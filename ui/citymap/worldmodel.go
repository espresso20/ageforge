package citymap

import (
	"math"
	"sort"
)

// worldmodel.go builds the world map's GEOGRAPHY: a seeded continent — a single
// central landmass surrounded by ocean, with real coastlines (bays + peninsulas),
// downhill rivers that reach the sea, and relief anchors for mountain/hill/forest
// symbols. It replaces the old flat FBM wash (a canvas of edge-to-edge noise) with a
// heightmap-driven model, but keeps the SAME biome vocabulary (classifyBiome /
// passableBiome / biomeColor) and produces a *terrainField so every existing
// downstream consumer — the terrain paint, the settlement gate, the civ-dot land snap
// — works unchanged. It just now sees a real continent instead of a noise field.
//
// The model is a pure, deterministic function of (w, h, seed): the same inputs always
// yield the same continent. It is panic-safe on tiny/zero canvases (an empty model
// whose derived field guards every query). Age-independence is preserved by the caller
// seeding from the display name (worldTerrainSeed), NOT the age — aging up must never
// rearrange the land.
//
// Resolution strategy: elevation + moisture live on a LOW-RES grid (~1 cell per
// worldGridPx pixels) so flow accumulation for rivers is cheap and the coastline reads
// as broad geography, not per-pixel fuzz. The grid is upsampled bilinearly to the
// full pixel canvas when classifying biomes / painting, so the coast is smooth.

// worldGridPx is the pixel span of one low-res model cell. The elevation/moisture
// fields and the river flow graph live at this coarser resolution (a continent's shape
// is large-scale; per-pixel elevation would just be noise) and are upsampled to the
// pixel canvas at paint time. ~4px keeps rivers cheap while staying fine enough that
// the coastline reads smooth after bilinear upsample.
const worldGridPx = 4

// worldSeaLevel is the elevation cutoff (on the [0,1] combined field) below which a
// cell is ocean. The island mask forces the field to ~0 at the frame edge, so this
// threshold carves the coastline somewhere inside the frame — a continent ringed by
// sea. Tuned with worldIslandGain + islandMask's core so the continent fills roughly
// half the frame (a real landmass, not a speck) while the border stays open water.
const worldSeaLevel = 0.36

// worldIslandGain scales the radial island mask's falloff. The mask is 1 at the center
// and drops toward 0 at the frame edge; a higher gain makes the land pull in tighter
// (more surrounding ocean), a lower one lets the continent spread toward the edges.
// Kept just under 1 so the continent reaches most of the way to the frame before the
// shoulder drops it into sea — a broad landmass with a coastal margin, not a small isle.
const worldIslandGain = 0.92

// worldRiverFlowFrac is the fraction of the maximum accumulated flow a cell must carry
// to be drawn as a river. Lower → more, thinner tributaries; higher → only the major
// trunks. Kept moderate so a few readable rivers reach the sea without webbing the
// whole continent.
const worldRiverFlowFrac = 0.10

// river is a single watercourse: a downhill polyline of pixel points from an inland
// source to the sea, plus the accumulated flow at each point (drives the widening
// toward the mouth). Points are in PIXEL space so the render can stroke them directly.
type river struct {
	pts  []worldPt // ordered source → mouth, pixel coords
	flow []float64 // accumulated flow at each point (monotonic-ish, peaks at the mouth)
}

// worldPt is a pixel-space point.
type worldPt struct{ x, y int }

// reliefAnchor is a single relief symbol location: a pixel-space point plus which kind
// of dab to draw (mountain / hill / forest). Sampled sparsely from the biome field so
// the map has scattered relief without stamping every pixel.
type reliefAnchor struct {
	x, y int
	kind reliefKind
}

// reliefKind selects the relief dab drawn at an anchor.
type reliefKind uint8

const (
	reliefMountain reliefKind = iota
	reliefHill
	reliefForest
)

// worldModel is the full seeded world geography. It owns the low-res elevation +
// moisture grids (for reference / tests), the per-pixel derived terrainField (biome +
// passability — the SAME structure the old flat field produced, so downstream is
// unchanged), the rivers, and the relief anchors. seaLevel is carried so the render can
// depth-band the ocean from the same cutoff the land/water split used.
type worldModel struct {
	w, h     int // pixel dimensions (the canvas)
	gw, gh   int // low-res grid dimensions
	seaLevel float64

	elev  []float64 // low-res elevation grid [gy*gw+gx], in [0,1]
	moist []float64 // low-res moisture grid  [gy*gw+gx], in [0,1]

	field   *terrainField  // per-pixel biome + passability (upsampled), drives all consumers
	rivers  []river        // watercourses (pixel polylines), source → sea
	reliefs []reliefAnchor // scattered mountain/hill/forest symbol anchors
}

