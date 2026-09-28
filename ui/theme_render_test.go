package ui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// renderScreens draws the key screens into a SimulationScreen wrapped by
// theme.WrapScreen — exactly the path the real app takes — and returns the
// resulting cells per screen name.
func renderScreens(t *testing.T, w, h int) map[string][]tcell.SimCell {
	t.Helper()
	engine := game.NewGameEngine()
	app := tview.NewApplication()
	pages := tview.NewPages()
	d := NewDashboard(app, engine, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.refresh()

	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sim.Fini)
	sim.SetSize(w, h)
	screen := theme.WrapScreen(sim)

	draw := func() []tcell.SimCell {
		screen.Clear()
		pages.SetRect(0, 0, w, h)
		pages.Draw(screen)
		screen.Show()
		cells, _, _ := sim.GetContents()
		out := make([]tcell.SimCell, len(cells))
		copy(out, cells)
		return out
	}

	out := map[string][]tcell.SimCell{}
	out["dashboard"] = draw()

	// The prompt with ghost text: "adv" typed, "ance" drawn dim after it.
	d.inputField.SetText("adv")
	d.inputField.Focus(func(tview.Primitive) {})
	out["dashboard_ghost"] = draw()
	if !d.inputField.atEnd || d.inputField.Ghost() != "ance" {
		t.Fatalf("no ghost text drawn for %q (at end %v, ghost %q)", "adv", d.inputField.atEnd, d.inputField.Ghost())
	}
	d.inputField.Blur()
	d.inputField.SetText("")

	state := engine.GetState()
	for _, name := range []string{"stats", "help", "citymap", "worldmap"} {
		if !d.overlayMgr.Show(name, state) {
			t.Fatalf("overlay %q not registered", name)
		}
		out["overlay:"+name] = draw()
		d.overlayMgr.Hide()
	}

	// The Plan panel with an item of each status (ready, blocked on storage,
	// waiting with a progress bar), the selection chip and a feedback line.
	for _, it := range []struct {
		key string
		n   int
	}{{"hut", 3}, {"story_circle", 1}, {"wood_camp", 1}, {"stash", 2}} {
		if _, err := engine.PlanAddBuild(it.key, it.n); err != nil {
			t.Fatalf("plan %s: %v", it.key, err)
		}
	}
	d.planPanel.note, d.planPanel.noteGood = "Removed 1 × Farm.", true
	if !d.overlayMgr.Show("plan", engine.GetState()) {
		t.Fatal("overlay plan not registered")
	}
	out["overlay:plan"] = draw()
	d.overlayMgr.Hide()
	engine.PlanClear()
	d.planPanel.reset()

	picker := CreateThemePickerPage(app, pages, nil, "dashboard")
	pages.AddPage(themePickerPage, picker, true, true)
	out["theme_picker"] = draw()
	pages.RemovePage(themePickerPage)

	showWipeConfirmation(app, pages, engine, func() {}, "test")
	out["danger_modal"] = draw()
	pages.RemovePage("wipe_confirm")

	// Filled-button modals: Negative-fill ENDURE/SUCCUMB and Positive-fill ACCEPT,
	// with their escaped "[E]"/"[S]"/"[A]" shortcut labels.
	d.showCatastropheModal("stone_era")
	out["catastrophe_modal"] = draw()
	d.closeCatastropheModal()
	d.showAncientMemoryModal("fire_mastery", "")
	out["ancient_memory_modal"] = draw()
	d.closeAncientMemoryModal()

	// Harbinger: the panel (severity, keycaps, affordable and unaffordable
	// costs), the Epoch overlay's harbinger lines and the status-bar badge.
	if err := engine.SummonHarbingerForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	d.refresh()
	out["dashboard_harbinger_badge"] = draw()
	for _, name := range []string{"harbinger", "epoch"} {
		if !d.overlayMgr.Show(name, engine.GetState()) {
			t.Fatalf("overlay %q not registered", name)
		}
		out["overlay:"+name+"_harbinger"] = draw()
		d.overlayMgr.Hide()
	}

	// The Last Passage: the Cosmic Era panel, the modal variant with Succumb
	// open and closed (chip-filled), and its status-bar badge.
	if err := engine.SummonHarbingerForTest("galactic_age"); err != nil {
		t.Fatal(err)
	}
	if !d.overlayMgr.Show("harbinger", engine.GetState()) {
		t.Fatal("overlay harbinger not registered")
	}
	out["overlay:harbinger_last_passage"] = draw()
	d.overlayMgr.Hide()
	if err := engine.ForceLastPassageForTest("galactic_age"); err != nil {
		t.Fatal(err)
	}
	d.refresh()
	out["last_passage_modal"] = draw()
	d.closeCatastropheModal()
	out["dashboard_last_passage_badge"] = draw()
	engine.SetCosmicLegacyForTest(true)
	d.showCatastropheModal(config.LastPassageKey)
	out["last_passage_modal_succumb_closed"] = draw()
	d.closeCatastropheModal()

	return out
}

// pixelGlyph reports runes that are legitimately drawn with fg == bg: the
// half-block map pixels ('▄' with equal upper/lower pixels is a solid fill).
func pixelGlyph(r rune) bool { return r == '▄' || r == '▀' }

