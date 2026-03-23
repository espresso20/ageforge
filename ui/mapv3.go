package ui

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/espresso20/ageforge/game"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// v3Biome enumerates terrain biome types
type v3Biome int

const (
	biomeOcean v3Biome = iota
	biomeCoast
	biomePlains
	biomeGrassland
	biomeForest
	biomeJungle
	biomeDesert
	biomeSwamp
	biomeTundra
	biomeSnow
	biomeHills
	biomeMountains
	biomeRiver
)

// v3Cell holds the terrain data for a single map cell
type v3Cell struct {
	biome     v3Biome
	elevation float64 // 0.0–1.0
	moisture  float64 // 0.0–1.0
	river     bool
}

// v3TileStyle holds the rendering info for a single cell
type v3TileStyle struct {
	r  rune
	fg tcell.Color
	bg tcell.Color
}

// MapV3 is a terrain-rendered map widget using multi-octave noise and biome simulation
type MapV3 struct {
	mu         sync.Mutex
	state      game.GameState
	terrainW   int
	terrainH   int
	terrain    [][]v3Cell
}

// NewMapV3 creates a new MapV3 instance
func NewMapV3() *MapV3 { return &MapV3{} }

// Refresh stores the latest game state for the next draw
func (m *MapV3) Refresh(state game.GameState) {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()
}

// Build creates the tview primitive for the map
func (m *MapV3) Build(state game.GameState) tview.Primitive {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()

	box := tview.NewBox()
	box.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		m.mu.Lock()
		s := m.state
		m.mu.Unlock()
		m.renderV3(screen, s, x, y, w, h)
		return x, y, w, h
	})
	return box
}

// --- Noise functions ---

// v3Hash produces a deterministic float64 0–1 for integer lattice point
func v3Hash(x, y int, seed uint64) float64 {
	h := fnv.New64a()
	var buf [24]byte
	binary.LittleEndian.PutUint64(buf[0:], seed)
	binary.LittleEndian.PutUint64(buf[8:], uint64(x+100000))
	binary.LittleEndian.PutUint64(buf[16:], uint64(y+100000))
	h.Write(buf[:])
	return float64(h.Sum64()%100000) / 100000.0
}

// v3Noise produces bilinear-interpolated value noise with smoothstep
func v3Noise(x, y int, seed uint64, scale float64) float64 {
	fx := float64(x) / scale
	fy := float64(y) / scale
	x0 := int(math.Floor(fx))
	x1 := x0 + 1
	y0 := int(math.Floor(fy))
	y1 := y0 + 1
	tx := fx - math.Floor(fx)
	ty := fy - math.Floor(fy)
	tx = tx * tx * (3 - 2*tx) // smoothstep
	ty = ty * ty * (3 - 2*ty)
	return v3Lerp(
		v3Lerp(v3Hash(x0, y0, seed), v3Hash(x1, y0, seed), tx),
		v3Lerp(v3Hash(x0, y1, seed), v3Hash(x1, y1, seed), tx),
		ty,
	)
}

func v3Lerp(a, b, t float64) float64 { return a + t*(b-a) }

// v3FBM produces 4-octave fractional Brownian motion noise
func v3FBM(x, y int, seed uint64) float64 {
	v := v3Noise(x, y, seed, 60.0) * 0.50
	v += v3Noise(x, y, seed+1, 30.0) * 0.25
	v += v3Noise(x, y, seed+2, 15.0) * 0.15
	v += v3Noise(x, y, seed+3, 7.0) * 0.10
	return v // 0–1
}

// --- Terrain generation ---

func v3Smoothstep(edge0, edge1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-edge0)/(edge1-edge0)))
	return t * t * (3 - 2*t)
}

func whittakerBiome(elev, moisture float64) v3Biome {
	switch {
	case elev < 0.15:
		return biomeOcean
	case elev < 0.20:
		return biomeCoast
	case elev > 0.78:
		if moisture > 0.5 {
			return biomeSnow
		}
		return biomeMountains
	case elev > 0.65:
		if moisture > 0.6 {
			return biomeForest
		}
		return biomeHills
	case elev > 0.50:
		if moisture > 0.65 {
			return biomeForest
		}
		if moisture > 0.35 {
			return biomeGrassland
		}
		return biomePlains
	case elev > 0.35:
		if moisture > 0.75 {
			return biomeJungle
		}
		if moisture > 0.55 {
			return biomeGrassland
		}
		if moisture > 0.30 {
			return biomePlains
		}
		return biomeDesert
	default: // low elevation
		if moisture > 0.70 {
			return biomeSwamp
		}
		if moisture > 0.40 {
			return biomeGrassland
		}
		return biomePlains
	}
}

