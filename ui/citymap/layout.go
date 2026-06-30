package citymap

import (
	"image/color"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// layout.go is the P2 structure layer: per-age placement strategies plus road
// generation. Each strategy turns the player's actual built buildings — grouped
// into lineage districts — into 2.5D building placements and a set of road
// segments, deterministically seeded so the city is stable frame to frame. The
// palace stays central in every strategy; districts arrange around it per era.
//
// Eras group the 22 ages into the seven silhouettes from the map-overhaul design
// doc. The strategy is chosen by era; the terrain + hue still shift per age
// (terrain.go), so two ages in the same era share a silhouette but not a look.

// era is one of the seven layout silhouettes.
type era int

const (
	eraOrganic    era = iota // primitive, stone — clustered scatter + footpaths
	eraHubSpoke              // bronze, iron, classical — roads radiating from palace
	eraCastle                // medieval, renaissance — keep + wall ring + 4 quarters
	eraZonedGrid             // colonial, industrial, victorian — production/residential split
	eraCityBlocks            // electric, modern, atomic — regular blocks + avenues
	eraCampus                // information, digital, cyberpunk, fusion — cluster pods
	eraOrbital               // space, interstellar, galactic, quantum, transcendent — rings
)

// eraForAge maps an age key to its era band via the canonical age order, so the
// grouping is driven by age position and any future age slots in sensibly. The
// band cut points follow the design doc / P2 spec:
//
//	0-1   organic      (primitive, stone)
//	2-4   hub-and-spoke (bronze, iron, classical)
//	5-6   castle        (medieval, renaissance)
//	7-9   zoned grid    (colonial, industrial, victorian)
//	10-12 city blocks   (electric, atomic, modern)  [orders 10,11,12]
//	13-16 campus        (information, digital, cyberpunk, fusion)
//	17+   orbital       (space, interstellar, galactic, quantum, transcendent)
//
// An unknown age (test stub) defaults to organic — the gentlest strategy.
func eraForAge(ageKey string) era {
	order := config.AgeOrder()
	idx := -1
	for i, k := range order {
		if k == ageKey {
			idx = i
			break
		}
	}
	if idx < 0 {
		return eraOrganic
	}
	switch {
	case idx <= 1:
		return eraOrganic
	case idx <= 4:
		return eraHubSpoke
	case idx <= 6:
		return eraCastle
	case idx <= 9:
		return eraZonedGrid
	case idx <= 12:
		return eraCityBlocks
	case idx <= 16:
		return eraCampus
	default:
		return eraOrbital
	}
}

// importance tiers a placement for sizing and z-order.
type importance int

const (
	impNormal importance = iota
	impWonder
	impPalace
)

// placement is a single 2.5D building volume to draw: a pixel center, a footprint
// half-size (0 → 1px, 1 → 3×3, …), the lineage/category color, and its tier. It
// also carries the building's identity (key/name/lineage/category/tier) so the
// overlay pass can name the marker with the building's own config.BuildingDef.Name
// — one named marker per built building type — rather than a lineage banner. The
// palace placement leaves the identity fields zero (it is labeled "City Center").
type placement struct {
	cx, cy int
	size   int
	col    color.RGBA
	tier   importance

	// Building identity (empty for the palace). The label text is name; the label
	// color is derived from lineageKey/category via lineageColor at draw time so it
	// retints with the theme. ltier (LineageTier) + size prioritize labels when space
	// is tight (higher tier / bigger volume gets its name first per cluster).
	key        string
	name       string
	lineageKey string
	category   string
	ltier      int
}

// roadSeg is a straight road drawn between two pixel endpoints (Bresenham).
type roadSeg struct {
	x0, y0, x1, y1 int
}

// buildingItem is one distinct BUILT building type (Count > 0) the map draws as
// its own marker. It carries everything a placement needs to render + be named:
// the config key, the human Name, the lineage/category (for color + clustering),
// the instance count (nudges volume size), and the LineageTier (label priority).
type buildingItem struct {
	key      string
	name     string
	category string
	count    int
	tier     int // LineageTier — higher = more advanced; used to rank labels
}

// district is one lineage's (or special category's) buildings, grouped together so
// they cluster + share a color on the map. Each member of buildings draws as its
// OWN marker (one per distinct built type) within the district's region; the
// district itself contributes the shared lineage color (col) and the placement
// region the era strategy assigns it. lineageKey/category identify the group; col
// is the lineage color every member volume uses.
type district struct {
	lineageKey string
	category   string
	col        color.RGBA
	buildings  []buildingItem
}

// buildingMeta resolves a building key to its lineage key and category via the
// canonical config table. Pure data, no locks (config.BuildingByKey is a plain
// map build), so it is safe to call from the render path. Cached per call site by
// the caller; here it is a thin lookup. Unknown keys return ("","",false).
func buildingMeta(byKey map[string]config.BuildingDef, key string) (lineageKey, category string, ok bool) {
	d, ok := byKey[key]
	if !ok {
		return "", "", false
	}
	return d.LineageKey, d.Category, true
}

// builtDistricts groups the player's built buildings by lineage into districts,
// sorted for determinism. Crucially this no longer collapses a lineage to a few
// representatives: EVERY distinct built building type (Count > 0) becomes its own
// buildingItem inside its lineage's district, so the map draws one named marker per
// type. Same-lineage buildings land in the same district (so they cluster + share
// a color); storage/wonder/diplomacy/monument group by category so they read as a
// distinct neighborhood. keys must be sorted; counts maps key→instance count;
// byKey supplies each building's Name + LineageTier (pure data, no locks).
func builtDistricts(byKey map[string]config.BuildingDef, keys []string, counts map[string]int) []district {
	// Aggregate by a grouping key: production lineages group by lineageKey; the
	// special categories group by category so e.g. all wonders cluster together.
	groups := map[string]*district{}
	order := make([]string, 0) // preserve first-seen order before the stable sort
	for _, k := range keys {
		lineage, category, _ := buildingMeta(byKey, k)
		gkey := lineage
		switch category {
		case "wonder", "monument", "storage", "diplomacy":
			gkey = "cat:" + category
		}
		if gkey == "" {
			gkey = "misc"
		}
		d := groups[gkey]
		if d == nil {
			d = &district{
				lineageKey: lineage,
				category:   category,
				col:        lineageColor(lineage, category),
			}
			groups[gkey] = d
			order = append(order, gkey)
		}
		def := byKey[k]
		name := def.Name
		if name == "" {
			name = titleCaseKey(k) // unknown/test key: a sensible label rather than blank
		}
		d.buildings = append(d.buildings, buildingItem{
			key:      k,
			name:     name,
			category: category,
			count:    counts[k],
			tier:     def.LineageTier,
		})
	}
	sort.Strings(order)

	out := make([]district, 0, len(order))
	for _, g := range order {
		out = append(out, *groups[g])
	}
	return out
}

// buildingVolumeSize scales a single building's 2.5D volume by its own instance
// count so a building you've stamped many times reads bulkier than a lone one —
// exactly one marker per type either way, the count only nudges the footprint.
// Subtle by design: 0 → 1px up to 2 → 5×5.
func buildingVolumeSize(count int) int {
	switch {
	case count <= 2:
		return 0
	case count <= 10:
		return 1
	default:
		return 2
	}
}

// layoutSeed derives a deterministic seed from the age plus the set of built
// building keys, so the layout is stable while the same buildings are present but
// reshuffles sensibly when the empire changes. FNV-1a over age + sorted keys.
func layoutSeed(ageKey string, keys []string) uint32 {
	var h uint32 = 2166136261
	mix := func(s string) {
		for i := 0; i < len(s); i++ {
			h ^= uint32(s[i])
			h *= 16777619
		}
		h ^= '|'
		h *= 16777619
	}
	mix(ageKey)
	for _, k := range keys {
		mix(k)
	}
	return h | 1
}

// rng is a tiny deterministic PRNG (xorshift32) for jittering placements. Pure
// and seeded, so a given (age, building set) always produces the same city.
type rng struct{ s uint32 }

func newRNG(seed uint32) *rng {
	if seed == 0 {
		seed = 1
	}
	return &rng{s: seed}
}

func (r *rng) next() uint32 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 17
	r.s ^= r.s << 5
	return r.s
}

