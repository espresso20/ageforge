package game

import (
	"strings"
	"testing"
)

// lendBoonCrew grants a faction boon's temporary workers the way boon.Apply
// does for the Extra Hands boon.
func lendBoonCrew(ge *GameEngine, factionKey string, count, ticks int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	boonApplier{ge: ge, name: CivName(factionKey), key: factionKey}.GrantTempWorkers(count, ticks)
}

// lastLogWith returns the newest log entry containing s.
func lastLogWith(ge *GameEngine, s string) (LogEntry, bool) {
	logs := ge.GetLogs()
	for i := len(logs) - 1; i >= 0; i-- {
		if strings.Contains(logs[i].Message, s) {
			return logs[i], true
		}
	}
	return LogEntry{}, false
}

// TestBoonTempWorkersLeaveOnTime: the Extra Hands boon's workers used to stay
// for good. They now go home when their time is up, with a routine log line
// when nobody was staffing a building.
func TestBoonTempWorkersLeaveOnTime(t *testing.T) {
	ge := newSeededEngine(1)
	ge.mu.Lock()
	ge.Workers.AddLentWorkers(2) // the player's own
	ge.mu.Unlock()

	lendBoonCrew(ge, "merchant_guild", 5, 3)
	if got := ge.Workers.TotalPop(); got != 7 {
		t.Fatalf("population after the boon = %d, want 7", got)
	}

	ge.StepTicks(2)
	if got := ge.Workers.TotalPop(); got != 7 {
		t.Fatalf("the crew left early: population %d after 2 of 3 ticks, want 7", got)
	}
	ge.StepTicks(1)
	if got := ge.Workers.TotalPop(); got != 2 {
		t.Fatalf("population after the loan ran out = %d, want 2", got)
	}
	if n := len(ge.Diplomacy.boonLoans); n != 0 {
		t.Errorf("%d boon loans still held after they ran out", n)
	}
	e, ok := lastLogWith(ge, "went home")
	if !ok {
		t.Fatal("no log line for the crew going home")
	}
	if e.Message != "5 workers from the Merchant Guild went home." || e.Type != LogRoutine {
		t.Errorf("log = %q (%s), want %q (%s)", e.Message, e.Type, "5 workers from the Merchant Guild went home.", LogRoutine)
	}

	ge.StepTicks(5)
	if got := ge.Workers.TotalPop(); got != 2 {
		t.Errorf("population %d after more ticks: the crew left twice?", got)
	}
}

// TestBoonTempWorkersLeaveTheirPosts: a crew that was staffing buildings is
// released from its assignments the way other removals are (KillWorker: idle
// workers first, then the largest assignments), so no building claims
// workers who are gone, and the log says buildings lost staff.
func TestBoonTempWorkersLeaveTheirPosts(t *testing.T) {
	ge := newSeededEngine(1)
	ge.mu.Lock()
	ge.Workers.AddLentWorkers(2)
	ge.mu.Unlock()
	lendBoonCrew(ge, "riverlands_tribes", 5, 2)

	ge.mu.Lock()
	if !ge.Workers.Assign("", "gathering_camp", 6) {
		ge.mu.Unlock()
		t.Fatal("could not assign 6 of 7 workers")
	}
	ge.mu.Unlock()

	ge.StepTicks(2)
	if got := ge.Workers.TotalPop(); got != 2 {
		t.Fatalf("population after the loan ran out = %d, want 2", got)
	}
	if got := ge.Workers.GetAssignedCount("", "gathering_camp"); got != 2 {
		t.Errorf("gathering camps staffed by %d after the crew left, want 2 (the workers left)", got)
	}
	if got := ge.Workers.IdleCount(""); got != 0 {
		t.Errorf("idle workers = %d, want 0", got)
	}
	e, ok := lastLogWith(ge, "went home")
	if !ok {
		t.Fatal("no log line for the crew going home")
	}
	want := "5 workers from the Riverlands Tribes went home. 4 of them were staffing buildings."
	if e.Message != want || e.Type == LogRoutine {
		t.Errorf("log = %q (%s), want %q in the main log", e.Message, e.Type, want)
	}
}

// TestBoonTempWorkersSaved: a crew on loan is saved with its time left and
// still leaves on time after a load. A save with no crew has no field for it.
func TestBoonTempWorkersSaved(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	if data, err := ge.StateJSON(); err != nil {
		t.Fatal(err)
	} else if strings.Contains(string(data), "boon_workers") {
		t.Error("a save with no boon crew writes boon_workers; it should be omitted")
	}

	lendBoonCrew(ge, "merchant_guild", 4, 10)
	ge.StepTicks(4)
	if err := ge.SaveGame("boon-crew"); err != nil {
		t.Fatal(err)
	}

	b := NewGameEngine()
	if err := b.LoadGame("boon-crew"); err != nil {
		t.Fatal(err)
	}
	loans := b.Diplomacy.boonLoans
	if len(loans) != 1 || loans[0] != (BoonWorkerLoan{FactionKey: "merchant_guild", Count: 4, TicksLeft: 6}) {
		t.Fatalf("loaded boon loans = %+v, want one crew of 4 from merchant_guild with 6 ticks left", loans)
	}
	pop := b.Workers.TotalPop()
	b.StepTicks(5)
	if got := b.Workers.TotalPop(); got != pop {
		t.Fatalf("the crew left early after the load: population %d, want %d", got, pop)
	}
	b.StepTicks(1)
	if got := b.Workers.TotalPop(); got != pop-4 {
		t.Errorf("population after the loan ran out = %d, want %d", got, pop-4)
	}

	// A load replaces the loans held in memory: the save above had one, a
	// fresh save has none.
	c := newSeededEngine(2)
	if err := c.SaveGame("no-crew"); err != nil {
		t.Fatal(err)
	}
	lendBoonCrew(b, "merchant_guild", 3, 50)
	if err := b.LoadGame("no-crew"); err != nil {
		t.Fatal(err)
	}
	if n := len(b.Diplomacy.boonLoans); n != 0 {
		t.Errorf("loading a save with no boon crew kept %d loans from the game in memory", n)
	}
}
