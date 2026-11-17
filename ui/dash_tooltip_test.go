package ui

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// tipClock pins the tooltip's clock and moves it on.
type tipClock struct{ now time.Time }

func (c *tipClock) at() time.Time        { return c.now }
func (c *tipClock) wait(d time.Duration) { c.now = c.now.Add(d) }
func pinTip(d *Dashboard) *tipClock {
	c := &tipClock{time.Unix(5000, 0)}
	d.app.feel.tip.Now = c.at
	return c
}

// pointerTo moves the pointer to pt the way the app does: the tooltip first
// sees the event, then the screen.
func pointerTo(d *Dashboard, st *stage, pt image.Point) {
	m := kit.Mouse{Kind: kit.MouseMove, X: pt.X, Y: pt.Y}
	d.app.feel.tip.Move(m)
	st.mouse(m)
}

// tipRows is the text on the tooltip's plate now, empty when none shows.
func tipRows(d *Dashboard, st *stage) []string {
	st.paint()
	b := d.rootFx.Bounds()
	r := d.app.feel.tip.Rect(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y-promptRows))
	var rows []string
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := make([]rune, 0, r.Dx())
		for x := r.Min.X; x < r.Max.X; x++ {
			row = append(row, st.buf.At(x, y).R)
		}
		if line := strings.TrimSpace(strings.Trim(string(row), "+-|")); line != "" {
			rows = append(rows, line)
		}
	}
	return rows
}

// showCard turns the Buildings box until the card of key is on the screen.
func showCard(t *testing.T, d *Dashboard, st *stage, key string) placedCard {
	t.Helper()
	box := d.economyTab.buildBox
	box.expanded = true
	for box.first = 0; box.first < 60; box.first++ {
		d.refresh()
		st.paint()
		for _, pc := range box.placed {
			if pc.card.key == key {
				return pc
			}
		}
	}
	t.Fatalf("the Buildings box never shows %s", key)
	return placedCard{}
}

// buildTip rests the pointer on the Build button of key and returns what the
// tooltip says (empty when none shows).
func buildTip(t *testing.T, d *Dashboard, st *stage, clock *tipClock, key string) []string {
	t.Helper()
	pc := showCard(t, d, st, key)
	pointerTo(d, st, centre(d, d.economyTab.buildBox.cardKeys.keys[pc.build].r))
	if got := tipRows(d, st); len(got) > 0 {
		t.Fatalf("the tooltip shows before the pointer rested: %q", got)
	}
	clock.wait(kit.TipDelay)
	return tipRows(d, st)
}

// An off Build button says why: the resource short and by how much, with
// two short the one furthest away, and the cap that blocks it.
func TestOffBuildButtonSaysWhy(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 45)
	clock := pinTip(d)
	var key string
	var cost map[string]float64
	for k, b := range eng.GetState().Buildings {
		if b.Unlocked && !b.IsLegacy && b.Category != "wonder" && b.AgeKey == "medieval_age" && len(b.NextCost) >= 2 && (key == "" || k < key) {
			key, cost = k, b.NextCost
		}
	}
	if key == "" {
		t.Fatal("staging: no building of the age takes two resources")
	}
	res := sortedKeysOf(cost)
	a, b := res[0], res[1]

	// One short: the other is held in full.
	setStock(eng, 1e9)
	eng.SetStockForTest(a, cost[a]*0.4)
	d.refresh()
	rows := buildTip(t, d, st, clock, key)
	want := "! short of " + game.Amount(cost[a]-cost[a]*0.4, a)
	if len(rows) != 1 || rows[0] != want {
		t.Errorf("one short: tooltip %q, want %q", rows, want)
	}

	// Two short: the card's reason names one of them, and the tooltip says
	// exactly what the card says.
	clock.wait(time.Second)
	d.app.feel.tip.Clear()
	eng.SetStockForTest(a, cost[a]*0.4)
	eng.SetStockForTest(b, cost[b]*0.1)
	d.refresh()
	rows = buildTip(t, d, st, clock, key)
	pc := showCard(t, d, st, key)
	short := map[string]bool{
		"! short of " + game.Amount(cost[a]-cost[a]*0.4, a): true,
		"! short of " + game.Amount(cost[b]-cost[b]*0.1, b): true,
	}
	if len(rows) != 1 || rows[0] != "! "+pc.card.reason || !short[rows[0]] {
		t.Errorf("two short: tooltip %q, card says %q, want one of %v", rows, pc.card.reason, short)
	}

	// Enough to build: the button is on and has nothing to say.
	d.app.feel.tip.Clear()
	setStock(eng, 1e9)
	d.refresh()
	if rows := buildTip(t, d, st, clock, key); len(rows) != 0 {
		t.Errorf("a button that is on says %q", rows)
	}
}

func TestOffBuildButtonAtItsCapSaysSo(t *testing.T) {
	restoreForge(t)
	d, pages, eng := wallTown(t, map[string]int{"grand_amphitheatre_monument": 1})
	st := newStage(pages, 160, 45)
	clock := pinTip(d)
	setStock(eng, 1e12)
	rows := buildTip(t, d, st, clock, "grand_amphitheatre_monument")
	if len(rows) != 1 || rows[0] != "! "+txtLimit.String() {
		t.Errorf("at the cap the tooltip says %q, want %q", rows, "! "+txtLimit.String())
	}
}

