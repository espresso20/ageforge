package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// expedition_grid.go is the cell grid the Expeditions panel draws into: a
// pure picture of what it shows, painted on the screen in the active
// theme's colours by one painter, so the tests read the same grid the
// player sees. A cell says how it is painted by a kind (a theme role or
// a map colour class), never by a colour value.

// xk is how a cell is painted.
type xk uint8

const (
	xText        xk = iota // body text
	xDim                   // quiet text, hints
	xFaint                 // the map's dust, separators
	xBorder                // box edges
	xTitle                 // box titles, numbers that matter
	xAccent                // keys, the tier label
	xBright                // the strongest ink: names, the highlighted pin
	xGood                  // gains, affordable, a success
	xBad                   // a refusal, a failure, a short stock
	xGold                  // the twist: the offer's flavour, in italics
	xLegend                // a Legendary pin, gold
	xLegend2               // a Legendary pin, brighter (the blink)
	xWater                 // sea
	xGrass                 // meadow
	xHill                  // hills and mountains
	xSel                   // the selection
	xSelGold               // the selection on a Legendary pin
	xLabel                 // a place name on the map
	xWork                  // people: workers
	xRes                   // people: the reserve
	xArmy                  // people: soldiers
	xLore                  // people: scholars
	xRoute                 // a route's dots and a party's road
	xDust                  // the map's terrain dust
	xPlace                 // a pin's place name, beside it
	xChipShort             // a pin as a chip: the tier colours, Short ○
	xChipLong              //   Long ◇
	xChipMil               //   Military ▲
	xChipCivic             //   Civil ■
	xChipLegend            //   Legendary ◈
	xChipGoal              // a party's goal ◆
	xChipShortHi           // the same chips in their bright phase (the selected pin pulses)
	xChipLongHi
	xChipMilHi
	xChipCivicHi
	xChipLegendHi
	xChipGoalHi
	xCount
)

// chipTone is which pin chip an offer or a goal is drawn as.
type chipTone uint8

const (
	toneShort chipTone = iota
	toneLong
	toneMil
	toneCivic
	toneLegend
	toneGoal
)

// chipKind is the chip for a tone; hi is its bright phase.
func chipKind(t chipTone, hi bool) xk {
	k := xChipShort + xk(t)
	if hi {
		k = xChipShortHi + xk(t)
	}
	return k
}

// xcell is one cell, painted by an xk.
type xcell = gcell[xk]

// xgrid is a block of cells.
type xgrid struct {
	glyphGrid[xk]
}

func newXGrid(w, h int) *xgrid {
	return &xgrid{newGlyphGrid[xk](w, h)}
}

// put writes s from (x, y), one cell a rune, and returns the next x.
func (g *xgrid) put(x, y int, s string, k xk) int {
	for _, r := range s {
		g.set(x, y, r, k)
		x++
	}
	return x
}

// putIt is put in italics.
func (g *xgrid) putIt(x, y int, s string, k xk) int {
	for _, r := range s {
		g.store(x, y, xcell{r: r, k: k, lane: -1, fl: cfItalic})
		x++
	}
	return x
}

// putFit writes s in at most room cells, ending in an ellipsis if it is cut.
func (g *xgrid) putFit(x, y, room int, s string, k xk) int {
	return g.put(x, y, fitText(s, room), k)
}

// rput writes s so that it ends at column xr.
func (g *xgrid) rput(xr, y int, s string, k xk) {
	g.put(xr-len([]rune(s))+1, y, s, k)
}

// fill clears a block.
func (g *xgrid) fill(x, y, w, h int) {
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			g.set(x+i, y+j, ' ', xText)
		}
	}
}

// box draws a frame with a title on its top edge.
func (g *xgrid) box(x, y, w, h int, k xk, title string, tk xk, double bool) {
	if w < 2 || h < 2 {
		return
	}
	b := []rune("┌┐└┘─│")
	if double {
		b = []rune("╔╗╚╝═║")
	}
	g.set(x, y, b[0], k)
	g.set(x+w-1, y, b[1], k)
	g.set(x, y+h-1, b[2], k)
	g.set(x+w-1, y+h-1, b[3], k)
	for i := 1; i < w-1; i++ {
		g.set(x+i, y, b[4], k)
		g.set(x+i, y+h-1, b[4], k)
	}
	for j := 1; j < h-1; j++ {
		g.set(x, y+j, b[5], k)
		g.set(x+w-1, y+j, b[5], k)
	}
	if title != "" {
		g.put(x+2, y, " "+fitText(title, w-6)+" ", tk)
	}
}

// seg is a run of text in one kind.
type xseg struct {
	s string
	k xk
}

// segs writes runs one after another and returns the next x.
func (g *xgrid) segs(x, y int, parts ...xseg) int {
	for _, p := range parts {
		x = g.put(x, y, p.s, p.k)
	}
	return x
}

// segsFit is segs cut to room cells.
func (g *xgrid) segsFit(x, y, room int, parts ...xseg) int {
	end := x + room
	for _, p := range parts {
		if x >= end {
			break
		}
		x = g.putFit(x, y, end-x, p.s, p.k)
	}
	return x
}

// fitText cuts s to room cells, ending in an ellipsis if it is cut.
func fitText(s string, room int) string {
	r := []rune(s)
	if room <= 0 {
		return ""
	}
	if len(r) <= room {
		return s
	}
	if room == 1 {
		return "…"
	}
	return string(r[:room-1]) + "…"
}

