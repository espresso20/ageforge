package mapmodel

import "github.com/espresso20/ageforge/config"

// movers.go is the traffic roster both map styles read: every kind of thing
// that moves through a town, the age that introduces it and the last age it
// is seen in, how it travels and how fast. An age shows only movers that it,
// or an earlier age, introduced, so the map never shows a vehicle from an age
// the player has not reached (the no-spoilers rule). Older movers retire as
// their era passes, which keeps the streets readable.

// Mover is one kind of moving thing.
type Mover uint8

const (
	MoverNone Mover = iota
	MoverWalker
	MoverHunter
	MoverOxCart
	MoverRider
	MoverRowboat
	MoverWagon
	MoverSailShip
	MoverSteamTrain
	MoverTram
	MoverSteamship
	MoverEarlyCar
	MoverCar
	MoverTruck
	MoverPlane
	MoverBoxShip
	MoverTrain
	MoverMaglev
	MoverDrone
	MoverHovercar
	MoverShuttle
	MoverSatellite
	MoverHabitat
	// The Earth arc (glyphs_city.go): the news helicopter, the sky train,
	// the megacity's crowds and the space elevator's climbers.
	MoverNewsHeli
	MoverSkyTrain
	MoverCrowd
	MoverClimber
	NumMovers
)

// Way is what a mover travels on.
type Way uint8

const (
	WayFoot   Way = iota // paths between homes, works and the wilds
	WayStreet            // streets, avenues and trails
	WayRail              // the railway
	WayWater             // rivers, lakes and the sea
	WaySky               // overhead
)

// MoverInfo describes one kind of mover.
type MoverInfo struct {
	Key   string // "steam_train"
	Name  string // legend label: "steam train"
	Title string // inspect title: "Steam train"
	Sym   Sym
	Class Class
	Way   Way
	// From is the index of the age that introduces it; Until the last age
	// it appears in (-1: to the end of time).
	From, Until int
	// Pace is animation frames per cell: people stroll at 5 (about 1.6
	// cells a second), a maglev takes 1.
	Pace int
	// Cars is how many cells long it is (trains); 0 or 1 is one cell.
	Cars  int
	Smoke bool // trails smoke (steam)
	// Night: before the Cosmic Era it shows only at night (satellites).
	Night bool
	// Lines are what inspecting one says; each mover keeps one of them.
	Lines []string
}

// In reports whether the mover is about in the age with index age.
func (i MoverInfo) In(age int) bool {
	return i.From >= 0 && age >= i.From && (i.Until < 0 || age <= i.Until)
}

// Line is the inspect line for mover number n of this kind.
func (i MoverInfo) Line(n int) string {
	if len(i.Lines) == 0 {
		return ""
	}
	return i.Lines[Hash(int64(n), HashStr(i.Key))%uint64(len(i.Lines))]
}

type moverDef struct {
	key, name, title string
	sym              Sym
	class            Class
	way              Way
	from, until      string // age keys; until "" runs to the end
	pace, cars       int
	smoke, night     bool
	lines            []string
}

