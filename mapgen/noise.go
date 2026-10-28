package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Noise and hashing in whole numbers: nothing here touches a float, so the
// same seed gives the same bits on every machine.

// hash folds integers into one well-mixed number (the model's own hash).
func hash(vals ...int64) uint64 { return mapmodel.Hash(vals...) }

// roll is a stable pick in [0, n) from the hash of vals.
func roll(n int, vals ...int64) int { return int(hash(vals...) % uint64(n)) }

// chance is true about pct times in a hundred, for the same vals.
func chance(pct int, vals ...int64) bool { return roll(100, vals...) < pct }

// noiseMax is one past the largest value the noise returns.
const noiseMax = 1024

// smooth is the smoothstep curve 3t²-2t³ on 0..256.
func smooth(t int) int { return t * t * (768 - 2*t) >> 16 }

func lerp(a, b, t int) int { return a + (b-a)*t>>8 }

func lattice(seed int64, x, y int) int {
	return int(hash(seed, int64(x), int64(y)) % noiseMax)
}

// valueNoise is smooth noise on an integer lattice. x and y are fixed point
// with 8 fraction bits; the result is in [0, noiseMax).
func valueNoise(seed int64, x, y int) int {
	ix, iy := x>>8, y>>8
	fx, fy := smooth(x&255), smooth(y&255)
	top := lerp(lattice(seed, ix, iy), lattice(seed, ix+1, iy), fx)
	bot := lerp(lattice(seed, ix, iy+1), lattice(seed, ix+1, iy+1), fx)
	return lerp(top, bot, fy)
}

// fbm layers octaves of value noise, each at twice the frequency and half the
// weight of the one before.
func fbm(seed int64, x, y, octaves int) int {
	sum, weight, total := 0, 256, 0
	for o := 0; o < octaves; o++ {
		sum += weight * valueNoise(seed+int64(o)*7919, x, y)
		total += weight
		weight >>= 1
		x, y = x*2, y*2
	}
	return sum / total
}

// isqrt is the integer square root.
func isqrt(n int) int {
	if n <= 0 {
		return 0
	}
	r := n
	for r*r > n {
		r = (r + n/r) / 2
	}
	for (r+1)*(r+1) <= n {
		r++
	}
	return r
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func clamp(v, lo, hi int) int { return min(max(v, lo), hi) }
