package game

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
)

// accountSchemaVersion is the on-disk schema version of account.json. Bumped only
// for migrations; readers default unknown/older fields to zero (accounts.md §3.3).
const accountSchemaVersion = 1

// accountFileName is the fixed base name of the account file under the data dir.
// v1 resolves a single account.json — no multi-profile layout yet (accounts.md §3.1).
const accountFileName = "account.json"

// AccountUnlocks holds account-wide cosmetic unlocks (DATA, accounts.md §3.3).
// Phase 1 only round-trips these; the unlock API lands in Phase 3.
type AccountUnlocks struct {
	Themes []string `json:"themes,omitempty"`
}

// AccountStats holds lifetime, cross-save aggregates (DATA, accounts.md §3.3).
// Phase 1 only round-trips these; the engine hooks land in Phase 6.
type AccountStats struct {
	TotalPrestiges       int    `json:"total_prestiges,omitempty"`
	HighestAge           string `json:"highest_age,omitempty"`
	CivilizationsStarted int    `json:"civilizations_started,omitempty"`
	SavesCompleted       int    `json:"saves_completed,omitempty"`
}

// AccountPrefs holds preferences that travel with the account (accounts.md §3.3).
// Phase 1 only round-trips these; SetActiveTheme et al. land in Phase 3.
type AccountPrefs struct {
	ActiveTheme string `json:"active_theme,omitempty"`
}

// Account is the per-player identity + meta-progression record, persisted to
// data/account.json. It is the single source of truth for identity (account ID,
// display name) and DATA (unlocks, lifetime stats, achievements, prefs).
//
// Integrity mirrors saves: Signature is the HMAC-SHA256 of the payload (with
// Signature zeroed) under saveHMACKey. A tampered file still loads — Tampered is
// set instead, the cosmetic analogue of the save CheaterBadge (accounts.md §3.4).
// The schema follows accounts.md §3.3; all DATA fields are present so the file
// round-trips, but the unlock/stats/prefs APIs that mutate them arrive in later
// phases.
type Account struct {
	Version     int       `json:"version"`
	AccountID   string    `json:"account_id"`
	DisplayName string    `json:"display_name,omitempty"`
	Created     time.Time `json:"created"`
	LastSeen    time.Time `json:"last_seen,omitempty"`

	// --- meta-progression (DATA) ---
	Unlocks      AccountUnlocks `json:"unlocks,omitempty"`
	Stats        AccountStats   `json:"stats,omitempty"`
	Achievements []string       `json:"achievements,omitempty"`

	// --- preferences (travel with the account) ---
	Prefs AccountPrefs `json:"prefs,omitempty"`

	// --- integrity (same scheme as saves) ---
	Signature string `json:"_sig,omitempty"`

	// Tampered is the in-memory, non-persisted cosmetic flag set when a loaded
	// file's signature does not match (accounts.md §3.4). It mirrors the save
	// CheaterBadge: signalling, not a lockout. json:"-" keeps it off disk.
	Tampered bool `json:"-"`

	// FreshlyCreated is the in-memory, non-persisted signal that LoadOrCreate just
	// minted this account on its fresh-create path (no file existed). Boot code reads
	// it to surface a one-time, non-blocking first-run notice (accounts.md §6) without
	// changing LoadOrCreate's signature. json:"-" keeps it off disk; it is false for
	// any account loaded from an existing file.
	FreshlyCreated bool `json:"-"`

	// dirty is the in-memory, non-persisted write-debounce flag for the lifetime-stats
	// hooks (accounts.md §8 "debounce writes"). RecordPrestige/RecordAgeReached mutate
	// the in-memory stats and set dirty=true WITHOUT touching the disk — they run under
	// the engine write lock (advanceAge/DoPrestige) where file I/O is forbidden. The
	// engine's periodic autosave block (outside ge.mu) calls FlushIfDirty, which Saves
	// once if dirty and clears the flag. json:"-" keeps it off disk and out of the sig.
	dirty bool `json:"-"`

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
// hex-encoded (accounts.md §3.3 — 128 random bits, not a v4 UUID specifically).
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
		// prompts for a fresh name (accounts.md §7).
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
	return &acct, true, nil
}

