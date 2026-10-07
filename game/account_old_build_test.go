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

// Two builds on one machine: a build of the game from before badges and
// this one, taking turns with the same account. The older build must never
// be able to mark a healthy account as modified, and nothing it does may
// cost the account a badge.
//
// The fixtures under testdata/account_split are a real round trip:
//
//   - new/ was written by this code: account.json, the badge file beside
//     it, and the account's export.
//   - after_old_resave/account.json, after_old_play/account.json and
//     after_old_import/account.json were written by master's own code from
//     before badges (commit 310f3a5, run from a copy of that tree against a
//     temp data root holding new/): it loaded the account and changed a
//     setting; it played and earned two old-style achievements; it imported
//     new/export.json. That run itself failed if the old code read the
//     account as modified, changed or removed the badge file, or refused
//     the export.
//
// These tests are the other half: this code loading what the old code left.
// Every test copies fixtures into a temp data root (isolateAccountDir) and
// never touches data/.

const splitDir = "testdata/account_split"

// installSplit puts an account.json fixture and the new-layout badge file
// into a slot of the temp data root, and returns the ID and both files'
// bytes. badges == false leaves the badge file out.
func installSplit(t *testing.T, accountFile string, badges bool) (id string, accountRaw, badgesRaw []byte) {
	t.Helper()
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(splitDir, name))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		return b
	}
	accountRaw = read(accountFile)
	var head struct {
		AccountID string `json:"account_id"`
	}
	if err := json.Unmarshal(accountRaw, &head); err != nil {
		t.Fatalf("fixture %s: %v", accountFile, err)
	}
	id = head.AccountID
	if err := os.MkdirAll(accountDir(id), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir(id), accountFileName), accountRaw, 0644); err != nil {
		t.Fatal(err)
	}
	if badges {
		badgesRaw = read("new/badges.json")
		if err := os.WriteFile(filepath.Join(accountDir(id), badgeFileName), badgesRaw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return id, accountRaw, badgesRaw
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

// TestNewLayoutFixtureIsCurrent: the new/ fixture is what this code still
// writes and reads. If the layout changes, the round trip must be run again
// with the old code, not patched by hand.
func TestNewLayoutFixtureIsCurrent(t *testing.T) {
	isolateAccountDir(t)
	id, accountRaw, badgesRaw := installSplit(t, "new/account.json", true)
	assertOldShape(t, accountRaw)
	acct := slotAccount(t, id)
	if acct.Tampered || acct.BadgesTampered {
		t.Fatalf("the fixture loads flagged: %v %v", acct.Tampered, acct.BadgesTampered)
	}
	// Saving it writes the same bytes.
	acct.mu.Lock()
	err := acct.Save()
	acct.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if string(slotFile(t, id)) != string(accountRaw) || string(slotBadgeFile(t, id)) != string(badgesRaw) {
		t.Error("this code writes the fixture account differently: run the round trip again with the old code")
	}
}

// TestOldBuildResaveKeepsBadges: the older build loaded the account,
// changed a setting and saved it, never seeing the badge file. Loaded here
// afterwards the account is unflagged, every badge is still earned and
// dated as before, and neither file is rewritten, on the first load or the
// second.
func TestOldBuildResaveKeepsBadges(t *testing.T) {
	isolateAccountDir(t)
	id, accountRaw, badgesRaw := installSplit(t, "after_old_resave/account.json", true)
	before := storedBadges(t, badgesRaw)
	if len(before.Badges) < 4 {
		t.Fatalf("precondition: the fixture holds %d badges", len(before.Badges))
	}

	for load := 1; load <= 2; load++ {
		ge, acct := openIn(t, id)
		if acct.Tampered || acct.BadgesTampered {
			t.Fatalf("load %d: the account reads as modified after the older build saved it (account %v, badge file %v)", load, acct.Tampered, acct.BadgesTampered)
		}
		if !reflect.DeepEqual(acct.Badges, before.Badges) {
			t.Errorf("load %d: badges\n got %+v\nwant %+v", load, acct.Badges, before.Badges)
		}
		for key, e := range acct.Badges {
			if e.At == 0 || e.Run == "" || e.Flags != 0 {
				t.Errorf("load %d: %s is %+v, want dated, with its run, uncrossed", load, key, e)
			}
		}
		if !reflect.DeepEqual(acct.Counters, before.Counters) || !slices.Equal(acct.Days, before.Days) {
			t.Errorf("load %d: counters %v days %v, want %v %v", load, acct.Counters, acct.Days, before.Counters, before.Days)
		}
		// What the older build changed is there.
		if _, glyphs, _ := acct.MapPrefs(); glyphs != "ascii" {
			t.Errorf("load %d: the older build's setting was lost: map glyphs %q", load, glyphs)
		}
		if earned := ge.DrainEarnedBadges(); len(earned) != 0 {
			t.Errorf("load %d announced badges: %+v", load, earned)
		}
		if err := acct.FlushIfDirty(); err != nil {
			t.Fatal(err)
		}
		if got := slotFile(t, id); string(got) != string(accountRaw) {
			t.Errorf("load %d rewrote account.json:\nwas %s\nnow %s", load, accountRaw, got)
		}
		if got := slotBadgeFile(t, id); string(got) != string(badgesRaw) {
			t.Errorf("load %d rewrote the badge file:\nwas %s\nnow %s", load, badgesRaw, got)
		}
	}
}

// TestOldBuildAchievementsBecomeBadges is the reverse case: with the badge
// file sitting in the slot, the older build played on and earned two
// old-style achievements (reached_modern, prestige_x10) and nine more
// prestiges. This build picks all of it up: the badges are granted, undated
// and uncrossed, the counter catches up, what was earned before keeps its
// date, account.json is left as the older build wrote it, and a second load
// writes identical bytes.
func TestOldBuildAchievementsBecomeBadges(t *testing.T) {
	isolateAccountDir(t)
	id, accountRaw, badgesRaw := installSplit(t, "after_old_play/account.json", true)
	before := storedBadges(t, badgesRaw)

	ge, acct := openIn(t, id)
	if acct.Tampered || acct.BadgesTampered {
		t.Fatalf("the account reads as modified after the older build played it (account %v, badge file %v)", acct.Tampered, acct.BadgesTampered)
	}
	want := []string{badgeIron, badgeModern, badgeStone, badgePrestige1, badgePrestige3, badgePrestige10, badgeHousing1}
	slices.Sort(want)
	if got := acct.EarnedBadges(); !slices.Equal(got, want) {
		t.Fatalf("badges %v, want %v", got, want)
	}
	for key, e := range acct.Badges {
		if was, held := before.Badges[key]; held {
			if e != was {
				t.Errorf("%s, earned before the older build's turn, changed: %+v, was %+v", key, e, was)
			}
			continue
		}
		if e.At != 0 || e.Run != "" || e.Flags != 0 {
			t.Errorf("%s, picked up from the older build, is %+v; want undated and uncrossed", key, e)
		}
	}
	if got := counter(acct, config.BadgeEvPrestige); got != 10 {
		t.Errorf("the prestige counter is %v, want the older build's 10", got)
	}
	if got := counter(acct, config.BadgeEvBuiltLineage+".housing"); got != before.Counters[config.BadgeEvBuiltLineage+".housing"] {
		t.Errorf("a counter the older build knows nothing about changed: %v", got)
	}
	if earned := ge.DrainEarnedBadges(); len(earned) != 0 {
		t.Errorf("picking up the older build's play announced badges: %+v", earned)
	}

	// account.json is the older build's, untouched; the badge file caught up.
	if got := slotFile(t, id); string(got) != string(accountRaw) {
		t.Errorf("account.json was rewritten:\nwas %s\nnow %s", accountRaw, got)
	}
	first := slotBadgeFile(t, id)
	if string(first) == string(badgesRaw) {
		t.Fatal("the badge file was not updated with what the older build earned")
	}
	if bf := storedBadges(t, first); !verifyBadgeFile(&bf) || bf.AccountID != id || bf.Tampered {
		t.Errorf("the updated badge file: verifies %v, account %s, tampered %v", verifyBadgeFile(&bf), bf.AccountID, bf.Tampered)
	}

	// A second load changes nothing.
	ge2, again := openIn(t, id)
	if err := again.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if again.Tampered || again.BadgesTampered || len(ge2.DrainEarnedBadges()) != 0 {
		t.Errorf("the second load: flags %v %v", again.Tampered, again.BadgesTampered)
	}
	if got := slotFile(t, id); string(got) != string(accountRaw) {
		t.Error("the second load rewrote account.json")
	}
	if got := slotBadgeFile(t, id); string(got) != string(first) {
		t.Errorf("the second load rewrote the badge file:\nfirst  %s\nsecond %s", first, got)
	}
}

// TestOldBuildImportsNewExport: the older build imported an export this
// build made (the badge store rides in it under its own signature, outside
// the part the older build signs). The account it wrote is healthy, holds
// everything the older build knows, and gets its record's badges here.
func TestOldBuildImportsNewExport(t *testing.T) {
	isolateAccountDir(t)
	id, importedRaw, _ := installSplit(t, "after_old_import/account.json", false)
	assertOldShape(t, importedRaw)
	source, err := os.ReadFile(filepath.Join(splitDir, "new/account.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want, got Account
	if err := json.Unmarshal(source, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(importedRaw, &got); err != nil {
		t.Fatal(err)
	}
	if !verifyAccount(&got) || got.Tampered {
		t.Fatalf("the account the older build imported: verifies %v, tampered %v", verifyAccount(&got), got.Tampered)
	}
	if got.AccountID != want.AccountID || got.DisplayName != want.DisplayName ||
		!reflect.DeepEqual(got.Stats, want.Stats) || !reflect.DeepEqual(got.Unlocks, want.Unlocks) ||
		!slices.Equal(got.Achievements, want.Achievements) || !reflect.DeepEqual(got.Prefs, want.Prefs) {
		t.Errorf("the older build's import lost something:\n got %+v %+v %v %+v\nwant %+v %+v %v %+v",
			got.Stats, got.Unlocks, got.Achievements, got.Prefs, want.Stats, want.Unlocks, want.Achievements, want.Prefs)
	}

	// The older build does not know the badge file, so it wrote none. This
	// build gives the account what its record proves.
	_, acct := openIn(t, id)
	if acct.Tampered || acct.BadgesTampered {
		t.Errorf("flags after loading the older build's import: %v %v", acct.Tampered, acct.BadgesTampered)
	}
	if earned := acct.EarnedBadges(); !slices.Equal(earned, []string{badgeIron, badgeStone, badgePrestige1}) {
		t.Errorf("badges from the record: %v", earned)
	}
}

// TestBothBuildsTakeTurns: the whole loop on one account. This build earns,
// the older build's saved file replaces account.json (its turn), this build
// earns again. Nothing is flagged or crossed at any point, and what this
// build earns shows in the older build's list.
func TestBothBuildsTakeTurns(t *testing.T) {
	isolateAccountDir(t)
	id, _, _ := installSplit(t, "after_old_play/account.json", true)
	ge, acct := openIn(t, id)
	if err := ge.StartNewNamedGame("turn"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 15; i++ {
		prestigeNow(t, ge)
	}
	if !hasBadge(acct, badgePrestige25) {
		t.Fatalf("precondition: 25 prestiges did not earn the top rung: %v", acct.EarnedBadges())
	}
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	assertOldShape(t, slotFile(t, id))
	disk := slotAccount(t, id)
	if disk.Tampered || disk.BadgesTampered || disk.Stats.TotalPrestiges != 25 {
		t.Fatalf("after this build's turn: flags %v %v, %d prestiges", disk.Tampered, disk.BadgesTampered, disk.Stats.TotalPrestiges)
	}
	for key, e := range disk.Badges {
		if e.Flags != 0 {
			t.Errorf("%s is crossed on a healthy account: %+v", key, e)
		}
	}

	// The older build's turn: its own earlier file comes back (as if it had
	// been running all along and saved last). Fewer prestiges than the
	// counter now holds; nothing is lost or flagged for it.
	older, err := os.ReadFile(filepath.Join(splitDir, "after_old_play/account.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountDir(id), accountFileName), older, 0644); err != nil {
		t.Fatal(err)
	}
	_, back := openIn(t, id)
	if back.Tampered || back.BadgesTampered {
		t.Fatalf("after the older build's turn: flags %v %v", back.Tampered, back.BadgesTampered)
	}
	if !hasBadge(back, badgePrestige25) || counter(back, config.BadgeEvPrestige) != 25 {
		t.Errorf("a badge was lost to the older build's save: %v, counter %v", back.EarnedBadges(), counter(back, config.BadgeEvPrestige))
	}
}
