package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// The sky plates. When the empire leaves the planet the atlas zooms out:
// the orrery (the home system), the star chart (the neighbourhood) and the
// galactic atlas. Same grammar as the planet: a field, lines, marks, labels,
// the same colour classes and the same inspector.

type cosmicCtx struct {
	cv     *Canvas
	rc     Rect
	a      *Atlas
	p      *Plate
	pal    *Pal
	v      View
	frame  int
	pois   []POI
	legend []line
	bodies map[int]body // POI ref -> body, for the inspector
}

type body struct {
	Name, Kind string
	At         Pt
	Colony     bool
	Civ        *Civ
	Note       string
}

func cosmicFitScale(p *Plate, rc Rect) float64 {
	R := map[string]float64{"orrery": 58, "stars": 34, "galaxy": 44}[p.Model]
	Ry := R
	if p.Model == "orrery" {
		Ry = R * 0.62
	}
	return math.Max(2*R/float64(rc.W), 2*Ry/float64(2*rc.H)) * 1.04
}

func (c *cosmicCtx) screen(p Pt) (float64, float64) {
	return (p.X-c.v.CX)/c.v.S + float64(c.rc.W)/2, (p.Y-c.v.CY)/(2*c.v.S) + float64(c.rc.H)/2
}

func (c *cosmicCtx) cell(p Pt) (int, int, bool) {
	fx, fy := c.screen(p)
	x, y := int(math.Floor(fx)), int(math.Floor(fy))
	return c.rc.X + x, c.rc.Y + y, x >= 0 && y >= 0 && x < c.rc.W && y < c.rc.H
}

func (c *cosmicCtx) world(i, j int) Pt {
	return Pt{c.v.CX + (float64(i)+0.5-float64(c.rc.W)/2)*c.v.S, c.v.CY + (float64(j)+0.5-float64(c.rc.H)/2)*2*c.v.S}
}

func (c *cosmicCtx) ring(center Pt, r float64, g rune, col tcell.Color, z int8, step int) {
	n := int(2*math.Pi*r/c.v.S) + 8
	for k := 0; k < n; k += step {
		ang := float64(k) / float64(n) * 2 * math.Pi
		if x, y, ok := c.cell(Pt{center.X + math.Cos(ang)*r, center.Y + math.Sin(ang)*r}); ok {
			c.cv.Put(x, y, g, col, z)
		}
	}
}

func (c *cosmicCtx) lane(a, b Pt, col tcell.Color, phase float64, mover bool) {
	ax, ay := c.screen(a)
	bx, by := c.screen(b)
	c.cv.Line(float64(c.rc.X)+ax, float64(c.rc.Y)+ay, float64(c.rc.X)+bx, float64(c.rc.Y)+by, func(i, x, y int, dx, dy float64) (rune, bool) {
		return '·', i%2 == 1
	}, col, zLine)
	if mover {
		t := math.Mod(float64(c.frame)*0.05+phase, 1)
		if x, y, ok := c.cell(Pt{lerp(a.X, b.X, t), lerp(a.Y, b.Y, t)}); ok {
			c.cv.PutA(x, y, c.p.Ship, col, tcell.AttrBold, zMover)
		}
	}
}

func (c *cosmicCtx) label(p Pt, s string, col tcell.Color, attr tcell.AttrMask) {
	if x, y, ok := c.cell(p); ok {
		c.cv.Label(x, y, s, col, attr, zLabel, nearOffsets, [4]int{c.rc.X, c.rc.Y, c.rc.X + c.rc.W, c.rc.Y + c.rc.H})
	}
}

func (c *cosmicCtx) mark(p Pt, g rune, col tcell.Color, attr tcell.AttrMask, poi POI, b body) {
	if x, y, ok := c.cell(p); ok {
		c.cv.PutA(x, y, g, col, attr, zMark)
		c.cv.Reserve(x, y, 1, 1)
	}
	poi.At = p
	poi.Ref = len(c.pois)
	b.At = p
	c.bodies[poi.Ref] = b
	c.pois = append(c.pois, poi)
}

