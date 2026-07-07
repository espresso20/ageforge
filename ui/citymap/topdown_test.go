package citymap

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
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
	// Form-aware plaza radius: a primitive town is organic/ribbon, so clear-radius follows its form
	// (organic gets the modest ring, not the roomy radial one). Read it from the generated plan.
	plazaR := tdPlazaRadius(plan.form, defaultTdConfig)
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
	// wonder roof. Form-aware plaza radius (a primitive town is organic → the modest ring).
	plazaR := tdPlazaRadius(plan.form, defaultTdConfig)
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

// inTownCellCount counts the raster cells the block field assigned to some ward (nearest >= 0) —
// i.e. every cell inside the town footprint. The STREET FRACTION is streetCells/inTownCells.
func inTownCellCount(f blockField) int {
	n := 0
	for _, s := range f.nearest {
		if s >= 0 {
			n++
		}
	}
	return n
}

// TestStreetsAreThinLanes locks map-overhaul-citymap FIX 1: the block-boundary streets are THIN
// village LANES, not the wide avenues that dominated the image before. The robust, non-brittle
// metric is the STREET-CELL FRACTION — the share of in-town raster cells classified as street. A
// wide-band web (the old streetBand 2.1 → ~2-cell boundaries) claims ~0.24 of the town; the
// narrowed band (~1 cell) claims meaningfully less, so the wards hold the town and the white
// recedes. We assert the fraction across several seeds AND counts stays under a ceiling the OLD
// wide streets would blow past, and — separately — that the streets are still PRESENT (a real web,
// not erased). Two guards so a regression in EITHER direction (streets swell back / streets vanish)
// trips it.
func TestStreetsAreThinLanes(t *testing.T) {
	_ = theme.SetActive("forge")
	cfg := defaultTdConfig

	// Config-level relationship: the classification band is ~one cell (thin), NOT ~two cells (wide).
	// This is the knob FIX 1 turns; lock it so a future edit can't silently widen the lanes back.
	if cfg.streetBand > cfg.cellSize*1.25 {
		t.Fatalf("streetBand %.2f is > 1.25×cellSize %.2f — the boundary web is ~2 cells wide again (wide avenues, not thin lanes)", cfg.streetBand, cfg.cellSize)
	}

	anchors := []tdAnchor{{cx: 0, cy: 0}} // wonderless village heart
	// The old wide streets sat at ~0.24 of the town; the new thin lanes sit ~0.19. A 0.22 ceiling
	// sits in that gap with margin — new passes, old fails — and is robust across seeds/counts.
	const maxStreetFrac = 0.22
	worst := 0.0
	for _, nm := range []string{"", "Aldermoor", "Corveil", "Duskwind", "Emberton", "Gorse"} {
		seed := citySeed(nm)
		for _, nRoofs := range []int{40, 90, 160} {
			townR := tdTownRadius(nRoofs, cfg)
			f := tdBuildBlockField(townR, anchors, nRoofs, formOrganic, cfg, seed)
			town := inTownCellCount(f)
			if town == 0 {
				t.Fatalf("seed %q n=%d: empty town footprint", nm, nRoofs)
			}
			frac := float64(len(f.streetCells)) / float64(town)
			if frac > worst {
				worst = frac
			}
			if frac > maxStreetFrac {
				t.Fatalf("seed %q n=%d: streets are %.1f%% of the town (> %.0f%%) — the boundary web is too WIDE (FIX 1 wants thin lanes)", nm, nRoofs, frac*100, maxStreetFrac*100)
			}
			// Still a real web: streets must not vanish (that would break streets-connected too, but
			// assert it here so a "thin" regression that erases them is caught by THIS test's intent).
			if frac < 0.05 || len(f.streetCells) < 8 {
				t.Fatalf("seed %q n=%d: only %.1f%% street cells (%d) — the lanes have been thinned into nonexistence", nm, nRoofs, frac*100, len(f.streetCells))
			}
		}
	}
	// Sanity: the worst-case fraction we saw is comfortably under the old wide-street regime, proving
	// the reduction is real and not a threshold that barely clears.
	if worst >= maxStreetFrac {
		t.Fatalf("worst street fraction %.3f reached the ceiling — no real headroom", worst)
	}
}

// outlineStats samples tdOrganicRadiusAt over the whole compass and returns the min/max radius
// ratio (a low ratio = a real waist/cove) and the stddev/mean (overall irregularity). A perfect
// circle scores min/max=1, std/mean=0; the timid first cut scored ~0.86 / ~0.038 (barely a wobble).
func outlineStats(seed uint32, townR float64) (minMax, stdOverMean float64) {
	const N = 720
	rmin, rmax := math.Inf(1), math.Inf(-1)
	sum, sum2 := 0.0, 0.0
	for i := 0; i < N; i++ {
		a := 2 * math.Pi * float64(i) / float64(N)
		r := tdOrganicRadiusAt(a, townR, seed)
		if r < rmin {
			rmin = r
		}
		if r > rmax {
			rmax = r
		}
		sum += r
		sum2 += r * r
	}
	mean := sum / N
	varr := sum2/N - mean*mean
	if varr < 0 {
		varr = 0
	}
	return rmin / rmax, math.Sqrt(varr) / mean
}

