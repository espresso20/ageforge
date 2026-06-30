package citymap

import "math"

// Self-contained value/FBM noise — no external dependency. Salvaged in spirit
// from the retired MapV4's FBM fallback but reimplemented cleanly here so the
// terrain field is deterministic, cheap, and free of any image assets.

// hash2 is a small integer hash producing a well-mixed uint32 from two ints plus
// a seed. Deterministic, so the same (x,y,seed) always yields the same value.
func hash2(x, y, seed uint32) uint32 {
	h := (x+seed)*2654435761 ^ (y+seed)*2246822519 ^ seed*3266489917
	h ^= h >> 15
	h *= 0x2c1b3c6d
	h ^= h >> 12
	h *= 0x297a2d39
	h ^= h >> 15
	return h
}

// hashUnit returns a hash in [0,1).
func hashUnit(x, y, seed uint32) float64 {
	return float64(hash2(x, y, seed)) / float64(^uint32(0))
}

// valueNoise is smooth bilinear-interpolated value noise sampled at (fx,fy).
// Output is in [0,1].
func valueNoise(fx, fy float64, seed uint32) float64 {
	x0 := int32(math.Floor(fx))
	y0 := int32(math.Floor(fy))

	tx := fx - math.Floor(fx)
	ty := fy - math.Floor(fy)
	// Smoothstep the interpolants so the field has no grid-aligned creases.
	tx = tx * tx * (3 - 2*tx)
	ty = ty * ty * (3 - 2*ty)

	v00 := hashUnit(uint32(x0), uint32(y0), seed)
	v10 := hashUnit(uint32(x0+1), uint32(y0), seed)
	v01 := hashUnit(uint32(x0), uint32(y0+1), seed)
	v11 := hashUnit(uint32(x0+1), uint32(y0+1), seed)

	a := v00*(1-tx) + v10*tx
	b := v01*(1-tx) + v11*tx
	return a*(1-ty) + b*ty
}

// fbm sums a few octaves of value noise into a soft elevation field in [0,1].
// Lower frequencies dominate so the terrain reads as broad landmasses with
// gentle detail rather than static. It is the elevation field's entry point and
// fixes the base frequency the terrain has always used.
func fbm(x, y float64, seed uint32) float64 {
	return fbmFreq(x, y, seed, 1.0/38.0)
}

// fbmFreq is fbm with an explicit base frequency, so a second independent field
// (the biome moisture field) can sample the same noise at a different scale — a
// broader, lower frequency reads as large damp/dry regions rather than tracking
// the elevation contours. Same 4-octave shape, same [0,1] output.
func fbmFreq(x, y float64, seed uint32, baseFreq float64) float64 {
	var sum, amp, norm float64
	freq := baseFreq
	amp = 0.5
	for octave := 0; octave < 4; octave++ {
		sum += valueNoise(x*freq, y*freq, seed+uint32(octave)*1013) * amp
		norm += amp
		freq *= 2.05
		amp *= 0.5
	}
	if norm == 0 {
		return 0
	}
	return sum / norm
}

// moistureSeed derives a well-separated, independent seed for the moisture field
// from the elevation seed, so moisture and elevation are uncorrelated (a wet
// lowland and a wet peak are both possible). Deterministic per render.
func moistureSeed(elevSeed uint32) uint32 {
	h := hash2(elevSeed, 0x9e3779b9, 0x517cc1b7)
	return h | 1
}
