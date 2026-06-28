package ui

import (
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
)

// restoreForge resets the process-global active theme back to the default after a
// test mutates it, so the switched theme doesn't leak into other cases (the remap
// is process-wide state). Registered via t.Cleanup by every test that switches.
func restoreForge(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		if err := theme.SetActive(theme.DefaultKey); err != nil {
			t.Fatalf("restoreForge: SetActive(%q) failed: %v", theme.DefaultKey, err)
		}
	})
}

func TestCmdThemeListShowsAllThemes(t *testing.T) {
	// Pin the active theme so the marker assertion is independent of test order
	// (the active theme is process-global; another case may have switched it).
	restoreForge(t)
	if err := theme.SetActive(theme.DefaultKey); err != nil {
		t.Fatalf("setup SetActive(%q) failed: %v", theme.DefaultKey, err)
	}

	res := cmdTheme([]string{"list"}, nil)
	if res.Type == "error" {
		t.Fatalf("theme list returned an error: %q", res.Message)
	}
	// Every registered theme's name and key must appear in the listing.
	for _, th := range theme.All() {
		if !strings.Contains(res.Message, th.Name) {
			t.Errorf("theme list missing name %q\ngot: %s", th.Name, res.Message)
		}
		if !strings.Contains(res.Message, th.Key) {
			t.Errorf("theme list missing key %q\ngot: %s", th.Key, res.Message)
		}
	}
	// The active theme must carry the active marker.
	if !strings.Contains(res.Message, "●") {
		t.Errorf("theme list missing active marker ●\ngot: %s", res.Message)
	}
}

func TestCmdThemeBareIsInfoNotError(t *testing.T) {
	// Bare `theme` (no args) directs the player rather than erroring — the
	// interactive picker is opened by the dashboard intercept, not this path.
	res := cmdTheme(nil, nil)
	if res.Type == "error" {
		t.Errorf("bare theme should not be an error, got %q: %q", res.Type, res.Message)
	}
}

func TestCmdThemeValidKeySwitchesActive(t *testing.T) {
	restoreForge(t)
	// Pick a non-default accessible theme so the switch is observable.
	const target = "high_contrast"
	if _, ok := theme.ByKey(target); !ok {
		t.Fatalf("test precondition: theme %q not registered", target)
	}

	res := cmdTheme([]string{target}, nil)
	if res.Type != "success" {
		t.Errorf("theme %s should succeed, got %q: %q", target, res.Type, res.Message)
	}
	if got := theme.Active().Key; got != target {
		t.Errorf("after `theme %s`, active = %q, want %q", target, got, target)
	}
}

func TestCmdThemeBadKeyIsErrorAndDoesNotSwitch(t *testing.T) {
	restoreForge(t)
	// Start from a known non-default theme to prove a bad key doesn't revert or
	// otherwise mutate the active theme.
	if err := theme.SetActive("deuteranopia"); err != nil {
		t.Fatalf("setup SetActive(deuteranopia) failed: %v", err)
	}
	before := theme.Active().Key

	res := cmdTheme([]string{"no_such_theme_xyz"}, nil)
	if res.Type != "error" {
		t.Errorf("unknown theme should be an error, got %q: %q", res.Type, res.Message)
	}
	// The error should hint the valid keys.
	if !strings.Contains(res.Message, theme.DefaultKey) {
		t.Errorf("unknown-theme error should list valid keys (e.g. %q)\ngot: %s", theme.DefaultKey, res.Message)
	}
	if got := theme.Active().Key; got != before {
		t.Errorf("a bad key changed the active theme: %q → %q", before, got)
	}
}

func TestThemeKeysCoversRegistry(t *testing.T) {
	keys := themeKeys()
	if len(keys) != len(theme.All()) {
		t.Fatalf("themeKeys() returned %d keys, want %d", len(keys), len(theme.All()))
	}
	want := map[string]bool{}
	for _, th := range theme.All() {
		want[th.Key] = true
	}
	for _, k := range keys {
		if !want[k] {
			t.Errorf("themeKeys() returned unexpected key %q", k)
		}
	}
}

