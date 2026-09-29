package roguelike

import (
	"sort"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// draw.go draws the map area: the tiles at the current zoom, the moving
// life on top, labels on free cells and the cursor.

func (v *view) resetOcc(n int) {
	if cap(v.occ) < n {
		v.occ = make([]bool, n)
	}
	v.occ = v.occ[:n]
	clear(v.occ)
}

// drawTiles draws the settlement or district zoom, or a plate-less
// down-sampled view (the compact fallback) when g.zoom is zRegion.
func (v *view) drawTiles(cv *mapstyle.Canvas, g geom) {
	v.resetOcc(g.w * g.h)
	for i := 0; i < g.w*g.h; i++ {
		cx, cy := i%g.w, i/g.w
		var gl glyph
		switch {
		case g.zoom == zRegion: // the most salient tile in the block wins
			gl = glyph{r: ' ', sal: -2}
			for j := 0; j < g.scale*g.scale; j++ {
				if t := v.tile(g.vx+cx*g.scale+j%g.scale, g.vy+cy*g.scale+j/g.scale); t.sal > gl.sal {
					gl = t
				}
			}
		case g.cellW == 2:
			tx, ty := g.vx+cx/2, g.vy+cy
			gl = v.tile(tx, ty)
			if cx%2 == 1 {
				sal := gl.sal
				gl = v.rightHalf(tx, ty, gl)
				gl.sal = sal
			}
		default:
			gl = v.tile(g.vx+cx, g.vy+cy)
		}
		// district labels may sit on a street; elsewhere streets stay clear
		v.occ[i] = gl.sal >= 40 && !(g.cellW == 2 && gl.sal < 45)
		cv.Put(g.x+cx, g.y+cy, gl.r, v.style(gl))
	}
}

// on draws one moving thing over the map, keeping the cell's background.
// It never covers a building unless over is set.
func (v *view) on(cv *mapstyle.Canvas, p mapmodel.Pt, r rune, c mapmodel.Class, over bool, id lgID) {
	s := v.sc
	k := s.at(p.X, p.Y).k
	cx, cy, ok := v.g.cellOf(p)
	if !ok || s.seen(p) < 2 || !over && (k == kTile || k == kWonder || k == kSite || k == kCentre) {
		return
	}
	_, cur := cv.Get(cx, cy)
	_, bg, _ := cur.Decompose()
	g := glyph{r: r, c: c, bg: bg}
	if c == mapmodel.CLife || c == mapmodel.CIdle || c == mapmodel.CDanger {
		g.attr = tcell.AttrBold
	}
	if v.flows && id == lgIdle {
		g.bg = v.pal.FlowBg
	}
	st := v.style(g)
	cv.Put(cx, cy, r, st)
	v.reg(id, r, st)
}

// drawLife puts the moving things on top: smoke, workers, idle hands,
// caravans, war marks, scouts, raiders, the harbinger and the catastrophe.
func (v *view) drawLife(cv *mapstyle.Canvas) {
	s, m, f := v.sc, v.sc.m, v.anim
	sym := func(sy mapmodel.Sym) rune { return mapmodel.R(sy, v.tier) }
	if v.g.zoom != zRegion {
		puffs := []rune(s.d.smoke)
		for _, src := range s.smoke {
			if ph := (f/2 + src.ph) % 7; ph < 3 && len(puffs) > 0 {
				p := pt(src.p.X+ph/2, src.p.Y-1-ph)
				if k := s.at(p.X, p.Y).k; k == kNone || k == kStreet {
					v.on(cv, p, puffs[ph%len(puffs)], mapmodel.CMemory, false, lgSmoke)
				}
			}
		}
		for i, path := range s.walks { // there and back again
			n := len(path)
			t := (f/2 + int(mapmodel.Hash(int64(i), 3)%uint64(2*n-2))) % (2*n - 2)
			v.on(cv, path[min(t, 2*n-2-t)], sym(mapmodel.SymWorker), mapmodel.CLife, false, lgWorker)
		}
		for j, p := range s.idle {
			r := sym(mapmodel.SymWorker)
			if (f/6+j)%4 == 0 {
				r = sym(mapmodel.SymIdle)
			}
			v.on(cv, p, r, mapmodel.CIdle, true, lgIdle)
		}
	}
	for ri, rt := range m.Routes {
		fi := factionIndex(m, rt.Civ)
		if fi < 0 || len(s.trails[fi]) < 4 {
			continue
		}
		tr := s.trails[fi]
		if rt.Disrupted { // a stalled caravan
			v.on(cv, tr[len(tr)/3], sym(mapmodel.SymCaravan), mapmodel.CDanger, false, lgStalled)
			continue
		}
		for c := 0; c < 2; c++ {
			v.on(cv, tr[(f/2+c*len(tr)/2+ri*7)%len(tr)], sym(mapmodel.SymCaravan), mapmodel.CWealth, false, lgCaravan)
		}
	}
	for fi, fc := range m.Factions {
		if tr := s.trails[fi]; fc.Relation == mapmodel.RelWar && len(tr) > 3 && (f/3)%2 == 0 {
			v.on(cv, tr[len(tr)/2], sym(mapmodel.SymWar), mapmodel.CDanger, true, lgWar)
		}
	}
	if n := len(s.scout); n > 1 {
		v.on(cv, s.scout[(f/2)%n], sym(mapmodel.SymScout), mapmodel.CLife, false, lgScout)
	}
	if n := len(s.raid); n > 1 {
		v.on(cv, s.raid[f%n], sym(mapmodel.SymRaider), mapmodel.CDanger, false, lgRaider)
	}
	if s.hasHb && (f/2)%4 != 3 { // the harbinger blinks
		v.on(cv, s.harb, sym(mapmodel.SymHarbinger), mapmodel.CCivic, true, lgHarbinger)
	}
	if hz := []rune(s.d.hazard); len(hz) > 0 { // the ring pulses
		for i, p := range s.hazard {
			v.on(cv, p, hz[(f/2+int(mapmodel.Hash(int64(i), 5)%4))%len(hz)], mapmodel.CDanger, false, lgHazard)
		}
	}
}

func factionIndex(m *mapmodel.Model, key string) int {
	for i := range m.Factions {
		if key != "" && m.Factions[i].Key == key {
			return i
		}
	}
	return -1
}

// free claims n cells from (x, y) for a label if they and a one-cell margin
// are unoccupied.
func (v *view) free(x, y, n int) bool {
	g := &v.g
	if y < g.y || y >= g.y+g.h || x < g.x || x+n > g.x+g.w {
		return false
	}
	row := v.occ[(y-g.y)*g.w : (y-g.y+1)*g.w]
	for i := max(x-1, g.x); i <= min(x+n, g.x+g.w-1); i++ {
		if row[i-g.x] {
			return false
		}
	}
	for i := x; i < x+n; i++ {
		row[i-g.x] = true
	}
	return true
}

// label writes text next to a tile, in the first free spot around it, and
// reports whether it found one.
func (v *view) label(cv *mapstyle.Canvas, p mapmodel.Pt, text string, st tcell.Style) bool {
	cx, cy, ok := v.g.cellOf(p)
	n := mapstyle.TextLen(text)
	for _, o := range [6][2]int{{2, 0}, {-n - 1, 0}, {-n / 2, -1}, {-n / 2, 1}, {2, 1}, {2, -1}} {
		if ok && v.free(cx+o[0], cy+o[1], n) {
			cv.Text(cx+o[0], cy+o[1], n, text, st)
			return true
		}
	}
	return false
}

// districtLabelBudget is how many building names the district zoom prints
// in a w x h map area: about one per 160 cells, at least 4 and at most 24,
// so a dense town reads as a map with a few names, not a wall of text.
func districtLabelBudget(w, h int) int {
	return min(24, max(4, w*h/160))
}

// drawLabels names things on the same grid as the art: the settlement and
// the civs (with their relation) at region zoom, wonders and civs at
// settlement zoom, every building type at district zoom.
func (v *view) drawLabels(cv *mapstyle.Canvas) {
	s, m := v.sc, v.sc.m
	if v.g.zoom == zRegion {
		v.label(cv, pt(s.w.CX, s.w.CY), s.w.Name, v.cls(mapmodel.CWealth).Bold(true))
	}
	if v.g.cellW == 2 {
		// The names nearest the cursor first, within the label budget.
		var anchors []*mapmodel.TownTile
		for i := range m.Town.Tiles {
			if tt := &m.Town.Tiles[i]; tt.Anchor && m.Building(tt.Key) != nil {
				if _, _, ok := v.g.cellOf(tt.Pt); ok {
					anchors = append(anchors, tt)
				}
			}
		}
		dist := func(p mapmodel.Pt) int { return abs(p.X-v.cur.X) + abs(p.Y-v.cur.Y) }
		sort.SliceStable(anchors, func(i, j int) bool {
			if di, dj := dist(anchors[i].Pt), dist(anchors[j].Pt); di != dj {
				return di < dj
			}
			return anchors[i].Key < anchors[j].Key
		})
		budget := districtLabelBudget(v.g.w, v.g.h)
		v.names = 0
		for _, tt := range anchors {
			if v.names == budget {
				break
			}
			if v.label(cv, tt.Pt, m.Building(tt.Key).Name, v.cls(mapmodel.CLabel)) {
				v.names++
			}
		}
	}
	for _, wd := range m.Town.Wonders {
		if d := m.Catalog.Defs[wd.Key]; d != nil && v.g.zoom != zRegion {
			v.label(cv, pt(wd.At.X+1, wd.At.Y), d.Name, v.cls(mapmodel.CWealth))
		}
	}
	for i := range m.Factions {
		f := &m.Factions[i]
		if p, ok := s.site(f); ok && f.Discovered && s.seen(p) == 2 {
			name := f.Name
			if v.g.cellW == 1 {
				name += " (" + f.Relation.String() + ")"
			}
			v.label(cv, p, name, v.cls(mapmodel.RelationClass(f.Relation)))
		}
	}
}

// drawMap draws the whole map area for the current zoom, then the cursor.
func (v *view) drawMap(cv *mapstyle.Canvas, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	if v.zoom == zRegion {
		v.drawRegion(cv, x, y, w, h)
	} else {
		v.g = v.fit(x, y, w, h, v.zoom, v.sc, false)
		v.drawTiles(cv, v.g)
		v.drawLife(cv)
		v.drawLabels(cv)
	}
	v.recentre = false
	if cx, cy, ok := v.g.cellOf(v.cur); ok && v.inspect {
		for i := 0; i < v.g.cellW && cx+i < v.g.x+v.g.w; i++ {
			r, _ := cv.Get(cx+i, cy)
			cv.Put(cx+i, cy, r, v.pal.Cursor.Bold(true))
		}
	}
}