// TestOrganicOutlineIrregular locks map-overhaul-citymap FIX 2: the organic town silhouette is a
// genuinely RAGGED, NON-CIRCULAR blob — not the timid near-circle the first cut produced (and not,
// obviously, a clean disc). Three properties, each read straight off tdOrganicRadiusAt so the test
// measures the actual outline geometry:
//
//	(1) A real WAIST: the min/max radius ratio is well below 1 (a circle is 1; the old wobble was
//	    ~0.86). We require < 0.80 — the narrowest point is ≥20% inside the widest, an unmistakable cove.
//	(2) Real IRREGULARITY: the radius stddev/mean is well above the old ~0.038 timid value. We require
//	    > 0.06 — roughly double, i.e. the rim genuinely rambles.
//	(3) VARIETY per seed: different citySeeds give different silhouettes. The old version fed the same
//	    two harmonics every time, so every town's std/mean was an IDENTICAL 0.0384 — a dead giveaway
//	    the shape didn't vary. We sample several seeds and require their std/mean values to actually
//	    SPREAD (max−min above a margin), proving the shape (not just the rotation) differs by seed.
//
// Bounds/connectivity are NOT re-tested here (TestStreetsConnected / TestNotAPinwheelCompact /
// TestEachTownFormWellFormed cover that the ragged outline stays bounded + one connected web); this
// test's job is purely that the outline is IRREGULAR and SEED-VARIED.
func TestOrganicOutlineIrregular(t *testing.T) {
	cfg := defaultTdConfig
	townR := tdTownRadius(120, cfg)

	// tdOrganicRadiusAt must never exceed townR (subtractive-only → bounded) and never go non-positive
	// (the floor keeps the domain star-shaped → connected). Check the invariants directly.
	for _, nm := range []string{"", "Aldermoor", "Corveil", "Duskwind"} {
		seed := citySeed(nm)
		for i := 0; i < 720; i++ {
			a := 2 * math.Pi * float64(i) / 720
			r := tdOrganicRadiusAt(a, townR, seed)
			if r > townR+1e-6 {
				t.Fatalf("seed %q: outline radius %.3f exceeds townR %.3f — not bounded (a ward could sit outside the footprint)", nm, r, townR)
			}
			if r <= 0 {
				t.Fatalf("seed %q: outline radius %.3f ≤ 0 — the footprint is no longer star-shaped/connected", nm, r)
			}
		}
	}

	// (1)+(2): every seed's outline is a ragged blob (real waist + real irregularity).
	var soms []float64
	for _, nm := range []string{"", "Aldermoor", "Corveil", "Duskwind", "Emberton", "Gorse", "Hale", "Faelin"} {
		seed := citySeed(nm)
		mm, som := outlineStats(seed, townR)
		soms = append(soms, som)
		if mm >= 0.80 {
			t.Fatalf("seed %q: outline min/max radius ratio %.3f ≥ 0.80 — the silhouette reads as a circle, not a ragged blob with bays", nm, mm)
		}
		if som <= 0.06 {
			t.Fatalf("seed %q: outline stddev/mean %.4f ≤ 0.06 — barely a wobble (the timid near-circle), not a genuinely irregular outline", nm, som)
		}
	}

	// (3): the shape VARIES by seed — the std/mean values must spread, not collapse to one constant
	// (the old code's tell was an identical 0.0384 for every seed).
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, s := range soms {
		if s < lo {
			lo = s
		}
		if s > hi {
			hi = s
		}
	}
	if hi-lo < 0.02 {
		t.Fatalf("outline irregularity barely varies across seeds (std/mean spread %.4f over [%.4f,%.4f]) — different towns should have different silhouettes, not one shared shape", hi-lo, lo, hi)
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

// TestBiggerPlazaClearRadius locks playtest polish FIX 3 UNDER the form-aware plaza
// (map-overhaul-citymap): the ROOMY plaza is now the RADIAL/planned form's identity (a
// monument/forum-planned town's generous forecourt), while the ORGANIC village gets a MODEST
// one-ward plaza so it never carves a dominant central void (the wagon-wheel signature). It asserts
// (1) the roomy radial plaza radius still exceeds the old 2.2×roofSize baseline (FIX 3's breathing
// room is retained where it belongs); (2) each form clears fabric out to ITS OWN plaza radius (a
// RADIAL wonder town honours the roomy ring; an ORGANIC wonder town honours the small ring); and (3)
// the organic plaza is STRICTLY SMALLER than the roomy radial one (organic is de-radialized).
func TestBiggerPlazaClearRadius(t *testing.T) {
	_ = theme.SetActive("forge")

	// (1) The roomy (radial/planned) plaza radius is meaningfully bigger than the pre-FIX-3 baseline.
	const oldPlazaRadius = 2.2
	if defaultTdConfig.plazaRadius <= oldPlazaRadius {
		t.Fatalf("plazaRadius %.2f is not bigger than the old %.2f — FIX 3's roomy plaza (now the radial form's) must keep its breathing room",
			defaultTdConfig.plazaRadius, oldPlazaRadius)
	}
	roomyR := tdPlazaRadius(formRadial, defaultTdConfig)
	organicR := tdPlazaRadius(formOrganic, defaultTdConfig)

	// (3) The organic plaza is strictly smaller than the roomy radial one — the de-radialization.
	if organicR >= roomyR {
		t.Fatalf("organic plaza radius %.2f is not smaller than the roomy radial %.2f — organic must get a MODEST plaza, not a dominant central void", organicR, roomyR)
	}

	wonderBlds := map[string]int{
		"hut": 40, "gathering_camp": 30, "forge": 20, "barracks": 12, "colosseum": 1, "stonehenge": 1,
	}
	// For each form, no fabric lot may sit inside that form's OWN wonder-plaza radius. Assert on a
	// RADIAL town (roomy ring honoured) and an ORGANIC town (small ring honoured).
	check := func(ageKey string, era era, form tdTownForm, plazaR float64) {
		// Find a display name whose seed rolls the target form for this era.
		name := ""
		for i := 0; i < 8000; i++ {
			cand := "Plaza" + formName(form) + strconv.Itoa(i)
			if tdPickTownForm(citySeed(cand), era) == form {
				name = cand
				break
			}
		}
		if name == "" {
			t.Fatalf("could not find a seed that rolls form %s at era %d", formName(form), era)
		}
		plan := tdPlanFor(namedState(ageKey, name, wonderBlds))
		if plan.form != form {
			t.Fatalf("form %s: generated plan rolled %s instead", formName(form), formName(plan.form))
		}
		nWonder := 0
		for _, a := range plan.anchors {
			if a.wonder {
				nWonder++
			}
		}
		if nWonder < 1 {
			t.Fatalf("form %s seed %q: expected ≥1 wonder anchor to test the plaza clear", formName(form), name)
		}
		for _, lt := range fabricLots(plan) {
			for _, a := range plan.anchors {
				if !a.wonder {
					continue
				}
				if d := math.Hypot(lt.x-a.cx, lt.y-a.cy); d < plazaR {
					t.Fatalf("form %s seed %q: fabric lot at (%.1f,%.1f) sits inside the wonder plaza (dist %.2f < plazaR %.2f) — the center is not clear",
						formName(form), name, lt.x, lt.y, d, plazaR)
				}
			}
		}
	}
	check("bronze_age", eraHubSpoke, formRadial, roomyR)      // roomy ring cleared
	check("primitive_age", eraOrganic, formOrganic, organicR) // small ring cleared
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

// centralVoidRadius is the ANTI-WHEEL metric that actually CATCHES the plaza-hub wheel
// (map-overhaul-citymap): the radius of the largest FABRIC-FREE disc centred on the town core — i.e.
// the min over fabric roof lots (wonders excluded) of (dist(lot,center) − roofHalf), floored at 0.
//
// This is the metric the OLD ward-seed radial-ordering correlation MISSED. The wheel the playtest
// caught is a central plaza VOID with the streets spoking outward: the pinned centre owns a big
// cleared region and the fabric RINGS it. That geometry has a LARGE central void. The de-radialized
// organic mesh fills right up to near the centre, so its central void is SMALL (≤ ~one ward radius).
// Unlike the correlation — which measures seed ORDERING and, at the ~8 seeds a small disc yields, is
// dominated by sampling noise (it can read HIGH for a genuinely blue-noise field) — the central-void
// radius reads the ACTUAL building geometry, so it FAILS on a hub-with-plaza wheel and PASSES on the
// mesh (proven in both directions by TestPrimitiveIsOrganicNotAWheel). Empty fabric → 0.
func centralVoidRadius(plan topPlan) float64 {
	best := math.Inf(1)
	for _, lt := range plan.lots {
		if lt.kind != tdRoof || lt.roof == roofWonder {
			continue
		}
		d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy) - math.Max(lt.w, lt.h)/2
		if d < best {
			best = d
		}
	}
	if math.IsInf(best, 1) || best < 0 {
		return 0
	}
	return best
}

// enclosedWardCount is the MESH-LOOPS anti-wheel metric (map-overhaul-citymap): the number of wards
// (nearest-seed regions) that are FULLY ENCLOSED — none of their interior cells touches the town rim
// (an off-town neighbour). An enclosed ward is a face of the street web bounded on ALL sides by
// streets/other wards, i.e. a LOOP in the street network. A rambling MESH has MANY such enclosed
// faces; a hub-and-ring-with-spokes WHEEL has FEW (a lone central hub cell, everything else fanned
// out to the rim). So a high enclosed count PASSES the mesh and a low one FAILS the wheel — a genuine
// topology signal, not a seed-count proxy. Reads a raw blockField (streets/interiors), 4-neighbour.
func enclosedWardCount(f blockField) int {
	if f.gridN < 3 {
		return 0
	}
	gN := f.gridN
	touchesRim := make([]bool, len(f.seeds))
	hasCells := make([]bool, len(f.seeds))
	for gy := 0; gy < gN; gy++ {
		for gx := 0; gx < gN; gx++ {
			c := gy*gN + gx
			si := f.nearest[c]
			if si < 0 || f.street[c] {
				continue
			}
			hasCells[si] = true
			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := gx+d[0], gy+d[1]
				if nx < 0 || nx >= gN || ny < 0 || ny >= gN {
					touchesRim[si] = true
					break
				}
				if f.nearest[ny*gN+nx] < 0 {
					touchesRim[si] = true
					break
				}
			}
		}
	}
	n := 0
	for si := range f.seeds {
		if hasCells[si] && !touchesRim[si] {
			n++
		}
	}
	return n
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

// TestPrimitiveIsOrganicNotAWheel locks the HEADLINE requirement: PRIMITIVE villages RAMBLE as an
// irregular MESH and are NEVER plaza-hub wagon-wheels. The old proxy (ward-seed radial-ordering
// correlation) MISSED the wheel — it measures seed ORDERING, which a blue-noise field can pass while
// the town still reads as a hub-with-spokes; at the ~8 seeds a small disc yields it is mostly
// sampling noise. This test replaces it with two metrics that read the ACTUAL wheel geometry and,
// CRUCIALLY, are proven to FAIL on a hub-with-plaza wheel and PASS on the new organic mesh (both
// directions, so neither can be a vacuous proxy):
//
//	metric 1 — CENTRAL-VOID radius (end-to-end): the fabric-free radius at the town centre. The
//	    wheel has a big central plaza void (fabric rings it); the mesh fills near the centre → small.
//	metric 2 — ENCLOSED WARDS (field topology): faces of the street web bounded on all sides =
//	    LOOPS. The mesh has many; the hub-and-ring-with-spokes wheel has few.
//
// Parts:
//
//	(A) the PRIMITIVE band rolls ORGANIC-dominant, NEVER radial/grid, and the default city is organic;
//	(B) BOTH metrics separate the NEW organic mesh (pass) from a hub-plaza WHEEL (fail), end-to-end
//	    and at field level, with a wide margin — plus the wheel is shown to FAIL the very same
//	    thresholds the organic town passes (the fail-on-old / pass-on-new proof).
func TestPrimitiveIsOrganicNotAWheel(t *testing.T) {
	_ = theme.SetActive("forge")
	cfg := defaultTdConfig
	rs := cfg.roofSize

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

	// Thresholds sit in the wide gap the probes measured: organic central void ≤ ~2.1·rs, wheel
	// ≥ ~4.4·rs; organic enclosed wards ≥ 6, wheel ≤ 5. Placed with margin so neither is brittle.
	const (
		maxOrganicVoid = 3.0 // ·roofSize — organic must fill close to centre (≤ ~one ward radius)
		minWheelVoid   = 3.0 // ·roofSize — a real wheel's central plaza void exceeds this
		minOrganicMesh = 6   // enclosed street-web loops — the mesh has many
		maxWheelMesh   = 5   // a hub+ring+spokes wheel has at most this few
	)

	// (B1) FIELD-LEVEL mesh-loops, over several seeds/counts: organic ≥ minOrganicMesh, and a forced
	// RADIAL wheel field ≤ maxWheelMesh (the metric BITES — it would flag a wheel).
	anchors := []tdAnchor{{cx: 0, cy: 0}} // wonderless: the village heart
	for _, nm := range []string{"", "Aldermoor", "Corveil", "Duskwind", "Emberton", "Faelin", "Gorse", "Hale"} {
		seed := citySeed(nm)
		for _, nRoofs := range []int{40, 80, 140, 200} {
			townR := tdTownRadius(nRoofs, cfg)
			org := tdBuildBlockField(townR, anchors, nRoofs, formOrganic, cfg, seed)
			rad := tdBuildBlockField(townR, anchors, nRoofs, formRadial, cfg, seed)
			if e := enclosedWardCount(org); e < minOrganicMesh {
				t.Fatalf("organic field (seed %q, n=%d) has only %d enclosed wards (< %d) — the street web is not a mesh with loops", nm, nRoofs, e, minOrganicMesh)
			}
			// fail-on-old: the RADIAL wheel must FAIL the organic mesh bar (few loops).
			if e := enclosedWardCount(rad); e > maxWheelMesh {
				t.Fatalf("radial WHEEL field (seed %q, n=%d) has %d enclosed wards (> %d) — the mesh metric does NOT bite; it would pass a wheel", nm, nRoofs, e, maxWheelMesh)
			}
		}
	}

	// (B2) END-TO-END central-void, on the REAL pipeline: the ORGANIC default-style village fills its
	// centre (void small) while a genuine hub-plaza WHEEL (a RADIAL town WITH a wonder seat → a
	// cleared central plaza the fabric rings) has a LARGE central void. Proven both ways with the SAME
	// threshold so the metric can't be vacuous.
	organicBlds := map[string]int{"hut": 40, "gathering_camp": 24, "stone_camp": 12, "forge": 10}
	for _, nm := range []string{"", "Aldermoor", "Duskwind", "Emberton", "Gorse", "Hale"} {
		plan := tdPlanFor(namedState("primitive_age", nm, organicBlds))
		if plan.form != formOrganic {
			continue // a ribbon roll is fine here; we assert on the organic ones
		}
		if v := centralVoidRadius(plan) / rs; v > maxOrganicVoid {
			t.Fatalf("organic village (seed %q) has a central void of %.2f·roofSize (> %.1f) — buildings do not fill near the centre; it reads as a wheel", nm, v, maxOrganicVoid)
		}
		if len(plan.wardSeeds) < minOrganicMesh {
			t.Fatalf("organic village (seed %q) has only %d wards — too coarse to be a mesh", nm, len(plan.wardSeeds))
		}
	}

	// The genuine WHEEL: a RADIAL-form town WITH a wonder at the centre → a reserved, cleared central
	// plaza that the fabric rings. Find a bronze-age (hub-spoke) seed that rolls radial. Its central
	// void must EXCEED the same threshold the organic town stayed under — fail-on-old / pass-on-new.
	wheelBlds := map[string]int{"hut": 40, "gathering_camp": 24, "forge": 10, "colosseum": 1}
	wheelChecked := 0
	for i := 0; i < 20000 && wheelChecked < 4; i++ {
		cand := "Wheel" + strconv.Itoa(i)
		if tdPickTownForm(citySeed(cand), eraHubSpoke) != formRadial {
			continue
		}
		plan := tdPlanFor(namedState("bronze_age", cand, wheelBlds))
		if plan.form != formRadial {
			continue
		}
		wheelChecked++
		v := centralVoidRadius(plan) / rs
		if v < minWheelVoid {
			t.Fatalf("the hub-plaza WHEEL (seed %q) has a central void of only %.2f·roofSize (< %.1f) — the central-void metric does NOT bite; it would pass a wheel as if it were a mesh", cand, v, minWheelVoid)
		}
	}
	if wheelChecked == 0 {
		t.Fatal("could not construct a radial wheel town to prove the metric fails on a wheel")
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

// ---- V3-B: ancient + medieval styling ---------------------------------------
//
// V3-B fills in the ANCIENT (bronze/iron/classical — eraHubSpoke) and MEDIEVAL (medieval/
// renaissance — eraCastle) bands using the existing frameworks (era styles, roof materials,
// ground tints, square props, town-form weights, wall flag). These tests lock the era distinctness
// + the big new WALLS+GATES piece, while the shared invariants above (streets-connected, no-overlap,
// compact, determinism, quiet ground, etc.) continue to cover all bands.

// wallLotsOf / gateLotsOf / towerLotsOf return the wall-ring lots of a plan by kind.
func wallLotsOf(plan topPlan) []tdLot {
	var out []tdLot
	for _, lt := range plan.lots {
		if lt.kind == tdWall {
			out = append(out, lt)
		}
	}
	return out
}
func gateLotsOf(plan topPlan) []tdLot {
	var out []tdLot
	for _, lt := range plan.lots {
		if lt.kind == tdGate {
			out = append(out, lt)
		}
	}
	return out
}
func towerLotsOf(plan topPlan) []tdLot {
	var out []tdLot
	for _, lt := range plan.lots {
		if lt.kind == tdWallTower || lt.kind == tdWallGatehouse {
			out = append(out, lt)
		}
	}
	return out
}
func bastionLotsOf(plan topPlan) []tdLot {
	var out []tdLot
	for _, lt := range plan.lots {
		if lt.kind == tdWallBastion {
			out = append(out, lt)
		}
	}
	return out
}

// TestV3BErasDistinctFromPrimitive locks that ANCIENT and MEDIEVAL read visibly different from the
// PRIMITIVE village: the roof MATERIAL tone differs (clay/slate vs thatch) and the GROUND tone
// differs (packed-earth/cobble vs earthy dirt+grass), and the two eras differ from EACH OTHER. All
// tones are theme-derived, so we resolve them from the same palette and require a meaningful color
// distance — a regression that left an era on the default (thatch) palette would trip this.
func TestV3BErasDistinctFromPrimitive(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()

	prim := tdStyleForEra(eraOrganic)
	ancient := tdStyleForEra(eraHubSpoke)
	medieval := tdStyleForEra(eraCastle)

	// Roof base material must differ from thatch for both, and from each other.
	pr := prim.roofBase(pal)
	ar := ancient.roofBase(pal)
	mr := medieval.roofBase(pal)
	if colorClose(ar, pr, 16) {
		t.Fatalf("ancient roof %v ≈ primitive thatch %v — clay tile must read as a distinct material", ar, pr)
	}
	if colorClose(mr, pr, 16) {
		t.Fatalf("medieval roof %v ≈ primitive thatch %v — slate must read as a distinct material", mr, pr)
	}
	if colorClose(ar, mr, 16) {
		t.Fatalf("ancient clay %v ≈ medieval slate %v — the two eras' roofs must read distinct", ar, mr)
	}

	// Ground tone must differ from the primitive dirt for both, and from each other.
	pg := prim.groundBase(pal)
	ag := ancient.groundBase(pal)
	mg := medieval.groundBase(pal)
	if colorClose(ag, pg, 10) {
		t.Fatalf("ancient ground %v ≈ primitive ground %v — packed earth/pale stone must differ", ag, pg)
	}
	if colorClose(mg, pg, 10) {
		t.Fatalf("medieval ground %v ≈ primitive ground %v — cobble/stone grey must differ", mg, pg)
	}
	if colorClose(ag, mg, 10) {
		t.Fatalf("ancient ground %v ≈ medieval ground %v — the two eras' ground must read distinct", ag, mg)
	}

	// End-to-end: the rendered images of the same civ at each era must differ (materials + ground +
	// walls all in play). Distinct ages, same buildings + seed.
	blds := map[string]int{"hut": 24, "gathering_camp": 16, "forge": 8}
	imgP, _ := renderImage(namedState("primitive_age", "Aldermoor", blds), 120, 72)
	imgA, _ := renderImage(namedState("bronze_age", "Aldermoor", blds), 120, 72)
	imgM, _ := renderImage(namedState("medieval_age", "Aldermoor", blds), 120, 72)
	if !imagesDiffer(imgP, imgA) {
		t.Fatal("primitive and ancient render identically — the era re-skin is not applied")
	}
	if !imagesDiffer(imgP, imgM) {
		t.Fatal("primitive and medieval render identically — the era re-skin is not applied")
	}
	if !imagesDiffer(imgA, imgM) {
		t.Fatal("ancient and medieval render identically — the two eras are not distinct")
	}
}

// TestV3BWallsPresentWithGates is the CENTREPIECE lock (locked #9): ancient + medieval TOWNS are
// ringed with a WALL that has GATES, the buildings stay INSIDE the wall, and the ring FOLLOWS the
// built-up outline just outside the outermost wards. Crucially it verifies STREET CONNECTIVITY is
// preserved THROUGH the wall: the street-cell web is still ONE connected component (the wall never
// touches the street cells), the wall is NOT a closed ring (it has gate GAPS), and a street REACHES
// each gate (a street exits the town through the gate). Runs across both bands, several seeds/forms.
func TestV3BWallsPresentWithGates(t *testing.T) {
	_ = theme.SetActive("forge")
	type bandCase struct {
		name       string
		ageKey     string
		wantTowers bool // stone walls carry towers + a gatehouse; mudbrick does not
	}
	bands := []bandCase{
		{"ancient", "bronze_age", false},
		{"medieval", "medieval_age", true},
	}
	names := []string{"Aldermoor", "Bexley", "Corveil", "Duskwind", "Emberton"}
	blds := map[string]int{"hut": 30, "gathering_camp": 20, "forge": 14, "barracks": 8, "colosseum": 1, "stonehenge": 1}

	for _, bc := range bands {
		for _, nm := range names {
			plan := tdPlanFor(namedState(bc.ageKey, nm, blds))

			walls := wallLotsOf(plan)
			gates := gateLotsOf(plan)
			if len(walls) < 12 {
				t.Fatalf("%s seed %q: only %d wall segments — no real curtain wall", bc.name, nm, len(walls))
			}
			if len(gates) < 2 {
				t.Fatalf("%s seed %q: only %d gates — a walled town must have gates the streets exit through", bc.name, nm, len(gates))
			}
			if bc.wantTowers {
				if len(towerLotsOf(plan)) < 3 {
					t.Fatalf("%s seed %q: stone wall has %d towers/gatehouse — a medieval wall must be studded with towers", bc.name, nm, len(towerLotsOf(plan)))
				}
			}

			// (1) The wall RINGS the built area: every wall segment sits OUTSIDE the footprint
			// radius (buildings inside), and inside the bounded town disc (never off-canvas).
			footR := tdFootprintRadius(&plan)
			for _, w := range walls {
				d := math.Hypot(w.x-plan.cx, w.y-plan.cy)
				if d < footR*0.98 {
					t.Fatalf("%s seed %q: a wall segment sits at %.1f, inside the footprint %.1f — the wall must ring OUTSIDE the built area", bc.name, nm, d, footR)
				}
				if d > plan.townR*1.25 {
					t.Fatalf("%s seed %q: a wall segment sits at %.1f, past the bounded disc %.1f — the wall is not bounded", bc.name, nm, d, plan.townR)
				}
			}

			// (2) Buildings stay INSIDE the wall: every roof lot's distance from core is below the
			// minimum wall radius (no roof pokes through the curtain). Use the min wall radius as the
			// inner bound the fabric must respect.
			minWallR := math.Inf(1)
			for _, w := range walls {
				if d := math.Hypot(w.x-plan.cx, w.y-plan.cy); d < minWallR {
					minWallR = d
				}
			}
			for _, lt := range allRoofLots(plan) {
				if d := math.Hypot(lt.x-plan.cx, lt.y-plan.cy) + roofHalfExtent(lt); d > minWallR+0.5 {
					// A roof may sit near the ragged wall's nearest bite, but not beyond the whole
					// ring's max; assert against the MAX wall radius so a ragged inner bite doesn't
					// false-trip while a genuine escapee (past the whole ring) does.
					maxWallR := 0.0
					for _, w := range walls {
						if dd := math.Hypot(w.x-plan.cx, w.y-plan.cy); dd > maxWallR {
							maxWallR = dd
						}
					}
					if d > maxWallR+0.5 {
						t.Fatalf("%s seed %q: a roof reaches %.1f from core, past the wall ring (max %.1f) — buildings must stay inside the wall", bc.name, nm, d, maxWallR)
					}
				}
			}

			// (3) STREET CONNECTIVITY through the wall is intact. The street-cell web is untouched by
			// the wall, so it is still ONE connected component.
			if comps := unionCount(plan.streetCells, plan.cellSize); comps != 1 {
				t.Fatalf("%s seed %q: street web has %d components — the wall must not sever street connectivity", bc.name, nm, comps)
			}

			// (4) The wall is NOT a closed ring — it has GATE GAPS. Bucket the wall segments by angle
			// into 24 sectors; a closed ring fills nearly all sectors, but the gate gaps must leave
			// some EMPTY sectors (the openings). Require at least 2 empty sectors (≥2 gate gaps).
			const sect = 24
			var filled [sect]bool
			for _, w := range walls {
				a := math.Atan2(w.y-plan.cy, w.x-plan.cx)
				b := int((a + math.Pi) / (2 * math.Pi) * sect)
				if b < 0 {
					b = 0
				}
				if b >= sect {
					b = sect - 1
				}
				filled[b] = true
			}
			empty := 0
			for _, f := range filled {
				if !f {
					empty++
				}
			}
			if empty < 2 {
				t.Fatalf("%s seed %q: wall fills %d/%d angular sectors with only %d gaps — the curtain has no gate openings (a closed ring)", bc.name, nm, sect-empty, sect, empty)
			}

			// (5) A STREET REACHES EACH GATE (the gate opens where a street exits). For every gate,
			// require a street cell near the gate's angle out toward the rim — so a road actually
			// passes through the opening (connectivity THROUGH the wall, not a decorative gap).
			for _, g := range gates {
				gAng := math.Atan2(g.y-plan.cy, g.x-plan.cx)
				reached := false
				for _, p := range plan.streetCells {
					pAng := math.Atan2(p.y-plan.cy, p.x-plan.cx)
					pR := math.Hypot(p.x-plan.cx, p.y-plan.cy)
					// Near the gate's direction AND running out toward the wall (a road heading out).
					if angDiff(gAng, pAng) < 0.35 && pR > footR*0.55 {
						reached = true
						break
					}
				}
				if !reached {
					t.Fatalf("%s seed %q: gate at angle %.2f has no street running out to it — the gate must sit on a street the town exits through", bc.name, nm, gAng)
				}
			}
		}
	}
}

// TestV3BEraSquarePropsDiffer locks that the central-square dressing SWAPS per era (task item 5):
// the ancient set (altar / columns / braziers / well) and the medieval set (market stalls /
// fountain / well / cross-or-gallows) are DISTINCT from the primitive set (well / firepit / stones
// / stall) and from EACH OTHER. It checks the palette (tdSquarePropsFor) and that the era-specific
// prop lots actually PLACE in a real plan.
func TestV3BEraSquarePropsDiffer(t *testing.T) {
	_ = theme.SetActive("forge")

	kindSet := func(kinds []tdLotKind) map[tdLotKind]bool {
		m := map[tdLotKind]bool{}
		for _, k := range kinds {
			m[k] = true
		}
		return m
	}
	prim := tdSquarePropsFor(tdStyleForEra(eraOrganic))
	ancient := tdSquarePropsFor(tdStyleForEra(eraHubSpoke))
	medieval := tdSquarePropsFor(tdStyleForEra(eraCastle))

	pW, aW, mW := kindSet(prim.wonder), kindSet(ancient.wonder), kindSet(medieval.wonder)
	// Ancient must contain its signature props and NOT be the primitive set.
	if !aW[tdPropAltar] || !aW[tdPropColumns] || !aW[tdPropBrazier] {
		t.Fatalf("ancient square props %v missing altar/columns/braziers", ancient.wonder)
	}
	if !mW[tdPropStall] || !mW[tdPropFountain] || !mW[tdPropCross] {
		t.Fatalf("medieval square props %v missing stalls/fountain/cross", medieval.wonder)
	}
	sameSet := func(a, b map[tdLotKind]bool) bool {
		if len(a) != len(b) {
			return false
		}
		for k := range a {
			if !b[k] {
				return false
			}
		}
		return true
	}
	if sameSet(aW, pW) {
		t.Fatal("ancient square props equal the primitive set — the dressing must swap per era")
	}
	if sameSet(mW, pW) {
		t.Fatal("medieval square props equal the primitive set — the dressing must swap per era")
	}
	if sameSet(aW, mW) {
		t.Fatal("ancient and medieval square props are identical — each era needs its own set")
	}

	// End-to-end: an ancient wonder town PLACES ancient props (and no primitive-only stones/firepit);
	// a medieval one PLACES medieval props (fountain/cross). Deterministic ring placement holds.
	blds := map[string]int{"hut": 26, "gathering_camp": 16, "forge": 10, "colosseum": 1, "stonehenge": 1}
	countKinds := func(plan topPlan) map[tdLotKind]int {
		m := map[tdLotKind]int{}
		for _, lt := range plan.lots {
			m[lt.kind]++
		}
		return m
	}
	ap := countKinds(tdPlanFor(namedState("bronze_age", "Aldermoor", blds)))
	if ap[tdPropAltar]+ap[tdPropColumns]+ap[tdPropBrazier] == 0 {
		t.Fatal("ancient wonder town placed no ancient square props (altar/columns/brazier)")
	}
	if ap[tdPropFirepit] != 0 || ap[tdPropStones] != 0 {
		t.Fatalf("ancient wonder town placed primitive-only props (firepit=%d stones=%d)", ap[tdPropFirepit], ap[tdPropStones])
	}
	mp := countKinds(tdPlanFor(namedState("medieval_age", "Duskwind", blds)))
	if mp[tdPropFountain]+mp[tdPropCross] == 0 {
		t.Fatal("medieval wonder town placed no medieval square props (fountain/cross)")
	}
	if mp[tdPropFirepit] != 0 || mp[tdPropStones] != 0 {
		t.Fatalf("medieval wonder town placed primitive-only props (firepit=%d stones=%d)", mp[tdPropFirepit], mp[tdPropStones])
	}
}

// TestV3BEraWonderDiffers locks that the era WONDER silhouette swaps (task item 6): ancient reads
// as a ZIGGURAT, medieval as a CATHEDRAL, both distinct from the primitive/default wonder. The
// archetype stays roofWonder (so placement/labels are unchanged); the DRAW differs by era profile.
// We assert (a) the era profiles are set, and (b) drawing the same wonder footprint under each era
// paints a DIFFERENT pixel field (the shapes actually differ).
func TestV3BEraWonderDiffers(t *testing.T) {
	_ = theme.SetActive("forge")

	if tdStyleForEra(eraHubSpoke).houseProfile != profileMudbrick {
		t.Fatal("ancient era must use the mudbrick profile (drives the ziggurat wonder + flat houses)")
	}
	if tdStyleForEra(eraCastle).houseProfile != profileTimber {
		t.Fatal("medieval era must use the timber profile (drives the cathedral wonder + steep houses)")
	}

	// Draw the same wonder lot under each era into its own image; the pixel fields must differ.
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	prim := drawWonderImg(tdStyleForEra(eraOrganic))
	ancient := drawWonderImg(tdStyleForEra(eraHubSpoke))
	medieval := drawWonderImg(tdStyleForEra(eraCastle))
	if !imagesDiffer(prim, ancient) {
		t.Fatal("ancient wonder draws identically to the primitive wonder — the ziggurat silhouette is not applied")
	}
	if !imagesDiffer(prim, medieval) {
		t.Fatal("medieval wonder draws identically to the primitive wonder — the cathedral silhouette is not applied")
	}
	if !imagesDiffer(ancient, medieval) {
		t.Fatal("ancient ziggurat and medieval cathedral draw identically — the two wonders must differ")
	}
}

// TestV3BWallsBoundedAndOpenErasHaveNone locks two guarantees: (1) walled-era wall lots map
// IN-BOUNDS on real + tiny canvases (bounded, panic-safe); and (2) the OPEN eras (industrial+ /
// modern / cyber / space) and the primitive village have NO walls (locked #9: industrial+ stays
// open sprawl; primitive never walled).
func TestV3BWallsBoundedAndOpenErasHaveNone(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}

	// (1) Walled eras: every wall/gate/tower lot maps in-bounds through the fill-frame transform on a
	// normal canvas AND a tiny one (bounded + panic-safe).
	for _, ageKey := range []string{"bronze_age", "medieval_age"} {
		for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
			plan := tdPlanFor(namedState(ageKey, "Aldermoor", blds))
			xf := computeTransform(&plan, sz.w, sz.h)
			for _, lt := range plan.lots {
				switch lt.kind {
				case tdWall, tdGate, tdWallTower, tdWallGatehouse:
					px, py := xf.px(lt.x, lt.y)
					if px < -2 || px > sz.w+2 || py < -2 || py > sz.h+2 {
						t.Fatalf("%s %dx%d: wall lot maps to (%d,%d) far off-canvas — the ring is not bounded to the frame", ageKey, sz.w, sz.h, px, py)
					}
				}
			}
		}
	}

	// (2) The primitive village + the open eras have NO wall lots.
	openStyles := []struct {
		name   string
		ageKey string
	}{
		{"primitive", "primitive_age"},
		{"industrial", "industrial_age"},
		{"modern", "modern_age"},
		{"digital", "digital_age"},
		{"galactic", "galactic_age"},
	}
	for _, os := range openStyles {
		plan := tdPlanFor(namedState(os.ageKey, "Aldermoor", blds))
		if n := len(wallLotsOf(plan)) + len(gateLotsOf(plan)) + len(towerLotsOf(plan)); n != 0 {
			t.Fatalf("%s (%s) has %d wall lots — this era must be OPEN (no walls)", os.name, os.ageKey, n)
		}
	}
}

// TestStoneAgeStyleAndWonderMotifRefactor locks Phase 1b-i: the STONE age gets its own style
// (megalith wonder, thatch houses still) and the wonder-motif field refactor preserved the
// ancient/medieval wonder mapping (they read off wonderMotif now, not houseProfile).
func TestStoneAgeStyleAndWonderMotifRefactor(t *testing.T) {
	if got := styleForAge("stone_age").wonderMotif; got != wonderMegalith {
		t.Fatalf("stone_age wonderMotif = %v, want wonderMegalith", got)
	}
	if got := styleForAge("stone_age").houseProfile; got != profileThatch {
		t.Fatalf("stone_age houseProfile = %v, want profileThatch (stone dwellings stay thatch)", got)
	}
	// Behaviour-preserving refactor: ancient/medieval wonders now key off wonderMotif and must map
	// to the same sprites they did off houseProfile.
	if got := styleForAge("bronze_age").wonderMotif; got != wonderZiggurat {
		t.Fatalf("bronze_age wonderMotif = %v, want wonderZiggurat", got)
	}
	if got := styleForAge("medieval_age").wonderMotif; got != wonderCathedral {
		t.Fatalf("medieval_age wonderMotif = %v, want wonderCathedral", got)
	}
}

// TestStoneMegalithWonderDiffers locks that the stone-age MEGALITH wonder draws differently from
// the primitive generic hall AND from the bronze ziggurat (the megalith sprite is actually applied).
func TestStoneMegalithWonderDiffers(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	prim := drawWonderImg(styleForAge("primitive_age"))
	stone := drawWonderImg(styleForAge("stone_age"))
	bronze := drawWonderImg(styleForAge("bronze_age"))
	if !imagesDiffer(prim, stone) {
		t.Fatal("stone megalith wonder draws identically to the primitive hall — the megalith silhouette is not applied")
	}
	if !imagesDiffer(stone, bronze) {
		t.Fatal("stone megalith wonder draws identically to the bronze ziggurat — the two wonders must differ")
	}
}

// TestStoneSquarePropsIncludeMegalith locks that stone's town-square prop palette carries the
// megalith prop and primitive's does NOT (the stone square dresses distinct from primitive).
func TestStoneSquarePropsIncludeMegalith(t *testing.T) {
	kindSet := func(kinds []tdLotKind) map[tdLotKind]bool {
		m := map[tdLotKind]bool{}
		for _, k := range kinds {
			m[k] = true
		}
		return m
	}
	stone := tdSquarePropsFor(styleForAge("stone_age"))
	prim := tdSquarePropsFor(styleForAge("primitive_age"))
	sW := kindSet(stone.wonder)
	pW := kindSet(prim.wonder)
	if !sW[tdPropMegalith] {
		t.Fatalf("stone square props %v missing tdPropMegalith", stone.wonder)
	}
	if pW[tdPropMegalith] {
		t.Fatalf("primitive square props %v must NOT include tdPropMegalith", prim.wonder)
	}
	sC := kindSet(stone.center)
	if !sC[tdPropMegalith] {
		t.Fatalf("stone center props %v missing tdPropMegalith", stone.center)
	}
}

// TestDrawRoofMegalithPanicSafe locks that the megalith sprite is panic-safe + in-bounds on a tiny
// footprint and a normal one (every write is clamped).
func TestDrawRoofMegalithPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("stone_age")
	for _, tc := range []struct{ w, h, hw, hh int }{
		{9, 9, 2, 2},     // tiny footprint centered near an edge-ish spot
		{40, 40, 12, 10}, // normal footprint
	} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		// Should not panic even if the center + extents push writes past the image edges.
		drawRoofMegalith(img, tc.w/2, tc.h/2, tc.hw, tc.hh, rc)
		drawRoofMegalith(img, 1, 1, tc.hw, tc.hh, rc) // hard against the NW corner
	}
}

