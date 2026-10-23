package skyline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// zoomed returns a view at zoom z, free of the present, whose view is
// centred on world column centre, and the frame composed at w by h.
func zoomed(m *mapmodel.Model, z zoom, centre, w, h int) (*view, *scene) {
	v := newView()
	v.zoom, v.sw = z, w
	v.follow, v.changes = false, false
	v.cam = centre - z.cols(w)/2
	return v, v.compose(mapstyle.Frame{Model: m, Anim: 3}, w, h)
}

func press(v *view, m *mapmodel.Model, k tcell.Key, r rune, mod tcell.ModMask) bool {
	return v.HandleKey(tcell.NewEventKey(k, r, mod), mapstyle.Frame{Model: m})
}

// TestZoomScale: a building is drawn twice as wide at close and half as
// wide at far as it is at mid, and the three levels show the same building.
func TestZoomScale(t *testing.T) {
	for _, age := range []string{"bronze_age", "medieval_age", "industrial_age"} {
		m := build(fixture.Options{Age: age, Seed: 5, Tick: tickAt(0.5)}, 0)
		_, s := zoomed(m, zMid, m.Skyline.Width/2, 160, 45)
		tried := 0
		for i := range s.lay.lots {
			lv := &s.lay.lots[i]
			if lv.ml.Wonder || lv.spr.h > 12 || lv.spr.w < 8 || lv.spr.w > 30 || !alone(s.lay, i, 6) || lv.ml.X < 100 || lv.ml.X > m.Skyline.Width-100 {
				continue
			}
			tried++
			width := map[zoom]int{}
			cells := map[zoom]int{}
			for _, z := range zoomSteps {
				v, _ := zoomed(m, z, lv.ml.X, 160, 45)
				v.noTraffic = true
				sc := v.compose(mapstyle.Frame{Model: m, Anim: 3}, 160, 45)
				lo, hi := 1<<30, -1
				// the columns the building and a margin of the world's own
				// columns stand on (nothing else is within that margin)
				x0, x1 := max(0, sc.sx(lv.x0-sc.cam-4)), min(sc.W, sc.sx(lv.x0-sc.cam+lv.spr.w+4)+1)
				for y := sc.top; y < sc.top+sc.S; y++ {
					for x := x0; x < x1; x++ {
						if c := sc.fb.at(x, y); c.d == lv.depth {
							lo, hi = min(lo, x), max(hi, x)
							cells[z]++
						}
					}
				}
				if hi < lo {
					t.Fatalf("%s: %s is not drawn at %s", age, lv.ml.Key, z)
				}
				width[z] = hi - lo + 1
			}
			if width[zClose] != 2*width[zMid] {
				t.Errorf("%s %s: drawn %d wide at mid, %d at close", age, lv.ml.Key, width[zMid], width[zClose])
			}
			// (at far a lamp post or the like in the cell of the building's edge
			// takes the cell's depth, so the edge cell may count as the lamp's)
			if w, half := width[zFar], (width[zMid]+1)/2; w < half-1 || w > half+1 {
				t.Errorf("%s %s: drawn %d wide at mid, %d at far", age, lv.ml.Key, width[zMid], w)
			}
			if cells[zClose] != 4*cells[zMid] {
				t.Errorf("%s %s: %d cells at mid, %d at close, not four times as many", age, lv.ml.Key, cells[zMid], cells[zClose])
			}
			if cells[zFar] >= cells[zMid] || cells[zFar] == 0 {
				t.Errorf("%s %s: %d cells at mid, %d at far", age, lv.ml.Key, cells[zMid], cells[zFar])
			}
			if tried == 3 {
				break
			}
		}
		if tried == 0 {
			t.Fatalf("%s: no lone building to measure", age)
		}
	}
}

// alone reports whether lot i has no other lot within gap columns of it.
func alone(l *layout, i, gap int) bool {
	a := &l.lots[i]
	for j := range l.lots {
		b := &l.lots[j]
		if j != i && b.x0 < a.x0+a.spr.w+gap && a.x0 < b.x0+b.spr.w+gap {
			return false
		}
	}
	return true
}

