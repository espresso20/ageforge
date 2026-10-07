package game

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// --- faith strength -----------------------------------------------------------

// faithBuildingsUpTo lists the faith buildings a town can hold in age: every
// producer of faith that is not a wonder, from the Primitive Age to age.
func faithBuildingsUpTo(age string) []string {
	order := ageOrders()
	var keys []string
	for _, d := range config.BaseBuildings() {
		if d.Category == "wonder" {
			continue
		}
		if j, ok := order[d.RequiredAge]; !ok || j > order[age] {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == "faith" && e.Value > 0 {
				keys = append(keys, d.Key)
				break
			}
		}
	}
	return keys
}

// unstaffedPerSet is how many copies of a faith building with nobody in them
// make what config.FlowCopies fully staffed ones do: an empty building makes
// a fifth of a full one.
const unstaffedPerSet = 25

// faithTown is an engine in age with sets moderate sets' worth of faith
// buildings (unstaffed copies, which make the same faith with no workers to
// feed), every faith wonder of an earlier age standing, and a store nothing
// here could fill. sets may be fractional in steps of 1/25.
func faithTown(t *testing.T, age string, sets float64) *GameEngine {
	t.Helper()
	ge := catEngine(t, age, 7)
	order := ageOrders()
	for _, a := range ageKeys() {
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
	}
	setFaithBuildings(ge, age, sets)
	for _, d := range config.BaseBuildings() {
		if d.Category != "wonder" || order[d.RequiredAge] >= order[age] {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == "faith" {
				ge.Buildings.counts[d.Key] = 1
			}
		}
	}
	ge.Buildings.counts["quantum_vault"] = 1 // storage for everything, faith included
	return ge
}

// setFaithBuildings gives ge sets moderate sets' worth of every faith
// building age has.
func setFaithBuildings(ge *GameEngine, age string, sets float64) {
	for _, key := range faithBuildingsUpTo(age) {
		ge.Buildings.counts[key] = int(math.Round(sets * unstaffedPerSet))
	}
}

// The bands: under a quarter of full strength is low, over three quarters
// high.
func TestFaithBandAt(t *testing.T) {
	for _, c := range []struct {
		strength float64
		want     FaithBand
		good     float64
	}{
		{0, FaithBandLow, 0.40},
		{0.2499, FaithBandLow, 0.40},
		{0.25, FaithBandMid, 0.50},
		{0.75, FaithBandMid, 0.50},
		{0.7501, FaithBandHigh, 0.60},
		{1, FaithBandHigh, 0.60},
	} {
		if got := FaithBandAt(c.strength); got != c.want {
			t.Errorf("strength %v: band %s, want %s", c.strength, got, c.want)
		}
		if got := goodChanceFor(c.strength); got != c.good || EpochGoodChanceIn(c.want) != c.good {
			t.Errorf("strength %v: good chance %v (band's %v), want %v", c.strength, got, EpochGoodChanceIn(c.want), c.good)
		}
	}
	// A moderate town sits under the middle band with room to spare, and the
	// top band takes under three and a half times the moderate set.
	if moderate := 1 / FaithFullSets; moderate >= FaithMidAt || moderate < 0.2 {
		t.Errorf("a moderate town reads %.3f: want the low band, not far under %v", moderate, FaithMidAt)
	}
	if top := FaithHighAbove * FaithFullSets; top <= 3 || top >= 3.5 {
		t.Errorf("the top band starts at %.3f times the moderate set, want over 3 and under 3.5", top)
	}
}

