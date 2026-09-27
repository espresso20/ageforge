package config

import (
	"math"
	"regexp"
	"strings"
	"testing"
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

// TestPaybackRule: a production building's output of a construction
// resource repays the price of its first copy in PaybackTicks, valued at
// its age's parity (economy.md, Law 3).
func TestPaybackRule(t *testing.T) {
	levels := priceLevels(BaseBuildings())
	checked := 0
	for _, d := range BaseBuildings() {
		if d.Category == "wonder" {
			continue
		}
		lv := levels[d.RequiredAge]
		n := 0
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && lv[e.Target] > 0 {
				n++
			}
		}
		for _, e := range d.Effects {
			if e.Type != "production" || e.Value <= 0 || lv[e.Target] <= 0 {
				continue
			}
			// Output over one payback, in price units of the age.
			earned := e.Value * PaybackTicks(d.RequiredAge) / lv[e.Target] * float64(n)
			price := priceUnits(d.BaseCost, lv)
			if math.Abs(earned/price-1) > 0.01 {
				t.Errorf("%s: %s %.4g/tick earns %.3g price units per payback, first copy costs %.3g", d.Key, e.Target, e.Value, earned, price)
			}
			checked++
		}
	}
	if checked < 80 {
		t.Errorf("only %d production effects checked; the rule should cover most producers", checked)
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
// its age's share (economy.md, Law 2).
func TestBuildAndResearchCaps(t *testing.T) {
	for _, d := range BaseBuildings() {
		div := BuildTimeDivisor
		if d.Category == "storage" {
			div = StorageBuildTimeDivisor
		}
		if limit := AgeTargetTicks(d.RequiredAge) / div; limit > 0 && float64(d.BuildTicks) > limit {
			t.Errorf("%s builds in %d ticks, over its cap of %.0f", d.Key, d.BuildTicks, limit)
		}
	}
	for _, tech := range Technologies() {
		if limit := AgeTargetTicks(tech.Age) / ResearchTimeDivisor; limit > 0 && float64(tech.ResearchTicks) > limit {
			t.Errorf("%s researches in %d ticks, over its cap of %.0f", tech.Key, tech.ResearchTicks, limit)
		}
	}
}

// TestWondersSizedToTheirAge: every wonder costs WonderPriceUnits of its age
// in construction resources (rounding aside).
func TestWondersSizedToTheirAge(t *testing.T) {
	levels := priceLevels(BaseBuildings())
	for _, d := range BaseBuildings() {
		if d.Category != "wonder" {
			continue
		}
		u := priceUnits(d.BaseCost, levels[d.RequiredAge])
		if math.Abs(u/WonderPriceUnits-1) > 0.05 {
			t.Errorf("%s costs %.1f price units of %s, want %.0f", d.Key, u, d.RequiredAge, WonderPriceUnits)
		}
	}
}

// TestDescriptionRatesMatchEffects: the rate a description shows is the rate
// the building has, since the build list shows the description.
func TestDescriptionRatesMatchEffects(t *testing.T) {
	for _, d := range BaseBuildings() {
		for _, e := range d.Effects {
			if e.Type != "production" || e.Value <= 0 {
				continue
			}
			re := regexp.MustCompile(`\+([0-9][0-9.,]*[KMBTQ]?)\s+` + regexp.QuoteMeta(e.Target) + `/tick`)
			m := re.FindStringSubmatch(d.Description)
			if m == nil {
				continue
			}
			if want := FormatRateValue(e.Value); m[1] != want {
				t.Errorf("%s description says +%s %s/tick, effect is %s", d.Key, m[1], e.Target, want)
			}
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