// TestZoomCentred: a zoom change keeps the view centred on the same column
// of the town, in and out and back.
func TestZoomCentred(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3}, 0)
	for _, w := range []int{160, 161, 100} {
		v := newView()
		v.sw = w
		_ = v.compose(mapstyle.Frame{Model: m}, w, 45)
		v.follow = false
		v.cam = m.Skyline.Width / 2
		centre := func() int { return v.cam + v.zoom.cols(w)/2 }
		want := centre()
		for i, k := range []struct {
			key  tcell.Key
			r    rune
			into zoom
		}{{tcell.KeyRune, '+', zClose}, {tcell.KeyRune, '-', zMid}, {tcell.KeyRune, '-', zFar}, {tcell.KeyRune, '=', zMid}, {tcell.KeyRune, '=', zClose}, {tcell.KeyPgUp, 0, zMid}} {
			if !press(v, m, k.key, k.r, 0) || v.zoom != k.into {
				t.Fatalf("w=%d step %d: zoom is %s, want %s", w, i, v.zoom, k.into)
			}
			if c := centre(); abs(c-want) > 1 {
				t.Errorf("w=%d step %d (%s): the view is centred on %d, was %d", w, i, v.zoom, c, want)
			}
			s := v.compose(mapstyle.Frame{Model: m}, w, 45) // the frame agrees
			if c := s.cam + s.lw/2; abs(c-want) > 1 {
				t.Errorf("w=%d step %d (%s): drawn centred on %d, was %d", w, i, v.zoom, c, want)
			}
		}
	}
	// following the present keeps following it, with the present at the edge
	v := newView()
	v.sw = 160
	_ = v.compose(mapstyle.Frame{Model: m}, 160, 45)
	for _, z := range []zoom{zClose, zFar, zMid} {
		v.setZoom(z)
		s := v.compose(mapstyle.Frame{Model: m}, 160, 45)
		if !v.follow || s.cam != min(presentCam(m, s.lw), max(0, m.Skyline.Width-s.lw)) {
			t.Errorf("%s: cam %d, follow %v: not on the present", z, s.cam, v.follow)
		}
	}
}

// TestZoomKeys: the zoom keys step far, mid, close and stop at the ends;
// Shift-arrows are the half screen; the sky arc, which has no zoom, leaves
// them to the prompt.
func TestZoomKeys(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3}, 0)
	v := newView()
	_ = v.compose(mapstyle.Frame{Model: m}, 100, 30)
	if v.ZoomLevel() != "mid" {
		t.Fatalf("a new view is at %s", v.ZoomLevel())
	}
	for _, k := range []struct {
		key  tcell.Key
		r    rune
		want string
	}{
		{tcell.KeyRune, '+', "close"}, {tcell.KeyRune, '+', "close"}, {tcell.KeyPgDn, 0, "close"},
		{tcell.KeyRune, '-', "mid"}, {tcell.KeyPgUp, 0, "far"}, {tcell.KeyRune, '-', "far"}, {tcell.KeyRune, '_', "far"},
		{tcell.KeyRune, '=', "mid"}, {tcell.KeyPgDn, 0, "close"},
	} {
		if !press(v, m, k.key, k.r, 0) {
			t.Errorf("key %v %q was not used", k.key, k.r)
		}
		if got := v.ZoomLevel(); got != k.want {
			t.Fatalf("after %v %q the zoom is %s, want %s", k.key, k.r, got, k.want)
		}
	}
	v.SetZoom("far")
	if v.ZoomLevel() != "far" {
		t.Errorf("SetZoom(far): %s", v.ZoomLevel())
	}
	v.SetZoom("nonsense")
	if v.ZoomLevel() != "far" {
		t.Errorf("an unknown level changed the zoom to %s", v.ZoomLevel())
	}
	v.SetZoom("")
	if v.ZoomLevel() != "mid" {
		t.Errorf("the empty level is mid, not %s", v.ZoomLevel())
	}
	// the arrows move the same four screen columns at every level
	for _, z := range zoomSteps {
		v.SetZoom(z.String())
		_ = v.compose(mapstyle.Frame{Model: m}, 100, 30)
		v.follow, v.cam = false, 600
		press(v, m, tcell.KeyRight, 0, 0)
		if got := (v.cam - 600) * 100 / z.cols(100); got != 4 {
			t.Errorf("%s: an arrow moves %d screen columns", z, got)
		}
		v.cam = 600
		press(v, m, tcell.KeyRight, 0, tcell.ModShift)
		if v.cam-600 != z.cols(100)/2 {
			t.Errorf("%s: Shift-Right moved %d columns of %d", z, v.cam-600, z.cols(100))
		}
	}
	// the sky arc has no zoom
	for _, age := range []string{"space_age", "quantum_age"} {
		sm := build(fixture.Options{Age: age, Seed: 3}, 0)
		sv := newView()
		_ = sv.compose(mapstyle.Frame{Model: sm}, 100, 30)
		for _, r := range "+-=_" {
			if press(sv, sm, tcell.KeyRune, r, 0) {
				t.Errorf("%s: %q was used in the sky arc", age, r)
			}
		}
		if press(sv, sm, tcell.KeyPgUp, 0, 0) || press(sv, sm, tcell.KeyPgDn, 0, 0) {
			t.Errorf("%s: PgUp or PgDn was used in the sky arc", age)
		}
	}
}

