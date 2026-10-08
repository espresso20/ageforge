package ui

import "math"

// badge_art_legend.go holds the hand-drawn badges of the rarest few: the
// legendary specials and the Transcendent Age. Each is a sprite of 17 by 9
// and a role map the same size, like the tier sprites of medal.go, and one
// painter draws them all: the prism turns through whatever is marked as
// prism, a band of light crosses the metal now and then, and the stars
// twinkle. At rest (the motion setting off) a sprite is its own drawing in
// the hues of frame 0.
//
// A role says what a cell is:
//
//	p, P, q   the prism, at its base, pale and dark stop
//	g, G      gold, and gold's pale stop
//	s         silver          l  platinum          b  bronze
//	c, m      cyan, magenta
//	w         the brightest ink of the theme (lettering)
//	d         the theme's dim ink
//	T         a star that twinkles
//
// Every frame is a pure function of the frame number.

// The sprites' keys (config.BadgeDef.Emblem names them through the emblem
// table, badge_emblems.go).
const (
	specialChains    = "legend.chains"
	specialBoxes     = "legend.boxes"
	specialMuseum    = "legend.museum"
	specialWonders   = "legend.wonders"
	specialRogues    = "legend.rogues"
	specialEndings   = "legend.endings"
	specialUndying   = "legend.undying"
	specialTranscend = "legend.transcend"
)

// legendSprite is a hand-drawn legendary badge: its drawing and what each
// cell of it is.
type legendSprite struct{ rows, roles [][]rune }

func legend(rows, roles []string) legendSprite {
	return legendSprite{rows: rows17(rows...), roles: rows17(roles...)}
}

