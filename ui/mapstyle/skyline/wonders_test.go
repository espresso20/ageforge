package skyline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// The ground ages are the first seventeen (Primitive to Fusion); from the
// Space Age the maps leave the ground (mapmodel.SkyScene), and their wonders
// stand in the sky arc's scenes.
const groundAges = 17

// wonderKeyOf is the key of the wonder of age index a.
func wonderKeyOf(a int) string { return wonderKeys(catalog())[a] }

// wonderModel is the model of age a with its wonder built, or, when share is
// above zero (and built is false), under way with that share of its cost
// banked.
func wonderModel(o fixture.Options, built bool, share float64) *mapmodel.Model {
	c := catalog()
	a := c.AgeIdx[o.Age]
	key := wonderKeyOf(a)
	st := fixture.State(o)
	bs := st.Buildings[key]
	switch {
	case built:
		bs.Count = 1
		bs.WonderBank, bs.WonderBankFull = nil, false
		st.CurrentAgeWonderKey, st.CurrentAgeWonderName = "", ""
	default:
		bs.Count = 0
		bs.NextCost = map[string]float64{"gold": 1000, "stone": 500}
		bs.WonderBank = map[string]float64{"gold": 1000 * share, "stone": 500 * share}
		st.CurrentAgeWonderKey, st.CurrentAgeWonderName = key, bs.Name
	}
	st.Buildings[key] = bs
	return mapmodel.NewBuilder(c).Build(&st, nil)
}

func ageKeys(from, to int) []string {
	var out []string
	for a := from; a < to; a++ {
		out = append(out, catalog().Ages[a])
	}
	return out
}

// wonderLot is the lot of the wonder of the model's present age.
func wonderLot(t *testing.T, m *mapmodel.Model) *mapmodel.Lot {
	t.Helper()
	key := wonderKeyOf(m.AgeIdx)
	for i := range m.Skyline.Lots {
		if l := &m.Skyline.Lots[i]; l.Wonder && l.Key == key {
			return l
		}
	}
	t.Fatalf("%s: no lot for the wonder %s", m.Age, key)
	return nil
}

// TestEveryWonderDraws: every age's wonder is drawn, at every height, as a
// sprite exactly as tall as the scene gives it, with something in every row
// band (no empty hull) and a foot on the ground; the ground ages' own scenes
// show it on screen, and the sky ages' scenes carry it too.
func TestEveryWonderDraws(t *testing.T) {
	c := catalog()
	for a := range c.ByAge {
		key := wonderKeyOf(a)
		if key == "" {
			t.Fatalf("age %d has no wonder", a)
		}
		// the heights a scene can ask for: the nominal height at every scale
		// from the smallest terminal to the largest
		var sizes []int
		for _, sc := range []float64{0.42, 0.5, 0.65, 0.8, 1, 1.25} {
			sizes = append(sizes, wonderRows(key, sc, 200))
		}
		for _, rows := range sizes {
			s := wonderArt(key, rows)
			if s.h != rows {
				t.Errorf("%s at %d rows: the sprite is %d rows tall", key, rows, s.h)
			}
			solid, foot := 0, 0
			for i, cell := range s.c {
				if cell.ch != 0 {
					solid++
					if i/s.w == s.h-1 {
						foot++
					}
				}
			}
			if solid < rows*2 {
				t.Errorf("%s at %d rows: only %d cells drawn", key, rows, solid)
			}
			if foot == 0 && key != "dyson_scaffold" && key != "singularity_core" {
				t.Errorf("%s at %d rows: nothing stands on the ground row", key, rows)
			}
			if s.w > 72 {
				t.Errorf("%s at %d rows is %d columns wide", key, rows, s.w)
			}
		}
	}
	for _, age := range ageKeys(0, groundAges) {
		m := wonderModel(fixture.Options{Age: age, Seed: 3, Tick: tickAt(0.5)}, true, 0)
		lot := wonderLot(t, m)
		v := newView()
		v.follow, v.cam = false, lot.X-40
		s := v.compose(mapstyle.Frame{Model: m, Anim: 3}, 80, 40)
		n := 0
		for i, cell := range v.fb.c {
			if cell.d == dWonder && cell.ch != ' ' && i/v.fb.w >= s.top {
				n++
			}
		}
		if n < 30 {
			t.Errorf("%s: the wonder shows %d cells on screen", age, n)
		}
	}
	for _, age := range ageKeys(groundAges, len(c.ByAge)-1) { // the mandala has no lots
		m := wonderModel(fixture.Options{Age: age, Seed: 3, Tick: tickAt(0.5)}, true, 0)
		lot := wonderLot(t, m)
		v := newView()
		v.follow, v.cam = false, lot.X-40
		v.compose(mapstyle.Frame{Model: m, Anim: 3}, 80, 40)
		lay := v.slay
		if lay == nil {
			t.Fatalf("%s: no sky layout", age)
		}
		i := lay.find(lot.Key, 0)
		if i < 0 || lay.lots[i].spr.h < 3 {
			t.Errorf("%s: the sky scene draws no landmark for %s", age, lot.Key)
		}
	}
}

