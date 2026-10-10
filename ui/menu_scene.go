package ui

import (
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// menu_scene.go draws what moves on the main menu: the AGEFORGE wordmark as
// iron on the fire, the forge (a fire with an anvil standing in front of
// it), the sparks, and the night sky over the town. menu.go lays the page
// out and menu_town.go draws the player's own map behind it.
//
// Everything is drawn into a grid of cells first (mGrid) and put on the
// screen in one pass, so a page is a value a test can read. No colour is
// chosen here: every one is derived from the active theme's roles
// (menuPalette), so the fire is red and gold on Forge, green on Source and
// ink on paper on a light theme.
//
// The fire and the sparks carry state from frame to frame, so unlike the
// badge case a frame is not a function of its number alone. They draw their
// chance from a generator of their own with a fixed seed: the same number
// of frames from the same layout is always the same picture, which is what
// the render tests and the wiki's pictures rely on.

// mCell is one cell of the menu.
type mCell struct {
	r     rune
	fg    tcell.Color
	bg    tcell.Color
	bold  bool
	hasBg bool // bg is set; otherwise the cell stands on the page's ground
}

// mGrid is the page as cells.
type mGrid struct {
	w, h int
	c    []mCell
	// clipped lists text that was cut at an edge. A layout that fits writes
	// nothing here; the render tests fail on anything in it.
	clipped []string
}

func newMGrid(w, h int) *mGrid {
	g := &mGrid{w: max(w, 0), h: max(h, 0)}
	g.c = make([]mCell, g.w*g.h)
	for i := range g.c {
		g.c[i].r = ' '
	}
	return g
}

func (g *mGrid) in(x, y int) bool { return x >= 0 && y >= 0 && x < g.w && y < g.h }

// put writes a glyph and its ink, and leaves the cell's ground as it is.
func (g *mGrid) put(x, y int, r rune, fg tcell.Color) {
	if g.in(x, y) {
		c := &g.c[y*g.w+x]
		c.r, c.fg, c.bold = r, fg, false
	}
}

// set writes a whole cell: glyph, ink, ground and weight.
func (g *mGrid) set(x, y int, r rune, fg, bg tcell.Color, bold bool) {
	if g.in(x, y) {
		g.c[y*g.w+x] = mCell{r: r, fg: fg, bg: bg, bold: bold, hasBg: true}
	}
}

// fill paints a block's ground and clears what was on it.
func (g *mGrid) fill(x, y, w, h int, fg, bg tcell.Color) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			g.set(i, j, ' ', fg, bg, false)
		}
	}
}

// text writes s from (x, y) on the ground already there and returns its
// length in cells. Text that runs off the grid is cut and noted.
func (g *mGrid) text(x, y int, s string, fg tcell.Color, bold bool) int {
	n := 0
	for _, r := range s {
		if g.in(x+n, y) {
			c := &g.c[y*g.w+x+n]
			c.r, c.fg, c.bold = r, fg, bold
		} else if len(g.clipped) == 0 || g.clipped[len(g.clipped)-1] != s {
			g.clipped = append(g.clipped, s)
		}
		n++
	}
	return n
}

// textOn is text on a ground of its own.
func (g *mGrid) textOn(x, y int, s string, fg, bg tcell.Color, bold bool) int {
	n := 0
	for _, r := range s {
		if g.in(x+n, y) {
			g.set(x+n, y, r, fg, bg, bold)
		} else if len(g.clipped) == 0 || g.clipped[len(g.clipped)-1] != s {
			g.clipped = append(g.clipped, s)
		}
		n++
	}
	return n
}

// box draws a frame, heavy or light, and leaves its inside alone.
func (g *mGrid) box(x, y, w, h int, heavy bool, fg tcell.Color) {
	if w < 2 || h < 2 {
		return
	}
	c := []rune("┌┐└┘─│")
	if heavy {
		c = []rune("┏┓┗┛━┃")
	}
	for i := 1; i < w-1; i++ {
		g.put(x+i, y, c[4], fg)
		g.put(x+i, y+h-1, c[4], fg)
	}
	for j := 1; j < h-1; j++ {
		g.put(x, y+j, c[5], fg)
		g.put(x+w-1, y+j, c[5], fg)
	}
	g.put(x, y, c[0], fg)
	g.put(x+w-1, y, c[1], fg)
	g.put(x, y+h-1, c[2], fg)
	g.put(x+w-1, y+h-1, c[3], fg)
}

