package ui

import (
	"github.com/gdamore/tcell/v2"

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
	tsEffect                // what a tech does: its effect sentence, bold in the effect colour
	tsCount                 // how many styles there are
)

// tCell is one cell of a panel, painted by a tStyle.
type tCell = gcell[tStyle]

// tGrid is a block of cells.
type tGrid struct {
	glyphGrid[tStyle]
}

func newTGrid(w, h int) *tGrid {
	return &tGrid{newGlyphGrid[tStyle](w, h)}
}

func (g *tGrid) put(x, y int, r rune, st tStyle, lane int) {
	g.store(x, y, tCell{r: r, k: st, lane: int8(lane)})
}

// paint writes a cell in an ink, with flags. The style is what the cell
// falls back to where an ink has no say (a test that reads styles).
func (g *tGrid) paint(x, y int, r rune, k ink, st tStyle, fl uint8) {
	g.store(x, y, tCell{r: r, k: st, lane: -1, ink: k, fl: fl})
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

// card draws a rounded frame cw by ch at (x0, y0), blank inside, with name
// on its top edge and, when the edge has the room, esc (the hint that Esc
// closes it) at the right.
func (g *tGrid) card(x0, y0, cw, ch int, name, esc string) {
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			r := ' '
			switch {
			case y == 0 && x == 0:
				r = '╭'
			case y == 0 && x == cw-1:
				r = '╮'
			case y == ch-1 && x == 0:
				r = '╰'
			case y == ch-1 && x == cw-1:
				r = '╯'
			case y == 0 || y == ch-1:
				r = '─'
			case x == 0 || x == cw-1:
				r = '│'
			}
			g.put(x0+x, y0+y, r, tsBorder, -1)
		}
	}
	g.text(x0+2, y0, clipRunes(" "+name+" ", cw-4), tsBright, -1)
	if cw > runeLen(esc)+runeLen(name)+8 {
		g.text(x0+cw-2-runeLen(esc), y0, esc, tsDim, -1)
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
	p.styles[tsEffect] = on(theme.RoleEffect, p.bg, 4.5).Bold(true)
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
	st := p.styles[c.k]
	if c.k == tsLane && int(c.lane) >= 0 && int(c.lane) < len(p.lane) {
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
	case selected && c.k != tsSel:
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
	g.draw(scr, x0, y0, func(c tCell) (rune, tcell.Style) { return c.r, p.style(c) })
}
