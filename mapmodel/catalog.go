package mapmodel

import (
	"sort"

	"github.com/espresso20/ageforge/config"
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

// LineageNames are the player-facing names of each lineage.
var LineageNames = map[string]string{
	LinHousing: "housing", LinFood: "farms and food", LinWood: "wood and fibre",
	LinMines: "quarries and mines", LinMetal: "metalworks", LinEngineer: "engineering",
	LinEnergy: "power", LinHarbor: "harbor", LinHacker: "hacker dens",
	LinKnowledge: "knowledge", LinFaith: "faith", LinCulture: "culture", LinMonument: "monuments",
	LinTrade: "trade", LinStorage: "storage", LinMilitary: "military", LinDiplomacy: "diplomacy",
	LinWonder: "wonders",
}

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
	WorkerCap  int
	Wonder     bool
	// Slot is the type's fixed position among its age's types in the
	// skyline district (most prominent first).
	Slot int
}

// Catalog is the static data the model reads, built once from config. It
// is immutable after NewCatalog and safe to share.
type Catalog struct {
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
}

// NewCatalog reads config. It is not cheap (config rebuilds its tables on
// every call), so build it once and keep it.
func NewCatalog() *Catalog {
	c := &Catalog{AgeIdx: map[string]int{}, EpochIdx: map[string]int{}, Defs: map[string]*Def{},
		ByLineage: map[string][]*Def{}, Catastrophes: map[string]string{}}
	ages := config.AgeByKey()
	for i, k := range config.AgeOrder() {
		c.Ages = append(c.Ages, k)
		c.AgeNames = append(c.AgeNames, ages[k].Name)
		c.AgeIdx[k] = i
	}
	for i, e := range config.Epochs() {
		c.Epochs = append(c.Epochs, e.Key)
		c.EpochName = append(c.EpochName, e.Name)
		c.EpochIdx[e.Key] = i
		name, _ := config.CatastropheInfo(e.Key)
		c.Catastrophes[e.Key] = name
	}
	for _, k := range c.Ages {
		c.AgeEpoch = append(c.AgeEpoch, c.EpochIdx[config.EpochForAge(k)])
	}
	defs := config.BuildingByKey()
	keys := make([]string, 0, len(defs))
	for k := range defs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		d := defs[k]
		md := &Def{Key: k, Name: d.Name, Category: d.Category, Tier: d.LineageTier,
			Age: c.AgeIdx[d.RequiredAge], Output: d.OutputResource, WorkerCap: d.WorkerCapacity}
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
	for lin, ds := range c.ByLineage {
		sort.Slice(ds, func(i, j int) bool {
			if ds[i].Age != ds[j].Age {
				return ds[i].Age < ds[j].Age
			}
			if ds[i].Tier != ds[j].Tier {
				return ds[i].Tier < ds[j].Tier
			}
			return ds[i].Key < ds[j].Key
		})
		for r, d := range ds {
			d.Rank = r
		}
		c.ByLineage[lin] = ds
	}
	c.Wonders = c.ByLineage[LinWonder]
	c.ByAge = make([][]*Def, len(c.Ages))
	for _, k := range keys {
		d := c.Defs[k]
		c.ByAge[d.Age] = append(c.ByAge[d.Age], d)
	}
	for a, ds := range c.ByAge {
		sort.Slice(ds, func(i, j int) bool {
			pi, pj := prominence(ds[i]), prominence(ds[j])
			if pi != pj {
				return pi > pj
			}
			return ds[i].Key < ds[j].Key
		})
		for s, d := range ds {
			d.Slot = s
		}
		c.ByAge[a] = ds
	}
	for _, f := range config.BaseFactions() {
		c.Factions = append(c.Factions, f.Key)
	}
	return c
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
