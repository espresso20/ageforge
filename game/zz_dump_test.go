package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZZDump(t *testing.T) {
	dir := os.Getenv("GS_DUMP_DIR")
	if dir == "" {
		t.Skip()
	}
	os.MkdirAll(dir, 0o755)
	for name, ge := range snapshotCases(t) {
		if err := os.WriteFile(filepath.Join(dir, name+".txt"), []byte(dumpSnapshot(ge.GetState())), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
