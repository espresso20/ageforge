package game

import (
	"strings"
	"testing"
)

// pathHostile is the set of characters that must never appear in a generated
// save name, since the name becomes a save filename (<name>.json).
const pathHostile = `/\:.'"*?<>|`

func TestGenerateSaveNameFilesystemSafe(t *testing.T) {
	for i := 0; i < 500; i++ {
		name := GenerateSaveName()
		if name == "" {
			t.Fatalf("GenerateSaveName() returned empty string")
		}
		if strings.ContainsAny(name, pathHostile) {
			t.Fatalf("GenerateSaveName() = %q contains a path-hostile char from %q", name, pathHostile)
		}
		if name != strings.TrimSpace(name) {
			t.Fatalf("GenerateSaveName() = %q has leading/trailing whitespace", name)
		}
	}
}

func TestGenerateSaveNameVariety(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 200; i++ {
		seen[GenerateSaveName()] = struct{}{}
	}
	if len(seen) <= 5 {
		t.Errorf("GenerateSaveName() produced only %d distinct values across 200 calls, want > 5", len(seen))
	}
}

// TestGenerateSaveNameDistribution is the regression guard against connective
// clustering. The old structure produced "Grand Duchy of" in ~11% of names and
// a "The " prefix in ~33%; both would fail the thresholds below. The flattened
// design (varied polity prefixes + optional article + more templates) passes
// these with wide margin at N=30000, where variance is negligible.
func TestGenerateSaveNameDistribution(t *testing.T) {
	const n = 30000

	twoWordPrefix := make(map[string]int)
	grandDuchy := 0
	therePrefix := 0
	distinct := make(map[string]struct{})

	for i := 0; i < n; i++ {
		name := GenerateSaveName()
		distinct[name] = struct{}{}

		if strings.HasPrefix(name, "The ") {
			therePrefix++
		}
		if strings.HasPrefix(name, "Grand Duchy of") {
			grandDuchy++
		}

		fields := strings.Fields(name)
		if len(fields) >= 2 {
			twoWordPrefix[fields[0]+" "+fields[1]]++
		}
	}

	// Most common two-word prefix must be < 4% — kills the old 11% cluster.
	var topPrefix string
	var topCount int
	for p, c := range twoWordPrefix {
		if c > topCount {
			topPrefix, topCount = p, c
		}
	}
	if frac := float64(topCount) / float64(n); frac >= 0.04 {
		t.Errorf("top two-word prefix %q = %.2f%% of names, want < 4%%", topPrefix, frac*100)
	}

	if frac := float64(grandDuchy) / float64(n); frac >= 0.03 {
		t.Errorf("%q prefix = %.2f%% of names, want < 3%%", "Grand Duchy of", frac*100)
	}

	if frac := float64(therePrefix) / float64(n); frac >= 0.22 {
		t.Errorf("%q prefix = %.2f%% of names, want < 22%%", "The ", frac*100)
	}

	if frac := float64(len(distinct)) / float64(n); frac <= 0.70 {
		t.Errorf("distinct names = %.2f%% of %d, want > 70%%", frac*100, n)
	}
}
