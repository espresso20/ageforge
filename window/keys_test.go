package window

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestMapKey(t *testing.T) {
	cases := []struct {
		name string
		k    ebiten.Key
		m    mods
		want press
		ok   bool
	}{
		{"enter", ebiten.KeyEnter, mods{}, press{key: tcell.KeyEnter, hold: true}, true},
		{"esc", ebiten.KeyEscape, mods{}, press{key: tcell.KeyEsc, hold: true}, true},
		{"tab", ebiten.KeyTab, mods{}, press{key: tcell.KeyTab, hold: true}, true},
		{"backtab", ebiten.KeyTab, mods{shift: true}, press{key: tcell.KeyBacktab, hold: true}, true},
		{"backspace", ebiten.KeyBackspace, mods{}, press{key: tcell.KeyBackspace2, hold: true}, true},
		{"delete", ebiten.KeyDelete, mods{}, press{key: tcell.KeyDelete, hold: true}, true},
		{"up", ebiten.KeyArrowUp, mods{}, press{key: tcell.KeyUp, hold: true}, true},
		{"home", ebiten.KeyHome, mods{}, press{key: tcell.KeyHome, hold: true}, true},
		{"pgdn", ebiten.KeyPageDown, mods{}, press{key: tcell.KeyPgDn, hold: true}, true},
		{"f5", ebiten.KeyF5, mods{}, press{key: tcell.KeyF5, hold: true}, true},
		{"f12", ebiten.KeyF12, mods{}, press{key: tcell.KeyF12, hold: true}, true},
		{"option+up keeps alt for mac_keys", ebiten.KeyArrowUp, mods{alt: true}, press{key: tcell.KeyUp, mod: tcell.ModAlt, hold: true}, true},
		{"shift+left", ebiten.KeyArrowLeft, mods{shift: true}, press{key: tcell.KeyLeft, mod: tcell.ModShift, hold: true}, true},
		{"ctrl+c", ebiten.KeyC, mods{ctrl: true}, press{key: tcell.KeyRune, r: 'c', mod: tcell.ModCtrl, hold: true}, true},
		{"alt+b", ebiten.KeyB, mods{alt: true}, press{key: tcell.KeyRune, r: 'b', mod: tcell.ModAlt, hold: true}, true},
		{"plain letter is text", ebiten.KeyA, mods{}, press{}, false},
		{"cmd is never the game's", ebiten.KeyC, mods{meta: true}, press{}, false},
		{"cmd+enter", ebiten.KeyEnter, mods{meta: true}, press{}, false},
	}
	for _, c := range cases {
		got, ok := mapKey(c.k, c.m)
		if ok != c.ok || got != c.want {
			t.Errorf("%s: got %+v %v, want %+v %v", c.name, got, ok, c.want, c.ok)
		}
	}
}

// Ctrl+C must arrive as tcell's KeyCtrlC once the simulation screen has built the event.
func TestCtrlCBecomesKeyCtrlC(t *testing.T) {
	p, _ := mapKey(ebiten.KeyC, mods{ctrl: true})
	ev := tcell.NewEventKey(p.key, p.r, p.mod)
	if ev.Key() != tcell.KeyCtrlC {
		t.Fatalf("got key %v, want KeyCtrlC", ev.Key())
	}
}

func TestMapChars(t *testing.T) {
	got := mapChars([]rune{'a', 'Z', ' ', '\r', 0x7f, 0xF700, 'é'})
	want := []rune{'a', 'Z', ' ', 'é'}
	if len(got) != len(want) {
		t.Fatalf("got %d presses, want %d: %+v", len(got), len(want), got)
	}
	for i, r := range want {
		if got[i].key != tcell.KeyRune || got[i].r != r {
			t.Errorf("press %d: %+v, want rune %q", i, got[i], r)
		}
	}
	if takesText(mods{ctrl: true}) || takesText(mods{alt: true}) || !takesText(mods{shift: true}) || !takesText(mods{ctrl: true, alt: true}) {
		t.Error("takesText: chords must not type")
	}
}

func TestOneLine(t *testing.T) {
	if got := oneLine("a\r\nb \n\n c\t"); got != "a b c" {
		t.Fatalf("oneLine = %q", got)
	}
}
