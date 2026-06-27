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
