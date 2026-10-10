package ui

import (
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// The playtest's screen and town: a MacBook at a large font, where the
// middle column is about 60 columns by 45 rows, and a Primitive Age town of
// 160 people with 32 huts, 20 stashes and 15 gathering camps. The mini map
// took the top third of that column and the Buildings list was cut off
// mid-entry at the bottom ("Stash x20" showed its name and nothing else).
const (
	playtestCols = 2*60 + sidebarW
	playtestRows = 45 + dashHeaderRows + promptRows
)

// playtestTown is a dashboard on a town like the playtest's.
func playtestTown(t *testing.T) (*Dashboard, *tview.Pages) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()
	eng.SetStockForTest("food", 10000)
	eng.SetStockForTest("wood", 10000)
	eng.SetTownForTest(map[string]int{
		"hut": 32, "stash": 20, "gathering_camp": 15, "wood_camp": 12, "story_circle": 6, "shrine": 4, "sacred_grove": 2,
	}, 160)
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.refresh()
	return d, pages
}

// buildingsBox returns the rows inside the Buildings box as drawn, with the
// trailing blanks trimmed off each.
func buildingsBox(t *testing.T, d *Dashboard, pages *tview.Pages, w, h int) []string {
	t.Helper()
	drawDashboard(t, d, pages, w, h)
	screen := strings.Split(drawDashboard(t, d, pages, w, h), "\n")
	x, y, bw, bh := d.economyTab.buildingTV.GetInnerRect()
	var rows []string
	for j := y; j < y+bh && j < len(screen); j++ {
		line := []rune(screen[j])
		if x >= len(line) {
			rows = append(rows, "")
			continue
		}
		rows = append(rows, strings.TrimRight(string(line[x:min(x+bw, len(line))]), " "))
	}
	return rows
}

// wholeEntries fails if the box shows part of an entry: every entry of the
// list is on screen whole, line for line, or not at all. It returns how
// many are on screen.
func wholeEntries(t *testing.T, where string, d *Dashboard, rows []string) int {
	t.Helper()
	_, _, w, _ := d.economyTab.buildingTV.GetInnerRect()
	_, entries := buildingEntries(d.economyTab.state, w)
	screen := strings.Join(rows, "\n") + "\n"
	shown := 0
	for _, e := range entries {
		var plain []string
		for _, line := range e {
			plain = append(plain, strings.TrimRight(untag(line), " "))
		}
		// The entry's own first line: a building's name, or the first-steps
		// guide's heading (which opens with a blank line).
		name := ""
		for _, line := range plain {
			if name = line; strings.TrimSpace(line) != "" {
				break
			}
		}
		if !strings.Contains(screen, name+"\n") {
			continue // not on screen at all
		}
		shown++
		if !strings.Contains(screen, strings.Join(plain, "\n")+"\n") {
			t.Errorf("%s: the Buildings box shows part of an entry, %q, and not the rest of it:\n%s", where, strings.TrimSpace(name), strings.Join(rows, "\n"))
		}
	}
	if shown == 0 {
		t.Errorf("%s: the Buildings box shows no entry:\n%s", where, strings.Join(rows, "\n"))
	}
	return shown
}

// TestBuildingsBoxOnThePlaytestScreen: on the playtest's screen and town
// the dashboard has no mini map (it is off until asked for) and the
// Buildings list has the whole column. With more entries than fit, the box
// shows whole entries only, says how many are below and names the key; a
// page down shows the rest, with how many are above.
func TestBuildingsBoxOnThePlaytestScreen(t *testing.T) {
	d, pages := playtestTown(t)
	txt := drawDashboard(t, d, pages, playtestCols, playtestRows)
	if d.mapDock.shown || strings.Contains(txt, " Map · ") {
		t.Fatalf("the mini map shows on a dashboard nobody turned it on for:\n%s", txt)
	}
	rows := buildingsBox(t, d, pages, playtestCols, playtestRows)
	_, _, bw, bh := d.economyTab.buildingTV.GetInnerRect()
	if bw < 55 || bw > 65 || bh < 40 {
		t.Fatalf("the Buildings box is %d by %d inside; the playtest's column was about 60 by 45", bw, bh)
	}
	_, entries := buildingEntries(d.economyTab.state, bw)
	first := wholeEntries(t, "the first page", d, rows)
	if first >= len(entries) {
		t.Fatalf("all %d entries fit a %d-row box: the test needs a list longer than its box", len(entries), bh)
	}
	last := rows[len(rows)-1]
	for len(rows) > 0 && last == "" {
		rows = rows[:len(rows)-1]
		last = rows[len(rows)-1]
	}
	if want := "↓ " + itoa(len(entries)-first) + " more below: PgDn"; strings.TrimSpace(last) != want {
		t.Errorf("the box does not end by saying what is below it: %q, want %q", last, want)
	}
	if strings.Contains(strings.Join(rows, "\n"), "more above") {
		t.Errorf("the first page says there is more above it:\n%s", strings.Join(rows, "\n"))
	}
	// The stash, which the playtest saw cut down to its name, is whole or
	// absent on each page.
	seen := map[string]bool{}
	note := func(rows []string) {
		for _, r := range rows {
			if name := strings.TrimSpace(r); strings.HasPrefix(name, "✗ ") || strings.HasPrefix(name, "✓ ") || strings.HasPrefix(name, "MAX ") {
				seen[name] = true
			}
		}
	}
	note(rows)

	// Page down until the end: every page is whole entries, and names
	// what is above it.
	for page := 2; page < 10 && d.economyTab.bldFirst+d.economyTab.bldShown < len(entries); page++ {
		d.economyTab.ScrollDown()
		rows = buildingsBox(t, d, pages, playtestCols, playtestRows)
		wholeEntries(t, "page "+itoa(page), d, rows)
		if want := "↑ " + itoa(d.economyTab.bldFirst) + " more above: PgUp"; !strings.Contains(strings.Join(rows, "\n"), want) {
			t.Errorf("page %d does not say what is above it (%q):\n%s", page, want, strings.Join(rows, "\n"))
		}
		note(rows)
	}
	if len(seen) < len(entries)-1 { // the first-steps guide is an entry with no name line
		t.Errorf("paging through the list showed %d buildings of %d entries: %v", len(seen), len(entries), seen)
	}
	// And back up to the top.
	for i := 0; i < 10 && d.economyTab.bldFirst > 0; i++ {
		d.economyTab.ScrollUp()
		wholeEntries(t, "paging back", d, buildingsBox(t, d, pages, playtestCols, playtestRows))
	}
	if d.economyTab.bldFirst != 0 {
		t.Errorf("PgUp did not come back to the top: the list starts at entry %d", d.economyTab.bldFirst)
	}
}

