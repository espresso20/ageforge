package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// epochProvider generates the epoch overlay text from the current game state.
// Shows current epoch status, epoch history, legacy bonuses, and civilization log.
func epochProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	sb.WriteString("[gold]═══ Epoch ═══[-]\n\n")

	epochProviderCurrentEpoch(&sb, state)
	sb.WriteString("\n")
	epochProviderHistory(&sb, state)
	sb.WriteString("\n")
	epochProviderHarbingers(&sb, state)
	sb.WriteString("\n")
	epochProviderLegacyBonuses(&sb, state)
	sb.WriteString("\n")
	epochProviderCivilizationLog(&sb, state)

	return sb.String()
}

// epochProviderCurrentEpoch renders the current epoch status section.
func epochProviderCurrentEpoch(sb *strings.Builder, state game.GameState) {
	sb.WriteString("[gold]── Current epoch ──[-]\n\n")

	if state.EpochKey == "" {
		sb.WriteString(" [gray]No epoch data yet.[-]\n")
		return
	}

	epochs := config.EpochByKey()
	ep, ok := epochs[state.EpochKey]
	if !ok {
		fmt.Fprintf(sb, " [gold]%s %s[-]\n", state.EpochIcon, state.EpochName)
	} else {
		// Ages the player cannot see named yet are counted, not named.
		sight := game.SightOf(&state)
		ageNames := make([]string, 0, len(ep.Ages))
		unseen := 0
		for _, a := range ep.Ages {
			if sight.Age(a) {
				ageNames = append(ageNames, game.AgeName(a))
			} else {
				unseen++
			}
		}
		if unseen > 0 {
			ageNames = append(ageNames, fmt.Sprintf("%d more to come", unseen))
		}
		fmt.Fprintf(sb, " %s%s %s[-]   Ages: %s\n",
			theme.NameTag(ep.Color), ep.Icon, ep.Name,
			strings.Join(ageNames, " · "))
		fmt.Fprintf(sb, " [gray]Primary resource: %s   Energy: %s[-]\n",
			game.ResourceName(ep.PrimaryResource), game.ResourceName(ep.EnergyResource))
	}

	sb.WriteString("\n")

	// Current epoch event (the transition's; a catastrophe is shown below)
	currentEvent := findEpochEvent(state.EpochEventHistory, state.EpochKey)
	if currentEvent == nil {
		sb.WriteString(" Epoch event: [gray]No event this epoch[-]\n")
	} else {
		evColor := epochEventColor(currentEvent.EventType)
		if currentEvent.EventType == "catastrophe" {
			fmt.Fprintf(sb, " Epoch event: [%s]%s[-]   [red](catastrophe)[-]\n",
				evColor, currentEvent.EventName)
		} else {
			evDefs := config.EpochEventByKey()
			flavorStr := ""
			durationStr := ""
			if evDef, found := evDefs[currentEvent.EventKey]; found {
				if evDef.FlavorText != "" {
					flavorStr = evDef.FlavorText
				}
				if evDef.Duration > 0 {
					durationStr = fmt.Sprintf("   %s duration", formatTicks(evDef.Duration, state))
				}
			}
			fmt.Fprintf(sb, " Epoch event: [%s]%s[-]%s\n", evColor, currentEvent.EventName, durationStr)
			if flavorStr != "" {
				fmt.Fprintf(sb, " [gray]  %s[-]\n", flavorStr)
			}
		}
	}

	sb.WriteString("\n")

	// Catastrophe status. Only what the player has seen: a doom not yet
	// foretold reads the same as none (the no-leak rule, game/fate.go).
	o := state.CatastropheOutlook
	switch {
	case state.PendingCatastrophe != "":
		sb.WriteString(" Catastrophe: [red]pending. Type catastrophe to choose Endure or Succumb.[-]\n")
		sb.WriteString(" [gray]  Advancing and prestige are blocked until you decide.[-]\n")
	case state.LastPassage.Pending:
		sb.WriteString(" Catastrophe: [red]the Last Passage. Type catastrophe to choose Endure or Succumb.[-]\n")
		sb.WriteString(" [gray]  Prestige waits until you decide; nothing else does.[-]\n")
	case !config.CatastropheAllowed(state.EpochKey):
		fmt.Fprintf(sb, " Catastrophe: %s\n", theme.Paint(theme.RoleDim, "none in the "+currentEraName(state)))
	default:
		if r := latestCatastrophe(state.EpochEventHistory, state.EpochKey); r != nil {
			fmt.Fprintf(sb, " Catastrophe: %s\n", catastropheOutcomeLabel(r.Outcome))
		} else if r := eraDoomRecord(state); r != nil && r.Outcome == game.HarbingerOutcomeSpared {
			sb.WriteString(" Catastrophe: [green]spared[-] [gray](nothing more will strike this era)[-]\n")
		} else if o.Warned {
			sb.WriteString(" Catastrophe: [yellow]foretold[-]\n")
		} else {
			sb.WriteString(" Catastrophe: [gray]none so far[-]\n")
		}
	}

	// The outlook (same wording the `catastrophe` command uses: a figure from
	// the Industrial Age on, a severity before it).
	if o.Warned {
		fmt.Fprintf(sb, " Doom foretold: [yellow]%s[-] [gray](more faith, lower odds)[-]\n", doomWarningText(state))
	}
	switch {
	case o.Possible && o.Passage == game.PassagePrestige:
		fmt.Fprintf(sb, " Next passage (prestige, the Last Passage): [yellow]%s[-] [gray](more faith, lower odds)[-]\n",
			outlookRiskText(state))
	case o.Possible && !o.Warned && o.Passage == game.PassageEpoch:
		sb.WriteString(" Outlook: [gray]no harbinger has come. Quiet, for now.[-]\n")
	}

	// Harbinger status.
	if h := state.Harbinger; h != nil {
		status := "no word of when"
		if h.WhenText != "" {
			status = "doom " + h.WhenText
		}
		switch {
		case h.PassageCame:
			status = "the Last Passage has come"
		case h.LastPassage && h.Invited:
			status = "invited, the Last Passage will come at prestige"
		case h.LastPassage:
			status = "waiting for your prestige"
		case h.Invited:
			status = "invited, the catastrophe will come"
		}
		fmt.Fprintf(sb, " Harbinger: [warning]%s is here[-] [gray](%s; appease %d/%d, brace %d/%d; type 'harbinger')[-]\n",
			capFirstUI(h.Name), status, h.AppeaseLevel, game.HarbingerMaxAppease, h.BraceLevel, game.HarbingerMaxBrace)
	} else if r := latestHarbinger(state.HarbingerHistory, state.EpochKey); r != nil {
		fmt.Fprintf(sb, " Harbinger: %s\n", harbingerRecordText(*r))
	}
}

