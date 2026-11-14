package kit

import (
	"image"
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
		"┌───Top────┐",
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
	tx.OnRegion = func(name string) { got = append(got, name) }
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
