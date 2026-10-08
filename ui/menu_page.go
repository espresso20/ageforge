package ui

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/ui/mapstyle"
)

// menu_page.go lays the main menu out and draws it. There are two pages:
//
//   - the town page, for an account with a game: the forged wordmark in a
//     band of night sky, the player's town under it (menu_town.go), embers
//     rising, the menu in a box on the right and a caption that names the
//     town and its age;
//   - the forge page, for a first visit: a title page with the wordmark as
//     hot iron, the menu set as a table of contents with dot leaders, and a
//     plate of the anvil over the fire.
//
// A page is a function of the terminal's size, what the menu says
// (menuView) and the scene's state (menuScene), drawn into a grid. Nothing
// here reads a file or the engine.

// The menu's entries.
const (
	miContinue = "continue"
	miNew      = "new"
	miLoad     = "load"
	miBadges   = "badges"
	miThemes   = "themes"
	miAccounts = "accounts"
	miUpdates  = "updates"
	miQuit     = "quit"
)

// menuItem is one entry of the menu.
type menuItem struct {
	id    string
	label string
	key   rune
	// extras is what stands at the right of the row, fullest first; the
	// first that fits is drawn. details is the line under the row.
	extras  []string
	details []string
}

// menuView is what a page says.
type menuView struct {
	items []menuItem
	sel   int
	// forge picks the forge page; otherwise the town page is drawn.
	forge bool
	// captions is the town page's caption, fullest first.
	captions []string
	// version is the version line, fullest first (an update notice may
	// lead it).
	versions []string
	// edition is the edition the forge page's title names ("fourth"), ""
	// when the version does not say.
	edition string
	// ages is how many ages the game has, for the title's lines.
	ages int
	// elite is the forge master's line: a mark, the words, a mark. Empty
	// for everyone else.
	elite [3]string
	// plain: the plain glyph set, which has no blocks.
	plain bool
}

// menuLayout is where a page's parts sit at a size.
type menuLayout struct {
	w, h int
	// compact: the terminal is under the game's least size, and the page
	// is the list alone.
	compact bool
	forge   bool
	// sx is how many cells wide a pixel of the wordmark is; (wx, wy) is
	// its corner.
	sx, wx, wy int
	// tag is the row of the line under the wordmark.
	tag int
	// (mx, my) is the first menu row, mw its width and gap the rows from
	// one entry to the next.
	mx, my, mw, gap int
	// (px, py, pw, ph) is the box: the menu's on the town page, the
	// plate's on the forge page.
	px, py, pw, ph int

	// The town page: top is the first row of land, landW how far the land
	// runs and cap the caption's row.
	top, landW, cap int

	// The forge page: the rows of the line over the wordmark, the rule and
	// the foot, and whether the inner frame is drawn.
	line1, rule, foot int
	inner             bool
}

// The game's least terminal: under it the page is the list alone.
const (
	menuMinW = 80
	menuMinH = 24
)

// menuRows is how many rows n entries take at a gap, the Continue entry's
// line under it counted when there is one.
func menuRows(n, gap int, detail bool) int {
	rows := n + max(n-1, 0)*(gap-1)
	if detail {
		rows++
	}
	return rows
}

// townLayout is the town page at w by h for a menu of n entries.
//
// The wordmark is doubled from 114 columns. The box is 50 wide from 110
// columns and 40 under it, with a blank row between entries from 36 rows.
// The land runs from under the sky to a little past the middle of the box,
// so the town sits in the clear and its edge fades out under the menu.
func townLayout(w, h, n int) menuLayout {
	L := menuLayout{w: w, h: h, sx: 1, wy: 1, tag: 8, top: 9, pw: 40, gap: 1, cap: h - 2}
	margin := 3
	if w >= 114 {
		L.sx = 2
	}
	if w >= 110 {
		L.pw, margin = 50, 4
	}
	if h >= 36 {
		L.gap = 2
	}
	L.wx = (w - menuWordW*L.sx) / 2
	L.ph = menuRows(n, L.gap, true) + 2
	L.px = w - L.pw - margin
	L.py = L.top + 1 + (h-menuMinH)/3
	if L.gap > 1 {
		L.ph += 2
		L.py = L.top + 4 + (h-40)/2
	}
	L.py = max(L.top+1, min(L.py, L.cap-1-L.ph))
	L.mx, L.my, L.mw = L.px+2, L.py+L.gap, L.pw-4
	L.landW = min(w, L.px+L.pw*29/50)
	return L
}

