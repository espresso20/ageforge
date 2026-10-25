package ui

import (
	"strings"

	"github.com/espresso20/ageforge/config"
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
