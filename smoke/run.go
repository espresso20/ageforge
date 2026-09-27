// Package smoke plays AgeForge end to end without a UI and reports whether
// anything broke: panics, soft-locks, invariant violations, plus pacing and
// event counts. It drives the real GameEngine through its public API and
// steps ticks synchronously with GameEngine.StepTicks, so a run to the
// prestige age takes seconds instead of days.
//
// cmd/smoke is the command-line front end; `make smoke` runs it.
package smoke

import (
	"bufio"
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

// Config controls one smoke session. Durations are simulated time at 1x.
type Config struct {
	Seeds       []int64
	Catastrophe string // "endure" or "succumb"
	// PrestigeAge is the age at which the bot prestiges; "" means the first
	// age where prestige is allowed.
	PrestigeAge string
	// Cycles is how many times to prestige before stopping.
	Cycles int
	// FinalAge, if set, keeps playing after the last prestige until this age.
	FinalAge string
	// StopAge, if set, ends the run as soon as this age is entered, before
	// any prestige. Useful for short runs and tests.
	StopAge string
	// Harbinger is the bot's policy for harbinger answers (HarbingerIgnore, ...).
	Harbinger string

	DecideEvery  int           // ticks between bot decisions
	CheckEvery   int           // ticks between invariant sweeps
	SoftlockSpan time.Duration // no progress for this long is a soft-lock
	AgeTimeout   time.Duration // more than this in one age is a stall
	MaxSim       time.Duration // hard cap per run
	Horizon      time.Duration // bot saves instead of investing inside this

	// TraceDir, if set, receives trace-<seed>.log with every bot action.
	TraceDir string
}

// DefaultConfig is the quick-mode configuration.
func DefaultConfig() Config {
	return Config{
		Seeds:        []int64{1, 2, 3, 4, 5},
		Catastrophe:  "endure",
		Harbinger:    HarbingerIgnore,
		Cycles:       1,
		DecideEvery:  5,
		CheckEvery:   25,
		SoftlockSpan: 30 * time.Minute,
		AgeTimeout:   48 * time.Hour,
		MaxSim:       2000 * time.Hour,
		Horizon:      30 * time.Minute,
	}
}

// Anomaly kinds. Any of these fails the session.
const (
	KindPanic     = "panic"
	KindSoftlock  = "softlock"
	KindInvariant = "invariant"
)

// Anomaly is one problem found during a run. Repeats of the same check are
// folded into Count; Dump is the state at the first occurrence.
type Anomaly struct {
	Kind    string `json:"kind"`
	Check   string `json:"check"`
	Message string `json:"message"`
	Seed    int64  `json:"seed"`
	Cycle   int    `json:"cycle"`
	Age     string `json:"age"`
	Tick    int    `json:"tick"`
	Count   int    `json:"count"`
	Dump    string `json:"dump,omitempty"`
}

// AgeSplit is the time spent in one age.
type AgeSplit struct {
	Cycle   int     `json:"cycle"`
	Age     string  `json:"age"`
	Ticks   int     `json:"ticks"`
	Seconds float64 `json:"seconds_1x"`
	// Unfinished marks the age a run ended in without advancing.
	Unfinished bool `json:"unfinished,omitempty"`
}

// CycleSplit is the time from a fresh start to prestige.
type CycleSplit struct {
	Cycle      int     `json:"cycle"`
	Ticks      int     `json:"ticks"`
	Seconds    float64 `json:"seconds_1x"`
	Points     int     `json:"prestige_points"`
	FinalAge   string  `json:"final_age"`
	Prestiged  bool    `json:"prestiged"`
	Succumbed  int     `json:"succumbed"`
	TechsTotal int     `json:"techs"`
}

// Stats counts events over a run.
type Stats struct {
	CatastrophesRolled    int            `json:"catastrophes_rolled"`
	CatastrophesEndured   int            `json:"catastrophes_endured"`
	CatastrophesSuccumbed int            `json:"catastrophes_succumbed"`
	EpochEvents           map[string]int `json:"epoch_events"`
	Awakenings            int            `json:"awakenings"`
	Milestones            int            `json:"milestones"`
	TechsResearched       int            `json:"techs_researched"`
	BuildingsCompleted    int            `json:"buildings_completed"`
	AgesAdvanced          int            `json:"ages_advanced"`
	Prestiges             int            `json:"prestiges"`
	TimedEvents           map[string]int `json:"timed_events"`
	StarvationDeaths      int            `json:"starvation_deaths"`
	Actions               map[string]int `json:"bot_actions"`
	ActionErrors          map[string]int `json:"bot_action_errors"`

	HarbingerThreads  map[string]int `json:"harbinger_threads_by_target_epoch"`
	HarbingerHandoffs map[string]int `json:"harbinger_handoffs_by_target_epoch"`
	HarbingerVerdicts map[string]int `json:"harbinger_verdicts"`
	// FalseProphets counts resolved threads revealed as false.
	FalseProphets int `json:"false_prophets_revealed"`
}

// RunResult is the outcome of one seed.
type RunResult struct {
	Seed       int64              `json:"seed"`
	Outcome    string             `json:"outcome"`
	Ticks      int                `json:"ticks"`
	Seconds    float64            `json:"seconds_1x"`
	WallMillis int64              `json:"wall_ms"`
	FinalAge   string             `json:"final_age"`
	Ages       []AgeSplit         `json:"ages"`
	Cycles     []CycleSplit       `json:"cycles"`
	Anomalies  []*Anomaly         `json:"anomalies"`
	Harbingers []*HarbingerThread `json:"harbingers"`
	Stats      Stats              `json:"stats"`
}

// Failed reports whether the run hit a panic, soft-lock or invariant
// violation.
func (r *RunResult) Failed() bool { return len(r.Anomalies) > 0 }

// Outcomes.
const (
	OutcomeDone     = "done"
	OutcomePanic    = "panic"
	OutcomeSoftlock = "softlock"
	OutcomeStalled  = "stalled"
	OutcomeBudget   = "budget"
)

type runner struct {
	cfg  Config
	seed int64
	ge   *game.GameEngine
	bot  *Bot
	res  *RunResult

	ageIdx  map[string]int
	ticks   int // total ticks across cycles
	sim     time.Duration
	cycle   int
	age     string
	ageT0   int
	ageS0   time.Duration
	cycT0   int
	cycS0   time.Duration
	succ    int
	logTick int // highest log tick already scanned
	prevEvt map[string]bool

	// Soft-lock detection: building and research high-water marks, and the
	// resource total at the previous decision. Saving counts as progress, so
	// resources only need to be rising, not above an old peak.
	hiBuild    int
	hiTech     int
	lastRes    float64
	lastProg   time.Duration
	byCheck    map[string]*Anomaly
	thread     *HarbingerThread // live harbinger thread being tracked
	stopReason string
}

// Run plays one seed to completion and returns what happened.
func Run(cfg Config, seed int64) *RunResult {
	r := &runner{
		cfg:     cfg,
		seed:    seed,
		res:     &RunResult{Seed: seed},
		ageIdx:  make(map[string]int),
		byCheck: make(map[string]*Anomaly),
		prevEvt: make(map[string]bool),
		cycle:   1,
	}
	r.res.Stats.EpochEvents = make(map[string]int)
	r.res.Stats.TimedEvents = make(map[string]int)
	r.res.Stats.HarbingerThreads = make(map[string]int)
	r.res.Stats.HarbingerHandoffs = make(map[string]int)
	r.res.Stats.HarbingerVerdicts = make(map[string]int)
	for i, k := range config.AgeOrder() {
		r.ageIdx[k] = i
	}
	start := time.Now()
	defer func() {
		r.res.WallMillis = time.Since(start).Milliseconds()
	}()

	r.ge = game.NewGameEngine()
	r.ge.SeedRNG(seed)
	r.bot = NewBot(r.ge)
	r.bot.Harbinger = cfg.Harbinger
	r.bot.HorizonTicks = cfg.Horizon.Seconds() / game.BaseTickInterval.Seconds()
	if cfg.TraceDir != "" {
		if f, err := os.Create(filepath.Join(cfg.TraceDir, fmt.Sprintf("trace-%d.log", seed))); err == nil {
			w := bufio.NewWriter(f)
			r.bot.Trace = w
			defer func() { w.Flush(); f.Close() }()
		}
	}
	r.subscribe()
	r.play()
	return r.res
}

// subscribe counts bus events. Handlers run under the engine write lock, so
// they only bump counters and never call back into the engine.
func (r *runner) subscribe() {
	s := &r.res.Stats
	bus := r.ge.Bus
	bus.Subscribe(game.EventEpochEventFired, func(e game.EventData) {
		t, _ := e.Payload["event_type"].(string)
		s.EpochEvents[t]++
		if t == "catastrophe" {
			s.CatastrophesRolled++
		}
	})
	bus.Subscribe(game.EventHarbingerArrived, func(e game.EventData) {
		ep, _ := e.Payload["epoch_key"].(string)
		if h, _ := e.Payload["handoff"].(bool); h {
			s.HarbingerHandoffs[ep]++
		} else {
			s.HarbingerThreads[ep]++
		}
	})
	bus.Subscribe(game.EventAwakeningFired, func(game.EventData) { s.Awakenings++ })
	bus.Subscribe(game.EventMilestoneCompleted, func(game.EventData) { s.Milestones++ })
	bus.Subscribe(game.EventResearchDone, func(game.EventData) { s.TechsResearched++ })
	bus.Subscribe(game.EventBuildingBuilt, func(game.EventData) { s.BuildingsCompleted++ })
}

func (r *runner) play() {
	defer func() {
		if rec := recover(); rec != nil {
			r.panicked(rec, debug.Stack())
		}
		r.finish()
	}()

	st := r.ge.GetState()
	r.enterAge(st)
	r.checkStorageFeasible(st)
	for {
		if r.sim >= r.cfg.MaxSim {
			r.stop(OutcomeBudget)
			r.anomaly(KindSoftlock, "run_budget", fmt.Sprintf("run did not finish within %s simulated", r.cfg.MaxSim), st, true)
			return
		}
		if r.ticks%r.cfg.DecideEvery == 0 {
			st = r.ge.GetState()
			r.observe(st)
			if r.ticks%r.cfg.CheckEvery < r.cfg.DecideEvery {
				r.checkInvariants(st)
			}
			if r.stopReason != "" {
				return
			}
			if done := r.control(&st); done {
				r.stop(OutcomeDone)
				return
			}
			r.bot.Play(st)
		}
		r.sim += r.ge.StepTicks(1)
		r.ticks++
	}
}

func (r *runner) stop(outcome string) {
	if r.stopReason == "" {
		r.stopReason = outcome
	}
}

func (r *runner) finish() {
	res := r.res
	res.Outcome = r.stopReason
	res.Ticks = r.ticks
	res.Seconds = r.sim.Seconds()
	res.FinalAge = r.age
	res.Stats.Actions = r.bot.Actions
	res.Stats.ActionErrors = r.bot.Errors
	if r.stopReason != OutcomeDone {
		res.Ages = append(res.Ages, AgeSplit{Cycle: r.cycle, Age: r.age, Ticks: r.ticks - r.ageT0,
			Seconds: (r.sim - r.ageS0).Seconds(), Unfinished: true})
	}
}

func (r *runner) panicked(rec interface{}, stack []byte) {
	r.stop(OutcomePanic)
	dump := func() (s string) {
		defer func() {
			if rec2 := recover(); rec2 != nil {
				s = fmt.Sprintf("(state dump panicked too: %v)", rec2)
			}
		}()
		return Dump(r.ge.GetState(), r.cycle, r.sim)
	}()
	r.res.Anomalies = append(r.res.Anomalies, &Anomaly{
		Kind: KindPanic, Check: "panic", Seed: r.seed, Cycle: r.cycle, Age: r.age, Tick: r.ticks, Count: 1,
		Message: fmt.Sprintf("%v", rec),
		Dump:    dump + "\n\nstack:\n" + string(stack),
	})
}

// anomaly records a problem, folding repeats of the same check per age.
func (r *runner) anomaly(kind, check, msg string, st game.GameState, withDump bool) {
	key := kind + "/" + check + "/" + r.age
	if a, ok := r.byCheck[key]; ok {
		a.Count++
		return
	}
	a := &Anomaly{Kind: kind, Check: check, Message: msg, Seed: r.seed, Cycle: r.cycle, Age: r.age, Tick: r.ticks, Count: 1}
	if withDump {
		a.Dump = Dump(st, r.cycle, r.sim)
	}
	r.byCheck[key] = a
	r.res.Anomalies = append(r.res.Anomalies, a)
}

// enterAge starts the pacing clock for st.Age and resets the progress
// watermarks.
func (r *runner) enterAge(st game.GameState) {
	r.age = st.Age
	r.ageT0, r.ageS0 = r.ticks, r.sim
	r.hiBuild, r.hiTech, r.lastRes = -1, -1, -1
	r.lastProg = r.sim
}

func (r *runner) closeAge() {
	r.res.Ages = append(r.res.Ages, AgeSplit{
		Cycle: r.cycle, Age: r.age,
		Ticks: r.ticks - r.ageT0, Seconds: (r.sim - r.ageS0).Seconds(),
	})
}

// observe updates pacing, event counts and the soft-lock detector.
func (r *runner) observe(st game.GameState) {
	if st.Age != r.age {
		// An age change the runner did not make (Succumb resets to primitive).
		r.closeAge()
		r.enterAge(st)
	}

	for _, ev := range st.ActiveEvents {
		if !r.prevEvt[ev.Key] {
			r.res.Stats.TimedEvents[ev.Key]++
		}
	}
	r.prevEvt = make(map[string]bool, len(st.ActiveEvents))
	for _, ev := range st.ActiveEvents {
		r.prevEvt[ev.Key] = true
	}
	r.trackHarbinger(st)
	for _, l := range st.Log {
		if l.Tick <= r.logTick {
			continue
		}
		if l.Type == "error" && strings.Contains(l.Message, "died of starvation") {
			r.res.Stats.StarvationDeaths++
		}
	}
	if n := len(st.Log); n > 0 {
		r.logTick = st.Log[n-1].Tick
	}

	builds := len(st.BuildQueue)
	for _, b := range st.Buildings {
		builds += b.Count
	}
	total := 0.0
	for _, rs := range st.Resources {
		if rs.Unlocked {
			total += rs.Amount
		}
	}
	if w := st.CurrentAgeWonderKey; w != "" {
		for _, v := range st.Buildings[w].WonderBank {
			total += v
		}
	}
	progressed := false
	if builds > r.hiBuild {
		r.hiBuild, progressed = builds, true
	}
	if st.Research.TotalResearched > r.hiTech {
		r.hiTech, progressed = st.Research.TotalResearched, true
	}
	if total > r.lastRes*(1+1e-9) {
		progressed = true
	}
	r.lastRes = total
	if progressed {
		r.lastProg = r.sim
	}
	if r.sim-r.lastProg > r.cfg.SoftlockSpan {
		r.stop(OutcomeSoftlock)
		r.anomaly(KindSoftlock, "no_progress",
			fmt.Sprintf("no progress in buildings, research or resources for %s; blocked on: %s",
				r.cfg.SoftlockSpan, Blockers(st)), st, true)
		return
	}
	if r.sim-r.ageS0 > r.cfg.AgeTimeout {
		r.stop(OutcomeStalled)
		r.anomaly(KindSoftlock, "age_timeout",
			fmt.Sprintf("still in %s after %s; blocked on: %s", st.Age, r.cfg.AgeTimeout, Blockers(st)), st, true)
	}
}

// control resolves catastrophes, advances ages and prestiges. It reports
// true when the run has reached its goal. st is refreshed after any
// transition.
func (r *runner) control(st *game.GameState) bool {
	if st.PendingCatastrophe != "" {
		var err error
		if r.cfg.Catastrophe == "succumb" {
			err = r.ge.Succumb()
			if err == nil {
				r.res.Stats.CatastrophesSuccumbed++
				r.succ++
			}
		} else {
			err = r.ge.Endure()
			if err == nil {
				r.res.Stats.CatastrophesEndured++
			}
		}
		if err != nil {
			r.anomaly(KindInvariant, "catastrophe_choice", "resolving a pending catastrophe failed: "+err.Error(), *st, true)
		}
		*st = r.ge.GetState()
		if st.Age != r.age {
			r.closeAge()
			r.enterAge(*st)
		}
	}

	if st.AgeReady {
		pendingBefore := st.PendingCatastrophe
		from := st.Age
		if err := r.ge.AdvanceAge(); err == nil {
			after := r.ge.GetState()
			if pendingBefore != "" {
				r.anomaly(KindInvariant, "advance_with_pending_catastrophe",
					fmt.Sprintf("AdvanceAge succeeded from %s while catastrophe %q was pending", from, pendingBefore), after, true)
			}
			if p := after.PendingCatastrophe; p != "" && p != config.EpochForAge(after.Age) {
				r.anomaly(KindInvariant, "stale_pending_catastrophe",
					fmt.Sprintf("after advancing to %s the pending catastrophe is %q, not the new epoch", after.Age, p), after, true)
			}
			r.res.Stats.AgesAdvanced++
			r.closeAge()
			r.enterAge(after)
			if r.cfg.StopAge != "" && r.ageIdx[after.Age] >= r.ageIdx[r.cfg.StopAge] {
				return true
			}
			r.checkStorageFeasible(after)
			*st = after
		} else {
			r.bot.Errors["advance"]++
		}
	}

	if r.cycle > r.cfg.Cycles {
		// Past the last prestige: play on to FinalAge.
		return r.cfg.FinalAge == "" || r.ageIdx[st.Age] >= r.ageIdx[r.cfg.FinalAge]
	}
	if st.Prestige.CanPrestige && (r.cfg.PrestigeAge == "" || r.ageIdx[st.Age] >= r.ageIdx[r.cfg.PrestigeAge]) {
		points := st.Prestige.PendingPoints
		if err := r.ge.DoPrestige(); err != nil {
			r.bot.Errors["prestige"]++
			return false
		}
		r.res.Stats.Prestiges++
		r.closeAge()
		r.res.Cycles = append(r.res.Cycles, CycleSplit{
			Cycle: r.cycle, Ticks: r.ticks - r.cycT0, Seconds: (r.sim - r.cycS0).Seconds(),
			Points: points, FinalAge: st.Age, Prestiged: true, Succumbed: r.succ,
			TechsTotal: st.Research.TotalResearched,
		})
		r.cycle++
		r.succ = 0
		r.cycT0, r.cycS0 = r.ticks, r.sim
		r.buyPrestigeUpgrades()
		*st = r.ge.GetState()
		r.enterAge(*st)
		if r.cycle > r.cfg.Cycles && r.cfg.FinalAge == "" {
			return true
		}
	}
	return false
}

// buyPrestigeUpgrades spends points on the cheapest upgrades first.
func (r *runner) buyPrestigeUpgrades() {
	for i := 0; i < 100; i++ {
		ps := r.ge.GetState().Prestige
		best, bestCost := "", math.MaxInt
		for _, k := range sortedKeys(ps.Upgrades) {
			u := ps.Upgrades[k]
			if u.NextCost > 0 && u.NextCost <= ps.Available && u.NextCost < bestCost {
				best, bestCost = k, u.NextCost
			}
		}
		if best == "" || !r.bot.act("prestige_upgrade", best, r.ge.BuyPrestigeUpgrade(best)) {
			return
		}
	}
}
