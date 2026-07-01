package citymap

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// citygen.go is the count-driven city synthesizer (citymap v2 — see
// design-and-architecture/city-synthesis.md). It supersedes the old per-era
// placement STRATEGIES (layout.go): rather than dropping one marker per built type
// and stringing MST roads between them, it grows a whole top-down CITY whose form
// and density derive from the player's actual building COUNTS and current AGE.
//
// The generator is pure and deterministic from a seed. Its output is a cityPlan —
// streets + blocks + lots — that the render path paints in order (streets → block
// interiors → building lots), then reads back the landmark lots to build the same
// layoutGeometry the overlay already consumes, so the named hero buildings (Shrine,
// Gathering Camp, …) stay labeled exactly as before, now embedded in the fabric.
//
// Phasing (city-synthesis.md §Phasing): this is Phase A — the framework plus the
// `organic` village and `ancient` gridded-Mesopotamia eras fully tuned; the other
// five era bands fall back to defaultEraStyle (a plain grid) until Phase B/C tunes
// them. Every era still renders, terrain-gated and panic-safe.

// ---- model ------------------------------------------------------------------

// point is an integer pixel coordinate on the render canvas.
type point struct{ x, y int }

// lotKind classifies a placed parcel so the renderer knows how to paint it and the
// geometry pass knows which lots are labeled landmarks.
type lotKind int

const (
	lotHouse    lotKind = iota // a dwelling — the count-scaled residential fabric
	lotWorkshop                // a production structure in the production zone
	lotGarden                  // greenery filling leftover block area
	lotPlaza                   // a paved civic open space (stone tone)
	lotWall                    // a defensive wall segment (era extra)
	lotTower                   // a keep / tower (era extra)
	lotLandmark                // a named hero building — carries the overlay label
)

// street is one polyline of the road network with a width class. pts are pixel
// waypoints (already terrain-routed, so the polyline never crosses water); width is
// the half-thickness in pixels (0 → 1px, 1 → 3px band); paved marks the wider
// avenues so the renderer can tint them a touch brighter than alleys.
type street struct {
	pts   []point
	width int
	paved bool
}

// block is a rectangular parcel bounded by streets. Grid eras carve clean rects;
// the organic era uses a few loose blocks around the cluster centers. Lots are
// placed inside blocks; leftover interior area becomes gardens.
type block struct {
	x, y, w, h int
}

// lot is one placed structure (or open parcel) in the plan. rect is (x,y,w,h) in
// pixels; kind drives the paint + whether it is a labeled landmark; domain drives
// the lineage color (via lineageColor); tier nudges the volume size / style; label
// is "" except on lotLandmark, which carries the building's config Name so the
// overlay can name it.
type lot struct {
	x, y, w, h int
	kind       lotKind
	domain     string // lineageKey (or special category) → lineageColor
	category   string // building category, for the landmark's label color source
	tier       int    // LineageTier — style + label priority
	label      string // landmark name; "" otherwise
}

// cityPlan is the render-ready output of the generator: the street network, the
// parcels between streets, and the structures filling them.
type cityPlan struct {
	streets []street
	blocks  []block
	lots    []lot
}

// ---- era style presets ------------------------------------------------------

// streetPattern names the network topology an era grows.
type streetPattern int

const (
	patternOrganic    streetPattern = iota // meandering dirt paths cluster→cluster
	patternRadial                          // spokes from a central plaza + a coarse grid
	patternGrid                            // plain orthogonal grid (the default fallback)
	patternSuperblock                      // wide-avenue superblock grid (Phase B/C)
	patternCampus                          // hex/campus pods + links (Phase C)
	patternOrbital                         // concentric rings (Phase C)
)

// houseStyle names the residential silhouette an era uses. Phase A renders houses
// as small volumes regardless; the style is carried so Phase B/C can specialize the
// footprint (hut → mudbrick → timber → rowhouse → tower → arcology).
type houseStyle int

const (
	houseHut houseStyle = iota
	houseMudbrick
	houseTimber
	houseRowhouse
	houseTower
	houseArcology
	houseDome
)

// eraStyle is the single struct that drives an era's whole silhouette: the street
// topology + widths, block coarseness, how organic (jittered) vs. strict the grid
// is, how much leftover area becomes garden, the era extras (walls/keep/plaza), the
// house style, and an overall density multiplier on the count→house scaling.
type eraStyle struct {
	streetPattern streetPattern
	mainWidth     int     // half-thickness of avenues (0 → 1px, 1 → 3px)
	alleyWidth    int     // half-thickness of minor streets/alleys
	blockSize     int     // target block edge in pixels (grid granularity)
	gridJitter    float64 // 0 = strict grid … 1 = fully organic wander (px = jitter*blockSize)
	gardenRatio   float64 // fraction of eligible leftover parcels turned to gardens
	hasWalls      bool    // draw a perimeter wall ring (era extra)
	hasKeep       bool    // place a central keep/tower lot
	hasPlaza      bool    // place a central plaza lot under the hero landmark
	houseStyle    houseStyle
	density       float64 // multiplier on residential count → house population
}

