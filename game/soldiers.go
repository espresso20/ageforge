package game

import (
	"fmt"
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
// The army is the soldiers at their posts, which is every soldier in this
// change: nobody goes out on a mission yet. Everything that reads soldiers
// (the Defense Rating, raids, the Brace view, a campaign's price, the soldier
// milestones) still reads the "soldiers" resource, and syncSoldiers holds that
// resource at the posted army every tick. Nothing else writes it.

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
		if ge.soldiersPosted[key] < ge.soldierPostsOf(key) {
			waiting = append(waiting, key)
		}
	}
	if len(waiting) == 0 {
		return 0
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
		for ge.soldiersPosted[key] < ge.soldierPostsOf(key) {
			if !fed || net-per < margin {
				break
			}
			if ge.Workers.TakeIdle(1) == 0 {
				// A new arrival eats: it must leave the margin.
				net -= per
			}
			ge.soldiersPosted[key]++
			added++
		}
		total += added
		def := ge.Buildings.defs[key]
		if added > 0 {
			ge.logSoldiersArrived(def, added)
		}
		if ge.soldiersPosted[key] >= ge.soldierPostsOf(key) {
			delete(ge.soldierWaitLogged, key)
		} else if !ge.soldierWaitLogged[key] {
			if ge.soldierWaitLogged == nil {
				ge.soldierWaitLogged = map[string]bool{}
			}
			ge.soldierWaitLogged[key] = true
			ge.addLog(LogRoutine, fmt.Sprintf("The %s has posts waiting on food: its soldiers come as the food allows.", def.Name))
		}
	}
	return total
}

// syncSoldiers holds the soldiers resource at the posted army, the compatibility
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
	posted := ge.soldierTotal()
	if posted > ge.soldiersSeen {
		ge.Stats.RecordSoldiersTrained(float64(posted - ge.soldiersSeen))
	}
	ge.soldiersSeen = posted
	home := float64(posted)
	if ge.garrisonForTestSet {
		home = ge.garrisonForTest
	}
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

// loadSoldiersPosted restores the posted soldiers from a save, after the
// buildings are back. A save from before the posts has none: they fill again as
// the food allows. The army that stands is not counted as trained again.
func (ge *GameEngine) loadSoldiersPosted(save *GameSave) {
	ge.soldiersPosted, ge.soldierWaitLogged = nil, nil
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
