package mapview

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/espresso20/ageforge/world"
)

// A Face is the pixel font the labels are written in: Mask is the string
// drawn at one art pixel per font pixel, as an opaque-or-clear bitmap with its
// top left at 0, 0. The window supplies it (the font lives there).
type Face interface {
	Mask(s string) *image.Alpha
}

// A label is a name wanting a place beside an object.
type label struct {
	x, y   float64 // where the object is
	r      float64 // how far it draws, so the name clears it
	text   string
	mask   *image.Alpha
	prio   int
	chosen bool            // the selected object: always shown, in the highlight colour
	box    image.Rectangle // what the object and its ring fill (empty: derived from x, y, r)
}

const (
	prioSelected = 1000
	labelGap     = 2 // art pixels between an object and its name
	labelPad     = 2 // clear art pixels kept round a name (its keyline takes one)
)

// ranks says which kinds win a collision: Sol, then galaxies and the lit
// colonies, then bodies, then arm features, then the rest.
var ranks = [...]int{world.KindStar: 60, world.KindGalaxy: 50, world.KindColony: 45, world.KindPlanet: 40, world.KindMoon: 38,
	world.KindStation: 35, world.KindFeature: 30}

// named says which objects of a frame get a label when it is the master.
func named(f *world.Frame, o *world.Object) bool {
	if o.Label == "" {
		return false
	}
	switch f.Kind {
	case world.FrameLocalGroup:
		return o.Kind == world.KindGalaxy
	case world.FrameRegion:
		return o.Kind == world.KindFeature || o.Key == "sol"
	case world.FrameNeighbourhood:
		return o.Kind == world.KindStar || o.Kind == world.KindColony
	case world.FrameSystem:
		return o.Kind == world.KindPlanet || o.Kind == world.KindStar
	case world.FramePlanet:
		return o.Kind == world.KindMoon || o.Kind == world.KindStation || o.Kind == world.KindShip
	}
	return false
}

// boxOf is the square an object (and anything drawn round it) fills.
func boxOf(x, y, r float64) image.Rectangle { return boxXY(x, y, r, r) }

func boxXY(x, y, rx, ry float64) image.Rectangle {
	ox, oy, kx, ky := int(math.Round(x)), int(math.Round(y)), int(math.Ceil(rx)), int(math.Ceil(ry))
	return image.Rect(ox-kx, oy-ky, ox+kx+1, oy+ky+1)
}

// ringR is the radius of o's selection ring: just outside what is drawn. A
// dwarf's lit blob fills about 70% of its nominal radius, so its ring hugs that.
func ringR(p pen, o *world.Object) float64 {
	r := p.radius(o)
	if o.Galaxy != nil && o.Galaxy.Irr && r >= 1.5 {
		r *= .7
	}
	return math.Max(r, 2) + 3
}

// extent is the box an object fills on the canvas: its sprite and glow, or
// the selection ring round it.
func extent(p pen, o *world.Object, ringed bool) image.Rectangle {
	x, y := p.at(o.Pos)
	r := math.Max(p.radius(o), 1)
	if o.Kind == world.KindGalaxy && !ringed && p.radius(o) >= 1.5 { // a tilted galaxy fills an ellipse, not a square
		gm := geomOf(o)
		return boxXY(x, y, r*math.Hypot(gm.cos, gm.sin*gm.q), r*gm.ys*math.Hypot(gm.sin, gm.cos*gm.q))
	}
	if o.Body != nil && o.Body.Rings && p.f.Kind == world.FrameSystem { // Saturn's rings
		r *= 2.3
	}
	switch {
	case ringed:
		r = ringR(p, o) + 1 // the ring and its width
	case o.Kind == world.KindStar:
		r += 3 // the glow ring
	}
	return boxOf(x, y, r)
}

