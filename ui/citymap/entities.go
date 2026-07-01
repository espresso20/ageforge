package citymap

import (
	"image"
	"image/color"
	"sort"

	"github.com/espresso20/ageforge/game"
)

// entities.go holds the shared building-set queries and the low-level pixel
// primitives the map render paths draw with. It once owned the isometric 2.5D
// structure layer (per-era placement strategies → drawVolume, terrain-routed roads),
// but that whole path was retired with the citymap v3 top-down rewrite
// (design-and-architecture/city-synthesis.md): the citymap now renders top-down roofs
// via topdown.go and no longer plants isometric volumes on terrain. What remains here
// is the still-shared surface:
//
//   - builtBuildingCount / builtBuildingKeys — the sorted built-set queries both the
//     citymap (topdown gather) and the worldmap (cache key / progress) read;
//   - nearestPassablePx — the worldmap's land-snap for its civ dots;
//   - drawBlock / drawHSpan / drawRoad — the pixel primitives the top-down roof atlas
//     (topdown.go) and the worldmap rasterize with;
//   - absInt / clampInt — the shared int helpers used across citygen, pathfind, trade.

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
// layout is stable frame-to-frame (map iteration order is otherwise random and
// would make districts jump). counts maps each key to its instance count.
func builtBuildingKeys(state game.GameState) (keys []string, counts map[string]int) {
	counts = make(map[string]int, len(state.Buildings))
	keys = make([]string, 0, len(state.Buildings))
	for k, bs := range state.Buildings {
		if bs.Count > 0 {
			keys = append(keys, k)
			counts[k] = bs.Count
		}
	}
	sort.Strings(keys)
	return keys, counts
}

// nearestPassablePx ring-searches outward from (x,y) for the closest passable
// pixel, returning it. Bounded by the canvas span; returns ok=false only if no
// passable pixel exists at all (a fully-flooded canvas, which never happens with
// the real terrain). The search walks ring perimeters so the FIRST hit is the
// nearest, giving the smallest possible nudge. Used by the worldmap to snap its
// civ dots onto land (the citymap top-down path has no terrain to gate against).
func nearestPassablePx(f *terrainField, x, y int) (int, int, bool) {
	if f.passableAt(x, y) {
		return x, y, true
	}
	maxR := f.w
	if f.h > maxR {
		maxR = f.h
	}
	for r := 1; r <= maxR; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if absInt(dx) != r && absInt(dy) != r {
					continue // ring perimeter only — nearest-first
				}
				ax, ay := x+dx, y+dy
				if f.passableAt(ax, ay) {
					return ax, ay, true
				}
			}
		}
	}
	return x, y, false
}

// drawHSpan fills a horizontal run of pixels [x0,x1] at row y with color c,
// clipped to the image. Shared by the top-down roof atlas (topdown.go).
func drawHSpan(img *image.RGBA, x0, x1, y int, c color.RGBA) {
	b := img.Bounds()
	if y < b.Min.Y || y >= b.Max.Y {
		return
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		if x < b.Min.X || x >= b.Max.X {
			continue
		}
		img.SetRGBA(x, y, c)
	}
}

// drawRoad rasterizes a road segment onto img with Bresenham's line algorithm,
// painting each pixel the (theme-derived) color. Clean single-pixel lines. Shared
// by the top-down atlas line helper (topdown.go's drawLineC) and the worldmap.
func drawRoad(img *image.RGBA, s roadSeg, c color.RGBA) {
	b := img.Bounds()
	x0, y0, x1, y1 := s.x0, s.y0, s.x1, s.y1
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if x0 >= b.Min.X && x0 < b.Max.X && y0 >= b.Min.Y && y0 < b.Max.Y {
			img.SetRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// drawBlock fills a (size+1)×(size+1) pixel square centered on (cx,cy), clipped
// to the image bounds. size 0 paints a single pixel; size 1 a 3×3 blob, etc.
// Shared by the top-down atlas (props, walls/gates) and other draw paths.
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

// absInt returns the absolute value of v.
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
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
