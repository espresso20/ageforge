package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestTradeProvider_NoDiplomacySection locks the removal of the Trade overlay's
// old Diplomacy section: faction standing lives on the Factions panel, and the
// Trade overlay only points there.
func TestTradeProvider_NoDiplomacySection(t *testing.T) {
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
		"riverlands_tribes": {Name: "Riverlands Tribes", Specialty: "food", Discovered: true, Status: "friendly", Opinion: 30},
		"ironhold_clans":    {Name: "Ironhold Clans", Specialty: "iron", Discovered: false},
	}}}
	for name, s := range map[string]game.GameState{"zero": {}, "met": state} {
		out := tradeProvider(s, panelWidth)
		if strings.Contains(out, "Diplomacy ═") {
			t.Errorf("%s: trade overlay still renders a Diplomacy section:\n%s", name, out)
		}
		if !strings.Contains(out, "Opinion and deals with other civilizations: type factions") {
			t.Errorf("%s: trade overlay is missing the pointer to the Factions panel:\n%s", name, out)
		}
		// Faction rows belong to the Factions panel, not here.
		for _, gone := range []string{"Riverlands Tribes", "Ironhold Clans", "Undiscovered", "diplomacy ally"} {
			if strings.Contains(out, gone) {
				t.Errorf("%s: trade overlay still shows faction detail %q", name, gone)
			}
		}
		// The Trade sections themselves still render.
		for _, want := range []string{"Market rates", "Trade routes"} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: trade overlay missing %q", name, want)
			}
		}
	}
}

// TestTradeProvider_AllyBonuses checks the allied trade bonus, the one faction
// effect that changes trade income, is still shown on the Trade overlay, and
// only for allies that actually grant it (the same gate as GetTradeBonus).
func TestTradeProvider_AllyBonuses(t *testing.T) {
	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{
		"merchant_guild":    {Name: "Merchant Guild", Specialty: "gold", Discovered: true, Status: "allied", TradeBonus: 0.20},
		"riverlands_tribes": {Name: "Riverlands Tribes", Specialty: "food", Discovered: true, Status: "friendly", TradeBonus: 0.15},
		"ironhold_clans":    {Name: "Ironhold Clans", Specialty: "iron", Discovered: true, Status: "allied", AtWar: true, TradeBonus: 0.15},
	}}}
	out := tradeProvider(state, panelWidth)

	if !strings.Contains(out, "Allied bonuses") {
		t.Fatalf("an ally with a trade bonus should get an Allied bonuses block:\n%s", out)
	}
	line := lineContaining(out, "Merchant Guild")
	for _, want := range []string{"+20% gold", "route imports and production"} {
		if !strings.Contains(line, want) {
			t.Errorf("ally bonus line %q missing %q", line, want)
		}
	}
	// A friendly civ grants nothing, and neither does an ally at war.
	for _, none := range []string{"Riverlands Tribes", "Ironhold Clans"} {
		if strings.Contains(out, none) {
			t.Errorf("%s grants no trade bonus but is listed:\n%s", none, out)
		}
	}

	if out := tradeProvider(game.GameState{}, panelWidth); strings.Contains(out, "Allied bonuses") {
		t.Errorf("no allies, yet the Allied bonuses block renders:\n%s", out)
	}
}

// TestTradeProvider_SoldBoughtAndRouteNeeds: market totals are split into
// what you sold and what you bought, and a locked route names the building
// it needs, not its key (black_market is the Black Market Hub).
func TestTradeProvider_SoldBoughtAndRouteNeeds(t *testing.T) {
	var s game.GameState
	s.Trade.TotalSold = map[string]float64{"iron_ore": 1500, "food": 20}
	s.Trade.TotalBought = map[string]float64{"gold": 300}
	s.Trade.AvailableRoutes = []game.TradeRouteInfo{{Name: "Crypto Market", Key: "crypto_market",
		RequiredBld: "black_market", MinCount: 1, Export: map[string]float64{"crypto": 50}, Import: map[string]float64{"gold": 1000}}}
	out := plainText(tradeProvider(s, panelWidth))
	for _, want := range []string{"Sold: 20 food, 1.5K iron ore", "Bought: 300 gold", "needs 1 Black Market Hub (have 0)"} {
		if !strings.Contains(out, want) {
			t.Errorf("trade panel missing %q:\n%s", want, out)
		}
	}
	for _, gone := range []string{"given plus received", "black_market"} {
		if strings.Contains(out, gone) {
			t.Errorf("trade panel still shows %q:\n%s", gone, out)
		}
	}
}