// TestDistrictKeys: [ and ] put the left edge of the view on the start of
// the district before and after, stopping at the ends.
func TestDistrictKeys(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3}, 0)
	ds := m.Skyline.Districts
	for _, z := range zoomSteps {
		v := newView()
		v.SetZoom(z.String())
		_ = v.compose(mapstyle.Frame{Model: m}, 160, 45)
		press(v, m, tcell.KeyHome, 0, 0)
		for i := range ds {
			if !press(v, m, tcell.KeyRune, ']', 0) || v.follow || v.cam != ds[i].X0 {
				t.Fatalf("%s: ] number %d put the view at %d (follow %v), want district %d's start %d", z, i+1, v.cam, v.follow, i, ds[i].X0)
			}
			s := v.compose(mapstyle.Frame{Model: m}, 160, 45) // the district is where the frame starts
			if want := min(ds[i].X0, max(0, m.Skyline.Width-s.lw)); s.cam != want {
				t.Errorf("%s: district %d drawn from %d, want %d", z, i, s.cam, want)
			}
			v.cam = ds[i].X0
		}
		last := v.cam
		press(v, m, tcell.KeyRune, ']', 0)
		if v.cam != last {
			t.Errorf("%s: ] past the last district moved the view to %d", z, v.cam)
		}
		for i := len(ds) - 2; i >= 0; i-- {
			if !press(v, m, tcell.KeyRune, '[', 0) || v.cam != ds[i].X0 {
				t.Fatalf("%s: [ put the view at %d, want district %d's start %d", z, v.cam, i, ds[i].X0)
			}
		}
		press(v, m, tcell.KeyRune, '[', 0)
		if v.cam != ds[0].X0 {
			t.Errorf("%s: [ before the first district moved the view to %d", z, v.cam)
		}
	}
	// from the present, [ goes to the district before the view's left edge
	v := newView()
	_ = v.compose(mapstyle.Frame{Model: m}, 160, 45)
	cam := v.cam
	press(v, m, tcell.KeyRune, '[', 0)
	if v.follow || v.cam >= cam || v.cam < 0 {
		t.Errorf("[ from the present: cam %d to %d, follow %v", cam, v.cam, v.follow)
	}
}

// TestOverviewBox: the bright box on the overview bar covers the part of the
// town in view at each zoom.
func TestOverviewBox(t *testing.T) {
	for _, ww := range []int{201, 1451, 4979} {
		for _, W := range []int{80, 160, 240} {
			for _, vw := range []int{W / 2, W, 2 * W} {
				for _, cam := range []int{0, 1, ww / 3, max(0, ww-vw)} {
					c0, c1 := overviewBox(cam, vw, ww, W)
					last := min(cam+vw, ww) - 1 // the last column in view
					switch {
					case c0 < 0 || c1 > W || c1 <= c0:
						t.Errorf("ww=%d W=%d vw=%d cam=%d: box [%d,%d) is out of the bar", ww, W, vw, cam, c0, c1)
					case c0 != cam*W/ww:
						t.Errorf("ww=%d W=%d vw=%d cam=%d: box [%d,%d) does not begin on the view's first column", ww, W, vw, cam, c0, c1)
					case c1 != last*W/ww+1:
						t.Errorf("ww=%d W=%d vw=%d cam=%d: box [%d,%d) does not end on the view's last column", ww, W, vw, cam, c0, c1)
					}
					for col := cam; col <= last; col++ { // every column in view is in the box
						if c := col * W / ww; c < c0 || c >= c1 {
							t.Fatalf("ww=%d W=%d vw=%d cam=%d: column %d is in cell %d, outside the box [%d,%d)", ww, W, vw, cam, col, c, c0, c1)
						}
					}
				}
			}
		}
	}
	m := build(fixture.Options{Age: "medieval_age", Seed: 3, Tick: tickAt(0.5)}, 1)
	ww := m.Skyline.Width
	for _, sz := range [][2]int{{160, 45}, {80, 24}, {240, 60}} {
		w, h := sz[0], sz[1]
		var spans []int
		for _, z := range zoomSteps {
			_, s := zoomed(m, z, ww/2, w, h)
			c0, c1 := overviewBox(s.cam, s.lw, ww, w)
			if s.ovBox != [2]int{c0, c1} {
				t.Errorf("%dx%d %s: the bar's box is %v, the view is at cells [%d,%d)", w, h, z, s.ovBox, c0, c1)
			}
			spans = append(spans, c1-c0)
			// the box is lit (a plain cell inside has the box's colour, one
			// outside has the bar's)
			litBg := s.fb.at((c0+c1)/2, h-2).bg
			if c0 > 0 && s.fb.at(0, h-2).bg == litBg {
				t.Errorf("%dx%d %s: the box looks like the rest of the bar", w, h, z)
			}
			if c1-c0 < w && s.fb.at(w-1, h-2).bg == litBg {
				t.Errorf("%dx%d %s: the box (%d of %d cells) is not lit differently from the bar", w, h, z, c1-c0, w)
			}
		}
		if !(spans[0] > spans[1] && spans[1] > spans[2]) {
			t.Errorf("%dx%d: the box does not narrow as the view zooms in: far, mid, close = %v", w, h, spans)
		}
	}
}

