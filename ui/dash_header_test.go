package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/window/kit"
)

// The header is the framed box at the top of the dashboard (dash_header.go).
// These tests hold its height and place, what row A gives up and in which
// order, that an alert is never dropped, and that every part takes a click.

// headerFixture is a view with every part of the header filled, for a
// width and a gate: the long names are the worst case for the first row.
func headerFixture(t *testing.T) headerView {
	t.Helper()
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "quantum_age")
	st := eng.GetState()
	st.Prestige.Level = 12
	st.Workers = game.WorkerState{TotalPop: 8311, MaxPop: 51785785}
	st.Morale, st.MoraleMultiplier = 1.1, 1.2
	prev := game.DevModeActive
	game.DevModeActive = true
	t.Cleanup(func() { game.DevModeActive = prev })
	st.PendingCatastrophe = "iron_era"
	st.Harbinger = &game.HarbingerView{}
	st.AccountStats = &game.AccountStatsView{DisplayName: "A Very Long Account Name Indeed", BadgeSummary: game.BadgeSummary{Title: "Magistrate", TitleRank: 2}}
	st.Milestones.CurrentTitle = "Aspiring"
	l := ledgerOf(st, eng)
	v := statusLine(st, l.data.v.YearText)
	v.Led = &l.data
	return v
}

// TestHeaderHeights: the rows the header takes for each size, and the
// dashboard gives it exactly those, at the top, with the boxes right under it.
func TestHeaderHeights(t *testing.T) {
	for _, c := range []struct{ w, h, want int }{
		{80, 24, 4}, {99, 60, 4}, {100, 29, 4}, // two rows inside
		{100, 30, 5}, {119, 60, 5}, {120, 39, 5}, {160, 45, 5}, {240, 60, 5}, // three
	} {
		if got := headerHeight(c.w, c.h); got != c.want {
			t.Errorf("%dx%d: %d rows, want %d", c.w, c.h, got, c.want)
		}
	}
	if dashHeaderRows != headerSmallRows {
		t.Errorf("dashHeaderRows = %d", dashHeaderRows)
	}
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	for _, sz := range append(layoutSizes, ledgerSizes...) {
		w, h := sz[0], sz[1]
		rows := headerRowsOn(t, d, pages, w, h)
		x, y, hw, hh := d.headerTV.GetRect()
		if want := headerHeight(w, h); x != 0 || y != 0 || hw != w || hh != want || len(rows) != want {
			t.Errorf("%dx%d: the header is at (%d,%d) %dx%d with %d rows, want (0,0) %dx%d", w, h, x, y, hw, hh, len(rows), w, want)
		}
		if _, cy, _, _ := d.contentArea.GetRect(); cy != hh {
			t.Errorf("%dx%d: the boxes start at row %d, the header ends at %d", w, h, cy, hh)
		}
		for i, row := range rows {
			if strings.TrimSpace(row) == "" {
				t.Errorf("%dx%d: row %d of the header is blank", w, h, i+1)
			}
		}
	}
}

// TestHeaderAlertsAreNeverDropped: at every width from 60 to 260, with the
// longest names, every alert is on the first row, at least as its mark, and
// the ones with a command are click targets.
func TestHeaderAlertsAreNeverDropped(t *testing.T) {
	v := headerFixture(t)
	v.Alerts = append([]headerAlert{{text: "You cheated, shame", short: "You cheated…", k: xBad}}, v.Alerts...)
	if len(v.Alerts) != 3 {
		t.Fatalf("staging: %d alerts", len(v.Alerts))
	}
	for w := 60; w <= 260; w++ {
		for _, rows := range []int{2, 3} {
			g := renderHeader(v, w, rows)
			row := g.Lines()[1]
			for _, mark := range []string{"☄", "⚑", "Y"} {
				if !strings.Contains(row, mark) {
					t.Fatalf("width %d, %d rows: no %q on the first row: %q", w, rows, mark, row)
				}
			}
			for i, a := range v.Alerts {
				found := false
				for _, h := range g.hits {
					found = found || (h.id == hitHeader && h.n == headerAlert0+i)
				}
				if found != (a.cmd != "") {
					t.Fatalf("width %d: alert %d has a click target: %v", w, i, found)
				}
			}
		}
	}
}

