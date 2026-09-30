package ui

import (
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// logRoutingDashboard builds a dashboard over newCaseTestEngine (the
// Primitive Age, stocked so every command below succeeds).
func logRoutingDashboard(t *testing.T) (*Dashboard, *game.GameEngine) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := newCaseTestEngine(t, "")
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	return d, eng
}

// submitForTest types line at the prompt and presses Enter.
func submitForTest(d *Dashboard, line string) {
	d.inputField.SetText(line)
	d.submitInput()
}

// TestRoutineLinesSkipTheMainLog: routine confirmations (a build started,
// workers recruited and assigned, a gather, a plan item) reach the logs
// panel and never the main window's log, while errors and notable events
// still reach the main log (playtest 2026-09-29: "Started building Hut"
// belongs in the logs panel only).
func TestRoutineLinesSkipTheMainLog(t *testing.T) {
	d, eng := logRoutingDashboard(t)
	eng.Resources.LoadAmounts(map[string]float64{"food": 100}) // room to gather into
	routine := map[string]string{
		"gather food":             "Gathered ",
		"build hut":               "Started building Hut",
		"recruit":                 "Recruited ",
		"assign gathering_camp":   "Assigned ",
		"unassign gathering_camp": "Unassigned ",
		"plan build hut":          "Planned ",
		"plan clear":              "Cleared the plan",
	}
	for _, cmd := range []string{"gather food", "build hut", "recruit", "assign gathering_camp", "unassign gathering_camp", "plan build hut", "plan clear"} {
		submitForTest(d, cmd)
	}
	submitForTest(d, "build no_such_building")

	st := eng.GetState()
	d.refreshLog(st)
	main := d.logTV.GetText(true)
	panel := guardRendered(logsProvider(st, guardPanelWidth))
	for cmd, line := range routine {
		if strings.Contains(main, line) {
			t.Errorf("%q: the main log shows the routine line %q:\n%s", cmd, line, main)
		}
		if !strings.Contains(panel, line) {
			t.Errorf("%q: the logs panel lacks %q:\n%s", cmd, line, panel)
		}
	}
	for _, notable := range []string{"Unknown building", "Welcome to AgeForge"} {
		if !strings.Contains(main, notable) {
			t.Errorf("the main log lacks the notable line %q:\n%s", notable, main)
		}
	}
}

// TestLogRoutingRule pins which category reaches which log.
func TestLogRoutingRule(t *testing.T) {
	for _, c := range []struct {
		typ         string
		main, panel bool
	}{
		{"info", true, true},
		{"success", true, true},
		{"warning", true, true},
		{"error", true, true},
		{"event", true, true},
		{game.LogRoutine, false, true},
		{"debug", false, false},
	} {
		e := game.LogEntry{Type: c.typ, Message: "x"}
		if got := mainLogShows(e); got != c.main {
			t.Errorf("%s: main log shows it = %v, want %v", c.typ, got, c.main)
		}
		if got := logsPanelShows(e); got != c.panel {
			t.Errorf("%s: logs panel shows it = %v, want %v", c.typ, got, c.panel)
		}
	}
}
