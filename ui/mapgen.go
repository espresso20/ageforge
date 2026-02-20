package ui

import (
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// TerrainPalette defines the color palette for a terrain era
type TerrainPalette struct {
	Ground, GroundAlt                      color.RGBA
	Water, WaterLight, WaterDeep           color.RGBA
	Tree, TreeDark, TreeLight              color.RGBA
	Road, RoadEdge                         color.RGBA
	Hill, HillLight                        color.RGBA
	Farmland, FarmAlt                      color.RGBA
	Accent1, Accent2                       color.RGBA // era-specific accents
}

// MapGenConfig holds parameters for map generation
type MapGenConfig struct {
	Width, Height int
	DetailLevel   int // 0=mini, 1=full
	Buildings     map[string]game.BuildingState
	AgeKey        string
}

func eraFromAge(ageKey string) int {
	ages := config.AgeByKey()
	if a, ok := ages[ageKey]; ok {
		o := a.Order
		switch {
		case o <= 1:
			return 0
		case o <= 4:
			return 1
		case o <= 7:
			return 2
		case o <= 10:
			return 3
		case o <= 13:
			return 4
		case o <= 15:
			return 5
		case o == 16:
			return 6
		case o <= 19:
			return 7
		default:
			return 8
		}
	}
	return 0
}

func c(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

func getTerrainPalette(era int) TerrainPalette {
	switch era {
	case 0: // primitive — lush green wilderness
		return TerrainPalette{
			Ground: c(34, 85, 34), GroundAlt: c(44, 100, 42),
			Water: c(28, 55, 150), WaterLight: c(50, 80, 175), WaterDeep: c(18, 40, 120),
			Tree: c(20, 110, 30), TreeDark: c(12, 75, 18), TreeLight: c(35, 130, 45),
			Road: c(85, 65, 42), RoadEdge: c(65, 50, 32),
			Hill: c(50, 105, 50), HillLight: c(65, 120, 62),
			Farmland: c(90, 120, 40), FarmAlt: c(80, 110, 35),
			Accent1: c(180, 120, 40), Accent2: c(200, 100, 30), // campfire orange
		}
	case 1: // ancient — sandy fields and farms
		return TerrainPalette{
			Ground: c(110, 130, 55), GroundAlt: c(120, 140, 62),
			Water: c(30, 68, 140), WaterLight: c(45, 85, 162), WaterDeep: c(20, 50, 110),
			Tree: c(50, 100, 30), TreeDark: c(38, 78, 22), TreeLight: c(65, 118, 42),
			Road: c(145, 125, 82), RoadEdge: c(115, 98, 65),
			Hill: c(140, 120, 80), HillLight: c(160, 140, 95),
			Farmland: c(150, 170, 60), FarmAlt: c(140, 160, 50),
			Accent1: c(180, 160, 100), Accent2: c(160, 140, 80), // sandstone
		}
	case 2: // medieval — dark forest and stone
		return TerrainPalette{
			Ground: c(30, 60, 25), GroundAlt: c(38, 72, 32),
			Water: c(20, 48, 128), WaterLight: c(32, 65, 148), WaterDeep: c(12, 35, 100),
			Tree: c(15, 80, 20), TreeDark: c(8, 52, 12), TreeLight: c(25, 98, 30),
			Road: c(105, 95, 72), RoadEdge: c(80, 72, 55),
			Hill: c(45, 75, 40), HillLight: c(58, 90, 52),
			Farmland: c(70, 95, 30), FarmAlt: c(80, 105, 35),
			Accent1: c(140, 130, 110), Accent2: c(120, 110, 95), // stone walls
		}
	case 3: // industrial — smoke and brick
		return TerrainPalette{
			Ground: c(68, 68, 62), GroundAlt: c(78, 78, 72),
			Water: c(38, 58, 88), WaterLight: c(50, 70, 102), WaterDeep: c(28, 42, 68),
			Tree: c(48, 78, 38), TreeDark: c(35, 58, 28), TreeLight: c(58, 90, 48),
			Road: c(80, 75, 70), RoadEdge: c(60, 58, 55),
			Hill: c(85, 85, 78), HillLight: c(98, 98, 90),
			Farmland: c(75, 85, 55), FarmAlt: c(82, 92, 60),
			Accent1: c(140, 60, 40), Accent2: c(90, 90, 90), // brick + steel
		}
	case 4: // modern — asphalt and glass
		return TerrainPalette{
			Ground: c(52, 58, 62), GroundAlt: c(62, 68, 72),
			Water: c(28, 78, 148), WaterLight: c(42, 95, 168), WaterDeep: c(18, 58, 118),
			Tree: c(38, 88, 38), TreeDark: c(28, 68, 28), TreeLight: c(50, 105, 50),
			Road: c(50, 50, 52), RoadEdge: c(70, 70, 72),
			Hill: c(72, 78, 82), HillLight: c(85, 90, 95),
			Farmland: c(60, 80, 45), FarmAlt: c(68, 88, 52),
			Accent1: c(140, 180, 220), Accent2: c(200, 200, 210), // glass + white
		}
	case 5: // digital — server glow
		return TerrainPalette{
			Ground: c(10, 15, 38), GroundAlt: c(15, 22, 48),
			Water: c(18, 38, 118), WaterLight: c(28, 52, 138), WaterDeep: c(10, 25, 88),
			Tree: c(8, 28, 58), TreeDark: c(5, 18, 42), TreeLight: c(12, 38, 72),
			Road: c(0, 80, 130), RoadEdge: c(0, 60, 100),
			Hill: c(18, 25, 55), HillLight: c(25, 35, 68),
			Farmland: c(15, 35, 60), FarmAlt: c(18, 42, 72),
			Accent1: c(0, 200, 255), Accent2: c(0, 150, 200), // cyan glow
		}
	case 6: // cyberpunk — neon on black
		return TerrainPalette{
			Ground: c(8, 8, 12), GroundAlt: c(14, 12, 20),
			Water: c(78, 0, 118), WaterLight: c(100, 0, 148), WaterDeep: c(55, 0, 85),
			Tree: c(0, 38, 18), TreeDark: c(0, 22, 10), TreeLight: c(0, 55, 28),
			Road: c(30, 25, 35), RoadEdge: c(20, 18, 25),
			Hill: c(15, 12, 22), HillLight: c(22, 18, 32),
			Farmland: c(0, 50, 25), FarmAlt: c(0, 60, 30),
			Accent1: c(255, 0, 128), Accent2: c(0, 255, 80), // neon pink + green
		}
	case 7: // space — starfield and domes
		return TerrainPalette{
			Ground: c(5, 5, 15), GroundAlt: c(8, 8, 22),
			Water: c(15, 25, 78), WaterLight: c(22, 35, 98), WaterDeep: c(8, 15, 55),
			Tree: c(8, 8, 28), TreeDark: c(5, 5, 18), TreeLight: c(12, 12, 38),
			Road: c(58, 78, 138), RoadEdge: c(42, 58, 105),
			Hill: c(10, 10, 28), HillLight: c(15, 15, 38),
			Farmland: c(10, 18, 35), FarmAlt: c(12, 22, 42),
			Accent1: c(100, 160, 255), Accent2: c(200, 220, 255), // blue-white
		}
	case 8: // cosmic — transcendent void
		return TerrainPalette{
			Ground: c(3, 3, 8), GroundAlt: c(8, 5, 15),
			Water: c(38, 10, 78), WaterLight: c(58, 15, 98), WaterDeep: c(25, 5, 55),
			Tree: c(5, 5, 15), TreeDark: c(3, 3, 10), TreeLight: c(8, 8, 22),
			Road: c(78, 58, 18), RoadEdge: c(58, 42, 12),
			Hill: c(8, 5, 18), HillLight: c(12, 8, 25),
			Farmland: c(10, 8, 20), FarmAlt: c(12, 10, 25),
			Accent1: c(255, 200, 50), Accent2: c(200, 150, 255), // gold + violet
		}
	}
	return getTerrainPalette(0)
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func noise2D(x, y int, seed uint64) float64 {
	h1 := hashKey(string(rune(seed)) + string(rune(x*7919+y*6271)))
	h2 := hashKey(string(rune(seed+77)) + string(rune((x/3)*4909+(y/3)*3571)))
	return float64(h1%10000)/10000.0*0.6 + float64(h2%10000)/10000.0*0.4
}

type bldInfo struct {
	key      string
	category string
	x, y     int
	size     int
}

// GenerateMapImage creates a procedural pixel map as an image.RGBA
func GenerateMapImage(cfg MapGenConfig) *image.RGBA {
	w, h := cfg.Width, cfg.Height
	if w < 4 || h < 4 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	era := eraFromAge(cfg.AgeKey)
	pal := getTerrainPalette(era)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	seed := hashKey(cfg.AgeKey)
	cx, cy := w/2, h/2
	dl := cfg.DetailLevel

	// ═══════════════════════════════════════════
	// 1. BASE TERRAIN
	// ═══════════════════════════════════════════
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := noise2D(x, y, seed)
			elev := noise2D(x/4, y/4, seed+200)
			if elev > 0.65 {
				hillT := (elev - 0.65) / 0.35
				base := lerp(pal.Ground, pal.GroundAlt, n)
				img.SetRGBA(x, y, lerp(base, lerp(pal.Hill, pal.HillLight, n), hillT))
			} else {
				img.SetRGBA(x, y, lerp(pal.Ground, pal.GroundAlt, n))
			}
		}
	}

	// ═══════════════════════════════════════════
	// 2. ERA-SPECIFIC BACKGROUND FEATURES
	// ═══════════════════════════════════════════
	switch {
	case era >= 7: // space/cosmic — stars and nebulae
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				n := noise2D(x, y, seed+99)
				if n < 0.03 {
					br := uint8(90 + int(n*5500))
					if era == 8 {
						img.SetRGBA(x, y, c(br, uint8(float64(br)*0.8), uint8(float64(br)*0.3)))
					} else {
						img.SetRGBA(x, y, c(br, br, uint8(min(255, int(float64(br)*1.2)))))
					}
				}
				// Nebula clouds
				if era == 8 {
					neb := noise2D(x/6, y/6, seed+500)
					if neb > 0.7 {
						t := (neb - 0.7) / 0.3 * 0.15
						existing := img.RGBAAt(x, y)
						img.SetRGBA(x, y, lerp(existing, pal.Accent2, t))
					}
				}
			}
		}
	case era == 6: // cyberpunk — grid lines on ground
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x%(w/12) == 0 || y%(h/12) == 0 {
					existing := img.RGBAAt(x, y)
					img.SetRGBA(x, y, lerp(existing, pal.Accent1, 0.08))
				}
			}
		}
	case era == 5: // digital — faint circuit traces
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x%(w/8) == 0 || y%(h/8) == 0 {
					existing := img.RGBAAt(x, y)
					img.SetRGBA(x, y, lerp(existing, pal.Accent1, 0.06))
				}
			}
		}
	}

	// ═══════════════════════════════════════════
	// 3. RIVER
	// ═══════════════════════════════════════════
	riverBaseX := float64(w) * 0.28
	riverW := 3 + dl*3
	bankW := 1 + dl
	// No river in space/cosmic
	if era < 7 {
		for y := 0; y < h; y++ {
			rx := riverBaseX + math.Sin(float64(y)*0.06)*float64(w)*0.10 + math.Sin(float64(y)*0.15)*float64(w)*0.03
			for dx := -bankW; dx < riverW+bankW; dx++ {
				px := int(rx) + dx
				if px < 0 || px >= w {
					continue
				}
				if dx < 0 || dx >= riverW {
					existing := img.RGBAAt(px, y)
					img.SetRGBA(px, y, lerp(existing, pal.Water, 0.3))
				} else {
					centerDist := math.Abs(float64(dx)-float64(riverW)/2.0) / (float64(riverW) / 2.0)
					wc := lerp(pal.WaterDeep, pal.WaterLight, centerDist)
					wc = lerp(wc, pal.Water, noise2D(px, y, seed+7)*0.2)
					img.SetRGBA(px, y, wc)
				}
			}
		}
	}

	// ═══════════════════════════════════════════
	// 4. VEGETATION (era-specific)
	// ═══════════════════════════════════════════
	drawVegetation(img, w, h, era, dl, pal, seed)

	// ═══════════════════════════════════════════
	// 5. COLLECT BUILDING PLACEMENTS
	// ═══════════════════════════════════════════
	placements := placeBuildingsRadial(cfg.Buildings, cx, cy, w, h, era, dl)

	// Sort: furthest first so close buildings draw on top
	sort.Slice(placements, func(i, j int) bool {
		di := (placements[i].x-cx)*(placements[i].x-cx) + (placements[i].y-cy)*(placements[i].y-cy)
		dj := (placements[j].x-cx)*(placements[j].x-cx) + (placements[j].y-cy)*(placements[j].y-cy)
		return di > dj
	})

	// ═══════════════════════════════════════════
	// 6. SURROUNDINGS (farmland, parking lots, etc)
	// ═══════════════════════════════════════════
	drawSurroundings(img, w, h, era, dl, pal, seed, placements)

	// ═══════════════════════════════════════════
	// 7. INFRASTRUCTURE (paths → roads → rails → highways → glowing lines)
	// ═══════════════════════════════════════════
	drawInfrastructure(img, w, h, era, dl, pal, cx, cy, placements)

	// ═══════════════════════════════════════════
	// 8. BUILDINGS (era-specific rendering)
	// ═══════════════════════════════════════════
	drawBuildings(img, w, h, era, dl, pal, placements)

	// ═══════════════════════════════════════════
	// 9. ERA DECORATIONS (smokestacks, power lines, neon, etc)
	// ═══════════════════════════════════════════
	drawDecorations(img, w, h, era, dl, pal, seed, placements)

	return img
}

