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
	Ground     color.RGBA
	GroundAlt  color.RGBA
	Water      color.RGBA
	WaterLight color.RGBA
	WaterDeep  color.RGBA
	Tree       color.RGBA
	TreeDark   color.RGBA
	TreeLight  color.RGBA
	Road       color.RGBA
	RoadEdge   color.RGBA
	Hill       color.RGBA
	HillLight  color.RGBA
	Farmland   color.RGBA
	FarmAlt    color.RGBA
}

// MapGenConfig holds parameters for map generation
type MapGenConfig struct {
	Width       int // pixels
	Height      int // pixels
	DetailLevel int // 0=mini, 1=full
	Buildings   map[string]game.BuildingState
	AgeKey      string
}

// eraFromAge returns the era index (0-8) from an age key
func eraFromAge(ageKey string) int {
	ages := config.AgeByKey()
	if a, ok := ages[ageKey]; ok {
		order := a.Order
		switch {
		case order <= 1:
			return 0
		case order <= 4:
			return 1
		case order <= 7:
			return 2
		case order <= 10:
			return 3
		case order <= 13:
			return 4
		case order <= 15:
			return 5
		case order == 16:
			return 6
		case order <= 19:
			return 7
		default:
			return 8
		}
	}
	return 0
}

func getTerrainPalette(era int) TerrainPalette {
	switch era {
	case 0: // primitive
		return TerrainPalette{
			Ground: c(34, 85, 34), GroundAlt: c(44, 100, 42),
			Water: c(28, 55, 150), WaterLight: c(50, 80, 175), WaterDeep: c(18, 40, 120),
			Tree: c(20, 110, 30), TreeDark: c(12, 75, 18), TreeLight: c(35, 130, 45),
			Road: c(85, 65, 42), RoadEdge: c(65, 50, 32),
			Hill: c(50, 105, 50), HillLight: c(65, 120, 62),
			Farmland: c(90, 120, 40), FarmAlt: c(80, 110, 35),
		}
	case 1: // ancient
		return TerrainPalette{
			Ground: c(120, 100, 60), GroundAlt: c(135, 115, 72),
			Water: c(30, 68, 140), WaterLight: c(45, 85, 162), WaterDeep: c(20, 50, 110),
			Tree: c(50, 100, 30), TreeDark: c(38, 78, 22), TreeLight: c(65, 118, 42),
			Road: c(145, 125, 82), RoadEdge: c(115, 98, 65),
			Hill: c(140, 120, 80), HillLight: c(160, 140, 95),
			Farmland: c(100, 120, 45), FarmAlt: c(110, 130, 50),
		}
	case 2: // medieval
		return TerrainPalette{
			Ground: c(30, 60, 25), GroundAlt: c(38, 72, 32),
			Water: c(20, 48, 128), WaterLight: c(32, 65, 148), WaterDeep: c(12, 35, 100),
			Tree: c(15, 80, 20), TreeDark: c(8, 52, 12), TreeLight: c(25, 98, 30),
			Road: c(105, 95, 72), RoadEdge: c(80, 72, 55),
			Hill: c(45, 75, 40), HillLight: c(58, 90, 52),
			Farmland: c(70, 95, 30), FarmAlt: c(80, 105, 35),
		}
	case 3: // industrial
		return TerrainPalette{
			Ground: c(68, 68, 62), GroundAlt: c(78, 78, 72),
			Water: c(38, 58, 88), WaterLight: c(50, 70, 102), WaterDeep: c(28, 42, 68),
			Tree: c(48, 78, 38), TreeDark: c(35, 58, 28), TreeLight: c(58, 90, 48),
			Road: c(115, 112, 105), RoadEdge: c(90, 88, 82),
			Hill: c(85, 85, 78), HillLight: c(98, 98, 90),
			Farmland: c(75, 85, 55), FarmAlt: c(82, 92, 60),
		}
	case 4: // modern
		return TerrainPalette{
			Ground: c(52, 58, 62), GroundAlt: c(62, 68, 72),
			Water: c(28, 78, 148), WaterLight: c(42, 95, 168), WaterDeep: c(18, 58, 118),
			Tree: c(38, 88, 38), TreeDark: c(28, 68, 28), TreeLight: c(50, 105, 50),
			Road: c(125, 125, 125), RoadEdge: c(100, 100, 100),
			Hill: c(72, 78, 82), HillLight: c(85, 90, 95),
			Farmland: c(60, 80, 45), FarmAlt: c(68, 88, 52),
		}
	case 5: // digital
		return TerrainPalette{
			Ground: c(10, 15, 38), GroundAlt: c(15, 22, 48),
			Water: c(18, 38, 118), WaterLight: c(28, 52, 138), WaterDeep: c(10, 25, 88),
			Tree: c(8, 28, 58), TreeDark: c(5, 18, 42), TreeLight: c(12, 38, 72),
			Road: c(0, 115, 175), RoadEdge: c(0, 85, 135),
			Hill: c(18, 25, 55), HillLight: c(25, 35, 68),
			Farmland: c(15, 35, 60), FarmAlt: c(18, 42, 72),
		}
	case 6: // cyber
		return TerrainPalette{
			Ground: c(8, 8, 12), GroundAlt: c(14, 12, 20),
			Water: c(78, 0, 118), WaterLight: c(100, 0, 148), WaterDeep: c(55, 0, 85),
			Tree: c(0, 38, 18), TreeDark: c(0, 22, 10), TreeLight: c(0, 55, 28),
			Road: c(250, 0, 125), RoadEdge: c(180, 0, 90),
			Hill: c(15, 12, 22), HillLight: c(22, 18, 32),
			Farmland: c(0, 50, 25), FarmAlt: c(0, 60, 30),
		}
	case 7: // space
		return TerrainPalette{
			Ground: c(5, 5, 15), GroundAlt: c(8, 8, 22),
			Water: c(15, 25, 78), WaterLight: c(22, 35, 98), WaterDeep: c(8, 15, 55),
			Tree: c(8, 8, 28), TreeDark: c(5, 5, 18), TreeLight: c(12, 12, 38),
			Road: c(58, 78, 138), RoadEdge: c(42, 58, 105),
			Hill: c(10, 10, 28), HillLight: c(15, 15, 38),
			Farmland: c(10, 18, 35), FarmAlt: c(12, 22, 42),
		}
	case 8: // cosmic
		return TerrainPalette{
			Ground: c(3, 3, 8), GroundAlt: c(8, 5, 15),
			Water: c(38, 10, 78), WaterLight: c(58, 15, 98), WaterDeep: c(25, 5, 55),
			Tree: c(5, 5, 15), TreeDark: c(3, 3, 10), TreeLight: c(8, 8, 22),
			Road: c(78, 58, 18), RoadEdge: c(58, 42, 12),
			Hill: c(8, 5, 18), HillLight: c(12, 8, 25),
			Farmland: c(10, 8, 20), FarmAlt: c(12, 10, 25),
		}
	}
	return getTerrainPalette(0)
}

