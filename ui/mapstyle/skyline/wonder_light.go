package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

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
			c := math.Cos(t/9 + ph)
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
		case beamSky:
			// a searchlight swinging over the roofs
			ang := 0.55 * math.Sin(t/13+ph)
			for j := 1; j <= b.n; j++ {
				a := 0.42 * k * (1 - float64(j)/float64(b.n+1))
				x := bx + int(math.Round(float64(j)*math.Tan(ang)))
				s.fb.tint(x, s.Y(by-j), glow, a, d)
				s.fb.tint(x+1, s.Y(by-j), glow, a*0.5, d)
			}
		case beamPillar:
			pulse := 0.75 + 0.25*math.Sin(t/6+ph)
			for j := 1; j <= b.n; j++ {
				a := 0.55 * k * pulse * (1 - float64(j)/float64(b.n+1))
				s.fb.tint(bx, s.Y(by-j), glow, a, d)
				s.fb.tint(bx-1, s.Y(by-j), glow, a*0.4, d)
				s.fb.tint(bx+1, s.Y(by-j), glow, a*0.4, d)
			}
		}
	}
}
