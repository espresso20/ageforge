package ui

import (
	"image"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// TestCapPartsNameTheKeys: a keycap's label is split into the keys a click
// can press, and what names no key is left out.
func TestCapPartsNameTheKeys(t *testing.T) {
	codes := func(label string) (out []kit.KeyCode) {
		for _, p := range capParts(label) {
			out = append(out, p.key.Code)
		}
		return out
	}
	for label, want := range map[string][]kit.KeyCode{
		"Esc":              {kit.KeyEsc},
		"Enter":            {kit.KeyEnter},
		"↑↓":               {kit.KeyUp, kit.KeyDown},
		"↑/↓":              {kit.KeyUp, kit.KeyDown},
		"←↑↓→":             {kit.KeyLeft, kit.KeyUp, kit.KeyDown, kit.KeyRight},
		"PgUp PgDn":        {kit.KeyPgUp, kit.KeyPgDn},
		"PgUp/PgDn":        {kit.KeyPgUp, kit.KeyPgDn},
		"Home End":         {kit.KeyHome, kit.KeyEnd},
		"PgUp PgDn (⌥↑↓)":  {kit.KeyPgUp, kit.KeyPgDn},
		"A":                {kit.KeyRune},
		"+/−":              {kit.KeyRune, kit.KeyRune},
		"U/D":              {kit.KeyRune, kit.KeyRune},
		"1-9":              nil,
		"theme <key>":      nil,
		"PgUp PgDn (\\/↓)": {kit.KeyPgUp, kit.KeyPgDn},
	} {
		if got := codes(label); !reflect.DeepEqual(got, want) {
			t.Errorf("capParts(%q) = %v, want %v", label, got, want)
		}
	}
	if p := capParts("+/−"); len(p) != 2 || p[0].key.Rune != '+' || p[1].key.Rune != '-' {
		t.Errorf("capParts(+/−) = %+v, want plus then the minus the keyboard types", p)
	}
	if p := capParts("A"); len(p) != 1 || p[0].key.Rune != 'a' {
		t.Errorf("capParts(A) = %+v, want the letter as typed without Shift", p)
	}
}

// keyFrame is a text panel's frame on a stage, with three keys on its foot
// and a record of what a click pressed.
func keyFrame(t *testing.T) (*panelFrame, *stage, *[]kit.Key) {
	t.Helper()
	keepTheme(t)
	f := newPanelFrame(txtExpeditions, newNativeText())
	f.keys = func() [][2]string {
		return [][2]string{{"↑↓", "scroll"}, {"PgUp PgDn", "page"}, {"Esc", "close"}}
	}
	pressed := &[]kit.Key{}
	f.buttons.press = func(k kit.Key) { *pressed = append(*pressed, k) }
	st := newStage(f, 80, 24)
	st.paint()
	return f, st, pressed
}

// keyCell is a cell of the frame's button for the key code, its first or
// its last.
func keyCell(t *testing.T, f *panelFrame, code kit.KeyCode, last bool) image.Point {
	t.Helper()
	var hit *xkey
	for i := range f.grid.keys {
		if k := &f.grid.keys[i]; k.key.Code == code && (hit == nil || last) {
			hit = k
		}
	}
	if hit == nil {
		t.Fatalf("no button for key %v among %+v", code, f.grid.keys)
	}
	if last {
		return image.Pt(hit.r.Max.X-1, hit.r.Min.Y)
	}
	return hit.r.Min
}

// TestPanelFrameKeysAreButtons: a click on a key along the foot presses it,
// each key of a shared cap has its own cells, and the close button in the top
// edge presses Esc.
func TestPanelFrameKeysAreButtons(t *testing.T) {
	f, st, pressed := keyFrame(t)
	click := func(pt image.Point) {
		st.mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y})
		st.mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y})
	}
	foot := 24 - 2
	if got := st.buf.Line(foot); !strings.HasPrefix(got, "│  ↑↓  scroll   PgUp PgDn  page   Esc  close ") {
		t.Fatalf("the foot reads %q", got)
	}
	for _, c := range []kit.KeyCode{kit.KeyUp, kit.KeyDown, kit.KeyPgUp, kit.KeyPgDn} {
		click(keyCell(t, f, c, false))
	}
	// Esc has two buttons: the close button, drawn first, and the foot's,
	// where a click on the name counts too.
	closeAt, footEsc := keyCell(t, f, kit.KeyEsc, false), keyCell(t, f, kit.KeyEsc, true)
	if closeAt.Y != 0 || st.buf.At(closeAt.X+1, 0).R != '×' {
		t.Errorf("the close button is at %v showing %q, want × in the top edge", closeAt, st.buf.At(closeAt.X+1, 0).R)
	}
	if footEsc.Y != foot {
		t.Errorf("the foot's Esc button ends at %v, want row %d", footEsc, foot)
	}
	click(closeAt)
	click(footEsc)
	var got []kit.KeyCode
	for _, k := range *pressed {
		got = append(got, k.Code)
	}
	if want := []kit.KeyCode{kit.KeyUp, kit.KeyDown, kit.KeyPgUp, kit.KeyPgDn, kit.KeyEsc, kit.KeyEsc}; !reflect.DeepEqual(got, want) {
		t.Errorf("the clicks pressed %v, want %v", got, want)
	}
	// The name beside a cap of two keys is nobody's, and neither is the box.
	*pressed = nil
	down := keyCell(t, f, kit.KeyDown, true)
	click(image.Pt(down.X+2, foot))
	click(image.Pt(40, 10))
	if len(*pressed) != 0 {
		t.Errorf("clicks off the buttons pressed %v", *pressed)
	}
}

