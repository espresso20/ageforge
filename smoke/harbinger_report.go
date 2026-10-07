package smoke

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

func pointsStr(pts map[string]ResourcePoint, cost map[string]float64) string {
	var parts []string
	for _, r := range sortedKeys(cost) {
		p := pts[r]
		parts = append(parts, fmt.Sprintf("%s %s/%s", r, num(p.Stock), num(p.Storage)))
	}
	return orDefault(strings.Join(parts, ", "), "-")
}

// capsStr is costStr for storage caps, where -1 means no cap.
func capsStr(m map[string]float64) string {
	var parts []string
	for _, k := range sortedKeys(m) {
		v := "no cap"
		if m[k] >= 0 {
			v = num(m[k])
		}
		parts = append(parts, k+" "+v)
	}
	return strings.Join(parts, ", ")
}

func affordStr(ever bool, after float64) string {
	if !ever {
		return "never"
	}
	return "after " + dur(after)
}

// quartiles returns the 25th, 50th and 75th percentiles of v (nearest rank).
func quartiles(v []float64) (q1, med, q3 float64) {
	if len(v) == 0 {
		return 0, 0, 0
	}
	c := append([]float64(nil), v...)
	sort.Float64s(c)
	at := func(p float64) float64 { return c[int(float64(p*float64(len(c)-1))+0.5)] }
	return at(0.25), at(0.5), at(0.75)
}

