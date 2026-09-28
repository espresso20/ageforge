package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// Zoom levels share one grammar: the same glyphs, sampled differently.
const (
	ZRegion     = 0 // the whole world, down-sampled to fit
	ZSettlement = 1 // one tile per cell
	ZDistrict   = 2 // one tile per two cells: square tiles, counts and names
)

var zoomNames = []string{"REGION", "SETTLEMENT", "DISTRICT"}

// View is the interactive map: a scene, a palette, a cursor and a clock.
type View struct {
	S      *Scene
	Pal    Palette
	Zoom   int
	CurX   int
	CurY   int
	Frame  int
	Legend bool // sidebar legend when there is room

	legend map[string]legendItem
	// last map geometry, for tile<->cell conversion
	mx, my, mw, mh int
	vx, vy, scale  int
	cellW          int
	camX, camY     int
	camKey         [3]int
	occ            []bool // cells holding something a label must not cover
}

type legendItem struct {
	r     rune
	c     Class
	bg    tcell.Color
	label string
	group int
}

func NewView(s *Scene) *View {
	v := &View{S: s, Pal: NewPalette(s.D), Zoom: ZSettlement, CurX: s.W.CX, CurY: s.W.CY, Legend: true}
	return v
}

func (v *View) style(c Class) tcell.Style {
	return tcell.StyleDefault.Foreground(v.Pal.Fg[c]).Background(v.Pal.Bg)
}

func (v *View) reg(r rune, c Class, label string, group int) {
	if v.legend == nil {
		return
	}
	if _, ok := v.legend[label]; !ok {
		v.legend[label] = legendItem{r: r, c: c, label: label, group: group}
	}
}

func put(scr tcell.Screen, x, y int, r rune, st tcell.Style) { scr.SetContent(x, y, r, nil, st) }

func puts(scr tcell.Screen, x, y, max int, s string, st tcell.Style) int {
	n := 0
	for _, r := range s {
		if n >= max {
			break
		}
		put(scr, x+n, y, r, st)
		n += 1
	}
	return n
}

// ---------------------------------------------------------------- tiles

var boxLight = []rune("·│─└││┌├─┘─┴┐┤┬┼")
var boxHeavy = []rune("·┃━┗┃┃┏┣━┛━┻┓┫┳╋")
var boxDouble = []rune("·║═╚║║╔╠═╝═╩╗╣╦╬")

func boxRune(set []rune, m int) rune {
	// m: N=1 E=2 S=4 W=8
	return set[m&15]
}

func (v *View) mask(x, y int, like func(Kind) bool) int {
	m := 0
	for bit, d := range [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
		if like(v.S.cell(x+d[0], y+d[1]).K) {
			m |= 1 << bit
		}
	}
	return m
}

