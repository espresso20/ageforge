package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// View holds what the player chose, not what the game is: the selected limb
// (the inspect verb), whether to mark growth since the last visit, and the
// animation clock.
type View struct {
	Sel   int     // index into Organism.Limbs; -1 for none
	Since bool    // highlight growth since Prev
	T     float64 // seconds, for animation only; 0 draws the still frame
}

// Render draws o (and its growth since prev, which may be nil) into a W x H
// canvas. It is a pure function of its arguments.
func Render(o, prev *Organism, v View, W, H int) *Canvas {
	d := dialFor(o.EpochKey)
	p := newPalette(d)
	c := NewCanvas(W, H, p.Bg)
	if W < 24 || H < 8 {
		c.Text(0, 0, "organism: too small", p.Dim, 9, 0)
		return c
	}
	r := &renderer{c: c, o: o, prev: prev, v: v, d: d, p: p, W: W, H: H}
	r.layout()
	r.plan()
	r.sky()
	r.strata()
	r.neighbors()
	r.roots()
	r.tree()
	r.fruit()
	r.weather()
	r.labels()
	r.hud()
	if v.Sel >= 0 && v.Sel < len(o.Limbs) {
		r.inspect()
	}
	return c
}

type renderer struct {
	c       *Canvas
	o, prev *Organism
	v       View
	d       dial
	p       palette
	W, H    int
	mini    bool

	top, ground, bottom int     // sky rows [top, ground), soil rows [ground, bottom)
	cx, gy              float64 // trunk base, in cells
	ux, uy              float64 // cells per tree unit, across and up (cells are 1:2)
	stature             float64
	trunkW              float64 // trunk half-width at the base, in columns
	trunkTop            float64 // units
	lean                float64
	youth               float64 // 0.3..1: how grown the sapling is

	limbs   []limbGeom
	roots_  []seg
	crownY  float64 // crown centre height, units
	strataB []band
}

type limbGeom struct {
	li     int   // index into o.Limbs
	segs   []seg // the revealed segments, laid out
	n      int   // revealed
	prevN  int   // revealed at the last visit
	col    tcell.Color
	tip    [2]float64 // outermost revealed point, cells
	hasTip bool
	dimmed bool
	isSel  bool
	vigor  float64
}

type band struct {
	s      Stratum
	y0, y1 int // rows [y0, y1)
	bg     tcell.Color
}

func (r *renderer) layout() {
	r.mini = r.W < 64 || r.H < 20
	r.top = 1
	footer := 2
	if r.mini {
		footer = 0
	}
	r.bottom = r.H - footer
	sceneH := r.bottom - r.top
	frac := 0.72
	if r.mini {
		frac = 0.76
	}
	r.ground = r.top + int(math.Round(float64(sceneH)*frac))
	r.cx = float64(r.W) / 2
	r.gy = float64(r.ground)
	r.uy = float64(r.ground-r.top) * 0.97
	r.ux = r.uy * 2.5 // a spreading crown suits a wide terminal
	ageFrac := float64(r.o.AgeIdx) / float64(max(1, r.o.AgeCount-1))
	r.stature = 0.6 + 0.4*math.Pow(ageFrac, 0.8)
	// A young camp is a sapling: the first sixty buildings grow the tree to
	// its age's stature. After that only the age changes its size.
	r.youth = math.Min(1, 0.3+0.7*float64(r.o.LimbTotal()+r.o.Housing+r.o.Storage)/60)
	r.trunkTop = r.stature * 0.34 * r.youth
	r.lean = (float64(r.seed("lean")%100)/100 - 0.5) * 0.08
}

// at maps tree units to fractional cells.
func (r *renderer) at(x, y float64) (float64, float64) {
	return r.cx + x*r.ux, r.gy - y*r.uy
}

func (r *renderer) seed(parts ...string) uint64 {
	return hash64(append([]string{fmt.Sprint(r.o.Seed)}, parts...)...)
}

func (r *renderer) h(parts ...any) uint64 {
	s := make([]string, 0, len(parts)+1)
	s = append(s, fmt.Sprint(r.o.Seed))
	for _, p := range parts {
		s = append(s, fmt.Sprint(p))
	}
	return hash64(s...)
}

func frac01(h uint64) float64 { return float64(h%100000) / 100000 }

