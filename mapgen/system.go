package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// The solar system is the next scale out: the sun in the middle, the planets
// on their orbits, the lanes between them. Parcels are marks along the
// orbits, in the order the people reached them; the planets stand on the
// orbits among them.

const sunRX, sunRY = 13, 5

var systemSpec = ringSpec{
	cx: SystemW / 2, cy: SystemH / 2, hub: 7, rings: []int{14, 22, 31, 41, 59, 68, 76, 83},
	lotW: 3, lotH: 1, step: 6, spread: [2]int{30, 10},
}

// homeRing is the orbit of the home world, the third from the sun.
const homeRing = 2

// NewSystemPlan lays out the lanes, a planet on each, the wonder's ground and
// the marks along the orbits.
func NewSystemPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newRingPlan(seed, systemSpec, System)
	p := b.plan
	fixed := make([]int, 0, len(p.Roads)-1)
	for i := 1; i < len(p.Roads); i++ {
		fixed = append(fixed, i)
	}
	p.Fixed = fixed
	for ring := range systemSpec.rings {
		cells := p.Roads[2+2*ring+roll(2, seed, 30, int64(ring))].Cells
		p.Bodies = append(p.Bodies, cells[len(cells)/8+roll(len(cells)*3/4, seed, 31, int64(ring))])
	}
	for _, body := range p.Bodies {
		b.take(Rect{body.X, body.Y, 1, 1})
	}
	b.plotBesideRing(18, len(systemSpec.rings)-1)
	b.cutRingLots(seed, cat, systemSpec)
	return p
}

// paintSystem draws the blazing high-res 3D rotating Sun and solar corona.
func (f *frame) paintSystem() {
	c, seed := f.plan.Centre, f.m.Seed
	subRY := 4 * sunRY
	subRX := 4 * sunRX
	rX2 := subRX * subRX
	rY2 := subRY * subRY
	rMax2 := rX2 * rY2

	subOffsets := [4][2]int{
		{1, 1}, // bit 0 (1): TL
		{3, 1}, // bit 1 (2): TR
		{1, 3}, // bit 2 (4): BL
		{3, 3}, // bit 3 (8): BR
	}

	for y := c.Y - sunRY - 3; y <= c.Y+sunRY+3; y++ {
		dyCell := y - c.Y
		for x := c.X - sunRX - 6; x <= c.X+sunRX+6; x++ {
			if !f.g.in(x, y) {
				continue
			}
			dxCell := x - c.X

			// 1. Compute 4-bit subpixel disc mask
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

			spaceBg := f.g.At(x, y).Bg

			// Outside the disc: Coronal Halo
			if mask == 0 {
				if rim(x, y, c.X, c.Y, sunRX+5, sunRY+3) <= 1000 {
					flare := fbm(seed+311, (dxCell*256)/8+f.anim, (dyCell*256)/4-f.anim, 2)
					if flare > 440 {
						r := '·'
						fg := P(HAmber, 240)
						if flare > 610 {
							r = '∙'
							fg = P(HWhite, 255)
						} else if flare > 520 {
							r = '·'
							fg = P(HSun, 255)
						}
						f.g.paint(x, y, Cell{R: r, Fg: fg, Bg: spaceBg})
					}
				}
				continue
			}

			// Solar Rim (0 < mask < 15)
			if mask < 15 {
				f.g.paint(x, y, Cell{
					R:  quadrantRunes[mask],
					Fg: P(HSun, 255),
					Bg: spaceBg,
				})
				f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "sun"})
				continue
			}

			// Solar Interior (mask == 15)
			var (
				coreMask   int
				photoMask  int
				plasmaMask int
			)
			for q := 0; q < 4; q++ {
				subX := subOffsets[q][0]
				subY := subOffsets[q][1]
				dx := 4*dxCell + subX - 2
				dy := 4*dyCell + subY - 2

				phi := asinFrac(dy, subRY)
				rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
				lambda0 := asinFrac(dx, rxLatSub)

				solarShift := f.anim * 2
				lambdaS := ((lambda0-solarShift)%4000 + 4000) % 4000

				plasma := fbmPeriodic(seed+301, lambdaS, phi, 6, 3, 3)

				// Distance fraction from core (0 at center, 1000 at edge)
				dFrac := (dx*dx*rY2 + dy*dy*rX2) * 1000 / rMax2

				bit := 1 << q
				if dFrac < 250 || (dFrac < 520 && plasma > 520) || plasma > 670 {
					coreMask |= bit
				}
				if plasma > 440 || dFrac < 680 {
					photoMask |= bit
				}
				if plasma > 330 {
					plasmaMask |= bit
				}
			}

			var cell Cell
			switch {
			case coreMask == 15:
				cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: P(HWhite, 255)}
			case coreMask > 0:
				cell = Cell{R: quadrantRunes[coreMask], Fg: P(HWhite, 255), Bg: P(HSun, 255)}
			case photoMask == 15:
				cell = Cell{R: '█', Fg: P(HSun, 255), Bg: P(HGold, 255)}
			case photoMask > 0:
				cell = Cell{R: quadrantRunes[photoMask], Fg: P(HSun, 255), Bg: P(HAmber, 255)}
			case plasmaMask == 15:
				cell = Cell{R: '█', Fg: P(HAmber, 255), Bg: P(HGold, 255)}
			case plasmaMask > 0:
				cell = Cell{R: quadrantRunes[plasmaMask], Fg: P(HAmber, 255), Bg: P(HFire, 255)}
			default:
				cell = Cell{R: '█', Fg: P(HAmber, 255), Bg: P(HFire, 255)}
			}

			f.g.paint(x, y, cell)
			f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "sun"})
		}
	}
}

