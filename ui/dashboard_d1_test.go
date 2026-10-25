package ui

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// The dashboard's second layout: the skyline strip under the Buildings list,
// the People box in the right half of the Under construction row, and a
// Panels list that says what each panel is for. See dash_layout.go,
// dash_skyline.go, people_rich.go and panels_table.go.

// d1Sizes are the terminals these tests draw at.
var d1Sizes = [][2]int{{80, 24}, {100, 30}, {120, 40}, {160, 45}, {240, 60}}

// stripFacts is what the dashboard drew of the strip at one size.
type stripFacts struct {
	shown          bool
	x, y, w, h     int // the strip's rectangle
	bx, by, bw, bh int // the Buildings box's
	ly, lh         int // the Log box's top row and height
}

// stripAt draws the dashboard at w x h and reports the strip and the boxes
// beside it.
func stripAt(t *testing.T, d *Dashboard, pages *tview.Pages, w, h int) (stripFacts, []tcell.SimCell) {
	t.Helper()
	cells := drawnDashboard(t, d, pages, w, h)
	var f stripFacts
	f.shown = d.mapDock.stripShown
	f.x, f.y, f.w, f.h = d.skyline.GetRect()
	f.bx, f.by, f.bw, f.bh = d.economyTab.buildingTV.GetRect()
	_, f.ly, _, f.lh = d.logTV.GetRect()
	return f, cells
}

// TestSkylineStripFollowsTheLog: the strip is as wide as the Buildings box
// and half as tall as the Log box (round down), directly under the list; it
// shows only when that leaves 6 rows inside its border and the terminal is
// 100x30 or larger; and when it hides the list has the whole column.
func TestSkylineStripFollowsTheLog(t *testing.T) {
	restoreForge(t)
	sawShown := map[[2]int]bool{}
	for _, age := range []string{"primitive_age", "medieval_age"} {
		d, pages, _ := stagedDashboard(t, age)
		for _, sz := range d1Sizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			f, cells := stripAt(t, d, pages, w, h)
			half := f.lh / 2
			if !f.shown {
				if (half-2 >= stripMinInner) && stripFits(w, h) {
					t.Errorf("%s: the Log box is %d rows, so half is %d and the strip should show", where, f.lh, half)
				}
				if strings.Contains(strings.Join(rectText(cells, w, 0, 0, w, h), "\n"), " Skyline ") {
					t.Errorf("%s: the strip is hidden but its title is on screen", where)
				}
				continue
			}
			sawShown[sz] = true
			if !stripFits(w, h) {
				t.Errorf("%s: the strip shows on a terminal under %dx%d", where, stripMinCols, stripMinRows)
			}
			if f.w != f.bw || f.x != f.bx {
				t.Errorf("%s: the strip is %d wide at column %d, the Buildings box %d wide at column %d", where, f.w, f.x, f.bw, f.bx)
			}
			if f.h != half {
				t.Errorf("%s: the strip is %d rows, half the %d-row Log box is %d", where, f.h, f.lh, half)
			}
			if f.h-2 < stripMinInner {
				t.Errorf("%s: the strip has %d rows inside, under %d", where, f.h-2, stripMinInner)
			}
			if f.by+f.bh != f.y {
				t.Errorf("%s: the Buildings box ends at row %d and the strip starts at %d", where, f.by+f.bh, f.y)
			}
			// It ends where the column does, level with the Log box.
			if f.y+f.h != f.ly+f.lh {
				t.Errorf("%s: the strip ends at row %d, the Log box at %d", where, f.y+f.h, f.ly+f.lh)
			}
			// What is drawn is a bordered box titled " Skyline " with a picture in it.
			rows := rectText(cells, w, f.x, f.y, f.w, f.h)
			if !strings.Contains(rows[0], " Skyline ") || !strings.HasPrefix(rows[0], "┌") || !strings.HasPrefix(rows[len(rows)-1], "└") {
				t.Errorf("%s: the strip is not a box titled Skyline: %q", where, rows[0])
			}
			if body := strings.TrimSpace(strings.Join(rows[1:len(rows)-1], "")); len([]rune(body)) < f.w {
				t.Errorf("%s: the strip draws almost nothing: %q", where, body)
			}
		}
	}
	// 240x60 is the size the brief checks the strip at; it has to show there.
	if !sawShown[[2]int{240, 60}] {
		t.Errorf("the strip does not show at 240x60 in the primitive or medieval age")
	}
	// At 120x40 the Log is 7 rows once the People box has its row, half of it
	// too little for a skyline: the list has the column.
	if sawShown[[2]int{120, 40}] {
		t.Errorf("the strip shows at 120x40, where half the Log box leaves under %d rows inside", stripMinInner)
	}
}