// The rule itself: the faith held that the town's own faith buildings made,
// against four and a half moderate sets'.
func TestFaithStrengthOf(t *testing.T) {
	for _, c := range []struct {
		name string
		held float64
		m    FaithSave
		want float64
	}{
		{"nothing measured yet", 500, FaithSave{}, 0},
		{"no faith held", 0, FaithSave{Moderate: 100, Own: 100}, 0},
		{"a moderate town that kept everything", 100, FaithSave{Moderate: 100, Own: 100}, 1 / 4.5},
		{"twice the moderate set", 200, FaithSave{Moderate: 100, Own: 200}, 2 / 4.5},
		{"three times", 300, FaithSave{Moderate: 100, Own: 300}, 3 / 4.5},
		{"past full", 900, FaithSave{Moderate: 100, Own: 900}, 1},
		// A windfall is held too, and counts for nothing past the faith the
		// town's income made.
		{"a moderate town with a windfall", 100 + 5000, FaithSave{Moderate: 100, Own: 100}, 1 / 4.5},
		{"a windfall after spending", 50 + 5000, FaithSave{Moderate: 100, Own: 100}, 1 / 4.5},
		// A wonder's faith is held too, and counts for nothing.
		{"a moderate town with a wonder", 100 + 400, FaithSave{Moderate: 100, Own: 100, Other: 400}, 1 / 4.5},
		{"a wonder and no faith buildings", 400, FaithSave{Moderate: 100, Other: 400}, 0},
		{"three times, with a wonder", 300 + 400, FaithSave{Moderate: 100, Own: 300, Other: 400}, 3 / 4.5},
		// Spending takes its share of the strength.
		{"three times, half of it spent", 350, FaithSave{Moderate: 100, Own: 300, Other: 400}, 1.5 / 4.5},
	} {
		if got := FaithStrengthOf(c.held, c.m); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("%s: strength %v, want %v", c.name, got, c.want)
		}
	}
	m := FaithSave{Moderate: 100, Own: 300, Other: 400}
	if d, k := FaithDevotionOf(m), FaithKeptOf(350, m); d != 3 || k != 0.5 {
		t.Errorf("devotion %v and share kept %v, want 3 and 0.5", d, k)
	}
	if k := FaithKeptOf(9999, m); k != 1 {
		t.Errorf("a windfall past the run's income: share kept %v, want 1", k)
	}
	if d, k := FaithDevotionOf(FaithSave{}), FaithKeptOf(10, FaithSave{}); d != 0 || k != 0 {
		t.Errorf("nothing measured: devotion %v, kept %v, want 0 and 0", d, k)
	}
}

// Through the engine, tick by tick: a town's faith strength is its devotion
// over four and a half, whatever age it is in and whatever faith its wonders
// hand it. A moderate town reads 22%, the low band; twice the moderate set is
// the middle band, three times still is, and the top band takes a little
// under three and a half times. A town with wonders and no faith buildings
// holds plenty of faith and reads nothing.
func TestFaithStrengthFollowsDevotion(t *testing.T) {
	for _, age := range []string{"bronze_age", "iron_age", "medieval_age", "colonial_age", "victorian_age", "modern_age", "interstellar_age", "transcendent_age"} {
		// One moderate set's worth of empty buildings: a devotion of 1, less
		// what a worker output bonus adds to the moderate set's full crews
		// (the late ages' milestones give one).
		one := faithTown(t, age, 1)
		one.StepTicks(120)
		base := one.CatastropheOutlook().FaithDevotion
		if base > 1+1e-9 || base < 0.95 {
			t.Errorf("%s: one moderate set's worth of faith buildings reads a devotion of %.4f, want 1 (or a little under)", age, base)
		}
		for _, c := range []struct {
			sets float64
			band FaithBand
		}{{0, FaithBandLow}, {1, FaithBandLow}, {2, FaithBandMid}, {3, FaithBandMid}, {3.6, FaithBandHigh}} {
			ge := faithTown(t, age, c.sets)
			ge.StepTicks(120)
			o := ge.CatastropheOutlook()
			devotion := c.sets * base
			if math.Abs(o.FaithDevotion-devotion) > 1e-9 || math.Abs(o.FaithStrength-devotion/FaithFullSets) > 1e-9 || o.FaithBand != c.band {
				t.Errorf("%s, %v moderate sets: strength %.4f (%s), devotion %.3f; want %.4f (%s) and %.3f", age, c.sets, o.FaithStrength, o.FaithBand, o.FaithDevotion, devotion/FaithFullSets, c.band, devotion)
			}
			held := ge.Resources.Get("faith")
			if o.FaithKept != 1 && c.sets > 0 {
				t.Errorf("%s, %v sets: %.0f%% of %v faith kept, want all of it", age, c.sets, o.FaithKept*100, held)
			}
			if st := ge.GetState(); st.CatastropheOutlook != o {
				t.Errorf("%s, %v sets: GetState shows %+v, the roll reads %+v", age, c.sets, st.CatastropheOutlook, o)
			}
			// Reading takes nothing.
			if got := ge.Resources.Get("faith"); got != held {
				t.Errorf("%s: %v faith after the strength was read, %v before", age, got, held)
			}
			// A town with wonders and no faith buildings holds their faith and
			// reads nothing for it.
			if c.sets == 0 && age != "bronze_age" && (held <= 0 || o.FaithStrength != 0) {
				t.Errorf("%s, wonders and no faith buildings: %v faith held, strength %v; want faith in hand and no strength", age, held, o.FaithStrength)
			}
		}
	}
}

