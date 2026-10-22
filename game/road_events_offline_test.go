package game

import (
	"reflect"
	"testing"
	"time"
)

// roadRun is what a long mission shows: the road events as they stand, the
// story lines they told, the mission's clock, odds and haul, the returns log
// and the town's gold.
type roadRun struct {
	active    bool
	events    []RoadEvent
	seen      []string
	ticksLeft int
	elapsed   int
	odds      float64
	success   map[string]float64
	failure   map[string]float64
	gold      float64
	returns   []MissionReturn
}

// roadMissionEngine is a seeded engine with a long mission sent, at least two
// of its road events due inside the first 60% of its length, and the length
// of that mission in ticks. Gold has no income here: production made while
// away runs at OfflineEfficiency, so with income the two engines could not
// end with the same gold by design, and gold would hide the toll check. The
// toll decisions are compared through the event stories instead.
func roadMissionEngine(t *testing.T) (*GameEngine, int) {
	t.Helper()
	for seed := int64(1); seed <= 200; seed++ {
		ge := missionReady(t, seed, "bronze_age")
		setIncome(ge, "gold", 0)
		// Gold sits at the town's real storage cap: a stock above it is clipped
		// by the first live tick's rate pass, which would be the test's doing.
		ge.mu.Lock()
		ge.recalculateRates()
		gold := ge.Resources.resources["gold"]
		gold.Amount = gold.Storage
		ge.mu.Unlock()
		if err := ge.LaunchExpedition(keyOffer(ge, "trade_caravan")); err != nil {
			continue
		}
		ge.mu.RLock()
		a := ge.Military.ActiveByTier(TierLong)
		total := a.TotalTicks
		due := 0
		for _, e := range a.RoadEvents {
			if e.Tick <= total*6/10 {
				due++
			}
		}
		ge.mu.RUnlock()
		if due >= 2 {
			return ge, total
		}
	}
	t.Fatal("no seed in 1..200 gives a long mission two road events in its first 60%")
	return nil, 0
}

func snapRoadRun(ge *GameEngine) roadRun {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	r := roadRun{gold: ge.Resources.Get("gold"), returns: append([]MissionReturn(nil), ge.Military.returns...)}
	a := ge.Military.ActiveByTier(TierLong)
	if a == nil {
		return r
	}
	r.active = true
	r.events = append([]RoadEvent(nil), a.RoadEvents...)
	r.seen = append([]string(nil), a.RoadEventsSeen...)
	r.ticksLeft = a.TicksLeft
	r.elapsed = a.ElapsedTicks
	r.odds = a.SuccessChance
	r.success = map[string]float64{}
	for k, v := range a.SuccessRewards {
		r.success[k] = v
	}
	r.failure = map[string]float64{}
	for k, v := range a.FailureRewards {
		r.failure[k] = v
	}
	return r
}

// TestRoadEventsFireOfflineLikeLive: the same mission, from the same seed and
// state, reaches the same road events at the same ticks whether the time is
// played live, tick by tick, or caught up offline in one go. The mission
// comes home in both, with the same returns entry.
func TestRoadEventsFireOfflineLikeLive(t *testing.T) {
	live, total := roadMissionEngine(t)
	offline, _ := roadMissionEngine(t)
	if before, after := snapRoadRun(live), snapRoadRun(offline); !reflect.DeepEqual(before, after) {
		t.Fatalf("the two engines did not start from the same mission:\n live %+v\n away %+v", before, after)
	}
	window := total * 6 / 10
	interval := live.getTickInterval()

	// Part one: partway through, the mission is still out and its road
	// events have been told, the same in both.
	for i := 0; i < window; i++ {
		live.doTick()
	}
	offTickBefore := offline.tick
	offline.SimulateOffline(time.Duration(window) * interval)
	if got := offline.tick - offTickBefore; got != window {
		t.Fatalf("offline catch-up credited %d ticks, want %d", got, window)
	}
	want, got := snapRoadRun(live), snapRoadRun(offline)
	if !want.active || !got.active {
		t.Fatalf("mission should still be out at tick %d of %d: live active %v, away active %v", window, total, want.active, got.active)
	}
	if len(want.seen) < 2 {
		t.Fatalf("live play told %d road events by tick %d, want at least 2", len(want.seen), window)
	}
	compareRoadRuns(t, "partway", want, got)

	// Part two: the rest of the way. The mission comes home in both.
	rest := total - window
	for i := 0; i < rest; i++ {
		live.doTick()
	}
	offTickBefore = offline.tick
	offline.SimulateOffline(time.Duration(rest) * interval)
	if got := offline.tick - offTickBefore; got != rest {
		t.Fatalf("offline catch-up credited %d ticks, want %d", got, rest)
	}
	want, got = snapRoadRun(live), snapRoadRun(offline)
	if want.active || got.active {
		t.Fatalf("mission should have come home by tick %d: live active %v, away active %v", total, want.active, got.active)
	}
	if len(want.returns) != 1 {
		t.Fatalf("live play wrote %d returns, want 1", len(want.returns))
	}
	compareRoadRuns(t, "home", want, got)
}

