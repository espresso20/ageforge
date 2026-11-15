package ui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// SCRATCH: a design draft of the dashboard's header. Not for commit as is.

type headAlert struct {
	text, short, cmd string
	k                xk
}

type headView struct {
	Name, Worn, Honour string
	Age, Year          string
	Prestige           int
	Dev                bool
	People             string
	Morale             int
	MoraleK            xk
	Delta              string
	Alerts             []headAlert
	Notice             string
	// the ledger
	Final, AllStand, Ready bool
	Next                   string
	Share                  float64
	Pct                    int
	Pillars                []ledgerPillar
	Waits                  []string
	Then                   string
	Short                  []string
}

const (
	hitHead     = 40
	headPeople  = 0
	headCharter = 1
	headAlert0  = 2
)

// renderHead draws the header w wide with rows content rows (2 or 3).
func renderHead(v headView, w, rows int) *xgrid {
	h := rows + 2
	g := newXGrid(w, h)
	title := v.Name
	if v.Worn != "" {
		title += " · " + v.Worn
	}
	if v.Honour != "" {
		title += " · \"" + v.Honour + "\""
	}
	// The top edge, from the right: the menu button, then the notice; the
	// name takes what is left.
	bx := w - 2 - keyButtonW("Esc", "menu")
	room := bx - 6
	notice := ""
	if v.Notice != "" {
		notice = " " + v.Notice + " "
		room -= len([]rune(notice)) + 2
	}
	g.box(0, 0, w, h, xBorder, fitText(title, max(room, 8)), xTitle, false)
	g.keyButton(bx, 0, "Esc", "menu")
	if notice != "" {
		g.put(bx-2-len([]rune(notice)), 0, notice, xLegend)
	}

	// Row A: now. It gives up, in order: what morale does to production,
	// the alerts' long words, the year, Prestige and DEV, and at the last
	// the alerts' words (their marks stay: an alert is never dropped).
	y := 1
	for level := 0; ; level++ {
		if headNow(g, v, w, level, false) || level == headLevels {
			headNow(g, v, w, level, true)
			break
		}
	}

	// Row B: the way to the next age.
	x := 0
	y = 2
	if v.Final {
		g.put(2, y, "You have reached the final age.", xAccent)
		return g
	}
	lead := "To the " + v.Next
	barK := xAccent
	if v.AllStand {
		barK = xGood
	}
	inner := w - 4
	pct := fmt.Sprintf(" %d%%", v.Pct)
	if rows == 2 {
		// One row: the bar, then what it waits on.
		bw := clamp(w/8, 6, 16)
		x = g.put(2, y, lead+" ", xTitle)
		advBar(g, x, y, bw, v.Share, barK)
		x = g.put(x+bw, y, pct, xText)
		headWaits(g, v, x+2, y, w-2-(x+2))
		g.hit(hitHead, headCharter, 2, y, inner, 1)
		return g
	}
	// The bar takes up to a third of the row; the pillars share the rest in
	// equal columns with a gutter between them, each a name, a meter when
	// there is room (one size for all four), and its count.
	n, least := len(v.Pillars), 0
	for _, p := range v.Pillars {
		least = max(least, len([]rune(p.Name))+1+len([]rune(headCount(p)))+2)
	}
	bw := clamp(min(inner/3, inner-n*least-3)-len([]rune(lead))-1-len(pct), 6, 40)
	x = g.put(2, y, lead+" ", xTitle)
	advBar(g, x, y, bw, v.Share, barK)
	x = g.put(x+bw, y, pct, xText) + 3
	if n > 0 {
		col := (w - 2 - x) / n
		meter := min(col-least-1, 10)
		if meter < 3 {
			meter = 0
		}
		for i, p := range v.Pillars {
			px := x + i*col
			nk, ck, mk := xText, xDim, xAccent
			if p.Met {
				nk, ck, mk = xGood, xGood, xGood
			}
			px = g.put(px, y, p.Name+" ", nk)
			if meter > 0 {
				advBar(g, px, y, meter, p.Share, mk)
				px += meter + 1
			}
			g.put(px, y, headCount(p), ck)
		}
	}
	g.hit(hitHead, headCharter, 2, y, inner, 1)

	// Row C: what it waits on, and what you are short of.
	y = 3
	sw := 0
	if len(v.Short) > 0 && !v.AllStand {
		sx := w - 2
		for i := len(v.Short) - 1; i >= 0; i-- {
			n := " " + v.Short[i] + " "
			sx -= len([]rune(n))
			g.put(sx, y, n, xChipMil)
			sx--
		}
		sx -= len("Short of:")
		g.put(sx, y, "Short of:", xDim)
		sw = w - 2 - sx + 3
	}
	headWaits(g, v, 2, y, inner-sw)
	g.hit(hitHead, headCharter, 2, y, inner, 1)
	return g
}

