package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// fakeWindow is a window host that counts what the Settings page asks of it.
type fakeWindow struct {
	size, zooms, toggles int
	full                 bool
}

func (w *fakeWindow) FontChoice() string   { return "mono" }
func (w *fakeWindow) NextFont()            {}
func (w *fakeWindow) CellSize() (int, int) { return 12, 24 }
func (w *fakeWindow) TextSize() int        { return w.size }
func (w *fakeWindow) ZoomBy(d int)         { w.size += d; w.zooms++ }
func (w *fakeWindow) Fullscreen() bool     { return w.full }
func (w *fakeWindow) ToggleFullscreen()    { w.full = !w.full; w.toggles++ }

// withWindow gives the rig's menu a window host, with pictures on, for the
// length of the test.
func withWindow(t *testing.T, r *menuRig) *fakeWindow {
	t.Helper()
	w := &fakeWindow{size: 16}
	r.m.window = w
	was := pictureCell
	pictureCell = w.CellSize
	t.Cleanup(func() { pictureCell = was })
	return w
}

// paint is the menu painted on a buffer of w by h cells.
func (r *menuRig) paint(w, h int) *kit.Buffer {
	r.t.Helper()
	r.m.SetRect(0, 0, w, h)
	root := kit.NewRoot(r.m)
	root.Resize(w, h)
	b := kit.NewBuffer(w, h)
	root.Paint(b)
	return b
}

// entryRect is the cells of the entry with id on the page last drawn.
func (r *menuRig) entryRect(id string) image.Rectangle {
	r.t.Helper()
	for i, it := range r.m.view.items {
		if it.id == id {
			return r.m.scene.hits[i]
		}
	}
	r.t.Fatalf("no entry %q", id)
	return image.Rectangle{}
}

// click presses and releases the left button at pt on the menu and lets the
// key it pressed act.
func (r *menuRig) click(pt image.Point) {
	r.t.Helper()
	r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: pt.X, Y: pt.Y})
	r.m.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y})
	r.m.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y})
	pump(r.app)
	r.release()
}

func TestSettingsRowsForTheWindowAreOnlyThereInIt(t *testing.T) {
	r := newMenuRig(t)
	r.press('s')
	if got := ids(r.m.view.items); strings.Contains(got, miSetSize) || strings.Contains(got, miSetFull) {
		t.Errorf("without a window the Settings page has %q", got)
	}
	r.press('b')
	withWindow(t, r)
	r.press('s')
	if got := ids(r.m.view.items); !strings.Contains(got, miSetSize) || !strings.Contains(got, miSetFull) {
		t.Errorf("in a window the Settings page has %q", got)
	}
}

func TestTextSizeAndFullscreenGoThroughTheWindowsOwnCalls(t *testing.T) {
	r := newMenuRig(t)
	win := withWindow(t, r)
	r.press('s')
	r.draw(120, 40)
	// The keys on the row: plus and minus, as typed or as clicked.
	r.press('+')
	if win.size != 17 || win.zooms != 1 {
		t.Errorf("plus: size %d after %d calls", win.size, win.zooms)
	}
	r.press('-')
	r.press('-')
	if win.size != 15 || win.zooms != 3 {
		t.Errorf("minus: size %d after %d calls", win.size, win.zooms)
	}
	r.draw(120, 40)
	var plus, minus image.Rectangle
	for _, k := range r.m.scene.keys.keys {
		switch k.key.Rune {
		case '+':
			plus = k.r
		case '-':
			minus = k.r
		}
	}
	if plus.Empty() || minus.Empty() {
		t.Fatal("the text size row shows no plus and minus keys")
	}
	r.click(plus.Min.Add(image.Pt(1, 0)))
	if win.size != 16 {
		t.Errorf("a click on plus left size %d, want 16", win.size)
	}
	r.draw(120, 40)
	r.click(minus.Min.Add(image.Pt(1, 0)))
	if win.size != 15 {
		t.Errorf("a click on minus left size %d, want 15", win.size)
	}
	if got := r.entryRect(miSetSize); !got.Empty() {
		scr := r.draw(120, 40)
		if !scr.has("Text size") || !scr.has("15") || !scr.has(zoomKeysHint()) {
			t.Errorf("the text size row does not show its label, size and shortcut")
		}
	}
	// Fullscreen: Enter on the row and a click on it, both toggle.
	row := r.entryRect(miSetFull)
	r.click(row.Min.Add(image.Pt(3, 0)))
	if !win.full || win.toggles != 1 {
		t.Errorf("a click on the fullscreen row: full %v after %d calls", win.full, win.toggles)
	}
	r.press('u')
	if win.full || win.toggles != 2 {
		t.Errorf("the letter on the fullscreen row: full %v after %d calls", win.full, win.toggles)
	}
}