// plan lays out every limb and fits the crown to the sky. Each lineage
// owns a wedge of the crown in crown order, its width the lineage's share
// (sub-linear, so a small lineage still shows); its reach is how far its
// genome has grown. The fit uses a count-free bound on reach, so building
// one more farm never rescales the tree.
func (r *renderer) plan() {
	since := r.prev != nil && r.v.Since
	prevCount := map[int]int{}
	if since {
		for _, l := range r.prev.Limbs {
			prevCount[l.Index] = l.Count
		}
	}
	L := r.stature * 0.27 * r.youth
	// Reach bound: the longest path through the first 45 buildings' worth
	// of any genome, times the bough.
	reach := 0.0
	gens := make([]genome, len(limbSpecs))
	for slot, sp := range limbSpecs {
		gens[slot] = r.o.genome(r.seed("limb", sp.Key), 7)
		g := gens[slot]
		n := reveal(45, len(g.segs))
		path := make([]float64, n)
		for i := 0; i < n; i++ {
			path[i] = g.segs[i].Len
			if p := g.segs[i].Parent; p >= 0 {
				path[i] += path[p]
			}
			reach = math.Max(reach, path[i])
		}
	}
	reach *= L
	fitX := (float64(r.W)/2 - 2) / ((reach*0.95 + 0.05) * r.ux)
	// Leave room above for the top clusters' foliage (about uy/14 rows).
	fitY := (r.gy - float64(r.top) - 2) / ((r.trunkTop + reach*0.95 + 1.0/14) * r.uy)
	if k := math.Min(1, math.Min(fitX, fitY)); k < 1 {
		r.ux *= k
		r.uy *= k
	}
	r.trunkW = 0.6 + 3.2*math.Sqrt(math.Min(1, float64(r.o.Pop)/6000))*r.uy/28
	r.crownY = r.trunkTop + reach*0.45

	// Wedges.
	n := len(r.o.Limbs)
	fan := math.Min(1.3, 0.55+0.1*float64(n))
	var total float64
	weights := make([]float64, n)
	for i, l := range r.o.Limbs {
		weights[i] = math.Pow(float64(max(1, l.Count)), 0.6)
		total += weights[i]
	}
	acc := 0.0
	for li, l := range r.o.Limbs {
		half := fan * weights[li] / total
		centre := -fan + 2*fan*acc/total + half
		acc += weights[li]
		g := gens[l.Index]
		ay := r.trunkTop * (0.7 + 0.3*(1-math.Abs(centre)/1.3))
		lg := limbGeom{li: li, n: reveal(l.Count, len(g.segs)), col: r.p.Limb[l.Index], vigor: l.Vigor()}
		lg.segs = g.layout(lg.n, 0, ay, centre, half*0.9, L, 0.12)
		lg.prevN = lg.n
		if since {
			lg.prevN = reveal(prevCount[l.Index], len(g.segs))
		}
		if r.v.Sel >= 0 {
			lg.isSel = r.v.Sel == li
			lg.dimmed = !lg.isSel
		}
		r.limbs = append(r.limbs, lg)
	}

	// Roots: storage feeds them; they reach down through history.
	rootL := float64(r.bottom-r.ground) / r.uy * 0.32 * math.Sqrt(r.youth)
	rg := r.o.genome(r.seed("roots"), 6)
	r.roots_ = rg.layout(reveal(r.o.Storage+2, len(rg.segs)), 0, 0, math.Pi, 1.25, rootL, 0)
}

// sky: sun or moon by game time, stars at night and in the late epochs.
func (r *renderer) sky() {
	c, p := r.c, r.p
	if r.o.PendingCatastrophe != "" {
		for y := r.top; y < r.ground; y++ {
			t := float64(y-r.top) / float64(max(1, r.ground-r.top))
			for x := 0; x < r.W; x++ {
				c.Bg(x, y, theme.Mix(p.Bg, p.Negative, 0.04+0.10*t))
			}
		}
	}
	// A day is an hour of play (1800 ticks at 2s).
	phase := float64(r.o.Tick%1800) / 1800
	night := phase >= 0.5
	if r.d.Stars || night {
		n := r.W * (r.ground - r.top) / 90
		rg := &rng{s: r.seed("stars")}
		for i := 0; i < n; i++ {
			x := int(rg.next() % uint64(r.W))
			y := r.top + int(rg.next()%uint64(r.ground-r.top))
			kind := rg.next() % 10
			g, col := '·', p.Dim
			switch {
			case kind == 0:
				g, col = '+', p.Text
			case kind < 4:
				g = '.'
			}
			if r.v.T > 0 && (int(r.v.T*2)+i)%9 == 0 {
				g = '✦'
				col = p.Text
			}
			c.Glyph(x, y, g, col, 0)
		}
	}
	if r.d.Leaf == leafStar {
		return // no sun in deep space
	}
	f := math.Mod(phase*2, 1) // 0..1 across the sky
	sx := 2 + int(f*float64(r.W-5))
	sy := r.top + int(float64(r.ground-r.top-3)*(1-math.Sin(f*math.Pi))*0.7)
	if night {
		c.Glyph(sx, sy, '☾', p.Text, 6)
	} else {
		c.GlyphA(sx, sy, '☼', p.Highlight, 6, tcell.AttrBold)
	}
}