// TestWonderIsTheTallest: in its own district the wonder stands at least half
// as tall again as the tallest building (every ground age, every terminal
// height, a thin town and a crowded one), and within the room the scene has.
func TestWonderIsTheTallest(t *testing.T) {
	for _, scale := range []float64{1, 4} {
		for _, age := range ageKeys(0, groundAges) {
			for _, h := range []int{16, 24, 30, 36, 45, 60} {
				m := wonderModel(fixture.Options{Age: age, Seed: 5, Scale: scale, Tick: tickAt(0.5)}, true, 0)
				v := newView()
				ref := h - 6 // the scene's room above the ground at mid zoom
				lay := v.layoutSized(m, ref, ref)
				a := m.AgeIdx
				wh, tall, tk := 0, 0, ""
				for i := range lay.lots {
					lv := &lay.lots[i]
					if lv.ml.Age != a {
						continue
					}
					if lv.ml.Wonder {
						wh = lv.spr.h
					} else if lv.spr.h > tall {
						tall, tk = lv.spr.h, lv.ml.Key
					}
				}
				if wh == 0 {
					t.Fatalf("%s %d rows: no wonder in the layout", age, h)
				}
				if wh > ref-2 {
					t.Errorf("%s at %d rows: the wonder is %d rows, the scene holds %d", age, h, wh, ref-2)
				}
				if 2*wh < 3*tall {
					t.Errorf("%s at %d rows (x%v): the wonder is %d rows against %s at %d: under one and a half times as tall", age, h, scale, wh, tk, tall)
				}
			}
		}
	}
}

// TestWonderFitsTheScene: whatever the terminal, a landmark and its
// scaffolding draw inside the scene's rows, never over the header or under
// the ground.
func TestWonderFitsTheScene(t *testing.T) {
	for _, sz := range [][2]int{{60, 16}, {80, 24}, {100, 30}, {120, 36}, {160, 45}, {200, 60}} {
		for _, zm := range zoomSteps {
			for _, age := range []string{"stone_age", "classical_age", "victorian_age", "atomic_age", "information_age", "fusion_age"} {
				for _, built := range []bool{true, false} {
					m := wonderModel(fixture.Options{Age: age, Seed: 4, Tick: tickAt(0.5)}, built, 0.6)
					x := wonderX(m.Skyline.Districts[m.AgeIdx], wonderKeyOf(m.AgeIdx)) // the plot
					v := newView()
					v.zoom, v.follow = zm, false
					v.cam = x - zm.cols(sz[0])/2
					s := v.compose(mapstyle.Frame{Model: m, Anim: 5}, sz[0], sz[1])
					// the world frame is v.lfb away from mid zoom
					f := &v.fb
					if zm != zMid {
						f = &v.lfb
					}
					seen := 0
					for i, c := range f.c {
						if c.d != dWonder {
							continue
						}
						seen++
						y := i / f.w
						if y < 1 || y > 1+s.groundY-1 {
							t.Fatalf("%dx%d %s %s built=%v: a landmark cell at world row %d, outside the scene's 1..%d", sz[0], sz[1], zm, age, built, y, s.groundY)
						}
					}
					if seen == 0 {
						t.Errorf("%dx%d %s %s built=%v: no landmark cells in the frame", sz[0], sz[1], zm, age, built)
					}
				}
			}
		}
	}
}

