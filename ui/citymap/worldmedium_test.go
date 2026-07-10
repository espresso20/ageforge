package citymap

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// worldmedium_test.go covers Phase B — the cartographic medium abstraction (worldmedium.go):
// that mediumForAge wires the four validation ages to their bespoke mediums (and everything
// else to the default atlas), that each medium render is panic-safe / exact-size /
// deterministic, and — the whole point — that the four mediums read VISIBLY different from
// each other and from the default. It also carries the opt-in Phase-B PNG dump.

// mediumWorld builds the standard fixture: the SAME display name (so the same continent) at
// a given age, with a small discovered-civ roster so the overlay shows up on the dump.
func mediumWorld(age string) game.GameState {
	facs := map[string]game.FactionInfo{
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 70, Strength: 4},
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -50, AtWar: true, Strength: 5},
		"egypt":    {Name: "Egypt", Discovered: true, Status: "neutral", Opinion: 5, Strength: 2},
	}
	blds := map[string]int{"forge": 12, "barracks": 6, "granary": 4}
	return worldState(age, blds, "Aldermoor", facs)
}

// TestMediumForAgeWiring pins the age→medium mapping: the four validation ages resolve to
// their bespoke styles, and a couple of other ages fall through to the default atlas.
func TestMediumForAgeWiring(t *testing.T) {
	_ = theme.SetActive("forge")
	cases := []struct {
		age   string
		name  string
		style worldMediumStyle
	}{
		{"primitive_age", "charcoal", styleCharcoal},
		{"medieval_age", "parchment", styleParchment},
		{"modern_age", "satellite", styleSatellite},
		{"cyberpunk_age", "neon", styleNeon},
		// A spread of other ages must all be the neutral atlas default.
		{"classical_age", "atlas", styleAtlas},
		{"iron_age", "atlas", styleAtlas},
		{"stone_age", "atlas", styleAtlas},
		{"transcendent_age", "atlas", styleAtlas},
		{"made_up_age", "atlas", styleAtlas}, // unknown ages default too
	}
	for _, tc := range cases {
		med := mediumForAge(tc.age)
		if med.name != tc.name {
			t.Errorf("mediumForAge(%q).name = %q, want %q", tc.age, med.name, tc.name)
		}
		if med.style != tc.style {
			t.Errorf("mediumForAge(%q).style = %d, want %d", tc.age, med.style, tc.style)
		}
		if med.draw == nil {
			t.Errorf("mediumForAge(%q) has nil draw func", tc.age)
		}
	}
}

// TestMediumRenderSizeAndPanicSafe drives every medium's full render (via renderWorldImage,
// which selects the medium from the age) across a spread of sizes including tiny/odd/zero.
// Each must produce an image of EXACTLY the requested pixel size, fully opaque, and never
// panic — the chrome (frames, compass, HUD brackets) and per-pixel passes must all clip.
func TestMediumRenderSizeAndPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	ages := []string{"primitive_age", "medieval_age", "modern_age", "cyberpunk_age", "classical_age"}
	sizes := []struct{ w, h int }{{0, 0}, {1, 1}, {1, 2}, {3, 5}, {7, 4}, {13, 9}, {40, 26}, {200, 130}}
	for _, age := range ages {
		for _, sz := range sizes {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s render panicked at %dx%d: %v", age, sz.w, sz.h, r)
					}
				}()
				st := mediumWorld(age)
				img, _ := renderWorldImage(st, sz.w, sz.h)
				if img == nil {
					t.Fatalf("%s: nil image at %dx%d", age, sz.w, sz.h)
				}
				ew, eh := sz.w, sz.h
				if ew < 0 {
					ew = 0
				}
				if eh < 0 {
					eh = 0
				}
				if got := img.Bounds(); got.Dx() != ew || got.Dy() != eh {
					t.Fatalf("%s: size = %dx%d, want %dx%d", age, got.Dx(), got.Dy(), ew, eh)
				}
				for i := 3; i < len(img.Pix); i += 4 {
					if img.Pix[i] != 0xff {
						t.Fatalf("%s: pixel %d alpha = %d, want 255 at %dx%d", age, i/4, img.Pix[i], sz.w, sz.h)
					}
				}
			}()
		}
	}
}

