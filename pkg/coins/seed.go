package coins

import "math"

// The hand-struck ages differ in how a copy varies; vary switches on one of these.
const (
	kindClassical = iota
	kindMedieval
)

var (
	tonedClassical = newRamp(0x1a2230, 0x3a4a64, 0x687ea0, 0x98accb, 0xd0e0f6) // toned silver, bluish
	tonedMedieval  = newRamp(0x1a1e1c, 0x3c4440, 0x68726a, 0x929c92, 0xc2cabc) // toned silver, pewter green
	toneSilver     = newRamp(0x14181f, 0x303a4c, 0x586478, 0x8892a4, 0xb4bccc) // grey first, then blue-grey
	tonePewter     = newRamp(0x16181a, 0x363a3e, 0x5e6468, 0x8a9094, 0xb4b8bc)
)

// ---- the seed: a few independent numbers out of one ----

func mixSeed(seed uint32, k int) uint32 {
	n := seed*0x9e3779b1 + uint32(k)*0x85ebca6b + 0x27d4eb2f
	for i := 0; i < 2; i++ {
		n ^= n >> 16
		n *= 0x7feb352d
		n ^= n >> 15
		n *= 0x846ca68b
		n ^= n >> 16
	}
	return n
}

func rnd(seed uint32, k int) float64 { return float64(mixSeed(seed, k)>>8) / float64(1<<24) } // 0 to 1
func rs(seed uint32, k int) float64  { return 2*rnd(seed, k) - 1 }                            // -1 to 1

// ---- the hand-struck finish ----

// edgeRadius is the radius at which the outline d0 lies along angle a, searched from outside in.
func edgeRadius(d0 func(x, y float64) float64, a float64) float64 {
	for r := 90.0; r > 0; r -= 0.5 {
		if d0(r*math.Cos(a), r*math.Sin(a)) < 0 {
			return r
		}
	}
	return 40
}

// handStruck gives levels 1 to 5 the hand-made finish: a darker metal, low
// relief, an uneven rim, a few nicks in the edge, the strike off centre and
// the field one pure tone. Level 6 is untouched.
func handStruck(sp *spec, l int) {
	if l > 5 || l < 1 {
		return
	}
	k := []float64{0, 0.55, 0.9, 1.2, 1.1, 1.05}[l]
	amp := []float64{0, 3.0, 2.0, 1.2, 0.8, 0.4}[l]
	dx, dy := []float64{0, 2.6, 1.5, 0.8, 0.4, 0.2}[l], []float64{0, -2, 1.0, 0.5, 0.3, 0.1}[l]
	nn := []int{0, 3, 2, 1, 1, 0}[l]
	angs := []float64{0.9, 2.7, 4.6}
	sizes := []float64{2.8, 2.4, 3.0}
	d0 := sp.sd
	type nick struct{ x, y, r float64 }
	var nicks []nick
	for i := 0; i < nn; i++ {
		rb := edgeRadius(d0, angs[i])
		nicks = append(nicks, nick{(rb + 0.6) * math.Cos(angs[i]), (rb + 0.6) * math.Sin(angs[i]), sizes[i]})
	}
	sp.field = 3
	sp.bright += []float64{0, -0.16, -0.10, -0.06, -0.04, -0.02}[l]
	sp.rimH *= []float64{0, 0.55, 0.8, 1, 1, 1}[l]
	f0 := sp.face
	sp.face = func(x, y float64) sample {
		st := f0(x-dx, y-dy)
		st.h = ground + (st.h-ground)*k
		return st
	}
	sp.sd = func(x, y float64) float64 {
		a := math.Atan2(y, x)
		d := d0(x, y) + amp*(0.6*math.Sin(2*a+0.5)+0.4*math.Sin(5*a+2.0))
		for _, n := range nicks {
			d = sdSub(d, sdCircle(x, y, n.x, n.y, n.r))
		}
		return d
	}
}

// ---- toning ----

// relief returns how far (in world units, up to 12) a point lies from the
// nearest raised or sunk part of the face, built once on a one-unit grid.
func relief(f func(x, y float64) sample) func(x, y float64) float64 {
	const N, O = 150, 75
	rel := make([]bool, N*N)
	for j := 0; j < N; j++ {
		for i := 0; i < N; i++ {
			rel[j*N+i] = math.Abs(f(float64(i-O), float64(j-O)).h-ground) > 0.4
		}
	}
	dist := make([]float32, N*N)
	for j := 0; j < N; j++ {
		for i := 0; i < N; i++ {
			best := 12.0
			if rel[j*N+i] {
				best = 0
			} else {
				for dj := -12; dj <= 12; dj++ {
					for di := -12; di <= 12; di++ {
						ii, jj := i+di, j+dj
						if ii < 0 || jj < 0 || ii >= N || jj >= N || !rel[jj*N+ii] {
							continue
						}
						if d := math.Hypot(float64(di), float64(dj)); d < best {
							best = d
						}
					}
				}
			}
			dist[j*N+i] = float32(best)
		}
	}
	return func(x, y float64) float64 {
		i, j := int(math.Floor(x+0.5))+O, int(math.Floor(y+0.5))+O
		if i < 0 || j < 0 || i >= N || j >= N {
			return 12
		}
		return float64(dist[j*N+i])
	}
}

