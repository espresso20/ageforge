package mapmodel

import "math"

// rng.go holds the stateless hashing and the little float maths the model
// needs. The model is part of the determinism contract (see Fingerprint), so
// it follows the float rules for simulation code: every product that feeds a
// sum is rounded explicitly with float64(...), and there are no
// transcendental calls from package math.

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

// b2i is 1 for true and 0 for false, for hashing a flag.
func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
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
