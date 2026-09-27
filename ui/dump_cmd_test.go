package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestDumpWritesToDataDir: `dump` wrote to a relative data/logs, so the file
// landed wherever the game was launched from instead of beside the saves.
func TestDumpWritesToDataDir(t *testing.T) {
	root := t.TempDir()
	defer game.SetDataDirForTest(root)()
	ge := game.NewGameEngine()

	res := cmdDump(nil, ge)
	if res.Type == "error" {
		t.Fatalf("dump: %s", res.Message)
	}
	dir := filepath.Join(game.DataDir(), "logs")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || !strings.HasPrefix(entries[0].Name(), "dump_") {
		t.Fatalf("expected one dump in %s, got %v (%v)", dir, entries, err)
	}
	if !strings.HasPrefix(dir, root) {
		t.Errorf("logs dir %s is outside the data root %s", dir, root)
	}
	if !strings.Contains(res.Message, dir) {
		t.Errorf("dump message %q does not name the file's real location", res.Message)
	}
}
