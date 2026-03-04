package game

import (
	"testing"
)

// gathering_camp is a stone_age food-domain building (WorkerDomain:"food", WorkerCapacity:5)

func TestVillagerManager_RecruitAndPop(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")

	if vm.TotalPop() != 0 {
		t.Errorf("initial pop = %v, want 0", vm.TotalPop())
	}

	ok := vm.Recruit("worker", 3, 10)
	if !ok {
		t.Error("Recruit should succeed under pop cap")
	}
	if vm.TotalPop() != 3 {
		t.Errorf("pop after recruit = %v, want 3", vm.TotalPop())
	}
}

func TestVillagerManager_RecruitOverCap(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")

	ok := vm.Recruit("worker", 5, 3)
	if ok {
		t.Error("Recruit should fail when count exceeds pop cap")
	}
	if vm.TotalPop() != 0 {
		t.Errorf("pop should be 0 after failed recruit, got %v", vm.TotalPop())
	}
}

func TestVillagerManager_RecruitUnlocked(t *testing.T) {
	vm := NewVillagerManager()

	// Not unlocked
	ok := vm.Recruit("worker", 1, 10)
	if ok {
		t.Error("Recruit should fail for locked type")
	}
}

func TestVillagerManager_AssignAndUnassign(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)

	// Assign 3 workers to gathering_camp (food domain building)
	ok := vm.Assign("worker", "gathering_camp", 3)
	if !ok {
		t.Error("Assign should succeed for valid domain+building pair")
	}
	if vm.IdleCount("worker") != 2 {
		t.Errorf("idle after assign = %v, want 2", vm.IdleCount("worker"))
	}

	// Unassign 1
	ok = vm.Unassign("worker", "gathering_camp", 1)
	if !ok {
		t.Error("Unassign should succeed")
	}
	if vm.IdleCount("worker") != 3 {
		t.Errorf("idle after unassign = %v, want 3", vm.IdleCount("worker"))
	}

	// Can't unassign more than assigned
	ok = vm.Unassign("worker", "gathering_camp", 10)
	if ok {
		t.Error("Unassign should fail when count > assigned")
	}
}

func TestVillagerManager_AssignMoreThanIdle(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 2, 10)

	ok := vm.Assign("worker", "gathering_camp", 5)
	if ok {
		t.Error("Assign should fail when count > idle")
	}
}

func TestVillagerManager_AssignWrongDomain(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("scholar") // knowledge domain
	vm.Recruit("scholar", 3, 10)

	// gathering_camp is food domain, not knowledge — should fail
	ok := vm.Assign("scholar", "gathering_camp", 1)
	if ok {
		t.Error("Assign should fail when building domain doesn't match worker domain")
	}
}

func TestVillagerManager_FoodDrain(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")

	if vm.FoodDrain() != 0 {
		t.Errorf("drain with no villagers = %v, want 0", vm.FoodDrain())
	}

	vm.Recruit("worker", 5, 10)
	drain := vm.FoodDrain()
	if drain <= 0 {
		t.Errorf("drain with 5 workers = %v, want > 0", drain)
	}

	// Drain should scale with count
	vm.Recruit("worker", 5, 20)
	drain2 := vm.FoodDrain()
	if drain2 <= drain {
		t.Errorf("drain should increase: 5 workers=%v, 10 workers=%v", drain, drain2)
	}
}

func TestVillagerManager_GetAssignedCount(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)

	vm.Assign("worker", "gathering_camp", 3)

	count := vm.GetAssignedCount("food", "gathering_camp")
	if count != 3 {
		t.Errorf("GetAssignedCount = %v, want 3", count)
	}

	// Zero for unassigned building
	count2 := vm.GetAssignedCount("food", "farm")
	if count2 != 0 {
		t.Errorf("GetAssignedCount for unassigned building = %v, want 0", count2)
	}
}

func TestVillagerManager_RemoveSoldiers(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("soldier")
	vm.Recruit("soldier", 5, 10)

	vm.RemoveSoldiers(2)
	if vm.TotalPop() != 3 {
		t.Errorf("pop after removing 2 soldiers = %v, want 3", vm.TotalPop())
	}
}

func TestVillagerManager_SaveLoadRoundTrip(t *testing.T) {
	vm := NewVillagerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)
	vm.Assign("worker", "gathering_camp", 3)

	saved := vm.GetAll()

	vm2 := NewVillagerManager()
	vm2.UnlockType("worker")
	vm2.LoadVillagers(saved)

	if vm2.TotalPop() != 5 {
		t.Errorf("loaded pop = %v, want 5", vm2.TotalPop())
	}

	// Building assignments should survive the round trip
	count := vm2.GetAssignedCount("food", "gathering_camp")
	if count != 3 {
		t.Errorf("loaded assigned count = %v, want 3", count)
	}
}

func TestVillagerManager_LegacyLoadCompat(t *testing.T) {
	// Simulate an old-format save with resource-based assignments — should load without crash
	vm := NewVillagerManager()
	vm.UnlockType("worker")

	oldData := map[string]VillagerInfo{
		"worker": {Count: 5, FoodCost: 0.1, Assignment: map[string]int{"food": 3, "wood": 2}},
	}
	vm.LoadVillagers(oldData)

	// Count should be restored, but legacy resource assignments are discarded (workers idle)
	if vm.TotalPop() != 5 {
		t.Errorf("loaded pop from legacy save = %v, want 5", vm.TotalPop())
	}
	if vm.IdleCount("food") != 5 {
		t.Errorf("idle count after legacy load = %v, want 5 (all idle)", vm.IdleCount("food"))
	}
}
