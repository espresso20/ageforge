package skyline

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/ui/mapstyle/styletest"
)

// sceneAges are the sky ages that have a panorama: all of them but the
// Transcendent Age's mandala.
var sceneAges = skyAges[:4]

// TestSkyZoomLevels: the sky arc's scenes zoom like the ground: wide shows
// two screens of the panorama, close half of one at twice the size with the
// baseline in the window, and far fits all of it, every district standing in
// its stretch of the screen.
func TestSkyZoomLevels(t *testing.T) {
	for _, age := range sceneAges {
		m := skyModel(age, 3, 0)
		for _, sz := range [][2]int{{100, 30}, {160, 45}, {241, 61}} {
			w, h := sz[0], sz[1]
			ww := m.Skyline.Width
			for _, z := range zoomSteps {
				v := newView()
				v.zoom = z
				v.follow, v.cam = false, ww/2
				s := v.compose(mapstyle.Frame{Model: m, Anim: 5}, w, h)
				switch z {
				case zMid:
					if s.lw != w || s.zm != zMid {
						t.Errorf("%s %dx%d mid: shows %d columns", age, w, h, s.lw)
					}
				case zWide:
					if s.lw != 2*w || s.zm != zWide {
						t.Errorf("%s %dx%d wide: shows %d columns, want %d", age, w, h, s.lw, 2*w)
					}
				case zClose:
					if s.lw != (w+1)/2 || s.zm != zClose {
						t.Errorf("%s %dx%d close: shows %d columns", age, w, h, s.lw)
					}
					if y := s.sy(s.groundY); y < 0 || y >= s.ls {
						t.Errorf("%s %dx%d close: the baseline is on scene row %d of %d", age, w, h, y, s.ls)
					}
					if v.panHi <= v.panLo || s.pan <= 0 {
						t.Errorf("%s %dx%d close: pan %d in %d to %d: nothing to pan over", age, w, h, s.pan, v.panLo, v.panHi)
					}
				case zFar:
					if ww <= 2*w {
						if s.fit {
							t.Errorf("%s %dx%d far: fits a town of %d columns that wide already shows", age, w, h, ww)
						}
						continue
					}
					if !s.fit || s.lw != ww || s.cam != 0 {
						t.Fatalf("%s %dx%d far: shows %d of %d columns from %d (fit %v)", age, w, h, s.lw, ww, s.cam, s.fit)
					}
					if got := s.wsx(ww - 1); got != w-1 {
						t.Errorf("%s %dx%d far: the panorama's last column is on screen column %d", age, w, h, got)
					}
					for di, d := range m.Skyline.Districts {
						a, b := s.wsx(d.X0), s.wsx(d.X0+d.W)
						if b <= a+1 {
							continue
						}
						lots := 0
						for _, l := range m.Skyline.Lots {
							if l.X >= d.X0 && l.X < d.X0+d.W && !l.Wonder {
								lots++
							}
						}
						ink := 0
						for y := s.top; y < s.top+s.S; y++ {
							for x := a; x < b && x < w; x++ {
								if c := s.fb.at(x, y); c.d >= dWonder && c.d <= dRow2 {
									ink++
								}
							}
						}
						if lots > 3 && ink == 0 {
							t.Errorf("%s %dx%d far: district %d (%d lots) has nothing drawn in it", age, w, h, di, lots)
						}
					}
				}
				// the bars are the bars at every level, and nothing is left undrawn
				if s.fb.w != w || s.fb.h != h {
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
			}
		}
	}
}

// TestSkyZoomScale: a form is drawn twice as wide at close as at mid and half
// as wide at wide, in the sky as on the ground.
func TestSkyZoomScale(t *testing.T) {
	for _, age := range sceneAges {
		m := skyModel(age, 3, 0)
		_, base := zoomed(m, zMid, m.Skyline.Width/2, 160, 45)
		tried := 0
		for i := range base.lay.lots {
			lv := &base.lay.lots[i]
			if lv.ml.Wonder || lv.spr.w < 8 || lv.spr.w > 30 || !alone(base.lay, i, 6) || lv.ml.X < 100 || lv.ml.X > m.Skyline.Width-100 {
				continue
			}
			width := map[zoom]int{}
			for _, z := range walkSteps {
				v, _ := zoomed(m, z, lv.ml.X, 160, 45)
				v.noTraffic = true
				sc := v.compose(mapstyle.Frame{Model: m, Anim: 3}, 160, 45)
				lo, hi := 1<<30, -1
				x0, x1 := max(0, sc.sx(lv.x0-sc.cam-4)), min(sc.W, sc.sx(lv.x0-sc.cam+lv.spr.w+4)+1)
				for y := sc.top; y < sc.top+sc.S; y++ {
					for x := x0; x < x1; x++ {
						if c := sc.fb.at(x, y); c.d == lv.depth {
							lo, hi = min(lo, x), max(hi, x)
						}
					}
				}
				if hi < lo {
					width[z] = 0
					continue
				}
				width[z] = hi - lo + 1
			}
			if width[zMid] == 0 || width[zClose] == 0 || width[zWide] == 0 {
				continue
			}
			tried++
			if width[zClose] != 2*width[zMid] {
				t.Errorf("%s %s: drawn %d wide at mid, %d at close", age, lv.ml.Key, width[zMid], width[zClose])
			}
			if w, half := width[zWide], (width[zMid]+1)/2; w < half-1 || w > half+1 {
				t.Errorf("%s %s: drawn %d wide at mid, %d at wide", age, lv.ml.Key, width[zMid], w)
			}
			if tried == 2 {
				break
			}
		}
		if tried == 0 {
			t.Errorf("%s: no lone form to measure", age)
		}
	}
}

// TestSkyPanKeys: in the sky scenes the close level pans as the ground's
// does: up and down raise and lower the window, clamped at the scene's top
// and its bottom, Home brings it back to where it starts, and the other
// levels leave the arrows alone.
func TestSkyPanKeys(t *testing.T) {
	for _, age := range sceneAges {
		m := skyModel(age, 3, 0)
		v := newView()
		v.SetZoom("close")
		v.follow, v.cam = false, m.Skyline.Width/2
		f := mapstyle.Frame{Model: m, Anim: 5}
		s := v.compose(f, 160, 45)
		start := s.pan
		if v.panLo >= 0 || v.panHi <= 0 {
			t.Fatalf("%s: the window can be moved %d to %d from where it starts", age, v.panLo, v.panHi)
		}
		for i := 0; i < 200; i++ {
			press(v, m, tcell.KeyUp, 0, 0)
		}
		if v.pan != v.panHi {
			t.Errorf("%s: Up stopped at %d of %d", age, v.pan, v.panHi)
		}
		if s = v.compose(f, 160, 45); s.pan != s.ls-1-(s.ls-1)/2 {
			t.Errorf("%s: at the top the window is raised by %d", age, s.pan)
		}
		for i := 0; i < 200; i++ {
			press(v, m, tcell.KeyDown, 0, 0)
		}
		if v.pan != v.panLo {
			t.Errorf("%s: Down stopped at %d of %d", age, v.pan, v.panLo)
		}
		if s = v.compose(f, 160, 45); s.pan != 0 {
			t.Errorf("%s: at the bottom the window is raised by %d", age, s.pan)
		}
		press(v, m, tcell.KeyUp, 0, 0)
		press(v, m, tcell.KeyHome, 0, 0)
		if s = v.compose(f, 160, 45); v.pan != 0 || s.pan != start {
			t.Errorf("%s: Home left the window at %d (pan %d), it starts at %d", age, s.pan, v.pan, start)
		}
		v.SetZoom("mid")
		v.compose(f, 160, 45)
		if press(v, m, tcell.KeyUp, 0, 0) {
			t.Errorf("%s: Up was taken at mid", age)
		}
	}
}

// TestMandalaZoom: the Transcendent Age closes in on the middle of its mandala
// and back; close in, the arrows pan the window along and up while the cursor
// is away (clamped), Home brings it back to the middle before it goes to the
// core, the cursor takes the arrows when it is out and the window goes to it,
// and there is no level wider than the whole mandala.
func TestMandalaZoom(t *testing.T) {
	m := skyModel("transcendent_age", 3, 1)
	const w, h = 120, 40
	v := newView()
	f := mapstyle.Frame{Model: m, Anim: 5}
	s := v.compose(f, w, h)
	if s.zm != zMid || v.panHi != v.panLo && v.zoom == zClose {
		t.Fatalf("the mandala starts at %s", s.zm)
	}
	// the wide levels are the whole mandala
	for _, z := range []zoom{zWide, zFar} {
		v.zoom = z
		if s = v.compose(f, w, h); s.zm != zMid {
			t.Errorf("%s: the mandala is drawn at %s", z, s.zm)
		}
	}
	press(v, m, tcell.KeyRune, '+', 0)
	if v.ZoomLevel() != "close" {
		t.Fatalf("+ zoomed to %s", v.ZoomLevel())
	}
	s = v.compose(f, w, h)
	if s.zm != zClose || v.panHi <= v.panLo || v.panXHi <= v.panXLo {
		t.Fatalf("close: zoom %s, window range %d..%d by %d..%d", s.zm, v.panLo, v.panHi, v.panXLo, v.panXHi)
	}
	// the window starts on the middle: the core's glyph is at the screen's
	core := mapmodel.R(mapmodel.SymCore, s.tier)
	cx, cy := w/2, 1+(h-2-1)/2
	found := false
	for dy := -2; dy <= 2 && !found; dy++ {
		for dx := -3; dx <= 3; dx++ {
			if c := s.fb.at(cx+dx, cy+dy); c != nil && c.ch == core {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("the core is not in the middle of the close window")
	}
	// the bars: right edge and bottom edge
	if c := s.fb.at(w-1, 2); c == nil || c.d != dBar {
		t.Errorf("no vertical bar at the right edge")
	}
	if c := s.fb.at(3, h-2); c == nil || c.d != dBar {
		t.Errorf("no horizontal bar on the bottom scene row")
	}
	if line := rowText(draw(v, m, w, h, 5, mapmodel.TierUnicode, false), h-1); !strings.Contains(line, "◄► ↑↓ pan") || !strings.Contains(line, "+ - zoom") {
		t.Errorf("the status line does not say the arrows pan close in: %q", line)
	}
	for i := 0; i < 100; i++ {
		press(v, m, tcell.KeyUp, 0, 0)
		press(v, m, tcell.KeyRight, 0, 0)
	}
	if v.pan != v.panHi || v.panX != v.panXHi || v.inspect {
		t.Errorf("Up and Right ended at %d, %d of %d, %d (cursor out %v)", v.pan, v.panX, v.panHi, v.panXHi, v.inspect)
	}
	for i := 0; i < 100; i++ {
		press(v, m, tcell.KeyDown, 0, 0)
		press(v, m, tcell.KeyLeft, 0, 0)
	}
	if v.pan != v.panLo || v.panX != v.panXLo {
		t.Errorf("Down and Left ended at %d, %d of %d, %d", v.pan, v.panX, v.panLo, v.panXLo)
	}
	press(v, m, tcell.KeyHome, 0, 0)
	if v.pan != 0 || v.panX != 0 || v.inspect {
		t.Errorf("Home left pan %d,%d (cursor out %v): it brings the window back first", v.pan, v.panX, v.inspect)
	}
	press(v, m, tcell.KeyHome, 0, 0)
	if !v.inspect || v.cur.kind != tCore {
		t.Errorf("a second Home did not put the cursor on the core: %v", v.cur)
	}
	// the cursor is out: the arrows are its, and Tab to the farthest ring
	// brings the window to it
	before := v.pan
	press(v, m, tcell.KeyUp, 0, 0)
	if v.pan != before {
		t.Errorf("Up panned the window with the cursor out")
	}
	ts := mandalaTargets(m)
	v.cur, v.reveal = ts[len(ts)-1].target, true // the outermost ring's
	s = v.compose(f, w, h)
	rings := m.Mandala()
	l := v.mandalaFor(len(rings), w, h-2)
	dx, dy, ok := mandalaSpot(&l.g, m, rings, v.cur)
	if !ok {
		t.Fatalf("the cursor's target %v has no place", v.cur)
	}
	px, py := s.sx(l.cx+dx), s.sy(l.cy+dy)
	if px < 0 || px >= w || py < 0 || py >= h-2 {
		t.Errorf("the cursor's target is off the close window, at screen %d,%d", px, py)
	}
	if v.pan == 0 && v.panX == 0 {
		t.Errorf("the window stayed on the middle for the outermost ring")
	}
	// back out
	v.inspect = false
	press(v, m, tcell.KeyRune, '-', 0)
	if v.ZoomLevel() != "mid" {
		t.Errorf("- left the zoom at %s", v.ZoomLevel())
	}
	if press(v, m, tcell.KeyRune, '-', 0); v.ZoomLevel() != "mid" {
		t.Errorf("- went wider than the whole mandala: %s", v.ZoomLevel())
	}
}

// TestSkyZoomFrameGuard: a sky frame at every zoom stays inside the 25 ms
// guard at 160x45 and 240x60 (and logs the times).
func TestSkyZoomFrameGuard(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, sz := range [][2]int{{160, 45}, {240, 60}} {
		for _, age := range skyAges {
			m := skyModel(age, 7, 1)
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
				t.Logf("%dx%d %-17s %-5s frame %v", sz[0], sz[1], age, z, per)
				if per > 25*time.Millisecond {
					t.Errorf("%dx%d %s at %s: a frame takes %v", sz[0], sz[1], age, z, per)
				}
			}
		}
	}
}

// TestSkyZoomSweep draws every sky age at every zoom, glyph tier and theme
// with the cursor out, without a panic, nothing '?', no glyph in its own
// colour, and in the ASCII tier nothing outside ASCII.
func TestSkyZoomSweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	themes := theme.All()
	for i, age := range skyAges {
		m := skyModel(age, 4, 1)
		for _, z := range zoomSteps {
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
				if err := theme.SetActive(themes[(i+int(z))%len(themes)].Key); err != nil {
					t.Fatal(err)
				}
				v := newView()
				v.zoom, v.inspect, v.legend, v.flows = z, true, true, true
				for _, sz := range [][2]int{{120, 36}, {80, 24}} {
					scr := draw(v, m, sz[0], sz[1], 7, tier, false)
					cells, _, _ := scr.GetContents()
					for _, c := range cells {
						if len(c.Runes) == 0 {
							continue
						}
						if f := queryRules.Fault(c.Runes[0], tier); f != "" {
							t.Fatalf("%s %s %s: rune %q: %s", age, z, tier, c.Runes[0], f)
						}
						if styletest.HiddenInk(c) {
							t.Fatalf("%s %s %s: %q has fg == bg", age, z, tier, c.Runes[0])
						}
					}
				}
			}
		}
	}
}
