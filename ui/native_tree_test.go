package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// treePage is a small page on nativeFlex: a title, a framed body with a list
// beside a text, a filler row and an input line.
type treePage struct {
	page  *nativeFlex
	list  *nativeList
	text  *nativeText
	input *nativeInput
}

func newTreePage() *treePage {
	p := &treePage{list: newNativeList(), text: newNativeText(), input: newNativeInput()}
	for _, s := range []string{"Ashford", "Birch Hollow", "Cinder Reach", "Dunmere"} {
		p.list.AddItem(s, "", 0, nil)
	}
	p.list.SetBorder(true).SetTitle(" Towns ")
	p.text.SetDynamicColors(true).SetText(strings.Repeat("[yellow]one[-] two three four five six seven\n", 30))
	p.text.SetBorder(true)
	p.input.SetLabel("> ").SetText("build hut")
	title := surfaceText("A page", kit.AlignCenter)
	body := newNativeFlex().AddItem(p.list, 0, 1, true).AddItem(p.text, 0, 2, false)
	body.SetBorder(true).SetTitle(" Body ").SetTitleColor(theme.Of(theme.RoleAccent)).SetBorderColor(theme.Of(theme.RoleBorder)).
		SetBackgroundColor(theme.Of(theme.RoleNegative))
	p.page = newNativeFlex().SetDirection(flexRow).
		AddItem(title, 1, 0, false).AddItem(body, 0, 1, true).AddItem(newNativeBox(), 1, 0, false).AddItem(p.input, 1, 0, false)
	return p
}

// A flex shares its room out: fixed items get their cells, the rest is split
// by proportion with each share rounded down in turn, a gap takes room and
// shows nothing, and a border takes a cell on every side.
func TestNativeFlexSharesItsRoom(t *testing.T) {
	a, b, c := newNativeBox(), newNativeBox(), newNativeBox()
	rect := func(p Primitive) [4]int {
		x, y, w, h := p.GetRect()
		return [4]int{x, y, w, h}
	}
	row := newNativeFlex().AddItem(a, 3, 0, false).AddItem(nil, 0, 1, false).AddItem(b, 0, 2, false).AddItem(c, 5, 0, false)
	row.SetRect(3, 2, 40, 12)
	drawOn(newSimScreen(t, 60, 20), row)
	// 40 - 3 - 5 = 32 to share 1:2: the gap takes 10, b the 22 left.
	if got, want := [3][4]int{rect(a), rect(b), rect(c)}, [3][4]int{{3, 2, 3, 12}, {16, 2, 22, 12}, {38, 2, 5, 12}}; got != want {
		t.Errorf("a row of 40: items at %v, want %v", got, want)
	}
	col := newNativeFlex().SetDirection(flexRow).AddItem(a, 0, 1, false).AddItem(b, 2, 0, false)
	col.SetBorder(true)
	col.SetRect(0, 0, 10, 9)
	drawOn(newSimScreen(t, 60, 20), col)
	if got, want := [2][4]int{rect(a), rect(b)}, [2][4]int{{1, 1, 8, 5}, {1, 6, 8, 2}}; got != want {
		t.Errorf("a framed column of 9 rows: items at %v, want %v", got, want)
	}
	if x, y, w, h := col.GetInnerRect(); [4]int{x, y, w, h} != [4]int{1, 1, 8, 7} {
		t.Errorf("the framed column's inside is %v", [4]int{x, y, w, h})
	}
}

// A page shows its children where the flexes put them, the body's border on
// the body's own ground and doubled while the keyboard is inside.
func TestAFlexPageDrawsItsChildren(t *testing.T) {
	keepTheme(t)
	p := newTreePage()
	f := drawWrapped(t, p.page, 50, 14, 2)
	rows := f.Rows()
	if !strings.Contains(rows[0], "A page") || !strings.Contains(rows[1], " Body ") || !strings.Contains(rows[2], " Towns ") {
		t.Fatalf("the page is not laid out as built:\n%s", f.Text())
	}
	if !strings.HasPrefix(rows[13], "> build hut") || !strings.Contains(rows[3], "Ashford") || !strings.Contains(rows[3], "one two three") {
		t.Errorf("a child is not where its flex put it:\n%s", f.Text())
	}
	if f.Rune(0, 1) != '┌' {
		t.Errorf("with the keyboard elsewhere the body's border is %q", f.Rune(0, 1))
	}
	if _, bg, _ := decompose(f.cells[1*f.w].Style); bg != theme.Of(theme.RoleNegative) {
		t.Errorf("the body's border is drawn on %06x, want the ground the body was given", bg.Hex())
	}
	p.list.Focus(nil)
	if f = drawWrapped(t, p.page, 50, 14, 2); f.Rune(0, 1) != '╔' || f.Rune(1, 2) != '╔' {
		t.Errorf("with the keyboard on the list the borders are %q and %q, want both doubled", f.Rune(0, 1), f.Rune(1, 2))
	}
}

