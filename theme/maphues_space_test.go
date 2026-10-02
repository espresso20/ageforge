package theme

import "testing"

// TestSpaceHues: every sky arc colour is set and valid, and an unknown hue
// falls back instead of panicking.
func TestSpaceHues(t *testing.T) {
	for h := SpaceHue(1); h < numSpaceHues; h++ {
		if spaceHues[h] == 0 {
			t.Errorf("space hue %d has no colour", h)
		}
		if !SpaceColor(h).Valid() {
			t.Errorf("space hue %d is not a valid colour", h)
		}
	}
	if SpaceColor(numSpaceHues+3) != SpaceColor(SpaceNone) {
		t.Error("an unknown space hue does not fall back")
	}
	if NumSpaceHues != int(numSpaceHues) {
		t.Error("NumSpaceHues is stale")
	}
}
