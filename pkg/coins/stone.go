package coins

import "math"

// The Stone Age: a bead string, chalk discs up to graded shell and red stone; the legendary winds twice round a shell.

var (
	stChalk = newRamp(0x32322a, 0x6c6a5a, 0xa09c86, 0xc8c4ac, 0xe6e2cc) // levels 1-2 beads
	stGreen = newRamp(0x1c1e1a, 0x3c4038, 0x6a705f, 0x9aa08a, 0xc8ccb4) // level 3 beads, levels 1-2 edge
	stBone  = newRamp(0x3a3024, 0x7a6a52, 0xb8a888, 0xe0d4b8, 0xfaf2de) // levels 4-6 shell
	stRed   = newRamp(0x2a0a06, 0x6a1c0e, 0xa8381c, 0xd9622e, 0xf29a62) // levels 5-6 barrel beads
	stDark  = newRamp(0x101214, 0x262a30, 0x444c58, 0x6c7888, 0x98a4b4) // levels 4-6 spacers
	stCord  = newRamp(0x20150c, 0x3e2a18, 0x6a4a2a, 0x96703e, 0xc09a5c)
)

// stBead is one bead: a half-width across the cord, b half its length along it; ribs > 0 makes a ribbed oval shell.
type stBead struct {
	x, y, rot, a, b, wob float64
	z, ribs              int
	hi                   bool
}

func (b *stBead) sd(x, y float64) float64 {
	p, q := rotate(x-b.x, y-b.y, b.rot)
	d := sdTrueEllipse(p, q, b.b, b.a, 0)
	if b.wob > 0 {
		d += b.wob * (math.Sin(3*math.Atan2(q, p)+b.a) + 0.6*math.Sin(5*math.Atan2(q, p)+7*b.b))
	}
	return d
}

// stLay seats beads along a polyline, margin m, spare length spread by gw; fill > 1 grows them (to fill times) until they touch.
func stLay(bs []stBead, pts []vec, m, fill float64, gw func(int) float64) {
	cum, ws, gs := []float64{0}, 0.0, 0.0
	for i := 1; i < len(pts); i++ {
		cum = append(cum, cum[i-1]+math.Hypot(pts[i].x-pts[i-1].x, pts[i].y-pts[i-1].y))
	}
	P := cum[len(cum)-1]
	for i := range bs {
		ws, gs = ws+2*bs[i].b, gs+gw(i)
	}
	k := clamp((P-2*m)/ws, math.Min(1.5, fill), fill)
	ws *= k
	s := m
	for i := range bs {
		b := &bs[i]
		b.a, b.b = b.a*k, b.b*k
		s += b.b
		j := 0
		for j < len(pts)-2 && cum[j+1] < s {
			j++
		}
		f, p0, p1 := (s-cum[j])/(cum[j+1]-cum[j]), pts[j], pts[j+1]
		b.x, b.y, b.rot = p0.x+f*(p1.x-p0.x), p0.y+f*(p1.y-p0.y), b.rot+math.Atan2(p1.y-p0.y, p1.x-p0.x)
		s += b.b + math.Max(P-ws-2*m, 0)/gs*gw(i)
	}
}

func stSpec(bs []stBead, lines [][]vec, knots []stBead, ramps [4]ramp, w float64) *spec {
	cd := func(x, y float64) (d float64) {
		d = 1e9
		for _, l := range lines {
			d = math.Min(d, polyline(x, y, l, w))
		}
		return d
	}
	bs, sp := append(bs, knots...), &spec{seed: 331, edge: 3, ramp: ramps}
	near := func(x, y float64) (float64, int) {
		best, bi := 0.0, -1
		for i := range bs {
			if d := bs[i].sd(x, y); d < best {
				best, bi = d, i
			}
		}
		return best, bi
	}
	sp.sd = func(x, y float64) float64 {
		if best, bi := near(x, y); bi >= 0 {
			return best
		}
		return cd(x, y)
	}
	sp.face = func(x, y float64) sample {
		best, bi := near(x, y)
		if bi < 0 {
			return sample{h: ground + 1.4*bevel(cd(x, y), 2), z: 3}
		}
		b := &bs[bi]
		p, q := rotate(x-b.x, y-b.y, b.rot)
		if b.ribs == 0 && b.z != 3 && b.b > 0.62*b.a && sdTrueEllipse(p, q, b.b*0.3, b.a*0.28, 0) < 0 { // the drilled hole
			return sample{h: ground - 1, z: 3, tone: 0.08, seal: true}
		}
		mn := math.Min(b.a, b.b)
		h := ground + 2.6*bevel(best, mn*0.9)
		if b.ribs > 0 { // ribs across the width
			h += 0.4 * math.Cos(float64(b.ribs)*math.Pi*q/b.a) * smooth(0, 3, -best)
		}
		if b.hi && best < -1 && math.Hypot(x-b.x+0.32*mn, y-b.y+0.36*mn) < 0.8+0.14*mn { // hard highlight
			return sample{h: h, z: b.z, tone: 0.97, seal: true}
		}
		return sample{h: h, z: b.z}
	}
	return sp
}

