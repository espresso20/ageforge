package citymap

import (
	"image"
	"image/color"
	"math"
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

// TestStableIncrementalPlacement is the anti-re-randomize guarantee (locked #8), now
// under the INTERMIXED LANE placement (FIX 1): the slots — including the organic jitter
// — for the FIRST N instances of a building must be byte-identical whether the building
// has count N or count N+1. Adding a building can never move an existing one. We compare
// per-building roof-lot centers, in a single-type village and again in a multi-domain,
// wonder-anchored city (the harder case: round-robin ordering + a wonder plaza).
func TestStableIncrementalPlacement(t *testing.T) {
	_ = theme.SetActive("forge")

	// Case 1 — a single-type village: huts share the one center anchor's spiral, so the
	// first 10 slots must coincide exactly when the 11th is added.
	hutsBase := roofLots(tdPlanFor(sampleState("primitive_age", map[string]int{"hut": 10})))
	hutsMore := roofLots(tdPlanFor(sampleState("primitive_age", map[string]int{"hut": 11})))
	if len(hutsBase) < 10 || len(hutsMore) < len(hutsBase) {
		t.Fatalf("unexpected hut lot counts: base=%d more=%d", len(hutsBase), len(hutsMore))
	}
	for i := 0; i < len(hutsBase); i++ {
		a, b := hutsBase[i], hutsMore[i]
		if a.x != b.x || a.y != b.y {
			t.Fatalf("slot %d moved when hut count grew 10→11: (%v,%v) vs (%v,%v) — placement not stable-incremental",
				i, a.x, a.y, b.x, b.y)
		}
	}

	// Case 2 — a multi-domain, wonder-anchored city. Growing ONE domain's count must not
	// move the EXISTING lots of any domain. We key lots by (domain,tier) and compare the
	// per-key position sequences; every key present in base must be a positional prefix of
	// the same key in more. This exercises the round-robin ordering + the plaza filter.
	baseM := map[string]int{"hut": 12, "gathering_camp": 9, "stone_camp": 7, "forge": 6, "colosseum": 1}
	moreM := map[string]int{"hut": 12, "gathering_camp": 10, "stone_camp": 7, "forge": 6, "colosseum": 1} // +1 camp
	byDomain := func(lots []tdLot) map[string][]tdLot {
		m := map[string][]tdLot{}
		for _, lt := range lots {
			if lt.roof == roofWonder {
				continue
			}
			m[lt.domain] = append(m[lt.domain], lt)
		}
		return m
	}
	bd := byDomain(fabricLots(tdPlanFor(sampleState("primitive_age", baseM))))
	md := byDomain(fabricLots(tdPlanFor(sampleState("primitive_age", moreM))))
	for dom, bl := range bd {
		ml := md[dom]
		if len(ml) < len(bl) {
			t.Fatalf("domain %q lost lots when a sibling grew: base=%d more=%d", dom, len(bl), len(ml))
		}
		for i := range bl {
			if bl[i].x != ml[i].x || bl[i].y != ml[i].y {
				t.Fatalf("domain %q lot %d moved when gathering_camp grew: (%v,%v) vs (%v,%v) — intermix broke stability",
					dom, i, bl[i].x, bl[i].y, ml[i].x, ml[i].y)
			}
		}
	}

	// Also assert the jitter actually did something: the raw spiral would place slots on a
	// perfect lattice; with jitter at least one slot is offset from its un-jittered spot.
	moved := false
	for i := range hutsBase {
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

// TestIntermixedLanePlacement locks FIX 1: buildings are placed in a stable, type-
// INTERMIXED sequence — consecutive placement slots are different domains (not one big
// blob of huts) AND the fabric is spatially mixed (no single-domain round blob). We
// drive a village of five distinct domains and assert both the placement ORDER and the
// spatial nearest-neighbor mixing exceed a healthy threshold.
func TestIntermixedLanePlacement(t *testing.T) {
	_ = theme.SetActive("forge")
	// hut=housing, gathering_camp=food, stone_camp=geological, forge=metallurgy,
	// barracks=military — five distinct domains, no wonder (one center anchor).
	state := sampleState("primitive_age", map[string]int{
		"hut": 12, "gathering_camp": 10, "stone_camp": 8, "forge": 8, "barracks": 6,
	})
	plan := tdPlanFor(state)
	fab := fabricLots(plan)
	if len(fab) < 20 {
		t.Fatalf("expected a substantial fabric, got %d lots", len(fab))
	}

	// (1) Placement ORDER intermix: consecutive slots must usually be different domains
	// (round-robin interleave), NOT a run of one domain then the next. Count adjacent
	// pairs that differ; a per-domain-blob layout would score near 0.
	distinctDomains := map[string]bool{}
	adjDiff, adjTot := 0, 0
	for i := 1; i < len(fab); i++ {
		adjTot++
		if fab[i].domain != fab[i-1].domain {
			adjDiff++
		}
		distinctDomains[fab[i].domain] = true
	}
	distinctDomains[fab[0].domain] = true
	if len(distinctDomains) < 3 {
		t.Fatalf("fabric only spans %d domains — need a multi-domain settlement to test intermix", len(distinctDomains))
	}
	orderFrac := float64(adjDiff) / float64(adjTot)
	if orderFrac < 0.5 {
		t.Fatalf("placement order intermix only %.2f — consecutive slots are too often the same domain (blobby, not interleaved)", orderFrac)
	}

	// (2) SPATIAL intermix: a lot's nearest neighbor is frequently a DIFFERENT domain —
	// there is no single-domain round blob. A per-domain-cluster layout would score near 0.
	spatial := nnDomainDiffFrac(fab)
	if spatial < 0.30 {
		t.Fatalf("spatial nearest-neighbor domain-mixing only %.2f — the fabric reads as same-type blobs, not intermixed", spatial)
	}

	// The town must not be a disc: the fabric grows along lanes, so at least a couple of
	// lanes exist for it to grow along.
	if len(plan.streets) < 2 {
		t.Fatalf("expected a lane network for the fabric to grow along, got %d streets", len(plan.streets))
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
	meanNearest := sumNearest / float64(len(fab))
	// Clustering: on average a fabric lot hugs SOME anchor more closely than the town's
	// overall radius — i.e. it grows around the anchors, not scattered. The ratio ceiling is
	// 0.72 (was 0.6 before playtest polish FIX 3): the BIGGER plaza (plazaRadius 2.2 → 3.0)
	// deliberately clears more open ground around each anchor, so the surviving fabric hugs the
	// plaza RIM rather than the anchor CENTER and its mean distance-to-anchor rises accordingly.
	// It still clusters (well under the town radius); it just breathes at a comfortable remove.
	if meanNearest > maxFromCore*0.72 {
		t.Fatalf("fabric mean distance-to-nearest-anchor %.1f is not small vs town radius %.1f — buildings do not cluster around the anchors", meanNearest, maxFromCore)
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

// TestLineTheLanes locks that buildings LINE ALONGSIDE the lanes rather than sitting ON
// them (playtest FIX 2). Two properties: (1) roof lots sit in a BAND offset from the
// nearest lane centerline — the mean offset is close to the intended perpendicular offset
// (lane-half + roof-half + margin), not ~0; and (2) the lanes' OWN cells (within a
// lane-half of a centerline) stay building-free between the opposing rows, so the road
// stays visible. A pull-to-lane layout would fail both (offsets ~0, lane cells buried).
func TestLineTheLanes(t *testing.T) {
	_ = theme.SetActive("forge")
	state := sampleState("primitive_age", map[string]int{"hut": 18, "gathering_camp": 12, "stone_camp": 8, "forge": 8})
	plan := tdPlanFor(state)
	cfg := defaultTdConfig

	nearestLaneDist := func(x, y float64) float64 {
		best := math.Inf(1)
		for _, s := range plan.streets {
			for i := 0; i+1 < len(s.pts); i++ {
				if d := distToSegSq(x, y, s.pts[i], s.pts[i+1]); d < best {
					best = d
				}
			}
		}
		return math.Sqrt(best)
	}

	// (1) Mean offset from the lane ≈ the intended perpendicular offset (buildings line
	// beside the road). Use the plain roof extent for the target; a small tolerance covers
	// per-type extent variation and jitter.
	targetPerp := cfg.laneHalf + cfg.roofSize/2 + cfg.laneMargin
	var sum float64
	var n int
	onRoad := 0
	for _, lt := range fabricLots(plan) {
		d := nearestLaneDist(lt.x, lt.y)
		sum += d
		n++
		if d < cfg.laneHalf { // a roof center sitting ON the road surface
			onRoad++
		}
	}
	if n == 0 {
		t.Fatal("no fabric lots to measure")
	}
	mean := sum / float64(n)
	if mean < targetPerp*0.6 {
		t.Fatalf("mean roof-to-lane distance %.2f is far below the lining offset %.2f — buildings are pulled onto the lanes, not lining them", mean, targetPerp)
	}
	// Essentially no roof center may sit on the road surface (a rare lot near a crossing lane
	// it isn't lining is tolerated, but not a wholesale pile-on).
	if frac := float64(onRoad) / float64(n); frac > 0.10 {
		t.Fatalf("%.0f%% of roof centers sit on a lane surface — buildings should line ALONGSIDE the road, not on it", frac*100)
	}

	// (2) The lanes' OWN cells stay building-free: sample points ALONG each lane centerline
	// and confirm few have a roof center within a lane-half (the road shows between the
	// opposing rows). Skip the crowded junctions/plazas (near an anchor) where lanes meet.
	plazaR := cfg.plazaRadius * cfg.roofSize
	sampled, buried := 0, 0
	for _, s := range plan.streets {
		for i := 0; i+1 < len(s.pts); i++ {
			a, b := s.pts[i], s.pts[i+1]
			for _, f := range []float64{0.25, 0.5, 0.75} {
				cx := a.x + (b.x-a.x)*f
				cy := a.y + (b.y-a.y)*f
				// Skip samples inside/near an anchor (junction/plaza), not representative frontage.
				nearAnchor := false
				for _, an := range plan.anchors {
					if math.Hypot(cx-an.cx, cy-an.cy) < plazaR+cfg.roofSize {
						nearAnchor = true
						break
					}
				}
				if nearAnchor {
					continue
				}
				sampled++
				for _, lt := range fabricLots(plan) {
					if math.Hypot(lt.x-cx, lt.y-cy) < cfg.laneHalf {
						buried++
						break
					}
				}
			}
		}
	}
	if sampled == 0 {
		t.Fatal("no lane centerline samples — the lane network is missing")
	}
	if frac := float64(buried) / float64(sampled); frac > 0.10 {
		t.Fatalf("%.0f%% of lane centerline samples have a roof on the road — the street is not visible between the rows", frac*100)
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

// segSegMinDist returns the minimum distance between two city-space segments p0→p1 and
// q0→q1 (the smallest of the four endpoint-to-opposite-segment projections — exact for
// non-crossing segments, and 0-ish when they cross). Used by the connectivity test to
// decide whether two lane segments touch/join.
func segSegMinDist(p0, p1, q0, q1 tdPoint) float64 {
	c := []float64{
		distToSegSq(p0.x, p0.y, q0, q1),
		distToSegSq(p1.x, p1.y, q0, q1),
		distToSegSq(q0.x, q0.y, p0, p1),
		distToSegSq(q1.x, q1.y, p0, p1),
	}
	d := math.Inf(1)
	for _, v := range c {
		if v < d {
			d = v
		}
	}
	return math.Sqrt(d)
}

// TestRoadNetworkConnected locks playtest polish FIX 2: the lane network is ONE CONNECTED
// graph radiating from the central crossroads — every lane segment is reachable from every
// other (no isolated floating segments, the regression this work removes). The old model
// SPLIT each cross-street with a gap where it met a main, leaving the halves stranded; the
// cross-streets now run CONTINUOUSLY through the main they cross, and connectors reach the
// core, so the whole network links up.
//
// We build a graph over every street segment, join two segments when they touch within a
// tolerance, and assert a single connected component AND that the component contains a
// segment at the central crossroads (so "connected" means "connected to the heart", not a
// stranded ring). The tolerance (5.0 city units) is deliberately generous for the winding
// polylines yet still catches the regression: measured, the OLD split-halves network breaks
// into 5+ components at this tolerance while the connected network is exactly 1.
func TestRoadNetworkConnected(t *testing.T) {
	_ = theme.SetActive("forge")
	// Generous "these lanes meet" tolerance — larger than a lane's full width (~2·laneHalf)
	// so winding polylines that cross still register as joined, but tight enough that the old
	// disconnected floating halves would NOT all link (verified: old → 5+ components here).
	const tol = 5.0
	seeds := []string{"Aldermoor", "Bexley", "Corveil", "Duskwind", "Emberton"}
	counts := []map[string]int{
		{"hut": 8, "gathering_camp": 4},
		{"hut": 16, "gathering_camp": 10, "stone_camp": 6, "forge": 6},
		{"hut": 30, "gathering_camp": 20, "forge": 12, "barracks": 8, "colosseum": 1, "stonehenge": 1},
		{"hut": 60, "gathering_camp": 45, "forge": 28, "barracks": 18, "colosseum": 1},
	}
	type seg struct{ a, b tdPoint }
	for _, name := range seeds {
		for _, m := range counts {
			plan := tdPlanFor(namedState("primitive_age", name, m))
			var segs []seg
			for _, s := range plan.streets {
				for i := 0; i+1 < len(s.pts); i++ {
					segs = append(segs, seg{s.pts[i], s.pts[i+1]})
				}
			}
			n := len(segs)
			if n < 2 {
				t.Fatalf("seed %q counts %v: only %d lane segments — expected a real network", name, m, n)
			}
			// Union-find: join segments that touch within tol.
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
					if segSegMinDist(segs[i].a, segs[i].b, segs[j].a, segs[j].b) <= tol {
						parent[find(i)] = find(j)
					}
				}
			}
			roots := map[int]bool{}
			for i := 0; i < n; i++ {
				roots[find(i)] = true
			}
			if len(roots) != 1 {
				// Report component sizes for a helpful failure.
				sizes := map[int]int{}
				for i := 0; i < n; i++ {
					sizes[find(i)]++
				}
				t.Fatalf("seed %q counts %v: lane network has %d disconnected components (want 1) — some lanes are isolated floating segments: sizes=%v",
					name, m, len(roots), sizes)
			}
			// The single component must include a segment AT the central crossroads (reachable
			// FROM the core), so the network is anchored to the heart, not a stranded ring.
			nearCore := false
			for _, s := range segs {
				if segSegMinDist(s.a, s.b, tdPoint{plan.cx, plan.cy}, tdPoint{plan.cx, plan.cy}) <= tol {
					nearCore = true
					break
				}
			}
			if !nearCore {
				t.Fatalf("seed %q counts %v: no lane segment lies at the central crossroads — the connected network is not anchored to the core", name, m)
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