// ─── Vegetation ──────────────────────────────────────────
func drawVegetation(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, seed uint64) {
	// Primitive/ancient/medieval: dense forests
	// Industrial: sparse remaining trees
	// Modern+: decorative parks only
	// Digital/cyber: none
	// Space/cosmic: none
	if era >= 5 {
		return
	}

	density := 0.10
	switch era {
	case 0:
		density = 0.14
	case 1:
		density = 0.08
	case 2:
		density = 0.10
	case 3:
		density = 0.025
	case 4:
		density = 0.015
	}
	treeR := 2 + dl
	step := treeR + 1

	for y := treeR; y < h-treeR; y += step {
		for x := treeR; x < w-treeR; x += step {
			n := noise2D(x, y, seed+42)
			if n >= density {
				continue
			}
			// Check not in water
			px := img.RGBAAt(x, y)
			if px == pal.Water || px == pal.WaterLight || px == pal.WaterDeep {
				continue
			}
			// Draw canopy
			for dy := -treeR; dy <= treeR; dy++ {
				for dx := -treeR; dx <= treeR; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(treeR) {
						continue
					}
					tx, ty := x+dx, y+dy
					if tx < 0 || tx >= w || ty < 0 || ty >= h {
						continue
					}
					shade := float64(dy+treeR) / float64(treeR*2)
					tc := lerp(pal.TreeLight, pal.TreeDark, shade)
					edgeFade := d / float64(treeR)
					existing := img.RGBAAt(tx, ty)
					img.SetRGBA(tx, ty, lerp(existing, tc, 1.0-edgeFade*0.4))
				}
			}
			img.SetRGBA(x, y, pal.TreeDark)
		}
	}
}

