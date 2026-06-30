package citymap

import (
	"image"
	"image/color"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// entities.go places the structures on top of the terrain. P2 upgrades P1's flat
// markers to a real structure layer: buildings are grouped into lineage districts
// (layout.go), arranged by the era's placement strategy, connected by roads, and
// drawn as small 2.5D volumes (lit roof + shaded wall + drop shadow) colored per
// lineage (palette.go). Roads draw under the buildings and over the terrain.

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

// drawStructures is the P2/P3 structure entry point: it resolves districts, runs
// the era strategy, draws roads under the buildings, then draws each 2.5D volume.
// byKey is the config lineage/category table (passed in so the render path builds it
// once). It returns the layoutGeometry — the palace pixel plus per-district
// centroids — so the P3 overlay pass can stamp district labels exactly where the
// volumes landed. (Trade-lane endpoints are filled in later by the trade pass.)
func drawStructures(img *image.RGBA, pal terrainPalette, state game.GameState, byKey map[string]config.BuildingDef) layoutGeometry {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	keys, counts := builtBuildingKeys(state)
	districts := builtDistricts(byKey, keys, counts)

	e := eraForAge(state.Age)
	seed := layoutSeed(state.Age, keys)
	res := buildLayout(e, w, h, districts, pal, seed)

	// Roads first (under the buildings, over the terrain).
	for _, rseg := range res.roads {
		drawRoad(img, rseg, pal.road)
	}

	// Buildings: draw normal tier first, then wonders, then the palace, so the
	// important volumes sit on top where placements overlap.
	drawPlacementsByTier(img, pal, res.placements, impNormal)
	drawPlacementsByTier(img, pal, res.placements, impWonder)
	drawPlacementsByTier(img, pal, res.placements, impPalace)

	return geometryFor(w, h, districts, res.placements)
}

// geometryFor extracts the overlay anchors from a completed layout: the palace
// center and one centroid per district worth labeling. A district's centroid is the
// mean of its representative volumes; districts are matched to their placements by
// color (lineageColor gives each district a distinct color, which scatterDistrict
// copies into every placement), excluding the palace tier. Only districts with
// enough buildings to warrant a banner (reps >= 2, or a wonder/monument) are kept,
// so a single hut doesn't get a label.
func geometryFor(w, h int, districts []district, placements []placement) layoutGeometry {
	geo := layoutGeometry{palaceX: w / 2, palaceY: h / 2}

	// Recover the palace center from its placement (every strategy centers it, but
	// read it back rather than assume, in case a strategy ever offsets it).
	for _, p := range placements {
		if p.tier == impPalace {
			geo.palaceX, geo.palaceY = p.cx, p.cy
			break
		}
	}

	for _, d := range districts {
		// Worth a banner? Skip thin districts to keep the map from over-labeling.
		worth := d.reps >= 2 || d.category == "wonder" || d.category == "monument" || d.category == "diplomacy"
		if !worth {
			continue
		}
		var sx, sy, n int
		for _, p := range placements {
			if p.tier == impPalace {
				continue
			}
			if p.col == d.col {
				sx += p.cx
				sy += p.cy
				n++
			}
		}
		if n == 0 {
			continue
		}
		geo.districts = append(geo.districts, districtCentroid{
			px:         sx / n,
			py:         sy / n,
			lineageKey: d.lineageKey,
			category:   d.category,
			count:      d.count,
		})
	}
	return geo
}

// drawPlacementsByTier draws only the placements at the given tier.
func drawPlacementsByTier(img *image.RGBA, pal terrainPalette, ps []placement, tier importance) {
	for _, p := range ps {
		if p.tier == tier {
			drawVolume(img, pal, p)
		}
	}
}

// drawVolume renders one building as a 2.5D volume:
//
//	shadow — RoleDim-derived, offset down-right, drawn first so it sits beneath.
//	wall   — the lineage color darkened, one row below the roof (the shaded side).
//	roof   — the lineage color brightened (the lit top), drawn last on top.
//
// The roof/wall/shadow stack lifts the structure off the terrain and reads as
// dimensional even at half-block resolution, without a literal iso projection.
// Size scales the footprint; palace/wonder tiers are nudged up for prominence.
func drawVolume(img *image.RGBA, pal terrainPalette, p placement) {
	// Size by importance, uniformly across every strategy: the palace dominates,
	// wonders/monuments read a touch bigger than ordinary buildings. Done here
	// (not at placement time) so all strategies get consistent showpiece sizing.
	size := p.size
	switch p.tier {
	case impPalace:
		size += 1
	case impWonder:
		size += 1
	}
	if size < 0 {
		size = 0
	}

	roof := brighten(p.col, 0.28)
	wall := darken(p.col, 0.35)
	shadow := pal.shadow

	// Shadow: offset down-right by 1px, same footprint, drawn underneath.
	drawBlock(img, p.cx+1, p.cy+size+2, size, shadow)

	// Wall: the shaded side, one-to-two rows beneath the roof. For size 0 this is
	// a single pixel directly below; for larger volumes it's a band.
	wallRows := 1
	if size >= 2 {
		wallRows = 2
	}
	for dy := 1; dy <= wallRows; dy++ {
		drawHSpan(img, p.cx-size, p.cx+size, p.cy+size+dy, wall)
	}

	// Roof: the lit top, drawn last so it crowns the volume.
	drawBlock(img, p.cx, p.cy, size, roof)
}

// drawHSpan fills a horizontal run of pixels [x0,x1] at row y with color c,
// clipped to the image.
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
// painting each pixel the (theme-derived) road color. Clean single-pixel lines so
// the spokes / grid / rings stay legible against the terrain.
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
