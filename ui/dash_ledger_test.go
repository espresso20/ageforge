package ui

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// The Ledger is what the header's way to the next age reads: the long bar,
// what it waits on and the four pillars. These tests hold its numbers to the
// Advance panel's, hold what it waits on and what it is short of to the rules
// in dash_ledger.go, and draw it in the real header at the sizes it is
// written for (dash_header_test.go holds the header's own layout).

// ledgerOf works the Ledger out for a snapshot, as a refresh does.
func ledgerOf(st game.GameState, eng *game.GameEngine) *ledger {
	l := &ledger{}
	l.update(st, eng)
	return l
}

// ledgerSizes are the three terminals the header is written for: the
// smallest (two rows inside), the common one (three) and a big one.
var ledgerSizes = [][2]int{{80, 24}, {100, 30}, {160, 45}}

// plainTier puts the dashboard on the plain glyph tier, for the session.
func plainTier(d *Dashboard) {
	s := resolveMapSettings(nil)
	s.Tier = mapmodel.TierASCII
	d.mapLocal = &s
}

// headerRowsOn draws the dashboard and returns what the header wrote, row by
// row, held to its box.
func headerRowsOn(t *testing.T, d *Dashboard, pages *nativePages, w, h int) []string {
	t.Helper()
	cells := drawnDashboard(t, d, pages, w, h)
	return checkHeader(t, fmt.Sprintf("header at %dx%d", w, h), d, cells, w)
}

// checkHeader holds the header to what it wrote: the box is the top of the
// screen, as wide as it, and the rows on screen are the rows of its grid.
func checkHeader(t *testing.T, where string, d *Dashboard, cells []simCell, screenW int) []string {
	t.Helper()
	x, y, w, h := d.headerTV.GetRect()
	g := d.headerTV.grid
	if g == nil || w <= 0 || h <= 0 {
		t.Fatalf("%s: the header has no grid (%dx%d)", where, w, h)
	}
	if x != 0 || y != 0 || w != screenW || g.w != w || g.h != h {
		t.Errorf("%s: the header is at (%d,%d) %dx%d with a grid of %dx%d, want the top row, %d wide", where, x, y, w, h, g.w, g.h, screenW)
	}
	drawn := rectText(cells, screenW, x, y, w, h)
	for row, want := range g.Lines() {
		if row >= len(drawn) {
			break
		}
		if drawn[row] != want {
			t.Errorf("%s: row %d shows %q, the header wrote %q", where, row+1, drawn[row], want)
		}
	}
	return drawn
}

// GetText is the header's rows as text, one to a line.
func (b *headerBox) GetText(bool) string {
	if b.grid == nil {
		return ""
	}
	return strings.Join(b.grid.Lines(), "\n")
}

// TestLedgerAtThreeSizes draws the real dashboard at 80x24, 100x30 and
// 160x45 and holds each header to what it is for.
func TestLedgerAtThreeSizes(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"bronze_age", "medieval_age", "quantum_age"} {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		next := st.NextAgeName
		for _, sz := range ledgerSizes {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			rows := headerRowsOn(t, d, pages, w, h)
			text := strings.Join(rows, "\n")
			// Every size: the long bar to the next age, with its percent, and
			// what it waits on.
			if !strings.Contains(rows[2], "To the "+next) || !strings.Contains(rows[2], "%") || !strings.ContainsAny(rows[2], "█░") {
				t.Errorf("%s: no long bar to the %s:\n%s", where, next, text)
			}
			if !strings.Contains(text, "Waiting on: ") {
				t.Errorf("%s: the header does not say what it waits on:\n%s", where, text)
			}
			pillars := 0
			for _, name := range []string{"Knowledge", "People", "Industry", "Monument"} {
				if strings.Contains(strings.Join(rows[2:], "\n"), name) {
					pillars++
				}
			}
			if headerHeight(w, h) == headerSmallRows {
				// Two rows inside: the bar with what it waits on, and no pillars.
				if len(rows) != 4 || strings.Contains(text, "Knowledge") || !strings.Contains(rows[2], "Waiting on: ") {
					t.Errorf("%s: the small header is not the bar with what it waits on:\n%s", where, text)
				}
			} else if pillars != 4 || len(rows) != 5 || !strings.Contains(rows[3], "Waiting on: ") {
				t.Errorf("%s: the header lacks its waiting line or its pillars:\n%s", where, text)
			}
			// The resources are the Resources panel's: the header has no cards,
			// no trend lines and no count of resources it leaves out.
			if strings.ContainsAny(text, "▲▼▬▶▁▂▃▄▅▆▇") || strings.Contains(text, "rising") || strings.Contains(text, "falling") {
				t.Errorf("%s: the header draws a resource card:\n%s", where, text)
			}
		}
	}
}

