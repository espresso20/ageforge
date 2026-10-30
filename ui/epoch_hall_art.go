package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// epoch_hall_art.go draws the Epoch screen's front page: a canvas of cells
// (the main menu's mGrid) that the blocks are drawn onto, each centred as a
// unit, and one pass that turns the canvas into tview text. The look is
// the menu's: a banner painted cell by cell over a strip of stars, boxed
// tiles, badge cards in their rarity's colour.

// hallMaxWidth is the widest the blocks spread; past it they keep their
// size and the margins grow.
const hallMaxWidth = 190

// hallMotion is what moves on the page: the clock in seconds when the
// motion setting is on. Off, t is held at zero and the page is one still
// frame.
type hallMotion struct {
	on bool
	t  float64
}

// hallCtx is the canvas and what the blocks need to draw on it.
type hallCtx struct {
	end      int // the row under the last block, before the sky
	g        *mGrid
	w, cw    int // the canvas, and the width the blocks may spread over
	rows     int // the screen's rows
	y        int // the next free row
	plain    bool
	tier     mapmodel.GlyphTier
	mo       hallMotion
	pal      *menuPalette
	duo      bool
	gap      int // rows between two blocks
	bg, text tcell.Color
}

func newHallCtx(w, rows int, tier mapmodel.GlyphTier, mo hallMotion) *hallCtx {
	th := theme.Active()
	c := &hallCtx{
		g: newMGrid(w, 600), w: w, rows: rows, tier: tier, plain: tier == mapmodel.TierASCII, mo: mo,
		pal: newMenuPalette(th), duo: th.Duotone,
		bg: th.Color(theme.RoleBackground), text: th.Color(theme.RoleText),
	}
	c.cw = min(w, hallMaxWidth)
	c.gap = 1
	switch {
	case rows >= 80:
		c.gap = 3
	case rows >= 60:
		c.gap = 2
	}
	return c
}

// ink is a colour the page may draw with: folded to one ink on one paper
// in a duotone theme (Monochrome, Parchment), and held to the contrast
// rule for art.
func (c *hallCtx) ink(col tcell.Color) tcell.Color {
	if c.duo {
		col = theme.Mix(c.bg, c.text, 0.2+0.8*theme.RelativeLuminance(col))
	}
	return theme.Legible(col, c.bg, 3)
}

// role is a theme role as an ink.
func (c *hallCtx) role(r theme.Role) tcell.Color { return c.ink(theme.Color(r)) }

// hue is a fixed identity colour (a rarity's, a tile's) as an ink.
func (c *hallCtx) hue(hex string) tcell.Color {
	var r, g, b int32
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return c.role(theme.RoleAccent)
	}
	return c.ink(tcell.NewRGBColor(r, g, b))
}

// mark picks a glyph for the Unicode set or the plain one.
func (c *hallCtx) mark(uni, ascii string) string { return hallMark(c.plain, uni, ascii) }

// left is the column that centres something n cells wide.
func (c *hallCtx) left(n int) int { return max((c.w-n)/2, 0) }

// centre writes s centred on the row y, and returns its width.
func (c *hallCtx) centre(y int, s string, fg tcell.Color, bold bool) int {
	n := runeLen(s)
	c.g.text(c.left(n), y, s, fg, bold)
	return n
}

// heading is a section's title between two rules, centred.
func (c *hallCtx) heading(title string) {
	rule := strings.Repeat(c.mark("─", "-"), 6)
	n := runeLen(title) + 2*len(rule) + 4
	x := c.left(n)
	c.g.text(x, c.y, rule, c.role(theme.RoleDim), false)
	c.g.text(x+len(rule)+2, c.y, title, c.role(theme.RoleAccent), true)
	c.g.text(x+len(rule)+4+runeLen(title), c.y, rule, c.role(theme.RoleDim), false)
	c.y += 2
}

// paragraph writes plain text wrapped to the content width, each line
// centred, never cut.
func (c *hallCtx) paragraph(s string, fg tcell.Color) {
	for _, l := range wrapText(s, c.cw) {
		c.centre(c.y, l, fg, false)
		c.y++
	}
}

