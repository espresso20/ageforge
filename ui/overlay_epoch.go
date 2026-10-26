package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// epochProvider generates the epoch overlay text from the current game state.
// Shows current epoch status, epoch history, legacy bonuses, and civilization log.
func epochProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	sb.WriteString(vibe.Line("ui.overlay_epoch.epochProvider.01", "[gold]═══ Epoch ═══[-]\n\n"))

	epochProviderCurrentEpoch(&sb, state)
	sb.WriteString("\n")
	epochProviderHistory(&sb, state)
	sb.WriteString("\n")
	epochProviderHarbingers(&sb, state)
	sb.WriteString("\n")
	epochProviderLegacyBonuses(&sb, state)
	sb.WriteString("\n")
	writeMasterySection(&sb, state)
	sb.WriteString("\n")
	epochProviderCivilizationLog(&sb, state)

	return sb.String()
}

// epochProviderCurrentEpoch renders the current epoch status section.
func epochProviderCurrentEpoch(sb *strings.Builder, state game.GameState) {
	sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.01", "[gold]── Current epoch ──[-]\n\n"))

	if state.EpochKey == "" {
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.02", " [gray]No epoch data yet.[-]\n"))
		return
	}

	set := state.Ruleset()
	ep, ok := set.Era(state.EpochKey)
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
			ageNames = append(ageNames, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.03", "%d more to come", unseen))
		}
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.04", " %s%s %s[-]   Ages: %s\n",
			theme.NameTag(ep.Color), ep.Icon, ep.Name,
			strings.Join(ageNames, " · ")))
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.05", " [gray]Primary resource: %s   Energy: %s[-]\n",
			game.ResourceName(ep.PrimaryResource), game.ResourceName(ep.EnergyResource)))
	}

	sb.WriteString("\n")

	// Current epoch event (the transition's; a catastrophe is shown below)
	currentEvent := findEpochEvent(state.EpochEventHistory, state.EpochKey)
	if currentEvent == nil {
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.06", " Epoch event: [gray]No event this epoch[-]\n"))
	} else {
		evColor := epochEventColor(currentEvent.EventType)
		if currentEvent.EventType == "catastrophe" {
			fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.07", " Epoch event: [%s]%s[-]   [red](catastrophe)[-]\n",
				evColor, currentEvent.EventName))
		} else {
			flavorStr := ""
			durationStr := ""
			if evDef, found := set.EraEvent(currentEvent.EventKey); found {
				if evDef.FlavorText != "" {
					flavorStr = evDef.FlavorText
				}
				if evDef.Duration > 0 {
					durationStr = vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.08", "   %s duration", formatTicks(evDef.Duration, state))
				}
			}
			fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.09", " Epoch event: [%s]%s[-]%s\n", evColor, currentEvent.EventName, durationStr))
			if flavorStr != "" {
				fmt.Fprintf(sb, " [comment]  %s[-]\n", flavorStr)
			}
		}
	}

	// What a good event at the next era's door could be: the tiers the
	// town's culture strength opens (game/culture.go). Shown once culture is
	// held, while there is an era to come.
	if _, next := set.NextEra(state.EpochKey); next && state.Resources["culture"].Unlocked {
		o := state.CatastropheOutlook
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.10", " Good events at the next era: %s\n", cultureTierText(o.CultureTier)))
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.11", "   Culture strength %s. Major from %.0f%%, the Legendary from %.0f%%.",
			cultureStrengthFigure(state), game.CultureMajorAbove*100, game.CultureLegendaryAbove*100)) + "\n")
	}

	sb.WriteString("\n")

	// Catastrophe status. Only what the player has seen: a doom not yet
	// foretold reads the same as none (the no-leak rule, game/fate.go).
	o := state.CatastropheOutlook
	switch {
	case state.PendingCatastrophe != "":
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.12", " Catastrophe: [red]pending. Type catastrophe to choose Endure or Succumb.[-]\n"))
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.13", " [gray]  Advancing and prestige are blocked until you decide.[-]\n"))
	case state.LastPassage.Pending:
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.14", " Catastrophe: [red]the Last Passage. Type catastrophe to choose Endure or Succumb.[-]\n"))
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.15", " [gray]  Prestige waits until you decide; nothing else does.[-]\n"))
	case !set.CatastropheAllowed(state.EpochKey):
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.16", " Catastrophe: %s\n", theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.17", "none in the ")+currentEraName(state))))
	default:
		if r := latestCatastrophe(state.EpochEventHistory, state.EpochKey); r != nil {
			fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.18", " Catastrophe: %s\n", catastropheOutcomeLabel(r.Outcome)))
		} else if r := eraDoomRecord(state); r != nil && r.Outcome == game.HarbingerOutcomeSpared {
			sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.19", " Catastrophe: [green]spared[-] [gray](nothing more will strike this era)[-]\n"))
		} else if o.Warned {
			sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.20", " Catastrophe: [yellow]foretold[-]\n"))
		} else {
			sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.21", " Catastrophe: [gray]none so far[-]\n"))
		}
	}

	// The outlook (same wording the `catastrophe` command uses: a figure from
	// the Industrial Age on, a severity before it).
	if o.Warned {
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.22", " Doom foretold: [yellow]%s[-] [gray](more faith, lower odds)[-]\n", doomWarningText(state)))
	}
	switch {
	case o.Possible && o.Passage == game.PassagePrestige:
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.23", " Next passage (prestige, the Last Passage): [yellow]%s[-] [gray](more faith, lower odds)[-]\n",
			outlookRiskText(state)))
	case o.Possible && !o.Warned && o.Passage == game.PassageEpoch:
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.24", " Outlook: [gray]no harbinger has come. Quiet, for now.[-]\n"))
	}

	// Harbinger status.
	if h := state.Harbinger; h != nil {
		status := vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.25", "no word of when")
		if h.WhenText != "" {
			status = vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.26", "doom ") + h.WhenText
		}
		switch {
		case h.PassageCame:
			status = vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.27", "the Last Passage has come")
		case h.LastPassage && h.Invited:
			status = vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.28", "invited, the Last Passage will come at prestige")
		case h.LastPassage:
			status = vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.29", "waiting for your prestige")
		case h.Invited:
			status = vibe.Line("ui.overlay_epoch.epochProviderCurrentEpoch.30", "invited, the catastrophe will come")
		}
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.31", " Harbinger: [warning]%s is here[-] [gray](%s; appease %d/%d, brace %d/%d; type 'harbinger')[-]\n",
			textfmt.Capitalize(h.Name), status, h.AppeaseLevel, game.HarbingerMaxAppease, h.BraceLevel, game.HarbingerMaxBrace))
	} else if r := latestHarbinger(state.HarbingerHistory, state.EpochKey); r != nil {
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_epoch.epochProviderCurrentEpoch.32", " Harbinger: %s\n", harbingerRecordText(state.Ruleset(), *r)))
	}
}

