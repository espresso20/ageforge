package smoke

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
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
	// PacingFailures are the first-cycle ages whose median across seeds left
	// the band, under -pacing enforce (see NewSummary).
	PacingFailures []PacingRow `json:"pacing_failures,omitempty"`
	// QuietFailures are the first-cycle ages whose median longest quiet
	// stretch is over QuietMax, under -pacing enforce.
	QuietFailures []PacingRow `json:"quiet_failures,omitempty"`
	// FirstRun is the first run to the Modern Age (nil when no run got
	// there, as in the fast tier, which stops at the Bronze Age).
	// FirstRunFailed marks a median outside FirstRunLow-FirstRunHigh under
	// -pacing enforce.
	FirstRun       *FirstRunRow `json:"first_run,omitempty"`
	FirstRunFailed bool         `json:"first_run_failed,omitempty"`
	// Early is the Primitive and Stone Ages together on known ground (Era
	// Mastery; nil for a new player), EarlyFailed a median over
	// VeteranEarlyMax under -pacing enforce.
	Early       *EarlyRow `json:"early_known_ground,omitempty"`
	EarlyFailed bool      `json:"early_failed,omitempty"`
	// LaterRun grades push cycles (Config.PushCycles); LaterRunFailed is set
	// when it misses its bar under -pacing enforce.
	LaterRun       *LaterRunRow `json:"later_run,omitempty"`
	LaterRunFailed bool         `json:"later_run_failed,omitempty"`
	// Depth is the depth-pays check (new players' first runs; nil when they
	// reached fewer than two of its ages); DepthFailed marks a deeper
	// prestige paying under DepthPaysMin times the one before, under
	// -pacing enforce.
	Depth       *DepthRow `json:"depth_pays,omitempty"`
	DepthFailed bool      `json:"depth_failed,omitempty"`
}

// FirstRunRow is the 1x time from a fresh start to entering the Modern Age
// (the first prestige) in the first cycle, across seeds, against the band
// FirstRunLow to FirstRunHigh. A run that never got there counts as slower
// than every run that did, so most seeds missing it is a slow median.
type FirstRunRow struct {
	Samples    int     `json:"samples"`
	Reached    int     `json:"reached"`
	MinSecs    float64 `json:"min_seconds"`
	MedianSecs float64 `json:"median_seconds"` // -1 when the median run never got there
	MaxSecs    float64 `json:"max_seconds"`
	Verdict    string  `json:"verdict"`
	// LowSecs and HighSecs are the band the median was held to
	// (firstRunBand; 0 and 0 when reported only).
	LowSecs  float64 `json:"band_low_seconds,omitempty"`
	HighSecs float64 `json:"band_high_seconds,omitempty"`
}

// firstRunToModern is the 1x time run r took from its start to entering the
// Modern Age in its first cycle, and whether it got there: every first-cycle
// age before the first visit to the Modern Age or later, replays (Succumb)
// included.
func firstRunToModern(r *RunResult, order map[string]int) (float64, bool) {
	modern := order[game.PrestigeRunAge]
	secs := 0.0
	for _, a := range r.Ages {
		if a.Cycle != 1 {
			continue
		}
		if order[a.Age] >= modern {
			return secs, true
		}
		secs += a.Seconds
	}
	return secs, false
}

// newFirstRun grades the runs' first runs to the Modern Age; nil when none
// got there.
func newFirstRun(runs []*RunResult, order map[string]int, low, high time.Duration) *FirstRunRow {
	var got []float64
	missed := 0
	for _, r := range runs {
		if secs, ok := firstRunToModern(r, order); ok {
			got = append(got, secs)
		} else {
			missed++
		}
	}
	if len(got) == 0 {
		return nil
	}
	lo, _, hi := spread(got)
	row := &FirstRunRow{Samples: len(runs), Reached: len(got), MinSecs: lo, MaxSecs: hi}
	// The median over every run, a missed one ranking after every reached one.
	all := append([]float64(nil), got...)
	sort.Float64s(all)
	for i := 0; i < missed; i++ {
		all = append(all, -1)
	}
	n := len(all)
	mid := func(i int) float64 { return all[i] }
	switch {
	case n%2 == 1:
		row.MedianSecs = mid(n / 2)
	case mid(n/2) < 0 || mid(n/2-1) < 0:
		row.MedianSecs = -1
	default:
		row.MedianSecs = (mid(n/2-1) + mid(n/2)) / 2
	}
	row.LowSecs, row.HighSecs = low.Seconds(), high.Seconds()
	switch {
	case high <= 0:
		row.Verdict = VerdictNone // reported only (firstRunBand)
	case row.MedianSecs < 0 || row.MedianSecs > high.Seconds():
		row.Verdict = VerdictSlow
	case row.MedianSecs < low.Seconds():
		row.Verdict = VerdictFast
	default:
		row.Verdict = VerdictOK
	}
	return row
}

