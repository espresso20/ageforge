package roguelike

import (
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// plates.go draws the region zoom on the epoch's cartographic plate: the
// same world, drawn the way maps were drawn in that era. A plate is a row
// of dials; the few furniture pieces (frames, rhumb lines, a compass, a
// radar sweep, the planet's rings) are small functions. Glyphs and theme
// classes only, no raster.

const (
	frNone uint8 = iota
	frHide
	frDouble
	frNeat
	frSurvey
	frHUD
)

const (
	fogSmudge uint8 = iota
	fogDragons
	fogBlank
	fogHatch
)

type plate struct {
	name, title                         string
	frame, fog, coast                   uint8 // coast: 0 none, 1 line, 2 bold line
	coastC                              mapmodel.Class
	marks, sea                          float64 // density of land marks and sea strokes
	lines, hachure, contour, biome      bool    // water-lining, hill strokes, relief lines, colour cells
	night, space, rhumbs, compass, refs bool
	scan, sweep                         bool
	grid                                [2]int // graticule spacing in cells
}

var plates = [7]plate{
	{name: "charcoal on hide", title: "the lands we know, drawn by the fire of %s", frame: frHide, fog: fogSmudge,
		marks: 0.35, sea: 0.08, coast: 1, coastC: mapmodel.CRock},
	{name: "portolan chart", title: "Carta of %s", frame: frDouble, fog: fogDragons, marks: 0.2, coast: 2,
		coastC: mapmodel.CWall, rhumbs: true},
	{name: "engraved map", title: "A New and Accurate Map of %s", frame: frNeat, fog: fogBlank, marks: 0.45,
		coast: 1, coastC: mapmodel.CText, lines: true, hachure: true, compass: true, grid: [2]int{16, 8}},
	{name: "survey sheet", title: "Survey of %s", frame: frSurvey, fog: fogHatch, marks: 0.25, coast: 1,
		coastC: mapmodel.CWater, contour: true, refs: true, grid: [2]int{20, 8}},
	{name: "satellite mosaic", title: "EARTH OBSERVATION  %s", frame: frHUD, fog: fogBlank, marks: 0.12, biome: true, scan: true},
	{name: "orbital night pass", title: "NIGHTSIDE  %s", frame: frHUD, fog: fogBlank, coast: 1,
		coastC: mapmodel.CMemory, night: true, sweep: true},
	{name: "planet from orbit", title: "%s from orbit", fog: fogBlank, marks: 0.5, sea: 0.3, space: true},
}

type regionBuf struct {
	land, off []bool
	terr      []mapmodel.Terrain
	vis, lvl  []uint8
	dist      []int8
	st        []glyph
	over      []rune
	oc        []mapmodel.Class
	queue     []int
}

func (b *regionBuf) reset(n int) {
	if cap(b.land) < n {
		*b = regionBuf{land: make([]bool, n), off: make([]bool, n), terr: make([]mapmodel.Terrain, n),
			vis: make([]uint8, n), lvl: make([]uint8, n), dist: make([]int8, n), st: make([]glyph, n),
			over: make([]rune, n), oc: make([]mapmodel.Class, n), queue: make([]int, 0, n)}
	}
	b.land, b.off, b.terr, b.vis, b.lvl = b.land[:n], b.off[:n], b.terr[:n], b.vis[:n], b.lvl[:n]
	b.dist, b.st, b.over, b.oc = b.dist[:n], b.st[:n], b.over[:n], b.oc[:n]
	for i := 0; i < n; i++ {
		b.over[i], b.st[i].r = 0, 0
	}
}

var terrRank = [8]int{0, 0, 6, 2, 1, 3, 4, 5} // deep, shallow, river, beach, grass, forest, hills, mountain

// sample fills the region buffer: per cell, land or sea, the dominant
// terrain, what is known and the most salient structure; then the distance
// to land over the sea, for water-lining.
func (v *view) sample(g geom) {
	s, rb := v.sc, &v.rbuf
	n := g.w * g.h
	rb.reset(n)
	for i := 0; i < n; i++ {
		x0, y0 := g.vx+i%g.w*g.scale, g.vy+i/g.w*g.scale
		land, tot, best, vis := 0, 0, mapmodel.TDeep, uint8(0)
		stg := glyph{sal: -1}
		for j := 0; j < g.scale*g.scale; j++ {
			x, y := x0+j%g.scale, y0+j/g.scale
			if !s.in(x, y) {
				continue
			}
			k := y*s.w.W + x
			t := s.w.T[k]
			tot++
			if t.Land() || t == mapmodel.TRiver {
				land++
			}
			if terrRank[t] > terrRank[best] || best == mapmodel.TDeep && t == mapmodel.TShallow {
				best = t
			}
			vis = max(vis, s.vis[k])
			if c := s.cells[k]; c.k != kNone && s.vis[k] == 2 && (c.k != kSite || s.m.Factions[c.civ-1].Discovered) {
				if gl := v.tile(x, y); gl.sal >= 30 && gl.sal > stg.sal {
					stg = gl
				}
			}
		}
		rb.off[i], rb.land[i] = tot == 0, tot > 0 && land*2 >= tot
		if !rb.land[i] && best.Land() {
			best = mapmodel.TShallow
		}
		rb.terr[i], rb.vis[i], rb.lvl[i] = best, vis, 0
		if rb.land[i] {
			rb.lvl[i] = [8]uint8{1, 1, 1, 1, 1, 1, 2, 3}[best] // relief steps
		}
		if stg.sal >= 30 {
			rb.st[i] = stg
		}
	}
	rb.queue = rb.queue[:0]
	for i := 0; i < n; i++ {
		rb.dist[i] = 9
		if rb.land[i] {
			rb.dist[i], rb.queue = 0, append(rb.queue, i)
		}
	}
	for q := 0; q < len(rb.queue); q++ {
		i := rb.queue[q]
		for _, d := range dirs4 {
			x, y := i%g.w+d[0], i/g.w+d[1]
			if rb.dist[i] < 5 && x >= 0 && y >= 0 && x < g.w && y < g.h && rb.dist[y*g.w+x] > rb.dist[i]+1 {
				rb.dist[y*g.w+x] = rb.dist[i] + 1
				rb.queue = append(rb.queue, y*g.w+x)
			}
		}
	}
}

// around builds a 4-neighbour mask (N1 E2 S4 W8) of cells passing ok;
// off-area neighbours count as off.
func around(g geom, cx, cy int, off bool, ok func(i int) bool) int {
	m := 0
	for bit, d := range dirs4 {
		x, y := cx+d[0], cy+d[1]
		if x < 0 || y < 0 || x >= g.w || y >= g.h {
			if off {
				m |= 1 << bit
			}
		} else if ok(y*g.w + x) {
			m |= 1 << bit
		}
	}
	return m
}

var coastRunes = [16]rune{' ', '─', '│', '╮', '─', '─', '╯', '·', '│', '╭', '│', '·', '╰', '·', '·', '·'}

// plateTerrain draws one known cell's ground in the plate's manner.
func (v *view) plateTerrain(p *plate, g geom, cx, cy int) glyph {
	rb, pal := &v.rbuf, v.pal
	i := cy*g.w + cx
	t, land := rb.terr[i], rb.land[i]
	h := mapmodel.HashF(v.sc.w.Seed, 13, int64(cx), int64(cy))
	out := glyph{r: ' ', bg: pal.tint}
	if out.bg == pal.Bg {
		out.bg = 0
	}
	switch {
	case p.biome:
		out.bg = pal.biome[t]
		if land && (t == mapmodel.TMountain || t == mapmodel.THills && h < 0.3) {
			out.r, out.c = '^', mapmodel.CRock
		} else if land && t == mapmodel.TForest && h < p.marks {
			out.r, out.c = '♣', mapmodel.CFlora
		}
		return out
	case !land:
		if p.space || p.contour {
			out.bg = pal.WaterBg
		}
		switch {
		case p.night:
		case p.lines && rb.dist[i] == 2:
			out.r, out.c = v.waterLine(g, cx, cy), mapmodel.CWater
		case h < p.sea:
			out.r, out.c, out.bg = '~', mapmodel.CWater, pal.WaterBg
			if p.space && mapmodel.Hash(int64(cx), int64(cy), int64(v.anim/2))%9 != 0 {
				out.r = '≈'
			}
		}
		return out
	}
	lv := rb.lvl[i]
	if p.contour && lv >= 2 { // relief steps as lines joining the same step
		onStep := func(j int) bool {
			return rb.lvl[j] == lv && around(g, j%g.w, j/g.w, false, func(k int) bool { return rb.lvl[k] < lv }) != 0
		}
		if onStep(i) {
			out.r, out.c = boxLight[around(g, cx, cy, false, onStep)], mapmodel.CHill
			if lv == 3 {
				out.c, out.attr = mapmodel.CRock, tcell.AttrBold
			}
			return out
		}
	}
	if mk := around(g, cx, cy, true, func(j int) bool { return !rb.land[j] }); p.coast > 0 && mk != 0 {
		out.r, out.c = coastRunes[mk], p.coastC
		if p.coast == 2 {
			out.attr = tcell.AttrBold
		}
		return out
	}
	switch {
	case p.night:
		if h < 0.04 {
			out.r, out.c = '·', mapmodel.CMemory
		}
	case t == mapmodel.TRiver:
		out.r, out.c = '~', mapmodel.CWater
	case t == mapmodel.TMountain:
		out.r, out.c = '▲', mapmodel.CRock
		if v.sc.epoch < 2 {
			out.r = '^'
		}
	case t == mapmodel.THills && p.hachure: // strokes down the slope
		wl, el := cx > 0 && rb.lvl[i-1] < lv, cx+1 < g.w && rb.lvl[i+1] < lv
		out.c = mapmodel.CHill
		switch {
		case wl && !el:
			out.r = '╱'
		case el && !wl:
			out.r = '╲'
		case h < 0.5:
			out.r = '^'
		}
	case t == mapmodel.THills && h < p.marks:
		out.r, out.c = '^', mapmodel.CHill
		if v.sc.epoch == 0 {
			out.r = '∩'
		}
	case t == mapmodel.TForest && h < p.marks:
		out.r, out.c = '♣', mapmodel.CFlora
	case h < 0.05 && p.frame == frHide:
		out.r, out.c = ',', mapmodel.CGround
	}
	return out
}

// waterLine is a stroke parallel to the nearest coast.
func (v *view) waterLine(g geom, cx, cy int) rune {
	d := func(x, y int) int {
		if x < 0 || y < 0 || x >= g.w || y >= g.h {
			return 9
		}
		return int(v.rbuf.dist[y*g.w+x])
	}
	gx, gy := d(cx+1, cy)-d(cx-1, cy), d(cx, cy+1)-d(cx, cy-1)
	switch {
	case gx == 0 && gy == 0:
		return '·'
	case gx == 0 || abs(gy) > 2*abs(gx):
		return '─'
	case gy == 0 || abs(gx) > 2*abs(gy):
		return '│'
	case gx*gy > 0:
		return '╱'
	}
	return '╲'
}

func abs(a int) int { return max(a, -a) }

// fogGlyph is how the plate draws what nobody has seen.
func (v *view) fogGlyph(p *plate, g geom, cx, cy int) glyph {
	h := mapmodel.HashF(v.sc.w.Seed, 14, int64(cx), int64(cy))
	out := glyph{r: ' ', c: mapmodel.CMemory}
	switch p.fog {
	case fogSmudge: // a charcoal smudge where knowledge gives out
		near := false
		for j := 0; j < 9 && !near; j++ {
			x, y := cx+j%3-1, cy+j/3-1
			near = x >= 0 && y >= 0 && x < g.w && y < g.h && v.rbuf.vis[y*g.w+x] > 0
		}
		if near && h < 0.3 {
			out.r = '░'
		} else if near && h < 0.6 {
			out.r = '·'
		}
	case fogDragons, fogBlank:
		if h < 0.05 {
			out.r = '·'
		}
	case fogHatch:
		if (cx+2*cy)%9 == 0 {
			out.r = '╱'
		}
	}
	return out
}

// overlays lays the plate's line work (graticule, rhumbs, sweep, halos)
// into the overlay buffer; it only shows on blank cells.
func (v *view) overlays(p *plate, g geom) {
	rb := &v.rbuf
	set := func(x, y int, r rune, c mapmodel.Class) {
		if x >= 0 && y >= 0 && x < g.w && y < g.h {
			rb.over[y*g.w+x], rb.oc[y*g.w+x] = r, c
		}
	}
	// ray draws from (ox, oy) at a turns; sea rays skip known land and dot
	// their minor winds.
	ray := func(ox, oy int, a float64, n int, r rune, c mapmodel.Class, sea bool) {
		step := 1
		if r == '·' && sea {
			step = 2
		}
		for t := 1; t < n; t += step {
			x := ox + int(math.Round(float64(mapmodel.Cos(a)*float64(2*t))))
			y := oy + int(math.Round(float64(mapmodel.Sin(a)*float64(t))))
			if !sea || x < 0 || y < 0 || x >= g.w || y >= g.h || !rb.land[y*g.w+x] || rb.vis[y*g.w+x] == 0 {
				set(x, y, r, c)
			}
		}
	}
	if gx, gy := p.grid[0], p.grid[1]; gx > 0 {
		c := mapmodel.CMemory
		if p.refs {
			c = mapmodel.CWater
		}
		for i := 0; i < g.w*g.h; i++ {
			vx, hy := i%g.w%gx == gx/2, i/g.w%gy == gy/2
			switch {
			case vx && hy:
				set(i%g.w, i/g.w, '┼', c)
			case vx:
				set(i%g.w, i/g.w, '┊', c)
			case hy:
				set(i%g.w, i/g.w, '┈', c)
			}
		}
	}
	if p.rhumbs {
		for _, o := range v.roses(g) {
			for k := 0; k < 16; k++ {
				r, c := '·', mapmodel.CMemory
				switch {
				case k%8 == 0:
					r, c = '─', mapmodel.CHill
				case k%8 == 4:
					r, c = '│', mapmodel.CHill
				case k == 2 || k == 10:
					r, c = '╲', mapmodel.CHill
				case k%4 == 2:
					r, c = '╱', mapmodel.CHill
				}
				ray(o.X, o.Y, float64(k)/16, max(g.w, g.h)*2/3, r, c, true)
			}
		}
	}
	if p.night { // each civ glows in its relation colour
		for _, f := range v.sc.m.Factions {
			if cx, cy, ok := v.siteCell(f); ok {
				for j := 0; j < 21; j++ {
					if x, y := cx+j%7-3, cy+j/7-1; mapmodel.HashF(int64(x), int64(y), 21) < 0.5 {
						set(x-g.x, y-g.y, '∙', mapmodel.RelationClass(f.Relation))
					}
				}
			}
		}
	}
	if ox, oy, ok := g.cellOf(pt(v.sc.w.CX, v.sc.w.CY)); p.sweep && ok { // the radar sweep
		a := float64(v.anim%240) / 240
		for k := 2; k >= 0; k-- {
			c := mapmodel.CMemory
			if k == 0 {
				c = mapmodel.CFresh
			}
			ray(ox-g.x, oy-g.y, a-float64(k)*0.012, min(g.w/2, g.h)*2/3, '·', c, false)
		}
	}
}

func (v *view) siteCell(f mapmodel.Faction) (int, int, bool) {
	if p, ok := v.sc.site(&f); ok && f.Discovered && v.sc.seen(p) == 2 {
		return v.g.cellOf(p)
	}
	return 0, 0, false
}

// roses are the portolan's wind roses: one or two, off known land.
func (v *view) roses(g geom) []mapmodel.Pt {
	var out []mapmodel.Pt
	for _, c := range [2][2]int{{g.w / 4, g.h / 3}, {g.w * 3 / 4, g.h * 2 / 3}} {
		bx, by, best := -1, -1, 1<<30
		for dy := -g.h / 5; dy <= g.h/5; dy++ {
			for dx := -g.w / 6; dx <= g.w/6; dx++ {
				x, y := c[0]+dx, c[1]+dy
				if x < 1 || y < 1 || x >= g.w-1 || y >= g.h-1 || v.rbuf.land[y*g.w+x] && v.rbuf.vis[y*g.w+x] > 0 {
					continue
				}
				if d := abs(dx)/2 + abs(dy); d < best {
					bx, by, best = x, y, d
				}
			}
		}
		if bx >= 0 {
			out = append(out, pt(bx, by))
		}
	}
	return out
}

// drawRegion draws the region zoom on the epoch's plate.
func (v *view) drawRegion(cv *mapstyle.Canvas, x, y, w, h int) {
	s := v.sc
	p := &plates[s.epoch]
	ix, iy, iw, ih := x, y, w, h
	if (p.frame == frDouble || p.frame == frNeat || p.frame == frSurvey) && w >= 12 && h >= 6 {
		ix, iy, iw, ih = x+1, y+1, w-2, h-2
	}
	g := v.fit(ix, iy, iw, ih, zRegion, s, p.space)
	v.g = g
	v.resetOcc(iw * ih)
	v.sample(g)
	v.overlays(p, g)
	rb := &v.rbuf
	wcx, wcy := float64(s.w.W/2-g.vx)/float64(g.scale), float64(s.w.H/2-g.vy)/float64(g.scale)
	rx, ry := float64(s.w.W)/float64(2*g.scale), float64(s.w.H)/float64(2*g.scale)
	scanRow := -1
	if p.scan {
		scanRow = (v.anim / 2) % ih
	}
	for i := 0; i < iw*ih; i++ {
		cx, cy := i%iw, i/iw
		nx, ny := (float64(cx)+0.5-wcx)/rx, (float64(cy)+0.5-wcy)/ry
		d2 := float64(nx*nx) + float64(ny*ny)
		outside := p.frame == frHide && !v.inHide(g, cx, cy) || p.space && d2 > 1
		var gl glyph
		switch {
		case p.frame == frHide && outside:
			gl = glyph{r: ' '}
			if (cx+cy)%2 == 0 && around(g, cx, cy, false, func(j int) bool { return v.inHide(g, j%g.w, j/g.w) }) != 0 {
				gl = glyph{r: '·', c: mapmodel.CRock} // the hide's stitched edge
			}
		case outside:
			gl = v.space(cx, cy, math.Sqrt(d2), ry)
		case rb.off[i]:
			gl = glyph{r: ' '}
		case rb.st[i].r != 0:
			gl = rb.st[i]
			if p.night {
				gl = nightLight(gl, v.anim, cx, cy)
			}
			if p.biome {
				gl.bg = v.pal.biome[rb.terr[i]]
			}
		case rb.vis[i] == 0:
			gl = v.fogGlyph(p, g, cx, cy)
		default:
			gl = v.plateTerrain(p, g, cx, cy)
			if rb.vis[i] == 1 && gl.r != ' ' {
				gl.c, gl.attr, gl.fg = mapmodel.CMemory, 0, 0
			}
		}
		if gl.r == ' ' && rb.over[i] != 0 && !outside {
			gl.r, gl.c = rb.over[i], rb.oc[i]
		}
		if cy == scanRow {
			gl.bg = v.pal.scan
		}
		v.occ[i] = gl.sal >= 40
		cv.Put(ix+cx, iy+cy, gl.r, v.style(gl))
	}
	if p.space {
		v.orbit(cv, g, wcx, wcy, rx, ry)
	}
	v.drawLife(cv)
	v.drawLabels(cv)
	v.furniture(cv, p, g)
	v.frame(cv, p, x, y, w, h, g)
}

// nightLight turns a structure into city lights.
func nightLight(g glyph, anim, cx, cy int) glyph {
	switch {
	case g.sal >= 100: // a civ keeps its mark
	case g.sal >= 55:
		g.r, g.c, g.fg, g.bg, g.attr = '•', mapmodel.CWealth, 0, 0, tcell.AttrBold
		if mapmodel.Hash(int64(cx), int64(cy), int64(anim/4))%11 == 0 {
			g.r = '∙'
		}
	case g.c == mapmodel.CRoad:
		g.r, g.bg = '·', 0
	default:
		g.r, g.c, g.fg, g.bg, g.attr = '∙', mapmodel.CWealth, 0, 0, 0
	}
	return g
}

// inHide is the stretched hide's ragged outline.
func (v *view) inHide(g geom, cx, cy int) bool {
	nx := (float64(cx)+0.5)/float64(g.w)*2 - 1
	ny := (float64(cy)+0.5)/float64(g.h)*2 - 1
	x2, y2 := float64(nx*nx), float64(ny*ny)
	return float64(x2*x2)+float64(y2*y2)+float64(0.22*(mapmodel.Noise(v.sc.w.Seed+9, float64(cx)/5, float64(cy)/2.5)-0.5)) < 0.92
}

// space is the sky round the planet: stars and the polar rings.
func (v *view) space(cx, cy int, d, ry float64) glyph {
	tol := 0.5 / math.Max(ry, 1)
	switch {
	case math.Abs(d-1.18) < tol:
		return glyph{r: '·', c: mapmodel.CCivic}
	case math.Abs(d-1.5) < tol && (cx+cy)%2 == 0:
		return glyph{r: '·', c: mapmodel.CMemory}
	}
	h := mapmodel.Hash(int64(cx), int64(cy), 7)
	g := glyph{r: ' ', c: mapmodel.CStar}
	switch h % 37 {
	case 0:
		g.r = '·'
		if (v.anim/3+int(h>>8))%5 == 0 {
			g.r, g.attr = '*', tcell.AttrBold
		}
	case 1, 2:
		g.r = '·'
	case 3:
		g.r = '.'
	}
	return g
}

var starNames = []string{"Vega", "Altair", "Deneb", "Rigel", "Sirius", "Capella", "Arcturus", "Procyon", "Spica", "Antares"}

// orbit draws habitats riding the inner ring, traffic on the outer one and
// a few named stars.
func (v *view) orbit(cv *mapstyle.Canvas, g geom, wcx, wcy, rx, ry float64) {
	put := func(a, r float64, sym mapmodel.Sym, c mapmodel.Class) {
		x := g.x + int(math.Floor(wcx+float64(mapmodel.Cos(a)*float64(r*rx))))
		y := g.y + int(math.Floor(wcy+float64(mapmodel.Sin(a)*float64(r*ry))))
		if x >= g.x && y >= g.y && x < g.x+g.w && y < g.y+g.h {
			cv.Put(x, y, mapmodel.R(sym, v.tier), v.cls(c).Bold(true))
			v.occ[(y-g.y)*g.w+x-g.x] = true
		}
	}
	for k := 0; k < 4; k++ {
		if k < 3 {
			put(float64(v.anim)*0.002+float64(k)/3, 1.18, mapmodel.SymHabitat, mapmodel.CWealth)
		}
		put(-float64(v.anim)*0.004+float64(k)/4+0.1, 1.5, mapmodel.SymSatellite, mapmodel.CLife)
	}
	seed := v.sc.w.Seed
	for k := int64(0); k < 6; k++ {
		x := g.x + int(mapmodel.Hash(seed, 31, k)%uint64(max(1, g.w)))
		y := g.y + int(mapmodel.Hash(seed, 32, k)%uint64(max(1, g.h)))
		nx, ny := (float64(x-g.x)+0.5-wcx)/rx, (float64(y-g.y)+0.5-wcy)/ry
		if float64(nx*nx)+float64(ny*ny) < 2.6 || !v.free(x, y, 1) {
			continue
		}
		cv.Put(x, y, '✦', v.cls(mapmodel.CLabel).Bold(true))
		if name := starNames[mapmodel.Hash(seed, 33, k)%uint64(len(starNames))]; v.free(x+2, y, mapstyle.TextLen(name)) {
			cv.Text(x+2, y, 0, name, v.cls(mapmodel.CStar))
		}
	}
}

// furniture: the wind roses, the compass, the dragons, the serpent and the
// HUD's pass stamp.
func (v *view) furniture(cv *mapstyle.Canvas, p *plate, g geom) {
	rb := &v.rbuf
	put := func(x, y int, s string, st tcell.Style) bool {
		ok := v.free(g.x+x, g.y+y, mapstyle.TextLen(s))
		if ok {
			cv.Text(g.x+x, g.y+y, 0, s, st)
		}
		return ok
	}
	acc := v.cls(mapmodel.CAccent)
	for _, o := range v.roses(g) {
		if p.rhumbs && put(o.X, o.Y, "✦", acc.Bold(true)) {
			put(o.X, o.Y-1, "N", acc)
		}
	}
	for _, c := range [3][2]int{{2, g.h - 4}, {g.w - 5, g.h - 4}, {2, 1}} {
		if p.compass && c[0] >= 0 && c[1] >= 0 && c[0]+3 <= g.w && c[1]+3 <= g.h && !rb.land[(c[1]+1)*g.w+c[0]+1] &&
			put(c[0], c[1], " N ", acc) {
			put(c[0], c[1]+1, "W✦E", acc.Bold(true))
			put(c[0], c[1]+2, " S ", acc)
			break
		}
	}
	if p.fog == fogDragons {
		v.inRun(cv, g, "here be dragons", func(i int) bool { return rb.vis[i] == 0 }, v.cls(mapmodel.CDim).Italic(true))
		v.inRun(cv, g, "~^~^~o<", func(i int) bool { return !rb.land[i] && rb.dist[i] >= 2 && rb.vis[i] > 0 }, v.cls(mapmodel.CWater))
	}
	if p.frame == frHUD && g.h > 2 {
		stamp := "PASS " + strconv.Itoa(1+v.sc.m.Clock.Day%97) + "  " + v.sc.m.Clock.String()
		put(g.w-mapstyle.TextLen(stamp)-2, g.h-1, stamp, v.cls(mapmodel.CDim))
		if p.scan {
			put(2, g.h-1, "DOWNLINK", v.cls(mapmodel.CFresh))
		}
	}
}

// inRun writes text on the first horizontal run of cells that pass ok,
// scanning from a seeded row.
func (v *view) inRun(cv *mapstyle.Canvas, g geom, text string, ok func(i int) bool, st tcell.Style) {
	n := mapstyle.TextLen(text)
	if g.h <= 0 || n+2 > g.w {
		return
	}
	start := int(mapmodel.Hash(v.sc.w.Seed, 41) % uint64(g.h))
	for k := 0; k < g.h; k++ {
		cy, run := (start+k)%g.h, 0
		for cx := 0; cx < g.w; cx++ {
			if run++; !ok(cy*g.w + cx) {
				run = 0
			} else if run == n+2 && v.free(g.x+cx-n, g.y+cy, n) {
				cv.Text(g.x+cx-n, g.y+cy, n, text, st)
				return
			}
		}
	}
}

// frame draws the plate's border and title.
func (v *view) frame(cv *mapstyle.Canvas, p *plate, x, y, w, h int, g geom) {
	title := strings.Replace(p.title, "%s", v.sc.w.Name, 1)
	lab, bst := v.cls(mapmodel.CLabel), v.pal.border
	centre := func(row int, s string, st tcell.Style) {
		if n := mapstyle.TextLen(s); n <= w-4 {
			cv.Text(x+(w-n)/2, row, n, s, st)
		}
	}
	switch p.frame {
	case frDouble, frNeat, frSurvey:
		if g.x == x {
			return // no room for a frame
		}
		set := []rune("┌─┐│└┘")
		if p.frame == frDouble {
			set = []rune("╔═╗║╚╝")
		}
		for i := 0; i < w; i++ {
			cv.Put(x+i, y, set[1], bst)
			cv.Put(x+i, y+h-1, set[1], bst)
		}
		for j := 0; j < h; j++ {
			cv.Put(x, y+j, set[3], bst)
			cv.Put(x+w-1, y+j, set[3], bst)
		}
		for k, c := range [4][2]int{{x, y}, {x + w - 1, y}, {x, y + h - 1}, {x + w - 1, y + h - 1}} {
			cv.Put(c[0], c[1], set[[4]int{0, 2, 4, 5}[k]], bst)
		}
		for cx, gx := p.grid[0]/2, p.grid[0]; gx > 0 && cx < g.w-2; cx += gx {
			s := strconv.Itoa(abs(cx-g.w/2)/gx*10) + "°"
			if p.refs {
				s = pad2(cx / gx)
			}
			cv.Text(g.x+cx, y, 0, s, bst)
		}
		for cy, gy := p.grid[1]/2, p.grid[1]; p.refs && cy < g.h; cy += gy {
			cv.Text(x, g.y+cy, 1, strconv.Itoa(cy/gy%10), bst)
		}
		row := y
		if p.frame == frSurvey {
			row = y + h - 1
		}
		centre(row, " "+title+" ", lab.Bold(true))
	case frHUD:
		for k, c := range [4][2]int{{x, y}, {x + w - 2, y}, {x, y + h - 1}, {x + w - 2, y + h - 1}} {
			cv.Put(c[0], c[1], []rune("┌──┐└──┘")[2*k], bst)
			cv.Put(c[0]+1, c[1], []rune("┌──┐└──┘")[2*k+1], bst)
		}
		cv.Text(x+3, y, w-6, title, lab.Bold(true))
	case frHide:
		parts := strings.SplitN(title, ", ", 2)
		centre(y, parts[0], lab.Italic(true))
		if len(parts) > 1 && h > 2 {
			centre(y+h-1, parts[1], v.cls(mapmodel.CDim).Italic(true))
		}
	default:
		cv.Text(x+1, y, w-2, title, lab.Bold(true))
	}
}

func pad2(n int) string {
	if n < 10 && n >= 0 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
