package ui

import (
	"image"
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

func TestSplitKeyedAndKeyHints(t *testing.T) {
	for _, c := range []struct{ in, key, label string }{
		{"[E] ENDURE", "E", "ENDURE"}, {"[A[] Accept", "A", "Accept"}, {"plain", "", "plain"}, {"[gray]x", "", "[gray]x"},
	} {
		if k, l := splitKeyed(c.in); k != c.key || l != c.label {
			t.Errorf("splitKeyed(%q) = %q, %q", c.in, k, l)
		}
	}
	hints, rest := keyHints("[gray]Enter: confirm  ·  Tab: reroll name  ·  Esc: cancel[-]")
	if want := [][2]string{{"Enter", "confirm"}, {"Tab", "reroll name"}, {"Esc", "cancel"}}; !reflect.DeepEqual(hints, want) || rest != "" {
		t.Errorf("hints %v rest %q", hints, rest)
	}
	hints, rest = keyHints("[gray]Esc: decide later · advancing waits · type 'x' to reopen[-]")
	if len(hints) != 1 || hints[0] != [2]string{"Esc", "decide later"} || rest != "advancing waits · type 'x' to reopen" {
		t.Errorf("hints %v rest %q", hints, rest)
	}
}

// A dialog's choices are action buttons: the default names Enter and is
// primary, the way out names Esc and is plain, and a destructive one is
// danger with the way out first. A click on each answers as its key does.
func TestDialogChoicesAreActionButtons(t *testing.T) {
	keepTheme(t)
	for _, danger := range []bool{false, true} {
		app := newAppLoop()
		var got []int
		d := newNativeDialog(app).SetText("Sure?").AddButtons([]string{"Yes", "Maybe", "No"}).
			SetDoneFunc(func(i int, _ string) { got = append(got, i) })
		d.asActions(danger)
		kinds, keys := []keyKind{}, []string{}
		for _, b := range d.buttons {
			kinds, keys = append(kinds, b.act), append(keys, b.key)
		}
		wantKinds, wantKeys := []keyKind{keyPrimary, keyPlain, keyPlain}, []string{"Enter", "", "Esc"}
		if danger {
			wantKinds, wantKeys = []keyKind{keyPlain, keyPlain, keyDanger}, []string{"Esc", "", ""}
		}
		if !reflect.DeepEqual(kinds, wantKinds) || !reflect.DeepEqual(keys, wantKeys) {
			t.Errorf("danger %v: kinds %v keys %v, want %v %v", danger, kinds, keys, wantKinds, wantKeys)
		}
		app.SetFocus(d)
		st := newStage(d, 70, 20)
		for i, b := range d.buttons {
			x, y, _, _ := b.GetRect()
			st.click(x+1, y+1)
			if n := len(got); n != i+1 || got[n-1] != i {
				t.Errorf("danger %v: a click on button %d gave %v", danger, i, got)
			}
		}
	}
}

// A row of choices: a click presses its choice, and one marked off does
// nothing.
func TestChoiceRowClicks(t *testing.T) {
	keepTheme(t)
	var got []string
	mk := func(name string, off bool) choice {
		return choice{label: name, key: "K", kind: keyPlain, off: off, press: func() { got = append(got, name) }}
	}
	row, bs := choiceRow([]choice{mk("one", false), mk("two", true), mk("three", false)}, 2, -1)
	st := newStage(row, 70, actionRows)
	for _, b := range bs {
		x, y, _, _ := b.GetRect()
		st.click(x+1, y+1)
	}
	if !reflect.DeepEqual(got, []string{"one", "three"}) {
		t.Errorf("the clicks pressed %v", got)
	}
}

// A click on a tab selects it, the open tab and the others paint differently
// with pictures, and without them the cells keep the text.
func TestTabPlates(t *testing.T) {
	keepTheme(t)
	withPictures(t)
	var s keySet
	s.tab(image.Rect(2, 1, 10, 2), "ONE", 0, true)
	s.tab(image.Rect(11, 1, 19, 2), "TWO", 1, false)
	if n, ok := s.tabAt(image.Pt(12, 1)); !ok || n != 1 {
		t.Errorf("tabAt over the second tab: %d %v", n, ok)
	}
	if _, ok := s.tabAt(image.Pt(30, 1)); ok {
		t.Error("tabAt past the tabs found one")
	}
	var k keyPresser
	b := kit.NewBuffer(30, 3)
	k.paintTabs(kit.NewPainter(b), &s, image.Point{}, newExpPalette())
	var open, back *kit.Pic
	for i := range b.Pics {
		switch b.Pics[i].Rect.Min.X {
		case 2:
			open = &b.Pics[i]
		case 11:
			back = &b.Pics[i]
		}
	}
	if open == nil || back == nil || reflect.DeepEqual(open.Img.Pix, back.Img.Pix) {
		t.Fatal("the open tab and a back tab are not two different pictures")
	}
	// Hovered: a back tab comes forward.
	k.press, k.in, k.over = func(kit.Key) {}, true, image.Pt(12, 1)
	b2 := kit.NewBuffer(30, 3)
	k.paintTabs(kit.NewPainter(b2), &s, image.Point{}, newExpPalette())
	same := false
	for _, p := range b2.Pics {
		if p.Rect.Min.X == 11 && reflect.DeepEqual(p.Img.Pix, back.Img.Pix) {
			same = true
		}
	}
	if same {
		t.Error("a tab under the pointer looks as it does at rest")
	}
	// Without pictures nothing is painted over the cells.
	was := pictureCell
	pictureCell = nil
	b3 := kit.NewBuffer(30, 3)
	k.paintTabs(kit.NewPainter(b3), &s, image.Point{}, newExpPalette())
	pictureCell = was
	if len(b3.Pics) != 0 {
		t.Error("tabs painted pictures where there are none")
	}
}

// A click on a key prompt of the map's bar presses that key.
func TestMapBarKeysArePressable(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	p := d.mapPanel
	drawMapPanel(t, d, 160, 48)
	var got []kit.Key
	p.buttons.press = func(k kit.Key) { got = append(got, k) }
	found := false
	for _, k := range p.keys.keys {
		if k.key.Code != kit.KeyEsc {
			continue
		}
		found = true
		at := p.Bounds().Min.Add(k.r.Min)
		if !p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: at.X, Y: at.Y}) || len(got) != 1 || got[0].Code != kit.KeyEsc {
			t.Errorf("a click on Esc pressed %v", got)
		}
		if p.PointerAt(at) != kit.PointerHand {
			t.Error("no hand over the key")
		}
	}
	if !found {
		t.Error("the map's bar has no Esc key prompt")
	}
}

// A click on a tab of the Expeditions panel switches to its view.
func TestExpeditionTabsAreClickable(t *testing.T) {
	engine, p := expFixture(t)
	p.open(engine.GetState())
	p.Layout(image.Rect(0, 0, 100, 27))
	at := func(n int) image.Point { return p.grid.tabs[n].r.Min.Add(image.Pt(1, 0)) }
	click := func(pt image.Point) bool {
		return p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
	if len(p.grid.tabs) != 2 || p.view.report {
		t.Fatalf("%d tabs, report %v", len(p.grid.tabs), p.view.report)
	}
	if !click(at(1)) || !p.view.report {
		t.Error("a click on the second tab did not open the Report")
	}
	p.Layout(image.Rect(0, 0, 100, 27))
	if p.PointerAt(at(0)) != kit.PointerHand || !click(at(0)) || p.view.report {
		t.Error("a click on the first tab did not open the Ready Room")
	}
}
