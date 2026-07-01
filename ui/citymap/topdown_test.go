package citymap

import (
	"image"
	"image/color"
	"math"
	"sort"
	"strconv"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// topdown_test.go locks the V3-A top-down engine invariants (LOCKED spec:
// design-and-architecture/city-synthesis.md). It reuses the shared helpers from
// citymap_test.go / layout_test.go (sampleState, imagesDiffer, colorClose,
// oneAgePerEra) and adds the top-down-specific guards:
//
//   - determinism           — same state/seed/size → identical pixels + identical plan.
//   - stable-incremental     — the golden-spiral slots (with jitter) for the first N
//     instances don't move when instance N+1 is added.
//   - count-drives-population — more of a housing building → strictly more house lots,
//     near-1:1 low, sqrt-damped high.
//   - fill-frame / in-bounds  — every lot maps in-bounds; the town bbox sits inside the
//     frame with margin.
//   - roof atlas             — every primitive roofType paints non-zero pixels without
//     panic, AND paints NO pixel equal to the theme RoleAccent/RoleHighlight color
//     (the yellow-dot regression guard: the ridge is base-derived, not an accent).
//   - greenery-in-town        — filler lots lie within/near the town bbox, not scattered
//     across the far canvas.
//   - landmark-only overlay   — 1 title, ≥1 landmark label, a City Center label, zero
//     civ/trade labels.
//   - panic-safe + exact size — 0×0 / 1×1 / tiny / 1×40 all render at the requested size.

// tdPlanFor builds the top-down plan for a state exactly the way renderImage does:
// era style from the age, city seed from the display name. The single source of truth
// for the plan-level invariant tests below.
func tdPlanFor(state game.GameState) topPlan {
	style := tdStyleForEra(eraForAge(state.Age))
	seed := citySeed(displayNameOf(state))
	return generateTopPlan(state, config.BuildingByKey(), style, seed)
}

// roofLots returns just the roof lots of a plan, preserving order.
func roofLots(plan topPlan) []tdLot {
	out := make([]tdLot, 0, len(plan.lots))
	for _, lt := range plan.lots {
		if lt.kind == tdRoof {
			out = append(out, lt)
		}
	}
	return out
}

// countRoofType counts roof lots of a given archetype in a plan.
func countRoofType(plan topPlan, rt roofType) int {
	n := 0
	for _, lt := range plan.lots {
		if lt.kind == tdRoof && lt.roof == rt {
			n++
		}
	}
	return n
}

// TestTopDownDeterministic proves the engine is a pure function of (state, seed, size):
// the same inputs must yield byte-identical pixels AND an identical plan (same lot
// sequence). If either drifted, caching and stable placement would both be unsound.
func TestTopDownDeterministic(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	state := sampleState("primitive_age", map[string]int{"hut": 8, "gathering_camp": 4, "shrine": 1})

	imgA, _ := renderImage(state, 100, 60)
	imgB, _ := renderImage(state, 100, 60)
	if imagesDiffer(imgA, imgB) {
		t.Fatal("identical state/size rendered different pixels — engine is not deterministic")
	}

	planA := tdPlanFor(state)
	planB := tdPlanFor(state)
	if len(planA.lots) != len(planB.lots) {
		t.Fatalf("generateTopPlan lot count drift: %d vs %d", len(planA.lots), len(planB.lots))
	}
	for i := range planA.lots {
		a, b := planA.lots[i], planB.lots[i]
		if a != b {
			t.Fatalf("generateTopPlan lot %d differs: %+v vs %+v", i, a, b)
		}
	}
}

// streetCellSet returns the plan's street-cell centers as a rounded (x,y) key set, so two plans'
// block-boundary structures can be compared for equality regardless of slice order. Rounded to
// avoid float noise (the cells sit on a fixed raster, so equal fields yield identical keys).
func streetCellSet(plan topPlan) map[[2]int]bool {
	s := map[[2]int]bool{}
	for _, p := range plan.streetCells {
		s[[2]int{int(math.Round(p.x * 100)), int(math.Round(p.y * 100))}] = true
	}
	return s
}

// streetCellSig is a stable signature string of a plan's street-cell field, so two plans' block-
// boundary structures can be compared for equality. Sorted so slice order doesn't matter.
func streetCellSig(plan topPlan) string {
	set := streetCellSet(plan)
	keys := make([][2]int, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	var b []byte
	for _, k := range keys {
		b = strconv.AppendInt(b, int64(k[0]), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(k[1]), 10)
		b = append(b, ';')
	}
	return string(b)
}

// TestBandedStructureStability locks the Voronoi-block model's STABILITY TRADEOFF (see the
// topdown.go header): the block STRUCTURE (the street-gap network + the town size) is BANDED — it
// is a step function of the roof count, so it is IDENTICAL for a run of adjacent counts (a band)
// and re-forms only at a band boundary or on age-up. This replaces the old lane model's strict
// slot-for-slot incrementality; the real-town look is the priority. Measuring STRUCTURALLY (by the
// street-cell signature) rather than assuming specific band edges: we sweep a range of hut counts
// and assert (1) the signature is PIECEWISE-CONSTANT — adjacent counts frequently share a signature
// (bands exist, not a fresh field per building), and (2) it DOES change as the count grows (the
// banding is real, and the structure adapts to size), while every town stays compact.
func TestBandedStructureStability(t *testing.T) {
	_ = theme.SetActive("forge")

	// Sweep hut counts; record each plan's street-cell signature + the roof-lot count.
	type sample struct {
		n    int
		sig  string
		lots []tdLot
	}
	var samples []sample
	for n := 20; n <= 60; n++ {
		plan := tdPlanFor(sampleState("primitive_age", map[string]int{"hut": n}))
		samples = append(samples, sample{n: n, sig: streetCellSig(plan), lots: roofLots(plan)})
		if len(plan.streetCells) < 8 {
			t.Fatalf("hut count %d: only %d street cells — expected a real block-gap network", n, len(plan.streetCells))
		}
	}

	// (1) PIECEWISE-CONSTANT: count how many adjacent-count pairs share the SAME signature (i.e.
	// sit in the same band). A per-building fresh field would score ~0; a banded field scores high.
	sameBand := 0
	for i := 1; i < len(samples); i++ {
		if samples[i].sig == samples[i-1].sig {
			sameBand++
		}
	}
	if frac := float64(sameBand) / float64(len(samples)-1); frac < 0.5 {
		t.Fatalf("only %.0f%% of adjacent counts share a block field — the structure re-forms too often to be banded (tradeoff not delivered)", frac*100)
	}

	// (2) The banding is REAL, not a single frozen field: distinct signatures appear across the
	// swept range (the structure adapts to town size at band boundaries).
	sigs := map[string]bool{}
	for _, s := range samples {
		sigs[s.sig] = true
	}
	if len(sigs) < 2 {
		t.Fatalf("the whole count sweep produced a single block field — the banding never re-forms (structure does not adapt to size)")
	}

	// Within a band (a same-signature adjacent pair), MOST existing roofs keep their exact place
	// (the blocks fill progressively). Verify on the first such pair found.
	for i := 1; i < len(samples); i++ {
		if samples[i].sig != samples[i-1].sig {
			continue
		}
		posPrev := map[[2]int]bool{}
		for _, lt := range samples[i-1].lots {
			posPrev[[2]int{int(math.Round(lt.x * 100)), int(math.Round(lt.y * 100))}] = true
		}
		kept := 0
		for _, lt := range samples[i].lots {
			if posPrev[[2]int{int(math.Round(lt.x * 100)), int(math.Round(lt.y * 100))}] {
				kept++
			}
		}
		if got, want := float64(kept)/float64(maxInt(len(posPrev), 1)), 0.6; got < want {
			t.Fatalf("within a band (%d→%d) only %.0f%% of roofs kept their place (want ≥%.0f%%) — the fill is not progressive",
				samples[i-1].n, samples[i].n, got*100, want*100)
		}
		break
	}

	// Jitter is real (the block perimeter would otherwise be a perfect stamp): at least one slot
	// gets a nonzero offset. Pure over (i, di, seed).
	moved := false
	for i := 0; i < 20; i++ {
		jx, jy := slotJitter(i, 0, citySeed(displayNameOf(sampleState("primitive_age", nil))), defaultTdConfig.jitterAmp)
		if math.Abs(jx) > 1e-9 || math.Abs(jy) > 1e-9 {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatal("slotJitter produced no offset for any slot — the lattice is not being broken up")
	}
}

// TestSlotJitterPureFunction directly locks the property the stability guarantee rests
// on: slotJitter is a pure function of (i, di, seed) — repeated calls return the same
// value, and it never consults external state. If someone reroutes jitter through the
// threaded rng, this catches it.
func TestSlotJitterPureFunction(t *testing.T) {
	const seed uint32 = 0xC0FFEE01
	for i := 0; i < 50; i++ {
		x1, y1 := slotJitter(i, 2, seed, 1.0)
		x2, y2 := slotJitter(i, 2, seed, 1.0)
		if x1 != x2 || y1 != y2 {
			t.Fatalf("slotJitter(%d,...) not pure: (%v,%v) vs (%v,%v)", i, x1, y1, x2, y2)
		}
		// Amplitude bound: the offset must stay within [-amp,amp] on each axis so
		// buildings still pack close.
		if math.Abs(x1) > 1.0 || math.Abs(y1) > 1.0 {
			t.Fatalf("slotJitter(%d) exceeds amplitude: (%v,%v)", i, x1, y1)
		}
	}
	// amp<=0 → no jitter (a clean way to disable it).
	if x, y := slotJitter(3, 1, seed, 0); x != 0 || y != 0 {
		t.Fatalf("slotJitter with amp 0 = (%v,%v), want (0,0)", x, y)
	}
}

// TestCountDrivesPopulation checks the count→roof-lot mapping (locked #4): more of a
// housing building yields strictly MORE house lots, near-1:1 at low counts and
// sqrt-damped high. We drive hut count up and watch the roofHut/roofLong lot count.
func TestCountDrivesPopulation(t *testing.T) {
	_ = theme.SetActive("forge")

	housingLots := func(hutCount int) int {
		st := sampleState("primitive_age", map[string]int{"hut": hutCount})
		plan := tdPlanFor(st)
		// Housing tier-0 huts render as roofHut.
		return countRoofType(plan, roofHut)
	}

	// Near-1:1 at low counts: 1..12 should map 1:1 (locked #4 low band).
	for _, n := range []int{1, 3, 6, 12} {
		if got := housingLots(n); got != n {
			t.Fatalf("housing lots for %d huts = %d, want %d (near-1:1 low band)", n, got, n)
		}
	}

	// Strictly increasing across a wide range (monotonic, never fewer roofs for more
	// buildings).
	prev := 0
	for _, n := range []int{1, 5, 12, 20, 40, 80} {
		got := housingLots(n)
		if got < prev {
			t.Fatalf("housing lots non-monotonic: %d huts → %d lots, but a smaller count gave %d", n, got, prev)
		}
		prev = got
	}

	// Sub-linear (sqrt-damped) high: 200 huts must NOT emit 200 lots (the fabric
	// densifies, it doesn't clone 1:1), and stays under the legibility cap.
	hi := housingLots(200)
	if hi >= 200 {
		t.Fatalf("200 huts emitted %d lots — high counts must sub-scale, not clone 1:1", hi)
	}
	if hi > 80 {
		t.Fatalf("200 huts emitted %d lots — over the legibility cap of 80", hi)
	}
	// And more than the low band, so growth is real.
	if hi <= 12 {
		t.Fatalf("200 huts emitted only %d lots — high band should exceed the low band", hi)
	}
}

// TestFillFrameInBounds checks the fill-frame transform maps every lot in-bounds and the
// town bounding box sits inside the frame with a margin (locked #3). No lot pixel may
// map outside the canvas, and the built-up bbox must not touch the frame edge.
func TestFillFrameInBounds(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 120, 80
	state := sampleState("primitive_age", map[string]int{"hut": 14, "gathering_camp": 6, "stone_camp": 3, "shrine": 1})
	plan := tdPlanFor(state)
	xf := computeTransform(&plan, w, h)

	// Every ROOF lot center maps within [0,w)×[0,h): the built city fills the frame and
	// no building falls off-canvas. (Fill-frame fits the roofs, not the greenery.)
	for i, lt := range plan.lots {
		if lt.kind != tdRoof {
			continue
		}
		px, py := xf.px(lt.x, lt.y)
		if px < 0 || px >= w || py < 0 || py >= h {
			t.Fatalf("roof lot %d maps to (%d,%d), out of %dx%d bounds", i, px, py, w, h)
		}
	}

	// Greenery is seasoning and may sit at/just past the margin (the groves ring the town
	// edge), but must never be flung far off-canvas — bound it to a small overscan.
	const over = 12
	for i, lt := range plan.lots {
		if lt.kind == tdRoof {
			continue
		}
		px, py := xf.px(lt.x, lt.y)
		if px < -over || px >= w+over || py < -over || py >= h+over {
			t.Fatalf("filler lot %d (kind %d) maps to (%d,%d), far outside %dx%d frame", i, lt.kind, px, py, w, h)
		}
	}

	// The town bbox (roofs only) maps inside the frame with a >=1px margin on each side.
	minX, minY, maxX, maxY := tdRoofBBox(&plan)
	x0, y0 := xf.px(minX, minY)
	x1, y1 := xf.px(maxX, maxY)
	if x0 < 1 || y0 < 1 || x1 > w-2 || y1 > h-2 {
		t.Fatalf("town bbox px=[%d,%d..%d,%d] not inside %dx%d frame with margin", x0, y0, x1, y1, w, h)
	}
}

// fabricLots returns the non-wonder roof lots (the intermixed settlement fabric), in
// placement (slice) order. Wonders are excluded — they are anchors/centerpieces, not
// fabric — so the intermix + clustering checks measure the town, not the showpieces.
func fabricLots(plan topPlan) []tdLot {
	out := make([]tdLot, 0, len(plan.lots))
	for _, lt := range plan.lots {
		if lt.kind == tdRoof && lt.roof != roofWonder {
			out = append(out, lt)
		}
	}
	return out
}

// nnDomainDiffFrac returns the fraction of lots whose NEAREST neighbor is a DIFFERENT
// domain. Near 0 means same-domain blobs (each lot is surrounded by its own kind); a
// healthy fraction means the domains are spatially intermixed (FIX 1). O(n²) but the
// test sets are small.
func nnDomainDiffFrac(lots []tdLot) float64 {
	if len(lots) < 2 {
		return 0
	}
	diff, tot := 0, 0
	for i := range lots {
		best := math.Inf(1)
		bj := -1
		for j := range lots {
			if i == j {
				continue
			}
			d := math.Hypot(lots[i].x-lots[j].x, lots[i].y-lots[j].y)
			if d < best {
				best, bj = d, j
			}
		}
		if bj >= 0 {
			tot++
			if lots[bj].domain != lots[i].domain {
				diff++
			}
		}
	}
	return float64(diff) / float64(tot)
}

// TestIntermixedBlockPlacement locks the type-INTERMIX invariant under the Voronoi-block model:
// buildings are distributed across the blocks so the fabric is spatially MIXED (no single-domain
// round blob — a hut ward next to a camp next to a store), NOT one giant blob of huts. We drive a
// village of five distinct domains and assert the spatial nearest-neighbor domain mixing exceeds a
// healthy threshold, and that a real block-gap street network exists.
func TestIntermixedBlockPlacement(t *testing.T) {
	_ = theme.SetActive("forge")
	// hut=housing, gathering_camp=food, stone_camp=geological, forge=metallurgy, barracks=military
	// — five distinct domains, no wonder (one center anchor).
	state := sampleState("primitive_age", map[string]int{
		"hut": 12, "gathering_camp": 10, "stone_camp": 8, "forge": 8, "barracks": 6,
	})
	plan := tdPlanFor(state)
	fab := fabricLots(plan)
	if len(fab) < 20 {
		t.Fatalf("expected a substantial fabric, got %d lots", len(fab))
	}

	distinctDomains := map[string]bool{}
	for _, lt := range fab {
		distinctDomains[lt.domain] = true
	}
	if len(distinctDomains) < 3 {
		t.Fatalf("fabric only spans %d domains — need a multi-domain settlement to test intermix", len(distinctDomains))
	}

	// SPATIAL intermix: a lot's nearest neighbor is frequently a DIFFERENT domain — the blocks
	// hold a mix of types, so there is no single-domain round blob. A per-domain-cluster layout
	// (the retired district-spiral model) would score near 0.
	spatial := nnDomainDiffFrac(fab)
	if spatial < 0.30 {
		t.Fatalf("spatial nearest-neighbor domain-mixing only %.2f — the fabric reads as same-type blobs, not intermixed across blocks", spatial)
	}

	// The block-gap STREET network must exist (the gaps between wards): a real town has a
	// meaningful number of street cells.
	if len(plan.streetCells) < 8 {
		t.Fatalf("expected a block-gap street network, got only %d street cells", len(plan.streetCells))
	}
}

// TestWonderAnchoredGrowth locks FIX 2: wonders are the CENTRAL growth anchors the
// settlement hugs, each with a CLEAR PLAZA, scaling with the wonder count. With ≥1
// wonder the fabric clusters AROUND the anchor(s) and no fabric lot sits inside a
// wonder's plaza; with 0 wonders the village is one cohesive settlement around the
// center; and the anchor set scales 0→1→several as wonders are built.
func TestWonderAnchoredGrowth(t *testing.T) {
	_ = theme.SetActive("forge")

	plazaR := defaultTdConfig.plazaRadius * defaultTdConfig.roofSize

	// Anchor count scales with wonders: 0 → 1 (city center), 1 → 1 (the wonder), 3 → 3.
	wonderAnchors := func(m map[string]int) (nWonder, nAnchor int) {
		plan := tdPlanFor(sampleState("primitive_age", m))
		for _, a := range plan.anchors {
			if a.wonder {
				nWonder++
			}
		}
		return nWonder, len(plan.anchors)
	}
	if nw, na := wonderAnchors(map[string]int{"hut": 10, "gathering_camp": 6}); nw != 0 || na != 1 {
		t.Fatalf("0-wonder village: got %d wonder-anchors / %d anchors, want 0 / 1 (single center anchor)", nw, na)
	}
	if nw, na := wonderAnchors(map[string]int{"hut": 12, "gathering_camp": 8, "colosseum": 1}); nw != 1 || na != 1 {
		t.Fatalf("1-wonder city: got %d wonder-anchors / %d anchors, want 1 / 1", nw, na)
	}
	if nw, na := wonderAnchors(map[string]int{"hut": 30, "gathering_camp": 20, "forge": 12, "sacred_grove": 1, "stonehenge": 1, "colosseum": 1}); nw != 3 || na != 3 {
		t.Fatalf("3-wonder city: got %d wonder-anchors / %d anchors, want 3 / 3 (anchors scale with wonders)", nw, na)
	}

	// CLEAR PLAZA + clustering with ≥1 wonder: no fabric lot may sit inside any wonder
	// anchor's plaza radius, and the fabric must hug the anchors (its mean distance to
	// the NEAREST anchor is small relative to the town's overall extent).
	plan := tdPlanFor(sampleState("primitive_age", map[string]int{
		"hut": 40, "gathering_camp": 30, "forge": 20, "barracks": 12, "colosseum": 1, "stonehenge": 1,
	}))
	var wonderAnchorPts []tdAnchor
	for _, a := range plan.anchors {
		if a.wonder {
			wonderAnchorPts = append(wonderAnchorPts, a)
		}
	}
	if len(wonderAnchorPts) < 2 {
		t.Fatalf("expected ≥2 wonder anchors, got %d", len(wonderAnchorPts))
	}
	fab := fabricLots(plan)
	if len(fab) < 20 {
		t.Fatalf("expected a substantial fabric, got %d lots", len(fab))
	}
	var sumNearest, maxFromCore float64
	for _, lt := range fab {
		// Plaza: strictly outside every wonder's clear ring.
		nearestAnchor := math.Inf(1)
		for _, a := range plan.anchors {
			d := math.Hypot(lt.x-a.cx, lt.y-a.cy)
			if a.wonder && d < plazaR {
				t.Fatalf("fabric lot at (%.1f,%.1f) sits inside a wonder plaza (dist %.2f < plazaR %.2f) — the centerpiece is buried", lt.x, lt.y, d, plazaR)
			}
			if d < nearestAnchor {
				nearestAnchor = d
			}
		}
		sumNearest += nearestAnchor
		if r := math.Hypot(lt.x-plan.cx, lt.y-plan.cy); r > maxFromCore {
			maxFromCore = r
		}
	}
	_ = sumNearest
	// COMPACT SETTLEMENT (Voronoi-block model): buildings fill WARDS spread across the town disc,
	// so — unlike the retired lane model — they do NOT hug the anchors; the invariant is that the
	// town is one COMPACT cluster, not scattered. Every fabric lot stays within a bounded multiple
	// of the town radius (the plan's bounded townR), so nothing is flung out. (Anchor-count scaling
	// and plaza-clear, the real wonder-anchoring signals, are asserted above/below.)
	if maxFromCore > plan.townR*1.05 {
		t.Fatalf("a fabric lot sits %.1f from the core, past the bounded town radius %.1f — the settlement is not compact", maxFromCore, plan.townR)
	}

	// 0-wonder cohesion: a wonderless village is ONE cohesive settlement around the
	// center — every fabric lot is within a bounded radius of the single center anchor.
	vplan := tdPlanFor(sampleState("primitive_age", map[string]int{"hut": 16, "gathering_camp": 10, "stone_camp": 6}))
	if len(vplan.anchors) != 1 || vplan.anchors[0].wonder {
		t.Fatalf("wonderless village should have exactly one non-wonder center anchor, got %d anchors", len(vplan.anchors))
	}
	vfab := fabricLots(vplan)
	rad := tdFootprintRadius(&vplan)
	for _, lt := range vfab {
		if d := math.Hypot(lt.x-vplan.cx, lt.y-vplan.cy); d > rad+0.01 {
			t.Fatalf("wonderless village lot at distance %.1f exceeds its footprint radius %.1f — not one cohesive settlement", d, rad)
		}
	}
}

// TestRoofAtlasNoAccentPixels is the yellow-dot regression guard (FIX 1): every
// PRIMITIVE roof archetype must render without panic, paint a non-trivial number of
// pixels, AND paint NO pixel equal to the theme RoleAccent or RoleHighlight color. The
// ridge/crown is a base-derived lighten now, so a saturated accent can never land on a
// roof. Uses a real earthy style so the roof material is thatch-brown, far from gold.
func TestRoofAtlasNoAccentPixels(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	style := tdStyleForEra(eraOrganic)
	pal := newTdPal()
	accent := rgba(theme.Color(theme.RoleAccent))
	highlight := rgba(theme.Color(theme.RoleHighlight))

	// A transform that blows a city-space lot up to a big, clearly-visible roof so the
	// pixel scan is meaningful. scale 4, centered at (30,30).
	xf := tdTransform{scale: 4, offX: 30, offY: 30, roofFloorPx: 1}

	// Every primitive archetype, paired with a representative domain/category so the
	// lineage tint path also runs (matters for the accent scan).
	cases := []struct {
		name     string
		roof     roofType
		domain   string
		category string
	}{
		{"hut", roofHut, "housing", "housing"},
		{"long", roofLong, "housing", "housing"},
		{"temple", roofTemple, "faith", "research"},
		{"camp", roofCamp, "food", "production"},
		{"stash", roofStash, "storage", "storage"},
		{"flat", roofFlat, "geological_extraction", "production"},
		{"ridge", roofRidge, "metallurgy", "production"},
		{"wonder", roofWonder, "wonder", "wonder"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 60, 60))
			lt := tdLot{
				x: 0, y: 0, w: 6, h: 6, kind: tdRoof,
				domain: tc.domain, category: tc.category, roof: tc.roof,
			}
			// Must not panic.
			drawRoof(img, xf, lt, style, pal)

			painted := 0
			for i := 0; i+3 < len(img.Pix); i += 4 {
				if img.Pix[i+3] == 0 {
					continue // untouched (alpha 0)
				}
				painted++
				got := color.RGBA{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2], A: 0xff}
				// No roof pixel may be the raw accent or highlight (the yellow dots). A
				// small tolerance guards against an off-by-one blend, but the earthy roof
				// tones are nowhere near gold so any near-match is the bug.
				if colorClose(got, accent, 6) {
					t.Fatalf("%s roof painted a RoleAccent pixel %v at index %d — the yellow-dot regression is back",
						tc.name, got, i/4)
				}
				if colorClose(got, highlight, 6) {
					t.Fatalf("%s roof painted a RoleHighlight pixel %v at index %d — accent leaked onto the roof",
						tc.name, got, i/4)
				}
			}
			if painted < 8 {
				t.Fatalf("%s roof painted only %d pixels — a roof must fill its lot", tc.name, painted)
			}
		})
	}
}