func c(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

// buildingColor returns the base color for a building category
func buildingColor(category string) color.RGBA {
	switch category {
	case "housing":
		return c(160, 82, 45)  // warm brown
	case "production":
		return c(180, 160, 50) // industrial yellow
	case "research":
		return c(50, 160, 200) // cyan
	case "military":
		return c(180, 50, 50)  // red
	case "storage":
		return c(60, 110, 180) // blue
	case "wonder":
		return c(255, 215, 0)  // gold
	default:
		return c(160, 160, 160)
	}
}

// roofColor returns a darker roof/top color for a building
func roofColor(category string) color.RGBA {
	switch category {
	case "housing":
		return c(120, 55, 28)
	case "production":
		return c(100, 90, 30)
	case "research":
		return c(30, 100, 140)
	case "military":
		return c(130, 30, 30)
	case "storage":
		return c(40, 75, 130)
	case "wonder":
		return c(200, 170, 0)
	default:
		return c(110, 110, 110)
	}
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

// noise2D generates layered noise for more natural terrain
func noise2D(x, y int, seed uint64) float64 {
	// Two octaves of hash noise for more organic look
	h1 := hashKey(string(rune(seed)) + string(rune(x*7919+y*6271)))
	h2 := hashKey(string(rune(seed+77)) + string(rune((x/3)*4909+(y/3)*3571)))
	n1 := float64(h1%10000) / 10000.0
	n2 := float64(h2%10000) / 10000.0
	return n1*0.6 + n2*0.4
}

// buildingPos stores a placed building's position for path drawing
type buildingPos struct {
	x, y int
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

	// --- 1. Terrain with elevation ---
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := noise2D(x, y, seed)
			// Elevation: gentle hills using lower-frequency noise
			elev := noise2D(x/4, y/4, seed+200)
			if elev > 0.65 {
				// Hill terrain
				hillT := (elev - 0.65) / 0.35
				base := lerp(pal.Ground, pal.GroundAlt, n)
				img.SetRGBA(x, y, lerp(base, lerp(pal.Hill, pal.HillLight, n), hillT))
			} else {
				img.SetRGBA(x, y, lerp(pal.Ground, pal.GroundAlt, n))
			}
		}
	}

	// --- 2. Stars in space eras ---
	if era >= 7 {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				n := noise2D(x, y, seed+99)
				if n < 0.025 {
					brightness := uint8(100 + int(n*6000))
					if era == 8 {
						img.SetRGBA(x, y, c(brightness, uint8(float64(brightness)*0.82), uint8(float64(brightness)*0.35)))
					} else {
						img.SetRGBA(x, y, c(brightness, brightness, uint8(min(255, int(float64(brightness)*1.15)))))
					}
				}
			}
		}
	}

	// --- 3. River with banks and depth ---
	riverBaseX := float64(w) * 0.28
	riverWidth := 3
	if cfg.DetailLevel > 0 {
		riverWidth = 6
	}
	bankWidth := 1
	if cfg.DetailLevel > 0 {
		bankWidth = 2
	}
	for y := 0; y < h; y++ {
		// Two sine waves for more natural meander
		rx := riverBaseX +
			math.Sin(float64(y)*0.06)*float64(w)*0.10 +
			math.Sin(float64(y)*0.15)*float64(w)*0.03
		for dx := -bankWidth; dx < riverWidth+bankWidth; dx++ {
			px := int(rx) + dx
			if px < 0 || px >= w {
				continue
			}
			if dx < 0 || dx >= riverWidth {
				// River bank — blend ground toward water
				bankT := 0.3
				existing := img.RGBAAt(px, y)
				img.SetRGBA(px, y, lerp(existing, pal.Water, bankT))
			} else {
				// River body — center is deeper
				centerDist := math.Abs(float64(dx)-float64(riverWidth)/2.0) / (float64(riverWidth) / 2.0)
				wc := lerp(pal.WaterDeep, pal.WaterLight, centerDist)
				// Slight shimmer
				shimmer := noise2D(px, y, seed+7)
				wc = lerp(wc, pal.Water, shimmer*0.2)
				img.SetRGBA(px, y, wc)
			}
		}
	}

	// --- 4. Trees as organic clusters ---
	treeDensity := 0.08
	if era >= 3 {
		treeDensity = 0.03
	}
	if era >= 5 {
		treeDensity = 0.01
	}
	treeRadius := 2
	if cfg.DetailLevel > 0 {
		treeRadius = 3
	}
	for y := treeRadius; y < h-treeRadius; y += treeRadius + 1 {
		for x := treeRadius; x < w-treeRadius; x += treeRadius + 1 {
			n := noise2D(x, y, seed+42)
			if n >= treeDensity {
				continue
			}
			// Avoid planting trees in water
			existing := img.RGBAAt(x, y)
			if existing == pal.Water || existing == pal.WaterLight || existing == pal.WaterDeep {
				continue
			}
			// Draw circular tree canopy
			for dy := -treeRadius; dy <= treeRadius; dy++ {
				for dx := -treeRadius; dx <= treeRadius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(treeRadius) {
						continue
					}
					tx, ty := x+dx, y+dy
					if tx < 0 || tx >= w || ty < 0 || ty >= h {
						continue
					}
					// Canopy shading: lighter at top, darker at bottom
					shade := float64(dy+treeRadius) / float64(treeRadius*2)
					tc := lerp(pal.TreeLight, pal.TreeDark, shade)
					// Edge fade
					edgeFade := d / float64(treeRadius)
					existing := img.RGBAAt(tx, ty)
					img.SetRGBA(tx, ty, lerp(existing, tc, 1.0-edgeFade*0.4))
				}
			}
			// Tree trunk (tiny dark center pixel)
			if x >= 0 && x < w && y >= 0 && y < h {
				img.SetRGBA(x, y, pal.TreeDark)
			}
		}
	}

	// --- 5. Collect building positions for paths/farmland ---
	type bldInfo struct {
		key      string
		category string
		x, y     int
		size     int
	}
	var placements []bldInfo

	maxDist := float64(min(w, h)) / 2.0

	for key, bs := range cfg.Buildings {
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

			size := 3
			if bs.Category == "wonder" {
				size = 6
			} else if bs.Category == "military" {
				size = 4
			}
			if cfg.DetailLevel > 0 {
				size += 2
			}

			placements = append(placements, bldInfo{key, bs.Category, bx, by, size})
		}
	}

	// Sort by distance from center so closer buildings draw on top
	sort.Slice(placements, func(i, j int) bool {
		di := (placements[i].x-cx)*(placements[i].x-cx) + (placements[i].y-cy)*(placements[i].y-cy)
		dj := (placements[j].x-cx)*(placements[j].x-cx) + (placements[j].y-cy)*(placements[j].y-cy)
		return di > dj // furthest first so close buildings draw on top
	})

	// --- 6. Farmland patches around production buildings ---
	for _, b := range placements {
		if b.category != "production" {
			continue
		}
		farmRadius := b.size + 4
		if cfg.DetailLevel > 0 {
			farmRadius += 3
		}
		for dy := -farmRadius; dy <= farmRadius; dy++ {
			for dx := -farmRadius; dx <= farmRadius; dx++ {
				d := math.Sqrt(float64(dx*dx + dy*dy))
				if d > float64(farmRadius) || d < float64(b.size) {
					continue
				}
				px, py := b.x+dx, b.y+dy
				if px < 0 || px >= w || py < 0 || py >= h {
					continue
				}
				// Striped farmland pattern
				n := noise2D(px, py, seed+300)
				fc := pal.Farmland
				if (px+py)%4 < 2 {
					fc = pal.FarmAlt
				}
				fade := (d - float64(b.size)) / float64(farmRadius-b.size)
				existing := img.RGBAAt(px, py)
				img.SetRGBA(px, py, lerp(fc, existing, fade*0.5+n*0.2))
			}
		}
	}

	// --- 7. Roads connecting buildings to center ---
	if len(placements) > 0 && era >= 1 {
		for _, b := range placements {
			drawRoad(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, cfg.DetailLevel)
		}
	}

	// --- 8. Draw buildings ---
	for _, b := range placements {
		bc := buildingColor(b.category)
		rc := roofColor(b.category)
		bx, by, size := b.x, b.y, b.size

		// Draw shadow first (offset down-right)
		shadowOff := 1
		if cfg.DetailLevel > 0 {
			shadowOff = 2
		}
		shadowC := c(0, 0, 0)
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				px := bx - size/2 + dx + shadowOff
				py := by - size/2 + dy + shadowOff
				if px >= 0 && px < w && py >= 0 && py < h {
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, shadowC, 0.35))
				}
			}
		}

		// Building body
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				px := bx - size/2 + dx
				py := by - size/2 + dy
				if px < 0 || px >= w || py < 0 || py >= h {
					continue
				}
				// Border
				if dy == 0 || dy == size-1 || dx == 0 || dx == size-1 {
					img.SetRGBA(px, py, lerp(bc, c(0, 0, 0), 0.4))
					continue
				}
				// Roof area (top third)
				if dy < size/3 {
					img.SetRGBA(px, py, rc)
				} else {
					// Wall with slight window detail
					if cfg.DetailLevel > 0 && dx > 1 && dx < size-2 && dy > size/3+1 && (dx+dy)%3 == 0 {
						img.SetRGBA(px, py, lerp(bc, c(200, 220, 255), 0.5)) // window
					} else {
						img.SetRGBA(px, py, bc)
					}
				}
			}
		}

		// Wonder: radial glow
		if b.category == "wonder" {
			glowR := size + 3
			if cfg.DetailLevel > 0 {
				glowR += 3
			}
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
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, c(255, 215, 0), fade*0.25))
				}
			}
		}

		// Military: small flag/banner on top
		if b.category == "military" && by-size/2-2 >= 0 {
			flagX := bx
			for fy := by - size/2 - 3; fy < by-size/2; fy++ {
				if fy >= 0 && fy < h && flagX >= 0 && flagX < w {
					img.SetRGBA(flagX, fy, c(200, 30, 30))
				}
			}
		}
	}

	return img
}

// drawRoad draws a road line between two points with anti-aliased edges
func drawRoad(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC, edgeC color.RGBA, detail int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	roadW := 1
	if detail > 0 {
		roadW = 2
	}
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		for rw := -roadW; rw <= roadW; rw++ {
			// Road runs perpendicular to direction
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

// settlementLabel returns "Village", "Town", "City", or "Metropolis"
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