// f01 returns a float in [0,1).
func (r *rng) f01() float64 { return float64(r.next()) / float64(^uint32(0)) }

// span returns an int in [-half, half].
func (r *rng) span(half int) int {
	if half <= 0 {
		return 0
	}
	return int(r.next()%uint32(2*half+1)) - half
}

// layoutResult is what every strategy returns: the building volumes (palace
// first) and the road network.
type layoutResult struct {
	placements []placement
	roads      []roadSeg
}

// buildLayout dispatches to the era strategy and returns placements + roads.
// districts must already be lineage-grouped; palaceCol/wonderCol come from the
// terrain palette (so the palace tracks the theme accent). w,h are pixels.
func buildLayout(e era, w, h int, districts []district, pal terrainPalette, seed uint32) layoutResult {
	r := newRNG(seed)
	switch e {
	case eraOrganic:
		return layoutOrganic(w, h, districts, pal, r)
	case eraHubSpoke:
		return layoutHubSpoke(w, h, districts, pal, r)
	case eraCastle:
		return layoutCastle(w, h, districts, pal, r)
	case eraZonedGrid:
		return layoutZonedGrid(w, h, districts, pal, r)
	case eraCityBlocks:
		return layoutCityBlocks(w, h, districts, pal, r)
	case eraCampus:
		return layoutCampus(w, h, districts, pal, r)
	case eraOrbital:
		return layoutOrbital(w, h, districts, pal, r)
	default:
		return layoutOrganic(w, h, districts, pal, r)
	}
}