// buildWorldModel builds the seeded continent for a w×h pixel canvas. Pure +
// deterministic from (w, h, seed); panic-safe on tiny/zero sizes (returns a model whose
// field guards every query, so callers that iterate rivers/relief simply find none).
//
// Pipeline:
//  1. LOW-RES ELEVATION — FBM octaves at a continent scale, MULTIPLIED by a radial
//     island mask (1 at center → 0 at the edge) plus a couple of large seeded low-freq
//     "lobes" that pull the coastline in/out so it grows bays and peninsulas instead of
//     a clean circle. Normalized to [0,1].
//  2. SEA / LAND — cells with elevation < seaLevel are ocean.
//  3. MOISTURE — FBM (distinct salt) blended with normalized distance-from-ocean (a BFS
//     over the sea) so coasts read wetter than the deep interior.
//  4. PER-PIXEL BIOME — upsample elevation + moisture bilinearly and run the SHARED
//     classifyBiome, filling a terrainField (biome + passability) exactly like the old
//     path, so every downstream consumer is unchanged.
//  5. RIVERS — flow accumulation on the low-res grid: every land cell drains to its
//     lowest neighbour; flow accumulates downstream; sources are high, wet cells; each
//     kept course walks downhill to the sea, widening with accumulated flow.
//  6. RELIEF — sample mountain/hill/forest anchors from the biome field on a sparse
//     stride so relief symbols scatter without covering every cell.
func buildWorldModel(w, h int, seed uint32) *worldModel {
	m := &worldModel{w: w, h: h, seaLevel: worldSeaLevel, field: &terrainField{w: w, h: h}}
	if w <= 0 || h <= 0 {
		return m
	}

	// Low-res grid: at least 1 cell, round up so the whole canvas is covered.
	gw := (w + worldGridPx - 1) / worldGridPx
	gh := (h + worldGridPx - 1) / worldGridPx
	if gw < 1 {
		gw = 1
	}
	if gh < 1 {
		gh = 1
	}
	m.gw, m.gh = gw, gh

	m.buildElevation(seed)
	m.buildMoisture(seed)
	m.buildField(seed)
	m.buildRivers(seed)
	m.buildRelief(seed)
	return m
}

