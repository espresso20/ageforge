package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// cell_grid.go is the cell grid the drawn panels share: the Research panel
// (research_tree.go) and the badge case (badge_case.go) both render into a
// tGrid as a pure function of what they show, and one painter puts a grid
// on the screen in the active theme's colours. A cell says how it is
// painted by a style (a theme role) or by an ink (a colour that belongs to
// the art, such as a badge's metal), never by a colour value: the painter
// picks the value for the theme and holds it to the contrast rule, so a
// grid is the same in every theme and is checked in each.

// tStyle is how a cell is painted. The painter maps each to theme roles.
type tStyle uint8

const (
	tsText    tStyle = iota // body text
	tsDim                   // locked, later, hints
	tsBright                // ready to start; the strongest ink on the canvas
	tsHi                    // in progress, numbers
	tsLane                  // the cell's lane colour (researched); the accent without a lane
	tsGold                  // the selected tech's chain, steps still to do
	tsGoldDim               // the chain's steps already done
	tsGood                  // names you hold
	tsSel                   // the selection
	tsChip                  // title and key bars
	tsChipKey               // a key on a bar
	tsChipDim               // quiet text on a bar
	tsBorder                // the card's frame
	tsBad                   // a strike, a warning
	tsSlab                  // a hidden badge: the chip colour as ink
	tsLabel                 // a value's label
	tsCount                 // how many styles there are
)

// Cell flags.
const (
	cfBold uint8 = 1 << iota // drawn bold
	cfSel                    // on the selection's background
)

// tCell is one cell of a panel.
type tCell struct {
	r    rune
	st   tStyle
	lane int8 // index into the tree's lanes for tsLane, else -1
	fl   uint8
	// ink, when set, is the cell's colour in place of its style's.
	ink ink
}

// tGrid is a block of cells.
type tGrid struct {
	w, h int
	c    []tCell
}

func newTGrid(w, h int) *tGrid {
	g := &tGrid{w: max(w, 0), h: max(h, 0)}
	g.c = make([]tCell, g.w*g.h)
	for i := range g.c {
		g.c[i] = tCell{r: ' ', lane: -1}
	}
	return g
}

func (g *tGrid) in(x, y int) bool { return x >= 0 && y >= 0 && x < g.w && y < g.h }

func (g *tGrid) at(x, y int) tCell {
	if !g.in(x, y) {
		return tCell{r: ' ', lane: -1}
	}
	return g.c[y*g.w+x]
}

func (g *tGrid) put(x, y int, r rune, st tStyle, lane int) {
	if g.in(x, y) {
		g.c[y*g.w+x] = tCell{r: r, st: st, lane: int8(lane)}
	}
}

// paint writes a cell in an ink, with flags. The style is what the cell
// falls back to where an ink has no say (a test that reads styles).
func (g *tGrid) paint(x, y int, r rune, k ink, st tStyle, fl uint8) {
	if g.in(x, y) {
		g.c[y*g.w+x] = tCell{r: r, st: st, lane: -1, ink: k, fl: fl}
	}
}

// flag adds flags to a cell that is already drawn.
func (g *tGrid) flag(x, y int, fl uint8) {
	if g.in(x, y) {
		g.c[y*g.w+x].fl |= fl
	}
}

// text writes s from (x, y), one cell a rune, and returns the next x.
func (g *tGrid) text(x, y int, s string, st tStyle, lane int) int {
	for _, r := range s {
		g.put(x, y, r, st, lane)
		x++
	}
	return x
}

// stamp copies src onto g with its top left corner at (x, y). Cells of
// src outside g are dropped.
func (g *tGrid) stamp(x, y int, src *tGrid) {
	for sy := 0; sy < src.h; sy++ {
		for sx := 0; sx < src.w; sx++ {
			if g.in(x+sx, y+sy) {
				g.c[(y+sy)*g.w+x+sx] = src.c[sy*src.w+sx]
			}
		}
	}
}