// TestLedgerInPlainASCII: on the plain glyph tier the header is nothing but
// ASCII at every size, its bars in plain marks.
func TestLedgerInPlainASCII(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"bronze_age", "quantum_age"} {
		d, pages, _ := stagedDashboard(t, age)
		plainTier(d)
		for _, sz := range ledgerSizes {
			w, h := sz[0], sz[1]
			rows := headerRowsOn(t, d, pages, w, h)
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

// TestLedgerReadsTheAdvancePanel: the long bar, the pillars' counts and the
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
			rows := headerRowsOn(t, d, pages, w, h)
			if !strings.Contains(rows[2], fmt.Sprintf("%d%%", pct)) {
				t.Errorf("%s: the long bar's row has no %d%%: %q", where, pct, rows[2])
			}
			if !strings.Contains(rows[1], "Year "+v.YearText) {
				t.Errorf("%s: the first row does not show the year %s: %q", where, v.YearText, rows[1])
			}
			bar := rows[2][:strings.Index(rows[2], "%")]
			filled, empty := strings.Count(bar, "█"), strings.Count(bar, "░")
			if filled+empty < 6 || filled != barFill(v.Share, filled+empty) {
				t.Errorf("%s: the long bar fills %d of %d cells for a share of %.3f (the panel's fill is %d)", where, filled, filled+empty, v.Share, barFill(v.Share, filled+empty))
			}
			if headerHeight(w, h) == headerSmallRows {
				continue
			}
			for _, p := range l.data.Pillars {
				if !p.Met && !strings.Contains(rows[2], p.Name+" ") {
					t.Errorf("%s: pillar %s is not on the bar's row:\n%s", where, p.Name, rows[2])
				}
				if !p.Met && !strings.Contains(rows[2], p.Count) {
					t.Errorf("%s: %s's count %s is not on the bar's row:\n%s", where, p.Name, p.Count, rows[2])
				}
			}
		}
	}
}

// TestLedgerCalendarDoesNotRunBack: the year in the header and the share on
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
		if len(items) == 0 {
			t.Errorf("%s: a town with nothing built waits on nothing", ages[i].Key)
			continue
		}
		longest = max(longest, len(items)+min(len(l.data.Then), 1))
		for avail := 24; avail <= 200; avail++ {
			var plain strings.Builder
			for _, p := range l.data.waitText(avail) {
				plain.WriteString(p.s)
			}
			text := plain.String()
			if runeLen(text) > avail {
				t.Errorf("%s at %d: the text is %d cells: %q", ages[i].Key, avail, runeLen(text), text)
				continue
			}
			parts := strings.Split(text[strings.Index(text, ": ")+2:], " · ")
			then := false
			if last := parts[len(parts)-1]; strings.HasPrefix(last, "then ") {
				then = true
				parts = parts[:len(parts)-1]
			}
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
				t.Errorf("%s at %d: %d listed and %d counted, of %d:\n%s", ages[i].Key, avail, listed, counted, len(items), text)
			}
			if then && (counted > 0 || l.data.Then == "") {
				t.Errorf("%s at %d: the wonder comes after the list when it is cut or when there is none: %q", ages[i].Key, avail, text)
			}
		}
	}
	if longest < 14 {
		t.Errorf("no real gate waits on 14 things: the longest has %d, so the test no longer holds the longest gate", longest)
	}
}

