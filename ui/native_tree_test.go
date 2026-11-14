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

// flexSpec is one item of a flex under test: a gap when gap is set.
type flexSpec struct {
	fixed, prop int
	gap         bool
}

// A nativeFlex gives each item the rectangle tview's Flex gave it, whatever
// the mix of fixed and shared items, gaps, a border, or too little room.
func TestNativeFlexSharesRoomAsTviewDoes(t *testing.T) {
	cases := map[string][]flexSpec{
		"shared":           {{prop: 1}, {prop: 2}, {prop: 1}},
		"fixed and shared": {{fixed: 3}, {prop: 1}, {fixed: 5}},
		"gaps":             {{prop: 1, gap: true}, {fixed: 7}, {prop: 1, gap: true}},
		"uneven":           {{prop: 3}, {prop: 5}, {prop: 7}, {fixed: 1}},
		"nothing shared":   {{fixed: 4}, {fixed: 4}},
		"no proportion":    {{fixed: 2}, {}, {prop: 1}},
		"more than fits":   {{fixed: 30}, {fixed: 30}, {prop: 1}},
	}
	for name, specs := range cases {
		for _, down := range []bool{false, true} {
			for _, border := range []bool{false, true} {
				for _, size := range [][2]int{{40, 12}, {7, 3}, {1, 1}, {0, 0}} {
					where := fmt.Sprintf("%s, down %v, border %v, %dx%d", name, down, border, size[0], size[1])
					old, now := tview.NewFlex(), newNativeFlex()
					if down {
						old.SetDirection(tview.FlexRow)
						now.SetDirection(flexRow)
					}
					old.SetBorder(border)
					now.SetBorder(border)
					var olds []*tview.Box
					var nows []*nativeBox
					for _, s := range specs {
						if s.gap {
							old.AddItem(nil, s.fixed, s.prop, false)
							now.AddItem(nil, s.fixed, s.prop, false)
							continue
						}
						a, b := tview.NewBox(), newNativeBox()
						old.AddItem(a, s.fixed, s.prop, false)
						now.AddItem(b, s.fixed, s.prop, false)
						olds, nows = append(olds, a), append(nows, b)
					}
					sim := newSimScreen(t, 60, 20)
					old.SetRect(3, 2, size[0], size[1])
					now.SetRect(3, 2, size[0], size[1])
					old.Draw(sim)
					now.Draw(sim)
					for i := range olds {
						ax, ay, aw, ah := olds[i].GetRect()
						bx, by, bw, bh := nows[i].GetRect()
						if [4]int{ax, ay, aw, ah} != [4]int{bx, by, bw, bh} {
							t.Errorf("%s: item %d is at %v, tview put it at %v", where, i, [4]int{bx, by, bw, bh}, [4]int{ax, ay, aw, ah})
						}
					}
				}
			}
		}
	}
}

// A tvShim puts one of the game's primitives in a tview layout, for the
// tests that compare a page on nativeFlex with the same page on tview's Flex.
type tvShim struct {
	*tview.Box
	p Primitive
}

func shim(p Primitive) *tvShim { return &tvShim{Box: tview.NewBox(), p: p} }

func (s *tvShim) Draw(scr tcell.Screen) {
	s.p.SetRect(s.GetRect())
	s.p.Draw(scr)
}

func (s *tvShim) HasFocus() bool { return s.p.HasFocus() }

// treePage is a small page built twice over the same kind of children: on
// tview's Flex, each child drawing itself, and on nativeFlex, one tree.
type treePage struct {
	page  drawable
	list  *nativeList
	text  *nativeText
	input *nativeInput
}

func newTreePage(native bool) *treePage {
	p := &treePage{list: newNativeList(), text: newNativeText(), input: newNativeInput()}
	for _, s := range []string{"Ashford", "Birch Hollow", "Cinder Reach", "Dunmere"} {
		p.list.AddItem(s, "", 0, nil)
	}
	p.list.SetBorder(true).SetTitle(" Towns ")
	p.text.SetDynamicColors(true).SetText(strings.Repeat("[yellow]one[-] two three four five six seven\n", 30))
	p.text.SetBorder(true)
	p.input.SetLabel("> ").SetText("build hut")
	title := surfaceText("A page", kit.AlignCenter)
	if native {
		body := newNativeFlex().AddItem(p.list, 0, 1, true).AddItem(p.text, 0, 2, false)
		body.SetBorder(true).SetTitle(" Body ").SetTitleColor(theme.Color(theme.RoleAccent)).SetBorderColor(theme.Color(theme.RoleBorder)).
			SetBackgroundColor(theme.Color(theme.RoleNegative))
		p.page = newNativeFlex().SetDirection(flexRow).
			AddItem(title, 1, 0, false).AddItem(body, 0, 1, true).AddItem(newNativeBox(), 1, 0, false).AddItem(p.input, 1, 0, false)
		return p
	}
	body := tview.NewFlex().AddItem(shim(p.list), 0, 1, true).AddItem(shim(p.text), 0, 2, false)
	body.SetBorder(true).SetTitle(" Body ").SetTitleColor(theme.Color(theme.RoleAccent)).SetBorderColor(theme.Color(theme.RoleBorder)).
		SetBackgroundColor(theme.Color(theme.RoleNegative))
	p.page = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(shim(title), 1, 0, false).AddItem(body, 0, 1, true).AddItem(tview.NewBox(), 1, 0, false).AddItem(shim(p.input), 1, 0, false)
	return p
}

