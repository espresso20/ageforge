package game

import (
	"strings"
	"testing"
)

// The wonder's bank for a player who lingers. A player who stays in an age
// with full stores and does nothing is the case the storage wall exists for:
// production the stores cannot hold is lost, except what the age's wonder
// still needs, which goes into its bank while wonder overflow is on. These
// tests run the real engine's ticks (doTick), with the stores full from the
// start and an income into each of them, and check that
//
//   - the bank fills to the wonder's price and stops, never past it;
//   - once the wonder is built, overflow is simply lost;
//   - with wonder overflow off, nothing is banked.

// lingerEngine is a Primitive Age game with a town that out-produces its
// food and wood stores (the Sacred Grove wants 500 food and 1000 wood), every
// store full, and wonder overflow as asked. It returns the wonder's key.
func lingerEngine(t *testing.T, overflow bool) (*GameEngine, string) {
	t.Helper()
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 20
	ge.Buildings.counts["gathering_camp"] = 20
	ge.SetWonderOverflow(overflow)
	ge.recalculateRates()
	w := ge.progress.WonderForAge(ge.age)
	if w == "" || ge.Buildings.defs[w].Category != "wonder" {
		t.Fatalf("setup: the %s has no wonder (%q)", ge.age, w)
	}
	for res, c := range ge.Buildings.defs[w].BaseCost {
		if ge.Resources.GetRate(res) <= 0 {
			t.Fatalf("setup: no income of %s, which the %s costs %v of", res, w, c)
		}
		if c <= ge.Resources.GetStorage(res) {
			t.Fatalf("setup: a store of %v holds the whole %v %s the %s costs, so it never needs a bank", ge.Resources.GetStorage(res), c, res, w)
		}
	}
	for _, r := range ge.Resources.resources {
		r.Amount = r.Storage
	}
	return ge, w
}

// bankOf is a copy of wonder w's bank.
func bankOf(ge *GameEngine, w string) map[string]float64 {
	out := map[string]float64{}
	for res, v := range ge.Buildings.wonderBanks[w] {
		out[res] = v
	}
	return out
}

func TestWonderBank_FillsToThePriceAndStops(t *testing.T) {
	ge, w := lingerEngine(t, true)
	price := ge.Buildings.defs[w].BaseCost
	name := ge.Buildings.defs[w].Name

	prev := map[string]float64{}
	var fullAt int
	for tick := 1; tick <= 1500; tick++ {
		ge.doTick()
		bank := bankOf(ge, w)
		for res, need := range price {
			if bank[res] > need+1e-9 {
				t.Fatalf("tick %d: %s bank holds %v of a %v price: banked past it", tick, res, bank[res], need)
			}
			if bank[res] < prev[res] {
				t.Fatalf("tick %d: %s bank fell from %v to %v with nothing spent", tick, res, prev[res], bank[res])
			}
		}
		for res, r := range ge.Resources.resources {
			if r.Amount > r.Storage+1e-9 {
				t.Fatalf("tick %d: %s holds %v, over its cap of %v", tick, res, r.Amount, r.Storage)
			}
		}
		prev = bank
		if fullAt == 0 && ge.Buildings.IsWonderBankFull(w) {
			fullAt = tick
			break
		}
	}
	if fullAt == 0 {
		t.Fatalf("1500 ticks of full stores and an income left the %s bank at %v of %v", name, prev, price)
	}

	// Stopped: a long wait later the bank is exactly what it was, the stores
	// are still at their caps, and the bank's "full" line came once.
	full := bankOf(ge, w)
	for tick := 0; tick < 500; tick++ {
		ge.doTick()
	}
	after := bankOf(ge, w)
	for res := range price {
		if after[res] != full[res] {
			t.Errorf("%s bank moved from %v to %v after it was full", res, full[res], after[res])
		}
	}
	lines := 0
	for _, l := range ge.log {
		if strings.Contains(l.Message, "The "+name+" bank is full.") {
			lines++
		}
	}
	if lines != 1 {
		t.Errorf("the bank-is-full line came %d times, want once", lines)
	}
	for res, r := range ge.Resources.resources {
		if r.Amount > r.Storage+1e-9 {
			t.Errorf("%s holds %v, over its cap of %v", res, r.Amount, r.Storage)
		}
	}
}

func TestWonderBank_OnceBuiltOverflowIsLost(t *testing.T) {
	ge, w := lingerEngine(t, true)
	price := ge.Buildings.defs[w].BaseCost
	for tick := 0; tick < 1500 && !ge.Buildings.IsWonderBankFull(w); tick++ {
		ge.doTick()
	}
	if !ge.Buildings.IsWonderBankFull(w) {
		t.Fatalf("setup: the bank is %v of %v", bankOf(ge, w), price)
	}
	if err := ge.BuildBuilding(w); err != nil {
		t.Fatalf("setup: building the wonder from a full bank: %v", err)
	}
	for tick := 0; tick < 3000 && ge.Buildings.GetCount(w) == 0; tick++ {
		ge.doTick()
	}
	if ge.Buildings.GetCount(w) != 1 {
		t.Fatalf("setup: the wonder is not built (%d)", ge.Buildings.GetCount(w))
	}

	if got := ge.overflowWonder(); got != "" {
		t.Errorf("overflow still banks into %q after the wonder is built", got)
	}
	for _, r := range ge.Resources.resources {
		r.Amount = r.Storage
	}
	before := bankOf(ge, w)
	for tick := 0; tick < 300; tick++ {
		ge.doTick()
	}
	for res, v := range bankOf(ge, w) {
		if v != before[res] {
			t.Errorf("%s bank moved from %v to %v after the wonder was built", res, before[res], v)
		}
	}
	for res, r := range ge.Resources.resources {
		if r.Amount > r.Storage+1e-9 || r.Amount < 0 {
			t.Errorf("%s holds %v against a cap of %v", res, r.Amount, r.Storage)
		}
	}
}

func TestWonderBank_OverflowOffBanksNothing(t *testing.T) {
	ge, w := lingerEngine(t, false)
	for tick := 0; tick < 600; tick++ {
		ge.doTick()
	}
	for res, v := range ge.Buildings.wonderBanks[w] {
		if v != 0 {
			t.Errorf("wonder overflow is off, but the %s bank holds %v", res, v)
		}
	}
	if logHas(ge, "Overflow finished banking") {
		t.Error("wonder overflow is off, but a line says overflow banked")
	}
	for res, r := range ge.Resources.resources {
		if r.Amount > r.Storage+1e-9 {
			t.Errorf("%s holds %v, over its cap of %v", res, r.Amount, r.Storage)
		}
	}
	// Turned on, the very next ticks bank again.
	ge.SetWonderOverflow(true)
	for tick := 0; tick < 5; tick++ {
		ge.doTick()
	}
	if len(bankOf(ge, w)) == 0 {
		t.Error("turning wonder overflow on banked nothing")
	}
}
