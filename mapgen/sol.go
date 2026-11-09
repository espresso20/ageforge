package mapgen

import "github.com/espresso20/ageforge/world"

// Sol. The renderers draw the solar system from SolSpec, a literal: eight
// planets with their real moon counts, the belt, the dwarf planet and a comet.

// StarSpec is a star as the painters read it (hues, not sky roles).
type StarSpec struct {
	Core, Mid, Rim Hue
	Size           int
}

// PlanetSpec is one planet as the painters read it; world.PlanetSpec holds
// the data and its field notes.
type PlanetSpec struct {
	Size     int
	Close    int
	Orbit    int
	Moons    int
	Base     Paint
	Bands    []Paint
	Blotch   Paint
	Cap      bool
	Spot     bool
	Ring     bool
	World    bool
	Home     bool
	Outpost  int
	Stations int
}

// SystemSpec is a whole system.
type SystemSpec struct {
	Star    StarSpec
	Planets []PlanetSpec // inner to outer
	Belts   []int        // orbits of the asteroid belts
	Dwarf   bool         // a dwarf planet far out
	Comet   bool
}

func specOf(w world.SystemSpec) SystemSpec {
	out := SystemSpec{Star: StarSpec{hueOf(w.Star.Core), hueOf(w.Star.Mid), hueOf(w.Star.Rim), w.Star.Size}, Belts: w.Belts, Dwarf: w.Dwarf, Comet: w.Comet}
	for _, p := range w.Planets {
		q := PlanetSpec{Size: p.Size, Close: p.Close, Orbit: p.Orbit, Moons: p.Moons, Base: paintOf(p.Base), Blotch: paintOf(p.Blotch),
			Cap: p.Cap, Spot: p.Spot, Ring: p.Ring, World: p.World, Home: p.Home, Outpost: p.Outpost, Stations: p.Stations}
		for _, b := range p.Bands {
			q.Bands = append(q.Bands, paintOf(b))
		}
		out.Planets = append(out.Planets, q)
	}
	return out
}

// ThingOf is the thing planet i is drawn as: the home world is "home", the
// others "planet" numbered by Ref from 1.
func (sp *SystemSpec) ThingOf(i int) Thing {
	if sp.Planets[i].Home {
		return Thing{Kind: KindTerrain, Key: "home"}
	}
	return Thing{Kind: KindTerrain, Key: "planet", Ref: i + 1}
}

// The sun is a terrain thing of its own; its halo is the same thing with a
// Ref, which is not a body.
var (
	sunThing     = Thing{Kind: KindTerrain, Key: "sun"}
	sunHaloThing = Thing{Kind: KindTerrain, Key: "sun", Ref: 1}
)

// SolSpec is the home system, from world/.
var SolSpec = specOf(world.SolSpec)
