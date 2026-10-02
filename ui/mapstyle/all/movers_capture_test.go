//go:build mapcapture

package all

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// moverAges are the towns the traffic review looks at: one per era that
// brings new movers.
var moverAges = []string{"stone_age", "iron_age", "industrial_age", "digital_age", "cyberpunk_age", "space_age", "galactic_age"}

// TestWriteMoverCaptures draws the roguelike settlement of each era's town
// from the smoke-bot states, 12 frames a second apart, for reviewing the
// traffic (run TestGenerateStates in ui/mapstyle/capture first):
//
//	go test -tags mapcapture -run TestWriteMoverCaptures ./ui/mapstyle/all
//
// MAP_STATES_DIR picks the states, MOVERS_CAPTURE_DIR the output (default
// map_captures/movers at the repo root, which git ignores).
func TestWriteMoverCaptures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "map_captures")
	states := envOr("MAP_STATES_DIR", filepath.Join(root, "states"))
	out := envOr("MOVERS_CAPTURE_DIR", filepath.Join(root, "movers"))
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	if err := theme.SetActive("forge"); err != nil {
		t.Fatal(err)
	}
	for _, age := range moverAges {
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
			name := short(age)
			if tier == mapmodel.TierASCII {
				name += "_ascii"
			}
			frames := make([]int, 12)
			for i := range frames {
				frames[i] = 1000 + i*8
			}
			moverShot(t, states, out, "roguelike", age, name, tier, frames, "")
		}
	}
}

// moverShot draws one state in a style (the roguelike's settlement) at
// 160x48 at the given animation frames and writes name.txt, name.html and
// png/name.png.
func moverShot(t *testing.T, states, out, style, state, name string, tier mapmodel.GlyphTier, anims []int, note string) {
	t.Helper()
	st, err := capture.LoadState(filepath.Join(states, state+".json.gz"))
	if err != nil {
		t.Logf("skip %s: %v", state, err)
		return
	}
	if err := os.MkdirAll(filepath.Join(out, "png"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := mapmodel.NewBuilder(nil).Build(&st, nil)
	view, _ := Registry().New(style)
	view.SetOption(mapstyle.OptInspect, true)
	scr := capture.NewScreen(160, 48)
	defer scr.Fini()
	var html []string
	var txt strings.Builder
	for i, a := range anims {
		view.Draw(scr, mapstyle.Rect{W: 160, H: 48}, mapstyle.Frame{Model: m, Anim: a, Tier: tier})
		html = append(html, capture.Frame(scr))
		if i == 0 {
			_ = capture.PNG(filepath.Join(out, "png", name+".png"), scr)
		}
		fmt.Fprintf(&txt, "--- frame %d\n", a)
		txt.WriteString(capture.Text(scr))
	}
	title := fmt.Sprintf("%s · %s · %s · 160x48 · glyphs %s", style, state, name, tier)
	if note != "" {
		title += " · " + note
	}
	bg := capture.Hex(theme.Color(theme.RoleBackground))
	fg := capture.Hex(theme.Color(theme.RoleText))
	if err := os.WriteFile(filepath.Join(out, name+".html"), []byte(capture.HTML(title, bg, fg, capture.Fonts, html, 4)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, name+".txt"), []byte(title+"\n"+txt.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