// TestWonderPlotMatchesTheModel: a wonder under way stands where the model
// places it once built (wonderX is mapmodel's skylineFor rule, repeated).
func TestWonderPlotMatchesTheModel(t *testing.T) {
	c := catalog()
	for a := range c.ByAge {
		m := wonderModel(fixture.Options{Age: c.Ages[a], Seed: 8, Tick: tickAt(0.5)}, true, 0)
		lot := wonderLot(t, m)
		if x := wonderX(m.Skyline.Districts[a], lot.Key); x != lot.X {
			t.Errorf("%s: wonderX is %d, the model's lot stands at %d", lot.Key, x, lot.X)
		}
	}
}

// TestScaffoldFollowsTheBank: a wonder under way shows as a frame whose
// height rises with its banked share, from a footing to the whole landmark;
// nothing shows until something is banked.
func TestScaffoldFollowsTheBank(t *testing.T) {
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "victorian_age", "modern_age", "fusion_age"} {
		prev, prevCells := -1, -1
		for _, share := range []float64{0.02, 0.1, 0.25, 0.5, 0.75, 0.95, 1} {
			m := wonderModel(fixture.Options{Age: age, Seed: 2, Tick: tickAt(0.5)}, false, share)
			if w := wonderProgress(m); w < share-0.001 || w > share+0.001 {
				t.Fatalf("%s: the model says %.3f banked, want %.3f", age, w, share)
			}
			v := newView()
			lay := v.layoutFor(m, 39)
			if len(lay.sites) != 1 {
				t.Fatalf("%s at %.2f: %d scaffolds in the layout, want 1", age, share, len(lay.sites))
			}
			rows := wonderRows(wonderKeyOf(m.AgeIdx), lay.scale, lay.ref)
			site := lay.sites[0].spr
			want := min(rows, siteRows(share, rows)+2)
			if site.h != want {
				t.Errorf("%s at %.2f: the scaffold is %d rows, want %d (the wonder is %d)", age, share, site.h, want, rows)
			}
			if site.h < prev {
				t.Errorf("%s: the scaffold shrinks, %d rows at %.2f after %d", age, site.h, share, prev)
			}
			cells := 0
			for _, c := range site.c {
				if c.ch != 0 {
					cells++
				}
			}
			if cells < prevCells {
				t.Errorf("%s at %.2f: %d cells drawn, fewer than the %d before", age, share, cells, prevCells)
			}
			prev, prevCells = site.h, cells
			if share == 1 && site.h != rows {
				t.Errorf("%s: with everything banked the frame is %d rows, want the wonder's %d", age, site.h, rows)
			}
		}
	}
	// nothing banked, nothing built, nothing queued: no plot at all
	m := wonderModel(fixture.Options{Age: "medieval_age", Seed: 2, Tick: tickAt(0.5)}, false, 0)
	if lay := newView().layoutFor(m, 39); len(lay.sites) != 0 {
		t.Errorf("a wonder with nothing banked has %d scaffolds", len(lay.sites))
	}
	// the scaffold is no lot: no inspection stops on it
	m = wonderModel(fixture.Options{Age: "medieval_age", Seed: 2, Tick: tickAt(0.5)}, false, 0.5)
	for _, tg := range targets(m, 120, 0) {
		if tg.kind == tLot && tg.key == wonderKeyOf(m.AgeIdx) {
			t.Errorf("the cursor can stand on the unbuilt wonder")
		}
	}
}

