package ui

import (
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// medal.go is the one badge renderer. The Research panel's tech badges and
// the account's badges are drawn from the same pieces: a frame, a rim, a
// ribbon and an emblem on a cell grid (cell_grid.go).
//
//   - Frames: a box of three rows and an octagon of five, built from a set
//     of corner and edge runes (round, double, dashed), and a ring of half
//     blocks and shades.
//   - A badge is a sprite and a role map the same size: each cell of the
//     map says what the cell is (frame, a stop of the rim, ribbon, emblem,
//     a decoration, a glint), and one painter turns roles into inks for
//     every tier. A tier is a sprite and a metal, nothing else.
//   - Every frame of an animation is a pure function of the badge and the
//     frame number: nothing here keeps state or draws a random number.

// frameRunes is a badge frame: the four corners, then the flat and the
// upright edge.
type frameRunes struct{ tl, tr, bl, br, h, v rune }

var (
	frameRound  = frameRunes{'╭', '╮', '╰', '╯', '─', '│'}
	frameDouble = frameRunes{'╔', '╗', '╚', '╝', '═', '║'}
)

// dashed is f with dashed edges: a badge that waits.
func (f frameRunes) dashed() frameRunes {
	f.h, f.v = '┄', '┆'
	return f
}

// flat is n cells of f's flat edge.
func (f frameRunes) flat(n int) string { return strings.Repeat(string(f.h), n) }

// boxRows is a boxed badge, three rows: top between the upper corners, mid
// between the uprights and foot between the lower corners.
func boxRows(f frameRunes, top, mid, foot string) []string {
	return []string{
		string(f.tl) + top + string(f.tr),
		string(f.v) + mid + string(f.v),
		string(f.bl) + foot + string(f.br),
	}
}

// octagonRows is an eight-sided badge, five rows and nine cells wide: in1
// and in3 (five cells) are the rows above and below the middle, mid (seven
// cells) the middle and foot (five cells) the bottom edge.
func octagonRows(f frameRunes, in1, mid, in3, foot string) []string {
	return []string{
		" " + string(f.tl) + f.flat(5) + string(f.tr) + " ",
		string(f.tl) + string(f.br) + in1 + string(f.bl) + string(f.tr),
		string(f.v) + mid + string(f.v),
		string(f.bl) + string(f.tr) + in3 + string(f.tl) + string(f.br),
		" " + string(f.bl) + foot + string(f.br) + " ",
	}
}

// medalSprite is a badge at rest and what each of its cells is:
//
//	F frame      b, p, d  the rim's stops (base, pale, dark)
//	R ribbon     E        the emblem's cell
//	S a decoration beside the emblem
//	G a glint on the rim
type medalSprite struct{ rows, roles []string }

func (s medalSprite) size() (w, h int) {
	if len(s.rows) == 0 {
		return 0, 0
	}
	return len([]rune(s.rows[0])), len(s.rows)
}

// emblemCell is where the sprite's emblem sits.
func (s medalSprite) emblemCell() (x, y int) {
	for y, row := range s.roles {
		if x := strings.IndexByte(row, 'E'); x >= 0 {
			return x, y
		}
	}
	return 0, 0
}

// The tier sprites. Bronze is the box, silver the octagon; gold and
// platinum are the small ring, platinum with a star at each corner of the
// rim; legendary is the large ring.
var (
	spriteBronze = medalSprite{
		rows:  boxRows(frameRound, "───", "   ", "─▾─"),
		roles: []string{"FFFFF", "F E F", "FFRFF"},
	}
	spriteSilver = medalSprite{
		rows:  octagonRows(frameRound, " · · ", "       ", " · · ", "┬───┬"),
		roles: []string{" FFFFFFF ", "FF S S FF", "F   E   F", "FF S S FF", " FRFFFRF "},
	}
	spriteGold = medalSprite{
		rows: []string{
			"   ▄▓▓▓▓▓▄   ",
			" ▄▓▒░░░░░▒▓▄ ",
			"▐▓░  ✦ ✦  ░▓▌",
			"▐▓░       ░▓▌",
			"▐▓░  ✦ ✦  ░▓▌",
			" ▀▓▒░░░░░▒▓▀ ",
			"   ▀█▀ ▀█▀   ",
		},
		roles: []string{
			"   FbbbbbF   ",
			" FbpdddddpbF ",
			"Fbd  S S  dbF",
			"Fbd   E   dbF",
			"Fbd  S S  dbF",
			" FbpdddddpbF ",
			"   RRR RRR   ",
		},
	}
	spritePlatinum = medalSprite{
		rows: []string{
			"   ▄▓▓▓▓▓▄   ",
			" ✦▓▒░░░░░▒▓✦ ",
			"▐▓░  ✦ ✦  ░▓▌",
			"▐▓░       ░▓▌",
			"▐▓░  ✦ ✦  ░▓▌",
			" ✦▓▒░░░░░▒▓✦ ",
			"   ▀█▀ ▀█▀   ",
		},
		roles: []string{
			"   FbbbbbF   ",
			" GbpdddddpbG ",
			"Fbd  S S  dbF",
			"Fbd   E   dbF",
			"Fbd  S S  dbF",
			" GbpdddddpbG ",
			"   RRR RRR   ",
		},
	}
	spriteLegendary = medalSprite{
		rows: []string{
			"    ▄▄█████▄▄    ",
			"  ▄█▓▒░░░░░▒▓█▄  ",
			" █▓░ ✧     · ░▓█ ",
			"▐█▒    ✦ ✦    ▒█▌",
			"▐█▒           ▒█▌",
			"▐█▒    ✦ ✦    ▒█▌",
			" █▓░ ·     ✧ ░▓█ ",
			"  ▀█▓▒░░░░░▒▓█▀  ",
			"    ▀▀█▀▀▀█▀▀    ",
		},
		roles: []string{
			"    FFFFFFFFF    ",
			"  FFbpdddddpbFF  ",
			" Fbd G     G dbF ",
			"FFp    S S    pFF",
			"FFp     E     pFF",
			"FFp    S S    pFF",
			" Fbd G     G dbF ",
			"  FFbpdddddpbFF  ",
			"    FFRFFFRFF    ",
		},
	}
	// spriteStatic is an integrity badge without a sprite of its own: a
	// frame of static. (The three there are each have a sprite; this is
	// what a fourth would wear until it is drawn.)
	spriteStatic = medalSprite{
		rows:  []string{"▚▞▚▞▚▞▚▞▚", "▞       ▚", "▚       ▞", "▞       ▚", "▚▞▚▞▚▞▚▞▚"},
		roles: []string{"FFFFFFFFF", "F       F", "F   E   F", "F       F", "FFFFFFFFF"},
	}
	// spriteSlab is a hidden badge at full size: a slab that gives nothing
	// away, not even a tier.
	spriteSlab = medalSprite{
		rows:  []string{"▗▄▄▄▄▄▄▄▖", "▐███████▌", "▐███████▌", "▐███████▌", "▝▀▀▀▀▀▀▀▘"},
		roles: []string{"HHHHHHHHH", "HHHHHHHHH", "HHHHHHHHH", "HHHHHHHHH", "HHHHHHHHH"},
	}
)

// tierSprite is a tier's sprite; the bronze box for anything else.
func tierSprite(t config.BadgeTier) medalSprite {
	switch t {
	case config.BadgeSilver:
		return spriteSilver
	case config.BadgeGold:
		return spriteGold
	case config.BadgePlatinum:
		return spritePlatinum
	case config.BadgeLegendary:
		return spriteLegendary
	}
	return spriteBronze
}

// ringRows is the large ring with nothing on it, 17 by 9: the legendary
// badge's frame, and the ring of a tech's card. Its foot mirrors its top.
func ringRows() []string {
	out := make([]string, 0, len(spriteLegendary.rows))
	for y, row := range spriteLegendary.rows {
		rs := []rune(row)
		for x, role := range spriteLegendary.roles[y] {
			switch role {
			case 'G', 'S', 'E':
				rs[x] = ' '
			}
		}
		out = append(out, string(rs))
	}
	out[len(out)-1] = strings.Map(func(r rune) rune {
		if r == '▄' {
			return '▀'
		}
		return r
	}, out[0])
	return out
}

// The mini badges, 5 by 3: every badge in the case's grid. The frame is
// the tier, so two tiers differ in shape before they differ in colour.
var (
	miniBronze    = boxRows(frameRound, "───", "   ", "─▾─")
	miniSilver    = boxRows(frameDouble, "═══", "   ", "═▾═")
	miniGold      = []string{"▛▀▀▀▜", "▌   ▐", "▙▄▄▄▟"}
	miniPlatinum  = []string{"✦▀▀▀✦", "▌   ▐", "✦▄▄▄✦"}
	miniLegendary = []string{"▗▟█▙▖", "█   █", "▝▜█▛▘"}
	miniLocked    = boxRows(frameRound.dashed(), "┄┄┄", " ? ", "┄┄┄")
	miniHidden    = []string{"▗▄▄▄▖", "▐███▌", "▝▀▀▀▘"}
	// miniIntegrity is an integrity badge: a frame of static.
	miniIntegrity = []string{"▚▞▚▞▚", "▞   ▚", "▚▞▚▞▚"}
)

const (
	miniW, miniH = 5, 3
)

func tierMini(t config.BadgeTier) []string {
	switch t {
	case config.BadgeBronze:
		return miniBronze
	case config.BadgeSilver:
		return miniSilver
	case config.BadgeGold:
		return miniGold
	case config.BadgePlatinum:
		return miniPlatinum
	case config.BadgeLegendary:
		return miniLegendary
	}
	return miniIntegrity
}

// ----- inks -----

// ink is a colour that belongs to badge art, by name: one of the theme
// package's badge hues, or a hue of the legendary rim's wheel. The painter
// picks its value for a dark or a light theme and holds it to the contrast
// rule (gridPalette.inkOn).
type ink uint16

const (
	inkNone        = ink(theme.BadgeHueNone)
	inkBronze      = ink(theme.BadgeBronze)
	inkSilver      = ink(theme.BadgeSilver)
	inkGold        = ink(theme.BadgeGold)
	inkPlatinum    = ink(theme.BadgePlatinum)
	inkCookie      = ink(theme.BadgeCookie)
	inkCookieDark  = ink(theme.BadgeCookieDark)
	inkLid         = ink(theme.BadgeLid)
	inkLidDark     = ink(theme.BadgeLidDark)
	inkJarLabel    = ink(theme.BadgeJarLabel)
	inkGlass       = ink(theme.BadgeGlass)
	inkCyan        = ink(theme.BadgeCyan)
	inkMagenta     = ink(theme.BadgeMagenta)
	inkStatic      = ink(theme.BadgeStatic)
	inkLedgerRule  = ink(theme.BadgeLedgerRule)
	inkLedgerDigit = ink(theme.BadgeLedgerDigit)
	inkLedgerText  = ink(theme.BadgeLedgerText)
	inkLedgerHead  = ink(theme.BadgeLedgerHead)
	inkRain0       = ink(theme.BadgeRain0)
	inkSourceFrame = ink(theme.BadgeSourceFrame)
	inkSourceLabel = ink(theme.BadgeSourceLabel)
	inkSourceKey   = ink(theme.BadgeSourceKey)
	// inkPrism is the first of the legendary rim's inks: prismHues hues,
	// three stops each (dark, base, pale).
	inkPrism = ink(theme.NumBadgeHues)
)

// prismHues is how many hues the legendary rim's wheel has.
const prismHues = 36

// The metal inks of a tier: its dark, base and pale stops.
func tierInks(t config.BadgeTier) (dark, base, pale ink) {
	b := theme.BadgeBronze
	switch t {
	case config.BadgeSilver:
		b = theme.BadgeSilver
	case config.BadgeGold:
		b = theme.BadgeGold
	case config.BadgePlatinum:
		b = theme.BadgePlatinum
	}
	return ink(b - 1), ink(b), ink(b + 1)
}

// The stops of a rim.
const (
	stopDark = theme.PrismDark
	stopBase = theme.PrismBase
	stopPale = theme.PrismPale
)

// prismInk is the legendary rim's ink at an angle of hue degrees.
func prismInk(hue float64, stop int) ink {
	h := int(math.Round(hue/(360/prismHues))) % prismHues
	if h < 0 {
		h += prismHues
	}
	return inkPrism + ink(h*3+stop)
}

// inkValue is an ink's colour for a dark or a light theme, before the
// contrast rule.
func inkValue(k ink, light bool) tcell.Color {
	if k >= inkPrism {
		i := int(k - inkPrism)
		return theme.PrismColor(float64(i/3)*(360/prismHues), i%3, light)
	}
	return theme.BadgeHueColor(theme.BadgeHue(k), light)
}

// ----- the painter -----

// medalState is whether a badge is earned, in sight but not earned, or
// withheld by the spoiler rules.
type medalState uint8

const (
	medalEarned medalState = iota
	medalLocked
	medalHidden
)

// medal is what the renderer needs to know of a badge.
type medal struct {
	// tier is the badge's tier; BadgeNoTier is an integrity badge.
	tier   config.BadgeTier
	emblem rune
	state  medalState
	// crossed: earned in a modified game. A strike goes through it.
	crossed bool
	// special names a hand-drawn sprite (badge_art_special.go), "" for
	// none.
	special string
	// phase is the badge's own offset into the animations, so two badges
	// side by side do not move as one.
	phase int
}

// atRest is the frame number of a badge that does not move: the motion
// setting is off, or the badge is not one the case animates.
const atRest = -1

// The glints of a rim cycle through these, and the legendary frame steps
// through these stops.
var (
	glintRunes = [4]rune{'·', '✧', '✦', '✧'}
	pulseStops = [4]int{stopBase, stopPale, stopBase, stopDark}
)

// glintCorner is the phase of the glint in a corner of a w by h sprite.
func glintCorner(x, y, w, h int) int {
	top, left := float64(y) < float64(h-1)/2, float64(x) < float64(w-1)/2
	switch {
	case top && left:
		return 1
	case top, left:
		return 0
	}
	return 3
}

// prismHue is the hue of the legendary rim at a cell: the wheel turns
// round the middle of the sprite, three degrees a frame.
func prismHue(x, y, w, h, n int) float64 {
	a := math.Atan2((float64(y)-float64(h-1)/2)*2.1, float64(x)-float64(w-1)/2) * 180 / math.Pi
	return math.Mod(math.Mod(math.Round(a+140+float64(n)*3), 360)+360, 360)
}

// medalSize is the size of a badge's full art.
func medalSize(m medal) (w, h int) {
	switch {
	case m.state == medalHidden:
		return spriteSlab.size()
	case m.special != "" && m.state == medalEarned:
		return specialW, specialH
	case m.tier == config.BadgeNoTier:
		return spriteStatic.size()
	}
	return tierSprite(m.tier).size()
}

// drawMedal draws a badge's full art with its top left corner at (x, y).
// n is the frame number, or atRest. plain draws for the plain glyph tier:
// the caller still folds the grid, but what moves by a change of colour
// alone moves by a change of glyph instead.
func drawMedal(g *tGrid, x, y int, m medal, n int, plain bool) {
	switch {
	case m.state == medalHidden:
		for dy, row := range spriteSlab.rows {
			g.text(x, y+dy, row, tsSlab, -1)
		}
		return
	case m.special != "" && m.state == medalEarned:
		drawSpecial(g, x, y, m, n, plain)
		return
	}
	sp := tierSprite(m.tier)
	static := m.tier == config.BadgeNoTier
	if static {
		sp = spriteStatic
	}
	w, h := sp.size()
	ex, ey := sp.emblemCell()
	if m.state == medalLocked {
		// The tier's shape, hollow: what there is to earn.
		for dy, row := range sp.rows {
			for dx, r := range []rune(row) {
				switch sp.roles[dy][dx] {
				case ' ', 'S', 'G':
					continue
				case 'E':
					r = '?'
				}
				g.put(x+dx, y+dy, r, tsDim, -1)
			}
		}
		return
	}
	frame := max(n, 0) + m.phase
	if n == atRest {
		frame = 0
	}
	legend := m.tier == config.BadgeLegendary
	dark, base, pale := tierInks(m.tier)
	k := frame%24 - 8
	sweep := legend && n != atRest && k >= 0 && k < 8
	mid := 4 + float64(k)*24/7
	pulse := pulseStops[(frame/4)%4]
	for dy, row := range sp.rows {
		for dx, r := range []rune(row) {
			role := sp.roles[dy][dx]
			if role == ' ' {
				continue
			}
			px, py := x+dx, y+dy
			rim := strings.IndexByte("FbpdR", role) >= 0
			if !legend {
				col := base
				if static {
					// Static has no metal: two colours, split down the middle.
					col = inkCyan
					if dx > w/2 {
						col = inkMagenta
					}
					if role == 'E' {
						g.put(px, py, m.emblem, tsBright, -1)
						continue
					}
				}
				switch role {
				case 'p':
					col = pale
				case 'd':
					col = dark
				case 'E':
					g.paint(px, py, m.emblem, pale, tsBright, cfBold)
					continue
				case 'S':
					if m.tier == config.BadgeBronze || m.tier == config.BadgeSilver || m.tier == config.BadgeNoTier {
						col = pale
					}
				case 'R':
					if m.tier == config.BadgeGold || m.tier == config.BadgePlatinum {
						col = dark
					}
				case 'G':
					// Platinum's corner stars glint; at rest they are stars.
					if n != atRest {
						r = glintRunes[(frame/3+glintCorner(dx, dy, w, h))%4]
					}
					switch r {
					case '·':
						col = dark
					case '✦':
						col = pale
					}
				}
				g.paint(px, py, r, col, tsText, 0)
				continue
			}
			// The legendary ring: a wheel of hues that turns, a frame that
			// pulses, glints at four corners and a band of light that
			// crosses the rim.
			hue := prismHue(dx, dy, w, h, frame)
			if rim && sweep && math.Abs(float64(dx)+2*float64(dy)-mid) < 1 {
				if plain {
					r = '@'
				} else {
					r = '█'
				}
				g.put(px, py, r, tsBright, -1)
				continue
			}
			switch role {
			case 'F':
				g.paint(px, py, r, prismInk(hue, pulse), tsText, 0)
			case 'R':
				g.paint(px, py, r, prismInk(hue, stopDark), tsText, 0)
			case 'b':
				g.paint(px, py, r, prismInk(hue, stopBase), tsText, 0)
			case 'p':
				g.paint(px, py, r, prismInk(hue, stopPale), tsText, 0)
			case 'd':
				g.paint(px, py, r, prismInk(hue, stopDark), tsText, 0)
			case 'E':
				g.put(px, py, m.emblem, tsBright, -1)
			case 'S':
				if (frame/4)%2 == 1 {
					r = '✧'
				}
				g.paint(px, py, r, prismInk(hue, stopPale), tsText, 0)
			case 'G':
				r = glintRunes[(frame/3+glintCorner(dx, dy, w, h))%4]
				switch r {
				case '·':
					g.paint(px, py, r, prismInk(hue, stopBase), tsText, 0)
				case '✧':
					g.paint(px, py, r, prismInk(hue, stopPale), tsText, 0)
				default:
					g.put(px, py, r, tsBright, -1)
				}
			}
		}
	}
	if m.crossed {
		strike(g, x, y, w, h, ex, ey)
	}
}

// strike draws the strike of a badge earned in a modified game: a
// diagonal in the negative colour from the upper right to the lower left
// of a w by h badge, stepping round the emblem's cell.
func strike(g *tGrid, x, y, w, h, ex, ey int) {
	for dy := 0; dy < h; dy++ {
		dx := int(math.Round(0.75*float64(w-1) - float64(dy)*0.5*float64(w-1)/float64(max(h-1, 1))))
		if dx == ex && dy == ey {
			dx++
		}
		g.put(x+dx, y+dy, '╱', tsBad, -1)
	}
}

// drawMini draws a badge at the grid's size, 5 by 3, at (x, y). Only the
// legendary mini moves: its frame turns through the hues.
func drawMini(g *tGrid, x, y int, m medal, n int) {
	switch m.state {
	case medalHidden:
		for dy, row := range miniHidden {
			g.text(x, y+dy, row, tsSlab, -1)
		}
		return
	case medalLocked:
		for dy, row := range miniLocked {
			g.text(x, y+dy, row, tsDim, -1)
		}
		return
	}
	rows := tierMini(m.tier)
	frame := max(n, 0) + m.phase
	if n == atRest {
		frame = 0
	}
	_, base, pale := tierInks(m.tier)
	for dy, row := range rows {
		for dx, r := range []rune(row) {
			px, py := x+dx, y+dy
			if dx == miniW/2 && dy == miniH/2 {
				if m.tier == config.BadgeLegendary || m.tier == config.BadgeNoTier {
					g.put(px, py, m.emblem, tsBright, -1)
				} else {
					g.paint(px, py, m.emblem, pale, tsBright, cfBold)
				}
				continue
			}
			if r == ' ' {
				continue
			}
			switch m.tier {
			case config.BadgeLegendary:
				g.paint(px, py, r, prismInk(prismHue(dx, dy, miniW, miniH, frame), pulseStops[(frame/4)%4]), tsText, 0)
			case config.BadgeNoTier:
				col := inkCyan
				if dx > miniW/2 || dx == miniW/2 && dy == miniH-1 {
					col = inkMagenta
				}
				g.paint(px, py, r, col, tsText, 0)
			default:
				col := base
				if r == '✦' {
					col = pale
				}
				g.paint(px, py, r, col, tsText, 0)
			}
		}
	}
	if m.crossed {
		strike(g, x, y, miniW, miniH, miniW/2, miniH/2)
	}
}

// toastMark is a badge in three cells, for the one row of the toast bar:
// the emblem between two half discs. The caller colours it.
func toastMark(emblem rune) string { return "◖" + string(emblem) + "◗" }

// ----- the plain glyph tier -----

// badgeFold is what the plain tier draws for badge art. The map model's
// own fold would turn two stops of a rim into the same mark and a rounded
// corner into a plus, so badges fold by their own table: a rim reads # = :
// from the outside in, a top edge _ and a bottom edge ', a round corner .
// above and ' below.
var badgeFold = map[rune]rune{
	'█': '#', '▓': '#', '▒': '=', '░': ':', '▄': '_', '▀': '\'', '▐': '|', '▌': '|',
	'╭': '.', '╮': '.', '╰': '\'', '╯': '\'', '─': '-', '│': '|', '▾': 'v',
	'✦': '*', '·': '.', '✧': '+', '◆': '*', '╱': '/', '═': '=', '║': '|',
	'◂': '<', '▸': '>', '←': '<', '→': '>', '↑': '^', '↓': 'v',
	'┌': '+', '┐': '+', '└': '+', '┘': '+', '├': '+', '┤': '+', '┬': '|', '┴': '|',
	'┄': '.', '┆': ':', '▗': '.', '▖': '.', '▝': '\'', '▘': '\'',
	'▛': '#', '▜': '#', '▙': '#', '▟': '#', '╔': '.', '╗': '.', '╚': '\'', '╝': '\'',
	'✶': '*', '◖': '(', '◗': ')', '…': '.', '✓': 'v', '★': '*', '☆': 'o', '▲': '^', '▼': 'v',
	// The hand-drawn sprites, and the marks their noise throws up.
	'●': 'o', '∞': '8', '▚': '#', '▞': '#', '╴': '-', '¿': '?', '§': 'S',
}

// foldBadges rewrites a grid of badge art for the plain glyph tier.
func foldBadges(g *tGrid) { foldGrid(g, badgeFold) }
