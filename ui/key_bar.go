package ui

import (
	"image"

	"github.com/espresso20/ageforge/window/kit"
)

// key_bar.go makes the keys along a panel's foot into buttons. Each is drawn
// as a keycap with its name beside it, a click on it presses its key as the
// keyboard would, and it lights up under the pointer. The close button in a
// panel's top edge is one more of them: it presses Esc.

// xkey is a key a click can press: the key and the cells that press it.
type xkey struct {
	key kit.Key
	r   image.Rectangle
}

// keyAt is the key the cell pt presses.
func (g *xgrid) keyAt(pt image.Point) (xkey, bool) {
	for _, k := range g.keys {
		if pt.In(k.r) {
			return k, true
		}
	}
	return xkey{}, false
}

// capPart is one key named on a keycap, and the runes of the cap's label
// that are its own.
type capPart struct {
	key    kit.Key
	off, w int
}

// capParts splits a keycap's label into the keys it names: "Esc" is one key,
// "↑↓" two, "PgUp PgDn" two. A part that names no key ("1-9") is left out,
// and so is a note in brackets after the keys.
func capParts(label string) []capPart {
	var parts []capPart
	rs := []rune(label)
	start := 0
	word := func(end int) {
		if k, ok := capKey(string(rs[start:end])); ok {
			parts = append(parts, capPart{k, start, end - start})
		}
		start = end + 1
	}
	for i, r := range rs {
		if r == '(' {
			word(i)
			return parts
		}
		if r == ' ' || r == '/' {
			word(i)
		} else if k, ok := arrowKey(r); ok {
			word(i)
			parts = append(parts, capPart{k, i, 1})
		}
	}
	word(len(rs))
	return parts
}

// arrowKey is the key an arrow glyph stands for.
func arrowKey(r rune) (kit.Key, bool) {
	switch r {
	case '↑':
		return kit.Key{Code: kit.KeyUp}, true
	case '↓':
		return kit.Key{Code: kit.KeyDown}, true
	case '←':
		return kit.Key{Code: kit.KeyLeft}, true
	case '→':
		return kit.Key{Code: kit.KeyRight}, true
	}
	return kit.Key{}, false
}

// keyButton draws a key and its name at (x, y) as a button, records what a
// click on it presses, and returns the next x. A cap that names one key is
// one button with its name. A cap that names several is a button a key, each
// with its own cells of the cap; the name is then nobody's.
func (g *xgrid) keyButton(x, y int, key, name string) int {
	capEnd := g.put(x, y, " "+key+" ", xKeycap)
	end := g.put(capEnd, y, " "+name+" ", xKeyName)
	parts := capParts(key)
	if len(parts) == 1 {
		g.keys = append(g.keys, xkey{parts[0].key, image.Rect(x, y, end, y+1)})
		return end
	}
	for i, p := range parts {
		x0, x1 := x+1+p.off, x+1+p.off+p.w
		if i == 0 {
			x0 = x
		}
		if i == len(parts)-1 {
			x1 = capEnd
		}
		g.keys = append(g.keys, xkey{p.key, image.Rect(x0, y, x1, y+1)})
	}
	return end
}

// keyButtonW is the cells keyButton takes.
func keyButtonW(key, name string) int { return len([]rune(key)) + len([]rune(name)) + 4 }

const (
	// closeButtonMinW is the narrowest box that gets a close button.
	closeButtonMinW = 24
	// closeButtonCols is the columns at the right of a box's top edge that
	// the close button takes, the corner and the edge beside it included:
	// anything else written in that edge ends before them.
	closeButtonCols = 5
)

// closeButton draws the close button in the top edge of a box w wide, at its
// right, and records that a click on it presses Esc.
func (g *xgrid) closeButton(w int) {
	if w < closeButtonMinW {
		return
	}
	x := w - closeButtonCols
	g.put(x, 0, " × ", xKeyName)
	g.keys = append(g.keys, xkey{kit.Key{Code: kit.KeyEsc}, image.Rect(x, 0, x+3, 1)})
}

// keyPresser is what a panel keeps to make the keys it draws pressable: what
// presses a key, and where the pointer is.
type keyPresser struct {
	// press presses a key as the keyboard would: the key goes to whatever
	// has the keyboard. Nil in a panel drawn alone (a picture of the screen),
	// whose keys are then not buttons.
	press func(kit.Key)
	// over is the cell the pointer is over, in the panel's own cells, while
	// in says it is over the panel.
	over image.Point
	in   bool
}

// lit is the key under the pointer.
func (k *keyPresser) lit(g *xgrid) (xkey, bool) {
	if g == nil || k.press == nil || !k.in {
		return xkey{}, false
	}
	return g.keyAt(k.over)
}

// mouse takes m for the keys g drew, with the panel's top left at org. A left
// press on a key presses it. It reports whether the event was used: a key was
// pressed, or the key that is lit changed.
func (k *keyPresser) mouse(g *xgrid, org image.Point, m kit.Mouse) bool {
	was, wasLit := k.lit(g)
	if k.in = m.Kind != kit.MouseLeave; k.in {
		k.over = m.Pt().Sub(org)
	}
	now, isLit := k.lit(g)
	if isLit && m.Kind == kit.MousePress && m.Button == kit.ButtonLeft {
		k.press(now.key)
		return true
	}
	return isLit != wasLit || now.r != was.r
}

// pointer is the hand over a key a click would press.
func (k *keyPresser) pointer(g *xgrid, org, at image.Point) kit.Pointer {
	if g == nil || k.press == nil {
		return kit.PointerArrow
	}
	_, ok := g.keyAt(at.Sub(org))
	return handIf(ok)
}

// paintLit repaints the key under the pointer in its lit colours, over the
// grid already painted at org.
func (k *keyPresser) paintLit(p *kit.Painter, g *xgrid, org image.Point, pal *expPalette) {
	hit, ok := k.lit(g)
	if !ok {
		return
	}
	for y := hit.r.Min.Y; y < hit.r.Max.Y; y++ {
		for x := hit.r.Min.X; x < hit.r.Max.X; x++ {
			c := g.c[y*g.w+x]
			switch c.k {
			case xKeycap:
				c.k = xKeycapLit
			case xKeyName:
				c.k = xKeyNameLit
			}
			p.Put(org.X+x, org.Y+y, c.r, pal.style(c))
		}
	}
}
