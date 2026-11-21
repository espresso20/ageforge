package kit

import (
	"image"

	"github.com/espresso20/ageforge/theme"
)

// Text is tagged text in a box: one string, broken into lines.
type Text struct {
	Base
	// Wrap is what a line longer than the box does; Align where lines sit.
	Wrap  Wrap
	Align Align
	// Style is what the text starts in. With a ground set, the box is
	// filled with it first.
	Style Style
	// OnRegion is called with a region's name when the player clicks text
	// inside one (see the ["name"] tag), and with 2 for the second click of
	// a double click.
	OnRegion func(name string, clicks int)
	// OnHover, when set, is called with the region the pointer came to rest
	// on, "" when it left the regions.
	OnHover func(name string)

	hover   string // the region under the pointer, where a click acts on regions
	text    string
	lines   []Line
	longest int // the widest line
	// what lines was laid out for
	forText  string
	forWidth int
	forWrap  Wrap
	laid     bool
}

// NewText is text that wraps between words.
func NewText(s string) *Text { return &Text{text: s, Wrap: WrapWord} }

// Label is one line of text that does not wrap.
func Label(s string) *Text { return &Text{text: s} }

// SetText changes what it says.
func (t *Text) SetText(s string) { t.text = s }

// GetText is what it says, tags and all.
func (t *Text) GetText() string { return t.text }

// Lines is the text as it lies in a box w cells wide.
func (t *Text) Lines(w int) []Line {
	if !t.laid || t.forText != t.text || t.forWidth != w || t.forWrap != t.Wrap {
		t.lines = layout(t.text, t.Style, w, t.Wrap)
		t.longest = 0
		for i := range t.lines {
			t.longest = max(t.longest, t.lines[i].w)
		}
		t.forText, t.forWidth, t.forWrap, t.laid = t.text, w, t.Wrap, true
	}
	return t.lines
}

func (t *Text) Layout(r image.Rectangle) { t.Place(r) }

// Size is the widest line by the number of lines, wrapped to most.X.
func (t *Text) Size(most image.Point) image.Point {
	ls := t.Lines(most.X)
	return image.Pt(min(t.longest, most.X), len(ls))
}

// place is where line l starts in the box, and how many of its cells are
// cut off on the left. A line too long for the box shows its middle when
// centered; right-aligned lines end where the longest line ends.
func (t *Text) place(l *Line) (x, skip int) {
	w := t.rect.Dx()
	switch t.Align {
	case AlignCenter:
		if l.w > w {
			return t.rect.Min.X, (l.w - w) / 2
		}
		return t.rect.Min.X + (w-l.w)/2, 0
	case AlignRight:
		return t.rect.Min.X + max(w, t.longest) - l.w, 0
	}
	return t.rect.Min.X, 0
}

func (t *Text) Paint(p *Painter) {
	r := t.rect
	p = p.Clip(r)
	p.Fill(r, t.Style.Bg)
	ls := t.Lines(r.Dx())
	// Only the lines the painter can touch: a Scroll gives a tall text a
	// short window.
	c := p.Bounds()
	for y := max(c.Min.Y, r.Min.Y); y < c.Max.Y && y-r.Min.Y < len(ls); y++ {
		l := &ls[y-r.Min.Y]
		x, skip := t.place(l)
		p.line(x, y, l, skip, r.Max.X-x)
		if t.hover != "" {
			t.light(p, l, x, skip, y)
		}
	}
}

// light marks the region under the pointer on one line: its cells that have
// no ground of their own take the chip's, as a list's row under the pointer
// does. A line that holds no other region is marked across the whole box, so
// a region of whole lines reads as one block.
func (t *Text) light(p *Painter, l *Line, x, skip, y int) {
	chip := theme.RGBA(theme.RoleChip)
	in, alone := false, true
	for i := range l.g {
		switch l.g[i].region {
		case t.hover:
			in = true
		case "":
		default:
			alone = false
		}
	}
	if !in {
		return
	}
	if alone {
		for cx := t.rect.Min.X; cx < t.rect.Max.X; cx++ {
			if c := p.cell(cx, y); c != nil {
				c.Bg = chip
			}
		}
	}
	// The cells of the text: the region's take the chip (they may be all
	// that is marked), and one with a ground of its own gets it back.
	for i := range l.g {
		g := &l.g[i]
		if skip > 0 {
			skip -= g.w
			continue
		}
		for k := 0; k < g.w; k++ {
			c := p.cell(x+k, y)
			switch {
			case c == nil:
			case g.st.Bg.IsSet():
				ink(c, g.st)
			case g.region == t.hover:
				c.Bg = chip
			}
		}
		x += g.w
	}
}

// regionAt is the region of the text under a cell, "" for none.
func (t *Text) regionAt(pt image.Point) string {
	ls := t.Lines(t.rect.Dx())
	row := pt.Y - t.rect.Min.Y
	if row < 0 || row >= len(ls) {
		return ""
	}
	l := &ls[row]
	x, skip := t.place(l)
	for i := range l.g {
		g := &l.g[i]
		if skip > 0 {
			skip -= g.w
			continue
		}
		if pt.X >= x && pt.X < x+g.w {
			return g.region
		}
		x += g.w
	}
	return ""
}

// Mouse: a click on a region tells OnRegion, and the region under the
// pointer is marked while the pointer is on it.
func (t *Text) Mouse(m Mouse) bool {
	if t.OnRegion == nil {
		return false
	}
	switch m.Kind {
	case MouseLeave, MouseMove:
		was := t.hover
		if t.hover = ""; m.Kind == MouseMove {
			t.hover = t.regionAt(m.Pt())
		}
		if t.hover != was && t.OnHover != nil {
			t.OnHover(t.hover)
		}
		return t.hover != was
	case MousePress:
		if name := t.regionAt(m.Pt()); name != "" && m.Button == ButtonLeft {
			t.OnRegion(name, max(m.Clicks, 1))
			return true
		}
	}
	return false
}
