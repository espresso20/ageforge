package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// --- faith standing -----------------------------------------------------------

// The bands: under a quarter of full standing is low, over three quarters
// high, and an age with no measure of faith reads as the middle.
func TestFaithBandAt(t *testing.T) {
	for _, c := range []struct {
		standing float64
		ok       bool
		want     FaithBand
		good     float64
	}{
		{0, true, FaithBandLow, 0.40},
		{0.2499, true, FaithBandLow, 0.40},
		{0.25, true, FaithBandMid, 0.50},
		{0.75, true, FaithBandMid, 0.50},
		{0.7501, true, FaithBandHigh, 0.60},
		{1, true, FaithBandHigh, 0.60},
		{0, false, FaithBandMid, 0.50},
		{1, false, FaithBandMid, 0.50},
	} {
		if got := FaithBandAt(c.standing, c.ok); got != c.want {
			t.Errorf("standing %v (measured %v): band %s, want %s", c.standing, c.ok, got, c.want)
		}
		if got := goodChanceFor(c.standing, c.ok); got != c.good || EpochGoodChanceIn(c.want) != c.good {
			t.Errorf("standing %v (measured %v): good chance %v (band's %v), want %v", c.standing, c.ok, got, EpochGoodChanceIn(c.want), c.good)
		}
	}
}

// The problem the standing replaced, staged in a late age: with one Stellar
// Vault built, the store faith is kept in is two quadrillion, and what a
// devoted faith economy makes in the whole Interstellar Age is millionths of
// it. Read as a share of storage, as it used to be, no faith a run could
// hold leaves the bottom band, and the Last Passage always rolls at 18%.
// Read as a standing, the same faith moves through all three bands.
func TestFaithStandingReadsTheAgeNotTheStore(t *testing.T) {
	const age = "interstellar_age"
	ge := lpEngine(t, age, 3)
	ge.Buildings.counts["stellar_vault"] = 1
	ge.recalculateRates()
	store := ge.Resources.GetStorage("faith")
	if store < 2e15 {
		t.Fatalf("faith storage with one Stellar Vault is %v, want the vault's 2Q for every resource", store)
	}
	perTick, ageTicks := config.FlowIncome("faith", age), config.AgeTargetTicks(age)
	shortest := perTick * ageTicks * harbingerLeadMin // a moderate economy, a doom's shortest warning
	wholeAge := perTick * ageTicks
	full := FaithStandingFullIn(ge.rules, age)
	if want := wholeAge * harbingerLeadMax; math.Abs(full-want) > 1e-6*want {
		t.Fatalf("full standing in %s is %v, want what a moderate economy makes in three fifths of the age (%v)", age, full, want)
	}
	for _, c := range []struct {
		name     string
		faith    float64
		band     FaithBand
		passage  float64 // the Last Passage's chance
		standing float64
	}{
		{"none", 0, FaithBandLow, 0.18, 0},
		{"a moderate economy, a doom's shortest warning", shortest, FaithBandMid, 0.15, 1.0 / 3},
		{"a devoted one (3x), the same warning", 3 * shortest, FaithBandHigh, 0.12, 1},
		{"a moderate economy, the whole age", wholeAge, FaithBandHigh, 0.12, 1},
		{"a devoted one, the whole age", 3 * wholeAge, FaithBandHigh, 0.12, 1},
	} {
		ge.Resources.LoadAmounts(map[string]float64{"faith": c.faith})
		// What the old measure read: the share of the store.
		if fill := c.faith / store; fill >= 1e-5 {
			t.Errorf("%s: %v faith is %v of the store; the test means to show it is next to nothing", c.name, c.faith, fill)
		}
		o := ge.CatastropheOutlook()
		if o.FaithBand != c.band || math.Abs(o.FaithStanding-c.standing) > 1e-9 || o.FaithFull != full || math.Abs(o.Probability-c.passage) > 1e-9 {
			t.Errorf("%s: outlook standing %v (%s) of %v, the Last Passage at %v; want %v (%s) and %v", c.name, o.FaithStanding, o.FaithBand, o.FaithFull, o.Probability, c.standing, c.band, c.passage)
		}
		if st := ge.GetState(); st.CatastropheOutlook != o {
			t.Errorf("%s: GetState shows %+v, the roll reads %+v", c.name, st.CatastropheOutlook, o)
		}
		// Reading the standing takes nothing: the faith is still there, past
		// full standing included, and a recalculation leaves it.
		ge.recalculateRates()
		if got := ge.Resources.Get("faith"); got != c.faith {
			t.Errorf("%s: %v faith held after the standing was read, want %v untouched", c.name, got, c.faith)
		}
	}
}