// forgeLayout is the forge page at w by h: the mock-up's two layouts, the
// large from 110 by 38 and the small under it, each centred in whatever
// room the terminal has beyond its own size.
func forgeLayout(w, h int) menuLayout {
	if w >= 110 && h >= 38 {
		ox, oy := (w-120)/2, (h-40)/2
		return menuLayout{w: w, h: h, forge: true, sx: 2, wx: (w - menuWordW*2) / 2, wy: 5 + oy,
			line1: 3 + oy, tag: 13 + oy, rule: 15 + oy, foot: 36 + oy, inner: true,
			mx: 8 + ox, my: 18 + oy, mw: 50, gap: 2,
			px: 64 + ox, py: 17 + oy, pw: 50, ph: 17}
	}
	ox, oy := (w-menuMinW)/2, (h-menuMinH)/2
	return menuLayout{w: w, h: h, forge: true, sx: 1, wx: (w - menuWordW) / 2, wy: 2 + oy,
		line1: 1 + oy, tag: 9 + oy, rule: 10 + oy, foot: 21 + oy,
		mx: 5 + ox, my: 11 + oy, mw: 42, gap: 1,
		px: 50 + ox, py: 11 + oy, pw: 26, ph: 9}
}

// menuLayoutFor is the layout of a view at a size.
func menuLayoutFor(w, h int, v *menuView) menuLayout {
	if w < menuMinW || h < menuMinH {
		return menuLayout{w: w, h: h, compact: true, forge: v.forge}
	}
	if v.forge {
		return forgeLayout(w, h)
	}
	return townLayout(w, h, len(v.items))
}

// ---- the scene ----

// menuRestFrames is how many frames the scene has run when it is first
// drawn, so the picture a still menu holds (the motion setting off) has
// its fire lit and its sparks in the air.
const menuRestFrames = 36

// menuScene is the state of what moves on a page: the clock, the strike,
// the sparks, the fire and the stars.
type menuScene struct {
	L      menuLayout
	rnd    menuRand
	t      float64
	strike float64
	frame  int
	sparks menuSparks
	forge  *menuForge
	stars  []menuStar
}

// newMenuScene makes the scene for a layout, at rest.
func newMenuScene(L menuLayout) *menuScene {
	sc := &menuScene{L: L, rnd: menuRand{s: 0x9e3779b97f4a7c15}}
	if L.compact {
		return sc
	}
	if L.forge {
		sc.forge = newMenuForge(L.pw-2, (L.ph-2)*2)
		for i := 0; i < 70; i++ {
			sc.forge.step(&sc.rnd, 0)
		}
	} else {
		sc.stars = newMenuStars(&sc.rnd, L.w, L.top)
	}
	for i := 0; i < menuRestFrames; i++ {
		sc.step()
	}
	sc.frame = 0
	return sc
}

// anvilAt is where the anvil's face is on the screen.
func (sc *menuScene) anvilAt() (x, y float64) {
	return float64(sc.L.px+1) + sc.forge.ax, float64(sc.L.py+1) + sc.forge.ay
}

// step moves the scene on one animation frame.
func (sc *menuScene) step() {
	sc.frame++
	if sc.L.compact {
		return
	}
	for i := 0; i < menuSub; i++ {
		sc.t += 0.071 * menuPace
		sc.strike *= 0.938
		sc.sparks.step(&sc.rnd)
		if sc.L.forge {
			sc.forge.step(&sc.rnd, sc.strike)
			if sc.rnd.f() < 0.5*menuPace {
				x, y := sc.anvilAt()
				sc.sparks.emit(&sc.rnd, x, y, 1, float64(sc.forge.fw)*0.5, 0.55)
			}
		} else if sc.rnd.f() < 0.55*menuPace {
			// An ember off the town.
			sc.sparks.emit(&sc.rnd, sc.rnd.f()*float64(sc.L.w), float64(sc.L.top+1)+sc.rnd.f()*3, 1, 2, 0.5)
		}
	}
}

