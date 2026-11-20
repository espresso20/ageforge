package game

import (
	"path/filepath"
	"testing"
)

// A release build (DataHome "user") keeps its data in the user config folder.
func TestResolveBinaryDataDirUserHome(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv("HOME", cfg)
	t.Setenv("XDG_CONFIG_HOME", cfg)
	t.Setenv("AppData", cfg)
	old := DataHome
	t.Cleanup(func() { DataHome = old })

	DataHome = "user"
	got := resolveBinaryDataDir()
	if filepath.Base(got) != "AgeForge" || !filepath.IsAbs(got) {
		t.Fatalf("DataHome user: dir = %q, want an absolute path ending in AgeForge", got)
	}
	if rel, err := filepath.Rel(cfg, got); err != nil || rel == ".." || len(rel) > 2 && rel[:3] == "../" {
		t.Fatalf("DataHome user: dir = %q, want it under %q", got, cfg)
	}

	DataHome = ""
	if got := resolveBinaryDataDir(); filepath.Base(got) != "data" {
		t.Fatalf("DataHome empty: dir = %q, want data beside the binary", got)
	}
}
