package kit

import (
	"image"
	"image/color"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/theme"
)

// screen paints w at cols by rows and returns the root and the buffer.
func screen(w Widget, cols, rows int) (*Root, *Buffer) {
	r := NewRoot(w)
	r.Resize(cols, rows)
	b := NewBuffer(cols, rows)
	r.Paint(b)
	return r, b
}

func key(c KeyCode) Key { return Key{Code: c} }

func TestTagsSetColorsAndAttributes(t *testing.T) {
	_, b := screen(Label("[gold]a[-]b[:red]c[:-][::bu]d[::U]e[::-]f[#102030]g"), 20, 1)
	fg := func(x int) any { return b.At(x, 0).Fg }
	if fg(0) != theme.RGBA(theme.RoleAccent) || fg(1) != theme.RGBA(theme.RoleText) {
		t.Errorf("gold then reset: %v, %v", fg(0), fg(1))
	}
	if b.At(2, 0).Bg != theme.RGBA(theme.RoleNegative) || b.At(3, 0).Bg != theme.RGBA(theme.RoleBackground) {
		t.Errorf("a ground and its reset: %v, %v", b.At(2, 0).Bg, b.At(3, 0).Bg)
	}
	if a := b.At(3, 0).Attr; a != Bold|Underline {
		t.Errorf("[::bu] gave %b", a)
	}
	if a := b.At(4, 0).Attr; a != Bold {
		t.Errorf("[::U] should leave bold only, got %b", a)
	}
	if a := b.At(5, 0).Attr; a != 0 {
		t.Errorf("[::-] should clear the attributes, got %b", a)
	}
	if c := b.At(6, 0).Fg; c.R != 0x10 || c.G != 0x20 || c.B != 0x30 {
		t.Errorf("a hex color: %v", c)
	}
	if got := b.Line(0); got != "abcdefg" {
		t.Errorf("the text is %q", got)
	}
}

func TestATagFollowsTheThemeWhenItChanges(t *testing.T) {
	defer func() { _ = theme.SetActive(theme.DefaultKey) }()
	root := NewRoot(Label("[accent]x"))
	root.Resize(3, 1)
	b := NewBuffer(3, 1)
	root.Paint(b)
	before := b.At(0, 0)
	var other string
	for _, th := range theme.All() {
		if th.Color(theme.RoleAccent) != theme.Active().Color(theme.RoleAccent) && th.Color(theme.RoleBackground) != theme.Active().Color(theme.RoleBackground) {
			other = th.Key
			break
		}
	}
	if err := theme.SetActive(other); err != nil {
		t.Fatal(err)
	}
	root.Paint(b)
	if after := b.At(0, 0); after.Fg == before.Fg || after.Bg == before.Bg {
		t.Errorf("the cell kept the old theme's colors: %v then %v", before, after)
	}
}

func TestEscapedTagsWrapLikeTheirText(t *testing.T) {
	got := WordWrap("press [Enter[] to keep or [Esc[] to leave", 14)
	want := []string{"press [Enter[] ", "to keep or ", "[Esc[] to leave"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if s := Strip("[red[[] and [x[]"); s != "[red[] and [x]" {
		t.Errorf("Strip gave %q", s)
	}
}

func TestWideGlyphsNeverSplit(t *testing.T) {
	_, b := screen(Label("ab你好"), 5, 1)
	if got := b.Line(0); got != "ab你" {
		t.Errorf("a wide glyph with half its room: %q", got)
	}
	if !b.At(2, 0).Wide || !b.At(3, 0).Cont {
		t.Error("a wide glyph marks both its cells")
	}
	// Writing over half of one blanks the other half.
	p := NewPainter(b)
	p.Put(3, 0, 'x', Style{})
	if b.At(2, 0).R != ' ' || b.At(2, 0).Wide {
		t.Errorf("the left half stayed: %+v", b.At(2, 0))
	}
}

func TestFlexSharesTheRoom(t *testing.T) {
	a, b, c := &Custom{}, &Custom{}, Label("12345")
	row := Row().Add(a, 4, 0).Add(b, 0, 1).Add(nil, 0, 1).Add(c, 0, 0)
	row.Gap = 1
	screen(row, 30, 3)
	if got := a.Bounds(); got != image.Rect(0, 0, 4, 3) {
		t.Errorf("the fixed item: %v", got)
	}
	// 30 less 4 fixed, 5 natural and 3 gaps leaves 18 for two weights.
	if got := b.Bounds(); got != image.Rect(5, 0, 14, 3) {
		t.Errorf("the weighted item: %v", got)
	}
	if got := c.Bounds(); got != image.Rect(25, 0, 30, 3) {
		t.Errorf("the natural item: %v", got)
	}
	b.Hide = true
	screen(row, 30, 3)
	if got := c.Bounds(); got != image.Rect(25, 0, 30, 3) {
		t.Errorf("with one hidden the gap takes its room: %v", got)
	}
	col := Col().Add(Label("one"), 0, 0).Add(Label("two\nlines"), 0, 0)
	if got := col.Size(image.Pt(40, 40)); got != image.Pt(5, 3) {
		t.Errorf("a column's natural size: %v", got)
	}
}

func TestPanelDrawsItsBoxTitleAndFoot(t *testing.T) {
	pn := NewPanel("[gold]Top[-]", Label("in"))
	pn.Foot, pn.PadX = "foot", 1
	_, b := screen(pn, 12, 4)
	want := []string{
		"┌────Top───┐",
		"│ in       │",
		"│          │",
		"└───foot───┘",
	}
	if got := b.Lines(); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%s", strings.Join(got, "\n"))
	}
	if got := pn.Size(image.Pt(40, 40)); got != image.Pt(6, 3) {
		t.Errorf("natural size %v", got)
	}
}

