package skyline

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

const sentinel = 'Ж'

// drawBoxed draws into a rect inset in a larger screen pre-filled with a
// sentinel, and checks every cell inside was written and nothing outside.
func drawBoxed(t *testing.T, v *view, m *mapmodel.Model, w, h, anim int, tier mapmodel.GlyphTier, compact bool) {
	t.Helper()
	scr := capture.NewScreen(w+4, h+2)
	st := tcell.StyleDefault
	for y := 0; y < h+2; y++ {
		for x := 0; x < w+4; x++ {
			scr.SetContent(x, y, sentinel, nil, st)
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
	cells, sw, sh := scr.GetContents()
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			c := cells[y*sw+x]
			ch := ' '
			if len(c.Runes) > 0 {
				ch = c.Runes[0]
			}
			inside := x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
			if inside && ch == sentinel {
				t.Fatalf("%s %dx%d compact=%v: cell (%d,%d) left unwritten", m.Age, w, h, compact, x-r.X, y-r.Y)
			}
			if !inside && ch != sentinel {
				t.Fatalf("%s %dx%d compact=%v: wrote outside the rect at (%d,%d)", m.Age, w, h, compact, x, y)
			}
		}
	}
}

func TestEveryAgeEverySize(t *testing.T) {
	for i, age := range config.AgeOrder() {
		for _, tod := range []float64{0.5, 0.95} {
			o := fixture.Options{Age: age, Seed: int64(i + 1), Tick: tickAt(tod), Harbinger: i%2 == 1,
				Wars: i % 3 % 2, Catastrophe: i%3 == 2}
			m := build(o, 0)
			v := newView()
			v.flows = i%4 == 0
			v.inspect = i%5 == 0
			for _, sz := range [][2]int{{200, 60}, {120, 40}, {100, 30}, {80, 24}} {
				drawBoxed(t, v, m, sz[0], sz[1], 10+i, mapmodel.TierUnicode, false)
			}
			drawBoxed(t, v, m, 40, 15, 3, mapmodel.TierUnicode, true)
			drawBoxed(t, v, m, 40, 15, 3, mapmodel.TierUnicode, false)
		}
	}
}

func TestTinySizes(t *testing.T) {
	m := build(fixture.Options{Age: "atomic_age", Seed: 2, Harbinger: true, Catastrophe: true}, 1)
	v := newView()
	v.inspect, v.flows = true, true
	for w := 1; w <= 64; w += 3 {
		for h := 1; h <= 18; h++ {
			drawBoxed(t, v, m, w, h, 5, mapmodel.TierUnicode, false)
			drawBoxed(t, v, m, w, h, 5, mapmodel.TierASCII, true)
		}
	}
	scr := capture.NewScreen(10, 5)
	v.Draw(scr, mapstyle.Rect{W: 0, H: 3}, mapstyle.Frame{Model: m})
	v.Draw(scr, mapstyle.Rect{W: 10, H: 5}, mapstyle.Frame{})
	v.DrawCompact(scr, mapstyle.Rect{W: 10, H: 5}, mapstyle.Frame{})
}

func screenCells(s tcell.SimulationScreen) []tcell.SimCell {
	c, _, _ := s.GetContents()
	out := make([]tcell.SimCell, len(c))
	copy(out, c)
	return out
}

func sameCells(a, b []tcell.SimCell) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Style != b[i].Style || string(a[i].Runes) != string(b[i].Runes) {
			return false
		}
	}
	return true
}

