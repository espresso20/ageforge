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
	key    string  // the object's key
	text   string
	mask   *image.Alpha
	prio   int
	chosen bool            // the selected object: always shown, in the highlight colour
	must   bool            // the home planet's name shows even where the craft crowd it
	box    image.Rectangle // what the object and its ring fill (empty: derived from x, y, r)
}

const (
	ringGap      = 2 // art pixels between what is drawn and its selection ring
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
	case world.FramePlanet: // the swarm of satellites is not named, one by one
		return o.Pick && (o.Kind == world.KindPlanet || o.Kind == world.KindMoon || o.Kind == world.KindStation || o.Kind == world.KindShip)
	}
	return false
}

// boxOf is the square an object (and anything drawn round it) fills.
func boxOf(x, y, r float64) image.Rectangle { return boxXY(x, y, r, r) }

func boxXY(x, y, rx, ry float64) image.Rectangle {
	ox, oy, kx, ky := int(math.Round(x)), int(math.Round(y)), int(math.Ceil(rx)), int(math.Ceil(ry))
	return image.Rect(ox-kx, oy-ky, ox+kx+1, oy+ky+1)
}

// drawnR is how far o really draws from its centre, in art pixels, glow
// aside: a system body's sprite (its rings included) is wider than its radius.
func drawnR(p pen, o *world.Object) float64 {
	r := p.radius(o)
	if o.Galaxy != nil && o.Galaxy.Irr && r >= 1.5 { // a dwarf's lit blob fills about 70% of its nominal radius
		r *= .7
	}
	if o.Kind == world.KindStar {
		return r + 3.5 // the glow
	}
	if p.f.Kind == world.FramePlanet {
		if o.Body != nil && (o.Kind == world.KindPlanet || o.Kind == world.KindMoon) {
			return float64(stepR(r)) + 1
		}
		return craftR(craftOf(o), homeScale(p.ppu))
	}
	if p.f.Kind == world.FrameSystem && o.Body != nil && o.Kind != world.KindBelt {
		sr := stepR(r)
		if o.Body.Rings && sr >= 2 {
			return math.Ceil(float64(sr)*2.3) + 1
		}
		return float64(sr) + 1
	}
	return math.Max(r, 1)
}

// ringR is the radius of o's selection ring: two art pixels outside all that
// is drawn of it, so the object shows through.
func ringR(p pen, o *world.Object) float64 { return drawnR(p, o) + ringGap }

// silhouetteBox is the box the Horsehead's silhouette fills on the canvas: its
// name belongs beside the head, not out at the edge of the whole cloud.
func silhouetteBox(p pen, o *world.Object) (image.Rectangle, bool) {
	if o.Form != world.FormDark || p.f.Kind != world.FrameRegion {
		return image.Rectangle{}, false
	}
	sb := p.art.cloudFor(o).sb
	if sb.Empty() {
		return image.Rectangle{}, false
	}
	x, y := p.at(o.Pos)
	R := math.Max(p.radius(o), 3)
	px := func(t, c, ctr float64) int { // texel edge to canvas
		return int(math.Round(c + (ctr+(t/cloudN*2-1)*hwin)*R))
	}
	return image.Rect(px(float64(sb.Min.X), x, hx), px(float64(sb.Min.Y), y, hy),
		px(float64(sb.Max.X), x, hx), px(float64(sb.Max.Y), y, hy)), true
}

// markOf is where o's selection ring is centred and how big it is: round the
// drawn object, which for the Horsehead is the silhouette, not the whole cloud.
func markOf(p pen, o *world.Object) (x, y, r float64) {
	if b, ok := silhouetteBox(p, o); ok {
		return float64(b.Min.X+b.Max.X) / 2, float64(b.Min.Y+b.Max.Y) / 2, math.Hypot(float64(b.Dx()), float64(b.Dy()))/2 + ringGap
	}
	x, y = p.at(o.Pos)
	return x, y, ringR(p, o)
}

