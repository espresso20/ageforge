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
	r.Amount = amount
}

// --- Harbor lineage ---------------------------------------------------------

// importFlow is what a route's import adds to a resource's rate, read from the
// rate breakdown after a recalculation.
func importFlow(ge *GameEngine, res string) float64 {
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	ge.recalculateRates()
	return ge.Resources.resources[res].Breakdown.TradeFlow
}

// TestHarborRouteBonus_RaisesIncome is the load-bearing harbour test: with a
// harbour built, a route's import flow is larger than the same route's with no
// harbour, by the harbour's share.
func TestHarborRouteBonus_RaisesIncome(t *testing.T) {
	run := func(withHarbor bool) float64 {
		ge := NewGameEngine()
		ge.age = "industrial_age"
		ge.Buildings.counts["market"] = 2
		setRes(ge, "food", config.Million)
		setRes(ge, "wood", config.Million)
		if withHarbor {
			ge.Buildings.counts["harbor"] = 1 // +5% route income
		}
		if err := ge.Trade.StartRoute("local_barter", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
			t.Fatalf("StartRoute: %v", err)
		}
		return importFlow(ge, "wood")
	}

	base := run(false)
	boosted := run(true)
	if base <= 0 {
		t.Fatalf("baseline route brought no import (got %.6f)", base)
	}
	if math.Abs(boosted/base-1.05) > 1e-9 {
		t.Errorf("harbour flow = %.6f against %.6f without, want x1.05", boosted, base)
	}
}

// TestRouteCycle_RaisesOpinion: each minute a route runs gives +1 opinion to
// every civ you have met, except civs at war with you (trade.md, "Trade
// routes"). A route that is not running gives nothing.
func TestRouteCycle_RaisesOpinion(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "medieval_age"
	ge.Buildings.counts["market"] = 3
	setRes(ge, "gold", config.Million)
	ge.Diplomacy.factions["riverlands_tribes"] = &FactionState{Discovered: true, Opinion: 10, Status: "neutral"}
	ge.Diplomacy.factions["ironhold_clans"] = &FactionState{Discovered: true, Opinion: -60, Status: "rival", AtWar: true}
	ge.Diplomacy.factions["artisan_league"] = &FactionState{Discovered: false, Opinion: 0, Status: "neutral"}
	if err := ge.Trade.StartRoute("silk_road", ge.Buildings, ge.age, ge.progress.GetAgeOrder()); err != nil {
		t.Fatalf("StartRoute: %v", err)
	}
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	ge.recalculateRates()
	minute := func() {
		for i := 1; i <= minutesToTicks(1); i++ {
			ge.Trade.Tick(i, ge.Diplomacy, nil)
		}
	}
	minute()
	f := ge.Diplomacy.factions
	if got := f["riverlands_tribes"].Opinion; got != 11 {
		t.Errorf("met civ opinion after one minute = %d, want 11", got)
	}
	if got := f["riverlands_tribes"].TradeCount; got != 1 {
		t.Errorf("met civ trade count after one minute = %d, want 1", got)
	}
	if got := f["ironhold_clans"].Opinion; got != -60 {
		t.Errorf("civ at war: opinion = %d, want -60 (unchanged)", got)
	}
	if got := f["artisan_league"].Opinion; got != 0 {
		t.Errorf("unmet civ: opinion = %d, want 0 (unchanged)", got)
	}

	// A paused route is held and warms no one.
	if err := ge.Trade.StopRoute("silk_road"); err != nil {
		t.Fatal(err)
	}
	ge.recalculateRates()
	minute()
	if got := f["riverlands_tribes"].Opinion; got != 11 {
		t.Errorf("after a paused minute opinion = %d, want 11", got)
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
// specialty resource brings nothing while at war, then resumes once peace returns.
// Uses silk_road (imports culture) against an embargoed/at-war culture specialist
// (artisan_league, specialty "culture").
func TestDisruption_BlocksRouteWhenAtWar(t *testing.T) {
	setup := func() *GameEngine {
		ge := NewGameEngine()
		ge.age = "medieval_age"
		ge.Buildings.counts["market"] = 3
		setRes(ge, "gold", config.Million)
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

	// 1) At war → route is disrupted, no culture gained over a full cycle.
	ge := setup()
	hostileFaction(ge, "artisan_league", true)
	if got := importFlow(ge, "culture"); got != 0 {
		t.Errorf("at war: culture flow = %.6f, want 0 (route should be disrupted)", got)
	}
	// Disrupted flag should be reflected on the active route + state snapshot.
	disrupted := ge.Diplomacy.DisruptedResources()
	if !disrupted["culture"] {
		t.Errorf("DisruptedResources should include culture when at war with culture specialist")
	}

	// 2) Embargo (no war) also disrupts.
	ge = setup()
	hostileFaction(ge, "artisan_league", false)
	if got := importFlow(ge, "culture"); got != 0 {
		t.Errorf("embargo: culture flow = %.6f, want 0", got)
	}

	// 3) Peace → route runs normally and culture flows.
	ge = setup()
	if got := importFlow(ge, "culture"); got <= 0 {
		t.Errorf("peace: culture flow = %.6f, want > 0 (route should run)", got)
	}
}
