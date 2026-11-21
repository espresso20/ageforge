package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// armyPage is the Army panel: the soldiers, posts, defense and threat, the
// campaign under way, and the campaigns that could be sent.
func armyPage(state game.GameState) deckPage {
	mil := state.Military
	ageName := state.AgeName
	if ageName == "" {
		ageName = state.Age
	}
	pg := deckPage{
		Title: rowid.Line("ui.overlay_military.armyPage.01", "Army"),
		Hint:  "army",
	}
	pg.Notes = append(garrisonNotes(state), deckLine{Value: rowid.Line("ui.overlay_military.armyPage.02", "Commands: expedition opens the Expeditions panel; expedition <name> sends an offer")})

	soldiers := deckCard{Title: rowid.Line("ui.overlay_military.armyPage.03", "Soldiers"), Lines: []deckLine{
		{Value: rowid.Sprintf("ui.overlay_military.armyPage.04", "%s posted of %s", textfmt.Number(float64(mil.SoldierCount)), textfmt.Number(float64(mil.SoldierCap))), K: xBright}}}
	if mil.MilitaryBonus > 0 {
		soldiers.Lines = append(soldiers.Lines, deckLine{Value: rowid.Line("ui.overlay_military.armyPage.05", "Military Bonus") + fmt.Sprintf(" +%.0f%%", mil.MilitaryBonus*100), K: xGood})
	}
	if mil.ExpeditionBonus > 0 {
		soldiers.Lines = append(soldiers.Lines, deckLine{Value: rowid.Line("ui.overlay_military.armyPage.07", "Expedition Bonus") + fmt.Sprintf(" +%.0f%%", mil.ExpeditionBonus*100), K: xGood})
	}
	posts := deckCard{Title: rowid.Line("ui.overlay_military.armyPage.09", "Posts")}
	if len(mil.Posts) == 0 {
		// Soldiers begin with the first military building: until one is built
		// there is nobody at a post, and the panel says so.
		posts.Desc = rowid.Line("ui.overlay_military.armyPage.10", "none yet. A military building comes with its soldiers.")
	}
	for _, p := range mil.Posts {
		posts.Lines = append(posts.Lines, deckLine{fmt.Sprintf("%s x%d", p.Name, p.Count) + " ", fmt.Sprintf("%s/%s", textfmt.Number(float64(p.Posted)), textfmt.Number(float64(p.Posts))), xBright})
	}
	pg.Sections = append(pg.Sections, deckSection{Title: rowid.Line("ui.overlay_military.armyPage.13", "Overview"), MinW: 30, Cards: []deckCard{
		soldiers, posts,
		{Title: rowid.Line("ui.overlay_military.armyPage.14", "Defense"), Lines: []deckLine{{Value: textfmt.Number(mil.DefenseRating), K: xBright}}},
		{Title: rowid.Line("ui.overlay_military.armyPage.15", "Threat"), Lines: []deckLine{
			{Value: textfmt.Number(mil.Threat), K: xBright}, {Value: rowid.Sprintf("ui.overlay_military.armyPage.16", "raids in the %s", ageName), K: xDim}}},
	}})

	// Only the military campaign lives here; the active scout shows in the
	// Expeditions panel (the two run concurrently, one per category).
	if a := mil.ActiveMilitary; a != nil {
		c := deckCard{Title: a.Name, Chip: rowid.Line("ui.overlay_military.armyPage.17", "Campaign"), ChipK: xChipMil, Meter: true,
			MeterText: rowid.Sprintf("ui.overlay_military.armyPage.18", "%s left", formatTicks(a.TicksLeft, state))}
		if a.TotalTicks > 0 {
			c.Share = max(0, min(1, 1-float64(a.TicksLeft)/float64(a.TotalTicks)))
		}
		if a.Soldiers > 0 {
			c.Lines = []deckLine{{Value: rowid.Sprintf("ui.overlay_military.armyPage.19", "%d soldiers deployed", a.Soldiers)}}
		}
		pg.Sections = append(pg.Sections, deckSection{Title: rowid.Line("ui.overlay_military.armyPage.20", "Active Campaign"), Cards: []deckCard{c}})
	}

	// The Army panel lists only military campaigns (they cost soldiers).
	camps := deckSection{Title: rowid.Line("ui.overlay_military.armyPage.21", "Campaigns"), MinW: 46}
	if mil.ActiveMilitary == nil {
		camps.Notes = append(camps.Notes, deckLine{Value: rowid.Line("ui.overlay_military.armyPage.22", "No active campaign")})
	}
	camps.Notes = append(camps.Notes, deckLine{Value: rowid.Sprintf("ui.overlay_military.armyPage.23", "Completed: %d (expeditions and campaigns)", mil.CompletedCount)})
	for _, exp := range mil.Expeditions {
		if exp.Category == game.ExpeditionMilitary {
			camps.Cards = append(camps.Cards, campaignCard(exp, state))
		}
	}
	if len(camps.Cards) == 0 {
		camps.Notes = append(camps.Notes,
			deckLine{Value: rowid.Line("ui.overlay_military.armyPage.24", "No campaigns available yet.")},
			deckLine{Value: rowid.Line("ui.overlay_military.armyPage.25", "Campaigns are grand missions and cost soldiers,")},
			deckLine{Value: rowid.Line("ui.overlay_military.armyPage.26", "which come with military buildings. They open as your civilization grows.")})
	}
	pg.Sections = append(pg.Sections, camps)
	return pg
}