func wonderProgress(m *mapmodel.Model) float64 {
	for _, w := range m.Wonders {
		if w.Key == wonderKeyOf(m.AgeIdx) {
			return w.Progress
		}
	}
	return -1
}

// TestScaffoldDrawsOnScreen: the scaffold is on the frame, at the wonder's
// plot, and gets taller with the bank.
func TestScaffoldDrawsOnScreen(t *testing.T) {
	top := func(share float64) int {
		m := wonderModel(fixture.Options{Age: "industrial_age", Seed: 2, Tick: tickAt(0.5)}, false, share)
		x := wonderX(m.Skyline.Districts[m.AgeIdx], wonderKeyOf(m.AgeIdx))
		v := newView()
		v.follow, v.cam = false, x-40
		s := v.compose(mapstyle.Frame{Model: m, Anim: 2}, 80, 45)
		row := -1
		for i, c := range v.fb.c {
			if c.d == dWonder && c.ch != ' ' {
				if y := i / v.fb.w; row < 0 || y < row {
					row = y
				}
			}
		}
		if row < 0 {
			t.Fatalf("no scaffold on screen at %.2f", share)
		}
		return s.groundY + s.top - row
	}
	lo, mid, hi := top(0.1), top(0.5), top(0.95)
	if !(lo < mid && mid < hi) {
		t.Errorf("the scaffold stands %d, %d and %d rows high at 10%%, 50%% and 95%%: want it to rise", lo, mid, hi)
	}
}

// TestWonderGlyphs: every cell of every landmark, its close detail included,
// is a rune that is one cell wide and folds to ASCII without a '?'.
func TestWonderGlyphs(t *testing.T) {
	c := catalog()
	for a := range c.ByAge {
		key := wonderKeyOf(a)
		for _, rows := range []int{wonderMinRows, 12, 24, 36} {
			s := wonderArt(key, rows)
			for _, spr := range []*sprite{s, s.closeUp()} {
				for _, cell := range spr.c {
					for _, r := range []rune{cell.ch} {
						if r == 0 {
							continue
						}
						if uniseg.StringWidth(string(r)) != 1 {
							t.Fatalf("%s: %q is not one cell wide", key, r)
						}
						if f := mapmodel.Fold(r, mapmodel.TierASCII); f == '?' || f >= 0x80 {
							t.Fatalf("%s: %q folds to %q in the ASCII tier", key, r, f)
						}
						if r == '—' {
							t.Fatalf("%s: an em dash", key)
						}
					}
				}
			}
		}
	}
}

// TestWonderTiersAndThemes draws every ground age's landmark, built and
// under way, at day and night, at mid and close zoom, in every glyph tier,
// across the themes, with the inspect cursor beside it (so the plaque draws):
// no panic, no '?', nothing outside ASCII in the ASCII tier, no glyph in its
// own colour.
func TestWonderTiersAndThemes(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	themes := theme.All()
	n := 0
	for i, age := range ageKeys(0, groundAges) {
		for _, built := range []bool{true, false} {
			for _, night := range []bool{false, true} {
				for _, zm := range []zoom{zMid, zClose} {
					tod := 0.5
					if night {
						tod = 0.95
					}
					m := wonderModel(fixture.Options{Age: age, Seed: 6, Tick: tickAt(tod)}, built, 0.55)
					x := wonderX(m.Skyline.Districts[m.AgeIdx], wonderKeyOf(m.AgeIdx))
					for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
						n++
						if err := theme.SetActive(themes[(i+n)%len(themes)].Key); err != nil {
							t.Fatal(err)
						}
						v := newView()
						v.zoom, v.follow, v.inspect = zm, false, true
						v.cam = x - zm.cols(110)/2
						// the cursor on the lot nearest the landmark
						best, bd := -1, 1<<30
						for j := range m.Skyline.Lots {
							if l := &m.Skyline.Lots[j]; !l.Wonder && abs(l.X-x) < bd {
								best, bd = j, abs(l.X-x)
							}
						}
						if best >= 0 {
							l := &m.Skyline.Lots[best]
							v.cur = target{tLot, l.Key, l.Copy}
						}
						scr := draw(v, m, 110, 36, 9, tier, false)
						cells, _, _ := scr.GetContents()
						for _, c := range cells {
							if len(c.Runes) == 0 {
								continue
							}
							r := c.Runes[0]
							fg, bg, _ := c.Style.Decompose()
							switch {
							case tier == mapmodel.TierASCII && r >= 0x80:
								t.Fatalf("%s built=%v night=%v %s ascii: non-ASCII rune %q", age, built, night, zm, r)
							case tier != mapmodel.TierASCII && r == '?':
								t.Fatalf("%s built=%v night=%v %s %s: a '?' leaked into the frame", age, built, night, zm, tier)
							case tier != mapmodel.TierASCII && uniseg.StringWidth(string(r)) != 1:
								t.Fatalf("%s %s: %q is not one cell wide", age, tier, r)
							case r != ' ' && fg == bg:
								t.Fatalf("%s built=%v night=%v %s %s: %q has fg == bg", age, built, night, zm, tier, r)
							}
						}
					}
				}
			}
		}
	}
}