// moverDefs is the roster. Every era brings its own: Stone walkers and
// hunters; Iron ox carts, riders and rowboats; Steel wagons, sailing ships
// and, from the Industrial Age, the railway and its steam trains; Electric
// trams, steamships and early cars; Digital cars, trucks, planes, container
// ships and freight trains, news helicopters from the Information Age and
// drones from the Digital Age; Neon maglevs and hovercars, with the
// megacity's sky trains and crowds, the space elevator's climbers from the
// Fusion Age and shuttles from the Space Age's launch pads; Cosmic orbital
// habitats. Satellites cross the night sky from the Modern Age, as the
// skyline has always drawn them.
var moverDefs = [NumMovers]moverDef{
	MoverWalker: {key: "walker", name: "worker at work", title: "Worker", sym: SymWorker, class: CLife, way: WayFoot,
		from: "primitive_age", pace: 5, lines: []string{"On the way to work.", "Heading home after a shift."}},
	MoverHunter: {key: "hunter", name: "hunter", title: "Hunter", sym: SymHunter, class: CLife, way: WayFoot,
		from: "primitive_age", until: "bronze_age", pace: 5,
		lines: []string{"A hunter, off to the hills.", "A hunter, tracking game."}},
	MoverOxCart: {key: "ox_cart", name: "ox cart", title: "Ox cart", sym: SymOxCart, class: CWealth, way: WayStreet,
		from: "iron_age", until: "industrial_age", pace: 4,
		lines: []string{"An ox cart, creaking to market.", "An ox cart, loaded with grain."}},
	MoverRider: {key: "rider", name: "rider", title: "Rider", sym: SymRider, class: CLife, way: WayStreet,
		from: "iron_age", until: "victorian_age", pace: 2,
		lines: []string{"A rider, carrying news.", "A rider, in no hurry."}},
	MoverRowboat: {key: "rowboat", name: "rowboat", title: "Rowboat", sym: SymRowboat, class: CLife, way: WayWater,
		from: "iron_age", until: "atomic_age", pace: 5,
		lines: []string{"A rowboat, out fishing.", "A rowboat, ferrying goods downriver."}},
	MoverWagon: {key: "wagon", name: "wagon", title: "Wagon", sym: SymWagon, class: CWealth, way: WayStreet,
		from: "renaissance_age", until: "electric_age", pace: 3,
		lines: []string{"A wagon, hauling timber.", "A covered wagon, bound for market."}},
	MoverSailShip: {key: "sailing_ship", name: "sailing ship", title: "Sailing ship", sym: SymSailShip, class: CLife,
		way: WayWater, from: "renaissance_age", until: "electric_age", pace: 4,
		lines: []string{"A sailing ship, riding the wind.", "A sailing ship, heavy with cargo."}},
	MoverSteamTrain: {key: "steam_train", name: "steam train", title: "Steam train", sym: SymTrain, class: CWork,
		way: WayRail, from: "industrial_age", until: "atomic_age", pace: 2, cars: 4, smoke: true,
		lines: []string{"A steam train, hauling ore.", "A steam train, hauling coal."}},
	MoverTram: {key: "tram", name: "tram", title: "Tram", sym: SymTram, class: CCivic, way: WayStreet,
		from: "victorian_age", until: "digital_age", pace: 3,
		lines: []string{"A tram, full of commuters.", "A tram, ringing its bell."}},
	MoverSteamship: {key: "steamship", name: "steamship", title: "Steamship", sym: SymShip, class: CLife, way: WayWater,
		from: "victorian_age", until: "information_age", pace: 3, smoke: true,
		lines: []string{"A steamship, trailing smoke.", "A steamship, bound for the coast."}},
	MoverEarlyCar: {key: "early_car", name: "motor car", title: "Motor car", sym: SymAuto, class: CWealth,
		way: WayStreet, from: "electric_age", until: "atomic_age", pace: 3,
		lines: []string{"An early motor car, sputtering along.", "An early motor car, frightening the horses."}},
	MoverCar: {key: "car", name: "car", title: "Car", sym: SymAuto, class: CLife, way: WayStreet,
		from: "modern_age", until: "cyberpunk_age", pace: 2,
		lines: []string{"A car, stuck behind a truck.", "A car, on the school run."}},
	MoverTruck: {key: "truck", name: "truck", title: "Truck", sym: SymTruck, class: CWork, way: WayStreet,
		from: "modern_age", until: "fusion_age", pace: 3,
		lines: []string{"A truck, making deliveries.", "A truck, loaded to the roof."}},
	MoverPlane: {key: "plane", name: "plane", title: "Plane", sym: SymPlane, class: CLife, way: WaySky,
		from: "modern_age", until: "space_age", pace: 1,
		lines: []string{"A plane, passing overhead.", "A plane, on its approach."}},
	MoverBoxShip: {key: "container_ship", name: "container ship", title: "Container ship", sym: SymBoxShip,
		class: CWork, way: WayWater, from: "modern_age", pace: 4,
		lines: []string{"A container ship, stacked high.", "A container ship, low in the water."}},
	MoverTrain: {key: "freight_train", name: "freight train", title: "Freight train", sym: SymTrain, class: CWork,
		way: WayRail, from: "modern_age", until: "digital_age", pace: 2, cars: 5,
		lines: []string{"A freight train, hauling containers.", "A freight train, running on time."}},
	MoverMaglev: {key: "maglev", name: "maglev", title: "Maglev", sym: SymMaglev, class: CCivic, way: WayRail,
		from: "cyberpunk_age", pace: 1, cars: 4,
		lines: []string{"A maglev, gliding on its rail.", "A maglev, running silent."}},
	MoverDrone: {key: "drone", name: "drone", title: "Drone", sym: SymDrone, class: CLife, way: WaySky,
		from: "digital_age", pace: 2, lines: []string{"A delivery drone, on its rounds.", "A drone, watching the streets."}},
	MoverHovercar: {key: "hovercar", name: "hovercar", title: "Hovercar", sym: SymHovercar, class: CCivic,
		way: WayStreet, from: "cyberpunk_age", pace: 2,
		lines: []string{"A hovercar, skimming the street.", "A hovercar, cutting corners."}},
	MoverShuttle: {key: "shuttle", name: "shuttle", title: "Shuttle", sym: SymShuttle, class: CLife, way: WaySky,
		from: "space_age", pace: 1, lines: []string{"A shuttle, climbing to orbit.", "A shuttle, outbound."}},
	MoverSatellite: {key: "satellite", name: "satellite", title: "Satellite", sym: SymSatellite, class: CLife,
		way: WaySky, from: "modern_age", pace: 3, night: true,
		lines: []string{"A satellite, crossing the sky.", "A satellite, catching the sun."}},
	MoverHabitat: {key: "habitat", name: "orbital habitat", title: "Orbital habitat", sym: SymOrbital, class: CWealth,
		way: WaySky, from: "interstellar_age", pace: 12,
		lines: []string{"An orbital habitat, home to thousands.", "An orbital habitat, turning slowly."}},

	// The Earth arc (the Modern Age to the Fusion Age). The sky train runs
	// the megacity's elevated lines (the maglev takes them over once fusion
	// powers the city); the climber rides the space elevator's tether, which
	// rises in the Fusion Age and stays.
	MoverNewsHeli: {key: "news_helicopter", name: "news helicopter", title: "News helicopter", sym: SymNewsHeli,
		class: CLife, way: WaySky, from: "information_age", until: "digital_age", pace: 2,
		lines: []string{"A news helicopter, circling the story.", "A traffic helicopter, counting the jams."}},
	MoverSkyTrain: {key: "sky_train", name: "sky train", title: "Sky train", sym: SymSkyTrain, class: CCivic,
		way: WayRail, from: "cyberpunk_age", until: "cyberpunk_age", pace: 1, cars: 3,
		lines: []string{"A sky train, threading the towers.", "A sky train, packed to the doors."}},
	MoverCrowd: {key: "crowd", name: "crowd", title: "Crowd", sym: SymCrowd, class: CLife, way: WayFoot,
		from: "cyberpunk_age", pace: 6,
		lines: []string{"A crowd, under the neon.", "A crowd, pushing for the sky train."}},
	MoverClimber: {key: "climber", name: "tether climber", title: "Tether climber", sym: SymClimber, class: CWealth,
		way: WaySky, from: "fusion_age", pace: 4,
		lines: []string{"A climber, riding the tether up.", "A climber, hauling cargo to orbit."}},
}