// campaignCard is one military campaign on the board.
func campaignCard(exp game.ExpeditionInfo, state game.GameState) deckCard {
	c := deckCard{Title: exp.Name, Desc: exp.Description}
	if exp.SoldiersNeeded > 0 {
		c.Lines = append(c.Lines, deckLine{rowid.Line("ui.overlay_military.campaignCard.01", "Soldiers"), strconv.Itoa(exp.SoldiersNeeded), 0})
	}
	// Duration is rolled per launch, so preview the def's range rather than a
	// single number.
	c.Lines = append(c.Lines, deckLine{rowid.Line("ui.overlay_military.campaignCard.02", "Duration"), formatTickRange(exp.DurationMin, exp.DurationMax, state), 0})
	diff := xGood
	if exp.Difficulty > 0.5 {
		diff = xBad
	} else if exp.Difficulty > 0.3 {
		diff = xGold
	}
	// The listed difficulty, and what the army leaves of it today.
	note := rowid.Line("ui.overlay_military.campaignCard.03", "(chance of failure before bonuses)")
	if exp.Chance > 0 && exp.Chance < exp.Difficulty-0.0005 {
		note = rowid.Sprintf("ui.overlay_military.campaignCard.04", "(chance of failure; %.0f%% with your army)", exp.Chance*100)
	}
	c.Lines = append(c.Lines, deckLine{rowid.Line("ui.overlay_military.campaignCard.05", "Difficulty"), fmt.Sprintf("%.0f%%", exp.Difficulty*100), diff}, deckLine{Value: note, K: xDim})
	if cost := formatExpeditionCost(exp.Cost, state); cost != "" {
		c.Lines = append(c.Lines, deckLine{rowid.Line("ui.overlay_military.campaignCard.07", "Cost"), cost, 0})
	}
	if exp.CanLaunch {
		c.Chip, c.ChipK = rowid.Line("ui.overlay_military.campaignCard.08", "ready"), xChipCivic
		c.Action, c.Command = rowid.Line("ui.overlay_military.campaignCard.09", "send"), "expedition "+exp.Name
	} else {
		c.Chip, c.ChipK, c.Quiet = rowid.Line("ui.overlay_military.campaignCard.10", "locked"), xKeyName, true
		c.Why = exp.LaunchBlockReason
	}
	return c
}

// formatExpeditionCost renders an expedition's resource cost in a readable,
// player-facing form (e.g. "30 food, 30 wood"), with keys sorted for stable
// output. Returns "" for a free (empty) cost so callers can omit the line.
func formatExpeditionCost(cost map[string]float64, state game.GameState) string {
	if len(cost) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cost))
	for _, k := range sortedKeysOf(cost) {
		parts = append(parts, textfmt.Number(cost[k])+" "+resourceDisplayName(state, k))
	}
	return strings.Join(parts, ", ")
}