// wrapText breaks s into lines of at most w cells, at spaces.
func wrapText(s string, w int) []string {
	var out []string
	line := ""
	for _, word := range strings.Fields(s) {
		switch {
		case line == "":
			line = word
		case len([]rune(line))+1+len([]rune(word)) > w:
			out = append(out, line)
			line = word
		default:
			line += " " + word
		}
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

// foldExpedition is the plain glyph tier's fold: the pins and the bars
// keep a difference between them in ASCII, and the rest folds as the maps
// fold.
var foldExpedition = map[rune]rune{
	'○': 'o', '◇': 'd', '▲': '^', '■': '#', '◈': '*', '◆': '+', '▣': 'T', '†': '+',
	'▰': '#', '▱': '-', '█': '#', '░': '-', '·': '.', '•': '*', '≈': '~', '∩': 'n', '●': '*',
	'←': '<', '→': '>', '↑': '^', '↓': 'v', '…': '~', '▸': '>',
}

func (g *xgrid) fold() { g.foldWith(foldExpedition) }

// expPalette is the active theme's colours for the panel.
type expPalette struct {
	styles [xCount]tcell.Style
}

func newExpPalette() *expPalette {
	bg := theme.Color(theme.RoleBackground)
	mp := mapstyle.NewPalette(0)
	on := func(c tcell.Color, ratio float64) tcell.Style {
		return tcell.StyleDefault.Background(bg).Foreground(theme.Legible(c, bg, ratio))
	}
	role := func(r theme.Role, ratio float64) tcell.Style { return on(theme.Color(r), ratio) }
	p := &expPalette{}
	s := &p.styles
	s[xText] = role(theme.RoleText, 4.5)
	s[xDim] = role(theme.RoleDim, 3)
	s[xFaint] = on(theme.Mix(bg, theme.Color(theme.RoleDim), 0.55), 1.6)
	s[xBorder] = role(theme.RoleBorder, 3)
	s[xTitle] = role(theme.RoleHighlight, 4.5).Bold(true)
	s[xAccent] = role(theme.RoleAccent, 3).Bold(true)
	s[xBright] = role(theme.RoleBright, 4.5).Bold(true)
	s[xGood] = role(theme.RolePositive, 4.5)
	s[xBad] = role(theme.RoleNegative, 4.5)
	s[xGold] = role(theme.RoleHighlight, 4.5)
	gold := theme.Mix(theme.Color(theme.RoleHighlight), theme.MapHueColor(theme.HueGold), 0.6)
	s[xLegend] = on(gold, 3).Bold(true)
	s[xLegend2] = on(theme.Mix(gold, theme.Color(theme.RoleBright), 0.5), 3).Bold(true)
	s[xWater] = tcell.StyleDefault.Background(bg).Foreground(mp.Fg[mapmodel.CWater])
	s[xGrass] = tcell.StyleDefault.Background(bg).Foreground(mp.Fg[mapmodel.CGround])
	s[xHill] = tcell.StyleDefault.Background(bg).Foreground(mp.Fg[mapmodel.CHill])
	sel := theme.Color(theme.RoleSelection)
	s[xSel] = tcell.StyleDefault.Background(sel).Foreground(theme.Legible(theme.Color(theme.RoleSelectionText), sel, 4.5)).Bold(true)
	s[xSelGold] = tcell.StyleDefault.Background(gold).Foreground(theme.Legible(bg, gold, 4.5)).Bold(true)
	chip := theme.Color(theme.RoleChip)
	s[xWork] = role(theme.RoleHighlight, 3)
	s[xRes] = role(theme.RolePositive, 3)
	s[xArmy] = role(theme.RoleNegative, 3)
	s[xLore] = role(theme.RoleLabel, 3)
	s[xLabel] = tcell.StyleDefault.Background(chip).Foreground(theme.Legible(theme.Color(theme.RoleHighlight), chip, 4.5))
	s[xRoute] = role(theme.RoleAccent, 4.5)
	s[xDust] = on(theme.Mix(bg, theme.Color(theme.RoleDim), 0.55), mapstyle.TerrainFloor)
	s[xPlace] = role(theme.RoleDim, 4.5)
	// The pins: each tone is a chip, its glyph in the ink that reads on it.
	// A chip's bright phase is the same chip lifted toward the bright role.
	tones := [6]struct {
		bg  tcell.Color
		ink tcell.Color
	}{
		{theme.Color(theme.RolePositive), theme.Color(theme.RoleBackground)},
		{theme.Color(theme.RoleLabel), theme.Color(theme.RoleBackground)},
		{theme.Color(theme.RoleNegative), theme.Color(theme.RoleOnNegative)},
		{theme.Color(theme.RoleAccent), theme.Color(theme.RoleOnAccent)},
		{gold, theme.Color(theme.RoleBackground)},
		{theme.Color(theme.RolePositive), theme.Color(theme.RoleBackground)},
	}
	for t, c := range tones {
		chipOn := func(bg tcell.Color) tcell.Style {
			ink := theme.Legible(c.ink, bg, 4.5)
			if theme.ContrastRatio(ink, bg) < 4.5 { // a mid tone: black or white reads
				ink = theme.BestOn(bg)
			}
			return tcell.StyleDefault.Background(bg).Foreground(ink).Bold(true)
		}
		s[chipKind(chipTone(t), false)] = chipOn(c.bg)
		s[chipKind(chipTone(t), true)] = chipOn(theme.Mix(c.bg, theme.Color(theme.RoleBright), 0.4))
	}
	return p
}

// style is the screen style of a cell.
func (p *expPalette) style(c xcell) tcell.Style {
	st := p.styles[c.k]
	if c.fl&cfItalic != 0 {
		st = st.Italic(true)
	}
	return st
}

// paintExpedition puts a grid on the screen at (x0, y0).
func paintExpedition(scr tcell.Screen, g *xgrid, x0, y0 int, p *expPalette) {
	g.draw(scr, x0, y0, func(c xcell) (rune, tcell.Style) { return c.r, p.style(c) })
}
