package ui

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The unit of a rate: per second from 0.1 a second, per minute from 0.1 a
// minute, else per hour; zero is zero a second.
func TestRatePer(t *testing.T) {
	for _, c := range []struct {
		name string
		rate float64
		ms   int
		v    float64
		unit string
	}{
		{"one a second", 1, 1000, 1, "/s"},
		{"a slow tick halves it", 1, 2000, 0.5, "/s"},
		{"a fast tick doubles it", 3, 500, 6, "/s"},
		{"a speed bonus raises it", 1, 400, 2.5, "/s"},
		{"edge of a second", 0.1, 1000, 0.1, "/s"},
		{"under a second goes to a minute", 0.0999, 1000, 5.994, "/min"},
		{"negative minute", -0.05, 1000, -3, "/min"},
		{"edge of a minute", 0.1 / 60, 1000, 0.1, "/min"},
		{"under a minute goes to an hour", 0.0999 / 60, 1000, 5.994, "/h"},
		{"negative hour", -0.0001, 1000, -0.36, "/h"},
		{"zero", 0, 1000, 0, "/s"},
		{"negative second", -4, 2000, -2, "/s"},
		{"a tick of no known length is a second", 4, 0, 4, "/s"},
	} {
		v, unit := ratePer(c.rate, c.ms)
		if math.Abs(v-c.v) > 1e-9 || unit != c.unit {
			t.Errorf("%s: ratePer(%v, %d) = %v %s, want %v %s", c.name, c.rate, c.ms, v, unit, c.v, c.unit)
		}
	}
}

func TestRateText(t *testing.T) {
	for _, c := range []struct {
		rate float64
		ms   int
		want string
	}{
		{1, 2000, "+0.5/s"},
		{13.9, 1000, "+13.9/s"},
		{-0.4, 1000, "-0.4/s"},
		{0, 1000, "0/s"},
		{0.01, 1000, "+0.6/min"},
		{0.0001, 1000, "+0.36/h"},
		{1500, 1000, "+1.5K/s"},
		{1, 400, "+2.5/s"},
	} {
		if got := rateText(c.rate, c.ms); got != c.want {
			t.Errorf("rateText(%v, %d) = %q, want %q", c.rate, c.ms, got, c.want)
		}
	}
	if got := rateAmount(-3, 1000); got != "3/s" {
		t.Errorf("rateAmount(-3) = %q", got)
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
		{"+0.004 faith/tick", "+0.24 faith/min", 1000},
		{"Food -4.8/tick for ~10m.", "Food -4.8/s for ~10m.", 1000},
		{"Food +1/tick, stone +0.5/tick for 5m.", "Food +1/s, stone +0.5/s for 5m.", 1000},
		{"+0.5 opinion/tick per worker", "+0.5 opinion/s per worker", 1000},
		{"+240 housing", "+240 housing", 1000},
		{"+2% all production", "+2% all production", 1000},
	} {
		if got := perSecondText(c.in, c.ms); got != c.want {
			t.Errorf("perSecondText(%q, %d) = %q, want %q", c.in, c.ms, got, c.want)
		}
	}
}

// The Resources box writes every rate through the same helper.
func TestResourceRowsWriteRatesPerSecond(t *testing.T) {
	for _, c := range []struct {
		rate float64
		ms   int
		want string
	}{
		{1, 2000, "+0.5/s"},
		{13.9, 1000, "+13.9/s"},
		{-0.04, 1000, "-2.4/min"},
		{0.0001, 1000, "+0.36/h"},
		{0, 1000, "0/s"},
	} {
		row := storeRow(game.ResourceState{Name: "Food", Amount: 5, Rate: c.rate}, c.ms)
		if got, _ := (resFormat{}).rateText(row); got != c.want {
			t.Errorf("a rate of %v a tick of %d ms reads %q, want %q", c.rate, c.ms, got, c.want)
		}
	}
	row := storeRow(game.ResourceState{Name: "Food", Amount: 5, Rate: 0.01}, 1000)
	if got, _ := (resFormat{short: true}).rateText(row); got != "+0.6/m" {
		t.Errorf("the short unit reads %q", got)
	}
}

// No rate on the dashboard or the Economy box is in ticks.
func TestScreensShowNoRateInTicks(t *testing.T) {
	keepTheme(t)
	for _, age := range []string{"primitive_age", "medieval_age", "quantum_age"} {
		d, pages, _ := stagedDashboard(t, age)
		d.refresh()
		for _, line := range paintAlone(pages, 200, 60).Lines() {
			if strings.Contains(line, "/tick") || strings.Contains(line, "a tick") {
				t.Errorf("%s: a rate is in ticks: %q", age, strings.TrimSpace(line))
			}
		}
	}
}
