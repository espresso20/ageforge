package ui

import (
	"image"
	"sync/atomic"
	"time"

	"github.com/espresso20/ageforge/window/kit"
)

// epoch_hall_coins.go paints the epoch page's coins over its text. The page
// leaves their room blank and says where they go (hallCoin); the owner of the
// text view paints them as pictures, in the pose their show has come to.

// hallCoins is the coins of the page on show, the pictures they are drawn
// from, and whether any of them moves.
type hallCoins struct {
	list  []hallCoin
	cache *coinCache
	// moving is set by a paint that drew a coin in a show, so the
	// dashboard knows to keep redrawing.
	moving atomic.Bool
}

// paint paints each coin whose place is on view in the text, in a room the
// size of the text view: text is the view's rectangle and top the page's
// first row in view. Coins in a show are at frame n (atRest: none move), and
// a coin starts its show after the ones before it on the page.
func (h *hallCoins) paint(p *kit.Painter, text image.Rectangle, top, n int) {
	moving := false
	if h.cache == nil {
		h.cache = &coinCache{}
	}
	p = p.Clip(text)
	org := image.Pt(text.Min.X, text.Min.Y-top)
	for _, c := range h.list {
		r := c.r.Add(org)
		if r.Intersect(text).Empty() {
			continue
		}
		at := n
		if n != atRest {
			at -= 2 * c.i
		}
		p.Picture(r, h.cache.image(c.look, r.Size(), coinPoseAt(c.look, at)))
		moving = moving || c.look.earned()
	}
	h.moving.Store(moving && n != atRest)
}

// hallFrame is the pose number the coins' shows are at: the clock's, or
// atRest when the motion setting is off.
func hallFrame(motion bool, now time.Time) int {
	if !motion {
		return atRest
	}
	// Large, so a coin's own delay never takes it to atRest.
	return 1<<40 + int(now.UnixNano()/int64(coinStep))
}
