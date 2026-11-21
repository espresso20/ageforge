package ui

import (
	"image"
	"image/color"
)

// stats_pix.go draws the Stats page's pictures in art pixels, in the manner
// of the coins: flat tones, a pale lip, a dark foot, no blur.

// fillPx fills a rectangle of real pixels.
func fillPx(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	r = r.Intersect(img.Rect)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// statsInk is the colours the page's pictures are drawn in, all from the
// panel palette: the ground, a plate raised from it, the well a numeral sits
// in, the lip of the plate and the numeral's face and shadow.
type statsInk struct {
	ground, plate, well, lip, foot, face, shadow, accent color.RGBA
}

func newStatsInk(pal *expPalette) statsInk {
	ground, edge := pal.tone(xText, true), pal.tone(xBorder, false)
	in := statsInk{ground: ground, face: pal.tone(xBright, false), accent: pal.tone(xAccent, false)}
	in.plate = toward(ground, edge, .4)
	in.well = toward(ground, edge, .1)
	in.lip = toward(in.plate, in.face, .3)
	in.foot = toward(in.plate, ground, .5)
	in.shadow = toward(in.well, ground, .6)
	if !lightGround(ground) {
		in.shadow = toward(in.well, keyBlack, .5)
	}
	return in
}

// statsTilePic is a headline tile: a plate with a well, a sign at the left
// of the well when kind is not empty and the
// label rows left below the well. tw, th and wr (the well's rows) are cells.
// A tile that holds is framed in the accent tone.
func statsTilePic(tw, th, wr, cw, ch int, kind string, hold bool, in statsInk) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, tw*cw, th*ch))
	k := max(1, ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	a.fill(image.Rect(0, 0, W, H), 2, in.foot)
	a.fill(image.Rect(0, 0, W, H-1), 2, in.plate)
	a.hline(2, W-2, 0, in.lip)
	wellH := wr * ch / k
	well := image.Rect(2, 2, W-2, wellH)
	a.fill(well, 1, in.well)
	if hold {
		for t := 0; t < 2; t++ {
			a.hline(2, W-2, t, in.accent)
			a.hline(2, W-2, H-1-t, in.accent)
			for y := 2; y < H-2; y++ {
				a.px(t, y, in.accent)
				a.px(W-1-t, y, in.accent)
			}
		}
	}
	if kind != "" {
		drawSign(a, well.Min.X+2, well.Min.Y+(well.Dy()-36)/2, kind, signTints[kind], .1, in.well)
	}
	return img
}

// statsBlockPic is one age's block of the band: a flat tone with a lip
// along the top and a foot along the bottom. An open block fades out in
// dither on the right; a hot one (under the pointer) takes an outline.
func statsBlockPic(wc, hc, cw, ch int, tone color.RGBA, open, hot bool, in statsInk) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, wc*cw, hc*ch))
	k := max(1, ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	R := rampOf(tone)
	a.fill(image.Rect(1, 0, W-1, H), 0, R.base)
	a.hline(1, W-1, 0, R.pale)
	a.hline(1, W-1, H-1, R.dark)
	a.hline(1, W-1, H-2, R.dark)
	if open {
		for x := W - 8; x < W-1; x++ {
			for y := 0; y < H; y++ {
				if (x+y)%2 == 0 && x-(W-8) > y%4+1 {
					a.px(x, y, in.ground)
				}
			}
		}
	}
	if hot {
		edge := toward(R.pale, keyWhite, .4)
		a.hline(0, W, 0, edge)
		a.hline(0, W, H-1, edge)
		for y := 0; y < H; y++ {
			a.px(0, y, edge)
			a.px(W-1, y, edge)
		}
	}
	return img
}

// statsBarPic is a ranked bar: the track with a tick at each decade, and
// the bar from its left end to frac of the way across.
func statsBarPic(wc, hc, cw, ch int, frac float64, ticks []float64, tone color.RGBA, hot bool, in statsInk) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, wc*cw, hc*ch))
	k := max(1, ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	top, bot := H/4, H-H/4
	a.fill(image.Rect(0, top, W, bot), 0, in.well)
	for _, t := range ticks {
		x := int(t*float64(W-1) + .5)
		for y := top; y < bot; y++ {
			a.px(x, y, toward(in.well, in.face, .22))
		}
	}
	if n := int(frac*float64(W) + .5); n > 0 {
		if lightGround(in.ground) {
			tone = toward(tone, keyBlack, .3)
		}
		R := rampOf(tone)
		if hot {
			R = rampOf(toward(tone, keyWhite, .25))
		}
		a.fill(image.Rect(0, top, n, bot), 0, R.base)
		a.hline(0, n, top, R.pale)
		a.hline(0, n, bot-1, R.dark)
	}
	return img
}