// ─── Building placement ──────────────────────────────────
func placeBuildingsRadial(buildings map[string]game.BuildingState, cx, cy, w, h, era, dl int) []bldInfo {
	var placements []bldInfo
	maxDist := float64(min(w, h)) / 2.0

	for key, bs := range buildings {
		if !bs.Unlocked || bs.Count == 0 {
			continue
		}
		ringMin, ringMax := 0.08, 0.35
		switch bs.Category {
		case "housing":
			ringMin, ringMax = 0.05, 0.22
		case "production":
			ringMin, ringMax = 0.12, 0.32
		case "military":
			ringMin, ringMax = 0.22, 0.42
		case "wonder":
			ringMin, ringMax = 0.03, 0.16
		case "research":
			ringMin, ringMax = 0.08, 0.26
		case "storage":
			ringMin, ringMax = 0.06, 0.20
		}

		for i := 0; i < bs.Count; i++ {
			bHash := hashKey(key + string(rune(i)))
			angle := float64(bHash%3600) / 3600.0 * 2.0 * math.Pi
			distRatio := ringMin + float64(bHash%1000)/1000.0*(ringMax-ringMin)
			dist := distRatio * maxDist
			bx := cx + int(math.Cos(angle)*dist)
			by := cy + int(math.Sin(angle)*dist*0.7)

			// Building size scales with era
			size := 3
			switch {
			case era >= 6: // cyber+ = mega towers
				size = 5
			case era >= 4: // modern = skyscrapers
				size = 4
			case era >= 3: // industrial = bigger
				size = 4
			}
			if bs.Category == "wonder" {
				size += 3
			} else if bs.Category == "military" {
				size += 1
			}
			if dl > 0 {
				size += 2
			}

			placements = append(placements, bldInfo{key, bs.Category, bx, by, size})
		}
	}
	return placements
}

