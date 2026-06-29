package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- Phase B: multi-account API (enumerate / switch / create) ---
//
// All cases isolate the data ROOT via isolateAccountDir(t) (which also resets the
// process-global activeAccountID to "" for the test and restores it on cleanup), so
// CreateAccount/SwitchAccount/ListAccounts operate entirely inside a temp tree.

// TestCreateAccountMintsEmpty covers the fresh-create contract: a brand-new name mints an
// EMPTY established account in its own slot and makes it active — with NO carry-over from a
// previously-active account.
func TestCreateAccountMintsEmpty(t *testing.T) {
	isolateAccountDir(t)

	a, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	if a.DisplayName != "Alice" {
		t.Errorf("DisplayName = %q, want Alice", a.DisplayName)
	}
	if a.AccountID != accountIDFromName("Alice") {
		t.Errorf("AccountID = %q, want name-derived %q", a.AccountID, accountIDFromName("Alice"))
	}
	if !a.Established() {
		t.Error("created account is not Established (DisplayName empty?)")
	}
	// EMPTY data — no carry-over, nothing inherited.
	if len(a.Unlocks.Themes) != 0 || a.Stats.TotalPrestiges != 0 ||
		len(a.Achievements) != 0 || a.Prefs.ActiveTheme != "" {
		t.Errorf("fresh account carried DATA, want empty: unlocks=%v stats=%+v ach=%v prefs=%+v",
			a.Unlocks.Themes, a.Stats, a.Achievements, a.Prefs)
	}
	// Active + persisted pointer point at the new account.
	if getActiveAccountID() != a.AccountID {
		t.Errorf("active id = %q, want %q", getActiveAccountID(), a.AccountID)
	}
	if ptr, _ := readActivePointer(); ptr != a.AccountID {
		t.Errorf("active pointer = %q, want %q", ptr, a.AccountID)
	}
	// And it round-trips from its slot.
	loaded, found, err := loadAccountFromSlot(a.AccountID)
	if err != nil || !found {
		t.Fatalf("loadAccountFromSlot(new) = (found=%v, %v), want found", found, err)
	}
	if loaded.DisplayName != "Alice" {
		t.Errorf("reloaded DisplayName = %q, want Alice", loaded.DisplayName)
	}
}

// TestCreateAccountNoCarryOver proves the distinction from CreateNamedAccount: creating a
// second, differently-named account does NOT copy the first account's earned progress, and
// the first account's data on disk is left fully intact.
func TestCreateAccountNoCarryOver(t *testing.T) {
	isolateAccountDir(t)

	// First account with real progress.
	first, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	first.Unlocks.Themes = []string{"obsidian", "amber_crt"}
	first.Stats.TotalPrestiges = 7
	first.Stats.HighestAge = "bronze_age"
	first.Achievements = []string{"first_prestige"}
	first.Prefs.ActiveTheme = "obsidian"
	if err := first.Save(); err != nil {
		t.Fatalf("first.Save: %v", err)
	}
	firstID := first.AccountID

	// Second, brand-new account — must start EMPTY (no carry-over from Alice).
	second, err := CreateAccount("Bob")
	if err != nil {
		t.Fatalf("CreateAccount(Bob): %v", err)
	}
	if second.AccountID == firstID {
		t.Fatal("Bob derived the same id as Alice (names should differ)")
	}
	if len(second.Unlocks.Themes) != 0 || second.Stats.TotalPrestiges != 0 ||
		len(second.Achievements) != 0 || second.Prefs.ActiveTheme != "" {
		t.Errorf("Bob carried Alice's DATA, want empty: unlocks=%v stats=%+v ach=%v prefs=%+v",
			second.Unlocks.Themes, second.Stats, second.Achievements, second.Prefs)
	}

	// Alice's data on disk is untouched (load her slot directly — switching to Bob must
	// not have rewritten Alice).
	alice, found, err := loadAccountFromSlot(firstID)
	if err != nil || !found {
		t.Fatalf("loadAccountFromSlot(Alice) = (found=%v, %v), want found", found, err)
	}
	if alice.Stats.TotalPrestiges != 7 || alice.Stats.HighestAge != "bronze_age" {
		t.Errorf("Alice's stats changed: %+v, want {TotalPrestiges:7 HighestAge:bronze_age ...}", alice.Stats)
	}
	if len(alice.Unlocks.Themes) != 2 || alice.Prefs.ActiveTheme != "obsidian" {
		t.Errorf("Alice's unlocks/prefs changed: themes=%v active=%q", alice.Unlocks.Themes, alice.Prefs.ActiveTheme)
	}
}

