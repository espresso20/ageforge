package mapstyle

import (
	"slices"

	"github.com/espresso20/ageforge/mapgen"
)

// snap.go is the space cursor. In the space scenes the arrows jump from body
// to body (the sun, a planet, a station, a gate, a star) instead of stepping
// cell by cell through empty sky, and the cursor is drawn as four corners
// round the body it is on. A body is a run of touching cells of one thing in
// the grid being shown, so the renderer's own picture is the only list.

// body is one thing the cursor can stand on: its box and the cell the cursor
// takes, in the cells of the grid shown.
type body struct {
	box mapgen.Rect
	at  mapgen.Pt
	th  mapgen.Thing
}

var (
	homeThing = mapgen.Thing{Kind: mapgen.KindTerrain, Key: "home"}
	sunThing  = mapgen.Thing{Kind: mapgen.KindTerrain, Key: "sun"}
)

// shownCell is the grid the cursor stands in and its cell there: the
// overview's at the squeezed zooms, the world's at the others.
func (v *view) shownCell(sc *mapgen.Scene) (g *mapgen.Grid, x, y int, ok bool) {
	if v.overview && v.ov != nil {
		x, y, ok = v.ov.CellOf(v.cur)
		return v.ov.Grid, x, y, ok
	}
	return sc.Grid, v.cur.X, v.cur.Y, v.cur.X >= 0 && v.cur.Y >= 0 && v.cur.X < sc.Grid.W && v.cur.Y < sc.Grid.H
}

// snap moves the cursor to the nearest body in the direction d (an arrow key,
// or a step of it); it stays put when there is none that way. It reports
// false, having done nothing, in a scene with no bodies at all.
func (v *view) snap(sc *mapgen.Scene, d [2]int) bool {
	g, x, y, ok := v.shownCell(sc)
	if !ok {
		return false
	}
	bs := bodiesOf(g)
	if !v.overview && v.win.W > 0 { // in the close-up, those in view
		bs = slices.DeleteFunc(bs, func(b body) bool {
			return b.box.X >= v.win.X+v.win.W || b.box.X+b.box.W <= v.win.X || b.box.Y >= v.win.Y+v.win.H || b.box.Y+b.box.H <= v.win.Y
		})
	}
	if len(bs) == 0 {
		return false
	}
	from := mapgen.Rect{X: x, Y: y, W: 1, H: 1}
	if b, ok := v.bodyAt(g, x, y); ok {
		from = b.box
	}
	if b, ok := nearestBody(bs, from, sign(d[0]), sign(d[1])); ok {
		v.cur = mapgen.Pt{X: b.at.X, Y: b.at.Y}
		if v.overview && v.ov != nil {
			v.cur = v.ov.Pick(b.at.X, b.at.Y)
		}
		return true
	}
	return v.overview // the close-up steps a cell, which scrolls the next bodies into view
}

// hop moves the cursor on to the next (dir 1) or previous body of the picture
// shown, the Tab key's step on the star map's close views.
func (v *view) hop(sc *mapgen.Scene, dir int) {
	bs := bodiesOf(v.ov.Grid)
	if len(bs) == 0 {
		return
	}
	v.inspect = true
	n := len(bs) - 1 // a step back from nowhere lands on the last
	if dir > 0 {
		n = 0
	}
	if g, x, y, ok := v.shownCell(sc); ok {
		if here, ok := v.bodyAt(g, x, y); ok {
			if i := slices.IndexFunc(bs, func(b body) bool { return b.th == here.th }); i >= 0 {
				n = ((i+dir)%len(bs) + len(bs)) % len(bs)
			}
		}
	}
	v.cur = v.ov.Pick(bs[n].at.X, bs[n].at.Y)
}

// pickable says whether a run of n cells of a thing is a body.
func pickable(th mapgen.Thing, n int) bool {
	switch th.Kind {
	case mapgen.KindTerrain:
		switch th.Key {
		case "sun":
			return th.Ref == 0 // not its halo
		case "planet", "home", "pluto", "comet", "space_station":
			return true
		case "star":
			return th.Count == 0 // not its halo
		case "neighbor", "deepsky":
			return true
		case "satellite", "the_moon":
			return n >= 4 // a moon drawn as a dot is not worth a stop
		}
	case mapgen.KindBuilding, mapgen.KindWonder, mapgen.KindCiv, mapgen.KindVisitor, mapgen.KindHarbinger:
		return !th.Ruin
	}
	return false
}