// writeFirstRun renders the first-run line under a pacing table.
func (s *Summary) writeFirstRun(sb *strings.Builder) {
	f := s.FirstRun
	if f == nil {
		return
	}
	med := "never (most seeds did not get there)"
	if f.MedianSecs >= 0 {
		med = days(f.MedianSecs)
	}
	band := "reported only"
	if f.HighSecs > 0 {
		band = fmt.Sprintf("band %s to %s", days(f.LowSecs), days(f.HighSecs))
	}
	label := "First run"
	if s.Config.Preset != "" {
		label = "The " + s.Config.Preset + " preset's run"
	}
	fmt.Fprintf(sb, "\n%s to the Modern Age: median %s (%s to %s; %d of %d seeds got there), %s: %s.\n",
		label, med, days(f.MinSecs), days(f.MaxSecs), f.Reached, f.Samples, band, verdictMark(f.Verdict))
}

// days formats 1x seconds as days to two decimals ("5.30 days").
func days(secs float64) string {
	return fmt.Sprintf("%.2f days", secs/86400)
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
	Deals       bool    `json:"deals,omitempty"`
	Army        bool    `json:"army,omitempty"`
	Preset      string  `json:"preset,omitempty"`
	PushCycles  bool    `json:"push_cycles,omitempty"`
	Kit         bool    `json:"kit,omitempty"`
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
	// QuietSecs is the median across seeds of the age's longest quiet
	// stretch (AgeSplit.QuietSecs; the longest of a seed's visits).
	QuietSecs float64 `json:"quiet_median_seconds,omitempty"`
	// K is the age's Era Mastery speed (the slowest of the seeds' visits; 0
	// means 1): Ratio and Verdict are against the target ÷ K.
	K float64 `json:"k,omitempty"`
}

