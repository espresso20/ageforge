package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestMasterySectionNoSpoilers: the Era Mastery table names ages only up to
// the record, and shows each known age's mastery and speed.
func TestMasterySectionNoSpoilers(t *testing.T) {
	speeds := map[string]float64{}
	for _, a := range config.AgeOrder() {
		speeds[a] = 1
	}
	speeds["primitive_age"], speeds["stone_age"] = config.MasteryK(2), config.MasteryK(1)
	state := game.GameState{
		Age: "stone_age",
		Mastery: game.MasteryState{
			K: config.MasteryK(1), Level: 1, Record: "bronze_age", RunFurthest: "stone_age",
			Ages:      map[string]int{"primitive_age": 2, "stone_age": 1},
			Speeds:    speeds,
			NextGains: []string{"primitive_age"},
		},
	}
	var sb strings.Builder
	writeMasterySection(&sb, state)
	out := sb.String()
	for _, want := range []string{"Primitive Age", "Bronze Age", "2.4x", "mastery  2", "known ground, 2x faster (mastery 1)", "Next prestige: the Primitive Age gains a mastery level"} {
		if !strings.Contains(out, want) {
			t.Errorf("the mastery section lacks %q:\n%s", want, out)
		}
	}
	for _, a := range config.AgeOrder()[3:] {
		if name := game.AgeName(a); strings.Contains(out, name) {
			t.Errorf("the mastery section names %s, past the record:\n%s", name, out)
		}
	}
}
