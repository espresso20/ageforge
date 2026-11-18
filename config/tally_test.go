package config

import "testing"

// Rung r of n strikes level 1 + round((r-1)*5/(n-1)), halves up.
func TestTallyLevelFollowsTheRungRule(t *testing.T) {
	cases := []struct {
		rungs []float64
		want  []int // the level at each rung, in order
	}{
		{[]float64{1, 5, 11, 22, 33, 44}, []int{1, 2, 3, 4, 5, 6}},
		{[]float64{1, 3, 6, 9, 12}, []int{1, 2, 4, 5, 6}},
		{[]float64{10}, []int{1}},
		{[]float64{1, 2, 3}, []int{1, 4, 6}},
	}
	for _, c := range cases {
		td := TallyDef{Rungs: c.rungs}
		if got := td.Level(c.rungs[0] - 1); got != 0 {
			t.Errorf("rungs %v: below the first rung is level %d, want 0", c.rungs, got)
		}
		for i, r := range c.rungs {
			if got := td.Level(r); got != c.want[i] {
				t.Errorf("rungs %v: rung %d is level %d, want %d", c.rungs, i+1, got, c.want[i])
			}
			if got := td.Level(r + 0.5); got != c.want[i] {
				t.Errorf("rungs %v: just past rung %d is level %d, want %d", c.rungs, i+1, got, c.want[i])
			}
		}
	}
}

func TestTallyGiftFindsTheRungLevel(t *testing.T) {
	ts := Tallies{{Key: "ledger.endured", Rungs: []float64{1, 3, 6, 9, 12}}}
	if k, l, ok := ts.Gift("ledger.endured.3"); !ok || k != "ledger.endured" || l != 4 {
		t.Errorf("gift ledger.endured.3 = %q %d %v, want ledger.endured 4 true", k, l, ok)
	}
	if _, _, ok := ts.Gift("ledger.endured.9"); ok {
		t.Error("a rung past the table gave a gift")
	}
	if _, _, ok := ts.Gift("age.stone_age"); ok {
		t.Error("a badge key gave a gift")
	}
}
