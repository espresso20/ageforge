package smoke

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/espresso20/ageforge/game"
)

// Tiers.
const (
	// TierFast is the per-PR tier (`make smoke`): a few minutes.
	TierFast = "fast"
	// TierFull is the nightly tier (`make smoke-full`): everything, deeper.
	TierFull = "full"
	// TierDeep is the weekly tier: only the progression scenario, played
	// to a Quantum Age prestige through the Last Passage, so pacing is
	// graded on every age the nightly never reaches. CI runs one seed per
	// job and merges the reports (-merge).
	TierDeep = "deep"
)

// deepScenarios are the scenarios the deep tier runs by default.
var deepScenarios = map[string]bool{"static": true, "progression": true}

// Scenario statuses.
const (
	StatusPass = "pass"
	StatusFail = "fail"
	StatusSkip = "skip"
)

// Env is what every scenario runs with: the tier, seeds, pacing mode and
// where to write files. cmd/smoke fills it from its flags.
type Env struct {
	Tier string
	// Seeds overrides the scenario's tier default seed count when > 0.
	Seeds    int
	SeedBase int64
	Pacing   string
	Parallel int
	// OutDir receives per-scenario files next to report.md.
	OutDir string
	// RepoRoot is the module root (docsync reads the docs there; ui and perf
	// run `go test` there).
	RepoRoot string
	// Base is the run configuration the bot scenarios start from (flags such
	// as -catastrophe and -harbinger land here).
	Base Config
	// Overrides replace the progression scenario's tier defaults.
	Overrides Overrides
	// Style limits the styles scenario to one style ("" runs them all).
	Style string
	// CheckIn is the idle style's time between check-ins (0: IdleCheckIn).
	CheckIn time.Duration
	// Strict fails on known bugs too instead of reporting them as warnings.
	// There are none today; a scenario that learns of one checks it.
	Strict bool
	// FuzzCommands overrides the fuzz scenario's command count when > 0.
	FuzzCommands int
	// Only and Skip narrow the scenarios RunScenarios picks (after the tier
	// and the names asked for): with Only set, a scenario outside it is
	// dropped, and one in Skip always is. The nightly's CI shards use them
	// to split the full tier between jobs; a shard left with nothing writes
	// an empty report.
	Only, Skip []string
	// Logf prints progress; nil is silent.
	Logf func(format string, args ...interface{})
}

func (e *Env) logf(format string, args ...interface{}) {
	if e.Logf != nil {
		e.Logf(format, args...)
	}
}

func (e *Env) full() bool { return e.Tier == TierFull }

// seeds returns n seeds (or the -seeds override) from SeedBase.
func (e *Env) seeds(n int) []int64 {
	if e.Seeds > 0 {
		n = e.Seeds
	}
	out := make([]int64, n)
	for i := range out {
		out[i] = e.SeedBase + int64(i)
	}
	return out
}

func (e *Env) parallel() int {
	if e.Parallel > 0 {
		return e.Parallel
	}
	return runtime.NumCPU()
}

// runSeeds runs f for every seed, Parallel at a time, and returns the
// results in seed order.
func runSeeds[T any](e *Env, seeds []int64, f func(seed int64) T) []T {
	out := make([]T, len(seeds))
	sem := make(chan struct{}, e.parallel())
	var wg sync.WaitGroup
	for i, s := range seeds {
		wg.Add(1)
		go func(i int, s int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i] = f(s)
		}(i, s)
	}
	wg.Wait()
	return out
}

// Finding is one problem a scenario found (a failure) or one thing worth
// knowing that does not fail the run (a warning).
type Finding struct {
	Check   string `json:"check"`
	Message string `json:"message"`
	Seed    int64  `json:"seed,omitempty"`
	// Repro is how to reproduce it: a command line, a seed and command list.
	Repro string `json:"repro,omitempty"`
	// Detail is long text for the full report only (a state dump, a stack).
	Detail string `json:"detail,omitempty"`
}

