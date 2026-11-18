package config

import "testing"

func TestCoinLevelMarks(t *testing.T) {
	const age = "stone_age"
	ref := AgeTargetTicks(age)
	if ref <= 0 {
		t.Fatalf("no pacing target for %s", age)
	}
	cases := []struct {
		share float64
		want  int
	}{
		{0, 6}, {0.60, 6}, {0.61, 5}, {0.80, 5}, {0.81, 4},
		{1.00, 4}, {1.01, 3}, {1.30, 3}, {1.31, 2}, {1.80, 2}, {1.81, 1}, {50, 1},
	}
	for _, c := range cases {
		if got := CoinLevel(age, c.share*ref, 1); got != c.want {
			t.Errorf("share %v: level %d, want %d", c.share, got, c.want)
		}
	}
}

func TestCoinLevelSpeedHalvesTheMarks(t *testing.T) {
	const age = "iron_age"
	ref := AgeTargetTicks(age)
	if got := CoinLevel(age, 0.5*ref, 2); got != 4 {
		t.Errorf("half the reference at 2x: level %d, want 4", got)
	}
	if got := CoinLevel(age, 0.30*ref, 2); got != 6 {
		t.Errorf("0.30 of the reference at 2x: level %d, want 6", got)
	}
	if got := CoinLevel(age, 0.5*ref, 0); got != 6 {
		t.Errorf("a speed of 0 counts as 1: level %d, want 6", got)
	}
	if got := CoinLevel(age, 0.5*ref, -3); got != 6 {
		t.Errorf("a negative speed counts as 1: level %d, want 6", got)
	}
}

func TestCoinLevelUnknownAge(t *testing.T) {
	if got := CoinLevel("no_such_age", 0, 1); got != 1 {
		t.Errorf("unknown age: level %d, want 1", got)
	}
}

func TestCoinTypes(t *testing.T) {
	if len(CoinTypes) != 5 {
		t.Fatalf("%d coin types, want 5", len(CoinTypes))
	}
	for _, k := range CoinTypes {
		if !IsCoinType(k) {
			t.Errorf("%s is not a coin type", k)
		}
	}
	if IsCoinType("swift") {
		t.Error("swift is not a coin type")
	}
}
