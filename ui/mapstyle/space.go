package mapstyle

import (
	"strconv"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

// space.go says what the cursor says in the space ages: the home planet and
// the sun, the other planets, the stars, the rings, the elevator and the
// routes, the empty sky and the spaceport on the ground, each as itself. The
// five space ages are scenes 0 to 4.

// skyOf is the space scene of the model's age, or -1 on the ground.
func skyOf(m *mapmodel.Model) int {
	for i, age := range []string{"space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"} {
		if m.Age == age {
			return i
		}
	}
	return -1
}

// skyVoids are what open sky is called in each scene, and a line for it.
var skyVoids = [5][2]vibe.Entry{
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.01", "Open space"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.02", "low orbit, quiet tonight")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.03", "Deep space"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.04", "light-years of nothing much")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.05", "Between the stars"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.06", "the lanes run elsewhere")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.07", "A maybe"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.08", "nothing is certain here")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.09", "Light"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.10", "everything you built, remembered")},
}

// hubTitles are what the heart of the town is called in each scene; the town's
// name follows.
var hubTitles = [5]vibe.Entry{
	vibe.E("ui.mapstyle.roguelike.space_inspect.describeHub.01", "Ring station of "),
	vibe.E("ui.mapstyle.roguelike.space_inspect.describeHub.02", "Home system of "),
	vibe.E("ui.mapstyle.roguelike.space_inspect.describeHub.03", "Starbase "),
	vibe.E("ui.mapstyle.roguelike.space_inspect.describeHub.04", "Anchor of "),
	vibe.E("ui.mapstyle.roguelike.space_inspect.describeHub.05", "Core of "),
}

// describeSpace is a thing of the space scenes, or a spaceport on the ground,
// under the cursor; ok is false for a thing it has no words for.
func describeSpace(m *mapmodel.Model, th mapgen.Thing) (in Inspection, ok bool) {
	sky := skyOf(m)
	town := m.TownName
	switch th.Key {
	case "port":
		return Inspection{Title: vibe.Line("ui.mapstyle.port.01", "Spaceport"), Lines: []string{vibe.Line("ui.mapstyle.port.02", "where the shuttles lift off")}}, true
	case "space_station":
		return Inspection{
			Title: vibe.Line("ui.mapstyle.space.station.01", "Orbital Space Station"),
			Lines: []string{
				vibe.Line("ui.mapstyle.space.station.02", "A modular habitat and orbital citadel"),
				vibe.Line("ui.mapstyle.space.station.03", "docking bays hum with transfer shuttles"),
			},
		}, true
	case "dyson_harvester":
		return Inspection{
			Title: vibe.Line("ui.mapstyle.space.dyson.01", "Solar Dyson Swarm Collector"),
			Lines: []string{
				vibe.Line("ui.mapstyle.space.dyson.02", "Orbits deep in the solar corona"),
				vibe.Line("ui.mapstyle.space.dyson.03", "harvesting raw plasma and radiant stellar energy"),
			},
		}, true
	case "warp_gate":
		return Inspection{
			Title: vibe.Line("ui.mapstyle.space.warp_gate.01", "Interstellar Warp Gate"),
			Lines: []string{
				vibe.Line("ui.mapstyle.space.warp_gate.02", "Hub of faster-than-light corridors"),
				vibe.Line("ui.mapstyle.space.warp_gate.03", "connecting Sol to distant star systems"),
			},
		}, true
	case "asteroid_mine":
		return Inspection{
			Title: vibe.Line("ui.mapstyle.space.asteroid_mine.01", "Asteroid Mining Complex"),
			Lines: []string{
				vibe.Line("ui.mapstyle.space.asteroid_mine.02", "Tethered refinery and mass driver"),
				vibe.Line("ui.mapstyle.space.asteroid_mine.03", "extracting metals and volatiles from ancient chondrites"),
			},
		}, true
	case "comet":
		return Inspection{
			Title: vibe.Line("ui.mapstyle.space.comet.01", "Great Comet"),
			Lines: []string{
				vibe.Line("ui.mapstyle.space.comet.02", "An icy wanderer sweeping from the Oort cloud"),
				vibe.Line("ui.mapstyle.space.comet.03", "its luminescent ion tail points away from the Sun"),
			},
		}, true
	case "pluto":
		return Inspection{
			Title: vibe.Line("ui.mapstyle.space.pluto.01", "Pluto & Charon"),
			Lines: []string{
				vibe.Line("ui.mapstyle.space.pluto.02", "Binary dwarf planet in the frozen Kuiper belt"),
				vibe.Line("ui.mapstyle.space.pluto.03", "nitrogen glaciers and methane frost glint under distant Sol"),
			},
		}, true
	}
	if sky < 0 {
		return Inspection{}, false
	}
	switch {
	case th.Kind == mapgen.KindRoad && th.Key == "square":
		in = describeSquare(m, town)
		in.Title = hubTitles[sky].String() + town
		return in, true
	case th.Kind == mapgen.KindRoad:
		return describeLane(m, th.Key, town), true
	}
	switch th.Key {
	case "home":
		if sky >= 2 { // on the star map's galaxy the home world is a star: the heart of the town
			in = describeSquare(m, town)
			in.Title = hubTitles[sky].String() + town
			return in, true
		}
		return describeHomeworld(m), true
	case "neighbor":
		return describeNearStar(m, th), true
	case "deepsky":
		return describeSkyObject(th), true
	case "planet":
		if th.Ref >= 1 && th.Ref <= len(planetNames) {
			return Inspection{Title: planetNames[th.Ref-1].String()}, true
		}
		return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.08", "A planet of the home system"),
			Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.09", "it keeps its old orbit")}}, true
	case "sun":
		in = Inspection{Title: vibe.Line("ui.mapstyle.space.sun.01", "Sol"),
			Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.32", "the orrery of the home system"),
				town + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.33", " still turns round it")}}
		if sky >= 2 { // on the star map Sol is the town's seat
			in.Command = m.SquareCommand()
		}
		return in, true
	case "satellite":
		return Inspection{Title: vibe.Line("ui.mapstyle.space.satellite.01", "Satellite")}, true
	case "the_moon":
		return Inspection{Title: vibe.Line("ui.mapstyle.space.moon.01", "The Moon")}, true
	case "star":
		return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.50", "A star"), Lines: []string{skyVoids[sky][1].String()}}, true
	case "space":
		return Inspection{Title: skyVoids[sky][0].String(), Lines: []string{skyVoids[sky][1].String()}}, true
	}
	return Inspection{}, false
}

// describeHomeworld is the planet the town stands on, seen from orbit, as a
// planet of the solar system, or as the home star.
func describeHomeworld(m *mapmodel.Model) Inspection {
	return Inspection{Title: planetNames[2].String(), Command: m.SquareCommand(),
		Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.04", "where it all began"),
			plural(m.Workers.Pop, vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.05", "person"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.06", "people")) +
				vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.07", " in all")}}
}

// describeLane is a road of the space scenes: the elevator's tether, a ring
// of stations, the orbits of the solar system, a route between stars.
func describeLane(m *mapmodel.Model, key, town string) Inspection {
	switch key {
	case "tether":
		return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.14", "Space elevator"),
			Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.15", "the tether from ") + town + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.16", " to the ring"),
				vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.17", "climbers ride it day and night")}}
	case "orbit":
		return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.34", "The home system"),
			Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.35", "a world on every ring")}}
	case "route":
		return Inspection{Title: vibe.Line("ui.mapstyle.space.route.01", "Star route"),
			Lines: []string{vibe.Line("ui.mapstyle.space.route.02", "joins the systems your people have reached")}}
	}
	return Inspection{Title: vibe.Line("ui.mapstyle.space.ring.01", "Orbital ring"),
		Lines: []string{vibe.Line("ui.mapstyle.space.ring.02", "stations stand along it")}}
}

