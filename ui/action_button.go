package ui

import (
	"image"
	"image/color"
	"time"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// action_button.go is the action button: three rows tall, for what a card, a
// dialog or a panel asks of the player (send, start, advance). It is a key
// like any other (key_bar.go holds where it is and what a click does, and
// may carry a command line); its look is drawn here. With pictures it is a
// shaded plate with a sign at the left, the label as cell text over it, and
// a small keycap at the right; without them it is three rows of cells.

// signKind is the picture at the left of an action button.
type signKind uint8

const (
	signNone signKind = iota
	signTick
	signCross
	signPlay
	signStop
)

// signArt is a sign as 13 by 13 art pixels.
var signArt = map[signKind][]string{
	signTick: {
		".............",
		".............",
		".............",
		"...........##",
		"..........##.",
		".........##..",
		".##.....##...",
		"..##...##....",
		"...##.##.....",
		"....###......",
		".............",
		".............",
		".............",
	},
	signCross: {
		".............",
		".............",
		"..##......##.",
		"...##....##..",
		"....##..##...",
		".....####....",
		"......##.....",
		".....####....",
		"....##..##...",
		"...##....##..",
		"..##......##.",
		".............",
		".............",
	},
	signPlay: {
		".............",
		".##..........",
		".####........",
		".######......",
		".########....",
		".##########..",
		".############",
		".##########..",
		".########....",
		".######......",
		".####........",
		".##..........",
		".............",
	},
	signStop: {
		".............",
		".............",
		"..#########..",
		"..#########..",
		"..#########..",
		"..#########..",
		"..#########..",
		"..#########..",
		"..#########..",
		"..#########..",
		"..#########..",
		".............",
		".............",
	},
}

const (
	// actionRows is the rows an action button takes.
	actionRows = 3
	// actGlintPoses is how many poses a glint crosses a primary button in, and
	// actGlintRest how many animation steps it then rests: about three seconds.
	actGlintPoses = 4
	actGlintRest  = 24
)

// actionSpec is an action button: its colour, its label, the key it names
// (empty for none), the command line a click runs (empty to press the key),
// its sign, and whether it is unavailable. ask, with yes, is the question a
// click on a command asks first (xkey.ask).
type actionSpec struct {
	kind       keyKind
	label, key string
	cmd        string
	ask, yes   string
	sign       signKind
	off        bool
}

// actionW is the cells an action button takes: the label, the key when it
// has one, the sign when it has one, and the padding.
func actionW(label, key string, sign signKind) int {
	left := 2
	if sign != signNone {
		left = 4
	}
	w := left + len([]rune(label))
	if key == "" {
		return w + 2
	}
	return w + 1 + len([]rune(key)) + 2 + 1
}

// action records an action button at (x, y) in the set, writes its label and
// key through put (the grid's own way of writing text) and returns the next
// x. The cells it takes are the label and the key as they read without
// pictures; paint draws the button over them.
func (s *keySet) action(x, y int, a actionSpec, put func(x, y int, text string)) int {
	left := 2
	if a.sign != signNone {
		left = 4
	}
	w := actionW(a.label, a.key, a.sign)
	mid := y + 1
	label := image.Rect(x+left, mid, x+left+len([]rune(a.label)), mid+1)
	put(label.Min.X, mid, a.label)
	b := xkey{r: image.Rect(x, y, x+w, y+actionRows), word: label, kind: a.kind, cmd: a.cmd, ask: a.ask, yes: a.yes, sign: a.sign, off: a.off}
	if a.key != "" {
		cx := label.Max.X + 1
		b.cap = image.Rect(cx, mid, cx+len([]rune(a.key))+2, mid+1)
		put(cx, mid, " "+a.key+" ")
		if parts := capParts(a.key); len(parts) > 0 {
			b.key = parts[0].key
		}
	}
	b.still = b.key == (kit.Key{}) && a.cmd == ""
	s.keys = append(s.keys, b)
	return x + w
}

// actionButton draws an action button on the grid at (x, y) and returns the
// next x.
func (g *xgrid) actionButton(x, y int, a actionSpec) int {
	g.fill(x, y, actionW(a.label, a.key, a.sign), actionRows)
	return g.keyset().action(x, y, a, func(x, y int, text string) { g.put(x, y, text, xText) })
}

// actionKey is what an action plate was drawn for: its size in cells and
// pixels, the state and the pose of the glint, the sign, where the keycap is
// and the colours that went into it.
type actionKey struct {
	st                        keyState
	cols, cw, ch              int
	sign                      signKind
	capX, capW, pose          int
	base, chip, edge, back    color.RGBA
	capBase, ink              color.RGBA
	deep, dark, pale, capPale color.RGBA
	capDeep, capDark          color.RGBA
}

// action is the picture of an action button in a state.
func (s *plateSet) action(l keyLook, st keyState, cols, cw, ch int, sign signKind, capX, capW, pose int) *image.RGBA {
	s.fit(cw, ch)
	key := actionKey{st: st, cols: cols, cw: cw, ch: ch, sign: sign, capX: capX, capW: capW, pose: pose,
		base: l.ramp.base, chip: l.chip, edge: l.edge, back: l.ground, capBase: l.cap.base, ink: l.ink,
		deep: l.ramp.deep, dark: l.ramp.dark, pale: l.ramp.pale, capPale: l.cap.pale, capDeep: l.cap.deep, capDark: l.cap.dark}
	img, ok := s.acts[key]
	if !ok {
		if s.acts == nil {
			s.acts = map[actionKey]*image.RGBA{}
		}
		img = drawAction(key, l)
		s.acts[key] = img
		s.bytes += len(img.Pix)
	}
	return img
}

// shade is the colour of the face at art pixel (x, y), u of the way from the
// upper left to the lower right: three tones with an ordered dither between
// them, the light coming from the upper left.
func shade(R keyRamp, u float64, x, y int) color.RGBA {
	tones := [3]color.RGBA{toward(R.base, R.pale, .5), R.base, toward(R.base, R.dark, .6)}
	v := u * 1.4
	i := int(v)
	if i >= 2 {
		return tones[2]
	}
	if v-float64(i) > bayer4[y&3][x&3] {
		i++
	}
	return tones[i]
}

// drawAction draws an action button: a face shaded from the upper left over
// a lip, the sign at the left and a keycap at the right. A lifted face shows
// more lip, a pressed one sinks into it, an unavailable one lies flat and
// hatched toward the ground. The canvas is the button's cells, so it shows
// pixel for pixel; art pixels are max(1, height/21) real ones.
func drawAction(pk actionKey, l keyLook) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, pk.cols*pk.cw, actionRows*pk.ch))
	k := max(1, pk.ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	R := l.ramp
	x0, x1 := 1, W-1
	signAt := func(top, bot int, ink color.RGBA, withShade bool) {
		if pk.sign == signNone {
			return
		}
		rows := signArt[pk.sign]
		sx, sy := pk.cw*7/5/k, (top+bot)/2-6
		for pass := 0; pass < 2; pass++ {
			for y, row := range rows {
				for x, c := range row {
					if c != '#' {
						continue
					}
					switch {
					case pass == 0 && withShade:
						a.px(sx+x+1, sy+y+1, toward(R.base, keyBlack, .45))
					case pass == 1:
						a.px(sx+x, sy+y, ink)
					}
				}
			}
		}
	}
	if pk.st == keyOff {
		f := image.Rect(x0, 2, x1, H-2)
		a.fill(f, 2, pk.edge)
		in := f.Inset(1)
		for y := in.Min.Y; y < in.Max.Y; y++ {
			for x := in.Min.X; x < in.Max.X; x++ {
				c := pk.chip
				if (x+y)&1 == 0 {
					c = toward(pk.chip, pk.back, .5)
				}
				a.px(x, y, c)
			}
		}
		signAt(2, H-2, toward(pk.edge, pk.back, .2), false)
		return img
	}
	top, bot := 1, H-3
	switch pk.st {
	case keyHover:
		top, bot = 0, H-4
	case keyPressed:
		top, bot = 3, H-1
	}
	face := image.Rect(x0, top, x1, bot)
	skirt := image.Rect(x0, top+6, x1, H)
	a.fill(skirt, 2, R.deep)
	a.fill(image.Rect(skirt.Min.X+1, face.Max.Y-2, skirt.Max.X-1, skirt.Max.Y-1), 0, R.dark)
	a.fill(face, 2, R.deep)
	in := face.Inset(1)
	a.fill(in, 1, R.base)
	for y := in.Min.Y; y < in.Max.Y; y++ {
		for x := in.Min.X; x < in.Max.X; x++ {
			if min(x-in.Min.X, in.Max.X-1-x)+min(y-in.Min.Y, in.Max.Y-1-y) < 1 {
				continue
			}
			u := (float64(x-in.Min.X)/float64(max(1, in.Dx()-1)) + float64(y-in.Min.Y)/float64(max(1, in.Dy()-1))) / 2
			a.px(x, y, shade(R, u, x, y))
		}
	}
	a.hline(in.Min.X+1, in.Max.X-1, in.Min.Y, R.pale)
	a.hline(in.Min.X+1, in.Max.X-1, in.Max.Y-1, R.dark)
	if pk.st == keyPressed {
		a.hline(in.Min.X+1, in.Max.X-1, in.Min.Y, R.dark)
	}
	if pk.pose > 0 && pk.st == keyRest {
		// A band of light crosses the face, slanting like the light: wide
		// and pale at its edges, nearly white at its core.
		span := in.Dx() + in.Dy()
		cx := pk.pose * span / (actGlintPoses + 1)
		for y := in.Min.Y; y < in.Max.Y; y++ {
			for x := in.Min.X; x < in.Max.X; x++ {
				d := (x - in.Min.X) + (in.Max.Y - 1 - y) - cx
				if d < 0 {
					d = -d
				}
				switch {
				case d <= 3*k:
					a.px(x, y, toward(img.RGBAAt(x*k, y*k), keyWhite, .8))
				case d <= 9*k:
					a.px(x, y, toward(img.RGBAAt(x*k, y*k), keyWhite, .4))
				}
			}
		}
	}
	signAt(top, bot, l.ink, true)
	if pk.capW > 0 {
		cy := (top + bot) / 2
		c := image.Rect(pk.capX*pk.cw/k+1, cy-8, (pk.capX+pk.capW)*pk.cw/k-1, cy+9)
		a.fill(c, 1, pk.capDeep)
		ci := c.Inset(1)
		a.fill(ci, 0, pk.capBase)
		a.hline(ci.Min.X, ci.Max.X, ci.Min.Y, pk.capPale)
		a.hline(ci.Min.X, ci.Max.X, ci.Max.Y-1, pk.capDark)
	}
	return img
}

