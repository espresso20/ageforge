package smoke

import "time"

// Idle targets: how long a player who checks in now and then (the idle
// style, with the game running between visits) should take to the first
// prestige. PacingTargets describe an attentive player (about a week to the
// Modern Age on the one-week curve; the greedy bot gets there in about 5.3
// days); these describe the player AgeForge is built for, who looks in a few
// times a day and leaves a build plan behind (Bot.planAhead). The full tier's
// idle scenario plays IdleSeeds seeds at each interval and, under -pacing
// enforce, fails when the median misses its target. See
// design-and-architecture/economy.md, "Idle targets".
//
// These are the one-week curve's first targets: the medians measured when it
// landed (3 seeds, worker shares on: 5.63, 6.50 and 9.20 days) with about 20%
// headroom, so a balance change that slows check-in play noticeably fails the
// nightly. The 1-hour target sits on the floor TestIdleTargetsAreSane allows,
// the 167-hour sum of the greedy targets to the Modern Age.
// Away-proofing (Pacing v2, PR 2) tightens them, and starts enforcing the
// ratio to the active bot that the idle scenario reports.
type IdleTarget struct {
	CheckIn       time.Duration
	FirstPrestige time.Duration
}

// IdleTargets, shortest interval first.
var IdleTargets = []IdleTarget{
	{CheckIn: time.Hour, FirstPrestige: 168 * time.Hour},     // 7 days (measured 5.63)
	{CheckIn: 3 * time.Hour, FirstPrestige: 192 * time.Hour}, // 8 days (measured 6.50)
	{CheckIn: 8 * time.Hour, FirstPrestige: 264 * time.Hour}, // 11 days (measured 9.20)
}

// IdleSeeds is how many seeds the idle scenario plays per interval; the
// target is graded on their median.
const IdleSeeds = 3

// IdleBudgetFactor is how far past its target an idle run may go before it
// stops (a run that stops short of the prestige misses the target).
const IdleBudgetFactor = 2.0

// VeteranIdleWatch is the veteran check-in player's watch line (Era Mastery,
// Pacing v2 PR 5): the idle scenario also plays the veteran preset actively
// and at 3- and 8-hour check-ins, and reports each check-in run's time to
// the first prestige over the active veteran's on the same seed. Measured
// when Era Mastery landed (2 seeds): 1.83x at 3 hours and 3.40x at 8 hours,
// because a veteran's early ages are shorter than one visit and the plan
// only looks one age ahead. These lines sit about 20% above that and only
// warn; Pacing v2 PR 6's legacy kit (a plan template that chains ages
// between visits) is meant to bring them to 1.5x and 2.0x, and enforces it.
var VeteranIdleWatch = map[time.Duration]float64{
	3 * time.Hour: 2.2,
	8 * time.Hour: 4.1,
}
