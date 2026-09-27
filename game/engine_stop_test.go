package game

import (
	"testing"
	"time"
)

// startedLoop runs ge.Start on its own goroutine and returns the loop's done
// channel once the loop has registered it.
func startedLoop(t *testing.T, ge *GameEngine) chan struct{} {
	t.Helper()
	go ge.Start()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ge.mu.RLock()
		done := ge.loopDone
		ge.mu.RUnlock()
		if done != nil {
			return done
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("Start never registered its loop")
	return nil
}

// Stop must not return while the tick goroutine is still running: callers
// (the UI harness teardown, the ESC handler) rely on nothing of the loop's
// being in flight afterwards.
func TestStop_WaitsForTickLoop(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	done := startedLoop(t, ge)

	ge.Stop()
	select {
	case <-done:
	default:
		t.Fatal("Stop returned while the Start loop was still running")
	}

	// A second Stop is a no-op and must not block.
	returned := make(chan struct{})
	go func() { ge.Stop(); close(returned) }()
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		t.Fatal("second Stop blocked")
	}
}

// Stop on an engine whose loop never started returns at once.
func TestStop_BeforeStartDoesNotBlock(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	returned := make(chan struct{})
	go func() { ge.Stop(); close(returned) }()
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop blocked on an engine that never started")
	}
}

// ESC → splash → New Game restarts the loop on the same engine; the second
// Stop must wait for the second loop.
func TestStop_WaitsAfterRestart(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	first := startedLoop(t, ge)
	ge.Stop()
	<-first

	go ge.Start()
	var second chan struct{}
	deadline := time.Now().Add(5 * time.Second)
	for second == nil || second == first {
		if time.Now().After(deadline) {
			t.Fatal("restarted Start never registered its loop")
		}
		ge.mu.RLock()
		second = ge.loopDone
		ge.mu.RUnlock()
		time.Sleep(time.Millisecond)
	}
	ge.Stop()
	select {
	case <-second:
	default:
		t.Fatal("Stop returned while the restarted loop was still running")
	}
}