// placeName is what a building line is where the cursor finds it: a part of
// the space scene (a habitat module, a colony ark), else the line's own name.
func placeName(m *mapmodel.Model, lineage string) string {
	if sky := skyOf(m); sky >= 0 && sky < len(skyParts) {
		if part, ok := skyParts[sky][lineage]; ok {
			return part.String()
		}
	}
	return lineageLabel(lineage)
}

// planetNames are Sol's planets by Ref (mapgen.SystemSpec.ThingOf: index+1),
// inner to outer. Earth is the "home" thing; describeHomeworld takes its slot.
var planetNames = [...]vibe.Entry{
	vibe.E("ui.mapstyle.space.planet.mercury", "Mercury"),
	vibe.E("ui.mapstyle.space.planet.venus", "Venus"),
	vibe.E("ui.mapstyle.space.planet.earth", "Earth"),
	vibe.E("ui.mapstyle.space.planet.mars", "Mars"),
	vibe.E("ui.mapstyle.space.planet.jupiter", "Jupiter"),
	vibe.E("ui.mapstyle.space.planet.saturn", "Saturn"),
	vibe.E("ui.mapstyle.space.planet.uranus", "Uranus"),
	vibe.E("ui.mapstyle.space.planet.neptune", "Neptune"),
}

// starNames are the stars of Sol's neighbourhood, in the order of
// mapgen.NearStars.
var starNames = [len(mapgen.NearStars)]vibe.Entry{
	vibe.E("ui.mapstyle.space.star.alpha_centauri", "Alpha Centauri"),
	vibe.E("ui.mapstyle.space.star.barnards_star", "Barnard's Star"),
	vibe.E("ui.mapstyle.space.star.wolf_359", "Wolf 359"),
	vibe.E("ui.mapstyle.space.star.lalande_21185", "Lalande 21185"),
	vibe.E("ui.mapstyle.space.star.sirius", "Sirius"),
	vibe.E("ui.mapstyle.space.star.luyten_726_8", "Luyten 726-8"),
	vibe.E("ui.mapstyle.space.star.ross_154", "Ross 154"),
	vibe.E("ui.mapstyle.space.star.ross_248", "Ross 248"),
	vibe.E("ui.mapstyle.space.star.epsilon_eridani", "Epsilon Eridani"),
	vibe.E("ui.mapstyle.space.star.lacaille_9352", "Lacaille 9352"),
	vibe.E("ui.mapstyle.space.star.ross_128", "Ross 128"),
	vibe.E("ui.mapstyle.space.star.61_cygni", "61 Cygni"),
	vibe.E("ui.mapstyle.space.star.procyon", "Procyon"),
	vibe.E("ui.mapstyle.space.star.struve_2398", "Struve 2398"),
	vibe.E("ui.mapstyle.space.star.groombridge_34", "Groombridge 34"),
	vibe.E("ui.mapstyle.space.star.epsilon_indi", "Epsilon Indi"),
	vibe.E("ui.mapstyle.space.star.tau_ceti", "Tau Ceti"),
	vibe.E("ui.mapstyle.space.star.teegardens_star", "Teegarden's Star"),
	vibe.E("ui.mapstyle.space.star.kapteyns_star", "Kapteyn's Star"),
}

