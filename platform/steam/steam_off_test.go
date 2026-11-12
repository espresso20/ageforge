//go:build !steam

package steam

import "testing"

func TestOffBuildIsInert(t *testing.T) {
	if Build {
		t.Fatal("Build is true without the steam tag")
	}
	if err := Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	Unlock("anything")
	Frame()
}

func TestOffBuildHasNoDataDir(t *testing.T) {
	if d := DataDir(); d != "" {
		t.Fatalf("DataDir = %q without the steam tag, want empty", d)
	}
}
