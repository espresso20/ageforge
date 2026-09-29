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
	"github.com/espresso20/ageforge/mapmodel"
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
	// AgeTimeout, if set, is a fixed time allowed in every age. Zero derives
	// each age's timeout from the pacing table (see AgeTimeout).
	AgeTimeout time.Duration
	MaxSim     time.Duration // hard cap per run
	Horizon    time.Duration // bot saves instead of investing inside this
	// CheckIn, if set, makes the bot a player who looks at the game only
	// this often (the idle style): each decision point is a visit that
	// plays rounds until nothing more is worth doing (Bot.CheckIn).
	// ApplyStyle sets DecideEvery, Horizon and SoftlockSpan to match.
	CheckIn time.Duration
	// DigestEvery, if set, records the engine's state digest every this
	// many ticks in RunResult.Digests, so two runs of a seed (on different
	// machines) can be compared tick range by tick range: the first
	// differing entry brackets where they parted. The final digest is
	// always recorded (RunResult.StateDigest).
	DigestEvery int
	// NoPlan and NoOverflow switch off, for a check-in player, the two tools
	// the game gives one: leaving a build plan at each visit (Bot.planAhead)
	// and wonder overflow. Both are on by default; the switches exist to
	// measure what each is worth (-no-plan, -no-overflow).
	NoPlan     bool
	NoOverflow bool
	// Deals turns on the bot's faction-deal policy (Bot.Deals, -deals=on).
	// Off by default: deals are a side channel, and the pacing targets are
	// graded on the bot that ignores them.
	Deals bool
	// Army turns on the bot's garrison policy (Bot.Army, -army=on). Off by
	// default, like Deals: the pacing targets are graded on the bot that
	// ignores the army.
	Army bool

	// Pacing is PacingReport (default) or PacingEnforce. In report mode an
	// age past its timeout and a run out of MaxSim are pacing notes, not
	// failures, and play continues past a timeout; in enforce mode they fail
	// the run, and a first-cycle age whose median across seeds leaves the
	// target band fails the set (NewSummary).
	Pacing string
	// LastPassage is how the bot answers the Last Passage: "endure"
	// (default) or "succumb" (takes the Cosmic Legacy while it can).
	LastPassage string
	// InviteCosmic makes the bot Invite the Cosmic Era's harbinger, so its
	// next prestige brings the Last Passage.
	InviteCosmic bool
	// Style labels the run in reports ("greedy", "idle", ...).
	Style string

	// TraceDir, if set, receives trace-<seed>.log with every bot action.
	TraceDir string

	// hook, if set, runs at every decision point before the runner looks at
	// the state; returning true ends the run as done. Scenarios in this
	// package use it (saveload checkpoints, perf sampling).
	hook func(r *runner) bool
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
		MaxSim:       2000 * time.Hour,
		Horizon:      30 * time.Minute,
		Pacing:       PacingReport,
		LastPassage:  "endure",
		Style:        "greedy",
	}
}

// ageTimeout is the time allowed in age before it counts as timed out.
func (c Config) ageTimeout(age string) time.Duration {
	if c.AgeTimeout > 0 {
		return c.AgeTimeout
	}
	return AgeTimeout(age)
}

func (c Config) enforce() bool { return c.Pacing == PacingEnforce }

// Anomaly kinds. Any of these fails the session.
const (
	KindPanic     = "panic"
	KindSoftlock  = "softlock"
	KindInvariant = "invariant"
	// KindPacing is only raised under -pacing enforce.
	KindPacing = "pacing"
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
	// Prestiged marks an age the run left by prestige rather than by
	// advancing. It is not a completed age (a run that prestiges on entering
	// the first allowed age spends no time in it), so it is reported but never
	// graded.
	Prestiged bool `json:"prestiged,omitempty"`
	// TargetSecs is the pacing target (0 when the age has none), Verdict
	// grades Seconds against it, and TimedOut marks an age that ran past
	// its timeout (report mode keeps playing).
	TargetSecs float64 `json:"target_seconds_1x,omitempty"`
	Verdict    string  `json:"verdict,omitempty"`
	TimedOut   bool    `json:"timed_out,omitempty"`
}

