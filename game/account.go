package game

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
)

// accountSchemaVersion is the on-disk schema version of account.json. Bumped only
// for migrations; readers default unknown/older fields to zero (the accounts design §3.3).
//
// account.json is exactly what every build since accounts shipped signs and verifies.
// Badges do not live in it: they have a file of their own beside it (badgeFileName,
// account_badges.go), so a build from before badges reads and writes account.json as it
// always did and can never read it as modified.
const accountSchemaVersion = 1

// accountFileName is the fixed base name of the account file under the data dir.
// v1 resolves a single account.json — no multi-profile layout yet (the accounts design §3.1).
const accountFileName = "account.json"

// accountsDirName is the subdirectory under the data ROOT that holds the per-account
// slots: <root>/accounts/<account_id>/. Each slot carries that account's account.json
// and its own saves/ tree. The flat legacy layout (account.json + saves/ directly under
// the root) is migrated into a slot on first boot by migrateLegacyAccountIfNeeded.
const accountsDirName = "accounts"

// activePointerFileName is the base name of the active-account pointer file under the
// data ROOT: <root>/active-account. It holds the 32-hex account ID whose slot is the
// live one — the single source of truth for "which account is active" across boots.
const activePointerFileName = "active-account"

// activeAccountID is the process-global ID of the account whose scoped slot is live.
// dataDirectory() (the SCOPED dir) resolves through it, so reading the active account,
// listing saves and a run started under it resolve to <root>/accounts/<activeAccountID>/.
// Writes never follow it: Account.Save writes each account to its own slot, and SaveGame
// writes a run to the slot of the account that owns it. It is "" before any account is
// named/loaded (the brief first-run window) — see dataDirectory() for the empty-id
// behavior. Guarded by activeAccountMu so the boot goroutine and the UI never race the
// pointer.
var activeAccountID string

// activeAccountMu guards activeAccountID. It is a distinct, lightweight lock with no
// relation to the engine's ge.mu or an Account's a.mu — it only serializes reads/writes
// of the process-global active-account pointer.
var activeAccountMu sync.Mutex

// getActiveAccountID returns the currently-active account ID under the pointer lock.
func getActiveAccountID() string {
	activeAccountMu.Lock()
	defer activeAccountMu.Unlock()
	return activeAccountID
}

// setActiveAccountID sets the in-memory active-account ID under the pointer lock. It does
// NOT persist — callers that want the choice to survive a restart also call writeActivePointer.
func setActiveAccountID(id string) {
	activeAccountMu.Lock()
	defer activeAccountMu.Unlock()
	activeAccountID = id
}

// accountDir returns the on-disk slot directory for the account with the given ID:
// <root>/accounts/<id>/. An empty id collapses to <root>/accounts (filepath.Join drops
// the empty segment) — the brief pre-naming window, where nothing named is written yet.
func accountDir(id string) string {
	return filepath.Join(rootDataDir(), accountsDirName, id)
}

// dataDirectory is the SCOPED data dir: the active account's slot,
// <root>/accounts/<activeID>/. accountPath() and the saves dir resolve through it, so the
// active account is read from, and a new run saves into, the active account's slot (Phase A
// account-scoping). Account.Save does not use it (each account writes its own slot).
//
// When no account is active yet (empty id, the first-run window before a name is chosen),
// it collapses to <root>/accounts. That is harmless: nothing writes a *named* artifact in
// that window — LoadOrCreate returns a fresh unestablished account WITHOUT persisting it,
// and the create/migrate paths always have a real id before they MkdirAll/Save. So no file
// is ever created under an empty-id slot.
func dataDirectory() string {
	return accountDir(getActiveAccountID())
}

// activeAccountPointerPath returns the path of the active-account pointer under the ROOT:
// <root>/active-account. It deliberately uses rootDataDir() (NOT the scoped dataDirectory)
// — the pointer names which slot is active and therefore lives one level above the slots.
func activeAccountPointerPath() string {
	return filepath.Join(rootDataDir(), activePointerFileName)
}

