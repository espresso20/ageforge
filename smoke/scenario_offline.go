package smoke

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// playUntil plays seed with the bot until stop returns true (checked at
// every decision) and hands back the engine in that state.
func playUntil(cfg Config, seed int64, stop func(r *runner) bool) (*game.GameEngine, *RunResult) {
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	cfg.hook = stop
	r := newRunner(cfg, seed, ge)
	r.play()
	return ge, r.res
}

// inAgeFor stops a run once it has spent ticks in age.
func inAgeFor(age string, ticks int) func(r *runner) bool {
	return func(r *runner) bool { return r.age == age && r.ticks-r.ageT0 >= ticks }
}

// offlineDurations are the closures simulated; 24h is the cap itself, so
// 30h must pay exactly what 24h does.
var offlineDurations = []time.Duration{time.Hour, 8 * time.Hour, game.MaxOfflineTime, 30 * time.Hour}

func runOffline(e *Env, res *Result) {
	bases := []string{"stone_age"}
	if e.full() {
		bases = []string{"stone_age", "bronze_age", "iron_age"}
	}
	seed := e.SeedBase
	var rows []string
	for _, age := range bases {
		cfg := e.Base
		cfg.Cycles, cfg.MaxSim = 1, 1000*time.Hour
		base, run := playUntil(cfg, seed, inAgeFor(age, 1000))
		if run.Outcome != OutcomeDone || base.GetState().Age != age {
			res.warn("base", "could not reach %s on seed %d (%s at %s); skipped", age, seed, run.Outcome, run.FinalAge)
			continue
		}
		rows = append(rows, offlineBase(res, base, age, seed)...)
	}
	res.Summary = fmt.Sprintf("%d closure(s) simulated from %d base state(s)", len(rows), len(bases))
	res.section("Closures", "| base | away | offline ticks | resources gained | largest gain | notes |\n|---|---|---|---|---|---|\n%s", strings.Join(rows, "\n"))
}

type offlineRun struct {
	d     time.Duration
	pre   game.GameState
	post  game.GameState
	gains map[string]float64
}

func offlineBase(res *Result, base *game.GameEngine, age string, seed int64) []string {
	repro := fmt.Sprintf("go run ./cmd/smoke -scenario offline -seed-base %d -v", seed)
	fail := func(check, format string, args ...interface{}) {
		f := res.fail(check, "%s base: "+format, append([]interface{}{age}, args...)...)
		f.Seed, f.Repro = seed, repro
	}
	name := "offline-" + age
	var runs []offlineRun
	var rows []string
	for _, d := range offlineDurations {
		ge, err := freshLoad(base, name)
		if err != nil {
			fail("load_error", "%v", err)
			return rows
		}
		o := offlineRun{d: d, pre: ge.GetState(), gains: map[string]float64{}}
		ok := func() (ok bool) {
			defer func() {
				if rec := recover(); rec != nil {
					fail("offline_panic", "applying %s offline panicked: %v", d, rec)
					res.Failures[len(res.Failures)-1].Detail = string(debug.Stack())
					ok = false
				}
			}()
			ge.SimulateOffline(d)
			return true
		}()
		if !ok {
			continue
		}
		o.post = ge.GetState()
		for _, k := range sortedKeys(o.post.Resources) {
			if o.post.Resources[k].Unlocked {
				o.gains[k] = o.post.Resources[k].Amount - o.pre.Resources[k].Amount
			}
		}
		notes := checkOffline(o, fail)
		notes = append(notes, keepTicking(ge, 200, func(check, msg string) { fail(check, "after %s offline: %s", d, msg) })...)
		runs = append(runs, o)
		gained, top, topV := 0, "-", 0.0
		for k, g := range o.gains {
			if g > 0 {
				gained++
				if g/math.Max(o.pre.Resources[k].Storage, 1) > topV {
					top, topV = fmt.Sprintf("%s +%s", k, num(g)), g/math.Max(o.pre.Resources[k].Storage, 1)
				}
			}
		}
		rows = append(rows, fmt.Sprintf("| %s | %s | %d | %d | %s | %s |", age, d, o.post.Tick-o.pre.Tick, gained, top, orDefault(strings.Join(notes, "; "), "ok")))
	}
	// Longer closures never pay less, and past the cap they pay the same.
	byD := map[time.Duration]offlineRun{}
	for _, o := range runs {
		byD[o.d] = o
	}
	if a, ok := byD[time.Hour]; ok {
		if b, ok := byD[8*time.Hour]; ok {
			for k, g := range a.gains {
				if b.gains[k] < g-1e-9 {
					fail("offline_not_monotone", "%s gained %s over 1h but only %s over 8h", k, num(g), num(b.gains[k]))
				}
			}
		}
	}
	if a, ok := byD[game.MaxOfflineTime]; ok {
		if b, ok := byD[30*time.Hour]; ok {
			if d := firstDiff(a.gains, b.gains, nil); d != "" {
				fail("offline_cap", "30h away paid differently from the %s cap: %s", game.MaxOfflineTime, d)
			}
			if a.post.Tick != b.post.Tick {
				fail("offline_cap", "30h away advanced %d ticks, the %s cap %d", b.post.Tick-b.pre.Tick, game.MaxOfflineTime, a.post.Tick-a.pre.Tick)
			}
		}
	}
	// The real path end to end: a save written 8h ago must load with the
	// same catch-up SimulateOffline gave.
	if o, ok := byD[8*time.Hour]; ok {
		if msg := loadGamePath(base, name, 8*time.Hour, o); msg != "" {
			fail("offline_loadgame_path", "%s", msg)
		}
	}
	if row := offlinePlan(base, name, age, fail); row != "" {
		rows = append(rows, row)
	}
	return rows
}