// row is one row of the grid as text.
func (g *mGrid) row(y int) string {
	if y < 0 || y >= g.h {
		return ""
	}
	out := make([]rune, g.w)
	for x := range out {
		out[x] = g.c[y*g.w+x].r
	}
	return string(out)
}

// flush puts the grid on the screen with its top left corner at (x0, y0).
// In the plain glyph set every glyph that is not ASCII is folded to one
// that is (the wordmark and the fire were already drawn without blocks).
func (g *mGrid) flush(scr tcell.Screen, x0, y0 int, pal *menuPalette, plain bool) {
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			bg := pal.bg
			if c.hasBg {
				bg = c.bg
			}
			r := c.r
			if plain {
				r = mapmodel.Fold(r, mapmodel.TierASCII)
			}
			scr.SetContent(x0+x, y0+y, r, nil, tcell.StyleDefault.Foreground(c.fg).Background(bg).Bold(c.bold))
		}
	}
}

// ---- colours, all from the theme's roles ----

// menuLUT is how many steps the heat and fire ramps are drawn in.
const menuLUT = 64

// menuPalette is the menu's colours on one theme.
type menuPalette struct {
	key string
	// bg is the page; ink, dim and faint are text from strongest to
	// weakest; accent marks the selection and the headings.
	bg, ink, dim, faint, accent tcell.Color
	// band and onBand are the selected row of the boxed menu.
	band, onBand tcell.Color
	// ground is the boxed menu's own ground, with its inks.
	ground, groundInk, groundDim tcell.Color
	// label is the second colour of the elite line.
	label tcell.Color
	// anvil is the anvil's body, and anvilInk its ink where a glyph draws
	// it (the plain glyph set).
	anvil, anvilInk tcell.Color
	// heat is the wordmark's ramp from cold iron to white heat; fire the
	// forge's, from the page to its brightest.
	heat, fire [menuLUT]tcell.Color
	// heatStops and fireStops are the ramps' six stops (the tests read them).
	heatStops, fireStops [6]tcell.Color
}

// colourHue is a colour's hue in degrees and its saturation from 0 to 1.
func colourHue(c tcell.Color) (hue, sat float64) {
	r, g, b := c.RGB()
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	hi, lo := max(rf, gf, bf), min(rf, gf, bf)
	d := hi - lo
	if d == 0 {
		return 0, 0
	}
	switch hi {
	case rf:
		hue = math.Mod((gf-bf)/d, 6)
	case gf:
		hue = (bf-rf)/d + 2
	default:
		hue = (rf-gf)/d + 4
	}
	hue *= 60
	if hue < 0 {
		hue += 360
	}
	return hue, d / hi
}

// warmColour reports whether c is a colour of fire: red through gold.
func warmColour(c tcell.Color) bool {
	h, s := colourHue(c)
	return s >= 0.25 && (h <= 75 || h >= 345)
}

// menuColdIron is the least the wordmark's coldest iron stands out from
// the page: enough that a letter gone cold is still a letter.
const menuColdIron = 1.6

