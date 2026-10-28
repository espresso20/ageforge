package procmap

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/ui/mapstyle/styletest"
	"github.com/espresso20/ageforge/vibe"
)

var testBuilder = mapmodel.NewBuilder(nil)

// ageOptions are the fixture options the sweeps give age number ai: a few
// seeds, and a harbinger, a war, a catastrophe or ruins on some.
func ageOptions(ai int, age string) fixture.Options {
	return fixture.Options{Age: age, Seed: 7 + int64(ai%3)*4, Harbinger: ai%2 == 0, Wars: ai % 2,
		Catastrophe: ai%3 == 0, Ruins: ai%4 == 1}
}

func modelFor(t testing.TB, o fixture.Options) *mapmodel.Model {
	t.Helper()
	st := fixture.State(o)
	return testBuilder.Build(&st, nil)
}

// grown is a model one Grow step on, with the first as its baseline, so it
// carries fresh buildings and a recap.
func grown(t testing.TB, o fixture.Options) *mapmodel.Model {
	t.Helper()
	st := fixture.State(o)
	visit := mapmodel.VisitOf(testBuilder.Build(&st, nil))
	next := fixture.Grow(st, 1)
	return testBuilder.Build(&next, visit)
}

const sentinel = 'Ж'

// glyphRules are what every rune the map draws must pass, in every tier: one
// cell wide and never the canvas's '?' fallback.
var glyphRules = styletest.Rules{Query: styletest.Every, Width: styletest.Every}

// drawOffset draws a view into a w by h rectangle inset in a screen filled
// with the sentinel, so a test can tell what the view wrote outside it.
func drawOffset(v mapstyle.Style, m *mapmodel.Model, w, h, anim int, tier mapmodel.GlyphTier, compact bool) tcell.SimulationScreen {
	scr := capture.NewScreen(w+4, h+2)
	for y := 0; y < h+2; y++ {
		for x := 0; x < w+4; x++ {
			scr.SetContent(x, y, sentinel, nil, tcell.StyleDefault)
		}
	}
	r := mapstyle.Rect{X: 2, Y: 1, W: w, H: h}
	f := mapstyle.Frame{Model: m, Anim: anim, Tier: tier}
	if compact {
		v.DrawCompact(scr, r, f)
	} else {
		v.Draw(scr, r, f)
	}
	scr.Show()
	return scr
}

// checkRect fails if any cell inside r still holds the sentinel or any cell
// outside it does not.
func checkRect(t *testing.T, scr tcell.SimulationScreen, r mapstyle.Rect, what string) {
	t.Helper()
	cells, w, h := scr.GetContents()
	bad := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ru := ' '
			if c := cells[y*w+x]; len(c.Runes) > 0 {
				ru = c.Runes[0]
			}
			inside := x >= r.X && y >= r.Y && x < r.X+r.W && y < r.Y+r.H
			if inside == (ru == sentinel) {
				bad++
				if bad <= 2 {
					t.Errorf("%s: cell %d,%d inside=%v holds %q", what, x, y, inside, ru)
				}
			}
		}
	}
}

// TestDrawWritesEveryCellOfItsRectAndNothingOutside: at every size, every age
// writes its whole rectangle and leaves the cells round it alone, in the full
// view and the compact one, with the flows overlay on and off.
func TestDrawWritesEveryCellOfItsRectAndNothingOutside(t *testing.T) {
	fixture.WithTheme(t, "forge")
	sizes := [][2]int{{200, 60}, {120, 40}, {100, 30}, {80, 24}, {40, 15}}
	for ai, age := range config.AgeOrder() {
		m := modelFor(t, ageOptions(ai, age))
		v := New()
		v.SetOption(mapstyle.OptFlows, ai%2 == 1)
		for _, sz := range sizes {
			for _, compact := range []bool{false, true} {
				if compact && sz[0] != 40 {
					continue
				}
				scr := drawOffset(v, m, sz[0], sz[1], ai, mapmodel.TierUnicode, compact)
				checkRect(t, scr, mapstyle.Rect{X: 2, Y: 1, W: sz[0], H: sz[1]}, age+" "+strconv.Itoa(sz[0])+"x"+strconv.Itoa(sz[1]))
				scr.Fini()
			}
		}
	}
}

