package citymap

import (
	"image"
	"image/color"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// oneAgePerEra is a representative age key for each of the seven era bands, used
// to exercise every layout strategy. Keys are real ages from config so eraForAge
// classifies them by their canonical order.
var oneAgePerEra = []struct {
	name string
	age  string
	era  era
}{
	{"organic", "primitive_age", eraOrganic},
	{"hub-spoke", "bronze_age", eraHubSpoke},
	{"castle", "medieval_age", eraCastle},
	{"zoned-grid", "industrial_age", eraZonedGrid},
	{"city-blocks", "atomic_age", eraCityBlocks},
	{"campus", "digital_age", eraCampus},
	{"orbital", "galactic_age", eraOrbital},
}

// sampleBuilt is a mixed building set spanning several lineages plus a wonder and
// storage, so districts exercise the lineage-grouping and special categories.
func sampleBuilt() map[string]int {
	return map[string]int{
		"hut":            5,  // housing
		"gathering_camp": 12, // food
		"forge":          3,  // metallurgy
		"barracks":       2,  // military
		"library":        1,  // knowledge
		"granary":        4,  // storage (category)
		"stash":          2,  // storage (category)
	}
}

// TestEraForAgeCoversAllAges checks every real age classifies into a valid era
// band and that the named representatives land in the expected era.
func TestEraForAgeCoversAllAges(t *testing.T) {
	for _, k := range config.AgeOrder() {
		e := eraForAge(k)
		if e < eraOrganic || e > eraOrbital {
			t.Fatalf("age %q classified to out-of-range era %d", k, e)
		}
	}
	for _, tc := range oneAgePerEra {
		if got := eraForAge(tc.age); got != tc.era {
			t.Fatalf("eraForAge(%q) = %d, want %d", tc.age, got, tc.era)
		}
	}
	// Unknown age must default to organic, never panic.
	if got := eraForAge("made_up_age"); got != eraOrganic {
		t.Fatalf("eraForAge(unknown) = %d, want organic", got)
	}
}

// TestLayoutStrategiesInBoundsAndRoads verifies that every era's strategy returns
// placements strictly inside the canvas and a non-empty road network for a real
// building set. This is the core "the map means something" guarantee.
func TestLayoutStrategiesInBoundsAndRoads(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	const w, h = 80, 96 // 80 cols × 48 rows × 2 px
	byKey := config.BuildingByKey()
	pal := buildPalette(0)

	for _, tc := range oneAgePerEra {
		t.Run(tc.name, func(t *testing.T) {
			keys := sortedKeys(sampleBuilt())
			districts := builtDistricts(byKey, keys, sampleBuilt())
			if len(districts) == 0 {
				t.Fatal("expected non-empty districts for the sample building set")
			}
			seed := layoutSeed(tc.age, keys)
			res := buildLayout(tc.era, w, h, districts, pal, seed)

			if len(res.placements) == 0 {
				t.Fatal("no placements produced")
			}
			if len(res.roads) == 0 {
				t.Fatal("no roads produced — layout would be illegible")
			}
			// The palace must be present and centered.
			foundPalace := false
			for _, p := range res.placements {
				if p.tier == impPalace {
					foundPalace = true
					if p.cx != w/2 || p.cy != h/2 {
						t.Fatalf("palace at (%d,%d), want center (%d,%d)", p.cx, p.cy, w/2, h/2)
					}
				}
				if p.cx < 0 || p.cx >= w || p.cy < 0 || p.cy >= h {
					t.Fatalf("placement out of bounds: (%d,%d) in %dx%d", p.cx, p.cy, w, h)
				}
			}
			if !foundPalace {
				t.Fatal("no palace placement")
			}
			// Road endpoints must be in-bounds too (Bresenham clips, but the segment
			// data itself should be sane).
			for _, s := range res.roads {
				for _, pt := range [][2]int{{s.x0, s.y0}, {s.x1, s.y1}} {
					if pt[0] < 0 || pt[0] >= w || pt[1] < 0 || pt[1] >= h {
						t.Fatalf("road endpoint out of bounds: (%d,%d) in %dx%d", pt[0], pt[1], w, h)
					}
				}
			}
		})
	}
}

// TestLayoutDeterministic confirms the layout is stable frame-to-frame: the same
// (age, building set) yields identical placements and roads on repeated builds.
func TestLayoutDeterministic(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 80, 96
	byKey := config.BuildingByKey()
	pal := buildPalette(0)
	keys := sortedKeys(sampleBuilt())
	districts := builtDistricts(byKey, keys, sampleBuilt())
	seed := layoutSeed("bronze_age", keys)

	a := buildLayout(eraHubSpoke, w, h, districts, pal, seed)
	b := buildLayout(eraHubSpoke, w, h, districts, pal, seed)

	if len(a.placements) != len(b.placements) || len(a.roads) != len(b.roads) {
		t.Fatalf("nondeterministic counts: a=(%d,%d) b=(%d,%d)",
			len(a.placements), len(a.roads), len(b.placements), len(b.roads))
	}
	for i := range a.placements {
		if a.placements[i] != b.placements[i] {
			t.Fatalf("placement %d differs across builds: %+v vs %+v", i, a.placements[i], b.placements[i])
		}
	}
	for i := range a.roads {
		if a.roads[i] != b.roads[i] {
			t.Fatalf("road %d differs across builds", i)
		}
	}
}

// TestLineageColorStableDistinctAndThemeAware checks three properties of the
// lineage→color mapping: (1) it is stable for a given key under a fixed theme,
// (2) distinct lineages map to distinct colors, and (3) the same lineage retints
// when the theme changes.
func TestLineageColorStableDistinctAndThemeAware(t *testing.T) {
	lineages := []string{
		"food", "metallurgy", "military", "knowledge", "energy",
		"faith", "culture_arts", "harbor", "trade", "engineering",
	}

	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}

	// (1) stable.
	first := map[string]color.RGBA{}
	for _, lk := range lineages {
		first[lk] = lineageColor(lk, "production")
	}
	for _, lk := range lineages {
		if again := lineageColor(lk, "production"); again != first[lk] {
			t.Fatalf("lineageColor(%q) not stable: %v vs %v", lk, first[lk], again)
		}
	}

	// (2) distinct: no two lineages share a color under forge. (Allow a tiny
	// tolerance for HSL round-trip; require a meaningful channel difference.)
	for i := 0; i < len(lineages); i++ {
		for j := i + 1; j < len(lineages); j++ {
			a, b := first[lineages[i]], first[lineages[j]]
			if colorClose(a, b, 8) {
				t.Fatalf("lineages %q and %q map to near-identical colors %v / %v",
					lineages[i], lineages[j], a, b)
			}
		}
	}

	// (3) theme-aware: at least one lineage color must change under a different
	// theme (a recolor of the palette should move the role bases).
	if err := theme.SetActive("high_contrast"); err != nil {
		t.Fatalf("SetActive(high_contrast): %v", err)
	}
	changed := false
	for _, lk := range lineages {
		if lineageColor(lk, "production") != first[lk] {
			changed = true
			break
		}
	}
	_ = theme.SetActive("forge")
	if !changed {
		t.Fatal("no lineage color changed across themes — mapping is not theme-aware")
	}
}

