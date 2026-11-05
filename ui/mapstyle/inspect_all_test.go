package mapstyle

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// thingsOf lists one cell of every distinct thing a scene shows.
func thingsOf(g *mapgen.Grid) map[mapgen.Thing]mapgen.Pt {
	out := map[mapgen.Thing]mapgen.Pt{}
	for i, h := range g.Hit {
		if _, seen := out[g.Things[h]]; !seen {
			out[g.Things[h]] = mapgen.Pt{X: i % g.W, Y: i / g.W}
		}
	}
	return out
}

// TestEverythingInspectsAsItself: in every age, the cursor on every kind of
// thing the map shows gets a title of that thing, never the town's name
// alone and never nothing.
func TestEverythingInspectsAsItself(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for ai, age := range config.AgeOrder() {
		o := ageOptions(ai, age)
		o.Wars, o.Routes, o.Idle = 1, 3, 12
		m := modelFor(t, o)
		v := New().(*view)
		f := Frame{Model: m, Anim: 17, Clock: 17}
		for th, p := range thingsOf(v.scene(f).Grid) {
			v.cur = p
			in, ok := v.Inspect(f)
			if !ok || in.Title == "" || in.Title == m.TownName {
				t.Errorf("%s: %+v inspects as %+v", age, th, in)
			}
		}
	}
}

// TestTrafficInspectsAsWhatItIs: a cart, a car, a ship, a drone: the traffic of
// every age has a name of its own and a line, not the road's.
func TestTrafficInspectsAsWhatItIs(t *testing.T) {
	fixture.WithTheme(t, "forge")
	roads := []string{"Street", "Avenue", "Bridge"}
	seen := map[string]bool{}
	for ai, age := range config.AgeOrder() {
		m := modelFor(t, ageOptions(ai, age))
		agents := 0
		for _, anim := range []int{17, 700, 1500, 2300} { // the ships of space come and go with the frame
			v := New().(*view)
			f := Frame{Model: m, Anim: anim, Clock: anim}
			for th, p := range thingsOf(v.scene(f).Grid) {
				if th.Kind != mapgen.KindAgent {
					continue
				}
				agents++
				words, known := movers[th.Key]
				v.cur = p
				in, _ := v.Inspect(f)
				if !known || in.Title != words[0].String() || slices.Contains(roads, in.Title) || len(in.Lines) != 1 || in.Lines[0] == "" {
					t.Errorf("%s: the %s inspects as %+v", age, th.Key, in)
				}
				seen[th.Key] = true
			}
		}
		if agents == 0 && ai < m.Catalog.SpaceAge() { // space has stations that blink, not traffic
			t.Errorf("%s: no traffic to inspect", age)
		}
	}
	for key := range movers {
		if !seen[key] {
			t.Errorf("no age draws a %s", key)
		}
	}
}

// roadKeys are the keys of the roads of the space scenes.
var roadKeys = map[string]bool{"tether": true, "ring": true, "orbit": true, "route": true}

// TestSpaceThingsHaveTheirOwnWords: in the space ages the home planet, the sun,
// the other planets, the stars, the elevator and the routes inspect each as
// itself, and a station or a colony names the part it is.
func TestSpaceThingsHaveTheirOwnWords(t *testing.T) {
	fixture.WithTheme(t, "forge")
	want := map[string]map[string]string{
		"space_age":        {"home": "The homeworld", "space": "Open space"},
		"interstellar_age": {"home": "The homeworld", "sun": "The home sun", "planet": "A planet of the home system", "satellite": "Satellite", "orbit": "The home system", "space": "Deep space"},
		"galactic_age":     {"space": "Between the stars"},
		"quantum_age":      {"space": "A maybe"},
		"transcendent_age": {"space": "Light"},
	}
	for age, titles := range want {
		m := modelFor(t, fixture.Options{Age: age, Seed: 6, Wars: 1, Routes: 3})
		f := Frame{Model: m, Anim: 17, Clock: 17}
		found := map[string]bool{}
		for _, z := range []int{zOverview, zInner, zCloseUp, zDistrict} {
			v := New().(*view)
			v.zoom, v.zoomSet, v.overview = z, true, z <= zInner
			if z == zDistrict {
				v.zoom = zCloseUp
				v.zoomTo(v.scene(f), zDistrict)
			}
			grid, pick := v.scene(f).Grid, func(p mapgen.Pt) mapgen.Pt { return p }
			if z == zOverview || z == zInner { // the orrery is drawn in the overview itself
				if v.ov = v.scene(f).Overview(160, 48); z == zInner {
					v.ov = v.scene(f).OverviewInner(160, 48)
				}
				grid, pick = v.ov.Grid, func(p mapgen.Pt) mapgen.Pt { return v.ov.Pick(p.X, p.Y) }
			}
			for th, p := range thingsOf(grid) {
				v.cur = pick(p)
				in, _ := v.Inspect(f)
				if title, ok := titles[th.Key]; ok && (th.Kind == mapgen.KindRoad) == roadKeys[th.Key] {
					found[th.Key] = true
					if in.Title != title {
						t.Errorf("%s: %s inspects as %q, want %q", age, th.Key, in.Title, title)
					}
				}
				if th.Kind == mapgen.KindBuilding && !th.Ruin {
					if part := placeName(m, th.Lineage); len(in.Lines) == 0 || !strings.HasPrefix(in.Lines[0], part) {
						t.Errorf("%s: a %s station says %v, want its part %q", age, th.Lineage, in.Lines, part)
					}
				}
			}
		}
		for key := range titles {
			if !found[key] {
				t.Errorf("%s: no %s on the map", age, key)
			}
		}
	}
}

