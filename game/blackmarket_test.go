package game

import (
	"math"
	"math/rand"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// blackMarketSeedWin yields a first Float64 of 0.167 (< blackMarketWinChance,
// five in six) → a winning roll. blackMarketSeedLose yields 0.919 (>= five in
// six) → a loss.
func blackMarketSeedWin() *rand.Rand  { return rand.New(rand.NewSource(2)) }
func blackMarketSeedLose() *rand.Rand { return rand.New(rand.NewSource(7)) }

// bmEngine builds a colonial-age engine with a fixed culture income + balance so the
// deal cost is deterministic: 10000, the minutes of that income the deal asks.
func bmEngine(culture float64) *GameEngine {
	ge := NewGameEngine()
	ge.age = "colonial_age"
	learn(ge, "mercantilism", "drama") // the techs that open the black market and festivals
	if r, ok := ge.Resources.resources["culture"]; ok {
		// cost = max(5000, blackMarketCostShare of the age's stock scale in
		// minutes of this income) = 10000
		r.Rate = 10000 / float64(ge.rules.StockMinutes("culture", ge.age, blackMarketCostShare)*config.TicksPerMinute)
		r.Amount = culture
	}
	if r, ok := ge.Resources.resources["gold"]; ok {
		r.Amount = 0
	}
	return ge
}

// bmFloorPayout is what a win on res pays a town that makes under a quarter
// of the age's typical income of both culture and res (bmEngine's, which
// makes next to none of either): both incomes are then held to the quarter,
// and the stake's minutes of culture income come to winMult times that many
// minutes of res income, worked from the age's typical incomes alone.
func bmFloorPayout(ge *GameEngine, res string, stake float64) float64 {
	typC, typR := ge.rules.TypicalIncome("culture", ge.age), ge.rules.TypicalIncome(res, ge.age)
	if ge.incomeOf("culture") >= config.EventIncomeFloor*typC || ge.incomeOf(res) >= config.EventIncomeFloor*typR {
		panic("the town makes too much for the floor")
	}
	return blackMarketWinMult * stake * typR / typC
}

// TestBlackMarket_WinDeterministic: a winning roll deducts the culture cost and
// pays out gold at the configured multiplier.
func TestBlackMarket_WinDeterministic(t *testing.T) {
	ge := bmEngine(50000)
	ge.blackMarketRand = blackMarketSeedWin()

	wantCost := ge.BlackMarketStatus().Cost // 10000 with the fixed income
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
	// gold payout = blackMarketWinMult × the stake's minutes of culture income
	// in minutes of gold income; this town makes next to none of either.
	if want := bmFloorPayout(ge, "gold", wantCost); math.Abs(gain-want) > 1e-9*want {
		t.Errorf("gain = %.1f, want %.1f", gain, want)
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

// TestBlackMarket_GoesThroughFiveTimesInSix: over a long run of deals a deal
// goes through five times in six and pays the stake's worth; the sixth time
// the smugglers vanish with the culture and pay nothing. The status says so.
func TestBlackMarket_GoesThroughFiveTimesInSix(t *testing.T) {
	ge := bmEngine(1e12)
	ge.blackMarketRand = rand.New(rand.NewSource(11))
	const runs = 6000
	wins := 0
	for i := 0; i < runs; i++ {
		ge.blackMarketReadyTick = 0
		won, gain, err := ge.DoBlackMarket("gold")
		if err != nil {
			t.Fatalf("deal %d: %v", i, err)
		}
		if won != (gain > 0) {
			t.Fatalf("deal %d: won %v with a gain of %v, want a gain exactly when it went through", i, won, gain)
		}
		if won {
			wins++
		}
	}
	// Five in six of 6000 is 5000, with a spread of about 29: 150 either
	// side is five spreads.
	if wins < 4850 || wins > 5150 {
		t.Errorf("%d of %d deals went through, want about %d", wins, runs, runs*5/6)
	}
	st := ge.BlackMarketStatus()
	if st.WinChance != 5.0/6 || st.WinMult != 1 {
		t.Errorf("the status says %v and %vx, want five in six and the same worth (1x)", st.WinChance, st.WinMult)
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
	ge.tick += ge.stretchTicks(blackMarketCooldownTicks)
	ge.blackMarketRand = blackMarketSeedWin()
	if !ge.BlackMarketStatus().Ready {
		t.Errorf("black market should be ready after cooldown elapses")
	}
}
