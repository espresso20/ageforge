package game

import (
	"maps"
	"slices"
	"testing"
)

// TestResearchBonusesIndependentOfCompletionOrder: the smoke saveload
// scenario found an Iron Age reload with Research.Bonuses[food] at
// 0.7999999999999999 against the live 0.8, because live play summed effects
// in completion order and LoadState in sorted order. Bonuses must come out
// bit-identical whatever order the techs were finished in, and equal to what
// a load rebuilds.
func TestResearchBonusesIndependentOfCompletionOrder(t *testing.T) {
	complete := func(order []string) *ResearchManager {
		rm := NewResearchManager()
		for _, key := range order {
			rm.currentTech, rm.ticksLeft, rm.totalTicks = key, 1, 1
			if got := rm.Tick(); got != key {
				t.Fatalf("Tick completed %q, want %q", got, key)
			}
		}
		return rm
	}
	keys := NewResearchManager().order
	forward := complete(keys)
	rev := slices.Clone(keys)
	slices.Reverse(rev)
	backward := complete(rev)
	// A shuffled order too, fixed so the test is reproducible.
	shuffled := slices.Clone(keys)
	for i := range shuffled {
		j := (i*7 + 3) % len(shuffled)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	mixed := complete(shuffled)

	loaded := NewResearchManager()
	loaded.LoadState(forward.GetResearched(), "", 0, 0)

	for name, rm := range map[string]*ResearchManager{"backward": backward, "shuffled": mixed, "loaded": loaded} {
		if !maps.Equal(rm.bonuses, forward.bonuses) {
			for k, v := range forward.bonuses {
				if rm.bonuses[k] != v {
					t.Errorf("%s: bonus %s = %v, forward order gives %v", name, k, rm.bonuses[k], v)
				}
			}
		}
	}
}
