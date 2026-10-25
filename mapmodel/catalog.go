package mapmodel

import (
	"cmp"
	"maps"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// Lineage keys the model groups buildings by. They are the config lineages,
// plus "monument" and "diplomacy" for the few buildings config leaves without
// one, and "wonder".
const (
	LinHousing   = "housing"
	LinFood      = "food"
	LinWood      = "organic_extraction"
	LinMines     = "geological_extraction"
	LinMetal     = "metallurgy"
	LinEngineer  = "engineering"
	LinEnergy    = "energy"
	LinHarbor    = "harbor"
	LinHacker    = "hacker"
	LinKnowledge = "knowledge"
	LinFaith     = "faith"
	LinCulture   = "culture_arts"
	LinMonument  = "monument"
	LinTrade     = "trade"
	LinStorage   = "storage"
	LinMilitary  = "military"
	LinDiplomacy = "diplomacy"
	LinWonder    = "wonder"
)

// LineageOrder is the fixed order lineages are listed and laid out in. The
// town rotates it by seed; everything else reads it as is.
var LineageOrder = []string{
	LinHousing, LinFood, LinWood, LinMines, LinMetal, LinEngineer, LinEnergy, LinHarbor,
	LinHacker, LinKnowledge, LinFaith, LinCulture, LinMonument, LinTrade, LinStorage,
	LinMilitary, LinDiplomacy,
}

// ageIndex is each age key's place in the config age order. The rosters
// (movers, city features, city looks) name their ages by key and are built at
// start-up, before any ruleset is in reach.
var ageIndex = func() map[string]int {
	order := config.AgeOrder()
	out := make(map[string]int, len(order))
	for i, k := range order {
		out[k] = i
	}
	return out
}()

// ageNo is the index of the age with config key k, or -1 for "" or a key
// that is not an age.
func ageNo(k string) int {
	if i, ok := ageIndex[k]; ok {
		return i
	}
	return -1
}

// LineageNames are the player-facing names of each lineage.
var LineageNames = map[string]vibe.Entry{
	LinHousing: vibe.E("map.lineage.housing", "housing"), LinFood: vibe.E("map.lineage.food", "farms and food"), LinWood: vibe.E("map.lineage.organic_extraction", "wood and fibre"),
	LinMines: vibe.E("map.lineage.geological_extraction", "quarries and mines"), LinMetal: vibe.E("map.lineage.metallurgy", "metalworks"), LinEngineer: vibe.E("map.lineage.engineering", "engineering"),
	LinEnergy: vibe.E("map.lineage.energy", "power"), LinHarbor: vibe.E("map.lineage.harbor", "harbor"), LinHacker: vibe.E("map.lineage.hacker", "hacker dens"),
	LinKnowledge: vibe.E("map.lineage.knowledge", "knowledge"), LinFaith: vibe.E("map.lineage.faith", "faith"), LinCulture: vibe.E("map.lineage.culture_arts", "culture"), LinMonument: vibe.E("map.lineage.monument", "monuments"),
	LinTrade: vibe.E("map.lineage.trade", "trade"), LinStorage: vibe.E("map.lineage.storage", "storage"), LinMilitary: vibe.E("map.lineage.military", "military"), LinDiplomacy: vibe.E("map.lineage.diplomacy", "diplomacy"),
	LinWonder: vibe.E("map.lineage.wonder", "wonders"),
}

// LineageName is the name of a lineage as it reads now ("" for a key with none).
func LineageName(lin string) string { return LineageNames[lin].String() }

// Def is one building type as the map sees it: static catalogue data.
type Def struct {
	Key, Name, Lineage, Category string
	// Tier is the config lineage tier; Rank orders a lineage's types oldest
	// first and is unique within the lineage (config repeats tiers for
	// monuments and diplomacy).
	Tier, Rank int
	Age        int // index of the age it becomes available in
	Epoch      int // index of that age's epoch
	Output     string
	BaseOutput float64 // per-instance base production of Output
	Wonder     bool
	// Slot is the type's fixed position among its age's types in the
	// skyline district (most prominent first).
	Slot int
}

// Catalog is the static data the model reads, built once from a ruleset. It
// is immutable after NewCatalogFor and safe to share.
type Catalog struct {
	// Rules is the ruleset the catalogue was built from.
	Rules     *rules.Set
	Ages      []string
	AgeNames  []string
	AgeIdx    map[string]int
	Epochs    []string
	EpochName []string
	EpochIdx  map[string]int
	AgeEpoch  []int // age index -> epoch index
	Defs      map[string]*Def
	// ByAge lists each age's types in skyline slot order.
	ByAge [][]*Def
	// ByLineage lists each lineage's types in rank order.
	ByLineage map[string][]*Def
	// Wonders lists the wonder types by age.
	Wonders []*Def
	// Factions lists the civ keys in config order; a civ's index here is
	// its site on the region map.
	Factions []string
	// Catastrophes maps epoch key -> catastrophe display name.
	Catastrophes map[string]string
	// RouteModes maps trade route key -> how its goods travel.
	RouteModes map[string]RouteMode
}

// NewCatalog is the catalogue of the core ruleset.
func NewCatalog() *Catalog { return NewCatalogFor(rules.Core()) }

// NewCatalogFor builds the catalogue of set. It sorts and ranks every
// building, so build it once and keep it.
func NewCatalogFor(set *rules.Set) *Catalog {
	c := &Catalog{Rules: set, AgeIdx: map[string]int{}, EpochIdx: map[string]int{}, Defs: map[string]*Def{},
		ByLineage: map[string][]*Def{}, Catastrophes: map[string]string{}}
	for i, a := range set.Ages() {
		c.Ages = append(c.Ages, a.Key)
		c.AgeNames = append(c.AgeNames, a.Name)
		c.AgeIdx[a.Key] = i
	}
	for i, e := range set.Eras() {
		c.Epochs = append(c.Epochs, e.Key)
		c.EpochName = append(c.EpochName, e.Name)
		c.EpochIdx[e.Key] = i
		name, _ := set.Catastrophe(e.Key)
		c.Catastrophes[e.Key] = name
	}
	for _, k := range c.Ages {
		c.AgeEpoch = append(c.AgeEpoch, c.EpochIdx[set.EraOf(k)])
	}
	defs := set.BuildingMap()
	keys := slices.Sorted(maps.Keys(defs))
	for _, k := range keys {
		d := defs[k]
		md := &Def{Key: k, Name: d.Name, Category: d.Category, Tier: d.LineageTier,
			Age: c.AgeIdx[d.RequiredAge], Output: d.OutputResource}
		md.Epoch = c.AgeEpoch[md.Age]
		md.Lineage = lineageOf(d)
		md.Wonder = md.Lineage == LinWonder
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == d.OutputResource {
				md.BaseOutput += e.Value
			}
		}
		c.Defs[k] = md
		c.ByLineage[md.Lineage] = append(c.ByLineage[md.Lineage], md)
	}
	for _, ds := range c.ByLineage {
		slices.SortFunc(ds, func(a, b *Def) int {
			return cmp.Or(cmp.Compare(a.Age, b.Age), cmp.Compare(a.Tier, b.Tier), cmp.Compare(a.Key, b.Key))
		})
		for r, d := range ds {
			d.Rank = r
		}
	}
	c.Wonders = c.ByLineage[LinWonder]
	c.ByAge = make([][]*Def, len(c.Ages))
	for _, k := range keys {
		d := c.Defs[k]
		c.ByAge[d.Age] = append(c.ByAge[d.Age], d)
	}
	for _, ds := range c.ByAge {
		slices.SortFunc(ds, func(a, b *Def) int {
			return cmp.Or(cmp.Compare(prominence(b), prominence(a)), cmp.Compare(a.Key, b.Key))
		})
		for s, d := range ds {
			d.Slot = s
		}
	}
	for _, f := range set.Factions() {
		c.Factions = append(c.Factions, f.Key)
	}
	c.RouteModes = map[string]RouteMode{}
	for _, r := range set.TradeRoutes() {
		c.RouteModes[r.Key] = routeModeOf(r, c.Defs[r.RequiredBld])
	}
	return c
}

