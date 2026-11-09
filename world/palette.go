package world

import "image/color"

// A Palette is the sky's colours for one theme, resolved to plain RGBA by the
// ui when it publishes a world, so a renderer never reads a theme: the void
// and one colour for each Tone.
type Palette struct {
	Bg    color.RGBA
	Tones [NumTones]color.RGBA
}

// Of is shade s over the void: the tone at full weight is the tone's colour,
// less lets the void show through.
func (p *Palette) Of(s Shade) color.RGBA { return p.Mix(s.Tone, s.W) }

// Mix is tone t at weight w (255 is the colour itself) over the void.
func (p *Palette) Mix(t Tone, w uint8) color.RGBA {
	if t >= NumTones {
		return p.Bg
	}
	c, k := p.Tones[t], int(w)
	mix := func(a, b uint8) uint8 { return uint8((int(a)*(255-k) + int(b)*k + 127) / 255) }
	return color.RGBA{R: mix(p.Bg.R, c.R), G: mix(p.Bg.G, c.G), B: mix(p.Bg.B, c.B), A: 255}
}

// Hash is a number that changes when any colour does.
func (p *Palette) Hash() uint64 {
	h := uint64(0xcbf29ce484222325)
	add := func(c color.RGBA) {
		h = (h ^ (uint64(c.R)<<16 | uint64(c.G)<<8 | uint64(c.B))) * 0x100000001b3
	}
	add(p.Bg)
	for _, c := range p.Tones {
		add(c)
	}
	return h
}