func expeditionHaul(exp game.ExpeditionInfo, state game.GameState) string {
	if haul := formatExpeditionCost(exp.ExpectedHaulMin, state); haul != "" {
		return haul
	}
	return fmt.Sprintf(txtAboutMinutesOfIncome.String(), exp.HaulMinutes)
}

func expeditionOfferCost(exp game.ExpeditionInfo, state game.GameState) string {
	if cost := formatExpeditionCost(exp.Cost, state); cost != "" {
		return cost
	}
	if exp.Tier == game.TierGrand && exp.Grand == game.GrandCivil {
		return "materials"
	}
	if exp.Tier == game.TierGrand && exp.Grand == game.GrandMilitary {
		return "soldiers"
	}
	return "free"
}

// garrisonNotes says what the army does for the player: how much of a raid
// the garrison would blunt against the current age's threat, and what it has
// saved this run. Worded from the player's side.
func garrisonNotes(state game.GameState) []deckLine {
	mil := state.Military
	capPct := config.DefenseMitigationCap * 100
	var out []deckLine
	if mil.Mitigation <= 0 {
		out = append(out,
			deckLine{Value: rowid.Line("ui.overlay_military.garrisonNotes.01", "You have no garrison: raids hit you with full force."), K: xBad},
			deckLine{Value: rowid.Sprintf("ui.overlay_military.garrisonNotes.02", "Soldiers blunt raids, war raids and what an Endure takes (up to %.0f%%).", capPct)})
	} else {
		twice := config.DefenseMitigation(mil.DefenseRating*2, mil.Threat)
		out = append(out,
			deckLine{Value: rowid.Sprintf("ui.overlay_military.garrisonNotes.03", "Your garrison would blunt about %.0f%% of a raid.", mil.Mitigation*100), K: xGood},
			deckLine{Value: rowid.Line("ui.overlay_military.garrisonNotes.04", "Raids, war raids and an Endure's losses all hit you that much softer.")},
			deckLine{Value: rowid.Sprintf("ui.overlay_military.garrisonNotes.05", "Twice the garrison: about %.0f%%. No army blunts more than %.0f%%.", twice*100, capPct)})
	}
	if line := garrisonSavedSummary(mil.Saved); line != "" {
		out = append(out, deckLine{Label: rowid.Line("ui.overlay_military.garrisonNotes.06", "Saved this run: "), Value: kit.Strip(line), K: xText})
	}
	return out
}

// garrisonSavedSummary is a one-line summary of what the army has saved this
// run, largest resource amounts first (at most four), or "" when nothing.
func garrisonSavedSummary(t *game.DefenseTally) string {
	if t == nil {
		return ""
	}
	var parts []string
	if t.Buildings > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", t.Buildings, pluralize(rowid.Line("ui.overlay_military.garrisonSavedSummary.01", "building"), t.Buildings)))
	}
	if t.Workers > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", t.Workers, pluralize(txtWorker.String(), t.Workers)))
	}
	keys := make([]string, 0, len(t.Resources))
	for k, v := range t.Resources {
		if v >= 1 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if t.Resources[keys[i]] != t.Resources[keys[j]] {
			return t.Resources[keys[i]] > t.Resources[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > 4 {
		keys = keys[:4]
	}
	for _, k := range keys {
		parts = append(parts, textfmt.Number(t.Resources[k])+" "+k)
	}
	if len(parts) == 0 {
		return ""
	}
	out := strings.Join(parts, ", ")
	if t.Raids > 0 {
		out += rowid.Sprintf("ui.overlay_military.garrisonSavedSummary.03", " [gray](%d %s blunted)[-]", t.Raids, pluralize(rowid.Line("ui.overlay_military.garrisonSavedSummary.04", "raid"), t.Raids))
	}
	return out
}

// resourceDisplayName is a resource's player-facing name in lower case ("dark
// matter"), from the snapshot, or the key with spaces when it has none.
func resourceDisplayName(state game.GameState, key string) string {
	if r, ok := state.Resources[key]; ok && r.Name != "" {
		return strings.ToLower(r.Name)
	}
	return strings.ReplaceAll(key, "_", " ")
}