// CreateNamedAccount mints a signed account whose identity is derived from name:
// AccountID = accountIDFromName(name), DisplayName = the trimmed ORIGINAL name (display
// keeps casing/whitespace; the ID does not). Re-entering the same name reproduces the
// same ID — that IS the cross-machine recovery story (accounts.md §3.5).
//
// MIGRATION: if an account file already exists (e.g. a legacy random-id dev account from
// the old LoadOrCreate auto-create), its DATA — unlocks, lifetime stats, achievements,
// prefs — is carried over into the new named account so earned progress is not lost when
// the identity is re-keyed to the name. Identity fields (the old random ID, Created) are
// replaced; the new named identity wins. A corrupt/unparseable existing file is ignored
// (LoadAccount backs it up) and the named account starts with empty data.
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

	// Carry over DATA from any pre-existing account so unlocks/stats survive the
	// identity re-key. found=false (absent or corrupt) → start clean.
	if prior, found, err := LoadAccount(); err == nil && found && prior != nil {
		acct.Unlocks = prior.Unlocks
		acct.Stats = prior.Stats
		acct.Achievements = append([]string(nil), prior.Achievements...)
		acct.Prefs = prior.Prefs
	}

	if err := acct.Save(); err != nil {
		return nil, err
	}
	return acct, nil
}

// accountPath resolves the full path to account.json, mirroring savePath's
// resolution: prefer the binary-relative data/account.json; if that is absent,
// fall back to a CWD-relative data/account.json (legacy/dev-run); default to the
// binary-relative path for new writes (accounts.md §3.1).
func accountPath() string {
	primary := filepath.Join(dataDirectory(), accountFileName)
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	legacy := filepath.Join("data", accountFileName)
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return primary // default to canonical for new writes
}

// WipeAccount permanently deletes the account file (account.json) and its
// corrupt-backup sibling (account.json.corrupt) if present, so the next boot starts
// from a clean slate and re-prompts for a name. It is the account analogue of
// WipeAllSaves — destructive and irreversible (no server backup).
//
// SCOPE: it removes ONLY the account identity + meta-progression file(s). It NEVER
// touches data/saves (game saves are separate and survive a wipe) nor any progress
// export file (account-export.json) — those are independent backups the player may
// have deliberately created. A missing file is not an error: os.Remove's
// "not exist" is ignored so wiping twice (or with no account) is a clean no-op.
func WipeAccount() error {
	path := accountPath()
	for _, p := range []string{path, path + ".corrupt"} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to wipe account file %s: %w", p, err)
		}
	}
	return nil
}

// signAccount returns the HMAC-SHA256 hex of the account payload with Signature
// zeroed, so the signature covers the data only — identical construction to
// signSave, sharing the hmacSign helper (accounts.md §3.4).
//
// It takes a pointer (not a value) so the sync.Mutex field is never copied — a
// value copy would trip go vet's copylocks. The signed bytes are unchanged from a
// value-based marshal: json.Marshal already ignores unexported fields (mu) and the
// json:"-" Tampered/FreshlyCreated fields, and we build the payload from a
// freshly-constructed struct literal that copies only the serializable fields,
// with Signature/Tampered zeroed — so the signature covers exactly the on-disk data.
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
		// Signature deliberately zero; Tampered/FreshlyCreated/mu are json:"-"/unexported.
	}
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

// verifyAccount reports whether a's stored Signature matches a freshly computed
// one. An unsigned file (empty Signature) is treated as legacy/benign-valid, the
// same benefit-of-the-doubt verifySave grants unsigned saves (accounts.md §3.4).
func verifyAccount(a *Account) bool {
	if a.Signature == "" {
		return true
	}
	return hmac.Equal([]byte(a.Signature), []byte(signAccount(a)))
}

// Save signs the account with the shared HMAC helper and writes it atomically
// (temp file + os.Rename), creating the data dir if missing. Mirrors SaveGame's
// write discipline (accounts.md §3.4).
func (a *Account) Save() error {
	dir := dataDirectory()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	a.Signature = signAccount(a)
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	// Write to the resolved path (canonical, or the legacy CWD location if that is
	// where the existing file lives) so we don't orphan a dev-run account.json.
	path := accountPath()
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write account: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to finalize account: %w", err)
	}
	return nil
}