// TestSkylineStripNeedsA100x30Terminal: tall enough a Log for the strip, but a
// terminal under 100 columns or 30 rows keeps the list's whole column.
func TestSkylineStripNeedsA100x30Terminal(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "primitive_age")
	for _, c := range []struct {
		w, h int
		show bool
	}{
		{100, 60, true},  // the classic layout, with a tall Log
		{99, 60, false},  // a column short of 100
		{100, 29, false}, // a row short of 30
		{80, 70, false},
	} {
		f, _ := stripAt(t, d, pages, c.w, c.h)
		if f.shown != c.show {
			t.Errorf("%dx%d: strip shown = %v, want %v (Log box %d rows)", c.w, c.h, f.shown, c.show, f.lh)
		}
		if !c.show && f.bh+f.by != f.ly+f.lh {
			t.Errorf("%dx%d: the strip is hidden but the Buildings box does not take the column", c.w, c.h)
		}
	}
}

func TestStripHeight(t *testing.T) {
	for logRows, want := range map[int]int{0: 0, 11: 0, 12: 6, 15: 7, 16: 8, 17: 8, 18: 9, 40: 20} {
		if got := stripHeight(logRows); got != want {
			t.Errorf("a %d-row Log box: strip %d rows, want %d", logRows, got, want)
		}
	}
	for _, c := range []struct {
		w, h int
		fits bool
	}{{100, 30, true}, {99, 30, false}, {100, 29, false}, {240, 60, true}} {
		if stripFits(c.w, c.h) != c.fits {
			t.Errorf("stripFits(%d, %d) = %v", c.w, c.h, !c.fits)
		}
	}
}

// TestSkylineStripMotion: with the motion setting on the strip's picture
// moves between two moments; with it off it holds still.
func TestSkylineStripMotion(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "medieval_age")
	base := time.Now()
	moments := []time.Duration{0, 625 * time.Millisecond, 1250 * time.Millisecond, 2500 * time.Millisecond}
	picture := func() []string {
		var out []string
		for _, at := range moments {
			d.skyline.start = base
			d.skyline.now = func() time.Time { return base.Add(at) }
			f, cells := stripAt(t, d, pages, 240, 60)
			if !f.shown {
				t.Fatalf("the strip is not shown at 240x60")
			}
			out = append(out, strings.Join(rectText(cells, 240, f.x, f.y, f.w, f.h), "\n"))
		}
		return out
	}
	set := d.mapSettings()
	set.Motion = true
	d.saveMapSettings(set)
	moving := picture()
	still := 0
	for _, p := range moving[1:] {
		if p == moving[0] {
			still++
		}
	}
	if still == len(moving)-1 {
		t.Errorf("motion on: the strip looks the same at every moment")
	}
	set.Motion = false
	d.saveMapSettings(set)
	held := picture()
	for i, p := range held[1:] {
		if p != held[0] {
			t.Errorf("motion off: the strip changed between moment 0 and %d:\n%s\n\n%s", i+1, held[0], p)
		}
	}
}

