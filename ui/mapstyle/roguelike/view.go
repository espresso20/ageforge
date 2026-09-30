// Package roguelike is the default map style: a glyph world in the lineage
// of Brogue and Dwarf Fortress. Every cell is one coloured character that
// means something; the settlement grows out of the real buildings, quarter
// by quarter, with streets, walls, wonders and people walking, and a cursor
// inspects any tile in the game's own command words. One grammar covers the
// town and the world: the region zoom is the same glyphs down-sampled, drawn
// on the epoch's cartographic plate.
package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// Zoom levels share one grammar: the same glyphs, sampled differently.
const (
	zRegion     = 0 // the whole world down-sampled, on the epoch's plate
	zSettlement = 1 // one tile per cell
	zDistrict   = 2 // one tile per two cells, with counts and names
)

var zoomNames = [3]string{"REGION", "SETTLEMENT", "DISTRICT"}

// Entry is the registry entry for the roguelike style.
func Entry() mapstyle.Entry {
	return mapstyle.Entry{Name: "roguelike", Title: "Roguelike",
		Blurb: "A glyph world with an inspect cursor: your town, its people and the lands around it.",
		New:   func() mapstyle.Style { return newView() }}
}

// view is one open map. What it draws is a function of the frame, the
// view's own state and the active theme.
type view struct {
	zoom                            int
	cur                             mapmodel.Pt
	seed                            int64
	placed, recentre                bool
	flows, inspect, legend, changes bool
	cam                             [3]mapmodel.Pt
	camKey                          [3][2]int

	sc  *scene
	scM *mapmodel.Model
	pal *pal

	// per-frame scratch
	g    geom
	occ  []bool
	seen [numLg]lgEntry
	anim int
	tier mapmodel.GlyphTier
	rbuf regionBuf
	// names is how many building names the last district-zoom frame drew.
	names int
}

func newView() *view { return &view{zoom: zSettlement, inspect: true, legend: true, changes: true} }

func (v *view) Name() string { return "roguelike" }

func (v *view) SetOption(o mapstyle.Option, on bool) {
	if o == mapstyle.OptWorld {
		switch {
		case on:
			v.zoom = zRegion
		case v.zoom == zRegion:
			v.zoom = zSettlement
		}
		return
	}
	if p := v.option(o); p != nil {
		*p = on
	}
}

func (v *view) option(o mapstyle.Option) *bool {
	switch o {
	case mapstyle.OptFlows:
		return &v.flows
	case mapstyle.OptInspect:
		return &v.inspect
	case mapstyle.OptLegend:
		return &v.legend
	case mapstyle.OptChanges:
		return &v.changes
	}
	return nil
}

// sceneFor returns the scene for a model, building it when the model
// pointer changes; a new seed puts the cursor back on the square.
func (v *view) sceneFor(m *mapmodel.Model) *scene {
	if m == nil {
		return nil
	}
	if m != v.scM && v.sc != nil && v.scM != nil && m.LayoutKey == v.scM.LayoutKey {
		v.scM, v.sc.m = m, m // same layout, only the clock moved
	} else if m != v.scM {
		v.scM, v.sc = m, newScene(m)
	}
	if v.sc != nil && (!v.placed || v.seed != m.Seed) {
		v.seed, v.placed, v.recentre = m.Seed, true, true
		v.cur = mapmodel.Pt{X: v.sc.w.CX, Y: v.sc.w.CY}
	}
	return v.sc
}

// pal is the palette for one theme and epoch plus the few derived colours
// this style needs; rebuilt only when the theme or the epoch changes.
type pal struct {
	*mapstyle.Palette
	key    string
	dim    tcell.Color    // understaffed buildings
	tint   tcell.Color    // the plate's paper (hide, parchment)
	biome  [8]tcell.Color // satellite colour cells, by terrain
	scan   tcell.Color
	border tcell.Style
	memo   map[[2]tcell.Color]tcell.Color
}