// TestHeaderRowAGivesWayInOrder: the first row keeps the age, People, Morale
// and the alerts at any width, and gives up the rest in order: what morale
// does to production, the alerts' long words, the year, People's long
// numbers, Prestige and DEV, and last the alerts' words.
func TestHeaderRowAGivesWayInOrder(t *testing.T) {
	v := headerFixture(t)
	v.Name, v.Notice = "Ada", ""
	rowA := func(w int) string { return renderHeader(v, w, 2).Lines()[1] }
	least := func(piece string) int {
		for w := 40; w <= 300; w++ {
			if strings.Contains(rowA(w), piece) {
				return w
			}
		}
		return 1 << 20
	}
	people := v.People.Text
	order := []string{"(production +20%)", "type catastrophe", "Year " + v.Year[len("Year "):], people, "Prestige 12", "☄ Catastrophe"}
	for i := 1; i < len(order); i++ {
		if a, b := least(order[i-1]), least(order[i]); a < b {
			t.Errorf("%q stays on a narrower row (%d) than %q (%d), but is given up first", order[i-1], a, order[i], b)
		}
	}
	// People is written short before Prestige and DEV go.
	if short := v.PeopleShort.Text; least(short) >= 1<<20 || least(short) > least(people) || strings.Contains(short, ",") {
		t.Errorf("People never shortens: %q", short)
	}
	for w := 70; w <= 220; w++ {
		line := rowA(w)
		for _, want := range []string{v.Age, "Morale 110%", "☄", "⚑"} {
			if !strings.Contains(line, want) {
				t.Errorf("width %d: no %q on the first row: %q", w, want, line)
			}
		}
		if w >= 100 && !strings.Contains(line, "People") {
			t.Errorf("width %d: no People on the first row: %q", w, line)
		}
	}
}

// TestHeaderFoldsToASCII: every mark the header draws has a plain form.
func TestHeaderFoldsToASCII(t *testing.T) {
	v := headerFixture(t)
	v.Notice = "★ Badge earned: Stonecutter"
	ready := *v.Led
	ready.v.AllStand, ready.v.Ready = true, true
	ready.Pillars = append([]ledgerPillar(nil), ready.Pillars...)
	for i := range ready.Pillars {
		ready.Pillars[i].Met, ready.Pillars[i].Share = true, 1
	}
	for _, led := range []*ledgerData{v.Led, &ready} {
		v.Led = led
		g := renderHeader(v, 160, 3)
		g.fold()
		for y, line := range strings.Split(g.String(), "\n") {
			for _, r := range line {
				if r > 0x7e || r == '?' {
					t.Errorf("row %d has %q after the fold: %q", y, string(r), line)
					break
				}
			}
		}
	}
}

// headerStage draws the dashboard on the screen the clicks go to.
func headerStage(t *testing.T, d *Dashboard, pages *nativePages, w, h int) *stage {
	t.Helper()
	d.refresh()
	st := newStage(pages, w, h)
	st.paint()
	return st
}

// headerTarget is the middle cell of the header's click target n.
func headerTarget(t *testing.T, d *Dashboard, n int) image.Point {
	t.Helper()
	for _, h := range d.headerTV.grid.hits {
		if h.id == hitHeader && h.n == n {
			return image.Pt((h.r.Min.X+h.r.Max.X)/2, h.r.Min.Y)
		}
	}
	t.Fatalf("the header has no click target %d", n)
	return image.Point{}
}

