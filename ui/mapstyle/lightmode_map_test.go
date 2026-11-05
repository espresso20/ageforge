package mapstyle

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
)

// TestSpaceMapsAreDarkAcrossAllThemes verifies that outer space scales (Scale >= Orbit)
// render on a pitch-black / dark void across all themes (both dark and light),
// that white celestial bodies (Sol core, polar ice caps, cloud belts, spaceship hulls)
// keep their pure white color rather than inverting to black, and that block runes
// are never inverted by text legibility rules.
func TestSpaceMapsAreDarkAcrossAllThemes(t *testing.T) {
	spaceGrid := &mapgen.Grid{
		W:     80,
		H:     24,
		Wall:  mapgen.HMetal,
		Tint:  mapgen.HNone,
		Space: true,
	}

	for th := range fixture.Themes(t) {
		p := cellPaletteFor(nil, spaceGrid)

		// 1. Space background must always be dark space void (not off-white/grey in light themes).
		if th.Duotone {
			if lum := theme.RelativeLuminance(p.bg); lum > 0.3 {
				t.Errorf("%s (duotone): space background luminance = %.3f, want <= 0.3", th.Key, lum)
			}
		} else {
			wantVoid := theme.SpaceColor(theme.SpaceVoid)
			if p.bg != wantVoid {
				t.Errorf("%s: space background = %06x, want SpaceVoid %06x", th.Key, p.bg.Hex(), wantVoid.Hex())
			}
		}

		// 2. HVoid at weight 255 must be dark space void.
		voidCol := p.over(mapgen.P(mapgen.HVoid, 255))
		if !th.Duotone && voidCol != theme.SpaceColor(theme.SpaceVoid) {
			t.Errorf("%s: over(HVoid) = %06x, want %06x", th.Key, voidCol.Hex(), theme.SpaceColor(theme.SpaceVoid).Hex())
		}

		// 3. HWhite hue must be white across all non-duotone themes (not near-black RoleBright).
		if !th.Duotone {
			whiteHue := p.hue[mapgen.HWhite]
			if whiteHue != theme.SpaceWhiteColor() {
				t.Errorf("%s: hue[HWhite] = %06x, want %06x (SpaceWhiteColor)", th.Key, whiteHue.Hex(), theme.SpaceWhiteColor().Hex())
			}
		}

		// 4. Sol core (HWhite foreground on HSun yellow background with block runes) must stay white.
		solCoreCells := []mapgen.Cell{
			{R: '█', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSun, 255)},
			{R: '▘', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSun, 255)},
			{R: '▀', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSun, 255)},
			{R: '▄', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSun, 255)},
		}
		for _, cell := range solCoreCells {
			st := p.style(cell)
			fg, _, _ := st.Decompose()
			if !th.Duotone && fg != theme.SpaceWhiteColor() {
				t.Errorf("%s: Sol core rune %q fg = %06x, want white %06x (inverted to black!)",
					th.Key, cell.R, fg.Hex(), theme.SpaceWhiteColor().Hex())
			}
		}

		// 5. Mars polar ice cap (HWhite on HRock with quadrant runes) must stay bright white.
		marsCap := mapgen.Cell{R: '▙', Fg: mapgen.P(mapgen.HWhite, 235), Bg: mapgen.P(mapgen.HRock, 245)}
		stCap := p.style(marsCap)
		fgCap, _, _ := stCap.Decompose()
		if !th.Duotone && theme.RelativeLuminance(fgCap) < 0.7 {
			t.Errorf("%s: Mars polar cap fg = %06x, luminance = %.2f, want bright white (>= 0.7)",
				th.Key, fgCap.Hex(), theme.RelativeLuminance(fgCap))
		}

		// 6. Jupiter cloud belt (HWhite on HCloud with '█') must stay bright white.
		jupCloud := mapgen.Cell{R: '█', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HCloud, 220)}
		stJup := p.style(jupCloud)
		fgJup, _, _ := stJup.Decompose()
		if !th.Duotone && fgJup != theme.SpaceWhiteColor() {
			t.Errorf("%s: Jupiter cloud belt fg = %06x, want white %06x", th.Key, fgJup.Hex(), theme.SpaceWhiteColor().Hex())
		}

		// 7. Spaceship hull (HWhite on HSteel with '█') must stay bright white.
		shipHull := mapgen.Cell{R: '█', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSteel, 180)}
		stShip := p.style(shipHull)
		fgShip, _, _ := stShip.Decompose()
		if !th.Duotone && fgShip != theme.SpaceWhiteColor() {
			t.Errorf("%s: Spaceship hull fg = %06x, want white %06x", th.Key, fgShip.Hex(), theme.SpaceWhiteColor().Hex())
		}
	}
}

