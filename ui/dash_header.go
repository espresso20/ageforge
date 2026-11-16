package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// dash_header.go is the top of the dashboard: one framed box, the header.
//
//	┌─ Name · Title · "Honour" ─────── notice ── Esc menu ┐
//	│ Age  Year  │  People 8,310/12,400   Morale 38% ...  │   now, and any alert
//	│ To the Next ████░░ 62%   Knowledge ✓  People 2/3 .. │   the way to the next age
//	│ Waiting on: Castle 3/5 · ...        Short of: Stone │   what it waits on
//	└─────────────────────────────────────────────────────┘
//
// Its top edge carries the account's name and titles at the left, a notice
// (a toast) when there is one, and the Esc button. The first row is the age,
// the year, People and Morale, and any alert (a pending catastrophe, a
// harbinger) as a chip. The second is the bar to the next age with the four
// pillars; the third what the gate waits on and what the player is short of.
// Under 100x30 the pillars are left out and the waiting line shares the
// bar's row. Every part takes a click: People and Morale open People, the
// bar and the lines under it open the Advance panel, an alert opens its
// panel, and the Esc button presses Esc.
//
// The numbers are the Advance panel's: the Ledger (dash_ledger.go) reads
// them from the same game.AdvanceView.

const (
	// headerSmallRows and headerStdRows are the box's height (its frame
	// included): two content rows under headerStdCols x headerStdTerm, three
	// from there up.
	headerSmallRows = 4
	headerStdRows   = 5
	headerStdCols   = 100
	headerStdTerm   = 30
	// headerNominalW is the width the header is written for before it has a
	// size of its own.
	headerNominalW = 160
)

// headerHeight is the header's height on a w by h terminal. It changes with
// the size of the terminal and at no other time.
func headerHeight(w, h int) int {
	if w >= headerStdCols && h >= headerStdTerm {
		return headerStdRows
	}
	return headerSmallRows
}

// The kinds of thing the header's clicks pick (xhit.id), and which.
const (
	hitHeader     = 40
	headerPeople  = 0 // People and Morale: open People
	headerCharter = 1 // the bar, the pillars and the lines under them: open the Advance panel
	headerAlert0  = 2 // the first alert; the next is headerAlert0+1
)

// The Esc button in the header's top edge: the key it presses, and what it
// does (the menu is what Esc opens when nothing is open).
const headerEsc = "Esc"

var headerMenu = vibe.E("ui.dash_header.headerMenu.01", "menu")

// headerAlert is a chip on the first row: a pending catastrophe or a
// harbinger. Its words come in three lengths (text, short, and the mark
// alone), and cmd, when there is one, is what a click on it runs.
type headerAlert struct {
	text, short, cmd string
	k                xk
}

// headFact is a fact on the first row written out, and the part of it that
// is the value: "People 8,310/12,400" and "8,310/12,400". The rest of the
// text is the fact's label.
type headFact struct{ Text, Val string }

// segs is the fact as a label and a value, in their own kinds.
func (f headFact) segs(label, val xk) []xseg {
	i := strings.Index(f.Text, f.Val)
	if f.Val == "" || i < 0 {
		return []xseg{{f.Text, label}}
	}
	parts := []xseg{{f.Text[:i], label}, {f.Val, val}}
	if rest := f.Text[i+len(f.Val):]; rest != "" {
		parts = append(parts, xseg{rest, label})
	}
	if i == 0 {
		parts = parts[1:]
	}
	return parts
}

// headerView is everything the header draws.
type headerView struct {
	// the top edge
	Name, Worn, Honour string
	Notice             string
	// the first row
	Age, Year, Prestige, Dev string
	People, PeopleShort      headFact
	Morale, Delta            headFact
	MoraleK                  xk
	Alerts                   []headerAlert
	// the rows under it; nil until the first refresh
	Led *ledgerData
}