func TestListMovesChoosesAndScrolls(t *testing.T) {
	var chose, moved []int
	l := NewList("zero", "one", "two", "three", "four", "five")
	l.OnChoose = func(i int) { chose = append(chose, i) }
	l.OnMove = func(i int) { moved = append(moved, i) }
	r, b := screen(l, 8, 3)
	if r.Focused() != l {
		t.Fatal("the list should have the focus")
	}
	for _, k := range []KeyCode{KeyDown, KeyDown, KeyDown, KeyEnter, KeyEnd, KeyHome, KeyUp} {
		r.Key(key(k))
	}
	if !reflect.DeepEqual(chose, []int{3}) || !reflect.DeepEqual(moved, []int{1, 2, 3, 5, 0}) {
		t.Errorf("chose %v, moved %v", chose, moved)
	}
	r.Key(key(KeyPgDn))
	r.Key(key(KeyPgDn))
	r.Paint(b)
	for y, want := range []string{"two", "three", "four"} {
		if got := strings.Fields(b.Line(y))[0]; got != want {
			t.Errorf("after two pages row %d is %q, want %q", y, got, want)
		}
	}
	if b.At(0, 2).Bg != theme.RGBA(theme.RoleSelection) {
		t.Error("the selected row is not on the selection's ground")
	}
	if b.At(7, 0).R == ' ' && b.At(7, 1).R == ' ' && b.At(7, 2).R == ' ' {
		t.Error("no scroll bar though rows are out of view")
	}
	l.Loop = true
	r.Key(key(KeyEnd))
	r.Key(key(KeyDown))
	if l.Selected() != 0 {
		t.Errorf("a looping list should come round, at %d", l.Selected())
	}
}

func TestListMouse(t *testing.T) {
	var chose []int
	l := NewList("zero", "one", "two", "three", "four", "five")
	l.OnChoose = func(i int) { chose = append(chose, i) }
	r, b := screen(l, 8, 3)
	now := time.Unix(100, 0)
	r.Now = func() time.Time { return now }
	if !r.Mouse(Mouse{Kind: MouseMove, X: 2, Y: 1}) {
		t.Error("a move onto a row changes the screen")
	}
	r.Paint(b)
	if b.At(0, 1).Bg != theme.RGBA(theme.RoleChip) {
		t.Error("the row under the pointer is not marked")
	}
	r.Mouse(Mouse{Kind: MouseLeave})
	r.Paint(b)
	if b.At(0, 1).Bg == theme.RGBA(theme.RoleChip) {
		t.Error("the mark stayed after the pointer left")
	}
	r.Mouse(Mouse{Kind: MousePress, X: 2, Y: 1})
	r.Mouse(Mouse{Kind: MouseRelease, X: 2, Y: 1})
	if l.Selected() != 1 || len(chose) != 0 {
		t.Errorf("one click selects: at %d, chose %v", l.Selected(), chose)
	}
	now = now.Add(200 * time.Millisecond)
	r.Mouse(Mouse{Kind: MousePress, X: 2, Y: 1})
	if !reflect.DeepEqual(chose, []int{1}) {
		t.Errorf("a double click chooses: %v", chose)
	}
	r.Mouse(Mouse{Kind: MouseWheel, X: 2, Y: 1, Wheel: 1})
	r.Paint(b)
	if got := b.Line(0); got != "three" {
		t.Errorf("after a wheel notch the top row is %q", got)
	}
	l.ClickChooses = true
	now = now.Add(time.Minute)
	r.Mouse(Mouse{Kind: MousePress, X: 2, Y: 0})
	if !reflect.DeepEqual(chose, []int{1, 3}) {
		t.Errorf("with ClickChooses one click chooses: %v", chose)
	}
}

