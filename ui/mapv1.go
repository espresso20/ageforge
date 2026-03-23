package ui

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	"github.com/espresso20/ageforge/game"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MapV1 is a character-native Civ-1-style civilization map widget.
// Each terminal cell is drawn as a terrain tile or building icon using tcell.SetContent.
type MapV1 struct {
	mu    sync.Mutex
	state game.GameState
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
// calls drawMapV1 on every render pass.
func (m *MapV1) Build(state game.GameState) tview.Primitive {
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()

	box := tview.NewBox()
	box.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		m.mu.Lock()
		s := m.state
		m.mu.Unlock()
		drawMapV1(screen, s, x, y, w, h)
		return x, y, w, h
	})
	return box
}

// ---------------------------------------------------------------------------
// Terrain types
// ---------------------------------------------------------------------------

type terrainKind int

const (
	terrainGrassland terrainKind = iota
	terrainPlains
	terrainForest
	terrainHills
	terrainMountains
	terrainOcean
	terrainCoast
	terrainDesert
	terrainTundra
	terrainRiver
)

// tileInfo holds the rune and colors for a single terminal cell.
type tileInfo struct {
	r  rune
	fg tcell.Color
	bg tcell.Color
}

// rgb is a convenience wrapper to decompose a 24-bit hex color into NewRGBColor.
func rgb(hex uint32) tcell.Color {
	r := int32((hex >> 16) & 0xff)
	g := int32((hex >> 8) & 0xff)
	b := int32(hex & 0xff)
	return tcell.NewRGBColor(r, g, b)
}

// darkenHex multiplies each RGB channel of a hex color by factor (0.0–1.0) and adds a slight blue tint.
func darkenHex(hex uint32, factor float64) tcell.Color {
	r := int32(float64((hex>>16)&0xff) * factor)
	g := int32(float64((hex>>8)&0xff) * factor)
	b := int32(float64(hex&0xff)*factor) + 8
	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b > 255 {
		b = 255
	}
	return tcell.NewRGBColor(r, g, b)
}

// darken multiplies each RGB channel of a tcell.Color by factor and adds a slight blue tint.
// It relies on the fact that tcell.NewRGBColor packs as ColorIsRGB|(r<<16|g<<8|b).
func darken(c tcell.Color, factor float64) tcell.Color {
	// Extract packed value: strip the ColorIsRGB flag bit (top bits) to get 24-bit RGB.
	packed := uint32(c) & 0xffffff
	return darkenHex(packed, factor)
}

// ---------------------------------------------------------------------------
// Seed and hash helpers
// ---------------------------------------------------------------------------

func mapV1Seed(_ game.GameState) uint64 {
	h := fnv.New64a()
	h.Write([]byte("mapv1"))
	return h.Sum64()
}

func hashTile(seed uint64, x, y int) uint64 {
	h := fnv.New64a()
	var buf [24]byte
	binary.LittleEndian.PutUint64(buf[0:], seed)
	binary.LittleEndian.PutUint64(buf[8:], uint64(x))
	binary.LittleEndian.PutUint64(buf[16:], uint64(y))
	h.Write(buf[:])
	return h.Sum64()
}

// ---------------------------------------------------------------------------
// Age era mapping
// ---------------------------------------------------------------------------

func ageEra(ageKey string) int {
	switch ageKey {
	case "primitive_age", "stone_age":
		return 0
	case "bronze_age", "iron_age", "classical_age":
		return 1
	case "medieval_age", "renaissance_age":
		return 2
	case "colonial_age", "industrial_age", "victorian_age":
		return 3
	case "electric_age", "atomic_age":
		return 4
	case "modern_age", "information_age":
		return 5
	case "digital_age":
		return 6
	case "cyberpunk_age", "fusion_age":
		return 7
	default:
		// space_age and any later ages
		return 8
	}
}

// ---------------------------------------------------------------------------
// Terrain tile rendering
// ---------------------------------------------------------------------------

