package ui

import (
	"hash/fnv"
	"image"
	"image/color"
	"math"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// TerrainPalette defines the color palette for a terrain era
type TerrainPalette struct {
	Ground     color.RGBA
	GroundAlt  color.RGBA
	Water      color.RGBA
	WaterLight color.RGBA
	Tree       color.RGBA
	TreeDark   color.RGBA
	Road       color.RGBA
	Extra      color.RGBA
}

// MapGenConfig holds parameters for map generation
type MapGenConfig struct {
	Width       int // pixels
	Height      int // pixels
	DetailLevel int // 0=mini, 1=full
	Buildings   map[string]game.BuildingState
	AgeKey      string
	Tick        int
	TotalPop    int
}

// eraFromAge returns the era index (0-8) from an age key
func eraFromAge(ageKey string) int {
	ages := config.AgeByKey()
	if a, ok := ages[ageKey]; ok {
		order := a.Order
		switch {
		case order <= 1:
			return 0 // primitive
		case order <= 4:
			return 1 // ancient
		case order <= 7:
			return 2 // medieval
		case order <= 10:
			return 3 // industrial
		case order <= 13:
			return 4 // modern
		case order <= 15:
			return 5 // digital
		case order == 16:
			return 6 // cyber
		case order <= 19:
			return 7 // space
		default:
			return 8 // cosmic
		}
	}
	return 0
}

// getTerrainPalette returns the terrain colors for an era
func getTerrainPalette(era int) TerrainPalette {
	switch era {
	case 0: // primitive — lush greens
		return TerrainPalette{
			Ground:     color.RGBA{34, 85, 34, 255},
			GroundAlt:  color.RGBA{40, 95, 40, 255},
			Water:      color.RGBA{30, 60, 160, 255},
			WaterLight: color.RGBA{50, 80, 180, 255},
			Tree:       color.RGBA{20, 110, 30, 255},
			TreeDark:   color.RGBA{15, 80, 20, 255},
			Road:       color.RGBA{80, 60, 40, 255},
			Extra:      color.RGBA{60, 100, 30, 255},
		}
	case 1: // ancient — sandy warm
		return TerrainPalette{
			Ground:     color.RGBA{120, 100, 60, 255},
			GroundAlt:  color.RGBA{130, 110, 70, 255},
			Water:      color.RGBA{30, 70, 140, 255},
			WaterLight: color.RGBA{40, 85, 160, 255},
			Tree:       color.RGBA{50, 100, 30, 255},
			TreeDark:   color.RGBA{40, 80, 25, 255},
			Road:       color.RGBA{140, 120, 80, 255},
			Extra:      color.RGBA{100, 90, 50, 255},
		}
	case 2: // medieval — dark forest
		return TerrainPalette{
			Ground:     color.RGBA{30, 60, 25, 255},
			GroundAlt:  color.RGBA{35, 70, 30, 255},
			Water:      color.RGBA{20, 50, 130, 255},
			WaterLight: color.RGBA{30, 65, 150, 255},
			Tree:       color.RGBA{15, 80, 20, 255},
			TreeDark:   color.RGBA{10, 55, 15, 255},
			Road:       color.RGBA{100, 90, 70, 255},
			Extra:      color.RGBA{50, 50, 45, 255},
		}
	case 3: // industrial — grey smog
		return TerrainPalette{
			Ground:     color.RGBA{70, 70, 65, 255},
			GroundAlt:  color.RGBA{80, 80, 75, 255},
			Water:      color.RGBA{40, 60, 90, 255},
			WaterLight: color.RGBA{50, 70, 100, 255},
			Tree:       color.RGBA{50, 80, 40, 255},
			TreeDark:   color.RGBA{40, 60, 30, 255},
			Road:       color.RGBA{110, 110, 100, 255},
			Extra:      color.RGBA{90, 85, 80, 255},
		}
	case 4: // modern — urban grey-blue
		return TerrainPalette{
			Ground:     color.RGBA{55, 60, 65, 255},
			GroundAlt:  color.RGBA{65, 70, 75, 255},
			Water:      color.RGBA{30, 80, 150, 255},
			WaterLight: color.RGBA{40, 95, 170, 255},
			Tree:       color.RGBA{40, 90, 40, 255},
			TreeDark:   color.RGBA{30, 70, 30, 255},
			Road:       color.RGBA{120, 120, 120, 255},
			Extra:      color.RGBA{80, 85, 90, 255},
		}
	case 5: // digital — dark blue
		return TerrainPalette{
			Ground:     color.RGBA{10, 15, 40, 255},
			GroundAlt:  color.RGBA{15, 20, 50, 255},
			Water:      color.RGBA{20, 40, 120, 255},
			WaterLight: color.RGBA{30, 55, 140, 255},
			Tree:       color.RGBA{10, 30, 60, 255},
			TreeDark:   color.RGBA{8, 20, 45, 255},
			Road:       color.RGBA{0, 120, 180, 255},
			Extra:      color.RGBA{20, 60, 120, 255},
		}
	case 6: // cyber — neon on black
		return TerrainPalette{
			Ground:     color.RGBA{8, 8, 12, 255},
			GroundAlt:  color.RGBA{12, 12, 18, 255},
			Water:      color.RGBA{80, 0, 120, 255},
			WaterLight: color.RGBA{100, 0, 150, 255},
			Tree:       color.RGBA{0, 40, 20, 255},
			TreeDark:   color.RGBA{0, 25, 12, 255},
			Road:       color.RGBA{255, 0, 128, 255},
			Extra:      color.RGBA{0, 255, 80, 255},
		}
	case 7: // space — starfield
		return TerrainPalette{
			Ground:     color.RGBA{5, 5, 15, 255},
			GroundAlt:  color.RGBA{8, 8, 20, 255},
			Water:      color.RGBA{15, 25, 80, 255},
			WaterLight: color.RGBA{20, 35, 100, 255},
			Tree:       color.RGBA{10, 10, 30, 255},
			TreeDark:   color.RGBA{8, 8, 22, 255},
			Road:       color.RGBA{60, 80, 140, 255},
			Extra:      color.RGBA{40, 40, 80, 255},
		}
	case 8: // cosmic — deep void with gold
		return TerrainPalette{
			Ground:     color.RGBA{3, 3, 8, 255},
			GroundAlt:  color.RGBA{6, 4, 12, 255},
			Water:      color.RGBA{40, 10, 80, 255},
			WaterLight: color.RGBA{60, 15, 100, 255},
			Tree:       color.RGBA{5, 5, 15, 255},
			TreeDark:   color.RGBA{3, 3, 10, 255},
			Road:       color.RGBA{80, 60, 20, 255},
			Extra:      color.RGBA{180, 150, 50, 255},
		}
	}
	return getTerrainPalette(0)
}

// buildingColor returns the color for a building category
func buildingColor(category string) color.RGBA {
	switch category {
	case "housing":
		return color.RGBA{80, 200, 80, 255}
	case "production":
		return color.RGBA{220, 200, 50, 255}
	case "research":
		return color.RGBA{50, 200, 220, 255}
	case "military":
		return color.RGBA{220, 60, 60, 255}
	case "storage":
		return color.RGBA{60, 120, 220, 255}
	case "wonder":
		return color.RGBA{255, 215, 0, 255}
	default:
		return color.RGBA{200, 200, 200, 255}
	}
}

// hashKey returns a deterministic hash for a string key
func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

// lerp linearly interpolates between two colors
func lerp(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

// noise generates a simple deterministic noise value 0.0-1.0 for a coordinate
func noise(x, y int, seed uint64) float64 {
	h := hashKey(string(rune(seed)) + string(rune(x*7919+y*6271)))
	return float64(h%10000) / 10000.0
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

	// 1. Fill terrain with noise-varied ground
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := noise(x, y, seed)
			c := lerp(pal.Ground, pal.GroundAlt, n)
			img.SetRGBA(x, y, c)
		}
	}

	// 2. Scatter stars/specks in space eras
	if era >= 7 {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				n := noise(x, y, seed+99)
				if n < 0.03 {
					brightness := uint8(120 + int(n*4000))
					if era == 8 {
						// Cosmic: gold-tinted stars
						img.SetRGBA(x, y, color.RGBA{brightness, uint8(float64(brightness) * 0.85), uint8(float64(brightness) * 0.4), 255})
					} else {
						img.SetRGBA(x, y, color.RGBA{brightness, brightness, uint8(float64(brightness) * 1.1), 255})
					}
				}
			}
		}
	}

	// 3. Draw river (sinusoidal, width varies by detail)
	riverBaseX := float64(w) * 0.3
	riverWidth := 2
	if cfg.DetailLevel > 0 {
		riverWidth = 4
	}
	for y := 0; y < h; y++ {
		rx := riverBaseX + math.Sin(float64(y)*0.08)*float64(w)*0.12
		for dx := 0; dx < riverWidth; dx++ {
			px := int(rx) + dx
			if px >= 0 && px < w {
				// River edge gradient
				t := float64(dx) / float64(riverWidth)
				edgeFade := 1.0 - math.Abs(t-0.5)*2.0
				c := lerp(pal.Water, pal.WaterLight, edgeFade*0.5+noise(px, y, seed+7)*0.3)
				img.SetRGBA(px, y, c)
			}
		}
	}

	// 4. Scatter trees/vegetation
	treeDensity := 0.12
	if era >= 3 {
		treeDensity = 0.04
	}
	if era >= 5 {
		treeDensity = 0.015
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := noise(x, y, seed+42)
			if n < treeDensity {
				// Tree: 2x2 or 3x3 blob depending on detail
				treeSize := 2
				if cfg.DetailLevel > 0 {
					treeSize = 3
				}
				tc := lerp(pal.Tree, pal.TreeDark, noise(x, y, seed+43))
				for dy := 0; dy < treeSize; dy++ {
					for dx := 0; dx < treeSize; dx++ {
						tx, ty := x+dx, y+dy
						if tx < w && ty < h {
							img.SetRGBA(tx, ty, tc)
						}
					}
				}
			}
		}
	}

	// 5. Draw roads from center outward (in appropriate eras)
	cx, cy := w/2, h/2
	if era >= 2 && era <= 6 {
		roadAngles := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
		roadLen := float64(min(w, h)) * 0.35
		for _, angle := range roadAngles {
			for d := 3.0; d < roadLen; d += 1.0 {
				rx := cx + int(math.Cos(angle)*d)
				ry := cy + int(math.Sin(angle)*d)
				if rx >= 0 && rx < w && ry >= 0 && ry < h {
					img.SetRGBA(rx, ry, pal.Road)
					// Road width
					if cfg.DetailLevel > 0 {
						if rx+1 < w {
							img.SetRGBA(rx+1, ry, pal.Road)
						}
						if ry+1 < h {
							img.SetRGBA(rx, ry+1, pal.Road)
						}
					}
				}
			}
		}
	}

	// 6. Place buildings radially from center
	for key, bs := range cfg.Buildings {
		if !bs.Unlocked || bs.Count == 0 {
			continue
		}
		bc := buildingColor(bs.Category)

		// Ring distances by category
		ringMin, ringMax := 0.08, 0.35
		switch bs.Category {
		case "housing":
			ringMin, ringMax = 0.05, 0.20
		case "production":
			ringMin, ringMax = 0.10, 0.30
		case "military":
			ringMin, ringMax = 0.20, 0.40
		case "wonder":
			ringMin, ringMax = 0.03, 0.18
		case "research":
			ringMin, ringMax = 0.08, 0.25
		}

		maxDist := float64(min(w, h)) / 2.0

		for i := 0; i < bs.Count; i++ {
			bHash := hashKey(key + string(rune(i)))
			angle := float64(bHash%3600) / 3600.0 * 2.0 * math.Pi
			distRatio := ringMin + float64(bHash%1000)/1000.0*(ringMax-ringMin)
			dist := distRatio * maxDist
			bx := cx + int(math.Cos(angle)*dist)
			by := cy + int(math.Sin(angle)*dist*0.7)

			// Building size based on category and detail
			size := 3
			if bs.Category == "wonder" {
				size = 5
			}
			if cfg.DetailLevel > 0 {
				size += 2
			}

			// Draw building as a filled rectangle with border
			borderC := color.RGBA{
				R: uint8(float64(bc.R) * 0.5),
				G: uint8(float64(bc.G) * 0.5),
				B: uint8(float64(bc.B) * 0.5),
				A: 255,
			}
			for dy := -size / 2; dy <= size/2; dy++ {
				for dx := -size / 2; dx <= size/2; dx++ {
					px, py := bx+dx, by+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						if dy == -size/2 || dy == size/2 || dx == -size/2 || dx == size/2 {
							img.SetRGBA(px, py, borderC)
						} else {
							img.SetRGBA(px, py, bc)
						}
					}
				}
			}

			// Wonder gets a glow effect
			if bs.Category == "wonder" && cfg.DetailLevel > 0 {
				glowSize := size + 2
				for dy := -glowSize; dy <= glowSize; dy++ {
					for dx := -glowSize; dx <= glowSize; dx++ {
						px, py := bx+dx, by+dy
						if px >= 0 && px < w && py >= 0 && py < h {
							d := math.Sqrt(float64(dx*dx + dy*dy))
							if d > float64(size/2) && d < float64(glowSize) {
								fade := 1.0 - (d-float64(size/2))/float64(glowSize-size/2)
								existing := img.RGBAAt(px, py)
								glowed := lerp(existing, color.RGBA{255, 215, 0, 255}, fade*0.3)
								img.SetRGBA(px, py, glowed)
							}
						}
					}
				}
			}
		}
	}

	// 7. Place villager dots (animated by tick)
	villagerCount := cfg.TotalPop
	if villagerCount > 15 {
		villagerCount = 15
	}
	for i := 0; i < villagerCount; i++ {
		vHash := hashKey(string(rune(cfg.Tick*7 + i*31)))
		radius := float64(min(w, h))*0.1 + float64(vHash%100)/100.0*float64(min(w, h))*0.2
		angle := float64(vHash%3600) / 3600.0 * 2.0 * math.Pi
		vx := cx + int(math.Cos(angle)*radius)
		vy := cy + int(math.Sin(angle)*radius*0.7)
		if vx >= 1 && vx < w-1 && vy >= 1 && vy < h-1 {
			// Small white dot with slight color
			img.SetRGBA(vx, vy, color.RGBA{255, 255, 255, 255})
			img.SetRGBA(vx+1, vy, color.RGBA{220, 220, 240, 255})
			img.SetRGBA(vx, vy+1, color.RGBA{220, 220, 240, 255})
		}
	}

	// 8. Settlement center marker
	markerC := color.RGBA{255, 215, 0, 255}
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			px, py := cx+dx, cy+dy
			if px >= 0 && px < w && py >= 0 && py < h {
				img.SetRGBA(px, py, markerC)
			}
		}
	}

	return img
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
