// Package sprites holds the 16x16 pixel canvas the wonder icons are drawn on.
package sprites

import "math"

// Canvas is a 16x16 pixel art canvas with packed RGB pixels (0xRRGGBB; 0 is
// transparent).
type Canvas struct {
	Pixels [16][16]uint32
}

// Set draws a single pixel. A pixel outside the canvas is dropped.
func (c *Canvas) Set(x, y int, col uint32) {
	if x >= 0 && x < 16 && y >= 0 && y < 16 {
		c.Pixels[y][x] = col
	}
}

// FillRect fills a rectangle from its top-left corner (x, y), clipped to the
// canvas.
func (c *Canvas) FillRect(x, y, w, h int, col uint32) {
	for row := 0; row < h; row++ {
		for column := 0; column < w; column++ {
			c.Set(x+column, y+row, col)
		}
	}
}

// Hline draws a horizontal line of w pixels.
func (c *Canvas) Hline(x, y, w int, col uint32) {
	c.FillRect(x, y, w, 1, col)
}

// Vline draws a vertical line of h pixels.
func (c *Canvas) Vline(x, y, h int, col uint32) {
	c.FillRect(x, y, 1, h, col)
}

// Dome draws the upper half of a circle centered on (cx, cy) with radius r,
// its flat side on row cy.
func (c *Canvas) Dome(cx, cy, r int, col uint32) {
	for dy := -r; dy <= 0; dy++ {
		halfWidth := int(math.Sqrt(float64(r*r - dy*dy)))
		c.Hline(cx-halfWidth, cy+dy, halfWidth*2+1, col)
	}
}
