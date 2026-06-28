package game

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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

// signAccount returns the HMAC-SHA256 hex of the account payload with Signature
// zeroed, so the signature covers the data only — identical construction to
// signSave, sharing the hmacSign helper (accounts.md §3.4).
func signAccount(a Account) string {
	a.Signature = ""
	a.Tampered = false // json:"-" already excludes it, but keep the payload pristine
	data, _ := json.Marshal(a)
	return hmacSign(data, saveHMACKey)
}

// verifyAccount reports whether a's stored Signature matches a freshly computed
// one. An unsigned file (empty Signature) is treated as legacy/benign-valid, the
// same benefit-of-the-doubt verifySave grants unsigned saves (accounts.md §3.4).
func verifyAccount(a *Account) bool {
	if a.Signature == "" {
		return true
	}
	return hmac.Equal([]byte(a.Signature), []byte(signAccount(*a)))
}

// Save signs the account with the shared HMAC helper and writes it atomically
// (temp file + os.Rename), creating the data dir if missing. Mirrors SaveGame's
// write discipline (accounts.md §3.4).
func (a *Account) Save() error {
	dir := dataDirectory()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	a.Signature = signAccount(*a)
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
