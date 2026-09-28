package ui

// The prompt's keys on the real App: Enter runs a whole command as typed and
// an unfinished one's ghost completion (a Dangerous completion only fills
// the field), Tab takes and cycles completions, ↑/↓ stay command history.
// Enter once went to the autocomplete dropdown instead and only rewrote
// "advance" to "advance ", so every command needed a second Enter (#131).

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestPromptKeys(t *testing.T) {
	h := newReproHarness(t)

	ready := func() {
		t.Helper()
		h.waitFor("empty focused input", 3*time.Second, func() bool { return h.inputHasFocus() && h.inputText() == "" })
	}
	typed := func(text string) {
		t.Helper()
		ready()
		h.typeText(text)
		h.waitFor("typed "+text, 3*time.Second, func() bool { return h.inputText() == text })
	}
	clear := func() { h.onUI(func() { h.a.dashboard.inputField.SetText("") }) }
	history := func() []string {
		var out []string
		h.onUI(func() { out = append(out, h.a.dashboard.cmdHistory...) })
		return out
	}
	logged := func(sub string) int {
		n := 0
		for _, l := range h.eng.GetLogs() {
			if strings.Contains(l.Message, sub) {
				n++
			}
		}
		return n
	}
	lastRun := func() string {
		hs := history()
		if len(hs) == 0 {
			return ""
		}
		return hs[len(hs)-1]
	}

	// adv + Tab gives advance, with no trailing space: nothing follows it.
	typed("adv")
	h.key(tcell.KeyTab, 0)
	h.waitFor("Tab to complete adv", 3*time.Second, func() bool { return h.inputText() == "advance" })
	clear()

	// → at the end of the line takes the ghost too.
	typed("rat")
	h.key(tcell.KeyRight, 0)
	h.waitFor("→ to complete rat", 3*time.Second, func() bool { return h.inputText() == "rates" })
	clear()

	// Tab again cycles the candidates: re → recruit, research, ...
	typed("re")
	h.key(tcell.KeyTab, 0)
	h.waitFor("first candidate", 3*time.Second, func() bool { return h.inputText() == "recruit " })
	h.key(tcell.KeyTab, 0)
	h.waitFor("second candidate", 3*time.Second, func() bool { return h.inputText() == "research " })
	h.key(tcell.KeyBacktab, 0)
	h.waitFor("Backtab back to the first", 3*time.Second, func() bool { return h.inputText() == "recruit " })
	clear()

	// A complete multi-word command runs on the first Enter.
	typed("research list")
	h.key(tcell.KeyEnter, 0)
	h.waitFor("research list to run on one Enter", 3*time.Second, func() bool { return h.inputText() == "" && lastRun() == "research list" })
	h.waitFor("dashboard input focus", 3*time.Second, func() bool { return h.frontPage() == "dashboard" && h.inputHasFocus() })

	// plan bu + Tab + a building + Tab + Enter adds a plan item.
	typed("plan bu")
	h.key(tcell.KeyTab, 0)
	h.waitFor("Tab to complete the subcommand", 3*time.Second, func() bool { return h.inputText() == "plan build " })
	key := plannableBuildingKeys(h.eng.GetState())[0]
	h.typeText(key[:2])
	h.key(tcell.KeyTab, 0)
	var picked string
	h.waitFor("Tab to complete the building", 3*time.Second, func() bool {
		picked = strings.TrimPrefix(h.inputText(), "plan build ")
		return strings.HasPrefix(picked, key[:2]) && strings.HasSuffix(picked, " ")
	})
	picked = strings.TrimSpace(picked)
	// (The plan may start the item at once, and then it leaves the plan:
	// watch the log line instead.)
	h.key(tcell.KeyEnter, 0)
	h.waitFor("plan build "+picked+" to run", 3*time.Second, func() bool {
		return h.inputText() == "" && lastRun() == "plan build "+picked && logged("Planned 1 × ") == 1
	})

	// A Dangerous partial + Enter fills in the command without running it; a
	// second Enter runs it.
	typed("plan cle")
	h.key(tcell.KeyEnter, 0)
	h.waitFor("plan cle + Enter to fill in plan clear", 3*time.Second, func() bool { return h.inputText() == "plan clear" })
	time.Sleep(100 * time.Millisecond)
	if lastRun() == "plan clear" || logged("Cleared the plan") != 0 {
		t.Fatal("plan clear ran from a completion")
	}
	h.key(tcell.KeyEnter, 0)
	h.waitFor("the second Enter to clear the plan", 3*time.Second, func() bool {
		return h.inputText() == "" && lastRun() == "plan clear" && logged("Cleared the plan") == 1
	})

	// ↑/↓ still walk the history, back to the draft.
	typed("dra")
	h.key(tcell.KeyUp, 0)
	h.waitFor("↑ to recall the last command", 3*time.Second, func() bool { return h.inputText() == "plan clear" })
	h.key(tcell.KeyUp, 0)
	h.waitFor("↑ again for the one before", 3*time.Second, func() bool { return h.inputText() == "plan build "+picked })
	h.key(tcell.KeyDown, 0)
	h.waitFor("↓ forward", 3*time.Second, func() bool { return h.inputText() == "plan clear" })
	h.key(tcell.KeyDown, 0)
	h.waitFor("↓ back to the draft", 3*time.Second, func() bool { return h.inputText() == "dra" })
	clear()

	// Garbage runs as typed: the unknown-command message, no completion.
	typed("zzqx")
	h.key(tcell.KeyEnter, 0)
	h.waitFor("garbage to run as typed", 3*time.Second, func() bool { return h.inputText() == "" && lastRun() == "zzqx" })

	// adv + Enter runs advance.
	if !h.grantNextAge() {
		t.Fatal("already at the final age")
	}
	from := h.eng.GetState().Age
	h.advanceAndCheckWith(0, 'e', func() {
		typed("adv")
		h.key(tcell.KeyEnter, 0)
	})
	if got := h.eng.GetState().Age; got == from {
		t.Fatalf("age still %s after adv + Enter", got)
	}
	if got := lastRun(); got != "advance" {
		t.Errorf("history records %q, want the completed advance", got)
	}

	// advance + Enter runs once: one age, one history entry.
	if !h.grantNextAge() {
		t.Fatal("already at the final age")
	}
	st := h.eng.GetState()
	from, want := st.Age, st.NextAge
	n := len(history())
	h.advanceAndCheckWith(1, 'e', func() {
		typed("advance")
		h.key(tcell.KeyEnter, 0)
	})
	if got := h.eng.GetState().Age; got != want {
		t.Fatalf("advance + Enter went %s → %s, want one age to %s", from, got, want)
	}
	if got := len(history()); got != n+1 {
		t.Errorf("history grew by %d, want 1", got-n)
	}
}