// buildElevation fills the low-res elevation grid. The continent's SHAPE is deliberately
// NOT a clean disc: a soft island envelope only guarantees open sea at the frame border,
// while the actual coastline is CARVED by low-frequency FBM under a heavy DOMAIN WARP.
// The envelope is off-center, elliptical, and rotated (seeded), and the sample position
// that feeds both the envelope and the shape-noise is warped by two more noise fields —
// so bays, capes, peninsulas, and asymmetry emerge instead of a rimmed circle. Seeded
// lobes add a few extra large capes/bays for punch. Normalized to [0,1]; ~0 at the edge.
func (m *worldModel) buildElevation(seed uint32) {
	gw, gh := m.gw, m.gh
	m.elev = make([]float64, gw*gh)

	// Continent-scale noise frequencies (PIXEL space so feature size is resolution-stable).
	const shapeFreq = 1.0 / 78.0  // the noise that CARVES the coastline (dominant)
	const detailFreq = 1.0 / 34.0 // finer interior relief for mountains/valleys
	const warpFreq = 1.0 / 96.0   // very low freq → large, smooth domain-warp swirls

	// Off-center continent: seed a landmass center offset so the continent isn't pinned to
	// the frame middle. Kept modest so it still sits mostly on-canvas.
	ox := (hashUnit(0, 0xCE47, seed) - 0.5) * 0.55
	oy := (hashUnit(1, 0xCE47, seed) - 0.5) * 0.55
	// Elliptical + rotated envelope: per-seed axis stretch + rotation so the base shape is
	// an oval at a random angle, not a circle.
	rot := hashUnit(2, 0xE11B, seed) * math.Pi
	sinR, cosR := math.Sin(rot), math.Cos(rot)
	ax := 0.85 + 0.45*hashUnit(3, 0xE11B, seed) // x semi-axis scale
	ay := 0.85 + 0.45*hashUnit(4, 0xE11B, seed) // y semi-axis scale
	// Domain-warp strength (in normalized units): how far the sample position is displaced
	// by the warp fields. Large → very organic, ragged coast; too large → land fragments.
	const warpAmp = 0.55

	// A few large seeded lobes for extra capes (+) and bays (-) on top of the warp.
	type lobe struct {
		cx, cy, rad, amp float64
	}
	nLobes := 5
	lobes := make([]lobe, nLobes)
	for i := 0; i < nLobes; i++ {
		a := hashUnit(uint32(i), 0x10BE, seed) * 2 * math.Pi
		dist := 0.35 + 0.55*hashUnit(uint32(i), 0x515E, seed)
		lobes[i] = lobe{
			cx:  ox + math.Cos(a)*dist,
			cy:  oy + math.Sin(a)*dist,
			rad: 0.24 + 0.30*hashUnit(uint32(i), 0x4AD1, seed),
			amp: (0.30 + 0.35*hashUnit(uint32(i), 0xB09E, seed)) *
				signFromHash(hashUnit(uint32(i), 0x51D3, seed)),
		}
	}

	wSeedA := seed ^ 0x7ED55D16
	wSeedB := seed ^ 0x2AB0E2C1

	minE, maxE := math.MaxFloat64, -math.MaxFloat64
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			px := float64(gx)*worldGridPx + worldGridPx/2
			py := float64(gy)*worldGridPx + worldGridPx/2

			// Normalized position ([-1,1] on the shorter axis so the envelope isn't
			// aspect-squashed), recentered on the seeded continent offset.
			nx, ny := m.normPos(px, py)
			nx -= ox
			ny -= oy

			// DOMAIN WARP: displace the sample position by two low-frequency noise fields
			// (recentered to [-1,1]) so straight radial contours become wandering, organic
			// ones. This is what breaks the circle: the envelope and the shape-noise are both
			// read at the warped position.
			wx := (fbmFreq(px, py, wSeedA, warpFreq) - 0.5) * 2 * warpAmp
			wy := (fbmFreq(px, py, wSeedB, warpFreq) - 0.5) * 2 * warpAmp
			sx := nx + wx
			sy := ny + wy

			// Elliptical, rotated radius of the WARPED position → the base envelope.
			rx := (sx*cosR + sy*sinR) / ax
			ry := (-sx*sinR + sy*cosR) / ay
			r := math.Hypot(rx, ry)

			// Seeded lobes deform the effective radius (capes pull it in-bounds, bays push out).
			var lobeSum float64
			for _, lb := range lobes {
				d := math.Hypot(sx-lb.cx, sy-lb.cy)
				if d < lb.rad {
					t := 1 - d/lb.rad
					lobeSum += lb.amp * t * t
				}
			}
			effR := (r - lobeSum) * worldIslandGain

			// SIGNED envelope in [-1,1]: +1 at the continent center, 0 in the transition
			// band, and NEGATIVE toward the frame border so the sea reliably reclaims the
			// edges. It is a broad gradient, not the coastline — the coast is carved by the
			// warped shape-noise inside the transition band, where env is near 0 and the
			// noise decides land vs sea.
			envS := islandMask(effR)*2 - 1

			// Shape noise carves the actual coast: low-freq FBM at the WARPED position, so
			// the land/sea boundary follows wandering noise contours, not a circle. Detail
			// noise adds interior relief. Both recentered to [-0.5,0.5].
			shape := fbmFreq(px+wx*40, py+wy*40, seed, shapeFreq) - 0.5
			detail := fbmFreq(px, py, seed^0x1234, detailFreq) - 0.5

			// Combine: the signed envelope guarantees a central continent ringed by sea
			// (weighted strongly enough that the border stays water), while the heavily
			// weighted warped shape-noise makes the coast within the transition band ragged
			// and asymmetric — bays, capes, peninsulas — never a disc.
			e := 0.98*envS + 0.58*shape + 0.20*detail
			m.elev[gy*gw+gx] = e
			if e < minE {
				minE = e
			}
			if e > maxE {
				maxE = e
			}
		}
	}

	// Normalize to [0,1] so seaLevel is a stable cutoff independent of the noise range.
	span := maxE - minE
	if span < 1e-9 {
		span = 1
	}
	for i := range m.elev {
		m.elev[i] = (m.elev[i] - minE) / span
	}

	// Fill small interior LAKES: the shape-noise dipping below sea level mid-continent
	// leaves little water holes that read as a moth-eaten coast rather than intentional
	// inland seas. Flood the ocean in from the frame border; any below-sea pocket NOT
	// connected to that border ocean is an interior lake — lift the SMALL ones just above
	// sea level so they become land, keeping only genuinely large inland seas for character.
	m.fillSmallLakes()
}

