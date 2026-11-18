package game

import (
	"github.com/espresso20/ageforge/config"
)

// collection_strike.go strikes the coins of the five types every age has, in
// the places the matching badges are reported (report, advanceAge). Like the
// reports, it runs under the engine's write lock, reads the game and changes
// nothing in it: the account's own lock is the only one it takes.

// ageEntryTick is the tick the run entered age at, from the run's age log.
func (ge *GameEngine) ageEntryTick(age string) (int, bool) {
	for _, e := range ge.Stats.AgeLog {
		if e.Age == age {
			return e.Tick, true
		}
	}
	return 0, false
}

// strikeCoin strikes the coin of typ for age, for this run. The figure that
// sets the level is the ticks the run spent in age so far (a run that never
// entered it, a developer jump, counts from its start); the speed is the
// run's own for that age. Callers hold ge.mu for writing.
func (ge *GameEngine) strikeCoin(typ, age string) {
	acct := ge.accountForRecordsLocked()
	if acct == nil || age == "" {
		return
	}
	entered, _ := ge.ageEntryTick(age)
	acct.StrikeCoin(typ, age, max(0, ge.tick-entered), ge.Prestige.AgeSpeed(age), ge.activeSaveName, badgeClock())
}

// coinHeldLocked reports whether this run already holds a coin of typ for
// age, so the checks that read the town are skipped once it has one.
func (ge *GameEngine) coinHeldLocked(typ, age string) bool {
	acct := ge.accountForRecordsLocked()
	return acct == nil || acct.coinHeld(typ, age, ge.activeSaveName)
}

// strikeOn strikes the coin an event completes, if it does: a wonder raised, an
// age's buildings all standing, its techs all researched, its harbinger met.
// The coin belongs to the age of the thing, not the age the run is in.
func (ge *GameEngine) strikeOn(ev Event) {
	if ge.badges == nil {
		return
	}
	switch ev.Kind {
	case config.BadgeEvWonderRaised:
		if def, ok := ge.rules.Building(ev.Subject); ok {
			ge.strikeCoin(config.CoinWonder, def.RequiredAge)
		}
	case config.BadgeEvBuildingBuilt:
		def, ok := ge.rules.Building(ev.Subject)
		if !ok || def.Category == "wonder" || ge.coinHeldLocked(config.CoinMaximalist, def.RequiredAge) {
			return
		}
		// The age's own badge sets the count: the same share of the age's
		// buildings standing.
		i, ok := ge.badges.byKey["maximalist."+def.RequiredAge]
		if !ok {
			return
		}
		n := 0
		for _, k := range ge.badges.ageBuildings[def.RequiredAge] {
			n += ge.Buildings.GetCount(k)
		}
		if float64(n) >= ge.badges.defs[i].Threshold {
			ge.strikeCoin(config.CoinMaximalist, def.RequiredAge)
		}
	case config.BadgeEvResearchDone:
		tech, ok := ge.rules.Tech(ev.Subject)
		if !ok || len(ge.badges.ageTechs[tech.Age]) == 0 || ge.coinHeldLocked(config.CoinTechs, tech.Age) {
			return
		}
		for _, k := range ge.badges.ageTechs[tech.Age] {
			if !ge.Research.IsResearched(k) {
				return
			}
		}
		ge.strikeCoin(config.CoinTechs, tech.Age)
	case config.BadgeEvHarbingerMet:
		for _, age := range ge.rules.AgeKeys() {
			if h, ok := ge.rules.Harbinger(age); ok && h.Key == ev.Subject {
				ge.strikeCoin(config.CoinHarbinger, h.Age)
				return
			}
		}
	}
}