func TestButtonPressesByKeyAndMouse(t *testing.T) {
	n := 0
	bt := NewButton("Go", func() { n++ })
	r, b := screen(Row().Add(bt, 0, 0).Add(nil, 0, 1), 10, 1)
	if got := b.Line(0); got != " Go" {
		t.Errorf("painted %q", got)
	}
	r.Key(key(KeyEnter))
	r.Key(Key{Code: KeyRune, Rune: ' '})
	r.Mouse(Mouse{Kind: MousePress, X: 1, Y: 0})
	r.Mouse(Mouse{Kind: MouseRelease, X: 1, Y: 0})
	if n != 3 {
		t.Errorf("pressed %d times, want 3", n)
	}
	// A press that is let go somewhere else is not a press.
	r.Mouse(Mouse{Kind: MousePress, X: 1, Y: 0})
	r.Mouse(Mouse{Kind: MouseMove, X: 8, Y: 0})
	r.Mouse(Mouse{Kind: MouseRelease, X: 8, Y: 0})
	if n != 3 {
		t.Errorf("a press dragged off the button counted: %d", n)
	}
	// Nor is one that ends with the pointer leaving the screen.
	r.Mouse(Mouse{Kind: MousePress, X: 1, Y: 0})
	r.Mouse(Mouse{Kind: MouseLeave})
	r.Mouse(Mouse{Kind: MouseMove, X: 1, Y: 0})
	r.Mouse(Mouse{Kind: MouseRelease, X: 1, Y: 0})
	if n != 3 {
		t.Errorf("a press that left the screen counted: %d", n)
	}
	bt.Disabled = true
	r.Mouse(Mouse{Kind: MousePress, X: 1, Y: 0})
	r.Mouse(Mouse{Kind: MouseRelease, X: 1, Y: 0})
	if n != 3 || bt.CanFocus() {
		t.Error("a disabled button was pressed or takes the focus")
	}
}

// The pointer takes the shape of what it is over: the hand on a button and
// on a row of a list, the text bar on an input line, the arrow elsewhere and
// once it has left. A move that changes the shape is a change to the screen,
// and the frame carries the shape. While a press is held the shape stays.
func TestThePointerTakesTheShapeOfWhatItIsOver(t *testing.T) {
	bt := NewButton("Go", func() {})
	list := NewList("one", "two")
	in := NewInput("> ")
	r, b := screen(Col().Add(Row().Add(bt, 0, 0).Add(nil, 0, 1), 1, 0).Add(list, 3, 0).Add(in, 1, 0).Add(nil, 0, 1), 20, 6)
	at := func(x, y int) Pointer {
		t.Helper()
		r.Mouse(Mouse{Kind: MouseMove, X: x, Y: y})
		r.Paint(b)
		return b.Pointer
	}
	if b.Pointer != PointerArrow {
		t.Errorf("before the pointer arrives the shape is %d, want the arrow", b.Pointer)
	}
	for _, c := range []struct {
		what string
		x, y int
		want Pointer
	}{
		{"the button", 1, 0, PointerHand},
		{"the gap beside the button", 12, 0, PointerArrow},
		{"a row of the list", 2, 2, PointerHand},
		{"the list under its last row", 2, 3, PointerArrow},
		{"the input line", 5, 4, PointerText},
		{"the empty rows", 5, 5, PointerArrow},
	} {
		if got := at(c.x, c.y); got != c.want {
			t.Errorf("over %s the shape is %d, want %d", c.what, got, c.want)
		}
	}
	// A move inside the input line changes nothing; a move onto the gap
	// beside it changes only the shape, and that counts as a change.
	at(5, 4)
	if r.Mouse(Mouse{Kind: MouseMove, X: 6, Y: 4}) {
		t.Error("a move along the input line reported a change")
	}
	if !r.Mouse(Mouse{Kind: MouseMove, X: 6, Y: 5}) {
		t.Error("a move off the input line, which changes the shape, reported no change")
	}
	bt.Disabled = true
	if got := at(1, 0); got != PointerArrow {
		t.Errorf("over a disabled button the shape is %d, want the arrow", got)
	}
	bt.Disabled = false
	// A press held on the button keeps the hand while it is dragged off.
	at(1, 0)
	r.Mouse(Mouse{Kind: MousePress, X: 1, Y: 0})
	if got := at(12, 5); got != PointerHand {
		t.Errorf("with a press held, dragged off the button, the shape is %d, want the hand still", got)
	}
	r.Mouse(Mouse{Kind: MouseRelease, X: 12, Y: 5})
	r.Paint(b)
	if b.Pointer != PointerArrow {
		t.Errorf("after the button came up away from it the shape is %d, want the arrow", b.Pointer)
	}
	at(1, 0)
	r.Mouse(Mouse{Kind: MouseLeave})
	r.Paint(b)
	if b.Pointer != PointerArrow {
		t.Errorf("after the pointer left the shape is %d, want the arrow", b.Pointer)
	}
}

// Took says whether a widget took the last mouse event: a move onto a
// list's row is taken (the list marks the row) and so is a left press on
// it; a press of the right button on it is not, nor a press under the last
// row.
func TestTheRootSaysWhetherAWidgetTookTheMouse(t *testing.T) {
	l := NewList("zero", "one")
	r, _ := screen(l, 8, 3)
	if !r.Mouse(Mouse{Kind: MouseMove, X: 2, Y: 0}) || r.Took() != true {
		t.Errorf("a move onto a row: took %v, want the list to have taken it (it marks the row)", r.Took())
	}
	if r.Mouse(Mouse{Kind: MousePress, Button: ButtonRight, X: 2, Y: 0}); r.Took() {
		t.Error("a right press on a row was taken")
	}
	if r.Mouse(Mouse{Kind: MousePress, Button: ButtonLeft, X: 2, Y: 1}); !r.Took() {
		t.Error("a left press on a row was not taken")
	}
	r.Mouse(Mouse{Kind: MouseRelease, Button: ButtonLeft, X: 2, Y: 1})
	if r.Mouse(Mouse{Kind: MousePress, Button: ButtonLeft, X: 2, Y: 2}); r.Took() {
		t.Error("a press under the last row was taken")
	}
}

