package game

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Soldiers. A military building comes with its soldiers (config.BuildingDef.
// Soldiers a copy, derived in config/soldier_rule.go), as a knowledge building
// comes with its scholars: they are people from the moment the copy stands,
// they take housing and eat like anyone (population.go), and a sold or
// destroyed copy takes its soldiers with it. A new copy's soldiers move in
// whether or not the homes have room; the reserve stops growing until they do.
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
	// They are equal until a soldier can be away.
	Posted, Posts int
}

// soldierTotal is every soldier at his post: the posts of every standing
// military building. A plain loop over the military keys, as scholarTotal.
func (ge *GameEngine) soldierTotal() int {
	total := 0
	bm := ge.Buildings
	for _, key := range bm.soldierKeys {
		if n := bm.counts[key]; n > 0 {
			if s := bm.defs[key].Soldiers; s > 0 {
				total += n * s
			}
		}
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
		out = append(out, SoldierPost{Key: key, Name: def.Name, Count: count, PerCopy: def.Soldiers, Posted: count * def.Soldiers, Posts: count * def.Soldiers})
	})
	return out
}

// syncSoldiers holds the soldiers resource at the posted army, the compatibility
// bridge to every reader that still reads the resource. It also counts the
// soldiers that arrive (a copy built, an older copy upgraded) in the lifetime
// soldiers-trained total the milestones read (a load sets soldiersSeen to the
// army that stands, so a loaded army is not counted again). Caller holds the write lock. Called by
// recalculateRates, which runs every tick and after every change to the
// buildings.
func (ge *GameEngine) syncSoldiers() {
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

// logSoldiersArrived says, for a military building that has just been built,
// that its soldiers have moved in. Caller holds the lock.
func (ge *GameEngine) logSoldiersArrived(def config.BuildingDef) {
	if def.Soldiers <= 0 {
		return
	}
	w := config.Term(ge.age, config.TermSoldiers)
	ge.addLog(LogRoutine, fmt.Sprintf("%s moved in with the %s.", textfmt.Count(def.Soldiers, strings.ToLower(w.One), strings.ToLower(w.Many)), def.Name))
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