// glyph resolves one tile to a glyph, style and salience (for the region
// view's down-sampling: the most important thing in a block wins).
func (v *View) glyph(x, y int) (rune, tcell.Style, int) {
	s := v.S
	w := s.W
	pal := &v.Pal
	if !w.in(x, y) {
		return ' ', tcell.StyleDefault.Background(pal.Bg), -1
	}
	i := y*w.W + x
	vis := s.Vis[i]
	if vis == 0 {
		return ' ', tcell.StyleDefault.Background(pal.Bg), -1
	}
	t := w.T[i]
	c := s.Cells[i]
	bg := pal.Bg
	if t.water() {
		bg = pal.WaterBg
	}
	st := func(cl Class) tcell.Style { return tcell.StyleDefault.Foreground(pal.Fg[cl]).Background(bg) }
	if vis == 1 {
		r, _ := v.terrain(x, y, t, false)
		if r == ' ' {
			return ' ', tcell.StyleDefault.Background(pal.Bg), 0
		}
		return r, tcell.StyleDefault.Foreground(pal.Fg[CMemory]).Background(pal.Bg), 1
	}
	d := s.D
	switch c.K {
	case KStreet, KRoad:
		if d.Road == 0 && c.K == KStreet {
			// no streets yet: just trampled ground between the huts
			r, _ := v.terrain(x, y, t, true)
			return r, st(CGround), 2
		}
		if c.K == KRoad {
			// Out in the country every era's road is a line of dots: the
			// map's one symbol for "a way between places".
			v.reg('·', CRoad, "road to a civ", 2)
			return '·', st(CRoad), 40
		}
		m := v.mask(x, y, func(k Kind) bool { return roadLike(k) || k == KWall && d.Wall == 3 })
		sty := st(CRoad)
		var r rune
		switch {
		case !c.Major || d.Road == 1:
			r = boxRune(boxLight, m)
			v.reg('┼', CRoad, "street", 2)
			if c.Major {
				sty = st(CWall)
				v.reg('┼', CWall, "avenue", 2)
			}
		case d.Road == 2:
			r = boxRune(boxHeavy, m)
			sty = st(CWall)
			v.reg('╋', CWall, "avenue", 2)
		default:
			r = boxRune(boxDouble, m)
			sty = st(CWall)
			v.reg('╬', CWall, "boulevard", 2)
		}
		return r, sty, 40
	case KBridge:
		m := v.mask(x, y, roadLike)
		r := '═'
		if m&5 != 0 && m&10 == 0 {
			r = '║'
		}
		if d.Road == 0 {
			r = '='
		}
		v.reg(r, CWall, "bridge / ford", 2)
		return r, st(CWall), 41
	case KWall:
		if d.Wall == 1 {
			v.reg('#', CRoad, "palisade", 2)
			return '#', st(CRock), 50
		}
		m := v.mask(x, y, func(k Kind) bool { return k == KWall || k == KTower })
		r := boxRune(boxDouble, m)
		if m == 0 {
			r = '■'
		}
		v.reg('═', CWall, "city wall", 2)
		return r, st(CWall), 50
	case KTower:
		v.reg('■', CWall, "wall tower", 2)
		return '■', st(CWall).Bold(true), 51
	case KPlaza:
		if d.Epoch == "stone_era" {
			r, _ := v.terrain(x, y, t, true)
			return r, st(CGround), 30
		}
		return '·', st(CRoad), 30
	case KCentre:
		cl := CWealth
		if d.CentreHot {
			cl = CIdle
			if (v.Frame+1)%3 == 0 {
				cl = CDanger
			}
			v.reg(d.Centre, CIdle, "hearth", 1)
		} else {
			v.reg(d.Centre, CWealth, "town centre", 1)
		}
		return d.Centre, st(cl).Bold(true), 95
	case KBuilding:
		b := &s.Blds[c.B-1]
		r := c.G
		if r == '"' && hash(int64(x), int64(y), int64(v.Frame/2))%9 == 0 {
			r = '\'' // wind in the crops
		}
		fg := pal.Fg[c.C]
		bgc := pal.Bg
		sal := 60
		if c.Under {
			fg = pal.Dim
			v.reg(c.G, CMemory, "understaffed (dim)", 4)
		}
		if c.Fresh {
			bgc = pal.FreshBg
			sal = 65
			v.legendFresh()
		}
		v.reg(c.G, c.C, lineageNames[b.V.Lineage], 1)
		sty := tcell.StyleDefault.Foreground(fg).Background(bgc)
		if b.V.Lineage == "housing" || b.V.Lineage == "food" {
			sal = 55
		}
		return r, sty, sal
	case KWonder:
		b := &s.Blds[c.B-1]
		_ = b
		bgc := pal.Bg
		if c.Fresh {
			bgc = pal.FreshBg
		}
		v.reg('▲', CWealth, "wonder", 1)
		return c.G, tcell.StyleDefault.Foreground(pal.Fg[CWealth]).Background(bgc).Bold(true), 90
	case KRuin:
		v.reg('%', CRock, "ruins", 4)
		return '%', st(CRock), 45
	case KSite:
		f := s.V.Factions[c.F-1]
		if !f.Discovered {
			break
		}
		v.reg('▪', c.C, "civ settlement", 3)
		sty := st(c.C)
		if c.G != '▪' {
			sty = sty.Bold(true).Reverse(true)
		}
		return c.G, sty, 100
	}
	r, cl := v.terrain(x, y, t, false)
	sal := map[Terrain]int{TDeep: 12, TShallow: 14, TRiver: 35, TBeach: 8, TGrass: 0, TForest: 20, THills: 25, TMountain: 30}[t]
	if r != ' ' && t == TGrass {
		sal = 1
	}
	return r, st(cl), sal
}

func (v *View) legendFresh() {
	if v.legend == nil {
		return
	}
	if _, ok := v.legend["new since last visit"]; !ok {
		v.legend["new since last visit"] = legendItem{r: ' ', c: CLife, bg: v.Pal.FreshBg, label: "new since last visit", group: 4}
	}
}

func (v *View) terrain(x, y int, t Terrain, trampled bool) (rune, Class) {
	w := v.S.W
	h := hash(w.Seed, 11, int64(x), int64(y))
	f := int64(v.Frame / 2)
	switch t {
	case TDeep:
		v.reg('≈', CWater, "water", 0)
		if hash(int64(x), int64(y), f)%13 == 0 {
			return '~', CWater
		}
		return '≈', CWater
	case TShallow:
		v.reg('~', CWater, "shallows", 0)
		if hash(int64(x), int64(y), f)%11 == 0 {
			return '≈', CWater
		}
		return '~', CWater
	case TRiver:
		v.reg('≈', CWater, "water", 0)
		if hash(int64(x), int64(y), f)%7 == 0 {
			return '~', CWater
		}
		return '≈', CWater
	case TBeach:
		return '·', CRock
	case TForest:
		v.reg('♣', CFlora, "forest", 0)
		if h%5 == 0 {
			return 'τ', CFlora
		}
		return '♣', CFlora
	case THills:
		v.reg('^', CHill, "hills", 0)
		return '^', CHill
	case TMountain:
		v.reg('▲', CRock, "mountains", 0)
		return '▲', CRock
	}
	if trampled {
		if h%3 == 0 {
			return '.', CGround
		}
		return ' ', CGround
	}
	switch h % 19 {
	case 0, 1:
		return '.', CGround
	case 2:
		return ',', CGround
	case 3:
		return '\'', CGround
	}
	return ' ', CGround
}

