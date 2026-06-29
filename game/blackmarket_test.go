package game

import (
	"math"
	"math/rand"
	"testing"
)

// blackMarketSeedWin yields a first Float64 of 0.167 (< blackMarketWinChance
// 0.55) → a winning roll. blackMarketSeedLose yields 0.605 (>= 0.55) → a loss.
func blackMarketSeedWin() *rand.Rand  { return rand.New(rand.NewSource(2)) }
func blackMarketSeedLose() *rand.Rand { return rand.New(rand.NewSource(1)) }

// bmEngine builds a colonial-age engine with a fixed culture cap + balance so the
// deal cost is deterministic, and a gold sink with headroom for the payout.
func bmEngine(culture float64) *GameEngine {
	ge := NewGameEngine()
	ge.age = "colonial_age"
	if r, ok := ge.Resources.resources["culture"]; ok {
		r.Storage = 100000 // cost = max(5000, 0.10×100000) = 10000
		r.Amount = culture
	}
	if r, ok := ge.Resources.resources["gold"]; ok {
		r.Storage = 1e12
		r.Amount = 0
	}
	return ge
}

// TestBlackMarket_WinDeterministic: a winning roll deducts the culture cost and
// pays out gold at the configured multiplier.
func TestBlackMarket_WinDeterministic(t *testing.T) {
	ge := bmEngine(50000)
	ge.blackMarketRand = blackMarketSeedWin()

	wantCost := ge.BlackMarketStatus().Cost // 10000 with the fixed cap
	if math.Abs(wantCost-10000) > 1e-6 {
		t.Fatalf("setup: cost = %.1f, want 10000", wantCost)
	}

	won, gain, err := ge.DoBlackMarket("gold")
	if err != nil {
		t.Fatalf("DoBlackMarket: %v", err)
	}
	if !won {
		t.Fatalf("expected a win with the winning seed")
	}
	// gold payout = cost × blackMarketWinMult.
	if math.Abs(gain-wantCost*blackMarketWinMult) > 1e-6 {
		t.Errorf("gain = %.1f, want %.1f", gain, wantCost*blackMarketWinMult)
	}
	if got := ge.Resources.Get("gold"); math.Abs(got-gain) > 1e-6 {
		t.Errorf("gold balance = %.1f, want %.1f", got, gain)
	}
	if got := ge.Resources.Get("culture"); math.Abs(got-(50000-wantCost)) > 1e-6 {
		t.Errorf("culture after = %.1f, want %.1f", got, 50000-wantCost)
	}
}

// TestBlackMarket_LoseDeterministic: a losing roll still spends the culture but
// pays out nothing.
func TestBlackMarket_LoseDeterministic(t *testing.T) {
	ge := bmEngine(50000)
	ge.blackMarketRand = blackMarketSeedLose()

	cost := ge.BlackMarketStatus().Cost
	won, gain, err := ge.DoBlackMarket("gold")
	if err != nil {
		t.Fatalf("DoBlackMarket: %v", err)
	}
	if won {
		t.Fatalf("expected a loss with the losing seed")
	}
	if gain != 0 {
		t.Errorf("loss gain = %.1f, want 0", gain)
	}
	if got := ge.Resources.Get("gold"); got != 0 {
		t.Errorf("gold after loss = %.1f, want 0", got)
	}
	if got := ge.Resources.Get("culture"); math.Abs(got-(50000-cost)) > 1e-6 {
		t.Errorf("culture after loss = %.1f, want %.1f (cost still spent)", got, 50000-cost)
	}
}

// TestBlackMarket_RefusesWhenBroke: too little culture → error, nothing spent.
func TestBlackMarket_RefusesWhenBroke(t *testing.T) {
	ge := bmEngine(100) // cost is 10000; far short
	ge.blackMarketRand = blackMarketSeedWin()

	_, _, err := ge.DoBlackMarket("gold")
	if err == nil {
		t.Fatalf("expected an error when culture is insufficient")
	}
	if got := ge.Resources.Get("culture"); got != 100 {
		t.Errorf("culture after refusal = %.1f, want 100 (unchanged)", got)
	}
}

// TestBlackMarket_GatedByAge: below colonial age the deal is unavailable.
func TestBlackMarket_GatedByAge(t *testing.T) {
	ge := bmEngine(50000)
	ge.age = "medieval_age" // before colonial
	if ge.BlackMarketStatus().Available {
		t.Errorf("black market should be unavailable before colonial age")
	}
	if _, _, err := ge.DoBlackMarket("gold"); err == nil {
		t.Errorf("expected an error attempting a deal before colonial age")
	}
}

// TestBlackMarket_Cooldown: a successful deal arms a cooldown that blocks the
// next deal until it expires.
func TestBlackMarket_Cooldown(t *testing.T) {
	ge := bmEngine(50000)
	ge.blackMarketRand = blackMarketSeedWin()

	if _, _, err := ge.DoBlackMarket("gold"); err != nil {
		t.Fatalf("first deal: %v", err)
	}
	// Immediately retry — must be on cooldown.
	if _, _, err := ge.DoBlackMarket("gold"); err == nil {
		t.Errorf("second deal should be blocked by cooldown")
	}
	// Advance past the cooldown → ready again (still need a fresh seed; reuse win).
	ge.tick += blackMarketCooldownTicks
	ge.blackMarketRand = blackMarketSeedWin()
	if !ge.BlackMarketStatus().Ready {
		t.Errorf("black market should be ready after cooldown elapses")
	}
}