// RouteMode is how a trade route's goods travel, so a style can draw the
// right vehicles: caravans and trains, ships, or aircraft.
type RouteMode uint8

const (
	ModeLand RouteMode = iota
	ModeSea
	ModeAir
)

func (m RouteMode) String() string { return [...]string{"land", "sea", "air"}[m] }

// routeWords are the words that put a route in a mode, in the order they are
// tried: the sky and space first, then rails, roads and caravans (a route to
// "ship iron by rail" is a train), then ships and the sea.
var routeWords = []struct {
	mode  RouteMode
	words []string
}{
	{ModeAir, []string{"warp", "stellar", "galactic", "orbital", "space", "air", "sky"}},
	{ModeLand, []string{"rail", "road", "caravan", "pipeline", "train", "truck", "convoy"}},
	{ModeSea, []string{"port", "harbor", "ship", "clipper", "barge", "sea", "ocean", "steamship"}},
}

// routeModeOf reads a route's mode from its data: a route that needs a
// harbor building sails; one whose words name the sky or space (warp,
// stellar, galactic, orbital, air) flies; rails, roads and caravans go by
// land; ships or the sea sail; the rest go by land.
func routeModeOf(r config.TradeRouteDef, req *Def) RouteMode {
	if req != nil && req.Lineage == LinHarbor {
		return ModeSea
	}
	words := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(r.Key+" "+r.Name+" "+r.RequiredBld+" "+r.Description),
		func(c rune) bool { return c < 'a' || c > 'z' }) {
		words[strings.TrimSuffix(w, "s")] = true
	}
	for _, g := range routeWords {
		if slices.ContainsFunc(g.words, func(w string) bool { return words[w] }) {
			return g.mode
		}
	}
	return ModeLand
}

