package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// The dashboard holds together at the sizes people play at: every box that
// shows rows writes them for the width it is drawn at, so nothing is left
// for the terminal to wrap. These tests draw the real dashboard at the
// game's smallest size, the two common ones and a roomy one, in an early, a
// middle and a late age, and fail on a row that wrapped, was cut, or split a
// rate.

// layoutSizes are the terminals the dashboard is drawn at: the minimum, the
// small-terminal size, the size the docs call large, and a roomy one.
var layoutSizes = [][2]int{{80, 24}, {100, 30}, {120, 40}, {144, 46}}

// layoutAges are an early, a middle and a late age: four resources, twelve
// (culture and soldiers among them), and all twenty-six.
var layoutAges = []string{"primitive_age", "medieval_age", "quantum_age"}

// stagedDashboard is a dashboard over a game moved into age: a few of every
// building so far, stores at mixed levels (full, nearly empty, in between),
// and five kinds of building under construction.
func stagedDashboard(t *testing.T, age string) (*Dashboard, *tview.Pages, *game.GameEngine) {
	t.Helper()
	eng := game.NewGameEngine()
	eng.SeedRNG(11)
	for _, a := range eng.Rules().AgeKeys() {
		if eng.GetState().Age == age {
			break
		}
		if a == "primitive_age" {
			continue
		}
		if err := eng.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	if got := eng.GetState().Age; got != age {
		t.Fatalf("staged the %s, want the %s", got, age)
	}
	eng.GrantTechsForTest()
	st := eng.GetState()
	counts := map[string]int{}
	n := 0
	for _, k := range sortedKeysOf(st.Buildings) {
		if b := st.Buildings[k]; b.Unlocked && b.Category != "wonder" {
			counts[k] = 2 + n%4
			n++
		}
	}
	eng.Buildings.LoadCounts(counts)
	eng.StepTicks(2)
	st = eng.GetState()
	// stock sets every store to a share of what it holds.
	stock := func(share func(n int) float64) {
		n := 0
		for _, k := range sortedKeysOf(st.Resources) {
			if rs := st.Resources[k]; rs.Unlocked {
				eng.SetStockForTest(k, rs.Storage*share(n))
				n++
			}
		}
	}
	// Each kind is ordered from full stores, so what is under construction
	// does not hang on how an age's prices sit against its storage. The
	// mixed levels follow.
	queued := 0
	for _, k := range sortedKeysOf(st.Buildings) {
		if b := st.Buildings[k]; b.Unlocked && b.AgeKey == st.Age && b.Category != "wonder" && queued < 5 {
			stock(func(int) float64 { return 1 })
			if _, err := eng.BuildMultiple(k, 1+queued%3); err == nil {
				queued++
			}
		}
	}
	stock(func(n int) float64 { return []float64{0.97, 0.31, 0.62, 0.08, 0.99, 0.45}[n%6] })
	if queued < 5 {
		t.Fatalf("only %d kinds of building under construction in the %s", queued, age)
	}
	eng.StepTicks(3)
	app := tview.NewApplication()
	pages := tview.NewPages()
	d := NewDashboard(app, eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	return d, pages, eng
}

// drawnDashboard draws the dashboard at w x h, as the app does, and returns
// the screen's cells.
func drawnDashboard(t *testing.T, d *Dashboard, pages *tview.Pages, w, h int) []tcell.SimCell {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sim.Fini)
	sim.SetSize(w, h)
	screen := theme.WrapScreen(sim)
	// Twice: the first draw gives every box its size, and the refresh after
	// it writes what depends on the sizes (the status bar, the mini map).
	for i := 0; i < 2; i++ {
		d.refresh()
		screen.Clear()
		pages.SetRect(0, 0, w, h)
		pages.Draw(screen)
		screen.Show()
	}
	cells, _, _ := sim.GetContents()
	return append([]tcell.SimCell(nil), cells...)
}

// rectText is the text drawn in a rectangle of the screen, a line per row,
// trailing blanks trimmed.
func rectText(cells []tcell.SimCell, screenW, x, y, w, h int) []string {
	lines := make([]string, h)
	for row := 0; row < h; row++ {
		var b strings.Builder
		for col := 0; col < w; col++ {
			c := cells[(y+row)*screenW+x+col]
			if len(c.Runes) == 0 {
				b.WriteByte(' ')
			} else {
				b.WriteRune(c.Runes[0])
			}
		}
		lines[row] = strings.TrimRight(b.String(), " ")
	}
	return lines
}

// checkFitted holds a box to what it wrote: every line of its text fits its
// inner width (so none was wrapped or cut), and the rows on screen are those
// lines, one row each, from the line it is scrolled to.
func checkFitted(t *testing.T, where string, v *fitView, cells []tcell.SimCell, screenW int) []string {
	t.Helper()
	x, y, w, h := v.GetInnerRect()
	if w <= 0 || h <= 0 {
		t.Errorf("%s: the box has no room inside (%dx%d)", where, w, h)
		return nil
	}
	text := strings.Split(strings.TrimRight(v.GetText(true), "\n"), "\n")
	for i, line := range text {
		if n := runeLen(strings.TrimRight(line, " ")); n > w {
			t.Errorf("%s: line %d is %d cells in a box %d wide, so it is cut: %q", where, i+1, n, w, line)
		}
	}
	top, _ := v.GetScrollOffset()
	drawn := rectText(cells, screenW, x, y, w, h)
	for row, got := range drawn {
		want := ""
		if top+row < len(text) {
			want = strings.TrimRight(text[top+row], " ")
		}
		if got != want {
			t.Errorf("%s: row %d shows %q, the box wrote %q (a wrapped or cut row)", where, row+1, got, want)
		}
	}
	return drawn
}

func TestDashboardRowsNeverWrap(t *testing.T) {
	restoreForge(t)
	for _, age := range layoutAges {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		unlocked := 0
		for _, rs := range st.Resources {
			if rs.Unlocked {
				unlocked++
			}
		}
		for _, sz := range layoutSizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", st.AgeName, w, h)
			cells := drawnDashboard(t, d, pages, w, h)

			// The Resources box: one resource a row, never a rate split.
			rows := checkFitted(t, where+", Resources", d.economyTab.resourceTV, cells, w)
			shown := 0
			for _, row := range rows {
				if strings.HasSuffix(row, "/") || strings.HasPrefix(strings.TrimSpace(row), "tick") {
					t.Errorf("%s: a rate is split after its slash: %q", where, row)
				}
				if strings.Contains(row, "/t") && !strings.Contains(row, "Ctrl") {
					shown++
					if w >= 120 && !strings.Contains(row, "/tick") {
						t.Errorf("%s: a resource row lost its rate's unit at a size with room for it: %q", where, row)
					}
					if w >= 120 && !strings.ContainsAny(row, "█░▓") {
						t.Errorf("%s: a resource row has no bar at a size with room for one: %q", where, row)
					}
				}
			}
			if shown == 0 || shown > unlocked {
				t.Errorf("%s: the Resources box shows %d resource rows of %d", where, shown, unlocked)
			}
			if d.economyTab.resPages == 1 && shown != unlocked {
				t.Errorf("%s: the Resources box shows %d of %d resources and offers no other page", where, shown, unlocked)
			}

			// Under construction: a line a building, or a count of the rest.
			build := checkFitted(t, where+", Under construction", d.economyTab.constructionTV, cells, w)
			if len(build) == 0 || strings.TrimSpace(build[0]) == "" {
				t.Errorf("%s: Under construction shows nothing with buildings under way", where)
			}

			// The Next Age strip, the Panels list and the Buildings list.
			checkFitted(t, where+", Next Age", d.ageTV, cells, w)
			panels := strings.Join(checkFitted(t, where+", Panels", d.sidebar, cells, w), "\n")
			for _, name := range sidebarPanels {
				if !strings.Contains(panels, name) {
					t.Errorf("%s: the Panels list does not show %q", where, name)
				}
			}
			for _, row := range checkFitted(t, where+", Buildings", d.economyTab.buildingTV, cells, w) {
				if strings.HasSuffix(row, "/") {
					t.Errorf("%s: the Buildings list splits a rate after its slash: %q", where, row)
				}
			}

			// The status bar and the Workers box are one line a fact.
			_, sy, _, _ := d.statusTV.GetInnerRect()
			if status := rectText(cells, w, 0, sy, w, 1)[0]; !strings.Contains(status, st.AgeName) || !strings.Contains(status, "Pop:") || !strings.Contains(status, "Morale") {
				t.Errorf("%s: the status bar lost the age, the population or morale: %q", where, status)
			}
			wx, wy, ww, wh := d.workerMiniTV.GetInnerRect()
			mini := strings.Join(rectText(cells, w, wx, wy, ww, wh), "\n")
			for _, want := range []string{"Idle:", "Housing left:", "Food use:", "Food net:"} {
				if !strings.Contains(mini, want) {
					t.Errorf("%s: the Workers box does not show %q:\n%s", where, want, mini)
				}
			}
		}
	}
}