// palacePlacement builds the central palace volume. Size scales gently with the
// canvas so it dominates without swallowing the map.
func palacePlacement(w, h int, pal terrainPalette) placement {
	size := 2
	if w < 40 || h < 30 {
		size = 1
	}
	return placement{cx: w / 2, cy: h / 2, size: size, col: pal.palace, tier: impPalace}
}

// buildingPlacement builds the placement for one building of a district at a given
// pixel, copying the lineage color + the building's identity (so the overlay can
// name the marker) and sizing the volume by the building's own instance count.
// Wonders/monuments tier up so they render larger and on top.
func buildingPlacement(d district, bi buildingItem, px, py int) placement {
	tier := impNormal
	if bi.category == "wonder" || bi.category == "monument" {
		// Showpiece sizing is applied uniformly in drawVolume by tier, so only the
		// tier is set here — keeps wonder size consistent across every strategy.
		tier = impWonder
	}
	return placement{
		cx: px, cy: py, size: buildingVolumeSize(bi.count), col: d.col, tier: tier,
		key: bi.key, name: bi.name, lineageKey: d.lineageKey, category: bi.category, ltier: bi.tier,
	}
}

// scatterDistrict drops one volume per building in the district around an anchor
// with a little jitter, each sized by that building's own count and clipped inside
// a margin, so a lineage reads as a small neighborhood of its individual buildings.
func scatterDistrict(d district, ax, ay, spread, w, h int, r *rng) []placement {
	out := make([]placement, 0, len(d.buildings))
	for _, bi := range d.buildings {
		px := clampInt(ax+r.span(spread), 2, w-3)
		py := clampInt(ay+r.span(spread), 2, h-3)
		out = append(out, buildingPlacement(d, bi, px, py))
	}
	return out
}

// ---- Strategy: organic scatter (primitive, stone) -------------------------
//
// Clustered random-walk: districts wander out from the center on a drunk path,
// each lineage its own loose cluster. Roads are faint footpaths back to the
// center — short, kinked, not a grid.
func layoutOrganic(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	cx, cy := w/2, h/2
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	maxR := float64(minInt(w, h))/2 - 4
	if maxR < 5 {
		maxR = 5
	}
	for _, d := range districts {
		// Random-walk the cluster anchor outward a bit each district.
		ang := r.f01() * 2 * math.Pi
		step := maxR * (0.35 + 0.5*r.f01())
		ax := clampInt(cx+int(math.Cos(ang)*step), 3, w-4)
		ay := clampInt(cy+int(math.Sin(ang)*step), 3, h-4)
		res.placements = append(res.placements, scatterDistrict(d, ax, ay, 3, w, h, r)...)

		// Faint footpath: a kinked two-segment path from the cluster toward center.
		midX := (ax+cx)/2 + r.span(2)
		midY := (ay+cy)/2 + r.span(2)
		res.roads = append(res.roads,
			roadSeg{ax, ay, midX, midY},
			roadSeg{midX, midY, cx, cy},
		)
	}
	if len(districts) == 0 {
		// Even an empty village gets a stub path so the road layer is never blank.
		res.roads = append(res.roads, roadSeg{cx, cy, clampInt(cx+6, 0, w-1), cy})
	}
	return res
}

