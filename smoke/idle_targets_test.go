package smoke

import (
	"testing"
	"time"
)

// TestIdleTargetsAreSane: the idle targets are ordered by interval, each
// looser than the last, and none is tighter than the greedy pacing to the
// first prestige (a check-in player can't beat one who is always there).
func TestIdleTargetsAreSane(t *testing.T) {
	greedy := CumulativeTarget("modern_age")
	var prev IdleTarget
	for i, it := range IdleTargets {
		if it.CheckIn <= 0 || it.FirstPrestige <= 0 {
			t.Errorf("target %d is empty: %+v", i, it)
		}
		if it.FirstPrestige < greedy {
			t.Errorf("%s check-ins: target %s is under the greedy %s to the first prestige", it.CheckIn, it.FirstPrestige, greedy)
		}
		if i > 0 && (it.CheckIn <= prev.CheckIn || it.FirstPrestige < prev.FirstPrestige) {
			t.Errorf("targets out of order at %s", it.CheckIn)
		}
		prev = it
	}
	if IdleSeeds < 3 {
		t.Errorf("IdleSeeds = %d; the target is a median over at least 3 seeds", IdleSeeds)
	}
	if got := shortDur(90 * time.Minute); got != "90m" {
		t.Errorf("shortDur(90m) = %q", got)
	}
}
