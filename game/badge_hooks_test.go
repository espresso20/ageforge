package game

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// The reports the full catalog added (badge_hooks.go and the hook sites in
// the engine): each is driven here the way play drives it, and earns the
// badge that reads it. Every test runs in a temp data root
// (isolateAccountDir), never data/.

// say reports an event with attributes as the engine does, under its lock.
func say(ge *GameEngine, kind, subject string, attrs map[string]float64) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.report(Event{Kind: kind, Subject: subject, Attrs: attrs})
}

// runTally reads one of the run's tallies.
func runTally(ge *GameEngine, name string) float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.runFacts.Counts[name]
}

// want fails unless the account holds every badge named, and none of the
// ones after "not".
func wantBadges(t *testing.T, acct *Account, when string, held []string, not ...string) {
	t.Helper()
	for _, key := range held {
		if _, ok := rules.Core().Badge(key); !ok {
			t.Fatalf("no badge %s in the catalog", key)
		}
		if !hasBadge(acct, key) {
			t.Errorf("%s: %s was not earned (holds %v)", when, key, mainKeys(acct.EarnedBadges()))
		}
	}
	for _, key := range not {
		if hasBadge(acct, key) {
			t.Errorf("%s: %s was earned", when, key)
		}
	}
}

// TestProductionCountsTowardTheResourceLadders: what the tick's rates
// produce is batched and told to the account, so a resource's ladder
// climbs; a game with no account does none of the work; and the batches are
// the account's, never the run's facts.
func TestProductionCountsTowardTheResourceLadders(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	finishOne(ge, "hut")
	finishOne(ge, "hut")
	finishOne(ge, "gathering_camp")
	ge.StepTicks(300) // settlers arrive and go to work
	const food = config.BadgeEvProduced + ".food"
	got := counter(acct, food)
	if got <= 0 {
		t.Fatalf("300 ticks with a staffed gathering camp counted no food: %v", acct.Counters)
	}
	// The count is what the rates made, tick by tick: a batch every
	// BadgeProducedTicks ticks, so it is never more than what was stored
	// and eaten, and it grows.
	ge.StepTicks(config.BadgeProducedTicks * 3)
	if more := counter(acct, food); more <= got {
		t.Errorf("three more batches did not move the count: %v then %v", got, more)
	}
	if n := runTally(ge, food) + runTally(ge, config.BadgeEvProduced); n != 0 {
		t.Errorf("production is tallied in the run's facts (%v): it must be the account's alone", n)
	}
	// The first rung of a ladder is reachable this way.
	rung, ok := rules.Core().Badge("resource.food.1")
	if !ok || rung.Counter != food {
		t.Fatalf("the food ladder's first rung: %+v", rung)
	}
	ge.mu.Lock()
	ge.tell(Event{Kind: config.BadgeEvProduced, Subject: "food", N: rung.Threshold})
	ge.mu.Unlock()
	wantBadges(t, acct, "a rung's worth of food", []string{"resource.food.1"}, "resource.food.2")

	// No account: nothing is batched at all.
	bare := NewGameEngine()
	if err := bare.StartNewNamedGame("bare"); err != nil {
		t.Fatal(err)
	}
	finishOne(bare, "hut")
	finishOne(bare, "gathering_camp")
	bare.StepTicks(200)
	if bare.produced != nil {
		t.Errorf("a game with no account batches production: %v", bare.produced)
	}
}

// TestTimeAwayCountsItsProduction: what the time away produced is counted
// once, as the offline gains, and the return is reported with how long it
// was, whether the allowance cut it short, and what the plan did.
func TestTimeAwayCountsItsProduction(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	finishOne(ge, "hut")
	finishOne(ge, "hut")
	finishOne(ge, "gathering_camp")
	ge.StepTicks(300)
	const food = config.BadgeEvProduced + ".food"
	before := counter(acct, food)

	ge.mu.Lock()
	ge.applyOfflineProgress(2 * time.Hour)
	ge.mu.Unlock()
	if after := counter(acct, food); after <= before {
		t.Errorf("two hours away counted no food: %v then %v", before, after)
	}
	if n := runTally(ge, config.BadgeEvReturned); n != 1 {
		t.Errorf("the return was tallied %v times", n)
	}
	wantBadges(t, acct, "a short time away", nil, "special.maximum_leave")

	ge.mu.Lock()
	ge.applyOfflineProgress(MaxOfflineTime + time.Hour)
	ge.mu.Unlock()
	wantBadges(t, acct, "away past the allowance", []string{"special.maximum_leave"}, "special.while_you_were_out")

	// While You Were Out: the allowance used in full, the plan having
	// started ten things and finished.
	say(ge, config.BadgeEvReturned, "", map[string]float64{"ticks": 9000, "capped": 1, "plan_started": 9, "plan_left": 0})
	wantBadges(t, acct, "nine plan starts", nil, "special.while_you_were_out")
	say(ge, config.BadgeEvReturned, "", map[string]float64{"ticks": 9000, "capped": 1, "plan_started": 10, "plan_left": 2})
	wantBadges(t, acct, "a plan left unfinished", nil, "special.while_you_were_out")
	say(ge, config.BadgeEvReturned, "", map[string]float64{"ticks": 9000, "capped": 1, "plan_started": 10, "plan_left": 0})
	wantBadges(t, acct, "ten plan starts and an empty plan", []string{"special.while_you_were_out"})
}

