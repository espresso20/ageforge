package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// sceneFits holds a drawn scene to its frame: every row w cells, every hit
// inside it, and a place for every card the layout says can be picked.
func sceneFits(t *testing.T, where string, g *xgrid, lay deckLayout, w, h int) {
	t.Helper()
	for i, line := range strings.Split(strings.TrimSuffix(g.String(), "\n"), "\n") {
		if n := len([]rune(line)); n != w {
			t.Fatalf("%s: row %d is %d cells, want %d", where, i, n, w)
		}
	}
	for _, ht := range g.hits {
		if !ht.r.In(image.Rect(1, 1, w-1, h-2)) {
			t.Errorf("%s: a hit at %v is outside the panel", where, ht.r)
		}
	}
}

// TestTradeRoadsShowEveryRoute: the map names every route with its state
// and goods, the town by its name, the notes and the way to Factions; the
// picked route shows its button and command; nothing is lost at any size.
func TestTradeRoadsShowEveryRoute(t *testing.T) {
	keepTheme(t)
	st := tradeState()
	pg := tradePage(st)
	g, lay := renderTradeRoads(&pg, st, 2, 0, 153, 38)
	text := g.String()
	want := []string{mapmodel.TownNameOf(st.Seed), "◆◆", "Factions", " open "}
	for n := 0; n < pg.count(); n++ {
		c, _ := pg.card(n)
		want = append(want, c.Title)
		if c.Chip != "" {
			want = append(want, " "+c.Chip+" ")
		}
		if len(c.Lines) >= 2 {
			want = append(want, c.Lines[0].Value+" → "+c.Lines[1].Value)
		}
		if lay.cards[n].Empty() {
			t.Errorf("card %d (%s) has no place on the map", n, c.Title)
		}
	}
	for _, n := range pg.Notes {
		want = append(want, strings.Fields(n.Label + n.Value)[1])
	}
	picked, _ := pg.card(2)
	want = append(want, " Enter  "+picked.Action+" ", picked.Command)
	for _, s := range want {
		if !strings.Contains(text, s) {
			t.Errorf("the map does not show %q:\n%s", s, text)
		}
	}
	for w := 60; w <= 200; w += 11 {
		for h := 14; h <= 50; h += 4 {
			g, lay := renderTradeRoads(&pg, st, 1, 0, w, h)
			sceneFits(t, "trade", g, lay, w, h)
			for n := 0; n < pg.count(); n++ {
				if c, _ := pg.card(n); !strings.Contains(g.String(), fitText(c.Title, 12)[:4]) && lay.total <= lay.viewH {
					t.Errorf("trade at %dx%d: %q is not on the panel", w, h, c.Title)
				}
			}
		}
	}
	// A window too small for the map gets the cards.
	small, _ := renderTradeRoads(&pg, st, 0, 0, 100, 30)
	if !strings.Contains(small.String(), pg.Sections[0].Title+" ─") {
		t.Errorf("at 100x30 the Trade panel is not the deck:\n%s", small.String())
	}
}

// TestArmyFieldShowsTheBalanceAndTheOrders: the field sets the defense
// against the threat, musters every post, lists every campaign as an order,
// and only the orders can be picked.
func TestArmyFieldShowsTheBalanceAndTheOrders(t *testing.T) {
	keepTheme(t)
	st := armyState()
	pg := armyPage(st)
	first := armyOverview
	g, lay := renderArmyField(&pg, st, first, 0, 153, 38)
	text := g.String()
	over := pg.Sections[0].Cards
	want := []string{over[2].Title + " " + over[2].Lines[0].Value, over[3].Title + " " + over[3].Lines[0].Value, over[0].Title, "▲", musterLegend(musterUnit(100))}
	for _, n := range pg.Notes {
		want = append(want, strings.Fields(n.Label + n.Value)[0])
	}
	for _, l := range over[1].Lines {
		want = append(want, l.Label+l.Value)
	}
	for n := first; n < pg.count(); n++ {
		c, _ := pg.card(n)
		want = append(want, c.Title)
		if lay.cards[n].Empty() {
			t.Errorf("order %d (%s) has no place on the field", n, c.Title)
		}
	}
	for _, s := range want {
		if !strings.Contains(text, s) {
			t.Errorf("the field does not show %q:\n%s", s, text)
		}
	}
	for n := 0; n < first; n++ {
		if !lay.cards[n].Empty() {
			t.Errorf("overview card %d can be picked on the field", n)
		}
	}
	if got := lay.step(first, 0, -1); got != first {
		t.Errorf("Up from the first order went to card %d", got)
	}
	if strings.Count(text, "█") < 20 {
		t.Errorf("the field has no balance bar:\n%s", text)
	}
	for w := 60; w <= 200; w += 11 {
		for h := 14; h <= 50; h += 4 {
			g, lay := renderArmyField(&pg, st, first, 0, w, h)
			sceneFits(t, "army", g, lay, w, h)
		}
	}
}