// renderHeader draws the header w wide with rows content rows (2 or 3).
func renderHeader(v headerView, w, rows int) *xgrid {
	h := rows + 2
	g := newXGrid(w, h)
	var names []string
	for _, t := range []string{v.Name, v.Worn} {
		if t != "" {
			names = append(names, t)
		}
	}
	if v.Honour != "" {
		names = append(names, "\""+v.Honour+"\"")
	}
	title := strings.Join(names, " · ")
	// The top edge, from the right: the menu button, then the notice; the
	// name takes what is left.
	bx := w - 2 - keyButtonW(headerEsc, headerMenu.String())
	room := bx - 6
	notice := ""
	if v.Notice != "" && bx >= 14 {
		notice = " " + fitText(v.Notice, max(bx-18, 1)) + " "
		room -= runeLen(notice) + 2
	}
	g.box(0, 0, w, h, xBorder, fitText(title, max(room, 8)), xTitle, false)
	if bx >= 4 {
		g.keyButton(bx, 0, headerEsc, headerMenu.String())
		if notice != "" {
			g.put(bx-2-runeLen(notice), 0, notice, xLegend)
		}
	}

	// Row A: now. It gives up, in order: what morale does to production,
	// the alerts' long words, the year, People's long numbers, Prestige and
	// DEV, and at the last the alerts' words (their marks stay: an alert is
	// never dropped).
	for level := 0; ; level++ {
		if headerNow(g, v, w, level, false) || level == headerLevels {
			headerNow(g, v, w, level, true)
			break
		}
	}
	if v.Led == nil {
		return g
	}
	m := v.Led

	// Row B: the way to the next age.
	y := 2
	if m.v.Final {
		g.putFit(2, y, w-4, m.finalRow(), xAccent)
		return g
	}
	lead := m.leadText()
	barK := xAccent
	if m.v.AllStand {
		barK = xGood
	}
	inner := w - 4
	pct := fmt.Sprintf(" %d%%", m.Pct)
	if rows == 2 {
		// One row: the bar, then what it waits on.
		bw := clamp(w/8, 6, 16)
		lead = fitText(lead, max(inner-bw-runeLen(pct)-1, 1))
		x := g.put(2, y, lead+" ", xTitle)
		advBar(g, x, y, bw, m.v.Share, barK)
		x = g.put(x+bw, y, pct, xText)
		headerWaits(g, m, x+2, y, w-2-(x+2))
		g.hit(hitHeader, headerCharter, 2, y, inner, 1)
		return g
	}
	// The bar takes up to a third of the row; the pillars share the rest in
	// equal columns with a gutter between them, each a name, a meter when
	// there is room (one size for all four), and its count.
	n, least := len(m.Pillars), 0
	for _, p := range m.Pillars {
		least = max(least, runeLen(p.Name)+1+runeLen(headerCount(p))+2)
	}
	bw := clamp(min(inner/3, inner-n*least-3)-runeLen(lead)-1-runeLen(pct), 6, 40)
	// A bar that leaves the pillars too little room gives up cells to them,
	// down to four; when they still do not fit they are left out.
	left := func(bw int) int { return w - 2 - (2 + runeLen(lead) + 1 + bw + runeLen(pct) + 3) }
	for bw > 4 && n > 0 && left(bw)/n < least {
		bw--
	}
	if n > 0 && left(bw)/n < least {
		n = 0
		lead = fitText(lead, max(inner-bw-runeLen(pct)-1, 1))
	}
	x := g.put(2, y, lead+" ", xTitle)
	advBar(g, x, y, bw, m.v.Share, barK)
	x = g.put(x+bw, y, pct, xText) + 3
	if n > 0 {
		col := (w - 2 - x) / n
		meter := min(col-least-1, 10)
		if meter < 3 {
			meter = 0
		}
		for i, p := range m.Pillars {
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
			g.put(px, y, headerCount(p), ck)
		}
	}
	g.hit(hitHeader, headerCharter, 2, y, inner, 1)

	// Row C: what it waits on, and what you are short of. The resources
	// take up to a third of the row, as many as fit.
	y = 3
	sw := 0
	if len(m.Short) > 0 && !m.v.AllStand {
		lead := m.shortText()
		names, used := 0, runeLen(lead)
		for _, n := range m.Short {
			if used+runeLen(n)+3 > min(40, inner/3) {
				break
			}
			used += runeLen(n) + 3
			names++
		}
		if names > 0 {
			sx := w - 2
			for i := names - 1; i >= 0; i-- {
				n := " " + m.Short[i] + " "
				sx -= runeLen(n)
				g.put(sx, y, n, xChipMil)
				sx--
			}
			sx -= runeLen(lead)
			g.put(sx, y, lead, xDim)
			sw = w - 2 - sx + 3
		}
	}
	headerWaits(g, m, 2, y, inner-sw)
	g.hit(hitHeader, headerCharter, 2, y, inner, 1)
	return g
}

