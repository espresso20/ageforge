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
// hooks, UnlockTheme, ExportProgress, SaveGame) before any of this existed:
// one that had played and one that had not. Every test copies a fixture
// into a temp data root (isolateAccountDir) and never touches data/.

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
	noErr(t, os.MkdirAll(accountDir(head.AccountID), 0755))
	noErr(t, os.WriteFile(filepath.Join(accountDir(head.AccountID), accountFileName), raw, 0644))
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

// accountFileKeys is every key account.json may hold. The badges are in the
// badge file, not here.
var accountFileKeys = []string{
	"version", "format", "account_id", "display_name", "created", "last_seen",
	"unlocks", "stats", "achievements", "prefs", "_sig", "tampered",
}

// assertAccountShape fails if raw, an account.json, holds a key outside
// accountFileKeys or is not in the current format.
func assertAccountShape(t *testing.T, raw []byte) {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("account.json does not parse: %v", err)
	}
	for key := range m {
		if !slices.Contains(accountFileKeys, key) {
			t.Errorf("account.json holds the key %q, which does not belong in it", key)
		}
	}
	if string(m["version"]) != "1" {
		t.Errorf("account.json is version %s, want 1", m["version"])
	}
	if string(m["format"]) != "4" {
		t.Errorf("account.json is format %s, want 4", m["format"])
	}
}

// storedBadges parses a badge file's bytes.
func storedBadges(t *testing.T, raw []byte) badgeFile {
	t.Helper()
	var bf badgeFile
	if err := json.Unmarshal(raw, &bf); err != nil {
		t.Fatalf("badge file: %v", err)
	}
	return bf
}

// withFormat returns raw, an account.json written before 4.0, as the game writes it
// now: in the current format and signed again.
func withFormat(t *testing.T, raw []byte) []byte {
	t.Helper()
	var acct Account
	noErr(t, json.Unmarshal(raw, &acct))
	acct.Format = accountFormat
	acct.Signature = signAccount(&acct)
	out, err := json.MarshalIndent(&acct, "", "  ")
	noErr(t, err)
	return out
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
	noErr(t, json.Unmarshal(was, &a))
	noErr(t, json.Unmarshal(now, &b))
	for _, key := range accountFileKeys {
		if key == "unlocks" || key == "_sig" {
			continue
		}
		if string(a[key]) != string(b[key]) {
			t.Errorf("gaining a theme changed %q in account.json: was %s, now %s", key, a[key], b[key])
		}
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
	noErr(t, json.Unmarshal(blob, &raw))
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
	noErr(t, json.Unmarshal(otherBlob, &otherRaw))
	if _, err := ImportAccountExport(reblob(func(m map[string]json.RawMessage) { m["badge_store"] = otherRaw["badge_store"] }), true); err == nil {
		t.Error("an export carrying another account's badges was accepted")
	}
	// Without the badge store it is an export from before badges: refused.
	plain := reblob(func(m map[string]json.RawMessage) { delete(m, "badge_store") })
	isolateAccountDir(t)
	if _, err := ImportAccountExport(plain, true); err == nil || err.Error() != errBeforeFour().Error() {
		t.Errorf("the export without its badge store: %v, want the refusal for a file from before 4.0", err)
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

	// A backup whose badge file was flagged flags the account's, merged or replaced.
	local()
	backup.BadgeStore.Tampered = true
	acct.applyExportLocked(backup, true)
	if !acct.BadgesTampered || acct.Tampered {
		t.Errorf("a flagged badge store: badge file tampered %v, account tampered %v; want true and false", acct.BadgesTampered, acct.Tampered)
	}
}

// TestRunFactsRoundTrip: a run's facts are saved and load equal; a save with
// nothing recorded writes no run_facts key.
func TestRunFactsRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	finishOne(ge, "hut")
	sellSome(ge, "hut", 1)
	ge.mu.RLock()
	want := ge.runFacts.clone()
	ge.mu.RUnlock()
	if want.Counts[config.BadgeEvAgeReached] != 1 || want.Counts[config.BadgeEvBuildingSold] != 1 {
		t.Fatalf("facts of a new run: %+v", want)
	}
	noErr(t, ge.SaveGame("run"))
	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	noErr(t, ge2.LoadGame("run"))
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
	noErr(t, acct.FlushIfDirty())
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
	noErr(t, os.WriteFile(filepath.Join(accountDir(id), badgeFileName), out, 0644))
}

// TestAccountFileHoldsNoBadges: the badges are in badges.json, signed for
// the account, and account.json holds nothing a build from before badges
// does not sign.
func TestAccountFileHoldsNoBadges(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")

	assertAccountShape(t, slotFile(t, acct.AccountID))
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
	// An account that has earned nothing still has one, empty and signed.
	plain, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	noErr(t, plain.SetActiveTheme("bronze"))
	var empty badgeFile
	noErr(t, json.Unmarshal(slotBadgeFile(t, plain.AccountID), &empty))
	if !verifyBadgeFile(&empty) || empty.AccountID != plain.AccountID || len(empty.Badges) != 0 || len(empty.Counters) != 0 {
		t.Errorf("the badge file of an account with no badges: %+v", empty)
	}
	// A save that changes no badge leaves the badge file's bytes alone.
	before := slotBadgeFile(t, acct.AccountID)
	noErr(t, acct.SetMapGlyphs("ascii"))
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
	noErr(t, ge2.StartNewNamedGame("later"))
	advanceTo(ge2, "stone_age")
	advanceTo(ge2, "bronze_age")
	advanceTo(ge2, "iron_age")
	if e, ok := flagged.Badges[badgeIron]; !ok || e.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("a badge earned into a flagged badge file: %+v (earned %v)", e, ok)
	}

	// Throwing the edited file away does not keep what the edit bought: the
	// file starts again empty and marked (TestMissingBadgeFileMarksTheAccount).
	noErr(t, flagged.FlushIfDirty())
	noErr(t, os.Remove(filepath.Join(accountDir(id), badgeFileName)))
	if again := slotAccount(t, id); len(again.Badges) != 0 || !again.BadgesTampered {
		t.Errorf("after deleting the edited file: badges %v, badge file tampered %v; want none and marked", again.Badges, again.BadgesTampered)
	}
}

