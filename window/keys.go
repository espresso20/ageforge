package window

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
)

// keys.go turns what Ebitengine reports into the key events the game already
// understands. The table is plain data so a test can read it.

// mods are the modifier keys held at the moment.
type mods struct{ ctrl, alt, shift, meta bool }

func (m mods) mask() tcell.ModMask {
	var out tcell.ModMask
	if m.ctrl {
		out |= tcell.ModCtrl
	}
	if m.alt {
		out |= tcell.ModAlt
	}
	if m.shift {
		out |= tcell.ModShift
	}
	return out
}

// press is one key event to inject into the simulation screen.
type press struct {
	key  tcell.Key
	r    rune
	mod  tcell.ModMask
	hold bool // repeats while the key is held
}

// specialKeys are the keys that are not text.
var specialKeys = map[ebiten.Key]tcell.Key{
	ebiten.KeyEnter:       tcell.KeyEnter,
	ebiten.KeyNumpadEnter: tcell.KeyEnter,
	ebiten.KeyEscape:      tcell.KeyEsc,
	ebiten.KeyTab:         tcell.KeyTab,
	ebiten.KeyBackspace:   tcell.KeyBackspace2,
	ebiten.KeyDelete:      tcell.KeyDelete,
	ebiten.KeyInsert:      tcell.KeyInsert,
	ebiten.KeyArrowUp:     tcell.KeyUp,
	ebiten.KeyArrowDown:   tcell.KeyDown,
	ebiten.KeyArrowLeft:   tcell.KeyLeft,
	ebiten.KeyArrowRight:  tcell.KeyRight,
	ebiten.KeyHome:        tcell.KeyHome,
	ebiten.KeyEnd:         tcell.KeyEnd,
	ebiten.KeyPageUp:      tcell.KeyPgUp,
	ebiten.KeyPageDown:    tcell.KeyPgDn,
	ebiten.KeyF1:          tcell.KeyF1,
	ebiten.KeyF2:          tcell.KeyF2,
	ebiten.KeyF3:          tcell.KeyF3,
	ebiten.KeyF4:          tcell.KeyF4,
	ebiten.KeyF5:          tcell.KeyF5,
	ebiten.KeyF6:          tcell.KeyF6,
	ebiten.KeyF7:          tcell.KeyF7,
	ebiten.KeyF8:          tcell.KeyF8,
	ebiten.KeyF9:          tcell.KeyF9,
	ebiten.KeyF10:         tcell.KeyF10,
	ebiten.KeyF11:         tcell.KeyF11,
	ebiten.KeyF12:         tcell.KeyF12,
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
	if tk, ok := specialKeys[k]; ok {
		if tk == tcell.KeyTab && m.shift {
			return press{key: tcell.KeyBacktab, hold: true}, true
		}
		mod := m.mask()
		return press{key: tk, mod: mod, hold: true}, true
	}
	if l, ok := letter(k); ok && m.ctrl != m.alt { // both at once is AltGr, which types text
		mod := m.mask()
		if m.shift {
			l = unicode.ToUpper(l)
		}
		return press{key: tcell.KeyRune, r: l, mod: mod, hold: true}, true
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
		out = append(out, press{key: tcell.KeyRune, r: r})
	}
	return out
}

// takesText reports whether typed characters should reach the game: not while
// a chord is held, since the chord is sent as a key. Ctrl and Alt together are
// AltGr on Windows and Linux, which types text.
func takesText(m mods) bool { return !m.meta && m.ctrl == m.alt }
