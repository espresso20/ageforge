package skyline

import (
	"strings"
	"testing"
	"time"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

var allTiers = []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII, mapmodel.TierNerd}

// TestRosterEveryAge: the age constants are the game's ages in order, and
// every age has a roster: something on its road, and entries that make sense.
func TestRosterEveryAge(t *testing.T) {
	order := config.AgeOrder()
	if len(order) != numAges {
		t.Fatalf("the game has %d ages, the roster %d", len(order), numAges)
	}
	for k, want := range map[int]string{agePrimitive: "primitive_age", ageStone: "stone_age", ageBronze: "bronze_age",
		ageIron: "iron_age", ageClassical: "classical_age", ageMedieval: "medieval_age", ageRenaissance: "renaissance_age",
		ageColonial: "colonial_age", ageIndustrial: "industrial_age", ageVictorian: "victorian_age",
		ageElectric: "electric_age", ageAtomic: "atomic_age", ageModern: "modern_age", ageInformation: "information_age",
		ageDigital: "digital_age", ageCyberpunk: "cyberpunk_age", ageFusion: "fusion_age", ageSpace: "space_age",
		ageInterstellar: "interstellar_age", ageGalactic: "galactic_age", ageQuantum: "quantum_age",
		ageTranscendent: "transcendent_age"} {
		if order[k] != want {
			t.Errorf("age %d is %s, the roster says %s", k, order[k], want)
		}
	}
	for a := 0; a < numAges; a++ {
		if n := len(ageRoster[a][laneRoad].common); n == 0 {
			t.Errorf("%s: nothing common on the road", order[a])
		}
		total := 0
		for l := laneKind(0); l < numLanes; l++ {
			total += len(ageRoster[a][l].common) + len(ageRoster[a][l].rare)
		}
		if total < 2 {
			t.Errorf("%s: a roster of %d", order[a], total)
		}
	}
	for _, e := range rosterEntries {
		if e.from < 0 || e.from > e.until || e.until > e.rare || e.rare >= numAges || e.lo <= 0 || e.hi < e.lo || len(e.art) == 0 {
			t.Errorf("%s: a malformed entry %+v", e.name, e.from)
		}
	}
}

// TestRosterLanes: the rail lane from the Industrial Age, the sky lane from
// the age of flight (the Electric Age), the road always; and each age's
// listed vehicles are in its roster.
func TestRosterLanes(t *testing.T) {
	has := func(age int, lane laneKind, name string, rare bool) bool {
		l := ageRoster[age][lane].common
		if rare {
			l = ageRoster[age][lane].rare
		}
		for _, e := range l {
			if e.name == name {
				return true
			}
		}
		return false
	}
	for a := 0; a < numAges; a++ {
		rail := len(ageRoster[a][laneRail].common) + len(ageRoster[a][laneRail].rare)
		sky := len(ageRoster[a][laneSky].common) + len(ageRoster[a][laneSky].rare)
		if a < ageIndustrial && rail != 0 {
			t.Errorf("age %d: %d trains before the Industrial Age", a, rail)
		}
		if a < ageElectric && sky != 0 {
			t.Errorf("age %d: %d flyers before flight", a, sky)
		}
		if hasRailLane(a) != (rail > 0) {
			t.Errorf("age %d: hasRailLane disagrees with the roster", a)
		}
	}
	for _, c := range []struct {
		age  int
		lane laneKind
		name string
	}{
		{agePrimitive, laneRoad, "bundle walker"}, {ageStone, laneRoad, "pack dog"}, {ageStone, laneRoad, "goat"},
		{ageBronze, laneRoad, "ox cart"}, {ageBronze, laneRoad, "camel"}, {ageIron, laneRoad, "chariot"},
		{ageClassical, laneRoad, "horse and rider"}, {ageMedieval, laneRoad, "wagon"}, {ageRenaissance, laneRoad, "carriage"},
		{ageColonial, laneRoad, "carriage"}, {ageIndustrial, laneRoad, "carriage"},
		{ageIndustrial, laneRail, "steam train"}, {ageVictorian, laneRail, "steam train"}, {ageVictorian, laneRoad, "tram"},
		{ageElectric, laneRoad, "motor car"}, {ageElectric, laneSky, "biplane"},
		{ageAtomic, laneRoad, "car"}, {ageAtomic, laneRoad, "bus"}, {ageAtomic, laneRail, "train"},
		{ageAtomic, laneSky, "propeller plane"}, {ageAtomic, laneSky, "jet"}, {ageModern, laneSky, "jet"},
		{ageModern, laneRoad, "bus"}, {ageModern, laneRail, "train"},
		{ageInformation, laneRoad, "car"}, {ageInformation, laneRail, "bullet train"}, {ageInformation, laneSky, "jet"},
		{ageDigital, laneRail, "bullet train"}, {ageDigital, laneSky, "helicopter"},
		{ageCyberpunk, laneRoad, "hover bike"}, {ageCyberpunk, laneRail, "maglev"}, {ageCyberpunk, laneSky, "drone"},
		{ageFusion, laneRoad, "hover car"}, {ageFusion, laneSky, "shuttle"}, {ageSpace, laneRoad, "hover car"},
		{ageSpace, laneSky, "shuttle"}, {ageInterstellar, laneRoad, "hover car"}, {ageInterstellar, laneSky, "starship"},
		{ageInterstellar, laneRail, "portal tram"}, {ageTranscendent, laneSky, "starship"},
		{ageTranscendent, laneRail, "portal tram"},
	} {
		if !has(c.age, c.lane, c.name, false) {
			t.Errorf("age %d: no %s among the common traffic", c.age, c.name)
		}
	}
	// older vehicles linger rarely: a horse in the Modern Age, a steam train too
	if !has(ageModern, laneRoad, "horse and rider", true) || !has(ageModern, laneRail, "steam train", true) {
		t.Error("the Modern Age has no rare horse or steam train")
	}
	if has(ageModern, laneRoad, "horse and rider", false) {
		t.Error("a horse is common in the Modern Age")
	}
}

