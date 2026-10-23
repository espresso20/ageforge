package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
	count := textfmt.Count(bs.Scholars, "scholar", "scholars")
	line := ""
	if !strings.Contains(bs.Description, count) {
		line = textfmt.Capitalize(count+" a copy") + "."
	}
	if bs.FullRateCopies > 0 {
		if line != "" {
			line += " "
		}
		line += fmt.Sprintf("Past %s, each researches at half rate.", textfmt.Count(bs.FullRateCopies, "copy", "copies"))
	}
	if line == "" {
		return bs.Description
	}
	return strings.TrimSpace(bs.Description + " " + line)
}
