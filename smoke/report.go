package smoke

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Summary is a whole session: the config and every seed's result.
type Summary struct {
	Mode      string        `json:"mode"`
	Started   time.Time     `json:"started"`
	WallMs    int64         `json:"wall_ms"`
	Config    ConfigJSON    `json:"config"`
	Runs      []*RunResult  `json:"runs"`
	Pacing    []PacingRow   `json:"pacing"`
	Gates     []GateProblem `json:"static_gate_problems"`
	Slack     []GateSlack   `json:"static_gate_slack"`
	Prices    []PriceRow    `json:"harbinger_prices"`
	Failed    bool          `json:"failed"`
	Anomalies int           `json:"anomaly_count"`
}

// ConfigJSON is Config with durations in seconds.
type ConfigJSON struct {
	Seeds        []int64 `json:"seeds"`
	Catastrophe  string  `json:"catastrophe"`
	Harbinger    string  `json:"harbinger_policy"`
	PrestigeAge  string  `json:"prestige_age"`
	Cycles       int     `json:"cycles"`
	FinalAge     string  `json:"final_age"`
	DecideEvery  int     `json:"decide_every_ticks"`
	CheckEvery   int     `json:"check_every_ticks"`
	SoftlockSecs float64 `json:"softlock_seconds"`
	// AgeTimeout is 0 when each age's timeout comes from the pacing table.
	AgeTimeout  float64 `json:"age_timeout_seconds"`
	MaxSimSecs  float64 `json:"max_sim_seconds"`
	Pacing      string  `json:"pacing"`
	Style       string  `json:"style,omitempty"`
	LastPassage string  `json:"last_passage,omitempty"`
}

// PacingRow aggregates one (cycle, age) across seeds. Times are 1x seconds.
type PacingRow struct {
	Cycle       int     `json:"cycle"`
	Age         string  `json:"age"`
	Samples     int     `json:"samples"`
	MinSecs     float64 `json:"min_seconds"`
	MedianSecs  float64 `json:"median_seconds"`
	MaxSecs     float64 `json:"max_seconds"`
	MedianTicks int     `json:"median_ticks"`
	Unfinished  bool    `json:"unfinished,omitempty"`
	// Prestiged rows are ages left by prestige (AgeSplit.Prestiged): shown,
	// never graded.
	Prestiged bool `json:"prestiged,omitempty"`
	// TargetSecs is the pacing target, Ratio the median over it, Verdict
	// the median graded against the band (see targets.go).
	TargetSecs float64 `json:"target_seconds,omitempty"`
	Ratio      float64 `json:"ratio,omitempty"`
	Verdict    string  `json:"verdict,omitempty"`
	TimedOut   bool    `json:"timed_out,omitempty"`
}

// NewSummary aggregates results.
func NewSummary(mode string, cfg Config, started time.Time, runs []*RunResult) *Summary {
	s := &Summary{
		Mode: mode, Started: started, WallMs: time.Since(started).Milliseconds(), Runs: runs,
		Config: ConfigJSON{
			Seeds: cfg.Seeds, Catastrophe: cfg.Catastrophe, Harbinger: cfg.Harbinger, PrestigeAge: cfg.PrestigeAge,
			Cycles: cfg.Cycles, FinalAge: cfg.FinalAge, DecideEvery: cfg.DecideEvery, CheckEvery: cfg.CheckEvery,
			SoftlockSecs: cfg.SoftlockSpan.Seconds(), AgeTimeout: cfg.AgeTimeout.Seconds(), MaxSimSecs: cfg.MaxSim.Seconds(),
			Pacing: cfg.Pacing, Style: cfg.Style, LastPassage: cfg.LastPassage,
		},
	}
	type key struct {
		cycle     int
		age       string
		open      bool
		prestiged bool
	}
	s.Prices = HarbingerPrices()
	s.Gates, s.Slack = StaticGates()
	secs := map[key][]float64{}
	ticks := map[key][]float64{}
	timedOut := map[key]bool{}
	for _, r := range runs {
		s.Anomalies += len(r.Anomalies)
		if r.Failed() {
			s.Failed = true
		}
		// Sum repeated visits to an age (Succumb) into one sample per seed.
		perSeed := map[key][2]float64{}
		for _, a := range r.Ages {
			k := key{a.Cycle, a.Age, a.Unfinished, a.Prestiged}
			v := perSeed[k]
			perSeed[k] = [2]float64{v[0] + a.Seconds, v[1] + float64(a.Ticks)}
			if a.TimedOut {
				timedOut[k] = true
			}
		}
		for k, v := range perSeed {
			secs[k] = append(secs[k], v[0])
			ticks[k] = append(ticks[k], v[1])
		}
	}
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	for k, v := range secs {
		lo, med, hi := spread(v)
		_, tmed, _ := spread(ticks[k])
		row := PacingRow{
			Cycle: k.cycle, Age: k.age, Unfinished: k.open, Prestiged: k.prestiged, Samples: len(v),
			MinSecs: lo, MedianSecs: med, MaxSecs: hi, MedianTicks: int(tmed),
			Verdict: Verdict(k.age, med, !k.open), TimedOut: timedOut[k],
		}
		if k.prestiged {
			row.Verdict = VerdictNone
		}
		if t, ok := Target(k.age); ok {
			row.TargetSecs = t.Seconds()
			row.Ratio = med / t.Seconds()
		}
		s.Pacing = append(s.Pacing, row)
	}
	sort.Slice(s.Pacing, func(i, j int) bool {
		a, b := s.Pacing[i], s.Pacing[j]
		if a.Cycle != b.Cycle {
			return a.Cycle < b.Cycle
		}
		if a.Age != b.Age {
			return order[a.Age] < order[b.Age]
		}
		// Completed, then left by prestige, then unfinished.
		rank := func(p PacingRow) int {
			switch {
			case p.Unfinished:
				return 2
			case p.Prestiged:
				return 1
			}
			return 0
		}
		return rank(a) < rank(b)
	})
	return s
}

