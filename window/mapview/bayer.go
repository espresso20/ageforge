package mapview

import (
	"image/color"
	"math"

	"github.com/espresso20/ageforge/world"
)

// bayer4 is the 4x4 ordered-dither matrix.
var bayer4 = [4][4]uint8{{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5}}

// Bayer is the dither threshold at pixel x,y, in (0,1).
func Bayer(x, y int) float64 { return (float64(bayer4[y&3][x&3]) + 0.5) / 16 }

// Ramp is n colours from the void (index 0) to the full tone (n-1): the
// limited palette a sprite is shaded from.
func Ramp(p *world.Palette, t world.Tone, n int) []color.RGBA {
	out := make([]color.RGBA, max(n, 2))
	for i := range out {
		out[i] = p.Mix(t, uint8(i*255/(len(out)-1)))
	}
	return out
}

func hash2(x, y int, seed uint32) uint32 {
	h := uint32(x)*0x9E3779B1 ^ uint32(y)*0x85EBCA77 ^ seed*0xC2B2AE3D
	h ^= h >> 15
	h *= 0x2C1B3C6D
	h ^= h >> 12
	h *= 0x297A2D39
	return h ^ h>>15
}

// Noise is smooth value noise in [0,1) at x,y: cheap enough to ask for every
// pixel, which mapgen's fbm is not.
func Noise(x, y float64, seed uint32) float64 {
	ix, iy := math.Floor(x), math.Floor(y)
	fx, fy := x-ix, y-iy
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	at := func(dx, dy int) float64 { return float64(hash2(int(ix)+dx, int(iy)+dy, seed)&0xffff) / 65536 }
	a, b := at(0, 0)+(at(1, 0)-at(0, 0))*fx, at(0, 1)+(at(1, 1)-at(0, 1))*fx
	return a + (b-a)*fy
}
