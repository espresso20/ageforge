package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// logRoutingDashboard builds a dashboard over newCaseTestEngine (the
// Primitive Age, stocked so every command below succeeds).
func logRoutingDashboard(t *testing.T) (*Dashboard, *game.GameEngine, *tview.Pages) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := newCaseTestEngine(t, "")
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	return d, eng, pages
}

// submitForTest types line at the prompt and presses Enter.
func submitForTest(d *Dashboard, line string) {
	d.inputField.SetText(line)
	d.submitInput()
}

// tickPrefixRe matches a log line that starts with a tick number ("T133 ").
var tickPrefixRe = regexp.MustCompile(`^\s*T\d+\s`)

// logLineWith returns the first line of text that contains part, or "".
func logLineWith(text, part string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, part) {
			return line
		}
	}
	return ""
}

// TestCommandRepliesReachTheMainLog: the main window's log says what each
// command did, routine confirmations included (a gather, a build started or
// queued, workers recruited and assigned, a plan item), with no tick number
// in front. The logs panel shows the same lines with their tick numbers and
// category tags. Errors and notable events reach both logs in their colors.
// Playtest 2026-09-29 asked for the tick number ("T133 Started building
// Hut") to leave the main window; a first fix took the whole line out.
func TestCommandRepliesReachTheMainLog(t *testing.T) {
	d, eng, _ := logRoutingDashboard(t)
	eng.Resources.LoadAmounts(map[string]float64{"food": 100}) // room to gather into
	replies := []struct{ cmd, line string }{
		{"gather food", "Gathered "},
		{"build hut", "Started building Hut"},
		{"build hut 2", "Queued 2 Huts"},
		{"recruit", "Recruited "},
		{"assign gathering_camp", "Assigned "},
		{"unassign gathering_camp", "Unassigned "},
		{"plan build hut", "Planned "},
		{"plan clear", "Cleared the plan"},
		{"build no_such_building", "Unknown building"},
	}
	for _, r := range replies {
		submitForTest(d, r.cmd)
	}
	const ready = "✦ Ready to advance to the Stone Age."
	eng.AddLog("event", ready+" Type 'advance' when you're ready.")

	st := eng.GetState()
	d.refreshLog(st)
	main := d.logTV.GetText(true)
	panel := guardRendered(logsProvider(st, guardPanelWidth))
	for _, r := range append(replies, struct{ cmd, line string }{"(an event)", ready}) {
		if got := logLineWith(main, r.line); got == "" {
			t.Errorf("%s: the main log lacks %q:\n%s", r.cmd, r.line, main)
		} else if tickPrefixRe.MatchString(got) {
			t.Errorf("%s: the main log line starts with a tick number: %q", r.cmd, got)
		}
		if got := logLineWith(panel, r.line); got == "" {
			t.Errorf("%s: the logs panel lacks %q:\n%s", r.cmd, r.line, panel)
		} else if !tickPrefixRe.MatchString(got) {
			t.Errorf("%s: the logs panel line has no tick number: %q", r.cmd, got)
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(main), "\n") {
		if tickPrefixRe.MatchString(line) {
			t.Errorf("a main log line starts with a tick number: %q", line)
		}
	}

	// The main log keeps its look: each category's color, the message's own
	// marker (✦), no category tag. The logs panel keeps its tags.
	raw := d.logTV.GetText(false)
	for _, want := range []string{"[red]Unknown building", "[gold]" + ready, "[white]Gathered "} {
		if !strings.Contains(raw, want) {
			t.Errorf("the main log lacks %q:\n%s", want, raw)
		}
	}
	for _, c := range []struct{ line, tag string }{{"Unknown building", "[X]"}, {ready, "[*]"}, {"Gathered ", "·"}} {
		if got := logLineWith(main, c.line); strings.Contains(got, c.tag) {
			t.Errorf("the main log line carries the logs panel's tag %q: %q", c.tag, got)
		}
		got := logLineWith(panel, c.line)
		if i := strings.Index(got, c.line); i < 0 || !strings.Contains(got[:i], c.tag) {
			t.Errorf("the logs panel line lacks the tag %q before the message: %q", c.tag, got)
		}
	}
}

// TestMainLogOnScreen draws the dashboard and reads its Log box: the
// command's line is on screen with no tick number in front of it.
func TestMainLogOnScreen(t *testing.T) {
	d, eng, pages := logRoutingDashboard(t)
	eng.Resources.LoadAmounts(map[string]float64{"food": 100})
	submitForTest(d, "gather food")
	submitForTest(d, "build hut")
	d.refresh()
	screen := drawDashboard(t, d, pages, 160, 48)
	for _, line := range []string{"Gathered ", "Started building Hut"} {
		row := logLineWith(screen, line)
		if row == "" {
			t.Errorf("the main window lacks %q:\n%s", line, screen)
			continue
		}
		if regexp.MustCompile(`T\d+\s+` + regexp.QuoteMeta(line)).MatchString(row) {
			t.Errorf("the main window shows a tick number before %q: %q", line, row)
		}
	}
}

// TestMapSettingRepliesInTheMainLog: the map settings' confirmations
// (routine lines once an account keeps the setting) read as plain sentences
// in the main window's log.
func TestMapSettingRepliesInTheMainLog(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	cases := []struct{ cmd, line string }{
		{"minimap off", "Mini map off. Type minimap on to bring it back."},
		{"minimap on", "Mini map on. It shows above the Buildings list"},
		{"map style skyline", "Map style set to Skyline."},
		{"map glyphs ascii", "Map glyphs set to ascii."},
		{"map flows on", "Flows overlay on: full stores"},
	}
	for _, c := range cases {
		d.runForTest(c.cmd)
		logs := eng.GetLogs()
		if last := logs[len(logs)-1]; last.Type != game.LogRoutine {
			t.Errorf("%s: logged %q as %q, want %q (the routine path goes untested)", c.cmd, last.Message, last.Type, game.LogRoutine)
		}
	}
	d.refreshLog(eng.GetState())
	main := d.logTV.GetText(true)
	for _, c := range cases {
		got := logLineWith(main, c.line)
		if got == "" {
			t.Errorf("%s: the main log lacks %q:\n%s", c.cmd, c.line, main)
			continue
		}
		if !strings.HasPrefix(got, c.line) {
			t.Errorf("%s: the main log line does not start with the reply: %q", c.cmd, got)
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
		{game.LogRoutine, true, true},
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