// readActivePointer reads the active-account ID from <root>/active-account, trimming
// surrounding whitespace/newline. A missing pointer returns ("", nil): no active account
// has been persisted yet (fresh install / post-wipe), which is a normal state, not an
// error. Any other read error is surfaced.
func readActivePointer() (string, error) {
	data, err := os.ReadFile(activeAccountPointerPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read active-account pointer: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// writeActivePointer persists id as the active account in <root>/active-account, creating
// the root if needed and writing atomically (temp file + rename) so a crash mid-write can
// never leave a half-written pointer. It mirrors Save()/SaveGame()'s write discipline.
func writeActivePointer(id string) error {
	root := rootDataDir()
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("failed to create data root: %w", err)
	}
	path := activeAccountPointerPath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(id), 0644); err != nil {
		return fmt.Errorf("failed to write active-account pointer: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("failed to finalize active-account pointer: %w", err)
	}
	return nil
}

// migrateLegacyAccountIfNeeded moves a pre-Phase-A FLAT layout (account.json + saves/
// directly under the data ROOT) into the new account-scoped slot, non-destructively
// (the accounts design, Phase A). It is called at the top of LoadOrCreate/LoadAccount, BEFORE the
// active account is resolved, so the pointer it writes is in place for that resolution.
//
// TRIGGER (idempotent): it runs ONLY when <root>/account.json exists AND <root>/accounts
// does NOT yet exist. Once the accounts/ tree exists the layout is already migrated (or
// born scoped), so a second call is a clean no-op — safe to call on every boot.
//
// STEPS: read the legacy account.json to learn its AccountID → mkdir -p the slot →
// MOVE (os.Rename) the legacy saves/ into the slot FIRST, then MOVE account.json → write
// the active pointer. Renames are atomic and never delete data. Saves move BEFORE
// account.json so that if the saves rename fails the legacy account.json is still in place
// at the root and the trigger condition (no accounts/ dir) still holds, leaving the tree in
// a clean, re-migratable state. account-export.json (a user-created backup) is left alone.
func migrateLegacyAccountIfNeeded() error {
	root := rootDataDir()
	legacyAccount := filepath.Join(root, accountFileName)
	accountsRoot := filepath.Join(root, accountsDirName)

	// Trigger only when the flat account.json exists and we have NOT migrated yet.
	if _, err := os.Stat(legacyAccount); err != nil {
		return nil // no legacy account (absent) — nothing to migrate
	}
	if _, err := os.Stat(accountsRoot); err == nil {
		return nil // accounts/ already exists — already migrated (idempotent no-op)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat accounts dir during migration: %w", err)
	}

	// Learn the legacy account's ID so we name its slot correctly. A corrupt/unparseable
	// legacy file has no usable ID — leave it untouched (LoadOrCreate's corrupt path backs
	// it up later) rather than inventing a slot name.
	data, err := os.ReadFile(legacyAccount)
	if err != nil {
		return fmt.Errorf("failed to read legacy account during migration: %w", err)
	}
	var acct Account
	if err := json.Unmarshal(data, &acct); err != nil {
		return nil // unparseable legacy file — not a migration case; downstream handles it
	}
	if acct.AccountID == "" {
		return nil // no ID to key a slot on — leave the flat file for downstream handling
	}

	slot := accountDir(acct.AccountID)
	if err := os.MkdirAll(slot, 0755); err != nil {
		return fmt.Errorf("failed to create account slot during migration: %w", err)
	}

	// One-time pre-migration snapshot: with migration now confirmed (parseable id, accounts/
	// absent) but BEFORE the first os.Rename, copy the flat data/ aside so the player's
	// original state is recoverable. BEST-EFFORT — a snapshot failure must NOT abort the
	// migration: the renames below are atomic and corruption-safe on their own, so this is
	// belt-and-braces reassurance, and bricking startup over a failed bonus backup would be
	// strictly worse than skipping it. We capture the path/err locally; the game package has no
	// logger, so a failure is simply not propagated.
	_, _ = snapshotPreMigration()

	// Move saves/ FIRST (see doc above): a failure here leaves account.json at the root,
	// so the trigger still fires next boot and we retry cleanly.
	legacySaves := filepath.Join(root, "saves")
	if _, err := os.Stat(legacySaves); err == nil {
		if err := os.Rename(legacySaves, filepath.Join(slot, "saves")); err != nil {
			return fmt.Errorf("failed to migrate saves into account slot: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat legacy saves during migration: %w", err)
	}

	// Move account.json into the slot.
	if err := os.Rename(legacyAccount, filepath.Join(slot, accountFileName)); err != nil {
		return fmt.Errorf("failed to migrate account.json into account slot: %w", err)
	}

	// Record the migrated account as active so the next resolution finds it.
	if err := writeActivePointer(acct.AccountID); err != nil {
		return fmt.Errorf("failed to write active pointer during migration: %w", err)
	}
	return nil
}

// AccountUnlocks holds account-wide cosmetic unlocks (DATA, the accounts design §3.3).
// Phase 1 only round-trips these; the unlock API lands in Phase 3.
type AccountUnlocks struct {
	Themes []string `json:"themes,omitempty"`
}

// AccountStats holds lifetime, cross-save aggregates (DATA, the accounts design §3.3).
// Phase 1 only round-trips these; the engine hooks land in Phase 6.
type AccountStats struct {
	TotalPrestiges int    `json:"total_prestiges,omitempty"`
	HighestAge     string `json:"highest_age,omitempty"`
	// CivilizationsStarted counts the runs begun on the account: every new game, and
	// the run that follows each prestige and each Succumb (RecordCivilizationStarted).
	CivilizationsStarted int `json:"civilizations_started,omitempty"`
	// SavesCompleted is retired. Nothing ever counted it, and it has no meaning the
	// game can stand behind: a save is never completed (there is no ending, and the
	// closest thing, a prestige, is Total Prestiges). It is not shown anywhere. The
	// field stays so that a file or an export that carries the key still verifies and
	// keeps it.
	SavesCompleted int `json:"saves_completed,omitempty"`
	// PrestigesByAge counts prestiges by the age they were made from (Pacing
	// v2), so badges can tell an early taste (the Medieval to the Atomic
	// Age) from a full run (the Modern Age or deeper, PrestigeRunAge).
	// Prestiges recorded before it existed are only in TotalPrestiges.
	PrestigesByAge map[string]int `json:"prestiges_by_age,omitempty"`
}

// AccountPrefs holds preferences that travel with the account (the accounts design §3.3).
// Phase 1 only round-trips these; SetActiveTheme et al. land in Phase 3.
type AccountPrefs struct {
	ActiveTheme string `json:"active_theme,omitempty"`
	// Map settings (the Map panel and the dashboard mini map). Empty means
	// the default: the first registered style, the unicode glyph tier.
	MapStyle  string `json:"map_style,omitempty"`
	MapGlyphs string `json:"map_glyphs,omitempty"`
	// MapIconsHint records that the Map panel's one-time "Type icons" hint
	// was shown, so it is shown once per account.
	MapIconsHint bool `json:"map_icons_hint,omitempty"`
	// Minimap is the dashboard mini map setting: "on", "off", or empty for
	// the default (on).
	Minimap string `json:"minimap,omitempty"`
}

// Account is the per-player identity + meta-progression record, persisted to the active
// account's slot at <root>/accounts/<account_id>/account.json (Phase A account-scoping;
// pre-Phase-A it was the flat <root>/account.json). It is the single source of truth for
// identity (account ID, display name) and DATA (unlocks, lifetime stats, achievements, prefs).
//
// Integrity mirrors saves: Signature is the HMAC-SHA256 of the payload (with
// Signature zeroed) under saveHMACKey. A tampered file still loads — Tampered is
// set instead, the cosmetic analogue of the save CheaterBadge (the accounts design §3.4).
// The schema follows the accounts design §3.3; all DATA fields are present so the file
// round-trips, but the unlock/stats/prefs APIs that mutate them arrive in later
// phases.
type Account struct {
	Version     int       `json:"version"`
	AccountID   string    `json:"account_id"`
	DisplayName string    `json:"display_name,omitempty"`
	Created     time.Time `json:"created"`
	LastSeen    time.Time `json:"last_seen,omitempty"`

	// --- meta-progression (DATA) ---
	Unlocks AccountUnlocks `json:"unlocks,omitempty"`
	Stats   AccountStats   `json:"stats,omitempty"`
	// Achievements is the list of account achievements a build from before badges
	// keeps. The four became badges (each badge lists the key as an alias):
	// ensureBadges grants a badge for a key found here, and earning such a badge
	// adds its key here, so both kinds of build show the same thing.
	Achievements []string `json:"achievements,omitempty"`
	// Badges is the earned badges by key; Counters the lifetime counts badges are
	// judged on (only the ones a badge names); Days the calendar days the account
	// was played on. They are NOT part of account.json (json:"-") or its signature:
	// they are read from and written to the badge file beside it (account_badges.go).
	Badges   map[string]BadgeEarned `json:"-"`
	Counters map[string]float64     `json:"-"`
	Days     []string               `json:"-"`
	// BadgesTampered is the badge file's own tamper mark: its signature did not match,
	// or it was signed for another account. Every badge in the file is then crossed,
	// and so is every badge earned while it is set. It is saved in the badge file and
	// sticks, as Tampered does for account.json. An edited badge file never flags
	// account.json, and the other way round.
	BadgesTampered bool `json:"-"`

	// --- preferences (travel with the account) ---
	Prefs AccountPrefs `json:"prefs,omitempty"`

	// --- integrity (same scheme as saves) ---
	Signature string `json:"_sig,omitempty"`

	// Tampered is the cosmetic flag set when a loaded file's signature does not match
	// (the accounts design §3.4). It mirrors the save CheaterBadge: signalling, not a lockout.
	// It is persisted (omitempty, so clean files keep their bytes and signatures) and
	// covered by the signature, so it sticks: the next Save re-signs the file with the
	// flag in it instead of laundering the edit, and deleting the key by hand breaks the
	// signature and sets it again. Exports carry it too.
	Tampered bool `json:"tampered,omitempty"`

	// FreshlyCreated is the in-memory, non-persisted signal that LoadOrCreate just
	// minted this account on its fresh-create path (no file existed). Boot code reads
	// it to surface a one-time, non-blocking first-run notice (the accounts design §6) without
	// changing LoadOrCreate's signature. json:"-" keeps it off disk; it is false for
	// any account loaded from an existing file.
	FreshlyCreated bool `json:"-"`

	// dirty is the in-memory, non-persisted write-debounce flag for the lifetime-stats
	// hooks (the accounts design §8 "debounce writes"). RecordPrestige/RecordAgeReached mutate
	// the in-memory stats and set dirty=true WITHOUT touching the disk — they run under
	// the engine write lock (advanceAge/DoPrestige) where file I/O is forbidden. The
	// engine's periodic autosave block (outside ge.mu) calls FlushIfDirty, which Saves
	// once if dirty and clears the flag. json:"-" keeps it off disk and out of the sig.
	dirty bool `json:"-"`

	// pendingEarned is the badges earned and not yet announced: the dashboard drains
	// it (GameEngine.DrainEarnedBadges) and writes the toast and the log line. Not
	// saved: a badge earned just before the game closes is simply in the list.
	pendingEarned []string

	// badgeDisk is the badge file's bytes as last read or written (nil: no file), so
	// Save rewrites it only when the badges changed. badgeUnreadable marks a file
	// that would not parse: it is set aside, not overwritten, at the next write.
	badgeDisk       []byte
	badgeUnreadable bool

	// mu guards the unlock/prefs reads and writes (and the Save inside the mutating
	// methods): the account is read from the UI goroutine (HasTheme/ActiveTheme/
	// UnlockedThemes) while another goroutine writes (UnlockTheme/SetActiveTheme).
	// This is the account's OWN lock over the account's OWN file — fully independent
	// of the engine's ge.mu, so there is no engine-deadlock risk. It has no json tag
	// and is never serialized; a sync.Mutex zero value marshals fine, so signing and
	// round-trip are unaffected (signAccount marshals the struct verbatim).
	mu sync.Mutex
}

// newAccountID returns a fresh, stable account ID: 16 random bytes from crypto/rand,
// hex-encoded (the accounts design §3.3 — 128 random bits, not a v4 UUID specifically).
func newAccountID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to generate account id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// newAccount builds a fresh, unsigned Account with a new ID and Created/LastSeen
// set to now. The caller is responsible for calling Save to sign and persist it.
func newAccount() (*Account, error) {
	id, err := newAccountID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Account{
		Version:   accountSchemaVersion,
		AccountID: id,
		Created:   now,
		LastSeen:  now,
	}, nil
}

// normalizeAccountName canonicalizes a name for ID derivation: trimmed, lowercased,
// and with any internal whitespace run collapsed to a single space. So "  Bob   the
// Builder " and "bob the builder" normalize identically and therefore derive the same
// account ID. It is the identity key — re-entering the same name on a fresh machine
// regenerates the same ID (name-based recovery). DisplayName keeps the ORIGINAL trimmed
// casing; this normalized form never leaves the derivation.
func normalizeAccountName(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}

// accountIDFromName derives the 32-char hex account ID from a name:
// hex(sha256(normalize(name))[:16]). Deterministic — the same name always yields the
// same ID — and it keeps the existing 16-byte / 32-hex-char ID format intact, so the
// recovery code and save attribution keep working unchanged. The leading 16 bytes of
// the SHA-256 digest are ample collision resistance for a local-only cosmetic identity.
func accountIDFromName(name string) string {
	sum := sha256.Sum256([]byte(normalizeAccountName(name)))
	return hex.EncodeToString(sum[:16])
}

// AccountIDForName is the exported name→id derivation, so UI/command paths can resolve a typed
// account name to its slot id WITHOUT duplicating the hashing rule (e.g. `account switch <name>`
// checks whether accountDir(AccountIDForName(name)) holds an account). It is deterministic and
// side-effect-free — a pure function of the (normalized) name, never touching disk or the active
// account.
func AccountIDForName(name string) string {
	return accountIDFromName(name)
}

// Established reports whether this account has been named (DisplayName set). An account
// loaded from disk with no display name (e.g. a legacy random-id dev account) is NOT
// established, so boot/UI code can prompt the player to name it on first run.
func (a *Account) Established() bool {
	return a.DisplayName != ""
}

// LoadAccount loads an EXISTING account.json without ever creating one — the read-only
// counterpart to LoadOrCreate, used by the name-first boot flow (the UI mints the account
// after prompting for a name). Behavior by file state:
//
//   - Absent: return (nil, false, nil) — no file, nothing to load. The caller prompts.
//   - Present + parses + signature valid (or unsigned/legacy): return (acct, true, nil).
//   - Present + parses + signature INVALID: return (acct, true, nil) with Tampered=true —
//     a cosmetic flag, not a lockout. The file is NOT deleted.
//   - Present but UNPARSEABLE: back the bad file up to account.json.corrupt and return
//     (nil, false, nil) — there is no salvageable account, so the caller prompts fresh.
//
// It never writes on the happy/tampered paths; the only disk mutation is the .corrupt
// rename for an unparseable file (mirrors LoadOrCreate's corrupt handling, minus the
// create).
func LoadAccount() (*Account, bool, error) {
	// Same resolution as LoadOrCreate, minus the create: migrate a legacy flat layout,
	// then scope to the active account from the persisted pointer.
	if err := migrateLegacyAccountIfNeeded(); err != nil {
		return nil, false, err
	}
	id, err := readActivePointer()
	if err != nil {
		return nil, false, err
	}
	if id != "" {
		setActiveAccountID(id)
	}

	path := accountPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to read account: %w", err)
	}

	var acct Account
	if err := json.Unmarshal(data, &acct); err != nil {
		// Corrupt / unparseable → back it up and report "not found" so the caller
		// prompts for a fresh name (the accounts design §7).
		backup := path + ".corrupt"
		if renameErr := os.Rename(path, backup); renameErr != nil {
			return nil, false, fmt.Errorf("account file is corrupt and could not be backed up: %w", renameErr)
		}
		return nil, false, nil
	}

	if !verifyAccount(&acct) {
		// Signature present but mismatched → tampered. Flag it, still load it.
		acct.Tampered = true
	}
	acct.loadBadgeFile(filepath.Dir(path))
	// Confirm the loaded account as active (the loaded id is authoritative over the pointer).
	setActiveAccountID(acct.AccountID)
	return &acct, true, nil
}

// CreateNamedAccount mints a signed account whose identity is derived from name:
// AccountID = accountIDFromName(name), DisplayName = the trimmed ORIGINAL name (display
// keeps casing/whitespace; the ID does not). Re-entering the same name reproduces the
// same ID — that IS the cross-machine recovery story (the accounts design §3.5).
//
// MIGRATION: if an account file already exists (e.g. a legacy random-id dev account from
// the old LoadOrCreate auto-create), its DATA — unlocks, lifetime stats, achievements,
// prefs — is carried over into the new named account so earned progress is not lost when
// the identity is re-keyed to the name. Identity fields (the old random ID, Created) are
// replaced; the new named identity wins. A corrupt/unparseable existing file is ignored
// (LoadAccount backs it up) and the named account starts with empty data.
//
// SCOPING (Phase A): the new account is keyed to the name-derived id. Carry-over reads the
// CURRENTLY-active account FIRST (before the id switch), then the active id is repointed to
// the new name-derived id, its slot is created, the account is Saved into that slot, and the
// active pointer is persisted so the next boot resolves to it. The save mutex discipline is
// preserved — Save() takes no account mutex.
func CreateNamedAccount(name string) (*Account, error) {
	trimmed := strings.TrimSpace(name)
	now := time.Now()
	acct := &Account{
		Version:        accountSchemaVersion,
		AccountID:      accountIDFromName(trimmed),
		DisplayName:    trimmed,
		Created:        now,
		LastSeen:       now,
		FreshlyCreated: true,
	}

	// Carry over DATA from any pre-existing (currently-active) account so unlocks/stats
	// survive the identity re-key. Read this BEFORE switching the active id below, so it
	// resolves the prior slot. found=false (absent or corrupt) → start clean.
	if prior, found, err := LoadAccount(); err == nil && found && prior != nil {
		acct.Unlocks = prior.Unlocks
		acct.Stats = prior.Stats
		acct.Achievements = append([]string(nil), prior.Achievements...)
		// The badges come along too. They are signed again for the new ID when the
		// account is saved: this is the one way a badge file changes owner.
		acct.Badges = mergeBadges(prior.Badges, nil)
		acct.Counters = mergeCounters(prior.Counters, nil)
		acct.Days = mergeDays(prior.Days, nil)
		acct.BadgesTampered = prior.BadgesTampered
		acct.Prefs = prior.Prefs
		// Carried data keeps its tamper flag: re-keying must not launder an edited file.
		acct.Tampered = prior.Tampered
	}
	// And a badge file already in the slot the name leads to, if there is one.
	acct.adoptBadgeFile()

	// Switch the active account to the new name-derived id; Save writes into that
	// account's own slot, <root>/accounts/<id>/account.json.
	setActiveAccountID(acct.AccountID)
	if err := os.MkdirAll(accountDir(acct.AccountID), 0755); err != nil {
		return nil, fmt.Errorf("failed to create account slot: %w", err)
	}
	if err := acct.Save(); err != nil {
		return nil, err
	}
	if err := writeActivePointer(acct.AccountID); err != nil {
		return nil, err
	}
	return acct, nil
}

// makeActive makes the account in slot id the active one: the persisted pointer, then the
// in-memory id, so saves list from its slot and the next boot resolves to it. The pointer
// goes first so a failed write changes nothing. The caller has already confirmed the slot
// holds an account.
func makeActive(id string) error {
	if err := writeActivePointer(id); err != nil {
		return err
	}
	setActiveAccountID(id)
	return nil
}

// accountPath resolves the full path to the ACTIVE account's account.json:
// <root>/accounts/<activeID>/account.json. It resolves purely through the SCOPED
// dataDirectory() (Phase A account-scoping) — the pre-scoping binary/CWD-relative
// fallbacks are gone, since the flat legacy <root>/account.json is relocated into a slot
// by migrateLegacyAccountIfNeeded before any account read, so there is no flat file left
// to fall back to.
func accountPath() string {
	return filepath.Join(dataDirectory(), accountFileName)
}

// --- Multi-account API (Phase B: enumerate / switch / create) ---
//
// Phase A laid the account-scoped layout (<root>/active-account pointer +
// <root>/accounts/<id>/{account.json,saves/}); Phase B is the read/select API over it.
// It is start-screen plumbing ONLY — no UI, no running-game reset, no copy/export
// changes. The engine wrappers (engine.go) keep ge.account in sync after a switch/create;
// switching never touches game state.

// AccountSummary is a read-only, lock-free snapshot of one account slot for the
// account-picker UI (Phase B). It carries identity + the headline meta-progression a
// chooser needs, plus the integrity (Tampered) and selection (Active) flags. It is built
// from a slot's loaded account.json and never holds the live *Account, so a consumer can
// neither mutate account state nor race a writer.
type AccountSummary struct {
	AccountID      string `json:"account_id"`
	DisplayName    string `json:"display_name"`
	HighestAge     string `json:"highest_age"`
	TotalPrestiges int    `json:"total_prestiges"`
	// Badges is how many badges the account holds (integrity badges left out when the
	// count is made with a ruleset to tell them by).
	Badges   int       `json:"badges"`
	LastSeen time.Time `json:"last_seen"`
	Active   bool      `json:"active"`
	Tampered bool      `json:"tampered"`
}

// loadAccountFromSlot reads + verifies the account.json in a SPECIFIC slot by id, WITHOUT
// touching the process-global active account. It is the lock-free, side-effect-free core
// the enumeration (ListAccounts) and the switch/open paths share, so neither has to fall
// back to the global-mutating LoadAccount() to read a non-active slot.
//
// It does NOT call migrateLegacyAccountIfNeeded, readActivePointer, or setActiveAccountID
// — the slot path is built directly from id (accountDir(id)/account.json), so reading slot
// B can never repoint the active account at B. Behavior by file state mirrors LoadAccount's
// READ contract, minus any disk mutation:
//
//   - Absent: (nil, false, nil) — no account in that slot.
//   - Present + parses + signature valid (or unsigned/legacy): (acct, true, nil).
//   - Present + parses + signature INVALID: (acct, true, nil) with Tampered=true — the
//     cosmetic flag, honestly surfaced; the file is NOT touched.
//   - Present but UNPARSEABLE: (nil, false, nil). UNLIKE LoadAccount it does NOT rename the
//     bad file to .corrupt — enumeration is strictly read-only and must not mutate the disk
//     just by listing; the slot simply doesn't appear.
func loadAccountFromSlot(id string) (*Account, bool, error) {
	path := filepath.Join(accountDir(id), accountFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to read account slot %s: %w", id, err)
	}
	var acct Account
	if err := json.Unmarshal(data, &acct); err != nil {
		// Unparseable → not a salvageable account. Read-only: do not back it up here.
		return nil, false, nil
	}
	if !verifyAccount(&acct) {
		acct.Tampered = true
	}
	acct.loadBadgeFile(filepath.Dir(path))
	return &acct, true, nil
}

// ListAccounts enumerates every account slot under <root>/accounts/ into read-only
// summaries for the picker (Phase B). It is purely read-only and NEVER returns an error: a
// missing accounts/ dir (fresh install / pre-first-account) yields an empty slice, and any
// unreadable/unparseable slot is simply skipped.
//
// CRITICAL — it MUST NOT mutate the active account. It reads each slot directly via
// loadAccountFromSlot(id) (which keys off the slot id, not activeAccountID, and calls no
// setActiveAccountID), so listing can never repoint the active account. As belt-and-braces
// it also snapshots getActiveAccountID() up front and asserts nothing changed (the
// loadAccountFromSlot path makes that a guarantee, but the snapshot documents the invariant
// and guards against a future change to the load path).
//
// Each summary's Active is true only for the slot whose id == the (snapshotted) active id.
// A tampered-but-parseable slot appears with Tampered=true (integrity is surfaced, not
// hidden). Results are sorted Active-first, then DisplayName ascending, then AccountID — a
// stable order for the chooser.
func ListAccounts() []AccountSummary { return listAccounts(nil) }

// listAccounts is ListAccounts with a badge book to count each account's badges by: the
// badges it holds that the ruleset knows and that count toward completion, after bringing
// a file from an older version up in memory (nothing is written). With no book the count
// is what the file holds.
func listAccounts(book *badgeBook) []AccountSummary {
	active := getActiveAccountID()

	accountsRoot := filepath.Join(rootDataDir(), accountsDirName)
	entries, err := os.ReadDir(accountsRoot)
	if err != nil {
		// Missing accounts/ (or unreadable root) → no accounts yet. Never an error.
		return nil
	}

	var summaries []AccountSummary
	for _, e := range entries {
		if !e.IsDir() {
			continue // only <id>/ slots; ignore stray files
		}
		id := e.Name()
		acct, found, loadErr := loadAccountFromSlot(id)
		if loadErr != nil || !found || acct == nil {
			continue // no valid account.json in this slot → skip it
		}
		if book != nil {
			acct.ensureBadgesLocked(book) // a private copy: no lock, and never saved
		}
		summaries = append(summaries, AccountSummary{
			AccountID:      acct.AccountID,
			DisplayName:    acct.DisplayName,
			HighestAge:     acct.Stats.HighestAge,
			TotalPrestiges: acct.Stats.TotalPrestiges,
			Badges:         acct.countableBadges(book),
			LastSeen:       acct.LastSeen,
			Active:         acct.AccountID == active,
			Tampered:       acct.Tampered,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Active != summaries[j].Active {
			return summaries[i].Active // active sorts first
		}
		if summaries[i].DisplayName != summaries[j].DisplayName {
			return summaries[i].DisplayName < summaries[j].DisplayName
		}
		return summaries[i].AccountID < summaries[j].AccountID
	})

	// Invariant guard: enumeration must leave the active account untouched. The
	// loadAccountFromSlot path never calls setActiveAccountID, so this always holds; the
	// assignment is a no-op that documents (and future-proofs) the contract.
	setActiveAccountID(active)
	return summaries
}

// SwitchAccount makes the account in slot id the active one and returns it (Phase B). It is
// a start-screen operation: it repoints the active account (in-memory + persisted pointer)
// and loads that slot, but does NOT reset or otherwise touch running game state — the
// engine wrapper swaps ge.account; the running game (if any) is the caller's concern.
//
// It verifies accountDir(id)/account.json exists first, erroring with "no such account" if
// not (and leaving the active account unchanged). On success it sets the in-memory active
// id AND persists the pointer, so the save list and new games scope to the new slot via
// dataDirectory() and the next boot resolves to it. Tamper/verify behavior is preserved
// (the returned account carries Tampered=true if its signature is stale or it was flagged
// before). The engine's SwitchAccount wraps this with the live-account handling.
func SwitchAccount(id string) (*Account, error) {
	acct, found, err := loadAccountFromSlot(id)
	if err != nil {
		return nil, err
	}
	if !found || acct == nil {
		return nil, fmt.Errorf("There is no account with the ID %s.", id)
	}
	// Commit the switch only after the slot is confirmed loadable, so a failed switch
	// leaves the active account unchanged.
	if err := makeActive(id); err != nil {
		return nil, err
	}
	return acct, nil
}

// CreateAccount creates (or opens) the account whose identity is derived from name and
// makes it active (Phase B). It is deliberately DISTINCT from CreateNamedAccount: this path
// has NO CARRY-OVER. A brand-new account starts EMPTY — it never copies the currently-active
// account's unlocks/stats/achievements/prefs. (CreateNamedAccount keeps its carry-over for
// the first-run name-derivation flow, where re-keying a legacy random-id account to a name
// must preserve earned progress; it is left untouched.)
//
// Identity is name-derived (accountIDFromName), so the same name always maps to the same
// slot. Behavior by slot state:
//
//   - Slot already exists (this name maps to an existing account): do NOT clobber it. Open
//     it via SwitchAccount(id) and return it. "Create with an existing name" == "open that
//     account" — same name is the same identity, and wiping it would be data loss.
//   - Slot absent: mint a FRESH EMPTY established account (DisplayName=name, the name-derived
//     id, zero Unlocks/Stats/Achievements, default Prefs), set it active, MkdirAll its slot,
//     Save it in, and persist the active pointer. No prior account's DATA rides along.
func CreateAccount(name string) (*Account, error) {
	trimmed := strings.TrimSpace(name)
	id := accountIDFromName(trimmed)

	// Same-name → open the existing account rather than overwrite it. Slot existence is
	// keyed on its account.json, so a bare/empty slot dir doesn't count as "exists".
	if _, err := os.Stat(filepath.Join(accountDir(id), accountFileName)); err == nil {
		return SwitchAccount(id)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat account slot %s: %w", id, err)
	}

	// Fresh, EMPTY established account — no carry-over. DATA fields stay zero.
	now := time.Now()
	acct := &Account{
		Version:        accountSchemaVersion,
		AccountID:      id,
		DisplayName:    trimmed,
		Created:        now,
		LastSeen:       now,
		FreshlyCreated: true,
	}
	// A badge file left in the slot (its account.json was lost) is this account's.
	acct.adoptBadgeFile()
	// Make it active; Save writes into its own slot, <root>/accounts/<id>/account.json.
	setActiveAccountID(id)
	if err := os.MkdirAll(accountDir(id), 0755); err != nil {
		return nil, fmt.Errorf("failed to create account slot: %w", err)
	}
	if err := acct.Save(); err != nil {
		return nil, err
	}
	if err := writeActivePointer(id); err != nil {
		return nil, err
	}
	return acct, nil
}

// WipeAccount permanently deletes the ACTIVE account's slot —
// <root>/accounts/<activeID>/ and everything under it (account.json, its .corrupt
// sibling, and the slot's own saves/) — then clears the active-account pointer and the
// in-memory active id, so the next boot starts from a clean slate and re-prompts for a
// name. It is the account analogue of WipeAllSaves — destructive and irreversible (no
// server backup).
//
// SCOPE (Phase A): it removes only the ACTIVE account's slot. OTHER accounts' slots under
// <root>/accounts/ are spared, and any user progress-export file (account-export.json,
// which lives outside the slots) is never touched — those are independent backups the
// player may have deliberately created. With no active account (empty id) it is a clean
// no-op: it never falls back to removing the shared <root>/accounts root. Wiping twice (or
// with nothing active) is a clean no-op — os.RemoveAll on a missing path is not an error.
//
// BACKUP: it delegates to WipeAccountByID(activeID), which snapshots the slot into
// <root>/backups/ BEFORE removing it (the wipe is permanent, so we leave a recoverable copy).
// The backup path is discarded here — the active-wipe callers don't surface it — and a backup
// failure never blocks the wipe (only the wipe's own error is returned).
func WipeAccount() error {
	id := getActiveAccountID()
	if id == "" {
		// Nothing active to wipe. Do NOT remove accountDir("") — that is the shared
		// <root>/accounts root, never a single account's slot.
		return nil
	}
	_, err := WipeAccountByID(id)
	return err
}

// ExportAccountByID exports the progress blob for the account in slot id WITHOUT touching the
// active account (Phase D). It is the by-id sibling of (a *Account) ExportProgress, used by the
// Accounts panel to back up the SELECTED account (which may not be the live one). It loads the
// slot read-only via loadAccountFromSlot (no setActiveAccountID, no pointer write) and serializes
// it; ExportProgress only READS the loaded account, so the active account is never disturbed.
//
// Errors when the slot has no loadable account.json ("no such account: <id>"). A tampered-but-
// parseable slot still exports (Tampered is cosmetic; the blob re-signs on its own scheme).
func ExportAccountByID(id string) ([]byte, error) {
	acct, found, err := loadAccountFromSlot(id)
	if err != nil {
		return nil, err
	}
	if !found || acct == nil {
		return nil, fmt.Errorf("There is no account with the ID %s.", id)
	}
	return acct.ExportProgress()
}

// RecoveryCodeForID returns the recovery code for the account in slot id WITHOUT touching the
// active account (Phase D). The code is a pure function of the account ID (RecoveryCode derives
// it from AccountID alone), so loading the slot read-only is sufficient — and confirms the slot
// actually exists, so the panel can't surface a code for a non-existent account. Errors with
// "no such account: <id>" when the slot has no loadable account.json.
func RecoveryCodeForID(id string) (string, error) {
	acct, found, err := loadAccountFromSlot(id)
	if err != nil {
		return "", err
	}
	if !found || acct == nil {
		return "", fmt.Errorf("There is no account with the ID %s.", id)
	}
	return acct.RecoveryCode(), nil
}

// AccountExportPath returns the default on-disk export path for the account in slot id:
// <root>/accounts/<id>/account-<id8>-export.json. It is a pure path helper (no I/O) used by the
// Accounts panel so the by-id export of the SELECTED account lands inside that account's own
// slot — never in the ACTIVE account's DataDir(), which would be wrong for a non-active
// selection. The blob carries the full id regardless; the filename short-id is just a
// human-friendly disambiguator when several exports share a directory.
func AccountExportPath(id string) string {
	return filepath.Join(accountDir(id), fmt.Sprintf("account-%s-export.json", shortID(id)))
}

// WipeAccountByID permanently deletes the slot for account id — <root>/accounts/<id>/ and
// everything under it (Phase D). It is the by-id sibling of WipeAccount (which targets the
// ACTIVE slot): the Accounts panel uses it to wipe the SELECTED account, which may not be the
// live one. OTHER slots are always spared (os.RemoveAll(accountDir(id)) touches only that one
// directory), and any user export file outside the slots is untouched.
//
// Guard: an empty id is refused (a no-op error), since accountDir("") collapses to the shared
// <root>/accounts root — removing it would nuke EVERY account. When id IS the active account,
// it also clears the active-account pointer + resets the in-memory active id (mirroring
// WipeAccount), so the next boot starts clean and the UI re-prompts. Wiping a non-active slot
// leaves the active pointer alone.
//
// BACKUP (best-effort, BEFORE removal): the wipe is permanent, so we snapshot the slot into
// <root>/backups/ first via BackupAccount and RETURN that path so the UI can tell the player
// where the recoverable copy landed. A backup failure does NOT block the wipe — the player
// explicitly confirmed — it just yields backupPath="" (the backup error is intentionally not
// fatal; only the wipe's own error is returned). Because the backups live OUTSIDE the slot
// (<root>/backups, not <root>/accounts/<id>), the os.RemoveAll below never deletes the snapshot
// it just took.
func WipeAccountByID(id string) (backupPath string, err error) {
	if id == "" {
		// Never remove accountDir("") — that is the shared <root>/accounts root, not a slot.
		return "", fmt.Errorf("Cannot wipe an account without an ID.")
	}
	// Snapshot first. Best-effort: a backup error is swallowed (backupPath stays "") so the
	// player's confirmed wipe still proceeds; the snapshot lives outside the slot, so it
	// survives the removal below.
	if path, backupErr := BackupAccount(id); backupErr == nil {
		backupPath = path
	}
	if rmErr := os.RemoveAll(accountDir(id)); rmErr != nil {
		return backupPath, fmt.Errorf("failed to wipe account slot %s: %w", accountDir(id), rmErr)
	}
	// If we just wiped the ACTIVE slot, clear the pointer + in-memory id so nothing stale is
	// resolved next boot (a non-active wipe must leave the active account untouched).
	if id == getActiveAccountID() {
		if rmErr := os.Remove(activeAccountPointerPath()); rmErr != nil && !os.IsNotExist(rmErr) {
			return backupPath, fmt.Errorf("failed to clear active-account pointer: %w", rmErr)
		}
		setActiveAccountID("")
	}
	return backupPath, nil
}

// shortID returns the first 8 chars of an account id (enough to recognize a slot at a glance),
// or the whole string when shorter. The package-level twin of the UI's shortAccountID, kept here
// so AccountExportPath can name files without reaching across into the ui package.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// signAccount returns the HMAC-SHA256 hex of the account payload with Signature
// zeroed, so the signature covers the data only — identical construction to
// signSave, sharing the hmacSign helper (the accounts design §3.4).
//
// It takes a pointer (not a value) so the sync.Mutex field is never copied — a
// value copy would trip go vet's copylocks. The signed bytes are unchanged from a
// value-based marshal: json.Marshal already ignores unexported fields (mu) and the
// json:"-" FreshlyCreated field, and we build the payload from a freshly-constructed
// struct literal that copies only the serializable fields, with Signature zeroed — so
// the signature covers exactly the on-disk data. Tampered is part of that data: it is
// omitempty, so an untampered account signs to the same bytes as before the flag was
// persisted, and a flagged one cannot shed the flag without breaking the signature.
func signAccount(a *Account) string {
	payload := Account{
		Version:      a.Version,
		AccountID:    a.AccountID,
		DisplayName:  a.DisplayName,
		Created:      a.Created,
		LastSeen:     a.LastSeen,
		Unlocks:      a.Unlocks,
		Stats:        a.Stats,
		Achievements: a.Achievements,
		Prefs:        a.Prefs,
		Tampered:     a.Tampered,
		// Signature deliberately zero; FreshlyCreated/mu are json:"-"/unexported.
		// The badges are not here: they are signed in their own file.
	}
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

// verifyAccount reports whether a's stored Signature matches a freshly computed
// one. An unsigned file (empty Signature) is treated as legacy/benign-valid, the
// same benefit-of-the-doubt verifySave grants unsigned saves (the accounts design §3.4).
func verifyAccount(a *Account) bool {
	if a.Signature == "" {
		return true
	}
	return hmac.Equal([]byte(a.Signature), []byte(signAccount(a)))
}

// validAccountID reports whether id has the shape every account ID has: 32 lowercase
// hex characters (16 bytes, from newAccountID, accountIDFromName or a recovery code). An
// ID names a directory under <root>/accounts/, so anything else (an empty ID, which
// would collapse onto the shared accounts root, or a path from a hand-made export) is
// refused before it reaches the disk.
func validAccountID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// Save signs the account with the shared HMAC helper and writes it atomically
// (temp file + os.Rename) into the account's OWN slot, <root>/accounts/<AccountID>/,
// creating it if missing. Mirrors SaveGame's write discipline (the accounts design §3.4).
// It writes account.json, then the badge file beside it if the badges changed
// (saveBadgeFile). The two are signed apart; if the second write fails or the game
// stops between them, the next load tops the badges up from the account's record.
//
// It deliberately does not resolve through the active account: whichever account is
// active, an account's data only ever lands in its own file. So an old account object
// flushed after a switch, a recovery code restored while another account is active, or
// an import aimed at another slot can never overwrite a different account.
func (a *Account) Save() error {
	if !validAccountID(a.AccountID) {
		return fmt.Errorf("cannot save an account with the ID %q", a.AccountID)
	}
	dir := accountDir(a.AccountID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	a.Signature = signAccount(a)
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	path := filepath.Join(dir, accountFileName)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write account: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to finalize account: %w", err)
	}
	// The badges go to their own file in the same slot, when they changed.
	return a.saveBadgeFile(dir)
}

// LoadOrCreate resolves the active account (migrating a legacy flat layout, then reading
// the active-account pointer) and returns the live Account, creating one transparently when
// the active slot has none (the accounts design §6/§7). Behavior by file state:
//
//   - Absent: generate a fresh account (new ID, Created=now), Save it, return it.
//   - Present + parses + signature valid (or unsigned/legacy): return it.
//   - Present + parses + signature INVALID: return it with Tampered=true set —
//     a cosmetic flag, not a lockout. The file is NOT deleted.
//   - Present but UNPARSEABLE: back the bad file up to account.json.corrupt, then
//     create + Save a fresh account and return it.
//
// Note: this phase does not stamp LastSeen on load (that is Phase 2 startup
// wiring); LoadOrCreate's contract here is read-or-create, leaving the on-disk
// file untouched on the valid/tampered paths.
func LoadOrCreate() (*Account, error) {
	// Relocate any pre-Phase-A flat layout into a slot before resolving the active
	// account, so the pointer it writes is honored by the resolution below.
	if err := migrateLegacyAccountIfNeeded(); err != nil {
		return nil, err
	}

	// Resolve the active account from the persisted pointer. With an id set, scope the
	// reads to that slot so accountPath() resolves to <root>/accounts/<id>/account.json.
	id, err := readActivePointer()
	if err != nil {
		return nil, err
	}
	if id != "" {
		setActiveAccountID(id)
	}

	path := accountPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No account in the active slot → mint, sign, and persist a fresh one
			// (the accounts design §6/§7). It becomes the active account: set the in-memory id,
			// Save into its now-resolved slot, then persist the pointer so the next boot
			// finds it. First genuine run → FreshlyCreated=true.
			return createFreshActive(true)
		}
		return nil, fmt.Errorf("failed to read account: %w", err)
	}

	var acct Account
	if err := json.Unmarshal(data, &acct); err != nil {
		// Corrupt / unparseable → back it up to <slot>/account.json.corrupt, then mint a
		// fresh account (the accounts design §7). The fresh account has a new ID, so it gets its
		// own slot and becomes the active account (pointer included), the same as the
		// absent-file path; the .corrupt backup stays in the old slot. (It used to be saved
		// into the old slot under its new ID, leaving a slot whose folder named a different
		// account.) The corrupt-recovery path is NOT a first run → FreshlyCreated stays false.
		backup := path + ".corrupt"
		if renameErr := os.Rename(path, backup); renameErr != nil {
			return nil, fmt.Errorf("account file is corrupt and could not be backed up: %w", renameErr)
		}
		return createFreshActive(false)
	}

	if !verifyAccount(&acct) {
		// Signature present but mismatched → tampered. Flag it, still load it.
		acct.Tampered = true
	}
	acct.loadBadgeFile(filepath.Dir(path))
	// Established account loaded from its slot — confirm it as the active account.
	setActiveAccountID(acct.AccountID)
	return &acct, nil
}

// createFreshActive mints a fresh random-id account, makes it the active account, and
// persists it into its own slot plus the active pointer. It is the scoped equivalent of
// the pre-Phase-A "absent file → create + Save" path: Save writes to
// <root>/accounts/<id>/account.json (never an empty-id slot), and writeActivePointer
// records the choice for the next boot. freshlyCreated sets FreshlyCreated so boot code
// can surface the one-time first-run notice.
func createFreshActive(freshlyCreated bool) (*Account, error) {
	acct, err := newAccount()
	if err != nil {
		return nil, err
	}
	setActiveAccountID(acct.AccountID)
	if err := acct.Save(); err != nil {
		return nil, err
	}
	if err := writeActivePointer(acct.AccountID); err != nil {
		return nil, err
	}
	acct.FreshlyCreated = freshlyCreated
	return acct, nil
}

// --- Unlock API (the accounts design §8; the theming design §5 is the first caller) ---
//
// The account is key-agnostic: it stores and reports unlocked-theme keys and the
// active-theme key without judging which are valid or always-unlocked. Theming
// owns that policy (the always-unlocked accessibility + Forge set, and unknown-key
// fallback). Accounts is the persisted store underneath it.

// HasTheme reports whether key is in the account's unlocked-theme set.
func (a *Account) HasTheme(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.hasThemeLocked(key)
}

// hasThemeLocked is the lock-free core of HasTheme. Callers must hold a.mu.
func (a *Account) hasThemeLocked(key string) bool {
	for _, k := range a.Unlocks.Themes {
		if k == key {
			return true
		}
	}
	return false
}

// UnlockTheme records key as an unlocked theme and persists the account. If the
// theme was already unlocked it is a no-op: (false, nil) with no write. Otherwise
// it appends the key (deduped via the membership check), persists via Save, and
// returns (true, <Save error>) — newly is true even if the subsequent Save fails,
// since the in-memory set did change. The theming UI fires the unlock toast only on
// newly==true, so replayed milestone checks never re-toast an owned theme.
func (a *Account) UnlockTheme(key string) (newly bool, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.hasThemeLocked(key) {
		return false, nil
	}
	a.Unlocks.Themes = append(a.Unlocks.Themes, key)
	return true, a.Save()
}

// UnlockedThemes returns the unlocked-theme keys in deterministic (sorted) order.
// It returns a fresh copy, so callers can't mutate the account's backing slice.
func (a *Account) UnlockedThemes() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.Unlocks.Themes))
	copy(out, a.Unlocks.Themes)
	sort.Strings(out)
	return out
}