// TestSpecialCategoryColors verifies wonders, storage, monuments, and diplomacy
// get their dedicated treatment rather than a lineage rotation.
func TestSpecialCategoryColors(t *testing.T) {
	_ = theme.SetActive("forge")
	wonder := lineageColor("wonder", "wonder")
	storage := lineageColor("storage", "storage")
	monument := lineageColor("anything", "monument")
	diplo := lineageColor("anything", "diplomacy")

	accent := rgba(theme.Color(theme.RoleAccent))
	dim := rgba(theme.Color(theme.RoleDim))

	if wonder != accent {
		t.Fatalf("wonder color = %v, want RoleAccent %v", wonder, accent)
	}
	if storage != dim {
		t.Fatalf("storage color = %v, want RoleDim %v", storage, dim)
	}
	// Monument is a brightened accent — should differ from raw accent and storage.
	if monument == accent || colorClose(monument, dim, 8) {
		t.Fatalf("monument color %v not distinct (accent %v / dim %v)", monument, accent, dim)
	}
	if colorClose(diplo, dim, 8) {
		t.Fatalf("diplomacy color %v collides with storage/dim %v", diplo, dim)
	}
}

// TestDrawVolumeProducesRoofWallShadow renders a single volume on a blank canvas
// and asserts the three 2.5D layers landed: a brightened roof pixel at the
// center, a darkened wall pixel below it, and a shadow pixel offset down-right.
func TestDrawVolumeProducesRoofWallShadow(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := buildPalette(0)
	img := image.NewRGBA(image.Rect(0, 0, 40, 40))
	// Paint the canvas a known base so we can detect the volume layers against it.
	base := color.RGBA{R: 10, G: 10, B: 10, A: 0xff}
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = base.R
		img.Pix[i+1] = base.G
		img.Pix[i+2] = base.B
		img.Pix[i+3] = 0xff
	}

	col := color.RGBA{R: 120, G: 80, B: 40, A: 0xff}
	p := placement{cx: 20, cy: 20, size: 1, col: col, tier: impNormal}
	drawVolume(img, pal, p)

	roof := brighten(col, 0.28)
	wall := darken(col, 0.35)

	// Roof at the center.
	if got := img.RGBAAt(20, 20); got != roof {
		t.Fatalf("roof pixel = %v, want %v", got, roof)
	}
	// Wall just below the roof footprint (cy + size + 1 = 22).
	if got := img.RGBAAt(20, 22); got != wall {
		t.Fatalf("wall pixel = %v, want %v", got, wall)
	}
	// Shadow offset down-right (cx+1, cy+size+2 = 23) must differ from base and roof.
	shadow := img.RGBAAt(21, 23)
	if shadow == base {
		t.Fatalf("shadow pixel unchanged (%v) — shadow not drawn", shadow)
	}
	if shadow == roof {
		t.Fatalf("shadow pixel equals roof %v — layering wrong", shadow)
	}
}

