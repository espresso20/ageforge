package fixture

import (
	"iter"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// WithTheme makes the theme key the active one for the rest of the test. The
// theme that was active comes back when the test ends, pass or fail.
func WithTheme(t testing.TB, key string) {
	t.Helper()
	prev := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(prev) })
	if err := theme.SetActive(key); err != nil {
		t.Fatal(err)
	}
}

// Themes steps through every theme, each one active while the loop body
// runs. The theme that was active comes back when the loop ends, however it
// ends (a break, a failed test, a panic).
func Themes(t testing.TB) iter.Seq[theme.Theme] {
	return func(yield func(theme.Theme) bool) {
		prev := theme.Active().Key
		defer func() { _ = theme.SetActive(prev) }()
		for _, th := range theme.All() {
			if err := theme.SetActive(th.Key); err != nil {
				t.Fatal(err)
			}
			if !yield(th) {
				return
			}
		}
	}
}