func TestColorHexTag(t *testing.T) {
	// A swatch tag is a valid tview hex tag built from the role's literal RGB.
	forge, _ := theme.ByKey("forge")
	tag := colorHexTag(forge.Color(theme.RoleAccent))
	// Forge accent is gold #ffd700.
	if tag != "[#ffd700]" {
		t.Errorf("colorHexTag(forge accent) = %q, want [#ffd700]", tag)
	}
}

func TestThemeSwatchesLabelsEveryRole(t *testing.T) {
	forge, _ := theme.ByKey("forge")
	sw := themeSwatches(forge)
	for _, sr := range swatchRoles {
		if !strings.Contains(sw, sr.label) {
			t.Errorf("themeSwatches missing role label %q\ngot: %s", sr.label, sw)
		}
		// Each row must render a colored block, not bare text.
		if !strings.Contains(sw, "███") {
			t.Errorf("themeSwatches missing swatch block ███\ngot: %s", sw)
		}
	}
}

func TestThemeDetailShowsAccessibleNoteAndGlyphs(t *testing.T) {
	// Forge is not accessible — no glyph note. (Forge is always available.)
	forge, _ := theme.ByKey("forge")
	if d := themeDetailText(forge, true); strings.Contains(d, "Accessible") {
		t.Errorf("forge detail should not claim Accessible\ngot: %s", d)
	}
	// An accessible theme shows the note and its signed glyphs.
	deut, _ := theme.ByKey("deuteranopia")
	d := themeDetailText(deut, true)
	if !strings.Contains(d, "Accessible") {
		t.Errorf("deuteranopia detail should show the Accessible note\ngot: %s", d)
	}
	if deut.GainGlyph != "" && !strings.Contains(d, deut.GainGlyph) {
		t.Errorf("deuteranopia detail should show the gain glyph %q\ngot: %s", deut.GainGlyph, d)
	}
}

func TestThemeRowLabelMarksActive(t *testing.T) {
	forge, _ := theme.ByKey("forge")
	// Marked when it's the active (open-time) theme. All shipped themes are
	// available today, so pass available=true.
	if got := themeRowLabel(forge, "forge", true); !strings.Contains(got, "current") {
		t.Errorf("themeRowLabel for active theme should mark (current)\ngot: %q", got)
	}
	// Not marked when a different theme is active.
	if got := themeRowLabel(forge, "high_contrast", true); strings.Contains(got, "current") {
		t.Errorf("themeRowLabel for non-active theme should not mark current\ngot: %q", got)
	}
	// Accessible themes carry the accessible tag.
	deut, _ := theme.ByKey("deuteranopia")
	if got := themeRowLabel(deut, "forge", true); !strings.Contains(got, "accessible") {
		t.Errorf("themeRowLabel for accessible theme should tag it\ngot: %q", got)
	}
	// An unavailable theme carries the locked marker (Phase-3 path; simulated here
	// by passing available=false).
	if got := themeRowLabel(forge, "forge", false); !strings.Contains(got, "locked") {
		t.Errorf("themeRowLabel for unavailable theme should mark it locked\ngot: %q", got)
	}
}

// TestCreateThemePickerPageRegisters mirrors the load-game add-page check: the
// picker builds without a running app and registers under its page name. We use a
// real (un-run) tview.Application + Pages so the construction path (Track closures,
// list build, initial preview) executes end to end. restoreForge resets the active
// theme afterward since the initial preview applies the default for real.
func TestCreateThemePickerPageRegisters(t *testing.T) {
	restoreForge(t)
	app := tview.NewApplication()
	pages := tview.NewPages()

	// nil engine → accountless picker (construction must not require an account).
	page := CreateThemePickerPage(app, pages, nil, "splash")
	if page == nil {
		t.Fatal("CreateThemePickerPage returned nil")
	}
	pages.AddPage(themePickerPage, page, true, true)
	if !pages.HasPage(themePickerPage) {
		t.Errorf("picker page %q not registered after AddPage", themePickerPage)
	}
}
