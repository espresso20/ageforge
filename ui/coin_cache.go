package ui

import "image"

// coinCache is the pictures of coins, by look and pose, for the cell size
// they were struck for, and how many bytes they hold. The zero value is ready.
type coinCache struct {
	set   map[coinKey]*image.RGBA
	cell  [2]int
	bytes int
	// px, when set, is the size in pixels of a cell the coins are struck
	// for, in place of the window's (a picture made offscreen).
	px image.Point
}

// coinSetMax is how many bytes of coin pictures the panel keeps before it
// starts the set again.
const coinSetMax = 24 << 20

// image is a coin in a pose as a picture for a room of cells. Coins that
// look alike share one picture: it is kept by the look and the pose, not the
// badge, so a case of hundreds of badges holds a few dozen pictures at rest,
// and an image that has been painted is never drawn on. A change of cell
// size starts the set again, and so does a set grown too large.
func (c *coinCache) image(l coinLook, cells image.Point, pose coinPose) *image.RGBA {
	cw, ch := c.px.X, c.px.Y
	if cw == 0 || ch == 0 {
		cw, ch = cellPixels()
	}
	key := coinKey{l, cells.X * cw, cells.Y * ch, pose}
	if c.cell != [2]int{cw, ch} || c.bytes > coinSetMax {
		c.set, c.cell, c.bytes = nil, [2]int{cw, ch}, 0
	}
	img, ok := c.set[key]
	if !ok {
		if c.set == nil {
			c.set = map[coinKey]*image.RGBA{}
		}
		img = coinPicture(l, key.w, key.h, pose)
		c.set[key] = img
		c.bytes += len(img.Pix)
	}
	return img
}

// coinKey is what a coin's picture was struck for.
type coinKey struct {
	look coinLook
	w, h int
	pose coinPose
}