// ---------------------------------------------------------------- map

// geometry fixes the viewport for the current zoom and cursor.
func (v *View) geometry(mx, my, mw, mh int) {
	w := v.S.W
	v.mx, v.my, v.mw, v.mh = mx, my, mw, mh
	v.scale, v.cellW = 1, 1
	switch v.Zoom {
	case ZRegion:
		s := int(math.Ceil(math.Max(float64(w.W)/float64(mw), float64(w.H)/float64(mh))))
		if s < 2 {
			s = 2
		}
		v.scale = s
		// centre the world in the area
		v.vx = -(mw - (w.W+s-1)/s) / 2 * s
		v.vy = -(mh - (w.H+s-1)/s) / 2 * s
		return
	case ZDistrict:
		v.cellW = 2
	}
	tw, th := mw/v.cellW, mh
	// The camera stays put and only scrolls when the cursor nears an
	// edge, like a roguelike; it starts centred on the town.
	key := [3]int{v.Zoom, tw, th}
	if key != v.camKey {
		v.camKey = key
		v.camX, v.camY = w.CX-tw/2, w.CY-th/2
	}
	mx2, my2 := min(8, tw/4), min(4, th/4)
	if v.CurX < v.camX+mx2 {
		v.camX = v.CurX - mx2
	}
	if v.CurX >= v.camX+tw-mx2 {
		v.camX = v.CurX - tw + mx2 + 1
	}
	if v.CurY < v.camY+my2 {
		v.camY = v.CurY - my2
	}
	if v.CurY >= v.camY+th-my2 {
		v.camY = v.CurY - th + my2 + 1
	}
	v.camX = clamp(v.camX, 0, max(0, w.W-tw))
	v.camY = clamp(v.camY, 0, max(0, w.H-th))
	v.vx, v.vy = v.camX, v.camY
}

func clamp(a, lo, hi int) int {
	if a < lo {
		return lo
	}
	if a > hi {
		return hi
	}
	return a
}

// tileCell maps a tile to its screen cell, if on screen.
func (v *View) tileCell(p Pt) (int, int, bool) {
	var cx, cy int
	if v.Zoom == ZRegion {
		dx, dy := p.X-v.vx, p.Y-v.vy
		if dx < 0 || dy < 0 {
			return 0, 0, false
		}
		cx, cy = dx/v.scale, dy/v.scale
	} else {
		cx, cy = (p.X-v.vx)*v.cellW, p.Y-v.vy
	}
	if cx < 0 || cy < 0 || cx >= v.mw || cy >= v.mh {
		return 0, 0, false
	}
	return v.mx + cx, v.my + cy, true
}

func (v *View) drawMap(scr tcell.Screen, mx, my, mw, mh int) {
	v.geometry(mx, my, mw, mh)
	s := v.S
	pal := &v.Pal
	blank := tcell.StyleDefault.Background(pal.Bg)
	v.occ = make([]bool, mw*mh)
	for cy := 0; cy < mh; cy++ {
		for cx := 0; cx < mw; cx++ {
			var r rune
			var st tcell.Style
			sal := 0
			switch v.Zoom {
			case ZRegion:
				best := -2
				r, st = ' ', blank
				x0, y0 := v.vx+cx*v.scale, v.vy+cy*v.scale
				for dy := 0; dy < v.scale; dy++ {
					for dx := 0; dx < v.scale; dx++ {
						g, gs, sl := v.glyph(x0+dx, y0+dy)
						if sl > best {
							best, r, st = sl, g, gs
						}
					}
				}
				sal = best
				if s.D.Space {
					if o := v.orbit(x0+v.scale/2, y0+v.scale/2); o > 1.0 {
						r, st = v.star(cx, cy)
						sal = 0
						if o > 1.035 && o < 1.075 {
							// the orbital ring, with stations riding it
							r, st = '·', v.style(CCivic)
							a := math.Atan2(float64(y0-s.W.H/2)*2, float64(x0-s.W.W/2))
							for k := 0; k < 3; k++ {
								sa := math.Mod(float64(v.Frame)*0.06+float64(k)*2.1, 2*math.Pi) - math.Pi
								if math.Abs(a-sa) < 0.022 {
									r, st = '◘', v.style(CWealth).Bold(true)
									v.reg('◘', CWealth, "orbital station", 5)
								}
							}
							v.reg('·', CCivic, "orbital ring", 2)
						}
					}
				}
			case ZDistrict:
				tx, ty := v.vx+cx/2, v.vy+cy
				g, gs, sl := v.glyph(tx, ty)
				r, st, sal = g, gs, sl
				if cx%2 == 1 {
					r = v.rightHalf(tx, ty, g)
				}
			default:
				r, st, sal = v.glyph(v.vx+cx, v.vy+cy)
			}
			// District labels may sit on a street; elsewhere streets are
			// part of the picture and stay clear.
			v.occ[cy*mw+cx] = sal >= 40 && !(v.Zoom == ZDistrict && sal < 45)
			put(scr, mx+cx, my+cy, r, st)
		}
	}
	v.drawLife(scr)
	v.drawLabels(scr)
	// cursor
	if cx, cy, ok := v.tileCell(Pt{v.CurX, v.CurY}); ok {
		r, _, _, _ := scr.GetContent(cx, cy)
		put(scr, cx, cy, r, tcell.StyleDefault.Foreground(pal.CursorFg).Background(pal.CursorBg).Bold(true))
		if v.cellW == 2 && cx+1 < mx+mw {
			r2, _, _, _ := scr.GetContent(cx+1, cy)
			put(scr, cx+1, cy, r2, tcell.StyleDefault.Foreground(pal.CursorFg).Background(pal.CursorBg))
		}
	}
}

