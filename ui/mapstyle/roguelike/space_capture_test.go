package roguelike

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestSkyCaptures writes review captures of the sky scenes from the real
// smoke-bot states when ROGUELIKE_SKY_CAPTURES names a directory
// (map_captures/states must hold them: see ui/mapstyle/capture).
func TestSkyCaptures(t *testing.T) {
	dir := os.Getenv("ROGUELIKE_SKY_CAPTURES")
	if dir == "" {
		t.Skip("set ROGUELIKE_SKY_CAPTURES to write captures")
	}
	states := filepath.Join("..", "..", "..", "map_captures", "states")
	ages := strings.Fields(os.Getenv("SKY_AGES"))
	if len(ages) == 0 {
		ages = []string{"fusion_age", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}
	}
	th := os.Getenv("SKY_THEME")
	if th == "" {
		th = "forge"
	}
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	if err := theme.SetActive(th); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(dir, 0o755)
	for _, age := range ages {
		st, err := capture.LoadState(filepath.Join(states, age+".json.gz"))
		if err != nil {
			t.Logf("skip %s: %v", age, err)
			continue
		}
		m := mapmodel.NewBuilder(nil).Build(&st, nil)
		for _, sz := range [][3]int{{160, 48, 0}, {40, 15, 1}, {160, 48, 2}} {
			v := Entry().New()
			if sz[2] == 2 {
				v.HandleKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), mapstyle.Frame{Model: m})
			}
			scr := capture.NewScreen(sz[0], sz[1])
			var frames []string
			var txt strings.Builder
			for i, a := range []int{1000, 1003, 1006, 1009, 1012, 1015} {
				f := mapstyle.Frame{Model: m, Anim: a, Tier: mapmodel.TierUnicode}
				if sz[2] == 1 {
					v.DrawCompact(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, f)
				} else {
					v.Draw(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, f)
				}
				frames = append(frames, capture.Frame(scr))
				if i == 0 {
					txt.WriteString(capture.Text(scr))
				}
			}
			name := fmt.Sprintf("%s_%s", age, [3]string{"full", "mini", "close"}[sz[2]])
			if th != "forge" {
				name += "_" + th
			}
			_ = capture.PNG(filepath.Join(dir, name+".png"), scr)
			_ = os.WriteFile(filepath.Join(dir, name+".txt"), []byte(txt.String()), 0o644)
			bg := capture.Hex(theme.Color(theme.RoleBackground))
			fg := capture.Hex(theme.Color(theme.RoleText))
			_ = os.WriteFile(filepath.Join(dir, name+".html"), []byte(capture.HTML(name, bg, fg, capture.Fonts, frames, 4)), 0o644)
			scr.Fini()
		}
	}
}