// headerLevels is the last level of shortening row A has.
const headerLevels = 6

// headerNow draws row A at a level of shortening when draw is set, and
// reports whether it fits.
func headerNow(g *xgrid, v headerView, w, level int, draw bool) bool {
	const y = 1
	left := []xseg{{" " + v.Age + " ", xLabel}}
	if v.Year != "" && level < 3 {
		left = append(left, xseg{"  " + v.Year, xDim})
	}
	left = append(left, xseg{"  │  ", xFaint})
	people := len(left)
	pf := v.People
	if level >= 4 {
		pf = v.PeopleShort
	}
	left = append(left, pf.segs(xDim, xBright)...)
	left = append(left, xseg{"   ", xDim})
	left = append(left, v.Morale.segs(xDim, v.MoraleK)...)
	if v.Delta.Text != "" && level < 1 {
		left = append(left, xseg{" ", xDim})
		left = append(left, v.Delta.segs(xDim, v.MoraleK)...)
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
		case level >= 6:
			t = string([]rune(a.short + " ")[:1])
		case level >= 2:
			t = a.short
		}
		hit := -1
		if a.cmd != "" {
			hit = headerAlert0 + i
		}
		right = append(right, chip{" " + t + " ", a.k, hit})
	}
	if v.Prestige != "" && level < 5 {
		right = append(right, chip{v.Prestige, xLore, -1})
	}
	if v.Dev != "" && level < 5 {
		right = append(right, chip{v.Dev, xBad, -1})
	}
	lw, rw := 0, 0
	for _, p := range left {
		lw += runeLen(p.s)
	}
	for _, c := range right {
		rw += runeLen(c.s) + 2
	}
	if !draw {
		return 2+lw+rw <= w-2
	}
	// The chips sit at the right edge, or right after the left side when that
	// reaches further. At the last level, with no room for both, the left side
	// gives way: the chips stay.
	cx := max(w-rw, 2+lw+2)
	if cx+rw-2 > w-2 {
		cx = w - rw
	}
	x := 2
	for i, p := range left {
		if i == people {
			g.hit(hitHeader, headerPeople, x, y, min(lw-(x-2), max(cx-2-x, 0)), 1)
		}
		x = g.putFit(x, y, max(cx-2-x, 0), p.s, p.k)
	}
	x = cx
	for _, c := range right {
		if x+runeLen(c.s) > w-2 {
			break
		}
		if c.hit >= 0 {
			g.hit(hitHeader, c.hit, x, y, runeLen(c.s), 1)
		}
		x = g.put(x, y, c.s, c.k) + 2
	}
	return true
}

// headerCount is a pillar's count: how many of its needs are met, or a tick.
func headerCount(p ledgerPillar) string {
	if p.Met {
		return "✓"
	}
	return p.Count
}

// headerWaits writes what the gate waits on from (x, y) in room cells.
func headerWaits(g *xgrid, m *ledgerData, x, y, room int) {
	if m.v.AllStand {
		k := xLegend
		good, text := m.readyText()
		if good {
			k = xGood
		}
		g.putFit(x, y, room, text, k)
		return
	}
	g.segsFit(x, y, room, m.waitText(room)...)
}

// ===== The box =====

// headerBox is the header's primitive: it paints the grid renderHeader
// drew, keeps it for the clicks, and lights the part under the pointer.
type headerBox struct {
	*nativeHost
	kit.Base
	// view reads the header's facts for a box w cells wide (the notice is
	// written to fit).
	view func(w int) headerView
	// plain reads the glyph tier: the plain one folds the header to ASCII.
	plain func() bool
	// run runs a command line, as if it had been typed.
	run    func(string)
	grid   *xgrid
	alerts []headerAlert // the alerts the grid drew, for their commands
	sized  bool          // the box has had a size of its own
	w, h   int
	// esc presses the Esc button; its pointer is the box's.
	esc keyPresser
	// over is the cell the pointer is over, in the box's cells, while in
	// says it is over the box; lit is the click target that is lit.
	over image.Point
	in   bool
	lit  image.Rectangle
}

func newHeaderBox(view func(w int) headerView, plain func() bool, run func(string), app *appLoop) *headerBox {
	b := &headerBox{view: view, plain: plain, run: run}
	b.nativeHost = newNativeHost()
	b.esc.wire(app)
	return b
}

