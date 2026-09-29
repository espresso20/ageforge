package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

// RGB is a scene colour in 0..255 floats, so blends and lighting compose
// without rounding until the cell is written.
type RGB struct{ R, G, B float64 }

func hex(v uint32) RGB {
	return RGB{float64(v >> 16 & 0xff), float64(v >> 8 & 0xff), float64(v & 0xff)}
}

func (c RGB) Lerp(o RGB, t float64) RGB {
	if t <= 0 {
		return c
	}
	if t >= 1 {
		return o
	}
	return RGB{c.R + (o.R-c.R)*t, c.G + (o.G-c.G)*t, c.B + (o.B-c.B)*t}
}

func (c RGB) Mul(k float64) RGB { return RGB{c.R * k, c.G * k, c.B * k}.clamp() }

// Tint multiplies channel-wise by o/255 (a coloured light).
func (c RGB) Tint(o RGB) RGB { return RGB{c.R * o.R / 255, c.G * o.G / 255, c.B * o.B / 255} }

func (c RGB) Add(o RGB, k float64) RGB {
	return RGB{c.R + o.R*k, c.G + o.G*k, c.B + o.B*k}.clamp()
}

func (c RGB) clamp() RGB {
	f := func(v float64) float64 { return math.Max(0, math.Min(255, v)) }
	return RGB{f(c.R), f(c.G), f(c.B)}
}

func (c RGB) Luma() float64 { return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B }

func (c RGB) Gray() RGB { l := c.Luma(); return RGB{l, l, l} }

// Sat scales saturation around the pixel's luma (k<1 desaturates).
func (c RGB) Sat(k float64) RGB { return c.Gray().Lerp(c, k) }

func (c RGB) TC() tcell.Color {
	c = c.clamp()
	return tcell.NewRGBColor(int32(c.R+0.5), int32(c.G+0.5), int32(c.B+0.5))
}

func fromTC(tc tcell.Color) RGB {
	r, g, b := tc.RGB()
	if r < 0 {
		return RGB{}
	}
	return RGB{float64(r), float64(g), float64(b)}
}

func smoothstep(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

// hash is a small deterministic integer hash (splitmix-ish), the source of
// every "random" choice in the scene so a state always draws the same city.
func hash(v ...int) uint32 {
	h := uint64(0x9e3779b97f4a7c15)
	for _, x := range v {
		h ^= uint64(int64(x)) + 0x9e3779b97f4a7c15 + (h << 6) + (h >> 2)
		h *= 0xbf58476d1ce4e5b9
		h ^= h >> 31
	}
	return uint32(h ^ h>>32)
}

func hashf(v ...int) float64 { return float64(hash(v...)%100000) / 100000 }

// noise1 is smooth 1D value noise in [0,1].
func noise1(x float64, seed int) float64 {
	i := int(math.Floor(x))
	f := x - float64(i)
	a, b := hashf(i, seed), hashf(i+1, seed)
	return a + (b-a)*f*f*(3-2*f)
}

func fbm(x float64, seed, oct int) float64 {
	v, amp, tot := 0.0, 1.0, 0.0
	for o := 0; o < oct; o++ {
		v += noise1(x, seed+o*101) * amp
		tot += amp
		x *= 2.03
		amp *= 0.5
	}
	return v / tot
}
