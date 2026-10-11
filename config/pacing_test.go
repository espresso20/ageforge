package config

import (
	"math"
	"strings"
	"testing"
	"time"
)

// TestAgeTargetsCoverEveryAge: every age has a target, so every building and
// tech gets a payback, a build-time cap and a research-time cap.
func TestAgeTargetsCoverEveryAge(t *testing.T) {
	for _, a := range Ages() {
		if AgeTargets[a.Key] <= 0 {
			t.Errorf("age %s has no entry in AgeTargets", a.Key)
		}
	}
}

// TestAgeTargetsGolden pins the one-week curve: the base targets × 2.6 from
// the Bronze Age on, the Primitive and Stone Ages as they were. A change here
// changes every producer's rate, so it should be deliberate (and smoke's copy,
// smoke/targets.go, must follow).
func TestAgeTargetsGolden(t *testing.T) {
	m, h := time.Minute, time.Hour
	want := map[string]time.Duration{
		"primitive_age": 15 * m, "stone_age": 45 * m, "bronze_age": 3*h + 54*m,
		"iron_age": 6*h + 30*m, "classical_age": 9*h + 6*m, "medieval_age": 11*h + 42*m,
		"renaissance_age": 15*h + 36*m, "colonial_age": 18*h + 12*m, "industrial_age": 20*h + 48*m,
		"victorian_age": 23*h + 24*m, "electric_age": 26 * h, "atomic_age": 31*h + 12*m,
		"modern_age": 31*h + 12*m, "information_age": 36*h + 24*m, "digital_age": 41*h + 36*m,
		"cyberpunk_age": 46*h + 48*m, "fusion_age": 52 * h, "space_age": 57*h + 12*m,
		"interstellar_age": 62*h + 24*m, "galactic_age": 62*h + 24*m, "quantum_age": 62*h + 24*m,
		"transcendent_age": 62*h + 24*m,
	}
	if len(AgeTargets) != len(want) {
		t.Errorf("AgeTargets has %d ages, want %d", len(AgeTargets), len(want))
	}
	for age, d := range want {
		if got := AgeTargets[age]; got != d {
			t.Errorf("AgeTargets[%s] = %v, want %v", age, got, d)
		}
	}
	var toModern time.Duration
	for _, a := range AgeOrder() {
		if a == "modern_age" {
			break
		}
		toModern += AgeTargets[a]
	}
	if toModern < 160*h || toModern > 175*h {
		t.Errorf("targets to the Modern Age sum to %v; the curve is about a week", toModern)
	}
}

// TestAgeStretch pins the clock factor: 1 for the Primitive and Stone Ages
// and anything unknown, PacingStretch from the Bronze Age on, and
// StretchTicks rounds to the nearest tick.
func TestAgeStretch(t *testing.T) {
	if PacingStretch != 2.6 {
		t.Errorf("PacingStretch = %v, want 2.6 (the one-week curve)", PacingStretch)
	}
	for i, a := range AgeOrder() {
		want := PacingStretch
		if i < 2 {
			want = 1
		}
		if got := AgeStretch(a); got != want {
			t.Errorf("AgeStretch(%s) = %v, want %v", a, got, want)
		}
	}
	for _, a := range []string{"", "no_such_age"} {
		if got := AgeStretch(a); got != 1 {
			t.Errorf("AgeStretch(%q) = %v, want 1", a, got)
		}
	}
	for _, c := range []struct {
		age        string
		ticks, out int
	}{
		{"stone_age", 216, 216}, {"bronze_age", 216, 562}, {"iron_age", 25, 65}, {"atomic_age", 40, 104},
		{"modern_age", 300, 780}, {"bronze_age", 150, 390}, {"bronze_age", 600, 1560}, {"iron_age", 0, 0},
		{"iron_age", 10000, 26000}, {"iron_age", 50000, 130000},
	} {
		if got := StretchTicks(c.age, c.ticks); got != c.out {
			t.Errorf("StretchTicks(%s, %d) = %d, want %d", c.age, c.ticks, got, c.out)
		}
	}
}