// Section is one titled block of Markdown in the full report.
type Section struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Result is one scenario's outcome.
type Result struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	WallMs   int64     `json:"wall_ms"`
	Summary  string    `json:"summary"`
	Failures []Finding `json:"failures"`
	Warnings []Finding `json:"warnings,omitempty"`
	Sections []Section `json:"-"`
	// Progression holds bot-run summaries (progression, styles, prestige).
	Progression []*Summary  `json:"progression,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

func (r *Result) fail(check, format string, args ...interface{}) *Finding {
	r.Failures = append(r.Failures, Finding{Check: check, Message: fmt.Sprintf(format, args...)})
	return &r.Failures[len(r.Failures)-1]
}

func (r *Result) warn(check, format string, args ...interface{}) *Finding {
	r.Warnings = append(r.Warnings, Finding{Check: check, Message: fmt.Sprintf(format, args...)})
	return &r.Warnings[len(r.Warnings)-1]
}

func (r *Result) section(title, format string, args ...interface{}) {
	r.Sections = append(r.Sections, Section{Title: title, Body: fmt.Sprintf(format, args...)})
}

// Scenario is one named, individually runnable check.
type Scenario struct {
	Name string
	// Desc is one line for -list and the report.
	Desc string
	// FullOnly scenarios are skipped in the fast tier.
	FullOnly bool
	// Paced scenarios run with the session's pacing mode; the rest always
	// run in report mode. The pacing targets describe a greedy first run, so
	// only the progression scenario is held to them: the idle style checks
	// in every few hours, and the saveload and perf runs are cut short.
	Paced bool
	Run   func(e *Env, res *Result)
}

// Scenarios is the suite, in run order.
func Scenarios() []Scenario {
	return []Scenario{
		{Name: "static", Desc: "Gate Covenant check from config alone: every age gate fits storage and has a source, and every milestone can be completed", Run: runStatic},
		{Name: "docsync", Desc: "site and README headline numbers, lineage count and command reference match config and the command table", Run: runDocsync},
		{Name: "progression", Desc: "greedy bot plays seeds end to end: panics, soft-locks, invariants, and the pacing table", Paced: true, Run: runProgression},
		{Name: "veteran", Desc: "Era Mastery's veteran preset (mastery 10 through the Space Age): each age against its target ÷ k, the Primitive and Stone Ages under an hour, the Modern Age in 1.1 to 1.5 days; then the legacy kit bought after a scripted prestige, and the shop refund", Paced: true, Run: runVeteran},
		{Name: "saveload", Desc: "save at a checkpoint per age, load into a fresh engine, continue, and compare against the uninterrupted run", Run: runSaveload},
		{Name: "offline", Desc: "close the game for 1h, 8h and 30h through the offline-gains path: positive, sane, capped at 24h", Run: runOffline},
		{Name: "fuzz", Desc: "random, malformed and hostile commands through the real command handler: no panics, invariants hold, ticks go on", Run: runFuzz},
		{Name: "accounts", Desc: "create, switch, export, import, back up, recover and wipe accounts in a temp data dir", Run: runAccounts},
		{Name: "perf", Desc: "late-game tick and GetState latency against budgets, and memory growth over a long run", Run: runPerf},
		{Name: "ui", Desc: "UI sweep under the themes, plus the dashboard and every overlay at 80x24 and 100x30", Run: runUI},
		{Name: "styles", Desc: "the bot under different play styles: greedy, idle check-ins, harbinger buyer, succumber, cosmic legacy hunter, garrison keeper (report-only), early taste (prestige at the Medieval Age)", FullOnly: true, Run: runStyles},
		{Name: "idle", Desc: "check-in players at 1h, 3h and 8h leaving a build plan each visit: median time to the first prestige against the idle targets; the veteran with the legacy kit at 3h and 8h against the active veteran", FullOnly: true, Run: runIdle},
		{Name: "prestige", Desc: "two prestige cycles: the depth points formula, the legacy kit and the shop refund, and legacies, ruins and the Cosmic Legacy persisting", FullOnly: true, Run: runPrestige},
	}
}

// ScenarioNames lists the scenario names in run order.
func ScenarioNames() []string {
	var out []string
	for _, s := range Scenarios() {
		out = append(out, s.Name)
	}
	return out
}

// Session is one invocation of the suite.
type Session struct {
	Tier      string    `json:"tier"`
	Pacing    string    `json:"pacing"`
	Started   time.Time `json:"started"`
	WallMs    int64     `json:"wall_ms"`
	Scenarios []*Result `json:"scenarios"`
	Failed    bool      `json:"failed"`
}

// RunScenarios runs the named scenarios ("all" or empty runs the tier's
// set) one after another. Each gets its own temporary data directory, so no
// scenario sees another's saves or accounts, and a panic inside a scenario's
// own code fails that scenario instead of the session.
func RunScenarios(e *Env, names []string) (*Session, error) {
	want := map[string]bool{}
	all := len(names) == 0
	for _, n := range names {
		if n == "all" {
			all = true
			continue
		}
		want[n] = true
	}
	known := map[string]bool{}
	for _, s := range Scenarios() {
		known[s.Name] = true
	}
	only, skip := map[string]bool{}, map[string]bool{}
	for _, n := range e.Only {
		only[n] = true
	}
	for _, n := range e.Skip {
		skip[n] = true
	}
	for _, set := range []map[string]bool{want, only, skip} {
		for n := range set {
			if !known[n] {
				return nil, fmt.Errorf("unknown scenario %q (have: %s)", n, strings.Join(ScenarioNames(), ", "))
			}
		}
	}
	sess := &Session{Tier: e.Tier, Pacing: e.Pacing, Started: time.Now()}
	for _, sc := range Scenarios() {
		if !all && !want[sc.Name] {
			continue
		}
		if (len(only) > 0 && !only[sc.Name]) || skip[sc.Name] {
			continue
		}
		res := &Result{Name: sc.Name}
		if e.Tier == TierDeep && !deepScenarios[sc.Name] && !want[sc.Name] {
			res.Status = StatusSkip
			res.Summary = "not in the deep tier"
			sess.Scenarios = append(sess.Scenarios, res)
			continue
		}
		if sc.FullOnly && !e.full() && !want[sc.Name] {
			res.Status = StatusSkip
			res.Summary = "full tier only"
			sess.Scenarios = append(sess.Scenarios, res)
			continue
		}
		e.logf("scenario %s: running", sc.Name)
		start := time.Now()
		se := *e
		if !sc.Paced {
			se.Base.Pacing = PacingReport
		}
		runIsolated(&se, sc, res)
		res.WallMs = time.Since(start).Milliseconds()
		if res.Status == "" {
			res.Status = StatusPass
			if len(res.Failures) > 0 {
				res.Status = StatusFail
			}
		}
		if res.Status == StatusFail {
			sess.Failed = true
		}
		e.logf("scenario %s: %s in %s (%d failure(s), %d warning(s))", sc.Name, strings.ToUpper(res.Status),
			time.Since(start).Round(time.Second), len(res.Failures), len(res.Warnings))
		sess.Scenarios = append(sess.Scenarios, res)
	}
	sess.WallMs = time.Since(sess.Started).Milliseconds()
	return sess, nil
}

func runIsolated(e *Env, sc Scenario, res *Result) {
	dir, err := os.MkdirTemp("", "ageforge-smoke-"+sc.Name+"-")
	if err != nil {
		res.fail("harness", "temp data dir: %v", err)
		return
	}
	defer os.RemoveAll(dir)
	restore := game.SetDataDirForTest(filepath.Join(dir, "data"))
	defer restore()
	defer func() {
		if rec := recover(); rec != nil {
			res.fail("harness_panic", "the %s scenario panicked outside the game: %v", sc.Name, rec).Detail = string(debug.Stack())
		}
	}()
	sc.Run(e, res)
}
