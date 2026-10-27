//go:build procmaplab

package mapgen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/ui/mapstyle/procmap"
	"github.com/espresso20/ageforge/ui/mapstyle/roguelike"
)

// The lab captures: the engine's map as the player gets it (through the
// screen layer, in the theme, with the map's header and inspector round it),
// and today's map beside it for a few ages. Run it with
//
//	go test -tags procmaplab -run 'TestLab' ./mapgen
//
// PROCMAP_OUT picks the folder the files go to (default: outside the repo).

const (
	labSeed     = 4
	labNoon     = 225 + 1800*5
	labMidnight = 1125 + 1800*5
)

// labOldAges are the ages today's map is captured for, one to an era.
var labOldAges = map[string]bool{
	"bronze_age": true, "medieval_age": true, "industrial_age": true, "atomic_age": true,
	"digital_age": true, "space_age": true,
}

func outDir() string {
	if d := os.Getenv("PROCMAP_OUT"); d != "" {
		return d
	}
	return "/Users/echo/code/ageforge/design-private/labs/procmap/out"
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeShot saves one capture three ways: plain text, the coloured spans (one
// span per colour run) and a page that opens by itself.
func writeShot(t *testing.T, dir, base string, scr tcell.SimulationScreen) {
	t.Helper()
	spans := capture.Frame(scr)
	write(t, dir, base+".txt", capture.Text(scr))
	write(t, dir, base+".spans.html", spans)
	bg := theme.Color(theme.RoleBackground)
	fg := theme.Color(theme.RoleText)
	write(t, dir, base+".html", capture.HTML(base, capture.Hex(bg), capture.Hex(fg), capture.Fonts, []string{spans}, 1))
}

func labDir(t *testing.T) string {
	dir := outDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func labModel(age string, tick int) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: age, Seed: labSeed, Tick: tick})
	return mapmodel.NewBuilder(nil).Build(&st, nil)
}

func panel(v mapstyle.Style, m *mapmodel.Model, w, h int) tcell.SimulationScreen {
	scr := capture.NewScreen(w, h)
	v.Draw(scr, mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: m})
	return scr
}

// TestLabCaptures draws every age: the engine by day and by night at the big
// size, and as the whole-town overview at the small one; and, for the ages
// today's map is shown for, today's map by day at both sizes.
func TestLabCaptures(t *testing.T) {
	dir := labDir(t)
	fixture.WithTheme(t, "forge")
	for _, age := range mapmodel.NewCatalog().Ages {
		shots := []struct {
			name    string
			tick    int
			w, h    int
			old, nw bool
		}{
			{"day_120x36", labNoon, 120, 36, labOldAges[age], true},
			{"night_120x36", labMidnight, 120, 36, false, true},
			{"day_80x24", labNoon, 80, 24, labOldAges[age], true},
		}
		for _, shot := range shots {
			m := labModel(age, shot.tick)
			base := age + "_" + shot.name
			if shot.nw {
				scr := panel(procmap.New(), m, shot.w, shot.h)
				writeShot(t, dir, base+"_new", scr)
				scr.Fini()
			}
			if shot.old {
				scr := panel(roguelike.New(), m, shot.w, shot.h)
				writeShot(t, dir, base+"_old", scr)
				scr.Fini()
			}
		}
	}
}

// TestLabWonderList writes what the page says of each wonder: its key, its
// name, its age, and the line the game gives it.
func TestLabWonderList(t *testing.T) {
	dir := labDir(t)
	cat := mapmodel.NewCatalog()
	defs := cat.Rules.BuildingMap()
	var lines []string
	for _, w := range cat.Wonders {
		lines = append(lines, fmt.Sprintf("%s|%s|%s|%s", w.Key, w.Name, cat.AgeNames[w.Age], defs[w.Key].Flavor))
	}
	write(t, dir, "wonders.txt", strings.Join(lines, "\n")+"\n")
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

// TestLabWonders draws the wonder of each age, rising (as the fixture has it,
// part banked) and finished, by day and by night, as a close-up.
func TestLabWonders(t *testing.T) {
	dir := labDir(t)
	fixture.WithTheme(t, "forge")
	var gen mapgen.Generator
	for _, age := range mapmodel.NewCatalog().Ages {
		for _, stage := range []string{"rising", "built"} {
			for _, tick := range []struct {
				name string
				tick int
			}{{"day", labNoon}, {"night", labMidnight}} {
				st := fixture.State(fixture.Options{Age: age, Seed: labSeed, Tick: tick.tick})
				key := st.CurrentAgeWonderKey
				if stage == "built" {
					bs := st.Buildings[key]
					bs.Count = 1
					st.Buildings[key] = bs
				}
				scene := gen.Draw(mapmodel.NewBuilder(nil).Build(&st, nil), 0)
				x0, y0, w, h, ok := wonderBox(scene.Grid, key)
				if !ok {
					t.Fatalf("%s: the wonder %s is not in the world", age, key)
				}
				scr := capture.NewScreen(w, h)
				procmap.Blit(scr, mapstyle.Rect{W: w, H: h}, scene.Grid.Window(x0, y0, w, h), mapmodel.TierUnicode)
				writeShot(t, dir, fmt.Sprintf("%s_wonder_%s_%s", age, stage, tick.name), scr)
				scr.Fini()
			}
		}
	}
}

// TestLabThemes draws one age in two themes, a dark one and a light one, by
// day and by night.
func TestLabThemes(t *testing.T) {
	dir := labDir(t)
	for _, key := range []string{"forge", "daylight"} {
		fixture.WithTheme(t, key)
		for _, tick := range []struct {
			name string
			tick int
		}{{"day", labNoon}, {"night", labMidnight}} {
			scr := panel(procmap.New(), labModel("modern_age", tick.tick), 120, 36)
			writeShot(t, dir, fmt.Sprintf("theme_%s_modern_age_%s", key, tick.name), scr)
			scr.Fini()
		}
	}
}
