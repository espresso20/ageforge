package ui

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// The Ledger is the top of the dashboard under the status ribbon: the long
// bar to the next age, what it waits on and the four pillars. These tests
// draw the real dashboard at the three sizes it is written for, hold its
// numbers to the Advance panel's, and hold what it waits on and what it is
// short of to the rules in dash_ledger.go.

// ledgerOf works the Ledger out for a snapshot, as a refresh does.
func ledgerOf(st game.GameState, eng *game.GameEngine) *ledger {
	l := &ledger{}
	l.update(st, eng)
	return l
}

// ledgerSizes are the three terminals the Ledger is written for: the smallest
// (compact), the common one (standard) and a big one (full).
var ledgerSizes = [][2]int{{80, 24}, {100, 30}, {160, 45}}

// plainTier puts the dashboard on the plain glyph tier, for the session.
func plainTier(d *Dashboard) {
	s := resolveMapSettings(nil)
	s.Tier = mapmodel.TierASCII
	d.mapLocal = &s
}

// ledgerRowsOn draws the dashboard and returns what the Ledger wrote, row by
// row, held to its box.
func ledgerRowsOn(t *testing.T, d *Dashboard, pages *tview.Pages, w, h int) []string {
	t.Helper()
	cells := drawnDashboard(t, d, pages, w, h)
	return checkFittedBox(t, fmt.Sprintf("Ledger at %dx%d", w, h), d.ledgerTV, cells, w)
}

// checkFittedBox holds the Ledger to what it wrote, as checkFitted holds a
// text box: no row wider than the box, and the rows on screen are the rows it
// wrote, one each.
func checkFittedBox(t *testing.T, where string, b *ledgerBox, cells []tcell.SimCell, screenW int) []string {
	t.Helper()
	x, y, w, h := b.GetInnerRect()
	if w <= 0 || h <= 0 {
		t.Errorf("%s: the box has no room inside (%dx%d)", where, w, h)
		return nil
	}
	text := strings.Split(b.GetText(true), "\n")
	if len(b.lines) > h {
		t.Errorf("%s: %d rows written for a box %d tall", where, len(b.lines), h)
	}
	for i, line := range text {
		if n := runeLen(strings.TrimRight(line, " ")); n > w {
			t.Errorf("%s: row %d is %d cells in a box %d wide, so it is cut: %q", where, i+1, n, w, line)
		}
	}
	drawn := rectText(cells, screenW, x, y, w, h)
	for row, got := range drawn {
		want := ""
		if row < len(text) {
			want = strings.TrimRight(text[row], " ")
		}
		if got != want {
			t.Errorf("%s: row %d shows %q, the box wrote %q", where, row+1, got, want)
		}
	}
	return drawn
}

// TestLedgerHeights: the rows the Ledger takes for each size, and the
// dashboard gives it exactly those, right under the toast line.
func TestLedgerHeights(t *testing.T) {
	for _, c := range []struct{ w, h, want int }{
		{80, 24, 1}, {99, 60, 1}, {100, 29, 1}, // the compact Ledger: the long bar
		{100, 30, 2}, {119, 60, 2}, {120, 36, 2}, {120, 38, 2}, // the standard one: the bar and the pillars
		{120, 39, 4}, {160, 45, 4}, {240, 60, 4}, // the full one: the bar, the waiting line, the pillars in two rows
	} {
		if got := ledgerHeight(c.w, c.h); got != c.want {
			t.Errorf("%dx%d: %d rows, want %d", c.w, c.h, got, c.want)
		}
	}
	// The smallest terminal keeps the badge line, the ribbon, the toast line
	// and the one row of the Ledger over the boxes.
	if dashHeaderRows != 1+1+1+ledgerCompactRows {
		t.Errorf("dashHeaderRows = %d", dashHeaderRows)
	}
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	for _, sz := range append(layoutSizes, ledgerSizes...) {
		w, h := sz[0], sz[1]
		drawnDashboard(t, d, pages, w, h)
		x, y, lw, lh := d.ledgerTV.GetRect()
		if want := ledgerHeight(w, h); x != 0 || y != 3 || lw != w || lh != want {
			t.Errorf("%dx%d: the Ledger is at (%d,%d) %dx%d, want (0,3) %dx%d", w, h, x, y, lw, lh, w, want)
		}
		if _, cy, _, _ := d.contentArea.GetRect(); cy != 3+lh {
			t.Errorf("%dx%d: the boxes start at row %d, the Ledger ends at %d", w, h, cy, 3+lh)
		}
	}
}