// ---- Strategy: hub-and-spoke (bronze, iron, classical) --------------------
//
// Spokes radiate from the central palace; each lineage district sits along one
// spoke, buildings strung down the road. Roads ARE the spokes — legible radials.
func layoutHubSpoke(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	cx, cy := w/2, h/2
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	maxR := float64(minInt(w, h))/2 - 4
	if maxR < 5 {
		maxR = 5
	}
	n := len(districts)
	if n == 0 {
		// A lone spoke so the hub still reads.
		res.roads = append(res.roads, roadSeg{cx, cy, clampInt(cx+int(maxR), 0, w-1), cy})
		return res
	}
	for i, d := range districts {
		ang := 2 * math.Pi * float64(i) / float64(n)
		ex := clampInt(cx+int(math.Cos(ang)*maxR), 2, w-3)
		ey := clampInt(cy+int(math.Sin(ang)*maxR), 2, h-3)
		res.roads = append(res.roads, roadSeg{cx, cy, ex, ey})
		// String the district's buildings down the spoke at increasing radius — one
		// marker per building, packed along the road so the lineage reads as a street.
		nb := len(d.buildings)
		for j, bi := range d.buildings {
			t := 0.4 + 0.55*float64(j)/math.Max(1, float64(nb))
			px := clampInt(cx+int(math.Cos(ang)*maxR*t)+r.span(1), 2, w-3)
			py := clampInt(cy+int(math.Sin(ang)*maxR*t)+r.span(1), 2, h-3)
			res.placements = append(res.placements, buildingPlacement(d, bi, px, py))
		}
	}
	return res
}

// ---- Strategy: castle + quarters (medieval, renaissance) ------------------
//
// A central keep, a wall ring drawn as a closed road loop, four quarters (one per
// compass diagonal) each holding a slice of districts, and roads from the keep to
// each quarter.
func layoutCastle(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	cx, cy := w/2, h/2
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	ringR := minInt(w, h)/2 - 4
	if ringR < 6 {
		ringR = 6
	}
	// Wall ring as an octagon of road segments (a clean closed loop on the canvas).
	const ringPts = 8
	var prev [2]int
	for i := 0; i <= ringPts; i++ {
		ang := 2 * math.Pi * float64(i) / float64(ringPts)
		x := clampInt(cx+int(math.Cos(ang)*float64(ringR)), 1, w-2)
		y := clampInt(cy+int(math.Sin(ang)*float64(ringR)), 1, h-2)
		if i > 0 {
			res.roads = append(res.roads, roadSeg{prev[0], prev[1], x, y})
		}
		prev = [2]int{x, y}
	}

	// Four quarter anchors on the diagonals, roads from keep to each.
	quarterAng := []float64{math.Pi / 4, 3 * math.Pi / 4, 5 * math.Pi / 4, 7 * math.Pi / 4}
	qr := float64(ringR) * 0.55
	anchors := make([][2]int, 4)
	for i, a := range quarterAng {
		ax := clampInt(cx+int(math.Cos(a)*qr), 3, w-4)
		ay := clampInt(cy+int(math.Sin(a)*qr), 3, h-4)
		anchors[i] = [2]int{ax, ay}
		res.roads = append(res.roads, roadSeg{cx, cy, ax, ay})
	}
	// Distribute districts round-robin into the four quarters.
	for i, d := range districts {
		a := anchors[i%4]
		res.placements = append(res.placements, scatterDistrict(d, a[0], a[1], 3, w, h, r)...)
	}
	return res
}

