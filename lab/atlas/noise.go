package main

import "math"

// Seeded, allocation-free value noise. The atlas's world is a set of
// continuous functions of (x, y) in world units, so it can be sampled at any
// zoom without a pixel grid getting in the way.

func hash2(seed int64, x, y int) uint64 {
	h := uint64(seed)*0x9E3779B97F4A7C15 ^ uint64(int64(x))*0xBF58476D1CE4E5B9 ^ uint64(int64(y))*0x94D049BB133111EB
	h ^= h >> 31
	h *= 0xD6E8FEB86659FD93
	h ^= h >> 29
	h *= 0x9E3779B97F4A7C15
	h ^= h >> 32
	return h
}

// rnd01 is a hash-derived uniform in [0,1).
func rnd01(seed int64, x, y int) float64 {
	return float64(hash2(seed, x, y)>>11) / float64(1<<53)
}

func smooth(t float64) float64 { return t * t * (3 - 2*t) }

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// vnoise is value noise in [0,1).
func vnoise(seed int64, x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	ix, iy := int(x0), int(y0)
	fx, fy := smooth(x-x0), smooth(y-y0)
	a := rnd01(seed, ix, iy)
	b := rnd01(seed, ix+1, iy)
	c := rnd01(seed, ix, iy+1)
	d := rnd01(seed, ix+1, iy+1)
	return lerp(lerp(a, b, fx), lerp(c, d, fx), fy)
}

// fbm is fractal value noise in roughly [0,1).
func fbm(seed int64, x, y float64, oct int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for i := 0; i < oct; i++ {
		sum += amp * vnoise(seed+int64(i)*131, x, y)
		norm += amp
		amp *= 0.5
		x, y = x*2.03+17.1, y*2.03-9.7
	}
	return sum / norm
}

// ridged is ridged fbm in [0,1): sharp crests for mountain ranges.
func ridged(seed int64, x, y float64, oct int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for i := 0; i < oct; i++ {
		n := 1 - math.Abs(vnoise(seed+int64(i)*977, x, y)*2-1)
		sum += amp * n * n
		norm += amp
		amp *= 0.5
		x, y = x*2.1+3.3, y*2.1+7.7
	}
	return sum / norm
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func hashStr(s string) int64 {
	var h uint64 = 1469598103934665603
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return int64(h >> 1)
}