// fillSmallLakes lifts small landlocked below-sea pockets to just above sea level, so the
// continent reads as a solid landmass with (at most) a few large inland seas rather than a
// spray of tiny holes. Ocean connected to the frame border is left untouched; only interior
// water components below a size threshold are filled. Operates on the normalized elevation
// grid; deterministic (pure function of the grid).
func (m *worldModel) fillSmallLakes() {
	gw, gh := m.gw, m.gh
	n := gw * gh
	if n == 0 {
		return
	}
	// Flood-fill "open ocean" from every border water cell.
	openSea := make([]bool, n)
	queue := make([]int, 0, n)
	pushIfSea := func(x, y int) {
		if x < 0 || y < 0 || x >= gw || y >= gh {
			return
		}
		i := y*gw + x
		if openSea[i] || m.elev[i] >= m.seaLevel {
			return
		}
		openSea[i] = true
		queue = append(queue, i)
	}
	for x := 0; x < gw; x++ {
		pushIfSea(x, 0)
		pushIfSea(x, gh-1)
	}
	for y := 0; y < gh; y++ {
		pushIfSea(0, y)
		pushIfSea(gw-1, y)
	}
	for head := 0; head < len(queue); head++ {
		i := queue[head]
		x, y := i%gw, i/gw
		pushIfSea(x+1, y)
		pushIfSea(x-1, y)
		pushIfSea(x, y+1)
		pushIfSea(x, y-1)
	}

	// Interior lakes = below-sea cells not reached from the border. Group into components;
	// fill (lift above sea) any component smaller than the threshold. Keep large ones.
	lakeMax := (gw * gh) / 40 // components smaller than ~2.5% of the grid are filled
	if lakeMax < 3 {
		lakeMax = 3
	}
	visited := make([]bool, n)
	for start := 0; start < n; start++ {
		if visited[start] || openSea[start] || m.elev[start] >= m.seaLevel {
			continue
		}
		// BFS this interior lake component.
		comp := []int{start}
		visited[start] = true
		for head := 0; head < len(comp); head++ {
			i := comp[head]
			x, y := i%gw, i/gw
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+nb[0], y+nb[1]
				if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
					continue
				}
				ni := ny*gw + nx
				if visited[ni] || openSea[ni] || m.elev[ni] >= m.seaLevel {
					continue
				}
				visited[ni] = true
				comp = append(comp, ni)
			}
		}
		if len(comp) <= lakeMax {
			for _, i := range comp {
				m.elev[i] = m.seaLevel + 0.02 // just above the waterline → land
			}
		}
	}
}

// buildMoisture fills the low-res moisture grid: FBM (distinct salt) blended with a
// normalized distance-from-ocean so coasts read damp and the deep interior drier. The
// blend keeps large damp/dry regions (from the FBM) while still guaranteeing a wet
// coastal fringe (from the distance field), which reads naturally on the biome map.
func (m *worldModel) buildMoisture(seed uint32) {
	gw, gh := m.gw, m.gh
	m.moist = make([]float64, gw*gh)
	mSeed := moistureSeed(seed)

	// Distance-from-ocean via a multi-source BFS over the sea cells (Manhattan hops on
	// the low-res grid). Coast = 0 hops, deep interior = many hops; normalized to [0,1]
	// then inverted so coast = wet (1), interior = dry (0) before blending.
	dist := m.oceanDistance()
	maxD := 0
	for _, d := range dist {
		if d > maxD {
			maxD = d
		}
	}
	if maxD == 0 {
		maxD = 1
	}

	const moistFreq = 1.0 / 120.0 // broad damp/dry regions, broader than elevation
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			px := float64(gx)*worldGridPx + worldGridPx/2
			py := float64(gy)*worldGridPx + worldGridPx/2
			n := fbmFreq(px, py, mSeed, moistFreq)

			d := dist[gy*gw+gx]
			// Coastal wetness: near the sea (small d) → high; interior (large d) → low.
			coastal := 1.0 - float64(d)/float64(maxD)

			// Blend: mostly the FBM regions, lifted toward wet near the coast. This keeps
			// the biome classifier's forest/grass split organic while ensuring shorelines
			// tend damp.
			m.moist[gy*gw+gx] = clamp01(0.60*n + 0.40*coastal)
		}
	}
}

