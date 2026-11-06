package mapgen

// The two wide views of the Galactic map. OVERVIEW is a sector of the Orion
// Arm round Sol, its nebulae and star clusters set by their real direction
// with the distance compressed (radius grows with the square root of the
// distance); COSMOS is the void between galaxies, with the Milky Way, its
// neighbours and a scatter of far ones. Both are drawn once per window into a
// canvas of light channels and kept; a frame copies them and twinkles a few
// stars.

// SkyObject is a named thing of the wide views.
type SkyObject struct {
	Key string
	Ly  int // light-years from Sol; 0 for the Milky Way, which has none
	lon int // galactic longitude, degrees (arm objects)
}

// SkyObjects are the selectable objects: the arm sector's first, then the
// void's. A map thing's Ref is the index plus one.
var SkyObjects = [...]SkyObject{
	{"orion_nebula", 1344, 209}, {"horsehead_nebula", 1375, 206}, {"pleiades", 444, 167}, {"hyades", 153, 180},
	{"crab_nebula", 6500, 185}, {"eagle_nebula", 7000, 17}, {"lagoon_nebula", 4100, 6}, {"carina_nebula", 8500, 287},
	{"milky_way", 0, 0}, {"large_magellanic_cloud", 160000, 0}, {"small_magellanic_cloud", 200000, 0},
	{"andromeda", 2500000, 0}, {"triangulum", 2700000, 0},
}

const (
	armObjects = 8
	armReachLy = 8500
	armBandDeg = 20 // the arm's direction on the screen
	armSkew    = 6  // degrees the Horsehead is drawn aside so it does not sit on the Orion Nebula
)

// light channels, then the dust that darkens them
const (
	cCool = iota
	cWarm
	cPink
	cRed
	cGold
	cTeal
	cOrange
	cBlue
	nch
	cDust = nch
)

var chHue = [nch]Hue{HStar, HSun, HNeon, HRed, HGold, HHydro, HFire, HAir}

// wcanvas is a window of light, in pixels; ell and the rest take isotropic
// positions, where a pixel is one wide and two tall.
type wcanvas struct {
	w, h int
	ch   [][nch + 1]int16
	pt   []Paint
	tw   []int // pixels of the brighter stars
}

func newWcanvas(w, h int) *wcanvas {
	return &wcanvas{w: w, h: h, ch: make([][nch + 1]int16, w*h), pt: make([]Paint, w*h)}
}

// ell raises channel k by amp in an ellipse of radii rx, ry tilted phi.
func (c *wcanvas) ell(cx, cy, rx, ry, phi, k, amp int) {
	rx, ry = max(rx, 1), max(ry, 1)
	r := max(rx, ry)
	cs, sn := cos4096(phi), sin4096(phi)
	for y := max(0, floorDiv(cy-r, 2)); y <= min(c.h-1, (cy+r)/2+1); y++ {
		for x := max(0, cx-r); x <= min(c.w-1, cx+r); x++ {
			dx, dy := x-cx, 2*y-cy
			u, v := (dx*cs+dy*sn)/1024, (dy*cs-dx*sn)/1024
			if b := bump(sq(u*256/rx)+sq(v*256/ry), 65536); b > 0 {
				c.ch[y*c.w+x][k] += int16(amp * b / 256)
			}
		}
	}
}

// star sets pixel (x, y)'s own paint; a bright one may twinkle.
func (c *wcanvas) star(x, y int, p Paint) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	c.pt[y*c.w+x] = p
	if p.Weight >= 200 && len(c.tw) < 24 {
		c.tw = append(c.tw, y*c.w+x)
	}
}

// starAt is star at an isotropic position.
func (c *wcanvas) starAt(x, y int, p Paint) { c.star(x, floorDiv(y, 2), p) }

// glint is a star of a seeded colour at an isotropic position.
func (c *wcanvas) glint(seed int64, i, x, y, lo int) {
	hue := [6]Hue{HStar, HStar, HWhite, HWhite, HStar, HSun}[roll(6, seed, 7010, int64(i))]
	r := roll(256-lo, seed, 7011, int64(i))
	if lo < 100 {
		r = r * r / (256 - lo) // a low floor means mostly dim stars
	}
	c.starAt(x, y, P(hue, lo+r))
}

