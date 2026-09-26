package citymap

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// TestDumpThemeComparisonPNGs renders the citymap and worldmap for a spread of
// ages under a dark and several light themes and writes 3×-upscaled PNGs for
// side-by-side eyeballing of the luminance-aware derivations. Opt-in:
//
//	MAP_THEME_DUMP=/tmp/dump go test ./ui/citymap/ -run TestDumpThemeComparisonPNGs
func TestDumpThemeComparisonPNGs(t *testing.T) {
	dir := os.Getenv("MAP_THEME_DUMP")
	if dir == "" {
		t.Skip("set MAP_THEME_DUMP=<dir> to dump theme-comparison PNGs")
	}
	t.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	blds := map[string]int{
		"hut": 26, "gathering_camp": 16, "forge": 10, "stonehenge": 1, "farm": 12,
		"house": 20, "market": 6, "barracks": 5, "temple": 3, "granary": 4, "library": 3,
	}
	ages := []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "modern_age", "space_age", "galactic_age"}
	themes := []string{"forge", "daylight", "parchment", "high_contrast_light"}
	const cw, ch = 160, 100
	citySheet := image.NewRGBA(image.Rect(0, 0, cw*len(themes)+4*len(themes), ch*len(ages)+4*len(ages)))
	worldSheet := image.NewRGBA(image.Rect(0, 0, cw*len(themes)+4*len(themes), ch*len(ages)+4*len(ages)))
	for ti, key := range themes {
		_ = theme.SetActive(key)
		for ai, age := range ages {
			city, _ := renderImage(namedState(age, "Aldermoor", blds), cw, ch)
			world, _ := renderWorldImage(mediumWorld(age), cw, ch/2)
			writeScaledPNG(t, dir+"/city_"+age+"_"+key+".png", city, 4)
			at := image.Pt(ti*(cw+4), ai*(ch+4))
			blit(citySheet, city, at)
			blit(worldSheet, world, at)
		}
	}
	writeScaledPNG(t, dir+"/sheet_city.png", citySheet, 2)
	writeScaledPNG(t, dir+"/sheet_world.png", worldSheet, 2)
	t.Logf("columns: %v; rows: %v", themes, ages)
}

func blit(dst, src *image.RGBA, at image.Point) {
	b := src.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.SetRGBA(at.X+x, at.Y+y, src.RGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
}

func writeScaledPNG(t *testing.T, path string, img *image.RGBA, s int) {
	t.Helper()
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx()*s, b.Dy()*s))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
			for dy := 0; dy < s; dy++ {
				for dx := 0; dx < s; dx++ {
					out.SetRGBA(x*s+dx, y*s+dy, color.RGBA{c.R, c.G, c.B, 0xff})
				}
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		t.Fatal(err)
	}
}