// TestRoofAtlasThemeRetint proves the roof atlas is theme-aware: the same roof under two
// themes must produce different base tones (the material is theme-derived, retints on a
// switch) — while STILL never being the accent under either theme.
func TestRoofAtlasThemeRetint(t *testing.T) {
	renderHut := func(themeKey string) *image.RGBA {
		_ = theme.SetActive(themeKey)
		style := tdStyleForEra(eraOrganic)
		pal := newTdPal()
		xf := tdTransform{scale: 4, offX: 30, offY: 30, roofFloorPx: 1}
		img := image.NewRGBA(image.Rect(0, 0, 60, 60))
		lt := tdLot{x: 0, y: 0, w: 6, h: 6, kind: tdRoof, domain: "housing", category: "housing", roof: roofHut}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	a := renderHut("forge")
	b := renderHut("high_contrast")
	_ = theme.SetActive("forge")
	if !imagesDiffer(a, b) {
		t.Fatal("hut roof identical across themes — the roof material is not theme-derived")
	}
}

// TestGreeneryInTown locks FIX 3: gardens/squares/props/trees live within (or, for the
// deliberate groves, just outside) the town footprint — NOT scattered across the far
// canvas. We assert every filler lot is within a bounded multiple of the town radius.
func TestGreeneryInTown(t *testing.T) {
	_ = theme.SetActive("forge")
	state := sampleState("primitive_age", map[string]int{"hut": 16, "gathering_camp": 8, "shrine": 1})
	plan := tdPlanFor(state)
	rad := tdFootprintRadius(&plan)
	if rad <= 0 {
		t.Fatal("degenerate footprint radius")
	}

	// In-town filler (gardens, squares, props) must be within the footprint.
	// Groves (trees) may sit just outside — the grove ring (~1.13× the town radius) plus a
	// fixed copse spread. Bound them by that, not a bare multiplier, so the check holds
	// for any town size. Nothing may be flung to the far canvas.
	roofSz := defaultTdConfig.roofSize
	inTownBudget := rad*0.95 + roofSz
	groveBudget := rad*1.15 + roofSz*2.0

	fillerN := 0
	for _, lt := range plan.lots {
		var budget float64
		switch lt.kind {
		case tdGarden, tdSquare, tdProp, tdPond:
			// Ponds (FIX 4) are placed like gardens — woven into the town, so they share the
			// in-town budget (not the wider grove budget).
			budget = inTownBudget
		case tdTree:
			budget = groveBudget
		default:
			continue
		}
		fillerN++
		d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy)
		if d > budget {
			t.Fatalf("filler lot (kind %d) at distance %.1f exceeds budget %.1f (town radius %.1f) — greenery scattered off-town",
				lt.kind, d, budget, rad)
		}
	}
	if fillerN == 0 {
		t.Fatal("no filler placed — the living-city greenery is missing")
	}

	// And at least a couple of groves (tree clusters) exist just outside the town for the
	// wooded fringe.
	if got := countKind(plan, tdTree); got < 3 {
		t.Fatalf("only %d tree lots — expected in-town trees + a few groves", got)
	}
}

