package ui

import (
	"image"
	"image/color"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// history_plot.go is the History chart: the model the panel lays out (lines,
// age spans, marks, the pointer) and the two painters that draw it, one in
// pixels and one in cells. Both read the same model; only the plot differs.

// histSlots is how many measures show at once. Colour and shape belong to the
// slot, so the picker's mark for a shown measure is its slot's.
const histSlots = 3

// histInk is the colours of the chart, all from the theme.
type histInk struct {
	ground, text, dim, accent                    color.RGBA
	label, positive, highlight, negative, effect color.RGBA
	bright, surface                              color.RGBA
}

func newHistInk() histInk {
	bg := theme.Of(theme.RoleBackground)
	leg := func(r theme.Role) color.RGBA { return rgbOf(theme.Legible(theme.Of(r), bg, 3)) }
	return histInk{
		ground: rgbOf(bg), text: rgbOf(theme.Legible(theme.Of(theme.RoleText), bg, 4.5)),
		dim: leg(theme.RoleDim), accent: leg(theme.RoleAccent), label: leg(theme.RoleLabel),
		positive: leg(theme.RolePositive), highlight: leg(theme.RoleHighlight),
		negative: leg(theme.RoleNegative), effect: leg(theme.RoleEffect),
		bright: rgbOf(theme.Of(theme.RoleBright)), surface: rgbOf(theme.Of(theme.RoleSurface)),
	}
}

// slot is the colour of a slot.
func (c histInk) slot(i int) color.RGBA {
	return [histSlots]color.RGBA{c.label, c.positive, c.highlight}[i]
}

// ageColor is an age's colour: its epoch's colour, a step lighter or darker
// by its place in the epoch, kept clear of white and black on either theme.
func (c histInk) ageColor(i int) color.RGBA {
	var base color.RGBA
	switch e := min(i/3, 6); e {
	case 0:
		base = toward(c.text, c.ground, .2)
	case 1:
		base = c.negative
	case 2:
		base = c.highlight
	case 3:
		base = toward(c.effect, c.bright, .45)
	case 4:
		base = c.effect
	case 5:
		base = c.label
	default:
		base = toward(toward(c.negative, c.effect, .5), c.bright, .15)
	}
	switch i % 3 {
	case 0:
		base = toward(base, c.bright, .22)
	case 2:
		base = toward(base, c.ground, .28)
	}
	lum := func(c color.RGBA) float64 { return (0.3*float64(c.R) + 0.59*float64(c.G) + 0.11*float64(c.B)) / 255 }
	for n := 0; n < 8 && (lum(base) > 0.72 || lum(base) < 0.28); n++ {
		base = toward(base, color.RGBA{128, 128, 128, 255}, .2)
	}
	return base
}

// histLine is one shown measure: its series in shown units, the past run's
// behind it, and the scale it is drawn on.
type histLine struct {
	m      histMeasure
	slot   int
	cur    histSeries
	gh     histSeries
	lo, hi float64
	lg     bool
}

// frac is where v sits on the line's scale, 0 (bottom) to 1.
func (l histLine) frac(v float64) float64 {
	if !l.lg {
		return math.Min(1, math.Max(0, v/l.hi))
	}
	if v <= l.lo {
		return 0
	}
	return math.Min(1, math.Log10(v/l.lo)/math.Log10(l.hi/l.lo))
}

// at is the series' value at tick t: the last sample at or before it. ok is
// false past the series' end, and for a series with no samples.
func (s histSeries) at(t int) (float64, bool) {
	if len(s.T) == 0 || t > s.T[len(s.T)-1] {
		return 0, false
	}
	return s.V[max(0, sort.SearchInts(s.T, t+1)-1)], true
}

// histGeo is where the chart's parts sit, in the panel's cells.
type histGeo struct {
	w, h, px0, pcols int
	y0, axisY        int // the chart picture's top row; the time axis row
	legendY, footY   int
	listTop          int // the runs list's heading row, 0 when there is none
	listRows         int // the list's scrolling rows
}

func (g histGeo) picRows() int  { return max(0, g.axisY-g.y0) }
func (g histGeo) plotTop() int  { return g.y0 + 2 }
func (g histGeo) plotRows() int { return max(0, g.axisY-g.plotTop()) }

// histGeometry lays the panel out: the chart takes the height the runs list
// (in All runs) leaves.
func histGeometry(w, h int, all bool, runs int) histGeo {
	g := histGeo{w: w, h: h, px0: 31, y0: 3, footY: h - 2}
	g.pcols = max(0, w-2-g.px0)
	switch {
	case all && runs > 0:
		g.listRows = min(11, runs)
		g.listTop = g.footY - 1 - (2 + g.listRows)
		g.legendY = g.listTop - 1
	case all:
		g.listTop = g.footY - 4
		g.legendY = g.listTop - 1
	default:
		g.legendY = g.footY - 2
	}
	g.axisY = g.legendY - 1
	return g
}

// histMark is how a mark of the run is drawn: its sign and colour.
func (c histInk) markSign(kind string) (string, color.RGBA) {
	switch kind {
	case game.MarkWonder:
		return kind, c.highlight
	case game.MarkCatastrophe:
		return kind, c.negative
	case game.MarkHarbinger:
		return kind, c.label
	case game.MarkAway:
		return kind, c.dim
	case game.MarkEvent:
		return kind, c.positive
	}
	return game.MarkAge, c.accent
}

// histTag is an age's name on the tag row, in cells from the chart's left.
type histTag struct {
	x    int
	name string
	idx  int
	now  bool
}

// histView is the chart as the panel laid it out.
type histView struct {
	geo    histGeo
	ink    histInk
	set    *rules.Set
	keys   []string // the ruleset's age keys, in order
	all    bool
	logOn  bool
	lo, hi int   // the ticks the axis spans
	end    int   // the run in play's newest tick
	ticks  []int // the run in play's sample ticks in the range: the pointer's stops
	ptr    int   // the tick under the pointer, -1 when there are no samples

	lines  []histLine
	spans  []histSpan
	marks  []game.HistoryMark
	ghost  *game.RunRecord // nil: none picked, or none kept
	gEnd   int
	gSpans []histSpan
}

func (v *histView) span() int { return max(1, v.hi-v.lo) }

// ageIdx is an age's place in the ruleset's order.
func (v *histView) ageIdx(key string) int { return max(0, slices.Index(v.keys, key)) }

func (v *histView) ageName(key string) string {
	return v.set.Name(rules.KindAge, key)
}

// cellX is the cell, from the chart's left, tick t falls in.
func (v *histView) cellX(t int) int { return (t - v.lo) * v.geo.pcols / v.span() }

// frameAt is the tick at the middle of chart column x.
func (v *histView) tickAt(x, cols int) int { return v.lo + (2*x+1)*v.span()/(2*max(1, cols)) }

// nearest is the sample tick closest to t.
func (v *histView) nearest(t int) int {
	if len(v.ticks) == 0 {
		return -1
	}
	i := sort.SearchInts(v.ticks, t)
	if i == len(v.ticks) || i > 0 && t-v.ticks[i-1] <= v.ticks[i]-t {
		i--
	}
	return v.ticks[i]
}

// axisTicks are the ticks the time axis labels, at round times.
func (v *histView) axisTicks() []int {
	perCell := float64(v.span()) * game.BaseTickInterval.Minutes() / float64(max(1, v.geo.pcols))
	step := 10080.0
	for _, s := range []float64{1, 2, 5, 10, 20, 30, 60, 120, 180, 360, 720, 1440, 2880, 4320, 10080} {
		if s/perCell >= 11 {
			step = s
			break
		}
	}
	st := max(1, int(step*float64(time.Minute/game.BaseTickInterval)))
	var out []int
	for t := (v.lo + st - 1) / st * st; t <= v.hi; t += st {
		out = append(out, t)
	}
	return out
}

// histTime is a tick as a time into the run: "0", "45s", "1h 12m", "3d 4h 20m".
func histTime(t int) string {
	if t <= 0 {
		return "0"
	}
	return textfmt.DurationExact(time.Duration(t) * game.BaseTickInterval)
}

// tags are the ages' names: each sits inside its own band from one cell after
// its left edge, in full, else without " Age", else not at all.
func (v *histView) tags() []histTag {
	var out []histTag
	for i, sp := range v.spans {
		l, r := v.cellX(sp.Start), v.geo.pcols
		if i+1 < len(v.spans) {
			r = v.cellX(sp.End)
		}
		full := v.ageName(sp.Key)
		for _, name := range []string{full, strings.TrimSuffix(full, " Age")} {
			if l+1+len([]rune(name)) < r {
				out = append(out, histTag{l + 1, name, v.ageIdx(sp.Key), i == len(v.spans)-1})
				break
			}
		}
	}
	return out
}

// Signs are 9 by 9 art pixels on a slab of 11.
var histSignArt = map[string][9]string{
	game.MarkAge:         {".........", "......##.", "......##.", "....####.", "....####.", "..######.", "..######.", "########.", "#########"},
	game.MarkWonder:      {"....#....", "...###...", "..#####..", "#########", ".#.#.#.#.", ".#.#.#.#.", ".#.#.#.#.", "#########", "#########"},
	game.MarkCatastrophe: {".....###.", "....###..", "...###...", "..######.", "....###..", "...###...", "..###....", "..##.....", "..#......"},
	game.MarkHarbinger:   {"#........", "#####....", "#######..", "#########", "#######..", "#####....", "#........", "#........", "#........"},
	game.MarkEvent:       {"....#....", "...###...", "..#####..", ".#######.", "#########", ".#######.", "..#####..", "...###...", "....#...."},
	game.MarkAway:        {"...####..", "..##.....", ".##......", ".##......", ".##......", ".##......", "..##.....", "...####..", "........."},
	"end-prestige":       {"....#....", "...###...", "..#####..", ".#######.", "...###...", "...###...", "...###...", ".........", "#########"},
	"end-succumbed":      {"##.....##", ".##...##.", "..##.##..", "...###...", "....#....", "...###...", "..##.##..", ".##...##.", "##.....##"},
	"end-left":           {"#####....", "#...#....", "#...#.#..", "#...#.##.", "#...#####", "#...#.##.", "#...#.#..", "#...#....", "#####...."},
	"end-play":           {"##.......", "####.....", "######...", "########.", "#########", "########.", "######...", "####.....", "##......."},
	"crown":              {"#...#...#", "##.###.##", "#########", "#########", "#########", "#########", "#########", ".........", "........."},
	"ghost":              {"..#####..", ".#######.", "##.###.##", "#########", "#########", "#########", "#########", "#########", "#.##.##.#"},
}

// histGlyph is the letter a sign reads as in cells, where there are no pictures.
var histGlyph = map[string]string{
	game.MarkAge: "A", game.MarkWonder: "W", game.MarkCatastrophe: "!", game.MarkHarbinger: "H",
	game.MarkEvent: "*", game.MarkAway: "z", "end-prestige": "^", "end-succumbed": "x", "end-left": "<",
	"end-play": ">", "crown": "$", "ghost": "&",
}

// histDrawGlyph puts a sign's 9 by 9 pattern at (x, y), in art pixels.
func histDrawGlyph(a art, x, y int, kind string, col color.RGBA) {
	for j, row := range histSignArt[kind] {
		for i, r := range row {
			if r == '#' {
				a.px(x+i, y+j, col)
			}
		}
	}
}

// histDrawSign is a sign on its slab: an outline, a tinted face, the sign.
func histDrawSign(a art, x, y int, kind string, col, ground color.RGBA) {
	a.fill(image.Rect(x, y, x+11, y+11), 1, col)
	a.fill(image.Rect(x+1, y+1, x+10, y+10), 0, toward(ground, col, .22))
	histDrawGlyph(a, x+1, y+1, kind, col)
}

// Slot shapes: a circle, a square, a triangle, so colour is never the only cue.
var histShapeArt = [histSlots][7]string{
	{"..###..", ".#####.", "#######", "#######", "#######", ".#####.", "..###.."},
	{"#######", "#######", "#######", "#######", "#######", "#######", "#######"},
	{"...#...", "..###..", "..###..", ".#####.", ".#####.", "#######", "#######"},
}

// drawShape is a slot's shape centred at (cx, cy), with a one-pixel ring.
func drawShape(a art, cx, cy, shape int, col, ring color.RGBA) {
	ox, oy := cx-3, cy-3
	pat := histShapeArt[shape]
	on := func(x, y int) bool { return x >= 0 && x < 7 && y >= 0 && y < 7 && pat[y][x] == '#' }
	for y := -1; y < 8; y++ {
		for x := -1; x < 8; x++ {
			near := false
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					near = near || on(x+dx, y+dy)
				}
			}
			if near {
				a.px(ox+x, oy+y, ring)
			}
		}
	}
	for y := 0; y < 7; y++ {
		for x := 0; x < 7; x++ {
			if on(x, y) {
				a.px(ox+x, oy+y, col)
			}
		}
	}
}