// headLevels is the last level of shortening row A has.
const headLevels = 5

// headNow draws row A at a level of shortening when draw is set, and reports
// whether it fits.
func headNow(g *xgrid, v headView, w, level int, draw bool) bool {
	const y = 1
	left := []xseg{{" " + v.Age + " ", xLabel}}
	if v.Year != "" && level < 3 {
		left = append(left, xseg{"  " + v.Year, xDim})
	}
	left = append(left, xseg{"  │  ", xFaint})
	people := len(left)
	left = append(left, xseg{"People ", xDim}, xseg{v.People, xBright}, xseg{"   Morale ", xDim}, xseg{fmt.Sprintf("%d%%", v.Morale), v.MoraleK})
	if v.Delta != "" && level < 1 {
		left = append(left, xseg{" (production ", xDim}, xseg{v.Delta, v.MoraleK}, xseg{")", xDim})
	}
	type chip struct {
		s   string
		k   xk
		hit int
	}
	var right []chip
	for i, a := range v.Alerts {
		t := a.text
		switch {
		case level >= 5:
			t = string([]rune(a.short)[:1])
		case level >= 2:
			t = a.short
		}
		right = append(right, chip{" " + t + " ", a.k, headAlert0 + i})
	}
	if v.Prestige > 0 && level < 4 {
		right = append(right, chip{fmt.Sprintf("Prestige %d", v.Prestige), xLore, -1})
	}
	if v.Dev && level < 4 {
		right = append(right, chip{"DEV", xBad, -1})
	}
	lw, rw := 0, 0
	for _, p := range left {
		lw += len([]rune(p.s))
	}
	for _, c := range right {
		rw += len([]rune(c.s)) + 2
	}
	if !draw {
		return 2+lw+rw <= w-2
	}
	x := 2
	for i, p := range left {
		if i == people {
			g.hit(hitHead, headPeople, x, y, lw-(x-2), 1)
		}
		x = g.putFit(x, y, w-2-x, p.s, p.k)
	}
	x = max(w-rw, 2+lw+2)
	for _, c := range right {
		if x+len([]rune(c.s)) > w-2 {
			break
		}
		if c.hit >= 0 {
			g.hit(hitHead, c.hit, x, y, len([]rune(c.s)), 1)
		}
		x = g.put(x, y, c.s, c.k) + 2
	}
	return true
}

// headCount is a pillar's count: how many of its needs are met, or a tick.
func headCount(p ledgerPillar) string {
	if p.Met {
		return "✓"
	}
	return p.Note
}

// headWaits writes what the gate waits on from (x, y) in room cells.
func headWaits(g *xgrid, v headView, x, y, room int) {
	if v.AllStand {
		if v.Ready {
			g.putFit(x, y, room, "Every pillar stands. Type advance to go on.", xGood)
		} else {
			g.putFit(x, y, room, "Every pillar stands. Answer the catastrophe first (type catastrophe).", xLegend)
		}
		return
	}
	end := x + room
	x = g.put(x, y, "Waiting on: ", xDim)
	for i, it := range v.Waits {
		more := fmt.Sprintf("+%d more", len(v.Waits)-i)
		need := len([]rune(it))
		if i > 0 {
			need += 3
		}
		if i < len(v.Waits)-1 {
			need += 3 + len(more)
		}
		if x+need > end {
			if i > 0 {
				x = g.put(x, y, " · ", xFaint)
			}
			g.put(x, y, more, xDim)
			return
		}
		if i > 0 {
			x = g.put(x, y, " · ", xFaint)
		}
		x = g.put(x, y, it, xText)
	}
	if v.Then != "" && x+8+len([]rune(v.Then)) <= end {
		x = g.put(x, y, " · ", xFaint)
		g.put(x, y, "then "+v.Then, xDim)
	}
}