// TestPanelFrameKeysLightUp: the pointer is a hand over a button and the
// button changes colour under it, once, and goes back when it leaves.
func TestPanelFrameKeysLightUp(t *testing.T) {
	f, st, _ := keyFrame(t)
	esc := keyCell(t, f, kit.KeyEsc, true)
	rest := st.buf.At(esc.X, esc.Y)
	if f.PointerAt(esc) != kit.PointerHand || f.PointerAt(image.Pt(40, 10)) != kit.PointerArrow {
		t.Errorf("pointer over the button %v and off it %v, want the hand and the arrow", f.PointerAt(esc), f.PointerAt(image.Pt(40, 10)))
	}
	if !st.mouse(kit.Mouse{Kind: kit.MouseMove, X: esc.X, Y: esc.Y}) {
		t.Fatal("moving onto a button changed nothing")
	}
	st.paint()
	if lit := st.buf.At(esc.X, esc.Y); lit.Bg == rest.Bg {
		t.Errorf("the button under the pointer keeps its ground %v", lit.Bg)
	}
	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: 40, Y: 10})
	st.paint()
	if back := st.buf.At(esc.X, esc.Y); back.Bg != rest.Bg {
		t.Errorf("the button stays lit after the pointer left: %v, want %v", back.Bg, rest.Bg)
	}
}

// TestPanelFrameAloneHasNoButtons: a frame nobody gave a way to press keys
// (a picture of the screen) takes no click and shows no hand.
func TestPanelFrameAloneHasNoButtons(t *testing.T) {
	f, st, _ := keyFrame(t)
	f.buttons.press = nil
	esc := keyCell(t, f, kit.KeyEsc, true)
	if f.PointerAt(esc) != kit.PointerArrow {
		t.Error("the hand shows over a key nothing can press")
	}
	if f.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: esc.X, Y: esc.Y}) {
		t.Error("a click on a key nothing can press was taken")
	}
	st.paint()
}

// TestKeyBarKeepsTheCloseKey: when the foot has no room for every key, the
// last one (how to close the panel) stays and the others give way.
func TestKeyBarKeepsTheCloseKey(t *testing.T) {
	keys := [][2]string{{"A", "add"}, {"X", "remove"}, {"O", "optimize"}, {"Esc", "close"}}
	for _, w := range []int{60, 40, 30, 16} {
		g := newXGrid(w, 3)
		expKeyBar(g, expLayout{w: w, h: 3, footY: 1}, keys, "")
		line := strings.Split(g.String(), "\n")[1]
		if !strings.Contains(line, " Esc  close ") || len(g.keys) == 0 || g.keys[len(g.keys)-1].r.Max.X > w-2 {
			t.Errorf("at %d wide the foot reads %q with buttons %+v; want Esc close on it, inside the box", w, line, g.keys)
		}
	}
	g := newXGrid(60, 3)
	expKeyBar(g, expLayout{w: 60, h: 3, footY: 1}, keys, "")
	if len(g.keys) != 4 {
		t.Errorf("at 60 wide %d of 4 keys are buttons:\n%s", len(g.keys), g.String())
	}
}
