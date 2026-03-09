package game

import (
	"os"
	"testing"
)

func TestEngine_NewEngineStartsInPrimitive(t *testing.T) {
	ge := NewGameEngine()
	state := ge.GetState()

	if state.Age != "primitive_age" {
		t.Errorf("starting age = %v, want primitive_age", state.Age)
	}
	if state.Tick != 0 {
		t.Errorf("starting tick = %v, want 0", state.Tick)
	}
}

func TestEngine_StartingResources(t *testing.T) {
	ge := NewGameEngine()
	state := ge.GetState()

	food := state.Resources["food"]
	if food.Amount != 25 {
		t.Errorf("starting food = %v, want 25", food.Amount)
	}
	wood := state.Resources["wood"]
	if wood.Amount != 50 {
		t.Errorf("starting wood = %v, want 50", wood.Amount)
	}
}

func TestEngine_GatherResource(t *testing.T) {
	ge := NewGameEngine()

	_, err := ge.GatherResource("food", 5)
	if err != nil {
		t.Errorf("GatherResource failed: %v", err)
	}

	state := ge.GetState()
	if state.Resources["food"].Amount != 30 { // 25 starting + 5
		t.Errorf("food after gather = %v, want 30", state.Resources["food"].Amount)
	}
}

func TestEngine_GatherLockedResource(t *testing.T) {
	ge := NewGameEngine()

	_, err := ge.GatherResource("iron", 5)
	if err == nil {
		t.Error("GatherResource should fail for locked resource")
	}
}

func TestEngine_BuildBuilding(t *testing.T) {
	ge := NewGameEngine()

	// Give enough resources
	ge.mu.Lock()
	ge.Resources.Add("wood", 1000)
	ge.mu.Unlock()

	err := ge.BuildBuilding("hut")
	if err != nil {
		t.Errorf("BuildBuilding(hut) failed: %v", err)
	}

	state := ge.GetState()
	hutState := state.Buildings["hut"]
	if hutState.Count == 0 && len(state.BuildQueue) == 0 {
		t.Error("hut should be built or in queue")
	}
}

func TestEngine_BuildUnknownBuilding(t *testing.T) {
	ge := NewGameEngine()

	err := ge.BuildBuilding("nonexistent_building_xyz")
	if err == nil {
		t.Error("BuildBuilding should fail for unknown building")
	}
}

func TestEngine_BuildLockedBuilding(t *testing.T) {
	ge := NewGameEngine()

	// farm is stone_age, should be locked in primitive
	err := ge.BuildBuilding("farm")
	if err == nil {
		t.Error("BuildBuilding should fail for locked building")
	}
}

func TestEngine_RecruitWorker(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Resources.Add("wood", 5000)
	ge.Buildings.counts["hut"] = 5
	ge.mu.Unlock()

	err := ge.RecruitWorker("food", 2)
	if err != nil {
		t.Errorf("RecruitWorker failed: %v", err)
	}

	state := ge.GetState()
	if state.Workers.TotalPop != 2 {
		t.Errorf("pop after recruit = %v, want 2", state.Workers.TotalPop)
	}
}

func TestEngine_RecruitOverCap(t *testing.T) {
	ge := NewGameEngine()

	err := ge.RecruitWorker("food", 1)
	if err == nil {
		t.Error("RecruitWorker should fail with 0 pop cap")
	}
}

func TestEngine_AssignWorker(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Buildings.counts["hut"] = 5
	// gathering_camp is the food-domain building (unlocks at stone_age but we bypass that here)
	ge.Buildings.counts["gathering_camp"] = 1
	ge.Buildings.unlocked["gathering_camp"] = true
	ge.mu.Unlock()

	ge.RecruitWorker("food", 3)
	// Assign using building key only (domain auto-resolved)
	err := ge.AssignWorker("gathering_camp", 2)
	if err != nil {
		t.Errorf("AssignWorker failed: %v", err)
	}

	state := ge.GetState()
	if state.Workers.TotalIdle != 1 {
		t.Errorf("idle after assign = %v, want 1", state.Workers.TotalIdle)
	}
}

func TestEngine_StartResearch(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Resources.AddStorage("knowledge", 10000)
	ge.Resources.Add("knowledge", 2500)
	ge.mu.Unlock()

	err := ge.StartResearch("tool_making")
	if err != nil {
		t.Errorf("StartResearch failed: %v", err)
	}

	state := ge.GetState()
	if state.Research.CurrentTech != "tool_making" {
		t.Errorf("current tech = %v, want tool_making", state.Research.CurrentTech)
	}
}

