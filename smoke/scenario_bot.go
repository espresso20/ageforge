package smoke

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Overrides are command-line settings that replace a bot scenario's tier
// defaults (progression only; zero values keep the default).
type Overrides struct {
	MaxSim      time.Duration
	StopAge     string
	PrestigeAge string
	Cycles      int
	// FinalAge applies when FinalAgeSet ("" then means stop at prestige).
	FinalAge    string
	FinalAgeSet bool
}

// runBotSet plays cfg on every seed and returns the aggregated summary,
// with each run's anomalies folded into res as failures. name labels the
// per-set report file (<out>/<name>.md) and the repro command.
func runBotSet(e *Env, res *Result, name, scenario string, cfg Config, seeds []int64) *Summary {
	cfg.Seeds = seeds
	started := time.Now()
	runs := runSeeds(e, seeds, func(seed int64) *RunResult {
		r := Run(cfg, seed)
		e.logf("  %s seed %d: %s at %s after %d ticks (%s at 1x), %d anomalies, %s wall", name, seed, r.Outcome, r.FinalAge,
			r.Ticks, dur(r.Seconds), len(r.Anomalies), (time.Duration(r.WallMillis) * time.Millisecond).Round(100*time.Millisecond))
		return r
	})
	return foldBotSet(e, res, name, scenario, cfg, started, runs)
}

// foldBotSet grades a set of finished runs (NewSummary) and folds their
// pacing failures and anomalies into res; see runBotSet.
func foldBotSet(e *Env, res *Result, name, scenario string, cfg Config, started time.Time, runs []*RunResult) *Summary {
	sum := NewSummary(name, cfg, started, runs)
	for _, p := range sum.PacingFailures {
		res.fail(KindPacing+"/pacing_"+p.Verdict, "%s: cycle 1 %s took %s (median of %d seeds, %s to %s) against a %s target (band %gx to %gx)",
			name, p.Age, dur(p.MedianSecs), p.Samples, dur(p.MinSecs), dur(p.MaxSecs), dur(p.TargetSecs), PacingLow, pacingHigh(p.Age))
	}
	for _, p := range sum.QuietFailures {
		res.fail(KindPacing+"/quiet_stretch", "%s: cycle 1 %s went %s with no new building type or tech (median of %d seeds), over the %s limit",
			name, p.Age, dur(p.QuietSecs), p.Samples, dur(QuietMax.Seconds()))
	}
	if f := sum.FirstRun; sum.FirstRunFailed && f != nil {
		med := "never, for most seeds"
		if f.MedianSecs >= 0 {
			med = days(f.MedianSecs)
		}
		res.fail(KindPacing+"/first_run_"+f.Verdict, "%s: the first run to the Modern Age took %s (median of %d seeds, %d got there, %s to %s), outside %s to %s",
			name, med, f.Samples, f.Reached, days(f.MinSecs), days(f.MaxSecs), days(f.LowSecs), days(f.HighSecs))
	}
	if x := sum.Early; sum.EarlyFailed && x != nil {
		res.fail(KindPacing+"/known_ground_early", "%s: the Primitive and Stone Ages on known ground took %s together (median of %d seeds), over the %s limit",
			name, dur(x.MedianSecs), x.Samples, dur(VeteranEarlyMax.Seconds()))
	}
	if l := sum.LaterRun; sum.LaterRunFailed && l != nil {
		res.fail(KindPacing+"/later_run", "%s: cycle 2 covered cycle 1's ages %.2fx faster (median; want %gx or more) and ended %+.1f ages deeper (median; want %+d or more)",
			name, l.MedianSpeedup, LaterRunMinSpeedup, l.MedianDepth, LaterRunMinDepth)
	}
	if d := sum.Depth; sum.DepthFailed && d != nil {
		var parts []string
		for _, p := range d.Points {
			parts = append(parts, fmt.Sprintf("%s %.1f", p.Age, p.PerDay))
		}
		res.fail(KindPacing+"/depth_pays", "%s: points per day by prestige age (%s) do not rise by %gx at each step deeper",
			name, strings.Join(parts, ", "), DepthPaysMin)
	}
	for _, r := range runs {
		for _, a := range r.Anomalies {
			f := res.fail(a.Kind+"/"+a.Check, "%s: %s (cycle %d, %s, tick %d, seen %dx)", name, a.Message, a.Cycle, a.Age, a.Tick, a.Count)
			f.Seed = a.Seed
			f.Repro = reproCmd(e.Tier, scenario, cfg, a.Seed)
			f.Detail = a.Dump
		}
	}
	if e.OutDir != "" {
		if err := writeFileFunc(filepath.Join(e.OutDir, name+".md"), sum.WriteMarkdown); err != nil {
			res.warn("report", "writing %s.md: %v", name, err)
		}
	}
	res.Progression = append(res.Progression, sum)
	return sum
}

