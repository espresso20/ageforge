// Package rules holds a ruleset: the game's definitions and every table
// worked out from them, compiled once into one immutable value.
//
// An engine owns one Set and reads everything about its ages, buildings,
// techs, eras and prices from it. Today every engine is built on the core
// set, which Compile makes from package config's tables, so the numbers are
// the ones config gives. What changes is where they live: on the engine,
// not in tables shared by the whole process, so two engines can run side by
// side on different rules.
//
// rules imports config. config must never import rules.
package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Source is the flat definitions a Set is compiled from: the tables package
// config builds, already normalized (prices, rates, time caps). Compile
// copies every slice and map it keeps, so a Source can be changed and
// compiled again.
type Source struct {
	// Ages in order, and the eras that group them, in order.
	Ages []config.AgeDef
	Eras []config.EpochDef

	Buildings []config.BuildingDef
	Techs     []config.TechDef
	Resources []config.ResourceDef

	Milestones      []config.MilestoneDef
	MilestoneChains []config.MilestoneChainDef
	MilestoneTitles []config.TitleDef

	// Events is the random pool; EraEvents the ones only one era rolls.
	// GoodEraEvents and ChallengingEraEvents are the pools an era's entry
	// rolls from. Awakenings fire once, on entering their age.
	Events               []config.EventDef
	EraEvents            []config.EventDef
	GoodEraEvents        []config.EpochEventDef
	ChallengingEraEvents []config.EpochEventDef
	Awakenings           []config.AwakeningDef

	// Harbingers is the roster, one per age, derived fields filled in.
	Harbingers []config.HarbingerDef

	Factions      []config.FactionDef
	TradeRoutes   []config.TradeRouteDef
	ExchangeRates []config.ExchangeRateDef

	WorkerClasses []config.WorkerClassDef
	WorkerDomains []string

	// PrestigeUpgrades is the whole shop, retired perks included; LegacyKit
	// the kit's keys in shop order.
	PrestigeUpgrades []config.PrestigeUpgradeDef
	LegacyKit        []string

	// Targets is the time a player should spend in each age at 1x, and
	// Stretch the factor each age's tick clocks run at (an age it leaves
	// out runs at 1).
	Targets map[string]time.Duration
	Stretch map[string]float64

	// Catastrophes is each era's doom, by era key; UnknownCatastrophe what
	// an unknown key reads as; LastPassage the final era's passage. Legacy
	// is what succumbing in each era leaves: resource key -> share.
	// CatastropheGate is the first era a catastrophe can strike in.
	Catastrophes       map[string]Catastrophe
	UnknownCatastrophe Catastrophe
	LastPassage        Catastrophe
	Legacy             map[string]map[string]float64
	CatastropheGate    string
}

// Catastrophe is a doom's display name and its flavor line.
type Catastrophe struct {
	Name   string
	Flavor string
}

// FromConfig is today's game: every table package config defines, freshly
// built, so the caller may change what it gets.
func FromConfig() Source {
	src := Source{
		Ages:                 config.Ages(),
		Eras:                 config.Epochs(),
		Buildings:            config.BaseBuildings(),
		Techs:                config.Technologies(),
		Resources:            config.BaseResources(),
		Milestones:           config.Milestones(),
		MilestoneChains:      config.MilestoneChains(),
		MilestoneTitles:      config.MilestoneTitles(),
		Events:               config.RandomEvents(),
		EraEvents:            config.EpochExclusiveEvents(),
		GoodEraEvents:        config.GoodEpochEvents(),
		ChallengingEraEvents: config.ChallengingEpochEvents(),
		Awakenings:           config.Awakenings(),
		Harbingers:           config.Harbingers(),
		Factions:             config.BaseFactions(),
		TradeRoutes:          config.BaseTradeRoutes(),
		ExchangeRates:        config.BaseExchangeRates(),
		WorkerClasses:        config.WorkerClasses(),
		WorkerDomains:        config.WorkerDomains(),
		PrestigeUpgrades:     config.PrestigeUpgrades(),
		LegacyKit:            config.LegacyKit(),
		Targets:              maps.Clone(config.AgeTargets),
		Stretch:              map[string]float64{},
		Catastrophes:         map[string]Catastrophe{},
		Legacy:               map[string]map[string]float64{},
		CatastropheGate:      config.CatastropheGateEpoch,
	}
	for _, a := range src.Ages {
		src.Stretch[a.Key] = config.AgeStretch(a.Key)
	}
	for _, e := range src.Eras {
		name, flavor := config.CatastropheInfo(e.Key)
		src.Catastrophes[e.Key] = Catastrophe{Name: name, Flavor: flavor}
		if bonus := config.LegacyBonusForEpoch(e.Key); bonus != nil {
			src.Legacy[e.Key] = bonus
		}
	}
	name, flavor := config.CatastropheInfo("")
	src.UnknownCatastrophe = Catastrophe{Name: name, Flavor: flavor}
	name, flavor = config.LastPassageInfo()
	src.LastPassage = Catastrophe{Name: name, Flavor: flavor}
	return src
}

// core is the one set built from package config: the process's only
// once-built table of definitions. Every engine that is not handed a set of
// its own shares it.
var core = sync.OnceValue(func() *Set { return Compile(FromConfig()) })

// Core returns the set compiled from package config, built on first use and
// shared from then on. It is what an engine runs on unless it is given
// another, and what code with no engine or snapshot to hand reads.
func Core() *Set { return core() }

// Digest is a hash of everything the set holds. A set never changes, so its
// digest never does: tests take one before and after a run to prove that
// nothing wrote into the set. Two sets compiled from the same tables can
// differ in the last bit of a flow price level (config.FlowDealLevels sums
// over a map), so compare a set with itself, not with another compile.
func (s *Set) Digest() string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%v", *s)))
	return hex.EncodeToString(sum[:8])
}
