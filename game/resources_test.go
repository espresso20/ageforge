package game

import (
	"testing"
)

func TestResourceManager_AddAndGet(t *testing.T) {
	rm := NewResourceManager()
	rm.UnlockResource("food")
	rm.UnlockResource("wood")

	rm.Add("food", 20)
	if got := rm.Get("food"); got != 20 {
		t.Errorf("Get(food) = %v, want 20", got)
	}

	// Adding more (within storage)
	rm.Add("food", 10)
	if got := rm.Get("food"); got != 30 {
		t.Errorf("Get(food) after second add = %v, want 30", got)
	}

	// Unlocked resource at zero
	if got := rm.Get("wood"); got != 0 {
		t.Errorf("Get(wood) = %v, want 0", got)
	}
}

// Stock has no ceiling: a credit past the base store is kept whole, a
// production tick adds to it, and an offline credit does too.
func TestResourceManager_StockIsUncapped(t *testing.T) {
	rm := NewResourceManager()
	rm.UnlockResource("food")

	// Food once had a store of 50; adding 999 keeps all 999.
	rm.Add("food", 999)
	if got := rm.Get("food"); got != 999 {
		t.Errorf("food = %v after adding 999, want all 999", got)
	}
	rm.SetRate("food", 7)
	rm.ApplyRates()
	if got := rm.Get("food"); got != 1006 {
		t.Errorf("food = %v after a tick at +7 from 999, want 1006", got)
	}
	gained := 0.0
	rm.AddProduced(100, func(_ string, g float64) { gained += g })
	if got := rm.Get("food"); got != 1706 || gained != 700 {
		t.Errorf("food = %v, gained %v after 100 ticks away at +7, want 1706 and 700", got, gained)
	}
	// Spending still works down to nothing, and stock never goes below it.
	rm.Add("food", -5000)
	if got := rm.Get("food"); got != 0 {
		t.Errorf("food = %v after a debit past zero, want 0", got)
	}
}

func TestResourceManager_Remove(t *testing.T) {
	rm := NewResourceManager()
	rm.UnlockResource("food")
	rm.Add("food", 40)

	ok := rm.Remove("food", 15)
	if !ok {
		t.Error("Remove(15) should succeed with 40 available")
	}
	if got := rm.Get("food"); got != 25 {
		t.Errorf("food after remove = %v, want 25", got)
	}

	// Remove more than available
	ok = rm.Remove("food", 100)
	if ok {
		t.Error("Remove(100) should fail with only 25 available")
	}
}

func TestResourceManager_PayAndCanAfford(t *testing.T) {
	rm := NewResourceManager()
	rm.UnlockResource("food")
	rm.UnlockResource("wood")
	rm.Add("food", 40)
	rm.Add("wood", 30)

	costs := map[string]float64{"food": 10, "wood": 5}
	if !rm.CanAfford(costs) {
		t.Error("CanAfford should be true")
	}

	if !rm.Pay(costs) {
		t.Error("Pay should succeed")
	}
	if got := rm.Get("food"); got != 30 {
		t.Errorf("food after pay = %v, want 30", got)
	}
	if got := rm.Get("wood"); got != 25 {
		t.Errorf("wood after pay = %v, want 25", got)
	}

	// Can't afford now
	bigCosts := map[string]float64{"food": 200}
	if rm.CanAfford(bigCosts) {
		t.Error("CanAfford should be false for 200 food")
	}
	if rm.Pay(bigCosts) {
		t.Error("Pay should fail for unaffordable costs")
	}
	// Ensure atomic — food unchanged after failed pay
	if got := rm.Get("food"); got != 30 {
		t.Errorf("food should be unchanged after failed pay, got %v", got)
	}
}

func TestResourceManager_ApplyRates(t *testing.T) {
	rm := NewResourceManager()
	rm.UnlockResource("food")
	rm.Add("food", 20)
	rm.SetRate("food", 5.0)

	rm.ApplyRates()
	if got := rm.Get("food"); got != 25 {
		t.Errorf("food after ApplyRates = %v, want 25", got)
	}

	// Negative rate
	rm.SetRate("food", -10.0)
	rm.ApplyRates()
	if got := rm.Get("food"); got != 15 {
		t.Errorf("food after negative rate = %v, want 15", got)
	}
}

func TestResourceManager_UnlockedState(t *testing.T) {
	rm := NewResourceManager()
	if rm.IsUnlocked("food") {
		t.Error("food should not be unlocked initially")
	}
	rm.UnlockResource("food")
	if !rm.IsUnlocked("food") {
		t.Error("food should be unlocked after UnlockResource")
	}
}

func TestResourceManager_SaveLoadRoundTrip(t *testing.T) {
	rm := NewResourceManager()
	rm.UnlockResource("food")
	rm.UnlockResource("wood")
	rm.Add("food", 33.5)
	rm.Add("wood", 22.2)

	// Save
	amounts := rm.GetAll()

	// Load into fresh manager
	rm2 := NewResourceManager()
	rm2.UnlockResource("food")
	rm2.UnlockResource("wood")
	rm2.LoadAmounts(amounts)

	if got := rm2.Get("food"); got != 33.5 {
		t.Errorf("loaded food = %v, want 33.5", got)
	}
	if got := rm2.Get("wood"); got != 22.2 {
		t.Errorf("loaded wood = %v, want 22.2", got)
	}
}

// SetRate sets the production rate for a resource
func (rm *ResourceManager) SetRate(key string, rate float64) {
	if r, ok := rm.resources[key]; ok {
		r.Rate = rate
	}
}
