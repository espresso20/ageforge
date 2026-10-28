package ui

import "fmt"

// Wonder icons are two-cell half-block glyphs for the Wonders overlay lists.
// Each wonder's icon is four colours; the first column is the left cell and
// the second the right cell, in the order below.
type wonderIconColors struct {
	upperLeft, lowerLeft, upperRight, lowerRight uint32
}

// wonderIconFallback is the pyramid, for a wonder key with no entry.
var wonderIconFallback = wonderIconColors{0xe8c860, 0xa08030, 0x000000, 0xa08030}

var wonderIcons = map[string]wonderIconColors{
	"sacred_grove":         {0x3a7a20, 0x1a2a10, 0x1a2a10, 0x1a2a10},
	"great_monolith":       {0xa09080, 0xa09080, 0x303030, 0x505050},
	"stonehenge":           {0xc0b0a0, 0x202018, 0x202018, 0x202018},
	"colosseum":            {0x201808, 0x201808, 0xc8a868, 0x201808},
	"parthenon":            {0xd0c898, 0xb8b080, 0xd0c898, 0xb8b080},
	"great_library":        {0x808090, 0xa0a0b0, 0x808090, 0xa0a0b0},
	"sistine_chapel":       {0xd8d0a8, 0xc0b890, 0x181818, 0xc0b890},
	"grand_lighthouse":     {0x001830, 0xa09080, 0x001830, 0xa09080},
	"crystal_palace":       {0xb8e8f8, 0x101820, 0x607080, 0x607080},
	"eiffel_tower":         {0x101018, 0x101018, 0x101018, 0x808070},
	"hoover_dam":           {0xa0a0a0, 0x1040a0, 0xa0a0a0, 0x1040a0},
	"particle_accelerator": {0x4060a0, 0x080810, 0x080810, 0x080810},
	"space_program":        {0x050510, 0x050510, 0x050510, 0x050510},
	"global_network":       {0x404860, 0x404860, 0x404860, 0x404860},
	"world_simulation":     {0x080810, 0x00ffff, 0x080810, 0x00ffff},
	"neon_citadel":         {0x040814, 0x040814, 0x040814, 0x040814},
	"stellar_cradle":       {0x2060c0, 0x020208, 0x020208, 0x020208},
	"dyson_scaffold":       {0x020208, 0x020208, 0x607080, 0x020208},
	"warp_nexus":           {0xc87020, 0x020208, 0x020208, 0x020208},
	"cosmic_beacon":        {0x400080, 0x0a0018, 0xffaa10, 0x0a0018},
	"reality_anchor":       {0xcc8810, 0x0a0018, 0x0a0018, 0x0a0018},
	"singularity_core":     {0x040c04, 0x040c04, 0x040c04, 0x040c04},
}

// WonderSpriteIcon renders a wonder's icon as two half-block cells (tview color
// tags). Each cell is a '▄' whose foreground is the lower colour and whose
// background is the upper colour.
func WonderSpriteIcon(key string) string {
	colors, ok := wonderIcons[key]
	if !ok {
		colors = wonderIconFallback
	}
	return halfBlock(colors.upperLeft, colors.lowerLeft) + halfBlock(colors.upperRight, colors.lowerRight)
}

func halfBlock(upper, lower uint32) string {
	return fmt.Sprintf("[#%06x:#%06x]▄[-:-]", lower, upper)
}