// describeNearStar is a star of the neighbourhood: its name, how far it is
// from Sol and, if people live there, what the colony says of itself.
func describeNearStar(m *mapmodel.Model, th mapgen.Thing) Inspection {
	i := th.Ref - 1
	if i < 0 || i >= len(starNames) {
		return Inspection{}
	}
	tenths := (mapgen.NearStars[i].Dist100 + 5) / 10
	in := Inspection{Title: starNames[i].String(),
		Lines: []string{strconv.Itoa(tenths/10) + "." + strconv.Itoa(tenths%10) + vibe.Line("ui.mapstyle.space.star.distance", " light-years from Sol")}}
	if th.Lineage != "" {
		colony := describeSector(m, mapgen.Thing{Kind: mapgen.KindSector, Key: th.Lineage, Lineage: th.Lineage, Count: th.Count})
		in.Lines = append(in.Lines, colony.Title)
		in.Lines = append(in.Lines, colony.Lines...)
	}
	return in
}

// skyNames are the objects of the wide views, in the order of mapgen.SkyObjects.
var skyNames = [len(mapgen.SkyObjects)]vibe.Entry{
	vibe.E("ui.mapstyle.space.sky.orion_nebula", "Orion Nebula"),
	vibe.E("ui.mapstyle.space.sky.horsehead_nebula", "Horsehead Nebula"),
	vibe.E("ui.mapstyle.space.sky.pleiades", "The Pleiades"),
	vibe.E("ui.mapstyle.space.sky.hyades", "The Hyades"),
	vibe.E("ui.mapstyle.space.sky.crab_nebula", "Crab Nebula"),
	vibe.E("ui.mapstyle.space.sky.eagle_nebula", "Eagle Nebula"),
	vibe.E("ui.mapstyle.space.sky.lagoon_nebula", "Lagoon Nebula"),
	vibe.E("ui.mapstyle.space.sky.carina_nebula", "Carina Nebula"),
	vibe.E("ui.mapstyle.space.sky.milky_way", "The Milky Way"),
	vibe.E("ui.mapstyle.space.sky.large_magellanic_cloud", "Large Magellanic Cloud"),
	vibe.E("ui.mapstyle.space.sky.small_magellanic_cloud", "Small Magellanic Cloud"),
	vibe.E("ui.mapstyle.space.sky.andromeda", "Andromeda"),
	vibe.E("ui.mapstyle.space.sky.triangulum", "Triangulum"),
}

// describeSkyObject names an object of the wide views and says how far it is
// from Sol; the Milky Way has no distance, Sol is in it.
func describeSkyObject(th mapgen.Thing) Inspection {
	i := th.Ref - 1
	if i < 0 || i >= len(skyNames) {
		return Inspection{}
	}
	in := Inspection{Title: skyNames[i].String()}
	ly := mapgen.SkyObjects[i].Ly
	switch {
	case ly == 0:
		in.Lines = []string{vibe.Line("ui.mapstyle.space.sky.home", "Sol lies in one of its arms")}
	case ly >= 1000000:
		t := (ly + 50000) / 100000
		in.Lines = []string{strconv.Itoa(t/10) + "." + strconv.Itoa(t%10) + vibe.Line("ui.mapstyle.space.sky.million", " million light-years from Sol")}
	default:
		s := strconv.Itoa(ly)
		if len(s) > 3 {
			s = s[:len(s)-3] + "," + s[len(s)-3:]
		}
		in.Lines = []string{s + vibe.Line("ui.mapstyle.space.sky.distance", " light-years from Sol")}
	}
	return in
}