// TestSkylineStripTakesNoKeys: the strip is a picture; it has no focus, no
// cursor and takes no keys.
func TestSkylineStripTakesNoKeys(t *testing.T) {
	restoreForge(t)
	d, _, _ := stagedDashboard(t, "medieval_age")
	if d.skyline.GetInputCapture() != nil || d.skyline.HasFocus() {
		t.Errorf("the skyline strip takes keys or the focus")
	}
	d.mapDock.Focus(func(p tview.Primitive) {
		if p == tview.Primitive(d.skyline) {
			t.Errorf("the dock hands the focus to the strip")
		}
	})
}

// d1Boxes returns the rectangles the second layout is built from.
func d1Rect(v interface{ GetRect() (int, int, int, int) }) [4]int {
	x, y, w, h := v.GetRect()
	return [4]int{x, y, w, h}
}

// TestPeopleBoxBesideUnderConstruction: from 120x36 up the People box is the
// right half of the Under construction row, as tall as the row and with the
// stacked bar, and the Panels list has the whole right column; below that the
// People box is where it always was, under the Panels list.
func TestPeopleBoxBesideUnderConstruction(t *testing.T) {
	restoreForge(t)
	for _, age := range layoutAges {
		d, pages, _ := stagedDashboard(t, age)
		var st game.GameState
		for _, sz := range d1Sizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			cells := drawnDashboard(t, d, pages, w, h)
			st = *d.lastState
			con, ppl, pan := d1Rect(d.economyTab.constructionTV), d1Rect(d.workerMiniTV), d1Rect(d.sidebar)
			col := d1Rect(d.economyTab.leftCol)
			if !dashRich(w, h) {
				// The classic layout: People under the Panels list, Under
				// construction the column's whole width.
				if ppl[0] != pan[0] || ppl[1] != pan[1]+pan[3] || con[2] != col[2] {
					t.Errorf("%s: the classic layout moved: People %v, Panels %v, Under construction %v", where, ppl, pan, con)
				}
				continue
			}
			if con[0] != col[0] || con[1] != ppl[1] || con[3] != ppl[3] {
				t.Errorf("%s: Under construction %v and People %v are not one row", where, con, ppl)
			}
			if ppl[0] != con[0]+con[2] || ppl[0]+ppl[2] != col[0]+col[2] {
				t.Errorf("%s: People %v is not the right half of the row %v", where, ppl, col)
			}
			if d := con[2] - ppl[2]; d < 0 || d > 1 {
				t.Errorf("%s: the halves are %d and %d wide", where, con[2], ppl[2])
			}
			if want := richPeopleHeight(h); ppl[3] != want {
				t.Errorf("%s: the People box is %d rows, want %d", where, ppl[3], want)
			}
			_, bodyY, _, bodyH := d.contentArea.GetRect()
			if pan[1] != bodyY || pan[3] != bodyH {
				t.Errorf("%s: the Panels list %v is not the whole column (rows %d to %d)", where, pan, bodyY, bodyY+bodyH)
			}
			if pan[2] != richSidebarW {
				t.Errorf("%s: the Panels column is %d wide, want %d", where, pan[2], richSidebarW)
			}

			// The box says everything the brief asks of it, and fits.
			rows := checkFitted(t, where+", People", d.workerMiniTV, cells, w)
			text := strings.Join(rows, "\n")
			bar := false
			for _, row := range rows {
				if strings.Contains(row, "█") && !strings.ContainsAny(row, "abcdefghijklmnopqrstuvwxyz") {
					bar = true
				}
			}
			if !bar {
				t.Errorf("%s: the People box has no bar across the homes:\n%s", where, text)
			}
			for _, want := range []string{"Pop ", st.Term("workers").Many, st.Term("reserve").Many, st.Term("soldiers").Many, st.Term("scholars").Many} {
				if !strings.Contains(text, want) {
					t.Errorf("%s: the People box does not show %q:\n%s", where, want, text)
				}
			}
			if !strings.Contains(text, "Room to grow") && !strings.Contains(text, "Held by") {
				t.Errorf("%s: the People box does not say what holds the population back:\n%s", where, text)
			}
			checkFitted(t, where+", Under construction", d.economyTab.constructionTV, cells, w)
		}
	}
}