// TestMusterUnit: a mark stands for a round number of soldiers, and the
// largest post fits its row.
func TestMusterUnit(t *testing.T) {
	for most, want := range map[int]int{0: 1, 60: 1, 61: 2, 120: 2, 121: 5, 300: 5, 301: 10, 5000: 100} {
		if got := musterUnit(most); got != want {
			t.Errorf("musterUnit(%d) = %d, want %d", most, got, want)
		}
	}
}

// TestMileRoadsOrderAndPick: a road has its reached stones first, then the
// one being walked to, then the ones ahead; the picked stone is shown in
// full; Left and Right go along a road and Up and Down from road to road.
func TestMileRoadsOrderAndPick(t *testing.T) {
	keepTheme(t)
	st := milestoneState()
	pg := milestonesPage(st)
	g, lay := renderMileRoads(&pg, st, 0, 0, 153, 38)
	sceneFits(t, "milestones", g, lay, 153, 38)
	for si, s := range pg.Sections {
		n0 := 0
		for _, before := range pg.Sections[:si] {
			n0 += len(before.Cards)
		}
		lastDone, firstAhead := -1, 1<<30
		for ci := range s.Cards {
			x := lay.cards[n0+ci].Min.X
			if mileDone(&s.Cards[ci]) {
				lastDone = max(lastDone, x)
			} else if lay.cards[n0+ci].Min.Y == lay.cards[n0].Min.Y {
				firstAhead = min(firstAhead, x)
			}
		}
		if lastDone >= firstAhead {
			t.Errorf("%s: a reached stone at column %d stands after one not reached at %d", s.Title, lastDone, firstAhead)
		}
	}
	for n := 0; n < pg.count(); n++ {
		c, _ := pg.card(n)
		g, lay := renderMileRoads(&pg, st, n, 0, 153, 38)
		text := g.String()
		for _, want := range []string{c.Title, c.Desc, c.MeterText, c.Why} {
			if want != "" && !strings.Contains(text, fitText(want, 100)) {
				t.Errorf("with %q picked the panel does not show %q:\n%s", c.Title, want, text)
			}
		}
		for _, l := range c.Lines {
			if !strings.Contains(text, l.Value) {
				t.Errorf("with %q picked the panel does not show its line %q", c.Title, l.Value)
			}
		}
		right := lay.step(n, 1, 0)
		if a, b := lay.cards[n], lay.cards[right]; right != n && b.Min.Y == a.Min.Y && b.Min.X <= a.Min.X {
			t.Errorf("Right from %q went left along its road", c.Title)
		}
	}
	for w := 60; w <= 200; w += 11 {
		for h := 14; h <= 50; h += 4 {
			g, lay := renderMileRoads(&pg, st, 1, 0, w, h)
			sceneFits(t, "milestones", g, lay, w, h)
		}
	}
}

// TestMileRoadsScroll: more roads than fit scroll, and the pick stays in
// view.
func TestMileRoadsScroll(t *testing.T) {
	keepTheme(t)
	var pg deckPage
	for i := 0; i < 20; i++ {
		pg.Sections = append(pg.Sections, deckSection{Title: "Road", Meta: "0 of 2", Cards: []deckCard{{Title: "First"}, {Title: "Second"}}})
	}
	p := newDeckPanel(func(game.GameState) deckPage { return pg })
	p.draw = renderMileRoads
	p.open(game.GameState{})
	p.Layout(image.Rect(0, 0, 153, 38))
	if p.lay.total <= p.lay.viewH || !strings.Contains(p.grid.String(), "┃") {
		t.Fatalf("twenty roads fit (%d rows in %d) or show no scroll mark", p.lay.total, p.lay.viewH)
	}
	p.pick(len(p.lay.cards) - 1)
	p.Layout(p.Bounds())
	last := p.lay.cards[len(p.lay.cards)-1]
	if last.Min.Y < p.scroll || last.Max.Y > p.scroll+p.lay.viewH {
		t.Errorf("the last stone (rows %d to %d) is out of view at scroll %d with %d rows", last.Min.Y, last.Max.Y, p.scroll, p.lay.viewH)
	}
	hit := false
	for _, ht := range p.grid.hits {
		hit = hit || ht.n == len(p.lay.cards)-1
	}
	if !hit {
		t.Error("the picked stone cannot be clicked after the roads scrolled")
	}
}