// TestLedgerAtThreeSizes draws the real dashboard at 80x24, 100x30 and
// 160x45 and holds each Ledger to what it is for.
func TestLedgerAtThreeSizes(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"bronze_age", "medieval_age", "quantum_age"} {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		next := st.NextAgeName
		for _, sz := range ledgerSizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			rows := ledgerRowsOn(t, d, pages, w, h)
			text := strings.Join(rows, "\n")
			// Every size: the long bar to the next age, with its percent, and
			// what it waits on.
			if !strings.Contains(rows[0], "To the "+next) || !strings.Contains(rows[0], "%") || !strings.ContainsAny(rows[0], "█░") {
				t.Errorf("%s: no long bar to the %s:\n%s", where, next, text)
			}
			if !strings.Contains(text, "Waiting on: ") && !strings.Contains(text, "Needs: ") {
				t.Errorf("%s: the Ledger does not say what it waits on:\n%s", where, text)
			}
			pillars := 0
			for _, name := range []string{"Knowledge", "People", "Industry", "Monument"} {
				if strings.Contains(strings.ToLower(text), strings.ToLower(name)) {
					pillars++
				}
			}
			switch ledgerTierFor(w, h) {
			case ledgerCompact:
				// The long bar, on one row: no pillars.
				if len(rows) != 1 || strings.Contains(text, "Knowledge") {
					t.Errorf("%s: the compact Ledger is not the long bar alone:\n%s", where, text)
				}
			case ledgerStandard:
				if !strings.Contains(rows[1], "Knowledge") || !strings.Contains(rows[1], "Monument") || len(rows) != 2 {
					t.Errorf("%s: the standard Ledger is not the bar and a row of pillars:\n%s", where, text)
				}
			case ledgerFull:
				if pillars != 4 || len(rows) != 4 || !strings.Contains(rows[1], "Waiting on: ") || !strings.Contains(rows[2], "KNOWLEDGE") {
					t.Errorf("%s: the full Ledger lacks its waiting line or its pillars:\n%s", where, text)
				}
			}
			// The resources are the Resources panel's: the Ledger has no cards,
			// no trend lines and no count of resources it leaves out.
			if strings.ContainsAny(text, "▲▼▬▶▁▂▃▄▅▆▇") || strings.Contains(text, "rising") || strings.Contains(text, "falling") {
				t.Errorf("%s: the Ledger draws a resource card:\n%s", where, text)
			}
		}
	}
}

// TestLedgerInPlainASCII: on the plain glyph tier the Ledger is nothing but
// ASCII at every size, its bars in plain marks.
func TestLedgerInPlainASCII(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"bronze_age", "quantum_age"} {
		d, pages, _ := stagedDashboard(t, age)
		plainTier(d)
		for _, sz := range ledgerSizes {
			w, h := sz[0], sz[1]
			rows := ledgerRowsOn(t, d, pages, w, h)
			for i, row := range rows {
				for _, r := range row {
					if r > 0x7e {
						t.Errorf("%s at %dx%d: row %d has %q: %q", age, w, h, i+1, string(r), row)
						break
					}
				}
			}
			if !strings.Contains(strings.Join(rows, ""), "#") {
				t.Errorf("%s at %dx%d: no bar in plain marks:\n%s", age, w, h, strings.Join(rows, "\n"))
			}
		}
	}
}

