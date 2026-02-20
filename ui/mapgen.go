package ui

import (
	"hash/fnv"
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// MapCell represents a single cell in the generated map
type MapCell struct {
	Char  rune
	Style tcell.Style
}

// TerrainTheme defines the visual look of terrain for an era
type TerrainTheme struct {
	Ground      rune
	Water       rune
	Tree        rune
	GroundColor tcell.Color
	WaterColor  tcell.Color
	TreeColor   tcell.Color
	Road        rune
	RoadColor   tcell.Color
	Extra       rune
	ExtraColor  tcell.Color
}

// MapGenConfig holds parameters for map generation
type MapGenConfig struct {
	Width       int
	Height      int
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

// getTerrainTheme returns the terrain theme for an era
func getTerrainTheme(era int) TerrainTheme {
	switch era {
	case 0: // primitive
		return TerrainTheme{'.', '~', 'T', tcell.ColorGreen, tcell.ColorBlue, tcell.ColorDarkGreen, ' ', tcell.ColorDefault, ',', tcell.ColorDarkOliveGreen}
	case 1: // ancient
		return TerrainTheme{'.', '~', 'T', tcell.ColorDarkKhaki, tcell.ColorBlue, tcell.ColorForestGreen, ' ', tcell.ColorDefault, ',', tcell.ColorOlive}
	case 2: // medieval
		return TerrainTheme{'.', '~', 'T', tcell.ColorDarkGreen, tcell.ColorDarkBlue, tcell.ColorForestGreen, '=', tcell.ColorGray, '#', tcell.ColorDimGray}
	case 3: // industrial
		return TerrainTheme{'.', '~', 't', tcell.ColorDimGray, tcell.ColorDarkSlateGray, tcell.ColorDarkOliveGreen, '=', tcell.ColorGray, '|', tcell.ColorGray}
	case 4: // modern
		return TerrainTheme{'.', '~', 't', tcell.ColorDarkGray, tcell.ColorSlateBlue, tcell.ColorOliveDrab, '=', tcell.ColorLightGray, '#', tcell.ColorGray}
	case 5: // digital
		return TerrainTheme{' ', '~', ' ', tcell.ColorDarkBlue, tcell.ColorMediumBlue, tcell.ColorDarkBlue, '─', tcell.ColorDarkCyan, '·', tcell.ColorDodgerBlue}
	case 6: // cyber
		return TerrainTheme{' ', '≈', ' ', tcell.ColorBlack, tcell.ColorDarkMagenta, tcell.ColorBlack, '─', tcell.ColorHotPink, '·', tcell.ColorLime}
	case 7: // space
		return TerrainTheme{'·', ' ', ' ', tcell.ColorBlack, tcell.ColorDarkBlue, tcell.ColorBlack, ' ', tcell.ColorDefault, '░', tcell.ColorDarkSlateGray}
	case 8: // cosmic
		return TerrainTheme{' ', '∿', ' ', tcell.ColorBlack, tcell.ColorDarkViolet, tcell.ColorBlack, ' ', tcell.ColorDefault, '✧', tcell.ColorGold}
	}
	return TerrainTheme{'.', '~', 'T', tcell.ColorGreen, tcell.ColorBlue, tcell.ColorDarkGreen, ' ', tcell.ColorDefault, ',', tcell.ColorGreen}
}

// buildingGlyph returns the display character and color for a building category
func buildingGlyph(category string) (rune, tcell.Color) {
	switch category {
	case "housing":
		return '⌂', tcell.ColorGreen
	case "production":
		return '▣', tcell.ColorYellow
	case "research":
		return '◈', tcell.ColorDarkCyan
	case "military":
		return '⛊', tcell.ColorRed
	case "storage":
		return '□', tcell.ColorDodgerBlue
	case "wonder":
		return '★', tcell.ColorGold
	default:
		return '■', tcell.ColorWhite
	}
}

// hashKey returns a deterministic hash for a string key
func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

// GenerateMap creates a procedural map grid
func GenerateMap(cfg MapGenConfig) [][]MapCell {
	w, h := cfg.Width, cfg.Height
	if w < 5 || h < 5 {
		return nil
	}

	era := eraFromAge(cfg.AgeKey)
	theme := getTerrainTheme(era)

	// Initialize grid with ground
	grid := make([][]MapCell, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]MapCell, w)
		for x := 0; x < w; x++ {
			ch := theme.Ground
			grid[y][x] = MapCell{Char: ch, Style: tcell.StyleDefault.Foreground(theme.GroundColor)}
		}
	}

	// Add river (sinusoidal path)
	riverX := w / 3
	riverWidth := 1
	if cfg.DetailLevel > 0 {
		riverWidth = 2
	}
	for y := 0; y < h; y++ {
		rx := riverX + int(math.Sin(float64(y)*0.4)*float64(w/8))
		for dx := 0; dx < riverWidth; dx++ {
			px := rx + dx
			if px >= 0 && px < w {
				grid[y][px] = MapCell{Char: theme.Water, Style: tcell.StyleDefault.Foreground(theme.WaterColor)}
			}
		}
	}

	// Scatter trees/vegetation
	treeDensity := 0.15
	if era >= 3 {
		treeDensity = 0.05
	}
	if era >= 5 {
		treeDensity = 0.02
	}
	seed := hashKey(cfg.AgeKey + "_trees")
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if grid[y][x].Char == theme.Ground {
				v := hashKey(string(rune(seed)) + string(rune(x*1000+y)))
				if float64(v%1000)/1000.0 < treeDensity {
					if theme.Tree != ' ' {
						grid[y][x] = MapCell{Char: theme.Tree, Style: tcell.StyleDefault.Foreground(theme.TreeColor)}
					}
				} else if float64(v%1000)/1000.0 < treeDensity*2 {
					if theme.Extra != ' ' {
						grid[y][x] = MapCell{Char: theme.Extra, Style: tcell.StyleDefault.Foreground(theme.ExtraColor)}
					}
				}
			}
		}
	}

	// Place buildings radially from center
	cx, cy := w/2, h/2
	totalBuildings := 0
	for _, bs := range cfg.Buildings {
		if bs.Unlocked {
			totalBuildings += bs.Count
		}
	}

	for key, bs := range cfg.Buildings {
		if !bs.Unlocked || bs.Count == 0 {
			continue
		}
		glyph, color := buildingGlyph(bs.Category)

		// Determine ring based on category
		ringMin, ringMax := 2.0, float64(min(w, h)/3)
		switch bs.Category {
		case "housing":
			ringMin, ringMax = 1.0, float64(min(w, h)/5)
		case "production":
			ringMin, ringMax = 2.0, float64(min(w, h)/4)
		case "military":
			ringMin, ringMax = float64(min(w, h)/5), float64(min(w, h)/3)
		case "wonder":
			ringMin, ringMax = 1.0, float64(min(w, h)/4)
		}

		for i := 0; i < bs.Count; i++ {
			bHash := hashKey(key + string(rune(i)))
			angle := float64(bHash%360) * math.Pi / 180.0
			dist := ringMin + float64(bHash%100)/100.0*(ringMax-ringMin)
			bx := cx + int(math.Cos(angle)*dist)
			by := cy + int(math.Sin(angle)*dist*0.6) // compress vertically
			if bx >= 0 && bx < w && by >= 0 && by < h {
				if cfg.DetailLevel > 0 && (bs.Category == "wonder" || bs.Count >= 3) {
					// Larger glyph for detail mode
					grid[by][bx] = MapCell{Char: glyph, Style: tcell.StyleDefault.Foreground(color).Bold(true)}
				} else {
					grid[by][bx] = MapCell{Char: glyph, Style: tcell.StyleDefault.Foreground(color)}
				}
			}
		}
	}

	// Place villager dots (animated by tick)
	villagerCount := cfg.TotalPop
	if villagerCount > 10 {
		villagerCount = 10
	}
	for i := 0; i < villagerCount; i++ {
		vHash := hashKey(string(rune(cfg.Tick*7 + i*31)))
		radius := float64(min(w, h)/6) + float64(vHash%100)/100.0*float64(min(w, h)/6)
		angle := float64(vHash%360) * math.Pi / 180.0
		vx := cx + int(math.Cos(angle)*radius)
		vy := cy + int(math.Sin(angle)*radius*0.6)
		if vx >= 0 && vx < w && vy >= 0 && vy < h {
			grid[vy][vx] = MapCell{Char: '●', Style: tcell.StyleDefault.Foreground(tcell.ColorWhite)}
		}
	}

	return grid
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
