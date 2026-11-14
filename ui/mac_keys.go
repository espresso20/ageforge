package ui

import "github.com/espresso20/ageforge/window/kit"

// A MacBook has no Page Up, Page Down, Home or End, and Fn+arrows are often
// taken by the terminal itself (Apple Terminal scrolls its own window). The
// game uses all four in its panels, so one translation, ahead of every
// panel, lets Option do their work: Option+Up and Option+Down are PgUp and
// PgDn, Option+Left and Option+Right are Home and End. Nothing downstream
// changes; a panel sees the key it already knows.
//
// How the chord reaches the program depends on the terminal. A terminal that
// writes the xterm modifier form (ESC [ 1 ; 3 A) or puts ESC in front of the
// arrow (ESC ESC [ A) arrives as the arrow with ModAlt, which is the first
// case below. Apple Terminal's own default for Option+Left and Option+Right
// is ESC b and ESC f, which arrive as Alt+b and Alt+f: the second case. Those
// two are also the prompt's word-left and word-right, so they are taken only
// when the prompt has nothing in it to jump over.

// macKeys returns the key the program should see for k, given the widget
// that has the focus.
func macKeys(k *kit.Key, focus Primitive) *kit.Key {
	// Only Alt on its own: Shift, Ctrl and the rest keep their meanings.
	if k == nil || k.Mod != kit.ModAlt {
		return k
	}
	var to kit.KeyCode
	switch k.Code {
	case kit.KeyUp:
		to = kit.KeyPgUp
	case kit.KeyDown:
		to = kit.KeyPgDn
	case kit.KeyLeft:
		to = kit.KeyHome
	case kit.KeyRight:
		to = kit.KeyEnd
	case kit.KeyChord:
		switch k.Rune {
		case 'b':
			to = kit.KeyHome
		case 'f':
			to = kit.KeyEnd
		default:
			return k
		}
		if hasText(focus) {
			return k // word-left and word-right of what is being typed
		}
	default:
		return k
	}
	return &kit.Key{Code: to}
}

// hasText reports whether the focused widget is a text field with something
// typed in it.
func hasText(focus Primitive) bool {
	f, ok := focus.(interface{ GetText() string })
	return ok && f.GetText() != ""
}

// pageKeysHint is how a key bar spells Option+Up and Option+Down beside
// PgUp and PgDn. The plain glyph set has no ⌥, so it spells the chord out.
func pageKeysHint(plain bool) string {
	if plain {
		return "Alt+Up/Dn"
	}
	return "⌥↑↓"
}
