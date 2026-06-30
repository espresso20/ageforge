package citymap

import (
	"image"
	"image/color"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// entities.go places the structures on top of the terrain. P2 upgrades P1's flat
// markers to a real structure layer: every distinct built building type becomes its
// own 2.5D volume, clustered with its lineage-mates (layout.go) and arranged by the
// era's placement strategy, connected by roads. Volumes are drawn as a lit roof +
// shaded wall + drop shadow colored per lineage (palette.go) so a lineage reads as a
// same-colored neighborhood. Roads draw under the buildings and over the terrain.
// The overlay then names each volume with the building's own config Name.

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

// drawStructures is the P2/P3 structure entry point: it groups the built buildings
// into lineage districts (one buildingItem per distinct built type), runs the era
// strategy, draws roads under the buildings, then draws each 2.5D volume. byKey is
// the config lineage/category/name table (passed in so the render path builds it
// once). It returns the layoutGeometry — the palace pixel plus one label anchor per
// building — so the P3 overlay pass can stamp each building's NAME exactly where its
// volume landed. (Trade-lane endpoints are filled in later by the trade pass.)
func drawStructures(img *image.RGBA, pal terrainPalette, state game.GameState, byKey map[string]config.BuildingDef, field *terrainField, terrainSeed uint32) layoutGeometry {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	keys, counts := builtBuildingKeys(state)
	districts := builtDistricts(byKey, keys, counts)

	e := eraForAge(state.Age)
	seed := layoutSeed(state.Age, keys)
	res := buildLayout(e, w, h, districts, pal, seed)

	// Terrain-aware pass: the biome field (built once in renderImage from the terrain
	// seed, so it matches the painted terrain exactly) drives both the placement nudge
	// — buildings + City Center moved off water/mountain — and the road cost grid, so
	// roads route around obstacles. The cost-grid jitter is keyed off the terrain seed
	// so a given city always routes identically.
	res.placements = nudgePlacements(res.placements, field)
	grid := buildCostGrid(field, terrainSeed)

	// Roads first (under the buildings, over the terrain). Each layout road segment
	// is re-routed by A* on the biome cost grid so it bends AROUND lakes and peaks
	// and gently meanders; if a route is impossible (endpoint walled off by water)
	// we fall back to the straight Bresenham segment so the road layer is never blank.
	for _, rseg := range res.roads {
		drawTerrainRoad(img, grid, rseg, pal.road)
	}

	// Buildings: draw normal tier first, then wonders, then the palace, so the
	// important volumes sit on top where placements overlap.
	drawPlacementsByTier(img, pal, res.placements, impNormal)
	drawPlacementsByTier(img, pal, res.placements, impWonder)
	drawPlacementsByTier(img, pal, res.placements, impPalace)

	return geometryFor(w, h, districts, res.placements)
}

// nudgePlacements moves any placement whose center lands on an impassable cell
// (water/mountain) to the nearest passable cell via a spiral/ring search, so every
// building volume — and the City Center — sits on land. The per-era strategy and
// lineage clustering are otherwise untouched: only drowned/cliffed slots move, and
// only as far as the nearest open ground, so a district stays where the strategy
// put it. Returns the adjusted slice (placements are value types).
func nudgePlacements(ps []placement, f *terrainField) []placement {
	if f == nil || len(f.passable) == 0 {
		return ps
	}
	for i := range ps {
		if f.passableAt(ps[i].cx, ps[i].cy) {
			continue
		}
		if nx, ny, ok := nearestPassablePx(f, ps[i].cx, ps[i].cy); ok {
			ps[i].cx, ps[i].cy = nx, ny
		}
	}
	return ps
}

// nearestPassablePx ring-searches outward from (x,y) for the closest passable
// pixel, returning it. Bounded by the canvas span; returns ok=false only if no
// passable pixel exists at all (a fully-flooded canvas, which never happens with
// the real terrain). The search walks ring perimeters so the FIRST hit is the
// nearest, giving the smallest possible nudge.
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

// geometryFor extracts the overlay anchors from a completed layout: the palace
// center plus one label anchor per BUILDING (every non-palace placement). Each
// placement already carries its building identity (name + lineage/category + tier +
// size), so the anchor is just the marker pixel plus that identity — the overlay
// names every marker with its own building Name and colors the label by lineage.
// No filtering happens here: all markers get an anchor; the overlay is what
// collision-limits the labels (and still guarantees the prominent building per
// cluster). The districts arg is unused now that identity rides on the placement,
// but kept so the structure entry point's signature is stable for tests.
func geometryFor(w, h int, _ []district, placements []placement) layoutGeometry {
	geo := layoutGeometry{palaceX: w / 2, palaceY: h / 2}

	for _, p := range placements {
		if p.tier == impPalace {
			// Recover the palace center from its placement (every strategy centers it,
			// but read it back rather than assume, in case a strategy offsets it).
			geo.palaceX, geo.palaceY = p.cx, p.cy
			continue
		}
		if p.name == "" {
			continue // a non-building volume (none today) — nothing to label
		}
		geo.buildings = append(geo.buildings, buildingLabel{
			px:         p.cx,
			py:         p.cy,
			name:       p.name,
			lineageKey: p.lineageKey,
			category:   p.category,
			tier:       p.ltier,
			size:       p.size,
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

// drawTerrainRoad routes one layout road segment around the terrain and draws it.
// It asks the cost grid for an A* route between the segment endpoints; on success it
// smooths the polyline (so the coarse-grid staircase reads as a meandering curve)
// and rasterizes it as a connected road. If A* finds no route (an endpoint walled
// off by water/mountain), it falls back to the straight Bresenham segment so the
// road layer is never blank — the road still tries to exist even when boxed in.
func drawTerrainRoad(img *image.RGBA, grid *pfCostGrid, s roadSeg, c color.RGBA) {
	if grid != nil {
		if pts, ok := grid.findPath(s.x0, s.y0, s.x1, s.y1); ok && len(pts) >= 2 {
			drawPolyline(img, smoothPath(pts), c)
			return
		}
	}
	// Fallback: straight line (still better than nothing; A* only fails when a route
	// genuinely can't reach the goal across open water).
	drawRoad(img, s, c)
}

// drawPolyline rasterizes a road through a sequence of pixel waypoints, drawing a
// single-pixel Bresenham segment between each consecutive pair so the path is one
// continuous line. Single-pixel matches the weight of the old straight roads (the
// look is unchanged — only the route now bends around terrain); the smoothing pass
// upstream is what turns the coarse A* node steps into a meander.
func drawPolyline(img *image.RGBA, pts [][2]int, c color.RGBA) {
	for i := 0; i+1 < len(pts); i++ {
		drawRoad(img, roadSeg{pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1]}, c)
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
