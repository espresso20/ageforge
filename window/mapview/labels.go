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
	chosen bool // the selected object: always shown, in the highlight colour
}

const (
	prioSelected = 1000
	labelGap     = 3 // art pixels between an object and its name
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

// place chooses which labels show and where, highest priority first; one
// that would touch an earlier one or leave the canvas is dropped. It returns
// the top left of each shown label's text, in the order given.
func place(ls []label, w, h int) (shown []label, at []image.Point) {
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
		mw, mh := l.mask.Bounds().Dx(), l.mask.Bounds().Dy()
		ox, oy := int(math.Round(l.x)), int(math.Round(l.y))
		reach := int(math.Ceil(l.r)) + labelGap
		for _, x := range [...]int{ox + reach, ox - reach - mw} { // right of it, else left
			r := image.Rect(x, oy-mh/2, x+mw, oy-mh/2+mh)
			if r.Min.X < 0 || r.Min.Y < 0 || r.Max.X > w || r.Max.Y > h {
				continue
			}
			pad := r.Inset(-labelPad)
			clash := false
			for _, t := range taken {
				clash = clash || pad.Overlaps(t)
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
		ls = append(ls, label{x: x, y: y, r: math.Max(p.radius(o), 1), text: text, mask: v.face.Mask(text), prio: prio, chosen: chosen})
	}
	sel := v.objAt(v.sel.fi, v.sel.key)
	if sel != nil {
		p := v.penFor(img, v.sel.fi, off)
		x, y := p.at(sel.Pos)
		p.ring2(x, y, math.Max(p.radius(sel), 2)+3, v.pal.Mix(world.ToneWhite, 255))
		add(v.sel.fi, sel, true)
		if len(ls) > 0 {
			picked = ls[0].text
		}
	}
	f := v.ch[v.m].f
	for j := range f.Objects {
		if o := &f.Objects[j]; named(f, o) && !(sel == o) {
			add(v.m, o, false)
		}
	}
	shown, at := place(ls, img.Bounds().Dx(), img.Bounds().Dy())
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
