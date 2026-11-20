package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// pump runs what the app has queued, as its loop would.
func pump(a *appLoop) {
	for {
		select {
		case f := <-a.events:
			f()
		default:
			return
		}
	}
}

// litEntry is the entry the pointer is on, -1 for none.
func litEntry(m *mainMenu) int {
	k, ok := m.buttons.lit(&m.scene.keys)
	if !ok {
		return -1
	}
	for i, r := range m.scene.hits {
		if k.r.In(r) {
			return i
		}
	}
	return -1
}

func TestMenuMouseLightsAndOpens(t *testing.T) {
	r := newMenuRig(t, "ashford")
	r.m.set.Motion = false // a choice opens at once
	r.draw(120, 40)
	row := func(id string) image.Point {
		for i, it := range r.m.view.items {
			if it.id == id {
				return r.m.scene.hits[i].Min.Add(image.Pt(2, 0))
			}
		}
		t.Fatalf("no entry %q", id)
		return image.Point{}
	}
	idOf := func(i int) string {
		if i < 0 {
			return ""
		}
		return r.m.view.items[i].id
	}
	at := row(miSettings)
	if !r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y}) {
		t.Error("moving onto another entry should change the page")
	}
	if got := idOf(litEntry(r.m)); got != miSettings {
		t.Errorf("the entry under the pointer is lit, got %q", got)
	}
	if got := r.m.view.items[r.m.view.sel].id; got != miContinue {
		t.Errorf("the pointer does not move the picked row, got %q", got)
	}
	if r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X + 1, Y: at.Y}) {
		t.Error("moving along the same entry changes nothing")
	}
	if !r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: 0, Y: 0}) || litEntry(r.m) != -1 {
		t.Error("off the entries nothing is lit")
	}
	if r.m.PointerAt(at) != kit.PointerHand || r.m.PointerAt(image.Pt(0, 0)) != kit.PointerArrow {
		t.Error("the hand should show over an entry and not off the list")
	}
	r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	r.m.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X, Y: at.Y})
	r.m.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: at.X, Y: at.Y})
	pump(r.app)
	if !r.m.settings {
		t.Fatal("a click on Settings should open the Settings page")
	}
	// The Settings page is a list too: Back is its last row.
	scr := r.draw(120, 40)
	if !scr.has("Back") {
		t.Fatal("the Settings page should be drawn")
	}
	at = row(miSetBack)
	r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	r.m.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X, Y: at.Y})
	pump(r.app)
	if r.m.settings {
		t.Error("a click on Back should leave the Settings page")
	}
	// Under the game's least size the list alone is drawn, and still takes
	// the mouse.
	r.draw(40, 12)
	at = row(miNew)
	r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	if got := idOf(litEntry(r.m)); got != miNew {
		t.Errorf("on the small page the pointer lights too, got %q", got)
	}
}
