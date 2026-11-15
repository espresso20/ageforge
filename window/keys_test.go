package window

import (
	"testing"

	"github.com/espresso20/ageforge/window/kit"
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
		{"enter", ebiten.KeyEnter, mods{}, press{key: kit.Key{Code: kit.KeyEnter}, hold: true}, true},
		{"esc", ebiten.KeyEscape, mods{}, press{key: kit.Key{Code: kit.KeyEsc}, hold: true}, true},
		{"tab", ebiten.KeyTab, mods{}, press{key: kit.Key{Code: kit.KeyTab}, hold: true}, true},
		{"backtab", ebiten.KeyTab, mods{shift: true}, press{key: kit.Key{Code: kit.KeyBacktab}, hold: true}, true},
		{"backspace", ebiten.KeyBackspace, mods{}, press{key: kit.Key{Code: kit.KeyBackspace}, hold: true}, true},
		{"delete", ebiten.KeyDelete, mods{}, press{key: kit.Key{Code: kit.KeyDelete}, hold: true}, true},
		{"up", ebiten.KeyArrowUp, mods{}, press{key: kit.Key{Code: kit.KeyUp}, hold: true}, true},
		{"home", ebiten.KeyHome, mods{}, press{key: kit.Key{Code: kit.KeyHome}, hold: true}, true},
		{"pgdn", ebiten.KeyPageDown, mods{}, press{key: kit.Key{Code: kit.KeyPgDn}, hold: true}, true},
		{"f5", ebiten.KeyF5, mods{}, press{key: kit.Key{Code: kit.KeyF5}, hold: true}, true},
		{"f12", ebiten.KeyF12, mods{}, press{key: kit.Key{Code: kit.KeyF12}, hold: true}, true},
		{"option+up keeps alt for mac_keys", ebiten.KeyArrowUp, mods{alt: true}, press{key: kit.Key{Code: kit.KeyUp, Mod: kit.ModAlt}, hold: true}, true},
		{"shift+left", ebiten.KeyArrowLeft, mods{shift: true}, press{key: kit.Key{Code: kit.KeyLeft, Mod: kit.ModShift}, hold: true}, true},
		{"ctrl+c", ebiten.KeyC, mods{ctrl: true}, press{key: kit.Key{Code: kit.KeyChord, Rune: 'c', Mod: kit.ModCtrl}, hold: true}, true},
		{"alt+b", ebiten.KeyB, mods{alt: true}, press{key: kit.Key{Code: kit.KeyChord, Rune: 'b', Mod: kit.ModAlt}, hold: true}, true},
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

// Ctrl+C arrives as the chord the game leaves on, never as the letter.
func TestCtrlCIsAChord(t *testing.T) {
	p, _ := mapKey(ebiten.KeyC, mods{ctrl: true})
	if !p.key.Ctrl('c') || p.key.Code == kit.KeyRune {
		t.Fatalf("Ctrl+C is %+v, want the chord", p.key)
	}
}

func TestMapChars(t *testing.T) {
	got := mapChars([]rune{'a', 'Z', ' ', '\r', 0x7f, 0xF700, 'é'})
	want := []rune{'a', 'Z', ' ', 'é'}
	if len(got) != len(want) {
		t.Fatalf("got %d presses, want %d: %+v", len(got), len(want), got)
	}
	for i, r := range want {
		if got[i].key != (kit.Key{Code: kit.KeyRune, Rune: r}) {
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