func TestDeterminism(t *testing.T) {
	for _, age := range []string{"medieval_age", "cyberpunk_age", "space_age"} {
		m := build(fixture.Options{Age: age, Seed: 9, Routes: 4, Tick: tickAt(0.9)}, 1)
		a := screenCells(draw(newView(), m, 160, 45, 33, mapmodel.TierUnicode, false))
		b := screenCells(draw(newView(), m, 160, 45, 33, mapmodel.TierUnicode, false))
		if !sameCells(a, b) {
			t.Errorf("%s: the same frame drew differently twice", age)
		}
		c := screenCells(draw(newView(), m, 160, 45, 61, mapmodel.TierUnicode, false))
		if sameCells(a, c) {
			t.Errorf("%s: two animation frames drew the same cells: nothing moves", age)
		}
		ca := screenCells(draw(newView(), m, 40, 15, 33, mapmodel.TierUnicode, true))
		cb := screenCells(draw(newView(), m, 40, 15, 33, mapmodel.TierUnicode, true))
		if !sameCells(ca, cb) {
			t.Errorf("%s: the compact view is not deterministic", age)
		}
	}
}

func countKind(vs []vehicle, k vkind) int {
	n := 0
	for _, v := range vs {
		if v.kind == k {
			n++
		}
	}
	return n
}

func TestTrafficDensity(t *testing.T) {
	for _, age := range []string{"stone_age", "classical_age", "medieval_age", "industrial_age", "electric_age",
		"modern_age", "cyberpunk_age", "fusion_age", "galactic_age"} {
		none := build(fixture.Options{Age: age, Seed: 4, Routes: -1}, 0)
		busy := build(fixture.Options{Age: age, Seed: 4, Routes: 5}, 0)
		for _, anim := range []int{0, 17, 90} {
			if busy.Activity.Routes == 0 {
				break // no trade routes open this early
			}
			vn := trafficFor(none, anim, 0, 160, 39)
			vb := trafficFor(busy, anim, 0, 160, 39)
			if n := countKind(vn, vkRoute); n != 0 {
				t.Errorf("%s: %d route vehicles with no routes running", age, n)
			}
			if countKind(vb, vkRoute) <= countKind(vn, vkRoute) {
				t.Errorf("%s: more trade routes did not bring more route traffic", age)
			}
		}
		few, many := *busy, *busy
		few.Activity.Staffed, many.Activity.Staffed = 3, 5000
		if countKind(trafficFor(&many, 5, 0, 160, 39), vkFoot) <= countKind(trafficFor(&few, 5, 0, 160, 39), vkFoot) {
			t.Errorf("%s: more staffed workers did not bring more foot traffic", age)
		}
		idle := *busy
		idle.Activity.Staffed = 0
		if n := countKind(trafficFor(&idle, 5, 0, 160, 39), vkFoot); n != 0 {
			t.Errorf("%s: %d walkers with nobody at work", age, n)
		}
	}
	m := build(fixture.Options{Age: "medieval_age", Seed: 4, Wars: 2}, 0)
	if m.Activity.Wars > 0 {
		wb := 0
		for _, v := range trafficFor(m, 0, 0, 160, 39) {
			if v.t == &tWarband {
				wb++
			}
		}
		if wb != m.Activity.Wars {
			t.Errorf("%d warbands for %d wars", wb, m.Activity.Wars)
		}
	}
}