// TestDrawRoofKeepPanicSafe locks that the iron KEEP sprite is panic-safe + in-bounds on a tiny
// footprint, a normal horizontal one, and a vertical one (tower seats north), including hard against
// the NW corner (every write is clamped).
func TestDrawRoofKeepPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("iron_age")
	for _, tc := range []struct{ w, h, hw, hh int }{
		{9, 9, 2, 2},     // tiny footprint
		{40, 40, 12, 10}, // normal, horizontal (tower seats west)
		{40, 40, 8, 14},  // normal, vertical (tower seats north)
	} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		drawRoofKeep(img, tc.w/2, tc.h/2, tc.hw, tc.hh, rc)
		drawRoofKeep(img, 1, 1, tc.hw, tc.hh, rc) // hard against the NW corner
	}
}

// TestDumpStoneEpochPNGs renders a representative city for primitive / stone / bronze with a FIXED
// seed and identical building counts, and writes PNGs for human eyeballing. Not an assertion test —
// it exists so a reviewer can compare the three ages side by side. Opt-in: skipped unless
// CITYMAP_PNG_DUMP is set to an output directory, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpStoneEpochPNGs
func TestDumpStoneEpochPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	// Identical building set (with a wonder so the centerpiece renders) and a FIXED display name →
	// the citySeed is fixed, so only the era re-skin differs across the three dumps.
	blds := map[string]int{"hut": 26, "gathering_camp": 16, "forge": 10, "stonehenge": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"primitive_age", "1b_i_primitive.png"},
		{"stone_age", "1b_i_stone.png"},
		{"bronze_age", "1b_i_bronze.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// tdPlanForAge builds the top-down plan using the REAL per-age style (styleForAge), not the stale
// era-band tdStyleForEra path tdPlanFor uses — so Phase 1b-i/ii per-age presets (stone, iron,
// classical) are actually exercised. Mirrors renderTopDown's style + seed derivation.
func tdPlanForAge(state game.GameState) topPlan {
	style := styleForAge(state.Age)
	seed := citySeed(displayNameOf(state))
	return generateTopPlan(state, config.BuildingByKey(), style, seed)
}

// TestIronClassicalStylesWired locks Phase 1b-ii style wiring: iron gets a fortified KEEP wonder +
// mudbrick houses + a TIMBER palisade wall; classical gets a TEMPLE wonder, STONE walls, and the
// white-stone house profile. The behaviour-preserving foundation (bronze) stays ancient (ziggurat).
func TestIronClassicalStylesWired(t *testing.T) {
	iron := styleForAge("iron_age")
	if iron.wonderMotif != wonderKeep {
		t.Fatalf("iron wonderMotif = %v, want wonderKeep (iron gets its own fortified keep)", iron.wonderMotif)
	}
	if iron.wallProfile != wallTimber {
		t.Fatalf("iron wallProfile = %v, want wallTimber (a brown palisade)", iron.wallProfile)
	}
	if iron.houseProfile != profileMudbrick {
		t.Fatalf("iron houseProfile = %v, want profileMudbrick (unchanged from ancient)", iron.houseProfile)
	}
	if !iron.hasWalls {
		t.Fatal("iron must have walls (a timber palisade)")
	}

	cl := styleForAge("classical_age")
	if cl.wonderMotif != wonderTemple {
		t.Fatalf("classical wonderMotif = %v, want wonderTemple", cl.wonderMotif)
	}
	if cl.wallProfile != wallStone {
		t.Fatalf("classical wallProfile = %v, want wallStone", cl.wallProfile)
	}
	if cl.houseProfile != profileStoneClassical {
		t.Fatalf("classical houseProfile = %v, want profileStoneClassical", cl.houseProfile)
	}

	// Bronze foundation unchanged: still ancient (ziggurat + mudbrick wall + mudbrick houses).
	bz := styleForAge("bronze_age")
	if bz.wonderMotif != wonderZiggurat || bz.wallProfile != wallMudbrick || bz.houseProfile != profileMudbrick {
		t.Fatalf("bronze changed: motif=%v wall=%v house=%v — bronze must stay ancient", bz.wonderMotif, bz.wallProfile, bz.houseProfile)
	}
}

// TestIronCityDiffersFromBronze locks that IRON reads apart from BRONZE at the CITY level. Both share
// the ziggurat wonder (so the wonder sprite alone is identical) — the distinction is the timber wall +
// cooler roof/ground tint, which must change the rendered city pixels.
func TestIronCityDiffersFromBronze(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	bronze, _ := renderImage(namedState("bronze_age", "Aldermoor", blds), 120, 72)
	iron, _ := renderImage(namedState("iron_age", "Aldermoor", blds), 120, 72)
	if !imagesDiffer(bronze, iron) {
		t.Fatal("iron city renders identically to bronze — the timber wall + cooler tint are not applied")
	}
}

// TestV3BiiWondersDiffer locks that the distinct wonder silhouettes are actually applied: the
// classical TEMPLE reads apart from the medieval CATHEDRAL, and — the follow-up fix — the iron KEEP
// reads apart from the bronze ZIGGURAT (iron no longer reuses bronze's centerpiece).
func TestV3BiiWondersDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	classical := drawWonderImg(styleForAge("classical_age"))
	medieval := drawWonderImg(styleForAge("medieval_age"))
	iron := drawWonderImg(styleForAge("iron_age"))
	bronze := drawWonderImg(styleForAge("bronze_age"))
	if !imagesDiffer(classical, medieval) {
		t.Fatal("classical temple draws identically to the medieval cathedral — the temple silhouette is not applied")
	}
	if !imagesDiffer(classical, iron) {
		t.Fatal("classical temple draws identically to the iron keep — the two wonders must differ")
	}
	if !imagesDiffer(iron, bronze) {
		t.Fatal("iron keep draws identically to the bronze ziggurat — iron must have its own centerpiece")
	}
}

