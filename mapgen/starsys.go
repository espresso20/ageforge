package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Star systems. A system is a SystemSpec, and the renderers draw nothing but
// specs: the home system is SolSpec, a literal; every other star of the sky
// is GenSystem(seed, star, progress), drawn from the save's seed by stable
// rolls, so the same save and the same star always give the same system.
// Star 0 is the home sun; the others count up from 1.

// StarSpec is a star: three hues from the heart out, and its radius at the
// full zoom in 16ths of a pixel.
type StarSpec struct {
	Core, Mid, Rim Hue
	Size           int
}

// PlanetSpec is one planet. Sizes are in 16ths of a pixel at the full zoom.
type PlanetSpec struct {
	Size     int   // radius
	Close    int   // radius of its close-up disc, percent of the home world's
	Orbit    int   // distance from the star, percent of the system's reach
	Moons    int   // how many
	Base     Paint // the face, without bands
	Bands    []Paint
	Blotch   Paint // weight 0: none; else the odd fleck on the face
	Cap      bool  // a polar cap
	Spot     bool  // a great storm
	Ring     bool
	World    bool // seas, land and cloud
	Home     bool // the home world: the town's, with its air
	Outpost  int  // a station by it in the squeezed views: 0 none, 1 a hull pixel, 2 a glyph
	Stations int  // stations in its close-up
}

// SystemSpec is a whole system.
type SystemSpec struct {
	Star    StarSpec
	Planets []PlanetSpec // inner to outer
	Belts   []int        // orbits of the asteroid belts
	Dwarf   bool         // a dwarf planet far out
	Comet   bool
	Home    bool // the home system: it has the town's works
}

// StationCount is how many planets have a station by them.
func (sp SystemSpec) StationCount() (n int) {
	for _, p := range sp.Planets {
		if p.Outpost > 0 {
			n++
		}
	}
	return n
}

// ThingOf is the thing planet i is drawn as: the home world is "home", the
// others "planet" numbered by Ref from 1.
func (sp *SystemSpec) ThingOf(i int) Thing {
	if sp.Planets[i].Home {
		return Thing{Kind: KindTerrain, Key: "home"}
	}
	return Thing{Kind: KindTerrain, Key: "planet", Ref: i + 1}
}

// StarThing is the thing star k is drawn as: the home sun, or "star"
// numbered by Ref. The halo is its glow, which is not a body of its own.
func StarThing(k int, halo bool) Thing {
	th := Thing{Kind: KindTerrain, Key: "sun"}
	if k > 0 {
		th = Thing{Kind: KindTerrain, Key: "star", Ref: k}
	}
	if halo {
		if k == 0 {
			th.Ref = 1
		} else {
			th.Count = 1
		}
	}
	return th
}

// SolSpec is the home system as the sky has it: eight planets with their real
// moon counts, Saturn ringed, a station by the home world and one at Mars.
var SolSpec = SystemSpec{
	Star: StarSpec{Core: HWhite, Mid: HSun, Rim: HCore, Size: 88},
	Planets: []PlanetSpec{
		{Size: 22, Close: 55, Orbit: 18, Base: P(HRock, 255)},
		{Size: 32, Close: 90, Orbit: 28, Bands: []Paint{P(HSun, 250), P(HSun, 205), P(HSun, 250), P(HSun, 205), P(HSun, 250), P(HSun, 205)}},
		{Size: 35, Close: 100, Orbit: 39, Moons: 1, World: true, Home: true, Outpost: 1, Stations: 2},
		{Size: 27, Close: 70, Orbit: 50, Moons: 2, Base: P(HRed, 255), Blotch: P(HRed, 200), Cap: true, Outpost: 2},
		{Size: 68, Close: 110, Orbit: 68, Moons: 4, Spot: true, Bands: []Paint{P(HRock, 235), P(HAmber, 255), P(HSand, 255), P(HRock, 215), P(HAmber, 245), P(HSand, 240), P(HRock, 225), P(HAmber, 255), P(HRock, 235)}},
		{Size: 58, Close: 85, Orbit: 79, Moons: 5, Ring: true, Bands: []Paint{P(HSand, 235), P(HGold, 250), P(HSand, 255), P(HGold, 235), P(HSand, 250), P(HGold, 255), P(HSand, 240), P(HGold, 240), P(HSand, 230)}},
		{Size: 42, Close: 90, Orbit: 89, Moons: 4, Base: P(HCool, 255)},
		{Size: 42, Close: 90, Orbit: 98, Moons: 2, Bands: []Paint{P(HPanel, 255), P(HCyan, 215), P(HPanel, 255), P(HPanel, 255), P(HPanel, 255), P(HCyan, 215), P(HPanel, 255), P(HPanel, 255), P(HPanel, 255)}},
	},
	Belts: []int{59},
	Dwarf: true,
	Comet: true,
	Home:  true,
}