// TestTrafficOcclusion: in a neon city, a vehicle never draws over a cell a
// nearer building owns: those cells are identical with and without traffic.
func TestTrafficOcclusion(t *testing.T) {
	m := build(fixture.Options{Age: "cyberpunk_age", Seed: 7, Routes: 6, Tick: tickAt(0.9), Scale: 2}, 0)
	for _, anim := range []int{3, 40, 77, 120} {
		v := newView()
		v.noTraffic = true
		v.compose(mapstyle.Frame{Model: m, Anim: anim}, 160, 45)
		bare := append([]cell(nil), v.fb.c...)
		v.noTraffic = false
		s := v.compose(mapstyle.Frame{Model: m, Anim: anim}, 160, 45)
		hidden, shown := 0, 0
		nearest := map[int]uint8{} // the nearest vehicle over each cell
		for _, vh := range trafficFor(m, anim, s.cam, s.W, s.groundY) {
			vh := vh
			vh.cells(func(x, y int, ch rune, mk byte) {
				y = s.Y(y)
				if x < 0 || x >= s.W || y < 0 || y >= s.H {
					return
				}
				if d, ok := nearest[y*s.W+x]; !ok || vh.depth < d {
					nearest[y*s.W+x] = vh.depth
				}
			})
		}
		for i, d := range nearest {
			if bare[i].d < d && bare[i].d >= dWonder && bare[i].d <= dRow2 { // a nearer building owns it
				hidden++
				if v.fb.c[i] != bare[i] {
					t.Fatalf("anim %d: a vehicle at depth %d drew over a nearer cell (depth %d) at %d,%d", anim, d, bare[i].d, i%s.W, i/s.W)
				}
			} else {
				shown++
			}
		}
		if hidden == 0 || shown == 0 {
			t.Errorf("anim %d: occlusion not exercised (hidden %d, shown %d)", anim, hidden, shown)
		}
	}
	// the lanes sit between the rows
	depths := map[uint8]bool{}
	for _, vh := range trafficFor(m, 5, 0, 160, 39) {
		depths[vh.depth] = true
	}
	for _, d := range []uint8{dLane0, dLane1, dLane2} {
		if !depths[d] {
			t.Errorf("no cyberpunk traffic in the lane at depth %d", d)
		}
	}
}

// TestPlacementStable: after the city grows, every silhouette that was on
// screen is still centred on the same world column and still drawn there.
func TestPlacementStable(t *testing.T) {
	o := fixture.Options{Age: "renaissance_age", Seed: 11, Tick: tickAt(0.5)}
	b := mapmodel.NewBuilder(catalog())
	st := fixture.State(o)
	m1 := b.Build(&st, nil)
	st2 := fixture.Grow(st, 1)
	m2 := b.Build(&st2, nil)
	type id struct {
		key string
		cp  int
	}
	pos := map[id]int{}
	for _, l := range m1.Skyline.Lots {
		pos[id{l.Key, l.Copy}] = l.X
	}
	moved := 0
	for _, l := range m2.Skyline.Lots {
		if x, ok := pos[id{l.Key, l.Copy}]; ok && x != l.X {
			moved++
		}
	}
	if moved > 0 {
		t.Fatalf("%d lots moved after growth", moved)
	}
	v := newView()
	v.follow = false
	check := func(m *mapmodel.Model) map[id]bool {
		v.cam = 0
		v.follow = false
		s := v.compose(mapstyle.Frame{Model: m, Anim: 1}, 200, 60)
		out := map[id]bool{}
		for i := range s.lay.lots {
			lv := &s.lay.lots[i]
			x := lv.ml.X - s.cam
			if x < 0 || x >= s.W {
				continue
			}
			c := v.fb.at(x, s.Y(s.groundY-1))
			out[id{lv.ml.Key, lv.ml.Copy}] = c.d <= dRow2
		}
		return out
	}
	before, after := check(m1), check(m2)
	for k, drawn := range before {
		if drawn && !after[k] {
			t.Errorf("%v was drawn at its column before growth and not after", k)
		}
	}
}

func TestThemeSweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	ages := []string{"stone_age", "medieval_age", "industrial_age", "cyberpunk_age", "quantum_age"}
	models := map[string]*mapmodel.Model{}
	for _, a := range ages {
		for _, tod := range []float64{0.5, 0.95} {
			models[a+strconv.FormatFloat(tod, 'f', 2, 64)] = build(fixture.Options{Age: a, Seed: 5, Tick: tickAt(tod),
				Harbinger: true, Routes: 3}, 1)
		}
	}
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for name, m := range models {
			for _, compact := range []bool{false, true} {
				w, h := 120, 36
				if compact {
					w, h = 40, 15
				}
				v := newView()
				v.flows, v.inspect = true, true
				scr := draw(v, m, w, h, 7, mapmodel.TierUnicode, compact)
				cells, sw, _ := scr.GetContents()
				for i, c := range cells {
					fg, bg, _ := c.Style.Decompose()
					if len(c.Runes) > 0 && c.Runes[0] != ' ' && fg == bg {
						t.Errorf("theme %s, %s compact=%v: %q at (%d,%d) has fg == bg", th.Key, name, compact, c.Runes[0], i%sw, i/sw)
						break
					}
				}
			}
		}
	}
}