var moverTable = buildMovers()

func buildMovers() [NumMovers]MoverInfo {
	idx := map[string]int{}
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}
	age := func(k string, none int) int {
		if k == "" {
			return none
		}
		if i, ok := idx[k]; ok {
			return i
		}
		return -1 // an unknown age key: never about (TestMoverRoster fails on it)
	}
	var out [NumMovers]MoverInfo
	out[MoverNone].From, out[MoverNone].Until = -1, -1 // nothing introduces nothing
	for k := Mover(1); k < NumMovers; k++ {
		d := moverDefs[k]
		out[k] = MoverInfo{Key: d.key, Name: d.name, Title: d.title, Sym: d.sym, Class: d.class, Way: d.way,
			From: age(d.from, -1), Until: age(d.until, -1), Pace: max(1, d.pace), Cars: max(1, d.cars),
			Smoke: d.smoke, Night: d.night, Lines: d.lines}
		if d.until != "" && out[k].Until < 0 {
			out[k].From = -1
		}
	}
	return out
}

// Info returns a mover's roster entry.
func (k Mover) Info() MoverInfo {
	if k >= NumMovers {
		return moverTable[MoverNone]
	}
	return moverTable[k]
}

// Introduced reports whether the age with index age, or an earlier one, has
// introduced mover k (it may have retired since).
func Introduced(k Mover, age int) bool {
	i := k.Info()
	return i.From >= 0 && age >= i.From
}

// MoversAt lists the movers about in the age with index age, in roster order.
func MoversAt(age int) []Mover {
	var out []Mover
	for k := Mover(1); k < NumMovers; k++ {
		if moverTable[k].In(age) {
			out = append(out, k)
		}
	}
	return out
}
