package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// setRes ensures a resource exists with plenty of headroom and the given amount,
// so trade-route imports/exports aren't clipped by storage in tests.
func setRes(ge *GameEngine, key string, amount float64) {
	r, ok := ge.Resources.resources[key]
	if !ok {
		return
	}
	r.Storage = amount + 1e9
	r.Amount = amount
}

// --- New colonial→industrial routes exist & are valid ----------------------

// TestTradeExpansion_NewRoutesPresent confirms the six gap-filler routes were
// added (total now 20+) and that each new route is age-gated in the
// renaissance→industrial band with sane export/import maps and a real building.
func TestTradeExpansion_NewRoutesPresent(t *testing.T) {
	routes := config.TradeRouteByKey()
	bldKeys := config.BuildingByKey()
	ageOrder := fullAgeOrder()

	newKeys := []string{
		"mercantile_convoy", "triangular_trade", "tea_clippers",
		"coal_barges", "cotton_exchange", "steamship_line",
	}
	for _, k := range newKeys {
		def, ok := routes[k]
		if !ok {
			t.Errorf("new route %q missing from BaseTradeRoutes", k)
			continue
		}
		// Age gate must sit in the renaissance→industrial band.
		o := ageOrder[def.MinAge]
		if o < ageOrder["renaissance_age"] || o > ageOrder["industrial_age"] {
			t.Errorf("route %q MinAge %q outside renaissance→industrial band", k, def.MinAge)
		}
		if def.RequiredBld != "" {
			if _, ok := bldKeys[def.RequiredBld]; !ok {
				t.Errorf("route %q RequiredBld %q is not a real building", k, def.RequiredBld)
			}
		}
		if len(def.Export) == 0 || len(def.Import) == 0 {
			t.Errorf("route %q must have both exports and imports", k)
		}
		if def.TicksPerRun <= 0 {
			t.Errorf("route %q TicksPerRun = %d, want > 0", k, def.TicksPerRun)
		}
	}

	if len(routes) < 20 {
		t.Errorf("total trade routes = %d, want 20+ after expansion", len(routes))
	}
}

// --- Harbor lineage ---------------------------------------------------------

// TestHarborLineage_Exists verifies the 5-tier harbour lineage is registered,
// gated colonial→digital, and each tier carries a positive trade_route_income
// effect.
func TestHarborLineage_Exists(t *testing.T) {
	byKey := config.BuildingByKey()
	want := []struct {
		key string
		age string
	}{
		{"harbor", "colonial_age"},
		{"harbor_authority", "industrial_age"},
		{"seaport", "modern_age"},
		{"container_terminal", "information_age"},
		{"logistics_hub", "digital_age"},
	}
	for _, w := range want {
		def, ok := byKey[w.key]
		if !ok {
			t.Errorf("harbour building %q missing", w.key)
			continue
		}
		if def.RequiredAge != w.age {
			t.Errorf("%s RequiredAge = %q, want %q", w.key, def.RequiredAge, w.age)
		}
		if def.WorkerDomain != "trade" {
			t.Errorf("%s WorkerDomain = %q, want trade", w.key, def.WorkerDomain)
		}
		var income float64
		for _, eff := range def.Effects {
			if eff.Type == "trade_route_income" {
				income = eff.Value
			}
		}
		if income <= 0 {
			t.Errorf("%s has no positive trade_route_income effect", w.key)
		}
	}
}

