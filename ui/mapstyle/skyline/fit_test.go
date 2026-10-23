package skyline

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// farScene composes the far level of m on a w by h screen.
func farScene(m *mapmodel.Model, w, h int) (*view, *scene) {
	v := newView()
	v.zoom = zFar
	return v, v.compose(mapstyle.Frame{Model: m, Anim: 5}, w, h)
}

// TestFarFitsTheTown: the far level shows the whole town on the screen's
// width, whatever its size and the screen's: every district is on the screen
// with its buildings standing in it, the town's east end is inside the last
// column, and nothing of the view scrolls. A town no wider than two screens
// has no far level of its own: it is the wide one.
func TestFarFitsTheTown(t *testing.T) {
	towns := 0
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "modern_age", "cyberpunk_age", "fusion_age"} {
		for _, scale := range []float64{0.5, 1, 2, 4} {
			m := build(fixture.Options{Age: age, Seed: 3, Routes: 2, Scale: scale, Tick: tickAt(0.5)}, 0)
			for _, sz := range [][2]int{{80, 24}, {120, 36}, {160, 45}, {241, 61}} {
				w, h := sz[0], sz[1]
				v, s := farScene(m, w, h)
				ww := m.Skyline.Width
				if ww <= 2*w {
					if s.fit || s.zm != zWide {
						t.Errorf("%s x%v %dx%d: a town of %d columns is %s (fit %v), not the wide level", age, scale, w, h, ww, s.zm, s.fit)
					}
					if v.fitOK(w) {
						t.Errorf("%s x%v %dx%d: far is a level of its own for a town of %d columns", age, scale, w, h, ww)
					}
					continue
				}
				towns++
				if !s.fit || s.zm != zFar || s.lw != ww || s.cam != 0 || v.lfb.w != 2*w {
					t.Fatalf("%s x%v %dx%d: far shows %d columns from %d in a frame %d wide (fit %v), the town is %d", age, scale, w, h, s.lw, s.cam, v.lfb.w, s.fit, ww)
				}
				if got := s.wsx(ww - 1); got != w-1 {
					t.Errorf("%s x%v %dx%d: the town's last column lands on screen column %d, not %d", age, scale, w, h, got, w-1)
				}
				if got := s.wsx(0); got != 0 {
					t.Errorf("%s x%v %dx%d: the town's first column lands on %d", age, scale, w, h, got)
				}
				// every district has buildings drawn across its stretch of screen
				for di, d := range m.Skyline.Districts {
					a, b := s.wsx(d.X0), s.wsx(d.X0+d.W)
					if b <= a {
						continue // a district that is narrower than a screen column
					}
					ink := 0
					for y := s.top; y < s.top+s.S; y++ {
						for x := a; x < b && x < w; x++ {
							if c := s.fb.at(x, y); c.d >= dWonder && c.d <= dRow2 {
								ink++
							}
						}
					}
					lots := 0
					for _, l := range m.Skyline.Lots {
						if l.X >= d.X0 && l.X < d.X0+d.W {
							lots++
						}
					}
					if lots > 3 && ink == 0 {
						t.Errorf("%s x%v %dx%d: district %d (%d lots) shows no buildings on the far level", age, scale, w, h, di, lots)
					}
				}
				// the arrows and the district keys leave the view alone: it shows everything
				v.follow, v.cam = false, 7
				for _, k := range []tcell.Key{tcell.KeyLeft, tcell.KeyRight} {
					press(v, m, k, 0, 0)
					press(v, m, k, 0, tcell.ModShift)
				}
				press(v, m, tcell.KeyRune, ']', 0)
				if v.cam != 7 || v.follow {
					t.Errorf("%s x%v %dx%d: an arrow moved the far view (cam %d, follow %v)", age, scale, w, h, v.cam, v.follow)
				}
			}
		}
	}
	if towns < 20 {
		t.Fatalf("only %d towns were wide enough to fit: the table tests nothing", towns)
	}
}

