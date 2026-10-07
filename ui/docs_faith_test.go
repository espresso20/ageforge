package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The Faith page says where the faith bands begin, in moderate sets of faith
// buildings, and lists the moderate set age by age: how many faith buildings
// it holds, what it makes per tick and what a town's faith buildings must
// average for the middle and top bands. The figures follow the faith
// buildings' rates and the rule's constants, so a balance change moves them:
// this holds the page to the rule the rolls read.
func TestDocsFaithStrengthTable(t *testing.T) {
	set := rules.Core()
	page := readDoc(t, "../site/docs/faith.md")
	wantRow := func(row string) {
		t.Helper()
		if !strings.Contains(page, row) {
			t.Errorf("site/docs/faith.md has no row %q: the page is out of step with the game", row)
		}
	}
	// Faith strength by devotion, with all the faith kept.
	pct := func(sets float64) string {
		return fmt.Sprintf("%.0f%%", game.FaithStrengthOf(sets, game.FaithSave{Moderate: 1, Own: sets})*100)
	}
	mid, top := game.FaithMidAt*game.FaithFullSets, game.FaithHighAbove*game.FaithFullSets
	wantRow("| 1x (a moderate town) | " + pct(1) + " | bottom |")
	wantRow(fmt.Sprintf("| %gx | %s | the middle band begins |", mid, pct(mid)))
	wantRow("| 2x | " + pct(2) + " | middle |")
	wantRow("| 3x | " + pct(3) + " | middle |")
	wantRow(fmt.Sprintf("| %gx | %s | the top band begins just above |", top, pct(top)))
	wantRow("| 3.5x | " + pct(3.5) + " | top |")
	wantRow(fmt.Sprintf("| %gx or more | 100%% | top |", game.FaithFullSets))
	if game.FaithBandAt(1/game.FaithFullSets) != game.FaithBandLow || game.FaithBandAt(2/game.FaithFullSets) != game.FaithBandMid ||
		game.FaithBandAt(3/game.FaithFullSets) != game.FaithBandMid || game.FaithBandAt(3.5/game.FaithFullSets) != game.FaithBandHigh {
		t.Error("the page says a moderate town is in the bottom band, twice and three times the set in the middle and 3.5 times in the top; the rule no longer does")
	}

	// The moderate set, age by age.
	idx := set.Indexes()
	for _, key := range set.AgeKeys() {
		kinds := 0
		for _, d := range set.Buildings() {
			if d.Category == "wonder" || idx[d.RequiredAge] > idx[key] {
				continue
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Target == "faith" && e.Value > 0 {
					kinds++
					break
				}
			}
		}
		age, _ := set.Age(key)
		makes := set.FlowBuildingOutput("faith", key)
		wantRow(fmt.Sprintf("| %s | %d | %s | %s | %s |", age.Name, kinds*int(config.FlowCopies), FormatNumber(makes), FormatNumber(makes*mid), FormatNumber(makes*top)))
	}
}