// TestPeopleRichBoxLines: the box's lines for a town with soldiers on posts
// and out, a reserve and a held-back population, at the sizes it is drawn at.
func TestPeopleRichBoxLines(t *testing.T) {
	st := game.GameState{Age: "bronze_age"}
	st.Workers.TotalPop, st.Workers.MaxPop, st.Workers.TotalIdle, st.Workers.Scholars = 1061, 1113, 412, 30
	st.Workers.Soldiers = 108
	st.Military.SoldierCap, st.Military.SoldiersAway = 140, 20
	st.Workers.HeldBy, st.Workers.FoodNet, st.Workers.FoodFeeds = game.HeldFood, 0.45, 3

	// A roomy box: every fact, the soldiers' posts and the ones out.
	full := plainText(peopleRichText(st, 40, 8))
	flat := strings.Join(strings.Fields(full), " ")
	for _, want := range []string{"Pop 1061/1113", "95% full", "Craftsfolk 511", "Villagers 412", "Spearmen 108/140 · 20 out", "Scribes 30", "Held by: food", "+0.45 food a tick feeds 3 more"} {
		if !strings.Contains(flat, want) {
			t.Errorf("a roomy box lacks %q:\n%s", want, full)
		}
	}
	// And the bar, one cell for each share of the homes.
	if lines := strings.Split(full, "\n"); !strings.ContainsAny(lines[1], "█") || runeLen(lines[1]) != 40 {
		t.Errorf("the second line is not a 40-cell bar: %q", lines[1])
	}

	// Fewer rows give up lines in order: the food detail, the bar, the headcount.
	for _, c := range []struct {
		rows  int
		lines int
		has   []string
		lacks []string
	}{
		{8, 8, []string{"+0.45", "Pop"}, nil},
		{7, 7, []string{"Held by: food", "Pop"}, []string{"+0.45"}},
		{6, 6, []string{"Pop"}, []string{"█"}},
		{5, 5, []string{"Held by: food"}, []string{"Pop"}},
		{4, 4, []string{"Scribes"}, []string{"Held by"}},
	} {
		text := plainText(peopleRichText(st, 22, c.rows))
		lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
		if len(lines) != c.lines {
			t.Errorf("%d rows: %d lines:\n%s", c.rows, len(lines), text)
		}
		for _, line := range lines {
			if runeLen(line) > 22 {
				t.Errorf("%d rows: a line of %d cells in a box 22 wide: %q", c.rows, runeLen(line), line)
			}
		}
		for _, s := range c.has {
			if !strings.Contains(text, s) {
				t.Errorf("%d rows: no %q:\n%s", c.rows, s, text)
			}
		}
		for _, s := range c.lacks {
			if strings.Contains(text, s) {
				t.Errorf("%d rows: still has %q:\n%s", c.rows, s, text)
			}
		}
	}
	// With the headcount off its line, the title carries it.
	if title := peopleRichTitle(st, 5); !strings.Contains(title, "1.06K/1.11K") && !strings.Contains(title, "1061/1113") {
		t.Errorf("the short box's title lacks the headcount: %q", title)
	}
	if title := peopleRichTitle(st, 7); title != " People " {
		t.Errorf("a box with the headcount line has the title %q", title)
	}
	// Widths from 18 up (the narrowest box the game draws is 20 inside) never overflow, nor the unbounded one a never-drawn box is written for.
	for w := 18; w <= 70; w++ {
		for _, line := range strings.Split(plainText(peopleRichText(st, w, 8)), "\n") {
			if runeLen(line) > w {
				t.Errorf("width %d: a line of %d cells: %q", w, runeLen(line), line)
			}
		}
	}
	if huge := peopleRichText(st, 1<<16, 1<<16); len(huge) > 4096 {
		t.Errorf("the box written for no size at all is %d bytes", len(huge))
	}

	// Nothing holding the population back, a town full of people, and housing.
	st.Workers.HeldBy = game.HeldNone
	if text := plainText(peopleRichText(st, 30, 8)); !strings.Contains(text, "Room to grow") || !strings.Contains(text, "52 homes empty") {
		t.Errorf("a town with room says so:\n%s", text)
	}
	st.Workers.HeldBy = game.HeldHousing
	if text := plainText(peopleRichText(st, 30, 8)); !strings.Contains(text, "Held by: housing") || !strings.Contains(text, "every home is full") {
		t.Errorf("a town held by housing says so:\n%s", text)
	}
}

