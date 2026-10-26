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
