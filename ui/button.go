package ui

import (
	"image"
	"image/color"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// button.go is how the key buttons look. A key is a small physical keycap and
// the word beside it is plain text. The grid holds both as cells (the look
// without pictures: a chip and plain text); at paint time each key is
// restyled for its state, and where the window shows pictures a plate is laid
// under the cap's cells and the cap's text written back over it.

// keyKind is the colour of a key: the accent for a key that acts, steel for
// one that moves, and the close button's own. The action buttons (three rows
// tall, action_button.go) are primary in the accent, plain in steel and
// danger in the alert colour.
type keyKind uint8

const (
	keyAccent keyKind = iota
	keySteel
	keyClose
	keyPrimary
	keyPlain
	keyDanger
)

// action reports whether the kind is an action button's.
func (k keyKind) action() bool { return k >= keyPrimary }

// keyState is how a key looks: at rest, under the pointer, pressed or
// unavailable.
type keyState uint8

const (
	keyRest keyState = iota
	keyHover
	keyPressed
	keyOff
)

const (
	// keySteps is how many animation steps a typed key stays sunk.
	keySteps = 2
	// keyHoldMax bounds how long a held mouse button keeps a key sunk when
	// its release never arrives (the press closed the panel).
	keyHoldMax = 2 * time.Second
	// plateSetMax is how many bytes of key plates a presser keeps before it
	// starts the set again.
	plateSetMax = 4 << 20
)

// keyFeel is what the keys typed and the mouse buttons held do to the
// buttons: the key last pressed sinks, for keySteps steps after it went down
// and while a mouse button holds it. The app has one, shared by the panels.
type keyFeel struct {
	// now is the clock; nil means time.Now (a test pins it).
	now func() time.Time
	// key is the key last pressed; end and held (Unix nanoseconds, 0 for none)
	// say until when it stays sunk. They are atomic because the animation
	// tickers ask busy from another goroutine.
	key  kit.Key
	end  atomic.Int64
	held atomic.Int64
	// rect is the button a mouse button holds, for one that names no key.
	rect image.Rectangle
	// preview is the command line of the button the pointer rests on, "" for
	// none: the command bar shows it dim while the bar is empty.
	preview string
	// motion says whether the motion setting is on (nil: off); glintOn is set
	// by a paint that drew a button's glint and wake is the time (Unix
	// nanoseconds) from which the next pose of it is due.
	motion  func() bool
	glintOn atomic.Bool
	wake    atomic.Int64
}

func (f *keyFeel) clock() time.Time {
	if f.now != nil {
		return f.now()
	}
	return time.Now()
}

// typed notes that k was pressed now. A key typed puts the command preview
// away.
func (f *keyFeel) typed(k kit.Key) {
	f.key, f.preview = k, ""
	f.end.Store(f.clock().Add(keySteps * coinStep).UnixNano())
}

// hold notes that the button at r, which presses k, was pressed with a mouse
// button, which has it down.
func (f *keyFeel) hold(k kit.Key, r image.Rectangle) {
	f.typed(k)
	f.rect = r
	f.held.Store(f.clock().Add(keyHoldMax).UnixNano())
}

// heldAt reports whether a mouse button holds the button at r down.
func (f *keyFeel) heldAt(r image.Rectangle) bool {
	return f.clock().UnixNano() < f.held.Load() && f.rect == r
}

// release lets go of a held key, and says whether there was one.
func (f *keyFeel) release() bool { return f.held.Swap(0) != 0 }

// sunk reports whether k is sunk now.
func (f *keyFeel) sunk(k kit.Key) bool {
	n := f.clock().UnixNano()
	return (n < f.end.Load() || n < f.held.Load()) && sameKey(f.key, k)
}

// busy reports whether a frame drawn now may differ from one drawn a step
// ago: a key is sunk, or was in the step before.
func (f *keyFeel) busy() bool {
	n := f.clock().UnixNano()
	return n < f.end.Load()+int64(coinStep) || n < f.held.Load()
}

// sameKey reports whether two keys are the same press: letters count as the
// lower case, and Shift does not matter.
func sameKey(a, b kit.Key) bool {
	if a.Code != b.Code {
		return false
	}
	return a.Code != kit.KeyRune || unicode.ToLower(a.Rune) == unicode.ToLower(b.Rune)
}

// keyRamp is the four tones a key is drawn in, deepest to palest.
type keyRamp struct{ deep, dark, base, pale color.RGBA }

var (
	keyBlack = color.RGBA{A: 255}
	keyWhite = color.RGBA{255, 255, 255, 255}
)

func rampOf(c color.RGBA) keyRamp {
	return keyRamp{toward(c, keyBlack, .6), toward(c, keyBlack, .32), c, toward(c, keyWhite, .38)}
}

// inkOn is the ink that reads on c.
func inkOn(c color.RGBA) color.RGBA { return rgbOf(theme.BestOn(theme.RGB(c.R, c.G, c.B))) }

// tone is a palette style's colour.
func (p *expPalette) tone(k xk, bg bool) color.RGBA {
	st := p.styles[k]
	c := st.Fg
	if bg {
		c = st.Bg
	}
	rgba, _ := c.RGBA()
	return rgba
}

// keyLook is the colours of a key in a state: the ramp of its face, the ink
// of its cap, and the colours the unavailable key is drawn in.
type keyLook struct {
	ramp keyRamp
	// cap is the ramp of the small keycap on an action button.
	cap     keyRamp
	ink     color.RGBA
	chip    color.RGBA // the unavailable key's face
	edge    color.RGBA // and its rim
	ground  color.RGBA
	word    color.RGBA // the word's ink
	wordBar color.RGBA // the bar behind the word under the pointer
}

// look is the colours of a key of kind in st. Light comes from the upper
// left: a lifted key is lighter, a sunk one darker.
func (p *expPalette) look(kind keyKind, st keyState) keyLook {
	ground, text, dim := p.tone(xText, true), p.tone(xText, false), p.tone(xDim, false)
	chip, accent := p.tone(xKeyName, true), p.tone(xKeycap, true)
	base, capBase := accent, toward(accent, keyWhite, .78)
	switch kind {
	case keySteel, keyPlain:
		base = toward(chip, text, .42)
		capBase = accent
	case keyDanger:
		base = p.tone(xBad, false)
		capBase = toward(base, keyWhite, .78)
	case keyClose:
		base = toward(chip, text, .42)
		if st != keyRest && st != keyOff {
			base = p.tone(xBad, false)
		}
	}
	r := rampOf(base)
	switch st {
	case keyHover:
		r = keyRamp{r.deep, r.dark, toward(r.base, r.pale, .45), toward(r.pale, keyWhite, .4)}
	case keyPressed:
		r = keyRamp{r.deep, r.dark, toward(r.base, r.dark, .35), r.base}
	}
	l := keyLook{ramp: r, cap: rampOf(capBase), ink: inkOn(r.base), chip: toward(chip, ground, .45), edge: toward(chip, ground, .1), ground: ground, word: text,
		wordBar: toward(ground, accent, .16)}
	if st == keyOff {
		l.ink, l.word = toward(dim, ground, .2), toward(dim, ground, .25)
	}
	return l
}

// plateKey is what a plate was drawn for: the cap's size in cells and
// pixels, the state, and the colours that went into it.
type plateKey struct {
	st                     keyState
	cols, cw, ch           int
	base, chip, edge, back color.RGBA
}

// plateSet is the plates a presser has drawn. A plate that has been painted
// is never drawn on again: the window keeps one picture per image.
type plateSet struct {
	m     map[plateKey]*image.RGBA
	acts  map[actionKey]*image.RGBA
	bytes int
	cell  [2]int
}

// fit empties the set when the cell has changed size or the set has grown
// past its bound.
func (s *plateSet) fit(cw, ch int) {
	if s.cell != [2]int{cw, ch} || s.bytes > plateSetMax {
		s.m, s.acts, s.cell, s.bytes = nil, nil, [2]int{cw, ch}, 0
	}
}

// plate is the picture of a key cap cols cells wide in state st.
func (s *plateSet) plate(l keyLook, st keyState, cols, cw, ch int) *image.RGBA {
	s.fit(cw, ch)
	key := plateKey{st, cols, cw, ch, l.ramp.base, l.chip, l.edge, l.ground}
	img, ok := s.m[key]
	if !ok {
		if s.m == nil {
			s.m = map[plateKey]*image.RGBA{}
		}
		img = drawPlate(key, l.ramp)
		s.m[key] = img
		s.bytes += len(img.Pix)
	}
	return img
}

// art draws in art pixels, each a k by k block of real ones.
type art struct {
	img *image.RGBA
	k   int
}

func (a art) px(x, y int, c color.RGBA) {
	for dy := 0; dy < a.k; dy++ {
		for dx := 0; dx < a.k; dx++ {
			if p := image.Pt(x*a.k+dx, y*a.k+dy); p.In(a.img.Rect) {
				a.img.SetRGBA(p.X, p.Y, c)
			}
		}
	}
}

func (a art) hline(x0, x1, y int, c color.RGBA) {
	for x := x0; x < x1; x++ {
		a.px(x, y, c)
	}
}

// fill fills r with c, cutting each corner n steps (n 0: square).
func (a art) fill(r image.Rectangle, n int, c color.RGBA) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if n > 0 && min(x-r.Min.X, r.Max.X-1-x)+min(y-r.Min.Y, r.Max.Y-1-y) < n {
				continue
			}
			a.px(x, y, c)
		}
	}
}

