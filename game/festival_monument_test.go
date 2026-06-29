package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// ---------------------------------------------------------------------------
// Cultural Monuments — culture sink via construction cost + permanent bonus
// ---------------------------------------------------------------------------

// TestMonuments_ExistAndCostCulture asserts the four cultural monuments are real
// buildable defs, each priced in culture (the sink) with a permanent
// production_all "bonus" effect (the payoff), MaxCount 1, in the "monument"
// category.
func TestMonuments_ExistAndCostCulture(t *testing.T) {
	byKey := config.BuildingByKey()
	want := map[string]float64{
		"cultural_obelisk":            0.01,
		"grand_amphitheatre_monument": 0.02,
		"eternal_library_monument":    0.03,
		"monument_of_ages":            0.05,
	}
	for key, wantBonus := range want {
		b, ok := byKey[key]
		if !ok {
			t.Fatalf("monument %q missing from BuildingByKey", key)
		}
		if b.Category != "monument" {
			t.Errorf("monument %q category = %q, want \"monument\"", key, b.Category)
		}
		if b.MaxCount != 1 {
			t.Errorf("monument %q MaxCount = %d, want 1", key, b.MaxCount)
		}
		if got := b.BaseCost["culture"]; got <= 0 {
			t.Errorf("monument %q has no culture cost (BaseCost[culture]=%v) — it is supposed to be a culture sink", key, got)
		}
		// permanent production_all bonus effect present with the expected value.
		var bonus float64
		for _, eff := range b.Effects {
			if eff.Type == "bonus" && eff.Target == "production_all" {
				bonus = eff.Value
			}
		}
		if math.Abs(bonus-wantBonus) > 1e-9 {
			t.Errorf("monument %q production_all bonus = %v, want %v", key, bonus, wantBonus)
		}
	}
}

// TestMonuments_GrantPermanentProductionBonus proves a built monument feeds the
// production_all multiplier through the same static-bonus path wonders use:
// getWonderBonuses (now wonders+monuments) and the resolver both reflect it.
func TestMonuments_GrantPermanentProductionBonus(t *testing.T) {
	ge := NewGameEngine()

	// Register a monument def (mirrors config) and mark one built.
	ge.Buildings.defs["monument_of_ages"] = config.BuildingDef{
		Key:      "monument_of_ages",
		Name:     "Monument of Ages",
		Category: "monument",
		Effects:  []config.Effect{{Type: "bonus", Target: "production_all", Value: 0.05}},
	}

	if got := ge.getWonderBonuses()["production_all"]; got != 0 {
		t.Fatalf("baseline static production_all bonus = %v, want 0 (nothing built yet)", got)
	}

	ge.Buildings.counts["monument_of_ages"] = 1

	if got := ge.getWonderBonuses()["production_all"]; math.Abs(got-0.05) > 1e-9 {
		t.Errorf("getWonderBonuses[production_all] with monument built = %v, want 0.05", got)
	}
	// And it must surface in the resolver pool the engine multiplies rates by.
	if got := ge.buildResolver().AddTotal("production_all"); math.Abs(got-0.05) > 1e-9 {
		t.Errorf("resolver production_all with monument built = %v, want 0.05", got)
	}
}

// ---------------------------------------------------------------------------
// Festival — repeatable, player-initiated culture sink (temporary buff)
// ---------------------------------------------------------------------------

// TestFestival_RefusesWhenCultureInsufficient asserts DoFestival errors and
// spends nothing / injects nothing when the player can't afford it.
func TestFestival_RefusesWhenCultureInsufficient(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("culture")
	// Set culture below the minimum cost.
	ge.Resources.LoadAmounts(map[string]float64{"culture": festivalMinCost - 1})

	before := ge.Resources.Get("culture")
	if err := ge.DoFestival(); err == nil {
		t.Fatal("DoFestival should fail when culture is insufficient")
	}
	if got := ge.Resources.Get("culture"); got != before {
		t.Errorf("culture changed on a failed festival: %v -> %v (must spend nothing)", before, got)
	}
	if n := len(ge.Events.GetActive()); n != 0 {
		t.Errorf("a failed festival injected %d active events, want 0", n)
	}
}