// defaultEraStyle is the sane fallback for the five era bands Phase A does not yet
// tune (castle, zonedgrid, cityblocks, campus, orbital). A plain, slightly-jittered
// orthogonal grid with modest gardens — it renders a legible city without pretending
// to be its final silhouette. Phase B/C replaces these bands with bespoke presets.
var defaultEraStyle = eraStyle{
	streetPattern: patternGrid,
	mainWidth:     1,
	alleyWidth:    0,
	blockSize:     14,
	gridJitter:    0.10,
	gardenRatio:   0.20,
	hasWalls:      false,
	hasKeep:       false,
	hasPlaza:      false,
	houseStyle:    houseRowhouse,
	density:       1.0,
}

// eraStyles maps each era band to its preset. Phase A tunes organic + ancient
// (eraOrganic, eraHubSpoke); every other band resolves to defaultEraStyle via
// eraStyleFor. Keyed off the existing eraForAge buckets so the age→style mapping is
// unchanged and future ages slot in through eraForAge.
var eraStyles = map[era]eraStyle{
	// ORGANIC (primitive, stone): a small lived-in village. Meandering thin dirt
	// paths link a handful of hut clusters and the landmarks; garden patches dot the
	// gaps; no walls, no plaza, no grid. Density is generous so a few housing
	// buildings still read as a real cluster of huts rather than two lonely dots.
	eraOrganic: {
		streetPattern: patternOrganic,
		mainWidth:     0, // dirt paths are thin (1px)
		alleyWidth:    0,
		blockSize:     10, // loose clusters, small
		gridJitter:    1.0,
		gardenRatio:   0.35,
		hasWalls:      false,
		hasKeep:       false,
		hasPlaza:      false,
		houseStyle:    houseHut,
		density:       1.3,
	},
	// ANCIENT (bronze, iron, classical): Mesopotamia. A central plaza with a
	// ziggurat/temple hero on it, a few radiating avenues, and a coarse grid of
	// mudbrick house blocks with courtyard gardens. Wider streets than the village;
	// clearly a small CITY, a visible step up. Courtyard gardens, central plaza on.
	eraHubSpoke: {
		streetPattern: patternRadial,
		mainWidth:     1, // avenues are 3px
		alleyWidth:    0, // side lanes 1px
		blockSize:     16,
		gridJitter:    0.28, // slightly irregular mudbrick blocks, not machine-straight
		gardenRatio:   0.22,
		hasWalls:      false,
		hasKeep:       false,
		hasPlaza:      true,
		houseStyle:    houseMudbrick,
		density:       1.0,
	},
}

// eraStyleFor returns the tuned preset for an era, or defaultEraStyle for the bands
// Phase A leaves on the fallback. Centralizing the lookup keeps generateCityPlan
// agnostic to which eras are tuned yet.
func eraStyleFor(e era) eraStyle {
	if s, ok := eraStyles[e]; ok {
		return s
	}
	return defaultEraStyle
}

// ---- gather -----------------------------------------------------------------

// cityRole buckets a building by its function in the city fabric. Residential
// counts inflate into houses; production into workshops; landmarks (civic / faith /
// knowledge / wonder / storage) become the labeled hero lots the city grows around.
type cityRole int

const (
	roleResidential cityRole = iota
	roleProduction
	roleLandmark
)

// builtBuilding is one distinct built type gathered for synthesis: its config key,
// lineage/category (for color + role), tier, and instance count (drives population).
type builtBuilding struct {
	key      string
	name     string
	domain   string // lineageKey
	category string
	tier     int
	count    int
	role     cityRole
}

// classifyRole maps a building's category+lineage to its city role. Housing is
// residential; wonders/monuments/faith/knowledge/research/culture/diplomacy/storage
// read as civic landmarks the city centers on; everything else (the production
// lineages) is production. Pure data — safe on the render path.
func classifyRole(category, lineageKey string) cityRole {
	switch category {
	case "housing":
		return roleResidential
	case "wonder", "monument", "storage", "diplomacy", "research":
		return roleLandmark
	}
	switch lineageKey {
	case "faith", "knowledge", "culture_arts":
		return roleLandmark
	}
	return roleProduction
}