// glintPose is the pose of the glint at now: 1 to actGlintPoses while it
// crosses, 0 while it rests. wake is when the next pose begins.
func glintPose(now time.Time) (pose int, next time.Time) {
	step := now.UnixNano() / int64(mapAnimStep)
	c := int(step % (actGlintPoses + actGlintRest))
	if c < actGlintPoses {
		return c + 1, now
	}
	return 0, time.Unix(0, (step-int64(c)+actGlintPoses+actGlintRest)*int64(mapAnimStep))
}

// glint is the pose of the glint for a primary button at rest, and 0 when
// the motion setting is off. It tells the app's animation ticker that a
// redraw is due, at once while the glint crosses and at the next pose after.
func (f *keyFeel) glint() int {
	if f.motion == nil || !f.motion() {
		return 0
	}
	pose, next := glintPose(f.clock())
	f.wake.Store(next.UnixNano())
	f.glintOn.Store(true)
	return pose
}

// glinting reports whether a frame drawn now may differ from the last: the
// glint is crossing, or is due to start. A true answer is taken back until
// the next paint says a glint is still on the screen.
func (f *keyFeel) glinting() bool {
	if !f.glintOn.Load() || f.clock().UnixNano() < f.wake.Load() {
		return false
	}
	return f.glintOn.Swap(false)
}

