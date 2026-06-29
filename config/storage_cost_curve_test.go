package config

import (
	"math"
	"testing"
)

// storageCapPerCopy returns the per-copy storage cap a storage building adds
// (its "storage" Effect value), and whether such an effect exists.
func storageCapPerCopy(d BuildingDef) (float64, bool) {
	for _, e := range d.Effects {
		if e.Type == "storage" {
			return e.Value, true
		}
	}
	return 0, false
}

// dominantBaseCost returns the largest single-resource cost in BaseCost. Because
// storage buildings raise the cap of *every* resource equally (Effect.Target ==
// "all"), the binding affordability constraint is always the most expensive
// resource: if you can save up for that one, you can save up for the rest.
func dominantBaseCost(d BuildingDef) float64 {
	var max float64
	for _, v := range d.BaseCost {
		if v > max {
			max = v
		}
	}
	return max
}

// TestStorage_CostNeverWallsAboveCapProvided enforces the storage-lineage design
// invariant for EVERY storage building at its runtime (post-normalizeCostCurves)
// values: the cost of the N-th copy must never exceed the storage that N copies
// provide, so a player can always save up for the next one and a storage building
// never becomes an unaffordable wall.
//
// Literal, self-contained form (per the balance card):
//
//	cost(N) = dominantBaseCost × CostScale^(N-1)  ≤  N × capPerCopy
//
// for every N from 1 to MaxCount. This generalizes the single-building
// TestStash_FirstCopyAffordableWithinBaseWoodCap guard to the whole lineage so a
// future re-tune of any storage cost, cap, or MaxCount re-surfaces the regression
// loudly. (This is the same class of bug as the old stash deadlock.)
func TestStorage_CostNeverWallsAboveCapProvided(t *testing.T) {
	// If a storage building is unlimited (MaxCount 0) or absurdly high, we can't
	// loop forever — scan to this bound and additionally require the curve to stay
	// affordable for at least minUnlimitedStack copies (a meaningful stack).
	const (
		unlimitedScanBound = 200
		minUnlimitedStack  = 20
	)

	for key, d := range BuildingByKey() {
		if d.Category != "storage" {
			continue
		}

		cap, ok := storageCapPerCopy(d)
		if !ok || cap <= 0 {
			t.Errorf("storage building %q has no positive storage Effect — cannot provide cap", key)
			continue
		}
		dom := dominantBaseCost(d)
		if dom <= 0 {
			t.Errorf("storage building %q has no positive build cost", key)
			continue
		}

		// Decide how far to walk the cost curve.
		limit := d.MaxCount
		unlimited := d.MaxCount == 0 || d.MaxCount > unlimitedScanBound
		if unlimited {
			limit = unlimitedScanBound
		}

		// Walk every copy 1..limit and assert cost(N) <= N*cap.
		firstWall := -1
		for n := 1; n <= limit; n++ {
			cost := dom * math.Pow(d.CostScale, float64(n-1))
			capProvided := float64(n) * cap
			if cost > capProvided {
				firstWall = n
				break
			}
		}

		if firstWall != -1 {
			if unlimited {
				// Unlimited building: a wall anywhere short of a meaningful stack is
				// a softlock waiting to happen. (Walls >= minUnlimitedStack are still
				// unreachable in practice but cap MaxCount to be safe — see below.)
				if firstWall < minUnlimitedStack {
					t.Errorf("storage %q (MaxCount=0/unlimited) walls at copy %d: "+
						"cost %.0f > %d × capPerCopy %.0f — set MaxCount below the wall, "+
						"lower BaseCost, or raise the cap",
						key, firstWall, dom*math.Pow(d.CostScale, float64(firstWall-1)), firstWall, cap)
				} else {
					// Affordable for a meaningful stack but technically unbounded —
					// the design now caps every storage building, so flag the gap.
					t.Errorf("storage %q has MaxCount=0 (unlimited) and would wall at "+
						"copy %d — set an explicit MaxCount below the wall so the building "+
						"cannot be built into the unaffordable zone", key, firstWall)
				}
			} else {
				// Bounded building must stay affordable across its ENTIRE allowed range.
				t.Errorf("storage %q walls at copy %d of MaxCount %d: cost %.0f > %d × "+
					"capPerCopy %.0f — lower MaxCount below the wall, lower BaseCost, or "+
					"raise the cap",
					key, firstWall, d.MaxCount,
					dom*math.Pow(d.CostScale, float64(firstWall-1)), firstWall, cap)
			}
		}
	}
}