// shapeImg is a slot's shape in one cell.
func shapeImg(shape int, col, ground color.RGBA) *image.RGBA {
	cw, ch := cellPixels()
	img := image.NewRGBA(image.Rect(0, 0, cw, ch))
	k := max(1, ch/12)
	drawShape(art{img, k}, cw/k/2, ch/k/2, shape, col, ground)
	return img
}

// picture is the chart in pixels: bands, ribbon, grid, lines, signs, pointer.
func (v *histView) picture(cw, ch int) *image.RGBA {
	g, c := v.geo, v.ink
	k := max(1, ch/12)
	img := image.NewRGBA(image.Rect(0, 0, g.pcols*cw, g.picRows()*ch))
	a := art{img, k}
	aw, ah := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	ar := ch / k // art pixels in a row
	a.fill(image.Rect(0, 0, aw, ah), 0, c.ground)
	bandTop, plotTop := ar, 2*ar
	span := v.span()
	axf := func(t int) int { return (t - v.lo) * aw / span }

	for _, spans := range [][]histSpan{v.spans} {
		for _, sp := range spans {
			x0, x1 := max(0, axf(sp.Start)), min(aw, axf(sp.End))
			if x1 <= x0 {
				continue
			}
			col := c.ageColor(v.ageIdx(sp.Key))
			a.fill(image.Rect(x0, bandTop, x1, ah), 0, toward(c.ground, col, .09))
			a.fill(image.Rect(x0, plotTop, x1, plotTop+3), 0, col)
			if sp.Start > v.lo {
				for y := bandTop; y < plotTop; y++ {
					a.px(x0, y, toward(c.ground, col, .7))
				}
				for y := plotTop + 3; y < ah; y += 2 {
					a.px(x0, y, toward(c.ground, col, .35))
				}
			}
		}
	}
	grid := toward(c.ground, c.dim, .30)
	for q := 1; q < 4; q++ {
		y := plotTop + 4 + (ah-plotTop-5)*q/4
		for x := 0; x < aw; x += 3 {
			a.px(x, y, grid)
		}
	}
	for _, t := range v.axisTicks() {
		for y, x := plotTop+4, axf(t); y < ah; y += 3 {
			a.px(x, y, grid)
		}
	}

	plotH, topY := ah-plotTop-6, plotTop+5
	yOf := func(l histLine, s histSeries, t int) (int, bool) {
		val, ok := s.at(t)
		return topY + int(math.Round((1-l.frac(val))*float64(plotH))), ok
	}
	endX := min(aw, axf(v.end))
	// a dithered fill under the first line
	if len(v.lines) > 0 && !v.lines[0].lg && len(v.lines[0].cur.T) > 0 {
		l, f := v.lines[0], toward(c.ground, c.slot(v.lines[0].slot), .16)
		for x := 0; x < endX; x++ {
			y0, _ := yOf(l, l.cur, v.tickAt(x, aw))
			for y := y0 + 2; y < ah; y++ {
				if y-y0-2 < 4 && float64(4-(y-y0-2))/5 < bayer4[y&3][x&3] {
					continue
				}
				a.px(x, y, f)
			}
		}
	}
	// the past run: dotted, one pixel, dimmer
	if v.ghost != nil {
		for _, l := range v.lines {
			col, prev := toward(c.ground, c.slot(l.slot), .62), 0
			for x := 0; x < min(aw, axf(v.gEnd)); x++ {
				y, ok := yOf(l, l.gh, v.tickAt(x, aw))
				if !ok {
					continue
				}
				if x/2%2 == 0 {
					a.px(x, y, col)
					for yy := min(y, prev); x > 0 && prev != 0 && yy <= max(y, prev); yy += 2 {
						a.px(x, yy, col)
					}
				}
				prev = y
			}
		}
		for i, sp := range v.gSpans {
			if i > 0 {
				for y := ah - 5; y < ah; y++ {
					a.px(axf(sp.Start), y, toward(c.ground, c.ageColor(v.ageIdx(sp.Key)), .6))
				}
			}
		}
	}
	// the run in play: solid, two pixels thick, stepped
	for i := len(v.lines) - 1; i >= 0; i-- {
		l := v.lines[i]
		if len(l.cur.T) == 0 {
			continue
		}
		col, prev := c.slot(l.slot), 0
		ys := make([]int, aw)
		for x := 0; x < endX; x++ {
			y, _ := yOf(l, l.cur, v.tickAt(x, aw))
			ys[x] = y
			y0, y1 := y, y
			if x > 0 {
				y0, y1 = min(y, prev), max(y, prev)
			}
			for yy := y0; yy <= y1+1; yy++ {
				a.px(x, yy, col)
			}
			prev = y
		}
		if len(l.cur.T) > 1 && (l.cur.T[1]-l.cur.T[0])*aw/span >= 6 {
			for _, t := range l.cur.T {
				if x := axf(t); x >= 0 && x < endX {
					a.fill(image.Rect(x-1, ys[x]-1, x+3, ys[x]+3), 0, col)
					a.fill(image.Rect(x, ys[x], x+2, ys[x]+2), 0, c.ground)
				}
			}
		}
		if endX > 1 {
			drawShape(a, endX-1, ys[endX-1], l.slot, col, c.ground)
		}
	}
	// signs along the top, none closer than 13 pixels to the one before
	last := -100
	for _, mk := range v.marks {
		if x := axf(mk.Tick); mk.Tick >= v.lo && mk.Tick <= v.hi && x-last >= 13 {
			kind, col := c.markSign(mk.Kind)
			histDrawSign(a, max(0, min(x-6, aw-12)), 1, kind, col, c.ground)
			last = x
		}
	}
	// the pointer
	if v.ptr >= 0 {
		x := max(0, min(aw-1, axf(v.ptr)))
		for y := plotTop + 4; y < ah; y++ {
			a.px(x, y, c.bright)
		}
		for _, l := range v.lines {
			if y, ok := yOf(l, l.gh, v.ptr); ok && v.ghost != nil {
				drawShape(a, x, y, l.slot, c.ground, c.slot(l.slot))
			}
			if y, ok := yOf(l, l.cur, v.ptr); ok {
				drawShape(a, x, y, l.slot, c.slot(l.slot), c.ground)
			}
		}
	}
	return img
}

