package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

var listRows = []string{"[gold]● Ashford[-]  [gray]a1b2[-]", "Birch Hollow", "[dim]── Group ──[-]", "Cinder Reach", "Dunmere", "Eastwatch"}

func listPair() (*tview.List, *nativeList) {
	old, nl := tview.NewList().ShowSecondaryText(false), newNativeList().ShowSecondaryText(false)
	for _, r := range listRows {
		old.AddItem(r, "", 0, nil)
		nl.AddItem(r, "", 0, nil)
	}
	for _, l := range []interface {
		SetSelectedBackgroundColor(tcell.Color) *tview.List
	}{old} {
		l.SetSelectedBackgroundColor(theme.Ref(theme.RoleSelection)).SetSelectedTextColor(theme.Ref(theme.RoleSelectionText))
	}
	nl.SetSelectedBackgroundColor(theme.Ref(theme.RoleSelection)).SetSelectedTextColor(theme.Ref(theme.RoleSelectionText))
	return old, nl
}

func TestNativeListDrawsAsAListDid(t *testing.T) {
	keepTheme(t)
	for _, c := range []struct {
		name          string
		sel           int
		border, focus bool
		full          bool
		w, h          int
	}{
		{name: "first selected", w: 30, h: 8},
		{name: "middle selected", sel: 3, w: 30, h: 8},
		{name: "scrolled to the last", sel: 5, w: 30, h: 3},
		{name: "bordered with a title", sel: 1, border: true, w: 30, h: 8},
		{name: "bordered with the focus", sel: 1, border: true, focus: true, w: 30, h: 8},
		{name: "full line", sel: 2, full: true, w: 30, h: 8},
		{name: "narrow", sel: 0, w: 8, h: 4},
	} {
		old, nl := listPair()
		old.SetCurrentItem(c.sel)
		nl.SetCurrentItem(c.sel)
		old.SetBorder(c.border).SetTitle(" Themes ")
		nl.SetBorder(c.border).SetTitle(" Themes ")
		old.SetHighlightFullLine(c.full)
		nl.SetHighlightFullLine(c.full)
		if c.focus {
			old.Focus(func(tview.Primitive) {})
			nl.Focus(func(tview.Primitive) {})
		}
		want, got := cellsOf(t, old, c.w, c.h), cellsOf(t, nl, c.w, c.h)
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: cell %d,%d is %q, the list drew %q\nkit:\n%stview:\n%s", c.name, i%c.w, i/c.w, got[i], want[i], rowsOf(got, c.w), rowsOf(want, c.w))
				break
			}
		}
	}
}

func TestNativeListKeys(t *testing.T) {
	keepTheme(t)
	l := newNativeList().ShowSecondaryText(false)
	var moved, chosen []int
	l.SetChangedFunc(func(i int, _, _ string, _ rune) { moved = append(moved, i) })
	l.SetSelectedFunc(func(i int, _, _ string, _ rune) { chosen = append(chosen, i) })
	for i, r := range []rune{'a', 'b', 0, 'd', 0} {
		l.AddItem("row", "", r, nil)
		_ = i
	}
	if len(moved) != 1 || moved[0] != 0 {
		t.Errorf("the first row added announces itself: %v", moved)
	}
	cellsOf(t, l, 20, 3)
	press := func(k tcell.Key, r rune) {
		l.InputHandler()(tcell.NewEventKey(k, r, tcell.ModNone), func(tview.Primitive) {})
	}
	at := func(want int, what string) {
		t.Helper()
		if got := l.GetCurrentItem(); got != want {
			t.Errorf("%s: row %d, want %d", what, got, want)
		}
	}
	press(tcell.KeyDown, 0)
	at(1, "Down")
	press(tcell.KeyEnd, 0)
	at(4, "End")
	press(tcell.KeyDown, 0)
	at(0, "Down at the end comes round")
	press(tcell.KeyUp, 0)
	at(4, "Up at the top comes round")
	press(tcell.KeyHome, 0)
	at(0, "Home")
	press(tcell.KeyPgDn, 0)
	at(3, "PgDn is a screenful (3 rows)")
	press(tcell.KeyPgUp, 0)
	at(0, "PgUp")
	press(tcell.KeyTab, 0)
	at(1, "Tab")
	press(tcell.KeyEnter, 0)
	if len(chosen) != 1 || chosen[0] != 1 {
		t.Errorf("Enter chooses the row: %v", chosen)
	}
	press(tcell.KeyRune, 'd')
	at(3, "a shortcut selects")
	if len(chosen) != 2 || chosen[1] != 3 {
		t.Errorf("a shortcut chooses its row: %v", chosen)
	}
	press(tcell.KeyRune, 'z')
	if len(chosen) != 2 {
		t.Error("a letter that is no shortcut does nothing")
	}
	press(tcell.KeyRune, ' ')
	if len(chosen) != 3 || chosen[2] != 3 {
		t.Errorf("Space chooses the selected row: %v", chosen)
	}
	l.SetCurrentItem(-1)
	at(4, "SetCurrentItem counts back from the end")
	l.Clear()
	if l.GetItemCount() != 0 || l.GetCurrentItem() != 0 {
		t.Error("Clear empties the list")
	}
}

func TestNativeListMouse(t *testing.T) {
	keepTheme(t)
	_, l := listPair()
	var moved, chosen []int
	l.SetChangedFunc(func(i int, _, _ string, _ rune) { moved = append(moved, i) })
	l.SetSelectedFunc(func(i int, _, _ string, _ rune) { chosen = append(chosen, i) })
	cellsOf(t, l, 30, 6)
	l.root.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 3, Y: 3})
	l.root.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: 3, Y: 3})
	if l.GetCurrentItem() != 3 || len(moved) != 1 || moved[0] != 3 {
		t.Errorf("a click selects its row: at %d, moved %v", l.GetCurrentItem(), moved)
	}
	if len(chosen) != 1 || chosen[0] != 3 {
		t.Errorf("one click chooses, as it does in tview: %v", chosen)
	}
}
