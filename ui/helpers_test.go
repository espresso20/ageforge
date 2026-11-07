package ui

import (
	"strings"
	"testing"
)

// must fails the test at the caller's line when err is not nil.
func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// wantAll reports one error per entry of wants that text does not contain.
// where names the screen or line the text came from.
func wantAll(t testing.TB, where, text string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(text, w) {
			t.Errorf("%s: missing %q:\n%s", where, w, text)
		}
	}
}

// row is one row of the grid as text.
func (g *mGrid) row(y int) string {
	if y < 0 || y >= g.h {
		return ""
	}
	out := make([]rune, g.w)
	for x := range out {
		out[x] = g.c[y*g.w+x].r
	}
	return string(out)
}

// GetCurrent returns the current toast text or empty string if none active
func (tm *ToastManager) GetCurrent() string { return tm.CurrentFor(0) }