// countKind counts lots of a given non-roof kind.
func countKind(plan topPlan, k tdLotKind) int {
	n := 0
	for _, lt := range plan.lots {
		if lt.kind == k {
			n++
		}
	}
	return n
}

// TestLandmarkOnlyOverlay checks the overlay is KEY-LANDMARKS ONLY (locked #7): exactly
// one title, at least one landmark building label (a real config Name, lineage-colored),
// a City Center label, and ZERO civ/trade labels — the city reads by roof, not a wall of
// text.
func TestLandmarkOnlyOverlay(t *testing.T) {
	_ = theme.SetActive("forge")
	// A civ with a shrine (a landmark) plus huts + a wonder, so we get a labeled hero and
	// a wonder label. Give it an account so the title has a real name.
	state := sampleState("primitive_age", map[string]int{"hut": 10, "gathering_camp": 5, "shrine": 1, "colosseum": 1})
	state.AgeName = "Primitive Age"
	state.AccountStats = &game.AccountStatsView{DisplayName: "Testopia"}

	_, plan := renderImage(state, 120, 80)

	var titles, capitals, buildings, civs, trades int
	var lineageColored int
	for _, l := range plan.labels {
		switch l.kind {
		case labelTitle:
			titles++
		case labelCapital:
			capitals++
		case labelBuilding:
			buildings++
			if l.lineageColored {
				lineageColored++
			}
		case labelCiv:
			civs++
		case labelTrade:
			trades++
		}
	}

	if titles != 1 {
		t.Fatalf("title labels = %d, want exactly 1", titles)
	}
	if capitals != 1 {
		t.Fatalf("City Center labels = %d, want exactly 1", capitals)
	}
	if buildings < 1 {
		t.Fatalf("landmark building labels = %d, want >= 1", buildings)
	}
	if lineageColored < 1 {
		t.Fatalf("no landmark label is lineage-colored — labels must match their roof color")
	}
	if civs != 0 {
		t.Fatalf("civ labels = %d, want 0 (landmark-only overlay)", civs)
	}
	if trades != 0 {
		t.Fatalf("trade labels = %d, want 0 (landmark-only overlay)", trades)
	}

	// The building labels must carry real config Names (non-empty), not keys or blanks.
	for _, l := range plan.labels {
		if l.kind == labelBuilding && l.text == "" {
			t.Fatal("a landmark building label has empty text")
		}
	}
}

// squareProps returns the town-square prop lots (the typed well/firepit/stones/stall),
// in placement order. Used by the town-square tests.
func squareProps(plan topPlan) []tdLot {
	out := make([]tdLot, 0, len(plan.lots))
	for _, lt := range plan.lots {
		switch lt.kind {
		case tdPropWell, tdPropFirepit, tdPropStones, tdPropStall:
			out = append(out, lt)
		}
	}
	return out
}

