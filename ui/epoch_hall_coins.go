package ui

import (
	"image"

	"github.com/espresso20/ageforge/pkg/coins"
	"github.com/espresso20/ageforge/window/kit"
)

// epoch_hall_coins.go paints the epoch page's coins over its text. The page
// leaves their room blank and says where they go (hallCoin); the owner of the
// text view paints them as pictures.

// hallCoins is the coins of the page on show, and the mint they are drawn from.
type hallCoins struct {
	list []hallCoin
	mint *coins.Mint
	// px, when set, is the size in pixels of a cell the coins are drawn
	// for, in place of the window's (a picture made offscreen).
	px image.Point
}

// paint paints each coin whose place is on view in the text, in a room the
// size of the text view: text is the view's rectangle and top the page's
// first row in view.
func (h *hallCoins) paint(p *kit.Painter, text image.Rectangle, top int) {
	cw, ch := h.px.X, h.px.Y
	if cw == 0 || ch == 0 {
		cw, ch = cellPixels()
	}
	p = p.Clip(text)
	org := image.Pt(text.Min.X, text.Min.Y-top)
	for _, c := range h.list {
		if r := c.r.Add(org); !r.Intersect(text).Empty() {
			paintCoin(p, h.mint, c.coin, r, cw, ch)
		}
	}
}
