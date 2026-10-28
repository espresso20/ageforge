//go:build procmaplab

package mapgen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// The lab captures: the engine's map as the player gets it (through the
// screen layer, in the theme, with the map's header and inspector round it).
// Run it with
//
//	go test -tags procmaplab -run 'TestLab' ./mapgen
//
// PROCMAP_OUT picks the folder the files go to (default: outside the repo).

const (
	labSeed     = 4
	labNoon     = 225 + 1800*5
	labMidnight = 1125 + 1800*5
)

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

// TestLabCaptures draws every age: by day and by night at the big size, and as
// the whole-town overview at the small one.
func TestLabCaptures(t *testing.T) {
	dir := labDir(t)
	fixture.WithTheme(t, "forge")
	for _, age := range mapmodel.NewCatalog().Ages {
		shots := []struct {
			name string
			tick int
			w, h int
		}{
			{"day_120x36", labNoon, 120, 36},
			{"night_120x36", labMidnight, 120, 36},
			{"day_80x24", labNoon, 80, 24},
		}
		for _, shot := range shots {
			scr := panel(mapstyle.New(), labModel(age, shot.tick), shot.w, shot.h)
			writeShot(t, dir, age+"_"+shot.name+"_new", scr)
			scr.Fini()
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
				mapstyle.Blit(scr, mapstyle.Rect{W: w, H: h}, scene.Grid.Window(x0, y0, w, h), mapmodel.TierUnicode)
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
			scr := panel(mapstyle.New(), labModel("modern_age", tick.tick), 120, 36)
			writeShot(t, dir, fmt.Sprintf("theme_%s_modern_age_%s", key, tick.name), scr)
			scr.Fini()
		}
	}
}

// TestLabVisitors draws the rare visitor with the cursor on it, on the land,
// in the city, in orbit and in the solar system: the visitor, and the alien's
// craft, each at the middle of its visit.
func TestLabVisitors(t *testing.T) {
	dir := labDir(t)
	fixture.WithTheme(t, "forge")
	var gen mapgen.Generator
	cat := mapmodel.NewCatalog()
	for _, age := range []string{"bronze_age", "modern_age", "space_age", "interstellar_age"} {
		m := labModel(age, labNoon)
		for _, form := range []string{"visitor", "alien"} {
			late := m.AgeIdx >= cat.SpaceAge()
			for from := 0; ; {
				sg := mapmodel.NextSighting(m.Seed, late, from)
				from = sg.Start + sg.Frames
				frame := sg.Start + sg.Frames/2
				g := gen.DrawAt(m, 0, frame).Grid
				at, found := mapgen.Pt{}, false
				for i := range g.Cells {
					if th := g.Things[g.Hit[i]]; th.Kind == mapgen.KindVisitor && th.Key == form {
						at, found = mapgen.Pt{X: i % g.W, Y: i / g.W}, true
					}
				}
				if !found {
					continue
				}
				v := mapstyle.New()
				f := mapstyle.Frame{Model: m, Anim: frame, Clock: frame}
				scr := capture.NewScreen(120, 36)
				v.Draw(scr, mapstyle.Rect{W: 120, H: 36}, f)
				for p, _ := v.Pointer(); p.X != at.X || p.Y != at.Y; p, _ = v.Pointer() {
					key, mod := tcell.KeyRight, tcell.ModNone
					switch {
					case at.X < p.X:
						key = tcell.KeyLeft
					case at.X == p.X && at.Y < p.Y:
						key = tcell.KeyUp
					case at.X == p.X:
						key = tcell.KeyDown
					}
					v.HandleKey(tcell.NewEventKey(key, 0, mod), f)
				}
				v.Draw(scr, mapstyle.Rect{W: 120, H: 36}, f)
				writeShot(t, dir, fmt.Sprintf("visitor_%s_%s", age, form), scr)
				scr.Fini()
				break
			}
		}
	}
}

// TestLabWorld draws the world round the town: civs at the edge coloured by
// how they stand, trade routes with caravans, scouts and an army out, the
// harbinger and idle workers at the square, at the close-up and the region
// zoom, in an age of each scale.
func TestLabWorld(t *testing.T) {
	dir := labDir(t)
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"bronze_age", "modern_age", "space_age", "interstellar_age", "galactic_age"} {
		st := fixture.State(fixture.Options{Age: age, Seed: 6, Tick: labNoon, Wars: 1, Harbinger: true, Idle: 14, Routes: 4})
		st.Military.ActiveScout = &game.ExpeditionSnapshot{Name: "Scout the hills", TicksLeft: 40}
		st.Military.ActiveMilitary = &game.ExpeditionSnapshot{Name: "March on the raiders", TicksLeft: 25}
		m := mapmodel.NewBuilder(nil).Build(&st, nil)
		for _, shot := range []struct {
			name string
			w, h int
		}{{"close", 160, 48}, {"region", 80, 24}} {
			scr := capture.NewScreen(shot.w, shot.h)
			mapstyle.New().Draw(scr, mapstyle.Rect{W: shot.w, H: shot.h}, mapstyle.Frame{Model: m, Anim: 21, Clock: 21})
			writeShot(t, dir, fmt.Sprintf("world_%s_%s", age, shot.name), scr)
			scr.Fini()
		}
	}
}