// TestCensusReadsThePayroll: every BadgeCensusTicks ticks the account is
// told how many workers each domain has at work, and a payroll badge is
// earned when its domain has enough. The census is the account's: the run's
// facts do not tally it.
func TestCensusReadsThePayroll(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	first, _ := rules.Core().Badge("domain.food.1")
	second, _ := rules.Core().Badge("domain.food.2")
	if first.Threshold <= 0 || second.Threshold <= first.Threshold {
		t.Fatalf("the food payroll's rungs: %v then %v", first.Threshold, second.Threshold)
	}
	staff := func(n int) {
		ge.mu.Lock()
		ge.Workers.domains["worker"].assignments["gathering_camp"] = n
		ge.mu.Unlock()
		// To the next census.
		for i := 0; i < config.BadgeCensusTicks; i++ {
			ge.mu.Lock()
			ge.tick++
			ge.noteCensus()
			ge.mu.Unlock()
		}
	}
	staff(int(first.Threshold) - 1)
	wantBadges(t, acct, "one worker short", nil, "domain.food.1")
	staff(int(first.Threshold))
	wantBadges(t, acct, "the first payroll", []string{"domain.food.1"}, "domain.food.2", "domain.lumber.1")
	staff(int(second.Threshold))
	wantBadges(t, acct, "the second payroll", []string{"domain.food.2"})
	if n := runTally(ge, config.BadgeEvCensus); n != 0 {
		t.Errorf("the census is tallied in the run's facts (%v)", n)
	}
}

// TestCensusSeesEveryStoreFull: Everything Full asks for every capped
// resource the run has unlocked to be full at once, from the Industrial Age
// on. A resource that is spent as it is made (a flow resource) is left out:
// it cannot be saved up.
func TestCensusSeesEveryStoreFull(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	census := func() {
		ge.mu.Lock()
		ge.tick = (ge.tick/config.BadgeCensusTicks + 1) * config.BadgeCensusTicks
		ge.noteCensus()
		ge.mu.Unlock()
	}
	fill := func(short string) {
		ge.mu.Lock()
		for _, key := range ge.Resources.order {
			r := ge.Resources.resources[key]
			if !ge.Resources.unlocked[key] || r.Storage <= 0 {
				continue
			}
			r.Amount = r.Storage
			if key == short {
				r.Amount = r.Storage - 1
			}
		}
		ge.mu.Unlock()
	}
	fill("")
	census()
	wantBadges(t, acct, "every store full in the first age", nil, "special.everything_full")

	ge.mu.Lock()
	ge.age = "industrial_age"
	ge.applyAgeUnlocks("industrial_age")
	ge.ageTold = ge.age
	ge.mu.Unlock()
	fill("wood")
	census()
	wantBadges(t, acct, "one store a unit short", nil, "special.everything_full")
	fill("")
	// A flow resource below its cap does not stand in the way.
	ge.mu.Lock()
	flows := 0
	for _, key := range ge.Resources.order {
		if r := ge.Resources.resources[key]; ge.Resources.unlocked[key] && r.Storage > 0 && ge.rules.IsFlowResource(key) {
			r.Amount = 0
			flows++
		}
	}
	ge.mu.Unlock()
	census()
	wantBadges(t, acct, "every store full in the Industrial Age", []string{"special.everything_full"})
	if flows == 0 {
		t.Log("no flow resource is unlocked by the Industrial Age: the exemption was not exercised")
	}
}