func drawCosmic(cv *Canvas, rc Rect, a *Atlas, p *Plate, pal *Pal, v View, frame int) *cosmicCtx {
	c := &cosmicCtx{cv: cv, rc: rc, a: a, p: p, pal: pal, v: v, frame: frame, bodies: map[int]body{}}
	// the black between: a faint, fixed star field
	for j := 0; j < rc.H; j++ {
		for i := 0; i < rc.W; i++ {
			x, y := rc.X+i, rc.Y+j
			cv.SetBg(x, y, pal.Paper)
			w := c.world(i, j)
			r := rnd01(a.W.Seed+404, int(math.Floor(w.X/v.S)), int(math.Floor(w.Y/(2*v.S))))
			switch {
			case r < 0.006:
				cv.Put(x, y, '∙', pal.mix(pal.Paper, pal.Ink, 0.55), zGround)
			case r < 0.03 && (p.Model != "stars" || r < 0.012):
				col := pal.mix(pal.Paper, pal.Faint, 0.4)
				if (frame+i*7+j*3)%29 == 0 {
					col = pal.Ink // a twinkle
				}
				cv.Put(x, y, '·', col, zGround)
			}
		}
	}
	switch p.Model {
	case "orrery":
		c.orrery()
	case "stars":
		c.stars()
	case "galaxy":
		c.galaxy()
	}
	return c
}

func (c *cosmicCtx) knownCivs(celestial bool) []*Civ {
	var out []*Civ
	for _, cv := range c.a.Civs {
		if cv.Known && cv.Celestial == celestial {
			out = append(out, cv)
		}
	}
	return out
}

func (c *cosmicCtx) earthNote() string {
	var parts []string
	for _, cv := range c.knownCivs(false) {
		parts = append(parts, cv.Def.Name)
	}
	if len(parts) == 0 {
		return ""
	}
	return "shared with " + strings.Join(parts, ", ")
}

// colonies is how many worlds beyond home the realm holds.
func (c *cosmicCtx) colonies(max int) int {
	n := 1 + int(math.Log2(1+float64(c.a.Buildings)/150)) + (c.a.AgeIdx - 17)
	if n > max {
		n = max
	}
	if n < 1 {
		n = 1
	}
	return n
}

