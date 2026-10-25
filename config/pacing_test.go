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
	if len(AgeTargets) != len(Ages()) {
		t.Errorf("AgeTargets has %d ages, want %d", len(AgeTargets), len(Ages()))
	}
	for _, a := range Ages() {
		if AgeTargets[a.Key] <= 0 {
			t.Errorf("age %s has no entry in AgeTargets", a.Key)
		}
	}
}

// TestAgeTargetsAboutAWeek: the targets up to the Modern Age sum to about a
// week of play. A change to any age's target changes every producer's rate, so
// TestGoldenTables pins the targets by age; this holds the curve's total.
func TestAgeTargetsAboutAWeek(t *testing.T) {
	var toModern time.Duration
	for _, a := range AgeOrder() {
		if a == "modern_age" {
			break
		}
		toModern += AgeTargets[a]
	}
	if toModern < 160*time.Hour || toModern > 175*time.Hour {
		t.Errorf("targets to the Modern Age sum to %v; the curve is about a week", toModern)
	}
}

// TestAgeStretch holds the clock factor: 1 for the Primitive and Stone Ages
// and anything unknown, PacingStretch from the Bronze Age on, and
// StretchTicks rounds to the nearest tick.
func TestAgeStretch(t *testing.T) {
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
		ticks int
		s     float64
		out   int
	}{
		{216, 1, 216}, {216, 2.6, 562}, {25, 2.6, 65}, {0, 2.6, 0}, {150, 2.6, 390}, {10000, 2.6, 26000},
	} {
		if got := StretchTicksBy(c.ticks, c.s); got != c.out {
			t.Errorf("StretchTicksBy(%d, %v) = %d, want %d", c.ticks, c.s, got, c.out)
		}
	}
	for _, a := range AgeOrder() {
		if got, want := StretchTicks(a, 216), StretchTicksBy(216, AgeStretch(a)); got != want {
			t.Errorf("StretchTicks(%s, 216) = %d, want %d", a, got, want)
		}
	}
}

// TestEveryAgeHasAFeed: an age can grow as long as one of its construction
// resources flows in from a building a player can own by then: one of that
// age's, or an older one, since old buildings keep producing. (This is what
// the Renaissance lacked for steel before the rebalance, and every age from
// the Modern on for steel.)
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

// TestBuildCaps: nothing takes longer to build than its age's share (the
// economy design's Law 2).
func TestBuildCaps(t *testing.T) {
	for _, d := range BaseBuildings() {
		if limit := BuildingBuildCapTicks(d.RequiredAge, d.Category); limit > 0 && float64(d.BuildTicks) > math.Ceil(limit) {
			t.Errorf("%s builds in %d ticks, over its cap of %.0f", d.Key, d.BuildTicks, limit)
		}
	}
}

// TestWondersSizedToTheirAge: every wonder costs WonderPriceUnits of its age
// in construction resources (rounding aside), counted in the price units from
// before new copies took NewCopyPriceFactor: a wonder is set against what the
// town earns, which did not move, so it is WonderPriceUnitsToday of today's.
func TestWondersSizedToTheirAge(t *testing.T) {
	levels := priceLevels(BaseBuildings())
	for _, d := range BaseBuildings() {
		if d.Category != "wonder" {
			continue
		}
		u := priceUnits(d.BaseCost, levels[d.RequiredAge])
		if math.Abs(u/WonderPriceUnitsToday-1) > 0.05 {
			t.Errorf("%s costs %.1f price units of %s, want %.1f (%.0f before the new-copy factor)", d.Key, u, d.RequiredAge, WonderPriceUnitsToday, WonderPriceUnits)
		}
	}
}

func TestFormatRateValue(t *testing.T) {
	cases := map[float64]string{0.2: "0.2", 0.7111: "0.711", 44.83: "44.8", 3344: "3.34K", 2.72 * Million: "2.72M", 1.5 * Quadrillion: "1.5Q"}
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
)

// BuildingBuildCapTicks gives wonders WonderBuildFactor times the general cap: the long build that ends an age.
func BuildingBuildCapTicks(age, category string) float64 {
	div := 6.0
	if category == "storage" {
		div = 48
	}
	limit := LatePriceTargetTicks(age) / div
	if category == "wonder" {
		limit *= WonderBuildFactor
	}
	return limit
}