// paints turns the channels into paints in pix.
func (c *wcanvas) paints(pix *Pixels) {
	for i := range c.ch {
		row := &c.ch[i]
		best, total := 0, 0
		for k := 0; k < nch; k++ {
			total += int(row[k])
			if row[k] > row[best] {
				best = k
			}
		}
		w := total
		if d := int(row[cDust]); d > 0 {
			w = w * (256 - min(d, 235)) / 256
		}
		p := Paint{}
		if w >= 8 {
			p = P(chHue[best], w)
		}
		if c.pt[i].Weight > 0 {
			p = c.pt[i]
		}
		pix.Set(i%c.w, i/c.w, p, 0)
	}
}

// galLook is how a galaxy is drawn.
type galLook struct {
	r, q, phi int // radius, the squash of its tilt out of 256, its angle
	arms      int
	spin      int // turns of the arms, in 256ths
	core, bar int
	dust      int
	stars     int
	spurs     bool // two fainter, shorter arms between the main ones
}

// armXY is the point at t (0..255 along it) of the arm that leaves at angle
// off; with a bar the arms leave from its ends.
func armXY(g galLook, cx, cy, off, t int) (int, int, int) {
	lo := 30
	if g.bar > 0 {
		off += turn / 7
		lo = 97
	}
	th := off + t*g.spin*turn/65536
	r := g.r * (lo + (256-lo)*t*t/65536) / 256 // wide pitch outside, like a log spiral
	u, v := r*cos4096(th)/1024, r*sin4096(th)/1024*g.q/256
	cs, sn := cos4096(g.phi), sin4096(g.phi)
	return cx + (u*cs-v*sn)/1024, cy + (u*sn+v*cs)/1024, r
}

// galaxy paints a spiral at (cx, cy): a soft glow over the disc, arms of pale
// light and stars that widen as they wind out, dust on their inner edges, pink
// knots, a warm bar and core. The fainter spurs lie between the main arms.
func (c *wcanvas) galaxy(seed int64, id, cx, cy int, g galLook) {
	sq0 := max(g.q, 40)
	c.ell(cx, cy, g.r, g.r*sq0/256, g.phi, cCool, 9)
	c.ell(cx, cy, g.r*7/10, g.r*7/10*sq0/256, g.phi, cWarm, 12)
	n := max(60, g.r*3)
	for a := 0; a < g.arms*(1+boolInt(g.spurs)); a++ {
		spur := a >= g.arms
		off, lo, amp := a*turn/g.arms, 0, 130
		if spur {
			off, lo, amp = (a-g.arms)*turn/g.arms+turn/4, 90, 60
		}
		for i := lo * n / 256; i < (256-(lo+20)*boolInt(spur))*n/256; i++ {
			t := i * 256 / n
			x, y, r := armXY(g, cx, cy, off, t)
			wd := 1 + g.r/22 + g.r*t/256/8
			c.ell(x, y, wd, max(1, wd*sq0/256), 0, cCool, amp*(60+t/2)/256*7/(wd+2)) // wider arms overlap more, so each sample adds less
			for k := 0; k < g.stars; k++ {
				j := int64(id*100000 + a*10000 + i*10 + k)
				ox, oy := roll(2*wd+1, seed, 7020, j)-wd, roll(2*wd+1, seed, 7021, j)-wd
				c.starAt(x+ox, y+oy*sq0/256, P([3]Hue{HStar, HWhite, HStar}[roll(3, seed, 7023, j)], 50+roll(110, seed, 7024, j)))
			}
			if spur {
				continue
			}
			if g.dust > 0 && i%3 == 0 { // the dust lies along the inner edge of the arm
				ex, ey, _ := armXY(g, cx, cy, off-turn/40, t)
				c.ell(ex, ey, max(1, wd/2), max(1, wd/2*sq0/256), 0, cDust, g.dust*(40+t/3)/256)
			}
			if t > 70 && t < 230 && roll(30, seed, 7022, int64(id), int64(a), int64(i)) == 0 {
				c.ell(x, y, 1+r/18, 1+r/18*sq0/256, 0, cPink, 110)
			}
		}
	}
	if g.bar > 0 {
		c.ell(cx, cy, g.r*38/100, g.r*10/100, g.phi+turn/7, cWarm, g.bar)
	}
	c.ell(cx, cy, 2+g.r/5, 1+(2+g.r/5)*sq0/256, g.phi, cWarm, g.core)
	c.ell(cx, cy, 1+g.r/12, 1+(1+g.r/12)*sq0/256, g.phi, cWarm, g.core)
	c.starAt(cx, cy, P(HWhite, 255))
}

