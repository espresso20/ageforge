package mapgen

// The cosmic view of Sol (SystemViewCosmos): what lies round the planets. Back
// to front: the glow and stars of the Milky Way, the Oort cloud, the sky's own
// stars, the neighbor stars, the bubble of the Sun, then (orrery.go) the
// Kuiper belt, the orbits and the bodies, and last the relay beacons. All of
// it is placed from the seed and still, except slow twinkles and blinks.
//
// Distances are in rows from the Sun, the way the orbits are (a row is 2.3
// columns across); the pixel canvas is half a column by half a row a pixel.

const (
	kuiperIn, kuiperOut = 100, 166 // the belt's reach, in percent of maxRY
	bubbleAt            = 190      // the bubble's edge
	oortShell           = 260      // the thickest part of the Oort cloud
	oortOut             = 440      // and where it has all but gone
	beaconPeriod        = 80       // a relay beacon: lit beaconLit frames in beaconPeriod
	beaconLit           = 10
)

// fsin is sin(x) for x in radians.
func fsin(x float64) float64 { return float64(sin4096(int(x*turn/6.2831853))) / 1024 }

// toPx is the pixel (x, y) at e rows from the Sun in direction a (4096ths of a turn).
func (r *orrery) toPx(e float64, a int) (int, int) {
	f := e * 32 // 16ths of a pixel
	return (r.sx + int(f*2.3*float64(cos4096(a))/1024)) / pxUnit, (r.sy + int(f*float64(sin4096(a))/1024)) / pxUnit
}

// cosmos paints everything behind the Kuiper belt.
func (r *orrery) cosmos(o *Overview, maxRY int) {
	r.milkyWay(o)
	r.oort(maxRY)
	r.fieldStars(o)
	r.neighbors()
	r.bubble(maxRY)
}

// band is the Milky Way: a soft diagonal river across the sky.
type band struct {
	dx, dy, nx, ny float64 // along it and across it (unit vectors, on the screen)
	ox, oy         float64 // where it passes the middle
	sigma, phase   float64
}

func (r *orrery) newBand() band {
	H := 2 * float64(r.pix.H)          // the screen's height in pixel widths
	th := 300 + roll(500, r.seed, 401) // 4096ths of a turn
	if roll(2, r.seed, 402) == 0 {
		th = -th
	}
	if roll(2, r.seed, 403) == 0 {
		th += turn / 2 // which end it runs from
	}
	b := band{dx: float64(cos4096(th)) / 1024, dy: float64(sin4096(th)) / 1024, sigma: 0.14 * H, phase: float64(roll(628, r.seed, 404)) / 100}
	b.nx, b.ny = -b.dy, b.dx
	off := (0.08 + float64(roll(8, r.seed, 405))/100) * H
	if roll(2, r.seed, 406) == 0 {
		off = -off
	}
	b.ox, b.oy = b.nx*off, b.ny*off
	return b
}

// at is the strength of the glow (0 to 1), how open the dust lane leaves it
// (0 to 1) and a mottle (about 0.5 to 1.5) at pixel (x, y).
func (b *band) at(seed int64, r *orrery, x, y int) (glow, lane, mottle float64) {
	X := float64(x) - float64(r.sx)/pxUnit - b.ox
	Y := 2*(float64(y)-float64(r.sy)/pxUnit) - b.oy
	d, t := X*b.nx+Y*b.ny, X*b.dx+Y*b.dy
	d += 0.2 * b.sigma * fsin(t/(4*b.sigma)+b.phase)
	u := d / b.sigma
	g := 1 / (1 + u*u)
	glow = g * g
	dl := (d - 0.12*b.sigma + 0.05*b.sigma*fsin(t/(2*b.sigma)+2*b.phase)) / (0.17 * b.sigma)
	lane = 1 - 0.88/(1+dl*dl)
	if glow >= 0.02 {
		mottle = 0.5 + float64(fbm(seed+410, int(t*256/(0.9*b.sigma)), int(d*256/(0.9*b.sigma)), 3))/noiseMax
	}
	return
}

