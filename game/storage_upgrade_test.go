package game

import (
	"strings"
	"testing"
)

// The stash -> storage_pit upgrade trap: stashes can't be rebuilt after the
// Primitive Age (age lock), and each upgrade spent one of the 25 capped
// storage pit slots, so every upgrade lowered the most the player could ever
// store. With 32 stashes, upgrading a third one made Bronze's 30K food
// requirement unstorable, and the log told the player to do it.

func stoneAgeWithStashes(t *testing.T, stashes int) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.Buildings.UnlockBuilding("stash")
	ge.Buildings.counts["stash"] = stashes
	ge.advanceAge("stone_age")
	return ge
}

func TestAdvanceAge_StorageNeverOffersUpgrade(t *testing.T) {
	ge := stoneAgeWithStashes(t, 32)
	if to, ok := ge.Buildings.GetPendingUpgrade("stash"); ok {
		t.Fatalf("stash has a pending upgrade to %s; storage must never transform", to)
	}
	for _, u := range ge.GetAvailableUpgrades() {
		if ge.Buildings.defs[u.FromKey].Category == "storage" || ge.Buildings.defs[u.ToKey].Category == "storage" {
			t.Errorf("storage upgrade offered: %s -> %s", u.FromKey, u.ToKey)
		}
	}
	for _, l := range ge.GetState().Log {
		if strings.Contains(l.Message, "upgrade stash") {
			t.Errorf("log still advises upgrading stashes: %q", l.Message)
		}
	}
	if got := ge.Buildings.GetStorageBonuses()["all"]; got != 32*500 {
		t.Errorf("storage bonus after advancing = %v, want %v", got, 32*500)
	}
}

func TestLoadGame_StorageNeverOffersUpgrade(t *testing.T) {
	isolateAccountDir(t)
	ge := stoneAgeWithStashes(t, 32)
	const name = "storage_upgrade_trap"
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame(name); err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	if to, ok := loaded.Buildings.GetPendingUpgrade("stash"); ok {
		t.Errorf("reloaded save offers stash -> %s", to)
	}
}

func TestUpgradeBuilding_RespectsTargetMaxCount(t *testing.T) {
	ge := stoneAgeWithStashes(t, 32)
	// Force the old offer back, as a pre-fix save in memory would have it.
	ge.Buildings.SetPendingUpgrade("stash", "storage_pit")
	ge.Buildings.counts["storage_pit"] = 24
	for _, r := range []string{"wood", "stone"} {
		ge.Resources.AddStorage(r, 1e9)
		ge.Resources.Add(r, 1e9)
	}
	if err := ge.UpgradeBuilding("stash", 0, true); err != nil {
		t.Fatalf("UpgradeBuilding: %v", err)
	}
	if got := ge.Buildings.GetCount("storage_pit"); got != 25 {
		t.Errorf("storage_pit = %d after upgrading all, want the MaxCount 25", got)
	}
	if got := ge.Buildings.GetCount("stash"); got != 31 {
		t.Errorf("stash = %d, want 31 (only one slot was free)", got)
	}
	if err := ge.UpgradeBuilding("stash", 1, false); err == nil || !strings.Contains(err.Error(), "max count") {
		t.Errorf("upgrade into a full storage_pit: err = %v, want a max count error", err)
	}
	if moved := ge.Buildings.PartialTransform("stash", "storage_pit", 5, nil); moved != 0 {
		t.Errorf("PartialTransform moved %d copies past MaxCount", moved)
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