// TestResourcesBoxReachesEveryResource: when the box cannot show every
// resource it shows a page, says so with the key that turns it, and the key
// goes round every resource and back to the first page.
func TestResourcesBoxReachesEveryResource(t *testing.T) {
	restoreForge(t)
	d, pages, eng := stagedDashboard(t, "quantum_age")
	var names []string
	st := eng.GetState()
	for _, k := range sortedKeysOf(st.Resources) {
		if rs := st.Resources[k]; rs.Unlocked {
			names = append(names, rs.Name)
		}
	}
	if len(names) != 26 {
		t.Fatalf("the Quantum Age has %d resources unlocked, want all 26", len(names))
	}
	press := func() {
		if ev := d.root.GetInputCapture()(tcell.NewEventKey(resourcePageKey, 0, tcell.ModNone)); ev != nil {
			t.Fatalf("%s was not taken by the dashboard", resourcePageKeyName)
		}
	}
	for _, sz := range layoutSizes {
		w, h := sz[0], sz[1]
		where := fmt.Sprintf("%dx%d", w, h)
		box := func() string {
			cells := drawnDashboard(t, d, pages, w, h)
			x, y, bw, bh := d.economyTab.resourceTV.GetInnerRect()
			return strings.Join(rectText(cells, w, x, y, bw, bh), "\n")
		}
		first := box()
		n := d.economyTab.resPages
		if n == 1 {
			for _, name := range names {
				if !strings.Contains(first, shortName(name, resNameCap)) && !strings.Contains(first, name) {
					t.Errorf("%s: one page, and %q is not on it:\n%s", where, name, first)
				}
			}
			continue
		}
		if !strings.Contains(first, fmt.Sprintf("of %d", len(names))) || !strings.Contains(first, resourcePageKeyName) {
			t.Errorf("%s: the box shows a page of its rows without saying so and naming %s:\n%s", where, resourcePageKeyName, first)
		}
		seen := first
		for i := 1; i < n; i++ {
			press()
			seen += "\n" + box()
		}
		rows := 0
		for _, line := range strings.Split(seen, "\n") {
			if strings.Contains(line, "/t") && !strings.Contains(line, "Ctrl") {
				rows++
			}
		}
		if rows != len(names) {
			t.Errorf("%s: %d pages show %d resource rows, want each of the %d once", where, n, rows, len(names))
		}
		press()
		if again := box(); again != first {
			t.Errorf("%s: after the last page %s does not come back to the first", where, resourcePageKeyName)
		}
	}
	// With a panel open the box is behind it, and the key is the panel's.
	d.overlayMgr.Show("help", eng.GetState())
	if ev := d.root.GetInputCapture()(tcell.NewEventKey(resourcePageKey, 0, tcell.ModNone)); ev == nil {
		t.Errorf("%s turned the Resources box's page behind an open panel", resourcePageKeyName)
	}
	d.overlayMgr.Hide()
}