// drawPlate draws a key: a face with chamfered corners over a skirt. A
// lifted key shows more skirt, a pressed one sinks into it, and an
// unavailable one lies flat and checkered toward the ground. The canvas is
// the cap's cells, so it shows pixel for pixel; art pixels are
// max(1, height/21) real ones.
func drawPlate(pk plateKey, R keyRamp) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, pk.cols*pk.cw, pk.ch))
	k := max(1, pk.ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	x0, x1 := 1, W-1 // a pixel of air each side keeps neighbours apart
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
	a.hline(in.Min.X+1, in.Max.X-1, in.Min.Y, R.pale)
	a.hline(in.Min.X+1, in.Max.X-1, in.Max.Y-1, R.dark)
	if pk.st == keyPressed {
		a.hline(in.Min.X+1, in.Max.X-1, in.Min.Y, R.dark)
	}
	return img
}

// paint restyles the keys h drew, over the grid already painted at org:
// each key in the state it is in. A key under the pointer rises and its word
// is lit, a key being pressed (by the mouse, or typed) sinks, and one marked
// off is dull. Where the window shows pictures a plate goes under the cap and
// its text is written over; without them the cap is restyled as a chip.
func (k *keyPresser) paint(p *kit.Painter, h keyHolder, org image.Point, pal *expPalette) {
	s := heldKeys(h)
	if s == nil {
		return
	}
	hot, isHot := k.lit(s)
	pix := pictureCell != nil && !s.plain
	for i, key := range s.keys {
		if key.kind.action() {
			if !key.word.Empty() {
				st := keyRest
				switch {
				case key.off:
					st = keyOff
				case k.press != nil && k.sunkButton(key):
					st = keyPressed
				case k.press != nil && isHot && hot.r == key.r:
					st = keyHover
				}
				k.paintAction(p, s, org, pal, key, st, pix)
			}
			continue
		}
		if key.cap.Empty() {
			continue
		}
		st, first := keyRest, true
		for j, o := range s.keys {
			if o.cap != key.cap {
				continue
			}
			if j < i {
				first = false
				break
			}
			switch {
			case o.off:
				st = keyOff
			case k.press != nil && st != keyOff && !o.still && k.sense().sunk(o.key):
				st = keyPressed
			case k.press != nil && st == keyRest && isHot && hot.cap == o.cap:
				st = keyHover
			}
		}
		if !first {
			continue
		}
		k.paintKey(p, s, org, pal, key, st, pix)
	}
}

