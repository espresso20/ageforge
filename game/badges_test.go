package game

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// The badge engine: what earns a badge, what does not, and what the player
// is shown. Every test runs in a temp data root (isolateAccountDir, through
// devEngine), never data/.

// advanceTo moves a test engine to age the way play does: through
// advanceAge, under the lock.
func advanceTo(ge *GameEngine, age string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.advanceAge(age)
}

// finishOne completes one copy of a building as the build queue does.
func finishOne(ge *GameEngine, key string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.finishBuild(BuildQueueItem{BuildingKey: key})
}

// sellSome reports n copies sold, with the count lowered as a sale lowers it.
func sellSome(ge *GameEngine, key string, n int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Buildings.RemoveBuilding(key, n)
	ge.noteSold(key, n)
}

// incidental reports whether key is a badge an advance earns on the side:
// the looks the account wears as it advances (theme, map style, glyphs) and
// the awakening an age brings. The tests about other badges set them aside.
func incidental(key string) bool {
	for _, p := range []string{"theme.", "map.", "awakening."} {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// mainKeys is keys without the incidental badges.
func mainKeys(keys []string) []string {
	return slices.DeleteFunc(slices.Clone(keys), incidental)
}

// mainViews is views without the incidental badges.
func mainViews(views []BadgeView) []BadgeView {
	return slices.DeleteFunc(slices.Clone(views), func(v BadgeView) bool { return incidental(v.Key) })
}

func counter(a *Account, name string) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Counters[name]
}

// TestAgeBadgeEarnedOnAdvance: reaching an age earns its badge once, dated
// and filed under the save it was earned in, and queues it for the toast.
func TestAgeBadgeEarnedOnAdvance(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	if hasBadge(acct, badgeStone) {
		t.Fatal("a new account holds the Stone Age badge")
	}
	advanceTo(ge, "stone_age")
	if !hasBadge(acct, badgeStone) {
		t.Fatalf("reaching the Stone Age did not earn %s", badgeStone)
	}
	acct.mu.Lock()
	e := acct.Badges[badgeStone]
	acct.mu.Unlock()
	if e.At == 0 || e.Run != "run" || e.Flags != 0 {
		t.Errorf("the badge is %+v, want dated, earned in the save \"run\" and unflagged", e)
	}

	all := ge.DrainEarnedBadges()
	earned := mainViews(all)
	if len(earned) != 1 || earned[0].Key != badgeStone || earned[0].Name != "Rock Solid" || earned[0].Hidden {
		t.Fatalf("drained %+v, want the Stone Age badge with its text", earned)
	}
	// The advance also earns what the account wore as it advanced.
	for _, key := range []string{"theme.forge", "map.style.roguelike", "map.glyphs.unicode"} {
		if !slices.ContainsFunc(all, func(v BadgeView) bool { return v.Key == key }) {
			t.Errorf("advancing in the default looks did not earn %s: %v", key, earnedKeys(all))
		}
	}
	if got, want := BadgeLogLine(earned[0]), "Badge earned: Rock Solid (bronze). Reach the Stone Age."; got != want {
		t.Errorf("log line %q, want %q", got, want)
	}
	if again := ge.DrainEarnedBadges(); len(again) != 0 {
		t.Errorf("the badge was announced twice: %+v", again)
	}

	// Reaching it again (a later run) earns nothing new.
	advanceTo(ge, "stone_age")
	if again := ge.DrainEarnedBadges(); len(again) != 0 {
		t.Errorf("an earned badge was earned again: %+v", again)
	}
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if disk := slotAccount(t, acct.AccountID); !hasBadge(disk, badgeStone) || disk.Version != accountSchemaVersion || disk.Tampered {
		t.Errorf("on disk: version %d, tampered %v, badges %v", disk.Version, disk.Tampered, disk.Badges)
	}
}

// TestPrestigeLadderFromPlay: a real prestige moves the counter and earns
// the first rung, judged against the run that ended.
func TestPrestigeLadderFromPlay(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	prestigeNow(t, ge)
	if !hasBadge(acct, badgePrestige1) || hasBadge(acct, badgePrestige3) {
		t.Fatalf("after one prestige: %v", acct.EarnedBadges())
	}
	if got := counter(acct, config.BadgeEvPrestige); got != 1 {
		t.Errorf("prestige counter %v, want 1", got)
	}
	prestigeNow(t, ge)
	prestigeNow(t, ge)
	if !hasBadge(acct, badgePrestige3) {
		t.Errorf("three prestiges did not earn %s: %v", badgePrestige3, acct.EarnedBadges())
	}
	if stats, _ := acct.LifetimeStats(); stats.TotalPrestiges != 3 {
		t.Errorf("TotalPrestiges %d, want 3", stats.TotalPrestiges)
	}
	// The new run starts with its own facts.
	ge.mu.RLock()
	facts := ge.runFacts
	ge.mu.RUnlock()
	if !facts.Whole || len(facts.Counts) != 0 || facts.AgeTick != 0 {
		t.Errorf("the run after a prestige did not start with fresh facts: %+v", facts)
	}
}

// TestMilestonesJudgedOnTheTick: the tick's report is where milestones are
// judged, and each completion is reported in turn. What a milestone asks
// and gives is unchanged; this only checks the two layers share the report.
func TestMilestonesJudgedOnTheTick(t *testing.T) {
	isolateAccountDir(t)
	ge, _ := devEngine(t, "Ada")

	ge.mu.Lock()
	ge.Buildings.counts["hut"] = 1
	ge.mu.Unlock()
	ge.StepTicks(1)
	if !ge.Milestones.IsCompleted("first_shelter") {
		t.Fatal("a hut did not complete First Shelter on the tick")
	}
	ge.mu.RLock()
	done, tallied := ge.Milestones.CompletedCount(), ge.runFacts.Counts[config.BadgeEvMilestone]
	ge.mu.RUnlock()
	if float64(done) != tallied || done == 0 {
		t.Errorf("%d milestones completed, %v reported", done, tallied)
	}
}

// TestHutHoarderNeedsTheCountAndTheAge: a run badge is earned on its event
// when the run meets it, and not otherwise.
func TestHutHoarderNeedsTheCountAndTheAge(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	ge.mu.Lock()
	ge.Buildings.counts["hut"] = 58
	ge.mu.Unlock()
	finishOne(ge, "hut") // 59
	if hasBadge(acct, badgeHutHoarder) {
		t.Fatal("59 huts earned Hut Hoarder")
	}
	finishOne(ge, "hut") // 60
	if !hasBadge(acct, badgeHutHoarder) {
		t.Fatal("the 60th hut in the Primitive Age did not earn Hut Hoarder")
	}

	// The same count in a later age earns nothing.
	isolateAccountDir(t)
	ge2, acct2 := devEngine(t, "Bea")
	ge2.mu.Lock()
	ge2.age = "stone_age"
	ge2.Buildings.counts["hut"] = 59
	ge2.mu.Unlock()
	finishOne(ge2, "hut")
	if hasBadge(acct2, badgeHutHoarder) {
		t.Error("60 huts standing in the Stone Age earned a Primitive Age badge")
	}
}

// TestLineageLadderCountsABuildOnce: a lifetime build counter moves only
// when a build takes the building past the most the run has built of it,
// so selling and rebuilding adds nothing.
func TestLineageLadderCountsABuildOnce(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	const housing = config.BadgeEvBuiltLineage + ".housing"

	for i := 0; i < 13; i++ {
		finishOne(ge, "hut")
	}
	if got := counter(acct, housing); got != 13 || hasBadge(acct, badgeHousing1) {
		t.Fatalf("after 13 huts: counter %v, rung earned %v", got, hasBadge(acct, badgeHousing1))
	}
	// Sell five and build five back: nothing new was built.
	sellSome(ge, "hut", 5)
	for i := 0; i < 5; i++ {
		finishOne(ge, "hut")
	}
	if got := counter(acct, housing); got != 13 {
		t.Fatalf("selling 5 huts and rebuilding them moved the counter to %v", got)
	}
	finishOne(ge, "hut") // the 14th the run has built
	if got := counter(acct, housing); got != 14 || !hasBadge(acct, badgeHousing1) {
		t.Fatalf("the 14th hut: counter %v, rung earned %v", got, hasBadge(acct, badgeHousing1))
	}
	// A counter no badge names is not kept at all.
	acct.mu.Lock()
	names := sortedKeys(acct.Counters)
	acct.mu.Unlock()
	for _, name := range names {
		if !ge.badges.counters[name] {
			t.Errorf("the account keeps the counter %s, which no badge names", name)
		}
	}
	if !slices.Contains(names, housing) || slices.Contains(names, config.BadgeEvBuilt+".hut") {
		t.Errorf("counters kept: %v, want the housing lineage and not the hut itself", names)
	}

	// The marks are the run's: they are saved with it and come back.
	if err := ge.SaveGame("run"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	if err := ge2.LoadGame("run"); err != nil {
		t.Fatal(err)
	}
	sellSome(ge2, "hut", 3)
	for i := 0; i < 3; i++ {
		finishOne(ge2, "hut")
	}
	if got := counter(acct, housing); got != 14 {
		t.Errorf("after a save and a load, selling and rebuilding moved the counter to %v", got)
	}
}

// TestBuildHooksAreWired: a hut built through the public command and the
// tick loop reaches the counter (the hook is in the real path).
func TestBuildHooksAreWired(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	if err := ge.BuildBuilding("hut"); err != nil {
		t.Fatalf("BuildBuilding: %v", err)
	}
	ge.StepTicks(60)
	if ge.Buildings.GetCount("hut") != 1 {
		t.Fatalf("the hut did not finish: %d", ge.Buildings.GetCount("hut"))
	}
	if got := counter(acct, config.BadgeEvBuiltLineage+".housing"); got != 1 {
		t.Errorf("a hut built in play left the housing counter at %v", got)
	}
	ge.mu.RLock()
	built := ge.runFacts.Counts[config.BadgeEvBuildingBuilt]
	ge.mu.RUnlock()
	if built != 1 {
		t.Errorf("the run tallied %v builds, want 1", built)
	}
	// Selling opens in the Stone Age.
	ge.mu.Lock()
	ge.age = "stone_age"
	ge.mu.Unlock()
	if err := ge.SellBuilding("hut", 1); err != nil {
		t.Fatalf("SellBuilding: %v", err)
	}
	ge.mu.RLock()
	sold := ge.runFacts.Counts[config.BadgeEvBuildingSold]
	ge.mu.RUnlock()
	if sold != 1 {
		t.Errorf("the run tallied %v sales, want 1", sold)
	}
}

// TestSecretBadgeStaysASilhouette: a secret badge lists with its hint and
// no name or description until it is earned.
func TestSecretBadgeStaysASilhouette(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	views, _ := ge.Badges()
	v := viewOf(t, views, badgeSale)
	if !v.Hidden || !v.Secret || v.Name != BadgeHiddenName || v.Desc != "Something about a clearance." {
		t.Fatalf("the secret badge before it is earned: %+v", v)
	}
	ge.mu.Lock()
	ge.noteSold("hut", 99)
	ge.mu.Unlock()
	if hasBadge(acct, badgeSale) {
		t.Fatal("99 sales earned Liquidation Sale")
	}
	ge.mu.Lock()
	ge.noteSold("hut", 1)
	ge.mu.Unlock()
	if !hasBadge(acct, badgeSale) {
		t.Fatal("100 sales in one run did not earn Liquidation Sale")
	}
	views, _ = ge.Badges()
	if v := viewOf(t, views, badgeSale); v.Hidden || v.Name != "Liquidation Sale" || !v.Earned {
		t.Errorf("the secret badge once earned: %+v", v)
	}
}

// TestFashionablyLate: a named predicate over a time threshold, stated as a
// multiple of the age's pacing target.
func TestFashionablyLate(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	need := int(10 * ge.Rules().TargetTicks("primitive_age"))
	if need <= 0 {
		t.Fatal("the Primitive Age has no pacing target")
	}

	ge.mu.Lock()
	ge.tick = need - 5
	ge.mu.Unlock()
	ge.StepTicks(2)
	if hasBadge(acct, badgeLate) {
		t.Fatal("earned before ten times the target had passed")
	}
	ge.StepTicks(5)
	if !hasBadge(acct, badgeLate) {
		t.Fatal("ten times the Primitive Age's target, spent in the Primitive Age, did not earn the badge")
	}

	// A run that does not know when it entered its age cannot earn it.
	isolateAccountDir(t)
	ge2, acct2 := devEngine(t, "Bea")
	ge2.mu.Lock()
	ge2.runFacts = RunFacts{}
	ge2.tick = need + 10
	ge2.mu.Unlock()
	ge2.StepTicks(2)
	if hasBadge(acct2, badgeLate) {
		t.Error("a run with no known entry tick earned a time-in-age badge")
	}
}

// TestDevTouchedRunEarnsBadges: a run the developer console changed earns
// like any other. The game watches what happens; if a badge's condition is
// met, the badge is earned, counted and announced, uncrossed.
func TestDevTouchedRunEarnsBadges(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, acct := devEngine(t, "Dev Tester")

	if out := DevExecCommand("/give food 5", ge); !strings.Contains(out, "gave") {
		t.Fatalf("/give: %q", out)
	}
	if !ge.GetState().DevTouched {
		t.Fatal("precondition: the dev command did not mark the run")
	}
	advanceTo(ge, "stone_age")
	for i := 0; i < 20; i++ {
		finishOne(ge, "hut")
	}
	ge.mu.Lock()
	ge.noteSold("hut", 150)
	ge.mu.Unlock()
	prestigeNow(t, ge)

	want := []string{badgeStone, badgePrestige1, badgeHousing1, badgeSale}
	slices.Sort(want)
	if earned := mainKeys(acct.EarnedBadges()); !slices.Equal(earned, want) {
		t.Errorf("a dev-touched run earned %v, want %v", earned, want)
	}
	for _, key := range acct.EarnedBadges() {
		if e := acct.Badges[key]; e.At == 0 || e.Flags != 0 {
			t.Errorf("%s earned in a dev-touched run is %+v, want dated and uncrossed", key, e)
		}
	}
	if got := counter(acct, config.BadgeEvBuiltLineage+".housing"); got != 20 {
		t.Errorf("the housing counter is %v, want 20", got)
	}
	stats, _ := acct.LifetimeStats()
	if stats.TotalPrestiges != 1 || stats.HighestAge != "stone_age" {
		t.Errorf("lifetime stats from a dev-touched run: %+v", stats)
	}
	if pending := mainViews(ge.DrainEarnedBadges()); len(pending) != len(want) {
		t.Errorf("%d toasts queued, want %d", len(pending), len(want))
	}
}

// TestCookieJar: unlocking the developer console earns its own badge. The
// badge is worth nothing, counts toward nothing, and is listed only once
// earned.
func TestCookieJar(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	views, before := ge.Badges()
	for _, v := range views {
		if v.Integrity {
			t.Fatalf("an integrity badge is listed before it is earned: %+v", v)
		}
	}
	ge.NoteDevUnlocked()
	if !hasBadge(acct, badgeCookieJar) {
		t.Fatal("unlocking the console did not earn Hand in the Cookie Jar")
	}
	views, after := ge.Badges()
	v := viewOf(t, views, badgeCookieJar)
	if !v.Earned || !v.Integrity || v.Points != 0 || v.Tier != "" || v.Hidden {
		t.Errorf("the cookie jar badge: %+v", v)
	}
	if !reflect.DeepEqual(after, before) {
		t.Errorf("an integrity badge moved the totals: before %+v, after %+v", before, after)
	}
	// It is earned the same in a run the console has already changed, and
	// unlocking it twice earns it once.
	isolateAccountDir(t)
	withDevMode(t)
	ge2, acct2 := devEngine(t, "Bea")
	DevExecCommand("/fill", ge2)
	ge2.NoteDevUnlocked()
	ge2.NoteDevUnlocked()
	if got := acct2.EarnedBadges(); !slices.Equal(got, []string{badgeCookieJar}) {
		t.Errorf("after unlocking the console the account holds %v, want only the cookie jar", got)
	}
}

// TestModifiedSaveCrossesItsBadges: loading a save edited outside the game
// earns Creative Accounting, and every badge earned in that game is crossed
// and adds no points.
func TestModifiedSaveCrossesItsBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	if err := ge.SaveGame("run"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(accountDir(acct.AccountID), "saves", "run.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["resources"] = json.RawMessage(`{"food": 999999, "wood": 999999}`)
	edited, _ := json.Marshal(raw)
	if err := os.WriteFile(path, edited, 0644); err != nil {
		t.Fatal(err)
	}

	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	if err := ge2.LoadGame("run"); err != nil {
		t.Fatal(err)
	}
	if !ge2.GetState().CheaterBadge {
		t.Fatal("precondition: the edited save did not load as modified")
	}
	if !hasBadge(acct, badgeAccounting) {
		t.Fatal("loading an edited save did not earn Creative Accounting")
	}
	advanceTo(ge2, "stone_age")
	views, sum := ge2.Badges()
	stone := viewOf(t, views, badgeStone)
	if !stone.Earned || !stone.Crossed {
		t.Fatalf("a badge earned in a modified game: %+v", stone)
	}
	// Everything the advance earned is crossed, and a crossed badge is
	// worth nothing.
	crossed := 0
	for _, v := range views {
		if v.Integrity {
			continue
		}
		if v.Earned && !v.Crossed {
			t.Errorf("%s was earned uncrossed in a modified game", v.Key)
		}
		if v.Earned {
			crossed++
		}
	}
	if crossed == 0 || sum.Earned != crossed || sum.Points != 0 {
		t.Errorf("totals with %d crossed badges: %+v, want them earned and 0 points", crossed, sum)
	}
}

// TestEditedAccountCrossesItsBadges: a badge earned on an account whose
// file was edited is crossed too, and the flag still sticks.
func TestEditedAccountCrossesItsBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	acct.mu.Lock()
	acct.Tampered = true
	acct.mu.Unlock()

	advanceTo(ge, "stone_age")
	acct.mu.Lock()
	e := acct.Badges[badgeStone]
	acct.mu.Unlock()
	if e.Flags&BadgeFlagCrossed == 0 {
		t.Errorf("a badge earned on a flagged account is not crossed: %+v", e)
	}
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if disk := slotAccount(t, acct.AccountID); !disk.Tampered || disk.Badges[badgeStone].Flags&BadgeFlagCrossed == 0 {
		t.Errorf("on disk: tampered %v, badge %+v", disk.Tampered, disk.Badges[badgeStone])
	}
}

// TestAnotherAccountsRunEarnsNothing: the run in memory after an account
// change belongs to the account that started it, not the one now held.
func TestAnotherAccountsRunEarnsNothing(t *testing.T) {
	isolateAccountDir(t)
	ge, _ := devEngine(t, "Ada")
	other, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	ge.SetAccount(other) // the run still belongs to Ada

	advanceTo(ge, "stone_age")
	ge.NoteDevUnlocked()
	if got := other.EarnedBadges(); len(got) != 0 {
		t.Errorf("Ada's run earned Bea badges: %v", got)
	}
}

// TestBadgesHideWhatTheSpoilerRulesHide: a locked badge about an age the
// player cannot see named yet reaches the UI with no text in it at all, the
// next age's badge shows, and the totals keep hidden badges out of the
// "of" number.
func TestBadgesHideWhatTheSpoilerRulesHide(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")

	views, sum := ge.Badges()
	if v := viewOf(t, views, badgeStone); v.Hidden || v.Name != "Rock Solid" {
		t.Errorf("the next age's badge should show: %+v", v)
	}
	for _, key := range []string{badgeIron, badgeModern} {
		v := viewOf(t, views, key)
		if !v.Hidden || v.Name != BadgeHiddenName || v.Desc != "" {
			t.Errorf("%s should be a silhouette with no text: %+v", key, v)
		}
	}
	// A hidden badge carries nothing that says what it is.
	hidden := 0
	for _, v := range views {
		if !v.Hidden {
			continue
		}
		hidden++
		def, _ := rules.Core().Badge(v.Key)
		// A secret badge shows its one-line hint, and nothing else.
		if hint := v.Desc; (v.Secret && hint != def.Hint) || (!v.Secret && hint != "") {
			t.Errorf("the hidden badge %s shows %q", v.Key, hint)
		}
		if v.Name != BadgeHiddenName || v.Emblem != "" || v.Ladder != "" || v.RewardTheme != "" || v.RewardTitle != "" {
			t.Errorf("the hidden badge %s says something: %+v", v.Key, v)
		}
	}
	// The integrity badges are unlisted until earned; every other badge is
	// shown or hidden, and a new account is told only how many it can see.
	listed := 0
	for _, d := range rules.Core().Badges() {
		if !d.Integrity() {
			listed++
		}
	}
	if len(views) != listed {
		t.Errorf("%d badges listed, want %d (the integrity badges are left out)", len(views), listed)
	}
	if sum.Hidden != hidden || sum.Shown != listed-hidden || hidden == 0 || sum.Earned != 0 || sum.HiddenCounted {
		t.Errorf("totals for a new account: %+v, with %d of %d hidden", sum, hidden, listed)
	}
	// A new account sees the first age's badges and nothing of the last
	// era's: most of the catalog is still ahead of it.
	if sum.Shown >= sum.Hidden {
		t.Errorf("a new account sees %d badges and has %d hidden; most of the catalog should be ahead of it", sum.Shown, sum.Hidden)
	}

	// Reaching the Bronze Age makes the Iron Age the next one: its badge shows.
	advanceTo(ge, "stone_age")
	advanceTo(ge, "bronze_age")
	views, _ = ge.Badges()
	if v := viewOf(t, views, badgeIron); v.Hidden || v.Name != "Age of Iron" || v.Earned {
		t.Errorf("with the Iron Age next, its badge should show: %+v", v)
	}
	if v := viewOf(t, views, badgeModern); !v.Hidden {
		t.Errorf("the Modern Age badge showed in the Bronze Age: %+v", v)
	}
	// What the account has seen stays seen in a new run.
	if err := ge.StartNewNamedGame("second"); err != nil {
		t.Fatal(err)
	}
	views, _ = ge.Badges()
	if v := viewOf(t, views, badgeIron); v.Hidden {
		t.Errorf("a new run hid a badge the account had seen: %+v", v)
	}

	// The size of the hidden part shows only once the last age is reached.
	acct.mu.Lock()
	acct.Stats.HighestAge = "transcendent_age"
	acct.mu.Unlock()
	if _, sum := ge.Badges(); !sum.HiddenCounted {
		t.Errorf("an account that reached the last age still hides the count: %+v", sum)
	}
}

// TestLifetimeProgressShows: a locked ladder rung that shows carries its
// count and its threshold.
func TestLifetimeProgressShows(t *testing.T) {
	isolateAccountDir(t)
	ge, _ := devEngine(t, "Ada")
	for i := 0; i < 3; i++ {
		finishOne(ge, "hut")
	}
	views, _ := ge.Badges()
	if v := viewOf(t, views, badgeHousing1); v.Progress != 3 || v.Target != 14 || v.Earned {
		t.Errorf("Housing Hobbyist after 3 huts: %+v", v)
	}
}

// TestBadgesDoNotChangeTheRun: the same seed and the same play end in the
// same state with an account and without one. Badges are judged from the
// run; nothing flows back.
func TestBadgesDoNotChangeTheRun(t *testing.T) {
	isolateAccountDir(t)
	play := func(ge *GameEngine) map[string]json.RawMessage {
		ge.SeedRNG(11)
		for i := 0; i < 30; i++ {
			_, _ = ge.GatherResource("wood", 3)
			_, _ = ge.GatherResource("food", 3)
		}
		_ = ge.BuildBuilding("hut")
		ge.StepTicks(150)
		_ = ge.BuildBuilding("hut")
		ge.StepTicks(150)
		advanceTo(ge, "stone_age")
		ge.StepTicks(50)
		b, err := ge.StateJSON()
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}
		delete(raw, "account_id") // whose save it is, not what happened in it
		return raw
	}

	bare := NewGameEngine()
	withAccount, acct := devEngine(t, "Ada")
	a, b := play(bare), play(withAccount)
	if len(acct.EarnedBadges()) == 0 {
		t.Fatal("precondition: the account earned nothing, so the test proves nothing")
	}
	for _, key := range sortedKeys(a) {
		if string(a[key]) != string(b[key]) {
			t.Errorf("%s differs with an account:\n  without %s\n  with    %s", key, a[key], b[key])
		}
	}
	if len(a) != len(b) {
		t.Errorf("the saves have different fields: %d without an account, %d with", len(a), len(b))
	}
}

// TestBadgeForEarningBadges: earning a badge is an event of its own, so a
// badge for earning badges is a plain ladder, and integrity badges do not
// count toward it. On a ruleset with one more row; no engine code knows it.
func TestBadgeForEarningBadges(t *testing.T) {
	isolateAccountDir(t)
	src := rules.FromConfig()
	// A catalog of four: the integrity badges, two plain ones and the one
	// that counts them.
	src.BadgeFamilies = nil
	src.Badges = slices.DeleteFunc(src.Badges, func(d config.BadgeDef) bool { return !d.Integrity() })
	src.Badges = append(src.Badges,
		config.BadgeDef{
			Key: "age.stone_age", Family: "age", Subject: "stone_age", Name: "Rock Solid", Desc: "Reach the Stone Age.",
			Tier: config.BadgeBronze, Scope: config.BadgeMoment, Event: config.BadgeEvAgeReached,
			Proof: config.StaticProof(config.BadgeRuleGate),
		},
		config.BadgeDef{
			Key: "ladder.prestiges.1", Family: "ladder", Name: "First Prestige", Desc: "Prestige for the first time.",
			Tier: config.BadgeBronze, Scope: config.BadgeLifetime, Counter: config.BadgeEvPrestige, Threshold: 1,
			Proof: config.StaticProof(config.BadgeRuleLifetime),
		},
		config.BadgeDef{
			Key: "special.collector", Family: "special", Name: "Collector", Desc: "Earn 2 badges.",
			Tier: config.BadgeBronze, Scope: config.BadgeLifetime,
			Counter: config.BadgeEvBadge, Threshold: 2,
			Proof: config.BadgeProof{Kind: config.BadgeProofDerived},
		})
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngineWith(rules.Compile(src))
	ge.SetAccount(acct)
	if err := ge.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}

	ge.NoteDevUnlocked() // an integrity badge: not counted
	advanceTo(ge, "stone_age")
	if hasBadge(acct, "special.collector") {
		t.Fatalf("one countable badge earned Collector: %v", acct.EarnedBadges())
	}
	prestigeNow(t, ge)
	if !hasBadge(acct, "special.collector") {
		t.Fatalf("two countable badges did not earn Collector: %v", acct.EarnedBadges())
	}
	if got := counter(acct, config.BadgeEvBadge); got != 3 {
		t.Errorf("the badge counter is %v, want 3 (the two and Collector itself)", got)
	}
}

