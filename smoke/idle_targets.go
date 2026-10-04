package smoke

import "time"

// Idle targets: how long a player who checks in now and then (the idle
// style, with the game running between visits) should take to the first
// prestige. PacingTargets describe an attentive player (about a week to the
// Modern Age on the one-week curve; the greedy bot gets there in about 5.3
// days); these describe the player AgeForge is built for, who looks in a few
// times a day and leaves a build plan behind (Bot.planAhead). The full tier's
// idle scenario plays IdleSeeds seeds at each interval and, under -pacing
// enforce, fails when the median misses its target.
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

// VeteranIdleMax is the veteran check-in player's limit (Pacing v2, PR 6):
// the idle scenario plays the veteran preset with the legacy kit actively
// and at 3- and 8-hour check-ins, and holds each check-in run's time to the
// first prestige to this many times the active veteran's on the same seed
// (median across seeds), failing under -pacing enforce.
//
// Without the kit (when Era Mastery landed) they were 1.88x and 3.41x: a
// veteran's ages are shorter than one visit, and the plan only looked one
// age ahead. The kit's plan template re-adds each age's part at every
// advance, so the plan chains ages between visits. Measured with the canned
// kit (3 seeds): 1.47x at 3 hours and 2.34x at 8 hours (per seed 1.36 to
// 1.54x and 2.31 to 2.35x), once auto-recruit stopped hiring more workers
// than a mastered age's food fed (they were 1.55x and 2.42x before that
// fix). The limits are set from that with about 9% headroom; the Pacing v2
// plan had hoped for 1.5x and 2.0x.
//
// What the measurement leans on: the canned template holds the techs the
// veteran researched as planned research items (the player's own path,
// which the template keeps). With no techs in the template the same runs
// take 1.65x and 3.08x, since nothing researches between visits once the
// plan has moved on an age. What is left at 8 hours is mostly stone: a
// veteran reaches the Classical to the Renaissance Age inside one visit's
// gap with next to no stone coming in, and the stone its required buildings
// and its wonder (the Parthenon, the Great Library, the Sistine Chapel) take
// is bought at the market on the next visit; the plan's own trades wait on
// the market's recovery. The wonder's bank is not what holds it up:
// overflow at a full store goes into the age's wonder already, and letting
// a wonder queued in the plan take the plan's overflow too (with wonder
// overflow off) left both ratios as they were.
var VeteranIdleMax = map[time.Duration]float64{
	3 * time.Hour: 1.6,
	8 * time.Hour: 2.55,
}
