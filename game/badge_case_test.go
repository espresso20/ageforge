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

// What the badge case reads from the engine: the art and the rewards on a
// badge's view, the list kept between changes, the titles a score holds,
// the themes two integrity badges give, and the motion setting.

const badgeSource = "special.touched_by_the_source"

// freshViews is the badge list built from scratch, past the kept one.
func freshViews(ge *GameEngine) ([]BadgeView, BadgeSummary) {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	sight := ge.ageSightLocked()
	ge.account.mu.Lock()
	defer ge.account.mu.Unlock()
	return ge.account.buildBadgeViewsLocked(ge.badges, sight)
}

// TestBadgeListIsKeptAndNeverStale: the badge list is built once and
// handed out again for as long as nothing it shows has changed, and it is
// built again the moment something has. After every kind of change the
// kept list is, field for field, the list built from scratch.
func TestBadgeListIsKeptAndNeverStale(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	same := func(a, b []BadgeView) bool { return len(a) > 0 && len(b) > 0 && &a[0] == &b[0] }
	check := func(step string) []BadgeView {
		t.Helper()
		got, sum := ge.Badges()
		want, wantSum := freshViews(ge)
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(sum, wantSum) {
			t.Fatalf("after %s the kept list is stale:\n got %+v\nwant %+v", step, got, want)
		}
		again, _ := ge.Badges()
		if !same(got, again) {
			t.Errorf("after %s the list was built twice with nothing changed", step)
		}
		if st := ge.GetState(); st.AccountStats == nil || !same(st.AccountStats.Badges, got) {
			t.Errorf("after %s the snapshot built a list of its own", step)
		}
		return got
	}
	last := check("a new game")
	changed := func(step string) {
		t.Helper()
		now := check(step)
		if same(last, now) {
			t.Errorf("%s changed the badges and the list was not built again", step)
		}
		last = now
	}
	advanceTo(ge, "stone_age")
	changed("an age (a badge earned, and the next age in sight)")
	finishOne(ge, "hut")
	changed("a counted build (a ladder's progress)")
	for i := 0; i < 14; i++ {
		finishOne(ge, "hut")
	}
	changed("a rung earned")
	prestigeNow(t, ge)
	changed("a prestige")
	ge.NoteDevUnlocked()
	changed("an integrity badge")
	// A tick changes nothing the list shows.
	before, _ := ge.Badges()
	ge.doTick()
	if after := check("a tick"); !same(before, after) {
		t.Error("a tick that earned nothing built the list again")
	}

	// An import folds another copy's badges in: the list follows.
	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}
	other, _ := devEngine(t, "Bea")
	advanceTo(other, "stone_age")
	if _, err := ImportAccountExport(blob, true); err != nil {
		t.Fatal(err)
	}
	check("an import")

	// A load from disk starts from the file, with a list of its own.
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	ge2, _ := openIn(t, acct.AccountID)
	got, _ := ge2.Badges()
	want, _ := freshViews(ge2)
	if !reflect.DeepEqual(got, want) || len(earnedKeys(got)) == 0 {
		t.Errorf("a loaded account's list:\n got %+v\nwant %+v", got, want)
	}
}