// writeFates renders the fated-doom section: what each era's hidden fate held
// and how it went, per seed and per era, the strike and warning timing, and
// the catastrophes per first run against the baseline before fated dooms.
func (s *Summary) writeFates(sb *strings.Builder) {
	var rows []*FateRow
	for _, r := range s.Runs {
		rows = append(rows, r.Fates...)
	}
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(sb, "\n## Fated dooms\n\nEvery era's hidden fate as the bot lived it, from the engine's fate events (the bot itself never sees them). On entering an era from the Iron Era on a doom is fated %.0f%% of the time, its strike tick drawn across the era's ages at their pacing targets; its harbinger comes a lead before it, and it strikes at its tick or at an advance the player reaches first.\n\n", game.FateChance*100)
	sb.WriteString("| seed | eras entered | fated | false prophets | struck | spared | revealed | cleared or open | struck or spared at an advance | endured | succumbed | fated catastrophes in cycle 1 (expected) | Last Passage in cycle 1 |\n|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	var perRun, expRun []float64
	passages := 0
	for _, r := range s.Runs {
		var eras, fated, falses, struck, spared, revealed, cleared, atAdv int
		exp1 := 0.0
		for _, f := range r.Fates {
			if config.FateAllowed(f.Epoch) {
				eras++
			}
			if f.Fated {
				fated++
			}
			if f.FalseProphet {
				falses++
			}
			switch f.Outcome {
			case game.FateStruck:
				struck++
			case game.FateSpared:
				spared++
			case game.FateRevealed:
				revealed++
			case FateCleared, FateOpen:
				cleared++
			}
			if f.AtAdvance && f.Outcome != game.FateRevealed {
				atAdv++
			}
			if f.Cycle == 1 {
				exp1 += f.Expected
			}
		}
		// The Last Passage is the prestige's own roll, outside the fate
		// model (and never met on a first run to the Modern Age, the
		// baseline's), so the comparison leaves it out.
		lp1 := r.Stats.LastPassagesByCycle[1]
		cat1 := r.Stats.CatastrophesByCycle[1] - lp1
		fmt.Fprintf(sb, "| %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d (%.2f) | %d |\n", r.Seed, eras, fated, falses, struck, spared,
			revealed, cleared, atAdv, r.Stats.CatastrophesEndured, r.Stats.CatastrophesSuccumbed, cat1, exp1, lp1)
		for _, c := range r.Cycles {
			if c.Cycle == 1 && c.Prestiged {
				perRun = append(perRun, float64(cat1))
				expRun = append(expRun, exp1)
				passages += lp1
			}
		}
	}

	// Per era.
	type agg struct {
		entered, fated, struck, spared, atAdv int
		offsets, warnings                     []float64
	}
	byEra := map[string]*agg{}
	var offsets, warnings, warnHours []float64
	for _, f := range rows {
		if !config.FateAllowed(f.Epoch) {
			continue
		}
		a := byEra[f.Epoch]
		if a == nil {
			a = &agg{}
			byEra[f.Epoch] = a
		}
		a.entered++
		if !f.Fated {
			continue
		}
		a.fated++
		a.offsets = append(a.offsets, f.StrikeFrac)
		offsets = append(offsets, f.StrikeFrac)
		switch f.Outcome {
		case game.FateStruck:
			a.struck++
		case game.FateSpared:
			a.spared++
		default:
			continue
		}
		if f.AtAdvance {
			a.atAdv++
		}
		a.warnings = append(a.warnings, f.WarningAgeFrac)
		warnings = append(warnings, f.WarningAgeFrac)
		warnHours = append(warnHours, float64(f.WarningTicks)*game.BaseTickInterval.Seconds())
	}
	sb.WriteString("\n| era | entered | fated | struck | spared | at an advance | strike offset, median share of the era | warning, median share of the age |\n|---|---|---|---|---|---|---|---|\n")
	for _, ep := range config.Epochs() {
		a := byEra[ep.Key]
		if a == nil {
			continue
		}
		_, off, _ := spread(a.offsets)
		_, warn, _ := spread(a.warnings)
		fmt.Fprintf(sb, "| %s | %d | %d | %d | %d | %d | %.2f | %.2f |\n", ep.Key, a.entered, a.fated, a.struck, a.spared, a.atAdv, off, warn)
	}
	if len(offsets) > 0 {
		q1, med, q3 := quartiles(offsets)
		fmt.Fprintf(sb, "\nStrike offset, as a share of the era's window from its entry (%d fated eras): quartiles %.2f / %.2f / %.2f. Uniform would read 0.25 / 0.50 / 0.75.\n", len(offsets), q1, med, q3)
	}
	if len(warnings) > 0 {
		q1, med, q3 := quartiles(warnings)
		_, h, _ := spread(warnHours)
		fmt.Fprintf(sb, "\nHarbinger warning before the doom resolved (%d dooms), as a share of the age's target: quartiles %.2f / %.2f / %.2f; median %s at 1x.\n", len(warnings), q1, med, q3, dur(h))
	}
	if len(perRun) > 0 {
		mean, se := meanSE(perRun)
		expMean, _ := meanSE(expRun)
		fmt.Fprintf(sb, "\nCatastrophes per first run (cycle 1 to its prestige, %d seeds, the Last Passage left out): %.2f ± %.2f measured, %.2f expected from the faith the bot kept. Before fated dooms (%s, %d seeds): %.2f expected, %.2f measured.",
			len(perRun), mean, se, expMean, BaselineCommit, BaselineSeeds, BaselineCatastrophesExpected, BaselineCatastrophesMeasured)
		if passages > 0 {
			fmt.Fprintf(sb, " The Last Passage struck %d time(s) on top, at those prestiges.", passages)
		}
		sb.WriteString("\n")
	}
}

// meanSE is the mean of v and its standard error.
func meanSE(v []float64) (mean, se float64) {
	if len(v) == 0 {
		return 0, 0
	}
	for _, x := range v {
		mean += x
	}
	mean /= float64(len(v))
	if len(v) < 2 {
		return mean, 0
	}
	ss := 0.0
	for _, x := range v {
		d := x - mean
		ss += float64(d * d)
	}
	return mean, math.Sqrt(ss/float64(len(v)-1)) / math.Sqrt(float64(len(v)))
}

// passageLabel names what a harbinger thread warned of: its era's doom, or
// the Last Passage (a TargetEpoch of "", the final epoch's prestige).
func passageLabel(epoch, target string) string {
	if target == "" {
		return epoch + " → Last Passage"
	}
	if target == epoch {
		return epoch
	}
	return epoch + " → " + target
}

// writeHarbingers renders the harbinger section: event counts, the static
// price table and each thread's affordability as the bot lived it.
func (s *Summary) writeHarbingers(sb *strings.Builder) {
	fmt.Fprintf(sb, "\n## Harbingers\n\nBot policy `%s` (Appease both levels, Brace level 1).\n\n", orDefault(s.Config.Harbinger, HarbingerIgnore))

	sb.WriteString("| seed | threads started | speaker handoffs | verdicts | false prophets revealed | appease bought | brace bought |\n|---|---|---|---|---|---|---|\n")
	for _, r := range s.Runs {
		st := r.Stats
		fmt.Fprintf(sb, "| %d | %s | %s | %s | %d | %d | %d |\n", r.Seed, countStr(st.HarbingerThreads), countStr(st.HarbingerHandoffs),
			countStr(st.HarbingerVerdicts), st.FalseProphets, st.Actions["harbinger_appease"], st.Actions["harbinger_brace"])
	}
	sb.WriteString("\nThread counts are keyed by the era whose doom the harbinger warns of. A false prophet is only revealed when its thread resolves; the live view hides it.\n")

	sb.WriteString("\n### Level-1 prices against the most storage buildable in the age the harbinger arrives in\n\n")
	sb.WriteString("A thread's price is set when its harbinger arrives and stays the same in every age it lives through. Appease is priced on the age the harbinger arrives in (what that age makes in the thread's shortest warning: a fifth of the age for a doom, two thirds of it for the Last Passage), so every age one can arrive in has a row. Brace is priced by the era for a doom, and on the same warning as Appease for the Last Passage. The Last Passage's thread comes with its era's first age; its later rows price a thread that began there (an older save, the dev console). \"Max storage\" assumes every capped storage building up to that age was built out, plus storage techs.\n\n")
	sb.WriteString("| doom (harbinger arrives in) | Appease L1 | Brace L1 | max storage in that age | fits |\n|---|---|---|---|---|\n")
	for _, p := range s.Prices {
		var over []string
		for _, c := range []map[string]float64{p.AppeaseL1, p.BraceL1} {
			for _, res := range sortedKeys(c) {
				if m := p.MaxStorage[res]; m >= 0 && c[res] > m {
					over = append(over, fmt.Sprintf("no: %s %s > %s", res, num(c[res]), num(p.MaxStorage[res])))
				}
			}
		}
		fmt.Fprintf(sb, "| %s (%s) | %s | %s | %s | %s |\n", passageLabel(p.Epoch, p.TargetEpoch), p.Age,
			orDefault(costStr(p.AppeaseL1), "-"), orDefault(costStr(p.BraceL1), "-"), capsStr(p.MaxStorage),
			orDefault(strings.Join(over, "; "), "yes"))
	}

	var threads []*HarbingerThread
	seedOf := map[*HarbingerThread]int64{}
	for _, r := range s.Runs {
		for _, th := range r.Harbingers {
			threads = append(threads, th)
			seedOf[th] = r.Seed
		}
	}
	if len(threads) == 0 {
		sb.WriteString("\nNo harbinger thread was observed.\n")
		return
	}
	for _, kind := range []string{"Appease", "Brace"} {
		fmt.Fprintf(sb, "\n### %s affordability per thread\n\nStock/storage when the thread started and at the last decision before its doom resolved (or the run ended).\n\n", kind)
		sb.WriteString("| seed | cycle | doom | speakers | outcome | L1 cost | at start | at resolution | ever fits storage | ever affordable | bought |\n|---|---|---|---|---|---|---|---|---|---|---|\n")
		for _, th := range threads {
			cost, fits, afford, after, lvl := th.AppeaseL1, th.AppeaseFits, th.AppeaseAffordable, th.AppeaseAfterSecs, th.AppeaseLevel
			if kind == "Brace" {
				cost, fits, afford, after, lvl = th.BraceL1, th.BraceFits, th.BraceAffordable, th.BraceAfterSecs, th.BraceLevel
			}
			outcome := th.Outcome
			if th.FalseProphet != nil && *th.FalseProphet {
				outcome += " (false prophet)"
			}
			fmt.Fprintf(sb, "| %d | %d | %s | %s | %s | %s | %s | %s | %v | %s | %d |\n", seedOf[th], th.Cycle, passageLabel(th.Epoch, th.TargetEpoch),
				strings.Join(th.Speakers, ", "), outcome, orDefault(costStr(cost), "-"),
				pointsStr(th.AtStart, cost), pointsStr(th.AtPassage, cost), fits, affordStr(afford, after), lvl)
		}
	}

	s.writeAppeaseTiming(sb, threads)

	// Flag passages where no seed could ever afford an answer.
	type agg struct{ seen, appease, brace int }
	byPassage := map[string]*agg{}
	for _, th := range threads {
		k := passageLabel(th.Epoch, th.TargetEpoch)
		a := byPassage[k]
		if a == nil {
			a = &agg{}
			byPassage[k] = a
		}
		a.seen++
		if th.AppeaseAffordable {
			a.appease++
		}
		if th.BraceAffordable {
			a.brace++
		}
	}
	sb.WriteString("\n")
	for _, k := range sortedKeys(byPassage) {
		a := byPassage[k]
		for _, x := range []struct {
			name string
			n    int
		}{{"Appease", a.appease}, {"Brace", a.brace}} {
			if x.n == 0 {
				fmt.Fprintf(sb, "- **%s never affordable** in the %s thread across %d observed thread(s).\n", x.name, k, a.seen)
			}
		}
	}
}

// writeAppeaseTiming summarizes, per passage, when Appease level 1 and then
// level 2 (with level 1 bought) first became affordable, against the
// thread's length: the Appease pricing aims at level 1 partway through and
// level 2 near the passage.
func (s *Summary) writeAppeaseTiming(sb *strings.Builder, threads []*HarbingerThread) {
	type agg struct{ l1, l2, length []float64 }
	by := map[string]*agg{}
	var keys []string
	for _, th := range threads {
		k := passageLabel(th.Epoch, th.TargetEpoch)
		a := by[k]
		if a == nil {
			a = &agg{}
			by[k] = a
			keys = append(keys, k)
		}
		a.length = append(a.length, th.LengthSecs)
		if th.AppeaseAffordable {
			a.l1 = append(a.l1, th.AppeaseAfterSecs)
		}
		if th.AppeaseL2AfterSecs >= 0 {
			a.l2 = append(a.l2, th.AppeaseL2AfterSecs)
		}
	}
	cell := func(v []float64, n int, length float64) string {
		if len(v) == 0 {
			return fmt.Sprintf("never (0/%d)", n)
		}
		_, med, _ := spread(v)
		share := ""
		if length > 0 {
			share = fmt.Sprintf(", %.0f%% in", 100*med/length)
		}
		return fmt.Sprintf("%s (%d/%d%s)", dur(med), len(v), n, share)
	}
	sb.WriteString("\n### Appease timing per doom\n\nMedian time from the thread's start until level 1 was affordable, and until level 2 was affordable with level 1 bought, over the threads that got there; the share is of the median thread length. Only meaningful under an Appease policy.\n\n")
	sb.WriteString("| doom | threads | thread length (median) | L1 affordable | L2 affordable |\n|---|---|---|---|---|\n")
	for _, k := range keys {
		a := by[k]
		_, length, _ := spread(a.length)
		fmt.Fprintf(sb, "| %s | %d | %s | %s | %s |\n", k, len(a.length), dur(length),
			cell(a.l1, len(a.length), length), cell(a.l2, len(a.length), length))
	}
}