// TestAnySize: no size panics or strays, down to a single cell, with and
// without a model.
func TestAnySize(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"primitive_age", "medieval_age", "galactic_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 3, Harbinger: true, Catastrophe: true})
		for _, w := range []int{1, 2, 3, 5, 8, 13, 21, 34, 40, 59, 60, 61, 80, 99, 100, 109, 110, 111} {
			for _, h := range []int{1, 2, 3, 4, 5, 8, 15, 16, 17, 24, 29, 30, 31} {
				for _, flows := range []bool{false, true} {
					v := New()
					v.SetOption(mapstyle.OptFlows, flows)
					for _, compact := range []bool{false, true} {
						scr := drawOffset(v, m, w, h, 3, mapmodel.TierASCII, compact)
						checkRect(t, scr, mapstyle.Rect{X: 2, Y: 1, W: w, H: h}, age+" "+strconv.Itoa(w)+"x"+strconv.Itoa(h))
						scr.Fini()
					}
				}
			}
		}
	}
	for _, compact := range []bool{false, true} {
		scr := drawOffset(New(), nil, 80, 24, 0, mapmodel.TierUnicode, compact)
		checkRect(t, scr, mapstyle.Rect{X: 2, Y: 1, W: 80, H: 24}, "no model")
		scr.Fini()
	}
}

// TestOverviewAtEverySizeOfTheWorld: the whole-town view fits a window wider
// or taller than the world itself without reading past its edge.
func TestOverviewAtEverySizeOfTheWorld(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "modern_age", Seed: 7})
	v := New()
	v.SetOption(mapstyle.OptWorld, true)
	for w := 60; w <= 240; w += 17 {
		for h := 16; h <= 80; h += 9 {
			scr := drawOffset(v, m, w, h, 0, mapmodel.TierUnicode, false)
			checkRect(t, scr, mapstyle.Rect{X: 2, Y: 1, W: w, H: h}, strconv.Itoa(w)+"x"+strconv.Itoa(h))
			scr.Fini()
		}
	}
}

// TestDeterministicAndAlive: a frame is a function of the model, the frame
// number and the view: the same frame draws the same, a new view draws the
// same, and the world moves between frames.
func TestDeterministicAndAlive(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"primitive_age", "medieval_age", "cyberpunk_age", "galactic_age"} {
		m := grown(t, fixture.Options{Age: age, Seed: 9, Harbinger: true})
		for _, world := range []bool{false, true} {
			v := New()
			v.SetOption(mapstyle.OptWorld, world)
			first := capture.Text(drawOffset(v, m, 160, 48, 4, mapmodel.TierUnicode, false))
			again := capture.Text(drawOffset(v, m, 160, 48, 4, mapmodel.TierUnicode, false))
			if first != again {
				t.Errorf("%s world %v: the same frame drew differently", age, world)
			}
			fresh := New()
			fresh.SetOption(mapstyle.OptWorld, world)
			if got := capture.Text(drawOffset(fresh, m, 160, 48, 4, mapmodel.TierUnicode, false)); got != first {
				t.Errorf("%s world %v: a fresh view drew differently", age, world)
			}
			if later := capture.Text(drawOffset(v, m, 160, 48, 40, mapmodel.TierUnicode, false)); later == first {
				t.Errorf("%s world %v: nothing moved between frames", age, world)
			}
		}
	}
}

