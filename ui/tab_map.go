package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// MapTab displays a full-screen procedural settlement map
type MapTab struct {
	box            *tview.Box
	grid           [][]MapCell
	settlementName string
	lastHash       uint64
	lastAge        string
	lastTick       int
}

// NewMapTab creates a new full-screen map tab
func NewMapTab() *MapTab {
	t := &MapTab{}
	t.box = tview.NewBox()
	t.box.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if t.grid == nil {
			return x, y, width, height
		}

		// Draw settlement name centered at top
		if t.settlementName != "" {
			labelX := x + (width-len(t.settlementName))/2
			for i, ch := range t.settlementName {
				screen.SetContent(labelX+i, y, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGold).Bold(true))
			}
		}

		// Draw grid starting from row 1 (below label)
		startY := y + 1
		for gy := 0; gy < len(t.grid) && gy < height-1; gy++ {
			for gx := 0; gx < len(t.grid[gy]) && gx < width; gx++ {
				cell := t.grid[gy][gx]
				screen.SetContent(x+gx, startY+gy, cell.Char, nil, cell.Style)
			}
		}
		return x, y, width, height
	})
	return t
}

// Root returns the root primitive
func (t *MapTab) Root() tview.Primitive {
	return t.box
}

// Refresh updates the map with current game state
func (t *MapTab) Refresh(state game.GameState) {
	// Compute hash
	h := hashKey(state.Age)
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			h ^= hashKey(k) * uint64(bs.Count)
		}
	}

	needRegen := h != t.lastHash || state.Age != t.lastAge
	tickChanged := state.Tick != t.lastTick

	if !needRegen && !tickChanged {
		return
	}

	_, _, w, ht := t.box.GetRect()
	if w < 5 || ht < 5 {
		return
	}

	// Count buildings for settlement label
	totalBuildings := 0
	for _, bs := range state.Buildings {
		totalBuildings += bs.Count
	}

	t.grid = GenerateMap(MapGenConfig{
		Width:       w,
		Height:      ht - 1,
		DetailLevel: 1,
		Buildings:   state.Buildings,
		AgeKey:      state.Age,
		Tick:        state.Tick,
		TotalPop:    state.Villagers.TotalPop,
	})

	t.settlementName = settlementLabel(totalBuildings)
	t.lastHash = h
	t.lastAge = state.Age
	t.lastTick = state.Tick
}
