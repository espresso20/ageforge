package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Account version 2 and the files from before it. The fixtures under
// testdata/account_v1 were written by the version 1 code (CreateAccount,
// the Record hooks, UnlockTheme, ExportProgress, SaveGame) before any of
// this existed; account_v1_edited.json was changed by hand after signing.
// Every test copies a fixture into a temp data root (isolateAccountDir)
// and never touches data/.

const v1Dir = "testdata/account_v1"

// installV1 copies a version 1 account fixture into its own slot of the
// temp data root and returns its ID and its bytes.
func installV1(t *testing.T, name string) (id string, raw []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(v1Dir, name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var head struct {
		AccountID string `json:"account_id"`
		Version   int    `json:"version"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	if head.Version != 1 {
		t.Fatalf("fixture %s is version %d, want 1", name, head.Version)
	}
	if err := os.MkdirAll(accountDir(head.AccountID), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir(head.AccountID), accountFileName), raw, 0644); err != nil {
		t.Fatal(err)
	}
	return head.AccountID, raw
}

// openIn loads the account in slot id as a boot does and hands it to a
// fresh engine, which brings it up to version 2.
func openIn(t *testing.T, id string) (*GameEngine, *Account) {
	t.Helper()
	acct := slotAccount(t, id)
	ge := NewGameEngine()
	ge.SetAccount(acct)
	return ge, acct
}

// TestV1AccountBecomesV2 is the migration: an account file from before
// badges loads with nothing lost, its achievements are badges and stay
// earned, what its record already proves is granted, nothing is announced,
// and a second load changes nothing.
func TestV1AccountBecomesV2(t *testing.T) {
	cases := []struct {
		file     string
		tampered bool
		badges   []string
		prestige float64 // the prestige counter, seeded from the lifetime stat
	}{
		{
			// All four achievements, 11 prestiges, the Information Age reached.
			file:     "account_v1_full.json",
			badges:   []string{badgeIron, badgeModern, badgeStone, badgePrestige1, badgePrestige3, badgePrestige10},
			prestige: 11,
		},
		{
			// reached_iron and first_prestige; the Medieval Age reached.
			file:     "account_v1_partial.json",
			badges:   []string{badgeIron, badgeStone, badgePrestige1},
			prestige: 1,
		},
		{file: "account_v1_empty.json"},
		{
			// prestige_x10 typed into the file by hand: it becomes a badge, crossed.
			file:     "account_v1_edited.json",
			tampered: true,
			badges:   []string{badgePrestige1, badgePrestige10},
			prestige: 1,
		},
		{
			// Already flagged and re-signed by version 1.
			file:     "account_v1_flagged.json",
			tampered: true,
			badges:   []string{badgeIron, badgeStone},
		},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			isolateAccountDir(t)
			id, raw := installV1(t, tc.file)
			var old Account
			if err := json.Unmarshal(raw, &old); err != nil {
				t.Fatal(err)
			}

			ge, acct := openIn(t, id)
			if acct.Tampered != tc.tampered {
				t.Fatalf("tampered = %v, want %v: a clean version 1 file must still verify, an edited one must not", acct.Tampered, tc.tampered)
			}
			if acct.Version != accountSchemaVersion {
				t.Errorf("version %d after loading, want %d", acct.Version, accountSchemaVersion)
			}

			// Nothing lost.
			if acct.AccountID != old.AccountID || acct.DisplayName != old.DisplayName || !acct.Created.Equal(old.Created) || !acct.LastSeen.Equal(old.LastSeen) {
				t.Errorf("identity changed: %s %q", acct.AccountID, acct.DisplayName)
			}
			if !reflect.DeepEqual(acct.Stats, old.Stats) {
				t.Errorf("lifetime stats changed:\n got %+v\nwant %+v", acct.Stats, old.Stats)
			}
			if !reflect.DeepEqual(acct.Unlocks, old.Unlocks) || !reflect.DeepEqual(acct.Prefs, old.Prefs) {
				t.Errorf("unlocks or prefs changed: %+v %+v", acct.Unlocks, acct.Prefs)
			}
			if !slices.Equal(acct.Achievements, old.Achievements) {
				t.Errorf("the achievements list changed: %v, was %v", acct.Achievements, old.Achievements)
			}

			// The achievements are badges, with what the record proves.
			want := slices.Clone(tc.badges)
			slices.Sort(want)
			if got := acct.EarnedBadges(); !slices.Equal(got, want) {
				t.Errorf("badges %v, want %v", got, want)
			}
			for _, key := range acct.EarnedBadges() {
				e := acct.Badges[key]
				if e.At != 0 || e.Run != "" {
					t.Errorf("%s came over dated or with a run: %+v", key, e)
				}
				if crossed := e.Flags&BadgeFlagCrossed != 0; crossed != tc.tampered {
					t.Errorf("%s crossed = %v on an account with tampered = %v", key, crossed, tc.tampered)
				}
			}
			if got := counter(acct, config.BadgeEvPrestige); got != tc.prestige {
				t.Errorf("prestige counter %v, want %v", got, tc.prestige)
			}
			if earned := ge.DrainEarnedBadges(); len(earned) != 0 {
				t.Errorf("the migration announced badges: %+v", earned)
			}

			// It is saved at once, as a version 2 file that verifies.
			first := slotFile(t, id)
			disk := slotAccount(t, id)
			if disk.Version != accountSchemaVersion || !verifyAccount(disk) || disk.Tampered != tc.tampered {
				t.Errorf("on disk: version %d, verifies %v, tampered %v", disk.Version, verifyAccount(disk), disk.Tampered)
			}
			if !slices.Equal(disk.EarnedBadges(), want) {
				t.Errorf("badges on disk %v, want %v", disk.EarnedBadges(), want)
			}

			// A second load changes nothing.
			ge2, again := openIn(t, id)
			if err := again.FlushIfDirty(); err != nil {
				t.Fatal(err)
			}
			if second := slotFile(t, id); string(second) != string(first) {
				t.Errorf("the second load rewrote the file:\nfirst  %s\nsecond %s", first, second)
			}
			if earned := ge2.DrainEarnedBadges(); len(earned) != 0 {
				t.Errorf("the second load announced badges: %+v", earned)
			}
		})
	}
}

// TestMigratedAccountKeepsEarning: after the migration the account earns
// like any other. The next prestige continues the seeded counter, and a
// badge it already held is not earned again.
func TestMigratedAccountKeepsEarning(t *testing.T) {
	isolateAccountDir(t)
	id, _ := installV1(t, "account_v1_partial.json")
	ge, acct := openIn(t, id)
	if err := ge.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}

	advanceTo(ge, "stone_age") // held since the migration
	prestigeNow(t, ge)
	prestigeNow(t, ge) // the third prestige of the account's life
	earned := ge.DrainEarnedBadges()
	if len(earned) != 1 || earned[0].Key != badgePrestige3 {
		t.Fatalf("announced %+v, want only the rung at 3 prestiges", earned)
	}
	if got := counter(acct, config.BadgeEvPrestige); got != 3 {
		t.Errorf("prestige counter %v, want 3 (1 from before, 2 now)", got)
	}
	if e := acct.Badges[badgeStone]; e.At != 0 {
		t.Errorf("a badge held since the migration was earned again: %+v", e)
	}
}

// TestV1ExportImports: an export made before badges still verifies and
// lands as a version 2 account, in both import paths.
func TestV1ExportImports(t *testing.T) {
	blob, err := os.ReadFile(filepath.Join(v1Dir, "export_v1_full.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{badgeIron, badgeModern, badgeStone, badgePrestige1, badgePrestige3, badgePrestige10}
	slices.Sort(want)

	// Through the engine: brought up before it is saved.
	isolateAccountDir(t)
	ge := NewGameEngine()
	imported, err := ge.ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("the version 1 export was refused: %v", err)
	}
	disk := slotAccount(t, imported.AccountID)
	if disk.Tampered || disk.Version != accountSchemaVersion || !slices.Equal(disk.EarnedBadges(), want) {
		t.Errorf("imported through the engine: version %d, tampered %v, badges %v", disk.Version, disk.Tampered, disk.EarnedBadges())
	}

	// Without an engine the data lands as it is, and the first engine to
	// hold the account brings it up.
	isolateAccountDir(t)
	plain, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("the version 1 export was refused: %v", err)
	}
	if len(plain.Badges) != 0 || len(plain.Achievements) != 4 {
		t.Fatalf("a plain import changed the data: badges %v, achievements %v", plain.Badges, plain.Achievements)
	}
	_, opened := openIn(t, plain.AccountID)
	if got := opened.EarnedBadges(); !slices.Equal(got, want) {
		t.Errorf("after the first engine held it: %v, want %v", got, want)
	}
}

// TestExportCarriesBadges: an export carries the badges, the counters and
// the days, signed, and they come back.
func TestExportCarriesBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	for i := 0; i < 3; i++ {
		finishOne(ge, "hut")
	}
	acct.noteDay("2026-10-06")
	acct.noteDay("2026-10-07")
	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}

	isolateAccountDir(t) // another machine
	back, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatal(err)
	}
	if !hasBadge(back, badgeStone) || back.Badges[badgeStone].At == 0 {
		t.Errorf("the badge did not come back dated: %+v", back.Badges)
	}
	if got := back.Counters[config.BadgeEvBuiltLineage+".housing"]; got != 3 {
		t.Errorf("the housing counter came back as %v, want 3", got)
	}
	if !slices.Equal(back.Days, []string{"2026-10-06", "2026-10-07"}) {
		t.Errorf("days came back as %v", back.Days)
	}

	// A changed badge breaks the signature.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(blob, &raw); err != nil {
		t.Fatal(err)
	}
	raw["badges"] = json.RawMessage(`{"age.modern_age":{"at":1}}`)
	forged, _ := json.Marshal(raw)
	if _, err := ImportAccountExport(forged, true); err == nil {
		t.Error("an export with a badge typed in was accepted")
	}
}

// TestImportMergesBadges: importing a backup into an account that has
// played on keeps every badge either copy holds (the earlier of two), takes
// the larger of each counter (never the sum) and the later highest age.
// Replace takes the backup as it is.
func TestImportMergesBadges(t *testing.T) {
	isolateAccountDir(t)
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	local := func() {
		acct.Badges = map[string]BadgeEarned{
			badgeStone: {At: 200, Run: "later"},
			badgeIron:  {At: 300},
		}
		acct.Counters = map[string]float64{"prestige": 2, "built.lineage.housing": 40}
		acct.Stats.HighestAge = "iron_age"
		acct.Days = []string{"2026-10-02"}
	}
	backup := &progressExport{
		Version: 2, AccountID: acct.AccountID,
		Badges: map[string]BadgeEarned{
			badgeStone:     {At: 100, Run: "first"},
			badgePrestige1: {At: 150},
		},
		Counters: map[string]float64{"prestige": 5, "built.lineage.housing": 10},
		Stats:    AccountStats{HighestAge: "medieval_age"},
		Days:     []string{"2026-10-01", "2026-10-02"},
	}

	local()
	acct.applyExportLocked(backup, true)
	if e := acct.Badges[badgeStone]; e.At != 100 || e.Run != "first" {
		t.Errorf("merge kept the later copy of a badge both hold: %+v", e)
	}
	if !hasBadge(acct, badgeIron) || !hasBadge(acct, badgePrestige1) {
		t.Errorf("merge dropped a badge one side held: %v", acct.Badges)
	}
	if acct.Counters["prestige"] != 5 || acct.Counters["built.lineage.housing"] != 40 {
		t.Errorf("counters after merge %v, want the larger of each", acct.Counters)
	}
	if acct.Stats.HighestAge != "medieval_age" {
		t.Errorf("highest age after merge %q, want the later one", acct.Stats.HighestAge)
	}
	if !slices.Equal(acct.Days, []string{"2026-10-01", "2026-10-02"}) {
		t.Errorf("days after merge %v", acct.Days)
	}
	// Importing the same backup again doubles nothing.
	acct.applyExportLocked(backup, true)
	if acct.Counters["prestige"] != 5 {
		t.Errorf("a second import of the same backup moved a counter to %v", acct.Counters["prestige"])
	}

	// A lower highest age in the backup does not pull the account back.
	local()
	backup.Stats.HighestAge = "stone_age"
	acct.applyExportLocked(backup, true)
	if acct.Stats.HighestAge != "iron_age" {
		t.Errorf("an older backup lowered the highest age to %q", acct.Stats.HighestAge)
	}

	local()
	acct.applyExportLocked(backup, false)
	if hasBadge(acct, badgeIron) || acct.Counters["built.lineage.housing"] != 10 || len(acct.Days) != 2 {
		t.Errorf("replace did not take the backup as it is: %v %v %v", acct.Badges, acct.Counters, acct.Days)
	}
}

// TestSaveFromBeforeRunFacts: a save written before run facts loads
// unflagged, with the run's past marked unknown, tracks from there, and
// keeps what it tracked through a save and a load.
func TestSaveFromBeforeRunFacts(t *testing.T) {
	isolateAccountDir(t)
	id, _ := installV1(t, "account_v1_saver.json")
	raw, err := os.ReadFile(filepath.Join(v1Dir, "save_pre_run_facts.json"))
	if err != nil {
		t.Fatal(err)
	}
	saves := filepath.Join(accountDir(id), "saves")
	if err := os.MkdirAll(saves, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(saves, "fixture.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
	if err := makeActive(id); err != nil {
		t.Fatal(err)
	}

	ge, acct := openIn(t, id)
	if err := ge.LoadGame("fixture"); err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	st := ge.GetState()
	if st.CheaterBadge || st.DevTouched {
		t.Fatalf("a save from before run facts loaded flagged: modified %v, dev %v", st.CheaterBadge, st.DevTouched)
	}
	if hasBadge(acct, badgeAccounting) {
		t.Error("a clean old save earned Creative Accounting")
	}
	ge.mu.RLock()
	facts := ge.runFacts.clone()
	ge.mu.RUnlock()
	if facts.Whole {
		t.Error("the run's past reads as known for a save that never recorded it")
	}
	if !facts.AgeKnown || facts.AgeTick != 0 {
		t.Errorf("a run still in its first age entered it at tick 0: %+v", facts)
	}
	if got := ge.Buildings.GetCount("hut"); got < 2 {
		t.Fatalf("the save's two huts did not load: %d", got)
	}

	// It tracks from here, and the marks keep a rebuilt hut from counting.
	finishOne(ge, "hut")
	if got := counter(acct, config.BadgeEvBuiltLineage+".housing"); got != 1 {
		t.Errorf("a hut built after the load left the counter at %v", got)
	}
	if err := ge.SaveGame("fixture"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	if err := ge2.LoadGame("fixture"); err != nil {
		t.Fatal(err)
	}
	if ge2.GetState().CheaterBadge {
		t.Error("the save reads as modified after run facts were added to it")
	}
	ge2.mu.RLock()
	again := ge2.runFacts.clone()
	ge2.mu.RUnlock()
	if again.Whole || again.Net["hut"] != 1 || again.High["hut"] != 1 || again.Counts[config.BadgeEvBuildingBuilt] != 1 {
		t.Errorf("run facts after a save and a load: %+v", again)
	}
}

// TestRunFactsRoundTrip: a new run's facts are whole, are saved, and load
// equal; a save with nothing recorded writes no run_facts key.
func TestRunFactsRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	finishOne(ge, "hut")
	sellSome(ge, "hut", 1)
	ge.mu.RLock()
	want := ge.runFacts.clone()
	ge.mu.RUnlock()
	if !want.Whole || want.Counts[config.BadgeEvAgeReached] != 1 || want.Counts[config.BadgeEvBuildingSold] != 1 {
		t.Fatalf("facts of a new run: %+v", want)
	}
	if err := ge.SaveGame("run"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	if err := ge2.LoadGame("run"); err != nil {
		t.Fatal(err)
	}
	ge2.mu.RLock()
	got := ge2.runFacts.clone()
	ge2.mu.RUnlock()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("run facts after a load:\n got %+v\nwant %+v", got, want)
	}

	bare := NewGameEngine()
	bare.mu.Lock()
	bare.runFacts = RunFacts{}
	snap := bare.buildSaveSnapshot()
	bare.mu.Unlock()
	if snap.RunFacts != nil {
		t.Errorf("a run with nothing recorded writes run facts: %+v", snap.RunFacts)
	}
}
