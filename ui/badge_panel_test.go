package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// badgeDashboard is a dashboard over a fresh game on a fresh account in a
// temp data root, with its pages, for drawing the whole screen.
func badgeDashboard(t *testing.T) (*Dashboard, *nativePages, *game.GameEngine) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	a, err := game.CreateAccount("Badge Tester")
	must(t, err)
	eng.SetAccount(a)
	must(t, eng.StartNewNamedGame("run"))
	d, pages := newTestDashboard(eng)
	d.refresh()
	return d, pages, eng
}

// settle lets the dashboard show whatever a staged change raised (an age's
// splash) and closes it, so the next thing opened is what is on screen.
func settle(d *Dashboard) {
	d.refresh()
	if d.overlayMgr.HasActive() {
		d.overlayMgr.Hide()
	}
}

// earnForTest earns a spread of badges through the engine's own reports:
// the Stone Age, the first Housing rung and three rungs of the prestige
// ladder.
func earnForTest(t *testing.T, eng *game.GameEngine) {
	t.Helper()
	must(t, eng.EnterAgeForTest("stone_age"))
	for i := 0; i < 20; i++ {
		eng.ReportForTest(config.BadgeEvBuiltLineage, "housing")
	}
	for i := 0; i < 10; i++ {
		eng.ReportForTest(config.BadgeEvPrestige, "modern_age")
	}
}

// TestBadgesCommandOpensTheCase: badges and its alias open the case; a
// family's name opens its tab; next opens the nearest rungs; a name, or a
// part of one, opens that badge's detail; and a name the account cannot
// see yet finds nothing, so a command cannot be used to fish for it.
func TestBadgesCommandOpensTheCase(t *testing.T) {
	d, _, eng := badgeDashboard(t)
	earnForTest(t, eng)
	settle(d)
	p := d.badgePanel
	for _, cmd := range []string{"badges", "achievements"} {
		d.runForTest(cmd)
		if got := d.overlayMgr.ActiveName(); got != "badges" {
			t.Fatalf("%s opened %q", cmd, got)
		}
		if !d.inputField.HasFocus() && d.app.GetFocus() != nil && d.app.GetFocus() != Primitive(d.inputField) {
			t.Errorf("%s took the keyboard from the command bar", cmd)
		}
		if !strings.Contains(sidebarText(d.sidebarActive, 20, 30), theme.Selected(" badges     ")) {
			t.Errorf("%s: the Panels list does not mark badges as open", cmd)
		}
		d.overlayMgr.Hide()
	}
	for cmd, tab := range map[string]string{
		"badges ages": "age", "badges Lineages": "lineage", "badges ladder": "ladder", "badges specials": "special",
		"badges next": caseTabNext, "badges all": caseTabAll,
	} {
		d.runForTest(cmd)
		if d.overlayMgr.ActiveName() != "badges" || p.view.tab != tab || p.view.card {
			t.Errorf("%s: panel %q, tab %q, detail %v", cmd, d.overlayMgr.ActiveName(), p.view.tab, p.view.card)
		}
	}
	for cmd, key := range map[string]string{
		"badges rock solid":          "age.stone_age",
		"badges Rock":                "age.stone_age",
		"badges housing hobbyist":    "lineage.housing.1",
		"badges serial reincarnator": "ladder.prestiges.3",
		"badges hut hoarder":         "special.hut_hoarder",
	} {
		d.runForTest("badges ages") // from another tab
		d.runForTest(cmd)
		if p.view.sel != key || !p.view.card || p.view.tab != caseTabAll {
			t.Errorf("%s: selected %q, detail %v, tab %q", cmd, p.view.sel, p.view.card, p.view.tab)
		}
	}
	// A typed command over the open case leaves it open and changes what
	// it reads: the glyph tier, the motion setting.
	d.runForTest("map glyphs ascii")
	if d.overlayMgr.ActiveName() != "badges" || p.view.tier != mapmodel.TierASCII {
		t.Errorf("map glyphs over the case: panel %q, tier %s", d.overlayMgr.ActiveName(), p.view.tier)
	}
	d.runForTest("motion off")
	if p.motion || eng.Account().MotionOn() {
		t.Error("motion off did not reach the open case")
	}
	d.runForTest("motion on")
	if !p.motion {
		t.Error("motion on did not reach the open case")
	}
	d.runForTest("map glyphs unicode")
	d.overlayMgr.Hide()

	// Names the account may not see: a later age's badge and a secret.
	views, _ := eng.Badges()
	for _, v := range views {
		if v.Hidden && v.Name != game.BadgeHiddenName {
			t.Fatalf("a withheld badge has a name: %+v", v)
		}
	}
	for _, cmd := range []string{"badges modern", "badges into the modern age", "badges liquidation", "badges cookie", "badges zzz"} {
		res := HandleCommand(cmd, eng)
		if res.OverlayName != "" || res.Type != "error" || !strings.Contains(res.Message, "No badge or family in sight") {
			t.Errorf("%s: %+v", cmd, res)
		}
	}
	// Without an account the case still opens, and says so.
	bare := game.NewGameEngine()
	if res := HandleCommand("badges", bare); res.OverlayName != "badges" {
		t.Errorf("badges with no account: %+v", res)
	}
	m := buildCase(nil, game.BadgeSummary{}, false, caseTabAll, caseLayoutFor(80, 21, caseTabAll))
	if text := renderCase(m, caseView{}).String(); !strings.Contains(text, "No account is loaded") {
		t.Errorf("the case with no account:\n%s", text)
	}
}