// compareRoadRuns reports every way the offline run differs from the live one.
func compareRoadRuns(t *testing.T, when string, want, got roadRun) {
	t.Helper()
	if !reflect.DeepEqual(got.events, want.events) {
		t.Errorf("%s: road events differ offline:\n live %+v\n away %+v", when, want.events, got.events)
	}
	if !reflect.DeepEqual(got.seen, want.seen) {
		t.Errorf("%s: road event stories differ offline:\n live %q\n away %q", when, want.seen, got.seen)
	}
	if got.ticksLeft != want.ticksLeft {
		t.Errorf("%s: ticks left offline %d, live %d", when, got.ticksLeft, want.ticksLeft)
	}
	if got.elapsed != want.elapsed {
		t.Errorf("%s: elapsed ticks offline %d, live %d", when, got.elapsed, want.elapsed)
	}
	if got.odds != want.odds {
		t.Errorf("%s: odds offline %v, live %v", when, got.odds, want.odds)
	}
	if !reflect.DeepEqual(got.success, want.success) || !reflect.DeepEqual(got.failure, want.failure) {
		t.Errorf("%s: haul offline success %v failure %v, live success %v failure %v", when, got.success, got.failure, want.success, want.failure)
	}
	if !reflect.DeepEqual(got.returns, want.returns) {
		t.Errorf("%s: returns differ offline:\n live %+v\n away %+v", when, want.returns, got.returns)
	}
	if got.gold != want.gold {
		t.Errorf("%s: gold offline %v, live %v", when, got.gold, want.gold)
	}
}

// TestMissionLaunchedBeforeAwayComesHomeWhileAway: a mission sent before the
// player leaves comes home during the time away, at its own tick, and its
// returns entry is written then.
func TestMissionLaunchedBeforeAwayComesHomeWhileAway(t *testing.T) {
	ge, total := roadMissionEngine(t)
	ge.mu.RLock()
	launched := ge.tick
	ticksLeft := ge.Military.ActiveByTier(TierLong).TicksLeft
	ge.mu.RUnlock()
	if ticksLeft != total {
		t.Fatalf("mission has %d ticks left, want its full %d", ticksLeft, total)
	}

	ge.SimulateOffline(time.Duration(total+10) * ge.getTickInterval())

	ge.mu.RLock()
	defer ge.mu.RUnlock()
	if ge.Military.ActiveByTier(TierLong) != nil {
		t.Fatal("mission is still out after being away longer than it runs")
	}
	if len(ge.Military.returns) != 1 {
		t.Fatalf("returns log has %d entries, want 1", len(ge.Military.returns))
	}
	// A road event can shorten or lengthen the mission, so its own tick is
	// not always launch plus length; it must still be inside the time away.
	if back := ge.Military.returns[0].Tick; back <= launched || back >= ge.tick {
		t.Fatalf("mission came home at tick %d, outside the time away (launched %d, now %d)", back, launched, ge.tick)
	}
}
