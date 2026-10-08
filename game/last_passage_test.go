package game

import (
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/rules"
)

// --- helpers ------------------------------------------------------------------

// lpEngine returns a seeded engine standing in age (a Cosmic Era age) with the
// Last Passage thread started, as the first tick of the epoch would.
func lpEngine(t *testing.T, age string, seed int64) *GameEngine {
	t.Helper()
	ge := catEngine(t, age, seed)
	quietFate(ge) // the Cosmic Era's own doom has its tests in fate_test.go
	ge.harbingerTickCheck()
	if ge.harbinger == nil || ge.harbinger.TargetEpoch != "" {
		t.Fatalf("no Last Passage thread at %s: %+v", age, ge.harbinger)
	}
	return ge
}

// makeLastPassagePending prestiges ge with a roll that always hits.
func makeLastPassagePending(t *testing.T, ge *GameEngine) {
	t.Helper()
	ge.rng = riggedRNG(0.0, 0.5, 0.25, 0.75)
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	if !ge.pendingLastPassage {
		t.Fatalf("the Last Passage did not come on a roll of 0")
	}
}

// endureLastPassageIfPending lets a test that prestiges from the final epoch
// get past a Last Passage the seed happened to roll.
func endureLastPassageIfPending(t *testing.T, ge *GameEngine) {
	t.Helper()
	if !ge.GetState().LastPassage.Pending {
		return
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatalf("EndureLastPassage: %v", err)
	}
}

// hasCosmicLegacyModifier reports whether the resolver carries the Cosmic
// Legacy: a x1.1 multiplier beside the all-production pool, never in it.
func hasCosmicLegacyModifier(ge *GameEngine) bool {
	for _, m := range ge.buildResolver().All() {
		if m.Source == cosmicLegacySource && m.Target == "production_all" && m.Op == OpMul && m.Value == 1+CosmicLegacyProductionBonus {
			return true
		}
	}
	return false
}

// indexOfLog returns the index of the first log line containing substr, or -1.
func indexOfLog(ge *GameEngine, substr string) int {
	for i, e := range ge.log {
		if strings.Contains(e.Message, substr) {
			return i
		}
	}
	return -1
}

// --- the cosmic thread ----------------------------------------------------------

func TestCosmicThreadStartsAndHandsOff(t *testing.T) {
	ge := threadEngine(t, "cosmic_era", 21)
	h := ge.harbinger
	if h.EpochKey != "cosmic_era" || h.TargetEpoch != "" || h.Age != "interstellar_age" || h.FalseProphet {
		t.Fatalf("cosmic thread = %+v", h)
	}
	if countLogs(ge, "warning of the Last Passage") != 1 {
		t.Errorf("arrival log should name the Last Passage:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if v := ge.GetState().Harbinger; v == nil || !v.LastPassage || v.TargetEpochName != "the Last Passage" || v.TargetEpochKey != "" {
		t.Errorf("view = %+v", v)
	}

	// Answers carry across the handoffs.
	setStock(ge, map[string][2]float64{"faith": {5e10, 5e10}, "culture": {5e10, 5e10}})
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatalf("appease: %v", err)
	}
	fillBraceStock(ge, 5e16)
	if err := ge.HarbingerBrace(); err != nil {
		t.Fatalf("brace: %v", err)
	}
	if countLogs(ge, "you keep 70% of the run's prestige points") != 1 {
		t.Errorf("brace log should talk about points:\n%s", strings.Join(logMessages(ge), "\n"))
	}

	wantNames := map[string]string{
		"interstellar_age": "the Distress Beacon",
		"galactic_age":     "the Elder Relay",
		"quantum_age":      "your future self",
		"transcendent_age": "your unmade self",
	}
	chain := []string{"interstellar_age"}
	for _, age := range []string{"galactic_age", "quantum_age", "transcendent_age"} {
		walkTo(t, ge, age)
		chain = append(chain, age)
		h := ge.harbinger
		if h == nil || h.Age != age || h.TargetEpoch != "" || strings.Join(h.Chain, ",") != strings.Join(chain, ",") {
			t.Fatalf("at %s: thread = %+v", age, h)
		}
		if h.AppeaseLevel != 1 || h.BraceLevel != 1 {
			t.Errorf("at %s: levels appease %d brace %d, want 1 and 1", age, h.AppeaseLevel, h.BraceLevel)
		}
		def, _ := config.HarbingerFor(age)
		if def.Name != wantNames[age] {
			t.Errorf("at %s: speaker %q, want %q", age, def.Name, wantNames[age])
		}
		if v := ge.GetState().Harbinger; v == nil || v.Name != wantNames[age] {
			t.Errorf("at %s: view %+v", age, v)
		}
	}
	if n := countLogs(ge, "takes up the warning of the Last Passage"); n != 3 {
		t.Errorf("handoff logs = %d, want 3", n)
	}
	if def, _ := config.HarbingerFor("interstellar_age"); def.Name != wantNames["interstellar_age"] {
		t.Errorf("interstellar speaker = %q", def.Name)
	}
}

// A Last Passage thread's Appease and Brace are both priced on the age it
// begins in (the era's first in play; a later one for a thread that began
// there).
func TestCosmicThreadCostsByArrivalAge(t *testing.T) {
	prices := map[string][4]float64{ // faith, culture; dark matter, titanium
		"interstellar_age": {1300000000, 20000000000, 1.1e+16, 9.6e+15},
		"galactic_age":     {2600000000, 42000000000, 1.3e+17, 1.1e+16},
		"quantum_age":      {5400000000, 89000000000, 1.3e+17, 1.1e+16},
		"transcendent_age": {5900000000, 97000000000, 1.4e+17, 1.2e+16},
	}
	for _, age := range epochAges(t, "cosmic_era") {
		ge := lpEngine(t, age, 5)
		v := ge.GetState().Harbinger
		want, ok := prices[age]
		if !ok || v.AppeaseCost["faith"] != want[0] || v.AppeaseCost["culture"] != want[1] || len(v.AppeaseCost) != 2 {
			t.Errorf("%s: appease = %v, want %v faith and %v culture", age, v.AppeaseCost, want[0], want[1])
		}
		if !ok || v.BraceCost["dark_matter"] != want[2] || v.BraceCost["titanium"] != want[3] || len(v.BraceCost) != 2 {
			t.Errorf("%s: brace = %v, want %v dark matter and %v titanium", age, v.BraceCost, want[2], want[3])
		}
		if v.EndurePointsPct != 50 || v.NextEndurePointsPct != 70 {
			t.Errorf("%s: endure points %d%% next %d%%", age, v.EndurePointsPct, v.NextEndurePointsPct)
		}
	}
}

