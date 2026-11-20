package config

import "testing"

func TestRankFor(t *testing.T) {
	for pts, want := range map[int]int{0: 0, 24: 0, 25: 1, 99: 1, 100: 2, 190: 3, 320: 4, 459: 4, 460: 5, 700: 6, 5000: 6} {
		if got := RankFor(pts); got != want {
			t.Errorf("RankFor(%d) = %d, want %d", pts, got, want)
		}
	}
}