// hit strikes the title: a flare along the wordmark and a burst of sparks.
func (sc *menuScene) hit() {
	if sc.L.compact {
		return
	}
	sc.strike = 1
	if sc.L.forge {
		x, y := sc.anvilAt()
		sc.sparks.emit(&sc.rnd, x, y, 46, float64(sc.forge.fw)*0.7, 0.95)
		return
	}
	sc.sparks.emit(&sc.rnd, float64(sc.L.w)/2, float64(sc.L.top), 60, float64(menuWordW*sc.L.sx)*0.9, 0.8)
}

// ---- drawing ----

// fitOption is the first of options that fits in room cells, "" when none
// does.
func fitOption(options []string, room int) string {
	for _, o := range options {
		if o != "" && runeLen(o) <= room {
			return o
		}
	}
	return ""
}

// centred is the column that centres n cells in w.
func centred(w, n int) int { return (w - n) / 2 }

// spaced letterspaces s: one blank between its characters.
func spaced(s string) string {
	return strings.Join(strings.Split(s, ""), " ")
}

// render draws the page of a view into a new grid. town is the picture
// behind the town page (nil: the sky alone) and mf the frame it is drawn
// at.
func (sc *menuScene) render(pal *menuPalette, v *menuView, town *menuTown, mf mapstyle.Frame) *mGrid {
	L := sc.L
	g := newMGrid(L.w, L.h)
	switch {
	case L.compact:
		sc.renderCompact(g, pal, v)
	case L.forge:
		sc.renderForge(g, pal, v)
	default:
		sc.renderTown(g, pal, v, town, mf)
	}
	return g
}

// menuKeysHint is the line that says how the menu is driven.
const menuKeysHint = "↑ ↓ choose · Enter open"

// renderTown draws the town page.
func (sc *menuScene) renderTown(g *mGrid, pal *menuPalette, v *menuView, town *menuTown, mf mapstyle.Frame) {
	L := sc.L
	drawMenuStars(g, pal, sc.stars, sc.t)
	town.draw(g, pal, 0, L.top, L.landW, L.h-L.top, mf, sc.strike > 0.45)
	sc.sparks.draw(g, pal)
	drawWordmark(g, pal, L.wx, L.wy, L.sx, sc.t, sc.strike, 0.5, v.plain)
	if tag := fitOption([]string{" " + numberWords(v.ages) + " ages, one terminal "}, L.w); tag != "" {
		g.textOn(centred(L.w, runeLen(tag)), L.tag, tag, pal.ink, pal.bg, false)
	}

	// The menu's box, on a ground of its own.
	g.fill(L.px, L.py, L.pw, L.ph, pal.groundInk, pal.ground)
	g.box(L.px, L.py, L.pw, L.ph, false, pal.groundDim)
	sc.drawMenu(g, pal, v, false)
	if ver := fitOption(spacedOptions(v.versions), L.pw-4); ver != "" {
		g.text(L.px+L.pw-runeLen(ver)-2, L.py+L.ph-1, ver, pal.groundDim, false)
	}
	sc.drawElite(g, pal, v, L.px+2, L.py, L.pw-4, true)

	// The caption on the left and the keys on the right, under the land.
	room := L.w - 4
	hint := " " + menuKeysHint + " "
	if capt := fitOption(spacedOptions(v.captions), room); capt != "" {
		g.textOn(2, L.cap, capt, pal.groundInk, pal.ground, false)
		room -= runeLen(capt) + 2
	}
	if runeLen(hint) <= room {
		g.textOn(L.w-2-runeLen(hint), L.cap, hint, pal.groundDim, pal.ground, false)
	}
}

