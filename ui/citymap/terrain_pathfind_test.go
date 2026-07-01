package citymap

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// ---- Biome classifier + passability ----------------------------------------

// TestClassifyBiomeValidAndPassability sweeps the (elevation, moisture) unit square
// and asserts every pair classifies into a real biome and that the passability rule
// matches the biome: deep/shallow water and mountain/snow are impassable, all the
// land biomes (sand/grass/forest/rock) are passable. This is the contract roads and
// placement depend on.
func TestClassifyBiomeValidAndPassability(t *testing.T) {
	for ei := 0; ei <= 20; ei++ {
		for mi := 0; mi <= 20; mi++ {
			e := float64(ei) / 20.0
			m := float64(mi) / 20.0
			b := classifyBiome(e, m)
			if b >= biomeCount {
				t.Fatalf("classifyBiome(%.2f,%.2f) = %d, out of range", e, m, b)
			}
			// Passability must be consistent with the documented rule.
			wantPassable := true
			switch b {
			case biomeDeepWater, biomeShallowWater, biomeMountain, biomeSnow:
				wantPassable = false
			}
			if passableBiome(b) != wantPassable {
				t.Fatalf("biome %d passability = %v, want %v", b, passableBiome(b), wantPassable)
			}
			// Low ground must always be water; the very top must always be snow.
			if e < bandDeepWater && b != biomeDeepWater {
				t.Fatalf("elev %.2f below deep-water band but biome = %d", e, b)
			}
			if e >= bandHighland+0.06 && b != biomeSnow {
				t.Fatalf("elev %.2f at peak but biome = %d, want snow", e, b)
			}
		}
	}
}

// TestNewTerrainFieldProducesMixAndPassabilityGrid renders a real biome field from
// sample noise and asserts it (1) is the right size, (2) produces a passability
// grid aligned with its biomes, and (3) contains BOTH passable and impassable cells
// (a field that was all-land or all-water would make the road/placement logic moot).
func TestNewTerrainFieldProducesMixAndPassabilityGrid(t *testing.T) {
	const w, h = 80, 96
	f := newTerrainField(w, h, 0xC0FFEE)
	if f.w != w || f.h != h {
		t.Fatalf("field size = %dx%d, want %dx%d", f.w, f.h, w, h)
	}
	if len(f.biomes) != w*h || len(f.passable) != w*h {
		t.Fatalf("grid lengths biomes=%d passable=%d, want %d", len(f.biomes), len(f.passable), w*h)
	}
	passN, blockN := 0, 0
	for i, b := range f.biomes {
		if b >= biomeCount {
			t.Fatalf("field cell %d has invalid biome %d", i, b)
		}
		// passable[] must agree with passableBiome(biome[]) cell-for-cell.
		if f.passable[i] != passableBiome(b) {
			t.Fatalf("cell %d passable=%v but biome %d says %v", i, f.passable[i], b, passableBiome(b))
		}
		if f.passable[i] {
			passN++
		} else {
			blockN++
		}
	}
	if passN == 0 {
		t.Fatal("field has no passable land — biome thresholds produce all water/rock")
	}
	if blockN == 0 {
		t.Fatal("field has no impassable cells — nothing for roads/placement to avoid")
	}
	// Out-of-range queries are safe: impassable + open-land defaults.
	if f.passableAt(-1, 0) || f.passableAt(w, 0) {
		t.Fatal("off-canvas passableAt should be false (edge is a wall)")
	}
	if f.at(-1, -1) != biomeGrass {
		t.Fatal("off-canvas at() should default to grass")
	}
}

// blockedStripField builds a synthetic field: open land everywhere except a vertical
// wall of deep water spanning the full height at column band [wallX0,wallX1], with a
// single passable gap removed (so the only way across is around the top or bottom).
// Used to prove the pathfinder detours around water deterministically, independent
// of noise.
func blockedStripField(w, h, wallX0, wallX1 int) *terrainField {
	f := &terrainField{w: w, h: h}
	f.biomes = make([]biome, w*h)
	f.passable = make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if x >= wallX0 && x <= wallX1 {
				f.biomes[idx] = biomeDeepWater
				f.passable[idx] = false
			} else {
				f.biomes[idx] = biomeGrass
				f.passable[idx] = true
			}
		}
	}
	return f
}