// ActiveTheme returns the persisted active-theme key, or "" if none is set.
func (a *Account) ActiveTheme() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Prefs.ActiveTheme
}

// SetActiveTheme persists key as the active theme and returns the Save error.
// It does NOT validate that key is unlocked — accounts is key-agnostic and theming
// owns validity + the always-unlocked policy. The only error path is the Save error.
func (a *Account) SetActiveTheme(key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Prefs.ActiveTheme = key
	return a.Save()
}

// MapPrefs returns the persisted map settings: the style key and glyph tier
// name ("" for the defaults) and whether the icons hint was shown.
func (a *Account) MapPrefs() (style, glyphs string, hintShown bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Prefs.MapStyle, a.Prefs.MapGlyphs, a.Prefs.MapIconsHint
}

// SetMapStyle persists the map style key. Like SetActiveTheme it does not
// validate the key: the UI owns the style registry.
func (a *Account) SetMapStyle(key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Prefs.MapStyle = key
	return a.Save()
}

// SetMapGlyphs persists the map glyph tier name.
func (a *Account) SetMapGlyphs(tier string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Prefs.MapGlyphs = tier
	return a.Save()
}

// MinimapOn reports whether the dashboard's mini map is on (the default).
func (a *Account) MinimapOn() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Prefs.Minimap != "off"
}

