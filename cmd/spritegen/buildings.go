// cmd/spritegen/buildings.go — parametric 16×16 building sprite generator
package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
)

// ── Era color palettes ────────────────────────────────────────────────────────

type eraPalette struct {
	wall, roof, window, door, accent, glow uint32
}

var eraPalettes = [7]eraPalette{
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

var ageEraMap = map[string]int{
	"primitive_age": 0, "stone_age": 0, "bronze_age": 0, "iron_age": 0,
	"classical_age": 1, "medieval_age": 1, "renaissance_age": 1,
	"age_of_sail": 2, "colonial_age": 2, "industrial_age": 2, "victorian_age": 2, "gilded_age": 2,
	"electric_age": 3, "modern_age": 3, "atomic_age": 3, "information_age": 3,
	"digital_age": 4, "cyberpunk_age": 4,
	"nanotech_age": 5, "fusion_age": 5, "space_age": 5, "interstellar_age": 5,
	"galactic_age": 6, "quantum_age": 6, "singularity_age": 6, "transcendent_age": 6, "divine_age": 6, "cosmic_age": 6,
}

func getEra(ageKey string) int {
	if e, ok := ageEraMap[ageKey]; ok {
		return e
	}
	return 0
}

// ── Canvas and drawing helpers ────────────────────────────────────────────────

type canvas [16][16]uint32

// fillRect fills a rectangle. x,y = top-left, w,h = size. Clips to 16×16.
func (c *canvas) fillRect(x, y, w, h int, col uint32) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			rx, ry := x+dx, y+dy
			if rx >= 0 && rx < 16 && ry >= 0 && ry < 16 {
				c[ry][rx] = col
			}
		}
	}
}

// set draws a single pixel.
func (c *canvas) set(x, y int, col uint32) {
	if x >= 0 && x < 16 && y >= 0 && y < 16 {
		c[y][x] = col
	}
}

// hline draws a horizontal line.
func (c *canvas) hline(x, y, w int, col uint32) { c.fillRect(x, y, w, 1, col) }

// vline draws a vertical line.
func (c *canvas) vline(x, y, h int, col uint32) { c.fillRect(x, y, 1, h, col) }

// pitchedRoof draws a pitched roof: peak at (cx, peakY), base from (x, baseY) to (x+w-1, baseY).
func (c *canvas) pitchedRoof(cx, peakY, baseY, baseW int, col uint32) {
	for y := peakY; y <= baseY; y++ {
		spread := (y - peakY) * (baseW / 2) / (baseY - peakY + 1)
		c.hline(cx-spread, y, spread*2+1, col)
	}
}

