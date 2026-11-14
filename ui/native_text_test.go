package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// cellsOf draws p alone at w by h and returns each cell as "rune fg hex", row
// by row: enough to say what is where and in what ink.
func cellsOf(_ *testing.T, p Primitive, w, h int) []string {
	f := drawPrimitive(nil, p, w, h)
	out := make([]string, 0, w*h)
	for _, c := range f.cells {
		fg, _, _ := decompose(c.Style)
		out = append(out, fmt.Sprintf("%c fg %06x", c.Runes[0], fg.Hex()))
	}
	return out
}

func TestNativeTextKeysAndCalls(t *testing.T) {
	keepTheme(t)
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, fmt.Sprintf("line %02d", i))
	}
	kt := newNativeText().SetDynamicColors(true)
	changed := 0
	var done []kit.KeyCode
	kt.SetChangedFunc(func() { changed++ }).SetDoneFunc(func(k kit.KeyCode) { done = append(done, k) })
	kt.SetText(strings.Join(lines, "\n"))
	if changed != 1 || kt.GetText(false) != strings.Join(lines, "\n") {
		t.Fatalf("SetText: %d changes", changed)
	}
	top := func() string { return strings.TrimSpace(drawPrimitive(t, kt, 12, 5).Rows()[0]) }
	press := func(k kit.KeyCode, r rune) {
		keyTo(kt, tk(k, r, 0))
	}
	if top() != "line 00" {
		t.Fatalf("starts at %q", top())
	}
	for _, c := range []struct {
		key  kit.KeyCode
		r    rune
		want string
	}{
		{kit.KeyDown, 0, "line 01"}, {kit.KeyRune, 'j', "line 02"}, {kit.KeyRune, 'k', "line 01"},
		{kit.KeyPgDn, 0, "line 05"}, {kit.KeyRune, 'G', "line 35"}, {kit.KeyRune, 'g', "line 00"}, {kit.KeyEnd, 0, "line 35"},
	} {
		press(c.key, c.r)
		if got := top(); got != c.want {
			t.Errorf("after %v %q the top line is %q, want %q", c.key, c.r, got, c.want)
		}
	}
	press(kit.KeyEsc, 0)
	press(kit.KeyEnter, 0)
	if len(done) != 2 || done[0] != kit.KeyEsc {
		t.Errorf("Esc and Enter are done keys, got %v", done)
	}
	if row, _ := kt.GetScrollOffset(); row != 35 {
		t.Errorf("the scroll offset is %d", row)
	}
	// The wheel scrolls it, through the kit.
	newStage(kt, 12, 5).mouse(kit.Mouse{Kind: kit.MouseWheel, X: 2, Y: 2, Wheel: -2})
	if got := top(); got != "line 29" {
		t.Errorf("two notches up: %q", got)
	}
	kt.SetScrollable(false).SetText("a\nb\nc\nd\ne\nf\ng\nh")
	if got := top(); got != "d" {
		t.Errorf("not scrollable shows the end, top line %q", got)
	}
	// The older lines go after a draw, as in tview: the text is whole once.
	kt.SetMaxLines(2)
	if got := top(); got != "d" || kt.GetText(false) != "g\nh" {
		t.Errorf("SetMaxLines(2): drawn top line %q, text afterward %q", got, kt.GetText(false))
	}
	if got := kt.Clear().GetText(true); got != "" || changed != 3 {
		t.Errorf("Clear: %q, %d changes", got, changed)
	}
}

func TestNativeTextWriteTitleColorAndDrawFunc(t *testing.T) {
	keepTheme(t)
	kt := newNativeText()
	if _, err := fmt.Fprint(kt, "one "); err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(kt, "two")
	if got := kt.GetText(false); got != "one two" {
		t.Errorf("Write: %q", got)
	}
	var w int
	kt.SetBorder(true).SetTitle("T").SetTitleColor(kit.RGB(255, 0, 0)).SetDrawFunc(func(_, _, ww, _ int) { w = ww })
	if kt.GetTitle() != "T" {
		t.Errorf("GetTitle: %q", kt.GetTitle())
	}
	cells := cellsOf(t, kt, 20, 4)
	if w != 20 {
		t.Errorf("the draw func saw width %d", w)
	}
	if !strings.HasPrefix(cells[10], "T fg ff0000") {
		t.Errorf("the title is %q, want red T", cells[10])
	}
}

// TestPanelFrameIsOneTree: the frame and its text are one kit tree, so the
// wheel reaches the text through the frame.
func TestPanelFrameIsOneTree(t *testing.T) {
	keepTheme(t)
	var lines []string
	for i := 0; i < 60; i++ {
		lines = append(lines, fmt.Sprintf("row %02d", i))
	}
	tv := newNativeText().SetDynamicColors(true)
	tv.SetText(strings.Join(lines, "\n"))
	f := newPanelFrame(txtExpeditions, tv)
	st := newStage(f, 40, 12)
	first := func() string {
		st.paint()
		return string([]rune(st.buf.Line(1))[panelPadX : panelPadX+6])
	}
	if got := first(); got != "row 00" {
		t.Fatalf("the text starts at %q", got)
	}
	if !st.mouse(kit.Mouse{Kind: kit.MouseWheel, X: 5, Y: 5, Wheel: 2}) {
		t.Fatal("the wheel over the text changed nothing")
	}
	if got := first(); got == "row 00" {
		t.Errorf("after the wheel the text still starts at %q", got)
	}
	if x, y, w, h := tv.GetRect(); w <= 0 || h <= 0 || x != panelPadX || y != 1 {
		t.Errorf("the text view was told it is at %d,%d %dx%d", x, y, w, h)
	}
}