// TestHeaderClicks: People and Morale open People, the bar and the waiting
// line open the Advance panel, an alert opens its panel; the pointer is the
// hand over each of them.
func TestHeaderClicks(t *testing.T) {
	restoreForge(t)
	d, _, pages := harbDashboard(t, "medieval_age")
	st := headerStage(t, d, pages, 120, 40)
	if len(d.headerTV.alerts) != 1 {
		t.Fatalf("staging: %d alerts with a harbinger", len(d.headerTV.alerts))
	}
	for _, c := range []struct {
		what string
		n    int
		want string
	}{
		{"People", headerPeople, "people"}, {"the bar", headerCharter, advancePanelName}, {"the harbinger", headerAlert0, "harbinger"},
	} {
		d.overlayMgr.Hide()
		at := headerTarget(t, d, c.n)
		st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
		if st.paint(); st.buf.Pointer != kit.PointerHand {
			t.Errorf("over %s the pointer's shape is %d, want the hand", c.what, st.buf.Pointer)
		}
		st.click(at.X, at.Y)
		if got := d.overlayMgr.ActiveName(); got != c.want {
			t.Errorf("a click on %s opened %q, want %q", c.what, got, c.want)
		}
	}
	d.overlayMgr.Hide()
	// Morale is People's: its words are inside the target.
	for _, h := range d.headerTV.grid.hits {
		if h.n == headerPeople {
			line := d.headerTV.grid.Lines()[h.r.Min.Y]
			part := string([]rune(line)[h.r.Min.X:h.r.Max.X])
			if !strings.Contains(part, "Morale") || !strings.Contains(part, "People") {
				t.Errorf("the People target covers %q, want People and Morale", part)
			}
		}
	}
	// Off the targets the pointer is the arrow.
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: 1, Y: 2})
	if st.paint(); st.buf.Pointer == kit.PointerHand {
		t.Error("the hand shows over the frame's edge")
	}
}

// TestHeaderAlertRunsItsCommand: a catastrophe chip runs catastrophe, a
// cheater badge runs nothing.
func TestHeaderAlertRunsItsCommand(t *testing.T) {
	restoreForge(t)
	d, _, pages := lastPassageDashboard(t, "galactic_age", false)
	var ran []string
	d.headerTV.run = func(s string) { ran = append(ran, s) }
	st := headerStage(t, d, pages, 120, 40)
	// The modal the refresh raised is over the header: close it.
	_, front := pages.GetFrontPage()
	keyTo(front, tk(kit.KeyEsc, 0, 0))
	st.paint()
	if got := d.headerTV.GetText(true); !strings.Contains(got, "☄ Last Passage") {
		t.Fatalf("no Last Passage chip: %q", got)
	}
	at := headerTarget(t, d, headerAlert0)
	st.click(at.X, at.Y)
	if len(ran) != 1 || ran[0] != "catastrophe" {
		t.Errorf("a click on the chip ran %v, want catastrophe", ran)
	}
	d.shame = &shameMessages[0]
	d.refresh()
	st.paint()
	text := d.headerTV.GetText(true)
	if !strings.Contains(text, string([]rune(kit.Strip(shameMessages[0].String()))[:8])) {
		t.Fatalf("the cheater badge is not on the header:\n%s", text)
	}
	if len(d.headerTV.alerts) != 3 || d.headerTV.alerts[0].cmd != "" {
		t.Errorf("alerts %+v: the cheater badge comes first and runs nothing", d.headerTV.alerts)
	}
}

// TestHeaderEscButton: the Esc button in the top edge presses Esc, and the
// pointer is the hand over it.
func TestHeaderEscButton(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	var pressed []kit.Key
	d.headerTV.esc.press = func(k kit.Key) { pressed = append(pressed, k) }
	st := headerStage(t, d, pages, 120, 40)
	keys := d.headerTV.grid.keys
	if len(keys) != 1 || keys[0].key.Code != kit.KeyEsc || keys[0].r.Min.Y != 0 {
		t.Fatalf("the header's keys are %+v, want one Esc in the top edge", keys)
	}
	if !strings.Contains(d.headerTV.GetText(true), " Esc  pause ") {
		t.Errorf("the button does not read Esc pause:\n%s", d.headerTV.GetText(true))
	}
	at := keys[0].r.Min.Add(image.Pt(1, 0))
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	if st.paint(); st.buf.Pointer != kit.PointerHand {
		t.Errorf("over the button the pointer's shape is %d, want the hand", st.buf.Pointer)
	}
	st.click(at.X, at.Y)
	if len(pressed) != 1 || pressed[0].Code != kit.KeyEsc {
		t.Errorf("a click on the button pressed %v, want Esc", pressed)
	}
}