// ─── Surroundings ────────────────────────────────────────
func drawSurroundings(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, seed uint64, placements []bldInfo) {
	for _, b := range placements {
		radius := b.size + 4 + dl*2

		switch {
		case era <= 2 && b.category == "production":
			// Farmland with crop rows
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(radius) || d < float64(b.size) {
						continue
					}
					px, py := b.x+dx, b.y+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fc := pal.Farmland
					if (px+py)%4 < 2 {
						fc = pal.FarmAlt
					}
					fade := (d - float64(b.size)) / float64(radius-b.size)
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(fc, existing, fade*0.5))
				}
			}

		case era >= 4 && era <= 5 && b.category == "housing":
			// Parking lot / pavement around modern buildings
			asphalt := c(45, 45, 48)
			lineC := c(220, 220, 50)
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(radius) || d < float64(b.size) {
						continue
					}
					px, py := b.x+dx, b.y+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fade := (d - float64(b.size)) / float64(radius-b.size)
					pc := asphalt
					if dx%6 == 0 && dy > 0 {
						pc = lineC // parking lines
					}
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(pc, existing, fade*0.6))
				}
			}

		case era == 3 && b.category == "production":
			// Soot-stained ground around factories
			soot := c(40, 38, 35)
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(radius) || d < float64(b.size) {
						continue
					}
					px, py := b.x+dx, b.y+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fade := (d - float64(b.size)) / float64(radius-b.size)
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(soot, existing, fade*0.4+noise2D(px, py, seed+300)*0.3))
				}
			}
		}
	}
}