// gatherBuildings turns the built-building set into the synthesis input: one
// builtBuilding per distinct built type, classified into a city role, sorted for
// determinism. It reuses the same builtDistricts grouping the old path used only to
// resolve names/lineage/tier consistently, then flattens to the per-type list the
// generator populates from. byKey is the pure config table (no locks).
func gatherBuildings(state game.GameState, byKey map[string]config.BuildingDef) []builtBuilding {
	keys, counts := builtBuildingKeys(state)
	out := make([]builtBuilding, 0, len(keys))
	for _, k := range keys {
		lineage, category, _ := buildingMeta(byKey, k)
		def := byKey[k]
		name := def.Name
		if name == "" {
			name = titleCaseKey(k)
		}
		domain := lineage
		// Special categories color by category (via lineageColor), so carry the
		// category as the domain when the lineage is empty/irrelevant, matching the
		// old district grouping.
		switch category {
		case "wonder", "monument", "storage", "diplomacy":
			domain = category
		}
		if domain == "" {
			domain = "misc"
		}
		out = append(out, builtBuilding{
			key:      k,
			name:     name,
			domain:   domain,
			category: category,
			tier:     def.LineageTier,
			count:    counts[k],
			role:     classifyRole(category, lineage),
		})
	}
	return out
}

// ---- size -------------------------------------------------------------------

// cityMetrics is the sized envelope the generator lays streets and lots within:
// a center, a radius bounding the built-up area, and derived counts. All in pixels.
type cityMetrics struct {
	cx, cy int
	radius int
}

// sizeCity derives the built-up envelope from the total building count with sqrt
// scaling, so both a ~5-building hamlet and a ~300-building metropolis fit the same
// canvas legibly (counts are REPRESENTATIVE, not 1:1). The radius grows with
// sqrt(total) toward — but never past — the canvas half-min, leaving a terrain
// margin; a tiny empire still gets a minimum radius so the village isn't a dot.
func sizeCity(total, w, h int) cityMetrics {
	half := minInt(w, h) / 2
	maxR := half - 3
	if maxR < 4 {
		maxR = 4
	}
	// sqrt scaling: radius ~ k*sqrt(total), clamped into [minR, maxR]. The constant
	// is tuned so ~6 buildings ≈ 40% of maxR and ~200 buildings saturates maxR.
	minR := maxR / 3
	if minR < 4 {
		minR = 4
	}
	r := minR
	if total > 0 {
		grown := int(float64(maxR) * (0.35 + 0.09*math.Sqrt(float64(total))))
		r = clampInt(grown, minR, maxR)
	}
	return cityMetrics{cx: w / 2, cy: h / 2, radius: r}
}

// houseCountFor maps a residential building's instance count to how many house lots
// it contributes, sqrt-scaled by density so many-stamped housing reads as a bigger
// neighbourhood without a 1:1 pixel explosion. At least 1 house per residential
// type so a lone hut still appears.
func houseCountFor(count int, density float64) int {
	if count <= 0 {
		return 0
	}
	n := int(density * (1.0 + 1.6*math.Sqrt(float64(count))))
	if n < 1 {
		n = 1
	}
	if n > 40 {
		n = 40 // legibility cap: a metropolis of huts, not confetti
	}
	return n
}

// workshopCountFor maps a production building's count to workshop lots — a gentler
// sqrt than housing (workshops are bulkier and zoned), min 1 so every production
// type shows at least one shed.
func workshopCountFor(count int, density float64) int {
	if count <= 0 {
		return 0
	}
	n := int(density * (1.0 + 1.0*math.Sqrt(float64(count))))
	if n < 1 {
		n = 1
	}
	if n > 24 {
		n = 24
	}
	return n
}

// ---- generate ---------------------------------------------------------------

