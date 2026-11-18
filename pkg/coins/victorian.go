package coins

import "math"

// Victorian money is machine-made: exact symmetry, clean lines, no wear. The
// six levels climb bronze, silver and gold; each has a second ramp of the same
// metal that a copy may lean toward.
var (
	vicChoc  = newRamp(0x140a06, 0x2e1a10, 0x482b1a, 0x643c26, 0x825036) // level 1: dark chocolate bronze
	vicChocB = newRamp(0x100c08, 0x261c12, 0x3c2c1c, 0x544028, 0x70583a) // a browner, greyer bronze
	vicRed   = newRamp(0x2a0e06, 0x5c2612, 0x92421f, 0xc4602f, 0xea8c58) // level 2: bright red-brown bronze
	vicRedB  = newRamp(0x2a1408, 0x5a3416, 0x8c5c26, 0xbc8238, 0xe0a85a) // a yellower bronze
	vicGoldB = newRamp(0x4a2c0c, 0x8c5c18, 0xcc8c2c, 0xecb848, 0xfff0a0) // a redder gold, still gold
	vicProof = newRamp(0x120c05, 0x281c0a, 0x42300f, 0x634718, 0x8a6a28) // the legendary's dark mirror field
	vicFrost = newRamp(0x6a5020, 0xa8883c, 0xd8b865, 0xf2dc9c, 0xfffbe0) // the legendary's pale frosted devices
)

// vicRamps is the metal of level l and its second ramp.
func vicRamps(l int) (main, alt ramp) {
	switch l {
	case 1:
		return vicChoc, vicChocB
	case 2:
		return vicRed, vicRedB
	case 3, 4:
		return silver, tonedClassical
	}
	return gilt, vicGoldB
}

// die is one die's varieties: the charge in each quarter of the shield, the
// tooth count at 128 px, the shield scale, tooth length scale, width of the
// sunk cross (1 is the reference for each) and the rim, broader inward (0 is
// the reference).
type die struct {
	perm       [4]int
	nT         int
	ss, tl, cw float64
	rw         float64
}

var referenceDie = die{[4]int{0, 1, 2, 3}, 56, 1, 1, 1, 0}

func dieOf(seed uint32) die {
	d := die{[4]int{0, 1, 2, 3}, 52 + int(rnd(seed, 79)*9), 1 + 0.12*rs(seed, 85), 1 + 0.25*rs(seed, 86), 1 + 0.35*rs(seed, 87), 2.6 * rnd(seed, 88)}
	for i := 3; i > 0; i-- {
		j := int(rnd(seed, 80+i) * float64(i+1))
		d.perm[i], d.perm[j] = d.perm[j], d.perm[i]
	}
	return d
}

func victorian(l, lod int, seed uint32) *spec {
	if seed == 0 {
		return victorianGen(l, lod, referenceDie)
	}
	sp := victorianGen(l, lod, dieOf(seed))
	victorianVary(sp, l, seed)
	return sp
}

// victorianLegend is the proof piece, finer than level 6: a dark mirror field
// crossed by two hard light bands, pale frosted devices standing on it, a ring
// of beads inside the teeth, a closed wreath of hand-set leaves with berries,
// a tied bow with ribbon tails, a thin second line round the shield, finer charges.
func victorianLegend() *spec { return victorianGen(0, 0, referenceDie) }

// victorianRound is the round blank. reedN > 0 gives a reeded edge (the radius
// steps in and out every sector).
func victorianRound(reedN int, reedA float64) func(x, y float64) float64 {
	return func(x, y float64) float64 {
		r := 53.2
		if reedN > 0 {
			k := int(math.Floor((angle(x, y)+math.Pi)*float64(reedN)/(2*math.Pi))) & 1
			r = 53.6 - reedA*float64(k)
		}
		return math.Hypot(x, y) - r
	}
}

// victorianRim is the raised plain rim; w widens it inward (level 6).
func victorianRim(r, w float64) (sample, bool) {
	if r < 49-w {
		return sample{}, false
	}
	d := 53.4 - r
	return sample{h: ground + 2.0*smooth(0, 1.6, d)*(1-smooth(3+w, 4.4+w, d))}, true
}

// teeth are n teeth pointing inward from the rim.
func teeth(x, y float64, n int, uc, hl, hw float64) (sample, bool) {
	u, v := around(x, y, n, -math.Pi/2)
	if dd := sdRoundBox(u, v, uc, 0, hl, hw, 0.7); dd < 0 {
		return sample{h: ground + 1.2 + 0.8*bevel(dd, 1.2)}, true
	}
	return sample{}, false
}