func (c *cosmicCtx) orrery() {
	a, pal, cv := c.a, c.pal, c.cv
	const tilt = 0.55
	names := newNamer(a.W.Seed + 911)
	type planet struct {
		au   float64
		g    rune
		col  tcell.Color
		name string
		kind string
	}
	ps := []planet{
		{0.39, '•', pal.Rock, names.word(), "rock"}, {0.72, '•', pal.Sand, names.word(), "rock"},
		{1.0, '◉', pal.Accent, a.HomeName, "home"}, {1.52, '●', pal.Neg, names.word(), "rock"},
		{5.2, '◍', pal.Sand, names.word(), "giant"}, {9.5, '◍', pal.High, names.word(), "giant"},
		{19.2, '●', pal.Water, names.word(), "ice"}, {30, '●', pal.Water, names.word(), "ice"},
	}
	sun := Pt{0, 0}
	col := c.colonies(6)
	colonised := map[int]bool{}
	for _, k := range []int{3, 4, 1, 5, 6, 7}[:col] {
		colonised[k] = true
	}
	var home Pt
	pos := make([]Pt, len(ps))
	for i, pl := range ps {
		r := orbitR(pl.au)
		oc := pal.mix(pal.Paper, pal.Label, 0.6)
		if pl.kind == "home" {
			oc = pal.mix(pal.Paper, pal.Accent, 0.8)
		}
		// the ecliptic seen from above and a little to the side: ellipses
		n := int(2*math.Pi*r/c.v.S) + 8
		for k := 0; k < n; k++ {
			t := float64(k) / float64(n) * 2 * math.Pi
			if x, y, ok := c.cell(Pt{math.Cos(t) * r, math.Sin(t) * r * tilt}); ok {
				cv.Put(x, y, '·', oc, zGrid)
			}
		}
		ang := rnd01(a.W.Seed, i, 17)*2*math.Pi + float64(c.frame)*0.02/math.Pow(pl.au, 1.5)
		pos[i] = Pt{math.Cos(ang) * r, math.Sin(ang) * r * tilt}
		if pl.kind == "home" {
			home = pos[i]
		}
	}
	// the belt
	for k := 0; k < 160; k++ {
		r := orbitR(2.2 + rnd01(a.W.Seed, k, 3)*1.1)
		ang := rnd01(a.W.Seed, k, 4) * 2 * math.Pi
		if x, y, ok := c.cell(Pt{math.Cos(ang) * r, math.Sin(ang) * r * tilt}); ok {
			cv.Put(x, y, []rune{'·', '∴', '·', '∙'}[k%4], c.pal.mix(c.pal.Paper, pal.Rock, 0.7), zRelief)
		}
	}
	// lanes from home to each colony, goods moving on the traded ones
	k := 0
	for i := range ps {
		if colonised[i] {
			c.lane(home, pos[i], pal.mix(pal.Paper, pal.Accent, 0.7), float64(i)*0.3, k < len(a.Routes)+1)
			k++
		}
	}
	if x, y, ok := c.cell(sun); ok {
		cv.Stamp(x, y, []string{"╲ │ ╱", "─ ☼ ─", "╱ │ ╲"}, pal.High, zMark)
	}
	c.pois = append(c.pois, POI{Kind: "sun", Name: "the Sun", At: sun, Ref: -1})
	for i, pl := range ps {
		g, colr := pl.g, pl.col
		note := ""
		if colonised[i] {
			colr = pal.Accent
			note = "colony of your realm"
		}
		if pl.kind == "home" {
			note = c.earthNote()
		}
		c.mark(pos[i], g, colr, tcell.AttrBold, POI{Kind: "body", Name: pl.name}, body{Name: pl.name, Kind: pl.kind, Colony: colonised[i], Note: note})
		name := pl.name
		if colonised[i] {
			name += " ▪"
		}
		lc := pal.mix(pal.Paper, pal.Label, 0.85)
		if pl.kind == "home" || colonised[i] {
			lc = pal.Accent
		}
		c.label(pos[i], name, lc, 0)
		if pl.kind == "home" {
			// the moon, always colonised by now
			m := Pt{pos[i].X + 2.2, pos[i].Y - 1}
			if x, y, ok := c.cell(m); ok {
				cv.Put(x, y, '∘', pal.Accent, zMark)
			}
		}
	}
	// the stars beyond: known off-world civs at the edge, pointing their way
	for _, cv2 := range c.knownCivs(true) {
		ang := rnd01(hashStr(cv2.Key), 1, 1) * 2 * math.Pi
		edge := Pt{math.Cos(ang) * 60, math.Sin(ang) * 60}
		col := pal.Rel(relation(cv2))
		arrow := dirGlyph(math.Cos(ang), math.Sin(ang)/2, '→', '↓', '↗', '↘')
		if math.Cos(ang) < -0.38 {
			arrow = '←'
		}
		if math.Abs(math.Cos(ang)) <= 0.38 && math.Sin(ang) < 0 {
			arrow = '↑'
		}
		ly := 3 + rnd01(hashStr(cv2.Key), 2, 2)*8
		c.mark(edge, arrow, col, tcell.AttrBold, POI{Kind: "civ", Name: cv2.Def.Name}, body{Name: cv2.Def.Name, Kind: "civ", Civ: cv2, Note: fmt.Sprintf("%.1f light years out", ly)})
		c.label(edge, fmt.Sprintf("%s · %.1f ly", strings.ToUpper(cv2.Def.Name), ly), col, tcell.AttrBold)
	}
	c.omen(Pt{44, -30})
	c.legend = []line{
		{"☼  the Sun", pal.High, 0}, {"◉  " + a.HomeName + ", home", pal.Accent, 0},
		{"●  planet · ◍ giant", pal.Label, 0}, {"▪  colony (accent)", pal.Accent, 0},
		{"·  orbit · lane", pal.Faint, 0}, {"◆  cargo on a lane", pal.Accent, 0},
		{"∴  asteroid belt", pal.Rock, 0}, {"→  a people beyond", pal.Label, 0},
	}
	if len(a.Omens) > 0 {
		c.legend = append(c.legend, line{"☄  harbinger's omen", pal.Neg, 0})
	}
}

