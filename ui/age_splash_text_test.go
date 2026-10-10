package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// legacyAgeSplashText is the age splash's text as it was built before the
// arrival screen, word for word: the yardstick ageSplashLines is held to,
// so that no line the old splash showed is lost or changed.
func legacyAgeSplashText(newAge string, summary game.AgeAdvanceSummary,
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
		fmt.Fprintf(&sb, "[cyan]New buildings:[-] %s\n", strings.Join(bNames, ", "))
		if len(later) > 0 {
			fmt.Fprintf(&sb, "[cyan]Opened later by research:[-] %s\n", strings.Join(later, ", "))
		}
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
			if tech, ok := techs[def.RequiredTech]; ok {
				fmt.Fprintf(&sb, "[white]Bank its cost and research %s, its keystone, then build it.[-]\n", tech.Name)
			} else {
				sb.WriteString("[white]Bank its cost, then build it.[-]\n")
			}
			break
		}
	}

	// Upgrade summary. Advancing only offers these upgrades (the buildings
	// stay as they are until the player runs `upgrade`), so say that.
	if len(summary.BuildingsTransformed) > 0 || len(summary.BuildingsLegacy) > 0 {
		fmt.Fprintf(&sb, "\n")
		if len(summary.BuildingsTransformed) > 0 {
			fmt.Fprintf(&sb, "[yellow]── Upgrades available ──[-]\n")
			fmt.Fprintf(&sb, "  Type [cyan]upgrade[-] to see the costs, or [cyan]upgrade <building>[-] to upgrade one.\n")
			// Every line of the splash is centered on its own, so the
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
				fmt.Fprintf(&sb, "%s%s  %s%s\n", labels[i], strings.Repeat(" ", labelW-runeLen(labels[i])),
					counts[i], strings.Repeat(" ", countW-runeLen(counts[i])))
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

// splashCases is every age with and without upgrades on offer, and every
// age that opens an epoch with each kind of event.
func splashCases() []struct {
	age     string
	summary game.AgeAdvanceSummary
	epoch   bool
	event   game.EpochEventRecord
} {
	var out []struct {
		age     string
		summary game.AgeAdvanceSummary
		epoch   bool
		event   game.EpochEventRecord
	}
	full := game.AgeAdvanceSummary{
		BuildingsTransformed: []game.BuildingTransform{
			{OldName: "Elders' Hall", NewName: "Scriptorium", Count: 5},
			{OldKey: "long_house", NewKey: "house", Count: 15},
			{OldName: "Nuclear Extraction Plant", NewName: "Uranium Processing Works", Count: 120},
		},
		BuildingsLegacy: []string{"stash", "stone_camp", "no_such_building"},
	}
	events := config.EpochEventByKey()
	for _, a := range config.Ages() {
		out = append(out, struct {
			age     string
			summary game.AgeAdvanceSummary
			epoch   bool
			event   game.EpochEventRecord
		}{a.Key, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{}})
		out = append(out, struct {
			age     string
			summary game.AgeAdvanceSummary
			epoch   bool
			event   game.EpochEventRecord
		}{a.Key, full, false, game.EpochEventRecord{}})
		era := config.EpochForAge(a.Key)
		ep := config.EpochByKey()[era]
		if len(ep.Ages) == 0 || ep.Ages[0] != a.Key {
			continue
		}
		for _, kind := range []string{"good_minor", "good_major", "good_legendary", "bad_challenging", "catastrophe"} {
			rec := game.EpochEventRecord{EpochKey: era, EpochName: ep.Name, EventKey: "none", EventName: "An event", EventType: kind}
			for key, ev := range events {
				if ev.Type == kind {
					rec.EventKey, rec.EventName = key, ev.Name
					break
				}
			}
			out = append(out, struct {
				age     string
				summary game.AgeAdvanceSummary
				epoch   bool
				event   game.EpochEventRecord
			}{a.Key, full, true, rec})
		}
	}
	return out
}

// TestAgeSplashTextIsWhatItWas: the arrival's text, built line by line for
// the new screen, is the old splash's text to the letter, for every age,
// with and without upgrades, and with every kind of epoch event.
func TestAgeSplashTextIsWhatItWas(t *testing.T) {
	cases := splashCases()
	if len(cases) < 2*len(config.Ages())+5 {
		t.Fatalf("only %d cases", len(cases))
	}
	for _, c := range cases {
		got := buildAgeSplashText(c.age, c.summary, c.epoch, c.event)
		want := legacyAgeSplashText(c.age, c.summary, c.epoch, c.event)
		if got != want {
			t.Errorf("%s (epoch %v, %s): the text changed.\n--- was ---\n%s\n--- is ---\n%s", c.age, c.epoch, c.event.EventType, want, got)
		}
		for _, l := range ageSplashLines(c.age, c.summary, c.epoch, c.event) {
			if strings.Contains(l.text, "\n") {
				t.Errorf("%s: a line holds a line break: %q", c.age, l.text)
			}
			if (l.kind == skBlank) != (l.text == "") {
				t.Errorf("%s: line %q is of kind %d", c.age, l.text, l.kind)
			}
		}
	}
}