// TestMissingBadgeFileMarksTheAccount: a saved account always has a badge
// file, so one that is missing was taken away. It reads as an edited one
// would: the account loads, flagged for its badges and not as a whole, and the
// next save starts the file again, empty and marked. What is earned from then
// on is crossed. Tampering marks; it never locks a player out.
func TestMissingBadgeFileMarksTheAccount(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	noErr(t, os.Remove(filepath.Join(accountDir(id), badgeFileName)))

	loaded := slotAccount(t, id)
	if !loaded.BadgesTampered || loaded.Tampered || len(loaded.Badges) != 0 {
		t.Fatalf("a missing badge file: badge file tampered %v, account tampered %v, badges %v; want marked, not flagged, none", loaded.BadgesTampered, loaded.Tampered, loaded.Badges)
	}
	if slotBadgeFile(t, id) != nil {
		t.Error("reading the account wrote the badge file")
	}

	loaded.judge(coreBook(), Event{Kind: config.BadgeEvAgeReached, Subject: "stone_age"}, badgeCtx{})
	if e, ok := loaded.Badges[badgeStone]; !ok || e.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("a badge earned after the file went missing: %+v (earned %v), want crossed", e, ok)
	}
	noErr(t, loaded.FlushIfDirty())
	bf := storedBadges(t, slotBadgeFile(t, id))
	if !verifyBadgeFile(&bf) || !bf.Tampered || bf.AccountID != id {
		t.Errorf("the new badge file: verifies %v, marked %v, account %s", verifyBadgeFile(&bf), bf.Tampered, bf.AccountID)
	}
	if again := slotAccount(t, id); !again.BadgesTampered || again.Tampered || again.Badges[badgeStone].Flags&BadgeFlagCrossed == 0 {
		t.Errorf("the mark did not stick: badge file tampered %v, account tampered %v", again.BadgesTampered, again.Tampered)
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
	noErr(t, os.WriteFile(filepath.Join(accountDir(poor.AccountID), badgeFileName), slotBadgeFile(t, rich.AccountID), 0644))
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
	noErr(t, json.Unmarshal(slotFile(t, id), &m))
	m["stats"] = json.RawMessage(`{"total_prestiges": 99}`)
	edited, _ := json.MarshalIndent(m, "", "  ")
	noErr(t, os.WriteFile(filepath.Join(accountDir(id), accountFileName), edited, 0644))

	loaded := slotAccount(t, id)
	if !loaded.Tampered || loaded.BadgesTampered {
		t.Fatalf("after editing account.json: account tampered %v, badge file tampered %v; want true and false", loaded.Tampered, loaded.BadgesTampered)
	}
	if e := loaded.Badges[badgeStone]; e.Flags != 0 || e.At == 0 {
		t.Errorf("a badge earned before the edit was changed: %+v", e)
	}
}