// spacedOptions puts a blank on each side of every option, so text on a
// border or on the land has air round it.
func spacedOptions(options []string) []string {
	out := make([]string, 0, len(options))
	for _, o := range options {
		if o != "" {
			out = append(out, " "+o+" ")
		}
	}
	return out
}

// renderForge draws the forge page.
func (sc *menuScene) renderForge(g *mGrid, pal *menuPalette, v *menuView) {
	L := sc.L
	g.box(1, 0, L.w-2, L.h, true, pal.dim)
	if L.inner {
		g.box(3, 1, L.w-6, L.h-2, false, pal.faint)
	}
	// The plate: the fire, with the anvil standing in front of it.
	sc.forge.draw(g, pal, L.px+1, L.py+1, sc.strike, v.plain)
	g.box(L.px, L.py, L.pw, L.ph, true, pal.dim)
	if plate := fitOption([]string{" Plate I. The forge ", " The forge "}, L.pw-4); plate != "" {
		g.textOn(L.px+centred(L.pw, runeLen(plate)), L.py+L.ph-1, plate, pal.dim, pal.bg, false)
	}
	sc.sparks.draw(g, pal)

	inside := L.w - 8
	if line := fitOption([]string{spaced("AN ALMANAC OF " + strings.ToUpper(numberWords(v.ages)) + " AGES"), "AN ALMANAC OF " + strings.ToUpper(numberWords(v.ages)) + " AGES"}, inside); line != "" {
		g.text(centred(L.w, runeLen(line)), L.line1, line, pal.dim, false)
	}
	ax, _ := sc.anvilAt()
	drawWordmark(g, pal, L.wx, L.wy, L.sx, sc.t, sc.strike, (ax-float64(L.wx))/float64(menuWordW*L.sx), v.plain)
	tags := []string{"begun at a campfire"}
	if v.edition != "" {
		tags = []string{"begun at a campfire · the " + v.edition + " edition, much enlarged", "begun at a campfire"}
	}
	if tag := fitOption(tags, inside); tag != "" {
		g.text(centred(L.w, runeLen(tag)), L.tag, tag, pal.ink, false)
	}

	for x := L.mx - 1; x < L.px+L.pw; x++ {
		g.put(x, L.rule, '─', pal.faint)
	}
	g.text(L.mx, L.rule, " CONTENTS ", pal.accent, true)
	sc.drawElite(g, pal, v, L.mx+11, L.rule, L.px+L.pw-2-(L.mx+11), false)
	sc.drawMenu(g, pal, v, true)

	// The foot: the edition and the version on the left, the keys on the
	// right.
	room := L.px + L.pw - L.mx
	var feet []string
	for _, ver := range v.versions {
		if v.edition != "" {
			feet = append(feet, "the "+v.edition+" edition · "+ver)
		}
	}
	feet = append(feet, v.versions...)
	if runeLen(menuKeysHint) <= room {
		g.text(L.px+L.pw-runeLen(menuKeysHint), L.foot, menuKeysHint, pal.dim, false)
		room -= runeLen(menuKeysHint) + 2
	}
	if foot := fitOption(feet, room); foot != "" {
		g.text(L.mx, L.foot, foot, pal.dim, false)
	}
}

// drawElite draws the forge master's line centred in the w cells from x on
// row y: over the box's border on the town page (boxed), on the rule on
// the forge page.
func (sc *menuScene) drawElite(g *mGrid, pal *menuPalette, v *menuView, x, y, w int, boxed bool) {
	if v.elite[1] == "" {
		return
	}
	n := runeLen(v.elite[0]) + runeLen(v.elite[1]) + runeLen(v.elite[2]) + 4
	if n > w {
		return
	}
	at := x + w - n
	ground := pal.bg
	if boxed {
		at, ground = x+centred(w, n), pal.ground
	}
	at += g.textOn(at, y, " "+v.elite[0]+" ", pal.accent, ground, true)
	at += g.textOn(at, y, v.elite[1], pal.label, ground, true)
	g.textOn(at, y, " "+v.elite[2]+" ", pal.accent, ground, true)
}