// TestMediumRenderDeterministic proves each medium is a pure function of its inputs: two
// renders of the same (age, name, size) are byte-identical. Determinism underpins the cache
// key in imageFor.
func TestMediumRenderDeterministic(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 160, 100
	for _, age := range []string{"primitive_age", "medieval_age", "modern_age", "cyberpunk_age"} {
		a, _ := renderWorldImage(mediumWorld(age), w, h)
		b, _ := renderWorldImage(mediumWorld(age), w, h)
		if imagesDiffer(a, b) {
			t.Fatalf("%s render is not deterministic — two identical inputs produced different images", age)
		}
	}
}

// TestMediumsReadDistinct is the whole point of Phase B: the four bespoke mediums must render
// the SAME continent (same display name, same size) VISIBLY differently from one another AND
// from the default atlas. We render all five and assert every pair differs. (imagesDiffer is
// in citymap_test.go, same package.)
func TestMediumsReadDistinct(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 200, 130
	type shot struct {
		age string
		img *image.RGBA
	}
	shots := []shot{}
	for _, age := range []string{"primitive_age", "medieval_age", "modern_age", "cyberpunk_age", "classical_age"} {
		img, _ := renderWorldImage(mediumWorld(age), w, h)
		shots = append(shots, shot{age, img})
	}
	for i := 0; i < len(shots); i++ {
		for j := i + 1; j < len(shots); j++ {
			if !imagesDiffer(shots[i].img, shots[j].img) {
				t.Errorf("mediums %q and %q render IDENTICAL — they must read distinct", shots[i].age, shots[j].age)
			}
		}
	}
}

// TestMediumsDrawTerrainSelfContained sanity-checks each bespoke medium's draw function in
// isolation (no overlay): it must paint the terrain layer without panicking on a real model
// and leave the image fully opaque. Guards that a medium can't leave transparent gaps that
// the half-block streamer would render as black holes.
func TestMediumsDrawTerrainSelfContained(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 120, 80
	m := buildWorldModel(w, h, worldTerrainSeed("Aldermoor"))
	for _, age := range []string{"primitive_age", "medieval_age", "modern_age", "cyberpunk_age", "classical_age"} {
		med := mediumForAge(age)
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s draw panicked: %v", age, r)
				}
			}()
			med.draw(img, m, med)
		}()
		for i := 3; i < len(img.Pix); i += 4 {
			if img.Pix[i] != 0xff {
				t.Fatalf("%s: draw left pixel %d transparent (alpha %d)", age, i/4, img.Pix[i])
			}
		}
	}
}

// TestDumpWorldMediumPNGs renders the WORLD tab for the four medium ages + one default, all
// with the SAME display-name seed (the same continent re-skinned), at 400×260px. Opt-in:
// skipped unless CITYMAP_PNG_DUMP=<dir> is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpWorldMediumPNGs -count=1
func TestDumpWorldMediumPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump the Phase-B medium PNGs")
	}
	_ = theme.SetActive("forge")
	dumps := []struct {
		age  string
		file string
	}{
		{"primitive_age", "wm_B_charcoal.png"},
		{"medieval_age", "wm_B_parchment.png"},
		{"modern_age", "wm_B_satellite.png"},
		{"cyberpunk_age", "wm_B_neon.png"},
		{"classical_age", "wm_B_default.png"},
	}
	for _, d := range dumps {
		st := mediumWorld(d.age)
		// 400×260px: renderWorldImage takes cell dims and doubles height → 400 cols × 130 rows.
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
		t.Logf("wrote %s (age=%s, medium=%s)", path, d.age, mediumForAge(d.age).name)
	}
}