// TestBadgeCaseOnTheLiveScreen draws the whole screen with the case open
// at the four sizes, as the app does: the case takes every row but the
// command bar's, the command bar is still there under it, the selected
// badge's art is whole, and Esc closes the detail before the panel.
func TestBadgeCaseOnTheLiveScreen(t *testing.T) {
	d, pages, eng := badgeDashboard(t)
	earnForTest(t, eng)
	settle(d)
	for _, size := range treeSizes {
		w, h := size[0], size[1]
		d.runForTest("badges serial reincarnator")
		cells := drawnDashboard(t, d, pages, w, h)
		lines := rectText(cells, w, 0, 0, w, h)
		where := fmt.Sprintf("%dx%d", w, h)
		if !strings.Contains(lines[0], "BADGES") || !strings.Contains(lines[0], "Settler") {
			t.Errorf("%s: the case's title bar is not the screen's first row: %q", where, lines[0])
		}
		if !strings.Contains(lines[h-promptRows-1], "Esc") {
			t.Errorf("%s: the case's key bar is not over the command bar: %q", where, lines[h-promptRows-1])
		}
		if !strings.Contains(strings.Join(lines[h-promptRows:], "\n"), "Command") {
			t.Errorf("%s: the command bar is not under the case:\n%s", where, strings.Join(lines[h-promptRows:], "\n"))
		}
		text := strings.Join(lines, "\n")
		wantAll(t, fmt.Sprintf("%s: the detail", where), text, "Serial Reincarnator", "Gold · Rare · 25 pts", "Prestige 10 times.", "Rung 3 of 4 · Prestiges")
		// The gold medal is a picture in the detail's room, owning its cells.
		bp := d.badgePanel
		buf := kit.NewBuffer(w, h)
		bp.Layout(bp.Bounds())
		bp.Paint(kit.NewPainter(buf))
		room := bp.model.medalRoom.Add(bp.Bounds().Min)
		found := false
		for i, pic := range buf.Pics {
			if pic.Rect == room && int(buf.At(room.Min.X, room.Min.Y).Pic) == i+1 && !oneColour(pic.Img) {
				found = true
			}
		}
		if !found || bp.model.medalFor == nil || bp.model.medalFor.id != "ladder.prestiges.3" {
			t.Errorf("%s: the gold badge's medal is not painted in %v (%d pictures, for %v)", where, room, len(buf.Pics), bp.model.medalFor)
		}
		for i, c := range cells {
			if len(c.Runes) > 0 && c.Runes[0] != ' ' && i/w < h-promptRows && uniseg.StringWidth(string(c.Runes[0])) != 1 {
				t.Fatalf("%s: %q at %d,%d is not one cell wide", where, c.Runes[0], i%w, i/w)
			}
		}

		// Esc: the detail first, then the panel. The dashboard routes it.
		esc := newKey(kit.KeyEsc, 0, 0)
		d.app.SetFocus(d.inputField)
		if !d.badgePanel.view.card {
			t.Fatalf("%s: the detail is not open", where)
		}
		if ev := captured(d.root, *esc); ev != nil || d.badgePanel.view.card || d.overlayMgr.ActiveName() != "badges" {
			t.Errorf("%s: the first Esc did not close just the detail", where)
		}
		if captured(d.root, tk(kit.KeyRight, 0, 0)) != nil {
			t.Errorf("%s: the case did not take an arrow", where)
		}
		if ev := captured(d.root, tk(kit.KeyRune, 'b', 0)); ev == nil {
			t.Errorf("%s: the case took a letter from the prompt", where)
		}
		if ev := captured(d.root, *esc); ev != nil || d.overlayMgr.ActiveName() != "" {
			t.Errorf("%s: the second Esc did not close the panel", where)
		}
	}
}

