package theme

import "testing"

// TestDuotoneThemesSpanARamp: a Duotone theme's pictures (the skyline map)
// fold every color onto the ramp from Background to Text, so the two ends
// must be far enough apart for a picture to survive the fold, and at least
// one theme must use it.
func TestDuotoneThemesSpanARamp(t *testing.T) {
	n := 0
	for _, th := range All() {
		if !th.Duotone {
			continue
		}
		n++
		if r := ContrastRatio(th.Color(RoleBackground), th.Color(RoleText)); r < 7 {
			t.Errorf("%s is Duotone but Text on Background is only %.1f:1", th.Key, r)
		}
	}
	if n == 0 {
		t.Error("no theme is Duotone")
	}
}