// TestWonderCloseDetail: every ground landmark carries detail only the close
// zoom draws, drawn on its solid faces, and the mid zoom never uses it.
func TestWonderCloseDetail(t *testing.T) {
	for a := 0; a < groundAges; a++ {
		key := wonderKeyOf(a)
		s := wonderArt(key, 28)
		if len(s.fine) < 3 {
			t.Errorf("%s: %d cells of close detail, want some", key, len(s.fine))
			continue
		}
		cu := s.closeUp()
		if cu == s {
			t.Errorf("%s: closeUp returns the sprite itself", key)
			continue
		}
		diff := 0
		for i := range s.c {
			if s.c[i] != cu.c[i] {
				diff++
			}
		}
		if diff < 3 {
			t.Errorf("%s: the close zoom changes %d cells, want its detail to show", key, diff)
		}
		if cu.h != s.h || cu.w != s.w {
			t.Errorf("%s: closeUp is %dx%d, the sprite %dx%d", key, cu.w, cu.h, s.w, s.h)
		}
		if again := s.closeUp(); again != cu {
			t.Errorf("%s: closeUp is not cached", key)
		}
	}
	// a sprite with no detail is its own close-up
	if s := newSprite(3, 3); s.closeUp() != s {
		t.Errorf("a plain sprite should be its own close-up")
	}
}

// TestWonderLightsAtNight: the lamps of the landmarks that have them throw
// beams at night (frames with the beams differ from frames without), and
// the glow stands round the silhouette.
func TestWonderLightsAtNight(t *testing.T) {
	for _, age := range []string{"colonial_age", "victorian_age", "cyberpunk_age"} {
		m := wonderModel(fixture.Options{Age: age, Seed: 3, Tick: tickAt(0.95)}, true, 0)
		lot := wonderLot(t, m)
		v := newView()
		v.follow, v.cam = false, lot.X-45
		frame := func(anim int) []cell {
			v.compose(mapstyle.Frame{Model: m, Anim: anim}, 90, 45)
			return append([]cell(nil), v.fb.c...)
		}
		lay := v.layoutFor(m, 39)
		spr := lay.lots[lay.find(lot.Key, 0)].spr
		if len(spr.beams) == 0 {
			t.Fatalf("%s has no lamp to throw a beam", lot.Key)
		}
		if len(spr.haloCells()) < 10 {
			t.Errorf("%s: the glow has %d cells", lot.Key, len(spr.haloCells()))
		}
		with := map[int][]cell{}
		for anim := 0; anim < 64; anim += 3 {
			with[anim] = frame(anim)
		}
		saved := spr.beams
		spr.beams = nil
		most := 0
		for anim := 0; anim < 64; anim += 3 {
			without := frame(anim)
			diff := 0
			for i := range without {
				if without[i] != with[anim][i] {
					diff++
				}
			}
			most = max(most, diff)
		}
		spr.beams = saved
		if most < 12 {
			t.Errorf("%s: at night the beams change at most %d cells of a frame", lot.Key, most)
		}
		// by day there is no glow and no beam
		md := wonderModel(fixture.Options{Age: age, Seed: 3, Tick: tickAt(0.5)}, true, 0)
		vd := newView()
		vd.follow, vd.cam = false, lot.X-45
		vd.compose(mapstyle.Frame{Model: md, Anim: 5}, 90, 45)
		day := append([]cell(nil), vd.fb.c...)
		layd := vd.layoutFor(md, 39)
		sprd := layd.lots[layd.find(lot.Key, 0)].spr
		sprd.beams = nil
		vd.compose(mapstyle.Frame{Model: md, Anim: 5}, 90, 45)
		for i := range day {
			if day[i] != vd.fb.c[i] {
				t.Fatalf("%s: by day the beams change cell %d", lot.Key, i)
			}
		}
	}
}