// dome draws a dome: center cx,cy, radius r.
func (c *canvas) dome(cx, cy, r int, col uint32) {
	for dy := -r; dy <= 0; dy++ {
		hw := int(math.Sqrt(float64(r*r - dy*dy)))
		c.hline(cx-hw, cy+dy, hw*2+1, col)
	}
}

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
func buildingCompose(lin, tier, era, variation int) canvas {
	var c canvas
	pal := eraPalettes[era]

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
	c.fillRect(bx-1, 14, bodyW+2, 2, darken(pal.wall, 0.6))

	// Body walls
	c.fillRect(bx, by, bodyW, bodyH, pal.wall)

	// Body outline (slightly darker)
	c.hline(bx, by, bodyW, darken(pal.wall, 0.7))         // top edge
	c.vline(bx, by, bodyH, darken(pal.wall, 0.7))         // left edge
	c.vline(bx+bodyW-1, by, bodyH, darken(pal.wall, 0.7)) // right edge

	// Windows: 1×2 px each, spaced across body
	numWindows := 1 + tier*2/10 + variation%2
	if numWindows > 3 {
		numWindows = 3
	}
	for w := 0; w < numWindows; w++ {
		wx := bx + 1 + w*(bodyW-2)/(numWindows)
		wy := by + 1
		c.fillRect(wx, wy, 1, 2, pal.window)
	}

	// Door: 2 wide, at bottom of body, center
	dx := bx + bodyW/2 - 1
	c.fillRect(dx, by+bodyH-2, 2, 2, pal.door)

	// Roof — shape varies by lineage group
	roofTop := by - 1
	switch {
	case lin <= 2: // housing, food → pitched roof
		c.pitchedRoof(8, roofTop-2, roofTop, bodyW+2, pal.roof)
	case lin <= 4: // organic/geological extraction → flat roof with chimney
		c.hline(bx-1, roofTop, bodyW+2, pal.roof)
		cx2 := bx + bodyW - 2 + variation
		c.vline(cx2, roofTop-3, 3, darken(pal.roof, 0.8)) // chimney
		c.hline(cx2-1, roofTop-4, 3, pal.accent)           // chimney top
	case lin <= 6: // knowledge, faith → dome or spire
		if era >= 4 {
			c.vline(8, roofTop-4, 4, pal.accent) // antenna/spire
			c.set(7, roofTop-5, pal.glow)
			c.set(8, roofTop-5, pal.glow)
		} else {
			c.dome(8, roofTop, 2+bodyW/4, pal.roof)
		}
	case lin <= 9: // military, trade, engineering → flat industrial roofline
		c.hline(bx-1, roofTop, bodyW+2, pal.roof)
		if era >= 2 { // smokestacks for industrial+
			for i := 0; i < 2; i++ {
				sx := bx + 1 + i*(bodyW-3)
				c.vline(sx, roofTop-4+i, 4-i, darken(pal.roof, 0.7))
			}
		}
		if era >= 4 { // antenna array for digital+
			c.hline(bx+1, roofTop-2, bodyW-2, darken(pal.wall, 0.5))
			for i := 0; i < 3; i++ {
				c.vline(bx+2+i*2, roofTop-4, 2, pal.glow)
			}
		}
	case lin <= 11: // culture, metallurgy → ornate roof
		c.pitchedRoof(8, roofTop-3, roofTop, bodyW+2, pal.roof)
		c.set(8, roofTop-3, pal.accent) // finial
		c.set(7, roofTop-3, pal.accent)
	default: // energy, hacker → futuristic flat with glow
		c.hline(bx-1, roofTop, bodyW+2, pal.roof)
		if era >= 3 {
			c.hline(bx, roofTop-1, bodyW, pal.glow) // glow strip on roof
		}
	}

	// Era-specific extra details overlaid
	switch era {
	case 0: // primitive: keep clean
		// nothing extra
	case 1: // classical: add banner/flag on top
		c.vline(bx+bodyW-1, by-3, 3, pal.accent)
		c.hline(bx+bodyW-1, by-3, 2, pal.accent)
	case 2: // industrial: smoke puff above chimneys
		c.set(bx+bodyW-2+variation, roofTop-5, darken(pal.wall, 1.5))
	case 3: // modern: glass reflection on windows
		for w := 0; w < 3 && w < bodyW-2; w++ {
			wx := bx + 1 + w*(bodyW-2)/3
			c.set(wx, by+1, lighten(pal.window, 60))
		}
	case 4: // cyber: neon strip on body edge
		c.vline(bx+bodyW-1, by, bodyH, pal.glow)
	case 5: // fusion: energy beam from top
		c.vline(8, roofTop-5, 3, pal.glow)
		c.set(7, roofTop-5, pal.glow)
		c.set(9, roofTop-5, pal.glow)
	case 6: // cosmic: glow halo dots around building
		c.set(bx-1, by-1, pal.glow)
		c.set(bx+bodyW, by-1, pal.glow)
		c.set(bx-1, by+bodyH, pal.glow)
		c.set(bx+bodyW, by+bodyH, pal.glow)
	}

	return c
}

// ── Per-lineage overrides ─────────────────────────────────────────────────────