// TestBadgeToastFitsTheBar: the toast for a badge is the badge's own art
// at its smallest, three cells in its tier's colours, then its name, its
// tier and what it is for. It is written for the bar's width: the whole of
// it where it fits, without the description where that would be cut, and
// never longer than the bar. It reads in every theme.
func TestBadgeToastFitsTheBar(t *testing.T) {
	views, _ := caseFixture()
	by := map[string]game.BadgeView{}
	for _, v := range views {
		by[v.Key] = v
	}
	render := func(markup string) (text string, cells []simCell) {
		tv := newNativeText().SetDynamicColors(true)
		tv.SetText(safeTags(markup))
		sim := newSim()
		must(t, sim.Init())
		defer sim.Fini()
		sim.SetSize(200, 1)
		tv.SetRect(0, 0, 200, 1)
		drawOn(sim, tv)
		sim.Show()
		cells, _, _ = sim.GetContents()
		return strings.TrimRight(tv.GetText(true), "\n"), append([]simCell(nil), cells...)
	}
	hut := by["special.hut_hoarder"]
	if text, _ := render(badgeToast(hut, mapmodel.TierUnicode, 0)); text != "◖∩◗ Badge earned: Hut Hoarder (gold). "+hut.Desc {
		t.Errorf("the toast at any width: %q", text)
	}
	if text, _ := render(badgeToast(hut, mapmodel.TierASCII, 0)); !strings.HasPrefix(text, "(n) Badge earned: Hut Hoarder (gold).") {
		t.Errorf("the toast in the plain tier: %q", text)
	}
	for _, size := range treeSizes {
		w := size[0]
		for _, v := range views {
			if v.Hidden {
				continue
			}
			for _, tier := range caseTiers {
				text, _ := render(badgeToast(v, tier, w))
				if n := uniseg.StringWidth(text); n > w {
					t.Errorf("%d columns: the toast for %s is %d cells: %q", w, v.Key, n, text)
				}
				if !strings.Contains(text, "Badge earned: "+v.Name) {
					t.Errorf("%d columns: the toast for %s does not name it: %q", w, v.Key, text)
				}
				if strings.Contains(text, v.Desc[:8]) && !strings.HasSuffix(text, v.Desc) {
					t.Errorf("%d columns: the toast for %s cuts its description: %q", w, v.Key, text)
				}
			}
		}
	}
	if text, _ := render(badgeToast(hut, mapmodel.TierUnicode, 80)); text != "◖∩◗ Badge earned: Hut Hoarder (gold)." {
		t.Errorf("80 columns: the toast keeps a description it has no room for: %q", text)
	}
	if text, _ := render(badgeToast(by["age.stone_age"], mapmodel.TierUnicode, 80)); text != "◖*◗ Badge earned: Rock Solid (bronze). Reach the Stone Age." {
		t.Errorf("80 columns: %q", text)
	}
	long := hut
	long.Name = strings.Repeat("Very Long Name ", 8)
	if text, _ := render(badgeToast(long, mapmodel.TierUnicode, 60)); uniseg.StringWidth(text) > 60 || !strings.Contains(text, "…") {
		t.Errorf("a name too long for the bar: %q", text)
	}

	// Colours: the mark in the tier's ink, and everything legible, in
	// every theme.
	eachThemeInline(t, func(th theme.Theme) {
		bg := theme.Of(theme.RoleBackground)
		_, cells := render(badgeToast(hut, mapmodel.TierUnicode, 0))
		for x, c := range cells {
			if len(c.Runes) == 0 || c.Runes[0] == ' ' {
				continue
			}
			fg, _, _ := decompose(c.Style)
			if r := theme.ContrastRatio(fg, bg); r < 2.99 {
				t.Errorf("theme %s: the toast's %q at %d has a contrast of %.2f", th.Key, c.Runes[0], x, r)
			}
		}
		pal := newGridPalette(0, nil)
		if fg, _, _ := decompose(cells[0].Style); fg != pal.inkOn(inkGold, false) {
			t.Errorf("theme %s: the toast's mark is %06x, not gold (%06x)", th.Key, fg.Hex(), pal.inkOn(inkGold, false).Hex())
		}
	})
}

