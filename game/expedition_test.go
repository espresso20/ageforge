package game

import (
	"testing"
)

// setResource unlocks a resource, gives it headroom, zeroes its production rate,
// and sets its amount to EXACTLY amount (overriding any starting amount the engine
// seeds). Caller must NOT hold the engine lock.
func setResource(ge *GameEngine, key string, amount float64) {
	ge.mu.Lock()
	ge.Resources.UnlockResource(key)
	r, ok := ge.Resources.resources[key]
	if ok {
		r.Storage = amount + 1000
		r.Amount = amount
		r.Rate = 0
	}
	ge.mu.Unlock()
}

func setSoldiers(ge *GameEngine, amount float64) {
	setResource(ge, "soldiers", amount)
}

func setAge(ge *GameEngine, age string) {
	ge.mu.Lock()
	ge.age = age
	ge.mu.Unlock()
}

// TestLaunchExpedition_SpendsSoldiersAndCost verifies a launch deducts both the
// soldiers resource (SoldiersNeeded) and the Cost resources.
func TestLaunchExpedition_SpendsSoldiersAndCost(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "iron_age")

	// trade_escort: iron_age, SoldiersNeeded 3, no Cost.
	setSoldiers(ge, 10)

	if err := ge.LaunchExpedition("trade_escort"); err != nil {
		t.Fatalf("LaunchExpedition(trade_escort) returned error: %v", err)
	}

	got := ge.Resources.Get("soldiers")
	if got != 7 {
		t.Errorf("soldiers after launch = %v, want 7 (10 - 3)", got)
	}
}

// TestLaunchExpedition_DeductsCostResources verifies an expedition with a Cost
// map deducts those resources too (scout_party in stone_age, food+wood, no soldiers).
func TestLaunchExpedition_DeductsCostResources(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "stone_age")

	setResource(ge, "food", 100)
	setResource(ge, "wood", 100)

	if err := ge.LaunchExpedition("scout_party"); err != nil {
		t.Fatalf("LaunchExpedition(scout_party) returned error: %v", err)
	}

	if food := ge.Resources.Get("food"); food != 70 {
		t.Errorf("food after scout_party = %v, want 70 (100 - 30)", food)
	}
	if wood := ge.Resources.Get("wood"); wood != 70 {
		t.Errorf("wood after scout_party = %v, want 70 (100 - 30)", wood)
	}
}

// TestLaunchExpedition_InsufficientSoldiers_NoDeduction verifies a launch with
// too few soldiers errors and deducts nothing.
func TestLaunchExpedition_InsufficientSoldiers_NoDeduction(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "iron_age")

	// conquer_territory needs 10 soldiers; give only 4.
	setSoldiers(ge, 4)

	err := ge.LaunchExpedition("conquer_territory")
	if err == nil {
		t.Fatalf("expected error launching with insufficient soldiers, got nil")
	}
	if got := ge.Resources.Get("soldiers"); got != 4 {
		t.Errorf("soldiers after failed launch = %v, want 4 (unchanged)", got)
	}
	if ge.Military.HasActive() {
		t.Errorf("expedition should not be active after a failed launch")
	}
}

// TestLaunchExpedition_InsufficientCost_NoDeduction verifies a launch that can
// afford soldiers but not the Cost resources errors and deducts NOTHING
// (including soldiers).
func TestLaunchExpedition_InsufficientCost_NoDeduction(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "stone_age")

	// scout_party costs food 30 + wood 30, no soldiers. Give enough food, too little wood.
	setResource(ge, "food", 100)
	setResource(ge, "wood", 10)

	err := ge.LaunchExpedition("scout_party")
	if err == nil {
		t.Fatalf("expected error launching scout_party with insufficient wood, got nil")
	}
	if food := ge.Resources.Get("food"); food != 100 {
		t.Errorf("food after failed launch = %v, want 100 (unchanged — no partial charge)", food)
	}
	if wood := ge.Resources.Get("wood"); wood != 10 {
		t.Errorf("wood after failed launch = %v, want 10 (unchanged)", wood)
	}
	if ge.Military.HasActive() {
		t.Errorf("expedition should not be active after a failed launch")
	}
}

// TestScoutParty_LaunchableInStoneAge_RejectedPastBronze verifies scout_party is
// available with food+wood and no soldiers in stone_age, and rejected past its
// MaxAge (bronze_age) with no deduction.
func TestScoutParty_LaunchableInStoneAge_RejectedPastBronze(t *testing.T) {
	// Launchable in stone_age with no soldiers.
	ge := NewGameEngine()
	setAge(ge, "stone_age")
	setResource(ge, "food", 100)
	setResource(ge, "wood", 100)
	// Note: no soldiers resource at all.

	if err := ge.LaunchExpedition("scout_party"); err != nil {
		t.Fatalf("scout_party should be launchable in stone_age, got: %v", err)
	}

	// Rejected past MaxAge (iron_age is past bronze_age).
	ge2 := NewGameEngine()
	setAge(ge2, "iron_age")
	setResource(ge2, "food", 100)
	setResource(ge2, "wood", 100)

	err := ge2.LaunchExpedition("scout_party")
	if err == nil {
		t.Fatalf("expected scout_party to be rejected past bronze_age, got nil")
	}
	if food := ge2.Resources.Get("food"); food != 100 {
		t.Errorf("food after rejected past-MaxAge launch = %v, want 100 (unchanged)", food)
	}
	if ge2.Military.HasActive() {
		t.Errorf("expedition should not be active after a rejected launch")
	}
}

// TestExpedition_TickNoLongerLosesSoldiers verifies the worker-loss hack is gone:
// resolving an expedition (win or lose) does not touch the worker pool. Soldiers
// are spent at launch only.
func TestExpedition_ResolvesWithoutWorkerLoss(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "iron_age")
	setSoldiers(ge, 10)

	ge.mu.Lock()
	ge.Workers.UnlockType("worker")
	ge.Workers.Recruit("worker", 5, 100)
	popBefore := ge.Workers.TotalPop()
	ge.mu.Unlock()

	if err := ge.LaunchExpedition("trade_escort"); err != nil {
		t.Fatalf("LaunchExpedition error: %v", err)
	}

	// Drive the expedition to resolution. trade_escort's active duration is now
	// randomized (up to ~100 ticks, Bug RQPHYAHC), so budget generously; the loop
	// exits as soon as it resolves.
	for i := 0; i < 200 && ge.Military.HasActive(); i++ {
		ge.mu.Lock()
		ge.processExpeditions()
		ge.mu.Unlock()
	}

	if ge.Military.HasActive() {
		t.Fatalf("expedition did not resolve within tick budget")
	}
	if popAfter := ge.Workers.TotalPop(); popAfter != popBefore {
		t.Errorf("worker pop changed by expedition resolution: before=%d after=%d (should be unchanged)", popBefore, popAfter)
	}
}