// TestThemeSweep: in every theme, at both zooms and in the compact view, no
// glyph is drawn in its own ground's colour.
func TestThemeSweep(t *testing.T) {
	var models []*mapmodel.Model
	for i, age := range []string{"primitive_age", "medieval_age", "industrial_age", "digital_age", "cyberpunk_age", "galactic_age"} {
		models = append(models, grown(t, fixture.Options{Age: age, Seed: 5, Harbinger: true, Wars: 1, Catastrophe: i%2 == 0}))
	}
	for th := range fixture.Themes(t) {
		for _, m := range models {
			for _, world := range []bool{false, true} {
				v := New()
				v.SetOption(mapstyle.OptWorld, world)
				v.SetOption(mapstyle.OptFlows, !world)
				for _, compact := range []bool{false, true} {
					scr := drawOffset(v, m, 140, 44, 2, mapmodel.TierUnicode, compact)
					cells, _, _ := scr.GetContents()
					for i, c := range cells {
						if styletest.HiddenInk(c) && c.Runes[0] != sentinel {
							t.Errorf("theme %s, %s world %v: cell %d rune %q has fg == bg", th.Key, m.Age, world, i, c.Runes[0])
							break
						}
					}
					scr.Fini()
				}
			}
		}
	}
}

// TestGlyphWidths: in every glyph set every rune the map draws is one cell
// wide and never the canvas's '?' fallback (but on the key hints' last row).
func TestGlyphWidths(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "atomic_age",
		"digital_age", "cyberpunk_age", "galactic_age"} {
		m := grown(t, fixture.Options{Age: age, Seed: 2, Harbinger: true, Wars: 1, Catastrophe: true, Ruins: true})
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			for _, world := range []bool{false, true} {
				v := New()
				v.SetOption(mapstyle.OptWorld, world)
				v.SetOption(mapstyle.OptFlows, true)
				for _, compact := range []bool{false, true} {
					scr := drawOffset(v, m, 150, 46, 1, tier, compact)
					cells, width, _ := scr.GetContents()
					for i, c := range cells {
						if len(c.Runes) == 0 || c.Runes[0] == sentinel {
							continue
						}
						r := c.Runes[0]
						if r == '?' && i/width == 46 {
							continue
						}
						if fault := glyphRules.Fault(r, tier); fault != "" {
							t.Errorf("%s %s world %v: rune %q at row %d: %s", age, tier, world, r, i/width, fault)
							break
						}
					}
					scr.Fini()
				}
			}
		}
	}
}

// TestSourceRunesSingleWidth: every rune literal the map's code draws, here
// and in the engine, is one cell wide, so the canvas's wide-rune fallback is
// never needed, and none is an em dash or an exclamation mark.
func TestSourceRunesSingleWidth(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	engine, _ := filepath.Glob(filepath.Join("..", "..", "..", "mapgen", "*.go"))
	fset := token.NewFileSet()
	for _, f := range append(files, engine...) {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING && lit.Kind != token.CHAR {
				return true
			}
			text := lit.Value
			if unquoted, err := strconv.Unquote(text); err == nil {
				text = unquoted
			}
			for _, r := range text {
				if r >= 0x80 && uniseg.StringWidth(string(r)) != 1 {
					t.Errorf("%s: rune %q (U+%04X) is not one cell wide", fset.Position(lit.Pos()), r, r)
				}
				if r == '—' || r == '!' && lit.Kind == token.STRING {
					t.Errorf("%s: %q in a string: house style forbids it", fset.Position(lit.Pos()), r)
				}
			}
			return true
		})
	}
}

// commandWords are the first words of every command the map may put on its
// inspect line.
func commandWords(m *mapmodel.Model) map[string]bool {
	words := map[string]bool{}
	for _, c := range append(m.AllCommands(), mapmodel.CmdStatus, "build", "upgrade", "assign") {
		words[strings.Fields(c)[0]] = true
	}
	return words
}