func TestASwitchFlipsOnAClickAndTheRowFollows(t *testing.T) {
	r := newMenuRig(t)
	withWindow(t, r)
	r.press('s')
	r.paint(120, 40)
	if len(r.m.scene.switches) == 0 {
		t.Fatal("with pictures the plain on/off rows are switches")
	}
	pics := func() *image.RGBA {
		b := r.paint(120, 40)
		sw := r.m.scene.switches
		for _, s := range sw {
			if r.m.view.items[s.item].id == miSetMotion {
				return b.Pics[b.Cells[(s.r.Min.Y+0)*b.W+s.r.Min.X].Pic-1].Img
			}
		}
		t.Fatal("no motion switch")
		return nil
	}
	before := pics()
	if !r.acct.MotionOn() {
		t.Fatal("motion should start on")
	}
	row := r.entryRect(miSetMotion)
	r.click(row.Min.Add(image.Pt(2, 0)))
	if r.acct.MotionOn() {
		t.Error("a click on the row did not turn motion off")
	}
	after := pics()
	if string(before.Pix) == string(after.Pix) {
		t.Error("the switch plate is the same after the setting changed")
	}
	// A click on the switch itself.
	var sw image.Rectangle
	for _, s := range r.m.scene.switches {
		if r.m.view.items[s.item].id == miSetMotion {
			sw = s.r
		}
	}
	r.click(sw.Min.Add(image.Pt(1, 0)))
	if !r.acct.MotionOn() {
		t.Error("a click on the switch did not turn motion back on")
	}
	// Enter on the picked row flips it too.
	r.m.view.sel = 0
	for i, it := range r.m.view.items {
		if it.id == miSetMinimap {
			r.m.view.sel = i
		}
	}
	was := r.acct.MinimapOn()
	r.key(kit.KeyEnter, 0)
	r.release()
	if r.acct.MinimapOn() == was {
		t.Error("Enter on the picked row did not flip the switch")
	}
}

func TestWithoutPicturesASwitchIsTheTextInAChip(t *testing.T) {
	r := newMenuRig(t)
	r.press('s')
	scr := r.draw(120, 40)
	if len(r.m.scene.switches) != 0 {
		t.Error("without pictures there is no switch plate")
	}
	if !scr.has(" "+settingOnOff(true)+" ") && !scr.has(" "+settingOnOff(false)+" ") {
		t.Error("without pictures the on/off text should be on the row")
	}
}

// A row under the pointer and the picked row are lit and are told apart: the
// picked row by the bar at its left edge, the one under the pointer by the
// chevron at its right, in a dark theme and in a gold one.
func TestHoverAndPickedRowsDiffer(t *testing.T) {
	for _, th := range []string{"", "daylight", "forge", "bronze", "parchment", "high_contrast"} {
		r := newMenuRig(t)
		if th != "" {
			if err := theme.SetActive(th); err != nil {
				t.Logf("no theme %q", th)
				continue
			}
		}
		r.press('s')
		r.paint(120, 40)
		pick := r.entryRect(miSetTheme) // the first row is picked
		over := r.entryRect(miSetMotion)
		r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: over.Min.X + 4, Y: over.Min.Y})
		b := r.paint(120, 40)
		at := func(x, y int) kit.Cell { return b.Cells[y*b.W+x] }
		rest := r.entryRect(miSetGlyphs)
		if at(pick.Min.X, pick.Min.Y).R != '▎' {
			t.Errorf("theme %q: the picked row has no bar at its left edge", th)
		}
		if at(over.Min.X, over.Min.Y).R == '▎' || at(rest.Min.X, rest.Min.Y).R == '▎' {
			t.Errorf("theme %q: a row that is not picked has the bar", th)
		}
		chevron := over.Max.X - menuTail + 1
		if at(chevron, over.Min.Y).R != '›' {
			t.Errorf("theme %q: the row under the pointer has no chevron", th)
		}
		if at(chevron, rest.Min.Y).R == '›' || at(chevron, pick.Min.Y).R == '›' {
			t.Errorf("theme %q: a row not under the pointer has the chevron", th)
		}
		bgOf := func(r image.Rectangle) kit.Cell { return at(r.Min.X+3, r.Min.Y) }
		if bgOf(over).Bg == bgOf(rest).Bg || bgOf(pick).Bg == bgOf(rest).Bg {
			t.Errorf("theme %q: a lit row has the ground of a row at rest", th)
		}
		if bgOf(over).Bg == bgOf(pick).Bg {
			t.Errorf("theme %q: the picked row and the one under the pointer are lit the same", th)
		}
	}
}

