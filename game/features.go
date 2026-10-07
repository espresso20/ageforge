package game

import (
	"errors"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// Feature locks (config.FeatureLocks): a command that waits for a tech.
//
// A locked command is refused with the name of the tech that opens it. It is
// open when any of these holds:
//
//   - its tech is not in the tree (the lock is inert until a content change
//     adds the tech);
//   - its tech is researched;
//   - this run already used it (grantedFeatures): a command in use stays
//     open for the rest of the run, whatever is researched;
//   - the game is in its tree grace age (a save from before the lock).
//
// Every use of a command whose tech is not researched is recorded as granted,
// an inert lock's too. So when a later version adds the tech an inert lock
// names, a game already trading or holding festivals keeps doing so for that
// run, and meets the lock on its next one.

// FeatureState is one feature lock as the panels and the smoke bot read it.
type FeatureState struct {
	Name string
	// Tech and TechName are the tech that opens the command.
	Tech     string
	TechName string
	// Live reports that the lock's tech is in the tree; a lock that is not
	// live is always open.
	Live bool
	// Open reports that the command can be used now.
	Open bool
	// Granted reports that this run keeps the command without the tech.
	Granted bool
}

// featureLock returns the lock on feature key and whether it holds the
// command shut right now. An unknown key is not locked. Caller holds the
// lock.
func (ge *GameEngine) featureLock(key string) (config.FeatureLockDef, bool) {
	def, ok := ge.rules.FeatureLock(key)
	if !ok {
		return def, false
	}
	return def, ge.featureShut(def)
}

// featureShut reports whether def holds its command shut. Caller holds the
// lock.
func (ge *GameEngine) featureShut(def config.FeatureLockDef) bool {
	if !ge.rules.FeatureLockLive(def) {
		return false
	}
	if ge.Research.IsResearched(def.Tech) || ge.grantedFeatures[def.Key] {
		return false
	}
	return !ge.inTreeGrace()
}

// inTreeGrace reports whether the game is in its tree grace age: the age an
// older save was in when it met the tree's new locks. Caller holds the lock.
func (ge *GameEngine) inTreeGrace() bool {
	return ge.Buildings.graceAge != "" && ge.Buildings.graceAge == ge.age
}

// featureErr is the refusal for the first of keys that is locked ("" keys are
// skipped), nil when every one is open: "Campaigns need Military Tactics
// first. Research it to send one." Caller holds the lock.
func (ge *GameEngine) featureErr(keys ...string) error {
	for _, key := range keys {
		if key == "" {
			continue
		}
		if def, shut := ge.featureLock(key); shut {
			tech, _ := ge.rules.Tech(def.Tech)
			return errors.New(def.Refusal(tech.Name))
		}
	}
	return nil
}

// useFeature records that the run used the commands behind keys: each one
// whose tech is not researched is granted for the rest of the run. Call it
// once the command has gone through. Caller holds the write lock.
func (ge *GameEngine) useFeature(keys ...string) {
	for _, key := range keys {
		def, ok := ge.rules.FeatureLock(key)
		if !ok || ge.Research.IsResearched(def.Tech) || ge.grantedFeatures[key] {
			continue
		}
		if ge.grantedFeatures == nil {
			ge.grantedFeatures = make(map[string]bool)
		}
		ge.grantedFeatures[key] = true
	}
}

// expeditionFeatures is the locks an expedition sits behind: campaigns for a
// military one; for a scouting one past the Scout Party the expeditions
// lock, and the Naval Expedition its own on top.
func expeditionFeatures(def *ExpeditionDef) []string {
	switch {
	case def.Category == ExpeditionMilitary:
		return []string{config.FeatureCampaigns}
	case def.Key == "scout_party":
		return nil
	case def.Key == "naval_expedition":
		return []string{config.FeatureExpeditions, config.FeatureNavalExpedition}
	}
	return []string{config.FeatureExpeditions}
}

// routeFeatures is the locks a trade route sits behind: trade routes, and
// the route's own where it has one.
func routeFeatures(route string) []string {
	return []string{config.FeatureTradeRoutes, config.RouteFeature(route)}
}

// featureStates is every feature lock as the panels read it. Caller holds
// the lock.
func (ge *GameEngine) featureStates() map[string]FeatureState {
	out := make(map[string]FeatureState)
	for _, def := range ge.rules.FeatureLocks() {
		tech, live := ge.rules.Tech(def.Tech)
		out[def.Key] = FeatureState{
			Name: def.Name, Tech: def.Tech, TechName: tech.Name, Live: live,
			Open: !ge.featureShut(def), Granted: ge.grantedFeatures[def.Key],
		}
	}
	return out
}

// grantedFeatureKeys is the run's granted features for the save, sorted; nil
// when there are none, so the field is left out.
func (ge *GameEngine) grantedFeatureKeys() []string {
	if len(ge.grantedFeatures) == 0 {
		return nil
	}
	keys := make([]string, 0, len(ge.grantedFeatures))
	for k := range ge.grantedFeatures {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// loadGrantedFeatures restores the run's granted features from a save,
// dropping keys that name no lock.
func (ge *GameEngine) loadGrantedFeatures(keys []string) {
	ge.grantedFeatures = nil
	for _, k := range keys {
		if _, ok := ge.rules.FeatureLock(k); !ok {
			continue
		}
		if ge.grantedFeatures == nil {
			ge.grantedFeatures = make(map[string]bool)
		}
		ge.grantedFeatures[k] = true
	}
}

// featuresInUse is the commands a save from before the feature locks shows
// in use, by what it holds: those stay open for the rest of its run.
//
//   - trade routes: a route running, or anything a route ever imported or
//     exported this run; Rail Freight and Warp Commerce when that route is
//     running;
//   - campaigns: a campaign under way;
//   - expeditions past the Scout Party, and the Naval Expedition: that
//     expedition under way;
//   - a finished expedition (the save counts them but not their kind) counts
//     for campaigns and expeditions alike, and for the Naval Expedition once
//     the save has reached the age it sails from;
//   - festivals and the black market: their cooldown has been set this run;
//   - diplomacy: a civilization is met.
func featuresInUse(save *GameSave, ageIndex func(string) (int, bool)) []string {
	var out []string
	routes := save.Trade.ActiveRoutes
	if len(routes) > 0 || len(save.Trade.TotalImported) > 0 || len(save.Trade.TotalExported) > 0 {
		out = append(out, config.FeatureTradeRoutes)
	}
	for _, key := range sortedKeys(routes) {
		if f := config.RouteFeature(key); f != "" {
			out = append(out, f)
		}
	}
	if save.Military.ActiveMilitary != nil {
		out = append(out, config.FeatureCampaigns)
	}
	for _, exp := range []*ActiveExpedition{save.Military.ActiveScout, save.Military.ActiveExpedition} {
		if exp == nil || exp.Key == "scout_party" {
			continue
		}
		out = append(out, config.FeatureExpeditions)
		if exp.Key == "naval_expedition" {
			out = append(out, config.FeatureNavalExpedition)
		}
	}
	if save.Military.CompletedCount > 0 {
		out = append(out, config.FeatureCampaigns, config.FeatureExpeditions)
		here, okHere := ageIndex(save.Age)
		if sails, ok := ageIndex("renaissance_age"); ok && okHere && here >= sails {
			out = append(out, config.FeatureNavalExpedition)
		}
	}
	if save.FestivalReadyTick > 0 {
		out = append(out, config.FeatureFestivals)
	}
	if save.BlackMarketReadyTick > 0 {
		out = append(out, config.FeatureBlackMarket)
	}
	for _, key := range sortedKeys(save.Diplomacy.Factions) {
		if save.Diplomacy.Factions[key].Discovered {
			out = append(out, config.FeatureDiplomacy)
			break
		}
	}
	return out
}
