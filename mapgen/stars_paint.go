package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The territory of the people on the star map, painted over the galaxy: each
// building line's sector (hub, systems), the faint border round the lot, the
// civilizations met in clusters of their own, the lanes that join hubs, and
// the rift a harbinger opens. All glyphs, no words; the names are the
// inspect line's business.

// systemGlyphs are the systems of a sector, dimmest first.
var systemGlyphs = [4]rune{'∙', '•', '✦', '*'}

// lit is a sector with something built: the cells that shine, in order.
type lit struct {
	line  string
	idx   int
	count int
	cells []Pt // lit cells, hub first
	tier  int  // 0..3, the rank of the best building standing
	fresh bool // something of it grew since the last visit
}

// homeSectors lists the sectors that have something standing, in line order.
func (f *frame) homeSectors() []lit {
	var out []lit
	for _, l := range f.m.Lineages {
		i, ok := lineOf[l.Key]
		if !ok || l.Count <= 0 || len(f.plan.Sectors) <= i {
			continue
		}
		sec := f.plan.Sectors[i]
		n := min(systemsFor(l.Count, sectorMax), len(sec.Cells))
		tier := 0
		if kinds := len(f.m.Catalog.ByLineage[l.Key]); kinds > 1 && len(l.Buildings) > 0 {
			tier = clamp(l.Buildings[len(l.Buildings)-1].Rank*3/(kinds-1), 0, 3)
		}
		grew := slices.ContainsFunc(l.Buildings, func(b *mapmodel.Building) bool { return b.Delta > 0 })
		out = append(out, lit{line: l.Key, idx: i, count: l.Count, cells: sec.Cells[:n], tier: tier, fresh: grew})
	}
	return out
}

// rivalSite is a civilization met and the cluster it stands in.
type rivalSite struct {
	civ   *mapmodel.Faction
	cells []Pt
}

// rivalSites places every civilization met in a slot of its own, picked by its
// key and moved on to the next free one when taken.
func (f *frame) rivalSites() []rivalSite {
	var out []rivalSite
	n := len(f.plan.Rivals)
	if n == 0 {
		return nil
	}
	taken := make([]bool, n)
	for i := range f.m.Factions {
		civ := &f.m.Factions[i]
		if !civ.Discovered {
			continue
		}
		s := roll(n, f.m.Seed, 100, mapmodel.HashStr(civ.Key))
		for try := 0; try < n && taken[s]; try++ {
			s = (s + 1) % n
		}
		if taken[s] {
			continue
		}
		taken[s] = true
		cells := f.plan.Rivals[s].Cells
		out = append(out, rivalSite{civ, cells[:min(len(cells), 4+3*clamp(civ.Strength, 1, 5))]})
	}
	return out
}

// paintTerritory draws the people's part of the galaxy over its sky.
func (f *frame) paintTerritory() {
	homes := f.homeSectors()
	rivals := f.rivalSites()
	f.paintRelics() // first, so that nothing here covers one; they keep off the places sectors can take
	f.paintBorders(homes, rivals)
	f.paintLanes(homes, rivals)
	for _, h := range homes {
		f.paintSector(h)
	}
	for _, r := range rivals {
		f.paintRival(r)
	}
	f.paintRift()
	f.paintSkyLife()
}

// halo marks the cells within a short oval of any cell of a set.
func halo(cells []Pt, rx int, into []bool) {
	for _, c := range cells {
		for dy := -1; dy <= 1; dy++ {
			for dx := -rx; dx <= rx; dx++ {
				if abs(dx) == rx && dy != 0 {
					continue
				}
				if x, y := c.X+dx, c.Y+dy; inWorld(x, y) {
					into[at(x, y)] = true
				}
			}
		}
	}
}

// edgeOf lists the cells of a halo that touch outside it: an outline one cell
// wide.
func edgeOf(in []bool) []Pt {
	var out []Pt
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			if !in[at(x, y)] {
				continue
			}
			if x == 0 || y == 0 || x == worldW-1 || y == worldH-1 || !in[at(x-1, y)] || !in[at(x+1, y)] || !in[at(x, y-1)] || !in[at(x, y+1)] {
				out = append(out, Pt{x, y})
			}
		}
	}
	return out
}

// open sky is a cell nothing but the galaxy's own sky has claimed.
func (f *frame) sky(p Pt) bool {
	th := f.g.ThingAt(p.X, p.Y)
	return f.g.in(p.X, p.Y) && th.Kind == KindTerrain
}

