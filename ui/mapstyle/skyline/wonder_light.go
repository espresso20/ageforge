package skyline

import (
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// wonder_light.go is how a landmark shows at night: its outline picks up the
// light of its age, a faint glow stands round its silhouette and its lamps
// throw their beams.

// wonderGlow is the colour a landmark's light is: lamplight in the early
// ages, the pale white of the modern ones, neon in the cyberpunk and the cold
// cyan of the cosmic.
func (s *scene) wonderGlow(lv *lotView, row int) tcell.Color {
	h := theme.SkyLamp
	switch {
	case lv.fam >= 9:
		h = theme.SkyWinNeon
	case lv.fam == 8:
		h = theme.SkyNeonCyan
	case lv.fam == 7:
		h = theme.SkyWinPale
	}
	return s.p.hue(h, mEmit, row)
}

// rimShade is an outline cell's colour: darker than its face by day so the
// silhouette parts from its neighbours, lit by the landmark's own light at night.
func (s *scene) rimShade(c tcell.Color, lit bool, glow tcell.Color, k float64) tcell.Color {
	if lit {
		return theme.Mix(theme.Shade(c, outlineK), glow, 0.62*k)
	}
	return theme.Shade(c, outlineK)
}

// wonderLight draws the night glow round a landmark and its lamps' beams.
// x0 and y0 are the sprite's top left in world-frame cells.
func (s *scene) wonderLight(lv *lotView, spr *sprite, x0, y0 int, glow tcell.Color, k float64) {
	d := lv.depth
	for _, h := range spr.haloCells() {
		s.fb.tint(x0+h.x, s.Y(y0+h.y), glow, 0.17*k, d)
	}
	seed := int(lv.ml.Seed)
	for i, b := range spr.beams {
		bx, by := x0+b.x, y0+b.y
		t := float64(s.anim)
		ph := hashf(seed, i, 3) * 6.28
		switch b.kind {
		case beamFan:
			// a turning lamp: the beam heads out to one side, shortens as it comes
			// round to face us and goes out the other
			c := mapmodel.Cos((t/9 + ph) / (2 * math.Pi))
			n := int(float64(b.n) * math.Abs(c))
			dir := 1
			if c < 0 {
				dir = -1
			}
			for j := 1; j <= n; j++ {
				a := 0.5 * k * (1 - float64(j)/float64(b.n+1))
				half := j / 7
				for dy := -half; dy <= half; dy++ {
					aa := a
					if dy != 0 {
						aa *= 0.55
					}
					s.fb.tint(bx+dir*j, s.Y(by+dy), glow, aa, d)
				}
			}
		case beamDown:
			// the same turning lamp, its light falling away toward the roofs
			c := mapmodel.Cos((t/10 + ph) / (2 * math.Pi))
			n := int(float64(b.n) * math.Abs(c))
			dir := 1
			if c < 0 {
				dir = -1
			}
			for j := 1; j <= n; j++ {
				a := 0.5 * k * (1 - float64(j)/float64(b.n+1))
				y := by + j*2/5
				for dy := 0; dy <= 1+j/8; dy++ {
					aa := a
					if dy > 0 {
						aa *= 0.55
					}
					s.fb.tint(bx+dir*j, s.Y(y+dy), glow, aa, d)
				}
			}
		case beamSky:
			// a searchlight swinging over the roofs
			ang := 0.55 * mapmodel.Sin((t/13+ph)/(2*math.Pi))
			for j := 1; j <= b.n; j++ {
				a := 0.42 * k * (1 - float64(j)/float64(b.n+1))
				x := bx + int(math.Round(float64(j)*mapmodel.Sin(ang/(2*math.Pi))/mapmodel.Cos(ang/(2*math.Pi))))
				s.fb.tint(x, s.Y(by-j), glow, a, d)
				s.fb.tint(x+1, s.Y(by-j), glow, a*0.5, d)
			}
		case beamPillar:
			pulse := 0.75 + 0.25*mapmodel.Sin((t/6+ph)/(2*math.Pi))
			for j := 1; j <= b.n; j++ {
				a := 0.55 * k * pulse * (1 - float64(j)/float64(b.n+1))
				s.fb.tint(bx, s.Y(by-j), glow, a, d)
				s.fb.tint(bx-1, s.Y(by-j), glow, a*0.4, d)
				s.fb.tint(bx+1, s.Y(by-j), glow, a*0.4, d)
			}
		}
	}
}

// plaqueReach is how far (in world columns, past the landmark's own width)
// the inspect cursor may stand for the landmark to show its plaque.
const plaqueReach = 10

// plaques puts a small plaque with its name on the ground under every
// landmark the inspect cursor is near: on a lot within plaqueReach columns of
// it, but not on the landmark itself (the cursor labels that one). A wonder
// still going up also says how much of it is banked. It is drawn on the
// screen like the other labels, over the ground and its traffic but under the
// cursor, the weather and the tags.
func (s *scene) plaques() {
	if !s.v.inspect || s.sel < 0 || s.sel >= len(s.lay.lots) {
		return
	}
	cur := &s.lay.lots[s.sel]
	var gold tcell.Color
	row := s.sy(s.groundY + 1)                                                          // the road: the cursor's own bar is the row under it
	mark := s.wsx(cur.x0-s.cam) + (s.wsx(cur.x0-s.cam+cur.spr.w)-s.wsx(cur.x0-s.cam))/2 // the cursor's pointer
	draw := func(lv *lotView, label string) {
		reach := lv.spr.w/2 + plaqueReach
		if abs(cur.ml.X-lv.ml.X) > reach || cur == lv {
			return
		}
		if gold == 0 {
			gold = s.p.hue(theme.SkyWonderGold, mEmit, 0)
		}
		text := " " + label + " "
		n := textLen(text)
		x := clampInt(s.wsx(lv.ml.X-s.cam)-n/2, 0, max(0, s.W-n))
		if mark >= x-1 && mark <= x+n { // keep clear of the cursor's pointer
			if lv.ml.X >= cur.ml.X {
				x = min(mark+2, max(0, s.W-n))
			} else {
				x = max(0, mark-n-1)
			}
		}
		for i, r := range []rune(text) {
			c := s.fb.at(x+i, s.Y(row))
			if c == nil || c.d < dTag {
				continue // the cursor, the weather and the tags stay on top
			}
			bg := theme.Shade(c.show(), 0.45)
			s.fb.set(x+i, s.Y(row), r, theme.Legible(gold, bg, 4), bg, dTag)
		}
	}
	for _, li := range s.vis {
		if lv := &s.lay.lots[li]; lv.ml.Wonder {
			draw(lv, s.m.Catalog.Defs[lv.ml.Key].Name)
		}
	}
	for i := range s.lay.sites {
		lv := &s.lay.sites[i]
		if lv.x0 >= s.cam+s.lw+12 || lv.x0+lv.spr.w <= s.cam-12 {
			continue
		}
		pct := 0
		for _, w := range s.m.Wonders {
			if w.Key == lv.ml.Key {
				pct = int(w.Progress*100 + 0.5)
			}
		}
		draw(lv, s.m.Catalog.Defs[lv.ml.Key].Name+" "+strconv.Itoa(pct)+"%")
	}
}

// natural reports a slot that is rock or growth: a landmark's night outline
// follows its built edges, not the hill under it.
func natural(sl slot) bool {
	switch sl {
	case sRock, sRockDark, sLeaf, sLeafDark, sTrunk, sWater:
		return true
	}
	return false
}