// The click region under the pointer is marked with the chip's ground: just
// its own text where the line holds another region, the whole width of the
// box where it holds none, and never a cell that has a ground of its own.
// Text that takes no clicks marks nothing.
func TestTheRegionUnderThePointerIsMarked(t *testing.T) {
	const text = `["a"]one[""] ["b"]two[""]` + "\n" + `["c"]three[""]` + "\n" + `["c"]four [#000000:#ff0000]hot[-:-][""]` + "\nplain"
	tx := NewText(text)
	tx.OnRegion = func(string, int) {}
	r, b := screen(tx, 12, 4)
	chip := theme.RGBA(theme.RoleChip)
	marks := func() string {
		r.Paint(b)
		var sb strings.Builder
		for y := 0; y < 4; y++ {
			for x := 0; x < 12; x++ {
				if b.At(x, y).Bg == chip {
					sb.WriteByte('#')
				} else {
					sb.WriteByte('.')
				}
			}
			sb.WriteByte('\n')
		}
		return sb.String()
	}
	const none = "............\n............\n............\n............\n"
	for _, c := range []struct {
		what string
		x, y int
		want string
	}{
		{"before the pointer arrives", -1, -1, none},
		{"the first region of a line of two", 1, 0, "###.........\n............\n............\n............\n"},
		{"the second region of that line", 5, 0, "....###.....\n............\n............\n............\n"},
		{"a region of two whole lines", 1, 1, "............\n############\n#####...####\n............\n"},
		{"plain text", 1, 3, none},
	} {
		if c.x >= 0 {
			r.Mouse(Mouse{Kind: MouseMove, X: c.x, Y: c.y})
		}
		if got := marks(); got != c.want {
			t.Errorf("%s: marked\n%swant\n%s", c.what, got, c.want)
		}
	}
	r.Mouse(Mouse{Kind: MouseMove, X: 1, Y: 1})
	r.Mouse(Mouse{Kind: MouseLeave})
	if got := marks(); got != none {
		t.Errorf("after the pointer left the mark stayed:\n%s", got)
	}
	tx.OnRegion = nil
	r.Mouse(Mouse{Kind: MouseMove, X: 1, Y: 1})
	if got := marks(); got != none {
		t.Errorf("text that takes no clicks marked a region:\n%s", got)
	}
}

// A double click is two presses of the left button. A press of another
// button on the same cell just before does not make the next left press the
// second of a pair.
func TestARightPressDoesNotBeginADoubleClick(t *testing.T) {
	var chose []int
	l := NewList("zero", "one", "two")
	l.OnChoose = func(i int) { chose = append(chose, i) }
	r, _ := screen(l, 8, 3)
	now := time.Unix(100, 0)
	r.Now = func() time.Time { return now }
	r.Mouse(Mouse{Kind: MousePress, Button: ButtonRight, X: 2, Y: 1})
	r.Mouse(Mouse{Kind: MouseRelease, Button: ButtonRight, X: 2, Y: 1})
	now = now.Add(50 * time.Millisecond)
	r.Mouse(Mouse{Kind: MousePress, Button: ButtonLeft, X: 2, Y: 1})
	r.Mouse(Mouse{Kind: MouseRelease, Button: ButtonLeft, X: 2, Y: 1})
	if len(chose) != 0 || l.Selected() != 1 {
		t.Errorf("a right press then a left press chose %v (row %d selected), want the row selected and nothing chosen", chose, l.Selected())
	}
	// And a right press between the two halves of a double click breaks it.
	now = now.Add(time.Minute)
	r.Mouse(Mouse{Kind: MousePress, Button: ButtonLeft, X: 2, Y: 2})
	r.Mouse(Mouse{Kind: MousePress, Button: ButtonRight, X: 2, Y: 2})
	now = now.Add(50 * time.Millisecond)
	r.Mouse(Mouse{Kind: MousePress, Button: ButtonLeft, X: 2, Y: 2})
	if len(chose) != 0 {
		t.Errorf("left, right, left on one cell chose %v, want nothing chosen", chose)
	}
	now = now.Add(50 * time.Millisecond)
	r.Mouse(Mouse{Kind: MousePress, Button: ButtonLeft, X: 2, Y: 2})
	if !reflect.DeepEqual(chose, []int{2}) {
		t.Errorf("two left presses after that chose %v, want row 2", chose)
	}
}