// ─── Infrastructure ──────────────────────────────────────
func drawInfrastructure(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, cx, cy int, placements []bldInfo) {
	if len(placements) == 0 {
		return
	}

	switch {
	case era == 0:
		// Dirt footpaths — thin, irregular
		for _, b := range placements {
			drawPath(img, cx, cy, b.x, b.y, w, h, pal.Road, 0)
		}

	case era <= 2:
		// Stone/cobble roads
		for _, b := range placements {
			drawRoad(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, dl)
		}

	case era == 3:
		// Railways: dark track + sleeper ties
		trackC := c(60, 55, 50)
		tieC := c(90, 70, 45)
		for _, b := range placements {
			drawRailway(img, cx, cy, b.x, b.y, w, h, trackC, tieC, dl)
		}
		// Also some roads
		for i, b := range placements {
			if i%3 == 0 {
				drawRoad(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, dl)
			}
		}

	case era <= 5:
		// Multi-lane highways with lane markings
		for _, b := range placements {
			drawHighway(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, dl)
		}

	case era == 6:
		// Neon light trails
		for _, b := range placements {
			drawNeonTrail(img, cx, cy, b.x, b.y, w, h, pal.Accent1, pal.Accent2, dl)
		}

	case era >= 7:
		// Energy conduits — glowing blue/gold lines
		for _, b := range placements {
			drawNeonTrail(img, cx, cy, b.x, b.y, w, h, pal.Accent1, pal.Accent2, dl)
		}
	}
}

// ─── Building rendering ──────────────────────────────────
func drawBuildings(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, placements []bldInfo) {
	for _, b := range placements {
		bx, by, size := b.x, b.y, b.size

		// Shadow
		shadowOff := 1 + dl
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				px := bx - size/2 + dx + shadowOff
				py := by - size/2 + dy + shadowOff
				if px >= 0 && px < w && py >= 0 && py < h {
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, c(0, 0, 0), 0.35))
				}
			}
		}

		switch {
		case era <= 1:
			drawBuildingPrimitive(img, w, h, b, pal)
		case era == 2:
			drawBuildingMedieval(img, w, h, b, pal)
		case era == 3:
			drawBuildingIndustrial(img, w, h, b, pal)
		case era <= 5:
			drawBuildingModern(img, w, h, era, dl, b, pal)
		case era == 6:
			drawBuildingCyber(img, w, h, dl, b, pal)
		default:
			drawBuildingSpace(img, w, h, dl, b, pal)
		}

		// Wonder glow (all eras)
		if b.category == "wonder" {
			glowR := size + 3 + dl*3
			for dy := -glowR; dy <= glowR; dy++ {
				for dx := -glowR; dx <= glowR; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d <= float64(size/2) || d >= float64(glowR) {
						continue
					}
					px, py := bx+dx, by+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fade := 1.0 - (d-float64(size/2))/float64(glowR-size/2)
					gc := pal.Accent1
					if era >= 6 {
						gc = pal.Accent2
					}
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, gc, fade*0.25))
				}
			}
		}
	}
}

