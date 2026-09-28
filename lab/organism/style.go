package main

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// The organism has one rendering grammar. An epoch changes a handful of
// dials, never the grammar: the wood's hue, how branches are drawn (grown,
// forged, routed as circuit traces, or strung as constellations), what a
// leaf is (foliage, a solder pad, a star), and whether the sky holds stars.

type branchStyle int

const (
	styleGrown   branchStyle = iota // curved, organic
	styleForged                     // straighter, stiffer limbs
	styleTrace                      // 45-degree circuit traces
	styleStellar                    // dotted constellation lines
)

type leafStyle int

const (
	leafFoliage leafStyle = iota
	leafPad
	leafStar
)

type dial struct {
	Wood   uint32 // bark hue
	Hue    uint32 // the epoch's identity hue: soil, labels
	Branch branchStyle
	Leaf   leafStyle
	Stars  bool // stars in the sky by day too
	Soil   string
}

var epochDials = map[string]dial{
	"stone_era":    {Wood: 0x8b6b4a, Hue: 0xa89878, Branch: styleGrown, Leaf: leafFoliage, Soil: "stone"},
	"iron_era":     {Wood: 0x7c5a44, Hue: 0xa8645a, Branch: styleGrown, Leaf: leafFoliage, Soil: "iron"},
	"steel_era":    {Wood: 0x7d746c, Hue: 0x8c93a0, Branch: styleForged, Leaf: leafFoliage, Soil: "steel"},
	"electric_era": {Wood: 0x9a8458, Hue: 0xd0ae4c, Branch: styleForged, Leaf: leafFoliage, Soil: "electric"},
	"digital_era":  {Wood: 0x4f93ad, Hue: 0x3fa4cc, Branch: styleTrace, Leaf: leafPad, Soil: "digital"},
	"neon_era":     {Wood: 0xb45cf0, Hue: 0xc056dc, Branch: styleTrace, Leaf: leafPad, Stars: true, Soil: "neon"},
	"cosmic_era":   {Wood: 0xcfd6ff, Hue: 0x7466ff, Branch: styleStellar, Leaf: leafStar, Stars: true, Soil: "cosmic"},
}

func dialFor(epoch string) dial {
	if d, ok := epochDials[epoch]; ok {
		return d
	}
	return epochDials["stone_era"]
}

// palette is every colour one frame uses, resolved from the active theme.
// Chrome, sky, text and signals come from theme roles; the only fixed hues
// are identities (lineages, epochs, bark), and each goes through
// theme.Legible so it clears contrast on the theme's background, darker on
// light themes and untouched on dark ones.
type palette struct {
	Bg, Text, Dim, Accent, Highlight, Positive, Negative, Warning, Surface, Border tcell.Color
	Light                                                                          bool
	Wood, Root, Gold                                                               tcell.Color
	Limb                                                                           []tcell.Color
}

func rgb(h uint32) tcell.Color { return tcell.NewHexColor(int32(h)) }

func newPalette(d dial) palette {
	bg := theme.Color(theme.RoleBackground)
	p := palette{
		Bg: bg, Text: theme.Color(theme.RoleText), Dim: theme.Color(theme.RoleDim),
		Accent: theme.Color(theme.RoleAccent), Highlight: theme.Color(theme.RoleHighlight),
		Positive: theme.Color(theme.RolePositive), Negative: theme.Color(theme.RoleNegative),
		Warning: theme.Color(theme.RoleWarning), Surface: theme.Color(theme.RoleSurface),
		Border: theme.Color(theme.RoleBorder), Light: theme.IsLight(),
	}
	p.Wood = theme.Legible(rgb(d.Wood), bg, 2.4)
	p.Root = theme.Legible(theme.Mix(rgb(d.Wood), bg, 0.3), bg, 1.8)
	p.Gold = theme.Legible(rgb(0xffc83d), bg, 3.0)
	for _, s := range limbSpecs {
		p.Limb = append(p.Limb, theme.Legible(rgb(s.Hue), bg, 3.0))
	}
	return p
}

// soil is the background of an epoch's stratum: the epoch hue washed into
// the theme background, alternating a touch so neighbours separate.
func (p palette) soil(epoch string, i int) tcell.Color {
	amt := 0.2 + 0.06*float64(i%2)
	if p.Light {
		amt += 0.07
	}
	return theme.Mix(p.Bg, rgb(dialFor(epoch).Hue), amt)
}