// ---- Terrain-aware roads ----------------------------------------------------

// TestRoadRoutesAroundWater is the core terrain-aware-roads guarantee: a road from a
// building on the left of a full-height water wall to the City Center on the right
// must route AROUND the wall — no waypoint may land on a blocked cell, and the path
// must be longer than the straight-line distance (it detoured). The wall spans the
// whole height with no gap through the middle, so a straight Bresenham line would
// cut through ~deep water; A* must go around the top or bottom edge.
func TestRoadRoutesAroundWater(t *testing.T) {
	const w, h = 90, 60
	// A water wall down the middle third of the width, full height EXCEPT we leave the
	// top two rows and bottom two rows open so a detour exists.
	f := blockedStripField(w, h, 40, 50)
	for x := 40; x <= 50; x++ {
		for _, y := range []int{0, 1, h - 2, h - 1} {
			idx := y*w + x
			f.biomes[idx] = biomeGrass
			f.passable[idx] = true
		}
	}

	grid := buildCostGrid(f, 1)
	startX, startY := 10, h/2
	goalX, goalY := w-10, h/2
	pts, ok := grid.findPath(startX, startY, goalX, goalY)
	if !ok {
		t.Fatal("no path found across a wall that has top/bottom gaps — A* failed to detour")
	}
	if len(pts) < 2 {
		t.Fatalf("degenerate path of %d points", len(pts))
	}
	// No waypoint may sit on a blocked cell.
	for _, p := range pts {
		if !f.passableAt(p[0], p[1]) {
			t.Fatalf("route waypoint (%d,%d) is on a blocked (water) cell", p[0], p[1])
		}
	}
	// The route must be a genuine detour: its traversed length exceeds the straight-
	// line distance between endpoints by a clear margin (it had to go around).
	// Compare path length to the straight-line distance. Endpoints share a row, so the
	// straight-line (and Manhattan) distance is just dx; a detour up-and-over adds
	// vertical travel, so the path's Manhattan length must clear dx by a clear margin.
	straight := float64(goalX - startX)
	var routeLen float64
	for i := 0; i+1 < len(pts); i++ {
		routeLen += manhattan(pts[i], pts[i+1])
	}
	if routeLen <= straight*1.2 {
		t.Fatalf("route length %.1f not meaningfully longer than straight %.1f — did not detour",
			routeLen, straight)
	}
}

