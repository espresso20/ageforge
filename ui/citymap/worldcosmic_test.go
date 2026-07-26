package citymap

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// worldcosmic_test.go covers Phase C — the COSMIC STRATEGIC VIEWS (worldcosmic.go): that the
// dispatcher intercepts only the wired cosmic ages, that the space_age strategic view is
// panic-safe / deterministic across sizes, that it does NOT render the terrestrial continent
// (it's a strategy star-map, not a map), and it carries the opt-in Phase-C PNG dump.

// cosmicWorldAge builds a cosmic-age fixture with a mixed faction roster: several discovered civs
// of varying Strength, one at war, one trade partner, plus one undiscovered — so every wired
// cosmic view exercises nodes/systems/capitals + lanes/web/seams + territory + fog. The roster is
// identical across ages (only state.Age differs), so the same seeded world renders three ways.
func cosmicWorldAge(age string) game.GameState {
	facs := map[string]game.FactionInfo{
		"helios":   {Name: "Helios Combine", Discovered: true, Status: "allied", Opinion: 75, Strength: 4, Personality: "peaceful"},
		"vega":     {Name: "Vega Syndicate", Discovered: true, Status: "friendly", Opinion: 30, Strength: 3, Personality: "mercantile", TradeCount: 6},
		"drakon":   {Name: "Drakon Hegemony", Discovered: true, Status: "rival", Opinion: -55, Strength: 5, AtWar: true, Personality: "aggressive"},
		"tessarae": {Name: "Tessarae", Discovered: true, Status: "neutral", Opinion: 5, Strength: 2, Personality: "isolationist"},
		"unknown":  {Name: "Unknown Signal", Discovered: false, Strength: 3},
	}
	blds := map[string]int{"orbital_dock": 8, "fusion_plant": 5, "arcology": 3}
	return worldState(age, blds, "Aldermoor", facs)
}

// cosmicWorld is the space_age fixture (kept for the existing space-specific tests).
func cosmicWorld() game.GameState { return cosmicWorldAge("space_age") }

// cosmicWiredAges lists the ages the Phase-C dispatcher owns — the panic/determinism/not-continent
// sweeps run every one of them.
var cosmicWiredAges = []string{"space_age", "interstellar_age", "galactic_age"}

// TestCosmicWorldViewForDispatch pins the intercept: the three wired cosmic ages are owned by a
// strategic view; terrestrial ages, the still-unwired higher cosmic ages (quantum/transcendent),
// and unknown ages are not (they fall through to the medium path).
func TestCosmicWorldViewForDispatch(t *testing.T) {
	for _, age := range cosmicWiredAges {
		if _, ok := cosmicWorldViewFor(age); !ok {
			t.Errorf("cosmicWorldViewFor(%q) = false, want true", age)
		}
	}
	notCosmic := []string{
		"primitive_age", "modern_age", "cyberpunk_age", "fusion_age",
		// quantum + transcendent are NOT wired yet — they must still fall through to the atlas.
		"quantum_age", "transcendent_age",
		"made_up_age", "",
	}
	for _, age := range notCosmic {
		if _, ok := cosmicWorldViewFor(age); ok {
			t.Errorf("cosmicWorldViewFor(%q) = true, want false", age)
		}
	}
}

// TestCosmicRenderPanicSafe drives EVERY wired cosmic strategic view (via renderWorldImage) across
// tiny / odd / zero / large canvases with a populated roster. None may panic, and each must return
// an image of the exact requested pixel size.
func TestCosmicRenderPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	sizes := []struct{ w, h int }{
		{0, 0}, {1, 1}, {1, 40}, {40, 1}, {3, 7}, {7, 3},
		{2, 2}, {13, 9}, {80, 40}, {400, 260}, {123, 57},
	}
	for _, age := range cosmicWiredAges {
		st := cosmicWorldAge(age)
		for _, sz := range sizes {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("renderWorldImage(%s, %d, %d) panicked: %v", age, sz.w, sz.h, r)
					}
				}()
				img, _ := renderWorldImage(st, sz.w, sz.h)
				if img == nil {
					t.Fatalf("renderWorldImage(%s, %d,%d) returned nil", age, sz.w, sz.h)
				}
				b := img.Bounds()
				if b.Dx() != sz.w || b.Dy() != sz.h {
					t.Errorf("renderWorldImage(%s, %d,%d) image is %dx%d", age, sz.w, sz.h, b.Dx(), b.Dy())
				}
			}()
		}
	}
}