// No required building buys a band. Stonehenge makes 0.6 faith a tick where
// a moderate Bronze Age set makes 0.07, and the Sistine Chapel 1.8 where a
// Renaissance set makes 1.27; every player must build them to advance. With
// the wonder standing or not, from the age's first tick or not, a town reads
// what its faith buildings earn it.
func TestNoWonderBuysAFaithBand(t *testing.T) {
	for _, c := range []struct{ age, wonder string }{{"bronze_age", "stonehenge"}, {"renaissance_age", "sistine_chapel"}} {
		for _, sets := range []float64{0, 1, 2} {
			without := faithTown(t, c.age, sets)
			with := faithTown(t, c.age, sets)
			with.Buildings.counts[c.wonder] = 1
			late := faithTown(t, c.age, sets)
			without.StepTicks(400)
			with.StepTicks(400)
			late.StepTicks(200)
			late.Buildings.counts[c.wonder] = 1
			late.StepTicks(200)
			if w, wo := with.Resources.Get("faith"), without.Resources.Get("faith"); w < wo+100 {
				t.Fatalf("%s: %v faith with %s and %v without: the wonder makes none here", c.age, w, c.wonder, wo)
			}
			want := sets / FaithFullSets
			for name, ge := range map[string]*GameEngine{"without the wonder": without, "with it from the first tick": with, "with it built halfway": late} {
				if got := ge.faithStrength(); math.Abs(got-want) > 1e-9 {
					t.Errorf("%s, %v sets, %s: strength %.4f, want %.4f", c.age, sets, name, got, want)
				}
			}
		}
	}
}

// The measure is the town's own: its morale, its bonuses, the Cosmic Legacy,
// Era Mastery's speed and time spent away all move the faith it makes and
// what a moderate set would have made in its place alike, so none of them
// moves the strength.
func TestFaithStrengthIsTheTownsOwn(t *testing.T) {
	const age = "electric_age"
	want := 2 / FaithFullSets
	for name, stage := range map[string]func(ge *GameEngine){
		"plain":             func(*GameEngine) {},
		"high morale":       func(ge *GameEngine) { ge.morale = 1.5 },
		"low morale":        func(ge *GameEngine) { ge.morale = 0.3 },
		"the Cosmic Legacy": func(ge *GameEngine) { ge.cosmicLegacy = true },
		"mastered ground":   func(ge *GameEngine) { ge.Prestige.SetMastery(age, 9) },
		"a faith bonus":     func(ge *GameEngine) { ge.permanentBonuses["faith_rate"] = 0.5 },
		"all production":    func(ge *GameEngine) { ge.permanentBonuses["production_all"] = 1.5 },
	} {
		ge := faithTown(t, age, 2)
		stage(ge)
		ge.StepTicks(150)
		if got := ge.faithStrength(); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: strength %.5f, want %.5f", name, got, want)
		}
	}
	// A town away half the time makes half the faith, and so would the
	// moderate set in its place.
	ge := faithTown(t, age, 2)
	ge.StepTicks(50)
	before := ge.faithMeasure
	ge.SimulateOffline(3 * time.Hour)
	if ge.faithMeasure.Moderate <= before.Moderate || ge.faithMeasure.Own <= before.Own {
		t.Fatalf("time away added nothing to the measure: %+v, then %+v", before, ge.faithMeasure)
	}
	if got := ge.faithStrength(); math.Abs(got-want) > 1e-9 {
		t.Errorf("after three hours away: strength %.5f, want %.5f", got, want)
	}
}

