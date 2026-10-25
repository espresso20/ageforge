package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// newSimScreen is an initialised w×h UTF-8 SimulationScreen, finished when
// the test ends.
func newSimScreen(t testing.TB, w, h int) tcell.SimulationScreen {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	must(t, sim.Init())
	t.Cleanup(sim.Fini)
	sim.SetSize(w, h)
	return sim
}

// frame is the cells a SimulationScreen showed at one moment, w to a row.
// Text reads a cell as its first rune, an empty cell as a space, and trims
// the blanks off the end of each row.
type frame struct {
	w     int
	cells []tcell.SimCell
}

// simCells is what sim shows now (after Show), copied so that the next draw
// on the same screen does not change it.
func simCells(sim tcell.SimulationScreen) frame {
	cells, w, _ := sim.GetContents()
	return frame{w: w, cells: append([]tcell.SimCell(nil), cells...)}
}

// drawPrimitive draws p over a fresh w×h screen and returns what it showed.
func drawPrimitive(t testing.TB, p tview.Primitive, w, h int) frame {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	must(t, sim.Init())
	defer sim.Fini()
	sim.SetSize(w, h)
	p.SetRect(0, 0, w, h)
	p.Draw(sim)
	sim.Show()
	return simCells(sim)
}

// Cells is every cell, row by row.
func (f frame) Cells() []tcell.SimCell { return f.cells }

// Rune is the rune drawn at column x of row y, a space where the cell is empty.
func (f frame) Rune(x, y int) rune {
	if rs := f.cells[y*f.w+x].Runes; len(rs) > 0 {
		return rs[0]
	}
	return ' '
}

// Region is the text of the w×h rectangle whose corner is at x, y: a string
// a row, the blanks trimmed off the end of each.
func (f frame) Region(x, y, w, h int) []string {
	rows := make([]string, h)
	for row := range rows {
		var b strings.Builder
		for col := 0; col < w; col++ {
			b.WriteRune(f.Rune(x+col, y+row))
		}
		rows[row] = strings.TrimRight(b.String(), " ")
	}
	return rows
}

// Rows is the whole screen as text, a string a row.
func (f frame) Rows() []string {
	if f.w == 0 {
		return nil
	}
	return f.Region(0, 0, f.w, len(f.cells)/f.w)
}

// Text is the whole screen as text, a line a row.
func (f frame) Text() string { return strings.Join(f.Rows(), "\n") }

// newTestDashboard is a dashboard over eng (a fresh game when eng is nil)
// on the "dashboard" page of its own Pages, as the app lays them out.
func newTestDashboard(eng *game.GameEngine) (*Dashboard, *tview.Pages) {
	if eng == nil {
		eng = game.NewGameEngine()
	}
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	return d, pages
}