// box draws a rounded frame in fg around (x, y, w, h), with ground
// behind it when hasGround is set.
func (c *hallCtx) box(x, y, w, h int, fg tcell.Color, ground *tcell.Color, bold bool) {
	if ground != nil {
		c.g.fill(x, y, w, h, fg, *ground)
	}
	for i := 1; i < w-1; i++ {
		c.g.put(x+i, y, '─', fg)
		c.g.put(x+i, y+h-1, '─', fg)
	}
	for j := 1; j < h-1; j++ {
		c.g.put(x, y+j, '│', fg)
		c.g.put(x+w-1, y+j, '│', fg)
	}
	c.g.put(x, y, '╭', fg)
	c.g.put(x+w-1, y, '╮', fg)
	c.g.put(x, y+h-1, '╰', fg)
	c.g.put(x+w-1, y+h-1, '╯', fg)
	if bold {
		for j := 0; j < h; j++ {
			for i := 0; i < w; i++ {
				if i == 0 || j == 0 || i == w-1 || j == h-1 {
					if g := c.g; g.in(x+i, y+j) {
						g.c[(y+j)*g.w+x+i].bold = true
					}
				}
			}
		}
	}
}

// ---- the epoch's colours ----

// epochInks are the four stops of an epoch's gradient: a shadow that
// leans toward grey, its own hue, a pale one and the heat of a shimmer.
func (c *hallCtx) epochInks(order int) [4]tcell.Color {
	base := theme.MapHueColor(theme.EpochHue(order))
	grey := theme.Mix(c.bg, c.text, 0.55)
	white := tcell.NewRGBColor(255, 255, 255)
	return [4]tcell.Color{
		theme.Mix(theme.Mix(base, grey, 0.6), c.bg, 0.2),
		base,
		theme.Mix(base, white, 0.45),
		theme.Mix(base, white, 0.85),
	}
}

func stopsAt(stops [4]tcell.Color, v float64) tcell.Color {
	v = math.Max(0, math.Min(1, v)) * float64(len(stops)-1)
	i := int(v)
	if i >= len(stops)-1 {
		return stops[len(stops)-1]
	}
	return theme.Mix(stops[i], stops[i+1], v-float64(i))
}

// bannerInk is the colour of the banner's cell (x, y) of a w by h one:
// two slow waves along the name, brighter toward the foot, and, with
// motion, a glint that runs through it.
func (c *hallCtx) bannerInk(stops [4]tcell.Color, x, y, w, h int) tcell.Color {
	u, v := float64(x)/float64(max(w-1, 1)), float64(y)/float64(max(h-1, 1))
	ph := 0.6
	if c.mo.on {
		ph = c.mo.t * 0.7
	}
	val := 0.30 + 0.28*v + 0.18*math.Sin(u*7+ph) + 0.10*math.Sin(u*17-ph*1.7)
	if c.mo.on {
		pos := math.Mod(c.mo.t*0.12, 1.6) - 0.3
		d := (u - pos) / 0.10
		val += 0.55 * math.Exp(-d*d)
	}
	return c.ink(stopsAt(stops, val))
}

