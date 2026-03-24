// pkg/sprites/sprites.go — shared parametric 16×16 building sprite generator
package sprites

import "math"

// ── Era color palettes ────────────────────────────────────────────────────────

// EraPalette holds the six color slots for a single era.
type EraPalette struct {
	Wall, Roof, Window, Door, Accent, Glow uint32
}

// EraPalettes is the ordered set of 7 era palettes (index 0–6).
var EraPalettes = [7]EraPalette{
	// Era 0: Primitive/Ancient (primitive→iron age)
	{0x8b6a40, 0x5a3a20, 0xc8a060, 0x5a3a20, 0xc89040, 0xffe080},
	// Era 1: Classical/Medieval (classical→renaissance)
	{0xd0c8b0, 0x8b1a1a, 0xa0c0d0, 0x6b3a1a, 0xd4a017, 0xffe0a0},
	// Era 2: Colonial/Industrial (age_of_sail→gilded)
	{0x8a6050, 0x4a3a30, 0xc0c0a0, 0x3a2a20, 0xc87941, 0xffa060},
	// Era 3: Modern/Atomic (modern→information)
	{0x808890, 0x505860, 0x80c0e0, 0x404850, 0xc0c8d0, 0xe0f0ff},
	// Era 4: Digital/Cyber (digital→cyberpunk)
	{0x202030, 0x101020, 0x40c0ff, 0x181828, 0xff40c0, 0x40ffff},
	// Era 5: Fusion/Space (nanotech→interstellar)
	{0x0a1530, 0x050a20, 0x4080ff, 0x030810, 0xffe040, 0x80c0ff},
	// Era 6: Cosmic/Transcendent (galactic→divine)
	{0x080810, 0x040408, 0xc0a0ff, 0x020206, 0xffd040, 0xff80ff},
}

// ── Age → era mapping ─────────────────────────────────────────────────────────

// AgeEraMap maps age key strings to era indices (0–6).
var AgeEraMap = map[string]int{
	"primitive_age": 0, "stone_age": 0, "bronze_age": 0, "iron_age": 0,
	"classical_age": 1, "medieval_age": 1, "renaissance_age": 1,
	"age_of_sail": 2, "colonial_age": 2, "industrial_age": 2, "victorian_age": 2, "gilded_age": 2,
	"electric_age": 3, "modern_age": 3, "atomic_age": 3, "information_age": 3,
	"digital_age": 4, "cyberpunk_age": 4,
	"nanotech_age": 5, "fusion_age": 5, "space_age": 5, "interstellar_age": 5,
	"galactic_age": 6, "quantum_age": 6, "singularity_age": 6, "transcendent_age": 6, "divine_age": 6, "cosmic_age": 6,
}

// GetEra returns the era index for a given age key (defaults to 0 if unknown).
func GetEra(ageKey string) int {
	if e, ok := AgeEraMap[ageKey]; ok {
		return e
	}
	return 0
}

// ── Canvas type and drawing helpers ──────────────────────────────────────────

// Canvas is a 16×16 pixel art canvas with uint32 packed RGB pixels (0xRRGGBB; 0 = transparent).
type Canvas struct {
	Pixels [16][16]uint32
}

// fillRect fills a rectangle. x,y = top-left, w,h = size. Clips to 16×16.
func (c *Canvas) fillRect(x, y, w, h int, col uint32) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			rx, ry := x+dx, y+dy
			if rx >= 0 && rx < 16 && ry >= 0 && ry < 16 {
				c.Pixels[ry][rx] = col
			}
		}
	}
}

// set draws a single pixel.
func (c *Canvas) set(x, y int, col uint32) {
	if x >= 0 && x < 16 && y >= 0 && y < 16 {
		c.Pixels[y][x] = col
	}
}

// hline draws a horizontal line.
func (c *Canvas) hline(x, y, w int, col uint32) { c.fillRect(x, y, w, 1, col) }