func spread(v []float64) (lo, med, hi float64) {
	if len(v) == 0 {
		return 0, 0, 0
	}
	c := append([]float64(nil), v...)
	sort.Float64s(c)
	n := len(c)
	med = c[n/2]
	if n%2 == 0 {
		med = (c[n/2-1] + c[n/2]) / 2
	}
	return c[0], med, c[n-1]
}

// WriteJSON writes the machine-readable report.
func (s *Summary) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// dur formats 1x seconds as a short human duration.
func dur(secs float64) string {
	if secs <= 0 {
		return "0s"
	}
	d := time.Duration(secs * float64(time.Second))
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%.1fd", d.Hours()/24)
	case d >= time.Hour:
		return fmt.Sprintf("%.1fh", d.Hours())
	case d >= time.Minute:
		return fmt.Sprintf("%.0fm", d.Minutes())
	default:
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}

// WriteMarkdown writes the human-readable report.
func (s *Summary) WriteMarkdown(w io.Writer) error {
	var sb strings.Builder
	verdict := "PASS"
	if s.Failed {
		verdict = "FAIL"
	}
	fmt.Fprintf(&sb, "# AgeForge smoke report: %s\n\n", verdict)
	fmt.Fprintf(&sb, "Mode `%s`, %d seed(s), catastrophe choice `%s`, prestige at `%s`, %d prestige cycle(s)",
		s.Mode, len(s.Runs), s.Config.Catastrophe, orDefault(s.Config.PrestigeAge, "first allowed age"), s.Config.Cycles)
	if s.Config.FinalAge != "" {
		fmt.Fprintf(&sb, ", then on to `%s`", s.Config.FinalAge)
	}
	fmt.Fprintf(&sb, ". Took %s of wall time.\n\n", time.Duration(s.WallMs)*time.Millisecond)
	sb.WriteString("Times are simulated wall-clock at 1x speed (tick_speed bonuses included). ")
	if s.Config.Pacing == PacingEnforce {
		sb.WriteString("Pacing is enforced: a first-cycle age outside its target band, or any age past its timeout, fails the run (later cycles and ages left by prestige are graded only), as do panics, soft-locks and invariant violations.\n\n")
	} else {
		sb.WriteString("Pacing is report-only: ages are graded against their targets but never fail the run; panics, soft-locks and invariant violations do.\n\n")
	}

	sb.WriteString("## Runs\n\n| seed | outcome | reached | ticks | 1x time | cycles to prestige | anomalies | wall |\n|---|---|---|---|---|---|---|---|\n")
	for _, r := range s.Runs {
		var cyc []string
		for _, c := range r.Cycles {
			cyc = append(cyc, dur(c.Seconds))
		}
		fmt.Fprintf(&sb, "| %d | %s | %s | %d | %s | %s | %d | %s |\n", r.Seed, r.Outcome, r.FinalAge, r.Ticks, dur(r.Seconds),
			orDefault(strings.Join(cyc, ", "), "-"), len(r.Anomalies), time.Duration(r.WallMillis)*time.Millisecond)
	}

	sb.WriteString("\n## Pacing per age\n\n")
	fmt.Fprintf(&sb, "Time spent in each age, from entering it to entering the next, across seeds, against the target in smoke/targets.go (pass: %gx to %gx the target; the verdict grades the median).\n\n", PacingLow, PacingHigh)
	s.writePacingTable(&sb)
	for _, r := range s.Runs {
		for _, n := range r.Notes {
			fmt.Fprintf(&sb, "- seed %d: %s\n", r.Seed, n)
		}
	}
	var cyc []float64
	for _, r := range s.Runs {
		for _, c := range r.Cycles {
			if c.Cycle == 1 {
				cyc = append(cyc, c.Seconds)
			}
		}
	}
	if len(cyc) > 0 {
		lo, med, hi := spread(cyc)
		fmt.Fprintf(&sb, "\nFirst prestige: min %s, median %s, max %s (%d of %d seeds got there).\n", dur(lo), dur(med), dur(hi), len(cyc), len(s.Runs))
	} else {
		sb.WriteString("\nNo seed reached prestige.\n")
	}

	s.writeGates(&sb)

	sb.WriteString("\n## Events\n\n| seed | catastrophes rolled | endured | succumbed | epoch events | awakenings | milestones | techs | buildings | timed events | starvation deaths |\n|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range s.Runs {
		st := r.Stats
		fmt.Fprintf(&sb, "| %d | %d | %d | %d | %s | %d | %d | %d | %d | %d | %d |\n", r.Seed,
			st.CatastrophesRolled, st.CatastrophesEndured, st.CatastrophesSuccumbed, countStr(st.EpochEvents),
			st.Awakenings, st.Milestones, st.TechsResearched, st.BuildingsCompleted, countTotal(st.TimedEvents), st.StarvationDeaths)
	}
	sb.WriteString("\nBot actions (successful / rejected), summed over seeds:\n\n")
	acts, errs := map[string]int{}, map[string]int{}
	for _, r := range s.Runs {
		for k, v := range r.Stats.Actions {
			acts[k] += v
		}
		for k, v := range r.Stats.ActionErrors {
			errs[k] += v
		}
	}
	all := map[string]bool{}
	for k := range acts {
		all[k] = true
	}
	for k := range errs {
		all[k] = true
	}
	for _, k := range sortedKeys(all) {
		fmt.Fprintf(&sb, "- `%s`: %d / %d\n", k, acts[k], errs[k])
	}

	s.writeHarbingers(&sb)

	sb.WriteString("\n## Anomalies\n\n")
	if s.Anomalies == 0 {
		sb.WriteString("None.\n")
	}
	for _, r := range s.Runs {
		for _, a := range r.Anomalies {
			fmt.Fprintf(&sb, "### %s: %s (seed %d, cycle %d, %s, tick %d, seen %dx)\n\n%s\n\n", a.Kind, a.Check, a.Seed, a.Cycle, a.Age, a.Tick, a.Count, a.Message)
			if a.Dump != "" {
				fmt.Fprintf(&sb, "<details><summary>state dump</summary>\n\n```\n%s\n```\n\n</details>\n\n", strings.TrimRight(a.Dump, "\n"))
			}
		}
	}
	_, err := io.WriteString(w, sb.String())
	return err
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func countStr(m map[string]int) string {
	if len(m) == 0 {
		return "0"
	}
	var parts []string
	for _, k := range sortedKeys(m) {
		parts = append(parts, fmt.Sprintf("%s %d", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

// writePacingTable renders the per-age pacing rows with their targets and
// verdicts.
func (s *Summary) writePacingTable(sb *strings.Builder) {
	sb.WriteString("| cycle | age | seeds | min | median | max | target | ratio | verdict |\n|---|---|---|---|---|---|---|---|---|\n")
	for _, p := range s.Pacing {
		age := p.Age
		if p.Unfinished {
			age += " (unfinished)"
		}
		if p.Prestiged {
			age += " (left by prestige)"
		}
		if p.TimedOut {
			age += " (past timeout)"
		}
		target, ratio := "-", "-"
		if p.TargetSecs > 0 && !p.Prestiged {
			target = dur(p.TargetSecs)
			ratio = fmt.Sprintf("%.2gx", p.Ratio)
		}
		fmt.Fprintf(sb, "| %d | %s | %d | %s | %s | %s | %s | %s | %s |\n", p.Cycle, age, p.Samples,
			dur(p.MinSecs), dur(p.MedianSecs), dur(p.MaxSecs), target, ratio, verdictMark(p.Verdict))
	}
}

func countTotal(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