// SetMinimap persists the mini map setting.
func (a *Account) SetMinimap(on bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Prefs.Minimap = "off"
	if on {
		a.Prefs.Minimap = "on"
	}
	return a.Save()
}

// SetMapIconsHintShown records that the one-time icons hint was shown.
func (a *Account) SetMapIconsHintShown() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Prefs.MapIconsHint {
		return nil
	}
	a.Prefs.MapIconsHint = true
	return a.Save()
}

// --- Recovery code (identity backup, the accounts design §3.5 / §8 / §9 Phase 4) ---
//
// The recovery code encodes IDENTITY ONLY — the 16-byte account_id plus a 2-byte
// checksum — into a short, dash-grouped, uppercase, Crockford-base32 string with an
// `AGEF-` prefix (e.g. AGEF-7Q2K-9X4M-ZJ31-...). It restores who you are across
// machines/reinstalls; it does NOT restore earned progress (unlocks/stats) — that is
// DATA, backed up separately via export/import (Phase 5). The code is a convenience
// identifier, not a credential (the accounts design §3.5): the checksum guards against TYPOS,
// not forgery, and account state is cosmetic, not security-critical.

// recoveryCodePrefix is the human-readable namespace stamped on every recovery code.
const recoveryCodePrefix = "AGEF"

