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
	ratios  []float64 // each such run's time over the active bot's on the same seed
	seeds   int
	verdict string
}

// firstPrestiges maps each run's seed to its first prestige's 1x seconds,
// for the runs that got there.
func firstPrestiges(sum *Summary) map[int64]float64 {
	out := map[int64]float64{}
	for _, r := range sum.Runs {
		for _, c := range r.Cycles {
			if c.Cycle == 1 && c.Prestiged {
				out[r.Seed] = c.Seconds
			}
		}
	}
	return out
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
	// The active reference: the greedy bot on the same seeds, to the first
	// prestige. Each check-in run is reported as a ratio to it (Pacing v2's
	// away-proofing holds the ratios to 1.2, 1.3 and 1.5x).
	ref := e.Base
	ref.Cycles, ref.FinalAge, ref.MaxSim, ref.Pacing = 1, "", 1000*time.Hour, PacingReport
	active := firstPrestiges(runBotSet(e, res, "idle-active", "idle", ref, seeds))
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
		got := firstPrestiges(sum)
		for _, seed := range seeds {
			secs, ok := got[seed]
			if !ok {
				continue
			}
			ir.firsts = append(ir.firsts, secs)
			if a := active[seed]; a > 0 {
				ir.ratios = append(ir.ratios, secs/a)
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
		part := fmt.Sprintf("%s: %s (%s", shortDur(ir.target.CheckIn), dur(med), verdictMark(ir.verdict))
		if len(ir.ratios) > 0 {
			_, r, _ := spread(ir.ratios)
			part += fmt.Sprintf(", %.2fx active", r)
		}
		parts = append(parts, part+")")
	}
	res.Summary = "first prestige, median of " + fmt.Sprint(len(seeds)) + " seeds: " + strings.Join(parts, ", ")
	res.section("Idle targets", "%s", idleTargetTable(results, active, len(seeds)))
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

func idleTargetTable(results []idleResult, active map[int64]float64, seeds int) string {
	var sb strings.Builder
	sb.WriteString("A player who checks in every interval below and leaves a build plan each visit (the idle style, overflow on). Median over the seeds of the time to the first prestige; a run that never got there counts as past the target. The ratio is each run's time over the greedy bot's on the same seed, median over the seeds (reported, not graded).\n\n")
	var act []float64
	for _, v := range active {
		act = append(act, v)
	}
	fmt.Fprintf(&sb, "Active reference (the greedy bot, same seeds): %s, %d/%d seeds.\n\n", firstStr(act), len(act), seeds)
	sb.WriteString("| check-in | seeds | first prestige, median (min–max) | target | vs active | visits (median) | |\n|---|---|---|---|---|---|---|\n")
	for _, ir := range results {
		_, med, _ := spread(ir.firsts)
		visits := "-"
		if med > 0 {
			visits = fmt.Sprintf("%.0f", med/ir.target.CheckIn.Seconds())
		}
		ratio := "-"
		if len(ir.ratios) > 0 {
			_, r, _ := spread(ir.ratios)
			ratio = fmt.Sprintf("%.2fx", r)
		}
		fmt.Fprintf(&sb, "| %s | %d/%d | %s | %s | %s | %s | %s |\n", shortDur(ir.target.CheckIn), len(ir.firsts), ir.seeds,
			firstStr(ir.firsts), dur(ir.target.FirstPrestige.Seconds()), ratio, visits, verdictMark(ir.verdict))
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