// sunkButton reports whether b is sunk: its key was typed, or a mouse button
// holds it.
func (k *keyPresser) sunkButton(b xkey) bool {
	f := k.sense()
	return b.key != (kit.Key{}) && f.sunk(b.key) || f.heldAt(b.r)
}

// paintKey paints one key and its word in state st.
func (k *keyPresser) paintKey(p *kit.Painter, g *keySet, org image.Point, pal *expPalette, key xkey, st keyState, pix bool) {
	l := pal.look(key.kind, st)
	cap := key.cap.Add(org)
	var label []rune
	for x := key.cap.Min.X; x < key.cap.Max.X; x++ {
		label = append(label, g.glyph(x, key.cap.Min.Y))
	}
	// Under the pointer the close key widens over its whole slot and names
	// the key that does the same.
	if name := []rune(" " + escClose()[0] + " "); key.kind == keyClose && (st == keyHover || st == keyPressed) && len(name) == key.r.Dx() {
		cap, label = key.r.Add(org), name
	}
	ink := kit.Style{}.Foreground(theme.RGB(l.ink.R, l.ink.G, l.ink.B)).Bold(st != keyOff)
	if pix {
		cw, ch := pictureCell()
		p.Fill(cap, theme.RGB(l.ground.R, l.ground.G, l.ground.B))
		if cw > 0 && ch > 0 {
			p.Picture(cap, k.plates.plate(l, st, cap.Dx(), cw, ch))
		}
		p.Label(cap.Min.X, cap.Min.Y, kit.Escape(string(label)), ink)
	} else {
		face, fg := l.ramp.base, l.ink
		switch st {
		case keyPressed:
			face, fg = l.ink, l.ramp.base
		case keyOff:
			face = l.chip
		}
		cell := ink.Foreground(theme.RGB(fg.R, fg.G, fg.B)).Background(theme.RGB(face.R, face.G, face.B))
		for i, r := range label {
			p.Put(cap.Min.X+i, cap.Min.Y, r, cell)
		}
	}
	// The word is plain text: at rest as the grid has it, under the pointer on
	// a lit bar, unavailable dim.
	if st == keyRest || key.word.Empty() {
		return
	}
	bar := l.ground
	if st == keyHover || st == keyPressed {
		bar = l.wordBar
	}
	word := kit.Style{}.Foreground(theme.RGB(l.word.R, l.word.G, l.word.B)).Background(theme.RGB(bar.R, bar.G, bar.B))
	for x := key.word.Min.X; x < key.word.Max.X; x++ {
		p.Put(org.X+x, org.Y+key.word.Min.Y, g.glyph(x, key.word.Min.Y), word)
	}
}