func drawBuildingPrimitive(img *image.RGBA, w, h int, b bldInfo, pal TerrainPalette) {
	// Round huts with thatched roofs
	bc := c(120, 80, 40)  // wood/mud
	rc := c(100, 90, 30)  // thatch
	if b.category == "wonder" {
		bc = c(200, 180, 100)
		rc = c(220, 200, 60)
	}
	r := b.size / 2
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			d := math.Sqrt(float64(dx*dx + dy*dy))
			if d > float64(r) {
				continue
			}
			px, py := b.x+dx, b.y+dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			if d > float64(r)-1 {
				img.SetRGBA(px, py, lerp(bc, c(0, 0, 0), 0.3))
			} else if dy < 0 {
				img.SetRGBA(px, py, rc) // roof
			} else {
				img.SetRGBA(px, py, bc) // walls
			}
		}
	}
}

func drawBuildingMedieval(img *image.RGBA, w, h int, b bldInfo, pal TerrainPalette) {
	bc := c(140, 130, 110) // stone
	rc := c(100, 45, 25)   // dark wood roof
	if b.category == "wonder" {
		bc = c(220, 210, 180)
		rc = c(180, 160, 60)
	}
	size := b.size
	// Peaked roof (triangle top)
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			px := b.x - size/2 + dx
			py := b.y - size/2 + dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			// Peaked roof: triangle shape in top half
			if dy < size/3 {
				mid := size / 2
				roofSpan := mid - int(float64(dy)*float64(mid)/float64(size/3))
				if dx >= mid-roofSpan && dx <= mid+roofSpan {
					img.SetRGBA(px, py, rc)
				}
			} else if dy == 0 || dy == size-1 || dx == 0 || dx == size-1 {
				img.SetRGBA(px, py, lerp(bc, c(0, 0, 0), 0.35))
			} else {
				img.SetRGBA(px, py, bc)
			}
		}
	}
}

func drawBuildingIndustrial(img *image.RGBA, w, h int, b bldInfo, pal TerrainPalette) {
	bc := c(140, 60, 40)   // brick red
	rc := c(70, 70, 70)    // metal roof
	if b.category == "production" {
		bc = c(110, 100, 90) // factory grey
	}
	if b.category == "wonder" {
		bc = c(200, 180, 140)
		rc = c(160, 150, 130)
	}
	size := b.size
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			px := b.x - size/2 + dx
			py := b.y - size/2 + dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			if dy == 0 || dy == size-1 || dx == 0 || dx == size-1 {
				img.SetRGBA(px, py, lerp(bc, c(0, 0, 0), 0.4))
			} else if dy < size/4 {
				img.SetRGBA(px, py, rc) // flat metal roof
			} else {
				img.SetRGBA(px, py, bc)
				// Brick pattern
				if (dy+dx)%3 == 0 {
					img.SetRGBA(px, py, lerp(bc, c(0, 0, 0), 0.1))
				}
			}
		}
	}
}

func drawBuildingModern(img *image.RGBA, w, h, era, dl int, b bldInfo, pal TerrainPalette) {
	// Tall rectangular buildings with glass facades
	glass := c(120, 170, 220)
	frame := c(60, 65, 70)
	if b.category == "wonder" {
		glass = c(200, 220, 255)
		frame = c(180, 190, 200)
	}
	size := b.size
	// Taller buildings: extend upward
	heightBonus := size / 2
	if b.category == "housing" || b.category == "research" {
		heightBonus = size
	}

	for dy := -heightBonus; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			px := b.x - size/2 + dx
			py := b.y - size/2 + dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			if dx == 0 || dx == size-1 || dy == -heightBonus || dy == size-1 {
				img.SetRGBA(px, py, frame)
			} else {
				// Glass with window grid
				if dx%3 == 0 || dy%3 == 0 {
					img.SetRGBA(px, py, frame) // mullions
				} else {
					// Glass panels — slight color variation
					n := noise2D(px, py, 888)
					img.SetRGBA(px, py, lerp(glass, c(80, 140, 200), n*0.3))
				}
			}
		}
	}
}