// Carried-in faith counts for what it is. A town keeps its strength across
// an advance (the measure follows the run, not the age), a town that stays
// as devoted keeps it through the next age, and the epoch roll at the
// advance reads what the player was shown before it.
func TestFaithStrengthCarriesAcrossAges(t *testing.T) {
	for _, sets := range []float64{1, 2, 3.6} {
		ge := faithTown(t, "bronze_age", sets)
		ge.Buildings.counts["stonehenge"] = 1
		want := sets / FaithFullSets
		for _, next := range []string{"iron_age", "classical_age", "medieval_age", "renaissance_age"} {
			ge.StepTicks(300)
			shown := ge.GetState().CatastropheOutlook
			held := ge.Resources.Get("faith")
			ge.pendingCatastrophe = ""
			ge.advanceAge(next)
			after := ge.CatastropheOutlook()
			if got := ge.Resources.Get("faith"); got < held {
				t.Errorf("%v sets: the advance into %s took faith: %v, then %v", sets, next, held, got)
			}
			if math.Abs(shown.FaithStrength-want) > 1e-9 || math.Abs(after.FaithStrength-shown.FaithStrength) > 1e-9 || after.FaithBand != shown.FaithBand {
				t.Errorf("%v sets, into %s: strength %.4f before the advance and %.4f after, want %.4f both times", sets, next, shown.FaithStrength, after.FaithStrength, want)
			}
			// The new age has a new faith building: a town as devoted builds it.
			setFaithBuildings(ge, next, sets)
		}
	}

	// The epoch roll: a draw of 0.45 is a good event from the middle band up.
	roll := func(sets float64) string {
		ge := faithTown(t, "bronze_age", sets)
		ge.Buildings.counts["stonehenge"] = 1
		ge.StepTicks(300)
		ge.rng = riggedRNG(0.45, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5)
		ge.advanceAge("iron_age")
		if len(ge.epochEventHistory) == 0 {
			t.Fatal("no epoch event rolled on entering the Iron Era")
		}
		return ge.epochEventHistory[len(ge.epochEventHistory)-1].EventType
	}
	if got := roll(1); strings.HasPrefix(got, "good") {
		t.Errorf("a moderate town with Stonehenge, draw 0.45: %s, want a challenging event (the low band's 40%%)", got)
	}
	if got := roll(2); !strings.HasPrefix(got, "good") {
		t.Errorf("a town twice as devoted, draw 0.45: %s, want a good event (the middle band's 50%%)", got)
	}
}