// TestOverviewShowsTheSkyline: the bar's columns are the town's own heights in
// the colour of the tops, the wonders are marked, districts are named where
// the name fits, and what is new is in the positive colour.
func TestOverviewShowsTheSkyline(t *testing.T) {
	// the present age's wonder is built, so the bar has a wonder to mark
	// (the skyline draws no other), and the town grows past the baseline
	b := mapmodel.NewBuilder(catalog())
	st := fixture.State(fixture.Options{Age: "medieval_age", Seed: 3, Tick: tickAt(0.5)})
	key := wonderKeyOf(catalog().AgeIdx["medieval_age"])
	bs := st.Buildings[key]
	bs.Count, bs.WonderBank, bs.WonderBankFull = 1, nil, false
	st.Buildings[key] = bs
	st.CurrentAgeWonderKey, st.CurrentAgeWonderName = "", ""
	since := mapmodel.VisitOf(b.Build(&st, nil))
	for i := 1; i <= 2; i++ {
		st = fixture.Grow(st, i)
	}
	m := b.Build(&st, since)
	v, _ := zoomed(m, zMid, m.Skyline.Width/2, 160, 45)
	v.changes = true
	s := v.compose(mapstyle.Frame{Model: m, Anim: 3}, 160, 45)
	W, ww, y := 160, s.lay.width, 43
	maxH := 1
	for _, h := range s.lay.prof {
		maxH = max(maxH, h)
	}
	wonders := map[int]bool{}
	for _, l := range s.lay.lots {
		if l.ml.Wonder {
			wonders[overviewCell(l.ml.X, ww, W)] = true
		}
	}
	if len(wonders) == 0 {
		t.Fatal("no wonder to mark in this model")
	}
	tops := map[tcell.Color]bool{}
	var text strings.Builder
	for x := 0; x < W; x++ {
		c := s.fb.at(x, y)
		text.WriteRune(c.ch)
		if wonders[x] {
			if c.ch != '◆' {
				t.Errorf("column %d holds a wonder and the bar shows %q", x, c.ch)
			}
			continue
		}
		if at := strings.IndexRune("▁▂▃▄▅▆▇", c.ch); at >= 0 {
			i := at / len("▁")
			a, b := overviewSpan(x, ww, W)
			h := 0
			for k := a; k < b; k++ {
				h = max(h, s.lay.prof[k])
			}
			want := min(7, int((float64(h)/float64(maxH))*8+0.999))
			if got := i + 1; got != want {
				t.Errorf("column %d: the town is %d of %d tall there, the bar shows eighth %d, want %d", x, h, maxH, got, want)
			}
			tops[c.fg] = true
		}
	}
	if len(tops) < 3 {
		t.Errorf("the bar has %d colours: the building tops are not in their own", len(tops))
	}
	named := 0
	for _, d := range m.Skyline.Districts {
		if strings.Contains(text.String(), ageWord(m.Catalog.AgeNames[d.Age])) {
			named++
		}
	}
	if named < 2 {
		t.Errorf("only %d districts are named on the bar: %q", named, text.String())
	}
	// a name never covers a wonder
	if strings.Count(text.String(), "◆") < len(wonders) {
		t.Errorf("a name covered a wonder: %q", text.String())
	}
	pos := role(theme.RolePositive, theme.Color(theme.RoleBackground))
	fresh := 0
	for x := 0; x < W; x++ {
		if c := s.fb.at(x, y); c.fg == theme.Legible(pos, c.bg, 3) {
			fresh++
		}
	}
	if m.Recap.NewCount > 0 && fresh == 0 {
		t.Errorf("%d new buildings and none on the bar", m.Recap.NewCount)
	}
}