// TestCreateAccountExistingNameOpens covers same-name == same identity: CreateAccount with a
// name that already has a slot OPENS that account (switches to it) without wiping its data.
func TestCreateAccountExistingNameOpens(t *testing.T) {
	isolateAccountDir(t)

	// Seed Alice with progress, then switch away so she is not active.
	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	aliceID := alice.AccountID
	alice.Stats.TotalPrestiges = 5
	alice.Unlocks.Themes = []string{"obsidian"}
	if err := alice.Save(); err != nil {
		t.Fatalf("alice.Save: %v", err)
	}
	if _, err := CreateAccount("Bob"); err != nil {
		t.Fatalf("CreateAccount(Bob): %v", err)
	}
	if getActiveAccountID() == aliceID {
		t.Fatal("setup: expected Bob active, got Alice")
	}

	// Re-create Alice by name → must OPEN the existing slot, preserving her data, and make
	// her active again.
	reopened, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount(Alice) reopen: %v", err)
	}
	if reopened.AccountID != aliceID {
		t.Errorf("reopened id = %q, want existing %q", reopened.AccountID, aliceID)
	}
	if reopened.Stats.TotalPrestiges != 5 || len(reopened.Unlocks.Themes) != 1 {
		t.Errorf("reopen wiped Alice's data: stats=%+v themes=%v", reopened.Stats, reopened.Unlocks.Themes)
	}
	if getActiveAccountID() != aliceID {
		t.Errorf("active id = %q, want Alice %q after reopen", getActiveAccountID(), aliceID)
	}
}

