package ui

// badge_art_special.go holds the hand-drawn badges: the three integrity
// badges, each a sprite of 17 by 9 with a glitch of its own. They go
// through the same grid, inks and fold as every other badge (medal.go).
//
// Every frame is a pure function of the frame number. The noise is a hash
// of the frame and the cell, never a random number, so a frame always
// looks the same and drawing one draws nothing from the game's streams.

const (
	specialW, specialH = 17, 9
)

// The sprites' keys (config.BadgeDef.Emblem names them through the emblem
// table, badge_emblems.go).
const (
	specialCookieJar = "cookie_jar"
	specialLedger    = "ledger"
	specialSource    = "source"
)

// artHash is the sprites' noise: a number in [0, 1) from three integers.
func artHash(a, b, c int) float64 {
	x := uint32(int32(a))*374761393 ^ uint32(int32(b))*668265263 ^ uint32(int32(c))*1274126177
	x = (x ^ (x >> 13)) * 1103515245
	x ^= x >> 16
	return float64(x) / 4294967296
}

// artBurst is how hard a glitch that comes round every period frames and
// lasts length of them is hitting at frame n: 1 as it starts, falling to
// 0, and 0 between bursts.
func artBurst(n, period, length, seed int) float64 {
	k := (n + seed) % period
	if k < length {
		return 1 - float64(k)/float64(length)
	}
	return 0
}

// artNoise is what a glitched cell shows.
var artNoise = []rune("▓▒░█▚▞▀▄#%&@$§01")

func pickRune(set []rune, v float64) rune { return set[int(v*float64(len(set)))%len(set)] }

func row17(s string) []rune {
	rs := []rune(s)
	for len(rs) < specialW {
		rs = append(rs, ' ')
	}
	return rs[:specialW]
}

func rows17(rows ...string) [][]rune {
	out := make([][]rune, len(rows))
	for i, r := range rows {
		out[i] = row17(r)
	}
	return out
}

// drawSpecial draws a hand-drawn badge at (x, y). n is the frame number,
// or atRest for the sprite as it stands when nothing moves.
func drawSpecial(g *tGrid, x, y int, m medal, n int, plain bool) {
	rest := n == atRest
	frame := max(n, 0) + m.phase
	if rest {
		frame = 0
	}
	switch m.special {
	case specialCookieJar:
		drawCookieJar(g, x, y, frame, rest)
	case specialLedger:
		drawLedger(g, x, y, frame, rest)
	case specialSource:
		drawSource(g, x, y, frame, plain)
	}
	if m.crossed && rest {
		strike(g, x, y, specialW, specialH, -1, -1)
	}
}

// ----- Hand in the Cookie Jar -----

var (
	jarRows = rows17(
		"      ▄▄▄▄▄      ",
		"   ▐█████████▌   ",
		"   ╭─────────╮   ",
		"  ╭╯ ●  ●  ● ╰╮  ",
		"  │  ●  ●  ●  │  ",
		"  │ ●  ●  ●   │  ",
		"  │  COOKIES  │  ",
		"  │  ●  ●  ●  │  ",
		"  ╰───────────╯  ",
	)
	// jarLabels are what the jar's label reads: its own, then what it
	// slips to in a burst, and the prompt it winks.
	jarLabels = [][]rune{
		[]rune("  COOKIES  "), []rune("  C00K1E5  "), []rune("  /dev▌    "), []rune("  sudo nom "), []rune(" ACCESS ¿? "),
	}
)

func jarInk(r, c int, ch rune) ink {
	switch {
	case ch == '●':
		if (r+c)%2 == 1 {
			return inkCookie
		}
		return inkCookieDark
	case r == 0:
		return inkLid
	case r == 1:
		return inkLidDark
	case r == 6 && c > 2 && c < 14:
		return inkJarLabel
	}
	return inkGlass
}