// TestTabVisitsEveryTarget: Tab steps through every building and wonder once
// at either zoom, the inspector names each, and the command it offers is one
// the model knows.
func TestTabVisitsEveryTarget(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"medieval_age", "cyberpunk_age"} {
		m := grown(t, fixture.Options{Age: age, Seed: 7, Harbinger: true, Wars: 1, Ruins: true})
		words := commandWords(m)
		for _, world := range []bool{false, true} {
			v := New().(*view)
			v.SetOption(mapstyle.OptWorld, world)
			f := mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode}
			drawOffset(v, m, 160, 48, 0, mapmodel.TierUnicode, false).Fini()
			targets := v.scene(f).Targets
			if len(targets) < 5 {
				t.Fatalf("%s: only %d targets", age, len(targets))
			}
			seen := map[mapgen.Pt]bool{}
			for range targets {
				v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
				seen[v.cur] = true
				drawOffset(v, m, 160, 48, 0, mapmodel.TierUnicode, false).Fini()
				in, ok := v.Inspect(f)
				if !ok || in.Title == "" {
					t.Errorf("%s world %v target %v: no inspection", age, world, v.cur)
					continue
				}
				if in.Command != "" && !words[strings.Fields(in.Command)[0]] {
					t.Errorf("%s world %v %s: command %q is not a model command", age, world, in.Title, in.Command)
				}
			}
			if len(seen) != len(targets) {
				t.Errorf("%s world %v: Tab visited %d of %d targets", age, world, len(seen), len(targets))
			}
			in, _ := v.Inspect(f)
			text := capture.Text(drawOffset(v, m, 160, 48, 0, mapmodel.TierUnicode, false))
			if in.Command != "" && !strings.Contains(text, "type: "+in.Command) {
				t.Errorf("%s world %v: the inspector line is missing %q", age, world, in.Command)
			}
		}
	}
}

// TestInspectCommands: the square offers the square's command, a hidden
// cursor reports nothing, and the edge of the world is a place with a name.
func TestInspectCommands(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "medieval_age", Seed: 8, Harbinger: true, Catastrophe: true, Idle: 30})
	v := New().(*view)
	f := mapstyle.Frame{Model: m}
	drawOffset(v, m, 160, 48, 0, mapmodel.TierUnicode, false).Fini()
	v.HandleKey(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone), f)
	drawOffset(v, m, 160, 48, 0, mapmodel.TierUnicode, false).Fini()
	if in, ok := v.Inspect(f); !ok || in.Command != m.SquareCommand() {
		t.Errorf("on the square: command %q, want %q", in.Command, m.SquareCommand())
	}
	v.cur = mapgen.Pt{}
	if in, ok := v.Inspect(f); !ok || in.Title == "" {
		t.Errorf("at the corner of the world the inspector says %+v", in)
	}
	v.SetOption(mapstyle.OptInspect, false)
	if _, ok := v.Inspect(f); ok {
		t.Error("Inspect with the cursor hidden reported ok")
	}
	if _, ok := v.Pointer(); ok {
		t.Error("Pointer with the cursor hidden reported ok")
	}
}

// TestKeys: the view takes only keys that print nothing (typing goes to the
// prompt): the arrows move the cursor (Shift by 8), PgUp and PgDn zoom out and
// in, Home centres, and no letter does anything.
func TestKeys(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "medieval_age", Seed: 1})
	v := New().(*view)
	f := mapstyle.Frame{Model: m}
	press := func(k tcell.Key, r rune, mod tcell.ModMask) bool {
		return v.HandleKey(tcell.NewEventKey(k, r, mod), f)
	}
	centre := v.scene(f).Centre
	if !press(tcell.KeyPgUp, 0, 0) || !v.zoomSet || v.zoom != zOverview {
		t.Errorf("PgUp: zoom %d, chosen %v", v.zoom, v.zoomSet)
	}
	if !press(tcell.KeyPgDn, 0, 0) || v.zoom != zCloseUp {
		t.Errorf("PgDn: zoom %d", v.zoom)
	}
	press(tcell.KeyRight, 0, tcell.ModShift)
	press(tcell.KeyDown, 0, 0)
	if v.cur.X != centre.X+8 || v.cur.Y != centre.Y+1 {
		t.Errorf("the cursor is at %v, want %d,%d", v.cur, centre.X+8, centre.Y+1)
	}
	press(tcell.KeyHome, 0, 0)
	if v.cur != centre {
		t.Errorf("Home did not centre: %v", v.cur)
	}
	for _, r := range "hjklHJKLzxcfn?sgq+-=_" {
		if press(tcell.KeyRune, r, 0) {
			t.Errorf("%q was used; letters belong to the prompt", r)
		}
	}
	v.SetOption(mapstyle.OptFlows, true)
	if !v.flows || !v.changes || !v.legend {
		t.Errorf("options: flows %v changes %v legend %v", v.flows, v.changes, v.legend)
	}
}