// ring dots a ring at radius r about (cx, cy).
func (c *wcanvas) ring(cx, cy, r, n int, p Paint) {
	for k := 0; k < n; k++ {
		a := k * turn / n
		c.starAt(cx+r*cos4096(a)/1024, cy+r*sin4096(a)/1024, p)
	}
}

// nebula paints the object key as a soft glow of several hues at (x, y).
func (c *wcanvas) nebula(seed int64, key string, x, y, r int) {
	st := func(n, spread int, lo int) {
		for i := 0; i < n; i++ {
			c.glint(seed, 100+i+x, x+roll(2*spread+1, seed, 7030, int64(x), int64(i))-spread, y+2*(roll(2*spread+1, seed, 7031, int64(x), int64(i))-spread)/2, lo)
		}
	}
	switch key {
	case "orion_nebula":
		c.ell(x, y, 2*r, 3*r/2, turn/9, cRed, 60)
		c.ell(x, y, r, 3*r/4, turn/9, cPink, 120)
		c.ell(x, y, r/2, r/2, 0, cTeal, 150)
		st(5, r, 200)
	case "horsehead_nebula":
		c.ell(x, y, 2*r, 3*r/2, 0, cRed, 55)
		c.ell(x-r/3, y+r/2, r/2, r, turn/14, cDust, 235)
	case "pleiades":
		c.ell(x, y, 2*r, r, 0, cBlue, 90)
		st(7, r, 215)
	case "hyades":
		for i := 0; i < 6; i++ {
			c.starAt(x+(i-3)*r/2, y+abs(i-3)*r, P([3]Hue{HSun, HWhite, HAmber}[i%3], 200+roll(56, seed, 7040, int64(i))))
		}
	case "crab_nebula":
		c.ell(x, y, 3*r/2, r, turn/10, cOrange, 90)
		c.ell(x, y, r*3/4, r/2, turn/10, cBlue, 90)
		for i := 0; i < 8; i++ {
			c.ell(x, y, r, 1, roll(turn, seed, 7041, int64(i)), cOrange, 130)
		}
		c.starAt(x, y, P(HWhite, 255))
	case "eagle_nebula":
		c.ell(x, y, 2*r, 3*r/2, 0, cGold, 80)
		c.ell(x-r/3, y, r, r, 0, cTeal, 90)
		for i := -1; i <= 1; i++ {
			c.ell(x+i*r/2, y+r/3, 1+r/8, r, 0, cDust, 220)
		}
		st(4, r, 190)
	case "lagoon_nebula":
		c.ell(x, y, 2*r, 3*r/2, 0, cRed, 70)
		c.ell(x, y, r, r*3/4, 0, cPink, 100)
		c.ell(x+r/3, y, r/3, r/3, 0, cGold, 70)
		c.ell(x, y, 2*r, 1, turn/16, cDust, 210)
		st(4, r, 190)
	case "carina_nebula":
		c.ell(x, y, 5*r/2, 3*r/2, turn/12, cRed, 100)
		c.ell(x, y, 3*r/2, r, turn/12, cRed, 90)
		c.ell(x, y, r/3, r/3, 0, cPink, 150)
		c.ell(x-r/2, y, 2*r, 1, turn/7, cDust, 220)
		c.ell(x+r/2, y+r/3, 2*r, 1, turn/9, cDust, 200)
		st(6, r, 200)
	}
}

// wideKey is everything a wide layer is drawn from: the whole window.
type wideKey struct {
	seed             int64
	void             bool
	w, h, x0, y0, st int
	colonies         int
}

