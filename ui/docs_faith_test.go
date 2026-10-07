package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The Faith page lists, for every age, the faith that reads as full faith
// strength, where the middle and top bands begin and what a moderate faith
// economy makes per tick. The figures follow the faith buildings' rates and
// the pacing targets, so a balance change moves them: this holds the page's
// table to the rule the rolls read.
func TestDocsFaithStrengthTable(t *testing.T) {
	set := rules.Core()
	page := readDoc(t, "../site/docs/faith.md")
	rows := 0
	for _, m := range game.FaithMeasuresIn(set) {
		age, _ := set.Age(m.Age)
		want := fmt.Sprintf("| %s | %s | %s | %s | %s |", age.Name, FormatNumber(m.Full), FormatNumber(m.Full*game.FaithMidAt),
			FormatNumber(m.Full*game.FaithHighAbove), FormatNumber(set.FlowIncome("faith", m.Age)))
		if !strings.Contains(page, want) {
			t.Errorf("site/docs/faith.md has no row %q: the full-strength table is out of step with the game", want)
		}
		rows++
	}
	if rows != len(set.AgeKeys()) {
		t.Errorf("%d ages checked, want %d", rows, len(set.AgeKeys()))
	}
}