// RouteMode is the mode of the route with key k. A route the catalogue
// does not know (an old save's, a removed one) gets a stable pick, so it
// still draws the same way every time.
func (c *Catalog) RouteMode(k string) RouteMode {
	if m, ok := c.RouteModes[k]; ok {
		return m
	}
	return RouteMode(Hash(HashStr(k)) % 3)
}

// lineageOf normalises config's lineage key.
func lineageOf(d config.BuildingDef) string {
	switch {
	case d.Category == "wonder" || d.LineageKey == LinWonder:
		return LinWonder
	case d.Category == "monument":
		return LinMonument
	case d.Category == "storage":
		return LinStorage
	case d.LineageKey == "":
		if d.Category == "diplomacy" {
			return LinDiplomacy
		}
		return LinEngineer
	}
	return d.LineageKey
}

// prominence ranks how much a type stands out on a skyline: landmarks and
// civic buildings rise in a district's middle, sheds and stores sit at its
// edges.
func prominence(d *Def) int {
	switch d.Lineage {
	case LinWonder:
		return 9
	case LinMonument, LinFaith:
		return 8
	case LinKnowledge, LinCulture:
		return 7
	case LinHousing, LinMilitary, LinDiplomacy:
		return 6
	case LinTrade, LinEnergy, LinHacker:
		return 5
	case LinMetal, LinEngineer, LinHarbor:
		return 4
	case LinWood, LinMines:
		return 3
	case LinStorage:
		return 2
	}
	return 1
}

// EpochOfAge returns the epoch index of an age index (clamped).
func (c *Catalog) EpochOfAge(age int) int {
	if age < 0 {
		age = 0
	}
	if age >= len(c.AgeEpoch) {
		age = len(c.AgeEpoch) - 1
	}
	return c.AgeEpoch[age]
}
