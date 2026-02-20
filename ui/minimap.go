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

// UpdateState regenerates the map when buildings or age change
func (m *MiniMap) UpdateState(state game.GameState) {
	h := hashKey(state.Age)
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			h ^= hashKey(k) * uint64(bs.Count)
		}
	}

	if h == m.lastHash && state.Age == m.lastAge {
		return
	}

	_, _, w, ht := m.image.GetInnerRect()
	if w < 4 || ht < 4 {
		return
	}
	// Half-blocks give 2x vertical res; bump horizontal 2x for more detail
	pixW := w * 2
	pixH := ht * 4

	img := GenerateMapImage(MapGenConfig{
		Width:       pixW,
		Height:      pixH,
		DetailLevel: 0,
		Buildings:   state.Buildings,
		AgeKey:      state.Age,
	})

	m.image.SetImage(img)
	m.lastHash = h
	m.lastAge = state.Age
}
