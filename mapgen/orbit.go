package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Orbit is the first space scale: the home planet seen from straight above, a
// small disc in the middle with its land and sea drawn from the world, and
// rings of orbit round it with stations standing on them, joined by one space
// elevator. Everything else is the engine as it was: the same parcels, the
// same stable placement, the same roads, joins, traffic and light.

var orbitSpec = ringSpec{
	cx: worldW / 2, cy: worldH / 2, hub: 20, rings: []int{24, 29, 34, 39},
	lotW: 5, lotH: 3, step: 6, spread: [2]int{30, 10},
}

// the home planet's disc: its radii in cells (ratio 50/20 = 2.5 is a perfect circle in terminal fonts)
const discRX, discRY = 50, 20

var quadrantRunes = [16]rune{
	' ', '▘', '▝', '▀',
	'▖', '▌', '▞', '▛',
	'▗', '▚', '▐', '▜',
	'▄', '▙', '▟', '█',
}

// asinTable maps index i in [0..10] (representing fraction i/10 from 0.0 to 1.0)
// to asin(i/10) / (pi/2) * 1000 in [0..1000].
var asinTable = [11]int{0, 64, 128, 194, 262, 333, 410, 494, 590, 713, 1000}

// asinFrac returns the spherical projection fraction in [-1000, 1000] for v in [-maxV, maxV].
func asinFrac(v, maxV int) int {
	if maxV <= 0 || v == 0 {
		return 0
	}
	neg := false
	if v < 0 {
		neg = true
		v = -v
	}
	if v >= maxV {
		if neg {
			return -1000
		}
		return 1000
	}
	s := v * 1000 / maxV
	idx := s / 100
	rem := s % 100
	val := asinTable[idx] + (asinTable[idx+1]-asinTable[idx])*rem/100
	if neg {
		return -val
	}
	return val
}

// noisePeriodic samples 2D value noise where x wraps with periodX lattice cells.
func noisePeriodic(seed int64, x, y, periodX int) int {
	p256 := periodX * 256
	x = ((x % p256) + p256) % p256
	ix0 := (x >> 8) % periodX
	ix1 := (ix0 + 1) % periodX
	iy := y >> 8
	fx := smooth(x & 255)
	fy := smooth(y & 255)
	top := lerp(lattice(seed, ix0, iy), lattice(seed, ix1, iy), fx)
	bot := lerp(lattice(seed, ix0, iy+1), lattice(seed, ix1, iy+1), fx)
	return lerp(top, bot, fy)
}

// fbmPeriodic samples multi-octave fractal noise periodic in lambda ∈ [0, 4000).
func fbmPeriodic(seed int64, lambda, phi, basePeriodX, scaleY, octaves int) int {
	lambda = ((lambda % 4000) + 4000) % 4000
	if phi < -1000 {
		phi = -1000
	} else if phi > 1000 {
		phi = 1000
	}
	total, amp, maxAmp := 0, 512, 0
	periodX := basePeriodX
	freqY := scaleY
	for i := 0; i < octaves; i++ {
		x := (lambda * periodX * 256) / 4000
		y := ((phi + 1000) * freqY * 256) / 2000
		total += noisePeriodic(seed+int64(i)*101, x, y, periodX) * amp
		maxAmp += amp
		amp >>= 1
		periodX <<= 1
		freqY <<= 1
	}
	return total / maxAmp
}

// NewOrbitPlan lays out the elevator, the rings, the wonder's plot and the
// station parcels, and hands the parcels to the building lines.
func NewOrbitPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newRingPlan(seed, orbitSpec, Orbit)
	fixed := make([]int, 0, len(b.plan.Roads)-1)
	for i := 1; i < len(b.plan.Roads); i++ {
		fixed = append(fixed, i)
	}
	b.plan.Fixed = fixed
	b.plotBesideRing(17, len(orbitSpec.rings)-1)
	b.cutRingLots(seed, cat, orbitSpec)
	return b.plan
}

// rim is 1000 on the edge of the ellipse of radii rx, ry cells round
// (cx, cy), less inside it and more outside.
func rim(x, y, cx, cy, rx, ry int) int {
	dx, dy := x-cx, y-cy
	return (dx*dx*ry*ry + dy*dy*rx*rx) * 1000 / (rx * rx * ry * ry)
}

