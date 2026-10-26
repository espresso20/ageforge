package ui

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// wonderProgressBar returns a simple fill bar with colored segments.
// Filled portion uses BarFillColor (Accent role), empty uses BarEmptyColor (Dim role).
func wonderProgressBar(pct float64, width int) string {
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return BarFillColor() + strings.Repeat("█", filled) + BarEmptyColor() + strings.Repeat("░", empty) + "[-]"
}

// wonderInfo holds config + state for a wonder
type wonderInfo struct {
	key     string
	name    string
	ageKey  string
	ageName string
	def     config.BuildingDef
}

// getWonderList returns all wonders in age order
func getWonderList(set *rules.Set) []wonderInfo {
	ages := set.Ages()

	var wonders []wonderInfo
	for _, age := range ages {
		for _, bKey := range age.UnlockBuildings {
			if def, ok := set.Building(bKey); ok && def.Category == "wonder" {
				wonders = append(wonders, wonderInfo{
					key:     bKey,
					name:    def.Name,
					ageKey:  age.Key,
					ageName: age.Name,
					def:     def,
				})
				break // one wonder per age
			}
		}
	}
	return wonders
}

// currentWonder is the wonder of the age the town is in, nil when the age has
// none.
func currentWonder(state game.GameState) *wonderInfo {
	for _, w := range getWonderList(state.Ruleset()) {
		if w.ageKey == state.Age {
			return &w
		}
	}
	return nil
}

// bankRow is one resource of a wonder's price: what is banked of what it
// takes, as a share from 0 to 1, and the colour the bank's lines give it
// (red for none, yellow for some, green for all).
type bankRow struct {
	res          string
	banked, need float64
	share        float64
	color        string
}

// bankRows is a wonder's price a resource at a time, in resource order,
// against what is banked.
func bankRows(cost, banked map[string]float64) []bankRow {
	var rows []bankRow
	for _, k := range sortedKeysOf(cost) {
		r := bankRow{res: k, banked: banked[k], need: cost[k], color: "red"}
		if r.need > 0 {
			r.share = min(r.banked/r.need, 1)
		}
		if r.share >= 1 {
			r.color = "green"
		} else if r.share > 0 {
			r.color = "yellow"
		}
		rows = append(rows, r)
	}
	return rows
}

// wonderCollectHint is the registry's wonder collect form, escaped for tview.
func wonderCollectHint() string {
	return lit(helpRow("wonder collect").Form)
}

// formatEffect formats a building effect for display, in the same words as
// formatTechEffect ("+10% all production", "unlocks Lumber Mill"), colored
// by kind.
func formatEffect(eff config.Effect) string {
	text := formatTechEffect(eff)
	switch eff.Type {
	case "production":
		return "[green]" + text + "[-]"
	case "capacity":
		return "[yellow]" + text + "[-]"
	case "bonus":
		return "[cyan]" + text + "[-]"
	}
	return text
}