// TestWarsAreReported: a war starting, a war ending by waiting it out and
// a war bought off with tribute are each reported once, in the order they
// happened, with how many wars there are and whether tribute ended it.
func TestWarsAreReported(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	const a, b = "ironhold_clans", "riverlands_tribes"
	defs := config.FactionByKey()
	ge.mu.Lock()
	dm := ge.Diplomacy
	dm.factions[a] = &FactionState{Discovered: true, Status: "embargo", Opinion: -90, Provocations: warProvocationThreshold - 1}
	dm.factions[b] = &FactionState{Discovered: true, Status: "embargo", Opinion: -90, Provocations: warProvocationThreshold - 1}
	if !dm.recordProvocation(dm.factions[a], defs[a], 0) {
		t.Fatal("the last provocation did not start a war")
	}
	ge.noteWars(false)
	ge.mu.Unlock()
	if n := runTally(ge, config.BadgeEvWarStarted); n != 1 {
		t.Fatalf("one war started, %v tallied", n)
	}
	wantBadges(t, acct, "one war", nil, "special.popular")

	ge.mu.Lock()
	dm.recordProvocation(dm.factions[b], defs[b], 0)
	ge.noteWars(false)
	ge.mu.Unlock()
	wantBadges(t, acct, "two wars at once", []string{"special.popular"})

	// Bought off: the engine's own tribute path reports the ending.
	ge.mu.Lock()
	ge.Resources.unlocked["gold"], ge.Resources.unlocked["culture"] = true, true
	ge.Resources.resources["gold"].Storage, ge.Resources.resources["gold"].Amount = 1e9, 1e9
	ge.Resources.resources["culture"].Storage, ge.Resources.resources["culture"].Amount = 1e9, 1e9
	ge.mu.Unlock()
	if err := ge.SendTribute(a); err != nil {
		t.Fatalf("SendTribute: %v", err)
	}
	wantBadges(t, acct, "a war bought off", []string{"special.protection_money", "civ." + a + ".peace"}, "special.cold_shoulder")

	// Waited out: the war ends by itself once the grudge has run its course.
	ge.mu.Lock()
	fs := dm.factions[b]
	fs.LastProvocationTick = 0
	for tick := 1; tick < 2000000 && fs.AtWar; tick += 500 {
		dm.processWar(tick, ge.age)
	}
	ended := !fs.AtWar
	ge.noteWars(false)
	ge.mu.Unlock()
	if !ended {
		t.Fatal("the second war never ended by itself")
	}
	wantBadges(t, acct, "a war waited out", []string{"special.cold_shoulder", "civ." + b + ".peace"})
	if n := runTally(ge, config.BadgeEvWarEnded); n != 2 {
		t.Errorf("two wars ended, %v tallied", n)
	}
	// Told once: a second look finds nothing new.
	ge.mu.Lock()
	ge.noteWars(false)
	ge.mu.Unlock()
	if n := runTally(ge, config.BadgeEvWarEnded); n != 2 {
		t.Errorf("an ended war was reported again: %v", n)
	}
}

// TestLendsGiftsAndAlliances: workers on loan, a gift and an alliance with
// everyone met are reported with what their badges read.
func TestLendsGiftsAndAlliances(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	say(ge, config.BadgeEvWorkersLent, "riverlands_tribes", nil)
	wantBadges(t, acct, "workers on loan", []string{"special.regifting"})

	say(ge, config.BadgeEvCivAllied, "riverlands_tribes", map[string]float64{"met": 3, "unallied": 0})
	wantBadges(t, acct, "three met, all allied", []string{"civ.riverlands_tribes.allied"}, "special.everybodys_friend")
	say(ge, config.BadgeEvCivAllied, "ironhold_clans", map[string]float64{"met": 4, "unallied": 1})
	wantBadges(t, acct, "four met, one not allied", nil, "special.everybodys_friend")
	say(ge, config.BadgeEvCivAllied, "merchant_guild", map[string]float64{"met": 4, "unallied": 0})
	wantBadges(t, acct, "four met, all allied", []string{"special.everybodys_friend"})

	// A civilization's regular: five deals from it, across runs; ten from
	// one in a single run is Loyal Customer.
	for i := 0; i < 5; i++ {
		say(ge, config.BadgeEvDeal, "riverlands_tribes", nil)
	}
	wantBadges(t, acct, "five deals from one civilization", []string{"civ.riverlands_tribes.regular"}, "special.loyal_customer", "civ.ironhold_clans.regular")
	for i := 0; i < 9; i++ {
		say(ge, config.BadgeEvDeal, "ironhold_clans", nil)
	}
	wantBadges(t, acct, "nine deals from another", nil, "special.loyal_customer")
	say(ge, config.BadgeEvDeal, "ironhold_clans", nil)
	wantBadges(t, acct, "ten deals from one civilization in one run", []string{"special.loyal_customer"})
}

// TestLooksWornAtAnAdvance: advancing an age tells the account what it is
// wearing, so the theme and map badges are earned for the looks in use, and
// only for those.
func TestLooksWornAtAnAdvance(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	if err := acct.SetActiveTheme("daylight"); err != nil {
		t.Fatal(err)
	}
	if err := acct.SetMapStyle("skyline"); err != nil {
		t.Fatal(err)
	}
	if err := acct.SetMapGlyphs("ascii"); err != nil {
		t.Fatal(err)
	}
	advanceTo(ge, "stone_age")
	wantBadges(t, acct, "an advance in Daylight, skyline, plain glyphs",
		[]string{"theme.daylight", "map.style.skyline", "map.glyphs.ascii"},
		"theme.forge", "map.style.roguelike", "map.glyphs.unicode", "map.glyphs.nerd")
	for _, kind := range []string{config.BadgeEvAdvancedTheme, config.BadgeEvAdvancedStyle, config.BadgeEvAdvancedGlyphs} {
		if n := runTally(ge, kind) + runTally(ge, kind+".daylight") + runTally(ge, kind+".skyline") + runTally(ge, kind+".ascii"); n != 0 {
			t.Errorf("the looks are tallied in the run's facts (%s: %v): they are the account's settings", kind, n)
		}
	}
}