// menuRampStops works out the six stops of a theme's heat ramp (the
// wordmark) and fire ramp (the forge).
//
// A ramp runs from the page to the theme's strongest ink through the
// theme's own colours, so it stands out more with every step whichever way
// the theme runs: brighter on a dark theme, darker on a light one, where
// hot iron is ink on paper. A theme whose accent is a colour of fire burns
// from embers to gold (its Negative warmed with its Accent, then its Accent
// or Highlight). Any other burns in its own two colours (Accent and Label):
// green on Source, magenta to cyan on Glitch. A theme of one ink runs
// straight from paper to ink.
func menuRampStops(th theme.Theme) (heat, fire [6]tcell.Color) {
	bg, ink, bright := th.Color(theme.RoleBackground), th.Color(theme.RoleText), th.Color(theme.RoleBright)
	if th.Duotone {
		for i, k := range []float64{0.24, 0.40, 0.56, 0.72, 0.86, 1} {
			heat[i] = theme.Mix(bg, ink, k)
		}
		heat[0] = theme.Legible(heat[0], bg, menuColdIron)
		fire = heat
		fire[0], fire[1] = bg, theme.Mix(bg, ink, 0.22)
		return heat, fire
	}
	cold := theme.Legible(theme.Mix(bg, th.Color(theme.RoleDim), 0.30), bg, menuColdIron)
	accent := th.Color(theme.RoleAccent)
	if warmColour(accent) {
		// A fire: embers are the theme's red warmed with its accent (a
		// burnt orange on Forge), and the flame climbs to the stronger of
		// its accent and its highlight.
		ember, hi := theme.Mix(th.Color(theme.RoleNegative), accent, 0.25), accent
		if hl := th.Color(theme.RoleHighlight); warmColour(hl) && theme.ContrastRatio(hl, bg) > theme.ContrastRatio(hi, bg) {
			hi = hl
		}
		mid, top := theme.Mix(ember, hi, 0.5), theme.Mix(hi, bright, 0.75)
		heat = [6]tcell.Color{cold, theme.Mix(bg, ember, 0.42), theme.Mix(bg, ember, 0.78), mid, hi, top}
		fire = [6]tcell.Color{bg, theme.Mix(bg, ember, 0.30), theme.Mix(bg, ember, 0.70), mid, hi, top}
	} else {
		lo, hi := accent, th.Color(theme.RoleLabel)
		if theme.ContrastRatio(lo, bg) > theme.ContrastRatio(hi, bg) {
			lo, hi = hi, lo
		}
		mid, top := theme.Mix(lo, hi, 0.5), theme.Mix(hi, bright, 0.75)
		heat = [6]tcell.Color{cold, theme.Mix(bg, lo, 0.5), lo, mid, hi, top}
		fire = [6]tcell.Color{bg, theme.Mix(bg, lo, 0.30), theme.Mix(bg, lo, 0.70), mid, hi, top}
	}
	// Hotter always stands out more than cooler: a stop that does not is
	// pushed toward the theme's ink until it does.
	for _, ramp := range []*[6]tcell.Color{&heat, &fire} {
		for i := 1; i < len(ramp); i++ {
			if want := theme.ContrastRatio(ramp[i-1], bg) * 1.06; theme.ContrastRatio(ramp[i], bg) < want {
				ramp[i] = theme.Legible(ramp[i], bg, want)
			}
		}
	}
	return heat, fire
}

// menuRamp spreads six stops over the ramp's steps.
func menuRamp(stops [6]tcell.Color) (lut [menuLUT]tcell.Color) {
	for i := range lut {
		v := float64(i) / float64(menuLUT-1) * float64(len(stops)-1)
		j := min(int(v), len(stops)-2)
		lut[i] = theme.Mix(stops[j], stops[j+1], v-float64(j))
	}
	return lut
}

// rampAt picks a ramp's colour for v from 0 to 1.
func rampAt(lut *[menuLUT]tcell.Color, v float64) tcell.Color {
	switch {
	case v <= 0 || v != v:
		return lut[0]
	case v >= 1:
		return lut[menuLUT-1]
	}
	return lut[int(v*float64(menuLUT-1))]
}

// newMenuPalette builds the menu's colours for a theme.
func newMenuPalette(th theme.Theme) *menuPalette {
	bg, text := th.Color(theme.RoleBackground), th.Color(theme.RoleText)
	p := &menuPalette{key: th.Key, bg: bg, ink: text}
	p.dim = theme.Legible(th.Color(theme.RoleDim), bg, 3)
	p.faint = theme.Legible(theme.Mix(bg, th.Color(theme.RoleDim), 0.42), bg, 1.5)
	p.accent = theme.Legible(th.Color(theme.RoleAccent), bg, 3)
	p.band = th.Color(theme.RoleAccent)
	p.onBand = theme.Legible(th.Color(theme.RoleOnAccent), p.band, 4.5)
	// The boxed menu stands a shade off the page: deeper on a dark theme,
	// paler on a light one.
	p.ground = theme.Mix(bg, theme.BestOn(text), 0.25)
	if p.ground == bg {
		// A page already as dark or as pale as a colour gets (the two
		// high contrast themes): the box takes a step toward the ink.
		p.ground = theme.Mix(bg, text, 0.08)
	}
	p.groundInk = theme.Legible(text, p.ground, 7)
	p.groundDim = theme.Legible(th.Color(theme.RoleDim), p.ground, 3)
	p.label = theme.Legible(th.Color(theme.RoleLabel), bg, 3)
	p.anvil = theme.Mix(bg, text, 0.10)
	p.anvilInk = theme.Mix(bg, text, 0.42)
	p.heatStops, p.fireStops = menuRampStops(th)
	p.heat, p.fire = menuRamp(p.heatStops), menuRamp(p.fireStops)
	return p
}

