package skyline

import (
	"strconv"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// buildings_test.go pins the skyline's buildings: a silhouette for every
// building (up to a cap), districts as wide as what they hold, an outline on
// every silhouette, the back rows dimmer than the front, variants that
// differ between copies and hold for one, lit windows that flicker and the
// frame budget.

func TestSkyCopiesFollowTheCount(t *testing.T) {
	if mapmodel.SkyCopies(0) != 0 || mapmodel.SkyCopies(-3) != 0 {
		t.Error("no buildings, no silhouettes")
	}
	prev := 0
	for n := 1; n <= 200; n++ {
		want := (n + mapmodel.SkyCopiesPer - 1) / mapmodel.SkyCopiesPer
		if want > mapmodel.SkyMaxCopies {
			want = mapmodel.SkyMaxCopies
		}
		got := mapmodel.SkyCopies(n)
		if got != want {
			t.Fatalf("%d buildings draw %d silhouettes, want %d", n, got, want)
		}
		if got < prev {
			t.Fatalf("%d buildings draw fewer silhouettes (%d) than %d did (%d)", n, got, n-1, prev)
		}
		prev = got
	}
	if mapmodel.SkyMaxCopies < 24 {
		t.Errorf("the cap is %d, want at least 24", mapmodel.SkyMaxCopies)
	}
}

// TestSilhouettesPerKey: every type draws SkyCopies of its real count, and a
// town ten times bigger draws more, to the cap.
func TestSilhouettesPerKey(t *testing.T) {
	for _, scale := range []float64{1, 4, 20} {
		m := build(fixture.Options{Age: "industrial_age", Seed: 5, Scale: scale, Tick: tickAt(0.5)}, 0)
		per := map[string]int{}
		for _, l := range m.Skyline.Lots {
			per[l.Key]++
		}
		capped := false
		for key, n := range per {
			b := m.Building(key)
			if b == nil {
				t.Fatalf("scale %v: lots of %s, which is not built", scale, key)
			}
			if b.Wonder {
				if n != 1 {
					t.Errorf("scale %v: the wonder %s draws %d silhouettes, want 1", scale, key, n)
				}
				continue
			}
			if want := mapmodel.SkyCopies(b.Count); n != want {
				t.Errorf("scale %v: %s ×%d draws %d silhouettes, want %d", scale, key, b.Count, n, want)
			}
			capped = capped || n == mapmodel.SkyMaxCopies
		}
		if scale >= 20 && !capped {
			t.Errorf("scale %v: no type reaches the cap of %d", scale, mapmodel.SkyMaxCopies)
		}
	}
}

// TestDistrictsWidenToFit: a bigger town gets wider districts, every
// silhouette stands inside its own district, and no two stand in one place.
func TestDistrictsWidenToFit(t *testing.T) {
	for _, age := range []string{"stone_age", "medieval_age", "industrial_age", "cyberpunk_age"} {
		small := build(fixture.Options{Age: age, Seed: 5, Scale: 1, Tick: tickAt(0.5)}, 0)
		big := build(fixture.Options{Age: age, Seed: 5, Scale: 6, Tick: tickAt(0.5)}, 0)
		if big.Skyline.Width <= small.Skyline.Width {
			t.Errorf("%s: a town six times as big is %d columns wide, the small one %d", age, big.Skyline.Width, small.Skyline.Width)
		}
		for _, m := range []*mapmodel.Model{small, big} {
			for _, l := range m.Skyline.Lots {
				d := m.Skyline.Districts[l.Age]
				lo, hi := d.X0, d.X0+d.LandW
				if d.Bay && (l.Quay || l.Key == "grand_lighthouse") {
					lo, hi = d.BayX, d.BayX+d.BayW
				}
				if l.X < lo || l.X >= hi {
					t.Errorf("%s: %s#%d stands at %d, outside its district's %d..%d", age, l.Key, l.Copy, l.X, lo, hi)
				}
			}
			for i, d := range m.Skyline.Districts {
				if i > 0 && d.X0 != m.Skyline.Districts[i-1].X0+m.Skyline.Districts[i-1].W {
					t.Errorf("%s: district %d does not start where the last ends", age, i)
				}
				if d.LandW < 30 {
					t.Errorf("%s: district %d is %d columns of land", age, i, d.LandW)
				}
			}
		}
	}
}

// edgeCellsDarker finds, for the lots of a composed scene in the front row,
// edge cells that were drawn solid on a plain wall, and reports how many are
// darker than the face colour the palette gives that slot, and how many are not.
func edgeCellsDarker(t *testing.T, s *scene, v *view) (darker, other int) {
	t.Helper()
	for _, li := range s.vis {
		lv := &s.lay.lots[li]
		if lv.depth != dRow0 || lv.ml.Wonder {
			continue
		}
		spr := lv.spr
		edge := spr.edges()
		for sy := 0; sy < spr.h; sy++ {
			for sx := 0; sx < spr.w; sx++ {
				c := spr.c[sy*spr.w+sx]
				if !edge[sy*spr.w+sx] || c.ch != '█' || !isWall(c.fg) || c.bg != sNone {
					continue
				}
				X, Y := lv.x0+sx-s.cam, s.Y(s.groundY-spr.h+sy)
				cl := v.fb.at(X, Y)
				if cl == nil || cl.d != lv.depth || cl.ch != '█' {
					continue
				}
				face := s.p.col(lv.fam, lv.vr, c.fg, lv.ml.Row)
				if luma(cl.fg) < luma(face)-1 {
					darker++
				} else {
					other++ // (or a neighbour in the same row drew its own face over it)
				}
			}
		}
	}
	return
}

func TestOutlines(t *testing.T) {
	for _, age := range []string{"stone_age", "medieval_age", "industrial_age", "modern_age"} {
		m := build(fixture.Options{Age: age, Seed: 5, Tick: tickAt(0.5)}, 0)
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 3}, 160, 45)
		edges, cells := 0, 0
		for _, li := range s.vis {
			spr := s.lay.lots[li].spr
			e := spr.edges()
			for i, c := range spr.c {
				cells++
				if e[i] {
					edges++
					x, y := i%spr.w, i/spr.w
					if c.ch == 0 {
						t.Fatalf("%s: an empty cell is an edge", age)
					}
					if spr.at(x-1, y).ch != 0 && spr.at(x+1, y).ch != 0 && spr.at(x, y-1).ch != 0 {
						t.Fatalf("%s: an interior cell is an edge", age)
					}
				}
			}
		}
		if edges == 0 || edges*2 > cells {
			t.Errorf("%s: %d of %d cells are outline", age, edges, cells)
		}
		darker, other := edgeCellsDarker(t, s, v)
		if darker == 0 {
			t.Errorf("%s: no front-row edge is drawn darker than its face (%d are not)", age, other)
		}
		if other > darker {
			t.Errorf("%s: most edge cells are not darker than their face (%d are not, %d are)", age, other, darker)
		}
	}
}

