package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/pkg/sprites"
)

// Wonder icons — small two-cell half-block glyphs rendered from in-code 16×16
// wonder sprites. Extracted from the retired MapV4 (ui/map.go) so the Wonders
// overlays (overlay_wonders.go, wonder_gallery.go) keep their per-wonder icons
// after the map rewrite. The new procedural map (ui/citymap) does not use these.

// Wonder sprite color constants (sandstone / gold palette).
const (
	v4wnS = 0xc8a840 // sandstone
	v4wnL = 0xe8c860 // light face
	v4wnD = 0xa08030 // shadow
	v4wnG = 0xd4a017 // gold cap
)

// v4SpriteWonder is the generic pyramid fallback used when a wonder key has no
// bespoke sprite in v4WonderSpriteByKey.
func v4SpriteWonder() [16][16]uint32 {
	return [16][16]uint32{
		{0, 0, 0, 0, 0, 0, 0, v4wnG, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wnG, v4wnG, v4wnG, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, v4wnL, v4wnS, v4wnD, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, 0, 0, 0, 0, 0},
		{0, 0, 0, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, 0, 0, 0, 0},
		{0, 0, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0, 0, 0},
		{0, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0, 0},
		{v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnL, v4wnS, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, v4wnS, 0},
		{v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, v4wnD, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

var v4SpriteWonderPixels = v4SpriteWonder()

// v4WonderSpriteByKey returns a unique 16×16 sprite for each of the 22 age wonders.
// Falls back to the generic pyramid if the key is unrecognised.
func v4WonderSpriteByKey(key string) [16][16]uint32 {
	switch key {
	case "sacred_grove":
		// Primitive age — circle of standing stones with a tree canopy
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x1A2A10)
		c.Dome(7, 6, 5, 0x3A7A20)
		c.Dome(7, 5, 3, 0x4A9A30)
		c.Set(7, 2, 0x60C040)
		c.Vline(7, 10, 3, 0x6B4020)
		c.Vline(8, 10, 3, 0x6B4020)
		stones := [][2]int{{2, 12}, {5, 14}, {10, 14}, {13, 12}, {14, 9}, {13, 6}, {2, 9}}
		for _, s := range stones {
			c.FillRect(s[0], s[1], 2, 2, 0xA09080)
		}
		return c.Pixels

	case "great_monolith":
		// Stone age — megalith row of standing stones
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x303030)
		c.Hline(0, 12, 16, 0x505050)
		posts := []int{1, 4, 7, 10, 13}
		for _, x := range posts {
			c.FillRect(x, 3, 2, 10, 0xA09080)
			c.Hline(x, 3, 2, 0xC0B0A0)
		}
		c.FillRect(4, 2, 5, 2, 0xC0B0A0)
		return c.Pixels

	case "stonehenge":
		// Bronze age — Stonehenge-style trilithon ring
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x202018)
		ring := [][2]int{{7, 1}, {11, 2}, {14, 5}, {14, 10}, {11, 13}, {7, 14}, {3, 13}, {2, 10}, {2, 5}, {4, 2}}
		for _, s := range ring {
			c.FillRect(s[0]-1, s[1]-1, 2, 3, 0x8C7C6C)
		}
		c.FillRect(5, 5, 2, 6, 0xA09080)
		c.FillRect(9, 5, 2, 6, 0xA09080)
		c.FillRect(4, 4, 8, 2, 0xC0B0A0)
		c.FillRect(6, 10, 4, 2, 0xD0C0B0)
		return c.Pixels

	case "colosseum":
		// Classical age — oval ring of arched walls
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x201808)
		c.Dome(7, 8, 7, 0xC8A868)
		c.Dome(7, 8, 5, 0x201808)
		c.FillRect(1, 10, 14, 5, 0xC8A868)
		c.FillRect(3, 10, 10, 3, 0x201808)
		for x := 1; x < 14; x += 2 {
			c.Set(x, 10, 0x201808)
		}
		return c.Pixels

	case "parthenon":
		// Iron age — Greek temple with columns and pediment
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x181810)
		c.FillRect(1, 13, 14, 2, 0xD0C898)
		c.FillRect(0, 12, 16, 1, 0xB8B080)
		for x := 1; x < 15; x += 2 {
			c.Vline(x, 6, 7, 0xE8E0C0)
		}
		c.FillRect(1, 5, 14, 2, 0xC8C090)
		c.FillRect(3, 3, 10, 2, 0xD0C898)
		c.FillRect(5, 2, 6, 1, 0xD0C898)
		c.Set(7, 1, 0xE8E0C0)
		c.Set(8, 1, 0xE8E0C0)
		return c.Pixels

	case "great_library":
		// Medieval age — library with arched windows
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x181818)
		c.FillRect(1, 13, 14, 2, 0x909090)
		c.FillRect(2, 5, 12, 9, 0xA0A0B0)
		c.FillRect(1, 3, 14, 3, 0x808090)
		c.FillRect(3, 1, 10, 3, 0x909090)
		c.Set(7, 0, 0xC0C0C0)
		c.Set(8, 0, 0xC0C0C0)
		for _, x := range []int{3, 7, 11} {
			c.FillRect(x, 7, 2, 4, 0x404060)
			c.Set(x, 6, 0x404060)
			c.Set(x+1, 6, 0x404060)
		}
		c.Hline(3, 9, 10, 0x808080)
		c.Hline(3, 11, 10, 0x808080)
		return c.Pixels

	case "sistine_chapel":
		// Renaissance age — chapel dome with cross
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x181818)
		c.FillRect(1, 9, 14, 6, 0xC0B890)
		c.FillRect(4, 6, 8, 4, 0xA8A080)
		c.Dome(7, 6, 4, 0xD8D0A8)
		c.FillRect(6, 1, 4, 3, 0xB0A880)
		c.Set(7, 0, 0xE8E0C0)
		c.Set(8, 0, 0xE8E0C0)
		c.Set(7, 4, 0x303028)
		c.Set(8, 4, 0x303028)
		return c.Pixels

	case "grand_lighthouse":
		// Colonial age — tall lighthouse with light beam
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x001830)
		c.FillRect(0, 13, 16, 3, 0x1040A0)
		c.Hline(0, 12, 16, 0x2060C0)
		c.FillRect(5, 3, 6, 10, 0xD0C8B0)
		c.FillRect(6, 1, 4, 3, 0xC0B8A0)
		c.FillRect(5, 0, 6, 2, 0xFFFF80)
		c.Set(4, 1, 0xFFFF40)
		c.Set(11, 1, 0xFFFF40)
		c.Set(1, 2, 0xFFFF80)
		c.Set(2, 1, 0xFFFF80)
		c.FillRect(3, 12, 10, 2, 0xA09080)
		return c.Pixels

	case "crystal_palace":
		// Industrial age — glass grid building
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101820)
		c.FillRect(0, 2, 16, 4, 0x80C8E8)
		c.Dome(7, 4, 5, 0xA0E0F8)
		for x := 0; x < 16; x += 3 {
			c.Vline(x, 2, 12, 0x607080)
		}
		for y := 2; y < 14; y += 3 {
			c.Hline(0, y, 16, 0x607080)
		}
		c.FillRect(1, 3, 2, 2, 0xB8E8F8)
		c.FillRect(4, 3, 2, 2, 0xB8E8F8)
		c.FillRect(7, 3, 2, 2, 0xB8E8F8)
		c.FillRect(10, 3, 2, 2, 0xB8E8F8)
		c.FillRect(13, 3, 2, 2, 0xB8E8F8)
		c.FillRect(0, 13, 16, 3, 0x506070)
		return c.Pixels

	case "eiffel_tower":
		// Victorian age — Eiffel Tower silhouette
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101018)
		c.FillRect(1, 10, 3, 5, 0x808070)
		c.FillRect(12, 10, 3, 5, 0x808070)
		c.FillRect(4, 8, 3, 4, 0x909080)
		c.FillRect(9, 8, 3, 4, 0x909080)
		c.Set(4, 10, 0x101018)
		c.Set(11, 10, 0x101018)
		c.FillRect(5, 5, 6, 4, 0xA0A090)
		c.FillRect(7, 1, 2, 5, 0xC0C0B0)
		c.Set(7, 0, 0xD8D8C8)
		c.Set(8, 0, 0xD8D8C8)
		c.Hline(0, 14, 16, 0x606060)
		return c.Pixels

	case "hoover_dam":
		// Electric age — dam wall with water
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x101820)
		c.FillRect(0, 0, 16, 6, 0x203050)
		c.FillRect(2, 3, 12, 10, 0x909090)
		c.FillRect(3, 4, 10, 9, 0xA0A0A0)
		c.FillRect(0, 0, 16, 4, 0x1050A0)
		c.Hline(0, 3, 16, 0x2080D0)
		c.FillRect(0, 12, 16, 4, 0x1040A0)
		c.Vline(4, 5, 7, 0x606060)
		c.Vline(7, 5, 7, 0x606060)
		c.Vline(10, 5, 7, 0x606060)
		c.Hline(3, 11, 10, 0xFFCC20)
		return c.Pixels

	case "particle_accelerator":
		// Atomic age — circular ring accelerator
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x080810)
		c.Dome(7, 8, 6, 0x4060A0)
		c.Dome(7, 8, 4, 0x080810)
		c.Dome(7, 8, 3, 0x2040C0)
		c.Dome(7, 8, 1, 0x080810)
		c.Set(7, 8, 0x80C0FF)
		c.Set(8, 8, 0x80C0FF)
		c.Hline(0, 8, 2, 0x6080FF)
		c.Hline(14, 8, 2, 0x6080FF)
		c.Vline(7, 0, 2, 0x6080FF)
		c.Vline(7, 14, 2, 0x6080FF)
		return c.Pixels

	case "space_program":
		// Space age — rocket on launch pad
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x050510)
		c.Set(2, 1, 0xFFFFFF)
		c.Set(6, 3, 0xFFFFFF)
		c.Set(11, 0, 0xFFFFFF)
		c.Set(14, 4, 0xFFFFFF)
		c.Vline(2, 4, 10, 0x707070)
		c.Hline(2, 6, 3, 0x707070)
		c.Hline(2, 9, 3, 0x707070)
		c.FillRect(6, 3, 4, 9, 0xD0D0D0)
		c.FillRect(7, 1, 2, 3, 0xE0E0E0)
		c.Set(7, 0, 0xC0C0C0)
		c.Set(8, 0, 0xC0C0C0)
		c.FillRect(5, 10, 2, 3, 0xA0A0B0)
		c.FillRect(9, 10, 2, 3, 0xA0A0B0)
		c.Hline(6, 12, 4, 0xFF8020)
		c.Hline(7, 13, 2, 0xFFAA40)
		c.FillRect(4, 13, 8, 2, 0x606070)
		return c.Pixels

	case "global_network":
		// Information age — data center server rack pattern
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0A18)
		c.FillRect(1, 1, 6, 14, 0x303048)
		c.FillRect(9, 1, 6, 14, 0x303048)
		for y := 2; y < 14; y += 2 {
			c.Hline(2, y, 4, 0x404860)
			c.Set(5, y, 0x00FF80)
			c.Hline(10, y, 4, 0x404860)
			c.Set(13, y, 0x0080FF)
		}
		c.Hline(7, 4, 2, 0x00FFFF)
		c.Hline(7, 8, 2, 0x00FFFF)
		c.Hline(7, 12, 2, 0x00FFFF)
		return c.Pixels

	case "world_simulation":
		// Cyberpunk age — neon sprawl cityscape
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x080810)
		c.FillRect(0, 8, 3, 8, 0x202028)
		c.FillRect(4, 6, 2, 10, 0x202028)
		c.FillRect(7, 4, 3, 12, 0x202028)
		c.FillRect(11, 7, 2, 9, 0x202028)
		c.FillRect(13, 5, 3, 11, 0x202028)
		c.Hline(0, 9, 16, 0xCC00FF)
		c.Hline(0, 12, 16, 0x00FFFF)
		c.Hline(0, 14, 16, 0xFF0080)
		c.Set(1, 10, 0xFF40FF)
		c.Set(5, 8, 0x40FFFF)
		c.Set(8, 6, 0xFF40FF)
		c.Set(12, 9, 0x40FFFF)
		c.Set(14, 7, 0xFF40FF)
		return c.Pixels

	case "neon_citadel":
		// Digital age — quantum hexagonal pattern
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x040814)
		hexPts := [][2]int{
			{7, 1}, {11, 3}, {13, 7}, {11, 11}, {7, 13}, {3, 11}, {1, 7}, {3, 3},
		}
		for _, p := range hexPts {
			c.Set(p[0], p[1], 0xA0C0FF)
			if p[0]+1 < 16 {
				c.Set(p[0]+1, p[1], 0xA0C0FF)
			}
		}
		inner := [][2]int{{7, 4}, {10, 6}, {10, 10}, {7, 12}, {4, 10}, {4, 6}}
		for _, p := range inner {
			c.Set(p[0], p[1], 0x60A0FF)
		}
		c.FillRect(6, 6, 4, 4, 0x2040A0)
		c.Set(7, 7, 0xE0F0FF)
		c.Set(8, 7, 0xE0F0FF)
		c.Set(7, 8, 0xE0F0FF)
		c.Set(8, 8, 0xE0F0FF)
		return c.Pixels

	case "stellar_cradle":
		// Fusion age — torus ring with plasma core
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		c.Dome(7, 8, 6, 0x2060C0)
		c.Dome(7, 8, 4, 0x020208)
		c.Dome(7, 8, 3, 0xFF6020)
		c.Dome(7, 8, 1, 0xFFB040)
		c.Set(7, 8, 0xFFFF80)
		c.Set(8, 8, 0xFFFF80)
		c.Hline(2, 8, 3, 0x4080D0)
		c.Hline(11, 8, 3, 0x4080D0)
		return c.Pixels

	case "dyson_scaffold":
		// Interstellar age — elongated silver ark ship
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		c.Set(1, 2, 0xFFFFFF)
		c.Set(5, 0, 0xFFFFFF)
		c.Set(12, 3, 0xFFFFFF)
		c.Set(14, 1, 0xFFFFFF)
		c.FillRect(1, 6, 14, 4, 0xA0B0C0)
		c.FillRect(2, 5, 12, 6, 0xB0C0D0)
		c.Set(0, 7, 0x808090)
		c.Set(0, 8, 0x808090)
		c.FillRect(12, 4, 3, 2, 0x607080)
		c.FillRect(12, 10, 3, 2, 0x607080)
		c.Set(15, 5, 0x40C0FF)
		c.Set(15, 6, 0x40C0FF)
		c.Set(15, 9, 0x40C0FF)
		c.Set(15, 10, 0x40C0FF)
		c.Hline(4, 7, 7, 0x6080FF)
		return c.Pixels

	case "warp_nexus":
		// Galactic age — dyson sphere orange/gold ring
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x020208)
		c.Set(1, 1, 0xFFFFFF)
		c.Set(13, 2, 0xFFFFFF)
		c.Set(3, 13, 0xFFFFFF)
		c.Set(14, 12, 0xFFFFFF)
		c.Dome(7, 8, 6, 0xC87020)
		c.Dome(7, 8, 4, 0x020208)
		c.Dome(7, 8, 3, 0xFFAA20)
		c.FillRect(5, 6, 6, 6, 0xFFCC40)
		c.Set(7, 7, 0xFFFF80)
		c.Set(8, 7, 0xFFFF80)
		c.Set(7, 8, 0xFFFF80)
		c.Set(8, 8, 0xFFFF80)
		c.Set(7, 2, 0xE09030)
		c.Set(7, 14, 0xE09030)
		c.Set(1, 8, 0xE09030)
		c.Set(14, 8, 0xE09030)
		return c.Pixels

	case "cosmic_beacon":
		// Quantum age — reality anchor deep purple with gold geometric star
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0018)
		c.Vline(7, 0, 3, 0xFFCC20)
		c.Vline(7, 13, 3, 0xFFCC20)
		c.Hline(0, 7, 3, 0xFFCC20)
		c.Hline(13, 7, 3, 0xFFCC20)
		c.Set(3, 3, 0xFFAA10)
		c.Set(4, 4, 0xFFAA10)
		c.Set(11, 3, 0xFFAA10)
		c.Set(12, 4, 0xFFAA10)
		c.Set(3, 12, 0xFFAA10)
		c.Set(4, 11, 0xFFAA10)
		c.Set(11, 12, 0xFFAA10)
		c.Set(12, 11, 0xFFAA10)
		c.FillRect(4, 4, 8, 8, 0x400080)
		c.FillRect(5, 5, 6, 6, 0x600090)
		c.Set(7, 7, 0xFFCC20)
		c.Set(8, 7, 0xFFCC20)
		c.Set(7, 8, 0xFFCC20)
		c.Set(8, 8, 0xFFCC20)
		return c.Pixels

	case "reality_anchor":
		// Quantum age — reality anchor with radiating chain lines
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x0A0018)
		c.Vline(7, 0, 16, 0x5000A0)
		c.Hline(0, 7, 16, 0x5000A0)
		c.Vline(8, 0, 16, 0x5000A0)
		c.Hline(0, 8, 16, 0x5000A0)
		c.Set(7, 1, 0xFFCC20)
		c.Set(7, 14, 0xFFCC20)
		c.Set(1, 7, 0xFFCC20)
		c.Set(14, 7, 0xFFCC20)
		c.Set(4, 4, 0xCC8810)
		c.Set(11, 4, 0xCC8810)
		c.Set(4, 11, 0xCC8810)
		c.Set(11, 11, 0xCC8810)
		c.FillRect(5, 5, 6, 6, 0x300060)
		c.FillRect(6, 6, 4, 4, 0x500090)
		c.Set(7, 7, 0xE0A020)
		c.Set(8, 7, 0xE0A020)
		c.Set(7, 8, 0xE0A020)
		c.Set(8, 8, 0xE0A020)
		return c.Pixels

	case "singularity_core":
		// Transcendent age — world tree with luminous tip
		var c sprites.Canvas
		c.FillRect(0, 0, 16, 16, 0x040C04)
		c.FillRect(6, 13, 4, 3, 0x5A3010)
		c.Set(4, 14, 0x3A2008)
		c.Set(5, 15, 0x3A2008)
		c.Set(10, 14, 0x3A2008)
		c.Set(11, 15, 0x3A2008)
		c.FillRect(7, 7, 2, 7, 0x6B4020)
		c.Set(5, 8, 0x5A3810)
		c.Set(6, 7, 0x5A3810)
		c.Set(9, 7, 0x5A3810)
		c.Set(10, 8, 0x5A3810)
		c.Dome(7, 9, 4, 0x206820)
		c.Dome(7, 7, 3, 0x30A030)
		c.Dome(7, 5, 3, 0x50C040)
		c.Dome(7, 3, 2, 0x70E060)
		c.Set(7, 1, 0xC0FF80)
		c.Set(8, 1, 0xC0FF80)
		c.Set(7, 0, 0xFFFFCC)
		return c.Pixels

	default:
		return v4SpriteWonderPixels
	}
}

// WonderSpriteIcon renders a wonder's 16×16 sprite as a two-cell half-block icon
// (tview color tags), sampling the upper/lower pixel pairs at two columns.
// Each character is a '▄' half-block: BG = upper sample pixel, FG = lower sample pixel.
func WonderSpriteIcon(key string) string {
	px := v4WonderSpriteByKey(key)
	c := func(row, col int) (r, g, b uint8) {
		packed := px[row][col]
		return uint8(packed >> 16), uint8(packed >> 8), uint8(packed)
	}
	r1u, g1u, b1u := c(4, 4)
	r1l, g1l, b1l := c(12, 4)
	r2u, g2u, b2u := c(4, 12)
	r2l, g2l, b2l := c(12, 12)
	ch1 := fmt.Sprintf("[#%02x%02x%02x:#%02x%02x%02x]▄[-:-]", r1l, g1l, b1l, r1u, g1u, b1u)
	ch2 := fmt.Sprintf("[#%02x%02x%02x:#%02x%02x%02x]▄[-:-]", r2l, g2l, b2l, r2u, g2u, b2u)
	return ch1 + ch2
}