// GenerateBuildingSprite returns a 16×16 sprite for a building given its lineage (1-13),
// tier (0-20), age key, and variation (0, 1, or 2).
func GenerateBuildingSprite(lineage, tier int, ageKey string, variation int) [16][16]uint32 {
	era := getEra(ageKey)
	pal := eraPalettes[era]
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
			c.vline(4, 10, 5, darken(pal.roof, 0.8))
			c.hline(4, 10, 3, darken(pal.roof, 0.8))
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
			c.set(cx2, by2, pal.accent)
			c.set(cx2-1, by2+1, pal.accent)
			c.set(cx2, by2+1, pal.accent)
			c.set(cx2+1, by2+1, pal.accent)
			c.set(cx2, by2+2, pal.accent)
		} else if era >= 4 { // cyber/future faith: star/glow
			c.set(8, 2, pal.glow)
			c.set(7, 3, pal.glow)
			c.set(9, 3, pal.glow)
			c.set(8, 4, pal.glow)
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
				c.set(bx+i, by2-1, pal.wall)
			}
		}

	case 8: // Trade — add coin/scale symbol
		if era <= 3 {
			c.set(8, 4, 0xd4a017) // gold coin dot
			c.set(7, 4, 0xd4a017)
		}

	case 9: // Engineering — add gear on roof
		if era >= 2 {
			c.set(8, 3, pal.accent)
			c.set(7, 3, pal.accent)
			c.set(9, 3, pal.accent)
			c.set(8, 2, pal.accent)
			c.set(8, 4, pal.accent)
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
		c.hline(bx+1, by2-1, bw-2, pal.accent) // decorative top bar

	case 11: // Metallurgy — add fire/glow at base
		c.set(7, 13, 0xff6000)
		c.set(8, 13, 0xff8000)
		c.set(7, 12, 0xffb000)

	case 12: // Energy — add energy conduit lines
		if era >= 3 {
			c.hline(2, 8, 12, pal.glow)
		} else {
			c.vline(3, 8, 4, pal.accent) // pipe
			c.vline(12, 8, 4, pal.accent)
		}

	case 13: // Hacker — show server rack pattern on body
		bx := 8 - (4+tier*6/20)/2
		bw := 4 + tier*6/20
		if bw > 10 {
			bw = 10
		}
		by2 := 14 - (3 + tier*5/20)
		for row := 0; row < 3; row++ {
			c.hline(bx+1, by2+row*2, bw-2, darken(pal.wall, 0.7))
			c.set(bx+1, by2+row*2, pal.glow) // status LED
		}
	}

	return c
}

// ── Building catalog ──────────────────────────────────────────────────────────

type buildingDef struct {
	key     string
	lineage int
	tier    int
	ageKey  string
}