// TestResourceRowsLineUp: on any width the rate column is as wide as the
// widest rate shown, so the rates end in one column and the bars start in
// one, and a row never outgrows the box.
func TestResourceRowsLineUp(t *testing.T) {
	rows := []resRow{
		storeRow(game.ResourceState{Name: "Food", Amount: 49600, Storage: 69400, Rate: 13.9}),
		storeRow(game.ResourceState{Name: "Gold", Amount: 2.66e6, Storage: 2.66e6, Rate: 3760}), // "+3.76K/tick", the longest
		storeRow(game.ResourceState{Name: "Knowledge", Amount: 32400, Storage: 69400, Rate: 17}),
		storeRow(game.ResourceState{Name: "Dark Matter Crystals", Amount: 12, Storage: 69400, Rate: -0.04}),
		faithRow(game.ResourceState{Name: "Faith", Amount: 20, Storage: 69400, Rate: 0.04}, game.CatastropheOutlook{FaithStrength: 0.41, FaithBand: game.FaithBandAt(0.41)}),
		cultureRow(game.ResourceState{Name: "Culture", Amount: 1200, Storage: 69400, Rate: 7.1}),
	}
	// From a little under the narrowest box the game draws (27 cells, at 80
	// columns) up to a roomy one.
	for w := 24; w <= 90; w++ {
		f := fitResourceFormat(rows, w)
		rateEnd, barStart := -1, -1
		for _, r := range rows {
			line := untag(f.line(r))
			if runeLen(line) > w {
				t.Errorf("width %d: row is %d cells: %q", w, runeLen(line), line)
			}
			unit := f.unit
			at := strings.Index(line, unit)
			if at < 0 {
				t.Fatalf("width %d: row has no rate: %q", w, line)
			}
			end := runeLen(line[:at]) + runeLen(unit)
			if rateEnd >= 0 && end != rateEnd {
				t.Errorf("width %d: rates end at columns %d and %d: %q", w, rateEnd, end, line)
			}
			rateEnd = end
			if f.bar > 0 {
				start := runeLen(line[:strings.IndexAny(line, "█░▓")])
				if barStart >= 0 && start != barStart {
					t.Errorf("width %d: bars start at columns %d and %d: %q", w, barStart, start, line)
				}
				barStart = start
			}
		}
	}
	// At the width of a 120-column terminal's box the long name gives way
	// to its initials, the rate keeps its unit and there is a bar.
	f := fitResourceFormat(rows, 47)
	if line := untag(f.line(rows[3])); !strings.Contains(line, "D. M. Crystals") || !strings.Contains(line, "-0.04/tick") || f.bar < resBarMin {
		t.Errorf("47 cells: want initials, the full unit and a bar: %q (bar %d)", line, f.bar)
	}
}

