package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
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
	engine := game.NewGameEngine() // Primitive Age: the Iron Era passage can roll
	txt := untag(harbingerPanelText(engine.GetState(), "", false, false))
	for _, want := range []string{"No harbinger is here", "Iron Era could bring a catastrophe", "The risk is"} {
		if !strings.Contains(txt, want) {
			t.Errorf("panel missing %q:\n%s", want, txt)
		}
	}
	if strings.Contains(txt, "%") {
		t.Errorf("pre-industrial outlook must not print a figure:\n%s", txt)
	}
}

func TestHarbingerPanelPreIndustrial(t *testing.T) {
	_, engine, _ := harbDashboard(t, "bronze_age")
	st := engine.GetState()
	h := st.Harbinger
	if h == nil || h.Numeric {
		t.Fatalf("bronze age harbinger = %+v", h)
	}
	txt := untag(harbingerPanelText(st, "", false, false))
	for _, want := range []string{
		"THE SOOTHSAYER", h.Description, "Iron Era", "Severity:", "The omens give no figure",
		"Appease — " + h.AppeaseLabel, "Brace — " + h.BraceLabel, "Invite — " + h.InviteLabel,
		"Level 0 / 2", "Next level costs:", "20% of buildings fall, 15% of stock is kept", "Next level: 15% fall, 30% kept",
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
	if strings.Contains(txt, "Odds published") {
		t.Error("pre-industrial panel printed the odds")
	}
}

func TestHarbingerPanelPostIndustrial(t *testing.T) {
	_, engine, _ := harbDashboard(t, "industrial_age")
	st := engine.GetState()
	if st.Harbinger == nil || !st.Harbinger.Numeric {
		t.Fatalf("industrial age harbinger = %+v", st.Harbinger)
	}
	txt := untag(harbingerPanelText(st, "", false, false))
	want := "Odds published: " + harbingerPercent(st.Harbinger.Probability)
	if !strings.Contains(txt, want) {
		t.Errorf("panel missing %q:\n%s", want, txt)
	}
}

// The panel's keys run the actions against the engine, report back in the
// panel, and Invite needs a second press.
func TestHarbingerPanelKeys(t *testing.T) {
	d, engine, pages := harbDashboard(t, "bronze_age")
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
	if !strings.Contains(d.harbPanel.note, "cannot afford to brace") || d.harbPanel.noteGood {
		t.Errorf("brace note = %q", d.harbPanel.note)
	}
	if engine.GetState().Harbinger.BraceLevel != 0 {
		t.Error("refused brace changed the level")
	}

	// Appease: the passage price (12000 faith in the Stone Era) is beyond a
	// new settlement's storage, and the panel says what storage it needs.
	press('A')
	if st := engine.GetState(); st.Harbinger.AppeaseLevel != 0 || d.harbPanel.noteGood {
		t.Fatalf("appease: level %d note %q", st.Harbinger.AppeaseLevel, d.harbPanel.note)
	}
	if !strings.Contains(untag(renderText(t, pages, 160, 60)), "faith storage must reach 12000") {
		t.Error("panel did not show the appease refusal")
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
	if strings.Contains(d.statusTV.GetText(true), "HARBINGER") {
		t.Error("badge shown with no harbinger")
	}
	d2, _, _ := harbDashboard(t, "medieval_age")
	d2.refresh()
	if !strings.Contains(d2.statusTV.GetText(true), "⚑ HARBINGER") {
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
	if r := HandleCommand("harbinger appease", engine); r.Type != "error" || !strings.Contains(r.Message, "no harbinger") {
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
	if strings.Contains(msg, "% catastrophe chance") || !strings.Contains(msg, "The Soothsayer warns of "+string(st.Harbinger.Tier)) {
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
	txt := untag(epochProvider(st, 120))
	for _, want := range []string{
		"── Harbingers ──",
		"The Wild Man, The Hermit, The Soothsayer → Iron Era: Discredited",
		"The Town Crier → Steel Era: Vindicated (appeased 2, braced 1)",
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