// A picture marks the cells of its rectangle that the painter may touch, and
// whatever is painted on one of them afterwards takes that cell back: text,
// a fill, another picture. A new frame starts with no pictures.
func TestAPictureKeepsItsCellsUntilSomethingIsPaintedOnThem(t *testing.T) {
	b := NewBuffer(10, 4)
	p := NewPainter(b)
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	p.Print(0, 1, -1, "abcdefghij", Style{})
	p.Picture(image.Rect(2, 1, 6, 3), img)
	shows := func() string {
		var sb strings.Builder
		for y := 0; y < 4; y++ {
			for x := 0; x < 10; x++ {
				sb.WriteByte("._23456789"[min(int(b.At(x, y).Pic), 9)])
			}
			sb.WriteByte('\n')
		}
		return sb.String()
	}
	if got, want := shows(), "..........\n..____....\n..____....\n..........\n"; got != want {
		t.Fatalf("the picture's cells:\n%swant\n%s", got, want)
	}
	if len(b.Pics) != 1 || b.Pics[0].Rect != image.Rect(2, 1, 6, 3) || b.Pics[0].Img != img {
		t.Fatalf("the buffer lists %v", b.Pics)
	}
	if b.At(2, 1).R != ' ' || b.At(1, 1).R != 'b' {
		t.Errorf("under the picture the text should be gone and beside it stay: %q %q", b.At(2, 1).R, b.At(1, 1).R)
	}
	// Text, a fill and a second picture each take cells back.
	p.Print(3, 1, -1, "X", Style{})
	p.Fill(image.Rect(5, 2, 8, 3), Role(theme.RoleSurface))
	p.Picture(image.Rect(2, 2, 3, 4), img)
	if got, want := shows(), "..........\n.._.__....\n..2__.....\n..2.......\n"; got != want {
		t.Errorf("after text, a fill and another picture:\n%swant\n%s", got, want)
	}
	// A clipped painter marks only what it may touch, but the picture keeps its whole rectangle.
	b.Reset()
	if len(b.Pics) != 0 || b.At(2, 1).Pic != 0 {
		t.Fatal("a new frame still has the last one's pictures")
	}
	p.Clip(image.Rect(0, 0, 4, 4)).Picture(image.Rect(2, 0, 8, 2), img)
	if got, want := shows(), "..__......\n..__......\n..........\n..........\n"; got != want {
		t.Errorf("a clipped picture:\n%swant\n%s", got, want)
	}
	if b.Pics[0].Rect != image.Rect(2, 0, 8, 2) {
		t.Errorf("a clipped picture's rectangle is %v, want the whole of it", b.Pics[0].Rect)
	}
	p.Picture(image.Rect(0, 0, 2, 2), nil)
	if len(b.Pics) != 1 {
		t.Error("a picture with no image was listed")
	}
}

// A label is text over a picture: its cells keep the picture and the ground,
// take the glyph and the ink, and honour the clip; Print on the same cells
// still takes them back.
func TestALabelKeepsThePictureUnderItsText(t *testing.T) {
	b := NewBuffer(10, 2)
	p := NewPainter(b)
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	ground := RGB(1, 2, 3)
	p.Fill(image.Rect(0, 0, 10, 2), ground)
	p.Picture(image.Rect(2, 0, 7, 1), img)
	ink := Style{}.Foreground(RGB(200, 100, 50)).Bold(true)
	if n := p.Label(3, 0, "Esc", ink); n != 3 {
		t.Fatalf("Label took %d cells, want 3", n)
	}
	for i, r := range "Esc" {
		c := b.At(3+i, 0)
		if c.R != r || c.Pic != 1 || c.Fg != (color.RGBA{200, 100, 50, 255}) || c.Attr&Bold == 0 {
			t.Errorf("cell %d is %+v, want %q over the picture in the ink", 3+i, c, r)
		}
		if bg, _ := ground.RGBA(); c.Bg != bg {
			t.Errorf("cell %d lost its ground: %v", 3+i, c.Bg)
		}
	}
	if b.At(2, 0).Pic != 1 || b.At(2, 0).R != ' ' {
		t.Errorf("a cell beside the text changed: %+v", b.At(2, 0))
	}
	// A style with a ground of its own does not repaint the cell's.
	p.Label(2, 0, "x", Style{}.Background(RGB(9, 9, 9)))
	if bg, _ := ground.RGBA(); b.At(2, 0).Bg != bg || b.At(2, 0).Pic != 1 {
		t.Errorf("a label's own ground replaced the cell's: %+v", b.At(2, 0))
	}
	// The clip holds.
	if n := p.Clip(image.Rect(0, 0, 4, 2)).Label(3, 0, "abc", ink); n != 3 || b.At(3, 0).R != 'a' || b.At(4, 0).R != 's' {
		t.Errorf("a clipped label wrote past its clip: %q %q (took %d)", b.At(3, 0).R, b.At(4, 0).R, n)
	}
	// Print takes the picture back.
	p.Print(3, 0, -1, "P", ink)
	if b.At(3, 0).Pic != 0 || b.At(4, 0).Pic != 1 {
		t.Errorf("Print left the picture on its cell, or took the next: %d %d", b.At(3, 0).Pic, b.At(4, 0).Pic)
	}
}

