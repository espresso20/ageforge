package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/window/kit"
)

// runLoop starts a loop on root at w by h and stops it when the test ends.
// It returns the loop and a func that waits until the loop has handled
// everything sent so far and returns the frame the window would show.
func runLoop(t *testing.T, root Primitive, w, h int) (*appLoop, func() *kit.Buffer) {
	t.Helper()
	a := newAppLoop()
	a.SetRoot(root, true)
	ended := make(chan error, 1)
	go func() { ended <- a.Run() }()
	t.Cleanup(func() {
		a.Stop()
		select {
		case <-ended:
		case <-time.After(3 * time.Second):
			t.Error("Run did not return after Stop")
		}
	})
	a.Resize(w, h)
	frame := func() *kit.Buffer {
		a.QueueUpdate(func() {}) // everything queued before this has run
		var b kit.Buffer
		a.Frame(&b)
		return &b
	}
	return a, frame
}

// The loop is the whole path from the window to the screen: a key typed is
// in the next frame, an update queued from another goroutine is painted, and
// the frame is the size the window said.
func TestTheLoopTakesKeysAndPaintsFrames(t *testing.T) {
	keepTheme(t)
	in := newNativeInput().SetLabel("> ")
	a, frame := runLoop(t, in, 20, 3)
	if f := frame(); f.W != 20 || f.H != 3 || !strings.HasPrefix(f.Line(0), ">") {
		t.Fatalf("the first frame is %dx%d: %q", f.W, f.H, f.Line(0))
	}
	for _, r := range "hut" {
		a.Key(tk(kit.KeyRune, r, 0))
	}
	a.Key(tk(kit.KeyBackspace, 0, 0))
	if f := frame(); f.Line(0) != "> hu" || !f.CursorOn || f.Cursor.X != 4 {
		t.Errorf("after typing: %q, cursor %v shown %v", f.Line(0), f.Cursor, f.CursorOn)
	}
	// An update from another goroutine runs on the loop and is painted
	// before QueueUpdateDraw returns.
	a.QueueUpdateDraw(func() { in.SetText("farm") })
	var b kit.Buffer
	a.Frame(&b)
	if b.Line(0) != "> farm" {
		t.Errorf("after a queued update the frame reads %q", b.Line(0))
	}
	a.Resize(30, 2)
	if f := frame(); f.W != 30 || f.H != 2 {
		t.Errorf("after a resize the frame is %dx%d", f.W, f.H)
	}
}

// A key passes the app's capture first, which may take it or put another in
// its place, and reaches nothing when nothing in the root has the keyboard.
func TestTheLoopsCaptureSeesAKeyFirst(t *testing.T) {
	keepTheme(t)
	in := newNativeInput()
	a, frame := runLoop(t, in, 20, 1)
	a.SetInputCapture(func(k *kit.Key) *kit.Key {
		switch {
		case k.Is('x'):
			return nil
		case k.Is('a'):
			return &kit.Key{Code: kit.KeyRune, Rune: 'b'}
		}
		return k
	})
	for _, r := range "xaz" {
		a.Key(tk(kit.KeyRune, r, 0))
	}
	if frame(); in.GetText() != "bz" {
		t.Errorf("x taken, a made b, z let by: the field reads %q", in.GetText())
	}
	a.QueueUpdate(func() { a.SetFocus(nil) })
	a.Key(tk(kit.KeyRune, 'q', 0))
	if frame(); in.GetText() != "bz" {
		t.Errorf("with the keyboard nowhere a key still typed: %q", in.GetText())
	}
}

// Ctrl+C ends the loop, as it always ended the game, and a stopped loop
// turns away what is sent to it without making the sender wait.
func TestCtrlCStopsTheLoop(t *testing.T) {
	a := newAppLoop()
	a.SetRoot(newNativeInput(), true)
	ended := make(chan error, 1)
	go func() { ended <- a.Run() }()
	a.Resize(10, 1)
	a.Key(kit.Key{Code: kit.KeyChord, Rune: 'c', Mod: kit.ModCtrl})
	select {
	case err := <-ended:
		if err != nil {
			t.Errorf("Run returned %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Ctrl+C did not end the loop")
	}
	done := make(chan struct{})
	go func() {
		a.QueueUpdateDraw(func() { t.Error("an update ran on a stopped loop") })
		a.Key(tk(kit.KeyEnter, 0, 0))
		a.Mouse(kit.Mouse{Kind: kit.MouseMove})
		a.Resize(5, 5)
		a.Stop() // twice is fine
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("a call on a stopped loop did not return")
	}
}

// The mouse goes through the loop to the widget under the pointer.
func TestTheLoopTakesTheMouse(t *testing.T) {
	keepTheme(t)
	l := newNativeList().ShowSecondaryText(false)
	for _, s := range []string{"one", "two", "three", "four"} {
		l.AddItem(s, "", 0, nil)
	}
	a, frame := runLoop(t, l, 20, 4)
	frame()
	a.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 2, Y: 2})
	a.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: 2, Y: 2})
	frame()
	var at int
	a.QueueUpdate(func() { at = l.GetCurrentItem() })
	if at != 2 {
		t.Errorf("a click on the third row left the list on row %d", at)
	}
}

// A right click presses Esc where the app says Esc goes back, and does
// nothing where it says no or says nothing; a left click is never Esc.
func TestARightClickIsEscWhereTheAppSaysSo(t *testing.T) {
	keepTheme(t)
	a, frame := runLoop(t, newNativeBox(), 20, 4)
	var keys []kit.KeyCode
	a.SetInputCapture(func(k *kit.Key) *kit.Key {
		keys = append(keys, k.Code)
		return nil
	})
	press := func(b kit.MouseButton) []kit.KeyCode {
		a.QueueUpdate(func() { keys = nil })
		a.Mouse(kit.Mouse{Kind: kit.MousePress, Button: b, X: 3, Y: 1})
		a.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: b, X: 3, Y: 1})
		frame()
		var got []kit.KeyCode
		a.QueueUpdate(func() { got = append(got, keys...) })
		return got
	}
	if got := press(kit.ButtonRight); len(got) != 0 {
		t.Errorf("with nothing said, a right click sent the keys %v", got)
	}
	allow := false
	a.SetBack(func() bool { return allow })
	if got := press(kit.ButtonRight); len(got) != 0 {
		t.Errorf("where Esc must not be pressed, a right click sent the keys %v", got)
	}
	allow = true
	if got := press(kit.ButtonRight); len(got) != 1 || got[0] != kit.KeyEsc {
		t.Errorf("where Esc goes back, a right click sent the keys %v, want one Esc", got)
	}
	if got := press(kit.ButtonLeft); len(got) != 0 {
		t.Errorf("a left click sent the keys %v", got)
	}
}
