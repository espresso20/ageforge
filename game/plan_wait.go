package game

import "github.com/espresso20/ageforge/rowid"

// planWaitEvery is how often the plan's wait is looked at, in ticks.
const planWaitEvery = 10

// planWatch notes, for the log, when the plan starts waiting on money, when it
// goes on again and when it is finished. It looks at the plan every
// planWaitEvery ticks. Under the write lock.
func (ge *GameEngine) planWatch() {
	if ge.tick%planWaitEvery != 0 {
		return
	}
	if len(ge.plan) == 0 {
		if ge.planSeen {
			ge.planSeen, ge.planWaitKey = false, ""
			ge.addLog(OfficePlan, "success", rowid.Line("game.plan.planWatch.03", "Plan finished."))
		}
		return
	}
	ge.planSeen = true
	key, short := "", ""
	for _, v := range ge.planViews() {
		if v.Status == PlanStatusWaiting && v.Short != "" {
			key, short = v.Kind+":"+v.Key, v.Short
		}
		break // only the item at the head holds the plan up
	}
	switch {
	case key != "" && ge.planWaitKey == "":
		ge.planWaitKey, ge.planWaitSince = key, ge.tick
		ge.addLog(OfficePlan, "info", rowid.Sprintf("game.plan.planWatch.01", "Plan is waiting on %s.", ge.rules.ResourceLabel(short)), logTrouble, logSubj(key))
	case key == "" && ge.planWaitKey != "":
		ge.addLog(OfficePlan, "info", rowid.Sprintf("game.plan.planWatch.02", "Plan went on after waiting %s.", ge.durationLocked(ge.tick-ge.planWaitSince)), logTrouble, logSubj(ge.planWaitKey))
		ge.planWaitKey = ""
	}
}