// CycleSplit is the time from a fresh start to prestige.
type CycleSplit struct {
	Cycle   int     `json:"cycle"`
	Ticks   int     `json:"ticks"`
	Seconds float64 `json:"seconds_1x"`
	Points  int     `json:"prestige_points"`
	// Expected is what the documented formula pays for the run; a Last
	// Passage ending (Ending endured or succumbed) pays part or none of it.
	Expected   int    `json:"expected_points"`
	Ending     string `json:"ending,omitempty"`
	FinalAge   string `json:"final_age"`
	Prestiged  bool   `json:"prestiged"`
	Succumbed  int    `json:"succumbed"`
	TechsTotal int    `json:"techs"`
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

	// Army defense. EndureBuildingsLost sums the buildings every Endure
	// destroyed, EndureStockKept the stock share each Endure kept (summed:
	// divide by CatastrophesEndured for the mean), EndureGarrison the
	// garrison's share at each Endure (summed likewise). RaidsBlunted,
	// DefenseBuildingsSaved and DefenseWorkersSaved come from the engine's
	// GameStats.Defense tally, summed over prestige cycles.
	EndureBuildingsLost   int     `json:"endure_buildings_lost"`
	EndureStockKept       float64 `json:"endure_stock_kept_sum,omitempty"`
	EndureGarrison        float64 `json:"endure_garrison_sum,omitempty"`
	RaidsBlunted          int     `json:"raids_blunted,omitempty"`
	DefenseBuildingsSaved int     `json:"defense_buildings_saved,omitempty"`
	DefenseWorkersSaved   int     `json:"defense_workers_saved,omitempty"`
	// defSeen is the last GameStats.Defense tally folded in (the engine's
	// resets with the run at prestige or Succumb).
	defSeen game.DefenseTally
}

// RunResult is the outcome of one seed.
type RunResult struct {
	Seed       int64              `json:"seed"`
	Style      string             `json:"style,omitempty"`
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
	// Notes are non-failing observations (report-mode pacing timeouts,
	// budget exhaustion).
	Notes []string `json:"notes,omitempty"`
	// CosmicLegacy is true when the run ended holding the Cosmic Legacy.
	CosmicLegacy bool `json:"cosmic_legacy,omitempty"`
	// StateDigest is the engine's state digest when the run ended
	// (GameEngine.StateDigest): one seed must end on the same digest on
	// every machine. Digests is the trail Config.DigestEvery asks for.
	StateDigest string       `json:"state_digest,omitempty"`
	Digests     []TickDigest `json:"digests,omitempty"`
	// MapDigest is the map model's fingerprint of the final state
	// (mapmodel.Model.Fingerprint): the maps must lay one state out the
	// same way on every machine too.
	MapDigest string `json:"map_digest,omitempty"`
}

// TickDigest is the engine's state digest at a tick (total across cycles).
type TickDigest struct {
	Tick   int    `json:"tick"`
	Digest string `json:"digest"`
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
	// OutcomeSlow ends a run whose age timed out under -pacing enforce.
	OutcomeSlow = "slow"
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
	timedOut   bool // the current age is past its timeout
	byCheck    map[string]*Anomaly
	thread     *HarbingerThread // live harbinger thread being tracked
	stopReason string
	// advancing is set while control calls AdvanceAge, so the age-advance
	// bus handler leaves that advance to control.
	advancing bool
}

// Run plays one seed to completion and returns what happened.
func Run(cfg Config, seed int64) *RunResult {
	start := time.Now()
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	r := newRunner(cfg, seed, ge)
	defer func() {
		r.res.WallMillis = time.Since(start).Milliseconds()
	}()
	if cfg.TraceDir != "" {
		if f, err := os.Create(filepath.Join(cfg.TraceDir, fmt.Sprintf("trace-%d.log", seed))); err == nil {
			w := bufio.NewWriter(f)
			r.bot.Trace = w
			defer func() { w.Flush(); f.Close() }()
		}
	}
	r.play()
	return r.res
}

