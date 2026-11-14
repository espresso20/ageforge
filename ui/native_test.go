package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

func TestKitKeyNamesTcellsKeys(t *testing.T) {
	for _, c := range []struct {
		ev   *tcell.EventKey
		want kit.Key
	}{
		{tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), kit.Key{Code: kit.KeyRune, Rune: 'x'}},
		{tcell.NewEventKey(tcell.KeyRune, 'X', tcell.ModNone), kit.Key{Code: kit.KeyRune, Rune: 'X'}},
		{tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), kit.Key{Code: kit.KeyEnter}},
		{tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), kit.Key{Code: kit.KeyTab}},
		{tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone), kit.Key{Code: kit.KeyBacktab}},
		{tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), kit.Key{Code: kit.KeyBackspace}},
		{tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModShift), kit.Key{Code: kit.KeyPgDn, Mod: kit.ModShift}},
		{tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone), kit.Key{Code: kit.KeyF5}},
		// Ctrl with a letter, both ways tcell may hand it over.
		{tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModCtrl), kit.Key{Code: kit.KeyChord, Rune: 'r', Mod: kit.ModCtrl}},
		{tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModCtrl), kit.Key{Code: kit.KeyChord, Rune: 'r', Mod: kit.ModCtrl}},
		{tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModAlt), kit.Key{Code: kit.KeyChord, Rune: 'r', Mod: kit.ModAlt}},
		// A key with no name of its own still arrives, for the screens that
		// move on at any key.
		{tcell.NewEventKey(tcell.KeyF20, 0, tcell.ModNone), kit.Key{Code: kit.KeyOther}},
	} {
		if got := kitKey(c.ev); got != c.want {
			t.Errorf("%v came out as %+v, want %+v", c.ev.Name(), got, c.want)
		}
	}
}

// spot is a kit widget that fills its box with a letter and notes the mouse.
type spot struct {
	kit.Custom
	got []string
}

func newSpot(r rune, underline bool) *spot {
	s := &spot{}
	s.Draw = func(p *kit.Painter, box image.Rectangle) {
		st := kit.Style{Fg: kit.RGB(1, 2, 3), Bg: kit.RGB(4, 5, 6)}
		if underline {
			st.Attr = kit.Bold | kit.Underline
		}
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				p.Put(x, y, r, st)
			}
		}
		p.Fill(image.Rect(box.Min.X, box.Min.Y, box.Min.X+1, box.Min.Y+1), kit.Clear)
	}
	s.OnMouse = func(m kit.Mouse) bool {
		s.got = append(s.got, []string{"move", "press", "release", "wheel", "leave"}[m.Kind]+" "+m.Pt().String())
		return true
	}
	return s
}

func TestAHostPaintsItsTreeOnTheScreen(t *testing.T) {
	sim := tcell.NewSimulationScreen("UTF-8")
	must(t, sim.Init())
	defer sim.Fini()
	sim.SetSize(10, 4)
	scr := newNativeScreen(theme.WrapScreen(sim))
	h := newNativeHost(newSpot('a', true))
	h.SetRect(2, 1, 6, 2)
	h.Draw(scr)
	scr.Show()
	cells, w, _ := sim.GetContents()
	at := func(x, y int) tcell.SimCell { return cells[y*w+x] }
	if r := at(3, 1).Runes; len(r) == 0 || r[0] != 'a' {
		t.Fatalf("the host's cell holds %q", r)
	}
	fg, bg, attr := at(3, 1).Style.Decompose()
	if fg != tcell.NewRGBColor(1, 2, 3) || bg != tcell.NewRGBColor(4, 5, 6) {
		t.Errorf("colors came through as %v on %v", fg, bg)
	}
	if attr&tcell.AttrBold == 0 || attr&tcell.AttrUnderline == 0 {
		t.Errorf("bold and underline should both reach the screen, got %b", attr)
	}
	if _, bg, _ := at(2, 1).Style.Decompose(); bg != theme.TcellTransparent {
		t.Errorf("a cell with no ground should be transparent, got %v", bg)
	}
	if r := at(0, 0).Runes; len(r) > 0 && r[0] == 'a' {
		t.Error("the host drew outside its rectangle")
	}
}

