package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// WondersTab shows the player's collected wonders with pixel art and perks
type WondersTab struct {
	root     *tview.Flex
	listTV   *tview.TextView
	detailTV *tview.TextView
	imgView  *tview.Image
	lastHash uint64
}

// NewWondersTab creates the wonders collection tab
func NewWondersTab() *WondersTab {
	t := &WondersTab{}

	t.listTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.listTV.SetBorder(true).SetTitle(" Wonder Collection ").SetTitleColor(ColorTitle)

	t.detailTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.detailTV.SetBorder(true).SetTitle(" Details ").SetTitleColor(ColorTitle)

	t.imgView = tview.NewImage()
	t.imgView.SetColors(tview.TrueColor)
	t.imgView.SetBorder(true).SetTitle(" Gallery ").SetTitleColor(ColorTitle)

	// Layout: left list | right (image on top, details below)
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.imgView, 0, 1, false).
		AddItem(t.detailTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.listTV, 0, 1, false).
		AddItem(rightPanel, 0, 2, false)

	return t
}

// Root returns the root primitive
func (t *WondersTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the wonders tab
func (t *WondersTab) Refresh(state game.GameState) {
	// Hash wonder state
	var h uint64
	h = hashKey(state.Age)
	for _, w := range getWonderList() {
		if bs, ok := state.Buildings[w.key]; ok {
			if bs.Count > 0 {
				h ^= hashKey(w.key) * 7
			}
			if bs.Unlocked {
				h ^= hashKey(w.key) * 3
			}
		}
	}
	if h == t.lastHash {
		return
	}
	t.lastHash = h

	wonders := getWonderList()
	builtCount := 0
	totalCount := len(wonders)
	maxSpeed := 1.0

	var listSB strings.Builder
	var detailSB strings.Builder
	var lastBuiltWonder *wonderInfo

	for _, w := range wonders {
		built := false
		unlocked := false
		if bs, ok := state.Buildings[w.key]; ok {
			built = bs.Count > 0
			unlocked = bs.Unlocked
		}

		if built {
			builtCount++
			maxSpeed += 0.5
			wCopy := w
			lastBuiltWonder = &wCopy

			// List entry: built
			fmt.Fprintf(&listSB, " [gold]★[-] [green]%s[-]\n", w.name)
			fmt.Fprintf(&listSB, "   [gray]%s[-]\n", w.ageName)
			for _, eff := range w.def.Effects {
				fmt.Fprintf(&listSB, "   %s\n", formatEffect(eff))
			}
			fmt.Fprintf(&listSB, "   [gold]+0.5x speed[-]\n\n")
		} else if unlocked {
			// List entry: available
			fmt.Fprintf(&listSB, " [yellow]○[-] [yellow]%s[-]\n", w.name)
			fmt.Fprintf(&listSB, "   [gray]%s — available[-]\n\n", w.ageName)
		} else {
			// List entry: locked
			fmt.Fprintf(&listSB, " [gray]?[-] [gray]???[-]\n")
			fmt.Fprintf(&listSB, "   [gray]%s — locked[-]\n\n", w.ageName)
		}
	}

	// Summary header
	var headerSB strings.Builder
	fmt.Fprintf(&headerSB, "[gold::b]Wonders: %d / %d[-]\n", builtCount, totalCount)
	fmt.Fprintf(&headerSB, "[cyan]Max Speed: %.1fx[-]\n", maxSpeed)
	fmt.Fprintf(&headerSB, "[gray]Each wonder grants +0.5x game speed[-]\n\n")

	t.listTV.SetText(headerSB.String() + listSB.String())

	// Detail panel: show most recent built wonder's full info
	if lastBuiltWonder != nil {
		fmt.Fprintf(&detailSB, "[gold::b]%s[-]\n", lastBuiltWonder.name)
		fmt.Fprintf(&detailSB, "[gray]%s[-]\n\n", lastBuiltWonder.ageName)
		fmt.Fprintf(&detailSB, "%s\n\n", lastBuiltWonder.def.Description)
		fmt.Fprintf(&detailSB, "[cyan]Effects:[-]\n")
		for _, eff := range lastBuiltWonder.def.Effects {
			fmt.Fprintf(&detailSB, "  %s\n", formatEffect(eff))
		}
		fmt.Fprintf(&detailSB, "  [gold]+0.5x game speed[-]\n")

		// Render the wonder pixel art
		_, _, imgW, imgH := t.imgView.GetInnerRect()
		if imgW > 4 && imgH > 4 {
			pixW := imgW * 2
			pixH := imgH * 4
			img := renderWonderIcon(*lastBuiltWonder, pixW, pixH, true)
			t.imgView.SetImage(img)
		}
	} else {
		detailSB.WriteString("[gray]No wonders built yet.\n\nBuild the wonder for your current age\nto unlock game speed bonuses![-]")
		// Show ghost of first available wonder
		for _, w := range wonders {
			if bs, ok := state.Buildings[w.key]; ok && bs.Unlocked {
				_, _, imgW, imgH := t.imgView.GetInnerRect()
				if imgW > 4 && imgH > 4 {
					pixW := imgW * 2
					pixH := imgH * 4
					img := renderWonderIcon(w, pixW, pixH, false)
					t.imgView.SetImage(img)
				}
				break
			}
		}
	}

	t.detailTV.SetText(detailSB.String())
}
