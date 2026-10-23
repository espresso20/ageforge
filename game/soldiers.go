package game

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Soldiers. A military building comes with its soldiers (config.BuildingDef.
// Soldiers a copy, derived in config/soldier_rule.go), as a knowledge building
// comes with its scholars. They are people: they eat like anyone and take a
// place in the homes (population.go), and a sold or destroyed copy takes its
// soldiers with it.
//
// Posts fill from the reserve first (a person with no job takes a post), then
// from new arrivals, which need no room in the homes. Both only while the food
// margin test the reserve's growth uses passes (recruitFoodMargin): net food
// stays at or above max(one person's food, recruitFoodMarginShare of the food
// made) after the newcomer eats. A post that cannot be filled waits, and the
// staffing pass (fillSoldierPosts) tries again, so soldiers arriving never
// starve the town.
//
// The army is the soldiers at their posts. A mission takes a party out of it
// (military.go, parties): the party leaves its posts while away, so the army
// at home is the posted soldiers less the soldiers away. Everything that reads
// soldiers (the Defense Rating, raids, the Brace view, a party's launch, the
// soldier milestones) reads the "soldiers" resource, and syncSoldiers holds
// that resource at the army at home every tick. Nothing else writes it. Away
// or home, a soldier is a person: he eats and takes a place in the homes.
//
// Soldiers fall: on a mission that fails or costs lives, in an ambush on the
// road, and when a raid lands against the garrison. A fallen soldier leaves the
// population (soldiersFall) and his post is empty. The empty posts of the
// fallen are not filled by arrivals: they train from the reserve, a few at a
// time (soldierTrainShare of the posts a staffing pass), under the same food
// margin. soldiersVacant holds them by building.

// soldierTrainPercent is the share of all the posts that may finish training
// in one staffing pass, at least one soldier.
const soldierTrainPercent = 1

// SoldierPost is one kind of military building's posts for a snapshot.
type SoldierPost struct {
	Key  string
	Name string
	// Count is the copies standing; PerCopy the soldiers one copy comes with.
	Count, PerCopy int
	// Posted is the soldiers at their posts; Posts the posts the copies have.
	// A post that food cannot yet fill is not posted.
	Posted, Posts int
}

// soldierPostsOf is the posts of the copies of key that stand.
func (ge *GameEngine) soldierPostsOf(key string) int {
	bm := ge.Buildings
	if n := bm.counts[key]; n > 0 {
		return n * bm.defs[key].Soldiers
	}
	return 0
}

// soldierTotal is every soldier at his post: the posted soldiers of every
// standing military building, never more than its posts. A plain loop over the
// military keys, as scholarTotal.
func (ge *GameEngine) soldierTotal() int {
	total := 0
	for key, n := range ge.soldiersPosted {
		total += min(n, ge.soldierPostsOf(key))
	}
	return total
}

// soldierPostsTotal is every post the standing military buildings have, filled
// or not.
func (ge *GameEngine) soldierPostsTotal() int {
	total := 0
	for _, key := range ge.Buildings.soldierKeys {
		total += ge.soldierPostsOf(key)
	}
	return total
}

// soldierPosts lists the posts of each standing military building, oldest
// first (sorted by key within a tier).
func (ge *GameEngine) soldierPosts() []SoldierPost {
	var out []SoldierPost
	ge.Buildings.eachBuiltOf(ge.Buildings.soldierKeys, func(key string, count int, def config.BuildingDef) {
		if def.Soldiers <= 0 {
			return
		}
		posts := count * def.Soldiers
		out = append(out, SoldierPost{Key: key, Name: def.Name, Count: count, PerCopy: def.Soldiers, Posted: min(ge.soldiersPosted[key], posts), Posts: posts})
	})
	sort.SliceStable(out, func(i, j int) bool {
		return ge.Buildings.defs[out[i].Key].LineageTier < ge.Buildings.defs[out[j].Key].LineageTier
	})
	return out
}

// soldiersAway is the soldiers out on missions: the parties under way.
func (ge *GameEngine) soldiersAway() int { return ge.Military.SoldiersAway() }

// soldiersHome is the army at home: the posted soldiers less those away.
func (ge *GameEngine) soldiersHome() int {
	posted := ge.soldierTotal()
	if ge.garrisonForTestSet {
		posted = int(ge.garrisonForTest)
	}
	return max(0, posted-ge.soldiersAway())
}

// postKeys is the military buildings that stand, oldest first.
func (ge *GameEngine) postKeys() []string {
	var keys []string
	for _, key := range ge.Buildings.soldierKeys {
		if ge.soldierPostsOf(key) > 0 {
			keys = append(keys, key)
		}
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return ge.Buildings.defs[keys[i]].LineageTier < ge.Buildings.defs[keys[j]].LineageTier
	})
	return keys
}