// TestHarborRouteBonus_RaisesIncome is the load-bearing harbour test: with a
// harbour built, a route's imports must pay out MORE than the same route with no
// harbour. Drives the real engine income path (processTrade → Trade.Tick).
func TestHarborRouteBonus_RaisesIncome(t *testing.T) {
	run := func(withHarbor bool) float64 {
		ge := NewGameEngine()
		ge.age = "industrial_age"
		// Provide a market for the route requirement + the simplest route.
		ge.Buildings.counts["market"] = 2
		setRes(ge, "food", 1e6)
		setRes(ge, "wood", 1e6)
		setRes(ge, "gold", 0)
		if withHarbor {
			ge.Buildings.counts["harbor"] = 1 // +5% route income
		}
		// Start a deterministic route: stone_trade (export wood, import stone)?
		// Use local_barter (food→wood) so we measure wood gain cleanly.
		if err := ge.Trade.StartRoute("local_barter", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
			t.Fatalf("StartRoute: %v", err)
		}
		before := ge.Resources.Get("wood")
		// Tick until the route completes exactly one cycle.
		def := config.TradeRouteByKey()["local_barter"]
		for i := 0; i < def.TicksPerRun; i++ {
			ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, ge.harborRouteBonus())
		}
		return ge.Resources.Get("wood") - before
	}

	base := run(false)
	boosted := run(true)
	if base <= 0 {
		t.Fatalf("baseline route produced no import (got %.3f)", base)
	}
	if boosted <= base {
		t.Errorf("harbour bonus did not raise income: base=%.3f boosted=%.3f", base, boosted)
	}
	// local_barter imports 8 wood; +5% harbour → 8.4.
	if math.Abs(boosted-8.4) > 1e-6 {
		t.Errorf("boosted import = %.3f, want 8.4 (8 × 1.05)", boosted)
	}
}

// --- Disruption tied to diplomacy war/embargo -------------------------------

// hostileFaction marks a faction at war (or embargoed) directly so its specialty
// resource becomes disrupted.
func hostileFaction(ge *GameEngine, key string, atWar bool) {
	fs := &FactionState{Discovered: true, Opinion: -80}
	if atWar {
		fs.AtWar = true
		fs.Status = "rival"
	} else {
		fs.Status = "embargo"
	}
	ge.Diplomacy.factions[key] = fs
}

// TestDisruption_BlocksRouteWhenAtWar verifies a route importing a hostile civ's
// specialty resource earns nothing while at war, then resumes once peace returns.
// Uses silk_road (imports culture) against an embargoed/at-war culture specialist
// (artisan_league, specialty "culture").
func TestDisruption_BlocksRouteWhenAtWar(t *testing.T) {
	setup := func() *GameEngine {
		ge := NewGameEngine()
		ge.age = "medieval_age"
		ge.Buildings.counts["market"] = 3
		setRes(ge, "gold", 1e6)
		setRes(ge, "culture", 0)
		if err := ge.Trade.StartRoute("silk_road", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
			t.Fatalf("StartRoute: %v", err)
		}
		return ge
	}

	// Sanity: artisan_league specialises in culture (the disrupted resource).
	if config.FactionByKey()["artisan_league"].Specialty != "culture" {
		t.Fatalf("test assumes artisan_league specialty culture")
	}

	def := config.TradeRouteByKey()["silk_road"]

	// 1) At war → route is disrupted, no culture gained over a full cycle.
	ge := setup()
	hostileFaction(ge, "artisan_league", true)
	for i := 0; i < def.TicksPerRun; i++ {
		ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, 0)
	}
	if got := ge.Resources.Get("culture"); got != 0 {
		t.Errorf("at war: culture gained = %.1f, want 0 (route should be disrupted)", got)
	}
	// Disrupted flag should be reflected on the active route + state snapshot.
	disrupted := ge.Diplomacy.DisruptedResources()
	if !disrupted["culture"] {
		t.Errorf("DisruptedResources should include culture when at war with culture specialist")
	}

	// 2) Embargo (no war) also disrupts.
	ge = setup()
	hostileFaction(ge, "artisan_league", false)
	for i := 0; i < def.TicksPerRun; i++ {
		ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, 0)
	}
	if got := ge.Resources.Get("culture"); got != 0 {
		t.Errorf("embargo: culture gained = %.1f, want 0", got)
	}

	// 3) Peace → route runs normally and culture flows.
	ge = setup()
	for i := 0; i < def.TicksPerRun; i++ {
		ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, 0)
	}
	if got := ge.Resources.Get("culture"); got <= 0 {
		t.Errorf("peace: culture gained = %.1f, want > 0 (route should run)", got)
	}
}
