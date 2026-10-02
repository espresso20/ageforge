package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fate_guard_test.go holds the no-leak rule for fated dooms (game/fate.go):
// until a doom's harbinger arrives, nothing on screen may tell whether one is
// fated. The engine keeps the fate out of GameState; this guard keeps the UI
// and the map from reaching for it another way. No non-test source under ui/
// or mapmodel/ may name the fate's bus events, its save type or the engine's
// test-only accessors for it.
func TestUINeverReadsTheHiddenFate(t *testing.T) {
	banned := []string{"EventFateRolled", "EventFateResolved", "FateSave", "FateForTest", "ForceFateForTest", "ForceQuietFateForTest", "ForceFalseProphetForTest"}
	for _, dir := range []string{".", "../mapmodel"} {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, name := range banned {
				if strings.Contains(string(src), name) {
					t.Errorf("%s names %s: the UI may not read an era's hidden fate", path, name)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