// TestFarReadsTheTallest: each column of the far frame stands for a run of
// the town's columns and is as tall as the tallest of them, in the colours
// of that building.
func TestFarReadsTheTallest(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3, Scale: 2, Tick: tickAt(0.5)}, 0)
	v, s := farScene(m, 160, 45)
	cols := s.lay.fitCols(v.lfb.w)
	if len(cols) != v.lfb.w || v.lfb.w != 320 {
		t.Fatalf("%d far columns for a frame %d wide", len(cols), v.lfb.w)
	}
	for x, c := range cols {
		a := int(float64(x) * s.kx)
		b := max(a+1, int(float64(x+1)*s.kx))
		h := 0
		for w := a; w < b && w < len(s.lay.prof); w++ {
			h = max(h, s.lay.prof[w])
		}
		if int(c.h) != h {
			t.Fatalf("far column %d (world %d to %d): %d tall, the town stands %d there", x, a, b, c.h, h)
		}
		if h > 0 {
			lv := &s.lay.lots[c.lot]
			if sx := int(c.at) - lv.x0; sx < 0 || sx >= lv.spr.w || lv.spr.topAt(sx) < 0 {
				t.Fatalf("far column %d reads %s at column %d, which holds nothing of it", x, lv.ml.Key, c.at)
			}
		}
	}
}

// TestFarMarksTheWonder: this age's wonder is starred on the far level
// however low it stands among the towers round it, and a wonder going up is
// starred too.
func TestFarMarksTheWonder(t *testing.T) {
	for _, age := range []string{"medieval_age", "modern_age", "fusion_age"} {
		for _, built := range []bool{true, false} {
			m := wonderModel(fixture.Options{Age: age, Seed: 4, Scale: 3, Tick: tickAt(0.5)}, built, 0.5)
			_, s := farScene(m, 160, 45)
			x := s.wsx(wonderX(m.Skyline.Districts[m.AgeIdx], wonderKeyOf(m.AgeIdx)))
			star := mapmodel.R(mapmodel.SymWonder, s.tier)
			found := false
			for y := s.top; y < s.top+s.S && !found; y++ {
				for dx := -1; dx <= 1; dx++ {
					if c := s.fb.at(x+dx, y); c != nil && c.ch == star && c.d == dTag {
						found = true
					}
				}
			}
			if !found {
				t.Errorf("%s built=%v: no star over the wonder at column %d", age, built, x)
			}
		}
	}
}

// TestFarCursor: Tab works on the far level too: the cursor's name and bar
// stand on the screen at the lot's own place.
func TestFarCursor(t *testing.T) {
	m := build(fixture.Options{Age: "medieval_age", Seed: 3, Scale: 2, Tick: tickAt(0.5)}, 0)
	v := newView()
	v.zoom, v.inspect = zFar, true
	f := mapstyle.Frame{Model: m, Anim: 5}
	_ = v.compose(f, 160, 45)
	for i := 0; i < 40; i++ {
		press(v, m, tcell.KeyTab, 0, 0)
	}
	s := v.compose(f, 160, 45)
	if !v.follow {
		t.Errorf("Tab at the far level stopped the view following the present")
	}
	if s.sel < 0 {
		t.Fatalf("the cursor is on no lot")
	}
	lv := &s.lay.lots[s.sel]
	x0 := s.wsx(lv.x0 - s.cam)
	x1 := max(x0, s.wsx(lv.x0-s.cam+lv.spr.w)-1)
	row := s.Y(s.sy(s.groundY + 2))
	for _, x := range []int{x0, x1} {
		if c := s.fb.at(x, row); c == nil || c.d != dTop {
			t.Errorf("no cursor bar under %s at column %d (it stands on %d to %d)", lv.ml.Key, x, x0, x1)
		}
	}
}

