package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/pkg/cabinet"
	"github.com/espresso20/ageforge/window/kit"
)

// TestMacKeysTranslation: Option+arrows are PgUp, PgDn, Home and End for
// every panel, and nothing else is touched.
func TestMacKeysTranslation(t *testing.T) {
	ev := func(code kit.KeyCode, r rune, m kit.Mod) *kit.Key { return &kit.Key{Code: code, Rune: r, Mod: m} }
	for _, c := range []struct{ from, to kit.KeyCode }{
		{kit.KeyUp, kit.KeyPgUp}, {kit.KeyDown, kit.KeyPgDn}, {kit.KeyLeft, kit.KeyHome}, {kit.KeyRight, kit.KeyEnd},
	} {
		got := macKeys(ev(c.from, 0, kit.ModAlt))
		if got.Code != c.to || got.Mod != 0 {
			t.Errorf("Alt with key %v became %v with modifiers %v, want %v", c.from, got.Code, got.Mod, c.to)
		}
		for _, m := range []kit.Mod{0, kit.ModShift, kit.ModCtrl, kit.ModAlt | kit.ModShift, kit.ModAlt | kit.ModCtrl} {
			if in := ev(c.from, 0, m); macKeys(in) != in {
				t.Errorf("key %v with modifiers %v was changed", c.from, m)
			}
		}
	}
	other := ev(kit.KeyChord, 'x', kit.ModAlt)
	if macKeys(other) != other {
		t.Error("Alt+x was changed")
	}
	if macKeys(nil) != nil {
		t.Error("a nil event came back as something")
	}

}

// TestOptionKeysReachThePanels runs the real app: Option+Up and Option+Down
// reach the panels as PgUp and PgDn, Option+Left and Option+Right as
// Home and End (the cabinet takes the last two), and the map zooms on - and = when it is.
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
	h.waitFor("the cabinet", 5*time.Second, func() bool {
		var ok bool
		h.onUI(func() { ok = d.overlayMgr.ActiveName() == "badges" })
		return ok
	})
	sel := func() int {
		var i int
		h.onUI(func() { i = d.cabinet.sel })
		return i
	}
	alt := kit.ModAlt
	press(kit.KeyRight, 0, alt)
	if sel() != cabinet.Slots-1 {
		t.Errorf("Alt+Right selected drawer %d, want the last", sel())
	}
	press(kit.KeyLeft, 0, alt)
	if sel() != 0 {
		t.Errorf("Alt+Left selected drawer %d, want the first", sel())
	}

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