// blueShare is how much of a colour is blue.
func blueShare(c tcell.Color) float64 {
	r, g, b := c.RGB()
	return float64(b) / float64(max(1, int(r+g+b)))
}

// TestBackRowsRecede: in every theme, by day and by night, the rows step back
// in a built material: in a dark theme each row behind is dimmer and cooler
// than the one in front; in a light theme it fades toward the paper.
func TestBackRowsRecede(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, tod := range []float64{0.5, 0.95} {
			m := build(fixture.Options{Age: "medieval_age", Seed: 5, Tick: tickAt(tod)}, 0)
			p := newPal(m)
			if p.mono {
				continue // a duotone theme has its own ramp
			}
			for fam := 0; fam < theme.SkyFamilies; fam++ {
				for _, sl := range []slot{sWall, sWallLit, sWallShade, sRoof, sTrim} {
					row := [3]tcell.Color{p.col(fam, 0, sl, 0), p.col(fam, 0, sl, 1), p.col(fam, 0, sl, 2)}
					at := func() string {
						return th.Key + " tod " + strconv.FormatFloat(tod, 'f', 2, 64) + " family " + strconv.Itoa(fam) + " slot " + strconv.Itoa(int(sl))
					}
					for r := 1; r < 3; r++ {
						if p.light {
							if gap := func(c tcell.Color) float64 { return absf(luma(c) - luma(p.horizon)) }; gap(row[r]) > gap(row[r-1])+0.5 {
								t.Errorf("%s: row %d is farther from the paper than row %d", at(), r, r-1)
							}
							continue
						}
						if luma(row[r]) > luma(row[r-1])+0.5 {
							t.Errorf("%s: row %d (%.0f) is brighter than row %d (%.0f)", at(), r, luma(row[r]), r-1, luma(row[r-1]))
						}
						// (the cool filter itself, before the haze mixes in the horizon)
						if raw := p.raw(p.mats[fam][0], sl); blueShare(p.tone(raw, r)) < blueShare(raw)-0.001 {
							t.Errorf("%s: row %d's tone is warmer than the material", at(), r)
						}
					}
					if !p.light && luma(row[0]) > 20 && luma(row[2]) >= luma(row[0])-1 {
						t.Errorf("%s: the back row (%.0f) is no dimmer than the front (%.0f)", at(), luma(row[2]), luma(row[0]))
					}
				}
			}
		}
	}
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// spriteSig is a sprite's picture as a number.
func spriteSig(s *sprite) uint64 {
	h := uint64(s.w)*131 + uint64(s.h)
	for _, c := range s.c {
		h = h*1099511628211 ^ uint64(c.ch)<<16 ^ uint64(c.fg)<<8 ^ uint64(c.bg)
	}
	return h
}

