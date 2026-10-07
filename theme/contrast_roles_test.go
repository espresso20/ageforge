package theme

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// contrastPair is one "text role drawn on surface role" combination the UI
// actually produces, with the WCAG floor it must clear. Standard floors are AA
// (4.5 body text, 3.0 secondary / large / non-text UI); accessibility themes use
// the stricter column (AAA 7.0 for primary text, 4.5 for everything else).
type contrastPair struct {
	fg, bg     Role
	floor      float64
	accessible float64
	why        string
}

// surfacePairs is the full matrix. Every role that carries text is checked on
// both the canvas (Background) and the panel color (Surface), because overlays
// and modals draw the same tags on Surface.
var surfacePairs = func() []contrastPair {
	var ps []contrastPair
	for _, bg := range []Role{RoleBackground, RoleSurface} {
		ps = append(ps,
			contrastPair{RoleText, bg, 4.5, 7.0, "primary text"},
			contrastPair{RoleBright, bg, 4.5, 7.0, "emphasis text"},
			contrastPair{RoleLabel, bg, 4.5, 4.5, "labels / values"},
			contrastPair{RoleHighlight, bg, 4.5, 4.5, "numbers"},
			contrastPair{RolePositive, bg, 4.5, 4.5, "gains"},
			contrastPair{RoleNegative, bg, 4.5, 4.5, "losses"},
			contrastPair{RoleWarning, bg, 4.5, 4.5, "warnings"},
			contrastPair{RoleDim, bg, 3.0, 4.5, "secondary text"},
			contrastPair{RoleAccent, bg, 3.0, 4.5, "titles"},
			contrastPair{RoleBorder, bg, 3.0, 3.0, "borders (non-text UI)"},
		)
	}
	ps = append(ps,
		contrastPair{RoleSelectionText, RoleSelection, 4.5, 7.0, "selected row"},
		contrastPair{RoleOnAccent, RoleAccent, 4.5, 4.5, "keycaps / primary buttons"},
		contrastPair{RoleText, RoleChip, 4.5, 7.0, "keycap labels"},
		contrastPair{RoleOnNegative, RoleNegative, 3.0, 4.5, "danger modal text (bold)"},
	)
	return ps
}()

// TestContrast_RoleMatrix asserts, for every shipped theme, the WCAG floor for
// every text-on-surface pair the UI draws. This is the gate that lets a light
// theme ship: its roles are tuned against this matrix, not by eye.
func TestContrast_RoleMatrix(t *testing.T) {
	for _, th := range All() {
		th := th
		t.Run(th.Key, func(t *testing.T) {
			for _, p := range surfacePairs {
				floor := p.floor
				if th.Accessible {
					floor = p.accessible
				}
				r := ContrastRatio(th.Color(p.fg), th.Color(p.bg))
				if r < floor {
					t.Errorf("%s on %s = %.2f:1, want >= %.1f (%s)", p.fg, p.bg, r, floor, p.why)
				}
			}
		})
	}
}

// TestRoles_AllSet guarantees no theme ships a role left at ColorDefault — an
// unset role would resolve to the terminal's own color and break the "every
// theme paints its own surface" rule.
func TestRoles_AllSet(t *testing.T) {
	for _, th := range All() {
		for r := Role(0); r < numRoles; r++ {
			if th.Color(r) == tcell.ColorDefault {
				t.Errorf("%s: role %s is unset", th.Key, r)
			}
			if !th.Color(r).IsRGB() {
				t.Errorf("%s: role %s is not a true-RGB color", th.Key, r)
			}
		}
	}
}

// TestDerivedRoles_PreserveDarkThemes pins the no-regression contract: for the
// themes that predate the extended roles, the derived values reproduce what the
// UI drew before (Surface was the canvas, borders were Accent, selected rows drew
// Text, warnings were Highlight).
func TestDerivedRoles_PreserveDarkThemes(t *testing.T) {
	for _, th := range []Theme{Forge, Deuteranopia, Protanopia, HighContrast, Bronze, Cyberpunk, Monochrome, Cosmic} {
		eq := func(a, b Role) {
			if th.Color(a) != th.Color(b) {
				t.Errorf("%s: %s = %06x, want = %s %06x", th.Key, a, th.Color(a).Hex(), b, th.Color(b).Hex())
			}
		}
		eq(RoleSurface, RoleBackground)
		eq(RoleBorder, RoleAccent)
		eq(RoleSelectionText, RoleText)
		eq(RoleBright, RoleText)
		eq(RoleWarning, RoleHighlight)
	}
	// Forge's keycaps were black on gold with a #30363d label chip, and danger
	// modals drew white text; all three must survive.
	if got := Forge.Color(RoleOnAccent).Hex(); got != 0x000000 {
		t.Errorf("Forge OnAccent = %06x, want 000000", got)
	}
	if got := Forge.Color(RoleChip).Hex(); got != 0x30363d {
		t.Errorf("Forge Chip = %06x, want 30363d", got)
	}
	if got := Forge.Color(RoleOnNegative).Hex(); got != 0xffffff {
		t.Errorf("Forge OnNegative = %06x, want ffffff", got)
	}
}

// TestIsLight classifies the shipped themes.
func TestIsLight(t *testing.T) {
	light := map[string]bool{"daylight": true, "high_contrast_light": true, "parchment": true}
	for _, th := range All() {
		if th.IsLight() != light[th.Key] {
			t.Errorf("%s: IsLight() = %v, want %v", th.Key, th.IsLight(), light[th.Key])
		}
		want := "Dark"
		if light[th.Key] {
			want = "Light"
		}
		if th.Variant() != want {
			t.Errorf("%s: Variant() = %q, want %q", th.Key, th.Variant(), want)
		}
	}
}