func stPath(n int, f func(t float64) vec) (p []vec) {
	for i := 0; i <= n; i++ {
		p = append(p, f(float64(i)/float64(n)))
	}
	return p
}

func stKnot(p vec, r float64) stBead { return stBead{x: p.x, y: p.y, a: r, b: r, z: 3} }

func stEll(cx, cy, rx, ry, tilt, th, ph1, ph2 float64) vec {
	r := 1 + 0.06*math.Sin(2*th+ph1) + 0.035*math.Sin(3*th+ph2)
	x, y := rx*r*math.Cos(th), ry*r*math.Sin(th)
	return vec{cx + x*math.Cos(tilt) - y*math.Sin(tilt), cy + x*math.Sin(tilt) + y*math.Cos(tilt)}
}

// stTails is two unequal cord tails leaving the knot k outward (heading nu).
func stTails(k vec, nu float64, seed uint32) [][]vec {
	v, out := float64(min(seed, 1)), [][]vec{}
	sg := 1 + v*(2*math.Floor(2*rnd(seed, 20))-2)
	for j, al := range []float64{0.45, -0.4} {
		l := []float64{21, 15}[j] * (1 + 0.3*v*rs(seed, 21+j))
		a := nu + sg*(al+0.25*v*rs(seed, 23+j))
		out = append(out, []vec{k, {k.x + 0.5*l*math.Cos(a), k.y + 0.5*l*math.Sin(a)}, {k.x + l*math.Cos(a+sg*0.18), k.y + l*math.Sin(a+sg*0.18)}})
	}
	return out
}