// strata: every epoch reached is a layer of soil, newest on top, its
// thickness the number of ages lived in it; each age inside is a shade of
// the layer, and the epoch's events are fossils.
func (r *renderer) strata() {
	c, p := r.c, r.p
	rows := r.bottom - r.ground
	if rows <= 0 {
		return
	}
	n := len(r.o.Strata)
	ageFrac := float64(r.o.AgeIdx) / float64(max(1, r.o.AgeCount-1))
	depth := int(math.Round(float64(rows) * (0.4 + 0.6*ageFrac)))
	depth = max(depth, min(rows, n))
	alloc := make([]int, n)
	left := depth
	var sum float64
	for i, s := range r.o.Strata {
		sum += float64(max(1, s.Ticks))
		if left > 0 {
			alloc[i]++
			left--
		}
	}
	rest := left
	for i, s := range r.o.Strata {
		add := int(math.Floor(float64(rest) * float64(max(1, s.Ticks)) / sum))
		alloc[i] += add
		left -= add
	}
	for i := n - 1; left > 0 && i >= 0; i-- {
		alloc[i]++
		left--
	}
	y := r.ground
	for i := n - 1; i >= 0; i-- { // newest on top
		s := r.o.Strata[i]
		b := band{s: s, y0: y, y1: min(r.bottom, y+alloc[i]), bg: p.soil(s.EpochKey, i)}
		y = b.y1
		r.strataB = append(r.strataB, b)
	}
	bedrock := theme.Mix(p.Bg, p.Dim, 0.06)
	for yy := y; yy < r.bottom; yy++ {
		for x := 0; x < r.W; x++ {
			c.Bg(x, yy, bedrock)
		}
	}
	for bi, b := range r.strataB {
		rowsB := b.y1 - b.y0
		ages := len(b.s.Ages)
		for yy := b.y0; yy < b.y1; yy++ {
			// Ages are shades inside the layer when there is room for them;
			// newest age on top.
			shade := b.bg
			if ages > 1 && rowsB >= ages {
				k := (yy - b.y0) * ages / rowsB
				shade = theme.Mix(b.bg, p.Bg, 0.12*float64(k))
			}
			for x := 0; x < r.W; x++ {
				c.Bg(x, yy, shade)
			}
		}
		// Pebbles.
		peb := theme.Mix(b.bg, p.Text, 0.2)
		rg := &rng{s: r.seed("pebbles", b.s.EpochKey)}
		for k := 0; k < r.W*rowsB/30; k++ {
			x := int(rg.next() % uint64(r.W))
			yy := b.y0 + int(rg.next()%uint64(max(1, rowsB)))
			g := []rune{'.', '·', ',', '˙', '°'}[rg.next()%5]
			c.Glyph(x, yy, g, peb, 0)
		}
		if !r.mini {
			label := strings.ToLower(b.s.Name)
			if rowsB >= 2 || bi == 0 {
				label += ": " + strings.ToLower(strings.Join(shortAges(b.s.Ages), ", "))
			}
			col := theme.Legible(theme.Mix(rgb(dialFor(b.s.EpochKey).Hue), p.Text, 0.3), b.bg, 3.5)
			lx := r.W - len([]rune(label)) - 1
			c.Text(lx, b.y0, label, col, 5, tcell.AttrItalic)
		}
		r.fossils(b)
	}
	// Grass, while the tree is still a tree.
	if r.d.Branch == styleGrown && len(r.strataB) > 0 {
		grass := theme.Legible(theme.Mix(p.Positive, p.Bg, 0.35), p.Bg, 2.2)
		rg := &rng{s: r.seed("grass")}
		for x := 0; x < r.W; x++ {
			if rg.next()%4 == 0 {
				c.Glyph(x, r.ground-1, []rune{'"', ',', '\'', '`'}[rg.next()%4], grass, 0)
			}
		}
	}
}

func shortAges(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = strings.TrimSuffix(n, " Age")
	}
	return out
}

// Fossil glyphs. Shape carries the meaning, so the capture reads in
// monochrome: open marks for blessings, crosses and bursts for blows.
func fossilGlyph(f Fossil) (rune, string) {
	switch f.Kind {
	case "good_minor":
		return '∘', "pos"
	case "good_major":
		return '✧', "pos"
	case "good_legendary":
		return '❖', "gold"
	case "bad_challenging":
		return '†', "warn"
	case "catastrophe":
		switch f.Outcome {
		case "succumbed":
			return '✕', "neg"
		case "pending":
			return '!', "neg"
		}
		return '✹', "neg"
	case "harbinger":
		return '◉', "accent"
	}
	return '·', "dim"
}