// NewSummary aggregates results.
func NewSummary(mode string, cfg Config, started time.Time, runs []*RunResult) *Summary {
	s := &Summary{
		Mode: mode, Started: started, WallMs: time.Since(started).Milliseconds(), Runs: runs,
		Config: ConfigJSON{
			Seeds: cfg.Seeds, Catastrophe: cfg.Catastrophe, Harbinger: cfg.Harbinger, PrestigeAge: cfg.PrestigeAge,
			Cycles: cfg.Cycles, FinalAge: cfg.FinalAge, DecideEvery: cfg.DecideEvery, CheckEvery: cfg.CheckEvery,
			SoftlockSecs: cfg.SoftlockSpan.Seconds(), AgeTimeout: cfg.AgeTimeout.Seconds(), MaxSimSecs: cfg.MaxSim.Seconds(),
			Pacing: cfg.Pacing, Style: cfg.Style, LastPassage: cfg.LastPassage, Deals: cfg.Deals, Army: cfg.Army,
			Preset: cfg.Preset, PushCycles: cfg.PushCycles, Kit: cfg.Kit,
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
	quiet := map[key][]float64{}
	timedOut := map[key]bool{}
	kOf := map[key]float64{}
	for _, r := range runs {
		s.Anomalies += len(r.Anomalies)
		if r.Failed() {
			s.Failed = true
		}
		// Sum repeated visits to an age (Succumb) into one sample per seed;
		// its quiet stretch is the longest of the visits.
		perSeed := map[key][3]float64{}
		for _, a := range r.Ages {
			k := key{a.Cycle, a.Age, a.Unfinished, a.Prestiged}
			v := perSeed[k]
			perSeed[k] = [3]float64{v[0] + a.Seconds, v[1] + float64(a.Ticks), max(v[2], a.QuietSecs)}
			if a.TimedOut {
				timedOut[k] = true
			}
			ak := max(a.K, 1)
			if was, ok := kOf[k]; !ok || ak < was {
				kOf[k] = ak
			}
		}
		for k, v := range perSeed {
			secs[k] = append(secs[k], v[0])
			ticks[k] = append(ticks[k], v[1])
			quiet[k] = append(quiet[k], v[2])
		}
	}
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	for k, v := range secs {
		lo, med, hi := spread(v)
		_, tmed, _ := spread(ticks[k])
		_, qmed, _ := spread(quiet[k])
		row := PacingRow{
			Cycle: k.cycle, Age: k.age, Unfinished: k.open, Prestiged: k.prestiged, Samples: len(v),
			MinSecs: lo, MedianSecs: med, MaxSecs: hi, MedianTicks: int(tmed),
			Verdict: VerdictK(k.age, med, !k.open, kOf[k]), TimedOut: timedOut[k], QuietSecs: qmed,
		}
		if kOf[k] > 1 {
			row.K = kOf[k]
		}
		if k.prestiged {
			row.Verdict = VerdictNone
		}
		if t, ok := Target(k.age); ok {
			row.TargetSecs = t.Seconds()
			row.Ratio = float64(med*max(kOf[k], 1)) / t.Seconds()
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
	// Enforcement grades the median across seeds, like the table: one seed's
	// age can run fast or slow on its events (a lucky Renaissance on a good
	// epoch roll) without the game having changed pace. Only completed ages
	// of the first cycle count; later cycles run on Era Mastery, with the legacy kit.
	if cfg.Pacing == PacingEnforce {
		for _, p := range s.Pacing {
			if p.Cycle == 1 && !p.Unfinished && !p.Prestiged && (p.Verdict == VerdictSlow || p.Verdict == VerdictFast) {
				s.PacingFailures = append(s.PacingFailures, p)
				s.Failed = true
			}
			if p.Cycle == 1 && !p.Unfinished && !p.Prestiged && quietGraded(p.Age) && p.QuietSecs > QuietMax.Seconds() {
				s.QuietFailures = append(s.QuietFailures, p)
				s.Failed = true
			}
		}
	}
	// The first run to the Modern Age, graded like an age: on its median,
	// and failing the set only under enforce (only the progression scenario
	// runs paced; the styles scenario reports it).
	lo, hi, graded := firstRunBand(cfg)
	if !graded {
		lo, hi = 0, 0
	}
	s.FirstRun = newFirstRun(runs, order, lo, hi)
	if s.FirstRun != nil && cfg.Pacing == PacingEnforce && graded && s.FirstRun.Verdict != VerdictOK {
		s.FirstRunFailed, s.Failed = true, true
	}
	// Era Mastery: the known-ground Primitive and Stone Ages together, and
	// the push cycles.
	s.Early = newEarly(runs)
	if s.Early != nil && cfg.Pacing == PacingEnforce && s.Early.Verdict != VerdictOK {
		s.EarlyFailed, s.Failed = true, true
	}
	if cfg.PushCycles && cfg.Cycles > 1 {
		s.LaterRun = newLaterRun(runs, order)
		if s.LaterRun != nil && cfg.Pacing == PacingEnforce && s.LaterRun.Failed {
			s.LaterRunFailed, s.Failed = true, true
		}
	}
	// Depth pays: a deeper prestige pays more per day.
	s.Depth = newDepth(cfg, runs, order)
	if s.Depth != nil && cfg.Pacing == PacingEnforce && s.Depth.Failed {
		s.DepthFailed, s.Failed = true, true
	}
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
		s.Mode, len(s.Runs), s.Config.Catastrophe, orDefault(s.Config.PrestigeAge, game.PrestigeRunAge), s.Config.Cycles)
	if s.Config.FinalAge != "" {
		fmt.Fprintf(&sb, ", then on to `%s`", s.Config.FinalAge)
	}
	if s.Config.Deals {
		sb.WriteString(", faction deals on")
	}
	if s.Config.Army {
		sb.WriteString(", garrison kept")
	}
	fmt.Fprintf(&sb, ". Took %s of wall time.\n\n", time.Duration(s.WallMs)*time.Millisecond)
	sb.WriteString("Times are simulated wall-clock at 1x speed (tick_speed bonuses included). ")
	if s.Config.Pacing == PacingEnforce {
		sb.WriteString("Pacing is enforced: a first-cycle age whose median across seeds is outside its target band, or (up to the " + QuietLastAge + ") whose median longest quiet stretch is over " + dur(QuietMax.Seconds()) + ", fails the set, and any age past its timeout fails its run (later cycles and ages left by prestige are graded only), as do panics, soft-locks and invariant violations.\n\n")
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
	fmt.Fprintf(&sb, "Time spent in each age, from entering it to entering the next, across seeds, against the target in smoke/targets.go (pass: %gx to %gx the target, %s; the verdict grades the median). The longest quiet stretch is the median of each seed's longest stretch in the age with no new building type built and no tech finished; under enforce a first-cycle age over %s fails, up to the %s (later ages are reported only).\n\n", PacingLow, PacingHigh, highForText(), dur(QuietMax.Seconds()), QuietLastAge)
	s.writePacingTable(&sb)
	s.writeAgeGates(&sb)
	s.writeMarket(&sb)
	s.writeQuiet(&sb)
	s.writeFirstRun(&sb)
	s.writeEarly(&sb)
	s.writeLaterRun(&sb)
	s.writeDepth(&sb)
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
	s.writeDefense(&sb)
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
	// What the rejected actions were told (the first few of each run).
	told := false
	for _, r := range s.Runs {
		for _, f := range r.Stats.Refusals {
			if !told {
				sb.WriteString("\nWhat the rejected actions were told (the first of each run):\n\n")
				told = true
			}
			fmt.Fprintf(&sb, "- seed %d, %s, tick %d: `%s %s`: %s\n", r.Seed, f.Age, f.Tick, f.Kind, f.Detail, f.Message)
		}
	}

	s.writeFates(&sb)
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
	sb.WriteString("| cycle | age | seeds | min | median | max | target | ratio | verdict | longest quiet |\n|---|---|---|---|---|---|---|---|---|---|\n")
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
			if p.K > 1 {
				// Era Mastery: the target this age is held to is ÷ k.
				target = fmt.Sprintf("%s (÷%s)", dur(p.TargetSecs/p.K), game.SpeedText(p.K))
			}
			ratio = fmt.Sprintf("%.2gx", p.Ratio)
		}
		fmt.Fprintf(sb, "| %d | %s | %d | %s | %s | %s | %s | %s | %s | %s |\n", p.Cycle, age, p.Samples,
			dur(p.MinSecs), dur(p.MedianSecs), dur(p.MaxSecs), target, ratio, verdictMark(p.Verdict), quietMark(p))
	}
}

// writeAgeGates is what each first-cycle age waited for: when each thing the
// advance asks for was first in place (AgeSplit.Gates), as the median across
// seeds of its share of the age, and how much of the age's knowledge the
// market sold. Written only when some run recorded gates.
func (s *Summary) writeAgeGates(sb *strings.Builder) {
	type row struct {
		tech, funded, built, blds, res, bought []float64
		faith, culture                         []float64
		last                                   map[string]int
	}
	rows := map[string]*row{}
	any := false
	for _, r := range s.Runs {
		for _, a := range r.Ages {
			if a.Cycle != 1 || a.Gates == nil || a.Ticks <= 0 || a.Unfinished || a.Prestiged {
				continue
			}
			any = true
			w := rows[a.Age]
			if w == nil {
				w = &row{last: map[string]int{}}
				rows[a.Age] = w
			}
			g := a.Gates
			share := func(dst *[]float64, at int) {
				if at >= 0 {
					*dst = append(*dst, float64(at)/float64(a.Ticks))
				}
			}
			share(&w.tech, g.WonderTech)
			share(&w.funded, g.WonderFunded)
			share(&w.built, g.WonderBuilt)
			share(&w.blds, g.Buildings)
			share(&w.res, g.Resources)
			if total := a.KnowledgeBought + a.KnowledgeMade; total > 0 {
				w.bought = append(w.bought, a.KnowledgeBought/total)
			}
			w.faith, w.culture = append(w.faith, a.FaithStrength), append(w.culture, a.CultureStrength)
			// What came last. A wonder is named with what held its start
			// back: its keystone tech or its price.
			last := "the wonder, once paid for"
			if g.WonderTech > g.WonderFunded {
				last = "the wonder, once its keystone was known"
			}
			if g.Buildings > g.WonderBuilt {
				last = "buildings"
			}
			if g.Resources > max(g.WonderBuilt, g.Buildings) {
				last = "resources"
			}
			w.last[last]++
		}
	}
	if !any {
		return
	}
	sb.WriteString("\nWhat each age of the first cycle waited for: when each thing the advance asks for was first in place, as a share of the time the age took (median across seeds; - is never, or an age without one). The wonder's build runs from the later of its keystone tech and its price to the wonder standing. Knowledge bought is the share of the age's knowledge the market sold. Faith and culture strength are what the rolls would have read as the age ended (faith: under a quarter is the bottom band; culture: over two fifths opens Major good epoch events).\n\n")
	sb.WriteString("| age | keystone tech | wonder paid for | wonder built | buildings | resources | waited longest for | knowledge bought | faith strength | culture strength |\n|---|---|---|---|---|---|---|---|---|---|\n")
	pct := func(v []float64) string {
		if len(v) == 0 {
			return "-"
		}
		_, med, _ := spread(v)
		return fmt.Sprintf("%.0f%%", med*100)
	}
	for _, age := range config.AgeOrder() {
		w := rows[age]
		if w == nil {
			continue
		}
		var last []string
		for _, k := range sortedKeys(w.last) {
			last = append(last, fmt.Sprintf("%s (%d)", k, w.last[k]))
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", age, pct(w.tech), pct(w.funded), pct(w.built), pct(w.blds), pct(w.res), strings.Join(last, ", "), pct(w.bought), pct(w.faith), pct(w.culture))
	}
}

// writeMarket is the market's part in each first-cycle age: the resources
// an age got mostly at the market (bought, as a share of what it had of
// them: made plus bought) and what paid for them (sold, as a share of what
// it made of them), as medians across seeds. Only shares of a tenth or more
// are listed, largest first. Written only when some run recorded them.
func (s *Summary) writeMarket(sb *strings.Builder) {
	bought, sold := map[string]map[string][]float64{}, map[string]map[string][]float64{}
	seeds := map[string]int{}
	for _, r := range s.Runs {
		for _, a := range r.Ages {
			if a.Cycle != 1 || a.Made == nil || a.Unfinished || a.Prestiged {
				continue
			}
			seeds[a.Age]++
			if bought[a.Age] == nil {
				bought[a.Age], sold[a.Age] = map[string][]float64{}, map[string][]float64{}
			}
			for res, v := range a.Bought {
				bought[a.Age][res] = append(bought[a.Age][res], v/(v+a.Made[res]))
			}
			for res, v := range a.Sold {
				if made := a.Made[res]; made > 0 {
					sold[a.Age][res] = append(sold[a.Age][res], v/made)
				}
			}
		}
	}
	if len(seeds) == 0 {
		return
	}
	list := func(shares map[string][]float64, n int) string {
		type item struct {
			res string
			v   float64
		}
		var items []item
		for res, v := range shares {
			// A resource fewer than half the seeds traded has no median.
			if 2*len(v) < n {
				continue
			}
			if _, med, _ := spread(v); med >= 0.1 {
				items = append(items, item{res, med})
			}
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i].v != items[j].v {
				return items[i].v > items[j].v
			}
			return items[i].res < items[j].res
		})
		var parts []string
		for _, it := range items {
			parts = append(parts, fmt.Sprintf("%s %.0f%%", it.res, it.v*100))
		}
		return orDefault(strings.Join(parts, ", "), "-")
	}
	sb.WriteString("\nThe market in each age of the first cycle (median across seeds; shares of a tenth or more). Bought: the share of what the age had of a resource (made plus bought) that the market sold it. Sold: what the age gave the market, as a share of what it made of that resource (over 100% is stock from before).\n\n")
	sb.WriteString("| age | bought | sold |\n|---|---|---|\n")
	for _, age := range config.AgeOrder() {
		if seeds[age] == 0 {
			continue
		}
		fmt.Fprintf(sb, "| %s | %s | %s |\n", age, list(bought[age], seeds[age]), list(sold[age], seeds[age]))
	}
}

// writeQuiet says what each first-cycle age's longest quiet stretch was made
// of (AgeSplit.Quiet), as medians across seeds: whether the player had
// anything left to open when it began, and what there was to do in it.
// Written only when some run recorded one.
func (s *Summary) writeQuiet(sb *strings.Builder) {
	type row struct {
		secs, open, left, types, built, researching, affordable, banking, nothing []float64
	}
	rows := map[string]*row{}
	for _, r := range s.Runs {
		for _, a := range r.Ages {
			q := a.Quiet
			if a.Cycle != 1 || q == nil || q.Samples == 0 || a.Unfinished || a.Prestiged {
				continue
			}
			w := rows[a.Age]
			if w == nil {
				w = &row{}
				rows[a.Age] = w
			}
			n := float64(q.Samples)
			w.secs = append(w.secs, a.QuietSecs)
			w.open, w.left, w.types = append(w.open, float64(q.TechsOpen)), append(w.left, float64(q.TechsLeft)), append(w.types, float64(q.TypesLeft))
			w.built = append(w.built, float64(q.Built))
			w.researching, w.affordable = append(w.researching, float64(q.Researching)/n), append(w.affordable, float64(q.Affordable)/n)
			w.banking, w.nothing = append(w.banking, float64(q.Banking)/n), append(w.nothing, float64(q.Nothing)/n)
		}
	}
	if len(rows) == 0 {
		return
	}
	med := func(v []float64) float64 { _, m, _ := spread(v); return m }
	sb.WriteString("\nWhat the longest quiet stretch of each first-cycle age was made of (median across seeds). When it began: the techs that could be researched, the age's techs not yet researched, and the age's building types never built. In it: the buildings finished (copies of types already seen), and the share of the bot's decisions at which a tech was being researched, at which some building of the age or open tech could be paid for at once, at which the wonder had a price left and something in store to pay toward it, and at which none of that was true and nothing was under construction (the player could start nothing).\n\n")
	sb.WriteString("| age | longest quiet | open techs | age's techs left | building types left | buildings finished | researching | something affordable | wonder to pay toward | nothing to start |\n|---|---|---|---|---|---|---|---|---|---|\n")
	for _, age := range config.AgeOrder() {
		w := rows[age]
		if w == nil {
			continue
		}
		fmt.Fprintf(sb, "| %s | %s | %.0f | %.0f | %.0f | %.0f | %.0f%% | %.0f%% | %.0f%% | %.0f%% |\n", age, dur(med(w.secs)), med(w.open), med(w.left), med(w.types), med(w.built),
			med(w.researching)*100, med(w.affordable)*100, med(w.banking)*100, med(w.nothing)*100)
	}
}

// quietMark renders a row's longest quiet stretch, flagged when it is over
// QuietMax.
func quietMark(p PacingRow) string {
	if p.QuietSecs > QuietMax.Seconds() && !p.Prestiged && quietGraded(p.Age) {
		return dur(p.QuietSecs) + " (over " + dur(QuietMax.Seconds()) + ")"
	}
	return dur(p.QuietSecs)
}

// highForText lists the ages PacingHighFor holds to a tighter band.
func highForText() string {
	var parts []string
	for _, a := range sortedKeys(PacingHighFor) {
		parts = append(parts, fmt.Sprintf("%s at most %gx", a, PacingHighFor[a]))
	}
	return strings.Join(parts, ", ")
}

func countTotal(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

// writeDefense is the army table: what each run lost to Endure and what its
// garrison saved. Written only when some run endured a catastrophe or had a
// garrison blunt a raid.
func (s *Summary) writeDefense(sb *strings.Builder) {
	any := false
	for _, r := range s.Runs {
		if r.Stats.CatastrophesEndured > 0 || r.Stats.RaidsBlunted > 0 {
			any = true
		}
	}
	if !any {
		return
	}
	sb.WriteString("\n## Army\n\n| seed | endured | buildings lost to Endure | mean stock kept | mean garrison share | buildings saved | raids blunted | workers saved |\n|---|---|---|---|---|---|---|---|\n")
	for _, r := range s.Runs {
		st := r.Stats
		kept, guard := "-", "-"
		if n := float64(st.CatastrophesEndured); n > 0 {
			kept = fmt.Sprintf("%.1f%%", st.EndureStockKept/n*100)
			guard = fmt.Sprintf("%.1f%%", st.EndureGarrison/n*100)
		}
		fmt.Fprintf(sb, "| %d | %d | %d | %s | %s | %d | %d | %d |\n", r.Seed, st.CatastrophesEndured, st.EndureBuildingsLost,
			kept, guard, st.DefenseBuildingsSaved, st.RaidsBlunted, st.DefenseWorkersSaved)
	}
}