// stone is the bead string at level l, lod 0 or 1; seed 0 is the plain table.
func stone(l, lod int, seed uint32) *spec {
	v, small := float64(min(seed, 1)), float64(lod)
	hs := func(i, k int) float64 { return hash(i, k, 331+17*int(seed)) }
	R := func(k int, w float64) float64 { return w * rs(seed, k) * v }
	n := int(math.Round(level(l, 7, 11, 15, 11, 15, 15)*(1-0.4*small))) + int(math.Round(R(3, 1.6)))
	var bs []stBead
	rl := 2 + int(hs(0, 9)*2.5)
	for i, run := 0, 0; i < n; i++ {
		bs = append(bs, stBead{z: int(level(l, 1, 1, 0, 1, 1, 1))})
		if run++; l >= 4 && run >= rl && i < n-1 {
			bs, run, rl = append(bs, stBead{}), 0, 2+int(hs(i, 9)*2.5)
		}
	}
	m, pk := float64(len(bs)), 0.56+R(8, 0.08)
	for i := range bs {
		b := &bs[i]
		g := math.Max(0, 1-math.Abs((float64(i)+0.5)/m-pk)/0.55) // 1 at the bottom of the loop
		if l >= 4 && b.z == 0 {
			b.a, b.b, b.rot = 5.2+0.9*hs(i, 1), 2.2+0.5*hs(i, 2), (hs(i, 4)-0.5)*0.3
			continue
		}
		h1, h2 := hs(i, 1), hs(i, 2)
		b.a = level(l, 8.4+3.4*h1, 7.5+3*h1, 7.2+0.8*h1, 6.4+0.9*h1, 3.8+6.2*math.Pow(g, 1.1)+0.5*h1, 3.6+6.4*math.Pow(g, 1.2)+0.4*h1)
		b.b = level(l, b.a*(0.42+0.5*h2), b.a*(0.42+0.5*h2), 4.4+0.4*h2, b.a*(0.9+0.08*h2), b.a*0.5+0.5*h2, b.a*0.46+0.5*h2)
		b.rot = (hs(i, 4) - 0.5) * level(l, 0.9, 0.6, 0.3, 0.1, 0.3, 0.2)
		b.wob, b.hi = level(l, 1.3, 0.9, 0, 0, 0, 0), l == 6
	}
	if l >= 5 { // red stone beads at hand-set places
		for j, f := range map[int][]float64{5: {0.2, 0.52, 0.8}, 6: {0.16, 0.38, 0.62, 0.86}}[l] {
			k := int((f + R(30+j, 0.05)) * m)
			for bs[k].z != 1 {
				k++
			}
			bs[k].z, bs[k].a, bs[k].b = 2, bs[k].a*0.9, bs[k].b*1.35
		}
	}
	rx, ry := level(l, 43, 43, 38, 37, 44, 40)*(1+R(10, 0.05)), level(l, 33, 33, 29, 31, 30, 26)*(1+R(11, 0.07))*(1+0.22*small)
	cy, tilt, th0, ph1, ph2 := level(l, 2, 0, 0, 0, 0, -10), 0.22+R(1, 0.3), -1.2+R(2, 0.8), 1+R(12, 3.5), R(13, 3.5)
	pt := func(th float64) vec { return stEll(0, cy, rx, ry, tilt, th, ph1, ph2) }
	gw := func(i int) float64 {
		return level(l, 0.4+1.6*math.Pow(hs(i, 7), 2), 0.4+1.6*math.Pow(hs(i, 7), 2), 0.7+0.6*hs(i, 7), 0.6+0.8*hs(i, 7))
	}
	gap, margin, fill := level(l, 0.95, 0), level(l, 14, 10), 1.0
	a0, a1 := th0+gap, th0+2*math.Pi-gap
	if lod == 1 {
		fill, margin = 2.1, margin*0.6
	}
	pts := stPath(120, func(t float64) vec { return pt(a0 + (a1-a0)*t) })
	stLay(bs, pts, margin, fill, gw)
	kr := 1 + 0.4*small
	lines, knots := [][]vec{pts}, []stBead{stKnot(pts[0], 3.4*kr), stKnot(pts[120], 3.4*kr)}
	if k := pts[0]; l > 1 {
		lines, knots = append(lines, stTails(k, math.Atan2(k.y-cy, k.x), seed)...), []stBead{stKnot(k, 5.2*kr)}
	}
	if l == 6 { // the carved scallop hanging at the bottom on its own cord
		o := pt(th0 + math.Pi)
		nu, d := math.Atan2(o.y-cy, o.x), 24+4*small
		c := vec{o.x + d*math.Cos(nu), o.y + d*math.Sin(nu)}
		lines = append(lines, []vec{o, c})
		bs = append(bs, stBead{x: c.x, y: c.y, rot: nu, a: 10 + 2*small + R(14, 0.6), b: 14 + 2*small, z: 1, ribs: 5, hi: true})
	}
	sp := stSpec(bs, lines, knots, [4]ramp{stGreen, stChalk, stRed, stCord}, lineWidth(lod)(1.9))
	if l >= 4 {
		sp.ramp[0], sp.ramp[1] = stDark, stBone
	}
	sp.field, sp.gain, sp.bright = int(level(l, 0, 0, 0, 0, 0, 1)), level(l, 1, 1, 1, 1, 1, 1.25), level(l, 0.09, 0.06, 0.02, 0, 0, 0)
	return sp
}

// stoneLegend: one string wound twice in a spiral round a carved shell.
func stoneLegend() *spec {
	hs := func(i, k int) float64 { return hash(i, k, 977) }
	var bs []stBead
	pts := stPath(240, func(t float64) vec { return stEll(0, 0, 51*(1-0.5*t), 56*(1-0.5*t), 0.2, -0.9+4*math.Pi*t, 1, 0.4) })
	for i := 0; i < 44; i++ {
		a := 5.6 - 2.2*float64(i)/44 + 0.4*hs(i, 1)
		b := stBead{a: a, b: a*0.72 + 0.4*hs(i, 4), rot: (hs(i, 7) - 0.5) * 0.2, z: 1, hi: true}
		switch i {
		case 6, 17, 29, 38:
			b.z, b.a, b.b = 2, a*0.9, b.b*1.3
		case 10, 23, 33:
			b = stBead{a: 4 + 0.6*hs(i, 3), b: 1.9 + 0.4*hs(i, 5), rot: b.rot}
		}
		bs = append(bs, b)
	}
	stLay(bs, pts, 8, 1, func(i int) float64 { return 0.5 + 0.5*hs(i, 7) })
	end := pts[240]
	bs = append(bs, stBead{x: 0, y: 1, rot: math.Atan2(end.y, end.x) + math.Pi, a: 11, b: 15, z: 1, ribs: 5, hi: true})
	lines := append([][]vec{pts, {end, {0, 1}}}, stTails(pts[0], math.Atan2(pts[0].y, pts[0].x), 0)...)
	sp := stSpec(bs, lines, []stBead{stKnot(pts[0], 4.6)}, [4]ramp{stDark, stBone, stRed, stCord}, 2.4)
	sp.field, sp.gain, sp.edge = 1, 1.25, 4.5
	return sp
}