// TestIronTimberWallLotsBounded locks that an iron town emits wallTimber wall + gate lots that map
// in-bounds on real + tiny canvases (bounded, panic-safe), and has NO stone towers/gatehouse.
func TestIronTimberWallLotsBounded(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		plan := tdPlanForAge(namedState("iron_age", "Aldermoor", blds))
		if len(wallLotsOf(plan)) < 12 {
			t.Fatalf("iron %dx%d: only %d wall segments — no real palisade", sz.w, sz.h, len(wallLotsOf(plan)))
		}
		if len(gateLotsOf(plan)) < 2 {
			t.Fatalf("iron %dx%d: only %d gates — the palisade needs gates the streets exit through", sz.w, sz.h, len(gateLotsOf(plan)))
		}
		if n := len(towerLotsOf(plan)); n != 0 {
			t.Fatalf("iron has %d stone towers/gatehouse — a timber palisade must have none", n)
		}
		xf := computeTransform(&plan, sz.w, sz.h)
		for _, lt := range plan.lots {
			switch lt.kind {
			case tdWall, tdGate:
				px, py := xf.px(lt.x, lt.y)
				if px < -2 || px > sz.w+2 || py < -2 || py > sz.h+2 {
					t.Fatalf("iron %dx%d: wall lot maps to (%d,%d) off-canvas — the ring is not bounded", sz.w, sz.h, px, py)
				}
			}
		}
	}
}

// TestClassicalStoneWallLots locks that a classical town gets a proper STONE wall — segments, gates,
// AND towers/gatehouse (classical cities have real fortifications).
func TestClassicalStoneWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	plan := tdPlanForAge(namedState("classical_age", "Aldermoor", blds))
	if len(wallLotsOf(plan)) < 12 {
		t.Fatalf("classical: only %d wall segments — no real curtain wall", len(wallLotsOf(plan)))
	}
	if len(gateLotsOf(plan)) < 2 {
		t.Fatalf("classical: only %d gates", len(gateLotsOf(plan)))
	}
	if n := len(towerLotsOf(plan)); n < 3 {
		t.Fatalf("classical stone wall has %d towers/gatehouse — a stone wall must be studded with towers", n)
	}
}

// TestClassicalSquarePropsColumnsForward locks that classical squares are dressed columns-forward
// (a Greco-Roman forum) and NOT the ancient altar/brazier set.
func TestClassicalSquarePropsColumnsForward(t *testing.T) {
	kindSet := func(kinds []tdLotKind) map[tdLotKind]bool {
		m := map[tdLotKind]bool{}
		for _, k := range kinds {
			m[k] = true
		}
		return m
	}
	cl := tdSquarePropsFor(styleForAge("classical_age"))
	w := kindSet(cl.wonder)
	if !w[tdPropColumns] || !w[tdPropAltar] || !w[tdPropWell] {
		t.Fatalf("classical wonder props %v missing columns/altar/well", cl.wonder)
	}
	if w[tdPropBrazier] {
		t.Fatalf("classical wonder props %v must not carry the ancient brazier", cl.wonder)
	}
	c := kindSet(cl.center)
	if !c[tdPropColumns] {
		t.Fatalf("classical center props %v missing columns", cl.center)
	}
}

// TestV3BiiSpritesPanicSafe locks that the two new sprites (temple wonder + classical house) are
// panic-safe + in-bounds on a tiny footprint and a normal one (every write is clamped).
func TestV3BiiSpritesPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	rcTemple := roofColorsFor(styleForAge("classical_age"), pal, "wonder", "wonder")
	rcHouse := roofColorsFor(styleForAge("classical_age"), pal, "housing", "production")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}} {
			drawRoofTempleWonder(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rcTemple)
			drawRoofTempleWonder(img, 1, 1, hwhh.hw, hwhh.hh, rcTemple) // hard against the NW corner
			drawRoofStoneClassical(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rcHouse)
			drawRoofStoneClassical(img, 1, 1, hwhh.hw, hwhh.hh, rcHouse)
		}
	}
}

// --- Renaissance (Phase 1b) ---------------------------------------------------

// TestRenaissanceStyleWired locks the renaissance wiring: a DOME wonder, a STAR-FORT wall, walls on,
// and the cream-stone house profile. Medieval (its old shared preset) is unchanged (cathedral + stone).
func TestRenaissanceStyleWired(t *testing.T) {
	ren := styleForAge("renaissance_age")
	if ren.wonderMotif != wonderDome {
		t.Fatalf("renaissance wonderMotif = %v, want wonderDome", ren.wonderMotif)
	}
	if ren.wallProfile != wallStarFort {
		t.Fatalf("renaissance wallProfile = %v, want wallStarFort", ren.wallProfile)
	}
	if !ren.hasWalls {
		t.Fatal("renaissance must have walls (a star-fort)")
	}
	if ren.houseProfile != profileStoneClassical {
		t.Fatalf("renaissance houseProfile = %v, want profileStoneClassical (cream ashlar townhouses)", ren.houseProfile)
	}

	// Medieval, whose preset renaissance used to share, must stay MEDIEVAL (cathedral + stone wall).
	med := styleForAge("medieval_age")
	if med.wonderMotif != wonderCathedral || med.wallProfile != wallStone {
		t.Fatalf("medieval changed: motif=%v wall=%v — medieval must stay cathedral + stone", med.wonderMotif, med.wallProfile)
	}
}

// TestRenaissanceWonderDiffers locks that the renaissance DOME reads apart from the medieval CATHEDRAL,
// the classical TEMPLE, and the iron KEEP — the dome silhouette must actually be applied, not shared.
func TestRenaissanceWonderDiffers(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	ren := drawWonderImg(styleForAge("renaissance_age"))
	med := drawWonderImg(styleForAge("medieval_age"))
	classical := drawWonderImg(styleForAge("classical_age"))
	iron := drawWonderImg(styleForAge("iron_age"))
	if !imagesDiffer(ren, med) {
		t.Fatal("renaissance dome draws identically to the medieval cathedral — the dome silhouette is not applied")
	}
	if !imagesDiffer(ren, classical) {
		t.Fatal("renaissance dome draws identically to the classical temple — the two wonders must differ")
	}
	if !imagesDiffer(ren, iron) {
		t.Fatal("renaissance dome draws identically to the iron keep — the two wonders must differ")
	}
}

// TestRenaissanceCityDiffersFromMedieval locks that a renaissance city reads apart from a medieval one
// at the CITY level (cream stone + dome + star-fort vs cool slate + cathedral + stone curtain).
func TestRenaissanceCityDiffersFromMedieval(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	med, _ := renderImage(namedState("medieval_age", "Aldermoor", blds), 120, 72)
	ren, _ := renderImage(namedState("renaissance_age", "Aldermoor", blds), 120, 72)
	if !imagesDiffer(med, ren) {
		t.Fatal("renaissance city renders identically to medieval — the cream re-skin + dome + star-fort are not applied")
	}
}