// vline draws a vertical line.
func (c *Canvas) vline(x, y, h int, col uint32) { c.fillRect(x, y, 1, h, col) }

// pitchedRoof draws a pitched roof: peak at (cx, peakY), base from (x, baseY) to (x+w-1, baseY).
func (c *Canvas) pitchedRoof(cx, peakY, baseY, baseW int, col uint32) {
	for y := peakY; y <= baseY; y++ {
		spread := (y - peakY) * (baseW / 2) / (baseY - peakY + 1)
		c.hline(cx-spread, y, spread*2+1, col)
	}
}

// dome draws a dome: center cx,cy, radius r.
func (c *Canvas) dome(cx, cy, r int, col uint32) {
	for dy := -r; dy <= 0; dy++ {
		hw := int(math.Sqrt(float64(r*r - dy*dy)))
		c.hline(cx-hw, cy+dy, hw*2+1, col)
	}
}

// FillRect is the exported version of fillRect for use outside the package.
func (c *Canvas) FillRect(x, y, w, h int, col uint32) { c.fillRect(x, y, w, h, col) }

// Set is the exported version of set for use outside the package.
func (c *Canvas) Set(x, y int, col uint32) { c.set(x, y, col) }

// Hline is the exported version of hline for use outside the package.
func (c *Canvas) Hline(x, y, w int, col uint32) { c.hline(x, y, w, col) }

// Vline is the exported version of vline for use outside the package.
func (c *Canvas) Vline(x, y, h int, col uint32) { c.vline(x, y, h, col) }

// Dome is the exported version of dome for use outside the package.
func (c *Canvas) Dome(cx, cy, r int, col uint32) { c.dome(cx, cy, r, col) }