// ---- chance ----

// menuRand is the scene's own generator: small, fast and the same on every
// machine.
type menuRand struct{ s uint64 }

func (r *menuRand) next() uint64 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 7
	r.s ^= r.s << 17
	return r.s
}

// f is a number from 0 up to, not including, 1.
func (r *menuRand) f() float64 { return float64(r.next()>>11) / float64(1<<53) }

// ---- the wordmark ----

// menuFont is the pixel alphabet the wordmark and the arrival screens are
// set in: seven pixels tall, strokes two pixels thick, corners cut by one.
// A letter is six pixels wide; I is four, and M and W, which need a middle,
// are seven. The six letters of AGEFORGE came first (the menu's wordmark)
// and the rest are drawn in their hand. It covers A to Z, the hyphen and
// the space, which is every age's and every era's name and any that may
// follow.
var menuFont = map[rune][7]string{
	'A': {".####.", "##..##", "##..##", "######", "##..##", "##..##", "##..##"},
	'B': {"#####.", "##..##", "##..##", "#####.", "##..##", "##..##", "#####."},
	'C': {".#####", "##....", "##....", "##....", "##....", "##....", ".#####"},
	'D': {"#####.", "##..##", "##..##", "##..##", "##..##", "##..##", "#####."},
	'E': {"######", "##....", "##....", "#####.", "##....", "##....", "######"},
	'F': {"######", "##....", "##....", "#####.", "##....", "##....", "##...."},
	'G': {".#####", "##....", "##....", "##.###", "##..##", "##..##", ".#####"},
	'H': {"##..##", "##..##", "##..##", "######", "##..##", "##..##", "##..##"},
	'I': {"####", ".##.", ".##.", ".##.", ".##.", ".##.", "####"},
	'J': {"...###", "....##", "....##", "....##", "##..##", "##..##", ".####."},
	'K': {"##..##", "##.##.", "####..", "###...", "####..", "##.##.", "##..##"},
	'L': {"##....", "##....", "##....", "##....", "##....", "##....", "######"},
	'M': {"##...##", "###.###", "#######", "##.#.##", "##...##", "##...##", "##...##"},
	'N': {"##..##", "###.##", "######", "##.###", "##..##", "##..##", "##..##"},
	'O': {".####.", "##..##", "##..##", "##..##", "##..##", "##..##", ".####."},
	'P': {"#####.", "##..##", "##..##", "#####.", "##....", "##....", "##...."},
	'Q': {".####.", "##..##", "##..##", "##..##", "##.###", "##..##", ".#####"},
	'R': {"#####.", "##..##", "##..##", "#####.", "##.##.", "##..##", "##..##"},
	'S': {".#####", "##....", "##....", ".####.", "....##", "....##", "#####."},
	'T': {"######", "..##..", "..##..", "..##..", "..##..", "..##..", "..##.."},
	'U': {"##..##", "##..##", "##..##", "##..##", "##..##", "##..##", ".####."},
	'V': {"##..##", "##..##", "##..##", "##..##", "##..##", ".####.", "..##.."},
	'W': {"##...##", "##...##", "##...##", "##.#.##", "#######", "###.###", "##...##"},
	'X': {"##..##", "##..##", ".####.", "..##..", ".####.", "##..##", "##..##"},
	'Y': {"##..##", "##..##", "##..##", ".####.", "..##..", "..##..", "..##.."},
	'Z': {"######", "....##", "...##.", "..##..", ".##...", "##....", "######"},
	'-': {"....", "....", "....", "####", "....", "....", "...."},
	' ': {"...", "...", "...", "...", "...", "...", "..."},
}

