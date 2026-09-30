package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// mapTestDashboard builds a dashboard on a fresh game in a temp data dir,
// with a named account (acct false: no account at all).
func mapTestDashboard(t *testing.T, acct bool) (*Dashboard, *game.GameEngine) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	if acct {
		a, err := game.CreateNamedAccount("Map Tester")
		if err != nil {
			t.Fatal(err)
		}
		eng.SetAccount(a)
	}
	app := tview.NewApplication()
	pages := tview.NewPages()
	d := NewDashboard(app, eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.refresh()
	return d, eng
}

// run types a line into the prompt and presses Enter.
func (d *Dashboard) runForTest(line string) {
	d.inputField.SetText(line)
	d.submitInput()
}

// drawMapPanel draws the open Map panel on a w x h screen and returns it as text.
func drawMapPanel(t *testing.T, d *Dashboard, w, h int) string {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	d.mapPanel.SetRect(0, 0, w, h)
	d.mapPanel.Draw(sim)
	sim.Show()
	return simText(sim)
}

func simText(sim tcell.SimulationScreen) string {
	cells, w, h := sim.GetContents()
	var sb strings.Builder
	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			if len(c.Runes) == 0 {
				line.WriteByte(' ')
			} else {
				line.WriteRune(c.Runes[0])
			}
		}
		sb.WriteString(strings.TrimRight(line.String(), " "))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func key(r rune) *tcell.EventKey { return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone) }

// TestMapAliasesOpenThePanel: map, citymap and worldmap all open the one Map
// panel; worldmap opens roguelike on the known world (its region zoom).
func TestMapAliasesOpenThePanel(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	for _, cmd := range []string{"map", "citymap", "worldmap"} {
		d.runForTest(cmd)
		if got := d.overlayMgr.ActiveName(); got != "map" {
			t.Fatalf("%s opened %q, want the map panel", cmd, got)
		}
		txt := drawMapPanel(t, d, 120, 40)
		region := strings.Contains(txt, "REGION")
		if region != (cmd == "worldmap") {
			t.Errorf("%s: region zoom shown = %v\n%s", cmd, region, txt)
		}
		d.overlayMgr.Hide()
	}
}

// TestMapPanelCycleKeys: s cycles the style and g the glyph tier, and both
// are saved to the account.
func TestMapPanelCycleKeys(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	d.runForTest("map")
	p := d.mapPanel
	if p.set.Style != "roguelike" || p.set.Tier != mapmodel.TierUnicode {
		t.Fatalf("defaults = %s/%s, want roguelike/unicode", p.set.Style, p.set.Tier)
	}
	names := all.Registry().Names()
	for i := 1; i <= len(names); i++ {
		p.handleKey(key('s'))
		want := names[i%len(names)]
		if p.set.Style != want {
			t.Fatalf("after %d s: style %s, want %s", i, p.set.Style, want)
		}
		if got, _, _ := eng.Account().MapPrefs(); got != want {
			t.Errorf("after %d s: account style %q, want %q", i, got, want)
		}
		if txt := drawMapPanel(t, d, 120, 40); !strings.Contains(txt, "style: "+styleTitle(all.Registry(), want)) {
			t.Errorf("key bar does not name the style %s:\n%s", want, txt)
		}
	}
	for _, want := range []string{"nerd", "ascii", "unicode"} {
		p.handleKey(key('g'))
		if p.set.Tier.String() != want {
			t.Fatalf("g: tier %s, want %s", p.set.Tier, want)
		}
		if _, got, _ := eng.Account().MapPrefs(); got != want {
			t.Errorf("g: account glyphs %q, want %q", got, want)
		}
		drawMapPanel(t, d, 120, 40)
	}
}

// TestMapInspectStagesCommand: Enter on an inspected target closes the
// panel and puts its command in the prompt, unrun, for the player's Enter.
func TestMapInspectStagesCommand(t *testing.T) {
	for _, style := range all.Registry().Names() {
		t.Run(style, func(t *testing.T) {
			d, eng := mapTestDashboard(t, true)
			if err := eng.Account().SetMapStyle(style); err != nil {
				t.Fatal(err)
			}
			// A civ met: something on the map with a command in every style.
			if err := eng.MeetFactionForTest("riverlands_tribes", 30); err != nil {
				t.Fatal(err)
			}
			d.runForTest("map")
			p := d.mapPanel
			drawMapPanel(t, d, 160, 48)
			var cmd string
			for i := 0; i < 40 && cmd == ""; i++ {
				p.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
				drawMapPanel(t, d, 160, 48)
				if in, ok := p.current().Inspect(p.frame()); ok {
					cmd = in.Command
				}
			}
			if cmd == "" {
				t.Fatal("Tab never reached a target with a command")
			}
			p.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			if d.overlayMgr.ActiveName() != "" {
				t.Errorf("the panel is still open after Enter")
			}
			if got := d.inputField.GetText(); got != cmd {
				t.Errorf("prompt holds %q, want the staged %q", got, cmd)
			}
			if !d.inputField.comp.complete(cmd) {
				t.Errorf("staged %q is not a whole command", cmd)
			}
		})
	}
}