// TestDefaultLooksAreInTheCatalog: an account that chose no theme, map
// style or glyph set wears the defaults, and each has a badge.
func TestDefaultLooksAreInTheCatalog(t *testing.T) {
	th, style, glyphs := (&Account{}).looks()
	if th != defaultThemeKey || !slices.Contains(config.BadgeMapStyles, style) || !slices.Contains(config.BadgeMapGlyphs, glyphs) {
		t.Errorf("the looks of an account that chose none: %s %s %s", th, style, glyphs)
	}
	for _, key := range []string{"theme." + th, "map.style." + style, "map.glyphs." + glyphs} {
		if _, ok := rules.Core().Badge(key); !ok {
			t.Errorf("no badge %s for a default look", key)
		}
	}
}

// TestBadgeExpeditionsMatchTheGame: the catalog's table of expeditions is
// the game's, each in the age it comes with.
func TestBadgeExpeditionsMatchTheGame(t *testing.T) {
	set := rules.Core()
	order := set.Indexes()
	mm := NewMilitaryManager()
	first := map[string]string{}
	for _, age := range set.AgeKeys() {
		for _, x := range mm.GetAvailableExpeditions(age, order) {
			if _, seen := first[x.Key]; !seen {
				first[x.Key] = age
			}
		}
	}
	listed := map[string]bool{}
	for _, x := range config.BadgeExpeditions() {
		listed[x.Key] = true
		if first[x.Key] == "" {
			t.Errorf("the catalog lists the expedition %s, which the game does not have", x.Key)
			continue
		}
		if first[x.Key] != x.MinAge {
			t.Errorf("%s comes in the %s; the catalog says %s", x.Key, first[x.Key], x.MinAge)
		}
		if _, ok := set.Badge("expedition." + x.Key); !ok {
			t.Errorf("no badge for the expedition %s", x.Key)
		}
	}
	for key := range first {
		if !listed[key] {
			t.Errorf("the game has the expedition %s, which the catalog does not list", key)
		}
	}
}

// TestInvitedDoomsAreMarked: Endure and Succumb report whether the doom
// was invited, read from the run's own fate.
func TestInvitedDoomsAreMarked(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	strike := func(invited bool) {
		ge.mu.Lock()
		ge.age = "iron_age"
		ge.ageTold = ge.age
		ge.currentEpoch = "iron_era"
		ge.Buildings.counts["gathering_camp"] = 10
		ge.fate = &FateSave{EpochKey: "iron_era", Fated: true, Arrived: true, Invited: invited}
		ge.pendingCatastrophe = "iron_era"
		ge.mu.Unlock()
	}
	strike(false)
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	wantBadges(t, acct, "a doom endured that nobody asked for", []string{"doom.iron_era.endured", "ladder.endured.1"}, "special.asked_for_it")
	strike(true)
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	wantBadges(t, acct, "an invited doom endured", []string{"special.asked_for_it"})

	strike(false)
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	wantBadges(t, acct, "a doom succumbed to", []string{"doom.iron_era.succumbed", "ladder.succumbed.1"}, "special.signed_the_guest_book")
	strike(true)
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	wantBadges(t, acct, "an invited doom succumbed to", []string{"special.signed_the_guest_book"})
}

// TestEraPace: leaving an era reports the ticks spent in it over the
// pacing targets of its ages, so a swift era is earned inside its target
// and not outside it; a run that does not know when it entered the era (a
// save from before that was kept) earns none.
func TestEraPace(t *testing.T) {
	set := rules.Core()
	stone, _ := set.Era("stone_era")
	target := 0.0
	for _, a := range stone.Ages {
		target += set.TargetTicks(a)
	}
	leave := func(t *testing.T, ticks int, known bool) *Account {
		isolateAccountDir(t)
		ge, acct := devEngine(t, "Ada")
		for _, a := range stone.Ages[1:] {
			advanceTo(ge, a)
		}
		ge.mu.Lock()
		ge.tick = ticks
		if !known {
			ge.runFacts.EraKnown = false
		}
		ge.mu.Unlock()
		advanceTo(ge, "iron_age")
		if n := runTally(ge, config.BadgeEvEraLeft); n != 1 {
			t.Fatalf("leaving the Stone Era was tallied %v times", n)
		}
		return acct
	}
	wantBadges(t, leave(t, int(target), true), "the Stone Era on its target", []string{"swift.stone_era"})
	wantBadges(t, leave(t, int(target)+1, true), "a tick over", nil, "swift.stone_era")
	wantBadges(t, leave(t, 10, false), "an era entered before the run kept track", nil, "swift.stone_era")
}