// crockfordAlphabet is the Crockford base32 symbol set: 0-9 then A-Z excluding the
// ambiguous I, L, O, U. Index = 5-bit value; the string is the canonical encoder.
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// crc16CCITT computes the CRC-16/CCITT-FALSE checksum of data: 16-bit, polynomial
// 0x1021, init 0xFFFF, no reflection, no final XOR. It is a small, standard,
// dependency-free typo guard for the recovery code (the accounts design §3.5).
func crc16CCITT(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// crockfordEncode encodes data as an uppercase Crockford base32 string (no padding).
// Bits are packed MSB-first; a trailing partial group is left-padded with zero bits,
// matching crockfordDecode's symmetric unpacking.
func crockfordEncode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var out []byte
	var buf uint32
	bits := 0
	for _, b := range data {
		buf = (buf << 8) | uint32(b)
		bits += 8
		for bits >= 5 {
			bits -= 5
			out = append(out, crockfordAlphabet[(buf>>uint(bits))&0x1F])
		}
	}
	if bits > 0 {
		out = append(out, crockfordAlphabet[(buf<<uint(5-bits))&0x1F])
	}
	return string(out)
}

// crockfordDecodeChar maps a single character to its 5-bit value using Crockford's
// lenient rules: case-insensitive, with I/L→1, O→0, U→V (per the Crockford spec, U is
// treated as V to absorb a common transcription slip). Returns (value, ok).
func crockfordDecodeChar(c byte) (byte, bool) {
	switch {
	case c >= 'a' && c <= 'z':
		c -= 'a' - 'A' // normalize to upper
	}
	switch c {
	case 'O':
		c = '0'
	case 'I', 'L':
		c = '1'
	case 'U':
		c = 'V'
	}
	for i := 0; i < len(crockfordAlphabet); i++ {
		if crockfordAlphabet[i] == c {
			return byte(i), true
		}
	}
	return 0, false
}

