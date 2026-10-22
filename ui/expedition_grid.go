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
	xText    xk = iota // body text
	xDim               // quiet text, hints
	xFaint             // the map's dust, separators
	xBorder            // box edges
	xTitle             // box titles, numbers that matter
	xAccent            // keys, the tier label
	xBright            // the strongest ink: names, the highlighted pin
	xGood              // gains, affordable, a success
	xBad               // a refusal, a failure, a short stock
	xGold              // the twist: the offer's flavour, in italics
	xLegend            // a Legendary pin, gold
	xLegend2           // a Legendary pin, brighter (the blink)
	xWater             // sea
	xGrass             // meadow
	xHill              // hills and mountains
	xSel               // the selection
	xSelGold           // the selection on a Legendary pin
	xLabel             // a place name on the map
	xWork              // people: workers
	xRes               // people: the reserve
	xArmy              // people: soldiers
	xLore              // people: scholars
	xCount
)

// xcell is one cell.
type xcell struct {
	r  rune
	k  xk
	it bool // drawn in italics
}

// xgrid is a block of cells.
type xgrid struct {
	w, h int
	c    []xcell
}

func newXGrid(w, h int) *xgrid {
	g := &xgrid{w: max(w, 0), h: max(h, 0)}
	g.c = make([]xcell, g.w*g.h)
	for i := range g.c {
		g.c[i] = xcell{r: ' '}
	}
	return g
}

func (g *xgrid) in(x, y int) bool { return x >= 0 && y >= 0 && x < g.w && y < g.h }

func (g *xgrid) at(x, y int) xcell {
	if !g.in(x, y) {
		return xcell{r: ' '}
	}
	return g.c[y*g.w+x]
}

func (g *xgrid) set(x, y int, r rune, k xk) {
	if g.in(x, y) {
		g.c[y*g.w+x] = xcell{r: r, k: k}
	}
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
		if g.in(x, y) {
			g.c[y*g.w+x] = xcell{r: r, k: k, it: true}
		}
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

// String is the grid as lines of text, trailing spaces kept.
func (g *xgrid) String() string {
	var sb strings.Builder
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			sb.WriteRune(g.c[y*g.w+x].r)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Lines are the rows as text, trailing spaces trimmed.
func (g *xgrid) Lines() []string {
	lines := strings.Split(strings.TrimRight(g.String(), "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return lines
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

func (g *xgrid) fold() {
	for i, c := range g.c {
		if c.r < 0x80 {
			continue
		}
		if r, ok := foldExpedition[c.r]; ok {
			g.c[i].r = r
		} else {
			g.c[i].r = mapmodel.Fold(c.r, mapmodel.TierASCII)
		}
	}
}

// expPalette is the active theme's colours for the panel.
type expPalette struct {
	styles [xCount]tcell.Style
	italic tcell.Style
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
	return p
}

// paintExpedition puts a grid on the screen at (x0, y0).
func paintExpedition(scr tcell.Screen, g *xgrid, x0, y0 int, p *expPalette) {
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			st := p.styles[c.k]
			if c.it {
				st = st.Italic(true)
			}
			scr.SetContent(x0+x, y0+y, c.r, nil, st)
		}
	}
}
