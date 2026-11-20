package game

import (
	"math"
	"testing"
)

// staffed is an engine with two gathering camps (six jobs), all filled, and
// waiting more people on top.
func staffed(t *testing.T, waiting int) *GameEngine {
	t.Helper()
	ge := sharesEngine(t, 10, map[string]int{"gathering_camp": 2})
	ge.Workers.Unlock()
	jobs := slotsOf(ge, "gathering_camp")
	if !ge.Workers.Recruit(jobs+waiting, ge.popCapLocked()) || !ge.Workers.Assign("gathering_camp", jobs) {
		t.Fatal("setup failed")
	}
	ge.recalculateRates()
	return ge
}

func TestPause_GrowthStopsWhenTheWaitingEqualThoseAtWork(t *testing.T) {
	jobs := slotsOf(sharesEngine(t, 10, map[string]int{"gathering_camp": 2}), "gathering_camp")
	for _, c := range []struct {
		waiting int
		grows   bool
	}{{jobs, false}, {jobs - 1, true}, {jobs + 3, false}} {
		ge := staffed(t, c.waiting)
		pop := ge.popLocked()
		ge.keepSharesLive()
		if grew := ge.popLocked() > pop; grew != c.grows {
			t.Errorf("%d waiting of %d at work: grew %v, want %v", c.waiting, jobs, grew, c.grows)
		}
		st := ge.GetState()
		want, wantHeld := RecruitActive, HeldNone
		if !c.grows {
			want, wantHeld = RecruitWork, HeldWork
		}
		if c.grows {
			continue // growing may stop at the next pass for other reasons
		}
		if got := RecruitStatus(st); got != want || st.Workers.HeldBy != wantHeld {
			t.Errorf("%d waiting: status %q held by %q, want %q and %q", c.waiting, got, st.Workers.HeldBy, want, wantHeld)
		}
	}
}

func TestPause_HandRecruitingStillWorks(t *testing.T) {
	ge := staffed(t, 6)
	pop := ge.popLocked()
	if err := ge.RecruitWorker("worker", 3); err != nil || ge.popLocked() != pop+3 {
		t.Fatalf("recruit by hand while paused: %v, population %d from %d", err, ge.popLocked(), pop)
	}
}

func TestPause_AFreshGameStillFillsItsFirstJobs(t *testing.T) {
	ge := sharesEngine(t, 3, map[string]int{"gathering_camp": 2})
	ge.keepSharesLive()
	if ge.popLocked() == 0 {
		t.Fatal("no first person")
	}
	growFully(ge)
	if got, want := ge.Workers.GetAssignedCount("gathering_camp"), slotsOf(ge, "gathering_camp"); got != want {
		t.Errorf("first jobs: %d of %d filled", got, want)
	}
}

func TestMovers_TheirSumIsWhatOneTickDoes(t *testing.T) {
	for _, waiting := range []int{0, 40} {
		ge := staffed(t, waiting)
		ge.morale = 0.62
		sum := 0.0
		for _, m := range ge.moraleMoversLocked() {
			sum += m.PerMinute
		}
		before := ge.morale
		ge.updateMoraleTick()
		if got := float64(ge.morale-before) * 30; math.Abs(got-sum) > 1e-9 {
			t.Errorf("%d waiting: movers sum %v a minute, one tick makes %v", waiting, sum, got)
		}
	}
}

func TestMovers_OnlyTermsThatAreNotZero(t *testing.T) {
	ge := staffed(t, 40)
	ge.morale = 0.5
	seen := map[string]bool{}
	for _, m := range ge.moraleMoversLocked() {
		if m.PerMinute == 0 {
			t.Errorf("%s listed at zero", m.Key)
		}
		seen[m.Key] = true
	}
	if !seen[MoverWaiting] || seen[MoverSettling] {
		t.Errorf("movers %v: want waiting and no settling at 50%%", seen)
	}
}

// One more worker: the figure equals the change in the rate from assigning
// one, for a food building and another, and is nothing when the building is
// full.
func TestOneMore_MatchesTheRateChange(t *testing.T) {
	ge := sharesEngine(t, 20, map[string]int{"gathering_camp": 2, "wood_camp": 2})
	ge.Workers.Unlock()
	ge.Workers.Recruit(20, ge.popCapLocked())
	ge.Workers.Assign("gathering_camp", 2)
	ge.Workers.Assign("wood_camp", 2)
	ge.recalculateRates()
	for _, c := range []struct{ key, res string }{{"gathering_camp", "food"}, {"wood_camp", "wood"}} {
		one := ge.GetState().Buildings[c.key].OneMore[c.res]
		if one <= 0 {
			t.Fatalf("%s: no gain listed for %s", c.key, c.res)
		}
		was := ge.Resources.GetRate(c.res)
		ge.Workers.Assign(c.key, 1)
		ge.recalculateRates()
		if got := ge.Resources.GetRate(c.res) - was; math.Abs(got-one) > 1e-9 {
			t.Errorf("%s: one more listed %v, the rate changed by %v", c.key, one, got)
		}
	}
	ge.Workers.Assign("gathering_camp", slotsOf(ge, "gathering_camp"))
	if m := ge.GetState().Buildings["gathering_camp"].OneMore; len(m) != 0 {
		t.Errorf("full building lists %v", m)
	}
}
