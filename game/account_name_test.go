package game

import (
	"testing"
)

// TestAccountIDFromNameDeterministic confirms the derived ID is stable for the same
// name and 32 hex chars (16 bytes) — the format the recovery code + save attribution
// depend on.
func TestAccountIDFromNameDeterministic(t *testing.T) {
	id1 := accountIDFromName("Imperium")
	id2 := accountIDFromName("Imperium")
	if id1 != id2 {
		t.Errorf("same name derived different IDs: %q != %q", id1, id2)
	}
	if len(id1) != 32 {
		t.Errorf("derived ID = %q (len %d), want 32 hex chars", id1, len(id1))
	}
}

// TestAccountIDFromNameNormalizes confirms the normalization rules: case-insensitive,
// surrounding whitespace trimmed, internal whitespace runs collapsed — all map to one ID.
func TestAccountIDFromNameNormalizes(t *testing.T) {
	base := accountIDFromName("bob")
	variants := map[string]string{
		"trailing-space": "Bob ",
		"leading-space":  "  bob",
		"uppercase":      "BOB",
		"mixed-case":     "Bob",
	}
	for name, v := range variants {
		if got := accountIDFromName(v); got != base {
			t.Errorf("%s: accountIDFromName(%q) = %q, want %q (same as %q)", name, v, got, base, "bob")
		}
	}

	// Internal whitespace collapse: "bob  the   builder" == "bob the builder".
	if a, b := accountIDFromName("bob  the   builder"), accountIDFromName("bob the builder"); a != b {
		t.Errorf("internal whitespace not collapsed: %q != %q", a, b)
	}

	// Different names → different IDs.
	if accountIDFromName("alice") == accountIDFromName("bob") {
		t.Error("different names derived the same ID")
	}
}

// TestCreateNamedAccount confirms identity derivation: DisplayName keeps the original
// (trimmed) name, AccountID is name-derived, and the account is Established + signed.
func TestCreateNamedAccount(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateNamedAccount("  Imperium Romanum  ")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	if acct.DisplayName != "Imperium Romanum" {
		t.Errorf("DisplayName = %q, want %q (trimmed original)", acct.DisplayName, "Imperium Romanum")
	}
	if acct.AccountID != accountIDFromName("Imperium Romanum") {
		t.Errorf("AccountID = %q, want derived %q", acct.AccountID, accountIDFromName("Imperium Romanum"))
	}
	if !acct.Established() {
		t.Error("named account not Established()")
	}
	if acct.Signature == "" {
		t.Error("named account not signed")
	}
	if !verifyAccount(acct) {
		t.Error("named account fails signature verification")
	}
}

// TestCreateNamedAccountCarriesOverData confirms the migration: a pre-existing account
// with an earned unlock has that DATA carried into the new named account when the
// identity is re-keyed to a name.
func TestCreateNamedAccountCarriesOverData(t *testing.T) {
	isolateAccountDir(t)

	// Seed a legacy random-id account with an unlock + a lifetime stat.
	legacy, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (legacy seed): %v", err)
	}
	legacy.Unlocks.Themes = []string{"obsidian"}
	legacy.Stats.TotalPrestiges = 7
	legacy.Achievements = []string{"first_prestige"}
	if err := legacy.Save(); err != nil {
		t.Fatalf("Save legacy: %v", err)
	}

	named, err := CreateNamedAccount("Carthage")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	if named.AccountID != accountIDFromName("Carthage") {
		t.Errorf("AccountID = %q, want name-derived", named.AccountID)
	}
	if named.AccountID == legacy.AccountID {
		t.Error("named account kept the legacy random ID; should be name-derived")
	}
	if len(named.Unlocks.Themes) != 1 || named.Unlocks.Themes[0] != "obsidian" {
		t.Errorf("unlock not carried over: %v", named.Unlocks.Themes)
	}
	if named.Stats.TotalPrestiges != 7 {
		t.Errorf("lifetime stat not carried over: TotalPrestiges = %d, want 7", named.Stats.TotalPrestiges)
	}
	if len(named.Achievements) != 1 || named.Achievements[0] != "first_prestige" {
		t.Errorf("achievement not carried over: %v", named.Achievements)
	}
}

// TestLoadAccountFoundFalseWhenNoFile confirms LoadAccount never creates: with no file
// it returns found=false and a nil account.
func TestLoadAccountFoundFalseWhenNoFile(t *testing.T) {
	isolateAccountDir(t)

	acct, found, err := LoadAccount()
	if err != nil {
		t.Fatalf("LoadAccount (absent): %v", err)
	}
	if found {
		t.Error("LoadAccount reported found=true with no file on disk")
	}
	if acct != nil {
		t.Errorf("LoadAccount returned a non-nil account with no file: %+v", acct)
	}
}

// TestLoadAccountFoundTrueForExisting confirms LoadAccount loads an existing account.
func TestLoadAccountFoundTrueForExisting(t *testing.T) {
	isolateAccountDir(t)

	created, err := CreateNamedAccount("Babylon")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}

	loaded, found, err := LoadAccount()
	if err != nil {
		t.Fatalf("LoadAccount (existing): %v", err)
	}
	if !found {
		t.Fatal("LoadAccount reported found=false for an existing account")
	}
	if loaded.AccountID != created.AccountID {
		t.Errorf("loaded ID = %q, want %q", loaded.AccountID, created.AccountID)
	}
	if loaded.DisplayName != "Babylon" {
		t.Errorf("loaded DisplayName = %q, want Babylon", loaded.DisplayName)
	}
}

// TestRecoveryViaName confirms the cross-machine recovery story: re-entering the same
// name regenerates the same AccountID.
func TestRecoveryViaName(t *testing.T) {
	isolateAccountDir(t)

	first, err := CreateNamedAccount("Persia")
	if err != nil {
		t.Fatalf("CreateNamedAccount (first): %v", err)
	}
	second, err := CreateNamedAccount("Persia")
	if err != nil {
		t.Fatalf("CreateNamedAccount (second): %v", err)
	}
	if first.AccountID != second.AccountID {
		t.Errorf("re-entering the same name produced a different ID: %q != %q", first.AccountID, second.AccountID)
	}
}