// starTypes are the kinds of star, with the weight each is dealt at: the
// common ones are small and dim.
var starTypes = [...]struct {
	StarSpec
	weight int
}{
	{StarSpec{HCore, HRed, HRed, 56}, 45},      // red dwarf
	{StarSpec{HWhite, HSun, HCore, 88}, 30},    // yellow
	{StarSpec{HWhite, HWhite, HCool, 104}, 17}, // white
	{StarSpec{HWhite, HCyan, HPanel, 150}, 8},  // blue giant
}

// giantBands are the looks of a big planet: the bands from pole to pole.
var giantBands = [...][]Paint{
	{P(HSand, 235), P(HGold, 250), P(HSand, 255), P(HGold, 235), P(HSand, 250), P(HGold, 255)},
	{P(HRock, 235), P(HAmber, 255), P(HSand, 255), P(HRock, 215), P(HAmber, 245), P(HSand, 240)},
	{P(HPanel, 255), P(HCyan, 215), P(HPanel, 255), P(HPanel, 255), P(HCyan, 215), P(HPanel, 255)},
	{P(HCool, 255), P(HCool, 215), P(HCool, 255), P(HCool, 215)},
	{P(HRed, 255), P(HRock, 230), P(HRed, 220), P(HRock, 255), P(HRed, 240), P(HRock, 215)},
	{P(HHydro, 235), P(HPanel, 255), P(HHydro, 215), P(HPanel, 240), P(HHydro, 255), P(HPanel, 220)},
}

// rockyBases are the colours of a small planet.
var rockyBases = [...]Hue{HRock, HRed, HSand, HSun, HSteel, HAmber, HRock, HCool}

// genStar is the star k of the save.
func genStar(seed int64, k int) StarSpec {
	total := 0
	for _, t := range starTypes {
		total += t.weight
	}
	r := roll(total, seed, 7001, int64(k))
	for _, t := range starTypes {
		if r < t.weight {
			t.Size += roll(13, seed, 7002, int64(k)) - 6 // no two alike
			return t.StarSpec
		}
		r -= t.weight
	}
	return starTypes[1].StarSpec
}

// maxStations caps the stations of a generated system.
const maxStations = 6

