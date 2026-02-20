package ui

import (
	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// MiniMap is a widget that displays a procedural pixel settlement map
type MiniMap struct {
	image    *tview.Image
	lastHash uint64
	lastAge  string
	lastTick int
}

// NewMiniMap creates a new mini-map widget
func NewMiniMap() *MiniMap {
	m := &MiniMap{}
	m.image = tview.NewImage()
	m.image.SetBorder(true).SetTitle(" Map ").SetTitleColor(ColorTitle)
	m.image.SetColors(tview.TrueColor)
	return m
}

// Primitive returns the underlying tview primitive
func (m *MiniMap) Primitive() tview.Primitive {
	return m.image
}

// UpdateState regenerates the map when buildings/age change or tick advances
func (m *MiniMap) UpdateState(state game.GameState) {
	// Quick hash to avoid unnecessary regen
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

	// Use inner rect to determine pixel dimensions
	// Half-blocks give 2x vertical resolution, so pixels = cols x (rows*2)
	_, _, w, ht := m.image.GetInnerRect()
	if w < 4 || ht < 4 {
		return
	}
	pixW := w
	pixH := ht * 2 // half-block chars double vertical resolution

	img := GenerateMapImage(MapGenConfig{
		Width:       pixW,
		Height:      pixH,
		DetailLevel: 0,
		Buildings:   state.Buildings,
		AgeKey:      state.Age,
		Tick:        state.Tick,
		TotalPop:    state.Villagers.TotalPop,
	})

	m.image.SetImage(img)

	m.lastHash = h
	m.lastAge = state.Age
	m.lastTick = state.Tick
}