// pixelWord sets text in the pixel alphabet, one pixel between letters,
// and reports whether every character of it has a letter there. Lower
// case is set as upper.
func pixelWord(text string) (rows [menuWordH]string, ok bool) {
	first := true
	for _, r := range strings.ToUpper(text) {
		glyph, has := menuFont[r]
		if !has {
			return rows, false
		}
		for y := range rows {
			if !first {
				rows[y] += "."
			}
			rows[y] += glyph[y]
		}
		first = false
	}
	return rows, !first
}

// ironMode is how the pixels of a word in iron are put on cells: sx cells
// wide and sy rows tall each, or, with half set, two pixels to a cell (each
// half a cell wide and sy rows tall), for a long name on a narrow screen.
type ironMode struct {
	sx, sy int
	half   bool
}

// size is how many cells a word px pixels wide takes in the mode.
func (m ironMode) size(px int) (w, h int) {
	if m.half {
		return (px + 1) / 2, menuWordH * m.sy
	}
	return px * m.sx, menuWordH * m.sy
}

// ironHeat is how hot pixel (px, py) of a word w pixels wide is at time t:
// wordHeat for a word of any width.
func ironHeat(px, py, w int, t, strike, uc float64) float64 {
	u, v := 0.5, float64(py)/float64(menuWordH-1)
	if w > 1 {
		u = float64(px) / float64(w-1)
	}
	grain := mapmodel.HashF(int64(px), int64(py), int64(math.Floor(t*6)))
	return 0.28 + 0.2*math.Sin(t*0.9-u*5) + 0.14*math.Sin(t*0.37+u*11) + 0.24*v + 0.05*(2*grain-1) + strike*(1-0.55*math.Abs(u-uc))
}

// drawIron draws a word set by pixelWord with its top left corner at
// (x0, y0), each pixel as hot as heat says. In the plain glyph set a pixel
// is a # on a ground of nearly its own colour, in place of a block, and
// there are no half cells: the caller picks a whole-cell mode.
func drawIron(g *mGrid, pal *menuPalette, rows [menuWordH]string, x0, y0 int, mode ironMode, heat func(px, py int) float64, plain bool) {
	w := len(rows[0])
	lit := func(px, py int) bool { return px < w && rows[py][px] == '#' }
	for py := 0; py < menuWordH; py++ {
		if mode.half {
			for cx := 0; cx*2 < w; cx++ {
				l, r := lit(cx*2, py), lit(cx*2+1, py)
				if !l && !r {
					continue
				}
				for dy := 0; dy < mode.sy; dy++ {
					x, y := x0+cx, y0+py*mode.sy+dy
					switch {
					case l && r:
						g.set(x, y, '▌', rampAt(&pal.heat, heat(cx*2, py)), rampAt(&pal.heat, heat(cx*2+1, py)), false)
					case l:
						g.put(x, y, '▌', rampAt(&pal.heat, heat(cx*2, py)))
					default:
						g.put(x, y, '▐', rampAt(&pal.heat, heat(cx*2+1, py)))
					}
				}
			}
			continue
		}
		for px := 0; px < w; px++ {
			if !lit(px, py) {
				continue
			}
			col := rampAt(&pal.heat, heat(px, py))
			for dy := 0; dy < mode.sy; dy++ {
				for dx := 0; dx < mode.sx; dx++ {
					x, y := x0+px*mode.sx+dx, y0+py*mode.sy+dy
					if plain {
						// A ground of the letter's ink, a little back from
						// it so the mark shows, and never too far back to
						// read as a block on the page.
						g.set(x, y, '#', col, theme.Legible(theme.Mix(col, pal.bg, 0.3), pal.bg, 3), true)
					} else {
						g.put(x, y, '█', col)
					}
				}
			}
		}
	}
}

// menuWord is AGEFORGE in the font, one pixel between letters.
var menuWord = func() (rows [7]string) {
	for i, l := range "AGEFORGE" {
		for r := 0; r < 7; r++ {
			if i > 0 {
				rows[r] += "."
			}
			rows[r] += menuFont[l][r]
		}
	}
	return rows
}()

// menuWordW and menuWordH are the wordmark's size in pixels.
const (
	menuWordW = 55
	menuWordH = 7
)

