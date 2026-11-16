package ui

import (
	"image"

	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// key_foot.go is the keys along the foot of a menu page (Load game, Themes,
// Accounts): key prompts on a keySet, so a click presses the key as the
// keyboard would and the key sinks when it is typed. How they look is
// button.go's.

// footerKey is one action on a key bar: its key, its name, and a shorter
// name for a bar with little room ("" when the name is short already).
type footerKey struct {
	key          string
	label, short vibe.Entry
}

// name is the key's word on a bar: the short one when short is set and
// there is one.
func (k footerKey) name(short bool) string {
	if short && k.short.Plain() != "" {
		return k.short.String()
	}
	return k.label.String()
}

// footLine is the keys on one line of a bar, as key and word.
type footLine [][2]string

// footW is the cells a line takes with gap cells between its keys.
func (l footLine) footW(gap int) int {
	n := 0
	for i, k := range l {
		if i > 0 {
			n += gap
		}
		n += keyButtonW(k[0], k[1])
	}
	return n
}

// fitKeys lays a key bar out for w cells and rows lines. It writes the
// fullest form that fits on one line: every name in full, then the shorter
// names, then those set closer. A bar still too long for one line is set in
// full over two when there are two. The gap between keys comes with it.
func fitKeys(keys []footerKey, w, rows int) (lines []footLine, gap int) {
	line := func(ks []footerKey, short bool) footLine {
		out := make(footLine, len(ks))
		for i, k := range ks {
			out[i] = [2]string{k.key, k.name(short)}
		}
		return out
	}
	for _, form := range []struct {
		short bool
		gap   int
	}{{false, 1}, {true, 1}, {true, 0}} {
		if l := line(keys, form.short); l.footW(form.gap) <= w {
			return []footLine{l}, form.gap
		}
	}
	if rows >= 2 {
		for _, short := range []bool{false, true} {
			for cut := (len(keys) + 1) / 2; cut < len(keys); cut++ {
				a, b := line(keys[:cut], short), line(keys[cut:], short)
				if a.footW(1) <= w && b.footW(1) <= w {
					return []footLine{a, b}, 1
				}
			}
		}
	}
	return []footLine{line(keys, true)}, 0
}

// keyFoot is the foot of a page: its keys, centred, as buttons.
type keyFoot struct {
	*nativeHost
	kit.Base
	keys    []footerKey
	buttons keyPresser
	grid    *xgrid
}

// newKeyFoot is the foot with keys, whose buttons press their keys on app.
func newKeyFoot(app *appLoop, keys []footerKey) *keyFoot {
	f := &keyFoot{keys: keys}
	f.nativeHost = newNativeHost()
	if app != nil {
		f.buttons.wire(app)
	}
	return f
}

func (f *keyFoot) Layout(r image.Rectangle) { f.Place(r) }

// render draws the foot at w by h cells.
func (f *keyFoot) render(w, h int) *xgrid {
	g := newXGrid(w, h)
	lines, gap := fitKeys(f.keys, w, h)
	for y, l := range lines {
		if y >= h {
			break
		}
		x := max((w-l.footW(gap))/2, 0)
		for i, k := range l {
			if i > 0 {
				x += gap
			}
			x = g.keyButton(x, y, k[0], k[1])
		}
	}
	return g
}

func (f *keyFoot) Paint(p *kit.Painter) {
	r := f.Bounds()
	if r.Empty() {
		return
	}
	g := f.render(r.Dx(), r.Dy())
	pal := newExpPalette()
	paintExpeditionKit(p, g, r.Min, pal)
	f.buttons.paint(p, g, r.Min, pal)
	f.grid = g
}

// Mouse: a click on a key presses it.
func (f *keyFoot) Mouse(m kit.Mouse) bool {
	return f.buttons.mouse(f.grid, f.Bounds().Min, m)
}

// PointerAt: the hand over a key a click would press.
func (f *keyFoot) PointerAt(at image.Point) kit.Pointer {
	return f.buttons.pointer(f.grid, f.Bounds().Min, at)
}