// TestNameSpot: a name goes in the middle of its district, clear of marks
// and of the district's edges, or not at all.
func TestNameSpot(t *testing.T) {
	marked := make([]bool, 40)
	if x, ok := nameSpot(marked, 10, 30, 6); !ok || x != 17 {
		t.Errorf("a name in cells [10,30) goes at %d (%v), want the middle, 17", x, ok)
	}
	marked[19] = true
	x, ok := nameSpot(marked, 10, 30, 6)
	if !ok || x-1 <= 19 && 19 <= x+6 {
		t.Errorf("a name over a mark at 19: %d (%v)", x, ok)
	}
	if _, ok := nameSpot(marked, 10, 17, 6); ok {
		t.Error("a name that leaves no cell each side fitted")
	}
	for i := range marked {
		marked[i] = i%4 == 0
	}
	if _, ok := nameSpot(marked, 10, 30, 6); ok {
		t.Error("a name fitted between marks closer than itself")
	}
}

// TestZoomFooter: the footer says the zoom level and the keys at every zoom
// and size.
func TestZoomFooter(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3}, 0)
	for _, sz := range [][2]int{{80, 24}, {160, 45}, {240, 60}} {
		for _, z := range zoomSteps {
			v := newView()
			v.zoom = z
			txt := capture.Text(draw(v, m, sz[0], sz[1], 1, mapmodel.TierUnicode, false))
			rows := strings.Split(strings.TrimRight(txt, "\n"), "\n")
			foot := rows[len(rows)-1]
			if !strings.Contains(foot, "zoom: "+z.String()) {
				t.Errorf("%dx%d at %s: the footer is %q", sz[0], sz[1], z, foot)
			}
			if !strings.Contains(foot, "+ - zoom") {
				t.Errorf("%dx%d at %s: the footer does not name the zoom keys: %q", sz[0], sz[1], z, foot)
			}
		}
	}
}

// TestZoomInspect: Tab works at every zoom: the cursor stands on a building,
// stays in view, names it on the footer and on a label, and the arrows move it.
func TestZoomInspect(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3, Tick: tickAt(0.5)}, 0)
	for _, sz := range [][2]int{{160, 45}, {80, 24}} {
		w, h := sz[0], sz[1]
		for _, z := range zoomSteps {
			v := newView()
			v.SetZoom(z.String())
			f := mapstyle.Frame{Model: m, Anim: 5}
			_ = draw(v, m, w, h, 5, mapmodel.TierUnicode, false)
			if !v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, 0), f) || !v.inspect {
				t.Fatalf("%dx%d %s: Tab did not put the cursor out", w, h, z)
			}
			seen := map[string]bool{}
			for step := 0; step < 8; step++ {
				scr := draw(v, m, w, h, 5, mapmodel.TierUnicode, false)
				in, ok := v.Inspect(f)
				if !ok {
					t.Fatalf("%dx%d %s: no inspection at step %d", w, h, z, step)
				}
				seen[in.Title] = true
				txt := capture.Text(scr)
				foot := txt[strings.LastIndex(strings.TrimRight(txt, "\n"), "\n"):]
				if !strings.Contains(foot, in.Title) {
					t.Errorf("%dx%d %s: the footer %q does not name %q", w, h, z, foot, in.Title)
				}
				if v.cur.kind == tLot {
					lv := v.lay.lots[v.lay.find(v.cur.key, v.cur.cp)]
					if mid := lv.ml.X; mid < v.cam || mid >= v.cam+v.lastW {
						t.Errorf("%dx%d %s: %s at column %d is outside the view [%d,%d)", w, h, z, in.Title, mid, v.cam, v.cam+v.lastW)
					}
					if label := " " + lv.b.Name + " "; lv.b != nil && !lv.ml.Wonder && !strings.Contains(txt, label) {
						t.Errorf("%dx%d %s: no label %q on the screen", w, h, z, label)
					}
				}
				key := tcell.KeyRight
				if step%4 == 3 {
					key = tcell.KeyDown
				}
				v.HandleKey(tcell.NewEventKey(key, 0, 0), f)
			}
			if len(seen) < 3 {
				t.Errorf("%dx%d %s: the cursor visited only %v", w, h, z, seen)
			}
		}
	}
}