// drawCookieJar: a jar that will not hold still. Its label slips, rows
// tear sideways and a strike flickers through it in every burst.
func drawCookieJar(g *tGrid, x, y, n int, rest bool) {
	b := 0.0
	wink := false
	if !rest {
		b = artBurst(n, 26, 6, 15)
		wink = b == 0 && n%30 >= 18 && n%30 < 24
	}
	label := jarLabels[0]
	switch {
	case b > 0:
		label = jarLabels[1+int(artHash(n, 6, 1)*4)]
	case wink:
		label = jarLabels[2]
	}
	for r := 0; r < specialH; r++ {
		off := 0
		if b > 0 && artHash(n, r, 41) < 0.4*b+0.06 {
			off = 1
			if artHash(n, r, 42) < 0.5 {
				off = -1
			}
		}
		for c := 0; c < specialW; c++ {
			sc := c - off
			ch := ' '
			if sc >= 0 && sc < specialW {
				ch = jarRows[r][sc]
			}
			if r == 6 && sc >= 3 && sc < 14 {
				ch = label[sc-3]
			}
			if ch == ' ' {
				continue
			}
			col := jarInk(r, sc, ch)
			rate := 0.004
			if b > 0 {
				rate = 0.12*b + 0.015
			}
			if !rest && artHash(n, r*specialW+c, 19) < rate {
				ch = pickRune(artNoise, artHash(n, c, r+7))
				col = inkMagenta
				if artHash(n, c, r) < 0.5 {
					col = inkStatic
				}
			}
			if off != 0 {
				col = inkMagenta
				if off < 0 {
					col = inkCyan
				}
			}
			if ch == '▌' && r == 6 && n%4 < 2 {
				continue // the prompt's cursor blinks
			}
			g.paint(x+c, y+r, ch, col, tsText, 0)
		}
	}
	if b > 0.3 {
		for r := 0; r < specialH; r++ {
			if artHash(n, r, 99) > 0.2 {
				g.put(x+int(float64(13)-float64(r)*1.25+0.5), y+r, '╱', tsBad, -1)
			}
		}
	}
}

// ----- Creative Accounting -----

var ledgerRows = rows17(
	" ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄ ",
	"█ LEDGER    p.1 █",
	"█ ───────────── █",
	"█ ore     1,204 █",
	"█ gold   99,999 █",
	"█ wood      ??? █",
	"█ ───────────── █",
	"█ TOTAL   ∞.00  █",
	" ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀ ",
)

func ledgerInk(r, c int, ch rune) ink {
	switch {
	case ch == '█' || ch == '▄' || ch == '▀':
		if c < 8 {
			return inkCyan
		}
		return inkMagenta
	case ch == '─':
		return inkLedgerRule
	case ch >= '0' && ch <= '9', ch == ',', ch == '.', ch == '∞':
		return inkLedgerDigit
	case ch == '?':
		return inkMagenta
	case r == 1:
		return inkLedgerHead
	}
	return inkLedgerText
}

// drawLedger: a ledger with sums that cannot be. The digits churn, rows
// tear into two colours, noise eats the edges and a stamp flickers in.
func drawLedger(g *tGrid, x, y, n int, rest bool) {
	b := 0.0
	stamp := false
	if !rest {
		b = artBurst(n, 22, 5, 9)
		stamp = n%46 >= 30 && n%46 < 40
	}
	for r := 0; r < specialH; r++ {
		off := 0
		if b > 0 && artHash(n, r, 71) < 0.45*b+0.08 {
			off = 1
			if artHash(n, r, 72) < 0.5 {
				off = -1
			}
			if artHash(n, r, 73) < 0.3 {
				off *= 2
			}
		}
		for c := 0; c < specialW; c++ {
			sc := c - off
			ch := ' '
			if sc >= 0 && sc < specialW {
				ch = ledgerRows[r][sc]
			}
			if ch == ' ' {
				continue
			}
			col := ledgerInk(r, sc, ch)
			if b > 0 && ch >= '0' && ch <= '9' {
				ch = rune('0' + int(artHash(n, r*specialW+c, 5)*10))
			}
			rate := 0.008
			if b > 0 {
				rate = 0.16*b + 0.02
			}
			if !rest && artHash(n, r*specialW+c, 9) < rate {
				ch = pickRune(artNoise, artHash(n, c, r+3))
				col = inkMagenta
				if artHash(n, r, c) < 0.5 {
					col = inkCyan
				}
			}
			if off != 0 {
				col = inkMagenta
				if off < 0 {
					col = inkCyan
				}
			}
			g.paint(x+c, y+r, ch, col, tsText, 0)
		}
	}
	if stamp {
		for i, ch := range []rune("V O I D") {
			if ch == ' ' {
				continue
			}
			if artHash(n, 4, i) < 0.15 {
				ch = pickRune(artNoise, artHash(n, i, 4))
			}
			g.put(x+5+i, y+4, ch, tsBad, -1)
		}
	}
}