// TestRosterSprites: every drawing is 1 to 4 cells wide and 1 to 2 tall (a
// train: every unit), the same size in every glyph tier and frame, with a
// mask to match, nothing the ASCII tier cannot show, one cell for the Nerd
// tier's mark, and fits its lane.
func TestRosterSprites(t *testing.T) {
	check := func(name string, tp *vtemplate, maxW int) {
		t.Helper()
		if len(tp.rows) < 1 || len(tp.rows) > 2 {
			t.Errorf("%s: %d rows", name, len(tp.rows))
		}
		if len(tp.mask) != len(tp.rows) {
			t.Errorf("%s: %d mask rows for %d rows", name, len(tp.mask), len(tp.rows))
		}
		if w := tp.width(); w < 1 || w > maxW {
			t.Errorf("%s: %d cells wide", name, w)
		}
		for _, tier := range allTiers {
			for frame := 0; frame < 2; frame++ {
				v := vehicle{t: tp, tier: tier, frame: frame}
				rows := v.rowsNow()
				if len(rows) != len(tp.rows) {
					t.Errorf("%s (tier %d, frame %d): %d rows, want %d", name, tier, frame, len(rows), len(tp.rows))
					continue
				}
				for r, row := range rows {
					rr := []rune(row)
					if len(rr) != len([]rune(tp.rows[r])) || len(rr) != len([]rune(tp.mask[r])) {
						t.Errorf("%s (tier %d, frame %d): row %d is %d cells, the drawing %d, the mask %d", name, tier, frame, r,
							len(rr), len([]rune(tp.rows[r])), len([]rune(tp.mask[r])))
					}
					for _, ch := range rr {
						if f := mapmodel.Fold(ch, tier); tier == mapmodel.TierASCII && (f >= 0x80 || f == '?' && ch != '?') {
							t.Errorf("%s: %q has no ASCII form (folds to %q)", name, ch, f)
						}
						if ch != ' ' && uniseg.StringWidth(string(ch)) != 1 {
							t.Errorf("%s: %q is not one cell wide", name, ch)
						}
					}
				}
			}
		}
		if tp.sym != mapmodel.SymNone {
			if r := []rune(string(mapmodel.R(tp.sym, mapmodel.TierNerd))); len(r) != 1 {
				t.Errorf("%s: the Nerd mark is %d runes", name, len(r))
			}
		}
	}
	for _, e := range rosterEntries {
		for i, tp := range e.art {
			check(e.name, tp, 4+20) // the longest train: four carriages and an engine
			_ = i
		}
		for _, p := range e.parts {
			check(e.name+" unit", p, 4)
		}
		// fits its lane: the road, the rail and the sky are two rows deep
		for _, tp := range e.art {
			if len(tp.rows) > 2 {
				t.Errorf("%s does not fit its lane", e.name)
			}
		}
	}
	for _, p := range []*vtemplate{pMan, pWoman, pChild, pBundle, pSpear} {
		check("person", p, 4)
	}
	for _, t1 := range []*vtemplate{tpDog, tpGoat} {
		check("animal", t1, 4)
	}
	for e := range groupArt {
		for n := range groupArt[e] {
			for v, g := range groupArt[e][n] {
				check("group", g, 5*4+4)
				if !g.people {
					t.Errorf("group %d/%d/%d is not marked as people", e, n, v)
				}
			}
		}
	}
}

