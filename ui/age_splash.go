package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// ShowAgeSplash displays a full-screen overlay celebrating an age advancement.
// It auto-dismisses after 8 seconds or on any keypress.
func ShowAgeSplash(app *tview.Application, pages *tview.Pages, oldAge, newAge string) {
	ages := config.AgeByKey()
	newDef := ages[newAge]

	// Generate a pixel art scene for the new age
	dummyBuildings := make(map[string]game.BuildingState)
	for _, bKey := range newDef.UnlockBuildings {
		dummyBuildings[bKey] = game.BuildingState{Count: 3, Unlocked: true}
	}

	img := GenerateMapImage(MapGenConfig{
		Width:       160,
		Height:      80,
		DetailLevel: 1,
		Buildings:   dummyBuildings,
		AgeKey:      newAge,
	})

	// Build the image widget
	mapImage := tview.NewImage()
	mapImage.SetColors(tview.TrueColor)
	mapImage.SetImage(img)

	// Age title overlay
	titleTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	var sb strings.Builder
	fmt.Fprintf(&sb, "\n\n")
	fmt.Fprintf(&sb, "[gold]══════════════════════════════════[-]\n")
	fmt.Fprintf(&sb, "[gold::b]★  %s  ★[-]\n", strings.ToUpper(newDef.Name))
	fmt.Fprintf(&sb, "[white]\"%s\"[-]\n", newDef.Description)
	fmt.Fprintf(&sb, "[gold]══════════════════════════════════[-]\n")
	fmt.Fprintf(&sb, "\n")

	// Show new unlocks
	allBuildings := config.BuildingByKey()
	if len(newDef.UnlockBuildings) > 0 {
		bNames := make([]string, 0, len(newDef.UnlockBuildings))
		for _, bKey := range newDef.UnlockBuildings {
			if def, ok := allBuildings[bKey]; ok {
				bNames = append(bNames, def.Name)
			} else {
				bNames = append(bNames, bKey)
			}
		}
		fmt.Fprintf(&sb, "[cyan]New Buildings:[-] %s\n", strings.Join(bNames, ", "))
	}
	if len(newDef.UnlockResources) > 0 {
		fmt.Fprintf(&sb, "[green]New Resources:[-] %s\n", strings.Join(newDef.UnlockResources, ", "))
	}
	if len(newDef.UnlockVillagers) > 0 {
		fmt.Fprintf(&sb, "[yellow]New Villagers:[-] %s\n", strings.Join(newDef.UnlockVillagers, ", "))
	}

	// Highlight the wonder for this age
	for _, bKey := range newDef.UnlockBuildings {
		if def, ok := allBuildings[bKey]; ok && def.Category == "wonder" {
			fmt.Fprintf(&sb, "\n[gold::b]★ Wonder Unlocked: %s[-]\n", def.Name)
			fmt.Fprintf(&sb, "[white]Build it to unlock +0.5x game speed![-]\n")
			break
		}
	}

	fmt.Fprintf(&sb, "\n[gray]Press any key to continue[-]")
	titleTV.SetText(sb.String())

	// Layout: image behind, text overlay on top via Flex
	overlay := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mapImage, 0, 2, false).
		AddItem(titleTV, 0, 1, false)

	// Dismiss function
	dismissed := false
	dismiss := func() {
		if dismissed {
			return
		}
		dismissed = true
		pages.RemovePage("age_splash")
		pages.SwitchToPage("dashboard")
	}

	// Capture any key to dismiss
	overlay.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		app.QueueUpdateDraw(func() {
			dismiss()
		})
		return nil
	})

	// Auto-dismiss after 8 seconds
	go func() {
		time.Sleep(8 * time.Second)
		app.QueueUpdateDraw(func() {
			dismiss()
		})
	}()

	pages.AddPage("age_splash", overlay, true, true)
	pages.SwitchToPage("age_splash")
}
