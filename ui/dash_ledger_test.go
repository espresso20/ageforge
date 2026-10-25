package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// The Ledger is the top of the dashboard under the status ribbon: the long
// bar to the next age, the four pillars and a card a resource. These tests
// draw the real dashboard at the three sizes it is written for, hold its
// numbers to the Advance panel's, and hold its memory, its cards and its
// highlight to the rules in dash_ledger.go.

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
	s := resolveMapSettings(nil, d.mapViews.reg)
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
	for _, c := range []struct{ w, h, slots, want int }{
		{80, 24, 26, 2}, {99, 60, 26, 2}, {100, 29, 26, 2}, // the compact Ledger: the strip's two rows
		{100, 30, 26, 4}, {119, 60, 3, 4}, {120, 36, 26, 4}, {120, 38, 26, 4}, // the standard one: bar, pillars, a card
		{120, 39, 26, 7}, {120, 40, 26, 7}, {120, 41, 26, 9}, {120, 43, 26, 11}, // the full one: five rows and the cards
		{160, 45, 26, 11}, {160, 45, 6, 9}, {160, 45, 3, 7}, {240, 60, 3, 7}, {240, 60, 26, 11},
	} {
		if got := ledgerHeight(c.w, c.h, c.slots); got != c.want {
			t.Errorf("%dx%d with %d resources: %d rows, want %d", c.w, c.h, c.slots, got, c.want)
		}
	}
	// The compact Ledger takes the rows the Next Age strip had, so a small
	// terminal loses nothing to it.
	if dashHeaderRows != 1+1+1+ledgerCompactRows {
		t.Errorf("dashHeaderRows = %d", dashHeaderRows)
	}
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	for _, sz := range append(layoutSizes, ledgerSizes...) {
		w, h := sz[0], sz[1]
		drawnDashboard(t, d, pages, w, h)
		x, y, lw, lh := d.ledgerTV.GetRect()
		if want := ledgerHeight(w, h, d.ledger.slots); x != 0 || y != 3 || lw != w || lh != want {
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
				// The long bar and one row of cards: no pillars.
				if len(rows) != 2 || strings.Contains(text, "Knowledge") {
					t.Errorf("%s: the compact Ledger is not the bar and a row of cards:\n%s", where, text)
				}
				if !strings.Contains(rows[1], "/t") && !strings.ContainsAny(rows[1], "▲▼▬") {
					t.Errorf("%s: the card row has no rates:\n%s", where, text)
				}
			case ledgerStandard:
				if !strings.Contains(rows[1], "Knowledge") || !strings.Contains(rows[1], "Monument") || len(rows) != 4 {
					t.Errorf("%s: the standard Ledger lacks its row of pillars:\n%s", where, text)
				}
				if !strings.ContainsAny(rows[3], "▁▂▃▄▅▆▇█") {
					t.Errorf("%s: the standard Ledger has no trend line in its card row:\n%s", where, text)
				}
			case ledgerFull:
				if pillars != 4 || !strings.Contains(rows[1], "Waiting on: ") || !strings.Contains(rows[2], "KNOWLEDGE") {
					t.Errorf("%s: the full Ledger lacks its waiting line or its pillars:\n%s", where, text)
				}
				if !strings.Contains(text, "rising") && !strings.Contains(text, "falling") && !strings.Contains(text, "steady") {
					t.Errorf("%s: the cards do not say which way they are going:\n%s", where, text)
				}
			}
		}
	}
}