// TestHeaderHoverLights: the part under the pointer lights up and goes back
// when the pointer leaves.
func TestHeaderHoverLights(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	st := headerStage(t, d, pages, 120, 40)
	at := headerTarget(t, d, headerCharter)
	bg := func() kit.Cell { return st.buf.At(at.X, at.Y) }
	rest := bg()
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	st.paint()
	if lit := bg(); lit.Bg == rest.Bg {
		t.Errorf("the bar's row does not light under the pointer: %+v", lit)
	}
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: 6})
	st.paint()
	if back := bg(); back != rest {
		t.Errorf("the bar's row stays lit after the pointer leaves: %+v, was %+v", back, rest)
	}
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	st.mouse(kit.Mouse{Kind: kit.MouseLeave, X: -1, Y: -1})
	st.paint()
	if back := bg(); back != rest {
		t.Errorf("the bar's row stays lit after the pointer leaves the screen: %+v", back)
	}
}

// TestHeaderNotice: the toast on show is in the top edge, between the name
// and the Esc button, and a long name gives way to it.
func TestHeaderNotice(t *testing.T) {
	v := headerFixture(t)
	v.Notice = "★ Badge earned: Stonecutter"
	for w := 60; w <= 200; w += 10 {
		top := renderHeader(v, w, 2).Lines()[0]
		if !strings.Contains(top, " ★ Badge earned: Stonecutter ") || !strings.Contains(top, "Esc") {
			t.Errorf("width %d: the top edge reads %q", w, top)
		}
		// The name stays, cut, and its end comes before the notice.
		if !strings.Contains(top, " A Very") || strings.Index(top, "★") < strings.Index(top, "A Very") {
			t.Errorf("width %d: the name is not before the notice: %q", w, top)
		}
	}
	long := renderHeader(v, 90, 2).Lines()[0]
	v.Notice = ""
	bare := renderHeader(v, 90, 2).Lines()[0]
	if !strings.Contains(bare, "\"Aspiring\"") || strings.Contains(long, "Aspiring") || !strings.Contains(long, "A Very Long") {
		t.Errorf("the long title should give way to the notice:\n%s\n%s", bare, long)
	}
	// A notice is cut to the room, not run over the name or the button.
	v.Notice = strings.Repeat("long notice ", 20)
	if top := renderHeader(v, 80, 2).Lines()[0]; runeLen(top) != 80 || !strings.Contains(top, "Esc") || !strings.Contains(top, "…") {
		t.Errorf("a very long notice: %q", top)
	}
}

// The plain tier folds the header the dashboard draws.
func TestHeaderPlainTier(t *testing.T) {
	if !headerPlain(mapSettings{Tier: mapmodel.TierASCII}) || headerPlain(mapSettings{Tier: mapmodel.TierUnicode}) {
		t.Error("headerPlain does not follow the glyph tier")
	}
}

// TestHeaderTitleWithoutAName: with no account the top edge holds the
// civilization title alone, with no dot before it.
func TestHeaderTitleWithoutAName(t *testing.T) {
	top := renderHeader(headerView{Age: "Bronze Age", Honour: "Aspiring"}, 100, 2).Lines()[0]
	if !strings.HasPrefix(top, "┌─ \"Aspiring\" ") {
		t.Errorf("the top edge reads %q", top)
	}
	if top := renderHeader(headerView{Age: "Bronze Age"}, 100, 2).Lines()[0]; strings.Contains(top, "·") || strings.Contains(top, "  ") && !strings.Contains(top, "Esc") {
		t.Errorf("an empty title leaves marks: %q", top)
	}
}