// reproCmd is the command line that replays one seed of a bot scenario.
func reproCmd(tier, scenario string, cfg Config, seed int64) string {
	parts := []string{"go run ./cmd/smoke"}
	if tier != "" && tier != TierFast {
		parts = append(parts, "-tier "+tier)
	}
	parts = append(parts, "-scenario "+scenario, fmt.Sprintf("-seed-base %d -seeds 1", seed))
	if cfg.Catastrophe != "" && cfg.Catastrophe != "endure" {
		parts = append(parts, "-catastrophe "+cfg.Catastrophe)
	}
	if cfg.Harbinger != "" && cfg.Harbinger != HarbingerIgnore {
		parts = append(parts, "-harbinger "+cfg.Harbinger)
	}
	if cfg.StopAge != "" {
		parts = append(parts, "-stop-age "+cfg.StopAge)
	}
	if cfg.Style != "" && cfg.Style != "greedy" {
		parts = append(parts, "-style "+cfg.Style)
	}
	if cfg.CheckIn > 0 && cfg.CheckIn != IdleCheckIn {
		parts = append(parts, "-check-in "+cfg.CheckIn.String())
	}
	if cfg.Preset != "" && scenario != "veteran" {
		parts = append(parts, "-preset "+cfg.Preset)
	}
	if cfg.Kit && scenario != "idle" {
		parts = append(parts, "-kit")
	}
	parts = append(parts, "-trace -v")
	return strings.Join(parts, " ")
}