// legendSprites are the sprites by key.
var legendSprites = map[string]legendSprite{
	// Six for Six: six links, every one closed.
	specialChains: legend(
		[]string{
			"  ╭───╮   ╭───╮  ",
			"  │ ◆ ├───┤ ◆ │  ",
			"  ╰─┬─╯   ╰─┬─╯  ",
			"  ╭─┴─╮   ╭─┴─╮  ",
			"  │ ◆ ├───┤ ◆ │  ",
			"  ╰─┬─╯   ╰─┬─╯  ",
			"  ╭─┴─╮   ╭─┴─╮  ",
			"  │ ◆ ├───┤ ◆ │  ",
			"  ╰───╯   ╰───╯  ",
		},
		[]string{
			"  ppppp   ppppp  ",
			"  p G pgggp G p  ",
			"  ppgpp   ppgpp  ",
			"  ppgpp   ppgpp  ",
			"  p G pgggp G p  ",
			"  ppgpp   ppgpp  ",
			"  ppgpp   ppgpp  ",
			"  p G pgggp G p  ",
			"  ppppp   ppppp  ",
		}),
	// Box Ticker: a list with nothing left on it.
	specialBoxes: legend(
		[]string{
			" ╔═════════════╗ ",
			" ║ ▣ ▣ ▣ ▣ ▣ ▣ ║ ",
			" ║ ▣ ▣ ▣ ▣ ▣ ▣ ║ ",
			" ║ ▣ ▣ ▣ ▣ ▣ ▣ ║ ",
			" ║ ▣ ▣ ▣ ▣ ▣ ▣ ║ ",
			" ║ ▣ ▣ ▣ ▣ ▣ ▣ ║ ",
			" ║ ───────── ✓ ║ ",
			" ║  ALL DONE   ║ ",
			" ╚═════════════╝ ",
		},
		[]string{
			" ppppppppppppppp ",
			" p g g g g g g p ",
			" p g g g g g g p ",
			" p g g g g g g p ",
			" p g g g g g g p ",
			" p g g g g g g p ",
			" p ddddddddd G p ",
			" p  www wwww   p ",
			" ppppppppppppppp ",
		}),
	// Museum Piece: the museum, and the one piece on show.
	specialMuseum: legend(
		[]string{
			"        ▲        ",
			"     ▄▄███▄▄     ",
			"  ▄▄█████████▄▄  ",
			"  ═════════════  ",
			"   ║ ║ ║ ║ ║ ║   ",
			"   ║ ║ ║✦║ ║ ║   ",
			"   ║ ║ ║ ║ ║ ║   ",
			"  ═════════════  ",
			" ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀ ",
		},
		[]string{
			"        P        ",
			"     ppppppp     ",
			"  ppppppppppppp  ",
			"  ggggggggggggg  ",
			"   l l l l l l   ",
			"   l l lTl l l   ",
			"   l l l l l l   ",
			"  ggggggggggggg  ",
			" qqqqqqqqqqqqqqq ",
		}),
	// Twenty-Two Wonders: three of them under one sky.
	specialWonders: legend(
		[]string{
			"   ·    ✦    ·   ",
			"        │        ",
			"    ╲   ▲   ╱    ",
			"  ▄    ▟█▙    ▄  ",
			" ▟█▙  ▟███▙  ▟█▙ ",
			" ███ ▟█████▙ ███ ",
			" █▒█ ███▒███ █▒█ ",
			"▄█▒█▄███▒███▄█▒█▄",
			"▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀",
		},
		[]string{
			"   T    T    T   ",
			"        G        ",
			"    G   P   G    ",
			"  s    ppp    g  ",
			" sss  ppppp  ggg ",
			" sss ppppppp ggg ",
			" sds pppdppp gdg ",
			"qsdsqpppdpppqgdgq",
			"qqqqqqqqqqqqqqqqq",
		}),
	// Rogues' Gallery: every one of them, framed.
	specialRogues: legend(
		[]string{
			"╭───╮ ╭───╮ ╭───╮",
			"│ ● │ │ ● │ │ ● │",
			"│▟█▙│ │▟█▙│ │▟█▙│",
			"╰───╯ ╰───╯ ╰───╯",
			"  ▾     ▾     ▾  ",
			"╭───╮ ╭───╮ ╭───╮",
			"│ ● │ │ ● │ │ ● │",
			"│▟█▙│ │▟█▙│ │▟█▙│",
			"╰───╯ ╰───╯ ╰───╯",
		},
		[]string{
			"ggggg ggggg ggggg",
			"g P g g P g g P g",
			"gpppg gpppg gpppg",
			"ggggg ggggg ggggg",
			"  d     d     d  ",
			"ggggg ggggg ggggg",
			"g P g g P g g P g",
			"gpppg gpppg gpppg",
			"ggggg ggggg ggggg",
		}),
	// Connoisseur of Endings: what every ending leaves, over the embers.
	specialEndings: legend(
		[]string{
			"     ▄▄███▄▄     ",
			"   ▄█████████▄   ",
			"  ▐███████████▌  ",
			"  ▐█▀  ▀█▀  ▀█▌  ",
			"  ▐█▄  ▄█▄  ▄█▌  ",
			"   ▀████▀████▀   ",
			"    █▌█▌█▐█▐█    ",
			"    ▀▀▀▀▀▀▀▀▀    ",
			"  ·  ✦  ·  ✦  ·  ",
		},
		[]string{
			"     ppppppp     ",
			"   ppppppppppp   ",
			"  ppppppppppppp  ",
			"  ppp  ppp  ppp  ",
			"  ppp  ppp  ppp  ",
			"   ppppppppppp   ",
			"    lllllllll    ",
			"    qqqqqqqqq    ",
			"  T  T  T  T  T  ",
		}),
	// Unkillable: a shield, and the flame it kept.
	specialUndying: legend(
		[]string{
			"  ▄▄▄▄▄▄▄▄▄▄▄▄▄  ",
			"  █▀▀▀▀▀▀▀▀▀▀▀█  ",
			"  █  ✦  ▲  ✦  █  ",
			"  █    ▟█▙    █  ",
			"  █   ▟███▙   █  ",
			"  ▐▙   ▀█▀   ▟▌  ",
			"   ▜▙   █   ▟▛   ",
			"    ▀▙▄▄█▄▄▟▀    ",
			"      ▀▀▀▀▀      ",
		},
		[]string{
			"  ppppppppppppp  ",
			"  ppppppppppppp  ",
			"  p  T  G  T  p  ",
			"  p    GGG    p  ",
			"  p   GGGGG   p  ",
			"  pp   GGG   pp  ",
			"   pp   G   pp   ",
			"    ppppGpppp    ",
			"      ppppp      ",
		}),
	// The Transcendent Age: a ring round a ring round a star.
	specialTranscend: legend(
		[]string{
			"   ·    ✦    ·   ",
			"     ▄▄▀▀▀▄▄     ",
			"  ▄▀▀  ▄▄▄  ▀▀▄  ",
			" ▐▌  ▄▀   ▀▄  ▐▌ ",
			" ▐▌  █  ✦  █  ▐▌ ",
			" ▐▌  ▀▄   ▄▀  ▐▌ ",
			"  ▀▄▄  ▀▀▀  ▄▄▀  ",
			"     ▀▀▄▄▄▀▀     ",
			"   ·    ✦    ·   ",
		},
		[]string{
			"   T    T    T   ",
			"     ppppppp     ",
			"  ppp  PPP  ppp  ",
			" pp  PP   PP  pp ",
			" pp  P  T  P  pp ",
			" pp  PP   PP  pp ",
			"  ppp  PPP  ppp  ",
			"     ppppppp     ",
			"   T    T    T   ",
		}),
}

