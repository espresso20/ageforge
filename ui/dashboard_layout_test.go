package ui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
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
func stagedDashboard(t *testing.T, age string) (*Dashboard, *nativePages, *game.GameEngine) {
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
		must(t, eng.EnterAgeForTest(a))
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
	d, pages := newTestDashboard(eng)
	return d, pages, eng
}

// drawnDashboard draws the dashboard at w x h, as the app does, and returns
// the screen's cells.
func drawnDashboard(t *testing.T, d *Dashboard, pages *nativePages, w, h int) []simCell {
	t.Helper()
	sim := newSimScreen(t, w, h)
	screen := sim
	// Twice: the first draw gives every box its size, and the refresh after
	// it writes what depends on the sizes (the status bar, the mini map).
	for i := 0; i < 2; i++ {
		d.refresh()
		screen.Clear()
		pages.SetRect(0, 0, w, h)
		drawOn(screen, pages)
		screen.Show()
	}
	return simCells(sim).Cells()
}

// rectText is the text drawn in a rectangle of the screen, a line per row,
// trailing blanks trimmed.
func rectText(cells []simCell, screenW, x, y, w, h int) []string {
	return frame{w: screenW, cells: cells}.Region(x, y, w, h)
}

// checkFitted holds a box to what it wrote: every line of its text fits its
// inner width (so none was wrapped or cut), and the rows on screen are those
// lines, one row each, from the line it is scrolled to.
func checkFitted(t *testing.T, where string, v *fitView, cells []simCell, screenW int) []string {
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
			if rs.Unlocked && k != "soldiers" && !config.IsResourceEnded(k, st.Age) {
				unlocked++
			}
		}
		for _, sz := range layoutSizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", st.AgeName, w, h)
			cells := drawnDashboard(t, d, pages, w, h)

			// The Resources box: one resource a row, never a rate split.
			rows := checkFitted(t, where+", Resources", d.economyTab.resourceTV.fitView, cells, w)
			shown := 0
			for _, row := range rows {
				if strings.HasSuffix(row, "/") || strings.HasPrefix(strings.TrimSpace(row), "tick") {
					t.Errorf("%s: a rate is split after its slash: %q", where, row)
				}
				if isRateRow(row) {
					shown++
					if w >= 120 && !strings.Contains(row, "/s") && !strings.Contains(row, "/min") {
						t.Errorf("%s: a resource row lost its rate's unit at a size with room for it: %q", where, row)
					}
					// Culture and faith show progress toward something; an ordinary
					// resource shows magnitude, for stock has no ceiling.
					progress := strings.HasPrefix(row, " Culture") || strings.HasPrefix(row, " Faith")
					if w >= 120 && progress && !strings.ContainsAny(row, "█░▓") {
						t.Errorf("%s: a progress row has no bar at a size with room for one: %q", where, row)
					}
					if !progress && w >= 120 && !strings.ContainsAny(row, "█░") {
						t.Errorf("%s: a resource row has no magnitude bar at a size with room for one: %q", where, row)
					}
					if !progress && strings.ContainsAny(row, "▓") {
						t.Errorf("%s: a resource row shows a progress bar: %q", where, row)
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

			// The header, the Panels list and the Buildings list.
			checkHeader(t, where+", header", d, cells, w)
			panels := strings.Join(checkFitted(t, where+", Panels", d.sidebar, cells, w), "\n")
			wantAll(t, fmt.Sprintf("%s: the Panels list", where), panels, sidebarPanels...)
			checkCards(t, where+", Buildings", d, cells, w)

			// The header's first row and the People box are one line a fact.
			if status := rectText(cells, w, 0, 1, w, 1)[0]; !strings.Contains(status, st.AgeName) || !strings.Contains(status, "People") || !strings.Contains(status, "Morale") {
				t.Errorf("%s: the header lost the age, the people or morale: %q", where, status)
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
			wantAll(t, fmt.Sprintf("%s: the People box", where), mini, pop, st.Term("workers").Many, st.Term("reserve").Many, st.Term("soldiers").Many, st.Term("scholars").Many)
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
		if rs := st.Resources[k]; rs.Unlocked && k != "soldiers" && !config.IsResourceEnded(k, st.Age) {
			names = append(names, rs.Name)
		}
	}
	if len(names) != 9 {
		t.Fatalf("the Quantum Age has %d active resources besides the soldiers, want 9", len(names))
	}
	press := func() {
		if ev := captured(d.root, tk(kit.KeyChord, resourcePageKey, kit.ModCtrl)); ev != nil {
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
			if isRateRow(line) {
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
	if ev := captured(d.root, tk(kit.KeyChord, resourcePageKey, kit.ModCtrl)); ev == nil {
		t.Errorf("%s turned the Resources box's page behind an open panel", resourcePageKeyName)
	}
	d.overlayMgr.Hide()
}

// TestResourceRowsLineUp: on any width the rate column is as wide as the
// widest rate shown, so the rates end in one column and the bars start in
// one, and a row never outgrows the box.
func TestResourceRowsLineUp(t *testing.T) {
	rows := []resRow{
		storeRow(game.ResourceState{Name: "Food", Amount: 49600, Rate: 13.9}, 0),
		storeRow(game.ResourceState{Name: "Gold", Amount: 2.66 * config.Million, Rate: 3760}, 0), // "+3.76K/s", the longest
		storeRow(game.ResourceState{Name: "Knowledge", Amount: 32400, Rate: 17}, 0),
		storeRow(game.ResourceState{Name: "Dark Matter Crystals", Amount: 12, Rate: -0.04}, 0),
		faithRow(game.ResourceState{Name: "Faith", Amount: 20, Rate: 0.04}, game.CatastropheOutlook{FaithStrength: 0.41, FaithBand: game.FaithBandAt(0.41)}, 0),
		cultureRow(game.ResourceState{Name: "Culture", Amount: 1200, Rate: 7.1}, 0),
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
			rate, _ := f.rateText(r)
			at := strings.Index(line, rate)
			if at < 0 {
				t.Fatalf("width %d: row has no rate: %q", w, line)
			}
			end := runeLen(line[:at]) + runeLen(rate)
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
	if line := untag(f.line(rows[3])); !strings.Contains(line, "D. M. Crystals") || !strings.Contains(line, "-2.4/min") || f.bar < resBarMin {
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
		rows = append(rows, storeRow(game.ResourceState{Name: fmt.Sprintf("Resource %d", i), Amount: float64(i) * 1000, Rate: float64(i) * 1.5}, 0))
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
	if lines, _ := layoutResourceBox(seven, 60, 15, 0); len(lines) != 13 || lines[1] != "" {
		t.Errorf("a roomy box puts a line between rows and ends on the last row: %d lines", len(lines))
	}
	if lines, _ := layoutResourceBox(seven, 60, 9, 0); len(lines) != 7 || lines[1] == "" {
		t.Errorf("a shorter box closes the rows up: %q", lines)
	}
	if lines, pages := layoutResourceBox(seven, 60, 7, 0); len(lines) != 7 || pages != 1 {
		t.Errorf("a box as tall as its rows shows them all: %d lines, %d pages", len(lines), pages)
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

// A click on a panel's name in the dashboard's Panels list opens the panel,
// as typing the name and Enter does, and leaves what is typed in the prompt
// alone. The pointer is the hand over a name.
func TestAClickOnThePanelsListOpensThePanel(t *testing.T) {
	d, pages := newTestDashboard(game.NewGameEngine())
	d.refresh()
	d.runForTest("research")
	want := d.overlayMgr.ActiveName()
	d.overlayMgr.Hide()
	if want == "" || d.overlayMgr.HasActive() {
		t.Fatalf("staging: typing research opened %q, and Hide left a panel open: %v", want, d.overlayMgr.HasActive())
	}
	for _, size := range [][2]int{{160, 42}, {100, 30}, {90, 24}} { // the rich list, two columns, the packed list
		st := newStage(pages, size[0], size[1])
		x, y, w, h := d.sidebar.GetInnerRect()
		at := image.Pt(-1, -1)
		for j := y; j < y+h && at.X < 0; j++ {
			row := make([]rune, w)
			for i := range row {
				row[i] = st.buf.At(x+i, j).R
			}
			if i := strings.Index(string(row), "research"); i >= 0 {
				at = image.Pt(x+len([]rune(string(row)[:i]))+2, j)
			}
		}
		if at.X < 0 {
			t.Fatalf("%dx%d: the Panels list does not show research", size[0], size[1])
		}
		d.inputField.SetText("bui")
		typed := len(d.cmdHistory)
		st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
		if st.paint(); st.buf.Pointer != kit.PointerHand {
			t.Errorf("%dx%d: over a panel's name the pointer's shape is %d, want the hand", size[0], size[1], st.buf.Pointer)
		}
		st.click(at.X, at.Y)
		if got := d.overlayMgr.ActiveName(); got != want {
			t.Errorf("%dx%d: a click on research opened %q, want %q", size[0], size[1], got, want)
		}
		if d.inputField.GetText() != "bui" || len(d.cmdHistory) != typed {
			t.Errorf("%dx%d: the click changed the prompt (%q) or its history (%d lines, was %d)", size[0], size[1], d.inputField.GetText(), len(d.cmdHistory), typed)
		}
		d.inputField.SetText("")
		d.overlayMgr.Hide()
	}
}

// The wheel over the Resources box turns its page, on for a turn toward the
// player and back for one away, and stops at the first and the last page.
func TestTheWheelTurnsTheResourcesBoxPage(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "quantum_age")
	for _, sz := range layoutSizes {
		drawnDashboard(t, d, pages, sz[0], sz[1]) // the boxes get their sizes and the state its rows
		st := newStage(pages, sz[0], sz[1])
		n := d.economyTab.resPages
		if n < 2 {
			continue
		}
		x, y, _, _ := d.economyTab.resourceTV.GetInnerRect()
		st.mouse(kit.Mouse{Kind: kit.MouseMove, X: x + 2, Y: y + 1})
		wheel := func(rows int) bool {
			moved := st.mouse(kit.Mouse{Kind: kit.MouseWheel, X: x + 2, Y: y + 1, Wheel: rows})
			st.paint()
			return moved
		}
		if wheel(-1) || d.economyTab.resPage != 0 {
			t.Errorf("%dx%d: a turn back on the first page went to page %d", sz[0], sz[1], d.economyTab.resPage)
		}
		for i := 1; i < n; i++ {
			if !wheel(3) || d.economyTab.resPage != i {
				t.Errorf("%dx%d: turn %d left the box on page %d, want %d", sz[0], sz[1], i, d.economyTab.resPage, i)
			}
		}
		if wheel(1) || d.economyTab.resPage != n-1 {
			t.Errorf("%dx%d: a turn on from the last page went to page %d", sz[0], sz[1], d.economyTab.resPage)
		}
		if !wheel(-1) || d.economyTab.resPage != n-2 {
			t.Errorf("%dx%d: a turn back left the box on page %d, want %d", sz[0], sz[1], d.economyTab.resPage, n-2)
		}
		return
	}
	t.Fatal("staging: no size in layoutSizes gives the Resources box more than one page")
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

// TestDashboardLayoutInEveryTheme: the fitted rows, the page line's keycap
// and the two-column Panels list are legible in every theme: no glyph is
// drawn in its own background color, and no cell falls through to the
// terminal's default.
func TestDashboardLayoutInEveryTheme(t *testing.T) {
	restoreForge(t)
	for _, th := range theme.All() {
		must(t, theme.SetActive(th.Key))
		d, pages, _ := stagedDashboard(t, "quantum_age")
		for _, sz := range [][2]int{{80, 24}, {100, 30}, {120, 40}, {160, 45}, {240, 60}} {
			w, h := sz[0], sz[1]
			bad := 0
			for i, c := range drawnDashboard(t, d, pages, w, h) {
				fg, bg, _ := decompose(c.Style)
				blank := len(c.Runes) == 0 || c.Runes[0] == ' ' || pixelGlyph(c.Runes[0])
				if !(bg.Hex() >= 0) || (!blank && (!(fg.Hex() >= 0) || fg.Hex() == bg.Hex())) {
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
			must(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%dx%d.txt", age, sz[0], sz[1])), []byte(text), 0o644))
		}
	}
}

// rateInRow matches a rate as the Resources box writes it: a sign, a number
// and a per-second, per-minute or per-hour unit.
var rateInRow = regexp.MustCompile(`[+-][0-9][0-9.]*[A-Za-z]{0,2}/(s|min|m|h)\b`)

// isRateRow reports whether a line of the Resources box is a resource's row.
func isRateRow(row string) bool { return rateInRow.MatchString(row) }