func TestGlyphWidths(t *testing.T) {
	for _, age := range []string{"primitive_age", "medieval_age", "victorian_age", "cyberpunk_age", "space_age", "transcendent_age"} {
		m := build(fixture.Options{Age: age, Seed: 6, Routes: 4, Harbinger: true, Catastrophe: true, Wars: 1, Tick: tickAt(0.9)}, 1)
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierNerd, mapmodel.TierASCII} {
			for _, compact := range []bool{false, true} {
				v := newView()
				v.inspect, v.flows = true, true
				w, h := 160, 45
				if compact {
					w, h = 40, 15
				}
				scr := draw(v, m, w, h, 21, tier, compact)
				cells, _, _ := scr.GetContents()
				for _, c := range cells {
					if len(c.Runes) == 0 {
						continue
					}
					r := c.Runes[0]
					switch {
					case tier == mapmodel.TierASCII && r >= 0x80:
						t.Fatalf("%s ascii compact=%v: non-ASCII rune %q", age, compact, r)
					case tier != mapmodel.TierASCII && uniseg.StringWidth(string(r)) != 1:
						t.Fatalf("%s %s compact=%v: rune %q is not one cell wide", age, tier, compact, r)
					case tier != mapmodel.TierASCII && r == '?':
						t.Fatalf("%s %s compact=%v: a '?' fallback leaked into the frame", age, tier, compact)
					case r == '—':
						t.Fatalf("%s: an em dash in the frame", age)
					}
				}
			}
		}
	}
}

func commandWords(m *mapmodel.Model) map[string]bool {
	w := map[string]bool{}
	for _, c := range m.AllCommands() {
		w[strings.Fields(c)[0]] = true
	}
	return w
}

func TestInspect(t *testing.T) {
	for _, age := range []string{"medieval_age", "cyberpunk_age"} {
		m := build(fixture.Options{Age: age, Seed: 3, Harbinger: true, Wars: 1, Tick: tickAt(0.5)}, 1)
		words := commandWords(m)
		v := newView()
		f := mapstyle.Frame{Model: m}
		if _, ok := v.Inspect(f); ok {
			t.Fatal("Inspect reported a cursor before it was put out")
		}
		v.SetOption(mapstyle.OptInspect, true)
		in, ok := v.Inspect(f)
		if !ok || v.cur.kind != tLot || in.Command == "" {
			t.Fatalf("%s: the cursor did not start on a building (%+v)", age, in)
		}
		_ = draw(v, m, 160, 48, 1, mapmodel.TierUnicode, false)
		kinds := map[tkind]int{}
		seen := map[target]bool{}
		for i := 0; i < len(m.Skyline.Lots)+20; i++ {
			v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
			in, ok := v.Inspect(f)
			if !ok || in.Title == "" {
				t.Fatalf("%s: step %d: nothing under the cursor", age, i)
			}
			if in.Command == "" || !words[strings.Fields(in.Command)[0]] {
				t.Fatalf("%s: %q: command %q is not a model command", age, in.Title, in.Command)
			}
			kinds[v.cur.kind]++
			seen[v.cur] = true
			if v.cur.kind == tLot && !strings.Contains(strings.Join(in.Lines, " "), "·") && !strings.HasPrefix(strings.Join(in.Lines, " "), "wonder") {
				t.Errorf("%s: %q has no lineage line: %v", age, in.Title, in.Lines)
			}
		}
		if kinds[tLot] == 0 || kinds[tCiv] == 0 || kinds[tHarbinger] == 0 {
			t.Errorf("%s: Tab did not reach every kind of target: %v", age, kinds)
		}
		for _, k := range []tcell.Key{tcell.KeyLeft, tcell.KeyRight, tcell.KeyUp, tcell.KeyDown} {
			if !v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f) {
				t.Errorf("inspect mode ignored key %v", k)
			}
			if _, ok := v.Inspect(f); !ok {
				t.Errorf("cursor lost after key %v", k)
			}
		}
		scr := draw(v, m, 160, 48, 1, mapmodel.TierUnicode, false)
		if st := capture.Text(scr); !strings.Contains(st, "type: ") {
			t.Errorf("%s: the status line does not show the command", age)
		}
		// a civ says its relation in words
		for tgt := range seen {
			if tgt.kind == tCiv {
				v.cur = tgt
				in, _ := v.Inspect(f)
				if !strings.HasPrefix(in.Lines[0], "relation: ") {
					t.Errorf("civ inspection does not spell the relation: %v", in.Lines)
				}
				break
			}
		}
	}
}

