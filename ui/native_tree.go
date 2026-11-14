package ui

import (
	"image"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// native_tree.go is the layout boxes the screens are built from, as kit
// containers: a flex, a group that places its children by its own rule, a
// filler, and a canvas for the few things still drawn cell by cell. A
// container lays each child out, paints it into the page's one buffer and
// sends it the mouse.

// widgetOf is p as a kit container holds it: itself, or nothing for nil (a
// gap). A nil of some pointer type is no primitive either.
func widgetOf(p Primitive) kit.Widget {
	if p == nil {
		return nil
	}
	return p
}

// place gives a child of a kit container its rectangle: the kit's, in the
// buffer, and tview's, on the screen, which is off further on. Code that
// asks a primitive for its rectangle still gets the right answer.
func place(w kit.Widget, r image.Rectangle, off image.Point) {
	placeAt(w, r.Min.X, r.Min.Y, r.Dx(), r.Dy(), off)
}

// placeAt is place for a size that may be below nothing, which tview hands
// on as it is when the room has run out.
func placeAt(w kit.Widget, x, y, wd, ht int, off image.Point) {
	if p, ok := w.(interface{ SetRect(x, y, w, h int) }); ok {
		p.SetRect(x+off.X, y+off.Y, wd, ht)
	}
	w.Layout(image.Rect(x, y, x+max(wd, 0), y+max(ht, 0)))
}

// screenOffset is how far a primitive's rectangle on the screen is from the
// rectangle r it was just laid out in: what place adds for its children.
func screenOffset(p interface{ GetRect() (int, int, int, int) }, r image.Rectangle) image.Point {
	x, y, _, _ := p.GetRect()
	return image.Pt(x-r.Min.X, y-r.Min.Y)
}

// walkWidgets calls fn for w and everything inside it, outermost first.
func walkWidgets(w kit.Widget, fn func(kit.Widget)) {
	if w == nil {
		return
	}
	fn(w)
	if p, ok := w.(kit.Parent); ok {
		for _, c := range p.Children() {
			walkWidgets(c, fn)
		}
	}
}

// coveredBy is the rectangles w paints: what it says, or all of its own.
func coveredBy(w kit.Widget) []image.Rectangle {
	if c, ok := w.(interface{ covered() []image.Rectangle }); ok {
		return c.covered()
	}
	return []image.Rectangle{w.Bounds()}
}

// A nativeBox stands where a bare tview.Box stood as a filler: a rectangle
// of one ground.
type nativeBox struct {
	*nativeHost
	kit.Base
	ground kit.Color
}

// newNativeBox is a filler in the page's ground, as tview.NewBox is.
func newNativeBox() *nativeBox {
	b := &nativeBox{ground: kit.Role(theme.RoleBackground)}
	b.nativeHost = newNativeHost()
	return b
}

func (b *nativeBox) SetBackgroundColor(c kit.Color) *nativeBox {
	b.ground = c
	return b
}

func (b *nativeBox) Layout(r image.Rectangle) { b.Place(r) }
func (b *nativeBox) Paint(p *kit.Painter)     { p.Fill(b.Bounds(), b.ground) }

// A nativeGroup is a layout box that places its children by a rule of its
// own (the dashboard's body, its left column, the map dock): a kit container
// whose arrange func says where each child goes in the room it is given.
// Like a flex it paints nothing itself.
type nativeGroup struct {
	*nativeHost
	kit.Base
	// arrange places the children for the rectangle x, y, w, h by calling
	// put for each one that shows.
	arrange func(x, y, w, h int)
	kids    []kit.Widget
	seen    map[Primitive]kit.Widget
	off     image.Point
}

func newNativeGroup(arrange func(x, y, w, h int)) *nativeGroup {
	g := &nativeGroup{arrange: arrange, seen: map[Primitive]kit.Widget{}}
	g.nativeHost = newNativeHost()
	return g
}

func (g *nativeGroup) Children() []kit.Widget { return g.kids }

func (g *nativeGroup) Layout(r image.Rectangle) {
	g.Place(r)
	if g.nativeHost == nil {
		return // newNativeHost lays the tree out before the host is set
	}
	g.kids, g.off = g.kids[:0], screenOffset(g, r)
	g.arrange(r.Min.X, r.Min.Y, r.Dx(), r.Dy())
}

// put shows p at x, y, w by h cells, in the coordinates arrange was given.
func (g *nativeGroup) put(p Primitive, x, y, w, h int) {
	child, ok := g.seen[p]
	if !ok {
		child = widgetOf(p)
		g.seen[p] = child
	}
	g.kids = append(g.kids, child)
	placeAt(child, x, y, w, h, g.off)
}

func (g *nativeGroup) Paint(p *kit.Painter) {
	for _, w := range g.kids {
		if !w.Bounds().Empty() {
			w.Paint(p)
		}
	}
}

// Covers: a group is nothing but its children.
func (g *nativeGroup) Covers(image.Point) bool { return false }

func (g *nativeGroup) covered() []image.Rectangle {
	var out []image.Rectangle
	for _, w := range g.kids {
		out = append(out, coveredBy(w)...)
	}
	return out
}

// A nativeFlex stands where a tview.Flex stood: the same calls, and the same
// sharing out of the room. Like tview's it paints nothing of its own but a
// border, so where it has a gap the page beneath shows.
type nativeFlex struct {
	*nativeHost
	kit.Base
	down  bool // items one under the other (flexRow)
	items []flexItem
	// frame is the border and title, when SetBorder asked for one, and
	// ground what they are drawn on (unset: the page's).
	frame    kit.Panel
	ground   kit.Color
	drawFunc func(x, y, w, h int)
}

type flexItem struct {
	prim        Primitive  // as it was added; nil for a gap
	w           kit.Widget // prim in the tree
	fixed, prop int
	focus       bool
}

// newNativeFlex is an empty flex, its items side by side, as tview makes one.
func newNativeFlex() *nativeFlex {
	f := &nativeFlex{}
	f.frame.TitleAlign = kit.AlignCenter
	f.nativeHost = newNativeHost()
	return f
}

// flexDir is which way a flex lays its items: side by side, or one under
// the other (rows).
type flexDir uint8

const (
	flexColumn flexDir = iota
	flexRow
)

func (f *nativeFlex) SetDirection(direction flexDir) *nativeFlex {
	f.down = direction == flexRow
	return f
}

// AddItem adds p at the end: fixed cells along the flex, or with fixed 0 a
// share of what is left in proportion to prop. The first item added with
// focus takes the keyboard when the flex is given it. A nil p is a gap.
func (f *nativeFlex) AddItem(p Primitive, fixed, prop int, focus bool) *nativeFlex {
	f.items = append(f.items, flexItem{prim: p, w: widgetOf(p), fixed: fixed, prop: prop, focus: focus})
	return f
}

func (f *nativeFlex) RemoveItem(p Primitive) *nativeFlex {
	for i := len(f.items) - 1; i >= 0; i-- {
		if f.items[i].prim == p {
			f.items = append(f.items[:i], f.items[i+1:]...)
		}
	}
	return f
}

func (f *nativeFlex) Clear() *nativeFlex {
	f.items = nil
	return f
}

func (f *nativeFlex) ResizeItem(p Primitive, fixed, prop int) *nativeFlex {
	for i := range f.items {
		if f.items[i].prim == p {
			f.items[i].fixed, f.items[i].prop = fixed, prop
		}
	}
	return f
}

func (f *nativeFlex) SetBorder(on bool) *nativeFlex {
	f.frame.Border = kit.BorderNone
	if on {
		f.frame.Border = kit.BorderLine
	}
	return f
}

func (f *nativeFlex) SetTitle(s string) *nativeFlex {
	f.frame.Title = s
	return f
}

func (f *nativeFlex) SetTitleColor(c kit.Color) *nativeFlex {
	f.frame.TitleStyle.Fg = c
	return f
}

func (f *nativeFlex) SetBorderColor(c kit.Color) *nativeFlex {
	f.frame.BorderStyle.Fg = c
	return f
}

// SetBackgroundColor sets the ground the border and the title are drawn on.
// Inside the border the flex still paints nothing: its items do.
func (f *nativeFlex) SetBackgroundColor(c kit.Color) *nativeFlex {
	f.ground = c
	return f
}

// SetDrawFunc sets what is called with the flex's rectangle before its items
// are placed, so a screen can size them for the room.
func (f *nativeFlex) SetDrawFunc(fn func(x, y, w, h int)) *nativeFlex {
	f.drawFunc = fn
	return f
}

func (f *nativeFlex) framed() bool { return f.frame.Border != kit.BorderNone }

// inner is the rectangle the items share: r, less the border.
func (f *nativeFlex) inner(r image.Rectangle) image.Rectangle {
	if f.framed() {
		return image.Rect(r.Min.X+1, r.Min.Y+1, r.Min.X+1+max(r.Dx()-2, 0), r.Min.Y+1+max(r.Dy()-2, 0))
	}
	return r
}

// GetInnerRect is the rectangle the items share, on the screen.
func (f *nativeFlex) GetInnerRect() (x, y, w, h int) {
	x, y, w, h = f.GetRect()
	in := f.inner(image.Rect(x, y, x+w, y+h))
	return in.Min.X, in.Min.Y, in.Dx(), in.Dy()
}

func (f *nativeFlex) Children() []kit.Widget {
	out := make([]kit.Widget, 0, len(f.items))
	for _, it := range f.items {
		if it.w != nil {
			out = append(out, it.w)
		}
	}
	return out
}

// Layout shares the room out as tview's Flex does: the fixed items first,
// then the rest by proportion, each share rounded down in turn.
func (f *nativeFlex) Layout(r image.Rectangle) {
	f.Place(r)
	if f.nativeHost == nil {
		return // newNativeHost lays the tree out before the host is set
	}
	off := screenOffset(f, r)
	if f.drawFunc != nil {
		x, y, w, h := f.GetRect()
		f.drawFunc(x, y, w, h)
	}
	in := f.inner(r)
	left, props := in.Dx(), 0
	if f.down {
		left = in.Dy()
	}
	for _, it := range f.items {
		if it.fixed > 0 {
			left -= it.fixed
		} else {
			props += it.prop
		}
	}
	x, y := in.Min.X, in.Min.Y
	for _, it := range f.items {
		size := it.fixed
		if size <= 0 {
			size = 0
			if props > 0 {
				size = left * it.prop / props
				left -= size
				props -= it.prop
			}
		}
		switch {
		case it.w == nil:
		case f.down:
			placeAt(it.w, x, y, in.Dx(), size, off)
		default:
			placeAt(it.w, x, y, size, in.Dy(), off)
		}
		if f.down {
			y += size
		} else {
			x += size
		}
	}
}

// Paint draws the border, doubled while the keyboard is inside as tview
// draws it, and the items the kit draws.
func (f *nativeFlex) Paint(p *kit.Painter) {
	if f.framed() {
		for _, r := range f.own() {
			p.Fill(r, f.ground)
		}
		fr := f.frame
		if f.HasFocus() {
			fr.Border = kit.BorderDouble
		}
		fr.Layout(f.Bounds())
		fr.Paint(p)
	}
	for _, it := range f.items {
		if it.w != nil && !it.w.Bounds().Empty() {
			it.w.Paint(p)
		}
	}
}

// own is the cells the flex paints itself: its border, when it has one.
func (f *nativeFlex) own() []image.Rectangle {
	r := f.Bounds()
	if !f.framed() || r.Empty() {
		return nil
	}
	return []image.Rectangle{
		image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+1), image.Rect(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y),
		image.Rect(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), image.Rect(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y),
	}
}