// oceanDistance returns, per low-res cell, the Manhattan hop count to the nearest ocean
// cell (elevation < seaLevel), via a multi-source BFS seeded from every sea cell. Land
// cells with no sea reachable (a fully landlocked field — never happens with the island
// mask) get maxD+1 so they still read as "far from water." Ocean cells are distance 0.
func (m *worldModel) oceanDistance() []int {
	gw, gh := m.gw, m.gh
	dist := make([]int, gw*gh)
	for i := range dist {
		dist[i] = -1
	}
	queue := make([]int, 0, gw*gh)
	for i, e := range m.elev {
		if e < m.seaLevel {
			dist[i] = 0
			queue = append(queue, i)
		}
	}
	// BFS. If there's no ocean at all (degenerate), everything stays -1 → treat as 0
	// below so moisture doesn't NaN.
	for head := 0; head < len(queue); head++ {
		idx := queue[head]
		gx, gy := idx%gw, idx/gw
		d := dist[idx]
		for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := gx+nb[0], gy+nb[1]
			if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
				continue
			}
			ni := ny*gw + nx
			if dist[ni] != -1 {
				continue
			}
			dist[ni] = d + 1
			queue = append(queue, ni)
		}
	}
	for i := range dist {
		if dist[i] < 0 {
			dist[i] = 0
		}
	}
	return dist
}

// buildField fills the per-pixel terrainField (biome + passability) by upsampling the
// low-res elevation + moisture grids bilinearly and running the SHARED classifyBiome —
// so the world uses the exact same biome semantics as the city map and every consumer
// (paint, settlement gate, dot snap) is unchanged. Deep water is elevation-driven from
// the same seaLevel the land/water split used, so the coastline the render strokes and
// the passability the dots gate on agree to the pixel.
func (m *worldModel) buildField(seed uint32) {
	w, h := m.w, m.h
	f := m.field
	f.biomes = make([]biome, w*h)
	f.passable = make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			e := m.elevAtPx(x, y)
			mo := m.moistAtPx(x, y)
			b := m.classify(e, mo)
			idx := y*w + x
			f.biomes[idx] = b
			f.passable[idx] = passableBiome(b)
		}
	}
}

// classify maps an (elevation, moisture) pair to a biome. It anchors the water split on
// the model's seaLevel (so ocean matches the land/water cut exactly) and otherwise
// remaps the elevation above sea level onto the shared classifyBiome's [0,1] band scheme
// — the interior thus reuses the SAME grass/forest/rock/mountain/snow classifier the
// city map uses, keeping one source of truth for what a biome is.
func (m *worldModel) classify(elev, moist float64) biome {
	if elev < m.seaLevel {
		// Two-band ocean: the deepest third of the sub-sea range is deep water, the
		// shallow shelf near the coast is shallow water — so the coast reads as a lighter
		// rim. Below deep threshold → deep; above → shallow.
		if elev < m.seaLevel*0.55 {
			return biomeDeepWater
		}
		return biomeShallowWater
	}
	// Land: remap the above-sea elevation onto the shared classifier's land bands, but with
	// a GAMMA curve that keeps most of the continent as lowland (grass/forest) and reserves
	// the rock/mountain/snow bands for the genuinely high interior. A linear remap spread ~a
	// third of the land into bare rock, which reads as a uniform grey rock; the gamma pushes
	// the mid-elevations down into the grass/forest band so relief peaks stay a minority. The
	// top of the remap sits a touch above bandHighland so snow only caps the very highest
	// cells. Biome vocabulary stays identical to the city map (still classifyBiome).
	t := (elev - m.seaLevel) / (1 - m.seaLevel)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	const peakCap = bandHighland + 0.10 // top of the remap; snow (>=0.92) only near the ceiling
	shaped := math.Pow(t, 2.4)          // compress low/mid elevations well into the lowland bands
	remapped := bandShallowWater + shaped*(peakCap-bandShallowWater)
	// Moisture gamma: the coastal-wet blend leaves most land damp (all forest, no open
	// grass). Bias it drier so dry lowland reads as grassland and only the wettest ground is
	// forest — restoring green/tan variety instead of a monotone forest sheet.
	dm := math.Pow(clamp01(moist), 1.6)
	return classifyBiome(remapped, dm)
}