// TestSceneInsetIsTheChrome: the map says which rows of its full view are
// bars (SceneInset), and it is right: the age is named above the scene, the
// keys are listed below it, and neither is in the rows between. The main menu
// shows those rows alone, behind the menu.
func TestSceneInsetIsTheChrome(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "modern_age", "space_age", "quantum_age", "transcendent_age"} {
		for _, size := range [][2]int{{60, 16}, {66, 20}, {95, 36}, {144, 42}} {
			w, h := size[0], size[1]
			st := fixture.State(fixture.Options{Age: age, Seed: 7})
			m := mapmodel.NewBuilder(nil).Model(&st, nil)
			v := New()
			for _, o := range []mapstyle.Option{mapstyle.OptFlows, mapstyle.OptInspect, mapstyle.OptLegend, mapstyle.OptChanges} {
				v.SetOption(o, false)
			}
			scr := capture.NewScreen(w, h)
			v.Draw(scr, mapstyle.Rect{W: w, H: h}, mapstyle.Frame{Model: m, Anim: 3, Clock: 3})
			rows := strings.Split(strings.TrimSuffix(capture.Text(scr), "\n"), "\n")
			scr.Fini()

			top, bottom := v.SceneInset(w, h)
			where := st.AgeName + " at " + strconv.Itoa(w) + "x" + strconv.Itoa(h)
			if top < 1 || bottom < 1 || top+bottom >= h {
				t.Fatalf("%s: inset %d and %d leaves no scene", where, top, bottom)
			}
			ageWord := strings.ToLower(strings.TrimSuffix(st.AgeName, " Age"))
			if head := strings.ToLower(strings.Join(rows[:top], "\n")); !strings.Contains(head, ageWord) {
				t.Errorf("%s: the %d rows above the scene do not name the age: %q", where, top, rows[:top])
			}
			isKeys := func(row string) bool {
				return strings.Contains(row, "PgUp") || strings.Contains(row, "arrows move") || strings.Contains(row, "Tab ")
			}
			if !isKeys(strings.Join(rows[h-bottom:], "\n")) {
				t.Errorf("%s: the %d rows below the scene list no keys: %q", where, bottom, rows[h-bottom:])
			}
			for y := top; y < h-bottom; y++ {
				if isKeys(rows[y]) {
					t.Errorf("%s: row %d is in the scene and lists keys: %q", where, y, rows[y])
				}
			}
		}
	}
}

// BenchmarkMapFrame is one 200x60 frame of the map at a few ages, with
// VibeMode off and on: the cost a player pays every redraw of the map, and
// what the switch adds to it (an atomic load for each text the frame writes,
// and with it on, a lookup of the script for each).
func BenchmarkMapFrame(b *testing.B) {
	for _, age := range []string{"medieval_age", "modern_age", "fusion_age", "space_age", "galactic_age"} {
		for _, name := range []string{"off", "on"} {
			b.Run(age+"/vibemode_"+name, func(b *testing.B) {
				was := vibe.On()
				b.Cleanup(func() { vibe.Set(was) })
				vibe.Set(false)
				m := grown(b, fixture.Options{Age: age, Seed: 7, Harbinger: true, Wars: 1})
				v := New()
				scr := capture.NewScreen(200, 60)
				defer scr.Fini()
				r := mapstyle.Rect{W: 200, H: 60}
				v.Draw(scr, r, mapstyle.Frame{Model: m})
				vibe.Set(name == "on")
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					v.Draw(scr, r, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
				}
			})
		}
	}
}