// ---- Strategy: zoned grid (colonial, industrial, victorian) ---------------
//
// The map splits into a production zone (left) and a residential zone (right)
// with a road grid over each. Production-ish lineages go left, the rest right.
func layoutZonedGrid(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	cx := w / 2
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	// A road grid: a few verticals + horizontals, plus a strong central spine
	// dividing the two zones.
	margin := 3
	res.roads = append(res.roads, roadSeg{cx, margin, cx, h - margin - 1}) // central divider
	for _, fx := range []float64{0.22, 0.5, 0.78} {
		x := clampInt(int(float64(w)*fx), margin, w-margin-1)
		res.roads = append(res.roads, roadSeg{x, margin, x, h - margin - 1})
	}
	for _, fy := range []float64{0.3, 0.6, 0.85} {
		y := clampInt(int(float64(h)*fy), margin, h-margin-1)
		res.roads = append(res.roads, roadSeg{margin, y, w - margin - 1, y})
	}

	// Lineages considered "production" sit in the left zone; everything else right.
	isProd := map[string]bool{
		"food": true, "organic_extraction": true, "geological_extraction": true,
		"metallurgy": true, "energy": true, "engineering": true, "harbor": true,
		"hacker": true, "astronaut": true,
	}
	leftN, rightN := 0, 0
	for _, d := range districts {
		left := isProd[d.lineageKey] || d.category == "storage"
		var ax, ay int
		if left {
			ax = clampInt(w/4+r.span(w/10), margin+1, cx-2)
			ay = clampInt(margin+2+(leftN%3)*(h/4)+r.span(2), margin+1, h-margin-2)
			leftN++
		} else {
			ax = clampInt(3*w/4+r.span(w/10), cx+2, w-margin-1)
			ay = clampInt(margin+2+(rightN%3)*(h/4)+r.span(2), margin+1, h-margin-2)
			rightN++
		}
		res.placements = append(res.placements, scatterDistrict(d, ax, ay, 2, w, h, r)...)
	}
	return res
}

// ---- Strategy: city blocks (electric, atomic, modern) ---------------------
//
// Regular blocks separated by avenues: a full grid of roads, districts dropped
// into block centers in reading order.
func layoutCityBlocks(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	margin := 2
	cols, rows := 4, 3
	cellW := (w - 2*margin) / cols
	cellH := (h - 2*margin) / rows
	if cellW < 2 {
		cellW = 2
	}
	if cellH < 2 {
		cellH = 2
	}
	// Avenues: vertical + horizontal lines bounding every block.
	for c := 0; c <= cols; c++ {
		x := clampInt(margin+c*cellW, 0, w-1)
		res.roads = append(res.roads, roadSeg{x, margin, x, margin + rows*cellH})
	}
	for rw := 0; rw <= rows; rw++ {
		y := clampInt(margin+rw*cellH, 0, h-1)
		res.roads = append(res.roads, roadSeg{margin, y, margin + cols*cellW, y})
	}
	// Drop districts into block centers, skipping the central block (the palace).
	centerCol, centerRow := cols/2, rows/2
	bi := 0
	place := func(d district) {
		for {
			c := bi % cols
			rw := (bi / cols) % rows
			bi++
			if c == centerCol && rw == centerRow {
				continue // reserve the middle block for the palace
			}
			ax := clampInt(margin+c*cellW+cellW/2+r.span(1), 2, w-3)
			ay := clampInt(margin+rw*cellH+cellH/2+r.span(1), 2, h-3)
			res.placements = append(res.placements, scatterDistrict(d, ax, ay, 1, w, h, r)...)
			return
		}
	}
	for _, d := range districts {
		place(d)
	}
	return res
}