// rightHalf fills the second cell of a district tile: roads and walls
// continue, water and fields repeat, a building's anchor shows its count.
func (v *View) rightHalf(x, y int, g rune) rune {
	s := v.S
	if !s.W.in(x, y) || s.Vis[y*s.W.W+x] == 0 {
		return ' '
	}
	c := s.cell(x, y)
	east := s.cell(x+1, y).K
	switch c.K {
	case KRoad:
		return ' '
	case KStreet:
		if s.D.Road == 0 || !(roadLike(east) || east == KWall && s.D.Wall == 3) {
			return ' '
		}
		switch {
		case !c.Major || s.D.Road == 1:
			return '─'
		case s.D.Road == 2:
			return '━'
		}
		return '═'
	case KWall:
		if east == KWall || east == KTower {
			if s.D.Wall == 1 {
				return '#'
			}
			return '═'
		}
		return ' '
	case KBridge:
		if roadLike(east) {
			return g
		}
		return ' '
	case KBuilding:
		b := s.Blds[c.B-1]
		if c.Anchor {
			return superscript(b.V.Count)
		}
		if b.V.Lineage == "food" {
			return g
		}
		return ' '
	case KWonder:
		if east != KWonder {
			return ' '
		}
		switch g {
		case '█', '▓', '▀':
			return g
		case '╓', '╥', '┌', '╬', '╨':
			return '─'
		}
		return ' '
	case KNone, KPlaza:
		t := s.W.at(x, y)
		if t.water() || t == TForest && hash(int64(x), int64(y))%2 == 0 {
			return g
		}
	}
	return ' '
}

func superscript(n int) rune {
	sup := []rune("⁰¹²³⁴⁵⁶⁷⁸⁹")
	if n < 10 {
		return sup[n]
	}
	return '⁺'
}

// orbit: in the cosmic era the region view pulls back far enough to see
// the planet's limb, with stars and an orbital ring around it. It returns
// the normalised distance from the planet's centre (>1 is space).
func (v *View) orbit(x, y int) float64 {
	w := v.S.W
	dx := (float64(x) - float64(w.W)/2) / (float64(w.W) / 2 * 0.98)
	dy := (float64(y) - float64(w.H)/2) / (float64(w.H) / 2 * 1.0)
	return math.Sqrt(dx*dx + dy*dy)
}

func (v *View) star(cx, cy int) (rune, tcell.Style) {
	h := hash(int64(cx), int64(cy), 7)
	st := tcell.StyleDefault.Foreground(v.Pal.Fg[CStar]).Background(v.Pal.Bg)
	switch h % 23 {
	case 0:
		if (v.Frame+int(h>>8))%5 == 0 {
			return '*', st.Bold(true)
		}
		return '·', st
	case 1, 2:
		return '·', st
	case 3:
		return '.', st
	}
	return ' ', st
}