// TestConditionsOnARun: the extra conditions a row can carry. "At most"
// needs the run's whole story: a run that was not tracked from its first
// tick cannot show that something never happened.
func TestConditionsOnARun(t *testing.T) {
	isolateAccountDir(t)
	src := rules.FromConfig()
	src.Badges = append(src.Badges,
		config.BadgeDef{
			Key: "special.no_sales", Family: "special", Name: "Keeper", Desc: "Reach the Stone Age without selling a building.",
			Tier: config.BadgeBronze, Scope: config.BadgeRun, Event: config.BadgeEvAgeReached, Subject: "stone_age",
			When:  []config.BadgeCond{{Fact: "run." + config.BadgeEvBuildingSold, Op: config.BadgeAtMost, Value: 0}},
			Proof: config.StaticProof(config.BadgeRuleGate),
		},
		config.BadgeDef{
			Key: "special.braced", Family: "special", Name: "Sandbagged", Desc: "Endure a catastrophe with Brace at level 2.",
			Tier: config.BadgeSilver, Scope: config.BadgeMoment, Event: config.BadgeEvEndured,
			When:  []config.BadgeCond{{Fact: "ev.brace", Op: config.BadgeAtLeast, Value: 2}},
			Proof: config.StaticProof(config.BadgeRuleGate),
		},
	)
	set := rules.Compile(src)
	start := func(name string) (*GameEngine, *Account) {
		acct, err := CreateAccount(name)
		if err != nil {
			t.Fatal(err)
		}
		ge := NewGameEngineWith(set)
		ge.SetAccount(acct)
		if err := ge.StartNewNamedGame("run"); err != nil {
			t.Fatal(err)
		}
		return ge, acct
	}

	clean, a := start("Ada")
	advanceTo(clean, "stone_age")
	if !hasBadge(a, "special.no_sales") {
		t.Error("a run with no sales did not earn the no-sales badge")
	}

	sold, b := start("Bea")
	sellSome(sold, "hut", 1)
	advanceTo(sold, "stone_age")
	if hasBadge(b, "special.no_sales") {
		t.Error("a run with a sale earned the no-sales badge")
	}

	unknown, c := start("Cy")
	unknown.mu.Lock()
	unknown.runFacts = RunFacts{} // as a save from before run facts loads
	unknown.mu.Unlock()
	advanceTo(unknown, "stone_age")
	if hasBadge(c, "special.no_sales") {
		t.Error("a run whose past is unknown earned a badge for something that never happened")
	}

	ev := func(ge *GameEngine, brace float64) {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.report(Event{Kind: config.BadgeEvEndured, Subject: "iron_era", Attrs: map[string]float64{"brace": brace}})
	}
	ev(clean, 1)
	if hasBadge(a, "special.braced") {
		t.Error("Brace 1 earned the Brace 2 badge")
	}
	ev(clean, 2)
	if !hasBadge(a, "special.braced") {
		t.Error("Brace 2 did not earn its badge")
	}
}

