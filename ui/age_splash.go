package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// ShowAgeSplash is the simplified variant of ShowAgeSplashFull for callers that
// don't have an AgeAdvanceSummary or EpochEventRecord available.
func ShowAgeSplash(om *OverlayManager, oldAge, newAge string) {
	ShowAgeSplashFull(om, oldAge, newAge, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{})
}

// ShowAgeSplashFull puts up the arrival screen for an advance into newAge
// (arrival.go): the moment first, the age's name struck in iron over the
// player's town, then what the age opens, the upgrades on offer and the
// epoch's event. It moves on from the first to the second by itself or on
// any key, and closes on a key or after twenty seconds.
func ShowAgeSplashFull(om *OverlayManager, oldAge, newAge string,
	summary game.AgeAdvanceSummary, epochChanged bool, epochEvent game.EpochEventRecord) {
	a := newArrival(om, oldAge, newAge, summary, epochChanged, epochEvent)
	// With the engine at hand the screen has the game's state from its first
	// frame; without, the next refresh brings it (OverlayManager.Refresh).
	if om.engine != nil {
		cur := om.engine.GetState()
		a.feed(om.seen, &cur)
	}

	// Remove any currently active overlay first, then add the screen. It
	// holds the keyboard itself, so a re-focus of the page stack (a window
	// closing over it) lands on something that handles keys.
	om.dropArrival() // an arrival still up gives way to this one: they do not stack
	if om.active != "" {
		om.pages.RemovePage(om.active)
	}
	om.pages.AddPage(arrivalPageName, a, true, true)
	om.active, om.arrival, om.focus = arrivalPageName, a, a
	om.app.SetFocus(a)
	a.begin()
}

// splashKind says what a line of the arrival's text is, so the screen can
// lay each kind out in its place (arrival_page.go).
type splashKind uint8

const (
	skBlank       splashKind = iota
	skRule                   // the rules round the heading
	skTitle                  // the age's name
	skDesc                   // its description
	skQuip                   // its quip
	skBody                   // what the age opens: a label and a list
	skWonder                 // the wonder, and how to raise it
	skSection                // a heading inside the text
	skUpgradeHelp            // how to upgrade
	skUpgradeRow             // one upgrade on offer
	skLegacy                 // one building now legacy
	skEpoch                  // the new epoch and its event
	skPrompt                 // how to go on
)

// splashLine is one line of the arrival's text: as tview markup, and what
// kind of line it is.
type splashLine struct {
	text string
	kind splashKind
}

// buildAgeSplashText assembles the full text of an age advance: the age's
// name, description and quip, what it opens, the upgrades on offer and the
// epoch's event. It is the arrival screen's words, every one of them; the
// screen lays the same lines out (ageSplashLines). Pure function: same
// inputs → same string, no side effects.
func buildAgeSplashText(newAge string, summary game.AgeAdvanceSummary,
	epochChanged bool, epochEvent game.EpochEventRecord) string {
	lines := ageSplashLines(newAge, summary, epochChanged, epochEvent)
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.text
	}
	return strings.Join(out, "\n")
}

