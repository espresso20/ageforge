package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/uniseg"
)

func TestRenderDimensions(t *testing.T) {
	// Test seeds 1..40
	for seed := int64(1); seed <= 40; seed++ {
		m := generate(seed)
		lines := render(m, false, 60, 16)

		// Should have 37 lines (1 top + 32 grid + 1 middle + 1 metrics + 1 inspect + 1 bottom)
		if len(lines) != 37 {
			t.Errorf("seed %d: expected 37 lines, got %d", seed, len(lines))
		}

		// Each line should be exactly 122 cells wide
		for i, line := range lines {
			width := uniseg.StringWidth(line)
			if width != 122 {
				t.Errorf("seed %d line %d: expected width 122, got %d", seed, i, width)
			}
		}

		// Grid rows (lines 1..32) should start and end with │
		for i := 1; i <= 32; i++ {
			if !strings.HasPrefix(lines[i], "│") {
				t.Errorf("seed %d line %d: doesn't start with │", seed, i)
			}
			if !strings.HasSuffix(lines[i], "│") {
				t.Errorf("seed %d line %d: doesn't end with │", seed, i)
			}
		}
	}
}

func TestDeterministic(t *testing.T) {
	m1 := generate(42)
	lines1 := render(m1, false, 60, 16)

	m2 := generate(42)
	lines2 := render(m2, false, 60, 16)

	if len(lines1) != len(lines2) {
		t.Fatalf("same seed produces different number of lines")
	}

	for i := range lines1 {
		if lines1[i] != lines2[i] {
			t.Errorf("line %d differs between same seed", i)
		}
	}
}

func TestDifferentSeeds(t *testing.T) {
	lines42 := render(generate(42), false, 60, 16)
	lines7 := render(generate(7), false, 60, 16)

	// Grid rows should differ in at least 50% (nebula and stars must follow the seed) of cells
	gridDiffs := 0
	gridCells := 32 * 120

	for i := 1; i <= 32; i++ {
		for j := 1; j <= 120; j++ {
			if j < len(lines42[i]) && j < len(lines7[i]) {
				if lines42[i][j:j+1] != lines7[i][j:j+1] {
					gridDiffs++
				}
			}
		}
	}

	percentDiff := float64(gridDiffs) / float64(gridCells)
	if percentDiff < 0.50 {
		t.Errorf("only %.1f%% of cells differ between seed 42 and seed 7 (expected >= 50%%)", percentDiff*100)
	}
}

func TestNoLetters(t *testing.T) {
	m := generate(42)
	lines := render(m, false, 60, 16)

	for i := 1; i <= 32; i++ {
		for _, ch := range lines[i] {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				t.Errorf("line %d contains letter %c", i, ch)
			}
		}
	}
}

func TestGateAppears(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		m := generate(seed)
		lines := render(m, false, 60, 16)

		gateCount := 0
		for i := 1; i <= 32; i++ {
			for _, ch := range lines[i] {
				if ch == '⟐' {
					gateCount++
				}
			}
		}

		if gateCount != 1 {
			t.Errorf("seed %d: expected gate '⟐' to appear exactly once, got %d", seed, gateCount)
		}
	}
}

func TestLaneMetrics(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		m := generate(seed)
		lines := render(m, false, 60, 16)

		// Metrics line is line 34 (1 top + 32 grid + 1 middle)
		metricsLine := lines[34]

		// Extract lanes count from metrics line
		// Format: "│ stars X · stations Y · lanes Z · freighters W · rift Q│"
		parts := strings.Split(metricsLine, "·")
		if len(parts) < 3 {
			t.Errorf("seed %d: couldn't parse metrics line", seed)
			continue
		}

		lanesPart := strings.TrimSpace(parts[2])
		var laneCount int
		if _, err := fmt.Sscanf(lanesPart, "lanes %d", &laneCount); err != nil {
			t.Errorf("seed %d: couldn't parse lanes count: %v", seed, err)
			continue
		}

		if laneCount < 4 {
			t.Errorf("seed %d: expected lanes >= 4, got %d", seed, laneCount)
		}
	}
}

func TestFreighterMetrics(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		m := generate(seed)
		lines := render(m, false, 60, 16)

		// Metrics line is line 34 (1 top + 32 grid + 1 middle)
		metricsLine := lines[34]

		// Extract freighters count from metrics line
		// Format: "│ stars X · stations Y · lanes Z · freighters W · rift Q│"
		parts := strings.Split(metricsLine, "·")
		if len(parts) < 4 {
			t.Errorf("seed %d: couldn't parse metrics line", seed)
			continue
		}

		freightersPart := strings.TrimSpace(parts[3])
		var freighterCount int
		if _, err := fmt.Sscanf(freightersPart, "freighters %d", &freighterCount); err != nil {
			t.Errorf("seed %d: couldn't parse freighters count: %v", seed, err)
			continue
		}

		if freighterCount < 3 {
			t.Errorf("seed %d: expected freighters >= 3, got %d", seed, freighterCount)
		}
	}
}
