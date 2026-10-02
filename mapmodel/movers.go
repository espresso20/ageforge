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

// Info returns a mover's roster entry (the sky arc's movers too).
func (k Mover) Info() MoverInfo {
	if k >= NumMovers && k < numAllMovers {
		return skyMoverTable[k-NumMovers]
	}
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
// From the Space Age both map styles leave the ground (sky.go), and these
// movers travel the sky scenes' own lanes: the asteroid belt, the gulfs
// between colony worlds, the lanes of the starbase. (The tether's climbers
// are the Earth arc's MoverClimber: the elevator rises in the Fusion Age and
// the sky scenes carry it on.) They number on from NumMovers in a range of
// their own, so the ground roster above (and the arrays the styles size by
// NumMovers) is untouched; Info, Introduced and MoversAt cover both ranges.
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

var skyMoverDefs = [numAllMovers - NumMovers]moverDef{
	MoverMiningDrone - NumMovers: {key: "mining_drone", name: "mining drone", title: "Mining drone", sym: SymMiningDrone,
		class: CWork, way: WaySpace, from: "space_age", until: "interstellar_age", pace: 2,
		lines: []string{"A mining drone, hauling ore from the belt.", "A mining drone, out for another rock."}},
	MoverGenShip - NumMovers: {key: "generation_ship", name: "generation ship", title: "Generation ship", sym: SymGenShip,
		class: CLife, way: WaySpace, from: "interstellar_age", until: "interstellar_age", pace: 6,
		lines: []string{"A generation ship: its crew's grandchildren will land.", "A generation ship, a town in a hull."}},
	MoverStarship - NumMovers: {key: "starship", name: "starship", title: "Starship", sym: SymStarship,
		class: CLife, way: WaySpace, from: "galactic_age", until: "galactic_age", pace: 1,
		lines: []string{"A starship, about to jump to warp.", "A starship, home from the far side of the arm."}},
	MoverAlienShip - NumMovers: {key: "alien_ship", name: "alien ship", title: "Alien ship", sym: SymUFO,
		class: CFresh, way: WaySpace, from: "galactic_age", until: "galactic_age", pace: 2,
		lines: []string{"An alien ship, on ordinary business.", "An alien ship, docking like anyone else."}},
	MoverPhaseShip - NumMovers: {key: "phase_ship", name: "phase ship", title: "Phase ship", sym: SymPhaseShip,
		class: CCivic, way: WaySpace, from: "quantum_age", until: "quantum_age", pace: 2,
		lines: []string{"A phase ship, in several places at once.", "A phase ship, arriving before it left."}},
	MoverMote - NumMovers: {key: "mote", name: "mote of light", title: "Mote of light", sym: SymMote,
		class: CText, way: WaySpace, from: "transcendent_age", pace: 12,
		lines: []string{"A mote of light, drifting home.", "A mote of light. It was a city once."}},
}

var skyMoverTable = buildSkyMovers()

func buildSkyMovers() [numAllMovers - NumMovers]MoverInfo {
	idx := map[string]int{}
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}
	age := func(k string) int {
		if k == "" {
			return -1
		}
		if i, ok := idx[k]; ok {
			return i
		}
		return -2 // an unknown age key: never about (TestSkyMoverRoster fails on it)
	}
	var out [numAllMovers - NumMovers]MoverInfo
	for i, d := range skyMoverDefs {
		info := MoverInfo{Key: d.key, Name: d.name, Title: d.title, Sym: d.sym, Class: d.class, Way: d.way,
			From: age(d.from), Until: age(d.until), Pace: max(1, d.pace), Cars: max(1, d.cars),
			Smoke: d.smoke, Night: d.night, Lines: d.lines}
		if info.From < 0 || info.Until == -2 {
			info.From, info.Until = -1, -1
		}
		out[i] = info
	}
	return out
}
