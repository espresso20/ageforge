package mapgen

import (
	"bytes"
	"image/color"
	"slices"
	"testing"
)

var testHue = func(h Hue) color.RGBA {
	r := spaceRGB[h]
	return color.RGBA{uint8(r[0]), uint8(r[1]), uint8(r[2]), 255}
}

func voidScene(seed int64) *Scene {
	s := nearScene(seed, nil)
	s.Seed = seed
	return s
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

// The Milky Way's arms are thin and open: out of its middle the light has dark
// gaps between the arms, not an unbroken ring. On every ray, the share of the
// way from 0.4 to 0.95 of the radius that is darker than half the arm's light
// says how open the pattern is.
func TestVoidArmsHaveGaps(t *testing.T) {
	for _, sz := range [][2]int{{60, 20}, {120, 40}, {200, 60}} {
		w, h, cw, ch := sz[0], sz[1], 10, 20
		img := voidScene(7).VoidImage(w, h, cw, ch, color.RGBA{4, 6, 12, 255}, testHue)
		W, H := 2*w, 4*h // isotropic units, as the layout counts them
		R := float64(min(W*23/100, H*36/100))
		cx, cy := float64(W*26/100*cw)/2, float64(H*46/100*ch)/4
		sx, sy := float64(cw)/2, float64(ch)/4
		const rays = 24
		var open float64
		for k := 0; k < rays; k++ {
			a := float64(k) * 4096 / rays
			var v []int
			for r := 0.4; r <= 0.95; r += 0.005 {
				c := img.RGBAAt(int(cx+r*R*sx*cosF(a)), int(cy+r*R*sy*sinF(a)))
				v = append(v, int(c.R)+int(c.G)+int(c.B))
			}
			s := slices.Clone(v)
			slices.Sort(s)
			arm := s[len(s)*98/100]
			dark := 0
			for _, x := range v {
				if x*2 < arm {
					dark++
				}
			}
			open += float64(dark) / float64(len(v)) / rays
		}
		t.Logf("%dx%d: %.0f%% of the way along a ray is dark", w, h, open*100)
		if open < 0.5 {
			t.Errorf("%dx%d: only %.0f%% of the disc between the arms is dark: it reads as a ring", w, h, open*100)
		}
	}
}