// TestZoomNothingOutside: the scene fills the rows between the header and the
// bars at every zoom and size, odd ones too, the bars are the bars, and a
// draw into a rect writes nothing outside it.
func TestZoomNothingOutside(t *testing.T) {
	for _, age := range []string{"primitive_age", "industrial_age", "cyberpunk_age"} {
		m := build(fixture.Options{Age: age, Seed: 3, Routes: 3, Tick: tickAt(0.9)}, 1)
		for _, sz := range [][2]int{{80, 24}, {60, 16}, {61, 17}, {101, 31}, {160, 45}, {241, 61}} {
			w, h := sz[0], sz[1]
			for _, z := range zoomSteps {
				v := newView()
				v.zoom, v.changes = z, false
				s := v.compose(mapstyle.Frame{Model: m, Anim: 9}, w, h)
				if s.fb.w != w || s.fb.h != h || len(s.fb.c) != w*h {
					t.Fatalf("%s %dx%d %s: the frame is %dx%d", age, w, h, z, s.fb.w, s.fb.h)
				}
				for y := 0; y < h; y++ {
					for x := 0; x < w; x++ {
						c := s.fb.at(x, y)
						switch bar := y == 0 || y >= h-2; {
						case c.d == dEmpty:
							t.Fatalf("%s %dx%d %s: cell %d,%d was never drawn", age, w, h, z, x, y)
						case bar && c.d != dTop:
							t.Fatalf("%s %dx%d %s: bar row %d has scene art at column %d (depth %d)", age, w, h, z, y, x, c.d)
						case !bar && c.d == dTop:
							t.Fatalf("%s %dx%d %s: the scene has a bar cell at %d,%d (%q)", age, w, h, z, x, y, c.ch)
						}
					}
				}
				// a rect inside a larger screen: nothing outside it changes
				scr := capture.NewScreen(w+6, h+4)
				sentinel := tcell.StyleDefault.Foreground(tcell.NewRGBColor(1, 2, 3)).Background(tcell.NewRGBColor(3, 2, 1))
				for y := 0; y < h+4; y++ {
					for x := 0; x < w+6; x++ {
						scr.SetContent(x, y, 'S', nil, sentinel)
					}
				}
				v.Draw(scr, mapstyle.Rect{X: 3, Y: 2, W: w, H: h}, mapstyle.Frame{Model: m, Anim: 9})
				scr.Show()
				cells, sw, _ := scr.GetContents()
				for i, c := range cells {
					x, y := i%sw, i/sw
					inside := x >= 3 && x < 3+w && y >= 2 && y < 2+h
					same := len(c.Runes) == 1 && c.Runes[0] == 'S' && c.Style == sentinel
					if inside == same {
						t.Fatalf("%s %dx%d %s: cell %d,%d (inside %v) is %q", age, w, h, z, x, y, inside, c.Runes)
					}
				}
			}
		}
	}
}

// TestZoomTextIsNotScaled: text keeps its size at every zoom: the cursor's
// name is one run of single cells, never doubled.
func TestZoomTextIsNotScaled(t *testing.T) {
	m := build(fixture.Options{Age: "medieval_age", Seed: 3, Harbinger: true, Tick: tickAt(0.5)}, 0)
	for _, z := range zoomSteps {
		v := newView()
		v.SetZoom(z.String())
		v.inspect = true
		txt := capture.Text(draw(v, m, 160, 45, 4, mapmodel.TierUnicode, false))
		b := m.Building(v.cur.key)
		if b == nil {
			t.Fatalf("%s: the cursor stands on nothing", z)
		}
		name := " " + b.Name + " "
		if !strings.Contains(txt, name) {
			t.Errorf("%s: the cursor's label %q is not on the screen", z, name)
		}
		if strings.Contains(txt, strings.Repeat(string([]rune(name)[1]), 2)+strings.Repeat(string([]rune(name)[2]), 2)) {
			t.Errorf("%s: the label is doubled", z)
		}
	}
}

