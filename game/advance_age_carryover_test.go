package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestAdvanceAge_CapsHoardedResource verifies that a player who over-accumulates
// a resource the new age uses has it capped to ~carryoverStarterBuildings of the
// cheapest new-age building's cost — NOT carried over as a flat percentage of the
// hoard. (EPIC: age-pacing economy rebalance, sub-ticket 2.)
func TestAdvanceAge_CapsHoardedResource(t *testing.T) {
	ge := NewGameEngine()

	// HUGE stockpile of a resource stone_age buildings use (wood + stone).
	const hoard = 1_000_000.0
	ge.Resources.LoadAmounts(map[string]float64{
		"wood":  hoard,
		"stone": hoard,
	})

	ge.advanceAge("stone_age")

	entry := config.AgeEntryCosts("stone_age")
	for _, res := range []string{"wood", "stone"} {
		cap := carryoverStarterBuildings * entry[res]
		got := ge.Resources.Get(res)
		if got != cap {
			t.Errorf("%s: hoard should be capped to %.2f (8×%.2f), got %.2f",
				res, cap, entry[res], got)
		}
		// Sanity: the cap must be far below the old flat-10%% carryover.
		if old := hoard * 0.10; got >= old {
			t.Errorf("%s: cap %.2f is not below old 10%% carryover %.2f — model not applied",
				res, got, old)
		}
	}
}

// TestAdvanceAge_PreservesModestStockpile verifies that a player whose stockpile
// is below the starter cap keeps what they had — it is NOT cut to a percentage.
func TestAdvanceAge_PreservesModestStockpile(t *testing.T) {
	ge := NewGameEngine()

	entry := config.AgeEntryCosts("stone_age")
	// Pick a modest amount well under the wood cap (8 × cheapest wood building).
	modest := entry["wood"] * 2 // two starter buildings' worth
	if modest <= 0 {
		t.Fatal("stone_age wood entry cost is zero — test cannot proceed")
	}
	ge.Resources.LoadAmounts(map[string]float64{"wood": modest})

	ge.advanceAge("stone_age")

	got := ge.Resources.Get("wood")
	if got != modest {
		t.Errorf("modest wood stockpile should be preserved: want %.2f, got %.2f", modest, got)
	}
}

// TestAdvanceAge_ResidualForUnusedResource verifies the fallback path: a resource
// that no new-age building uses is reduced to the small residual percentage.
func TestAdvanceAge_ResidualForUnusedResource(t *testing.T) {
	ge := NewGameEngine()

	entry := config.AgeEntryCosts("stone_age")
	// Find a resource that exists on the engine but is NOT in the stone_age entry
	// costs (no stone_age non-wonder building uses it).
	var unused string
	var startAmt float64
	for key := range ge.Resources.resources {
		if key == "faith" {
			continue
		}
		if _, used := entry[key]; !used {
			unused = key
			break
		}
	}
	if unused == "" {
		t.Skip("no unused-by-stone_age resource available to test residual path")
	}

	startAmt = 1000.0
	ge.Resources.LoadAmounts(map[string]float64{unused: startAmt})

	ge.advanceAge("stone_age")

	want := startAmt * carryoverResidualPct
	got := ge.Resources.Get(unused)
	if got != want {
		t.Errorf("unused resource %q should drop to residual %.2f (%.0f%%), got %.2f",
			unused, want, carryoverResidualPct*100, got)
	}
}

// TestAdvanceAge_FaithExcluded confirms faith is never reduced on age transition.
func TestAdvanceAge_FaithExcluded(t *testing.T) {
	ge := NewGameEngine()
	const faith = 500.0
	ge.Resources.LoadAmounts(map[string]float64{"faith": faith})

	ge.advanceAge("stone_age")

	if got := ge.Resources.Get("faith"); got != faith {
		t.Errorf("faith should be untouched on age transition: want %.2f, got %.2f", faith, got)
	}
}