// place chooses which labels show and where, highest priority first. A name
// starts labelGap art pixels from its own object's box, to the right, else the
// left, above, below; the first spot that stays on the canvas and keeps
// labelPad clear of every other drawn thing (the boxes in objs, earlier names)
// wins, and a name with no such spot is dropped. It returns the top left of
// each shown label's text, in the order given.
func place(ls []label, w, h int, objs ...image.Rectangle) (shown []label, at []image.Point) {
	order := make([]int, len(ls))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return ls[order[a]].prio > ls[order[b]].prio })
	var taken []image.Rectangle
	for _, i := range order {
		l := ls[i]
		if l.mask == nil {
			continue
		}
		own := l.box
		if own.Empty() {
			own = boxOf(l.x, l.y, math.Max(l.r, 1))
		}
		mw, mh := l.mask.Bounds().Dx(), l.mask.Bounds().Dy()
		cx, cy := (own.Min.X+own.Max.X-mw)/2, (own.Min.Y+own.Max.Y-mh)/2
		for _, p := range [...]image.Point{{own.Max.X + labelGap, cy}, {own.Min.X - labelGap - mw, cy},
			{cx, own.Min.Y - labelGap - mh}, {cx, own.Max.Y + labelGap}} {
			r := image.Rectangle{Min: p, Max: p.Add(image.Pt(mw, mh))}
			if r.Min.X < 0 || r.Min.Y < 0 || r.Max.X > w || r.Max.Y > h {
				continue
			}
			pad := r.Inset(-labelPad)
			clash := false
			for _, t := range taken {
				clash = clash || pad.Overlaps(t)
			}
			for _, o := range objs {
				clash = clash || (o != own && pad.Overlaps(o))
			}
			if clash {
				continue
			}
			taken = append(taken, pad)
			shown, at = append(shown, l), append(at, r.Min)
			break
		}
	}
	return shown, at
}

// overlay draws the selection ring and the names over the finished canvas.
func (v *View) overlay(img *image.RGBA, off world.Vec) {
	if v.face == nil || v.names == nil {
		return
	}
	var ls []label
	var objs []image.Rectangle
	var picked string
	add := func(i int, o *world.Object, chosen bool) {
		text := v.names(o.Label)
		if text == "" {
			text = v.names(o.Inspect)
		}
		if text == "" || (!chosen && text == picked) {
			return // the same name twice (a star seen from two frames) reads once
		}
		p := v.penFor(img, i, off)
		x, y := p.at(o.Pos)
		prio := ranks[o.Kind]
		if chosen {
			prio = prioSelected
		}
		ls = append(ls, label{x: x, y: y, r: p.radius(o), box: extent(p, o, chosen), text: text, mask: v.face.Mask(text), prio: prio, chosen: chosen})
	}
	sel := v.objAt(v.sel.fi, v.sel.key)
	if sel != nil {
		p := v.penFor(img, v.sel.fi, off)
		x, y := p.at(sel.Pos)
		p.ring2(x, y, ringR(p, sel), v.pal.Mix(world.ToneWhite, 255))
		add(v.sel.fi, sel, true)
		if len(ls) > 0 {
			picked = ls[0].text
		}
	}
	f := v.ch[v.m].f
	p := v.penFor(img, v.m, off)
	for j := range f.Objects {
		o := &f.Objects[j]
		if named(f, o) && !(sel == o) {
			add(v.m, o, false)
		}
		switch o.Kind {
		case world.KindStar, world.KindPlanet, world.KindMoon, world.KindGalaxy, world.KindColony, world.KindStation, world.KindShip:
		default:
			continue // orbits, belts and washes are not obstacles
		}
		if o == sel {
			continue // the selected one comes in with its ring
		}
		objs = append(objs, extent(p, o, false))
	}
	if sel != nil && v.sel.fi == v.m {
		objs = append(objs, extent(p, sel, true))
	}
	shown, at := place(ls, img.Bounds().Dx(), img.Bounds().Dy(), objs...)
	for i, l := range shown {
		fg := v.pal.Mix(world.ToneWhite, 255)
		if l.chosen {
			fg = v.pal.Mix(world.ToneSun, 255)
		}
		blit(img, l.mask, at[i], v.pal.Bg, 1) // a one pixel keyline so it reads over the sky
		blit(img, l.mask, at[i], fg, 0)
	}
}

// blit paints the mask's set pixels in c; grow > 0 paints each one's
// neighbours too, for the keyline.
func blit(img *image.RGBA, m *image.Alpha, at image.Point, c color.RGBA, grow int) {
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if m.AlphaAt(x, y).A < 128 {
				continue
			}
			for dy := -grow; dy <= grow; dy++ {
				for dx := -grow; dx <= grow; dx++ {
					px, py := at.X+x-b.Min.X+dx, at.Y+y-b.Min.Y+dy
					if image.Pt(px, py).In(img.Rect) {
						i := img.PixOffset(px, py)
						img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, 255
					}
				}
			}
		}
	}
}
