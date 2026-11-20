package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestTradeRoutes_ShortAndRunning: a route short of its export shows as
// short, one that has stock shows as running.
func TestTradeRoutes_ShortAndRunning(t *testing.T) {
	st := tradeState()
	st.Trade.ActiveRoutes[0].Starved = true
	st.Resources = map[string]game.ResourceState{"food": {Amount: 2}}
	st.Trade.ActiveRoutes[1].Disrupted = false
	st.Trade.ActiveRoutes[1].Starved = true
	st.Resources["stone"] = game.ResourceState{Amount: 500}
	state := map[string]string{}
	for _, r := range tradeRoutes(st) {
		state[r.key] = r.state
	}
	if state["local_barter"] != "short" {
		t.Errorf("local_barter = %q, want short", state["local_barter"])
	}
	if state["gold_caravan"] != "running" {
		t.Errorf("a restocked route = %q, want running", state["gold_caravan"])
	}
}

// TestTradeStartAll: the command starts every ready route and only those.
func TestTradeStartAll(t *testing.T) {
	_, _, eng := stagedDashboard(t, "colonial_age")
	ready := map[string]bool{}
	for _, r := range eng.GetState().Trade.AvailableRoutes {
		if r.CanStart {
			ready[r.Key] = true
		}
	}
	if len(ready) == 0 {
		t.Skip("no route is ready in this town")
	}
	active := map[string]bool{}
	for _, a := range eng.GetState().Trade.ActiveRoutes {
		active[a.Key] = true
	}
	if res := cmdTradeStartAll(eng); res.Type == "error" {
		t.Fatalf("start all: %v", res.Message)
	}
	for _, a := range eng.GetState().Trade.ActiveRoutes {
		if !active[a.Key] && !ready[a.Key] {
			t.Errorf("started %s, which was not ready", a.Key)
		}
		delete(ready, a.Key)
	}
	if len(ready) != 0 {
		t.Errorf("ready routes left unstarted: %v", ready)
	}
	if res := cmdTradeStartAll(eng); res.Type != "error" {
		t.Error("a second start all with nothing ready should fail")
	}
}

// TestTradeBoard_Fits: the board has no overlapping cards and nothing past the
// last content row, in an early, a middle and a late town.
func TestTradeBoard_Fits(t *testing.T) {
	for _, age := range layoutAges {
		for _, h := range []int{71, 68} {
			_, _, eng := stagedDashboard(t, age)
			st := eng.GetState()
			pg := tradePage(st)
			g, lay := renderTradeBoard(&pg, st, 0, 0, 279, h)
			if g == nil {
				t.Fatalf("%s %d: no grid", age, h)
			}
			if lay.viewH == 0 {
				for i, a := range lay.cards {
					if !a.Empty() && a.Max.Y > h-4 {
						t.Errorf("%s %d: card %d ends at row %d, past the content rows", age, h, i, a.Max.Y)
					}
					for j, b := range lay.cards[:i] {
						if a.Overlaps(b) {
							t.Errorf("%s %d: cards %d and %d overlap: %v %v", age, h, i, j, a, b)
						}
					}
				}
			}
		}
	}
}