// crockfordDecode decodes a Crockford base32 string (already stripped of the prefix,
// dashes, and spaces) back to bytes. byteLen is the expected decoded length; any
// trailing partial-group bits beyond byteLen*8 are discarded (they are the encoder's
// zero padding). Returns an error on an unknown symbol or insufficient input.
func crockfordDecode(s string, byteLen int) ([]byte, error) {
	out := make([]byte, 0, byteLen)
	var buf uint32
	bits := 0
	for i := 0; i < len(s); i++ {
		v, ok := crockfordDecodeChar(s[i])
		if !ok {
			return nil, fmt.Errorf("That recovery code has a character it cannot contain ('%s'). Check it and try again.", string(s[i]))
		}
		buf = (buf << 5) | uint32(v)
		bits += 5
		if bits >= 8 {
			bits -= 8
			out = append(out, byte((buf>>uint(bits))&0xFF))
		}
	}
	if len(out) < byteLen {
		return nil, fmt.Errorf("That recovery code is too short. Check it and try again.")
	}
	return out[:byteLen], nil
}

// groupBy4 inserts a dash after every 4 characters, leaving any trailing partial
// group as-is (e.g. 29 chars → 7 groups of 4 + 1). Matches the doc's example shape.
func groupBy4(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if i > 0 && i%4 == 0 {
			b.WriteByte('-')
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// RecoveryCode returns this account's identity-recovery code (the accounts design §3.5/§8):
// the 16 raw account-id bytes plus a 2-byte CRC-16 checksum, Crockford-base32 encoded,
// uppercased, dash-grouped in 4s, with the AGEF- prefix. Identity only — never progress.
func (a *Account) RecoveryCode() string {
	a.mu.Lock()
	id := a.AccountID
	a.mu.Unlock()

	idBytes, err := hex.DecodeString(id)
	if err != nil || len(idBytes) != 16 {
		// An account ID that isn't 16 hex bytes can't form a code; return the prefix
		// only rather than panic. LoadOrCreate always produces a valid 16-byte ID.
		return recoveryCodePrefix + "-"
	}
	sum := crc16CCITT(idBytes)
	payload := make([]byte, 0, 18)
	payload = append(payload, idBytes...)
	payload = append(payload, byte(sum>>8), byte(sum&0xFF))
	body := groupBy4(crockfordEncode(payload))
	return recoveryCodePrefix + "-" + body
}

// RecoveryCodeID decodes a recovery code, verifies its checksum, and returns the account
// ID it carries, without touching the disk. Input is normalized leniently: uppercased,
// the AGEF- prefix stripped, dashes and spaces removed, and ambiguous Crockford
// characters mapped (I/L→1, O→0, U→V). A checksum mismatch returns a clear typo-guard
// error rather than silently naming a wrong account. The UI uses it to check a code (and
// to spot the player's own code) before asking for a confirm or changing anything.
func RecoveryCodeID(code string) (string, error) {
	// Normalize: drop spaces, uppercase, strip the AGEF- prefix, strip dashes.
	norm := strings.ToUpper(strings.TrimSpace(code))
	norm = strings.ReplaceAll(norm, " ", "")
	norm = strings.ReplaceAll(norm, "-", "")
	if p := strings.ToUpper(recoveryCodePrefix); strings.HasPrefix(norm, p) {
		norm = norm[len(p):]
	}
	if norm == "" {
		return "", fmt.Errorf("The recovery code is empty. Type account recover <code>.")
	}

	payload, err := crockfordDecode(norm, 18)
	if err != nil {
		return "", err
	}
	idBytes := payload[:16]
	gotSum := uint16(payload[16])<<8 | uint16(payload[17])
	if gotSum != crc16CCITT(idBytes) {
		return "", fmt.Errorf("That recovery code does not check out (a character is probably wrong). Check it and try again.")
	}
	return hex.EncodeToString(idBytes), nil
}

// ImportRecoveryCode restores the identity in a recovery code into THAT account's own
// slot, <root>/accounts/<id>/ (the accounts design §3.5/§8), and returns the account there:
//
//   - The slot already holds an account (the code of an account on this machine, the
//     player's own included): it is returned untouched. Recovering never overwrites an
//     account.
//   - The slot is empty: a FRESH signed account.json carrying the recovered ID with EMPTY
//     data is written there. Identity only; unlocks, stats and achievements come back
//     only from an export. An unreadable account.json already in the slot is moved aside
//     to account.json.corrupt first, never overwritten.
//
// It never touches the active account or any other slot (it used to write the empty
// account over whichever account was active). The engine's RecoverAccount switches to the
// restored account afterwards. It reuses the normal signed atomic Save path; it never
// hand-rolls a second integrity scheme.
func ImportRecoveryCode(code string) (*Account, error) {
	id, err := RecoveryCodeID(code)
	if err != nil {
		return nil, err
	}
	existing, found, err := loadAccountFromSlot(id)
	if err != nil {
		return nil, err
	}
	if found && existing != nil {
		return existing, nil
	}
	path := filepath.Join(accountDir(id), accountFileName)
	if _, statErr := os.Stat(path); statErr == nil {
		if err := os.Rename(path, path+".corrupt"); err != nil {
			return nil, fmt.Errorf("The account file for that code is damaged and could not be set aside: %w", err)
		}
	}

	now := time.Now()
	acct := &Account{
		Version:   accountSchemaVersion,
		AccountID: id,
		Created:   now,
		LastSeen:  now,
		// EMPTY data: identity only. Unlocks/Stats/Achievements/Prefs stay zero —
		// progress is carried by export/import (Phase 5), not the recovery code.
	}
	// A badge file left in the slot (its account.json was lost) is this account's.
	acct.adoptBadgeFile()
	if err := acct.Save(); err != nil {
		return nil, err
	}
	return acct, nil
}

// --- Progress export / import (single-account backup, the accounts design §3.6 / §8 / §9 Phase 5/C) ---
//
// Distinct from the recovery code: the recovery code carries IDENTITY only (account id),
// while an export is a full single-account BACKUP — identity (AccountID + DisplayName) AND
// progress (unlocks, lifetime stats, achievements, prefs). With no server, DATA cannot be
// reconstituted from nothing, so export is the explicit one-action backup (the accounts design §3.6).
//
// Phase C re-homes import: the old `(a *Account) ImportProgress` folded a blob into whatever
// account was live, which silently cross-contaminated accounts (importing B's backup while A
// was active wrote B's data into A). Import is now the package function ImportAccountExport,
// which resolves the blob's OWN slot by its AccountID and lands the data THERE — creating the
// slot if absent, merging/replacing if present — without disturbing the active account. The
// engine's ImportAccountExport folds a backup of the account in use into the live object
// instead. Neither switches accounts.

// progressExport is the self-describing on-disk shape of an export blob: a format
// version, the owning account's IDENTITY (AccountID + DisplayName), the DATA fields
// (unlocks, lifetime stats, achievements, prefs), and an HMAC signature over the
// sig-zeroed payload (same scheme as Account.Save — reuse, don't reinvent).
//
// Phase C binds the export to its account: AccountID/DisplayName ride along so an
// export is a full single-account BACKUP (identity + progress), and import lands the
// blob in that account's OWN slot rather than folding into whoever happens to be active.
// The stats/achievements fields ride along so Phase 6 data exports without a format bump.
// Tampered carries the account's tamper flag (omitempty, signed), so exporting an edited
// account and importing it elsewhere keeps it flagged instead of laundering it.
type progressExport struct {
	Version      int            `json:"version"`
	AccountID    string         `json:"account_id,omitempty"`
	DisplayName  string         `json:"display_name,omitempty"`
	Unlocks      AccountUnlocks `json:"unlocks,omitempty"`
	Stats        AccountStats   `json:"stats,omitempty"`
	Achievements []string       `json:"achievements,omitempty"`
	Prefs        AccountPrefs   `json:"prefs,omitempty"`
	Tampered     bool           `json:"tampered,omitempty"`
	Signature    string         `json:"_sig,omitempty"`
	// BadgeStore is the account's badge file, carried whole with its own signature
	// (which binds it to the account ID). It is NOT under Signature above: that one
	// covers exactly what a build from before badges signs, so such a build still
	// verifies and imports the export, and simply does not see this key.
	BadgeStore *badgeFile `json:"badge_store,omitempty"`
}

// signProgressExport returns the HMAC-SHA256 hex of the export payload with Signature
// zeroed, under saveHMACKey — identical construction to signAccount/signSave. The
// signature COVERS AccountID + DisplayName, so an export cannot be re-attributed to a
// different account without breaking the sig, and Tampered, so the flag cannot be
// dropped from a flagged export by hand.
func signProgressExport(p *progressExport) string {
	payload := progressExport{
		Version:      p.Version,
		AccountID:    p.AccountID,
		DisplayName:  p.DisplayName,
		Unlocks:      p.Unlocks,
		Stats:        p.Stats,
		Achievements: p.Achievements,
		Prefs:        p.Prefs,
		Tampered:     p.Tampered,
		// Signature deliberately zero. BadgeStore is signed on its own.
	}
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

// ExportProgress builds a signed export blob — a full single-account BACKUP carrying this
// account's IDENTITY (AccountID + DisplayName) AND its DATA (unlocks, lifetime stats,
// achievements, prefs) (the accounts design §3.6, Phase C). It signs the blob with the shared HMAC
// helper (same zero-sig-then-marshal pattern as Save) and returns pretty JSON bytes the
// caller writes to a file/clipboard. Because the blob carries its account id, import can
// land it back in that account's OWN slot regardless of which account is active.
func (a *Account) ExportProgress() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	exp := progressExport{
		Version:      accountSchemaVersion,
		AccountID:    a.AccountID,
		DisplayName:  a.DisplayName,
		Unlocks:      a.Unlocks,
		Stats:        a.Stats,
		Achievements: a.Achievements,
		Prefs:        a.Prefs,
		Tampered:     a.Tampered,
		BadgeStore:   a.exportBadgeFileLocked(),
	}
	exp.Signature = signProgressExport(&exp)

	data, err := json.MarshalIndent(&exp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal progress export: %w", err)
	}
	return data, nil
}

// ImportAccountExport unmarshals and verifies a single-account export blob, then lands it
// in the account's OWN slot — the slot named by the blob's AccountID, NOT whatever account
// is currently active (the accounts design §3.6, Phase C). It returns the imported *Account.
//
// INTEGRITY (first, before any disk mutation): the blob is unmarshalled and its signature
// recomputed over the sig-zeroed payload; a mismatch (or a missing/empty sig — exports are
// always written signed by ExportProgress) is rejected with the same wording as before, and
// nothing is written. A blob whose AccountID is empty after unmarshal is an old/foreign
// export with no slot to target; it is rejected with a clear "missing account id" error.
//
// TARGET SLOT (resolved by the blob's AccountID via loadAccountFromSlot, which reads a slot
// WITHOUT touching the active global):
//
//   - Slot EXISTS: apply the merge/replace policy onto THAT slot's account and Save it back
//     into its own slot. merge == true (the safe default) UNIONs themes + achievements, takes
//     the MAX per numeric lifetime stat (bests never regress), keeps the slot's local
//     HighestAge/ActiveTheme unless empty (then adopts the blob's). merge == false REPLACEs the
//     DATA fields wholesale. Identity (AccountID/DisplayName/Created) is left as the slot's.
//   - Slot ABSENT: mint a fresh established account carrying the blob's identity
//     (AccountID + DisplayName) and the blob's DATA verbatim, MkdirAll its slot, and Save it in.
//     merge vs replace is moot for a brand-new slot — there is no local data to fold into.
//
// TAMPER: a flagged export flags the account it lands in (merge or replace), and a flagged
// account stays flagged; see Account.Tampered.
//
// ACTIVE-ID DISCIPLINE: Save writes into the account's own slot, so import never touches the
// process-global active account at all. (It used to point the global at the blob's slot for
// the length of the write, and an autosave on the tick goroutine in that window wrote the
// live game and account into the imported slot.) With account A active, importing B's backup
// leaves A active; the caller switches if it wants.
func ImportAccountExport(blob []byte, merge bool) (*Account, error) {
	exp, err := decodeAccountExport(blob)
	if err != nil {
		return nil, err
	}
	return importExportToSlot(exp, merge, nil)
}

// decodeAccountExport unmarshals and verifies an export blob: the integrity half of
// ImportAccountExport, shared with the engine's live-account import. Nothing is written.
func decodeAccountExport(blob []byte) (*progressExport, error) {
	var exp progressExport
	if err := json.Unmarshal(blob, &exp); err != nil {
		return nil, fmt.Errorf("That file is not a readable progress export: %w", err)
	}
	// Verify integrity BEFORE touching the disk. An unsigned blob is rejected (an export is
	// always signed by ExportProgress; a missing sig means it was not produced by us).
	if !hmac.Equal([]byte(exp.Signature), []byte(signProgressExport(&exp))) {
		return nil, fmt.Errorf("That progress export has been changed or damaged since it was made, so it cannot be imported.")
	}
	if exp.AccountID == "" {
		return nil, fmt.Errorf("That progress export has no account ID (it comes from an older version) and cannot be imported.")
	}
	// The ID names the slot the backup lands in, so it must be a real account ID, never a path.
	if !validAccountID(exp.AccountID) {
		return nil, fmt.Errorf("That progress export has an account ID the game cannot use, so it cannot be imported.")
	}
	// The badges it carries must be this account's and as the game signed them. An
	// export is refused whole rather than imported in part.
	if b := exp.BadgeStore; b != nil && (b.AccountID != exp.AccountID || !verifyBadgeFile(b)) {
		return nil, fmt.Errorf("That progress export has been changed or damaged since it was made, so it cannot be imported.")
	}
	return &exp, nil
}

// importExportToSlot lands a verified export in its account's own slot, read from disk:
// merged into (or replacing) the account already there, or as a new account. See
// ImportAccountExport.
//
// With a badge book (the engine's import), the account it lands in is brought up to the
// ruleset's badges before it is saved, so achievements from an older export are badges at
// once. Without one that happens the first time an engine holds the account.
func importExportToSlot(exp *progressExport, merge bool, book *badgeBook) (*Account, error) {
	target, found, err := loadAccountFromSlot(exp.AccountID)
	if err != nil {
		return nil, err
	}
	if !found || target == nil {
		// No slot yet — mint one carrying the blob's identity + DATA verbatim. merge is moot.
		now := time.Now()
		target = &Account{
			Version:      accountSchemaVersion,
			AccountID:    exp.AccountID,
			DisplayName:  exp.DisplayName,
			Created:      now,
			LastSeen:     now,
			Unlocks:      exp.Unlocks,
			Stats:        exp.Stats,
			Achievements: append([]string(nil), exp.Achievements...),
			Prefs:        exp.Prefs,
			Tampered:     exp.Tampered,
		}
		// With a badge file left in the slot, if any: the two are merged.
		target.adoptBadgeFile()
		target.takeBadgeStoreLocked(exp.BadgeStore, true)
	} else {
		target.applyExportLocked(exp, merge) // target is private to this call; no lock needed
	}
	if book != nil {
		target.ensureBadgesLocked(book)
	}
	if err := target.Save(); err != nil {
		return nil, err
	}
	return target, nil
}

// applyExportLocked folds a verified export's DATA into a; identity (AccountID, DisplayName,
// Created) stays a's. merge == true (the safe default) UNIONs themes, achievements and
// badges (the earlier copy of a badge both hold), takes the MAX of each numeric lifetime
// stat and each counter (bests never regress, and nothing is ever summed), takes the later
// HighestAge, and keeps a's active theme and map settings unless empty. merge == false
// REPLACEs the DATA fields wholesale. A flagged export flags a; a flagged a stays flagged.
// Callers hold a.mu when a is shared (the live account).
func (a *Account) applyExportLocked(exp *progressExport, merge bool) {
	a.Tampered = a.Tampered || exp.Tampered
	if !merge {
		a.Unlocks = exp.Unlocks
		a.Stats = exp.Stats
		a.Achievements = append([]string(nil), exp.Achievements...)
		a.takeBadgeStoreLocked(exp.BadgeStore, false)
		a.Prefs = exp.Prefs
		return
	}
	// Union themes (preserve local, add blob's) and achievements.
	a.Unlocks.Themes = unionStrings(a.Unlocks.Themes, exp.Unlocks.Themes)
	a.Achievements = unionStrings(a.Achievements, exp.Achievements)
	// Badges: every badge either copy holds, the earlier of two. Counters: the larger
	// of each, never the sum, so importing your own backup doubles nothing.
	a.takeBadgeStoreLocked(exp.BadgeStore, true)
	// Max each numeric lifetime stat — bests don't regress.
	a.Stats.TotalPrestiges = maxInt(a.Stats.TotalPrestiges, exp.Stats.TotalPrestiges)
	for _, age := range sortedKeys(exp.Stats.PrestigesByAge) {
		if a.Stats.PrestigesByAge == nil {
			a.Stats.PrestigesByAge = map[string]int{}
		}
		a.Stats.PrestigesByAge[age] = maxInt(a.Stats.PrestigesByAge[age], exp.Stats.PrestigesByAge[age])
	}
	a.Stats.CivilizationsStarted = maxInt(a.Stats.CivilizationsStarted, exp.Stats.CivilizationsStarted)
	a.Stats.SavesCompleted = maxInt(a.Stats.SavesCompleted, exp.Stats.SavesCompleted)
	// HighestAge is a key, not a number: the later age of the two wins.
	if ageOrderOf(exp.Stats.HighestAge) > ageOrderOf(a.Stats.HighestAge) {
		a.Stats.HighestAge = exp.Stats.HighestAge
	}
	// Active theme: keep local if set, else adopt the blob's.
	if a.Prefs.ActiveTheme == "" {
		a.Prefs.ActiveTheme = exp.Prefs.ActiveTheme
	}
	// Map settings: the same rule.
	if a.Prefs.MapStyle == "" {
		a.Prefs.MapStyle = exp.Prefs.MapStyle
	}
	if a.Prefs.MapGlyphs == "" {
		a.Prefs.MapGlyphs = exp.Prefs.MapGlyphs
	}
	if a.Prefs.Minimap == "" {
		a.Prefs.Minimap = exp.Prefs.Minimap
	}
	a.Prefs.MapIconsHint = a.Prefs.MapIconsHint || exp.Prefs.MapIconsHint
}

// importExport folds a verified export into this LIVE account and saves it: the engine's
// path for a backup of the account in use. Working on the live object keeps records it has
// not flushed yet, and leaves no second copy on disk for the live one to overwrite at its
// next save. On a Save error the account stays dirty so the next flush retries.
func (a *Account) importExport(exp *progressExport, merge bool, book *badgeBook) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.applyExportLocked(exp, merge)
	if book != nil {
		// What the backup brought (achievements, a higher age, more prestiges) may
		// prove badges the account did not hold: granted without a toast.
		a.ensureBadgesLocked(book)
	}
	if err := a.Save(); err != nil {
		a.dirty = true
		return err
	}
	a.dirty = false
	return nil
}

// unionStrings returns the set union of a and b in a fresh slice, preserving a's order
// then appending any of b not already present. Used to merge themes/achievements so an
// import never drops an entry either side already holds.
func unionStrings(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// maxInt returns the larger of a and b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- Lifetime stats + achievements (the accounts design §3.3 / §8 / §9 Phase 6) ---
//
// Lifetime stats are CROSS-SAVE aggregates living on the ACCOUNT — distinct from the
// per-save ge.Stats system, which resets on every new game/prestige. The account's badges
// (account_badges.go) are judged from the engine's reports, not by these hooks.
//
// LOCKING DISCIPLINE (the load-bearing constraint, the accounts design §8 + project rule):
// the engine calls RecordPrestige/RecordAgeReached from UNDER the engine write lock
// (ge.mu — advanceAge and DoPrestige both hold it). Those methods therefore MUST be
// in-memory only: they take the account's OWN mutex (a.mu, fully independent of ge.mu),
// do NO file I/O, and NEVER call back into ge.* (AddLog/GetState/…), or the non-reentrant
// ge.mu would deadlock. The actual Save happens later via FlushIfDirty, called from the
// engine's periodic-autosave block which runs OUTSIDE ge.mu (the accounts design §8 write cadence).

// RecordPrestige increments the lifetime prestige count and marks the account dirty for
// the next flush (the accounts design §8). It is IN-MEMORY ONLY: it takes a.mu, performs no
// file I/O, and never calls back into the engine — so it is safe to call from DoPrestige
// while the engine write lock is held. The write is deferred to FlushIfDirty (engine
// autosave block, outside ge.mu). The prestige badges are judged from the engine's own
// report of the prestige (badges.go), not here.
func (a *Account) RecordPrestige() {
	a.RecordPrestigeFrom("")
}

// RecordPrestigeFrom is RecordPrestige for a prestige made from age, which it
// also counts in PrestigesByAge ("" counts nowhere but the total). Same
// IN-MEMORY-ONLY discipline as RecordPrestige.
func (a *Account) RecordPrestigeFrom(age string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Stats.TotalPrestiges++
	if age != "" {
		if a.Stats.PrestigesByAge == nil {
			a.Stats.PrestigesByAge = map[string]int{}
		}
		a.Stats.PrestigesByAge[age]++
	}
	a.dirty = true
}

// RecordCivilizationStarted counts one more civilization started on the account: a run
// began. IN-MEMORY ONLY (same discipline as RecordPrestige): the engine calls it under its
// write lock when a new game starts and when a prestige or a Succumb starts the next run.
func (a *Account) RecordCivilizationStarted() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Stats.CivilizationsStarted++
	a.dirty = true
}

// RecordAgeReached records that the account's player has reached the given age, lifting
// HighestAge only when ageOrder exceeds the order of the currently-stored highest age (so
// a lower age never regresses the lifetime best), and marks the account dirty (the accounts
// design §8). IN-MEMORY ONLY (same discipline as RecordPrestige). The age badges are judged
// from the engine's report of the advance (badges.go).
//
// DEVIATION from the accounts design §8: the doc sketches RecordAgeReached(ageKey) with no order.
// The account stores only the highest age KEY (AccountStats.HighestAge, the accounts design §3.3),
// and a bare key can't be ranked without consulting the age table — which would couple the
// account to config and re-derive order on every call. We take ageOrder explicitly from the
// engine (which already knows it) so the comparison is a cheap int compare and the account
// stays config-free. The stored highest order is derived from HighestAge on demand via
// highestOrderLocked, so it needs no extra persisted field.
func (a *Account) RecordAgeReached(ageKey string, ageOrder int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if ageOrder > a.highestOrderLocked() {
		a.Stats.HighestAge = ageKey
		a.dirty = true
	}
}

// highestOrderLocked returns the Order of the currently-stored HighestAge, or -1 if none
// is set (or the stored key is unknown — treated as "below everything" so any real age
// wins). Callers must hold a.mu. It consults the pure config age table (no locks).
func (a *Account) highestOrderLocked() int { return ageOrderOf(a.Stats.HighestAge) }

// ageOrderOf is the Order of an age key in the config age table, or -1 for "" or a key
// it does not have (below every real age). Pure config: no locks.
func ageOrderOf(age string) int {
	if age == "" {
		return -1
	}
	if def, ok := config.AgeByKey()[age]; ok {
		return def.Order
	}
	return -1
}

// FlushIfDirty persists the account once if the in-memory stats/achievements have changed
// since the last write, then clears the dirty flag (the accounts design §8 write cadence). It is the
// ONLY place lifetime-stat changes touch the disk, and it MUST be called from OUTSIDE the
// engine write lock (the autosave block / clean-exit path) — never from a Record* call site.
//
// Self-deadlock avoidance: Save() does NOT acquire a.mu itself (the theme mutators call
// a.Save() while already holding a.mu, by design). So FlushIfDirty can hold a.mu across the
// Save() call without re-entering the same non-reentrant mutex. If dirty is false it is a
// pure no-op (no Save, no error). On a Save error the dirty flag is left SET so the next
// flush retries rather than silently dropping the unsaved progress.
func (a *Account) FlushIfDirty() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.dirty {
		return nil
	}
	if err := a.Save(); err != nil {
		return err // keep dirty set: retry on the next flush
	}
	a.dirty = false
	return nil
}

// Name returns the account's display name under a.mu, for the lock-safe UI snapshot
// (GetState). DisplayName is chosen-once identity and never mutated by the Record*
// writers, but reading it under a.mu keeps the snapshot consistent with the rest of
// LifetimeStats and free of any data-race tooling complaint.
func (a *Account) Name() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.DisplayName
}

// LifetimeStats returns a lock-guarded COPY of the account's lifetime stats and the keys
// of the badges it holds (sorted), for the UI snapshot (GetState). Nothing it returns is
// the account's own, so a snapshot consumer can neither mutate account state nor race the
// Record* writers.
func (a *Account) LifetimeStats() (AccountStats, []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	st := a.Stats
	st.PrestigesByAge = maps.Clone(a.Stats.PrestigesByAge) // no alias of the live map
	return st, sortedKeys(a.Badges)
}