// The tooltip of a price lists each resource as held and needed and marks
// exactly the short ones.
func TestPriceTooltipMarksTheShortResources(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 45)
	clock := pinTip(d)
	var key string
	var cost map[string]float64
	for k, b := range eng.GetState().Buildings {
		if b.Unlocked && !b.IsLegacy && b.Category != "wonder" && b.AgeKey == "medieval_age" && len(b.NextCost) >= 2 && (key == "" || k < key) {
			key, cost = k, b.NextCost
		}
	}
	res := sortedKeysOf(cost)
	setStock(eng, 1e9)
	eng.SetStockForTest(res[0], cost[res[0]]*0.5)
	d.refresh()
	pc := showCard(t, d, st, key)
	pointerTo(d, st, image.Pt(0, 0)) // away from everything
	box := d.economyTab.buildBox
	x, y, _, _ := box.GetRect()
	pointerTo(d, st, image.Pt(x+pc.tx, y+pc.y+2))
	clock.wait(kit.TipDelay)
	rows := tipRows(d, st)
	if len(rows) != len(pc.card.price) {
		t.Fatalf("the price has %d resources, the tooltip %d lines: %q", len(pc.card.price), len(rows), rows)
	}
	for i, p := range pc.card.price {
		short := strings.HasPrefix(rows[i], "! ")
		if short == p.ok {
			t.Errorf("%s: short in the card %v, marked %v (%q)", p.res, !p.ok, short, rows[i])
		}
		if !strings.Contains(rows[i], game.ResourceName(p.res)) || !strings.Contains(rows[i], "have") || !strings.Contains(rows[i], "need") {
			t.Errorf("line %q lacks the resource, held or needed", rows[i])
		}
	}
	if !strings.HasPrefix(rows[0], "! ") && !strings.HasPrefix(rows[1], "! ") {
		t.Errorf("staging: no resource of %s is short: %q", key, rows)
	}
}

// It goes on a move off the target, a key, a click and a panel opening.
func TestTooltipGoesAndStaysInside(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 45)
	clock := pinTip(d)
	setStock(eng, 0)
	d.refresh()
	st.paint()
	d.economyTab.buildBox.expanded = true
	d.refresh()
	st.paint()
	key := ""
	for _, pc := range d.economyTab.buildBox.placed {
		if !pc.card.canBuild && pc.build >= 0 {
			key = pc.card.key
			break
		}
	}
	if key == "" {
		t.Fatal("staging: no card that cannot be built")
	}
	shown := func(how string, act func()) {
		t.Helper()
		pc := showCard(t, d, st, key)
		at := centre(d, d.economyTab.buildBox.cardKeys.keys[pc.build].r)
		pointerTo(d, st, image.Pt(0, 0))
		pointerTo(d, st, at)
		clock.wait(kit.TipDelay)
		if len(tipRows(d, st)) == 0 {
			t.Fatalf("%s: staging: the tooltip does not show", how)
		}
		act()
		if got := tipRows(d, st); len(got) != 0 {
			t.Errorf("%s: the tooltip still shows %q", how, got)
		}
	}
	shown("a move off the target", func() { pointerTo(d, st, image.Pt(0, 0)) })
	shown("a key", func() { d.app.key(kit.Key{Code: kit.KeyRune, Rune: 'x'}) })
	shown("a click", func() {
		m := kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 0, Y: 0}
		d.app.feel.tip.Move(m)
	})
	shown("a panel opening", func() { d.runForTest("map") })
	d.overlayMgr.Hide()

	// Inside the window, off the target and clear of the command bar, at the
	// right edge of the screen as well.
	pc := showCard(t, d, st, key)
	target := d.economyTab.buildBox.cardKeys.keys[pc.build].r
	x, y, _, _ := d.economyTab.buildBox.GetRect()
	tr := target.Add(image.Pt(x, y))
	pointerTo(d, st, centre(d, target))
	clock.wait(kit.TipDelay)
	st.paint()
	b := d.rootFx.Bounds()
	area := image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y-promptRows)
	r := d.app.feel.tip.Rect(area)
	if r.Empty() || !r.In(area) || r.Overlaps(tr) {
		t.Errorf("the plate %v is outside %v or covers its target %v", r, area, tr)
	}
}

// The pick moving onto a card by key shows its tooltip after the same delay
// as the pointer, with no pointer in the box.
func TestTooltipOnTheKeyPick(t *testing.T) {
	d, st, eng := medievalTown(t, 160, 45)
	clock := pinTip(d)
	setStock(eng, 0)
	d.economyTab.buildBox.expanded = true
	d.refresh()
	st.paint()
	key := d.economyTab.buildBox.placed[0].card.key
	d.economyTab.buildBox.pick(key)
	st.paint()
	if rows := tipRows(d, st); len(rows) != 0 {
		t.Fatalf("shown at once: %q", rows)
	}
	clock.wait(kit.TipDelay - time.Millisecond)
	if rows := tipRows(d, st); len(rows) != 0 {
		t.Fatalf("shown early: %q", rows)
	}
	clock.wait(time.Millisecond)
	rows := tipRows(d, st)
	pc := showCard(t, d, st, key)
	if len(rows) != 1 || rows[0] != "! "+pc.card.reason {
		t.Errorf("the key pick shows %q, want the reason %q", rows, pc.card.reason)
	}
	d.app.key(kit.Key{Code: kit.KeyRune, Rune: 'x'})
	if rows := tipRows(d, st); len(rows) != 0 {
		t.Errorf("a key left %q", rows)
	}
}
