package coins

import "math"

// The Iron Age: spits bound into a bundle, with a clasp at the top.

var (
	ironRough  = newRamp(0x1a1410, 0x342a22, 0x524438, 0x76655a, 0xa08e80) // levels 1-2: brown raw iron
	ironGrey   = newRamp(0x14161a, 0x2a2d33, 0x464a52, 0x6a6f78, 0x9aa0aa) // levels 3-5: forged iron
	ironSteel  = newRamp(0x0c1018, 0x1c2434, 0x2e3a52, 0x4a5a78, 0x8a9cc0) // level 6 and legendary
	ironCord   = newRamp(0x1e140a, 0x3e2c18, 0x5e4628, 0x80663e, 0xa88e60) // levels 1-4 bindings
	ironBright = newRamp(0x1c1e22, 0x3a3e46, 0x5e646e, 0x8c929c, 0xc4cad2)
	ironWire   = newRamp(0x2a1a08, 0x5c3a14, 0x8e5e22, 0xc08a3c, 0xe8c070) // levels 5-6 bindings and clasp
)

// ironLie is how far a copy may lie off the -45 degree axis, in degrees.
const ironLie = 11.0

// ironRod is one spit: across offset, centre and length along the axis, bend.
type ironRod struct{ off, c, ln, b float64 }

// ironBand is a binding w wide at axis position tb, leaning tilt across.
type ironBand struct{ tb, w, tilt float64 }

// ironBundle is spits on one axis from origin (ox, oy). head: 0 round, 1 paddle, 2 eye, 3 leaf.
type ironBundle struct {
	ang, ox, oy, splay, tb0, fan, r, taper, W, clasp float64
	hasClasp                                         bool
	head                                             int
	rods                                             []ironRod
	bands                                            []ironBand
}

func ironPick(a []float64, i int) float64 { return a[min(i, len(a)-1)] }

func (b *ironBundle) width() float64 {
	m := 0.0
	for _, rd := range b.rods {
		m = math.Max(m, math.Abs(rd.off))
	}
	return m + b.r
}

// span holds t within rod i.
func (b *ironBundle) span(i int, t float64) float64 {
	r := b.rods[i]
	return clamp(t, r.c-r.ln/2, r.c+r.ln/2)
}

// centre is rod i's centre line across, at t.
func (b *ironBundle) centre(i int, t float64) float64 {
	r := b.rods[i]
	k := (t - r.c) / (r.ln / 2)
	return r.off + r.b*(1-k*k) + r.off*(b.splay*math.Max(0, b.tb0-10-t)/10+b.fan*math.Max(0, t-30)/10)
}

// radius tapers rod i from r to a 0.8 tip over the points.
func (b *ironBundle) radius(i int, t float64) float64 {
	if b.taper == 0 {
		return b.r
	}
	t0 := b.rods[i].c - b.rods[i].ln/2
	return 0.8 + (b.r-0.8)*clamp((t-t0)/b.taper, 0, 1)
}

func (b *ironBundle) dome(o float64) float64 {
	return ground + 2.8*math.Cos(math.Pi/2*clamp(o/b.W, -1, 1))
}

// eye is the centre of rod i's eye, 2 past its end on the rod line.
func (b *ironBundle) eye(i int) (float64, float64) {
	ec := b.rods[i].c + b.rods[i].ln/2 + 2
	return ec, b.centre(i, ec)
}

func (b *ironBundle) rodSd(i int, t, o float64) float64 {
	tt := b.span(i, t)
	return math.Hypot(t-tt, o-b.centre(i, tt)) - b.radius(i, tt)
}

// headSd is the distance to rod i's head, negative inside.
func (b *ironBundle) headSd(i int, t, o float64) float64 {
	t1 := b.rods[i].c + b.rods[i].ln/2
	switch b.head {
	case 1:
		return sdEllipse(t, o, t1-2.5, b.centre(i, t1-2.5), 5, 1.8*b.r, 0)
	case 2:
		ec, eo := b.eye(i)
		return sdSub(sdCircle(t, o, ec, eo, 4.2), sdCircle(t, o, ec, eo, 2))
	case 3:
		return sdEllipse(t, o, t1+3, b.centre(i, t1+3), 9, 4.6, 0)
	}
	return 1e9
}

func (b *ironBundle) bandSd(bd ironBand, t, o float64) float64 {
	return sdBox(t-bd.tb-bd.tilt*o, o, 0, 0, bd.w/2, b.W+1)
}

func (b *ironBundle) claspSd(t, o float64) float64 {
	return sdRoundBox(t, o, b.clasp, 0, 5, b.W+1.5, 2)
}