func drawBuildingCyber(img *image.RGBA, w, h, dl int, b bldInfo, pal TerrainPalette) {
	// Dark mega-towers with neon outlines
	body := c(15, 15, 22)
	neon1 := pal.Accent1
	neon2 := pal.Accent2
	size := b.size
	heightBonus := size

	for dy := -heightBonus; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			px := b.x - size/2 + dx
			py := b.y - size/2 + dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			isEdge := dx == 0 || dx == size-1 || dy == -heightBonus || dy == size-1
			if isEdge {
				// Neon edge — alternating colors
				nc := neon1
				if (dx+dy)%6 < 3 {
					nc = neon2
				}
				img.SetRGBA(px, py, nc)
			} else {
				img.SetRGBA(px, py, body)
				// Occasional lit windows
				if (dx*3+dy*7)%11 < 2 {
					wc := neon1
					if (dx+dy)%2 == 0 {
						wc = neon2
					}
					img.SetRGBA(px, py, lerp(body, wc, 0.5))
				}
			}
		}
	}
}

func drawBuildingSpace(img *image.RGBA, w, h, dl int, b bldInfo, pal TerrainPalette) {
	// Domed structures with energy glow
	center := pal.Accent1
	dome := c(40, 50, 80)
	if b.category == "wonder" {
		center = pal.Accent2
		dome = c(60, 70, 110)
	}
	r := b.size / 2
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			d := math.Sqrt(float64(dx*dx + dy*dy))
			if d > float64(r) {
				continue
			}
			px, py := b.x+dx, b.y+dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			t := d / float64(r)
			if t > 0.85 {
				// Dome rim — glowing
				img.SetRGBA(px, py, lerp(center, dome, 0.3))
			} else {
				// Dome body with highlight at top
				highlight := 1.0 - float64(dy+r)/float64(r*2)
				img.SetRGBA(px, py, lerp(dome, center, highlight*0.4+t*0.2))
			}
		}
	}
}

// ─── Infrastructure drawing helpers ──────────────────────

func drawPath(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		if px >= 0 && px < w && py >= 0 && py < h {
			existing := img.RGBAAt(px, py)
			img.SetRGBA(px, py, lerp(existing, roadC, 0.5))
		}
	}
}

func drawRoad(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC, edgeC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	roadW := 1 + dl
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		for rw := -roadW; rw <= roadW; rw++ {
			rpx := px + int(float64(rw)*(-dy/dist))
			rpy := py + int(float64(rw)*(dx/dist))
			if rpx < 0 || rpx >= w || rpy < 0 || rpy >= h {
				continue
			}
			if rw == -roadW || rw == roadW {
				existing := img.RGBAAt(rpx, rpy)
				img.SetRGBA(rpx, rpy, lerp(existing, edgeC, 0.5))
			} else {
				img.SetRGBA(rpx, rpy, roadC)
			}
		}
	}
}

func drawRailway(img *image.RGBA, x0, y0, x1, y1, w, h int, trackC, tieC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		if px < 0 || px >= w || py < 0 || py >= h {
			continue
		}
		// Rails
		img.SetRGBA(px, py, trackC)
		// Sleeper ties every few pixels perpendicular
		if i%4 == 0 {
			for rw := -2; rw <= 2; rw++ {
				rpx := px + int(float64(rw)*(-dy/dist))
				rpy := py + int(float64(rw)*(dx/dist))
				if rpx >= 0 && rpx < w && rpy >= 0 && rpy < h {
					img.SetRGBA(rpx, rpy, tieC)
				}
			}
		}
	}
}