// TestPeopleBarMatchesThePanel: the box's bar divides the homes as the
// People panel's does.
func TestPeopleBarMatchesThePanel(t *testing.T) {
	for _, c := range []struct {
		counts []int
		homes  int
		w      int
		want   []int
	}{
		{[]int{50, 25, 15, 10}, 100, 20, []int{10, 5, 3, 2}},
		{[]int{50, 25, 15, 10}, 200, 20, []int{5, 3, 2, 1}}, // half the homes empty
		{[]int{100, 0, 0, 1}, 50, 10, []int{10, 0, 0, 1}},   // more people than homes: the bar is the people
		{[]int{0, 0, 0, 0}, 0, 10, []int{0, 0, 0, 0}},
		{[]int{1, 1, 1, 1}, 10000, 6, []int{1, 1, 1, 1}}, // every kind is at least a cell
	} {
		got := stackWidths(c.counts, c.homes, c.w)
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%v of %d homes in %d cells: %v, want %v", c.counts, c.homes, c.w, got, c.want)
		}
	}
}

// TestPanelsTableCoversEveryPanel: the grouped Panels list names every panel
// of the Panels list once, in the groups the owner asked for, and each job
// line is short and plain.
func TestPanelsTableCoversEveryPanel(t *testing.T) {
	want := map[string][]string{
		"Town":     {"people", "plan", "wonders", "map"},
		"Progress": {"milestones", "research", "charter", "epoch", "history"},
		"World":    {"expedition", "army", "trade", "factions", "harbinger"},
		"You":      {"badges", "stats", "logs", "help"},
	}
	var all []string
	for _, g := range panelGroups {
		var names []string
		for _, p := range g.panels {
			names = append(names, p.name)
			all = append(all, p.name)
			if p.job.String() == "" || runeLen(p.job.String()) > panelJobMax {
				t.Errorf("%s: the job line %q is empty or over %d cells", p.name, p.job.String(), panelJobMax)
			}
			if strings.ContainsAny(p.job.String(), "!—–") {
				t.Errorf("%s: the job line %q has an exclamation mark or a dash", p.name, p.job.String())
			}
			if runeLen(p.name) > panelNameW {
				t.Errorf("%s is wider than the name column (%d)", p.name, panelNameW)
			}
		}
		if fmt.Sprint(names) != fmt.Sprint(want[g.title.String()]) {
			t.Errorf("group %s lists %v, want %v", g.title.String(), names, want[g.title.String()])
		}
	}
	sort.Strings(all)
	have := append([]string(nil), sidebarPanels...)
	sort.Strings(have)
	if fmt.Sprint(all) != fmt.Sprint(have) {
		t.Errorf("the table lists %v, the Panels list %v", all, have)
	}
	if got, want := panelsRichRows(), len(sidebarPanels)+len(panelGroups); got != want {
		t.Errorf("the rich list is %d rows, want %d", got, want)
	}
}

