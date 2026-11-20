package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// tradeState is a mid-game state: two routes running (one blockaded), one
// open, one that waits for a building, and an ally that boosts trade.
func tradeState() game.GameState {
	s := game.GameState{Age: "medieval_age"}
	s.Trade.ActiveRoutes = []game.ActiveRouteInfo{
		{Name: "Local Barter", Key: "local_barter", Scale: 1, Run: 1,
			Export: map[string]float64{"food": 10}, Import: map[string]float64{"wood": 8}},
		{Name: "Gold Caravan", Key: "gold_caravan", Disrupted: true, DisruptedBy: "gold",
			Export: map[string]float64{"stone": 50}, Import: map[string]float64{"gold": 5}},
	}
	s.Trade.AvailableRoutes = []game.TradeRouteInfo{
		{Name: "Stone Trade", Key: "stone_trade", CanStart: true, Description: "Send wood to the quarries for stone.",
			Export: map[string]float64{"wood": 15}, Import: map[string]float64{"stone": 12}},
		{Name: "Spice Trade", Key: "spice_trade", RequiredBld: "port", MinCount: 1,
			Export: map[string]float64{"gold": 100}, Import: map[string]float64{"food": 200}},
	}
	s.Trade.DisruptedResources = []string{"gold"}
	s.Diplomacy.Factions = map[string]game.FactionInfo{
		"merchant_guild": {Name: "Merchant Guild", Specialty: "gold", Discovered: true, Status: "allied", TradeBonus: 0.20},
	}
	return s
}

// TestTradePage_Cards: a mid-game state gives a card to every route, in the
// order the board lists them, each with the state and the command it has.
func TestTradePage_Cards(t *testing.T) {
	pg := tradePage(tradeState())
	open := deckSectionNamed(t, pg, "Open routes")
	if len(open.Cards) != 4 {
		t.Fatalf("want 4 route cards:\n%s", deckText(pg))
	}
	out := deckText(pg)
	wantAll(t, "trade page", out,
		"2 running", "1 open", "1 locked", // counts
		"⚠ Routes importing gold are suspended (a war, or your embargo).",
		"Lift your embargo with: diplomacy neutral <civ>",
		"-10 food/s", "+8 wood/s", "Suspended: gold imports are blockaded",
		"running at 100% of its flow", "-15 wood/s", "+12 stone/s",
		"needs 1 Port (have 0)", "trade route start/stop <key>")
	barter, _ := deckCardNamed(t, pg, "Local Barter")
	if !barter.Meter || barter.Share != 1 || barter.Command != "trade route stop local_barter" {
		t.Errorf("Local Barter = %+v", barter)
	}
	if road, _ := deckCardNamed(t, pg, "Gold Caravan"); road.Meter || road.Why == "" || road.Command != "trade route stop gold_caravan" {
		t.Errorf("Gold Caravan = %+v", road)
	}
	if c, _ := deckCardNamed(t, pg, "Stone Trade"); c.Command != "trade route start stone_trade" {
		t.Errorf("Stone Trade command = %q", c.Command)
	}
	if c, _ := deckCardNamed(t, pg, "Spice Trade"); c.Command != "" || !c.Quiet {
		t.Errorf("a locked route should only show: %+v", c)
	}
}

// TestTradePage_RunsItsCommands: Enter on a card runs what the old page told
// the player to type.
func TestTradePage_RunsItsCommands(t *testing.T) {
	for title, cmd := range map[string]string{
		"Local Barter": "trade route stop local_barter",
		"Stone Trade":  "trade route start stone_trade",
	} {
		wantDeckEnter(t, tradePage, tradeState(), title, cmd)
	}
}

func TestTradePage_Fits(t *testing.T) {
	wantDeckFits(t, tradePage, tradeState())
	wantDeckFits(t, tradePage, game.GameState{})
}

// TestTradePage_NoDiplomacySection: faction standing lives on the Factions
// panel; the page names no civilization.
func TestTradePage_NoDiplomacySection(t *testing.T) {
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
		"riverlands_tribes": {Name: "Riverlands Tribes", Specialty: "food", Discovered: true, Status: "friendly", Opinion: 30},
		"ironhold_clans":    {Name: "Ironhold Clans", Specialty: "iron", Discovered: false},
	}}}
	for name, s := range map[string]game.GameState{"zero": {}, "met": state} {
		out := deckText(tradePage(s))
		wantAll(t, name, out, "No trade routes in this age yet.")
		for _, gone := range []string{"Riverlands Tribes", "Ironhold Clans", "Undiscovered", "diplomacy ally", "Market rates", "Traded at the market", "trade <give>"} {
			if strings.Contains(out, gone) {
				t.Errorf("%s: trade page shows %q", name, gone)
			}
		}
	}
}

// TestTradePage_AllyBonuses: the allied trade bonus, the one faction effect
// that changes trade income, is shown, and only for allies that grant it
// (the same gate as GetTradeBonus).
func TestTradePage_AllyBonuses(t *testing.T) {
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
		"merchant_guild":    {Name: "Merchant Guild", Specialty: "gold", Discovered: true, Status: "allied", TradeBonus: 0.20},
		"riverlands_tribes": {Name: "Riverlands Tribes", Specialty: "food", Discovered: true, Status: "friendly", TradeBonus: 0.15},
		"ironhold_clans":    {Name: "Ironhold Clans", Specialty: "iron", Discovered: true, Status: "allied", AtWar: true, TradeBonus: 0.15},
	}}}
	out := deckText(tradePage(state))
	wantAll(t, "ally bonus", out, "Allied bonuses", "Merchant Guild: +20% gold (route imports and production)")
	for _, none := range []string{"Riverlands Tribes", "Ironhold Clans"} {
		if strings.Contains(out, none) {
			t.Errorf("%s grants no trade bonus but is listed:\n%s", none, out)
		}
	}
	if out := deckText(tradePage(game.GameState{})); strings.Contains(out, "Allied bonuses") {
		t.Errorf("no allies, yet Allied bonuses shows:\n%s", out)
	}
}

// TestTradePageLocksRoutesWithTheFeature: while trade routes themselves are
// locked, a route's card has no button: it says what the game would answer.
func TestTradePageLocksRoutesWithTheFeature(t *testing.T) {
	st := game.GameState{
		Age:      "medieval_age",
		Features: map[string]game.FeatureState{config.FeatureTradeRoutes: {Open: false, Tech: "the_wheel", TechName: "The Wheel"}},
	}
	st.Trade.AvailableRoutes = []game.TradeRouteInfo{{Key: "local_barter", Name: "Local Barter", CanStart: true}}
	pg := tradePage(st)
	c, ok := pg.card(0)
	if !ok || c.Command != "" || c.Why == "" || !strings.Contains(c.Why, "The Wheel") {
		t.Errorf("with trade routes locked the card is %+v; want no command and the refusal that names The Wheel", c)
	}
	st.Features[config.FeatureTradeRoutes] = game.FeatureState{Open: true}
	pg = tradePage(st)
	if c, _ := pg.card(0); c.Command != "trade route start local_barter" || c.Why != "" {
		t.Errorf("with trade routes open the card is %+v; want its start command", c)
	}
}