// A page on nativeFlex shows the cells the same page showed on tview's Flex,
// with the focus nowhere, on the list and on the input line.
func TestAFlexPageDrawsAsItDidOnTview(t *testing.T) {
	keepTheme(t)
	for _, focus := range []string{"none", "list", "input"} {
		var frames [2]frame
		for i, native := range []bool{false, true} {
			p := newTreePage(native)
			switch focus {
			case "list":
				p.list.Focus(nil)
			case "input":
				p.input.Focus(nil)
			}
			frames[i] = drawWrapped(t, p.page, 50, 14, 2)
		}
		old, now := frames[0], frames[1]
		for i := range old.cells {
			a, b := old.cells[i], now.cells[i]
			if string(a.Runes) != string(b.Runes) || a.Style != b.Style {
				t.Fatalf("focus %s: cell %d,%d is %q %v, on tview it was %q %v\nnow:\n%s\ntview:\n%s",
					focus, i%old.w, i/old.w, string(b.Runes), b.Style, string(a.Runes), a.Style, now.Text(), old.Text())
			}
		}
	}
}

// The focus and the keys go where tview's Flex sent them: the flex hands the
// focus to its focus item, a capture on the page sees a key before one on
// the child, and either can stop it.
func TestAFlexPageRoutesFocusAndKeys(t *testing.T) {
	keepTheme(t)
	p := newTreePage(true)
	app := newAppLoop()
	page := p.page.(*nativeFlex)
	app.SetFocus(page)
	if app.GetFocus() != p.list || !page.HasFocus() || !p.list.HasFocus() || p.input.HasFocus() {
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
	press := func(code kit.KeyCode, r rune) { page.InputHandler()(tev(code, r, 0), func(Primitive) {}) }

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
	bare.InputHandler()(tev(kit.KeyEnter, 0, 0), func(Primitive) {})
	if app.GetFocus() != bare || heard != 1 {
		t.Errorf("a flex with no focus item: focus on %T, capture heard %d keys", app.GetFocus(), heard)
	}
}

// The text cursor shows on the page's screen while the input line inside it
// has the focus, and not while something else has.
func TestAFlexPageShowsTheCursorOfItsInput(t *testing.T) {
	keepTheme(t)
	sim := newSimScreen(t, 50, 14)
	p := newTreePage(true)
	app := newAppLoop()
	draw := func() (int, int, bool) {
		sim.HideCursor()
		p.page.SetRect(0, 0, 50, 14)
		p.page.Draw(sim)
		sim.Show()
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
	sim := newSimScreen(t, 50, 14)
	scr := newNativeScreen(theme.WrapScreen(sim))
	p := newTreePage(true)
	draw := func() {
		p.page.SetRect(0, 0, 50, 14)
		p.page.Draw(scr)
		scr.Show()
	}
	draw()
	lx, ly, _, _ := p.list.GetInnerRect()
	scr.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: lx + 2, Y: ly + 2})
	scr.mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: lx + 2, Y: ly + 2})
	if p.list.GetCurrentItem() != 2 {
		t.Errorf("a click on the third row left the list on row %d", p.list.GetCurrentItem())
	}
	draw()
	tx, ty, _, _ := p.text.GetInnerRect()
	scr.mouse(kit.Mouse{Kind: kit.MouseWheel, X: tx + 3, Y: ty + 1, Wheel: 4})
	draw()
	if row, _ := p.text.GetScrollOffset(); row <= 0 {
		t.Errorf("the wheel over the text left it at row %d", row)
	}
}

// Where a flex has a gap the page beneath still shows, and its cells are
// still that page's for the mouse; a filler box clears them and takes them.
func TestAFlexLeavesItsGapsToThePageBeneath(t *testing.T) {
	keepTheme(t)
	for _, clear := range []bool{false, true} {
		sim := newSimScreen(t, 30, 9)
		scr := newNativeScreen(theme.WrapScreen(sim))
		under := newSpot('u', false)
		lower := newNativeHost(under)
		lower.SetRect(0, 0, 30, 9)
		window := centerModal(solidBox(theme.RoleSurface), 10, 3, clear)
		window.SetRect(0, 0, 30, 9)
		lower.Draw(scr)
		window.Draw(scr)
		scr.Show()
		f := simCells(sim)
		margin, middle := f.Rune(2, 1), f.Rune(15, 4)
		if middle == 'u' {
			t.Errorf("clear %v: the window did not draw over the page", clear)
		}
		if got := margin == 'u'; got == clear {
			t.Errorf("clear %v: the margin shows %q", clear, margin)
		}
		if got := scr.hostAt(2, 1) == lower; got == clear {
			t.Errorf("clear %v: the margin's cells belong to the page beneath: %v", clear, got)
		}
		if scr.hostAt(15, 4) != window.nativeHost {
			t.Errorf("clear %v: the window's own cells are not its own", clear)
		}
	}
}