// TestPanelsRichText: the list in groups, at every box size that takes it,
// is whole, fits, marks the open panel, and is what sidebarText gives a box
// that big and the packed names for one that is not.
func TestPanelsRichText(t *testing.T) {
	for h := panelsRichRows(); h <= panelsRichRows()+6; h++ {
		for _, w := range []int{panelsRichWidth(), panelsRichWidth() + 1, 40, 70} {
			text := sidebarText("trade", w, h)
			lines := strings.Split(strings.TrimRight(untag(text), "\n"), "\n")
			if len(lines) > h {
				t.Errorf("%dx%d: %d lines", w, h, len(lines))
			}
			for _, line := range lines {
				if runeLen(line) > w {
					t.Errorf("%dx%d: a line of %d cells: %q", w, h, runeLen(line), line)
				}
			}
			plain := untag(text)
			for _, g := range panelGroups {
				if !strings.Contains(plain, strings.ToUpper(g.title.String())) {
					t.Errorf("%dx%d: no heading for %s", w, h, g.title.String())
				}
				for _, p := range g.panels {
					if !strings.Contains(plain, p.name+strings.Repeat(" ", panelNameW-len(p.name))+" "+p.job.String()) {
						t.Errorf("%dx%d: %s and its job are not on one row", w, h, p.name)
					}
				}
			}
			if !strings.Contains(text, "[onaccent:accent] trade") {
				t.Errorf("%dx%d: the open panel is not marked", w, h)
			}
		}
	}
	// Too narrow or too short for the descriptions: today's packed names.
	for _, c := range [][2]int{{panelsRichWidth() - 1, 40}, {40, panelsRichRows() - 1}, {20, 14}, {20, 8}} {
		text := untag(sidebarText("", c[0], c[1]))
		if strings.Contains(text, "TOWN") || strings.Contains(text, "tech tree") {
			t.Errorf("%dx%d: a box too small for descriptions has them:\n%s", c[0], c[1], text)
		}
		for _, name := range sidebarPanels {
			if !strings.Contains(text, name) {
				t.Errorf("%dx%d: %q is missing", c[0], c[1], name)
			}
		}
	}
	// With room to spare the groups are set apart by a blank row.
	spare := strings.Split(untag(panelsRichText("", 40, panelsRichRows()+len(panelGroups)-1)), "\n")
	blanks := 0
	for _, line := range spare {
		if strings.TrimSpace(line) == "" {
			blanks++
		}
	}
	if blanks != len(panelGroups)-1+1 { // one between each pair, and the line break at the end
		t.Errorf("a list with room to spare has %d blank rows: %q", blanks, spare)
	}
}

// TestPanelsListOnTheDashboard: at 120x40 and up the Panels list is the
// grouped one with a line on each panel's job; at 80x24 and 100x30 it is
// today's packed names. Every panel name is on screen either way.
func TestPanelsListOnTheDashboard(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "medieval_age")
	for _, sz := range d1Sizes {
		w, h := sz[0], sz[1]
		where := fmt.Sprintf("%dx%d", w, h)
		cells := drawnDashboard(t, d, pages, w, h)
		panels := strings.Join(checkFitted(t, where+", Panels", d.sidebar, cells, w), "\n")
		for _, name := range sidebarPanels {
			if !strings.Contains(panels, name) {
				t.Errorf("%s: the Panels list does not show %q", where, name)
			}
		}
		described := strings.Contains(panels, "the tech tree") && strings.Contains(panels, "TOWN") && strings.Contains(panels, "YOU")
		if rich := dashRich(w, h); described != rich {
			t.Errorf("%s: the Panels list has descriptions = %v, want %v:\n%s", where, described, rich, panels)
		}
		if described {
			for _, g := range panelGroups {
				for _, p := range g.panels {
					if !strings.Contains(panels, p.job.String()) {
						t.Errorf("%s: the Panels list lost the job line %q", where, p.job.String())
					}
				}
			}
		}
	}
}

// TestEconHeightsRich: the row under the Resources box is the People box's
// height whenever the column has room, the Resources box gives way before the
// People box or the Log do, and the three always add up to the column.
func TestEconHeightsRich(t *testing.T) {
	for h := 14; h <= 90; h++ {
		for want := 4; want <= 40; want++ {
			for _, people := range []int{richPeopleRows, richPeopleDetailRows} {
				r, c, l := econHeightsRich(h, want, people)
				if r+c+l != h || r < 0 || c < 0 || l < 0 {
					t.Fatalf("h %d want %d: %d + %d + %d", h, want, r, c, l)
				}
				if h >= 28 && (c != people || l < 5) {
					t.Errorf("h %d want %d people %d: construction %d, log %d", h, want, people, c, l)
				}
				// The Resources box has its half, or what it needs when that is
				// more, as far as the People box and the Log leave it.
				if wantR := min(max(h*3/6, want), max(6, h-people-5)); h >= 28 && r != wantR {
					t.Errorf("h %d want %d people %d: Resources %d, want %d", h, want, people, r, wantR)
				}
			}
		}
	}
}

