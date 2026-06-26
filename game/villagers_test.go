package game

import (
	"testing"
)

// All workers belong to the single "worker" pool.
// Any building with WorkerCapacity > 0 accepts workers regardless of domain.
// gathering_camp has WorkerCapacity:5.

func TestWorkerManager_RecruitAndPop(t *testing.T) {
	vm := NewWorkerManager()
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

func TestWorkerManager_RecruitOverCap(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")

	ok := vm.Recruit("worker", 5, 3)
	if ok {
		t.Error("Recruit should fail when count exceeds pop cap")
	}
	if vm.TotalPop() != 0 {
		t.Errorf("pop should be 0 after failed recruit, got %v", vm.TotalPop())
	}
}

func TestWorkerManager_RecruitUnlocked(t *testing.T) {
	vm := NewWorkerManager()

	// Not unlocked — any domain key should fail
	ok := vm.Recruit("worker", 1, 10)
	if ok {
		t.Error("Recruit should fail for locked workers")
	}
}

func TestWorkerManager_AssignAndUnassign(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)

	// Assign 3 workers to gathering_camp (has WorkerCapacity:5)
	ok := vm.Assign("worker", "gathering_camp", 3)
	if !ok {
		t.Error("Assign should succeed for valid building with WorkerCapacity")
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

func TestWorkerManager_AssignMoreThanIdle(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 2, 10)

	ok := vm.Assign("worker", "gathering_camp", 5)
	if ok {
		t.Error("Assign should fail when count > idle")
	}
}

func TestWorkerManager_AssignAnyBuilding(t *testing.T) {
	// Workers can now be assigned to any building with WorkerCapacity > 0,
	// regardless of what resource domain that building produces.
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 3, 10)

	// gathering_camp produces food — workers can go there
	ok := vm.Assign("worker", "gathering_camp", 1)
	if !ok {
		t.Error("Assign to food-producing building should succeed")
	}
}

func TestWorkerManager_FoodDrain(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")

	if vm.FoodDrain() != 0 {
		t.Errorf("drain with no workers = %v, want 0", vm.FoodDrain())
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

func TestWorkerManager_GetAssignedCount(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)

	vm.Assign("worker", "gathering_camp", 3)

	// domain arg is ignored — always reads from single pool
	count := vm.GetAssignedCount("worker", "gathering_camp")
	if count != 3 {
		t.Errorf("GetAssignedCount = %v, want 3", count)
	}
	count2 := vm.GetAssignedCount("food", "gathering_camp") // legacy domain arg — still works
	if count2 != 3 {
		t.Errorf("GetAssignedCount with legacy domain arg = %v, want 3", count2)
	}

	// Zero for unassigned building
	count3 := vm.GetAssignedCount("worker", "farm")
	if count3 != 0 {
		t.Errorf("GetAssignedCount for unassigned building = %v, want 0", count3)
	}
}

func TestWorkerManager_SaveLoadRoundTrip(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)
	vm.Assign("worker", "gathering_camp", 3)

	saved := vm.GetAll()

	vm2 := NewWorkerManager()
	vm2.UnlockType("worker")
	vm2.LoadWorkers(saved)

	if vm2.TotalPop() != 5 {
		t.Errorf("loaded pop = %v, want 5", vm2.TotalPop())
	}

	// Building assignments should survive the round trip
	count := vm2.GetAssignedCount("worker", "gathering_camp")
	if count != 3 {
		t.Errorf("loaded assigned count = %v, want 3", count)
	}
}

func TestWorkerManager_LegacyLoadCompat(t *testing.T) {
	// Old saves had separate domain entries (food, faith, knowledge, etc.).
	// LoadWorkers now sums all domain counts into the single "worker" pool.
	vm := NewWorkerManager()
	vm.UnlockType("worker")

	oldData := map[string]WorkerInfo{
		"food":      {Count: 3, FoodCost: 0.1, Assignment: map[string]int{"gathering_camp": 2}},
		"knowledge": {Count: 2, FoodCost: 0.8, Assignment: map[string]int{}},
	}
	vm.LoadWorkers(oldData)

	// All domain counts should be summed into the single pool
	if vm.TotalPop() != 5 {
		t.Errorf("loaded pop from legacy save = %v, want 5", vm.TotalPop())
	}
	// Assignments from old domains are preserved
	assigned := vm.GetAssignedCount("worker", "gathering_camp")
	if assigned != 2 {
		t.Errorf("loaded assigned count from legacy save = %v, want 2", assigned)
	}
}
