package roguelike

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

var testBuilder = mapmodel.NewBuilder(nil)

func opts(ai int, age string) fixture.Options {
	return fixture.Options{Age: age, Seed: 7 + int64(ai%3)*4, Harbinger: ai%2 == 0, Wars: ai % 2,
		Catastrophe: ai%3 == 0, Ruins: ai%4 == 1}
}

func modelFor(t testing.TB, o fixture.Options) *mapmodel.Model {
	t.Helper()
	st := fixture.State(o)
	return testBuilder.Build(&st, nil)
}

// grown is a model one Grow step on, with the first as its baseline, so it
// carries fresh tiles and a recap.
func grown(t testing.TB, o fixture.Options) (*mapmodel.Model, game.GameState) {
	t.Helper()
	st := fixture.State(o)
	m0 := testBuilder.Build(&st, nil)
	v := mapmodel.VisitOf(m0)
	st1 := fixture.Grow(st, 1)
	return testBuilder.Build(&st1, v), st
}

const sentinel = 'Ж'

func fillScreen(s tcell.SimulationScreen) {
	w, h := s.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s.SetContent(x, y, sentinel, nil, tcell.StyleDefault)
		}
	}
}

// checkRect fails if any cell inside r still holds the sentinel or any cell
// outside it does not.
func checkRect(t *testing.T, s tcell.SimulationScreen, r mapstyle.Rect, what string) {
	t.Helper()
	cells, w, h := s.GetContents()
	bad := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ru := ' '
			if c := cells[y*w+x]; len(c.Runes) > 0 {
				ru = c.Runes[0]
			}
			in := x >= r.X && y >= r.Y && x < r.X+r.W && y < r.Y+r.H
			if in == (ru == sentinel) {
				bad++
				if bad <= 2 {
					t.Errorf("%s: cell %d,%d inside=%v holds %q", what, x, y, in, ru)
				}
			}
		}
	}
}

func draw(v *view, m *mapmodel.Model, w, h, anim int, tier mapmodel.GlyphTier, compact bool) tcell.SimulationScreen {
	s := capture.NewScreen(w+4, h+2)
	fillScreen(s)
	r := mapstyle.Rect{X: 2, Y: 1, W: w, H: h}
	f := mapstyle.Frame{Model: m, Anim: anim, Tier: tier}
	if compact {
		v.DrawCompact(s, r, f)
	} else {
		v.Draw(s, r, f)
	}
	s.Show()
	return s
}

func TestEveryAgeEverySize(t *testing.T) {
	sizes := [][2]int{{200, 60}, {120, 40}, {100, 30}, {80, 24}, {40, 15}}
	for ai, age := range config.AgeOrder() {
		m := modelFor(t, opts(ai, age))
		for z := zRegion; z <= zDistrict; z++ {
			v := newView()
			v.zoom = z
			v.flows = ai%2 == 1
			for _, sz := range sizes {
				for _, compact := range []bool{false, true} {
					if compact && sz[0] != 40 {
						continue
					}
					s := draw(v, m, sz[0], sz[1], ai, mapmodel.TierUnicode, compact)
					checkRect(t, s, mapstyle.Rect{X: 2, Y: 1, W: sz[0], H: sz[1]},
						age+" zoom "+strconv.Itoa(z)+" "+strconv.Itoa(sz[0])+"x"+strconv.Itoa(sz[1]))
				}
			}
		}
	}
}

func TestAnySize(t *testing.T) {
	for _, age := range []string{"primitive_age", "medieval_age", "galactic_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 3, Harbinger: true, Catastrophe: true})
		for _, w := range []int{1, 2, 3, 5, 8, 13, 21, 34, 40, 59, 60, 61, 80} {
			for _, h := range []int{1, 2, 3, 4, 5, 8, 15, 16, 17, 24} {
				for z := zRegion; z <= zDistrict; z++ {
					v := newView()
					v.zoom, v.flows = z, true
					for _, compact := range []bool{false, true} {
						s := draw(v, m, w, h, 3, mapmodel.TierASCII, compact)
						checkRect(t, s, mapstyle.Rect{X: 2, Y: 1, W: w, H: h}, age+" "+strconv.Itoa(w)+"x"+strconv.Itoa(h))
					}
				}
			}
		}
	}
	// no model at all
	v := newView()
	s := draw(v, nil, 80, 24, 0, mapmodel.TierUnicode, false)
	checkRect(t, s, mapstyle.Rect{X: 2, Y: 1, W: 80, H: 24}, "nil model")
}

