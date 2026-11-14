package ui

import (
	"testing"

	"github.com/espresso20/ageforge/window/kit"
	"github.com/rivo/tview"
)

// TestButtonRingMovesAndPresses: Tab and Shift-Tab go round the ring, a
// button's letter (either case) presses it without moving the keyboard, and
// any other key is left for the caller.
func TestButtonRingMovesAndPresses(t *testing.T) {
	var pressed []string
	var held tview.Primitive
	ring := newButtonRing(tview.NewApplication(), func(p tview.Primitive) { held = p })
	for _, b := range []struct {
		name   string
		letter rune
	}{{"first", 'f'}, {"second", 's'}, {"third", 0}} {
		name := b.name
		ring.add(tview.NewButton(name).SetSelectedFunc(func() { pressed = append(pressed, name) }), b.letter)
	}

	key := func(k kit.KeyCode, r rune) bool { return ring.key(newKey(k, r, 0)) }
	for _, step := range []struct {
		key  kit.KeyCode
		want int
	}{{kit.KeyTab, 1}, {kit.KeyTab, 2}, {kit.KeyTab, 0}, {kit.KeyBacktab, 2}, {kit.KeyLeft, 1}, {kit.KeyRight, 2}} {
		if !key(step.key, 0) || ring.at != step.want || held != ring.buttons[step.want] {
			t.Fatalf("key %v: on button %d (held %v), want %d", step.key, ring.at, held, step.want)
		}
	}

	if !key(kit.KeyRune, 'S') || len(pressed) != 1 || pressed[0] != "second" || ring.at != 2 {
		t.Errorf("S pressed %v, now on %d: want second pressed, still on 2", pressed, ring.at)
	}
	if key(kit.KeyRune, 'x') || key(kit.KeyEnter, 0) || len(pressed) != 1 {
		t.Errorf("a key that is no button's was taken, or pressed something: %v", pressed)
	}
}

// TestTakeKeysRunsAndSwallowsOnlyItsKeys: a key in the map runs its function
// and is taken; every other key goes on.
func TestTakeKeysRunsAndSwallowsOnlyItsKeys(t *testing.T) {
	window := newNativeFlex()
	ran := 0
	takeKeys(window, map[kit.KeyCode]func(){kit.KeyEsc: func() { ran++ }})
	esc, tab := tev(kit.KeyEsc, 0, 0), tev(kit.KeyTab, 0, 0)
	if got := window.GetInputCapture()(esc); got != nil || ran != 1 {
		t.Errorf("Esc: returned %v, ran %d times; want it taken and run once", got, ran)
	}
	if got := window.GetInputCapture()(tab); got != tab || ran != 1 {
		t.Errorf("Tab: returned %v, ran %d times; want it passed on", got, ran)
	}
}
