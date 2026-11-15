package mapstyle

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/window/kit"
)

// pictureCells reads the cells of the map's picture (below the header, above
// the bars, left of the sidebar) off a 100 by 36 screen, and the header row.
func pictureCells(t *testing.T, themeKey string) (picture, header []kit.Style, runes []rune) {
	t.Helper()
	fixture.WithTheme(t, themeKey)
	m := grown(t, fixture.Options{Age: "medieval_age", Seed: 5})
	scr := draw(New(), m, 100, 36, false, mapmodel.TierUnicode)
	defer scr.Fini()
	cells, w, _ := scr.GetContents()
	for y := 1; y < 1+(36-5); y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			picture = append(picture, c.Style)
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			runes = append(runes, r)
		}
	}
	for x := 0; x < w; x++ {
		header = append(header, cells[x].Style)
	}
	return picture, header, runes
}

// TestPictureIgnoresTheTheme: the map's picture is drawn the same in every
// theme, the full-colour dark one of Forge: the same runes, foregrounds and
// backgrounds under a light theme (Daylight) and a one-ink theme (Monochrome)
// as under Forge. The words and frame round it still follow the theme: the
// header row differs.
func TestPictureIgnoresTheTheme(t *testing.T) {
	forge, forgeHeader, forgeRunes := pictureCells(t, "forge")
	for _, key := range []string{"daylight", "monochrome"} {
		picture, header, runes := pictureCells(t, key)
		if !slices.Equal(runes, forgeRunes) {
			t.Errorf("%s: the picture's runes differ from Forge's", key)
		}
		for i := range forge {
			if picture[i] != forge[i] {
				t.Errorf("%s: picture cell %d is %v, under Forge %v", key, i, picture[i], forge[i])
				break
			}
		}
		if slices.Equal(header, forgeHeader) {
			t.Errorf("%s: the header row is drawn as under Forge, the chrome no longer follows the theme", key)
		}
	}
}