// TestLedgerReadsTheAdvancePanel: the long bar, the pillars' meters and the
// year are the Advance panel's own numbers, from the same view and the same
// held calendar.
func TestLedgerReadsTheAdvancePanel(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"stone_age", "bronze_age", "medieval_age"} {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		v := eng.HoldCalendar(st.AdvanceView())
		panel := renderAdvance(v, advanceUI{}, 100, 30).String()
		l := ledgerOf(st, eng)

		pct := sharePercent(v.Share, v.Share >= 1)
		if l.data.Pct != pct {
			t.Errorf("%s: the long bar says %d%%, the panel's share is %d%%", age, l.data.Pct, pct)
		}
		if want := fmt.Sprintf("%d%% of the way to the %s", pct, v.NextName); !strings.Contains(panel, want) {
			t.Errorf("%s: the panel does not say %q:\n%s", age, want, panel)
		}
		if !strings.Contains(panel, "Year "+v.YearText) {
			t.Errorf("%s: the panel does not show the year %s", age, v.YearText)
		}
		if len(l.data.Pillars) != len(v.Pillars) || len(v.Pillars) != 4 {
			t.Fatalf("%s: %d pillars in the Ledger, %d in the view", age, len(l.data.Pillars), len(v.Pillars))
		}
		for i, p := range v.Pillars {
			status, _ := pillarStatus(p)
			lp := l.data.Pillars[i]
			if lp.Name != p.Name || lp.Pct != sharePercent(p.Progress, false) || lp.Met != p.Met || lp.Status != status || lp.Share != p.Progress {
				t.Errorf("%s: pillar %s is %+v in the Ledger, %+v in the view", age, p.Name, lp, p)
			}
			if !strings.Contains(panel, status) {
				t.Errorf("%s: the panel does not show %q for %s", age, status, p.Name)
			}
		}

		// What is drawn: the percent and the year, and the same fill as the
		// panel's bars give the same share.
		for _, sz := range ledgerSizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			rows := ledgerRowsOn(t, d, pages, w, h)
			if !strings.Contains(rows[0], fmt.Sprintf("%3d%%", pct)) {
				t.Errorf("%s: the long bar's row has no %d%%: %q", where, pct, rows[0])
			}
			cells := drawnDashboard(t, d, pages, w, h)
			if ribbon := rectText(cells, w, 0, 1, w, 1)[0]; !strings.Contains(ribbon, "Year "+v.YearText) {
				t.Errorf("%s: the ribbon does not show the year %s: %q", where, v.YearText, ribbon)
			}
			filled, empty := strings.Count(rows[0], "█"), strings.Count(rows[0], "░")
			if filled+empty < 6 || filled != barFill(v.Share, filled+empty) {
				t.Errorf("%s: the long bar fills %d of %d cells for a share of %.3f (the panel's fill is %d)", where, filled, filled+empty, v.Share, barFill(v.Share, filled+empty))
			}
			if ledgerTierFor(w, h) == ledgerCompact {
				continue
			}
			text := strings.Join(rows, "\n")
			for _, p := range v.Pillars {
				if !p.Met && !strings.Contains(text, fmt.Sprintf("%d%%", sharePercent(p.Progress, false))) {
					t.Errorf("%s: %s's %d%% is not in the Ledger:\n%s", where, p.Name, sharePercent(p.Progress, false), text)
				}
			}
		}
	}
}

// TestLedgerCalendarDoesNotRunBack: the year on the ribbon and the share on
// the bar are held by the engine, as the Advance panel's are, so selling a
// building shows in the pillars and does not turn the town back.
func TestLedgerCalendarDoesNotRunBack(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	set := eng.Rules()
	age, _ := set.Age(set.Next("bronze_age"))
	town := map[string]int{}
	for k, n := range age.BuildingReqs {
		town[k] = n
	}
	eng.SetTownForTest(town, 0)
	eng.StepTicks(1)
	high := ledgerOf(eng.GetState(), eng).data
	if high.v.Progress < 0.5 {
		t.Fatalf("staging: progress %.2f", high.v.Progress)
	}
	for k := range town {
		town[k] = 0
	}
	eng.SetTownForTest(town, 0)
	eng.StepTicks(1)
	low := ledgerOf(eng.GetState(), eng).data
	if low.v.Progress >= high.v.Progress {
		t.Fatalf("staging: the gate did not slip (%.2f then %.2f)", high.v.Progress, low.v.Progress)
	}
	if low.Pct != high.Pct || low.v.YearText != high.v.YearText {
		t.Errorf("the bar went from %d%% to %d%% and the year from %s to %s when the town shrank", high.Pct, low.Pct, high.v.YearText, low.v.YearText)
	}
	slipped := false
	for i, p := range low.Pillars {
		slipped = slipped || p.Pct < high.Pillars[i].Pct
	}
	if !slipped {
		t.Error("no pillar shows the loss")
	}
}

