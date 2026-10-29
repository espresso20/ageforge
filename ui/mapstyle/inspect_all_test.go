package mapstyle

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapgen"
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

// starTraffic are the movers of the star map, which has none drawn since its
// lanes became line work.
var starTraffic = map[string]bool{"starship": true, "phase_ship": true, "mote": true}

// TestTrafficInspectsAsWhatItIs: a cart, a car, a ship, a drone: the traffic of
// every age has a name of its own and a line, not the road's.
func TestTrafficInspectsAsWhatItIs(t *testing.T) {
	fixture.WithTheme(t, "forge")
	roads := []string{"Street", "Avenue", "Bridge"}
	seen := map[string]bool{}
	for ai, age := range config.AgeOrder() {
		m := modelFor(t, ageOptions(ai, age))
		v := New().(*view)
		f := Frame{Model: m, Anim: 17, Clock: 17}
		agents := 0
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
		if agents == 0 && age != "galactic_age" && age != "quantum_age" && age != "transcendent_age" { // the star map has lanes, not traffic
			t.Errorf("%s: no traffic to inspect", age)
		}
	}
	for key := range movers {
		if starTraffic[key] {
			continue
		}
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
		"space_age":        {"home": "The homeworld", "tether": "Space elevator", "ring": "Orbital ring", "space": "Open space"},
		"interstellar_age": {"home": "The homeworld", "sun": "The home sun", "planet": "A planet of the home system", "orbit": "The home system", "space": "Deep space"},
		"galactic_age":     {"space": "Between the stars"},
		"quantum_age":      {"space": "A maybe"},
		"transcendent_age": {"space": "Light"},
	}
	for age, titles := range want {
		m := modelFor(t, fixture.Options{Age: age, Seed: 6, Wars: 1, Routes: 3})
		v := New().(*view)
		f := Frame{Model: m, Anim: 17, Clock: 17}
		found := map[string]bool{}
		for th, p := range thingsOf(v.scene(f).Grid) {
			v.cur = p
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
		sc := v.scene(f)
		v.cur = sc.Centre
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