// TestCosmicEmptyRosterOK checks the empty case: no discovered factions still renders (the hub
// alone in the starfield) without panicking and with a non-degenerate image.
func TestCosmicEmptyRosterOK(t *testing.T) {
	_ = theme.SetActive("forge")
	st := worldState("space_age", map[string]int{"arcology": 2}, "Solo", nil)
	img, plan := renderWorldImage(st, 120, 80)
	if img == nil {
		t.Fatal("renderWorldImage(empty roster) returned nil")
	}
	// The cosmic view owns the whole frame → the overlay plan is empty (no labels).
	if len(plan.labels) != 0 {
		t.Errorf("cosmic view should emit no overlay labels, got %d", len(plan.labels))
	}
	// It should have drawn SOMETHING beyond a flat fill (the hub + starfield).
	if uniformImage(img) {
		t.Error("cosmic view with a lone hub rendered a uniform image")
	}
}

// TestCosmicDeterministic asserts byte-for-byte stability for every wired cosmic view: the same
// (state, size, seed) renders identically twice (hash/seed only, no math/rand).
func TestCosmicDeterministic(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 240, 150
	for _, age := range cosmicWiredAges {
		st := cosmicWorldAge(age)
		a, _ := renderWorldImage(st, w, h)
		b, _ := renderWorldImage(st, w, h)
		if !bytes.Equal(a.Pix, b.Pix) {
			t.Errorf("%s strategic view is not deterministic: two renders differ", age)
		}
	}
}

// TestCosmicNotAContinent is the whole point: every cosmic strategic view must NOT look like the
// terrestrial atlas/continent render. We render a terrestrial age (atlas continent) and each wired
// cosmic age off the SAME display name (same world seed family) and assert the pixels differ
// substantially — the cosmic views are strategy star-maps, not maps of land.
func TestCosmicNotAContinent(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 240, 150
	// A roster with an ally, an at-war rival, and a trade partner so the new views exercise
	// territory + seams + lanes too. Same display name across ages → same underlying world seed.
	facs := map[string]game.FactionInfo{
		"helios": {Name: "Helios", Discovered: true, Status: "allied", Opinion: 70, Strength: 4},
		"drakon": {Name: "Drakon", Discovered: true, Status: "rival", Opinion: -50, Strength: 5, AtWar: true, Personality: "aggressive"},
		"vega":   {Name: "Vega", Discovered: true, Status: "friendly", Opinion: 20, Strength: 3, Personality: "mercantile", TradeCount: 4},
	}
	blds := map[string]int{"arcology": 4}
	atlas := worldState("modern_age", blds, "Aldermoor", facs) // satellite/atlas terrestrial render
	aImg, _ := renderWorldImage(atlas, w, h)

	for _, age := range cosmicWiredAges {
		cosmic := worldState(age, blds, "Aldermoor", facs)
		cImg, _ := renderWorldImage(cosmic, w, h)
		if bytes.Equal(cImg.Pix, aImg.Pix) {
			t.Fatalf("%s strategic view is byte-identical to the terrestrial render", age)
		}
		if diff := diffFraction(cImg, aImg); diff < 0.5 {
			t.Errorf("%s view differs from terrestrial render in only %.1f%% of pixels; expected a wholly different image", age, diff*100)
		}
	}
}

// TestCosmicViewsDistinct is the coherent-but-DISTINCT guarantee: for one fixed (state, size,
// seed), the three cosmic views must not read near-identical. We call each view function directly
// with the SAME seed (isolating the difference to the rendering, not the per-age seed) and assert
// every pair differs in a meaningful fraction of pixels — the dominant metaphor (home cluster vs
// route web vs territorial spiral) has to change the picture, not just the palette.
func TestCosmicViewsDistinct(t *testing.T) {
	_ = theme.SetActive("forge")
	st := cosmicWorld()
	const w, h = 240, 150
	const seed uint32 = 0xC0FFEE
	render := func(fn func(*image.RGBA, game.GameState, int, int, uint32)) *image.RGBA {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		fn(img, st, w, h, seed)
		return img
	}
	space := render(drawSpaceStrategicView)
	inter := render(interstellarStrategicView)
	galac := render(galacticStrategicView)
	pairs := []struct {
		name string
		a, b *image.RGBA
	}{
		{"space vs interstellar", space, inter},
		{"space vs galactic", space, galac},
		{"interstellar vs galactic", inter, galac},
	}
	for _, pr := range pairs {
		if bytes.Equal(pr.a.Pix, pr.b.Pix) {
			t.Errorf("%s: renders are byte-identical", pr.name)
			continue
		}
		if diff := diffFraction(pr.a, pr.b); diff < 0.10 {
			t.Errorf("%s: renders differ in only %.1f%% of pixels; the dominant metaphor should make them read distinct", pr.name, diff*100)
		}
	}
}