func TestTheMouseGoesToTheHostOnTop(t *testing.T) {
	sim := tcell.NewSimulationScreen("UTF-8")
	must(t, sim.Init())
	defer sim.Fini()
	sim.SetSize(20, 6)
	scr := newNativeScreen(theme.WrapScreen(sim))
	left, right := newSpot('l', false), newSpot('r', false)
	hl, hr := newNativeHost(left), newNativeHost(right)
	hl.SetRect(0, 0, 10, 6)
	hr.SetRect(10, 0, 10, 6)
	cover := tview.NewBox() // one of tview's own widgets, drawn over part of left
	cover.SetRect(0, 0, 5, 6)
	frame := func() {
		hl.Draw(scr)
		hr.Draw(scr)
		cover.Draw(scr)
		scr.Show()
	}
	frame()

	scr.mouse(kit.Mouse{Kind: kit.MouseMove, X: 2, Y: 2}) // under the cover
	if len(left.got) != 0 {
		t.Errorf("a covered cell took the mouse: %v", left.got)
	}
	scr.mouse(kit.Mouse{Kind: kit.MouseMove, X: 7, Y: 2})
	scr.mouse(kit.Mouse{Kind: kit.MouseMove, X: 12, Y: 3})
	if got := strings.Join(left.got, ", "); got != "move (7,2), leave (0,0)" {
		t.Errorf("left saw %q", got)
	}
	// A press is held by its host: the drag and the release are its own,
	// in its own cells, though the pointer is over the other host.
	scr.mouse(kit.Mouse{Kind: kit.MousePress, X: 12, Y: 3})
	scr.mouse(kit.Mouse{Kind: kit.MouseMove, X: 7, Y: 2})
	scr.mouse(kit.Mouse{Kind: kit.MouseRelease, X: 7, Y: 2})
	if got := strings.Join(right.got, ", "); got != "move (2,3), press (2,3), move (-3,2), release (-3,2)" {
		t.Errorf("right saw %q", got)
	}
	// A host that is no longer drawn takes nothing.
	left.got = nil
	hr.Draw(scr)
	scr.Show()
	scr.mouse(kit.Mouse{Kind: kit.MouseMove, X: 8, Y: 2})
	if len(left.got) != 0 {
		t.Errorf("a host off the screen took the mouse: %v", left.got)
	}
}

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

// tev is a key press as tview takes one, for the tests that hand a key to a
// primitive the way tview does.
func tev(code kit.KeyCode, r rune, mod kit.Mod) *tcell.EventKey {
	return tcellKey(kit.Key{Code: code, Rune: r, Mod: mod})
}

// A key goes to tcell's form and back unchanged, so the few places that
// still hand a key to tview lose nothing.
func TestTcellKeyIsKitKeyReversed(t *testing.T) {
	for _, k := range []kit.Key{
		{Code: kit.KeyRune, Rune: 'x'}, {Code: kit.KeyRune, Rune: 'É'},
		{Code: kit.KeyEnter}, {Code: kit.KeyEsc}, {Code: kit.KeyTab}, {Code: kit.KeyBacktab},
		{Code: kit.KeyBackspace}, {Code: kit.KeyDelete}, {Code: kit.KeyHome}, {Code: kit.KeyF12},
		{Code: kit.KeyPgDn, Mod: kit.ModShift}, {Code: kit.KeyUp, Mod: kit.ModAlt},
		{Code: kit.KeyChord, Rune: 'r', Mod: kit.ModCtrl}, {Code: kit.KeyChord, Rune: 'b', Mod: kit.ModAlt},
		{Code: kit.KeyOther},
	} {
		if got := kitKey(tcellKey(k)); got != k {
			t.Errorf("%+v came back as %+v", k, got)
		}
	}
}

