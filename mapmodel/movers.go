package mapmodel

import "github.com/espresso20/ageforge/vibe"

// movers.go is the traffic roster the map reads: every kind of thing
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

	// def is the table entry these were written as, read again while VibeMode is on
	def *moverDef
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
	key          string
	name, title  vibe.Entry
	sym          Sym
	class        Class
	way          Way
	from, until  string // age keys; until "" runs to the end
	pace, cars   int
	smoke, night bool
	lines        []vibe.Entry
}

// moverDefs is the roster. Every era brings its own: Stone walkers and
// hunters; Iron ox carts, riders and rowboats; Steel wagons, sailing ships
// and, from the Industrial Age, the railway and its steam trains; Electric
// trams, steamships and early cars; Digital cars, trucks, planes, container
// ships and freight trains, news helicopters from the Information Age and
// drones from the Digital Age; Neon maglevs and hovercars, with the
// megacity's sky trains and crowds, the space elevator's climbers from the
// Fusion Age and shuttles from the Space Age's launch pads; Cosmic orbital
// habitats. Satellites cross the night sky from the Modern Age.
var moverDefs = [numAllMovers]moverDef{
	MoverWalker: {key: "walker", name: vibe.E("map.mover.walker.name", "worker at work"), title: vibe.E("map.mover.walker.title", "Worker"), sym: SymWorker, class: CLife, way: WayFoot,
		from: "primitive_age", pace: 5, lines: []vibe.Entry{vibe.E("map.mover.walker.line.01", "On the way to work."), vibe.E("map.mover.walker.line.02", "Heading home after a shift.")}},
	MoverHunter: {key: "hunter", name: vibe.E("map.mover.hunter.name", "hunter"), title: vibe.E("map.mover.hunter.title", "Hunter"), sym: SymHunter, class: CLife, way: WayFoot,
		from: "primitive_age", until: "bronze_age", pace: 5,
		lines: []vibe.Entry{vibe.E("map.mover.hunter.line.01", "A hunter, off to the hills."), vibe.E("map.mover.hunter.line.02", "A hunter, tracking game.")}},
	MoverOxCart: {key: "ox_cart", name: vibe.E("map.mover.ox_cart.name", "ox cart"), title: vibe.E("map.mover.ox_cart.title", "Ox cart"), sym: SymOxCart, class: CWealth, way: WayStreet,
		from: "iron_age", until: "industrial_age", pace: 4,
		lines: []vibe.Entry{vibe.E("map.mover.ox_cart.line.01", "An ox cart, creaking to market."), vibe.E("map.mover.ox_cart.line.02", "An ox cart, loaded with grain.")}},
	MoverRider: {key: "rider", name: vibe.E("map.mover.rider.name", "rider"), title: vibe.E("map.mover.rider.title", "Rider"), sym: SymRider, class: CLife, way: WayStreet,
		from: "iron_age", until: "victorian_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.rider.line.01", "A rider, carrying news."), vibe.E("map.mover.rider.line.02", "A rider, in no hurry.")}},
	MoverRowboat: {key: "rowboat", name: vibe.E("map.mover.rowboat.name", "rowboat"), title: vibe.E("map.mover.rowboat.title", "Rowboat"), sym: SymRowboat, class: CLife, way: WayWater,
		from: "iron_age", until: "atomic_age", pace: 5,
		lines: []vibe.Entry{vibe.E("map.mover.rowboat.line.01", "A rowboat, out fishing."), vibe.E("map.mover.rowboat.line.02", "A rowboat, ferrying goods downriver.")}},
	MoverWagon: {key: "wagon", name: vibe.E("map.mover.wagon.name", "wagon"), title: vibe.E("map.mover.wagon.title", "Wagon"), sym: SymWagon, class: CWealth, way: WayStreet,
		from: "renaissance_age", until: "electric_age", pace: 3,
		lines: []vibe.Entry{vibe.E("map.mover.wagon.line.01", "A wagon, hauling timber."), vibe.E("map.mover.wagon.line.02", "A covered wagon, bound for market.")}},
	MoverSailShip: {key: "sailing_ship", name: vibe.E("map.mover.sailing_ship.name", "sailing ship"), title: vibe.E("map.mover.sailing_ship.title", "Sailing ship"), sym: SymSailShip, class: CLife,
		way: WayWater, from: "renaissance_age", until: "electric_age", pace: 4,
		lines: []vibe.Entry{vibe.E("map.mover.sailing_ship.line.01", "A sailing ship, riding the wind."), vibe.E("map.mover.sailing_ship.line.02", "A sailing ship, heavy with cargo.")}},
	MoverSteamTrain: {key: "steam_train", name: vibe.E("map.mover.steam_train.name", "steam train"), title: vibe.E("map.mover.steam_train.title", "Steam train"), sym: SymTrain, class: CWork,
		way: WayRail, from: "industrial_age", until: "atomic_age", pace: 2, cars: 4, smoke: true,
		lines: []vibe.Entry{vibe.E("map.mover.steam_train.line.01", "A steam train, hauling ore."), vibe.E("map.mover.steam_train.line.02", "A steam train, hauling coal.")}},
	MoverTram: {key: "tram", name: vibe.E("map.mover.tram.name", "tram"), title: vibe.E("map.mover.tram.title", "Tram"), sym: SymTram, class: CCivic, way: WayStreet,
		from: "victorian_age", until: "digital_age", pace: 3,
		lines: []vibe.Entry{vibe.E("map.mover.tram.line.01", "A tram, full of commuters."), vibe.E("map.mover.tram.line.02", "A tram, ringing its bell.")}},
	MoverSteamship: {key: "steamship", name: vibe.E("map.mover.steamship.name", "steamship"), title: vibe.E("map.mover.steamship.title", "Steamship"), sym: SymShip, class: CLife, way: WayWater,
		from: "victorian_age", until: "information_age", pace: 3, smoke: true,
		lines: []vibe.Entry{vibe.E("map.mover.steamship.line.01", "A steamship, trailing smoke."), vibe.E("map.mover.steamship.line.02", "A steamship, bound for the coast.")}},
	MoverEarlyCar: {key: "early_car", name: vibe.E("map.mover.early_car.name", "motor car"), title: vibe.E("map.mover.early_car.title", "Motor car"), sym: SymAuto, class: CWealth,
		way: WayStreet, from: "electric_age", until: "atomic_age", pace: 3,
		lines: []vibe.Entry{vibe.E("map.mover.early_car.line.01", "An early motor car, sputtering along."), vibe.E("map.mover.early_car.line.02", "An early motor car, frightening the horses.")}},
	MoverCar: {key: "car", name: vibe.E("map.mover.car.name", "car"), title: vibe.E("map.mover.car.title", "Car"), sym: SymAuto, class: CLife, way: WayStreet,
		from: "modern_age", until: "cyberpunk_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.car.line.01", "A car, stuck behind a truck."), vibe.E("map.mover.car.line.02", "A car, on the school run.")}},
	MoverTruck: {key: "truck", name: vibe.E("map.mover.truck.name", "truck"), title: vibe.E("map.mover.truck.title", "Truck"), sym: SymTruck, class: CWork, way: WayStreet,
		from: "modern_age", until: "fusion_age", pace: 3,
		lines: []vibe.Entry{vibe.E("map.mover.truck.line.01", "A truck, making deliveries."), vibe.E("map.mover.truck.line.02", "A truck, loaded to the roof.")}},
	MoverPlane: {key: "plane", name: vibe.E("map.mover.plane.name", "plane"), title: vibe.E("map.mover.plane.title", "Plane"), sym: SymPlane, class: CLife, way: WaySky,
		from: "modern_age", until: "space_age", pace: 1,
		lines: []vibe.Entry{vibe.E("map.mover.plane.line.01", "A plane, passing overhead."), vibe.E("map.mover.plane.line.02", "A plane, on its approach.")}},
	MoverBoxShip: {key: "container_ship", name: vibe.E("map.mover.container_ship.name", "container ship"), title: vibe.E("map.mover.container_ship.title", "Container ship"), sym: SymBoxShip,
		class: CWork, way: WayWater, from: "modern_age", pace: 4,
		lines: []vibe.Entry{vibe.E("map.mover.container_ship.line.01", "A container ship, stacked high."), vibe.E("map.mover.container_ship.line.02", "A container ship, low in the water.")}},
	MoverTrain: {key: "freight_train", name: vibe.E("map.mover.freight_train.name", "freight train"), title: vibe.E("map.mover.freight_train.title", "Freight train"), sym: SymTrain, class: CWork,
		way: WayRail, from: "modern_age", until: "digital_age", pace: 2, cars: 5,
		lines: []vibe.Entry{vibe.E("map.mover.freight_train.line.01", "A freight train, hauling containers."), vibe.E("map.mover.freight_train.line.02", "A freight train, running on time.")}},
	MoverMaglev: {key: "maglev", name: vibe.E("map.mover.maglev.name", "maglev"), title: vibe.E("map.mover.maglev.title", "Maglev"), sym: SymMaglev, class: CCivic, way: WayRail,
		from: "cyberpunk_age", pace: 1, cars: 4,
		lines: []vibe.Entry{vibe.E("map.mover.maglev.line.01", "A maglev, gliding on its rail."), vibe.E("map.mover.maglev.line.02", "A maglev, running silent.")}},
	MoverDrone: {key: "drone", name: vibe.E("map.mover.drone.name", "drone"), title: vibe.E("map.mover.drone.title", "Drone"), sym: SymDrone, class: CLife, way: WaySky,
		from: "digital_age", pace: 2, lines: []vibe.Entry{vibe.E("map.mover.drone.line.01", "A delivery drone, on its rounds."), vibe.E("map.mover.drone.line.02", "A drone, watching the streets.")}},
	MoverHovercar: {key: "hovercar", name: vibe.E("map.mover.hovercar.name", "hovercar"), title: vibe.E("map.mover.hovercar.title", "Hovercar"), sym: SymHovercar, class: CCivic,
		way: WayStreet, from: "cyberpunk_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.hovercar.line.01", "A hovercar, skimming the street."), vibe.E("map.mover.hovercar.line.02", "A hovercar, cutting corners.")}},
	MoverShuttle: {key: "shuttle", name: vibe.E("map.mover.shuttle.name", "shuttle"), title: vibe.E("map.mover.shuttle.title", "Shuttle"), sym: SymShuttle, class: CLife, way: WaySky,
		from: "space_age", pace: 1, lines: []vibe.Entry{vibe.E("map.mover.shuttle.line.01", "A shuttle, climbing to orbit."), vibe.E("map.mover.shuttle.line.02", "A shuttle, outbound.")}},
	MoverSatellite: {key: "satellite", name: vibe.E("map.mover.satellite.name", "satellite"), title: vibe.E("map.mover.satellite.title", "Satellite"), sym: SymSatellite, class: CLife,
		way: WaySky, from: "modern_age", pace: 3, night: true,
		lines: []vibe.Entry{vibe.E("map.mover.satellite.line.01", "A satellite, crossing the sky."), vibe.E("map.mover.satellite.line.02", "A satellite, catching the sun.")}},
	MoverHabitat: {key: "habitat", name: vibe.E("map.mover.habitat.name", "orbital habitat"), title: vibe.E("map.mover.habitat.title", "Orbital habitat"), sym: SymOrbital, class: CWealth,
		way: WaySky, from: "interstellar_age", pace: 12,
		lines: []vibe.Entry{vibe.E("map.mover.habitat.line.01", "An orbital habitat, home to thousands."), vibe.E("map.mover.habitat.line.02", "An orbital habitat, turning slowly.")}},

	// The Earth arc (the Modern Age to the Fusion Age). The sky train runs
	// the megacity's elevated lines (the maglev takes them over once fusion
	// powers the city); the climber rides the space elevator's tether, which
	// rises in the Fusion Age and stays.
	MoverNewsHeli: {key: "news_helicopter", name: vibe.E("map.mover.news_helicopter.name", "news helicopter"), title: vibe.E("map.mover.news_helicopter.title", "News helicopter"), sym: SymNewsHeli,
		class: CLife, way: WaySky, from: "information_age", until: "digital_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.news_helicopter.line.01", "A news helicopter, circling the story."), vibe.E("map.mover.news_helicopter.line.02", "A traffic helicopter, counting the jams.")}},
	MoverSkyTrain: {key: "sky_train", name: vibe.E("map.mover.sky_train.name", "sky train"), title: vibe.E("map.mover.sky_train.title", "Sky train"), sym: SymSkyTrain, class: CCivic,
		way: WayRail, from: "cyberpunk_age", until: "cyberpunk_age", pace: 1, cars: 3,
		lines: []vibe.Entry{vibe.E("map.mover.sky_train.line.01", "A sky train, threading the towers."), vibe.E("map.mover.sky_train.line.02", "A sky train, packed to the doors.")}},
	MoverCrowd: {key: "crowd", name: vibe.E("map.mover.crowd.name", "crowd"), title: vibe.E("map.mover.crowd.title", "Crowd"), sym: SymCrowd, class: CLife, way: WayFoot,
		from: "cyberpunk_age", pace: 6,
		lines: []vibe.Entry{vibe.E("map.mover.crowd.line.01", "A crowd, under the neon."), vibe.E("map.mover.crowd.line.02", "A crowd, pushing for the sky train.")}},
	MoverClimber: {key: "climber", name: vibe.E("map.mover.climber.name", "tether climber"), title: vibe.E("map.mover.climber.title", "Tether climber"), sym: SymClimber, class: CWealth,
		way: WaySky, from: "fusion_age", pace: 4,
		lines: []vibe.Entry{vibe.E("map.mover.climber.line.01", "A climber, riding the tether up."), vibe.E("map.mover.climber.line.02", "A climber, hauling cargo to orbit.")}},
	// The sky arc's movers (see below).
	MoverMiningDrone: {key: "mining_drone", name: vibe.E("map.mover.mining_drone.name", "mining drone"), title: vibe.E("map.mover.mining_drone.title", "Mining drone"), sym: SymMiningDrone,
		class: CWork, way: WaySpace, from: "space_age", until: "interstellar_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.mining_drone.line.01", "A mining drone, hauling ore from the belt."), vibe.E("map.mover.mining_drone.line.02", "A mining drone, out for another rock.")}},
	MoverGenShip: {key: "generation_ship", name: vibe.E("map.mover.generation_ship.name", "generation ship"), title: vibe.E("map.mover.generation_ship.title", "Generation ship"), sym: SymGenShip,
		class: CLife, way: WaySpace, from: "interstellar_age", until: "interstellar_age", pace: 6,
		lines: []vibe.Entry{vibe.E("map.mover.generation_ship.line.01", "A generation ship: its crew's grandchildren will land."), vibe.E("map.mover.generation_ship.line.02", "A generation ship, a town in a hull.")}},
	MoverStarship: {key: "starship", name: vibe.E("map.mover.starship.name", "starship"), title: vibe.E("map.mover.starship.title", "Starship"), sym: SymStarship,
		class: CLife, way: WaySpace, from: "galactic_age", until: "galactic_age", pace: 1,
		lines: []vibe.Entry{vibe.E("map.mover.starship.line.01", "A starship, about to jump to warp."), vibe.E("map.mover.starship.line.02", "A starship, home from the far side of the arm.")}},
	MoverAlienShip: {key: "alien_ship", name: vibe.E("map.mover.alien_ship.name", "alien ship"), title: vibe.E("map.mover.alien_ship.title", "Alien ship"), sym: SymUFO,
		class: CFresh, way: WaySpace, from: "galactic_age", until: "galactic_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.alien_ship.line.01", "An alien ship, on ordinary business."), vibe.E("map.mover.alien_ship.line.02", "An alien ship, docking like anyone else.")}},
	MoverPhaseShip: {key: "phase_ship", name: vibe.E("map.mover.phase_ship.name", "phase ship"), title: vibe.E("map.mover.phase_ship.title", "Phase ship"), sym: SymPhaseShip,
		class: CCivic, way: WaySpace, from: "quantum_age", until: "quantum_age", pace: 2,
		lines: []vibe.Entry{vibe.E("map.mover.phase_ship.line.01", "A phase ship, in several places at once."), vibe.E("map.mover.phase_ship.line.02", "A phase ship, arriving before it left.")}},
	MoverMote: {key: "mote", name: vibe.E("map.mover.mote.name", "mote of light"), title: vibe.E("map.mover.mote.title", "Mote of light"), sym: SymMote,
		class: CText, way: WaySpace, from: "transcendent_age", pace: 12,
		lines: []vibe.Entry{vibe.E("map.mover.mote.line.01", "A mote of light, drifting home."), vibe.E("map.mover.mote.line.02", "A mote of light. It was a city once.")}},
}

