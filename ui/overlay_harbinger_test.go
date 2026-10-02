package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// tagRe matches tview colour/attribute tags, for reading panel text plainly.
var tagRe = regexp.MustCompile(`\[[a-zA-Z0-9#:\-]*\]`)

func untag(s string) string { return tagRe.ReplaceAllString(s, "") }

// harbDashboard builds a dashboard over a fresh engine whose harbinger for age
// has been summoned.
func harbDashboard(t *testing.T, age string) (*Dashboard, *game.GameEngine, *tview.Pages) {
	t.Helper()
	engine := game.NewGameEngine()
	app := tview.NewApplication()
	pages := tview.NewPages()
	d := NewDashboard(app, engine, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	if age != "" {
		if err := engine.SummonHarbingerForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	return d, engine, pages
}

// renderText draws pages at w×h into a SimulationScreen and returns the rows.
func renderText(t *testing.T, pages *tview.Pages, w, h int) string {
	t.Helper()
	var rows []string
	for _, r := range screenGrid(t, pages, w, h) {
		rows = append(rows, strings.TrimRight(string(r), " "))
	}
	return strings.Join(rows, "\n")
}

func TestHarbingerPanelWithNoHarbinger(t *testing.T) {
	engine := game.NewGameEngine() // Primitive Age: nothing can strike in the Stone Era
	txt := untag(harbingerPanelText(engine.GetState(), "", false, false))
	for _, want := range []string{"No harbinger is here", "A harbinger comes only when doom is on its way", "No catastrophe can strike in the Stone Era"} {
		if !strings.Contains(txt, want) {
			t.Errorf("panel missing %q:\n%s", want, txt)
		}
	}
	if strings.Contains(txt, "%") {
		t.Errorf("pre-industrial outlook must not print a figure:\n%s", txt)
	}
	assertNoLaterEras(t, txt, "stone_era")

	// An era that can be fated reads quiet, for now.
	iron := game.NewGameEngine()
	if err := iron.SummonHarbingerForTest("classical_age"); err != nil {
		t.Fatal(err)
	}
	if err := iron.ForceQuietFateForTest("iron_era"); err != nil {
		t.Fatal(err)
	}
	txt = untag(harbingerPanelText(iron.GetState(), "", false, false))
	if !strings.Contains(txt, "No harbinger has come: the Iron Era is quiet, for now.") {
		t.Errorf("quiet era panel:\n%s", txt)
	}
	assertNoLaterEras(t, txt, "iron_era")
}

// Until its harbinger comes, a fated doom shows nowhere: the Harbinger and
// Epoch panels, the catastrophe command, the status bar and the map read the
// same for a fated era as for a quiet one.
func TestNoScreenLeaksTheFate(t *testing.T) {
	type view struct{ panel, cmd, epoch, status, mapPrint string }
	render := func(fated bool) view {
		t.Helper()
		d, engine, _ := harbDashboard(t, "")
		engine.SeedRNG(5)
		if err := engine.SummonHarbingerForTest("classical_age"); err != nil { // places the engine
			t.Fatal(err)
		}
		var err error
		if fated {
			err = engine.ForceFateForTest("iron_era", 40000)
		} else {
			err = engine.ForceQuietFateForTest("iron_era")
		}
		if err != nil {
			t.Fatal(err)
		}
		engine.StepTicks(20)
		st := engine.GetState()
		if st.Harbinger != nil {
			t.Fatal("the harbinger came early")
		}
		d.refresh()
		m := mapmodel.NewBuilder(nil).Build(&st, nil)
		return view{
			panel:    harbingerPanelText(st, "", false, false),
			cmd:      HandleCommand("catastrophe", engine).Message,
			epoch:    epochProvider(st, 120),
			status:   d.statusTV.GetText(false),
			mapPrint: m.Fingerprint(),
		}
	}
	fated, quiet := render(true), render(false)
	if fated.panel != quiet.panel {
		t.Errorf("Harbinger panel differs:\n%s\n---\n%s", fated.panel, quiet.panel)
	}
	if fated.cmd != quiet.cmd {
		t.Errorf("catastrophe command differs:\n%s\n---\n%s", fated.cmd, quiet.cmd)
	}
	if fated.epoch != quiet.epoch {
		t.Errorf("Epoch panel differs:\n%s\n---\n%s", fated.epoch, quiet.epoch)
	}
	if fated.status != quiet.status {
		t.Errorf("status bar differs:\n%s\n---\n%s", fated.status, quiet.status)
	}
	if fated.mapPrint != quiet.mapPrint {
		t.Errorf("the map model differs:\n%s\n---\n%s", fated.mapPrint, quiet.mapPrint)
	}
}

// assertNoLaterEras fails when txt names an era after epochKey: the player
// has not reached it (the no-spoiler rule).
func assertNoLaterEras(t *testing.T, txt, epochKey string) {
	t.Helper()
	order := config.EpochByKey()[epochKey].Order
	for _, ep := range config.Epochs() {
		if ep.Order > order && strings.Contains(txt, ep.Name) {
			t.Errorf("panel names %s, an era the player has not reached:\n%s", ep.Name, txt)
		}
	}
}

func TestHarbingerPanelPreIndustrial(t *testing.T) {
	_, engine, _ := harbDashboard(t, "classical_age")
	st := engine.GetState()
	h := st.Harbinger
	if h == nil || h.Numeric || h.When != game.WhenThisAge {
		t.Fatalf("classical age harbinger = %+v", h)
	}
	txt := untag(harbingerPanelText(st, "", false, false))
	assertNoLaterEras(t, txt, "iron_era")
	for _, want := range []string{
		"The Oracle", h.Description, "Warning of impending doom before this age is out.", "Severity:", "The omens give no figure",
		"Appease: " + h.AppeaseLabel, "Brace: " + h.BraceLabel, "Invite: " + h.InviteLabel,
		"Level 0 / 2", "Next level costs:", "20% of buildings fall, 15% of stock is kept", "Next level: 15% fall, 30% kept",
		"it still comes when it was fated to",
		" A ", " B ", " I ", "Esc",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("panel missing %q:\n%s", want, txt)
		}
	}
	for _, l := range h.Lines {
		if !strings.Contains(txt, l) {
			t.Errorf("panel missing warning line %q", l)
		}
	}
	if strings.Contains(txt, "Published odds") {
		t.Error("pre-industrial panel printed the odds")
	}

	// The earliest figures give no timing; in the Stone Era, where nothing
	// can strike, a harbinger is a false prophet with nothing to answer.
	_, stone, _ := harbDashboard(t, "bronze_age")
	txt = untag(harbingerPanelText(stone.GetState(), "", false, false))
	for _, want := range []string{"Warning of impending doom. The Soothsayer gives no word of when.", "Unavailable: no catastrophe can strike in the Stone Era."} {
		if !strings.Contains(txt, want) {
			t.Errorf("stone era panel missing %q:\n%s", want, txt)
		}
	}
	assertNoLaterEras(t, txt, "stone_era")
}

func TestHarbingerPanelPostIndustrial(t *testing.T) {
	_, engine, _ := harbDashboard(t, "industrial_age")
	st := engine.GetState()
	if st.Harbinger == nil || !st.Harbinger.Numeric {
		t.Fatalf("industrial age harbinger = %+v", st.Harbinger)
	}
	txt := untag(harbingerPanelText(st, "", false, false))
	want := "Published odds: " + harbingerPercent(st.Harbinger.Probability)
	if !strings.Contains(txt, want) {
		t.Errorf("panel missing %q:\n%s", want, txt)
	}
}

// The panel's keys run the actions against the engine, report back in the
// panel, and Invite needs a second press.
func TestHarbingerPanelKeys(t *testing.T) {
	d, engine, pages := harbDashboard(t, "iron_age")
	d.harbPanel.reset()
	if !d.overlayMgr.Show("harbinger", engine.GetState()) {
		t.Fatal("harbinger overlay not registered")
	}
	capture := d.overlayMgr.entries["harbinger"].tv.GetInputCapture()
	press := func(r rune) *tcell.EventKey {
		return capture(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}

	// Brace with empty stores: refused, reported, nothing changes.
	if press('b') != nil {
		t.Error("B was not consumed")
	}
	if !strings.Contains(d.harbPanel.note, "Cannot afford to brace") || d.harbPanel.noteGood {
		t.Errorf("brace note = %q", d.harbPanel.note)
	}
	if engine.GetState().Harbinger.BraceLevel != 0 {
		t.Error("refused brace changed the level")
	}

	// Appease: the Iron Era price is beyond a new settlement's storage, and
	// the panel says what storage it needs.
	need := engine.GetState().Harbinger.AppeaseCost["faith"]
	press('A')
	if st := engine.GetState(); st.Harbinger.AppeaseLevel != 0 || d.harbPanel.noteGood {
		t.Fatalf("appease: level %d note %q", st.Harbinger.AppeaseLevel, d.harbPanel.note)
	}
	if want := "faith storage must reach " + textfmt.Number(need); !strings.Contains(untag(renderText(t, pages, 160, 60)), want) {
		t.Errorf("panel did not show the appease refusal %q", want)
	}

	// Invite: first press only arms it.
	press('i')
	if engine.GetState().Harbinger.Invited || !d.harbPanel.inviteArmed {
		t.Fatal("a single I press must not invite")
	}
	// Any other action disarms it.
	press('b')
	if d.harbPanel.inviteArmed {
		t.Error("another key should disarm the invite")
	}
	press('i')
	press('I')
	if st := engine.GetState(); !st.Harbinger.Invited || st.Harbinger.AppeaseBlocked == "" {
		t.Fatalf("invite after two presses: %+v", st.Harbinger)
	}
	press('a')
	if !strings.Contains(d.harbPanel.note, "invited") {
		t.Errorf("appease after invite note = %q", d.harbPanel.note)
	}

	// Other keys pass through; Esc closes.
	if ev := press('x'); ev == nil {
		t.Error("an unrelated key was swallowed")
	}
	capture(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
	if d.overlayMgr.HasActive() {
		t.Error("Esc did not close the panel")
	}
}

func TestHarbingerStatusBadgeAndSidebar(t *testing.T) {
	d, _, _ := harbDashboard(t, "")
	d.refresh()
	if strings.Contains(d.statusTV.GetText(true), "⚑ Harbinger") {
		t.Error("badge shown with no harbinger")
	}
	d2, _, _ := harbDashboard(t, "medieval_age")
	d2.refresh()
	if !strings.Contains(d2.statusTV.GetText(true), "⚑ Harbinger.") {
		t.Errorf("no harbinger badge in status bar: %q", d2.statusTV.GetText(true))
	}
	if !strings.Contains(buildSidebarText(""), "harbinger") {
		t.Error("harbinger missing from the sidebar Panels list")
	}
}

func TestHarbingerCommand(t *testing.T) {
	engine := game.NewGameEngine()
	for _, cmd := range []string{"harbinger", "harb"} {
		if r := HandleCommand(cmd, engine); r.OverlayName != "harbinger" {
			t.Errorf("%q → %+v, want the harbinger overlay", cmd, r)
		}
	}
	if r := HandleCommand("harbinger appease", engine); r.Type != "error" || !strings.Contains(r.Message, "No harbinger") {
		t.Errorf("appease with no harbinger → %+v", r)
	}
	if r := HandleCommand("harbinger dance", engine); r.Type != "info" || !strings.Contains(r.Message, "Usage") {
		t.Errorf("bad subcommand → %+v", r)
	}
	if err := engine.SummonHarbingerForTest("medieval_age"); err != nil {
		t.Fatal(err)
	}
	if r := HandleCommand("harb invite", engine); r.Type != "success" || !engine.GetState().Harbinger.Invited {
		t.Errorf("harb invite → %+v", r)
	}
}

// Pre-industrial, no screen prints the real odds while a harbinger is present:
// the catastrophe command and the Epoch overlay repeat the harbinger's word.
func TestOutlookFollowsHarbingerPrecision(t *testing.T) {
	_, engine, _ := harbDashboard(t, "bronze_age")
	st := engine.GetState()
	msg := untag(catastropheOutlookText(st))
	if strings.Contains(msg, "% catastrophe chance") || !strings.Contains(msg, "The Soothsayer warns of doom, with no word of when: "+string(st.Harbinger.Tier)) {
		t.Errorf("catastrophe outlook with a pre-industrial harbinger = %q", msg)
	}
	ep := untag(epochProvider(st, 120))
	if strings.Contains(ep, "% catastrophe chance") || !strings.Contains(ep, "The Soothsayer is here") {
		t.Errorf("epoch overlay with a pre-industrial harbinger:\n%s", ep)
	}

	_, engine2, _ := harbDashboard(t, "industrial_age")
	msg = untag(catastropheOutlookText(engine2.GetState()))
	if !strings.Contains(msg, "% catastrophe chance") {
		t.Errorf("industrial outlook should print the odds: %q", msg)
	}
}

func TestEpochOverlayHarbingerHistory(t *testing.T) {
	st := game.NewGameEngine().GetState()
	st.HarbingerHistory = []game.HarbingerRecord{
		{Age: "bronze_age", Name: "the Soothsayer", Chain: []string{"primitive_age", "stone_age", "bronze_age"},
			TargetEpochKey: "iron_era", TargetEpochName: "Iron Era",
			Outcome: game.HarbingerOutcomeDiscredited, FalseProphet: true},
		{Age: "medieval_age", Name: "the Town Crier", TargetEpochKey: "steel_era", TargetEpochName: "Steel Era",
			Outcome: game.HarbingerOutcomeVindicated, AppeaseLevel: 2, BraceLevel: 1},
	}
	st.HarbingerHistory = append(st.HarbingerHistory, game.HarbingerRecord{Age: "classical_age", Name: "the Oracle",
		EpochKey: "iron_era", TargetEpochKey: "iron_era", TargetEpochName: "Iron Era", Outcome: game.HarbingerOutcomeSpared, AtAdvance: true})
	txt := untag(epochProvider(st, 120))
	for _, want := range []string{
		"── Harbingers ──",
		"The Wild Man, The Hermit, The Soothsayer → Iron Era: Discredited",
		"The Town Crier → Steel Era: Vindicated (appeased 2, braced 1)",
		"The Oracle → doom in the Iron Era: Spared (settled as you advanced)",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("epoch overlay missing %q:\n%s", want, txt)
		}
	}
}

// A handoff toasts the new figure, and the panel names who spoke before.
func TestHarbingerHandoffToastAndEarlierFigures(t *testing.T) {
	d, engine, _ := harbDashboard(t, "bronze_age")
	engine.Bus.Publish(game.EventData{Type: game.EventHarbingerArrived, Payload: map[string]interface{}{
		"harbinger_name": "the Soothsayer", "handoff": true,
	}})
	found := false
	for i := 0; i < 3 && !found; i++ { // an arrival toast may be queued ahead of it
		found = strings.Contains(d.toastMgr.GetCurrent(), "The Soothsayer takes up the warning")
		if !found {
			d.toastMgr.mu.Lock()
			d.toastMgr.current = nil
			if len(d.toastMgr.queue) > 0 {
				next := d.toastMgr.queue[0]
				d.toastMgr.current, d.toastMgr.queue = &next, d.toastMgr.queue[1:]
			}
			d.toastMgr.mu.Unlock()
		}
	}
	if !found {
		t.Error("no handoff toast")
	}
	st := engine.GetState()
	st.Harbinger.Earlier = []string{"the Wild Man", "the Hermit"}
	txt := untag(harbingerPanelText(st, "", false, false))
	if !strings.Contains(txt, "Took up the warning from The Wild Man, then The Hermit. Your answers stand.") {
		t.Errorf("panel does not name the earlier figures:\n%s", txt)
	}
}

// In the Cosmic Era a fated doom (the Reality Tear) can speak while the Last
// Passage's thread waits: the panel says so, and the catastrophe command and
// the Epoch panel show both, each at its own odds.
func TestCosmicDoomScreens(t *testing.T) {
	_, engine, _ := harbDashboard(t, "galactic_age") // the Last Passage thread, over a quiet fate
	if h := engine.GetState().Harbinger; h == nil || !h.LastPassage {
		t.Fatalf("setup: harbinger %+v", h)
	}
	if err := engine.ForceFateForTest("cosmic_era", 2000); err != nil {
		t.Fatal(err)
	}
	engine.StepTicks(1)
	st := engine.GetState()
	h := st.Harbinger
	if h == nil || h.LastPassage || !h.LastPassageWaiting {
		t.Fatalf("doom's harbinger = %+v", h)
	}
	panel := untag(harbingerPanelText(st, "", false, false))
	if !strings.Contains(panel, "The Last Passage still waits at your next prestige.") {
		t.Errorf("panel:\n%s", panel)
	}
	doomOdds := harbingerPercent(h.Probability)
	passOdds := harbingerPercent(st.CatastropheOutlook.Probability)
	if doomOdds == passOdds {
		t.Fatalf("setup: the doom and the Last Passage share odds %s", doomOdds)
	}
	for name, txt := range map[string]string{
		"catastrophe": untag(HandleCommand("catastrophe", engine).Message),
		"epoch":       untag(epochProvider(st, 140)),
	} {
		if !strings.Contains(txt, "warns of doom") || !strings.Contains(txt, "Next passage (prestige, the Last Passage)") {
			t.Errorf("%s should show the doom and the Last Passage:\n%s", name, txt)
		}
		for _, line := range strings.Split(txt, "\n") {
			if strings.Contains(line, "Last Passage") && strings.Contains(line, doomOdds) {
				t.Errorf("%s: the Last Passage line carries the doom's %s: %q", name, doomOdds, line)
			}
		}
	}
}
