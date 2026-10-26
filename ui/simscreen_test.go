package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
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

// drawWrapped draws p on a fresh w×h screen through theme.WrapScreen, as the
// app does, passes times (the first gives every box its size, the next ones
// lay out what depends on it) and returns what the last showed.
func drawWrapped(t testing.TB, p tview.Primitive, w, h, passes int) frame {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	must(t, sim.Init())
	defer sim.Fini()
	sim.SetSize(w, h)
	screen := theme.WrapScreen(sim)
	for i := 0; i < passes; i++ {
		screen.Clear()
		p.SetRect(0, 0, w, h)
		p.Draw(screen)
		screen.Show()
	}
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

// keepTheme puts the active theme back when the test ends.
func keepTheme(t testing.TB) {
	t.Helper()
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
}

// eachTheme runs fn as a subtest named for each registered theme, with that
// theme active, and puts the active theme back afterwards.
func eachTheme(t *testing.T, fn func(t *testing.T)) {
	t.Helper()
	keepTheme(t)
	for _, th := range theme.All() {
		must(t, theme.SetActive(th.Key))
		t.Run(th.Key, fn)
	}
}

// eachThemeInline calls fn for each registered theme, with that theme
// active, in the caller's own test, and puts the active theme back
// afterwards.
func eachThemeInline(t testing.TB, fn func(th theme.Theme)) {
	t.Helper()
	keepTheme(t)
	for _, th := range theme.All() {
		must(t, theme.SetActive(th.Key))
		fn(th)
	}
}

// foldASCII is s as the plain glyph set shows it when plain, and s as it is
// when not.
func foldASCII(s string, plain bool) string {
	if !plain {
		return s
	}
	out := []rune(plainSafe(s))
	for i, c := range out {
		out[i] = mapmodel.Fold(c, mapmodel.TierASCII)
	}
	return string(out)
}

// flushRows clears sim, puts g on it as the app does (through
// theme.WrapScreen) and returns the rune in every cell, a string a row. In
// the plain glyph set a rune beyond ASCII fails the test; where names the
// screen for the failure.
func flushRows(t testing.TB, where string, sim tcell.SimulationScreen, g *mGrid, pal *menuPalette, plain bool) []string {
	t.Helper()
	sim.Clear()
	g.flush(theme.WrapScreen(sim), 0, 0, pal, plain)
	w, h := sim.Size()
	rows := make([]string, h)
	for y := range rows {
		row := make([]rune, w)
		for x := range row {
			row[x], _, _, _ = sim.GetContent(x, y)
			if plain && row[x] > 0x7e {
				t.Fatalf("%s: cell (%d,%d) holds %q, which the plain glyph set does not have", where, x, y, row[x])
			}
		}
		rows[y] = string(row)
	}
	return rows
}

// wordReadable fails the test unless word is on one of rows (what flushRows
// returned for sim) and each of its letters is drawn at a contrast of 3 or
// more on what is behind it. Only the first place it is found is held to
// that. where names the screen for the failure.
func wordReadable(t testing.TB, where string, sim tcell.SimulationScreen, rows []string, word string) {
	t.Helper()
	if word = strings.TrimSpace(word); word == "" {
		return
	}
	for y, row := range rows {
		at := strings.Index(row, word)
		if at < 0 {
			continue
		}
		x0 := len([]rune(row[:at]))
		for i, c := range []rune(word) {
			if c == ' ' {
				continue
			}
			_, _, st, _ := sim.GetContent(x0+i, y)
			fg, bg, _ := st.Decompose()
			if k := theme.ContrastRatio(fg, bg); k < 3-0.02 {
				t.Errorf("%s: %q is drawn at a contrast of %.2f (%06x on %06x)", where, word, k, fg.Hex(), bg.Hex())
				return
			}
		}
		return
	}
	t.Errorf("%s: %q is not on the screen", where, word)
}

// extraCells counts the cells of sim that show a rune other than the one g
// drew there (no rune and a blank are the same): what something else, such as
// a theme's ambient effect, put on the page.
func extraCells(sim tcell.SimulationScreen, g *mGrid) (n int) {
	w, h := sim.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if c, _, _, _ := sim.GetContent(x, y); c != g.c[y*w+x].r && !(c == ' ' && g.c[y*w+x].r == 0) {
				n++
			}
		}
	}
	return n
}
