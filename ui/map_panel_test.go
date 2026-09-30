package ui

import (
	"strings"
	"testing"
	"time"

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

// TestMapPanelLettersAreThePrompts: no printable key is a map control any
// more (the old s and g keys went to the prompt's letters); the key bar
// names the setting commands instead, and they change the open panel.
func TestMapPanelLettersAreThePrompts(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	d.runForTest("map")
	p := d.mapPanel
	if p.set.Style != "roguelike" || p.set.Tier != mapmodel.TierUnicode {
		t.Fatalf("defaults = %s/%s, want roguelike/unicode", p.set.Style, p.set.Tier)
	}
	for _, r := range "sgzxhjklfnc?" {
		if p.routeKey(key(r), "") {
			t.Errorf("%q went to the map; printable keys belong to the prompt", r)
		}
	}
	if p.set.Style != "roguelike" || p.set.Tier != mapmodel.TierUnicode {
		t.Errorf("letters changed the settings: %s/%s", p.set.Style, p.set.Tier)
	}
	txt := drawMapPanel(t, d, 160, 48)
	for _, want := range []string{"map style Roguelike", "map glyphs unicode", "Esc close"} {
		if !strings.Contains(txt, want) {
			t.Errorf("key bar has no %q:\n%s", want, txt)
		}
	}
	d.runForTest("map style skyline")
	d.runForTest("map glyphs ascii")
	if d.overlayMgr.ActiveName() != "map" {
		t.Fatal("a setting command closed the panel")
	}
	d.refresh()
	if p.set.Style != "skyline" || p.set.Tier != mapmodel.TierASCII {
		t.Errorf("after the commands the open panel is %s/%s", p.set.Style, p.set.Tier)
	}
	if got, glyphs, _ := eng.Account().MapPrefs(); got != "skyline" || glyphs != "ascii" {
		t.Errorf("account settings %s/%s", got, glyphs)
	}
	if txt := drawMapPanel(t, d, 160, 48); !strings.Contains(txt, "map style Skyline") {
		t.Errorf("key bar does not name the new style:\n%s", txt)
	}
}

// TestMapFlowsCommand: map flows turns the flows overlay on and off in the
// open panel, and bare it switches it.
func TestMapFlowsCommand(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	d.runForTest("map")
	for _, c := range []struct {
		line string
		want bool
	}{{"map flows on", true}, {"map flows off", false}, {"map flows", true}, {"map flows", false}} {
		d.runForTest(c.line)
		if d.mapPanel.flows != c.want {
			t.Errorf("%s: flows %v, want %v", c.line, d.mapPanel.flows, c.want)
		}
		if !strings.Contains(d.mapPanel.note, "Flows overlay") {
			t.Errorf("%s: key bar note %q", c.line, d.mapPanel.note)
		}
	}
	d.runForTest("map flows on")
	if txt := drawMapPanel(t, d, 160, 48); !strings.Contains(txt, "FLOWS") {
		t.Errorf("the roguelike shows no flows panel with map flows on:\n%s", txt)
	}
}

// TestMapInspectStagesCommand: Enter on an inspected target puts its
// command in the prompt, unrun, under the still-open panel; with something
// typed, Tab and Enter are the prompt's.
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
				if !p.routeKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), "") {
					t.Fatal("Tab on an empty prompt did not go to the map")
				}
				drawMapPanel(t, d, 160, 48)
				if in, ok := p.current().Inspect(p.frame()); ok {
					cmd = in.Command
				}
			}
			if cmd == "" {
				t.Fatal("Tab never reached a target with a command")
			}
			if txt := drawMapPanel(t, d, 160, 48); !strings.Contains(txt, "Enter types "+cmd) {
				t.Errorf("key bar does not offer %q:\n%s", cmd, txt)
			}
			p.routeKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), "")
			if d.overlayMgr.ActiveName() != "map" {
				t.Errorf("Enter closed the panel")
			}
			if got := d.inputField.GetText(); got != cmd {
				t.Errorf("prompt holds %q, want the staged %q", got, cmd)
			}
			if !d.inputField.comp.complete(cmd) {
				t.Errorf("staged %q is not a whole command", cmd)
			}
			if txt := drawMapPanel(t, d, 160, 48); !strings.Contains(txt, "Enter runs the command below") {
				t.Errorf("key bar does not say Enter runs the staged command:\n%s", txt)
			}
			for _, k := range []tcell.Key{tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEnter} {
				if p.routeKey(tcell.NewEventKey(k, 0, tcell.ModNone), cmd) {
					t.Errorf("%v went to the map with %q in the prompt", k, cmd)
				}
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

// TestCommandBarWorksWithMapOpen drives the real event loop: with the Map
// open the prompt keeps the keyboard, so typed letters land in it and Enter
// runs the line with the map still open, while the arrows and Tab (on an
// empty prompt) move the map, Enter on an empty prompt types the inspected
// command, and Esc closes the map. Adam found the command bar dead while
// the Map was open.
func TestCommandBarWorksWithMapOpen(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	// A civ met: a target for Tab at the region zoom, with a command.
	if err := eng.MeetFactionForTest("riverlands_tribes", 30); err != nil {
		t.Fatal(err)
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	sim.SetSize(160, 48)
	d.app.SetScreen(sim).SetRoot(d.pages, true).SetFocus(d.inputField)
	done := make(chan error, 1)
	go func() { done <- d.app.Run() }()
	t.Cleanup(func() {
		d.app.Stop()
		<-done
	})
	onLoop := func(f func()) {
		t.Helper()
		ok := make(chan struct{})
		go d.app.QueueUpdateDraw(func() { f(); close(ok) })
		select {
		case <-ok:
		case <-time.After(5 * time.Second):
			t.Fatal("the event loop did not answer")
		}
	}
	waitFor := func(what string, cond func() bool) {
		t.Helper()
		for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
			var ok bool
			onLoop(func() { ok = cond() })
			if ok {
				return
			}
		}
		t.Fatalf("timed out waiting for %s", what)
	}
	var prompt, active, title string
	var focused bool
	look := func() {
		prompt, active = d.inputField.GetText(), d.overlayMgr.ActiveName()
		focused = d.app.GetFocus() == d.inputField
		title = ""
		if in, ok := d.mapPanel.current().Inspect(d.mapPanel.frame()); ok {
			title = in.Title
		}
	}
	typeText := func(s string) {
		for _, r := range s {
			sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
		}
	}

	typeText("map")
	sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor("the Map panel", func() bool { look(); return active == "map" && prompt == "" && focused })
	var txt string
	onLoop(func() { txt = simText(sim) })
	if !strings.Contains(txt, "Command") || !strings.Contains(txt, "Esc close") {
		t.Errorf("the command bar or the key bar is missing with the map open:\n%s", txt)
	}

	typeText("stat")
	waitFor("letters in the prompt", func() bool { look(); return prompt == "stat" })
	for range "stat" {
		sim.InjectKey(tcell.KeyBackspace2, 0, tcell.ModNone)
	}
	waitFor("Backspace in the prompt", func() bool { look(); return prompt == "" })

	sim.InjectKey(tcell.KeyPgUp, 0, tcell.ModNone)
	waitFor("PgUp to zoom out to the region", func() bool { return strings.Contains(simText(sim), "REGION") })
	onLoop(look)
	before := title
	sim.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	waitFor("Tab to move the map's cursor", func() bool { look(); return title != before && title != "" })
	sim.InjectKey(tcell.KeyRight, 0, tcell.ModNone)
	sim.InjectKey(tcell.KeyLeft, 0, tcell.ModNone)
	onLoop(look)
	if prompt != "" || active != "map" {
		t.Errorf("the arrows reached the prompt (%q) or closed the map (%q)", prompt, active)
	}

	typeText("status")
	sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor("status to run", func() bool { look(); return prompt == "" && d.mapPanel.note != "" })
	if active != "map" {
		t.Errorf("running a command closed the map (active %q)", active)
	}

	sim.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	waitFor("the cursor on a target with a command", func() bool {
		look()
		in, ok := d.mapPanel.current().Inspect(d.mapPanel.frame())
		if !ok || in.Command == "" {
			sim.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
			return false
		}
		return true
	})
	sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	waitFor("Enter to type the inspected command", func() bool { look(); return prompt != "" })
	staged := prompt
	sim.InjectKey(tcell.KeyEsc, 0, tcell.ModNone)
	waitFor("Esc to close the map", func() bool { look(); return active == "" && focused })
	if prompt != staged {
		t.Errorf("closing the map changed the prompt from %q to %q", staged, prompt)
	}
}