// banner draws the epoch's name in iron, painted cell by cell in the
// epoch's colours over a strip of stars the width of the page, and returns
// the colours' stops for the lines under it.
func (c *hallCtx) banner(name string, order int, seed int64) {
	limit := 8
	if c.rows >= 36 {
		limit = 16
	}
	room := c.cw
	big := []ironMode{{sx: 2, sy: 2}, {sx: 1, sy: 2}}
	modes := []ironMode{{sx: 2, sy: 1}, {sx: 1, sy: 1}}
	if !c.plain {
		modes = append(modes, ironMode{sy: 1, half: true})
	}
	p := placeIron(name, append(big, modes...), room, limit)
	h := p.h
	if !p.ok {
		h = 1
	}
	band := h + 2
	stops := c.epochInks(order)

	// The stars.
	rnd := &menuRand{s: uint64(seed)*0x9e3779b97f4a7c15 | 1}
	for i := 0; i < 3; i++ {
		rnd.next()
	}
	sky := newMGrid(c.w, band)
	t := 0.0
	if c.mo.on {
		t = c.mo.t
	}
	drawMenuStars(sky, c.pal, newMenuStars(rnd, c.w, band), t)
	for y := 0; y < band; y++ {
		for x := 0; x < c.w; x++ {
			if cell := sky.c[y*sky.w+x]; cell.r != ' ' {
				c.g.put(x, c.y+y, cell.r, c.ink(cell.fg))
			}
		}
	}

	// The name.
	if !p.ok {
		text := strings.ToUpper(name)
		if s := spaced(text); runeLen(s) <= room {
			text = s
		}
		x := c.left(runeLen(text))
		for i, r := range []rune(text) {
			c.g.put(x+i, c.y+1, r, c.bannerInk(stops, i, 0, runeLen(text), 1))
			c.g.c[(c.y+1)*c.g.w+x+i].bold = true
		}
		c.y += band
		return
	}
	art := newMGrid(p.w, p.h)
	drawIron(art, c.pal, p.rows, 0, 0, p.mode, func(int, int) float64 { return 0.7 }, c.plain)
	if p.tail != "" {
		tl := spaced(strings.ToUpper(p.tail))
		art.text(max((p.w-runeLen(tl))/2, 0), p.h-1, tl, c.text, true)
	}
	x0 := c.left(p.w)
	for y := 0; y < art.h; y++ {
		for x := 0; x < art.w; x++ {
			cell := art.c[y*art.w+x]
			if cell.r == ' ' && !cell.hasBg {
				continue
			}
			ink := c.bannerInk(stops, x, y, art.w, art.h)
			out := cell
			out.fg, out.bold = ink, true
			if cell.hasBg {
				out.bg = theme.Mix(ink, c.bg, 0.55)
				if cell.r == '▌' || cell.r == '▐' {
					out.bg = c.bannerInk(stops, min(x+1, art.w-1), y, art.w, art.h)
				}
			}
			c.g.store(x0+x, c.y+1+y, out)
		}
	}
	c.y += band
}

// ---- numbers in blocks ----

// hallGlyphs is the big digits: three rows each, built from the block
// glyphs, one cell wide for the point and three for the rest.
var hallGlyphs = map[rune][3]string{
	'0': {"█▀█", "█ █", "▀▀▀"}, '1': {"▄█ ", " █ ", "▄█▄"}, '2': {"▀▀█", "▄▀▀", "▀▀▀"},
	'3': {"▀▀█", " ▀█", "▀▀▀"}, '4': {"█ █", "▀▀█", "  ▀"}, '5': {"█▀▀", "▀▀█", "▀▀▀"},
	'6': {"█▀▀", "█▀█", "▀▀▀"}, '7': {"▀▀█", "  █", "  ▀"}, '8': {"█▀█", "█▀█", "▀▀▀"},
	'9': {"█▀█", "▀▀█", "▀▀▀"}, '.': {" ", " ", "."},
	'K': {"█ █", "█▀▄", "▀ ▀"}, 'M': {"█▄█", "█▀█", "▀ ▀"}, 'B': {"█▀▄", "█▀▄", "▀▀▀"},
	'T': {"▀█▀", " █ ", " ▀ "}, 'Q': {"█▀█", "█▄█", " ▀▄"},
}

// bigNumber is s in the big digits, ok false for a text the font has no
// glyph for.
func bigNumber(s string, sx int) (rows [3]string, ok bool) {
	for i, r := range s {
		gl, found := hallGlyphs[r]
		if !found {
			return rows, false
		}
		for k := 0; k < 3; k++ {
			if i > 0 {
				rows[k] += strings.Repeat(" ", sx)
			}
			for _, cell := range gl[k] {
				rows[k] += strings.Repeat(string(cell), sx)
			}
		}
	}
	return rows, s != ""
}

// tileInks are the tiles' accents, one each, from the theme's roles and
// three fixed hues for the rest.
func (c *hallCtx) tileInk(i int) tcell.Color {
	switch i % 5 {
	case 0:
		return c.role(theme.RolePositive)
	case 1:
		return c.role(theme.RoleHighlight)
	case 2:
		return c.role(theme.RoleEffect)
	case 3:
		return c.role(theme.RoleAccent)
	}
	return c.hue("#d070e0")
}