// TestNightRimIsLit: a landmark's outline is lit at night where a building's
// is darker than its face, and stays dark by day.
func TestNightRimIsLit(t *testing.T) {
	m := wonderModel(fixture.Options{Age: "classical_age", Seed: 3, Tick: tickAt(0.5)}, true, 0)
	s := &scene{p: newPal(m)}
	face := theme.Color(theme.RoleText)
	glow := s.p.hue(theme.SkyLamp, mEmit, 0)
	dark := s.rimShade(face, false, glow, 0)
	lit := s.rimShade(face, true, glow, 1)
	if dark == lit {
		t.Fatalf("a lit rim looks like a dark one")
	}
	if theme.Shade(face, outlineK) != dark {
		t.Errorf("an unlit rim is not the building outline")
	}
	if luma(lit) <= luma(dark) {
		t.Errorf("the lit rim (%v) is no lighter than the dark one (%v)", lit, dark)
	}
}

// TestWonderPlaque: with the inspect cursor on a lot near a landmark its name
// stands under it; with the cursor on the landmark (which the cursor itself
// names), far away, or off, there is no plaque. A wonder going up says how
// much is banked.
func TestWonderPlaque(t *testing.T) {
	name := func(m *mapmodel.Model) string { return m.Catalog.Defs[wonderKeyOf(m.AgeIdx)].Name }
	rowText := func(v *view, m *mapmodel.Model, x int) string {
		scr := capture.NewScreen(100, 40)
		v.cam = x - 50
		v.Draw(scr, mapstyle.Rect{W: 100, H: 40}, mapstyle.Frame{Model: m, Anim: 3})
		lines := strings.Split(capture.Text(scr), "\n")
		return strings.Join(lines[:len(lines)-4], "\n") // the scene, not the key bar and the cursor's line under it
	}
	m := wonderModel(fixture.Options{Age: "classical_age", Seed: 3, Tick: tickAt(0.5)}, true, 0)
	lot := wonderLot(t, m)
	near, far := -1, -1
	for i := range m.Skyline.Lots {
		l := &m.Skyline.Lots[i]
		d := abs(l.X - lot.X)
		switch {
		case l.Wonder:
		case d > 3 && d < 14 && near < 0:
			near = i
		case d > 60 && far < 0:
			far = i
		}
	}
	if near < 0 || far < 0 {
		t.Fatalf("no lots near (%d) and far (%d) from the wonder", near, far)
	}
	v := newView()
	v.follow, v.inspect = false, true
	on := func(l *mapmodel.Lot) { v.cur = target{tLot, l.Key, l.Copy} }
	on(&m.Skyline.Lots[near])
	if txt := rowText(v, m, lot.X); !strings.Contains(txt, " "+name(m)+" ") {
		t.Errorf("the cursor beside %s shows no plaque:\n%s", name(m), txt)
	}
	on(&m.Skyline.Lots[far])
	if txt := rowText(v, m, lot.X); strings.Count(txt, name(m)) != 0 {
		t.Errorf("the cursor far from %s shows its plaque", name(m))
	}
	v.cur = target{tLot, lot.Key, lot.Copy}
	if txt := rowText(v, m, lot.X); strings.Count(txt, name(m)) != 1 {
		t.Errorf("the cursor on %s shows its name %d times, want once (the cursor's own label)", name(m), strings.Count(txt, name(m)))
	}
	on(&m.Skyline.Lots[near])
	v.inspect = false
	if txt := rowText(v, m, lot.X); strings.Contains(txt, name(m)) {
		t.Errorf("the plaque shows with the cursor off")
	}
	// under way: the plaque says how much is banked
	mu := wonderModel(fixture.Options{Age: "classical_age", Seed: 3, Tick: tickAt(0.5)}, false, 0.4)
	x := wonderX(mu.Skyline.Districts[mu.AgeIdx], wonderKeyOf(mu.AgeIdx))
	best, bd := -1, 1<<30
	for i := range mu.Skyline.Lots {
		if d := abs(mu.Skyline.Lots[i].X - x); d > 3 && d < bd {
			best, bd = i, d
		}
	}
	v = newView()
	v.follow, v.inspect = false, true
	l := &mu.Skyline.Lots[best]
	v.cur = target{tLot, l.Key, l.Copy}
	if txt := rowText(v, mu, x); !strings.Contains(txt, name(mu)+" 40%") {
		t.Errorf("the plaque of a wonder 40%% banked does not say so:\n%s", txt)
	}
}

