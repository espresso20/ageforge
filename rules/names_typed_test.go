package rules

import (
	"testing"

	"github.com/espresso20/ageforge/vibe"
)

func TestKeyFor_NamesKeysAndPlainNames(t *testing.T) {
	was := vibe.On()
	vibe.Set(true)
	defer vibe.Set(was)
	s := CoreFor(true)
	if got := s.Typed(KindBuilding, "hut"); got != "crash_pad" {
		t.Fatalf("Typed hut = %q, want crash_pad", got)
	}
	for _, typed := range []string{"hut", "crash_pad", "Crash Pad", "crash"} {
		if k, ok := s.KeyFor(KindBuilding, typed, true); !ok || k != "hut" {
			t.Errorf("KeyFor(%q) = %q, %v, want hut", typed, k, ok)
		}
	}
	if _, ok := s.KeyFor(KindBuilding, "crash", false); ok {
		t.Error("a first word alone named a building without prefix matching")
	}
	vibe.Set(false)
	if got := s.Typed(KindBuilding, "hut"); got != "hut" {
		t.Errorf("with VibeMode off, Typed = %q", got)
	}
}