// drawLife puts the moving things on top: workers, idle hands, caravans,
// scouts, raiders, smoke, the harbinger and the catastrophe.
func (v *View) drawLife(scr tcell.Screen) {
	s := v.S
	pal := &v.Pal
	on := func(p Pt, r rune, c Class, over bool, label string, group int) {
		if !s.W.in(p.X, p.Y) || s.Vis[s.idx(p.X, p.Y)] < 2 {
			return
		}
		k := s.cell(p.X, p.Y).K
		if !over && (k == KBuilding || k == KWonder || k == KSite || k == KCentre) {
			return
		}
		if cx, cy, ok := v.tileCell(p); ok {
			_, _, cur, _ := scr.GetContent(cx, cy)
			_, bg, _ := cur.Decompose()
			put(scr, cx, cy, r, tcell.StyleDefault.Foreground(pal.Fg[c]).Background(bg).Bold(c == CLife || c == CIdle))
			if label != "" {
				v.reg(r, c, label, group)
			}
		}
	}
	f := v.Frame
	// smoke over the works (only close up)
	if v.Zoom != ZRegion && len(s.D.Smoke) > 0 {
		for bi, b := range s.Blds {
			if !smoky(b.V.Lineage) || b.V.Workers == 0 {
				continue
			}
			for ti, t := range b.Tiles {
				if ti > 3 {
					break
				}
				ph := (f + int(hash(int64(bi), int64(ti))%7)) % 7
				if ph >= 3 {
					continue
				}
				p := Pt{t.X + ph/2, t.Y - 1 - ph}
				if s.cell(p.X, p.Y).K == KNone || s.cell(p.X, p.Y).K == KStreet {
					on(p, s.D.Smoke[ph%len(s.D.Smoke)], CMemory, false, "", 5)
				}
			}
		}
	}
	for i, path := range s.Walks {
		n := len(path)
		t := (f + int(hash(int64(i), 3)%uint64(2*n-2))) % (2*n - 2)
		if t >= n {
			t = 2*n - 2 - t
		}
		on(path[t], '☺', CLife, false, "worker at work", 5)
	}
	for j, p := range s.Idle {
		r := '☺'
		if (f/3+j)%4 == 0 {
			r = '☻'
		}
		on(p, r, CIdle, true, "idle worker (assign!)", 5)
	}
	// caravans on the roads to civs that trade with us
	routes := s.V.Routes
	k := 0
	for fi, tr := range s.Trails {
		if len(tr) < 4 {
			continue
		}
		fv := s.V.Factions[fi]
		trading := fv.TradeCount > 0 || k < len(routes)
		k++
		if fv.AtWar {
			mid := tr[len(tr)/2]
			if f%2 == 0 {
				on(mid, '×', CDanger, true, "war on this road", 5)
			}
			continue
		}
		if !trading {
			continue
		}
		for c := 0; c < 2; c++ {
			pos := (f + c*len(tr)/2 + fi*5) % len(tr)
			on(tr[pos], '&', CWealth, false, "caravan", 5)
		}
	}
	if n := len(s.Scout); n > 1 {
		on(s.Scout[f%n], '@', CLife, false, "scouts", 5)
	}
	if n := len(s.Raid); n > 1 {
		on(s.Raid[(f*2)%n], '»', CDanger, false, "raiders", 5)
	}
	// the harbinger waits at the edge of town
	if s.V.Harbinger != "" {
		a := hf(s.W.Seed, 97) * 2 * math.Pi
		r := math.Max(s.WallR, 4) + 2.5
		p := Pt{s.WallC.X + int(math.Cos(a)*r*2), s.WallC.Y + int(math.Sin(a)*r)}
		if f%4 != 3 {
			on(p, 'Ψ', CCivic, true, "harbinger", 5)
		}
	}
	// catastrophe pressing on the walls
	if len(s.Hazard) > 0 {
		g, label := catastropheGlyph(s.V.PendingCatastrophe)
		for i, p := range s.Hazard {
			ph := (f + int(hash(int64(i), 5)%4)) % 4
			r := g[ph%len(g)]
			on(p, r, CDanger, false, label, 5)
		}
	}
}

func catastropheGlyph(epoch string) ([]rune, string) {
	switch epoch {
	case "stone_era":
		return []rune("*·˙*"), "meteor fall"
	case "iron_era":
		return []rune("ЖЖ»Ж"), "barbarian host"
	case "steel_era":
		return []rune("▒░▒░"), "smog of collapse"
	case "electric_era":
		return []rune("░▒░·"), "fallout"
	case "digital_era":
		return []rune("#%?0"), "glitch"
	case "neon_era":
		return []rune("☼*☼·"), "solar flare"
	}
	return []rune("╳╱╲╳"), "reality fracture"
}

// drawLabels writes names on the same grid as the art.
func (v *View) drawLabels(scr tcell.Screen) {
	s := v.S
	pal := &v.Pal
	lst := tcell.StyleDefault.Foreground(pal.Fg[CLabel]).Background(pal.Bg)
	free := func(x, y, n int) bool {
		if y < v.my || y >= v.my+v.mh || x < v.mx || x+n > v.mx+v.mw {
			return false
		}
		for i := -1; i <= n; i++ {
			if x+i < v.mx || x+i >= v.mx+v.mw {
				continue
			}
			if v.occ[(y-v.my)*v.mw+x+i-v.mx] {
				return false
			}
		}
		for i := 0; i < n; i++ {
			v.occ[(y-v.my)*v.mw+x+i-v.mx] = true
		}
		return true
	}
	label := func(p Pt, text string, st tcell.Style) {
		cx, cy, ok := v.tileCell(p)
		if !ok {
			return
		}
		n := utf8.RuneCountInString(text)
		for _, o := range [][2]int{{2, 0}, {-n - 1, 0}, {-n / 2, -1}, {-n / 2, 1}, {2, 1}, {2, -1}} {
			if free(cx+o[0], cy+o[1], n) {
				puts(scr, cx+o[0], cy+o[1], n, text, st)
				return
			}
		}
	}
	switch v.Zoom {
	case ZRegion:
		label(Pt{s.W.CX, s.W.CY}, s.W.Name, lst.Foreground(pal.Fg[CWealth]).Bold(true))
		for fi, f := range s.V.Factions {
			if f.Discovered && fi < len(s.W.Sites) && s.Vis[s.idx(s.W.Sites[fi].X, s.W.Sites[fi].Y)] == 2 {
				label(s.W.Sites[fi], f.Name, lst)
			}
		}
	case ZDistrict:
		for _, b := range s.Blds {
			if len(b.Tiles) == 0 {
				continue
			}
			a := b.Tiles[0]
			if b.Wonder {
				a = Pt{b.Tiles[0].X + 1, b.Tiles[0].Y}
			}
			label(a, b.V.Name, lst)
		}
	default:
		for _, b := range s.Blds {
			if b.Wonder && len(b.Tiles) > 0 {
				label(Pt{b.Tiles[0].X + 1, b.Tiles[0].Y}, b.V.Name, lst.Foreground(pal.Fg[CWealth]))
			}
		}
		for fi, f := range s.V.Factions {
			if f.Discovered && fi < len(s.W.Sites) {
				label(s.W.Sites[fi], f.Name, lst)
			}
		}
	}
}

