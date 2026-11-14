package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// textViewLike is what the tests ask of both text views.
type textViewLike interface {
	tview.Primitive
}

// cellsOf draws p at w by h and returns each cell as glyph and style.
func cellsOf(t *testing.T, p tview.Primitive, w, h int) []string {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	must(t, sim.Init())
	defer sim.Fini()
	sim.SetSize(w, h)
	scr := theme.WrapScreen(sim)
	p.SetRect(0, 0, w, h)
	p.Draw(scr)
	scr.Show()
	cells, cw, _ := sim.GetContents()
	out := make([]string, 0, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*cw+x]
			r := ' '
			if len(c.Runes) > 0 && c.Runes[0] != 0 {
				r = c.Runes[0]
			}
			fg, bg, attr := theme.ResolveStyle(c.Style).Decompose()
			if r == ' ' {
				fg, attr = 0, 0 // a blank's ink does not show
			}
			out = append(out, fmt.Sprintf("%c fg %06x bg %06x %b", r, fg.Hex(), bg.Hex(), attr&^tcell.AttrUnderline))
		}
	}
	return out
}

func rowsOf(cells []string, w int) string {
	var sb strings.Builder
	for i, c := range cells {
		sb.WriteString(c[:strings.Index(c, " fg")])
		if i%w == w-1 {
			sb.WriteString("|\n")
		}
	}
	return sb.String()
}

func TestNativeTextDrawsAsATextViewDid(t *testing.T) {
	keepTheme(t)
	const text = "[gold]Wonders[-]\n\nThe [green]Great Library[-] holds [yellow]1,200[-] scrolls and a rather long line that has to wrap at the edge of its box.\n[gray]press [Esc[] to close[-]\nlast line"
	type opts struct {
		name                string
		dynamic, wrap, word bool
		align               int
		border, focus       bool
		title               string
		scroll              int
		w, h                int
	}
	for _, o := range []opts{
		{name: "plain", dynamic: true, wrap: true, w: 40, h: 10},
		{name: "word wrap", dynamic: true, wrap: true, word: true, w: 31, h: 12},
		{name: "no wrap centered", dynamic: true, align: tview.AlignCenter, w: 50, h: 8},
		{name: "right", dynamic: true, wrap: true, align: tview.AlignRight, w: 44, h: 9},
		{name: "border and title", dynamic: true, wrap: true, border: true, title: " [gold]Help[-] ", w: 36, h: 9},
		{name: "focused border", dynamic: true, wrap: true, border: true, focus: true, title: "Log", w: 36, h: 9},
		{name: "scrolled", dynamic: true, wrap: true, scroll: 3, w: 30, h: 5},
		{name: "scrolled past the end", dynamic: true, wrap: true, scroll: 400, w: 30, h: 5},
		{name: "tags off", wrap: true, w: 40, h: 10},
	} {
		old := tview.NewTextView().SetDynamicColors(o.dynamic).SetWrap(o.wrap).SetWordWrap(o.word).SetTextAlign(o.align)
		old.SetBorder(o.border).SetTitle(o.title)
		old.SetText(text)
		old.ScrollTo(o.scroll, 0)
		kt := newNativeText().SetDynamicColors(o.dynamic).SetWrap(o.wrap).SetWordWrap(o.word).SetTextAlign(o.align)
		kt.SetBorder(o.border).SetTitle(o.title)
		kt.SetText(text)
		kt.ScrollTo(o.scroll, 0)
		if o.focus {
			old.Focus(func(tview.Primitive) {})
			kt.Focus(func(tview.Primitive) {})
		}
		want, got := cellsOf(t, old, o.w, o.h), cellsOf(t, kt, o.w, o.h)
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: cell %d,%d is %q, the text view drew %q\nkit:\n%stview:\n%s", o.name, i%o.w, i/o.w, got[i], want[i], rowsOf(got, o.w), rowsOf(want, o.w))
				break
			}
		}
	}
}

func TestNativeTextKeysAndCalls(t *testing.T) {
	keepTheme(t)
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, fmt.Sprintf("line %02d", i))
	}
	kt := newNativeText().SetDynamicColors(true)
	changed := 0
	var done []tcell.Key
	kt.SetChangedFunc(func() { changed++ }).SetDoneFunc(func(k tcell.Key) { done = append(done, k) })
	kt.SetText(strings.Join(lines, "\n"))
	if changed != 1 || kt.GetText(false) != strings.Join(lines, "\n") {
		t.Fatalf("SetText: %d changes", changed)
	}
	top := func() string { return strings.TrimSpace(rowsOf(cellsOf(t, kt, 12, 5), 12)[:12]) }
	press := func(k tcell.Key, r rune) {
		kt.InputHandler()(tcell.NewEventKey(k, r, tcell.ModNone), func(tview.Primitive) {})
	}
	if top() != "line 00" {
		t.Fatalf("starts at %q", top())
	}
	for _, c := range []struct {
		key  tcell.Key
		r    rune
		want string
	}{
		{tcell.KeyDown, 0, "line 01"}, {tcell.KeyRune, 'j', "line 02"}, {tcell.KeyRune, 'k', "line 01"},
		{tcell.KeyPgDn, 0, "line 05"}, {tcell.KeyRune, 'G', "line 35"}, {tcell.KeyRune, 'g', "line 00"}, {tcell.KeyEnd, 0, "line 35"},
	} {
		press(c.key, c.r)
		if got := top(); got != c.want {
			t.Errorf("after %v %q the top line is %q, want %q", c.key, c.r, got, c.want)
		}
	}
	press(tcell.KeyEsc, 0)
	press(tcell.KeyEnter, 0)
	if len(done) != 2 || done[0] != tcell.KeyEsc {
		t.Errorf("Esc and Enter are done keys, got %v", done)
	}
	if row, _ := kt.GetScrollOffset(); row != 35 {
		t.Errorf("the scroll offset is %d", row)
	}
	// The wheel scrolls it, through the kit.
	kt.root.Mouse(kit.Mouse{Kind: kit.MouseWheel, X: 2, Y: 2, Wheel: -2})
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
	kt.SetBorder(true).SetTitle("T").SetTitleColor(tcell.ColorRed).SetDrawFunc(func(_ tcell.Screen, _, _, ww, _ int) (int, int, int, int) {
		w = ww
		return 0, 0, 0, 0
	})
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