// GenSystem is the system of star k (1 and up) of the save. progress is how
// far the player has come into space: 0 before the interstellar ages, when no
// system but the home one has a station, then the number of building types of
// those ages the town owns. A system gets one station at first and one more
// for every two after.
func GenSystem(seed int64, k, progress int) SystemSpec {
	r := func(n int, salt int64, more ...int64) int {
		return roll(n, append([]int64{seed, 7100 + salt, int64(k)}, more...)...)
	}
	sp := SystemSpec{Star: genStar(seed, k), Comet: r(3, 0) == 0}
	// one to ten planets, most systems having three to eight
	n := [...]int{1, 2, 2, 3, 3, 4, 4, 4, 5, 5, 5, 6, 6, 6, 7, 7, 8, 8, 9, 10}[r(20, 1)]
	rocky := (n*55 + 99) / 100
	// belts: none, one after the small planets, or that and one at the far end
	belts := [...]int{0, 0, 1, 1, 2}[r(5, 3)]
	if n < 3 {
		belts = min(belts, 1)
	}
	// orbits come from random gaps, spread from clear of the star to the edge
	slots := n + belts
	gap, total := make([]int, slots), 0
	for i := range gap {
		gap[i] = 6 + r(7, 4, int64(i))
		total += gap[i]
	}
	lo := sp.Star.Size / 5
	at, slot := 0, 0
	next := func() int {
		at += gap[slot]
		slot++
		return lo + (98-lo)*at/total
	}
	for i := 0; i < n; i++ {
		if belts > 0 && i == rocky && rocky < n {
			sp.Belts = append(sp.Belts, next())
		}
		p := PlanetSpec{Orbit: next()}
		id := int64(i)
		if i < rocky {
			p.Size = 18 + r(18, 5, id)
			p.Close = 55 + r(45, 6, id)
			p.Base = P(rockyBases[r(len(rockyBases), 8, id)], 255)
			switch r(6, 9, id) {
			case 0:
				p.Cap = true
			case 1:
				p.Blotch, p.Cap = P(p.Base.Hue, 200), true
			case 2:
				p.Bands = []Paint{p.Base, P(p.Base.Hue, 205), p.Base, P(p.Base.Hue, 205), p.Base, P(p.Base.Hue, 205)}
			case 3:
				p.World = r(2, 10, id) == 0
			}
		} else {
			p.Size = 42 + r(30, 5, id)
			p.Close = 85 + r(30, 6, id)
			p.Bands = giantBands[r(len(giantBands), 8, id)]
			p.Ring = r(3, 11, id) == 0
			p.Spot = r(4, 12, id) == 0
		}
		p.Moons = r(1+p.Size*p.Size/260, 7, id) // 0 to about 20, by size
		sp.Planets = append(sp.Planets, p)
	}
	for len(sp.Belts) < belts { // the far belt, past the last planet
		sp.Belts = append(sp.Belts, next())
	}
	// stations: none until the interstellar ages
	if progress > 0 {
		for j := min(n, maxStations, 1+(progress-1)/2); j > 0; j-- {
			p := &sp.Planets[(r(n, 13)+j)%n]
			p.Outpost, p.Stations = 1+r(2, 14, int64(j)), 1
		}
	}
	return sp
}

// systemOf is the spec of star k: the home system, or a generated one.
func systemOf(seed int64, k, progress int) *SystemSpec {
	if k == 0 {
		return &SolSpec
	}
	sp := GenSystem(seed, k, progress)
	return &sp
}

// starSites places the stars round the home system in the cosmos view, in
// thousandths of the way from its middle to the edge of the window. It does
// not depend on the window, so star k is the same star at every size.
func starSites(seed int64, most int) (out [][2]int) {
next:
	for a := int64(0); a < 400 && len(out) < most; a++ {
		x, y := roll(1801, seed, 7200, a)-900, roll(1701, seed, 7201, a)-850
		if abs(x) < 400 && abs(y) < 550 { // the home system's room
			continue
		}
		for _, o := range out {
			if abs(x-o[0]) < 140 && abs(y-o[1]) < 300 {
				continue next
			}
		}
		out = append(out, [2]int{x, y})
	}
	return out
}

// spaceProgress is how far the town has come into the interstellar ages: the
// number of building types of those ages it owns, 0 before them.
func spaceProgress(m *mapmodel.Model) (n int) {
	from, ok := m.Catalog.AgeIdx["interstellar_age"]
	if !ok {
		return 0
	}
	for _, b := range m.Buildings {
		if b.Count > 0 && b.Age >= from {
			n++
		}
	}
	return n
}