var allBuildings = []buildingDef{
	// Lineage 1: Housing
	{"hut", 1, 0, "primitive_age"},
	{"longhouse", 1, 1, "stone_age"},
	{"house", 1, 2, "bronze_age"},
	{"townhouse", 1, 3, "iron_age"},
	{"villa", 1, 4, "classical_age"},
	{"manor", 1, 5, "medieval_age"},
	{"estate", 1, 6, "renaissance_age"},
	{"settlement_block", 1, 7, "colonial_age"},
	{"tenement", 1, 8, "industrial_age"},
	{"row_house", 1, 9, "victorian_age"},
	{"apartment_block", 1, 10, "electric_age"},
	{"housing_project", 1, 11, "atomic_age"},
	{"tower_block", 1, 12, "modern_age"},
	{"smart_complex", 1, 13, "information_age"},
	{"megaplex", 1, 14, "digital_age"},
	{"arcology_pod", 1, 15, "cyberpunk_age"},
	{"habitat_ring", 1, 16, "fusion_age"},
	{"orbital_habitat", 1, 17, "space_age"},
	{"generation_ship", 1, 18, "interstellar_age"},
	{"dyson_sphere_habitat", 1, 19, "galactic_age"},
	{"reality_fold", 1, 20, "quantum_age"},
	{"transcendent_nexus", 1, 21, "transcendent_age"},

	// Lineage 2: Food
	{"gathering_camp", 2, 0, "primitive_age"},
	{"forager_post", 2, 1, "stone_age"},
	{"farm", 2, 2, "bronze_age"},
	{"field_works", 2, 3, "iron_age"},
	{"estate_farm", 2, 4, "classical_age"},
	{"demesne", 2, 5, "medieval_age"},
	{"market_garden", 2, 6, "renaissance_age"},
	{"plantation", 2, 7, "colonial_age"},
	{"agricultural_works", 2, 8, "industrial_age"},
	{"mechanized_farm", 2, 9, "victorian_age"},
	{"industrial_farm", 2, 10, "electric_age"},
	{"agricultural_complex", 2, 11, "atomic_age"},
	{"agri_complex", 2, 12, "modern_age"},
	{"smart_farm", 2, 13, "information_age"},
	{"nano_farm", 2, 14, "digital_age"},
	{"vat_farm", 2, 15, "cyberpunk_age"},
	{"bio_reactor_farm", 2, 16, "fusion_age"},
	{"hydroponic_bay", 2, 17, "space_age"},
	{"protein_synthesizer", 2, 18, "interstellar_age"},
	{"matter_converter", 2, 19, "galactic_age"},
	{"quantum_cultivator", 2, 20, "quantum_age"},

	// Lineage 3: Organic Extraction
	{"wood_camp", 3, 0, "primitive_age"},
	{"woodcutter_camp", 3, 1, "stone_age"},
	{"lumber_mill", 3, 2, "bronze_age"},
	{"timber_yard", 3, 3, "iron_age"},
	{"wood_workshop", 3, 4, "classical_age"},
	{"sawmill", 3, 5, "medieval_age"},
	{"coal_mine", 3, 6, "renaissance_age"},
	{"coal_works", 3, 7, "colonial_age"},
	{"steam_coal_plant", 3, 8, "industrial_age"},
	{"oil_derrick", 3, 9, "victorian_age"},
	{"oil_field", 3, 10, "electric_age"},
	{"petroleum_refinery", 3, 11, "atomic_age"},
	{"oil_platform", 3, 12, "modern_age"},
	{"smart_refinery", 3, 13, "information_age"},
	{"bio_fabrication_lab", 3, 14, "digital_age"},
	{"nanobot_vat", 3, 15, "cyberpunk_age"},
	{"molecular_synthesizer", 3, 16, "fusion_age"},
	{"quantum_organic_extractor", 3, 17, "space_age"},
	{"reality_matter_weaver", 3, 18, "interstellar_age"},
	{"cosmic_organic_works", 3, 19, "galactic_age"},
	{"reality_harvester", 3, 20, "quantum_age"},

	// Lineage 4: Geological
	{"stone_camp", 4, 0, "stone_age"},
	{"stone_pit", 4, 1, "stone_age"},
	{"quarry", 4, 2, "bronze_age"},
	{"marble_quarry", 4, 3, "iron_age"},
	{"marble_works", 4, 4, "classical_age"},
	{"stonemasons_guild", 4, 5, "medieval_age"},
	{"iron_mine", 4, 6, "renaissance_age"},
	{"deep_iron_mine", 4, 7, "colonial_age"},
	{"steam_mine", 4, 8, "industrial_age"},
	{"uranium_mine", 4, 9, "victorian_age"},
	{"nuclear_extraction_plant", 4, 10, "electric_age"},
	{"uranium_processing_works", 4, 11, "atomic_age"},
	{"titanium_mine", 4, 12, "modern_age"},
	{"precision_mine", 4, 13, "information_age"},
	{"nano_drill_complex", 4, 14, "digital_age"},
	{"dark_crystal_mine", 4, 15, "cyberpunk_age"},
	{"exotic_mineral_extractor", 4, 16, "fusion_age"},
	{"asteroid_crystal_mine", 4, 17, "space_age"},
	{"stellar_core_drill", 4, 18, "interstellar_age"},
	{"neutron_star_mine", 4, 19, "galactic_age"},
	{"reality_excavator", 4, 20, "quantum_age"},

	// Lineage 5: Knowledge
	{"story_circle", 5, 0, "primitive_age"},
	{"elders_hall", 5, 1, "stone_age"},
	{"scriptorium", 5, 2, "bronze_age"},
	{"agora", 5, 3, "iron_age"},
	{"library", 5, 4, "classical_age"},
	{"monastery_library", 5, 5, "medieval_age"},
	{"university", 5, 6, "renaissance_age"},
	{"natural_philosophy_hall", 5, 7, "colonial_age"},
	{"research_institute", 5, 8, "industrial_age"},
	{"academy", 5, 9, "victorian_age"},
	{"physics_laboratory", 5, 10, "electric_age"},
	{"research_campus", 5, 11, "atomic_age"},
	{"think_tank", 5, 12, "modern_age"},
	{"innovation_hub", 5, 13, "information_age"},
	{"ai_research_lab", 5, 14, "digital_age"},
	{"neuro_research_center", 5, 15, "cyberpunk_age"},
	{"theoretical_institute", 5, 16, "fusion_age"},
	{"deep_space_observatory", 5, 17, "space_age"},
	{"xenology_institute", 5, 18, "interstellar_age"},
	{"cosmic_research_station", 5, 19, "galactic_age"},
	{"reality_academy", 5, 20, "quantum_age"},

	// Lineage 6: Faith
	{"shrine", 6, 0, "primitive_age"},
	{"standing_stones", 6, 1, "stone_age"},
	{"altar", 6, 2, "bronze_age"},
	{"temple", 6, 3, "iron_age"},
	{"oracle_house", 6, 4, "classical_age"},
	{"cathedral", 6, 5, "medieval_age"},
	{"basilica", 6, 6, "renaissance_age"},
	{"mission", 6, 7, "colonial_age"},
	{"church", 6, 8, "industrial_age"},
	{"grand_cathedral", 6, 9, "victorian_age"},
	{"revival_hall", 6, 10, "electric_age"},
	{"spiritual_center", 6, 11, "atomic_age"},
	{"meditation_center", 6, 12, "modern_age"},
	{"digital_temple", 6, 13, "information_age"},
	{"cyber_shrine", 6, 14, "digital_age"},
	{"neon_sanctuary", 6, 15, "cyberpunk_age"},
	{"quantum_chapel", 6, 16, "fusion_age"},
	{"orbital_sanctuary", 6, 17, "space_age"},
	{"void_monastery", 6, 18, "interstellar_age"},
	{"stellar_shrine", 6, 19, "galactic_age"},
	{"transcendence_hall", 6, 20, "quantum_age"},

	// Lineage 7: Military
	{"war_camp", 7, 0, "stone_age"},
	{"barracks", 7, 1, "bronze_age"},
	{"hunting_lodge", 7, 2, "iron_age"},
	{"legion_fort", 7, 3, "iron_age"},
	{"military_academy", 7, 4, "classical_age"},
	{"castle_keep", 7, 5, "medieval_age"},
	{"fortress", 7, 6, "renaissance_age"},
	{"fort", 7, 7, "colonial_age"},
	{"military_base", 7, 8, "industrial_age"},
	{"garrison", 7, 9, "victorian_age"},
	{"command_post", 7, 10, "electric_age"},
	{"bunker_complex", 7, 11, "atomic_age"},
	{"special_ops_hq", 7, 12, "modern_age"},
	{"cyber_command", 7, 13, "information_age"},
	{"drone_warfare_center", 7, 14, "digital_age"},
	{"combat_aug_center", 7, 15, "cyberpunk_age"},
	{"plasma_command", 7, 16, "fusion_age"},
	{"space_force_base", 7, 17, "space_age"},
	{"fleet_command", 7, 18, "interstellar_age"},
	{"stellar_armada_hq", 7, 19, "galactic_age"},
	{"probability_war_room", 7, 20, "quantum_age"},
	{"omniversal_war_council", 7, 21, "transcendent_age"},

	// Lineage 8: Trade
	{"market", 8, 0, "bronze_age"},
	{"trading_post", 8, 1, "iron_age"},
	{"merchant_quarter", 8, 2, "classical_age"},
	{"guildhall", 8, 3, "medieval_age"},
	{"exchange", 8, 4, "renaissance_age"},
	{"port", 8, 5, "colonial_age"},
	{"stock_exchange", 8, 6, "industrial_age"},
	{"bank", 8, 7, "victorian_age"},
	{"financial_district", 8, 8, "electric_age"},
	{"corporate_hq", 8, 9, "atomic_age"},
	{"investment_firm", 8, 10, "modern_age"},
	{"venture_hub", 8, 11, "information_age"},
	{"crypto_exchange", 8, 12, "digital_age"},
	{"black_market", 8, 13, "cyberpunk_age"},
	{"energy_exchange", 8, 14, "fusion_age"},
	{"asteroid_market", 8, 15, "space_age"},
	{"galactic_trade_hub", 8, 16, "interstellar_age"},
	{"stellar_exchange", 8, 17, "galactic_age"},
	{"probability_market", 8, 18, "quantum_age"},
	{"omniversal_bazaar", 8, 19, "transcendent_age"},

	// Lineage 9: Engineering
	{"smithy", 9, 0, "bronze_age"},
	{"ironworks", 9, 1, "iron_age"},
	{"aqueduct", 9, 2, "classical_age"},
	{"workshop", 9, 3, "medieval_age"},
	{"mill", 9, 4, "renaissance_age"},
	{"dockyard", 9, 5, "colonial_age"},
	{"iron_works_complex", 9, 6, "industrial_age"},
	{"steam_works", 9, 7, "victorian_age"},
	{"power_station", 9, 8, "electric_age"},
	{"nuclear_plant", 9, 9, "atomic_age"},
	{"power_grid_hub", 9, 10, "modern_age"},
	{"smart_grid_node", 9, 11, "information_age"},
	{"neural_grid", 9, 12, "digital_age"},
	{"augmentation_foundry", 9, 13, "cyberpunk_age"},
	{"fusion_reactor", 9, 14, "fusion_age"},
	{"launch_complex", 9, 15, "space_age"},
	{"warp_drive_plant", 9, 16, "interstellar_age"},
	{"dyson_assembly", 9, 17, "galactic_age"},
	{"reality_forge", 9, 18, "quantum_age"},
	{"singularity_engine", 9, 19, "transcendent_age"},

	// Lineage 10: Culture
	{"amphitheater", 10, 0, "classical_age"},
	{"great_hall", 10, 1, "medieval_age"},
	{"art_studio", 10, 2, "renaissance_age"},
	{"concert_hall", 10, 3, "colonial_age"},
	{"opera_house", 10, 4, "industrial_age"},
	{"grand_museum", 10, 5, "victorian_age"},
	{"radio_station", 10, 6, "electric_age"},
	{"cinema", 10, 7, "atomic_age"},
	{"tv_studio", 10, 8, "modern_age"},
	{"media_center", 10, 9, "information_age"},
	{"vr_studio", 10, 10, "digital_age"},
	{"holographic_theater", 10, 11, "cyberpunk_age"},
	{"neural_art_complex", 10, 12, "fusion_age"},
	{"zero_g_gallery", 10, 13, "space_age"},
	{"cultural_beacon", 10, 14, "interstellar_age"},
	{"civilization_archive", 10, 15, "galactic_age"},
	{"reality_art_engine", 10, 16, "quantum_age"},

	// Lineage 11: Metallurgy
	{"smelter", 11, 0, "iron_age"},
	{"forge", 11, 1, "classical_age"},
	{"ironmonger", 11, 2, "medieval_age"},
	{"foundry", 11, 3, "renaissance_age"},
	{"iron_works", 11, 4, "colonial_age"},
	{"steel_mill", 11, 5, "industrial_age"},
	{"bessemer_plant", 11, 6, "victorian_age"},
	{"electric_arc_furnace", 11, 7, "electric_age"},
	{"advanced_alloy_plant", 11, 8, "atomic_age"},
	{"titanium_smelter", 11, 9, "modern_age"},
	{"aerospace_foundry", 11, 10, "information_age"},
	{"nano_alloy_plant", 11, 11, "digital_age"},
	{"dark_matter_refinery", 11, 12, "cyberpunk_age"},
	{"exotic_matter_forge", 11, 13, "fusion_age"},
	{"orbital_refinery", 11, 14, "space_age"},
	{"antimatter_forge", 11, 15, "interstellar_age"},
	{"stellar_metallurgy", 11, 16, "galactic_age"},
	{"quantum_metal_works", 11, 17, "quantum_age"},

	// Lineage 12: Energy
	{"coal_plant", 12, 0, "industrial_age"},
	{"steam_turbine", 12, 1, "victorian_age"},
	{"power_generator", 12, 2, "electric_age"},
	{"nuclear_reactor", 12, 3, "atomic_age"},
	{"oil_refinery", 12, 4, "modern_age"},
	{"smart_energy_grid", 12, 5, "information_age"},
	{"quantum_battery_array", 12, 6, "digital_age"},
	{"dark_energy_tap", 12, 7, "cyberpunk_age"},
	{"fusion_reactor_array", 12, 8, "fusion_age"},
	{"solar_collector_array", 12, 9, "space_age"},
	{"pulsar_tap", 12, 10, "interstellar_age"},
	{"quasar_tap", 12, 11, "galactic_age"},
	{"zero_point_generator", 12, 12, "quantum_age"},

	// Lineage 13: Hacker
	{"server_farm", 13, 0, "information_age"},
	{"data_center", 13, 1, "digital_age"},
	{"cyber_hub", 13, 2, "cyberpunk_age"},
	{"quantum_server_farm", 13, 3, "fusion_age"},
	{"orbital_data_relay", 13, 4, "space_age"},
	{"galactic_network_node", 13, 5, "interstellar_age"},
	{"consciousness_upload_hub", 13, 6, "galactic_age"},
	{"reality_processor", 13, 7, "quantum_age"},
}