// TestBadgeToastOnTheLiveScreen: an earned badge's toast is the notice in the
// header's top edge at every size, whole, in one row: the badge and its name
// always, its description once the edge has the room.
func TestBadgeToastOnTheLiveScreen(t *testing.T) {
	for _, size := range treeSizes {
		w, h := size[0], size[1]
		d, pages, eng := badgeDashboard(t)
		drawnDashboard(t, d, pages, w, h) // the bar gets its width
		// The report alone: an age entered for real raises its splash
		// over the bar.
		eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
		cells := drawnDashboard(t, d, pages, w, h)
		row := rectText(cells, w, 0, 0, w, 1)[0]
		whole := "◖*◗ Badge earned: Rock Solid (bronze)."
		if w >= 100 {
			whole += " Reach the Stone Age."
		}
		if !strings.Contains(row, " "+whole+" ") {
			t.Errorf("%dx%d: the header's top edge shows %q, want %q in it", w, h, row, whole)
		}
	}
}

// TestHelpListsBadgeKeys: the Help panel lists the badge case with the
// panels, its command with the commands and every key it takes, at any
// width, and the sidebar lists it at every height.
func TestHelpListsBadgeKeys(t *testing.T) {
	for _, screenW := range []int{0, 80, 144} {
		help := visible(safeTags(helpProvider(game.GameState{}, screenW)))
		at := strings.Index(help, "The badge case")
		if at < 0 {
			t.Fatal("the Help panel has no badge case section")
		}
		section := help[at:]
		if end := strings.Index(section, "═══ Shortcuts"); end > 0 {
			section = section[:end]
		}
		for _, k := range badgeKeys {
			if !strings.Contains(section, "  "+k.key) {
				t.Errorf("%d columns: the Help panel's badge case section does not list %s", screenW, k.key)
			}
		}
		flat := strings.Join(strings.Fields(help), " ")
		wantAll(t, fmt.Sprintf("%d columns: the Help panel", screenW), flat, "badges next", "badges <family>", "badges <name>", "motion [on|off]", "achievements=badges", "The badge case: every badge")
	}
	listed := ""
	for _, k := range badgeKeys {
		listed += k.key + " " + k.what.String() + " "
	}
	wantAll(t, "the badge case's key list", listed, "Arrows", "Tab", "Shift-Tab", "PgUp", "PgDn", "Home", "End", "Enter", "Esc")
	found := false
	for _, name := range sidebarPanels {
		found = found || name == "badges"
	}
	if !found {
		t.Error("the Panels list does not list badges")
	}
}

