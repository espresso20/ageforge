package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolateAccountDir points the data root at an isolated temp directory for the
// duration of the test, so we never read or clobber a real ./data/account.json.
//
// It sets the package-level dataDirOverride test seam (see save.go) to a fresh
// t.TempDir() and restores the prior value on cleanup. With the override set,
// dataDirectory() — and therefore accountPath(), Save(), and LoadOrCreate() —
// resolve entirely inside the temp dir, deterministically and without depending on
// CWD or the test binary's location. Go runs tests in a package sequentially unless
// t.Parallel() is called (none here do), so the shared override is safe.
func isolateAccountDir(t *testing.T) string {
	t.Helper()
	prior := dataDirOverride
	tmp := t.TempDir()
	dataDirOverride = tmp
	t.Cleanup(func() {
		dataDirOverride = prior
	})
	return tmp
}

// accountFilePath returns the account.json path inside the isolated data root.
func accountFilePath() string {
	return filepath.Join(dataDirOverride, accountFileName)
}

// TestLoadOrCreateCreatesAndSigns covers the absent-file path: a fresh account is
// created, signed, persisted, and re-loadable with a valid signature.
func TestLoadOrCreateCreatesAndSigns(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (create): %v", err)
	}
	if acct.AccountID == "" {
		t.Fatal("created account has empty AccountID")
	}
	if len(acct.AccountID) != 32 { // 16 bytes hex-encoded
		t.Errorf("AccountID = %q (len %d), want 32 hex chars", acct.AccountID, len(acct.AccountID))
	}
	if acct.Version != accountSchemaVersion {
		t.Errorf("Version = %d, want %d", acct.Version, accountSchemaVersion)
	}
	if acct.Created.IsZero() {
		t.Error("Created timestamp is zero")
	}
	if acct.Signature == "" {
		t.Error("created account was not signed (empty Signature)")
	}
	if acct.Tampered {
		t.Error("freshly created account flagged Tampered")
	}

	// File must exist on disk and verify.
	if _, err := os.Stat(accountFilePath()); err != nil {
		t.Fatalf("account.json not written: %v", err)
	}
	if !verifyAccount(acct) {
		t.Error("freshly created account fails signature verification")
	}
}

// TestLoadOrCreateRoundTrips covers create → load: a second LoadOrCreate returns
// the same account (same ID), with a valid signature and no tamper flag.
func TestLoadOrCreateRoundTrips(t *testing.T) {
	isolateAccountDir(t)

	first, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (first): %v", err)
	}
	second, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (second): %v", err)
	}
	if first.AccountID != second.AccountID {
		t.Errorf("ID changed across load: %q != %q", first.AccountID, second.AccountID)
	}
	if second.Tampered {
		t.Error("round-tripped account flagged Tampered")
	}
	if !verifyAccount(second) {
		t.Error("round-tripped account fails signature verification")
	}
}

// TestSaveRoundTripsDataFields confirms the DATA fields persist and reload, and the
// signature still verifies after a Save that mutates them.
func TestSaveRoundTripsDataFields(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	id := acct.AccountID
	acct.DisplayName = "Adam"
	acct.Unlocks.Themes = []string{"amber_crt", "obsidian"}
	acct.Stats.TotalPrestiges = 14
	acct.Stats.HighestAge = "bronze_age"
	acct.Achievements = []string{"first_prestige"}
	acct.Prefs.ActiveTheme = "obsidian"
	if err := acct.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (reload): %v", err)
	}
	if reloaded.AccountID != id {
		t.Errorf("ID changed: %q != %q", reloaded.AccountID, id)
	}
	if reloaded.DisplayName != "Adam" {
		t.Errorf("DisplayName = %q, want Adam", reloaded.DisplayName)
	}
	if len(reloaded.Unlocks.Themes) != 2 || reloaded.Unlocks.Themes[0] != "amber_crt" {
		t.Errorf("Unlocks.Themes = %v, want [amber_crt obsidian]", reloaded.Unlocks.Themes)
	}
	if reloaded.Stats.TotalPrestiges != 14 || reloaded.Stats.HighestAge != "bronze_age" {
		t.Errorf("Stats not round-tripped: %+v", reloaded.Stats)
	}
	if reloaded.Prefs.ActiveTheme != "obsidian" {
		t.Errorf("Prefs.ActiveTheme = %q, want obsidian", reloaded.Prefs.ActiveTheme)
	}
	if reloaded.Tampered {
		t.Error("reloaded account flagged Tampered after legitimate Save")
	}
	if !verifyAccount(reloaded) {
		t.Error("reloaded account fails signature verification after legitimate Save")
	}
}