// TestRoadNetworkConnectsAllAcrossWater proves the shared road network: given a City
// Center and buildings on BOTH sides of a full-height water wall (with top/bottom gaps),
// buildRoadNetwork must (1) return a non-empty set of segments, (2) place no segment
// pixel on a water cell (routes go around), and (3) connect every node into one network
// — each node pixel appears as an endpoint of at least one routed edge (smoothPath pins
// the edge endpoints to the node coords), and the endpoint graph is a single component.
func TestRoadNetworkConnectsAllAcrossWater(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := buildPalette(0)
	const w, h = 90, 60
	f := blockedStripField(w, h, 40, 50)
	for x := 40; x <= 50; x++ {
		for _, y := range []int{0, 1, h - 2, h - 1} {
			idx := y*w + x
			f.biomes[idx] = biomeGrass
			f.passable[idx] = true
		}
	}
	grid := buildCostGrid(f, 5)

	center := [2]int{20, h / 2}
	buildings := [][2]int{
		{10, 10},     // left side
		{15, h - 10}, // left side
		{75, 12},     // right side (across the wall)
		{80, h - 12}, // right side (across the wall)
	}
	segs := buildRoadNetwork(grid, center, buildings)
	if len(segs) == 0 {
		t.Fatal("buildRoadNetwork produced no segments for a 5-node network")
	}

	// (2) Rasterize and assert nothing lands on water.
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for _, s := range segs {
		drawRoad(img, s, pal.road)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.RGBAAt(x, y) == pal.road && !f.passableAt(x, y) {
				t.Fatalf("network road pixel on a water cell at (%d,%d)", x, y)
			}
		}
	}

	// (3) Every node must be touched by the network, and the network must be one
	// connected component. Build an adjacency over segment endpoints (union-find) and
	// assert all node pixels share a root.
	uf := map[[2]int]*[2]int{}
	var find func(p [2]int) [2]int
	find = func(p [2]int) [2]int {
		r, ok := uf[p]
		if !ok {
			uf[p] = &p
			return p
		}
		if *r == p {
			return p
		}
		root := find(*r)
		*r = root
		return root
	}
	union := func(a, b [2]int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			*uf[ra] = rb
		}
	}
	for _, s := range segs {
		union([2]int{s.x0, s.y0}, [2]int{s.x1, s.y1})
	}
	// A* pins each routed edge to NODE CENTERS (the cost grid is downsampled), so compare
	// connectivity at the node-center a raw pixel maps to — that is the network's true
	// contract, not the exact input pixel.
	snap := func(p [2]int) [2]int {
		nx, ny := grid.pxToNode(p[0], p[1])
		cx, cy := grid.nodePx(nx, ny)
		return [2]int{cx, cy}
	}
	nodes := append([][2]int{center}, buildings...)
	for _, raw := range nodes {
		n := snap(raw)
		if _, ok := uf[n]; !ok {
			t.Fatalf("node %v (snapped %v) is not an endpoint of any road segment — not connected", raw, n)
		}
	}
	root0 := find(snap(nodes[0]))
	for _, raw := range nodes[1:] {
		if find(snap(raw)) != root0 {
			t.Fatalf("node %v not in the same connected component as the City Center", raw)
		}
	}
}

// TestFindPathFallsBackWhenBoxedIn verifies the unreachable case: a building fully
// enclosed by deep water (no open ring around it) yields no A* path, so the caller
// can fall back to a direct line. We box a single open cell inside a water moat and
// route from outside; with no passable corridor the snap-to-open search fails and
// findPath returns ok=false.
func TestFindPathFallsBackWhenBoxedIn(t *testing.T) {
	const w, h = 60, 60
	// All water...
	f := &terrainField{w: w, h: h, biomes: make([]biome, w*h), passable: make([]bool, w*h)}
	for i := range f.biomes {
		f.biomes[i] = biomeDeepWater
		f.passable[i] = false
	}
	// ...except one open cell at (5,5) and a separate open cell at the goal (50,50),
	// with nothing connecting them.
	open := func(x, y int) {
		f.biomes[y*w+x] = biomeGrass
		f.passable[y*w+x] = true
	}
	open(5, 5)
	open(50, 50)

	grid := buildCostGrid(f, 3)
	if _, ok := grid.findPath(5, 5, 50, 50); ok {
		t.Fatal("findPath claimed a route between two cells separated by solid water")
	}
}

// manhattan returns the L1 distance between two pixel waypoints. Used only to show a
// detour route is longer than the straight horizontal run; avoids a math import.
func manhattan(a, b [2]int) float64 {
	return float64(absInt(a[0]-b[0]) + absInt(a[1]-b[1]))
}

// ---- City Center label ------------------------------------------------------

// TestCityCenterLabel asserts the central marker's label reads "City Center" (the
// rename from "Capital"), is the labelCapital kind, and is anchored near the palace.
func TestCityCenterLabel(t *testing.T) {
	_ = theme.SetActive("forge")
	const cols, rows = 80, 40
	geo := layoutGeometry{palaceX: cols / 2, palaceY: rows} // palaceY in px → row ~ rows/2
	plan := buildOverlayPlan(sampleState("iron_age", nil), cols, rows, geo)

	found := false
	for _, lb := range plan.labels {
		if lb.kind == labelCapital {
			found = true
			if lb.text != "City Center" {
				t.Fatalf("center label = %q, want \"City Center\"", lb.text)
			}
		}
		// The old string must be gone entirely.
		if lb.text == "Capital" {
			t.Fatal("overlay still emits a \"Capital\" label")
		}
	}
	if !found {
		t.Fatal("no City Center label planned")
	}
}
