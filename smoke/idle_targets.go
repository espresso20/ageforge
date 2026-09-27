package smoke

import "time"

// Idle targets: how long a player who checks in now and then (the idle
// style, with the game running between visits) should take to the first
// prestige. PacingTargets describe an attentive player (about three days to
// the Modern Age); these describe the player AgeForge is built for, who looks
// in a few times a day and leaves a build plan behind (Bot.planAhead). The
// full tier's idle scenario plays IdleSeeds seeds at each interval and, under
// -pacing enforce, fails when the median misses its target. See
// design-and-architecture/economy.md, "Idle targets".
type IdleTarget struct {
	CheckIn       time.Duration
	FirstPrestige time.Duration
}

// IdleTargets, shortest interval first.
var IdleTargets = []IdleTarget{
	{CheckIn: time.Hour, FirstPrestige: 84 * time.Hour},      // 3.5 days
	{CheckIn: 3 * time.Hour, FirstPrestige: 120 * time.Hour}, // 5 days
	{CheckIn: 8 * time.Hour, FirstPrestige: 192 * time.Hour}, // 8 days
}

// IdleSeeds is how many seeds the idle scenario plays per interval; the
// target is graded on their median.
const IdleSeeds = 3

// IdleBudgetFactor is how far past its target an idle run may go before it
// stops (a run that stops short of the prestige misses the target).
const IdleBudgetFactor = 2.0