func carveRivers(cells [][]v3Cell, seed uint64, w, h int) {
	rng := rand.New(rand.NewSource(int64(seed)))
	numRivers := 3 + rng.Intn(3)

	for r := 0; r < numRivers; r++ {
		startX := w/5 + rng.Intn(w*3/5)
		startY := h/5 + rng.Intn(h*3/5)
		if cells[startY][startX].elevation < 0.55 {
			continue
		}

		x, y := startX, startY
		for steps := 0; steps < w+h; steps++ {
			if x < 0 || y < 0 || x >= w || y >= h {
				break
			}
			if cells[y][x].biome == biomeOcean || cells[y][x].biome == biomeCoast {
				break
			}

			cells[y][x].river = true

			bestX, bestY := x, y
			bestElev := cells[y][x].elevation
			for _, d := range [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				if cells[ny][nx].elevation < bestElev {
					bestElev = cells[ny][nx].elevation
					bestX, bestY = nx, ny
				}
			}
			if bestX == x && bestY == y {
				break // local minimum — stop
			}
			x, y = bestX, bestY
		}
	}
}

func generateV3Terrain(seed uint64, w, h int) [][]v3Cell {
	cells := make([][]v3Cell, h)
	for y := range cells {
		cells[y] = make([]v3Cell, w)
	}

	cx, cy := float64(w)/2, float64(h)/2

	// Pass 1: elevation and moisture maps
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			elev := v3FBM(x, y, seed)
			moist := v3FBM(x, y, seed+9999)

			// Radial falloff — ocean around edges
			dx := (float64(x) - cx) / cx
			dy := (float64(y) - cy) / cy
			dist := math.Sqrt(dx*dx + dy*dy)
			falloff := 1.0 - v3Smoothstep(0.55, 0.90, dist)
			elev = elev * falloff

			cells[y][x] = v3Cell{elevation: elev, moisture: moist}
		}
	}

	// Pass 2: biome assignment (Whittaker diagram)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := &cells[y][x]
			c.biome = whittakerBiome(c.elevation, c.moisture)
		}
	}

	// Pass 3: river carving
	carveRivers(cells, seed, w, h)

	// Pass 4: coast detection
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if cells[y][x].biome != biomeOcean {
				continue
			}
			for _, d := range [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				if cells[ny][nx].biome != biomeOcean {
					cells[y][x].biome = biomeCoast
					break
				}
			}
		}
	}

	return cells
}

// --- Biome tile rendering ---

func v3BiomeTile(biome v3Biome, river bool, era int) v3TileStyle {
	if river {
		return v3TileStyle{'~', tcell.NewHexColor(0x3a6aff), tcell.NewHexColor(0x0a1a40)}
	}

	dark := era >= 6
	space := era >= 8

	if space {
		switch biome {
		case biomeOcean:
			return v3TileStyle{' ', tcell.NewHexColor(0x050510), tcell.NewHexColor(0x050510)}
		case biomeCoast:
			return v3TileStyle{' ', tcell.NewHexColor(0x080818), tcell.NewHexColor(0x080818)}
		case biomeMountains:
			return v3TileStyle{'▲', tcell.NewHexColor(0x505060), tcell.NewHexColor(0x202030)}
		default:
			return v3TileStyle{'·', tcell.NewHexColor(0x202028), tcell.NewHexColor(0x0f0f18)}
		}
	}

	dimFactor := 1.0
	if dark {
		dimFactor = 0.5
	}

	dim := func(hex uint32) tcell.Color {
		r := uint8(float64((hex>>16)&0xff) * dimFactor)
		g := uint8(float64((hex>>8)&0xff) * dimFactor)
		b := uint8(float64(hex&0xff) * dimFactor)
		return tcell.NewRGBColor(int32(r), int32(g), int32(b))
	}

	switch biome {
	case biomeOcean:
		return v3TileStyle{' ', dim(0x0a2a5a), dim(0x0a2a5a)}
	case biomeCoast:
		return v3TileStyle{' ', dim(0x1a4a7a), dim(0x1a4a7a)}
	case biomePlains:
		return v3TileStyle{'·', dim(0x7a8a3a), dim(0x8a9a4a)}
	case biomeGrassland:
		return v3TileStyle{'·', dim(0x3a6a2a), dim(0x4a7c3a)}
	case biomeForest:
		return v3TileStyle{'♣', dim(0x0e3a08), dim(0x1e4a10)}
	case biomeJungle:
		return v3TileStyle{'♣', dim(0x0a3005), dim(0x144008)}
	case biomeDesert:
		return v3TileStyle{'·', dim(0xc0a060), dim(0xa08040)}
	case biomeSwamp:
		return v3TileStyle{'·', dim(0x405030), dim(0x304020)}
	case biomeTundra:
		return v3TileStyle{'·', dim(0x9aaba0), dim(0x607870)}
	case biomeSnow:
		return v3TileStyle{'*', dim(0xe0e8ec), dim(0xc0c8cc)}
	case biomeHills:
		return v3TileStyle{'n', dim(0x6a7a2a), dim(0x5a6a30)}
	case biomeMountains:
		return v3TileStyle{'▲', dim(0x7a6a50), dim(0x3a2e20)}
	default:
		return v3TileStyle{'·', dim(0x4a7c3a), dim(0x3a6a2a)}
	}
}