// TestGroundMapsInLightThemes verifies that ground scales in light themes render
// on paper-like backgrounds, that material colors of buildings and wonders
// using block characters are preserved, and that night dimming darkens naturally
// with contrast remaining >= 1.5.
func TestGroundMapsInLightThemes(t *testing.T) {
	groundGrid := &mapgen.Grid{
		W:     80,
		H:     24,
		Wall:  mapgen.HStone,
		Tint:  mapgen.HNone,
		Space: false,
	}

	lightThemes := []string{"daylight", "high_contrast_light"}
	for _, thKey := range lightThemes {
		fixture.WithTheme(t, thKey)
		p := cellPaletteFor(nil, groundGrid)

		// 1. Background must be light paper.
		if lum := theme.RelativeLuminance(p.bg); lum < 0.8 {
			t.Errorf("%s: ground background luminance = %.2f, want >= 0.8", thKey, lum)
		}

		// 2. Building block runes should keep material colors and NOT invert to black ink (RoleText).
		textInk := theme.Color(theme.RoleText)
		materials := []struct {
			name string
			hue  mapgen.Hue
			bg   mapgen.Hue
		}{
			{"thatch", mapgen.HThatch, mapgen.HEarth},
			{"stone", mapgen.HStone, mapgen.HGrass},
			{"gold", mapgen.HGold, mapgen.HLawn},
			{"brick", mapgen.HBrick, mapgen.HPath},
			{"timber", mapgen.HTimber, mapgen.HEarth},
		}

		for _, mat := range materials {
			cell := mapgen.Cell{
				R:  '█',
				Fg: mapgen.P(mat.hue, 255),
				Bg: mapgen.P(mat.bg, 36),
			}
			st := p.style(cell)
			fg, bg, _ := st.Decompose()

			// The material must not be pulled to RoleText (near-black ink).
			if fg == textInk {
				t.Errorf("%s: %s building block was inverted to RoleText ink (%06x)", thKey, mat.name, textInk.Hex())
			}

			// Contrast must be at least 1.5.
			if cr := theme.ContrastRatio(fg, bg); cr < 1.5 {
				t.Errorf("%s: %s building has contrast %.2f, want >= 1.5", thKey, mat.name, cr)
			}
		}

		// 3. Test night dimming across various Dim levels.
		for _, dim := range []uint8{50, 120, 165, 200, 255} {
			cell := mapgen.Cell{
				R:   '█',
				Fg:  mapgen.P(mapgen.HStone, 235),
				Bg:  mapgen.P(mapgen.HGrass, 36),
				Dim: dim,
			}
			st := p.style(cell)
			fg, bg, _ := st.Decompose()

			// Night dimming must darken both fg and bg.
			if theme.RelativeLuminance(bg) >= theme.RelativeLuminance(p.bg) {
				t.Errorf("%s dim %d: bg luminance did not darken", thKey, dim)
			}

			// Contrast must clear the 1.5 floor at night so glyphs do not sink into the ground.
			if cr := theme.ContrastRatio(fg, bg); cr < 1.5 {
				t.Errorf("%s dim %d: contrast = %.2f, want >= 1.5", thKey, dim, cr)
			}
		}
	}
}

// TestRenderedAgesLightAndDarkThemes tests end-to-end rendering of ground and space
// maps across standard, accessibility, and flavor themes to ensure space maps are
// dark with glowing stars/sun, and ground maps are clean ink-on-paper with legible glyphs.
func TestRenderedAgesLightAndDarkThemes(t *testing.T) {
	testThemes := []string{
		"forge",               // Dark standard
		"daylight",            // Light standard
		"high_contrast_light", // Light accessibility
		"parchment",           // Light duotone
		"monochrome",          // Dark duotone
		"cosmic",              // Dark flavor
	}

	testAges := []struct {
		age     string
		isSpace bool
	}{
		{"stone_age", false},
		{"modern_age", false},
		{"space_age", true},
		{"interstellar_age", true},
	}

	for _, thKey := range testThemes {
		fixture.WithTheme(t, thKey)
		th := theme.Active()

		for _, ta := range testAges {
			for _, tick := range []int{noon, midnight} {
				m := model(ta.age, tick)
				scr := draw(New(), m, 120, 36, false, mapmodel.TierUnicode)
				cells, w, h := scr.GetContents()

				if ta.isSpace {
					side := 0
					if 120 >= 110 {
						side = sideWidth
					}
					mapW, mapH := 120-side, 36-5

					for y := 1; y < 1+mapH; y++ {
						for x := 0; x < mapW; x++ {
							c := cells[y*w+x]
							fg, bg, _ := c.Style.Decompose()
							r := rune(' ')
							if len(c.Runes) > 0 {
								r = c.Runes[0]
							}

							// Void cells in space must be dark (not washed out off-white in light themes).
							if r == ' ' || r == '·' || r == '∙' || r == '*' || r == '•' || r == '✦' {
								if !th.Duotone && th.IsLight() {
									if lum := theme.RelativeLuminance(bg); lum > 0.15 {
										t.Errorf("%s %s tick %d at (%d,%d): space cell bg luminance = %.2f (%06x), want dark (<= 0.15)",
											thKey, ta.age, tick, x, y, lum, bg.Hex())
										break
									}
								}
							}

							// Verify Sol core cells with HSun background stay white (never inverted to black).
							if !th.Duotone && bg == theme.SpaceColor(theme.SpaceSun) && isBlockRune(r) {
								if fg != theme.SpaceWhiteColor() {
									t.Errorf("%s %s tick %d at (%d,%d): Sol core rune %q on SpaceSun (%06x) has fg=%06x, want SpaceWhiteColor (inverted to black!)",
										thKey, ta.age, tick, x, y, r, bg.Hex(), fg.Hex())
								}
							}
						}
					}
				} else {
					// Ground ages: verify all glyphs have contrast >= 1.5.
					for i, c := range cells[:w*h] {
						if len(c.Runes) == 0 || c.Runes[0] == ' ' {
							continue
						}
						fg, bg, _ := c.Style.Decompose()
						if cr := theme.ContrastRatio(fg, bg); cr < 1.5 {
							t.Fatalf("theme %s, %s tick %d: %q at (%d,%d) has contrast %.2f (fg=%06x bg=%06x)",
								thKey, ta.age, tick, c.Runes[0], i%w, i/w, cr, fg.Hex(), bg.Hex())
						}
					}
				}
				scr.Fini()
			}
		}
	}
}

func isYellowGold(c tcell.Color) bool {
	r, g, b := c.RGB()
	// Yellow/Gold has high red and green, lower blue
	return r > 180 && g > 140 && b < 120
}