// The focus and the keys go where the flex sends them: it hands the focus to
// its focus item, a capture on the page sees a key before one on the child,
// and either can stop it.
func TestAFlexPageRoutesFocusAndKeys(t *testing.T) {
	keepTheme(t)
	p := newTreePage()
	app := newAppLoop()
	page := p.page
	app.SetFocus(page)
	if app.GetFocus() != Primitive(p.list) || !page.HasFocus() || !p.list.HasFocus() || p.input.HasFocus() {
		t.Fatalf("the page's focus item should have the focus, it is on %T", app.GetFocus())
	}
	drawWrapped(t, page, 50, 14, 1)
	var seen []string
	stop := ""
	page.SetInputCapture(func(k *kit.Key) *kit.Key {
		seen = append(seen, "page")
		if stop == "page" {
			return nil
		}
		return k
	})
	p.list.SetInputCapture(func(k *kit.Key) *kit.Key {
		seen = append(seen, "list")
		if stop == "list" {
			return nil
		}
		return k
	})
	press := func(code kit.KeyCode, r rune) { keyTo(page, tk(code, r, 0)) }

	press(kit.KeyDown, 0)
	if got := strings.Join(seen, " "); got != "page list" || p.list.GetCurrentItem() != 1 {
		t.Errorf("Down: captures ran %q and the list is on row %d; want page then list, row 1", got, p.list.GetCurrentItem())
	}
	seen, stop = nil, "list"
	press(kit.KeyDown, 0)
	if got := strings.Join(seen, " "); got != "page list" || p.list.GetCurrentItem() != 1 {
		t.Errorf("a key the list's capture took still moved it: %q, row %d", got, p.list.GetCurrentItem())
	}
	seen, stop = nil, "page"
	press(kit.KeyDown, 0)
	if got := strings.Join(seen, " "); got != "page" || p.list.GetCurrentItem() != 1 {
		t.Errorf("a key the page's capture took went on: %q, row %d", got, p.list.GetCurrentItem())
	}

	// The focus moved to the input line: the same page now types into it.
	seen, stop = nil, ""
	app.SetFocus(p.input)
	press(kit.KeyRune, '!')
	if got := strings.Join(seen, " "); got != "page" || p.input.GetText() != "build hut!" || p.list.GetCurrentItem() != 1 {
		t.Errorf("typing with the input focused: captures %q, input %q, list row %d", got, p.input.GetText(), p.list.GetCurrentItem())
	}

	// A flex with no focus item keeps the focus itself, and its capture
	// still hears the keys.
	bare := newNativeFlex().AddItem(newNativeBox(), 0, 1, false)
	heard := 0
	bare.SetInputCapture(func(k *kit.Key) *kit.Key { heard++; return k })
	app.SetFocus(bare)
	keyTo(bare, tk(kit.KeyEnter, 0, 0))
	if app.GetFocus() != Primitive(bare) || heard != 1 {
		t.Errorf("a flex with no focus item: focus on %T, capture heard %d keys", app.GetFocus(), heard)
	}
}

// The text cursor shows on the page's screen while the input line inside it
// has the focus, and not while something else has.
func TestAFlexPageShowsTheCursorOfItsInput(t *testing.T) {
	keepTheme(t)
	sim := newSimScreen(t, 50, 14)
	p := newTreePage()
	app := newAppLoop()
	draw := func() (int, int, bool) {
		sim.HideCursor()
		p.page.SetRect(0, 0, 50, 14)
		drawOn(sim, p.page)
		return sim.GetCursor()
	}
	app.SetFocus(p.input)
	x, y, on := draw()
	ix, iy, _, _ := p.input.GetInnerRect()
	if want := ix + len("> build hut"); !on || y != iy || x != want {
		t.Errorf("with the input focused the cursor is at %d,%d (shown %v), want %d,%d", x, y, on, want, iy)
	}
	app.SetFocus(p.list)
	if _, _, on := draw(); on {
		t.Error("the cursor shows though the list has the focus")
	}
}

// The mouse reaches a child through the page's one tree: a click picks a
// list row and the wheel scrolls the text.
func TestAFlexPageTakesTheMouse(t *testing.T) {
	keepTheme(t)
	p := newTreePage()
	st := newStage(p.page, 50, 14)
	lx, ly, _, _ := p.list.GetInnerRect()
	st.click(lx+2, ly+2)
	if p.list.GetCurrentItem() != 2 {
		t.Errorf("a click on the third row left the list on row %d", p.list.GetCurrentItem())
	}
	st.paint()
	tx, ty, _, _ := p.text.GetInnerRect()
	st.mouse(kit.Mouse{Kind: kit.MouseWheel, X: tx + 3, Y: ty + 1, Wheel: 4})
	st.paint()
	if row, _ := p.text.GetScrollOffset(); row <= 0 {
		t.Errorf("the wheel over the text left it at row %d", row)
	}
}

// Where a window's margin is a gap the page beneath still shows, and a
// filler box clears it. Either way the pointer is the window's: the page
// beneath gets nothing while the window is up.
func TestAWindowShowsThePageBeneathAndKeepsThePointer(t *testing.T) {
	keepTheme(t)
	for _, clear := range []bool{false, true} {
		under := newNativeText().SetText(strings.Repeat(strings.Repeat("u", 30)+"\n", 40))
		window := centerModal(solidBox(theme.RoleSurface), 10, 3, clear)
		pages := newNativePages()
		pages.AddPage("under", under, true, true)
		pages.AddPage("window", window, true, true)
		st := newStage(pages, 30, 9)
		margin, middle := st.buf.At(2, 1).R, st.buf.At(15, 4).R
		if middle == 'u' {
			t.Errorf("clear %v: the window did not draw over the page", clear)
		}
		if got := margin == 'u'; got == clear {
			t.Errorf("clear %v: the margin shows %q", clear, margin)
		}
		// The wheel over the margin never scrolls the page beneath.
		before, _ := under.GetScrollOffset()
		st.mouse(kit.Mouse{Kind: kit.MouseWheel, X: 2, Y: 1, Wheel: 1})
		if after, _ := under.GetScrollOffset(); after != before {
			t.Errorf("clear %v: the wheel over the margin moved the page beneath from row %d to %d", clear, before, after)
		}
	}
}