// darken darkens a color by factor (0.0–1.0).
func darken(col uint32, factor float64) uint32 {
	r := uint8(float64(col>>16&0xff) * factor)
	g := uint8(float64(col>>8&0xff) * factor)
	b := uint8(float64(col&0xff) * factor)
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

// lighten lightens a color by adding to each channel.
func lighten(col uint32, add uint8) uint32 {
	r := min16(int(col>>16&0xff)+int(add), 255)
	g := min16(int(col>>8&0xff)+int(add), 255)
	b := min16(int(col&0xff)+int(add), 255)
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

func min16(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Core building composer ────────────────────────────────────────────────────

// buildingCompose draws the shared structural frame. lineage and era control shapes.
// variation (0-2) adds small stable differences.
func buildingCompose(lin, tier, era, variation int) Canvas {
	var c Canvas
	pal := EraPalettes[era]

	bodyW := 4 + tier*6/20
	if bodyW%2 != 0 {
		bodyW++
	} // keep even
	if bodyW > 10 {
		bodyW = 10
	}
	bodyH := 3 + tier*5/20
	if bodyH > 8 {
		bodyH = 8
	}

	bx := 8 - bodyW/2 // body left x
	by := 14 - bodyH  // body top y (foundation is 14-15)

	// Foundation
	c.fillRect(bx-1, 14, bodyW+2, 2, darken(pal.Wall, 0.6))

	// Body walls
	c.fillRect(bx, by, bodyW, bodyH, pal.Wall)

	// Body outline (slightly darker)
	c.hline(bx, by, bodyW, darken(pal.Wall, 0.7))         // top edge
	c.vline(bx, by, bodyH, darken(pal.Wall, 0.7))         // left edge
	c.vline(bx+bodyW-1, by, bodyH, darken(pal.Wall, 0.7)) // right edge

	// Windows: 1×2 px each, spaced across body
	numWindows := 1 + tier*2/10 + variation%2
	if numWindows > 3 {
		numWindows = 3
	}
	for w := 0; w < numWindows; w++ {
		wx := bx + 1 + w*(bodyW-2)/(numWindows)
		wy := by + 1
		c.fillRect(wx, wy, 1, 2, pal.Window)
	}

	// Door: 2 wide, at bottom of body, center
	dx := bx + bodyW/2 - 1
	c.fillRect(dx, by+bodyH-2, 2, 2, pal.Door)

	// Roof — shape varies by lineage group
	roofTop := by - 1
	switch {
	case lin <= 2: // housing, food → pitched roof
		c.pitchedRoof(8, roofTop-2, roofTop, bodyW+2, pal.Roof)
	case lin <= 4: // organic/geological extraction → flat roof with chimney
		c.hline(bx-1, roofTop, bodyW+2, pal.Roof)
		cx2 := bx + bodyW - 2 + variation
		c.vline(cx2, roofTop-3, 3, darken(pal.Roof, 0.8)) // chimney
		c.hline(cx2-1, roofTop-4, 3, pal.Accent)           // chimney top
	case lin <= 6: // knowledge, faith → dome or spire
		if era >= 4 {
			c.vline(8, roofTop-4, 4, pal.Accent) // antenna/spire
			c.set(7, roofTop-5, pal.Glow)
			c.set(8, roofTop-5, pal.Glow)
		} else {
			c.dome(8, roofTop, 2+bodyW/4, pal.Roof)
		}
	case lin <= 9: // military, trade, engineering → flat industrial roofline
		c.hline(bx-1, roofTop, bodyW+2, pal.Roof)
		if era >= 2 { // smokestacks for industrial+
			for i := 0; i < 2; i++ {
				sx := bx + 1 + i*(bodyW-3)
				c.vline(sx, roofTop-4+i, 4-i, darken(pal.Roof, 0.7))
			}
		}
		if era >= 4 { // antenna array for digital+
			c.hline(bx+1, roofTop-2, bodyW-2, darken(pal.Wall, 0.5))
			for i := 0; i < 3; i++ {
				c.vline(bx+2+i*2, roofTop-4, 2, pal.Glow)
			}
		}
	case lin <= 11: // culture, metallurgy → ornate roof
		c.pitchedRoof(8, roofTop-3, roofTop, bodyW+2, pal.Roof)
		c.set(8, roofTop-3, pal.Accent) // finial
		c.set(7, roofTop-3, pal.Accent)
	default: // energy, hacker → futuristic flat with glow
		c.hline(bx-1, roofTop, bodyW+2, pal.Roof)
		if era >= 3 {
			c.hline(bx, roofTop-1, bodyW, pal.Glow) // glow strip on roof
		}
	}

	// Era-specific extra details overlaid
	switch era {
	case 0: // primitive: keep clean
		// nothing extra
	case 1: // classical: add banner/flag on top
		c.vline(bx+bodyW-1, by-3, 3, pal.Accent)
		c.hline(bx+bodyW-1, by-3, 2, pal.Accent)
	case 2: // industrial: smoke puff above chimneys
		c.set(bx+bodyW-2+variation, roofTop-5, darken(pal.Wall, 1.5))
	case 3: // modern: glass reflection on windows
		for w := 0; w < 3 && w < bodyW-2; w++ {
			wx := bx + 1 + w*(bodyW-2)/3
			c.set(wx, by+1, lighten(pal.Window, 60))
		}
	case 4: // cyber: neon strip on body edge
		c.vline(bx+bodyW-1, by, bodyH, pal.Glow)
	case 5: // fusion: energy beam from top
		c.vline(8, roofTop-5, 3, pal.Glow)
		c.set(7, roofTop-5, pal.Glow)
		c.set(9, roofTop-5, pal.Glow)
	case 6: // cosmic: glow halo dots around building
		c.set(bx-1, by-1, pal.Glow)
		c.set(bx+bodyW, by-1, pal.Glow)
		c.set(bx-1, by+bodyH, pal.Glow)
		c.set(bx+bodyW, by+bodyH, pal.Glow)
	}

	return c
}

// ── Public sprite generator ───────────────────────────────────────────────────

// GenerateBuildingSprite returns a 16×16 sprite for a building given its lineage (1-13),
// tier (0-20), age key, and variation (0, 1, or 2).
func GenerateBuildingSprite(lineage, tier int, ageKey string, variation int) [16][16]uint32 {
	era := GetEra(ageKey)
	pal := EraPalettes[era]
	_ = pal
	c := buildingCompose(lineage, tier, era, variation)

	// Lineage-specific overrides that make shapes more recognizable
	switch lineage {
	case 1: // Housing — normal compose is good, vary roof color slightly per variation
		// already handled in compose

	case 2: // Food — add crop rows at base
		bw := 4 + tier*6/20
		if bw > 10 {
			bw = 10
		}
		bx := 8 - bw/2
		if era <= 1 { // primitive/classical: show crop field base
			for i := 0; i < 3; i++ {
				c.set(bx+i*2, 13, 0x4a8020)
				c.set(bx+i*2+1, 13, 0x2a6010)
			}
		}

	case 3: // Organic extraction (lumber/oil) — show resource indicator
		if era <= 1 { // show tree stump
			c.set(5, 12, 0x5a3a20)
			c.set(6, 12, 0x5a3a20)
			c.set(5, 11, 0x2d6b1a)
			c.set(6, 11, 0x2d6b1a)
			c.set(4, 10, 0x2d6b1a)
			c.set(7, 10, 0x2d6b1a)
		} else if era >= 3 { // show pipe/conduit
			c.vline(4, 10, 5, darken(pal.Roof, 0.8))
			c.hline(4, 10, 3, darken(pal.Roof, 0.8))
		}

	case 4: // Geological (quarry/mine) — show pit marker
		if era <= 2 {
			c.hline(3, 14, 3, 0x4a4a4a)
			c.hline(3, 15, 3, 0x333333)
			c.set(3, 13, 0x666666)
		}

	case 5: // Knowledge — show book/scroll detail
		if era <= 1 {
			bw := 4 + tier*6/20
			if bw > 10 {
				bw = 10
			}
			bx := 8 - bw/2
			c.fillRect(bx+1, 7, 2, 3, 0xf0f0e0) // book on wall
			c.vline(bx+2, 7, 3, 0xd0c8b0)
		}

	case 6: // Faith — add cross or star symbol
		if era <= 2 {
			bx := 8 - (4+tier*6/20)/2
			bw := 4 + tier*6/20
			if bw > 10 {
				bw = 10
			}
			cx2 := bx + bw/2
			by2 := 14 - (3+tier*5/20) - 4
			c.set(cx2, by2, pal.Accent)
			c.set(cx2-1, by2+1, pal.Accent)
			c.set(cx2, by2+1, pal.Accent)
			c.set(cx2+1, by2+1, pal.Accent)
			c.set(cx2, by2+2, pal.Accent)
		} else if era >= 4 { // cyber/future faith: star/glow
			c.set(8, 2, pal.Glow)
			c.set(7, 3, pal.Glow)
			c.set(9, 3, pal.Glow)
			c.set(8, 4, pal.Glow)
		}

	case 7: // Military — add battlements on top
		bw := 4 + tier*6/20
		if bw > 10 {
			bw = 10
		}
		bx := 8 - bw/2
		by2 := 14 - (3 + tier*5/20)
		if era <= 2 {
			for i := 0; i < bw; i += 2 {
				c.set(bx+i, by2-1, pal.Wall)
			}
		}

	case 8: // Trade — add coin/scale symbol
		if era <= 3 {
			c.set(8, 4, 0xd4a017) // gold coin dot
			c.set(7, 4, 0xd4a017)
		}

	case 9: // Engineering — add gear on roof
		if era >= 2 {
			c.set(8, 3, pal.Accent)
			c.set(7, 3, pal.Accent)
			c.set(9, 3, pal.Accent)
			c.set(8, 2, pal.Accent)
			c.set(8, 4, pal.Accent)
		}

	case 10: // Culture — arched facade
		bx := 8 - (4+tier*6/20)/2
		bw := 4 + tier*6/20
		if bw > 10 {
			bw = 10
		}
		by2 := 14 - (3 + tier*5/20)
		c.set(bx, by2, 0)           // carve arch opening
		c.set(bx+bw-1, by2, 0)
		c.hline(bx+1, by2-1, bw-2, pal.Accent) // decorative top bar

	case 11: // Metallurgy — add fire/glow at base
		c.set(7, 13, 0xff6000)
		c.set(8, 13, 0xff8000)
		c.set(7, 12, 0xffb000)

	case 12: // Energy — add energy conduit lines
		if era >= 3 {
			c.hline(2, 8, 12, pal.Glow)
		} else {
			c.vline(3, 8, 4, pal.Accent) // pipe
			c.vline(12, 8, 4, pal.Accent)
		}

	case 13: // Hacker — show server rack pattern on body
		bx := 8 - (4+tier*6/20)/2
		bw := 4 + tier*6/20
		if bw > 10 {
			bw = 10
		}
		by2 := 14 - (3 + tier*5/20)
		for row := 0; row < 3; row++ {
			c.hline(bx+1, by2+row*2, bw-2, darken(pal.Wall, 0.7))
			c.set(bx+1, by2+row*2, pal.Glow) // status LED
		}
	}

	return c.Pixels
}

// PalacePixels returns a 16×16 palace sprite for the given era (0–6).
func PalacePixels(era int) [16][16]uint32 {
	switch era {
	case 0: // Primitive: chief's longhouse
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0x5a3a20, 0x5a3a20, 0x5a3a20, 0x5a3a20, 0x5a3a20, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x5a3a20, 0x8b6a40, 0x8b6a40, 0x8b6a40, 0x8b6a40, 0x8b6a40, 0x5a3a20, 0, 0, 0, 0, 0},
			{0, 0, 0, 0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0, 0, 0},
			{0, 0, 0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0, 0, 0},
			{0, 0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0, 0},
			{0x5a3a20, 0x8b6a40, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6a40, 0x5a3a20, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x5a3a20, 0x5a3a20, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x8b6040, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0xc8a060, 0x8b6040, 0, 0},
			{0x5a3a20, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x8b6040, 0x5a3a20, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 1: // Classical/Medieval: stone castle
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0xd4a017, 0, 0, 0, 0xd4a017, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0xd4a017, 0xd4a017, 0xd4a017, 0, 0xd4a017, 0xd4a017, 0xd4a017, 0, 0, 0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0x908070, 0xd0c8b0, 0x908070, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x908070, 0xd0c8b0, 0x908070, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xa0c0d0, 0xa0c0d0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd4a017, 0xd4a017, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd4a017, 0xd4a017, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x6b3a1a, 0x6b3a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x6b3a1a, 0x6b3a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x8b1a1a, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0xd0c8b0, 0x8b1a1a, 0},
			{0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0x908070, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 2: // Industrial: Victorian town hall with clock tower
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0xc87941, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x4a3a30, 0x8a6050, 0xc0c0a0, 0x8a6050, 0x4a3a30, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0x8a6050, 0x8a6050, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0, 0, 0, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0x8a6050, 0xc0c0a0, 0xc0c0a0, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x3a2a20, 0x3a2a20, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x3a2a20, 0x3a2a20, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0x8a6050, 0, 0},
			{0, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0x4a3a30, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 3: // Modern: glass capitol tower
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0xc0c8d0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x808890, 0x80c0e0, 0x808890, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x808890, 0x80c0e0, 0xe0f0ff, 0x80c0e0, 0x808890, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x808890, 0x80c0e0, 0x80c0e0, 0xe0f0ff, 0x80c0e0, 0x80c0e0, 0x808890, 0, 0, 0, 0, 0},
			{0, 0, 0x505860, 0x505860, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x505860, 0x505860, 0, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x404850, 0x404850, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x404850, 0x404850, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x808890, 0x808890, 0x808890, 0x808890, 0x80c0e0, 0x80c0e0, 0x808890, 0x505860, 0, 0},
			{0, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0x505860, 0, 0},
			{0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0x404850, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 4: // Digital/Cyber: neon megaplex
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0x40ffff, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x101020, 0x40c0ff, 0x101020, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x202030, 0x40c0ff, 0x40ffff, 0x40c0ff, 0x202030, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x202030, 0x40c0ff, 0x202030, 0x202030, 0x202030, 0x40c0ff, 0x202030, 0, 0, 0, 0, 0},
			{0, 0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0, 0},
			{0, 0x101020, 0x202030, 0x40c0ff, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x40c0ff, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x40c0ff, 0x202030, 0x40c0ff, 0x40c0ff, 0xff40c0, 0x40c0ff, 0x40c0ff, 0x202030, 0x40c0ff, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x40c0ff, 0x202030, 0x40c0ff, 0x40c0ff, 0xff40c0, 0x40c0ff, 0x40c0ff, 0x202030, 0x40c0ff, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x181828, 0x181828, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x202030, 0x202030, 0x202030, 0x202030, 0x181828, 0x181828, 0x202030, 0x202030, 0x202030, 0x202030, 0x202030, 0x101020, 0, 0},
			{0, 0x101020, 0x40c0ff, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x40c0ff, 0x101020, 0, 0},
			{0x40c0ff, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x101020, 0x40c0ff, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 5: // Fusion/Space: orbital command spire
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0x80c0ff, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0x4080ff, 0xffe040, 0x4080ff, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0x0a1530, 0x4080ff, 0x80c0ff, 0x4080ff, 0x0a1530, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0x0a1530, 0x4080ff, 0x0a1530, 0x80c0ff, 0x0a1530, 0x4080ff, 0x0a1530, 0, 0, 0, 0, 0},
			{0, 0, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x4080ff, 0xffe040, 0x4080ff, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x4080ff, 0xffe040, 0x4080ff, 0x4080ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x80c0ff, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x0a1530, 0x80c0ff, 0x0a1530, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0, 0},
			{0, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0x4080ff, 0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0, 0},
			{0x4080ff, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x0a1530, 0x4080ff, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	default: // Era 6: Cosmic — glowing transcendent pillar
		return [16][16]uint32{
			{0, 0, 0, 0, 0, 0, 0, 0xff80ff, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0xffd040, 0xff80ff, 0xffd040, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0xc0a0ff, 0xff80ff, 0xffffff, 0xff80ff, 0xc0a0ff, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0xc0a0ff, 0xffd040, 0x080810, 0xffffff, 0x080810, 0xffd040, 0xc0a0ff, 0, 0, 0, 0, 0},
			{0, 0, 0, 0xc0a0ff, 0x080810, 0x080810, 0x080810, 0xffffff, 0x080810, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0, 0},
			{0, 0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xc0a0ff, 0xffffff, 0xc0a0ff, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0},
			{0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xffd040, 0xffd040, 0xffffff, 0xffd040, 0xffd040, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0},
			{0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xffd040, 0xffd040, 0xffffff, 0xffd040, 0xffd040, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0},
			{0, 0, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0xc0a0ff, 0xffffff, 0xc0a0ff, 0xc0a0ff, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0},
			{0, 0, 0, 0xc0a0ff, 0x080810, 0x080810, 0x080810, 0xffffff, 0x080810, 0x080810, 0x080810, 0xc0a0ff, 0, 0, 0, 0},
			{0, 0, 0, 0, 0xc0a0ff, 0xffd040, 0x080810, 0xffffff, 0x080810, 0xffd040, 0xc0a0ff, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0xc0a0ff, 0xff80ff, 0xffffff, 0xff80ff, 0xc0a0ff, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0xffd040, 0xffd040, 0xffd040, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}
	}
}