// ── Main output function ──────────────────────────────────────────────────────

// GenerateAllBuildingSprites generates 3 variations of each building sprite and saves them as PNGs.
func GenerateAllBuildingSprites(outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	total := 0
	for _, b := range allBuildings {
		// Generate 3 variations (0, 1, 2)
		for v := 0; v < 3; v++ {
			pixels := GenerateBuildingSprite(b.lineage, b.tier, b.ageKey, v)
			// Convert canvas to image.NRGBA
			img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
			for row := 0; row < 16; row++ {
				for col := 0; col < 16; col++ {
					pv := pixels[row][col]
					if pv == 0 {
						img.SetNRGBA(col, row, color.NRGBA{0, 0, 0, 0})
					} else {
						img.SetNRGBA(col, row, color.NRGBA{uint8(pv >> 16), uint8(pv >> 8), uint8(pv), 255})
					}
				}
			}
			suffix := ""
			if v > 0 {
				suffix = fmt.Sprintf("_v%d", v+1)
			}
			fname := filepath.Join(outDir, b.key+suffix+".png")
			if err := savePNG(fname, img); err != nil {
				return err
			}
		}
		total++
	}
	fmt.Printf("Generated %d building sprites (%d files) → %s\n", total, total*3, outDir)
	return nil
}
