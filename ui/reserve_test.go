package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// reserveState is a real Primitive Age state with workers, set to the given
// reserve and hold.
func reserveState(t *testing.T, by string, net float64, feeds int) game.GameState {
	t.Helper()
	ge := sharesTestEngine(t)
	game.DevExecCommand("/recruit 5", ge)
	st := ge.GetState()
	st.Workers.TotalIdle, st.Workers.HoldTicks = 3, 0
	st.Workers.HeldBy, st.Workers.FoodNet, st.Workers.FoodFeeds = by, net, feeds
	if wt, ok := st.Workers.Types["worker"]; ok {
		wt.IdleCount = 3
		st.Workers.Types["worker"] = wt
	}
	return st
}

// The People box names each kind of person in the age's own word, says what
// holds the population back, and keeps every row whole in the 20 cells it
// has.
func TestReserve_DashboardBox(t *testing.T) {
	cases := []struct {
		by    string
		net   float64
		feeds int
		want  []string
	}{
		{game.HeldHousing, 4, 40, []string{"Held by: housing"}},
		{game.HeldFood, 0.45, 4, []string{"Held by: food", "+0.45 feeds 4 more"}},
		{game.HeldNone, 4, 40, nil},
	}
	for _, c := range cases {
		st := reserveState(t, c.by, c.net, c.feeds)
		text := plainText(peopleBoxText(st, sidebarW-2, true))
		for _, want := range []string{"Pop ", "Gatherers", "Clanfolk", "Hunters", "Elders"} {
			if !strings.Contains(text, want) {
				t.Errorf("held by %q: the box does not show %q:\n%s", c.by, want, text)
			}
		}
		if reserve := lineContaining(text, "Clanfolk"); !strings.HasSuffix(reserve, " 3") {
			t.Errorf("held by %q: the reserve's line should end with its count, 3: %q", c.by, reserve)
		}
		for _, want := range c.want {
			if !strings.Contains(text, want) {
				t.Errorf("held by %q: the box does not show %q:\n%s", c.by, want, text)
			}
		}
		if c.want == nil && strings.Contains(text, "Held by") {
			t.Errorf("growing, but the box says it is held:\n%s", text)
		}
		for _, row := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
			if runeLen(row) > sidebarW-2 {
				t.Errorf("held by %q: row %q is wider than the box", c.by, row)
			}
		}
	}
	// With room, the food line says it all in one sentence.
	st := reserveState(t, game.HeldFood, 0.45, 4)
	if got := heldByText(st.Workers); got != "Held by: food (+0.45 food a tick feeds 4 more)" {
		t.Errorf("heldByText = %q", got)
	}
}

// The People panel's Overview shows the Held-by line and the age's names, and
// its Work tab names what holds growth rather than a filled-slots limit.
func TestReserve_PeoplePanel(t *testing.T) {
	for _, c := range []struct{ by, held, recruit string }{
		{game.HeldHousing, "held by housing", "held by housing"},
		{game.HeldFood, "held by food (+0.45 food a tick feeds 4 more)", "held by food"},
	} {
		st := reserveState(t, c.by, 0.45, 4)
		over := renderPeople(st, peopleView{tab: tabOverview}, 100, 30).String()
		if !strings.Contains(over, "Clanfolk") || !strings.Contains(over, c.held) {
			t.Errorf("held by %q: the Overview lacks the reserve or %q:\n%s", c.by, c.held, over)
		}
		work := renderPeople(st, peopleView{tab: tabWork}, 100, 30).String()
		if !strings.Contains(work, c.recruit) || strings.Contains(work, "every worker slot is filled") {
			t.Errorf("held by %q: the Work tab's auto-recruit line does not say %q, or still blames the slots:\n%s", c.by, c.recruit, work)
		}
	}
}