// TestVariants: the three variants all occur, a copy's variant never changes,
// and the copies of one type do not all look alike; the same town drawn twice
// is the same town.
func TestVariants(t *testing.T) {
	seen := map[int]int{}
	for seed := uint64(1); seed < 300; seed++ {
		v := variantOf(seed * 2654435761)
		if v != variantOf(seed*2654435761) {
			t.Fatal("a variant is not stable")
		}
		seen[v]++
	}
	for v := 0; v < numVariants; v++ {
		if seen[v] < 40 {
			t.Errorf("variant %d came up %d times in 299 copies", v, seen[v])
		}
	}
	m := build(fixture.Options{Age: "medieval_age", Seed: 5, Scale: 5, Tick: tickAt(0.5)}, 0)
	v1, v2 := newView(), newView()
	l1, l2 := v1.layoutFor(m, 36), v2.layoutFor(m, 36)
	if len(l1.lots) != len(l2.lots) {
		t.Fatal("two layouts of one town differ in size")
	}
	byKey := map[string]map[uint64]bool{}
	variants := map[string]map[int]bool{}
	for i := range l1.lots {
		a, b := &l1.lots[i], &l2.lots[i]
		if a.ml.Key != b.ml.Key || a.ml.Copy != b.ml.Copy || spriteSig(a.spr) != spriteSig(b.spr) || a.variant != b.variant {
			t.Fatalf("%s#%d is not the same in two layouts", a.ml.Key, a.ml.Copy)
		}
		if a.ml.Wonder {
			continue
		}
		if byKey[a.ml.Key] == nil {
			byKey[a.ml.Key], variants[a.ml.Key] = map[uint64]bool{}, map[int]bool{}
		}
		byKey[a.ml.Key][spriteSig(a.spr)] = true
		variants[a.ml.Key][a.variant] = true
	}
	checked, alike := 0, 0
	for key, sigs := range byKey {
		n := 0
		for _, l := range m.Skyline.Lots {
			if l.Key == key {
				n++
			}
		}
		if n < 6 {
			continue
		}
		checked++
		if len(variants[key]) < 2 {
			t.Errorf("%s: %d copies and one variant", key, n)
		}
		if len(sigs) < 3 {
			alike++
		}
	}
	if checked == 0 {
		t.Fatal("no type has six copies to compare")
	}
	if alike > checked/5 {
		t.Errorf("%d of %d types with six or more copies draw fewer than three different silhouettes", alike, checked)
	}
	// the variants themselves differ: one form dressed three ways
	form := aVilla
	sigs := map[uint64]bool{}
	for v := 0; v < numVariants; v++ {
		sigs[spriteSig(dress(form(newRnd(77), 9), v, 77, 4, 40))] = true
	}
	if len(sigs) != numVariants {
		t.Errorf("one building dressed as the %d variants makes %d pictures", numVariants, len(sigs))
	}
}