func (c *cosmicCtx) omen(at Pt) {
	for _, o := range c.a.Omens {
		if !o.Active {
			continue
		}
		col := c.pal.Neg
		if c.frame%2 == 1 {
			col = c.pal.Warn
		}
		for k := 1; k <= 6; k++ {
			if x, y, ok := c.cell(Pt{at.X + float64(k)*1.6*c.v.S, at.Y - float64(k)*0.9*c.v.S}); ok {
				c.cv.Put(x, y, '·', c.pal.mix(c.pal.Paper, col, 1-float64(k)*0.13), zMark)
			}
		}
		c.mark(at, c.p.Omen, col, tcell.AttrBold, POI{Kind: "omen", Name: o.Name}, body{Name: o.Name, Kind: "omen", Note: o.Target})
		s := o.Name
		if o.Numeric {
			s += fmt.Sprintf(" %.0f%%", o.Prob*100)
		}
		c.label(at, s, c.pal.Neg, tcell.AttrItalic)
	}
}

type star struct {
	At   Pt
	Name string
	Mag  float64
}

func (c *cosmicCtx) starList(n int, R float64, seed int64) []star {
	names := newNamer(seed)
	greek := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Tau", "Sigma", "Eta", "Kappa", "Lambda", "Omicron", "Zeta"}
	var out []star
	for k := 0; len(out) < n && k < n*20; k++ {
		r := math.Sqrt(rnd01(seed, k, 1)) * R
		ang := rnd01(seed, k, 2) * 2 * math.Pi
		p := Pt{math.Cos(ang) * r, math.Sin(ang) * r}
		ok := r > 3
		for _, s := range out {
			if s.At.Dist(p) < 3.2 {
				ok = false
			}
		}
		if !ok {
			continue
		}
		name := greek[int(rnd01(seed, k, 3)*float64(len(greek)))] + " " + names.word()
		out = append(out, star{p, name, rnd01(seed, k, 4)})
	}
	return out
}

