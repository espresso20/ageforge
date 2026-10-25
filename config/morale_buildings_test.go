package config

import "testing"

// moraleEffectValue returns the morale effect value for a building def, and
// whether it has a morale effect at all.
func moraleEffectValue(b BuildingDef) (float64, bool) {
	for _, eff := range b.Effects {
		if eff.Type == "morale" {
			return eff.Value, true
		}
	}
	return 0, false
}

// TestMoraleBuildings_OnlySpiritLineagesCarryMorale: every faith and
// culture_arts building has a positive morale effect, and no other lineage
// does (a pure-producer lineage must not turn into a morale building).
func TestMoraleBuildings_OnlySpiritLineagesCarryMorale(t *testing.T) {
	spirit := 0
	for _, b := range BaseBuildings() {
		v, has := moraleEffectValue(b)
		if b.LineageKey != "faith" && b.LineageKey != "culture_arts" {
			if has {
				t.Errorf("building %q (lineage %q) unexpectedly has a morale effect; only faith/culture_arts should",
					b.Key, b.LineageKey)
			}
			continue
		}
		spirit++
		if !has {
			t.Errorf("%s building %q (tier %d) is missing a morale effect", b.LineageKey, b.Key, b.LineageTier)
		} else if v <= 0 {
			t.Errorf("%s building %q has non-positive morale value %v", b.LineageKey, b.Key, v)
		}
	}
	if spirit == 0 {
		t.Error("no faith or culture_arts buildings found")
	}
}

// TestMoraleBuildings_RampTrendsUpByTier checks the gentle-ramp intent: later,
// costlier buildings give MORE morale. The ramp is hand-rounded off
// ~0.0005·1.15^tier, and faith's shrine (t0) + temple (t3) keep their
// grandfathered Stage-1 value of 0.0006 (so a couple of local dips are expected
// and fine). We assert the overall trend, not strict step-by-step monotonicity:
// the top tier must hold the lineage's max value, and that max must meaningfully
// exceed the bottom tier.
func TestMoraleBuildings_RampTrendsUpByTier(t *testing.T) {
	for _, lineage := range []string{"faith", "culture_arts"} {
		var bottomVal, topVal, maxVal float64
		bottomTier, topTier := 1<<30, -1
		seen := false
		for _, b := range BaseBuildings() {
			if b.LineageKey != lineage {
				continue
			}
			v, ok := moraleEffectValue(b)
			if !ok {
				continue
			}
			seen = true
			if v > maxVal {
				maxVal = v
			}
			if b.LineageTier < bottomTier {
				bottomTier, bottomVal = b.LineageTier, v
			}
			if b.LineageTier > topTier {
				topTier, topVal = b.LineageTier, v
			}
		}
		if !seen {
			t.Fatalf("%s: no morale buildings found", lineage)
		}
		// Top tier should carry the lineage's largest morale value.
		if topVal < maxVal {
			t.Errorf("%s: top tier %d value %v is below lineage max %v", lineage, topTier, topVal, maxVal)
		}
		// And the top should be clearly larger than the bottom (ramp, not flat).
		if topVal <= bottomVal*2 {
			t.Errorf("%s: top tier %d=%v not meaningfully above bottom tier %d=%v",
				lineage, topTier, topVal, bottomTier, bottomVal)
		}
	}
}
