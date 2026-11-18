package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The festival, the upgrade and the three kinds of pact each move their count,
// and entering an age zeroes them all.
func TestDeedsCountTheirEventsAndResetOnEnteringAnAge(t *testing.T) {
	isolateAccountDir(t)
	ge := dealEngine(t, "classical_age")
	stock(ge)
	learn(ge, "drama", "envoys")
	meet(ge, "merchant_guild", 60)
	want := func(what string, got, n int) {
		t.Helper()
		if got != n {
			t.Errorf("%s: %d, want %d", what, got, n)
		}
	}
	noErr(t, ge.DoFestival())
	want("festivals", ge.Stats.Deeds.Festivals, 1)
	noErr(t, ge.SendGift("merchant_guild"))
	_, err := ge.AcceptFactionDeal("merchant_guild", 1)
	noErr(t, err)
	noErr(t, ge.SetDiplomaticStatus("merchant_guild", "allied"))
	want("pacts", ge.Stats.Deeds.Pacts, 3)

	ge.mu.Lock()
	ge.Stats.Deeds.KeystoneTick = 9
	ge.Stats.Deeds.RouteCycle("local_barter", 5)
	ge.advanceAge("medieval_age")
	got := ge.Stats.Deeds
	ge.mu.Unlock()
	if got.Festivals != 0 || got.Pacts != 0 || got.KeystoneTick != 0 || len(got.RouteCycles) != 0 || len(got.RouteClean) != 0 {
		t.Errorf("deeds after entering an age: %+v", got)
	}
}

func TestDeedsCountUpgradesAndTheKeystone(t *testing.T) {
	isolateAccountDir(t)
	ge := campsInTheStoneAge(t, 4)
	noErr(t, ge.UpgradeBuilding("gathering_camp", 2, false))
	if ge.Stats.Deeds.Upgrades != 2 {
		t.Errorf("upgrades: %d, want 2", ge.Stats.Deeds.Upgrades)
	}
	key := ge.rules.Keystone(ge.age)
	if key == "" {
		t.Fatalf("%s has no keystone", ge.age)
	}
	ge.mu.Lock()
	ge.tick = 77
	ge.finishResearch("the_wheel")
	if ge.Stats.Deeds.KeystoneTick != 0 {
		t.Errorf("a tech that is not the keystone set the tick: %d", ge.Stats.Deeds.KeystoneTick)
	}
	ge.finishResearch(key)
	ge.mu.Unlock()
	if ge.Stats.Deeds.KeystoneTick != 77 {
		t.Errorf("keystone tick %d, want 77", ge.Stats.Deeds.KeystoneTick)
	}
}

// A route's cycles are counted; a missed cycle or a stop ends its unbroken run
// and the next finished cycle begins a new one.
func TestDeedsCountRouteCyclesAndBreaks(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "industrial_age"
	ge.Buildings.counts["market"] = 2
	setRes(ge, "food", config.Million)
	setRes(ge, "wood", config.Million)
	d := &ge.Stats.Deeds
	tm := ge.Trade
	tm.deeds = d
	noErr(t, tm.StartRoute("local_barter", ge.Buildings, ge.age, ge.progress.GetAgeOrder()))
	per := config.TradeRouteByKey()["local_barter"].TicksPerRun
	cycle := func() {
		for range per {
			tm.now++
			tm.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, 0, nil)
		}
	}
	cycle()
	cycle()
	if d.RouteCycles["local_barter"] != 2 || d.RouteClean["local_barter"] != per {
		t.Fatalf("after two cycles: %+v", *d)
	}
	setRes(ge, "food", 0) // the export is short: a missed cycle
	cycle()
	if _, ok := d.RouteClean["local_barter"]; ok || d.RouteCycles["local_barter"] != 2 {
		t.Errorf("after a missed cycle: %+v", *d)
	}
	setRes(ge, "food", config.Million)
	cycle()
	if d.RouteCycles["local_barter"] != 3 || d.RouteClean["local_barter"] != 4*per {
		t.Errorf("after the next cycle: %+v", *d)
	}
	noErr(t, tm.StopRoute("local_barter"))
	if _, ok := d.RouteClean["local_barter"]; ok {
		t.Error("a stop left the unbroken run")
	}
}

func TestDeedsSurviveASaveRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(3)
	ge.mu.Lock()
	ge.Stats.Deeds = AgeDeeds{Festivals: 2, Upgrades: 3, Pacts: 4, KeystoneTick: 50,
		RouteCycles: map[string]int{"local_barter": 6}, RouteClean: map[string]int{"local_barter": 12}}
	ge.mu.Unlock()
	got := saveAndLoad(t, ge, "deeds").Stats.Deeds
	if got.Festivals != 2 || got.Upgrades != 3 || got.Pacts != 4 || got.KeystoneTick != 50 ||
		got.RouteCycles["local_barter"] != 6 || got.RouteClean["local_barter"] != 12 {
		t.Errorf("deeds after a round trip: %+v", got)
	}
}
