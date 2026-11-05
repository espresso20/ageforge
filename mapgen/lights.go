package mapgen

// Blinking lights on the space stations: a light is one pixel, lit one frame
// in blinkPeriod, each on a phase of its own from the seed so they never blink
// together. The ships are gone; the stations are what moves the eye now.
const blinkPeriod = 8

// blinkOn says whether light i of the seed is on at frame anim.
func blinkOn(seed int64, anim, i int) bool {
	return (anim+roll(blinkPeriod, seed, 5400, int64(i)))%blinkPeriod == 0
}

// light sets pixel (x, y) in hue h when light i is on.
func (p *Pixels) light(seed int64, anim, i, x, y int, h Hue, own uint8) {
	if blinkOn(seed, anim, i) {
		p.Set(x, y, P(h, 255), own)
	}
}