// tiles draws the headline figures as boxed tiles in a centred row (or
// rows, where the width does not hold them all), each a big number over
// its label.
func (c *hallCtx) tiles(stats []hallStat) {
	gap := 2
	if c.cw >= 150 {
		gap = 3
	}
	big := c.rows >= 30
	// Where the page is wide the digits are twice as wide, as long as all
	// the tiles still fit in one row.
	sx := 1
	if c.rows >= 40 && c.cw >= 140 {
		sx = 2
	}
	var bigs [][3]string
	need := 12
	for {
		bigs = make([][3]string, len(stats))
		allBig := big
		for i, s := range stats {
			r, ok := bigNumber(s.big, sx)
			if !ok {
				allBig = false
			}
			bigs[i] = r
		}
		need = 12
		for i, s := range stats {
			n := runeLen(s.label) + 4
			if allBig {
				n = max(n, runeLen(bigs[i][0])+4)
			} else {
				n = max(n, runeLen(s.big)+4)
			}
			if s.hasBar {
				n = max(n, runeLen(s.value)+4)
			}
			need = max(need, n)
		}
		fits := len(stats)*need+(len(stats)-1)*gap <= c.cw
		if fits || (sx == 1 && !allBig) {
			big = allBig
			break
		}
		if sx == 1 {
			// Even single-width digits do not fit in one row: small text does.
			big = false
			continue
		}
		sx--
	}
	tw := max(min((c.cw-(len(stats)-1)*gap)/len(stats), 30), need)
	per := min(max((c.cw+gap)/(tw+gap), 1), len(stats))
	nrows := (len(stats) + per - 1) / per
	per = (len(stats) + nrows - 1) / nrows
	th := 5
	if big {
		th = 8
	}
	ground := c.pal.ground
	for r := 0; r < nrows; r++ {
		row := stats[r*per : min((r+1)*per, len(stats))]
		total := len(row)*tw + (len(row)-1)*gap
		x := c.left(total)
		for k, s := range row {
			i := r*per + k
			ink := c.tileInk(i)
			edge := theme.Mix(ink, c.bg, 0.35)
			y := c.y
			c.box(x, y, tw, th, edge, &ground, false)
			ctr := func(row int, s string, fg tcell.Color, bold bool) {
				c.g.text(x+(tw-runeLen(s))/2, y+row, s, fg, bold)
			}
			if big {
				for k := 0; k < 3; k++ {
					ctr(2+k, bigs[i][k], ink, true)
				}
				ctr(5, s.label, c.role(theme.RoleText), true)
				if s.hasBar {
					ctr(6, s.value, c.role(theme.RoleDim), false)
				}
			} else {
				ctr(1, s.big, ink, true)
				ctr(2, s.label, c.role(theme.RoleText), true)
				if s.hasBar {
					ctr(3, s.value, c.role(theme.RoleDim), false)
				}
			}
			if s.hasBar {
				// The progress is the tile's lower edge.
				inner := tw - 2
				filled := int(math.Round(s.frac * float64(inner)))
				if s.frac > 0 && filled == 0 {
					filled = 1
				}
				for j := 0; j < inner; j++ {
					if j < filled {
						c.g.put(x+1+j, y+th-1, '━', ink)
					} else {
						c.g.put(x+1+j, y+th-1, '╌', edge)
					}
				}
			}
			x += tw + gap
		}
		c.y += th + 1
	}
}

// ---- the ages ----