// TestLedgerInPlainASCII: on the plain glyph tier the Ledger is nothing but
// ASCII at every size, its bars and trend lines in plain marks.
func TestLedgerInPlainASCII(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"bronze_age", "quantum_age"} {
		d, pages, _ := stagedDashboard(t, age)
		plainTier(d)
		// Some history, so the trend lines have something to draw.
		drawnDashboard(t, d, pages, 80, 24)
		for _, r := range d.ledger.trail.rings {
			for n := 0; n < 10; n++ {
				r.push(float64(10 + n*n%7))
			}
		}
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
	// The trend line has eight heights in either tier.
	if got := ledgerSpark([]float64{1, 2, 3, 4, 5, 6, 7, 8}, 8, true); got != ledgerSparkASCII {
		t.Errorf("the plain trend line is %q, want %q", got, ledgerSparkASCII)
	}
	if got := ledgerSpark([]float64{1, 2, 3, 4, 5, 6, 7, 8}, 8, false); got != ledgerSparkUnicode {
		t.Errorf("the trend line is %q, want %q", got, ledgerSparkUnicode)
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

// TestLedgerTrailIsBounded: the memory of the amounts is a few samples a
// resource, taken every few ticks, and starts over when the game does.
func TestLedgerTrailIsBounded(t *testing.T) {
	st := game.GameState{Resources: map[string]game.ResourceState{
		"food": {Unlocked: true, Amount: 1, Rate: 1}, "wood": {Unlocked: true, Amount: 2}, "gold": {Unlocked: false, Amount: 9},
	}}
	var tr ledgerTrail
	for tick := 0; tick <= 100000; tick++ {
		st.Tick = tick
		food := st.Resources["food"]
		food.Amount = float64(tick)
		st.Resources["food"] = food
		tr.observe(st)
	}
	if len(tr.rings) != 2 || tr.rings["gold"] != nil {
		t.Errorf("the trail keeps %d resources, want the two unlocked", len(tr.rings))
	}
	r := tr.rings["food"]
	if r.n != ledgerTrailLen {
		t.Errorf("the trail holds %d samples, want its limit %d", r.n, ledgerTrailLen)
	}
	vals := r.appendTo(nil)
	if len(vals) != ledgerTrailLen {
		t.Fatalf("the trail reads back %d samples", len(vals))
	}
	// Oldest first, a sample every ledgerSampleTicks ticks, ending at the last.
	for i := 1; i < len(vals); i++ {
		if vals[i]-vals[i-1] != ledgerSampleTicks {
			t.Errorf("samples %d and %d are %v and %v: not %d ticks apart", i-1, i, vals[i-1], vals[i], ledgerSampleTicks)
			break
		}
	}
	if vals[len(vals)-1] != 100000 {
		t.Errorf("the newest sample is %v, want 100000", vals[len(vals)-1])
	}
	// Fewer ticks than a sample apart take none.
	n := tr.rings["wood"].n
	st.Tick = 100001
	tr.observe(st)
	if tr.rings["wood"].n != n {
		t.Error("a sample was taken a tick after the last")
	}
	// The tick going back (a new game) starts the memory over; so does reset.
	st.Tick = 3
	tr.observe(st)
	if tr.rings["food"].n != 1 {
		t.Errorf("after the tick went back the trail holds %d samples, want 1", tr.rings["food"].n)
	}
	tr.reset()
	if len(tr.rings) != 0 {
		t.Error("reset left a trail")
	}
}

// TestLedgerTrendAndSpark: which way an amount has gone, and the line that
// shows it.
func TestLedgerTrendAndSpark(t *testing.T) {
	for _, c := range []struct {
		trail []float64
		rate  float64
		want  int
	}{
		{[]float64{1, 2, 3, 4}, 0, 1}, {[]float64{4, 3, 2, 1}, 5, -1}, {[]float64{100, 100.5, 100.2}, 5, 0},
		{[]float64{0, 0, 0}, 5, 0}, {nil, 2, 1}, {[]float64{7}, -2, -1}, {nil, 0, 0},
		{[]float64{1, 2, 3, 4}, -1, -1}, // up over the minutes, down now
	} {
		if got := ledgerTrend(c.trail, c.rate); got != c.want {
			t.Errorf("trend of %v at rate %v = %d, want %d", c.trail, c.rate, got, c.want)
		}
	}
	// The newest sample is at the right, the cells before the first are blank
	// and a flat line sits in the middle.
	if got := ledgerSpark([]float64{1, 9}, 5, false); got != "   ▁█" {
		t.Errorf("a two-sample line = %q", got)
	}
	if got := ledgerSpark([]float64{5, 5, 5}, 3, false); got != "▄▄▄" {
		t.Errorf("a flat line = %q", got)
	}
	if got := ledgerSpark(make([]float64, 40), 12, false); runeLen(got) != 12 {
		t.Errorf("a long trail draws %d cells, want 12", runeLen(got))
	}
}

// TestLedgerHidesIdleResources: a resource with nothing held, no rate, or not
// yet unlocked takes no card, unless the gate is short of it.
func TestLedgerHidesIdleResources(t *testing.T) {
	st := game.GameState{Age: "bronze_age", Resources: map[string]game.ResourceState{
		"food":  {Name: "Food", Unlocked: true, Amount: 50, Rate: 1},
		"coal":  {Name: "Coal", Unlocked: true, Amount: 0, Rate: 0},   // nothing going on
		"stone": {Name: "Stone", Unlocked: true, Amount: 0, Rate: 2},  // arriving
		"iron":  {Name: "Iron", Unlocked: false, Amount: 30, Rate: 3}, // not unlocked
		"gold":  {Name: "Gold", Unlocked: true, Amount: 7, Rate: 0},   // held, idle
		"wood":  {Name: "Wood", Unlocked: true, Amount: 0, Rate: 0},   // nothing going on, but the gate costs it
	}}
	l := ledgerOf(st, nil)
	if !l.data.Need["wood"] || l.data.Need["coal"] {
		t.Fatalf("staging: the gate is short of %v, want wood and not coal", l.data.Need)
	}
	var got []string
	for _, c := range l.data.Cards {
		got = append(got, c.Key)
	}
	// In the ruleset's order, with only the ones that have something going on
	// and the one the gate is short of.
	if strings.Join(got, " ") != "food wood stone gold" {
		t.Errorf("cards for %v, want food wood stone gold", got)
	}
	for _, c := range []struct {
		tier  ledgerTier
		first int // the first row of cards
	}{{ledgerCompact, 1}, {ledgerStandard, 2}, {ledgerFull, 5}} {
		lines := ledgerLines(&l.data, c.tier, 160, 11, mapmodel.TierUnicode)
		text := untag(strings.Join(lines[c.first:], "\n"))
		for _, hidden := range []string{"Coal", "Iron"} {
			if strings.Contains(text, hidden) {
				t.Errorf("tier %d draws a card for %q:\n%s", c.tier, hidden, text)
			}
		}
		wantAll(t, fmt.Sprintf("tier %d cards", c.tier), text, "Food", "Wood", "Stone", "Gold")
	}
}

// TestLedgerCardsStayPut: a card's place does not change with its numbers,
// however many digits they have, and a card does not push the next.
func TestLedgerCardsStayPut(t *testing.T) {
	names := []string{"Food", "Wood", "Stone", "Iron", "Gold", "Faith"}
	mk := func(scale float64) *ledgerData {
		d := &ledgerData{Need: map[string]bool{}}
		for i, name := range names {
			d.Cards = append(d.Cards, ledgerCard{Key: strings.ToLower(name), Name: name, Amount: scale * float64(i+1) * 1.37, Rate: scale * float64(i+2) / 3,
				Trail: []float64{1, 3, 2, 5, 4}, Trend: i%3 - 1})
		}
		return d
	}
	for _, tier := range []ledgerTier{ledgerCompact, ledgerStandard, ledgerFull} {
		const w = 160
		var cols [][]int
		for _, scale := range []float64{0.01, 1, 1234, 9.9e9, 4e18} {
			var at []int
			for _, line := range ledgerLines(mk(scale), tier, w, 11, mapmodel.TierUnicode) {
				plain := untag(line)
				if runeLen(plain) > w {
					t.Errorf("tier %d, scale %v: a row of %d cells", tier, scale, runeLen(plain))
				}
				for _, name := range names {
					if i := strings.Index(plain, name); i >= 0 {
						at = append(at, runeLen(plain[:i]))
					}
				}
			}
			cols = append(cols, at)
		}
		for i := 1; i < len(cols); i++ {
			if fmt.Sprint(cols[i]) != fmt.Sprint(cols[0]) || len(cols[0]) == 0 {
				t.Errorf("tier %d: the cards moved with their numbers: %v then %v", tier, cols[0], cols[i])
			}
		}
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

// TestLedgerBottleneckFollowsTheBlockers: the resource a waiting building
// costs and the town lacks is the one the Ledger marks, and a resource no
// waiting building costs is not; when the gate is met nothing is marked.
func TestLedgerBottleneckFollowsTheBlockers(t *testing.T) {
	restoreForge(t)
	full := bottleneckTown(t, "stone_age", 1e12)
	if l := ledgerOf(full.GetState(), full); len(l.data.Need) != 0 || l.data.Worst != "" {
		t.Fatalf("with every store full the gate is short of %v (worst %q)", l.data.Need, l.data.Worst)
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
			t.Errorf("with %s empty the Ledger marks it = %v, but a waiting building costs it = %v", res, got, costs[res])
		}
		if costs[res] && l.data.Worst != res {
			t.Errorf("%s is the only store empty and a waiting building costs it, but the worst is %q", res, l.data.Worst)
		}
		if !costs[res] && l.data.Worst != "" {
			t.Errorf("%s is empty but nothing waiting costs it, and the worst is %q", res, l.data.Worst)
		}
		if !costs[res] {
			continue
		}
		// The marks on the cards: a chip with the arrow on the worst, and the
		// name in the "Short of" line.
		marked := strings.Join(ledgerLines(&l.data, ledgerFull, 160, 11, mapmodel.TierUnicode), "\n")
		if chip := theme.TagFgBg(theme.RoleOnAccent, theme.RoleAccent) + "▶" + rs.Name; !strings.Contains(marked, chip) {
			t.Errorf("no card carries the mark for %s:\n%s", res, untag(marked))
		}
		if !strings.Contains(untag(marked), "Short of: "+rs.Name) {
			t.Errorf("the Ledger does not name %s as short:\n%s", rs.Name, untag(marked))
		}
	}
	// The gate met: nothing waited on, nothing marked, and the line says so.
	meet := bottleneckTown(t, "stone_age", 0)
	meetTheGate(t, meet)
	l := ledgerOf(meet.GetState(), meet)
	if len(l.data.Waits) != 0 || l.data.Then != "" || len(l.data.Need) != 0 || l.data.Worst != "" {
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
				}{{ledgerCompact, 2}, {ledgerStandard, 4}, {ledgerFull, 7}, {ledgerFull, 11}} {
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
		if _, _, _, lh := d.ledgerTV.GetRect(); lh != ledgerHeight(sz.w, sz.h, d.ledger.slots) {
			t.Errorf("%dx%d: the Ledger is %d rows after the resize, want %d", sz.w, sz.h, lh, ledgerHeight(sz.w, sz.h, d.ledger.slots))
		}
	}
}