// covered is the cells the flex paints: its border and its items. The rest
// of its rectangle is left as it was on the screen.
func (f *nativeFlex) covered() []image.Rectangle {
	out := f.own()
	for _, it := range f.items {
		if it.w != nil {
			out = append(out, coveredBy(it.w)...)
		}
	}
	return out
}

// Covers: the flex itself is only its border; the rest is its items' or a
// gap, which leaves the pointer to the page beneath.
func (f *nativeFlex) Covers(pt image.Point) bool {
	return f.framed() && !pt.In(f.inner(f.Bounds()))
}

// Focus gives the keyboard to the first item added with focus, or keeps it
// when there is none: as tview's Flex does.
func (f *nativeFlex) Focus(delegate func(p Primitive)) {
	for _, it := range f.items {
		if it.prim != nil && it.focus {
			delegate(it.prim)
			return
		}
	}
	f.primBase.Focus(delegate)
}

// HasFocus reports whether the flex or anything in it holds the keyboard.
func (f *nativeFlex) HasFocus() bool {
	for _, it := range f.items {
		if it.prim != nil && it.prim.HasFocus() {
			return true
		}
	}
	return f.primBase.HasFocus()
}

// A nativeCanvas is a box filled by a func that still writes tcell's cells,
// at the box's place on the screen (the theme picker's sample).
type nativeCanvas struct {
	*nativeHost
	kit.Base
	draw func(scr mapstyle.Cells, x, y, w, h int)
}