// TestResourceBoxLayouts: the box's lines never outnumber its rows or
// outgrow its width, whatever the size, and they degrade in order: the line
// between rows goes first, then the detail lines, then the legend's blank
// line, then the legend, and only then a page of rows.
func TestResourceBoxLayouts(t *testing.T) {
	var rows []resRow
	for i := 0; i < 26; i++ {
		rows = append(rows, storeRow(game.ResourceState{Name: fmt.Sprintf("Resource %d", i), Amount: float64(i) * 1000, Storage: 30000, Rate: float64(i) * 1.5}))
		for w := 24; w <= 70; w += 3 {
			for h := 1; h <= 60; h += 2 {
				lines, pages := layoutResourceBox(rows, w, h, 0)
				if len(lines) > h {
					t.Fatalf("%d rows in %dx%d: %d lines", len(rows), w, h, len(lines))
				}
				for _, line := range lines {
					if n := visibleLen(line); n > w {
						t.Fatalf("%d rows in %dx%d: a line of %d cells: %q", len(rows), w, h, n, untag(line))
					}
				}
				if pages < 1 || (pages == 1 && h > 1 && len(rows) > h) {
					t.Fatalf("%d rows in %dx%d: %d pages", len(rows), w, h, pages)
				}
			}
		}
	}
	seven := rows[:7]
	if lines, _ := layoutResourceBox(seven, 60, 15, 0); len(lines) != 15 || lines[1] != "" {
		t.Errorf("a roomy box puts a line between rows and ends on the legend: %d lines", len(lines))
	}
	if lines, _ := layoutResourceBox(seven, 60, 9, 0); len(lines) != 9 || lines[1] == "" || lines[7] != "" {
		t.Errorf("a shorter box closes the rows up and keeps the legend apart: %q", lines)
	}
	if lines, _ := layoutResourceBox(seven, 60, 8, 0); len(lines) != 8 || !strings.Contains(untag(lines[7]), "falling") {
		t.Errorf("a box with one row to spare keeps the legend: %q", lines)
	}
	if lines, pages := layoutResourceBox(seven, 60, 7, 0); len(lines) != 7 || pages != 1 {
		t.Errorf("a box as tall as its rows shows them all and no legend: %d lines, %d pages", len(lines), pages)
	}
	if lines, pages := layoutResourceBox(seven, 60, 6, 0); pages != 2 || !strings.Contains(untag(lines[5]), "1-5 of 7") {
		t.Errorf("a box a row short shows a page and says which: %d pages, %q", pages, lines)
	}
}