// In every age, the engine reads a moderate faith economy that saved through
// a doom's shortest warning as the middle band and a devoted one as the top,
// whatever storage the player has built, and Era Mastery does not move the
// measure.
func TestFaithStandingReachableInEveryAge(t *testing.T) {
	for _, m := range FaithMeasuresIn(rules.Core()) {
		if m.Full <= 0 || m.WarningTicks != harbingerLeadMin*config.AgeTargetTicks(m.Age) || m.Epoch != config.EpochForAge(m.Age) {
			t.Fatalf("%s: measure %+v", m.Age, m)
		}
		moderate := config.FlowIncome("faith", m.Age) * m.WarningTicks
		for _, c := range []struct {
			faith float64
			band  FaithBand
		}{{0, FaithBandLow}, {moderate, FaithBandMid}, {3 * moderate, FaithBandHigh}} {
			for _, mastery := range []int{0, 4} {
				ge := catEngine(t, m.Age, 1)
				ge.Prestige.SetMastery(m.Age, mastery)
				// A general store far past anything faith could fill.
				ge.Resources.LoadStorage(map[string]float64{"faith": 1e30})
				ge.Resources.LoadAmounts(map[string]float64{"faith": c.faith})
				if o := ge.CatastropheOutlook(); o.FaithBand != c.band || o.FaithFull != m.Full {
					t.Errorf("%s at mastery %d with %v faith: standing %v (%s) of %v, want %s of %v", m.Age, mastery, c.faith, o.FaithStanding, o.FaithBand, o.FaithFull, c.band, m.Full)
				}
			}
		}
	}
	if n := len(FaithMeasuresIn(rules.Core())); n != len(config.AgeOrder()) {
		t.Errorf("%d faith measures for %d ages", n, len(config.AgeOrder()))
	}
}

// The epoch roll at an advance reads the standing in the age being left, the
// one the player was shown before advancing. Measured in the new age, which
// asks for several times the faith, a full standing would read as a low one.
func TestEpochRollReadsTheAgeBeingLeft(t *testing.T) {
	roll := func(standing, draw float64) string {
		ge := catEngine(t, "bronze_age", 5)
		setFaithStanding(ge, standing)
		held := ge.Resources.Get("faith")
		if shown := ge.GetState().CatastropheOutlook; math.Abs(shown.FaithStanding-standing) > 1e-9 {
			t.Fatalf("shown standing %v, want %v", shown.FaithStanding, standing)
		}
		if inNew, _ := ge.faithStandingIn("iron_age"); standing >= 0.9 && inNew >= FaithMidAt {
			t.Fatalf("%v faith is %v of the Iron Age's full standing: the test needs it under the middle band there", held, inNew)
		}
		ge.rng = riggedRNG(draw, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5)
		ge.advanceAge("iron_age")
		if got := ge.Resources.Get("faith"); got < held {
			t.Errorf("the advance took faith: %v held, %v after", held, got)
		}
		if n := len(ge.epochEventHistory); n == 0 {
			t.Fatal("no epoch event rolled on entering the Iron Era")
		}
		return ge.epochEventHistory[len(ge.epochEventHistory)-1].EventType
	}
	// A draw of 0.55 is a good event only at the high band's 60%.
	if got := roll(0.9, 0.55); !strings.HasPrefix(got, "good") {
		t.Errorf("high standing in the Bronze Age, draw 0.55: %s, want a good event (60%%)", got)
	}
	if got := roll(0.5, 0.55); strings.HasPrefix(got, "good") {
		t.Errorf("mid standing, draw 0.55: %s, want a challenging event (50%%)", got)
	}
	// A draw of 0.45 is a good event from the middle band up.
	if got := roll(0.5, 0.45); !strings.HasPrefix(got, "good") {
		t.Errorf("mid standing, draw 0.45: %s, want a good event (50%%)", got)
	}
	if got := roll(0.1, 0.45); strings.HasPrefix(got, "good") {
		t.Errorf("low standing, draw 0.45: %s, want a challenging event (40%%)", got)
	}
}

// A doom's Appease is priced on the measure the standing reads: level 1 asks
// a quarter of full standing in faith, the bottom of the middle band, and
// both of an ordinary doom's levels three quarters, the bottom of the top.
// So faith worth one Appease is the middle band, and paying it from a
// standing built in the shortest warning drops the band and still lowers the
// odds.
func TestAppeaseAndStandingShareAMeasure(t *testing.T) {
	for _, ep := range config.Epochs() {
		if !config.FateAllowed(ep.Key) {
			continue
		}
		for _, age := range ep.Ages {
			full := FaithStandingFullIn(rules.Core(), age)
			l1 := doomAppeaseCost(ep.Key, age, 1)["faith"]
			// Rounded up to two significant figures: within 10% over.
			if share := l1 / full; share < FaithMidAt || share > 1.1*FaithMidAt {
				t.Errorf("a doom of the %s foretold in the %s: Appease level 1 asks %v faith, %.3f of full standing; want about %v", ep.Key, age, l1, share, FaithMidAt)
			}
			if config.IsFinalEpoch(ep.Key) {
				continue
			}
			both := l1 + doomAppeaseCost(ep.Key, age, 2)["faith"]
			if share := both / full; share < FaithHighAbove || share > 1.1*FaithHighAbove {
				t.Errorf("a doom of the %s foretold in the %s: both Appease levels ask %v faith, %.3f of full standing; want about %v", ep.Key, age, both, share, FaithHighAbove)
			}
		}
	}
	// Middle band → pay level 1 → bottom band at ×0.6: 54%, under the 75% it was.
	ge := fateEngine(t, "classical_age", 2)
	forceFate(t, ge, 26000)
	tickTo(ge, arrivalTick(ge))
	setFaithStanding(ge, 1.0/3)
	before := ge.strikeChance()
	if err := ge.HarbingerAppease(); err != nil {
		t.Fatal(err)
	}
	after := ge.strikeChance()
	if o := ge.CatastropheOutlook(); math.Abs(before-0.75) > 1e-9 || math.Abs(after-0.54) > 1e-9 || o.FaithBand != FaithBandLow {
		t.Errorf("Appease from a third of full standing: %v before, %v after in the %s band; want 0.75, then 0.54 in the low band", before, after, o.FaithBand)
	}
}