// ---------------------------------------------------------------- chrome

func (v *View) chrome(role theme.Role) tcell.Style {
	return tcell.StyleDefault.Foreground(theme.Color(role)).Background(theme.Color(theme.RoleBackground))
}

// Draw renders the full view: header, map, legend, inspector.
func (v *View) Draw(scr tcell.Screen, W, H int) {
	s := v.S
	bgst := tcell.StyleDefault.Background(theme.Color(theme.RoleBackground))
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			put(scr, x, y, ' ', bgst)
		}
	}
	if W < 60 || H < 16 {
		v.DrawCompact(scr, W, H)
		return
	}
	v.legend = map[string]legendItem{}
	legendW := 0
	if v.Legend && W >= 110 {
		legendW = 26
	}
	footH := 3
	mapW, mapH := W-legendW, H-1-footH-1
	v.drawMap(scr, 0, 1, mapW, mapH)

	// header
	acc := v.chrome(theme.RoleAccent).Bold(true)
	dim := v.chrome(theme.RoleDim)
	lab := v.chrome(theme.RoleLabel)
	x := puts(scr, 0, 0, W, " ◆ "+s.W.Name, acc)
	x += puts(scr, x, 0, W-x, " · "+s.V.AgeName+" · "+s.V.EpochName, lab)
	met := 0
	for _, f := range s.V.Factions {
		if f.Discovered {
			met++
		}
	}
	right := fmt.Sprintf("☺ %s  idle %d · civs met %d/%d · %s ", comma(s.V.Pop), s.V.Idle, met, len(s.V.Factions), zoomNames[v.Zoom])
	if rw := utf8.RuneCountInString(right); x+rw+2 < W {
		puts(scr, W-rw, 0, rw, right, dim)
		idleX := W - rw + strings.Index(right, "idle")
		_ = idleX
	}
	// rule under the map
	ry := 1 + mapH
	for i := 0; i < W; i++ {
		put(scr, i, ry, '─', v.chrome(theme.RoleBorder))
	}
	if legendW > 0 {
		v.drawLegend(scr, mapW, 1, legendW, mapH)
	}
	v.drawInspector(scr, 0, ry+1, W)
}

