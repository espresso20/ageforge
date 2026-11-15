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
	// Round its box the dialog paints nothing, so the page beneath shows,
	// but the pointer is the dialog's there too.
	if !d.Covers(image.Pt(0, 0)) || !d.Covers(image.Pt(x+1, y)) || len(d.covered()) != 1 {
		t.Error("the dialog should paint its box only and keep the pointer over the whole page")
	}
}

// A press or the wheel beside a dialog reaches nothing on the page beneath:
// the dialog has the keyboard, and the pointer goes with it.
func TestADialogKeepsThePointerFromThePageBeneath(t *testing.T) {
	keepTheme(t)
	under := newNativeList()
	for _, s := range []string{"one", "two", "three", "four"} {
		under.AddItem(s, "", 0, nil)
	}
	pages := newNativePages()
	pages.AddPage("under", under, true, true)
	asked := 0
	d := showButtonModal(newAppLoop(), pages, "ask", "Sure?", []string{"Yes", "No"}, func(int, string) { asked++ })
	st := newStage(pages, 60, 20)
	bx, by, _, _ := d.buttons[0].GetRect()
	if bx < 3 || by < 3 {
		t.Fatalf("the dialog's first button is at %d,%d: no room beside the box to test", bx, by)
	}
	st.click(2, 2) // the third row of the list beneath, beside the dialog
	st.mouse(kit.Mouse{Kind: kit.MouseWheel, X: 2, Y: 2, Wheel: 2})
	if under.GetCurrentItem() != 0 || asked != 0 {
		t.Errorf("a press beside the dialog moved the list beneath to row %d (answers given: %d)", under.GetCurrentItem(), asked)
	}
	st.paint()
	// The pointer is the hand on the dialog's button, and the arrow beside
	// the dialog, though a row of the list is under it there.
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: bx + 1, Y: by})
	if st.paint(); st.buf.Pointer != kit.PointerHand {
		t.Errorf("over the dialog's button the pointer's shape is %d, want the hand", st.buf.Pointer)
	}
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: 2, Y: 1})
	if st.paint(); st.buf.Pointer != kit.PointerArrow {
		t.Errorf("beside the dialog, over a row of the list beneath, the shape is %d, want the arrow", st.buf.Pointer)
	}
	st.click(bx+1, by)
	if asked != 1 {
		t.Errorf("a click on the dialog's button gave %d answers, want 1", asked)
	}
}