// milkyWay lays the band's glow into the ground of the cells and its stars
// into the pixels.
func (r *orrery) milkyWay(o *Overview) {
	b := r.newBand()
	g := o.Grid
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			glow, lane, m := b.at(r.seed, r, 2*x+1, 2*y+1)
			if w := int(46 * glow * lane * m); w > 0 {
				g.Cells[y*g.W+x].Bg = P(HCool, w)
			}
		}
	}
	for y := 0; y < r.pix.H; y++ {
		for x := 0; x < r.pix.W; x++ {
			glow, lane, m := b.at(r.seed, r, x, y)
			if glow < 0.02 {
				continue
			}
			h := hash(r.seed, 411, int64(x), int64(y))
			if float64(h%10000) >= 1500*glow*lane*m {
				continue
			}
			w := int(45+h>>16%130) * int(100*(0.5+0.5*glow)) / 100
			hue := HStar
			switch h >> 32 % 14 {
			case 0, 1, 2:
				hue = HCool
			case 3:
				hue = HSand
			}
			r.pix.Set(x, y, P(hue, w), 0)
		}
	}
}

// oort scatters the icy specks of the Oort cloud: a shell well outside the
// Kuiper belt that thins outward, very dim, a few of them twinkling slowly.
func (r *orrery) oort(maxRY int) {
	m := float64(maxRY) / 100
	n := r.pix.W * r.pix.H / 28
	for i := 0; i < n; i++ {
		u := func(salt int64) float64 { return float64(roll(1000, r.seed, 420, int64(i), salt)) / 1000 }
		var e float64
		switch k := roll(100, r.seed, 421, int64(i)); {
		case k < 55: // the shell
			e = (oortShell + (u(1)+u(2)+u(3)-1.5)*0.9*100) * m
		case k < 85: // thinning outward
			e = (oortShell + u(4)*u(4)*(oortOut-oortShell)) * m
		default: // and the inner cloud
			e = (bubbleAt + 20 + u(5)*(oortShell-bubbleAt-20)) * m
		}
		x, y := r.toPx(e, roll(turn, r.seed, 422, int64(i)))
		w := 26 + roll(40, r.seed, 423, int64(i))
		if i%9 == 0 { // a slow twinkle
			w -= w * tri(r.s.Anim/7+roll(24, r.seed, 424, int64(i)), 12) / 20
		}
		hue := [...]Hue{HStar, HStar, HCool, HCool, HWhite}[roll(5, r.seed, 425, int64(i))]
		r.pix.Set(x, y, P(hue, w), 0)
	}
}

// fieldStars moves the sky's own stars from the cells into the pixels, over
// the haze and under the rest.
func (r *orrery) fieldStars(o *Overview) {
	g := o.Grid
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := &g.Cells[y*g.W+x]
			if c.R == ' ' || c.R == 0 {
				continue
			}
			px, py := 2*x+roll(2, r.seed, 430, int64(x), int64(y)), 2*y+roll(2, r.seed, 431, int64(x), int64(y))
			r.pix.Set(px, py, P(c.Fg.Hue, int(c.Fg.Weight)*3/4), 0)
			if c.R == '✧' || c.R == '+' {
				r.pix.Set(px+1, py, P(c.Fg.Hue, int(c.Fg.Weight)/3), 0)
			}
			c.R = ' '
		}
	}
}

// neighbors are the bright stars next door, near the edges of the sky: a tight
// pair, like a binary, on one side and a single one on the other.
func (r *orrery) neighbors() {
	W, H := r.pix.W, r.pix.H
	left := roll(2, r.seed, 440) == 0
	at := func(salt int64, onLeft bool) (int, int) {
		x := 7 + roll(max(1, W/9), r.seed, 441, salt)
		if !onLeft {
			x = W - 9 - roll(max(1, W/9), r.seed, 441, salt)
		}
		return x, H/5 + roll(max(1, H*3/5), r.seed, 442, salt)
	}
	star := func(x, y int, h Hue, w int) {
		r.pix.Set(x, y, P(h, w), 0)
		r.pix.Set(x+1, y, P(h, w), 0)
		r.pix.Glow(x*pxUnit+pxUnit, y*pxUnit+pxUnit/2, pxUnit, 5*pxUnit, h, 110, 4, 0)
	}
	x, y := at(1, left)
	star(x, y, HWhite, 255)
	star(x+4+roll(2, r.seed, 443), y+roll(3, r.seed, 444)-1, HSun, 215)
	x, y = at(2, !left)
	star(x, y, HCool, 245)
}