// TestEveryAgeHasAFeed: the market turns any construction resource of an age
// into any other at parity, so an age can grow as long as one of its
// construction resources flows in from a building a player can own by then:
// one of that age's, or an older one, since old buildings keep producing.
// (This is what the Renaissance lacked for steel before the rebalance, and
// every age from the Modern on for steel.)
func TestEveryAgeHasAFeed(t *testing.T) {
	defs := BaseBuildings()
	levels := priceLevels(defs)
	order := map[string]int{}
	for i, a := range Ages() {
		order[a.Key] = i
	}
	for _, a := range Ages() {
		lv := levels[a.Key]
		if len(lv) == 0 {
			continue
		}
		fed := false
		for _, d := range defs {
			if d.RequiredAge == "" || order[d.RequiredAge] > order[a.Key] || d.Category == "wonder" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 && lv[e.Target] > 0 {
					fed = true
				}
			}
		}
		if !fed {
			t.Errorf("%s: no building a player can own produces any of its construction resources", a.Key)
		}
	}
}

// TestMarketNeverBeatsBuilding: a round trip through the market at parity
// always loses value, so trading can't manufacture resources.
func TestMarketNeverBeatsBuilding(t *testing.T) {
	for _, a := range Ages() {
		for _, x := range MarketPairs(a.Key) {
			back, ok := MarketRate(x.To, x.From, a.Key)
			if !ok {
				continue
			}
			lv := exchangeLevels()[a.Key]
			if lv[x.From] <= 0 || lv[x.To] <= 0 {
				continue // a listed flow pair keeps its literal rates
			}
			if x.BaseRate*back >= 1 {
				t.Errorf("%s: %s->%s->%s returns %.3g of what went in", a.Key, x.From, x.To, x.From, x.BaseRate*back)
			}
		}
	}
}

// TestBuildAndResearchCaps: nothing takes longer to build or research than
// its age's share (the economy design's Law 2).
func TestBuildAndResearchCaps(t *testing.T) {
	for _, d := range BaseBuildings() {
		div := buildShare
		if d.Category == "storage" {
			div = storageBuildShare
		}
		if limit := AgeTargetTicks(d.RequiredAge) / div; limit > 0 && float64(d.BuildTicks) > limit {
			t.Errorf("%s builds in %d ticks, over its cap of %.0f", d.Key, d.BuildTicks, limit)
		}
	}
	for _, tech := range Technologies() {
		if limit := AgeTargetTicks(tech.Age) / researchShare; limit > 0 && float64(tech.ResearchTicks) > limit {
			t.Errorf("%s researches in %d ticks, over its cap of %.0f", tech.Key, tech.ResearchTicks, limit)
		}
	}
}

// TestWondersSizedToTheirAge: every wonder costs wonderPriceUnits of its age
// in construction resources (rounding aside).
func TestWondersSizedToTheirAge(t *testing.T) {
	levels := priceLevels(BaseBuildings())
	for _, d := range BaseBuildings() {
		if d.Category != "wonder" {
			continue
		}
		u := priceUnits(d.BaseCost, levels[d.RequiredAge])
		if math.Abs(u/wonderPriceUnits-1) > 0.05 {
			t.Errorf("%s costs %.1f price units of %s, want %.0f", d.Key, u, d.RequiredAge, wonderPriceUnits)
		}
	}
}

func TestFormatRateValue(t *testing.T) {
	cases := map[float64]string{0.2: "0.2", 0.7111: "0.711", 44.83: "44.8", 3344: "3.34K", 2.72e6: "2.72M", 1.5e15: "1.5Q"}
	for v, want := range cases {
		if got := FormatRateValue(v); got != want {
			t.Errorf("FormatRateValue(%v) = %q, want %q", v, got, want)
		}
	}
	if strings.Contains(FormatRateValue(999.96), "1000") {
		t.Errorf("FormatRateValue(999.96) = %q; should roll over to 1K", FormatRateValue(999.96))
	}
}

// The shares the caps and the wonder price are checked against. They were
// constants of the load-time rules; the tables are written out now, so they
// live with the tests that hold the tables to them.
const (
	buildShare        = 6.0  // no building takes longer than a sixth of its age's target
	storageBuildShare = 48.0 // a storage building, a forty-eighth
	researchShare     = 8.0  // no tech takes longer than an eighth of its age's target
	wonderPriceUnits  = 40.0 // a wonder costs forty price units of its age
)