// litCells counts window cells of a frame that are lit (brighter than the
// same cell by day would be) and returns the colours of those cells.
func windowColours(t *testing.T, m *mapmodel.Model, anim int) map[[2]int]tcell.Color {
	t.Helper()
	v := newView()
	s := v.compose(mapstyle.Frame{Model: m, Anim: anim}, 160, 45)
	out := map[[2]int]tcell.Color{}
	for _, li := range s.vis {
		lv := &s.lay.lots[li]
		spr := lv.spr
		for sy := 0; sy < spr.h; sy++ {
			for sx := 0; sx < spr.w; sx++ {
				if c := spr.c[sy*spr.w+sx]; c.fg == sWin && c.ch != 0 {
					X, Y := lv.x0+sx-s.cam, s.Y(s.groundY-spr.h+sy)
					if cl := v.fb.at(X, Y); cl != nil && cl.d == lv.depth && cl.ch == c.ch {
						out[[2]int{X, Y}] = cl.fg
					}
				}
			}
		}
	}
	return out
}

// TestLitWindows: windows are lit at night in every age's own style, they
// waver slowly from frame to frame, and with the animation held (motion off)
// the same frame is the same picture.
func TestLitWindows(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	_ = theme.SetActive("forge")
	for _, age := range []string{"bronze_age", "medieval_age", "modern_age", "cyberpunk_age"} {
		night := build(fixture.Options{Age: age, Seed: 5, Scale: 2, Tick: tickAt(0.95)}, 0)
		day := build(fixture.Options{Age: age, Seed: 5, Scale: 2, Tick: tickAt(0.5)}, 0)
		a := windowColours(t, night, 40)
		d := windowColours(t, day, 40)
		if len(a) < 10 {
			t.Errorf("%s: only %d window cells on screen at night", age, len(a))
			continue
		}
		bright := 0
		for k, c := range a {
			if dc, ok := d[k]; ok && luma(c) > luma(dc)+25 {
				bright++
			}
		}
		if bright < 5 {
			t.Errorf("%s: %d windows are lit at night", age, bright)
		}
		// held still: the same frame twice
		again := windowColours(t, night, 40)
		for k, c := range a {
			if again[k] != c {
				t.Fatalf("%s: window %v is not the same in two frames of one animation step", age, k)
			}
		}
		// moving: the light wavers
		changed := 0
		for _, an := range []int{46, 52, 58} {
			for k, c := range windowColours(t, night, an) {
				if a[k] != c {
					changed++
				}
			}
		}
		if st := styleOfAge(night.AgeIdx); st != wElectric && st != wLED && changed == 0 {
			t.Errorf("%s: no window flickers over three steps", age)
		}
	}
}

// TestWindowGlowBounds: a window's light is always a fraction of full, slow
// (it never jumps from one frame to the next except in a brownout or a sign's
// stutter), and held at one frame it holds.
func TestWindowGlowBounds(t *testing.T) {
	for _, age := range []int{0, 5, 12, 16} {
		for k := 0; k < 40; k++ {
			prev := windowGlow(age, k*7, k, k/3, 0)
			for an := 0; an < 400; an++ {
				g := windowGlow(age, k*7, k, k/3, an)
				if g < 0.3 || g > 1 {
					t.Fatalf("age %d: glow %.2f out of range at frame %d", age, g, an)
				}
				if g != windowGlow(age, k*7, k, k/3, an) {
					t.Fatalf("age %d: glow is not a function of the frame", age)
				}
				if styleOfAge(age) == wTorch || styleOfAge(age) == wLamp {
					if d := absf(g - prev); d > 0.12 {
						t.Fatalf("age %d: glow jumped %.2f in one frame", age, d)
					}
				}
				prev = g
			}
		}
	}
}