// TestRunPace: reaching an age reports the run's ticks over the pacing
// targets of the ages behind it, for the badges that ask for a quick start.
func TestRunPace(t *testing.T) {
	set := rules.Core()
	behind := 0.0
	for _, a := range set.AgeKeys() {
		if a == "iron_age" {
			break
		}
		behind += set.TargetTicks(a)
	}
	reach := func(t *testing.T, ticks int) *Account {
		isolateAccountDir(t)
		ge, acct := devEngine(t, "Ada")
		advanceTo(ge, "stone_age")
		advanceTo(ge, "bronze_age")
		ge.mu.Lock()
		ge.tick = ticks
		ge.mu.Unlock()
		advanceTo(ge, "iron_age")
		return acct
	}
	wantBadges(t, reach(t, int(behind)), "the Iron Age on the targets", []string{"special.early_riser", "age.iron_age"})
	wantBadges(t, reach(t, int(behind)+1), "a tick over", []string{"age.iron_age"}, "special.early_riser")
}

// TestKitPurchasesAreReported: buying an item of the legacy kit is
// reported with how many are left to buy, so the first purchase and the
// whole kit each have their badge.
func TestKitPurchasesAreReported(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	kit := ge.rules.LegacyKit()
	if len(kit) < 2 {
		t.Fatalf("the legacy kit has %d items", len(kit))
	}
	ge.mu.Lock()
	ge.Prestige.available = 1e6
	ge.Prestige.level = 9
	ge.mu.Unlock()
	for i, key := range kit {
		if err := ge.BuyPrestigeUpgrade(key); err != nil {
			t.Fatalf("buying %s: %v", key, err)
		}
		if i == 0 {
			wantBadges(t, acct, "the first kit item", []string{"special.maxed_out"}, "special.fully_invested")
		}
	}
	wantBadges(t, acct, "the whole kit", []string{"special.fully_invested"})
}

// TestPlanSizeIsReported: adding to the build plan reports how long it now
// is.
func TestPlanSizeIsReported(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	grow := func(n int) {
		ge.mu.Lock()
		for len(ge.plan) < n {
			ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: "hut", Count: 1})
		}
		ge.notePlanSize()
		ge.mu.Unlock()
	}
	grow(19)
	wantBadges(t, acct, "nineteen items planned", nil, "special.long_game")
	grow(20)
	wantBadges(t, acct, "twenty items planned", []string{"special.long_game"})
}

// TestWonderOverflowIsReported: the first unit the caps put into a wonder's
// bank is reported once a run (Exact Change asks for a run without it),
// and a wonder that finishes from its bank says whether the player was
// away.
func TestWonderOverflowIsReported(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	for i := 0; i < 5; i++ {
		say(ge, config.BadgeEvWonderRaised, "sacred_grove", nil)
	}
	wantBadges(t, acct, "five wonders with no overflow", []string{"special.exact_change"})

	isolateAccountDir(t)
	ge, acct = devEngine(t, "Bea")
	say(ge, config.BadgeEvWonderOverflow, "", nil)
	for i := 0; i < 5; i++ {
		say(ge, config.BadgeEvWonderRaised, "sacred_grove", nil)
	}
	wantBadges(t, acct, "five wonders after overflow", nil, "special.exact_change")

	say(ge, config.BadgeEvWonderBanked, "sacred_grove", map[string]float64{"away": 0})
	wantBadges(t, acct, "a wonder banked at the keyboard", nil, "special.nothing_wasted")
	say(ge, config.BadgeEvWonderBanked, "sacred_grove", map[string]float64{"away": 1})
	wantBadges(t, acct, "a wonder banked while away", []string{"special.nothing_wasted"})
}

// TestDaysWithinASpan: Regular asks for seven different days played within
// ten calendar days, read from the account's own list of days.
func TestDaysWithinASpan(t *testing.T) {
	a := &Account{}
	days := func(list ...string) (float64, bool) {
		a.Days = list
		return a.accountFactLocked("days_within.10")
	}
	if n, ok := days(); !ok || n != 0 {
		t.Errorf("no days: %v %v", n, ok)
	}
	if n, _ := days("2026-10-01", "2026-10-02", "2026-10-03", "2026-10-05", "2026-10-07", "2026-10-09", "2026-10-10"); n != 7 {
		t.Errorf("seven days in ten: %v", n)
	}
	// The eleventh day back is outside the span.
	if n, _ := days("2026-09-30", "2026-10-02", "2026-10-03", "2026-10-05", "2026-10-07", "2026-10-09", "2026-10-10"); n != 6 {
		t.Errorf("six days in ten and one before: %v", n)
	}
	if _, ok := a.accountFactLocked("no_such_fact"); ok {
		t.Error("an unknown account fact reads as known")
	}

	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	acct.mu.Lock()
	acct.Days = []string{"2026-10-01", "2026-10-02", "2026-10-03", "2026-10-05", "2026-10-07", "2026-10-09"}
	acct.mu.Unlock()
	ge.mu.Lock()
	ge.tell(Event{Kind: config.BadgeEvDayPlayed})
	ge.mu.Unlock()
	wantBadges(t, acct, "six days in ten", nil, "special.regular")
	acct.mu.Lock()
	acct.Days = append(acct.Days, "2026-10-10")
	acct.mu.Unlock()
	ge.mu.Lock()
	ge.tell(Event{Kind: config.BadgeEvDayPlayed})
	ge.mu.Unlock()
	wantBadges(t, acct, "seven days in ten", []string{"special.regular"})
}