// sd is the outline, negative inside. Each eye's opening is cut out of it.
func (b *ironBundle) sd(t, o float64) float64 {
	d := 1e9
	for i := range b.rods {
		d = math.Min(d, math.Min(b.rodSd(i, t, o), b.headSd(i, t, o)))
		if b.head == 2 {
			ec, eo := b.eye(i)
			d = sdSub(d, sdCircle(t, o, ec, eo, 2))
		}
	}
	for _, bd := range b.bands {
		d = math.Min(d, b.bandSd(bd, t, o))
	}
	if b.hasClasp {
		d = math.Min(d, b.claspSd(t, o))
	}
	return d
}

// dist is the outline at (x, y).
func (b *ironBundle) dist(x, y float64) float64 {
	t, o := rotate(x-b.ox, y-b.oy, b.ang)
	return b.sd(t, o)
}

// headFace: a paddle 0.4 above the bundle, an eye as a torus, a leaf as a rounded blade.
func (b *ironBundle) headFace(i int, t, o float64) float64 {
	switch b.head {
	case 1:
		return b.dome(o) + 0.4
	case 2:
		ec, eo := b.eye(i)
		s := (math.Hypot(t-ec, o-eo) - 3.1) / 1.1
		return b.dome(o) + math.Sqrt(math.Max(0, 1-s*s))
	}
	d := (o - b.centre(i, b.rods[i].c+b.rods[i].ln/2+3)) / 3.4
	return b.dome(o) + math.Sqrt(math.Max(0, 1-d*d))
}

// face is what the bundle shows at (x, y), and whether (x, y) is inside it.
func (b *ironBundle) face(x, y float64) (sample, bool) {
	t, o := rotate(x-b.ox, y-b.oy, b.ang)
	if b.sd(t, o) >= 0 {
		return sample{h: ground}, false
	}
	if b.hasClasp && b.claspSd(t, o) < 0 {
		return sample{h: ground + 5.8, z: 1}, true
	}
	for _, bd := range b.bands {
		if b.bandSd(bd, t, o) < 0 {
			return sample{h: b.dome(o) + 1.2, z: 1}, true
		}
	}
	for i := range b.rods {
		if b.headSd(i, t, o) < 0 {
			return sample{h: b.headFace(i, t, o)}, true
		}
	}
	top := ground
	for i := range b.rods {
		if b.rodSd(i, t, o) < 0 {
			tt := b.span(i, t)
			d := (o - b.centre(i, tt)) / b.radius(i, tt)
			top = math.Max(top, b.dome(o)+math.Sqrt(math.Max(0, 1-d*d)))
		}
	}
	return sample{h: top}, true
}

// ironVary gives a copy its seed's variation. tbs is shifted in place.
func ironVary(b *ironBundle, l, lod int, seed uint32, tbs []float64) {
	if l >= 2 && lod == 0 {
		n := len(b.rods)
		if rnd(seed, 1) < 0.35 {
			b.rods = b.rods[:n-1]
		} else if rnd(seed, 1) > 0.70 {
			last, prev := b.rods[n-1], b.rods[n-2]
			last.off += last.off - prev.off
			b.rods = append(b.rods, last)
		}
	}
	for i := range b.rods {
		b.rods[i].ln += rs(seed, 10+i) * level(l, 9, 9, 6, 6, 5, 4)
		if l <= 2 {
			b.rods[i].b += rs(seed, 20+i) * 1.5
		}
	}
	if l >= 3 {
		k := int(rnd(seed, 30) * float64(len(b.rods)))
		b.rods[k].b = rs(seed, 31) * level(l, 0, 0, 1.6, 1.4, 1.2, 0.8)
	}
	shift := rs(seed, 40) * 6
	for i := range tbs {
		tbs[i] += shift
	}
	b.splay *= 0.6 + 0.8*rnd(seed, 50)
	b.ang += rs(seed, 60) * ironLie * math.Pi / 180
}

