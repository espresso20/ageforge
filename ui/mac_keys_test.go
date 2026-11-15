package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/window/kit"
)

// TestMacKeysTranslation: Option+arrows are PgUp, PgDn, Home and End for
// every panel, and nothing else is touched.
func TestMacKeysTranslation(t *testing.T) {
	ev := func(code kit.KeyCode, r rune, m kit.Mod) *kit.Key { return &kit.Key{Code: code, Rune: r, Mod: m} }
	for _, c := range []struct{ from, to kit.KeyCode }{
		{kit.KeyUp, kit.KeyPgUp}, {kit.KeyDown, kit.KeyPgDn}, {kit.KeyLeft, kit.KeyHome}, {kit.KeyRight, kit.KeyEnd},
	} {
		got := macKeys(ev(c.from, 0, kit.ModAlt), nil)
		if got.Code != c.to || got.Mod != 0 {
			t.Errorf("Alt with key %v became %v with modifiers %v, want %v", c.from, got.Code, got.Mod, c.to)
		}
		for _, m := range []kit.Mod{0, kit.ModShift, kit.ModCtrl, kit.ModAlt | kit.ModShift, kit.ModAlt | kit.ModCtrl} {
			if in := ev(c.from, 0, m); macKeys(in, nil) != in {
				t.Errorf("key %v with modifiers %v was changed", c.from, m)
			}
		}
	}
	other := ev(kit.KeyChord, 'x', kit.ModAlt)
	if macKeys(other, nil) != other {
		t.Error("Alt+x was changed")
	}
	if macKeys(nil, nil) != nil {
		t.Error("a nil event came back as something")
	}

	// Apple Terminal's Option+Left and Option+Right are ESC b and ESC f, which
	// are the prompt's word jumps: Home and End only when there is no word.
	empty, typed := newNativeInput(), newNativeInput().SetText("build hut")
	for _, f := range []struct {
		r        rune
		to       kit.KeyCode
		focus    Primitive
		wantMove bool
	}{
		{'b', kit.KeyHome, nil, true}, {'f', kit.KeyEnd, nil, true},
		{'b', kit.KeyHome, empty, true}, {'f', kit.KeyEnd, empty, true},
		{'b', kit.KeyHome, newNativeBox(), true},
		{'b', kit.KeyHome, typed, false}, {'f', kit.KeyEnd, typed, false},
	} {
		in := ev(kit.KeyChord, f.r, kit.ModAlt)
		got := macKeys(in, f.focus)
		if f.wantMove && got.Code != f.to || !f.wantMove && got != in {
			t.Errorf("Alt+%c with focus %T (text %v): got key %v", f.r, f.focus, f.focus == Primitive(typed), got.Code)
		}
	}
}

// TestOptionKeysReachThePanels runs the real app: Option+Up and Option+Down
// reach the badge case as PgUp and PgDn, Option+Left and Option+Right as
// Home and End, ESC b and ESC f too while the prompt is empty, and the map
// zooms on - and = when it is.
func TestOptionKeysReachThePanels(t *testing.T) {
	h := newReproHarness(t)
	press := func(k kit.KeyCode, r rune, m kit.Mod) {
		t.Helper()
		done := make(chan struct{})
		go func() { h.sim.InjectKey(k, r, m); close(done) }()
		select {
		case <-done:
		case <-time.After(4 * time.Second):
			t.Fatal("InjectKey blocked")
		}
		time.Sleep(15 * time.Millisecond)
		h.onUI(func() {}) // the key has been handled
	}
	d := h.a.dashboard

	h.submit("badges")
	h.waitFor("the badge case", 5*time.Second, func() bool {
		var ok bool
		h.onUI(func() {
			ok = d.overlayMgr.ActiveName() == "badges" && d.badgePanel.model != nil && len(d.badgePanel.model.items) > 1
		})
		return ok
	})
	sel := func() string {
		var s string
		h.onUI(func() { s = d.badgePanel.view.sel })
		return s
	}
	var first, last string
	h.onUI(func() {
		items := d.badgePanel.model.items
		first, last = items[0].id, items[len(items)-1].id
	})
	if first == last {
		t.Fatal("the badge case holds one badge")
	}
	alt := kit.ModAlt

	press(kit.KeyHome, 0, 0)
	if sel() != first {
		t.Fatalf("Home selects %q, want %q", sel(), first)
	}
	press(kit.KeyDown, 0, alt)
	pgdn := sel()
	if pgdn == first {
		t.Error("Alt+Down did not move a page")
	}
	press(kit.KeyHome, 0, 0)
	press(kit.KeyPgDn, 0, 0)
	if sel() != pgdn {
		t.Errorf("Alt+Down selected %q, PgDn %q", pgdn, sel())
	}
	press(kit.KeyUp, 0, alt)
	pgup := sel()
	press(kit.KeyHome, 0, 0)
	press(kit.KeyPgDn, 0, 0)
	press(kit.KeyPgUp, 0, 0)
	if sel() != pgup {
		t.Errorf("Alt+Up selected %q, PgUp %q", pgup, sel())
	}

	press(kit.KeyRight, 0, alt)
	if sel() != last {
		t.Errorf("Alt+Right selected %q, want the last badge %q", sel(), last)
	}
	press(kit.KeyLeft, 0, alt)
	if sel() != first {
		t.Errorf("Alt+Left selected %q, want the first badge %q", sel(), first)
	}
	press(kit.KeyChord, 'f', alt) // ESC f
	if sel() != last {
		t.Errorf("ESC f selected %q, want the last badge %q", sel(), last)
	}
	press(kit.KeyChord, 'b', alt) // ESC b
	if sel() != first {
		t.Errorf("ESC b selected %q, want the first badge %q", sel(), first)
	}
	// With a word in the prompt ESC f is the prompt's own (a word jump), so
	// the badge case does not move.
	press(kit.KeyRune, 'x', 0)
	press(kit.KeyChord, 'f', alt)
	if sel() != first || h.inputText() != "x" {
		t.Errorf("ESC f with text in the prompt: selection %q, prompt %q", sel(), h.inputText())
	}
	press(kit.KeyBackspace, 0, 0)

	// The map: Option+arrows and - = zoom it, the prompt having the rest.
	h.submit("map")
	h.waitFor("the map", 5*time.Second, func() bool {
		var ok bool
		h.onUI(func() { ok = d.overlayMgr.ActiveName() == "map" })
		return ok
	})
	shows := func(word string) bool { return strings.Contains(h.screenText(), word) }
	if !shows("SETTLEMENT") {
		t.Fatalf("the map does not open at the settlement zoom:\n%s", h.screenText())
	}
	for _, s := range []struct {
		k    kit.KeyCode
		r    rune
		m    kit.Mod
		want string
	}{
		{kit.KeyUp, 0, alt, "REGION"}, {kit.KeyDown, 0, alt, "SETTLEMENT"},
		{kit.KeyRune, '-', 0, "REGION"}, {kit.KeyRune, '=', 0, "SETTLEMENT"},
		{kit.KeyRune, '+', 0, "SETTLEMENT"},
	} {
		press(s.k, s.r, s.m)
		if !shows(s.want) {
			t.Fatalf("after key %v %c (modifiers %v) the map does not show %s:\n%s", s.k, s.r, s.m, s.want, h.screenText())
		}
	}
	press(kit.KeyRune, 'x', 0)
	press(kit.KeyRune, '-', 0)
	if got := h.inputText(); got != "x-" || !shows("SETTLEMENT") {
		t.Errorf("with text in the prompt - went to %q, not the prompt, or zoomed the map", got)
	}
}
