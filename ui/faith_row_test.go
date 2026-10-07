package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The faith row and the catastrophe outlook show the faith strength the rolls
// read: the band's own odds, the strength as the percentage, and the faith
// behind it. The row used to work the band out again from the fill of the
// store, and could disagree with the roll at the band's edge.
func TestFaithRowShowsTheStrengthTheRollsRead(t *testing.T) {
	const full = 8943.0 // the Classical Age
	for _, c := range []struct {
		strength float64
		label    string
	}{
		{0, "No faith"}, {0.10, "Dim faith"}, {0.2499, "Dim faith"},
		{0.25, "Low faith"}, {0.50, "Low faith"}, {0.60, "◈ Faith"}, {0.75, "◈ Faith"},
		{0.7501, "Strong faith"}, {0.99, "Strong faith"}, {1, "Faith full"},
	} {
		band := game.FaithBandAt(c.strength, true)
		o := game.CatastropheOutlook{FaithStrength: c.strength, FaithFull: full, FaithBand: band}
		odds := fmt.Sprintf("(epoch: %.0f%% good)", game.EpochGoodChanceIn(band)*100)
		row := untag(formatFaithRow(game.ResourceState{Name: "Faith", Amount: c.strength * full, Storage: 1e12, Rate: 0.5}, o))
		if !strings.Contains(row, c.label) || !strings.Contains(row, odds) || !strings.Contains(row, fmt.Sprintf("%.0f%%", c.strength*100)) {
			t.Errorf("strength %v (%s): row %q, want %q, %q and the strength as a percentage", c.strength, band, row, c.label, odds)
		}
		if strings.Contains(row, "prestige bonus") {
			t.Errorf("strength %v: the row promises a prestige bonus nothing gives: %q", c.strength, row)
		}
		// The bar is the strength, not the fill of a store a trillion wide.
		if filled := strings.Count(row, "▓"); filled != int(c.strength*10) {
			t.Errorf("strength %v: the bar has %d of 10 cells filled, want %d", c.strength, filled, int(c.strength*10))
		}
	}

	st := game.GameState{
		Resources:          map[string]game.ResourceState{"faith": {Amount: 3577.9, Storage: 1e12}},
		CatastropheOutlook: game.CatastropheOutlook{FaithStrength: 0.40, FaithFull: full, FaithBand: game.FaithBandMid},
	}
	if got, want := faithStrengthText(st), "faith strength 40% (3.58K of 8.94K)"; got != want {
		t.Errorf("faithStrengthText = %q, want %q", got, want)
	}
	// An age with no measure of faith has nothing to set the faith against.
	st.CatastropheOutlook = game.CatastropheOutlook{FaithBand: game.FaithBandMid}
	if got, want := faithStrengthText(st), "faith strength 0%"; got != want {
		t.Errorf("with no measure: faithStrengthText = %q, want %q", got, want)
	}
}