func (r *renderer) fossils(b band) {
	p := r.p
	for _, f := range r.o.Fossils {
		if f.EpochKey != b.s.EpochKey {
			continue
		}
		g, class := fossilGlyph(f)
		col := map[string]tcell.Color{"pos": p.Positive, "gold": p.Gold, "warn": p.Warning, "neg": p.Negative, "accent": p.Accent}[class]
		if col == 0 {
			col = p.Dim
		}
		col = theme.Legible(col, b.bg, 3.0)
		h := r.h("fossil", f.Name, f.Tick)
		rows := b.y1 - b.y0
		if rows <= 0 {
			return
		}
		y := b.y0 + int(h%uint64(rows))
		x := 2 + int((h>>16)%uint64(max(1, r.W-30)))
		if math.Abs(float64(x)-r.cx) < 4 {
			x += 8
		}
		for tries := 0; tries < 12 && !r.c.Free(x, y, 1, 4); tries++ {
			x = 2 + (x+5)%max(1, r.W-4)
		}
		r.c.GlyphA(x, y, g, col, 4, tcell.AttrBold)
	}
}

// stroke draws a line of branch glyphs between two cell points. The glyph
// follows the line's visual slope (a cell is twice as tall as it is wide),
// and width fattens it across its direction.
func (r *renderer) stroke(x0, y0, x1, y1, width float64, col tcell.Color, pri int8, style branchStyle) {
	dx, dy := x1-x0, y1-y0
	vx, vy := dx/2, -dy
	theta := math.Atan2(math.Abs(vx), math.Abs(vy)) // 0 = vertical
	var g rune
	slash := vx*vy > 0
	switch style {
	case styleTrace:
		switch {
		case theta < 0.40:
			g = '│'
		case theta > 1.17:
			g = '─'
		case slash:
			g = '╱'
		default:
			g = '╲'
		}
	case styleStellar:
		g = '·'
	default:
		switch {
		case theta < 0.40:
			g = '|'
		case theta > 1.17:
			g = '~'
			if style == styleForged {
				g = '─'
			}
		case slash:
			g = '/'
		default:
			g = '\\'
		}
	}
	n := int(math.Max(math.Abs(dx), math.Abs(dy))*2) + 1
	vertical := theta < 0.8
	for i := 0; i <= n; i++ {
		if style == styleStellar && i%3 != 0 {
			continue
		}
		t := float64(i) / float64(n)
		x, y := x0+dx*t, y0+dy*t
		cx, cy := int(math.Floor(x)), int(math.Floor(y))
		r.c.Glyph(cx, cy, g, col, pri)
		if width >= 0.75 {
			k := int(width)
			for j := 1; j <= k; j++ {
				if vertical {
					r.c.Glyph(cx-j, cy, g, col, pri)
					r.c.Glyph(cx+j, cy, g, col, pri)
				} else if j <= k/2 {
					r.c.Glyph(cx, cy-j, g, col, pri)
					r.c.Glyph(cx, cy+j, g, col, pri)
				}
			}
		}
	}
}

// segCells samples a (bowed) segment into cell points.
func (r *renderer) segCells(s seg, steps int) [][2]float64 {
	mx, my := (s.X0+s.X1)/2, (s.Y0+s.Y1)/2
	dx, dy := s.X1-s.X0, s.Y1-s.Y0
	bend := s.Bend
	if r.d.Branch != styleGrown {
		bend = 0
	}
	cxu, cyu := mx-dy*bend, my+dx*bend
	pts := make([][2]float64, 0, steps+1)
	for k := 0; k <= steps; k++ {
		t := float64(k) / float64(steps)
		u := 1 - t
		x := u*u*s.X0 + 2*u*t*cxu + t*t*s.X1
		y := u*u*s.Y0 + 2*u*t*cyu + t*t*s.Y1
		cx, cy := r.at(x, y)
		pts = append(pts, [2]float64{cx, cy})
	}
	return pts
}

// roots: storage grows the root system down through the strata.
func (r *renderer) roots() {
	col := r.p.Root
	if r.v.Sel >= 0 {
		col = theme.Mix(col, r.p.Bg, 0.5)
	}
	for _, s := range r.roots_ {
		s.X0, s.X1 = s.X0*1.5, s.X1*1.5
		pts := r.segCells(s, 3)
		w := math.Max(0, r.trunkW*0.8*math.Pow(0.55, float64(s.Depth+1)))
		for k := 1; k < len(pts); k++ {
			style := r.d.Branch
			if style == styleStellar {
				style = styleForged // roots stay roots, even among the stars
			}
			r.stroke(pts[k-1][0], pts[k-1][1]+0.5, pts[k][0], pts[k][1]+0.5, w, col, 2, style)
		}
	}
}

