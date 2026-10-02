package skyline

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// moverModels returns busy copies of m: one with m's own trade routes and
// one each running five land, sea and air routes, all with soldiers,
// wealth, workers and traffic enough to fill every stream. A town with no
// bay yet (bays come with the Colonial district) gets one on its first
// district, so the early bands' boats are exercised too.
func moverModels(m *mapmodel.Model) []*mapmodel.Model {
	base := *m
	a := &base.Activity
	a.Soldiers, a.Wealth, a.Traffic, a.Staffed = 400, 1, 1, 5000
	bay := false
	for _, d := range base.Skyline.Districts {
		bay = bay || d.Bay
	}
	if !bay && len(base.Skyline.Districts) > 0 {
		ds := append([]mapmodel.District(nil), base.Skyline.Districts...)
		ds[0].Bay, ds[0].BayX, ds[0].BayW = true, ds[0].X0+ds[0].LandW, 20
		base.Skyline.Districts = ds
	}
	out := []*mapmodel.Model{&base}
	for _, mode := range []mapmodel.RouteMode{mapmodel.ModeLand, mapmodel.ModeSea, mapmodel.ModeAir} {
		mm := base
		mm.Routes = nil
		for i := 0; i < 5; i++ {
			mm.Routes = append(mm.Routes, mapmodel.Route{Key: "r" + strconv.Itoa(i), Mode: mode})
		}
		mm.Activity.Routes = 5
		out = append(out, &mm)
	}
	return out
}

// sweep calls fn with the traffic of m over the whole panorama (cameras
// 80 columns apart on a view 160 wide, so every bay and launch pad passes
// through it) at a spread of frames.
func sweep(m *mapmodel.Model, fn func(vs []vehicle)) {
	for cam := 0; cam < m.Skyline.Width; cam += 80 {
		for _, anim := range []int{0, 9, 40, 77, 133, 400} {
			fn(trafficFor(m, anim, cam, 160, 39))
		}
	}
}

// TestNoMoverBeforeItsAge: no tagged vehicle shows before the age that
// introduces its mover in the shared roster (mapmodel/movers.go), in any
// stream: trade routes by land, sea and air, foot, soldiers, private
// traffic, the sky, the bays and the launch pads. The works wait for their
// movers too: no viaduct before the railway, no maglev rails before the
// maglev, no launches before the shuttle.
func TestNoMoverBeforeItsAge(t *testing.T) {
	seen := map[mapmodel.Mover]bool{}
	for ai, age := range config.AgeOrder() {
		m := build(fixture.Options{Age: age, Seed: 4, Routes: 5, Wars: 1, Scale: 2, Tick: tickAt(0.5)}, 0)
		if m.AgeIdx != ai {
			t.Fatalf("%s: model age %d, want %d", age, m.AgeIdx, ai)
		}
		bad := map[*vtemplate]bool{}
		for _, mm := range moverModels(m) {
			sweep(mm, func(vs []vehicle) {
				for _, v := range vs {
					k, tagged := moverTags[v.t]
					if !tagged {
						continue
					}
					seen[k] = true
					if !mapmodel.Introduced(k, ai) && !bad[v.t] {
						bad[v.t] = true
						t.Errorf("%s: a %s (kind %d, depth %d) shows before its age (from %d)",
							age, k.Info().Key, v.kind, v.depth, k.Info().From)
					}
				}
			})
		}
	}
	for _, k := range moverTags {
		// satellites and the tether's climbers are drawn by traffic(), not trafficFor
		if !seen[k] && k != mapmodel.MoverSatellite && k != mapmodel.MoverClimber {
			t.Errorf("no %s in any age: the sweep missed a stream", k.Info().Key)
		}
	}

	// the viaduct deck and the maglev rails, by their rows and depths
	works := func(age string) (viaduct, rails int) {
		m := build(fixture.Options{Age: age, Seed: 4, Routes: 5, Tick: tickAt(0.5)}, 0)
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 5}, 160, 45)
		for x := 0; x < s.W; x++ {
			if c := v.fb.at(x, s.Y(viaductY(s.groundY))); c != nil && c.d == dLane0+1 {
				viaduct++
			}
			for _, r := range railYs(s.groundY) {
				if c := v.fb.at(x, s.Y(r.y)); c != nil && c.d == r.d {
					rails++
				}
			}
		}
		return viaduct, rails
	}
	if n, _ := works("colonial_age"); n != 0 {
		t.Errorf("colonial_age: %d viaduct cells before the railway", n)
	}
	if n, _ := works("industrial_age"); n == 0 {
		t.Error("industrial_age: no viaduct")
	}
	if _, n := works("digital_age"); n != 0 {
		t.Errorf("digital_age: %d maglev rail cells before the maglev", n)
	}
	if _, n := works("cyberpunk_age"); n == 0 {
		t.Error("cyberpunk_age: no maglev rails")
	}

	// launches: the Fusion Age has a launch site (the Space Program) but no
	// shuttle yet; the Space Age launches
	for _, c := range []struct {
		age  string
		want bool
	}{{"fusion_age", false}, {"space_age", true}} {
		st := fixture.State(fixture.Options{Age: c.age, Seed: 4})
		if b, ok := st.Buildings["space_program"]; !ok || b.Count == 0 {
			st.Buildings["space_program"] = game.BuildingState{Name: "Space Program", Category: "wonder",
				AgeKey: "modern_age", Count: 1, Unlocked: true}
		}
		m := mapmodel.NewBuilder(catalog()).Build(&st, nil)
		if m.Building("space_program") == nil {
			t.Fatalf("%s: no launch site to test with", c.age)
		}
		got := false
		for cam := 0; cam < m.Skyline.Width; cam += 40 {
			for anim := 0; anim < 160; anim += 5 {
				got = got || len(launches(m, anim, cam, 160, 39, bandOf(m.AgeIdx))) > 0
			}
		}
		if got != c.want {
			t.Errorf("%s: launches %v, want %v", c.age, got, c.want)
		}
	}

	// the compact view's sky markers
	for _, c := range []struct {
		age  int
		want mapmodel.Sym
	}{{10, mapmodel.SymNone}, {11, mapmodel.SymNone}, {12, mapmodel.SymPlane}, {13, mapmodel.SymPlane},
		{14, mapmodel.SymDrone}, {15, mapmodel.SymDrone}, {16, mapmodel.SymCar}, {19, mapmodel.SymRocket}} {
		if got := compactSkySym(bandOf(c.age), c.age); got != c.want {
			t.Errorf("age %d: compact sky marker %v, want %v", c.age, got, c.want)
		}
	}
	atomic := build(fixture.Options{Age: "atomic_age", Seed: 4, Routes: 5, Tick: tickAt(0.5)}, 0)
	plane := mapmodel.R(mapmodel.SymPlane, mapmodel.TierUnicode)
	for _, anim := range []int{3, 40, 90} {
		if txt := capture.Text(draw(newView(), atomic, 40, 15, anim, mapmodel.TierUnicode, true)); strings.ContainsRune(txt, plane) {
			t.Errorf("atomic_age compact frame %d: a plane before the Modern Age", anim)
		}
	}
}

