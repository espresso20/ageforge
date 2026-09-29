package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// ShowAgeSplash is the simplified variant of ShowAgeSplashFull for callers that
// don't have an AgeAdvanceSummary or EpochEventRecord available.
// It auto-dismisses after 20 seconds or on any keypress.
func ShowAgeSplash(om *OverlayManager, oldAge, newAge string) {
	ShowAgeSplashFull(om, oldAge, newAge, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{})
}

// ShowAgeSplashFull is the full variant of ShowAgeSplash that also displays the
// transformation summary and optional epoch event reveal.
func ShowAgeSplashFull(om *OverlayManager, oldAge, newAge string,
	summary game.AgeAdvanceSummary, epochChanged bool, epochEvent game.EpochEventRecord) {
	// Age title overlay
	titleTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	titleTV.SetText(safeTags(buildAgeSplashText(newAge, summary, epochChanged, epochEvent)))

	// Layout: text-only flex with focusable=true so the overlay itself receives input
	overlay := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titleTV, 0, 1, true)

	// dismissed guards against double-dismiss (both the timer goroutine and a
	// keypress can fire simultaneously; we only want to remove the page once).
	dismissed := false
	dismiss := func() {
		if dismissed {
			return
		}
		dismissed = true
		om.Hide()
	}

	// SetInputCapture on the overlay Flex so all keypresses trigger dismiss.
	// A container's capture fires for keys routed to any focused child, and
	// keeping titleTV as the focusable item means a Pages re-focus (e.g. when
	// a page above is removed) lands on titleTV inside this Flex rather than
	// on a bare primitive with no key handler.
	overlay.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		dismiss()
		return nil
	})

	// Auto-dismiss after 20 seconds
	go func() {
		time.Sleep(20 * time.Second)
		om.app.QueueUpdateDraw(func() {
			dismiss()
		})
	}()

	// Remove any currently active overlay first, then add the splash
	if om.active != "" {
		om.pages.RemovePage(om.active)
	}
	om.pages.AddPage("age_splash", overlay, true, true)
	om.active = "age_splash"
	om.app.SetFocus(overlay)
}

// buildAgeSplashText assembles the full splash body string for an age advance.
// It is split out from ShowAgeSplashFull so it can be unit-tested without a live
// tview app. Pure function: same inputs → same string, no side effects.
func buildAgeSplashText(newAge string, summary game.AgeAdvanceSummary,
	epochChanged bool, epochEvent game.EpochEventRecord) string {
	ages := config.AgeByKey()
	newDef := ages[newAge]

	var sb strings.Builder
	fmt.Fprintf(&sb, "\n\n")
	fmt.Fprintf(&sb, "[gold]══════════════════════════════════[-]\n")
	fmt.Fprintf(&sb, "[gold::b]★  %s  ★[-]\n", strings.ToUpper(newDef.Name))
	fmt.Fprintf(&sb, "[white]\"%s\"[-]\n", newDef.Description)
	if newDef.Quip != "" {
		fmt.Fprintf(&sb, "[gray::i]%s[-:-:-]\n", newDef.Quip)
	}
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
		fmt.Fprintf(&sb, "[cyan]New buildings:[-] %s\n", strings.Join(bNames, ", "))
	}
	if len(newDef.UnlockResources) > 0 {
		rNames := make([]string, 0, len(newDef.UnlockResources))
		for _, rKey := range newDef.UnlockResources {
			rNames = append(rNames, game.ResourceName(rKey))
		}
		fmt.Fprintf(&sb, "[green]New resources:[-] %s\n", strings.Join(rNames, ", "))
	}
	if len(newDef.UnlockVillagers) > 0 {
		wNames := make([]string, 0, len(newDef.UnlockVillagers))
		for _, wKey := range newDef.UnlockVillagers {
			wNames = append(wNames, textfmt.Capitalize(strings.ReplaceAll(wKey, "_", " ")))
		}
		fmt.Fprintf(&sb, "[yellow]New worker types:[-] %s\n", strings.Join(wNames, ", "))
	}

	// Highlight the wonder for this age
	for _, bKey := range newDef.UnlockBuildings {
		if def, ok := allBuildings[bKey]; ok && def.Category == "wonder" {
			fmt.Fprintf(&sb, "\n[gold::b]★ Wonder unlocked: %s[-]\n", def.Name)
			fmt.Fprintf(&sb, "[white]Bank its cost, then build it. It raises the speed cap by 0.5x.[-]\n")
			break
		}
	}

	// Transformation summary
	if len(summary.BuildingsTransformed) > 0 || len(summary.BuildingsLegacy) > 0 {
		fmt.Fprintf(&sb, "\n")
		if len(summary.BuildingsTransformed) > 0 {
			fmt.Fprintf(&sb, "[yellow]── Buildings upgraded ──[-]\n")
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
			fmt.Fprintf(&sb, "[yellow]── Legacy buildings ──[-]\n")
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

		fmt.Fprintf(&sb, "\n[gold]══ New epoch ══[-]\n")
		fmt.Fprintf(&sb, "[%s]%s %s[-]\n", epochColor, epochIcon, epochEvent.EpochName)
		fmt.Fprintf(&sb, "\n[%s]%s[-]\n", eventColor, epochEvent.EventName)
		if flavorText != "" {
			fmt.Fprintf(&sb, "[white]\"%s\"[-]\n", flavorText)
		}
	}

	fmt.Fprintf(&sb, "\n[gray]Press any key to continue.[-]")
	return sb.String()
}
