package ui

import (
	"image"
	"reflect"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/cabinet"
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
	img, liftImg                  *image.RGBA
	liftAt                        image.Rectangle
}

// openDrawer opens the drawer of an age in sight.
func (p *cabinetPanel) openDrawer(i int) bool {
	if i < 0 || i >= len(config.Ages()) || !p.drawers[i].Reached {
		return false
	}
	p.tray.slot, p.tray.well = i, 0
	p.loadTray()
	return true
}

// closeDrawer goes back to the front; it reports whether a drawer was open.
func (p *cabinetPanel) closeDrawer() bool {
	defer func() { p.tray.slot = -1 }()
	return p.tray.slot >= 0
}

func (p *cabinetPanel) loadTray() {
	i := p.tray.slot
	if i < 0 {
		return
	}
	t := cabinet.Tray{Slot: i, Money: p.keys[i], Legend: p.drawers[i].Legend}
	for j, typ := range config.CoinTypes {
		t.Types[j] = typ
		if a := p.account; a != nil && a() != nil {
			for _, c := range a().CoinStackOf(typ, t.Money).Copies {
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
		t.cells, t.cell, t.img, t.liftWell = l.Cells, l.Cell, cabinet.Open(l, t.t, p.mint), -1
		t.renders++
	}
	if t.liftWell != t.well {
		t.liftWell = t.well
		t.liftAt, t.liftImg = cabinet.Lift(l, t.t, t.well, p.mint)
	}
	p2.Picture(r, t.img)
	p2.Picture(t.liftAt.Add(r.Min), t.liftImg)
	i, d := t.slot, p.drawers[t.slot]
	pr := func(c image.Rectangle, s string, dim bool) { p.printPlate(p2, r, c, s, dim) }
	pr(l.Name, p.names[i], false)
	pr(l.Count, vibe.Sprintf("ui.cabinet_drawer.paintTray.01", "%d of %d", d.Held, d.Total), false)
	pr(l.Hint, vibe.Line("ui.cabinet_drawer.paintTray.02", "Arrows move   Esc goes back"), false)
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
