package config

import (
	"strings"
	"testing"
)

// TestLogFlavorPoolsAreSane verifies every flavour pool is non-empty and every
// line is plain, safe text. The engine wraps each line as fmt.Sprintf("  [gray]%s[-]", q),
// so the line itself must carry no tview tags (square brackets) and no stray
// printf directives — either would corrupt the rendered log entry.
func TestLogFlavorPoolsAreSane(t *testing.T) {
	if len(logFlavorPools) == 0 {
		t.Fatal("logFlavorPools is empty")
	}
	for moment, pool := range logFlavorPools {
		if len(pool) == 0 {
			t.Errorf("log flavor moment %q has an empty pool", moment)
		}
		// A single-variant pool can never vary; require at least two so the
		// "picked at random, stays fresh" promise actually holds.
		if len(pool) < 2 {
			t.Errorf("log flavor moment %q has only %d variant(s); want >= 2 for variety", moment, len(pool))
		}
		seen := make(map[string]bool, len(pool))
		for i, line := range pool {
			if strings.TrimSpace(line) == "" {
				t.Errorf("log flavor moment %q line %d is empty/blank", moment, i)
			}
			if strings.ContainsAny(line, "[]") {
				t.Errorf("log flavor moment %q line %d has a square bracket (tview tag risk): %q", moment, i, line)
			}
			if strings.Contains(line, "%") {
				t.Errorf("log flavor moment %q line %d has a %% directive: %q", moment, i, line)
			}
			if seen[line] {
				t.Errorf("log flavor moment %q has a duplicate line: %q", moment, line)
			}
			seen[line] = true
		}
	}
}

// TestPickLogFlavorReturnsValidLine confirms the picker only ever returns a
// line that actually lives in the requested moment's pool (and never panics).
func TestPickLogFlavorReturnsValidLine(t *testing.T) {
	for _, moment := range LogFlavorMoments() {
		valid := make(map[string]bool)
		for _, line := range logFlavorPools[moment] {
			valid[line] = true
		}
		// Sample many times to exercise the rand path across the whole pool.
		for i := 0; i < 200; i++ {
			got := PickLogFlavor(moment)
			if got == "" {
				t.Fatalf("PickLogFlavor(%q) returned empty for a populated pool", moment)
			}
			if !valid[got] {
				t.Fatalf("PickLogFlavor(%q) returned %q, not in pool", moment, got)
			}
		}
	}
}

// TestPickLogFlavorVaries ensures the picker actually draws different lines over
// repeated calls — a degenerate picker that always returns pool[0] would defeat
// the entire point of the per-moment pools.
func TestPickLogFlavorVaries(t *testing.T) {
	// building_complete has the largest pool; over 300 draws we should see
	// several distinct lines unless the RNG or picker is broken.
	seen := make(map[string]bool)
	for i := 0; i < 300; i++ {
		seen[PickLogFlavor(LogFlavorBuildingComplete)] = true
	}
	if len(seen) < 2 {
		t.Errorf("PickLogFlavor never varied over 300 draws (saw %d distinct lines)", len(seen))
	}
}

// TestPickLogFlavorUnknownMoment confirms an unregistered or empty moment key
// degrades gracefully to "" rather than panicking — the engine guards on this.
func TestPickLogFlavorUnknownMoment(t *testing.T) {
	if got := PickLogFlavor("no_such_moment"); got != "" {
		t.Errorf("PickLogFlavor(unknown) = %q, want empty string", got)
	}
	if got := PickLogFlavor(""); got != "" {
		t.Errorf("PickLogFlavor(empty) = %q, want empty string", got)
	}
}