func TestKeys(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3}, 0)
	f := mapstyle.Frame{Model: m}
	v := newView()
	_ = draw(v, m, 100, 30, 0, mapmodel.TierUnicode, false)
	present := v.cam
	key := func(k tcell.Key, r rune, mod tcell.ModMask) bool {
		return v.HandleKey(tcell.NewEventKey(k, r, mod), f)
	}
	if !key(tcell.KeyLeft, 0, 0) || v.follow {
		t.Fatal("left did not scroll")
	}
	key(tcell.KeyHome, 0, 0)
	_ = draw(v, m, 100, 30, 0, mapmodel.TierUnicode, false)
	if v.cam != 0 {
		t.Errorf("Home: cam %d", v.cam)
	}
	key(tcell.KeyRune, 'L', 0)
	if v.cam != 50 {
		t.Errorf("L scrolls half a screen: cam %d", v.cam)
	}
	key(tcell.KeyEnd, 0, 0)
	_ = draw(v, m, 100, 30, 0, mapmodel.TierUnicode, false)
	if v.cam != present {
		t.Errorf("End: cam %d, present %d", v.cam, present)
	}
	for _, r := range "ifc" {
		if !key(tcell.KeyRune, r, 0) {
			t.Errorf("key %q unused", r)
		}
	}
	if !v.inspect || !v.flows || v.changes {
		t.Error("i f c did not toggle inspect, flows and changes")
	}
	if key(tcell.KeyRune, 'z', 0) {
		t.Error("an unbound key was used")
	}
	if !key(tcell.KeyEscape, 0, 0) || v.inspect {
		t.Error("Esc did not put the cursor away")
	}
}

// TestCopy: player-visible strings have no em dashes or exclamation marks.
func TestCopy(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	fset := token.NewFileSet()
	for _, fn := range files {
		if strings.HasSuffix(fn, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, fn, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if s, err := strconv.Unquote(lit.Value); err == nil && (strings.Contains(s, "—") || strings.Contains(s, "!")) {
					t.Errorf("%s: %q", fset.Position(lit.Pos()), s)
				}
			}
			return true
		})
	}
	if e := Entry(); e.Name != "skyline" || e.Title != "Skyline" || e.Blurb == "" || e.New() == nil {
		t.Errorf("bad entry %+v", e)
	}
}

func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("timing")
	}
	m := build(fixture.Options{Age: "cyberpunk_age", Seed: 7, Routes: 5, Tick: tickAt(0.9), Scale: 2}, 1)
	v := newView()
	scr := capture.NewScreen(200, 60)
	f := mapstyle.Frame{Model: m}
	v.Draw(scr, mapstyle.Rect{W: 200, H: 60}, f)
	const n = 30
	start := time.Now()
	for i := 0; i < n; i++ {
		f.Anim = i
		v.Draw(scr, mapstyle.Rect{W: 200, H: 60}, f)
	}
	per := time.Since(start) / n
	t.Logf("200x60 cyberpunk frame: %v", per)
	if per > 25*time.Millisecond {
		t.Errorf("a frame takes %v", per)
	}
}
