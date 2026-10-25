package game

import (
	"strings"
	"testing"
)

// closestKey breaks a tie by sorted key, so a mistyped key gets the same
// "Did you mean" suggestion every run (map order must not leak in).
func TestClosestKey_TieBreaksBySortedKey(t *testing.T) {
	keys := map[string]int{"bb": 1, "ba": 2, "bd": 3, "zz": 4}
	for i := 0; i < 50; i++ {
		if got := closestKey("bc", keys); got != "ba" {
			t.Fatalf("run %d: closestKey(bc) = %q, want the sorted-first tie %q", i, got, "ba")
		}
	}
	if got := closestKey("farm", keys); got != "" {
		t.Errorf("nothing within 2 edits should suggest nothing, got %q", got)
	}
}

// A mistyped building key is refused with the nearest real building, and one
// that is nothing like any building gets the plain hint.
func TestUnknownBuildingErr_Suggests(t *testing.T) {
	ge := NewGameEngine()
	err := ge.unknownBuildingErr("farmm")
	if err == nil || !strings.Contains(err.Error(), "Did you mean 'farm'?") {
		t.Errorf("a near miss on a building: %v", err)
	}
	err = ge.unknownBuildingErr("qqqqqqqqqqqq")
	if err == nil || strings.Contains(err.Error(), "Did you mean") {
		t.Errorf("a far miss should suggest nothing: %v", err)
	}
}