func lines(n int) string {
	var s []string
	for i := 0; i < n; i++ {
		s = append(s, "line "+string(rune('a'+i)))
	}
	return strings.Join(s, "\n")
}

func TestScrollByKeysWheelAndDrag(t *testing.T) {
	sc := NewScroll(Label(lines(20)))
	sc.Focus = true
	r, b := screen(sc, 10, 5)
	top := func() string { r.Paint(b); return b.Line(0)[:6] }
	if top() != "line a" {
		t.Fatalf("starts at %q", top())
	}
	r.Key(key(KeyDown))
	r.Key(key(KeyPgDn))
	if top() != "line f" {
		t.Errorf("down one and a page: %q", top())
	}
	r.Key(key(KeyEnd))
	if top() != "line p" || !sc.AtEnd() {
		t.Errorf("End: %q", top())
	}
	r.Mouse(Mouse{Kind: MouseWheel, X: 3, Y: 2, Wheel: -1})
	if top() != "line m" {
		t.Errorf("a wheel notch up: %q", top())
	}
	// Drag the thumb from the bottom of the bar to its top.
	r.Mouse(Mouse{Kind: MousePress, X: 9, Y: 4})
	r.Mouse(Mouse{Kind: MouseMove, X: 9, Y: 0})
	r.Mouse(Mouse{Kind: MouseRelease, X: 9, Y: 0})
	if top() != "line a" {
		t.Errorf("dragged to the top: %q", top())
	}
	if b.At(9, 0).R == ' ' {
		t.Error("no thumb at the top of the bar")
	}
}

func TestScrollFollowsAGrowingLog(t *testing.T) {
	tx := Label(lines(3))
	sc := NewScroll(tx)
	sc.Follow = true
	r, b := screen(sc, 10, 4)
	tx.SetText(lines(10))
	r.Paint(b)
	if got := b.Line(3)[:6]; got != "line j" {
		t.Errorf("the last line should stay in view, bottom row is %q", got)
	}
	sc.ScrollTo(0)
	tx.SetText(lines(12))
	r.Paint(b)
	if got := b.Line(0)[:6]; got != "line a" {
		t.Errorf("scrolled away, it should stay put: %q", got)
	}
	// A scroll asked for before the text is laid out holds at the next
	// paint, and one past the end stops at the end.
	tx.SetText(lines(26))
	sc.ScrollTo(20)
	r.Paint(b)
	if got := b.Line(0)[:6]; got != "line u" {
		t.Errorf("ScrollTo before the layout: %q", got)
	}
	sc.ScrollTo(500)
	r.Paint(b)
	if got := b.Line(3)[:6]; got != "line z" || sc.Top() != 22 {
		t.Errorf("past the end: bottom row %q, top %d", got, sc.Top())
	}
	sc.NoBar = true
	r.Paint(b)
	if b.At(9, 0).R != ' ' {
		t.Error("NoBar should leave the bar's column to the child")
	}
}

// a probe notes the keys and mouse events it is offered.
type probe struct {
	Custom
	name string
	log  *[]string
	use  bool
}

func newProbe(name string, log *[]string, use bool) *probe {
	p := &probe{name: name, log: log, use: use}
	p.Focus = true
	p.OnKey = func(Key) bool { *log = append(*log, name); return p.use }
	p.OnMouse = func(m Mouse) bool {
		*log = append(*log, name+[]string{" move", " press", " release", " wheel", " leave"}[m.Kind])
		return p.use
	}
	return p
}

// around is a parent that hears the keys its child left.
type around struct {
	Stack
	log *[]string
}

func (a *around) Key(Key) bool { *a.log = append(*a.log, "around"); return false }

func TestKeysGoFocusThenOutwardThenGlobal(t *testing.T) {
	var log []string
	in := newProbe("in", &log, false)
	out := &around{log: &log}
	out.Items = []Widget{in}
	r, _ := screen(out, 10, 2)
	r.Keys = func(Key) bool { log = append(log, "global"); return true }
	if !r.Key(key(KeyEnter)) {
		t.Error("the global handler used the key")
	}
	if want := []string{"in", "around", "global"}; !reflect.DeepEqual(log, want) {
		t.Errorf("order %v, want %v", log, want)
	}
	log, in.use = nil, true
	r.Key(key(KeyEnter))
	if want := []string{"in"}; !reflect.DeepEqual(log, want) {
		t.Errorf("a used key went on: %v", log)
	}
}

// sieve is a parent that filters the keys on their way to its child.
type sieve struct {
	Stack
	log  *[]string
	name string
	stop bool
	swap *Key
}

func (s *sieve) FilterKey(k Key) (Key, bool) {
	*s.log = append(*s.log, s.name)
	if s.swap != nil {
		k = *s.swap
	}
	return k, !s.stop
}