// legendMetal reports whether a role is metal: what the band of light
// crosses.
func legendMetal(role rune) bool {
	switch role {
	case 'p', 'P', 'q', 'g', 'G', 's', 'l', 'b':
		return true
	}
	return false
}

// drawLegend draws a hand-drawn legendary badge at (x, y). frame is the
// frame number; rest holds it still. plain marks the band of light with a
// glyph, for the tier where a change of colour alone would not show.
func drawLegend(g *tGrid, x, y int, sp legendSprite, frame int, rest, plain bool) {
	k := frame%32 - 12
	sweep := !rest && k >= 0 && k < 10
	mid := float64(k) * 34 / 9
	for dy := 0; dy < specialH; dy++ {
		for dx := 0; dx < specialW; dx++ {
			role, r := sp.roles[dy][dx], sp.rows[dy][dx]
			if role == ' ' || r == ' ' {
				continue
			}
			px, py := x+dx, y+dy
			hue := prismHue(dx, dy, specialW, specialH, frame)
			if sweep && legendMetal(role) && math.Abs(float64(dx)+2*float64(dy)-mid) < 1.2 {
				if plain {
					r = '@'
				}
				g.put(px, py, r, tsBright, -1)
				continue
			}
			switch role {
			case 'p':
				g.paint(px, py, r, prismInk(hue, stopBase), tsText, 0)
			case 'P':
				g.paint(px, py, r, prismInk(hue+120, stopPale), tsText, cfBold)
			case 'q':
				g.paint(px, py, r, prismInk(hue, stopDark), tsText, 0)
			case 'g':
				g.paint(px, py, r, inkGold, tsText, 0)
			case 'G':
				g.paint(px, py, r, inkGold+1, tsText, cfBold)
			case 's':
				g.paint(px, py, r, inkSilver, tsText, 0)
			case 'l':
				g.paint(px, py, r, inkPlatinum, tsText, 0)
			case 'b':
				g.paint(px, py, r, inkBronze, tsText, 0)
			case 'c':
				g.paint(px, py, r, inkCyan, tsText, 0)
			case 'm':
				g.paint(px, py, r, inkMagenta, tsText, 0)
			case 'w':
				g.put(px, py, r, tsBright, -1)
			case 'd':
				g.put(px, py, r, tsDim, -1)
			case 'T':
				if !rest {
					r = glintRunes[(frame/3+dx+2*dy)%4]
				}
				stop := stopPale
				if r == '·' {
					stop = stopBase
				}
				g.paint(px, py, r, prismInk(hue+200, stop), tsText, 0)
			}
		}
	}
}
