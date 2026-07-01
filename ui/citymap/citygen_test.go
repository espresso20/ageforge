package citymap

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// citygen_test.go covers the count-driven city synthesizer (citygen.go): the
// framework guarantees from design-and-architecture/city-synthesis.md §Pipeline —
// determinism, count→population, terrain gating (no lot or street on water), the
// organic/ancient era styles actually differing, landmark labels feeding the overlay
// geometry, panic-safety on degenerate canvases, and correct output size.

// lakeField builds a terrainField that is open land except a big central lake, so
// tests can assert nothing (lots or streets) lands on the water. Everything outside
// the lake rect is passable grass.
func lakeField(w, h, lx0, ly0, lx1, ly1 int) *terrainField {
	f := &terrainField{w: w, h: h, biomes: make([]biome, w*h), passable: make([]bool, w*h)}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if x >= lx0 && x <= lx1 && y >= ly0 && y <= ly1 {
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

// TestGenerateCityPlanDeterministic: the same seed yields a byte-identical plan.
func TestGenerateCityPlanDeterministic(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 90, 100
	byKey := config.BuildingByKey()
	st := sampleState("bronze_age", sampleBuilt())
	seed := layoutSeed("bronze_age", sortedKeys(sampleBuilt()))

	a := generateCityPlan(st, byKey, nil, eraHubSpoke, seed, w, h)
	b := generateCityPlan(st, byKey, nil, eraHubSpoke, seed, w, h)
	if !cityPlansEqual(a, b) {
		t.Fatal("same seed produced different plans — synthesis is not deterministic")
	}

	// A different seed should generally produce a different plan (sanity: the seed is
	// actually threaded through the generator, not ignored).
	c := generateCityPlan(st, byKey, nil, eraHubSpoke, seed+0x9e3779b9, w, h)
	if cityPlansEqual(a, c) {
		t.Fatal("different seeds produced identical plans — seed is not threaded")
	}
}

// TestCountDrivesPopulation: strictly more housing instances must yield strictly more
// house lots (the population tracks COUNT, not distinct-type). Same era/seed/size so
// only the count varies.
func TestCountDrivesPopulation(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 120
	byKey := config.BuildingByKey()
	seed := uint32(0x1234abcd)

	houseCount := func(built map[string]int) int {
		st := sampleState("primitive_age", built)
		plan := generateCityPlan(st, byKey, nil, eraOrganic, seed, w, h)
		n := 0
		for _, lt := range plan.lots {
			if lt.kind == lotHouse {
				n++
			}
		}
		return n
	}

	few := houseCount(map[string]int{"hut": 2})
	many := houseCount(map[string]int{"hut": 30})
	if few == 0 {
		t.Fatal("a 2-hut village produced zero house lots")
	}
	if many <= few {
		t.Fatalf("more housing instances did not yield more houses: 2 huts → %d, 30 huts → %d", few, many)
	}
}

// TestEveryLotOnPassableLand: with a real terrain field, NO lot may land on water.
// This is the §terrain-gate guarantee — every lot snapped onto passable ground.
func TestEveryLotOnPassableLand(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 120
	byKey := config.BuildingByKey()
	// Use the real render field so the test reflects the shipping terrain, plus a hard
	// central-lake field so the gate is definitely exercised.
	seed, _ := ageInfo("bronze_age")
	realField := newTerrainField(w, h, seed)
	lake := lakeField(w, h, 35, 45, 65, 75)

	st := sampleState("bronze_age", sampleBuilt())
	for _, tc := range []struct {
		name string
		f    *terrainField
	}{
		{"real-terrain", realField},
		{"central-lake", lake},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := generateCityPlan(st, byKey, tc.f, eraHubSpoke, layoutSeed("bronze_age", sortedKeys(sampleBuilt())), w, h)
			for _, lt := range plan.lots {
				// Judge by the lot's center, which is what gateLots snaps.
				cx := lt.x + lt.w/2
				cy := lt.y + lt.h/2
				if !tc.f.passableAt(cx, cy) {
					t.Fatalf("%s lot (kind %d) center (%d,%d) on impassable cell", tc.name, lt.kind, cx, cy)
				}
			}
		})
	}
}

