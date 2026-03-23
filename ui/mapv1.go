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

// ---------------------------------------------------------------------------
// Biome types
// ---------------------------------------------------------------------------

type v1Biome int

const (
	biomeOcean      v1Biome = iota
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

// v1Cell holds the terrain data for a single map cell.
type v1Cell struct {
	biome     v1Biome
	elevation float64
	moisture  float64
	river     bool
}

// v1TileStyle holds the rendering info for a single cell.
type v1TileStyle struct {
	r  rune
	fg tcell.Color
	bg tcell.Color
}

// ---------------------------------------------------------------------------
// MapV1 struct — v1 API + v3 terrain cache
// ---------------------------------------------------------------------------

// MapV1 is a character-native Civ-1-style civilization map widget.
// Uses 4-pass FBM terrain generation with Whittaker biomes, rivers, and coast
// detection. Terrain is cached and only regenerated on resize.
type MapV1 struct {
	mu       sync.Mutex
	state    game.GameState
	terrainW int
	terrainH int
	terrain  [][]v1Cell
}

// NewMapV1 creates a new MapV1 instance.
func NewMapV1() *MapV1 { return &MapV1{} }

// Refresh updates the stored state; the box reads it on its next draw cycle.
func (m *MapV1) Refresh(state game.GameState) {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()
}

// Build stores initial state and returns a tview.Primitive whose draw function
// calls renderV1 on every render pass.
func (m *MapV1) Build(state game.GameState) tview.Primitive {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()

	box := tview.NewBox()
	box.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		m.mu.Lock()
		s := m.state
		m.mu.Unlock()
		m.renderV1(screen, s, x, y, w, h)
		return x, y, w, h
	})
	return box
}

// ---------------------------------------------------------------------------
// Seed helper
// ---------------------------------------------------------------------------

func mapV1Seed(_ game.GameState) uint64 {
	h := fnv.New64a()
	h.Write([]byte("mapv1"))
	return h.Sum64()
}

// ---------------------------------------------------------------------------
// Age era and name helpers
// ---------------------------------------------------------------------------

func ageEraV1(ageKey string) int {
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

func v1AgeName(ageKey string) string {
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

// ---------------------------------------------------------------------------
// Noise functions (from v3)
// ---------------------------------------------------------------------------

func v1HashF(x, y int, seed uint64) float64 {
	h := fnv.New64a()
	var buf [24]byte
	binary.LittleEndian.PutUint64(buf[0:], seed)
	binary.LittleEndian.PutUint64(buf[8:], uint64(x+100000))
	binary.LittleEndian.PutUint64(buf[16:], uint64(y+100000))
	h.Write(buf[:])
	return float64(h.Sum64()%100000) / 100000.0
}

func v1Lerp(a, b, t float64) float64 { return a + t*(b-a) }

func v1NoiseAt(x, y int, seed uint64, scale float64) float64 {
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
	return v1Lerp(
		v1Lerp(v1HashF(x0, y0, seed), v1HashF(x1, y0, seed), tx),
		v1Lerp(v1HashF(x0, y1, seed), v1HashF(x1, y1, seed), tx),
		ty,
	)
}

// v1FBM produces 4-octave fractional Brownian motion noise.
func v1FBM(x, y int, seed uint64) float64 {
	v := v1NoiseAt(x, y, seed, 60.0) * 0.50
	v += v1NoiseAt(x, y, seed+1, 30.0) * 0.25
	v += v1NoiseAt(x, y, seed+2, 15.0) * 0.15
	v += v1NoiseAt(x, y, seed+3, 7.0) * 0.10
	return v // 0–1
}

// ---------------------------------------------------------------------------
// Terrain generation (4-pass from v3, renamed for v1)
// ---------------------------------------------------------------------------

func v1Smoothstep(edge0, edge1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-edge0)/(edge1-edge0)))
	return t * t * (3 - 2*t)
}

func whittakerBiomeV1(elev, moisture float64) v1Biome {
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
	default:
		if moisture > 0.70 {
			return biomeSwamp
		}
		if moisture > 0.40 {
			return biomeGrassland
		}
		return biomePlains
	}
}