// TestVisitorInspectedEarnsItsBadge: the map tells the engine the player
// looked at a visitor; the account earns the secret, and the run's facts
// hold nothing of it.
func TestVisitorInspectedEarnsItsBadge(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	ge.NoteVisitorInspected("alien")
	wantBadges(t, acct, "a visitor inspected", []string{"special.we_are_not_alone"})
	if n := runTally(ge, config.BadgeEvVisitor) + runTally(ge, config.BadgeEvVisitor+".alien"); n != 0 {
		t.Errorf("the visitor is tallied in the run's facts: %v", n)
	}
	// With no account it is nothing, and breaks nothing.
	NewGameEngine().NoteVisitorInspected("alien")
}

// TestSetsAndTopRungs: earning a badge of a set counts toward the badge for
// the whole set, and the top rung of a lineage ladder earns Full Set.
func TestSetsAndTopRungs(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	set := rules.Core()
	civs := set.BadgesInSet("civ.met")
	if len(civs) != len(set.Factions()) {
		t.Fatalf("the set civ.met has %d badges for %d civilizations", len(civs), len(set.Factions()))
	}
	for i, f := range set.Factions() {
		say(ge, config.BadgeEvCivMet, f.Key, nil)
		if i < len(civs)-1 && hasBadge(acct, "special.small_world") {
			t.Fatalf("Small World was earned after %d of %d civilizations", i+1, len(civs))
		}
	}
	wantBadges(t, acct, "every civilization met", []string{"special.small_world"})

	top, _ := set.Badge("lineage.harbor.5")
	say(ge, config.BadgeEvBuiltLineage, "harbor", nil)
	wantBadges(t, acct, "one harbor", nil, "special.full_set")
	ge.mu.Lock()
	ge.report(Event{Kind: config.BadgeEvBuiltLineage, Subject: "harbor", N: top.Threshold})
	ge.mu.Unlock()
	wantBadges(t, acct, "the top of a lineage ladder", []string{"lineage.harbor.5", "special.full_set"})
}

// TestTitlesFromBadges: a title comes with its badge (or with a whole
// set), can be worn, and is refused to an account that does not hold it. A
// crossed badge gives no title.
func TestTitlesFromBadges(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	_, sum := ge.Badges()
	if !slices.Equal(sum.Titles, []string{sum.Title}) || sum.Worn != sum.Title {
		t.Fatalf("a new account's titles: %v, worn %q", sum.Titles, sum.Worn)
	}
	if err := acct.SetTitle("Survivor", sum.Titles); err == nil {
		t.Error("a title the account does not hold was worn")
	}
	for i := 0; i < 15; i++ {
		say(ge, config.BadgeEvEndured, "iron_era", nil)
	}
	_, sum = ge.Badges()
	if !slices.Contains(sum.Titles, "Survivor") {
		t.Fatalf("the gold rung of the endured ladder did not give Survivor: %v", sum.Titles)
	}
	if err := acct.SetTitle("survivor", sum.Titles); err != nil {
		t.Fatal(err)
	}
	if _, sum = ge.Badges(); sum.Worn != "Survivor" {
		t.Errorf("worn %q, want Survivor", sum.Worn)
	}
	// Every title in the catalog names a badge or a set that exists.
	for _, tt := range rules.Core().BadgeWornTitles() {
		if tt.Badge != "" {
			def, ok := rules.Core().Badge(tt.Badge)
			if !ok || def.Reward.Title != tt.Title {
				t.Errorf("the title %s names the badge %s, which gives %q", tt.Title, tt.Badge, def.Reward.Title)
			}
		}
		if tt.Set != "" && len(rules.Core().BadgesInSet(tt.Set)) == 0 {
			t.Errorf("the title %s names the set %s, which is empty", tt.Title, tt.Set)
		}
		if (tt.Badge == "") == (tt.Set == "") {
			t.Errorf("the title %s must come with one badge or one set", tt.Title)
		}
	}
	// It is written beside the account, not in it.
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	assertOldShape(t, slotFile(t, acct.AccountID))
	if strings.Contains(string(slotFile(t, acct.AccountID)), "Survivor") {
		t.Error("the title worn was written into account.json")
	}
	// Crossed: the rung is held, the title is not.
	isolateAccountDir(t)
	ge, acct = devEngine(t, "Bea")
	acct.mu.Lock()
	acct.Tampered = true
	acct.mu.Unlock()
	for i := 0; i < 15; i++ {
		say(ge, config.BadgeEvEndured, "iron_era", nil)
	}
	if _, sum = ge.Badges(); !hasBadge(acct, "ladder.endured.3") || slices.Contains(sum.Titles, "Survivor") {
		t.Errorf("a crossed rung gave its title: %v", sum.Titles)
	}
}

