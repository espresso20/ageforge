package ui

import (
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// ink.go holds the inks of badge art: a colour that belongs to the art, by
// name, and the tables the glyph tiers fold it by.

// ink is a colour that belongs to badge art, by name: one of the theme
// package's badge hues, or a hue of the prism wheel. The painter picks its
// value for a dark or a light theme and holds it to the contrast rule
// (gridPalette.inkOn).
type ink uint16

const (
	inkNone     = ink(theme.BadgeHueNone)
	inkBronze   = ink(theme.BadgeBronze)
	inkSilver   = ink(theme.BadgeSilver)
	inkGold     = ink(theme.BadgeGold)
	inkPlatinum = ink(theme.BadgePlatinum)
	inkCyan     = ink(theme.BadgeCyan)
	inkMagenta  = ink(theme.BadgeMagenta)
	// inkPrism is the first of the prism wheel's inks: prismHues hues,
	// three stops each (dark, base, pale).
	inkPrism = ink(theme.NumBadgeHues)
)

// prismHues is how many hues the prism wheel has.
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

// The stops of a prism hue.
const (
	stopBase = theme.PrismBase
	stopPale = theme.PrismPale
)

// prismInk is the prism wheel's ink at an angle of hue degrees.
func prismInk(hue float64, stop int) ink {
	h := int(math.Round(hue/(360/prismHues))) % prismHues
	if h < 0 {
		h += prismHues
	}
	return inkPrism + ink(h*3+stop)
}

// inkValue is an ink's colour for a dark or a light theme, before the
// contrast rule.
func inkValue(k ink, light bool) kit.Color {
	if k >= inkPrism {
		i := int(k - inkPrism)
		return theme.PrismColor(float64(i/3)*(360/prismHues), i%3, light)
	}
	return theme.BadgeHueColor(theme.BadgeHue(k), light)
}

// atRest is the frame number of a badge that does not move: the motion
// setting is off, or the badge is not one the case animates.
const atRest = -1
