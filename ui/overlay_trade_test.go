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
		if !strings.Contains(out, "Faction standing and deals: type factions") {
			t.Errorf("%s: trade overlay is missing the pointer to the Factions panel:\n%s", name, out)
		}
		// Faction rows belong to the Factions panel, not here.
		for _, gone := range []string{"Riverlands Tribes", "Ironhold Clans", "Undiscovered", "diplomacy ally"} {
			if strings.Contains(out, gone) {
				t.Errorf("%s: trade overlay still shows faction detail %q", name, gone)
			}
		}
		// The Trade sections themselves still render.
		for _, want := range []string{"Exchange Rates", "Trade Routes"} {
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

	if !strings.Contains(out, "Allied Bonuses") {
		t.Fatalf("an ally with a trade bonus should get an Allied Bonuses block:\n%s", out)
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

	if out := tradeProvider(game.GameState{}, panelWidth); strings.Contains(out, "Allied Bonuses") {
		t.Errorf("no allies, yet the Allied Bonuses block renders:\n%s", out)
	}
}
