package ui

import (
	"strings"
	"testing"
)

func TestComputeMoraleBand_Bonus(t *testing.T) {
	// High band: morale 0.88, multiplier 1.18 → green, "+18%".
	b := computeMoraleBand(0.88, 1.18)
	if !b.Bonus || b.Penalty {
		t.Fatalf("expected bonus band, got %+v", b)
	}
	if b.Color != "green" {
		t.Errorf("bonus color = %q, want green", b.Color)
	}
	if b.DeltaPct != 18 {
		t.Errorf("bonus delta = %d, want 18", b.DeltaPct)
	}
	if !strings.Contains(b.Status, "88%") || !strings.Contains(b.Status, "+18%") {
		t.Errorf("bonus status = %q, want it to mention 88%% and +18%%", b.Status)
	}
	if strings.Contains(strings.ToLower(b.Status), "penalty") {
		t.Errorf("bonus status should not mention penalty: %q", b.Status)
	}
}

func TestComputeMoraleBand_Neutral(t *testing.T) {
	// Only an exact-1.0 multiplier is neutral now, which the continuous curve
	// produces solely at the 0.50 pivot → neutral color, "steady", no penalty.
	b := computeMoraleBand(0.50, 1.0)
	if b.Bonus || b.Penalty {
		t.Fatalf("expected neutral band, got %+v", b)
	}
	if b.Color != "white" {
		t.Errorf("neutral color = %q, want white", b.Color)
	}
	if b.DeltaPct != 0 {
		t.Errorf("neutral delta = %d, want 0", b.DeltaPct)
	}
	if !strings.Contains(b.Status, "50%") || !strings.Contains(b.Status, "steady") {
		t.Errorf("neutral status = %q, want it to mention 50%% and steady", b.Status)
	}
	if strings.Contains(strings.ToLower(b.Status), "penalty") {
		t.Errorf("neutral status should not mention penalty: %q", b.Status)
	}
}

func TestComputeMoraleBand_NearPivotHonest(t *testing.T) {
	// A real bonus that rounds to 0% must read "+<1%", not a dishonest "+0%".
	b := computeMoraleBand(0.51, 1.004)
	if !b.Bonus || b.Penalty {
		t.Fatalf("expected bonus band just above pivot, got %+v", b)
	}
	if !strings.Contains(b.Status, "+<1%") {
		t.Errorf("near-pivot bonus status = %q, want it to contain +<1%%", b.Status)
	}
	// A real penalty that rounds to 0% must read "-<1%".
	b = computeMoraleBand(0.49, 0.996)
	if b.Bonus || !b.Penalty {
		t.Fatalf("expected penalty band just below pivot, got %+v", b)
	}
	if !strings.Contains(b.Status, "-<1%") {
		t.Errorf("near-pivot penalty status = %q, want it to contain -<1%%", b.Status)
	}
}

func TestComputeMoraleBand_Penalty(t *testing.T) {
	// Low band: morale 0.20, multiplier 0.66 → red, "−34%".
	b := computeMoraleBand(0.20, 0.66)
	if b.Bonus || !b.Penalty {
		t.Fatalf("expected penalty band, got %+v", b)
	}
	if b.Color != "red" {
		t.Errorf("penalty color = %q, want red", b.Color)
	}
	if b.DeltaPct != -34 {
		t.Errorf("penalty delta = %d, want -34", b.DeltaPct)
	}
	if !strings.Contains(b.Status, "20%") || !strings.Contains(b.Status, "-34%") {
		t.Errorf("penalty status = %q, want it to mention 20%% and -34%%", b.Status)
	}
}

func TestRoundHalf(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{17.4, 17},
		{17.6, 18},
		{-33.6, -34},
		{-33.4, -33},
		{0.0, 0},
		{0.5, 1},
		{-0.5, -1},
	}
	for _, c := range cases {
		if got := roundHalf(c.in); got != c.want {
			t.Errorf("roundHalf(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestAbsFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{1.5, 1.5},
		{-1.5, 1.5},
		{0.0, 0.0},
		{-0.0, 0.0},
	}
	for _, c := range cases {
		if got := absFloat(c.in); got != c.want {
			t.Errorf("absFloat(%v) = %v, want %v", c.in, got, c.want)
		}
	}

	// Guard the epsilon used by the Active Multipliers "is morale active" check:
	// a neutral 1.0 multiplier must read as inactive, a +18% one as active.
	const eps = 0.0005
	if absFloat(1.0-1.0) > eps {
		t.Error("neutral multiplier (1.0) should be treated as inactive")
	}
	if !(absFloat(1.18-1.0) > eps) {
		t.Error("bonus multiplier (1.18) should be treated as active")
	}
}
