package game

import (
	"math"

	"github.com/espresso20/ageforge/config"
)

// feats.go judges the legendary feat of each age, one hard thing that strikes
// the age's feat coin. A feat is judged only while the run is in its age and
// does not yet hold that age's feat coin, either on an event (judgeFeatOn,
// from strikeOn) or on leaving the age (judgeFeatLeaving, from advanceAge).
// Both run under the engine's write lock and read the game without changing it.

// strikeFeat strikes the feat coin of age for this run. Callers hold ge.mu for writing.
func (ge *GameEngine) strikeFeat(age string) {
	acct := ge.accountForRecordsLocked()
	if acct == nil {
		return
	}
	entered, _ := ge.ageEntryTick(age)
	acct.StrikeFeat(age, max(0, ge.tick-entered), ge.activeSaveName, badgeClock(), ge.cheaterBadge)
}

// ticksIn is the ticks the run has spent in age so far.
func (ge *GameEngine) ticksIn(age string) float64 {
	entered, _ := ge.ageEntryTick(age)
	return float64(max(0, ge.tick-entered))
}

// within reports whether the run has spent fewer ticks than share of age's
// reference time at the run's speed for it.
func (ge *GameEngine) within(age string, share float64) bool {
	return ge.ticksIn(age) < float64(share*ge.rules.CoinRefTicks(age, ge.Prestige.AgeSpeed(age)))
}

// ageTechsDone reports whether every tech of age is researched.
func (ge *GameEngine) ageTechsDone(age string) bool {
	techs := ge.badges.ageTechs[age]
	for _, k := range techs {
		if !ge.Research.IsResearched(k) {
			return false
		}
	}
	return len(techs) > 0
}

// staffedIn is the workers posted in the jobs of domain and in all jobs. The
// soldiers are not counted: they hold posts no building has.
func (ge *GameEngine) staffedIn(domain string) (of, all int) {
	for building, n := range ge.Workers.assignments {
		if def, ok := ge.Workers.buildingDefs[building]; ok && def.WorkerDomain != "" {
			all += n
			if def.WorkerDomain == domain {
				of += n
			}
		}
	}
	return of, all
}

// wonderOf reports whether key is a wonder of age.
func (ge *GameEngine) wonderOf(key, age string) bool {
	def, ok := ge.rules.Building(key)
	return ok && def.Category == "wonder" && def.RequiredAge == age
}

// judgeFeatOn judges the feats an event decides, for the age the run is in.
func (ge *GameEngine) judgeFeatOn(ev Event) {
	age := ge.age
	if age == "" || ge.coinHeldLocked(config.CoinFeat, age) {
		return
	}
	mark, d, done := ge.rules.FeatMark(age), ge.Stats.Deeds, false
	switch age {
	case "primitive_age":
		done = ev.Kind == config.BadgeEvBuildingBuilt && ge.Buildings.GetCount("hut") >= int(mark)
	case "iron_age":
		entered, _ := ge.ageEntryTick(age)
		done = ev.Kind == config.BadgeEvResearchDone && ev.Subject == ge.rules.Keystone(age) &&
			ge.Stats.Deeds.KeystoneTick > 0 &&
			float64(ge.Stats.Deeds.KeystoneTick-entered) <= mark/ge.Prestige.AgeSpeed(age)
	case "classical_age":
		done = ev.Kind == config.BadgeEvFestival && ge.Stats.Deeds.Festivals >= max(1, int(math.Ceil(mark/ge.Prestige.AgeSpeed(age))))
	case "renaissance_age":
		done = ev.Kind == config.BadgeEvResearchDone && ge.ageTechsDone(age)
	case "medieval_age":
		_, began := d.Wars[ev.Subject]
		done = ev.Kind == config.BadgeEvWarEnded && began && ev.Attrs["tribute"] == 0 && d.Tributes == 0
	case "colonial_age":
		done = ev.Kind == config.BadgeEvTick && len(d.RouteClean) >= int(mark)
	case "industrial_age":
		// On the census cadence, from the workers' posts (the census event
		// only totals the domains some badge asks for).
		if ev.Kind == config.BadgeEvTick && ge.tick%config.BadgeCensusTicks == 0 {
			of, all := ge.staffedIn("energy")
			done = all > 0 && float64(of) >= float64(mark*float64(all))
		}
	case "victorian_age":
		done = ev.Kind == config.BadgeEvWonderRaised && ge.wonderOf(ev.Subject, age) && ge.within(age, mark)
	case "information_age":
		done = ev.Kind == config.BadgeEvMilestone && ev.Subject == "tech_master"
	case "digital_age":
		done = ev.Kind == config.BadgeEvWonderRaised && ge.wonderOf(ev.Subject, age) && ge.Diplomacy.warCount() > 0
	case "fusion_age":
		done = ev.Kind == config.BadgeEvCivAllied && int(ev.Attrs["met"]-ev.Attrs["unallied"]) >= int(mark) && ge.Diplomacy.warCount() == 0
	case "galactic_age":
		h := ge.harbinger
		done = ev.Kind == config.BadgeEvAppeased && h != nil && h.TargetEpoch == "" && h.AppeaseLevel >= int(mark)
	case "quantum_age":
		h := ge.harbinger
		done = (ev.Kind == config.BadgeEvInvited || ev.Kind == config.BadgeEvBraced) &&
			h != nil && h.TargetEpoch == "" && h.Invited && h.BraceLevel >= int(mark)
	case "transcendent_age":
		done = ev.Kind == config.BadgeEvLastPassage && ev.Subject == "endured" && int(ev.Attrs["brace"]) >= int(mark)
	}
	if done {
		ge.strikeFeat(age)
	}
}

// judgeFeatLeaving judges the feats decided on leaving age, before anything of
// the advance has changed the run: the stock, the age and the deeds are still
// the age's own.
func (ge *GameEngine) judgeFeatLeaving(age string) {
	if age == "" || ge.coinHeldLocked(config.CoinFeat, age) {
		return
	}
	mark, d, done := ge.rules.FeatMark(age), ge.Stats.Deeds, false
	switch age {
	case "stone_age":
		done = ge.within(age, mark)
	case "bronze_age":
		done = d.Upgrades == 0 && ge.within(age, mark)
	case "electric_age":
		plant, _ := ge.rules.Building("nuclear_plant")
		need := float64(mark * plant.BaseCost["electricity"])
		done = need > 0 && ge.Resources.Get("electricity") >= need
	case "atomic_age":
		done = d.Pacts == 0 && ge.within(age, mark)
	case "modern_age":
		sum := 0.0
		for _, a := range ge.rules.AgeKeys() {
			sum += ge.rules.CoinRefTicks(a, ge.Prestige.AgeSpeed(a))
			if a == age {
				break
			}
		}
		done = float64(ge.tick) < float64(mark*sum)
	case "cyberpunk_age":
		fs, _ := ge.Diplomacy.StateOf("shadow_syndicate")
		done = fs.Status == "allied"
	case "space_age":
		done = FaithBandAt(ge.faithStrength()) == FaithBandHigh
	case "interstellar_age":
		if start, ok := d.RouteClean["warp_commerce"]; ok {
			done = float64(ge.tick-start) >= float64(mark*ge.ticksIn(age))
		}
	}
	if done {
		ge.strikeFeat(age)
	}
}