// TestSpaceHubIsTheHeartOfTheTown: the square in the space ages is the ring
// station, the home system, the starbase, the anchor and the core.
func TestSpaceHubIsTheHeartOfTheTown(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for i, age := range []string{"space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 6})
		v := New().(*view)
		f := Frame{Model: m}
		if i < 2 {
			v.zoom, v.zoomSet, v.overview = zOverview, true, true // the square is on the overview
		}
		sc := v.scene(f)
		v.cur = sc.Centre
		if age == "galactic_age" || age == "quantum_age" || age == "transcendent_age" {
			v.cur = mapgen.Pt{X: 120, Y: 48} // the home star stays at the core; the camera starts on the town
		}
		if age == "space_age" || age == "interstellar_age" {
			v.cur = mapgen.Pt{X: sc.Centre.X, Y: firstRoadRow(sc.Grid, sc.Centre.X)}
		}
		in, _ := v.Inspect(f)
		if want := hubTitles[i].String() + m.TownName; in.Title != want {
			t.Errorf("%s: the heart of the town inspects as %q, want %q", age, in.Title, want)
		}
	}
}

// firstRoadRow is the row of the square in a ring scene, found by its key.
func firstRoadRow(g *mapgen.Grid, x int) int {
	for y := 0; y < g.H; y++ {
		if th := g.ThingAt(x, y); th.Kind == mapgen.KindRoad && th.Key == "square" {
			return y
		}
	}
	return 0
}

// TestSpaceConstructsInspection verifies that all newly introduced megastructures
// and space stations have rich inspection tooltips with exact title and descriptive lines.
func TestSpaceConstructsInspection(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := &mapmodel.Model{Age: "interstellar_age", TownName: "Alpha Centauri"}
	tests := []struct {
		key      string
		expTitle string
		expLines []string
	}{
		{
			key:      "space_station",
			expTitle: "Orbital Space Station",
			expLines: []string{"A modular habitat and orbital citadel", "docking bays hum with transfer shuttles"},
		},
		{
			key:      "dyson_harvester",
			expTitle: "Solar Dyson Swarm Collector",
			expLines: []string{"Orbits deep in the solar corona", "harvesting raw plasma and radiant stellar energy"},
		},
		{
			key:      "warp_gate",
			expTitle: "Interstellar Warp Gate",
			expLines: []string{"Hub of faster-than-light corridors", "connecting Sol to distant star systems"},
		},
		{
			key:      "asteroid_mine",
			expTitle: "Asteroid Mining Complex",
			expLines: []string{"Tethered refinery and mass driver", "extracting metals and volatiles from ancient chondrites"},
		},
	}

	for _, tc := range tests {
		th := mapgen.Thing{Kind: mapgen.KindBuilding, Key: tc.key}
		in, ok := describeSpace(m, th)
		if !ok {
			t.Fatalf("describeSpace(%s) returned ok=false", tc.key)
		}
		if in.Title != tc.expTitle {
			t.Errorf("key %s: got title %q, want %q", tc.key, in.Title, tc.expTitle)
		}
		if len(in.Lines) != len(tc.expLines) {
			t.Errorf("key %s: got %d lines, want %d", tc.key, len(in.Lines), len(tc.expLines))
		}
		for i := range tc.expLines {
			if i < len(in.Lines) && in.Lines[i] != tc.expLines[i] {
				t.Errorf("key %s line %d: got %q, want %q", tc.key, i, in.Lines[i], tc.expLines[i])
			}
		}
	}
}

// TestCloseUpMoonInspects: the cursor over the Moon in the calm Earth
// close-up names it, with words the map already has.
func TestCloseUpMoonInspects(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "space_age", Seed: 6})
	v := New().(*view)
	f := Frame{Model: m, Anim: 17, Clock: 17}
	for th, p := range thingsOf(v.scene(f).Grid) {
		if th.Kind == mapgen.KindTerrain && th.Key == "satellite" {
			v.cur = p
			if in, _ := v.Inspect(f); in.Title != "Satellite" {
				t.Errorf("the Moon inspects as %q", in.Title)
			}
			return
		}
	}
	t.Error("no Moon in the close-up")
}
