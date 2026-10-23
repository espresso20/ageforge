package game

import (
	"strings"
	"testing"
)

// Storage is in no line, so it is never an older building: a stash can't be
// rebuilt after the Primitive Age (age lock), and upgrading one would spend a
// copy of the storage the age lock never lets you rebuild, which lowers the
// most the player can ever store.

func stoneAgeWithStashes(t *testing.T, stashes int) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.Buildings.UnlockBuilding("stash")
	ge.Buildings.counts["stash"] = stashes
	ge.advanceAge("stone_age")
	return ge
}

func TestAdvanceAge_StorageIsNeverOlder(t *testing.T) {
	ge := stoneAgeWithStashes(t, 32)
	if to, ok := ge.Buildings.UpgradeTarget("stash"); ok {
		t.Fatalf("stash upgrades to %s; storage is in no line", to.Key)
	}
	for _, u := range ge.GetAvailableUpgrades() {
		if ge.Buildings.defs[u.FromKey].Category == "storage" || ge.Buildings.defs[u.ToKey].Category == "storage" {
			t.Errorf("storage upgrade listed: %s -> %s", u.FromKey, u.ToKey)
		}
	}
	for _, l := range ge.GetState().Log {
		if strings.Contains(l.Message, "upgrade stash") {
			t.Errorf("log advises upgrading stashes: %q", l.Message)
		}
	}
	// A Stash holds nothing: it boosts production, and the store measure
	// takes nothing from it.
	if got := ge.Buildings.GetStorageBonuses()["all"]; got != 0 {
		t.Errorf("storage bonus after advancing = %v with 32 Stashes, want 0", got)
	}
	if err := ge.UpgradeBuilding("stash", 0, true); err == nil || !strings.Contains(err.Error(), "has no upgrade") {
		t.Errorf("upgrade stash: err = %v, want a no-upgrade refusal", err)
	}
	if got := ge.Buildings.GetCount("stash"); got != 32 {
		t.Errorf("stash = %d after the refused upgrade, want 32", got)
	}
}

func TestBuildPreviousAgeStorage_DoesNotAdviseUpgrade(t *testing.T) {
	ge := stoneAgeWithStashes(t, 32)
	err := ge.BuildBuilding("stash")
	if err == nil {
		t.Fatal("built a Primitive Age stash in the Stone Age")
	}
	if strings.Contains(err.Error(), "upgrade") {
		t.Errorf("error advises an upgrade that does not exist: %v", err)
	}
}