// twinkle is a bright star's cell and how it looks between its flashes.
type twinkle struct {
	ci  int
	off Cell
}

// wideLayer keeps the drawn view between frames.
type wideLayer struct {
	key   wideKey
	ok    bool
	cells []Cell
	hit   []uint8 // per cell: 0 nothing, 255 Sol, else the object's index plus one
	tw    []twinkle
}

// OverviewArm draws the Orion Arm sector round Sol into w by h cells.
func (s *Scene) OverviewArm(w, h int) *Overview { return s.wideView(w, h, false) }

// OverviewVoid draws the void between galaxies into w by h cells.
func (s *Scene) OverviewVoid(w, h int) *Overview { return s.wideView(w, h, true) }

func (s *Scene) wideView(w, h int, void bool) *Overview {
	gw, gh := s.Grid.W, s.Grid.H
	o := &Overview{Grid: newGrid(w, h), Scale: 1, step: 256, picks: make([]Pt, w*h), X0: gw/2 - w/2, Y0: gh/2 - h/2}
	o.Grid.Wall, o.Grid.Tint, o.Grid.Space = s.Grid.Wall, s.Grid.Tint, true
	l := s.wide
	if l == nil {
		l = &wideLayer{}
	}
	if k := (wideKey{s.Seed, void, w, h, o.X0, o.Y0, o.step, len(s.Colonies)}); !l.ok || l.key != k {
		*l = *buildWide(k)
	}
	copy(o.Grid.Cells, l.cells)
	var idx [256]uint16
	for ci, k := range l.hit {
		if k == 0 {
			continue
		}
		if idx[k] == 0 {
			t := sunThing
			if k != 255 {
				t = Thing{Kind: KindTerrain, Key: "deepsky", Ref: int(k)}
			}
			idx[k] = o.Grid.thingIndex(t)
		}
		o.Grid.Hit[ci] = idx[k]
	}
	for i, t := range l.tw {
		if !blinkWith(s.Seed, s.Anim, 800+i, 40, 30) {
			o.Grid.Cells[t.ci] = t.off
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o.picks[y*w+x] = Pt{o.X0 + x, o.Y0 + y}
		}
	}
	return o
}

// buildWide draws the layer for k.
func buildWide(k wideKey) *wideLayer {
	c := newWcanvas(2*k.w, 2*k.h)
	hits := make([]uint8, k.w*k.h)
	mark := func(n uint8, x, y, r int) { // cells whose middle is within r of the isotropic point
		for cy := max(0, floorDiv(y-r, 4)); cy <= min(k.h-1, (y+r)/4); cy++ {
			for cx := max(0, floorDiv(x-r, 2)); cx <= min(k.w-1, (x+r)/2); cx++ {
				if sq(2*cx+1-x)+sq(4*cy+2-y) <= r*r {
					hits[cy*k.w+cx] = n
				}
			}
		}
	}
	if k.void {
		c.void(k.seed, mark)
	} else {
		c.arm(k.seed, k.colonies, mark)
	}
	pix := NewPixels(k.w, k.h)
	defer pix.Release()
	c.paints(pix)
	g := newGrid(k.w, k.h)
	for i := range g.Cells {
		g.Cells[i] = Cell{R: ' ', Bg: P(HVoid, 255)}
	}
	pix.Resolve(g)
	for i, cell := range g.Cells {
		if f, b := int(cell.Fg.Weight), int(cell.Bg.Weight); cell.R != ' ' && (f < 64 && b < 64 || cell.Fg.Hue == cell.Bg.Hue && abs(f-b) < 24) {
			// a haze glyph a themed ground cannot tell from its ground is a blank
			g.Cells[i] = Cell{R: ' ', Fg: cell.Bg, Bg: cell.Bg}
			if cell.Fg.Weight > cell.Bg.Weight {
				g.Cells[i].Bg = cell.Fg
			}
		}
	}
	l := &wideLayer{key: k, ok: true, cells: g.Cells, hit: hits}
	for _, p := range c.tw {
		ci := p/c.w/2*k.w + p%c.w/2
		if cell := g.Cells[ci]; cell.R != ' ' && cell.Fg.Hue != HNone {
			cell.R, cell.Fg.Weight = '·', cell.Fg.Weight/3 // dimmed to a dot
			l.tw = append(l.tw, twinkle{ci, cell})
		}
	}
	return l
}

