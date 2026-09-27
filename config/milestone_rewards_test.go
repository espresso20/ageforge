package config

import "testing"

// TestMilestoneRewardsAreUnlocked: a milestone may only grant a resource the
// player already has by the earliest age the milestone can complete. "Iron
// Forged" granted 40 coal on reaching the Iron Age, three ages before coal
// unlocks, so the reward went into a resource the player could not see or
// spend. The earliest age is the latest of the milestone's MinAge, the ages
// of the buildings and techs it requires, and the unlock ages of the
// resources it asks for. Chains grant only a speed boost (MilestoneChainDef
// has no resource rewards), so milestones are the whole check.
func TestMilestoneRewardsAreUnlocked(t *testing.T) {
	order := map[string]int{}
	for i, a := range AgeOrder() {
		order[a] = i
	}
	unlockedAt := map[string]int{}
	for i, a := range Ages() {
		for _, r := range a.UnlockResources {
			unlockedAt[r] = i
		}
	}
	defs := BuildingByKey()
	techs := TechByKey()
	resources := ResourceByKey()
	for _, m := range Milestones() {
		earliest := order[m.MinAge]
		for key := range m.MinBuildings {
			earliest = max(earliest, order[defs[key].RequiredAge])
		}
		for _, key := range m.RequiredTechs {
			earliest = max(earliest, order[techs[key].Age])
		}
		for res := range m.MinResources {
			earliest = max(earliest, unlockedAt[res])
		}
		for _, e := range m.Rewards {
			if _, isResource := resources[e.Target]; !isResource {
				continue // a bonus key (iron_rate, production_all, ...)
			}
			at, ok := unlockedAt[e.Target]
			if !ok || at > earliest {
				t.Errorf("milestone %s can complete in %s but grants %s, which unlocks in %s",
					m.Key, AgeOrder()[earliest], e.Target, ageName(at, ok))
			}
		}
	}
}

func ageName(i int, ok bool) string {
	if !ok {
		return "no age"
	}
	return AgeOrder()[i]
}
