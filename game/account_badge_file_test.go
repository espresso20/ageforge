package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// The badge file (badges.json, beside account.json in the account's slot)
// and the accounts from before it. The fixtures under testdata/account_v1
// were written by the code from before badges (CreateAccount, the Record
// hooks, UnlockTheme, ExportProgress, SaveGame) before any of this existed;
// account_v1_edited.json was changed by hand after signing. Every test
// copies a fixture into a temp data root (isolateAccountDir) and never
// touches data/.

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
// fresh engine, which reconciles its badges with its record.
func openIn(t *testing.T, id string) (*GameEngine, *Account) {
	t.Helper()
	acct := slotAccount(t, id)
	ge := NewGameEngine()
	ge.SetAccount(acct)
	return ge, acct
}

// slotBadgeFile returns the raw bytes of the badge file in slot id, nil when
// the slot has none.
func slotBadgeFile(t *testing.T, id string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(accountDir(id), badgeFileName))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read the badge file of slot %s: %v", id, err)
	}
	return b
}

// accountFileKeys is every key account.json may hold: what the code from
// before badges writes. A key outside it would be one that code does not
// sign, and it would read the file as modified.
var accountFileKeys = []string{
	"version", "account_id", "display_name", "created", "last_seen",
	"unlocks", "stats", "achievements", "prefs", "_sig", "tampered",
}

// assertOldShape fails if raw, an account.json, holds anything the code
// from before badges does not know.
func assertOldShape(t *testing.T, raw []byte) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("account.json does not parse: %v", err)
	}
	for key := range m {
		if !slices.Contains(accountFileKeys, key) {
			t.Errorf("account.json holds the key %q, which a build from before badges does not sign", key)
		}
	}
	if string(m["version"]) != "1" {
		t.Errorf("account.json is version %s, want 1: its format has not changed", m["version"])
	}
}

// ageBadges is the badge of every age after the first, up to and including
// age: what an account that has reached age is granted from its record.
func ageBadges(age string) []string {
	var out []string
	for _, k := range rules.Core().AgeKeys()[1:] {
		out = append(out, "age."+k)
		if k == age {
			return out
		}
	}
	panic("no age " + age)
}

// themesGiven is the themes the badges in keys unlock, sorted.
func themesGiven(keys []string) []string {
	var out []string
	for _, k := range keys {
		if def, ok := rules.Core().Badge(k); ok && def.Reward.Theme != "" {
			out = append(out, def.Reward.Theme)
		}
	}
	slices.Sort(out)
	return out
}

