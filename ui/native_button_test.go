package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

func TestNativeButtonDrawsAsAButtonDid(t *testing.T) {
	keepTheme(t)
	for _, c := range []struct {
		name, label string
		fill        theme.Role
		colored     bool
		focus       bool
		w           int
	}{
		{name: "plain", label: "Close", w: 9},
		{name: "plain with the focus", label: "Close", focus: true, w: 9},
		{name: "wider than its label", label: "OK", w: 14},
		{name: "escaped tag", label: tview.Escape("[E] ENDURE"), colored: true, fill: theme.RoleNegative, w: 14},
		{name: "colored with the focus", label: "[ OK ]", colored: true, fill: theme.RoleAccent, focus: true, w: 8},
		{name: "chip", label: "Close", colored: true, fill: theme.RoleChip, w: 9},
		{name: "chip with the focus", label: "Close", colored: true, fill: theme.RoleChip, focus: true, w: 9},
	} {
		old, nb := tview.NewButton(c.label), newNativeButton(c.label)
		if c.colored {
			old.SetBackgroundColor(theme.Color(c.fill))
			old.SetLabelColor(onFill(c.fill))
			styleFilledButton(nb, c.fill)
		}
		if c.focus {
			old.Focus(func(tview.Primitive) {})
			nb.Focus(func(tview.Primitive) {})
		}
		want, got := cellsOf(t, old, c.w, 1), cellsOf(t, nb, c.w, 1)
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: cell %d is %q, the button drew %q\nkit:\n%stview:\n%s", c.name, i, got[i], want[i], rowsOf(got, c.w), rowsOf(want, c.w))
				break
			}
		}
	}
}

func TestNativeButtonPressedByEnterAndAClick(t *testing.T) {
	keepTheme(t)
	n := 0
	b := newNativeButton("Go").SetSelectedFunc(func() { n++ })
	key := func(k tcell.Key, r rune) {
		b.InputHandler()(tcell.NewEventKey(k, r, tcell.ModNone), func(tview.Primitive) {})
	}
	key(tcell.KeyRune, ' ')
	key(tcell.KeyEsc, 0)
	key(tcell.KeyTab, 0)
	if n != 0 {
		t.Errorf("only Enter presses a tview button, got %d presses", n)
	}
	key(tcell.KeyEnter, 0)
	if n != 1 {
		t.Errorf("Enter pressed %d times", n)
	}
	cellsOf(t, b, 8, 1)
	b.root.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 2, Y: 0})
	b.root.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: 2, Y: 0})
	if n != 2 {
		t.Errorf("a click pressed it %d times in all, want 2", n)
	}
}
