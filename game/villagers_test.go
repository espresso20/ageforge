package game

import (
	"testing"
)

// gathering_camp is a stone_age food-domain building (WorkerDomain:"food", WorkerCapacity:5)

func TestWorkerManager_RecruitAndPop(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")

	if vm.TotalPop() != 0 {
		t.Errorf("initial pop = %v, want 0", vm.TotalPop())
	}

	ok := vm.Recruit("food", 3, 10)
	if !ok {
		t.Error("Recruit should succeed under pop cap")
	}
	if vm.TotalPop() != 3 {
		t.Errorf("pop after recruit = %v, want 3", vm.TotalPop())
	}
}

func TestWorkerManager_RecruitOverCap(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")

	ok := vm.Recruit("food", 5, 3)
	if ok {
		t.Error("Recruit should fail when count exceeds pop cap")
	}
	if vm.TotalPop() != 0 {
		t.Errorf("pop should be 0 after failed recruit, got %v", vm.TotalPop())
	}
}

func TestWorkerManager_RecruitUnlocked(t *testing.T) {
	vm := NewWorkerManager()

	// Not unlocked
	ok := vm.Recruit("food", 1, 10)
	if ok {
		t.Error("Recruit should fail for locked type")
	}
}

func TestWorkerManager_AssignAndUnassign(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")
	vm.Recruit("food", 5, 10)

	// Assign 3 workers to gathering_camp (food domain building)
	ok := vm.Assign("food", "gathering_camp", 3)
	if !ok {
		t.Error("Assign should succeed for valid domain+building pair")
	}
	if vm.IdleCount("food") != 2 {
		t.Errorf("idle after assign = %v, want 2", vm.IdleCount("food"))
	}

	// Unassign 1
	ok = vm.Unassign("food", "gathering_camp", 1)
	if !ok {
		t.Error("Unassign should succeed")
	}
	if vm.IdleCount("food") != 3 {
		t.Errorf("idle after unassign = %v, want 3", vm.IdleCount("food"))
	}

	// Can't unassign more than assigned
	ok = vm.Unassign("food", "gathering_camp", 10)
	if ok {
		t.Error("Unassign should fail when count > assigned")
	}
}

func TestWorkerManager_AssignMoreThanIdle(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")
	vm.Recruit("food", 2, 10)

	ok := vm.Assign("food", "gathering_camp", 5)
	if ok {
		t.Error("Assign should fail when count > idle")
	}
}

func TestWorkerManager_AssignWrongDomain(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("knowledge") // knowledge domain
	vm.Recruit("knowledge", 3, 10)

	// gathering_camp is food domain, not knowledge — should fail
	ok := vm.Assign("knowledge", "gathering_camp", 1)
	if ok {
		t.Error("Assign should fail when building domain doesn't match worker domain")
	}
}

func TestWorkerManager_FoodDrain(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")

	if vm.FoodDrain() != 0 {
		t.Errorf("drain with no workers = %v, want 0", vm.FoodDrain())
	}

	vm.Recruit("food", 5, 10)
	drain := vm.FoodDrain()
	if drain <= 0 {
		t.Errorf("drain with 5 workers = %v, want > 0", drain)
	}

	// Drain should scale with count
	vm.Recruit("food", 5, 20)
	drain2 := vm.FoodDrain()
	if drain2 <= drain {
		t.Errorf("drain should increase: 5 workers=%v, 10 workers=%v", drain, drain2)
	}
}

func TestWorkerManager_GetAssignedCount(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")
	vm.Recruit("food", 5, 10)

	vm.Assign("food", "gathering_camp", 3)

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

func TestWorkerManager_RemoveSoldiers(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("military")
	vm.Recruit("military", 5, 10)

	vm.RemoveSoldiers(2)
	if vm.TotalPop() != 3 {
		t.Errorf("pop after removing 2 soldiers = %v, want 3", vm.TotalPop())
	}
}

func TestWorkerManager_SaveLoadRoundTrip(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("food")
	vm.Recruit("food", 5, 10)
	vm.Assign("food", "gathering_camp", 3)

	saved := vm.GetAll()

	vm2 := NewWorkerManager()
	vm2.UnlockType("food")
	vm2.LoadWorkers(saved)

	if vm2.TotalPop() != 5 {
		t.Errorf("loaded pop = %v, want 5", vm2.TotalPop())
	}

	// Building assignments should survive the round trip
	count := vm2.GetAssignedCount("food", "gathering_camp")
	if count != 3 {
		t.Errorf("loaded assigned count = %v, want 3", count)
	}
}

func TestWorkerManager_LegacyLoadCompat(t *testing.T) {
	// Simulate an old-format save with legacy type keys ("worker") — should be silently skipped.
	// After the legacy alias removal, "worker" is not a known domain, so workers are not loaded.
	// This test verifies the loader does not crash and simply ignores unknown keys.
	vm := NewWorkerManager()
	vm.UnlockType("food")

	oldData := map[string]WorkerInfo{
		"worker": {Count: 5, FoodCost: 0.1, Assignment: map[string]int{"food": 3, "wood": 2}},
	}
	vm.LoadWorkers(oldData)

	// "worker" is no longer a known domain — workers are silently discarded.
	if vm.TotalPop() != 0 {
		t.Errorf("loaded pop from legacy save = %v, want 0 (unknown key skipped)", vm.TotalPop())
	}
	if vm.IdleCount("food") != 0 {
		t.Errorf("idle count after legacy load = %v, want 0 (unknown key skipped)", vm.IdleCount("food"))
	}
}
