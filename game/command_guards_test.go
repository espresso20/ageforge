package game

import (
	"math"
	"testing"
)

// The smoke fuzzer typed "recruit 9223372036854775807" and the pop-cap check
// (TotalPop()+count > popCap) wrapped negative, so the recruit went through
// and population landed near -9.2e18. These tests pin the guards on every
// count and amount the engine takes from a command.

func TestWorkerManager_RecruitNoOverflow(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)

	for _, n := range []int{math.MaxInt, math.MaxInt - 4, math.MinInt, -1, 0} {
		if vm.Recruit("worker", n, 10) {
			t.Errorf("Recruit(%d) with 5/10 pop succeeded", n)
		}
		if got := vm.TotalPop(); got != 5 {
			t.Fatalf("after Recruit(%d): pop = %d, want 5", n, got)
		}
	}
	if !vm.Recruit("worker", 5, 10) {
		t.Error("Recruit filling the cap exactly should succeed")
	}
}

func TestWorkerManager_AssignUnassignRejectNonPositive(t *testing.T) {
	vm := NewWorkerManager()
	vm.UnlockType("worker")
	vm.Recruit("worker", 5, 10)
	vm.Assign("worker", "gathering_camp", 2)

	for _, n := range []int{0, -1, math.MinInt} {
		if vm.Assign("worker", "gathering_camp", n) {
			t.Errorf("Assign(%d) succeeded", n)
		}
		if vm.Unassign("worker", "gathering_camp", n) {
			t.Errorf("Unassign(%d) succeeded", n)
		}
	}
	if got := vm.GetAssignedCount("worker", "gathering_camp"); got != 2 {
		t.Errorf("assigned = %d, want 2", got)
	}
	if got := vm.IdleCount("worker"); got != 3 {
		t.Errorf("idle = %d, want 3", got)
	}
}

func TestEngineCountGuards(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Workers.UnlockType("worker")
	ge.Buildings.counts["gathering_camp"] = 2
	ge.Buildings.counts["hut"] = 2
	ge.recalculateRates()
	if err := ge.RecruitWorker("worker", 2); err != nil {
		t.Fatalf("setup recruit: %v", err)
	}
	if err := ge.AssignWorker("gathering_camp", 1); err != nil {
		t.Fatalf("setup assign: %v", err)
	}
	before := ge.GetState()

	for _, n := range []int{math.MaxInt, math.MinInt, -1, 0} {
		if ge.RecruitWorker("worker", n) == nil {
			t.Errorf("RecruitWorker(%d) succeeded", n)
		}
		if n <= 0 {
			if ge.AssignWorker("gathering_camp", n) == nil {
				t.Errorf("AssignWorker(%d) succeeded", n)
			}
			if ge.UnassignWorker("gathering_camp", n) == nil {
				t.Errorf("UnassignWorker(%d) succeeded", n)
			}
			if _, err := ge.BuildMultiple("gathering_camp", n); err == nil {
				t.Errorf("BuildMultiple(%d) succeeded", n)
			}
		}
	}
	after := ge.GetState()
	if after.Workers.TotalPop != before.Workers.TotalPop || after.Workers.TotalIdle != before.Workers.TotalIdle {
		t.Errorf("workers changed: pop %d->%d, idle %d->%d",
			before.Workers.TotalPop, after.Workers.TotalPop, before.Workers.TotalIdle, after.Workers.TotalIdle)
	}
	if after.Buildings["gathering_camp"].Count != 2 {
		t.Errorf("gathering_camp count = %d, want 2", after.Buildings["gathering_camp"].Count)
	}
}

func TestEngineAmountGuards(t *testing.T) {
	ge := newSeededEngine(1)
	ge.applyAgeUnlocks("stone_age")
	ge.applyAgeUnlocks("bronze_age")
	ge.age = "bronze_age"
	ge.Buildings.counts["market"] = 1
	ge.recalculateRates()
	ge.Resources.Add("food", 40)
	ge.Resources.Add("wood", 40)

	bad := []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0, -5}
	for _, amt := range bad {
		if _, err := ge.ExchangeResources("food", "wood", amt); err == nil {
			t.Errorf("ExchangeResources(%v) succeeded", amt)
		}
		if _, err := ge.GatherResource("food", amt); err == nil {
			t.Errorf("GatherResource(%v) succeeded", amt)
		}
		if _, err := ge.BankWonderResource("great_monolith", "stone", amt); err == nil {
			t.Errorf("BankWonderResource(%v) succeeded", amt)
		}
	}
	for key, r := range ge.Resources.resources {
		if math.IsNaN(r.Amount) || math.IsInf(r.Amount, 0) || r.Amount < 0 {
			t.Errorf("resource %s = %v after hostile amounts", key, r.Amount)
		}
	}
}

func TestResourceManagerRefusesNaN(t *testing.T) {
	rm := NewResourceManager()
	rm.Add("food", 10)
	rm.Add("food", math.NaN())
	if rm.Remove("food", math.NaN()) {
		t.Error("Remove(NaN) succeeded")
	}
	if rm.Pay(map[string]float64{"food": math.NaN()}) {
		t.Error("Pay(NaN) succeeded")
	}
	if got := rm.Get("food"); got != 10 {
		t.Errorf("food = %v, want 10", got)
	}
}