// track draws the ages as a row of blocks: done ones filled in their own
// epoch's colour, the current one lit (and pulsing, with motion), the next
// as an outline and the ones the player may not see as dim dots.
func (c *hallCtx) track(v hallView) {
	type seg struct {
		w     int
		draw  func(x, y int)
		isCur bool
	}
	var segs []seg
	count := len(v.doneEp) + 1 + len(v.ahead)
	// The blocks are as wide as the page can spare.
	bw := 4
	for _, try := range []int{8, 6} {
		if c.cw >= 100 && count*(try+1)+v.unseen*2 <= c.cw*8/10 {
			bw = try
			break
		}
	}
	filled := func(ink tcell.Color) func(x, y int) {
		return func(x, y int) {
			for r := 0; r < 2; r++ {
				for i := 0; i < bw; i++ {
					c.g.put(x+i, y+1+r, '█', ink)
				}
			}
		}
	}
	for _, o := range v.doneEp {
		segs = append(segs, seg{w: bw, draw: filled(c.ink(c.epochInks(o)[1]))})
	}
	curStops := c.epochInks(v.curEp)
	pulse := 0.7
	if c.mo.on {
		pulse = 0.5 + 0.5*math.Sin(c.mo.t*1.6)
	}
	hot := c.ink(theme.Mix(curStops[2], curStops[3], pulse))
	curInk := c.ink(curStops[1])
	segs = append(segs, seg{w: bw, isCur: true, draw: func(x, y int) {
		c.g.put(x+bw/2-1, y, []rune(c.mark("▼", "v"))[0], curInk)
		c.g.put(x+bw/2, y, []rune(c.mark("▼", "v"))[0], curInk)
		filled(hot)(x, y)
	}})
	dim := c.role(theme.RoleDim)
	for range v.ahead {
		segs = append(segs, seg{w: bw, draw: func(x, y int) {
			c.g.put(x, y+1, '┌', dim)
			c.g.put(x+bw-1, y+1, '┐', dim)
			c.g.put(x, y+2, '└', dim)
			c.g.put(x+bw-1, y+2, '┘', dim)
			for i := 1; i < bw-1; i++ {
				c.g.put(x+i, y+1, '─', dim)
				c.g.put(x+i, y+2, '─', dim)
			}
		}})
	}
	for i := 0; i < v.unseen; i++ {
		segs = append(segs, seg{w: 1, draw: func(x, y int) { c.g.put(x, y+1, '·', dim); c.g.put(x, y+2, '·', dim) }})
	}
	// Lay the segments out in rows no wider than the content.
	for i := 0; i < len(segs); {
		n, j := 0, i
		for j < len(segs) && (j == i || n+1+segs[j].w <= c.cw) {
			if j > i {
				n++
			}
			n += segs[j].w
			j++
		}
		x := c.left(n)
		for k := i; k < j; k++ {
			segs[k].draw(x, c.y)
			x += segs[k].w + 1
		}
		c.y += 4
		i = j
	}
	// The current age's name, and what is next.
	c.centre(c.y, v.cur, c.ink(curStops[2]), true)
	c.y++
	var after []string
	if len(v.ahead) > 0 {
		after = append(after, vibe.Sprintf("ui.epoch_hall.hallFrontText.09", "next: %s", v.ahead[0]))
	}
	if v.unseen > 0 {
		after = append(after, vibe.Sprintf("ui.epoch_hall.hallFrontText.03", "%d more to come", v.unseen))
	}
	if len(after) > 0 {
		c.paragraph(strings.Join(after, " "+c.mark("·", "-")+" "), dim)
	}
}

// ---- the badges ----

// rarityInk is a rarity's colour: dim for common, up through green, blue
// and violet to gold.
func (c *hallCtx) rarityInk(word string) tcell.Color {
	switch word {
	case "uncommon":
		return c.hue("#5fc77a")
	case "rare":
		return c.hue("#4aa3ff")
	case "epic":
		return c.hue("#c05cff")
	case "legendary", "mythic":
		return c.hue("#ffb020")
	}
	return c.role(theme.RoleDim)
}