// sortedUnion is every string in a or b, once, sorted.
func sortedUnion(a, b []string) []string {
	out := slices.Clone(a)
	for _, k := range b {
		if !slices.Contains(out, k) {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return out
}

// assertOnlyUnlocksDiffer fails if two account.json files differ in
// anything but their unlocks (and so their signature): what a theme gained
// from a badge may change, and nothing else.
func assertOnlyUnlocksDiffer(t *testing.T, was, now []byte) {
	t.Helper()
	var a, b map[string]json.RawMessage
	if err := json.Unmarshal(was, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(now, &b); err != nil {
		t.Fatal(err)
	}
	for _, key := range accountFileKeys {
		if key == "unlocks" || key == "_sig" {
			continue
		}
		if string(a[key]) != string(b[key]) {
			t.Errorf("gaining a theme changed %q in account.json: was %s, now %s", key, a[key], b[key])
		}
	}
}

// TestOldAccountGetsItsBadgeFile is the migration: an account from before
// badges loads with nothing lost, its achievements are badges in a badge
// file beside it, what its record already proves is granted, nothing is
// announced, and a second load changes nothing. account.json is left as it
// was, but for the themes the granted badges unlock: an age's theme comes
// with the badge of that age, so an account that reached the age has it.
func TestOldAccountGetsItsBadgeFile(t *testing.T) {
	cases := []struct {
		file     string
		tampered bool
		badges   []string
		prestige float64 // the prestige counter, seeded from the lifetime stat
	}{
		{
			// All four achievements, 11 prestiges, the Information Age reached.
			file: "account_v1_full.json",
			// Thirteen ages and three prestige rungs: enough badges for
			// the first rung of the ladder over earned badges.
			badges:   append(ageBadges("information_age"), badgePrestige1, badgePrestige3, badgePrestige10, badgeCollector),
			prestige: 11,
		},
		{
			// reached_iron and first_prestige; the Medieval Age reached.
			file:     "account_v1_partial.json",
			badges:   append(ageBadges("medieval_age"), badgePrestige1),
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
			// Already flagged and re-signed before badges.
			file:     "account_v1_flagged.json",
			tampered: true,
			badges:   ageBadges("iron_age"),
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
				t.Fatalf("tampered = %v, want %v: a clean file must still verify, an edited one must not", acct.Tampered, tc.tampered)
			}
			if acct.BadgesTampered {
				t.Error("an account with no badge file reads as having an edited one")
			}

			// Nothing lost.
			if acct.Version != old.Version || acct.AccountID != old.AccountID || acct.DisplayName != old.DisplayName ||
				!acct.Created.Equal(old.Created) || !acct.LastSeen.Equal(old.LastSeen) {
				t.Errorf("identity changed: version %d, %s %q", acct.Version, acct.AccountID, acct.DisplayName)
			}
			if !reflect.DeepEqual(acct.Stats, old.Stats) {
				t.Errorf("lifetime stats changed:\n got %+v\nwant %+v", acct.Stats, old.Stats)
			}
			if !reflect.DeepEqual(acct.Prefs, old.Prefs) {
				t.Errorf("prefs changed: %+v", acct.Prefs)
			}
			// No theme is lost, and the only ones gained are the granted
			// badges' own.
			for _, th := range old.Unlocks.Themes {
				if !acct.HasTheme(th) {
					t.Errorf("the account lost the theme %s", th)
				}
			}
			gained := themesGiven(tc.badges)
			wantThemes := sortedUnion(old.Unlocks.Themes, gained)
			if got := sortedUnion(acct.Unlocks.Themes, nil); !slices.Equal(got, wantThemes) || len(acct.Unlocks.Themes) != len(wantThemes) {
				t.Errorf("unlocked themes %v, want %v (it had %v, and its badges give %v)", acct.Unlocks.Themes, wantThemes, old.Unlocks.Themes, gained)
			}
			themesGrew := len(wantThemes) > len(old.Unlocks.Themes)
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

			// account.json keeps its shape, and a healthy one that gained no
			// theme keeps its bytes: the badges went to their own file. One
			// that gained a theme differs in its unlocks and nothing else.
			first := slotFile(t, id)
			assertOldShape(t, first)
			if !tc.tampered && !themesGrew && string(first) != string(raw) {
				t.Errorf("a healthy account.json was rewritten:\nwas %s\nnow %s", raw, first)
			}
			if !tc.tampered && themesGrew {
				assertOnlyUnlocksDiffer(t, raw, first)
			}
			disk := slotAccount(t, id)
			if !verifyAccount(disk) || disk.Tampered != tc.tampered || disk.BadgesTampered {
				t.Errorf("on disk: verifies %v, tampered %v, badge file tampered %v", verifyAccount(disk), disk.Tampered, disk.BadgesTampered)
			}
			if !slices.Equal(disk.EarnedBadges(), want) {
				t.Errorf("badges on disk %v, want %v", disk.EarnedBadges(), want)
			}
			firstBadges := slotBadgeFile(t, id)
			if (firstBadges != nil) != (len(want) > 0) {
				t.Errorf("badge file written = %v for an account with %d badges", firstBadges != nil, len(want))
			}
			if firstBadges != nil {
				var bf badgeFile
				if err := json.Unmarshal(firstBadges, &bf); err != nil || !verifyBadgeFile(&bf) || bf.AccountID != id || bf.Tampered {
					t.Errorf("the badge file: err %v, verifies %v, account %s, tampered %v", err, verifyBadgeFile(&bf), bf.AccountID, bf.Tampered)
				}
			}

			// A second load changes nothing, in either file.
			ge2, again := openIn(t, id)
			if err := again.FlushIfDirty(); err != nil {
				t.Fatal(err)
			}
			if second := slotFile(t, id); string(second) != string(first) {
				t.Errorf("the second load rewrote account.json:\nfirst  %s\nsecond %s", first, second)
			}
			if second := slotBadgeFile(t, id); string(second) != string(firstBadges) {
				t.Errorf("the second load rewrote the badge file:\nfirst  %s\nsecond %s", firstBadges, second)
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
	earned := mainViews(ge.DrainEarnedBadges())
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
// imports, in both import paths, and its achievements become badges.
func TestV1ExportImports(t *testing.T) {
	blob, err := os.ReadFile(filepath.Join(v1Dir, "export_v1_full.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := append(ageBadges("information_age"), badgePrestige1, badgePrestige3, badgePrestige10, badgeCollector)
	slices.Sort(want)

	// Through the engine: brought up before it is saved.
	isolateAccountDir(t)
	ge := NewGameEngine()
	imported, err := ge.ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("the version 1 export was refused: %v", err)
	}
	disk := slotAccount(t, imported.AccountID)
	if disk.Tampered || disk.BadgesTampered || !slices.Equal(disk.EarnedBadges(), want) {
		t.Errorf("imported through the engine: tampered %v, badge file tampered %v, badges %v", disk.Tampered, disk.BadgesTampered, disk.EarnedBadges())
	}
	assertOldShape(t, slotFile(t, imported.AccountID))

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

// TestExportCarriesBadges: an export carries the badge file whole, under
// its own signature, beside the part a build from before badges signs. The
// badges come back as they were; an export whose badges were changed, or
// taken from another account, is refused; and with the badges taken out it
// is still a valid export of the account.
func TestExportCarriesBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	for i := 0; i < 3; i++ {
		finishOne(ge, "hut")
	}
	// Two days from the past, beside today's (starting the game noted it).
	acct.noteDay("2001-01-01")
	acct.noteDay("2001-01-02")
	days := slices.Clone(acct.Days)
	if len(days) != 3 || !slices.IsSorted(days) {
		t.Fatalf("days on the account: %v, want two past days and today, in order", days)
	}
	earned := acct.Badges[badgeStone]
	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}
	other, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	other.judge(coreBook(), Event{Kind: config.BadgeEvAgeReached, Subject: "stone_age"}, badgeCtx{})
	otherBlob, err := other.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}

	isolateAccountDir(t) // another machine
	back, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := back.Badges[badgeStone]; got != earned || got.At == 0 || back.BadgesTampered {
		t.Errorf("the badge came back as %+v, want %+v, unflagged", got, earned)
	}
	if got := back.Counters[config.BadgeEvBuiltLineage+".housing"]; got != 3 {
		t.Errorf("the housing counter came back as %v, want 3", got)
	}
	if !slices.Equal(back.Days, days) {
		t.Errorf("days came back as %v, want %v", back.Days, days)
	}
	if disk := slotAccount(t, back.AccountID); disk.BadgesTampered || disk.Tampered || disk.Badges[badgeStone] != earned {
		t.Errorf("on disk after the import: %+v, tampered %v/%v", disk.Badges, disk.Tampered, disk.BadgesTampered)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(blob, &raw); err != nil {
		t.Fatal(err)
	}
	reblob := func(change func(m map[string]json.RawMessage)) []byte {
		m := map[string]json.RawMessage{}
		for k, v := range raw {
			m[k] = v
		}
		change(m)
		out, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	// A badge typed into the badge store breaks its signature.
	var store map[string]json.RawMessage
	if err := json.Unmarshal(raw["badge_store"], &store); err != nil {
		t.Fatalf("the export carries no badge store: %v", err)
	}
	store["badges"] = json.RawMessage(`{"age.modern_age":{"at":1}}`)
	forgedStore, _ := json.Marshal(store)
	if _, err := ImportAccountExport(reblob(func(m map[string]json.RawMessage) { m["badge_store"] = forgedStore }), true); err == nil {
		t.Error("an export with a badge typed in was accepted")
	}
	// Another account's badge store, validly signed for that account, is not this one's.
	var otherRaw map[string]json.RawMessage
	if err := json.Unmarshal(otherBlob, &otherRaw); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportAccountExport(reblob(func(m map[string]json.RawMessage) { m["badge_store"] = otherRaw["badge_store"] }), true); err == nil {
		t.Error("an export carrying another account's badges was accepted")
	}
	// Without the badge store it is the export a build from before badges
	// writes and reads: every other key is one that build signs.
	plain := reblob(func(m map[string]json.RawMessage) { delete(m, "badge_store") })
	for key := range raw {
		if key != "badge_store" && !slices.Contains([]string{"version", "account_id", "display_name", "unlocks", "stats", "achievements", "prefs", "tampered", "_sig"}, key) {
			t.Errorf("the export holds the key %q outside the badge store, which a build from before badges does not sign", key)
		}
	}
	isolateAccountDir(t)
	if _, err := ImportAccountExport(plain, true); err != nil {
		t.Errorf("the export without its badge store was refused: %v", err)
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
		Version: 1, AccountID: acct.AccountID,
		Stats: AccountStats{HighestAge: "medieval_age"},
		BadgeStore: &badgeFile{
			Version: badgeFileVersion, AccountID: acct.AccountID,
			Badges: map[string]BadgeEarned{
				badgeStone:     {At: 100, Run: "first"},
				badgePrestige1: {At: 150},
			},
			Counters: map[string]float64{"prestige": 5, "built.lineage.housing": 10},
			Days:     []string{"2026-10-01", "2026-10-02"},
		},
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

	// An export from before badges carries no badge store. It says nothing about
	// badges, so even a replace leaves the account's as they are.
	local()
	older := &progressExport{Version: 1, AccountID: acct.AccountID, Stats: AccountStats{HighestAge: "stone_age"}}
	acct.applyExportLocked(older, false)
	if !hasBadge(acct, badgeIron) || !hasBadge(acct, badgeStone) || acct.Counters["built.lineage.housing"] != 40 {
		t.Errorf("replacing with an export from before badges cost the account its badges: %v %v", acct.Badges, acct.Counters)
	}
	if acct.Stats.HighestAge != "stone_age" {
		t.Errorf("the replace did not take the export's stats: %+v", acct.Stats)
	}

	// A backup whose badge file was flagged flags the account's, merged or replaced.
	local()
	backup.BadgeStore.Tampered = true
	acct.applyExportLocked(backup, true)
	if !acct.BadgesTampered || acct.Tampered {
		t.Errorf("a flagged badge store: badge file tampered %v, account tampered %v; want true and false", acct.BadgesTampered, acct.Tampered)
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

// earnSome gives a fresh account a dated badge, a counter and a prestige
// through an engine, flushed to disk, and returns both.
func earnSome(t *testing.T, name string) (*GameEngine, *Account) {
	t.Helper()
	ge, acct := devEngine(t, name)
	advanceTo(ge, "stone_age")
	for i := 0; i < 3; i++ {
		finishOne(ge, "hut")
	}
	prestigeNow(t, ge)
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	return ge, acct
}

// editBadgeFile rewrites the badge file in slot id by hand, without
// signing it again.
func editBadgeFile(t *testing.T, id string, change func(m map[string]json.RawMessage)) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(slotBadgeFile(t, id), &m); err != nil {
		t.Fatalf("the badge file of %s: %v", id, err)
	}
	change(m)
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir(id), badgeFileName), out, 0644); err != nil {
		t.Fatal(err)
	}
}

// TestAccountFileHoldsNoBadges: the badges are in badges.json, signed for
// the account, and account.json holds nothing a build from before badges
// does not sign.
func TestAccountFileHoldsNoBadges(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")

	assertOldShape(t, slotFile(t, acct.AccountID))
	var bf badgeFile
	if err := json.Unmarshal(slotBadgeFile(t, acct.AccountID), &bf); err != nil {
		t.Fatalf("the badge file: %v", err)
	}
	if bf.Version != badgeFileVersion || bf.AccountID != acct.AccountID || !verifyBadgeFile(&bf) || bf.Tampered {
		t.Errorf("the badge file: version %d, account %s, verifies %v, tampered %v", bf.Version, bf.AccountID, verifyBadgeFile(&bf), bf.Tampered)
	}
	if _, ok := bf.Badges[badgeStone]; !ok || bf.Counters[config.BadgeEvPrestige] != 1 || len(bf.Days) != 1 {
		t.Errorf("the badge file holds %v, %v, %v", bf.Badges, bf.Counters, bf.Days)
	}
	// An account that has earned nothing gets no badge file. (Created by
	// hand: starting a game would note a day.)
	plain, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	if err := plain.SetActiveTheme("bronze"); err != nil {
		t.Fatal(err)
	}
	if b := slotBadgeFile(t, plain.AccountID); b != nil {
		t.Errorf("an account with no badges has a badge file: %s", b)
	}
	// A save that changes no badge leaves the badge file's bytes alone.
	before := slotBadgeFile(t, acct.AccountID)
	if err := acct.SetMapGlyphs("ascii"); err != nil {
		t.Fatal(err)
	}
	if after := slotBadgeFile(t, acct.AccountID); string(after) != string(before) {
		t.Error("saving a pref rewrote the badge file with different bytes")
	}
}

// TestEditedBadgeFileCrossesItsBadges: a badge file changed outside the
// game is flagged and its badges crossed. The flag sticks through saves and
// cannot be taken out by hand, badges earned afterwards are crossed too,
// and account.json is not flagged for it.
func TestEditedBadgeFileCrossesItsBadges(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	accountBefore := slotFile(t, id)

	editBadgeFile(t, id, func(m map[string]json.RawMessage) {
		m["counters"] = json.RawMessage(`{"prestige": 25}`)
	})
	ge, loaded := openIn(t, id)
	if !loaded.BadgesTampered || loaded.Tampered {
		t.Fatalf("after editing the badge file: badge file tampered %v, account tampered %v; want true and false", loaded.BadgesTampered, loaded.Tampered)
	}
	for _, key := range loaded.EarnedBadges() {
		if loaded.Badges[key].Flags&BadgeFlagCrossed == 0 {
			t.Errorf("%s is not crossed in an edited badge file", key)
		}
	}
	// The edit bought the top rung of the ladder; it is crossed, and worth nothing.
	if !hasBadge(loaded, badgePrestige25) {
		t.Fatalf("precondition: the edited counter did not reach the top rung: %v", loaded.EarnedBadges())
	}
	if _, sum := ge.Badges(); sum.Points != 0 {
		t.Errorf("an edited badge file scores %d points", sum.Points)
	}
	if string(slotFile(t, id)) != string(accountBefore) {
		t.Error("an edited badge file changed account.json")
	}

	// The flag is written and stays.
	var bf badgeFile
	if err := json.Unmarshal(slotBadgeFile(t, id), &bf); err != nil || !bf.Tampered || !verifyBadgeFile(&bf) {
		t.Fatalf("the badge file after the load: err %v, tampered %v, verifies %v", err, bf.Tampered, verifyBadgeFile(&bf))
	}
	if again := slotAccount(t, id); !again.BadgesTampered || again.Tampered {
		t.Errorf("after a save and a load: badge file tampered %v, account tampered %v", again.BadgesTampered, again.Tampered)
	}
	// Taking the flag out by hand breaks the signature, which sets it again.
	editBadgeFile(t, id, func(m map[string]json.RawMessage) { delete(m, "tampered") })
	if again := slotAccount(t, id); !again.BadgesTampered {
		t.Error("deleting the tampered key by hand cleared the flag")
	}
	// What is earned from here on is crossed too.
	ge2, flagged := openIn(t, id)
	if err := ge2.StartNewNamedGame("later"); err != nil {
		t.Fatal(err)
	}
	advanceTo(ge2, "stone_age")
	advanceTo(ge2, "bronze_age")
	advanceTo(ge2, "iron_age")
	if e, ok := flagged.Badges[badgeIron]; !ok || e.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("a badge earned into a flagged badge file: %+v (earned %v)", e, ok)
	}

	// Throwing the edited file away does not keep what the edit bought: the
	// badges come back from account.json's record, which never took it (a
	// crossed badge is not written to the old achievements list).
	if err := flagged.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(slotAccount(t, id).Achievements, "prestige_x10") {
		t.Error("the edit reached account.json's achievements list")
	}
	if err := os.Remove(filepath.Join(accountDir(id), badgeFileName)); err != nil {
		t.Fatal(err)
	}
	_, rebuilt := openIn(t, id)
	if hasBadge(rebuilt, badgePrestige10) || hasBadge(rebuilt, badgePrestige25) || rebuilt.BadgesTampered {
		t.Errorf("after deleting the edited file: %v, badge file tampered %v", rebuilt.EarnedBadges(), rebuilt.BadgesTampered)
	}
}

// TestBadgeFileIsBoundToItsAccount: a badge file signed for one account,
// copied into another's slot, does not carry its badges over clean.
func TestBadgeFileIsBoundToItsAccount(t *testing.T) {
	isolateAccountDir(t)
	_, rich := earnSome(t, "Ada")
	poor, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir(poor.AccountID), badgeFileName), slotBadgeFile(t, rich.AccountID), 0644); err != nil {
		t.Fatal(err)
	}
	got := slotAccount(t, poor.AccountID)
	if !got.BadgesTampered || got.Tampered {
		t.Fatalf("another account's badge file: badge file tampered %v, account tampered %v; want true and false", got.BadgesTampered, got.Tampered)
	}
	for _, key := range got.EarnedBadges() {
		if got.Badges[key].Flags&BadgeFlagCrossed == 0 {
			t.Errorf("%s came over uncrossed from another account's file", key)
		}
	}
	if clean := slotAccount(t, rich.AccountID); clean.BadgesTampered {
		t.Error("the account the file was copied from is flagged")
	}
}

// TestEditedAccountLeavesBadgeFileAlone: an edited account.json flags the
// account, as it always did, and does not flag the badge file or cross what
// it already holds. What is earned while the account is flagged is crossed.
func TestEditedAccountLeavesBadgeFileAlone(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	var m map[string]json.RawMessage
	if err := json.Unmarshal(slotFile(t, id), &m); err != nil {
		t.Fatal(err)
	}
	m["stats"] = json.RawMessage(`{"total_prestiges": 99}`)
	edited, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(accountDir(id), accountFileName), edited, 0644); err != nil {
		t.Fatal(err)
	}

	loaded := slotAccount(t, id)
	if !loaded.Tampered || loaded.BadgesTampered {
		t.Fatalf("after editing account.json: account tampered %v, badge file tampered %v; want true and false", loaded.Tampered, loaded.BadgesTampered)
	}
	if e := loaded.Badges[badgeStone]; e.Flags != 0 || e.At == 0 {
		t.Errorf("a badge earned before the edit was changed: %+v", e)
	}
}

// TestMissingBadgeFileIsRebuiltFromTheRecord: without its badge file an
// account keeps what account.json proves (its old achievements, the ages
// up to its highest, its prestiges), undated and uncrossed. That is all an
// account that was only ever opened by a build from before badges has.
func TestMissingBadgeFileIsRebuiltFromTheRecord(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	if err := os.Remove(filepath.Join(accountDir(id), badgeFileName)); err != nil {
		t.Fatal(err)
	}
	_, again := openIn(t, id)
	if again.Tampered || again.BadgesTampered {
		t.Errorf("a missing badge file flagged the account: %v %v", again.Tampered, again.BadgesTampered)
	}
	if got := again.EarnedBadges(); !slices.Equal(got, []string{badgeStone, badgePrestige1}) {
		t.Errorf("rebuilt from the record: %v, want the Stone Age badge and the first prestige", got)
	}
	for _, key := range again.EarnedBadges() {
		if e := again.Badges[key]; e.At != 0 || e.Flags != 0 {
			t.Errorf("%s rebuilt as %+v, want undated and uncrossed", key, e)
		}
	}
	if slotBadgeFile(t, id) == nil {
		t.Error("the rebuilt badges were not written")
	}
}

// TestUnreadableBadgeFileIsSetAside: a badge file that does not parse is
// never overwritten: it is moved aside when there are badges to write.
func TestUnreadableBadgeFileIsSetAside(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	path := filepath.Join(accountDir(id), badgeFileName)
	if err := os.WriteFile(path, []byte("{ not json"), 0644); err != nil {
		t.Fatal(err)
	}
	accountBefore := slotFile(t, id)

	_, again := openIn(t, id) // rebuilds from the record and saves
	if again.Tampered {
		t.Error("a damaged badge file flagged account.json")
	}
	if set, err := os.ReadFile(path + ".corrupt"); err != nil || string(set) != "{ not json" {
		t.Errorf("the damaged file was not set aside: %q, %v", set, err)
	}
	var bf badgeFile
	if err := json.Unmarshal(slotBadgeFile(t, id), &bf); err != nil || !verifyBadgeFile(&bf) {
		t.Errorf("the new badge file: err %v, verifies %v", err, verifyBadgeFile(&bf))
	}
	if string(slotFile(t, id)) != string(accountBefore) {
		t.Error("a damaged badge file changed account.json")
	}
}

// TestBackupCopiesBadgeFile: a full backup of a slot holds the badge file.
func TestBackupCopiesBadgeFile(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	dir, err := BackupAccount(acct.AccountID)
	if err != nil {
		t.Fatalf("BackupAccount: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, badgeFileName))
	if err != nil || string(got) != string(slotBadgeFile(t, acct.AccountID)) {
		t.Errorf("the backup's badge file: %v (same bytes: %v)", err, string(got) == string(slotBadgeFile(t, acct.AccountID)))
	}
}

// TestSwitchAndRecoverKeepBadges: switching to an account, and recovering
// the code of one that is on this machine, open it with its badge file.
func TestSwitchAndRecoverKeepBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, ada := earnSome(t, "Ada")
	bea, err := ge.CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	if len(bea.EarnedBadges()) != 0 {
		t.Fatalf("a new account holds badges: %v", bea.EarnedBadges())
	}
	if _, err := ge.SwitchAccount(ada.AccountID); err != nil {
		t.Fatalf("SwitchAccount: %v", err)
	}
	if got := ge.Account(); got.AccountID != ada.AccountID || !hasBadge(got, badgeStone) || got.Badges[badgeStone].At == 0 {
		t.Errorf("after switching back: %s holds %v", got.AccountID, got.Badges)
	}
	if _, err := ge.SwitchAccount(bea.AccountID); err != nil {
		t.Fatal(err)
	}
	got, _, err := ge.RecoverAccount(ada.RecoveryCode())
	if err != nil {
		t.Fatalf("RecoverAccount: %v", err)
	}
	if !hasBadge(got, badgeStone) || got.BadgesTampered || got.Tampered {
		t.Errorf("after recovering Ada's code: badges %v, flags %v %v", got.EarnedBadges(), got.Tampered, got.BadgesTampered)
	}
	if b := slotBadgeFile(t, bea.AccountID); b != nil {
		t.Errorf("Bea, who earned nothing, has a badge file: %s", b)
	}
}

// TestRenamedAccountCarriesItsBadges: naming an unnamed account gives it a
// new ID. Its badges come along and are signed for the new ID, the one way
// a badge file changes owner.
func TestRenamedAccountCarriesItsBadges(t *testing.T) {
	isolateAccountDir(t)
	legacy, err := LoadOrCreate()
	if err != nil {
		t.Fatal(err)
	}
	recordAge(t, legacy, "stone_age")
	if err := legacy.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	earned := legacy.Badges[badgeStone]

	named, err := CreateNamedAccount("Carthage")
	if err != nil {
		t.Fatal(err)
	}
	if named.AccountID == legacy.AccountID {
		t.Fatal("precondition: the named account kept the old ID")
	}
	disk := slotAccount(t, named.AccountID)
	if disk.BadgesTampered || disk.Tampered || disk.Badges[badgeStone] != earned {
		t.Errorf("the named account: badges %+v, flags %v %v; want %+v, unflagged", disk.Badges, disk.Tampered, disk.BadgesTampered, earned)
	}
}

// TestBadgeFileSurvivesALostAccountFile: if account.json is gone and the
// account is made again under the same name, the badge file still in its
// slot is the account's. It is read back, and saving the new account does
// not write over it.
func TestBadgeFileSurvivesALostAccountFile(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	earned := acct.Badges[badgeStone]
	if err := os.Remove(filepath.Join(accountDir(id), accountFileName)); err != nil {
		t.Fatal(err)
	}

	again, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	if again.AccountID != id {
		t.Fatalf("precondition: the same name gave another ID")
	}
	if again.Badges[badgeStone] != earned || again.BadgesTampered {
		t.Errorf("the new account did not take up its badge file: %+v, tampered %v", again.Badges, again.BadgesTampered)
	}
	// Earning and saving keeps what was there.
	recordAge(t, again, "iron_age")
	if err := again.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	disk := slotAccount(t, id)
	if disk.Badges[badgeStone] != earned || !hasBadge(disk, badgeIron) || disk.BadgesTampered || disk.Tampered {
		t.Errorf("on disk: %+v, flags %v %v", disk.Badges, disk.Tampered, disk.BadgesTampered)
	}

	// The same through a recovery code.
	isolateAccountDir(t)
	_, bea := earnSome(t, "Bea")
	code, beaStone := bea.RecoveryCode(), bea.Badges[badgeStone]
	if err := os.Remove(filepath.Join(accountDir(bea.AccountID), accountFileName)); err != nil {
		t.Fatal(err)
	}
	recovered, err := ImportRecoveryCode(code)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Badges[badgeStone] != beaStone {
		t.Errorf("recovering the ID did not take up the badge file in its slot: %+v", recovered.Badges)
	}
}