func TestDeterministicAndAlive(t *testing.T) {
	for _, age := range []string{"primitive_age", "medieval_age", "cyberpunk_age", "galactic_age"} {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 9, Harbinger: true})
		for z := zRegion; z <= zDistrict; z++ {
			v := newView()
			v.zoom = z
			a := capture.Text(draw(v, m, 160, 48, 4, mapmodel.TierUnicode, false))
			b := capture.Text(draw(v, m, 160, 48, 4, mapmodel.TierUnicode, false))
			if a != b {
				t.Errorf("%s zoom %d: same frame drew differently", age, z)
			}
			v2 := newView()
			v2.zoom = z
			if c := capture.Text(draw(v2, m, 160, 48, 4, mapmodel.TierUnicode, false)); c != a {
				t.Errorf("%s zoom %d: a fresh view drew differently", age, z)
			}
			d := capture.Text(draw(v, m, 160, 48, 9, mapmodel.TierUnicode, false))
			if d == a {
				t.Errorf("%s zoom %d: nothing moved between frames", age, z)
			}
		}
	}
}

func TestThemeSweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	var models []*mapmodel.Model
	for i, age := range []string{"primitive_age", "medieval_age", "industrial_age", "digital_age", "cyberpunk_age", "galactic_age"} {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 5, Harbinger: true, Wars: 1, Catastrophe: i%2 == 0})
		models = append(models, m)
	}
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, m := range models {
			for z := zRegion; z <= zDistrict; z++ {
				v := newView()
				v.zoom, v.flows = z, z == zSettlement
				for _, compact := range []bool{false, true} {
					s := draw(v, m, 140, 44, 2, mapmodel.TierUnicode, compact)
					cells, _, _ := s.GetContents()
					for i, c := range cells {
						if len(c.Runes) == 0 || c.Runes[0] == ' ' || c.Runes[0] == sentinel {
							continue
						}
						fg, bg, _ := c.Style.Decompose()
						if fg == bg {
							t.Errorf("theme %s, %s zoom %d: cell %d rune %q has fg == bg", th.Key, m.Age, z, i, c.Runes[0])
							break
						}
					}
				}
			}
		}
	}
}

func TestGlyphWidths(t *testing.T) {
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "atomic_age",
		"digital_age", "cyberpunk_age", "galactic_age"} {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 2, Harbinger: true, Wars: 1, Catastrophe: true, Ruins: true})
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			for z := zRegion; z <= zDistrict; z++ {
				v := newView()
				v.zoom, v.flows = z, true
				for _, compact := range []bool{false, true} {
					s := draw(v, m, 150, 46, 1, tier, compact)
					cells, sw, _ := s.GetContents()
					for i, c := range cells {
						if len(c.Runes) == 0 || c.Runes[0] == sentinel {
							continue
						}
						r := c.Runes[0]
						// '?' is the canvas's fallback for a rune the tier cannot
						// show; only the key hints on the last row may use it.
						if r == '?' && i/sw != 46 {
							t.Errorf("%s %s zoom %d: fallback '?' at row %d", age, tier, z, i/sw)
							break
						}
						if tier == mapmodel.TierASCII && r >= 0x80 {
							t.Errorf("%s ascii zoom %d: non-ASCII rune %q", age, z, r)
							break
						}
						if uniseg.StringWidth(string(r)) != 1 {
							t.Errorf("%s %s zoom %d: rune %q is not one cell", age, tier, z, r)
							break
						}
					}
				}
			}
		}
	}
}