// soldiersFall takes n soldiers out of the population: they are dead. They
// leave the posts, oldest building first, and each post stays empty until the
// reserve trains someone for it (fillSoldierPosts). It returns how many fell
// (never more than the posted soldiers). Caller holds the write lock.
func (ge *GameEngine) soldiersFall(n int) int {
	if n <= 0 {
		return 0
	}
	if ge.garrisonForTestSet {
		// A garrison set by a test hook has no posts: the army itself shrinks.
		n = min(n, int(ge.garrisonForTest))
		ge.garrisonForTest -= float64(n)
		ge.syncSoldiers()
		return n
	}
	fell := 0
	for _, key := range ge.postKeys() {
		if fell >= n {
			break
		}
		take := min(n-fell, ge.soldiersPosted[key])
		if take <= 0 {
			continue
		}
		ge.soldiersPosted[key] -= take
		if ge.soldiersVacant == nil {
			ge.soldiersVacant = map[string]int{}
		}
		ge.soldiersVacant[key] += take
		fell += take
	}
	if fell > 0 {
		ge.syncSoldiers()
	}
	return fell
}

// trainFallenPosts is the training pass for the posts the fallen left: oldest
// building first, from the reserve only, at most soldierTrainPercent of all
// the posts (at least one soldier) a pass, and only while the food margin the
// reserve's growth uses passes. It returns how many finished training.
func (ge *GameEngine) trainFallenPosts() int {
	if len(ge.soldiersVacant) == 0 {
		return 0
	}
	if ge.soldiersPosted == nil {
		ge.soldiersPosted = map[string]int{}
	}
	budget := max(1, (soldierTrainPercent*ge.soldierPostsTotal()+99)/100)
	net, per, margin := ge.recruitFood()
	if ge.Resources.Get("food") <= 0 || net-per < margin {
		return 0
	}
	total := 0
	for _, key := range ge.postKeys() {
		vacant := ge.soldiersVacant[key]
		if vacant <= 0 || budget <= 0 {
			continue
		}
		got := ge.Workers.TakeIdle(min(vacant, budget))
		if got <= 0 {
			continue
		}
		budget -= got
		total += got
		ge.soldiersPosted[key] += got
		ge.soldiersVacant[key] -= got
		if ge.soldiersVacant[key] <= 0 {
			delete(ge.soldiersVacant, key)
		}
		ge.logSoldiersTrained(ge.Buildings.defs[key], got)
	}
	return total
}

// logSoldiersTrained says that n soldiers have finished training for a post
// at a military building. Caller holds the lock.
func (ge *GameEngine) logSoldiersTrained(def config.BuildingDef, n int) {
	w := config.Term(ge.age, config.TermSoldiers)
	ge.addLog(LogRoutine, fmt.Sprintf("%s finished training at the %s.", textfmt.Count(n, strings.ToLower(w.One), strings.ToLower(w.Many)), def.Name))
}

// postAllSoldiers puts a soldier at every post, with no food test: a town set
// up by hand (a test hook, or staged for a screen) starts with its army.
func (ge *GameEngine) postAllSoldiers() {
	ge.soldiersPosted = map[string]int{}
	for _, key := range ge.Buildings.soldierKeys {
		if n := ge.soldierPostsOf(key); n > 0 {
			ge.soldiersPosted[key] = n
		}
	}
}

