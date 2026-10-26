// Package mapgen draws the settlement map: a seeded, layered, deterministic
// generator that turns (map model, viewport, animation frame) into a grid of
// coloured cells plus an index from cell to thing. It knows nothing of the
// screen; the ui only blits the grid and draws the cursor.
//
// The layers, bottom to top: fields (height, moisture, sea, rivers), the
// settlement plan (roads and parcels), placement (what stands on which
// parcel), age style (what it looks like), structures, joins, detail scatter,
// agents and light. Orbit scenes swap the ground for a starfield, a planet
// and ring roads. Every choice is a hash of integers, so a frame is a pure
// function of its input, byte for byte, on every machine.
package mapgen

// Cell is one character of the picture: a glyph in a foreground paint over a
// background paint, pulled toward the night tint by Dim (of 255).
type Cell struct {
	R      rune
	Fg, Bg Paint
	Dim    uint8
}

// Kind says what sort of thing a cell shows.
type Kind uint8

const (
	KindNone Kind = iota
	KindTerrain
	KindRoad
	KindBuilding
	KindWonder
	KindAgent
	KindAmbient // the ordinary houses of the people, which no building type owns
)

// Thing is what the inspect index points at.
type Thing struct {
	Kind    Kind
	Key     string // building or wonder key, terrain name, agent kind
	Lineage string
	Count   int // copies a building parcel stands for
	Legacy  bool
	Ruin    bool
}

// Grid is a finished picture: the cells, and for each cell the thing it
// shows (an index into Things, 0 for nothing). Wall is the hue of the age's
// walls, which the building lines' hues lean on, and Tint the hue the dark
// falls toward.
type Grid struct {
	W, H   int
	Wall   Hue
	Tint   Hue
	Cells  []Cell
	Hit    []uint16
	Things []Thing
	index  map[Thing]uint16
}

func newGrid(w, h int) *Grid {
	return &Grid{W: w, H: h, Cells: make([]Cell, w*h), Hit: make([]uint16, w*h),
		Things: []Thing{{}}, index: map[Thing]uint16{}}
}

// At returns the cell at (x, y); off the grid it is blank.
func (g *Grid) At(x, y int) Cell {
	if !g.in(x, y) {
		return Cell{}
	}
	return g.Cells[y*g.W+x]
}

// ThingAt is the thing shown at (x, y).
func (g *Grid) ThingAt(x, y int) Thing {
	if !g.in(x, y) {
		return Thing{}
	}
	return g.Things[g.Hit[y*g.W+x]]
}

func (g *Grid) in(x, y int) bool { return x >= 0 && y >= 0 && x < g.W && y < g.H }

// draw writes a glyph and keeps the background below it.
func (g *Grid) draw(x, y int, r rune, fg Paint) {
	if g.in(x, y) {
		c := &g.Cells[y*g.W+x]
		c.R, c.Fg = r, fg
	}
}

// paint writes a whole cell.
func (g *Grid) paint(x, y int, c Cell) {
	if g.in(x, y) {
		g.Cells[y*g.W+x] = c
	}
}

// mark says what the cell at (x, y) shows.
func (g *Grid) mark(x, y int, t Thing) {
	if !g.in(x, y) {
		return
	}
	i, ok := g.index[t]
	if !ok {
		i = uint16(len(g.Things))
		g.Things = append(g.Things, t)
		g.index[t] = i
	}
	g.Hit[y*g.W+x] = i
}

// Window copies the w by h part of g whose top-left corner is (x0, y0).
func (g *Grid) Window(x0, y0, w, h int) *Grid {
	out := newGrid(w, h)
	out.Wall, out.Tint = g.Wall, g.Tint
	out.Things = g.Things
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if g.in(x0+x, y0+y) {
				out.Cells[y*w+x] = g.Cells[(y0+y)*g.W+x0+x]
				out.Hit[y*w+x] = g.Hit[(y0+y)*g.W+x0+x]
			}
		}
	}
	return out
}

// Text is the grid as plain text, trailing spaces trimmed.
func (g *Grid) Text() string {
	var b []byte
	for y := 0; y < g.H; y++ {
		row := []rune{}
		for x := 0; x < g.W; x++ {
			row = append(row, g.Cells[y*g.W+x].R)
		}
		end := len(row)
		for end > 0 && (row[end-1] == ' ' || row[end-1] == 0) {
			end--
		}
		for _, r := range row[:end] {
			if r == 0 {
				r = ' '
			}
			b = append(b, string(r)...)
		}
		b = append(b, '\n')
	}
	return string(b)
}
