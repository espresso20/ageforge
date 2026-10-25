package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestMacKeysTranslation: Option+arrows are PgUp, PgDn, Home and End for
// every panel, and nothing else is touched.
func TestMacKeysTranslation(t *testing.T) {
	ev := func(k tcell.Key, r rune, m tcell.ModMask) *tcell.EventKey { return tcell.NewEventKey(k, r, m) }
	for _, c := range []struct {
		from tcell.Key
		to   tcell.Key
	}{{tcell.KeyUp, tcell.KeyPgUp}, {tcell.KeyDown, tcell.KeyPgDn}, {tcell.KeyLeft, tcell.KeyHome}, {tcell.KeyRight, tcell.KeyEnd}} {
		got := macKeys(ev(c.from, 0, tcell.ModAlt), nil)
		if got.Key() != c.to || got.Modifiers() != tcell.ModNone {
			t.Errorf("Alt+%s became %s with modifiers %v, want %s", tcell.KeyNames[c.from], tcell.KeyNames[got.Key()], got.Modifiers(), tcell.KeyNames[c.to])
		}
		for _, m := range []tcell.ModMask{tcell.ModNone, tcell.ModShift, tcell.ModCtrl, tcell.ModAlt | tcell.ModShift, tcell.ModAlt | tcell.ModCtrl} {
			if in := ev(c.from, 0, m); macKeys(in, nil) != in {
				t.Errorf("%s with modifiers %v was changed", tcell.KeyNames[c.from], m)
			}
		}
	}
	other := ev(tcell.KeyRune, 'x', tcell.ModAlt)
	if macKeys(other, nil) != other {
		t.Error("Alt+x was changed")
	}
	if macKeys(nil, nil) != nil {
		t.Error("a nil event came back as something")
	}

	// Apple Terminal's Option+Left and Option+Right are ESC b and ESC f, which
	// are the prompt's word jumps: Home and End only when there is no word.
	empty, typed := tview.NewInputField(), tview.NewInputField().SetText("build hut")
	for _, f := range []struct {
		r        rune
		to       tcell.Key
		focus    tview.Primitive
		wantMove bool
	}{
		{'b', tcell.KeyHome, nil, true}, {'f', tcell.KeyEnd, nil, true},
		{'b', tcell.KeyHome, empty, true}, {'f', tcell.KeyEnd, empty, true},
		{'b', tcell.KeyHome, tview.NewBox(), true},
		{'b', tcell.KeyHome, typed, false}, {'f', tcell.KeyEnd, typed, false},
	} {
		in := ev(tcell.KeyRune, f.r, tcell.ModAlt)
		got := macKeys(in, f.focus)
		if f.wantMove && got.Key() != f.to || !f.wantMove && got != in {
			t.Errorf("Alt+%c with focus %T (text %v): got %s", f.r, f.focus, f.focus == typed, tcell.KeyNames[got.Key()])
		}
	}
}

// TestOptionKeysReachThePanels runs the real app: Option+Up and Option+Down
// reach the badge case as PgUp and PgDn, Option+Left and Option+Right as
// Home and End, ESC b and ESC f too while the prompt is empty, and the map
// zooms on - and = when it is.
func TestOptionKeysReachThePanels(t *testing.T) {
	h := newReproHarness(t)
	press := func(k tcell.Key, r rune, m tcell.ModMask) {
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
	alt := tcell.ModAlt

	press(tcell.KeyHome, 0, tcell.ModNone)
	if sel() != first {
		t.Fatalf("Home selects %q, want %q", sel(), first)
	}
	press(tcell.KeyDown, 0, alt)
	pgdn := sel()
	if pgdn == first {
		t.Error("Alt+Down did not move a page")
	}
	press(tcell.KeyHome, 0, tcell.ModNone)
	press(tcell.KeyPgDn, 0, tcell.ModNone)
	if sel() != pgdn {
		t.Errorf("Alt+Down selected %q, PgDn %q", pgdn, sel())
	}
	press(tcell.KeyUp, 0, alt)
	pgup := sel()
	press(tcell.KeyHome, 0, tcell.ModNone)
	press(tcell.KeyPgDn, 0, tcell.ModNone)
	press(tcell.KeyPgUp, 0, tcell.ModNone)
	if sel() != pgup {
		t.Errorf("Alt+Up selected %q, PgUp %q", pgup, sel())
	}

	press(tcell.KeyRight, 0, alt)
	if sel() != last {
		t.Errorf("Alt+Right selected %q, want the last badge %q", sel(), last)
	}
	press(tcell.KeyLeft, 0, alt)
	if sel() != first {
		t.Errorf("Alt+Left selected %q, want the first badge %q", sel(), first)
	}
	press(tcell.KeyRune, 'f', alt) // ESC f
	if sel() != last {
		t.Errorf("ESC f selected %q, want the last badge %q", sel(), last)
	}
	press(tcell.KeyRune, 'b', alt) // ESC b
	if sel() != first {
		t.Errorf("ESC b selected %q, want the first badge %q", sel(), first)
	}
	// With a word in the prompt ESC f is the prompt's own (a word jump), so
	// the badge case does not move.
	press(tcell.KeyRune, 'x', tcell.ModNone)
	press(tcell.KeyRune, 'f', alt)
	if sel() != first || h.inputText() != "x" {
		t.Errorf("ESC f with text in the prompt: selection %q, prompt %q", sel(), h.inputText())
	}
	press(tcell.KeyBackspace2, 0, tcell.ModNone)

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
		k    tcell.Key
		r    rune
		m    tcell.ModMask
		want string
	}{
		{tcell.KeyUp, 0, alt, "REGION"}, {tcell.KeyDown, 0, alt, "SETTLEMENT"},
		{tcell.KeyRune, '-', tcell.ModNone, "REGION"}, {tcell.KeyRune, '=', tcell.ModNone, "SETTLEMENT"},
		{tcell.KeyRune, '+', tcell.ModNone, "DISTRICT"},
	} {
		press(s.k, s.r, s.m)
		if !shows(s.want) {
			t.Fatalf("after %s %c (modifiers %v) the map does not show %s:\n%s", tcell.KeyNames[s.k], s.r, s.m, s.want, h.screenText())
		}
	}
	press(tcell.KeyRune, 'x', tcell.ModNone)
	press(tcell.KeyRune, '-', tcell.ModNone)
	if got := h.inputText(); got != "x-" || !shows("DISTRICT") {
		t.Errorf("with text in the prompt - went to %q, not the prompt, or zoomed the map", got)
	}
}