// TestMoverFallbacks pins what a band shows in the ages before its own
// vehicles arrive: the band below's, on lanes the frame draws.
func TestMoverFallbacks(t *testing.T) {
	type pick func(v vehicle) bool
	land := func(v vehicle) bool { return v.kind == vkRoute && v.depth != dShip }
	ships := func(v vehicle) bool { return v.kind == vkRoute && v.depth == dShip }
	boats := func(v vehicle) bool { return v.kind == vkAmbient && v.depth == dShip }
	kind := func(k vkind) pick { return func(v vehicle) bool { return v.kind == k && v.depth != dShip } }
	models := map[string][]*mapmodel.Model{}
	for _, c := range []struct {
		age   string
		model int // moverModels index: 1 land, 2 sea, 3 air routes
		which pick
		want  []*vtemplate
		max   int // at most this many (0: any number)
	}{
		{"bronze_age", 1, land, []*vtemplate{&tPorters}, 0},
		{"iron_age", 1, land, []*vtemplate{&tOxCart}, 0},
		{"colonial_age", 1, land, []*vtemplate{&tTradeCoach}, 0},
		{"industrial_age", 1, land, []*vtemplate{&tTradeCoach, &tSteamTrain}, 0},
		{"atomic_age", 1, land, []*vtemplate{&tTram}, 0},
		{"atomic_age", 3, land, []*vtemplate{&tTram}, 0}, // air routes go overland before the plane
		{"modern_age", 1, land, []*vtemplate{&tTruck}, 0},
		{"modern_age", 3, land, []*vtemplate{&tCargoJet}, 0},
		{"digital_age", 1, land, []*vtemplate{&tTruck}, 0},
		{"digital_age", 3, land, []*vtemplate{&tCargoJet}, 0},
		{"cyberpunk_age", 1, land, []*vtemplate{&tMaglev}, 0},
		{"cyberpunk_age", 3, land, []*vtemplate{&tFlyCargo}, 0},
		{"digital_age", 1, kind(vkPrivate), []*vtemplate{&tCar}, 0},
		{"cyberpunk_age", 1, kind(vkPrivate), []*vtemplate{&tFlyCar}, 0},
		{"atomic_age", 1, kind(vkAmbient), []*vtemplate{&tZeppelin, &tBiplane}, 3},
		{"modern_age", 1, kind(vkAmbient), []*vtemplate{&tJet, &tAirliner, &tHeli}, 0},
		{"digital_age", 1, kind(vkAmbient), []*vtemplate{&tDrone, &tHeli}, 0}, // the drones arrive
		{"cyberpunk_age", 1, kind(vkAmbient), []*vtemplate{&tDrone}, 0},
		{"digital_age", 1, kind(vkArmy), []*vtemplate{&tJeep, &tWarband}, 0},
		{"cyberpunk_age", 1, kind(vkArmy), []*vtemplate{&tHoverTank, &tWarband}, 0},
		{"bronze_age", 2, ships, []*vtemplate{&tCanoe}, 0},
		{"bronze_age", 2, boats, []*vtemplate{&tCanoe}, 0},
		{"iron_age", 2, ships, []*vtemplate{&tTrireme}, 0},
		{"medieval_age", 2, ships, []*vtemplate{&tTrireme}, 0},
		{"medieval_age", 2, boats, []*vtemplate{&tTrireme}, 0},
		{"renaissance_age", 2, ships, []*vtemplate{&tSail}, 0},
		{"atomic_age", 2, ships, []*vtemplate{&tSteamer}, 0},
		{"atomic_age", 2, boats, []*vtemplate{&tBarge}, 0},
		{"modern_age", 2, ships, []*vtemplate{&tFreighter}, 0},
		{"digital_age", 2, boats, []*vtemplate{&tBarge}, 0},
		{"cyberpunk_age", 2, boats, []*vtemplate{&tHoverCar}, 0},
	} {
		if models[c.age] == nil {
			models[c.age] = moverModels(build(fixture.Options{Age: c.age, Seed: 4, Routes: 5, Wars: 1, Tick: tickAt(0.5)}, 0))
		}
		m := models[c.age][c.model]
		ok := map[*vtemplate]bool{}
		for _, w := range c.want {
			ok[w] = true
		}
		found := 0
		sweep(m, func(vs []vehicle) {
			n := 0
			for _, v := range vs {
				if !c.which(v) {
					continue
				}
				n++
				found++
				if !ok[v.t] {
					t.Errorf("%s (routes %d): drew %+v, want one of %d templates", c.age, c.model, *v.t, len(c.want))
					return
				}
			}
			if c.max > 0 && n > c.max {
				t.Errorf("%s: %d vehicles, want at most %d", c.age, n, c.max)
			}
		})
		if found == 0 {
			t.Errorf("%s (routes %d): nothing drawn in the stream", c.age, c.model)
		}
	}
	// the Industrial Age has its railway: every third land route is a train
	trains := 0
	sweep(models["industrial_age"][1], func(vs []vehicle) {
		for _, v := range vs {
			if v.t == &tSteamTrain {
				trains++
			}
		}
	})
	if trains == 0 {
		t.Error("industrial_age: no steam train on the viaduct")
	}
}