// offlinePlanBudget is how long a full day of offline catch-up with a build
// plan may take on the wall clock. It runs in steps (game.OfflineStepTicks),
// not tick by tick; a millisecond-scale step keeps it well under a second.
const offlinePlanBudget = 2 * time.Second

// offlinePlan closes the game for a day with a build plan queued: every
// producer of the age, 25 copies each. The catch-up must start plan items as
// the day's income pays for them (not only what the store held when the
// player left), keep the invariants, log a summary, and do it quickly. The
// same closure twice must come back identical.
func offlinePlan(base *game.GameEngine, name, age string, fail func(check, format string, args ...interface{})) string {
	var runs []game.GameState
	var took time.Duration
	var started int
	for i := 0; i < 2; i++ {
		ge, err := freshLoad(base, name)
		if err != nil {
			fail("load_error", "%v", err)
			return ""
		}
		pre := ge.GetState()
		planned := 0
		for _, key := range sortedKeys(pre.Buildings) {
			bs, def := pre.Buildings[key], config.BuildingByKey()[key]
			if bs.Unlocked && def.RequiredAge == age && def.Category == "production" && planned < 6 {
				if n, err := ge.PlanAddBuild(key, 25); err == nil && n > 0 {
					planned++
				}
			}
		}
		if planned == 0 {
			return ""
		}
		before := len(pre.BuildQueue) + totalBuilt(pre)
		start := time.Now()
		ge.SimulateOffline(game.MaxOfflineTime)
		took = time.Since(start)
		post := ge.GetState()
		started = len(post.BuildQueue) + totalBuilt(post) - before
		if started <= 0 {
			fail("offline_plan_idle", "a day away with %d producers planned started none of them", planned)
		}
		summary := false
		for _, l := range post.Log {
			summary = summary || strings.HasPrefix(l.Message, "While you were away your plan started")
		}
		if started > 0 && !summary {
			fail("offline_plan_log", "the plan started %d buildings offline but the log has no summary", started)
		}
		for _, p := range invariantProblems(post, config.BuildingByKey()) {
			fail(p.check, "after a day offline with a plan: %s", p.msg)
		}
		runs = append(runs, post)
	}
	if took > offlinePlanBudget {
		fail("offline_plan_slow", "a day offline with a plan took %s (budget %s)", took, offlinePlanBudget)
	}
	skip := func(p string) bool {
		return p == "Log" || p == "SaveExists" || strings.HasPrefix(p, "Stats.PlayTime")
	}
	if d := firstDiff(runs[0], runs[1], skip); d != "" {
		fail("offline_plan_nondeterministic", "the same day offline with the same plan came back different: %s", d)
	}
	return fmt.Sprintf("| %s | 24h with a plan | %d | - | %d buildings started | %s wall |", age, runs[0].Tick-base.GetState().Tick, started, took.Round(time.Millisecond))
}

// totalBuilt is every building copy standing.
func totalBuilt(st game.GameState) int {
	n := 0
	for _, bs := range st.Buildings {
		n += bs.Count
	}
	return n
}

// freshLoad saves base under name and loads it into a new engine at once,
// so LoadGame's own catch-up (only for saves 5s or older) stays out of it.
func freshLoad(base *game.GameEngine, name string) (*game.GameEngine, error) {
	if err := base.SaveGame(name); err != nil {
		return nil, err
	}
	ge := game.NewGameEngine()
	if err := ge.LoadGame(name); err != nil {
		return nil, err
	}
	return ge, nil
}

