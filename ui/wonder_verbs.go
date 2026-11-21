package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
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
	txtWonderBuild   = rowid.E("ui.wonder_verbs.build.01", "Build")
	txtWonderCollect = rowid.E("ui.wonder_verbs.collect.01", "Collect")
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
			g.chip(x, row, chipSpec{kind: keySteel, label: txtWonderCollect.String(), cmd: b.collectCmd, off: b.collectOff})
			g.keyset().keys[len(g.keyset().keys)-1].key = kit.Key{Code: kit.KeyRune, Rune: wonderCollectKey}
		}
	}
	if row := textRow(b.text, b.build, text.Dx()) - top; row >= 0 && row+chipRows <= text.Dy() {
		g.chip(wonderPad, row, chipSpec{kind: keyAccent, label: txtWonderBuild.String(), key: "B", cmd: b.buildCmd, off: b.buildOff})
		g.keyset().keys[len(g.keyset().keys)-1].key = kit.Key{Code: kit.KeyRune, Rune: wonderBuildKey}
	}
	return g
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
	wv.attach(om, "wonders", wv.layer)
	om.SetKeyHandler("wonders", func(ev *kit.Key) *kit.Key {
		// UI goroutine, no engine lock held (see the harbinger handler).
		if !wv.handleKey(ev) {
			return ev
		}
		refresh()
		return nil
	})
}