// LoadOrCreate resolves data/account.json and returns the live Account, creating
// one transparently when needed (accounts.md §6/§7). Behavior by file state:
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
	path := accountPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			acct, err := newAccount()
			if err != nil {
				return nil, err
			}
			if err := acct.Save(); err != nil {
				return nil, err
			}
			// First genuine run: no file existed. Flag it so boot code can surface a
			// one-time non-blocking notice (accounts.md §6). The corrupt-recovery path
			// below is NOT a first run — it leaves FreshlyCreated false.
			acct.FreshlyCreated = true
			return acct, nil
		}
		return nil, fmt.Errorf("failed to read account: %w", err)
	}

	var acct Account
	if err := json.Unmarshal(data, &acct); err != nil {
		// Corrupt / unparseable → back it up, then start fresh (accounts.md §7).
		backup := path + ".corrupt"
		if renameErr := os.Rename(path, backup); renameErr != nil {
			return nil, fmt.Errorf("account file is corrupt and could not be backed up: %w", renameErr)
		}
		fresh, newErr := newAccount()
		if newErr != nil {
			return nil, newErr
		}
		if saveErr := fresh.Save(); saveErr != nil {
			return nil, saveErr
		}
		return fresh, nil
	}

	if !verifyAccount(&acct) {
		// Signature present but mismatched → tampered. Flag it, still load it.
		acct.Tampered = true
	}
	return &acct, nil
}

// --- Unlock API (accounts.md §8; theming.md §5 is the first caller) ---
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
// since the in-memory set did change. theming.md fires the unlock toast only on
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

// --- Recovery code (identity backup, accounts.md §3.5 / §8 / §9 Phase 4) ---
//
// The recovery code encodes IDENTITY ONLY — the 16-byte account_id plus a 2-byte
// checksum — into a short, dash-grouped, uppercase, Crockford-base32 string with an
// `AGEF-` prefix (e.g. AGEF-7Q2K-9X4M-ZJ31-...). It restores who you are across
// machines/reinstalls; it does NOT restore earned progress (unlocks/stats) — that is
// DATA, backed up separately via export/import (Phase 5). The code is a convenience
// identifier, not a credential (accounts.md §3.5): the checksum guards against TYPOS,
// not forgery, and account state is cosmetic, not security-critical.

// recoveryCodePrefix is the human-readable namespace stamped on every recovery code.
const recoveryCodePrefix = "AGEF"

