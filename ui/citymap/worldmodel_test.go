package citymap

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// worldmodel_test.go covers the seeded continent model (worldmodel.go): determinism,
// that it's a REAL continent (land + ocean, one central connected mass, not edge-to-edge),
// that rivers flow downhill and reach the sea, and the neutral render's size/panic-safety.
// It also carries the opt-in PNG dump for eyeballing whole continents.

// modelLandWater counts passable (land) vs impassable (water/mountain) cells in a model.
func modelLandWater(m *worldModel) (land, water int) {
	for _, p := range m.field.passable {
		if p {
			land++
		} else {
			water++
		}
	}
	return land, water
}

// largestLandComponent returns the size of the biggest 4-connected land region and the
// total land count, via flood fill — used to assert the continent is a coherent central
// mass rather than confetti scattered across the frame.
func largestLandComponent(m *worldModel) (largest, total int) {
	w, h := m.w, m.h
	seen := make([]bool, w*h)
	for i, p := range m.field.passable {
		if p {
			total++
		}
		_ = i
	}
	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			si := sy*w + sx
			if !m.field.passable[si] || seen[si] {
				continue
			}
			stack := []int{si}
			seen[si] = true
			sz := 0
			for len(stack) > 0 {
				c := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				sz++
				cx, cy := c%w, c/w
				for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					nx, ny := cx+nb[0], cy+nb[1]
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					ni := ny*w + nx
					if m.field.passable[ni] && !seen[ni] {
						seen[ni] = true
						stack = append(stack, ni)
					}
				}
			}
			if sz > largest {
				largest = sz
			}
		}
	}
	return largest, total
}

// TestBuildWorldModelDeterministic proves the model is a pure function of (w,h,seed):
// two builds with the same inputs produce identical elevation, biomes, passability,
// river count, and relief count. (Determinism is the contract the whole two-seed
// age-stability design rests on.)
func TestBuildWorldModelDeterministic(t *testing.T) {
	const w, h = 160, 100
	seed := worldTerrainSeed("Rome")
	a := buildWorldModel(w, h, seed)
	b := buildWorldModel(w, h, seed)

	if len(a.elev) != len(b.elev) {
		t.Fatalf("elevation grid sizes differ: %d vs %d", len(a.elev), len(b.elev))
	}
	for i := range a.elev {
		if a.elev[i] != b.elev[i] {
			t.Fatalf("elevation differs at cell %d: %v vs %v — not deterministic", i, a.elev[i], b.elev[i])
		}
	}
	for i := range a.field.biomes {
		if a.field.biomes[i] != b.field.biomes[i] {
			t.Fatalf("biome differs at pixel %d — not deterministic", i)
		}
		if a.field.passable[i] != b.field.passable[i] {
			t.Fatalf("passability differs at pixel %d — not deterministic", i)
		}
	}
	if len(a.rivers) != len(b.rivers) {
		t.Fatalf("river count differs: %d vs %d — not deterministic", len(a.rivers), len(b.rivers))
	}
	if len(a.reliefs) != len(b.reliefs) {
		t.Fatalf("relief count differs: %d vs %d — not deterministic", len(a.reliefs), len(b.reliefs))
	}
}

// TestBuildWorldModelIsRealContinent proves the model is a continent surrounded by ocean:
// across several seeds it has BOTH land and ocean, the land forms one dominant connected
// mass (not scattered islands), and the frame EDGE is overwhelmingly water (the landmass
// is central, ringed by sea, not painted edge-to-edge like the old flat field).
func TestBuildWorldModelIsRealContinent(t *testing.T) {
	const w, h = 200, 130
	for _, name := range []string{"Rome", "Carthage", "Memphis", "Babylon", "Sparta"} {
		m := buildWorldModel(w, h, worldTerrainSeed(name))

		land, water := modelLandWater(m)
		if land == 0 {
			t.Fatalf("%s: no land — whole canvas is ocean", name)
		}
		if water == 0 {
			t.Fatalf("%s: no ocean — whole canvas is land (not a continent)", name)
		}
		// A meaningful continent: land is a real fraction of the frame, not a speck and not
		// the whole thing. (A round continent in a wide frame; ~15–55% is the healthy band.)
		frac := float64(land) / float64(land+water)
		if frac < 0.12 {
			t.Fatalf("%s: land only %.1f%% — continent too small (a speck, not a landmass)", name, 100*frac)
		}
		if frac > 0.75 {
			t.Fatalf("%s: land %.1f%% — barely any ocean (not a continent ringed by sea)", name, 100*frac)
		}

		// One dominant landmass: the biggest connected component is the vast majority of all
		// land — the continent is coherent, not confetti.
		largest, total := largestLandComponent(m)
		if total == 0 {
			t.Fatalf("%s: no land component", name)
		}
		if share := float64(largest) / float64(total); share < 0.80 {
			t.Fatalf("%s: largest landmass is only %.1f%% of land — scattered islands, not a continent", name, 100*share)
		}

		// Frame edge is mostly water: sample the four borders; the landmass is central.
		edgeWater, edgeTot := 0, 0
		for x := 0; x < w; x++ {
			for _, y := range []int{0, h - 1} {
				edgeTot++
				if !m.field.passableAt(x, y) {
					edgeWater++
				}
			}
		}
		for y := 0; y < h; y++ {
			for _, x := range []int{0, w - 1} {
				edgeTot++
				if !m.field.passableAt(x, y) {
					edgeWater++
				}
			}
		}
		// The continent is ringed by sea, but a large landmass can legitimately run off one
		// edge — the warp-carved shape isn't a disc pinned inside the frame. The border must
		// still be MOSTLY water (≥60%): that proves the map is ocean-framed, not the old
		// edge-to-edge noise field, without demanding a precision the generator doesn't (and
		// shouldn't) guarantee.
		if ew := float64(edgeWater) / float64(edgeTot); ew < 0.60 {
			t.Fatalf("%s: frame edge only %.1f%% water — continent bleeds to the border, not ringed by sea", name, 100*ew)
		}
	}
}