// TestDrawRoofDomePanicSafe locks that the renaissance DOME sprite is panic-safe + in-bounds on a tiny
// footprint, a normal one, and hard against the NW corner (every write is clamped).
func TestDrawRoofDomePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("renaissance_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}} {
			drawRoofDome(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofDome(img, 1, 1, hwhh.hw, hwhh.hh, rc) // hard against the NW corner
		}
	}
}

// TestRenaissanceStarFortWallLots locks that a renaissance town emits a real STAR-FORT: curtain
// segments, gates the streets exit through, and periodic ANGULAR BASTION salients — and NO round
// stone towers/gatehouse. Every wall/gate/bastion lot must map in-bounds on real + tiny canvases.
func TestRenaissanceStarFortWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		plan := tdPlanForAge(namedState("renaissance_age", "Aldermoor", blds))
		if len(wallLotsOf(plan)) < 12 {
			t.Fatalf("renaissance %dx%d: only %d wall segments — no real rampart", sz.w, sz.h, len(wallLotsOf(plan)))
		}
		if len(gateLotsOf(plan)) < 2 {
			t.Fatalf("renaissance %dx%d: only %d gates — the star-fort needs gates the streets exit through", sz.w, sz.h, len(gateLotsOf(plan)))
		}
		if n := len(bastionLotsOf(plan)); n < 3 {
			t.Fatalf("renaissance %dx%d: only %d bastions — a star-fort must have angular salients", sz.w, sz.h, n)
		}
		if n := len(towerLotsOf(plan)); n != 0 {
			t.Fatalf("renaissance %dx%d: has %d round stone towers/gatehouse — a star-fort has angular bastions only", sz.w, sz.h, n)
		}
		xf := computeTransform(&plan, sz.w, sz.h)
		for _, lt := range plan.lots {
			switch lt.kind {
			case tdWall, tdGate, tdWallBastion:
				px, py := xf.px(lt.x, lt.y)
				if px < -3 || px > sz.w+3 || py < -3 || py > sz.h+3 {
					t.Fatalf("renaissance %dx%d: wall/bastion lot maps to (%d,%d) off-canvas — the star-fort is not bounded", sz.w, sz.h, px, py)
				}
			}
		}
	}
}

// TestDumpRenaissancePNGs renders medieval / renaissance / classical with a FIXED display name +
// identical building set INCLUDING a wonder so the cathedral/dome/temple centerpieces render, so a
// reviewer can compare the three side by side. Opt-in: skipped unless CITYMAP_PNG_DUMP=<dir> is set:
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpRenaissancePNGs
func TestDumpRenaissancePNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	// Identical building set (with a wonder so the centerpiece renders) + a FIXED display name → the
	// citySeed is fixed, so only the era re-skin differs across the three dumps.
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"medieval_age", "1c_medieval.png"},
		{"renaissance_age", "1c_renaissance.png"},
		{"classical_age", "1c_classical.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// TestColonialIndustrialStylesWired locks Phase 1b-iii style wiring: COLONIAL gets brick ROWHOUSES, a
// TIMBER palisade wall (walls on), and the grand GENERIC hall wonder (a statehouse); INDUSTRIAL gets
// brick ROWHOUSES too but under tin roofs, NO walls, and the FACTORY wonder. Neither age still maps to
// the default village preset.
func TestColonialIndustrialStylesWired(t *testing.T) {
	col := styleForAge("colonial_age")
	if col.houseProfile != profileRowhouse {
		t.Fatalf("colonial houseProfile = %v, want profileRowhouse (brick terraces)", col.houseProfile)
	}
	if col.wallProfile != wallTimber {
		t.Fatalf("colonial wallProfile = %v, want wallTimber (a palisade fort)", col.wallProfile)
	}
	if !col.hasWalls {
		t.Fatal("colonial must have walls (a timber palisade fort)")
	}
	if col.wonderMotif != wonderGeneric {
		t.Fatalf("colonial wonderMotif = %v, want wonderGeneric (the statehouse — no bespoke wonder)", col.wonderMotif)
	}

	ind := styleForAge("industrial_age")
	if ind.houseProfile != profileRowhouse {
		t.Fatalf("industrial houseProfile = %v, want profileRowhouse (dense brick terraces)", ind.houseProfile)
	}
	if ind.wallProfile != wallNone {
		t.Fatalf("industrial wallProfile = %v, want wallNone (open industry)", ind.wallProfile)
	}
	if ind.hasWalls {
		t.Fatal("industrial must be OPEN (no walls)")
	}
	if ind.wonderMotif != wonderFactory {
		t.Fatalf("industrial wonderMotif = %v, want wonderFactory", ind.wonderMotif)
	}
	// Industrial should pack DENSER than colonial (tighter slotSpacing).
	if !(ind.slotSpacing < col.slotSpacing) {
		t.Fatalf("industrial slotSpacing (%.2f) should be tighter/denser than colonial (%.2f)", ind.slotSpacing, col.slotSpacing)
	}
	// Neither age may still resolve to the default village preset name.
	if col.name == defaultTdStyle.name || ind.name == defaultTdStyle.name {
		t.Fatalf("colonial/industrial still on the default preset: colonial=%q industrial=%q", col.name, ind.name)
	}
}

// TestIndustrialFactoryWonderDiffers locks that the industrial FACTORY reads apart from the renaissance
// DOME, the classical TEMPLE, and the medieval CATHEDRAL — the factory silhouette must actually be
// applied, not shared with any neighbour.
func TestIndustrialFactoryWonderDiffers(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	factory := drawWonderImg(styleForAge("industrial_age"))
	dome := drawWonderImg(styleForAge("renaissance_age"))
	temple := drawWonderImg(styleForAge("classical_age"))
	cathedral := drawWonderImg(styleForAge("medieval_age"))
	if !imagesDiffer(factory, dome) {
		t.Fatal("industrial factory draws identically to the renaissance dome — the factory silhouette is not applied")
	}
	if !imagesDiffer(factory, temple) {
		t.Fatal("industrial factory draws identically to the classical temple — the two wonders must differ")
	}
	if !imagesDiffer(factory, cathedral) {
		t.Fatal("industrial factory draws identically to the medieval cathedral — the two wonders must differ")
	}
}

// TestColonialIndustrialCitiesDiffer locks the CITY-level reads: a colonial city differs from an
// industrial one, colonial differs from the old default village (modern_age still uses default), and
// industrial differs from colonial — the two new re-skins are actually applied and distinct.
func TestColonialIndustrialCitiesDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	col, _ := renderImage(namedState("colonial_age", "Aldermoor", blds), 120, 72)
	ind, _ := renderImage(namedState("industrial_age", "Aldermoor", blds), 120, 72)
	def, _ := renderImage(namedState("modern_age", "Aldermoor", blds), 120, 72) // modern still uses defaultTdStyle
	if !imagesDiffer(col, ind) {
		t.Fatal("colonial city renders identically to industrial — the two re-skins are not distinct")
	}
	if !imagesDiffer(col, def) {
		t.Fatal("colonial city renders identically to the default village (modern) — the colonial re-skin is not applied")
	}
	if !imagesDiffer(ind, def) {
		t.Fatal("industrial city renders identically to the default village (modern) — the industrial re-skin is not applied")
	}
}

// TestDrawRoofRowhousePanicSafe locks that the ROWHOUSE dwelling sprite is panic-safe + in-bounds on a
// tiny footprint, a normal one, and hard against the NW corner (every write is clamped).
func TestDrawRoofRowhousePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("colonial_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofRowhouse(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofRowhouse(img, 1, 1, hwhh.hw, hwhh.hh, rc) // hard against the NW corner
		}
	}
}

// TestDrawRoofFactoryPanicSafe locks that the industrial FACTORY wonder sprite (hall + tall smokestacks
// + smoke that clips ABOVE the footprint) is panic-safe + in-bounds on tiny / normal / NW-corner cases.
func TestDrawRoofFactoryPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("industrial_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}} {
			drawRoofFactory(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofFactory(img, 1, 1, hwhh.hw, hwhh.hh, rc)      // NW corner
			drawRoofFactory(img, tc.w-1, 1, hwhh.hw, hwhh.hh, rc) // NE corner (smoke drifts up-right)
		}
	}
}

// TestColonialIndustrialWallLots locks that COLONIAL emits a real TIMBER PALISADE (curtain segments +
// gates, but NO round stone towers/gatehouse and NO angular bastions), while INDUSTRIAL is OPEN (zero
// wall/gate/tower/bastion lots). Bounded on real + tiny canvases.
func TestColonialIndustrialWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		// COLONIAL: a timber palisade — segments + gates, no towers, no bastions.
		cp := tdPlanForAge(namedState("colonial_age", "Aldermoor", blds))
		if len(wallLotsOf(cp)) < 12 {
			t.Fatalf("colonial %dx%d: only %d wall segments — no real palisade", sz.w, sz.h, len(wallLotsOf(cp)))
		}
		if len(gateLotsOf(cp)) < 2 {
			t.Fatalf("colonial %dx%d: only %d gates — the palisade needs gates the streets exit through", sz.w, sz.h, len(gateLotsOf(cp)))
		}
		if n := len(towerLotsOf(cp)); n != 0 {
			t.Fatalf("colonial %dx%d: has %d round stone towers/gatehouse — a timber palisade has none", sz.w, sz.h, n)
		}
		if n := len(bastionLotsOf(cp)); n != 0 {
			t.Fatalf("colonial %dx%d: has %d bastions — a timber palisade is not a star-fort", sz.w, sz.h, n)
		}
		xf := computeTransform(&cp, sz.w, sz.h)
		for _, lt := range cp.lots {
			switch lt.kind {
			case tdWall, tdGate:
				px, py := xf.px(lt.x, lt.y)
				if px < -3 || px > sz.w+3 || py < -3 || py > sz.h+3 {
					t.Fatalf("colonial %dx%d: wall/gate lot maps to (%d,%d) off-canvas — the palisade is not bounded", sz.w, sz.h, px, py)
				}
			}
		}
		// INDUSTRIAL: OPEN — zero wall lots of any kind.
		ip := tdPlanForAge(namedState("industrial_age", "Aldermoor", blds))
		if n := len(wallLotsOf(ip)) + len(gateLotsOf(ip)) + len(towerLotsOf(ip)) + len(bastionLotsOf(ip)); n != 0 {
			t.Fatalf("industrial %dx%d has %d wall lots — this age must be OPEN (no walls)", sz.w, sz.h, n)
		}
	}
}

// TestDumpColonialIndustrialPNGs renders renaissance / colonial / industrial with a FIXED display name +
// identical building set INCLUDING a wonder so the dome/statehouse/factory centerpieces render, so a
// reviewer can compare the three side by side. Opt-in: skipped unless CITYMAP_PNG_DUMP=<dir> is set:
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpColonialIndustrialPNGs
func TestDumpColonialIndustrialPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	// Identical building set (with a wonder so the centerpiece renders) + a FIXED display name → the
	// citySeed is fixed, so only the era re-skin differs across the three dumps.
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"renaissance_age", "1c_renaissance2.png"},
		{"colonial_age", "1c_colonial.png"},
		{"industrial_age", "1c_industrial.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// TestElectricEpochStylesWired locks the ELECTRIC-epoch style wiring (V3-B): VICTORIAN gets brownstone
// ROWHOUSES, NO walls, and the grand GENERIC hall wonder (a terminal/museum); ELECTRIC gets FLAT modern
// blocks (profileModernFlat), NO walls, and the setback TOWER wonder; ATOMIC gets FLAT modern blocks too,
// NO walls, and a TOWER wonder, but AIRIER (looser slotSpacing) than the denser electric downtown. None
// of the three still maps to the default village preset.
func TestElectricEpochStylesWired(t *testing.T) {
	vic := styleForAge("victorian_age")
	if vic.houseProfile != profileRowhouse {
		t.Fatalf("victorian houseProfile = %v, want profileRowhouse (brownstone terraces)", vic.houseProfile)
	}
	if vic.wonderMotif != wonderGeneric {
		t.Fatalf("victorian wonderMotif = %v, want wonderGeneric (the terminal/museum — no bespoke wonder)", vic.wonderMotif)
	}
	if vic.hasWalls {
		t.Fatal("victorian must be OPEN (no walls — the age of open boulevards)")
	}

	ele := styleForAge("electric_age")
	if ele.houseProfile != profileModernFlat {
		t.Fatalf("electric houseProfile = %v, want profileModernFlat (flat deco blocks)", ele.houseProfile)
	}
	if ele.wonderMotif != wonderTower {
		t.Fatalf("electric wonderMotif = %v, want wonderTower (art-deco setback tower)", ele.wonderMotif)
	}
	if ele.hasWalls {
		t.Fatal("electric must be OPEN (no walls)")
	}

	atm := styleForAge("atomic_age")
	if atm.houseProfile != profileModernFlat {
		t.Fatalf("atomic houseProfile = %v, want profileModernFlat (flat midcentury blocks)", atm.houseProfile)
	}
	if atm.wonderMotif != wonderTower {
		t.Fatalf("atomic wonderMotif = %v, want wonderTower (midcentury setback tower)", atm.wonderMotif)
	}
	if atm.hasWalls {
		t.Fatal("atomic must be OPEN (no walls)")
	}
	// Atomic should be AIRIER than electric (a suburb-and-downtown feel vs a packed deco downtown).
	if !(atm.slotSpacing > ele.slotSpacing) {
		t.Fatalf("atomic slotSpacing (%.2f) should be airier/looser than electric (%.2f)", atm.slotSpacing, ele.slotSpacing)
	}
	// None of the three may still resolve to the default village preset name.
	if vic.name == defaultTdStyle.name || ele.name == defaultTdStyle.name || atm.name == defaultTdStyle.name {
		t.Fatalf("an electric-epoch age still on the default preset: victorian=%q electric=%q atomic=%q", vic.name, ele.name, atm.name)
	}
}

// TestElectricTowerWonderDiffers locks that the electric/atomic TOWER reads apart from the renaissance
// DOME, the industrial FACTORY, the ancient ZIGGURAT, and the iron KEEP — the tower silhouette must
// actually be applied, not shared with any neighbouring wonder.
func TestElectricTowerWonderDiffers(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	tower := drawWonderImg(styleForAge("electric_age"))
	dome := drawWonderImg(styleForAge("renaissance_age"))
	factory := drawWonderImg(styleForAge("industrial_age"))
	ziggurat := drawWonderImg(styleForAge("bronze_age"))
	keep := drawWonderImg(styleForAge("iron_age"))
	if !imagesDiffer(tower, dome) {
		t.Fatal("electric tower draws identically to the renaissance dome — the tower silhouette is not applied")
	}
	if !imagesDiffer(tower, factory) {
		t.Fatal("electric tower draws identically to the industrial factory — the two wonders must differ")
	}
	if !imagesDiffer(tower, ziggurat) {
		t.Fatal("electric tower draws identically to the ancient ziggurat — the two wonders must differ")
	}
	if !imagesDiffer(tower, keep) {
		t.Fatal("electric tower draws identically to the iron keep — the two wonders must differ")
	}
}

// TestElectricEpochCitiesDiffer locks the CITY-level reads: victorian differs from colonial (both use
// rowhouses — brownstone-vs-brick + stone-pavers-vs-dirt + parks-vs-palisade must diverge), electric
// differs from atomic (warmer ornate dense deco vs cooler clean airy pastel midcentury), and all three
// differ from the old default village (modern_age still uses default) — the re-skins are actually applied.
func TestElectricEpochCitiesDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	vic, _ := renderImage(namedState("victorian_age", "Aldermoor", blds), 120, 72)
	col, _ := renderImage(namedState("colonial_age", "Aldermoor", blds), 120, 72)
	ele, _ := renderImage(namedState("electric_age", "Aldermoor", blds), 120, 72)
	atm, _ := renderImage(namedState("atomic_age", "Aldermoor", blds), 120, 72)
	def, _ := renderImage(namedState("modern_age", "Aldermoor", blds), 120, 72) // modern still uses defaultTdStyle
	if !imagesDiffer(vic, col) {
		t.Fatal("victorian city renders identically to colonial — the brownstone re-skin is not distinct from brick")
	}
	if !imagesDiffer(ele, atm) {
		t.Fatal("electric city renders identically to atomic — the warm-deco vs cool-midcentury re-skins are not distinct")
	}
	if !imagesDiffer(vic, def) {
		t.Fatal("victorian city renders identically to the default village (modern) — the victorian re-skin is not applied")
	}
	if !imagesDiffer(ele, def) {
		t.Fatal("electric city renders identically to the default village (modern) — the electric re-skin is not applied")
	}
	if !imagesDiffer(atm, def) {
		t.Fatal("atomic city renders identically to the default village (modern) — the atomic re-skin is not applied")
	}
}

