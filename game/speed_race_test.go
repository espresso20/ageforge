package game

import (
	"testing"
	"time"
)

// TestSpeedWritesVsTickLoop exercises the tick loop's getTickInterval read of
// speedMultiplier/tickSpeedBonus while the /speed dev command writes them from
// another goroutine. It only fails under -race (the read used to be
// unlocked); without -race it is a smoke test that the loop keeps ticking and
// shuts down cleanly.
func TestSpeedWritesVsTickLoop(t *testing.T) {
	isolateAccountDir(t)
	prevDev := DevModeActive
	DevModeActive = true
	t.Cleanup(func() { DevModeActive = prevDev })

	ge := NewGameEngine()
	DevExecCommand("/speed 10", ge) // interval clamps to MinTickInterval from the first tick
	done := make(chan struct{})
	go func() {
		ge.Start()
		close(done)
	}()

	// Wait for the first tick, so the writer loop below cannot slip a 1.0x in
	// before Start arms its first timer (that would push tick one out to 2s).
	firstTick := time.Now().Add(3 * time.Second)
	for ge.GetState().Tick == 0 {
		if time.Now().After(firstTick) {
			ge.Stop()
			t.Fatal("tick loop never ticked")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Every tick re-arms its timer via getTickInterval while these writes land.
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		// Down to 1x, then straight back to 10x so the loop keeps ticking at
		// the fast interval.
		DevExecCommand("/speed 1", ge)
		DevExecCommand("/speed 10", ge)
		time.Sleep(5 * time.Millisecond)
	}
	ge.Stop()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("tick loop did not stop")
	}
}