// epochProviderHarbingers lists this run's resolved harbingers, oldest first.
func epochProviderHarbingers(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Harbingers ──[-]\n")
	if len(state.HarbingerHistory) == 0 {
		sb.WriteString("   [gray]None have come and gone this run.[-]\n")
		return
	}
	for _, r := range state.HarbingerHistory {
		fmt.Fprintf(sb, "   · %s\n", harbingerRecordText(r))
	}
}

// latestHarbinger returns the newest harbinger record of epochKey's doom, or
// nil. (Saves from before fates held records keyed to the era the passage led
// into; those match too, which is what they meant.)
func latestHarbinger(history []game.HarbingerRecord, epochKey string) *game.HarbingerRecord {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].TargetEpochKey == epochKey {
			return &history[i]
		}
	}
	return nil
}

// harbingerRecordText renders a resolved harbinger in one line.
func harbingerRecordText(r game.HarbingerRecord) string {
	var verdict string
	switch r.Outcome {
	case game.HarbingerOutcomeFulfilled:
		verdict = "[red]Fulfilled[-] [gray](invited)[-]"
	case game.HarbingerOutcomeVindicated:
		verdict = "[red]Vindicated[-]"
	case game.HarbingerOutcomeSpared:
		verdict = "[green]Spared[-]"
	case game.HarbingerOutcomeDiscredited:
		verdict = "[yellow]Discredited[-]"
	default:
		verdict = "[gray]unknown[-]"
	}
	if r.FalseProphet && r.Outcome != game.HarbingerOutcomeDiscredited {
		verdict += " [gray](a false prophet)[-]"
	}
	var extras []string
	if r.AppeaseLevel > 0 {
		extras = append(extras, fmt.Sprintf("appeased %d", r.AppeaseLevel))
	}
	if r.BraceLevel > 0 {
		extras = append(extras, fmt.Sprintf("braced %d", r.BraceLevel))
	}
	extra := ""
	if len(extras) > 0 {
		extra = " [gray](" + strings.Join(extras, ", ") + ")[-]"
	}
	chain := []string{capFirstUI(r.Name)}
	if len(r.Chain) > 1 {
		chain = chain[:0]
		for _, a := range r.Chain {
			if def, ok := config.HarbingerFor(a); ok {
				chain = append(chain, capFirstUI(def.Name))
			}
		}
	}
	target := r.TargetEpochName
	if r.TargetEpochKey != "" && r.TargetEpochKey == r.EpochKey {
		target = "doom in the " + r.TargetEpochName
	}
	if r.AtAdvance && r.Outcome != game.HarbingerOutcomeDiscredited {
		extra += " [gray](settled as you advanced)[-]"
	}
	return fmt.Sprintf("%s → %s: %s%s", strings.Join(chain, ", "), target, verdict, extra)
}

