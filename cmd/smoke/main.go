// Command smoke runs AgeForge's smoke suite: named scenarios that play the
// game headless (a greedy bot, save/load round trips, offline catch-up,
// command fuzzing, accounts, performance, the UI on a simulated terminal,
// docs against config) and report what broke.
//
// It exits 1 if any scenario failed: a panic, a soft-lock, an invariant
// violation, a save/load divergence, a fuzz crash, an account failure, a
// docs mismatch or a blown performance budget. Pacing fails nothing unless
// -pacing enforce.
//
// Usage:
//
//	go run ./cmd/smoke                         # fast tier (make smoke)
//	go run ./cmd/smoke -tier full              # everything, deeper (make smoke-full)
//	go run ./cmd/smoke -scenario saveload -v   # one scenario
//	go run ./cmd/smoke -list                   # what the scenarios are
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/smoke"
)

func main() {
	os.Exit(run())
}

func run() int {
	tier := flag.String("tier", smoke.TierFast, "fast (per PR, a few minutes), full (nightly: every scenario, deeper, more seeds) or deep (weekly: progression to a Quantum Age prestige)")
	merge := flag.String("merge", "", "comma-separated report.json files (or directories holding one) to merge instead of running anything: progression runs are pooled and graded again under -pacing, static runs again, other scenarios carry over (the weekly job's per-seed jobs, the nightly's shards)")
	mode := flag.String("mode", "", "deprecated alias for -tier (quick = fast, full = full)")
	scenarios := flag.String("scenario", "all", "comma-separated scenarios to run, or all (the tier's set); see -list")
	only := flag.String("only", "", "comma-separated scenarios: of those -scenario picks, run only these (a CI shard's share; nothing left writes an empty report)")
	skip := flag.String("skip", "", "comma-separated scenarios: of those -scenario picks, leave these out (a CI shard's share)")
	list := flag.Bool("list", false, "list the scenarios and exit")
	pacing := flag.String("pacing", smoke.PacingReport, "report (grade ages against smoke/targets.go, never fail) or enforce (fail when a first-cycle age's median across seeds leaves the band, or any age passes its timeout; progression only)")
	seeds := flag.Int("seeds", 0, "seeds per bot scenario (0 = the scenario's tier default)")
	seedBase := flag.Int64("seed-base", 1, "first seed; seeds are seed-base, seed-base+1, ...")
	catastrophe := flag.String("catastrophe", "endure", "how the bot answers a catastrophe: endure or succumb")
	harbinger := flag.String("harbinger", smoke.HarbingerIgnore, "bot answer to harbingers: ignore, appease, brace or both (Appease both levels, Brace level 1)")
	style := flag.String("style", "", "styles scenario: run only this style ("+strings.Join(smoke.StyleNames(), ", ")+")")
	checkIn := flag.Duration("check-in", 0, "idle style: simulated 1x time between check-ins (0 = 3h)")
	noPlan := flag.Bool("no-plan", false, "idle style: leave no build plan at check-ins (to measure what the plan is worth)")
	noOverflow := flag.Bool("no-overflow", false, "turn wonder overflow off for the bot runs (to measure what it is worth)")
	noShares := flag.Bool("no-shares", false, "idle style: recruit and assign by hand at check-ins, with auto-recruit off, instead of leaving workers to worker shares (to measure what shares are worth)")
	army := flag.String("army", "off", "bot policy for the army: off (ignore it beyond what age gates ask for, the default the pacing targets assume) or on (keep a modest garrison of the age's newest military buildings, bought from spare stock)")
	deals := flag.String("deals", "off", "bot policy for faction trade deals: off (ignore them, the default the pacing targets assume) or on (take deals for what the age needs, paid from surplus)")
	prestigeAge := flag.String("prestige-age", "", "progression: age at which to prestige (default: first age where prestige is allowed)")
	cycles := flag.Int("cycles", 0, "progression: prestige cycles to play (0 = tier default)")
	finalAge := flag.String("final-age", "-", "progression: after the last prestige keep playing to this age (\"\" = stop at prestige; default: tier preset)")
	stopAge := flag.String("stop-age", "", "progression: end each run as soon as this age is entered")
	softlock := flag.Duration("softlock", 0, "simulated 1x span without progress that counts as a soft-lock (0 = default 30m)")
	ageTimeout := flag.Duration("age-timeout", 0, "fixed simulated 1x time allowed in every age (0 = derive each age's from the pacing table)")
	maxSim := flag.Duration("max-sim", 0, "progression and styles: simulated 1x cap per seed (0 = the scenario default)")
	checkEvery := flag.Int("check-every", 0, "ticks between invariant sweeps (0 = default 25)")
	digestEvery := flag.Int("digest-every", 0, "record the engine state digest every N ticks in report.json (runs[].digests), to find where two machines' runs of a seed part (0 = final digest only)")
	strict := flag.Bool("strict", false, "fail on known bugs too (they are reported as warnings otherwise)")
	fuzzCommands := flag.Int("fuzz-commands", 0, "fuzz: commands per seed (0 = tier default)")
	parallel := flag.Int("parallel", runtime.NumCPU(), "seeds to run at once")
	outDir := flag.String("out", "smoke-report", "directory for report.md, summary.md, report.json and per-scenario files")
	trace := flag.Bool("trace", false, "write every bot action to <out>/trace-<seed>.log")
	verbose := flag.Bool("v", false, "print progress to stderr")
	flag.Parse()

	if *list {
		for _, s := range smoke.Scenarios() {
			tag := ""
			if s.FullOnly {
				tag = " (full tier)"
			}
			fmt.Printf("%-12s %s%s\n", s.Name, s.Desc, tag)
		}
		return 0
	}
	switch *mode {
	case "":
	case "quick":
		*tier = smoke.TierFast
	case "full":
		*tier = smoke.TierFull
	default:
		fmt.Fprintf(os.Stderr, "unknown -mode %q (want quick or full; prefer -tier)\n", *mode)
		return 2
	}
	if *tier != smoke.TierFast && *tier != smoke.TierFull && *tier != smoke.TierDeep {
		fmt.Fprintf(os.Stderr, "unknown -tier %q (want fast, full or deep)\n", *tier)
		return 2
	}
	if *pacing != smoke.PacingReport && *pacing != smoke.PacingEnforce {
		fmt.Fprintf(os.Stderr, "unknown -pacing %q (want report or enforce)\n", *pacing)
		return 2
	}

	base := smoke.DefaultConfig()
	base.Catastrophe = *catastrophe
	base.Harbinger = *harbinger
	base.Pacing = *pacing
	base.NoPlan, base.NoOverflow, base.NoShares = *noPlan, *noOverflow, *noShares
	switch *deals {
	case "on", "off":
		base.Deals = *deals == "on"
	default:
		fmt.Fprintf(os.Stderr, "unknown -deals %q (want on or off)\n", *deals)
		return 2
	}
	switch *army {
	case "on", "off":
		base.Army = *army == "on"
	default:
		fmt.Fprintf(os.Stderr, "unknown -army %q (want on or off)\n", *army)
		return 2
	}
	switch base.Harbinger {
	case smoke.HarbingerIgnore, smoke.HarbingerAppease, smoke.HarbingerBrace, smoke.HarbingerBoth:
	default:
		fmt.Fprintf(os.Stderr, "unknown -harbinger %q (want ignore, appease, brace or both)\n", base.Harbinger)
		return 2
	}
	if base.Catastrophe != "endure" && base.Catastrophe != "succumb" {
		fmt.Fprintf(os.Stderr, "unknown -catastrophe %q (want endure or succumb)\n", base.Catastrophe)
		return 2
	}
	if *softlock > 0 {
		base.SoftlockSpan = *softlock
	}
	if *ageTimeout > 0 {
		base.AgeTimeout = *ageTimeout
	}
	if *checkEvery > 0 {
		base.CheckEvery = *checkEvery
	}
	if *digestEvery > 0 {
		base.DigestEvery = *digestEvery
	}
	if *checkIn < 0 {
		fmt.Fprintf(os.Stderr, "-check-in must not be negative (got %s)\n", *checkIn)
		return 2
	}
	if *style != "" {
		if _, err := smoke.ApplyStyle(base, *style); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
	}
	ages := config.AgeByKey()
	for _, a := range []string{*prestigeAge, *stopAge, strings.TrimPrefix(*finalAge, "-")} {
		if _, ok := ages[a]; a != "" && !ok {
			fmt.Fprintf(os.Stderr, "unknown age %q\n", a)
			return 2
		}
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	absOut, _ := filepath.Abs(*outDir)
	if *trace {
		base.TraceDir = absOut
	}

	env := &smoke.Env{
		Tier: *tier, Seeds: *seeds, SeedBase: *seedBase, Pacing: *pacing, Parallel: *parallel,
		OutDir: absOut, RepoRoot: repoRoot(), Base: base, Style: *style, CheckIn: *checkIn, FuzzCommands: *fuzzCommands, Strict: *strict,
		Overrides: smoke.Overrides{MaxSim: *maxSim, StopAge: *stopAge, PrestigeAge: *prestigeAge, Cycles: *cycles},
	}
	if *finalAge != "-" {
		env.Overrides.FinalAge, env.Overrides.FinalAgeSet = *finalAge, true
	}
	if *verbose {
		env.Logf = func(format string, args ...interface{}) {
			fmt.Fprintf(os.Stderr, "[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
		}
	}

	names := splitList(*scenarios)
	env.Only, env.Skip = splitList(*only), splitList(*skip)
	var sess *smoke.Session
	var err error
	if *merge != "" {
		sess, err = mergeReports(env, *merge)
	} else {
		sess, err = smoke.RunScenarios(env, names)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	for _, f := range []struct {
		name  string
		write func(io.Writer) error
	}{{"report.md", sess.WriteMarkdown}, {"summary.md", sess.WriteSummary}, {"report.json", sess.WriteJSON}} {
		if err := writeFile(filepath.Join(absOut, f.name), f.write); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
	}
	verdict := "PASS"
	if sess.Failed {
		verdict = "FAIL"
	}
	var parts []string
	for _, r := range sess.Scenarios {
		parts = append(parts, fmt.Sprintf("%s %s", r.Name, r.Status))
	}
	if len(parts) == 0 {
		parts = append(parts, "no scenario selected")
	}
	fmt.Printf("smoke %s (%s tier, %s): %s. Report: %s\n", verdict, sess.Tier,
		(time.Duration(sess.WallMs) * time.Millisecond).Round(time.Second), strings.Join(parts, ", "), filepath.Join(*outDir, "report.md"))
	if sess.Failed {
		return 1
	}
	return 0
}

// mergeReports reads the listed report.json files (a directory stands for
// the report.json inside it) and merges them (smoke.MergeSessions).
func mergeReports(env *smoke.Env, list string) (*smoke.Session, error) {
	var parts []*smoke.Session
	for _, p := range strings.Split(list, ",") {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			p = filepath.Join(p, "report.json")
		}
		s, err := smoke.ReadSession(p)
		if err != nil {
			return nil, err
		}
		parts = append(parts, s)
	}
	if env.Logf != nil {
		env.Logf("merging %d report(s)", len(parts))
	}
	return smoke.MergeSessions(env, parts)
}

// splitList splits a comma-separated flag value, dropping blanks.
func splitList(v string) []string {
	var out []string
	for _, n := range strings.Split(v, ",") {
		if n = strings.TrimSpace(n); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// repoRoot walks up from the working directory to the directory holding
// go.mod ("" if there is none).
func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func writeFile(path string, write func(io.Writer) error) error {
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
