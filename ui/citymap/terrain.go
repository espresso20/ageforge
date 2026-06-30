package citymap

import (
	"image"
	"image/color"

	"github.com/espresso20/ageforge/config"
)

// terrain.go turns the FBM elevation field into a theme-tinted band fill on the
// RGBA canvas. Soft and ambient — broad water, lowland, and high ground, with no
// hard edges. The colors come entirely from the active-theme palette (palette.go),
// so the terrain visibly retints when the theme changes.

// Elevation band thresholds on the [0,1] FBM field. Tuned so most of the map is
// walkable land with a little water at the low end and highlights up top.
const (
	bandDeepWater    = 0.30
	bandShallowWater = 0.40
	bandLowland      = 0.55
	bandMidland      = 0.72
	bandHighland     = 0.86
	// >= bandHighland is peak.
)

// ageInfo derives a stable seed and a faint hue shift from the current age key,
// using the canonical age order as an index. Early ages warm, late ages cool —
// a subtle progression in P1; richer per-age styling is P2.
func ageInfo(ageKey string) (seed uint32, hueShift float64) {
	order := config.AgeOrder()
	idx := -1
	for i, k := range order {
		if k == ageKey {
			idx = i
			break
		}
	}
	if idx < 0 {
		// Unknown age (e.g. a test stub): derive a seed from the string so the
		// field is still deterministic, and use a neutral hue.
		var h uint32 = 2166136261
		for i := 0; i < len(ageKey); i++ {
			h ^= uint32(ageKey[i])
			h *= 16777619
		}
		return h | 1, 0
	}
	// Seed: spread ages apart so successive ages get visibly different terrain.
	seed = uint32(idx+1) * 2654435761
	// Hue: walk from +0.5 (warm) at the first age to -0.5 (cool) at the last.
	n := len(order)
	if n > 1 {
		hueShift = 0.5 - float64(idx)/float64(n-1)
	}
	return seed, hueShift
}

// colorForElevation picks the band color for an elevation value.
func colorForElevation(e float64, p terrainPalette) color.RGBA {
	switch {
	case e < bandDeepWater:
		return p.deepWater
	case e < bandShallowWater:
		// Soften the shore: blend shallow→lowland across the band.
		t := (e - bandDeepWater) / (bandShallowWater - bandDeepWater)
		return blend(p.shallowWater, p.lowland, t*0.5)
	case e < bandLowland:
		return p.lowland
	case e < bandMidland:
		return p.midland
	case e < bandHighland:
		return p.highland
	default:
		return p.peak
	}
}

// drawTerrain fills img with the procedural, theme-tinted terrain field.
func drawTerrain(img *image.RGBA, p terrainPalette, seed uint32) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			e := fbm(float64(x), float64(y), seed)
			img.SetRGBA(x, y, colorForElevation(e, p))
		}
	}
}
