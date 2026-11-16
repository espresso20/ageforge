package ui

import (
	"image"
	"image/color"

	"github.com/espresso20/ageforge/window/kit"
)

// switch.go is the slide switch, for a setting that is on or off: a track
// and a knob, drawn in the same manner as the keys and the coins (a ramp of
// four tones, the light coming from the upper left). Off, the knob is at the
// left on a dull track; on, it is at the right on the track of the accent;
// held down, the knob widens; unavailable, it lies flat and dim.

const (
	// switchCols is the cells a switch plate covers.
	switchCols = 4
	// switchW and switchH are the switch in art pixels: 38 by 15 at the
	// default cell size.
	switchW, switchH = 38, 15
)

// switchKey is what a switch plate was drawn for.
type switchKey struct {
	on, down, off   bool
	cols, cw, ch    int
	track, knob     color.RGBA
	chip, edge, bak color.RGBA
}

// switchPlate is the picture of a switch.
func (s *plateSet) switchPlate(l keyLook, knob color.RGBA, on, down, off bool, back color.RGBA, cols, cw, ch int) *image.RGBA {
	s.fit(cw, ch)
	key := switchKey{on: on, down: down, off: off, cols: cols, cw: cw, ch: ch, track: l.ramp.base, knob: knob, chip: l.chip, edge: l.edge, bak: back}
	img, ok := s.sw[key]
	if !ok {
		if s.sw == nil {
			s.sw = map[switchKey]*image.RGBA{}
		}
		img = drawSwitch(key, l.ramp)
		s.sw[key] = img
		s.bytes += len(img.Pix)
	}
	return img
}

// drawSwitch draws a switch on a canvas of the plate's cells, centred.
func drawSwitch(sk switchKey, R keyRamp) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, sk.cols*sk.cw, sk.ch))
	k := max(1, sk.ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	a.fill(image.Rect(0, 0, W, H), 0, sk.bak)
	w, h := min(switchW, W-2), min(switchH, H-2)
	x0, y0 := (W-w)/2, (H-h)/2
	track := image.Rect(x0, y0, x0+w, y0+h)
	pad := 2
	kh := h - 2*pad
	kw := kh + 2
	if sk.down {
		kw += 5
	}
	knob := image.Rect(track.Min.X+pad, track.Min.Y+pad, track.Min.X+pad+kw, track.Min.Y+pad+kh)
	if sk.on {
		knob = image.Rect(track.Max.X-pad-kw, knob.Min.Y, track.Max.X-pad, knob.Max.Y)
	}
	if sk.off {
		// Unavailable: flat, checkered toward the ground, the knob a plain block.
		a.fill(track, 3, sk.edge)
		in := track.Inset(1)
		for y := in.Min.Y; y < in.Max.Y; y++ {
			for x := in.Min.X; x < in.Max.X; x++ {
				c := sk.chip
				if (x+y)&1 == 0 {
					c = toward(sk.chip, sk.bak, .5)
				}
				a.px(x, y, c)
			}
		}
		a.fill(knob, 2, sk.edge)
		return img
	}
	// The track: a rim, a face, the shade of the rim's upper left falling
	// into it and the light on its lower right.
	a.fill(track, 3, R.deep)
	face := track.Inset(1)
	a.fill(face, 2, R.dark)
	in := face.Inset(1)
	a.fill(in, 2, R.base)
	a.hline(in.Min.X+2, in.Max.X-2, in.Min.Y, R.dark)
	a.hline(in.Min.X+2, in.Max.X-2, in.Max.Y-1, R.pale)
	// The knob: a face with its light on the upper left and its shade on the
	// lower right, over a skirt.
	kr := rampOf(sk.knob)
	a.fill(image.Rect(knob.Min.X, knob.Min.Y+1, knob.Max.X, knob.Max.Y+1), 2, R.deep)
	a.fill(knob, 2, kr.deep)
	kin := knob.Inset(1)
	a.fill(kin, 1, kr.base)
	a.hline(kin.Min.X+1, kin.Max.X-1, kin.Min.Y, kr.pale)
	a.hline(kin.Min.X+1, kin.Max.X-1, kin.Max.Y-1, kr.dark)
	if sk.down {
		a.hline(kin.Min.X+1, kin.Max.X-1, kin.Min.Y, kr.dark)
	}
	return img
}

// paintSwitch paints a switch over the cells r: with pictures its plate,
// without them the chip of text the grid has there. down is the switch with
// the button held on its row.
func (k *keyPresser) paintSwitch(p *kit.Painter, pal *expPalette, r image.Rectangle, on, down, off bool) {
	cw, ch := pictureCell()
	if cw <= 0 || ch <= 0 {
		return
	}
	kind := keySteel
	if on {
		kind = keyAccent
	}
	l := pal.look(kind, keyRest)
	back := p.At(r.Min.X, r.Min.Y).Bg
	p.Picture(r, k.plates.switchPlate(l, pal.tone(xText, false), on, down, off, back, r.Dx(), cw, ch))
}