func classifyTerrain(seed uint64, tx, ty, mapW, mapH int) terrainKind {
	nearEdge := tx < 4 || ty < 4 || tx >= mapW-4 || ty >= mapH-4
	// Hash at region level so adjacent cells form coherent terrain patches
	const regionSize = 6
	h := hashTile(seed, tx/regionSize, ty/regionSize)
	slot := h % 100

	// Ocean: near edge always, or slot 71-74 elsewhere
	if nearEdge || slot >= 71 && slot <= 74 {
		return terrainOcean
	}

	// Tundra: near top/bottom (but not already ocean)
	if ty < 8 || ty >= mapH-8 {
		if slot >= 92 {
			return terrainTundra
		}
	}

	// River: thin horizontal band across mid-map
	midY := mapH / 2
	if ty == midY || ty == midY+1 {
		if slot >= 96 {
			return terrainRiver
		}
	}

	switch {
	case slot <= 34:
		return terrainGrassland
	case slot <= 49:
		return terrainPlains
	case slot <= 59:
		return terrainForest
	case slot <= 66:
		return terrainHills
	case slot <= 70:
		return terrainMountains
	case slot <= 82:
		// Would be ocean but we're not near edge and slot >74, so use grassland
		return terrainGrassland
	case slot <= 88:
		return terrainDesert
	case slot <= 91:
		return terrainCoast
	case slot <= 95:
		return terrainTundra
	default:
		return terrainRiver
	}
}

func terrainTile(seed uint64, tx, ty, mapW, mapH int) tileInfo {
	kind := classifyTerrain(seed, tx, ty, mapW, mapH)
	switch kind {
	case terrainGrassland:
		return tileInfo{'·', rgb(0x2a5a1a), rgb(0x3a6b2a)}
	case terrainPlains:
		return tileInfo{'·', rgb(0x6a7a2a), rgb(0x7a8a3a)}
	case terrainForest:
		return tileInfo{'♣', rgb(0x0e3a08), rgb(0x1e4a10)}
	case terrainHills:
		return tileInfo{'n', rgb(0x4a5a20), rgb(0x5a6a30)}
	case terrainMountains:
		return tileInfo{'▲', rgb(0x7a6a50), rgb(0x3a2e20)}
	case terrainOcean:
		return tileInfo{'~', rgb(0x1a4a8a), rgb(0x0a2a5a)}
	case terrainCoast:
		return tileInfo{'~', rgb(0x3a7aaa), rgb(0x1a4a7a)}
	case terrainDesert:
		return tileInfo{'·', rgb(0xc0a060), rgb(0xa08040)}
	case terrainTundra:
		return tileInfo{'·', rgb(0x809090), rgb(0x607870)}
	case terrainRiver:
		// River uses terrain background but river-blue FG
		base := tileInfo{'~', rgb(0x3a6aff), rgb(0x3a6b2a)}
		return base
	default:
		return tileInfo{'·', rgb(0x2a5a1a), rgb(0x3a6b2a)}
	}
}

// applyEraPalette adjusts tile colors for eras 6+ and era 8 (space).
func applyEraPalette(t tileInfo, era int, kind terrainKind) tileInfo {
	if era < 6 {
		return t
	}
	if era == 8 {
		// Space: dark grey craters for land tiles
		switch kind {
		case terrainOcean, terrainCoast, terrainRiver:
			// keep water tiles but darken
		default:
			return tileInfo{'·', rgb(0x303040), rgb(0x050510)}
		}
	}
	// Eras 6+: darken all bg, add blue tint
	t.bg = darken(t.bg, 0.7)
	t.fg = darken(t.fg, 0.7)
	return t
}

// ---------------------------------------------------------------------------
// City / building placement
// ---------------------------------------------------------------------------

func cityRadius(buildingCount int) int {
	switch {
	case buildingCount < 5:
		return 1
	case buildingCount < 15:
		return 2
	case buildingCount < 35:
		return 4
	case buildingCount < 70:
		return 6
	case buildingCount < 140:
		return 8
	default:
		return 10
	}
}

func countBuiltBuildings(state game.GameState) int {
	count := 0
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			count++
		}
	}
	return count
}

// domainTile returns the rune and fg color for a building's worker domain.
func domainTile(bs game.BuildingState) (rune, tcell.Color) {
	if bs.Category == "wonder" {
		return '★', rgb(0xffd700)
	}
	switch bs.WorkerDomain {
	case "food":
		return '⌂', rgb(0xd4a050)
	case "lumber":
		return '♣', rgb(0x50a050)
	case "masonry":
		return '▪', rgb(0xa0a0a0)
	case "metallurgy":
		return '⚒', rgb(0xe07030)
	case "energy":
		return '⚡', rgb(0xf0d020)
	case "military":
		return '⚔', rgb(0xe04040)
	case "knowledge":
		return '◎', rgb(0x60b0f0)
	case "faith":
		return '✚', rgb(0xf0f0c0)
	case "trade":
		return '$', rgb(0xf0c040)
	case "engineering":
		return '⚙', rgb(0x40d0d0)
	case "hacker":
		return '#', rgb(0x40f040)
	case "astronaut":
		return '◆', rgb(0xf0f0f0)
	default:
		return '▪', rgb(0xa0a0a0)
	}
}