// TestBuildingsBoxUnderTheMiniMap: with the mini map turned on at the
// playtest's screen, the list has two thirds of the column and still shows
// whole entries only, and says what is below.
func TestBuildingsBoxUnderTheMiniMap(t *testing.T) {
	d, pages := playtestTown(t)
	miniOn(t, d)
	txt := drawDashboard(t, d, pages, playtestCols, playtestRows)
	if !d.mapDock.shown || !strings.Contains(txt, " Map · ") {
		t.Fatalf("minimap on did not show the mini map at %dx%d", playtestCols, playtestRows)
	}
	rows := buildingsBox(t, d, pages, playtestCols, playtestRows)
	shown := wholeEntries(t, "under the mini map", d, rows)
	_, _, bw, _ := d.economyTab.buildingTV.GetInnerRect()
	_, entries := buildingEntries(d.economyTab.state, bw)
	if want := "↓ " + itoa(len(entries)-shown) + " more below: PgDn"; !strings.Contains(strings.Join(rows, "\n"), want) {
		t.Errorf("under the mini map the box does not say what is below it (%q):\n%s", want, strings.Join(rows, "\n"))
	}
}

// TestLayoutBuildings: the paging rule itself, on entries of known heights.
func TestLayoutBuildings(t *testing.T) {
	entry := func(name string, n int) []string {
		out := []string{name}
		for i := 1; i < n; i++ {
			out = append(out, name+" line "+itoa(i+1))
		}
		return out
	}
	entries := [][]string{entry("a", 4), entry("b", 5), entry("c", 3), entry("d", 6)}
	// Everything fits: the whole list, no notes.
	lines, from, shown := layoutBuildings("head", entries, 40, 19, 2)
	if len(lines) != 19 || from != 0 || shown != 4 || strings.Contains(strings.Join(lines, "\n"), "more") {
		t.Errorf("a box with room for it all: %d lines from entry %d, %d shown: %v", len(lines), from, shown, lines)
	}
	// One row short: d does not fit whole, so it is left out and counted.
	lines, from, shown = layoutBuildings("head", entries, 40, 18, 0)
	if from != 0 || shown != 3 || len(lines) != 14 || untag(lines[len(lines)-1]) != " ↓ 1 more below: PgDn" {
		t.Errorf("a box one row short: from %d, %d shown, %v", from, shown, lines)
	}
	// The next page: what is above, then d whole.
	lines, from, shown = layoutBuildings("head", entries, 40, 18, 3)
	if from != 3 || shown != 1 || untag(lines[1]) != " ↑ 3 more above: PgUp" || len(lines) != 8 {
		t.Errorf("the last page: from %d, %d shown, %v", from, shown, lines)
	}
	// A middle page says both.
	lines, _, shown = layoutBuildings("head", entries, 40, 9, 1)
	if shown != 1 || untag(lines[1]) != " ↑ 1 more above: PgUp" || untag(lines[len(lines)-1]) != " ↓ 2 more below: PgDn" {
		t.Errorf("a middle page: %d shown, %v", shown, lines)
	}
	// No box ever holds more rows than it has.
	for h := 1; h < 22; h++ {
		for first := 0; first < len(entries); first++ {
			if lines, _, _ := layoutBuildings("head", entries, 40, h, first); len(lines) > max(h, 1) {
				t.Errorf("a %d-row box from entry %d was given %d lines", h, first, len(lines))
			}
		}
	}
}

// itoa is strconv.Itoa, for the short messages above.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg, s := n < 0, ""
	if neg {
		n = -n
	}
	for ; n > 0; n /= 10 {
		s = string(rune('0'+n%10)) + s
	}
	if neg {
		s = "-" + s
	}
	return s
}
