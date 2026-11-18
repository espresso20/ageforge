package coins

import "math"

// stepTone is a tone that lands exactly on ramp step k (0 to 4), so a flat
// area never dithers.
func stepTone(k int) float64 { return float64(k)*0.25 + 0.02 }

// flat is a sealed point at ground height in ramp z, at ramp step k.
func flat(z, k int) sample { return sample{h: ground, z: z, tone: stepTone(k), seal: true} }

// sphere is a small ball inlaid in a groove: a lit tone, a dark tone and one
// hard highlight. ok is false outside the ball.
func sphere(x, y, cx, cy, rad float64, z, lit, dark int) (sample, bool) {
	dx, dy := x-cx, y-cy
	if math.Hypot(dx, dy) >= rad {
		return sample{}, false
	}
	if math.Hypot(dx+rad*0.4, dy+rad*0.4) < rad*0.34 {
		return flat(z, 4), true
	}
	if dx+dy > rad*0.35 {
		return flat(z, dark), true
	}
	return flat(z, lit), true
}