// TestNothingDrawsOutsideTheScene: a building lands only in the scene's own
// rows, whatever the size of the terminal, the camera or the town.
func TestNothingDrawsOutsideTheScene(t *testing.T) {
	for _, age := range []string{"stone_age", "medieval_age", "cyberpunk_age", "space_age"} {
		m := build(fixture.Options{Age: age, Seed: 9, Scale: 3, Tick: tickAt(0.9)}, 0)
		for _, sz := range [][2]int{{160, 45}, {80, 24}, {200, 60}, {40, 15}} {
			for _, cam := range []int{0, 37, m.Skyline.Width / 2, m.Skyline.Width} {
				v := newView()
				v.follow, v.cam = false, cam
				var s *scene
				if sz[1] < 20 {
					scr := capture.NewScreen(sz[0], sz[1])
					v.DrawCompact(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, mapstyle.Frame{Model: m, Anim: 5})
					scr.Fini()
					continue
				}
				if m.Sky() != mapmodel.SkyGround {
					scr := capture.NewScreen(sz[0], sz[1])
					v.Draw(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, mapstyle.Frame{Model: m, Anim: 5})
					scr.Fini()
					continue
				}
				s = v.compose(mapstyle.Frame{Model: m, Anim: 5}, sz[0], sz[1])
				for i, c := range v.fb.c {
					if c.d < dWonder || c.d > dRow2 {
						continue
					}
					x, y := i%v.fb.w, i/v.fb.w
					if y < s.top || y >= s.Y(s.groundY) || x >= s.W {
						t.Fatalf("%s %dx%d cam %d: a building cell (depth %d) at %d,%d, outside the scene's rows %d..%d",
							age, sz[0], sz[1], cam, c.d, x, y, s.top, s.Y(s.groundY))
					}
				}
			}
		}
	}
}

// TestBuildingFrameBudget: a town as big as the cap allows still draws inside
// the frame guard, in six ages at 160x45, at night when every window is
// working.
func TestBuildingFrameBudget(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "modern_age", "cyberpunk_age"} {
		m := build(fixture.Options{Age: age, Seed: 7, Routes: 4, Scale: 6, Tick: tickAt(0.95)}, 1)
		v := newView()
		scr := capture.NewScreen(160, 45)
		f := mapstyle.Frame{Model: m}
		v.Draw(scr, mapstyle.Rect{W: 160, H: 45}, f)
		const n = 40
		start := time.Now()
		for i := 0; i < n; i++ {
			f.Anim = i
			v.Draw(scr, mapstyle.Rect{W: 160, H: 45}, f)
		}
		per := time.Since(start) / n
		t.Logf("160x45 %s (%d silhouettes): frame %v", age, len(m.Skyline.Lots), per)
		if per > 25*time.Millisecond {
			t.Errorf("%s: a frame takes %v", age, per)
		}
		scr.Fini()
	}
}

// TestDriftStaysOnScreen: a far object's parallax offset follows the camera
// while the panorama is short and stops short of its limit however long it is.
func TestDriftStaysOnScreen(t *testing.T) {
	prev := -1.0
	for cam := 0; cam <= 20000; cam += 50 {
		d := drift(cam, 0.03, 60)
		if d < prev {
			t.Fatalf("the drift goes back at camera %d", cam)
		}
		if d >= 60 {
			t.Fatalf("the drift reaches its limit (%.1f) at camera %d", d, cam)
		}
		if float64(cam)*0.03 <= 30 && d != float64(cam)*0.03 {
			t.Fatalf("under half the limit the drift is %.2f, want %.2f", d, float64(cam)*0.03)
		}
		prev = d
	}
}