// leaf is a pointed lens along a direction; at (cx, cy), angle ang, length L,
// breadth b.
type leaf struct{ cx, cy, ang, L, b float64 }

func leafAt(px, py, dx, dy, L, b float64) leaf {
	return leaf{px + dx*L, py + dy*L, math.Atan2(dy, dx), L, b}
}

// leavesAt reports the leaf (if any) covering a point, with its midrib.
func leavesAt(ls []leaf, x, y float64) (sample, bool) {
	for _, l := range ls {
		if math.Abs(x-l.cx) > l.L+1 || math.Abs(y-l.cy) > l.L+1 {
			continue
		}
		u, v := rotate(x-l.cx, y-l.cy, l.ang)
		if math.Abs(u) >= l.L {
			continue
		}
		w := l.b * math.Pow(1-(u/l.L)*(u/l.L), 0.8)
		if math.Abs(v) >= w {
			continue
		}
		if math.Abs(v) < 0.55 {
			return sample{h: ground + 1.5, z: 1}, true
		}
		return sample{h: ground + 2.0 + 1.2*smooth(0, 0.7, 1-math.Abs(v)/w), z: 1}, true
	}
	return sample{}, false
}

// spray is leaves in pairs along the wreath circle from angle t0 to t1
// (degrees from the horizontal, down is positive), on the right side. j > 0
// gives a hand-set spread: length, breadth, angle and place all vary leaf to
// leaf; 0 is one even spray.
func spray(t0, t1, step, k, j float64, seed int) []leaf {
	var ls []leaf
	n := 0
	for deg := t0; deg >= t1-1e-9; deg -= step {
		n++
		for si, sg := range []float64{1, -1} {
			h := func(m int) float64 { return hash(n*7+si, m, seed) - 0.5 }
			th := (deg + j*3.2*h(1)) * math.Pi / 180
			R := 34.5 + j*1.1*h(2)
			px, py := R*math.Cos(th), R*math.Sin(th)
			tx, ty := math.Sin(th), -math.Cos(th)
			nx, ny := math.Cos(th), math.Sin(th)
			off := 0.7 + j*0.3*h(3)
			if sg > 0 { // outward leaves lean closer to the stem so none reaches the bead ring
				off = 0.6 + j*0.28*h(3)
			}
			ca, sa := math.Cos(off), math.Sin(off)
			dx, dy := tx*ca+sg*nx*sa, ty*ca+sg*ny*sa
			ls = append(ls, leafAt(px, py, dx, dy, k*5.2*(1+j*0.30*h(4)), k*2.7*(1+j*0.26*h(5))))
		}
	}
	return ls
}