// TestPanKeys: at the close level the up and down arrows raise and lower the
// window over the scene, a step at a time (Shift a half), stop at the foot
// of the scene and at its top, and Home brings the window back to the foot
// before it does anything else. The cursor keeps the arrows when it is out,
// and at the other levels they are not taken.
func TestPanKeys(t *testing.T) {
	m := build(fixture.Options{Age: "modern_age", Seed: 3, Scale: 2, Tick: tickAt(0.5)}, 0)
	for _, sz := range [][2]int{{80, 24}, {160, 45}, {241, 61}} {
		w, h := sz[0], sz[1]
		v := newView()
		v.SetZoom("close")
		v.follow, v.cam = false, m.Skyline.Width/2
		f := mapstyle.Frame{Model: m, Anim: 5}
		s := v.compose(f, w, h)
		if v.panLo != 0 || v.panHi < 3 {
			t.Fatalf("%dx%d: the window may be raised %d to %d rows", w, h, v.panLo, v.panHi)
		}
		if s.pan != 0 {
			t.Fatalf("%dx%d: the window starts raised by %d", w, h, s.pan)
		}
		if !press(v, m, tcell.KeyUp, 0, 0) || v.pan != panStep {
			t.Fatalf("%dx%d: Up moved the window to %d, not %d", w, h, v.pan, panStep)
		}
		if s = v.compose(f, w, h); s.pan != panStep {
			t.Errorf("%dx%d: the frame is raised by %d after one Up", w, h, s.pan)
		}
		press(v, m, tcell.KeyDown, 0, 0)
		press(v, m, tcell.KeyDown, 0, 0)
		if v.pan != 0 {
			t.Errorf("%dx%d: Down went below the foot of the scene: %d", w, h, v.pan)
		}
		for i := 0; i < 200; i++ {
			press(v, m, tcell.KeyUp, 0, 0)
		}
		if v.pan != v.panHi {
			t.Errorf("%dx%d: Up stopped at %d, the top is %d", w, h, v.pan, v.panHi)
		}
		s = v.compose(f, w, h)
		if s.pan != s.ls-1-(s.ls-1)/2 {
			t.Errorf("%dx%d: at the top the window is raised by %d", w, h, s.pan)
		}
		// the top of the scene is what the screen's first scene row shows
		if got := s.sy(0); got != 0 && got != -1 {
			t.Errorf("%dx%d: at the top, scene row 0 is on screen row %d", w, h, got)
		}
		press(v, m, tcell.KeyDown, 0, tcell.ModShift)
		if v.pan >= v.panHi || v.pan <= 0 {
			t.Errorf("%dx%d: Shift-Down left the window at %d of %d", w, h, v.pan, v.panHi)
		}
		cam := v.cam
		press(v, m, tcell.KeyHome, 0, 0)
		if v.pan != 0 || v.cam != cam {
			t.Errorf("%dx%d: Home left pan %d, cam %d (was %d): it brings the window to the foot first", w, h, v.pan, v.cam, cam)
		}
		press(v, m, tcell.KeyHome, 0, 0)
		if v.cam != 0 {
			t.Errorf("%dx%d: a second Home did not go to the oldest district", w, h)
		}
		// the cursor out: the arrows are its, and it brings the window home
		press(v, m, tcell.KeyUp, 0, 0)
		press(v, m, tcell.KeyTab, 0, 0)
		if v.pan != 0 {
			t.Errorf("%dx%d: the cursor came out with the window still raised by %d", w, h, v.pan)
		}
		pan := v.pan
		press(v, m, tcell.KeyUp, 0, 0)
		if v.pan != pan {
			t.Errorf("%dx%d: Up raised the window while the cursor is out", w, h)
		}
		// other levels: not taken
		for _, z := range []zoom{zMid, zWide, zFar} {
			v.inspect = false
			v.SetZoom(z.String())
			v.compose(f, w, h)
			if press(v, m, tcell.KeyUp, 0, 0) || press(v, m, tcell.KeyDown, 0, 0) {
				t.Errorf("%dx%d %s: Up or Down was taken with nothing to pan", w, h, z)
			}
		}
	}
}

