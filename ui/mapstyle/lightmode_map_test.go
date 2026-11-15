package mapstyle

import (
	"testing"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
)

// TestSpaceMapIsDark verifies that outer space scales (Scale >= Orbit) render
// on a pitch-black void, that white celestial bodies (Sol core, polar ice caps,
// cloud belts, spaceship hulls) keep their pure white rather than inverting to
// black, and that block runes are never inverted by text legibility rules. The
// palette is the same in every theme (see TestPictureIgnoresTheTheme).
func TestSpaceMapIsDark(t *testing.T) {
	spaceGrid := &mapgen.Grid{
		W:     80,
		H:     24,
		Wall:  mapgen.HMetal,
		Tint:  mapgen.HNone,
		Space: true,
	}
	p := cellPaletteFor(nil, spaceGrid)
	white := theme.SpaceWhiteColor()

	if wantVoid := theme.SpaceColor(theme.SpaceVoid); p.bg != wantVoid {
		t.Errorf("space background = %06x, want SpaceVoid %06x", p.bg.Hex(), wantVoid.Hex())
	}
	if voidCol := p.over(mapgen.P(mapgen.HVoid, 255)); voidCol != theme.SpaceColor(theme.SpaceVoid) {
		t.Errorf("over(HVoid) = %06x, want %06x", voidCol.Hex(), theme.SpaceColor(theme.SpaceVoid).Hex())
	}
	if whiteHue := p.hue[mapgen.HWhite]; whiteHue != white {
		t.Errorf("hue[HWhite] = %06x, want %06x (SpaceWhiteColor)", whiteHue.Hex(), white.Hex())
	}

	// Sol core (HWhite foreground on HSun yellow background with block runes) must stay white.
	for _, r := range []rune{'█', '▘', '▀', '▄'} {
		cell := mapgen.Cell{R: r, Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSun, 255)}
		if fg, _, _ := p.style(cell).Decompose(); fg != white {
			t.Errorf("Sol core rune %q fg = %06x, want white %06x (inverted to black!)", r, fg.Hex(), white.Hex())
		}
	}

	// Mars polar ice cap (HWhite on HRock with quadrant runes) must stay bright white.
	marsCap := mapgen.Cell{R: '▙', Fg: mapgen.P(mapgen.HWhite, 235), Bg: mapgen.P(mapgen.HRock, 245)}
	if fg, _, _ := p.style(marsCap).Decompose(); theme.RelativeLuminance(fg) < 0.7 {
		t.Errorf("Mars polar cap fg = %06x, luminance = %.2f, want bright white (>= 0.7)", fg.Hex(), theme.RelativeLuminance(fg))
	}

	// Jupiter cloud belt (HWhite on HCloud with '█') and a spaceship hull (HWhite on HSteel) stay white.
	for name, cell := range map[string]mapgen.Cell{
		"Jupiter cloud belt": {R: '█', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HCloud, 220)},
		"Spaceship hull":     {R: '█', Fg: mapgen.P(mapgen.HWhite, 255), Bg: mapgen.P(mapgen.HSteel, 180)},
	} {
		if fg, _, _ := p.style(cell).Decompose(); fg != white {
			t.Errorf("%s fg = %06x, want white %06x", name, fg.Hex(), white.Hex())
		}
	}
}

// TestGroundMapKeepsItsDarkGround verifies that the ground scale renders on
// Forge's dark background, that material colors of buildings and wonders
// using block characters are preserved, and that night dimming keeps contrast
// >= 1.5.
func TestGroundMapKeepsItsDarkGround(t *testing.T) {
	groundGrid := &mapgen.Grid{
		W:     80,
		H:     24,
		Wall:  mapgen.HStone,
		Tint:  mapgen.HNone,
		Space: false,
	}
	p := cellPaletteFor(nil, groundGrid)

	if want := pictureRole(theme.RoleBackground); p.bg != want {
		t.Errorf("ground background = %06x, want Forge's %06x", p.bg.Hex(), want.Hex())
	}
	if lum := theme.RelativeLuminance(p.bg); lum > 0.1 {
		t.Errorf("ground background luminance = %.2f, want dark (<= 0.1)", lum)
	}

	// Building block runes keep their material colors and are not inverted to RoleText ink.
	textInk := pictureRole(theme.RoleText)
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
		cell := mapgen.Cell{R: '█', Fg: mapgen.P(mat.hue, 255), Bg: mapgen.P(mat.bg, 36)}
		fg, bg, _ := p.style(cell).Decompose()
		if fg == textInk {
			t.Errorf("%s building block was inverted to RoleText ink (%06x)", mat.name, textInk.Hex())
		}
		if cr := theme.ContrastRatio(fg, bg); cr < 1.5 {
			t.Errorf("%s building has contrast %.2f, want >= 1.5", mat.name, cr)
		}
	}

	// The glyph stays clear of its ground across the Dim levels of the night.
	for _, dim := range []uint8{50, 120, 165, 200, 255} {
		cell := mapgen.Cell{R: '█', Fg: mapgen.P(mapgen.HStone, 235), Bg: mapgen.P(mapgen.HGrass, 36), Dim: dim}
		fg, bg, _ := p.style(cell).Decompose()
		if cr := theme.ContrastRatio(fg, bg); cr < 1.5 {
			t.Errorf("dim %d: contrast = %.2f, want >= 1.5", dim, cr)
		}
	}
}

// TestRenderedAgesLightAndDarkThemes tests end-to-end rendering of ground and space
// maps across standard, accessibility, and flavor themes to ensure space maps are
// dark with glowing stars/sun, and ground maps keep legible glyphs, in every theme.
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

							// Void cells in space must be dark in every theme (not washed out off-white in light themes).
							if r == ' ' || r == '·' || r == '∙' || r == '*' || r == '•' || r == '✦' {
								if lum := theme.RelativeLuminance(bg); lum > 0.15 {
									t.Errorf("%s %s tick %d at (%d,%d): space cell bg luminance = %.2f (%06x), want dark (<= 0.15)",
										thKey, ta.age, tick, x, y, lum, bg.Hex())
									break
								}
							}

							// Verify Sol core cells with HSun background stay white (never inverted to black).
							if bg == theme.SpaceColor(theme.SpaceSun) && isBlockRune(r) {
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

func isYellowGold(c theme.Color) bool {
	r, g, b := c.RGB()
	// Yellow/Gold has high red and green, lower blue
	return r > 180 && g > 140 && b < 120
}