// iron is the Iron Age spit bundle at level l (1 to 6), lod 0 or 1. Seed 0 is the plain table.
func iron(l, lod int, seed uint32) *spec {
	tbs := [][]float64{{6}, {0}, {10}, {-8, 16}, {-8, 4, 16}, {-8, 16}}[l-1]
	wide := []float64{-12.5, -7.5, -2.5, 2.5, 7.5, 12.5}
	offs := [2][][]float64{
		{{-7.5, 0, 6.5}, {-11, -5.5, 0, 5.5, 11}, wide, wide, wide, {-13.5, -9, -4.5, 0, 4.5, 9, 13.5}},
		{{-8, 0, 8}, {-9, -3, 3, 9}, {-9, -3, 3, 9}, {-9, -3, 3, 9}, {-9, -3, 3, 9}, {-10, -5, 0, 5, 10}},
	}[lod][l-1]
	rad := [][2]float64{{3.4, 4.2}, {3.0, 3.6}, {2.6, 3.2}, {2.6, 3.2}, {2.6, 3.2}, {2.2, 2.8}}[l-1][lod]
	lens := [][]float64{{70, 60, 78}, {84, 76, 88, 80, 82}, {96}, {96}, {96}, {100}}[l-1]
	cs := [][]float64{{4, -6, 0}, {0, -4, 2, -2, 3}, {0}, {0}, {0}, {0}}[l-1]
	bends := [][]float64{{2.5, -1.8, 1.2}, {1.2, -0.8, 0.6, -1.0, 0.9}, {0}, {0}, {0}, {0}}[l-1]

	b := &ironBundle{ang: -math.Pi / 4, r: rad, hasClasp: l == 6, head: []int{0, 0, 0, 1, 2, 2}[l-1],
		taper: level(l, 0, 10, 14, 16, 16, 18), splay: level(l, 0, 0, 0.10, 0.14, 0.16, 0.18)}
	for i, o := range offs {
		b.rods = append(b.rods, ironRod{off: o, c: ironPick(cs, i), ln: ironPick(lens, i), b: ironPick(bends, i)})
	}
	if seed != 0 {
		ironVary(b, l, lod, seed, tbs)
	}
	b.tb0, b.W, b.clasp = tbs[0], b.width(), tbs[len(tbs)-1]+7
	bf := 1.0
	if lod == 1 {
		bf = 1.6
	}
	tie, wrap := math.Tan(12*math.Pi/180), 1.6*bf
	for _, tb := range tbs {
		switch {
		case l == 1:
			b.bands = append(b.bands, ironBand{tb, 4 * bf, tie})
		case l <= 4:
			b.bands = append(b.bands, ironBand{tb, level(l, 0, 5, 6, 6) * bf, 0})
		case l == 5:
			for k := -1.0; k <= 1; k++ {
				b.bands = append(b.bands, ironBand{tb + k*(wrap+0.6), wrap, 0})
			}
		default:
			b.bands = append(b.bands, ironBand{tb - 2.5, 3 * bf, 0}, ironBand{tb + 2.5, 3 * bf, 0})
		}
	}
	body := []ramp{ironRough, ironRough, ironGrey, ironBright, ironBright, ironSteel}[l-1]
	bind := ironCord
	if l >= 4 {
		bind = ironWire
	}
	sc := 1.15
	if lod == 1 {
		sc = 1.40
	}
	sp := &spec{seed: 307, edge: 2.2, rimW: 0, rimH: 0, field: 3, ramp: [4]ramp{body, bind, body, bind}}
	if l == 6 {
		sp.bright = 0.08
	}
	sp.sd = func(x, y float64) float64 { return sc * b.dist(x/sc, y/sc) }
	sp.face = func(x, y float64) sample { s, _ := b.face(x/sc, y/sc); return s }
	return sp
}

// ironLegend is two crossed steel bundles: a small one at +30 degrees with a
// gold band, and the larger on top with fanned leaf heads and three gold bands.
func ironLegend() *spec {
	sec := &ironBundle{ang: math.Pi / 6, ox: 6, oy: 14, r: 2.2}
	for _, o := range []float64{-6, -2, 2, 6} {
		sec.rods = append(sec.rods, ironRod{off: o, ln: 64})
	}
	sec.bands, sec.W = []ironBand{{tb: 0, w: 4}}, sec.width()
	main := &ironBundle{ang: -math.Pi / 4, r: 2.2, taper: 18, splay: 0.22, fan: 0.25, head: 3, tb0: -10}
	for _, o := range []float64{-13.5, -9, -4.5, 0, 4.5, 9, 13.5} {
		main.rods = append(main.rods, ironRod{off: o, ln: 104})
	}
	main.bands = []ironBand{{tb: -10, w: 3.5}, {tb: 6, w: 3.5}, {tb: 22, w: 3.5}}
	main.W = main.width()

	sp := &spec{seed: 307, edge: 2.6, field: 3, bright: 0.10, ramp: [4]ramp{ironSteel, gilt, ironSteel, gilt}}
	sp.sd = func(x, y float64) float64 { return math.Min(main.dist(x, y), sec.dist(x, y)) }
	sp.face = func(x, y float64) sample {
		if s, ok := main.face(x, y); ok {
			return s
		}
		s, _ := sec.face(x, y)
		return s
	}
	return sp
}
