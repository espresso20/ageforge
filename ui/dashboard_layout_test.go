package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
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
// small-terminal size, the size the docs call large, a roomy one, a big
// laptop's and a very big one. The first two get the classic layout, the rest
// the rich one (dash_layout.go).
var layoutSizes = [][2]int{{80, 24}, {100, 30}, {120, 40}, {144, 46}, {160, 45}, {240, 60}}

// layoutAges are an early, a middle and a late age: four resources, twelve
// (culture and soldiers among them), and all twenty-six.
var layoutAges = []string{"primitive_age", "medieval_age", "quantum_age"}

// stagedDashboard is a dashboard over a game moved into age: a few of every
// building so far, stores at mixed levels (full, nearly empty, in between),
// and five kinds of building under construction.
func stagedDashboard(t *testing.T, age string) (*Dashboard, *tview.Pages, *game.GameEngine) {
	t.Helper()
	eng := game.NewGameEngine()
	eng.SeedRNG(12)
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
	// stock sets every resource to a share of a good stock of the age.
	good := eng.Rules().StockScale(st.Age)
	stock := func(share func(n int) float64) {
		n := 0
		for _, k := range sortedKeysOf(st.Resources) {
			if rs := st.Resources[k]; rs.Unlocked {
				eng.SetStockForTest(k, good*share(n))
				n++
			}
		}
	}
	// Each kind is ordered from a full stock, so what is under construction
	// does not hang on how an age's prices sit against its stock. The
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
		unlocked := 0 // the soldiers are people: the People box lists them, not this one
		for k, rs := range st.Resources {
			if rs.Unlocked && k != "soldiers" {
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
					// Only culture and faith show progress toward something; an
					// ordinary resource has no bar, for stock has no ceiling.
					progress := strings.HasPrefix(row, " Culture") || strings.HasPrefix(row, " Faith")
					if w >= 120 && progress && !strings.ContainsAny(row, "█░▓") {
						t.Errorf("%s: a progress row has no bar at a size with room for one: %q", where, row)
					}
					if !progress && strings.ContainsAny(row, "█░▓") {
						t.Errorf("%s: a resource row shows a store bar: %q", where, row)
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

			// The Ledger, the Panels list and the Buildings list.
			checkFittedBox(t, where+", Ledger", d.ledgerTV, cells, w)
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

			// The status ribbon and the People box are one line a fact.
			_, sy, _, _ := d.statusTV.GetInnerRect()
			if status := rectText(cells, w, 0, sy, w, 1)[0]; !strings.Contains(status, st.AgeName) || !strings.Contains(status, "People") || !strings.Contains(status, "Morale") {
				t.Errorf("%s: the status ribbon lost the age, the people or morale: %q", where, status)
			}
			mini := strings.Join(checkFitted(t, where+", People", d.workerMiniTV, cells, w), "\n")
			// With something holding the population back the headcount moves to
			// the box's title (peopleBoxTitle) and its row goes to the Panels list.
			// The rich layout's box (dash_layout.go) has a row for it.
			pop := "Pop "
			if peopleBoxCompact(st, d.workerDetail) {
				pop = ""
				if title := peopleBoxTitle(st, d.workerDetail); !strings.Contains(title, "/") {
					t.Errorf("%s: the People box is compact but its title has no headcount: %q", where, title)
				}
			}
			for _, want := range []string{pop, st.Term("workers").Many, st.Term("reserve").Many, st.Term("soldiers").Many, st.Term("scholars").Many} {
				if !strings.Contains(mini, want) {
					t.Errorf("%s: the People box does not show %q:\n%s", where, want, mini)
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
		if rs := st.Resources[k]; rs.Unlocked && k != "soldiers" {
			names = append(names, rs.Name)
		}
	}
	if len(names) != 24 {
		t.Fatalf("the Quantum Age has %d resources unlocked besides the soldiers, want all 24", len(names))
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
		if want := len(names); rows != want {
			t.Errorf("%s: %d pages show %d resource rows, want each of the %d once", where, n, rows, want)
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
		storeRow(game.ResourceState{Name: "Food", Amount: 49600, Rate: 13.9}),
		storeRow(game.ResourceState{Name: "Gold", Amount: 2.66 * config.Million, Rate: 3760}), // "+3.76K/tick", the longest
		storeRow(game.ResourceState{Name: "Knowledge", Amount: 32400, Rate: 17}),
		storeRow(game.ResourceState{Name: "Dark Matter Crystals", Amount: 12, Rate: -0.04}),
		faithRow(game.ResourceState{Name: "Faith", Amount: 20, Rate: 0.04}, game.CatastropheOutlook{FaithStrength: 0.41, FaithBand: game.FaithBandAt(0.41)}),
		cultureRow(game.ResourceState{Name: "Culture", Amount: 1200, Rate: 7.1}),
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
			if f.bar > 0 && r.progress {
				at := strings.IndexAny(line, "█░▓")
				if at < 0 {
					t.Fatalf("width %d: a progress row has no bar: %q", w, line)
				}
				start := runeLen(line[:at])
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
		rows = append(rows, storeRow(game.ResourceState{Name: fmt.Sprintf("Resource %d", i), Amount: float64(i) * 1000, Rate: float64(i) * 1.5}))
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
	if two := untag(sidebarText("", 20, 14)); !strings.HasPrefix(two, "milestones stats") {
		t.Errorf("14 rows: want two columns read down then across, got:\n%s", two)
	}
	if one := sidebarText("", 20, 24); one != buildSidebarText("") {
		t.Errorf("24 rows: want the one-column list")
	}
}

// TestStatusLineGivesUpTheHintFirst: the status ribbon keeps the age, the
// badges, the people and morale at any width, and sheds the rest in order:
// the hint, the next age, the badge's words, what morale does, the title and
// last the calendar year.
func TestStatusLineGivesUpTheHintFirst(t *testing.T) {
	// The widths below are counted without the DEV badge, and another test
	// in this package leaves developer mode on.
	prevDev := game.DevModeActive
	game.DevModeActive = false
	t.Cleanup(func() { game.DevModeActive = prevDev })
	st := game.GameState{AgeName: "Bronze Age", EpochKey: "stone_era", EpochName: "Stone Era", EpochIcon: "◈",
		Morale: 1.1, MoraleMultiplier: 1.2, Workers: game.WorkerState{TotalPop: 216, MaxPop: 875},
		Milestones: game.MilestoneState{CurrentTitle: "Aspiring"}, Harbinger: &game.HarbingerView{}}
	facts := ribbonFacts{year: "1,948 BC", next: "Iron Age"}
	full := untag(statusLine(st, facts, 0))
	for _, want := range []string{"Bronze Age", "\"Aspiring\"", "⚑ Harbinger. Type harbinger to read it.", "Year 1,948 BC", "People 216/875", "Morale 110% (production +20%)", "Next: Iron Age", "type a panel name to open it · Esc: close or menu"} {
		if !strings.Contains(full, want) {
			t.Errorf("the whole line is missing %q: %q", want, full)
		}
	}
	// The facts come in the order the card names: age, year, people, morale, next age.
	last := -1
	for _, want := range []string{"Bronze Age", "Year 1,948 BC", "People 216/875", "Morale 110%", "Next: Iron Age"} {
		at := strings.Index(full, want)
		if at <= last {
			t.Errorf("%q is out of order in %q", want, full)
		}
		last = at
	}
	for w := 70; w <= 220; w++ {
		line := untag(statusLine(st, facts, w))
		if runeLen(line) > w {
			t.Errorf("width %d: the line is %d cells: %q", w, runeLen(line), line)
		}
		for _, want := range []string{"Bronze Age", "⚑ Harbinger", "People 216/875", "Morale 110%"} {
			if !strings.Contains(line, want) {
				t.Errorf("width %d: the line lost %q: %q", w, want, line)
			}
		}
	}
	// What goes first needs the most room to stay: the narrowest bar each
	// piece is still on never grows as the pieces go in the order above.
	least := func(piece string) int {
		for w := 70; w <= 260; w++ {
			if strings.Contains(untag(statusLine(st, facts, w)), piece) {
				return w
			}
		}
		return 1 << 20
	}
	order := []string{"type a panel name", "Esc: close or menu", "Next: Iron Age", "Type harbinger to read it.", "(production +20%)", "\"Aspiring\"", "Year 1,948 BC"}
	for i := 1; i < len(order); i++ {
		if least(order[i-1]) < least(order[i]) {
			t.Errorf("%q stays on a narrower bar (%d) than %q (%d), but is given up first", order[i-1], least(order[i-1]), order[i], least(order[i]))
		}
	}
	if least("type a panel name") <= least("Esc: close or menu") || least("Esc: close or menu") <= least("Type harbinger to read it.") {
		t.Errorf("the hint does not go before the badge's words: %d, %d, %d", least("type a panel name"), least("Esc: close or menu"), least("Type harbinger to read it."))
	}
	// Developer mode adds its own badge, and the line still fits.
	game.DevModeActive = true
	for w := 70; w <= 200; w++ {
		if line := untag(statusLine(st, facts, w)); runeLen(line) > w || !strings.Contains(line, "DEV") || !strings.Contains(line, "Morale 110%") {
			t.Errorf("width %d in developer mode: %d cells, %q", w, runeLen(line), line)
		}
	}
	game.DevModeActive = false
	// The worst case at the smallest size: both badges and long counts.
	st.PendingCatastrophe, st.Workers = "iron_era", game.WorkerState{TotalPop: 8311, MaxPop: 51785785}
	if line := untag(statusLine(st, facts, 80)); runeLen(line) > 80 || !strings.Contains(line, "☄ Catastrophe") || !strings.Contains(line, "⚑ Harbinger") || !strings.Contains(line, "People 8.31K/51.8M") || !strings.Contains(line, "Morale 110%") {
		t.Errorf("80 cells with both badges: %d cells, %q", runeLen(line), line)
	}
	// The plain glyph tier sets the facts off with a bar, not a dot.
	if line := untag(statusLine(st, ribbonFacts{year: "1,948 BC", plain: true}, 0)); !strings.Contains(line, "Year 1,948 BC  |  People") {
		t.Errorf("the plain tier does not set the facts off with a bar: %q", line)
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
		for _, sz := range [][2]int{{80, 24}, {100, 30}, {120, 40}, {160, 45}, {240, 60}} {
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
