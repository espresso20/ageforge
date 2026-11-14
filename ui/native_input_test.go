package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

func TestNativeInputKeysAndCalls(t *testing.T) {
	keepTheme(t)
	in := newNativeInput().SetLabel("> ")
	var changes []string
	var done []kit.KeyCode
	in.SetChangedFunc(func(s string) { changes = append(changes, s) }).SetDoneFunc(func(k kit.KeyCode) { done = append(done, k) })
	in.Focus(func(Primitive) {})
	press := func(k kit.KeyCode, r rune) {
		keyTo(in, tk(k, r, 0))
	}
	for _, r := range "hut" {
		press(kit.KeyRune, r)
	}
	press(kit.KeyBackspace, 0)
	if in.GetText() != "hu" || len(changes) != 4 {
		t.Errorf("typed %q with changes %q", in.GetText(), changes)
	}
	press(kit.KeyEnter, 0)
	press(kit.KeyEsc, 0)
	press(kit.KeyTab, 0)
	if len(done) != 3 || done[0] != kit.KeyEnter {
		t.Errorf("Enter, Esc and Tab are done keys: %v", done)
	}
	// A click puts the cursor at the character.
	newStage(in, 20, 1).mouse(kit.Mouse{Kind: kit.MousePress, X: 3, Y: 0})
	press(kit.KeyRune, 'X')
	if got := in.GetText(); got != "hXu" {
		t.Errorf("after a click and a letter: %q", got)
	}
	in.SetFieldBackgroundColor(theme.Of(theme.RoleBackground)).SetLabelColor(theme.Of(theme.RoleHighlight)).SetFieldTextColor(theme.Of(theme.RoleText))
	if in.GetLabel() != "> " || in.SetTitle("T").GetTitle() != "T" {
		t.Error("the label and the title read back")
	}
}

func TestNativeInputMask(t *testing.T) {
	keepTheme(t)
	in := newNativeInput().SetFieldWidth(10).SetMaskCharacter('·').SetText("abc")
	if got := drawPrimitive(t, in, 10, 1).Rows()[0]; !strings.HasPrefix(got, "···") || strings.Contains(got, "a") {
		t.Errorf("a masked field draws the mask: %q", got)
	}
	if in.GetText() != "abc" {
		t.Errorf("the text is kept: %q", in.GetText())
	}
}
