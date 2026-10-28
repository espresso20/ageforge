package mapstyle

import (
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
		if sky >= 2 { // on the star map the home world is a star: the heart of the town
			in = describeSquare(m, town)
			in.Title = hubTitles[sky].String() + town
			return in, true
		}
		return describeHomeworld(m), true
	case "planet":
		return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.08", "A planet of the home system"),
			Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.09", "it keeps its old orbit")}}, true
	case "sun":
		return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.31", "The home sun"),
			Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.32", "the orrery of the home system"),
				town + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.33", " still turns round it")}}, true
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
	return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.03", "The homeworld"), Command: m.SquareCommand(),
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
