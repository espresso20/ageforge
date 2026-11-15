package ui

import (
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

func TestNativeButtonPressedByEnterAndAClick(t *testing.T) {
	keepTheme(t)
	n := 0
	b := newNativeButton("Go").SetSelectedFunc(func() { n++ })
	key := func(k kit.KeyCode, r rune) {
		keyTo(b, tk(k, r, 0))
	}
	key(kit.KeyRune, ' ')
	key(kit.KeyEsc, 0)
	key(kit.KeyTab, 0)
	if n != 0 {
		t.Errorf("only Enter presses a button, got %d presses", n)
	}
	key(kit.KeyEnter, 0)
	if n != 1 {
		t.Errorf("Enter pressed %d times", n)
	}
	newStage(b, 8, 1).click(2, 0)
	if n != 2 {
		t.Errorf("a click pressed it %d times in all, want 2", n)
	}
}