// TestBadgeViewsCarryArtAndRewards: a badge in sight carries its emblem,
// its tier as a number, its ladder and rung and what it gives. A withheld
// badge carries none of it: an emblem or a ladder's name would say what
// the badge is about.
func TestBadgeViewsCarryArtAndRewards(t *testing.T) {
	isolateAccountDir(t)
	ge, _ := devEngine(t, "Ada")
	views, sum := ge.Badges()

	stone := viewOf(t, views, badgeStone)
	if stone.Emblem != "centre.0" || stone.Level != config.BadgeBronze || stone.Ladder != "" {
		t.Errorf("the Stone Age badge: %+v", stone)
	}
	for i, key := range []string{badgePrestige1, badgePrestige3, badgePrestige10, badgePrestige25} {
		v := viewOf(t, views, key)
		if v.Ladder != "Prestiges" || v.Rung != i+1 || v.Rungs != 4 || v.Emblem != "star" {
			t.Errorf("%s: ladder %q rung %d of %d, emblem %q", key, v.Ladder, v.Rung, v.Rungs, v.Emblem)
		}
	}
	if v := viewOf(t, views, badgePrestige25); v.Level != config.BadgeLegendary || v.Tier != "legendary" {
		t.Errorf("the top rung: %+v", v)
	}
	for i, key := range []string{badgeHousing1, badgeHousing2} {
		v := viewOf(t, views, key)
		if v.Ladder != "Housing" || v.Rung != i+1 || v.Rungs != 5 || v.Emblem != "lineage.housing" {
			t.Errorf("%s: ladder %q rung %d of %d, emblem %q", key, v.Ladder, v.Rung, v.Rungs, v.Emblem)
		}
	}
	if v := viewOf(t, views, badgeHutHoarder); v.Emblem != "hut" || v.Ladder != "" || v.Rung != 0 {
		t.Errorf("Hut Hoarder: %+v", v)
	}

	// Withheld: the Modern Age badge (a later age) and the secret one.
	for _, key := range []string{badgeModern, badgeSale} {
		v := viewOf(t, views, key)
		if !v.Hidden {
			t.Fatalf("precondition: %s is in sight on a new account", key)
		}
		if v.Emblem != "" || v.Ladder != "" || v.Rung != 0 || v.RewardTheme != "" || v.RewardTitle != "" || v.Name != BadgeHiddenName {
			t.Errorf("the withheld %s gives something away: %+v", key, v)
		}
	}

	// A new account holds the first title, and is told the next.
	if sum.Title != "Settler" || sum.TitleRank != 0 || sum.NextTitle != "Headman" || sum.NextTitleAt != 250 {
		t.Errorf("a new account's title: %+v", sum)
	}

	// The integrity badges show once earned, with their sprite's emblem
	// and the theme they give.
	ge.ReportForTest(config.BadgeEvSaveElite, "")
	views, _ = ge.Badges()
	src := viewOf(t, views, badgeSource)
	if !src.Earned || !src.Integrity || src.Emblem != "source" || src.RewardTheme != "source" || src.Points != 0 || src.Level != config.BadgeNoTier {
		t.Errorf("Touched by the Source: %+v", src)
	}
}

// TestBadgeTitles: the title a score holds, the next one and what it asks
// for, and the title of an account that holds every badge that counts.
func TestBadgeTitles(t *testing.T) {
	set := rules.Core()
	for _, c := range []struct {
		points   int
		complete bool
		title    string
		rank     int
		next     string
		at       int
	}{
		{0, false, "Settler", 0, "Headman", 250},
		{249, false, "Settler", 0, "Headman", 250},
		{250, false, "Headman", 1, "Magistrate", 1000},
		{1245, false, "Magistrate", 2, "Sovereign", 3000},
		{3000, false, "Sovereign", 3, "Paragon", 6000},
		{6000, false, "Paragon", 4, "Eternal", 10000},
		{13345, false, "Eternal", 5, "", 0},
		{13345, true, "Completionist", 6, "", 0},
	} {
		title, rank, next, at := set.BadgeTitle(c.points, c.complete)
		if title != c.title || rank != c.rank || next != c.next || at != c.at {
			t.Errorf("%d points (complete %v): %q rank %d, then %q at %d; want %q rank %d, then %q at %d",
				c.points, c.complete, title, rank, next, at, c.title, c.rank, c.next, c.at)
		}
	}
	titles := config.BadgeScoreTitles()
	if titles[0].Points != 0 {
		t.Error("the first title asks for points: an account would hold none")
	}
	for i := 1; i < len(titles); i++ {
		if titles[i].Points <= titles[i-1].Points {
			t.Errorf("the titles are not in order at %s", titles[i].Title)
		}
	}

	// Completionist takes every badge that counts, none of them crossed.
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	book := coreBook()
	grantAll := func(crossOne bool) {
		acct.mu.Lock()
		defer acct.mu.Unlock()
		acct.Badges = map[string]BadgeEarned{}
		for i := range book.defs {
			if def := &book.defs[i]; !def.Integrity() {
				acct.grantLocked(book, def, badgeCtx{}, true, maxBadgeChain)
			}
		}
		if crossOne {
			e := acct.Badges[badgeStone]
			e.Flags |= BadgeFlagCrossed
			acct.Badges[badgeStone] = e
			acct.badgeRev++
		}
	}
	grantAll(false)
	if _, sum := ge.Badges(); sum.Title != config.BadgeCompleteTitle || sum.NextTitle != "" {
		t.Errorf("every badge earned: %+v", sum)
	}
	grantAll(true)
	if _, sum := ge.Badges(); sum.Title == config.BadgeCompleteTitle {
		t.Errorf("a crossed badge still completes the account: %+v", sum)
	}
}