// TestDistrictCap: the soft cap keeps order, leaves the low buildings alone
// and never reaches the cap from below.
func TestDistrictCap(t *testing.T) {
	for maxH := 6; maxH <= 60; maxH += 3 {
		for capH := 3; capH <= maxH; capH++ {
			prev := 0
			for h := 2; h <= maxH; h++ {
				got := softCap(h, capH, maxH)
				if got > capH {
					t.Fatalf("softCap(%d, %d, %d) = %d: over the cap", h, capH, maxH, got)
				}
				if got < prev {
					t.Fatalf("softCap(%d, %d, %d) = %d after %d: not monotone", h, capH, maxH, got, prev)
				}
				if h <= capH*6/10 && got != h {
					t.Fatalf("softCap(%d, %d, %d) = %d: a building under the knee changed", h, capH, maxH, got)
				}
				prev = got
			}
			if maxH > capH*6/10 && softCap(maxH, capH, maxH) != capH {
				t.Fatalf("softCap(%d, %d, %d): the tallest allowed does not reach the cap", maxH, capH, maxH)
			}
		}
	}
	if districtCap(30)*3 > 2*30+2 || districtCap(30) < 19 {
		t.Errorf("districtCap(30) = %d", districtCap(30))
	}
	if districtCap(1) < 2 {
		t.Errorf("districtCap(1) = %d", districtCap(1))
	}
}

// TestWonderFrameGuard: with every wonder built, the present one rising in
// its scaffold, a night frame at 160x45 stays inside the 25 ms guard at each
// zoom for six ages (the times are logged).
func TestWonderFrameGuard(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, age := range []string{"bronze_age", "classical_age", "industrial_age", "modern_age", "cyberpunk_age", "fusion_age"} {
		m := build(fixture.Options{Age: age, Seed: 7, Routes: 5, Tick: tickAt(0.95), Scale: 2}, 1)
		for _, zm := range zoomSteps {
			v := newView()
			v.zoom, v.inspect = zm, true
			scr := capture.NewScreen(160, 45)
			f := mapstyle.Frame{Model: m}
			// in the middle of the wonder's district, where the landmark and its glow are
			d := m.Skyline.Districts[len(m.Skyline.Districts)-1]
			v.follow = false
			v.cam = wonderX(d, wonderKeyOf(m.AgeIdx)) - zm.cols(160)/2
			v.Draw(scr, mapstyle.Rect{W: 160, H: 45}, f)
			const n = 30
			start := time.Now()
			for i := 0; i < n; i++ {
				f.Anim = i
				v.Draw(scr, mapstyle.Rect{W: 160, H: 45}, f)
			}
			per := time.Since(start) / n
			t.Logf("160x45 %-14s %-5s night frame %v", age, zm, per)
			if per > 25*time.Millisecond {
				t.Errorf("%s at %s: a frame takes %v", age, zm, per)
			}
			scr.Fini()
		}
	}
}