// flyby finds the first flyby visit of a late (Space Age on) town.
func flyby(seed int64) mapmodel.Sighting {
	s := mapmodel.NextSighting(seed, true, 0)
	for s.Kind != mapmodel.SightFlyby {
		s = mapmodel.NextSighting(seed, true, s.Start+s.Frames)
	}
	return s
}

// domeAt draws m at frame anim (160x45) and reports what the screen shows
// at the saucer's dome, and whether something nearer than the sky stands
// in front of it there. ok is false when no visit is showing. The view is
// returned with the frame still in its buffer.
func domeAt(m *mapmodel.Model, anim int, tier mapmodel.GlyphTier) (v *view, scr tcell.SimulationScreen, x, y int, r rune, hidden, ok bool) {
	v = newView()
	scr = draw(v, m, 160, 45, anim, tier, false)
	sx, sy, ok := saucerAt(m, anim, 160)
	if !ok {
		return v, scr, 0, 0, 0, false, false
	}
	x, y = sx+1, sy+1 // the dome; scene row 0 is screen row 1
	r, _, _, _ = scr.GetContent(x, y)
	c := v.fb.at(x, y)
	return v, scr, x, y, r, c != nil && c.d < dAir, true
}

// TestVisitorSaucer: while a visit shows, a saucer crosses or hovers in the
// sky, behind anything nearer; outside a visit there is none, in any age.
func TestVisitorSaucer(t *testing.T) {
	m := build(fixture.Options{Age: "space_age", Seed: 4, Tick: tickAt(0.95)}, 0)
	clear := *m
	clear.Weather = mapmodel.Weather{Kind: mapmodel.Clear} // no rain streaks over it
	m = &clear
	dome := mapmodel.R(mapmodel.SymUFO, mapmodel.TierUnicode)
	sg := flyby(m.Seed)
	mid := sg.Start + sg.Frames/2
	if got, ok := m.SightingAt(mid); !ok || got.Start != sg.Start {
		t.Fatalf("frame %d is not inside the visit at %d", mid, sg.Start)
	}
	shown := -1
	for anim := mid - 8; anim <= mid+8; anim++ {
		_, _, x, y, r, hidden, ok := domeAt(m, anim, mapmodel.TierUnicode)
		switch {
		case !ok:
			t.Fatalf("frame %d: no saucer mid-visit", anim)
		case x < 0 || x >= 160 || y < 1:
			t.Fatalf("frame %d: mid-flyby the dome is off screen at (%d,%d)", anim, x, y)
		case r == dome:
			if shown < 0 {
				shown = anim
			}
		case !hidden:
			t.Errorf("frame %d: %q at the dome (%d,%d) and nothing in front of it", anim, r, x, y)
		}
	}
	if shown < 0 {
		t.Fatal("the saucer never showed mid-flyby")
	}

	// the wings, in every tier
	for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierNerd, mapmodel.TierASCII} {
		v, scr, x, y, r, _, _ := domeAt(m, shown, tier)
		want := []rune{mapmodel.Fold('◄', tier), mapmodel.R(mapmodel.SymUFO, tier), mapmodel.Fold('►', tier)}
		if tier == mapmodel.TierASCII && string(want) != "<O>" {
			t.Errorf("the ascii saucer folds to %q, want \"<O>\"", string(want))
		}
		if r != want[1] {
			t.Errorf("%s: the dome is %q, want %q", tier, r, want[1])
		}
		for i, dx := range []int{-1, 1} {
			c := v.fb.at(x+dx, y)
			if c == nil || c.d < dAir {
				continue // off the screen, or something nearer stands there
			}
			if w, _, _, _ := scr.GetContent(x+dx, y); w != want[i*2] {
				t.Errorf("%s: wing %q at (%d,%d), want %q", tier, w, x+dx, y, want[i*2])
			}
		}
	}

	// it flies behind buildings and the ridge towns and in front of the
	// clouds: stand a tower, a ridge town and a cloud where it is and
	// draw it again
	{
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: shown}, 160, 45)
		x, y, _ := saucerAt(m, shown, 160)
		red := theme.SkyColor(theme.SkyWarRed)
		depths := [saucerW]uint8{dRow0, dRidgeTown, dCloud} // wing, dome, wing
		for i, d := range depths {
			v.fb.set(x+i, s.Y(y), '█', red, red, d)
		}
		s.visitor()
		for i, d := range depths {
			c := v.fb.at(x+i, s.Y(y))
			if c == nil {
				t.Fatalf("mid-flyby the saucer is off screen at column %d", x+i)
			}
			kept := c.ch == '█' && c.d == d
			switch {
			case d < dAir && !kept:
				t.Errorf("the saucer drew over a cell at depth %d, nearer than the sky", d)
			case d > dAir && kept:
				t.Errorf("the saucer hid behind a cloud (depth %d)", d)
			}
		}
	}

	// legible in every theme, and never in a legend
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		_, scr, x, y, r, _, _ := domeAt(m, shown, mapmodel.TierUnicode)
		for dx := -1; dx <= 1; dx++ {
			c, _, st, _ := scr.GetContent(x+dx, y)
			if fg, bg, _ := st.Decompose(); r == dome && c != ' ' && fg == bg {
				t.Errorf("theme %s: %q at (%d,%d) is drawn in its background colour", th.Key, c, x+dx, y)
			}
		}
	}
	_ = theme.SetActive(orig)
	lv := newView()
	lv.SetOption(mapstyle.OptLegend, true)
	if scr := draw(lv, m, 160, 45, shown, mapmodel.TierUnicode, false); strings.ContainsRune(rowText(scr, 44), dome) {
		t.Error("the saucer is in the legend")
	}

	// outside a visit there is no saucer
	for _, anim := range []int{7, sg.Start - 1, sg.Start + sg.Frames} {
		if _, ok := m.SightingAt(anim); ok {
			t.Fatalf("frame %d: a visit is showing", anim)
		}
		if txt := capture.Text(draw(newView(), m, 160, 45, anim, mapmodel.TierUnicode, false)); strings.ContainsRune(txt, dome) {
			t.Errorf("frame %d: a saucer with no visit showing", anim)
		}
	}

	// the first ages: none while nothing shows, and the long-odds joke
	// visit draws the saucer like a late one
	pm := build(fixture.Options{Age: "primitive_age", Seed: 4, Tick: tickAt(0.5)}, 0)
	if _, ok := pm.SightingAt(40); ok {
		t.Fatal("primitive_age: a visit at frame 40")
	}
	if txt := capture.Text(draw(newView(), pm, 160, 45, 40, mapmodel.TierUnicode, false)); strings.ContainsRune(txt, dome) {
		t.Error("primitive_age: a saucer with no visit showing")
	}
	js := mapmodel.NextSighting(pm.Seed, false, 0)
	jmid := js.Start + js.Frames/2
	if _, ok := pm.SightingAt(jmid); !ok {
		t.Fatalf("primitive_age: no visit at frame %d", jmid)
	}
	if _, _, x, y, r, hidden, ok := domeAt(pm, jmid, mapmodel.TierUnicode); !ok || (r != dome && !hidden) {
		t.Errorf("primitive_age: the joke visit drew %q at (%d,%d), no saucer", r, x, y)
	}
}