// generateCityPlan synthesizes the whole cityPlan for one render: pure and fully
// deterministic from seed. Pipeline (city-synthesis.md §Pipeline):
//
//	(a) gather   — built buildings → per-type role/domain/tier/count.
//	(b) size     — sqrt-scaled built-up radius from the total count.
//	(c) streets  — the era pattern, terrain-routed so no street crosses water.
//	(d) blocks   — parcels derived between the streets.
//	(e) populate — count-driven lots: houses / workshops / landmarks / gardens +
//	    era extras (walls / keep / plaza).
//	(f) gate     — every lot snapped onto passable land.
//
// field may be nil (tests without terrain) — the routing + gating degrade to a
// straight/identity pass. w,h are the render pixel dimensions.
func generateCityPlan(state game.GameState, byKey map[string]config.BuildingDef, field *terrainField, e era, seed uint32, w, h int) cityPlan {
	plan := cityPlan{}
	if w <= 0 || h <= 0 {
		return plan
	}
	style := eraStyleFor(e)
	r := newRNG(seed)

	// (a) gather.
	blds := gatherBuildings(state, byKey)

	// (b) size.
	total := 0
	for _, b := range blds {
		total += b.count
	}
	m := sizeCity(total, w, h)

	// Snap the city center onto passable land so the whole city (and every street that
	// radiates from it) starts on solid ground. If the center is drowned, the streets
	// and the central plaza/hero would otherwise anchor in a lake.
	if field != nil {
		if nx, ny, ok := nearestPassablePx(field, m.cx, m.cy); ok {
			m.cx, m.cy = nx, ny
		}
	}

	// Terrain routing grid (nil-safe): built once from the field so streets bend
	// around water/mountain. Seed reuse keeps a city routing identically frame to
	// frame. A nil field yields a nil grid → straight-line streets (tests).
	var grid *pfCostGrid
	if field != nil {
		grid = buildCostGrid(field, seed)
	}

	// (c) streets + (d) blocks, per era pattern.
	switch style.streetPattern {
	case patternOrganic:
		plan.streets, plan.blocks = organicStreets(m, blds, style, grid, field, r, w, h)
	case patternRadial:
		plan.streets, plan.blocks = radialStreets(m, style, grid, field, r, w, h)
	default: // patternGrid and the not-yet-tuned patterns fall back to a plain grid.
		plan.streets, plan.blocks = gridStreets(m, style, grid, field, r, w, h)
	}

	// (e) populate the blocks, count-driven.
	plan.lots = populate(m, blds, plan.blocks, style, r, w, h)

	// (f) terrain-gate every lot onto passable land (snap off water).
	if field != nil {
		gateLots(plan.lots, field)
	}
	return plan
}

// ---- street generators ------------------------------------------------------

// fieldSegClear reports whether the straight pixel segment a→b stays entirely on
// passable land at 1px resolution, walking the segment by its dominant axis. This is
// the FINE clearance test (the pathfinder's segClear samples at node granularity,
// ~3px, which can miss a 1px water sliver); streets validate against this so no
// street pixel is ever painted on water — the §terrain-gate guarantee for streets.
func fieldSegClear(f *terrainField, a, b point) bool {
	if f == nil {
		return true // no terrain (tests): nothing to avoid
	}
	steps := absInt(b.x-a.x) + absInt(b.y-a.y)
	if steps == 0 {
		return f.passableAt(a.x, a.y)
	}
	for i := 0; i <= steps; i++ {
		x := a.x + (b.x-a.x)*i/steps
		y := a.y + (b.y-a.y)*i/steps
		if !f.passableAt(x, y) {
			return false
		}
	}
	return true
}

// routeStreet turns a straight A→B intent into a terrain-routed street polyline that
// NEVER crosses water. It asks the cost grid for an A* route (node centers, all on
// passable land), then simplifies to as few segments as possible using a 1px field
// clearance test: a chord is kept only when it is field-clear, otherwise the route
// falls back to the finer A* waypoints between those points. Any residual raw segment
// that still fails the 1px test (a diagonal node hop grazing a 1px sliver) is dropped
// rather than drawn, leaving an invisible gap instead of a pixel on water. With no
// grid/field (tests) it degrades to the straight two-point segment.
func routeStreet(grid *pfCostGrid, field *terrainField, a, b point, width int, paved bool) street {
	if grid == nil {
		return street{pts: []point{a, b}, width: width, paved: paved}
	}
	raw, ok := grid.findPath(a.x, a.y, b.x, b.y)
	if !ok || len(raw) < 2 {
		// A* found nothing (endpoint walled off). Only emit a straight fallback if it
		// is itself field-clear — never draw a fallback across water.
		if fieldSegClear(field, a, b) {
			return street{pts: []point{a, b}, width: width, paved: paved}
		}
		return street{width: width, paved: paved} // no clear route — empty (drawn as nothing)
	}
	pts := make([]point, len(raw))
	for i, p := range raw {
		pts[i] = point{p[0], p[1]}
	}
	return street{pts: simplifyStreet(pts, field), width: width, paved: paved}
}

// simplifyStreet greedily collapses a dense pixel polyline to a few segments whose
// every chord is field-clear at 1px, returning ONE connected polyline. From each kept
// point it extends to the farthest later point whose straight chord stays on land. If
// even the immediate next hop grazes water (a diagonal node step over a 1px sliver),
// the street is TRUNCATED there rather than jumping across the gap — drawStreet
// connects consecutive points, so every emitted segment must be clear. A shorter
// street never crosses water; the network stays connected via the other streets.
func simplifyStreet(pts []point, field *terrainField) []point {
	if len(pts) < 2 || field == nil {
		return pts
	}
	out := []point{pts[0]}
	i := 0
	for i < len(pts)-1 {
		// Farthest j > i whose chord pts[i]→pts[j] is field-clear at 1px.
		j := -1
		for k := len(pts) - 1; k > i; k-- {
			if fieldSegClear(field, pts[i], pts[k]) {
				j = k
				break
			}
		}
		if j < 0 {
			// Not even the next point is reachable on land — truncate here.
			break
		}
		out = append(out, pts[j])
		i = j
	}
	return out
}