// paintBodies draws what stands on the lanes of a space scale over them: the
// planets of the solar system, the systems of the star map.
func (f *frame) paintBodies() {
	switch f.st.Scale {
	case Orbit:
		f.paintOrbitStructures()
	case System:
		f.paintSystemBodies()
	case Stars:
		f.paintTerritory()
		n, glyph := f.plan.Centre, []rune(f.st.Hub[0])[0]
		f.g.draw(n.X, n.Y, glyph, f.ink(glyph, HNone))
		f.g.mark(n.X, n.Y, Thing{Kind: KindTerrain, Key: "home"})
	}
}

// paintSystemBodies draws the full 8-planet 3D Solar System, planetary rings,
// asteroid belt, and interplanetary transfer corridors with active spacecraft.
func (f *frame) paintSystemBodies() {
	c := f.plan.Centre

	// 1. Compute Live Orbital Loops & Positions for All 8 Planets
	var planetPos [8]Pt
	divs := [8]int{12, 20, 32, 50, 80, 120, 180, 260}
	for i := 0; i < len(systemSpec.rings) && i < 8; i++ {
		east := f.plan.Roads[2+2*i].Cells
		west := f.plan.Roads[2+2*i+1].Cells
		loop := make([]Pt, 0, len(east)+len(west)+1)
		loop = append(loop, Pt{c.X, c.Y - systemSpec.rings[i]})
		loop = append(loop, east...)
		for j := len(west) - 1; j >= 0; j-- {
			loop = append(loop, west[j])
		}
		if len(loop) == 0 {
			continue
		}

		// Golden-angle sector distribution around Sol:
		// Disperses the planets across different quadrants (360 degrees) so they never bunch together:
		phaseFrac := (int(f.m.Seed%1000) + i*382) % 1000
		startIdx := (phaseFrac * len(loop)) / 1000

		advance := f.anim / divs[i]
		pos := loop[(startIdx+advance)%len(loop)]
		planetPos[i] = pos

	}

	// 3. Main Asteroid Belt (Between Mars Ring 3 ry=41 and Jupiter Ring 4 ry=59)
	beltRX, beltRY := 118, 50
	for y := c.Y - beltRY - 2; y <= c.Y+beltRY+2; y++ {
		for x := c.X - beltRX - 3; x <= c.X+beltRX+3; x++ {
			if !f.g.in(x, y) {
				continue
			}
			rimVal := rim(x, y, c.X, c.Y, beltRX, beltRY)
			if rimVal >= 940 && rimVal <= 1060 {
				th := f.g.ThingAt(x, y)
				if th.Kind != KindNone {
					continue
				}
				if !chance(18, f.m.Seed+808, int64(x), int64(y), 0) {
					continue
				}
				pick := roll(3, f.m.Seed+909, int64(x), int64(y))
				var r rune
				var fg Paint
				switch pick {
				case 0:
					r, fg = '·', P(HRock, 195)
				case 1:
					r, fg = '∙', P(HSand, 180)
				default:
					r, fg = '•', P(HMetal, 200)
				}
				spaceBg := f.g.At(x, y).Bg
				f.g.paint(x, y, Cell{R: r, Fg: fg, Bg: spaceBg})
				f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "planet"})
			}
		}
	}

	// 8. Ships between the planets, in quadrant pixels. They
	// keep clear of the bodies (the radii are cells of clearance) and of Sol,
	// and of anything else already drawn.
	clearance := [8]int{8, 9, 12, 10, 22, 25, 10, 12}
	pos := planetPos[:]
	bodies := make([]shipBody, len(pos))
	for i, p := range pos {
		bodies[i] = shipBody{(2*p.X + 1) * pxUnit, (2*p.Y + 1) * pxUnit, 32 * clearance[i]}
	}
	pix := NewPixels(f.g.W, f.g.H)
	pix.Ships(f.m.Seed, f.anim, bodies, nil, func(px, py int) bool {
		x, y := px/2, py/2
		if !f.g.in(x, y) || rim(x, y, c.X, c.Y, sunRX+5, sunRY+3) <= 1000 {
			return true
		}
		th := f.g.ThingAt(x, y)
		return th.Kind == KindWonder || th.Kind == KindBuilding || th.Key == "sun" || th.Key == "planet" || th.Key == "home"
	})
	pix.Resolve(f.g)
	pix.Release()

	// 4. Planet Rendering Helpers
	paintSatellite := func(sx, sy int, r rune, fg Paint, key string) {
		if !f.g.in(sx, sy) {
			return
		}
		th := f.g.ThingAt(sx, sy)
		if th.Kind == KindWonder || th.Key == "sun" || th.Key == "square" {
			return
		}
		spaceBg := f.g.At(sx, sy).Bg
		f.g.paint(sx, sy, Cell{R: r, Fg: fg, Bg: spaceBg})
		f.g.mark(sx, sy, Thing{Kind: KindTerrain, Key: key})
	}

	paintSaturnRings := func(px, py int, isBack bool) {
		rxRing, ryRing := 18, 5
		subOffsets := [4][2]int{
			{1, 1}, // bit 0 (1): TL
			{3, 1}, // bit 1 (2): TR
			{1, 3}, // bit 2 (4): BL
			{3, 3}, // bit 3 (8): BR
		}

		subRXRing := 4 * rxRing
		subRYRing := 4 * ryRing

		for dyCell := -ryRing - 1; dyCell <= ryRing+1; dyCell++ {
			y := py + dyCell
			for dxCell := -rxRing - 1; dxCell <= rxRing+1; dxCell++ {
				x := px + dxCell
				if !f.g.in(x, y) {
					continue
				}
				th := f.g.ThingAt(x, y)
				if th.Kind == KindWonder || th.Key == "sun" || th.Key == "square" {
					continue
				}

				// Check whether this cell is on the back or front of Saturn's tilted ring plane
				// Tilted line: v = 4*dyCell - dxCell/2
				vRough := 4*dyCell - dxCell/2
				if isBack && vRough > 2 {
					continue
				}
				if !isBack && vRough < -2 {
					continue
				}

				// Check if inside Saturn's globe (rx=8, ry=3)
				dxG := 4 * dxCell
				dyG := 4 * dyCell
				inGlobe := dxG*dxG*(4*3)*(4*3)+dyG*dyG*(4*8)*(4*8) <= (4*8)*(4*8)*(4*3)*(4*3)
				if isBack && inGlobe {
					// Ring is occluded behind Saturn's globe
					continue
				}

				ringMask := 0
				ringBMask := 0
				ringAMask := 0
				crepeMask := 0
				cassiniMask := 0

				for q := 0; q < 4; q++ {
					subX := subOffsets[q][0]
					subY := subOffsets[q][1]
					dx := 4*dxCell + subX - 2
					dy := 4*dyCell + subY - 2

					u := dx
					v := dy - dx/8 // 3D ring tilt

					if isBack && v > 0 {
						continue
					}
					if !isBack && v <= 0 {
						continue
					}

					uNorm := (u * 1000) / subRXRing
					vNorm := (v * 1000) / subRYRing
					rFrac := isqrt(uNorm*uNorm + vNorm*vNorm)

					bit := 1 << q
					switch {
					case rFrac >= 810 && rFrac <= 1000: // Outer Ring A
						ringMask |= bit
						ringAMask |= bit
					case rFrac > 730 && rFrac < 810: // Cassini Division gap
						cassiniMask |= bit
					case rFrac >= 500 && rFrac <= 730: // Inner Ring B (brightest, richest)
						ringMask |= bit
						ringBMask |= bit
					case rFrac >= 380 && rFrac < 500: // Crepe Ring C
						ringMask |= bit
						crepeMask |= bit
					}
				}

				if ringMask == 0 && cassiniMask == 0 {
					continue
				}

				spaceBg := f.g.At(x, y).Bg
				underBg := spaceBg
				if inGlobe && !isBack {
					underBg = P(HAmber, 220)
				}

				var cell Cell
				switch {
				case ringBMask > 0:
					fg := P(HGold, 255)
					bg := P(HSand, 240)
					if ringBMask < 15 {
						bg = underBg
						cell = Cell{R: quadrantRunes[ringBMask], Fg: fg, Bg: bg}
					} else {
						cell = Cell{R: '█', Fg: fg, Bg: bg}
					}
				case ringAMask > 0:
					fg := P(HSand, 245)
					bg := underBg
					if ringAMask < 15 {
						cell = Cell{R: quadrantRunes[ringAMask], Fg: fg, Bg: bg}
					} else {
						cell = Cell{R: '█', Fg: fg, Bg: P(HGold, 215)}
					}
				case crepeMask > 0:
					fg := P(HAmber, 180)
					bg := underBg
					if crepeMask < 15 {
						cell = Cell{R: quadrantRunes[crepeMask], Fg: fg, Bg: bg}
					} else {
						cell = Cell{R: '█', Fg: fg, Bg: underBg}
					}
				case cassiniMask > 0:
					cell = Cell{R: '·', Fg: P(HSand, 130), Bg: underBg}
				}

				f.g.paint(x, y, cell)
				f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "planet"})
			}
		}
	}

	drawPlanetSphere := func(px, py, rx, ry int, planetIdx int, key string) {
		subOffsets := [4][2]int{
			{1, 1}, // bit 0 (1): TL
			{3, 1}, // bit 1 (2): TR
			{1, 3}, // bit 2 (4): BL
			{3, 3}, // bit 3 (8): BR
		}

		subRX := 4 * rx
		subRY := 4 * ry
		rX2 := subRX * subRX
		rY2 := subRY * subRY
		rMax2 := rX2 * rY2

		sunDX := c.X - px
		sunDY := (c.Y - py) * 24 / 10
		lenSun := isqrt(sunDX*sunDX + sunDY*sunDY)
		sunLX, sunLY := 0, 0
		if lenSun > 0 {
			sunLX = (sunDX * 1000) / lenSun
			sunLY = (sunDY * 1000) / lenSun
		}

		for dyCell := -ry; dyCell <= ry; dyCell++ {
			y := py + dyCell
			for dxCell := -rx; dxCell <= rx; dxCell++ {
				x := px + dxCell
				if !f.g.in(x, y) {
					continue
				}
				th := f.g.ThingAt(x, y)
				if th.Kind == KindWonder || th.Key == "sun" || th.Key == "square" {
					continue
				}

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

				spaceBg := f.g.At(x, y).Bg

				// Shading & lighting at cell center
				dxMid := 4 * dxCell
				dyMid := 4 * dyCell
				nxMid := (dxMid * 1000) / subRX
				nyMid := (dyMid * 1000) / subRY
				dotMid := (nxMid*sunLX + nyMid*sunLY) / 1000
				sunlit := dotMid > -80

				var cell Cell
				switch planetIdx {
				case 0: // Mercury (rx=7, ry=3)
					if mask < 15 {
						fg := P(HRock, 245)
						if !sunlit {
							fg = P(HRock, 190)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						brightMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaM := ((lambda0-f.anim)%4000 + 4000) % 4000
							cNoise := fbmPeriodic(f.m.Seed+501, lambdaM, phi, 8, 4, 3)
							rayNoise := fbmPeriodic(f.m.Seed+502, lambdaM*2, phi*2, 12, 6, 2)
							if (rayNoise > 620 && cNoise > 480) || cNoise > 650 {
								brightMask |= (1 << q)
							}
						}
						if sunlit {
							fg := P(HWhite, 240)
							bg := P(HRock, 245)
							if brightMask == 15 {
								cell = Cell{R: '█', Fg: fg, Bg: P(HSteel, 245)}
							} else if brightMask > 0 {
								cell = Cell{R: quadrantRunes[brightMask], Fg: fg, Bg: bg}
							} else {
								cell = Cell{R: '█', Fg: bg, Bg: P(HMetal, 200)}
							}
						} else {
							fg := P(HSteel, 185)
							bg := P(HRock, 180)
							if brightMask > 0 {
								cell = Cell{R: quadrantRunes[brightMask], Fg: fg, Bg: bg}
							} else {
								cell = Cell{R: '█', Fg: bg, Bg: P(HVoid, 255)}
							}
						}
					}

				case 1: // Venus (rx=8, ry=3)
					if mask < 15 {
						fg := P(HSun, 255)
						if !sunlit {
							fg = P(HAmber, 215)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						creamMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaV := ((lambda0+f.anim*3)%4000 + 4000) % 4000
							shear := (phi * phi) / 2800
							cloudWarp := (fbmPeriodic(f.m.Seed+601, lambdaV+shear, phi/2, 6, 3, 3) - 512) * 140 / 512
							swirl := fbmPeriodic(f.m.Seed+602, (lambdaV*2+cloudWarp)%4000, phi*2, 8, 4, 2)
							if swirl > 530 || abs(phi) > 750 {
								creamMask |= (1 << q)
							}
						}
						if sunlit {
							fg := P(HWhite, 255)
							bg := P(HSun, 255)
							if creamMask == 15 {
								cell = Cell{R: '█', Fg: fg, Bg: bg}
							} else if creamMask > 0 {
								cell = Cell{R: quadrantRunes[creamMask], Fg: fg, Bg: bg}
							} else {
								cell = Cell{R: '█', Fg: bg, Bg: P(HAmber, 240)}
							}
						} else {
							fg := P(HAmber, 220)
							bg := P(HGold, 195)
							if creamMask > 0 {
								cell = Cell{R: quadrantRunes[creamMask], Fg: fg, Bg: bg}
							} else {
								cell = Cell{R: '█', Fg: bg, Bg: P(HAmber, 185)}
							}
						}
					}

				case 2: // Earth (rx=9, ry=4)
					if mask < 15 {
						fg := P(HWater, 255)
						if !sunlit {
							fg = P(HWater, 190)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						polarMask := 0
						cloudMask := 0
						landMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaE := ((lambda0-f.anim)%4000 + 4000) % 4000
							cloudRot := ((lambda0-f.anim*2+(phi*phi)/3000)%4000 + 4000) % 4000

							bit := 1 << q
							if abs(phi) > 780 {
								polarMask |= bit
							}
							cNoise := fbmPeriodic(f.m.Seed+103, cloudRot, phi, 8, 4, 3)
							if cNoise > 620 {
								cloudMask |= bit
							}
							elev := fbmPeriodic(f.m.Seed+101, lambdaE, phi, 6, 3, 4)
							if elev > 500 {
								landMask |= bit
							}
						}

						switch {
						case polarMask > 0:
							if polarMask == 15 {
								cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: P(HWhite, 255)}
							} else {
								bg := P(HWater, 255)
								if !sunlit {
									bg = P(HWater, 180)
								}
								cell = Cell{R: quadrantRunes[polarMask], Fg: P(HWhite, 255), Bg: bg}
							}
						case cloudMask > 0:
							baseBg := P(HWater, 255)
							if landMask > 0 {
								baseBg = P(HLife, 255)
							}
							if !sunlit {
								baseBg = P(HVoid, 255)
							}
							if cloudMask == 15 {
								cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: baseBg}
							} else {
								cell = Cell{R: quadrantRunes[cloudMask], Fg: P(HWhite, 255), Bg: baseBg}
							}
						case sunlit:
							if landMask == 15 {
								cell = Cell{R: '█', Fg: P(HLife, 255), Bg: P(HSand, 245)}
							} else if landMask > 0 {
								cell = Cell{R: quadrantRunes[landMask], Fg: P(HLife, 255), Bg: P(HWater, 255)}
							} else {
								cell = Cell{R: '█', Fg: P(HWater, 255), Bg: P(HSea, 255)}
							}
						default: // night side
							if landMask > 0 {
								cell = Cell{R: quadrantRunes[landMask], Fg: P(HGold, 255), Bg: P(HVoid, 255)}
							} else {
								cell = Cell{R: '█', Fg: P(HWater, 180), Bg: P(HVoid, 255)}
							}
						}
					}

				case 3: // Mars (rx=7, ry=3)
					if mask < 15 {
						fg := P(HRed, 255)
						if !sunlit {
							fg = P(HRed, 205)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						iceMask := 0
						basaltMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaM := ((lambda0-f.anim)%4000 + 4000) % 4000

							bit := 1 << q
							if phi < -770 || phi > 810 {
								iceMask |= bit
							}
							volc := fbmPeriodic(f.m.Seed+201, lambdaM, phi, 6, 3, 3)
							if volc < 420 {
								basaltMask |= bit
							}
						}

						switch {
						case iceMask > 0:
							if iceMask == 15 {
								cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: P(HWhite, 255)}
							} else {
								cell = Cell{R: quadrantRunes[iceMask], Fg: P(HWhite, 255), Bg: P(HRed, 255)}
							}
						case basaltMask > 0:
							fg := P(HRock, 220)
							bg := P(HRed, 255)
							if !sunlit {
								fg = P(HRock, 180)
								bg = P(HRed, 195)
							}
							if basaltMask == 15 {
								cell = Cell{R: '█', Fg: fg, Bg: P(HMetal, 190)}
							} else {
								cell = Cell{R: quadrantRunes[basaltMask], Fg: fg, Bg: bg}
							}
						case sunlit:
							cell = Cell{R: '█', Fg: P(HRed, 255), Bg: P(HFire, 240)}
						default:
							cell = Cell{R: '█', Fg: P(HRed, 195), Bg: P(HRock, 185)}
						}
					}

				case 4: // Jupiter (rx=14, ry=6)
					if mask < 15 {
						fg := P(HAmber, 255)
						if !sunlit {
							fg = P(HAmber, 205)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						grsCoreMask := 0
						grsBorderMask := 0
						zoneMask := 0

						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaJ := ((lambda0-f.anim*4)%4000 + 4000) % 4000

							wave := (fbmPeriodic(f.m.Seed+301, lambdaJ, phi/2, 4, 2, 2) - 512) * 60 / 512
							phiB := clamp(phi+wave, -1000, 1000)

							// Great Red Spot elliptical vortex in South Tropical Belt
							dLam := ((lambdaJ-2200)+6000)%4000 - 2000
							dPhi := phiB - 310
							grsMetric := (dLam*dLam*100)/(380*380) + (dPhi*dPhi*100)/(120*120)

							bit := 1 << q
							if grsMetric < 35 {
								grsCoreMask |= bit
							} else if grsMetric < 100 {
								grsBorderMask |= bit
							}

							// Ammonia White Zones vs Amber Belts
							isZone := (phiB >= -380 && phiB <= -180) || (phiB >= -30 && phiB <= 80) || (phiB >= 260 && phiB <= 460) || grsMetric < 150
							if isZone {
								zoneMask |= bit
							}
						}

						switch {
						case grsCoreMask > 0:
							fg := P(HRed, 255)
							bg := P(HAmber, 240)
							if !sunlit {
								fg = P(HRed, 210)
								bg = P(HAmber, 195)
							}
							if grsCoreMask == 15 {
								cell = Cell{R: '█', Fg: fg, Bg: fg}
							} else {
								cell = Cell{R: quadrantRunes[grsCoreMask], Fg: fg, Bg: bg}
							}
						case grsBorderMask > 0:
							fg := P(HAmber, 240)
							bg := P(HFire, 245)
							if !sunlit {
								fg = P(HAmber, 195)
								bg = P(HFire, 190)
							}
							if grsBorderMask == 15 {
								cell = Cell{R: '█', Fg: fg, Bg: bg}
							} else {
								cell = Cell{R: quadrantRunes[grsBorderMask], Fg: fg, Bg: bg}
							}
						case sunlit:
							if zoneMask == 15 {
								cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: P(HGold, 255)}
							} else if zoneMask > 0 {
								cell = Cell{R: quadrantRunes[zoneMask], Fg: P(HWhite, 255), Bg: P(HAmber, 255)}
							} else {
								cell = Cell{R: '█', Fg: P(HAmber, 255), Bg: P(HFire, 240)}
							}
						default: // night side
							if zoneMask > 0 {
								cell = Cell{R: quadrantRunes[zoneMask], Fg: P(HAmber, 210), Bg: P(HFire, 190)}
							} else {
								cell = Cell{R: '█', Fg: P(HFire, 190), Bg: P(HAmber, 180)}
							}
						}
					}

				case 5: // Saturn (rx=11, ry=5)
					if mask < 15 {
						fg := P(HGold, 255)
						if !sunlit {
							fg = P(HGold, 210)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						brightMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaS := ((lambda0-f.anim*3)%4000 + 4000) % 4000
							haze := fbmPeriodic(f.m.Seed+551, lambdaS, phi, 6, 3, 2)
							if (phi >= -140 && phi <= 140) || (abs(phi) > 340 && abs(phi) < 560 && haze > 460) {
								brightMask |= (1 << q)
							}
						}
						if sunlit {
							if brightMask == 15 {
								cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: P(HGold, 255)}
							} else if brightMask > 0 {
								cell = Cell{R: quadrantRunes[brightMask], Fg: P(HWhite, 255), Bg: P(HGold, 255)}
							} else {
								cell = Cell{R: '█', Fg: P(HGold, 255), Bg: P(HAmber, 240)}
							}
						} else {
							cell = Cell{R: '█', Fg: P(HGold, 205), Bg: P(HAmber, 190)}
						}
					}

				case 6: // Uranus (rx=9, ry=4)
					if mask < 15 {
						fg := P(HCyan, 255)
						if !sunlit {
							fg = P(HCyan, 205)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						paleMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaU := ((lambda0-f.anim)%4000 + 4000) % 4000
							haze := fbmPeriodic(f.m.Seed+651, lambdaU, phi, 6, 3, 2)
							if abs(phi) < 220 || (abs(phi) > 580 && haze > 480) {
								paleMask |= (1 << q)
							}
						}
						if sunlit {
							if paleMask == 15 {
								cell = Cell{R: '█', Fg: P(HCyan, 255), Bg: P(HAir, 245)}
							} else if paleMask > 0 {
								cell = Cell{R: quadrantRunes[paleMask], Fg: P(HCyan, 255), Bg: P(HCool, 240)}
							} else {
								cell = Cell{R: '█', Fg: P(HCool, 240), Bg: P(HWater, 230)}
							}
						} else {
							cell = Cell{R: '█', Fg: P(HCool, 195), Bg: P(HWater, 185)}
						}
					}

				case 7: // Neptune (rx=9, ry=4)
					if mask < 15 {
						fg := P(HWater, 255)
						if !sunlit {
							fg = P(HWater, 200)
						}
						cell = Cell{R: quadrantRunes[mask], Fg: fg, Bg: spaceBg}
					} else {
						cirrusMask := 0
						darkSpotMask := 0
						bandMask := 0
						for q := 0; q < 4; q++ {
							subX := subOffsets[q][0]
							subY := subOffsets[q][1]
							dx := 4*dxCell + subX - 2
							dy := 4*dyCell + subY - 2
							phi := asinFrac(dy, subRY)
							rxLatSub := isqrt(max(1, rX2-(dy*dy*rX2)/rY2))
							lambda0 := asinFrac(dx, rxLatSub)
							lambdaN := ((lambda0-f.anim*2)%4000 + 4000) % 4000
							cirrusRot := ((lambda0-f.anim*5)%4000 + 4000) % 4000

							bit := 1 << q
							cNoise := fbmPeriodic(f.m.Seed+701, cirrusRot, phi, 8, 4, 3)
							if cNoise > 630 {
								cirrusMask |= bit
							}
							if fbmPeriodic(f.m.Seed+702, lambdaN, phi, 6, 3, 3) < 360 {
								darkSpotMask |= bit
							}
							if abs(phi) < 260 {
								bandMask |= bit
							}
						}

						switch {
						case cirrusMask > 0:
							if cirrusMask == 15 {
								cell = Cell{R: '█', Fg: P(HWhite, 255), Bg: P(HWater, 255)}
							} else {
								cell = Cell{R: quadrantRunes[cirrusMask], Fg: P(HWhite, 255), Bg: P(HWater, 255)}
							}
						case darkSpotMask > 0:
							if darkSpotMask == 15 {
								cell = Cell{R: '█', Fg: P(HDeep, 255), Bg: P(HVoid, 255)}
							} else {
								cell = Cell{R: quadrantRunes[darkSpotMask], Fg: P(HDeep, 255), Bg: P(HWater, 255)}
							}
						case sunlit:
							if bandMask == 15 {
								cell = Cell{R: '█', Fg: P(HCyan, 245), Bg: P(HWater, 255)}
							} else if bandMask > 0 {
								cell = Cell{R: quadrantRunes[bandMask], Fg: P(HCyan, 245), Bg: P(HWater, 255)}
							} else {
								cell = Cell{R: '█', Fg: P(HWater, 255), Bg: P(HCool, 230)}
							}
						default:
							cell = Cell{R: '█', Fg: P(HWater, 200), Bg: P(HVoid, 255)}
						}
					}
				}

				f.g.paint(x, y, cell)
				f.g.mark(x, y, Thing{Kind: KindTerrain, Key: key})
			}
		}
	}

	// 5. Draw All 8 Worlds with Planetary Rings & Moons
	for i := 0; i < 8; i++ {
		pos := planetPos[i]
		px, py := pos.X, pos.Y
		key := "planet"
		if i == 2 {
			key = "home"
		}

		switch i {
		case 0: // Mercury (rx=5, ry=2)
			drawPlanetSphere(px, py, 5, 2, 0, key)
			a := (f.anim * 12) % 4096
			paintSatellite(px+(7*cos4096(a))/1024, py+(3*sin4096(a))/1024, '▫', P(HCyan, 255), key)

		case 1: // Venus (rx=6, ry=2)
			drawPlanetSphere(px, py, 6, 2, 1, key)
			a := (f.anim * 9) % 4096
			paintSatellite(px+(8*cos4096(a))/1024, py+(3*sin4096(a))/1024, '◈', P(HCyan, 255), key)

		case 2: // Earth (rx=6, ry=2)
			drawPlanetSphere(px, py, 6, 2, 2, key)
			la := (f.anim * 4) % 4096
			paintSatellite(px+(10*cos4096(la))/1024, py+(4*sin4096(la))/1024, '●', P(HWhite, 255), key)

		case 3: // Mars (rx=5, ry=2)
			drawPlanetSphere(px, py, 5, 2, 3, key)
			pa := (f.anim * 12) % 4096
			paintSatellite(px+(8*cos4096(pa))/1024, py+(3*sin4096(pa))/1024, '∙', P(HRock, 230), key)

		case 4: // Jupiter (rx=10, ry=4)
			drawPlanetSphere(px, py, 10, 4, 4, key)
			ioA := (f.anim * 10) % 4096
			paintSatellite(px+(14*cos4096(ioA))/1024, py+(6*sin4096(ioA))/1024, '●', P(HSun, 255), key)
			euA := (f.anim * 8) % 4096
			paintSatellite(px+(16*cos4096(euA))/1024, py+(7*sin4096(euA))/1024, '●', P(HCyan, 255), key)
			gaA := (f.anim * 5) % 4096
			paintSatellite(px+(18*cos4096(gaA))/1024, py+(8*sin4096(gaA))/1024, '●', P(HWhite, 255), key)
			caA := (f.anim * 3) % 4096
			paintSatellite(px+(20*cos4096(caA))/1024, py+(9*sin4096(caA))/1024, '∙', P(HRock, 230), key)

		case 5: // Saturn (rx=8, ry=3, rings rx=18, ry=5)
			paintSaturnRings(px, py, true)
			drawPlanetSphere(px, py, 8, 3, 5, key)
			paintSaturnRings(px, py, false)
			ta := (f.anim * 4) % 4096
			paintSatellite(px+(23*cos4096(ta))/1024, py+(7*sin4096(ta))/1024, '●', P(HAmber, 255), key)

		case 6: // Uranus (rx=6, ry=2)
			drawPlanetSphere(px, py, 6, 2, 6, key)

		case 7: // Neptune (rx=6, ry=2)
			drawPlanetSphere(px, py, 6, 2, 7, key)
			tra := (4096 - (f.anim*4)%4096) % 4096
			paintSatellite(px+(10*cos4096(tra))/1024, py+(4*sin4096(tra))/1024, '●', P(HAir, 245), key)
		}
	}
}