// wordHeat is how hot the wordmark's pixel (px, py) is at time t: two slow
// waves along the word, hotter toward the foot of each letter, a grain
// that changes six times a second, and the strike, which is hottest at uc
// (a place along the word from 0 to 1).
func wordHeat(px, py int, t, strike, uc float64) float64 {
	u, v := float64(px)/float64(menuWordW-1), float64(py)/float64(menuWordH-1)
	grain := mapmodel.HashF(int64(px), int64(py), int64(math.Floor(t*6)))
	return 0.28 + 0.2*math.Sin(t*0.9-u*5) + 0.14*math.Sin(t*0.37+u*11) + 0.24*v + 0.05*(2*grain-1) + strike*(1-0.55*math.Abs(u-uc))
}

// drawWordmark draws the wordmark with its top left corner at (x0, y0),
// each pixel sx cells wide. In the plain glyph set a pixel is a # on a
// ground of nearly its own colour, in place of a block.
func drawWordmark(g *mGrid, pal *menuPalette, x0, y0, sx int, t, strike, uc float64, plain bool) {
	for py := 0; py < menuWordH; py++ {
		for px := 0; px < menuWordW; px++ {
			if menuWord[py][px] != '#' {
				continue
			}
			col := rampAt(&pal.heat, wordHeat(px, py, t, strike, uc))
			for d := 0; d < sx; d++ {
				if plain {
					g.set(x0+px*sx+d, y0+py, '#', col, theme.Mix(col, pal.bg, 0.3), true)
				} else {
					g.put(x0+px*sx+d, y0+py, '█', col)
				}
			}
		}
	}
}

// ---- the sparks ----

// menuSub is how many steps of the scene one animation frame takes. The
// scene was tuned at about fourteen steps a second; two to a frame at the
// map's rate is sixteen, and menuPace slows each step to match.
const (
	menuSub  = 2
	menuPace = 0.88
)

type menuSpark struct{ x, y, vx, vy, life, max float64 }

type menuSparks struct{ a []menuSpark }

// emit throws n sparks from about (x, y): spread is how far apart they
// start and lift how fast the quickest rises.
func (s *menuSparks) emit(rnd *menuRand, x, y float64, n int, spread, lift float64) {
	for i := 0; i < n; i++ {
		life := (14 + rnd.f()*34) / menuPace
		s.a = append(s.a, menuSpark{
			x: x + (rnd.f()-0.5)*spread, y: y,
			vx: (rnd.f() - 0.5) * 0.7 * menuPace, vy: -(0.18 + rnd.f()*lift) * menuPace,
			life: life, max: life,
		})
	}
	if over := len(s.a) - 400; over > 0 {
		s.a = s.a[over:]
	}
}

func (s *menuSparks) step(rnd *menuRand) {
	live := s.a[:0]
	for _, sp := range s.a {
		sp.x += sp.vx
		sp.y += sp.vy
		sp.vx += (rnd.f() - 0.5) * 0.08 * menuPace
		sp.vy *= 0.987
		sp.life--
		if sp.life > 0 && sp.y >= 0 {
			live = append(live, sp)
		}
	}
	s.a = live
}

func (s *menuSparks) draw(g *mGrid, pal *menuPalette) {
	for _, sp := range s.a {
		k := sp.life / sp.max
		r := '·'
		switch {
		case k > 0.7:
			r = '*'
		case k > 0.35:
			r = '∙'
		}
		g.put(int(math.Round(sp.x)), int(math.Round(sp.y)), r, rampAt(&pal.fire, 0.35+0.65*k))
	}
}

// ---- the forge ----

// menuAnvil is the anvil, a pixel a character. Its top row is the face,
// which glows when it is struck.
var menuAnvil = [12]string{
	"...###################",
	".#####################",
	"#####################.",
	"..##################..",
	"......############....",
	"........########......",
	"........########......",
	".......##########.....",
	"......############....",
	".....##############...",
	"....################..",
	"....################..",
}

// menuForge is the fire and the anvil in front of it: a plate fw cells
// wide and fh half cells tall.
type menuForge struct {
	fw, fh int
	fire   []float64
	mask   []uint8 // 0 fire, 1 the anvil, 2 the anvil's face
	// ax and ay are where the face is, in cells from the plate's corner.
	ax, ay float64
}