// TestSkylineStripInEveryTheme: the strip is legible in every theme, light
// ones included: every cell has real colors, and a glyph is never drawn in
// the color of its own background.
func TestSkylineStripInEveryTheme(t *testing.T) {
	restoreForge(t)
	for _, th := range theme.All() {
		must(t, theme.SetActive(th.Key))
		d, pages, _ := stagedDashboard(t, "medieval_age")
		f, cells := stripAt(t, d, pages, 240, 60)
		if !f.shown {
			t.Fatalf("%s: the strip is not shown at 240x60", th.Key)
		}
		bad := 0
		for row := 0; row < f.h; row++ {
			for col := 0; col < f.w; col++ {
				c := cells[(f.y+row)*240+f.x+col]
				fg, bg, _ := c.Style.Decompose()
				blank := len(c.Runes) == 0 || c.Runes[0] == ' ' || pixelGlyph(c.Runes[0])
				if !bg.IsRGB() || (!blank && (!fg.IsRGB() || fg.Hex() == bg.Hex())) {
					if bad++; bad <= 3 {
						t.Errorf("%s: strip cell (%d,%d) %q is not legible (fg %v on bg %v)", th.Key, col, row, string(c.Runes), fg, bg)
					}
				}
			}
		}
	}
}

// TestPeopleBoxFollowsAResize: a terminal resized across the rich layout's
// edge redraws the People box in the other layout's words at once, without
// waiting for the next refresh.
func TestPeopleBoxFollowsAResize(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	drawn := func(w, h int) {
		sim := tcell.NewSimulationScreen("UTF-8")
		must(t, sim.Init())
		defer sim.Fini()
		sim.SetSize(w, h)
		scr := theme.WrapScreen(sim)
		pages.SetRect(0, 0, w, h)
		pages.Draw(scr)
		scr.Show()
	}
	hasBar := func() bool { return strings.Contains(d.workerMiniTV.GetText(true), "█") }
	drawnDashboard(t, d, pages, 100, 30)
	if hasBar() {
		t.Fatalf("the classic People box has a bar")
	}
	for i, c := range []struct {
		w, h int
		bar  bool
	}{{120, 40, true}, {100, 30, false}, {160, 45, true}, {119, 40, false}, {120, 35, false}, {120, 36, true}} {
		drawn(c.w, c.h)
		if hasBar() != c.bar {
			t.Errorf("step %d, resized to %dx%d: the People box has a bar = %v, want %v", i, c.w, c.h, hasBar(), c.bar)
		}
	}
	// The edge can be crossed with the box the same size on both sides: one
	// row taller at 121 columns, with something holding the population back.
	drawnDashboard(t, d, pages, 121, 35)
	st := *d.lastState
	st.Workers.HeldBy, st.Workers.FoodNet, st.Workers.FoodFeeds = game.HeldFood, 0.45, 3
	d.lastState = &st
	d.refreshWorkerMini(st)
	drawn(121, 35)
	if x, y, w, h := d.workerMiniTV.GetInnerRect(); hasBar() {
		t.Fatalf("the classic People box at 121x35 has a bar (%d,%d %dx%d)", x, y, w, h)
	}
	_, _, cw, ch := d.workerMiniTV.GetInnerRect()
	drawn(121, 36)
	if _, _, w, h := d.workerMiniTV.GetInnerRect(); w != cw || h != ch {
		t.Skipf("the box changed size across the edge (%dx%d to %dx%d), so the resize alone rewrites it", cw, ch, w, h)
	}
	if !hasBar() {
		t.Errorf("a box the same size on both sides of the edge was not rewritten:\n%s", d.workerMiniTV.GetText(true))
	}
}