// organicStreets grows the village network: a handful of hut-cluster anchors are
// scattered on a drunk walk out from the center, and a meandering dirt path links
// each anchor back toward the center (and thus, via shared trunks near the hub, to
// its neighbours). One loose block is emitted per anchor so populate has cluster
// parcels to fill. Everything terrain-routed so a path never fords a lake.
func organicStreets(m cityMetrics, blds []builtBuilding, style eraStyle, grid *pfCostGrid, field *terrainField, r *rng, w, h int) ([]street, []block) {
	// Cluster count tracks the settlement size but stays small — a village is a few
	// clusters, not a grid. One per ~2 built types, clamped to [2,7].
	nClusters := len(blds)/2 + 2
	nClusters = clampInt(nClusters, 2, 7)

	center := point{m.cx, m.cy}
	streets := make([]street, 0, nClusters)
	blocks := make([]block, 0, nClusters+1)

	// A central block so the hero landmark + plaza have a home even with no clusters.
	bs := style.blockSize
	blocks = append(blocks, centeredBlock(m.cx, m.cy, bs+2, bs+2, w, h))

	for i := 0; i < nClusters; i++ {
		ang := r.f01() * 2 * math.Pi
		step := float64(m.radius) * (0.4 + 0.55*r.f01())
		ax := clampInt(m.cx+int(math.Cos(ang)*step), 2, w-3)
		ay := clampInt(m.cy+int(math.Sin(ang)*step), 2, h-3)
		streets = append(streets, routeStreet(grid, field, point{ax, ay}, center, style.mainWidth, false))
		blocks = append(blocks, centeredBlock(ax, ay, bs, bs, w, h))
	}
	return streets, blocks
}

// radialStreets grows the ancient city: a few avenues radiate from the central
// plaza to the built-up edge, and a coarse grid of cross-streets carves the mudbrick
// blocks between them. Wider (paved) avenues than the village; the grid is lightly
// jittered so blocks read as hand-laid mudbrick, not machine-straight. Blocks are
// the grid cells (excluding the very center, reserved for the plaza + ziggurat).
func radialStreets(m cityMetrics, style eraStyle, grid *pfCostGrid, field *terrainField, r *rng, w, h int) ([]street, []block) {
	center := point{m.cx, m.cy}
	streets := make([]street, 0, 8)

	// Radiating avenues (paved, wider) at evenly-spaced angles with a slight offset.
	const spokes = 6
	for i := 0; i < spokes; i++ {
		ang := 2*math.Pi*float64(i)/float64(spokes) + 0.2
		ex := clampInt(m.cx+int(math.Cos(ang)*float64(m.radius)), 1, w-2)
		ey := clampInt(m.cy+int(math.Sin(ang)*float64(m.radius)), 1, h-2)
		streets = append(streets, routeStreet(grid, field, center, point{ex, ey}, style.mainWidth, true))
	}

	// A coarse orthogonal grid over the built-up square, jittered. This both draws
	// the cross-streets and defines the mudbrick blocks between them.
	gstreets, blocks := gridWithin(m, style, grid, field, r, w, h, true)
	streets = append(streets, gstreets...)
	return streets, blocks
}

// gridStreets is the default/fallback network: a plain orthogonal grid over the
// built-up square, lightly jittered by the era's gridJitter. Used for every era band
// not yet tuned (and as the ancient era's block-carving grid). Blocks are the grid
// cells.
func gridStreets(m cityMetrics, style eraStyle, grid *pfCostGrid, field *terrainField, r *rng, w, h int) ([]street, []block) {
	return gridWithin(m, style, grid, field, r, w, h, false)
}

