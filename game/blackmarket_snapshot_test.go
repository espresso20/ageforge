package game

import (
	"math"
	"reflect"
	"testing"
)

// The snapshot's payout for a resource is what a winning run on it pays.
func TestBlackMarketSnapshotPayoutsMatchAWinningRun(t *testing.T) {
	probe := bmEngine(50000)
	pays := probe.GetState().BlackMarket.Payouts
	if len(pays) == 0 {
		t.Fatal("no payouts in the Colonial Age")
	}
	for _, p := range pays {
		ge := bmEngine(50000)
		ge.blackMarketRand = blackMarketSeedWin()
		won, gain, err := ge.DoBlackMarket(p.Resource)
		if err != nil || !won {
			t.Fatalf("%s: won=%v err=%v", p.Resource, won, err)
		}
		if math.Abs(gain-p.Amount) > 1e-9*p.Amount {
			t.Errorf("%s: snapshot says %v, a win paid %v", p.Resource, p.Amount, gain)
		}
	}
}

// The payouts follow the rules' resource order, one line per resource the
// smugglers deal in.
func TestBlackMarketSnapshotPayoutsAreInResourceOrder(t *testing.T) {
	ge := bmEngine(50000)
	pays := ge.GetState().BlackMarket.Payouts
	at := map[string]int{}
	for i, def := range ge.rules.Resources() {
		at[def.Key] = i
	}
	for i := 1; i < len(pays); i++ {
		if at[pays[i-1].Resource] >= at[pays[i].Resource] {
			t.Errorf("%s comes before %s", pays[i-1].Resource, pays[i].Resource)
		}
	}
	seen := map[string]bool{}
	for _, p := range pays {
		if seen[p.Resource] || !blackMarketDeals[p.Resource] || p.Amount <= 0 {
			t.Errorf("bad payout line %+v", p)
		}
		seen[p.Resource] = true
	}
}

// Before the Colonial Age the list is empty and the market is not available.
func TestBlackMarketSnapshotIsEmptyBeforeTheColonialAge(t *testing.T) {
	st := NewGameEngine().GetState().BlackMarket
	if st.Available || len(st.Payouts) != 0 {
		t.Errorf("snapshot before the Colonial Age: %+v", st)
	}
}

// The snapshot and the public method are one status.
func TestBlackMarketSnapshotAgreesWithTheStatusMethod(t *testing.T) {
	ge := bmEngine(50000)
	if a, b := ge.GetState().BlackMarket, ge.BlackMarketStatus(); !reflect.DeepEqual(a, b) {
		t.Errorf("snapshot %+v, method %+v", a, b)
	}
	ge.blackMarketRand = blackMarketSeedWin()
	if _, _, err := ge.DoBlackMarket("gold"); err != nil {
		t.Fatal(err)
	}
	if a, b := ge.GetState().BlackMarket, ge.BlackMarketStatus(); !reflect.DeepEqual(a, b) || a.CooldownLeft == 0 {
		t.Errorf("after a run: snapshot %+v, method %+v", a, b)
	}
}