// paintAction paints one action button in state st over the cells it took:
// with pictures its plate, the label and the key written over it; without,
// three rows of cells.
func (k *keyPresser) paintAction(p *kit.Painter, s *keySet, org image.Point, pal *expPalette, b xkey, st keyState, pix bool) {
	l := pal.look(b.kind, st)
	r := b.r.Add(org)
	var label, key []rune
	for x := b.word.Min.X; x < b.word.Max.X; x++ {
		label = append(label, s.glyph(x, b.word.Min.Y))
	}
	for x := b.cap.Min.X; x < b.cap.Max.X; x++ {
		key = append(key, s.glyph(x, b.cap.Min.Y))
	}
	ink := kit.Style{}.Foreground(theme.RGB(l.ink.R, l.ink.G, l.ink.B)).Bold(st != keyOff)
	cap := b.cap.Add(org)
	if pix {
		cw, ch := pictureCell()
		p.Fill(r, theme.RGB(l.ground.R, l.ground.G, l.ground.B))
		if cw > 0 && ch > 0 {
			pose := 0
			if b.kind == keyPrimary && st == keyRest {
				pose = k.sense().glint()
			}
			capX, capW := 0, 0
			if !b.cap.Empty() {
				capX, capW = b.cap.Min.X-b.r.Min.X, b.cap.Dx()
			}
			p.Picture(r, k.plates.action(l, st, r.Dx(), cw, ch, b.sign, capX, capW, pose))
		}
		p.Label(b.word.Min.X+org.X, b.word.Min.Y+org.Y, kit.Escape(string(label)), ink)
		if !b.cap.Empty() {
			capInk := inkOn(l.cap.base)
			if st == keyOff {
				capInk = l.ink
			}
			p.Label(cap.Min.X, cap.Min.Y, kit.Escape(string(key)), kit.Style{}.Foreground(theme.RGB(capInk.R, capInk.G, capInk.B)).Bold(st != keyOff))
		}
		return
	}
	face, fg := l.ramp.base, l.ink
	switch st {
	case keyPressed:
		face, fg = l.ink, l.ramp.base
	case keyOff:
		face = l.chip
	}
	cell := func(bg, fg color.RGBA) kit.Style {
		return ink.Foreground(theme.RGB(fg.R, fg.G, fg.B)).Background(theme.RGB(bg.R, bg.G, bg.B))
	}
	block := cell(face, fg)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p.Put(x, y, ' ', block)
		}
	}
	for i, ch := range label {
		p.Put(b.word.Min.X+org.X+i, b.word.Min.Y+org.Y, ch, block)
	}
	// The key is a chip of the opposite colours.
	chip := cell(fg, face)
	if st == keyOff {
		chip = block
	}
	for i, ch := range key {
		p.Put(cap.Min.X+i, cap.Min.Y, ch, chip)
	}
}