// TestLoadOrCreateTamperedSetsFlag covers the tampered path: a field is changed on
// disk WITHOUT re-signing, so the signature no longer matches. LoadOrCreate must
// still return the account but with Tampered=true, and must NOT delete the file.
func TestLoadOrCreateTamperedSetsFlag(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (create): %v", err)
	}
	id := acct.AccountID

	// Read the signed file, mutate a field without re-signing, write it back.
	path := accountFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatalf("unmarshal account: %v", err)
	}
	onDisk["display_name"] = "intruder" // changes payload, stale _sig now mismatches
	tampered, err := json.Marshal(onDisk)
	if err != nil {
		t.Fatalf("re-marshal tampered: %v", err)
	}
	if err := os.WriteFile(path, tampered, 0644); err != nil {
		t.Fatalf("write tampered account: %v", err)
	}

	loaded, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (tampered): %v", err)
	}
	if !loaded.Tampered {
		t.Error("tampered account not flagged Tampered")
	}
	if loaded.AccountID != id {
		t.Errorf("tampered account ID changed: %q != %q (file should not be recreated)", loaded.AccountID, id)
	}
	if loaded.DisplayName != "intruder" {
		t.Errorf("tampered account loaded with DisplayName = %q, want intruder (still loaded, not discarded)", loaded.DisplayName)
	}
	// File must NOT be deleted or backed up — tamper is cosmetic, not a lockout.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("tampered account file was removed: %v", err)
	}
	if _, err := os.Stat(path + ".corrupt"); !os.IsNotExist(err) {
		t.Error("tampered (parseable) account wrongly produced a .corrupt backup")
	}
}

// TestLoadOrCreateCorruptBacksUpAndRecreates covers the unparseable path: garbage
// JSON on disk must be renamed to account.json.corrupt and a fresh account created.
func TestLoadOrCreateCorruptBacksUpAndRecreates(t *testing.T) {
	isolateAccountDir(t)

	// Seed a valid account first so the file lives at the resolved path, then
	// clobber it with non-JSON.
	first, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (seed): %v", err)
	}
	path := accountFilePath()
	if err := os.WriteFile(path, []byte("this is not json {{{"), 0644); err != nil {
		t.Fatalf("write corrupt account: %v", err)
	}

	fresh, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (corrupt): %v", err)
	}
	// A brand-new account: valid sig, no tamper flag, and a different ID than the
	// (now-corrupted) original.
	if fresh.AccountID == "" {
		t.Fatal("fresh account after corrupt has empty AccountID")
	}
	if fresh.AccountID == first.AccountID {
		t.Error("fresh account reused the corrupt account's ID (should be regenerated)")
	}
	if fresh.Tampered {
		t.Error("fresh account after corrupt flagged Tampered")
	}
	if !verifyAccount(fresh) {
		t.Error("fresh account after corrupt fails signature verification")
	}

	// The .corrupt backup must exist and hold the original garbage bytes.
	backup := path + ".corrupt"
	got, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("expected .corrupt backup, got: %v", err)
	}
	if string(got) != "this is not json {{{" {
		t.Errorf(".corrupt backup = %q, want the original garbage bytes", string(got))
	}
	// And the live file must parse + verify again.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("fresh account.json not written after corrupt recovery: %v", err)
	}
}