// bodyAt is the body that has cell (x, y) of g in it. seen, if given, marks
// the cells already gathered into a body.
func bodyAt(g *mapgen.Grid, x, y int, seen []bool) (body, bool) {
	b, ok, _ := bodyAtCells(g, x, y, seen)
	return b, ok
}

// bodyAt is the body at a cell, found on the view's own marks so a frame
// allocates nothing to ask; the marks are cleared again.
func (v *view) bodyAt(g *mapgen.Grid, x, y int) (body, bool) {
	if n := g.W * g.H; len(v.seen) < n {
		v.seen = make([]bool, n)
	}
	b, ok, cells := bodyAtCells(g, x, y, v.seen)
	for _, c := range cells {
		v.seen[c.Y*g.W+c.X] = false
	}
	return b, ok
}

// bodyAtCells is bodyAt that also says which cells it marked in seen.
func bodyAtCells(g *mapgen.Grid, x, y int, seen []bool) (body, bool, []mapgen.Pt) {
	hit := g.Hit[y*g.W+x]
	if hit == 0 || !pickable(g.Things[hit], 1<<30) {
		return body{}, false, nil
	}
	if seen == nil {
		seen = make([]bool, g.W*g.H)
	}
	cells := []mapgen.Pt{{X: x, Y: y}}
	seen[y*g.W+x] = true
	x0, y0, x1, y1 := x, y, x, y
	for i := 0; i < len(cells); i++ {
		c := cells[i]
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nx, ny := c.X+dx, c.Y+dy
				if nx < 0 || ny < 0 || nx >= g.W || ny >= g.H || seen[ny*g.W+nx] || g.Hit[ny*g.W+nx] != hit {
					continue
				}
				seen[ny*g.W+nx] = true
				cells = append(cells, mapgen.Pt{X: nx, Y: ny})
				x0, y0, x1, y1 = min(x0, nx), min(y0, ny), max(x1, nx), max(y1, ny)
			}
		}
	}
	th := g.Things[hit]
	if !pickable(th, len(cells)) {
		return body{}, false, cells
	}
	// the cursor takes the cell of the body nearest its middle
	best, gap := cells[0], 1<<60
	for _, c := range cells {
		dx, dy := 2*c.X-x0-x1, 2*(2*c.Y-y0-y1)
		if d := dx*dx + dy*dy; d < gap {
			best, gap = c, d
		}
	}
	return body{mapgen.Rect{X: x0, Y: y0, W: x1 - x0 + 1, H: y1 - y0 + 1}, best, th}, true, cells
}

// bodiesOf lists the bodies of g.
func bodiesOf(g *mapgen.Grid) []body {
	seen := make([]bool, g.W*g.H)
	var out []body
	for i, h := range g.Hit {
		if h != 0 && !seen[i] {
			if b, ok := bodyAt(g, i%g.W, i/g.W, seen); ok {
				out = append(out, b)
			}
		}
	}
	return out
}

// nearestBody is the body nearest the middle of from that lies within 45
// degrees of direction (dx, dy), measured the way the screen shows it (a
// cell is twice as tall as wide); ok is false when there is none.
func nearestBody(bs []body, from mapgen.Rect, dx, dy int) (best body, ok bool) {
	fx, fy := 2*from.X+from.W, 2*(2*from.Y+from.H) // doubled middle, rows doubled again
	gap := 1 << 60
	for _, b := range bs {
		if b.box == from {
			continue
		}
		rx, ry := 2*b.box.X+b.box.W-fx, 2*(2*b.box.Y+b.box.H)-fy
		along, across := rx*dx+ry*dy, abs(rx*dy)+abs(ry*dx) // across: distance from the key's axis
		if along <= 0 || across > along {
			continue
		}
		if d := rx*rx + ry*ry; d < gap {
			best, gap, ok = b, d, true
		}
	}
	return best, ok
}

func abs(v int) int { return max(v, -v) }

func sign(v int) int { return min(max(v, -1), 1) }
