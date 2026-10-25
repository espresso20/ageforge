package game

import "testing"

// TestLogRing_QueuedBuildsDoNotPushOutHistory: the log ring holds the last
// MaxLogSize lines and is the one the player reads, so a line written every
// tick eats it. The queue summary and one line per queued building used to
// be written as "debug" lines each tick, and a few long builds pushed the
// real history out of the ring within a minute or two.
func TestLogRing_QueuedBuildsDoNotPushOutHistory(t *testing.T) {
	isolateAccountDir(t)
	ge := NewGameEngine()
	var first LogEntry
	for _, e := range ge.GetLogs() {
		if e.Type != "debug" {
			first = e
			break
		}
	}
	if first.Message == "" {
		t.Fatal("no real line was logged before the queue")
	}

	ge.mu.Lock()
	for _, r := range []string{"wood", "food", "stone"} {
		ge.Resources.Add(r, 500_000)
	}
	// Builds that outlast the run, so the queue is busy on every tick.
	def := ge.Buildings.defs["shrine"]
	def.BuildTicks = 700
	ge.Buildings.defs["shrine"] = def
	ge.mu.Unlock()
	for i := 0; i < 3; i++ {
		if err := ge.BuildBuilding("shrine"); err != nil {
			t.Fatalf("shrine copy %d: %v", i+1, err)
		}
	}
	for i := 0; i < 600; i++ {
		ge.doTick()
	}

	if n := len(ge.GetState().BuildQueue); n != 3 {
		t.Fatalf("%d builds still queued after 600 ticks, want 3 (the test needs a busy queue)", n)
	}
	for _, e := range ge.GetLogs() {
		if e.Type == first.Type && e.Tick == first.Tick && e.Message == first.Message {
			return
		}
	}
	t.Errorf("the first real line %q was pushed out of the %d-line log ring by 600 ticks of queued builds", first.Message, MaxLogSize)
}