// bubble is the Sun's bubble, the heliosphere: a thin dim line round the
// planets and the belt, blunt and squashed on the nose side, the side the
// Milky Way runs from, and longer behind.
func (r *orrery) bubble(maxRY int) {
	b := r.newBand()
	rows := float64(maxRY) * bubbleAt / 100 * 32
	n := min(4096, int(rows*3)/4+60)
	for _, ring := range []struct {
		scale float64
		c     Paint
	}{{1, P(HCool, 70)}, {1.05, P(HCool, 26)}} {
		lx, ly := -1<<30, -1<<30
		for i := 0; i < n; i++ {
			a := i * turn / n
			ca, sa := float64(cos4096(a))/1024, float64(sin4096(a))/1024
			k := ca*-b.dx + sa*-b.dy*0.5 // toward the nose, roughly
			f := 1 - 0.14*k
			if k > 0.6 {
				f -= 0.08 * (k - 0.6) // blunt on the nose
			}
			f *= ring.scale
			x := floorDiv(r.sx+int(rows*2.3*f*ca), pxUnit)
			y := floorDiv(r.sy+int(rows*f*sa), pxUnit)
			if x != lx || y != ly {
				lx, ly = x, y
				r.pix.Set(x, y, ring.c, 0)
			}
		}
	}
}

// kuiper is the Kuiper belt at the cosmic view: a dense band that is thickest
// in the middle, varied in brightness, with a few icy bodies two pixels wide
// and one gap in it.
func (r *orrery) kuiper(maxRY int, own uint8) {
	n := maxRY * maxRY * 6
	gap := roll(turn, r.seed, 450)
	drift := r.s.Anim / 96
	for i := 0; i < n; i++ {
		a0 := roll(turn, r.seed, 451, int64(i))
		if d := (a0-gap+turn+turn/2)%turn - turn/2; d > -170 && d < 170 {
			continue
		}
		t := float64(roll(1000, r.seed, 452, int64(i))+roll(1000, r.seed, 453, int64(i))) / 2000
		e := (kuiperIn + (kuiperOut-kuiperIn)*t) * float64(maxRY) / 100
		x, y := r.toPx(e, a0+drift)
		c := P(HRock, 45+roll(120, r.seed, 454, int64(i)))
		switch k := roll(40, r.seed, 455, int64(i)); {
		case k < 6:
			c = P(HSand, 110+roll(70, r.seed, 456, int64(i)))
		case k == 7: // an icy body
			c = P(HCool, 190)
			r.pix.Set(x+1, y, c, own)
		}
		r.pix.Set(x, y, c, own)
	}
}

// beacons are the relay beacons at the edge of the Oort cloud: one pixel each,
// dim when dark and lit for a moment, slowly.
func (r *orrery) beacons() {
	n := 3 + roll(3, r.seed, 460)
	e := float64(oortShell*r.maxRY) / 100
	for k := 0; k < n; k++ {
		a := k*turn/n + roll(turn/(2*n), r.seed, 461, int64(k))
		x, y := r.toPx(e, a)
		for tries := 0; tries < 30 && (x < 3 || y < 3 || x >= r.pix.W-3 || y >= r.pix.H-3); tries++ {
			e *= 0.94
			x, y = r.toPx(e, a)
		}
		r.pix.Set(x, y, P(HSteel, 120), 0)
		if blinkWith(r.seed, r.s.Anim, 20+k, beaconPeriod, beaconLit) {
			r.pix.Set(x, y, P(HCyan, 255), 0)
		}
	}
}