// TestSourceRunesSingleWidth: every rune literal the package draws is one
// cell wide, so the canvas's wide-rune fallback is never needed.
func TestSourceRunesSingleWidth(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING && lit.Kind != token.CHAR {
				return true
			}
			s := lit.Value
			if u, err := strconv.Unquote(s); err == nil {
				s = u
			}
			for _, r := range s {
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

func commandWords(m *mapmodel.Model) map[string]bool {
	words := map[string]bool{}
	for _, c := range append(m.AllCommands(), mapmodel.CmdStatus, mapmodel.CmdExpedition, "build", "upgrade", "assign") {
		words[strings.Fields(c)[0]] = true
	}
	return words
}

func TestInspectTab(t *testing.T) {
	for _, age := range []string{"medieval_age", "cyberpunk_age"} {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 7, Harbinger: true, Wars: 1, Ruins: true})
		words := commandWords(m)
		for _, z := range []int{zSettlement, zRegion} {
			v := newView()
			v.zoom = z
			v.SetOption(mapstyle.OptInspect, true)
			f := mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode}
			draw(v, m, 160, 48, 0, mapmodel.TierUnicode, false)
			s := v.sceneFor(m)
			n := len(s.targets)
			if z == zRegion {
				n += 1 + len(s.siteTargets())
			}
			if n < 5 {
				t.Fatalf("%s: only %d targets", age, n)
			}
			seen := map[mapmodel.Pt]bool{}
			for i := 0; i < n; i++ {
				v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
				seen[v.cur] = true
				in, ok := v.Inspect(f)
				if !ok || in.Title == "" {
					t.Errorf("%s zoom %d target %v: no inspection", age, z, v.cur)
					continue
				}
				c := s.at(v.cur.X, v.cur.Y)
				if c.k != kTile && c.k != kWonder && c.k != kSite {
					continue
				}
				if in.Command == "" || !words[strings.Fields(in.Command)[0]] {
					t.Errorf("%s zoom %d %s: command %q is not a model command", age, z, in.Title, in.Command)
				}
			}
			if len(seen) != n {
				t.Errorf("%s zoom %d: tab visited %d of %d targets", age, z, len(seen), n)
			}
			// draw with the cursor on the last target: the inspector shows it
			txt := capture.Text(draw(v, m, 160, 48, 0, mapmodel.TierUnicode, false))
			in, _ := v.Inspect(f)
			if in.Command != "" && !strings.Contains(txt, "type: "+in.Command) {
				t.Errorf("%s zoom %d: inspector line missing %q", age, z, in.Command)
			}
		}
	}
}

func TestInspectKinds(t *testing.T) {
	m := modelFor(t, fixture.Options{Age: "medieval_age", Seed: 8, Harbinger: true, Catastrophe: true, Idle: 30})
	v := newView()
	draw(v, m, 160, 48, 0, mapmodel.TierUnicode, false)
	s := v.sceneFor(m)
	f := mapstyle.Frame{Model: m}
	check := func(p mapmodel.Pt, want string) {
		t.Helper()
		v.cur = p
		in, ok := v.Inspect(f)
		if !ok || in.Command != want {
			t.Errorf("at %v (%s): command %q, want %q", p, in.Title, in.Command, want)
		}
	}
	check(mapmodel.Pt{X: s.w.CX, Y: s.w.CY}, m.SquareCommand())
	if s.hasHb {
		check(s.harb, mapmodel.CmdHarbinger)
	} else {
		t.Error("no harbinger placed")
	}
	if len(s.idle) == 0 || len(s.hazard) == 0 {
		t.Fatal("no idle workers or hazard ring")
	}
	check(s.idle[0], m.IdleCommand())
	check(s.hazard[0], mapmodel.CmdCatastrophe)
	check(mapmodel.Pt{X: 0, Y: 0}, mapmodel.CmdExpedition)
	v.SetOption(mapstyle.OptInspect, false)
	if _, ok := v.Inspect(f); ok {
		t.Error("Inspect with the cursor hidden reported ok")
	}
}

func TestPlacementStable(t *testing.T) {
	for _, age := range []string{"bronze_age", "medieval_age", "industrial_age", "cyberpunk_age"} {
		o := fixture.Options{Age: age, Seed: 4}
		st := fixture.State(o)
		m0 := testBuilder.Build(&st, nil)
		st1 := fixture.Grow(st, 1)
		m1 := testBuilder.Build(&st1, nil)
		v := newView()
		v.SetOption(mapstyle.OptInspect, false)
		a := draw(v, m0, 200, 60, 0, mapmodel.TierUnicode, false)
		g := v.g
		b := draw(v, m1, 200, 60, 0, mapmodel.TierUnicode, false)
		if v.g.vx != g.vx || v.g.vy != g.vy {
			t.Fatalf("%s: the camera moved", age)
		}
		checked := 0
		for _, tt := range m0.Town.Tiles {
			x, y, ok := g.cellOf(tt.Pt)
			if !ok || tt.Ruin || v.sceneFor(m1).seen(tt.Pt) < 2 {
				continue
			}
			ra, _, _, _ := a.GetContent(x+2, y+1)
			rb, _, _, _ := b.GetContent(x+2, y+1)
			want := mapmodel.R(mapmodel.LineageSym(tt.Lineage, m0.Epoch), mapmodel.TierUnicode)
			if ra != rb || rb != want && !(want == '"' && rb == '\'') {
				t.Errorf("%s: tile %v (%s) drew %q then %q, want %q", age, tt.Pt, tt.Key, ra, rb, want)
			}
			checked++
		}
		if checked < 10 {
			t.Errorf("%s: only %d tiles checked", age, checked)
		}
	}
}