// writeFileFunc creates path and fills it with write.
func writeFileFunc(path string, write func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := write(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// DeepPrestigeAge is where the deep tier's runs prestige.
const DeepPrestigeAge = "quantum_age"

// progressionConfig is the progression scenario's run for the tier.
func progressionConfig(e *Env, o Overrides) (Config, []int64) {
	cfg := e.Base
	var seeds []int64
	switch {
	case e.Tier == TierDeep:
		// One cycle to a Quantum Age prestige, the final epoch's passage:
		// every age from the Primitive to the Galactic is graded, and the
		// Invite makes the Last Passage come (endured) on every seed. The
		// targets to it sum to about 557 hours on the one-week curve.
		seeds = e.seeds(3)
		cfg.Cycles, cfg.PrestigeAge, cfg.FinalAge, cfg.MaxSim = 1, DeepPrestigeAge, "", 1000*time.Hour
		cfg.InviteCosmic = true
	case e.full():
		// Two cycles: the first to a Modern Age prestige, graded as a first
		// run; the second a push (Era Mastery's later-run check): it plays
		// for as long as the first took, then prestiges from wherever it got.
		// About 260 hours for the bot; 1,000 leaves room for a slow seed.
		seeds = e.seeds(8)
		cfg.Cycles, cfg.FinalAge, cfg.MaxSim, cfg.PushCycles = 2, "", 1000*time.Hour, true
	default:
		seeds = e.seeds(3)
		cfg.Cycles, cfg.StopAge, cfg.MaxSim = 1, "bronze_age", 300*time.Hour
	}
	if o.MaxSim > 0 {
		cfg.MaxSim = o.MaxSim
	}
	if o.StopAge != "" {
		cfg.StopAge = o.StopAge
	}
	if o.PrestigeAge != "" {
		cfg.PrestigeAge = o.PrestigeAge
	}
	if o.Cycles > 0 {
		cfg.Cycles = o.Cycles
	}
	if o.FinalAgeSet {
		cfg.FinalAge = o.FinalAge
	}
	return cfg, seeds
}

func runProgression(e *Env, res *Result) {
	cfg, seeds := progressionConfig(e, e.Overrides)
	describeProgression(res, runBotSet(e, res, "progression", "progression", cfg, seeds))
}

// describeProgression writes the progression scenario's summary line and
// sections from its graded runs.
func describeProgression(res *Result, sum *Summary) {
	reached := map[string]int{}
	for _, r := range sum.Runs {
		reached[r.FinalAge]++
	}
	var parts []string
	for _, a := range config.AgeOrder() {
		if n := reached[a]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d at %s", n, a))
		}
	}
	slow := 0
	for _, p := range sum.Pacing {
		if p.Verdict == VerdictSlow || p.Verdict == VerdictFast {
			slow++
		}
	}
	res.Summary = fmt.Sprintf("%d seed(s) ended %s; %d anomaly(ies); %d age(s) off the pacing band", len(sum.Runs),
		strings.Join(parts, ", "), sum.Anomalies, slow)
	if f := sum.FirstRun; f != nil && f.MedianSecs >= 0 {
		res.Summary += fmt.Sprintf("; first run to the Modern Age %s (%s)", days(f.MedianSecs), verdictMark(f.Verdict))
	}
	var sb strings.Builder
	sum.writePacingTable(&sb)
	sum.writeFirstRun(&sb)
	sum.writeLaterRun(&sb)
	sum.writeDepth(&sb)
	res.section("Pacing per age", "%s", sb.String())
	res.section("Full bot report", "See progression.md next to this report for runs, events, harbingers and state dumps.")
}

func runStatic(e *Env, res *Result) {
	problems, slack := StaticGates()
	s := &Summary{Gates: problems, Slack: slack}
	var sb strings.Builder
	s.writeGates(&sb)
	for _, g := range problems {
		res.fail("gate_"+g.Kind, "%s -> %s: %s %s (need %s, max storage %s)", g.From, g.To, g.Key, g.Resource, num(g.Need), num(g.MaxStorage))
	}
	rows := StaticStorage()
	short := 0
	for _, r := range rows {
		if !r.OK() {
			short++
			res.fail("storage_covenant", "%s: max %s storage %s holds %.2f h of typical income %s/tick (want %g h)", r.Age, r.Resource, num(r.MaxStorage), r.Hours, num(r.Income), r.Want())
		}
	}
	mp, mr := StaticMilestones()
	for _, p := range mp {
		res.fail("milestone_"+p.Kind, "%s", p.Why)
	}
	hp := StaticHarbingerPrices()
	for _, p := range hp {
		res.fail("harbinger_price", "%s %s costs %s %s, over the %s storage buildable in %s", p.Epoch, p.Answer, num(p.Price), p.Resource, num(p.MaxStorage), p.Age)
	}
	ar := StaticAppeaseRules()
	for _, p := range ar {
		res.fail("appease_"+p.Rule, "%s", p)
	}
	br := StaticBraceRules()
	for _, p := range br {
		res.fail("brace_"+p.Rule, "%s", p)
	}
	fp, fr := StaticFaithStrength()
	for _, p := range fp {
		res.fail("faith_"+p.Rule, "%s", p)
	}
	dp := StaticDepth()
	for _, p := range dp {
		res.fail("depth_points", "%s", p)
	}
	rr := StaticResearch()
	research := 0
	for _, r := range rr {
		for _, p := range r.Problems() {
			research++
			res.fail("research_covenant", "%s", p)
		}
	}
	bp, breach := StaticBadges()
	for _, p := range bp {
		res.fail("badge_"+p.Kind, "%s", p.Why)
	}
	res.Summary = fmt.Sprintf("%d gate problem(s) across %d advances; %d age(s) short of the Storage Covenant; %d milestone problem(s); %d harbinger price(s) over storage; %d Appease price(s) off the warning; %d Brace price(s) off their warning; %d faith band(s) out of reach; %d depth point problem(s); %d research problem(s); %d badge problem(s)", len(problems), len(slack), short, len(mp), len(hp), len(ar), len(br), len(fp), len(dp), research, len(bp))
	res.section("Static gate check", "%s", strings.TrimPrefix(sb.String(), "\n## Static gate check\n\n"))
	var st strings.Builder
	writeStorage(&st, rows)
	res.section("Storage Covenant", "%s", st.String())
	var mf strings.Builder
	writeMilestones(&mf, mp, mr)
	res.section("Milestone feasibility", "%s", mf.String())
	var hf strings.Builder
	writeHarbingerPrices(&hf, hp, ar, br)
	res.section("Harbinger prices against storage and the warning", "%s", hf.String())
	var ff strings.Builder
	writeFaithStrength(&ff, fp, fr)
	res.section("Faith strength", "%s", ff.String())
	var df strings.Builder
	writeDepthStatic(&df, dp)
	res.section("Depth points", "%s", df.String())
	var rf strings.Builder
	writeResearch(&rf, rr)
	res.section("Research Covenant", "%s", rf.String())
	var bf strings.Builder
	writeBadges(&bf, bp, breach)
	res.section("Badge Covenant", "%s", strings.TrimPrefix(bf.String(), "\n### Badges\n\n"))
}
