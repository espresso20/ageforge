package ui

// Enter in the command input runs what was typed, even while the autocomplete
// dropdown is open; Tab takes the suggestion. Enter used to be taken by the
// dropdown and only rewrote "advance" to "advance ", so every command needed
// a second Enter.

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestInputEnterRunsTypedCommand(t *testing.T) {
	h := newReproHarness(t)

	// typeAndPress types text, checks the dropdown has something to offer (so
	// the key really meets an open list) and presses k once.
	complete := NewAutoCompleter(h.eng)
	typeAndPress := func(text string, k tcell.Key) {
		t.Helper()
		h.waitFor("empty focused input", 3*time.Second, func() bool { return h.inputHasFocus() && h.inputText() == "" })
		h.typeText(text)
		h.waitFor("typed "+text, 3*time.Second, func() bool { return h.inputText() == text })
		if len(complete(text)) == 0 {
			t.Fatalf("no suggestions for %q; the dropdown would not be open", text)
		}
		h.key(k, 0)
	}

	// Tab still completes, one word at a time, and Enter then runs the result.
	typeAndPress("plan bu", tcell.KeyTab)
	h.waitFor("Tab to complete the subcommand", 3*time.Second, func() bool { return h.inputText() == "plan build " })
	key := plannableBuildingKeys(h.eng.GetState())[0]
	h.typeText(key)
	h.key(tcell.KeyTab, 0)
	var picked string
	h.waitFor("Tab to complete the building", 3*time.Second, func() bool {
		picked = strings.TrimPrefix(h.inputText(), "plan build ")
		return strings.HasPrefix(picked, key) && strings.HasSuffix(picked, " ")
	})
	picked = strings.TrimSpace(picked)
	h.key(tcell.KeyEnter, 0)
	h.waitFor("plan build "+picked+" to run", 3*time.Second, func() bool {
		plan := h.eng.GetState().Plan
		return h.inputText() == "" && len(plan) == 1 && plan[0].Key == picked
	})

	// A complete multi-word command runs on the first Enter.
	typeAndPress("research list", tcell.KeyEnter)
	h.waitFor("research list to run on one Enter", 3*time.Second, func() bool { return h.inputText() == "" })
	h.waitFor("dashboard input focus", 3*time.Second, func() bool { return h.frontPage() == "dashboard" && h.inputHasFocus() })

	// advance + one Enter advances the age.
	if !h.grantNextAge() {
		t.Fatal("already at the final age")
	}
	from := h.eng.GetState().Age
	h.advanceAndCheckWith(0, 'e', func() { typeAndPress("advance", tcell.KeyEnter) })
	if got := h.eng.GetState().Age; got == from {
		t.Fatalf("age still %s after advance + Enter", got)
	}
}