// On the real dashboard at the smallest sizes the People box grows by the
// lines that say what holds the population back, and nothing else gives:
// every row of the box and of the Panels list is whole.
func TestReserve_DashboardAt80x24(t *testing.T) {
	restoreForge(t)
	d, pages, eng := stagedDashboard(t, "bronze_age")
	_ = eng
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		w, h := sz[0], sz[1]
		for _, by := range []string{game.HeldHousing, game.HeldFood} {
			drawnDashboard(t, d, pages, w, h)
			st := *d.lastState
			st.Workers.TotalIdle = 494
			st.Workers.HeldBy, st.Workers.FoodNet, st.Workers.FoodFeeds = by, 51.2, 512
			d.lastState = &st
			d.refreshWorkerMini(st)
			sim := tcell.NewSimulationScreen("UTF-8")
			if err := sim.Init(); err != nil {
				t.Fatal(err)
			}
			sim.SetSize(w, h)
			screen := theme.WrapScreen(sim)
			for i := 0; i < 2; i++ {
				screen.Clear()
				pages.SetRect(0, 0, w, h)
				pages.Draw(screen)
				screen.Show()
			}
			cells, _, _ := sim.GetContents()
			cells = append([]tcell.SimCell(nil), cells...)
			sim.Fini()
			where := fmt.Sprintf("held by %s at %dx%d", by, w, h)
			rows := checkFitted(t, where+", People", d.workerMiniTV, cells, w)
			text := strings.Join(rows, "\n")
			want := []string{"Held by: housing"}
			if by == game.HeldFood {
				want = []string{"Held by: food"}
				if h >= workerDetailMinRows {
					want = append(want, "+51.2 feeds 512 more")
				} else if strings.Contains(text, "feeds") {
					t.Errorf("%s: the food detail line takes a row a 24-row terminal needs:\n%s", where, text)
				}
			}
			for _, s := range append([]string{"Craftsfolk", "Villagers 494", "Spearmen", "Scribes"}, want...) {
				if !strings.Contains(text, s) && !strings.Contains(strings.Join(strings.Fields(text), " "), s) {
					t.Errorf("%s: the People box does not show %q:\n%s", where, s, text)
				}
			}
			// The headcount is a line of its own, or, on a short terminal with
			// a Held-by line up, in the box's title.
			if h >= workerDetailMinRows {
				if !strings.Contains(text, "Pop ") {
					t.Errorf("%s: the People box has no Pop line:\n%s", where, text)
				}
			} else if title := d.workerMiniTV.GetTitle(); !strings.Contains(title, "People ") || !strings.Contains(title, "/") {
				t.Errorf("%s: the title %q does not carry the headcount", where, title)
			}
			panels := strings.Join(checkFitted(t, where+", Panels", d.sidebar, cells, w), "\n")
			for _, name := range sidebarPanels {
				if !strings.Contains(panels, name) {
					t.Errorf("%s: the Panels list does not show %q", where, name)
				}
			}
		}
	}
}

// A late-game town, with counts that need the whole box, still leaves the
// Panels list whole at 80x24 with the Held-by line up.
func TestReserve_BigTownAt80x24(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "quantum_age")
	drawnDashboard(t, d, pages, 80, 24)
	st := *d.lastState
	st.Workers.TotalPop, st.Workers.MaxPop, st.Workers.TotalIdle, st.Workers.Scholars = 7680, 97800000, 4321, 120
	st.Military.SoldierCount = 1234567
	st.Workers.HeldBy = game.HeldFood
	d.lastState = &st
	d.refreshWorkerMini(st)
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(80, 24)
	screen := theme.WrapScreen(sim)
	for i := 0; i < 2; i++ {
		screen.Clear()
		pages.SetRect(0, 0, 80, 24)
		pages.Draw(screen)
		screen.Show()
	}
	cells, _, _ := sim.GetContents()
	cells = append([]tcell.SimCell(nil), cells...)
	text := strings.Join(checkFitted(t, "big town, People", d.workerMiniTV, cells, 80), "\n")
	if title := d.workerMiniTV.GetTitle(); !strings.Contains(title, "People") {
		t.Errorf("the title %q lost the box's name", title)
	}
	for _, s := range []string{"Operators", "Entangled", "Wardens", "Quantum minds", "Held by: food"} {
		if !strings.Contains(text, s) {
			t.Errorf("the People box does not show %q:\n%s", s, text)
		}
	}
	panels := strings.Join(checkFitted(t, "big town, Panels", d.sidebar, cells, 80), "\n")
	for _, name := range sidebarPanels {
		if !strings.Contains(panels, name) {
			t.Errorf("the Panels list does not show %q", name)
		}
	}
}
