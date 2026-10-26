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
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// renderScreens draws the key screens into a SimulationScreen wrapped by
// theme.WrapScreen — exactly the path the real app takes — and returns the
// resulting cells per screen name.
func renderScreens(t *testing.T, w, h int) map[string][]tcell.SimCell {
	t.Helper()
	engine := game.NewGameEngine()
	d, pages := newTestDashboard(engine)
	miniOn(t, d) // the mini map's colours are part of the sweep
	d.refresh()

	sim := newSimScreen(t, w, h)
	screen := theme.WrapScreen(sim)

	draw := func() []tcell.SimCell {
		screen.Clear()
		pages.SetRect(0, 0, w, h)
		pages.Draw(screen)
		screen.Show()
		return simCells(sim).Cells()
	}

	out := map[string][]tcell.SimCell{}
	// shot shows the overlay name, draws it into out under key and hides it.
	shot := func(key, name string) {
		if !d.overlayMgr.Show(name, engine.GetState()) {
			t.Fatalf("overlay %q not registered", name)
		}
		out[key] = draw()
		d.overlayMgr.Hide()
	}
	out["dashboard"] = draw()
	if !d.mapDock.shown {
		t.Fatalf("no mini map on the %dx%d dashboard", w, h)
	}

	// The prompt with ghost text: "adv" typed, "ance" drawn dim after it.
	d.inputField.SetText("adv")
	d.inputField.Focus(func(tview.Primitive) {})
	out["dashboard_ghost"] = draw()
	if !d.inputField.atEnd || d.inputField.Ghost() != "ance" {
		t.Fatalf("no ghost text drawn for %q (at end %v, ghost %q)", "adv", d.inputField.atEnd, d.inputField.Ghost())
	}
	d.inputField.Blur()
	d.inputField.SetText("")

	for _, name := range []string{"stats", "help", "trade", "milestones", "logs", "history", "buildings", "wonders"} {
		shot("overlay:"+name, name)
	}

	// The Map panel and the mini map in every glyph tier.
	saved := d.mapLocal
	for _, tier := range mapmodel.TierNames {
		tr, _ := mapmodel.ParseTier(tier)
		d.mapLocal = &mapSettings{Tier: tr, Minimap: true, HintShown: true}
		d.refresh()
		out["dashboard:"+tier] = draw()
		shot("overlay:map/"+tier, "map")
	}
	d.mapLocal = saved

	// The Factions panel with trade deals: a civ with open offers (one of
	// them unaffordable) and a hostile one that offers none.
	for _, c := range []struct {
		key     string
		opinion int
	}{{"riverlands_tribes", 30}, {"ironhold_clans", -80}} {
		must(t, engine.MeetFactionForTest(c.key, c.opinion))
	}
	if len(engine.GetState().Diplomacy.Factions["riverlands_tribes"].Deals) == 0 {
		t.Fatal("the Factions panel sweep has no deals to draw")
	}
	shot("overlay:factions_deals", "factions")

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
	shot("overlay:plan", "plan")
	engine.PlanClear()
	d.planPanel.reset()

	picker := CreateThemePickerPage(d.app, pages, nil, "dashboard")
	pages.AddPage(themePickerPage, picker, true, true)
	out["theme_picker"] = draw()
	pages.RemovePage(themePickerPage)

	showWipeSavesConfirmation(pages, "Test Account", func() {}, func() {})
	out["danger_modal"] = draw()
	pages.RemovePage(accountsSavesConfirm)

	// Filled-button modals: Negative-fill ENDURE/SUCCUMB and Positive-fill ACCEPT,
	// with their escaped "[E]"/"[S]"/"[A]" shortcut labels.
	d.showCatastropheModal("stone_era")
	out["catastrophe_modal"] = draw()
	d.closeCatastropheModal()
	d.showAncientMemoryModal("fire_mastery", "")
	out["ancient_memory_modal"] = draw()
	d.closeAncientMemoryModal()

	// The army: the Army panel with a garrison (the mitigation lines and the
	// saved-this-run tally after an Endure), and the catastrophe modal whose
	// Endure preview counts that garrison. Endured afterwards so nothing stays
	// pending for the harbinger screens below, which then count it too.
	engine.SetGarrisonForTest("iron_age", 640000)
	must(t, engine.ForceCatastropheForTest())
	d.showCatastropheModal(engine.GetState().PendingCatastrophe)
	out["catastrophe_modal_garrison"] = draw()
	d.closeCatastropheModal()
	must(t, engine.Endure())
	engine.SetGarrisonForTest("iron_age", 640000)
	shot("overlay:army_garrison", "army")

	// Harbinger: the panel (severity, keycaps, affordable and unaffordable
	// costs), the Epoch overlay's harbinger lines and the status-bar badge.
	must(t, engine.SummonHarbingerForTest("bronze_age"))
	d.refresh()
	out["dashboard_harbinger_badge"] = draw()
	for _, name := range []string{"harbinger", "epoch"} {
		shot("overlay:"+name+"_harbinger", name)
	}

	// The Last Passage: the Cosmic Era panel, the modal variant with Succumb
	// open and closed (chip-filled), and its status-bar badge.
	must(t, engine.SummonHarbingerForTest("galactic_age"))
	shot("overlay:harbinger_last_passage", "harbinger")
	must(t, engine.ForceLastPassageForTest("galactic_age"))
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

// pixelGlyph reports runes that are legitimately drawn with fg == bg:
// half-block pixels (wonder icons, the skyline map), where '▄' with equal
// upper and lower pixels is a solid fill.
func pixelGlyph(r rune) bool { return r == '▄' || r == '▀' }

// TestThemeRender_NoInvisibleText draws the dashboard, several overlays
// (including the Map panel and the mini map in every style and glyph tier),
// the theme picker and a danger modal
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
		t.Run(th.Key, func(t *testing.T) {
			must(t, theme.SetActive(th.Key))
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
	d, pages := newTestDashboard(nil) // constructed under Forge
	d.refresh()

	sim := newSimScreen(t, 120, 40)
	screen := theme.WrapScreen(sim)

	for _, key := range []string{"daylight", "forge", "high_contrast_light"} {
		_ = theme.SetActive(key) // live switch; no widget is rebuilt
		screen.Clear()
		pages.SetRect(0, 0, 120, 40)
		pages.Draw(screen)
		screen.Show()
		cells := simCells(sim).Cells()
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
			must(t, os.WriteFile(base+".html", []byte(cellsHTML(cells, w, h, key+" / "+name)), 0o644))
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
			must(t, os.WriteFile(base+".ans", []byte(sb.String()), 0o644))
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