// TestDrawRoadBresenham checks the road rasterizer paints a connected line of the
// road color between two points, including both endpoints.
func TestDrawRoadBresenham(t *testing.T) {
	_ = theme.SetActive("forge")
	pal := buildPalette(0)
	img := image.NewRGBA(image.Rect(0, 0, 30, 30))
	drawRoad(img, roadSeg{2, 2, 20, 12}, pal.road)

	if got := img.RGBAAt(2, 2); got != pal.road {
		t.Fatalf("road start pixel = %v, want %v", got, pal.road)
	}
	if got := img.RGBAAt(20, 12); got != pal.road {
		t.Fatalf("road end pixel = %v, want %v", got, pal.road)
	}
	// Count painted pixels: a Bresenham line of this length must paint > max(dx,dy)
	// pixels (every step paints exactly one).
	painted := 0
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] == pal.road.R && img.Pix[i+1] == pal.road.G && img.Pix[i+2] == pal.road.B {
			painted++
		}
	}
	if painted < 18 { // dx = 18; line must be at least that many pixels
		t.Fatalf("road painted only %d pixels, want >= 18", painted)
	}
}

// TestRenderImageAllEras renders the full pipeline for one age per era band with a
// real building set, asserting a correct-size, fully-opaque, panic-free image.
func TestRenderImageAllEras(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	const w, h = 100, 60
	for _, tc := range oneAgePerEra {
		t.Run(tc.name, func(t *testing.T) {
			img := renderImage(sampleState(tc.age, sampleBuilt()), w, h)
			if img == nil {
				t.Fatal("renderImage returned nil")
			}
			if got := img.Bounds(); got.Dx() != w || got.Dy() != h {
				t.Fatalf("image size = %dx%d, want %dx%d", got.Dx(), got.Dy(), w, h)
			}
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xff {
					t.Fatalf("pixel %d alpha = %d, want 255", i/4, img.Pix[i])
				}
			}
		})
	}
}

// --- helpers ---------------------------------------------------------------

// sortedKeys returns the sorted keys of a built map (matches builtBuildingKeys
// ordering so districts are built identically to the render path).
func sortedKeys(built map[string]int) []string {
	keys := make([]string, 0, len(built))
	for k := range built {
		keys = append(keys, k)
	}
	// Use the same ordering as the renderer (sort.Strings).
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// colorClose reports whether two colors are within tol on every channel.
func colorClose(a, b color.RGBA, tol int) bool {
	d := func(x, y uint8) int {
		v := int(x) - int(y)
		if v < 0 {
			return -v
		}
		return v
	}
	return d(a.R, b.R) <= tol && d(a.G, b.G) <= tol && d(a.B, b.B) <= tol
}
