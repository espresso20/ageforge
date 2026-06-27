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

// TestMoraleBuildings_FaithAndCultureCarryMoraleEffect asserts that the
// Stage 2A broadening landed: every faith worship building and every
// culture_arts entertainment building parses with a positive morale effect,
// and that BuildingByKey surfaces a couple of spot-check keys with it.
func TestMoraleBuildings_FaithAndCultureCarryMoraleEffect(t *testing.T) {
	byKey := BuildingByKey()

	faithCount := 0
	cultureCount := 0
	for _, b := range BaseBuildings() {
		switch b.LineageKey {
		case "faith":
			v, ok := moraleEffectValue(b)
			if !ok {
				t.Errorf("faith building %q (tier %d) is missing a morale effect", b.Key, b.LineageTier)
				continue
			}
			if v <= 0 {
				t.Errorf("faith building %q has non-positive morale value %v", b.Key, v)
			}
			faithCount++
		case "culture_arts":
			v, ok := moraleEffectValue(b)
			if !ok {
				t.Errorf("culture_arts building %q (tier %d) is missing a morale effect", b.Key, b.LineageTier)
				continue
			}
			if v <= 0 {
				t.Errorf("culture_arts building %q has non-positive morale value %v", b.Key, v)
			}
			cultureCount++
		}
	}

	// Sanity: the lineages are non-trivial in size (guards against an empty
	// iteration silently passing).
	if faithCount < 20 {
		t.Errorf("expected at least 20 faith buildings with morale effects, got %d", faithCount)
	}
	if cultureCount < 16 {
		t.Errorf("expected at least 16 culture_arts buildings with morale effects, got %d", cultureCount)
	}

	// Spot-check via BuildingByKey: shrine (pre-existing) and a Stage-2A addition
	// in each lineage must resolve with a morale effect.
	for _, key := range []string{"shrine", "standing_stones", "amphitheater"} {
		b, ok := byKey[key]
		if !ok {
			t.Fatalf("BuildingByKey missing %q", key)
		}
		if _, has := moraleEffectValue(b); !has {
			t.Errorf("BuildingByKey[%q] has no morale effect", key)
		}
	}
}

// TestMoraleBuildings_OnlySpiritLineagesGotMorale guards against accidentally
// turning a pure-producer lineage into a morale building. Only faith and
// culture_arts buildings should carry morale effects.
func TestMoraleBuildings_OnlySpiritLineagesGotMorale(t *testing.T) {
	for _, b := range BaseBuildings() {
		if _, has := moraleEffectValue(b); !has {
			continue
		}
		if b.LineageKey != "faith" && b.LineageKey != "culture_arts" {
			t.Errorf("building %q (lineage %q) unexpectedly has a morale effect — "+
				"only faith/culture_arts should", b.Key, b.LineageKey)
		}
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