// tree: trunk, limbs and leaves.
func (r *renderer) tree() {
	c, p := r.c, r.p
	wood := p.Wood
	if r.v.Sel >= 0 {
		wood = theme.Mix(wood, p.Bg, 0.4)
	}
	// Trunk: a tapered column of bark, flared at the foot.
	topRow := int(math.Floor(r.gy - r.trunkTop*r.uy))
	baseRow := r.ground - 1
	for y := baseRow; y >= topRow; y-- {
		t := float64(baseRow-y) / float64(max(1, baseRow-topRow))
		xc := r.cx + r.lean*math.Sin(t*math.Pi)*r.ux
		hw := r.trunkW * (1 - 0.4*t)
		if y == baseRow {
			hw += 1.2
		}
		x0, x1 := int(math.Floor(xc-hw)), int(math.Floor(xc+hw))
		for x := x0; x <= x1; x++ {
			g := r.bark(x, y, x0, x1, y == baseRow)
			c.Glyph(x, y, g, wood, 3)
		}
		// Sap: rising pulses, more with people, faster with morale.
		if r.v.T > 0 {
			pulses := min(8, 1+r.o.Pop/600)
			for k := 0; k < pulses; k++ {
				f := math.Mod(r.v.T*(0.2+0.25*r.o.Morale)+float64(k)/float64(pulses), 1)
				if int(f*float64(baseRow-topRow+1)) == baseRow-y {
					c.GlyphA(int(math.Floor(xc)), y, '•', p.Highlight, 6, tcell.AttrBold)
				}
			}
		}
	}

	// Limbs, back to front: dimmed ones first so the selection sits on top.
	order := make([]int, len(r.limbs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return !r.limbs[order[a]].isSel && r.limbs[order[b]].isSel })
	for _, i := range order {
		r.drawBranches(&r.limbs[i], wood)
	}
	for _, i := range order {
		r.drawLeaves(&r.limbs[i])
	}
	if len(r.limbs) == 0 {
		// Seed leaves: the camp before its first lineage.
		sprout := theme.Legible(p.Positive, p.Bg, 3.0)
		tx := int(math.Floor(r.cx + r.lean*r.ux*0.0))
		c.Glyph(tx-1, topRow-1, '\\', wood, 3)
		c.Glyph(tx+1, topRow-1, '/', wood, 3)
		for _, d := range [][2]int{{-3, -2}, {-2, -2}, {-4, -1}, {-3, -1}, {2, -2}, {3, -2}, {3, -1}, {4, -1}, {0, -2}} {
			c.GlyphA(tx+d[0], topRow+d[1], '♣', sprout, 4, tcell.AttrBold)
		}
	}
}

func (r *renderer) bark(x, y, x0, x1 int, base bool) rune {
	switch r.d.Branch {
	case styleTrace:
		if x == x0 || x == x1 {
			return '┃'
		}
		if (y+x)%3 == 0 {
			return '╂'
		}
		return '┃'
	case styleStellar:
		return '║'
	case styleForged:
		if x == x0 || x == x1 {
			return '║'
		}
		return '|'
	}
	if base {
		switch {
		case x == x0:
			return '/'
		case x == x1:
			return '\\'
		}
		return '_'
	}
	if x == x0 || x == x1 {
		return '|'
	}
	h := hxy(r.seed("bark"), x, y)
	return []rune{'|', '|', ':', '\'', '|', '.'}[h%6]
}

func (r *renderer) limbColors(lg *limbGeom, wood tcell.Color) (branch, leaf, fresh tcell.Color) {
	p := r.p
	leaf = lg.col
	// Health: a starving or demoralised civ wilts, every limb at once.
	if r.o.FoodRate < 0 {
		leaf = theme.Mix(leaf, p.Warning, 0.5)
	} else if r.o.Morale < 0.3 {
		leaf = theme.Mix(leaf, p.Dim, 0.4)
	}
	branch = theme.Mix(wood, lg.col, 0.25)
	fresh = theme.Legible(theme.Mix(lg.col, p.Text, 0.55), p.Bg, 4.5)
	if lg.dimmed {
		leaf = theme.Mix(leaf, p.Bg, 0.8)
		branch = theme.Mix(branch, p.Bg, 0.75)
		fresh = theme.Mix(fresh, p.Bg, 0.8)
	}
	if lg.isSel {
		branch = theme.Mix(r.p.Wood, lg.col, 0.5)
	}
	return
}