// medalOn draws a badge's art on the canvas with its top left corner at
// (x, y): the case's full art where there is room, its mini where not.
func (c *hallCtx) medalOn(v game.BadgeView, x, y int, mini bool) {
	md := medalOf(v, c.tier)
	w, h := miniW, miniH
	if !mini {
		w, h = medalSize(md)
	}
	tg := newTGrid(w, h)
	n := atRest
	if c.mo.on {
		n = int(c.mo.t * 6)
	}
	if mini {
		drawMini(tg, 0, 0, md, caseFrame(n, md.state == medalEarned && md.tier == config.BadgeLegendary))
	} else {
		drawMedal(tg, 0, 0, md, caseFrame(n, md.animated()), c.plain)
	}
	if c.plain {
		foldBadges(tg)
	}
	pal := newGridPalette(0, nil)
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			cell := tg.c[dy*tg.w+dx]
			if cell.r == ' ' || cell.r == 0 {
				continue
			}
			fg, _, attr := pal.style(cell).Decompose()
			c.g.put(x+dx, y+dy, cell.r, fg)
			if attr&tcell.AttrBold != 0 && c.g.in(x+dx, y+dy) {
				c.g.c[(y+dy)*c.g.w+x+dx].bold = true
			}
		}
	}
}

// badges draws the cards: art, name, and the rarity in its colour, in a
// frame of that colour.
func (c *hallCtx) badges(list []game.BadgeView) {
	if len(list) == 0 {
		c.paragraph(vibe.Line("ui.epoch_hall.hallFrontText.05", "No badges earned yet."), c.role(theme.RoleDim))
		return
	}
	gap := 2
	mini := c.rows < 36 || c.cw < 110
	artW, artH := miniW, miniH
	if !mini {
		for _, b := range list {
			w, h := medalSize(medalOf(b, c.tier))
			artW, artH = max(artW, w), max(artH, h)
		}
	}
	longest := 0
	for _, b := range list {
		for _, wd := range strings.Fields(b.Name) {
			longest = max(longest, runeLen(wd))
		}
		longest = max(longest, runeLen(b.Rarity))
	}
	cardW := max(min((c.cw-(len(list)-1)*gap)/len(list), 28), 14, artW+4, longest+4)
	per := min(max((c.cw+gap)/(cardW+gap), 1), len(list))
	nrows := (len(list) + per - 1) / per
	per = (len(list) + nrows - 1) / nrows
	ground := c.pal.ground
	for r := 0; r < nrows; r++ {
		row := list[r*per : min((r+1)*per, len(list))]
		names := make([][]string, len(row))
		lines := 1
		for i, b := range row {
			names[i] = wrapText(b.Name, cardW-4)
			lines = max(lines, len(names[i]))
		}
		ch := 1 + 1 + artH + 1 + lines + 1 + 1
		x := c.left(len(row)*cardW + (len(row)-1)*gap)
		for i, b := range row {
			rar := c.rarityInk(b.Rarity)
			c.box(x, c.y, cardW, ch, rar, &ground, b.Rarity == "legendary" || b.Rarity == "epic")
			aw, ah := miniW, miniH
			if !mini {
				aw, ah = medalSize(medalOf(b, c.tier))
			}
			c.medalOn(b, x+(cardW-aw)/2, c.y+2+(artH-ah)/2, mini)
			ty := c.y + 2 + artH + 1
			for k, l := range names[i] {
				c.g.text(x+(cardW-runeLen(l))/2, ty+k, l, c.role(theme.RoleBright), true)
			}
			if b.Rarity != "" {
				word := textfmt.Capitalize(b.Rarity)
				c.g.text(x+(cardW-runeLen(word))/2, ty+lines, word, rar, true)
			}
			x += cardW + gap
		}
		c.y += ch + 1
	}
}

// ---- the feats ----

// feat kinds, for the icon and its colour.
const (
	featCatastrophe = iota
	featHarbinger
	featMilestone
	featFastest
)

// feats draws the feats as one block, centred, a coloured icon to a line.
func (c *hallCtx) feats(v hallView) {
	if len(v.feats) == 0 {
		c.paragraph(vibe.Line("ui.epoch_hall.hallFrontText.07", "No feats yet. They come with the ages."), c.role(theme.RoleDim))
		return
	}
	icons := [4]string{c.mark("✸", "*"), c.mark("◈", "@"), c.mark("★", "*"), c.mark("»", ">")}
	inks := [4]tcell.Color{c.role(theme.RoleNegative), c.role(theme.RoleWarning), c.role(theme.RoleHighlight), c.role(theme.RolePositive)}
	lines := make([]string, len(v.feats))
	width := 0
	for i, f := range v.feats {
		lines[i] = strings.TrimSpace(cardTagRe.ReplaceAllString(f, ""))
		width = max(width, runeLen(lines[i])+3)
	}
	x := c.left(width)
	for i, l := range lines {
		k := v.featKinds[i]
		c.g.text(x, c.y, icons[k], inks[k], true)
		c.g.text(x+3, c.y, l, c.role(theme.RoleText), false)
		c.y++
	}
	c.y++
}