// ----- Touched by the Source -----

var (
	sourceTop    = row17("╭─[ SOURCE ]────╮")
	sourceBottom = row17("╰───────────────╯")
	// sourceKey is the key the rain settles into: its cells by row and
	// column.
	sourceKey = map[[2]int]rune{
		{3, 4}: '╭', {3, 5}: '─', {3, 6}: '╮',
		{4, 4}: '│', {4, 6}: '├', {4, 7}: '─', {4, 8}: '┬', {4, 9}: '─', {4, 10}: '┬', {4, 11}: '─', {4, 12}: '╴',
		{5, 4}: '╰', {5, 5}: '─', {5, 6}: '╯',
	}
	// rainGlyphs is what falls: half-width kana and digits, or letters and
	// digits in the plain tier.
	rainGlyphs      = []rune("ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ0123456789")
	rainGlyphsPlain = []rune("0123456789ABCDEFHJKLMNPRSTUVXYZ")
	rainInks        = [6]ink{inkRain0, inkRain1, inkRain2, inkRain3, inkRain4, inkRain5}
)

// rainSet is the rain's glyphs for a glyph tier.
func rainSet(plain bool) []rune {
	if plain {
		return rainGlyphsPlain
	}
	return rainGlyphs
}

// drawSource: code rain inside a ring marked SOURCE. Every few seconds it
// settles into a key, then the rain takes it back. At rest it shows the
// key.
func drawSource(g *tGrid, x, y, n int, plain bool) {
	set := rainSet(plain)
	showKey := n%72 < 18
	for r := 0; r < specialH; r++ {
		for c := 0; c < specialW; c++ {
			px, py := x+c, y+r
			if r == 0 || r == specialH-1 {
				ch := sourceTop[c]
				if r != 0 {
					ch = sourceBottom[c]
				}
				if r == 0 && c >= 2 && c <= 11 {
					col := inkSourceLabel
					if artHash(n>>1, c, 1) < 0.08 {
						col = inkRain0
					}
					g.paint(px, py, ch, col, tsText, cfBold)
				} else {
					g.paint(px, py, ch, inkSourceFrame, tsText, 0)
				}
				continue
			}
			if c == 0 || c == specialW-1 {
				g.paint(px, py, '│', inkSourceFrame, tsText, 0)
				continue
			}
			if k, ok := sourceKey[[2]int{r, c}]; ok && showKey {
				g.paint(px, py, k, inkSourceKey, tsText, cfBold)
				continue
			}
			speed := 1 + int(artHash(c, 1, 2)*3)
			const length = 16
			head := (n*speed/3+int(artHash(c, 3, 4)*length))%length - 1
			if d := head - r; d >= 0 && d < len(rainInks) {
				fl := uint8(0)
				if d == 0 {
					fl = cfBold
				}
				g.paint(px, py, pickRune(set, artHash((n>>1)+d, r, c)), rainInks[d], tsText, fl)
			}
		}
	}
}