func mirrorLeaves(ls []leaf) []leaf {
	var out []leaf
	for _, l := range ls {
		out = append(out, leaf{-l.cx, l.cy, math.Pi - l.ang, l.L, l.b})
	}
	return out
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// charge is the signed distance to one of the four charges (cog, anchor,
// sheaf, hammer), in a box about 9 by 9 units centred on 0,0.
func charge(k int, x, y float64) float64 {
	switch k {
	case 0: // cog
		u, v := around(x, y, 6, 0)
		d := math.Min(math.Hypot(x, y)-3.4, sdBox(u, v, 4.1, 0, 1.4, 1.45))
		return sdSub(d, math.Hypot(x, y)-1.2)
	case 1: // anchor
		d := sdSeg(x, y, 0, -4.6, 0, 3.8, 1.25)
		d = math.Min(d, sdSeg(x, y, -3.4, -2.6, 3.4, -2.6, 1.2))
		d = math.Min(d, sdSeg(x, y, 0, 4.2, -3.9, 0.4, 1.2))
		d = math.Min(d, sdSeg(x, y, 0, 4.2, 3.9, 0.4, 1.2))
		d = math.Min(d, math.Hypot(math.Abs(x)-4.2, y-0.2)-1.6)
		return d
	case 2: // wheat sheaf
		d := sdSeg(x, y, 0, 4.6, 0, -3, 1.1)
		d = math.Min(d, sdSeg(x, y, 0, 4.6, -3.6, -2, 1.1))
		d = math.Min(d, sdSeg(x, y, 0, 4.6, 3.6, -2, 1.1))
		d = math.Min(d, sdEllipse(x, y, 0, -3.8, 1.5, 2.2, 0))
		d = math.Min(d, sdEllipse(x, y, -4.1, -3.0, 1.5, 2.2, 0.55))
		d = math.Min(d, sdEllipse(x, y, 4.1, -3.0, 1.5, 2.2, -0.55))
		d = math.Min(d, sdSeg(x, y, -2.6, 1.6, 2.6, 1.6, 0.8))
		return d
	}
	xx := x + 1.7 // hammer
	d := sdSeg(xx, y, -3.2, 4.4, 2.2, -1.0, 1.15)
	return math.Min(d, sdSeg(xx, y, 0.14, -3.66, 4.66, 0.86, 1.9))
}

// chargeFine is the legendary's more worked version of each charge.
func chargeFine(k int, x, y float64) float64 {
	switch k {
	case 0: // cog: toothed wheel, eight teeth, an open hub
		u, v := around(x, y, 8, 0)
		d := math.Min(math.Hypot(x, y)-3.4, sdBox(u, v, 4.0, 0, 1.2, 1.2))
		d = sdSub(d, math.Hypot(x, y)-2.0)
		return math.Min(d, math.Hypot(x, y)-0.9)
	case 1: // anchor with a ring and broad flukes
		d := sdSeg(x, y, 0, -3.6, 0, 3.8, 1.1)
		d = math.Min(d, sdSeg(x, y, -3.2, -2.2, 3.2, -2.2, 1.0))
		d = math.Min(d, sdSeg(x, y, 0, 4.2, -4.0, 0.6, 1.1))
		d = math.Min(d, sdSeg(x, y, 0, 4.2, 4.0, 0.6, 1.1))
		d = math.Min(d, math.Hypot(math.Abs(x)-4.2, y-0.2)-1.5)
		return math.Min(d, math.Abs(math.Hypot(x, y+4.6)-1.5)-0.55)
	case 2: // sheaf: five ears, a band
		d := sdSeg(x, y, 0, 4.6, 0, -3.2, 1.0)
		d = math.Min(d, sdSeg(x, y, 0, 4.6, -3.4, -1.6, 1.0))
		d = math.Min(d, sdSeg(x, y, 0, 4.6, 3.4, -1.6, 1.0))
		d = math.Min(d, sdSeg(x, y, 0, 4.6, -1.7, -2.8, 0.9))
		d = math.Min(d, sdSeg(x, y, 0, 4.6, 1.7, -2.8, 0.9))
		for _, e := range []vec{{0, -4.2}, {-4.0, -2.9}, {4.0, -2.9}} {
			d = math.Min(d, sdEllipse(x, y, e.x, e.y, 1.3, 2.0, e.x*0.14))
		}
		return math.Min(d, sdSeg(x, y, -2.6, 1.8, 2.6, 1.8, 0.7))
	}
	d := sdSeg(x, y, -3.4, 4.4, 2.6, -1.4, 1.0) // crossed hammers
	d = math.Min(d, sdSeg(x, y, 3.4, 4.4, -2.6, -1.4, 1.0))
	d = math.Min(d, sdSeg(x, y, 0.2, -3.6, 4.4, 0.6, 1.8))
	return math.Min(d, sdSeg(x, y, -0.2, -3.6, -4.4, 0.6, 1.8))
}

// bow is the legendary's tied bow, with two ribbon tails ending in a notch.
func bow(x, y float64) (sample, bool) {
	ax := math.Abs(x)
	d := sdEllipse(ax, y, 6.6, 34.2, 6.2, 3.1, -0.38)
	d = math.Min(d, sdCircle(ax, y, 0, 34.8, 2.9))
	t := sdSub(sdSeg(ax, y, 1.4, 36.4, 5.4, 41.2, 1.5), sdCircle(ax, y, 5.9, 42.0, 1.1))
	d = math.Min(d, t)
	if d < 0 {
		return sample{h: ground + 2.4 + 1.2*bevel(d, 2.5), z: 3}, true
	}
	return sample{}, false
}

// victorianGen draws level l (1 to 6) or, for l 0, the legendary. lod 1 is
// the 48 px version.
func victorianGen(l, lod int, dv die) *spec {
	leg := l == 0
	lw := lineWidth(lod)
	sp := &spec{seed: 171, edge: 5}
	reedN, reedA := 0, 1.25
	if leg || l == 6 {
		reedN = 96
		if lod == 1 {
			reedN, reedA = 40, 2.4
		}
	}
	sp.sd = victorianRound(reedN, reedA)
	if reedN > 0 {
		sp.reed = reedN
	}
	if leg {
		sp.ramp = [4]ramp{vicProof, vicFrost, gilt, silver}
		sp.field = 2
		sp.fieldTone = 0.5
		sp.band = [2]float64{-13, -7}
		er := gilt
		sp.edgeR = &er
		sp.edgeT = 0.3
	} else {
		m, _ := vicRamps(l)
		sp.ramp = [4]ramp{m, m, m, m}
		sp.field = 0
		if l >= 3 {
			sp.field = 1
		}
		if l == 6 {
			sp.field = 2
			sp.fieldTone = 0.68
			sp.band = [2]float64{9999, 9999}
		}
		if l >= 4 {
			sp.gain = 1.25
		}
		er := m
		sp.edgeR = &er
		sp.edgeT = 0.34
	}
	shield := []vec{{-16, -16}, {16, -16}, {16, 6}, {10, 18}, {0, 28}, {-10, 18}, {-16, 6}}
	ccs := []vec{{-8.4, -9.8}, {8.4, -9.8}, {-8.2, 6.5}, {8.2, 6.5}}
	for i := range shield {
		shield[i] = vec{shield[i].x * dv.ss, shield[i].y * dv.ss}
	}
	for i := range ccs {
		ccs[i] = vec{ccs[i].x * dv.ss, ccs[i].y * dv.ss}
	}
	nT, hl, hw := dv.nT, 1.9, 1.3
	if lod == 1 {
		nT, hl, hw = dv.nT*32/56, 2.4, 2.2
	}
	hl *= dv.tl
	// level 3 has the plain rim only; the tooth ring and coronet start at 4; level 6 trades the teeth for a broad rim
	quarter, charges, crown, toothed := l >= 2 || leg, l >= 3 || leg, l >= 4 || leg, l == 4 || l == 5 || leg
	rimW := dv.rw
	if l == 6 {
		rimW += 3.4
	}
	var sprayR, sprayL []leaf
	k := 1.0
	if l == 6 {
		k = 1.5
	}
	if lod == 1 {
		k *= 1.25
	}
	switch {
	case l == 6:
		sprayR = spray(66, -32, 14, k, 0, 3)
		sprayL = mirrorLeaves(sprayR)
	case leg:
		sprayR = spray(70, -40, 11, 1, 1, 5)
		sprayL = mirrorLeaves(spray(66, -46, 10.5, 1, 1, 9)) // not a mirror image: set again by hand
	}
	// berries: hand-set places and sizes on each side, between the leaf pairs
	type berry struct{ th, r, rad float64 }
	berR := []berry{{57, 31.4, 1.8}, {40, 30.2, 1.6}, {19, 31.8, 1.9}, {1, 30.6, 1.6}, {-21, 31.6, 1.8}}
	berL := []berry{{62, 30.8, 1.7}, {45, 31.6, 1.9}, {27, 30.0, 1.6}, {8, 31.4, 1.8}, {-14, 30.4, 1.7}, {-36, 31.4, 1.6}}
	lift := 0.0 // level 6 stands the shield higher off the field
	if l == 6 {
		lift = 1.5
	}
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		if crown {
			if leg {
				for _, px := range []float64{-9, 0, 9} { // pearls on stalks above a band
					if dd := math.Hypot(x-px, y+32.2) - 3.0; dd < 0 {
						return sample{h: ground + 2.6 + 1.4*bevel(dd, 3), z: 3}
					}
					if math.Abs(x-px) < lw(2.4)/2 && y > -32 && y < -26.5 {
						return sample{h: ground + 2.2, z: 1}
					}
				}
			} else { // plain upright points on a band
				for i, px := range []float64{-9, 0, 9} {
					hh := 4.4
					if i == 1 {
						hh = 6.4
					}
					tri := []vec{{px - lw(5.2)/2, -26.9}, {px + lw(5.2)/2, -26.9}, {px, -26.9 - hh}}
					if sdPoly(x, y, tri) < 0 {
						return sample{h: ground + 2.6, z: 1}
					}
				}
			}
			if sdBox(x, y, 0, -24.6, 12.5, 2.3) < 0 {
				return sample{h: ground + 2.8, z: 1}
			}
		}
		if sd := sdPoly(x, y, shield); sd < 0 {
			if sd > -1.5 {
				return sample{h: ground + 3.0 + lift, z: 1}
			}
			if quarter {
				if math.Abs(x) < lw(2.8*dv.cw)/2 || math.Abs(y+3*dv.ss) < lw(2.8*dv.cw)/2 {
					return sample{h: ground}
				}
				if charges {
					if lod == 0 {
						for i, c := range ccs {
							d := charge(dv.perm[i], x-c.x, y-c.y)
							if leg {
								d = chargeFine(dv.perm[i], x-c.x, y-c.y)
							}
							if d < 0 {
								return sample{h: ground + 3.6 + lift, z: 1}
							}
						}
					} else if q := b2i(x >= 0) + 2*b2i(y >= -3); dv.perm[q]%3 == 0 {
						return sample{h: ground + 3.6 + lift, z: 1}
					}
				}
			}
			return sample{h: ground + 1.9 + lift, z: 1}
		} else if leg && sd > 2.1 && sd < 3.6 {
			return sample{h: ground + 1.3, z: 1} // the thin second line
		}
		if leg {
			if s, ok := bow(x, y); ok {
				return s
			}
			for side, bs := range [][]berry{berR, berL} {
				for _, b := range bs {
					th := b.th * math.Pi / 180
					bx, by := b.r*math.Cos(th), b.r*math.Sin(th)
					if side == 1 {
						bx = -bx
					}
					if dd := math.Hypot(x-bx, y-by) - b.rad*1.3; dd < 0 {
						return sample{h: ground + 2.6 + 1.3*bevel(dd, b.rad*1.3), z: 3}
					}
				}
			}
		}
		if sprayR != nil {
			for _, ls := range [][]leaf{sprayR, sprayL} {
				if s, ok := leavesAt(ls, x, y); ok {
					if s.h < ground+2 {
						s.h = ground + 2.2
					}
					return s
				}
			}
			if math.Abs(r-34.5) < lw(2.5)/2 {
				th := math.Atan2(y, math.Abs(x)) * 180 / math.Pi
				lo, hi := -38.0, 70.0
				if leg {
					lo, hi = -44, 72
				}
				if th < hi && th > lo {
					return sample{h: ground + 1.8, z: 1}
				}
			}
		}
		if leg && r > 41 && r < 44.4 && !(math.Abs(x) < 9 && y > 0) {
			u, v := around(x, y, 44, -math.Pi/2) // a ring of beads inside the teeth
			if dd := math.Hypot(u-42.8, v) - 1.55; dd < 0 {
				return sample{h: ground + 1.8 + 1.2*bevel(dd, 1.55)}
			}
		}
		if toothed {
			if s, ok := teeth(x, y, nT, 46.9, hl, hw); ok {
				return s
			}
		}
		if s, ok := victorianRim(r, rimW); ok {
			return s
		}
		return sample{h: ground}
	}
	if leg { // the field alone stays z 0 (dark); every raised plain part (rim, teeth, beads, stem) is bright gold, z 2
		f := sp.face
		sp.face = func(x, y float64) sample {
			st := f(x, y)
			if st.z == 0 && st.h != ground {
				st.z = 2
			}
			return st
		}
	}
	return sp
}

