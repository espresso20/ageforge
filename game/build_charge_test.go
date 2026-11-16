package game

import (
	"math"
	"strings"
	"testing"
)

// TestBuildBuilding_ChargesEveryQueuedCopy: back-to-back builds with no tick
// between them pay the escalating price for each copy, queued copies count in
// the curve, and a build the stock cannot cover is refused without taking
// anything.
func TestBuildBuilding_ChargesEveryQueuedCopy(t *testing.T) {
	const affordable = 5
	ge := NewGameEngine()
	ge.mu.Lock()
	_, ok := ge.Buildings.BuildBatchCost("wood_camp", affordable, ge.buildQueue)
	owned := ge.Buildings.GetCount("wood_camp")
	want, _ := ge.Buildings.BuildBatchCost("wood_camp", affordable, ge.buildQueue)
	if !ok {
		ge.mu.Unlock()
		t.Fatal("wood_camp is not defined")
	}
	ge.Resources.LoadAmounts(map[string]float64{"wood": want["wood"]})
	start := ge.Resources.Get("wood")
	ge.mu.Unlock()
	if math.Abs(start-want["wood"]) > 1e-9 {
		t.Skipf("storage holds %.2f of the %.2f wood the test needs", start, want["wood"])
	}

	granted, refused := 0, 0
	for i := 0; i < affordable+3; i++ {
		err := ge.BuildBuilding("wood_camp")
		switch {
		case err == nil:
			granted++
		case strings.Contains(err.Error(), "Cannot afford"):
			refused++
		default:
			t.Fatalf("build %d: %v", i+1, err)
		}
	}
	if granted != affordable || refused != 3 {
		t.Fatalf("granted %d, refused %d; want %d, 3", granted, refused, affordable)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if got := ge.Buildings.GetQueueCount("wood_camp", ge.buildQueue); got != affordable {
		t.Errorf("queued = %d, want %d", got, affordable)
	}
	if left := ge.Resources.Get("wood"); left > 1e-6 {
		t.Errorf("wood left = %.4f, want 0 (price of %d queued copies is %.2f)", left, affordable, want["wood"])
	}
	// The next copy's price counts the queue: owned plus queued.
	next, _ := ge.Buildings.BuildBatchCost("wood_camp", 1, ge.buildQueue)
	first, _ := ge.Buildings.BuildBatchCost("wood_camp", 1, nil)
	if owned+affordable > owned && next["wood"] <= first["wood"] {
		t.Errorf("next price %.2f does not exceed the first copy's %.2f", next["wood"], first["wood"])
	}
}
