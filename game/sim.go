package game

import "time"

// StepTicks runs n game ticks synchronously on the calling goroutine, as fast
// as the CPU allows, and returns the wall-clock time those ticks would have
// taken in the real timer-driven loop (the sum of each tick's interval at the
// current speed, tick_speed bonuses included).
//
// It exists for simulation and tests (the smoke autoplayer in package smoke),
// not for the UI: the game itself runs on Start's timer. Do not call it while
// Start is running on the same engine. Unlike Start it does not autosave and
// does not recover panics, so a crash inside a tick reaches the caller.
func (ge *GameEngine) StepTicks(n int) time.Duration {
	var elapsed time.Duration
	for i := 0; i < n; i++ {
		elapsed += ge.getTickInterval()
		ge.doTick()
	}
	return elapsed
}