func (c *cosmicCtx) stars() {
	a, pal, cv := c.a, c.pal, c.cv
	sol := Pt{0, 0}
	// polar graticule: distance rings and right-ascension spokes
	for _, r := range []float64{10, 20, 30} {
		if c.p.Mini {
			break
		}
		c.ring(sol, r, '·', pal.mix(pal.Paper, pal.Faint, 0.5), zGrid, 2)
		c.label(Pt{r * 0.72, -r * 0.72}, fmt.Sprintf("%.0f ly", r), pal.mix(pal.Paper, pal.Faint, 0.9), tcell.AttrItalic)
	}
	for h := 0; h < 12 && !c.p.Mini; h++ {
		ang := float64(h) / 12 * 2 * math.Pi
		if h%2 == 1 {
			c.label(Pt{math.Cos(ang) * 32.5, math.Sin(ang) * 32.5}, fmt.Sprintf("%dh", h*2), c.pal.mix(pal.Paper, pal.Faint, 0.9), 0)
			continue
		}
		for r := 4.0; r < 31; r += 2.6 {
			if x, y, ok := c.cell(Pt{math.Cos(ang) * r, math.Sin(ang) * r}); ok {
				cv.Put(x, y, '·', pal.mix(pal.Paper, pal.Faint, 0.35), zGrid)
			}
		}
		c.label(Pt{math.Cos(ang) * 32.5, math.Sin(ang) * 32.5}, fmt.Sprintf("%dh", h*2), pal.mix(pal.Paper, pal.Faint, 0.9), 0)
	}
	ss := c.starList(40, 31, a.W.Seed+333)
	// constellations: faint lines between the brightest neighbours
	sort.Slice(ss, func(i, j int) bool { return ss[i].Mag > ss[j].Mag })
	bright := ss[:12]
	for i, s := range bright {
		best, bd := -1, 11.0
		for j, t := range bright {
			if j != i && s.At.Dist(t.At) < bd {
				best, bd = j, s.At.Dist(t.At)
			}
		}
		if best > i {
			c.lane(s.At, bright[best].At, pal.mix(pal.Paper, pal.Faint, 0.4), 0, false)
		}
	}
	// colonies: the nearest systems, linked by hyperlanes
	byDist := append([]star(nil), ss...)
	sort.Slice(byDist, func(i, j int) bool { return byDist[i].At.Dist(sol) < byDist[j].At.Dist(sol) })
	ncol := c.colonies(9)
	col := map[string]bool{}
	for i := 0; i < ncol && i < len(byDist); i++ {
		col[byDist[i].Name] = true
		from := sol
		bd := byDist[i].At.Dist(sol)
		for j := 0; j < i; j++ { // attach to the nearest colony already held
			if d := byDist[j].At.Dist(byDist[i].At); d < bd {
				from, bd = byDist[j].At, d
			}
		}
		c.lane(from, byDist[i].At, pal.mix(pal.Paper, pal.Accent, 0.75), float64(i)*0.23, i < len(a.Routes)+1)
	}
	// off-world peoples: a cluster of systems and a sphere of influence
	for _, cv2 := range c.knownCivs(true) {
		h := hashStr(cv2.Key)
		ang := rnd01(h, 1, 1) * 2 * math.Pi
		d := 17 + rnd01(h, 2, 2)*9
		seat := Pt{math.Cos(ang) * d, math.Sin(ang) * d}
		rc := pal.Rel(relation(cv2))
		c.ring(seat, 4+float64(cv2.Def.Strength), '┄', pal.mix(pal.Paper, rc, 0.7), zBorder, 1)
		for k := 0; k < 3+cv2.Def.Strength; k++ {
			pp := Pt{seat.X + (rnd01(h, k, 5)-0.5)*7, seat.Y + (rnd01(h, k, 6)-0.5)*7}
			if x, y, ok := c.cell(pp); ok {
				cv.Put(x, y, '•', rc, zMark-1)
			}
		}
		c.mark(seat, c.p.CivCapital, rc, tcell.AttrBold, POI{Kind: "civ", Name: cv2.Def.Name}, body{Name: cv2.Def.Name, Kind: "civ", Civ: cv2, Note: fmt.Sprintf("%.1f light years out", d)})
		c.label(seat, strings.ToUpper(cv2.Def.Name), rc, tcell.AttrBold)
		if cv2.Info.AtWar {
			mid := Pt{seat.X * 0.55, seat.Y * 0.55}
			for k := -2; k <= 2; k++ {
				if x, y, ok := c.cell(Pt{mid.X - math.Sin(ang)*float64(k)*1.5, mid.Y + math.Cos(ang)*float64(k)*1.5}); ok {
					cv.PutA(x, y, '✕', pal.Neg, tcell.AttrBold, zMark)
				}
			}
		}
	}
	c.mark(sol, c.p.Capital, pal.Accent, tcell.AttrBold, POI{Kind: "body", Name: "Sol"}, body{Name: "Sol", Kind: "home", Note: c.earthNote()})
	c.label(sol, "SOL · "+a.HomeName, pal.Accent, tcell.AttrBold)
	for _, s := range ss {
		g, lc := '·', pal.mix(pal.Paper, pal.Label, 0.8)
		switch {
		case s.Mag > 0.85:
			g = '✶'
		case s.Mag > 0.6:
			g = '•'
		case s.Mag > 0.3:
			g = '∙'
		}
		colr := pal.mix(pal.Paper, pal.Ink, 0.5+s.Mag*0.5)
		note := ""
		if col[s.Name] {
			colr, lc, note = pal.Accent, pal.Accent, "colony of your realm"
			if g == '·' || g == '∙' {
				g = '•'
			}
		}
		c.mark(s.At, g, colr, 0, POI{Kind: "body", Name: s.Name}, body{Name: s.Name, Kind: "star", Colony: col[s.Name], Note: note})
		if s.Mag > 0.6 || col[s.Name] {
			c.label(s.At, s.Name, lc, 0)
		}
	}
	c.omen(Pt{-26, -14})
	c.legend = []line{
		{"☉  Sol, home", pal.Accent, 0}, {"•  colony (accent)", pal.Accent, 0},
		{"✶  bright star · ∙ dim", pal.Label, 0}, {"·  hyperlane · ◆ cargo", pal.Accent, 0},
		{"┄  sphere of influence", pal.Label, 0}, {"✶  seat of a people", pal.High, 0},
		{"◌  10 ly rings, RA spokes", pal.Faint, 0},
	}
}