// crockfordAlphabet is the Crockford base32 symbol set: 0-9 then A-Z excluding the
// ambiguous I, L, O, U. Index = 5-bit value; the string is the canonical encoder.
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// crc16CCITT computes the CRC-16/CCITT-FALSE checksum of data: 16-bit, polynomial
// 0x1021, init 0xFFFF, no reflection, no final XOR. It is a small, standard,
// dependency-free typo guard for the recovery code (accounts.md §3.5).
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
			return nil, fmt.Errorf("invalid recovery code (bad character %q)", string(s[i]))
		}
		buf = (buf << 5) | uint32(v)
		bits += 5
		if bits >= 8 {
			bits -= 8
			out = append(out, byte((buf>>uint(bits))&0xFF))
		}
	}
	if len(out) < byteLen {
		return nil, fmt.Errorf("invalid recovery code (too short)")
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

// RecoveryCode returns this account's identity-recovery code (accounts.md §3.5/§8):
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

// ImportRecoveryCode decodes a recovery code, verifies its checksum, and writes a
// FRESH signed account.json carrying the recovered account ID with EMPTY data —
// identity only, never restoring unlocks/stats (accounts.md §3.5/§8). It reuses the
// normal signed atomic Save path; it never hand-rolls a second integrity scheme.
//
// Input is normalized leniently: uppercased, the AGEF- prefix stripped, dashes and
// spaces removed, and ambiguous Crockford characters mapped (I/L→1, O→0, U→V). A
// checksum mismatch returns a clear typo-guard error rather than silently minting a
// wrong account.
func ImportRecoveryCode(code string) (*Account, error) {
	// Normalize: drop spaces, uppercase, strip the AGEF- prefix, strip dashes.
	norm := strings.ToUpper(strings.TrimSpace(code))
	norm = strings.ReplaceAll(norm, " ", "")
	norm = strings.ReplaceAll(norm, "-", "")
	if p := strings.ToUpper(recoveryCodePrefix); strings.HasPrefix(norm, p) {
		norm = norm[len(p):]
	}
	if norm == "" {
		return nil, fmt.Errorf("invalid recovery code (empty)")
	}

	payload, err := crockfordDecode(norm, 18)
	if err != nil {
		return nil, err
	}
	idBytes := payload[:16]
	gotSum := uint16(payload[16])<<8 | uint16(payload[17])
	if gotSum != crc16CCITT(idBytes) {
		return nil, fmt.Errorf("invalid recovery code (checksum failed)")
	}

	now := time.Now()
	acct := &Account{
		Version:   accountSchemaVersion,
		AccountID: hex.EncodeToString(idBytes),
		Created:   now,
		LastSeen:  now,
		// EMPTY data: identity only. Unlocks/Stats/Achievements/Prefs stay zero —
		// progress is carried by export/import (Phase 5), not the recovery code.
	}
	if err := acct.Save(); err != nil {
		return nil, err
	}
	return acct, nil
}

// --- Progress export / import (DATA backup, accounts.md §3.6 / §8 / §9 Phase 5) ---
//
// Distinct from the recovery code: the recovery code carries IDENTITY only (account
// id), while an export carries PROGRESS only (unlocks, lifetime stats, achievements,
// prefs) and NO identity. With no server, DATA cannot be reconstituted from nothing,
// so export is the explicit one-action backup of earned progress (accounts.md §3.6).
//
// DEVIATION from accounts.md §8: the doc sketches a package-level `ImportProgress(blob,
// merge)`. We implement ImportProgress as a METHOD on the live *Account instead. The
// engine holds this exact pointer (SetAccount, Phase 2), so mutating in place + Save
// makes engine.Account() reflect the import immediately, with no re-wiring. A package
// function would mint a new Account the engine never sees. The export side matches the
// doc's `(a *Account) ExportProgress()` signature as written.

// progressExport is the self-describing on-disk shape of an export blob: a format
// version, the DATA fields ONLY (no AccountID / identity), and an HMAC signature over
// the sig-zeroed payload (same scheme as Account.Save — reuse, don't reinvent). The
// stats/achievements fields ride along now so Phase 6 data exports without a format bump.
type progressExport struct {
	Version      int            `json:"version"`
	Unlocks      AccountUnlocks `json:"unlocks,omitempty"`
	Stats        AccountStats   `json:"stats,omitempty"`
	Achievements []string       `json:"achievements,omitempty"`
	Prefs        AccountPrefs   `json:"prefs,omitempty"`
	Signature    string         `json:"_sig,omitempty"`
}

// signProgressExport returns the HMAC-SHA256 hex of the export payload with Signature
// zeroed, under saveHMACKey — identical construction to signAccount/signSave.
func signProgressExport(p *progressExport) string {
	payload := progressExport{
		Version:      p.Version,
		Unlocks:      p.Unlocks,
		Stats:        p.Stats,
		Achievements: p.Achievements,
		Prefs:        p.Prefs,
		// Signature deliberately zero.
	}
	data, _ := json.Marshal(&payload)
	return hmacSign(data, saveHMACKey)
}

// ExportProgress builds a signed export blob carrying this account's DATA only —
// unlocks, lifetime stats, achievements, prefs — and NOT its identity (accounts.md
// §3.6). It signs the blob with the shared HMAC helper (same zero-sig-then-marshal
// pattern as Save) and returns pretty JSON bytes the caller writes to a file/clipboard.
func (a *Account) ExportProgress() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	exp := progressExport{
		Version:      accountSchemaVersion,
		Unlocks:      a.Unlocks,
		Stats:        a.Stats,
		Achievements: a.Achievements,
		Prefs:        a.Prefs,
	}
	exp.Signature = signProgressExport(&exp)

	data, err := json.MarshalIndent(&exp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal progress export: %w", err)
	}
	return data, nil
}

