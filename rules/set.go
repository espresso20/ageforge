package rules

import (
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Set is one compiled ruleset. It never changes after Compile, so any number
// of engines, snapshots and goroutines may share one without a lock.
//
// Every slice and map a method returns belongs to the set: read it, never
// write to it. The methods named ...Map are the exception: they return a
// fresh map the caller owns (the definitions inside it still share their
// own maps and slices with the set).
type Set struct {
	ages     []config.AgeDef
	ageKeys  []string
	agePos   map[string]int
	ageByKey map[string]config.AgeDef
	wonders  map[string]string // age key -> its wonder's building key

	eras       []config.EpochDef
	eraByKey   map[string]config.EpochDef
	eraPos     map[string]int
	eraOfAge   map[string]string
	eraFirst   map[string]int // era key -> position of its first age
	gateOrder  int            // Order of the era catastrophes start in
	dooms      map[string]Catastrophe
	unknown    Catastrophe
	lastDoom   Catastrophe
	legacy     map[string]map[string]float64
	depthW     map[string]int // age key -> depth weight
	depthPts   map[string]int // age key -> depth points
	harbingers []config.HarbingerDef
	harbinger  map[string]config.HarbingerDef // by age key
	awakenings []config.AwakeningDef
	awakening  map[string]config.AwakeningDef // by trigger age

	buildings     []config.BuildingDef
	buildingByKey map[string]config.BuildingDef
	techs         []config.TechDef
	techByKey     map[string]config.TechDef
	techsByAge    map[string][]config.TechDef
	resources     []config.ResourceDef
	resourceByKey map[string]config.ResourceDef

	milestones     []config.MilestoneDef
	milestoneByKey map[string]config.MilestoneDef
	chains         []config.MilestoneChainDef
	chainByKey     map[string]config.MilestoneChainDef
	titles         []config.TitleDef

	events        []config.EventDef
	eraEvents     []config.EventDef
	eventByKey    map[string]config.EventDef
	goodEvents    []config.EpochEventDef
	badEvents     []config.EpochEventDef
	eraEventByKey map[string]config.EpochEventDef

	factions     []config.FactionDef
	factionByKey map[string]config.FactionDef
	routes       []config.TradeRouteDef
	routeByKey   map[string]config.TradeRouteDef
	exchange     []config.ExchangeRateDef
	exchangeBy   map[string]config.ExchangeRateDef // "from:to"

	classes       []config.WorkerClassDef
	classByDomain map[string][]config.WorkerClassDef
	domains       []string

	upgrades     []config.PrestigeUpgradeDef
	upgradeByKey map[string]config.PrestigeUpgradeDef
	shop         []config.PrestigeUpgradeDef // the upgrades the shop sells
	kit          []string

	targets     map[string]time.Duration
	targetTicks map[string]float64
	stretch     map[string]float64

	// Worked out from the buildings and techs: age -> resource -> value.
	priceLevels map[string]map[string]float64
	priced      map[string][]string // age -> its construction resources, sorted
	flowLevels  map[string]map[string]float64
	flowIncome  map[string]map[string]float64
	typIncome   map[string]map[string]float64

	names  [numKinds]map[string]string
	counts map[string]int
}

// Compile builds a Set from src: it copies the tables and works out every
// index and derived number once, so no lookup rebuilds anything.
func Compile(src Source) *Set {
	s := &Set{
		ages:       slices.Clone(src.Ages),
		eras:       slices.Clone(src.Eras),
		buildings:  slices.Clone(src.Buildings),
		techs:      slices.Clone(src.Techs),
		resources:  slices.Clone(src.Resources),
		milestones: slices.Clone(src.Milestones),
		chains:     slices.Clone(src.MilestoneChains),
		titles:     slices.Clone(src.MilestoneTitles),
		events:     slices.Clone(src.Events),
		eraEvents:  slices.Clone(src.EraEvents),
		goodEvents: slices.Clone(src.GoodEraEvents),
		badEvents:  slices.Clone(src.ChallengingEraEvents),
		awakenings: slices.Clone(src.Awakenings),
		harbingers: slices.Clone(src.Harbingers),
		factions:   slices.Clone(src.Factions),
		routes:     slices.Clone(src.TradeRoutes),
		exchange:   slices.Clone(src.ExchangeRates),
		classes:    slices.Clone(src.WorkerClasses),
		domains:    slices.Clone(src.WorkerDomains),
		upgrades:   slices.Clone(src.PrestigeUpgrades),
		kit:        slices.Clone(src.LegacyKit),
		targets:    maps.Clone(src.Targets),
		stretch:    maps.Clone(src.Stretch),
		dooms:      maps.Clone(src.Catastrophes),
		unknown:    src.UnknownCatastrophe,
		lastDoom:   src.LastPassage,
		legacy:     make(map[string]map[string]float64, len(src.Legacy)),
	}
	for era, bonus := range src.Legacy {
		s.legacy[era] = maps.Clone(bonus)
	}
	s.indexAges()
	s.indexEras(src.CatastropheGate)
	s.indexDefs()
	s.derive()
	s.buildNames()
	s.buildCounts()
	return s
}

func (s *Set) indexAges() {
	s.ageKeys = make([]string, len(s.ages))
	s.agePos = make(map[string]int, len(s.ages))
	s.ageByKey = make(map[string]config.AgeDef, len(s.ages))
	for i, a := range s.ages {
		s.ageKeys[i] = a.Key
		s.agePos[a.Key] = i
		s.ageByKey[a.Key] = a
	}
	s.targetTicks = make(map[string]float64, len(s.targets))
	for age, d := range s.targets {
		s.targetTicks[age] = d.Seconds() / config.TickSeconds
	}
}

func (s *Set) indexEras(gate string) {
	s.eraByKey = make(map[string]config.EpochDef, len(s.eras))
	s.eraPos = make(map[string]int, len(s.eras))
	s.eraOfAge = map[string]string{}
	s.eraFirst = map[string]int{}
	s.depthW = map[string]int{}
	for i, e := range s.eras {
		s.eraByKey[e.Key] = e
		s.eraPos[e.Key] = i
		weight := 1
		for n := 0; n < e.Order; n++ {
			weight *= 3
		}
		for _, a := range e.Ages {
			// An age listed by two eras belongs to the first.
			if _, taken := s.eraOfAge[a]; !taken {
				s.eraOfAge[a] = e.Key
				s.depthW[a] = weight
			}
		}
		if len(e.Ages) > 0 {
			s.eraFirst[e.Key] = s.agePos[e.Ages[0]]
		}
	}
	s.gateOrder = s.eraByKey[gate].Order
	s.depthPts = make(map[string]int, len(s.ageKeys))
	total := 0
	for _, a := range s.ageKeys {
		// An age named twice pays what its first place in the order pays.
		if _, seen := s.depthPts[a]; !seen {
			s.depthPts[a] = total
		}
		total += s.depthW[a]
	}
	s.harbinger = make(map[string]config.HarbingerDef, len(s.harbingers))
	for _, h := range s.harbingers {
		s.harbinger[h.Age] = h
	}
	s.awakening = make(map[string]config.AwakeningDef, len(s.awakenings))
	for _, a := range s.awakenings {
		if _, taken := s.awakening[a.TriggerAge]; !taken {
			s.awakening[a.TriggerAge] = a
		}
	}
}

func (s *Set) indexDefs() {
	s.buildingByKey = make(map[string]config.BuildingDef, len(s.buildings))
	for _, b := range s.buildings {
		s.buildingByKey[b.Key] = b
	}
	s.wonders = map[string]string{}
	for _, a := range s.ages {
		for _, key := range a.UnlockBuildings {
			if def, ok := s.buildingByKey[key]; ok && def.Category == "wonder" {
				s.wonders[a.Key] = key
				break
			}
		}
	}
	s.techByKey = make(map[string]config.TechDef, len(s.techs))
	s.techsByAge = map[string][]config.TechDef{}
	for _, t := range s.techs {
		s.techByKey[t.Key] = t
		s.techsByAge[t.Age] = append(s.techsByAge[t.Age], t)
	}
	s.resourceByKey = make(map[string]config.ResourceDef, len(s.resources))
	for _, r := range s.resources {
		s.resourceByKey[r.Key] = r
	}
	s.milestoneByKey = make(map[string]config.MilestoneDef, len(s.milestones))
	for _, m := range s.milestones {
		s.milestoneByKey[m.Key] = m
	}
	s.chainByKey = make(map[string]config.MilestoneChainDef, len(s.chains))
	for _, c := range s.chains {
		s.chainByKey[c.Key] = c
	}
	s.eventByKey = make(map[string]config.EventDef, len(s.events)+len(s.eraEvents))
	for _, e := range s.events {
		s.eventByKey[e.Key] = e
	}
	for _, e := range s.eraEvents {
		s.eventByKey[e.Key] = e
	}
	s.eraEventByKey = make(map[string]config.EpochEventDef, len(s.goodEvents)+len(s.badEvents))
	for _, e := range s.goodEvents {
		s.eraEventByKey[e.Key] = e
	}
	for _, e := range s.badEvents {
		s.eraEventByKey[e.Key] = e
	}
	s.factionByKey = make(map[string]config.FactionDef, len(s.factions))
	for _, f := range s.factions {
		s.factionByKey[f.Key] = f
	}
	s.routeByKey = make(map[string]config.TradeRouteDef, len(s.routes))
	for _, r := range s.routes {
		s.routeByKey[r.Key] = r
	}
	s.exchangeBy = make(map[string]config.ExchangeRateDef, len(s.exchange))
	for _, x := range s.exchange {
		s.exchangeBy[x.From+":"+x.To] = x
	}
	s.classByDomain = map[string][]config.WorkerClassDef{}
	for _, c := range s.classes {
		s.classByDomain[c.Domain] = append(s.classByDomain[c.Domain], c)
	}
	s.upgradeByKey = make(map[string]config.PrestigeUpgradeDef, len(s.upgrades))
	for _, u := range s.upgrades {
		s.upgradeByKey[u.Key] = u
		if !u.Retired {
			s.shop = append(s.shop, u)
		}
	}
}

// ===== Ages =====

// Ages returns the ages in order.
func (s *Set) Ages() []config.AgeDef { return s.ages }

// AgeKeys returns the age keys in order.
func (s *Set) AgeKeys() []string { return s.ageKeys }

// Age returns an age's definition.
func (s *Set) Age(key string) (config.AgeDef, bool) {
	a, ok := s.ageByKey[key]
	return a, ok
}

// Index is an age's place in the order, from 0.
func (s *Set) Index(age string) (int, bool) {
	i, ok := s.agePos[age]
	return i, ok
}

// Next returns the age after age, or "" for the last age and an unknown one.
func (s *Set) Next(age string) string {
	i, ok := s.agePos[age]
	if !ok || i >= len(s.ageKeys)-1 {
		return ""
	}
	return s.ageKeys[i+1]
}

// Wonder returns the key of the wonder age unlocks ("" when it has none).
func (s *Set) Wonder(age string) string { return s.wonders[age] }

// ===== Eras =====

// Eras returns the eras in order.
func (s *Set) Eras() []config.EpochDef { return s.eras }

// Era returns an era's definition by its key.
func (s *Set) Era(key string) (config.EpochDef, bool) {
	e, ok := s.eraByKey[key]
	return e, ok
}

// EraOf returns the key of the era age belongs to. An unknown age reads as
// the first era ("" in a set with no eras).
func (s *Set) EraOf(age string) string {
	if e, ok := s.eraOfAge[age]; ok {
		return e
	}
	if len(s.eras) > 0 {
		return s.eras[0].Key
	}
	return ""
}

// EraFirstAge is the place in the age order of an era's first age.
func (s *Set) EraFirstAge(era string) (int, bool) {
	i, ok := s.eraFirst[era]
	return i, ok
}

// NextEra returns the era after era; ok is false for the final era and an
// unknown one.
func (s *Set) NextEra(era string) (config.EpochDef, bool) {
	i, ok := s.eraPos[era]
	if !ok || i+1 >= len(s.eras) {
		return config.EpochDef{}, false
	}
	return s.eras[i+1], true
}

// IsFinalEra reports whether era is the last one, whose passage is prestige.
// An unknown key is not final.
func (s *Set) IsFinalEra(era string) bool {
	i, ok := s.eraPos[era]
	return ok && i+1 >= len(s.eras)
}

// CatastropheAllowed reports whether a catastrophe may strike in era: from
// the gate era on. An unknown key is not allowed.
func (s *Set) CatastropheAllowed(era string) bool {
	e, ok := s.eraByKey[era]
	return ok && e.Order >= s.gateOrder
}

// FateAllowed reports whether a doom can be fated inside era.
func (s *Set) FateAllowed(era string) bool { return s.CatastropheAllowed(era) }

// Catastrophe returns the name and flavor of an era's doom, and the
// stand-in an unknown key reads as.
func (s *Set) Catastrophe(era string) (name, flavor string) {
	if c, ok := s.dooms[era]; ok {
		return c.Name, c.Flavor
	}
	return s.unknown.Name, s.unknown.Flavor
}

// LastPassage returns the name and flavor of the final era's passage.
func (s *Set) LastPassage() (name, flavor string) { return s.lastDoom.Name, s.lastDoom.Flavor }

// LegacyBonus returns what succumbing in era leaves behind: resource key ->
// share of production. nil for an era that leaves nothing.
func (s *Set) LegacyBonus(era string) map[string]float64 { return s.legacy[era] }

// DepthWeight is what completing age pays at prestige (0 for an unknown age).
func (s *Set) DepthWeight(age string) int { return s.depthW[age] }

// DepthPoints is what a prestige from age pays: the depth weight of every
// age before it (0 for an unknown age).
func (s *Set) DepthPoints(age string) int { return s.depthPts[age] }

// Harbingers returns the roster in age order.
func (s *Set) Harbingers() []config.HarbingerDef { return s.harbingers }

// Harbinger returns the harbinger of an age.
func (s *Set) Harbinger(age string) (config.HarbingerDef, bool) {
	h, ok := s.harbinger[age]
	return h, ok
}

// Awakening returns the awakening that fires on entering age.
func (s *Set) Awakening(age string) (config.AwakeningDef, bool) {
	a, ok := s.awakening[age]
	return a, ok
}

// ===== Buildings, techs, resources =====

// Buildings returns every building, in definition order.
func (s *Set) Buildings() []config.BuildingDef { return s.buildings }

// Building returns a building's definition.
func (s *Set) Building(key string) (config.BuildingDef, bool) {
	b, ok := s.buildingByKey[key]
	return b, ok
}

// BuildingMap returns a fresh map of every building by key.
func (s *Set) BuildingMap() map[string]config.BuildingDef { return maps.Clone(s.buildingByKey) }

// NextTier returns the building one tier up lineage from tier that newAge
// unlocks: what a building of that tier becomes when the age turns.
func (s *Set) NextTier(lineage string, tier int, newAge string) (config.BuildingDef, bool) {
	for _, b := range s.buildings {
		if b.LineageKey == lineage && b.LineageTier == tier+1 && b.RequiredAge == newAge {
			return b, true
		}
	}
	return config.BuildingDef{}, false
}

// AgeEntryCosts returns, per resource, the cheapest first-copy price of it
// among the buildings age unlocks, wonders aside. The map is the caller's.
func (s *Set) AgeEntryCosts(age string) map[string]float64 {
	out := make(map[string]float64)
	for _, b := range s.buildings {
		if b.RequiredAge != age || b.Category == "wonder" {
			continue
		}
		for res, amt := range b.BaseCost {
			if amt <= 0 {
				continue
			}
			if cur, ok := out[res]; !ok || amt < cur {
				out[res] = amt
			}
		}
	}
	return out
}

// Techs returns every tech, in definition order.
func (s *Set) Techs() []config.TechDef { return s.techs }

// Tech returns a tech's definition.
func (s *Set) Tech(key string) (config.TechDef, bool) {
	t, ok := s.techByKey[key]
	return t, ok
}

// TechMap returns a fresh map of every tech by key.
func (s *Set) TechMap() map[string]config.TechDef { return maps.Clone(s.techByKey) }

// TechsOf returns the techs of an age, in definition order.
func (s *Set) TechsOf(age string) []config.TechDef { return s.techsByAge[age] }

// Resources returns every resource, in definition order.
func (s *Set) Resources() []config.ResourceDef { return s.resources }

// Resource returns a resource's definition.
func (s *Set) Resource(key string) (config.ResourceDef, bool) {
	r, ok := s.resourceByKey[key]
	return r, ok
}

// ResourceMap returns a fresh map of every resource by key.
func (s *Set) ResourceMap() map[string]config.ResourceDef { return maps.Clone(s.resourceByKey) }

// ResourceLabel is a resource's name in running text: "iron ore", "dark
// matter". An unknown key falls back to the key with spaces.
func (s *Set) ResourceLabel(key string) string {
	if r, ok := s.resourceByKey[key]; ok {
		return strings.ToLower(r.Name)
	}
	return strings.ReplaceAll(key, "_", " ")
}

// ===== Milestones and events =====

// Milestones returns every milestone, in definition order.
func (s *Set) Milestones() []config.MilestoneDef { return s.milestones }

// Milestone returns a milestone's definition.
func (s *Set) Milestone(key string) (config.MilestoneDef, bool) {
	m, ok := s.milestoneByKey[key]
	return m, ok
}

// MilestoneChains returns every milestone chain, in definition order.
func (s *Set) MilestoneChains() []config.MilestoneChainDef { return s.chains }

// MilestoneChain returns a chain's definition.
func (s *Set) MilestoneChain(key string) (config.MilestoneChainDef, bool) {
	c, ok := s.chainByKey[key]
	return c, ok
}

// MilestoneTitles returns the fallback title ladder.
func (s *Set) MilestoneTitles() []config.TitleDef { return s.titles }

// Events returns the random event pool.
func (s *Set) Events() []config.EventDef { return s.events }

// EraEvents returns the events only one era can roll.
func (s *Set) EraEvents() []config.EventDef { return s.eraEvents }

// EventMap returns a fresh map of every event by key, the era ones included.
func (s *Set) EventMap() map[string]config.EventDef { return maps.Clone(s.eventByKey) }

// GoodEraEvents returns the good events an era's entry can roll.
func (s *Set) GoodEraEvents() []config.EpochEventDef { return s.goodEvents }

// ChallengingEraEvents returns the challenging events an era's entry can roll.
func (s *Set) ChallengingEraEvents() []config.EpochEventDef { return s.badEvents }

// EraEvent returns an era-entry event by key, from either pool.
func (s *Set) EraEvent(key string) (config.EpochEventDef, bool) {
	e, ok := s.eraEventByKey[key]
	return e, ok
}

// ===== Civilizations, trade, workers, the shop =====

// Factions returns every civilization, in definition order.
func (s *Set) Factions() []config.FactionDef { return s.factions }

// Faction returns a civilization's definition.
func (s *Set) Faction(key string) (config.FactionDef, bool) {
	f, ok := s.factionByKey[key]
	return f, ok
}

// TradeRoutes returns every trade route, in definition order.
func (s *Set) TradeRoutes() []config.TradeRouteDef { return s.routes }

// TradeRoute returns a trade route's definition.
func (s *Set) TradeRoute(key string) (config.TradeRouteDef, bool) {
	r, ok := s.routeByKey[key]
	return r, ok
}

// ExchangeRates returns the listed market pairs, in definition order.
func (s *Set) ExchangeRates() []config.ExchangeRateDef { return s.exchange }

// ExchangeRateMap returns a fresh map of the listed market pairs, keyed
// "from:to".
func (s *Set) ExchangeRateMap() map[string]config.ExchangeRateDef { return maps.Clone(s.exchangeBy) }

// WorkerDomains returns the worker domain keys in their fixed order.
func (s *Set) WorkerDomains() []string { return s.domains }

// WorkerClass returns the class a domain's workers hold in age: the class
// written for that age, else the latest one from an earlier age, else the
// domain's first. An unknown age reads as later than every age.
func (s *Set) WorkerClass(domain, age string) (config.WorkerClassDef, bool) {
	classes := s.classByDomain[domain]
	for _, c := range classes {
		if c.AgeKey == age {
			return c, true
		}
	}
	current := 1 << 30
	if a, ok := s.ageByKey[age]; ok {
		current = a.Order
	}
	best, bestOrder := -1, -1
	for i, c := range classes {
		if a, ok := s.ageByKey[c.AgeKey]; ok && a.Order <= current && a.Order > bestOrder {
			best, bestOrder = i, a.Order
		}
	}
	if best >= 0 {
		return classes[best], true
	}
	if len(classes) > 0 {
		return classes[0], true
	}
	return config.WorkerClassDef{}, false
}

// PrestigeUpgrades returns the whole shop table, retired perks included.
func (s *Set) PrestigeUpgrades() []config.PrestigeUpgradeDef { return s.upgrades }

// ShopUpgrades returns the upgrades the shop sells, in shop order.
func (s *Set) ShopUpgrades() []config.PrestigeUpgradeDef { return s.shop }

// PrestigeUpgrade returns a shop upgrade's definition.
func (s *Set) PrestigeUpgrade(key string) (config.PrestigeUpgradeDef, bool) {
	u, ok := s.upgradeByKey[key]
	return u, ok
}

// LegacyKit returns the legacy kit's keys in shop order.
func (s *Set) LegacyKit() []string { return s.kit }
