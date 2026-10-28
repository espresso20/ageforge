package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
)

// glyph_grid.go is the block of cells the drawn panels render into before
// anything reaches the screen: the Research panel and the badge case (tGrid,
// cell_grid.go), the Expeditions, People and Advance panels (xgrid,
// expedition_grid.go) and the main menu (mGrid, menu_scene.go). A panel is a
// value a test can read, and one painter puts it on the screen.

// grid is w by h cells, row after row. blank is what a cell outside it reads
// as, and what a new cell is.
type grid[C any] struct {
	w, h  int
	c     []C
	blank C
}

func newGrid[C any](w, h int, blank C) grid[C] {
	g := grid[C]{w: max(w, 0), h: max(h, 0), blank: blank}
	g.c = make([]C, g.w*g.h)
	for i := range g.c {
		g.c[i] = blank
	}
	return g
}

func (g *grid[C]) in(x, y int) bool { return x >= 0 && y >= 0 && x < g.w && y < g.h }

func (g *grid[C]) at(x, y int) C {
	if !g.in(x, y) {
		return g.blank
	}
	return g.c[y*g.w+x]
}

// store writes a cell; outside the grid it does nothing.
func (g *grid[C]) store(x, y int, c C) {
	if g.in(x, y) {
		g.c[y*g.w+x] = c
	}
}

// draw puts the grid on the screen with its top left corner at (x0, y0); look
// is the glyph and the style each cell is drawn in.
func (g *grid[C]) draw(scr tcell.Screen, x0, y0 int, look func(C) (rune, tcell.Style)) {
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			r, style := look(g.c[y*g.w+x])
			scr.SetContent(x0+x, y0+y, r, nil, style)
		}
	}
}

// Cell flags.
const (
	cfBold   uint8 = 1 << iota // drawn bold
	cfSel                      // on the selection's background
	cfItalic                   // drawn in italics
)

// gcell is one cell of a glyph grid: a rune and how it is painted. The kind
// k is the grid's own (a theme role, a map colour class); the rest is used by
// the grids that need it.
type gcell[K comparable] struct {
	r    rune
	k    K
	lane int8 // index into the tree's lanes for tsLane, else -1
	fl   uint8
	// ink, when set, is the cell's colour in place of its kind's.
	ink ink
}

// glyphGrid is a grid of gcells, with the plain glyph tier's fold.
type glyphGrid[K comparable] struct {
	grid[gcell[K]]
}

func newGlyphGrid[K comparable](w, h int) glyphGrid[K] {
	return glyphGrid[K]{newGrid(w, h, gcell[K]{r: ' ', lane: -1})}
}

// set writes rune r painted as k.
func (g *glyphGrid[K]) set(x, y int, r rune, k K) {
	g.store(x, y, gcell[K]{r: r, k: k, lane: -1})
}

// foldWith rewrites the grid for the plain glyph tier: nothing but ASCII.
// table is the panel's own fold, for the marks and edges where the map
// model's fold would lose the difference between two of them; what it does
// not list folds as the map does.
func (g *glyphGrid[K]) foldWith(table map[rune]rune) {
	for i := range g.c {
		g.c[i].r = foldRune(g.c[i].r, table)
	}
}

// foldRune is r in ASCII: as table has it, else as the map model folds it.
func foldRune(r rune, table map[rune]rune) rune {
	if r < 0x80 {
		return r
	}
	if folded, ok := table[r]; ok {
		return folded
	}
	return mapmodel.Fold(r, mapmodel.TierASCII)
}