// groupSize counts the heads of a group's drawing.
func groupSize(g *vtemplate) int {
	n := 0
	for _, ch := range strings.Join(g.rows, "") {
		if ch == '●' {
			n++
		}
		if ch == '☻' {
			n++
		}
	}
	return n
}

// peopleOf is the groups of people in a frame's traffic.
func peopleOf(vs []vehicle) (out []vehicle) {
	for _, v := range vs {
		if v.t.people {
			out = append(out, v)
		}
	}
	return out
}

// TestPeopleTwoCell: from 40 rows up, people are groups of two to five
// figures a head over a body, at different speeds; below 40 rows the
// one-cell markers stay and none of the roster is drawn.
func TestPeopleTwoCell(t *testing.T) {
	for _, age := range []string{"primitive_age", "iron_age", "industrial_age", "modern_age", "cyberpunk_age"} {
		m := build(fixture.Options{Age: age, Seed: 4, Routes: 3, Tick: tickAt(0.5), Scale: 2}, 0)
		m2 := *m
		m2.Activity.Staffed = 5000
		for _, h := range []int{40, 45, 52} {
			gy := h - 6
			groups := peopleOf(trafficFor(&m2, 5, 0, 160, gy))
			if len(groups) < 3 {
				t.Errorf("%s at %d rows: %d groups of people", age, h, len(groups))
			}
			for _, g := range groups {
				if n := groupSize(g.t); n < 2 || n > 5 {
					t.Errorf("%s at %d rows: a group of %d", age, h, n)
				}
				if len(g.t.rows) != 2 || g.t.rows[0] == "" || !strings.ContainsAny(g.t.rows[0], "●") {
					t.Errorf("%s at %d rows: a group is not two cells tall: %q", age, h, g.t.rows)
				}
				if g.y != gy+1 || g.depth != dStreet {
					t.Errorf("%s: a group stands at row %d depth %d, want the street (%d, %d)", age, g.y, g.depth, gy+1, dStreet)
				}
			}
			speeds := map[int]bool{}
			for _, g := range groups {
				// the speed shows as the distance walked in 40 frames
				a := trafficFor(&m2, 5, 0, 160, gy)
				b := trafficFor(&m2, 45, 0, 160, gy)
				for i := range a {
					if a[i].t == g.t && a[i].seed == g.seed {
						speeds[b[i].x-a[i].x] = true
					}
				}
			}
			if len(speeds) < 2 {
				t.Errorf("%s at %d rows: every group walks at the same speed", age, h)
			}
		}
		// below 40 rows: the old markers, none of the roster
		for _, h := range []int{24, 30, 39} {
			vs := trafficFor(&m2, 5, 0, 160, h-6)
			if n := countKind(vs, vkLife); n != 0 {
				t.Errorf("%s at %d rows: %d roster vehicles on a short view", age, h, n)
			}
			if countKind(vs, vkFoot) == 0 {
				t.Errorf("%s at %d rows: no one-cell walkers", age, h)
			}
			for _, v := range vs {
				if v.kind == vkFoot && len(v.t.rows) != 1 {
					t.Errorf("%s at %d rows: a %d-row walker", age, h, len(v.t.rows))
				}
			}
		}
	}
	// drawn: the head sits over the body, at every glyph tier
	m := build(fixture.Options{Age: "medieval_age", Seed: 4, Tick: tickAt(0.5), Scale: 2}, 0)
	m.Activity.Staffed = 5000
	for _, tier := range allTiers {
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 9, Tier: tier}, 160, 45)
		head := '●'
		if tier == mapmodel.TierASCII {
			head = 'o'
		}
		heads := 0
		for x := 0; x < s.W; x++ {
			c := v.fb.at(x, s.Y(s.groundY))
			if c != nil && c.ch == head && c.d == dStreet {
				heads++
				if b := v.fb.at(x, s.Y(s.groundY+1)); b == nil || b.d != dStreet || b.ch == ' ' {
					t.Errorf("tier %d: a head at column %d with no body under it", tier, x)
				}
			}
		}
		if heads < 4 {
			t.Errorf("tier %d: %d heads drawn", tier, heads)
		}
	}
	// and not on a short view
	v := newView()
	s := v.compose(mapstyle.Frame{Model: m, Anim: 9}, 160, 39)
	for x := 0; x < s.W; x++ {
		if c := v.fb.at(x, s.Y(s.groundY)); c != nil && c.ch == '●' && c.d == dStreet {
			t.Fatalf("a head drawn at 39 rows")
		}
	}
}