// TestLedgerWaitingOnKeepsItemsWhole: what the gate waits on is listed whole
// or counted, never cut, at every width, for every real gate.
func TestLedgerWaitingOnKeepsItemsWhole(t *testing.T) {
	ages := config.Ages()
	longest := 0
	for i := 0; i+1 < len(ages); i++ {
		st := game.GameState{Age: ages[i].Key, AgeName: ages[i].Name, NextAge: ages[i+1].Key, NextAgeName: ages[i+1].Name,
			Resources: map[string]game.ResourceState{"food": {Name: "Food", Unlocked: true, Amount: 5, Rate: 1}}}
		l := ledgerOf(st, nil)
		items := l.data.Waits
		if l.data.Then != "" {
			items = append(append([]string(nil), items...), "then "+l.data.Then)
		}
		if len(items) == 0 {
			t.Errorf("%s: a town with nothing built waits on nothing", ages[i].Key)
			continue
		}
		longest = max(longest, len(items))
		lv := ledgerView{m: &l.data}
		for avail := 10; avail <= 200; avail++ {
			text, tw := lv.waitText(avail)
			if tw == 0 {
				if text != "" {
					t.Errorf("%s at %d: no width but text %q", ages[i].Key, avail, text)
				}
				continue
			}
			plain := untag(text)
			if runeLen(plain) != tw || tw > avail {
				t.Errorf("%s at %d: the text is %d cells (said %d): %q", ages[i].Key, avail, runeLen(plain), tw, plain)
				continue
			}
			body := plain[strings.Index(plain, ": ")+2:]
			parts := strings.Split(body, " · ")
			listed := len(parts)
			counted := 0
			if last := parts[len(parts)-1]; strings.HasPrefix(last, "+") && strings.HasSuffix(last, " more") {
				listed--
				fmt.Sscanf(last, "+%d more", &counted)
			}
			for j := 0; j < listed; j++ {
				if parts[j] != items[j] {
					t.Errorf("%s at %d: item %d reads %q, the gate's is %q", ages[i].Key, avail, j+1, parts[j], items[j])
				}
			}
			if listed+counted != len(items) {
				t.Errorf("%s at %d: %d listed and %d counted, of %d:\n%s", ages[i].Key, avail, listed, counted, len(items), plain)
			}
		}
	}
	if longest < 14 {
		t.Errorf("no real gate waits on 14 things: the longest has %d, so the test no longer holds the longest gate", longest)
	}
}

// TestLedgerLongestGateOnTheDashboard: the age whose gate asks for the most
// kinds of building, on the real dashboard at the two smallest terminals:
// whole rows, none cut or wrapped, with the count of what it leaves out.
func TestLedgerLongestGateOnTheDashboard(t *testing.T) {
	restoreForge(t)
	ages := config.Ages()
	age, kinds := "", 0
	for i := 0; i+1 < len(ages); i++ {
		if n := len(ages[i+1].BuildingReqs); n > kinds {
			age, kinds = ages[i].Key, n
		}
	}
	d, pages, _ := stagedDashboard(t, age)
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		rows := ledgerRowsOn(t, d, pages, sz[0], sz[1])
		if text := strings.Join(rows, "\n"); !strings.Contains(text, " more") {
			t.Errorf("%s at %dx%d: the Ledger does not say how many it leaves out:\n%s", age, sz[0], sz[1], text)
		}
	}
}

// TestLedgerShowsTheWonderBank: a wonder's bank is a row of the gate, and the
// Ledger shows the share the Wonders list shows.
func TestLedgerShowsTheWonderBank(t *testing.T) {
	eng := game.NewGameEngine()
	st := eng.GetState()
	const wonder = "sacred_grove"
	def, ok := st.Ruleset().Building(wonder)
	if !ok {
		t.Fatalf("the ruleset has no %s", wonder)
	}
	amount := def.BaseCost["food"] / 2
	eng.SetStockForTest("food", amount)
	if result := HandleCommand(fmt.Sprintf("wonder bank food %.0f", amount), eng); result.Type != "success" {
		t.Fatalf("banking food: %+v", result)
	}
	st = eng.GetState()
	pct := wonderBankPercent(st, wonder)
	if pct == 0 {
		t.Fatal("staging: the partly banked Sacred Grove reads 0%")
	}
	want := fmt.Sprintf("Sacred Grove bank %.0f%%", pct)
	l := ledgerOf(st, eng)
	if !strings.Contains(strings.Join(l.data.Waits, " · "), want) {
		t.Errorf("the Ledger waits on %v, not %q", l.data.Waits, want)
	}
	// Once the wonder is built its bank is no longer waited on.
	bank := make(map[string]float64, len(def.BaseCost))
	for resource, cost := range def.BaseCost {
		bank[resource] = cost
	}
	eng.Buildings.LoadWonderBanks(map[string]map[string]float64{wonder: bank})
	eng.StepTicks(1)
	if err := eng.BuildBuilding(wonder); err != nil {
		t.Fatalf("building Sacred Grove: %v", err)
	}
	eng.StepTicks(def.BuildTicks + 1)
	built := ledgerOf(eng.GetState(), eng)
	if strings.Contains(strings.Join(built.data.Waits, " "), "bank") || built.data.Then != "" {
		t.Errorf("a built wonder is still waited on: %v then %q", built.data.Waits, built.data.Then)
	}
}