// TestMapIconsHintOnce: the first open shows the icons hint and records it
// on the account; the next open does not.
func TestMapIconsHintOnce(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	d.runForTest("map")
	if txt := drawMapPanel(t, d, 120, 40); !strings.Contains(txt, mapIconsHint) {
		t.Errorf("first open: no icons hint in\n%s", txt)
	}
	if _, _, shown := eng.Account().MapPrefs(); !shown {
		t.Error("the hint was not recorded on the account")
	}
	d.overlayMgr.Hide()
	d.runForTest("map")
	if txt := drawMapPanel(t, d, 120, 40); strings.Contains(txt, mapIconsHint) {
		t.Errorf("second open shows the hint again")
	}
}

// TestMapSettingsPersist: the settings live on the account, so they survive
// a save and load, and an account switch swaps them.
func TestMapSettingsPersist(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	first := eng.Account()
	for _, line := range []string{"map style skyline", "map glyphs ascii"} {
		d.runForTest(line)
	}
	if s := d.mapSettings(); s.Style != "skyline" || s.Tier != mapmodel.TierASCII {
		t.Fatalf("after the commands: %s/%s", s.Style, s.Tier)
	}
	if err := eng.SaveGame("maptest"); err != nil {
		t.Fatal(err)
	}
	d.runForTest("map style roguelike")
	if err := eng.LoadGame("maptest"); err != nil {
		t.Fatal(err)
	}
	if s := d.mapSettings(); s.Style != "roguelike" || s.Tier != mapmodel.TierASCII {
		t.Errorf("a load changed the account's map settings: %s/%s", s.Style, s.Tier)
	}
	d.runForTest("map style skyline")

	second, err := game.CreateAccount("Second")
	if err != nil {
		t.Fatal(err)
	}
	eng.SetAccount(second)
	if s := d.mapSettings(); s.Style != "roguelike" || s.Tier != mapmodel.TierUnicode {
		t.Errorf("a new account has %s/%s, want the defaults", s.Style, s.Tier)
	}
	back, err := game.SwitchAccount(first.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	eng.SetAccount(back)
	if s := d.mapSettings(); s.Style != "skyline" || s.Tier != mapmodel.TierASCII {
		t.Errorf("switching back (read from disk): %s/%s, want skyline/ascii", s.Style, s.Tier)
	}
}

// TestMapSettingsWithoutAccount: with no account the setting commands
// still work, for the session, and say so.
func TestMapSettingsWithoutAccount(t *testing.T) {
	d, eng := mapTestDashboard(t, false)
	for _, line := range []string{"map style skyline", "map glyphs ascii", "minimap off"} {
		d.runForTest(line)
	}
	if s := d.mapSettings(); s.Style != "skyline" || s.Tier != mapmodel.TierASCII || s.Minimap {
		t.Errorf("session settings %+v, want skyline, ascii, mini map off", s)
	}
	res := HandleCommand("map style roguelike", eng)
	if res.Type != "info" || res.MapPref.Key != "style" || !strings.Contains(res.Message, "this session") {
		t.Errorf("map style with no account: %+v", res)
	}
}

// TestMapSettingWordsMatchRegistries: the registry's literal words for map
// style (and its shortcut style) and map glyphs are the style registry's
// and the glyph tiers'.
func TestMapSettingWordsMatchRegistries(t *testing.T) {
	c := lookup(registry(), "map")
	want := map[string][]string{"style": all.Registry().Names(), "glyphs": mapmodel.TierNames}
	for sub, names := range want {
		s := lookup(c.Subs, sub)
		if s == nil || len(s.Args) != 1 {
			t.Fatalf("map %s is not registered with one slot", sub)
		}
		if strings.Join(s.Args[0].Words, ",") != strings.Join(names, ",") {
			t.Errorf("map %s takes %v, want %v", sub, s.Args[0].Words, names)
		}
	}
	st := lookup(registry(), "style")
	if st == nil || len(st.Args) != 1 {
		t.Fatal("style is not registered with one slot")
	}
	if got, want := strings.Join(st.Args[0].Words, ","), strings.Join(all.Registry().Names(), ","); got != want {
		t.Errorf("style takes %s, want %s", got, want)
	}
}

// TestStyleShortcut: style is map style under a shorter name: it shows the
// style bare and saves a new one to the account.
func TestStyleShortcut(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	d.runForTest("style skyline")
	if got, _, _ := eng.Account().MapPrefs(); got != "skyline" {
		t.Errorf("style skyline saved %q, want skyline", got)
	}
	if res := HandleCommand("style", eng); res.Type != "info" || !strings.Contains(res.Message, "Map style: Skyline") {
		t.Errorf("bare style: %+v", res)
	}
	if res := HandleCommand("style cubist", eng); res.Type != "error" {
		t.Errorf("style cubist: %+v", res)
	}
}
