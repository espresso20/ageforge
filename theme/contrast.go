package theme

import (
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// This file implements the contrast-safety guard math (theming.md §8). Two
// independent properties:
//
//  1. WCAG luminance contrast (RelativeLuminance / ContrastRatio) — "is the text
//     readable against the background." Used for the AA floors.
//  2. Colorblind distinguishability (Distinguishable) — "do two role colors stay
//     perceptibly different under deutan/protan simulation." Luminance contrast is
//     blind to this; it's what catches an Accent ≈ Positive collision for a
//     colorblind player.
//
// The contrast TEST (contrast_test.go) drives all of this against the shipped
// themes; these are the reusable primitives.

// srgbToLinear inverts the sRGB transfer function for one channel in [0,1].
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// RelativeLuminance returns the WCAG relative luminance of c in [0,1]:
// 0.2126 R + 0.7152 G + 0.0722 B over sRGB-linearized channels (theming.md §8).
func RelativeLuminance(c tcell.Color) float64 {
	r, g, b := rgbUnit(c)
	rl := srgbToLinear(r)
	gl := srgbToLinear(g)
	bl := srgbToLinear(b)
	return 0.2126*rl + 0.7152*gl + 0.0722*bl
}

// ContrastRatio returns the WCAG contrast ratio between a and b, in [1.0, 21.0]:
// (Llighter + 0.05) / (Ldarker + 0.05). Order-independent.
func ContrastRatio(a, b tcell.Color) float64 {
	la := RelativeLuminance(a)
	lb := RelativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// DistinguishableThreshold is the minimum CIEDE2000 distance between two colors,
// after colorblind simulation, for them to count as perceptibly different.
//
// NOTE ON SCALE: go-colorful's DistanceCIEDE2000 operates on its Lab values, which
// are normalized to [0,1]-ish channels rather than the textbook L*∈[0,100] scale.
// Its distances therefore land ~1/100th of conventional ΔE numbers — e.g. a vivid
// red vs green here is ~0.72, not ~72. Do not reach for a "ΔE ~ 1 = JND" intuition;
// these are this library's units. The threshold below is calibrated empirically
// against the shipped palettes (see contrast_test.go):
//
//	accessible blue/orange ± under deut/protan sim: ~0.63–0.78  (must PASS)
//	red/green under deut/protan sim:                ~0.26–0.43  (must FAIL — collapses)
//	red/green with no sim:                          ~0.72       (passes — they differ)
//
// 0.55 cleanly separates the safe pairs (margin above) from a CVD-collapsed
// red/green (well below), while still treating un-simulated red/green as distinct.
const DistinguishableThreshold = 0.55

// Distinguishable reports whether colors a and b remain perceptibly different after
// simulating the given color-vision deficiency. sim is "deuteranopia" or
// "protanopia" (case-insensitive prefix "deut"/"prot"); any other value simulates
// nothing (compares the colors as-is). Distance is CIEDE2000 ΔE in Lab space.
//
// This is the §8 colorblind check: it catches role-color collisions that luminance
// contrast cannot see.
func Distinguishable(a, b tcell.Color, sim string) bool {
	ca := simulate(a, sim)
	cb := simulate(b, sim)
	return ca.DistanceCIEDE2000(cb) >= DistinguishableThreshold
}

// toColorful converts a tcell.Color to a go-colorful Color via its RGB components.
func toColorful(c tcell.Color) colorful.Color {
	r, g, b := rgbUnit(c)
	return colorful.Color{R: r, G: g, B: b}
}

// rgbUnit returns c's RGB as floats in [0,1].
func rgbUnit(c tcell.Color) (r, g, b float64) {
	ri, gi, bi := c.RGB()
	return float64(ri) / 255.0, float64(gi) / 255.0, float64(bi) / 255.0
}

// simulate applies a dichromat color-vision-deficiency transform to c and returns
// the perceived color. Uses the Viénot/Brettel-style approach: linearize sRGB,
// project onto the dichromat plane in LMS via a fixed matrix, then de-linearize.
// "deut*" => deuteranopia, "prot*" => protanopia; anything else is identity.
func simulate(c tcell.Color, sim string) colorful.Color {
	r, g, b := rgbUnit(c)
	// Linearize for the matrix math (the CVD matrices operate on linear light).
	r, g, b = srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)

	var rr, gg, bb float64
	switch normSim(sim) {
	case "deuteranopia":
		// Viénot 1999 deuteranopia matrix (sRGB-linear domain).
		rr = 0.625*r + 0.375*g + 0.0*b
		gg = 0.70*r + 0.30*g + 0.0*b
		bb = 0.0*r + 0.30*g + 0.70*b
	case "protanopia":
		// Viénot 1999 protanopia matrix.
		rr = 0.567*r + 0.433*g + 0.0*b
		gg = 0.558*r + 0.442*g + 0.0*b
		bb = 0.0*r + 0.242*g + 0.758*b
	default:
		rr, gg, bb = r, g, b
	}

	// Clamp then de-linearize back to sRGB for a perceptual (Lab) comparison.
	rr = clamp01(rr)
	gg = clamp01(gg)
	bb = clamp01(bb)
	return colorful.Color{
		R: linearToSRGB(rr),
		G: linearToSRGB(gg),
		B: linearToSRGB(bb),
	}
}

// linearToSRGB applies the forward sRGB transfer function to one linear channel.
func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

// normSim normalizes a CVD name to its canonical form by prefix.
func normSim(sim string) string {
	s := strings.ToLower(strings.TrimSpace(sim))
	switch {
	case strings.HasPrefix(s, "deut"):
		return "deuteranopia"
	case strings.HasPrefix(s, "prot"):
		return "protanopia"
	default:
		return ""
	}
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