// TestAccountFromBefore4IsRefused: an account without the current format was
// written before 4.0, played or not. It is refused with the one line saves from
// before 4.0 get, is not listed, and is left on disk as it is; an account the
// game saves carries the format.
func TestAccountFromBefore4IsRefused(t *testing.T) {
	for _, fixture := range []string{"account_v1_full.json", "account_v1_empty.json"} {
		isolateAccountDir(t)
		id, _ := installV1(t, fixture)
		if acct, found, err := loadAccountFromSlot(id); acct != nil || found || err == nil || err.Error() != errBeforeFour().Error() {
			t.Fatalf("%s: %v %v %v, want the refusal", fixture, acct, found, err)
		}
		noErr(t, makeActive(id))
		if _, found, err := LoadAccount(); found || err == nil {
			t.Errorf("%s: LoadAccount found it (found %v, err %v)", fixture, found, err)
		}
		if got := ListAccounts(); len(got) != 0 {
			t.Errorf("%s is listed: %+v", fixture, got)
		}
		if _, err := SwitchAccount(id); err == nil {
			t.Errorf("switching to %s was accepted", fixture)
		}
		if on, _ := os.ReadFile(filepath.Join(accountDir(id), accountFileName)); len(on) == 0 {
			t.Errorf("%s was removed", fixture)
		}
	}

	isolateAccountDir(t)
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	if acct.Format != accountFormat {
		t.Errorf("a new account has format %d, want %d", acct.Format, accountFormat)
	}
	if _, found, err := loadAccountFromSlot(acct.AccountID); !found || err != nil {
		t.Errorf("an account the game saved: found %v, err %v, want it to load", found, err)
	}
}

// TestUnreadableBadgeFileIsSetAside: a badge file that does not parse is
// never overwritten: it is moved aside when there are badges to write.
func TestUnreadableBadgeFileIsSetAside(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	id := acct.AccountID
	path := filepath.Join(accountDir(id), badgeFileName)
	noErr(t, os.WriteFile(path, []byte("{ not json"), 0644))
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
	if len(slotAccount(t, bea.AccountID).Badges) != 0 {
		t.Errorf("Bea, who earned nothing, holds badges: %v", slotAccount(t, bea.AccountID).Badges)
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
	noErr(t, os.Remove(filepath.Join(accountDir(id), accountFileName)))

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
	noErr(t, again.FlushIfDirty())
	disk := slotAccount(t, id)
	if disk.Badges[badgeStone] != earned || !hasBadge(disk, badgeIron) || disk.BadgesTampered || disk.Tampered {
		t.Errorf("on disk: %+v, flags %v %v", disk.Badges, disk.Tampered, disk.BadgesTampered)
	}

	// The same through a recovery code.
	isolateAccountDir(t)
	_, bea := earnSome(t, "Bea")
	code, beaStone := bea.RecoveryCode(), bea.Badges[badgeStone]
	noErr(t, os.Remove(filepath.Join(accountDir(bea.AccountID), accountFileName)))
	recovered, err := ImportRecoveryCode(code)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Badges[badgeStone] != beaStone {
		t.Errorf("recovering the ID did not take up the badge file in its slot: %+v", recovered.Badges)
	}
}