// TestPanShowsTheTower: a tower the close level crops at the top is whole
// once the window is raised, and the marker that stands over it follows.
func TestPanShowsTheTower(t *testing.T) {
	m := build(fixture.Options{Age: "fusion_age", Seed: 3, Scale: 2, Tick: tickAt(0.5)}, 0)
	v := newView()
	v.SetZoom("close")
	f := mapstyle.Frame{Model: m, Anim: 5}
	v.follow, v.cam = false, 0
	s := v.compose(f, 160, 45)
	best := 0
	for i := range s.lay.lots {
		if s.lay.lots[i].spr.h > s.lay.lots[best].spr.h {
			best = i
		}
	}
	lv := &s.lay.lots[best]
	v.cam = lv.ml.X - s.lw/2
	s = v.compose(f, 160, 45)
	before := s.sy(s.groundY - lv.spr.h)
	if before >= 0 {
		t.Fatalf("the tallest tower (%d rows) is not cropped at the close level on 45 rows", lv.spr.h)
	}
	for i := 0; i < 50; i++ {
		press(v, m, tcell.KeyUp, 0, 0)
	}
	s = v.compose(f, 160, 45)
	if top := s.sy(s.groundY - lv.spr.h); top <= before || top < -2 {
		t.Errorf("raising the window left the tower's roof at screen row %d (it was %d)", top, before)
	}
	// and the roof is drawn there: the tower's own cells fill the screen's top rows
	n := 0
	for x := 0; x < s.W; x++ {
		if c := s.fb.at(x, s.Y(1)); c != nil && c.d >= dWonder && c.d <= dRow2 {
			n++
		}
	}
	if n == 0 {
		t.Errorf("no building is drawn on the screen's second scene row with the window at the top")
	}
}

// TestPanBar: the close level draws where the window stands as a bar down the
// right edge when there is more scene than window, and no other level does.
func TestPanBar(t *testing.T) {
	m := build(fixture.Options{Age: "industrial_age", Seed: 3, Tick: tickAt(0.5)}, 0)
	bar := func(s *scene) (thumb, track int, first, last int) {
		first, last = -1, -1
		for y := s.top; y < s.top+s.S; y++ {
			switch c := s.fb.at(s.W-1, y); c.ch {
			case '▐':
				if c.d == dBar {
					thumb++
					if first < 0 {
						first = y - s.top
					}
					last = y - s.top
				}
			case '▕':
				if c.d == dBar {
					track++
				}
			}
		}
		return
	}
	v := newView()
	v.SetZoom("close")
	v.follow, v.cam = false, 300
	f := mapstyle.Frame{Model: m, Anim: 5}
	s := v.compose(f, 160, 45)
	thumb, track, first, last := bar(s)
	if thumb == 0 || track == 0 || thumb+track != s.S {
		t.Fatalf("the bar has %d thumb and %d track cells for %d rows", thumb, track, s.S)
	}
	if last != s.S-1 {
		t.Errorf("at the foot of the scene the thumb ends at row %d of %d", last, s.S)
	}
	for i := 0; i < 100; i++ {
		press(v, m, tcell.KeyUp, 0, 0)
	}
	s = v.compose(f, 160, 45)
	if _, _, first2, _ := bar(s); first2 != 0 || first2 >= first {
		t.Errorf("at the top of the scene the thumb starts at row %d (it was %d)", first2, first)
	}
	for _, z := range []zoom{zMid, zWide, zFar} {
		v.SetZoom(z.String())
		s = v.compose(f, 160, 45)
		if thumb, track, _, _ := bar(s); thumb+track != 0 {
			t.Errorf("%s draws a pan bar", z)
		}
	}
}
