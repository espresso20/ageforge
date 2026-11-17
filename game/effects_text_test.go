package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A timed production effect is logged per second at the tick length it is
// given, never per tick.
func TestEffectsTextSaysPerSecond(t *testing.T) {
	cases := []struct {
		value, seconds float64
		want           string
	}{
		{-3, 2, "-1.5/s"},
		{5, 2, "+2.5/s"},
		{5, 0.5, "+10/s"},
		{0.01, 2, "+0.3/min"},
		{0.0001, 2, "+0.18/h"},
	}
	for _, c := range cases {
		got := effectsText([]config.Effect{{Type: "production", Target: "food", Value: c.value}}, c.seconds)
		if !strings.Contains(got, c.want) || strings.Contains(strings.ToLower(got), "tick") {
			t.Errorf("effectsText(%v at %vs) = %q, want %q and no tick", c.value, c.seconds, got, c.want)
		}
	}
}
