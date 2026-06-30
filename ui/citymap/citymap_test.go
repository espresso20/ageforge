package citymap

import (
	"image"
	"image/color"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// sampleState builds a small GameState for a given age with an optional set of
// built buildings (name → count). Only the fields the renderer reads (Age,
// Buildings) are populated.
func sampleState(age string, built map[string]int) game.GameState {
	bs := map[string]game.BuildingState{}
	for k, n := range built {
		bs[k] = game.BuildingState{Count: n, Name: k}
	}
	return game.GameState{Age: age, Buildings: bs}
}

// imagesDiffer reports whether two same-sized RGBA images have any differing
// pixel.
func imagesDiffer(a, b *image.RGBA) bool {
	if a.Bounds() != b.Bounds() {
		return true
	}
	if len(a.Pix) != len(b.Pix) {
		return true
	}
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			return true
		}
	}
	return false
}

// TestRenderImageNoPanic renders across a couple of ages and both empty and
// populated states. It must not panic and must produce an image of exactly the
// requested pixel size.
func TestRenderImageNoPanic(t *testing.T) {
	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	cases := []struct {
		name  string
		age   string
		built map[string]int
	}{
		{"primitive-empty", "primitive_age", nil},
		{"primitive-with-buildings", "primitive_age", map[string]int{"hut": 3, "gathering_camp": 2}},
		{"iron-with-buildings", "iron_age", map[string]int{"forge": 1, "barracks": 2, "granary": 1, "hunting_lodge": 1}},
		{"unknown-age", "made_up_age", map[string]int{"x": 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const w, h = 80, 48
			img := renderImage(sampleState(tc.age, tc.built), w, h)
			if img == nil {
				t.Fatal("renderImage returned nil")
			}
			if got := img.Bounds(); got.Dx() != w || got.Dy() != h {
				t.Fatalf("image size = %dx%d, want %dx%d", got.Dx(), got.Dy(), w, h)
			}
			// Every pixel must be fully opaque (we set A=0xff everywhere).
			for i := 3; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 0xff {
					t.Fatalf("pixel %d alpha = %d, want 255", i/4, img.Pix[i])
				}
			}
		})
	}
}

// TestThemeAwareTerrainDiffers proves the renderer is theme-aware: the SAME state
// rendered under two different themes must yield different pixels. If it didn't,
// the terrain would not be pulling from the active palette.
func TestThemeAwareTerrainDiffers(t *testing.T) {
	state := sampleState("bronze_age", map[string]int{"forge": 2, "granary": 1})

	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	imgForge := renderImage(state, 80, 48)

	if err := theme.SetActive("high_contrast"); err != nil {
		t.Fatalf("SetActive(high_contrast): %v", err)
	}
	imgHC := renderImage(state, 80, 48)

	// Restore the default so other tests in the package start from forge.
	_ = theme.SetActive("forge")

	if !imagesDiffer(imgForge, imgHC) {
		t.Fatal("terrain pixels identical across forge vs high_contrast — renderer is not theme-aware")
	}
}

// TestCacheKeyIncludesTheme verifies the on-demand cache invalidates on a theme
// switch: imageFor must return a freshly-tinted image after SetActive, not the
// stale cached one.
func TestCacheKeyIncludesTheme(t *testing.T) {
	cm := NewCityMap()
	cm.Refresh(sampleState("classical_age", map[string]int{"forge": 1}))

	if err := theme.SetActive("forge"); err != nil {
		t.Fatalf("SetActive(forge): %v", err)
	}
	first := cm.imageFor(80, 48)
	// A second call under the same theme/size/state must hit the cache (same pointer).
	if again := cm.imageFor(80, 48); again != first {
		t.Fatal("imageFor regenerated despite identical cache key (cache not working)")
	}

	if err := theme.SetActive("high_contrast"); err != nil {
		t.Fatalf("SetActive(high_contrast): %v", err)
	}
	switched := cm.imageFor(80, 48)
	_ = theme.SetActive("forge")

	if switched == first {
		t.Fatal("imageFor returned the cached image after a theme switch — theme key not in cache key")
	}
	if !imagesDiffer(first, switched) {
		t.Fatal("imageFor produced identical pixels after a theme switch")
	}
}

// TestHalfBlockStreamingMapping verifies the cols/rows mapping of the half-block
// streamer: a w×(h*2) image is written into a w×h cell area where each cell is a
// '▄' whose background is the UPPER pixel and foreground the LOWER pixel.
func TestHalfBlockStreamingMapping(t *testing.T) {
	const cols, rows = 6, 4
	img := image.NewRGBA(image.Rect(0, 0, cols, rows*2))
	// Paint a deterministic gradient so each pixel is unique and we can verify the
	// upper/lower split per cell.
	for y := 0; y < rows*2; y++ {
		for x := 0; x < cols; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x * 7),
				G: uint8(y * 11),
				B: uint8((x + y) * 5),
				A: 0xff,
			})
		}
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen.Init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(cols, rows)

	streamHalfBlocks(screen, img, 0, 0, cols, rows)
	screen.Show()

	cells, gotW, gotH := screen.GetContents()
	if gotW < cols || gotH < rows {
		t.Fatalf("screen size = %dx%d, want at least %dx%d", gotW, gotH, cols, rows)
	}

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := cells[row*gotW+col]
			if len(cell.Runes) != 1 || cell.Runes[0] != '▄' {
				t.Fatalf("cell (%d,%d) rune = %q, want '▄'", col, row, string(cell.Runes))
			}
			fg, bg, _ := cell.Style.Decompose()
			wantUpper := img.RGBAAt(col, row*2)
			wantLower := img.RGBAAt(col, row*2+1)
			if !colorMatches(bg, wantUpper) {
				t.Fatalf("cell (%d,%d) bg = %v, want upper pixel %v", col, row, bg, wantUpper)
			}
			if !colorMatches(fg, wantLower) {
				t.Fatalf("cell (%d,%d) fg = %v, want lower pixel %v", col, row, fg, wantLower)
			}
		}
	}
}

// colorMatches compares a tcell.Color against an RGBA via true-RGB components.
func colorMatches(c tcell.Color, want color.RGBA) bool {
	r, g, b := c.RGB()
	return uint8(r) == want.R && uint8(g) == want.G && uint8(b) == want.B
}