// --- the choice ---------------------------------------------------------------------

// lastPassageLoss is the share of the run's points a player who would Endure
// expects to lose to the Last Passage with these levels bought on its thread
// and faith storage at fill: the engine's chance for the roll, times what an
// Endure gives up.
func lastPassageLoss(t *testing.T, fill float64, appease, brace int) float64 {
	t.Helper()
	ge := lpEngine(t, "interstellar_age", 7)
	setFaithStrength(ge, fill)
	ge.harbinger.AppeaseLevel, ge.harbinger.BraceLevel = appease, brace
	return ge.CatastropheOutlook().Probability * (1 - LastPassageKeepFor(ge.lastPassageBraceLevel()))
}

// incomeHours is how long a moderate economy in age takes to make cost, in
// hours at 1x: the slowest of its resources at config.TypicalIncome (which is
// config.FlowIncome for faith and culture).
func incomeHours(t *testing.T, cost map[string]float64, age string) float64 {
	t.Helper()
	if len(cost) == 0 {
		t.Fatalf("no price to time in %s", age)
	}
	slowest := 0.0
	for res, c := range cost {
		rate := config.TypicalIncome(res, age)
		if rate <= 0 {
			t.Fatalf("nothing makes %s in %s", res, age)
		}
		slowest = math.Max(slowest, c/rate)
	}
	return slowest * config.TickSeconds / 3600
}

// lastPassageChoiceMinRatio is the least Appease level 1 may save per hour of
// a moderate economy's income, as a share of what Brace level 1 saves per
// hour of it. Under a half, Appease is the answer nobody buys.
const lastPassageChoiceMinRatio = 0.5

// The Last Passage's two answers are a choice, not a formality. For a player
// who would Endure, with nothing bought, each level 1 saves the same expected
// points, so the choice is in the prices: Appease level 1 is the dearer, but
// saves at least half as many expected points per hour of income as Brace
// level 1 (about two thirds, at these prices). Change a price or an effect
// and this is the test to argue with. The prices before it (Brace at 12% of
// the era's largest requirement, Appease on the whole age) gave a ratio of 1
// to 13,500.
func TestLastPassageChoice(t *testing.T) {
	const low, mid, high = 0.1, 0.5, 0.9
	// What is left at risk at mid faith, in percent of the run's points, by
	// [Appease level][Brace level]: the table the wiki prints.
	want := [3][3]float64{{7.5, 4.5, 2.25}, {4.5, 2.7, 1.35}, {2.7, 1.62, 0.81}}
	var loss [3][3]float64
	for a := range want {
		for b := range want[a] {
			loss[a][b] = lastPassageLoss(t, mid, a, b)
			if math.Abs(loss[a][b]*100-want[a][b]) > 1e-9 {
				t.Errorf("Appease %d, Brace %d at mid faith: %.4f%% of the run's points at risk, want %v%%", a, b, loss[a][b]*100, want[a][b])
			}
		}
	}
	appeaseSaves, braceSaves := loss[0][0]-loss[1][0], loss[0][0]-loss[0][1]
	if appeaseSaves <= 0 || braceSaves <= 0 {
		t.Fatalf("level 1 saves nothing: Appease %v, Brace %v", appeaseSaves, braceSaves)
	}
	// Appease level 1's saving per hour of income over Brace level 1's.
	valueRatio := func(appeaseHours, braceHours float64) float64 {
		return (appeaseSaves / appeaseHours) / (braceSaves / braceHours)
	}

	ages := epochAges(t, "cosmic_era")
	for _, age := range ages {
		appeaseHours := incomeHours(t, lastPassageAppeaseCost("cosmic_era", age, 1), age)
		braceHours := incomeHours(t, lastPassageBraceCost("cosmic_era", age, 1), age)
		ratio := valueRatio(appeaseHours, braceHours)
		t.Logf("%s: Brace 1 is %.1f h of income, Appease 1 %.1f h; Appease 1 saves %.3fx what Brace 1 does per hour", age, braceHours, appeaseHours, ratio)
		if appeaseHours <= braceHours {
			t.Errorf("%s: Appease level 1 takes %.1f h of income, Brace level 1 %.1f h: Appease must stay the dearer", age, appeaseHours, braceHours)
		}
		if ratio < lastPassageChoiceMinRatio {
			t.Errorf("%s: Appease level 1 saves %.2fx what Brace level 1 does per hour of income (%.1f h against %.1f h), want at least %vx", age, ratio, appeaseHours, braceHours, lastPassageChoiceMinRatio)
		}
	}

	// The thread a run meets, foretold in the era's first age: the hours the
	// wiki quotes.
	first := ages[0]
	appeaseHours := incomeHours(t, lastPassageAppeaseCost("cosmic_era", first, 1), first)
	braceHours := incomeHours(t, lastPassageBraceCost("cosmic_era", first, 1), first)
	if math.Abs(braceHours-22.5) > 0.05 || math.Abs(appeaseHours-32.9) > 0.05 {
		t.Errorf("foretold in %s: Brace level 1 is %.2f h of income and Appease level 1 %.2f h, want 22.5 and 32.9", first, braceHours, appeaseHours)
	}
	if ratio := valueRatio(appeaseHours, braceHours); math.Abs(ratio-0.686) > 0.005 {
		t.Errorf("foretold in %s: the value ratio is %.3f, want 0.686", first, ratio)
	}

	// The faith band moves the odds, not the choice: both savings scale with
	// the chance of the roll.
	for _, fill := range []float64{low, high} {
		none := lastPassageLoss(t, fill, 0, 0)
		a, b := none-lastPassageLoss(t, fill, 1, 0), none-lastPassageLoss(t, fill, 0, 1)
		if math.Abs(a/b-appeaseSaves/braceSaves) > 1e-9 {
			t.Errorf("faith fill %v: Appease level 1 saves %vx what Brace level 1 does, at mid faith %vx", fill, a/b, appeaseSaves/braceSaves)
		}
	}

	// The prices this replaced fail it: Brace at the era's price, which the
	// Reality Tear's thread keeps, and Appease on the whole age.
	oldBrace := incomeHours(t, eraBraceCost("cosmic_era", 1), first)
	oldAppease := incomeHours(t, warningAppeaseCostIn(rules.Core(), "cosmic_era", first, config.AgeTargetTicks(first), 1), first)
	oldRatio := valueRatio(oldAppease, oldBrace)
	t.Logf("the old prices: Brace 1 %.4f h, Appease 1 %.1f h; ratio 1 to %.0f", oldBrace, oldAppease, 1/oldRatio)
	if oldRatio >= lastPassageChoiceMinRatio/1000 {
		t.Errorf("the old prices (Brace %.4f h, Appease %.1f h) give a ratio of %v: the test would not have caught them", oldBrace, oldAppease, oldRatio)
	}
}