// TestListAccountsCountsBadges: the Accounts panel's count is the badges
// that count, read without changing the file.
func TestListAccountsCountsBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	ge.NoteDevUnlocked()
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	before := slotFile(t, acct.AccountID)
	want := 0
	for _, key := range acct.EarnedBadges() {
		if def, _ := rules.Core().Badge(key); !def.Integrity() {
			want++
		}
	}
	list := ge.ListAccounts()
	if len(list) != 1 || want == 0 || list[0].Badges != want {
		t.Errorf("the account list counts %+v, want %d badges (the integrity badge is left out)", list, want)
	}
	if after := slotFile(t, acct.AccountID); string(after) != string(before) {
		t.Error("listing the accounts rewrote an account file")
	}
}

// TestEveryBadgePredicateIsKnown: a badge that names a predicate the engine
// does not have could never be earned.
func TestEveryBadgePredicateIsKnown(t *testing.T) {
	known := BadgePredNames()
	for _, def := range rules.Core().Badges() {
		if def.Pred != "" && !slices.Contains(known, def.Pred) {
			t.Errorf("%s asks for the predicate %q, which the engine does not have (%v)", def.Key, def.Pred, known)
		}
	}
}

// TestEveryDeclaredEventIsReported: config lists the events a badge may be
// judged on. Each must be reported somewhere in the engine, or a badge on
// it could never be earned; and the engine must report nothing config does
// not list, or the guard would call a working badge unreachable.
func TestEveryDeclaredEventIsReported(t *testing.T) {
	fset := token.NewFileSet()
	cfg, err := parser.ParseFile(fset, filepath.Join("..", "config", "badges.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	// The BadgeEv constants, by name and by value.
	names := map[string]string{}
	for _, decl := range cfg.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "BadgeEv") || i >= len(vs.Values) {
					continue
				}
				if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
					names[name.Name] = strings.Trim(lit.Value, `"`)
				}
			}
		}
	}
	if len(names) < 30 {
		t.Fatalf("found only %d BadgeEv constants in config/badges.go", len(names))
	}
	listed := config.BadgeEventKinds()
	for name, kind := range names {
		if !slices.Contains(listed, kind) {
			t.Errorf("config.%s (%q) is not in BadgeEventKinds", name, kind)
		}
	}
	if len(listed) != len(names) {
		t.Errorf("BadgeEventKinds lists %d events, config declares %d", len(listed), len(names))
	}

	// Every one is used in the engine's own code (test files left out).
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for name := range names {
			if strings.Contains(string(src), "config."+name) {
				used[name] = true
			}
		}
	}
	for _, name := range sortedKeys(names) {
		if !used[name] {
			t.Errorf("config.%s is declared but the engine never reports it", name)
		}
	}
}