// TestSkyWonderArt: the cosmic landmarks draw at every height too, exactly
// that tall, so the sky arc's scenes can ask for any size.
func TestSkyWonderArt(t *testing.T) {
	for _, key := range []string{"dyson_scaffold", "warp_nexus", "cosmic_beacon", "reality_anchor", "singularity_core", "stellar_cradle"} {
		for rows := 6; rows <= 40; rows++ {
			s := wonderArt(key, rows)
			if s.h != rows || s.w < 3 {
				t.Fatalf("%s at %d rows: %dx%d", key, rows, s.w, s.h)
			}
		}
	}
}

// TestWonderCaptures writes review captures of every age's wonder when
// SKYLINE_CAPTURE names a directory: built at mid and close zoom, by day and
// by night (a few frames each, so the lamps turn), and the ground ages' wonder
// under way at a fifth, half and all of its cost.
func TestWonderCaptures(t *testing.T) {
	dir := os.Getenv("SKYLINE_CAPTURE")
	if dir == "" {
		t.Skip("set SKYLINE_CAPTURE to a directory to write captures")
	}
	c := catalog()
	write := func(name string, m *mapmodel.Model, x int, z zoom, w, h int) {
		v := newView()
		v.zoom, v.follow = z, false
		v.cam = x - z.cols(w)/2
		var frames []string
		for k := 0; k < 6; k++ {
			scr := draw(v, m, w, h, 40+k*5, mapmodel.TierUnicode, false)
			frames = append(frames, capture.Frame(scr))
			if k == 0 {
				_ = capture.PNG(filepath.Join(dir, name+".png"), scr)
			}
		}
		bg, fg := capture.Hex(theme.Color(theme.RoleBackground)), capture.Hex(theme.Color(theme.RoleText))
		_ = os.WriteFile(filepath.Join(dir, name+".html"), []byte(capture.HTML(name, bg, fg, capture.Fonts, frames, 6)), 0o644)
	}
	for a := range c.ByAge {
		key := wonderKeyOf(a)
		for _, night := range []bool{false, true} {
			tod, when := 0.5, "day"
			if night {
				tod, when = 0.95, "night"
			}
			m := wonderModel(fixture.Options{Age: c.Ages[a], Seed: 3, Routes: 2, Tick: tickAt(tod)}, true, 0)
			x := wonderX(m.Skyline.Districts[a], key)
			if a == len(c.ByAge)-1 { // the mandala has no lots to centre on
				continue
			}
			for _, z := range zoomSteps {
				if z != zMid && z != zClose || m.Sky() != mapmodel.SkyGround && z == zClose {
					continue
				}
				write(fmt.Sprintf("wonder_%02d_%s_%s_%s", a, key, z, when), m, x, z, 100, 40)
			}
		}
		if a < groundAges {
			for _, share := range []float64{0.2, 0.5, 1} {
				m := wonderModel(fixture.Options{Age: c.Ages[a], Seed: 3, Routes: 2, Tick: tickAt(0.5)}, false, share)
				write(fmt.Sprintf("wonder_%02d_%s_site_%d", a, key, int(share*100)), m, wonderX(m.Skyline.Districts[a], key), zMid, 100, 40)
			}
		}
	}
}