// victorianVary gives a machine-made copy its own look, with no damage of any
// kind: the piece lies turned up to 25 degrees either way (the light stays
// upper left, since it is lit after the turn), carries a small mint mark in
// one of three fixed places, and takes a slightly different alloy, one copy
// in three on the second ramp of its metal. The strike stays centred.
func victorianVary(sp *spec, l int, seed uint32) {
	a := rs(seed, 70) * 25 * math.Pi / 180
	f, sd := sp.face, sp.sd
	places := []vec{{0, 38.5}, {-27, -28}, {27, -28}}
	pl := places[int(rnd(seed, 72)*3)%3]
	kind := int(rnd(seed, 73)*4) % 4
	mark := func(x, y float64) float64 {
		x, y = x-pl.x, y-pl.y
		switch kind {
		case 0: // pellet
			return math.Hypot(x, y) - 4.3
		case 1: // four-point star
			return sdPoly(x, y, []vec{{0, -5}, {1.7, -1.7}, {5, 0}, {1.7, 1.7}, {0, 5}, {-1.7, 1.7}, {-5, 0}, {-1.7, -1.7}})
		case 2: // crescent
			return sdSub(math.Hypot(x, y)-4.5, math.Hypot(x-2.1, y+0.7)-3.3)
		}
		return sdPoly(x, y, []vec{{0, -5}, {3.6, 0}, {0, 5}, {-3.6, 0}}) // lozenge
	}
	sp.sd = func(x, y float64) float64 { u, v := rotate(x, y, a); return sd(u, v) }
	sp.face = func(x, y float64) sample {
		u, v := rotate(x, y, a)
		if mark(u, v) < 0 && math.Hypot(u, v) < 46 {
			return sample{h: ground + 2.8, z: 1}
		}
		return f(u, v)
	}
	main, alt := vicRamps(l)
	t := 0.5 * rnd(seed, 74)
	if rnd(seed, 77) < 0.34 {
		t = 0.65 + 0.35*rnd(seed, 78)
	}
	m := mixRamp(main, alt, t)
	sp.ramp = [4]ramp{m, m, m, m}
	sp.edgeR = &m
}