// ImportProgress unmarshals and verifies a progress export blob, then folds its DATA
// into the live account and persists it (accounts.md §3.6). It NEVER touches identity
// (AccountID/DisplayName/Created) — that is the recovery code's job.
//
// The signature is verified first: a recomputed HMAC over the sig-zeroed payload must
// match the blob's _sig, otherwise the blob is corrupt or tampered and we return an
// error WITHOUT mutating the account — we never import data that failed integrity.
//
// merge == true (the safe default): UNION the data into the current account —
//   - unlocked themes: union (never drop a theme the local account already has; add
//     any the blob carries) — so import(merge) restores wiped unlocks AND does not drop
//     newer local unlocks the backup predates.
//   - achievements: union.
//   - lifetime stats: take the MAX per numeric stat (lifetime bests never regress);
//     HighestAge keeps the local value unless empty, then adopts the blob's.
//   - active theme: keep the local one if set, else adopt the blob's.
//
// merge == false: REPLACE the DATA fields wholesale with the blob's.
//
// Either way it Saves through the normal signed atomic path. Because the engine holds
// this exact *Account, the import is visible via engine.Account() immediately.
func (a *Account) ImportProgress(blob []byte, merge bool) error {
	var exp progressExport
	if err := json.Unmarshal(blob, &exp); err != nil {
		return fmt.Errorf("invalid or corrupt progress export: %w", err)
	}
	// Verify integrity BEFORE mutating anything. An unsigned blob is rejected here
	// (unlike account.json's benign-unsigned-legacy rule): an export is always written
	// signed by ExportProgress, so a missing/empty sig means it was not produced by us.
	if !hmac.Equal([]byte(exp.Signature), []byte(signProgressExport(&exp))) {
		return fmt.Errorf("invalid or corrupt progress export (signature mismatch)")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !merge {
		// Wholesale replace of the DATA fields; identity is left untouched.
		a.Unlocks = exp.Unlocks
		a.Stats = exp.Stats
		a.Achievements = append([]string(nil), exp.Achievements...)
		a.Prefs = exp.Prefs
		return a.Save()
	}

	// Merge: union themes (preserve local, add blob's).
	a.Unlocks.Themes = unionStrings(a.Unlocks.Themes, exp.Unlocks.Themes)
	// Union achievements.
	a.Achievements = unionStrings(a.Achievements, exp.Achievements)
	// Max each numeric lifetime stat — bests don't regress.
	a.Stats.TotalPrestiges = maxInt(a.Stats.TotalPrestiges, exp.Stats.TotalPrestiges)
	a.Stats.CivilizationsStarted = maxInt(a.Stats.CivilizationsStarted, exp.Stats.CivilizationsStarted)
	a.Stats.SavesCompleted = maxInt(a.Stats.SavesCompleted, exp.Stats.SavesCompleted)
	// HighestAge is a key, not a number: keep local unless empty, then adopt blob's.
	if a.Stats.HighestAge == "" {
		a.Stats.HighestAge = exp.Stats.HighestAge
	}
	// Active theme: keep local if set, else adopt the blob's.
	if a.Prefs.ActiveTheme == "" {
		a.Prefs.ActiveTheme = exp.Prefs.ActiveTheme
	}
	return a.Save()
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

// --- Lifetime stats + achievements (accounts.md §3.3 / §8 / §9 Phase 6) ---
//
// Lifetime stats are CROSS-SAVE aggregates living on the ACCOUNT — distinct from the
// per-save ge.Stats system, which resets on every new game/prestige. Achievements are
// one-time, account-wide unlock keys appended on first satisfaction and never removed.
//
// LOCKING DISCIPLINE (the load-bearing constraint, accounts.md §8 + project rule):
// the engine calls RecordPrestige/RecordAgeReached from UNDER the engine write lock
// (ge.mu — advanceAge and DoPrestige both hold it). Those methods therefore MUST be
// in-memory only: they take the account's OWN mutex (a.mu, fully independent of ge.mu),
// do NO file I/O, and NEVER call back into ge.* (AddLog/GetState/…), or the non-reentrant
// ge.mu would deadlock. The actual Save happens later via FlushIfDirty, called from the
// engine's periodic-autosave block which runs OUTSIDE ge.mu (accounts.md §8 write cadence).

// accountAchievement is one entry in the in-file achievement table: a stable unlock key
// plus a human-readable name and a predicate over the lifetime stats. The predicate is
// pure (no locks, no I/O) and is evaluated by recordEvaluateLocked while a.mu is held.
type accountAchievement struct {
	Key  string
	Name string
	// met reports whether the given stats satisfy this achievement. ageOrder is the
	// order of the age that triggered the current evaluation (-1 when the trigger was
	// a prestige, not an age-up), so age achievements can key off the just-reached age.
	met func(s AccountStats, ageOrder int) bool
}

// achievementAge* are the age Order thresholds the age achievements fire at, named so
// the table reads as intent rather than magic numbers (config/ages.go: Order is 0-indexed;
// iron_age = 3, modern_age = 12). Kept here, not imported from config, to keep the table
// dependency-free and the predicate pure.
const (
	achievementOrderIron   = 3  // iron_age
	achievementOrderModern = 12 // modern_age
)

// accountAchievements is the small, sensible achievement set for Phase 6. Two prestige
// tiers and two age milestones — enough to prove the wiring without a sprawling table.
// New entries are purely additive (the key is the only persisted artifact).
var accountAchievements = []accountAchievement{
	{Key: "first_prestige", Name: "First Prestige", met: func(s AccountStats, _ int) bool {
		return s.TotalPrestiges >= 1
	}},
	{Key: "prestige_x10", Name: "Serial Reincarnator", met: func(s AccountStats, _ int) bool {
		return s.TotalPrestiges >= 10
	}},
	// Age achievements: the trigger's ageOrder must be at/above the threshold. We gate on
	// the live trigger order (not HighestAge) so a single RecordAgeReached call evaluates
	// only the age just reached; HighestAge has already been updated to that age by then,
	// so a later re-eval would still hold, but the trigger gate keeps each unlock crisp.
	{Key: "reached_iron", Name: "Age of Iron", met: func(_ AccountStats, ageOrder int) bool {
		return ageOrder >= achievementOrderIron
	}},
	{Key: "reached_modern", Name: "Into the Modern Age", met: func(_ AccountStats, ageOrder int) bool {
		return ageOrder >= achievementOrderModern
	}},
}

// AchievementName returns the human-readable name for an achievement key, or the key
// itself if it is unknown (so the UI degrades gracefully on a future/renamed key).
func AchievementName(key string) string {
	for _, a := range accountAchievements {
		if a.Key == key {
			return a.Name
		}
	}
	return key
}

// recordEvaluateLocked evaluates the achievement table against the current stats and
// appends any newly-satisfied keys to a.Achievements (deduped). Callers MUST hold a.mu.
// ageOrder is the order of the age that triggered this evaluation, or -1 for a prestige
// trigger. It performs NO Save — the caller sets a.dirty and the flush persists later.
func (a *Account) recordEvaluateLocked(ageOrder int) {
	for _, def := range accountAchievements {
		if a.hasAchievementLocked(def.Key) {
			continue
		}
		if def.met(a.Stats, ageOrder) {
			a.Achievements = append(a.Achievements, def.Key)
		}
	}
}

// hasAchievementLocked reports whether key is already unlocked. Callers must hold a.mu.
func (a *Account) hasAchievementLocked(key string) bool {
	for _, k := range a.Achievements {
		if k == key {
			return true
		}
	}
	return false
}

// RecordPrestige increments the lifetime prestige count, evaluates prestige achievements,
// and marks the account dirty for the next flush (accounts.md §8). It is IN-MEMORY ONLY:
// it takes a.mu, performs no file I/O, and never calls back into the engine — so it is
// safe to call from DoPrestige while the engine write lock is held. The write is deferred
// to FlushIfDirty (engine autosave block, outside ge.mu).
func (a *Account) RecordPrestige() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Stats.TotalPrestiges++
	a.recordEvaluateLocked(-1) // -1: prestige trigger, not an age-up
	a.dirty = true
}

// RecordAgeReached records that the account's player has reached the given age, lifting
// HighestAge only when ageOrder exceeds the order of the currently-stored highest age (so
// a lower age never regresses the lifetime best), then evaluates age achievements and marks
// the account dirty (accounts.md §8). IN-MEMORY ONLY (same discipline as RecordPrestige).
//
// DEVIATION from accounts.md §8: the doc sketches RecordAgeReached(ageKey) with no order.
// The account stores only the highest age KEY (AccountStats.HighestAge, accounts.md §3.3),
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
	}
	a.recordEvaluateLocked(ageOrder)
	a.dirty = true
}

// highestOrderLocked returns the Order of the currently-stored HighestAge, or -1 if none
// is set (or the stored key is unknown — treated as "below everything" so any real age
// wins). Callers must hold a.mu. It consults the pure config age table (no locks).
func (a *Account) highestOrderLocked() int {
	if a.Stats.HighestAge == "" {
		return -1
	}
	if def, ok := config.AgeByKey()[a.Stats.HighestAge]; ok {
		return def.Order
	}
	return -1
}

// FlushIfDirty persists the account once if the in-memory stats/achievements have changed
// since the last write, then clears the dirty flag (accounts.md §8 write cadence). It is the
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

// LifetimeStats returns a lock-guarded COPY of the account's lifetime stats and unlocked
// achievement keys, for the UI snapshot (GetState). It never exposes the account's backing
// slice — the returned achievements slice is freshly allocated — so a snapshot consumer can
// neither mutate account state nor race the Record* writers.
func (a *Account) LifetimeStats() (AccountStats, []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ach := make([]string, len(a.Achievements))
	copy(ach, a.Achievements)
	return a.Stats, ach
}
