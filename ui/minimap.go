package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// MiniMap is a widget that displays a procedural ASCII settlement map
type MiniMap struct {
	box       *tview.Box
	grid      [][]MapCell
	lastHash  uint64
	lastAge   string
	lastTick  int
}

// NewMiniMap creates a new mini-map widget
func NewMiniMap() *MiniMap {
	m := &MiniMap{}
	m.box = tview.NewBox().SetBorder(true).SetTitle(" Map ").SetTitleColor(ColorTitle)
	m.box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// Draw the cached grid
		if m.grid == nil {
			return x, y, width, height
		}
		for gy := 0; gy < len(m.grid) && gy < height; gy++ {
			for gx := 0; gx < len(m.grid[gy]) && gx < width; gx++ {
				cell := m.grid[gy][gx]
				screen.SetContent(x+gx, y+gy, cell.Char, nil, cell.Style)
			}
		}
		return x, y, width, height
	})
	return m
}

// Box returns the underlying tview.Box primitive
func (m *MiniMap) Box() *tview.Box {
	return m.box
}

// UpdateState regenerates the map when buildings/age change
func (m *MiniMap) UpdateState(state game.GameState) {
	// Compute a quick hash of building counts + age to avoid unnecessary regeneration
	h := hashKey(state.Age)
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			h ^= hashKey(k) * uint64(bs.Count)
		}
	}

	needRegen := h != m.lastHash || state.Age != m.lastAge
	tickChanged := state.Tick != m.lastTick

	if !needRegen && !tickChanged {
		return
	}

	_, _, w, ht := m.box.GetInnerRect()
	if w < 5 || ht < 5 {
		return
	}

	m.grid = GenerateMap(MapGenConfig{
		Width:       w,
		Height:      ht,
		DetailLevel: 0,
		Buildings:   state.Buildings,
		AgeKey:      state.Age,
		Tick:        state.Tick,
		TotalPop:    state.Villagers.TotalPop,
	})

	m.lastHash = h
	m.lastAge = state.Age
	m.lastTick = state.Tick
}