// TestZoomTransforms pins the cell maths: a block glyph magnifies exactly
// into four solid cells, other glyphs repeat, and shrinking keeps the nearest
// thing of each quadrant.
func TestZoomTransforms(t *testing.T) {
	red, blue := tcell.NewRGBColor(200, 0, 0), tcell.NewRGBColor(0, 0, 200)
	for _, tc := range []struct {
		ch   rune
		want [4]tcell.Color // upper left, upper right, lower left, lower right
	}{
		{'█', [4]tcell.Color{red, red, red, red}},
		{' ', [4]tcell.Color{blue, blue, blue, blue}},
		{'▀', [4]tcell.Color{red, red, blue, blue}},
		{'▄', [4]tcell.Color{blue, blue, red, red}},
		{'▌', [4]tcell.Color{red, blue, red, blue}},
		{'▐', [4]tcell.Color{blue, red, blue, red}},
		{'▘', [4]tcell.Color{red, blue, blue, blue}},
		{'▟', [4]tcell.Color{blue, red, red, red}},
	} {
		for q := 0; q < 4; q++ {
			got := zoomCell(cell{ch: tc.ch, fg: red, bg: blue, d: 20}, q)
			if gotc := got.show(); gotc != tc.want[q] || got.d != 20 {
				t.Errorf("%q quadrant %d magnifies to %q on %v, want %v", tc.ch, q, got.ch, gotc, tc.want[q])
			}
		}
	}
	if g := zoomCell(cell{ch: '╱', fg: red, bg: blue, d: 20}, 3); g.ch != '╱' || g.fg != red {
		t.Errorf("a line glyph is not repeated: %q", g.ch)
	}

	// magnify: the foot of the world frame, bottom-aligned, each cell 2x2
	src := &fb{}
	src.reset(3, 3+4) // 4 scene rows, each a colour of its own
	for y := 0; y < 4; y++ {
		for x := 0; x < 3; x++ {
			src.set(x, y+1, '█', tcell.NewRGBColor(int32(y*10), int32(x*10), 0), 0, 10)
		}
	}
	dst := &fb{}
	dst.reset(6, 3+5) // 5 scene rows: two and a half world rows
	magnify(dst, src, 0)
	for r := 0; r < 5; r++ {
		wantRow := 3 - (4-r)/2 // the bottom screen row is half of the last world row
		for x := 0; x < 6; x++ {
			row, col, _ := dst.at(x, r+1).fg.RGB()
			if int(row) != wantRow*10 || int(col) != (x/2)*10 {
				t.Errorf("screen row %d column %d shows world row %d column %d, want row %d column %d", r, x, int(row)/10, int(col)/10, wantRow, x/2)
			}
		}
	}

	// shrink: a tower one cell wide keeps its colour against the sky
	world := &fb{}
	world.reset(4, 3+4)
	sky, tower := tcell.NewRGBColor(10, 10, 80), tcell.NewRGBColor(200, 200, 0)
	for y := 1; y <= 4; y++ {
		for x := 0; x < 4; x++ {
			world.set(x, y, ' ', sky, sky, dSky)
		}
	}
	world.set(1, 2, '█', tower, tower, dRow0)
	world.set(1, 3, '█', tower, tower, dRow0)
	scr := &fb{}
	scr.reset(2, 3+2)
	shrink(scr, world, false)
	if c := scr.at(0, 1); c.ch != '▗' && c.ch != '▐' && c.ch != '▝' || c.fg != tower || c.bg != sky || c.d != dRow0 {
		t.Errorf("the tower's top cell is %q fg %v bg %v depth %d", c.ch, c.fg, c.bg, c.d)
	}
	if c := scr.at(1, 1); c.ch != '█' || c.fg != sky || c.d != bgDepth(dSky) {
		t.Errorf("a sky cell is %q fg %v depth %d", c.ch, c.fg, c.d)
	}
	scr.reset(2, 3+2)
	shrink(scr, world, true)
	if c := scr.at(0, 1); c.ch != '█' || c.fg != tower {
		t.Errorf("ASCII tier: the tower's cell is %q fg %v", c.ch, c.fg)
	}
}

// TestZoomFrameGuard: a frame at every zoom stays inside the 25 ms guard at
// 160x45 and 240x60 for six ages (and logs the times).
func TestZoomFrameGuard(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, sz := range [][2]int{{160, 45}, {240, 60}} {
		for _, age := range []string{"bronze_age", "medieval_age", "industrial_age", "modern_age", "cyberpunk_age", "fusion_age"} {
			m := build(fixture.Options{Age: age, Seed: 7, Routes: 5, Tick: tickAt(0.9), Scale: 2}, 1)
			for _, z := range zoomSteps {
				v := newView()
				v.zoom = z
				scr := capture.NewScreen(sz[0], sz[1])
				f := mapstyle.Frame{Model: m}
				v.Draw(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, f)
				const n = 20
				start := time.Now()
				for i := 0; i < n; i++ {
					f.Anim = i
					v.Draw(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, f)
				}
				per := time.Since(start) / n
				t.Logf("%dx%d %-14s %-5s frame %v", sz[0], sz[1], age, z, per)
				if per > 25*time.Millisecond {
					t.Errorf("%dx%d %s at %s: a frame takes %v", sz[0], sz[1], age, z, per)
				}
			}
		}
	}
}