func TestZZHeaderMock(t *testing.T) {
	out := os.Getenv("HEADER_MOCK_DIR")
	if out == "" {
		t.Skip("set HEADER_MOCK_DIR")
	}
	keepTheme(t)
	early := headView{
		Name: "The Sorrowing March…", Honour: "Headman", Age: "Primitive Age", Year: "Late Cretaceous", Dev: true,
		People: "0/0", Morale: 50, MoraleK: xGood, Delta: "+<1%",
		Next: "Stone Age", Share: 0.0, Pct: 0,
		Pillars: []ledgerPillar{{Name: "Knowledge", Note: "0/1"}, {Name: "People", Note: "0/3"}, {Name: "Industry", Note: "0/2"}, {Name: "Monument", Note: "0/2"}},
		Waits:   []string{"Story Circle 0/5", "Gathering Camp 0/8", "Hut 0/10", "Shrine 0/8", "Stash 0/5", "Wood Camp 0/8", "Sacred Grove bank 0%"},
		Then:    "Sacred Grove", Short: []string{"Food", "Wood"},
	}
	mid := headView{
		Name: "Ada", Worn: "Magistrate", Honour: "Consul", Age: "Medieval Age", Year: "1066 AD", Prestige: 2,
		People: "8,310/12,400", Morale: 38, MoraleK: xBad, Delta: "-6%",
		Alerts: []headAlert{{"☄ Catastrophe: type catastrophe", "☄ Catastrophe", "catastrophe", xChipMil}, {"⚑ Harbinger: type harbinger", "⚑ Harbinger", "harbinger", xChipLegend}},
		Notice: "★ Badge earned: Stonecutter",
		Next:   "Renaissance", Share: 0.62, Pct: 62,
		Pillars: []ledgerPillar{{Name: "Knowledge", Share: 1, Met: true}, {Name: "People", Share: 0.66, Note: "2/3"}, {Name: "Industry", Share: 0.5, Note: "1/2"}, {Name: "Monument", Share: 0.3, Note: "0/1"}},
		Waits:   []string{"Castle 3/5", "University 0/2", "Guild Hall 1/4"},
		Then:    "Great Cathedral", Short: []string{"Stone", "Iron", "Gold"},
	}
	ready := mid
	ready.Pillars = append([]ledgerPillar(nil), mid.Pillars...)
	ready.Alerts, ready.Notice, ready.AllStand, ready.Ready, ready.Share, ready.Pct = nil, "", true, true, 1, 100
	for i := range ready.Pillars {
		ready.Pillars[i].Met, ready.Pillars[i].Share = true, 1
	}
	shots := []struct {
		name    string
		v       headView
		w, rows int
	}{
		{"early-153", early, 153, 3}, {"mid-153", mid, 153, 3}, {"ready-153", ready, 153, 3},
		{"early-120", early, 120, 3}, {"mid-120", mid, 120, 3},
		{"early-100", early, 100, 3}, {"mid-100", mid, 100, 3},
		{"early-80", early, 80, 2}, {"mid-80", mid, 80, 2},
	}
	for _, s := range shots {
		g := renderHead(s.v, s.w, s.rows)
		b := kit.NewBuffer(s.w, g.h)
		p := kit.NewPainter(b)
		paintExpeditionKit(p, g, image.Point{}, newExpPalette())
		f, err := os.Create(filepath.Join(out, "HD-"+s.name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, ScreenImage(b, shotCellW, shotCellH)); err != nil {
			t.Fatal(err)
		}
		f.Close()
		t.Log("\n" + g.String())
	}
}