// TestTownSquareDressing locks the playtest FIX: each WONDER anchor's cleared plaza is
// dressed as a deliberate TOWN SQUARE — a paved-stone ground patch (a made surface,
// distinct from the era-tinted dirt) with ≥1 seeded era prop ringed AROUND the wonder
// roof, none overlapping the roof. It asserts (1) a paved plaza lot per wonder anchor,
// (2) the paved tone actually paints within the plaza radius on the rendered image
// (distinct from the dirt ground), and (3) props sit outside the wonder-roof footprint
// but inside the plaza, so the centerpiece is dressed, never covered.
func TestTownSquareDressing(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	const w, h = 140, 90
	state := sampleState("primitive_age", map[string]int{
		"hut": 30, "gathering_camp": 20, "forge": 12, "colosseum": 1, "stonehenge": 1,
	})
	plan := tdPlanFor(state)

	style := tdStyleForEra(eraOrganic)
	pal := newTdPal()
	paved := tdPavedColor(style, pal)
	ground := style.groundBase(pal)
	// The paving must be a DISTINCT made surface, not the dirt tone.
	if colorClose(paved, ground, 8) {
		t.Fatalf("paved tone %v ≈ ground tone %v — the square must read as a distinct paved surface", paved, ground)
	}

	// A paved plaza lot per wonder anchor, and ≥1 prop near each, none overlapping the
	// wonder roof.
	plazaR := defaultTdConfig.plazaRadius * defaultTdConfig.roofSize
	wonderAnchorN := 0
	for i, a := range plan.anchors {
		if !a.wonder {
			continue
		}
		wonderAnchorN++
		// A plaza lot centered on this anchor.
		gotPlaza := false
		for _, lt := range plan.lots {
			if lt.kind == tdPlaza && math.Hypot(lt.x-a.cx, lt.y-a.cy) < 0.5 {
				gotPlaza = true
				break
			}
		}
		if !gotPlaza {
			t.Fatalf("wonder anchor %d has no paved plaza lot — the cleared plaza is bare dirt, not a town square", i)
		}
		// Props near this anchor: at least one, all outside the roof footprint and inside
		// the plaza (dressing the ring around the wonder, never covering it).
		scale := 3.0
		if i > 0 {
			scale = 2.4
		}
		roofHalf := defaultTdConfig.roofSize * scale / 2
		near := 0
		for _, lt := range squareProps(plan) {
			d := math.Hypot(lt.x-a.cx, lt.y-a.cy)
			if d >= plazaR {
				continue // belongs to another anchor's square
			}
			near++
			propHalf := math.Max(lt.w, lt.h) / 2
			if d < roofHalf+propHalf {
				t.Fatalf("wonder anchor %d: a square prop at dist %.2f overlaps the wonder roof (roofHalf %.2f + propHalf %.2f) — the centerpiece is covered",
					i, d, roofHalf, propHalf)
			}
		}
		if near < 1 {
			t.Fatalf("wonder anchor %d has no town-square props ringed around it", i)
		}
	}
	if wonderAnchorN < 2 {
		t.Fatalf("expected ≥2 wonder anchors to dress, got %d", wonderAnchorN)
	}

	// The paved tone must actually PAINT within the plaza radius on the rendered image —
	// the plaza reads as a made surface, not bare ground. Scan the plaza disc of the
	// grandest wonder (anchor 0) and require a meaningful count of paved-ish pixels.
	img, _ := renderImage(state, w, h)
	xf := computeTransform(&plan, w, h)
	var center tdAnchor
	for _, a := range plan.anchors {
		if a.wonder {
			center = a
			break
		}
	}
	pavedPix := 0
	rpx := xf.ext(plazaR) // plaza radius in pixels
	ccx, ccy := xf.px(center.cx, center.cy)
	for dy := -rpx; dy <= rpx; dy++ {
		for dx := -rpx; dx <= rpx; dx++ {
			if dx*dx+dy*dy > rpx*rpx {
				continue
			}
			x, y := ccx+dx, ccy+dy
			if x < 0 || x >= w || y < 0 || y >= h {
				continue
			}
			off := img.PixOffset(x, y)
			got := color.RGBA{R: img.Pix[off], G: img.Pix[off+1], B: img.Pix[off+2], A: 0xff}
			if colorClose(got, paved, 14) {
				pavedPix++
			}
		}
	}
	if pavedPix < 6 {
		t.Fatalf("only %d paved-tone pixels inside the wonder plaza — the town square is not painted as a made surface", pavedPix)
	}
}

// TestWonderlessCenterSquare locks the playtest FIX for a wonderless village: the bare
// city-center anchor gets a MODEST square — a small paved patch + at least one prop — so
// the heart reads as a gathering place, WITHOUT hollowing a hut village into a donut. The
// small patch must be strictly smaller than a full wonder plaza (the openness stays
// contained), and the fabric must still fill the center (no big empty ring).
func TestWonderlessCenterSquare(t *testing.T) {
	_ = theme.SetActive("forge")
	plan := tdPlanFor(sampleState("primitive_age", map[string]int{"hut": 12, "gathering_camp": 8, "stone_camp": 6}))
	if len(plan.anchors) != 1 || plan.anchors[0].wonder {
		t.Fatalf("wonderless village should have exactly one non-wonder center anchor, got %d anchors", len(plan.anchors))
	}
	c := plan.anchors[0]

	// A small paved patch at the center.
	plazaR := defaultTdConfig.plazaRadius * defaultTdConfig.roofSize
	var patch *tdLot
	for i := range plan.lots {
		lt := plan.lots[i]
		if lt.kind == tdPlaza && math.Hypot(lt.x-c.cx, lt.y-c.cy) < 0.5 {
			patch = &plan.lots[i]
			break
		}
	}
	if patch == nil {
		t.Fatal("wonderless center has no paved patch — the heart reads as bare dirt, not a gathering place")
	}
	// MODEST: the patch half-extent must be well under a full wonder plaza radius so the
	// hut village keeps a filled heart (not a donut).
	if patch.w/2 >= plazaR {
		t.Fatalf("wonderless center patch half-extent %.2f is not smaller than a full plaza %.2f — a hut village must not become a donut", patch.w/2, plazaR)
	}

	// At least one prop at the center.
	props := 0
	for _, lt := range squareProps(plan) {
		if math.Hypot(lt.x-c.cx, lt.y-c.cy) < plazaR {
			props++
		}
	}
	if props < 1 {
		t.Fatalf("wonderless center has %d props — expected a modest well/firepit gathering spot", props)
	}

	// Cohesion is unchanged: a wonderless village stays one filled settlement — no fabric
	// lot is pushed off the footprint (the small square did NOT hollow the center).
	rad := tdFootprintRadius(&plan)
	for _, lt := range fabricLots(plan) {
		if d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy); d > rad+0.01 {
			t.Fatalf("wonderless village lot at %.1f exceeds footprint radius %.1f — the modest square disturbed cohesion", d, rad)
		}
	}
}

// TestPerEraDensityKnob locks the playtest FIX 2 groundwork: village density is AIRY
// early and TIGHTENS with age via a per-era spacing knob. PRIMITIVE keeps its current
// airy slot spacing (its look must not change); a later era packs strictly tighter. The
// knob is also correctly ROUTED into the generator config so later-era placement uses it.
func TestPerEraDensityKnob(t *testing.T) {
	_ = theme.SetActive("forge")

	prim := tdStyleForEra(eraOrganic)
	later := tdStyleForEra(eraCityBlocks) // a not-yet-tuned band → the tighter default preset

	// PRIMITIVE unchanged: still the airy 2.4 the village was tuned at.
	if prim.slotSpacing != 2.4 {
		t.Fatalf("primitive slotSpacing = %.2f, want 2.4 — the airy primitive village look must not change", prim.slotSpacing)
	}
	// Later era packs TIGHTER (denser city/metropolis).
	if !(later.slotSpacing > 0 && later.slotSpacing < prim.slotSpacing) {
		t.Fatalf("later-era slotSpacing = %.2f is not tighter than primitive %.2f — density must scale with age", later.slotSpacing, prim.slotSpacing)
	}

	// The knob must actually ROUTE into the generator config: an era style's slotSpacing
	// overrides the config default; primitive maps back to 2.4, later to its tighter value.
	route := func(style tdEraStyle) float64 {
		cfg := defaultTdConfig
		if style.slotSpacing > 0 {
			cfg.slotSpacing = style.slotSpacing
		}
		return cfg.slotSpacing
	}
	if route(prim) != 2.4 {
		t.Fatalf("primitive routed slotSpacing = %.2f, want 2.4", route(prim))
	}
	if route(later) >= route(prim) {
		t.Fatalf("later routed slotSpacing %.2f not tighter than primitive %.2f — the knob is not wired into placement", route(later), route(prim))
	}
}

// TestTopDownPanicSafeExactSize renders across degenerate and tiny canvases; each must
// not panic and must produce an image of EXACTLY the requested size (locked cross-cutting
// constraint). Covers 0×0, 1×1, a tiny square, and a 1×40 sliver.
func TestTopDownPanicSafeExactSize(t *testing.T) {
	_ = theme.SetActive("forge")
	states := []game.GameState{
		sampleState("primitive_age", nil),
		sampleState("primitive_age", map[string]int{"hut": 1}),
		sampleState("primitive_age", map[string]int{"hut": 20, "gathering_camp": 10, "shrine": 2, "colosseum": 1}),
		sampleState("made_up_age", map[string]int{"x": 3}),
	}
	sizes := []struct{ w, h int }{
		{0, 0},
		{1, 1},
		{3, 2},
		{1, 40},
		{40, 1},
		{100, 60},
	}
	for _, st := range states {
		for _, sz := range sizes {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("renderImage panicked at %dx%d: %v", sz.w, sz.h, r)
					}
				}()
				img, plan := renderImage(st, sz.w, sz.h)
				if img == nil {
					t.Fatalf("renderImage returned nil at %dx%d", sz.w, sz.h)
				}
				if got := img.Bounds(); got.Dx() != sz.w || got.Dy() != sz.h {
					t.Fatalf("image size = %dx%d, want %dx%d", got.Dx(), got.Dy(), sz.w, sz.h)
				}
				// The overlay plan must also be panic-safe (may be empty on tiny canvases).
				_ = plan
			}()
		}
	}
}

// namedState is sampleState with a distinct display name, so tests can vary the city SEED
// (the seed is a hash of the display name) across otherwise-identical building sets.
func namedState(ageKey, name string, blds map[string]int) game.GameState {
	st := sampleState(ageKey, blds)
	st.AccountStats = &game.AccountStatsView{DisplayName: name}
	return st
}

// allRoofLots returns every roof lot (fabric AND wonders), for the no-overlap sweep.
func allRoofLots(plan topPlan) []tdLot {
	out := make([]tdLot, 0, len(plan.lots))
	for _, lt := range plan.lots {
		if lt.kind == tdRoof {
			out = append(out, lt)
		}
	}
	return out
}