// TestConstructionRowsFit: every line fits, the bar gives way before the
// time does, and a box too short for the queue says how much more there is.
func TestConstructionRowsFit(t *testing.T) {
	st := game.GameState{TickIntervalMs: 2000, BuildQueue: []game.BuildQueueSnapshot{
		{Name: "Farm", TicksLeft: 130, TotalTicks: 150}, {Name: "Farm", TicksLeft: 140, TotalTicks: 150},
		{Name: "Stonemasons' Guild", TicksLeft: 900, TotalTicks: 1000},
		{Name: "Nuclear Extraction Plant", TicksLeft: 40, TotalTicks: 400},
	}}
	for w := 20; w <= 80; w++ {
		for h := 1; h <= 5; h++ {
			lines := layoutConstruction(st, w, h)
			if len(lines) == 0 || len(lines) > h {
				t.Fatalf("%dx%d: %d lines", w, h, len(lines))
			}
			for _, line := range lines {
				if n := visibleLen(line); n > w {
					t.Errorf("%dx%d: a line of %d cells: %q", w, h, n, untag(line))
				}
			}
		}
	}
	wide := untag(strings.Join(layoutConstruction(st, 57, 4), "\n"))
	if !strings.Contains(wide, "Farm x2") || !strings.Contains(wide, "~4m 20s left") || strings.Count(wide, "\n") != 2 {
		t.Errorf("a roomy box shows each building with its time left:\n%s", wide)
	}
	if narrow := untag(strings.Join(layoutConstruction(st, 27, 4), "\n")); strings.ContainsAny(narrow, "█░") || !strings.Contains(narrow, "~4m 20s") {
		t.Errorf("a narrow box drops the bar and keeps the time:\n%s", narrow)
	}
	if short := untag(strings.Join(layoutConstruction(st, 57, 2), "\n")); !strings.Contains(short, "+2 more under way") {
		t.Errorf("a short box says how many more are under way:\n%s", short)
	}
	if one := untag(layoutConstruction(st, 47, 1)[0]); !strings.Contains(one, "4 under way, the first in ~1m 20s") {
		t.Errorf("a box of one line counts what is under way: %q", one)
	}
	if empty := untag(layoutConstruction(game.GameState{}, 27, 1)[0]); !strings.Contains(empty, "nothing") {
		t.Errorf("an empty queue says so in 27 cells: %q", empty)
	}
}

// TestPanelsListAlwaysWhole: the Panels list shows every panel at any
// height the dashboard gives it, in one column when that fits and two when
// the screen is short, the open panel marked either way.
func TestPanelsListAlwaysWhole(t *testing.T) {
	for h := 8; h <= 30; h++ {
		text := sidebarText("factions", 20, h)
		lines := strings.Split(strings.TrimRight(untag(text), "\n"), "\n")
		if len(lines) > h {
			t.Errorf("height %d: %d lines", h, len(lines))
		}
		for _, line := range lines {
			if runeLen(line) > 20 {
				t.Errorf("height %d: a line of %d cells: %q", h, runeLen(line), line)
			}
		}
		for _, name := range sidebarPanels {
			if !strings.Contains(untag(text), name) {
				t.Errorf("height %d: %q is missing", h, name)
			}
		}
		if !strings.Contains(text, "[onaccent:accent]") {
			t.Errorf("height %d: the open panel is not marked", h)
		}
	}
	// 100x30 gives the list 14 rows: two columns. 120x40 gives it 24: one.
	if two := untag(sidebarText("", 20, 14)); !strings.HasPrefix(two, "milestones wonders") {
		t.Errorf("14 rows: want two columns read down then across, got:\n%s", two)
	}
	if one := sidebarText("", 20, 24); one != buildSidebarText("") {
		t.Errorf("24 rows: want the one-column list")
	}
}

