package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

func TestMenuMouseSelectsAndOpens(t *testing.T) {
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
	at := row(miSettings)
	if !r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y}) {
		t.Error("moving onto another entry should change the page")
	}
	if got := r.m.view.items[r.m.view.sel].id; got != miSettings {
		t.Errorf("the entry under the pointer is selected, got %q", got)
	}
	if r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X + 1, Y: at.Y}) {
		t.Error("moving along the same entry changes nothing")
	}
	if r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: 0, Y: 0}) || r.m.view.items[r.m.view.sel].id != miSettings {
		t.Error("off the entries the selection stays where it was")
	}
	r.m.Mouse(kit.Mouse{Kind: kit.MousePress, X: at.X, Y: at.Y})
	if !r.m.settings {
		t.Fatal("a click on Settings should open the Settings page")
	}
	// The Settings page is a list too: Back is its last row.
	scr := r.draw(120, 40)
	if !scr.has("Back") {
		t.Fatal("the Settings page should be drawn")
	}
	at = row(miSetBack)
	r.m.Mouse(kit.Mouse{Kind: kit.MousePress, X: at.X, Y: at.Y})
	if r.m.settings {
		t.Error("a click on Back should leave the Settings page")
	}
	// Under the game's least size the list alone is drawn, and still takes
	// the mouse.
	r.draw(40, 12)
	at = row(miNew)
	r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	if got := r.m.view.items[r.m.view.sel].id; got != miNew {
		t.Errorf("on the small page the pointer selects too, got %q", got)
	}
}
