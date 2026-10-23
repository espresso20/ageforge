package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEveryTextOfTheScreensSheetIsWiredOrLeftAlone: a text of the screens and
// the help that can follow VibeMode does: it is a vibe call (Line, Sprintf,
// Errorf) or an entry of a table (E) that carries its row id. A new label that
// is not fails this test, and the fix is one command:
//
//	go run ./cmd/words vibe-wire screens-and-help.csv
//
// What is left alone has a reason that is final (a key, a command, another
// program's menus); none is left for later ("made once", "compared").
func TestEveryTextOfTheScreensSheetIsWiredOrLeftAlone(t *testing.T) {
	c := realCatalog(t)
	var loose, later []string
	wired := 0
	left := map[string]int{}
	for _, r := range c.rows {
		if r.area != areaScreens {
			continue
		}
		where := fmt.Sprintf("%s (%s:%d)", r.id, r.u.f.rel, r.u.line)
		switch v := c.wireVerdict(r); {
		case r.u.vibeID != "":
			wired++
		case v.kind != wireNone:
			loose = append(loose, where)
		case !forGood(v.reason):
			later = append(later, where+": "+v.reason)
		default:
			left[v.reason]++
		}
	}
	if wired < 1900 {
		t.Errorf("only %d texts of the screens sheet are wired", wired)
	}
	if len(loose) > 0 {
		t.Errorf("%d texts of the screens sheet can follow VibeMode and are not written as vibe calls; run `go run ./cmd/words vibe-wire screens-and-help.csv`:\n  %s",
			len(loose), strings.Join(loose, "\n  "))
	}
	if len(later) > 0 {
		t.Errorf("%d texts of the screens sheet cannot follow VibeMode as the code is written; change the code or judge them in vibe_left.go:\n  %s",
			len(later), strings.Join(later, "\n  "))
	}
	for _, reason := range []string{leftKeycap, leftCommand, leftOtherApp, leftDev, leftBoot} {
		if left[reason] == 0 {
			t.Errorf("no text of the screens sheet is left alone as %q: the rule or the list is dead", reason)
		}
	}
}

// TestKeycapsAndMarksAreNotWords: the name of a key, and a text with nothing in
// it a script could write, are judged by their words alone.
func TestKeycapsAndMarksAreNotWords(t *testing.T) {
	for text, want := range map[string]string{
		"Esc": leftKeycap, "Enter": leftKeycap, "PgUp/PgDn": leftKeycap, "Home End": leftKeycap,
		"Ctrl+R": leftKeycap, "Up/Down": leftKeycap, "Arrows": leftKeycap, "Shift-Tab": leftKeycap,
		"[gold]Tab[-]": leftKeycap,
		"%d%%":         leftNoWords, " · ": leftNoWords, "[red]▼[-]": leftNoWords, "%s\n": leftNoWords,
		"Esc closes": "", "Close": "", "Take the completion": "", "Left": leftKeycap, "Tab moves between buttons": "",
	} {
		r := &row{area: areaScreens, current: text, id: "ui.x.y.01"}
		if got := screenReason(r); got != want {
			t.Errorf("%q: left alone as %q, want %q", text, got, want)
		}
	}
	// The rules are for the screens only.
	if got := screenReason(&row{area: areaLog, current: "Esc", id: "game.x.y.01"}); got != "" {
		t.Errorf("a text of the log is left alone as %q", got)
	}
}

// TestFixedWidthLabelsCarryTheirWidth: a label the code lays out in a fixed
// width has that width as its max in the export, whichever way the code gives
// it: padRight, a field of a format, a constant of the table, or its own
// padding.
func TestFixedWidthLabelsCarryTheirWidth(t *testing.T) {
	c := realCatalog(t)
	for id, want := range map[string]int{
		"ui.overlay_help.helpProvider.29":           12, // padRight(label, 12)
		"ui.expedition_view.briefing.01":            7,  // padRight(label, 7)
		"ui.expedition_view.drawConfirm.06":         6,  // a column of labels
		"ui.icons.view.03":                          19, // the pair of lines in the icons check
		"ui.overlay_stats.statsProvider.21":         20, // Fprintf(" %-20s ", label)
		"ui.overlay_stats.statsProvider.30":         16, // %-16s
		"ui.panels_table.panelGroups.02":            18, // panelJobMax
		"ui.panels_table.panelGroups.22":            18,
		"ui.advance_view.advMarch.01":               0, // a heading drawn in a box of any width
		"ui.overlay_help.helpProvider.28":           0,
		"ui.dashboard.NewDashboard.01":              0,
		"ui.input.cmdWonder.08":                     0,
		"ui.menu.buildViewWith.01":                  0,
		"ui.expedition_view.drawConfirm.08":         0, // " Back in  " between two values, not a column
		"ui.overlay_stats.statsProvider.10":         0, // the %-14s there is an age's name, not this text
		"ui.people_tabs.overview.04":                0,
		"ui.overlay_factions.dealLine.04":           0,
		"ui.expedition_view.drawReady.03":           0,
		"ui.overlay_harbinger.harbingerCostLine.04": 0,
	} {
		r := c.byID[id]
		if r == nil {
			t.Errorf("%s: no such row", id)
			continue
		}
		if r.max != want {
			t.Errorf("%s (%q): max %d, want %d", id, r.current, r.max, want)
		}
	}
}

// TestVibeLintHoldsEdgeSpaces: the spaces and line breaks a text starts and
// ends with are how the code joins it to the text next to it, so a line that
// drops them is an error.
func TestVibeLintHoldsEdgeSpaces(t *testing.T) {
	c := realCatalog(t)
	var spaced *row
	for _, r := range c.rows {
		if r.area == areaScreens && r.u.vibeID != "" && strings.HasPrefix(r.current, " ") && strings.HasSuffix(r.current, " ") && len(r.kept()) == 0 {
			spaced = r
			break
		}
	}
	if spaced == nil {
		t.Fatal("no wired text of the screens starts and ends with a space")
	}
	dir := t.TempDir()
	body := "id,text\n" + spaced.id + ",\"" + strings.TrimSpace(spaced.current) + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "1.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if n, err := vibeLint(c, realRoot(t), dir, &out); err != nil || n != 1 || !strings.Contains(out.String(), "as the original does") {
		t.Errorf("vibe-lint on a line that lost its edge spaces: %d errors, %v:\n%s", n, err, out.String())
	}
	body = "id,text\n" + spaced.id + ",\"" + strings.Repeat(" ", len(spaced.current)-len(strings.TrimLeft(spaced.current, " "))) + "zzz" + strings.Repeat(" ", len(spaced.current)-len(strings.TrimRight(spaced.current, " "))) + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "1.csv"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if n, err := vibeLint(c, realRoot(t), dir, &out); err != nil || n != 0 {
		t.Errorf("vibe-lint refused a line that keeps its edge spaces: %d errors, %v:\n%s", n, err, out.String())
	}
}
