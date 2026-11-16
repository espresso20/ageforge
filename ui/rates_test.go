package ui

import "testing"

func TestPerSecond(t *testing.T) {
	for _, c := range []struct {
		rate float64
		ms   int
		v    float64
		unit string
	}{
		{1, 1000, 1, "/s"},
		{1, 2000, 0.5, "/s"},
		{3, 500, 6, "/s"},
		{0.0001, 1000, 0.006, "/min"}, // 0.0001 a second is 0.006 a minute
		{-0.0001, 1000, -0.006, "/min"},
		{0, 1000, 0, "/s"},
		{0.01, 1000, 0.01, "/s"},
		{4, 0, 4, "/s"}, // a tick of no known length counts as a second
	} {
		v, unit := perSecond(c.rate, c.ms)
		if diff := v - c.v; diff > 1e-9 || diff < -1e-9 || unit != c.unit {
			t.Errorf("perSecond(%v, %d) = %v %s, want %v %s", c.rate, c.ms, v, unit, c.v, c.unit)
		}
	}
}

func TestPerSecondText(t *testing.T) {
	for _, c := range []struct {
		in, want string
		ms       int
	}{
		{"+1 culture/tick", "+0.5 culture/s", 2000},
		{"+73.5 iron/tick (6 workers)", "+73.5 iron/s (6 workers)", 1000},
		{"+1.2K iron ore/tick, -3 wood/tick", "+1.2K iron ore/s, -3 wood/s", 1000},
		{"+0.0004 faith/tick", "+0.024 faith/min", 1000},
		{"+240 housing", "+240 housing", 1000},
		{"+2% all production", "+2% all production", 1000},
	} {
		if got := perSecondText(c.in, c.ms); got != c.want {
			t.Errorf("perSecondText(%q, %d) = %q, want %q", c.in, c.ms, got, c.want)
		}
	}
}