// TestMotionCommand: motion shows the setting, turns it off and on, and
// says what it holds still. Off, the map's frame is held at its first.
func TestMotionCommand(t *testing.T) {
	d, _, eng := badgeDashboard(t)
	if res := HandleCommand("motion", eng); !strings.Contains(res.Message, "Motion: on") {
		t.Errorf("motion: %+v", res)
	}
	d.runForTest("motion off")
	if eng.Account().MotionOn() || d.mapSettings().Motion {
		t.Fatal("motion off did not turn it off")
	}
	if res := HandleCommand("motion", eng); !strings.Contains(res.Message, "Motion: off") {
		t.Errorf("motion: %+v", res)
	}
	if got := animFrame(false, 40*mapAnimStep); got != 0 {
		t.Errorf("with motion off the frame is %d", got)
	}
	if got := animFrame(true, 40*mapAnimStep); got != 40 {
		t.Errorf("with motion on the frame is %d, want 40", got)
	}
	d.runForTest("map")
	d.mapPanel.update(eng.GetState())
	if d.mapPanel.set.Motion {
		t.Error("the open map does not read the motion setting")
	}
	d.overlayMgr.Hide()
	d.runForTest("motion on")
	if !eng.Account().MotionOn() {
		t.Error("motion on did not turn it on")
	}
	for _, bad := range []string{"motion sideways", "motion on off"} {
		if res := HandleCommand(bad, eng); res.Type != "error" {
			t.Errorf("%s: %+v", bad, res)
		}
	}
	// With no account the setting lasts for the session.
	bare := game.NewGameEngine()
	if res := HandleCommand("motion off", bare); !strings.Contains(res.Message, "this session") || res.MapPref.Key != "motion" {
		t.Errorf("motion off with no account: %+v", res)
	}
}