// A key passes each filter from the outside in before the focused widget
// sees it; a filter may stop it or put another key in its place.
func TestFiltersSeeAKeyOnItsWayDown(t *testing.T) {
	var log []string
	var got []Key
	in := newProbe("in", &log, true)
	in.OnKey = func(k Key) bool { log, got = append(log, "in"), append(got, k); return true }
	inner := &sieve{log: &log, name: "inner"}
	inner.Items = []Widget{in}
	outer := &sieve{log: &log, name: "outer"}
	outer.Items = []Widget{inner}
	r, _ := screen(outer, 10, 2)

	r.Key(key(KeyEnter))
	if want := []string{"outer", "inner", "in"}; !reflect.DeepEqual(log, want) {
		t.Errorf("order %v, want %v", log, want)
	}
	log, got = nil, nil
	home := key(KeyHome)
	outer.swap = &home
	r.Key(key(KeyEnter))
	if len(got) != 1 || got[0].Code != KeyHome {
		t.Errorf("the key a filter put in its place did not arrive: %v", got)
	}
	log, outer.swap, inner.stop = nil, nil, true
	if !r.Key(key(KeyEnter)) {
		t.Error("a key a filter stopped counts as used")
	}
	if want := []string{"outer", "inner"}; !reflect.DeepEqual(log, want) {
		t.Errorf("a stopped key went on: %v", log)
	}
}

// A widget that is no Focusable can still be given the keys: its filters and
// handlers hear them, and Tab never lands on it.
func TestAnyWidgetCanBeGivenTheKeys(t *testing.T) {
	var log []string
	a, b := newProbe("a", &log, true), newProbe("b", &log, true)
	box := &sieve{log: &log, name: "box"}
	box.Items = []Widget{b}
	row := Row().Add(a, 0, 1).Add(box, 0, 1)
	r, _ := screen(row, 10, 2)
	r.Focus(box)
	if r.Focused() != Widget(box) || a.Focused() {
		t.Fatalf("the box should hold the keys, %v does", r.Focused())
	}
	r.Key(key(KeyEnter))
	if want := []string{"box"}; !reflect.DeepEqual(log, want) {
		t.Errorf("keys went to %v, want only the box's filter", log)
	}
	r.Focus(b)
	if !b.Focused() {
		t.Error("the focus did not move on from a widget that is no Focusable")
	}
}

// holed is a layer that is only really there on its left half.
type holed struct{ Flex }

func (h *holed) Covers(pt image.Point) bool { return pt.X < 5 }

// The pointer goes through a hole in the layer on top to the widget under
// it; where the top layer has something of its own, that takes it, and
// where the layer is there but empty nothing does.
func TestThePointerGoesThroughAHole(t *testing.T) {
	var log []string
	under := newProbe("under", &log, true)
	knob := newProbe("knob", &log, true)
	top := &holed{}
	top.Add(knob, 3, 0).Add(nil, 0, 1)
	r, _ := screen(&Stack{Items: []Widget{under, top}}, 10, 3)
	r.Mouse(Mouse{Kind: MousePress, X: 8, Y: 0}) // a hole: the widget under
	r.Mouse(Mouse{Kind: MouseRelease, X: 8, Y: 0})
	r.Mouse(Mouse{Kind: MousePress, X: 1, Y: 1}) // the knob
	r.Mouse(Mouse{Kind: MouseRelease, X: 1, Y: 1})
	r.Mouse(Mouse{Kind: MousePress, X: 4, Y: 1}) // the layer itself: nobody's
	want := []string{"under press", "under release", "under leave", "knob press", "knob release", "knob leave"}
	if !reflect.DeepEqual(log, want) {
		t.Errorf("the mouse went %v, want %v", log, want)
	}
}

func TestTabWalksTheFocus(t *testing.T) {
	var log []string
	a, b, c := newProbe("a", &log, false), newProbe("b", &log, false), newProbe("c", &log, false)
	b.Hide = true
	r, _ := screen(Row().Add(a, 0, 1).Add(b, 0, 1).Add(c, 0, 1), 12, 1)
	if r.Focused() != a {
		t.Fatal("the first widget that takes the focus has it")
	}
	r.Key(key(KeyTab))
	if r.Focused() != c || !c.Focused() || a.Focused() {
		t.Error("Tab should pass the hidden widget and land on c")
	}
	r.Key(key(KeyTab))
	if r.Focused() != a {
		t.Error("Tab should come round")
	}
	r.Key(key(KeyBacktab))
	if r.Focused() != c {
		t.Error("Shift+Tab should go back")
	}
}

