package main

import (
	"strings"
	"testing"

	"github.com/rivo/uniseg"
)

func TestRenderDimensions(t *testing.T) {
	// Test seeds 1..40
	for seed := int64(1); seed <= 40; seed++ {
		m := generate(seed)
		lines := render(m, false, 60, 16)

		// Should have 38 lines (1 top + 32 grid + 1 middle + 1 metrics + 1 inspect + 1 bottom)
		if len(lines) != 38 {
			t.Errorf("seed %d: expected 38 lines, got %d", seed, len(lines))
		}

		// Each line should be exactly 122 cells wide (except the last blank line)
		for i, line := range lines {
			if i == len(lines)-1 && line == "" {
				// Skip blank line at end
				continue
			}
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

	// Grid rows should differ in at least 20% of cells
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
	if percentDiff < 0.20 {
		t.Errorf("only %.1f%% of cells differ between seed 42 and seed 7 (expected >= 20%%)", percentDiff*100)
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