// newRunner wires a runner and a fresh bot to ge, which the caller has
// already seeded or loaded.
func newRunner(cfg Config, seed int64, ge *game.GameEngine) *runner {
	r := &runner{
		cfg:     cfg,
		seed:    seed,
		ge:      ge,
		res:     &RunResult{Seed: seed, Style: cfg.Style},
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
	r.bot = NewBot(ge)
	r.bot.Harbinger = cfg.Harbinger
	r.bot.HorizonTicks = cfg.Horizon.Seconds() / game.BaseTickInterval.Seconds()
	r.bot.CheckInTicks = cfg.CheckIn.Seconds() / game.BaseTickInterval.Seconds()
	r.bot.UsePlan = cfg.CheckIn > 0 && !cfg.NoPlan
	r.bot.Deals = cfg.Deals
	r.bot.Army = cfg.Army
	if cfg.NoOverflow {
		ge.SetWonderOverflow(false)
	}
	r.subscribe()
	return r
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
	// A build plan's advance item moves the age between decisions: the old
	// age ends at that tick, not at the next visit. (The runner's own
	// AdvanceAge is recorded in control.) No engine calls here: the handler
	// runs under the engine lock.
	bus.Subscribe(game.EventAgeAdvanced, func(e game.EventData) {
		if r.advancing {
			return
		}
		next, _ := e.Payload["new_age"].(string)
		s.AgesAdvanced++
		r.closeAge()
		r.enterAge(game.GameState{Age: next})
	})
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
	for !r.step() {
	}
}

// step runs one tick of the loop, with a decision first every DecideEvery
// ticks, and reports whether the run is over. Panics reach the caller.
func (r *runner) step() bool {
	if r.sim >= r.cfg.MaxSim {
		r.stop(OutcomeBudget)
		msg := fmt.Sprintf("run did not finish within %s simulated", r.cfg.MaxSim)
		if r.cfg.enforce() {
			r.anomaly(KindPacing, "run_budget", msg, r.ge.GetState(), true)
		} else {
			r.note(msg + " (a pacing outcome: reported, not failed, in report mode)")
		}
		return true
	}
	if r.cfg.DigestEvery > 0 && r.ticks%r.cfg.DigestEvery == 0 {
		r.res.Digests = append(r.res.Digests, TickDigest{Tick: r.ticks, Digest: r.ge.StateDigest()})
	}
	if r.ticks%r.cfg.DecideEvery == 0 {
		if r.cfg.hook != nil && r.cfg.hook(r) {
			r.stop(OutcomeDone)
			return true
		}
		st := r.ge.GetState()
		r.observe(st)
		r.res.Stats.foldDefense(st.Military.Saved)
		if r.ticks%r.cfg.CheckEvery < r.cfg.DecideEvery {
			r.checkInvariants(st)
		}
		if r.stopReason != "" {
			return true
		}
		if done := r.control(&st); done {
			r.stop(OutcomeDone)
			return true
		}
		if r.bot.CheckInTicks > 0 {
			r.bot.CheckIn(st)
		} else {
			r.bot.Play(st)
		}
	}
	r.sim += r.ge.StepTicks(1)
	r.ticks++
	return false
}

// note records a non-failing observation once.
func (r *runner) note(msg string) {
	for _, n := range r.res.Notes {
		if n == msg {
			return
		}
	}
	r.res.Notes = append(r.res.Notes, msg)
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
		res.Ages = append(res.Ages, r.split(true))
	}
	func() {
		defer func() { _ = recover() }() // the engine may be wedged after a panic
		st := r.ge.GetState()
		res.CosmicLegacy = st.LastPassage.CosmicLegacy
		res.StateDigest = r.ge.StateDigest()
		res.MapDigest = mapmodel.NewBuilder(nil).Build(&st, nil).Fingerprint()
	}()
}

// split is the pacing record of the current age so far.
func (r *runner) split(unfinished bool) AgeSplit {
	secs := (r.sim - r.ageS0).Seconds()
	a := AgeSplit{Cycle: r.cycle, Age: r.age, Ticks: r.ticks - r.ageT0, Seconds: secs,
		Unfinished: unfinished, TimedOut: r.timedOut, Verdict: Verdict(r.age, secs, !unfinished)}
	if t, ok := Target(r.age); ok {
		a.TargetSecs = t.Seconds()
	}
	return a
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
	r.timedOut = false
}

// closeAge records the age just completed. Its verdict is graded across
// seeds in NewSummary, which is where -pacing enforce fails a set.
func (r *runner) closeAge() {
	r.res.Ages = append(r.res.Ages, r.split(false))
}

