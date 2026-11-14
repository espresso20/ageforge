package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

func TestNativeInputDrawsAsAnInputFieldDid(t *testing.T) {
	keepTheme(t)
	for _, c := range []struct {
		name          string
		label, text   string
		width         int
		border, focus bool
		title         string
		w, h          int
	}{
		{name: "plain", label: "Name: ", text: "Ashford", w: 30, h: 1},
		{name: "fixed width", label: "Name: ", text: "Ashford", width: 12, w: 30, h: 1},
		{name: "bordered", label: "❯ ", text: "build hut", border: true, title: " Command ", w: 40, h: 3},
		{name: "bordered with the focus", label: "❯ ", text: "build hut", border: true, focus: true, w: 40, h: 3},
		{name: "empty", label: "> ", w: 20, h: 1},
	} {
		old := tview.NewInputField().SetLabel(c.label).SetFieldWidth(c.width).SetText(c.text)
		old.SetBorder(c.border).SetTitle(c.title)
		in := newNativeInput().SetLabel(c.label).SetFieldWidth(c.width).SetText(c.text)
		in.SetBorder(c.border).SetTitle(c.title)
		if c.focus {
			old.Focus(func(tview.Primitive) {})
			in.Focus(func(tview.Primitive) {})
		}
		want, got := cellsOf(t, old, c.w, c.h), cellsOf(t, in, c.w, c.h)
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: cell %d,%d is %q, the input field drew %q\nkit:\n%stview:\n%s", c.name, i%c.w, i/c.w, got[i], want[i], rowsOf(got, c.w), rowsOf(want, c.w))
				break
			}
		}
	}
}

func TestNativeInputKeysAndCalls(t *testing.T) {
	keepTheme(t)
	in := newNativeInput().SetLabel("> ")
	var changes []string
	var done []tcell.Key
	in.SetChangedFunc(func(s string) { changes = append(changes, s) }).SetDoneFunc(func(k tcell.Key) { done = append(done, k) })
	in.Focus(func(tview.Primitive) {})
	press := func(k tcell.Key, r rune) {
		in.InputHandler()(tcell.NewEventKey(k, r, tcell.ModNone), func(tview.Primitive) {})
	}
	for _, r := range "hut" {
		press(tcell.KeyRune, r)
	}
	press(tcell.KeyBackspace2, 0)
	if in.GetText() != "hu" || len(changes) != 4 {
		t.Errorf("typed %q with changes %q", in.GetText(), changes)
	}
	press(tcell.KeyEnter, 0)
	press(tcell.KeyEsc, 0)
	press(tcell.KeyTab, 0)
	if len(done) != 3 || done[0] != tcell.KeyEnter {
		t.Errorf("Enter, Esc and Tab are done keys: %v", done)
	}
	// A click puts the cursor at the character.
	cellsOf(t, in, 20, 1)
	in.root.Mouse(kit.Mouse{Kind: kit.MousePress, X: 3, Y: 0})
	press(tcell.KeyRune, 'X')
	if got := in.GetText(); got != "hXu" {
		t.Errorf("after a click and a letter: %q", got)
	}
	in.SetFieldBackgroundColor(theme.Color(theme.RoleBackground)).SetLabelColor(theme.Color(theme.RoleHighlight)).SetFieldTextColor(theme.Color(theme.RoleText))
	if in.GetLabel() != "> " || in.SetTitle("T").GetTitle() != "T" {
		t.Error("the label and the title read back")
	}
}

func TestNativeInputMask(t *testing.T) {
	keepTheme(t)
	in := newNativeInput().SetFieldWidth(10).SetMaskCharacter('·').SetText("abc")
	cells := cellsOf(t, in, 10, 1)
	if got := rowsOf(cells, 10); !strings.HasPrefix(got, "···") || strings.Contains(got, "a") {
		t.Errorf("a masked field draws the mask: %q", got)
	}
	if in.GetText() != "abc" {
		t.Errorf("the text is kept: %q", in.GetText())
	}
}
