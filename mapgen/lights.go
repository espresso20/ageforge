package mapgen

// Blinking lights on the space stations: a light is one pixel, lit blinkLit
// frames in blinkPeriod (a slow beacon, not a flicker), each on a phase of its
// own from the seed so they never blink together.
const (
	blinkPeriod = 28
	blinkLit    = 3
)

// blinkOn says whether light i of the seed is on at frame anim.
func blinkOn(seed int64, anim, i int) bool { return blinkWith(seed, anim, i, blinkPeriod, blinkLit) }

// blinkWith is blinkOn for a light that is lit lit frames in period.
func blinkWith(seed int64, anim, i, period, lit int) bool {
	return (anim+roll(period, seed, 5400, int64(i)))%period < lit
}

// light sets pixel (x, y) in hue h when light i is on.
func (p *Pixels) light(seed int64, anim, i, x, y int, h Hue, own uint8) {
	if blinkOn(seed, anim, i) {
		p.Set(x, y, P(h, 255), own)
	}
}