// changed says what the box shows has changed, and draws it again for the
// box's size; until the box has a size of its own it is drawn at a wide one,
// so what it says can be read before the first frame.
func (b *headerBox) changed() {
	if b.sized {
		b.draw(b.w, b.h)
	} else {
		b.draw(headerNominalW, headerStdRows)
	}
}

// Layout takes the box's rectangle and draws the header for it.
func (b *headerBox) Layout(r image.Rectangle) {
	b.Place(r)
	w, h := r.Dx(), r.Dy()
	if w <= 0 || h < headerSmallRows {
		b.grid = nil
		return
	}
	b.sized = true
	if b.grid == nil || w != b.w || h != b.h {
		b.draw(w, h)
	}
}

// draw writes the header for a w by h box.
func (b *headerBox) draw(w, h int) {
	v := b.view(w)
	b.grid, b.alerts = renderHeader(v, w, h-2), v.Alerts
	if b.plain != nil && b.plain() {
		b.grid.fold()
	}
	b.w, b.h = w, h
}

// CanFocus: the header takes no keys.
func (b *headerBox) CanFocus() bool { return false }

// target is the click target under the cell pt of the box: the first thing
// of the header's own kind that covers it.
func (b *headerBox) target(pt image.Point) (xhit, bool) {
	if b.grid == nil {
		return xhit{}, false
	}
	for _, h := range b.grid.hits {
		if h.id == hitHeader && pt.In(h.r) {
			return h, true
		}
	}
	return xhit{}, false
}

// Mouse: a click on the Esc button presses Esc; a click on People or Morale
// opens People, on the bar or the lines under it the Advance panel, on an
// alert its panel. A move that changes what lights up asks for a repaint.
func (b *headerBox) Mouse(m kit.Mouse) bool {
	org := b.Bounds().Min
	if b.in = m.Kind != kit.MouseLeave; b.in {
		b.over = m.Pt().Sub(org)
	}
	was := b.lit
	b.lit = image.Rectangle{}
	if b.in {
		if h, ok := b.target(b.over); ok {
			b.lit = h.r
		}
	}
	if b.esc.mouse(b.grid, org, m) {
		return true
	}
	if m.Kind == kit.MousePress && m.Button == kit.ButtonLeft && b.in {
		if h, ok := b.target(b.over); ok {
			if cmd := b.command(h); cmd != "" {
				b.run(cmd)
				return true
			}
		}
	}
	return was != b.lit
}

// command is what a click on a target runs.
func (b *headerBox) command(h xhit) string {
	switch h.n {
	case headerPeople:
		return "people"
	case headerCharter:
		return advancePanelName
	}
	if i := h.n - headerAlert0; i >= 0 && i < len(b.alerts) {
		return b.alerts[i].cmd
	}
	return ""
}

// PointerAt: the hand over a target and over the Esc button.
func (b *headerBox) PointerAt(at image.Point) kit.Pointer {
	pt := at.Sub(b.Bounds().Min)
	if _, ok := b.target(pt); ok {
		return kit.PointerHand
	}
	return b.esc.pointer(b.grid, b.Bounds().Min, at)
}

// Paint puts the header on the painter in theme colours, with what is under
// the pointer lit.
func (b *headerBox) Paint(p *kit.Painter) {
	r := b.Bounds()
	if r.Empty() || b.grid == nil {
		return
	}
	pal := newExpPalette()
	paintExpeditionKit(p, b.grid, r.Min, pal)
	if b.in && !b.lit.Empty() {
		paintHot(p, b.grid, r.Min, pal, b.lit)
	}
	b.esc.paint(p, b.grid, r.Min, pal)
}

// paintHot repaints the cells of r, a click target, as they look under the
// pointer, over the grid already painted at org: a chip in its bright phase,
// anything else on the chip ground under its ink.
func paintHot(p *kit.Painter, g *xgrid, org image.Point, pal *expPalette, r image.Rectangle) {
	r = r.Intersect(image.Rect(0, 0, g.w, g.h))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := g.c[y*g.w+x]
			st := pal.style(c).Background(pal.styles[xKeyName].Bg)
			if c.k >= xChipShort && c.k <= xChipGoal {
				st = pal.styles[c.k+xChipShortHi-xChipShort]
			}
			p.Put(org.X+x, org.Y+y, c.r, st)
		}
	}
}

// headerPlain says the glyph tier is the plain one.
func headerPlain(s mapSettings) bool { return s.Tier == mapmodel.TierASCII }