func comma(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func itoa(n int) string { return fmt.Sprint(n) }

func (v *View) drawLegend(scr tcell.Screen, x0, y0, w, h int) {
	bst := v.chrome(theme.RoleBorder)
	for y := y0; y < y0+h; y++ {
		put(scr, x0, y, '│', bst)
	}
	puts(scr, x0+2, y0, w-3, "LEGEND", v.chrome(theme.RoleAccent).Bold(true))
	items := make([]legendItem, 0, len(v.legend))
	for _, it := range v.legend {
		items = append(items, it)
	}
	sort.Slice(items, func(a, b int) bool {
		if items[a].group != items[b].group {
			return items[a].group < items[b].group
		}
		return items[a].label < items[b].label
	})
	heads := []string{"land", "buildings", "ways & walls", "civs", "state", "life"}
	y := y0 + 1
	last := -1
	for _, it := range items {
		if y >= y0+h {
			break
		}
		if it.group != last {
			if y+1 >= y0+h {
				break
			}
			puts(scr, x0+2, y, w-3, heads[it.group], v.chrome(theme.RoleDim))
			y++
			last = it.group
		}
		st := tcell.StyleDefault.Foreground(v.Pal.Fg[it.c]).Background(v.Pal.Bg)
		if it.bg != 0 {
			st = st.Background(it.bg)
		}
		put(scr, x0+3, y, it.r, st)
		puts(scr, x0+5, y, w-6, it.label, v.chrome(theme.RoleText))
		y++
	}
}

// Describe is the inspector: what is under the cursor, in the game's own
// vocabulary (the build key is the one `build` takes).
func (v *View) Describe(x, y int) (string, string) {
	s := v.S
	w := s.W
	if !w.in(x, y) || s.Vis[s.idx(x, y)] == 0 {
		return "░ Unexplored", "send an expedition to lift the fog"
	}
	c := s.cell(x, y)
	t := w.at(x, y)
	tname := map[Terrain]string{TDeep: "Open water", TShallow: "Shallows", TRiver: "River", TBeach: "Shore",
		TGrass: "Grassland", TForest: "Forest", THills: "Hills", TMountain: "Mountains"}[t]
	switch c.K {
	case KBuilding, KWonder, KRuin:
		b := s.Blds[c.B-1]
		g := b.G
		if c.K == KWonder {
			g = '▲'
		}
		l1 := fmt.Sprintf("%c %s ×%d", g, b.V.Name, b.V.Count)
		if b.V.WorkCap > 0 {
			l1 += fmt.Sprintf(" · %d/%d workers", b.V.Workers, b.V.WorkCap)
		}
		if b.V.Rate > 0 {
			l1 += fmt.Sprintf(" · ~+%s %s/s", short(b.V.Rate), strings.ToLower(b.V.RateName))
		}
		l1 += " · " + lineageNames[b.V.Lineage]
		var l2 []string
		if b.Delta > 0 {
			l2 = append(l2, fmt.Sprintf("+%d since your last visit", b.Delta))
		}
		if c.Under {
			l2 = append(l2, "understaffed: assign workers")
		}
		if b.V.Legacy {
			l2 = append(l2, "legacy (superseded)")
		} else if !b.Wonder {
			l2 = append(l2, "build "+b.V.Key)
		}
		if b.V.Ruins > 0 {
			l2 = append(l2, fmt.Sprintf("%d in ruins", b.V.Ruins))
		}
		return l1, strings.Join(l2, " · ")
	case KCentre, KPlaza:
		name := "Town square"
		if s.D.CentreHot {
			name = "The hearth"
		}
		l2 := fmt.Sprintf("%s people, %d idle", comma(s.V.Pop), s.V.Idle)
		if s.V.Idle > 0 {
			l2 += " · assign them"
		}
		return fmt.Sprintf("%c %s of %s", s.D.Centre, name, w.Name), l2
	case KSite:
		f := s.V.Factions[c.F-1]
		st := f.Status
		if f.AtWar {
			st = "AT WAR"
		}
		return fmt.Sprintf("%c %s · %s (opinion %+d) · strength %d", []rune(f.Name)[0], f.Name, st, f.Opinion, f.Strength),
			fmt.Sprintf("specialty %s · %d trades · diplomacy %s", f.Specialty, f.TradeCount, f.Key)
	case KStreet, KRoad, KBridge:
		for fi, tr := range s.Trails {
			for _, p := range tr {
				if p.X == x && p.Y == y {
					f := s.V.Factions[fi]
					return "· Road to " + f.Name, fmt.Sprintf("%d trades so far", f.TradeCount)
				}
			}
		}
		return "┼ Street", tname
	case KWall, KTower:
		return "═ City wall", fmt.Sprintf("ring %0.f tiles out; the city grows past it", s.WallR)
	}
	extra := ""
	switch t {
	case TForest:
		extra = "wood · woodcutters settle at its edge"
	case THills, TMountain:
		extra = "stone and ore · quarries and mines dig here"
	case TRiver, TShallow:
		extra = "water · harbors and mills settle on its banks"
	}
	if s.Vis[s.idx(x, y)] == 1 {
		extra = "seen once, not in sight"
	}
	return tname, extra
}

func short(f float64) string {
	switch {
	case f >= 1e9:
		return fmt.Sprintf("%.1fB", f/1e9)
	case f >= 1e6:
		return fmt.Sprintf("%.1fM", f/1e6)
	case f >= 1e3:
		return fmt.Sprintf("%.1fk", f/1e3)
	}
	return fmt.Sprintf("%.1f", f)
}

func (v *View) drawInspector(scr tcell.Screen, x0, y0, W int) {
	s := v.S
	l1, l2 := v.Describe(v.CurX, v.CurY)
	puts(scr, x0+1, y0, W-2, "▸ "+l1, v.chrome(theme.RoleBright).Bold(true))
	x := puts(scr, x0+3, y0+1, W-4, l2, v.chrome(theme.RoleText))
	// alerts and "since last visit" share the second line
	var alerts []string
	if s.V.PendingCatastrophe != "" {
		name, _ := catastropheLabel(s.V)
		alerts = append(alerts, "⚠ "+name)
	}
	if len(s.Changes) > 0 {
		n := len(s.Changes)
		if n > 3 {
			n = 3
		}
		alerts = append(alerts, "since last visit: "+strings.Join(s.Changes[:n], ", "))
	}
	a := strings.Join(alerts, "  ")
	if a != "" && W < 110 {
		// no room beside the inspector: the news takes the hint line
		st := v.chrome(theme.RolePositive)
		if s.V.PendingCatastrophe != "" {
			st = v.chrome(theme.RoleNegative).Bold(true)
		}
		puts(scr, x0+1, y0+2, W-2, a, st)
		return
	}
	if a != "" {
		aw := utf8.RuneCountInString(a)
		ax := W - aw - 1
		if ax < x+6 {
			ax = x + 6
		}
		st := v.chrome(theme.RolePositive)
		if s.V.PendingCatastrophe != "" {
			st = v.chrome(theme.RoleNegative).Bold(true)
		}
		puts(scr, ax, y0+1, W-ax-1, a, st)
	}
	keys := "←↑↓→ move  ⇧ fast  z/x zoom  tab next building  c centre  ? legend  q quit"
	if W < 100 {
		keys = "←↑↓→ move  z/x zoom  tab next  ? legend  q quit"
	}
	puts(scr, x0+1, y0+2, W-2, keys, v.chrome(theme.RoleDim))
}

func catastropheLabel(v MapView) (string, string) {
	if v.CatastropheName != "" {
		return v.CatastropheName, ""
	}
	_, l := catastropheGlyph(v.PendingCatastrophe)
	return strings.ToUpper(l[:1]) + l[1:] + " looms", ""
}

// DrawCompact is the glanceable mini-map for a sidebar: the settlement
// down-sampled to fit, one header line, one line of news.
func (v *View) DrawCompact(scr tcell.Screen, W, H int) {
	s := v.S
	v.legend = nil
	bgst := tcell.StyleDefault.Background(theme.Color(theme.RoleBackground))
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			put(scr, x, y, ' ', bgst)
		}
	}
	head := fmt.Sprintf("%s · %s", s.W.Name, strings.TrimSuffix(s.V.AgeName, " Age"))
	x := puts(scr, 0, 0, W, head, v.chrome(theme.RoleAccent).Bold(true))
	idle := fmt.Sprintf("☺%s", shortInt(s.V.Pop))
	if s.V.Idle > 0 {
		idle += fmt.Sprintf(" %d idle", s.V.Idle)
	}
	if iw := utf8.RuneCountInString(idle); x+iw+1 < W {
		st := v.chrome(theme.RoleDim)
		if s.V.Idle > 0 {
			st = v.chrome(theme.RoleWarning)
		}
		puts(scr, W-iw, 0, iw, idle, st)
	}
	// fit the built area
	minX, minY, maxX, maxY := s.W.CX-8, s.W.CY-3, s.W.CX+8, s.W.CY+3
	for _, b := range s.Blds {
		for _, t := range b.Tiles {
			minX, maxX = min(minX, t.X), max(maxX, t.X)
			minY, maxY = min(minY, t.Y), max(maxY, t.Y)
		}
	}
	mh := H - 2
	sc := int(math.Ceil(math.Max(float64(maxX-minX+3)/float64(W), float64(maxY-minY+3)/float64(mh))))
	saveZ, saveX, saveY := v.Zoom, v.CurX, v.CurY
	if sc <= 1 {
		v.Zoom = ZSettlement
		v.CurX, v.CurY = (minX+maxX)/2, (minY+maxY)/2
		v.drawMap(scr, 0, 1, W, mh)
	} else {
		// region-style sampling at a custom scale, centred on the city
		v.Zoom = ZRegion
		v.mx, v.my, v.mw, v.mh = 0, 1, W, mh
		v.scale, v.cellW = sc, 1
		v.vx = (minX+maxX)/2 - W*sc/2
		v.vy = (minY+maxY)/2 - mh*sc/2
		v.drawSampled(scr)
		v.drawLife(scr)
	}
	v.Zoom, v.CurX, v.CurY = saveZ, saveX, saveY
	news := ""
	st := v.chrome(theme.RolePositive)
	switch {
	case s.V.PendingCatastrophe != "":
		name, _ := catastropheLabel(s.V)
		news, st = "⚠ "+name, v.chrome(theme.RoleNegative).Bold(true)
	case len(s.Changes) > 0:
		news = "▲ " + strings.Join(s.Changes[:min(2, len(s.Changes))], ", ")
	default:
		news = "all quiet"
		st = v.chrome(theme.RoleDim)
	}
	puts(scr, 0, H-1, W, news, st)
}

