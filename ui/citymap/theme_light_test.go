package citymap

import (
	"image/color"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// lum is WCAG relative luminance of an image color (via the theme package math).
func lum(c color.RGBA) float64 {
	return theme.RelativeLuminance(tcellFromRGBA(c))
}

// TestMapPolarity_AllThemes pins the luminance contract the map recipes depend on,
// for every theme, AFTER the light-theme re-key: streets read lighter than the
// ground, drop shadows and shaded roofs darker than what they sit on, lit roofs
// lighter than their shaded side, and markers stand off the ground. On a light
// theme this is exactly the set of relationships that used to invert.
func TestMapPolarity_AllThemes(t *testing.T) {
	t.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	styles := map[string]tdEraStyle{
		"primitive_age": styleForAge("primitive_age"),
		"bronze_age":    styleForAge("bronze_age"),
		"medieval_age":  styleForAge("medieval_age"),
		"modern_age":    styleForAge("modern_age"),
	}
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		page := rgba(theme.Color(theme.RoleBackground))
		final := func(c color.RGBA) color.RGBA {
			if theme.IsLight() {
				return liftColor(c, page)
			}
			return c
		}
		p := newTdPal()
		for age, s := range styles {
			ground := final(s.groundBase(p))
			street := final(s.streetCol(p))
			// drawShadow blends the shadow tone 28% over whatever is beneath.
			shadowed := final(blend(s.groundBase(p), p.shadow, 0.28))
			roof := final(s.roofBase(p))
			roofDark := final(s.roofDark(p))
			if !(lum(street) > lum(ground)) {
				t.Errorf("%s/%s: street (%.3f) not lighter than ground (%.3f)", th.Key, age, lum(street), lum(ground))
			}
			// Light themes: a shadow must darken the ground. (Dark themes are pinned to
			// their pre-overhaul look, where the soft grey shadow is a touch lighter than
			// the darkest grounds — deliberately left alone.)
			if th.IsLight() && !(lum(shadowed) < lum(ground)) {
				t.Errorf("%s/%s: shadowed ground (%.3f) not darker than ground (%.3f)", th.Key, age, lum(shadowed), lum(ground))
			}
			if !(lum(roofDark) < lum(roof)) {
				t.Errorf("%s/%s: shaded roof (%.3f) not darker than lit roof (%.3f)", th.Key, age, lum(roofDark), lum(roof))
			}
		}
		// The palace marker must stand off the ground it sits on.
		pal := buildPalette(0)
		ground := final(pal.lowland)
		palace := final(pal.palace)
		if r := theme.ContrastRatio(tcellFromRGBA(palace), tcellFromRGBA(ground)); r < 1.5 {
			t.Errorf("%s: palace marker vs ground contrast %.2f, want >= 1.5", th.Key, r)
		}
	}
}

// TestMapColor_DarkThemesUnchanged guarantees the proxy is light-only: under every
// dark theme mapColor is exactly the theme role, so dark maps render as before.
func TestMapColor_DarkThemesUnchanged(t *testing.T) {
	t.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	for _, th := range theme.All() {
		if th.IsLight() {
			continue
		}
		_ = theme.SetActive(th.Key)
		for r := theme.Role(0); int(r) < theme.NumRoles; r++ {
			if mapColor(r) != rgba(theme.Color(r)) {
				t.Errorf("%s: mapColor(%s) differs from the theme role on a dark theme", th.Key, r)
			}
		}
	}
}

// TestMapColor_LightProxyIsDarkPolarity checks the light proxy: dark canvas, light
// pole, and marker roles bright enough to pop on that canvas.
func TestMapColor_LightProxyIsDarkPolarity(t *testing.T) {
	t.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	for _, th := range theme.All() {
		if !th.IsLight() {
			continue
		}
		_ = theme.SetActive(th.Key)
		bg := mapColor(theme.RoleBackground)
		if lum(bg) > 0.1 {
			t.Errorf("%s: proxy canvas luminance %.3f, want dark (<= 0.1)", th.Key, lum(bg))
		}
		if lum(mapColor(theme.RoleText)) < 0.5 {
			t.Errorf("%s: proxy light pole too dark", th.Key)
		}
		for _, r := range []theme.Role{theme.RoleAccent, theme.RolePositive, theme.RoleNegative, theme.RoleLabel, theme.RoleHighlight} {
			if cr := theme.ContrastRatio(tcellFromRGBA(mapColor(r)), tcellFromRGBA(bg)); cr < 4.5 {
				t.Errorf("%s: proxy %s vs proxy canvas = %.2f, want >= 4.5", th.Key, r, cr)
			}
		}
	}
}