func (r *renderer) drawBranches(lg *limbGeom, wood tcell.Color) {
	branch, _, fresh := r.limbColors(lg, wood)
	pri := int8(3)
	if lg.isSel {
		pri = 6
	}
	for i := 0; i < lg.n; i++ {
		s := lg.segs[i]
		w := r.trunkW * 0.7 * math.Pow(0.5, float64(s.Depth))
		if r.d.Branch == styleStellar {
			w = 0 // constellations are drawn with a fine pen
		}
		col := branch
		if i >= lg.prevN {
			col = fresh
		}
		pts := r.segCells(s, 4)
		if r.d.Branch == styleTrace {
			// Circuit traces: 45 degrees first, then straight, with a via.
			a, b := pts[0], pts[len(pts)-1]
			ddx, ddy := b[0]-a[0], b[1]-a[1]
			d := math.Min(math.Abs(ddx)/2, math.Abs(ddy))
			mx := a[0] + math.Copysign(d*2, ddx)
			my := a[1] + math.Copysign(d, ddy)
			r.stroke(a[0], a[1], mx, my, 0, col, pri, styleTrace)
			r.stroke(mx, my, b[0], b[1], 0, col, pri, styleTrace)
			r.c.Glyph(int(math.Floor(mx)), int(math.Floor(my)), '•', col, pri)
		} else {
			for k := 1; k < len(pts); k++ {
				r.stroke(pts[k-1][0], pts[k-1][1], pts[k][0], pts[k][1], w, col, pri, r.d.Branch)
			}
		}
		if r.d.Branch == styleStellar {
			// Constellations: every joint is a star.
			node := '+'
			if s.Depth <= 2 {
				node = '✦'
			}
			e := pts[len(pts)-1]
			star := theme.Mix(lg.col, r.p.Text, 0.35)
			if lg.dimmed {
				star = theme.Mix(star, r.p.Bg, 0.6)
			}
			r.c.GlyphA(int(math.Floor(e[0])), int(math.Floor(e[1])), node, star, pri, tcell.AttrBold)
		}
		end := pts[len(pts)-1]
		score := func(pt [2]float64) float64 { return math.Abs(pt[0]-r.cx)/2 + (r.gy - pt[1]) }
		if !lg.hasTip || score(end) > score(lg.tip) {
			lg.tip, lg.hasTip = end, true
		}
	}
}

// drawLeaves: a cluster of the lineage's glyph at every growing tip. How
// full a cluster is shows how well the lineage is staffed.
func (r *renderer) drawLeaves(lg *limbGeom) {
	c, p := r.c, r.p
	l := r.o.Limbs[lg.li]
	_, leaf, fresh := r.limbColors(lg, p.Wood)
	shadow := theme.Mix(leaf, p.Bg, 0.3)
	pri := int8(4)
	if lg.isSel {
		pri = 7
	}
	ry := math.Max(0.6, r.uy/14)
	rx := ry * 2.0
	fill := 0.35 + 0.6*lg.vigor
	glyph := l.Spec.Leaf
	leafKey := r.seed("leaf", l.Spec.Key)
	tick := int(r.v.T * 3)
	type cluster struct {
		x, y, k float64
		seg     int
	}
	var cls []cluster
	isTip := map[int]bool{}
	for _, ti := range tips(lg.segs) {
		isTip[ti] = true
		s := lg.segs[ti]
		x, y := r.at(s.X1, s.Y1)
		cls = append(cls, cluster{x, y, 1, ti})
	}
	// Foliage along the outer twigs too, so the crown has volume.
	for i := 0; i < lg.n; i++ {
		s := lg.segs[i]
		if isTip[i] || s.Depth < 2 {
			continue
		}
		x, y := r.at((s.X0+s.X1)/2, (s.Y0+s.Y1)/2)
		cls = append(cls, cluster{x, y, 0.65, i})
	}
	for _, cl := range cls {
		tx, ty := cl.x, cl.y
		ry, rx := ry*cl.k, rx*cl.k
		isFresh := cl.seg >= lg.prevN
		for y := int(math.Floor(ty - ry - 1)); y <= int(math.Ceil(ty+ry)); y++ {
			for x := int(math.Floor(tx - rx - 1)); x <= int(math.Ceil(tx+rx)); x++ {
				ddx := (float64(x) + 0.5 - tx) / rx
				ddy := (float64(y) + 0.5 - ty) / ry
				d2 := ddx*ddx + ddy*ddy
				if d2 > 1 {
					continue
				}
				h := hxy(leafKey, x, y)
				if frac01(h) > fill*(1-0.45*d2) {
					continue
				}
				col := leaf
				att := tcell.AttrNone
				switch {
				case isFresh:
					col, att = fresh, tcell.AttrBold
				case (h>>17)%3 == 0:
					col = shadow
				}
				if r.v.T > 0 && int((h>>23)%13) == tick%13 {
					col = theme.Mix(col, p.Text, 0.45) // leaves catch the light
				}
				if r.d.Leaf == leafStar && (h>>29)%5 == 0 {
					c.GlyphA(x, y, '✦', col, pri, att)
					continue
				}
				c.GlyphA(x, y, glyph, col, pri, att)
			}
		}
	}
	// Starving: a few leaves fall.
	if r.o.FoodRate < 0 {
		tips := tips(lg.segs)
		for k := 0; k < 6 && len(tips) > 0; k++ {
			h := r.h("fall", l.Spec.Key, k)
			s := lg.segs[tips[int(h%uint64(len(tips)))]]
			tx, ty := r.at(s.X1, s.Y1)
			f := math.Mod(r.v.T*0.15+frac01(h>>8), 1)
			y := ty + f*(r.gy-ty)
			c.Glyph(int(tx)+int(f*3), int(y), ',', p.Warning, 5)
		}
	}
}