// A row that cannot be used is dim, shows no hover and does nothing.
func TestAnUnavailableRowDoesNothing(t *testing.T) {
	r := newMenuRig(t)
	r.press('s')
	r.paint(120, 40)
	idx := -1
	for i, it := range r.m.view.items {
		if it.id == miSetDiscord {
			idx = i
		}
	}
	if idx < 0 {
		t.Skip("no Discord row in this build")
	}
	r.m.view.items[idx].off = true
	r.paint(120, 40)
	row := r.entryRect(miSetDiscord)
	was := r.acct.DiscordOn()
	r.click(row.Min.Add(image.Pt(3, 0)))
	r.press('d')
	if r.acct.DiscordOn() != was {
		t.Error("an unavailable row changed its setting")
	}
	r.m.Mouse(kit.Mouse{Kind: kit.MouseMove, X: row.Min.X + 4, Y: row.Min.Y})
	b := r.paint(120, 40)
	if b.Cells[row.Min.Y*b.W+row.Max.X-menuTail+1].R == '›' {
		t.Error("an unavailable row shows the hover chevron")
	}
}

// A click on an entry of the main menu does what its letter does, and the
// letter's key is drawn as a cap at the row's right end.
func TestAClickOnAnEntryIsItsLetter(t *testing.T) {
	r := newMenuRig(t, "ashford")
	r.m.set.Motion = false
	scr := r.draw(120, 40)
	row := r.entryRect(miLoad)
	cap := r.m.scene.keys.keys[0].cap
	if cap.Empty() || !scr.has(" L ") {
		t.Errorf("the entries do not show their keys as caps")
	}
	r.click(row.Min.Add(image.Pt(3, 0)))
	if r.front() != loadGamePage {
		t.Errorf("a click on Load game opened %q", r.front())
	}
}

// A list row under the pointer and the picked row are lit and told apart: the
// picked one has the bar at its left, the one under the pointer the chevron.
func TestListRowsAreLitUnderThePointerAndWhenPicked(t *testing.T) {
	r := newMenuRig(t, "ashford", "brighton", "calder")
	b := newLoadGameBrowser(r.app, r.pages, r.eng, "splash", true)
	b.list.SetCurrentItem(0)
	b.root.SetRect(0, 0, 120, 40)
	root := kit.NewRoot(b.root)
	root.Resize(120, 40)
	paint := func() *kit.Buffer {
		buf := kit.NewBuffer(120, 40)
		root.Paint(buf)
		return buf
	}
	paint()
	root.Mouse(kit.Mouse{Kind: kit.MouseMove, X: 20, Y: 4})
	buf := paint()
	at := func(x, y int) kit.Cell { return buf.Cells[y*buf.W+x] }
	picked, over, rest := at(0, 2), at(0, 4), at(0, 3)
	if picked.R != '▎' || over.R == '▎' || rest.R == '▎' {
		t.Errorf("the bar: picked %q, hovered %q, rest %q", picked.R, over.R, rest.R)
	}
	if at(118, 4).R != '›' || at(118, 2).R == '›' {
		t.Error("only the row under the pointer has the chevron")
	}
	if picked.Bg == rest.Bg || over.Bg == rest.Bg || picked.Bg == over.Bg {
		t.Error("picked, hovered and resting rows are not three different grounds")
	}
}

// A danger dialog's confirming button is a danger action button, the way out
// a plain one; the keyboard starts on the way out, and both take clicks.
func TestDangerDialogButtonsAreActionButtons(t *testing.T) {
	r := newMenuRig(t, "ashford", "brighton")
	b := newLoadGameBrowser(r.app, r.pages, r.eng, "splash", true)
	r.pages.AddPage(loadGamePage, b.root, true, true)
	r.app.SetFocus(b.list)
	b.doDelete()
	var d *nativeDialog
	walkWidgets(r.pages, func(w kit.Widget) {
		if x, ok := w.(*nativeDialog); ok {
			d = x
		}
	})
	if d == nil || len(d.buttons) != 2 || !d.actions {
		t.Fatal("the delete question is not a dialog with action buttons")
	}
	if d.buttons[0].act != keyPlain || d.buttons[1].act != keyDanger {
		t.Errorf("kinds: way out %d, confirm %d", d.buttons[0].act, d.buttons[1].act)
	}
	if d.at != 0 {
		t.Error("the keyboard does not start on the way out")
	}
	root := kit.NewRoot(r.pages)
	root.Resize(120, 40)
	buf := kit.NewBuffer(120, 40)
	root.Paint(buf)
	for _, btn := range d.buttons {
		if btn.Bounds().Dy() != actionRows {
			t.Errorf("a button is %d rows, want %d", btn.Bounds().Dy(), actionRows)
		}
	}
	n := len(b.saves)
	c := d.buttons[1].Bounds().Min.Add(image.Pt(2, 1))
	root.Mouse(kit.Mouse{Kind: kit.MouseMove, X: c.X, Y: c.Y})
	root.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: c.X, Y: c.Y})
	root.Mouse(kit.Mouse{Kind: kit.MouseRelease, Button: kit.ButtonLeft, X: c.X, Y: c.Y})
	if len(b.saves) != n-1 {
		t.Errorf("a click on the danger button left %d saves of %d", len(b.saves), n)
	}
}
