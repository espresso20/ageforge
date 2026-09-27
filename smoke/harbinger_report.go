package smoke

import (
	"fmt"
	"strings"
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
	sb.WriteString("\nThread counts are keyed by the epoch the harbinger warns about. A false prophet is only revealed when its thread resolves; the live view hides it.\n")

	sb.WriteString("\n### Level-1 prices against the most storage buildable by the passage\n\n")
	sb.WriteString("Read from a harbinger summoned on a scratch engine (never the played one). \"Max storage\" assumes every capped storage building up to the epoch's last age was built out, plus storage techs.\n\n")
	sb.WriteString("| passage | Appease L1 | Brace L1 | max storage by last age | can ever hold |\n|---|---|---|---|---|\n")
	for _, p := range s.Prices {
		var over []string
		for _, c := range []map[string]float64{p.AppeaseL1, p.BraceL1} {
			for _, res := range sortedKeys(c) {
				if m := p.MaxStorage[res]; m >= 0 && c[res] > m {
					over = append(over, fmt.Sprintf("no: %s %s > %s", res, num(c[res]), num(p.MaxStorage[res])))
				}
			}
		}
		fmt.Fprintf(sb, "| %s → %s (%s) | %s | %s | %s | %s |\n", p.Epoch, p.TargetEpoch, p.LastAge,
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
		fmt.Fprintf(sb, "\n### %s affordability per thread\n\nStock/storage when the thread started and at the last decision before it resolved (or the run ended).\n\n", kind)
		sb.WriteString("| seed | cycle | passage | speakers | outcome | L1 cost | at start | at passage | ever fits storage | ever affordable | bought |\n|---|---|---|---|---|---|---|---|---|---|---|\n")
		for _, th := range threads {
			cost, fits, afford, after, lvl := th.AppeaseL1, th.AppeaseFits, th.AppeaseAffordable, th.AppeaseAfterSecs, th.AppeaseLevel
			if kind == "Brace" {
				cost, fits, afford, after, lvl = th.BraceL1, th.BraceFits, th.BraceAffordable, th.BraceAfterSecs, th.BraceLevel
			}
			outcome := th.Outcome
			if th.FalseProphet != nil && *th.FalseProphet {
				outcome += " (false prophet)"
			}
			fmt.Fprintf(sb, "| %d | %d | %s → %s | %s | %s | %s | %s | %s | %v | %s | %d |\n", seedOf[th], th.Cycle, th.Epoch, th.TargetEpoch,
				strings.Join(th.Speakers, ", "), outcome, orDefault(costStr(cost), "-"),
				pointsStr(th.AtStart, cost), pointsStr(th.AtPassage, cost), fits, affordStr(afford, after), lvl)
		}
	}

	s.writeAppeaseTiming(sb, threads)

	// Flag passages where no seed could ever afford an answer.
	type agg struct{ seen, appease, brace int }
	byPassage := map[string]*agg{}
	for _, th := range threads {
		k := th.Epoch + " → " + th.TargetEpoch
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
		k := th.Epoch + " → " + th.TargetEpoch
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
	sb.WriteString("\n### Appease timing per passage\n\nMedian time from the thread's start until level 1 was affordable, and until level 2 was affordable with level 1 bought, over the threads that got there; the share is of the median thread length. Only meaningful under an Appease policy.\n\n")
	sb.WriteString("| passage | threads | thread length (median) | L1 affordable | L2 affordable |\n|---|---|---|---|---|\n")
	for _, k := range keys {
		a := by[k]
		_, length, _ := spread(a.length)
		fmt.Fprintf(sb, "| %s | %d | %s | %s | %s |\n", k, len(a.length), dur(length),
			cell(a.l1, len(a.length), length), cell(a.l2, len(a.length), length))
	}
}