// TestNoStreetPixelOnWater: no street polyline may pass through an impassable cell.
// Streets are terrain-routed via the cost grid, so every waypoint — and the whole
// rasterized line between waypoints — must stay on passable land.
func TestNoStreetPixelOnWater(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 120
	byKey := config.BuildingByKey()
	f := lakeField(w, h, 30, 40, 70, 80)
	st := sampleState("bronze_age", sampleBuilt())

	plan := generateCityPlan(st, byKey, f, eraHubSpoke, layoutSeed("bronze_age", sortedKeys(sampleBuilt())), w, h)
	for si, s := range plan.streets {
		for i := 0; i+1 < len(s.pts); i++ {
			a, b := s.pts[i], s.pts[i+1]
			// Walk the segment at ~1px resolution and assert every sample is on land.
			steps := absInt(b.x-a.x) + absInt(b.y-a.y)
			if steps == 0 {
				if !f.passableAt(a.x, a.y) {
					t.Fatalf("street %d waypoint (%d,%d) on water", si, a.x, a.y)
				}
				continue
			}
			for k := 0; k <= steps; k++ {
				x := a.x + (b.x-a.x)*k/steps
				y := a.y + (b.y-a.y)*k/steps
				if !f.passableAt(x, y) {
					t.Fatalf("street %d crosses water at (%d,%d) (segment (%d,%d)->(%d,%d))",
						si, x, y, a.x, a.y, b.x, b.y)
				}
			}
		}
	}
}

// TestOrganicVsAncientDiffer: the two tuned eras must produce visibly different
// plans — the eraStyle is actually applied, not ignored. Organic is a handful of
// meandering paths with loose blocks; ancient is radial avenues + a coarse grid, so
// the street count and block presence differ. Same state/seed isolates the era.
func TestOrganicVsAncientDiffer(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 120
	byKey := config.BuildingByKey()
	st := sampleState("bronze_age", sampleBuilt()) // age irrelevant here; era is passed explicitly
	seed := layoutSeed("bronze_age", sortedKeys(sampleBuilt()))

	organic := generateCityPlan(st, byKey, nil, eraOrganic, seed, w, h)
	ancient := generateCityPlan(st, byKey, nil, eraHubSpoke, seed, w, h)

	// Ancient (radial + grid) lays down many more streets than the village's few paths.
	if len(ancient.streets) <= len(organic.streets) {
		t.Fatalf("ancient should have more streets than organic: ancient=%d organic=%d",
			len(ancient.streets), len(organic.streets))
	}
	// Ancient carves a grid of blocks; organic has only a few loose cluster blocks.
	if len(ancient.blocks) <= len(organic.blocks) {
		t.Fatalf("ancient should have more blocks than organic: ancient=%d organic=%d",
			len(ancient.blocks), len(organic.blocks))
	}
	// The ancient era sets hasPlaza — a plaza lot must be present; the village has none.
	ancientHasPlaza, organicHasPlaza := false, false
	for _, lt := range ancient.lots {
		if lt.kind == lotPlaza {
			ancientHasPlaza = true
		}
	}
	for _, lt := range organic.lots {
		if lt.kind == lotPlaza {
			organicHasPlaza = true
		}
	}
	if !ancientHasPlaza {
		t.Fatal("ancient era should place a central plaza")
	}
	if organicHasPlaza {
		t.Fatal("organic era should not place a plaza")
	}
}

