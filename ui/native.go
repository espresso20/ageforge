package ui

import (
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeHost is what a screen or a drop-in widget embeds to be a
// Primitive: the rectangle and focus flag every primitive has (primBase),
// and a capture, which sees the primitive's keys before it does.
type nativeHost struct {
	*primBase
	// capture sees the keys on their way to the primitive, or to whatever
	// in it has the keyboard (SetInputCapture).
	capture func(*kit.Key) *kit.Key
}

func newNativeHost() *nativeHost { return &nativeHost{primBase: newPrimBase()} }

// SetInputCapture sets what sees the primitive's keys before it does. It
// returns the key to go on with, or nil to stop it there.
func (h *nativeHost) SetInputCapture(fn func(*kit.Key) *kit.Key) *nativeHost {
	h.capture = fn
	return h
}

// FilterKey is the capture, where the kit asks for it: on a key's way down
// to whatever has the focus.
func (h *nativeHost) FilterKey(k kit.Key) (kit.Key, bool) {
	if h.capture == nil {
		return k, true
	}
	out := h.capture(&k)
	if out == nil {
		return k, false
	}
	return *out, true
}

// newKey is a key press, for the code that makes one up.
func newKey(code kit.KeyCode, r rune, mod kit.Mod) *kit.Key {
	return &kit.Key{Code: code, Rune: r, Mod: mod}
}