// TestNextAgeRowKeepsGoalsWhole: a requirement is never broken across the
// strip's two rows, and a strip too small for all of them counts the ones
// already met instead of listing them, then the ones it has no room for.
func TestNextAgeRowKeepsGoalsWhole(t *testing.T) {
	st := game.GameState{
		NextAge: "iron_age", NextAgeName: "Iron Age",
		NextAgeBldReqs:      map[string]int{"lumber_mill": 8, "quarry": 8, "scriptorium": 5, "smelter": 6, "ironworks": 4, "temple": 5, "market": 3},
		Buildings:           map[string]game.BuildingState{"lumber_mill": {Count: 4}, "quarry": {Count: 3}, "scriptorium": {Count: 5}, "smelter": {Count: 2}, "ironworks": {Count: 1}, "temple": {Count: 2}},
		CurrentAgeWonderKey: "stonehenge", CurrentAgeWonderName: "Stonehenge",
	}
	goals := ageGoals(st)
	for w := 30; w <= 160; w++ {
		lines := ageProgressLines(st, w, 2)
		if len(lines) > 2 {
			t.Fatalf("width %d: %d lines", w, len(lines))
		}
		joined := untag(strings.Join(lines, "\n"))
		for _, line := range strings.Split(joined, "\n") {
			if runeLen(line) > w {
				t.Errorf("width %d: a line of %d cells: %q", w, runeLen(line), line)
			}
		}
		listed := 0
		for _, g := range goals {
			if strings.Contains(joined, g.text) {
				listed++
			} else if first := strings.Fields(g.text)[0]; strings.Contains(strings.ReplaceAll(joined, "Next Age: Iron Age", ""), first+" ") && first != "Iron" {
				t.Errorf("width %d: %q is broken across the strip:\n%s", w, g.text, joined)
			}
		}
		if listed < len(goals) && !strings.Contains(joined, "met") && !strings.Contains(joined, "more") {
			t.Errorf("width %d: %d of %d requirements listed and no count of the rest:\n%s", w, listed, len(goals), joined)
		}
	}
	if all := untag(strings.Join(ageProgressLines(st, 140, 2), "\n")); !strings.Contains(all, "✓ "+game.BuildingName("scriptorium")+" 5/5") || !strings.Contains(all, "✗ Wonder: Stonehenge") {
		t.Errorf("a wide strip lists every requirement:\n%s", all)
	}
	if small := untag(strings.Join(ageProgressLines(st, 78, 2), "\n")); !strings.Contains(small, "✓ 1 met") && !strings.Contains(small, "more") {
		t.Errorf("a narrow strip counts what it leaves out:\n%s", small)
	}
	if done := untag(ageProgressLines(game.GameState{}, 80, 2)[0]); !strings.Contains(done, "final age") {
		t.Errorf("the last age says so: %q", done)
	}
}