func newNativeCanvas(draw func(scr mapstyle.Cells, x, y, w, h int)) *nativeCanvas {
	c := &nativeCanvas{draw: draw}
	c.nativeHost = newNativeHost()
	return c
}

func (c *nativeCanvas) Layout(r image.Rectangle) { c.Place(r) }

// Paint clears the box to the page's ground, as a tview.Box did, then draws.
func (c *nativeCanvas) Paint(p *kit.Painter) {
	if c.nativeHost == nil || c.Bounds().Empty() {
		return
	}
	p.Fill(c.Bounds(), kit.Role(theme.RoleBackground))
	x, y, w, h := c.GetRect()
	c.draw(cellsFor(p, c, c.Bounds()), x, y, w, h)
}

// kitCells is a kit painter as the map draws on one (mapstyle.Cells). The
// map works in the screen's coordinates; off is how far those are from the
// painter's.
type kitCells struct {
	p   *kit.Painter
	off image.Point
}

// cellsFor is the painter as the map draws on it, for a primitive laid out
// at r: the map is handed the primitive's rectangle on the screen.
func cellsFor(p *kit.Painter, prim interface{ GetRect() (int, int, int, int) }, r image.Rectangle) kitCells {
	return kitCells{p: p, off: screenOffset(prim, r)}
}

func (c kitCells) SetContent(x, y int, r rune, _ []rune, st kit.Style) {
	c.p.Put(x-c.off.X, y-c.off.Y, r, st)
}

// GetContent reads a cell back as it was painted: its colors are literals,
// and a ground that is none is Transparent.
func (c kitCells) GetContent(x, y int) (rune, []rune, kit.Style, int) {
	cell := c.p.At(x-c.off.X, y-c.off.Y)
	st := kit.Style{Fg: kit.RGB(cell.Fg.R, cell.Fg.G, cell.Fg.B), Bg: kit.Clear, Attr: cell.Attr}
	if !cell.Clear() {
		st.Bg = kit.RGB(cell.Bg.R, cell.Bg.G, cell.Bg.B)
	}
	return cell.R, nil, st, 1
}