func drawHighway(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC, edgeC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	roadW := 2 + dl
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		for rw := -roadW; rw <= roadW; rw++ {
			rpx := px + int(float64(rw)*(-dy/dist))
			rpy := py + int(float64(rw)*(dx/dist))
			if rpx < 0 || rpx >= w || rpy < 0 || rpy >= h {
				continue
			}
			if rw == -roadW || rw == roadW {
				img.SetRGBA(rpx, rpy, edgeC)
			} else if rw == 0 && i%6 < 3 {
				// Dashed center line
				img.SetRGBA(rpx, rpy, c(220, 220, 50))
			} else {
				img.SetRGBA(rpx, rpy, roadC)
			}
		}
	}
}

func drawNeonTrail(img *image.RGBA, x0, y0, x1, y1, w, h int, neon1, neon2 color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		if px < 0 || px >= w || py < 0 || py >= h {
			continue
		}
		// Core bright line
		nc := neon1
		if i%8 < 4 {
			nc = neon2
		}
		img.SetRGBA(px, py, nc)
		// Glow around the line
		for rw := -1; rw <= 1; rw++ {
			rpx := px + int(float64(rw)*(-dy/dist))
			rpy := py + int(float64(rw)*(dx/dist))
			if rpx >= 0 && rpx < w && rpy >= 0 && rpy < h && rw != 0 {
				existing := img.RGBAAt(rpx, rpy)
				img.SetRGBA(rpx, rpy, lerp(existing, nc, 0.25))
			}
		}
	}
}

// ─── Decorations ─────────────────────────────────────────
func drawDecorations(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, seed uint64, placements []bldInfo) {
	switch {
	case era == 3:
		// Smokestacks above production buildings
		smokeC := c(100, 100, 105)
		for _, b := range placements {
			if b.category != "production" {
				continue
			}
			// Chimney
			chimneyX := b.x + b.size/4
			for cy := b.y - b.size/2 - 4; cy < b.y-b.size/2; cy++ {
				if cy >= 0 && cy < h && chimneyX >= 0 && chimneyX < w {
					img.SetRGBA(chimneyX, cy, c(80, 50, 35))
					if chimneyX+1 < w {
						img.SetRGBA(chimneyX+1, cy, c(80, 50, 35))
					}
				}
			}
			// Smoke puff
			for dy := -3; dy <= 0; dy++ {
				for dx := -2; dx <= 2; dx++ {
					px, py := chimneyX+dx, b.y-b.size/2-5+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						existing := img.RGBAAt(px, py)
						img.SetRGBA(px, py, lerp(existing, smokeC, 0.3))
					}
				}
			}
		}

	case era == 4 || era == 5:
		// Power lines between some buildings
		lineC := c(60, 60, 65)
		for i := 0; i+1 < len(placements); i += 2 {
			a, b := placements[i], placements[i+1]
			drawPowerLine(img, a.x, a.y-a.size/2, b.x, b.y-b.size/2, w, h, lineC)
		}

	case era == 6:
		// Holographic billboards above some buildings
		for _, b := range placements {
			if b.category != "housing" && b.category != "production" {
				continue
			}
			bHash := hashKey(b.key + "holo")
			if bHash%3 != 0 {
				continue
			}
			holoW, holoH := 4+dl*2, 2+dl
			holoC := pal.Accent1
			if bHash%2 == 0 {
				holoC = pal.Accent2
			}
			for dy := 0; dy < holoH; dy++ {
				for dx := 0; dx < holoW; dx++ {
					px := b.x - holoW/2 + dx
					py := b.y - b.size - 2 + dy
					if px >= 0 && px < w && py >= 0 && py < h {
						existing := img.RGBAAt(px, py)
						img.SetRGBA(px, py, lerp(existing, holoC, 0.6))
					}
				}
			}
		}
	}
}

func drawPowerLine(img *image.RGBA, x0, y0, x1, y1, w, h int, lineC color.RGBA) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		// Slight sag (catenary)
		sag := math.Sin(t*math.Pi) * 3.0
		py := y0 + int(dy*t+sag)
		if px >= 0 && px < w && py >= 0 && py < h {
			img.SetRGBA(px, py, lineC)
		}
	}
}

func settlementLabel(buildingCount int) string {
	switch {
	case buildingCount <= 5:
		return "Village"
	case buildingCount <= 20:
		return "Town"
	case buildingCount <= 50:
		return "City"
	default:
		return "Metropolis"
	}
}
