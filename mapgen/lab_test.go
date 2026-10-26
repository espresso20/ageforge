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
	"github.com/espresso20/ageforge/theme"
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
	labSeed   = 4
	labNoon   = 225 + 1800*5
	labMidnig = 1125 + 1800*5
)

func outDir() string {
	if d := os.Getenv("PROCMAP_OUT"); d != "" {
		return d
	}
	return "/Users/echo/code/ageforge/design-private/labs/procmap/out"
}

// blit puts the w by h part of a grid whose corner is (x0, y0) on a simulation
// screen, the one thing the ui does with it.
func blit(g *mapgen.Grid, x0, y0, w, h int) tcell.SimulationScreen {
	scr := capture.NewScreen(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := g.At(x0+x, y0+y)
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

// writeShot saves one capture three ways: plain text, the coloured spans
// (one span per colour run) and a page that opens by itself.
func writeShot(t *testing.T, dir, base string, scr tcell.SimulationScreen) {
	t.Helper()
	spans := capture.Frame(scr)
	write(t, dir, base+".txt", capture.Text(scr))
	write(t, dir, base+".spans.html", spans)
	bg := theme.Color(theme.RoleBackground)
	fg := theme.Color(theme.RoleText)
	write(t, dir, base+".html", capture.HTML(base, capture.Hex(bg), capture.Hex(fg), capture.Fonts, []string{spans}, 1))
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

				g := gen.Render(m, w, h, 0)
				scr := blit(g, 0, 0, w, h)
				writeShot(t, dir, base+"_new", scr)
				scr.Fini()

				scr = capture.NewScreen(w, h)
				roguelike.New().Draw(scr, mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: m})
				writeShot(t, dir, base+"_old", scr)
				scr.Fini()
			}
		}
	}
}

// wonderBox is a w by h part of the grid centred on the cells of one wonder:
// its top-left corner and size.
func wonderBox(g *mapgen.Grid, key string) (x0, y0, w, h int, ok bool) {
	x1, y1 := -1, -1
	x0, y0 = g.W, g.H
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			if th := g.ThingAt(x, y); th.Kind == mapgen.KindWonder && th.Key == key {
				x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
			}
		}
	}
	if x1 < 0 {
		return 0, 0, 0, 0, false
	}
	w, h = 31, 11
	x0 = max(0, min(g.W-w, (x0+x1)/2-w/2))
	y0 = max(0, min(g.H-h, (y0+y1)/2-h/2))
	return x0, y0, w, h, true
}

// TestLabWonders draws the wonder of each test age, rising (as the fixture
// has it, part banked) and finished, by day and by night, as a close-up.
func TestLabWonders(t *testing.T) {
	dir := outDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture.WithTheme(t, "forge")
	var gen mapgen.Generator
	for _, age := range labAges {
		for _, stage := range []string{"rising", "built"} {
			for _, tick := range []struct {
				name string
				tick int
			}{{"day", labNoon}, {"night", labMidnig}} {
				st := fixture.State(fixture.Options{Age: age, Seed: labSeed, Tick: tick.tick})
				key := st.CurrentAgeWonderKey
				if stage == "built" {
					bs := st.Buildings[key]
					bs.Count = 1
					st.Buildings[key] = bs
				}
				g := gen.Render(mapmodel.NewBuilder(nil).Build(&st, nil), 120, 36, 0)
				x0, y0, w, h, ok := wonderBox(g, key)
				if !ok {
					t.Fatalf("%s: the wonder %s is not in view", age, key)
				}
				scr := blit(g, x0, y0, w, h)
				base := fmt.Sprintf("%s_wonder_%s_%s", age, stage, tick.name)
				writeShot(t, dir, base, scr)
				scr.Fini()
			}
		}
	}
}