// toning lays silver toning on the piece in whole ramp steps: a crescent along
// part of the rim, a halo hugging the design, or one side in shade; one copy
// in four has none. The level sets the depth.
func toning(sp *spec, kind, l int, seed uint32) {
	lv := []float64{0, 2.3, 1.9, 1.5, 1.1, 0.8, 0}[l]
	k := int(rnd(seed, 60) * 4) // 0 crescent, 1 halo, 2 shade, 3 none
	if lv == 0 || k == 3 {
		return
	}
	sp.tonR = toneSilver
	if kind == kindMedieval {
		sp.tonR = tonePewter
	}
	str := 0.8 + 0.4*rnd(seed, 61)
	a0 := rnd(seed, 62) * 2 * math.Pi
	hw := 0.7 + 0.6*rnd(seed, 63)
	f := sp.face
	var near func(x, y float64) float64
	si := int(seed)
	sp.tonF = func(x, y, d float64, px, py int) int {
		var a float64
		switch k {
		case 0:
			dl := math.Abs(wrap(math.Atan2(y, x) - a0))
			aw := 1 - smooth(hw*0.55, hw, dl)
			rb := 1 - smooth(2.5, 12, d)
			a = aw * (0.35 + 0.85*rb) * (0.85 + 0.3*fbm(x, y, 9, si+71))
			if rb < 0.05 {
				a = 0
			}
		case 1:
			if near == nil {
				near = relief(f)
			}
			nd := near(x, y)
			if nd == 0 {
				return 0
			}
			a = (1 - smooth(1.5, 9, nd)) * (0.8 + 0.4*fbm(x, y, 8, si+72))
			a += 0.5 * (1 - smooth(2, 6, d)) // and a little along the rim
		case 2:
			pr := (x*math.Cos(a0) + y*math.Sin(a0)) / 50
			a = smooth(0.05, 0.85, pr+0.18*(fbm(x, y, 40, si+73)-0.5))
		}
		m := a * lv * str
		i := math.Floor(m)
		fr := m - i
		switch {
		case fr > 0.7:
			i++
		case fr >= 0.3 && (fr-0.3)/0.4 > bayer[py&3][px&3]:
			i++
		}
		if i > 3 {
			i = 3
		}
		return int(i)
	}
}

// vary turns the reference spec of a hand-struck age into one copy's own: the
// blank's outline turned and its radius drifting in lobes, the strike off
// centre and turned, one copy in three on a second ramp, the high points worn
// where the seed says, and toning.
func vary(sp *spec, kind, l int, seed uint32) {
	const R = 54.0
	w := wear(l)
	si := int(seed)
	type lobe struct{ f, ph, a float64 }
	nl := 2 + int(rnd(seed, 2)*3)
	var lobes []lobe
	for i := 0; i < nl; i++ {
		lobes = append(lobes, lobe{float64(2 + int(rnd(seed, 10+i)*4)), rnd(seed, 20+i) * 2 * math.Pi,
			0.04 * R / float64(nl) * (0.55 + 0.45*rnd(seed, 30+i))})
	}
	turn := rnd(seed, 1) * 2 * math.Pi
	clip := kind == kindMedieval && rnd(seed, 3) < 0.334
	ca := rnd(seed, 4) * 2 * math.Pi
	cd := 52.5 + 1.5*rnd(seed, 5)
	d0 := sp.sd
	sp.sd = func(x, y float64) float64 {
		rx, ry := rotate(x, y, turn)
		d := d0(rx, ry)
		a := math.Atan2(y, x)
		for _, lb := range lobes {
			d -= lb.a * math.Sin(lb.f*a+lb.ph)
		}
		if clip {
			d = math.Max(d, x*math.Cos(ca)+y*math.Sin(ca)-cd)
		}
		return d
	}
	off := 0.06 * R * (0.25 + 0.75*rnd(seed, 6))
	ad := rnd(seed, 7) * 2 * math.Pi
	ox, oy := off*math.Cos(ad), off*math.Sin(ad)
	rot := 12 * math.Pi / 180 * rs(seed, 8)
	f0 := sp.face
	sp.face = func(x, y float64) sample {
		u, v := rotate(x-ox, y-oy, rot)
		return f0(u, v)
	}
	alt := tonedClassical
	if kind == kindMedieval {
		alt = tonedMedieval
	}
	if rnd(seed, 9) < 0.334 {
		for i := range sp.ramp {
			if sp.ramp[i] == silver {
				sp.ramp[i] = alt
			}
		}
		if sp.edgeR != nil && *sp.edgeR == silver {
			sp.edgeR = &alt
		}
	}
	f1 := sp.face
	sp.face = func(x, y float64) sample {
		st := f1(x, y)
		if w > 0 && st.h > ground {
			m := smooth(0.5, 0.72, fbm(x, y, 14, si+21))
			st.h = ground + (st.h-ground)*(1-0.8*w*m)
		}
		return st
	}
	toning(sp, kind, l, seed)
}