// paintHomePlanet draws the rotating 3D spherical Earth. The globe rotates from
// West to East, showing continents, mountain ridges, biomes, and oceans with
// 2x2 subpixel quadrant resolution, differential cloud decks with Coriolis shear
// and cyclonic storms, twilight Rayleigh scattering, ocean specular sun glint,
// and night-side megacities rotating with the land.
func (f *frame) paintHomePlanet() {
	c, seed := f.plan.Centre, f.m.Seed
	subRY := 4 * discRY
	subRX := 4 * discRX
	rX2 := subRX * subRX
	rY2 := subRY * subRY
	rMax2 := rX2 * rY2

	subOffsets := [4][2]int{
		{1, 1}, // bit 0 (1): TL
		{3, 1}, // bit 1 (2): TR
		{1, 3}, // bit 2 (4): BL
		{3, 3}, // bit 3 (8): BR
	}

	groundShift := f.anim * 4
	cloudShift := f.anim * 7

	type stormCenter struct {
		lon, lat int
	}
	storms := [3]stormCenter{
		{lon: 900, lat: 360},
		{lon: 2700, lat: -400},
		{lon: 1900, lat: 180},
	}

	glintX := -subRX / 3
	glintY := -subRY / 5

	for y := c.Y - discRY; y <= c.Y+discRY; y++ {
		dyCell := y - c.Y
		vCell := (dyCell * 1000) / discRY
		for x := c.X - discRX; x <= c.X+discRX; x++ {
			dxCell := x - c.X

			// 1. Compute subpixel disc mask
			mask := 0
			for q := 0; q < 4; q++ {
				subX := subOffsets[q][0]
				subY := subOffsets[q][1]
				dx := 4*dxCell + subX - 2
				dy := 4*dyCell + subY - 2
				if dx*dx*rY2+dy*dy*rX2 <= rMax2 {
					mask |= (1 << q)
				}
			}

			if mask == 0 {
				continue
			}

			// If 1..3 subpixels inside: Outer Atmospheric Rim!
			if mask < 15 {
				spaceBg := f.g.At(x, y).Bg
				night := clamp((dxCell+discRX)*120/(2*discRX), 0, 120)
				rimCell := Cell{
					R:   quadrantRunes[mask],
					Fg:  P(HCyan, 255),
					Bg:  spaceBg,
					Dim: uint8(night),
				}
				// Polar auroras dancing at the limb
				if abs(vCell) > 840 {
					urot := (dxCell*1000)/discRX + (vCell*f.anim*2)/1000 + f.anim*3
					aurora := fbm(seed+77, (urot*256)/50+f.anim*3, (vCell*256)/30, 2)
					if aurora > 550 {
						rimCell.Fg = P(HForest, 240)
						rimCell.Dim = 0
					}
				}
				f.g.paint(x, y, rimCell)
				f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "home"})
				continue
			}

			// All 4 subpixels inside: Planet Interior!
			var (
				landMask      int
				cloudMask     int
				coreCloudMask int
				wispMask      int
				polarMask     int
				mtnMask       int
				forestMask    int
				shelfMask     int
				subElev       [4]int
				subMoist      [4]int
			)

			for q := 0; q < 4; q++ {
				subX := subOffsets[q][0]
				subY := subOffsets[q][1]
				dx := 4*dxCell + subX - 2
				dy := 4*dyCell + subY - 2
				v := (dy * 1000) / subRY
				phi := asinFrac(dy, subRY)
				rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
				lambda0 := asinFrac(dx, rxLatSub)

				// Ground rotation (West to East)
				lambdaG := ((lambda0-groundShift)%4000 + 4000) % 4000

				// Cloud rotation with Coriolis shear
				coriolis := (v*v - 350000) / 3000
				lambdaC := ((lambda0-cloudShift+coriolis)%4000 + 4000) % 4000

				// Procedural terrain:
				warpL := (fbmPeriodic(seed+11, lambdaG, phi, 4, 2, 2) - 512) * 120 / 512
				warpP := (fbmPeriodic(seed+12, lambdaG, phi, 4, 2, 2) - 512) * 80 / 512
				phiW := clamp(phi+warpP, -1000, 1000)
				elev := fbmPeriodic(seed+5, lambdaG+warpL, phiW, 6, 3, 4)
				ridgeVal := fbmPeriodic(seed+19, lambdaG, phi, 12, 6, 3)*2 - 1000
				if ridgeVal < 0 {
					ridgeVal = -ridgeVal
				}
				ridge := 1000 - ridgeVal
				moisture := fbmPeriodic(seed+41, lambdaG, phi, 8, 4, 3)
				subElev[q] = elev
				subMoist[q] = moisture

				bit := 1 << q
				isLand := elev > 515
				if isLand {
					landMask |= bit
					if abs(phi)+(elev-500)/3 > 840 {
						polarMask |= bit
					} else {
						if elev > 725 || (ridge > 875 && elev > 605) {
							mtnMask |= bit
						}
						if moisture > 510 && abs(phi) < 680 {
							forestMask |= bit
						}
					}
				} else {
					// Ocean
					if abs(phi) > 855 {
						polarMask |= bit
					}
					ddx := dx - glintX
					ddy := (dy - glintY) * 5 / 2
					isGlint := (ddx*ddx + ddy*ddy) < 160
					if elev > 480 || isGlint {
						shelfMask |= bit
					}
				}

				// Clouds and cyclonic storms:
				swirlL := 0
				swirlP := 0
				eyeBoost := 0
				for _, st := range storms {
					dLam := abs(lambdaC - st.lon)
					if dLam > 2000 {
						dLam = 4000 - dLam
					}
					dPhi := phi - st.lat
					dist2 := (dLam*dLam)/4 + dPhi*dPhi
					if dist2 < 90000 {
						r := isqrt(dist2)
						swirl := (300 - r) * 160 / 300
						if st.lat > 0 {
							swirlL += swirl * dPhi / 300
							swirlP -= swirl * dLam / 300
						} else {
							swirlL -= swirl * dPhi / 300
							swirlP += swirl * dLam / 300
						}
						if dist2 < 40000 && dist2 > 1600 {
							eyeBoost += (40000 - dist2) * 160 / 40000
						} else if dist2 <= 1600 {
							eyeBoost -= 150
						}
					}
				}

				cDensity := fbmPeriodic(seed+89, lambdaC+swirlL, phi+swirlP, 8, 4, 3) + eyeBoost
				if cDensity > 675 {
					coreCloudMask |= bit
				}
				if cDensity > 605 {
					cloudMask |= bit
				}
				if cDensity > 550 {
					wispMask |= bit
				}
			}

			rxLatCell := isqrt(max(1, discRX*discRX-(dyCell*dyCell*discRX*discRX)/(discRY*discRY)))
			l0Cell := asinFrac(dxCell, rxLatCell)
			phiCell := asinFrac(dyCell, discRY)
			lgCell := ((l0Cell-groundShift)%4000 + 4000) % 4000
			rotHash := hash(seed, 23, int64(lgCell/16), int64(phiCell/16))

			// Surface classification
			var baseCell Cell
			var underlyingBg Paint

			switch {
			case landMask > 0 && landMask < 15:
				// Coastline: sub-pixel quadrant blocks
				underlyingBg = P(HSea, 255)
				baseCell = Cell{
					R:  quadrantRunes[landMask],
					Fg: P(HLand, 255),
					Bg: P(HSea, 255),
				}

			case landMask == 15:
				// Pure land
				underlyingBg = P(HLand, 255)
				avgElev := (subElev[0] + subElev[1] + subElev[2] + subElev[3]) / 4
				switch {
				case polarMask == 15:
					underlyingBg = P(HSea, 240)
					glyph := ' '
					if (rotHash>>8)%5 == 0 {
						glyph = '*'
					}
					baseCell = Cell{R: glyph, Fg: P(HWhite, 240), Bg: P(HCloud, 210)}
					if abs(vCell) > 840 {
						aurora := fbm(seed+77, (dxCell*256)/50+f.anim*3, (vCell*256)/30, 2)
						if aurora > 570 {
							baseCell.Fg = P(HForest, 240)
						}
					}
				case polarMask > 0:
					baseCell = Cell{R: quadrantRunes[polarMask], Fg: P(HCloud, 220), Bg: P(HLand, 240)}
				case mtnMask == 15:
					glyph := ' '
					fg := P(HCloud, 230)
					if avgElev > 765 && rotHash%3 == 0 {
						glyph = '▲'
						fg = P(HWhite, 250)
					} else if rotHash%4 == 0 {
						glyph = '^'
					}
					baseCell = Cell{R: glyph, Fg: fg, Bg: P(HLand, 195)}
					underlyingBg = P(HLand, 195)
				case mtnMask > 0:
					baseCell = Cell{R: quadrantRunes[mtnMask], Fg: P(HLand, 185), Bg: P(HLand, 245)}
					underlyingBg = P(HLand, 240)
				case forestMask > 0 && forestMask < 15:
					baseCell = Cell{R: quadrantRunes[forestMask], Fg: P(HForest, 235), Bg: P(HLand, 245)}
					underlyingBg = P(HLand, 245)
				case forestMask == 15:
					glyph := ' '
					if rotHash%7 == 0 {
						glyph = '♣'
					}
					baseCell = Cell{R: glyph, Fg: P(HForest, 230), Bg: P(HLand, 255)}
					underlyingBg = P(HLand, 255)
				default:
					avgMoist := (subMoist[0] + subMoist[1] + subMoist[2] + subMoist[3]) / 4
					if avgMoist <= 410 && abs(vCell) < 550 {
						baseCell = Cell{R: ' ', Fg: P(HGold, 210), Bg: P(HLand, 235)}
					} else {
						baseCell = Cell{R: ' ', Fg: P(HLand, 210), Bg: P(HLand, 255)}
					}
					underlyingBg = P(HLand, 255)
				}

			default:
				// Pure ocean (landMask == 0)
				underlyingBg = P(HSea, 255)
				switch {
				case polarMask == 15:
					underlyingBg = P(HSea, 240)
					glyph := ' '
					if (rotHash>>8)%5 == 0 {
						glyph = '*'
					}
					baseCell = Cell{R: glyph, Fg: P(HWhite, 240), Bg: P(HCloud, 200)}
				case polarMask > 0:
					baseCell = Cell{R: quadrantRunes[polarMask], Fg: P(HCloud, 210), Bg: P(HSea, 255)}
				case shelfMask > 0 && shelfMask < 15:
					baseCell = Cell{R: quadrantRunes[shelfMask], Fg: P(HCyan, 185), Bg: P(HSea, 255)}
				case shelfMask == 15:
					underlyingBg = P(HSea, 230)
					baseCell = Cell{R: '~', Fg: P(HCyan, 190), Bg: P(HSea, 230)}
				default:
					glyph := ' '
					if rotHash%100 < 18 {
						glyph = '≈'
					} else if rotHash%100 < 40 {
						glyph = '~'
					}
					baseCell = Cell{R: glyph, Fg: P(HWater, 150), Bg: P(HSea, 255)}
				}
			}

			// Clouds: sculpted with 3-tier 2x2 subpixel quadrant shading (no solid white blocks or yellow beams)
			var cell Cell
			switch {
			case coreCloudMask == 15:
				cell = Cell{R: ' ', Fg: P(HCloud, 255), Bg: P(HCloud, 255)}
			case coreCloudMask > 0:
				cell = Cell{R: quadrantRunes[coreCloudMask], Fg: P(HCloud, 255), Bg: P(HCloud, 195)}
			case cloudMask == 15:
				cell = Cell{R: ' ', Fg: P(HCloud, 195), Bg: P(HCloud, 195)}
			case cloudMask > 0:
				cell = Cell{R: quadrantRunes[cloudMask], Fg: P(HCloud, 215), Bg: underlyingBg}
			case wispMask > 0:
				cell = Cell{R: quadrantRunes[wispMask], Fg: P(HCloud, 145), Bg: underlyingBg}
			default:
				cell = baseCell
			}

			// Day/Night terminator and limb darkening
			rVal := rim(x, y, c.X, c.Y, discRX, discRY)
			limb := clamp((rVal-700)*256/300, 0, 180)
			night := clamp((dxCell+discRX/3)*215/(discRX+discRX/3), 0, 215)
			cell.Dim = dimmer(uint8(night), uint8(limb))

			// Night-Side Megacities (Rotate with continents)
			if dxCell > 4 && abs(vCell) < 750 && cloudMask < 12 {

				isCity := false
				if landMask > 0 {
					cluster := fbmPeriodic(seed+33, lgCell, phiCell, 16, 8, 3)
					if landMask < 15 {
						// Coastal ports tracing continental coastlines
						if cluster > 430 && chance(65, seed, 16, int64(lgCell/15), int64(phiCell/15)) {
							isCity = true
						}
					} else if mtnMask == 0 {
						// Inland metropolises
						if cluster > 470 && chance(60, seed, 16, int64(lgCell/15), int64(phiCell/15)) {
							isCity = true
						}
					}
				}
				if isCity {
					hCity := hash(seed, 16, int64(lgCell/15), int64(phiCell/15), int64(f.anim/3))
					cityGlyphs := []rune{'·', '∙', '*'}
					cell.R = cityGlyphs[hCity%uint64(len(cityGlyphs))]
					cell.Fg = P(HGold, 255)
					cell.Dim = 0
					f.lights = append(f.lights, Pt{X: x, Y: y})
				}
			}

			f.g.paint(x, y, cell)
			f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "home"})
		}
	}
}

