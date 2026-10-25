package game

import (
	"encoding/json"
	"testing"
)

// TestFactionState_SaveJSON pins how a civ is written into a save: these keys, in this
// order, the deal fields left out when empty, and the transient opinion accumulator never
// written.
func TestFactionState_SaveJSON(t *testing.T) {
	full := FactionState{Discovered: true, Opinion: 33, Status: "friendly", TradeCount: 4, OpinionAccum: 0.7,
		AtWar: true, Provocations: 2, RaidedRoutes: 1, Embargoes: 3, LastProvocationTick: 99,
		Deals:     []FactionDeal{{ID: 1, Kind: "sell", Give: "wood", GiveAmt: 5.5, Taken: true}},
		DealRound: 3, DealTicks: 17, DealsFor: "bronze_age"}
	const wantFull = `{"discovered":true,"opinion":33,"status":"friendly","trade_count":4,` +
		`"at_war":true,"provocations":2,"raided_routes":1,"embargoes":3,"last_provocation_tick":99,` +
		`"deals":[{"id":1,"kind":"sell","give":"wood","give_amt":5.5,"taken":true}],` +
		`"deal_round":3,"deal_ticks":17,"deals_for":"bronze_age"}`
	if b, _ := json.Marshal(full); string(b) != wantFull {
		t.Errorf("full civ:\n got %s\nwant %s", b, wantFull)
	}
	const wantBare = `{"discovered":true,"opinion":0,"status":"neutral","trade_count":0,` +
		`"at_war":false,"provocations":0,"raided_routes":0,"embargoes":0,"last_provocation_tick":0}`
	if b, _ := json.Marshal(FactionState{Discovered: true, Status: "neutral"}); string(b) != wantBare {
		t.Errorf("bare civ:\n got %s\nwant %s", b, wantBare)
	}
}

// TestFactionState_SaveLoadCopies covers that a civ goes into the save and comes back
// without the accumulator, and that neither side shares its deal list with the live state.
func TestFactionState_SaveLoadCopies(t *testing.T) {
	dm := NewDiplomacyManager()
	dm.factions["merchant_guild"] = &FactionState{Discovered: true, Opinion: 10, OpinionAccum: 0.4,
		Deals: []FactionDeal{{ID: 1, Give: "wood", GiveAmt: 2}}}

	saved := dm.GetFactionsForSave()
	if saved["merchant_guild"].OpinionAccum != 0 {
		t.Errorf("save carries the opinion accumulator: %v", saved["merchant_guild"].OpinionAccum)
	}
	saved["merchant_guild"].Deals[0].GiveAmt = 99
	if got := dm.factions["merchant_guild"].Deals[0].GiveAmt; got != 2 {
		t.Errorf("editing the saved deals changed the live ones: %v", got)
	}

	in := map[string]FactionState{"merchant_guild": {Discovered: true, OpinionAccum: 0.9,
		Deals: []FactionDeal{{ID: 7, Give: "gold", GiveAmt: 1}}}}
	loaded := NewDiplomacyManager()
	loaded.LoadState(in, nil)
	fs := loaded.factions["merchant_guild"]
	if fs.OpinionAccum != 0 {
		t.Errorf("load kept the opinion accumulator: %v", fs.OpinionAccum)
	}
	fs.Deals[0].GiveAmt = 50
	if in["merchant_guild"].Deals[0].GiveAmt != 1 {
		t.Error("the loaded deals share their list with the save")
	}
}
