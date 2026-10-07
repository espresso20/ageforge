package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The faith row and the catastrophe outlook show the faith strength the
// rolls read: the band's own odds, the strength as the bar and the
// percentage, and the two things it comes from. The row used to work the
// band out again from the fill of the store, and promised a prestige bonus
// nothing gave.
func TestFaithRowShowsTheStrengthTheRollsRead(t *testing.T) {
	for _, c := range []struct {
		strength float64
		label    string
	}{
		{0, "Dim faith"}, {0.10, "Dim faith"}, {0.2499, "Dim faith"},
		{0.25, "Low faith"}, {0.50, "Low faith"}, {0.60, "◈ Faith"}, {0.75, "◈ Faith"},
		{0.7501, "Strong faith"}, {0.99, "Strong faith"}, {1, "Faith full"},
	} {
		band := game.FaithBandAt(c.strength)
		o := game.CatastropheOutlook{FaithStrength: c.strength, FaithBand: band, FaithDevotion: c.strength * game.FaithFullSets, FaithKept: 1}
		odds := fmt.Sprintf("(epoch: %.0f%% good)", game.EpochGoodChanceIn(band)*100)
		// The store is a trillion wide and half full: none of the row's business.
		row, bar := faithRowText(game.ResourceState{Name: "Faith", Amount: 5e11, Storage: 1e12, Rate: 0.5}, o)
		if !strings.Contains(row, c.label) || !strings.Contains(row, odds) || !strings.Contains(row, fmt.Sprintf("%.0f%%", c.strength*100)) {
			t.Errorf("strength %v (%s): row %q, want %q, %q and the strength as a percentage", c.strength, band, row, c.label, odds)
		}
		if strings.Contains(row, "prestige bonus") {
			t.Errorf("strength %v: the row promises a prestige bonus nothing gives: %q", c.strength, row)
		}
		if filled := strings.Count(row, "▓"); bar != resBarMax || filled != int(c.strength*float64(bar)) {
			t.Errorf("strength %v: the bar has %d of %d cells filled, want %d of %d", c.strength, filled, bar, int(c.strength*resBarMax), resBarMax)
		}
	}
	// A town holding no faith at all is told so.
	if row, _ := faithRowText(game.ResourceState{Name: "Faith"}, game.CatastropheOutlook{FaithBand: game.FaithBandLow}); !strings.Contains(row, "No faith") || !strings.Contains(row, "(epoch: 40% good)") {
		t.Errorf("no faith held: row %q", row)
	}

	for _, c := range []struct {
		o    game.CatastropheOutlook
		want string
	}{
		{game.CatastropheOutlook{FaithStrength: 1 / 4.5, FaithDevotion: 1, FaithKept: 1}, "faith strength 22% (devotion 1.0x, 100% of your faith kept)"},
		{game.CatastropheOutlook{FaithStrength: 0.4, FaithDevotion: 3.6, FaithKept: 0.5}, "faith strength 40% (devotion 3.6x, 50% of your faith kept)"},
		{game.CatastropheOutlook{FaithStrength: 1, FaithDevotion: 12.4, FaithKept: 1}, "faith strength 100% (devotion 12x, 100% of your faith kept)"},
		{game.CatastropheOutlook{FaithStrength: 0.01, FaithDevotion: 0.04, FaithKept: 1}, "faith strength 1% (devotion 0.04x, 100% of your faith kept)"},
		// Wonders and no faith buildings: faith in hand, and nothing to show for it.
		{game.CatastropheOutlook{}, "faith strength 0% (your faith buildings have made no faith yet)"},
	} {
		if got := faithStrengthText(game.GameState{CatastropheOutlook: c.o}); got != c.want {
			t.Errorf("faithStrengthText = %q, want %q", got, c.want)
		}
	}
}

// faithRowText is the faith row as a roomy Resources box shows it (its line
// on the grid and the detail line under it), without tags, and how many
// cells its bar has.
func faithRowText(rs game.ResourceState, o game.CatastropheOutlook) (string, int) {
	rows := []resRow{faithRow(rs, o)}
	lines, _ := layoutResourceBox(rows, 60, 10, 0)
	return untag(strings.Join(lines[:2], "\n")), fitResourceFormat(rows, 60).bar
}