// TestAgeJumpTellsEveryAgePassed: a run that is in a later age without
// having advanced into it (the developer console's jump) has the account
// told of every age it passed, on the next tick, so the age badges and
// their themes are earned. A run that has just been loaded is where it is:
// nothing is told again.
func TestAgeJumpTellsEveryAgePassed(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, acct := devEngine(t, "Dev Tester")
	ge.StepTicks(1)
	if out := DevExecCommand("/age classical_age", ge); !strings.Contains(out, "jumped") {
		t.Fatalf("/age: %q", out)
	}
	wantBadges(t, acct, "before the next tick", nil, "age.stone_age")
	ge.StepTicks(1)
	wantBadges(t, acct, "a jump to the Classical Age",
		[]string{"age.stone_age", "age.bronze_age", "age.iron_age", "age.classical_age"}, "age.medieval_age")
	if !acct.HasTheme("bronze") {
		t.Error("the jump passed the Bronze Age and did not unlock its theme")
	}
	for _, key := range acct.EarnedBadges() {
		if e := acct.Badges[key]; e.At == 0 || e.Flags != 0 {
			t.Errorf("%s from a jump is %+v, want dated and uncrossed", key, e)
		}
	}
	// No advance happened in the run: its own facts say so.
	if n := runTally(ge, config.BadgeEvAgeReached); n != 0 {
		t.Errorf("a jump tallied %v advances in the run", n)
	}
	ge.DrainEarnedBadges()

	// Saved and loaded in a fresh engine: the run is where it is.
	if err := ge.SaveGame("run"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	if err := ge2.LoadGame("run"); err != nil {
		t.Fatal(err)
	}
	ge2.StepTicks(3)
	if earned := mainViews(ge2.DrainEarnedBadges()); len(earned) != 0 {
		t.Errorf("loading a run in a later age announced %v", earnedKeysOf(earned))
	}
	// A jump back tells nothing.
	if out := DevExecCommand("/age stone_age", ge2); !strings.Contains(out, "jumped") {
		t.Fatalf("/age: %q", out)
	}
	ge2.StepTicks(2)
	if earned := mainViews(ge2.DrainEarnedBadges()); len(earned) != 0 {
		t.Errorf("a jump back announced %v", earnedKeysOf(earned))
	}
}

func earnedKeysOf(views []BadgeView) []string {
	var out []string
	for _, v := range views {
		out = append(out, v.Key)
	}
	return out
}

// TestHeldBadgesCountTowardTheCollectors: an account from before the
// badges for earning badges existed has no counter for them. Bringing it up
// counts what it holds, grants the rungs that reaches, and a second pass
// changes nothing.
func TestHeldBadgesCountTowardTheCollectors(t *testing.T) {
	isolateAccountDir(t)
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	set := rules.Core()
	held := 0
	acct.mu.Lock()
	acct.Badges = map[string]BadgeEarned{}
	for _, def := range set.Badges() {
		// Thirty badges that are not themselves about earning badges, and
		// one integrity badge, which never counts.
		if def.Integrity() && held < 31 && len(acct.Badges) == 0 {
			acct.Badges[def.Key] = BadgeEarned{At: 5, Run: "old"}
			continue
		}
		if held < 30 && !def.Integrity() && !strings.HasPrefix(def.Counter, config.BadgeEvBadge) && def.Family == "lineage" {
			acct.Badges[def.Key] = BadgeEarned{At: 5, Run: "old"}
			held++
		}
	}
	acct.mu.Unlock()
	ge := NewGameEngine()
	ge.SetAccount(acct)
	collector, _ := set.Badge("special.collector")
	if held < int(collector.Threshold) {
		t.Fatalf("precondition: %d held, Collector asks %v", held, collector.Threshold)
	}
	wantBadges(t, acct, "thirty badges held from before", []string{"special.collector"}, "special.curator")
	if e := acct.Badges["special.collector"]; e.At != 0 || e.Run != "" {
		t.Errorf("Collector, granted from the record, is dated: %+v", e)
	}
	// Collector is itself a badge held: the count includes it.
	if got := counter(acct, config.BadgeEvBadge); got != float64(held+1) {
		t.Errorf("the badge counter is %v, want %d (the thirty and Collector)", got, held+1)
	}
	if earned := ge.DrainEarnedBadges(); len(earned) != 0 {
		t.Errorf("bringing the account up announced %v", earnedKeysOf(earned))
	}
	before := counter(acct, config.BadgeEvBadge)
	if acct.ensureBadges(ge.badges) {
		t.Error("a second pass changed the account")
	}
	if counter(acct, config.BadgeEvBadge) != before {
		t.Error("a second pass moved the counter")
	}
}

// TestPrestigeReportsWhatItCarries: a prestige with Cosmic Legacy held
// counts toward Heirloom Universe, and ruins carried into the next run are
// reported.
func TestPrestigeReportsWhatItCarries(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	for i := 0; i < 4; i++ {
		say(ge, config.BadgeEvLegacyPrestige, "modern_age", nil)
	}
	wantBadges(t, acct, "four prestiges with the legacy", nil, "special.heirloom_universe")
	say(ge, config.BadgeEvLegacyPrestige, "modern_age", nil)
	wantBadges(t, acct, "five prestiges with the legacy", []string{"special.heirloom_universe"})
	say(ge, config.BadgeEvRuinsCarried, "", nil)
	wantBadges(t, acct, "ruins carried into a run", []string{"special.old_bones"})
}

// TestEveryEventAttributeIsSet: a badge that reads an attribute of its
// event ("ev.brace") needs the engine to set that attribute somewhere, or
// an "at least" on it could never hold and an "at most" never be vouched
// for.
func TestEveryEventAttributeIsSet(t *testing.T) {
	src := engineSource(t)
	seen := map[string]bool{}
	check := func(def config.BadgeDef, fact string) {
		attr, ok := strings.CutPrefix(fact, factEvent)
		if !ok {
			return
		}
		if d, staffed := strings.CutPrefix(attr, "staffed."); staffed {
			if !slices.Contains(rules.Core().WorkerDomains(), d) {
				t.Errorf("%s reads the payroll of %q, which is not a domain", def.Key, d)
			}
			attr = "staffed."
		}
		if seen[attr] {
			return
		}
		seen[attr] = true
		if !strings.Contains(src, `"`+attr) {
			t.Errorf("%s reads ev.%s, which the engine never sets", def.Key, attr)
		}
	}
	for _, def := range rules.Core().Badges() {
		check(def, def.Counter)
		for _, c := range def.When {
			check(def, c.Fact)
		}
	}
	if len(seen) < 15 {
		t.Errorf("only %d event attributes are read by the catalog: %v", len(seen), sortedKeys(seen))
	}
}

// engineSource is the engine's own source (test files left out), as one
// string.
func engineSource(t *testing.T) string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		sb.Write(src)
	}
	return sb.String()
}

