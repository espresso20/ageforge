//go:build procmaplab

package mapgen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/ui/mapstyle/roguelike"
)

// The lab captures: the new engine and today's map side by side, for three
// ages, by day and by night, at two sizes. Run it with
//
//	go test -tags procmaplab -run TestLabCaptures ./mapgen
//
// PROCMAP_OUT picks the folder the files go to (default: outside the repo).

var labAges = []string{"bronze_age", "modern_age", "space_age"}

var labSizes = [][2]int{{120, 36}, {80, 24}}

const (
	labSeed   = 7
	labNoon   = 225 + 1800*5
	labMidnig = 1125 + 1800*5
)

func outDir() string {
	if d := os.Getenv("PROCMAP_OUT"); d != "" {
		return d
	}
	return "/Users/echo/code/ageforge/design-private/labs/procmap/out"
}

// blit puts a grid on a simulation screen, the one thing the ui does with it.
func blit(g *mapgen.Grid) tcell.SimulationScreen {
	scr := capture.NewScreen(g.W, g.H)
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.At(x, y)
			r := c.R
			if r == 0 {
				r = ' '
			}
			style := tcell.StyleDefault.
				Foreground(tcell.NewRGBColor(int32(c.Fg.R), int32(c.Fg.G), int32(c.Fg.B))).
				Background(tcell.NewRGBColor(int32(c.Bg.R), int32(c.Bg.G), int32(c.Bg.B)))
			scr.SetContent(x, y, r, nil, style)
		}
	}
	return scr
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLabCaptures(t *testing.T) {
	dir := outDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture.WithTheme(t, "forge")
	var gen mapgen.Generator
	for _, age := range labAges {
		for _, tick := range []struct {
			name string
			tick int
		}{{"day", labNoon}, {"night", labMidnig}} {
			st := fixture.State(fixture.Options{Age: age, Seed: labSeed, Tick: tick.tick})
			m := mapmodel.NewBuilder(nil).Build(&st, nil)
			for _, size := range labSizes {
				w, h := size[0], size[1]
				base := fmt.Sprintf("%s_%s_%dx%d", age, tick.name, w, h)

				scr := blit(gen.Render(m, w, h, 0))
				write(t, dir, base+"_new.txt", capture.Text(scr))
				write(t, dir, base+"_new.spans.html", capture.Frame(scr))
				scr.Fini()

				scr = capture.NewScreen(w, h)
				roguelike.New().Draw(scr, mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: m})
				write(t, dir, base+"_old.txt", capture.Text(scr))
				write(t, dir, base+"_old.spans.html", capture.Frame(scr))
				scr.Fini()
			}
		}
	}
}