// TestLedgerLongestGateOnTheDashboard: the age whose gate asks for the most
// kinds of building, on the real dashboard at the two smallest terminals:
// whole rows, with the count of what it leaves out.
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
		rows := headerRowsOn(t, d, pages, sz[0], sz[1])
		if text := strings.Join(rows, "\n"); !strings.Contains(text, " more") {
			t.Errorf("%s at %dx%d: the header does not say how many it leaves out:\n%s", age, sz[0], sz[1], text)
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
			t.Errorf("with %s empty the header lists it as short = %v, but a waiting building costs it = %v", res, got, costs[res])
		}
		g := renderHeader(headerView{Led: &l.data}, 160, 3)
		lines := strings.Join(g.Lines(), "\n")
		if !costs[res] {
			if len(l.data.Short) != 0 || strings.Contains(lines, "Short of") {
				t.Errorf("%s is empty but nothing waiting costs it, and the Ledger is short of %v:\n%s", res, l.data.Short, lines)
			}
			continue
		}
		// The name is the only one in the "Short of" line, as a chip.
		if len(l.data.Short) != 1 || l.data.Short[0] != rs.Name {
			t.Errorf("%s is the only store empty and a waiting building costs it, but the Ledger is short of %v", res, l.data.Short)
		}
		row := g.Lines()[3]
		at := strings.LastIndex(row, " "+rs.Name+" ")
		if !strings.Contains(row, "Short of:") || at < 0 {
			t.Errorf("the Ledger does not name %s as short:\n%s", rs.Name, lines)
		} else if k := g.c[3*g.w+utf8.RuneCountInString(row[:at])+1].k; k != xChipMil {
			t.Errorf("the name of %s is drawn as kind %d, want a chip", res, k)
		}
	}
	// The gate met: nothing waited on, nothing marked, and the line says so.
	meet := bottleneckTown(t, "stone_age", 0)
	meetTheGate(t, meet)
	l := ledgerOf(meet.GetState(), meet)
	if len(l.data.Waits) != 0 || l.data.Then != "" || len(l.data.Need) != 0 || len(l.data.Short) != 0 {
		t.Errorf("a town that meets the gate still waits on %v then %q, short of %v", l.data.Waits, l.data.Then, l.data.Need)
	}
	if _, text := l.data.readyText(); !strings.Contains(text, "Every pillar stands") {
		t.Errorf("a met gate says %q", text)
	}
	if lines := renderHeader(headerView{Led: &l.data}, 160, 3).Lines(); !strings.Contains(lines[3], "Every pillar stands") || strings.Contains(strings.Join(lines, ""), "Short of") {
		t.Errorf("a met gate reads %q", lines)
	}
}

// TestLedgerRowsFitAtEveryWidth: no row spills out of the header at any
// width, in either glyph tier, for a town and a final age alike: the frame's
// edges hold, the air beside the text stays empty, and the plain tier is
// ASCII.
func TestLedgerRowsFitAtEveryWidth(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "quantum_age")
	l := ledgerOf(eng.GetState(), eng)
	final := &ledger{}
	final.data = l.data
	final.data.v.Final = true
	v := statusLine(eng.GetState(), l.data.v.YearText)
	v.Notice = "★ Badge earned: Stonecutter"
	v.Alerts = append(v.Alerts, headerAlert{"⚑ Harbinger: type harbinger", "⚑ Harbinger", "harbinger", xChipLegend})
	for _, m := range []*ledgerData{&l.data, &final.data} {
		v.Led = m
		for _, glyphs := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			for w := 30; w <= 260; w++ {
				for _, rows := range []int{2, 3} {
					g := renderHeader(v, w, rows)
					if glyphs == mapmodel.TierASCII {
						g.fold()
					}
					lines := strings.Split(strings.TrimRight(g.String(), "\n"), "\n")
					if len(lines) != rows+2 {
						t.Fatalf("width %d: %d rows for %d inside", w, len(lines), rows)
					}
					for y := 1; y < len(lines)-1; y++ {
						r := []rune(lines[y])
						if len(r) != w {
							t.Fatalf("width %d, %d rows: row %d is %d cells", w, rows, y, len(r))
						}
						for _, c := range lines[y] {
							if glyphs == mapmodel.TierASCII && c > 0x7e {
								t.Fatalf("width %d: the plain tier has a mark that is not ASCII: %q", w, lines[y])
							}
						}
						if glyphs == mapmodel.TierASCII && strings.ContainsRune(lines[y], '?') {
							t.Fatalf("width %d: a mark has no plain form: %q", w, lines[y])
						}
						if r[1] != ' ' || r[w-2] != ' ' {
							t.Fatalf("width %d, %d rows, glyphs %v: row %d runs into the frame: %q", w, rows, glyphs, y, lines[y])
						}
					}
				}
			}
		}
	}
}

// TestLedgerResizeShowsTheRightLedger: a resize shows the header of the new
// size at the next draw, with no refresh between.
func TestLedgerResizeShowsTheRightLedger(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	drawnDashboard(t, d, pages, 80, 24)
	sim := newSimScreen(t, 160, 45)
	screen := sim
	for _, sz := range []struct{ w, h int }{{160, 45}, {80, 24}, {100, 30}, {160, 45}} {
		sim.SetSize(sz.w, sz.h)
		pages.SetRect(0, 0, sz.w, sz.h)
		drawOn(screen, pages) // no d.refresh in between
		if _, _, w, h := d.headerTV.GetRect(); h != headerHeight(sz.w, sz.h) || w != sz.w || d.headerTV.grid.h != h || d.headerTV.grid.w != w {
			t.Errorf("%dx%d: the header is %dx%d after the resize (grid %dx%d), want %d rows", sz.w, sz.h, w, h, d.headerTV.grid.w, d.headerTV.grid.h, headerHeight(sz.w, sz.h))
		}
	}
}
