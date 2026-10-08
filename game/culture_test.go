package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// --- culture strength ---------------------------------------------------------

// cultureBuildingsUpTo lists the culture buildings a town can hold in age:
// every producer of culture that is not a wonder, up to age.
func cultureBuildingsUpTo(age string) []string {
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
			if e.Type == "production" && e.Target == "culture" && e.Value > 0 {
				keys = append(keys, d.Key)
				break
			}
		}
	}
	return keys
}

// culturePerSet is how many copies of a culture building make a moderate
// set's worth: config.FlowCopies, since culture buildings take no crew and
// an empty one makes all it can.
const culturePerSet = int(config.FlowCopies)

// cultureTown is faithTown for culture: an engine in age with sets moderate
// sets' worth of culture buildings, every wonder of an earlier age that
// makes culture standing, and a store nothing here could fill. sets may be
// fractional in steps of 1/5.
func cultureTown(t *testing.T, age string, sets float64) *GameEngine {
	t.Helper()
	ge := catEngine(t, age, 7)
	order := ageOrders()
	for _, a := range ageKeys() {
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
	}
	for _, key := range cultureBuildingsUpTo(age) {
		ge.Buildings.counts[key] = int(math.Round(sets * float64(culturePerSet)))
	}
	for _, d := range config.BaseBuildings() {
		if d.Category != "wonder" || order[d.RequiredAge] >= order[age] {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == "culture" {
				ge.Buildings.counts[d.Key] = 1
			}
		}
	}
	ge.Buildings.counts["quantum_vault"] = 1 // storage for everything, culture included
	return ge
}

// The tiers: Minor up to 40% of full strength, Major above it, the
// Legendary above 75%.
func TestCultureTierAt(t *testing.T) {
	for _, c := range []struct {
		strength float64
		want     CultureTier
	}{{0, CultureTierMinor}, {1 / CultureFullSets, CultureTierMinor}, {0.40, CultureTierMinor}, {0.41, CultureTierMajor},
		{2 / CultureFullSets, CultureTierMajor}, {0.75, CultureTierMajor}, {3.5 / CultureFullSets, CultureTierLegendary}, {1, CultureTierLegendary}} {
		if got := CultureTierAt(c.strength); got != c.want {
			t.Errorf("strength %.3f opens %s, want %s", c.strength, got, c.want)
		}
	}
	// What the thresholds are placed for: a moderate town stays on Minor
	// events, twice the culture buildings open Major ones, and three and a
	// half times open the Legendary.
	if CultureFullSets*CultureMajorAbove <= 1.5 || CultureFullSets*CultureMajorAbove >= 2 {
		t.Errorf("Major opens at %.2f moderate sets, want over one and a half and under two", CultureFullSets*CultureMajorAbove)
	}
	if CultureFullSets*CultureLegendaryAbove <= 3 || CultureFullSets*CultureLegendaryAbove >= 3.5 {
		t.Errorf("the Legendary opens at %.2f moderate sets, want over three and under three and a half", CultureFullSets*CultureLegendaryAbove)
	}
}

// The rule: devotion × the share kept / CultureFullSets, capped at 1.
func TestCultureStrengthOf(t *testing.T) {
	for _, c := range []struct {
		name string
		held float64
		m    CultureSave
		want float64
	}{
		{"nothing measured", 500, CultureSave{}, 0},
		{"a moderate town that kept it all", 100, CultureSave{Moderate: 100, Own: 100}, 1 / CultureFullSets},
		{"twice the buildings", 200, CultureSave{Moderate: 100, Own: 200}, 2 / CultureFullSets},
		{"twice the buildings, half spent", 100, CultureSave{Moderate: 100, Own: 200}, 1 / CultureFullSets},
		{"a wonder's culture is no devotion", 300, CultureSave{Moderate: 100, Own: 100, Other: 200}, 1 / CultureFullSets},
		{"bought culture past what was made counts for nothing", 5000, CultureSave{Moderate: 100, Own: 200}, 2 / CultureFullSets},
		{"bought culture makes up for culture spent", 200, CultureSave{Moderate: 100, Own: 200, Other: 0}, 2 / CultureFullSets},
		{"no culture buildings", 300, CultureSave{Moderate: 100, Other: 300}, 0},
		{"nine sets", 900, CultureSave{Moderate: 100, Own: 900}, 1},
	} {
		if got := CultureStrengthOf(c.held, c.m); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("%s: strength %v, want %v", c.name, got, c.want)
		}
	}
}