// extent is the box an object fills on the canvas: its sprite and glow, or
// the selection ring round it.
func extent(p pen, o *world.Object, ringed bool) image.Rectangle {
	if b, ok := silhouetteBox(p, o); ok {
		if ringed {
			cx, cy, r := markOf(p, o)
			return boxOf(cx, cy, r+1)
		}
		return b
	}
	x, y := p.at(o.Pos)
	r := math.Max(p.radius(o), 1)
	if o.Kind == world.KindGalaxy && !ringed && p.radius(o) >= 1.5 { // a tilted galaxy fills an ellipse, not a square
		gm := geomOf(o)
		return boxXY(x, y, r*math.Hypot(gm.cos, gm.sin*gm.q), r*gm.ys*math.Hypot(gm.sin, gm.cos*gm.q))
	}
	if (o.Body != nil && p.f.Kind == world.FrameSystem) || p.f.Kind == world.FramePlanet { // sprites and Saturn's rings
		r = drawnR(p, o)
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
		placed, spare := false, image.Point{-1, -1}
		for _, p := range [...]image.Point{{own.Max.X + labelGap, cy}, {own.Min.X - labelGap - mw, cy},
			{cx, own.Min.Y - labelGap - mh}, {cx, own.Max.Y + labelGap}} {
			r := image.Rectangle{Min: p, Max: p.Add(image.Pt(mw, mh))}
			if r.Min.X < 0 || r.Min.Y < 0 || r.Max.X > w || r.Max.Y > h {
				continue
			}
			if spare.X < 0 {
				spare = p
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
			placed = true
			break
		}
		if !placed && (l.chosen || l.must) && spare.X >= 0 { // the selected name always shows
			shown, at = append(shown, l), append(at, spare)
			taken = append(taken, image.Rectangle{Min: spare, Max: spare.Add(image.Pt(mw, mh))}.Inset(-labelPad))
		}
	}
	return shown, at
}

// overlay draws the selection ring and the names over the finished canvas.
func (v *View) overlay(img *image.RGBA, off world.Vec) {
	shown, at := v.choose(img, off)
	for i, l := range shown {
		fg := v.pal.Mix(world.ToneWhite, 255)
		if l.chosen {
			fg = v.pal.Mix(world.ToneSun, 255)
		}
		blit(img, l.mask, at[i], v.pal.Bg, 1) // a one pixel keyline so it reads over the sky
		blit(img, l.mask, at[i], fg, 0)
	}
}

// choose draws the selection ring and chooses the names: which show and where.
// Every name sits labelGap from its own object or is dropped; the selected
// one's is never dropped, and takes the first spot on the canvas that clears
// its ring when none clears everything.
func (v *View) choose(img *image.RGBA, off world.Vec) ([]label, []image.Point) {
	if v.face == nil || v.names == nil {
		return nil, nil
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
		prio := 0 // kinds past the table (the gate, a comet) rank last
		if int(o.Kind) < len(ranks) {
			prio = ranks[o.Kind]
		}
		if chosen {
			prio = prioSelected
		}
		ls = append(ls, label{x: x, y: y, r: p.radius(o), box: extent(p, o, chosen), key: o.Key, text: text, mask: v.face.Mask(text), prio: prio, chosen: chosen, must: p.f.Kind == world.FramePlanet && o.Kind == world.KindPlanet})
	}
	sfi := v.selDrawn()
	sel := v.objAt(sfi, v.sel.key)
	if sel != nil {
		p := v.penFor(img, sfi, off)
		x, y, r := markOf(p, sel)
		p.ring2(x, y, r, v.pal.Mix(world.ToneWhite, 255))
		add(sfi, sel, true)
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
		if o.Body != nil && o.Body.Kind == "satellite" {
			continue // the swarm is unnamed specks; a name may sit over them
		}
		if o == sel {
			continue // the selected one comes in with its ring
		}
		objs = append(objs, extent(p, o, false))
	}
	if sel != nil && sfi == v.m {
		objs = append(objs, extent(p, sel, true))
	}
	return place(ls, img.Bounds().Dx(), img.Bounds().Dy(), objs...)
}

// selDrawn is the frame the selection is drawn in: its own while that frame
// is on screen, else the deeper frame that carries the same object (Earth
// picked in the System stays picked in the Planet frame).
func (v *View) selDrawn() int {
	if v.sel.fi != v.m && v.objAt(v.m, v.sel.key) != nil {
		return v.m
	}
	return v.sel.fi
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