// TestCosmicEmpireLayout sanity-checks the reusable layout helper: the home node is first,
// flagged isHome, near center; discovered factions get distinct positions inside the frame;
// undiscovered factions are NOT placed here.
func TestCosmicEmpireLayout(t *testing.T) {
	_ = theme.SetActive("forge")
	st := cosmicWorld()
	const w, h = 200, 120
	nodes := strategicEmpires(st, w, h, 0xABCD)
	if len(nodes) == 0 {
		t.Fatal("strategicEmpires returned no nodes")
	}
	home := nodes[0]
	if !home.isHome {
		t.Error("first node is not the home hub")
	}
	// 4 discovered factions in the fixture → 1 home + 4 = 5 nodes (undiscovered excluded).
	if len(nodes) != 5 {
		t.Errorf("expected 5 nodes (home + 4 discovered), got %d", len(nodes))
	}
	// Home near center.
	if abs(home.cx-w/2) > w/4 || abs(home.cy-h/2) > h/4 {
		t.Errorf("home hub at (%d,%d) not near center (%d,%d)", home.cx, home.cy, w/2, h/2)
	}
	// Faction nodes: on-canvas and not stacked on the hub.
	seen := map[[2]int]bool{}
	for _, n := range nodes[1:] {
		if n.cx < 0 || n.cx >= w || n.cy < 0 || n.cy >= h {
			t.Errorf("faction node %q off-canvas at (%d,%d)", n.key, n.cx, n.cy)
		}
		if n.cx == home.cx && n.cy == home.cy {
			t.Errorf("faction node %q stacked on the hub", n.key)
		}
		seen[[2]int{n.cx, n.cy}] = true
	}
	if len(seen) < 2 {
		t.Error("faction nodes collapsed to a single position")
	}
}

// TestCosmicLayoutDeterministic confirms node placement is stable for the same (state,size,seed).
func TestCosmicLayoutDeterministic(t *testing.T) {
	_ = theme.SetActive("forge")
	st := cosmicWorld()
	a := strategicEmpires(st, 200, 120, 7)
	b := strategicEmpires(st, 200, 120, 7)
	if len(a) != len(b) {
		t.Fatalf("layout length differs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].cx != b[i].cx || a[i].cy != b[i].cy || a[i].role != b[i].role {
			t.Errorf("node %d differs between runs: %+v vs %+v", i, a[i], b[i])
		}
	}
}

// TestDumpWorldCosmicPNG writes the Phase-C proof-of-concept PNGs — one per wired cosmic view.
// Skipped unless CITYMAP_PNG_DUMP=<dir> is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDump -count=1
func TestDumpWorldCosmicPNG(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump the Phase-C cosmic strategic PNGs")
	}
	_ = theme.SetActive("forge")
	dumps := []struct{ age, file string }{
		{"space_age", "wm_C_space.png"},
		{"interstellar_age", "wm_C_interstellar.png"},
		{"galactic_age", "wm_C_galactic.png"},
	}
	for _, d := range dumps {
		st := cosmicWorldAge(d.age)
		img, _ := renderWorldImage(st, 400, 260)
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
		t.Logf("wrote %s (%s strategic view)", path, d.age)
	}
}

// ---- small test helpers -----------------------------------------------------

// uniformImage reports whether every pixel in img is identical (a flat fill).
func uniformImage(img *image.RGBA) bool {
	if len(img.Pix) < 8 {
		return true
	}
	first := img.RGBAAt(img.Bounds().Min.X, img.Bounds().Min.Y)
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y) != first {
				return false
			}
		}
	}
	return true
}

// diffFraction returns the fraction of pixels that differ between two equal-sized images.
func diffFraction(a, b *image.RGBA) float64 {
	if len(a.Pix) != len(b.Pix) || len(a.Pix) == 0 {
		return 1.0
	}
	diff := 0
	total := len(a.Pix) / 4
	for i := 0; i < len(a.Pix); i += 4 {
		if a.Pix[i] != b.Pix[i] || a.Pix[i+1] != b.Pix[i+1] ||
			a.Pix[i+2] != b.Pix[i+2] || a.Pix[i+3] != b.Pix[i+3] {
			diff++
		}
	}
	if total == 0 {
		return 1.0
	}
	return float64(diff) / float64(total)
}

// abs is a tiny int abs for the layout test.
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