// TestMaximalistCountsAnAgesBuildings: an age's maximalist badge is judged
// as each building finishes, on how many of that age's buildings stand,
// while the run is still in the age. The event is about the building and
// the badge about the age (AnySubject).
func TestMaximalistCountsAnAgesBuildings(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	def, ok := rules.Core().Badge("maximalist.primitive_age")
	if !ok || !def.AnySubject || def.InAge != "primitive_age" {
		t.Fatalf("the Primitive Age's maximalist row: %+v", def)
	}
	ge.mu.Lock()
	ge.Buildings.counts["hut"] = int(def.Threshold) - 2
	ge.mu.Unlock()
	finishOne(ge, "gathering_camp")
	wantBadges(t, acct, "one building short", nil, def.Key)
	// A wonder does not count.
	finishOne(ge, rules.Core().Wonder("primitive_age"))
	wantBadges(t, acct, "one short and a wonder", nil, def.Key)
	finishOne(ge, "stash")
	wantBadges(t, acct, "the count reached in the age", []string{def.Key}, "maximalist.stone_age")

	// The same count after leaving the age earns nothing.
	isolateAccountDir(t)
	ge, acct = devEngine(t, "Bea")
	ge.mu.Lock()
	ge.Buildings.counts["hut"] = int(def.Threshold)
	ge.mu.Unlock()
	advanceTo(ge, "stone_age")
	finishOne(ge, "gathering_camp")
	wantBadges(t, acct, "the count reached an age late", nil, def.Key)
}

// TestSyllabusCountsAnAgesTechs: an age's syllabus badge is judged as each
// tech finishes, on how many of that age's techs the run has researched.
func TestSyllabusCountsAnAgesTechs(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	set := rules.Core()
	def, ok := set.Badge("techs.primitive_age")
	techs := set.TechsOf("primitive_age")
	if !ok || !def.AnySubject || int(def.Threshold) != len(techs) || len(techs) == 0 {
		t.Fatalf("the Primitive Age's syllabus row: %+v, for %d techs", def, len(techs))
	}
	for i, tech := range techs {
		ge.mu.Lock()
		ge.Research.researched[tech.Key] = true
		ge.note(config.BadgeEvResearchDone, tech.Key)
		ge.mu.Unlock()
		if i < len(techs)-1 {
			wantBadges(t, acct, "part of the age's techs", nil, def.Key)
		}
	}
	wantBadges(t, acct, "every tech of the age", []string{def.Key}, "techs.stone_age")
}