// TestWipeAccount covers the destructive wipe: an established account is created,
// confirmed present, then WipeAccount() must delete the file so a subsequent
// LoadAccount reports found=false; the corrupt-backup sibling is also removed; and
// wiping with no file present is a clean no-op (no error).
func TestWipeAccount(t *testing.T) {
	isolateAccountDir(t)

	// Create a named account and confirm the file exists.
	acct, err := CreateNamedAccount("Adam")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	if acct.DisplayName != "Adam" {
		t.Fatalf("DisplayName = %q, want Adam", acct.DisplayName)
	}
	path := accountFilePath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("account.json not written: %v", err)
	}

	// Seed a corrupt-backup sibling to prove WipeAccount sweeps it too.
	backup := path + ".corrupt"
	if err := os.WriteFile(backup, []byte("garbage"), 0644); err != nil {
		t.Fatalf("seed .corrupt backup: %v", err)
	}

	// Wipe: file and its .corrupt sibling must be gone, with no error.
	if err := WipeAccount(); err != nil {
		t.Fatalf("WipeAccount: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("account.json still present after wipe (stat err = %v)", err)
	}
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Errorf("account.json.corrupt still present after wipe (stat err = %v)", err)
	}

	// A subsequent LoadAccount must report found=false (nothing to load).
	loaded, found, err := LoadAccount()
	if err != nil {
		t.Fatalf("LoadAccount after wipe: %v", err)
	}
	if found || loaded != nil {
		t.Errorf("LoadAccount after wipe = (%v, found=%v), want (nil, false)", loaded, found)
	}

	// Wiping again (no file present) is a clean no-op.
	if err := WipeAccount(); err != nil {
		t.Errorf("WipeAccount on missing file returned error: %v", err)
	}
}

// --- Phase 4: recovery code (accounts.md §3.5 / §8 / §9) ---

// TestRecoveryCodeRoundTrip covers the core contract: a.RecoveryCode() →
// ImportRecoveryCode(thatCode) yields an account with the SAME AccountID.
func TestRecoveryCodeRoundTrip(t *testing.T) {
	isolateAccountDir(t)

	orig, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	code := orig.RecoveryCode()

	restored, err := ImportRecoveryCode(code)
	if err != nil {
		t.Fatalf("ImportRecoveryCode(%q): %v", code, err)
	}
	if restored.AccountID != orig.AccountID {
		t.Errorf("round-trip ID mismatch: %q != %q", restored.AccountID, orig.AccountID)
	}
}

// TestRecoveryCodeFormat checks the produced shape: AGEF- prefix, all uppercase,
// dash-grouped in 4-char groups, and that it decodes back to the same ID.
func TestRecoveryCodeFormat(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	code := acct.RecoveryCode()

	if !strings.HasPrefix(code, "AGEF-") {
		t.Errorf("code %q missing AGEF- prefix", code)
	}
	if code != strings.ToUpper(code) {
		t.Errorf("code %q is not uppercase", code)
	}
	// Each dash-separated group after the prefix is at most 4 chars (4s with a
	// trailing partial group).
	groups := strings.Split(code, "-")
	if groups[0] != "AGEF" {
		t.Errorf("first group = %q, want AGEF", groups[0])
	}
	for _, g := range groups[1:] {
		if len(g) == 0 || len(g) > 4 {
			t.Errorf("group %q has invalid length %d (want 1..4)", g, len(g))
		}
	}
	// Must still decode back to the same ID.
	restored, err := ImportRecoveryCode(code)
	if err != nil {
		t.Fatalf("ImportRecoveryCode(%q): %v", code, err)
	}
	if restored.AccountID != acct.AccountID {
		t.Errorf("decoded ID mismatch: %q != %q", restored.AccountID, acct.AccountID)
	}
}

