package mapgen

import (
	"bytes"
	"image/color"
	"testing"
)

var testHue = func(h Hue) color.RGBA {
	r := spaceRGB[h]
	return color.RGBA{uint8(r[0]), uint8(r[1]), uint8(r[2]), 255}
}

func TestVoidImageRepeats(t *testing.T) {
	bg := color.RGBA{4, 6, 12, 255}
	mk := func(seed int64) *Scene {
		s := nearScene(seed, nil)
		s.Seed = seed
		return s
	}
	a := mk(7).VoidImage(60, 20, 8, 16, bg, testHue)
	b := mk(7).VoidImage(60, 20, 8, 16, bg, testHue)
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("same seed and size drew two pictures")
	}
	if c := mk(3).VoidImage(60, 20, 8, 16, bg, testHue); bytes.Equal(a.Pix, c.Pix) {
		t.Error("another seed drew the same picture")
	}
	if a.Bounds().Dx() != 480 || a.Bounds().Dy() != 320 {
		t.Errorf("size %v", a.Bounds())
	}
	lit := 0
	for i := 0; i < len(a.Pix); i += 4 {
		if a.Pix[i] > 60 || a.Pix[i+1] > 60 || a.Pix[i+2] > 60 {
			lit++
		}
	}
	if lit < 200 {
		t.Errorf("only %d bright pixels: the stars are missing", lit)
	}
}