// ---- Strategy: campus clusters (information, digital, cyberpunk, fusion) ---
//
// Cluster pods arranged on a hex ring around a central hub, connected by paths
// (pod → hub). Each district is one pod; volumes pack tightly inside the pod.
func layoutCampus(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	cx, cy := w/2, h/2
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	maxR := float64(minInt(w, h))/2 - 4
	if maxR < 5 {
		maxR = 5
	}
	n := len(districts)
	if n == 0 {
		res.roads = append(res.roads, roadSeg{cx, cy, clampInt(cx+int(maxR), 0, w-1), cy})
		return res
	}
	// Pods evenly spaced on a ring around the hub, with a hex-ish angular stagger.
	for i, d := range districts {
		ang := 2*math.Pi*float64(i)/float64(n) + 0.3
		podX := clampInt(cx+int(math.Cos(ang)*maxR), 4, w-5)
		podY := clampInt(cy+int(math.Sin(ang)*maxR), 4, h-5)
		res.roads = append(res.roads, roadSeg{cx, cy, podX, podY})
		// Pack the pod: tight hex-ish offsets around the pod center, one slot per
		// building. If a lineage has more buildings than base slots, stack further
		// rows (every building still gets its own marker — never capped).
		offsets := [][2]int{{0, 0}, {2, 0}, {-2, 0}, {1, 2}, {-1, 2}}
		for j, bi := range d.buildings {
			off := offsets[j%len(offsets)]
			extraRow := (j / len(offsets)) * 2 // push overflow buildings down a band
			px := clampInt(podX+off[0], 2, w-3)
			py := clampInt(podY+off[1]+extraRow, 2, h-3)
			res.placements = append(res.placements, buildingPlacement(d, bi, px, py))
		}
	}
	return res
}

// ---- Strategy: orbital rings (space, interstellar, galactic, quantum) ------
//
// Concentric rings around a central hub, with radial connectors. Each district
// occupies an arc of a ring; rings fill from the inside out as districts grow.
func layoutOrbital(w, h int, districts []district, pal terrainPalette, r *rng) layoutResult {
	res := layoutResult{}
	cx, cy := w/2, h/2
	res.placements = append(res.placements, palacePlacement(w, h, pal))

	maxR := float64(minInt(w, h))/2 - 3
	if maxR < 6 {
		maxR = 6
	}
	n := len(districts)
	// Up to 3 rings; distribute districts across them.
	rings := 1
	if n > 3 {
		rings = 2
	}
	if n > 7 {
		rings = 3
	}
	ringRadius := func(ri int) float64 {
		// Evenly spaced rings between ~40% and 100% of maxR.
		if rings == 1 {
			return maxR * 0.7
		}
		return maxR * (0.4 + 0.6*float64(ri)/float64(rings-1))
	}
	// Draw the ring circles as many-segment loops.
	for ri := 0; ri < rings; ri++ {
		rad := ringRadius(ri)
		const seg = 16
		var prev [2]int
		for i := 0; i <= seg; i++ {
			ang := 2 * math.Pi * float64(i) / float64(seg)
			x := clampInt(cx+int(math.Cos(ang)*rad), 1, w-2)
			y := clampInt(cy+int(math.Sin(ang)*rad), 1, h-2)
			if i > 0 {
				res.roads = append(res.roads, roadSeg{prev[0], prev[1], x, y})
			}
			prev = [2]int{x, y}
		}
	}
	// Radial connectors from hub to the outer ring at the cardinal+diagonal angles.
	outer := ringRadius(rings - 1)
	for k := 0; k < 8; k++ {
		ang := 2 * math.Pi * float64(k) / 8
		ex := clampInt(cx+int(math.Cos(ang)*outer), 1, w-2)
		ey := clampInt(cy+int(math.Sin(ang)*outer), 1, h-2)
		res.roads = append(res.roads, roadSeg{cx, cy, ex, ey})
	}
	// Place each district along its ring arc.
	for i, d := range districts {
		ri := i % rings
		rad := ringRadius(ri)
		baseAng := 2 * math.Pi * float64(i) / math.Max(1, float64(n))
		// One marker per building, fanned along the ring arc so the lineage occupies
		// a contiguous slice of its ring.
		for j, bi := range d.buildings {
			ang := baseAng + float64(j)*0.18
			px := clampInt(cx+int(math.Cos(ang)*rad)+r.span(1), 2, w-3)
			py := clampInt(cy+int(math.Sin(ang)*rad)+r.span(1), 2, h-3)
			res.placements = append(res.placements, buildingPlacement(d, bi, px, py))
		}
	}
	return res
}

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
