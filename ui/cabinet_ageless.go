package ui

import (
	"image"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/cabinet"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// cabinet_ageless.go is the open drawer of Ledgers, Curios and Themes: one
// well to a coin, on a plain grid that pages when the wells do not all fit.
// The wells are empty for now; the plates carry the names.

// agelessWell is one coin of an ageless drawer.
type agelessWell struct {
	name string
	held bool
}

// agelessSlot reports whether a drawer is Ledgers, Curios or Themes.
func agelessSlot(slot int) bool { return slot >= len(config.Ages()) }

// curioFamilies are the badge families whose coins are the Curios drawer's.
var curioFamilies = []string{"special", "map", "swift", "awakening", "doom"}

// agelessWells lists what an ageless drawer holds, in order.
func (p *cabinetPanel) agelessWells(slot int) []agelessWell {
	if p.set == nil || !agelessSlot(slot) {
		return nil
	}
	var acct *game.Account
	if p.account != nil {
		acct = p.account()
	}
	var out []agelessWell
	if p.keys[slot] == "ledgers" {
		for _, t := range p.set.Tallies() {
			l := 0
			if acct != nil {
				l, _, _ = acct.TallyCoin(t.Key)
			}
			out = append(out, agelessWell{t.Name, l > 0})
		}
		return out
	}
	var earned []string
	if acct != nil {
		earned = acct.EarnedBadges()
	}
	for _, b := range p.set.Badges() {
		if p.keys[slot] == "themes" && b.Family == "theme" || p.keys[slot] == "curios" && slices.Contains(curioFamilies, b.Family) {
			out = append(out, agelessWell{b.Name, slices.Contains(earned, b.Key)})
		}
	}
	return out
}

// gridDims is how many wells a page holds and how many to a row.
func (p *cabinetPanel) gridDims() (per, cols int) {
	cw, ch := cellPixels()
	r := p.Bounds()
	if l, ok := cabinet.PlanGrid(r.Dx(), r.Dy(), cw, ch); ok {
		return l.Per, l.Cols
	}
	return 1, 1
}

// moveGrid moves the selected well; the grid does not wrap.
func (p *cabinetPanel) moveGrid(key kit.KeyCode, n int) {
	per, cols := p.gridDims()
	w := &p.tray.well
	switch key {
	case kit.KeyLeft:
		*w--
	case kit.KeyRight:
		*w++
	case kit.KeyUp:
		*w -= cols
	case kit.KeyDown:
		*w += cols
	case kit.KeyPgUp:
		*w -= per
	case kit.KeyPgDn:
		*w += per
	case kit.KeyHome:
		*w = 0
	case kit.KeyEnd:
		*w = n - 1
	}
	*w = max(0, min(*w, n-1))
}

// gridKey is what the kept grid picture depends on.
type gridKey struct {
	cells, cell image.Point
	slot, page  int
	held        string
}

// paintGrid paints an open ageless drawer, and reports false when it does
// not fit.
func (p *cabinetPanel) paintGrid(p2 *kit.Painter, r image.Rectangle) bool {
	cw, ch := cellPixels()
	l, ok := cabinet.PlanGrid(r.Dx(), r.Dy(), cw, ch)
	if !ok {
		return false
	}
	t, i := &p.tray, p.tray.slot
	ws := p.agelessWells(i)
	t.well = max(0, min(t.well, len(ws)-1))
	page, pages := t.well/l.Per, (len(ws)+l.Per-1)/l.Per
	first := page * l.Per
	n := min(l.Per, len(ws)-first)
	g, held := cabinet.Grid{Slot: i, Held: make([]bool, n)}, 0
	var hs strings.Builder
	for j := range n {
		g.Held[j] = ws[first+j].held
		if g.Held[j] {
			hs.WriteByte('1')
		} else {
			hs.WriteByte('0')
		}
	}
	for _, w := range ws {
		if w.held {
			held++
		}
	}
	if k := (gridKey{l.Cells, l.Cell, i, page, hs.String()}); t.gridKey != k {
		t.gridKey, t.img, t.liftWell = k, cabinet.GridOpen(l, g, n), -1
		t.renders++
	}
	if sel := t.well - first; t.liftWell != t.well && n > 0 {
		t.liftWell = t.well
		t.liftAt, t.liftImg = cabinet.GridLift(l, g, n, sel)
	}
	p2.Picture(r, t.img)
	if n > 0 {
		p2.Picture(t.liftAt.Add(r.Min), t.liftImg)
	}
	pr := func(c image.Rectangle, s string, dim bool) { p.printPlate(p2, r, c, s, dim) }
	pr(l.Name, p.names[i], false)
	pr(l.Count, vibe.Sprintf("ui.cabinet_ageless.paintGrid.01", "%d of %d", held, len(ws)), false)
	pr(l.Hint, vibe.Line("ui.cabinet_ageless.paintGrid.02", "Arrows move   PgUp PgDn turn   Esc goes back"), false)
	a, b := "", ""
	if len(ws) > 0 {
		a = vibe.Sprintf("ui.cabinet_ageless.paintGrid.03", "Page %d of %d", page+1, pages)
		b = ws[t.well].name
	}
	pr(image.Rect(l.Ticket.Min.X, l.Ticket.Min.Y, l.Ticket.Max.X, l.Ticket.Min.Y+1), a, false)
	pr(image.Rect(l.Ticket.Min.X, l.Ticket.Min.Y+1, l.Ticket.Max.X, l.Ticket.Max.Y), b, false)
	for j := range n {
		pr(l.Plates[j], ws[first+j].name, !ws[first+j].held)
	}
	if i > 0 {
		pr(l.Above, p.names[i-1], p.dim[i-1])
	}
	if i+1 < cabinet.Slots {
		pr(l.Below, p.names[i+1], p.dim[i+1])
	}
	return true
}

// gridWellAt is the well under a screen cell on the page showing, or -1.
func (p *cabinetPanel) gridWellAt(at image.Point) int {
	cw, ch := cellPixels()
	r := p.Bounds()
	l, ok := cabinet.PlanGrid(r.Dx(), r.Dy(), cw, ch)
	if !ok {
		return -1
	}
	w := l.At(at.X-r.Min.X, at.Y-r.Min.Y)
	first := p.tray.well / l.Per * l.Per
	if w < 0 || first+w >= len(p.agelessWells(p.tray.slot)) {
		return -1
	}
	return first + w
}
