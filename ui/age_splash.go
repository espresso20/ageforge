package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// ShowAgeSplash is the simplified variant of ShowAgeSplashFull for callers that
// don't have an AgeAdvanceSummary or EpochEventRecord available.
// It auto-dismisses after 20 seconds or on any keypress.
func ShowAgeSplash(app *tview.Application, pages *tview.Pages, oldAge, newAge string) {
	ShowAgeSplashFull(app, pages, oldAge, newAge, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{})
}

// ShowAgeSplashFull is the full variant of ShowAgeSplash that also displays the
// transformation summary and optional epoch event reveal.
func ShowAgeSplashFull(app *tview.Application, pages *tview.Pages, oldAge, newAge string,
	summary game.AgeAdvanceSummary, epochChanged bool, epochEvent game.EpochEventRecord) {
	ages := config.AgeByKey()
	newDef := ages[newAge]

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
		fmt.Fprintf(&sb, "[yellow]New Workers:[-] %s\n", strings.Join(newDef.UnlockVillagers, ", "))
	}

	// Highlight the wonder for this age
	for _, bKey := range newDef.UnlockBuildings {
		if def, ok := allBuildings[bKey]; ok && def.Category == "wonder" {
			fmt.Fprintf(&sb, "\n[gold::b]★ Wonder Unlocked: %s[-]\n", def.Name)
			fmt.Fprintf(&sb, "[white]Build it to unlock +0.5x game speed![-]\n")
			break
		}
	}

	// Transformation summary
	if len(summary.BuildingsTransformed) > 0 || len(summary.BuildingsLegacy) > 0 {
		fmt.Fprintf(&sb, "\n")
		if len(summary.BuildingsTransformed) > 0 {
			fmt.Fprintf(&sb, "[yellow]── Buildings Transformed ──[-]\n")
			for _, t := range summary.BuildingsTransformed {
				oldName := t.OldName
				if oldName == "" {
					oldName = strings.ReplaceAll(t.OldKey, "_", " ")
				}
				newName := t.NewName
				if newName == "" {
					newName = strings.ReplaceAll(t.NewKey, "_", " ")
				}
				fmt.Fprintf(&sb, "  %-28s x%d\n",
					fmt.Sprintf("%s → %s", oldName, newName), t.Count)
			}
		}
		if len(summary.BuildingsLegacy) > 0 {
			fmt.Fprintf(&sb, "[yellow]── Legacy Buildings ──[-]\n")
			for _, legKey := range summary.BuildingsLegacy {
				legName := legKey
				if def, ok := allBuildings[legKey]; ok {
					legName = def.Name
				} else {
					legName = strings.ReplaceAll(legKey, "_", " ")
				}
				fmt.Fprintf(&sb, "  [gray]%s[-]\n", legName)
			}
		}
	}

	// Epoch transition reveal
	if epochChanged && epochEvent.EpochKey != "" {
		epochColor := "cyan"
		eventColor := "green"
		switch epochEvent.EventType {
		case "bad_challenging":
			epochColor = "red"
			eventColor = "red"
		case "catastrophe":
			epochColor = "red"
			eventColor = "red"
		case "good_legendary":
			epochColor = "gold"
			eventColor = "gold"
		case "good_major":
			eventColor = "cyan"
		}

		// Look up the epoch icon
		epochIcon := "✦"
		if ep, ok := config.EpochByKey()[epochEvent.EpochKey]; ok {
			epochIcon = ep.Icon
		}

		// Look up the flavor text from config
		flavorText := ""
		if evDef, ok := config.EpochEventByKey()[epochEvent.EventKey]; ok {
			flavorText = evDef.FlavorText
		}

		fmt.Fprintf(&sb, "\n[gold]══ EPOCH TRANSITION ══[-]\n")
		fmt.Fprintf(&sb, "[%s]%s %s[-]\n", epochColor, epochIcon, epochEvent.EpochName)
		fmt.Fprintf(&sb, "\n[%s]%s[-]\n", eventColor, epochEvent.EventName)
		if flavorText != "" {
			fmt.Fprintf(&sb, "[white]\"%s\"[-]\n", flavorText)
		}
	}

	fmt.Fprintf(&sb, "\n[gray]Press any key to continue[-]")
	titleTV.SetText(sb.String())

	// Layout: text-only flex
	overlay := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titleTV, 0, 1, false)

	// dismissed guards against double-dismiss (both the timer goroutine and a
	// keypress can fire simultaneously; we only want to remove the page once).
	dismissed := false
	dismiss := func() {
		if dismissed {
			return
		}
		dismissed = true
		pages.RemovePage("age_splash")
		pages.SwitchToPage("dashboard")
	}

	// SetInputCapture on titleTV (the focused child) so keypresses actually reach
	// the capture handler. QueueUpdateDraw is required because pages.RemovePage
	// modifies the layout and must run on the main goroutine.
	titleTV.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		app.QueueUpdateDraw(func() {
			dismiss()
		})
		return nil
	})

	// Auto-dismiss after 20 seconds
	go func() {
		time.Sleep(20 * time.Second)
		app.QueueUpdateDraw(func() {
			dismiss()
		})
	}()

	pages.AddPage("age_splash", overlay, true, true)
	pages.SwitchToPage("age_splash")
	app.SetFocus(titleTV)
}