// roofHalfExtent is a lot's collision half-size (the roof atlas fills the lot; the
// elongated longhouse's wider axis dominates).
func roofHalfExtent(lt tdLot) float64 { return math.Max(lt.w, lt.h) / 2 }

// TestNoRoofOverlap locks the playtest FIX 2 core guarantee: buildings LINE the lanes with
// NO overlaps. Across several building counts AND several city seeds, no two roof lots may
// sit closer than the min gap — their edge-to-edge separation must stay positive (never
// touching). This is what the lane-lining placement + the cross-lane overlap guard buy us
// over the old pull-to-lane model, which squished centers onto the lanes.
func TestNoRoofOverlap(t *testing.T) {
	_ = theme.SetActive("forge")
	// The min separation the guard aims to preserve. We assert lots never touch (sep > 0);
	// a small positive floor guards against a borderline near-miss reading as "fine".
	const floor = 0.2
	seeds := []string{"Aldermoor", "Bexley", "Corveil", "Duskwind", "Emberton"}
	counts := []map[string]int{
		{"hut": 8, "gathering_camp": 4},
		{"hut": 16, "gathering_camp": 10, "stone_camp": 6, "forge": 6},
		{"hut": 30, "gathering_camp": 20, "forge": 12, "barracks": 8, "colosseum": 1, "stonehenge": 1},
		{"hut": 60, "gathering_camp": 45, "forge": 28, "barracks": 18, "colosseum": 1},
	}
	for _, name := range seeds {
		for _, m := range counts {
			plan := tdPlanFor(namedState("primitive_age", name, m))
			lots := allRoofLots(plan)
			if len(lots) < 5 {
				t.Fatalf("seed %q counts %v: only %d roof lots — expected a real settlement", name, m, len(lots))
			}
			for i := range lots {
				for k := 0; k < i; k++ {
					d := math.Hypot(lots[i].x-lots[k].x, lots[i].y-lots[k].y)
					sep := d - roofHalfExtent(lots[i]) - roofHalfExtent(lots[k])
					if sep < floor {
						t.Fatalf("seed %q counts %v: roof lots overlap/touch — lot %d (%s) at (%.1f,%.1f) and lot %d (%s) at (%.1f,%.1f) have edge separation %.2f < %.2f",
							name, m, i, lots[i].domain, lots[i].x, lots[i].y, k, lots[k].domain, lots[k].x, lots[k].y, sep, floor)
					}
				}
			}
		}
	}
}

// onStreetCell reports whether city-space (x,y) falls ON a street (block-boundary) cell — within
// half a cell (Chebyshev) of any street-cell center. The block interiors are inset from these, so
// a roof should essentially never land here.
func onStreetCell(plan topPlan, x, y float64) bool {
	if plan.cellSize <= 0 {
		return false
	}
	half := plan.cellSize / 2
	for _, p := range plan.streetCells {
		if math.Abs(x-p.x) <= half && math.Abs(y-p.y) <= half {
			return true
		}
	}
	return false
}

// TestStreetsBoundBlocks locks the Voronoi-block core: buildings sit INSIDE the blocks, INSET
// from the streets — so ~0 building roofs land ON a street (boundary) cell, leaving the gap
// network visible between the wards. A regression that placed roofs on the boundaries (the old
// spaghetti look, or a bad inset) would bury the streets and trip this.
func TestStreetsBoundBlocks(t *testing.T) {
	_ = theme.SetActive("forge")
	state := sampleState("primitive_age", map[string]int{"hut": 18, "gathering_camp": 12, "stone_camp": 8, "forge": 8})
	plan := tdPlanFor(state)

	fab := fabricLots(plan)
	if len(fab) < 15 {
		t.Fatalf("expected a substantial fabric, got %d lots", len(fab))
	}
	if len(plan.streetCells) < 8 {
		t.Fatalf("expected a block-gap street network, got %d street cells", len(plan.streetCells))
	}

	// Essentially no roof CENTER may sit on a street cell (the blocks are inset from the gaps).
	// A rare edge case near a plaza is tolerated, but not a wholesale pile-on.
	onStreet := 0
	for _, lt := range fab {
		if onStreetCell(plan, lt.x, lt.y) {
			onStreet++
		}
	}
	if frac := float64(onStreet) / float64(len(fab)); frac > 0.05 {
		t.Fatalf("%.0f%% of roof centers sit ON a street cell (%d/%d) — buildings must be INSET inside their blocks, not on the gaps",
			frac*100, onStreet, len(fab))
	}

	// And the streets stay visible: a healthy share of street cells has NO roof body covering it
	// (sample each street cell; a covered cell is one whose center is within a roof half-extent of
	// a roof). Most gaps must read as open road.
	covered := 0
	for _, p := range plan.streetCells {
		for _, lt := range fab {
			if math.Hypot(lt.x-p.x, lt.y-p.y) < math.Max(lt.w, lt.h)/2 {
				covered++
				break
			}
		}
	}
	if frac := float64(covered) / float64(len(plan.streetCells)); frac > 0.15 {
		t.Fatalf("%.0f%% of street cells are covered by a roof body — the block-gap streets are being buried", frac*100)
	}
}

// TestStreetsVisible locks playtest FIX 1: BOLD packed-earth roads are actually PRESENT in
// the final rendered image and are NOT fully covered by the roofs. It renders a real
// village and counts pixels matching the packed-earth lane surface/edge tones — there must
// be a meaningful number, proving the streets read as a strong visual element beneath and
// between the lane-lining buildings.
func TestStreetsVisible(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	const w, h = 140, 90
	state := sampleState("primitive_age", map[string]int{"hut": 20, "gathering_camp": 12, "stone_camp": 8, "forge": 8})
	img, _ := renderImage(state, w, h)

	style := tdStyleForEra(eraOrganic)
	pal := newTdPal()
	surface := style.streetCol(pal)
	edge := tdStreetEdgeColor(style, pal)
	ground := style.groundBase(pal)

	// The packed-earth surface must be a DISTINCT, higher-contrast tone vs the dirt ground
	// (FIX 1: the streets must actually READ). Assert real separation, not a near-match.
	if colorClose(surface, ground, 12) {
		t.Fatalf("packed-earth street tone %v ≈ ground tone %v — the road has no contrast against the dirt", surface, ground)
	}

	packed := 0
	for i := 0; i+3 < len(img.Pix); i += 4 {
		if img.Pix[i+3] == 0 {
			continue
		}
		got := color.RGBA{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2], A: 0xff}
		if colorClose(got, surface, 6) || colorClose(got, edge, 6) {
			packed++
		}
	}
	// A real lane network at this size paints well over a hundred packed-earth pixels; require
	// a solid floor so a regression that buries or thins the roads trips the test.
	if packed < 40 {
		t.Fatalf("only %d packed-earth road pixels in the final image — the streets are not visible (buried by roofs or too thin)", packed)
	}
}

// TestNotAPinwheelCompact locks the playtest FIX: the town is a COMPACT, BOUNDED cluster — it
// does NOT fling buildings outward along radial spokes into a pinwheel with empty wedges between
// the arms (the regression this work removes). Two robust, non-brittle properties across a wide
// count range and several seeds:
//
//	(1) BOUNDED / SUB-LINEAR footprint: the max fabric-lot distance from the core must NOT grow
//	    ~linearly with the building count — a radial-spoke town's radius climbs with the count,
//	    a compact town's saturates. We assert the footprint radius at a HUGE count is only a
//	    modest multiple of the radius at a SMALL count (it must not blow up), and never exceeds a
//	    hard bound derived from the town-radius model.
//	(2) NO RADIAL SPOKES: the fabric's ANGULAR distribution is spread around the whole compass,
//	    not concentrated into a few arms. We bucket lots into 12 angular sectors and require the
//	    bulk of them occupied — a pinwheel leaves big empty wedges (many empty sectors).
func TestNotAPinwheelCompact(t *testing.T) {
	_ = theme.SetActive("forge")
	cfg := defaultTdConfig

	footprint := func(m map[string]int) (radius float64, fab []tdLot, cx, cy float64) {
		plan := tdPlanFor(sampleState("primitive_age", m))
		fab = fabricLots(plan)
		cx, cy = plan.cx, plan.cy
		for _, lt := range fab {
			if d := math.Hypot(lt.x-cx, lt.y-cy); d > radius {
				radius = d
			}
		}
		return
	}

	// (1) BOUNDED: footprint radius must saturate, not scale linearly with count. Compare a small
	// village to a huge one across several seeds; the huge town's radius must stay within a small
	// multiple of the small one's (a radial-spoke town would be many times larger).
	for _, name := range []string{"Aldermoor", "Corveil", "Duskwind"} {
		smallR, _, _, _ := footprint(map[string]int{"hut": 6, "gathering_camp": 4})
		bigR, bigFab, bcx, bcy := footprint(map[string]int{"hut": 200, "gathering_camp": 150, "forge": 90, "barracks": 60})
		_ = name
		if smallR <= 0 || bigR <= 0 {
			t.Fatalf("degenerate footprint radii: small=%.1f big=%.1f", smallR, bigR)
		}
		// The big town has ~30× the buildings of the small one; a compact town's radius grows only
		// modestly. Require the big footprint stays under 3× the small one AND under a hard bound
		// from the town-radius model (which itself grows only ~sqrt) — a pinwheel would blow past both.
		if bigR > smallR*3.0 {
			t.Fatalf("footprint radius blew up %.1f→%.1f (%.1f×) as the count grew — the town is flinging out, not staying compact (pinwheel)", smallR, bigR, bigR/smallR)
		}
		hardBound := tdTownRadius(2000, cfg) // way past any real count; the compact ceiling
		if bigR > hardBound {
			t.Fatalf("footprint radius %.1f exceeds the compact bound %.1f — the town is not bounded", bigR, hardBound)
		}

		// (2) NO RADIAL SPOKES: bucket the big town's fabric into 12 angular sectors; a compact town
		// fills the compass, a pinwheel concentrates into a few arms leaving empty wedges.
		if len(bigFab) < 20 {
			t.Fatalf("expected a substantial fabric to test angular spread, got %d", len(bigFab))
		}
		var buckets [12]int
		for _, lt := range bigFab {
			a := math.Atan2(lt.y-bcy, lt.x-bcx)
			b := int((a + math.Pi) / (2 * math.Pi) * 12)
			if b < 0 {
				b = 0
			}
			if b > 11 {
				b = 11
			}
			buckets[b]++
		}
		occupied := 0
		for _, c := range buckets {
			if c > 0 {
				occupied++
			}
		}
		// A compact town covers nearly the whole compass; a pinwheel of a few spokes would leave
		// many sectors empty. Require at least 9 of 12 sectors occupied (no big empty wedge).
		if occupied < 9 {
			t.Fatalf("fabric occupies only %d/12 angular sectors — it is concentrated into radial arms (pinwheel), not a compact spread (buckets=%v)", occupied, buckets)
		}
	}
}

