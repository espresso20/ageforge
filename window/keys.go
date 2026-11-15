package window

import (
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/espresso20/ageforge/window/kit"
)

// keys.go turns what Ebitengine reports into the UI kit's keys. The table is
// plain data so a test can read it.

// mods are the modifier keys held at the moment.
type mods struct{ ctrl, alt, shift, meta bool }

// kit is the modifiers as the UI kit names them.
func (m mods) kit() kit.Mod {
	var out kit.Mod
	if m.ctrl {
		out |= kit.ModCtrl
	}
	if m.alt {
		out |= kit.ModAlt
	}
	if m.shift {
		out |= kit.ModShift
	}
	return out
}

// press is one key event for the game.
type press struct {
	key  kit.Key
	hold bool // repeats while the key is held
}

// specialKeys are the keys that are not text.
var specialKeys = map[ebiten.Key]kit.KeyCode{
	ebiten.KeyEnter:       kit.KeyEnter,
	ebiten.KeyNumpadEnter: kit.KeyEnter,
	ebiten.KeyEscape:      kit.KeyEsc,
	ebiten.KeyTab:         kit.KeyTab,
	ebiten.KeyBackspace:   kit.KeyBackspace,
	ebiten.KeyDelete:      kit.KeyDelete,
	ebiten.KeyInsert:      kit.KeyInsert,
	ebiten.KeyArrowUp:     kit.KeyUp,
	ebiten.KeyArrowDown:   kit.KeyDown,
	ebiten.KeyArrowLeft:   kit.KeyLeft,
	ebiten.KeyArrowRight:  kit.KeyRight,
	ebiten.KeyHome:        kit.KeyHome,
	ebiten.KeyEnd:         kit.KeyEnd,
	ebiten.KeyPageUp:      kit.KeyPgUp,
	ebiten.KeyPageDown:    kit.KeyPgDn,
	ebiten.KeyF1:          kit.KeyF1,
	ebiten.KeyF2:          kit.KeyF2,
	ebiten.KeyF3:          kit.KeyF3,
	ebiten.KeyF4:          kit.KeyF4,
	ebiten.KeyF5:          kit.KeyF5,
	ebiten.KeyF6:          kit.KeyF6,
	ebiten.KeyF7:          kit.KeyF7,
	ebiten.KeyF8:          kit.KeyF8,
	ebiten.KeyF9:          kit.KeyF9,
	ebiten.KeyF10:         kit.KeyF10,
	ebiten.KeyF11:         kit.KeyF11,
	ebiten.KeyF12:         kit.KeyF12,
}

// pressedKeys lists the keys the loop asks about each frame.
var pressedKeys = func() []ebiten.Key {
	var ks []ebiten.Key
	for k := range specialKeys {
		ks = append(ks, k)
	}
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		ks = append(ks, k)
	}
	return ks
}()

// letter is the lower-case letter of an ebiten letter key.
func letter(k ebiten.Key) (rune, bool) {
	if k >= ebiten.KeyA && k <= ebiten.KeyZ {
		return 'a' + rune(k-ebiten.KeyA), true
	}
	return 0, false
}

// mapKey is the event for one key going down (or repeating) with m held.
// Text is not handled here (see mapChars), only the keys that are not text and
// the chords: Ctrl+letter and Alt+letter arrive as the key, because the
// characters a chord types differ by layout and platform (Option+b types a
// symbol on a Mac).
func mapKey(k ebiten.Key, m mods) (press, bool) {
	if m.meta {
		return press{}, false // Cmd and the Windows key are the window's, never the game's
	}
	if code, ok := specialKeys[k]; ok {
		if code == kit.KeyTab && m.shift {
			return press{key: kit.Key{Code: kit.KeyBacktab}, hold: true}, true
		}
		return press{key: kit.Key{Code: code, Mod: m.kit()}, hold: true}, true
	}
	if l, ok := letter(k); ok && m.ctrl != m.alt { // both at once is AltGr, which types text
		if m.shift {
			l = unicode.ToUpper(l)
		}
		return press{key: kit.Key{Code: kit.KeyChord, Rune: l, Mod: m.kit()}, hold: true}, true
	}
	return press{}, false
}

// mapChars is the text typed this frame. Control characters and the private
// use codes some platforms give the cursor keys are dropped.
func mapChars(chars []rune) []press {
	var out []press
	for _, r := range chars {
		if r < ' ' || r == 0x7f || !unicode.IsPrint(r) || (r >= 0xE000 && r <= 0xF8FF) {
			continue
		}
		out = append(out, press{key: kit.Key{Code: kit.KeyRune, Rune: r}})
	}
	return out
}

// takesText reports whether typed characters should reach the game: not while
// a chord is held, since the chord is sent as a key. Ctrl and Alt together are
// AltGr on Windows and Linux, which types text.
func takesText(m mods) bool { return !m.meta && m.ctrl == m.alt }