// paintOrbitStructures draws climber cars along the elevator tether and satellites along the rings.
func (f *frame) paintOrbitStructures() {
	// 1. Space elevator tether structures and climber cars:
	if len(f.plan.Roads) > 1 {
		spoke := f.plan.Roads[1]
		n := len(spoke.Cells)
		if n > 0 {
			// Top of tether: Geostationary Counterweight Anchor
			top := spoke.Cells[n-1]
			canAnchor := true
			for dx := -2; dx <= 2; dx++ {
				p := Pt{X: top.X + dx, Y: top.Y}
				if !f.g.in(p.X, p.Y) || f.g.ThingAt(p.X, p.Y).Kind == KindBuilding || f.g.ThingAt(p.X, p.Y).Kind == KindWonder {
					canAnchor = false
					break
				}
				if dx != 0 && f.busy[at(p.X, p.Y)] {
					canAnchor = false
					break
				}
			}
			if canAnchor {
				f.g.draw(top.X-2, top.Y, '▤', P(HCyan, 220))
				f.g.draw(top.X-1, top.Y, '═', P(HWhite, 240))
				f.g.draw(top.X, top.Y, '╬', P(HWhite, 240))
				f.g.draw(top.X+1, top.Y, '═', P(HWhite, 240))
				f.g.draw(top.X+2, top.Y, '▤', P(HCyan, 220))
			}

			// Pod 1: climbing up (▲)
			upIdx := (f.anim / 2) % n
			p1 := spoke.Cells[upIdx]
			if f.g.ThingAt(p1.X, p1.Y).Kind != KindBuilding && f.g.ThingAt(p1.X, p1.Y).Kind != KindWonder {
				f.g.draw(p1.X, p1.Y, '▲', P(HEngine, 255))
			}

			// Pod 2: descending down (▼)
			downIdx := (n - 1) - ((f.anim/2 + n/2) % n)
			p2 := spoke.Cells[downIdx]
			if f.g.ThingAt(p2.X, p2.Y).Kind != KindBuilding && f.g.ThingAt(p2.X, p2.Y).Kind != KindWonder {
				f.g.draw(p2.X, p2.Y, '▼', P(HEngine, 255))
			}

			// Pod 3: cargo car (■)
			cargoIdx := (f.anim/2 + n/3) % n
			p3 := spoke.Cells[cargoIdx]
			if f.g.ThingAt(p3.X, p3.Y).Kind != KindBuilding && f.g.ThingAt(p3.X, p3.Y).Kind != KindWonder {
				f.g.draw(p3.X, p3.Y, '■', P(HEngine, 255))
			}

			// Base beacon at Earth's north pole
			base := spoke.Cells[0]
			if f.g.ThingAt(base.X, base.Y).Kind != KindBuilding {
				f.g.draw(base.X, base.Y, '◈', P(HCyan, 255))
			}
		}
	}

	// 2. Orbital satellites & communication probes in orbital sectors:
	// draw travelling probes along the rings outside parcel lots.
	for roadIdx := 2; roadIdx < len(f.plan.Roads); roadIdx++ {
		road := f.plan.Roads[roadIdx]
		n := len(road.Cells)
		if n == 0 {
			continue
		}
		// Travelling probe on each ring half:
		probeIdx := (f.anim + roadIdx*17) % n
		pp := road.Cells[probeIdx]
		if f.g.ThingAt(pp.X, pp.Y).Kind != KindBuilding && f.g.ThingAt(pp.X, pp.Y).Kind != KindWonder {
			f.g.draw(pp.X, pp.Y, '▫', P(HCyan, 240))
		}
	}
}

// relicSpot is where the wonder of a ground age shines on the home planet.
func (f *frame) orbitRelic(age, try int) Pt {
	c := f.plan.Centre
	for attempt := 0; attempt < 10; attempt++ {
		p := Pt{X: c.X - discRX + 2 + roll(2*discRX-3, f.m.Seed, 17, int64(age), int64(try*10+attempt)), Y: c.Y - discRY + 2 + roll(2*discRY-3, f.m.Seed, 18, int64(age), int64(try*10+attempt))}

		if rim(p.X, p.Y, c.X, c.Y, discRX, discRY) <= 800 {
			return p
		}
	}
	return Pt{X: c.X, Y: c.Y}
}