// checkOffline checks one closure's catch-up and returns table notes.
func checkOffline(o offlineRun, fail func(check, format string, args ...interface{})) []string {
	var notes []string
	capped := o.d
	if capped > game.MaxOfflineTime {
		capped = game.MaxOfflineTime
	}
	interval := time.Duration(o.pre.TickIntervalMs) * time.Millisecond
	ticks := o.post.Tick - o.pre.Tick
	if interval > 0 {
		want := float64(capped) / float64(interval)
		if math.Abs(float64(ticks)-want) > math.Max(2, want*0.002) {
			fail("offline_ticks", "%s away advanced %d ticks; %s at a %s tick is %.0f", o.d, ticks, capped, interval, want)
		}
	}
	anyRate, anyGain := false, false
	for _, k := range sortedKeys(o.gains) {
		g, pre, post := o.gains[k], o.pre.Resources[k], o.post.Resources[k]
		// Construction finishes while the player is away, so the rate can
		// rise during the closure; it never exceeds the higher of the rates
		// before and after (buildings only complete, they are not lost).
		limit := math.Max(math.Max(pre.Rate, post.Rate), 0) * float64(ticks) * game.OfflineEfficiency
		switch {
		case bad(g):
			fail("offline_nan", "%s away: %s gain is %v", o.d, k, g)
		case g < -1e-9:
			fail("offline_loss", "%s away: %s fell by %s", o.d, k, num(-g))
		case g > float64(limit*(1+1e-9))+1e-6:
			fail("offline_overpaid", "%s away: %s gained %s, more than rate %.4g (the higher of before and after) x %d ticks x %.0f%% = %s", o.d, k, num(g), math.Max(pre.Rate, post.Rate), ticks, game.OfflineEfficiency*100, num(limit))
		}
		if post.Amount > float64(post.Storage*(1+1e-9))+1e-6 {
			fail("offline_over_storage", "%s away: %s is %s over its %s cap", o.d, k, num(post.Amount), num(post.Storage))
		}
		if pre.Rate > 0 && pre.Amount < pre.Storage-1 {
			anyRate = true
			if g > 0 {
				anyGain = true
			} else {
				fail("offline_no_gain", "%s away: %s was producing %+.4g/t below its cap but gained nothing", o.d, k, pre.Rate)
			}
		}
	}
	if anyRate && !anyGain {
		fail("offline_nothing", "%s away: nothing was gained", o.d)
	}
	for _, p := range invariantProblems(o.post, config.BuildingByKey()) {
		fail(p.check, "after %s offline: %s", o.d, p.msg)
	}
	welcome := false
	for _, l := range o.post.Log {
		welcome = welcome || strings.HasPrefix(l.Message, "Welcome back!")
	}
	if !welcome {
		fail("offline_no_welcome", "%s away: no welcome-back log line", o.d)
	}
	if o.d > game.MaxOfflineTime {
		notes = append(notes, "capped at "+game.MaxOfflineTime.String())
	}
	return notes
}

// keepTicking steps ge n ticks and reports a panic, a stalled tick counter
// or a broken invariant through fail. It returns table notes.
func keepTicking(ge *game.GameEngine, n int, fail func(check, msg string)) (notes []string) {
	before := ge.GetState().Tick
	defer func() {
		if rec := recover(); rec != nil {
			fail("tick_panic", fmt.Sprintf("a tick panicked: %v\n%s", rec, debug.Stack()))
			notes = append(notes, "tick panicked")
		}
	}()
	ge.StepTicks(n)
	st := ge.GetState()
	if st.Tick < before+n && st.PendingCatastrophe == "" {
		fail("not_ticking", fmt.Sprintf("%d ticks stepped but the tick counter moved %d", n, st.Tick-before))
	}
	for _, p := range invariantProblems(st, config.BuildingByKey()) {
		fail(p.check, p.msg)
	}
	return nil
}

// loadGamePath writes base's save as if it were d old and loads it through
// LoadGame, whose catch-up must match SimulateOffline's. The signature
// covers the timestamp, so the backdated copy is written unsigned (LoadGame
// accepts unsigned saves as pre-integrity saves).
func loadGamePath(base *game.GameEngine, name string, d time.Duration, want offlineRun) string {
	if err := base.SaveGame(name); err != nil {
		return err.Error()
	}
	path := filepath.Join(game.DataDir(), "saves", name+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return err.Error()
	}
	ts, _ := json.Marshal(time.Now().Add(-d))
	m["timestamp"] = ts
	delete(m, "_sig")
	delete(m, "_proof")
	out, _ := json.Marshal(m)
	back := name + "-backdated"
	if err := os.WriteFile(filepath.Join(game.DataDir(), "saves", back+".json"), out, 0o644); err != nil {
		return err.Error()
	}
	ge := game.NewGameEngine()
	if err := ge.LoadGame(back); err != nil {
		return "LoadGame: " + err.Error()
	}
	st := ge.GetState()
	dt := st.Tick - want.post.Tick
	if dt < -1 || dt > 1 {
		return fmt.Sprintf("a save %s old loaded at tick %d; SimulateOffline(%s) reached %d", d, st.Tick, d, want.post.Tick)
	}
	for _, k := range sortedKeys(want.post.Resources) {
		a, b := st.Resources[k].Amount, want.post.Resources[k].Amount
		// The wall clock moved on while writing and loading: one tick more
		// or less of catch-up is fine.
		slack := float64(math.Abs(float64(dt))*math.Max(want.pre.Resources[k].Rate, 0)*game.OfflineEfficiency) + 1e-6 + float64(1e-9*math.Abs(b))
		if math.Abs(a-b) > slack {
			return fmt.Sprintf("a save %s old loaded with %s %s; SimulateOffline(%s) gave %s", d, k, num(a), d, num(b))
		}
	}
	return ""
}