// drawMenu draws the entries from (L.mx, L.my). With leaders set it is a
// table of contents: a row of dots runs from each name to its key, and the
// selected entry is marked and lit. Without, each entry is a band across
// the box and the selected one is filled.
func (sc *menuScene) drawMenu(g *mGrid, pal *menuPalette, v *menuView, leaders bool) {
	L := sc.L
	x, w, row := L.mx, L.mw, L.my
	for i, it := range v.items {
		on := i == v.sel
		left := "  " + it.label
		if on {
			left = "▸ " + it.label
		}
		llen := runeLen(left)
		if leaders {
			ink, soft, dots := pal.ink, pal.dim, pal.faint
			if on {
				ink, soft, dots = pal.accent, pal.ink, pal.dim
			}
			extra := fitOption(it.extras, w-llen-8)
			rlen := 1
			if extra != "" {
				rlen += runeLen(extra) + 2
			}
			g.text(x, row, left, ink, on)
			for d := x + llen + 1; d < x+w-rlen-1; d++ {
				g.put(d, row, '.', dots)
			}
			if extra != "" {
				g.text(x+w-rlen, row, extra, soft, false)
			}
			g.text(x+w-1, row, string(it.key), ternaryColor(on, pal.accent, pal.dim), on)
		} else {
			band, ink, soft := pal.ground, pal.groundInk, pal.groundDim
			if on {
				band, ink, soft = pal.band, pal.onBand, pal.onBand
			}
			g.fill(x, row, w, 1, ink, band)
			g.textOn(x, row, left, ink, band, on)
			if extra := fitOption(it.extras, w-llen-6); extra != "" {
				g.textOn(x+w-3-runeLen(extra)-1, row, extra, soft, band, false)
			}
			g.textOn(x+w-2, row, string(it.key), soft, band, on)
		}
		row++
		if len(it.details) > 0 {
			if d := fitOption(it.details, w-4); d != "" {
				g.text(x+4, row, d, ternaryColor(leaders, pal.dim, pal.groundDim), false)
			}
			row++
		}
		row += L.gap - 1
	}
}

func ternaryColor(cond bool, a, b tcell.Color) tcell.Color {
	if cond {
		return a
	}
	return b
}

// renderCompact draws the page for a terminal under the game's least
// size: the name and the list, nothing that moves.
func (sc *menuScene) renderCompact(g *mGrid, pal *menuPalette, v *menuView) {
	w := min(g.w, 44)
	x := centred(g.w, w)
	title := "AGEFORGE"
	if runeLen(title) <= g.w {
		g.text(centred(g.w, runeLen(title)), 0, title, pal.accent, true)
	}
	for i, it := range v.items {
		y := 2 + i
		if y >= g.h {
			break
		}
		on := i == v.sel
		left := truncate("  "+it.label, max(w-2, 0))
		if on {
			left = truncate("▸ "+it.label, max(w-2, 0))
		}
		g.text(x, y, left, ternaryColor(on, pal.accent, pal.ink), on)
		if w >= runeLen(left)+2 {
			g.text(x+w-1, y, string(it.key), pal.dim, false)
		}
	}
	if y := 3 + len(v.items); y < g.h {
		if ver := fitOption(v.versions, w); ver != "" {
			g.text(x, y, ver, pal.dim, false)
		}
	}
}

// numberWords spells a whole number from 0 to 99 ("twenty-two"); anything
// else is written in figures.
func numberWords(n int) string {
	ones := []string{"no", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten",
		"eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
	switch {
	case n < 0 || n > 99:
		return strconv.Itoa(n)
	case n < 20:
		return ones[n]
	case n%10 == 0:
		return tens[n/10]
	}
	return tens[n/10] + "-" + ones[n%10]
}

// ordinalWords names a small ordinal ("fourth"), "" for one it has no
// word for.
func ordinalWords(n int) string {
	words := []string{"", "first", "second", "third", "fourth", "fifth", "sixth", "seventh", "eighth", "ninth", "tenth"}
	if n < 1 || n >= len(words) {
		return ""
	}
	return words[n]
}