// TestKeys: the view takes only keys that print nothing (typing goes to
// the prompt): the arrows move the cursor (Shift by 8), PgUp and PgDn zoom
// out and in, Home centers, and no letter does anything.
func TestKeys(t *testing.T) {
	m := modelFor(t, fixture.Options{Age: "medieval_age", Seed: 1})
	v := newView()
	f := mapstyle.Frame{Model: m}
	key := func(k tcell.Key, r rune, mod tcell.ModMask) bool {
		return v.HandleKey(tcell.NewEventKey(k, r, mod), f)
	}
	s := v.sceneFor(m)
	if !key(tcell.KeyPgDn, 0, 0) || v.zoom != zDistrict || !key(tcell.KeyPgUp, 0, 0) || !key(tcell.KeyPgUp, 0, 0) || v.zoom != zRegion {
		t.Errorf("zoom keys: zoom %d", v.zoom)
	}
	key(tcell.KeyPgUp, 0, 0)
	if v.zoom != zRegion {
		t.Errorf("PgUp past the region zoom: zoom %d", v.zoom)
	}
	v.zoom = zSettlement
	key(tcell.KeyRight, 0, tcell.ModShift)
	key(tcell.KeyDown, 0, 0)
	if v.cur.X != s.w.CX+8 || v.cur.Y != s.w.CY+1 {
		t.Errorf("cursor at %v, want %d,%d", v.cur, s.w.CX+8, s.w.CY+1)
	}
	key(tcell.KeyHome, 0, 0)
	if v.cur.X != s.w.CX || v.cur.Y != s.w.CY {
		t.Errorf("Home did not center: %v", v.cur)
	}
	for _, r := range "hjklHJKLzxcfn?sgq" {
		if key(tcell.KeyRune, r, 0) {
			t.Errorf("%q was used; letters belong to the prompt", r)
		}
	}
	v.SetOption(mapstyle.OptFlows, true)
	if !v.flows || !v.changes || !v.legend {
		t.Errorf("options: flows %v changes %v legend %v", v.flows, v.changes, v.legend)
	}
	if Entry().Name != "roguelike" || Entry().New().Name() != "roguelike" {
		t.Error("entry")
	}
}

func BenchmarkDraw200x60(b *testing.B) {
	m, _ := grown(b, fixture.Options{Age: "galactic_age", Seed: 7, Harbinger: true, Wars: 1})
	for _, z := range []int{zRegion, zSettlement, zDistrict} {
		b.Run(zoomNames[z], func(b *testing.B) {
			v := newView()
			v.zoom = z
			s := capture.NewScreen(200, 60)
			r := mapstyle.Rect{W: 200, H: 60}
			for i := 0; i < b.N; i++ {
				v.Draw(s, r, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
			}
		})
	}
}

func BenchmarkScene(b *testing.B) {
	m, _ := grown(b, fixture.Options{Age: "galactic_age", Seed: 7, Harbinger: true, Wars: 1})
	for i := 0; i < b.N; i++ {
		newScene(m)
	}
}

// TestCaptures writes text captures for review when ROGUELIKE_CAPTURES
// names a directory.
func TestCaptures(t *testing.T) {
	dir := os.Getenv("ROGUELIKE_CAPTURES")
	if dir == "" {
		t.Skip("set ROGUELIKE_CAPTURES to write captures")
	}
	for ai, age := range config.AgeOrder() {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 7, Harbinger: ai%2 == 0, Wars: 1, Catastrophe: ai == 9})
		for z := zRegion; z <= zDistrict; z++ {
			v := newView()
			v.zoom = z
			v.flows = z == zDistrict
			txt := capture.Text(draw(v, m, 160, 48, 3, mapmodel.TierUnicode, false))
			_ = os.WriteFile(filepath.Join(dir, age+"_"+strings.ToLower(zoomNames[z])+".txt"), []byte(txt), 0o644)
		}
		txt := capture.Text(draw(newView(), m, 40, 15, 3, mapmodel.TierUnicode, true))
		_ = os.WriteFile(filepath.Join(dir, age+"_mini.txt"), []byte(txt), 0o644)
		txt = capture.Text(draw(newView(), m, 80, 24, 3, mapmodel.TierASCII, false))
		_ = os.WriteFile(filepath.Join(dir, age+"_80x24_ascii.txt"), []byte(txt), 0o644)
	}
}
