package ui

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// MapTab displays a full-screen procedural pixel settlement map
type MapTab struct {
	root     *tview.Flex
	image    *tview.Image
	titleTV  *tview.TextView
	lastHash uint64
	lastAge  string
}

// NewMapTab creates a new full-screen map tab
func NewMapTab() *MapTab {
	t := &MapTab{}
	t.image = tview.NewImage()
	t.image.SetColors(tview.TrueColor)

	t.titleTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	t.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.titleTV, 1, 0, false).
		AddItem(t.image, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *MapTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the map with current game state
func (t *MapTab) Refresh(state game.GameState) {
	h := hashKey(state.Age)
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			h ^= hashKey(k) * uint64(bs.Count)
		}
	}

	if h == t.lastHash && state.Age == t.lastAge {
		return
	}

	_, _, w, ht := t.image.GetRect()
	if w < 4 || ht < 4 {
		return
	}
	// Higher resolution for full tab: 2x horizontal, 4x vertical (half-block doubled)
	pixW := w * 2
	pixH := ht * 4

	totalBuildings := 0
	for _, bs := range state.Buildings {
		totalBuildings += bs.Count
	}

	img := GenerateMapImage(MapGenConfig{
		Width:       pixW,
		Height:      pixH,
		DetailLevel: 1,
		Buildings:   state.Buildings,
		AgeKey:      state.Age,
	})

	t.image.SetImage(img)
	label := settlementLabel(totalBuildings)
	t.titleTV.SetText(fmt.Sprintf("[gold]── %s ──[-]", label))

	t.lastHash = h
	t.lastAge = state.Age
}