// TestListAccountsEnumerates covers the core enumeration: multiple slots are listed, exactly
// one is marked Active, and the headline fields are populated from each account.
func TestListAccountsEnumerates(t *testing.T) {
	isolateAccountDir(t)

	if got := ListAccounts(); len(got) != 0 {
		t.Errorf("ListAccounts with no accounts/ dir = %d entries, want 0", len(got))
	}

	if _, err := CreateAccount("Alice"); err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	if _, err := CreateAccount("Bob"); err != nil {
		t.Fatalf("CreateAccount(Bob): %v", err)
	}
	// Give Carol some progress so the headline fields are non-zero, and leave her active.
	carol, err := CreateAccount("Carol")
	if err != nil {
		t.Fatalf("CreateAccount(Carol): %v", err)
	}
	carol.Stats.TotalPrestiges = 3
	carol.Stats.HighestAge = "iron_age"
	carol.Achievements = []string{"first_prestige", "reached_iron"}
	if err := carol.Save(); err != nil {
		t.Fatalf("carol.Save: %v", err)
	}

	list := ListAccounts()
	if len(list) != 3 {
		t.Fatalf("ListAccounts = %d entries, want 3: %+v", len(list), list)
	}

	activeCount := 0
	byName := map[string]AccountSummary{}
	for _, s := range list {
		byName[s.DisplayName] = s
		if s.Active {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Errorf("Active-marked summaries = %d, want exactly 1", activeCount)
	}
	if !byName["Carol"].Active {
		t.Error("Carol (last created) not marked Active")
	}
	// Active sorts first.
	if !list[0].Active || list[0].DisplayName != "Carol" {
		t.Errorf("first entry = %q (active=%v), want Carol active-first", list[0].DisplayName, list[0].Active)
	}
	c := byName["Carol"]
	if c.AccountID != carol.AccountID || c.TotalPrestiges != 3 || c.HighestAge != "iron_age" || c.Achievements != 2 {
		t.Errorf("Carol summary fields wrong: %+v", c)
	}
	if c.LastSeen.IsZero() {
		t.Error("Carol summary LastSeen is zero (should be populated from the account)")
	}
}

// TestListAccountsDoesNotAlterActive is the load-bearing invariant: enumerating slots must
// not change the active account, even though it loads several non-active slots.
func TestListAccountsDoesNotAlterActive(t *testing.T) {
	isolateAccountDir(t)

	if _, err := CreateAccount("Alice"); err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	if _, err := CreateAccount("Bob"); err != nil {
		t.Fatalf("CreateAccount(Bob): %v", err)
	}
	// Bob is active now.
	before := getActiveAccountID()
	beforePtr, _ := readActivePointer()

	_ = ListAccounts()

	if after := getActiveAccountID(); after != before {
		t.Errorf("ListAccounts changed active id: %q -> %q", before, after)
	}
	if afterPtr, _ := readActivePointer(); afterPtr != beforePtr {
		t.Errorf("ListAccounts changed active pointer: %q -> %q", beforePtr, afterPtr)
	}
}

// TestSwitchAccount covers both arms: switching to an existing id changes the active id +
// pointer and returns it; switching to a missing id errors and leaves active unchanged.
func TestSwitchAccount(t *testing.T) {
	isolateAccountDir(t)

	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	bob, err := CreateAccount("Bob") // Bob active now
	if err != nil {
		t.Fatalf("CreateAccount(Bob): %v", err)
	}
	if getActiveAccountID() != bob.AccountID {
		t.Fatalf("setup: expected Bob active")
	}

	// Switch back to Alice.
	got, err := SwitchAccount(alice.AccountID)
	if err != nil {
		t.Fatalf("SwitchAccount(Alice): %v", err)
	}
	if got.AccountID != alice.AccountID || got.DisplayName != "Alice" {
		t.Errorf("SwitchAccount returned %+v, want Alice", got)
	}
	if getActiveAccountID() != alice.AccountID {
		t.Errorf("active id = %q, want Alice %q", getActiveAccountID(), alice.AccountID)
	}
	if ptr, _ := readActivePointer(); ptr != alice.AccountID {
		t.Errorf("active pointer = %q, want Alice %q", ptr, alice.AccountID)
	}

	// Switch to a missing id → error, active unchanged.
	beforeBad := getActiveAccountID()
	if _, err := SwitchAccount("deadbeefdeadbeefdeadbeefdeadbeef"); err == nil {
		t.Error("SwitchAccount(missing) returned nil error, want 'no such account'")
	}
	if getActiveAccountID() != beforeBad {
		t.Errorf("failed SwitchAccount changed active id: %q -> %q", beforeBad, getActiveAccountID())
	}
}

// TestListAccountsSurvivesCorruptSlot confirms a corrupt/tampered slot does not crash
// enumeration. A tampered-but-parseable account.json surfaces with Tampered=true; an
// unparseable one is simply skipped. Either way the valid slots still list cleanly.
func TestListAccountsSurvivesCorruptSlot(t *testing.T) {
	isolateAccountDir(t)

	good, err := CreateAccount("Good")
	if err != nil {
		t.Fatalf("CreateAccount(Good): %v", err)
	}
	tampered, err := CreateAccount("Tampered") // active after this
	if err != nil {
		t.Fatalf("CreateAccount(Tampered): %v", err)
	}

	// Tamper with the Tampered slot: mutate a signed field without re-signing.
	tpath := filepath.Join(accountDir(tampered.AccountID), accountFileName)
	data, err := os.ReadFile(tpath)
	if err != nil {
		t.Fatalf("read tampered slot: %v", err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatalf("unmarshal tampered slot: %v", err)
	}
	onDisk["display_name"] = "intruder" // stale _sig now mismatches
	mutated, _ := json.Marshal(onDisk)
	if err := os.WriteFile(tpath, mutated, 0644); err != nil {
		t.Fatalf("write tampered slot: %v", err)
	}

	// Add a fully-unparseable third slot — it must be skipped, not crash ListAccounts.
	junkID := "ffffffffffffffffffffffffffffffff"
	if err := os.MkdirAll(accountDir(junkID), 0755); err != nil {
		t.Fatalf("mkdir junk slot: %v", err)
	}
	if err := os.WriteFile(filepath.Join(accountDir(junkID), accountFileName), []byte("not json {{{"), 0644); err != nil {
		t.Fatalf("write junk account.json: %v", err)
	}

	list := ListAccounts() // must not panic
	byName := map[string]AccountSummary{}
	for _, s := range list {
		byName[s.DisplayName] = s
	}
	// Good is present and not tampered.
	if g, ok := byName["Good"]; !ok {
		t.Errorf("Good slot missing from list: %+v", list)
	} else if g.Tampered || g.AccountID != good.AccountID {
		t.Errorf("Good summary wrong: %+v", g)
	}
	// The tampered slot surfaces as "intruder" with Tampered=true (signature surfaced, honest).
	if ts, ok := byName["intruder"]; !ok {
		t.Errorf("tampered slot missing from list (want surfaced with Tampered): %+v", list)
	} else if !ts.Tampered {
		t.Error("tampered slot not flagged Tampered=true")
	}
	// The unparseable junk slot does not appear.
	if len(list) != 2 {
		t.Errorf("ListAccounts = %d entries, want 2 (junk slot skipped): %+v", len(list), list)
	}
}