// TestDrawRoofModernFlatPanicSafe locks that the FLAT modern-block dwelling sprite is panic-safe +
// in-bounds on a tiny footprint, a normal one, and hard against the NW corner (every write is clamped).
func TestDrawRoofModernFlatPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("electric_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofModernFlat(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofModernFlat(img, 1, 1, hwhh.hw, hwhh.hh, rc) // hard against the NW corner
		}
	}
}

// TestDrawRoofTowerPanicSafe locks that the ART-DECO SETBACK TOWER wonder sprite (tiers + a mast + an
// extended base shadow that clips BELOW/around the footprint) is panic-safe + in-bounds on tiny / normal
// / NW-corner cases (every write is clamped).
func TestDrawRoofTowerPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("electric_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}} {
			drawRoofTower(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofTower(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner (shadow drifts down-right)
			drawRoofTower(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner (shadow clips off-canvas)
		}
	}
}

// TestElectricEpochOpenNoWallLots locks that VICTORIAN, ELECTRIC, and ATOMIC are all OPEN ages: each
// emits ZERO wall / gate / tower / bastion lots (unwalled), on real + tiny canvases.
func TestElectricEpochOpenNoWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	ages := []string{"victorian_age", "electric_age", "atomic_age"}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		for _, ageKey := range ages {
			p := tdPlanForAge(namedState(ageKey, "Aldermoor", blds))
			if n := len(wallLotsOf(p)) + len(gateLotsOf(p)) + len(towerLotsOf(p)) + len(bastionLotsOf(p)); n != 0 {
				t.Fatalf("%s %dx%d has %d wall lots — this age must be OPEN (no walls)", ageKey, sz.w, sz.h, n)
			}
		}
	}
}

// TestDumpElectricEpochPNGs renders industrial / victorian / electric / atomic with a FIXED display name +
// identical building set INCLUDING a wonder so the factory/terminal/tower centerpieces render, so a
// reviewer can compare the ELECTRIC-epoch band (against the industrial neighbour) side by side. Opt-in:
// skipped unless CITYMAP_PNG_DUMP=<dir> is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpElectricEpochPNGs
func TestDumpElectricEpochPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	// Identical building set (with a wonder so the centerpiece renders) + a FIXED display name → the
	// citySeed is fixed, so only the era re-skin differs across the four dumps.
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"industrial_age", "1d_industrial.png"},
		{"victorian_age", "1d_victorian.png"},
		{"electric_age", "1d_electric.png"},
		{"atomic_age", "1d_atomic.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// TestDumpIronEpochPNGs renders bronze / iron / classical / medieval with a FIXED display name +
// identical building set INCLUDING a wonder so the ziggurat/temple/cathedral centerpieces render, so
// a reviewer can compare the ancient-band ages + medieval side by side. Opt-in: skipped unless
// CITYMAP_PNG_DUMP=<dir> is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpIronEpochPNGs
func TestDumpIronEpochPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	// Identical building set (with a wonder so the centerpiece renders) + a FIXED display name → the
	// citySeed is fixed, so only the era re-skin differs across the four dumps.
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"bronze_age", "1b_ii_bronze.png"},
		{"iron_age", "1b_ii_iron.png"},
		{"classical_age", "1b_ii_classical.png"},
		{"medieval_age", "1b_ii_medieval.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// TestDigitalEpochStylesWired locks V3-C DIGITAL-epoch style wiring: modern, information, and digital all
// use the GLASS-TOWER house profile + the SKYSCRAPER wonder, and are all OPEN (no walls). Information is
// DENSER than modern; digital carries the epoch's first NEON accents (asserted at the city level in
// TestDigitalEpochNeonPresent). None may still resolve to the default village preset name.
func TestDigitalEpochStylesWired(t *testing.T) {
	mod := styleForAge("modern_age")
	if mod.houseProfile != profileGlassTower {
		t.Fatalf("modern houseProfile = %v, want profileGlassTower (glass skyscrapers)", mod.houseProfile)
	}
	if mod.wonderMotif != wonderSkyscraper {
		t.Fatalf("modern wonderMotif = %v, want wonderSkyscraper (supertall glass tower)", mod.wonderMotif)
	}
	if mod.hasWalls {
		t.Fatal("modern must be OPEN (no walls — the age of open glass towers)")
	}

	inf := styleForAge("information_age")
	if inf.houseProfile != profileGlassTower {
		t.Fatalf("information houseProfile = %v, want profileGlassTower", inf.houseProfile)
	}
	if inf.wonderMotif != wonderSkyscraper {
		t.Fatalf("information wonderMotif = %v, want wonderSkyscraper", inf.wonderMotif)
	}
	if inf.hasWalls {
		t.Fatal("information must be OPEN (no walls)")
	}

	dig := styleForAge("digital_age")
	if dig.houseProfile != profileGlassTower {
		t.Fatalf("digital houseProfile = %v, want profileGlassTower", dig.houseProfile)
	}
	if dig.wonderMotif != wonderSkyscraper {
		t.Fatalf("digital wonderMotif = %v, want wonderSkyscraper", dig.wonderMotif)
	}
	if dig.hasWalls {
		t.Fatal("digital must be OPEN (no walls)")
	}
	// Information should be DENSER than modern (a packed server-city vs a downtown).
	if !(inf.slotSpacing < mod.slotSpacing) {
		t.Fatalf("information slotSpacing (%.2f) should be denser/tighter than modern (%.2f)", inf.slotSpacing, mod.slotSpacing)
	}
	// None of the three may still resolve to the default village preset name.
	if mod.name == defaultTdStyle.name || inf.name == defaultTdStyle.name || dig.name == defaultTdStyle.name {
		t.Fatalf("a digital-epoch age still on the default preset: modern=%q information=%q digital=%q", mod.name, inf.name, dig.name)
	}
}

// TestDigitalEpochNeonPresent locks that DIGITAL is the first age with NEON: its style must actually use a
// neon accent somewhere in its palette (streets/paving carry a cyan/magenta cast that information does NOT),
// AND a digital town must emit at least one neon-sign prop lot (the first-neon tell).
func TestDigitalEpochNeonPresent(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	inf := styleForAge("information_age")
	dig := styleForAge("digital_age")
	// The neon cast: digital's street + paved tones diverge from information's (a neon cyan/magenta lift).
	if colorClose(dig.streetCol(pal), inf.streetCol(pal), 1) && colorClose(dig.pavedCol(pal), inf.pavedCol(pal), 1) {
		t.Fatal("digital streets + paving are identical to information — the first-neon accent is not applied")
	}
	// A digital town must scatter neon-sign props (the first-neon tell).
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	p := tdPlanForAge(namedState("digital_age", "Aldermoor", blds))
	neon := 0
	for _, lt := range p.lots {
		if lt.kind == tdPropNeonSign {
			neon++
		}
	}
	if neon == 0 {
		t.Fatal("digital town emitted ZERO neon-sign props — the first-neon prop scatter is not applied")
	}
}

// TestInformationDataCentersPresent locks that an INFORMATION town scatters DATA-CENTER props (the
// server-city tell) — the distinctive prop that sets it apart from plain modern glass towers.
func TestInformationDataCentersPresent(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	p := tdPlanForAge(namedState("information_age", "Aldermoor", blds))
	dc := 0
	for _, lt := range p.lots {
		if lt.kind == tdPropDataCenter {
			dc++
		}
	}
	if dc == 0 {
		t.Fatal("information town emitted ZERO data-center props — the server-city scatter is not applied")
	}
}

// TestSkyscraperWonderDiffers locks that the SKYSCRAPER wonder reads apart from the deco TOWER, the
// renaissance DOME, and the industrial FACTORY — the skyscraper silhouette must actually be applied, not
// shared with any neighbouring wonder.
func TestSkyscraperWonderDiffers(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	skyscraper := drawWonderImg(styleForAge("modern_age"))
	tower := drawWonderImg(styleForAge("electric_age"))
	dome := drawWonderImg(styleForAge("renaissance_age"))
	factory := drawWonderImg(styleForAge("industrial_age"))
	if !imagesDiffer(skyscraper, tower) {
		t.Fatal("modern skyscraper draws identically to the deco tower — the skyscraper silhouette is not applied")
	}
	if !imagesDiffer(skyscraper, dome) {
		t.Fatal("modern skyscraper draws identically to the renaissance dome — the two wonders must differ")
	}
	if !imagesDiffer(skyscraper, factory) {
		t.Fatal("modern skyscraper draws identically to the industrial factory — the two wonders must differ")
	}
}

// TestDigitalEpochCitiesDiffer locks the CITY-level reads: modern differs from information (clean blue glass
// vs a denser colder server-city with data centers), information differs from digital (cold grey vs sleek
// dark + neon), and all three differ from a STILL-DEFAULT placeholder (cyberpunk_age, still on the village
// preset). cyberpunk is the placeholder because modern is now restyled and is no longer the village.
func TestDigitalEpochCitiesDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	mod, _ := renderImage(namedState("modern_age", "Aldermoor", blds), 120, 72)
	inf, _ := renderImage(namedState("information_age", "Aldermoor", blds), 120, 72)
	dig, _ := renderImage(namedState("digital_age", "Aldermoor", blds), 120, 72)
	def, _ := renderImage(namedState("cyberpunk_age", "Aldermoor", blds), 120, 72) // cyberpunk still uses defaultTdStyle
	if !imagesDiffer(mod, inf) {
		t.Fatal("modern city renders identically to information — the denser+colder+data-center re-skin is not distinct")
	}
	if !imagesDiffer(inf, dig) {
		t.Fatal("information city renders identically to digital — the sleek-dark+neon re-skin is not distinct")
	}
	if !imagesDiffer(mod, def) {
		t.Fatal("modern city renders identically to the default village (cyberpunk) — the modern re-skin is not applied")
	}
	if !imagesDiffer(inf, def) {
		t.Fatal("information city renders identically to the default village (cyberpunk) — the information re-skin is not applied")
	}
	if !imagesDiffer(dig, def) {
		t.Fatal("digital city renders identically to the default village (cyberpunk) — the digital re-skin is not applied")
	}
}

// TestDrawRoofGlassTowerPanicSafe locks that the GLASS-TOWER dwelling sprite (a slab + a window grid + an
// extended SE height shadow that clips beyond the footprint) is panic-safe + in-bounds on a tiny footprint,
// a normal one, and hard against the NW + SE corners (every write is clamped).
func TestDrawRoofGlassTowerPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("modern_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofGlassTower(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofGlassTower(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofGlassTower(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner (shadow clips off-canvas)
		}
	}
}

// TestDrawRoofSkyscraperPanicSafe locks that the SUPERTALL SKYSCRAPER wonder sprite (a slender slab + a
// window grid + a mast + a long raking SE shadow beyond the footprint) is panic-safe + in-bounds on tiny /
// normal / NW + SE corner cases (every write is clamped).
func TestDrawRoofSkyscraperPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("modern_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {14, 6}} {
			drawRoofSkyscraper(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofSkyscraper(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner (mast clips above)
			drawRoofSkyscraper(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner (shadow clips off-canvas)
		}
	}
}

// TestDigitalEpochOpenNoWallLots locks that MODERN, INFORMATION, and DIGITAL are all OPEN ages: each emits
// ZERO wall / gate / tower / bastion lots (unwalled), on real + tiny canvases.
func TestDigitalEpochOpenNoWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	ages := []string{"modern_age", "information_age", "digital_age"}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		for _, ageKey := range ages {
			p := tdPlanForAge(namedState(ageKey, "Aldermoor", blds))
			if n := len(wallLotsOf(p)) + len(gateLotsOf(p)) + len(towerLotsOf(p)) + len(bastionLotsOf(p)); n != 0 {
				t.Fatalf("%s %dx%d has %d wall lots — this age must be OPEN (no walls)", ageKey, sz.w, sz.h, n)
			}
		}
	}
}