// arm paints the Orion Arm sector.
func (c *wcanvas) arm(seed int64, colonies int, mark func(n uint8, x, y, r int)) {
	cx, cy := 2*(c.w/4)+1, 2*(2*(c.h/4)+1) // Sol, in pixels and isotropic rows
	reach := min(c.w/2, c.h) * 90 / 100
	phi := armBandDeg * turn / 360
	cs, sn := cos4096(phi), sin4096(phi)
	wd := reach * 55 / 100
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			dx, dy := x-cx, 2*y-cy
			u, v := (dx*cs-dy*sn)/1024, (dx*sn+dy*cs)/1024
			v += reach / 12 * sin4096(u*turn/(reach*2)) / 1024
			d := 20 + bump(v*v, wd*wd)*200/256
			i := y*c.w + x
			n := fbm(seed+3, x*256/9, y*256/5, 3) // 0..1023
			c.ch[i][cCool] = int16(d * n / 1024 * 5 / 10)
			if r := abs(fbm(seed+4, x*256/12, y*256/6, 3) - 512); r < 70 { // a lane where the noise crosses its middle
				c.ch[i][cDust] = int16((70 - r) * 3 * d / 256)
			}
			if roll(1000, seed, 7050, int64(x), int64(y)) < 3+d*28/256 {
				c.glint(seed, i, x, 2*y, 40)
			}
		}
	}
	for i := 0; i < 11; i++ { // pink knots along the arm
		u := (roll(2*reach, seed, 7051, int64(i)) - reach) * 11 / 10
		v := (roll(2*wd, seed, 7052, int64(i)) - wd) / 2
		c.ell(cx+(u*cs+v*sn)/1024, cy+(-u*sn+v*cs)/1024, 2+roll(3, seed, 7053, int64(i)), 2, 0, cPink, 90)
	}
	rs := reach * (10 + min(colonies, 17)) / 100 // the people's territory, a warm tint
	c.ell(cx, cy, rs, rs, 0, cGold, 24)
	br := reach * 9 / 100
	c.ell(cx, cy, br, br, 0, cCool, 16)
	c.ring(cx, cy, br, 48, P(HOrbit, 70))
	for i, o := range SkyObjects[:armObjects] {
		a := (o.lon - 60 + armSkew*boolInt(o.Key == "horsehead_nebula")) * turn / 360
		r := reach * isqrt(o.Ly*10000/armReachLy) / 100
		x, y := cx+r*cos4096(a)/1024, cy-r*sin4096(a)/1024
		rad := [armObjects]int{7, 4, 4, 4, 5, 7, 7, 10}[i]
		c.nebula(seed, o.Key, x, y, rad)
		mark(uint8(i+1), x, y, max(rad, 5))
	}
	for dx := -2; dx <= 2; dx++ {
		c.star(cx+dx, cy/2, P([5]Hue{HSun, HWhite, HWhite, HWhite, HSun}[dx+2], 245))
	}
	c.star(cx, cy/2-1, P(HSun, 210))
	c.star(cx, cy/2+1, P(HSun, 210))
	mark(255, cx, cy, 5)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// void paints the space between galaxies.