// paintBorders outlines the home sectors as one territory, and each rival's
// cluster as its own.
func (f *frame) paintBorders(homes []lit, rivals []rivalSite) {
	if len(homes) > 0 {
		in := make([]bool, worldW*worldH)
		for _, h := range homes {
			halo(h.cells, 3, in)
		}
		bw := 160
		if f.close {
			bw = 95 // a faint edge, not another texture
		}
		ink := P(f.ink([]rune(f.st.Hub[0])[0], HNone).Hue, bw)
		for _, p := range edgeOf(in) {
			if f.sky(p) {
				f.g.draw(p.X, p.Y, '·', ink)
				f.g.mark(p.X, p.Y, Thing{Kind: KindSector, Key: "border"})
			}
		}
	}
	rw := 190
	if f.close {
		rw = 120
	}
	for _, r := range rivals {
		in := make([]bool, worldW*worldH)
		halo(r.cells, 2, in)
		for _, p := range edgeOf(in) {
			if (p.X+p.Y)%2 == 0 && f.sky(p) {
				f.g.draw(p.X, p.Y, '·', P(relationHue[r.civ.Relation], rw))
				f.g.mark(p.X, p.Y, Thing{Kind: KindCiv, Key: r.civ.Key})
			}
		}
	}
}

// clearCore takes the core's glow from under a system: sectors win, so a
// cluster stands on the dark and not on the tan.
func (f *frame) clearCore(p Pt) {
	if c := &f.g.Cells[at(p.X, p.Y)]; c.Bg.Hue == HCore {
		c.Bg = P(HNebula, int(c.Bg.Weight)/4)
	}
}

// paintSector lights the systems of one line's sector and its hub.
func (f *frame) paintSector(h lit) {
	hue := lineHue(h.line)
	n := len(h.cells)
	for j, p := range h.cells {
		f.clearCore(p)
		if f.fresh && h.fresh {
			f.g.Cells[at(p.X, p.Y)].Bg = freshTint
		}
		if j == 0 {
			f.g.draw(p.X, p.Y, '◉', P(hue, f.breathe(255, 55, int(p.X))))
			f.g.mark(p.X, p.Y, Thing{Kind: KindSector, Key: h.line, Lineage: h.line, Count: n, Ref: 1})
			f.targets = append(f.targets, p)
			continue
		}
		g := h.tier
		if j > n/3 {
			g--
		}
		if j > 2*n/3 {
			g--
		}
		g = clamp(g, 0, 3)
		// a hashed nudge in glyph and brightness: a cluster, not a pattern
		g = clamp(g+roll(3, f.m.Seed, 370, int64(p.X), int64(p.Y))-1, 0, 3)
		w := 110 + 30*g + roll(70, f.m.Seed, 371, int64(p.X), int64(p.Y))
		if f.close {
			w = 215 + roll(41, f.m.Seed, 371, int64(p.X), int64(p.Y)) // the brightest things on the screen
		}
		f.g.draw(p.X, p.Y, systemGlyphs[g], P(hue, w))
		f.g.mark(p.X, p.Y, Thing{Kind: KindSector, Key: h.line, Lineage: h.line, Count: n})
	}
}

// paintRival lights a civilization's cluster: a hub and a few systems.
func (f *frame) paintRival(r rivalSite) {
	hue := relationHue[r.civ.Relation]
	thing := Thing{Kind: KindCiv, Key: r.civ.Key}
	for j, p := range r.cells {
		f.clearCore(p)
		glyph, w := '∙', 130+roll(80, f.m.Seed, 372, int64(p.X), int64(p.Y))
		if f.close {
			w = 200 + roll(41, f.m.Seed, 372, int64(p.X), int64(p.Y))
		}
		if j == 0 {
			glyph, w = '◉', f.breathe(255, 55, int(p.X))
		} else if j < len(r.cells)/3 {
			glyph = '•'
		}
		f.g.draw(p.X, p.Y, glyph, P(hue, w))
		f.g.mark(p.X, p.Y, thing)
	}
	f.targets = append(f.targets, r.cells[0])
}

type edge struct{ a, b, d int }