// ageSplashLines is the text of an age advance, line by line.
func ageSplashLines(newAge string, summary game.AgeAdvanceSummary,
	epochChanged bool, epochEvent game.EpochEventRecord) []splashLine {
	ages := config.AgeByKey()
	newDef := ages[newAge]

	var lines []splashLine
	add := func(kind splashKind, format string, args ...any) {
		lines = append(lines, splashLine{text: fmt.Sprintf(format, args...), kind: kind})
	}
	blank := func() { add(skBlank, "") }

	blank()
	blank()
	add(skRule, "[gold]══════════════════════════════════[-]")
	add(skTitle, "[gold::b]★  %s  ★[-]", strings.ToUpper(newDef.Name))
	add(skDesc, "[white]\"%s\"[-]", newDef.Description)
	if newDef.Quip != "" {
		add(skQuip, "[gray::i]%s[-:-:-]", newDef.Quip)
	}
	add(skRule, "[gold]══════════════════════════════════[-]")
	blank()

	// Show new unlocks
	allBuildings := config.BuildingByKey()
	techs := config.TechByKey()
	if len(newDef.UnlockBuildings) > 0 {
		bNames := make([]string, 0, len(newDef.UnlockBuildings))
		var later []string // buildings a tech of this age opens later
		for _, bKey := range newDef.UnlockBuildings {
			def, ok := allBuildings[bKey]
			switch {
			case !ok:
				bNames = append(bNames, bKey)
			case def.RequiredTech != "" && def.Category != "wonder":
				// (A wonder is open for banking from the first tick; its
				// keystone is named with the wonder below.)
				later = append(later, fmt.Sprintf("%s (%s)", def.Name, techs[def.RequiredTech].Name))
			default:
				bNames = append(bNames, def.Name)
			}
		}
		add(skBody, "[cyan]New buildings:[-] %s", strings.Join(bNames, ", "))
		if len(later) > 0 {
			add(skBody, "[cyan]Opened later by research:[-] %s", strings.Join(later, ", "))
		}
	}
	if len(newDef.UnlockResources) > 0 {
		rNames := make([]string, 0, len(newDef.UnlockResources))
		for _, rKey := range newDef.UnlockResources {
			rNames = append(rNames, game.ResourceName(rKey))
		}
		add(skBody, "[green]New resources:[-] %s", strings.Join(rNames, ", "))
	}
	if len(newDef.UnlockVillagers) > 0 {
		wNames := make([]string, 0, len(newDef.UnlockVillagers))
		for _, wKey := range newDef.UnlockVillagers {
			wNames = append(wNames, textfmt.Capitalize(strings.ReplaceAll(wKey, "_", " ")))
		}
		add(skBody, "[yellow]New worker types:[-] %s", strings.Join(wNames, ", "))
	}

	// Highlight the wonder for this age
	for _, bKey := range newDef.UnlockBuildings {
		if def, ok := allBuildings[bKey]; ok && def.Category == "wonder" {
			blank()
			add(skWonder, "[gold::b]★ Wonder unlocked: %s[-]", def.Name)
			if tech, ok := techs[def.RequiredTech]; ok {
				add(skWonder, "[white]Bank its cost and research %s, its keystone, then build it.[-]", tech.Name)
			} else {
				add(skWonder, "[white]Bank its cost, then build it.[-]")
			}
			break
		}
	}

	// Upgrade summary. Advancing only offers these upgrades (the buildings
	// stay as they are until the player runs `upgrade`), so say that.
	if len(summary.BuildingsTransformed) > 0 || len(summary.BuildingsLegacy) > 0 {
		blank()
		if len(summary.BuildingsTransformed) > 0 {
			add(skSection, "[yellow]── Upgrades available ──[-]")
			add(skUpgradeHelp, "  Type [cyan]upgrade[-] to see the costs, or [cyan]upgrade <building>[-] to upgrade one.")
			// Every line of the text is centered on its own, so the
			// counts only line up when the rows are all one width: each
			// label is padded to the longest, each count to the widest.
			labels := make([]string, len(summary.BuildingsTransformed))
			counts := make([]string, len(summary.BuildingsTransformed))
			labelW, countW := 0, 0
			for i, t := range summary.BuildingsTransformed {
				oldName := t.OldName
				if oldName == "" {
					oldName = strings.ReplaceAll(t.OldKey, "_", " ")
				}
				newName := t.NewName
				if newName == "" {
					newName = strings.ReplaceAll(t.NewKey, "_", " ")
				}
				labels[i] = fmt.Sprintf("%s → %s", oldName, newName)
				counts[i] = fmt.Sprintf("x%d", t.Count)
				labelW, countW = max(labelW, runeLen(labels[i])), max(countW, runeLen(counts[i]))
			}
			for i := range labels {
				add(skUpgradeRow, "%s%s  %s%s", labels[i], strings.Repeat(" ", labelW-runeLen(labels[i])),
					counts[i], strings.Repeat(" ", countW-runeLen(counts[i])))
			}
		}
		if len(summary.BuildingsLegacy) > 0 {
			add(skSection, "[yellow]── Legacy buildings ──[-]")
			for _, legKey := range summary.BuildingsLegacy {
				legName := legKey
				if def, ok := allBuildings[legKey]; ok {
					legName = def.Name
				} else {
					legName = strings.ReplaceAll(legKey, "_", " ")
				}
				add(skLegacy, "  [gray]%s[-]", legName)
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

		blank()
		add(skEpoch, "[gold]══ New epoch ══[-]")
		add(skEpoch, "[%s]%s %s[-]", epochColor, epochIcon, epochEvent.EpochName)
		blank()
		add(skEpoch, "[%s]%s[-]", eventColor, epochEvent.EventName)
		if flavorText != "" {
			add(skEpoch, "[white]\"%s\"[-]", flavorText)
		}
	}

	blank()
	add(skPrompt, "[gray]Press any key to continue.[-]")
	return lines
}