// buildRivers computes watercourses by flow accumulation on the low-res grid:
//
//  1. Every land cell drains to its LOWEST 8-neighbour (steepest descent). Cells at a
//     local minimum (a pit) drain to themselves and terminate a course early — rare on
//     the smoothed field, and harmless (a short stub, filtered out by the flow gate).
//  2. Process cells from HIGHEST to LOWEST elevation, pushing each cell's accumulated
//     flow (1 + inflow) to its drain target. After this pass every cell holds the total
//     upstream area draining through it — high at valley trunks, low on ridges.
//  3. SOURCES are land cells that are high AND wet AND carry enough accumulated flow.
//     From each, walk downhill along the drain pointers to the sea, emitting a pixel
//     polyline whose per-point flow drives the render's widening toward the mouth. To
//     avoid a hairball, sources are thinned so we keep only a handful of the biggest
//     trunks that actually reach open water.
func (m *worldModel) buildRivers(seed uint32) {
	gw, gh := m.gw, m.gh
	n := gw * gh
	if n == 0 {
		return
	}

	// Drain target per cell (index into the grid), and a flag for cells that are land.
	drain := make([]int, n)
	isLand := make([]bool, n)
	for i := 0; i < n; i++ {
		drain[i] = i // default: self (pit / sea)
		isLand[i] = m.elev[i] >= m.seaLevel
	}
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			i := gy*gw + gx
			if !isLand[i] {
				continue // sea cells terminate flow
			}
			lowest := m.elev[i]
			best := i
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx, ny := gx+dx, gy+dy
					if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
						continue
					}
					ni := ny*gw + nx
					if m.elev[ni] < lowest {
						lowest = m.elev[ni]
						best = ni
					}
				}
			}
			drain[i] = best
		}
	}

	// Accumulate flow: process high→low so a cell's inflow is finalized before it pushes
	// downstream. flow starts at 1 per land cell (unit rainfall).
	order := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if isLand[i] {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(a, b int) bool { return m.elev[order[a]] > m.elev[order[b]] })
	flow := make([]float64, n)
	for i := 0; i < n; i++ {
		if isLand[i] {
			flow[i] = 1
		}
	}
	for _, i := range order {
		d := drain[i]
		if d != i {
			flow[d] += flow[i]
		}
	}

	// Max flow → the threshold for "this is a river."
	maxFlow := 0.0
	for _, f := range flow {
		if f > maxFlow {
			maxFlow = f
		}
	}
	if maxFlow < 2 {
		return // no meaningful drainage (tiny canvas) — no rivers
	}
	flowThresh := maxFlow * worldRiverFlowFrac

	// Candidate sources: high + reasonably wet headwater cells that carry above-threshold
	// flow. We collect them, then pick a SPREAD-OUT handful of the biggest catchments so
	// the continent gets a few readable rivers fanning to different coasts — not a dozen
	// tributaries all converging on one mouth (which stacks into a single bright blob).
	type src struct {
		idx    int
		gx, gy int
		flow   float64
	}
	var srcs []src
	for gy := 0; gy < gh; gy++ {
		for gx := 0; gx < gw; gx++ {
			i := gy*gw + gx
			if !isLand[i] {
				continue
			}
			// A source sits high and damp: upper part of the land elevation range and
			// wet-ish, and it is a HEADWATER — small own-flow that grows into a trunk
			// downstream — so the drawn course starts near the top of the catchment.
			if m.elev[i] < m.seaLevel+0.30*(1-m.seaLevel) {
				continue // too low to be a headwater
			}
			if m.moist[i] < 0.45 {
				continue // dry ground doesn't spring a river
			}
			if flow[i] < 1.5 || flow[i] > flowThresh {
				continue // small-flow headwaters that GROW into a trunk, not mid-river cells
			}
			srcs = append(srcs, src{idx: i, gx: gx, gy: gy, flow: flow[i]})
		}
	}
	// Biggest catchments first (own-flow desc, index tiebreak for determinism) so the
	// rivers we keep are the map's most significant watercourses.
	sort.Slice(srcs, func(a, b int) bool {
		if srcs[a].flow != srcs[b].flow {
			return srcs[a].flow > srcs[b].flow
		}
		return srcs[a].idx < srcs[b].idx
	})

	// Keep a spread-out set: each kept river must have a DISTINCT mouth cell (no two rivers
	// sharing an outlet) AND a source at least minSourceGap grid cells from every already-
	// kept source (so they fan across the continent instead of clustering). Cap the count so
	// the atlas has a few clear rivers, not a web.
	const maxRivers = 5
	minSourceGap := (gw + gh) / 8 // ~an eighth of the grid span apart
	if minSourceGap < 2 {
		minSourceGap = 2
	}
	keptSrc := make([]src, 0, maxRivers)
	usedMouth := make(map[int]bool)
	for _, s := range srcs {
		if len(m.rivers) >= maxRivers {
			break
		}
		// Spacing gate: reject a source too close to one we already kept.
		tooClose := false
		for _, k := range keptSrc {
			if absInt(s.gx-k.gx)+absInt(s.gy-k.gy) < minSourceGap {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}
		course, cellIdx := m.traceCourse(drain, flow, isLand, s.idx)
		if len(course.pts) < 3 {
			continue
		}
		last := cellIdx[len(cellIdx)-1]
		if !m.cellTouchesSea(last, isLand) {
			continue // must reach open water
		}
		if usedMouth[last] {
			continue // another river already exits here — keep one mouth per outlet
		}
		// Must carry a real trunk somewhere (peak flow above threshold).
		peak := 0.0
		for _, f := range course.flow {
			if f > peak {
				peak = f
			}
		}
		if peak < flowThresh {
			continue
		}
		usedMouth[last] = true
		keptSrc = append(keptSrc, s)
		m.rivers = append(m.rivers, course)
	}
}