// TestPeopleGrowWithTheTown: more people are out in a bigger, busier town.
func TestPeopleGrowWithTheTown(t *testing.T) {
	m := build(fixture.Options{Age: "renaissance_age", Seed: 4, Tick: tickAt(0.5)}, 0)
	prev := -1
	for _, staffed := range []int{0, 2, 20, 200, 2000, 20000} {
		mm := *m
		mm.Activity.Staffed = staffed
		n := len(peopleOf(trafficFor(&mm, 5, 0, 160, 39)))
		if n < prev {
			t.Errorf("%d staffed: %d groups, fewer than before (%d)", staffed, n, prev)
		}
		if staffed == 0 && n != 0 {
			t.Errorf("%d groups with nobody at work", n)
		}
		prev = n
	}
	if prev < 5 {
		t.Errorf("a town of 20000 workers has %d groups", prev)
	}
	// the road, rail and sky fill up with the town's wealth and traffic
	quiet, busy := *m, *m
	quiet.Activity.Wealth, quiet.Activity.Traffic = 0, 0
	busy.Activity.Wealth, busy.Activity.Traffic = 1, 1
	qr, qa, qs := rosterCounts(&quiet)
	br, ba, bs := rosterCounts(&busy)
	if br <= qr || ba <= qa || bs <= qs {
		t.Errorf("a busy town has %d/%d/%d vehicles on its lanes, a quiet one %d/%d/%d", br, ba, bs, qr, qa, qs)
	}
}

// roster is a frame's roster vehicles by lane.
func rosterByLane(vs []vehicle) (out [numLanes][]vehicle) {
	for _, v := range vs {
		if v.kind != vkLife || v.t.people {
			continue
		}
		switch v.depth {
		case dAirLow:
			out[laneSky] = append(out[laneSky], v)
		case dStreet - 1:
			out[laneRail] = append(out[laneRail], v)
		default:
			out[laneRoad] = append(out[laneRoad], v)
		}
	}
	return out
}

// TestRosterOnScreen: each age's traffic, on the lanes, from the roster
// table: nothing before its age, the lanes at their rows and depths, the
// rare older ones showing now and then, and the rail track laid only where
// there are trains.
func TestRosterOnScreen(t *testing.T) {
	order := config.AgeOrder()
	for a := 0; a < ageSpace; a++ { // the ground panorama: the sky scenes draw their own
		m := build(fixture.Options{Age: order[a], Seed: 4, Routes: 3, Tick: tickAt(0.5), Scale: 2}, 0)
		m.Activity.Wealth, m.Activity.Traffic = 1, 1
		common := map[*vtemplate]bool{}
		rare := map[*vtemplate]bool{}
		for l := laneKind(0); l < numLanes; l++ {
			for _, e := range ageRoster[a][l].common {
				for _, tp := range e.art {
					common[tp] = true
				}
			}
			for _, e := range ageRoster[a][l].rare {
				for _, tp := range e.art {
					rare[tp] = true
				}
			}
		}
		seen := [numLanes]int{}
		rareSeen := 0
		for cam := 0; cam < m.Skyline.Width; cam += 40 {
			for _, anim := range []int{0, 30, 90, 200} {
				for _, gy := range []int{34, 39, 46} {
					by := rosterByLane(trafficFor(m, anim, cam, 160, gy))
					for l := laneKind(0); l < numLanes; l++ {
						for _, v := range by[l] {
							seen[l]++
							switch {
							case common[v.t]:
							case rare[v.t]:
								rareSeen++
							default:
								t.Fatalf("%s: a vehicle that is not in the age's roster", order[a])
							}
							switch l {
							case laneRoad:
								if v.y != gy+1 {
									t.Errorf("%s: a road vehicle at row %d, the road is %d", order[a], v.y, gy+1)
								}
							case laneRail:
								if v.y != gy+2 {
									t.Errorf("%s: a train at row %d, the rail is %d", order[a], v.y, gy+2)
								}
							case laneSky:
								if v.depth <= dRow2 || v.y-len(v.t.rows)+1 < 1 || v.y > gy*3/5 {
									t.Errorf("%s: a flyer at row %d depth %d", order[a], v.y, v.depth)
								}
							}
						}
					}
				}
			}
		}
		if seen[laneRoad] == 0 {
			t.Errorf("%s: nothing on the road", order[a])
		}
		if want := hasRailLane(a); (seen[laneRail] > 0) != want {
			t.Errorf("%s: %d trains drawn, rail lane %v", order[a], seen[laneRail], want)
		}
		if want := len(ageRoster[a][laneSky].common)+len(ageRoster[a][laneSky].rare) > 0; (seen[laneSky] > 0) != want {
			t.Errorf("%s: %d flyers drawn, sky lane %v", order[a], seen[laneSky], want)
		}
		hasRare := false
		for l := laneKind(0); l < numLanes; l++ {
			hasRare = hasRare || len(ageRoster[a][l].rare) > 0
		}
		if hasRare && rareSeen == 0 {
			t.Errorf("%s: its rare older vehicles never show", order[a])
		}
		// the track
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 5}, 160, 45)
		track := 0
		for x := 0; x < s.W; x++ {
			if c := v.fb.at(x, s.Y(s.groundY+2)); c != nil && (c.ch == '═' || c.ch == '╪') && c.d == dStreet {
				track++
			}
		}
		if hasRailLane(a) != (track > s.W/3) {
			t.Errorf("%s: %d columns of track, rail lane %v", order[a], track, hasRailLane(a))
		}
	}
}