// --- City overlay ---

func v3CityTile(domain string, era int) v3TileStyle {
	r, fg := v3DomainRune(domain)
	bg := tcell.NewHexColor(0x1a1a1a)
	if era >= 6 {
		bg = tcell.NewHexColor(0x050510)
	}
	return v3TileStyle{r, fg, bg}
}

func v3DomainRune(domain string) (rune, tcell.Color) {
	switch domain {
	case "food":
		return '⌂', tcell.NewHexColor(0xd4a050)
	case "lumber":
		return '♣', tcell.NewHexColor(0x50a050)
	case "masonry":
		return '▪', tcell.NewHexColor(0xa0a0a0)
	case "metallurgy":
		return '⚒', tcell.NewHexColor(0xe07030)
	case "energy":
		return '⚡', tcell.NewHexColor(0xf0d020)
	case "military":
		return '⚔', tcell.NewHexColor(0xe04040)
	case "knowledge":
		return '◎', tcell.NewHexColor(0x60b0f0)
	case "faith":
		return '✚', tcell.NewHexColor(0xf0f0c0)
	case "trade":
		return '$', tcell.NewHexColor(0xf0c040)
	case "engineering":
		return '⚙', tcell.NewHexColor(0x40d0d0)
	case "hacker":
		return '#', tcell.NewHexColor(0x40f040)
	case "astronaut":
		return '◆', tcell.NewHexColor(0xf0f0f0)
	default:
		return '▣', tcell.NewHexColor(0xd0d0d0)
	}
}

// --- Center tile ---

func v3CenterTile(era int) v3TileStyle {
	switch {
	case era >= 8:
		return v3TileStyle{'◈', tcell.ColorWhite, tcell.NewHexColor(0x000010)}
	case era >= 6:
		return v3TileStyle{'▣', tcell.ColorAqua, tcell.NewHexColor(0x050520)}
	case era >= 4:
		return v3TileStyle{'▣', tcell.ColorSkyblue, tcell.NewHexColor(0x102040)}
	case era >= 2:
		return v3TileStyle{'♜', tcell.ColorWhite, tcell.NewHexColor(0x404040)}
	default:
		return v3TileStyle{'⌂', tcell.ColorWhite, tcell.NewHexColor(0x6a4a20)}
	}
}

// --- City radius ---

func v3CityRadius(buildingCount int) int {
	switch {
	case buildingCount < 5:
		return 1
	case buildingCount < 15:
		return 3
	case buildingCount < 35:
		return 5
	case buildingCount < 70:
		return 7
	case buildingCount < 140:
		return 9
	default:
		return 11
	}
}

// --- Seed and era helpers ---

func mapV3Seed(state game.GameState) uint64 {
	h := fnv.New64a()
	h.Write([]byte("mapv3"))
	return h.Sum64()
}

func ageEraV3(ageKey string) int {
	switch ageKey {
	case "primitive_age", "stone_age":
		return 0
	case "bronze_age", "iron_age", "classical_age":
		return 1
	case "medieval_age", "renaissance_age", "colonial_age":
		return 2
	case "industrial_age", "victorian_age", "electric_age":
		return 3
	case "atomic_age", "modern_age", "information_age":
		return 4
	case "digital_age":
		return 5
	case "cyberpunk_age", "fusion_age":
		return 6
	case "space_age", "interstellar_age":
		return 7
	default:
		return 8
	}
}

