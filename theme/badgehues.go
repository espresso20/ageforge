package theme

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

// badgehues.go holds the badges' identity hues: the metals of the tiers and
// the colours of the hand-drawn badges. Like the maps' hues (maphues.go)
// they are never drawn as they are: the badge renderer picks the value for
// a dark or a light theme, folds it onto the ramp of a one-ink theme and
// holds it to the contrast rule for art against what it is drawn on
// (ui/cell_grid.go). Keeping them here keeps the "no raw colours outside
// theme/" rule.
//
// Every hue has two values. On a dark canvas the metals are the design's:
// bronze B87333, silver C0C0C0, gold D4AF37, platinum E5E4E2, each with a
// dark and a pale stop for a rim's gradient. On a light canvas those pale
// metals would all be darkened to one grey, so each tier has a ramp of its
// own there, and it runs the other way: the stop that stands out is the
// darkest. Silver is a neutral grey on light and platinum a steel blue, so
// the two stay apart.

// BadgeHue names a badge identity hue.
type BadgeHue uint16

const (
	BadgeHueNone BadgeHue = iota
	// The metals: dark, base and pale for each tier.
	BadgeBronzeDark
	BadgeBronze
	BadgeBronzePale
	BadgeSilverDark
	BadgeSilver
	BadgeSilverPale
	BadgeGoldDark
	BadgeGold
	BadgeGoldPale
	BadgePlatinumDark
	BadgePlatinum
	BadgePlatinumPale
	// The hand-drawn badges: a cookie jar, a ledger and a rain of code.
	BadgeCookie
	BadgeCookieDark
	BadgeLid
	BadgeLidDark
	BadgeJarLabel
	BadgeGlass
	BadgeCyan
	BadgeMagenta
	BadgeStatic
	BadgeLedgerRule
	BadgeLedgerDigit
	BadgeLedgerText
	BadgeLedgerHead
	BadgeRain0 // the head of a falling column, then its tail
	BadgeRain1
	BadgeRain2
	BadgeRain3
	BadgeRain4
	BadgeRain5
	BadgeSourceFrame
	BadgeSourceLabel
	BadgeSourceKey
	// NumBadgeHues is how many hues there are.
	NumBadgeHues
)

// badgeHues is each hue for a dark theme, then for a light one.
var badgeHues = [NumBadgeHues][2]int32{
	BadgeHueNone:      {0x808080, 0x808080},
	BadgeBronzeDark:   {0x7a4a1f, 0xb8804e},
	BadgeBronze:       {0xb87333, 0x96531c},
	BadgeBronzePale:   {0xe3a869, 0x6e3a0e},
	BadgeSilverDark:   {0x7d848c, 0x8a929b},
	BadgeSilver:       {0xc0c0c0, 0x5f6b77},
	BadgeSilverPale:   {0xf2f4f7, 0x39434d},
	BadgeGoldDark:     {0x8c6a14, 0xb8922e},
	BadgeGold:         {0xd4af37, 0x8f6a00},
	BadgeGoldPale:     {0xf4e29a, 0x6a4d00},
	BadgePlatinumDark: {0x8796a6, 0x8fa0c8},
	BadgePlatinum:     {0xe5e4e2, 0x566aa6},
	BadgePlatinumPale: {0xf2f8ff, 0x33457d},

	BadgeCookie:      {0xc68642, 0x9a5f22},
	BadgeCookieDark:  {0x8b5a2b, 0x6b3f14},
	BadgeLid:         {0xe0a95a, 0x8a5a12},
	BadgeLidDark:     {0xc98a3a, 0x70470c},
	BadgeJarLabel:    {0xf5f5f5, 0x1f2328},
	BadgeGlass:       {0x8fd3ff, 0x1f6f9c},
	BadgeCyan:        {0x39e6ff, 0x00758a},
	BadgeMagenta:     {0xff3ea5, 0xb3166b},
	BadgeStatic:      {0x39ff88, 0x0b7a3a},
	BadgeLedgerRule:  {0x2f6f93, 0x6f97b0},
	BadgeLedgerDigit: {0xffe066, 0x7a5a00},
	BadgeLedgerText:  {0x9fe8ff, 0x1d5f77},
	BadgeLedgerHead:  {0xe6f7ff, 0x0f3a4a},
	BadgeRain0:       {0xeaffea, 0x06290f},
	BadgeRain1:       {0x8dffa8, 0x0c4a1d},
	BadgeRain2:       {0x3ddc6e, 0x13672b},
	BadgeRain3:       {0x2fbf5c, 0x1b7f38},
	BadgeRain4:       {0x1c8f43, 0x2a8f47},
	BadgeRain5:       {0x126b31, 0x3b9954},
	BadgeSourceFrame: {0x1f7a3a, 0x1f7a3a},
	BadgeSourceLabel: {0x5dff8a, 0x0c5a24},
	BadgeSourceKey:   {0xc9ffd6, 0x06290f},
}

// BadgeHueColor returns a badge identity hue for a dark or a light theme.
func BadgeHueColor(h BadgeHue, light bool) tcell.Color {
	if h >= NumBadgeHues {
		h = BadgeHueNone
	}
	if light {
		return tcell.NewHexColor(badgeHues[h][1])
	}
	return tcell.NewHexColor(badgeHues[h][0])
}

// The stops of a rim's gradient.
const (
	PrismDark = iota
	PrismBase
	PrismPale
)

// PrismColor is the colour of a legendary badge's rim at an angle of hue
// degrees round its wheel, at a stop of its gradient. On a dark canvas the
// pale stop is the lightest; on a light one the same hues run deeper and
// the pale stop is the deepest.
func PrismColor(hue float64, stop int, light bool) tcell.Color {
	sl := [3][2]float64{{60, 42}, {70, 62}, {80, 84}}
	if light {
		sl = [3][2]float64{{65, 46}, {75, 36}, {85, 25}}
	}
	if stop < PrismDark || stop > PrismPale {
		stop = PrismBase
	}
	return hsl(hue, sl[stop][0], sl[stop][1])
}

// hsl is a colour from a hue in degrees and a saturation and a lightness
// in percent.
func hsl(h, s, l float64) tcell.Color {
	s, l = s/100, l/100
	a := s * math.Min(l, 1-l)
	f := func(n float64) int32 {
		k := math.Mod(n+h/30, 12)
		v := l - a*math.Max(-1, math.Min(math.Min(k-3, 9-k), 1))
		return int32(math.Round(v * 255))
	}
	return tcell.NewRGBColor(f(0), f(8), f(4))
}