// secondLevelMinRatio is the least level 2 of an answer may save per hour of
// a moderate economy's income, as a share of what its level 1 saves per hour
// when bought first. Under a half, the second level is the purchase nobody
// makes.
const secondLevelMinRatio = 0.5

// A second level is worth buying, in every thread: level 2 of Brace and of
// Appease, on the Last Passage's thread, on the Reality Tear's and on an
// ordinary doom's, each save at least half as much per hour of income as
// that answer's level 1 does when bought first. They cost the same again as
// level 1. At double, which level 2 cost outside the final era until the
// ordinary Brace rule, three of the five fall short.
func TestSecondLevelsWorthBuying(t *testing.T) {
	const mid = 0.5
	// What each level saves. The Last Passage: the expected share of the
	// run's points an Endure loses. A doom: Appease moves the chance it
	// strikes; Brace what an Endure takes, in buildings and in stock, with no
	// garrison (a garrison's share comes off both levels alike until the cap).
	loss := func(appease, brace int) float64 { return lastPassageLoss(t, mid, appease, brace) }
	strike := func(appease int) float64 { return StrikeChanceAt(mid, appease) }
	ge := catEngine(t, "interstellar_age", 1)
	endure := func(brace int) EndureOutcome { return ge.endurePreview(brace, ge.age) }
	if g := endure(0).Garrison; g != 0 {
		t.Fatalf("the test engine has a garrison (%v): the Brace figures below assume none", g)
	}
	type answer struct {
		name  string
		saves [2]float64 // level 1 bought first, then level 2
	}
	// What the effects alone give, level 2 over level 1: the ratio at the
	// same price. At double the price it halves.
	lpBrace := answer{"the Last Passage's Brace", [2]float64{loss(0, 0) - loss(0, 1), loss(0, 1) - loss(0, 2)}}
	lpAppease := answer{"the Last Passage's Appease", [2]float64{loss(0, 0) - loss(1, 0), loss(1, 0) - loss(2, 0)}}
	doomBuildings := answer{"a doom's Brace, buildings", [2]float64{endure(0).DestroyPct - endure(1).DestroyPct, endure(1).DestroyPct - endure(2).DestroyPct}}
	doomStock := answer{"a doom's Brace, stock", [2]float64{endure(1).KeepFrac - endure(0).KeepFrac, endure(2).KeepFrac - endure(1).KeepFrac}}
	doomAppease := answer{"a doom's Appease", [2]float64{strike(0) - strike(1), strike(1) - strike(2)}}
	answers := []answer{lpBrace, lpAppease, doomBuildings, doomStock, doomAppease}
	wantEffect := []float64{0.75, 0.6, 1, 1, 0.6}
	for i, a := range answers {
		if a.saves[0] <= 0 || a.saves[1] <= 0 {
			t.Fatalf("%s: level 1 saves %v, level 2 %v", a.name, a.saves[0], a.saves[1])
		}
		if got := a.saves[1] / a.saves[0]; math.Abs(got-wantEffect[i]) > 1e-9 {
			t.Errorf("%s: level 2 saves %.3f of what level 1 does, want %v", a.name, got, wantEffect[i])
		}
	}
	// Every thread a run can meet, at the prices of every age its harbinger
	// can arrive in.
	type priced struct {
		answer
		cost func(age string, level int) map[string]float64
	}
	checked := 0
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) {
			continue
		}
		epoch := ep.Key
		threads := []priced{
			{doomBuildings, func(age string, level int) map[string]float64 { return doomBraceCost(epoch, age, level) }},
			{doomStock, func(age string, level int) map[string]float64 { return doomBraceCost(epoch, age, level) }},
			{doomAppease, func(age string, level int) map[string]float64 { return doomAppeaseCost(epoch, age, level) }},
		}
		if config.IsFinalEpoch(epoch) {
			threads = append(threads,
				priced{lpBrace, func(age string, level int) map[string]float64 { return lastPassageBraceCost(epoch, age, level) }},
				priced{lpAppease, func(age string, level int) map[string]float64 { return lastPassageAppeaseCost(epoch, age, level) }})
		}
		for _, age := range ep.Ages {
			for _, a := range threads {
				checked++
				h1, h2 := incomeHours(t, a.cost(age, 1), age), incomeHours(t, a.cost(age, 2), age)
				ratio := (a.saves[1] / h2) / (a.saves[0] / h1)
				if ratio < secondLevelMinRatio {
					t.Errorf("%s, foretold in %s: level 2 (%.1f h of income) saves %.2fx what level 1 (%.1f h) does per hour, want at least %vx", a.name, age, h2, ratio, h1, secondLevelMinRatio)
				}
				if h2 != h1 {
					t.Errorf("%s, foretold in %s: level 2 takes %.2f h of income and level 1 %.2f h; they cost the same", a.name, age, h2, h1)
				}
				if age == "interstellar_age" || age == "victorian_age" {
					t.Logf("%s, foretold in %s: level 1 %.1f h, level 2 %.1f h more; level 2 saves %.2fx what level 1 does per hour (%.2fx at double the price)", a.name, age, h1, h2, ratio, ratio/2)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no thread checked")
	}
	// At double the price the bar catches the Last Passage's Brace and both
	// kinds of Appease; a doom's Brace sits exactly on it.
	short := 0
	for i := range answers {
		if wantEffect[i]/2 < secondLevelMinRatio {
			short++
		}
	}
	if short != 3 {
		t.Errorf("at double the price %d of the %d second levels fall short of %v, want 3: the test would not have caught the old prices", short, len(answers), secondLevelMinRatio)
	}
}

// doomChoiceMinRatio is the least an ordinary doom's Appease level 1 may
// save per hour of a moderate economy's income, as a share of what its
// Brace level 1 saves per hour: the bar the Last Passage's choice is held
// to (lastPassageChoiceMinRatio), for every other thread.
const doomChoiceMinRatio = 0.5

// An ordinary doom's two answers are a choice in every age its harbinger
// can arrive in. Appease level 1 takes two fifths off the chance the doom
// strikes, so it saves two fifths of whatever an Endure would cost. Brace
// level 1 softens the Endure itself: a quarter of the buildings that would
// fall stand, and 15 points more of every stock are kept (0.176 of what
// would be lost). Appease is the dearer (three quarters of the warning
// against a third), and still saves at least half as much per hour of
// income as Brace does on either count. Brace used to cost seconds of
// income from an era's second age on, which made the ratio thousands to
// one.
//
// Brace has a ceiling too: no material asks for more than three fifths of
// a moderate builder's store. In most ages one material stays under it and
// still takes a third of the warning, so the ratio stands. In two ages
// every material is on the ceiling, Brace is made in about a sixth of the
// warning, and Appease falls under the bar: doomChoiceUnderTheCeiling
// names them with what they read. The ceiling is the rule that wins there
// (a Brace nobody's store holds is no answer at all), and the list is
// where that is said out loud.
func TestOrdinaryDoomChoice(t *testing.T) {
	const mid = 0.5
	ge := catEngine(t, "iron_age", 1)
	e0, e1 := ge.endurePreview(0, ge.age), ge.endurePreview(1, ge.age)
	if e0.Garrison != 0 {
		t.Fatalf("the test engine has a garrison (%v)", e0.Garrison)
	}
	// As shares of what an unanswered Endure costs.
	appeaseSaves := 1 - StrikeChanceAt(mid, 1)/StrikeChanceAt(mid, 0)
	braceBuildings := (e0.DestroyPct - e1.DestroyPct) / e0.DestroyPct
	braceStock := (e1.KeepFrac - e0.KeepFrac) / (1 - e0.KeepFrac)
	if math.Abs(appeaseSaves-0.4) > 1e-9 || math.Abs(braceBuildings-0.25) > 1e-9 || math.Abs(braceStock-0.15/0.85) > 1e-9 {
		t.Fatalf("level 1 saves: Appease %v of the loss, Brace %v of the buildings and %v of the stock; want 0.4, 0.25 and 0.176", appeaseSaves, braceBuildings, braceStock)
	}
	braceSaves := math.Max(braceBuildings, braceStock)
	checked := 0
	seen := map[string]bool{}
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) || config.IsFinalEpoch(ep.Key) {
			continue
		}
		for _, age := range ep.Ages {
			checked++
			appeaseHours := incomeHours(t, doomAppeaseCost(ep.Key, age, 1), age)
			braceHours := incomeHours(t, doomBraceCost(ep.Key, age, 1), age)
			ratio := (appeaseSaves / appeaseHours) / (braceSaves / braceHours)
			if age == ep.Ages[0] {
				t.Logf("%s: Brace 1 is %.2f h of income, Appease 1 %.2f h; Appease 1 saves %.2fx what Brace 1 does per hour", age, braceHours, appeaseHours, ratio)
			}
			if appeaseHours <= braceHours {
				t.Errorf("%s: Appease level 1 takes %.2f h of income, Brace level 1 %.2f h: Appease must stay the dearer", age, appeaseHours, braceHours)
			}
			// Every material on its ceiling?
			capped := true
			for res, c := range doomBraceCost(ep.Key, age, 1) {
				if c < 0.9*ordinaryDoomBraceStoreShare*rules.Core().TypicalStorage(res, age) {
					capped = false
				}
			}
			warning := harbingerLeadMin * config.AgeTargetTicks(age) * config.TickSeconds / 3600
			share := braceHours / warning
			if want, under := doomChoiceUnderTheCeiling[age]; under {
				seen[age] = true
				if !capped || ratio >= doomChoiceMinRatio || math.Abs(ratio-want) > 0.02 {
					t.Errorf("%s is listed as under the bar at %.2fx with every material on its ceiling; it reads %.3fx (every material capped: %v)", age, want, ratio, capped)
				}
				if share >= ordinaryDoomBraceShare {
					t.Errorf("%s: Brace level 1 takes %.2f of the shortest warning, want under a third on the ceiling", age, share)
				}
				t.Logf("%s, every material on its ceiling: Brace 1 is %.2f h of income (%.2f of the warning), Appease 1 %.2f h; Appease 1 saves %.2fx what Brace 1 does per hour, under the %vx bar", age, braceHours, share, appeaseHours, ratio, doomChoiceMinRatio)
				continue
			}
			if ratio < doomChoiceMinRatio {
				t.Errorf("%s: Appease level 1 saves %.2fx what Brace level 1 does per hour of income (%.2f h against %.2f h), want at least %vx", age, ratio, appeaseHours, braceHours, doomChoiceMinRatio)
			}
			// The warning pays for Brace: a third of it, give or take the
			// rounding up to two figures.
			if capped || share < ordinaryDoomBraceShare || share > 1.1*ordinaryDoomBraceShare {
				t.Errorf("%s: Brace level 1 takes %.2f of the shortest warning (%.2f h of %.2f; every material capped: %v), want a third", age, share, braceHours, warning, capped)
			}
			// The old price fails the bar from the era's second age on.
			if age != ep.Ages[0] {
				oldHours := incomeHoursOfMade(eraBraceCost(ep.Key, 1), age)
				if old := (appeaseSaves / appeaseHours) / (braceSaves / oldHours); old >= doomChoiceMinRatio {
					t.Errorf("%s: the era's old price (%.4f h of income) gives a ratio of %.3f: the test would not have caught it", age, oldHours, old)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no age checked")
	}
	if len(seen) != len(doomChoiceUnderTheCeiling) {
		t.Errorf("ages listed under the bar: %v; met: %v", doomChoiceUnderTheCeiling, seen)
	}
}

// doomChoiceUnderTheCeiling is the ages where every material of an ordinary
// doom's Brace sits on its ceiling (three fifths of a moderate builder's
// store), with what Appease level 1 then saves per hour of income against
// Brace level 1: under doomChoiceMinRatio, because Brace is made in about a
// sixth of the warning there. Both ages build with two materials that
// their vaults hold little of beside what they make.
var doomChoiceUnderTheCeiling = map[string]float64{
	"industrial_age": 0.33,
	"modern_age":     0.39,
}

// incomeHoursOfMade is incomeHours over the resources of cost the age makes
// at all: the old era price named some an age only buys.
func incomeHoursOfMade(cost map[string]float64, age string) float64 {
	slowest := 0.0
	for res, c := range cost {
		if rules.Core().BuildingOutput(res, age) <= 0 {
			continue
		}
		slowest = math.Max(slowest, c/config.TypicalIncome(res, age))
	}
	return slowest * config.TickSeconds / 3600
}

// --- the outlook -------------------------------------------------------------------

func TestOutlookPassageKind(t *testing.T) {
	for _, ep := range config.Epochs() {
		ge := catEngine(t, ep.Ages[0], 1)
		o := ge.CatastropheOutlook()
		want := PassageEpoch
		if config.IsFinalEpoch(ep.Key) {
			want = PassagePrestige
		}
		if o.Passage != want {
			t.Errorf("%s: passage %q, want %q", ep.Key, o.Passage, want)
		}
		if want == PassagePrestige && (o.NextEpochKey != "" || !o.Possible) {
			t.Errorf("%s: prestige outlook = %+v", ep.Key, o)
		}
		if want == PassageEpoch && o.NextEpochKey == "" {
			t.Errorf("%s: epoch outlook has no next epoch: %+v", ep.Key, o)
		}
	}
	// Invited: certain. Appeased twice: ×0.36.
	ge := lpEngine(t, "galactic_age", 1)
	ge.harbinger.AppeaseLevel = 2
	if p := ge.CatastropheOutlook().Probability; p < 0.0647 || p > 0.0649 {
		t.Errorf("appeased twice: %v, want 0.18 × 0.36", p)
	}
	ge.catastropheInvited = true
	if p := ge.CatastropheOutlook().Probability; p != 1 {
		t.Errorf("invited: %v, want 1", p)
	}
}

// --- the roll ------------------------------------------------------------------------

func TestLastPassageOnlyRollsFromTheCosmicEra(t *testing.T) {
	// A stray invite before the Cosmic Era is not a Last Passage.
	ge := catEngine(t, "space_age", 3)
	ge.catastropheInvited = true
	ge.rng = riggedRNG(0.0)
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if ge.pendingLastPassage || ge.Prestige.GetLevel() != 1 || ge.age != "primitive_age" {
		t.Errorf("space-age prestige: pending %v level %d age %s", ge.pendingLastPassage, ge.Prestige.GetLevel(), ge.age)
	}

	ge = lpEngine(t, "interstellar_age", 3)
	makeLastPassagePending(t, ge)
	if ge.Prestige.GetLevel() != 0 || ge.age != "interstellar_age" {
		t.Errorf("pending Last Passage completed the prestige: level %d age %s", ge.Prestige.GetLevel(), ge.age)
	}
	st := ge.GetState()
	if !st.LastPassage.Pending || st.CatastropheOutlook.Possible || st.PendingCatastrophe != "" {
		t.Errorf("pending state = %+v, outlook %+v", st.LastPassage, st.CatastropheOutlook)
	}
	if countLogs(ge, "The Last Passage has come") != 1 {
		t.Errorf("no arrival log:\n%s", strings.Join(logMessages(ge), "\n"))
	}
}

func TestLastPassageSparedCompletesPrestige(t *testing.T) {
	ge := lpEngine(t, "transcendent_age", 4)
	full := ge.GetState().Prestige.PendingPoints
	ge.rng = riggedRNG(0.99, 0.3, 0.6, 0.1)
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if ge.pendingLastPassage || ge.Prestige.GetLevel() != 1 || ge.Prestige.Snapshot().Available != full || ge.age != "primitive_age" {
		t.Fatalf("spared: pending %v level %d points %d (want %d) age %s", ge.pendingLastPassage,
			ge.Prestige.GetLevel(), ge.Prestige.Snapshot().Available, full, ge.age)
	}
	verdict := indexOfLog(ge, "you were spared")
	done := indexOfLog(ge, "Prestige complete.")
	if verdict < 0 || done < 0 || verdict > done {
		t.Fatalf("verdict %d, prestige line %d:\n%s", verdict, done, strings.Join(logMessages(ge), "\n"))
	}
	if !strings.Contains(ge.log[verdict].Message, "The Last Passage opens") {
		t.Errorf("verdict = %q", ge.log[verdict].Message)
	}
	if ge.harbinger != nil || ge.harbingerHistory != nil || ge.catastropheInvited {
		t.Errorf("thread not cleared: %+v", ge.harbinger)
	}
	if len(ge.catastropheHistory) != 0 {
		t.Errorf("a spared passage wrote the civilization log: %v", ge.catastropheHistory)
	}
}

func TestLastPassageEndureKeepsAShareOfThePoints(t *testing.T) {
	for brace, keep := range []float64{0.50, 0.70, 0.85} {
		ge := lpEngine(t, "transcendent_age", 6)
		ge.harbinger.BraceLevel = brace
		full := ge.GetState().Prestige.PendingPoints
		want := int(float64(full) * keep)
		if lp := ge.GetState().LastPassage; lp.KeepPct != int(keep*100+0.5) || lp.PointsIfEndured != want || lp.BraceLevel != brace {
			t.Errorf("brace %d: state %+v, want keep %v and %d points", brace, lp, keep, want)
		}
		makeLastPassagePending(t, ge)
		if err := ge.EndureLastPassage(); err != nil {
			t.Fatalf("brace %d: %v", brace, err)
		}
		if got := ge.Prestige.Snapshot().Available; got != want || ge.Prestige.GetLevel() != 1 || ge.age != "primitive_age" {
			t.Errorf("brace %d: %d points level %d age %s, want %d of %d", brace, got, ge.Prestige.GetLevel(), ge.age, want, full)
		}
		if ge.cosmicLegacy || ge.pendingLastPassage {
			t.Errorf("brace %d: Endure granted the legacy or left it pending", brace)
		}
		if indexOfLog(ge, "was right") < 0 || indexOfLog(ge, "Endure: The Last Passage") < 0 {
			t.Errorf("brace %d: missing verdict or outcome:\n%s", brace, strings.Join(logMessages(ge), "\n"))
		}
		if e, _ := countCatastropheOutcomes(ge.catastropheHistory); e != 1 {
			t.Errorf("brace %d: endured count %d, want 1", brace, e)
		}
	}
	// A run too short to keep anything keeps nothing, never less.
	ge := lpEngine(t, "interstellar_age", 6)
	if got := ge.lastPassagePoints(lastPassageEndured, 1); got != 0 {
		t.Errorf("Endure of 1 point = %d, want 0", got)
	}
}

func TestLastPassageSuccumbGrantsTheCosmicLegacy(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := lpEngine(t, "quantum_age", 7)
	ge.Prestige.LoadState(2, 10, 4, nil) // a veteran: level 2, 4 points in the bank
	if hasCosmicLegacyModifier(ge) {
		t.Fatal("legacy before Succumb")
	}
	makeLastPassagePending(t, ge)
	if err := ge.SuccumbLastPassage(); err != nil {
		t.Fatal(err)
	}
	snap := ge.Prestige.Snapshot()
	if snap.Level != 3 || snap.Available != 4 || snap.TotalEarned != 10 || !ge.cosmicLegacy || ge.age != "primitive_age" {
		t.Fatalf("after Succumb: %+v legacy %v age %s", snap, ge.cosmicLegacy, ge.age)
	}
	if indexOfLog(ge, "Cosmic Legacy: all production +10%") < 0 || indexOfLog(ge, "Succumb: The Last Passage") < 0 {
		t.Errorf("log:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if _, s := countCatastropheOutcomes(ge.catastropheHistory); s != 1 {
		t.Errorf("succumbed count %d, want 1", s)
	}

	check := func(when string) {
		t.Helper()
		if !ge.cosmicLegacy || !hasCosmicLegacyModifier(ge) {
			t.Fatalf("%s: legacy %v, modifier %v", when, ge.cosmicLegacy, hasCosmicLegacyModifier(ge))
		}
		found := false
		for _, m := range ge.GetState().Modifiers {
			if m.Source == cosmicLegacySource && m.Target == "production_all" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: the Cosmic Legacy is missing from GameState.Modifiers", when)
		}
		if _, ok := ge.permanentBonuses["production_all"]; ok {
			t.Errorf("%s: the legacy was stored as a bonus value", when)
		}
	}
	check("after Succumb")
	base := ge.buildResolver().AddTotal("production_all")

	// Two more prestiges, one of them from the Cosmic Era.
	ge.age, ge.currentEpoch = "modern_age", config.EpochForAge("modern_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	check("after a Modern Age prestige")
	ge.age, ge.currentEpoch = "galactic_age", config.EpochForAge("galactic_age")
	quietFate(ge)
	ge.harbingerTickCheck()
	makeLastPassagePending(t, ge)
	// Owned already: Succumb is refused, Endure is the only answer.
	if err := ge.SuccumbLastPassage(); err == nil || !strings.Contains(err.Error(), "already carry the Cosmic Legacy") {
		t.Fatalf("second Succumb: err = %v", err)
	}
	if err := ge.Succumb(); err == nil || !ge.pendingLastPassage {
		t.Fatalf("Succumb() routed past the refusal: err %v pending %v", err, ge.pendingLastPassage)
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	check("after a Cosmic Era prestige")
	// The epoch catastrophe's Succumb resets the run too.
	ge.age, ge.currentEpoch = "iron_age", "iron_era"
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	check("after an epoch Succumb")

	if err := ge.SaveGame("legacy"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("legacy"); err != nil {
		t.Fatal(err)
	}
	ge = ge2
	check("after save/load")
	// It multiplies production after the caps: the pool is as it was, and
	// the factor is back.
	if got := ge.buildResolver().AddTotal("production_all"); got != base {
		t.Errorf("production_all pool after load = %v, want %v: the legacy is not part of the pool", got, base)
	}
	if got := ge.cosmicLegacyFactor(); got != 1+CosmicLegacyProductionBonus {
		t.Errorf("cosmicLegacyFactor after load = %v, want %v", got, 1+CosmicLegacyProductionBonus)
	}
	if ge.cheaterBadge {
		t.Error("save with the Cosmic Legacy failed its signature")
	}

	// Only a full wipe clears it.
	ge.Reset()
	if ge.cosmicLegacy || hasCosmicLegacyModifier(ge) {
		t.Error("Reset kept the Cosmic Legacy")
	}
}

func TestInviteGuaranteesTheLastPassage(t *testing.T) {
	ge := lpEngine(t, "quantum_age", 8)
	ge.harbinger.BraceLevel = 2
	if err := ge.HarbingerInvite(); err != nil {
		t.Fatal(err)
	}
	if countLogs(ge, "Your next prestige will bring the Last Passage") != 1 {
		t.Errorf("invite log:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if err := ge.HarbingerAppease(); err == nil {
		t.Error("Appease after an invite should refuse")
	}
	full := ge.GetState().Prestige.PendingPoints
	ge.rng = riggedRNG(0.999) // would miss any real odds
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if !ge.pendingLastPassage || ge.catastropheInvited {
		t.Fatalf("invited: pending %v invite still armed %v", ge.pendingLastPassage, ge.catastropheInvited)
	}
	if !ge.GetState().LastPassage.Invited {
		t.Error("state does not say the Last Passage was invited")
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	if indexOfLog(ge, "You invited it, and it came") < 0 {
		t.Errorf("no Fulfilled verdict:\n%s", strings.Join(logMessages(ge), "\n"))
	}
	if got, want := ge.Prestige.Snapshot().Available, int(float64(full)*0.85); got != want {
		t.Errorf("braced twice: %d points, want %d of %d", got, want, full)
	}
}

func TestPendingLastPassageBlocksOnlyPrestigeAndSurvivesSaveLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := lpEngine(t, "galactic_age", 9)
	ge.harbinger.BraceLevel = 1
	makeLastPassagePending(t, ge)

	if err := ge.DoPrestige(); err == nil || !strings.Contains(err.Error(), "The Last Passage is upon you") {
		t.Fatalf("second prestige: err = %v", err)
	}
	if err := ge.HarbingerBrace(); err == nil || !strings.Contains(err.Error(), "already come") {
		t.Errorf("brace while pending: err = %v", err)
	}
	if v := ge.GetState().Harbinger; v == nil || !v.PassageCame || v.BraceBlocked == "" || v.AppeaseBlocked == "" || v.InviteBlocked == "" {
		t.Errorf("view while pending = %+v", v)
	}
	if err := ge.forceCatastrophe(); err == nil {
		t.Error("an epoch catastrophe was stacked on the pending Last Passage")
	}

	if err := ge.SaveGame("pending_lp"); err != nil {
		t.Fatal(err)
	}
	infos, err := ListSaveDetails()
	if err != nil || len(infos) != 1 || infos[0].PendingCatastrophe != "The Last Passage" {
		t.Errorf("save details = %+v, err %v", infos, err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("pending_lp"); err != nil {
		t.Fatal(err)
	}
	if ge2.cheaterBadge {
		t.Error("pending save failed its signature")
	}
	st := ge2.GetState()
	if !st.LastPassage.Pending || st.LastPassage.BraceLevel != 1 || st.PendingCatastrophe != "" {
		t.Fatalf("after load: %+v", st.LastPassage)
	}
	if err := ge2.DoPrestige(); err == nil {
		t.Error("prestige after load was not blocked")
	}

	// Advancing is not blocked; the next figure takes up the thread.
	ge2.mu.Lock()
	ge2.ageReady = true
	ge2.mu.Unlock()
	if err := ge2.AdvanceAge(); err != nil {
		t.Fatalf("AdvanceAge while the Last Passage waits: %v", err)
	}
	if ge2.age != "quantum_age" || ge2.harbinger == nil || ge2.harbinger.Age != "quantum_age" || !ge2.pendingLastPassage {
		t.Fatalf("after advancing: age %s thread %+v pending %v", ge2.age, ge2.harbinger, ge2.pendingLastPassage)
	}

	full := ge2.GetState().Prestige.PendingPoints
	if err := ge2.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	if got := ge2.Prestige.Snapshot().Available; got != int(float64(full)*0.70) {
		t.Errorf("Endure after load: %d points, want 70%% of %d", got, full)
	}
}

func TestLastPassageIsDeterministic(t *testing.T) {
	run := func(seed int64) (bool, []string) {
		ge := lpEngine(t, "transcendent_age", seed)
		if err := ge.DoPrestige(); err != nil {
			t.Fatal(err)
		}
		came := ge.pendingLastPassage
		if came {
			if err := ge.EndureLastPassage(); err != nil {
				t.Fatal(err)
			}
		}
		return came, logMessages(ge)
	}
	came, spared := 0, 0
	for seed := int64(1); seed <= 60; seed++ {
		a, la := run(seed)
		b, lb := run(seed)
		if a != b || strings.Join(la, "\n") != strings.Join(lb, "\n") {
			t.Fatalf("seed %d: outcome or log differs between runs", seed)
		}
		if a {
			came++
		} else {
			spared++
		}
	}
	if came == 0 || spared == 0 {
		t.Errorf("60 seeds gave %d Last Passages and %d spared; want both", came, spared)
	}
}

// --- the ending lines -------------------------------------------------------------------

// runEndingCorpus is every line RunEnding can say at age with the Subject the
// engine sends (the pools are slotless, so a few thousand draws see them all).
func runEndingCorpus(age string) map[string]bool {
	req := flavor.Request{Moment: flavor.RunEnding, Age: age}
	if config.IsFinalEpoch(config.EpochForAge(age)) {
		def, _ := config.HarbingerFor(age)
		req.Subject = def.Name
	}
	rng := rand.New(rand.NewSource(1))
	out := map[string]bool{}
	for i := 0; i < 4000; i++ {
		out[flavor.Line(req, rng)] = true
	}
	return out
}

func TestRunEndingLineAtEveryPrestige(t *testing.T) {
	for _, age := range []string{"modern_age", "information_age", "cyberpunk_age", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"} {
		ge := catEngine(t, age, 10)
		quietFate(ge)
		ge.harbingerTickCheck()
		ge.rng = rand.New(&seqSource{vals: []float64{0.99, 0.41, 0.07, 0.63, 0.29, 0.85}})
		if err := ge.DoPrestige(); err != nil {
			t.Fatal(err)
		}
		if ge.pendingLastPassage {
			t.Fatalf("%s: rolled a Last Passage on 0.99", age)
		}
		done := indexOfLog(ge, "Prestige complete.")
		if done < 1 {
			t.Fatalf("%s: no line before the prestige summary:\n%s", age, strings.Join(logMessages(ge), "\n"))
		}
		ending := strings.TrimSuffix(strings.TrimPrefix(ge.log[done-1].Message, "  [gray]"), "[-]")
		corpus := runEndingCorpus(age)
		if !corpus[ending] {
			t.Errorf("%s: the line before the summary is not a RunEnding line for this age: %q", age, ending)
		}
		if ge.log[done-1].Tick != 0 {
			t.Errorf("%s: carried line kept its old tick %d", age, ge.log[done-1].Tick)
		}
		// The plain register before the Cosmic Era: no Cosmic Era line can be
		// drawn there, and no figure is named.
		if !config.IsFinalEpoch(config.EpochForAge(age)) {
			for line := range runEndingCorpus("galactic_age") {
				if corpus[line] && strings.Contains(line, "Elder Relay") {
					t.Errorf("%s reaches a Cosmic Era line: %q", age, line)
				}
			}
		}
	}
}

// The Cosmic Legacy multiplies production after the pools, so it counts in
// full in every age. In the all-production pool it would count a quarter
// once the pool is past +200%. Each age from the Victorian on is read with
// the pool far past that knee, where a pooled +10% is worth +2.5%.
func TestCosmicLegacyCountsAfterTheCap(t *testing.T) {
	keys := ageKeys()
	from := ageOrders()["victorian_age"]
	for _, age := range append([]string{keys[0]}, keys[from:]...) {
		ge := newTruthEngine(age, truthTypical)
		if ageOrders()[age] >= from {
			// Milestone rewards on top of the wonders: the pool is past its knee.
			ge.permanentBonuses["production_all"] += 3
			ge.recalculateRates()
			if p := ge.bonusPoolLocked(ge.buildResolver(), "production_all"); !p.Soft {
				t.Fatalf("%s: the all-production pool is not past its knee (%+v): the test would prove nothing", age, p)
			}
		}
		pool := ge.buildResolver().AddTotal("production_all")
		type reading struct{ rate, gross float64 }
		read := func() map[string]reading {
			ge.recalculateRates()
			out := map[string]reading{}
			k := ge.speedK()
			for _, key := range ge.Resources.order {
				r := ge.Resources.resources[key]
				// What the resource makes before the drain and Era Mastery.
				out[key] = reading{rate: r.Rate, gross: r.Rate/k - r.Breakdown.FoodDrain}
			}
			return out
		}
		off := read()
		ge.cosmicLegacy = true
		on := read()
		if got := ge.buildResolver().AddTotal("production_all"); got != pool {
			t.Errorf("%s: the legacy moved the all-production pool from %v to %v", age, pool, got)
		}
		if note := ge.capNoteLocked(config.Effect{Type: "production_all", Value: CosmicLegacyProductionBonus}, false); ageOrders()[age] >= from && note == "" {
			t.Errorf("%s: a pooled +10%% would count in full here: the test would prove nothing", age)
		}
		moved := 0
		for _, key := range ge.Resources.order {
			if off[key].gross <= 0 {
				if on[key].rate != off[key].rate {
					t.Errorf("%s: %s makes nothing, but its rate moved from %v to %v", age, key, off[key].rate, on[key].rate)
				}
				continue
			}
			moved++
			if got, want := on[key].gross/off[key].gross, 1+CosmicLegacyProductionBonus; math.Abs(got-want) > 1e-9 {
				t.Errorf("%s: %s production x%v with the legacy, want x%v", age, key, got, want)
			}
			if b := ge.Resources.resources[key].Breakdown; math.Abs(b.LegacyRate-CosmicLegacyProductionBonus*off[key].gross) > 1e-9*off[key].gross {
				t.Errorf("%s: %s breakdown says the legacy adds %v, want a tenth of %v", age, key, b.LegacyRate, off[key].gross)
			}
		}
		if moved == 0 {
			t.Errorf("%s: nothing is made here", age)
		}
		// The food drain is not production: the legacy leaves it alone.
		if food := ge.Resources.resources["food"]; food.Breakdown.FoodDrain >= 0 {
			t.Errorf("%s: no food drain to check", age)
		} else if got, want := on["food"].rate-off["food"].rate, CosmicLegacyProductionBonus*off["food"].gross*ge.speedK(); math.Abs(got-want) > 1e-9*math.Abs(want) {
			t.Errorf("%s: food rate rose by %v, want a tenth of what is grown (%v): the drain must not be multiplied", age, got, want)
		}
	}
}

// With food in deficit the legacy still only helps: it multiplies what is
// grown, never the drain.
func TestCosmicLegacyNeverDeepensAFoodDeficit(t *testing.T) {
	ge := newTruthEngine("iron_age", truthClean)
	ge.Workers.domains["worker"].count += 100000 // far more mouths than the farms feed
	ge.recalculateRates()
	before := ge.Resources.resources["food"].Rate
	if before >= 0 {
		t.Fatalf("food rate %v: want a deficit", before)
	}
	ge.cosmicLegacy = true
	ge.recalculateRates()
	if after := ge.Resources.resources["food"].Rate; after <= before {
		t.Errorf("food rate %v with the legacy, %v without: it must rise", after, before)
	}
}