// bottleneckTown is a game in age with its stock set: every resource the
// town has at level, except those in low, which have none.
func bottleneckTown(t *testing.T, age string, level float64, low ...string) *game.GameEngine {
	t.Helper()
	eng := game.NewGameEngine()
	for _, a := range eng.Rules().AgeKeys() {
		if eng.GetState().Age == age {
			break
		}
		if a == "primitive_age" {
			continue
		}
		must(t, eng.EnterAgeForTest(a))
	}
	for k, rs := range eng.GetState().Resources {
		if rs.Unlocked {
			eng.SetStockForTest(k, level)
		}
	}
	for _, k := range low {
		eng.SetStockForTest(k, 0)
	}
	return eng
}

// waitingPrices are the resources the gate's waiting buildings cost for
// their next copy, and what is left to bank of the wonder's cost.
func waitingPrices(st game.GameState) map[string]bool {
	costs := map[string]bool{}
	for _, p := range st.AdvanceView().Pillars {
		for _, r := range p.Rows {
			if r.Met {
				continue
			}
			switch r.Kind {
			case game.GateRowCount:
				for res, n := range buildingPrice(&st, r.Key, st.Buildings[r.Key]) {
					if n > 0 {
						costs[res] = true
					}
				}
			case game.GateRowShare:
				def, _ := st.Ruleset().Building(r.Key)
				for res, n := range def.BaseCost {
					if n-st.Buildings[r.Key].WonderBank[res] > 0 {
						costs[res] = true
					}
				}
			}
		}
	}
	return costs
}

// TestLedgerShortOfFollowsTheBlockers: the resource a waiting building costs
// and the town lacks is the one the Ledger names as short, and a resource no
// waiting building costs is not; when the gate is met nothing is named.
func TestLedgerShortOfFollowsTheBlockers(t *testing.T) {
	restoreForge(t)
	full := bottleneckTown(t, "stone_age", 1e12)
	if l := ledgerOf(full.GetState(), full); len(l.data.Need) != 0 || len(l.data.Short) != 0 {
		t.Fatalf("with every store full the gate is short of %v (%v)", l.data.Need, l.data.Short)
	}
	st := full.GetState()
	costs := waitingPrices(st)
	if len(costs) == 0 {
		t.Fatal("staging: no building the Stone Age gate waits on has a price")
	}
	for res, rs := range st.Resources {
		if !rs.Unlocked {
			continue
		}
		// One store empty, the rest full.
		e := bottleneckTown(t, "stone_age", 1e12, res)
		l := ledgerOf(e.GetState(), e)
		if got := l.data.Need[res]; got != costs[res] {
			t.Errorf("with %s empty the Ledger lists it as short = %v, but a waiting building costs it = %v", res, got, costs[res])
		}
		lines := untag(strings.Join(ledgerLines(&l.data, ledgerFull, 160, ledgerFullHeight, mapmodel.TierUnicode), "\n"))
		if !costs[res] {
			if len(l.data.Short) != 0 || strings.Contains(lines, "Short of") {
				t.Errorf("%s is empty but nothing waiting costs it, and the Ledger is short of %v:\n%s", res, l.data.Short, lines)
			}
			continue
		}
		// The name is the only one in the "Short of" line, on the accent.
		if len(l.data.Short) != 1 || l.data.Short[0] != rs.Name {
			t.Errorf("%s is the only store empty and a waiting building costs it, but the Ledger is short of %v", res, l.data.Short)
		}
		if !strings.Contains(lines, "Short of: "+rs.Name) {
			t.Errorf("the Ledger does not name %s as short:\n%s", rs.Name, lines)
		}
		if chip := theme.TagFgBg(theme.RoleOnAccent, theme.RoleAccent) + rs.Name; !strings.Contains(strings.Join(ledgerLines(&l.data, ledgerFull, 160, ledgerFullHeight, mapmodel.TierUnicode), "\n"), chip) {
			t.Errorf("the name of %s is not on the accent", res)
		}
	}
	// The gate met: nothing waited on, nothing marked, and the line says so.
	meet := bottleneckTown(t, "stone_age", 0)
	meetTheGate(t, meet)
	l := ledgerOf(meet.GetState(), meet)
	if len(l.data.Waits) != 0 || l.data.Then != "" || len(l.data.Need) != 0 || len(l.data.Short) != 0 {
		t.Errorf("a town that meets the gate still waits on %v then %q, short of %v", l.data.Waits, l.data.Then, l.data.Need)
	}
	lv := ledgerView{m: &l.data, w: 160}
	if text, _ := lv.waitText(150); !strings.Contains(untag(text), "Every pillar stands") {
		t.Errorf("a met gate says %q", untag(text))
	}
}

