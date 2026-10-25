package fixture

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// otherTheme is a theme key that is not the active one.
func otherTheme(t *testing.T) string {
	t.Helper()
	for _, th := range theme.All() {
		if th.Key != theme.Active().Key {
			return th.Key
		}
	}
	t.Fatal("only one theme")
	return ""
}

func TestWithThemeRestores(t *testing.T) {
	prev, other := theme.Active().Key, otherTheme(t)
	t.Run("sub", func(t *testing.T) {
		WithTheme(t, other)
		if got := theme.Active().Key; got != other {
			t.Errorf("the active theme is %s, want %s", got, other)
		}
	})
	if got := theme.Active().Key; got != prev {
		t.Errorf("after the test the active theme is %s, want %s again", got, prev)
	}
}

func TestThemesVisitsAllAndRestores(t *testing.T) {
	prev := theme.Active().Key
	var seen []string
	for th := range Themes(t) {
		if theme.Active().Key != th.Key {
			t.Errorf("the active theme is %s inside the loop body for %s", theme.Active().Key, th.Key)
		}
		seen = append(seen, th.Key)
	}
	var want []string
	for _, th := range theme.All() {
		want = append(want, th.Key)
	}
	if !slices.Equal(seen, want) {
		t.Errorf("visited %v, want %v", seen, want)
	}
	if got := theme.Active().Key; got != prev {
		t.Errorf("after the loop the active theme is %s, want %s again", got, prev)
	}
	for range Themes(t) {
		break
	}
	if got := theme.Active().Key; got != prev {
		t.Errorf("after a break the active theme is %s, want %s again", got, prev)
	}
}
