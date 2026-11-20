package kit

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

func bigInput() (*Input, *Root, *Buffer) {
	in := NewInput("> ")
	in.Big = true
	r, b := screen(in, 24, 2)
	return in, r, b
}

// A big line reads by Lines() like a small one: the text on the band's first
// row, once, with the pad before the prompt.
func TestBigInputShowsTheTextOnTheFirstRow(t *testing.T) {
	in, r, b := bigInput()
	typeText(r, "build")
	r.Paint(b)
	if got := b.Line(0); got != "  > build" {
		t.Errorf("first row %q", got)
	}
	if got := b.Line(1); got != "" {
		t.Errorf("the second row holds only the lower halves, read as %q", got)
	}
	if c := b.At(6, 0); c.Big != 1 || c.R != 'b' {
		t.Errorf("the first letter is the group's top-left anchor: %+v", c)
	}
	if c := b.At(7, 1); c.Big != 4 || !c.Cont {
		t.Errorf("and its lower-right cell continues it: %+v", c)
	}
	if !b.CursorOn || !b.CursorBig || b.Cursor.X != 6+2*5 {
		t.Errorf("the cursor is big and after the text: %v big %v", b.Cursor, b.CursorBig)
	}
	if in.Text() != "build" {
		t.Errorf("text %q", in.Text())
	}
}

// Writing over a quarter of a group blanks the whole group.
func TestOverwritingAQuarterOfABigGroupBlanksIt(t *testing.T) {
	b := NewBuffer(8, 4)
	p := NewPainter(b)
	p.PrintBig(0, 0, 8, "ab", Style{})
	p.Put(1, 1, 'z', Style{})
	if c := b.At(0, 0); c.R != ' ' || c.Big != 0 || c.Cont {
		t.Errorf("the group's anchor is left over: %+v", c)
	}
	for _, at := range [][2]int{{1, 0}, {0, 1}} {
		if c := b.At(at[0], at[1]); c.Big != 0 || c.Cont || c.R != ' ' {
			t.Errorf("cell %v of the group is left over: %+v", at, c)
		}
	}
	if c := b.At(2, 0); c.R != 'b' || c.Big != 1 {
		t.Errorf("the next group is untouched: %+v", c)
	}
	if got := b.Line(1); got != " z" {
		t.Errorf("row 1 reads %q", got)
	}
}

// A click on either cell of a big column puts the caret before or after it.
func TestBigInputTakesAClickOnABigColumn(t *testing.T) {
	in, r, _ := bigInput()
	typeText(r, "abcd")
	// text starts at cell 6; column k covers cells 6+2k and 7+2k
	r.Mouse(Mouse{Kind: MousePress, X: 6 + 2*2, Y: 0})
	typeText(r, "X")
	if got := in.Text(); got != "abXcd" {
		t.Errorf("a click on the third column: %q", got)
	}
	r.Mouse(Mouse{Kind: MousePress, X: 6, Y: 1})
	typeText(r, "Y")
	if got := in.Text(); got != "YabXcd" {
		t.Errorf("a click on the lower half of the first column: %q", got)
	}
}

// The ghost follows the text at double size, dim, and goes when the cursor
// leaves the end.
func TestBigInputShowsAGhostAfterTheText(t *testing.T) {
	in := NewInput("> ")
	in.Big = true
	in.Ghost = func(s string) string {
		if s == "bu" {
			return "ild"
		}
		return ""
	}
	r, b := screen(in, 24, 2)
	typeText(r, "bu")
	r.Paint(b)
	if got := b.Line(0); got != "  > build" {
		t.Errorf("painted %q", got)
	}
	if c := b.At(10, 0); c.R != 'i' || c.Big != 1 || c.Fg != theme.RGBA(theme.RoleDim) {
		t.Errorf("the ghost is a dim big group after the text: %+v", c)
	}
	if b.At(8, 0).Fg == theme.RGBA(theme.RoleDim) {
		t.Error("the typed text is not dim")
	}
	r.Key(key(KeyLeft))
	r.Paint(b)
	if got := b.Line(0); got != "  > bu" {
		t.Errorf("with the cursor inside the line there is no ghost: %q", got)
	}
}

// Centred in a taller rect, and the size follows Big.
func TestBigInputSizeAndCentring(t *testing.T) {
	in := NewInput("> ")
	if got := in.Size(image.Pt(30, 5)); got.Y != 1 {
		t.Errorf("small size %v", got)
	}
	in.Big = true
	if got := in.Size(image.Pt(30, 5)); got.Y != 2 {
		t.Errorf("big size %v", got)
	}
	r, b := screen(in, 24, 4)
	typeText(r, "a")
	r.Paint(b)
	if got := b.Line(1); got != "  > a" {
		t.Errorf("a two-row line in four rows sits on row 1: %q", b.Lines())
	}
}
