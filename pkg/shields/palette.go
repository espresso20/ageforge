package shields

import "image/color"

// A pal is one tier's colours. Metal is the rim, crest and wings; field is the
// shield's face, light to dark; em is the emblem's metal.
type pal struct {
	outline, deep       color.RGBA
	metal               [3]color.RGBA // light, mid, shade
	field               [5]color.RGBA // light to dark
	em                  [3]color.RGBA // light, main, shade
	spark, sparkHi      color.RGBA    // the emblem's warm touch, and the crown gem
	rib, ribLo, ribHi   color.RGBA
	ray, glint, twinkle color.RGBA
}

func rgb(h uint32) color.RGBA { return color.RGBA{uint8(h >> 16), uint8(h >> 8), uint8(h), 255} }

var pals = [numTiers]pal{
	Bronze: {
		outline: rgb(0x2a150c), deep: rgb(0x3a1d10),
		metal: [3]color.RGBA{rgb(0xe3a468), rgb(0xb8743c), rgb(0x7a4626)},
		field: [5]color.RGBA{rgb(0x8c3a30), rgb(0x7c3229), rgb(0x6a2a25), rgb(0x58221f), rgb(0x461a1a)},
		em:    [3]color.RGBA{rgb(0xffd9a0), rgb(0xeab070), rgb(0xb8743c)},
		spark: rgb(0xffb347), sparkHi: rgb(0xffe0a0),
		glint: rgb(0xfff0d8), twinkle: rgb(0xffe9c8),
	},
	Silver: {
		outline: rgb(0x141a26), deep: rgb(0x1d2533),
		metal: [3]color.RGBA{rgb(0xf0f4fa), rgb(0xb6c2d2), rgb(0x76859b)},
		field: [5]color.RGBA{rgb(0x56657a), rgb(0x475569), rgb(0x3b4859), rgb(0x2f3a49), rgb(0x252e3b)},
		em:    [3]color.RGBA{rgb(0xffffff), rgb(0xdbe4ee), rgb(0x9aa8bb)},
		spark: rgb(0x9ad8ff), sparkHi: rgb(0xe0f4ff),
		rib: rgb(0x9a2f3a), ribLo: rgb(0x6c2029), ribHi: rgb(0xc4505a),
		glint: rgb(0xffffff), twinkle: rgb(0xffffff),
	},
	Gold: {
		outline: rgb(0x3a2506), deep: rgb(0x1a2150),
		metal: [3]color.RGBA{rgb(0xffe9a0), rgb(0xf2b52a), rgb(0xb27414)},
		field: [5]color.RGBA{rgb(0x2c4a9c), rgb(0x243e86), rgb(0x1d3470), rgb(0x172a5c), rgb(0x12214a)},
		em:    [3]color.RGBA{rgb(0xfff1b0), rgb(0xffd04a), rgb(0xd4901c)},
		spark: rgb(0xff7a3a), sparkHi: rgb(0xffc08a),
		rib: rgb(0xb02a3a), ribLo: rgb(0x7a1c28), ribHi: rgb(0xd8505e),
		glint: rgb(0xfffbe0), twinkle: rgb(0xfff6c8),
	},
	Platinum: {
		outline: rgb(0x15222e), deep: rgb(0x0e2c36),
		metal: [3]color.RGBA{rgb(0xf4fcff), rgb(0xc3dae8), rgb(0x87a6bd)},
		field: [5]color.RGBA{rgb(0x2a7f8e), rgb(0x226d7b), rgb(0x1b5b68), rgb(0x154955), rgb(0x103a45)},
		em:    [3]color.RGBA{rgb(0xffffff), rgb(0xe8f6fb), rgb(0x9cc2d6)},
		spark: rgb(0x5ee6ff), sparkHi: rgb(0xd8fbff),
		rib: rgb(0x2b6fb0), ribLo: rgb(0x1d4e80), ribHi: rgb(0x4a92d8),
		glint: rgb(0xffffff), twinkle: rgb(0xe8fbff),
	},
	Legendary: {
		outline: rgb(0x2a0a3a), deep: rgb(0x240c3e),
		metal: [3]color.RGBA{rgb(0xfff4c0), rgb(0xffb52e), rgb(0xc4602a)},
		field: [5]color.RGBA{rgb(0x7a32b8), rgb(0x66299c), rgb(0x531f82), rgb(0x41186a), rgb(0x301152)},
		em:    [3]color.RGBA{rgb(0xfffbe0), rgb(0xffe27a), rgb(0xe98a2c)},
		spark: rgb(0xff5ac8), sparkHi: rgb(0xffc0ec),
		rib: rgb(0xc2287a), ribLo: rgb(0x8a1a58), ribHi: rgb(0xe85aa0),
		ray: rgb(0xffd36a), glint: rgb(0xffffff), twinkle: rgb(0xfff4c0),
	},
}

// metalTone is a mask tone as the tier's metal.
func (p *pal) metalTone(t uint8) color.RGBA {
	switch t {
	case tLight:
		return p.metal[0]
	case tShade:
		return p.metal[2]
	case tDeep:
		return p.deep
	case tGem:
		return p.spark
	case tGemHi:
		return p.sparkHi
	case tRib:
		return p.rib
	case tRibLo:
		return p.ribLo
	case tRibHi:
		return p.ribHi
	}
	return p.metal[1]
}

// emTone is a mask tone as the tier's emblem metal.
func (p *pal) emTone(t uint8) color.RGBA {
	switch t {
	case tLight:
		return p.em[0]
	case tShade:
		return p.em[2]
	case tDeep:
		return p.deep
	case tGem:
		return p.spark
	case tGemHi:
		return p.sparkHi
	}
	return p.em[1]
}

// The gems set in a legendary rim, one after another.
var rimGems = [][2]color.RGBA{
	{rgb(0xe0304a), rgb(0xffa0ac)}, // ruby
	{rgb(0x3a7aff), rgb(0xa8c8ff)}, // sapphire
	{rgb(0x2ed67a), rgb(0xa8ffd0)}, // emerald
	{rgb(0xb05aff), rgb(0xe2b8ff)}, // amethyst
	{rgb(0xffd040), rgb(0xfff0a0)}, // topaz
}