// drawSampled is drawMap's region branch at the current geometry.
func (v *View) drawSampled(scr tcell.Screen) {
	blank := tcell.StyleDefault.Background(v.Pal.Bg)
	for cy := 0; cy < v.mh; cy++ {
		for cx := 0; cx < v.mw; cx++ {
			best := -2
			r, st := ' ', blank
			x0, y0 := v.vx+cx*v.scale, v.vy+cy*v.scale
			for dy := 0; dy < v.scale; dy++ {
				for dx := 0; dx < v.scale; dx++ {
					g, gs, sal := v.glyph(x0+dx, y0+dy)
					if sal > best {
						best, r, st = sal, g, gs
					}
				}
			}
			put(scr, v.mx+cx, v.my+cy, r, st)
		}
	}
}

func shortInt(n int) string {
	if n >= 10000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprint(n)
}

// NextBuilding moves the cursor to the next building cluster.
func (v *View) NextBuilding(dir int) {
	s := v.S
	var anchors []Pt
	for _, b := range s.Blds {
		if len(b.Tiles) > 0 {
			anchors = append(anchors, b.Tiles[0])
		}
	}
	if len(anchors) == 0 {
		return
	}
	sort.Slice(anchors, func(a, b int) bool {
		da := rdist(anchors[a].X, anchors[a].Y, s.W.CX, s.W.CY)
		db := rdist(anchors[b].X, anchors[b].Y, s.W.CX, s.W.CY)
		return da < db
	})
	cur := -1
	for i, a := range anchors {
		if a.X == v.CurX && a.Y == v.CurY {
			cur = i
		}
	}
	n := (cur + dir + len(anchors)) % len(anchors)
	v.CurX, v.CurY = anchors[n].X, anchors[n].Y
}
