package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// historyRecorder holds n history samples: population flat at 836, the food
// rate rising, and two age advances marked by key, as the engine stores them.
func historyRecorder(n int) *game.RunHistory {
	h := game.NewRunHistory()
	for i := 0; i < n; i++ {
		h.Record(10*(i+1), map[string]float64{"population": 836, "food_rate": float64(i), "know_rate": 111 + float64(i)/float64(n)})
	}
	h.Mark(40, game.MarkAge, "stone_age", 0)
	h.Mark(120, game.MarkAge, "bronze_age", 0)
	return h
}

// TestHistoryFooterNamesAges: the footer under the graphs names the ages
// advanced into, not their keys.
func TestHistoryFooterNamesAges(t *testing.T) {
	out := untag(renderHistoryOverlay(game.GameState{}, historyRecorder(30), 100))
	if !strings.Contains(out, "Ages: Stone Age → Bronze Age") {
		t.Errorf("the footer should name the ages:\n%s", lineContaining(out, "Ages:"))
	}
	for _, key := range []string{"stone_age", "bronze_age"} {
		if strings.Contains(out, key) {
			t.Errorf("the History panel prints the key %q", key)
		}
	}
}

// TestHistoryAxisLabels: a flat series is labeled once, on the base line it
// is drawn along, and a label never repeats the one under it.
func TestHistoryAxisLabels(t *testing.T) {
	for _, c := range []struct {
		min, max         float64
		top, mid, bottom string
	}{
		{836, 836, "", "", "836"},    // flat
		{0, 34, "34", "17", "0"},     // an ordinary series
		{111, 111.4, "", "", "111"},  // moves less than a label can show
		{499, 500, "500", "", "499"}, // the middle would repeat the top
		{11.2, 13.9, "13.9", "12.6", "11.2"},
	} {
		top, mid, bottom := historyAxisLabels(c.min, c.max)
		if top != c.top || mid != c.mid || bottom != c.bottom {
			t.Errorf("labels for %v to %v: got %q %q %q, want %q %q %q", c.min, c.max, top, mid, bottom, c.top, c.mid, c.bottom)
		}
	}
	// On the panel: the flat population graph carries its number once.
	out := untag(renderHistoryOverlay(game.GameState{}, historyRecorder(30), 100))
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "Population") {
			continue
		}
		n := 0
		for _, g := range lines[i+1 : i+6] {
			if strings.HasPrefix(strings.TrimSpace(g), "836") {
				n++
			}
		}
		if n != 1 {
			t.Errorf("the flat Population graph labels its axis %d times, want once:\n%s", n, strings.Join(lines[i:i+6], "\n"))
		}
	}
}

// TestHistoryNegativeRate: a series that dips below zero draws without
// panicking and its header keeps the sign.
func TestHistoryNegativeRate(t *testing.T) {
	h := game.NewRunHistory()
	for i, v := range []float64{1, -3.2, -1.5} {
		h.Record(10*(i+1), map[string]float64{"population": 5, "food_rate": v})
	}
	h.Mark(20, game.MarkAge, "bronze_age", 0)
	out := untag(renderHistoryOverlay(game.GameState{}, h, 100))
	if !strings.Contains(out, "Food rate") || !strings.Contains(out, "-1.5") {
		t.Errorf("the negative food rate is not shown:\n%s", out)
	}
}
