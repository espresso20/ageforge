package world

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// Moon is a named major moon: where it circles, its true orbit radius in
// radii of its planet, and the art size of its disc (a true size would leave
// most of them under a pixel).
type Moon struct {
	Planet string
	Key    string
	Orbit  float64
	Size   float64
	Tone   Tone
}

// Moons are the major moons of Sol's planets and Pluto, inner to outer. The
// Moon itself is the home frame's. Names are the keys the ui writes words for.
var Moons = []Moon{
	{"mars", "phobos", 2.76, .1, ToneRock},
	{"mars", "deimos", 6.9, .1, ToneRock},
	{"jupiter", "io", 5.9, .11, ToneSand},
	{"jupiter", "europa", 9.4, .1, ToneWhite},
	{"jupiter", "ganymede", 15, .14, ToneRock},
	{"jupiter", "callisto", 26.3, .13, ToneRock},
	{"saturn", "enceladus", 3.95, .1, ToneWhite},
	{"saturn", "rhea", 8.74, .1, ToneWhite},
	{"saturn", "titan", 20.3, .14, ToneAmber},
	{"saturn", "iapetus", 59.1, .1, ToneRock},
	{"uranus", "miranda", 5.07, .1, ToneCool},
	{"uranus", "titania", 17.1, .12, ToneRock},
	{"uranus", "oberon", 22.8, .12, ToneRock},
	{"neptune", "triton", 14.3, .16, ToneCool},
	{"pluto", "charon", 16.5, .5, ToneRock},
}

// MoonletCap is the most unnamed moonlets a planet draws.
const MoonletCap = 12

// planetFrames is the close-up of every planet of the system, the home world's
// first (the chain follows the first child), then the rest inner to outer and
// Pluto last.
func planetFrames(m *mapmodel.Model, seed int64) []*Frame {
	out := []*Frame{homeFrame(m, seed)}
	for i := range SolSpec.Planets {
		if i != Home {
			out = append(out, worldFrame(seed, SolNames[i], SolSpec.Planets[i], planetKind(SolSpec.Planets[i]), SolSpec.Planets[i].Base.Tone, SolEarths[i], i))
		}
	}
	pluto := PlanetSpec{Close: 20, Moons: 1}
	return append(out, worldFrame(seed, "pluto", pluto, "dwarf", ToneRock, .19, len(SolSpec.Planets)))
}

// worldFrame is a planet that is not the home world, in its own radii: the
// planet, its named moons, and seeded unnamed moonlets for the rest of its
// count. It is drawn inflated so the planet is as big at the handoff as the
// system draws it (the spec's Close, against the home world's).
func worldFrame(seed int64, key string, sp PlanetSpec, kind string, tone Tone, earths float64, idx int) *Frame {
	f := &Frame{Kind: FramePlanet, Key: key, Unit: UnitPlanetRadii, Anchor: key, Radius: 64,
		Band: Band{Min: 4.5, Max: 28}, Inflate: HomeInflate * 100 / float64(sp.Close)}
	f.Objects = append(f.Objects, Object{Kind: KindPlanet, Key: key, Label: key, Inspect: key, Size: 1, Tone: tone, Weight: 255, Pick: true,
		Body: &Body{Spec: sp, Phase: OrbitPhase(seed, idx), Rings: sp.Ring, Moons: sp.Moons, Kind: kind, Earths: earths}})
	named := 0
	for _, mo := range Moons {
		if mo.Planet != key {
			continue
		}
		named++
		ph := Roll(Turn, seed, 5230, int64(idx), int64(named))
		f.Objects = append(f.Objects, Object{Kind: KindMoon, Key: mo.Key, Label: mo.Key, Inspect: mo.Key, Pos: polar(mo.Orbit, ph),
			Size: mo.Size, Tone: mo.Tone, Weight: 255, Pick: true,
			Body: &Body{Parent: key, Orbit: mo.Orbit, Period: int(256 * 10 * math.Sqrt(mo.Orbit)), Phase: ph, Kind: "moon"}})
	}
	for k := 0; k < min(MoonletCap, sp.Moons-named); k++ {
		ph := Roll(Turn, seed, 5231, int64(idx), int64(k))
		orbit := 2.6 + float64(Roll(400, seed, 5232, int64(idx), int64(k)))/10
		f.Objects = append(f.Objects, Object{Kind: KindMoon, Key: "moonlet_" + key + "_" + itoa(k), Pos: polar(orbit, ph),
			Size: .07, Tone: ToneRock, Weight: 200, Body: &Body{Parent: key, Kind: "moon", Phase: ph}})
	}
	return f
}
