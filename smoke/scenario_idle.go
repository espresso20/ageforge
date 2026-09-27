package smoke

import (
	"fmt"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// idleResult is one interval of the idle scenario.
type idleResult struct {
	target  IdleTarget
	sum     *Summary
	firsts  []float64 // first-prestige times, seconds at 1x, of the runs that got there
	seeds   int
	verdict string
}

// runIdle plays the idle style at every IdleTargets interval on IdleSeeds
// seeds each and grades the median time to the first prestige against the
// target. Under -pacing enforce a miss fails the scenario (and so the
// nightly); otherwise it is reported. Soft-locks, invariant violations and
// panics always fail it, as in every bot scenario. The per-age greedy
// targets don't apply: a check-in player waits for the next visit to press
// advance, so early ages are slow by nature.
func runIdle(e *Env, res *Result) {
	enforce := e.Pacing == PacingEnforce
	seeds := e.seeds(IdleSeeds)
	var results []idleResult
	for _, t := range IdleTargets {
		base := e.Base
		base.Cycles, base.CheckIn = 1, t.CheckIn
		base.MaxSim = time.Duration(float64(t.FirstPrestige) * IdleBudgetFactor)
		if o := e.Overrides.MaxSim; o > 0 {
			base.MaxSim = o
		}
		cfg, err := ApplyStyle(base, StyleIdle)
		if err != nil {
			res.fail("config", "%v", err)
			return
		}
		cfg.Pacing = PacingReport
		name := "idle-" + shortDur(t.CheckIn)
		sum := runBotSet(e, res, name, "idle", cfg, seeds)
		ir := idleResult{target: t, sum: sum, seeds: len(seeds)}
		for _, r := range sum.Runs {
			for _, c := range r.Cycles {
				if c.Cycle == 1 && c.Prestiged {
					ir.firsts = append(ir.firsts, c.Seconds)
				}
			}
		}
		ir.verdict = VerdictOK
		_, med, _ := spread(ir.firsts)
		switch {
		case len(ir.firsts)*2 <= ir.seeds:
			// Most runs never prestiged: the median is past the budget.
			ir.verdict = VerdictSlow
		case med > t.FirstPrestige.Seconds():
			ir.verdict = VerdictSlow
		}
		if ir.verdict != VerdictOK {
			msg := fmt.Sprintf("checking in every %s, the median first prestige over %d seeds is %s (reached by %d), over the %s target",
				shortDur(t.CheckIn), ir.seeds, firstStr(ir.firsts), len(ir.firsts), dur(t.FirstPrestige.Seconds()))
			if enforce {
				res.fail("idle_target", "%s", msg)
			} else {
				res.warn("idle_target", "%s (reported, not failed, without -pacing enforce)", msg)
			}
		}
		results = append(results, ir)
	}
	var parts []string
	for _, ir := range results {
		_, med, _ := spread(ir.firsts)
		parts = append(parts, fmt.Sprintf("%s: %s (%s)", shortDur(ir.target.CheckIn), dur(med), verdictMark(ir.verdict)))
	}
	res.Summary = "first prestige, median of " + fmt.Sprint(len(seeds)) + " seeds: " + strings.Join(parts, ", ")
	res.section("Idle targets", "%s", idleTargetTable(results))
	res.section("Time per age", "%s", idleAgeTable(results))
}

func firstStr(v []float64) string {
	if len(v) == 0 {
		return "never"
	}
	lo, med, hi := spread(v)
	return fmt.Sprintf("%s (%s–%s)", dur(med), dur(lo), dur(hi))
}

// shortDur prints an interval as "1h", "90m".
func shortDur(d time.Duration) string {
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", d/time.Hour)
	}
	return fmt.Sprintf("%dm", d/time.Minute)
}

func idleTargetTable(results []idleResult) string {
	var sb strings.Builder
	sb.WriteString("A player who checks in every interval below and leaves a build plan each visit (the idle style, overflow on). Median over the seeds of the time to the first prestige; a run that never got there counts as past the target.\n\n")
	sb.WriteString("| check-in | seeds | first prestige, median (min–max) | target | visits (median) | |\n|---|---|---|---|---|---|\n")
	for _, ir := range results {
		_, med, _ := spread(ir.firsts)
		visits := "-"
		if med > 0 {
			visits = fmt.Sprintf("%.0f", med/ir.target.CheckIn.Seconds())
		}
		fmt.Fprintf(&sb, "| %s | %d/%d | %s | %s | %s | %s |\n", shortDur(ir.target.CheckIn), len(ir.firsts), ir.seeds,
			firstStr(ir.firsts), dur(ir.target.FirstPrestige.Seconds()), visits, verdictMark(ir.verdict))
	}
	return sb.String()
}

// idleAgeTable is the median time in each first-cycle age per interval,
// against the greedy target for scale.
func idleAgeTable(results []idleResult) string {
	var sb strings.Builder
	sb.WriteString("Median time in each age of the first cycle (the greedy target for scale; a check-in player presses advance only when they look in).\n\n| age | target |")
	for _, ir := range results {
		fmt.Fprintf(&sb, " %s |", shortDur(ir.target.CheckIn))
	}
	sb.WriteString("\n|---|---|" + strings.Repeat("---|", len(results)) + "\n")
	for _, age := range config.AgeOrder() {
		t, ok := Target(age)
		if !ok {
			continue
		}
		row := fmt.Sprintf("| %s | %s |", age, dur(t.Seconds()))
		any := false
		for _, ir := range results {
			cell := "-"
			for _, p := range ir.sum.Pacing {
				if p.Cycle == 1 && p.Age == age && !p.Prestiged {
					cell = dur(p.MedianSecs)
					any = true
				}
			}
			row += " " + cell + " |"
		}
		if any {
			sb.WriteString(row + "\n")
		}
	}
	return sb.String()
}