// String is the grid as lines of text, trailing spaces kept.
func (g *tGrid) String() string {
	var sb strings.Builder
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			sb.WriteRune(g.c[y*g.w+x].r)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// foldGrid rewrites a grid for the plain glyph tier: nothing but ASCII.
// table is the panel's own fold, for the marks and edges where the map
// model's fold would lose the difference between two of them; what it does
// not list folds as the map does.
func foldGrid(g *tGrid, table map[rune]rune) {
	for i, c := range g.c {
		if c.r < 0x80 {
			continue
		}
		if r, ok := table[c.r]; ok {
			g.c[i].r = r
		} else {
			g.c[i].r = mapmodel.Fold(c.r, mapmodel.TierASCII)
		}
	}
}

// gridPalette is the active theme's colours for a grid: a style for each
// tStyle, the lanes' hues, and the inks.
type gridPalette struct {
	styles  [tsCount]tcell.Style
	lane    []tcell.Style
	bg, sel tcell.Color
	light   bool
	duo     bool
	text    tcell.Color
	// inks caches each ink's colour on the canvas and on the selection.
	inks [2]map[ink]tcell.Color
}

// newGridPalette reads the active theme. Lane hues and inks go through the
// contrast rule for art (3.0); text roles are held to the rule for text.
func newGridPalette(lanes int, hue func(i int) string) *gridPalette {
	p := &gridPalette{
		bg:    theme.Color(theme.RoleBackground),
		sel:   theme.Color(theme.RoleSelection),
		text:  theme.Color(theme.RoleText),
		light: theme.IsLight(),
		duo:   theme.Active().Duotone,
	}
	p.inks[0], p.inks[1] = map[ink]tcell.Color{}, map[ink]tcell.Color{}
	chip := theme.Color(theme.RoleChip)
	on := func(role theme.Role, back tcell.Color, ratio float64) tcell.Style {
		return tcell.StyleDefault.Background(back).Foreground(theme.Legible(theme.Color(role), back, ratio))
	}
	p.styles[tsText] = on(theme.RoleText, p.bg, 4.5)
	p.styles[tsDim] = on(theme.RoleDim, p.bg, 3)
	p.styles[tsBright] = on(theme.RoleBright, p.bg, 4.5).Bold(true)
	p.styles[tsHi] = on(theme.RoleHighlight, p.bg, 4.5).Bold(true)
	p.styles[tsLane] = on(theme.RoleAccent, p.bg, 3)
	p.styles[tsGold] = on(theme.RoleHighlight, p.bg, 3).Bold(true)
	p.styles[tsGoldDim] = on(theme.RoleWarning, p.bg, 3)
	p.styles[tsGood] = on(theme.RolePositive, p.bg, 4.5)
	p.styles[tsSel] = tcell.StyleDefault.Background(p.sel).Foreground(theme.Legible(theme.Color(theme.RoleSelectionText), p.sel, 4.5)).Bold(true)
	p.styles[tsChip] = on(theme.RoleText, chip, 4.5)
	p.styles[tsChipKey] = on(theme.RoleAccent, chip, 3).Bold(true)
	p.styles[tsChipDim] = on(theme.RoleDim, chip, 3)
	p.styles[tsBorder] = on(theme.RoleBorder, p.bg, 3)
	p.styles[tsBad] = on(theme.RoleNegative, p.bg, 3).Bold(true)
	// The chip colour as ink is quiet on purpose: it is a shape to make
	// out, not text to read, so it is not pushed to the contrast rule.
	p.styles[tsSlab] = tcell.StyleDefault.Background(p.bg).Foreground(chip)
	p.styles[tsLabel] = on(theme.RoleLabel, p.bg, 4.5)
	for i := 0; i < lanes; i++ {
		p.lane = append(p.lane, tcell.StyleDefault.Background(p.bg).Foreground(theme.Hue(hue(i), p.bg)))
	}
	return p
}

// inkOn is an ink's colour on back (the canvas, or the selection).
func (p *gridPalette) inkOn(k ink, selected bool) tcell.Color {
	cache, back := p.inks[0], p.bg
	if selected {
		cache, back = p.inks[1], p.sel
	}
	if c, ok := cache[k]; ok {
		return c
	}
	c := inkValue(k, p.light)
	if p.duo {
		// One ink on one paper: keep only how much the colour stands out,
		// as more or less of the ink. That is read off its value for a
		// dark canvas, where the stop that stands out is the lightest.
		c = theme.Mix(p.bg, p.text, 0.2+0.8*theme.RelativeLuminance(inkValue(k, false)))
	}
	c = theme.Legible(c, back, 3)
	cache[k] = c
	return c
}

// style is the screen style of a cell.
func (p *gridPalette) style(c tCell) tcell.Style {
	st := p.styles[c.st]
	if c.st == tsLane && int(c.lane) >= 0 && int(c.lane) < len(p.lane) {
		st = p.lane[c.lane]
	}
	selected := c.fl&cfSel != 0
	switch {
	case c.ink != inkNone:
		back := p.bg
		if selected {
			back = p.sel
		}
		st = tcell.StyleDefault.Background(back).Foreground(p.inkOn(c.ink, selected))
	case selected && c.st != tsSel:
		fg, _, _ := st.Decompose()
		st = st.Background(p.sel).Foreground(theme.Legible(fg, p.sel, 3))
	}
	if c.fl&cfBold != 0 {
		st = st.Bold(true)
	}
	return st
}

// paintGrid puts a rendered grid on the screen in theme colours.
func paintGrid(scr tcell.Screen, g *tGrid, x0, y0 int, p *gridPalette) {
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			scr.SetContent(x0+x, y0+y, c.r, nil, p.style(c))
		}
	}
}