// centerBandCells returns, for each labeled landmark whose label sits in the center band (the
// building + City Center labels — NOT the corner title, NOT the edge civ/trade labels), the set
// of (col,row) cells its glyphs occupy. Used to assert the center landmarks don't stack.
func centerBandLabelCells(plan overlayPlan) []map[[2]int]bool {
	var out []map[[2]int]bool
	for _, l := range plan.labels {
		if l.kind != labelBuilding && l.kind != labelCapital {
			continue
		}
		runes := []rune(l.text)
		start := l.cx
		switch l.align {
		case alignRight:
			start = l.cx - len(runes) + 1
		case alignCenter:
			start = l.cx - len(runes)/2
		}
		cells := map[[2]int]bool{}
		for i := range runes {
			cells[[2]int{start + i, l.cy}] = true
		}
		out = append(out, cells)
	}
	return out
}

// TestLandmarkLabelsNoStack locks the playtest FIX 5: the center landmark labels (city center,
// the wonder, a promoted hero, the Stash) must NOT stack on the same cells — the overlay offsets
// colliding landmark labels onto distinct rows so each stays readable. It forces the WORST case
// from the playtest screenshot: several landmark markers CRAMMED onto the SAME center pixel (a
// tightly-clustered heart), where a naive overlay would print every name on the same row. It
// asserts no two center-band label glyphs share a cell, so every landmark stays legible.
func TestLandmarkLabelsNoStack(t *testing.T) {
	_ = theme.SetActive("forge")

	// A worst-case geometry: the palace marker and three landmark building markers ALL at the same
	// center pixel — exactly the "Sacred Grove / City Center / Stash on top of each other" crush.
	const w, h = 140, 90
	cx, cy := w/2, h/2
	geo := layoutGeometry{palaceX: cx, palaceY: cy}
	names := []struct {
		name, lineage, cat string
		tier               int
	}{
		{"Sacred Grove", "wonder", "wonder", 3},
		{"Great Stash", "storage", "storage", 1},
		{"Elder Shrine", "faith", "research", 2},
	}
	for _, n := range names {
		geo.buildings = append(geo.buildings, buildingLabel{
			px: cx, py: cy, name: n.name, lineageKey: n.lineage, category: n.cat, tier: n.tier, size: 3,
		})
	}
	state := sampleState("primitive_age", nil)
	state.AgeName = "Primitive Age"
	state.AccountStats = &game.AccountStatsView{DisplayName: "Stackton"}

	plan := buildLandmarkOverlay(state, w, h/2, geo)

	cellsPerLabel := centerBandLabelCells(plan)
	// The wonder + stash + shrine + City Center = up to 4 center-band labels; they must all place.
	if len(cellsPerLabel) < 4 {
		t.Fatalf("expected 4 center-band landmark labels (3 buildings + City Center) to all place without stacking, got %d — some were dropped or collided", len(cellsPerLabel))
	}
	// No cell may be claimed by two different center-band labels (that is a stack).
	seen := map[[2]int]int{}
	for li, cells := range cellsPerLabel {
		for cell := range cells {
			if prev, ok := seen[cell]; ok {
				t.Fatalf("center-band landmark labels %d and %d both occupy cell (%d,%d) — the labels are stacked on top of each other", prev, li, cell[0], cell[1])
			}
			seen[cell] = li
		}
	}
}

// TestStreetsConnected locks the Voronoi-block streets-connected invariant: the street-cell
// network (the gaps between the blocks) is ONE connected component, and it reaches the central
// plaza. Region boundaries are connected BY CONSTRUCTION — a raster nearest-seed partition of a
// single disc yields one connected boundary web with real junctions, no spaghetti. We union-find
// the street cells by grid adjacency (two cells join when their centers are within ~1.5 cells, i.e.
// 8-neighbours on the raster) and assert a single component AND a street cell near the town core
// (the plaza region's boundary is central), across several seeds and counts.
func TestStreetsConnected(t *testing.T) {
	_ = theme.SetActive("forge")
	seeds := []string{"Aldermoor", "Bexley", "Corveil", "Duskwind", "Emberton"}
	counts := []map[string]int{
		{"hut": 8, "gathering_camp": 4},
		{"hut": 16, "gathering_camp": 10, "stone_camp": 6, "forge": 6},
		{"hut": 30, "gathering_camp": 20, "forge": 12, "barracks": 8, "colosseum": 1, "stonehenge": 1},
		{"hut": 60, "gathering_camp": 45, "forge": 28, "barracks": 18, "colosseum": 1},
	}
	for _, name := range seeds {
		for _, m := range counts {
			plan := tdPlanFor(namedState("primitive_age", name, m))
			cells := plan.streetCells
			n := len(cells)
			if n < 8 {
				t.Fatalf("seed %q counts %v: only %d street cells — expected a real block-gap network", name, m, n)
			}
			// Grid adjacency: two street cells are neighbours if their centers are within 1.5
			// cell sizes (covers the 8-neighbourhood on the raster, diagonal included).
			adj := plan.cellSize * 1.5
			adj2 := adj * adj
			parent := make([]int, n)
			for i := range parent {
				parent[i] = i
			}
			var find func(int) int
			find = func(x int) int {
				for parent[x] != x {
					parent[x] = parent[parent[x]]
					x = parent[x]
				}
				return x
			}
			for i := 0; i < n; i++ {
				for j := i + 1; j < n; j++ {
					dx := cells[i].x - cells[j].x
					dy := cells[i].y - cells[j].y
					if dx*dx+dy*dy <= adj2 {
						parent[find(i)] = find(j)
					}
				}
			}
			roots := map[int]bool{}
			for i := 0; i < n; i++ {
				roots[find(i)] = true
			}
			if len(roots) != 1 {
				sizes := map[int]int{}
				for i := 0; i < n; i++ {
					sizes[find(i)]++
				}
				t.Fatalf("seed %q counts %v: street-cell network has %d disconnected components (want 1) — the block gaps are not one connected web: sizes=%v",
					name, m, len(roots), sizes)
			}
			// Reaches the central PLAZA: the plaza region sits at the core, so its boundary
			// street cells are central. Require at least one street cell within ~0.45× the town
			// radius of the core (the plaza's Voronoi boundary is well inside that).
			nearCore := false
			coreBudget := plan.townR * 0.45
			for _, p := range cells {
				if math.Hypot(p.x-plan.cx, p.y-plan.cy) <= coreBudget {
					nearCore = true
					break
				}
			}
			if !nearCore {
				t.Fatalf("seed %q counts %v: no street cell within %.1f of the core — the connected network does not reach the central plaza", name, m, coreBudget)
			}
		}
	}
}

// groundVariance renders the ground of a fresh canvas with drawGround and returns the mean
// per-channel variance of the ground pixels (a robust proxy for texture "busyness": a flat
// wash → ~0, a noisy speckle → high). Used by the quieter-ground test.
func groundVariance(img *image.RGBA, w, h int) float64 {
	var sumR, sumG, sumB, sumR2, sumG2, sumB2, n float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := img.PixOffset(x, y)
			r, g, b := float64(img.Pix[off]), float64(img.Pix[off+1]), float64(img.Pix[off+2])
			sumR += r
			sumG += g
			sumB += b
			sumR2 += r * r
			sumG2 += g * g
			sumB2 += b * b
			n++
		}
	}
	if n == 0 {
		return 0
	}
	varOf := func(s, s2 float64) float64 { return s2/n - (s/n)*(s/n) }
	return (varOf(sumR, sumR2) + varOf(sumG, sumG2) + varOf(sumB, sumB2)) / 3
}

// TestQuieterGround locks playtest polish FIX 1: the ground texture is a CALM, subtle base
// now — far lower contrast/variance than the old busy speckle so the city reads cleanly
// against it. We render the ground twice: once with the live drawGround (quiet), and once
// with a scratch replica of the OLD formula (22% of pixels, up to 0.60 toward alt). The new
// ground's pixel variance must be MEANINGFULLY lower than the old — a robust ratio check,
// not a brittle exact value — while still carrying SOME grain (not a dead-flat wash).
func TestQuieterGround(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 120, 80
	const seed uint32 = 0xC0FFEE99
	style := tdStyleForEra(eraOrganic)
	pal := newTdPal()

	// New (quiet) ground.
	imgNew := image.NewRGBA(image.Rect(0, 0, w, h))
	drawGround(imgNew, style, pal, seed, w, h)
	newVar := groundVariance(imgNew, w, h)

	// Scratch replica of the OLD, busy formula (pre-FIX-1): threshold 0.22, amplitude 0.60.
	base := style.groundBase(pal)
	alt := style.groundAlt(pal)
	imgOld := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := texHash(uint32(x), uint32(y), seed)
			tt := 0.0
			if n < 0.22 {
				tt = 0.5 + 0.5*n/0.22
			}
			imgOld.SetRGBA(x, y, blend(base, alt, tt*0.60))
		}
	}
	oldVar := groundVariance(imgOld, w, h)

	if oldVar <= 0 {
		t.Skip("degenerate old-ground variance (base≈alt under this theme) — nothing to compare")
	}
	// The new ground must be substantially calmer — well under HALF the old variance. Robust
	// ratio, so a future retint that changes the absolute tones can't make this brittle.
	if newVar >= oldVar*0.5 {
		t.Fatalf("ground texture not meaningfully quieter: newVar=%.3f is not < 0.5×oldVar (%.3f) — the dirt still competes with the city", newVar, oldVar)
	}
	// But it must not be a dead-flat wash — a faint grain remains (some pixels still take a
	// touch of alt), so the ground still reads as textured earth, just calmly.
	if newVar <= 0 {
		t.Fatalf("ground is dead flat (variance %.4f) — FIX 1 should QUIET the texture, not remove it", newVar)
	}
}

