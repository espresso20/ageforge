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

// The Workers box says Reserve, and what holds the population back, in the
// 20 cells it has, with every row whole.
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
		text := plainText(workerMiniText(st, sidebarW-2, true))
		if !strings.Contains(text, "Reserve: 3") || strings.Contains(text, "Idle") {
			t.Errorf("held by %q: the box does not show the reserve:\n%s", c.by, text)
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

// The Workers panel shows Reserve and the Held-by line in its Summary, and
// its Auto-recruit line names what holds growth rather than a filled-slots
// limit.
func TestReserve_RosterPanel(t *testing.T) {
	for _, c := range []struct{ by, held, recruit string }{
		{game.HeldHousing, "Held by: housing", "held by housing"},
		{game.HeldFood, "Held by: food (+0.45 food a tick feeds 4 more)", "held by food"},
	} {
		st := reserveState(t, c.by, 0.45, 4)
		text := plainText(workersProvider(st, 80))
		if !strings.Contains(text, "Reserve: 3") || !strings.Contains(text, c.held) {
			t.Errorf("held by %q: the panel's summary lacks the reserve or %q:\n%s", c.by, c.held, text)
		}
		if !strings.Contains(text, c.recruit) || strings.Contains(text, "every worker slot is filled") {
			t.Errorf("held by %q: the Auto-recruit line does not say %q, or still blames the slots:\n%s", c.by, c.recruit, text)
		}
		for _, row := range strings.Split(text, "\n") {
			if runeLen(row) > 80 {
				t.Errorf("held by %q: panel row wider than 80 cells: %q", c.by, row)
			}
		}
	}
}

// On the real dashboard at the smallest sizes the Workers box grows by the
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
			rows := checkFitted(t, where+", Workers", d.workerMiniTV, cells, w)
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
			for _, s := range append([]string{"Reserve: 494", "Housing left:", "Food net:"}, want...) {
				if !strings.Contains(text, s) {
					t.Errorf("%s: the Workers box does not show %q:\n%s", where, s, text)
				}
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
	st.Workers.TotalPop, st.Workers.MaxPop, st.Workers.TotalIdle = 7680, 97800000, 4321
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
	text := strings.Join(checkFitted(t, "big town, Workers", d.workerMiniTV, cells, 80), "\n")
	for _, s := range []string{"Reserve:", "Held by: food", "Food net:"} {
		if !strings.Contains(text, s) {
			t.Errorf("the Workers box does not show %q:\n%s", s, text)
		}
	}
	panels := strings.Join(checkFitted(t, "big town, Panels", d.sidebar, cells, 80), "\n")
	for _, name := range sidebarPanels {
		if !strings.Contains(panels, name) {
			t.Errorf("the Panels list does not show %q", name)
		}
	}
}