func (v *view) palette(epoch int) *pal {
	key := theme.Active().Key
	if v.pal != nil && v.pal.key == key && v.pal.Epoch == epoch {
		return v.pal
	}
	base := *mapstyle.NewPalette(epoch)
	p := &pal{Palette: &base, key: key, memo: map[[2]tcell.Color]tcell.Color{}}
	bg := p.Bg
	if dials[epochIdx(epoch)].night { // the land recedes so lit structures glow
		for _, c := range []mapmodel.Class{mapmodel.CGround, mapmodel.CFlora, mapmodel.CRock, mapmodel.CHill, mapmodel.CWater} {
			p.Fg[c] = theme.Legible(theme.Mix(p.Fg[c], bg, 0.35), bg, 1.5)
		}
	}
	p.dim = theme.Legible(theme.Mix(bg, theme.Color(theme.RoleDim), 0.6), bg, 1.6)
	p.tint = bg
	switch epochIdx(epoch) {
	case 0:
		p.tint = theme.Mix(bg, p.Fg[mapmodel.CRock], 0.08)
	case 1:
		p.tint = theme.Mix(bg, p.Fg[mapmodel.CWealth], 0.06)
	}
	for t, c := range [8]mapmodel.Class{mapmodel.CWater, mapmodel.CWater, mapmodel.CWater, mapmodel.CRock,
		mapmodel.CGround, mapmodel.CFlora, mapmodel.CHill, mapmodel.CRock} {
		p.biome[t] = theme.Mix(bg, p.Fg[c], [8]float64{0.30, 0.22, 0.26, 0.18, 0.22, 0.30, 0.24, 0.20}[t])
	}
	p.scan = theme.Mix(bg, p.Fg[mapmodel.CLife], 0.14)
	p.border = tcell.StyleDefault.Foreground(theme.Legible(theme.Color(theme.RoleBorder), bg, 1.8)).Background(bg)
	v.pal = p
	return p
}

// on is fg made legible on bg (memoised: a frame has few such pairs).
func (p *pal) on(fg, bg tcell.Color) tcell.Color {
	if bg == p.Bg {
		return fg
	}
	k := [2]tcell.Color{fg, bg}
	c, ok := p.memo[k]
	if !ok {
		c = theme.Legible(fg, bg, 3)
		p.memo[k] = c
	}
	return c
}

// glyph is one resolved cell before styling.
type glyph struct {
	r      rune
	c      mapmodel.Class
	fg, bg tcell.Color // 0: the class colour, the palette background
	attr   tcell.AttrMask
	sal    int // salience: in a down-sampled block the most important wins
}

func (v *view) style(g glyph) tcell.Style {
	fg, bg := g.fg, g.bg
	if fg == 0 {
		fg = v.pal.Fg[g.c]
	}
	if bg == 0 {
		bg = v.pal.Bg
	} else {
		fg = v.pal.on(fg, bg)
	}
	return tcell.StyleDefault.Foreground(fg).Background(bg).Attributes(g.attr)
}

func (v *view) cls(c mapmodel.Class) tcell.Style { return v.pal.Style(c) }

func (v *view) role(r theme.Role) tcell.Style { return v.pal.Role(r) }

// geom is the map area's geometry for one frame.
type geom struct {
	x, y, w, h int // map area on the canvas
	zoom       int
	scale      int // region: tiles per cell
	cellW      int // district: cells per tile
	vx, vy     int // world tile under the area's top-left cell
}

// fit fixes the viewport for a zoom. The camera stays put and scrolls only
// when the cursor nears an edge; it jumps only when the cursor is off screen.
func (v *view) fit(x, y, w, h, zoom int, s *scene, spaceFit bool) geom {
	g := geom{x: x, y: y, w: w, h: h, zoom: zoom, scale: 1, cellW: 1}
	W, H := s.w.W, s.w.H
	if zoom == zRegion {
		fw, fh := float64(w), float64(h)
		if spaceFit {
			fw, fh = float64(fw*0.72), float64(fh*0.72)
		}
		g.scale = max(2, int(ceilDiv(float64(W), fw)), int(ceilDiv(float64(H), fh)))
		g.vx, g.vy = W/2-w*g.scale/2, H/2-h*g.scale/2
		return g
	}
	if zoom == zDistrict {
		g.cellW = 2
	}
	tw, th := max(1, w/g.cellW), h
	c := &v.cam[zoom]
	if key := [2]int{tw, th}; key != v.camKey[zoom] || v.recentre ||
		v.cur.X < c.X || v.cur.X >= c.X+tw || v.cur.Y < c.Y || v.cur.Y >= c.Y+th {
		v.camKey[zoom] = key
		c.X, c.Y = v.cur.X-tw/2, v.cur.Y-th/2
	}
	mx, my := min(8, tw/4), min(4, th/4)
	c.X = clamp(c.X, v.cur.X-tw+mx+1, v.cur.X-mx)
	c.Y = clamp(c.Y, v.cur.Y-th+my+1, v.cur.Y-my)
	c.X = clamp(c.X, 0, max(0, W-tw))
	c.Y = clamp(c.Y, 0, max(0, H-th))
	g.vx, g.vy = c.X, c.Y
	return g
}