func (c *cosmicCtx) galaxy() {
	a, pal, cv := c.a, c.pal, c.cv
	seed := a.W.Seed + 77
	arms := 4.0
	glow := [3]tcell.Color{pal.mix(pal.Paper, pal.Label, 0.07), pal.mix(pal.Paper, pal.High, 0.12), pal.mix(pal.Paper, pal.High, 0.24)}
	for j := 0; j < c.rc.H; j++ {
		for i := 0; i < c.rc.W; i++ {
			w := c.world(i, j)
			r := math.Hypot(w.X, w.Y)
			th := math.Atan2(w.Y, w.X)
			d := 1.0
			for k := 0.0; k < arms; k++ {
				ta := k*2*math.Pi/arms + 2.3*math.Log(math.Max(r, 1)/3)
				dd := math.Abs(math.Mod(th-ta+5*math.Pi, 2*math.Pi) - math.Pi)
				d = math.Min(d, dd)
			}
			den := math.Exp(-math.Pow(d*r/9, 2))*math.Exp(-r/30) + 1.4*math.Exp(-r*r/22)
			den *= 0.65 + 0.7*fbm(seed, w.X/4, w.Y/4, 3)
			if r > 46 {
				den *= 0.2
			}
			x, y := c.rc.X+i, c.rc.Y+j
			rr := rnd01(seed, int(math.Floor(w.X/c.v.S)), int(math.Floor(w.Y/(2*c.v.S))))
			// the glow under the stars: three classes, brightest at the core
			switch {
			case den > 1.0:
				cv.SetBg(x, y, glow[2])
			case den > 0.6:
				cv.SetBg(x, y, glow[1])
			case den > 0.32:
				cv.SetBg(x, y, glow[0])
			}
			switch {
			case den > 1.15 && rr < 0.7:
				cv.Put(x, y, '✶', pal.High, zRelief)
			case den > 0.7 && rr < 0.75:
				cv.Put(x, y, '•', pal.mix(pal.Paper, pal.High, 0.85), zRelief)
			case den > 0.4 && rr < 0.6:
				cv.Put(x, y, '∙', pal.mix(pal.Paper, pal.Ink, 0.8), zRelief)
			case den > 0.18 && rr < 0.45:
				cv.Put(x, y, '·', pal.mix(pal.Paper, pal.Label, 0.75), zRelief)
			}
		}
	}
	// sectors: the realm's grows by age; the known peoples hold theirs
	solR, solA := 27.0, rnd01(seed, 1, 1)*2*math.Pi
	sol := Pt{math.Cos(solA) * solR, math.Sin(solA) * solR}
	realm := 5 + float64(a.AgeIdx-19)*4 + math.Log2(1+float64(a.Buildings)/400)
	for j := 0; j < c.rc.H; j++ {
		for i := 0; i < c.rc.W; i++ {
			w := c.world(i, j)
			if w.Dist(sol) < realm {
				cv.SetBg(c.rc.X+i, c.rc.Y+j, pal.mix(pal.Paper, pal.Accent, 0.13))
			}
		}
	}
	c.ring(sol, realm, '┈', pal.Accent, zBorder, 1)
	seats := map[string]Pt{
		"stellar_federation": {math.Cos(solA+0.45) * 31, math.Sin(solA+0.45) * 31},
		"void_reavers":       {math.Cos(solA-1.9) * 40, math.Sin(solA-1.9) * 40},
		"quantum_collective": {math.Cos(solA+2.6) * 9, math.Sin(solA+2.6) * 9},
	}
	c.mark(sol, c.p.Capital, pal.Accent, tcell.AttrBold, POI{Kind: "body", Name: "Sol"}, body{Name: "Sol", Kind: "home", Note: c.earthNote()})
	c.label(sol, "SOL · YOUR REALM", pal.Accent, tcell.AttrBold)
	for k, cv2 := range c.knownCivs(true) {
		seat := seats[cv2.Key]
		rc := pal.Rel(relation(cv2))
		rad := 3 + float64(cv2.Def.Strength)*1.3
		for j := 0; j < c.rc.H; j++ {
			for i := 0; i < c.rc.W; i++ {
				if c.world(i, j).Dist(seat) < rad {
					cv.SetBg(c.rc.X+i, c.rc.Y+j, pal.mix(pal.Paper, rc, 0.12))
				}
			}
		}
		c.ring(seat, rad, '┄', rc, zBorder, 1)
		c.lane(sol, seat, pal.mix(pal.Paper, rc, 0.8), float64(k)*0.31, !cv2.Info.AtWar && cv2.Info.Status != "embargo")
		c.mark(seat, c.p.CivCapital, rc, tcell.AttrBold, POI{Kind: "civ", Name: cv2.Def.Name}, body{Name: cv2.Def.Name, Kind: "civ", Civ: cv2,
			Note: fmt.Sprintf("%.1f kly from Sol", seat.Dist(sol))})
		c.label(seat, strings.ToUpper(cv2.Def.Name), rc, tcell.AttrBold)
		if cv2.Info.AtWar {
			dx, dy := seat.X-sol.X, seat.Y-sol.Y
			l := math.Hypot(dx, dy)
			mid := Pt{sol.X + dx/l*(realm+1.5), sol.Y + dy/l*(realm+1.5)}
			for q := -3; q <= 3; q++ {
				if x, y, ok := c.cell(Pt{mid.X - dy/l*float64(q)*1.2, mid.Y + dx/l*float64(q)*1.2}); ok {
					cv.PutA(x, y, '✕', pal.Neg, tcell.AttrBold, zMark)
				}
			}
		}
	}
	// the colonies scattered in the realm
	for k := 0; k < c.colonies(14)+3; k++ {
		r := math.Sqrt(rnd01(seed, k, 21)) * realm * 0.9
		ang := rnd01(seed, k, 22) * 2 * math.Pi
		p := Pt{sol.X + math.Cos(ang)*r, sol.Y + math.Sin(ang)*r}
		if x, y, ok := c.cell(p); ok {
			cv.Put(x, y, '•', pal.Accent, zMark-1)
		}
	}
	c.pois = append(c.pois, POI{Kind: "body", Name: "the Core", At: Pt{0, 0}, Ref: -1})
	names := newNamer(seed + 5)
	for k := 0; k < 4 && !c.p.Mini; k++ {
		r := 38.0
		ang := float64(k)*2*math.Pi/arms + 2.3*math.Log(r/3)
		c.label(Pt{math.Cos(ang) * r, math.Sin(ang) * r}, spaced(names.word()+" arm"), pal.mix(pal.Paper, pal.Label, 0.75), tcell.AttrItalic)
	}
	c.omen(Pt{3, -4})
	c.legend = []line{
		{"☉  Sol, heart of your realm", pal.Accent, 0}, {"┈  your sector's edge", pal.Accent, 0},
		{"•  your colonies", pal.Accent, 0}, {"✶  seat of a people", pal.High, 0},
		{"┄  their sector", pal.Label, 0}, {"◆  cargo on a hyperlane", pal.Accent, 0},
		{"✕  war front", pal.Neg, 0}, {"▓  core · ∙ arm stars", pal.High, 0},
	}
}

