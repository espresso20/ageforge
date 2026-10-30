package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// vibeEngine is a new game with VibeMode on, where the hut reads "Crash Pad".
func vibeEngine(t *testing.T) *game.GameEngine {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	was := vibe.On()
	vibe.Set(true)
	t.Cleanup(func() { vibe.Set(was) })
	return game.NewGameEngine()
}

func TestVibeNames_CompletionOffersDisplayedNames(t *testing.T) {
	e := vibeEngine(t)
	comp := NewAutoCompleter(e)
	for _, line := range []string{"build ", "plan build "} {
		got := comp(line)
		if !slices.Contains(got, line+"crash_pad") {
			t.Errorf("completing %q = %v, want crash_pad", line, got)
		}
		if slices.Contains(got, line+"hut") {
			t.Errorf("completing %q = %v, offers the plain key hut", line, got)
		}
	}
	if got := comp("plan build cra"); !slices.Contains(got, "plan build crash_pad") {
		t.Errorf("completing a prefix = %v, want crash_pad", got)
	}
}

func TestVibeNames_CommandsTakeEveryForm(t *testing.T) {
	e := vibeEngine(t)
	for _, typed := range []string{"crash_pad", "Crash_Pad", "crash", "hut"} {
		if got := vibeArgs("plan", []string{"build", typed, "2"}, e); !slices.Equal(got, []string{"build", "hut", "2"}) {
			t.Errorf("plan build %s resolves to %v, want the hut", typed, got)
		}
	}
	// A name typed with a space reads as one name, and the count after it stays.
	if got := vibeArgs("build", []string{"crash", "pad", "3"}, e); !slices.Equal(got, []string{"hut", "3"}) {
		t.Errorf("build crash pad 3 resolves to %v", got)
	}
	// A word that names nothing is left for the command to refuse.
	if got := vibeArgs("build", []string{"zzzz"}, e); !slices.Equal(got, []string{"zzzz"}) {
		t.Errorf("an unknown word became %v", got)
	}
	// Words the slot takes are not names.
	if got := vibeArgs("assign", []string{"hut", "all"}, e); !slices.Equal(got, []string{"hut", "all"}) {
		t.Errorf("assign hut all = %v", got)
	}
	// The command runs with the displayed name and the plain key alike.
	for _, typed := range []string{"crash_pad", "hut"} {
		res := HandleCommand("plan build "+typed, e)
		if res.Type == "error" {
			t.Errorf("plan build %s refused: %s", typed, res.Message)
		}
	}
}

func TestVibeNames_PlainTechAndResourceNames(t *testing.T) {
	e := vibeEngine(t)
	set := e.Rules()
	for _, tc := range set.Techs()[:3] {
		typed := strings.ReplaceAll(strings.ToLower(tc.Name), " ", "_")
		got := vibeArgs("research", []string{typed}, e)
		if len(got) != 1 || got[0] != tc.Key {
			t.Errorf("research %s resolves to %v, want %s", typed, got, tc.Key)
		}
	}
}

func TestVibeNames_OffChangesNothing(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	was := vibe.On()
	vibe.Set(false)
	t.Cleanup(func() { vibe.Set(was) })
	e := game.NewGameEngine()
	if got := vibeArgs("build", []string{"crash"}, e); !slices.Equal(got, []string{"crash"}) {
		t.Errorf("with VibeMode off, build crash became %v", got)
	}
	comp := NewAutoCompleter(e)
	got := comp("build ")
	if !slices.Contains(got, "build hut") || slices.Contains(got, "build crash_pad") {
		t.Errorf("with VibeMode off, completing build = %v", got)
	}
}

func TestVibeNames_GatherTakesDisplayedNames(t *testing.T) {
	e := vibeEngine(t)
	food := e.Rules().Typed(rules.KindResource, "food")
	if food == "food" {
		t.Skip("the script has no other name for food")
	}
	if got := vibeArgs("gather", []string{food, "5"}, e); !slices.Equal(got, []string{"food", "5"}) {
		t.Errorf("gather %s resolves to %v", food, got)
	}
	got := NewAutoCompleter(e)("gather ")
	if !slices.Contains(got, "gather "+food) || slices.Contains(got, "gather food") {
		t.Errorf("completing gather = %v, want %s and not food", got, food)
	}
}