var moverTable = buildMovers()

func buildMovers() [numAllMovers]MoverInfo {
	var out [numAllMovers]MoverInfo
	out[MoverNone].From, out[MoverNone].Until = -1, -1 // nothing introduces nothing
	for k := Mover(1); k < numAllMovers; k++ {
		d := &moverDefs[k]
		out[k] = MoverInfo{Key: d.key, Name: d.name.Plain(), Title: d.title.Plain(), Sym: d.sym, Class: d.class, Way: d.way,
			From: ageNo(d.from), Until: ageNo(d.until), Pace: max(1, d.pace), Cars: max(1, d.cars),
			Smoke: d.smoke, Night: d.night, Lines: plainLines(d.lines), def: d}
		if d.until != "" && out[k].Until < 0 {
			out[k].From = -1 // an unknown age key: never about (TestMoverRoster fails on it)
		}
	}
	return out
}

// Info returns a mover's roster entry (the sky arc's movers too).
func (k Mover) Info() MoverInfo {
	if k >= numAllMovers {
		k = MoverNone
	}
	return moverTable[k].told()
}

// told is the roster entry with its words as they read now: itself while
// VibeMode is off, else read again from the entries the table was written as.
func (i MoverInfo) told() MoverInfo {
	if vibe.On() && i.def != nil { // (MoverNone has no entry)
		i.Name, i.Title, i.Lines = i.def.name.String(), i.def.title.String(), toldLines(i.def.lines)
	}
	return i
}