// cells paints the chart in cells, one column to a cell, at org (the chart's
// top left): the plain tier's plot, with the same bands, signs and pointer.
func (v *histView) cells(p *kit.Painter, org image.Point) {
	g, c := v.geo, v.ink
	fg := func(col color.RGBA) kit.Style { return kit.Style{}.Foreground(theme.RGB(col.R, col.G, col.B)) }
	rows := g.plotRows() - 1
	ageAt := func(t int) int {
		for _, sp := range v.spans {
			if t < sp.End || sp == v.spans[len(v.spans)-1] {
				return v.ageIdx(sp.Key)
			}
		}
		return 0
	}
	for x := 0; x < g.pcols; x++ {
		t := v.tickAt(x, g.pcols)
		if len(v.spans) > 0 {
			p.Print(org.X+x, org.Y+2, 1, "▀", fg(c.ageColor(ageAt(t))))
		}
		for i := len(v.lines) - 1; i >= 0; i-- {
			l := v.lines[i]
			val, ok := l.cur.at(min(t, v.end))
			if !ok || rows < 1 {
				continue
			}
			eighths := int(math.Round(l.frac(val) * float64(rows*8)))
			top := rows - (eighths+7)/8 // the row the line's top is in
			if l.slot > 0 {
				p.Print(org.X+x, org.Y+3+top, 1, string([]rune("●■▲")[l.slot]), fg(c.slot(l.slot)))
				continue
			}
			for r := top; r < rows; r++ {
				glyph := "█"
				if r == top && eighths%8 != 0 {
					glyph = string([]rune("▁▂▃▄▅▆▇")[eighths%8-1])
				}
				p.Print(org.X+x, org.Y+3+r, 1, glyph, fg(c.slot(0)))
			}
		}
	}
	for _, mk := range v.marks {
		if x := v.cellX(mk.Tick); mk.Tick >= v.lo && mk.Tick <= v.hi {
			kind, col := c.markSign(mk.Kind)
			p.Print(org.X+min(x, g.pcols-1), org.Y, 1, histGlyph[kind], fg(col))
		}
	}
	if v.ptr >= 0 {
		x := min(g.pcols-1, v.cellX(v.ptr))
		for r := 0; r <= rows; r++ {
			if l0 := v.lines; len(l0) == 0 || r < rows-(histEighths(l0[0], v.ptr, rows)+7)/8 {
				p.Print(org.X+x, org.Y+3+r, 1, "│", fg(c.bright))
			}
		}
	}
}

// histEighths is the height of a line at tick t, in eighths of a cell.
func histEighths(l histLine, t, rows int) int {
	val, _ := l.cur.at(t)
	return int(math.Round(l.frac(val) * float64(rows*8)))
}
