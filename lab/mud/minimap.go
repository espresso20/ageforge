package main

// minimap.go draws the room graph with box characters: a glyph per place (its
// meaning survives monochrome), the place's short name on the same grid, the
// roads between them, and goods moving along the roads toward the stores.
// It is also the compact, glanceable form of the whole view.

import (
	"fmt"

	"github.com/espresso20/ageforge/theme"
)

var placeGlyph = map[PlaceKey]rune{
	Square: '○', Homes: '⌂', Fields: '≡', Woods: '♣', Quarry: '▲', Forge: '¤', Works: '#',
	Temple: '†', Academy: '§', Market: '$', Harbour: '≈', Stores: '▪', Barracks: '‡', Wonders: '★', Gate: '∩',
}

func glyphFor(k PlaceKey) string { return string(placeGlyph[k]) }

// MapOpts selects the map's form.
type MapOpts struct {
	LabelW  int  // characters of label per place
	Gap     int  // road length between columns
	Counts  bool // second line per place: building count and marks
	Here    PlaceKey
	Changed map[PlaceKey]bool
	Frame   int
}

// mapSize is the grid a map with opts needs.
func (c *City) mapSize(o MapOpts) (w, h int) {
	w = c.Cols*(1+o.LabelW) + (c.Cols-1)*o.Gap
	rowH := 2
	if o.Counts {
		rowH = 3
	}
	h = c.Rows*rowH - 1
	return
}

// flows are the roads goods travel: from each producing place toward the
// stores, and between market and harbour while a trade route runs. The value
// is the direction of travel along the edge (true: A→B as listed in edges).
func (c *City) flows() map[[2]PlaceKey]bool {
	out := map[[2]PlaceKey]bool{}
	mark := func(a, b PlaceKey) {
		for _, e := range edges {
			if e.A == a && e.B == b {
				out[[2]PlaceKey{e.A, e.B}] = true
				return
			}
			if e.A == b && e.B == a {
				out[[2]PlaceKey{e.A, e.B}] = false
				return
			}
		}
	}
	dest := Stores
	if c.Places[Stores] == nil {
		dest = Square
	}
	for _, k := range []PlaceKey{Fields, Woods, Quarry, Forge, Works} {
		p := c.Places[k]
		if p == nil || p.Path || p.Total() == 0 {
			continue
		}
		if st := p.Staffing(); st >= 0 && st < 0.3 {
			continue // a stalled place sends nothing: the bottleneck shows as a still road
		}
		prev := k
		for _, step := range c.Route(k, dest) {
			mark(prev, step)
			prev = step
		}
	}
	if len(c.St.Trade.ActiveRoutes) > 0 && c.Places[Harbour] != nil && c.Places[Market] != nil {
		mark(Harbour, Market)
	}
	return out
}

// drawMap renders the room graph into a grid.
func (c *City) drawMap(o MapOpts) *Grid {
	w, h := c.mapSize(o)
	g := NewGrid(w, h)
	rowH := 2
	if o.Counts {
		rowH = 3
	}
	px := func(col int) int { return (col - c.MinCol) * (1 + o.LabelW + o.Gap) }
	py := func(row int) int { return (row - c.MinRow) * rowH }
	flows := c.flows()

	// Roads first, labels on top.
	for _, e := range edges {
		a, b := c.Places[e.A], c.Places[e.B]
		if a == nil || b == nil {
			continue
		}
		dir, moving := flows[[2]PlaceKey{e.A, e.B}]
		_, has := flows[[2]PlaceKey{e.A, e.B}]
		moving = has
		if e.D == East {
			y := py(a.Def.Row)
			x0, x1 := px(a.Def.Col)+1, px(b.Def.Col)-1
			for x := x0; x <= x1; x++ {
				g.Set(x, y, '─', theme.RoleBorder)
			}
			if moving {
				n := x1 - x0 + 1
				step := o.Frame % n
				if !dir {
					step = n - 1 - step
				}
				g.Set(x0+step, y, '•', theme.RoleHighlight)
			}
		} else { // South
			x := px(a.Def.Col)
			for y := py(a.Def.Row) + 1; y < py(b.Def.Row); y++ {
				g.Set(x, y, '│', theme.RoleBorder)
			}
			if moving && o.Frame%2 == 0 {
				y := py(a.Def.Row) + rowH - 1
				g.Set(x, y, '•', theme.RoleHighlight)
			}
		}
	}
	// Places.
	for _, k := range c.Order {
		p := c.Places[k]
		x, y := px(p.Def.Col), py(p.Def.Row)
		glyph, grole := placeGlyph[k], theme.RoleAccent
		lrole := theme.RoleText
		switch {
		case k == o.Here:
			glyph, grole, lrole = '@', theme.RoleBright, theme.RoleSelectionText
		case p.Path:
			glyph, grole, lrole = '·', theme.RoleDim, theme.RoleDim
		case p.Ruins() > 0 || (k == Gate && len(c.atWar()) > 0):
			grole = theme.RoleNegative
		case p.Staffing() >= 0 && p.Staffing() < 0.6:
			grole = theme.RoleWarning
		}
		g.Set(x, y, glyph, grole)
		label := []rune(p.Short)
		if p.Path {
			label = []rune("·····")
			if len(label) > o.LabelW {
				label = label[:o.LabelW]
			}
		}
		if len(label) > o.LabelW {
			label = label[:o.LabelW]
		}
		// Clear the road under the label.
		for i := 0; i < o.LabelW && x+1+i < w; i++ {
			if i < len(label) {
				g.Set(x+1+i, y, label[i], lrole)
				if k == o.Here {
					g.Cells[y*g.W+x+1+i].Bold = true
				}
			} else if g.At(x+1+i, y).R == 0 {
				// keep blank
			}
		}
		if k == o.Here {
			// mark the label as selected: the renderer paints Selection behind RoleSelectionText.
			for i := 0; i < len(label); i++ {
				g.Cells[y*g.W+x+1+i].Role = theme.RoleSelectionText
			}
		}
		if o.Changed[k] && k != o.Here {
			if xx := x + 1 + len(label); xx < w {
				g.Set(xx, y, '*', theme.RoleHighlight)
			}
		}
		if o.Counts && !p.Path {
			info := []rune(fmt.Sprintf("%d", p.Total()))
			if len(p.Holdings) == 0 {
				info = nil
			}
			for i, r := range info {
				if i < o.LabelW {
					g.Set(x+1+i, y+1, r, theme.RoleDim)
				}
			}
			mx := x + 1 + len(info) + 1
			put := func(r rune, role theme.Role) {
				if mx < x+1+o.LabelW {
					g.Set(mx, y+1, r, role)
					mx++
				}
			}
			if st := p.Staffing(); st >= 0 && st < 0.6 {
				put('!', theme.RoleWarning)
			}
			if p.Ruins() > 0 {
				put('x', theme.RoleNegative)
			}
			if k == Square && c.St.Harbinger != nil {
				put('?', theme.RoleWarning)
			}
			if k == Gate && len(c.atWar()) > 0 {
				put('×', theme.RoleNegative)
			}
			if k == Gate && c.St.Military.ActiveScout != nil {
				put('→', theme.RoleHighlight)
			}
		}
	}
	return g
}