// closeAgeByPrestige records the age a prestige left: reported, never graded
// (see AgeSplit.Prestiged).
func (r *runner) closeAgeByPrestige() {
	a := r.split(false)
	a.Prestiged, a.Verdict = true, VerdictNone
	r.res.Ages = append(r.res.Ages, a)
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
	// Float sums in sorted order: map order moves the last bit run to run.
	total := 0.0
	for _, k := range sortedKeys(st.Resources) {
		if rs := st.Resources[k]; rs.Unlocked {
			total += rs.Amount
		}
	}
	if w := st.CurrentAgeWonderKey; w != "" {
		bank := st.Buildings[w].WonderBank
		for _, k := range sortedKeys(bank) {
			total += bank[k]
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
	if limit := r.cfg.ageTimeout(st.Age); !r.timedOut && r.sim-r.ageS0 > limit {
		r.timedOut = true
		if r.cfg.enforce() {
			r.stop(OutcomeSlow)
			r.anomaly(KindPacing, "age_timeout",
				fmt.Sprintf("still in %s after %s (timeout %s); blocked on: %s", st.Age, dur((r.sim-r.ageS0).Seconds()), limit, Blockers(st)), st, true)
		} else {
			r.note(fmt.Sprintf("cycle %d: %s ran past its %s timeout; still progressing, so play went on", r.cycle, st.Age, limit))
		}
	}
}

// control resolves catastrophes, advances ages and prestiges. It reports
// true when the run has reached its goal. st is refreshed after any
// transition.
func (r *runner) control(st *game.GameState) bool {
	if st.LastPassage.Pending && st.PendingCatastrophe == "" {
		// A Last Passage left pending (by a loaded save): answer it.
		r.answerLastPassage(*st)
		*st = r.ge.GetState()
	}
	if r.cfg.InviteCosmic {
		if v := st.Harbinger; v != nil && v.LastPassage && !v.Invited && !v.PassageCame && v.InviteBlocked == "" {
			if r.bot.act("harbinger_invite", "last_passage", r.ge.HarbingerInvite()) {
				*st = r.ge.GetState()
			}
		}
	}
	if st.PendingCatastrophe != "" {
		var err error
		if r.cfg.Catastrophe == "succumb" {
			err = r.ge.Succumb()
			if err == nil {
				r.res.Stats.CatastrophesSuccumbed++
				r.succ++
			}
		} else {
			before := totalBuilt(*st)
			pending := st.PendingEndure
			err = r.ge.Endure()
			if err == nil {
				r.res.Stats.CatastrophesEndured++
				r.res.Stats.EndureBuildingsLost += before - totalBuilt(r.ge.GetState())
				r.res.Stats.EndureStockKept += pending.KeepFrac
				r.res.Stats.EndureGarrison += pending.Garrison
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
		r.advancing = true
		err := r.ge.AdvanceAge()
		r.advancing = false
		if err == nil {
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
		before := *st
		expected := PrestigePoints(before)
		if before.Prestige.PendingPoints != expected {
			r.anomaly(KindInvariant, "prestige_formula",
				fmt.Sprintf("prestige would pay %d points but the documented formula gives %d (age %s, %d milestones, %d techs, %d built, level %d)",
					before.Prestige.PendingPoints, expected, before.Age, before.Milestones.CompletedCount,
					before.Research.TotalResearched, before.Stats.TotalBuilt, before.Prestige.Level), before, true)
		}
		if err := r.ge.DoPrestige(); err != nil {
			r.bot.Errors["prestige"]++
			return false
		}
		ending := "plain"
		if mid := r.ge.GetState(); mid.LastPassage.Pending {
			ending = r.answerLastPassage(mid)
		}
		after := r.ge.GetState()
		if after.Prestige.Level != before.Prestige.Level+1 {
			r.anomaly(KindInvariant, "prestige_incomplete",
				fmt.Sprintf("prestige from %s left the level at %d (was %d)", before.Age, after.Prestige.Level, before.Prestige.Level), after, true)
			*st = after
			return false
		}
		points := after.Prestige.TotalEarned - before.Prestige.TotalEarned
		if ending == "plain" && points != expected {
			r.anomaly(KindInvariant, "prestige_formula",
				fmt.Sprintf("prestige paid %d points but the documented formula gives %d", points, expected), after, true)
		}
		r.checkPrestigeCarry(before, after, ending)
		r.res.Stats.Prestiges++
		r.closeAgeByPrestige()
		r.res.Cycles = append(r.res.Cycles, CycleSplit{
			Cycle: r.cycle, Ticks: r.ticks - r.cycT0, Seconds: (r.sim - r.cycS0).Seconds(),
			Points: points, Expected: expected, Ending: ending, FinalAge: before.Age, Prestiged: true,
			Succumbed: r.succ, TechsTotal: before.Research.TotalResearched,
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

// foldDefense adds what the engine's defense tally gained since the last look.
// A tally smaller than the last one seen means the run reset (prestige or
// Succumb) and counts from zero.
func (s *Stats) foldDefense(t *game.DefenseTally) {
	var cur game.DefenseTally
	if t != nil {
		cur = *t
	}
	if cur.Raids < s.defSeen.Raids || cur.Buildings < s.defSeen.Buildings || cur.Workers < s.defSeen.Workers {
		s.defSeen = game.DefenseTally{}
	}
	s.RaidsBlunted += cur.Raids - s.defSeen.Raids
	s.DefenseBuildingsSaved += cur.Buildings - s.defSeen.Buildings
	s.DefenseWorkersSaved += cur.Workers - s.defSeen.Workers
	s.defSeen = game.DefenseTally{Raids: cur.Raids, Buildings: cur.Buildings, Workers: cur.Workers}
}
