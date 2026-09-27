// Command smoke boots AgeForge headless, plays it end to end with a greedy
// bot over several seeds, and reports panics, soft-locks, invariant
// violations, pacing and event counts.
//
// It exits 1 if any seed hit a panic, a soft-lock or an invariant violation.
// Slow pacing is reported but never fails the run.
//
// Usage:
//
//	go run ./cmd/smoke                 # quick mode (make smoke)
//	go run ./cmd/smoke -mode full      # nightly mode (make smoke-full)
//	go run ./cmd/smoke -seeds 1 -v     # one seed, progress on stderr
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/smoke"
)

func main() {
	os.Exit(run())
}

func run() int {
	mode := flag.String("mode", "quick", "preset: quick (5 seeds, one run to prestige) or full (10 seeds, 2 prestige cycles, then on to digital_age)")
	seeds := flag.Int("seeds", 0, "number of seeds (0 = mode default)")
	seedBase := flag.Int64("seed-base", 1, "first seed; seeds are seed-base, seed-base+1, ...")
	catastrophe := flag.String("catastrophe", "endure", "how the bot answers a catastrophe: endure or succumb")
	prestigeAge := flag.String("prestige-age", "", "age at which to prestige (default: first age where prestige is allowed)")
	cycles := flag.Int("cycles", 0, "prestige cycles to play (0 = mode default)")
	finalAge := flag.String("final-age", "-", "after the last prestige keep playing to this age (\"\" = stop at prestige; default: mode preset)")
	harbinger := flag.String("harbinger", "ignore", "bot answer to harbingers: ignore, appease, brace or both (level 1 only)")
	stopAge := flag.String("stop-age", "", "end each run as soon as this age is entered (before prestige)")
	softlock := flag.Duration("softlock", 0, "simulated 1x span without progress that counts as a soft-lock (0 = default 30m)")
	ageTimeout := flag.Duration("age-timeout", 0, "simulated 1x time in one age that counts as stalled (0 = default)")
	maxSim := flag.Duration("max-sim", 0, "simulated 1x cap per seed (0 = default)")
	checkEvery := flag.Int("check-every", 0, "ticks between invariant sweeps (0 = default 25)")
	parallel := flag.Int("parallel", runtime.NumCPU(), "seeds to run at once")
	outDir := flag.String("out", "smoke-report", "directory for report.md and report.json")
	trace := flag.Bool("trace", false, "write every bot action to <out>/trace-<seed>.log")
	verbose := flag.Bool("v", false, "print each seed's result to stderr as it finishes")
	flag.Parse()

	cfg := smoke.DefaultConfig()
	nSeeds, defCycles, defFinal := 5, 1, ""
	switch *mode {
	case "quick":
	case "full":
		nSeeds, defCycles, defFinal = 10, 2, "digital_age"
	default:
		fmt.Fprintf(os.Stderr, "unknown -mode %q (want quick or full)\n", *mode)
		return 2
	}
	if *seeds > 0 {
		nSeeds = *seeds
	}
	if *cycles > 0 {
		defCycles = *cycles
	}
	if *finalAge != "-" {
		defFinal = *finalAge
	}
	cfg.Seeds = nil
	for i := 0; i < nSeeds; i++ {
		cfg.Seeds = append(cfg.Seeds, *seedBase+int64(i))
	}
	cfg.Cycles = defCycles
	cfg.FinalAge = defFinal
	cfg.PrestigeAge = *prestigeAge
	cfg.StopAge = *stopAge
	cfg.Harbinger = *harbinger
	switch cfg.Harbinger {
	case smoke.HarbingerIgnore, smoke.HarbingerAppease, smoke.HarbingerBrace, smoke.HarbingerBoth:
	default:
		fmt.Fprintf(os.Stderr, "unknown -harbinger %q (want ignore, appease, brace or both)\n", cfg.Harbinger)
		return 2
	}
	cfg.Catastrophe = *catastrophe
	if *softlock > 0 {
		cfg.SoftlockSpan = *softlock
	}
	if *ageTimeout > 0 {
		cfg.AgeTimeout = *ageTimeout
	}
	if *maxSim > 0 {
		cfg.MaxSim = *maxSim
	}
	if *checkEvery > 0 {
		cfg.CheckEvery = *checkEvery
	}
	if cfg.Catastrophe != "endure" && cfg.Catastrophe != "succumb" {
		fmt.Fprintf(os.Stderr, "unknown -catastrophe %q (want endure or succumb)\n", cfg.Catastrophe)
		return 2
	}
	ages := config.AgeByKey()
	for _, a := range []string{cfg.PrestigeAge, cfg.FinalAge, cfg.StopAge} {
		if _, ok := ages[a]; a != "" && !ok {
			fmt.Fprintf(os.Stderr, "unknown age %q\n", a)
			return 2
		}
	}

	// The engine stats a save file on every snapshot; keep that away from
	// the real data directory.
	tmp, err := os.MkdirTemp("", "ageforge-smoke-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	defer os.RemoveAll(tmp)
	defer game.SetDataDirForTest(tmp)()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if *trace {
		cfg.TraceDir = *outDir
	}
	started := time.Now()
	results := make([]*smoke.RunResult, len(cfg.Seeds))
	sem := make(chan struct{}, max(1, *parallel))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i, seed := range cfg.Seeds {
		wg.Add(1)
		go func(i int, seed int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := smoke.Run(cfg, seed)
			results[i] = r
			if *verbose {
				mu.Lock()
				fmt.Fprintf(os.Stderr, "seed %d: %s at %s after %d ticks (%s at 1x), %d anomalies, %s wall\n",
					seed, r.Outcome, r.FinalAge, r.Ticks, time.Duration(r.Seconds*float64(time.Second)).Round(time.Minute),
					len(r.Anomalies), time.Duration(r.WallMillis)*time.Millisecond)
				mu.Unlock()
			}
		}(i, seed)
	}
	wg.Wait()

	sum := smoke.NewSummary(*mode, cfg, started, results)
	mdPath := filepath.Join(*outDir, "report.md")
	jsonPath := filepath.Join(*outDir, "report.json")
	if err := writeFile(mdPath, sum.WriteMarkdown); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := writeFile(jsonPath, sum.WriteJSON); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	verdict := "PASS"
	if sum.Failed {
		verdict = "FAIL"
	}
	fmt.Printf("smoke %s: %d seed(s), %d anomaly(ies), %s wall. Report: %s, %s\n",
		verdict, len(results), sum.Anomalies, time.Since(started).Round(time.Second), mdPath, jsonPath)
	if sum.Failed {
		return 1
	}
	return 0
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
