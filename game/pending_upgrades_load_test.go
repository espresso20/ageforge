package game

import (
	"maps"
	"reflect"
	"slices"
	"testing"
)

// pendingUpgradeGame builds a Bronze Age game the way live play gets there:
// every Primitive Age building built, then two age advances. Offers made on
// entering the Stone Age stand in the Bronze Age for lineages with no Bronze
// tier (gathering_camp -> forager_post is one), which is exactly the set the
// old load rebuild dropped: seed 1 offered 7 upgrades live and 5 after a load.
func pendingUpgradeGame(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	for key, def := range ge.Buildings.defs {
		if def.RequiredAge == "primitive_age" && def.Category != "wonder" && ge.Buildings.IsUnlocked(key) {
			ge.Buildings.counts[key] = 2
		}
	}
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	ge.advanceAge("bronze_age")
	ge.pendingCatastrophe = ""
	if _, ok := ge.Buildings.GetPendingUpgrade("gathering_camp"); !ok {
		t.Fatal("setup: gathering_camp has no pending upgrade in the Bronze Age")
	}
	return ge
}

func TestLoadGame_KeepsPendingUpgradesFromEarlierAges(t *testing.T) {
	isolateAccountDir(t)
	ge := pendingUpgradeGame(t)
	live := ge.Buildings.GetAllPendingUpgrades()
	liveOffers := ge.GetAvailableUpgrades()

	if err := ge.SaveGame("pending-upgrades"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("pending-upgrades"); err != nil {
		t.Fatal(err)
	}
	if got := loaded.Buildings.GetAllPendingUpgrades(); !maps.Equal(got, live) {
		t.Errorf("pending upgrades after load:\n got %v\nwant %v", got, live)
	}
	if got := loaded.GetAvailableUpgrades(); !reflect.DeepEqual(got, liveOffers) {
		t.Errorf("available upgrades after load: got %d, want %d\n got %+v\nwant %+v", len(got), len(liveOffers), got, liveOffers)
	}
	if err := loaded.UpgradeBuilding("gathering_camp", 1, false); err != nil && err.Error() == "no upgrade available for gathering_camp" {
		t.Errorf("upgrade gathering_camp after load: %v", err)
	}
}

// TestRebuildPendingUpgrades_MatchesLive covers saves written before the
// pending set was persisted: the rebuild from the legacy set must give the
// same offers live play made.
func TestRebuildPendingUpgrades_MatchesLive(t *testing.T) {
	ge := pendingUpgradeGame(t)
	live := ge.Buildings.GetAllPendingUpgrades()
	legacy := ge.Buildings.GetLegacyBuildings()

	fresh := NewGameEngine()
	fresh.Buildings.LoadCounts(ge.Buildings.GetAll())
	fresh.Buildings.LoadLegacyBuildings(legacy)
	got := fresh.rebuildPendingUpgrades(legacy, ge.age)
	if !maps.Equal(got, live) {
		t.Errorf("rebuilt pending upgrades:\n got %v\nwant %v\n(legacy %v)", got, live, slices.Sorted(slices.Values(legacy)))
	}
}
