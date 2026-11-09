package window

import (
	"image"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// face writes the map's names in a pixel font, one font pixel to one art pixel,
// as the bitmaps mapview blits onto its canvas.
type face struct {
	font *fontFile
	once sync.Once
	f    font.Face
	mu   sync.Mutex
	memo map[string]*image.Alpha
}

func newFace(f *fontFile) *face { return &face{font: f, memo: map[string]*image.Alpha{}} }

// Mask is s drawn as an alpha bitmap with its top left at 0, 0; nil when the
// font did not load.
func (c *face) Mask(s string) *image.Alpha {
	c.once.Do(func() {
		c.font.load()
		if c.font.sf != nil {
			c.f, _ = opentype.NewFace(c.font.sf, &opentype.FaceOptions{Size: pixelEm, DPI: 72, Hinting: font.HintingFull})
		}
	})
	if c.f == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if m, ok := c.memo[s]; ok {
		return m
	}
	m := c.draw(s)
	c.memo[s] = m
	return m
}

func (c *face) draw(s string) *image.Alpha {
	met := c.f.Metrics()
	w := font.MeasureString(c.f, s).Ceil()
	h := (met.Ascent + met.Descent).Ceil()
	m := image.NewAlpha(image.Rect(0, 0, w, h))
	d := font.Drawer{Dst: m, Src: image.Opaque, Face: c.f, Dot: fixed.Point26_6{Y: met.Ascent}}
	d.DrawString(s)
	return m
}