// TestLandmarksCarryLabelsAndFeedOverlay: landmark lots carry the building name, and
// after drawing the plan those names surface in the layoutGeometry the overlay reads
// (one buildingLabel per landmark, with its lineage carried for the label color).
func TestLandmarksCarryLabelsAndFeedOverlay(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 100, 120
	byKey := config.BuildingByKey()
	st := sampleState("bronze_age", sampleBuilt())
	seed := layoutSeed("bronze_age", sortedKeys(sampleBuilt()))

	plan := generateCityPlan(st, byKey, nil, eraHubSpoke, seed, w, h)

	// Collect the landmark names from the plan.
	planLandmarks := map[string]bool{}
	for _, lt := range plan.lots {
		if lt.kind == lotLandmark {
			if lt.label == "" {
				t.Fatalf("landmark lot has no label: %+v", lt)
			}
			planLandmarks[lt.label] = true
		}
	}
	if len(planLandmarks) == 0 {
		t.Fatal("no landmark lots — nothing to label")
	}
	if !planLandmarks["Library"] {
		t.Fatalf("expected Library landmark, have %v", planLandmarks)
	}

	// Draw the plan and confirm the geometry the overlay consumes carries every
	// landmark as a named building anchor with a non-empty lineage.
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	geo := drawCityPlan(img, buildPalette(0), plan, w, h)
	geoNames := map[string]string{}
	for _, bl := range geo.buildings {
		if bl.name == "" {
			t.Fatalf("overlay building anchor has empty name: %+v", bl)
		}
		if bl.lineageKey == "" {
			t.Fatalf("overlay building anchor %q missing lineage", bl.name)
		}
		geoNames[bl.name] = bl.lineageKey
	}
	for name := range planLandmarks {
		if _, ok := geoNames[name]; !ok {
			t.Fatalf("landmark %q did not reach the overlay geometry (have %v)", name, geoNames)
		}
	}

	// The geometry must also feed the real overlay plan as building labels.
	oplan := buildOverlayPlan(st, w, h/2, geo)
	labeled := map[string]bool{}
	for _, lb := range oplan.labels {
		if lb.kind == labelBuilding {
			labeled[lb.text] = true
		}
	}
	if !labeled["Library"] {
		t.Fatalf("Library landmark not labeled by the overlay; have %v", labeled)
	}
}

// TestGenerateCityPlanPanicSafe: degenerate canvases (zero, tiny, 1px) must not
// panic and must return a well-formed (possibly empty) plan.
func TestGenerateCityPlanPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	byKey := config.BuildingByKey()
	st := sampleState("primitive_age", map[string]int{"hut": 3, "gathering_camp": 2, "library": 1})
	f := func(w, h int) *terrainField { return newTerrainField(w, h, 7) }

	for _, sz := range [][2]int{{0, 0}, {1, 1}, {2, 2}, {3, 5}, {1, 40}, {40, 1}} {
		w, hh := sz[0], sz[1]
		// nil field and a real field, both must be safe.
		_ = generateCityPlan(st, byKey, nil, eraOrganic, 12345, w, hh)
		_ = generateCityPlan(st, byKey, f(w, hh), eraHubSpoke, 12345, w, hh)
	}
}

// TestDrawCityPlanCorrectSize: drawing a plan preserves the image dimensions and
// full opacity, and does not panic on a real synthesized plan (the render contract
// renderImage relies on).
func TestDrawCityPlanCorrectSize(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 80, 96
	byKey := config.BuildingByKey()
	st := sampleState("bronze_age", sampleBuilt())
	seed := layoutSeed("bronze_age", sortedKeys(sampleBuilt()))
	field := newTerrainField(w, h, 42)
	plan := generateCityPlan(st, byKey, field, eraHubSpoke, seed, w, h)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill opaque first (renderImage paints terrain before structures).
	drawTerrainField(img, buildPalette(0), field)
	geo := drawCityPlan(img, buildPalette(0), plan, w, h)

	if got := img.Bounds(); got.Dx() != w || got.Dy() != h {
		t.Fatalf("image resized to %dx%d, want %dx%d", got.Dx(), got.Dy(), w, h)
	}
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0xff {
			t.Fatalf("pixel %d alpha = %d, want 255", i/4, img.Pix[i])
		}
	}
	// City center anchor must be in-bounds so the overlay's "City Center" label lands.
	if geo.palaceX < 0 || geo.palaceX >= w || geo.palaceY < 0 || geo.palaceY >= h {
		t.Fatalf("city center anchor (%d,%d) out of bounds", geo.palaceX, geo.palaceY)
	}
}
