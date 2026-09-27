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
	sum := NewSummary(name, cfg, started, runs)
	for _, r := range runs {
		for _, a := range r.Anomalies {
			f := res.fail(a.Kind+"/"+a.Check, "%s: %s (cycle %d, %s, tick %d, seen %dx)", name, a.Message, a.Cycle, a.Age, a.Tick, a.Count)
			f.Seed = a.Seed
			f.Repro = reproCmd(scenario, cfg, a.Seed)
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
func reproCmd(scenario string, cfg Config, seed int64) string {
	parts := []string{"go run ./cmd/smoke", "-scenario " + scenario, fmt.Sprintf("-seed-base %d -seeds 1", seed)}
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

// progressionConfig is the progression scenario's run for the tier.
func progressionConfig(e *Env, o Overrides) (Config, []int64) {
	cfg := e.Base
	var seeds []int64
	if e.full() {
		seeds = e.seeds(8)
		cfg.Cycles, cfg.FinalAge, cfg.MaxSim = 2, "digital_age", 1200*time.Hour
	} else {
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
	sum := runBotSet(e, res, "progression", "progression", cfg, seeds)
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
	var sb strings.Builder
	sum.writePacingTable(&sb)
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
	res.Summary = fmt.Sprintf("%d gate problem(s) across %d advances", len(problems), len(slack))
	res.section("Static gate check", "%s", strings.TrimPrefix(sb.String(), "\n## Static gate check\n\n"))
}
