package ui

import (
	"image"
	"image/color"
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
		{tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModCtrl), kit.Key{Code: kit.KeyRune, Rune: 'r', Mod: kit.ModCtrl}},
		{tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModAlt), kit.Key{Code: kit.KeyRune, Rune: 'r', Mod: kit.ModAlt}},
		// A key with no name of its own still arrives, for the screens that
		// move on at any key.
		{tcell.NewEventKey(tcell.KeyF20, 0, tcell.ModNone), kit.Key{Code: kit.KeyOther}},
	} {
		got, ok := kitKey(c.ev)
		if !ok || got != c.want {
			t.Errorf("%v came out as %+v (ok %v), want %+v", c.ev.Name(), got, ok, c.want)
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
	if _, bg, _ := at(2, 1).Style.Decompose(); bg != theme.Transparent {
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

func TestKitStyleReversesCellStyle(t *testing.T) {
	for _, attr := range []kit.Attr{0, kit.Bold, kit.Italic | kit.Dim, kit.Reverse | kit.Underline, kit.Strike | kit.Bold | kit.Underline} {
		c := &kit.Cell{R: 'x', Fg: color.RGBA{10, 20, 30, 255}, Bg: color.RGBA{40, 50, 60, 255}, Attr: attr}
		got := kitStyle(cellStyle(c))
		if got.Attr != attr {
			t.Errorf("attr %b came back as %b", attr, got.Attr)
		}
		if got.Fg != kit.RGB(10, 20, 30) || got.Bg != kit.RGB(40, 50, 60) {
			t.Errorf("colors came back as %+v", got)
		}
	}
	st := tcell.StyleDefault.Foreground(theme.Ref(theme.RoleAccent)).Background(theme.Transparent).Bold(true)
	got := kitStyle(st)
	if got.Fg != kit.Role(theme.RoleAccent) || got.Bg != kit.Clear || got.Attr != kit.Bold {
		t.Errorf("a role style came back as %+v", got)
	}
}