// TestZoomSweep draws ground ages at every zoom, glyph tier and theme
// without a panic, nothing '?', no glyph in its own colour, and in the ASCII
// tier nothing outside ASCII.
func TestZoomSweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	ages := []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "modern_age", "cyberpunk_age", "fusion_age"}
	themes := theme.All()
	for i, age := range ages {
		m := build(fixture.Options{Age: age, Seed: 4, Routes: 3, Harbinger: true, Tick: tickAt(0.95)}, 1)
		for _, z := range zoomSteps {
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
				if err := theme.SetActive(themes[(i+int(z))%len(themes)].Key); err != nil {
					t.Fatal(err)
				}
				v := newView()
				v.zoom, v.inspect, v.legend, v.flows = z, true, true, true
				scr := draw(v, m, 120, 36, 7, tier, false)
				cells, _, _ := scr.GetContents()
				for _, c := range cells {
					if len(c.Runes) == 0 {
						continue
					}
					r := c.Runes[0]
					fg, bg, _ := c.Style.Decompose()
					switch {
					case tier == mapmodel.TierASCII && r >= 0x80:
						t.Fatalf("%s %s ascii: non-ASCII rune %q", age, z, r)
					case tier != mapmodel.TierASCII && r == '?':
						t.Fatalf("%s %s %s: a '?' fallback leaked into the frame", age, z, tier)
					case r != ' ' && fg == bg:
						t.Fatalf("%s %s %s: %q has fg == bg", age, z, tier, r)
					}
				}
			}
		}
	}
}

// TestZoomCaptures writes review captures of the zoom levels when
// SKYLINE_CAPTURE names a directory.
func TestZoomCaptures(t *testing.T) {
	dir := os.Getenv("SKYLINE_CAPTURE")
	if dir == "" {
		t.Skip("set SKYLINE_CAPTURE to a directory to write captures")
	}
	shots := []struct {
		name    string
		age     string
		tod     float64
		w, h    int
		z       zoom
		inspect bool
	}{
		{"bronze", "bronze_age", 0.45, 160, 45, zFar, false},
		{"bronze", "bronze_age", 0.45, 160, 45, zMid, false},
		{"bronze", "bronze_age", 0.45, 160, 45, zClose, false},
		{"medieval", "medieval_age", 0.45, 160, 45, zFar, false},
		{"medieval", "medieval_age", 0.45, 160, 45, zClose, false},
		{"cyber", "cyberpunk_age", 0.95, 160, 45, zFar, false},
		{"cyber", "cyberpunk_age", 0.95, 160, 45, zClose, true},
		{"modern", "modern_age", 0.5, 160, 45, zFar, false},
		{"modern", "modern_age", 0.5, 160, 45, zClose, false},
		{"small", "renaissance_age", 0.5, 80, 24, zClose, false},
		{"small", "renaissance_age", 0.5, 80, 24, zFar, false},
	}
	for _, sh := range shots {
		m := build(fixture.Options{Age: sh.age, Seed: 3, Routes: 4, Tick: tickAt(sh.tod)}, 1)
		v := newView()
		v.zoom, v.inspect, v.sw = sh.z, sh.inspect, sh.w
		v.follow = false
		d := m.Skyline.Districts[len(m.Skyline.Districts)-1]
		v.cam = d.Centre - sh.z.cols(sh.w)/2
		var frames []string
		var txt string
		name := sh.name + "_" + sh.z.String()
		for k := 0; k < 6; k++ {
			scr := draw(v, m, sh.w, sh.h, 40+k*2, mapmodel.TierUnicode, false)
			frames = append(frames, capture.Frame(scr))
			if k == 0 {
				txt = capture.Text(scr)
				_ = capture.PNG(filepath.Join(dir, name+".png"), scr)
			}
		}
		bg, fg := capture.Hex(theme.Color(theme.RoleBackground)), capture.Hex(theme.Color(theme.RoleText))
		_ = os.WriteFile(filepath.Join(dir, name+".txt"), []byte(txt), 0o644)
		_ = os.WriteFile(filepath.Join(dir, name+".html"), []byte(capture.HTML(name, bg, fg, capture.Fonts, frames, 6)), 0o644)
	}
}
