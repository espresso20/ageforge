package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// MapTab displays a full-screen procedural pixel settlement map
type MapTab struct {
	root         *tview.Flex
	image        *tview.Image
	titleTV      *tview.TextView
	lastHash     uint64
	lastAge      string
	pendingState *game.GameState
	pendingHash  uint64
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

	totalBuildings := 0
	for _, bs := range state.Buildings {
		totalBuildings += bs.Count
	}
	label := settlementLabel(totalBuildings)
	t.titleTV.SetText(fmt.Sprintf("[gold]── %s ──[-]", label))

	t.pendingState = &state
	t.pendingHash = h

	// Defer image generation until the widget is drawn and its dimensions are known.
	// SetDrawFunc fires inside Image.Draw(), before pixels are rendered, so the newly
	// generated image is available for the current frame.
	t.image.SetDrawFunc(func(screen tcell.Screen, x, y, w, ht int) (int, int, int, int) {
		if t.pendingState != nil && w >= 4 && ht >= 4 {
			s := *t.pendingState
			ph := t.pendingHash
			t.pendingState = nil

			// Correct pixel dimensions for tview TrueColor half-block rendering:
			// each terminal cell = 1 pixel wide × 2 pixels tall
			pixW := w
			pixH := ht * 2

			img := GenerateMapImage(MapGenConfig{
				Width:       pixW,
				Height:      pixH,
				DetailLevel: 1,
				Buildings:   s.Buildings,
				AgeKey:      s.Age,
			})
			t.image.SetImage(img)
			t.lastHash = ph
			t.lastAge = s.Age
		}
		return x, y, w, ht
	})
}
