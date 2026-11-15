package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// A keycap button stands for one press of its key, and a pair of keys (↑↓)
// for none.
func TestWhatAKeycapStandsFor(t *testing.T) {
	for name, want := range map[string]kit.Key{
		"Enter": {Code: kit.KeyEnter}, "Esc": {Code: kit.KeyEsc}, "Tab": {Code: kit.KeyTab},
		"D": {Code: kit.KeyRune, Rune: 'd'}, "n": {Code: kit.KeyRune, Rune: 'n'},
	} {
		if got, ok := capKey(name); !ok || got != want {
			t.Errorf("the keycap %q stands for %+v (a key: %v), want %+v", name, got, ok, want)
		}
	}
	for _, name := range []string{"↑↓", "PgUp PgDn", ""} {
		if _, ok := capKey(name); ok {
			t.Errorf("the keycap %q should stand for no single press", name)
		}
	}
	if strings.Contains(footerButton("↑↓", "Move"), `["`) || !strings.HasPrefix(footerButton("Esc", "Back"), `["key Esc"]`) {
		t.Errorf("a button is a click region when its key is one press and only then: %q, %q", footerButton("↑↓", "Move"), footerButton("Esc", "Back"))
	}
	if got, want := visibleLen(footerButton("Esc", "Back")), visibleLen(footerButton("↑↓", "Back")); got != want+1 {
		t.Errorf("the region takes room on the bar: %d cells for Esc Back, %d for ↑↓ Back", got, want)
	}
}

// A click on a keycap button presses its key, as the keyboard would: on the
// theme picker a click on Esc Cancel leaves the picker, and the pointer is
// the hand over the button.
func TestAClickOnAKeycapButtonPressesItsKey(t *testing.T) {
	restoreForge(t)
	pages := newNativePages()
	pages.AddPage("home", newNativeBox(), true, false)
	app, frame := runLoop(t, pages, 120, 40)
	app.QueueUpdateDraw(func() {
		picker := CreateThemePickerPage(app, pages, nil, "home")
		pages.AddPage(themePickerPage, picker, true, true)
		app.SetFocus(picker)
	})
	b := frame()
	x, y := -1, -1
	for j := 0; j < b.H && x < 0; j++ {
		if i := strings.Index(b.Line(j), "Cancel"); i >= 0 {
			x, y = len([]rune(b.Line(j)[:i]))+1, j
		}
	}
	if x < 0 {
		t.Fatal("the picker's foot does not show Cancel")
	}
	app.Mouse(kit.Mouse{Kind: kit.MouseMove, X: x, Y: y})
	if b = frame(); b.Pointer != kit.PointerHand {
		t.Errorf("over the Cancel button the pointer's shape is %d, want the hand", b.Pointer)
	}
	app.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x, Y: y})
	app.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: x, Y: y})
	frame() // the click has run, and queued the key
	frame() // the key has run
	open := true
	app.QueueUpdate(func() { open = pages.HasPage(themePickerPage) })
	if open {
		t.Error("a click on Esc Cancel left the theme picker open")
	}
}