// MoversAt lists the movers about in the age with index age, in roster
// order (the sky arc's last).
func MoversAt(age int) []Mover {
	var out []Mover
	for k := Mover(1); k < numAllMovers; k++ {
		if k.Info().In(age) {
			out = append(out, k)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// The sky arc's movers (Space Age to Transcendent Age)
//
// From the Space Age the map leaves the ground (sky.go), and these
// movers travel the sky scenes' own lanes: the asteroid belt, the gulfs
// between colony worlds, the lanes of the starbase. (The tether's climbers
// are the Earth arc's MoverClimber: the elevator rises in the Fusion Age and
// the sky scenes carry it on.) They number on from NumMovers in a range of
// their own, so the ground roster above (and the arrays the map sizes by
// NumMovers) is untouched; Info and MoversAt cover both ranges.
// WaySpace keeps them off the ground's streets, rails, water and sky.
// ---------------------------------------------------------------------------

// WaySpace is open space: the lanes of the sky scenes.
const WaySpace Way = 16

const (
	MoverMiningDrone Mover = NumMovers + iota // a drone hauling ore from the belt
	MoverGenShip                              // a generation ship, trailing its engines
	MoverStarship                             // a starship, jumping to warp
	MoverAlienShip                            // an alien saucer: the visitor, now ordinary traffic
	MoverPhaseShip                            // a ship that tunnels between possibilities
	MoverMote                                 // a mote of light, drifting home
	numAllMovers
)

// NumSkyMovers is how many movers the sky arc adds.
const NumSkyMovers = int(numAllMovers - NumMovers)

// SkyMoverFirst is the first of the sky arc's movers; they run to
// SkyMoverFirst + NumSkyMovers.
const SkyMoverFirst = MoverMiningDrone
