package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/coins"
	"github.com/espresso20/ageforge/window/kit"
)

// coinSlot is the cells a square piece covers in the middle of room, for
// cells cw by ch pixels, and the piece's side in pixels.
func coinSlot(room image.Rectangle, cw, ch int) (image.Rectangle, int) {
	side := max(1, min(room.Dx()*cw, room.Dy()*ch))
	w, h := max(1, (side+cw/2)/cw), max(1, (side+ch/2)/ch)
	x, y := room.Min.X+(room.Dx()-w)/2, room.Min.Y+(room.Dy()-h)/2
	return image.Rect(x, y, x+w, y+h), side
}

// paintCoin paints copy c, as the mint draws it, in the middle of room. An
// empty copy, or one of an age whose money is not drawn, leaves the room as it is.
func paintCoin(p *kit.Painter, mint *coins.Mint, c game.CoinCopy, room image.Rectangle, cw, ch int) {
	if c.Type == "" {
		return
	}
	at, side := coinSlot(room, cw, ch)
	if img := mint.Piece(c.Age, c.Type, c.Level, c.Seed, side); img != nil {
		p.Picture(at, img)
	}
}
