package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// vibeOffAfter puts the process's VibeMode switch back where it was found.
func vibeOffAfter(t *testing.T) {
	t.Helper()
	was := vibe.On()
	t.Cleanup(func() { vibe.Set(was) })
	vibe.Set(false)
}

// helpTitleOf is the title on the frame of the Help panel.
func helpTitleOf(d *Dashboard) string { return d.overlayMgr.entries["help"].frame.title.String() }

// TestVibemodeCommand: vibemode shows the setting, turns it on and off, saves
// it to the account, and the words change live: the Buildings panel, the Help
// panel's title and the engine's ruleset follow, and back.
func TestVibemodeCommand(t *testing.T) {
	vibeOffAfter(t)
	d, _, eng := badgeDashboard(t)
	buildings := func() string { return buildingsProvider(eng.GetState(), 120) }

	if res := HandleCommand("vibemode", eng); res.Type != "info" || !strings.Contains(res.Message, "VibeMode: off") {
		t.Errorf("vibemode: %+v", res)
	}
	if !strings.Contains(buildings(), "Hut") || strings.Contains(buildings(), "Crash Pad") || !strings.Contains(helpTitleOf(d), "Help") {
		t.Fatal("the words are not plain before the switch")
	}

	d.runForTest("vibemode on")
	if !vibe.On() || !eng.Account().VibeOn() {
		t.Fatal("vibemode on did not turn it on, or did not save it to the account")
	}
	if res := HandleCommand("vibemode", eng); !strings.Contains(res.Message, "VibeMode: on") {
		t.Errorf("vibemode: %+v", res)
	}
	if got := buildings(); !strings.Contains(got, "Crash Pad") {
		t.Errorf("the Buildings panel did not change: %.300s", got)
	}
	if got := helpTitleOf(d); !strings.Contains(got, "Halp") || strings.Contains(got, "Help") {
		t.Errorf("the Help panel is titled %q with VibeMode on", got)
	}
	if got, _ := eng.Rules().Building("hut"); got.Name != "Crash Pad" {
		t.Errorf("the engine's ruleset calls the hut %q", got.Name)
	}

	d.runForTest("vibemode off")
	if vibe.On() || eng.Account().VibeOn() {
		t.Fatal("vibemode off did not turn it off")
	}
	if got := buildings(); !strings.Contains(got, "Hut") || strings.Contains(got, "Crash Pad") {
		t.Errorf("the Buildings panel did not go back: %.300s", got)
	}
	if got := helpTitleOf(d); !strings.Contains(got, "Help") || strings.Contains(got, "Halp") {
		t.Errorf("the Help panel is titled %q with VibeMode off again", got)
	}

	for _, bad := range []string{"vibemode sideways", "vibemode on off"} {
		if res := HandleCommand(bad, eng); res.Type != "error" || res.Vibe {
			t.Errorf("%s: %+v", bad, res)
		}
	}
	if vibe.On() {
		t.Error("a refused command turned VibeMode on")
	}
	// With no account the setting lasts for the session.
	bare := game.NewGameEngine()
	if res := HandleCommand("vibemode on", bare); !strings.Contains(res.Message, "this session") || !res.Vibe || !vibe.On() {
		t.Errorf("vibemode on with no account: %+v", res)
	}
}

// TestVibemodeSwitchedFromOutside: the switch can move with no command typed
// (an account with its own setting is switched to), and the Help title
// follows at the next refresh.
func TestVibemodeSwitchedFromOutside(t *testing.T) {
	vibeOffAfter(t)
	d, _, eng := badgeDashboard(t)
	for _, on := range []bool{true, false} {
		must(t, eng.SetVibeMode(on))
		d.refresh()
		want := map[bool]string{true: "Halp", false: "Help"}[on]
		if got := helpTitleOf(d); !strings.Contains(got, want) {
			t.Errorf("with VibeMode %v the Help panel is titled %q, want %s", on, got, want)
		}
	}
}