// TestThemeEffectsStayInEmptyCells: a theme's ambient effect draws only
// where the dashboard drew nothing, with clear air between it and any
// text; the same frame is the same picture; motion off draws none of it;
// and a theme without an effect is never touched.
func TestThemeEffectsStayInEmptyCells(t *testing.T) {
	eachThemeInline(t, func(th theme.Theme) {
		for _, size := range treeSizes {
			w, h := size[0], size[1]
			d, pages, eng := badgeDashboard(t)
			drawnDashboard(t, d, pages, w, h) // every box gets its size
			// One moment for the whole comparison: the effect's frame and
			// the mini map's are read off these clocks, and a slow machine
			// must not move either between two draws that are held to be
			// the same picture.
			clock := time.Now()
			d.fxNow = func() time.Time { return clock }
			d.miniMap.now = d.fxNow
			sim := newSim()
			must(t, sim.Init())
			sim.SetSize(w, h)
			scr := sim
			snap := func() []simCell {
				sim.Show()
				cells, _, _ := sim.GetContents()
				return append([]simCell(nil), cells...)
			}
			// The layout alone, then the effect's pass over it: what
			// differs is what the effect drew. The dashboard's root paints
			// the effect itself, so the layout alone is the root with the
			// effect taken off for that draw.
			fx := d.Root().(*fxRoot)
			layoutOnly := func() {
				after := fx.after
				fx.after = nil
				drawOn(scr, d.root)
				fx.after = after
			}
			scr.Clear()
			layoutOnly()
			still := snap()
			must(t, eng.Account().SetMotion(false))
			d.drawEffect(scr, 0, 0, w, h)
			if off := snap(); d.fxOn.Load() || fmt.Sprint(off) != fmt.Sprint(still) {
				t.Errorf("theme %s at %dx%d: an effect draws with motion off", th.Key, w, h)
			}
			must(t, eng.Account().SetMotion(true))
			bg := theme.Of(theme.RoleBackground)
			drawn := 0
			frames := 4
			if th.Effect != "" {
				frames = 130 // two tears of the Glitch theme
			}
			for n := 0; n < frames; n++ {
				scr.Clear()
				layoutOnly()
				still = snap() // the mini map moves on its own clock
				d.fxStart = clock.Add(-time.Duration(n) * mapAnimStep)
				d.drawEffect(scr, 0, 0, w, h)
				direct := snap()
				if n < 3 {
					// The root's own paint of the effect, through the kit,
					// is the same picture cell for cell.
					scr.Clear()
					drawOn(scr, fx)
					for i, c := range snap() {
						if c.Runes[0] != direct[i].Runes[0] || c.Style != direct[i].Style {
							t.Fatalf("theme %s at %dx%d, frame %d: the root painted %q at %d,%d where the effect alone draws %q",
								th.Key, w, h, n, c.Runes[0], i%w, i/w, direct[i].Runes[0])
						}
					}
				}
				for i, c := range direct {
					was := still[i]
					if c.Runes[0] == was.Runes[0] && c.Style == was.Style {
						continue
					}
					drawn++
					x, y := i%w, i/w
					if y >= h-promptRows {
						t.Fatalf("theme %s at %dx%d: the effect drew in the command bar (row %d)", th.Key, w, h, y)
					}
					for dx := -fxMargin; dx <= fxMargin; dx++ {
						if x+dx < 0 || x+dx >= w {
							continue
						}
						near := still[y*w+x+dx]
						if _, back, _ := decompose(near.Style); near.Runes[0] != ' ' || back != bg {
							t.Fatalf("theme %s at %dx%d: the effect drew %q at %d,%d, %d cells from %q", th.Key, w, h, c.Runes[0], x, y, dx, near.Runes[0])
						}
					}
					if fg, back, _ := decompose(c.Style); back != bg || fg == bg {
						t.Fatalf("theme %s: the effect's cell at %d,%d is %06x on %06x", th.Key, x, y, fg.Hex(), back.Hex())
					}
					if uniseg.StringWidth(string(c.Runes[0])) != 1 {
						t.Fatalf("theme %s: the effect drew %q, which is not one cell wide", th.Key, c.Runes[0])
					}
				}
			}
			sim.Fini()
			if (th.Effect != "") != (drawn > 0) {
				t.Errorf("theme %s (effect %q) at %dx%d: the effect drew %d cells", th.Key, th.Effect, w, h, drawn)
			}
			if (th.Effect != "") != d.fxOn.Load() {
				t.Errorf("theme %s: effect on is %v", th.Key, d.fxOn.Load())
			}
		}
	})
	// A frame is a function of its number.
	must(t, theme.SetActive("source"))
	frame := func(n int, plain bool) string {
		sim := newSim()
		must(t, sim.Init())
		defer sim.Fini()
		sim.SetSize(40, 12)
		scr := sim
		scr.Clear()
		drawThemeEffect(scr, 0, 0, 40, 12, theme.EffectRain, n, plain)
		sim.Show()
		cells, _, _ := sim.GetContents()
		var sb strings.Builder
		for _, c := range cells {
			if plain && c.Runes[0] >= 0x80 {
				t.Fatalf("the rain draws %q in the plain tier", c.Runes[0])
			}
			sb.WriteRune(c.Runes[0])
		}
		return sb.String()
	}
	if frame(5, false) != frame(5, false) || frame(5, false) == frame(40, false) {
		t.Error("the rain is not a function of its frame")
	}
	frame(5, true)
	// A window with no room, as a terminal a row or two tall leaves the
	// dashboard once the command bar is taken off: nothing is drawn, and
	// nothing breaks.
	for _, size := range [][2]int{{0, 0}, {40, 0}, {40, -2}, {-3, 5}} {
		for _, effect := range []string{theme.EffectRain, theme.EffectGlitch, theme.EffectEmbers, theme.EffectGreenbar, theme.EffectPrism} {
			drawThemeEffect(newSim(), 0, 0, size[0], size[1], effect, 7, false)
		}
	}
}