// A town reads what its culture buildings earn it, in every age culture is
// made in: one moderate set's worth is Minor events, twice opens Major
// ones, three and a half times the Legendary. Wonders and techs that make
// culture count for nothing.
func TestCultureStrengthFollowsDevotion(t *testing.T) {
	for _, age := range []string{"classical_age", "renaissance_age", "industrial_age", "modern_age", "space_age", "transcendent_age"} {
		one := cultureTown(t, age, 1)
		one.StepTicks(120)
		base := one.CatastropheOutlook().CultureDevotion
		if base > 1+1e-9 || base < 0.95 {
			t.Errorf("%s: one moderate set's worth of culture buildings reads a devotion of %.4f, want 1 (or a little under)", age, base)
		}
		for _, c := range []struct {
			sets float64
			tier CultureTier
		}{{0, CultureTierMinor}, {1, CultureTierMinor}, {2, CultureTierMajor}, {3, CultureTierMajor}, {3.6, CultureTierLegendary}} {
			ge := cultureTown(t, age, c.sets)
			ge.StepTicks(120)
			o := ge.CatastropheOutlook()
			devotion := c.sets * base
			if math.Abs(o.CultureDevotion-devotion) > 1e-9 || math.Abs(o.CultureStrength-devotion/CultureFullSets) > 1e-9 || o.CultureTier != c.tier {
				t.Errorf("%s, %v moderate sets: strength %.4f (%s), devotion %.3f; want %.4f (%s) and %.3f", age, c.sets, o.CultureStrength, o.CultureTier, o.CultureDevotion, devotion/CultureFullSets, c.tier, devotion)
			}
			if o.CultureKept < 1-1e-9 && c.sets > 0 {
				t.Errorf("%s, %v sets: %.0f%% of the culture kept, want all of it", age, c.sets, o.CultureKept*100)
			}
			if st := ge.GetState(); st.CatastropheOutlook != o {
				t.Errorf("%s, %v sets: GetState shows %+v, the roll reads %+v", age, c.sets, st.CatastropheOutlook, o)
			}
		}
	}
	// Before culture there is nothing to read.
	early := cultureTown(t, "iron_age", 1)
	early.StepTicks(120)
	if o := early.CatastropheOutlook(); o.CultureStrength != 0 || o.CultureTier != CultureTierMinor || early.cultureMeasure != (CultureSave{}) {
		t.Errorf("the Iron Age, before culture: strength %v, measure %+v", o.CultureStrength, early.cultureMeasure)
	}
}

// Spending culture takes its share of the strength, as spending faith does:
// an Appease paid in culture, a festival. Culture bought at the market does
// not raise it past what the town's own buildings earned, which is how the
// old rule (the fill of culture's store, the general one) was met.
func TestCultureStrengthAfterSpendingAndBuying(t *testing.T) {
	ge := cultureTown(t, "renaissance_age", 3.6)
	ge.StepTicks(300)
	top := ge.cultureStrength()
	if CultureTierAt(top) != CultureTierLegendary {
		t.Fatalf("a town of 3.6 sets reads %.3f, want the Legendary tier open", top)
	}
	held, made := ge.Resources.Get("culture"), ge.cultureMeasure.Own+ge.cultureMeasure.Other
	if held < made || !ge.Resources.Remove("culture", held-made/2) {
		t.Fatalf("holding %v of the %v culture made: could not spend down to half", held, made)
	}
	o := ge.CatastropheOutlook()
	if math.Abs(o.CultureStrength-top/2) > 1e-9 || o.CultureTier != CultureTierMinor || math.Abs(o.CultureKept-0.5) > 1e-9 {
		t.Errorf("half the culture spent: strength %.4f (%s), kept %.2f; want %.4f, Minor only, 0.50", o.CultureStrength, o.CultureTier, o.CultureKept, top/2)
	}
	// Buying it back restores what was spent and no more.
	ge.Resources.Add("culture", made*10)
	if got := ge.cultureStrength(); math.Abs(got-top) > 1e-9 {
		t.Errorf("after buying ten times the culture made: strength %.4f, want the %.4f the buildings earned", got, top)
	}

	// A moderate town with a store full of bought culture stays on Minor
	// events: under the old rule it had the Legendary tier.
	rich := cultureTown(t, "renaissance_age", 1)
	rich.StepTicks(300)
	rich.Resources.Add("culture", float64(0.9*rich.Resources.GetStorage("culture")))
	if fill := rich.Resources.Get("culture") / rich.Resources.GetStorage("culture"); fill < 0.75 {
		t.Fatalf("the store is %.2f full of culture, want over three quarters", fill)
	}
	if o := rich.CatastropheOutlook(); o.CultureTier != CultureTierMinor || math.Abs(o.CultureStrength-o.CultureDevotion/CultureFullSets) > 1e-9 {
		t.Errorf("a moderate town with a store of bought culture: strength %.4f (%s), want a moderate town's, Minor only", o.CultureStrength, o.CultureTier)
	}

	// An Appease paid in culture lowers it.
	th := threadEngine(t, "steel_era", 4)
	price := threadAppeaseCost(th.harbinger, 1)
	if price["culture"] <= 0 {
		t.Fatalf("a Steel Era Appease asks no culture: %v", price)
	}
	th.setFaithMeasure(10*price["faith"], FaithSave{Moderate: 10 * price["faith"], Own: 10 * price["faith"]})
	th.setCultureMeasure(4*price["culture"], CultureSave{Moderate: price["culture"], Own: 4 * price["culture"]})
	before := th.cultureStrength()
	if err := th.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	if got, want := th.cultureStrength(), before*0.75; math.Abs(got-want) > 1e-9 || CultureTierAt(before) != CultureTierLegendary || CultureTierAt(got) != CultureTierMajor {
		t.Errorf("Appease took a quarter of the culture: strength %.4f after %.4f, want %.4f (the Legendary tier down to Major)", got, before, want)
	}
}