func carveRiversV1(cells [][]v1Cell, seed uint64, w, h int) {
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

func generateV1Terrain(seed uint64, w, h int) [][]v1Cell {
	cells := make([][]v1Cell, h)
	for y := range cells {
		cells[y] = make([]v1Cell, w)
	}

	cx, cy := float64(w)/2, float64(h)/2

	// Pass 1: elevation and moisture maps
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			elev := v1FBM(x, y, seed)
			moist := v1FBM(x, y, seed+9999)

			// Radial falloff — ocean around edges
			dx := (float64(x) - cx) / cx
			dy := (float64(y) - cy) / cy
			dist := math.Sqrt(dx*dx + dy*dy)
			falloff := 1.0 - v1Smoothstep(0.55, 0.90, dist)
			elev = elev * falloff

			cells[y][x] = v1Cell{elevation: elev, moisture: moist}
		}
	}

	// Pass 2: biome assignment (Whittaker diagram)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := &cells[y][x]
			c.biome = whittakerBiomeV1(c.elevation, c.moisture)
		}
	}

	// Pass 3: river carving
	carveRiversV1(cells, seed, w, h)

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

// ---------------------------------------------------------------------------
// Biome tile rendering
// ---------------------------------------------------------------------------

func v1BiomeTile(biome v1Biome, river bool, era int) v1TileStyle {
	dark := era >= 6
	spaceEra := era >= 8

	if spaceEra {
		switch biome {
		case biomeOcean:
			return v1TileStyle{' ', tcell.NewHexColor(0x050510), tcell.NewHexColor(0x050510)}
		case biomeCoast:
			return v1TileStyle{' ', tcell.NewHexColor(0x080818), tcell.NewHexColor(0x080818)}
		case biomeMountains:
			return v1TileStyle{'▲', tcell.NewHexColor(0x505060), tcell.NewHexColor(0x202030)}
		default:
			return v1TileStyle{' ', tcell.NewHexColor(0x0f0f18), tcell.NewHexColor(0x0f0f18)}
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

	// River: blend into terrain — blue fg over terrain bg, no dark background
	if river {
		var terrainBg tcell.Color
		switch biome {
		case biomeGrassland:
			terrainBg = dim(0x4a7c3a)
		case biomePlains:
			terrainBg = dim(0x8a9a4a)
		default:
			terrainBg = dim(0x3a6a2a)
		}
		return v1TileStyle{'~', dim(0x5a9aff), terrainBg}
	}

	switch biome {
	case biomeOcean:
		return v1TileStyle{' ', dim(0x0a2a5a), dim(0x0a2a5a)}
	case biomeCoast:
		return v1TileStyle{' ', dim(0x1a4a7a), dim(0x1a4a7a)}
	// Flat terrain — space char, solid fill, no scan lines
	case biomePlains:
		return v1TileStyle{' ', dim(0x8a9a4a), dim(0x8a9a4a)}
	case biomeGrassland:
		return v1TileStyle{' ', dim(0x4a7c3a), dim(0x4a7c3a)}
	case biomeDesert:
		return v1TileStyle{' ', dim(0xa08040), dim(0xa08040)}
	case biomeSwamp:
		return v1TileStyle{' ', dim(0x304020), dim(0x304020)}
	case biomeTundra:
		return v1TileStyle{' ', dim(0x607870), dim(0x607870)}
	// Textured terrain — keep distinctive runes but fg close to bg to reduce stripe
	case biomeForest:
		return v1TileStyle{'♣', dim(0x1a4a10), dim(0x1e4a10)}
	case biomeJungle:
		return v1TileStyle{'♣', dim(0x0e3a08), dim(0x144008)}
	case biomeSnow:
		return v1TileStyle{' ', dim(0xd0d8dc), dim(0xd0d8dc)}
	case biomeHills:
		return v1TileStyle{' ', dim(0x5a6a30), dim(0x5a6a30)}
	case biomeMountains:
		return v1TileStyle{'▲', dim(0x6a5a40), dim(0x3a2e20)}
	default:
		return v1TileStyle{' ', dim(0x4a7c3a), dim(0x4a7c3a)}
	}
}

// ---------------------------------------------------------------------------
// Building placement — outward grid (from v3)
// ---------------------------------------------------------------------------

// v1OutwardGrid returns n grid offsets radiating outward from (0,0) in
// concentric square rings: ring 1 = 8 positions, ring 2 = 16, etc.
func v1OutwardGrid(n int) [][2]int {
	positions := make([][2]int, 0, n)
	for ring := 1; len(positions) < n; ring++ {
		for dx := -ring; dx <= ring; dx++ {
			for dy := -ring; dy <= ring; dy++ {
				adx, ady := dx, dy
				if adx < 0 {
					adx = -adx
				}
				if ady < 0 {
					ady = -ady
				}
				if adx == ring || ady == ring {
					positions = append(positions, [2]int{dx, dy})
				}
			}
		}
	}
	return positions
}

// ---------------------------------------------------------------------------
// Building domain icons (from v1 — better than v3)
// ---------------------------------------------------------------------------

func v1DomainTile(domain, category string) (rune, tcell.Color) {
	if category == "wonder" {
		return '★', tcell.NewHexColor(0xffd700)
	}
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

// ---------------------------------------------------------------------------
// City center tile (from v3)
// ---------------------------------------------------------------------------

func v1CenterTile(era int) v1TileStyle {
	switch {
	case era >= 8:
		return v1TileStyle{'◈', tcell.ColorWhite, tcell.NewHexColor(0x000010)}
	case era >= 6:
		return v1TileStyle{'▣', tcell.ColorAqua, tcell.NewHexColor(0x050520)}
	case era >= 4:
		return v1TileStyle{'▣', tcell.ColorSkyblue, tcell.NewHexColor(0x102040)}
	case era >= 2:
		return v1TileStyle{'♜', tcell.ColorWhite, tcell.NewHexColor(0x404040)}
	default:
		return v1TileStyle{'⌂', tcell.ColorWhite, tcell.NewHexColor(0x6a4a20)}
	}
}

// ---------------------------------------------------------------------------
// Main render function
// ---------------------------------------------------------------------------

func (m *MapV1) renderV1(screen tcell.Screen, state game.GameState, x, y, w, h int) {
	if w <= 0 || h <= 1 {
		return
	}

	// Regenerate terrain only when dimensions change
	m.mu.Lock()
	if m.terrainW != w || m.terrainH != h-1 {
		seed := mapV1Seed(state)
		m.terrain = generateV1Terrain(seed, w, h-1)
		m.terrainW = w
		m.terrainH = h - 1
	}
	terrain := m.terrain
	m.mu.Unlock()

	era := ageEraV1(state.Age)
	cx, cy := w/2, (h-1)/2

	// Collect built buildings sorted deterministically
	type bldEntry struct {
		key      string
		domain   string
		category string
	}
	var buildings []bldEntry
	for k, bs := range state.Buildings {
		if bs.Count <= 0 {
			continue
		}
		buildings = append(buildings, bldEntry{
			key:      k,
			domain:   bs.WorkerDomain,
			category: bs.Category,
		})
	}
	sort.Slice(buildings, func(i, j int) bool { return buildings[i].key < buildings[j].key })

	builtCount := len(buildings)

	// Build city positions using outward grid scan with spacing=2
	const v1Spacing = 2
	cityPositions := make(map[[2]int]bldEntry)
	gridPos := v1OutwardGrid(len(buildings))
	for i, bld := range buildings {
		if i >= len(gridPos) {
			break
		}
		bx := cx + gridPos[i][0]*v1Spacing
		by := cy + gridPos[i][1]*v1Spacing
		cityPositions[[2]int{bx, by}] = bld
	}

	// Render terrain + buildings
	for ty := 0; ty < h-1; ty++ {
		for tx := 0; tx < w; tx++ {
			var tile v1TileStyle

			if tx == cx && ty == cy {
				// City center
				tile = v1CenterTile(era)
			} else if b, ok := cityPositions[[2]int{tx, ty}]; ok {
				// Building icon — use v1 domain runes
				br, bfg := v1DomainTile(b.domain, b.category)
				bg := tcell.NewHexColor(0x1a1a1a)
				if era >= 6 {
					bg = tcell.NewHexColor(0x050510)
				}
				tile = v1TileStyle{br, bfg, bg}
			} else if ty < len(terrain) && tx < len(terrain[ty]) {
				cell := terrain[ty][tx]
				tile = v1BiomeTile(cell.biome, cell.river, era)
			} else {
				tile = v1TileStyle{'·', tcell.ColorGray, tcell.ColorBlack}
			}

			style := tcell.StyleDefault.Background(tile.bg).Foreground(tile.fg)
			screen.SetContent(x+tx, y+ty, tile.r, nil, style)
		}
	}

	// Status bar
	statusY := y + h - 1
	ageName := state.AgeName
	if ageName == "" {
		ageName = v1AgeName(state.Age)
	}
	status := fmt.Sprintf("  Map — %s — %d buildings  [MAPV1]", ageName, builtCount)
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