// TestWorldRiversFlowDownhillToSea proves the river model is physical: every kept river's
// elevation is (weakly) non-increasing from source to mouth — water runs DOWNHILL — and
// the final point sits in (or immediately beside) the ocean, so rivers actually reach the
// sea rather than dead-ending inland. Sampled across seeds that produce rivers.
func TestWorldRiversFlowDownhillToSea(t *testing.T) {
	const w, h = 220, 140
	sawRivers := false
	for _, name := range []string{"Rome", "Carthage", "Babylon", "Sparta", "Athens", "Thebes"} {
		m := buildWorldModel(w, h, worldTerrainSeed(name))
		if len(m.rivers) == 0 {
			continue
		}
		sawRivers = true
		for ri, r := range m.rivers {
			if len(r.pts) < 2 {
				t.Fatalf("%s river %d has < 2 points", name, ri)
			}
			// Downhill: elevation at each point must not rise meaningfully as we go
			// downstream. A tiny epsilon absorbs bilinear-sampling wobble at the coast.
			prev := m.elevAtPx(r.pts[0].x, r.pts[0].y)
			for i := 1; i < len(r.pts); i++ {
				e := m.elevAtPx(r.pts[i].x, r.pts[i].y)
				if e > prev+0.06 {
					t.Fatalf("%s river %d rises uphill at point %d: %.3f → %.3f", name, ri, i, prev, e)
				}
				prev = e
			}
			// Reaches the sea: the mouth is a water pixel (the trace stops at the first sea
			// cell), or is immediately adjacent to water.
			mouth := r.pts[len(r.pts)-1]
			touches := !m.field.passableAt(mouth.x, mouth.y)
			for _, nb := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if !m.field.passableAt(mouth.x+nb[0], mouth.y+nb[1]) {
					touches = true
				}
			}
			if !touches {
				t.Fatalf("%s river %d mouth (%d,%d) is inland — does not reach the sea", name, ri, mouth.x, mouth.y)
			}
			// Flow widens toward the mouth: peak flow is at/near the downstream end, not the
			// source (accumulation grows downstream).
			if r.flow[len(r.flow)-1] < r.flow[0] {
				t.Fatalf("%s river %d flow shrinks downstream (%.1f → %.1f) — accumulation backwards",
					name, ri, r.flow[0], r.flow[len(r.flow)-1])
			}
		}
	}
	if !sawRivers {
		t.Fatal("no seed produced any rivers — the river generator is dead")
	}
}

// TestWorldModelRenderSizeAndPanicSafe drives the neutral render at a spread of sizes,
// including tiny/odd/zero, and asserts the image is EXACTLY the requested size, fully
// opaque, and never panics. Guards the upsample + coastline + river/relief passes against
// out-of-range indexing on degenerate canvases.
func TestWorldModelRenderSizeAndPanicSafe(t *testing.T) {
	_ = theme.SetActive("forge")
	sizes := []struct{ w, h int }{{0, 0}, {1, 1}, {1, 2}, {3, 5}, {7, 4}, {40, 26}, {200, 130}}
	for _, sz := range sizes {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("render panicked at %dx%d: %v", sz.w, sz.h, r)
				}
			}()
			st := worldState("iron_age", map[string]int{"forge": 2}, "Rome", nil)
			img, _ := renderWorldImage(st, sz.w, sz.h)
			if img == nil {
				t.Fatalf("nil image at %dx%d", sz.w, sz.h)
			}
			ew, eh := sz.w, sz.h
			if ew < 0 {
				ew = 0
			}
			if eh < 0 {
				eh = 0
			}
			if got := img.Bounds(); got.Dx() != ew || got.Dy() != eh {
				t.Fatalf("size = %dx%d, want %dx%d", got.Dx(), got.Dy(), ew, eh)
			}
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xff {
					t.Fatalf("pixel %d alpha = %d, want 255 at %dx%d", i/4, img.Pix[i], sz.w, sz.h)
				}
			}
		}()
	}
}