func ceilDiv(a, b float64) float64 {
	if b <= 0 {
		return a
	}
	q := a / b
	if f := float64(int(q)); f < q {
		return f + 1
	}
	return q
}

// cellOf maps a tile to its cell on the canvas, if it is in the map area.
func (g *geom) cellOf(p mapmodel.Pt) (int, int, bool) {
	cx, cy := (p.X-g.vx)*g.cellW, p.Y-g.vy
	if g.zoom == zRegion {
		if p.X < g.vx || p.Y < g.vy {
			return 0, 0, false
		}
		cx, cy = (p.X-g.vx)/g.scale, (p.Y-g.vy)/g.scale
	}
	if cx < 0 || cy < 0 || cx >= g.w || cy >= g.h {
		return 0, 0, false
	}
	return g.x + cx, g.y + cy, true
}

var keyMoves = map[tcell.Key][2]int{tcell.KeyLeft: {-1, 0}, tcell.KeyRight: {1, 0}, tcell.KeyUp: {0, -1}, tcell.KeyDown: {0, 1}}

// HandleKey takes only keys that print nothing, so typing reaches the
// prompt: the arrows move the cursor (Shift by 8), PgUp and PgDn zoom out
// and in, Tab and Shift-Tab jump between buildings and wonders (and civs at
// region zoom), and Home centers on the town square. The flows overlay is
// the map flows command (SetOption).
func (v *view) HandleKey(ev *tcell.EventKey, f mapstyle.Frame) bool {
	s := v.sceneFor(f.Model)
	if ev == nil || s == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyTab, tcell.KeyBacktab:
		v.jump(s, map[bool]int{true: 1, false: -1}[ev.Key() == tcell.KeyTab])
		return true
	case tcell.KeyPgUp:
		v.zoom = max(zRegion, v.zoom-1)
		return true
	case tcell.KeyPgDn:
		v.zoom = min(zDistrict, v.zoom+1)
		return true
	case tcell.KeyHome:
		v.cur, v.recentre = mapmodel.Pt{X: s.w.CX, Y: s.w.CY}, true
		return true
	}
	d, ok := keyMoves[ev.Key()]
	if !ok {
		return false
	}
	if ev.Modifiers()&tcell.ModShift != 0 {
		d = [2]int{d[0] * 8, d[1] * 8}
	}
	if v.zoom == zRegion && v.g.zoom == zRegion && v.g.scale > 1 {
		d = [2]int{d[0] * v.g.scale, d[1] * v.g.scale}
	}
	v.inspect = true
	v.cur.X = clamp(v.cur.X+d[0], 0, s.w.W-1)
	v.cur.Y = clamp(v.cur.Y+d[1], 0, s.w.H-1)
	return true
}

// jump moves the cursor to the next (dir 1) or previous inspect target.
func (v *view) jump(s *scene, dir int) {
	ts := s.targets
	if v.zoom == zRegion {
		ts = append(append([]mapmodel.Pt{{X: s.w.CX, Y: s.w.CY}}, s.siteTargets()...), ts...)
	}
	if len(ts) == 0 {
		return
	}
	v.inspect = true
	n := len(ts) - 1 // a step back from nowhere lands on the last
	if dir > 0 {
		n = 0
	}
	for i, t := range ts {
		if t == v.cur {
			n = ((i+dir)%len(ts) + len(ts)) % len(ts)
			break
		}
	}
	v.cur = ts[n]
}
