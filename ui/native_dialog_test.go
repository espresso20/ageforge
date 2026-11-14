package ui

import (
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeDialog shows the cells tview's Modal showed, for the dialogs the
// game puts up: one button and two, a long text that wraps, text with blank
// lines and leading spaces, the danger colours, and no button at all.
func TestNativeDialogDrawsAsAModalDid(t *testing.T) {
	keepTheme(t)
	cases := map[string]struct {
		text    string
		buttons []string
		danger  bool
	}{
		"one button":   {"You're up to date.\n\nv4.0.0 is the latest version.", []string{"OK"}, false},
		"two buttons":  {"  ✦  Update available  ✦\n\n  Latest:   v4.1.0\n  Current:  v4.0.0\n\nDownload and install now?", []string{"Update now", "Later"}, false},
		"long text":    {"Every save of the account Test Account will be deleted. This cannot be undone, and there is no way to bring a deleted save back afterward.", []string{"Cancel", "Delete all saves"}, true},
		"tags":         {"[yellow]Careful.[-] This is [::b]bold[::-] and long enough to wrap onto another line of the box.", []string{"OK"}, false},
		"no buttons":   {"Checking for updates...", nil, false},
		"wide buttons": {"Pick.", []string{"The first of three long buttons", "The second of three long buttons", "The third of them"}, false},
	}
	for name, c := range cases {
		for _, size := range [][2]int{{80, 24}, {50, 15}, {30, 9}} {
			var frames [2]frame
			for i := range frames {
				var p drawable
				if i == 0 {
					m := tview.NewModal().SetText(c.text).AddButtons(c.buttons)
					if c.danger {
						m.SetBackgroundColor(theme.Tcell(theme.Of(theme.RoleNegative)))
						m.SetTextColor(theme.Tcell(theme.Of(theme.RoleOnNegative)))
					}
					tview.NewApplication().SetFocus(m)
					p = m
				} else {
					app := newAppLoop()
					d := newNativeDialog(app).SetText(c.text).AddButtons(c.buttons)
					if c.danger {
						styleDangerModal(d)
					}
					app.SetFocus(d)
					p = d
				}
				frames[i] = drawWrapped(t, p, size[0], size[1], 2)
			}
			old, now := frames[0], frames[1]
			for i := range old.cells {
				a, b := old.cells[i], now.cells[i]
				if string(a.Runes) != string(b.Runes) || a.Style != b.Style {
					t.Errorf("%s at %dx%d: cell %d,%d is %q %v, the modal drew %q %v\nnow:\n%s\nthe modal:\n%s",
						name, size[0], size[1], i%old.w, i/old.w, string(b.Runes), b.Style, string(a.Runes), a.Style, now.Text(), old.Text())
					break
				}
			}
		}
	}
}

// The dialog's keys are the Modal's: Tab and the arrows walk the buttons
// round the ends, Enter presses the one the keyboard is on, Esc is -1. A
// click presses a button too.
func TestNativeDialogKeysAndMouse(t *testing.T) {
	keepTheme(t)
	app := newAppLoop()
	var got []int
	d := newNativeDialog(app).SetText("Delete?").AddButtons([]string{"Cancel", "Delete", "Later"}).
		SetDoneFunc(func(i int, _ string) { got = append(got, i) })
	app.SetFocus(d)
	sim := newSimScreen(t, 60, 15)
	scr := newNativeScreen(theme.WrapScreen(sim))
	draw := func() {
		d.SetRect(0, 0, 60, 15)
		d.Draw(scr)
		scr.Show()
	}
	draw()
	press := func(code kit.KeyCode) { d.InputHandler()(tev(code, 0, 0), func(Primitive) {}) }
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
	draw()
	x, y, _, _ := d.buttons[2].GetRect()
	scr.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x + 1, Y: y})
	scr.mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: x + 1, Y: y})
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("a click on the third button called done with %v", got)
	}
	// Round the box the page beneath keeps its cells.
	if scr.hostAt(0, 0) == d.nativeHost || scr.hostAt(x+1, y) != d.nativeHost {
		t.Error("the dialog should own its box and nothing round it")
	}
}
