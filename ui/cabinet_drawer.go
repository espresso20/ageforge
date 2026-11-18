package ui

import (
	"image"
	"reflect"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/cabinet"
	"github.com/espresso20/ageforge/pkg/coins"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// cabinet_drawer.go is the cabinet's open drawer: one picture of the tray,
// kept while layout and holdings stay put, with the selected well's lift over it.

// cabinetWells are the wells in order, the legendary last.
var cabinetWells = [cabinet.Wells + 1]struct{ key, name string }{
	{config.CoinAge, vibe.Line("ui.cabinet_drawer.cabinetWells.01", "Age")},
	{config.CoinWonder, vibe.Line("ui.cabinet_drawer.cabinetWells.02", "Wonder")},
	{config.CoinTechs, vibe.Line("ui.cabinet_drawer.cabinetWells.03", "Techs")},
	{config.CoinMaximalist, vibe.Line("ui.cabinet_drawer.cabinetWells.04", "Maximalist")},
	{config.CoinHarbinger, vibe.Line("ui.cabinet_drawer.cabinetWells.05", "Harbinger")},
	{config.CoinFeat, vibe.Line("ui.cabinet_drawer.cabinetWells.06", "Legendary")},
}

// cabinetWell reads a word as a well, or -1.
func cabinetWell(word string) int {
	word = strings.ToLower(strings.TrimSpace(word))
	for i, w := range cabinetWells {
		if word == w.key || word == strings.ToLower(w.name) {
			return i
		}
	}
	return -1
}

// cabinetTray is the open drawer's state; slot is -1 while the front shows.
type cabinetTray struct {
	slot, well, liftWell, renders int
	t                             cabinet.Tray
	cells, cell                   image.Point
	img, liftImg, fanImg          *image.RGBA
	liftAt, fanAt                 image.Rectangle
	// fan is open on the selected well's stack, at copy pick; recs holds the
	// copies' records and counts every copy ever struck, per well.
	fan           bool
	fanWell, pick int
	recs          [cabinet.Wells][]game.CoinCopy
	counts        [cabinet.Wells]int
}

// openDrawer opens the drawer of an age in sight.
func (p *cabinetPanel) openDrawer(i int) bool {
	if i < 0 || i >= len(config.Ages()) || !p.drawers[i].Reached {
		return false
	}
	p.tray.slot, p.tray.well, p.tray.fan = i, 0, false
	p.loadTray()
	return true
}

// closeDrawer takes back the fan if one is open, else goes back to the front;
// it reports whether it closed anything.
func (p *cabinetPanel) closeDrawer() bool {
	if p.tray.fan {
		p.tray.fan = false
		return true
	}
	defer func() { p.tray.slot = -1 }()
	return p.tray.slot >= 0
}

// openFan spreads the selected well's copies, the finest selected.
func (p *cabinetPanel) openFan() bool {
	t := &p.tray
	if t.slot < 0 || t.fan || t.well >= cabinet.Wells || len(t.recs[t.well]) == 0 {
		return false
	}
	t.fan, t.pick = true, 0
	return true
}

// stepFan moves the selected copy along the fan; it does not wrap.
func (p *cabinetPanel) stepFan(key kit.KeyCode) {
	t, n := &p.tray, min(len(p.tray.recs[p.tray.well]), cabinet.FanMax)
	switch key {
	case kit.KeyLeft, kit.KeyUp:
		t.pick = max(t.pick-1, 0)
	case kit.KeyRight, kit.KeyDown:
		t.pick = min(t.pick+1, n-1)
	case kit.KeyHome:
		t.pick = 0
	case kit.KeyEnd:
		t.pick = n - 1
	}
}

// ticketWords are the two lines on the ticket plate: the selected copy's
// record while the fan is open, else how many copies the well's stack holds.
func (p *cabinetPanel) ticketWords() (a, b string) {
	t := &p.tray
	if t.slot < 0 || t.well >= cabinet.Wells || len(t.recs[t.well]) == 0 {
		return
	}
	n := vibe.Sprintf("ui.cabinet_drawer.ticketWords.01", "%d struck", t.counts[t.well])
	if !t.fan {
		return n, ""
	}
	c := t.recs[t.well][min(t.pick, len(t.recs[t.well])-1)]
	run := c.Run[:min(len(c.Run), 4)]
	a = vibe.Sprintf("ui.cabinet_drawer.ticketWords.02", "Level %d of %d   %s", c.Level, coins.Levels, n)
	b = vibe.Sprintf("ui.cabinet_drawer.ticketWords.03", "%s   Run %s   Figure %d", time.Unix(c.At, 0).Format("2006-01-02"), run, c.Figure)
	return a, b
}

func (p *cabinetPanel) loadTray() {
	i := p.tray.slot
	if i < 0 {
		return
	}
	t := cabinet.Tray{Slot: i, Money: p.keys[i], Legend: p.drawers[i].Legend}
	for j, typ := range config.CoinTypes {
		t.Types[j] = typ
		p.tray.recs[j], p.tray.counts[j] = nil, 0
		if a := p.account; a != nil && a() != nil {
			st := a().CoinStackOf(typ, t.Money)
			p.tray.recs[j], p.tray.counts[j] = st.Copies, st.Count
			for _, c := range st.Copies {
				t.Stacks[j] = append(t.Stacks[j], cabinet.Copy{Level: c.Level, Seed: c.Seed})
			}
		}
	}
	if !reflect.DeepEqual(t, p.tray.t) {
		p.tray.t, p.tray.img = t, nil
	}
}

// moveWell moves the selection: the five wells in a row, the legendary above.
func (p *cabinetPanel) moveWell(key kit.KeyCode) {
	w, top := &p.tray.well, p.tray.well == cabinet.Wells
	switch {
	case key == kit.KeyLeft:
		*w = max(*w-1, 0)
		if top {
			*w = 1
		}
	case key == kit.KeyRight:
		*w = min(*w+1, cabinet.Wells-1)
		if top {
			*w = cabinet.Wells - 2
		}
	case key == kit.KeyUp:
		*w = cabinet.Wells
	case key == kit.KeyDown && top:
		*w = cabinet.Wells / 2
	case key == kit.KeyHome:
		*w = 0
	case key == kit.KeyEnd:
		*w = cabinet.Wells - 1
	}
}

// wellAt is the well under a screen cell, or -1.
func (p *cabinetPanel) wellAt(at image.Point) int {
	cw, ch := cellPixels()
	r := p.Bounds()
	if l, ok := cabinet.PlanTray(r.Dx(), r.Dy(), cw, ch); ok {
		return l.At(at.X-r.Min.X, at.Y-r.Min.Y)
	}
	return -1
}

// printPlate writes words on one plate's cells.
func (p *cabinetPanel) printPlate(p2 *kit.Painter, r, c image.Rectangle, s string, dim bool) {
	if c.Empty() {
		return
	}
	c = c.Add(r.Min)
	ink, face := cabinet.PlateInk(dim), cabinet.PlateFace(dim)
	st := kit.Style{}.Foreground(theme.RGB(ink.R, ink.G, ink.B)).Background(theme.RGB(face.R, face.G, face.B))
	p2.Fill(c, st.Bg)
	p2.PrintAligned(c.Min.X, c.Min.Y, c.Dx(), kit.Escape(s), st, kit.AlignCenter)
}

// paintTray paints the open drawer, and reports false when it does not fit.
func (p *cabinetPanel) paintTray(p2 *kit.Painter, r image.Rectangle) bool {
	cw, ch := cellPixels()
	l, ok := cabinet.PlanTray(r.Dx(), r.Dy(), cw, ch)
	if !ok {
		return false
	}
	t := &p.tray
	if t.img == nil || l.Cells != t.cells || l.Cell != t.cell {
		t.cells, t.cell, t.img, t.liftWell, t.fanImg = l.Cells, l.Cell, cabinet.Open(l, t.t, p.mint), -1, nil
		t.renders++
	}
	if t.liftWell != t.well {
		t.liftWell = t.well
		t.liftAt, t.liftImg = cabinet.Lift(l, t.t, t.well, p.mint)
	}
	if t.fan && (t.fanImg == nil || t.fanWell != t.well) {
		t.fanWell = t.well
		t.fanAt, t.fanImg = cabinet.Fan(l, t.t, t.well, p.mint)
	}
	p2.Picture(r, t.img)
	p2.Picture(t.liftAt.Add(r.Min), t.liftImg)
	if t.fan && t.fanImg != nil {
		p2.Picture(t.fanAt.Add(r.Min), t.fanImg)
	}
	i, d := t.slot, p.drawers[t.slot]
	pr := func(c image.Rectangle, s string, dim bool) { p.printPlate(p2, r, c, s, dim) }
	pr(l.Name, p.names[i], false)
	pr(l.Count, vibe.Sprintf("ui.cabinet_drawer.paintTray.01", "%d of %d", d.Held, d.Total), false)
	hint := vibe.Line("ui.cabinet_drawer.paintTray.02", "Arrows move   Esc goes back")
	if t.fan {
		hint = vibe.Line("ui.cabinet_drawer.paintTray.03", "Arrows step along the copies   Esc closes the fan")
	}
	pr(l.Hint, hint, false)
	a, b := p.ticketWords()
	rows := l.Ticket
	pr(image.Rect(rows.Min.X, rows.Min.Y, rows.Max.X, rows.Min.Y+1), a, false)
	pr(image.Rect(rows.Min.X, rows.Min.Y+1, rows.Max.X, rows.Max.Y), b, false)
	for w, c := range l.Labels {
		held := w == cabinet.Wells && t.t.Legend || w < cabinet.Wells && len(t.t.Stacks[w]) > 0
		pr(c, cabinetWells[w].name, !held)
	}
	if i > 0 {
		pr(l.Above, p.names[i-1], p.dim[i-1])
	}
	pr(l.Below, p.names[i+1], p.dim[i+1])
	return true
}