// gridWithin lays an orthogonal street grid over the built-up square [center ±
// radius] at blockSize spacing, jittered by gridJitter, and returns both the street
// polylines and the block rects between them. Avenues (every other line, when
// widerMains is set) are paved + mainWidth; the rest are alleyWidth. Each grid line
// is terrain-routed end-to-end so it detours around water instead of crossing it.
// The center block is skipped from the returned blocks (reserved for the hero /
// plaza) when reserveCenter callers want it; here we return all cells and let
// populate reserve the center by proximity.
func gridWithin(m cityMetrics, style eraStyle, grid *pfCostGrid, field *terrainField, r *rng, w, h int, widerMains bool) ([]street, []block) {
	bs := style.blockSize
	if bs < 4 {
		bs = 4
	}
	x0 := clampInt(m.cx-m.radius, 1, w-2)
	y0 := clampInt(m.cy-m.radius, 1, h-2)
	x1 := clampInt(m.cx+m.radius, 1, w-2)
	y1 := clampInt(m.cy+m.radius, 1, h-2)
	if x1-x0 < bs || y1-y0 < bs {
		// Too small to grid — a single block covering the area, no interior streets.
		return nil, []block{{x: x0, y: y0, w: maxInt(x1-x0, 1), h: maxInt(y1-y0, 1)}}
	}

	jit := int(style.gridJitter * float64(bs) * 0.5)
	jitter := func() int {
		if jit <= 0 {
			return 0
		}
		return r.span(jit)
	}

	// Vertical grid line x-positions and horizontal y-positions, jittered.
	xs := make([]int, 0)
	for x := x0; x <= x1; x += bs {
		xs = append(xs, clampInt(x+jitter(), x0, x1))
	}
	if xs[len(xs)-1] < x1 {
		xs = append(xs, x1)
	}
	ys := make([]int, 0)
	for y := y0; y <= y1; y += bs {
		ys = append(ys, clampInt(y+jitter(), y0, y1))
	}
	if ys[len(ys)-1] < y1 {
		ys = append(ys, y1)
	}

	streets := make([]street, 0, len(xs)+len(ys))
	for i, x := range xs {
		width, paved := style.alleyWidth, false
		if widerMains && i%2 == 0 {
			width, paved = style.mainWidth, true
		}
		streets = append(streets, routeStreet(grid, field, point{x, y0}, point{x, y1}, width, paved))
	}
	for j, y := range ys {
		width, paved := style.alleyWidth, false
		if widerMains && j%2 == 0 {
			width, paved = style.mainWidth, true
		}
		streets = append(streets, routeStreet(grid, field, point{x0, y}, point{x1, y}, width, paved))
	}

	// Blocks: the rects between consecutive grid lines, inset by 1px so a lot never
	// paints directly on the street pixel.
	blocks := make([]block, 0, len(xs)*len(ys))
	for i := 0; i+1 < len(xs); i++ {
		for j := 0; j+1 < len(ys); j++ {
			bx := xs[i] + 1
			by := ys[j] + 1
			bw := xs[i+1] - xs[i] - 1
			bh := ys[j+1] - ys[j] - 1
			if bw < 2 || bh < 2 {
				continue
			}
			blocks = append(blocks, block{x: bx, y: by, w: bw, h: bh})
		}
	}
	return streets, blocks
}

// centeredBlock builds a block rect of (bw×bh) centered on (cx,cy), clamped so it
// stays on-canvas. Used by the organic era where blocks are loose cluster hulls
// rather than grid cells.
func centeredBlock(cx, cy, bw, bh, w, h int) block {
	x := clampInt(cx-bw/2, 1, maxInt(w-2, 1))
	y := clampInt(cy-bh/2, 1, maxInt(h-2, 1))
	if x+bw > w-1 {
		bw = maxInt(w-1-x, 1)
	}
	if y+bh > h-1 {
		bh = maxInt(h-1-y, 1)
	}
	return block{x: x, y: y, w: bw, h: bh}
}

// ---- populate (count-driven) ------------------------------------------------