// epochProviderHarbingers lists this run's resolved harbingers, oldest first.
func epochProviderHarbingers(sb *strings.Builder, state game.GameState) {
	sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderHarbingers.01", " [yellow]── Harbingers ──[-]\n"))
	if len(state.HarbingerHistory) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderHarbingers.02", "   [gray]None have come and gone this run.[-]\n"))
		return
	}
	for _, r := range state.HarbingerHistory {
		fmt.Fprintf(sb, "   · %s\n", harbingerRecordText(state.Ruleset(), r))
	}
}

// latestHarbinger returns the newest harbinger record of epochKey's doom, or
// nil.
func latestHarbinger(history []game.HarbingerRecord, epochKey string) *game.HarbingerRecord {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].TargetEpochKey == epochKey {
			return &history[i]
		}
	}
	return nil
}

// harbingerRecordText renders a resolved harbinger in one line, naming the
// figures of its chain from set.
func harbingerRecordText(set *rules.Set, r game.HarbingerRecord) string {
	var verdict string
	switch r.Outcome {
	case game.HarbingerOutcomeFulfilled:
		verdict = vibe.Line("ui.overlay_epoch.harbingerRecordText.01", "[red]Fulfilled[-] [gray](invited)[-]")
	case game.HarbingerOutcomeVindicated:
		verdict = vibe.Line("ui.overlay_epoch.harbingerRecordText.02", "[red]Vindicated[-]")
	case game.HarbingerOutcomeSpared:
		verdict = vibe.Line("ui.overlay_epoch.harbingerRecordText.03", "[green]Spared[-]")
	case game.HarbingerOutcomeDiscredited:
		verdict = vibe.Line("ui.overlay_epoch.harbingerRecordText.04", "[yellow]Discredited[-]")
	default:
		verdict = vibe.Line("ui.overlay_epoch.harbingerRecordText.05", "[gray]unknown[-]")
	}
	if r.FalseProphet && r.Outcome != game.HarbingerOutcomeDiscredited {
		verdict += vibe.Line("ui.overlay_epoch.harbingerRecordText.06", " [gray](a false prophet)[-]")
	}
	var extras []string
	if r.AppeaseLevel > 0 {
		extras = append(extras, vibe.Sprintf("ui.overlay_epoch.harbingerRecordText.07", "appeased %d", r.AppeaseLevel))
	}
	if r.BraceLevel > 0 {
		extras = append(extras, vibe.Sprintf("ui.overlay_epoch.harbingerRecordText.08", "braced %d", r.BraceLevel))
	}
	extra := ""
	if len(extras) > 0 {
		extra = " [gray](" + strings.Join(extras, ", ") + ")[-]"
	}
	chain := []string{textfmt.Capitalize(r.Name)}
	if len(r.Chain) > 1 {
		chain = chain[:0]
		for _, a := range r.Chain {
			if def, ok := set.Harbinger(a); ok {
				chain = append(chain, textfmt.Capitalize(def.Name))
			}
		}
	}
	target := r.TargetEpochName
	if r.TargetEpochKey != "" && r.TargetEpochKey == r.EpochKey {
		target = vibe.Line("ui.overlay_epoch.harbingerRecordText.09", "doom in the ") + r.TargetEpochName
	}
	if r.AtAdvance && r.Outcome != game.HarbingerOutcomeDiscredited {
		extra += vibe.Line("ui.overlay_epoch.harbingerRecordText.10", " [gray](settled as you advanced)[-]")
	}
	return fmt.Sprintf("%s → %s: %s%s", strings.Join(chain, ", "), target, verdict, extra)
}