// traceCourse walks the drain pointers from a source cell downhill until it reaches sea
// (or a pit / revisit), emitting a pixel-space polyline plus the accumulated flow at
// each step and the list of grid-cell indices visited (for de-dup / sea-touch checks).
func (m *worldModel) traceCourse(drain []int, flow []float64, isLand []bool, start int) (river, []int) {
	gw := m.gw
	var r river
	var cells []int
	seen := make(map[int]bool)
	cur := start
	for steps := 0; steps < len(drain)+1; steps++ {
		if seen[cur] {
			break // loop guard (shouldn't happen on a DAG, but be safe)
		}
		seen[cur] = true
		gx, gy := cur%gw, cur/gw
		px := gx*worldGridPx + worldGridPx/2
		py := gy*worldGridPx + worldGridPx/2
		r.pts = append(r.pts, worldPt{x: px, y: py})
		r.flow = append(r.flow, flow[cur])
		cells = append(cells, cur)
		if !isLand[cur] {
			break // reached the sea — stop at the first water cell (the mouth)
		}
		nxt := drain[cur]
		if nxt == cur {
			break // pit / local min — terminate
		}
		cur = nxt
	}
	return r, cells
}

// cellTouchesSea reports whether grid cell idx is water or has a water 4-neighbour —
// used to confirm a traced river actually reaches open water rather than dead-ending in
// a pit on the land.
func (m *worldModel) cellTouchesSea(idx int, isLand []bool) bool {
	if idx < 0 || idx >= len(isLand) {
		return false
	}
	if !isLand[idx] {
		return true
	}
	gw, gh := m.gw, m.gh
	gx, gy := idx%gw, idx/gw
	for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := gx+nb[0], gy+nb[1]
		if nx < 0 || ny < 0 || nx >= gw || ny >= gh {
			continue
		}
		if !isLand[ny*gw+nx] {
			return true
		}
	}
	return false
}

// buildRelief samples relief-symbol anchors from the per-pixel biome field on a sparse
// stride: mountain/snow cells → mountain dabs, rock/hill cells → hill dabs, forest cells
// → forest dabs. The stride + a per-cell hash gate keep the anchors scattered (not every
// eligible pixel) so the relief reads as a handful of symbols per region, not a texture.
func (m *worldModel) buildRelief(seed uint32) {
	w, h := m.w, m.h
	if w <= 0 || h <= 0 || m.field == nil || len(m.field.biomes) == 0 {
		return
	}
	// Stride roughly every 3 grid cells so relief is sparse; jitter each anchor within
	// its stride cell so symbols don't form a visible lattice.
	stride := worldGridPx * 3
	if stride < 4 {
		stride = 4
	}
	rSeed := seed ^ 0x9E3779B1
	for y := stride / 2; y < h; y += stride {
		for x := stride / 2; x < w; x += stride {
			// Jitter within the stride cell.
			jx := x + int((hashUnit(uint32(x), uint32(y), rSeed)-0.5)*float64(stride)*0.6)
			jy := y + int((hashUnit(uint32(y), uint32(x), rSeed^0x55)-0.5)*float64(stride)*0.6)
			if jx < 0 || jy < 0 || jx >= w || jy >= h {
				continue
			}
			b := m.field.at(jx, jy)
			var kind reliefKind
			switch b {
			case biomeMountain, biomeSnow:
				kind = reliefMountain
			case biomeRock:
				kind = reliefHill
			case biomeForest:
				kind = reliefForest
			default:
				continue // grass/sand/water carry no relief symbol
			}
			// Thin further with a hash gate so even eligible regions stay sparse: keep
			// ~55% of candidates. Forests a touch denser than peaks so tree cover reads.
			gate := 0.45
			if kind == reliefForest {
				gate = 0.35
			}
			if hashUnit(uint32(jx), uint32(jy), rSeed^0xABCD) < gate {
				continue
			}
			m.reliefs = append(m.reliefs, reliefAnchor{x: jx, y: jy, kind: kind})
		}
	}
}