// fillSoldierPosts is the staffing pass for the posts: the oldest building's
// posts first, each filled from the reserve and then by a new arrival, while
// the food margin test passes. It returns how many were posted, logs what
// moved in, and says once for each building whose posts are left waiting on
// food. Caller holds the write lock; the caller recalculates the rates when
// it posted anyone.
func (ge *GameEngine) fillSoldierPosts() int {
	var waiting []string
	for _, key := range ge.Buildings.soldierKeys {
		if ge.soldiersPosted[key]+ge.soldiersVacant[key] < ge.soldierPostsOf(key) {
			waiting = append(waiting, key)
		}
	}
	if len(waiting) == 0 {
		return ge.trainFallenPosts()
	}
	sort.SliceStable(waiting, func(i, j int) bool {
		return ge.Buildings.defs[waiting[i]].LineageTier < ge.Buildings.defs[waiting[j]].LineageTier
	})
	if ge.soldiersPosted == nil {
		ge.soldiersPosted = map[string]int{}
	}
	net, per, margin := ge.recruitFood()
	fed := ge.Resources.Get("food") > 0
	total := 0
	for _, key := range waiting {
		added := 0
		if need := ge.soldierPostsOf(key) - ge.soldiersPosted[key] - ge.soldiersVacant[key]; need > 0 && fed && net-per >= margin {
			// The reserve first: a person who already eats takes a post.
			fromReserve := ge.Workers.TakeIdle(need)
			need -= fromReserve
			added += fromReserve
			// Then arrivals, each of whom eats: as many as leave the margin
			// (arrival j is allowed while net - j x per stays at or above it).
			if need > 0 && per > 0 {
				k := int(math.Floor((net - margin) / per))
				k = max(0, min(k, need))
				net -= float64(float64(k) * per)
				added += k
			}
			ge.soldiersPosted[key] += added
		}
		total += added
		def := ge.Buildings.defs[key]
		if added > 0 {
			ge.logSoldiersArrived(def, added)
		}
		if ge.soldiersPosted[key]+ge.soldiersVacant[key] >= ge.soldierPostsOf(key) {
			delete(ge.soldierWaitLogged, key)
		} else if !ge.soldierWaitLogged[key] {
			if ge.soldierWaitLogged == nil {
				ge.soldierWaitLogged = map[string]bool{}
			}
			ge.soldierWaitLogged[key] = true
			ge.addLog(LogRoutine, fmt.Sprintf("The %s has posts waiting on food: its soldiers come as the food allows.", def.Name))
		}
	}
	return total + ge.trainFallenPosts()
}

// syncSoldiers holds the soldiers resource at the army at home (the posted army less
// the parties away), the compatibility
// bridge to every reader that still reads the resource. It also counts the
// soldiers that arrive (a copy built, an older copy upgraded) in the lifetime
// soldiers-trained total the milestones read (a load sets soldiersSeen to the
// army that stands, so a loaded army is not counted again). Caller holds the write lock. Called by
// recalculateRates, which runs every tick and after every change to the
// buildings.
func (ge *GameEngine) syncSoldiers() {
	// Posts that went with a lost copy are gone.
	for key, n := range ge.soldiersPosted {
		if limit := ge.soldierPostsOf(key); n > limit {
			if limit <= 0 {
				delete(ge.soldiersPosted, key)
			} else {
				ge.soldiersPosted[key] = limit
			}
		}
	}
	// A vacancy cannot outlast its post: a lost copy takes its empty posts too.
	for key, n := range ge.soldiersVacant {
		if room := ge.soldierPostsOf(key) - ge.soldiersPosted[key]; n > room {
			if room <= 0 {
				delete(ge.soldiersVacant, key)
			} else {
				ge.soldiersVacant[key] = room
			}
		}
	}
	posted := ge.soldierTotal()
	if posted > ge.soldiersSeen {
		ge.Stats.RecordSoldiersTrained(float64(posted - ge.soldiersSeen))
		// The soldiers who take their posts are what the soldiers' badge
		// ladder counts as produced (nothing makes them).
		if ge.account != nil && ge.badges != nil && ge.badges.countsProduction {
			if ge.produced == nil {
				ge.produced = make(map[string]float64, len(ge.Resources.order))
			}
			ge.produced[config.SoldiersKey] += float64(posted - ge.soldiersSeen)
		}
	}
	ge.soldiersSeen = posted
	home := float64(ge.soldiersHome())
	r := ge.Resources.resources["soldiers"]
	if r == nil {
		return
	}
	if home > 0 {
		ge.Resources.UnlockResource("soldiers")
	}
	r.Amount, r.Storage, r.Rate = home, home, 0
	delete(ge.Resources.grace, "soldiers")
}

// logSoldiersArrived says that n soldiers have taken their posts at a military
// building. Caller holds the lock.
func (ge *GameEngine) logSoldiersArrived(def config.BuildingDef, n int) {
	if n <= 0 {
		return
	}
	w := config.Term(ge.age, config.TermSoldiers)
	ge.addLog(LogRoutine, fmt.Sprintf("%s moved in with the %s.", textfmt.Count(n, strings.ToLower(w.One), strings.ToLower(w.Many)), def.Name))
}

