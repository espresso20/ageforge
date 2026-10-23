package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// buildingText is a building's description as the lists show it. A knowledge
// building also gives its copies against the most that can be built ("3 of 5
// built.") and, when the description does not say it (it does: "comes with 2
// scholars"), its scholars a copy.
func buildingText(bs game.BuildingState) string {
	if bs.Scholars <= 0 {
		return bs.Description
	}
	count := textfmt.Count(bs.Scholars, vibe.Line("ui.building_text.buildingText.01", "scholar"), vibe.Line("ui.building_text.buildingText.02", "scholars"))
	line := ""
	if !strings.Contains(bs.Description, count) {
		line = count + vibe.Line("ui.building_text.buildingText.03", " a copy")
	}
	if bs.ScholarCap > 0 {
		if line != "" {
			line += ", "
		}
		line += vibe.Sprintf("ui.building_text.buildingText.04", "%d of %d built", bs.Count, bs.ScholarCap)
	}
	if line == "" {
		return bs.Description
	}
	return strings.TrimSpace(bs.Description + " " + textfmt.Capitalize(line) + ".")
}

// atLimit is the words the Buildings panel gives a building that can be
// built no further, for a knowledge building at its cap in the lists that do
// not show them already.
func atLimit(bs game.BuildingState) string {
	if bs.Scholars > 0 && bs.AtMaxCount {
		return vibe.Line("ui.building_text.atLimit.01", "Building limit reached.")
	}
	return ""
}
