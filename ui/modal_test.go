package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
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

	key := func(k tcell.Key, r rune) bool { return ring.key(tcell.NewEventKey(k, r, tcell.ModNone)) }
	for _, step := range []struct {
		key  tcell.Key
		want int
	}{{tcell.KeyTab, 1}, {tcell.KeyTab, 2}, {tcell.KeyTab, 0}, {tcell.KeyBacktab, 2}, {tcell.KeyLeft, 1}, {tcell.KeyRight, 2}} {
		if !key(step.key, 0) || ring.at != step.want || held != ring.buttons[step.want] {
			t.Fatalf("key %v: on button %d (held %v), want %d", step.key, ring.at, held, step.want)
		}
	}

	if !key(tcell.KeyRune, 'S') || len(pressed) != 1 || pressed[0] != "second" || ring.at != 2 {
		t.Errorf("S pressed %v, now on %d: want second pressed, still on 2", pressed, ring.at)
	}
	if key(tcell.KeyRune, 'x') || key(tcell.KeyEnter, 0) || len(pressed) != 1 {
		t.Errorf("a key that is no button's was taken, or pressed something: %v", pressed)
	}
}

// TestTakeKeysRunsAndSwallowsOnlyItsKeys: a key in the map runs its function
// and is taken; every other key goes on.
func TestTakeKeysRunsAndSwallowsOnlyItsKeys(t *testing.T) {
	window := tview.NewFlex()
	ran := 0
	takeKeys(window, map[tcell.Key]func(){tcell.KeyEsc: func() { ran++ }})
	esc, tab := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone), tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	if got := window.GetInputCapture()(esc); got != nil || ran != 1 {
		t.Errorf("Esc: returned %v, ran %d times; want it taken and run once", got, ran)
	}
	if got := window.GetInputCapture()(tab); got != tab || ran != 1 {
		t.Errorf("Tab: returned %v, ran %d times; want it passed on", got, ran)
	}
}
