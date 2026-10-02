package game

import "testing"

// leakRunTimers sets the run timers a long run can end with: ready to
// advance, starving, a survey party counting down, a dispatch held up.
func leakRunTimers(ge *GameEngine) {
	ge.ageReady = true
	ge.starvationTicks = 7
	ge.autoExpeditionTicksLeft = 120
	ge.autoExpeditionStarved = true
}

// checkRunTimersFresh fails unless ge's run timers are a new game's, and
// `advance` and the first tick behave as in one.
func checkRunTimersFresh(t *testing.T, ge *GameEngine, after string) {
	t.Helper()
	fresh := NewGameEngine()
	ge.mu.RLock()
	got := [4]any{ge.ageReady, ge.starvationTicks, ge.autoExpeditionTicksLeft, ge.autoExpeditionStarved}
	ge.mu.RUnlock()
	want := [4]any{fresh.ageReady, fresh.starvationTicks, fresh.autoExpeditionTicksLeft, fresh.autoExpeditionStarved}
	if got != want {
		t.Errorf("after %s: ageReady, starvationTicks, autoExpeditionTicksLeft, autoExpeditionStarved = %v, want %v (a new game's)", after, got, want)
	}
	if ge.GetState().AgeReady {
		t.Errorf("after %s the snapshot says the Primitive Age is ready to advance", after)
	}
	if err := ge.AdvanceAge(); err == nil {
		t.Errorf("after %s, advance went through to %s without the Stone Age's requirements", after, ge.age)
	}
	ge.StepTicks(1)
	if e, ok := lastLogWith(ge, "Food is back"); ok {
		t.Errorf("after %s the first tick logged %q: the old run's starvation leaked in", after, e.Message)
	}
}

// TestPrestigeResetsRunTimers: prestige reset the run but not its timers, so
// a run that ended ready to advance let `advance` skip the Stone Age's
// requirements until the first tick, one that ended starving logged "Food is
// back" into the new run, and the old dispatch countdown showed until the
// first tick.
func TestPrestigeResetsRunTimers(t *testing.T) {
	ge := newSeededEngine(1)
	ge.mu.Lock()
	ge.age = "modern_age"
	leakRunTimers(ge)
	ge.completePrestige(prestigePlain)
	ge.mu.Unlock()
	checkRunTimersFresh(t, ge, "prestige")
}

// TestResetResetsRunTimers: a new game on the same engine (Reset, as the
// splash's New game and a wipe use) starts the timers over too.
func TestResetResetsRunTimers(t *testing.T) {
	ge := newSeededEngine(1)
	ge.mu.Lock()
	leakRunTimers(ge)
	ge.mu.Unlock()
	ge.Reset()
	checkRunTimersFresh(t, ge, "a new game")
}