// TestStatusLineGivesUpTheHintFirst: the status bar keeps the age, the
// epoch, the badges, the population and morale at any width, and sheds the
// rest in order.
func TestStatusLineGivesUpTheHintFirst(t *testing.T) {
	// The widths below are counted without the DEV badge, and another test
	// in this package leaves developer mode on.
	prevDev := game.DevModeActive
	game.DevModeActive = false
	t.Cleanup(func() { game.DevModeActive = prevDev })
	st := game.GameState{AgeName: "Bronze Age", EpochKey: "stone_era", EpochName: "Stone Era", EpochIcon: "◈", EpochColor: "white",
		Morale: 1.1, MoraleMultiplier: 1.2, Workers: game.WorkerState{TotalPop: 216, MaxPop: 875},
		Milestones: game.MilestoneState{CurrentTitle: "Aspiring"}, Harbinger: &game.HarbingerView{}}
	full := untag(statusLine(st, 0))
	for _, want := range []string{"Bronze Age", "\"Aspiring\"", "◈ Stone Era", "⚑ Harbinger. Type harbinger to read it.", "Pop: 216/875", "Morale 110% (production +20%)", "type a panel name to open it · Esc: close or menu"} {
		if !strings.Contains(full, want) {
			t.Errorf("the whole line is missing %q: %q", want, full)
		}
	}
	for w := 60; w <= 200; w++ {
		line := untag(statusLine(st, w))
		if runeLen(line) > w {
			t.Errorf("width %d: the line is %d cells: %q", w, runeLen(line), line)
		}
		for _, want := range []string{"Bronze Age", "◈", "⚑ Harbinger", "Pop: 216/875", "Morale 110%"} {
			if !strings.Contains(line, want) {
				t.Errorf("width %d: the line lost %q: %q", w, want, line)
			}
		}
	}
	if line := untag(statusLine(st, 150)); strings.Contains(line, "type a panel") || !strings.Contains(line, "Esc: close or menu") || !strings.Contains(line, "Type harbinger to read it.") {
		t.Errorf("150 cells: half the hint goes first: %q", line)
	}
	if line := untag(statusLine(st, 130)); strings.Contains(line, "Esc:") || !strings.Contains(line, "Type harbinger to read it.") {
		t.Errorf("130 cells: the whole hint goes before the badge is shortened: %q", line)
	}
	if line := untag(statusLine(st, 120)); !strings.Contains(line, "⚑ Harbinger: type harbinger") || !strings.Contains(line, "(production +20%)") {
		t.Errorf("120 cells: the badge is shortened before morale's effect goes: %q", line)
	}
	// Developer mode adds its own badge, and the line still fits.
	game.DevModeActive = true
	for w := 60; w <= 200; w++ {
		if line := untag(statusLine(st, w)); runeLen(line) > w || !strings.Contains(line, "DEV") || !strings.Contains(line, "Morale 110%") {
			t.Errorf("width %d in developer mode: %d cells, %q", w, runeLen(line), line)
		}
	}
	game.DevModeActive = false
	// The worst case at the smallest size: both badges and long counts.
	st.PendingCatastrophe, st.Workers = "iron_era", game.WorkerState{TotalPop: 8311, MaxPop: 51785785}
	if line := untag(statusLine(st, 80)); runeLen(line) > 80 || !strings.Contains(line, "☄ Catastrophe") || !strings.Contains(line, "⚑ Harbinger") || !strings.Contains(line, "Pop: 8.31K/51.8M") || !strings.Contains(line, "Morale 110%") {
		t.Errorf("80 cells with both badges: %d cells, %q", runeLen(line), line)
	}
}

// TestDashboardLayoutInEveryTheme: the fitted rows, the page line's keycap
// and the two-column Panels list are legible in every theme: no glyph is
// drawn in its own background color, and no cell falls through to the
// terminal's default.
func TestDashboardLayoutInEveryTheme(t *testing.T) {
	restoreForge(t)
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		d, pages, _ := stagedDashboard(t, "quantum_age")
		for _, sz := range [][2]int{{80, 24}, {120, 40}} {
			w, h := sz[0], sz[1]
			bad := 0
			for i, c := range drawnDashboard(t, d, pages, w, h) {
				fg, bg, _ := c.Style.Decompose()
				blank := len(c.Runes) == 0 || c.Runes[0] == ' ' || pixelGlyph(c.Runes[0])
				if !bg.IsRGB() || (!blank && (!fg.IsRGB() || fg.Hex() == bg.Hex())) {
					if bad++; bad <= 3 {
						t.Errorf("%s at %dx%d: cell (%d,%d) %q is not legible (fg %v on bg %v)", th.Key, w, h, i%w, i/w, string(c.Runes), fg, bg)
					}
				}
			}
		}
	}
}

// TestDashboardLayoutDump writes the staged dashboards as plain text, to
// read in a terminal. Opt-in:
//
//	DASHBOARD_LAYOUT_DUMP=/tmp/dash go test ./ui -run TestDashboardLayoutDump
func TestDashboardLayoutDump(t *testing.T) {
	dir := os.Getenv("DASHBOARD_LAYOUT_DUMP")
	if dir == "" {
		t.Skip("set DASHBOARD_LAYOUT_DUMP=<dir> to write the staged dashboards as text")
	}
	restoreForge(t)
	for _, age := range layoutAges {
		d, pages, _ := stagedDashboard(t, age)
		for _, sz := range layoutSizes {
			cells := drawnDashboard(t, d, pages, sz[0], sz[1])
			text := strings.Join(rectText(cells, sz[0], 0, 0, sz[0], sz[1]), "\n") + "\n"
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%dx%d.txt", age, sz[0], sz[1])), []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
