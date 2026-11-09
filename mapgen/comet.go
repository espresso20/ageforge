package mapgen

import "github.com/espresso20/ageforge/world"

// The comet crosses the cosmic view: it comes in at one side, takes minutes to
// reach the other on a gentle curve, and then the sky is quiet for a minute or
// two. Its head is a white speck with a wide, soft dust tail fanning away from
// the Sun and a thin blue ion tail beside it. One at a time, never faster than
// a pixel a frame.
const (
	cometCycle = world.CometCycle // frames in one crossing and its gap (about 5.8 minutes)
	cometTail  = 34               // the tail, in pixels
)

// cometPath is the head of the comet at frame anim over a canvas w by h
// pixels, in 16ths of a pixel; ok is false while the sky is quiet. The path
// and the timing come from the seed and the crossing's number.
func cometPath(seed int64, anim, w, h int) (x, y int, ok bool) {
	n := int64(anim / cometCycle)
	ph := anim % cometCycle
	dur := 1800 + roll(300, seed, 5500, n)
	start := 240 + roll(cometCycle-dur-480, seed, 5501, n)
	if ph < start || ph >= start+dur {
		return 0, 0, false
	}
	m := cometTail + 4
	span := (w + 2*m) * pxUnit
	along := (ph - start) * span / dur
	if roll(2, seed, 5502, n) == 1 {
		along = span - along
	}
	u := (ph - start) * 1024 / dur
	hh := h * pxUnit
	y0, y1 := hh*(15+roll(35, seed, 5503, n))/100, hh*(25+roll(55, seed, 5504, n))/100
	bow := (roll(21, seed, 5505, n) - 10) * hh / 100
	return along - m*pxUnit, y0 + (y1-y0)*u/1024 + bow*sin4096(2*u)/1024, true
}

// comet draws the comet, if one is crossing.
func (r *orrery) comet() {
	hx, hy, ok := cometPath(r.seed, r.s.Anim, r.pix.W, r.pix.H)
	if !ok {
		return
	}
	own := r.pix.Owner(Thing{Kind: KindTerrain, Key: "comet"}, 5)
	dx, dy := hx-r.sx, hy-r.sy
	l := max(1, isqrt(dx*dx+dy*dy))
	ux, uy := dx*1024/l, dy*1024/l // away from the Sun
	// at s half-pixels back from the head, o sixteenths across
	spot := func(s, o int) (int, int) {
		t := s * pxUnit / 2
		return floorDiv(hx+(ux*t-uy*o)/1024, pxUnit), floorDiv(hy+(uy*t+ux*o)/1024, pxUnit)
	}
	T := 2 * cometTail
	for s := T; s >= 1; s-- { // the dust tail, wider and fainter toward its end
		hw := pxUnit + s*4*pxUnit/T
		base := 170 * (T - s) / T
		for o := -hw; o <= hw; o += pxUnit / 2 {
			x, y := spot(s, o)
			r.pix.Set(x, y, P(HCyan, base*(hw-abs(o))/hw), own)
		}
	}
	T2 := T * 14 / 10
	for s := T2; s >= 1; s-- { // the ion tail, thin, blue, a little to one side
		x, y := spot(s, 2*pxUnit+s*pxUnit/24)
		r.pix.Set(x, y, P(HCool, 100*(T2-s)/T2), own)
	}
	r.pix.Glow(hx, hy, 22, 40, HCyan, 120, 2, own)
	r.pix.Disc(hx, hy, 22, own, func(nx, ny int) Paint { return P(HWhite, 255) })
}