// withSoldierFigures adds to a snapshot of the buildings the soldiers a copy of
// each military building comes with.
func (ge *GameEngine) withSoldierFigures(bs map[string]BuildingState) map[string]BuildingState {
	for key, b := range bs {
		if def, ok := ge.Buildings.defs[key]; ok && def.Soldiers > 0 {
			b.Soldiers = def.Soldiers
			bs[key] = b
		}
	}
	return bs
}

// soldiersPostedSave is the posted soldiers for a save: nil when there are
// none.
func (ge *GameEngine) soldiersPostedSave() map[string]int {
	var out map[string]int
	for key, n := range ge.soldiersPosted {
		if n = min(n, ge.soldierPostsOf(key)); n > 0 {
			if out == nil {
				out = map[string]int{}
			}
			out[key] = n
		}
	}
	return out
}

// soldiersVacantSave is the posts the fallen left, for a save: nil when none.
func (ge *GameEngine) soldiersVacantSave() map[string]int {
	var out map[string]int
	for key, n := range ge.soldiersVacant {
		if n = min(n, ge.soldierPostsOf(key)); n > 0 {
			if out == nil {
				out = map[string]int{}
			}
			out[key] = n
		}
	}
	return out
}

// loadSoldiersPosted restores the posted soldiers from a save, after the
// buildings are back. A save from before the posts has none: they fill again as
// the food allows. The army that stands is not counted as trained again.
func (ge *GameEngine) loadSoldiersPosted(save *GameSave) {
	ge.soldiersPosted, ge.soldierWaitLogged, ge.soldiersVacant = nil, nil, nil
	for key, n := range save.SoldiersVacant {
		if n = min(n, ge.soldierPostsOf(key)); n > 0 {
			if ge.soldiersVacant == nil {
				ge.soldiersVacant = map[string]int{}
			}
			ge.soldiersVacant[key] = n
		}
	}
	for key, n := range save.SoldiersPosted {
		if n = min(n, ge.soldierPostsOf(key)); n > 0 {
			if ge.soldiersPosted == nil {
				ge.soldiersPosted = map[string]int{}
			}
			ge.soldiersPosted[key] = n
		}
	}
	for _, key := range save.SoldiersWaiting {
		if ge.soldierWaitLogged == nil {
			ge.soldierWaitLogged = map[string]bool{}
		}
		ge.soldierWaitLogged[key] = true
	}
	ge.soldiersSeen = ge.soldierTotal()
}

// casualtyLine is the sentence the returns log and the mission's message add
// about the party: how many of it fell, counting the road's ambushes, or that
// all of it came home.
func (ge *GameEngine) casualtyLine(res ExpeditionResult) string {
	if res.Party <= 0 {
		return ""
	}
	if fell := res.Party - res.Marching + res.SoldiersFell; fell > 0 {
		return fmt.Sprintf(" %d of %d soldiers fell.", fell, res.Party)
	}
	return fmt.Sprintf(" All %d soldiers came home.", res.Party)
}

// partySizeLocked is the party a mission of the offer takes: its tier's share
// of the reference army of the age it is sent in (config.PartyShares).
func (ge *GameEngine) partySizeLocked(offer MissionOffer) int {
	kind := config.PartyShort
	switch {
	case offer.Tier == TierLong:
		kind = config.PartyLong
	case offer.Military():
		kind = config.PartyMilitary
	case offer.Tier == TierGrand:
		kind = config.PartyCivil
	}
	order, _ := ge.progress.AgeIndex(ge.age)
	return config.PartySize(kind, order)
}

// raidFallLocked is the soldiers a raid costs when it lands against the
// garrison: a twentieth of the army at home, rounded up, at least one. It
// returns the sentence for the log, or "" when nobody was home.
func (ge *GameEngine) raidFallLocked() string {
	home := ge.soldiersHome()
	if home <= 0 {
		return ""
	}
	fell := ge.soldiersFall(max(1, (raidFallPercent*home+99)/100))
	if fell <= 0 {
		return ""
	}
	return fmt.Sprintf("  [red]%s fell defending the town.[-]", textfmt.Count(fell, "soldier", "soldiers"))
}

// raidFallPercent is the share of the army at home that falls when a raid
// lands and the garrison blunts it.
const raidFallPercent = 5