func TestAModalTakesTheInput(t *testing.T) {
	var log []string
	under := newProbe("under", &log, true)
	r, b := screen(under, 20, 6)
	r.Keys = func(Key) bool { log = append(log, "global"); return true }
	closed := 0
	ok := NewButton("OK", nil)
	no := NewButton("No", nil)
	md := &Modal{Child: NewPanel("Sure?", Row().Add(ok, 0, 0).Add(no, 0, 0)), W: 12, H: 3, OnDismiss: func() { closed++ }}
	before := b.At(0, 0).Bg
	r.Push(md)
	r.Paint(b)
	if under.Focused() || r.Focused() != ok {
		t.Error("the focus should be inside the modal")
	}
	if b.At(0, 0).Bg == before && before != theme.RGBA(theme.RoleBackground) {
		t.Error("the screen behind is not dimmed")
	}
	if got := b.Line(1); !strings.Contains(got, "Sure?") {
		t.Errorf("the modal is not in the middle: %q", got)
	}
	r.Key(Key{Code: KeyRune, Rune: 'x'})
	r.Mouse(Mouse{Kind: MousePress, X: 0, Y: 5})
	if len(log) != 0 {
		t.Errorf("input reached past the modal: %v", log)
	}
	if closed != 1 {
		t.Errorf("a press outside should dismiss, got %d", closed)
	}
	r.Key(key(KeyTab))
	if r.Focused() != no {
		t.Error("Tab should walk inside the modal")
	}
	r.Key(key(KeyEsc))
	if closed != 2 {
		t.Errorf("Esc should dismiss, got %d", closed)
	}
	r.Remove(md)
	if r.Depth() != 1 || !under.Focused() {
		t.Error("with the modal gone the screen has the focus back")
	}
	r.Key(key(KeyEnter))
	if !reflect.DeepEqual(log, []string{"under"}) {
		t.Errorf("after the modal: %v", log)
	}
}

func TestMouseGoesToTheDeepestWidgetAndBubbles(t *testing.T) {
	var log []string
	left, right := newProbe("left", &log, false), newProbe("right", &log, true)
	back := newProbe("back", &log, true)
	st := &Stack{Items: []Widget{back, Row().Add(left, 0, 1).Add(right, 0, 1)}}
	r, _ := screen(st, 10, 2)
	r.Mouse(Mouse{Kind: MousePress, X: 7, Y: 0})
	if r.Focused() != right {
		t.Error("a press focuses what it lands on")
	}
	// The press was right's, so the drag and the release are too, wherever
	// the pointer goes.
	r.Mouse(Mouse{Kind: MouseMove, X: 1, Y: 0})
	r.Mouse(Mouse{Kind: MouseRelease, X: 1, Y: 0})
	if want := []string{"right press", "right move", "right release"}; !reflect.DeepEqual(log, want) {
		t.Errorf("got %v, want %v", log, want)
	}
	log = nil
	r.Mouse(Mouse{Kind: MouseWheel, X: 1, Y: 0, Wheel: 1})
	if want := []string{"right leave", "left wheel"}; !reflect.DeepEqual(log, want) {
		t.Errorf("the pointer moved to left, which does not use the wheel: %v", log)
	}
}

func TestHoleClearsItsCellsAndSaysWhere(t *testing.T) {
	var at image.Rectangle
	h := &Hole{OnPlace: func(r image.Rectangle) { at = r }}
	_, b := screen(Col().Add(Label("top"), 1, 0).Add(h, 0, 1), 6, 4)
	if at != image.Rect(0, 1, 6, 4) {
		t.Errorf("placed at %v", at)
	}
	if !b.At(3, 2).Clear() || b.At(3, 0).Clear() {
		t.Error("the hole's cells are clear and the rest are not")
	}
	NewPainter(b).Shade(image.Rect(0, 0, 6, 4), Role(theme.RoleBackground), 0.5)
	if !b.At(3, 2).Clear() {
		t.Error("shading filled a hole")
	}
}

func TestTextRegionsCanBeClicked(t *testing.T) {
	var got []string
	tx := Label(`go ["north"]north[""] or ["south"]south[""]`)
	tx.OnRegion = func(name string, _ int) { got = append(got, name) }
	r, b := screen(tx, 30, 1)
	if s := b.Line(0); s != "go north or south" {
		t.Fatalf("painted %q", s)
	}
	for _, x := range []int{0, 4, 13, 9} {
		r.Mouse(Mouse{Kind: MousePress, X: x, Y: 0})
		r.Mouse(Mouse{Kind: MouseRelease, X: x, Y: 0})
	}
	if want := []string{"north", "south"}; !reflect.DeepEqual(got, want) {
		t.Errorf("clicked %v, want %v", got, want)
	}
}

func TestPrintCutsAtItsWidth(t *testing.T) {
	b := NewBuffer(10, 1)
	p := NewPainter(b)
	if n := p.Print(1, 0, 4, "[red]abcdef", Style{}); n != 4 {
		t.Errorf("painted %d cells, want 4", n)
	}
	p.PrintAligned(6, 0, 4, "xy", Style{}, AlignRight)
	if got := b.Line(0); got != " abcd   xy" {
		t.Errorf("got %q", got)
	}
	// A clipped painter touches nothing outside its rectangle.
	p.Clip(image.Rect(0, 0, 3, 1)).Fill(image.Rect(0, 0, 10, 1), Role(theme.RoleNegative))
	if b.At(2, 0).Bg != theme.RGBA(theme.RoleNegative) || b.At(3, 0).Bg == theme.RGBA(theme.RoleNegative) {
		t.Error("the clip did not hold")
	}
}