// TestHackerThemesComeWithTheirBadges: Touched by the Source gives the
// Source theme and Creative Accounting gives Glitch, as the badge is
// earned. An account that already held the badge when it began to give a
// theme gets the theme on its next load. account.json keeps its shape:
// the themes are two more words in a list it already holds.
func TestHackerThemesComeWithTheirBadges(t *testing.T) {
	for _, c := range []struct{ badge, event, theme string }{
		{badgeSource, config.BadgeEvSaveElite, "source"},
		{badgeAccounting, config.BadgeEvSaveModified, "glitch"},
	} {
		isolateAccountDir(t)
		def, ok := rules.Core().Badge(c.badge)
		if !ok || !def.Integrity() || def.Reward.Theme != c.theme || def.Event != c.event {
			t.Fatalf("%s: %+v", c.badge, def)
		}
		ge, acct := devEngine(t, "Ada")
		if acct.HasTheme(c.theme) {
			t.Fatalf("precondition: a new account has %s", c.theme)
		}
		ge.ReportForTest(c.event, "")
		if !hasBadge(acct, c.badge) || !acct.HasTheme(c.theme) {
			t.Errorf("%s: earned %v, theme %v", c.badge, hasBadge(acct, c.badge), acct.HasTheme(c.theme))
		}
		if err := acct.FlushIfDirty(); err != nil {
			t.Fatal(err)
		}
		assertOldShape(t, slotFile(t, acct.AccountID))

		// The badge held from before it gave a theme: take the theme away,
		// as an account saved by the build before this one has it.
		acct.mu.Lock()
		acct.Unlocks.Themes = slices.DeleteFunc(acct.Unlocks.Themes, func(k string) bool { return k == c.theme })
		acct.mu.Unlock()
		if err := acct.Save(); err != nil {
			t.Fatal(err)
		}
		if slotAccount(t, acct.AccountID).HasTheme(c.theme) {
			t.Fatal("precondition: the theme is still on the saved account")
		}
		_, again := openIn(t, acct.AccountID)
		if !again.HasTheme(c.theme) {
			t.Errorf("an account that already held %s did not get %s on its next load", c.badge, c.theme)
		}
		if again.Tampered || again.BadgesTampered {
			t.Errorf("giving the theme flagged the account: %v %v", again.Tampered, again.BadgesTampered)
		}
	}
}

// TestMotionSettingLivesBesideTheAccountFile: motion is on by default.
// Turning it off writes settings.json in the account's slot and leaves
// account.json byte for byte as it was, so a build from before the setting
// still reads the account as its own. The setting comes back on a load and
// goes into a backup.
func TestMotionSettingLivesBesideTheAccountFile(t *testing.T) {
	isolateAccountDir(t)
	_, acct := earnSome(t, "Ada")
	if !acct.MotionOn() {
		t.Fatal("motion is off on a new account")
	}
	settings := filepath.Join(accountDir(acct.AccountID), settingsFileName)
	if _, err := os.Stat(settings); !os.IsNotExist(err) {
		t.Fatalf("an account that never set anything has a settings file (%v)", err)
	}
	before := slotFile(t, acct.AccountID)
	badges := slotBadgeFile(t, acct.AccountID)
	if err := acct.SetMotion(false); err != nil {
		t.Fatal(err)
	}
	if acct.MotionOn() {
		t.Error("motion is still on")
	}
	if got := slotFile(t, acct.AccountID); string(got) != string(before) {
		t.Error("setting motion rewrote account.json")
	}
	if got := slotBadgeFile(t, acct.AccountID); string(got) != string(badges) {
		t.Error("setting motion rewrote the badge file")
	}
	assertOldShape(t, slotFile(t, acct.AccountID))
	var file map[string]any
	data, err := os.ReadFile(settings)
	if err != nil || json.Unmarshal(data, &file) != nil || file["motion"] != "off" {
		t.Fatalf("settings.json: %s (%v)", data, err)
	}

	// A fresh load reads it; the account is not flagged.
	loaded := slotAccount(t, acct.AccountID)
	if loaded.MotionOn() || loaded.Tampered {
		t.Errorf("after a load: motion on %v, flagged %v", loaded.MotionOn(), loaded.Tampered)
	}
	// And a save of the account leaves the setting where it is.
	if err := loaded.Save(); err != nil {
		t.Fatal(err)
	}
	if slotAccount(t, acct.AccountID).MotionOn() {
		t.Error("saving the account lost the setting")
	}

	dir, err := BackupAccount(acct.AccountID)
	if err != nil {
		t.Fatalf("BackupAccount: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dir, settingsFileName)); err != nil || string(got) != string(data) {
		t.Errorf("the backup's settings file: %v", err)
	}

	// On again: the file says so by saying nothing.
	if err := loaded.SetMotion(true); err != nil {
		t.Fatal(err)
	}
	if !slotAccount(t, acct.AccountID).MotionOn() {
		t.Error("motion did not come back on")
	}
	// A settings file that does not parse reads as the defaults.
	if err := os.WriteFile(settings, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if !slotAccount(t, acct.AccountID).MotionOn() {
		t.Error("a damaged settings file turned motion off")
	}
}
