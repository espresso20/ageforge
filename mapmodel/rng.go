package mapmodel

import "math"

// rng.go holds the stateless hashing and the little float maths the model
// needs. The model is part of the determinism contract (it is fingerprinted
// by the smoke harness, see Fingerprint), so it follows the float rules for
// simulation code: every product that feeds a sum is rounded explicitly with
// float64(...), and there are no transcendental calls from package math.
// Sin and Cos below are a plain polynomial for the same reason.

// Hash is a small stateless integer hash, the basis of every "random" choice
// the maps make: the same inputs give the same answer forever, on every
// machine.
func Hash(vals ...int64) uint64 {
	h := uint64(1469598103934665603)
	for _, v := range vals {
		h ^= uint64(v)
		h *= 1099511628211
		h ^= h >> 29
		h *= 0xbf58476d1ce4e5b9
		h ^= h >> 32
	}
	return h
}

// HashF is Hash scaled to [0,1).
func HashF(vals ...int64) float64 { return float64(Hash(vals...)%1_000_000) / 1_000_000 }

// HashStr folds a string into an int64 for Hash.
func HashStr(s string) int64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return int64(h >> 1)
}

// smooth is the smoothstep weight 3t²-2t³.
func smooth(t float64) float64 { return float64(float64(t*t) * (3 - float64(2*t))) }

func lerp(a, b, t float64) float64 { return a + float64((b-a)*t) }

// valueNoise is bilinear value noise on an integer lattice, in [0,1).
func valueNoise(seed int64, x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	sx, sy := smooth(x-x0), smooth(y-y0)
	ix, iy := int64(x0), int64(y0)
	top := lerp(HashF(seed, ix, iy), HashF(seed, ix+1, iy), sx)
	bot := lerp(HashF(seed, ix, iy+1), HashF(seed, ix+1, iy+1), sx)
	return lerp(top, bot, sy)
}

// fbm sums oct octaves of value noise, normalised to [0,1).
func fbm(seed int64, x, y float64, oct int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for i := 0; i < oct; i++ {
		sum += float64(amp * valueNoise(seed+int64(i)*7919, x, y))
		norm += amp
		amp *= 0.5
		x, y = float64(x*2.03), float64(y*2.03)
	}
	return sum / norm
}

// Noise is value noise for the renderers (ridges, clouds, shimmer), with the
// same guarantees as the model's own.
func Noise(seed int64, x, y float64) float64 { return valueNoise(seed, x, y) }

// Sin returns sin(2π·turns). It is a Taylor polynomial on the first quarter
// wave with every product rounded, accurate to about 1e-7, and it gives the
// same bits on every architecture (math.Sin does not promise that).
func Sin(turns float64) float64 {
	t := turns - math.Floor(turns) // [0,1)
	sign := 1.0
	if t >= 0.5 {
		t -= 0.5
		sign = -1
	}
	if t > 0.25 {
		t = 0.5 - t
	}
	x := float64(t * (2 * math.Pi))
	x2 := float64(x * x)
	p := 1.0 / 39916800
	p = float64(p*x2) * -1
	p += 1.0 / 362880
	p = float64(p * x2)
	p -= 1.0 / 5040
	p = float64(p * x2)
	p += 1.0 / 120
	p = float64(p * x2)
	p -= 1.0 / 6
	p = float64(p * x2)
	p += 1
	return sign * float64(x*p)
}

// Cos returns cos(2π·turns).
func Cos(turns float64) float64 { return Sin(turns + 0.25) }

// Log2 returns log₂(x) for x > 0 using an exact integer part and a short
// series for the fraction; it is only used for sub-linear counts, where a
// few ulps do not matter but cross-machine agreement does.
func Log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	frac, exp := math.Frexp(x) // x = frac·2^exp, frac in [0.5,1)
	// log2(frac) via ln((1+s)/(1-s)) = 2(s + s³/3 + s⁵/5 + ...)
	s := (frac - 1) / (frac + 1)
	s2 := float64(s * s)
	sum := 0.0
	term := s
	for k := 1; k < 30; k += 2 {
		sum += term / float64(k)
		term = float64(term * s2)
	}
	return float64(exp) + float64(2*sum)/math.Ln2
}

// rdist is the "round" distance between two tiles: a terminal cell is about
// twice as tall as it is wide, so two columns count as one row.
func rdist(ax, ay, bx, by int) float64 {
	dx := float64(ax-bx) / 2
	dy := float64(ay - by)
	return math.Sqrt(float64(dx*dx) + float64(dy*dy))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