// populate distributes lots into the blocks, driven by building COUNTS:
//   - each landmark building → one lotLandmark at a prominent anchor (center-most
//     block first) carrying its label;
//   - each residential building → houseCountFor(count) lotHouse spread across the
//     residential blocks;
//   - each production building → workshopCountFor(count) lotWorkshop in the
//     production zone (blocks on the far side from the center);
//   - leftover eligible blocks → lotGarden by gardenRatio;
//   - era extras: a central plaza (hasPlaza), a keep (hasKeep), a wall ring
//     (hasWalls) — appended so they paint over the fabric.
//
// Deterministic: block ordering is stable and the rng is threaded, so the same seed
// yields the same city. All coordinates are clamped on-canvas; gating happens after.
func populate(m cityMetrics, blds []builtBuilding, blocks []block, style eraStyle, r *rng, w, h int) []lot {
	lots := make([]lot, 0, 64)
	if len(blocks) == 0 {
		return lots
	}

	// Order blocks by distance from the city center so landmarks take the innermost
	// parcels and production is pushed outward. Stable sort keeps determinism.
	type ranked struct {
		b        block
		occupied bool // carries any structure lot (blocks a garden)
	}
	rb := make([]ranked, len(blocks))
	dists := make([]float64, len(blocks))
	idxByDist := make([]int, len(blocks))
	for i, b := range blocks {
		rb[i] = ranked{b: b}
		bcx := float64(b.x + b.w/2)
		bcy := float64(b.y + b.h/2)
		dists[i] = math.Hypot(bcx-float64(m.cx), bcy-float64(m.cy))
		idxByDist[i] = i
	}
	sort.SliceStable(idxByDist, func(i, j int) bool { return dists[idxByDist[i]] < dists[idxByDist[j]] })

	// Split builtBuildings by role, preserving the gather order (already sorted).
	var landmarks, residential, production []builtBuilding
	for _, b := range blds {
		switch b.role {
		case roleLandmark:
			landmarks = append(landmarks, b)
		case roleResidential:
			residential = append(residential, b)
		default:
			production = append(production, b)
		}
	}

	// A settlement with no civic landmark (e.g. an early village of only huts +
	// gathering camps) would otherwise have no named hero for the overlay to label.
	// Promote its single most prominent PRODUCTION building (highest tier, then most-
	// built) to a landmark so the village always has one labeled anchor — this is why
	// a hut-and-camp village still reads "Gathering Camp" on its central structure.
	if len(landmarks) == 0 && len(production) > 0 {
		bestIdx := 0
		for i := 1; i < len(production); i++ {
			if production[i].tier > production[bestIdx].tier ||
				(production[i].tier == production[bestIdx].tier && production[i].count > production[bestIdx].count) {
				bestIdx = i
			}
		}
		hero := production[bestIdx]
		landmarks = append(landmarks, hero)
		production = append(production[:bestIdx], production[bestIdx+1:]...)
	}

	// --- landmarks: innermost blocks, one hero lot each, labeled (EXCLUSIVE). ---
	// The single most prominent landmark (highest tier) takes the dead-center anchor
	// so the plaza/ziggurat reads as the city's heart; the rest fan out inward-first.
	// A landmark's block is claimed exclusively (no house/workshop shares it) so a
	// hero volume stands clear. landmarkBlocks tracks which ranked indices are heroes.
	sort.SliceStable(landmarks, func(i, j int) bool {
		if landmarks[i].tier != landmarks[j].tier {
			return landmarks[i].tier > landmarks[j].tier
		}
		return landmarks[i].name < landmarks[j].name
	})
	landmarkBlock := make([]bool, len(rb))
	nextInner := 0 // cursor into idxByDist for the next free inner block
	for li, b := range landmarks {
		var lx, ly int
		if nextInner < len(idxByDist) {
			bk := idxByDist[nextInner]
			nextInner++
			landmarkBlock[bk] = true
			rb[bk].occupied = true
			lx = rb[bk].b.x + rb[bk].b.w/2
			ly = rb[bk].b.y + rb[bk].b.h/2
		} else {
			// More landmarks than blocks — ring the overflow around the center.
			ang := 2 * math.Pi * float64(li) / math.Max(1, float64(len(landmarks)))
			lx = clampInt(m.cx+int(math.Cos(ang)*float64(m.radius)*0.5), 2, w-3)
			ly = clampInt(m.cy+int(math.Sin(ang)*float64(m.radius)*0.5), 2, h-3)
		}
		lots = append(lots, lot{
			x: lx, y: ly, w: landmarkSize(b.tier), h: landmarkSize(b.tier),
			kind: lotLandmark, domain: b.domain, category: b.category, tier: b.tier, label: b.name,
		})
	}

	// Non-landmark blocks, still ordered center→out. Houses and workshops SHARE these
	// (a block holds several structures) so scarcity never starves a role to zero: the
	// production zone is the outer slice, residential the inner slice, but both fall
	// back to the whole pool when their slice is empty (tiny villages, few blocks).
	fabric := make([]int, 0, len(idxByDist))
	for _, bk := range idxByDist {
		if !landmarkBlock[bk] {
			fabric = append(fabric, bk)
		}
	}
	split := len(fabric) / 2
	innerPool := fabric[:split] // nearer center → residential
	outerPool := fabric[split:] // farther out → production
	if len(innerPool) == 0 {
		innerPool = fabric
	}
	if len(outerPool) == 0 {
		outerPool = fabric
	}

	// mark records that a block carried a structure (so it won't also become a garden).
	place := func(pool []int, cursor *int, kind lotKind, b builtBuilding, footprint int) {
		if len(pool) == 0 {
			return
		}
		bk := pool[*cursor%len(pool)]
		*cursor++
		rb[bk].occupied = true
		px, py := scatterInBlock(rb[bk].b, r)
		lots = append(lots, lot{
			x: px, y: py, w: footprint, h: footprint, kind: kind,
			domain: b.domain, category: b.category, tier: b.tier,
		})
	}

	// --- production: workshops cycle the outer pool. ------------------------
	pbCursor := 0
	for _, b := range production {
		n := workshopCountFor(b.count, style.density)
		for j := 0; j < n; j++ {
			place(outerPool, &pbCursor, lotWorkshop, b, 1)
		}
	}

	// --- residential: houses cycle the inner pool. --------------------------
	rbCursor := 0
	for _, b := range residential {
		n := houseCountFor(b.count, style.density)
		for j := 0; j < n; j++ {
			place(innerPool, &rbCursor, lotHouse, b, 0)
		}
	}

	// --- gardens: a fraction of the blocks that took NO structure turn green. ---
	if style.gardenRatio > 0 {
		freeBlocks := make([]int, 0)
		for _, bk := range idxByDist {
			if !rb[bk].occupied {
				freeBlocks = append(freeBlocks, bk)
			}
		}
		want := int(float64(len(freeBlocks)) * style.gardenRatio)
		for i := 0; i < want && i < len(freeBlocks); i++ {
			b := rb[freeBlocks[i]].b
			lots = append(lots, lot{x: b.x, y: b.y, w: b.w, h: b.h, kind: lotGarden})
		}
	}

	// --- era extras ---------------------------------------------------------
	if style.hasPlaza {
		// A plaza tile under the center (drawn before the hero landmark already placed
		// there paints on top via lot ordering in the renderer). Modest square.
		ps := clampInt(style.blockSize/2, 3, 8)
		lots = append(lots, lot{
			x: m.cx - ps/2, y: m.cy - ps/2, w: ps, h: ps, kind: lotPlaza,
		})
	}
	if style.hasKeep {
		lots = append(lots, lot{x: m.cx, y: m.cy, w: 2, h: 2, kind: lotTower})
	}
	if style.hasWalls {
		lots = append(lots, wallRing(m, w, h)...)
	}

	return lots
}