// TestRecoveryCodeTypoGuard flips a single payload character in a valid code and
// asserts ImportRecoveryCode rejects it with a checksum error (not a silent wrong ID).
func TestRecoveryCodeTypoGuard(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	code := acct.RecoveryCode()

	// Flip an early body character (a high-order payload bit, well inside the real
	// 144-bit payload — not the trailing padding bits) to a different valid Crockford
	// symbol. Mutating real ID/checksum bits must trip the checksum guard.
	b := []byte(code)
	// Index 5 is the first body char after the "AGEF-" prefix (indices 0..4).
	idx := 5
	orig := b[idx]
	repl := byte('Z')
	if orig == repl {
		repl = '2'
	}
	b[idx] = repl
	typo := string(b)
	if typo == code {
		t.Fatal("typo mutation did not change the code")
	}

	if _, err := ImportRecoveryCode(typo); err == nil {
		t.Errorf("typo'd code %q imported without error (typo guard failed)", typo)
	} else if !strings.Contains(err.Error(), "checksum") {
		t.Errorf("typo'd code error = %q, want a checksum error", err.Error())
	}
}

// TestImportRecoveryCodeWritesSignedEmptyAccount confirms the imported account is
// written as a valid signed account.json with EMPTY data (identity only).
func TestImportRecoveryCodeWritesSignedEmptyAccount(t *testing.T) {
	isolateAccountDir(t)

	orig, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	code := orig.RecoveryCode()

	restored, err := ImportRecoveryCode(code)
	if err != nil {
		t.Fatalf("ImportRecoveryCode: %v", err)
	}
	// Signed in-memory (no tamper), and EMPTY data — identity only.
	if !verifyAccount(restored) {
		t.Error("imported account fails signature verification")
	}
	if restored.Tampered {
		t.Error("imported account flagged Tampered")
	}
	if len(restored.Unlocks.Themes) != 0 || restored.Stats.TotalPrestiges != 0 ||
		len(restored.Achievements) != 0 || restored.Prefs.ActiveTheme != "" {
		t.Errorf("imported account carried DATA, want empty: unlocks=%v stats=%+v ach=%v prefs=%+v",
			restored.Unlocks.Themes, restored.Stats, restored.Achievements, restored.Prefs)
	}

	// And the on-disk file must reload, verify, and be untampered.
	reloaded, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (reload after import): %v", err)
	}
	if reloaded.AccountID != orig.AccountID {
		t.Errorf("reloaded ID = %q, want %q", reloaded.AccountID, orig.AccountID)
	}
	if reloaded.Tampered {
		t.Error("reloaded imported account flagged Tampered (signed Save failed)")
	}
	if !verifyAccount(reloaded) {
		t.Error("reloaded imported account fails signature verification")
	}
}

// TestImportRecoveryCodeLenient covers Crockford's lenient decode: lowercase, with
// spaces, and I-vs-1 substituted variants of a valid code all import to the same ID.
func TestImportRecoveryCodeLenient(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	code := acct.RecoveryCode()
	want := acct.AccountID

	variants := map[string]string{
		"lowercase":      strings.ToLower(code),
		"with-spaces":    strings.ReplaceAll(code, "-", " "),
		"no-dashes":      strings.ReplaceAll(code, "-", ""),
		"i-for-1":        strings.ReplaceAll(code, "1", "I"),
		"l-for-1":        strings.ReplaceAll(code, "1", "L"),
		"o-for-0":        strings.ReplaceAll(code, "0", "O"),
		"surround-space": "  " + code + "  ",
	}
	for name, v := range variants {
		got, err := ImportRecoveryCode(v)
		if err != nil {
			t.Errorf("%s: ImportRecoveryCode(%q): %v", name, v, err)
			continue
		}
		if got.AccountID != want {
			t.Errorf("%s: ID = %q, want %q", name, got.AccountID, want)
		}
	}
}
