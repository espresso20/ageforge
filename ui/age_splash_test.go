package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestAgeSplashUpgradeCountsLineUp: the splash centers every line by
// itself, so the upgrade rows' counts sit in one column only when the rows
// are all one width, however long a name is and however many digits a count
// has.
func TestAgeSplashUpgradeCountsLineUp(t *testing.T) {
	summary := game.AgeAdvanceSummary{BuildingsTransformed: []game.BuildingTransform{
		{OldName: "Elders' Hall", NewName: "Scriptorium", Count: 5},
		{OldName: "Longhouse", NewName: "House", Count: 15},
		{OldName: "Woodcutter Camp", NewName: "Lumber Mill", Count: 4},
		{OldName: "Nuclear Extraction Plant", NewName: "Uranium Processing Works", Count: 120},
	}}
	var rows []string
	for _, line := range strings.Split(untag(buildAgeSplashText("bronze_age", summary, false, game.EpochEventRecord{})), "\n") {
		if strings.Contains(line, " → ") {
			rows = append(rows, line)
		}
	}
	if len(rows) != len(summary.BuildingsTransformed) {
		t.Fatalf("found %d upgrade rows, want %d:\n%s", len(rows), len(summary.BuildingsTransformed), strings.Join(rows, "\n"))
	}
	width, column := runeLen(rows[0]), -1
	for _, row := range rows {
		if runeLen(row) != width {
			t.Errorf("upgrade row is %d wide, the first is %d: %q", runeLen(row), width, row)
		}
		at := runeLen(row[:strings.LastIndex(row, " x")])
		if column >= 0 && at != column {
			t.Errorf("count starts at column %d, the first at %d: %q", at, column, row)
		}
		column = at
	}
}