// ---- sampling helpers -------------------------------------------------------

// normPos maps a pixel coord to a normalized position in [-1,1] centered on the canvas,
// scaled by the SHORTER half-axis so the island mask is round-ish on any aspect (a wide
// canvas still gets ocean top+bottom, not a squashed oval). Returns (nx, ny).
func (m *worldModel) normPos(px, py float64) (float64, float64) {
	cx := float64(m.w) / 2
	cy := float64(m.h) / 2
	half := math.Min(cx, cy)
	if half < 1 {
		half = 1
	}
	return (px - cx) / half, (py - cy) / half
}

// elevAtPx samples the low-res elevation grid at a pixel coord with bilinear
// interpolation, so the upsampled continent has a smooth coastline. Out-of-range clamps
// to the nearest cell.
func (m *worldModel) elevAtPx(px, py int) float64 { return m.sampleBilinear(m.elev, px, py) }

// moistAtPx samples the low-res moisture grid at a pixel coord with bilinear
// interpolation. Out-of-range clamps to the nearest cell.
func (m *worldModel) moistAtPx(px, py int) float64 { return m.sampleBilinear(m.moist, px, py) }

// sampleBilinear bilinearly interpolates a low-res grid field at pixel (px,py). The
// grid cell centers sit at pixel (gx*worldGridPx + worldGridPx/2), so we map the pixel
// back to fractional grid coords and lerp the four surrounding cells. Clamped at the
// edges. Empty/degenerate grids return 0.
func (m *worldModel) sampleBilinear(grid []float64, px, py int) float64 {
	gw, gh := m.gw, m.gh
	if gw <= 0 || gh <= 0 || len(grid) < gw*gh {
		return 0
	}
	// Fractional grid position (cell centers at +worldGridPx/2).
	fx := (float64(px) - worldGridPx/2) / worldGridPx
	fy := (float64(py) - worldGridPx/2) / worldGridPx
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	tx := fx - float64(x0)
	ty := fy - float64(y0)
	at := func(gx, gy int) float64 {
		if gx < 0 {
			gx = 0
		}
		if gy < 0 {
			gy = 0
		}
		if gx >= gw {
			gx = gw - 1
		}
		if gy >= gh {
			gy = gh - 1
		}
		return grid[gy*gw+gx]
	}
	v00 := at(x0, y0)
	v10 := at(x0+1, y0)
	v01 := at(x0, y0+1)
	v11 := at(x0+1, y0+1)
	a := v00*(1-tx) + v10*tx
	b := v01*(1-tx) + v11*tx
	return a*(1-ty) + b*ty
}

// ---- small math -------------------------------------------------------------

// islandMask returns a broad, gentle radial envelope in [0,1]: 1 for r ≤ landCore, easing
// to 0 at r ≥ edge. It is NOT the coastline (the warped shape-noise carves that) — it only
// biases "there is a continent near the center" and "open sea at the frame border." A wide
// core + long shoulder keeps the envelope soft so it can't re-round the noisy coast, while
// still guaranteeing the landmass doesn't run off every edge.
func islandMask(r float64) float64 {
	const landCore = 0.34 // gentle plateau near the center
	const edge = 1.25     // fully ocean only well past the frame border
	if r <= landCore {
		return 1
	}
	if r >= edge {
		return 0
	}
	t := (r - landCore) / (edge - landCore) // 0..1 across the shoulder
	// Cosine ease from 1→0.
	return 0.5 * (1 + math.Cos(t*math.Pi))
}

// signFromHash maps a [0,1) hash to -1 or +1 (split at 0.5) so lobes alternate between
// peninsulas (+) and bays (-).
func signFromHash(u float64) float64 {
	if u < 0.5 {
		return -1
	}
	return 1
}

// clamp01 clamps v to [0,1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