func TestEngine_CancelResearch(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Resources.AddStorage("knowledge", 10000)
	ge.Resources.Add("knowledge", 2500)
	ge.mu.Unlock()

	ge.StartResearch("tool_making")
	err := ge.CancelResearch()
	if err != nil {
		t.Errorf("CancelResearch failed: %v", err)
	}

	state := ge.GetState()
	if state.Research.CurrentTech != "" {
		t.Errorf("current tech after cancel = %v, want empty", state.Research.CurrentTech)
	}
}

func TestEngine_GetState_Consistency(t *testing.T) {
	ge := NewGameEngine()

	state := ge.GetState()

	if state.AgeName == "" {
		t.Error("AgeName should not be empty")
	}
	if state.Resources == nil {
		t.Error("Resources should not be nil")
	}
	if state.Buildings == nil {
		t.Error("Buildings should not be nil")
	}
	if state.Milestones.Milestones == nil {
		t.Error("Milestones map should not be nil")
	}
	if state.Milestones.TotalCount == 0 {
		t.Error("TotalCount should be > 0")
	}
}

func TestEngine_SpeedMultiplier(t *testing.T) {
	ge := NewGameEngine()

	err := ge.SetSpeedMultiplier(1.5)
	if err == nil {
		t.Error("SetSpeedMultiplier(1.5) should fail with no wonders")
	}

	if ge.GetSpeedMultiplier() != 1.0 {
		t.Errorf("default speed = %v, want 1.0", ge.GetSpeedMultiplier())
	}
}

func TestEngine_Reset(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Resources.Add("wood", 500)
	ge.Buildings.counts["hut"] = 10
	ge.mu.Unlock()

	ge.Reset()
	state := ge.GetState()

	if state.Age != "primitive_age" {
		t.Errorf("age after reset = %v, want primitive_age", state.Age)
	}
	if state.Resources["wood"].Amount != 50 {
		t.Errorf("wood after reset = %v, want 50", state.Resources["wood"].Amount)
	}
}

func TestEngine_MilestoneEvents(t *testing.T) {
	ge := NewGameEngine()

	milestoneReceived := false
	ge.Bus.Subscribe(EventMilestoneCompleted, func(e EventData) {
		milestoneReceived = true
	})

	ge.mu.Lock()
	ge.Buildings.UnlockBuilding("hut")
	ge.Buildings.counts["hut"] = 1
	ge.checkMilestones()
	ge.mu.Unlock()

	if !milestoneReceived {
		t.Error("EventMilestoneCompleted should fire when milestone is achieved")
	}
}

func TestEngine_ChainEvents(t *testing.T) {
	ge := NewGameEngine()

	chainReceived := false
	ge.Bus.Subscribe(EventChainCompleted, func(e EventData) {
		chainReceived = true
	})

	ge.mu.Lock()
	// Military chain requires all 5 milestones: first_soldiers, war_machine,
	// iron_legion, fortress_state, military_superpower
	ge.Milestones.completed["first_soldiers"] = true
	ge.Milestones.completed["war_machine"] = true
	ge.Milestones.completed["iron_legion"] = true
	ge.Milestones.completed["fortress_state"] = true
	ge.Milestones.completed["military_superpower"] = true
	ge.checkMilestones()
	ge.mu.Unlock()

	if !chainReceived {
		t.Error("EventChainCompleted should fire when chain completes")
	}
}

func TestEngine_BuildMultiple(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Resources.AddStorage("wood", 100000)
	ge.Resources.Add("wood", 50000)
	ge.mu.Unlock()

	built, err := ge.BuildMultiple("hut", 5)
	if err != nil {
		t.Errorf("BuildMultiple failed: %v", err)
	}
	if built != 5 {
		t.Errorf("built = %v, want 5", built)
	}
}

func TestEngine_SaveLoadRoundTrip(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	ge.Resources.Add("wood", 500)
	ge.Resources.Add("food", 200)
	ge.Buildings.counts["hut"] = 3
	ge.mu.Unlock()

	ge.RecruitWorker("food", 2)

	err := ge.SaveGame("test_roundtrip")
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove("data/saves/test_roundtrip.json")

	ge2 := NewGameEngine()
	err = ge2.LoadGame("test_roundtrip")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	state := ge2.GetState()
	if state.Buildings["hut"].Count != 3 {
		t.Errorf("loaded hut count = %v, want 3", state.Buildings["hut"].Count)
	}
	if state.Workers.TotalPop != 2 {
		t.Errorf("loaded pop = %v, want 2", state.Workers.TotalPop)
	}
}