// The good epoch event's tier follows culture strength: Minor events only
// for a moderate town, Major ones too with twice the culture buildings, and
// the Legendary on some rolls with three and a half times (it is one event
// of ten, on 15% of the rolls: about one roll in 67).
func TestGoodEpochEventTierFollowsCultureStrength(t *testing.T) {
	const rolls = 1500
	types := map[string]string{}
	for _, ev := range config.GoodEpochEvents() {
		types[ev.Key] = ev.Type
	}
	tiers := func(sets float64) map[string]int {
		ge := cultureTown(t, "renaissance_age", sets)
		ge.StepTicks(120)
		want := CultureTierAt(ge.cultureStrength())
		ge.SeedRNG(11)
		seen := map[string]int{}
		for i := 0; i < rolls; i++ {
			ge.epochEventHistory = nil
			ge.rollGoodEpochEvent()
			if len(ge.epochEventHistory) != 1 {
				t.Fatalf("%v sets, roll %d: %d events rolled", sets, i, len(ge.epochEventHistory))
			}
			seen[types[ge.epochEventHistory[0].EventKey]]++
			// An event's gifts (culture among them) do not move the tier.
			if got := CultureTierAt(ge.cultureStrength()); got != want {
				t.Fatalf("%v sets, roll %d: the tier went from %s to %s", sets, i, want, got)
			}
		}
		return seen
	}
	if got := tiers(1); got["good_minor"] != rolls {
		t.Errorf("a moderate town's %d good rolls: %v, want Minor events only", rolls, got)
	}
	if got := tiers(2); got["good_major"] < rolls/4 || got["good_legendary"] != 0 || got["good_minor"] < rolls/4 {
		t.Errorf("twice the culture buildings, %d good rolls: %v, want Minor and Major events and no Legendary", rolls, got)
	}
	if got := tiers(3.6); got["good_legendary"] < rolls/200 || got["good_legendary"] > rolls/25 || got["good_major"] < rolls/4 {
		t.Errorf("3.6 times the culture buildings, %d good rolls: %v, want Major events and about one Legendary in 67", rolls, got)
	}
}

// A new run starts the measure over; a save carries it; and a save from
// before the measure existed is read as a moderate town that kept its
// culture: Minor events, where every town stood.
func TestCultureMeasureSaveLoadAndReset(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := cultureTown(t, "renaissance_age", 2)
	ge.StepTicks(100)
	want, strength := ge.cultureMeasure, ge.cultureStrength()
	if want.Moderate <= 0 || want.Own <= 0 {
		t.Fatalf("measure after 100 ticks: %+v", want)
	}
	if err := ge.SaveGame("culture"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("culture"); err != nil {
		t.Fatal(err)
	}
	if ge2.cultureMeasure != want || math.Abs(ge2.cultureStrength()-strength) > 1e-12 {
		t.Errorf("after load: measure %+v strength %v, want %+v and %v", ge2.cultureMeasure, ge2.cultureStrength(), want, strength)
	}

	old := cultureTown(t, "renaissance_age", 0)
	old.Resources.LoadAmounts(map[string]float64{"culture": 4321})
	old.restoreCultureState(&GameSave{})
	if got := old.cultureStrength(); math.Abs(got-1/CultureFullSets) > 1e-12 || CultureTierAt(got) != CultureTierMinor {
		t.Errorf("an older save holding culture: strength %v, want a moderate town's %v", got, 1/CultureFullSets)
	}
	fresh := cultureTown(t, "renaissance_age", 0)
	fresh.Resources.LoadAmounts(map[string]float64{"culture": 4321})
	fresh.restoreCultureState(&GameSave{CultureMeasured: true})
	if fresh.cultureMeasure != (CultureSave{}) {
		t.Errorf("a marked save with no measure was given one: %+v", fresh.cultureMeasure)
	}
	bad := cultureTown(t, "renaissance_age", 0)
	bad.restoreCultureState(&GameSave{CultureMeasured: true, CultureMeasure: &CultureSave{Moderate: math.NaN(), Own: -5, Other: math.Inf(1)}})
	if bad.cultureMeasure != (CultureSave{}) || bad.cultureStrength() != 0 {
		t.Errorf("a broken measure loaded as %+v (strength %v), want it cleared", bad.cultureMeasure, bad.cultureStrength())
	}
	if NewGameEngine().cultureSaveCopy() != nil {
		t.Error("a new run saves a culture measure")
	}

	ge.clearHarbingerRun()
	if ge.cultureMeasure != (CultureSave{}) || ge.cultureStrength() != 0 {
		t.Errorf("after a run reset: measure %+v", ge.cultureMeasure)
	}
}