// A capture that reads the kit's keys, set on a box tview owns, still stops a
// key, passes one on untouched, or puts another in its place.
func TestTcellCaptureStopsPassesAndReplaces(t *testing.T) {
	capture := tcellCapture(func(k *kit.Key) *kit.Key {
		switch {
		case k.Code == kit.KeyEsc:
			return nil
		case k.Is('a'):
			return newKey(kit.KeyHome, 0, 0)
		}
		return k
	})
	if got := capture(tev(kit.KeyEsc, 0, 0)); got != nil {
		t.Errorf("a key the capture took came out as %v", got.Name())
	}
	tab := tev(kit.KeyTab, 0, 0)
	if got := capture(tab); got != tab {
		t.Errorf("a key the capture passed on is not the same event")
	}
	if got := capture(tev(kit.KeyRune, 'a', 0)); got == nil || got.Key() != tcell.KeyHome {
		t.Errorf("a key the capture replaced did not come out as the new one")
	}
	if tcellCapture(nil) != nil {
		t.Error("no capture should stay no capture")
	}
}

// Ctrl with a letter is a chord, never the letter: a screen that reads
// characters does not see it, whichever way the terminal or the window sent it.
func TestACtrlChordIsNotALetter(t *testing.T) {
	for _, ev := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyCtrlQ, 0, tcell.ModCtrl),
		tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModCtrl),
	} {
		k := kitKey(ev)
		if k.Code == kit.KeyRune || k.Is('q') || !k.Ctrl('q') {
			t.Errorf("%v reads as %+v: want a chord that is Ctrl+q and not the letter", ev.Name(), k)
		}
	}
}

// themeColor is a color read back from a tcell screen as the theme's: a
// literal, Transparent, or no color.
func themeColor(c tcell.Color) theme.Color {
	if c == theme.TcellTransparent {
		return theme.Transparent
	}
	r, g, b := c.RGB()
	if r < 0 {
		return theme.ColorDefault
	}
	return theme.NewRGBColor(r, g, b)
}

// decompose is a style read back from a tcell screen as its parts in the
// kit's terms, for the tests that hold what was drawn against the theme.
func decompose(st tcell.Style) (fg, bg theme.Color, attr kit.Attr) {
	f, b, mask := st.Decompose()
	for _, a := range cellAttrs {
		if mask&a.tcell != 0 {
			attr |= a.kit
		}
	}
	if st.GetUnderlineStyle() != tcell.UnderlineStyleNone {
		attr |= kit.Underline
	}
	return themeColor(f), themeColor(b), attr
}

// tcellStyle is a kit style as tcell's, colors resolved.
func tcellStyle(st kit.Style) tcell.Style {
	out := tcell.StyleDefault.Foreground(theme.Tcell(st.Fg)).Background(theme.Tcell(st.Bg))
	var mask tcell.AttrMask
	for _, a := range cellAttrs {
		if st.Attr&a.kit != 0 {
			mask |= a.tcell
		}
	}
	return out.Attributes(mask).Underline(st.Attr&kit.Underline != 0)
}

// tcells is a tcell screen as the map and the weather draw on one
// (mapstyle.Cells), for the tests that still read a tcell screen back.
type tcells struct{ tcell.Screen }

func (t tcells) SetContent(x, y int, r rune, comb []rune, st kit.Style) {
	t.Screen.SetContent(x, y, r, comb, tcellStyle(st))
}

func (t tcells) GetContent(x, y int) (rune, []rune, kit.Style, int) {
	r, comb, st, w := t.Screen.GetContent(x, y)
	fg, bg, attr := decompose(st)
	return r, comb, kit.Style{Fg: fg, Bg: bg, Attr: attr}, w
}