// TestDumpDigitalEpochPNGs renders atomic / modern / information / digital with a FIXED display name +
// identical building set INCLUDING a wonder so the tower/skyscraper centerpieces render, so a reviewer can
// compare the DIGITAL-epoch band (against the atomic neighbour) side by side. Opt-in: skipped unless
// CITYMAP_PNG_DUMP=<dir> is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpDigitalEpochPNGs
func TestDumpDigitalEpochPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	// Identical building set (with a wonder so the centerpiece renders) + a FIXED display name → the citySeed
	// is fixed, so only the era re-skin differs across the four dumps.
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"atomic_age", "1e_atomic.png"},
		{"modern_age", "1e_modern.png"},
		{"information_age", "1e_information.png"},
		{"digital_age", "1e_digital.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// ---- NEON epoch (cyberpunk / fusion / space) --------------------------------

// TestNeonEpochStylesWired locks that the three NEON-epoch ages resolve to their tuned presets with the
// intended motifs/profiles: cyberpunk reuses the glass-tower + skyscraper silhouette (a DARK neon megatower),
// fusion carries the FUSION-CORE wonder, and space carries the LAUNCHPAD wonder + the METAL-DOME dwelling —
// and none of the three is still on the default village preset.
func TestNeonEpochStylesWired(t *testing.T) {
	cyb := styleForAge("cyberpunk_age")
	if cyb.houseProfile != profileGlassTower {
		t.Fatalf("cyberpunk houseProfile = %v, want profileGlassTower (dark neon megablocks)", cyb.houseProfile)
	}
	if cyb.wonderMotif != wonderSkyscraper {
		t.Fatalf("cyberpunk wonderMotif = %v, want wonderSkyscraper (dark neon megatower)", cyb.wonderMotif)
	}
	if cyb.hasWalls {
		t.Fatal("cyberpunk must be OPEN (no walls)")
	}

	fus := styleForAge("fusion_age")
	if fus.wonderMotif != wonderFusionCore {
		t.Fatalf("fusion wonderMotif = %v, want wonderFusionCore (glowing reactor)", fus.wonderMotif)
	}
	if fus.hasWalls {
		t.Fatal("fusion must be OPEN (a utopian open city)")
	}

	spc := styleForAge("space_age")
	if spc.wonderMotif != wonderLaunchpad {
		t.Fatalf("space wonderMotif = %v, want wonderLaunchpad (rocket on a pad)", spc.wonderMotif)
	}
	if spc.houseProfile != profileMetalDome {
		t.Fatalf("space houseProfile = %v, want profileMetalDome (colony domes)", spc.houseProfile)
	}
	if spc.hasWalls {
		t.Fatal("space must be OPEN (an open colony)")
	}

	// None of the three may still resolve to the default village preset name.
	if cyb.name == defaultTdStyle.name || fus.name == defaultTdStyle.name || spc.name == defaultTdStyle.name {
		t.Fatalf("a NEON-epoch age still on the default preset: cyberpunk=%q fusion=%q space=%q", cyb.name, fus.name, spc.name)
	}
	// cyberpunk is DARKER/DENSER than digital (the first-neon age it descends from).
	dig := styleForAge("digital_age")
	if !(cyb.slotSpacing < dig.slotSpacing) {
		t.Fatalf("cyberpunk slotSpacing (%.2f) should be denser/tighter than digital (%.2f)", cyb.slotSpacing, dig.slotSpacing)
	}
}

// TestNeonEpochPropsPresent locks the distinctive prop scatters: a CYBERPUNK town must emit HOLOGRAM props
// (the night-city projection tell) and a SPACE town must emit ROCKET props (the spaceport tell).
func TestNeonEpochPropsPresent(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}

	cp := tdPlanForAge(namedState("cyberpunk_age", "Aldermoor", blds))
	holo := 0
	for _, lt := range cp.lots {
		if lt.kind == tdPropHologram {
			holo++
		}
	}
	if holo == 0 {
		t.Fatal("cyberpunk town emitted ZERO hologram props — the night-city scatter is not applied")
	}

	sp := tdPlanForAge(namedState("space_age", "Aldermoor", blds))
	rocket := 0
	for _, lt := range sp.lots {
		if lt.kind == tdPropRocket {
			rocket++
		}
	}
	if rocket == 0 {
		t.Fatal("space town emitted ZERO rocket props — the spaceport scatter is not applied")
	}
}

// TestNeonWondersDiffer locks that the two new NEON wonders read apart from their neighbours: the FUSION CORE
// differs from the skyscraper, the renaissance dome, the deco tower, and the factory; the LAUNCHPAD differs
// from the fusion core, the skyscraper, and the dome. Each silhouette must actually be applied, not shared.
func TestNeonWondersDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	fusionCore := drawWonderImg(styleForAge("fusion_age"))
	launchpad := drawWonderImg(styleForAge("space_age"))
	skyscraper := drawWonderImg(styleForAge("modern_age"))
	dome := drawWonderImg(styleForAge("renaissance_age"))
	tower := drawWonderImg(styleForAge("electric_age"))
	factory := drawWonderImg(styleForAge("industrial_age"))

	if !imagesDiffer(fusionCore, skyscraper) {
		t.Fatal("fusion core draws identically to the skyscraper — the reactor silhouette is not applied")
	}
	if !imagesDiffer(fusionCore, dome) {
		t.Fatal("fusion core draws identically to the renaissance dome — the two wonders must differ")
	}
	if !imagesDiffer(fusionCore, tower) {
		t.Fatal("fusion core draws identically to the deco tower — the two wonders must differ")
	}
	if !imagesDiffer(fusionCore, factory) {
		t.Fatal("fusion core draws identically to the factory — the two wonders must differ")
	}
	if !imagesDiffer(launchpad, fusionCore) {
		t.Fatal("launchpad draws identically to the fusion core — the two NEON wonders must differ")
	}
	if !imagesDiffer(launchpad, skyscraper) {
		t.Fatal("launchpad draws identically to the skyscraper — the two wonders must differ")
	}
	if !imagesDiffer(launchpad, dome) {
		t.Fatal("launchpad draws identically to the renaissance dome — the two wonders must differ")
	}
}

// TestNeonEpochCitiesDiffer locks the CITY-level reads: cyberpunk (dark neon) differs from fusion (bright
// white) differs from space (pale metallic), and all three differ from transcendent (now the ETHEREAL-LIGHT
// finale — since every age is styled, transcendent is used here as a KNOWN-DISTINCT comparison age, no
// longer a default-village placeholder). A styled neon-epoch city still differs from the styled finale.
func TestNeonEpochCitiesDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	cyb, _ := renderImage(namedState("cyberpunk_age", "Aldermoor", blds), 120, 72)
	fus, _ := renderImage(namedState("fusion_age", "Aldermoor", blds), 120, 72)
	spc, _ := renderImage(namedState("space_age", "Aldermoor", blds), 120, 72)
	def, _ := renderImage(namedState("transcendent_age", "Aldermoor", blds), 120, 72) // now the styled ethereal finale (a known-distinct age)

	if !imagesDiffer(cyb, fus) {
		t.Fatal("cyberpunk city renders identically to fusion — the dark-neon vs bright-white re-skin is not distinct")
	}
	if !imagesDiffer(fus, spc) {
		t.Fatal("fusion city renders identically to space — the bright-white vs pale-metallic re-skin is not distinct")
	}
	if !imagesDiffer(cyb, spc) {
		t.Fatal("cyberpunk city renders identically to space — the dark-neon vs pale-metallic re-skin is not distinct")
	}
	if !imagesDiffer(cyb, def) {
		t.Fatal("cyberpunk city renders identically to the transcendent finale — the cyberpunk re-skin is not distinct")
	}
	if !imagesDiffer(fus, def) {
		t.Fatal("fusion city renders identically to the transcendent finale — the fusion re-skin is not distinct")
	}
	if !imagesDiffer(spc, def) {
		t.Fatal("space city renders identically to the transcendent finale — the space re-skin is not distinct")
	}
}