// ---- out to text ----

// sky scatters a few stars over the page's empty ground, to the foot of
// the screen, so a tall screen is a night sky around the trophies and not
// a void. A star keeps a cell of air from any text.
func (c *hallCtx) sky(seed int64) {
	h := max(c.y, c.rows-4)
	rnd := &menuRand{s: uint64(seed)*0xd1342543de82ef95 | 1}
	for i := 0; i < 3; i++ {
		rnd.next()
	}
	t := 0.0
	if c.mo.on {
		t = c.mo.t
	}
	g := c.g
	blank := func(x, y int) bool {
		if !g.in(x, y) {
			return true
		}
		cell := g.c[y*g.w+x]
		return cell.r == ' ' && !cell.hasBg
	}
	n := c.w * h / 110
	for i := 0; i < n; i++ {
		x, y := int(rnd.f()*float64(c.w)), int(rnd.f()*float64(h))
		ph := rnd.f() * 2 * math.Pi
		r := '·'
		if rnd.f() < 0.3 {
			r = '∙'
		}
		if !blank(x-2, y) || !blank(x-1, y) || !blank(x, y) || !blank(x+1, y) || !blank(x+2, y) {
			continue
		}
		g.put(x, y, r, c.ink(theme.Mix(c.pal.faint, c.pal.dim, 0.5+0.5*math.Sin(t*1.3+ph))))
	}
	c.y = max(c.y, h)
}

// hallFold is a glyph on the plain tier: the badges' table first, then the
// map's.
func hallFold(r rune) rune {
	if f, ok := badgeFold[r]; ok {
		return f
	}
	return mapmodel.Fold(r, mapmodel.TierASCII)
}

// render turns the canvas into tview text, one line to a row, the unused
// rows at the foot dropped.
func (c *hallCtx) render() string {
	g := c.g
	last := 0
	for y := 0; y < g.h && y < c.y+1; y++ {
		for x := 0; x < g.w; x++ {
			if cell := g.c[y*g.w+x]; cell.r != ' ' || cell.hasBg {
				last = y
			}
		}
	}
	var sb strings.Builder
	for y := 0; y <= last; y++ {
		end := 0
		for x := 0; x < g.w; x++ {
			if cell := g.c[y*g.w+x]; cell.r != ' ' || cell.hasBg {
				end = x + 1
			}
		}
		var cur *mCell
		styled := false
		for x := 0; x < end; x++ {
			cell := g.c[y*g.w+x]
			if cell.r == ' ' && !cell.hasBg {
				if styled && cur.hasBg {
					sb.WriteString("[-:-:-]")
					styled = false
				}
				sb.WriteByte(' ')
				continue
			}
			if !styled || cur.fg != cell.fg || cur.bg != cell.bg || cur.hasBg != cell.hasBg || cur.bold != cell.bold {
				sb.WriteString(hallTag(cell))
				cp := cell
				cur, styled = &cp, true
			}
			r := cell.r
			if c.plain {
				r = hallFold(r)
			}
			sb.WriteRune(r)
		}
		if styled {
			sb.WriteString("[-:-:-]")
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func hallTag(c mCell) string {
	hex := func(col tcell.Color) string {
		if v := col.Hex(); v >= 0 {
			return fmt.Sprintf("#%06x", v)
		}
		return "-"
	}
	bg, bold := "-", "-"
	if c.hasBg {
		bg = hex(c.bg)
	}
	if c.bold {
		bold = "b"
	}
	return "[" + hex(c.fg) + ":" + bg + ":" + bold + "]"
}
