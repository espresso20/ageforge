package citymap

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/espresso20/ageforge/game"
)

// entities.go places the structures on top of the terrain. P1 keeps this simple
// but visible: a central palace plus one bright marker per built building, laid
// out in a deterministic center-out scatter. Per-age layout strategies, roads,
// and 2.5D building volumes are P2.

// marker is a placed structure: a pixel center, a size in pixels, and whether it
// is the palace (drawn larger, in the accent color).
type marker struct {
	cx, cy int
	size   int
	palace bool
}

// builtBuildingCount counts distinct building types with Count > 0.
func builtBuildingCount(state game.GameState) int {
	n := 0
	for _, bs := range state.Buildings {
		if bs.Count > 0 {
			n++
		}
	}
	return n
}

// builtBuildingKeys returns the sorted keys of buildings with Count > 0, so the
// scatter layout is stable frame-to-frame (map iteration order is otherwise
// random and would make markers jump).
func builtBuildingKeys(state game.GameState) []string {
	keys := make([]string, 0, len(state.Buildings))
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// layoutMarkers produces the palace + a center-out scatter of building markers
// sized to the canvas. The scatter uses a phyllotaxis-style spiral (golden
// angle) so points spread evenly without clumping, scaled to fit the canvas.
func layoutMarkers(w, h int, keys []string) []marker {
	markers := make([]marker, 0, len(keys)+1)
	cx, cy := w/2, h/2

	// Central palace (size is a pixel radius: 2 → a 5×5 block, 1 → 3×3).
	palaceSize := 2
	if w < 40 || h < 30 {
		palaceSize = 1
	}
	markers = append(markers, marker{cx: cx, cy: cy, size: palaceSize, palace: true})

	if len(keys) == 0 {
		return markers
	}

	// Spread radius: keep markers inside a margin so nothing clips the edge.
	maxR := float64(min(w, h))/2 - 3
	if maxR < 4 {
		maxR = 4
	}
	const goldenAngle = 2.399963229728653 // radians (~137.5°)
	n := len(keys)
	for i, k := range keys {
		// Normalised radius grows with sqrt(i) for even areal density; the first
		// ring sits clear of the palace.
		frac := (float64(i) + 1.5) / (float64(n) + 1.5)
		r := maxR * math.Sqrt(frac)
		ang := float64(i) * goldenAngle
		px := cx + int(r*math.Cos(ang))
		py := cy + int(r*math.Sin(ang))
		px = clampInt(px, 1, w-2)
		py = clampInt(py, 1, h-2)
		// A tiny size variation keyed off the building name so the field doesn't
		// look mechanically uniform (0 → single pixel, 1 → 3×3 blob).
		size := 0
		if len(k) > 0 && k[0]%2 == 0 {
			size = 1
		}
		markers = append(markers, marker{cx: px, cy: py, size: size})
	}
	return markers
}

// drawMarkers paints the markers onto img, each with a 1px drop shadow to the
// lower-right so structures lift off the terrain even at half-block resolution.
func drawMarkers(img *image.RGBA, p terrainPalette, markers []marker) {
	for _, m := range markers {
		fill := p.building
		if m.palace {
			fill = p.palace
		}
		// Shadow first (offset down-right), then the body on top.
		drawBlock(img, m.cx+1, m.cy+1, m.size, p.shadow)
		drawBlock(img, m.cx, m.cy, m.size, fill)
	}
}

// drawBlock fills a (size+1)×(size+1) pixel square centered on (cx,cy), clipped
// to the image bounds. size 0 paints a single pixel; size 1 a 3×3 blob, etc.
func drawBlock(img *image.RGBA, cx, cy, size int, c color.RGBA) {
	b := img.Bounds()
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			x, y := cx+dx, cy+dy
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, c)
		}
	}
}

// clampInt clamps v into [lo,hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
