package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestTradeRoutes_ShortAndRunning: a route short of its export shows as
// short, one that has stock shows as running.
func TestTradeRoutes_ShortAndRunning(t *testing.T) {
	st := tradeState()
	st.Trade.ActiveRoutes[0].Run = 0.4
	st.Resources = map[string]game.ResourceState{"food": {Amount: 2}}
	st.Trade.ActiveRoutes[1].Disrupted = false
	st.Trade.ActiveRoutes[1].Run = 1
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
			for _, fd := range eng.Rules().Factions() {
				_ = eng.MeetFactionForTest(fd.Key, 60)
			}
			eng.StepTicks(3)
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

// offerTown is a Colonial town that has met two civilizations: one with a
// payable and an unpayable offer, one under embargo.
func offerTown(t *testing.T) game.GameState {
	t.Helper()
	_, _, eng := stagedDashboard(t, "colonial_age")
	st := eng.GetState()
	fs := st.Ruleset().Factions()
	st.Resources["food"] = game.ResourceState{Amount: 100}
	a, b := st.Diplomacy.Factions[fs[0].Key], st.Diplomacy.Factions[fs[1].Key]
	a.Discovered, a.Status, a.Name = true, "neutral", "Alpha"
	a.Deals = []game.DealInfo{{Num: 1, Kind: game.DealWant, Give: "food", GiveAmt: 10, Get: "gold", GetAmt: 5},
		{Num: 2, Kind: game.DealWant, Give: "food", GiveAmt: 1e12, Get: "gold", GetAmt: 5}}
	b.Discovered, b.Status, b.Name, b.Deals, b.DealsBlocked = true, "embargo", "Beta", nil, "under your embargo"
	st.Diplomacy.Factions[fs[0].Key], st.Diplomacy.Factions[fs[1].Key] = a, b
	return st
}

// TestTradeOffers_BoardAndDeck: the count line matches the offers, a payable
// and an unpayable offer differ in plate, an embargoed civilization says why,
// the card runs the typed accept, and the deck fallback lists the offers.
func TestTradeOffers_BoardAndDeck(t *testing.T) {
	st := offerTown(t)
	pg := tradePage(st)
	g, _ := renderTradeBoard(&pg, st, 0, 0, 279, 71)
	text := strings.Join(g.Lines(), "\n")
	for _, want := range []string{"2 open · 1 you can pay now", "no offers: under your embargo", "Alpha", "Beta"} {
		if !strings.Contains(text, want) {
			t.Errorf("the board lacks %q", want)
		}
	}
	var plates []xk
	for _, k := range g.keyset().keys {
		if k.cmd == "diplomacy accept "+st.Ruleset().Factions()[0].Key+" 1" || k.cmd == "diplomacy accept "+st.Ruleset().Factions()[0].Key+" 2" {
			plates = append(plates, g.at(k.r.Min.X+1, k.r.Min.Y).k)
		}
	}
	if len(plates) != 2 || plates[0] == plates[1] {
		t.Errorf("a payable and an unpayable offer should differ in plate: %v", plates)
	}
	card, ok := pg.card(firstOffer(pg))
	if !ok {
		t.Fatal("no offer card")
	}
	if want := fmt.Sprintf("diplomacy accept %s 1", st.Ruleset().Factions()[0].Key); card.Command != want {
		t.Errorf("offer card command = %q, want %q", card.Command, want)
	}
	// Too small for the board, the deck lists the same offers.
	if out := deckText(pg); !strings.Contains(out, "Alpha · Sell") || !strings.Contains(out, "2 open · 1 you can pay now") {
		t.Errorf("the deck lacks the offers:\n%s", out)
	}
}

// TestTradeOffers_Scroll: a list taller than its room says what is hidden and
// scrolls; the picked offer is never cut.
func TestTradeOffers_Scroll(t *testing.T) {
	st := offerTown(t)
	f := st.Diplomacy.Factions[st.Ruleset().Factions()[0].Key]
	for n := 3; n <= 60; n++ {
		f.Deals = append(f.Deals, game.DealInfo{Num: n, Kind: game.DealWant, Give: "food", GiveAmt: 1, Get: "gold", GetAmt: 5})
	}
	st.Diplomacy.Factions[st.Ruleset().Factions()[0].Key] = f
	pg := tradePage(st)
	g, lay := renderTradeBoard(&pg, st, firstOffer(pg)+10, 0, 279, 44)
	text := strings.Join(g.Lines(), "\n")
	if lay.viewH == 0 || lay.total <= lay.viewH || !strings.Contains(text, "more offers below") {
		t.Errorf("the list should scroll (total %d, view %d):\n%s", lay.total, lay.viewH, text)
	}
	if !strings.Contains(text, "type diplomacy accept") {
		t.Error("the picked offer's lines are cut")
	}
}

// firstOffer is the deck index of the first offer card.
func firstOffer(pg deckPage) int {
	for n := 0; ; n++ {
		c, ok := pg.card(n)
		if !ok || strings.HasPrefix(c.Command, "diplomacy accept") {
			return n
		}
	}
}