// fruit: wonders. A built wonder hangs as a golden fruit from a fixed
// segment of a fixed limb; the age's unbuilt wonder ripens with its bank.
func (r *renderer) fruit() {
	if len(r.limbs) == 0 {
		return
	}
	used := map[[2]int]bool{}
	for _, w := range r.o.Wonders {
		h := r.seed("wonder", w.Key)
		lg := r.limbs[int(h%uint64(len(r.limbs)))]
		if alt := r.limbs[int((h>>13)%uint64(len(r.limbs)))]; alt.n > lg.n {
			lg = alt // prefer the heavier of two draws
		}
		si := 1 + int((h>>7)%12)
		for si >= lg.n && si > 0 {
			si = lg.segs[si].Parent
		}
		si = max(si, 0)
		s := lg.segs[si]
		x, y := r.at(s.X1, s.Y1)
		cx, cy := int(x), int(y)+1
		for used[[2]int{cx, cy}] {
			cx += 2
		}
		used[[2]int{cx, cy}] = true
		g, col := '●', r.p.Gold
		if !w.Built {
			col = theme.Mix(r.p.Gold, r.p.Bg, 0.3)
			switch {
			case w.Progress >= 0.75:
				g = '◕'
			case w.Progress >= 0.5:
				g = '◑'
			case w.Progress >= 0.25:
				g = '◔'
			default:
				g = '○'
			}
		}
		if r.v.Sel >= 0 {
			col = theme.Mix(col, r.p.Bg, 0.5)
		}
		r.c.GlyphA(cx, cy, g, col, 7, tcell.AttrBold)
	}
}

// neighbors: every civ met stands on the horizon as a conifer, sized by
// strength and coloured by relation: war, friend, wary, or neutral.
type nbPos struct {
	n    Neighbor
	x, y int // top of the crown
	col  tcell.Color
}

func (r *renderer) neighborPositions() []nbPos {
	var out []nbPos
	for i, n := range r.o.Neighbors {
		side := 1.0
		if i%2 == 0 {
			side = -1
		}
		k := float64(i / 2)
		x := r.cx + side*(float64(r.W)/2-4-k*float64(r.W)*0.085)
		col := r.p.Dim
		switch {
		case n.AtWar:
			col = r.p.Negative
		case n.Opinion >= 40:
			col = r.p.Positive
		case n.Opinion <= -20:
			col = r.p.Warning
		}
		hgt := 1 + (n.Strength+1)/2
		out = append(out, nbPos{n, int(x), r.ground - 1 - hgt, col})
	}
	return out
}

func (r *renderer) neighbors() {
	c := r.c
	for _, np := range r.neighborPositions() {
		col := np.col
		if r.v.Sel >= 0 {
			col = theme.Mix(col, r.p.Bg, 0.5)
		}
		hgt := r.ground - 1 - np.y
		for k := 0; k < hgt; k++ {
			for dx := -k; dx <= k; dx++ {
				c.Glyph(np.x+dx, np.y+k, '▲', col, 1)
			}
		}
		c.Glyph(np.x, r.ground-1, '|', theme.Mix(col, r.p.Bg, 0.3), 1)
		if !r.mini {
			name := np.n.Name
			if strings.HasPrefix(name, "The ") {
				name = name[4:]
			}
			if i := strings.IndexByte(name, ' '); i > 0 {
				name = name[:i]
			}
			lx := max(0, min(r.W-len(name), np.x-len(name)/2))
			if c.Free(lx, np.y-1, len(name), 2) {
				c.Text(lx, np.y-1, name, col, 2, tcell.AttrItalic)
			}
		}
	}
}