// epochProviderHistory renders the epoch history section.
func epochProviderHistory(sb *strings.Builder, state game.GameState) {
	sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderHistory.01", " [yellow]── Epoch history ──[-]\n"))

	allEpochs := state.Ruleset().Eras()
	currentEpochOrder := -1
	if ep, ok := state.Ruleset().Era(state.EpochKey); ok {
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
			line.WriteString(vibe.Line("ui.overlay_epoch.epochProviderHistory.02", "   [gray](current)[-]"))
		} else {
			if record != nil {
				evColor := epochEventColor(record.EventType)
				fmt.Fprintf(&line, "   [%s]%s[-]", evColor, record.EventName)
			} else {
				line.WriteString(vibe.Line("ui.overlay_epoch.epochProviderHistory.03", "   [gray]no event[-]"))
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
	sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderLegacyBonuses.01", " [yellow]── Legacy bonuses (from Succumb) ──[-]\n"))

	if len(state.LegacyBonuses) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderLegacyBonuses.02", "   [gray]None yet[-]\n"))
		return
	}

	allEpochs := state.Ruleset().Eras()
	anyShown := false
	for _, ep := range allEpochs {
		if !state.LegacyBonuses[ep.Key] {
			continue
		}
		bonuses := state.Ruleset().LegacyBonus(ep.Key)
		if len(bonuses) == 0 {
			continue
		}
		anyShown = true

		fmt.Fprintf(sb, "   %s%s %s:[-]  %s\n",
			theme.NameTag(ep.Color), ep.Icon, ep.Name, legacyBonusLine(state, bonuses))
	}

	if !anyShown {
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderLegacyBonuses.03", "   [gray]None yet[-]\n"))
	}
}

// epochProviderCivilizationLog renders the civilization catastrophe log section.
func epochProviderCivilizationLog(sb *strings.Builder, state game.GameState) {
	sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCivilizationLog.01", " [yellow]── Civilization log ──[-]\n"))

	if len(state.CatastropheHistory) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_epoch.epochProviderCivilizationLog.02", "   [gray]No catastrophes yet.[-]\n"))
		return
	}

	for _, entry := range state.CatastropheHistory {
		fmt.Fprintf(sb, "   · %s\n", entry)
	}
}

// findEpochEvent returns the most recent transition event for the given epoch
// key, or nil. A catastrophe strikes inside its era now, after the
// transition's event, so it is skipped unless it is all the era has.
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
// actual Endure reads as Endured: a pending catastrophe says so.
func catastropheOutcomeLabel(outcome string) string {
	switch outcome {
	case game.CatastropheEndured:
		return vibe.Line("ui.overlay_epoch.catastropheOutcomeLabel.01", "[green]✓ Endured[-]")
	case game.CatastropheSuccumbed:
		return vibe.Line("ui.overlay_epoch.catastropheOutcomeLabel.02", "[yellow]Succumbed (legacy bonus)[-]")
	case game.CatastrophePending:
		return vibe.Line("ui.overlay_epoch.catastropheOutcomeLabel.03", "[red]Pending[-]")
	}
	return ""
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
