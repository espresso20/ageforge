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