// rowText is one screen row as text.
func rowText(scr tcell.SimulationScreen, y int) string {
	var b strings.Builder
	w, _ := scr.Size()
	for x := 0; x < w; x++ {
		r, _, _, _ := scr.GetContent(x, y)
		b.WriteRune(r)
	}
	return b.String()
}

// TestVisitorInspect: ↑ past the ridge puts the cursor on the saucer while
// it is in view, and Inspect names it an unknown craft of KindAlien; Tab
// never reaches it, ↓ leaves it, and the cursor lets go once it is gone.
func TestVisitorInspect(t *testing.T) {
	m := build(fixture.Options{Age: "space_age", Seed: 4, Tick: tickAt(0.5)}, 0)
	sg := mapmodel.NextSighting(m.Seed, true, 0)
	for sg.Kind == mapmodel.SightFlyby { // a hover sits still mid-visit
		sg = mapmodel.NextSighting(m.Seed, true, sg.Start+sg.Frames)
	}
	mid := sg.Start + sg.Frames/2
	f := mapstyle.Frame{Model: m, Anim: mid}
	v := newView()
	v.SetOption(mapstyle.OptInspect, true)
	_ = draw(v, m, 160, 45, mid, mapmodel.TierUnicode, false)
	press := func(k tcell.Key, f mapstyle.Frame) { v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f) }
	for i := 0; i < 8; i++ {
		press(tcell.KeyUp, f)
	}
	in, ok := v.Inspect(f)
	if !ok || in.Title != "Unknown craft" || len(in.Lines) != 1 || in.Lines[0] != "Not one of ours." ||
		in.Kind != mapstyle.KindAlien || in.Command != "" {
		t.Fatalf("↑ past the ridge did not reach the saucer: %+v", in)
	}
	scr := draw(v, m, 160, 45, mid, mapmodel.TierUnicode, false)
	if !strings.Contains(rowText(scr, 44), "Unknown craft") {
		t.Errorf("the status line does not name the craft: %q", rowText(scr, 44))
	}
	if x, y, _ := saucerAt(m, mid, 160); y >= 1 {
		if r, _, _, _ := scr.GetContent(x+1, y); r != '▼' {
			t.Errorf("no cursor mark over the saucer: %q", r)
		}
	}
	press(tcell.KeyDown, f)
	if in, _ := v.Inspect(f); in.Kind == mapstyle.KindAlien || v.cur.kind == tUFO {
		t.Error("↓ did not leave the saucer")
	}
	for i := 0; i < len(m.Skyline.Lots)+20; i++ {
		press(tcell.KeyTab, f)
		if in, _ := v.Inspect(f); in.Kind == mapstyle.KindAlien {
			t.Fatal("Tab reached the saucer")
		}
	}
	// the visit ends: the cursor lets go, and ↑ finds nothing up there
	v.cur = target{kind: tUFO}
	after := mapstyle.Frame{Model: m, Anim: sg.Start + sg.Frames}
	if in, ok := v.Inspect(after); !ok || in.Kind == mapstyle.KindAlien {
		t.Errorf("the cursor stayed on a saucer that has gone: %+v", in)
	}
	for i := 0; i < 8; i++ {
		press(tcell.KeyUp, after)
	}
	if v.cur.kind == tUFO {
		t.Error("↑ reached a saucer with no visit showing")
	}
}
