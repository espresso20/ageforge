package ui

import (
	"image"

	"github.com/espresso20/ageforge/window/kit"
)

// key_bar.go makes the keys along a panel's foot into buttons: where they are
// in the grid and what a click on them does. A click on one presses its key
// as the keyboard would. The close button in a panel's top edge is one more of
// them: it presses Esc. How they look is button.go's.

// xkey is a key a click can press: the key, the cells that press it, and the
// cells the key and its word take.
type xkey struct {
	key kit.Key
	r   image.Rectangle
	// cap is the cells of the key's own cap; the keys of one cap share it.
	// Empty for a key that is only a hit (a card's action in a scrolled view).
	cap image.Rectangle
	// word is the cells of the word beside the cap, empty when there is none.
	word image.Rectangle
	kind keyKind
	// off marks a key that is drawn but unavailable; still, one that names no
	// key to press ("1-9"): drawn as a key, never pressed.
	off, still bool
}

// keyAt is the key the cell pt presses.
func (g *xgrid) keyAt(pt image.Point) (xkey, bool) {
	for _, k := range g.keys {
		if pt.In(k.r) && !k.still {
			return k, true
		}
	}
	return xkey{}, false
}

// keysOff marks the keys drawn since the first len(g.keys) was from as
// unavailable.
func (g *xgrid) keysOff(from int) {
	for i := from; i < len(g.keys); i++ {
		g.keys[i].off = true
	}
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

// keyButton draws a key and its word at (x, y) as a button, records what a
// click on it presses, and returns the next x. A cap that names one key is
// one button with its word. A cap that names several is a button a key, each
// with its own cells of the cap; the word is then nobody's. The cells hold the
// cap and the word as they read without pictures; paint restyles them.
func (g *xgrid) keyButton(x, y int, key, name string) int {
	capEnd := g.put(x, y, " "+key+" ", xKeycap)
	end := g.put(capEnd, y, " "+name+" ", xText)
	cap, word := image.Rect(x, y, capEnd, y+1), image.Rect(capEnd, y, end, y+1)
	parts := capParts(key)
	kind := keySteel
	for _, p := range parts {
		if !moveKey(p.key) {
			kind = keyAccent
		}
	}
	if len(parts) == 0 {
		g.keys = append(g.keys, xkey{r: cap, cap: cap, word: word, still: true})
		return end
	}
	if len(parts) == 1 {
		g.keys = append(g.keys, xkey{key: parts[0].key, r: image.Rect(x, y, end, y+1), cap: cap, word: word, kind: kind})
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
		g.keys = append(g.keys, xkey{key: p.key, r: image.Rect(x0, y, x1, y+1), cap: cap, word: word, kind: kind})
	}
	return end
}

// moveKey reports whether k moves a selection or a view: the arrows, the
// paging keys, Home and End.
func moveKey(k kit.Key) bool {
	switch k.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight, kit.KeyPgUp, kit.KeyPgDn, kit.KeyHome, kit.KeyEnd:
		return true
	}
	return false
}

// escClose is the key that closes a panel, for a foot that ends with it.
func escClose() [2]string { return [2]string{"Esc", "close"} }

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
	g.put(x, 0, " × ", xKeycap)
	// The whole slot takes the click; the key itself rests on its first cells.
	g.keys = append(g.keys, xkey{key: kit.Key{Code: kit.KeyEsc}, r: image.Rect(x, 0, x+closeButtonCols, 1), cap: image.Rect(x, 0, x+3, 1), kind: keyClose})
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
	// feel is what the keys typed and the buttons held look like: shared
	// with the app (wire), or the presser's own.
	feel *keyFeel
	// plates are the pictures of the keys drawn so far.
	plates plateSet
}

// wire connects the presser to the app: a press is a key sent to it, and the
// keys typed on it show on the buttons.
func (k *keyPresser) wire(a *appLoop) { k.press, k.feel = a.Key, &a.feel }

// sense is the presser's feel, made when it has none.
func (k *keyPresser) sense() *keyFeel {
	if k.feel == nil {
		k.feel = &keyFeel{}
	}
	return k.feel
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
	switch m.Kind {
	case kit.MousePress:
		// A press is a key's or it is not: one that only moved the light is
		// left for the panel to use. The key stays sunk until the button is up.
		if isLit && m.Button == kit.ButtonLeft {
			k.press(now.key)
			k.sense().hold(now.key)
		}
		return isLit && m.Button == kit.ButtonLeft
	case kit.MouseRelease, kit.MouseLeave:
		if k.feel != nil && k.feel.release() {
			return true
		}
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
