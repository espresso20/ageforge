package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// Floors per theming.md §8.
const (
	floorBodyText  = 4.5 // Text/Label/Positive/Negative/Highlight vs Background (AA)
	floorLargeText = 3.0 // Dim and Accent vs Background (AA large-text / glyphs)
	floorSelection = 4.5 // Text drawn on the Selection background
)

// TestContrast_AllThemes enforces the WCAG AA luminance floors for every shipped
// theme (§8). Themes are tuned against this test, not by eye — a theme that fails
// here cannot ship.
func TestContrast_AllThemes(t *testing.T) {
	for _, th := range All() {
		th := th
		t.Run(th.Key, func(t *testing.T) {
			bg := th.Color(RoleBackground)

			body := []Role{RoleText, RoleLabel, RolePositive, RoleNegative, RoleHighlight}
			for _, role := range body {
				if r := ContrastRatio(th.Color(role), bg); r < floorBodyText {
					t.Errorf("%s: %s vs Background = %.2f, want >= %.1f (AA body text)",
						th.Key, role, r, floorBodyText)
				}
			}

			large := []Role{RoleDim, RoleAccent}
			for _, role := range large {
				if r := ContrastRatio(th.Color(role), bg); r < floorLargeText {
					t.Errorf("%s: %s vs Background = %.2f, want >= %.1f (AA large)",
						th.Key, role, r, floorLargeText)
				}
			}

			// Text must stay readable when drawn on a selected row.
			if r := ContrastRatio(th.Color(RoleText), th.Color(RoleSelection)); r < floorSelection {
				t.Errorf("%s: Text vs Selection = %.2f, want >= %.1f",
					th.Key, r, floorSelection)
			}
		})
	}
}

// TestContrast_AccessibleColorblind asserts that for every accessible theme, the
// ± colors (Positive vs Negative) stay perceptibly distinct under BOTH
// deuteranopia and protanopia simulation (§8). This is the check luminance contrast
// is blind to — it catches a blue/orange pair collapsing under CVD (or a stray
// red/green sneaking into an "accessible" theme).
func TestContrast_AccessibleColorblind(t *testing.T) {
	for _, th := range All() {
		if !th.Accessible {
			continue
		}
		th := th
		t.Run(th.Key, func(t *testing.T) {
			pos := th.Color(RolePositive)
			neg := th.Color(RoleNegative)

			for _, sim := range []string{"deuteranopia", "protanopia"} {
				if !Distinguishable(pos, neg, sim) {
					t.Errorf("%s: Positive vs Negative NOT distinguishable under %s "+
						"(ΔE < %.0f) — ± encoding collapses for colorblind players",
						th.Key, sim, DistinguishableThreshold)
				}
			}

			// Accent must not collide with Positive under CVD either (avoids
			// "is that a gain or a border?" ambiguity).
			acc := th.Color(RoleAccent)
			for _, sim := range []string{"deuteranopia", "protanopia"} {
				if !Distinguishable(acc, pos, sim) {
					t.Errorf("%s: Accent vs Positive NOT distinguishable under %s",
						th.Key, sim)
				}
			}

			// Belt-and-suspenders: accessible themes must carry signed glyphs so
			// the sign is encoded by shape, not just hue (§4).
			if th.GainGlyph == "" || th.LossGlyph == "" {
				t.Errorf("%s: accessible theme must set GainGlyph and LossGlyph", th.Key)
			}
		})
	}
}

// TestContrast_RedGreenCollapses is a guard on the guard: a naive red/green pair
// must FAIL the colorblind check, proving the threshold actually discriminates and
// isn't trivially satisfied by any two colors.
func TestContrast_RedGreenCollapses(t *testing.T) {
	red := tcell.NewRGBColor(0xf8, 0x51, 0x49)
	green := tcell.NewRGBColor(0x3f, 0xb9, 0x50)
	for _, sim := range []string{"deuteranopia", "protanopia"} {
		if Distinguishable(red, green, sim) {
			t.Errorf("red/green reported distinguishable under %s — threshold %.0f is too low",
				sim, DistinguishableThreshold)
		}
	}
	// Sanity: with no simulation, red and green ARE different.
	if !Distinguishable(red, green, "none") {
		t.Errorf("red/green should be distinguishable with no CVD simulation")
	}
}

// TestRelativeLuminance_Endpoints sanity-checks the WCAG math at the extremes:
// black -> 0, white -> 1, and white/black contrast == 21.
func TestRelativeLuminance_Endpoints(t *testing.T) {
	black := tcell.NewRGBColor(0, 0, 0)
	white := tcell.NewRGBColor(255, 255, 255)
	if l := RelativeLuminance(black); l > 1e-9 {
		t.Errorf("luminance(black) = %v, want ~0", l)
	}
	if l := RelativeLuminance(white); l < 1-1e-9 {
		t.Errorf("luminance(white) = %v, want ~1", l)
	}
	if r := ContrastRatio(white, black); r < 20.9 || r > 21.1 {
		t.Errorf("contrast(white,black) = %.3f, want ~21", r)
	}
}