// TestThemeRender_NoInvisibleText draws the dashboard, several overlays
// (including the citymap and worldmap), the theme picker and a danger modal
// under EVERY theme and asserts:
//
//  1. no non-space glyph has fg == bg (invisible text — e.g. [red] on a red
//     panel, or dark ink on a canvas that failed to repaint);
//  2. every cell has a concrete RGB background — nothing falls through to the
//     terminal default and no late-bound sentinel escapes the screen wrapper.
func TestThemeRender_NoInvisibleText(t *testing.T) {
	restoreForge(t)
	const w, h = 160, 48
	for _, th := range theme.All() {
		th := th
		t.Run(th.Key, func(t *testing.T) {
			if err := theme.SetActive(th.Key); err != nil {
				t.Fatal(err)
			}
			for name, cells := range renderScreens(t, w, h) {
				bad := 0
				for i, c := range cells {
					fg, bg, _ := c.Style.Decompose()
					if !bg.IsRGB() {
						if bad < 5 {
							t.Errorf("%s: cell (%d,%d) %q has non-RGB bg %v", name, i%w, i/w, string(c.Runes), bg)
						}
						bad++
						continue
					}
					if len(c.Runes) == 0 || c.Runes[0] == ' ' || pixelGlyph(c.Runes[0]) {
						continue
					}
					if !fg.IsRGB() {
						if bad < 5 {
							t.Errorf("%s: cell (%d,%d) %q has non-RGB fg %v", name, i%w, i/w, string(c.Runes), fg)
						}
						bad++
						continue
					}
					if fg.Hex() == bg.Hex() {
						if bad < 5 {
							t.Errorf("%s: cell (%d,%d) %q is invisible (fg == bg == %06x)", name, i%w, i/w, string(c.Runes), fg.Hex())
						}
						bad++
					}
				}
				if bad >= 5 {
					t.Errorf("%s: %d bad cells total", name, bad)
				}
			}
		})
	}
}

// TestThemeRender_CanvasIsThemeBackground checks the background decision end to
// end: with no content drawn over it, the dashboard's canvas cells carry the
// ACTIVE theme's Background — including right after a live switch, for widgets
// that were constructed under a different theme.
func TestThemeRender_CanvasIsThemeBackground(t *testing.T) {
	restoreForge(t)
	_ = theme.SetActive("forge")
	engine := game.NewGameEngine()
	app := tview.NewApplication()
	pages := tview.NewPages()
	d := NewDashboard(app, engine, pages) // constructed under Forge
	pages.AddPage("dashboard", d.Root(), true, true)
	d.refresh()

	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(120, 40)
	screen := theme.WrapScreen(sim)

	for _, key := range []string{"daylight", "forge", "high_contrast_light"} {
		_ = theme.SetActive(key) // live switch; no widget is rebuilt
		screen.Clear()
		pages.SetRect(0, 0, 120, 40)
		pages.Draw(screen)
		screen.Show()
		cells, _, _ := sim.GetContents()
		want := theme.Color(theme.RoleBackground).Hex()
		counts := map[int32]int{}
		for _, c := range cells {
			_, bg, _ := c.Style.Decompose()
			counts[bg.Hex()]++
		}
		// The canvas must dominate: most cells are plain Background.
		if counts[want]*2 < len(cells) {
			t.Errorf("%s: only %d/%d cells carry Background %06x; bg histogram %v", key, counts[want], len(cells), want, topHex(counts))
		}
	}
}

func topHex(m map[int32]int) string {
	var parts []string
	for k, v := range m {
		if v > 50 {
			parts = append(parts, fmt.Sprintf("%06x:%d", k, v))
		}
	}
	return strings.Join(parts, " ")
}

// TestThemeRender_DumpANSI writes each screen as an ANSI truecolor text file for
// eyeballing (cat it in a truecolor terminal). Opt-in:
//
//	THEME_RENDER_DUMP=/tmp/dump go test ./ui/ -run TestThemeRender_DumpANSI
func TestThemeRender_DumpANSI(t *testing.T) {
	dir := os.Getenv("THEME_RENDER_DUMP")
	if dir == "" {
		t.Skip("set THEME_RENDER_DUMP=<dir> to dump ANSI renders")
	}
	restoreForge(t)
	const w, h = 160, 48
	for _, key := range []string{"forge", "daylight", "high_contrast_light", "parchment"} {
		_ = theme.SetActive(key)
		for name, cells := range renderScreens(t, w, h) {
			base := fmt.Sprintf("%s/%s_%s", dir, key, strings.ReplaceAll(name, ":", "_"))
			if err := os.WriteFile(base+".html", []byte(cellsHTML(cells, w, h, key+" / "+name)), 0o644); err != nil {
				t.Fatal(err)
			}
			var sb strings.Builder
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					c := cells[y*w+x]
					fg, bg, _ := c.Style.Decompose()
					fr, fgc, fb := fg.RGB()
					br, bgc, bb := bg.RGB()
					r := " "
					if len(c.Runes) > 0 {
						r = string(c.Runes)
					}
					fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%s", fr, fgc, fb, br, bgc, bb, r)
				}
				sb.WriteString("\x1b[0m\n")
			}
			if err := os.WriteFile(base+".ans", []byte(sb.String()), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// cellsHTML renders cells as a static HTML page (one span per cell run) so a
// render can be eyeballed in a browser or screenshotted for a PR.
func cellsHTML(cells []tcell.SimCell, w, h int, title string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<!doctype html><meta charset=utf-8><title>%s</title>"+
		"<body style=\"margin:0;background:#888\"><pre style=\"margin:0;font:13px/15px Menlo,monospace\">", title)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			fg, bg, attr := c.Style.Decompose()
			r := " "
			if len(c.Runes) > 0 {
				r = string(c.Runes)
			}
			r = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(r)
			weight := ""
			if attr&tcell.AttrBold != 0 {
				weight = ";font-weight:bold"
			}
			fmt.Fprintf(&sb, "<span style=\"color:#%06x;background:#%06x%s\">%s</span>", fg.Hex(), bg.Hex(), weight, r)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("</pre>")
	return sb.String()
}
