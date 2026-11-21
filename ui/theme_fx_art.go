package ui

import "github.com/espresso20/ageforge/theme"

// theme_fx_art.go holds the pieces of theme_fx.go that are art: the noise
// and the rune sets an effect draws with.
//
// Every frame is a pure function of the frame number. The noise is a hash
// of the frame and the cell, never a random number, so a frame always
// looks the same and drawing one draws nothing from the game's streams.

// artHash is the effects' noise: a number in [0, 1) from three integers.
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

func pickRune(set []rune, v float64) rune { return set[int(v*float64(len(set)))%len(set)] }

// The glints of a prism cycle through these.
var glintRunes = [4]rune{'·', '✧', '✦', '✧'}

var (
	// rainGlyphs is what falls: half-width kana and digits, or letters and
	// digits in the plain tier.
	rainGlyphs = []rune("ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ0123456789")
	inkRain0   = ink(theme.BadgeRain0)
	rainInks   = [6]ink{inkRain0, inkRain0 + 1, inkRain0 + 2, inkRain0 + 3, inkRain0 + 4, inkRain0 + 5}
)
