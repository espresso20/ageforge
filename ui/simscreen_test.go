package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A simScreen is the screen of a test: a grid of cells that primitives are
// drawn on (drawOn) and read back from. It keeps the calls the tests were
// written with, from when the screen they read was a terminal's.
type simScreen struct {
	buf  kit.Buffer
	cur  image.Point
	curs bool
}

// A simCell is one cell read back: its rune (and what combines with it) and
// its style, the colors literal as they were painted.
type simCell struct {
	Runes []rune
	Style kit.Style
}

// newSim is a blank screen the size a terminal starts at.
func newSim() *simScreen {
	s := &simScreen{}
	s.SetSize(80, 25)
	return s
}

// newSimScreen is a blank w×h screen.
func newSimScreen(_ testing.TB, w, h int) *simScreen {
	s := &simScreen{}
	s.SetSize(w, h)
	return s
}

func (s *simScreen) Init() error      { return nil }
func (s *simScreen) Fini()            {}
func (s *simScreen) Show()            {}
func (s *simScreen) SetSize(w, h int) { s.buf.Resize(w, h) }
func (s *simScreen) Size() (int, int) { return s.buf.W, s.buf.H }

// Clear blanks the screen to the theme's text on its ground.
func (s *simScreen) Clear() {
	s.buf.Reset()
	s.curs = false
}

func (s *simScreen) HideCursor()                { s.curs = false }
func (s *simScreen) ShowCursor(x, y int)        { s.cur, s.curs = image.Pt(x, y), true }
func (s *simScreen) GetCursor() (int, int, bool) { return s.cur.X, s.cur.Y, s.curs }

// SetContent writes one cell: the screen is a mapstyle.Cells, so the map and
// the weather can be drawn straight onto it.
func (s *simScreen) SetContent(x, y int, r rune, _ []rune, st kit.Style) {
	kit.NewPainter(&s.buf).Put(x, y, r, st)
}

func (s *simScreen) cell(c kit.Cell) simCell {
	st := kit.Style{Fg: kit.RGB(c.Fg.R, c.Fg.G, c.Fg.B), Bg: kit.Clear, Attr: c.Attr}
	if !c.Clear() {
		st.Bg = kit.RGB(c.Bg.R, c.Bg.G, c.Bg.B)
	}
	rs := []rune{c.R}
	if c.Cont || c.R == 0 {
		rs[0] = ' '
	}
	return simCell{Runes: append(rs, []rune(c.Comb)...), Style: st}
}

func (s *simScreen) GetContent(x, y int) (rune, []rune, kit.Style, int) {
	c := s.cell(s.buf.At(x, y))
	return c.Runes[0], c.Runes[1:], c.Style, 1
}

// GetContents is every cell, row by row, and the screen's size.
func (s *simScreen) GetContents() ([]simCell, int, int) {
	out := make([]simCell, len(s.buf.Cells))
	for i, c := range s.buf.Cells {
		out[i] = s.cell(c)
	}
	return out, s.buf.W, s.buf.H
}

// drawOn paints p, at the rectangle it has been given, onto scr: the cells p
// covers and no others, so what was drawn under a gap in it still shows.
func drawOn(scr *simScreen, p Primitive) {
	x, y, w, h := p.GetRect()
	root := kit.NewRoot(p)
	root.Resize(max(w, 0), max(h, 0))
	var b kit.Buffer
	root.Paint(&b)
	for _, r := range coveredBy(p) {
		r = r.Intersect(image.Rect(0, 0, b.W, b.H))
		for cy := r.Min.Y; cy < r.Max.Y; cy++ {
			for cx := r.Min.X; cx < r.Max.X; cx++ {
				if sx, sy := x+cx, y+cy; sx >= 0 && sy >= 0 && sx < scr.buf.W && sy < scr.buf.H {
					scr.buf.Cells[sy*scr.buf.W+sx] = b.Cells[cy*b.W+cx]
				}
			}
		}
	}
	if b.CursorOn && p.HasFocus() {
		scr.ShowCursor(x+b.Cursor.X, y+b.Cursor.Y)
	}
}

// decompose is a cell's style in parts.
func decompose(st kit.Style) (fg, bg theme.Color, attr kit.Attr) { return st.Decompose() }

// tk is a key press.
func tk(code kit.KeyCode, r rune, mod kit.Mod) kit.Key { return kit.Key{Code: code, Rune: r, Mod: mod} }

// captured gives p's capture a key, as on its way to p, and returns what the
// capture passed on: nil when it took the key.
func captured(p kit.KeyFilter, k kit.Key) *kit.Key {
	out, on := p.FilterKey(k)
	if !on {
		return nil
	}
	return &out
}

// A stage is a primitive laid out and painted alone on a screen, with the
// mouse: what the window does with the game's root.
type stage struct {
	p    Primitive
	root *kit.Root
	buf  kit.Buffer
}

// newStage puts p on a w by h screen and paints it once.
func newStage(p Primitive, w, h int) *stage {
	s := &stage{p: p, root: kit.NewRoot(p)}
	p.SetRect(0, 0, w, h)
	s.root.Resize(w, h)
	s.paint()
	return s
}

func (s *stage) paint()                 { s.root.Paint(&s.buf) }
func (s *stage) mouse(m kit.Mouse) bool { return s.root.Mouse(m) }

// click presses and lets go of the left button at x, y.
func (s *stage) click(x, y int) {
	s.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x, Y: y})
	s.mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: x, Y: y})
}

// frame is the cells a screen showed at one moment, w to a row. Text reads a
// cell as its first rune, an empty cell as a space, and trims the blanks off
// the end of each row.
type frame struct {
	w     int
	cells []simCell
}

// simCells is what sim shows now, copied so that the next draw on the same
// screen does not change it.
func simCells(sim *simScreen) frame {
	cells, w, _ := sim.GetContents()
	return frame{w: w, cells: cells}
}

// drawPrimitive draws p over a fresh w×h screen and returns what it showed.
func drawPrimitive(_ testing.TB, p Primitive, w, h int) frame {
	sim := newSimScreen(nil, w, h)
	p.SetRect(0, 0, w, h)
	drawOn(sim, p)
	return simCells(sim)
}

// drawWrapped draws p on a fresh w×h screen passes times (the first gives
// every box its size, the next ones lay out what depends on it) and returns
// what the last showed.
func drawWrapped(_ testing.TB, p Primitive, w, h, passes int) frame {
	sim := newSimScreen(nil, w, h)
	for i := 0; i < passes; i++ {
		sim.Clear()
		p.SetRect(0, 0, w, h)
		drawOn(sim, p)
	}
	return simCells(sim)
}

// Cells is every cell, row by row.
func (f frame) Cells() []simCell { return f.cells }

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
func newTestDashboard(eng *game.GameEngine) (*Dashboard, *nativePages) {
	if eng == nil {
		eng = game.NewGameEngine()
	}
	pages := newNativePages()
	d := NewDashboard(newAppLoop(), eng, pages)
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
func flushRows(t testing.TB, where string, sim *simScreen, g *mGrid, pal *menuPalette, plain bool) []string {
	t.Helper()
	sim.Clear()
	w, h := sim.Size()
	g.paint(kit.NewPainter(&sim.buf), image.Point{}, pal, plain)
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
func wordReadable(t testing.TB, where string, sim *simScreen, rows []string, word string) {
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
			fg, bg, _ := decompose(st)
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
func extraCells(sim *simScreen, g *mGrid) (n int) {
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