// cityCenterTile returns the rune and style for the city center based on era.
func cityCenterTile(era int) (rune, tcell.Color, tcell.Color) {
	switch {
	case era <= 1:
		return '⌂', tcell.ColorWhite, rgb(0x6a4a20)
	case era <= 3:
		return '♜', tcell.ColorWhiteSmoke, rgb(0x404040)
	case era <= 5:
		return '▣', tcell.ColorAqua, rgb(0x102040)
	case era <= 7:
		return '▣', tcell.ColorTeal, rgb(0x050520)
	default:
		return '◈', tcell.ColorWhiteSmoke, rgb(0x000010)
	}
}

// distance returns Chebyshev distance (max of abs dx, abs dy) — square radius.
func distance(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

// ---------------------------------------------------------------------------
// Main draw function
// ---------------------------------------------------------------------------

func drawMapV1(screen tcell.Screen, state game.GameState, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}

	seed := mapV1Seed(state)
	era := ageEra(state.Age)

	// Collect built buildings sorted by key for deterministic placement
	type builtBuilding struct {
		key string
		bs  game.BuildingState
	}
	var builtList []builtBuilding
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			builtList = append(builtList, builtBuilding{k, bs})
		}
	}
	sort.Slice(builtList, func(i, j int) bool {
		return builtList[i].key < builtList[j].key
	})

	builtCount := len(builtList)
	radius := cityRadius(builtCount)
	cx := w / 2
	cy := h / 2

	// Reserve bottom row for status line
	drawH := h - 1
	if drawH < 1 {
		drawH = 1
	}

	// Draw terrain and city tiles
	for ty := 0; ty < drawH; ty++ {
		for tx := 0; tx < w; tx++ {
			screenX := x + tx
			screenY := y + ty

			dist := distance(tx, ty, cx, cy)

			var r rune
			var style tcell.Style

			if tx == cx && ty == cy {
				// City center
				cr, cfgColor, cbgColor := cityCenterTile(era)
				style = tcell.StyleDefault.Foreground(cfgColor).Background(cbgColor)
				r = cr
			} else if dist <= radius && len(builtList) > 0 {
				// Within city radius — potentially a building tile
				cellHash := hashTile(seed, tx, ty)
				if cellHash%3 == 0 {
					b := builtList[cellHash%uint64(len(builtList))]
					br, bfg := domainTile(b.bs)
					// Use terrain as background for building tiles
					terrain := terrainTile(seed, tx, ty, w, drawH)
					terrain = applyEraPalette(terrain, era, classifyTerrain(seed, tx, ty, w, drawH))
					style = tcell.StyleDefault.Foreground(bfg).Background(terrain.bg)
					r = br
				} else {
					// Plain terrain within city
					terrain := terrainTile(seed, tx, ty, w, drawH)
					terrain = applyEraPalette(terrain, era, classifyTerrain(seed, tx, ty, w, drawH))
					style = tcell.StyleDefault.Foreground(terrain.fg).Background(terrain.bg)
					r = terrain.r
				}
			} else {
				// Pure terrain
				terrain := terrainTile(seed, tx, ty, w, drawH)
				terrain = applyEraPalette(terrain, era, classifyTerrain(seed, tx, ty, w, drawH))
				style = tcell.StyleDefault.Foreground(terrain.fg).Background(terrain.bg)
				r = terrain.r
			}

			screen.SetContent(screenX, screenY, r, nil, style)
		}
	}

	// ---------------------------------------------------------------------------
	// Status line at bottom row
	// ---------------------------------------------------------------------------
	statusY := y + h - 1
	statusStyle := tcell.StyleDefault.
		Foreground(tcell.ColorGold).
		Background(tcell.ColorDarkSlateGray)

	ageName := state.AgeName
	if ageName == "" {
		ageName = state.Age
	}
	statusText := fmt.Sprintf("  Map v1 — %s — %d buildings   %d tile radius   MAPV1",
		ageName, builtCount, radius)

	// Fill entire row with background first
	for tx := 0; tx < w; tx++ {
		screen.SetContent(x+tx, statusY, ' ', nil, statusStyle)
	}
	// Write status text
	for i, ch := range []rune(statusText) {
		if i >= w {
			break
		}
		screen.SetContent(x+i, statusY, ch, nil, statusStyle)
	}
}