// TestGroups puts each theme in the expected picker section.
func TestGroups(t *testing.T) {
	want := map[string]Group{
		"forge": GroupStandard, "daylight": GroupStandard,
		"deuteranopia": GroupAccessibility, "protanopia": GroupAccessibility,
		"high_contrast": GroupAccessibility, "high_contrast_light": GroupAccessibility,
		"parchment": GroupUnlockable, "bronze": GroupUnlockable, "cyberpunk": GroupUnlockable,
		"monochrome": GroupUnlockable, "cosmic": GroupUnlockable,
		"source": GroupUnlockable, "glitch": GroupUnlockable,
	}
	for _, th := range All() {
		g, ok := want[th.Key]
		if !ok {
			t.Errorf("theme %q has no expected group; add it to this table", th.Key)
			continue
		}
		if th.Group() != g {
			t.Errorf("%s: Group() = %s, want %s", th.Key, th.Group(), g)
		}
	}
}

// TestLegible checks the identity-hue correction: an unreadable hue is pushed to
// the floor in the right direction; a readable one is untouched.
func TestLegible(t *testing.T) {
	blue := tcell.NewRGBColor(0, 0, 0xff)
	lightBlue := tcell.NewRGBColor(0xad, 0xd8, 0xe6)
	for _, tc := range []struct {
		name string
		c    tcell.Color
		bg   tcell.Color
	}{
		{"blue on forge", blue, Forge.Color(RoleBackground)},
		{"lightblue on daylight", lightBlue, Daylight.Color(RoleBackground)},
		{"lightblue on hc-light", lightBlue, HighContrastLight.Color(RoleBackground)},
	} {
		got := Legible(tc.c, tc.bg, 4.5)
		if r := ContrastRatio(got, tc.bg); r < 4.5 {
			t.Errorf("%s: Legible → %06x at %.2f:1, want >= 4.5", tc.name, got.Hex(), r)
		}
	}
	gold := Forge.Color(RoleAccent)
	if got := Legible(gold, Forge.Color(RoleBackground), 4.5); got != gold {
		t.Errorf("Legible changed an already-readable color: %06x → %06x", gold.Hex(), got.Hex())
	}
}

// TestRefResolve checks the sentinel round-trip and that no real color reads as
// a sentinel.
func TestRefResolve(t *testing.T) {
	t.Cleanup(func() { _ = SetActive(DefaultKey) })
	for _, key := range []string{"forge", "daylight"} {
		_ = SetActive(key)
		th, _ := ByKey(key)
		for r := Role(0); r < numRoles; r++ {
			ref := Ref(r)
			got, ok := RefRole(ref)
			if !ok || got != r {
				t.Fatalf("RefRole(Ref(%s)) = %v,%v", r, got, ok)
			}
			if Resolve(ref) != th.Color(r) {
				t.Errorf("%s: Resolve(Ref(%s)) = %06x, want %06x", key, r, Resolve(ref).Hex(), th.Color(r).Hex())
			}
			if _, isRef := RefRole(th.Color(r)); isRef {
				t.Errorf("%s: real color for %s parses as a sentinel", key, r)
			}
		}
	}
	for _, c := range []tcell.Color{tcell.ColorDefault, tcell.ColorReset, tcell.ColorWhite, tcell.PaletteColor(255)} {
		if _, isRef := RefRole(c); isRef {
			t.Errorf("%v parses as a sentinel", c)
		}
	}
}

// TestThemedScreen_ResolvesStyles draws sentinels and defaults through the wrapper
// and checks the cells hold the active theme's concrete colors.
func TestThemedScreen_ResolvesStyles(t *testing.T) {
	t.Cleanup(func() { _ = SetActive(DefaultKey) })
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(4, 1)
	ts := WrapScreen(sim)

	for _, key := range []string{"forge", "daylight"} {
		_ = SetActive(key)
		th, _ := ByKey(key)
		ts.Clear()
		ts.SetContent(0, 0, 'a', nil, tcell.StyleDefault)
		ts.SetContent(1, 0, 'b', nil, tcell.StyleDefault.Foreground(Ref(RoleAccent)).Background(Ref(RoleSurface)))
		ts.SetContent(2, 0, 'c', nil, tcell.StyleDefault.Foreground(tcell.NewRGBColor(1, 2, 3)).Bold(true))
		ts.Show()
		check := func(x int, wantFg, wantBg tcell.Color) {
			_, _, st, _ := sim.GetContent(x, 0)
			fg, bg, _ := st.Decompose()
			if fg != wantFg || bg != wantBg {
				t.Errorf("%s cell %d: fg/bg = %06x/%06x, want %06x/%06x", key, x, fg.Hex(), bg.Hex(), wantFg.Hex(), wantBg.Hex())
			}
		}
		check(0, th.Color(RoleText), th.Color(RoleBackground))
		check(1, th.Color(RoleAccent), th.Color(RoleSurface))
		check(2, tcell.NewRGBColor(1, 2, 3), th.Color(RoleBackground))
		// A cell nobody wrote is the theme canvas, not the terminal default.
		check(3, th.Color(RoleText), th.Color(RoleBackground))
	}
}

// ExampleKeycap documents the keycap helper's output.
func ExampleKeycap() {
	fmt.Println(Keycap("Esc"))
	// Output: [onaccent:accent:b] Esc [-:-:-]
}
