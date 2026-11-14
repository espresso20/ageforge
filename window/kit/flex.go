package kit

import "image"

// Flex lays its items out in a row or a column. Each item is fixed (so many
// cells), weighted (a share of what the others leave), or natural (it is
// asked: see Sizer). A nil widget is a gap that takes room and shows nothing.
type Flex struct {
	Base
	// Down makes it a column; otherwise it is a row.
	Down bool
	// Gap is the cells left between items.
	Gap   int
	items []flexItem
}

type flexItem struct {
	w             Widget
	fixed, weight int
}

// Row is an empty row.
func Row() *Flex { return &Flex{} }

// Col is an empty column.
func Col() *Flex { return &Flex{Down: true} }

// Add puts w at the end. With fixed above 0 it gets that many cells; else
// with weight above 0 it shares what is left by weight; with both 0 it gets
// its natural size (or, if it has none, a weight of 1).
func (f *Flex) Add(w Widget, fixed, weight int) *Flex {
	f.items = append(f.items, flexItem{w: w, fixed: fixed, weight: weight})
	return f
}

// Clear empties it.
func (f *Flex) Clear() *Flex {
	f.items = f.items[:0]
	return f
}

func (f *Flex) Children() []Widget {
	out := make([]Widget, 0, len(f.items))
	for _, it := range f.items {
		if it.w != nil {
			out = append(out, it.w)
		}
	}
	return out
}

func (f *Flex) main(p image.Point) int {
	if f.Down {
		return p.Y
	}
	return p.X
}

// sizes is each item's cells along the flex, for a box of the given size.
func (f *Flex) sizes(box image.Point) []int {
	out := make([]int, len(f.items))
	left, weights, shown := f.main(box), 0, 0
	for i, it := range f.items {
		if it.w != nil && hidden(it.w) {
			continue
		}
		shown++
		s, sizer := it.w.(Sizer)
		switch {
		case it.fixed > 0:
			out[i] = it.fixed
		case it.weight > 0:
			weights += it.weight
		case sizer:
			out[i] = f.main(s.Size(box))
		default:
			weights++
		}
		left -= out[i]
	}
	left = max(left-f.Gap*max(shown-1, 0), 0)
	for i, it := range f.items {
		if out[i] > 0 || it.fixed > 0 || (it.w != nil && hidden(it.w)) {
			continue
		}
		wt := it.weight
		if _, sizer := it.w.(Sizer); wt == 0 && !sizer {
			wt = 1
		}
		if wt > 0 && weights > 0 {
			out[i] = left * wt / weights
			left -= out[i]
			weights -= wt
		}
	}
	return out
}

func (f *Flex) Layout(r image.Rectangle) {
	f.Place(r)
	at := r.Min
	for i, n := range f.sizes(r.Size()) {
		it := f.items[i]
		if it.w != nil && hidden(it.w) {
			continue
		}
		box := image.Rect(at.X, at.Y, at.X+n, r.Max.Y)
		if f.Down {
			box = image.Rect(at.X, at.Y, r.Max.X, at.Y+n)
			at.Y += n + f.Gap
		} else {
			at.X += n + f.Gap
		}
		if it.w != nil {
			it.w.Layout(box.Intersect(r))
		}
	}
}

func (f *Flex) Paint(p *Painter) {
	for _, it := range f.items {
		if !hidden(it.w) {
			it.w.Paint(p)
		}
	}
}

// Size is the room the items ask for: fixed and natural sizes along the
// flex, the largest natural size across it.
func (f *Flex) Size(most image.Point) image.Point {
	var along, across, shown int
	for _, it := range f.items {
		if it.w != nil && hidden(it.w) {
			continue
		}
		shown++
		var nat image.Point
		if s, ok := it.w.(Sizer); ok {
			nat = s.Size(most)
		}
		n, x := nat.X, nat.Y
		if f.Down {
			n, x = nat.Y, nat.X
		}
		if it.fixed > 0 {
			n = it.fixed
		}
		along += n
		across = max(across, x)
	}
	along += f.Gap * max(shown-1, 0)
	if f.Down {
		return image.Pt(min(across, most.X), min(along, most.Y))
	}
	return image.Pt(min(along, most.X), min(across, most.Y))
}

// Stack shows its items one over the other, each in the whole rectangle,
// the last on top.
type Stack struct {
	Base
	Items []Widget
}

func (s *Stack) Children() []Widget { return s.Items }

func (s *Stack) Layout(r image.Rectangle) {
	s.Place(r)
	for _, w := range s.Items {
		if !hidden(w) {
			w.Layout(r)
		}
	}
}

func (s *Stack) Paint(p *Painter) {
	for _, w := range s.Items {
		if !hidden(w) {
			w.Paint(p)
		}
	}
}

// Center holds one child in the middle of its rectangle, at W by H cells.
// A zero W or H asks the child (see Sizer), or else takes the whole way.
type Center struct {
	Base
	Child Widget
	W, H  int
}

func (c *Center) Children() []Widget { return []Widget{c.Child} }

// centered is the rectangle of a w by h child in the middle of r, with a
// zero w or h taken from the child or the whole of r.
func centered(r image.Rectangle, child Widget, w, h int) image.Rectangle {
	if w <= 0 || h <= 0 {
		nat := r.Size()
		if s, ok := child.(Sizer); ok {
			nat = s.Size(r.Size())
		}
		if w <= 0 {
			w = nat.X
		}
		if h <= 0 {
			h = nat.Y
		}
	}
	w, h = min(w, r.Dx()), min(h, r.Dy())
	x, y := r.Min.X+(r.Dx()-w)/2, r.Min.Y+(r.Dy()-h)/2
	return image.Rect(x, y, x+w, y+h)
}

func (c *Center) Layout(r image.Rectangle) {
	c.Place(r)
	c.Child.Layout(centered(r, c.Child, c.W, c.H))
}

func (c *Center) Paint(p *Painter) { c.Child.Paint(p) }