// TestWorldDotsLandOnContinent re-asserts, against the new continent model, that your civ
// and every neighbour dot CENTER on land (passable). This is the render-facing contract
// the whole land model exists to keep: no civ floats at sea. (worldmap_test.go's
// TestWorldDotCentersOnLand covers the same via the direct draw seam; this pins it through
// the full renderWorldImage path on a model with real ocean.)
func TestWorldDotsLandOnContinent(t *testing.T) {
	_ = theme.SetActive("forge")
	const w, h = 160, 110
	facs := map[string]game.FactionInfo{
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 70, Strength: 4},
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -50, AtWar: true, Strength: 3},
		"egypt":    {Name: "Egypt", Discovered: true, Status: "neutral", Opinion: 5, Strength: 2},
	}
	st := worldState("classical_age", map[string]int{"forge": 8, "barracks": 6}, "Memphis", facs)
	f := worldFieldFor(st, w, h)
	water := 0
	for _, p := range f.passable {
		if !p {
			water++
		}
	}
	if water == 0 {
		t.Fatal("fixture has no ocean — pick a name/size with real sea")
	}
	_, hueShift := ageInfo(st.Age)
	pal := buildPalette(hueShift)
	seed, _ := ageInfo(st.Age)

	scratch := image.NewRGBA(image.Rect(0, 0, w, h))
	you := drawYourCiv(scratch, pal, f, w, h)
	if !f.passableAt(you.cx, you.cy) {
		t.Fatalf("your civ center (%d,%d) is on water", you.cx, you.cy)
	}
	civs := drawWorldCivs(scratch, pal, st, f, seed, you.cx, you.cy)
	for _, c := range civs {
		if !f.passableAt(c.cx, c.cy) {
			t.Fatalf("civ %q center (%d,%d) is on water", c.name, c.cx, c.cy)
		}
	}
}

// TestDumpWorldModelPNGs renders the WORLD MAP (not the citymap) at a large size for two
// different display-name seeds so a reviewer can eyeball two varied continents — coastline,
// rivers, relief, and the civ/label overlay. Opt-in: skipped unless CITYMAP_PNG_DUMP=<dir>
// is set, e.g.
//
//	CITYMAP_PNG_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpWorldModelPNGs
func TestDumpWorldModelPNGs(t *testing.T) {
	dir := os.Getenv("CITYMAP_PNG_DUMP")
	if dir == "" {
		t.Skip("set CITYMAP_PNG_DUMP=<dir> to dump the world-map continent PNGs")
	}
	_ = theme.SetActive("forge")
	// A discovered-civ roster so the overlay (your civ + neighbour dots + labels) shows up
	// on the atlas, including an at-war foe (the hot-red override + ⚔ mark).
	facs := map[string]game.FactionInfo{
		"rome":     {Name: "Rome", Discovered: true, Status: "allied", Opinion: 75, Strength: 4},
		"carthage": {Name: "Carthage", Discovered: true, Status: "rival", Opinion: -55, AtWar: true, Strength: 5},
		"egypt":    {Name: "Egypt", Discovered: true, Status: "neutral", Opinion: 5, Strength: 2},
		"nubia":    {Name: "Nubia", Discovered: true, Status: "friendly", Opinion: 30, Strength: 1},
	}
	blds := map[string]int{"forge": 14, "barracks": 8, "granary": 6}
	dumps := []struct {
		name string
		file string
	}{
		{"Aldermoor", "wm_A_world1.png"},
		{"Byzantium", "wm_A_world2.png"},
	}
	for _, d := range dumps {
		st := worldState("classical_age", blds, d.name, facs)
		// 400x260 pixels: renderWorldImage takes cell dims and doubles height, so pass
		// 400 cols × 130 rows → a 400×260px image.
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
		land, water := modelLandWater(buildWorldModel(400, 260, worldTerrainSeed(d.name)))
		t.Logf("wrote %s (name=%s, land=%.0f%%)", path, d.name, 100*float64(land)/float64(land+water))
	}
}
