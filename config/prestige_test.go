package config

import "testing"

// TestDepthWeights: every completed age pays 3^epoch, with no divisor: 1 per
// Stone Era age up to 729 per Cosmic Era age, 4,008 over all 22 ages.
func TestDepthWeights(t *testing.T) {
	want := map[string]int{
		"stone_era": 1, "iron_era": 3, "steel_era": 9, "electric_era": 27,
		"digital_era": 81, "neon_era": 243, "cosmic_era": 729,
	}
	total := 0
	for _, a := range AgeOrder() {
		w := DepthWeight(a)
		if w != want[EpochForAge(a)] {
			t.Errorf("%s (%s) weighs %d, want %d", a, EpochForAge(a), w, want[EpochForAge(a)])
		}
		total += w
	}
	if total != 4008 {
		t.Errorf("the ages weigh %d in all, want 4,008", total)
	}
	if DepthWeight("not_an_age") != 0 || DepthPoints("not_an_age") != 0 {
		t.Error("an unknown age pays something")
	}
}

// TestDepthPoints: the plan's table, and a second Medieval Age prestige
// after a Modern Age run gains under 10% (9 of 120).
func TestDepthPoints(t *testing.T) {
	for age, want := range map[string]int{
		"primitive_age": 0, "medieval_age": 9, "modern_age": 120, "information_age": 201,
		"cyberpunk_age": 363, "interstellar_age": 1092,
	} {
		if got := DepthPoints(age); got != want {
			t.Errorf("DepthPoints(%s) = %d, want %d", age, got, want)
		}
	}
	if m, r := DepthPoints("medieval_age"), DepthPoints("modern_age"); m*10 >= r {
		t.Errorf("a Medieval Age prestige pays %d, %.1f%% of a Modern Age run's %d; want under 10%%", m, float64(m*100)/float64(r), r)
	}
}

// TestPrestigeShop: nine retired perks (keys kept, costs frozen) and the
// four-item legacy kit at 9, 18, 36 and 54 points (117 in all).
func TestPrestigeShop(t *testing.T) {
	retired, kit := 0, 0
	prices := map[string]int{}
	for _, u := range PrestigeUpgrades() {
		if len(u.Costs) != u.MaxTier {
			t.Errorf("%s: %d costs for %d tiers", u.Key, len(u.Costs), u.MaxTier)
		}
		switch {
		case u.Retired:
			retired++
		case u.EffectType == "legacy":
			kit++
			prices[u.Key] = u.Costs[0]
			if u.MaxTier != 1 {
				t.Errorf("%s has %d tiers, want 1", u.Key, u.MaxTier)
			}
		default:
			t.Errorf("%s is neither retired nor a kit item", u.Key)
		}
	}
	if retired != 9 || kit != 4 || len(ActivePrestigeUpgrades()) != 4 {
		t.Errorf("%d retired, %d kit items, %d active; want 9, 4, 4", retired, kit, len(ActivePrestigeUpgrades()))
	}
	want := map[string]int{LegacyPlan: 9, LegacyResearch: 18, LegacyWorkers: 36, LegacyFactions: 54}
	sum := 0
	for _, k := range LegacyKit() {
		if prices[k] != want[k] {
			t.Errorf("%s costs %d, want %d", k, prices[k], want[k])
		}
		sum += prices[k]
	}
	if sum != 117 {
		t.Errorf("the kit costs %d in all, want 117", sum)
	}
}
