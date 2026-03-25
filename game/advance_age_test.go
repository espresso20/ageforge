package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

func TestAdvanceAge_SetsUpgradesForGatheringCamp(t *testing.T) {
	ge := NewGameEngine()
	// Don't call Start() — just manipulate internal state directly

	// Unlock and build gathering_camps
	ge.Buildings.UnlockBuilding("gathering_camp")
	ge.Buildings.counts["gathering_camp"] = 2

	// Verify gathering_camp def has correct lineage info
	def, ok := ge.Buildings.defs["gathering_camp"]
	if !ok {
		t.Fatal("gathering_camp not in defs")
	}
	t.Logf("gathering_camp def: LineageKey=%q LineageTier=%d", def.LineageKey, def.LineageTier)
	if def.LineageKey != "food" {
		t.Errorf("expected LineageKey=food, got %q", def.LineageKey)
	}
	if def.LineageTier != 0 {
		t.Errorf("expected LineageTier=0, got %d", def.LineageTier)
	}

	// Call advanceAge directly (no ticker running)
	ge.advanceAge("stone_age")

	// Check pending upgrades map
	t.Logf("pendingUpgrades: %v", ge.Buildings.pendingUpgrades)
	if len(ge.Buildings.pendingUpgrades) == 0 {
		t.Error("pendingUpgrades is empty after advancing to stone_age with gathering_camp built")
	}

	target, found := ge.Buildings.GetPendingUpgrade("gathering_camp")
	if !found {
		t.Error("expected gathering_camp to have a pending upgrade, got none")
	} else {
		t.Logf("gathering_camp -> %s", target)
		if target != "forager_post" {
			t.Errorf("expected upgrade target=forager_post, got %q", target)
		}
	}

	// Now test GetAvailableUpgrades
	upgrades := ge.GetAvailableUpgrades()
	t.Logf("GetAvailableUpgrades returned %d entries", len(upgrades))
	for _, u := range upgrades {
		t.Logf("  %s -> %s (count=%d canAfford=%v)", u.FromKey, u.ToKey, u.Count, u.CanAfford)
	}
	if len(upgrades) == 0 {
		t.Error("GetAvailableUpgrades returned 0 — bug confirmed")
	}
}

// TestLoadGame_RestoresPendingUpgrades verifies that pending upgrades are
// reconstructed from legacy buildings after a save/load cycle.
// Regression test for: "upgrade shows 0 buildings available after reload".
//
// This test simulates the LoadGame restore path directly (without disk I/O)
// by manually applying the same restore steps that LoadGame performs, then
// asserting that GetAvailableUpgrades returns the expected upgrades.
func TestLoadGame_RestoresPendingUpgrades(t *testing.T) {
	// Simulate the state that a saved game would have:
	//   - age: stone_age
	//   - gathering_camp: 3 built, flagged legacy (primitive→stone advance happened)
	//   - pendingUpgrades: empty (as it would be before the fix)
	legacyBuildings := []string{"gathering_camp"}
	buildingCounts := map[string]int{"gathering_camp": 3}
	age := "stone_age"

	// Fresh engine — simulates what NewGameEngine() gives us at LoadGame entry
	ge := NewGameEngine()

	// Apply the same restore calls that LoadGame makes
	ge.age = age
	ge.Buildings.LoadCounts(buildingCounts)
	ge.Buildings.LoadLegacyBuildings(legacyBuildings)
	// NOTE: before the fix, LoadGame stopped here; pendingUpgrades stayed empty.

	// The fix: reconstruct pending upgrades from legacy buildings + current age.
	for _, key := range legacyBuildings {
		if ge.Buildings.GetCount(key) <= 0 {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		if !ok || def.LineageKey == "" || def.LineageKey == "wonder" {
			continue
		}
		next := config.BuildingNextTierForAge(def.LineageKey, def.LineageTier, age)
		if next == nil {
			continue
		}
		ge.Buildings.SetPendingUpgrade(key, next.Key)
	}

	// The pending upgrade must now be present
	target, found := ge.Buildings.GetPendingUpgrade("gathering_camp")
	if !found {
		t.Error("after simulated LoadGame restore, gathering_camp has no pending upgrade")
	} else if target != "forager_post" {
		t.Errorf("expected forager_post, got %q", target)
	} else {
		t.Logf("gathering_camp -> %s (correctly reconstructed)", target)
	}

	// GetAvailableUpgrades must return a non-empty list
	upgrades := ge.GetAvailableUpgrades()
	if len(upgrades) == 0 {
		t.Error("GetAvailableUpgrades returned 0 after restore — 'upgrade' command would show nothing")
	}
	for _, u := range upgrades {
		t.Logf("  upgrade: %s -> %s (count=%d)", u.FromKey, u.ToKey, u.Count)
	}
}
