package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// tradeState is a mid-game state: two routes running (one blockaded), one
// open, one that waits for a building, and an ally that boosts trade.
func tradeState() game.GameState {
	s := game.GameState{Age: "medieval_age"}
	s.Trade.ActiveRoutes = []game.ActiveRouteInfo{
		{Name: "Local Barter", Key: "local_barter", TicksLeft: 30, CycleTicks: 50, CyclesDone: 14,
			Export: map[string]float64{"food": 10}, Import: map[string]float64{"wood": 8}},
		{Name: "Iron Road", Key: "iron_road", Disrupted: true, DisruptedBy: "iron",
			Export: map[string]float64{"stone": 40}, Import: map[string]float64{"iron": 12}},
	}
	s.Trade.AvailableRoutes = []game.TradeRouteInfo{
		{Name: "Stone Trade", Key: "stone_trade", CanStart: true, Description: "Send wood to the quarries for stone.",
			Export: map[string]float64{"wood": 15}, Import: map[string]float64{"stone": 12}},
		{Name: "Crypto Market", Key: "crypto_market", RequiredBld: "black_market", MinCount: 1,
			Export: map[string]float64{"crypto": 50}, Import: map[string]float64{"gold": 1000}},
	}
	s.Trade.DisruptedResources = []string{"iron"}
	s.Diplomacy.Factions = map[string]game.FactionInfo{
		"merchant_guild": {Name: "Merchant Guild", Specialty: "gold", Discovered: true, Status: "allied", TradeBonus: 0.20},
	}
	return s
}

// TestTradePage_Cards: a mid-game state gives a card to every route, each
// with the state the old page showed, and the commands the old page named.
func TestTradePage_Cards(t *testing.T) {
	pg := tradePage(tradeState())
	run := deckSectionNamed(t, pg, "Running")
	open := deckSectionNamed(t, pg, "Open routes")
	if len(run.Cards) != 2 || len(open.Cards) != 2 {
		t.Fatalf("want 2 running and 2 open cards:\n%s", deckText(pg))
	}
	out := deckText(pg)
	wantAll(t, "trade page", out,
		"2 running", "1 open", "1 locked", // counts
		"⚠ Routes importing iron are suspended (a war, or your embargo).",
		"Lift your embargo with: diplomacy neutral <civ>",
		"Export 10 food", "Import 8 wood", "Suspended: iron imports are blockaded",
		"left in this cycle", "14 cycles done",
		"Send wood to the quarries for stone.", "Export 15 wood", "Import 12 stone",
		"needs 1 Black Market Hub (have 0)",
		"Commands: trade route start/stop <key>")
	barter, _ := deckCardNamed(t, pg, "Local Barter")
	if !barter.Meter || barter.Share < 0.39 || barter.Share > 0.41 || barter.Command != "trade route stop local_barter" {
		t.Errorf("Local Barter = %+v", barter)
	}
	if road, _ := deckCardNamed(t, pg, "Iron Road"); road.Meter || road.Why == "" || road.Command != "trade route stop iron_road" {
		t.Errorf("Iron Road = %+v", road)
	}
	if c, _ := deckCardNamed(t, pg, "Stone Trade"); c.Command != "trade route start stone_trade" {
		t.Errorf("Stone Trade command = %q", c.Command)
	}
	if c, _ := deckCardNamed(t, pg, "Crypto Market"); c.Command != "" || !c.Quiet {
		t.Errorf("a locked route should only show: %+v", c)
	}
	for _, gone := range []string{"black_market", "given plus received"} {
		if strings.Contains(out, gone) {
			t.Errorf("trade page shows %q:\n%s", gone, out)
		}
	}
}

// TestTradePage_RunsItsCommands: Enter on a card runs what the old page told
// the player to type.
func TestTradePage_RunsItsCommands(t *testing.T) {
	for title, cmd := range map[string]string{
		"Local Barter": "trade route stop local_barter",
		"Stone Trade":  "trade route start stone_trade",
		"Factions":     "factions",
	} {
		wantDeckEnter(t, tradePage, tradeState(), title, cmd)
	}
}

func TestTradePage_Fits(t *testing.T) {
	wantDeckFits(t, tradePage, tradeState())
	wantDeckFits(t, tradePage, game.GameState{})
}

// TestTradePage_NoDiplomacySection: faction standing lives on the Factions
// panel, and the Trade panel only points there.
func TestTradePage_NoDiplomacySection(t *testing.T) {
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
		"riverlands_tribes": {Name: "Riverlands Tribes", Specialty: "food", Discovered: true, Status: "friendly", Opinion: 30},
		"ironhold_clans":    {Name: "Ironhold Clans", Specialty: "iron", Discovered: false},
	}}}
	for name, s := range map[string]game.GameState{"zero": {}, "met": state} {
		out := deckText(tradePage(s))
		wantAll(t, name, out, "Opinion and deals with other civilizations: type factions", "No trade routes in this age yet.")
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