// epochProviderHistory renders the epoch history section.
func epochProviderHistory(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Epoch history ──[-]\n")

	allEpochs := config.Epochs()
	currentEpochOrder := -1
	epochByKey := config.EpochByKey()
	if ep, ok := epochByKey[state.EpochKey]; ok {
		currentEpochOrder = ep.Order
	}

	for _, ep := range allEpochs {
		if ep.Order > currentEpochOrder {
			break
		}

		isCurrent := ep.Key == state.EpochKey
		record := findEpochEvent(state.EpochEventHistory, ep.Key)

		var line strings.Builder
		fmt.Fprintf(&line, "   %s%s %s[-]", theme.NameTag(ep.Color), ep.Icon, ep.Name)

		if isCurrent {
			line.WriteString("   [gray](current)[-]")
		} else {
			if record != nil {
				evColor := epochEventColor(record.EventType)
				fmt.Fprintf(&line, "   [%s]%s[-]", evColor, record.EventName)
			} else {
				line.WriteString("   [gray]no event[-]")
			}

			if r := latestCatastrophe(state.EpochEventHistory, ep.Key); r != nil {
				line.WriteString("   " + catastropheOutcomeLabel(r.Outcome))
			}
		}

		sb.WriteString(line.String())
		sb.WriteString("\n")
	}
}

// epochProviderLegacyBonuses renders the legacy bonuses section.
func epochProviderLegacyBonuses(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Legacy bonuses (from Succumb) ──[-]\n")

	if len(state.LegacyBonuses) == 0 {
		sb.WriteString("   [gray]None yet[-]\n")
		return
	}

	allEpochs := config.Epochs()
	anyShown := false
	for _, ep := range allEpochs {
		if !state.LegacyBonuses[ep.Key] {
			continue
		}
		bonuses := config.LegacyBonusForEpoch(ep.Key)
		if len(bonuses) == 0 {
			continue
		}
		anyShown = true

		fmt.Fprintf(sb, "   %s%s %s:[-]  %s\n",
			theme.NameTag(ep.Color), ep.Icon, ep.Name, legacyBonusLine(bonuses))
	}

	if !anyShown {
		sb.WriteString("   [gray]None yet[-]\n")
	}
}

// epochProviderCivilizationLog renders the civilization catastrophe log section.
func epochProviderCivilizationLog(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Civilization log ──[-]\n")

	if len(state.CatastropheHistory) == 0 {
		sb.WriteString("   [gray]No catastrophes yet.[-]\n")
		return
	}

	for _, entry := range state.CatastropheHistory {
		fmt.Fprintf(sb, "   · %s\n", entry)
	}
}

// findEpochEvent returns the most recent transition event for the given epoch
// key, or nil. A catastrophe strikes inside its era now, after the
// transition's event, so it is skipped unless it is all the era has (saves
// from before fates, when a catastrophe took the transition's place).
func findEpochEvent(history []game.EpochEventRecord, epochKey string) *game.EpochEventRecord {
	var last, cat *game.EpochEventRecord
	for i := range history {
		if history[i].EpochKey != epochKey {
			continue
		}
		if history[i].EventType == "catastrophe" {
			cat = &history[i]
		} else {
			last = &history[i]
		}
	}
	if last == nil {
		return cat
	}
	return last
}

// latestCatastrophe returns the most recent catastrophe record for epochKey, or nil.
func latestCatastrophe(history []game.EpochEventRecord, epochKey string) *game.EpochEventRecord {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].EpochKey == epochKey && history[i].EventType == "catastrophe" {
			return &history[i]
		}
	}
	return nil
}

// catastropheOutcomeLabel renders a catastrophe record's outcome. Only an
// actual Endure reads as Endured: a pending catastrophe says so, and a record
// whose outcome was never stored (older saves) says it is unknown rather than
// claiming the player endured it.
func catastropheOutcomeLabel(outcome string) string {
	switch outcome {
	case game.CatastropheEndured:
		return "[green]✓ Endured[-]"
	case game.CatastropheSuccumbed:
		return "[yellow]Succumbed (legacy bonus)[-]"
	case game.CatastrophePending:
		return "[red]Pending[-]"
	}
	return "[gray]catastrophe (outcome not recorded)[-]"
}

// epochEventColor returns a tview color tag name for an epoch event type.
func epochEventColor(eventType string) string {
	switch eventType {
	case "good_legendary":
		return "gold"
	case "good_major":
		return "green"
	case "good_minor":
		return "cyan"
	case "bad_challenging":
		return "red"
	case "catastrophe":
		return "red"
	}
	return "white"
}
