package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// The dialog's keys: Tab and the arrows walk the buttons
// round the ends, Enter presses the one the keyboard is on, Esc is -1. A
// click presses a button too.
func TestNativeDialogKeysAndMouse(t *testing.T) {
	keepTheme(t)
	app := newAppLoop()
	var got []int
	d := newNativeDialog(app).SetText("Delete?").AddButtons([]string{"Cancel", "Delete", "Later"}).
		SetDoneFunc(func(i int, _ string) { got = append(got, i) })
	app.SetFocus(d)
	st := newStage(d, 60, 15)
	press := func(code kit.KeyCode) { keyTo(d, tk(code, 0, 0)) }
	on := func() int {
		for i, b := range d.buttons {
			if b.HasFocus() {
				return i
			}
		}
		return -1
	}
	if on() != 0 {
		t.Fatalf("the keyboard starts on button %d", on())
	}
	for _, step := range []struct {
		key  kit.KeyCode
		want int
	}{{kit.KeyTab, 1}, {kit.KeyRight, 2}, {kit.KeyDown, 0}, {kit.KeyBacktab, 2}, {kit.KeyLeft, 1}, {kit.KeyUp, 0}} {
		press(step.key)
		if on() != step.want {
			t.Errorf("after key %d the keyboard is on button %d, want %d", step.key, on(), step.want)
		}
	}
	press(kit.KeyTab)
	press(kit.KeyEnter)
	press(kit.KeyEsc)
	if len(got) != 2 || got[0] != 1 || got[1] != -1 {
		t.Errorf("Enter on the second button then Esc called done with %v", got)
	}
	got = nil
	st.paint()
	x, y, _, _ := d.buttons[2].GetRect()
	st.click(x+1, y)
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("a click on the third button called done with %v", got)
	}
	// Round its box the dialog is not there: the page beneath shows, and
	// the pointer is that page's.
	if d.Covers(image.Pt(0, 0)) || !d.Covers(image.Pt(x+1, y)) || len(d.covered()) != 1 {
		t.Error("the dialog should cover its box and nothing round it")
	}
}