// Spending and losing faith take their share of the strength, and making it
// back brings it up again. The faith itself is never clamped to the measure.
func TestFaithStrengthAfterSpending(t *testing.T) {
	ge := faithTown(t, "classical_age", 3.6)
	ge.StepTicks(300)
	top := ge.faithStrength()
	if FaithBandAt(top) != FaithBandHigh {
		t.Fatalf("a town of 3.6 sets reads %.3f, want the top band", top)
	}
	// Spend down to half of what the town's income has made.
	held, made := ge.Resources.Get("faith"), ge.faithMeasure.Own+ge.faithMeasure.Other
	if held < made || !ge.Resources.Remove("faith", held-made/2) {
		t.Fatalf("holding %v of the %v faith made: could not spend down to half", held, made)
	}
	o := ge.CatastropheOutlook()
	if math.Abs(o.FaithStrength-top/2) > 1e-9 || o.FaithBand != FaithBandMid || math.Abs(o.FaithKept-0.5) > 1e-9 || math.Abs(o.FaithDevotion-3.6) > 1e-9 {
		t.Errorf("half the faith spent: strength %.4f (%s), kept %.2f, devotion %.2f; want %.4f in the middle band, 0.50 and 3.60", o.FaithStrength, o.FaithBand, o.FaithKept, o.FaithDevotion, top/2)
	}
	ge.StepTicks(300)
	if got := ge.faithStrength(); got <= top/2 || got >= top {
		t.Errorf("after as long again: strength %.4f, want between %.4f and %.4f", got, top/2, top)
	}
	// Past full strength the faith stays where it is.
	ge = faithTown(t, "classical_age", 9)
	ge.StepTicks(200)
	held = ge.Resources.Get("faith")
	if got := ge.faithStrength(); got != 1 {
		t.Errorf("nine moderate sets: strength %v, want 1", got)
	}
	ge.recalculateRates()
	if got := ge.Resources.Get("faith"); got != held {
		t.Errorf("faith past full strength: %v held, %v after a recalculation", held, got)
	}
}

// A new run starts the measure over; a save carries it; and a save from
// before the measure existed is read as a moderate town that kept its faith,
// which is the low band every roll read then.
func TestFaithMeasureSaveLoadAndReset(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := faithTown(t, "iron_age", 2)
	ge.StepTicks(100)
	want, strength := ge.faithMeasure, ge.faithStrength()
	if want.Moderate <= 0 || want.Own <= 0 || want.Other <= 0 {
		t.Fatalf("measure after 100 ticks with Stonehenge standing: %+v", want)
	}
	if err := ge.SaveGame("faith"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("faith"); err != nil {
		t.Fatal(err)
	}
	if ge2.faithMeasure != want || math.Abs(ge2.faithStrength()-strength) > 1e-12 {
		t.Errorf("after load: measure %+v strength %v, want %+v and %v", ge2.faithMeasure, ge2.faithStrength(), want, strength)
	}

	// A save from before the measure: no marker, no measure.
	old := faithTown(t, "iron_age", 0)
	old.Resources.LoadAmounts(map[string]float64{"faith": 4321})
	old.restoreFaithState(&GameSave{})
	if got := old.faithStrength(); math.Abs(got-1/FaithFullSets) > 1e-12 || FaithBandAt(got) != FaithBandLow {
		t.Errorf("an older save holding faith: strength %v, want a moderate town's %v, the low band", got, 1/FaithFullSets)
	}
	// One written since, at a run's first tick, has the marker and nothing to
	// measure yet: it is left alone.
	fresh := faithTown(t, "iron_age", 0)
	fresh.Resources.LoadAmounts(map[string]float64{"faith": 4321})
	fresh.restoreFaithState(&GameSave{FaithMeasured: true})
	if fresh.faithMeasure != (FaithSave{}) {
		t.Errorf("a marked save with no measure was given one: %+v", fresh.faithMeasure)
	}
	// A hand-edited measure cannot poison the rolls.
	bad := faithTown(t, "iron_age", 0)
	bad.restoreFaithState(&GameSave{FaithMeasured: true, FaithMeasure: &FaithSave{Moderate: math.NaN(), Own: -5, Other: math.Inf(1)}})
	if bad.faithMeasure != (FaithSave{}) || bad.faithStrength() != 0 {
		t.Errorf("a broken measure loaded as %+v (strength %v), want it cleared", bad.faithMeasure, bad.faithStrength())
	}
	// The save of a run with nothing measured keeps the field out.
	if NewGameEngine().faithSaveCopy() != nil {
		t.Error("a new run saves a faith measure")
	}

	// A new run starts over.
	ge.clearHarbingerRun()
	if ge.faithMeasure != (FaithSave{}) || ge.faithStrength() != 0 {
		t.Errorf("after a run reset: measure %+v", ge.faithMeasure)
	}
}