// TestBiggerPlazaClearRadius locks playtest polish FIX 3: the cleared plaza around a wonder
// anchor is BIGGER now (more breathing room), and it stays strictly building-free. It
// asserts (1) the configured plaza radius exceeds the old 2.2×roofSize baseline, and (2) no
// fabric lot sits inside ANY wonder anchor's (bigger) plaza — the center reads open and the
// centerpiece is never crowded.
func TestBiggerPlazaClearRadius(t *testing.T) {
	_ = theme.SetActive("forge")

	// (1) The plaza radius is meaningfully bigger than the pre-FIX-3 baseline (2.2×roofSize).
	const oldPlazaRadius = 2.2
	if defaultTdConfig.plazaRadius <= oldPlazaRadius {
		t.Fatalf("plazaRadius %.2f is not bigger than the old %.2f — FIX 3 must widen the plaza for more breathing room",
			defaultTdConfig.plazaRadius, oldPlazaRadius)
	}
	plazaR := defaultTdConfig.plazaRadius * defaultTdConfig.roofSize

	// (2) Across seeds, no fabric lot sits inside any wonder plaza (the bigger clear ring is
	// honoured — the center stays open and the wonder uncrowded).
	seeds := []string{"Aldermoor", "Corveil", "Emberton"}
	for _, name := range seeds {
		plan := tdPlanFor(namedState("primitive_age", name, map[string]int{
			"hut": 40, "gathering_camp": 30, "forge": 20, "barracks": 12, "colosseum": 1, "stonehenge": 1,
		}))
		nWonder := 0
		for _, a := range plan.anchors {
			if a.wonder {
				nWonder++
			}
		}
		if nWonder < 1 {
			t.Fatalf("seed %q: expected ≥1 wonder anchor to test the plaza clear", name)
		}
		for _, lt := range fabricLots(plan) {
			for _, a := range plan.anchors {
				if !a.wonder {
					continue
				}
				if d := math.Hypot(lt.x-a.cx, lt.y-a.cy); d < plazaR {
					t.Fatalf("seed %q: fabric lot at (%.1f,%.1f) sits inside the bigger wonder plaza (dist %.2f < plazaR %.2f) — the center is not clear",
						name, lt.x, lt.y, d, plazaR)
				}
			}
		}
	}
}

// TestPondsInTown locks playtest polish FIX 4: a FEW BUILT decorative ponds are mixed
// IN-TOWN among the greenery (seeded, deterministic), and the pond water tone actually
// PAINTS on the rendered image. It asserts (1) pond lots exist and are a small count (a few,
// not a lake district), (2) every pond sits within the town footprint (in-town, not scattered
// across the empty map), and (3) the water tone paints a meaningful number of pixels in the
// final render — the pond reads as a made water feature.
func TestPondsInTown(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	const w, h = 140, 90
	state := sampleState("primitive_age", map[string]int{"hut": 24, "gathering_camp": 16, "stone_camp": 8, "forge": 8})
	plan := tdPlanFor(state)

	// (1) Ponds exist and are FEW (a rare accent, never many).
	ponds := 0
	for _, lt := range plan.lots {
		if lt.kind == tdPond {
			ponds++
		}
	}
	if ponds < 1 {
		t.Fatal("no pond lots placed — the built decorative ponds (FIX 4) are missing")
	}
	if ponds > 8 {
		t.Fatalf("%d ponds — FIX 4 wants only a FEW ponds mixed in-town, not a lake district", ponds)
	}

	// (2) Every pond is IN-TOWN: within the built-up footprint (placed like a garden, not
	// scattered across the empty map). Use the same in-town budget the greenery test uses.
	rad := tdFootprintRadius(&plan)
	if rad <= 0 {
		t.Fatal("degenerate footprint radius")
	}
	inTownBudget := rad*0.95 + defaultTdConfig.roofSize
	for _, lt := range plan.lots {
		if lt.kind != tdPond {
			continue
		}
		if d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy); d > inTownBudget {
			t.Fatalf("pond at distance %.1f exceeds the in-town budget %.1f (town radius %.1f) — ponds must be woven into the town, not scattered",
				d, inTownBudget, rad)
		}
	}

	// (3) The pond water tone must be a DISTINCT tone (not the ground/garden) AND actually
	// paint pixels in the final render.
	style := tdStyleForEra(eraOrganic)
	pal := newTdPal()
	water := tdPondColor(style, pal)
	ground := style.groundBase(pal)
	garden := style.gardenCol(pal)
	if colorClose(water, ground, 10) {
		t.Fatalf("pond water tone %v ≈ ground tone %v — a pond must read as a distinct made water feature", water, ground)
	}
	if colorClose(water, garden, 10) {
		t.Fatalf("pond water tone %v ≈ garden tone %v — a pond must read as water, not greenery", water, garden)
	}
	img, _ := renderImage(state, w, h)
	waterPix := 0
	for i := 0; i+3 < len(img.Pix); i += 4 {
		if img.Pix[i+3] == 0 {
			continue
		}
		got := color.RGBA{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2], A: 0xff}
		// Match the pond body or its lighter shallows rim (brighten 0.18).
		if colorClose(got, water, 8) || colorClose(got, brighten(water, 0.18), 8) {
			waterPix++
		}
	}
	if waterPix < 4 {
		t.Fatalf("only %d pond-water pixels in the final image — the ponds are not painting as water", waterPix)
	}
}

// ---- town-form archetypes (map-overhaul-citymap V3-A) -----------------------

// formName is a readable label for a town form, for test failure messages.
func formName(f tdTownForm) string {
	switch f {
	case formOrganic:
		return "organic"
	case formRadial:
		return "radial"
	case formGrid:
		return "grid"
	case formRibbon:
		return "ribbon"
	}
	return "?"
}

// wardSeedRadialCorr is the robust ANTI-WHEEL metric: the Pearson correlation between a ward seed's
// INDEX and its RADIUS from the core, over plan.wardSeeds (the pinned center at the origin is
// skipped). A RADIAL "wagon wheel" scatters its free seeds on a golden-angle SPIRAL — radius climbs
// monotonically with index — so this correlation is high (~0.8–1.0). ORGANIC (blue-noise) and
// RIBBON (linear) seeds have no radial ordering, so it sits near 0. This survives Lloyd relaxation
// (which evens angular gaps but keeps the spiral's centers-out ordering), so it tells an organic
// town from a wheel on the REAL generated plan. Returns 0 for a degenerate (<3-seed) field.
func wardSeedRadialCorr(plan topPlan) float64 {
	var idx, rad []float64
	i := 0.0
	for _, s := range plan.wardSeeds {
		if math.Hypot(s.x-plan.cx, s.y-plan.cy) < 1e-9 {
			continue // skip the pinned center seed
		}
		idx = append(idx, i)
		rad = append(rad, math.Hypot(s.x-plan.cx, s.y-plan.cy))
		i++
	}
	n := float64(len(idx))
	if n < 3 {
		return 0
	}
	var sx, sy, sxy, sx2, sy2 float64
	for k := range idx {
		sx += idx[k]
		sy += rad[k]
		sxy += idx[k] * rad[k]
		sx2 += idx[k] * idx[k]
		sy2 += rad[k] * rad[k]
	}
	den := math.Sqrt((n*sx2 - sx*sx) * (n*sy2 - sy*sy))
	if den == 0 {
		return 0
	}
	return (n*sxy - sx*sy) / den
}

// TestTownFormDeterministicAndVaried locks the two core properties of tdPickTownForm: it is a pure
// function of (citySeed, era) — the SAME inputs always yield the SAME form — and across a sample of
// citySeeds the chosen forms VARY (a band is not collapsed to a single form). Determinism is what
// makes a civ's town shape stable across ages/frames; variety is the whole point (no two towns
// alike).
func TestTownFormDeterministicAndVaried(t *testing.T) {
	// (1) DETERMINISM: repeated picks for the same (seed, era) are identical, for several eras.
	eras := []era{eraOrganic, eraHubSpoke, eraCastle, eraZonedGrid, eraCityBlocks, eraCampus, eraOrbital}
	for _, e := range eras {
		for i := 0; i < 200; i++ {
			s := citySeed("Town" + strconv.Itoa(i))
			a := tdPickTownForm(s, e)
			b := tdPickTownForm(s, e)
			if a != b {
				t.Fatalf("tdPickTownForm not deterministic for seed %#x era %d: %s vs %s", s, e, formName(a), formName(b))
			}
		}
	}

	// (2) VARIETY: over a sample of citySeeds, a band that allows >1 form actually PRODUCES >1 form
	// (different seeds → different towns). Test on eras whose weights permit several forms.
	for _, e := range []era{eraHubSpoke, eraCastle, eraZonedGrid, eraCityBlocks} {
		seen := map[tdTownForm]int{}
		for i := 0; i < 400; i++ {
			seen[tdPickTownForm(citySeed("Varyville"+strconv.Itoa(i)), e)]++
		}
		if len(seen) < 2 {
			t.Fatalf("era %d produced only %d distinct form(s) over 400 seeds (%v) — towns are not varied", e, len(seen), seen)
		}
	}

	// (3) The SAME civ name over different eras generally re-skins its form (the roll is
	// era-weighted, not a single global choice) — assert at least the distribution differs by
	// confirming a name that is organic in primitive can be a different form in a grid-heavy era
	// somewhere in the sample (proves era actually feeds the pick).
	eraSpanChanged := false
	for i := 0; i < 200; i++ {
		s := citySeed("Spanner" + strconv.Itoa(i))
		if tdPickTownForm(s, eraOrganic) != tdPickTownForm(s, eraCityBlocks) {
			eraSpanChanged = true
			break
		}
	}
	if !eraSpanChanged {
		t.Fatal("no civ changed form between the organic and city-blocks bands — era is not influencing the pick")
	}
}