func (c *cosmicCtx) readout(x, y int) string {
	w := c.world(x, y)
	switch c.p.Model {
	case "orrery":
		r := math.Hypot(w.X, w.Y) / 10
		return fmt.Sprintf("r %.2f AU from the Sun", r*r)
	case "stars":
		ra := math.Mod(math.Atan2(w.Y, w.X)/(2*math.Pi)*24+24, 24)
		return fmt.Sprintf("RA %dh%02dm · %.1f ly from Sol", int(ra), int((ra-math.Floor(ra))*60), math.Hypot(w.X, w.Y))
	}
	l := math.Mod(math.Atan2(w.Y, w.X)*180/math.Pi+360, 360)
	return fmt.Sprintf("l %03.0f° · %.1f kly from the core", l, math.Hypot(w.X, w.Y))
}

func (c *cosmicCtx) inspect(p POI, pal *Pal) []line {
	b, ok := c.bodies[p.Ref]
	var L []line
	add := func(s string, col tcell.Color, at tcell.AttrMask) { L = append(L, line{s, col, at}) }
	if !ok {
		add(strings.ToUpper(p.Name), pal.High, tcell.AttrBold)
		return L
	}
	switch {
	case b.Civ != nil:
		f := b.Civ.Info
		add(strings.ToUpper(b.Name), pal.Rel(relation(b.Civ)), tcell.AttrBold)
		add(b.Note, pal.Faint, 0)
		add(fmt.Sprintf("%s · %s", f.Personality, stars(b.Civ.Def.Strength)), pal.Label, 0)
		add(fmt.Sprintf("standing %+d %s %s", f.Opinion, bar(f.Opinion), f.Status), pal.Rel(relation(b.Civ)), 0)
		add(fmt.Sprintf("trades %s · +%.0f%% allied", b.Civ.Def.Specialty, b.Civ.Def.TradeBonus*100), pal.Ink, 0)
		if f.AtWar {
			add("AT WAR", pal.Neg, tcell.AttrBold)
			add("» diplomacy tribute "+b.Civ.Key, pal.Accent, 0)
		} else {
			add("» diplomacy deals "+b.Civ.Key, pal.Accent, 0)
		}
	case b.Kind == "omen":
		add(strings.ToUpper(b.Name), pal.Neg, tcell.AttrBold)
		add("a harbinger speaks of "+b.Note, pal.Ink, tcell.AttrItalic)
		add("» harbinger brace", pal.Accent, 0)
	default:
		add(strings.ToUpper(b.Name), pal.Ink, tcell.AttrBold)
		add(b.Kind, pal.Faint, 0)
		if b.Note != "" {
			add(trunc(b.Note, 60), pal.Label, 0)
		}
		if b.Colony {
			add("» build (colony works)", pal.Accent, 0)
		}
	}
	return L
}

// orbitR spaces the orbits logarithmically, so the inner planets are not
// crushed against the Sun and Neptune still fits.
func orbitR(au float64) float64 { return 7 + 14.5*math.Log(1+au) }
