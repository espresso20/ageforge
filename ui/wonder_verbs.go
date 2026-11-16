package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// wonder_verbs.go is what the Wonders panel does by click: Build on the
// current age's wonder, and Collect beside its bank. Both are buttons laid
// over the panel's text, on the rows the summary keeps for them, and each is
// an xkey that carries its command line.

const (
	wonderBuildKey   = 'b'
	wonderCollectKey = 'c'
	// wonderCollectCmd banks every resource the wonder still needs, as far
	// as the town's stock goes.
	wonderCollectCmd = "wonder collect all"
)

// The words the buttons say.
var (
	txtWonderBuild   = vibe.E("ui.wonder_verbs.build.01", "Build")
	txtWonderCollect = vibe.E("ui.wonder_verbs.collect.01", "Collect")
)

// wonderBlock is the current wonder's summary and what its buttons need.
type wonderBlock struct {
	text string
	// bank and build are the byte offsets in text of the bank's heading and
	// of the rows kept for Build; -1 when the wonder has no bank to show.
	bank, build int
	// bankCol is the column the heading ends at.
	bankCol              int
	buildCmd, collectCmd string
	buildOff, collectOff bool
}

// wonderBankable says whether Collect would put anything in: the bank has
// room and the town holds some of a resource the wonder still needs.
func wonderBankable(state game.GameState, cost map[string]float64, bs game.BuildingState) bool {
	if bs.WonderBankFull {
		return false
	}
	for _, r := range bankRows(cost, bs.WonderBank) {
		if r.banked < r.need && state.Resources[r.res].Amount > 0 {
			return true
		}
	}
	return false
}

// textRow is the row the text at byte off starts on, in a view w cells wide.
func textRow(text string, off, w int) int {
	t := &kit.Text{Wrap: kit.WrapChar}
	t.SetText(safeTags(text[:off]) + "X")
	return len(t.Lines(w)) - 1
}

// wonderPad is the room the summary's lines leave before the buttons.
const wonderPad = 2

// layer is the buttons laid over the text for the part of it in view: the
// text's rectangle and its first row in view. It returns nil when the panel
// is not showing the summary.
func (wv *wondersView) layer(text image.Rectangle, top int) *xgrid {
	b := wv.blk
	if wv.gallery || b.bank < 0 || text.Dx() < 1 {
		return nil
	}
	g := newXGrid(text.Dx(), text.Dy())
	if row := textRow(b.text, b.bank, text.Dx()) - top; row >= 0 && row < text.Dy() {
		if x := b.bankCol; x+runeLen(txtWonderCollect.String())+2 <= text.Dx() {
			g.chip(x, row, txtWonderCollect.String(), keySteel, b.collectCmd, "", "", b.collectOff)
			g.keyset().keys[len(g.keyset().keys)-1].key = kit.Key{Code: kit.KeyRune, Rune: wonderCollectKey}
		}
	}
	if row := textRow(b.text, b.build, text.Dx()) - top; row >= 0 && row+actionRows <= text.Dy() {
		g.actionButton(wonderPad, row, actionSpec{kind: keyPrimary, label: txtWonderBuild.String(), key: "B", cmd: b.buildCmd, off: b.buildOff})
		g.keyset().keys[len(g.keyset().keys)-1].key = kit.Key{Code: kit.KeyRune, Rune: wonderBuildKey}
	}
	if wv.plain != nil && wv.plain() {
		g.fold()
	}
	return g
}

// over paints the buttons on the text.
func (wv *wondersView) over(p *kit.Painter, text image.Rectangle, top int) {
	wv.grid, wv.org = wv.layer(text, top), text.Min
	if wv.grid == nil {
		return
	}
	pal := newExpPalette()
	// Only the cells a button covers: the rest is the text underneath.
	for _, k := range wv.grid.keys {
		for y := k.r.Min.Y; y < k.r.Max.Y; y++ {
			for x := k.r.Min.X; x < k.r.Max.X; x++ {
				p.Put(text.Min.X+x, text.Min.Y+y, wv.grid.c[y*wv.grid.w+x].r, pal.style(wv.grid.c[y*wv.grid.w+x]))
			}
		}
	}
	wv.pr.paint(p, wv.grid, text.Min, pal)
}

// mouse gives a press or a move to the buttons.
func (wv *wondersView) mouse(m kit.Mouse) bool {
	if wv.grid == nil {
		return false
	}
	return wv.pr.mouse(wv.grid, wv.org, m)
}

// pointer is the hand over a button a click would press.
func (wv *wondersView) pointer(at image.Point) kit.Pointer {
	if wv.grid == nil {
		return kit.PointerArrow
	}
	return wv.pr.pointer(wv.grid, wv.org, at)
}

// press runs the command of the button that key stands for, as a typed key
// does, and reports whether it was one of theirs. An off button runs nothing.
func (wv *wondersView) press(ev *kit.Key) bool {
	if wv.gallery || wv.blk.bank < 0 || ev.Code != kit.KeyRune {
		return false
	}
	var cmd string
	var off bool
	switch strings.ToLower(string(ev.Rune)) {
	case string(wonderBuildKey):
		cmd, off = wv.blk.buildCmd, wv.blk.buildOff
	case string(wonderCollectKey):
		cmd, off = wv.blk.collectCmd, wv.blk.collectOff
	default:
		return false
	}
	if !off && wv.pr.run != nil {
		wv.pr.run(cmd)
	}
	return true
}

// wire makes the panel's buttons work: they paint over its text, take its
// mouse and run their commands through the app. refresh redraws the panel
// after a key.
func (wv *wondersView) wire(om *OverlayManager, refresh func()) {
	wv.pr.wire(om.app)
	wv.plain = func() bool { return om.plain != nil && om.plain() }
	om.SetOver("wonders", wv.over)
	om.SetButtons("wonders", wv.mouse, wv.pointer)
	om.SetKeyHandler("wonders", func(ev *kit.Key) *kit.Key {
		// UI goroutine, no engine lock held (see the harbinger handler).
		if !wv.handleKey(ev) {
			return ev
		}
		refresh()
		return nil
	})
}