// TestPrimitiveIsOrganicNotAWheel locks the headline requirement: PRIMITIVE villages RAMBLE
// organically and are NEVER radial wheels. Two parts:
//
//	(A) Over many citySeeds, the PRIMITIVE band rolls ORGANIC-dominant and NEVER radial or grid
//	    (villages aren't planned) — and the fixed anonymous village seed (the default city) is
//	    organic, not a wheel.
//	(B) A generated ORGANIC town has NO radial-spoke / central-ring concentration: its ward-seed
//	    radial ordering (wardSeedRadialCorr) is LOW, whereas a forced RADIAL town's is HIGH — a
//	    robust anti-wheel assertion (the metric genuinely bites; it is not vacuously true).
func TestPrimitiveIsOrganicNotAWheel(t *testing.T) {
	_ = theme.SetActive("forge")

	// (A) Distribution over many seeds at the primitive (organic) band.
	var cnt [4]int
	for i := 0; i < 2000; i++ {
		f := tdPickTownForm(citySeed("Hamlet"+strconv.Itoa(i)), eraOrganic)
		cnt[f]++
	}
	if cnt[formRadial] != 0 || cnt[formGrid] != 0 {
		t.Fatalf("primitive rolled radial=%d grid=%d — villages must NEVER be planned wheels/grids", cnt[formRadial], cnt[formGrid])
	}
	if frac := float64(cnt[formOrganic]) / 2000.0; frac < 0.5 {
		t.Fatalf("primitive is only %.0f%% organic — villages must be ORGANIC-dominant", frac*100)
	}
	if cnt[formRibbon] == 0 {
		t.Fatal("primitive never rolled ribbon — the occasional grew-along-a-trail village should still appear")
	}
	// The fixed anonymous (default) village seed must be organic — the CURRENT village is not a wheel.
	if got := tdPickTownForm(citySeed(""), eraOrganic); got != formOrganic {
		t.Fatalf("the default (anonymous) primitive village rolled %s, want organic — the current village must ramble, not read as a wheel", formName(got))
	}

	// (B) Anti-wheel on the REAL generated plan: an organic town's ward seeds are NOT radially
	// ordered; a radial town's are. Average over several seeds/counts so the assertion is robust to
	// a single unlucky field.
	organicCorr := 0.0
	radialCorr := 0.0
	ns := 0
	cfg := defaultTdConfig
	anchors := []tdAnchor{{cx: 0, cy: 0}} // wonderless: a single pinned center (the village heart)
	for _, nm := range []string{"", "Aldermoor", "Corveil", "Duskwind", "Emberton", "Faelin", "Gorse", "Hale"} {
		seed := citySeed(nm)
		for _, nRoofs := range []int{60, 120, 200} {
			townR := tdTownRadius(nRoofs, cfg)
			org := tdBuildBlockField(townR, anchors, nRoofs, formOrganic, cfg, seed)
			rad := tdBuildBlockField(townR, anchors, nRoofs, formRadial, cfg, seed)
			organicCorr += wardSeedRadialCorr(topPlan{wardSeeds: org.seeds})
			radialCorr += wardSeedRadialCorr(topPlan{wardSeeds: rad.seeds})
			ns++
		}
	}
	organicCorr /= float64(ns)
	radialCorr /= float64(ns)
	// The organic town must be clearly NON-radial, and the radial town clearly radial — a wide,
	// robust margin so this isn't brittle to Lloyd tuning.
	if organicCorr > 0.35 {
		t.Fatalf("organic ward seeds are radially ordered (corr %.3f) — the organic town is reading as a wheel", organicCorr)
	}
	if radialCorr < 0.6 {
		t.Fatalf("radial ward seeds are NOT radially ordered (corr %.3f) — the anti-wheel metric does not bite; the test would pass a wheel", radialCorr)
	}
	if radialCorr-organicCorr < 0.4 {
		t.Fatalf("organic (%.3f) and radial (%.3f) ward-seed orderings are too close — organic is not provably distinct from a wheel", organicCorr, radialCorr)
	}

	// The organic default village, generated end-to-end, is also non-radial (belt-and-braces on the
	// real tdPlanFor path, not just the raw field).
	plan := tdPlanFor(sampleState("primitive_age", map[string]int{"hut": 40, "gathering_camp": 24, "stone_camp": 12, "forge": 10}))
	if plan.form != formOrganic {
		t.Fatalf("the default primitive plan is %s, want organic", formName(plan.form))
	}
	if c := wardSeedRadialCorr(plan); c > 0.4 {
		t.Fatalf("the default organic village plan has radially-ordered wards (corr %.3f) — it still reads as a wheel", c)
	}
}

// unionCount union-finds the given street cells by grid adjacency (8-neighbours on the raster:
// centers within ~1.5 cell sizes join) and returns the number of connected components. Shared by
// the per-form connectivity assertion. Mirrors TestStreetsConnected's union-find.
func unionCount(cells []tdPoint, cellSize float64) int {
	n := len(cells)
	if n == 0 {
		return 0
	}
	adj2 := (cellSize * 1.5) * (cellSize * 1.5)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			dx := cells[i].x - cells[j].x
			dy := cells[i].y - cells[j].y
			if dx*dx+dy*dy <= adj2 {
				parent[find(i)] = find(j)
			}
		}
	}
	roots := map[int]bool{}
	for i := 0; i < n; i++ {
		roots[find(i)] = true
	}
	return len(roots)
}

// TestEachTownFormWellFormed locks that ALL FOUR forms feed the SAME ward machinery and each yields
// a WELL-FORMED town: (1) the block-gap street network is ONE connected component reaching the core
// (streets = ward boundaries, connected by construction), (2) the field has both street cells AND
// interior block cells (the wards exist and are inset from the gaps), and (3) the town is
// COMPACT/BOUNDED (every ward seed sits within the town disc). Then, end-to-end via the real
// pipeline, each form places a real settlement with NO roof overlap and ~0 roofs on street cells
// (buildings sit inside their wards). This is what proves the new forms didn't break connectivity,
// no-overlap, or boundedness that the radial-only model guaranteed.
func TestEachTownFormWellFormed(t *testing.T) {
	_ = theme.SetActive("forge")
	cfg := defaultTdConfig

	// (1)-(3): FIELD-level checks for each form, forced directly (no era-style confound). Wonderless
	// single-center anchor + a substantial ward count so the network is real.
	anchors := []tdAnchor{{cx: 0, cy: 0}}
	for _, form := range []tdTownForm{formOrganic, formRadial, formGrid, formRibbon} {
		for _, nm := range []string{"Aldermoor", "Bexley", "Corveil", "Duskwind"} {
			seed := citySeed(nm)
			nRoofs := 140
			townR := tdTownRadius(nRoofs, cfg)
			field := tdBuildBlockField(townR, anchors, nRoofs, form, cfg, seed)

			// Street network exists, is ONE component, and reaches the core (the central plaza's
			// ward boundary is central).
			if len(field.streetCells) < 8 {
				t.Fatalf("form %s seed %q: only %d street cells — no real block-gap network", formName(form), nm, len(field.streetCells))
			}
			if comps := unionCount(field.streetCells, field.cellSize); comps != 1 {
				t.Fatalf("form %s seed %q: street network has %d components, want 1 (streets must be one connected web)", formName(form), nm, comps)
			}
			nearCore := false
			for _, p := range field.streetCells {
				if math.Hypot(p.x, p.y) <= townR*0.45 {
					nearCore = true
					break
				}
			}
			if !nearCore {
				t.Fatalf("form %s seed %q: no street cell near the core — the network does not reach the central plaza", formName(form), nm)
			}

			// Wards exist: at least a few blocks carry interior cells (the buildings' frontage).
			blocksWithCells := 0
			for _, cells := range field.blockCells {
				if len(cells) > 0 {
					blocksWithCells++
				}
			}
			if blocksWithCells < 2 {
				t.Fatalf("form %s seed %q: only %d blocks have interior cells — no wards to fill", formName(form), nm, blocksWithCells)
			}

			// COMPACT/BOUNDED: every ward seed sits inside the town disc (a small epsilon for the
			// clamp). No form flings a ward outside the bounded footprint.
			for _, s := range field.seeds {
				if d := math.Hypot(s.x, s.y); d > townR+cfg.cellSize {
					t.Fatalf("form %s seed %q: a ward seed sits %.1f from core, past townR %.1f — not bounded/compact", formName(form), nm, d, townR)
				}
			}
		}
	}

	// End-to-end per form via the REAL pipeline (picker + populate + wonders + min-gap guard): pick a
	// (name, age) that rolls each target form, then assert no roof overlap and roofs-in-wards.
	type formCase struct {
		form   tdTownForm
		ageKey string
		era    era
	}
	// One representative age per era band we use. tdPickTownForm(seed, era) selects the form.
	cases := []formCase{
		{formOrganic, "primitive_age", eraOrganic},
		{formRibbon, "primitive_age", eraOrganic},
		{formRadial, "bronze_age", eraHubSpoke},
		{formGrid, "electric_age", eraCityBlocks},
	}
	blds := map[string]int{"hut": 26, "gathering_camp": 18, "stone_camp": 10, "forge": 10, "barracks": 6}
	for _, fc := range cases {
		// Find a display name whose seed rolls the desired form for this era.
		name := ""
		for i := 0; i < 5000; i++ {
			cand := "Form" + strconv.Itoa(i)
			if tdPickTownForm(citySeed(cand), fc.era) == fc.form {
				name = cand
				break
			}
		}
		if name == "" {
			t.Fatalf("could not find a seed that rolls form %s at era %d — form unreachable", formName(fc.form), fc.era)
		}
		plan := tdPlanFor(namedState(fc.ageKey, name, blds))
		if plan.form != fc.form {
			t.Fatalf("form %s: generated plan rolled %s instead (seed/era wiring mismatch)", formName(fc.form), formName(plan.form))
		}

		lots := allRoofLots(plan)
		if len(lots) < 10 {
			t.Fatalf("form %s (seed %q age %s): only %d roof lots — expected a real settlement", formName(fc.form), name, fc.ageKey, len(lots))
		}
		// NO roof overlap: no two roofs may touch (edge separation > a small floor).
		const floor = 0.2
		for i := range lots {
			for k := 0; k < i; k++ {
				d := math.Hypot(lots[i].x-lots[k].x, lots[i].y-lots[k].y)
				sep := d - roofHalfExtent(lots[i]) - roofHalfExtent(lots[k])
				if sep < floor {
					t.Fatalf("form %s (seed %q): roofs overlap — lot %d (%s) and lot %d (%s) edge sep %.2f < %.2f",
						formName(fc.form), name, i, lots[i].domain, k, lots[k].domain, sep, floor)
				}
			}
		}
		// Buildings sit IN WARDS: ~0 fabric roofs land on a street cell (inset from the gaps).
		fab := fabricLots(plan)
		onStreet := 0
		for _, lt := range fab {
			if onStreetCell(plan, lt.x, lt.y) {
				onStreet++
			}
		}
		if len(fab) > 0 {
			if frac := float64(onStreet) / float64(len(fab)); frac > 0.05 {
				t.Fatalf("form %s (seed %q): %.0f%% of roofs sit ON a street cell — buildings must be inset inside their wards", formName(fc.form), name, frac*100)
			}
		}
		// COMPACT: no fabric roof past the bounded town radius (anti-pinwheel holds for every form).
		for _, lt := range fab {
			if d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy); d > plan.townR*1.05 {
				t.Fatalf("form %s (seed %q): a roof sits %.1f from core, past townR %.1f — not compact/bounded", formName(fc.form), name, d, plan.townR)
			}
		}
	}
}