// TestLedgerRowsFitAtEveryWidth: no row is wider than its Ledger at any
// width, in either glyph tier, for a town and a final age alike.
func TestLedgerRowsFitAtEveryWidth(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "quantum_age")
	l := ledgerOf(eng.GetState(), eng)
	final := &ledger{}
	final.data = l.data
	final.data.v.Final = true
	for _, m := range []*ledgerData{&l.data, &final.data} {
		for _, g := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			for w := 30; w <= 260; w++ {
				for _, c := range []struct {
					tier ledgerTier
					h    int
				}{{ledgerCompact, ledgerCompactRows}, {ledgerStandard, ledgerStandardRows}, {ledgerFull, ledgerFullHeight}} {
					lines := ledgerLines(m, c.tier, w, c.h, g)
					if len(lines) > c.h {
						t.Fatalf("tier %d at %d: %d rows for %d", c.tier, w, len(lines), c.h)
					}
					for _, line := range lines {
						if n := visibleLen(line); n > w {
							t.Fatalf("tier %d, glyphs %v at width %d: a row of %d cells: %q", c.tier, g, w, n, untag(line))
						}
					}
				}
			}
		}
	}
}

// TestLedgerTierBoundaries: the Ledger a terminal gets.
func TestLedgerTierBoundaries(t *testing.T) {
	for _, c := range []struct {
		w, h int
		want ledgerTier
	}{
		{80, 24, ledgerCompact}, {99, 40, ledgerCompact}, {100, 29, ledgerCompact}, {100, 30, ledgerStandard},
		{119, 60, ledgerStandard}, {120, 35, ledgerStandard}, {120, 36, ledgerStandard}, {120, 38, ledgerStandard},
		{120, 39, ledgerFull}, {160, 45, ledgerFull},
	} {
		if got := ledgerTierFor(c.w, c.h); got != c.want {
			t.Errorf("%dx%d is tier %d, want %d", c.w, c.h, got, c.want)
		}
	}
}

// TestLedgerResizeShowsTheRightLedger: a resize shows the Ledger of the new
// size at the next draw, with no refresh between.
func TestLedgerResizeShowsTheRightLedger(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	drawnDashboard(t, d, pages, 80, 24)
	sim := newSimScreen(t, 160, 45)
	screen := theme.WrapScreen(sim)
	for _, sz := range []struct{ w, h int }{{160, 45}, {80, 24}, {100, 30}, {160, 45}} {
		sim.SetSize(sz.w, sz.h)
		pages.SetRect(0, 0, sz.w, sz.h)
		pages.Draw(screen) // no d.refresh in between
		if _, _, _, lh := d.ledgerTV.GetRect(); lh != ledgerHeight(sz.w, sz.h) {
			t.Errorf("%dx%d: the Ledger is %d rows after the resize, want %d", sz.w, sz.h, lh, ledgerHeight(sz.w, sz.h))
		}
	}
}

// GetText is the rows as text, one to a line, with the colour tags left in
// unless stripTags.
func (b *ledgerBox) GetText(stripTags bool) string {
	lines := make([]string, len(b.lines))
	for i, l := range b.lines {
		if stripTags {
			l = stripLedgerTags(l)
		}
		lines[i] = l
	}
	return strings.Join(lines, "\n")
}

// stripLedgerTags is s without the tags the Ledger writes.
func stripLedgerTags(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '[' {
			if n, _, _, _, _, ok := ledgerTag(s[i:]); ok {
				i += n
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		sb.WriteRune(r)
		i += size
	}
	return sb.String()
}