// TestDayPlayedIsNotedOncePerDay: starting or loading a game notes the
// calendar day on the account, once.
func TestDayPlayedIsNotedOncePerDay(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada") // starts a game
	today := time.Now().Format(accountDayLayout)
	if !slices.Equal(acct.Days, []string{today}) {
		t.Fatalf("days after starting a game: %v, want [%s]", acct.Days, today)
	}
	if err := ge.SaveGame("run"); err != nil {
		t.Fatal(err)
	}
	if err := ge.LoadGame("run"); err != nil {
		t.Fatal(err)
	}
	if err := ge.StartNewNamedGame("again"); err != nil {
		t.Fatal(err)
	}
	if len(acct.Days) != 1 {
		t.Errorf("the same day was noted again: %v", acct.Days)
	}
	// The day is the account's, never the run's: the run's facts must read
	// the same on any account and any date.
	ge.mu.RLock()
	_, tallied := ge.runFacts.Counts[config.BadgeEvDayPlayed]
	ge.mu.RUnlock()
	if tallied {
		t.Error("the day was tallied in the run's facts")
	}
}

// TestBadgeThemeReward: a badge that gives a theme unlocks it on the
// account when it is earned. On a ruleset with one more row.
func TestBadgeThemeReward(t *testing.T) {
	isolateAccountDir(t)
	src := rules.FromConfig()
	src.Badges = append(src.Badges, config.BadgeDef{
		Key: "special.bronze_look", Family: "special", Name: "Bronze Look", Desc: "Reach the Stone Age.",
		Tier: config.BadgeBronze, Scope: config.BadgeMoment, Event: config.BadgeEvAgeReached, Subject: "stone_age",
		Reward: config.BadgeReward{Theme: "bronze"},
		Proof:  config.StaticProof(config.BadgeRuleGate),
	})
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngineWith(rules.Compile(src))
	ge.SetAccount(acct)
	if err := ge.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	if acct.HasTheme("bronze") {
		t.Fatal("precondition: the account already has the theme")
	}
	advanceTo(ge, "stone_age")
	if !acct.HasTheme("bronze") {
		t.Error("earning the badge did not unlock its theme")
	}
	advanceTo(ge, "stone_age")
	if got := acct.UnlockedThemes(); len(got) != 1 {
		t.Errorf("the theme is unlocked %d times: %v", len(got), got)
	}
}

// TestSessionEventsStayOutOfTheRun: the events that are about the session
// or the account are judged, and never tallied in the run's facts.
func TestSessionEventsStayOutOfTheRun(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	ge.NoteDevUnlocked()
	advanceTo(ge, "stone_age") // earns a badge
	ge.StepTicks(3)
	if !hasBadge(acct, badgeCookieJar) || !hasBadge(acct, badgeStone) {
		t.Fatalf("precondition: %v", acct.EarnedBadges())
	}
	ge.mu.RLock()
	facts := ge.runFacts.clone()
	ge.mu.RUnlock()
	for _, ev := range config.BadgeSessionEvents() {
		for name := range facts.Counts {
			if name == ev || strings.HasPrefix(name, ev+".") {
				t.Errorf("the run's facts tally the session event %s: %v", ev, facts.Counts)
			}
		}
	}
	if facts.Counts[config.BadgeEvAgeReached] != 1 {
		t.Errorf("the run's own events are tallied: %v", facts.Counts)
	}
}