// paintLanes joins the hubs: every hub to its nearest neighbours so the web
// holds together, and home to every civilization it has a route to.
func (f *frame) paintLanes(homes []lit, rivals []rivalSite) {
	// node 0 is the home star; the rest are the hubs of the lit sectors
	pts := []Pt{f.plan.Centre}
	for _, h := range homes {
		pts = append(pts, h.cells[0])
	}
	var all []edge
	for a := range pts {
		for b := a + 1; b < len(pts); b++ {
			all = append(all, edge{a, b, sq(pts[a].X-pts[b].X) + 4*sq(pts[a].Y-pts[b].Y)})
		}
	}
	slices.SortStableFunc(all, func(x, y edge) int { return cmp.Compare(x.d, y.d) })
	set := map[[2]int]bool{}
	comp := make([]int, len(pts))
	for i := range comp {
		comp[i] = i
	}
	for _, e := range all { // a minimum span first, so nothing is left alone
		if comp[e.a] != comp[e.b] {
			from, to := comp[e.b], comp[e.a]
			for i := range comp {
				if comp[i] == from {
					comp[i] = to
				}
			}
			set[[2]int{e.a, e.b}] = true
		}
	}
	for a := range pts { // then each hub's two nearest, for loops in the web
		near := 0
		for _, e := range all {
			if (e.a == a || e.b == a) && near < 2 {
				set[[2]int{e.a, e.b}] = true
				near++
			}
		}
	}
	name := func(i int) (string, int) {
		if i == 0 {
			return "home", 0
		}
		return homes[i-1].line, homes[i-1].count
	}
	for _, e := range all {
		if !set[[2]int{e.a, e.b}] {
			continue
		}
		ka, ca := name(e.a)
		kb, cb := name(e.b)
		if e.a == 0 {
			ca = cb
		}
		t := 1
		switch m := min(ca, cb); {
		case m >= 40:
			t = 3
		case m >= 8:
			t = 2
		}
		f.paintLane(pts[e.a], pts[e.b], t, HStar, Thing{Kind: KindLane, Key: ka, Lineage: kb, Count: t})
	}
	tradeBuilt := 0
	for _, l := range f.m.Lineages {
		if l.Key == mapmodel.LinTrade || l.Key == mapmodel.LinHarbor {
			tradeBuilt += l.Count
		}
	}
	laid := map[string]bool{}
	for _, route := range f.m.Routes {
		if laid[route.Civ] {
			continue
		}
		for _, r := range rivals {
			if r.civ.Key != route.Civ {
				continue
			}
			laid[route.Civ] = true
			score := tradeBuilt/2 + r.civ.TradeCount*3 + r.civ.Strength
			t := 1
			switch {
			case score >= 24:
				t = 3
			case score >= 8:
				t = 2
			}
			f.paintLane(f.plan.Centre, r.cells[0], t, relationHue[r.civ.Relation], Thing{Kind: KindLane, Key: r.civ.Key, Lineage: "home", Count: t, Ref: 1})
		}
	}
}

// paintLane draws a line between two cells in line work. Traffic 1 is a dotted
// path, 2 a plain line, 3 a double or brighter one.
func (f *frame) paintLane(a, b Pt, traffic int, hue Hue, th Thing) {
	cells := straight(a, b)
	weight := [4]int{0, 90, 150, 235}[traffic]
	for i, c := range cells {
		if i == 0 || i == len(cells)-1 || !f.sky(c) {
			continue
		}
		dx, dy := cells[i+1].X-c.X, cells[i+1].Y-c.Y
		var r rune
		switch {
		case traffic == 1:
			r = '·'
		case dy == 0:
			r = '─'
		case dx == 0:
			r = '│'
		case dx*dy > 0:
			r = '╲'
		default:
			r = '╱'
		}
		if traffic == 3 {
			switch r {
			case '─':
				r = '═'
			case '│':
				r = '║'
			}
		}
		f.g.draw(c.X, c.Y, r, P(hue, weight))
		f.g.mark(c.X, c.Y, th)
	}
}

// straight is the cells of a line from a to b, a first, one step at a time in
// the longer direction.
func straight(a, b Pt) []Pt {
	dx, dy := b.X-a.X, b.Y-a.Y
	n := max(abs(dx), abs(dy))
	out := make([]Pt, 0, n+1)
	for i := 0; i <= n; i++ {
		if n == 0 {
			out = append(out, a)
			break
		}
		x := a.X + (dx*i*2+n*sign(dx))/(2*n)
		y := a.Y + (dy*i*2+n*sign(dy))/(2*n)
		out = append(out, Pt{x, y})
	}
	return out
}

func sign(v int) int { return cmp.Compare(v, 0) }

// paintRift cracks the galaxy's edge when a harbinger is near: a jagged line
// of slashes in dim red, with dark cells where it opens.
func (f *frame) paintRift() {
	h := f.m.Harbinger
	if h == nil {
		return
	}
	rift := f.plan.Rift
	thing := Thing{Kind: KindHarbinger, Key: h.Key}
	for i, c := range rift {
		var r rune
		next := c
		if i+1 < len(rift) {
			next = rift[i+1]
		} else if i > 0 {
			next = Pt{2*c.X - rift[i-1].X, 2*c.Y - rift[i-1].Y}
		}
		dx, dy := next.X-c.X, next.Y-c.Y
		switch {
		case dx == 0:
			r = '│'
		case dx*dy > 0:
			r = '╲'
		default:
			r = '╱'
		}
		if i < len(rift)/3 { // the rim: black where it opens
			for _, w := range []int{-1, 1} {
				if inWorld(c.X+w, c.Y) && f.sky(Pt{c.X + w, c.Y}) {
					f.g.Cells[at(c.X+w, c.Y)].Bg = P(HVoid, 255)
				}
			}
		}
		if f.sky(c) {
			f.g.draw(c.X, c.Y, r, P(HRed, 150+i))
			f.g.Cells[at(c.X, c.Y)].Bg = P(HVoid, 255)
			f.g.mark(c.X, c.Y, thing)
		}
	}
}