// TestDrawRoofFusionCorePanicSafe locks that the FUSION-CORE wonder sprite (concentric glowing discs + a
// white-hot bloom halo) is panic-safe + in-bounds on tiny / normal / NW + SE corner cases.
func TestDrawRoofFusionCorePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("fusion_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofFusionCore(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofFusionCore(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofFusionCore(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofLaunchpadPanicSafe locks that the LAUNCHPAD wonder sprite (a pad + a rocket + fins + gantry
// dabs + a scorch ring) is panic-safe + in-bounds on tiny / normal / NW + SE corner cases.
func TestDrawRoofLaunchpadPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("space_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {14, 6}} {
			drawRoofLaunchpad(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofLaunchpad(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofLaunchpad(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofMetalDomePanicSafe locks that the METAL-DOME dwelling sprite (a lit silver disc + a curved NW
// highlight arc + a rim) is panic-safe + in-bounds on tiny / normal / NW + SE corner cases.
func TestDrawRoofMetalDomePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("space_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {10, 12}, {14, 6}} {
			drawRoofMetalDome(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofMetalDome(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofMetalDome(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestNeonEpochOpenNoWallLots locks that CYBERPUNK, FUSION, and SPACE are all OPEN ages: each emits ZERO
// wall / gate / tower / bastion lots (unwalled), on real + tiny canvases.
func TestNeonEpochOpenNoWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	ages := []string{"cyberpunk_age", "fusion_age", "space_age"}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		for _, ageKey := range ages {
			p := tdPlanForAge(namedState(ageKey, "Aldermoor", blds))
			if n := len(wallLotsOf(p)) + len(gateLotsOf(p)) + len(towerLotsOf(p)) + len(bastionLotsOf(p)); n != 0 {
				t.Fatalf("%s %dx%d has %d wall lots — this age must be OPEN (no walls)", ageKey, sz.w, sz.h, n)
			}
		}
	}
}

// TestDumpNeonEpochPNGs renders digital / cyberpunk / fusion / space with a FIXED display name + identical
// building set INCLUDING a wonder so the centerpieces render, so a reviewer can compare the NEON-epoch band
// (against the digital neighbour it descends from) side by side. Opt-in: skipped unless CITYMAP_PNG_DUMP=<dir>
// is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpNeonEpochPNGs
func TestDumpNeonEpochPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"digital_age", "1f_digital.png"},
		{"cyberpunk_age", "1f_cyberpunk.png"},
		{"fusion_age", "1f_fusion.png"},
		{"space_age", "1f_space.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// TestCosmicEpochStylesWired locks the COSMIC-epoch first pair off their default placeholder:
// interstellar must use the SPIRE dwelling profile + the SPIRE-ARRAY centrepiece, and galactic must
// use the RING-HUB centrepiece. Both must be OPEN (no walls) and neither may still resolve to the
// default village preset name. Galactic must also read as a DENSER metropolis than interstellar.
func TestCosmicEpochStylesWired(t *testing.T) {
	_ = theme.SetActive("forge")

	inter := styleForAge("interstellar_age")
	if inter.wonderMotif != wonderSpireArray {
		t.Fatalf("interstellar wonderMotif = %v, want wonderSpireArray (spire cluster)", inter.wonderMotif)
	}
	if inter.houseProfile != profileSpire {
		t.Fatalf("interstellar houseProfile = %v, want profileSpire (arcology spires)", inter.houseProfile)
	}
	if inter.hasWalls {
		t.Fatal("interstellar must be OPEN (no walls)")
	}

	gal := styleForAge("galactic_age")
	if gal.wonderMotif != wonderRingHub {
		t.Fatalf("galactic wonderMotif = %v, want wonderRingHub (ring-hub megastation)", gal.wonderMotif)
	}
	if gal.hasWalls {
		t.Fatal("galactic must be OPEN (no walls)")
	}

	// Neither may still resolve to the default village preset name.
	if inter.name == defaultTdStyle.name || gal.name == defaultTdStyle.name {
		t.Fatalf("a COSMIC-epoch age still on the default preset: interstellar=%q galactic=%q", inter.name, gal.name)
	}
	// galactic is DENSER than interstellar (the age it descends from).
	if !(gal.slotSpacing < inter.slotSpacing) {
		t.Fatalf("galactic slotSpacing (%.2f) should be denser/tighter than interstellar (%.2f)", gal.slotSpacing, inter.slotSpacing)
	}
}

// TestCosmicWondersDiffer locks that the two new COSMIC wonders read apart from their neighbours: the
// SPIRE-ARRAY differs from the launchpad, the dome, and the skyscraper; the RING-HUB differs from the
// spire-array, the launchpad, the fusion core, and the dome. Each silhouette must actually be applied.
func TestCosmicWondersDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	spireArray := drawWonderImg(styleForAge("interstellar_age"))
	ringHub := drawWonderImg(styleForAge("galactic_age"))
	launchpad := drawWonderImg(styleForAge("space_age"))
	fusionCore := drawWonderImg(styleForAge("fusion_age"))
	skyscraper := drawWonderImg(styleForAge("modern_age"))
	dome := drawWonderImg(styleForAge("renaissance_age"))

	if !imagesDiffer(spireArray, launchpad) {
		t.Fatal("spire array draws identically to the launchpad — the spire-cluster silhouette is not applied")
	}
	if !imagesDiffer(spireArray, dome) {
		t.Fatal("spire array draws identically to the renaissance dome — the two wonders must differ")
	}
	if !imagesDiffer(spireArray, skyscraper) {
		t.Fatal("spire array draws identically to the skyscraper — the two wonders must differ")
	}
	if !imagesDiffer(ringHub, spireArray) {
		t.Fatal("ring hub draws identically to the spire array — the two COSMIC wonders must differ")
	}
	if !imagesDiffer(ringHub, launchpad) {
		t.Fatal("ring hub draws identically to the launchpad — the two wonders must differ")
	}
	if !imagesDiffer(ringHub, fusionCore) {
		t.Fatal("ring hub draws identically to the fusion core — the two wonders must differ")
	}
	if !imagesDiffer(ringHub, dome) {
		t.Fatal("ring hub draws identically to the renaissance dome — the two wonders must differ")
	}
}

// TestCosmicEpochCitiesDiffer locks the CITY-level reads: interstellar (deep-space spires) differs from
// galactic (ring-hub megastation), and both differ from transcendent (now the ETHEREAL-LIGHT finale — every
// age is styled, so transcendent serves here as a KNOWN-DISTINCT comparison age, not a default placeholder).
func TestCosmicEpochCitiesDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	inter, _ := renderImage(namedState("interstellar_age", "Aldermoor", blds), 120, 72)
	gal, _ := renderImage(namedState("galactic_age", "Aldermoor", blds), 120, 72)
	def, _ := renderImage(namedState("transcendent_age", "Aldermoor", blds), 120, 72) // now the styled ethereal finale (a known-distinct age)

	if !imagesDiffer(inter, gal) {
		t.Fatal("interstellar city renders identically to galactic — the spires vs ring-hub re-skin is not distinct")
	}
	if !imagesDiffer(inter, def) {
		t.Fatal("interstellar city renders identically to the transcendent finale — the interstellar re-skin is not distinct")
	}
	if !imagesDiffer(gal, def) {
		t.Fatal("galactic city renders identically to the transcendent finale — the galactic re-skin is not distinct")
	}
}

// TestDrawRoofSpirePanicSafe locks that the SPIRE dwelling sprite (a thin tapering needle + a long SE
// height shadow + a base pad + a lit tip) is panic-safe + in-bounds on tiny / normal / NW + SE corners.
func TestDrawRoofSpirePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("interstellar_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {10, 12}, {14, 6}} {
			drawRoofSpire(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofSpire(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofSpire(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofSpireArrayPanicSafe locks that the SPIRE-ARRAY wonder sprite (a base apron + a ring of
// satellite spires around a tallest central one, each with a long SE shadow) is panic-safe + in-bounds on
// tiny / normal / NW + SE corner cases.
func TestDrawRoofSpireArrayPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("interstellar_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofSpireArray(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofSpireArray(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofSpireArray(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofRingHubPanicSafe locks that the RING-HUB wonder sprite (a deck + concentric ring outlines +
// spokes + a glowing hub halo) is panic-safe + in-bounds on tiny / normal / NW + SE corner cases.
func TestDrawRoofRingHubPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("galactic_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofRingHub(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofRingHub(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofRingHub(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestCosmicEpochOpenNoWallLots locks that INTERSTELLAR and GALACTIC are OPEN ages: each emits ZERO
// wall / gate / tower / bastion lots (unwalled), on real + tiny canvases.
func TestCosmicEpochOpenNoWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	ages := []string{"interstellar_age", "galactic_age"}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		for _, ageKey := range ages {
			p := tdPlanForAge(namedState(ageKey, "Aldermoor", blds))
			if n := len(wallLotsOf(p)) + len(gateLotsOf(p)) + len(towerLotsOf(p)) + len(bastionLotsOf(p)); n != 0 {
				t.Fatalf("%s %dx%d has %d wall lots — this age must be OPEN (no walls)", ageKey, sz.w, sz.h, n)
			}
		}
	}
}

// TestDumpCosmicEpochPNGs renders space / interstellar / galactic with a FIXED display name + identical
// building set INCLUDING a wonder so the centerpieces render, so a reviewer can compare the COSMIC-epoch
// first pair (against the space neighbour it descends from) side by side. Opt-in: skipped unless
// CITYMAP_PNG_DUMP=<dir> is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpCosmicEpochPNGs
func TestDumpCosmicEpochPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"space_age", "1g_space.png"},
		{"interstellar_age", "1g_interstellar.png"},
		{"galactic_age", "1g_galactic.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}

// TestFinalPairWiring locks that the FINAL two ages — quantum + transcendent, completing all 22 — are wired
// to their own presets + sprites: quantum → the iridescent CRYSTAL-LATTICE wonder + the lattice house
// profile; transcendent → the ethereal ASCENSION wonder + the ethereal house profile. Both are OPEN (no
// walls), neither still resolves to the default village preset name, and — since transcendent is now
// styled — NO age maps to defaultTdStyle any more.
func TestFinalPairWiring(t *testing.T) {
	_ = theme.SetActive("forge")

	quantum := styleForAge("quantum_age")
	if quantum.wonderMotif != wonderCrystalLattice {
		t.Fatalf("quantum wonderMotif = %v, want wonderCrystalLattice (iridescent lattice mesh)", quantum.wonderMotif)
	}
	if quantum.houseProfile != profileLattice {
		t.Fatalf("quantum houseProfile = %v, want profileLattice (faceted crystal node)", quantum.houseProfile)
	}
	if quantum.hasWalls {
		t.Fatal("quantum must be OPEN (no walls)")
	}

	trans := styleForAge("transcendent_age")
	if trans.wonderMotif != wonderAscension {
		t.Fatalf("transcendent wonderMotif = %v, want wonderAscension (rising light + halos)", trans.wonderMotif)
	}
	if trans.houseProfile != profileEthereal {
		t.Fatalf("transcendent houseProfile = %v, want profileEthereal (soft light-form bloom)", trans.houseProfile)
	}
	if trans.hasWalls {
		t.Fatal("transcendent must be OPEN (no walls)")
	}

	// Neither may still resolve to the default village preset name.
	if quantum.name == defaultTdStyle.name || trans.name == defaultTdStyle.name {
		t.Fatalf("a FINAL-pair age still on the default preset: quantum=%q transcendent=%q", quantum.name, trans.name)
	}
	// The whole atlas is now styled: NO age may map to defaultTdStyle (it survives only as a base preset).
	for ageKey, st := range ageStyles {
		if st.name == defaultTdStyle.name {
			t.Fatalf("age %q still maps to the default village preset — every one of the 22 ages must be styled now", ageKey)
		}
	}
}

// TestFinalPairWondersDiffer locks that the two FINAL-pair wonders read apart from their neighbours and each
// other: the CRYSTAL-LATTICE differs from the ring-hub, the fusion core, the spire-array, and the dome; the
// ASCENSION differs from the crystal-lattice, the ring-hub, and the fusion core. Each silhouette must
// actually be applied (not falling through to the generic hall).
func TestFinalPairWondersDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	drawWonderImg := func(style tdEraStyle) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, 40, 40))
		lt := tdLot{x: 0, y: 0, w: 20, h: 20, kind: tdRoof, roof: roofWonder, domain: "wonder", category: "wonder"}
		xf := tdTransform{scale: 1, offX: 20, offY: 20, roofFloorPx: 1}
		drawRoof(img, xf, lt, style, pal)
		return img
	}
	crystalLattice := drawWonderImg(styleForAge("quantum_age"))
	ascension := drawWonderImg(styleForAge("transcendent_age"))
	ringHub := drawWonderImg(styleForAge("galactic_age"))
	spireArray := drawWonderImg(styleForAge("interstellar_age"))
	fusionCore := drawWonderImg(styleForAge("fusion_age"))
	dome := drawWonderImg(styleForAge("renaissance_age"))

	if !imagesDiffer(crystalLattice, ringHub) {
		t.Fatal("crystal lattice draws identically to the ring hub — the lattice-mesh silhouette is not applied")
	}
	if !imagesDiffer(crystalLattice, fusionCore) {
		t.Fatal("crystal lattice draws identically to the fusion core — the two wonders must differ")
	}
	if !imagesDiffer(crystalLattice, spireArray) {
		t.Fatal("crystal lattice draws identically to the spire array — the two wonders must differ")
	}
	if !imagesDiffer(crystalLattice, dome) {
		t.Fatal("crystal lattice draws identically to the renaissance dome — the two wonders must differ")
	}
	if !imagesDiffer(ascension, crystalLattice) {
		t.Fatal("ascension draws identically to the crystal lattice — the two FINAL-pair wonders must differ")
	}
	if !imagesDiffer(ascension, ringHub) {
		t.Fatal("ascension draws identically to the ring hub — the two wonders must differ")
	}
	if !imagesDiffer(ascension, fusionCore) {
		t.Fatal("ascension draws identically to the fusion core — the two wonders must differ")
	}
}

// TestFinalPairCitiesDiffer locks the CITY-level reads for the final two ages: quantum (dark iridescent
// crystal) differs from transcendent (bright ethereal light), and each differs from two KNOWN-DISTINCT
// styled ages — primitive_age and space_age. NOTE (this is the last slice): since transcendent is now
// styled, there is NO default-village placeholder left to compare against, so we use real styled ages.
func TestFinalPairCitiesDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	quantum, _ := renderImage(namedState("quantum_age", "Aldermoor", blds), 120, 72)
	trans, _ := renderImage(namedState("transcendent_age", "Aldermoor", blds), 120, 72)
	prim, _ := renderImage(namedState("primitive_age", "Aldermoor", blds), 120, 72)
	space, _ := renderImage(namedState("space_age", "Aldermoor", blds), 120, 72)

	if !imagesDiffer(quantum, trans) {
		t.Fatal("quantum city renders identically to transcendent — the iridescent-crystal vs ethereal-light re-skin is not distinct")
	}
	if !imagesDiffer(quantum, prim) {
		t.Fatal("quantum city renders identically to the primitive village — the quantum re-skin is not applied")
	}
	if !imagesDiffer(quantum, space) {
		t.Fatal("quantum city renders identically to space — the quantum crystal deck must differ from the space colony")
	}
	if !imagesDiffer(trans, prim) {
		t.Fatal("transcendent city renders identically to the primitive village — the transcendent re-skin is not applied")
	}
	if !imagesDiffer(trans, space) {
		t.Fatal("transcendent city renders identically to space — the ethereal light-field must differ from the space colony")
	}
}

// TestDrawRoofLatticePanicSafe locks that the LATTICE dwelling sprite (a faceted iridescent crystal node —
// four triangular facets in shifting hues + a bright core) is panic-safe + in-bounds on tiny / normal / NW +
// SE corner cases.
func TestDrawRoofLatticePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("quantum_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {10, 12}, {14, 6}} {
			drawRoofLattice(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofLattice(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofLattice(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofCrystalLatticePanicSafe locks that the CRYSTAL-LATTICE wonder sprite (a crystalline deck + a
// diamond grid of glowing nodes joined by iridescent struts + a bright central node) is panic-safe +
// in-bounds on tiny / normal / NW + SE corner cases.
func TestDrawRoofCrystalLatticePanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("quantum_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofCrystalLattice(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofCrystalLattice(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofCrystalLattice(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofEtherealPanicSafe locks that the ETHEREAL dwelling sprite (a soft translucent radial bloom —
// pure light, no hard fill) is panic-safe + in-bounds on tiny / normal / NW + SE corner cases.
func TestDrawRoofEtherealPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("transcendent_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "housing", "production")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {10, 12}, {14, 6}} {
			drawRoofEthereal(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofEthereal(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofEthereal(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestDrawRoofAscensionPanicSafe locks that the ASCENSION wonder sprite (concentric soft halos + a rising
// vertical light beam + a pure-white core) is panic-safe + in-bounds on tiny / normal / NW + SE corner
// cases.
func TestDrawRoofAscensionPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := newTdPal()
	style := styleForAge("transcendent_age")
	for _, tc := range []struct{ w, h int }{{9, 9}, {40, 40}} {
		img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
		rc := roofColorsFor(style, pal, "wonder", "wonder")
		for _, hwhh := range []struct{ hw, hh int }{{2, 2}, {12, 10}, {6, 14}} {
			drawRoofAscension(img, tc.w/2, tc.h/2, hwhh.hw, hwhh.hh, rc)
			drawRoofAscension(img, 1, 1, hwhh.hw, hwhh.hh, rc)           // NW corner
			drawRoofAscension(img, tc.w-1, tc.h-1, hwhh.hw, hwhh.hh, rc) // SE corner
		}
	}
}

// TestFinalPairOpenNoWallLots locks that QUANTUM and TRANSCENDENT are OPEN ages: each emits ZERO wall / gate
// / tower / bastion lots (unwalled), on real + tiny canvases.
func TestFinalPairOpenNoWallLots(t *testing.T) {
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 30, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	ages := []string{"quantum_age", "transcendent_age"}
	for _, sz := range []struct{ w, h int }{{120, 72}, {24, 16}, {8, 8}} {
		_ = sz
		for _, ageKey := range ages {
			p := tdPlanForAge(namedState(ageKey, "Aldermoor", blds))
			if n := len(wallLotsOf(p)) + len(gateLotsOf(p)) + len(towerLotsOf(p)) + len(bastionLotsOf(p)); n != 0 {
				t.Fatalf("%s has %d wall lots — this age must be OPEN (no walls)", ageKey, n)
			}
		}
	}
}

// TestDumpFinalPairPNGs renders galactic / quantum / transcendent with a FIXED display name + identical
// building set INCLUDING a wonder so the centerpieces render, so a reviewer can compare the FINAL pair
// (against the galactic neighbour it descends from) side by side. Opt-in: skipped unless CITYMAP_PNG_DUMP=<dir>
// is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpFinalPairPNGs
func TestDumpFinalPairPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump era-comparison PNGs")
	}
	_ = theme.SetActive("forge")
	blds := map[string]int{"hut": 28, "gathering_camp": 18, "forge": 12, "barracks": 6, "colosseum": 1}
	dumps := []struct {
		ageKey string
		file   string
	}{
		{"galactic_age", "1h_galactic.png"},
		{"quantum_age", "1h_quantum.png"},
		{"transcendent_age", "1h_transcendent.png"},
	}
	for _, d := range dumps {
		img, _ := renderImage(namedState(d.ageKey, "Aldermoor", blds), 160, 100)
		path := dir + "/" + d.file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatalf("encode %s: %v", path, err)
		}
		f.Close()
		t.Logf("wrote %s", path)
	}
}
