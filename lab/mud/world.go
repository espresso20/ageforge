package main

// world.go turns a GameState into the walkable settlement: a fixed compass
// graph of places, each fed by one or more building lineages, named for the
// epoch, and shown only once something stands there (or once the road needs
// to pass through it).

import (
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Dir is a compass direction. In the Cosmic Era the words change (spinward,
// hubward...) but the keys do not.
type Dir int

const (
	North Dir = iota
	East
	South
	West
)

func (d Dir) Opposite() Dir { return (d + 2) % 4 }

// dirWords are the direction names per register: planet-bound, then orbital.
var dirWords = [2][4]string{
	{"north", "east", "south", "west"},
	{"hubward", "spinward", "rimward", "antispinward"},
}

// PlaceKey identifies a place across ages; its NAME changes, its key never.
type PlaceKey string

const (
	Square   PlaceKey = "square"
	Homes    PlaceKey = "homes"
	Fields   PlaceKey = "fields"
	Woods    PlaceKey = "woods"
	Quarry   PlaceKey = "quarry"
	Forge    PlaceKey = "forge"
	Works    PlaceKey = "works"
	Temple   PlaceKey = "temple"
	Academy  PlaceKey = "academy"
	Market   PlaceKey = "market"
	Harbour  PlaceKey = "harbour"
	Stores   PlaceKey = "stores"
	Barracks PlaceKey = "barracks"
	Wonders  PlaceKey = "wonders"
	Gate     PlaceKey = "gate"
)

// PlaceDef is the static description of a place.
type PlaceDef struct {
	Key      PlaceKey
	Col, Row int
	Lineages []string
	// Names per epoch, in config.Epochs() order. Short are the minimap labels.
	Names [7]string
	Short [7]string
	// Always places exist from the first tick (the square and the gate).
	Always bool
}

var placeDefs = []PlaceDef{
	{Key: Temple, Col: 1, Row: 0, Lineages: []string{"faith"},
		Names: [7]string{"the Shrine", "the Temple Mount", "the Cathedral Close", "Chapel Row", "the Meditation Gardens", "the Neon Sanctuary", "the Void Cloister"},
		Short: [7]string{"Shrine", "Temple", "Close", "Chapels", "Gardens", "Sanctum", "Cloister"}},
	{Key: Academy, Col: 2, Row: 0, Lineages: []string{"knowledge", "hacker"},
		Names: [7]string{"the Story Circle", "the Library Steps", "the University", "the Institute", "the Campus", "the Server Halls", "the Observatory Deck"},
		Short: [7]string{"Stories", "Library", "Univ.", "Instit.", "Campus", "Servers", "Observ."}},
	{Key: Wonders, Col: 3, Row: 0, Lineages: []string{"wonder", "monument"},
		Names: [7]string{"the Hallowed Rise", "the Wonder Walk", "the Grand Promenade", "the Exposition Grounds", "the Monument Mall", "the Citadel Steps", "the Beacon Spire"},
		Short: [7]string{"Rise", "Wonders", "Promen.", "Expo", "Mall", "Citadel", "Spire"}},
	{Key: Woods, Col: 0, Row: 1, Lineages: []string{"organic_extraction"},
		Names: [7]string{"the Woods", "the Timber Yards", "the Coal Pits", "the Oil Fields", "the Refineries", "the Nanovats", "the Matter Weavers"},
		Short: [7]string{"Woods", "Timber", "Coal", "Oil", "Refin.", "Vats", "Weavers"}},
	{Key: Homes, Col: 1, Row: 1, Lineages: []string{"housing"},
		Names: [7]string{"the Huts", "the Lanes", "the Terraces", "the Rows", "the Blocks", "the Stacks", "the Habitat Rings"},
		Short: [7]string{"Huts", "Lanes", "Terrace", "Rows", "Blocks", "Stacks", "Habitat"}},
	{Key: Square, Col: 2, Row: 1, Lineages: []string{"culture_arts", ""}, Always: true,
		Names: [7]string{"the Village Green", "the Forum", "the Piazza", "the Plaza", "the Concourse", "the Arcology Atrium", "the Orbital Ring"},
		Short: [7]string{"Green", "Forum", "Piazza", "Plaza", "Concou.", "Atrium", "Ring"}},
	{Key: Market, Col: 3, Row: 1, Lineages: []string{"trade"},
		Names: [7]string{"the Barter Stones", "Guild Row", "the Exchange", "the Financial District", "the Venture Quarter", "the Night Market", "the Bazaar Ring"},
		Short: [7]string{"Barter", "Guilds", "Exch.", "Finance", "Venture", "Night", "Bazaar"}},
	{Key: Harbour, Col: 4, Row: 1, Lineages: []string{"harbor"},
		Names: [7]string{"the Landing", "the Quay", "the Harbour", "the Docks", "the Container Port", "the Logistics Hub", "the Docking Spindle"},
		Short: [7]string{"Landing", "Quay", "Harbour", "Docks", "Port", "Logist.", "Docking"}},
	{Key: Fields, Col: 0, Row: 2, Lineages: []string{"food"},
		Names: [7]string{"the Gathering Grounds", "the Demesne Fields", "the Plantations", "the Farm Belt", "the Agri-Complex", "the Vat Halls", "the Hydroponic Bays"},
		Short: [7]string{"Fields", "Fields", "Plant.", "Farms", "Agri", "Vats", "Hydro"}},
	{Key: Stores, Col: 1, Row: 2, Lineages: []string{"storage"},
		Names: [7]string{"the Stash Pits", "the Granaries", "the Warehouses", "the Depot", "the Archive Vaults", "the Cyber Vaults", "the Orbital Depots"},
		Short: [7]string{"Stash", "Granary", "Wareh.", "Depot", "Vaults", "Vaults", "Depots"}},
	{Key: Works, Col: 2, Row: 2, Lineages: []string{"engineering", "energy"},
		Names: [7]string{"Smithy Lane", "the Workshops", "the Mills", "the Power Yards", "the Grid", "the Augment Foundry", "the Launch Complex"},
		Short: [7]string{"Smithy", "Works", "Mills", "Power", "Grid", "Augment", "Launch"}},
	{Key: Forge, Col: 3, Row: 2, Lineages: []string{"metallurgy"},
		Names: [7]string{"the Smelting Pits", "the Forge Quarter", "the Foundries", "the Arc Furnaces", "the Alloy Plants", "the Dark Refinery", "the Star Forge"},
		Short: [7]string{"Smelt", "Forges", "Foundry", "Furnace", "Alloys", "Refin.", "StarFrg"}},
	{Key: Quarry, Col: 0, Row: 3, Lineages: []string{"geological_extraction"},
		Names: [7]string{"the Stone Pits", "the Quarry", "the Mines", "the Uranium Works", "the Deep Mines", "the Crystal Mines", "the Asteroid Tethers"},
		Short: [7]string{"Pits", "Quarry", "Mines", "Uranium", "Deep", "Crystal", "Tethers"}},
	{Key: Barracks, Col: 1, Row: 3, Lineages: []string{"military"},
		Names: [7]string{"the War Camp", "the Barracks Yard", "the Fort", "the Garrison", "the Command Post", "the Drone Pens", "the Fleet Yards"},
		Short: [7]string{"WarCamp", "Barrack", "Fort", "Garris.", "Command", "Drones", "Fleet"}},
	{Key: Gate, Col: 2, Row: 3, Always: true,
		Names: [7]string{"the Track Out", "the Town Gate", "the Toll Road", "the Railway Station", "the Interchange", "the Skyport", "the Jump Gate"},
		Short: [7]string{"Track", "Gate", "Toll", "Station", "Interch", "Skyport", "JumpGt"}},
}

// edges are the roads: a place, the direction you leave it in, and where you arrive.
var edges = []struct {
	A PlaceKey
	D Dir
	B PlaceKey
}{
	{Homes, East, Square}, {Square, East, Market}, {Market, East, Harbour},
	{Woods, East, Homes}, {Fields, East, Stores}, {Quarry, East, Barracks},
	{Woods, South, Fields}, {Fields, South, Quarry},
	{Temple, South, Homes}, {Academy, South, Square}, {Wonders, South, Market},
	{Temple, East, Academy}, {Academy, East, Wonders},
	{Homes, South, Stores}, {Square, South, Works}, {Market, South, Forge},
	{Stores, East, Works}, {Works, East, Forge},
	{Stores, South, Barracks}, {Works, South, Gate}, {Barracks, East, Gate},
}

// Holding is one building type standing in a place.
type Holding struct {
	Key      string
	Name     string
	Count    int
	Ruins    int
	Legacy   bool
	Tier     int // lineage tier: newer is higher
	Workers  int
	Capacity int // total worker slots across copies
	Lineage  string
	Category string
	Flavor   string
	Desc     string
}

// Place is one room of the settlement as it stands now.
type Place struct {
	Def      *PlaceDef
	Key      PlaceKey
	Name     string
	Short    string
	Holdings []Holding // newest tier first
	Settled  bool      // something stands here
	Path     bool      // present only because the road runs through it
	Exits    map[Dir]PlaceKey
}

// Total is the number of standing buildings.
func (p *Place) Total() int {
	n := 0
	for _, h := range p.Holdings {
		n += h.Count
	}
	return n
}

// Ruins is the number of ruined buildings.
func (p *Place) Ruins() int {
	n := 0
	for _, h := range p.Holdings {
		n += h.Ruins
	}
	return n
}

// Staffing is workers over worker slots, or -1 where nothing takes workers.
func (p *Place) Staffing() float64 {
	w, c := 0, 0
	for _, h := range p.Holdings {
		w += h.Workers
		c += h.Capacity
	}
	if c == 0 {
		return -1
	}
	return float64(w) / float64(c)
}

// City is the whole walkable settlement for one GameState.
type City struct {
	St       game.GameState
	Epoch    int // index into config.Epochs()
	AgeIdx   int
	Places   map[PlaceKey]*Place
	Order    []PlaceKey // stable display order (grid order)
	Orbital  bool       // cosmic register for directions
	Cols     int        // grid extent of the places that exist
	Rows     int
	MinCol   int
	MinRow   int
	defs     map[string]config.BuildingDef
	lineHome map[string]PlaceKey
}

// DirWord names d in this city's register.
func (c *City) DirWord(d Dir) string {
	if c.Orbital {
		return dirWords[1][d]
	}
	return dirWords[0][d]
}

func epochIndex(age string) int {
	ep := config.EpochForAge(age)
	for i, e := range config.Epochs() {
		if e.Key == ep {
			return i
		}
	}
	return 0
}

func ageIndex(age string) int {
	for i, a := range config.AgeOrder() {
		if a == age {
			return i
		}
	}
	return 0
}

// BuildCity lays the settlement out from st.
func BuildCity(st game.GameState) *City {
	c := &City{
		St:       st,
		Epoch:    epochIndex(st.Age),
		AgeIdx:   ageIndex(st.Age),
		Places:   map[PlaceKey]*Place{},
		defs:     config.BuildingByKey(),
		lineHome: map[string]PlaceKey{},
	}
	c.Orbital = c.Epoch == 6
	for i := range placeDefs {
		d := &placeDefs[i]
		for _, l := range d.Lineages {
			c.lineHome[l] = d.Key
		}
		if d.Col+1 > c.Cols {
			c.Cols = d.Col + 1
		}
		if d.Row+1 > c.Rows {
			c.Rows = d.Row + 1
		}
		c.Places[d.Key] = &Place{Def: d, Key: d.Key, Name: d.Names[c.Epoch], Short: d.Short[c.Epoch], Exits: map[Dir]PlaceKey{}}
	}
	// Primitive: there is no village yet, only a clearing.
	if c.AgeIdx == 0 {
		c.Places[Square].Name, c.Places[Square].Short = "the Clearing", "Fire"
	}

	for key, b := range st.Buildings {
		if b.Count == 0 && b.RuinCount == 0 {
			continue
		}
		def := c.defs[key]
		home, ok := c.lineHome[def.LineageKey]
		if !ok {
			home = Square
		}
		p := c.Places[home]
		p.Holdings = append(p.Holdings, Holding{
			Key: key, Name: b.Name, Count: b.Count, Ruins: b.RuinCount, Legacy: b.IsLegacy,
			Tier: ageIndex(def.RequiredAge), Workers: b.WorkersAssigned, Capacity: b.WorkerCapacity * b.Count,
			Lineage: def.LineageKey, Category: b.Category, Flavor: b.Flavor, Desc: b.Description,
		})
	}
	for _, p := range c.Places {
		sort.Slice(p.Holdings, func(i, j int) bool {
			a, b := p.Holdings[i], p.Holdings[j]
			if a.Tier != b.Tier {
				return a.Tier > b.Tier
			}
			return a.Key < b.Key
		})
		p.Settled = p.Total() > 0 || p.Ruins() > 0
	}
	// The harbour needs water; it appears with its first building. Everything
	// else settled must be reachable from the square: unsettled places on the
	// shortest road to a settled one become paths.
	present := map[PlaceKey]bool{}
	for k, p := range c.Places {
		if p.Settled || p.Def.Always {
			present[k] = true
		}
	}
	adj := c.adjacency(nil)
	for k := range present {
		for _, step := range bfsPath(adj, Square, k) {
			if !present[step] {
				present[step] = true
				c.Places[step].Path = true
				c.Places[step].Name = pathName(c, c.Places[step].Def)
			}
		}
	}
	for k := range c.Places {
		if !present[k] {
			delete(c.Places, k)
		}
	}
	for _, e := range edges {
		a, b := c.Places[e.A], c.Places[e.B]
		if a == nil || b == nil {
			continue
		}
		a.Exits[e.D] = e.B
		b.Exits[e.D.Opposite()] = e.A
	}
	c.MinCol, c.MinRow, c.Cols, c.Rows = 99, 99, 0, 0
	maxCol, maxRow := 0, 0
	for _, d := range placeDefs {
		if _, ok := c.Places[d.Key]; ok {
			c.Order = append(c.Order, d.Key)
			c.MinCol, c.MinRow = min(c.MinCol, d.Col), min(c.MinRow, d.Row)
			maxCol, maxRow = max(maxCol, d.Col), max(maxRow, d.Row)
		}
	}
	c.Cols, c.Rows = maxCol-c.MinCol+1, maxRow-c.MinRow+1
	sort.SliceStable(c.Order, func(i, j int) bool {
		a, b := c.Places[c.Order[i]].Def, c.Places[c.Order[j]].Def
		if a.Row != b.Row {
			return a.Row < b.Row
		}
		return a.Col < b.Col
	})
	return c
}

// adjacency is the road graph over the places in keep (nil = every place).
func (c *City) adjacency(keep map[PlaceKey]*Place) map[PlaceKey][]PlaceKey {
	adj := map[PlaceKey][]PlaceKey{}
	for _, e := range edges {
		if keep != nil && (keep[e.A] == nil || keep[e.B] == nil) {
			continue
		}
		adj[e.A] = append(adj[e.A], e.B)
		adj[e.B] = append(adj[e.B], e.A)
	}
	return adj
}

// bfsPath is the shortest road from a to b, excluding a, including b.
func bfsPath(adj map[PlaceKey][]PlaceKey, a, b PlaceKey) []PlaceKey {
	if a == b {
		return nil
	}
	prev := map[PlaceKey]PlaceKey{a: a}
	q := []PlaceKey{a}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		for _, m := range adj[n] {
			if _, seen := prev[m]; seen {
				continue
			}
			prev[m] = n
			if m == b {
				var path []PlaceKey
				for x := b; x != a; x = prev[x] {
					path = append([]PlaceKey{x}, path...)
				}
				return path
			}
			q = append(q, m)
		}
	}
	return nil
}

// Route is the walk from a to b over the roads that exist now.
func (c *City) Route(a, b PlaceKey) []PlaceKey {
	return bfsPath(c.adjacency(c.Places), a, b)
}

// Resolve matches what a player typed against places (by name, short label or
// key) and against buildings (by name or key); it returns the place and the
// building key when one matched.
func (c *City) Resolve(q string) (PlaceKey, string, bool) {
	q = strings.ToLower(strings.TrimSpace(q))
	q = strings.TrimPrefix(q, "the ")
	if q == "" {
		return "", "", false
	}
	norm := func(s string) string {
		s = strings.ToLower(s)
		s = strings.TrimPrefix(s, "the ")
		return strings.ReplaceAll(s, "_", " ")
	}
	for _, k := range c.Order {
		p := c.Places[k]
		if norm(p.Name) == q || norm(p.Short) == q || string(k) == q {
			return k, "", true
		}
	}
	// Buildings: exact, then singular/plural-insensitive prefix.
	for _, k := range c.Order {
		for _, h := range c.Places[k].Holdings {
			if norm(h.Name) == q || norm(h.Key) == q || norm(h.Name)+"s" == q {
				return k, h.Key, true
			}
		}
	}
	for _, k := range c.Order {
		p := c.Places[k]
		if strings.HasPrefix(norm(p.Name), q) || strings.HasPrefix(norm(p.Short), q) {
			return k, "", true
		}
		for _, h := range p.Holdings {
			if strings.HasPrefix(norm(h.Name), q) {
				return k, h.Key, true
			}
		}
	}
	return "", "", false
}

// pathName names a place that is only road so far, by where it lies from the
// square: "the south track", "the north-east avenue".
func pathName(c *City, d *PlaceDef) string {
	var parts []string
	if d.Row < 1 {
		parts = append(parts, c.DirWord(North))
	} else if d.Row > 1 {
		parts = append(parts, c.DirWord(South))
	}
	if d.Col < 2 {
		parts = append(parts, c.DirWord(West))
	} else if d.Col > 2 {
		parts = append(parts, c.DirWord(East))
	}
	noun := [4]string{"track", "road", "avenue", "corridor"}[band(c.Epoch)]
	return "the " + strings.Join(parts, "-") + " " + noun
}