// TestHackerThemesBelongToTheirBadges: each theme a badge gives names that
// badge, and each badge that gives a theme names a theme that names it
// back. The picker's hint for them does not say how they are earned.
func TestHackerThemesBelongToTheirBadges(t *testing.T) {
	gives := map[string]string{}
	for _, def := range rules.Core().Badges() {
		if def.Reward.Theme == "" {
			continue
		}
		th, ok := theme.ByKey(def.Reward.Theme)
		if !ok {
			t.Errorf("%s gives the theme %q, which does not exist", def.Key, def.Reward.Theme)
			continue
		}
		if th.UnlockBadge != def.Key {
			t.Errorf("%s gives %s, which says it comes from %q", def.Key, th.Key, th.UnlockBadge)
		}
		gives[th.Key] = def.Key
		if def.Reveal.Kind == config.BadgeSecret {
			hint := strings.ToLower(th.UnlockHint.Plain())
			for _, word := range strings.Fields(strings.ToLower(def.Name + " " + def.Desc)) {
				if len(word) > 4 && word != "badge" && strings.Contains(hint, strings.Trim(word, ".,")) {
					t.Errorf("the hint for %s (%q) gives away its secret badge with %q", th.Key, th.UnlockHint.Plain(), word)
				}
			}
		}
	}
	for _, th := range theme.All() {
		if th.UnlockBadge != "" && gives[th.Key] != th.UnlockBadge {
			t.Errorf("the theme %s waits for the badge %q, which does not give it", th.Key, th.UnlockBadge)
		}
	}
	if gives["source"] != "special.touched_by_the_source" || gives["glitch"] != "special.creative_accounting" {
		t.Errorf("the hacker themes: %v", gives)
	}
}

// TestStatusBarWearsTheTitle: once an account's badge score holds a title
// past the first, the header shows it beside the account's name in its top
// edge. The first title shows the name alone.
func TestStatusBarWearsTheTitle(t *testing.T) {
	st := game.NewGameEngine().GetState()
	st.AccountStats = &game.AccountStatsView{DisplayName: "Ada", BadgeSummary: game.BadgeSummary{Title: "Settler"}}
	top := func(v headerView, w int) string { return renderHeader(v, w, 2).Lines()[0] }
	if v := statusLine(st, ""); v.Name != "Ada" || v.Worn != "" || !strings.Contains(top(v, 100), " Ada ") || strings.Contains(top(v, 100), "Settler") {
		t.Errorf("the first title: %q", top(v, 100))
	}
	st.AccountStats.BadgeSummary = game.BadgeSummary{Title: "Magistrate", TitleRank: 2}
	v := statusLine(st, "")
	if line := top(v, 100); !strings.Contains(line, " Ada · Magistrate ") {
		t.Errorf("a wide header: %q", line)
	}
	for w := 40; w <= 200; w += 7 {
		if line := top(v, w); runeLen(line) != w || !strings.Contains(line, " Ada") {
			t.Errorf("%d columns: the top edge is %d cells: %q", w, runeLen(line), line)
		}
	}
}

// TestHackerThemesInTheList: until its badge is earned a badge's theme is
// locked, the list and the picker say only that a secret badge gives it,
// and the theme command refuses it with the same words. Earning the badge
// unlocks it at once.
func TestHackerThemesInTheList(t *testing.T) {
	keepTheme(t)
	_, _, eng := badgeDashboard(t)
	acct := eng.Account()
	st := eng.GetState()
	for key, event := range map[string]string{"source": config.BadgeEvSaveElite, "glitch": config.BadgeEvSaveModified} {
		th, _ := theme.ByKey(key)
		line := func() string {
			for _, l := range strings.Split(untag(cmdThemeList(acct, &st).Message), "\n") {
				if strings.Contains(l, th.Name.Plain()) {
					return l
				}
			}
			t.Fatalf("theme list has no %s", th.Name.Plain())
			return ""
		}
		if got := line(); !strings.Contains(got, "Given by a secret badge") {
			t.Errorf("the locked %s in the list: %q", key, got)
		}
		if got := untag(themeDetailText(th, themeAvailable(acct, th), &st)); !strings.Contains(got, "Locked") || !strings.Contains(got, "Given by a secret badge") {
			t.Errorf("the locked %s in the picker: %q", key, got)
		}
		if res := HandleCommand("theme "+key, eng); res.Type != "error" || theme.Active().Key == key {
			t.Errorf("theme %s before its badge: %+v", key, res)
		}
		eng.ReportForTest(event, "")
		if got := line(); strings.Contains(got, "secret badge") {
			t.Errorf("the unlocked %s in the list: %q", key, got)
		}
		if res := HandleCommand("theme "+key, eng); res.Type == "error" || theme.Active().Key != key {
			t.Errorf("theme %s after its badge: %+v (active %s)", key, res, theme.Active().Key)
		}
	}
}