func (c *wcanvas) void(seed int64, mark func(n uint8, x, y, r int)) {
	W, H := c.w, 2*c.h // isotropic size
	mwX, mwY, mwR := W*26/100, H*46/100, min(W*23/100, H*36/100)
	m31X, m31Y, m31R := W*79/100, H*30/100, W*14/100
	m33X, m33Y, m33R := W*88/100, H*72/100, W*6/100
	lmcX, lmcY := W*46/100, H*80/100
	smcX, smcY := W*56/100, H*88/100
	type major struct{ x, y, r int }
	majors := []major{{mwX, mwY, mwR}, {m31X, m31Y, m31R}, {m33X, m33Y, m33R}, {lmcX, lmcY, 10}, {smcX, smcY, 8}}

	mw := galLook{r: mwR, q: 256, phi: 0, arms: 2, spin: 330, core: 150, bar: 130, dust: 200, stars: 1, spurs: true}
	c.galaxy(seed, 1, mwX, mwY, mw)
	sx, sy, _ := armXY(mw, mwX, mwY, 0, 170) // Sol, in the Orion spur
	c.ring(sx, sy, 3, 12, P(HOrbit, 90))
	c.starAt(sx, sy, P(HSun, 255))
	c.starAt(sx+1, sy, P(HWhite, 255))
	mark(9, mwX, mwY, mwR*9/10)
	mark(255, sx, sy, 4)

	m31 := galLook{r: m31R, q: 85, phi: -turn / 12, arms: 2, spin: 256, core: 170, dust: 230, stars: 1}
	c.galaxy(seed, 2, m31X, m31Y, m31)
	c.ell(m31X, m31Y, m31R*11/10, m31R*11/10*85/256, -turn/12, cWarm, 40)
	mark(12, m31X, m31Y, m31R*7/10)
	m33 := galLook{r: m33R, q: 256, phi: 0, arms: 2, spin: 200, core: 90, dust: 0, stars: 1}
	c.galaxy(seed, 3, m33X, m33Y, m33)
	mark(13, m33X, m33Y, m33R*8/10)

	for i, p := range [2]struct{ x, y, rx, ry int }{{lmcX, lmcY, 9, 6}, {smcX, smcY, 6, 4}} { // the Clouds, irregular and blue-white
		for k := 0; k < 4; k++ {
			j := int64(i*10 + k)
			c.ell(p.x+roll(p.rx, seed, 7060, j)-p.rx/2, p.y+roll(p.ry, seed, 7061, j)-p.ry/2, p.rx*2/3, p.ry*2/3, roll(turn, seed, 7062, j), cCool, 60)
		}
		for k := 0; k < 14; k++ {
			j := int64(i*100 + k)
			c.glint(seed, int(j), p.x+roll(2*p.rx, seed, 7063, j)-p.rx, p.y+roll(2*p.ry, seed, 7064, j)-p.ry, 120)
		}
		mark(uint8(10+i), p.x, p.y, p.rx+2)
	}

	// the lane from the Milky Way toward Andromeda, ending short of it
	x0, y0, x1, y1 := mwX+mwR+4, mwY-mwR/4, m31X-m31R-8, m31Y+m31R/8
	n := max(abs(x1-x0), abs(y1-y0)/2)
	for k := 0; k < n; k++ {
		c.starAt(x0+(x1-x0)*k/n, y0+(y1-y0)*k/n, P(HStar, 110))
	}

	for i := 0; i < 24 && len(majors) < 16; i++ { // far galaxies, faint
		x, y := roll(W, seed, 7070, int64(i)), roll(H, seed, 7071, int64(i))
		free := true
		for _, m := range majors {
			if sq(x-m.x)+sq(y-m.y) < sq(m.r+14) {
				free = false
			}
		}
		if !free {
			continue
		}
		majors = append(majors, major{x, y, 6})
		ph, amp := roll(turn/2, seed, 7072, int64(i)), 45+roll(50, seed, 7073, int64(i))
		switch i % 3 {
		case 0: // edge-on
			c.ell(x, y, 8, 1, ph, cWarm, amp+30)
			c.ell(x, y, 5, 1, ph, cCool, amp)
		case 1: // a small spiral
			c.ell(x, y, 6, 4, ph, cCool, amp)
			c.ell(x, y, 2, 1, ph, cWarm, amp+60)
		default: // an elliptical
			c.ell(x, y, 5, 4, ph, cWarm, amp)
		}
		c.starAt(x, y, P(HWhite, 120+roll(60, seed, 7074, int64(i))))
	}
	for y := 0; y < c.h; y++ { // a very few foreground stars
		for x := 0; x < c.w; x++ {
			if roll(2400, seed, 7080, int64(x), int64(y)) == 0 {
				c.glint(seed, y*c.w+x, x, 2*y, 110)
			}
		}
	}
}
