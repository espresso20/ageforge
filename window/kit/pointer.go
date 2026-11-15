package kit

import "image"

// A Pointer is the shape the mouse pointer takes over a cell.
type Pointer uint8

const (
	PointerArrow Pointer = iota // nothing to say: the usual arrow
	PointerHand                 // over something a click acts on
	PointerText                 // over text that can be edited
)

// A Pointing widget says which shape the pointer takes at a point inside it.
// A widget that does not say, or says the arrow, leaves it to what holds it.
type Pointing interface {
	PointerAt(pt image.Point) Pointer
}

// pointerAt is the shape at the end of a path of widgets: the innermost one
// that names a shape other than the arrow.
func pointerAt(path []Widget, pt image.Point) Pointer {
	for i := len(path) - 1; i >= 0; i-- {
		if p, ok := path[i].(Pointing); ok {
			if s := p.PointerAt(pt); s != PointerArrow {
				return s
			}
		}
	}
	return PointerArrow
}

// PointerAt: a button that can be pressed shows the hand.
func (b *Button) PointerAt(image.Point) Pointer {
	if b.Disabled {
		return PointerArrow
	}
	return PointerHand
}

// PointerAt: a row of the list shows the hand.
func (l *List) PointerAt(pt image.Point) Pointer {
	if l.rowAt(pt) < 0 {
		return PointerArrow
	}
	return PointerHand
}

// PointerAt: an input line shows the text bar.
func (in *Input) PointerAt(image.Point) Pointer { return PointerText }

// PointerAt: a region of the text that a click acts on shows the hand.
func (t *Text) PointerAt(pt image.Point) Pointer {
	if t.OnRegion == nil || t.regionAt(pt) == "" {
		return PointerArrow
	}
	return PointerHand
}