// weather: trade pollen, war sparks, the harbinger's comet, the storm of a
// pending catastrophe, and motes for active events.
func (r *renderer) weather() {
	c, p := r.c, r.p
	nbs := r.neighborPositions()
	crownX, crownY := r.at(0, r.crownY)

	arc := func(x0, y0, x1, y1, lift, t float64) (float64, float64) {
		return x0 + (x1-x0)*t, y0 + (y1-y0)*t - lift*math.Sin(t*math.Pi)
	}
	for _, rt := range r.o.Routes {
		// Rendezvous hashing: a route keeps its partner as new civs are met.
		tx, ty := float64(r.W-1), float64(r.ground-3)
		if hash64(rt.Key)%2 == 0 {
			tx = 0
		}
		best := uint64(0)
		for _, nb := range nbs {
			if hh := hash64(rt.Key, nb.n.Key); hh >= best {
				best, tx, ty = hh, float64(nb.x), float64(nb.y)
			}
		}
		// Pollen leaves from the crown's edge on the partner's side.
		sx, sy := crownX, crownY
		for _, lg := range r.limbs {
			for _, s := range lg.segs {
				x, y := r.at(s.X1, s.Y1)
				if (tx > crownX && x > sx) || (tx < crownX && x < sx) {
					sx, sy = x, y
				}
			}
		}
		crownX, crownY := sx, sy
		lift := math.Abs(tx-crownX) * 0.12
		col := p.Highlight
		if rt.Disrupted {
			col = p.Negative
		}
		path := theme.Mix(col, p.Bg, 0.45)
		steps := int(math.Abs(tx-crownX)) + 1
		for k := 0; k <= steps; k += 2 {
			x, y := arc(crownX, crownY, tx, ty, lift, float64(k)/float64(steps))
			c.Glyph(int(x), int(y), '·', path, 0) // under labels and civs
		}
		if !rt.Disrupted {
			for j := 0; j < 3; j++ {
				f := math.Mod(r.v.T*0.06+float64(j)/3+frac01(hash64(rt.Key)), 1)
				mcol := p.Highlight
				if j%2 == 1 {
					f, mcol = 1-f, p.Accent // imports flow home
				}
				x, y := arc(crownX, crownY, tx, ty, lift, f)
				c.GlyphA(int(x), int(y), '•', mcol, 6, tcell.AttrBold)
			}
		}
	}
	for _, nb := range nbs {
		if !nb.n.AtWar {
			continue
		}
		gx := r.cx + (float64(nb.x)-r.cx)*0.5
		for k := 0; k < 3; k++ {
			f := math.Mod(r.v.T*0.4+float64(k)/3, 1)
			x := gx + (float64(nb.x)-gx)*f
			c.GlyphA(int(x), r.ground-1, '×', p.Negative, 6, tcell.AttrBold)
		}
	}

	if hb := r.o.Harbinger; hb != nil {
		// The comet: its tail lengthens with the odds of the blow.
		tail := 6 + int(22*hb.Probability)
		hx := r.W - tail - 2
		hy := r.top + 1
		for k := 1; k <= tail; k++ {
			g := '·'
			switch {
			case k <= 2:
				g = '═'
			case k <= tail/2:
				g = '-'
			}
			if r.v.T > 0 && (int(r.v.T*5)+k)%5 == 0 {
				continue
			}
			col := theme.Mix(p.Warning, p.Bg, float64(k)/float64(tail)*0.7)
			c.Glyph(hx+k, hy-k/5, g, col, 5)
		}
		c.GlyphA(hx, hy, '✶', p.Warning, 7, tcell.AttrBold)
		if !r.mini {
			lbl := "harbinger · " + strings.ToLower(hb.Name)
			n := len([]rune(lbl))
			for _, sp := range [][2]int{{hx - n - 2, hy}, {r.W - n - 1, hy + 1}, {r.W - n - 1, hy - 1}} {
				if sp[0] > 0 && c.Free(sp[0]-1, sp[1], n+2, 3) {
					c.Text(sp[0], sp[1], lbl, theme.Mix(p.Warning, p.Bg, 0.15), 5, tcell.AttrItalic)
					break
				}
			}
		}
	}

	if r.o.PendingCatastrophe != "" {
		// Lightning into the crown, flashing; the crown sheds.
		flash := r.v.T == 0 || math.Mod(r.v.T+0.2, 1.2) < 0.5
		if flash {
			x := r.cx + float64(int(r.seed("bolt")%16)) - 8
			for y := r.top; float64(y) < crownY-2; y++ {
				step := int(r.h("bolt", y)%3) - 1
				g := '│'
				if step < 0 {
					g = '╱'
				} else if step > 0 {
					g = '╲'
				}
				c.GlyphA(int(x), y, g, p.Highlight, 9, tcell.AttrBold)
				x -= float64(step)
			}
		}
		for k := 0; k < 14; k++ {
			h := r.h("shed", k)
			fx := crownX + float64(int(h%uint64(max(1, int(r.ux*0.9))))) - r.ux*0.45
			f := math.Mod(r.v.T*0.18+frac01(h>>9), 1)
			fy := crownY + f*(r.gy-crownY)
			c.Glyph(int(fx+math.Sin(r.v.T*2+float64(k))*2), int(fy), []rune{',', '`', '\''}[k%3], p.Warning, 6)
		}
	}
	for i, ev := range r.o.Events {
		for k := 0; k < 3; k++ {
			h := r.h("mote", ev, k)
			f := math.Mod(r.v.T*0.1+frac01(h), 1)
			x := crownX + float64(int((h>>8)%uint64(max(1, int(r.ux))))) - r.ux/2 + float64(i*3)
			y := crownY - 2 - f*(crownY-2-float64(r.top))
			g := '·'
			if k == 0 {
				g = '✧'
			}
			c.Glyph(int(x), int(y), g, p.Highlight, 5)
		}
	}
}

// hxy hashes a cell position under a key without allocating: the leaf and
// bark textures call it for every cell they touch.
func hxy(key uint64, x, y int) uint64 {
	z := key ^ uint64(int64(x))*0x9e3779b97f4a7c15 ^ uint64(int64(y))*0xc2b2ae3d27d4eb4f
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}