// landmarkSize scales a landmark's footprint half-size by its tier so a late-tier
// hero reads bigger. Kept small so it stays a legible marker, not a monolith.
func landmarkSize(tier int) int {
	s := 1 + tier/3
	if s > 3 {
		s = 3
	}
	return s
}

// scatterInBlock returns a random pixel inside a block (inset 0 so it can use the
// full parcel), deterministic from the threaded rng. Used to jitter houses and
// workshops within their parcel so a block reads as a cluster, not a single dot.
func scatterInBlock(b block, r *rng) (int, int) {
	if b.w <= 1 || b.h <= 1 {
		return b.x, b.y
	}
	px := b.x + int(r.f01()*float64(b.w))
	py := b.y + int(r.f01()*float64(b.h))
	return px, py
}

// wallRing emits a ring of lotWall segments around the built-up area as an octagon,
// so the renderer can paint a perimeter wall (a medieval era extra — Phase B uses
// this; Phase A only reaches it via a preset that sets hasWalls, which none do yet,
// but the code path is complete + tested-safe).
func wallRing(m cityMetrics, w, h int) []lot {
	const pts = 8
	out := make([]lot, 0, pts)
	for i := 0; i < pts; i++ {
		ang := 2 * math.Pi * float64(i) / float64(pts)
		x := clampInt(m.cx+int(math.Cos(ang)*float64(m.radius)), 1, w-2)
		y := clampInt(m.cy+int(math.Sin(ang)*float64(m.radius)), 1, h-2)
		out = append(out, lot{x: x, y: y, w: 1, h: 1, kind: lotWall})
	}
	return out
}

// ---- terrain gate -----------------------------------------------------------

// gateLots snaps every lot's origin onto passable land: if the lot's representative
// pixel (its top-left, or center for the larger landmark/plaza lots) sits on water/
// mountain, it is nudged to the nearest passable pixel via nearestPassablePx. This
// is the final guarantee from city-synthesis.md §terrain-gate — NO lot on water.
// Gardens and walls are gated too so a park never floats on a lake. Mutates in place.
func gateLots(lots []lot, f *terrainField) {
	if f == nil || len(f.passable) == 0 {
		return
	}
	for i := range lots {
		// Test the lot's center so a wide landmark/garden is judged by its middle, then
		// translate the whole rect by the same delta the center moved.
		cx := lots[i].x + lots[i].w/2
		cy := lots[i].y + lots[i].h/2
		if f.passableAt(cx, cy) {
			continue
		}
		if nx, ny, ok := nearestPassablePx(f, cx, cy); ok {
			lots[i].x += nx - cx
			lots[i].y += ny - cy
		}
	}
}
