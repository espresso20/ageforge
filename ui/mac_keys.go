package ui

import "github.com/espresso20/ageforge/window/kit"

// A MacBook has no Page Up, Page Down, Home or End. The game uses all four
// in its panels, so one translation, ahead of every panel, lets Option do
// their work: Option+Up and Option+Down are PgUp and
// PgDn, Option+Left and Option+Right are Home and End. Nothing downstream
// changes; a panel sees the key it already knows.

// macKeys returns the key the program should see for k.
func macKeys(k *kit.Key) *kit.Key {
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
	default:
		return k
	}
	return &kit.Key{Code: to}
}

// pageKeysHint is how a key bar spells Option+Up and Option+Down beside
// PgUp and PgDn. The plain glyph set has no ⌥, so it spells the chord out.
func pageKeysHint(plain bool) string {
	if plain {
		return "Alt+Up/Dn"
	}
	return "⌥↑↓"
}