func newMenuForge(fw, fh int) *menuForge {
	f := &menuForge{fw: max(fw, 1), fh: max(fh, 2)}
	f.fire = make([]float64, f.fw*f.fh)
	f.mask = make([]uint8, f.fw*f.fh)
	x0, y0 := (f.fw-22)/2, f.fh-12
	for r := 0; r < 12; r++ {
		for c := 0; c < 22; c++ {
			x, y := x0+c, y0+r
			if menuAnvil[r][c] == '#' && x >= 0 && x < f.fw && y >= 0 && y < f.fh {
				f.mask[y*f.fw+x] = 1
				if r == 0 {
					f.mask[y*f.fw+x] = 2
				}
			}
		}
	}
	f.ax, f.ay = float64(f.fw)/2, float64(y0/2)
	return f
}

// step burns the fire one step: the bottom row is fed, hottest in the
// middle, and every pixel rises a row, drifts a column and cools.
func (f *menuForge) step(rnd *menuRand, strike float64) {
	base, decay := (f.fh-1)*f.fw, 1/(0.8*float64(f.fh))
	for x := 0; x < f.fw; x++ {
		prof := 1.0
		if f.fw > 1 {
			prof = 0.5 + 0.5*math.Sin(math.Pi*float64(x)/float64(f.fw-1))
		}
		f.fire[base+x] = min(1, prof*(0.72+0.4*rnd.f())+strike*0.3)
	}
	for y := 1; y < f.fh; y++ {
		for x := 0; x < f.fw; x++ {
			dx := min(max(x-int(rnd.f()*3)+1, 0), f.fw-1)
			f.fire[(y-1)*f.fw+dx] = max(f.fire[y*f.fw+x]-rnd.f()*2*decay, 0)
		}
	}
}

// plainFire is the fire's glyphs in the plain glyph set, from embers up.
var plainFire = []rune(" .:-=+*%#")

// draw paints the plate's inside with its top left corner at (x0, y0).
// Two half cells make a cell; in the plain glyph set a cell is one glyph
// whose weight is the fire's heat there.
func (f *menuForge) draw(g *mGrid, pal *menuPalette, x0, y0 int, strike float64, plain bool) {
	half := func(i int) (tcell.Color, bool) {
		switch {
		case f.mask[i] == 2 && strike > 0.06:
			return rampAt(&pal.heat, 0.3+strike*0.7), true
		case f.mask[i] != 0:
			return pal.anvil, true
		}
		return rampAt(&pal.fire, f.fire[i]), false
	}
	for cy := 0; cy < f.fh/2; cy++ {
		for cx := 0; cx < f.fw; cx++ {
			a := cy*2*f.fw + cx
			b := a + f.fw
			top, topIron := half(a)
			bot, botIron := half(b)
			if !plain {
				g.set(x0+cx, y0+cy, '▀', top, bot, false)
				continue
			}
			switch {
			case topIron && f.mask[a] == 2 && strike > 0.06:
				g.set(x0+cx, y0+cy, '#', top, pal.anvil, true)
			case topIron || botIron:
				g.set(x0+cx, y0+cy, '#', pal.anvilInk, pal.anvil, false)
			default:
				v := max(f.fire[a], f.fire[b])
				r := plainFire[min(int(v*float64(len(plainFire))), len(plainFire)-1)]
				g.set(x0+cx, y0+cy, r, rampAt(&pal.fire, min(1, v+0.2)), rampAt(&pal.fire, v*0.45), false)
			}
		}
	}
}

// ---- the sky ----

type menuStar struct {
	x, y  int
	phase float64
	r     rune
}

// newMenuStars scatters stars over a band w wide and h tall.
func newMenuStars(rnd *menuRand, w, h int) []menuStar {
	if w <= 0 || h <= 0 {
		return nil
	}
	stars := make([]menuStar, 0, w*45/100)
	for i := 0; i < w*45/100; i++ {
		st := menuStar{x: int(rnd.f() * float64(w)), y: int(rnd.f() * float64(h)), phase: rnd.f() * 2 * math.Pi, r: '·'}
		if rnd.f() < 0.3 {
			st.r = '∙'
		}
		stars = append(stars, st)
	}
	return stars
}

func drawMenuStars(g *mGrid, pal *menuPalette, stars []menuStar, t float64) {
	for _, s := range stars {
		g.put(s.x, s.y, s.r, theme.Mix(pal.faint, pal.dim, 0.5+0.5*math.Sin(t*1.3+s.phase)))
	}
}
