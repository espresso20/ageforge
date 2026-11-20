package ui

import "testing"

func TestLaneSpeedClamp(t *testing.T) {
	if s := laneSpeed(0); s != 1 {
		t.Fatalf("idle lane speed %v, want 1", s)
	}
	if s := laneSpeed(1e12); s != 12 {
		t.Fatalf("top lane speed %v, want 12", s)
	}
	if laneSpeed(10) <= laneSpeed(1) {
		t.Fatal("faster flow must move faster")
	}
}
