package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// buildingText is a building's description as the lists show it. A knowledge
// building also says where its research halves ("Past 5 copies, each
// researches at half rate.") and, when the description does not say it (it
// does: "comes with 2 scholars"), its scholars a copy. Nothing refuses a copy
// past that count, so the line is all it says about the count.
func buildingText(bs game.BuildingState) string {
	if bs.Scholars <= 0 {
		return bs.Description
	}
	count := textfmt.Count(bs.Scholars, vibe.Line("ui.building_text.buildingText.01", "scholar"), vibe.Line("ui.building_text.buildingText.02", "scholars"))
	line := ""
	if !strings.Contains(bs.Description, count) {
		line = textfmt.Capitalize(count+vibe.Line("ui.building_text.buildingText.03", " a copy")) + "."
	}
	if bs.FullRateCopies > 0 {
		if line != "" {
			line += " "
		}
		line += vibe.Sprintf("ui.building_text.buildingText.07", "Past %s, each researches at half rate.", textfmt.Count(bs.FullRateCopies, vibe.Line("ui.building_text.buildingText.05", "copy"), vibe.Line("ui.building_text.buildingText.06", "copies")))
	}
	if line == "" {
		return bs.Description
	}
	return strings.TrimSpace(bs.Description + " " + line)
}