// TestSkyLaneBehindBuildings: the flyers' cells are never drawn over a
// building row that is nearer: those cells are the same with and without
// the roster's sky lane, and the lane has flyers over the roofs too.
func TestSkyLaneBehindBuildings(t *testing.T) {
	m := build(fixture.Options{Age: "information_age", Seed: 7, Routes: 4, Tick: tickAt(0.5), Scale: 2}, 0)
	m.Activity.Traffic = 1
	hidden, shown := 0, 0
	for _, anim := range []int{3, 40, 77, 120, 300} {
		v := newView()
		v.noTraffic = true
		v.compose(mapstyle.Frame{Model: m, Anim: anim}, 160, 45)
		bare := append([]cell(nil), v.fb.c...)
		v.noTraffic = false
		s := v.compose(mapstyle.Frame{Model: m, Anim: anim}, 160, 45)
		for _, vh := range rosterByLane(trafficFor(m, anim, s.cam, s.W, s.groundY))[laneSky] {
			vh := vh
			vh.cells(func(x, y int, ch rune, mk byte) {
				y = s.Y(y)
				if x < 0 || x >= s.W || y < 0 || y >= s.H {
					return
				}
				i := y*s.W + x
				if bare[i].d <= dRow2 && bare[i].d >= dWonder {
					hidden++
					if v.fb.c[i] != bare[i] {
						t.Fatalf("anim %d: a flyer drew over a building cell at %d,%d", anim, x, y)
					}
				} else {
					shown++
				}
			})
		}
	}
	if hidden == 0 || shown == 0 {
		t.Errorf("the sky lane's occlusion is not exercised (hidden %d, shown %d)", hidden, shown)
	}
}

// TestSkylineFrameTime: the foreground's people and lanes keep a frame
// inside the 25 ms budget (the map redraws every 125 ms) at 160x45, in every
// roster band.
func TestSkylineFrameTime(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, age := range []string{"primitive_age", "iron_age", "renaissance_age", "industrial_age", "electric_age", "modern_age",
		"digital_age", "cyberpunk_age", "fusion_age"} {
		m := build(fixture.Options{Age: age, Seed: 7, Routes: 5, Tick: tickAt(0.5), Scale: 2}, 1)
		m.Activity.Staffed, m.Activity.Wealth, m.Activity.Traffic = 20000, 1, 1
		v := newView()
		scr := capture.NewScreen(160, 45)
		f := mapstyle.Frame{Model: m}
		v.Draw(scr, mapstyle.Rect{W: 160, H: 45}, f)
		const n = 60
		start := time.Now()
		for i := 0; i < n; i++ {
			f.Anim = i
			v.Draw(scr, mapstyle.Rect{W: 160, H: 45}, f)
		}
		per := time.Since(start) / n
		t.Logf("160x45 %s frame: %v", age, per)
		if per > 25*time.Millisecond {
			t.Errorf("%s: a frame takes %v", age, per)
		}
	}
}