// TestFestival_DeductsCultureAndInjectsBuff is the happy path: enough culture →
// culture is spent, a temporary production_all buff event is injected, and the
// cooldown is armed.
func TestFestival_DeductsCultureAndInjectsBuff(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("culture")
	// Give plenty of culture so cost = festivalMinCost (cap is small early).
	ge.Resources.LoadAmounts(map[string]float64{"culture": 1_000_000})

	cost := ge.festivalCost()
	before := ge.Resources.Get("culture")

	if err := ge.DoFestival(); err != nil {
		t.Fatalf("DoFestival failed with ample culture: %v", err)
	}

	// Culture actually consumed.
	spent := before - ge.Resources.Get("culture")
	if math.Abs(spent-cost) > 1e-9 {
		t.Errorf("festival spent %v culture, want %v", spent, cost)
	}

	// Buff event injected with the right key, duration, and production_all value.
	active := ge.Events.GetActive()
	var found *ActiveEventState
	for i := range active {
		if active[i].Key == "cultural_festival" {
			found = &active[i]
			break
		}
	}
	if found == nil {
		t.Fatal("festival did not inject a cultural_festival active event")
	}
	if found.TicksLeft != festivalBuffTicks {
		t.Errorf("festival buff TicksLeft = %d, want %d", found.TicksLeft, festivalBuffTicks)
	}
	var buf float64
	for _, eff := range found.Effects {
		if eff.Type == "production_all" {
			buf = eff.Value
		}
	}
	if math.Abs(buf-festivalBuffPercent) > 1e-9 {
		t.Errorf("festival production_all = %v, want %v", buf, festivalBuffPercent)
	}

	// The buff must reach the resolver pool the engine applies to rates.
	if got := ge.buildResolver().AddTotal("production_all"); math.Abs(got-festivalBuffPercent) > 1e-9 {
		t.Errorf("resolver production_all during festival = %v, want %v", got, festivalBuffPercent)
	}
}

// TestFestival_RespectsCooldown asserts a second festival is refused while the
// cooldown is active, and allowed again once enough ticks have passed.
func TestFestival_RespectsCooldown(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("culture")
	ge.Resources.LoadAmounts(map[string]float64{"culture": 1_000_000})

	if err := ge.DoFestival(); err != nil {
		t.Fatalf("first festival failed: %v", err)
	}
	// Immediately: on cooldown.
	st := ge.FestivalStatus()
	if st.Ready {
		t.Errorf("festival reported Ready immediately after holding one; CooldownLeft=%d", st.CooldownLeft)
	}
	if err := ge.DoFestival(); err == nil {
		t.Error("second festival during cooldown should fail")
	}

	// Advance the game clock past the cooldown window.
	ge.tick += festivalCooldownTicks
	st = ge.FestivalStatus()
	if !st.Ready {
		t.Errorf("festival not Ready after cooldown elapsed; CooldownLeft=%d", st.CooldownLeft)
	}
	if err := ge.DoFestival(); err != nil {
		t.Errorf("festival after cooldown elapsed should succeed, got: %v", err)
	}
}

// TestFestivalStatus_ReportsCostAndBuff sanity-checks the snapshot the command
// renders: cost ≥ minimum, buff params match the constants.
func TestFestivalStatus_ReportsCostAndBuff(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("culture")
	st := ge.FestivalStatus()
	if st.Cost < festivalMinCost {
		t.Errorf("festival cost %v below minimum %v", st.Cost, festivalMinCost)
	}
	if st.BuffPercent != festivalBuffPercent {
		t.Errorf("status BuffPercent = %v, want %v", st.BuffPercent, festivalBuffPercent)
	}
	if st.BuffTicks != festivalBuffTicks {
		t.Errorf("status BuffTicks = %d, want %d", st.BuffTicks, festivalBuffTicks)
	}
	if st.CooldownTicks != festivalCooldownTicks {
		t.Errorf("status CooldownTicks = %d, want %d", st.CooldownTicks, festivalCooldownTicks)
	}
}