func v3AgeName(ageKey string) string {
	names := map[string]string{
		"primitive_age":    "Primitive Age",
		"stone_age":        "Stone Age",
		"bronze_age":       "Bronze Age",
		"iron_age":         "Iron Age",
		"classical_age":    "Classical Age",
		"medieval_age":     "Medieval Age",
		"renaissance_age":  "Renaissance Age",
		"colonial_age":     "Colonial Age",
		"industrial_age":   "Industrial Age",
		"victorian_age":    "Victorian Age",
		"electric_age":     "Electric Age",
		"atomic_age":       "Atomic Age",
		"modern_age":       "Modern Age",
		"information_age":  "Information Age",
		"digital_age":      "Digital Age",
		"cyberpunk_age":    "Cyberpunk Age",
		"fusion_age":       "Fusion Age",
		"space_age":        "Space Age",
		"interstellar_age": "Interstellar Age",
		"galactic_age":     "Galactic Age",
		"quantum_age":      "Quantum Age",
		"transcendent_age": "Transcendent Age",
	}
	if n, ok := names[ageKey]; ok {
		return n
	}
	return ageKey
}

// --- Main render function ---

func (m *MapV3) renderV3(screen tcell.Screen, state game.GameState, x, y, w, h int) {
	if w <= 0 || h <= 1 {
		return
	}

	// Regenerate terrain if dimensions changed
	m.mu.Lock()
	if m.terrainW != w || m.terrainH != h-1 {
		seed := mapV3Seed(state)
		m.terrain = generateV3Terrain(seed, w, h-1)
		m.terrainW = w
		m.terrainH = h - 1
	}
	terrain := m.terrain
	m.mu.Unlock()

	era := ageEraV3(state.Age)
	cx, cy := w/2, (h-1)/2

	// Count buildings and compute city radius
	builtCount := 0
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			builtCount++
		}
	}
	cityR := v3CityRadius(builtCount)

	// Collect built buildings sorted deterministically
	type bldEntry struct {
		key    string
		domain string
		wonder bool
	}
	var buildings []bldEntry
	for k, bs := range state.Buildings {
		if bs.Count <= 0 {
			continue
		}
		buildings = append(buildings, bldEntry{
			key:    k,
			domain: bs.WorkerDomain,
			wonder: bs.Category == "wonder",
		})
	}
	sort.Slice(buildings, func(i, j int) bool { return buildings[i].key < buildings[j].key })

	// Build a set of city positions (deterministic golden-angle spiral placement)
	cityPositions := make(map[[2]int]bldEntry)
	for i, bld := range buildings {
		angle := float64(i) * 2.399 // golden angle in radians
		radius := float64(3 + (i/8)*3)
		if radius > float64(cityR) {
			radius = float64(cityR)
		}
		bx := cx + int(math.Cos(angle)*radius)
		by := cy + int(math.Sin(angle)*radius)
		cityPositions[[2]int{bx, by}] = bld
	}

	// Render terrain
	for ty := 0; ty < h-1; ty++ {
		for tx := 0; tx < w; tx++ {
			var tile v3TileStyle

			if tx == cx && ty == cy {
				// City center
				tile = v3CenterTile(era)
			} else if b, ok := cityPositions[[2]int{tx, ty}]; ok {
				// Building icon
				tile = v3CityTile(b.domain, era)
				if b.wonder {
					tile.r = '★'
					tile.fg = tcell.NewHexColor(0xffd700)
				}
			} else if ty < len(terrain) && tx < len(terrain[ty]) {
				cell := terrain[ty][tx]
				tile = v3BiomeTile(cell.biome, cell.river, era)
			} else {
				tile = v3TileStyle{'·', tcell.ColorGray, tcell.ColorBlack}
			}

			style := tcell.StyleDefault.Background(tile.bg).Foreground(tile.fg)
			screen.SetContent(x+tx, y+ty, tile.r, nil, style)
		}
	}

	// Status bar
	statusY := y + h - 1
	ageName := v3AgeName(state.Age)
	status := fmt.Sprintf("  Map v3 (Noise Terrain) — %s — %d buildings  [MAPV3]", ageName, builtCount)
	statusRunes := []rune(status)
	for i := 0; i < w; i++ {
		r := ' '
		if i < len(statusRunes) {
			r = statusRunes[i]
		}
		style := tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorGold)
		screen.SetContent(x+i, statusY, r, nil, style)
	}
}
