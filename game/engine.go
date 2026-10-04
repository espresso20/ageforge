package game

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

const (
	// BaseTickInterval is the un-boosted tick period. Speed bonuses divide this
	// value, so higher bonuses produce shorter intervals (faster ticks).
	BaseTickInterval = 2 * time.Second
	// MinTickInterval is the floor imposed after all speed bonuses are applied.
	// It prevents the tick loop from spinning faster than the UI can render.
	MinTickInterval = 200 * time.Millisecond
	MaxLogSize      = 500

	// productionFloor caps how far a NEGATIVE additive production bonus can drag a
	// rate down. The additive pools (production_all, <res>_rate, gather_rate) are
	// applied as rate *= max(productionFloor, 1+Σ), so a -10% catastrophe debuff
	// (e.g. Reconstruction Effort) actually lands, but stacked penalties can never
	// push production below 10% of its pre-bonus value or flip it negative.
	productionFloor = 0.10

	// productionCap is the SYMMETRIC ceiling on the same pools. Until it existed
	// the additive pools were floored but unbounded above: nothing capped how many
	// stacking timed buffs (faction boons, events, wonders) could pile into
	// production_all or a "<res>_rate", and a soak measured a x20.3 multiplier on
	// knowledge_rate. Both pools are now applied as
	// rate *= clamp(1+Σ, productionFloor, productionCap), so stacked buffs
	// saturate at x3.0 instead of compounding without limit.
	//
	// NOTE: this bounds the applied MULTIPLIER, not the pool sum — the resolver
	// still reports the raw Σ for the breakdown panel, which is what a player
	// wants to see ("you are over the cap"). gather_rate keeps its floor-only
	// treatment: it is an additive re-add on worker output, not a multiplier on a
	// rate, so the same ceiling does not apply.
	productionCap = 3.0

	// Festival (culture sink) tuning. The tick counts here and the black
	// market's cooldown are typed for the base curve and stretched for the
	// current age (config.StretchTicks): from the Bronze Age on a festival
	// lasts, and waits, PacingStretch times as long, so an age holds as many.
	festivalBuffPercent   = 0.20 // +20% production_all while active
	festivalBuffTicks     = 150  // ~5 minutes at 2s/tick
	festivalCooldownTicks = 300  // ~10 minutes between festivals
	festivalMinCost       = 2000.0
	festivalCostFraction  = 0.05 // of culture storage cap

	// Black market (culture sink, high-risk/high-reward) tuning. The player
	// spends a lump of culture for a gamble: on a win the imported resource is
	// paid out at blackMarketWinMult × the stake's gold-value; on a loss the
	// culture is simply gone. Gated behind colonial age + a cooldown.
	blackMarketCostFraction  = 0.10           // of culture storage cap, per deal
	blackMarketMinCost       = 5000.0         // floor so it's meaningful early
	blackMarketWinChance     = 0.55           // probability of a payout
	blackMarketWinMult       = 2.5            // payout = stake_gold_value × this on a win
	blackMarketCooldownTicks = 240            // ~8 minutes between deals
	blackMarketMinAge        = "colonial_age" // smuggling networks open in the colonial era

	// lendEventDisplayTicks is how long the cosmetic "Workers on Loan" / "Under
	// Raid" timed events stay in the active-events panel. They carry no effects —
	// the actual worker/resource changes are applied immediately in processDiplomacy.
	lendEventDisplayTicks = 30

	// buildCostFloor / buildCostCap clamp the build-cost factor the engine derives
	// from the resolver's build_cost pool. build_cost values are negative cost
	// reductions; the factor is clamp(1+Σ, floor, cap). The 0.10 floor means costs
	// can never drop below 10% of base no matter how many reductions stack; the
	// 1.0 cap means a (hypothetical) positive build_cost can't RAISE costs.
	buildCostFloor = 0.10
	buildCostCap   = 1.0
)

// clamp constrains v to [lo, hi].
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// GameEngine is the central coordinator for all game systems. All subsystems
// are accessed through their manager fields rather than as globals, so multiple
// independent engine instances can coexist (useful for tests and prestige resets).
//
// Concurrency model: a single background goroutine calls doTick() while the UI
// and command handler access the engine from other goroutines. All reads and
// writes to mutable fields must be done under ge.mu (RLock for reads, Lock for
// writes). Bus handlers fire synchronously inside doTick while the write lock is
// already held — they MUST NOT call GetState() or any method that would attempt
// to re-acquire the lock.
type GameEngine struct {
	mu sync.RWMutex

	// rules is the ruleset this engine plays by: its ages, buildings, techs,
	// eras and prices. Set at construction; every manager holds the same
	// one. Rebind swaps it under the write lock.
	rules *rules.Set

	tick int
	age  string

	Resources  *ResourceManager
	Buildings  *BuildingManager
	Workers    *WorkerManager
	Research   *ResearchManager
	Military   *MilitaryManager
	Events     *EventManager
	Milestones *MilestoneManager
	Prestige   *PrestigeManager
	Trade      *TradeManager
	Diplomacy  *DiplomacyManager
	Stats      *GameStats
	Bus        *EventBus

	progress   *ProgressManager
	buildQueue []BuildQueueItem
	log        []LogEntry
	running    bool
	stopCh     chan struct{}
	stopOnce   sync.Once
	// loopDone is closed when the current Start loop has returned; nil until
	// Start first runs. Stop waits on it, so once Stop returns no tick,
	// autosave or account flush from that loop is still in flight. Guarded
	// by mu.
	loopDone chan struct{}

	// Permanent bonuses from milestones
	permanentBonuses map[string]float64
	// workerBonus is the worker output bonus the last rates pass applied
	// (gather_rate, floored): what workers add to their buildings is raised
	// by this share. The worker shares routine reads it to know what one
	// more food worker grows (foodWorkerFactor). Derived every pass; not
	// saved.
	workerBonus float64

	// Dynamic tick speed. speedMultiplier is only ever above 1 through the dev
	// console's /speed: players have no speed setting (see playerSpeedCap).
	tickSpeedBonus  float64
	speedMultiplier float64

	// Age advancement — set when requirements are met; player must type 'advance' to proceed
	ageReady bool

	// Starvation tracking — counts consecutive ticks with food <= 0 and active drain
	starvationTicks int

	// Automatic expedition dispatch (see auto_expedition.go).
	//
	//   autoExpeditionTicksLeft — countdown to the next automatic scouting dispatch,
	//     modelled on ActiveRoute.TicksLeft. 0 means a dispatch is DUE and is being
	//     retried each tick until the scouting slot frees up and the cost is covered.
	//     Persisted (MilitarySave.AutoExpeditionTicksLeft) so a reload cannot
	//     save-scum an instant dispatch.
	//   autoExpeditionStarved — true while a due dispatch is being blocked for want
	//     of supplies, so the warning is logged once per dry spell instead of every
	//     tick. Transient: recomputed within a tick of loading.
	autoExpeditionTicksLeft int
	autoExpeditionStarved   bool

	// Save integrity badges (set on load, never persisted separately)
	cheaterBadge bool
	eliteBadge   bool

	// activeSaveName is the slot a bare `save` writes to: the last name explicitly
	// saved or loaded this session. Empty until set; ActiveSaveName falls back to
	// AutosaveName. The periodic autosave does NOT touch this — it's a separate net.
	activeSaveName string

	// activeParentName tracks the parent of the current save in the save-lineage
	// tree (Phase 1: plumbed but always "" — branching populates it in Phase 2).
	activeParentName string

	// Phase 7: result of the most recent age advance transformation pass
	lastAgeAdvanceSummary AgeAdvanceSummary

	// Phase 8: epoch system
	currentEpoch string
	// epochEventFired ensures each epoch fires its roll at most once per civilisation cycle.
	epochEventFired    map[string]bool
	survivedEpochs     map[string]bool // epochs where player chose Endure
	pendingCatastrophe string          // epoch key when catastrophe modal should show; "" otherwise
	// catastropheInvited makes the next prestige from the final epoch bring
	// the Last Passage (armed by the Cosmic Era thread's Invite). Persisted.
	catastropheInvited bool
	// fate is the current era's hidden fate: whether a doom is fated, when it
	// strikes, when its harbinger comes (fate.go). nil in the final epoch and
	// until the first tick rolls it. Persisted; never shown to the player.
	fate *FateSave
	// Harbinger (see harbinger.go). harbinger is the live one, nil when none is
	// present. harbingerArrived records the epochs a harbinger has come in this
	// run (once per epoch). pendingBraceLevel is the Brace level handed to the
	// pending catastrophe, applied by Endure. harbingerHistory holds resolved
	// harbingers; like epochEventHistory it survives Succumb, not prestige.
	harbinger *HarbingerSave
	// parkedHarbinger is the Last Passage's thread while the Cosmic Era's
	// fated doom has the floor (fate.go); it resumes when the doom resolves.
	// Persisted.
	parkedHarbinger   *HarbingerSave
	harbingerArrived  map[string]bool
	pendingBraceLevel int
	harbingerHistory  []HarbingerRecord
	// sessionStart is the state the loaded save left (see SessionMark); nil
	// for a game that was not loaded. Not persisted.
	sessionStart *SessionMark
	// harbingerCheckedEpoch is the epoch the tick hook last looked for a
	// thread in (harbingerTickCheck). Not persisted.
	harbingerCheckedEpoch string
	// pendingLastPassage is set while a prestige from the final epoch waits
	// for Endure or Succumb (last_passage.go); cosmicLegacy is the one-time
	// Cosmic Legacy flag, kept across prestige and Succumb. Both persisted.
	pendingLastPassage bool
	cosmicLegacy       bool
	epochEventHistory  []EpochEventRecord
	// awakeningsFired tracks which one-time Age Awakenings have fired this run, so each
	// fires at most once per prestige cycle and a save/reload does not re-fire. Keyed by
	// AwakeningDef.Key. Cleared on prestige/reset alongside epochEventFired.
	awakeningsFired map[string]bool

	// Ancient Civilization Memory (Trello yn98pTQw): occasionally, early in a NEW
	// prestige run, the player discovers a cache holding a memory of their now-extinct
	// previous civilisation — an offer of one random tech, free of prerequisites but
	// at half research speed. One per run.
	//
	//   ancientMemoryUsed — set true the moment the cache is OFFERED (not just on
	//     accept), so declining still consumes the run's single chance and a save/reload
	//     cannot re-roll it. Reset on DoPrestige/Succumb/Reset so the next run can roll.
	//   pendingMemoryTech — the offered tech key while the accept/decline modal is up;
	//     "" otherwise. Surfaced to the UI via GameState.PendingMemoryTech. Transient
	//     (not persisted): a save taken mid-offer simply re-presents nothing; the run's
	//     chance is already spent via ancientMemoryUsed.
	//   memoryRand — RNG seam for the trigger roll + tech pick; nil means use the
	//     seeded ge.rng. Tests inject their own *rand.Rand to force outcomes.
	ancientMemoryUsed bool
	pendingMemoryTech string
	memoryRand        *rand.Rand

	// Phase 9: catastrophe system — these fields intentionally survive Succumb and Prestige
	// resets so that legacy bonuses and civilization history accumulate across multiple runs.
	legacyBonuses      map[string]bool // epochKey -> true if succumb legacy bonus active
	catastropheHistory []string        // narrative civilization log entries

	// History collector — periodic metric samples for the history overlay.
	History *HistoryCollector

	// account is the per-player identity + meta-progression record, loaded once at
	// boot (the accounts design §2/§8) and held here so the UI/dashboard — already sharing
	// this engine — can reach it via Account(). It is player-level, NOT per-save, so
	// Reset() must NOT clear it (a new game keeps the same player). May be nil if
	// LoadOrCreate failed at boot — account state is non-critical, the game runs anyway.
	account *Account

	// runAccountID is the account that owns the run in memory: the held account when the
	// run was started (StartNewNamedGame) or loaded (LoadGame). SaveGame files the run in
	// that account's slot and stamps it with that ID, and the run records to the account
	// only while that account is the one held, so switching accounts can never file the run
	// or its prestiges and age-ups under another account. "" when no run has been started
	// or loaded under an account (a fresh engine, accountless play, most tests): saves then
	// go to the active slot, as they always did. Reset clears it.
	runAccountID string
	// runOrphaned is set when the account that owns the run is wiped: the run has no slot
	// left, so SaveGame refuses rather than recreate the wiped slot or write elsewhere.
	// Starting or loading a run clears it.
	runOrphaned bool

	// devTouched marks a run the developer console has changed (markDevTouchedLocked).
	// Saved with the run (GameSave.DevTouched), kept through prestige and Succumb, cleared
	// only by a new game. A dev-touched run records nothing to the account.
	devTouched bool

	// Morale system — a managed two-way dial. Range [0.10, moraleCap()];
	// starts at moraleNeutral (0.50). Drives production via moraleMultiplier().
	morale          float64 // 0.10–moraleCap(); starts at moraleNeutral (0.50)
	lowMoraleWarned bool    // true after morale warning fired; reset when morale rises above 0.40

	// festivalReadyTick is the earliest game tick a new festival may be held
	// (cooldown anti-spam for the `festival` culture sink). 0 = ready now.
	festivalReadyTick int

	// blackMarketReadyTick is the earliest tick a black-market deal may run
	// (cooldown anti-spam). 0 = ready now. Both cooldowns are saved, and both
	// reset with the tick counter on prestige, Succumb and Reset.
	blackMarketReadyTick int

	// blackMarketRand is the RNG seam for the black-market win/lose roll; nil
	// means use the seeded ge.rng. Tests inject a seeded *rand.Rand so the
	// risk/reward outcome is deterministic.
	blackMarketRand *rand.Rand

	// Seeded RNG service — the master source for faction-encounter rolls (and the
	// home for future seeded systems). `seed` is generated once on new game,
	// persisted in the save (GameSave.Seed), and restored on load so a run's
	// encounter/buff stream is reproducible from its start. `rng` is (re)built from
	// `seed` via SeedRNG. rand.Rand is not safe for concurrent use, but every roll
	// runs inside doTick's write lock, so no extra synchronisation is needed.
	// The stream position round-trips too: rng draws from rngSrc, which counts
	// its steps; the save records the count and LoadGame replays to it (rng.go).
	// rngOwner is the *rand.Rand built on rngSrc, so a test that swaps in its
	// own ge.rng doesn't get rngSrc's count saved for it.
	seed     int64
	rng      *rand.Rand
	rngSrc   *countingSource
	rngOwner *rand.Rand

	// quip is a second stream seeded from the same master seed, used only for
	// the dim one-line log quips (config.PickLogFlavor and the coin flips that
	// decide whether a quip appears). They land in the log, so they must be
	// reproducible, but a separate stream keeps them from shifting gameplay
	// rolls: adding a quip, or a code path that logs one before a roll, cannot
	// change what the next ge.rng draw returns. Its position is saved like rng's.
	quip      *rand.Rand
	quipSrc   *countingSource
	quipOwner *rand.Rand

	// prose is the recent-history filter for generated flavour. One Stream for
	// the whole log, so an expedition line and a raid line cannot repeat each
	// other's sentence inside a screenful; see expedition_flavor.go. Cosmetic
	// state only — it is not persisted, and a reload simply starts with an empty
	// history.
	prose *flavor.Stream

	// plan is the build plan the engine works through as resources come in
	// (plan.go). Saved; cleared by prestige, Succumb and Reset.
	plan []PlanItem
	// planLog is the plan as written this run, age by age (legacy.go): the
	// Plan Template's source. Saved; folded into the template and cleared by
	// prestige and Succumb, cleared by Reset.
	planLog []PlanTemplateItem
	// wonderOverflowOff turns off banking what the caps would cut off into
	// the current age's wonder (overflow.go). The player's preference: saved,
	// kept across prestige and Succumb, cleared by Reset.
	wonderOverflowOff bool
	// overflowScratch is applyTickRates' reusable list of what the caps cut
	// off this tick, so a tick with overflow allocates nothing for it. Not
	// state: never saved, emptied before each use.
	overflowScratch []overflowLoss
	// Worker shares (shares.go). workerShares is the split the player set,
	// domain → percent (nil: every domain on auto); saved, put back on auto
	// by prestige, Succumb and Reset. autoRecruitOff turns off the routine's
	// recruiting; a preference like wonderOverflowOff. staffHoldUntil is the
	// tick the routine's wait after a worker command ends; saved, so a loaded
	// game waits as the saved one would have.
	workerShares   map[string]float64
	autoRecruitOff bool
	staffHoldUntil int
	// lastK is the Era Mastery speed the last rates pass ran at, so the next
	// one sees a drop and applies the grace rule (noteGraceLocked). Not
	// saved: LoadGame sets it to the loaded age's speed. 0 before any pass.
	lastK float64
}

// BuildQueueItem represents a building under construction
type BuildQueueItem struct {
	BuildingKey string
	TicksLeft   int
	TotalTicks  int
	// FromPlan marks a copy the build plan started: on completion it is
	// staffed (staffPlanCopy, plan.go). omitempty keeps older saves'
	// bytes, and their signatures, unchanged.
	FromPlan bool `json:",omitempty"`
}

// NewGameEngine creates a new game engine on the core ruleset (rules.Core),
// initialised to the Primitive Age. Callers must call Start() to begin the
// tick loop.
func NewGameEngine() *GameEngine { return NewGameEngineWith(rules.Core()) }

// NewGameEngineWith creates a new game engine that plays by set. The engine
// and every manager read their definitions from it and from nowhere else,
// so engines on different sets can run side by side.
func NewGameEngineWith(set *rules.Set) *GameEngine {
	ge := &GameEngine{
		rules:            set,
		age:              "primitive_age",
		Resources:        NewResourceManagerWith(set),
		Workers:          NewWorkerManagerWith(set),
		Research:         NewResearchManagerWith(set),
		Military:         NewMilitaryManagerWith(set),
		Events:           NewEventManagerWith(set),
		Milestones:       NewMilestoneManagerWith(set),
		Prestige:         NewPrestigeManagerWith(set),
		Trade:            NewTradeManagerWith(set),
		Diplomacy:        NewDiplomacyManagerWith(set),
		Stats:            NewGameStats(),
		Bus:              NewEventBus(),
		progress:         NewProgressManagerWith(set),
		permanentBonuses: make(map[string]float64),
		speedMultiplier:  1.0,
		morale:           moraleNeutral,
		stopCh:           make(chan struct{}),
		currentEpoch:     set.EraOf("primitive_age"),
		epochEventFired:  make(map[string]bool),
		harbingerArrived: make(map[string]bool),
		awakeningsFired:  make(map[string]bool),
		survivedEpochs:   make(map[string]bool),
		legacyBonuses:    make(map[string]bool),
		History:          NewHistoryCollector(),
	}
	ge.Buildings = ge.newBuildingManager()
	ge.applyAgeUnlocks("primitive_age")
	// Give starting resources — enough for first hut + a little food
	ge.Resources.Add("food", 25)
	ge.Resources.Add("wood", 50)
	// Startup flavor — the step-by-step onboarding now lives in the Buildings panel
	// (main-screen polish part 1), so the log stays clean for live events.
	ge.addLog("event", "Welcome to AgeForge. You have nothing but your hands.")
	// Subscribe to age advances to record markers in history.
	// IMPORTANT: Bus handlers run under the engine write lock — do NOT call GetState().
	ge.Bus.Subscribe(EventAgeAdvanced, func(e EventData) {
		ageName, _ := e.Payload["new_age"].(string)
		ge.History.MarkAge(ge.tick, ageName)
	})
	// Generate this run's master seed once. LoadGame overrides it with the saved
	// seed; Reset re-rolls a fresh one.
	ge.SeedRNG(newSeed())
	return ge
}

// newBuildingManager is a BuildingManager whose tech-gated buildings read
// this engine's research (whichever ResearchManager it holds at the time).
func (ge *GameEngine) newBuildingManager() *BuildingManager {
	bm := NewBuildingManagerWith(ge.rules)
	bm.researched = func(tech string) bool { return ge.Research != nil && ge.Research.IsResearched(tech) }
	return bm
}

// Rules returns the ruleset the engine plays by. A Set never changes, so the
// caller may keep it and read it without the engine's lock.
func (ge *GameEngine) Rules() *rules.Set {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.rules
}

// Rebind moves the engine and every manager onto set, keeping the run's
// state: the swap a game whose rules change while it runs will make.
// Nothing in play calls it yet. Rates and storage follow on the next tick,
// and snapshots taken before it keep the set they were made from.
func (ge *GameEngine) Rebind(set *rules.Set) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.rules = set
	ge.Resources.Rebind(set)
	ge.Buildings.Rebind(set)
	ge.Workers.Rebind(set)
	ge.Research.Rebind(set)
	ge.Military.Rebind(set)
	ge.Events.Rebind(set)
	ge.Milestones.Rebind(set)
	ge.Prestige.Rebind(set)
	ge.Trade.Rebind(set)
	ge.Diplomacy.Rebind(set)
	ge.progress.Rebind(set)
}

// techLockErr is why a building can't be built yet when a tech it needs is
// not researched (nil when nothing holds it back that way).
func (ge *GameEngine) techLockErr(def config.BuildingDef) error {
	if def.RequiredTech == "" || ge.Research.IsResearched(def.RequiredTech) {
		return nil
	}
	return fmt.Errorf("%s needs %s first. Research it to build here.", def.Name, ge.techName(def.RequiredTech))
}

// techName is a tech's display name (its key if it has no def).
func (ge *GameEngine) techName(key string) string {
	if d, ok := ge.Research.defs[key]; ok {
		return d.Name
	}
	return key
}

// newSeed returns a fresh master seed for a new run.
func newSeed() int64 { return time.Now().UnixNano() }

// SeedRNG sets the master seed and (re)initialises the seeded RNG service from it.
// Called once on new game (a fresh seed), on Reset (a new run → new seed), and on
// load (the saved seed, so the run stays reproducible from its start). Tests call
// it with a fixed seed for deterministic encounter/buff outcomes. Callers that are
// already under the engine write lock may call this directly (it only assigns two
// fields); NewGameEngine/Reset call it while single-threaded or locked.
func (ge *GameEngine) SeedRNG(seed int64) {
	ge.seed = seed
	ge.rngSrc = newCountingSource(seed)
	ge.rng = rand.New(ge.rngSrc)
	ge.rngOwner = ge.rng
	ge.quipSrc = newCountingSource(seed ^ quipSeedSalt)
	ge.quip = rand.New(ge.quipSrc)
	ge.quipOwner = ge.quip
}

// quipSeedSalt separates the quip stream from the gameplay stream so the two
// never replay each other's draws.
const quipSeedSalt int64 = 0x51_9C_0FFE_E0D1

// quipRNG returns the seeded quip stream. An engine not built by NewGameEngine
// gets one derived from whatever seed it has; the gameplay stream is left
// alone, so a test that swapped in its own ge.rng keeps it. Call only under the
// write lock.
func (ge *GameEngine) quipRNG() *rand.Rand {
	if ge.quip == nil {
		ge.quipSrc = newCountingSource(ge.seed ^ quipSeedSalt)
		ge.quip = rand.New(ge.quipSrc)
		ge.quipOwner = ge.quip
	}
	return ge.quip
}

// Seed returns this run's master RNG seed (persisted in the save; surfaced in
// GameState for reproducibility/debugging).
func (ge *GameEngine) Seed() int64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.seed
}

const AutosaveInterval = 60 * time.Second

// moraleCap returns 1.0 + 0.05 per wonder built.
func (ge *GameEngine) moraleCap() float64 {
	cap := 1.0
	ge.Buildings.eachBuilt(func(_ string, count int, def config.BuildingDef) {
		if def.Category == "wonder" {
			cap += float64(0.05 * float64(count))
		}
	})
	return cap
}

// clampMorale clamps ge.morale to [0.10, moraleCap()].
func (ge *GameEngine) clampMorale() {
	if ge.morale < 0.10 {
		ge.morale = 0.10
	}
	if c := ge.moraleCap(); ge.morale > c {
		ge.morale = c
	}
}

// applyMorale adds delta to morale and clamps.
func (ge *GameEngine) applyMorale(delta float64) {
	ge.morale += float64(delta) // callers pass products: round them, no FMA
	ge.clampMorale()
}

// Morale tuning constants. Morale is a managed two-way dial on a continuous
// curve pivoted at moraleNeutral: at the pivot production is untouched
// (preserving the historic economy baseline), above it the bonus you must EARN
// ramps up to +moraleMaxBonus at the cap, and below it the penalty you must
// AVOID ramps down to moraleMinMult at the 0.10 floor.
const (
	moraleNeutral  = 0.50   // starting/settling point; continuous-curve pivot
	moraleMaxBonus = 0.20   // max production bonus at moraleCap()
	moraleMinMult  = 0.50   // production multiplier at the 0.10 morale floor
	moraleDrift    = 0.0008 // per-tick gentle pull back toward moraleNeutral

	faithMoraleFactor = 0.0002 // morale lift per faith/tick produced
	faithMoraleCap    = 0.0040 // max per-tick morale lift from faith rate (saturates ~20 faith/tick; bounds late-game firehose)
)

// moraleMultiplier converts the current morale into a production multiplier
// using a CONTINUOUS curve pivoted at the neutral point (moraleNeutral = 0.50):
//
//   - At 0.50 the multiplier is exactly 1.0 (economy baseline preserved).
//   - Above 0.50 it ramps linearly to 1.0+moraleMaxBonus at moraleCap().
//   - Below 0.50 it ramps linearly down to moraleMinMult at the 0.10 floor.
//
// There is no neutral dead zone any more: any deviation from 0.50 produces a
// small, honest effect that grows with distance. Near the pivot the effect is
// tiny (e.g. 52% -> ~+0.8%); at the extremes it reaches the tuned endpoints
// (+20% at cap, x0.50 at the floor). The downside ramps steeper than the
// upside because the 0.10 floor is closer to the pivot than the cap is.
func (ge *GameEngine) moraleMultiplier() float64 {
	const moraleFloor = 0.10
	m := ge.morale

	if m > moraleNeutral {
		cap := ge.moraleCap()
		span := cap - moraleNeutral
		if span <= 0 {
			return 1.0
		}
		frac := (m - moraleNeutral) / span
		if frac > 1.0 {
			frac = 1.0
		}
		return 1.0 + float64(frac*moraleMaxBonus)
	}

	if m < moraleNeutral {
		span := moraleNeutral - moraleFloor
		if span <= 0 {
			return moraleMinMult
		}
		frac := (moraleNeutral - m) / span
		if frac > 1.0 {
			frac = 1.0
		}
		return 1.0 - float64(frac*(1.0-moraleMinMult))
	}

	return 1.0 // exactly neutral
}

// updateMoraleTick applies per-tick morale changes: starvation penalty,
// over-militarization drain, idle-worker drain, morale-building contribution,
// and a gentle drift back toward neutral. Must be called with the write lock
// held (inside doTick).
func (ge *GameEngine) updateMoraleTick() {
	foodRate := 0.0
	if fr, ok := ge.Resources.resources["food"]; ok {
		foodRate = fr.Rate
	}
	totalPop := ge.Workers.TotalPop()

	// Food deficit penalty. (No generic food-surplus boost any more — being fed
	// is the baseline, not a reward. High morale must be EARNED via buildings
	// and events; this only punishes outright starvation.)
	if foodRate < 0 && ge.Resources.Get("food") <= 0 {
		ge.applyMorale(-0.005)
	}

	// Military ratio — if military workers > 30% of pop, drain morale
	if totalPop > 0 {
		militaryAssigned := 0
		for key, bs := range ge.Buildings.counts {
			if bs == 0 {
				continue
			}
			def, ok := ge.Buildings.defs[key]
			if ok && def.WorkerDomain == "military" {
				militaryAssigned += ge.Workers.GetAssignedCount("military", key)
			}
		}
		ratio := float64(militaryAssigned) / float64(totalPop)
		if ratio > 0.30 {
			over := (ratio - 0.30) * 10
			ge.applyMorale(-0.003 * over)
		}
	}

	// Idle workers > 50% of pop
	if totalPop > 0 {
		idle := ge.Workers.IdleCount("worker")
		if float64(idle)/float64(totalPop) > 0.50 {
			ge.applyMorale(-0.002)
		}
	}

	// Morale-building contribution: each BUILT building with a "morale" effect
	// lifts spirits by existing (FLAT — not worker-scaled). Sum across all built
	// buildings and apply once. Read straight from the building manager's counts
	// and defs (lock-free; we already hold the engine write lock — do NOT call
	// GetState()).
	moraleFromBuildings := 0.0
	ge.Buildings.eachBuilt(func(_ string, count int, def config.BuildingDef) {
		for _, eff := range def.Effects {
			if eff.Type == "morale" {
				moraleFromBuildings += float64(eff.Value * float64(count))
			}
		}
	})
	if moraleFromBuildings != 0 {
		ge.applyMorale(moraleFromBuildings)
	}

	// Faith-rate morale lift: an active faith economy keeps spirits up. Scales with
	// faith PRODUCTION rate (not hoarded stock), capped per tick so a late-game
	// faith firehose can't peg morale in one step. Tunable via the consts above.
	faithRate := 0.0
	if fr, ok := ge.Resources.resources["faith"]; ok {
		faithRate = fr.Rate
	}
	if faithRate > 0 {
		ge.applyMorale(math.Min(faithRate*faithMoraleFactor, faithMoraleCap))
	}

	// Drift gently toward neutral. A stable, fed, non-over-militarized civ with
	// no morale buildings settles at moraleNeutral (~0.50). This makes neutral
	// the resting state: you must keep earning to hold the high-morale bonus,
	// and recover deliberately to escape the low-morale penalty. Move by at most
	// the remaining distance so drift never overshoots/oscillates past neutral.
	if ge.morale > moraleNeutral {
		step := moraleDrift
		if step > ge.morale-moraleNeutral {
			step = ge.morale - moraleNeutral
		}
		ge.applyMorale(-step)
	} else if ge.morale < moraleNeutral {
		step := moraleDrift
		if step > moraleNeutral-ge.morale {
			step = moraleNeutral - ge.morale
		}
		ge.applyMorale(step)
	}

	// Low morale warning (fires once, resets when morale recovers above 0.40)
	if ge.morale < 0.40 && !ge.lowMoraleWarned {
		ge.lowMoraleWarned = true
		ge.addLog("warning", fmt.Sprintf("⚠ Morale low (%s): all production %s. Faith output raises morale.",
			textfmt.Percent(ge.morale), textfmt.SignedPercent(ge.moraleMultiplier()-1)))
	} else if ge.morale >= 0.40 && ge.lowMoraleWarned {
		ge.lowMoraleWarned = false
	}
}

// Start begins the game tick loop in the calling goroutine. It blocks until
// Stop is called. Safe to call again after Stop — the stop channel is
// re-initialised so the engine can restart (e.g. ESC → splash → New Game).
// Stop waits for this function to return.
//
// NOTE: Do not call Start from inside the UI goroutine without a wrapper; it
// blocks indefinitely. Wrap with go ge.Start() or run via the app goroutine.
func (ge *GameEngine) Start() {
	ge.mu.Lock()
	// If this engine was previously stopped, reinitialise the stop channel so
	// Start can be called again after Stop (e.g. ESC → splash → New Game).
	// IMPORTANT: stopOnce must also be reset or the next Stop() call will be
	// a no-op and the tick goroutine will run forever.
	select {
	case <-ge.stopCh:
		ge.stopCh = make(chan struct{})
		ge.stopOnce = sync.Once{}
	default:
	}
	ge.running = true
	// Read the channel once, under the lock: a later restart replaces the
	// field, and this loop must keep listening to the channel it started with.
	stop := ge.stopCh
	done := make(chan struct{})
	ge.loopDone = done
	ge.mu.Unlock()
	defer close(done)

	timer := time.NewTimer(ge.getTickInterval())
	defer timer.Stop()

	lastAutosave := time.Now()

	for {
		select {
		case <-timer.C:
			// select picks at random between ready cases, so a timer that
			// fired alongside Stop could still tick. Check stop first.
			select {
			case <-stop:
				return
			default:
			}
			ge.safeTick()

			// Periodic autosave (outside the tick lock) → overwrite the active save
			// slot, not a fixed "autosave" file. ActiveSaveName takes its own RLock;
			// safe here because we are outside the tick write lock.
			if time.Since(lastAutosave) >= AutosaveInterval {
				if err := ge.SaveGame(ge.ActiveSaveName()); err != nil {
					ge.mu.Lock()
					ge.addLog("warning", fmt.Sprintf("Autosave failed: could not write the save file (%v). Try saving by hand.", err))
					ge.mu.Unlock()
				} else {
					ge.mu.Lock()
					ge.addLog("debug", "Autosave complete")
					ge.mu.Unlock()
				}

				// Account lifetime-stats flush (Phase 6): persist any pending
				// RecordPrestige/RecordAgeReached deltas. MUST be here, outside the
				// tick write lock — Save does file I/O. Use the locking accessor, not
				// ge.account directly, since we're outside ge.mu. No-op when not dirty.
				if acct := ge.Account(); acct != nil {
					if err := acct.FlushIfDirty(); err != nil {
						ge.mu.Lock()
						ge.addLog("warning", fmt.Sprintf("Could not save your lifetime stats (%v). The game tries again at the next autosave.", err))
						ge.mu.Unlock()
					}
				}

				lastAutosave = time.Now()
			}

			timer.Reset(ge.getTickInterval())
		case <-stop:
			return
		}
	}
}

// safeTick wraps doTick with panic recovery to keep the tick goroutine alive
// even if a subsystem panics. Panics are logged as errors rather than crashing
// the entire application.
func (ge *GameEngine) safeTick() {
	defer func() {
		if r := recover(); r != nil {
			ge.mu.Lock()
			ge.addLog("error", fmt.Sprintf("This tick hit an error and was skipped. Save your game and report the bug: %v", r))
			ge.mu.Unlock()
		}
	}()
	ge.doTick()
}

// getTickInterval computes the current tick interval from all speed sources.
// Called by the tick goroutine between ticks, outside the write lock. It takes
// the read lock because tickSpeedBonus and speedMultiplier are also written from
// other goroutines (the /speed dev command, Succumb/Prestige via
// recalculateTickSpeed, a load), not just inside doTick. Must NOT be called with
// ge.mu held; use tickIntervalLocked there.
func (ge *GameEngine) getTickInterval() time.Duration {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.tickIntervalLocked()
}

// tickIntervalLocked is getTickInterval for callers that already hold ge.mu
// (read or write).
func (ge *GameEngine) tickIntervalLocked() time.Duration {
	return ge.tickIntervalWithBonusLocked(ge.tickSpeedBonus)
}

// tickIntervalWithBonusLocked is tickIntervalLocked for a given total
// tick_speed bonus. Caller holds ge.mu (read or write).
func (ge *GameEngine) tickIntervalWithBonusLocked(bonus float64) time.Duration {
	mult := ge.speedMultiplier
	if mult < 1.0 {
		mult = 1.0
	}

	denom := (1.0 + bonus) * mult
	if denom <= 0 {
		// Guard: negative or zero denominator (e.g. tick_speed bonus ≤ -1.0)
		// would produce a timer of +Inf (292 years), freezing the game.
		return MinTickInterval
	}
	interval := time.Duration(float64(BaseTickInterval) / denom)
	if interval < MinTickInterval {
		interval = MinTickInterval
	}
	return interval
}

// recalculateTickSpeed sums all tick_speed bonuses from research, permanent
// bonuses, prestige, and active events. Must be called with the write lock held.
// The result is cached in ge.tickSpeedBonus; getTickInterval reads it.
func (ge *GameEngine) recalculateTickSpeed() {
	oldBonus := ge.tickSpeedBonus
	// tick_speed additive pool from the resolver (research + permanent + prestige
	// + active-event tick_speed). buildResolver reads only write-lock-held state
	// + pure config; recalculateTickSpeed runs under the write lock. UNgated, as
	// before — the (1 + bonus) below applies regardless of sign.
	bonus := ge.buildResolver().AddTotal("tick_speed")
	ge.tickSpeedBonus = bonus

	if bonus != oldBonus {
		mult := ge.speedMultiplier
		if mult < 1.0 {
			mult = 1.0
		}
		// Mirror getTickInterval's guard: a denominator ≤ 0 (tick_speed ≤ -1.0)
		// would yield a +Inf/garbage duration in the debug log. The real interval
		// the loop uses comes from getTickInterval, which guards identically.
		denom := (1.0 + bonus) * mult
		interval := MinTickInterval
		if denom > 0 {
			interval = time.Duration(float64(BaseTickInterval) / denom)
			if interval < MinTickInterval {
				interval = MinTickInterval
			}
		}
		ge.addLog("debug", fmt.Sprintf("Tick speed: +%.0f%% (interval: %dms)", bonus*100, interval.Milliseconds()))
	}
}

// playerSpeedCap is the most game speed a player runs at. Game speed is fixed
// so the calendar paces the game: there is no player speed setting, and
// wonders raise no cap. Only the dev console's /speed goes past it, for the
// session: a load clamps the saved multiplier back (clampPlayerSpeed) and a
// prestige, Succumb or wipe resets it.
const playerSpeedCap = 1.0

// clampPlayerSpeed limits a saved speed multiplier to what a player may run
// at: 1x at least (a NaN counts as 1x), playerSpeedCap at most. A save from
// before the speed setting was retired carries whatever the player had set,
// up to 12x with every wonder built.
func clampPlayerSpeed(mult float64) float64 {
	if !(mult >= 1.0) {
		return 1.0
	}
	return min(mult, playerSpeedCap)
}

// starvationDeathInterval is how many ticks pass between starvation deaths
// while food sits at zero.
const starvationDeathInterval = 5

// ActiveSaveName is the slot a bare `save` writes to: the last name explicitly
// saved or loaded this session, defaulting to AutosaveName until one is set.
func (ge *GameEngine) ActiveSaveName() string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	if ge.activeSaveName == "" {
		return AutosaveName
	}
	return ge.activeSaveName
}

// SetActiveSaveName records the slot a bare `save` should target. Call it after a
// successful explicit `save <name>`; LoadGame sets it directly under its own lock.
func (ge *GameEngine) SetActiveSaveName(name string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.activeSaveName = name
}

// ActiveParentName is the lineage parent of the current save ("" for a root).
func (ge *GameEngine) ActiveParentName() string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.activeParentName
}

// SetActiveParentName records the lineage parent of the current save.
func (ge *GameEngine) SetActiveParentName(name string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.activeParentName = name
}

// SetAccount installs the per-player account (at boot, and after a switch). May be nil.
// The account is player-level state and survives Reset (new game / succumb), so it
// is set here rather than in NewGameEngine or Reset (the accounts design §6).
//
// The account it replaces is flushed once detached, so records it gathered since the
// last autosave are kept; Save writes them into that account's own file, never the new
// one's. The flush runs outside ge.mu (it does file I/O). Never call it from a Bus
// handler or with ge.mu held.
func (ge *GameEngine) SetAccount(a *Account) {
	ge.mu.Lock()
	prev := ge.account
	ge.account = a
	ge.mu.Unlock()
	if prev != nil && prev != a {
		_ = prev.FlushIfDirty()
	}
}

// Account returns the per-player account, or nil if none was loaded at boot.
// Future phases (unlocks, themes, lifetime stats) read it through here.
func (ge *GameEngine) Account() *Account {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.account
}

// AccountID returns the current account's ID, or "" if no account is held. Used by
// buildSaveSnapshot to lazy-stamp the save's account_id on the next write — the
// stamp rides the normal SaveGame path so _sig re-signs over it (the accounts design §6).
func (ge *GameEngine) AccountID() string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	if ge.account == nil {
		return ""
	}
	return ge.account.AccountID
}

// ListAccounts enumerates the available account slots for the start-screen picker
// (Phase B). It is a plain passthrough to the read-only game.ListAccounts(): no engine
// lock is taken — it touches no engine state, only the account files under the data root.
// (It is start-screen plumbing, never called from a Bus handler / under ge.mu, so the
// Bus file-I/O rule isn't in play.)
func (ge *GameEngine) ListAccounts() []AccountSummary {
	return ListAccounts()
}

// SwitchAccount makes the account in slot id the live account: the Accounts panel's switch
// and the `account switch` command (Phase B). It checks the slot first and changes nothing if
// it holds no account. Switching to the account already in use keeps the live object, so
// records it has not flushed are not swapped for an older copy read from disk.
//
// Everything that belongs to the old account stays with it (see changeAccount): its pending
// records are flushed into its own file, and a live run is stopped and saved into its
// owner's slot before the switch. endedRun reports that a live run was stopped; the caller
// takes the player back to the main menu, since the run cannot go on under another account.
// On error nothing is switched.
func (ge *GameEngine) SwitchAccount(id string) (endedRun bool, err error) {
	next, found, err := loadAccountFromSlot(id)
	if err != nil {
		return false, err
	}
	if !found || next == nil {
		return false, fmt.Errorf("There is no account with the ID %s.", id)
	}
	if cur := ge.Account(); cur != nil && cur.AccountID == id {
		return false, makeActive(id)
	}
	return ge.changeAccount(next)
}

// RecoverAccount restores the identity in a recovery code and makes it the live account (the
// `account recover` command). The code's account lands in its own slot: an account already
// there is opened as it is, an empty slot gets a fresh identity-only account, and no other
// slot is written (see ImportRecoveryCode). The switch then works as SwitchAccount's does, so
// the account that was in use keeps everything it had. Recovering the code of the account in
// use changes nothing and returns it. A code that fails its checksum changes nothing.
func (ge *GameEngine) RecoverAccount(code string) (acct *Account, endedRun bool, err error) {
	id, err := RecoveryCodeID(code)
	if err != nil {
		return nil, false, err
	}
	if cur := ge.Account(); cur != nil && cur.AccountID == id {
		return cur, false, nil
	}
	next, err := ImportRecoveryCode(code)
	if err != nil {
		return nil, false, err
	}
	endedRun, err = ge.changeAccount(next)
	if err != nil {
		return nil, endedRun, err
	}
	return next, endedRun, nil
}

// changeAccount is the shared tail of SwitchAccount and RecoverAccount: it moves the engine
// from the live account to next, a different account already saved in its own slot.
//
//  1. The live account's pending records are flushed into its own file. A failure stops
//     here, with nothing changed.
//  2. A live run is stopped and saved into its owner's slot (endLiveRun), so no tick,
//     autosave or record from it can reach next.
//  3. next becomes the active account (pointer included) and is installed; SetAccount
//     flushes the old account once more after detaching it.
func (ge *GameEngine) changeAccount(next *Account) (endedRun bool, err error) {
	if cur := ge.Account(); cur != nil {
		if err := cur.FlushIfDirty(); err != nil {
			return false, fmt.Errorf("the lifetime records of the account in use could not be saved first: %w", err)
		}
	}
	endedRun, err = ge.endLiveRun()
	if err != nil {
		return endedRun, err
	}
	if err := makeActive(next.AccountID); err != nil {
		return endedRun, err
	}
	ge.SetAccount(next)
	return endedRun, nil
}

// Running reports whether the tick loop is going, i.e. a game is being played (Start has
// run and Stop has not). On the main menu and in headless play it is false.
func (ge *GameEngine) Running() bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.running
}

// endLiveRun stops a live run (the tick loop is going) and saves it into the slot of the
// account that owns it, ahead of an account change, and reports whether it did. With no live
// run (the main menu, headless play) it does nothing. If the save fails the loop is started
// again and the error returned, so the game carries on as it was.
func (ge *GameEngine) endLiveRun() (bool, error) {
	if !ge.Running() {
		return false, nil
	}
	ge.Stop()
	if err := ge.SaveGame(ge.ActiveSaveName()); err != nil {
		go ge.Start()
		return false, fmt.Errorf("the game could not be saved first, so it carries on: %w", err)
	}
	return true, nil
}

// ImportAccountExport restores a single-account backup blob into the account's OWN slot
// (Phase C) and returns the account it landed in. It never changes which account is active
// and never touches the run, so an import during play cannot move the game under another
// account.
//
// A backup of the account in use is folded into the live account itself, keeping records
// it has not flushed yet. (Folding it into a second copy read from disk would leave the live
// object stale, and its next save would overwrite the import.) Any other backup lands in its
// own slot through game.ImportAccountExport; the caller decides whether to switch to it.
func (ge *GameEngine) ImportAccountExport(blob []byte, merge bool) (*Account, error) {
	exp, err := decodeAccountExport(blob)
	if err != nil {
		return nil, err
	}
	if live := ge.Account(); live != nil && live.AccountID == exp.AccountID {
		if err := live.importExport(exp, merge); err != nil {
			return nil, err
		}
		return live, nil
	}
	return importExportToSlot(exp, merge)
}

// CreateAccount creates (or, for an existing same-name slot, opens) a name-derived account
// and installs it as ge.account (Phase B). It is the no-carry-over create: a brand-new
// account starts empty (see game.CreateAccount). It is a main-menu operation. The name of
// the account in use opens that account, keeping the live object; otherwise SetAccount
// flushes the account it replaces into that account's own file.
func (ge *GameEngine) CreateAccount(name string) (*Account, error) {
	if cur := ge.Account(); cur != nil && cur.AccountID == AccountIDForName(name) {
		return cur, makeActive(cur.AccountID)
	}
	acct, err := CreateAccount(name)
	if err != nil {
		return nil, err
	}
	ge.SetAccount(acct)
	return acct, nil
}

// ExportAccountByID exports the progress blob for the account in slot id (Phase D). Plain
// passthrough to game.ExportAccountByID: it reads the named slot WITHOUT touching ge.account or
// the active pointer, so the Accounts panel can back up a non-active selection safely. No ge.mu
// — file I/O over the slots, no engine state touched.
func (ge *GameEngine) ExportAccountByID(id string) ([]byte, error) {
	return ExportAccountByID(id)
}

// RecoveryCodeForID returns the recovery code for the account in slot id (Phase D). Plain
// passthrough to game.RecoveryCodeForID — read-only, no active-account or engine-state mutation.
func (ge *GameEngine) RecoveryCodeForID(id string) (string, error) {
	return RecoveryCodeForID(id)
}

// WipeAccountByID deletes the slot for account id (Phase D). It is the by-id sibling of the
// active-only WipeAccount, used by the Accounts panel to wipe the SELECTED account. game.
// WipeAccountByID snapshots the slot into <root>/backups/ BEFORE removal and returns that
// backupPath (empty if the backup failed — the wipe still proceeds), removes only that slot,
// and, if id was the active account, clears the active pointer + in-memory id. When the wiped
// id matches ge's current account we also detach ge.account (under the write lock) so the UI
// re-prompts/refreshes rather than holding a now-orphaned account whose slot is gone. A
// non-active wipe leaves ge.account alone.
//
// The live account is detached BEFORE the files go and without SetAccount's flush: a flush
// would write the account straight back into the slot being wiped. If the wiped account owns
// the run in memory, the run is marked orphaned so a later SaveGame (the exit handler's, say)
// refuses instead of recreating the wiped slot.
func (ge *GameEngine) WipeAccountByID(id string) (string, error) {
	ge.mu.Lock()
	var held *Account
	if id != "" && ge.account != nil && ge.account.AccountID == id {
		held = ge.account
		ge.account = nil
	}
	ownsRun := id != "" && ge.runAccountID == id
	ge.mu.Unlock()

	backupPath, err := WipeAccountByID(id)
	if err != nil {
		if held != nil {
			ge.mu.Lock()
			if ge.account == nil {
				ge.account = held // the wipe failed: put the account back as it was
			}
			ge.mu.Unlock()
		}
		return backupPath, err
	}
	if ownsRun {
		ge.mu.Lock()
		ge.runOrphaned = true
		ge.mu.Unlock()
	}
	return backupPath, nil
}

// StartNewNamedGame resets to a fresh game, makes `name` the active root save,
// and writes the initial save file. It does NOT start the ticker — the caller
// starts it. Returns the SaveGame error if any. The new run belongs to the account
// held now (runAccountID), and its saves go to that account's slot.
func (ge *GameEngine) StartNewNamedGame(name string) error {
	ge.Reset()
	// Set names AFTER Reset so it can't clobber them (Reset leaves them alone, but
	// the ordering keeps that guarantee local to this call).
	ge.SetActiveSaveName(name)
	ge.SetActiveParentName("") // root of a new lineage
	ge.mu.Lock()
	ge.runAccountID = ge.accountIDLocked()
	ge.runOrphaned = false
	ge.mu.Unlock()
	return ge.SaveGame(name)
}

// accountForRecordsLocked returns the account this run's achievements and lifetime stats go
// to, or nil when the run records nothing: accountless play, a run the developer console has
// touched (devTouched), or a run that belongs to another account (the game still in memory
// after an account switch). The dashboard's theme unlocks follow the same rule through
// GameState.AccountRecords. Callers hold ge.mu.
func (ge *GameEngine) accountForRecordsLocked() *Account {
	if ge.account == nil || ge.devTouched {
		return nil
	}
	if ge.runAccountID != "" && ge.runAccountID != ge.account.AccountID {
		return nil
	}
	return ge.account
}

// devTouchedLogLine is logged the first time the developer console changes a run.
const devTouchedLogLine = "Developer commands used: this run won't count toward account records."

// markDevTouchedLocked records that the developer console changed this run. From then on the
// run records nothing to the account: no achievements, no lifetime stats and no theme
// unlocks. The flag is saved with the run, kept through prestige and Succumb (a later run
// inherits the prestige, upgrades and legacies a dev-touched run earned), and cleared only
// by a new game. Dev mode alone does not set it: with the console unlocked but unused, a run
// records as usual. The first mark logs one line. Callers hold ge.mu.
func (ge *GameEngine) markDevTouchedLocked() {
	if ge.devTouched {
		return
	}
	ge.devTouched = true
	ge.addLog("warning", devTouchedLogLine)
}

// markDevTouched is markDevTouchedLocked for callers that do not hold ge.mu (the dev
// console's commands that lock inside their own helpers).
func (ge *GameEngine) markDevTouched() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.markDevTouchedLocked()
}

// Stop halts the game tick loop and waits for it to exit: when Stop returns,
// the Start goroutine has returned and no tick, autosave or account flush of
// its is still running. Safe to call multiple times, and before Start ever
// ran. Must not be called from the tick goroutine or with ge.mu held; nothing
// does (the tick loop never blocks on the UI, so the UI goroutine can wait).
func (ge *GameEngine) Stop() {
	first := false
	ge.stopOnce.Do(func() {
		first = true
		ge.mu.Lock()
		ge.running = false
		stop := ge.stopCh
		ge.mu.Unlock()
		close(stop)
	})

	ge.mu.RLock()
	done := ge.loopDone
	ge.mu.RUnlock()
	if done != nil {
		<-done
	}

	// Flush any pending account lifetime stats on a clean exit so a prestige/age-up
	// since the last autosave isn't lost (Phase 6). After the loop has exited, so it
	// cannot race an autosave's flush; outside ge.mu — Save does I/O — and via the
	// locking accessor. No-op when not dirty.
	if first {
		if acct := ge.Account(); acct != nil {
			_ = acct.FlushIfDirty()
		}
	}
}

// doTick processes one game tick. It holds the write lock for its entire
// duration, which means all Bus handlers that fire here (via Publish) also
// run under the write lock. Bus handlers MUST NOT call GetState() or any
// other method that acquires the lock — doing so will deadlock.
func (ge *GameEngine) doTick() {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.tick++

	// Harbinger and fate: roll the era's fate if it has none (new game,
	// Succumb, prestige, reset), bring a harbinger when one is due, strike at
	// the fated tick; in the final epoch, start the Last Passage thread.
	ge.harbingerTickCheck()

	// Process build queue
	ge.processBuildQueue()
	if len(ge.buildQueue) > 0 {
		ge.addLog("debug", fmt.Sprintf("Build queue: %s in progress", textfmt.Count(len(ge.buildQueue), "item", "items")))
	}

	// Process research
	ge.processResearch()

	// Process random events
	ge.processEvents()

	// Process expeditions
	ge.processExpeditions()

	// Process trade routes
	ge.processTrade()

	// Process diplomacy
	ge.processDiplomacy()

	// Apply building production
	ge.recalculateRates()

	// Snapshot the soldiers amount before rates are applied so we can credit
	// only the soldiers actually trained this tick (post-storage-clamp delta).
	soldiersBefore := ge.Resources.Get("soldiers")

	// Apply resource rates (production - consumption); what a cap cuts off
	// goes to the wonder bank while overflow is on (overflow.go).
	ge.applyTickRates()

	// Credit the lifetime soldiers-trained counter with the post-clamp delta.
	// Soldiers discarded at the storage cap don't count; the helper floors at 0
	// so a net drain never reduces the lifetime total.
	ge.Stats.RecordSoldiersTrained(ge.Resources.Get("soldiers") - soldiersBefore)

	// The build plan starts whatever this tick's income pays for (plan.go).
	ge.runPlanTick()

	// Worker shares (shares.go): idle workers go to work and recruits fill
	// the empty slots, every few ticks.
	if ge.tick%staffEveryTicks == 0 {
		ge.keepSharesLive()
	}

	// Log net food rate and capped resources every 10 ticks
	if ge.tick%10 == 0 {
		snap := ge.Resources.Snapshot()
		if f, ok := snap["food"]; ok {
			ge.addLog("debug", fmt.Sprintf("Food: %.1f (rate %+.3f/t), pop=%d", f.Amount, f.Rate, ge.Workers.TotalPop()))
		}
		for _, key := range sortedKeys(snap) {
			rs := snap[key]
			if rs.Unlocked && rs.Amount >= rs.Storage && rs.Storage > 0 {
				ge.addLog("debug", fmt.Sprintf("Resource at cap: %s (%.0f/%.0f)", key, rs.Amount, rs.Storage))
			}
		}
	}

	// Track gathered amounts in stats
	for key, r := range ge.Resources.Snapshot() {
		if r.Rate > 0 {
			ge.Stats.RecordGather(key, r.Rate)
		}
	}

	// Starvation: when food is at 0 with active drain, one worker dies every
	// starvationDeathInterval ticks.
	if ge.Resources.Get("food") <= 0 && ge.Workers.FoodDrain() > 0 {
		ge.starvationTicks++
		if ge.starvationTicks == 1 {
			ge.addLog("warning", fmt.Sprintf("⚠ Food has run out. One worker dies every %s until you have food again.",
				ge.durationLocked(starvationDeathInterval)))
			if q := config.PickLogFlavor(config.LogFlavorStarvation, ge.quipRNG()); q != "" {
				ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", q))
			}
		}
		if ge.starvationTicks%starvationDeathInterval == 0 && ge.Workers.TotalPop() > 0 {
			killed := ge.Workers.KillWorker(1)
			if killed > 0 {
				ge.addLog("error", fmt.Sprintf("☠ A worker starved to death. Population: %d.", ge.Workers.TotalPop()))
			}
		}
	} else if ge.starvationTicks > 0 {
		ge.starvationTicks = 0
		ge.addLog("info", "✓ Food is back. Workers have stopped starving.")
		if q := config.PickLogFlavor(config.LogFlavorStarvationEnded, ge.quipRNG()); q != "" {
			ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", q))
		}
	}

	// Morale tick — must run after recalculateRates() so foodRate is current
	ge.updateMoraleTick()

	// Periodic debug snapshot every 50 ticks
	if ge.tick%50 == 0 {
		snap := ge.Resources.Snapshot()
		foodAmt := snap["food"].Amount
		foodRate := snap["food"].Rate
		ge.addLog("debug", fmt.Sprintf("Tick %d snapshot: food=%.1f (%+.3f/t), pop=%d, queue=%d",
			ge.tick, foodAmt, foodRate, ge.Workers.TotalPop(), len(ge.buildQueue)))
	}

	// Check milestones
	ge.checkMilestones()

	// Check age advancement — notify once when ready, but require player to
	// type 'advance' to confirm. ageReady resets if requirements drop (e.g.
	// resources consumed) so the notification fires again if they're re-met.
	if nextAge := ge.progress.CheckAdvancement(ge.age, ge.Resources, ge.Buildings); nextAge != "" {
		if !ge.ageReady {
			ge.ageReady = true
			nextName := ge.progress.GetAgeName(nextAge)
			ge.addLog("event", fmt.Sprintf("✦ Ready to advance to the %s. Type 'advance' when you're ready.", nextName))
		}
	} else {
		ge.ageReady = false // requirements dropped — not ready anymore
	}

	// Recalculate tick speed from all sources
	ge.recalculateTickSpeed()

	// Record periodic history sample (every historySampleInterval ticks).
	// All data is read directly from engine fields — no GetState() call here.
	if ge.tick%historySampleInterval == 0 {
		ageOrderMap := ge.progress.GetAgeOrder()
		ageOrder := ageOrderMap[ge.age]
		snap := ge.Resources.Snapshot()
		foodRate := 0.0
		knowRate := 0.0
		if fr, ok := snap["food"]; ok {
			foodRate = fr.Rate
		}
		if kr, ok := snap["knowledge"]; ok {
			knowRate = kr.Rate
		}
		faith := snap["faith"].Amount
		ge.History.Sample(ge.tick, HistorySample{
			Population: float64(ge.Workers.TotalPop()),
			FoodRate:   foodRate,
			KnowRate:   knowRate,
			Faith:      faith,
			ProdAll:    ge.bonusPoolLocked(ge.buildResolver(), "production_all").Applied,
			TickSpeed:  ge.tickSpeedBonus,
			Morale:     ge.morale,
			AgeOrder:   ageOrder,
		})
	}
}

// processResearch handles research tick
func (ge *GameEngine) processResearch() {
	completed := ge.Research.Tick()
	if completed != "" {
		ge.finishResearch(completed)
	} else if ge.Research.currentTech != "" {
		ge.addLog("debug", fmt.Sprintf("Research: %s %d/%d ticks",
			ge.Research.currentTech, ge.Research.totalTicks-ge.Research.ticksLeft, ge.Research.totalTicks))
	}
}

// advanceResearch moves research on by n ticks at once (offline catch-up).
// Reports whether a tech completed.
func (ge *GameEngine) advanceResearch(n int) bool {
	if completed := ge.Research.Advance(n); completed != "" {
		ge.finishResearch(completed)
		return true
	}
	return false
}

// finishResearch logs a completed tech and publishes EventResearchDone.
func (ge *GameEngine) finishResearch(completed string) {
	def := ge.Research.defs[completed]
	ge.addLog("debug", fmt.Sprintf("Research complete: %s", def.Name))
	ge.addLog("success", fmt.Sprintf("Research complete: %s.", def.Name))
	// A bonus a cap keeps from counting says so.
	for _, capped := range ge.capLinesLocked(def.Effects, true) {
		ge.addLog("info", capped)
	}
	// Cosmetic flavour on roughly half of breakthroughs (varies, never spams).
	if ge.quipRNG().Intn(2) == 0 {
		if q := config.PickLogFlavor(config.LogFlavorResearchDone, ge.quipRNG()); q != "" {
			ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", q))
		}
	}
	ge.Bus.Publish(EventData{
		Type:    EventResearchDone,
		Payload: map[string]interface{}{"tech": completed},
	})
}

// processEvents handles random events
func (ge *GameEngine) processEvents() {
	ageOrder := ge.progress.GetAgeOrder()
	triggered, expired := ge.Events.Tick(ge.gameRNG(), ge.tick, ge.age, ageOrder, ge.currentEpoch)

	for _, def := range triggered {
		ge.addLog("debug", fmt.Sprintf("Event triggered: %s (sentiment: %s)", def.Name, def.Sentiment))
		// The log line says how long a timed event really lasts: its
		// duration is stretched with the age (EventManager.Tick).
		line := def.LogText(ge.durationLocked(ge.rules.StretchTicks(ge.age, def.Duration)))
		// Setbacks log as warnings so they don't read like windfalls.
		if def.Sentiment == "bad" {
			ge.addLog("warning", line)
		} else {
			ge.addLog("event", line)
		}
		ge.applyEventEffects(def)
	}

	for _, ae := range expired {
		ge.addLog("debug", fmt.Sprintf("Event expired: %s", ae.Key))
		suffix := buildLossSuffix(ae)
		ge.addLog("info", fmt.Sprintf("%s has ended.%s", ae.Name, suffix))
	}

	// Morale effects from triggered events
	for _, def := range triggered {
		switch def.Sentiment {
		case "good":
			ge.applyMorale(0.04)
		case "bad":
			ge.applyMorale(-0.04)
		}
	}
}

// processExpeditions handles military expedition progress
func (ge *GameEngine) processExpeditions() {
	militaryBonus, expeditionBonus := ge.militaryPower(), ge.expeditionReward()
	for _, cat := range []string{ExpeditionScouting, ExpeditionMilitary} {
		if active := ge.Military.ActiveByCategory(cat); active != nil {
			ge.addLog("debug", fmt.Sprintf("Expedition: %s %d ticks left", active.Name, active.TicksLeft))
		}
	}
	// Tick all active expeditions (one per category); each may resolve this tick.
	for _, res := range ge.Military.Tick(ge.gameRNG(), militaryBonus, expeditionBonus) {
		ge.addLog("debug", fmt.Sprintf("Expedition resolved: %s (rewards: %d types)", res.Key, len(res.Rewards)))
		ge.addLog("event", res.Message)
		// Cosmetic flavour, generated HERE rather than in MilitaryManager because
		// this is where the seeded rng lives. It is additive: res.Message above
		// already carried the mechanical outcome and the loot.
		if q := ge.expeditionFlavorLine(res); q != "" {
			ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", q))
		}
		// Add rewards to resources
		for resource, amount := range res.Rewards {
			ge.Resources.Add(resource, amount)
		}
		// A resolved expedition may turn up a civilization: roll a faction encounter
		// (discovery and/or a specialty-production boon) on the seeded RNG.
		for _, msg := range ge.rollExpeditionEncounter(res.Category, res.Success) {
			ge.addLog("event", msg)
		}
	}

	// The Geographic Society's standing orders. Runs LAST so a party that resolved
	// above has already freed the scouting slot — see processAutoExpeditions.
	ge.processAutoExpeditions()
}

// processTrade handles trade route ticks
func (ge *GameEngine) processTrade() {
	messages := ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy, ge.harborRouteBonus())
	for _, msg := range messages {
		ge.addLog("warning", msg)
	}
}

// harborRouteBonus sums the "trade_route_income" effect across all built harbour
// buildings (each tier adds a flat fractional bonus per instance). The result is
// an additive multiplier applied to every active trade route's imports. Pure
// read of held state — safe under the engine write lock. Runs every tick, so it
// reads the BuildingManager's already-built defs; config.BuildingByKey() would
// rebuild and re-normalize the entire building table on each call.
func (ge *GameEngine) harborRouteBonus() float64 {
	bonus := 0.0
	ge.Buildings.eachBuilt(func(_ string, count int, def config.BuildingDef) {
		for _, eff := range def.Effects {
			if eff.Type == "trade_route_income" {
				bonus += float64(eff.Value * float64(count))
			}
		}
	})
	return bonus
}

// processDiplomacy handles diplomacy ticks
func (ge *GameEngine) processDiplomacy() {
	ageOrder := ge.progress.GetAgeOrder()
	// Mercantile civs warm to trade activity: treat any active trade route as
	// "traded recently" this window. TradeManager.Tick (which calls RecordTrade) runs in
	// the same lock, so reading the active count here is safe.
	tradedRecently := ge.Trade.ActiveRouteCount() > 0
	// Old Friends (legacy.go): remembered civilizations within reach are
	// met before the age fallback would meet them.
	ge.meetOldFriendsLocked()
	messages := ge.Diplomacy.Tick(ge.gameRNG(), ge.age, ageOrder, ge.tick, tradedRecently)
	for _, msg := range messages {
		ge.addLog("event", msg)
	}

	// Apply queued worker-lending side effects to the worker pool.
	for _, req := range ge.Diplomacy.TakePendingLends() {
		ge.Workers.AddLentWorkers(req.Count)
		// Surface as a timed event so it shows in the active-events panel too.
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "worker_lending",
			Name:      "Workers on Loan",
			TicksLeft: lendEventDisplayTicks,
		})
	}
	// Return lent workers whose loans expired (remove from the pool).
	for _, n := range ge.Diplomacy.TakePendingReturns() {
		ge.Workers.KillWorker(n)
	}
	// And the crews faction boons lent, when their time is up.
	ge.returnBoonWorkers()
	ge.applyWarRaids()

	// Embassies passively generate opinion toward non-hostile factions.
	// Total/tick = Σ over embassy-type buildings of:
	//   perWorkerRate × workerCapacity × count × (0.20 + 0.80 × assigned/totalCap)
	// mirroring the production worker-fill curve so a staffed embassy outperforms
	// an empty one. perWorkerRate comes from the building's "opinion" effect.
	totalOpinion := 0.0
	// Every built building with an "opinion" effect counts (the Embassy and
	// the Grand Embassy today), in key order.
	ge.Buildings.eachBuilt(func(key string, count int, def config.BuildingDef) {
		var perWorker float64
		for _, eff := range def.Effects {
			if eff.Type == "opinion" {
				perWorker = eff.Value
				break
			}
		}
		if perWorker <= 0 || def.WorkerCapacity <= 0 {
			return
		}
		assigned := ge.Workers.GetAssignedCount("worker", key)
		totalCap := float64(count * def.WorkerCapacity)
		fill := float64(assigned) / totalCap
		if fill > 1.0 {
			fill = 1.0
		}
		totalOpinion += float64(perWorker * float64(def.WorkerCapacity) * float64(count) * (0.20 + float64(0.80*fill)))
	})
	if totalOpinion > 0 {
		ge.Diplomacy.AddPassiveOpinion(totalOpinion)
	}

	// Faction trade deals: advance the offer timers, re-roll what is due.
	ge.tickFactionDeals()
}

// checkMilestones checks for newly completed milestones and chains
func (ge *GameEngine) checkMilestones() {
	ageOrder := ge.progress.GetAgeOrder()
	researchedTechs := make(map[string]bool, len(ge.Research.researched))
	for key := range ge.Research.researched {
		researchedTechs[key] = true
	}

	// Soldier milestones key off cumulative lifetime soldiers trained, not the
	// live military-worker count. Knowledge workers are still a live domain count.
	soldiersTrained := int(ge.Stats.SoldiersTrained)
	knowledgeCount := ge.Workers.GetDomainCount("knowledge")

	// Count wonders
	wonderCount := 0
	for key, count := range ge.Buildings.counts {
		if def, ok := ge.Buildings.defs[key]; ok && def.Category == "wonder" && count > 0 {
			wonderCount += count
		}
	}

	completed := ge.Milestones.CheckMilestones(
		ge.tick, ge.age, ageOrder,
		ge.Resources, ge.Buildings,
		ge.Workers.TotalPop(),
		ge.Research.ResearchedCount(),
		ge.Stats.TotalBuilt,
		researchedTechs,
		soldiersTrained,
		wonderCount,
		knowledgeCount,
	)

	for _, ms := range completed {
		rewardText := formatMilestoneRewards(ms.Rewards)
		line := fmt.Sprintf("Milestone achieved: %s.", ms.Name)
		if parts := milestoneRewardParts(ms.Rewards); parts != "" {
			line += " " + parts + "."
		}
		ge.addLog("success", line)
		// A reward a cap keeps from counting says so, before it is granted.
		for _, capped := range ge.capLinesLocked(ms.Rewards, false) {
			ge.addLog("info", capped)
		}
		// Cosmetic flavour quip on its own dim line (never replaces the reward text).
		if ms.Flavor != "" {
			ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", ms.Flavor))
		}
		ge.applyMilestoneRewards(ms.Rewards)
		// Publish milestone event
		ge.Bus.Publish(EventData{
			Type: EventMilestoneCompleted,
			Payload: map[string]interface{}{
				"name":        ms.Name,
				"key":         ms.Key,
				"reward_text": rewardText,
				"flavor":      ms.Flavor,
			},
		})
	}

	// Check chains
	newChains := ge.Milestones.CheckChains()
	for _, chain := range newChains {
		// The boost's length is typed for the base curve and stretched for
		// the age it lands in, so it saves the same share of the age.
		boost := ge.stretchTicks(chain.BoostDuration)
		ge.addLog("success", fmt.Sprintf("Chain complete: %s. Title: %s. Game speed %s for %s.",
			chain.Name, chain.Title, textfmt.SignedPercent(chain.BoostValue),
			ge.durationWithSpeedBonusLocked(boost, chain.BoostValue)))
		// Cosmetic flavour quip on its own dim line (never replaces the title/boost).
		if chain.Flavor != "" {
			ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", chain.Flavor))
		}
		ge.startChainBoost(chain, boost)
		// Publish chain event
		ge.Bus.Publish(EventData{
			Type: EventChainCompleted,
			Payload: map[string]interface{}{
				"name":        chain.Name,
				"key":         chain.Key,
				"title":       chain.Title,
				"flavor":      chain.Flavor,
				"boost_ticks": boost,
			},
		})
	}

	// Recalculate title
	ge.Milestones.recalculateTitle()
}

// applyMilestoneRewards grants a completed milestone's rewards: an instant
// resource goes into the store (as much as fits, with a log line when the
// store cut it short), a permanent bonus into permanentBonuses for the rest
// of the run. Under the write lock.
func (ge *GameEngine) applyMilestoneRewards(rewards []config.Effect) {
	granted, fit := map[string]float64{}, map[string]float64{}
	for _, eff := range rewards {
		switch eff.Type {
		case "instant_resource":
			granted[eff.Target] += eff.Value
			fit[eff.Target] += ge.grantLocked(eff.Target, eff.Value)
		case "permanent_bonus":
			ge.permanentBonuses[eff.Target] += eff.Value
		}
	}
	// The reward line states the full grant: say so when a full store took less.
	if line := clippedLine(granted, fit); line != "" {
		ge.addLog("info", line)
	}
}

// startChainBoost starts a completed milestone chain's game speed boost: a
// timed tick_speed event lasting ticks. Under the write lock.
func (ge *GameEngine) startChainBoost(chain config.MilestoneChainDef, ticks int) {
	ge.Events.InjectEvent(ActiveEvent{
		Key:       chain.Key + "_boost",
		Name:      chain.Name + " Speed Boost",
		TicksLeft: ticks,
		Effects: []config.Effect{
			{Type: "tick_speed", Target: "tick_speed", Value: chain.BoostValue},
		},
	})
}

// recalculateRates recalculates all resource production rates
func (ge *GameEngine) recalculateRates() {
	// Reset all rates and breakdowns
	for _, def := range ge.Resources.defs {
		r := ge.Resources.resources[def.Key]
		if r != nil {
			r.Rate = 0
			r.Breakdown = RateBreakdown{}
		}
	}

	// Building production — worker fill ratio applied per building type
	// rate = base × count × (0.20 + 0.80 × assigned/totalCapacity) × moraleMultiplier
	// moraleMultiplier() is a banded curve: 1.0 across the neutral band, up to
	// 1.0+moraleMaxBonus when morale is high, down to moraleMinMult when low.
	mMult := ge.moraleMultiplier()
	production, workerOutput := ge.Buildings.productionWithWorkerOutput(ge.Workers.GetAssignedCount)
	for res, rate := range production {
		moraleRate := float64(rate * mMult)
		r := ge.Resources.resources[res]
		if r != nil {
			r.Rate += moraleRate
			r.Breakdown.BuildingRate += moraleRate
		}
	}

	permanentBonuses := make(map[string]float64)
	for k, v := range ge.permanentBonuses {
		permanentBonuses[k] = v
	}
	// Add prestige bonuses
	for k, v := range ge.Prestige.GetBonuses() {
		permanentBonuses[k] += v
	}
	// Add wonder bonus effects (Type "bonus" — multipliers such as production_all,
	// knowledge_rate, expedition_reward). These are computed dynamically from built
	// wonders each tick rather than stored in permanentBonuses so that save/load
	// and prestige resets don't require special migration logic.
	for k, v := range ge.getWonderBonuses() {
		permanentBonuses[k] += v
	}

	// Additive bonus pools now come from the resolver — the single source of
	// truth shared with Breakdown/UI. buildResolver reads only already-held
	// state + pure config, so it is lock-safe on this write-locked recalc path.
	// Application logic below (the >0 gates, morale via mMult) is unchanged:
	// Phase 3 only moves WHERE the additive sums come from.
	r := ge.buildResolver()

	// Build-cost factor (Fix A): fold the resolver's build_cost additive pool
	// (negative reductions from milestones + a research tech) into a single
	// multiplier and hand it to the BuildingManager. costMult = clamp(1 + Σ
	// build_cost, 0.10, 1.0). GetCost/BuildBatchCost/UpgradeCost all read it, so
	// the charged cost and the displayed cost are computed from the SAME factor.
	ge.Buildings.SetCostMultiplier(poolFactor("build_cost", r.AddTotal("build_cost")))

	// Apply production_all bonus (multiplier on all positive rates).
	// Pool: research + permanent + prestige + wonders + active-event production_all.
	// Fix B: UNgated with a floor. Previously gated `if prodAllBonus > 0`, which
	// silently swallowed negative additive bonuses (e.g. the Reconstruction Effort
	// catastrophe's -0.10 production_all) whenever the player lacked ≥10% positive
	// bonuses. Now always applied as ×clamp(1+Σ, productionFloor, productionCap),
	// so the debuff lands but production can neither drop below 10% of its
	// pre-bonus value nor run away above ×3.0 on stacked buffs.
	prodAllFactor := poolFactor("production_all", r.AddTotal("production_all"))
	if prodAllFactor != 1.0 {
		for _, def := range ge.Resources.defs {
			r := ge.Resources.resources[def.Key]
			if r != nil && r.Rate > 0 {
				r.Rate *= prodAllFactor
			}
		}
	}

	// Apply per-resource rate bonuses (e.g., "gold_rate", "iron_rate").
	// Includes legacy bonuses (stored in permanentBonuses["wood"] etc. after
	// reapplyLegacyBonuses). Fix B: same ungated+floored treatment as above, and
	// the same productionCap ceiling — stacked "<res>_rate" boons were the pool
	// the soak caught running to ×20.
	for _, def := range ge.Resources.defs {
		bonusKey := def.Key + "_rate"
		factor := poolFactor(bonusKey, r.AddTotal(bonusKey))
		if factor != 1.0 {
			r := ge.Resources.resources[def.Key]
			if r != nil && r.Rate > 0 {
				r.Rate *= factor
			}
		}
	}

	// Worker output (gather_rate): workers add that share more to the
	// buildings they staff. Worker output is what staffing adds to a building
	// (workerOutput above: 80% of its listed rate at a full crew, morale
	// included), and the bonus adds Σ gather_rate of it, ON TOP of the
	// multipliers above rather than through them, so it is never under the
	// production cap and never compounds with it. Fix B: ungated and floored,
	// so a negative total takes worker output down to 10% of itself at most.
	//
	// This block read WorkerManager.GetProductionRates, which has returned
	// nothing since worker output was folded into building output: every
	// worker output bonus in the game was dead. It reads the staffing share
	// of building output now.
	gatherDelta := poolFactor("gather_rate", r.AddTotal("gather_rate")) - 1.0
	ge.workerBonus = gatherDelta
	if gatherDelta != 0 {
		for _, def := range ge.Resources.defs {
			made := float64(workerOutput[def.Key] * mMult)
			if made == 0 {
				continue
			}
			if r := ge.Resources.resources[def.Key]; r != nil {
				bonus := float64(made * gatherDelta)
				r.Rate += bonus
				r.Breakdown.WorkerRate += bonus
			}
		}
	}

	// Research production effects (direct production from techs)
	for _, eff := range ge.getAllResearchProductionEffects() {
		if eff.Type == "production" {
			r := ge.Resources.resources[eff.Target]
			if r != nil {
				r.Rate += eff.Value
				r.Breakdown.ResearchRate += eff.Value
			}
		}
	}

	// Active event effects on production
	for _, eff := range ge.Events.GetActiveEffects() {
		if eff.Type == "production" {
			r := ge.Resources.resources[eff.Target]
			if r != nil {
				r.Rate += eff.Value
				r.Breakdown.EventRate += eff.Value
			}
		}
	}

	// Diplomacy trade bonuses on specific resource rates
	for _, def := range ge.Resources.defs {
		bonus := ge.Diplomacy.GetTradeBonus(def.Key)
		if bonus > 0 {
			r := ge.Resources.resources[def.Key]
			if r != nil && r.Rate > 0 {
				tradeBonus := float64(r.Rate * bonus)
				r.Rate += tradeBonus
				r.Breakdown.TradeRate += tradeBonus
			}
		}
	}

	// The Cosmic Legacy (last_passage.go): everything a resource makes × 1.1,
	// after the ×3 caps and every other bonus. Inside the all-production
	// pool it added nothing once the pool was full (the Victorian Age on a
	// typical run); here it counts in every age. It multiplies what is made,
	// before the food drain, so it never deepens a deficit. Its own
	// breakdown line says why.
	if legacy := ge.cosmicLegacyFactor(); legacy != 1 {
		for _, def := range ge.Resources.defs {
			if r := ge.Resources.resources[def.Key]; r != nil && r.Rate > 0 {
				scaled := float64(r.Rate * legacy)
				r.Breakdown.LegacyRate = scaled - r.Rate
				r.Rate = scaled
			}
		}
	}

	// Food consumption
	drain := ge.Workers.FoodDrain()
	if drain > 0 {
		r := ge.Resources.resources["food"]
		if r != nil {
			r.Rate -= drain
			r.Breakdown.FoodDrain = -drain
		}
	}

	// Calculate bonus rates (the difference from multipliers)
	for _, def := range ge.Resources.defs {
		r := ge.Resources.resources[def.Key]
		if r != nil {
			knownComponents := r.Breakdown.BuildingRate + r.Breakdown.WorkerRate +
				r.Breakdown.ResearchRate + r.Breakdown.EventRate + r.Breakdown.TradeRate + r.Breakdown.FoodDrain +
				r.Breakdown.LegacyRate
			r.Breakdown.BonusRate = r.Rate - knownComponents
		}
	}

	// Era Mastery (mastery.go): on known ground the whole economy runs k
	// times faster, so every net rate is multiplied by k at the very end,
	// after the ×3 caps, flat tech and event output, trade bonuses and the
	// food drain. Its own breakdown line says why.
	k := ge.speedK()
	ge.noteGraceLocked(k)
	if k != 1 {
		for _, def := range ge.Resources.defs {
			if r := ge.Resources.resources[def.Key]; r != nil && r.Rate != 0 {
				scaled := float64(r.Rate * k)
				r.Breakdown.MasteryRate = scaled - r.Rate
				r.Rate = scaled
			}
		}
	}

	// Recalculate storage from buildings + research + milestones
	storageBonuses := ge.Buildings.GetStorageBonuses()
	allBonus := storageBonuses["all"]
	// Add storage bonuses from research
	allBonus += ge.Research.StorageBonus("all")
	allBonus += permanentBonuses["all"]

	for _, def := range ge.Resources.defs {
		specific := storageBonuses[def.Key]
		specific += ge.Research.StorageBonus(def.Key)
		specific += permanentBonuses[def.Key]
		r := ge.Resources.resources[def.Key]
		// Storage grows with Era Mastery's k, as production does, so a store
		// holds the same hours of income at any speed.
		r.Storage = float64((def.BaseStorage + allBonus + specific) * k)
		// Storage can shrink (a storage building sold or destroyed). Add clamps
		// on the way in, but a resource with no production never passes through
		// Add again, so without this it sat above its new cap indefinitely.
		// Graced stock (the grace rule: k dropped) is the exception: it stays
		// until spent, and loses the grace once it is under the cap.
		if r.Amount > r.Storage {
			if !ge.Resources.grace[def.Key] {
				r.Amount = r.Storage
			}
		} else if ge.Resources.grace[def.Key] {
			delete(ge.Resources.grace, def.Key)
		}
	}
}

// getAllResearchProductionEffects returns production effects from researched techs
func (ge *GameEngine) getAllResearchProductionEffects() []config.Effect {
	var effects []config.Effect
	// Held defs in sorted order: recalculateRates runs every tick, and callers
	// sum these effects, so the order must not follow the researched map.
	allTechs := ge.Research.defs
	for _, key := range ge.Research.order {
		if !ge.Research.researched[key] {
			continue
		}
		if def, ok := allTechs[key]; ok {
			for _, eff := range def.Effects {
				if eff.Type == "production" {
					effects = append(effects, eff)
				}
			}
		}
	}
	return effects
}

// Age-transition resource carryover tuning (EPIC: age-pacing economy rebalance).
// See advanceAge for the model: each resource is capped to ~a handful of the
// cheapest new-age building rather than a flat percentage of the prior hoard.
const (
	// CarryoverStarterBuildings caps a carried-over resource to roughly this many
	// of the cheapest new-age building that uses it — a small head start.
	CarryoverStarterBuildings = 8
	// CarryoverResidualPct is the fallback fraction kept for resources that no
	// new-age (non-wonder) building uses as a build cost. Kept at the legacy 10%
	// because this branch mostly catches food (the worker-sustain resource) —
	// cutting it harder risks a starvation spiral right at the age transition,
	// and the mass-buy problem this rebalance fixes lives entirely in the
	// build-cost cap above.
	CarryoverResidualPct = 0.10
)

// advanceAge advances to newAge and applies all transition consequences:
//   - Building lineage transformation (old tier → new tier per lineage definition)
//   - Legacy flags for any lower-tier buildings that now have an unlocked replacement
//   - Age-gated unlock application (resources, buildings, workers)
//   - Resource carryover capped to ~a handful of new-age starter buildings
//   - Epoch detection and epoch event roll if the new age crosses an epoch boundary
//
// Caller must hold the write lock.
func (ge *GameEngine) advanceAge(newAge string) {
	oldAge := ge.age
	prevK := ge.speedK()
	ge.age = newAge
	ge.Prestige.NoteAgeEntered(newAge)
	ge.ageReady = false
	ge.Workers.SetAge(newAge)

	// Phase 7: Building lineage transformation pass.
	// Collect transforms first (safe iteration), then apply.
	type pendingTransform struct {
		oldKey, oldName, newKey, newName string
		count                            int
	}
	var transforms []pendingTransform
	for _, key := range sortedKeys(ge.Buildings.counts) {
		count := ge.Buildings.counts[key]
		if count == 0 {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		// Storage never transforms (decision log: storage is cumulative). An
		// upgrade would trade a copy the age lock never lets you rebuild for one
		// of the new tier's capped slots, lowering the most you can ever store.
		if !ok || def.LineageKey == "" || def.LineageKey == "wonder" || def.Category == "storage" {
			continue
		}
		next, ok := ge.rules.NextTier(def.LineageKey, def.LineageTier, newAge)
		if !ok {
			continue
		}
		transforms = append(transforms, pendingTransform{
			oldKey: key, oldName: def.Name,
			newKey: next.Key, newName: next.Name,
			count: count,
		})
	}
	summary := AgeAdvanceSummary{OldAge: oldAge, NewAge: newAge}
	for _, t := range transforms {
		ge.Buildings.SetPendingUpgrade(t.oldKey, t.newKey)
		ge.Buildings.MarkLegacy(t.oldKey)
		summary.BuildingsTransformed = append(summary.BuildingsTransformed, BuildingTransform{
			OldKey: t.oldKey, OldName: t.oldName,
			NewKey: t.newKey, NewName: t.newName,
			Count: t.count,
		})
		ge.addLog("info", fmt.Sprintf("↑ %s can upgrade to %s. Type 'upgrade %s'.",
			buildingCountIn(ge.rules, t.count, t.oldKey), t.newName, t.oldKey))
	}
	// Mark buildings as legacy if their lineage now has a higher-tier unlocked equivalent.
	for _, key := range sortedKeys(ge.Buildings.counts) {
		count := ge.Buildings.counts[key]
		if count == 0 || ge.Buildings.IsLegacy(key) {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		if !ok || def.LineageKey == "" || def.LineageKey == "wonder" {
			continue
		}
		for otherKey, otherDef := range ge.Buildings.defs {
			if otherDef.LineageKey == def.LineageKey &&
				otherDef.LineageTier > def.LineageTier &&
				ge.Buildings.IsUnlocked(otherKey) {
				ge.Buildings.MarkLegacy(key)
				summary.BuildingsLegacy = append(summary.BuildingsLegacy, key)
				break
			}
		}
	}
	ge.lastAgeAdvanceSummary = summary

	ge.applyAgeUnlocks(newAge)
	ge.Stats.RecordAge(newAge)

	// Account lifetime stat (Phase 6): record the highest age reached IN-MEMORY only.
	// advanceAge holds ge.mu, so RecordAgeReached must not do I/O or re-enter the
	// engine; the persisting flush runs later in the autosave block (outside ge.mu).
	// Order comes from the engine's ruleset (no locks) — the account stays
	// config-free and ranks ages by this int rather than re-deriving order itself.
	// A dev-touched run, or one that belongs to another account, records nothing.
	if acct := ge.accountForRecordsLocked(); acct != nil {
		reached, _ := ge.rules.Age(newAge)
		acct.RecordAgeReached(newAge, reached.Order)
	}

	// note: Age-transition carryover model (EPIC: age-pacing economy rebalance).
	// The old flat-10% reduction still left a huge stockpile (10% of a hoard is
	// plenty to mass-buy a new age's buildings). Instead we cap each resource to
	// ~CarryoverStarterBuildings of the CHEAPEST new-age building that uses it —
	// a small head start, not a fresh stockpile. Resources no new-age building
	// uses fall back to a small residual percentage. Players who didn't over-
	// accumulate keep what they had (amount below the cap is untouched).
	// Faith is excluded — it's cumulative.
	// The plan's overflow banks go back to the stores first, so the trim
	// treats them like anything else held: a bank never carries more into the
	// new age than the stores could (overflow.go).
	ge.returnPlanBanks("for the advance")
	entryCosts := ge.rules.AgeEntryCosts(newAge)
	for key, r := range ge.Resources.resources {
		if key == "faith" {
			continue
		}
		if entry, ok := entryCosts[key]; ok && entry > 0 {
			capAmt := CarryoverStarterBuildings * entry
			if r.Amount > capAmt {
				r.Amount = capAmt
			}
			// else: kept as-is — they didn't over-accumulate this resource.
		} else {
			r.Amount *= CarryoverResidualPct
		}
	}
	ge.addLog("info", fmt.Sprintf("Stockpiles trimmed for the new age: each resource is capped at %d times the cost of the cheapest new building that uses it, resources no new building uses keep %s, and faith is untouched.",
		CarryoverStarterBuildings, textfmt.Percent(CarryoverResidualPct)))

	oldName := ge.progress.GetAgeName(oldAge)
	newName := ge.progress.GetAgeName(newAge)
	unlocks := ge.progress.GetUnlocks(newAge)
	ge.addLog("debug", fmt.Sprintf("Age advance: %s → %s (unlocks: %d buildings, %d resources, %d workers)",
		oldAge, newAge, len(unlocks.UnlockBuildings), len(unlocks.UnlockResources), len(unlocks.UnlockVillagers)))
	ge.addLog("success", fmt.Sprintf("Advanced from the %s to the %s.", oldName, newName))
	if line := ge.masteryEntryLine(newAge, prevK); line != "" {
		ge.addLog("info", line)
	}
	// Cosmetic flavour echo in the log (distinct from the age splash quip — see ui/age_splash.go).
	// Age transitions are rare, so it fires every time.
	if q := config.PickLogFlavor(config.LogFlavorAgeAdvance, ge.quipRNG()); q != "" {
		ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", q))
	}

	// Notify player about the wonder available in this age
	for _, bKey := range unlocks.UnlockBuildings {
		if def, ok := ge.Buildings.defs[bKey]; ok && def.Category == "wonder" {
			ge.addLog("event", fmt.Sprintf("★ Wonder available: %s. Bank its cost, then build it.", def.Name))
			break
		}
	}

	ge.Bus.Publish(EventData{
		Type: EventAgeAdvanced,
		Payload: map[string]interface{}{
			"old_age": oldAge,
			"new_age": newAge,
		},
	})

	// Phase 8: detect epoch transition and roll epoch event
	ge.detectEpochTransition(newAge)

	// Age Awakening: one-time epoch awakening on first entry to its trigger age.
	// Fires after the epoch roll so the awakening's deterministic boost lands on top
	// of any epoch-event flavor, and logs as the pivotal "this is a new era" beat.
	ge.fireAwakening(newAge)

	// Harbinger: the new age's figure takes up the epoch's warning, or a new
	// epoch's thread starts. Non-blocking (log, toast, badge), so it never
	// competes with the age splash for focus.
	ge.harbingerOnAgeAdvance()

	// The legacy kit: the new age's template slice and old friends.
	ge.legacyOnAgeEnteredLocked()

	// Age advancement celebration morale boost
	ge.applyMorale(0.08)
}

// applyAgeUnlocks unlocks all content for an age
func (ge *GameEngine) applyAgeUnlocks(ageKey string) {
	age := ge.progress.GetUnlocks(ageKey)
	for _, r := range age.UnlockResources {
		ge.Resources.UnlockResource(r)
	}
	for _, b := range age.UnlockBuildings {
		ge.Buildings.UnlockBuilding(b)
	}
	for _, v := range age.UnlockVillagers {
		ge.Workers.UnlockType(v)
	}
}

// detectEpochTransition checks whether newAge belongs to a different epoch
// and, when an epoch boundary is crossed, fires the epoch event roll and
// rolls the new era's hidden fate. Must be called at the end of advanceAge
// while the engine write lock is held. Each epoch fires its event roll at
// most once per civilisation cycle (epochEventFired prevents double-fire on
// load or re-entry).
func (ge *GameEngine) detectEpochTransition(newAge string) {
	newEpoch := ge.rules.EraOf(newAge)
	if newEpoch == ge.currentEpoch {
		return // same epoch, no transition
	}
	// The advance commands settle the old era's doom before they advance
	// (fateBeforeAdvance). A direct advance (tests, dev tools) may still
	// leave a thread live; its era ends here, so settle it the same way.
	if f := ge.fate; f.open() && f.EpochKey == ge.currentEpoch {
		if ge.fateThread() != nil {
			if f.lying() {
				ge.revealFalseProphet(true)
			} else if ge.pendingCatastrophe == "" {
				ge.fateStrike(true)
			}
		}
	}
	if h := ge.harbinger; h != nil && h.TargetEpoch != "" && h.EpochKey == ge.currentEpoch {
		ge.harbinger = nil // an era's thread cannot outlive its era
	}
	ge.currentEpoch = newEpoch
	ep, _ := ge.rules.Era(newEpoch)
	ge.addLog("event", fmt.Sprintf("[%s]✦ The %s begins. %s[-]", ep.Color, ep.Name, ep.Description))
	ge.Bus.Publish(EventData{
		Type: EventEpochAdvanced,
		Payload: map[string]interface{}{
			"epoch_key":  newEpoch,
			"epoch_name": ep.Name,
			"epoch_icon": ep.Icon,
		},
	})
	ge.rollEpochEvent(newEpoch)
	// The new era's hidden fate, rolled on entry (none in the final epoch).
	ge.rollFate()
}

// fireAwakening fires the one-time Age Awakening for newAge, if one triggers on that
// age and it has not already fired this run. An awakening is deterministic (always
// grants its modest thematic boost, no downside) and one-time per prestige cycle:
// the awakeningsFired set guards against double-fire on re-entry or save/reload.
//
// The boost is delivered through the existing ActiveEvent / InjectEvent mechanism so
// it decays after Duration ticks and surfaces in the active-events panel like any
// other timed event. Must be called under the engine write lock (advanceAge holds it);
// it touches no lock-acquiring methods and is safe in that path.
func (ge *GameEngine) fireAwakening(newAge string) {
	def, ok := ge.rules.Awakening(newAge)
	if !ok {
		return // no awakening triggers on this age
	}
	if ge.awakeningsFired[def.Key] {
		return // already fired this run
	}
	ge.awakeningsFired[def.Key] = true

	ge.Events.InjectEvent(ActiveEvent{
		Key:       def.Key,
		Name:      def.Name,
		TicksLeft: def.Duration,
		Effects:   def.Effects,
	})

	ep, _ := ge.rules.Era(def.EpochKey)
	// One log line per awakening — the pivotal "new era" beat. Coloured by the epoch
	// so the awakening visually belongs to the era it ushers in.
	// The flavor text states the boost and its duration, so no separate effect line.
	ge.addLog("event", fmt.Sprintf("[%s]✦ Awakening: %s. %s[-]", ep.Color, def.Name, def.FlavorText))
	for _, capped := range ge.capLinesLocked(def.Effects, true) {
		ge.addLog("info", capped)
	}

	ge.Bus.Publish(EventData{
		Type: EventAwakeningFired,
		Payload: map[string]interface{}{
			"awakening_key":  def.Key,
			"awakening_name": def.Name,
			"epoch_key":      def.EpochKey,
		},
	})
}

// rollEpochEvent performs the epoch transition event roll.
//   - Faith fill % gates good-event probability (see epochGoodChance).
//   - Otherwise a challenging (non-catastrophe) bad event is applied
//     immediately. A transition never brings a catastrophe: an era's doom is
//     fated on entry and strikes inside it (fate.go).
//
// The roll comes from the seeded ge.rng. Must be called under engine write
// lock.
func (ge *GameEngine) rollEpochEvent(epochKey string) {
	// Prevent double-fire per epoch
	if ge.epochEventFired[epochKey] {
		return
	}
	ge.epochEventFired[epochKey] = true

	if ge.gameRNG().Float64() < ge.epochGoodChance() {
		ge.rollGoodEpochEvent()
		return
	}
	ge.rollChallengingEpochEvent(epochKey)
}

// rollGoodEpochEvent picks a good epoch event gated by culture fill %.
//   - >40% culture fill → major+minor events eligible.
//   - >75% culture fill with 15% chance → all tiers (legendary) eligible.
//
// Must be called under engine write lock.
func (ge *GameEngine) rollGoodEpochEvent() {
	cultureStorage := ge.Resources.GetStorage("culture")
	tier := "minor"
	if cultureStorage > 0 {
		culturePct := ge.Resources.Get("culture") / cultureStorage
		if culturePct > 0.75 && ge.gameRNG().Float64() < 0.15 {
			tier = "legendary"
		} else if culturePct > 0.40 {
			tier = "major"
		}
	}

	pool := ge.rules.GoodEraEvents()
	var eligible []config.EpochEventDef
	for _, ev := range pool {
		// The Cultural Festival pays in culture, which is locked until the
		// Classical Age: entering the Iron Era (the Iron Age) it would
		// promise culture the player cannot hold. It waits for culture.
		if ev.Key == "cultural_festival" && !ge.Resources.IsUnlocked("culture") {
			continue
		}
		switch tier {
		case "legendary":
			eligible = append(eligible, ev) // all tiers available
		case "major":
			if ev.Type == "good_minor" || ev.Type == "good_major" {
				eligible = append(eligible, ev)
			}
		default: // minor
			if ev.Type == "good_minor" {
				eligible = append(eligible, ev)
			}
		}
	}
	if len(eligible) == 0 {
		return
	}
	ev := eligible[ge.gameRNG().Intn(len(eligible))]
	ge.applyGoodEpochEvent(ev)

	ep, _ := ge.rules.Era(ge.currentEpoch)
	record := EpochEventRecord{
		EpochKey: ge.currentEpoch, EpochName: ep.Name,
		EventKey: ev.Key, EventName: ev.Name, EventType: ev.Type,
		Tick: ge.tick,
	}
	ge.epochEventHistory = append(ge.epochEventHistory, record)
	ge.Bus.Publish(EventData{
		Type:    EventEpochEventFired,
		Payload: map[string]interface{}{"event_key": ev.Key, "event_name": ev.Name, "event_type": ev.Type},
	})
}

// rollChallengingEpochEvent picks a bad (non-catastrophe) epoch event.
func (ge *GameEngine) rollChallengingEpochEvent(epochKey string) {
	pool := ge.rules.ChallengingEraEvents()
	if len(pool) == 0 {
		return
	}
	ev := pool[ge.gameRNG().Intn(len(pool))]
	ge.applyChallengingEpochEvent(ev, epochKey)

	ep, _ := ge.rules.Era(epochKey)
	record := EpochEventRecord{
		EpochKey: epochKey, EpochName: ep.Name,
		EventKey: ev.Key, EventName: ev.Name, EventType: ev.Type,
		Tick: ge.tick,
	}
	ge.epochEventHistory = append(ge.epochEventHistory, record)
	ge.Bus.Publish(EventData{
		Type:    EventEpochEventFired,
		Payload: map[string]interface{}{"event_key": ev.Key, "event_name": ev.Name, "event_type": ev.Type},
	})
}

// applyGoodEpochEvent applies the effects of a good epoch transition event. The
// headline's flavor text states each effect; "→" lines under it report only
// what the text cannot know (actual counts, amounts, names).
func (ge *GameEngine) applyGoodEpochEvent(ev config.EpochEventDef) {
	ge.addLog("success", fmt.Sprintf("✦ %s. %s", ev.Name, ev.FlavorText))
	ageOrder := ge.progress.GetAgeOrder()
	switch ev.Key {
	case "age_of_plenty":
		// ×2 all production for Duration ticks via production_all effect
		ge.injectEpochEffects("epoch_age_of_plenty", ev,
			[]config.Effect{{Type: "production_all", Value: 1.0}})
		ge.logCapped(config.Effect{Type: "production_all", Value: 1.0})
	case "population_surge":
		// +15% workers across all domains, instant
		before := ge.Workers.TotalPop()
		ge.Workers.AddPctAll(0.15)
		if gained := ge.Workers.TotalPop() - before; gained > 0 {
			ge.addLog("success", fmt.Sprintf("  → %s joined (population now %d).",
				textfmt.Count(gained, "new worker", "new workers"), ge.Workers.TotalPop()))
		}
	case "ancient_cache":
		// Fill 40% of each resource's storage cap
		const cacheFill = 0.40
		for _, def := range ge.Resources.defs {
			cap := ge.Resources.GetStorage(def.Key)
			if cap > 0 && ge.Resources.IsUnlocked(def.Key) {
				ge.Resources.Add(def.Key, cap*cacheFill)
			}
		}
	case "trade_winds":
		// Flat +5 gold/tick for Duration ticks
		ge.injectEpochEffects("epoch_trade_winds", ev,
			[]config.Effect{{Type: "production", Target: "gold", Value: 5.0}})
	case "cultural_festival":
		// Instant culture +30%, faith +20%; timed production boost
		culture := ge.gainResource("culture", ge.Resources.Get("culture")*0.30)
		faith := ge.gainResource("faith", ge.Resources.Get("faith")*0.20)
		if gained := Amounts(map[string]float64{"culture": culture, "faith": faith}); gained != "nothing" {
			ge.addLog("success", fmt.Sprintf("  → +%s.", gained))
		}
		ge.injectEpochEffects("epoch_cultural_festival", ev, []config.Effect{
			{Type: "production", Target: "culture", Value: 1.0},
			{Type: "production", Target: "faith", Value: 1.0},
		})
	case "grand_discovery":
		// Complete 3 free techs from current age
		completed := ge.Research.ForceCompleteN(3, ge.age, ageOrder)
		for _, key := range completed {
			ge.addLog("success", fmt.Sprintf("  → Free tech: %s.", ge.rules.Name(rules.KindTech, key)))
		}
		if len(completed) == 0 {
			ge.addLog("success", "  → No techs were left to research this age.")
		}
	case "worker_innovation":
		// Permanent +10% production_all
		ge.permanentBonuses["production_all"] += 0.10
		ge.logCapped(config.Effect{Type: "production_all", Value: 0.10})
	case "architects_gift":
		// 10 free buildings of the most common built non-wonder type
		const giftCount = 10
		bestKey := ""
		bestCount := 0
		for _, key := range sortedKeys(ge.Buildings.counts) {
			count := ge.Buildings.counts[key]
			if def, ok := ge.Buildings.defs[key]; ok && def.Category != "wonder" && count > bestCount {
				bestKey = key
				bestCount = count
			}
		}
		if bestKey != "" {
			ge.Buildings.counts[bestKey] += giftCount
			ge.addLog("success", fmt.Sprintf("  → %s, free.", buildingCountIn(ge.rules, giftCount, bestKey)))
		}
	case "peaceful_century":
		// +20% all production for Duration ticks
		ge.injectEpochEffects("epoch_peaceful_century", ev,
			[]config.Effect{{Type: "production_all", Value: 0.20}})
		ge.logCapped(config.Effect{Type: "production_all", Value: 0.20})
	case "epoch_blessing":
		// Permanent +15% production_all; recorded as a golden age
		ge.permanentBonuses["production_all"] += 0.15
		ge.logCapped(config.Effect{Type: "production_all", Value: 0.15})
	}
}

// applyChallengingEpochEvent applies a challenging (non-catastrophe) bad epoch
// event. As with good events, "→" lines report only what the flavor text
// cannot state (what burned, who died, how much was lost).
func (ge *GameEngine) applyChallengingEpochEvent(ev config.EpochEventDef, epochKey string) {
	ge.addLog("warning", fmt.Sprintf("⚠ %s. %s", ev.Name, ev.FlavorText))
	switch ev.Key {
	case "the_famine":
		ge.injectEpochEffects("epoch_famine", ev,
			[]config.Effect{{Type: "production", Target: "food", Value: -3.0}})
	case "merchant_betrayal":
		ge.loseShare("gold", 0.50)
		ge.injectEpochEffects("epoch_merchant_betrayal", ev,
			[]config.Effect{{Type: "production", Target: "gold", Value: -2.0}})
	case "the_great_fire":
		destroyed, _ := ge.Buildings.DestroyRandom(ge.gameRNG(), 8)
		ge.releaseWorkersFrom(destroyed)
		var lost []string
		for _, key := range sortedKeys(destroyed) {
			lost = append(lost, buildingCountIn(ge.rules, destroyed[key], key))
		}
		if len(lost) > 0 {
			ge.addLog("warning", fmt.Sprintf("  → Destroyed: %s.", textfmt.List(lost)))
		} else {
			ge.addLog("warning", "  → The fire burned out before it reached any buildings.")
		}
	case "epidemic":
		before := ge.Workers.TotalPop()
		ge.Workers.RemovePct(0.20)
		if lost := before - ge.Workers.TotalPop(); lost > 0 {
			ge.addLog("warning", fmt.Sprintf("  → %s lost.", textfmt.Count(lost, "worker", "workers")))
		}
		ge.injectEpochEffects("epoch_epidemic", ev,
			[]config.Effect{{Type: "production", Target: "food", Value: -1.5}})
	case "resource_drought":
		// Debuff epoch's primary resource
		primaryRes := "wood" // fallback
		if ep, ok := ge.rules.Era(epochKey); ok {
			primaryRes = ep.PrimaryResource
		}
		drought := []config.Effect{{Type: "production", Target: primaryRes, Value: -3.0}}
		ge.injectEpochEffects("epoch_resource_drought", ev, drought)
		// The text cannot name the resource (it depends on the epoch), so say it.
		ge.logTimedEffects("warning", drought, ev.Duration)
	case "political_instability":
		ge.loseShare("faith", 0.60)
		ge.injectEpochEffects("epoch_political_instability", ev, []config.Effect{
			{Type: "production", Target: "knowledge", Value: -2.0},
		})
	case "economic_crash":
		ge.loseShare("gold", 0.50)
		ge.injectEpochEffects("epoch_economic_crash", ev,
			[]config.Effect{{Type: "production", Target: "gold", Value: -3.0}})
	case "the_dark_age":
		if tech, ok := ge.Research.CancelResearch(); ok {
			ge.addLog("warning", fmt.Sprintf("  → Research on %s canceled (no refund).", ge.rules.Name(rules.KindTech, tech)))
		}
		ge.loseShare("knowledge", 0.80)
		ge.injectEpochEffects("epoch_dark_age", ev,
			[]config.Effect{{Type: "production", Target: "knowledge", Value: -3.0}})
	}
}

// injectEpochEffects starts an epoch event's timed effects. It logs nothing:
// the event's flavor text already states them. Caller holds ge.mu.
func (ge *GameEngine) injectEpochEffects(key string, ev config.EpochEventDef, effects []config.Effect) {
	ge.Events.InjectEvent(ActiveEvent{
		Key:       key,
		Name:      ev.Name,
		TicksLeft: ev.Duration,
		Effects:   effects,
	})
}

// logTimedEffects logs a timed boost or penalty as
// "  → All production +25% for ~6m." Caller holds ge.mu.
func (ge *GameEngine) logTimedEffects(logType string, effects []config.Effect, ticks int) {
	if text := effectsText(effects); text != "" {
		ge.addLog(logType, fmt.Sprintf("  → %s for %s.", textfmt.Capitalize(text), ge.durationLocked(ticks)))
	}
}

// logCapped logs a line for an effect that has just joined its pool and
// that a cap keeps from counting in full (capLinesLocked); nothing when it
// all counts. Caller holds ge.mu.
func (ge *GameEngine) logCapped(effects ...config.Effect) {
	for _, capped := range ge.capLinesLocked(effects, true) {
		ge.addLog("info", capped)
	}
}

// loseShare removes a fraction of a resource's stock and logs
// "  → Lost 500 gold (50% of your stock)." Caller holds ge.mu.
func (ge *GameEngine) loseShare(res string, frac float64) {
	loss := ge.Resources.Get(res) * frac
	if loss <= 0 || !ge.Resources.Remove(res, loss) {
		return
	}
	ge.addLog("warning", fmt.Sprintf("  → Lost %s (%s of your stock).", Amount(loss, res), textfmt.Percent(frac)))
}

// gainResource adds to a resource and returns what storage actually took.
func (ge *GameEngine) gainResource(res string, amount float64) float64 {
	before := ge.Resources.Get(res)
	return ge.Resources.Add(res, amount) - before
}

// effectsText describes timed event effects: "all production +25%",
// "food -3/tick and gold +5/tick". Unknown effect types are left out.
func effectsText(effects []config.Effect) string {
	var parts []string
	for _, e := range effects {
		switch e.Type {
		case "production_all":
			parts = append(parts, "all production "+textfmt.SignedPercent(e.Value))
		case "production":
			parts = append(parts, ResourceName(e.Target)+" "+textfmt.Rate(e.Value))
		}
	}
	return textfmt.List(parts)
}

// FestivalStatus is the snapshot the `festival` command renders: the live cost,
// the player's current culture, the buff parameters, and cooldown state.
type FestivalStatus struct {
	Cost          float64 // culture required to hold a festival right now
	Culture       float64 // player's current culture
	BuffPercent   float64 // production_all bonus the festival grants (e.g. 0.20)
	BuffTicks     int     // how long the buff lasts
	CooldownTicks int     // cooldown imposed after a festival
	CooldownLeft  int     // ticks remaining on the current cooldown (0 if ready)
	Ready         bool    // true when not on cooldown
	// CapNote says what the all-production cap would leave of the buff if a
	// festival were held now: "" when all of it would count (CapNote in
	// caps.go). The festival costs culture, so the command shows it before
	// the player pays.
	CapNote string
}

// stretchTicks re-times a base-curve tick count for the current age
// (config.StretchTicks). Caller holds ge.mu.
func (ge *GameEngine) stretchTicks(ticks int) int {
	return ge.rules.StretchTicks(ge.age, ticks)
}

// festivalCost returns the culture cost of a festival at the current progression:
// max(festivalMinCost, festivalCostFraction × culture storage cap). It scales with
// the player's culture cap so it stays a meaningful drain into the late game.
func (ge *GameEngine) festivalCost() float64 {
	cultureCap := ge.Resources.GetStorage("culture")
	cost := cultureCap * festivalCostFraction
	if cost < festivalMinCost {
		cost = festivalMinCost
	}
	return cost
}

// FestivalStatus returns the live festival cost, the player's culture, the buff
// parameters, and cooldown state for the `festival` command UI. Read-only.
func (ge *GameEngine) FestivalStatus() FestivalStatus {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	cd := ge.festivalReadyTick - ge.tick
	if cd < 0 {
		cd = 0
	}
	return FestivalStatus{
		CapNote:       ge.capNoteLocked(config.Effect{Type: "production_all", Value: festivalBuffPercent}, false),
		Cost:          ge.festivalCost(),
		Culture:       ge.Resources.Get("culture"),
		BuffPercent:   festivalBuffPercent,
		BuffTicks:     ge.stretchTicks(festivalBuffTicks),
		CooldownTicks: ge.stretchTicks(festivalCooldownTicks),
		CooldownLeft:  cd,
		Ready:         cd == 0,
	}
}

// DoFestival spends a lump of culture to inject a temporary empire-wide
// production buff (+festivalBuffPercent production_all for festivalBuffTicks).
// It is gated by a cooldown so it can't be spammed every tick. This is the
// repeatable, player-initiated culture sink; prestige gates remain primary.
func (ge *GameEngine) DoFestival() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.tick < ge.festivalReadyTick {
		return fmt.Errorf("The next festival is ready in %s.", ge.durationLocked(ge.festivalReadyTick-ge.tick))
	}
	cost := ge.festivalCost()
	have := ge.Resources.Get("culture")
	if have < cost {
		return fmt.Errorf("Not enough culture for a festival: need %s, have %s.", textfmt.Number(cost), textfmt.Number(have))
	}
	ge.Resources.Remove("culture", cost)
	buff := ge.stretchTicks(festivalBuffTicks)
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "cultural_festival",
		Name:      "Cultural Festival",
		TicksLeft: buff,
		Effects:   []config.Effect{{Type: "production_all", Value: festivalBuffPercent}},
	})
	ge.festivalReadyTick = ge.tick + ge.stretchTicks(festivalCooldownTicks)
	ge.addLog("success", fmt.Sprintf("Held a cultural festival for %s: all production %s for %s.",
		Amount(cost, "culture"), textfmt.SignedPercent(festivalBuffPercent), ge.durationLocked(buff)))
	ge.logCapped(config.Effect{Type: "production_all", Value: festivalBuffPercent})
	return nil
}

// bmRandFloat returns a [0,1) float from the black-market RNG seam, or the
// seeded ge.rng if no seam is set. Lets tests force a win or a loss.
func (ge *GameEngine) bmRandFloat() float64 {
	if ge.blackMarketRand != nil {
		return ge.blackMarketRand.Float64()
	}
	return ge.gameRNG().Float64()
}

// blackMarketCost returns the culture cost of one black-market deal at the
// current progression: max(blackMarketMinCost, fraction × culture storage cap).
func (ge *GameEngine) blackMarketCost() float64 {
	cost := ge.Resources.GetStorage("culture") * blackMarketCostFraction
	if cost < blackMarketMinCost {
		cost = blackMarketMinCost
	}
	return cost
}

// blackMarketReward computes the resource payout for a winning deal on the given
// resource. The culture stake is treated as its own gold-value; the win pays
// blackMarketWinMult × that, converted into the chosen resource via its
// "<res>:gold" exchange rate (so a unit of a "more valuable" resource pays out
// fewer units). Gold pays out directly. Returns 0 if the resource has no defined
// gold valuation (and thus can't be a black-market target).
func (ge *GameEngine) blackMarketReward(resource string, stake float64) float64 {
	goldValue := stake * blackMarketWinMult
	if resource == "gold" {
		return goldValue
	}
	def, ok := ge.rules.ListedRate(resource, "gold")
	if !ok || def.BaseRate <= 0 {
		return 0
	}
	return goldValue / def.BaseRate
}

// BlackMarketStatus is the snapshot the `blackmarket` command renders.
type BlackMarketStatus struct {
	Available     bool    // true when the colonial-age gate is met
	Cost          float64 // culture required per deal right now
	Culture       float64 // player's current culture
	WinChance     float64 // probability of a payout (0..1)
	WinMult       float64 // payout multiplier on a win
	CooldownTicks int     // cooldown imposed after a deal
	CooldownLeft  int     // ticks remaining on the current cooldown (0 if ready)
	Ready         bool    // true when off cooldown AND gate met
}

// BlackMarketStatus returns the live cost, odds, and cooldown for the
// `blackmarket` command UI. Read-only.
func (ge *GameEngine) BlackMarketStatus() BlackMarketStatus {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	available := ge.progress.GetAgeOrder()[blackMarketMinAge] <= ge.progress.GetAgeOrder()[ge.age]
	cd := ge.blackMarketReadyTick - ge.tick
	if cd < 0 {
		cd = 0
	}
	return BlackMarketStatus{
		Available:     available,
		Cost:          ge.blackMarketCost(),
		Culture:       ge.Resources.Get("culture"),
		WinChance:     blackMarketWinChance,
		WinMult:       blackMarketWinMult,
		CooldownTicks: ge.stretchTicks(blackMarketCooldownTicks),
		CooldownLeft:  cd,
		Ready:         cd == 0 && available,
	}
}

// DoBlackMarket runs one high-risk/high-reward black-market deal for the given
// resource: it spends a lump of culture (always consumed) and rolls; on a win it
// pays out blackMarketWinMult × the stake's gold-value in the chosen resource, on
// a loss the smugglers vanish with the culture and nothing comes back. Gated by
// the colonial-age unlock and a cooldown. Returns (won, gain, error).
func (ge *GameEngine) DoBlackMarket(resource string) (bool, float64, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	if ageOrder[blackMarketMinAge] > ageOrder[ge.age] {
		return false, 0, fmt.Errorf("The black market opens in %s.", ge.ageSightLocked().AgeRef(blackMarketMinAge))
	}
	if ge.tick < ge.blackMarketReadyTick {
		return false, 0, fmt.Errorf("The smugglers are lying low. Try again in %s.", ge.durationLocked(ge.blackMarketReadyTick-ge.tick))
	}
	// Validate the requested payout resource is something we can value in gold.
	if _, ok := ge.rules.Resource(resource); !ok {
		return false, 0, fmt.Errorf("Unknown resource '%s'.", resource)
	}
	cost := ge.blackMarketCost()
	have := ge.Resources.Get("culture")
	if have < cost {
		return false, 0, fmt.Errorf("Not enough culture for a smuggling run: need %s, have %s.", textfmt.Number(cost), textfmt.Number(have))
	}
	reward := ge.blackMarketReward(resource, cost)
	if reward <= 0 {
		return false, 0, fmt.Errorf("Smugglers do not deal in %s. Pick another resource.", ResourceName(resource))
	}

	// Culture is spent up front regardless of outcome — that's the risk.
	ge.Resources.Remove("culture", cost)
	ge.blackMarketReadyTick = ge.tick + ge.stretchTicks(blackMarketCooldownTicks)

	if ge.bmRandFloat() < blackMarketWinChance {
		got := ge.gainResource(resource, reward)
		ge.addLog("success", fmt.Sprintf("The smuggling run paid off: %s bought %s.", Amount(cost, "culture"), Amount(got, resource)))
		return true, reward, nil
	}
	ge.addLog("warning", fmt.Sprintf("The smuggling run failed. You lost %s.", Amount(cost, "culture")))
	return false, 0, nil
}

// getWonderBonuses returns a map of bonus-type effects from all currently built
// wonders and cultural monuments (those with count > 0). Effects with Type "bonus"
// represent percentage multipliers (e.g. production_all, knowledge_rate,
// expedition_reward) rather than flat resource production rates. The returned map
// is keyed by Target and holds the summed Value across all built wonders and monuments.
// Must be called with the engine lock held.
func (ge *GameEngine) getWonderBonuses() map[string]float64 {
	out := make(map[string]float64)
	ge.Buildings.eachBuilt(func(_ string, count int, def config.BuildingDef) {
		if def.Category != "wonder" && def.Category != "monument" {
			return
		}
		for _, eff := range def.Effects {
			if eff.Type == "bonus" {
				out[eff.Target] += float64(eff.Value * float64(count))
			}
		}
	})
	return out
}

// --- Modifier emitters (Phase 2 of the multiplier-resolver refactor) ---
//
// These build the parallel []Modifier view of the same bonus sources that
// recalculateRates / recalculateTickSpeed read directly today. They are pure
// reads of already-held engine state plus pure config.* lookups, so they are
// safe to call under the engine write lock (no GetState / lock-acquiring calls).
// This phase only feeds the golden test; runtime rate math is unchanged.

// wonderModifiers emits OpAdd Modifiers from built-wonder "bonus" effects,
// attributed to Source "wonders". Mirrors getWonderBonuses (count-scaled),
// collapsed to one modifier per target.
func (ge *GameEngine) wonderModifiers() []Modifier {
	bonuses := ge.getWonderBonuses()
	out := make([]Modifier, 0, len(bonuses))
	for t, v := range bonuses {
		out = append(out, Modifier{Source: "wonders", Target: t, Op: OpAdd, Value: v})
	}
	return out
}

// permanentModifiers emits OpAdd Modifiers from ge.permanentBonuses (milestones,
// legacy, epoch-permanent — already merged into that map), attributed to Source
// "permanent". One modifier per (target, value) entry.
func (ge *GameEngine) permanentModifiers() []Modifier {
	out := make([]Modifier, 0, len(ge.permanentBonuses))
	for t, v := range ge.permanentBonuses {
		out = append(out, Modifier{Source: "permanent", Target: t, Op: OpAdd, Value: v})
	}
	return out
}

// eventModifiers emits OpAdd Modifiers from currently active events for the
// multiplier buckets the engine reads (production_all, tick_speed). Source is
// "event:<name>" per active event. Delegates to EventManager.Modifiers.
func (ge *GameEngine) eventModifiers() []Modifier {
	return ge.Events.Modifiers()
}

// moraleModifiers contributes the morale factor as a single OpMul on
// "production_all". This reproduces the engine's `rate × moraleMultiplier() ×
// (1 + Σ production_all adds)` because Resolver.Total(production_all) evaluates
// to (1 + Σ OpAdd) × Π OpMul = (1 + adds) × moraleMultiplier().
//
// The multiplier is the BANDED curve moraleMultiplier(), not the raw ge.morale
// field: recalculateRates applies `rate × moraleMultiplier()` (1.0 across the
// neutral band, bonus above, penalty below), so the modifier must emit the same
// banded factor to stay equal to the live math.
func (ge *GameEngine) moraleModifiers() []Modifier {
	return []Modifier{{Source: "morale", Target: "production_all", Op: OpMul, Value: ge.moraleMultiplier()}}
}

// diplomacyModifiers emits the allied-faction trade bonus as an OpMul on each
// affected <resource>_rate so it surfaces in the Active Multipliers panel. The
// engine still APPLIES the bonus directly in recalculateRates (the additive
// pool uses AddTotal, which ignores OpMul, so there is no double-count); this
// Modifier is the panel's view of that same GetTradeBonus value — they cannot
// drift because both read GetTradeBonus. Must read only already-held state.
func (ge *GameEngine) diplomacyModifiers() []Modifier {
	out := make([]Modifier, 0, len(ge.Resources.defs))
	for _, def := range ge.Resources.defs {
		b := ge.Diplomacy.GetTradeBonus(def.Key)
		if b != 0 {
			out = append(out, Modifier{Source: "diplomacy", Target: def.Key + "_rate", Op: OpMul, Value: 1.0 + b})
		}
	}
	return out
}

// buildResolver constructs a fresh Resolver from every bonus source and returns
// it. Pull model: a NEW resolver is built each call so nothing mutable is shared
// across goroutines. Lock safety: only call from a context that already holds the
// engine write lock (e.g. the recalc path); every source emitter reads
// already-held state or pure config.* and never re-acquires a lock.
func (ge *GameEngine) buildResolver() *Resolver {
	r := NewResolver()
	r.AddAll(ge.Research.Modifiers())
	r.AddAll(ge.Prestige.Modifiers())
	r.AddAll(ge.wonderModifiers())
	r.AddAll(ge.permanentModifiers())
	r.AddAll(ge.cosmicLegacyModifiers())
	r.AddAll(ge.eventModifiers())
	r.AddAll(ge.moraleModifiers())
	r.AddAll(ge.diplomacyModifiers())
	return r
}

// processBuildQueue advances construction on queued buildings
func (ge *GameEngine) processBuildQueue() {
	var remaining []BuildQueueItem
	for _, item := range ge.buildQueue {
		item.TicksLeft--
		if item.TicksLeft <= 0 {
			ge.finishBuild(item)
		} else {
			def := ge.Buildings.defs[item.BuildingKey]
			ge.addLog("debug", fmt.Sprintf("Build queue: %s %d/%d ticks", def.Name, item.TotalTicks-item.TicksLeft, item.TotalTicks))
			remaining = append(remaining, item)
		}
	}
	ge.buildQueue = remaining
}

// advanceBuildQueue moves construction on by n ticks at once (offline
// catch-up), completing what finishes as processBuildQueue does, without its
// per-tick progress lines. Reports whether anything completed.
func (ge *GameEngine) advanceBuildQueue(n int) bool {
	if len(ge.buildQueue) == 0 {
		return false
	}
	var remaining []BuildQueueItem
	done := false
	for _, item := range ge.buildQueue {
		item.TicksLeft -= n
		if item.TicksLeft <= 0 {
			ge.finishBuild(item)
			done = true
		} else {
			remaining = append(remaining, item)
		}
	}
	ge.buildQueue = remaining
	return done
}

// finishBuild completes one queued copy: the count, the log lines, the stats
// and the bus event, and staffing for a plan's copy.
func (ge *GameEngine) finishBuild(item BuildQueueItem) {
	key := item.BuildingKey
	ge.Buildings.counts[key]++
	if item.FromPlan {
		ge.staffPlanCopy(key)
	}
	def := ge.Buildings.defs[key]
	ge.addLog("debug", fmt.Sprintf("Build complete: %s (count now %d)", def.Name, ge.Buildings.GetCount(key)))
	done := buildDoneLog(def)
	ge.addLog(done, fmt.Sprintf("%s built (you have %d).", def.Name, ge.Buildings.GetCount(key)))
	if def.Category == "wonder" || def.Category == "monument" {
		// Their bonuses join the pools; one a cap keeps from counting says so.
		for _, capped := range ge.capLinesLocked(def.Effects, true) {
			ge.addLog("info", capped)
		}
	}
	// Cosmetic flavour — present but not stale: ~1 in 3 completions get a quip,
	// so a long build queue stays lively without turning into wallpaper. The
	// quip goes where its line goes, so the main log never shows one alone.
	if ge.quipRNG().Intn(3) == 0 {
		if q := config.PickLogFlavor(config.LogFlavorBuildingComplete, ge.quipRNG()); q != "" {
			quip := "info"
			if done == LogRoutine {
				quip = LogRoutine
			}
			ge.addLog(quip, fmt.Sprintf("  [gray]%s[-]", q))
		}
	}
	ge.Stats.RecordBuild()
	ge.Bus.Publish(EventData{
		Type:    EventBuildingBuilt,
		Payload: map[string]interface{}{"building": key},
	})
}

// --- Public API for commands ---

// AdvanceAge manually advances to the next age if requirements are met.
func (ge *GameEngine) AdvanceAge() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingCatastrophe != "" {
		return ge.catastropheBlockErr("advancing")
	}

	if !ge.ageReady {
		nextAge := ge.progress.CheckAdvancement(ge.age, ge.Resources, ge.Buildings)
		if nextAge == "" {
			// Check if wonder is the only blocker
			wonderKey := ge.progress.WonderForAge(ge.age)
			if wonderKey != "" && ge.Buildings.GetCount(wonderKey) < 1 {
				if def, ok := ge.Buildings.defs[wonderKey]; ok {
					return fmt.Errorf("Build the %s wonder before advancing: bank its cost with 'wonder collect', then type 'build %s'.", def.Name, wonderKey)
				}
			}
			return fmt.Errorf("Not ready to advance. The Next Age bar lists what is missing.")
		}
	}
	nextAge := ge.progress.GetNextAge(ge.age)
	if nextAge == "" {
		return fmt.Errorf("You are already in the final age.")
	}
	// A fated doom cannot be outrun: it strikes (or its harbinger comes)
	// before an advance that would carry the player past it.
	if err := ge.fateBeforeAdvance(nextAge); err != nil {
		return err
	}
	ge.advanceAge(nextAge)
	return nil
}

// pastMedievalForGather reports whether the given age is strictly later than the
// Medieval Age in set's age order. Used to gate hand-gathering. It is pure
// (relies only on the ruleset) and acquires no locks, so it is safe to call
// while the engine write lock is held. Fails safe: if either age key is
// absent from the order, it returns false (gathering allowed) rather than panic.
func pastMedievalForGather(set *rules.Set, age string) bool {
	cur, okCur := set.Index(age)
	medieval, okMedieval := set.Index("medieval_age")
	return okCur && okMedieval && cur > medieval
}

// GatherResource manually gathers a resource
func (ge *GameEngine) GatherResource(resource string, amount float64) (float64, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	// Hand-gathering is only practical through the Medieval Age. Past it, the
	// economy is expected to run on buildings and workers. Lock is held here, so
	// we use the ge.age field and the engine's ruleset — no GetState().
	if pastMedievalForGather(ge.rules, ge.age) {
		return 0, fmt.Errorf("Gathering by hand ends after the Medieval Age. Build producers and assign workers instead.")
	}

	if !ge.Resources.IsUnlocked(resource) {
		return 0, fmt.Errorf("You cannot gather %s yet. Gather food, wood or stone.", ResourceName(resource))
	}
	if err := checkAmount(amount); err != nil {
		return 0, err
	}
	before := ge.Resources.Get(resource)
	actual := ge.Resources.Add(resource, amount)
	ge.Stats.RecordGather(resource, amount)
	ge.addLog("debug", fmt.Sprintf("Gather: %s +%.1f (total: %.1f)", resource, amount, actual))
	// Log what storage actually took, not what was asked for.
	switch gained := actual - before; {
	case gained <= 0:
		ge.addLog("warning", fmt.Sprintf("%s storage is full. Gathered nothing.", textfmt.Capitalize(ResourceName(resource))))
	case gained < amount:
		ge.addLog(LogRoutine, fmt.Sprintf("Gathered %s (%s storage is now full).", Amount(gained, resource), ResourceName(resource)))
	default:
		ge.addLog(LogRoutine, fmt.Sprintf("Gathered %s.", Amount(gained, resource)))
	}
	return actual, nil
}

// BankWonderResource deposits amount of resource from player storage into a
// wonder's bank, at most what the bank still needs, and returns what went in.
// An amount beyond what is on hand is refused, not partly banked.
func (ge *GameEngine) BankWonderResource(wonderKey, resource string, amount float64) (float64, error) {
	if err := checkAmount(amount); err != nil {
		return 0, err
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.bankWonderLocked(wonderKey, resource, amount)
}

// BankWonderMax deposits as much of resource as the wonder's bank still
// needs, up to what is on hand (`wonder collect <res> all`), and returns what
// went in.
func (ge *GameEngine) BankWonderMax(wonderKey, resource string) (float64, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.bankWonderLocked(wonderKey, resource, ge.Resources.Get(resource))
}

func (ge *GameEngine) bankWonderLocked(wonderKey, resource string, amount float64) (float64, error) {
	deposited, err := ge.Buildings.BankResource(wonderKey, resource, amount, ge.Resources)
	if err != nil {
		return 0, err
	}

	def := ge.Buildings.defs[wonderKey]
	banked := ge.Buildings.wonderBanks[wonderKey][resource]
	need := def.BaseCost[resource]
	ge.addLog(LogRoutine, fmt.Sprintf("Banked %s toward %s (%s of %s).",
		Amount(deposited, resource), def.Name, textfmt.Number(banked), textfmt.Number(need)))

	if ge.Buildings.IsWonderBankFull(wonderKey) {
		ge.addLog("success", fmt.Sprintf("%s is fully banked. Type 'build %s' to start construction.", def.Name, wonderKey))
	}
	return deposited, nil
}

// previousAgeBuildError explains why an older age's building can't be built.
// Only point at 'upgrade' when an upgrade is actually on offer: storage never
// transforms, and some lineages have no next tier this age.
func (ge *GameEngine) previousAgeBuildError(key string, def config.BuildingDef) error {
	if _, ok := ge.Buildings.GetPendingUpgrade(key); ok && ge.Buildings.GetCount(key) > 0 {
		return fmt.Errorf("%s belongs to a previous age. Type 'upgrade %s' to upgrade the ones you have.", def.Name, key)
	}
	if def.Category == "storage" {
		return fmt.Errorf("%s belongs to a previous age and can no longer be built. The ones you have still count, so build this age's storage instead.", def.Name)
	}
	return fmt.Errorf("%s belongs to a previous age and can no longer be built.", def.Name)
}

// BuildBuilding constructs a building (instant or queued)
func (ge *GameEngine) BuildBuilding(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.startBuildLocked(key, false)
}

// startBuildLocked is BuildBuilding under the write lock: every check the
// command makes, the payment, and the queueing (or instant build). The build
// plan starts its builds through it too, with quiet set so the per-copy
// "Started building" line gives way to the plan's own summary.
func (ge *GameEngine) startBuildLocked(key string, quiet bool) error {
	return ge.startBuildPaid(key, quiet, nil)
}

// startBuildPaid is startBuildLocked with part of the price already paid:
// prepaid (a plan item's overflow bank, at most the price per resource) is
// taken off what the stores pay. The caller takes it out of the bank once
// the build has started. Parts of the price left within planBankEpsilon of
// prepaid count as paid.
func (ge *GameEngine) startBuildPaid(key string, quiet bool, prepaid map[string]float64) error {
	def, exists := ge.Buildings.defs[key]
	if !exists {
		return ge.unknownBuildingErr(key)
	}
	if err := ge.techLockErr(def); err != nil && ge.Buildings.AgeUnlocked(key) {
		return err
	}
	if !ge.Buildings.IsUnlocked(key) {
		return fmt.Errorf("%s is not unlocked yet.", def.Name)
	}
	// Age lock: only allow building structures that belong to the current age.
	// Storage and wonder categories with no RequiredAge (empty string) are exempt.
	// Wonders always match because their RequiredAge equals the age they unlock in,
	// and the player can only be in that age when they attempt to build the wonder.
	if def.RequiredAge != "" && def.RequiredAge != ge.age {
		return ge.previousAgeBuildError(key, def)
	}
	// A unique building under construction says so. Only unique ones: a
	// capped building like storage queues copies up to its MaxCount, as
	// BuildMultiple always allowed. (This check used to cover every capped
	// building, so a player who checks in a few times a day could queue one
	// storage copy per visit with `build`, but any number with `build <key> N`.)
	if def.MaxCount == 1 {
		for _, item := range ge.buildQueue {
			if item.BuildingKey == key {
				return fmt.Errorf("%s is already under construction (%s left).", def.Name, ge.durationLocked(item.TicksLeft))
			}
		}
	}
	if def.MaxCount > 0 {
		inQueue := ge.Buildings.GetQueueCount(key, ge.buildQueue)
		if ge.Buildings.GetCount(key)+inQueue >= def.MaxCount {
			return fmt.Errorf("%s is at its max count of %d.", def.Name, def.MaxCount)
		}
	}

	if DevGodMode {
		// godmode: skip all cost/bank checks, build instantly below
	} else if def.Category == "wonder" {
		if !ge.Buildings.IsWonderBankFull(key) {
			return fmt.Errorf("%s is not fully banked yet. Bank its cost first with 'wonder collect <resource|all> [amount|all]'.", def.Name)
		}
		// Resources were already deducted when banked; nothing to pay here.
	} else {
		// Use queue-aware cost so that items already in the build queue are
		// factored into the cost curve (fixes queue-blindness exploit).
		cost, _ := ge.Buildings.BuildBatchCost(key, 1, ge.buildQueue)
		if len(prepaid) > 0 {
			cost, _ = splitBank(cost, prepaid)
		}
		if !ge.Resources.Pay(cost) {
			return fmt.Errorf("Cannot afford %s: need %s.", def.Name, ge.shortfallText(cost))
		}
	}

	ge.addLog("debug", fmt.Sprintf("Build start: %s", def.Name))
	if !DevGodMode && def.BuildTicks > 0 {
		// Queue for construction, ÷ k on known ground (Era Mastery).
		ticks := ge.buildTicksLocked(def)
		ge.buildQueue = append(ge.buildQueue, BuildQueueItem{
			BuildingKey: key,
			TicksLeft:   ticks,
			TotalTicks:  ticks,
			FromPlan:    quiet,
		})
		if !quiet {
			ge.addLog(LogRoutine, fmt.Sprintf("Started building %s (%s).", def.Name, ge.durationLocked(ticks)))
		}
	} else {
		// Instant build
		ge.Buildings.counts[key]++
		ge.Stats.RecordBuild()
		if quiet {
			ge.staffPlanCopy(key)
		}
		ge.recalculateRates()
		ge.addLog(buildDoneLog(def), fmt.Sprintf("%s built (you have %d).", def.Name, ge.Buildings.GetCount(key)))
		ge.Bus.Publish(EventData{
			Type:    EventBuildingBuilt,
			Payload: map[string]interface{}{"building": key},
		})
	}
	return nil
}

// BuildMultiple constructs up to count buildings, stopping when resources run out or max is hit.
// Returns the number actually built.
// Each successive unit is priced using the cumulative cost curve:
//
//	cost_i = floor(baseCost × scale^(built + queued + i))
//
// where built = fully-constructed count and queued = items already in the build
// queue for this key. This prevents batch purchases and the `max` command from
// bypassing cost scaling.
func (ge *GameEngine) BuildMultiple(key string, count int) (int, error) {
	if count <= 0 {
		return 0, fmt.Errorf("Build count must be positive (got %d).", count)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, exists := ge.Buildings.defs[key]
	if !exists {
		return 0, ge.unknownBuildingErr(key)
	}
	if err := ge.techLockErr(def); err != nil && ge.Buildings.AgeUnlocked(key) {
		return 0, err
	}
	if !ge.Buildings.IsUnlocked(key) {
		return 0, fmt.Errorf("%s is not unlocked yet.", def.Name)
	}
	// Age lock: only allow building structures that belong to the current age.
	if def.RequiredAge != "" && def.RequiredAge != ge.age {
		return 0, ge.previousAgeBuildError(key, def)
	}

	built := 0
	for i := 0; i < count; i++ {
		// Check MaxCount against fully-built + queued + what we're about to add
		if def.MaxCount > 0 {
			inQueue := ge.Buildings.GetQueueCount(key, ge.buildQueue)
			if ge.Buildings.GetCount(key)+inQueue >= def.MaxCount {
				break
			}
		}

		// Cost for this specific unit accounts for already-built and already-queued
		// instances so the exponential curve is not bypassed by batch purchases.
		unitCost, ok := ge.Buildings.BuildBatchCost(key, 1, ge.buildQueue)
		if !ok {
			break
		}
		if !ge.Resources.Pay(unitCost) {
			break
		}

		if def.BuildTicks > 0 {
			ticks := ge.buildTicksLocked(def)
			ge.buildQueue = append(ge.buildQueue, BuildQueueItem{
				BuildingKey: key,
				TicksLeft:   ticks,
				TotalTicks:  ticks,
			})
		} else {
			ge.Buildings.counts[key]++
			ge.Stats.RecordBuild()
			ge.Bus.Publish(EventData{
				Type:    EventBuildingBuilt,
				Payload: map[string]interface{}{"building": key},
			})
		}
		built++
	}

	if built == 0 {
		if def.MaxCount > 0 {
			inQueue := ge.Buildings.GetQueueCount(key, ge.buildQueue)
			if ge.Buildings.GetCount(key)+inQueue >= def.MaxCount {
				return 0, fmt.Errorf("%s is at its max count of %d.", def.Name, def.MaxCount)
			}
		}
		unitCost, _ := ge.Buildings.BuildBatchCost(key, 1, ge.buildQueue)
		return 0, fmt.Errorf("Cannot afford %s: need %s.", def.Name, ge.shortfallText(unitCost))
	}

	if def.BuildTicks > 0 {
		ge.addLog(LogRoutine, fmt.Sprintf("Queued %s.", buildingCountIn(ge.rules, built, key)))
	} else {
		ge.recalculateRates()
		ge.addLog(buildDoneLog(def), fmt.Sprintf("Built %s (you have %d).", buildingCountIn(ge.rules, built, key), ge.Buildings.GetCount(key)))
	}
	return built, nil
}

// RecruitMax recruits as many workers as possible up to the pop cap
func (ge *GameEngine) RecruitMax(vType string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if !ge.Workers.IsUnlocked(vType) {
		return 0, fmt.Errorf("Worker type '%s' is not unlocked yet.", vType)
	}

	popCap := ge.popCapLocked()

	available := popCap - ge.Workers.TotalPop()
	if available <= 0 {
		return 0, fmt.Errorf("No housing left (%d/%d). Build housing to recruit more.", ge.Workers.TotalPop(), popCap)
	}

	if !ge.Workers.Recruit(vType, available, popCap) {
		return 0, fmt.Errorf("Could not recruit any workers.")
	}
	ge.Stats.RecordRecruit(available)
	ge.holdStaffing()
	ge.addLog(LogRoutine, fmt.Sprintf("Recruited %s (population %d/%d).",
		textfmt.Count(available, "worker", "workers"), ge.Workers.TotalPop(), popCap))
	return available, nil
}

// RecruitWorker recruits workers
func (ge *GameEngine) RecruitWorker(vType string, count int) error {
	if count <= 0 {
		return fmt.Errorf("Recruit count must be positive (got %d).", count)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()

	popCap := ge.popCapLocked()

	if !ge.Workers.Recruit(vType, count, popCap) {
		totalPop := ge.Workers.TotalPop()
		if !ge.Workers.IsUnlocked(vType) {
			return fmt.Errorf("Worker type '%s' is not unlocked yet.", vType)
		}
		if totalPop >= popCap {
			return fmt.Errorf("No housing left (%d/%d). Build housing to recruit more.", totalPop, popCap)
		}
		return fmt.Errorf("Not enough housing for %s (%d/%d). Build housing to recruit more.",
			textfmt.Count(count, "more worker", "more workers"), totalPop, popCap)
	}
	ge.Stats.RecordRecruit(count)
	ge.holdStaffing()
	ge.addLog("debug", fmt.Sprintf("Recruit: %d (pop: %d/%d)", count, ge.Workers.TotalPop(), popCap))
	ge.addLog(LogRoutine, fmt.Sprintf("Recruited %s (population %d/%d).",
		textfmt.Count(count, "worker", "workers"), ge.Workers.TotalPop(), popCap))
	return nil
}

// AssignWorker assigns workers to a building.
// Any worker can be assigned to any building with WorkerCapacity > 0.
func (ge *GameEngine) AssignWorker(buildingKey string, count int) error {
	if count <= 0 {
		return fmt.Errorf("Assign count must be positive (got %d).", count)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, err := ge.staffableBuilding(buildingKey)
	if err != nil {
		return err
	}
	// Enforce capacity cap
	totalCap := def.WorkerCapacity * ge.Buildings.GetCount(buildingKey)
	alreadyAssigned := ge.Workers.GetAssignedCount("worker", buildingKey)
	available := totalCap - alreadyAssigned
	if available <= 0 {
		return fmt.Errorf("All %s at %s are full.", textfmt.Count(totalCap, "worker slot", "worker slots"), def.Name)
	}
	if count > available {
		return fmt.Errorf("Only %s free at %s (%d/%d filled).",
			textfmt.Count(available, "worker slot is", "worker slots are"), def.Name, alreadyAssigned, totalCap)
	}
	if !ge.Workers.Assign("worker", buildingKey, count) {
		idle := ge.Workers.IdleCount("worker")
		return fmt.Errorf("Only %s idle (you asked for %d). Recruit more or unassign some elsewhere.",
			textfmt.Count(idle, "worker is", "workers are"), count)
	}
	ge.holdStaffing()
	ge.recalculateRates()
	ge.addLog("debug", fmt.Sprintf("Assign: %d → %s", count, buildingKey))
	ge.addLog(LogRoutine, fmt.Sprintf("Assigned %s to %s.", textfmt.Count(count, "worker", "workers"), def.Name))
	return nil
}

// AssignAll assigns all idle workers to a building.
// Any worker can be assigned to any building with WorkerCapacity > 0.
func (ge *GameEngine) AssignAll(buildingKey string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, err := ge.staffableBuilding(buildingKey)
	if err != nil {
		return 0, err
	}
	// Cap at available capacity
	toAssign := ge.Workers.IdleCount("worker")
	if toAssign <= 0 {
		return 0, fmt.Errorf("No idle workers to assign. Recruit more first.")
	}
	totalCap := def.WorkerCapacity * ge.Buildings.GetCount(buildingKey)
	alreadyAssigned := ge.Workers.GetAssignedCount("worker", buildingKey)
	available := totalCap - alreadyAssigned
	if available <= 0 {
		return 0, fmt.Errorf("All %s at %s are full.", textfmt.Count(totalCap, "worker slot", "worker slots"), def.Name)
	}
	if toAssign > available {
		toAssign = available
	}
	if !ge.Workers.Assign("worker", buildingKey, toAssign) {
		return 0, fmt.Errorf("Could not assign workers to %s.", def.Name)
	}
	ge.holdStaffing()
	ge.recalculateRates()
	ge.addLog(LogRoutine, fmt.Sprintf("Assigned %s to %s.", textfmt.Count(toAssign, "worker", "workers"), def.Name))
	return toAssign, nil
}

// UnassignAll removes all workers from a building.
func (ge *GameEngine) UnassignAll(buildingKey string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, err := ge.workerBuilding(buildingKey)
	if err != nil {
		return 0, err
	}
	assigned := ge.Workers.GetAssignedCount("worker", buildingKey)
	if assigned <= 0 {
		return 0, fmt.Errorf("No workers are assigned to %s.", def.Name)
	}
	if !ge.Workers.Unassign("worker", buildingKey, assigned) {
		return 0, fmt.Errorf("Could not unassign workers from %s.", def.Name)
	}
	ge.holdStaffing()
	ge.recalculateRates()
	ge.addLog(LogRoutine, fmt.Sprintf("Unassigned %s from %s.", textfmt.Count(assigned, "worker", "workers"), def.Name))
	return assigned, nil
}

// UnassignWorker removes a specific number of workers from a building.
func (ge *GameEngine) UnassignWorker(buildingKey string, count int) error {
	if count <= 0 {
		return fmt.Errorf("Unassign count must be positive (got %d).", count)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, err := ge.workerBuilding(buildingKey)
	if err != nil {
		return err
	}
	if !ge.Workers.Unassign("worker", buildingKey, count) {
		assigned := ge.Workers.GetAssignedCount("worker", buildingKey)
		if assigned <= 0 {
			return fmt.Errorf("No workers are assigned to %s.", def.Name)
		}
		return fmt.Errorf("Only %s assigned to %s.", textfmt.Count(assigned, "worker is", "workers are"), def.Name)
	}
	ge.holdStaffing()
	ge.recalculateRates()
	ge.addLog("debug", fmt.Sprintf("Unassign: %d ← %s", count, buildingKey))
	ge.addLog(LogRoutine, fmt.Sprintf("Unassigned %s from %s.", textfmt.Count(count, "worker", "workers"), def.Name))
	return nil
}

// DismissWorkers removes workers from a building and from the population pool entirely.
func (ge *GameEngine) DismissWorkers(buildingKey string, count int, all bool) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, ok := ge.Buildings.defs[buildingKey]
	if !ok {
		return ge.unknownBuildingErr(buildingKey)
	}
	if all {
		count = ge.Workers.GetAssignedCount("worker", buildingKey)
	}
	if count <= 0 {
		return fmt.Errorf("No workers are assigned to %s.", def.Name)
	}
	dismissed := ge.Workers.Dismiss(buildingKey, count)
	if dismissed == 0 {
		return fmt.Errorf("No workers are assigned to %s.", def.Name)
	}
	ge.holdStaffing()
	ge.recalculateRates()
	ge.addLog(LogRoutine, fmt.Sprintf("Dismissed %s from %s. They left your population (now %d).",
		textfmt.Count(dismissed, "worker", "workers"), def.Name, ge.Workers.TotalPop()))
	return nil
}

// checkAmount rejects a resource amount no command can mean: NaN, an
// infinity, zero or a negative. NaN is the dangerous one: every comparison
// with it is false, so it slips past "have < need" checks and poisons
// whatever it is added to or subtracted from.
func checkAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return fmt.Errorf("Amount must be a positive number (got %v).", amount)
	}
	return nil
}

// unknownBuildingErr refuses a key that names no building, suggesting the
// closest one: "Unknown building 'farmm'. Did you mean 'farm'?"
func (ge *GameEngine) unknownBuildingErr(key string) error {
	if suggestion := ge.Buildings.SuggestKey(key); suggestion != "" {
		return fmt.Errorf("Unknown building '%s'. Did you mean '%s'?", key, suggestion)
	}
	return fmt.Errorf("Unknown building '%s'. Type 'build' to see available buildings.", key)
}

// workerBuilding returns the definition of a building that takes workers, or
// the refusal to show when the key is unknown or the building has no worker
// slots. Caller holds ge.mu.
func (ge *GameEngine) workerBuilding(key string) (config.BuildingDef, error) {
	def, ok := ge.Buildings.defs[key]
	if !ok {
		return def, ge.unknownBuildingErr(key)
	}
	if def.WorkerCapacity == 0 {
		return def, fmt.Errorf("%s does not take workers.", def.Name)
	}
	return def, nil
}

// staffableBuilding is workerBuilding that also requires at least one copy
// built, for assigning. Caller holds ge.mu.
func (ge *GameEngine) staffableBuilding(key string) (config.BuildingDef, error) {
	def, err := ge.workerBuilding(key)
	if err != nil {
		return def, err
	}
	if ge.Buildings.GetCount(key) == 0 {
		return def, fmt.Errorf("You have no %s yet. Build one first.", def.Name)
	}
	return def, nil
}

// shortfallText names what a cost is short of: "10 gold (have 5)" for each
// resource the player lacks, or the whole cost when nothing is short.
// Caller holds ge.mu.
func (ge *GameEngine) shortfallText(cost map[string]float64) string {
	var short []string
	for _, res := range sortedKeys(cost) {
		if have := ge.Resources.Get(res); have < cost[res] {
			short = append(short, fmt.Sprintf("%s (have %s)", Amount(cost[res], res), textfmt.Number(have)))
		}
	}
	if len(short) == 0 {
		return Amounts(cost)
	}
	return textfmt.List(short)
}

// SellBuilding removes n copies of a built building, refunds 50% of cost,
// and unassigns any workers that were in the sold slots.
func (ge *GameEngine) SellBuilding(key string, n int) error {
	if n <= 0 {
		return fmt.Errorf("Sell count must be positive (got %d).", n)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.age == "primitive_age" {
		return fmt.Errorf("You cannot sell buildings in the %s.", ge.rules.Name(rules.KindAge, "primitive_age"))
	}

	def, ok := ge.rules.Building(key)
	if !ok {
		return ge.unknownBuildingErr(key)
	}

	if def.Category == "wonder" {
		return fmt.Errorf("Wonders cannot be sold.")
	}
	// Storage is permanent, like wonders (see isDestroyable). Once its age has
	// passed the age lock never lets it be rebuilt, and an age's first storage
	// copy can cost more than the storage left after a sale, so a sold copy
	// (even this age's, sold just before advancing) could leave no way to
	// raise a cap ever again.
	if def.Category == "storage" {
		return fmt.Errorf("Storage cannot be sold: once its age has passed, it can never be rebuilt.")
	}

	current := ge.Buildings.GetCount(key)
	if current == 0 {
		return fmt.Errorf("You have no %s to sell.", def.Name)
	}

	if n > current {
		n = current
	}

	// Check build queue — reject if any copy of this building is queued
	for _, item := range ge.buildQueue {
		if item.BuildingKey == key {
			return fmt.Errorf("A %s is still under construction. Sell after it is finished.", def.Name)
		}
	}

	// Compute refund before removing
	refund, _ := ge.Buildings.SellCost(key, n)

	// Remove the buildings
	ge.Buildings.RemoveBuilding(key, n)

	// Unassign excess workers
	if def.WorkerCapacity > 0 {
		newCount := current - n
		newCap := def.WorkerCapacity * newCount
		assigned := ge.Workers.GetAssignedCount("worker", key)
		if newCap == 0 {
			ge.Workers.Unassign("worker", key, assigned)
		} else if assigned > newCap {
			ge.Workers.Unassign("worker", key, assigned-newCap)
		}
	}

	// Add refund to resources, keeping what storage actually took.
	got := make(map[string]float64, len(refund))
	clipped := false
	for _, res := range sortedKeys(refund) {
		got[res] = ge.gainResource(res, refund[res])
		if got[res] < refund[res] {
			clipped = true
		}
	}

	ge.holdStaffing()
	ge.recalculateRates()
	line := fmt.Sprintf("Sold %s. Refund: %s.", buildingCountIn(ge.rules, n, key), Amounts(got))
	if clipped {
		line += " Storage was full, so part of the refund was lost."
	}
	ge.addLog(LogRoutine, line)
	return nil
}

// StartResearch begins researching a technology
func (ge *GameEngine) StartResearch(techKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.startResearchLocked(techKey, false)
}

// startResearchLocked is StartResearch under the write lock; the build plan
// starts its techs through it with quiet set (its own summary line replaces
// "Started researching").
func (ge *GameEngine) startResearchLocked(techKey string, quiet bool) error {
	ageOrder := ge.progress.GetAgeOrder()
	knowledge := ge.Resources.Get("knowledge")

	// Combine research_speed from all sources (see combinedResearchSpeed). This
	// must be done before StartResearch so the combined value reduces tick count.
	combinedResearchSpeed := ge.combinedResearchSpeed()
	ge.Research.timeMult = ge.succumbResearchFactor() // Ancient Knowledge: × 0.8 per epoch
	ge.Research.timeK = ge.speedK()                   // Era Mastery: ÷ k after the speed step
	if err := ge.Research.StartResearchWithSpeed(techKey, ge.age, ageOrder, knowledge, combinedResearchSpeed); err != nil {
		return err
	}

	// Pay knowledge cost (waived in godmode)
	def, _ := ge.rules.Tech(techKey)
	if !DevGodMode {
		ge.Resources.Remove("knowledge", def.Cost)
	}
	if DevGodMode {
		// complete immediately
		ge.Research.ticksLeft = 0
	}
	ge.addLog("debug", fmt.Sprintf("Research start: %s (cost: %.0f knowledge, %d ticks)", def.Name, def.Cost, ge.Research.totalTicks))
	if !quiet {
		ge.addLog(LogRoutine, fmt.Sprintf("Started researching %s (%s).", def.Name, ge.durationLocked(ge.Research.totalTicks)))
	}
	return nil
}

// CancelResearch cancels current research (no refund)
func (ge *GameEngine) CancelResearch() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	tech, ok := ge.Research.CancelResearch()
	if !ok {
		return fmt.Errorf("No research in progress.")
	}
	ge.addLog("warning", fmt.Sprintf("Research on %s canceled (no refund).", ge.rules.Name(rules.KindTech, tech)))
	return nil
}

// ===== Ancient Civilization Memory (Trello yn98pTQw) =====

const (
	// ancientMemoryChance is the probability the cache appears at the start of a
	// qualifying new prestige run. Occasional by design — not every run.
	ancientMemoryChance = 0.40
	// ancientMemoryFlavor is shown in the offer modal and the log.
	ancientMemoryFlavor = "You have discovered an old cache. It appears to contain memories of a now-extinct civilization."
)

// ancientMemoryAges is the set of (early) ages in which a cache can surface. The
// design fires it early in a fresh run, before the player has rebuilt past it.
var ancientMemoryAges = map[string]bool{
	"primitive_age": true,
	"stone_age":     true,
}

// memRandFloat returns a [0,1) float from the injected seam, or the seeded
// ge.rng if no seam is set. Lets tests force/suppress the roll.
func (ge *GameEngine) memRandFloat() float64 {
	if ge.memoryRand != nil {
		return ge.memoryRand.Float64()
	}
	return ge.gameRNG().Float64()
}

// memRandIntn returns a non-negative int in [0,n) from the injected seam (or the
// seeded ge.rng). n must be > 0.
func (ge *GameEngine) memRandIntn(n int) int {
	if ge.memoryRand != nil {
		return ge.memoryRand.Intn(n)
	}
	return ge.gameRNG().Intn(n)
}

// maybeOfferAncientMemory rolls for, and possibly offers, an Ancient Memory cache.
// MUST be called with ge.mu held (it is invoked from the tail of DoPrestige and
// Succumb, which already hold the write lock) — it does not lock and must not call
// any lock-acquiring method.
//
// Gating (all must hold):
//   - prestige level >= 1 — there is no "previous civilization" on the first-ever
//     run, so the very first run never offers a cache.
//   - current age is primitive or stone (early in the run).
//   - this run has not already used its memory (ancientMemoryUsed false).
//   - the probability roll succeeds.
//
// On success it picks a candidate tech and sets pendingMemoryTech (the UI pops the
// accept/decline modal) AND marks ancientMemoryUsed — set on OFFER, so declining
// still spends the run's single chance and a save/reload can't re-roll it.
func (ge *GameEngine) maybeOfferAncientMemory() {
	if ge.ancientMemoryUsed {
		return
	}
	if ge.Prestige.GetLevel() < 1 {
		return // first-ever run: no extinct civilization to remember
	}
	if !ancientMemoryAges[ge.age] {
		return
	}
	if ge.memRandFloat() >= ancientMemoryChance {
		return // the cache stays buried this run
	}

	ageOrder := ge.progress.GetAgeOrder()
	techKey := ge.selectMemoryTech(ge.age, ageOrder, ge.Prestige.GetLevel())
	if techKey == "" {
		return // nothing valid to offer (e.g. everything already researched)
	}

	// Consume the run's chance on offer (no save-scum re-rolls), then present it.
	ge.ancientMemoryUsed = true
	ge.pendingMemoryTech = techKey
	def, _ := ge.rules.Tech(techKey)
	ge.addLog("event", fmt.Sprintf("✦ %s A memory of [cyan]%s[-] stirs. You can research it without its prerequisites, at half speed.", ancientMemoryFlavor, def.Name))
}

// selectMemoryTech picks a random tech appropriate to the current age, with the
// reachable tier gated by prestige level: low prestige offers a near-current-age
// tech; higher prestige can reach a higher age's tech (one extra age of reach per
// two prestige levels). Returns "" if no eligible, unresearched tech exists.
//
// Pure aside from the RNG seam — safe to call under ge.mu.
func (ge *GameEngine) selectMemoryTech(currentAge string, ageOrder map[string]int, prestigeLevel int) string {
	currentOrder, ok := ageOrder[currentAge]
	if !ok {
		return ""
	}
	// Reach: current age plus one age per two prestige levels.
	maxOrder := currentOrder + prestigeLevel/2

	var candidates []string
	for _, t := range ge.rules.Techs() {
		o, ok := ageOrder[t.Age]
		if !ok {
			continue
		}
		if o < currentOrder || o > maxOrder {
			continue
		}
		if ge.Research.IsResearched(t.Key) {
			continue
		}
		if t.Key == ge.Research.currentTech {
			continue // don't offer what's already in progress
		}
		candidates = append(candidates, t.Key)
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[ge.memRandIntn(len(candidates))]
}

// AcceptAncientMemory accepts the pending Ancient Memory offer: it begins
// researching the offered tech, bypassing prerequisites/age/cost, at 50% speed.
// Called from the UI (modal callback) on the UI goroutine — takes the write lock.
func (ge *GameEngine) AcceptAncientMemory() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingMemoryTech == "" {
		return fmt.Errorf("There is no ancient memory to accept.")
	}
	techKey := ge.pendingMemoryTech
	ge.pendingMemoryTech = ""

	// Same combined research_speed sources a normal research gets; the memory
	// penalty (2x ticks) is applied on top inside StartMemoryResearch.
	combinedResearchSpeed := ge.combinedResearchSpeed()
	ge.Research.timeMult = ge.succumbResearchFactor() // Ancient Knowledge: × 0.8 per epoch
	ge.Research.timeK = ge.speedK()                   // Era Mastery: ÷ k after the speed step
	if err := ge.Research.StartMemoryResearch(techKey, combinedResearchSpeed); err != nil {
		return err
	}
	def, _ := ge.rules.Tech(techKey)
	ge.addLog("success", fmt.Sprintf("Recovered the memory of %s. Researching it at half speed (%s).", def.Name, ge.durationLocked(ge.Research.totalTicks)))
	return nil
}

// DeclineAncientMemory dismisses the pending offer without effect. The cache is
// already consumed for this run (ancientMemoryUsed was set on offer), so declining
// does not refund the chance. Called from the UI (modal callback).
func (ge *GameEngine) DeclineAncientMemory() {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingMemoryTech == "" {
		return
	}
	ge.pendingMemoryTech = ""
	ge.addLog("info", "You leave the ancient cache sealed. Its memories crumble to dust.")
}

// LaunchExpedition starts a military expedition
func (ge *GameEngine) LaunchExpedition(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.launchExpeditionLocked(key)
}

// launchExpeditionLocked is the single expedition-launch path: it validates the
// cost, the age range and the one-active-per-category rule, then deducts soldiers
// and resources. Callers must already hold the write lock.
//
// Both entry points go through here — the player's `expedition` command
// (LaunchExpedition) and the Geographic Society's automatic dispatch (see
// auto_expedition.go). An auto-launch is charged and validated identically to a
// hand-dispatched one; nothing about the resulting expedition is special-cased,
// so it resolves, rewards and rolls its faction encounter on exactly the same
// path.
func (ge *GameEngine) launchExpeditionLocked(key string) error {
	ageOrder := ge.progress.GetAgeOrder()

	def := ge.Military.ExpeditionDefByKey(key)
	if def == nil {
		return fmt.Errorf("unknown expedition or campaign '%s'. Type expedition or campaign to see what you can send", key)
	}

	// --- Validate everything BEFORE any deduction so a failed launch never
	// partially charges the player. ---

	// Soldiers resource check (soldiers are now a real resource, not workers).
	haveSoldiers := int(ge.Resources.Get("soldiers"))
	if haveSoldiers < def.SoldiersNeeded {
		return fmt.Errorf("%s needs %d soldiers (you have %d). Military buildings train them; see the Army panel", def.Name, def.SoldiersNeeded, haveSoldiers)
	}

	// Additional resource cost check.
	for res, amount := range def.Cost {
		if ge.Resources.Get(res) < amount {
			return fmt.Errorf("%s needs %s %s (you have %s)", def.Name, amountText(amount), resourceLabel(res), amountText(ge.Resources.Get(res)))
		}
	}

	// Age range + active-expedition validation (does NOT touch resources).
	if err := ge.Military.LaunchExpedition(ge.gameRNG(), key, ge.age, ageOrder); err != nil {
		return err
	}

	// --- All checks passed: deduct soldiers + Cost. ---
	if def.SoldiersNeeded > 0 {
		ge.Resources.Remove("soldiers", float64(def.SoldiersNeeded))
	}
	for res, amount := range def.Cost {
		ge.Resources.Remove(res, amount)
	}

	ge.addLog("debug", fmt.Sprintf("Expedition start: %s (soldiers spent: %d)", def.Name, def.SoldiersNeeded))
	if def.Category == ExpeditionMilitary {
		ge.addLog(LogRoutine, fmt.Sprintf("Campaign launched: %s (%d soldiers).", def.Name, def.SoldiersNeeded))
	} else {
		ge.addLog(LogRoutine, fmt.Sprintf("Expedition sent: %s.", def.Name))
	}
	return nil
}

// DoPrestige resets the game with prestige bonuses
func (ge *GameEngine) DoPrestige() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingCatastrophe != "" {
		return ge.catastropheBlockErr("prestiging")
	}
	if ge.pendingLastPassage {
		return lastPassageBlockErr(ge.rules)
	}

	ageOrder := ge.progress.GetAgeOrder()
	if !ge.Prestige.CanPrestige(ge.age, ageOrder) {
		return fmt.Errorf("You can prestige once you reach %s.", ge.ageSightLocked().AgeRef(PrestigeMinAge))
	}

	// In the final epoch prestige is the passage, and it can bring the Last
	// Passage (last_passage.go). If it comes, prestige waits for the choice.
	// The era's fated doom cannot be outrun past it: it settles first.
	if ge.lastPassageApplies() {
		if err := ge.fateBeforePrestige(); err != nil {
			return err
		}
		if ge.rollLastPassage() {
			return nil
		}
		ge.completePrestige(lastPassageSpared)
		return nil
	}
	ge.completePrestige(prestigePlain)
	return nil
}

// completePrestige awards the run's prestige points (all, part or none of
// them, as how says), logs the run's last lines and resets for the new run.
// Called by DoPrestige and by the Last Passage choice, under the write lock.
func (ge *GameEngine) completePrestige(how prestigeEnding) {
	// What the player may see named, before the reset takes the run's ages.
	sight := ge.ageSightLocked()
	full := ge.Prestige.CalculatePoints(ge.age)
	points := ge.lastPassagePoints(how, full)
	ge.recordLastPassageOutcome(how, points, full)
	// The verdict and the ending belong to the old run; the reset below clears
	// the log, so they are written aside and carried over.
	carried := ge.runEndingLines(how, points, full)
	newLegacy := how == lastPassageSuccumbed && !ge.cosmicLegacy
	if how == lastPassageSuccumbed {
		ge.cosmicLegacy = true
	}

	ge.Prestige.Prestige(points)
	// Era Mastery: every age this run completed gains a level. The age
	// prestiged from counts as entered even if a test hook or the dev
	// console moved the run there without an advance.
	ge.Prestige.NoteAgeEntered(ge.age)
	masteryLine := masteryCommitLine(ge.rules, ge.Prestige.CommitRun())
	// The legacy kit remembers the run (plan, civilizations, shares) before
	// the managers holding it are reset.
	ge.captureLegacyLocked()
	prestigedFrom := ge.age

	// Preserve cross-run state before resetting managers
	savedRuins := ge.Buildings.GetAllRuins()

	// Reset all game systems
	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManagerWith(ge.rules)
	ge.Buildings = ge.newBuildingManager()
	ge.Workers = NewWorkerManagerWith(ge.rules)
	ge.Research = NewResearchManagerWith(ge.rules)
	ge.Military = NewMilitaryManagerWith(ge.rules)
	ge.Events = NewEventManagerWith(ge.rules)
	ge.Milestones = NewMilestoneManagerWith(ge.rules)
	ge.Trade = NewTradeManagerWith(ge.rules)
	ge.Diplomacy = NewDiplomacyManagerWith(ge.rules)
	ge.Stats = NewGameStats()
	// Bus intentionally kept — dashboard subscriptions must survive across resets.
	ge.permanentBonuses = make(map[string]float64)
	// A new run starts at 1x, as after a Succumb or a wipe: the dev console's
	// /speed override does not carry into it.
	ge.speedMultiplier = 1.0
	ge.buildQueue = nil
	ge.plan = nil
	ge.log = nil
	ge.currentEpoch = ge.rules.EraOf("primitive_age")
	ge.epochEventFired = make(map[string]bool)
	ge.clearHarbingerRun()
	ge.harbingerHistory = nil
	ge.awakeningsFired = make(map[string]bool)
	ge.survivedEpochs = make(map[string]bool)
	ge.pendingCatastrophe = ""
	ge.epochEventHistory = nil
	ge.morale = 0.70
	ge.lowMoraleWarned = false
	// Fresh run: this prestige cycle may roll a new Ancient Memory.
	ge.ancientMemoryUsed = false
	ge.pendingMemoryTech = ""
	// The cooldowns are tick numbers and the tick counter just went back to 0.
	ge.festivalReadyTick, ge.blackMarketReadyTick = 0, 0
	// The run's timers start over, as in a new game. A stale ageReady let
	// `advance` skip the Stone Age's requirements before the first tick.
	ge.ageReady = false
	ge.starvationTicks = 0
	ge.autoExpeditionTicksLeft = 0
	ge.autoExpeditionStarved = false
	ge.startRunShares()

	// Restore cross-run state
	ge.Buildings.LoadRuins(savedRuins)
	ge.reapplyLegacyBonuses()

	// Apply age unlocks for primitive age
	ge.applyAgeUnlocks("primitive_age")

	// Recompute derived caps and rates for the fresh managers now, not on the
	// first tick: the storage upgrade has to be in the first snapshot of the
	// new run, and in place before the starting resources land, or they are
	// clamped to the base storage it raises.
	ge.recalculateRates()

	// Apply starting resources (base + prestige bonus)
	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)
	for res, amount := range ge.Prestige.GetStartingResources() {
		ge.Resources.Add(res, amount)
	}

	ge.recalculateTickSpeed()

	ge.log = carried
	ge.addLog("success", fmt.Sprintf("Prestige complete. Level %d, %s earned.",
		ge.Prestige.GetLevel(), textfmt.Count(points, "prestige point", "prestige points")))
	// An early prestige pays little: say so, and what a deeper run pays.
	if line := EarlyPrestigeLine(sight, prestigedFrom, points, true); line != "" {
		ge.addLog("info", line)
	}
	if newLegacy {
		ge.addLog("success", fmt.Sprintf("✦ Cosmic Legacy: all production %s, permanent. It survives every prestige and every fall.", textfmt.SignedPercent(CosmicLegacyProductionBonus)))
	} else if ge.cosmicLegacy {
		ge.addLog("info", fmt.Sprintf("Cosmic Legacy active: all production %s.", textfmt.SignedPercent(CosmicLegacyProductionBonus)))
	}
	if masteryLine != "" {
		ge.addLog("info", masteryLine)
	}
	if line := ge.masteryEntryLine(ge.age, 1); line != "" {
		ge.addLog("info", line)
	}
	if n := ge.legacyEpochCount(); n > 0 {
		ge.addLog("info", fmt.Sprintf("Ancient Knowledge from %s you succumbed in: research time %s.",
			textfmt.Count(n, "epoch", "epochs"), ResearchFactorText(ge.succumbResearchFactor())))
	}
	if len(savedRuins) > 0 {
		ge.addLog("info", fmt.Sprintf("Ruins carried forward from past civilizations: %s.",
			textfmt.Count(len(savedRuins), "type", "types")))
	}
	// The legacy kit: shares, the first age's template slice, old friends.
	ge.startRunLegacyLocked()
	ge.addLog("info", "Type [cyan]help[-] to get started again.")

	// Account lifetime stat (Phase 6): record the prestige IN-MEMORY only — we hold
	// ge.mu here, so RecordPrestige must not do I/O or re-enter the engine. The write
	// is deferred to FlushIfDirty in the autosave block (outside ge.mu). A dev-touched
	// run, or one that belongs to another account, records nothing.
	if acct := ge.accountForRecordsLocked(); acct != nil {
		acct.RecordPrestigeFrom(prestigedFrom)
	}

	// Roll for an Ancient Memory cache — prestige level is now >= 1, the age is
	// primitive, and the flag was just cleared above, so this fresh run is eligible.
	ge.maybeOfferAncientMemory()
}

// BuyPrestigeUpgrade purchases a prestige upgrade tier
func (ge *GameEngine) BuyPrestigeUpgrade(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if err := ge.Prestige.BuyUpgrade(key); err != nil {
		return err
	}
	ge.recalculateRates() // a storage or rate upgrade shows at once, not next tick
	ge.addLog("success", ge.prestigeUpgradeLine(key))
	ge.legacyOnPurchaseLocked(key)
	return nil
}

// Reset completely reinitializes the engine to a fresh state (including prestige)
func (ge *GameEngine) Reset() {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.tick = 0
	ge.sessionStart = nil
	ge.age = "primitive_age"
	ge.Resources = NewResourceManagerWith(ge.rules)
	ge.Buildings = ge.newBuildingManager()
	ge.Workers = NewWorkerManagerWith(ge.rules)
	ge.Research = NewResearchManagerWith(ge.rules)
	ge.Military = NewMilitaryManagerWith(ge.rules)
	ge.Events = NewEventManagerWith(ge.rules)
	ge.Milestones = NewMilestoneManagerWith(ge.rules)
	ge.Prestige = NewPrestigeManagerWith(ge.rules)
	ge.Trade = NewTradeManagerWith(ge.rules)
	ge.Diplomacy = NewDiplomacyManagerWith(ge.rules)
	ge.Stats = NewGameStats()
	// Bus intentionally kept — dashboard subscriptions must survive across resets.
	ge.permanentBonuses = make(map[string]float64)
	ge.tickSpeedBonus = 0
	ge.lastK = 0 // no mastery left: the wiped game runs at 1x
	ge.speedMultiplier = 1.0
	ge.buildQueue = nil
	ge.plan = nil
	ge.planLog = nil
	ge.log = nil

	ge.applyAgeUnlocks("primitive_age")
	ge.Resources.Add("food", 25)
	ge.Resources.Add("wood", 50)

	ge.cheaterBadge = false
	ge.eliteBadge = false
	// A new game is a new run: no dev console history, and no owner until
	// StartNewNamedGame (or LoadGame) names one.
	ge.devTouched = false
	ge.runAccountID = ""
	ge.runOrphaned = false
	ge.wonderOverflowOff = false
	ge.currentEpoch = ge.rules.EraOf("primitive_age")
	ge.epochEventFired = make(map[string]bool)
	ge.clearHarbingerRun()
	ge.harbingerHistory = nil
	ge.awakeningsFired = make(map[string]bool)
	ge.survivedEpochs = make(map[string]bool)
	ge.pendingCatastrophe = ""
	ge.epochEventHistory = nil
	ge.legacyBonuses = make(map[string]bool)
	ge.catastropheHistory = nil
	ge.pendingLastPassage = false
	ge.cosmicLegacy = false
	ge.morale = moraleNeutral
	ge.lowMoraleWarned = false
	// Full wipe: no previous civilization, so no cache. Clear the run flag.
	ge.ancientMemoryUsed = false
	ge.pendingMemoryTech = ""
	ge.festivalReadyTick, ge.blackMarketReadyTick = 0, 0
	ge.ageReady = false
	ge.starvationTicks = 0
	ge.autoExpeditionTicksLeft = 0
	ge.autoExpeditionStarved = false
	ge.startRunShares()
	ge.autoRecruitOff = false
	// A wiped game is a brand-new run: re-roll the master seed.
	ge.SeedRNG(newSeed())

	ge.addLog("event", "Game wiped. Starting fresh.")
	ge.addLog("info", "Type [cyan]help[-] for commands.")
	// God mode is process-wide and outlives the run it was turned on in; left on, it
	// makes this run free to build, so the run starts dev-touched.
	if DevGodMode {
		ge.markDevTouchedLocked()
	}
}

// GetState returns a snapshot of the game state for UI
func (ge *GameEngine) GetState() GameState {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	resolver := ge.buildResolver()

	popCap := ge.popCapLocked()
	nextAge := ge.progress.GetNextAge(ge.age)
	sight := ge.ageSightLocked()

	// One epoch lookup per snapshot (was three full table rebuilds).
	epochDef, epochOK := ge.rules.Era(ge.currentEpoch)
	epochColor := "white"
	if epochOK {
		epochColor = epochDef.Color
	}

	logCopy := make([]LogEntry, len(ge.log))
	copy(logCopy, ge.log)

	var queue []BuildQueueSnapshot
	for _, item := range ge.buildQueue {
		def := ge.Buildings.defs[item.BuildingKey]
		queue = append(queue, BuildQueueSnapshot{
			Name:       def.Name,
			TicksLeft:  item.TicksLeft,
			TotalTicks: item.TotalTicks,
		})
	}

	var nextAgeName string
	var nextAgeResReqs map[string]float64
	var nextAgeBldReqs map[string]int
	if nextAge != "" {
		nextAgeName = ge.progress.GetAgeName(nextAge)
		nextAgeResReqs, nextAgeBldReqs = ge.progress.GetRequirementsForNext(ge.age)
	}

	ageOrder := ge.progress.GetAgeOrder()
	knowledgeCount := ge.Workers.GetDomainCount("knowledge")
	// Soldiers are now a real resource; the military panel reflects the resource
	// amount, not the derived military-worker count. The soldier milestones key
	// off the cumulative lifetime trained count instead (ge.Stats.SoldiersTrained).
	soldierResource := int(ge.Resources.Get("soldiers"))
	militaryBonus, expeditionBonus := ge.militaryPower(), ge.expeditionReward()

	// Prestige snapshot with pending points
	prestigeSnap := ge.Prestige.Snapshot()
	prestigeSnap.CanPrestige = ge.Prestige.CanPrestige(ge.age, ageOrder)
	prestigeSnap.PendingPoints = ge.Prestige.CalculatePoints(ge.age)
	if next := ge.progress.GetNextAge(ge.age); next != "" {
		prestigeSnap.NextAge, prestigeSnap.NextAgePoints = next, ge.Prestige.CalculatePoints(next)
	}

	speedMult := ge.speedMultiplier
	if speedMult < 1.0 {
		speedMult = 1.0
	}
	tickInterval := time.Duration(float64(BaseTickInterval) / ((1.0 + ge.tickSpeedBonus) * speedMult))
	if tickInterval < MinTickInterval {
		tickInterval = MinTickInterval
	}

	// MilitaryManager.Snapshot has no *GameEngine receiver, so it cannot see the
	// Geographic Society's buildings, workers or dispatch countdown. Build the
	// snapshot here and graft the automatic-dispatch view on afterwards.
	militarySnap := ge.Military.Snapshot(ge.age, ageOrder, soldierResource, int(ge.Resources.GetStorage("soldiers")), ge.Resources.GetRate("soldiers"), ge.Resources.GetAll(), militaryBonus, expeditionBonus)
	militarySnap.AutoExpedition = ge.autoExpeditionSnapshot()
	militarySnap.Threat = ge.ageThreat(ge.age)
	militarySnap.Mitigation = config.DefenseMitigation(militarySnap.DefenseRating, militarySnap.Threat)
	militarySnap.Saved = ge.Stats.Defense.clone()
	var pendingEndure EndureOutcome
	if ge.pendingCatastrophe != "" {
		pendingEndure = ge.endurePreview(ge.pendingBraceLevel, ge.age)
	}

	// Wonder gate: show which wonder must be built before advancing
	wonderKey := ge.progress.WonderForAge(ge.age)
	currentAgeWonderKey := ""
	currentAgeWonderName := ""
	if wonderKey != "" && ge.Buildings.GetCount(wonderKey) < 1 {
		currentAgeWonderKey = wonderKey
		if def, ok := ge.Buildings.defs[wonderKey]; ok {
			currentAgeWonderName = def.Name
		}
	}

	endured, succumbed := countCatastropheOutcomes(ge.catastropheHistory)
	rngDraws, quipDraws := ge.rngDraws()

	workers := ge.Workers.Snapshot(popCap)
	workers.Shares = cloneShares(ge.workerShares)
	workers.AutoRecruit = !ge.autoRecruitOff
	workers.HoldTicks = max(0, ge.staffHoldUntil-ge.tick)

	return GameState{
		Rules:                ge.rules,
		Tick:                 ge.tick,
		Age:                  ge.age,
		AgeName:              ge.progress.GetAgeName(ge.age),
		AgeReady:             ge.ageReady,
		CurrentAgeWonderKey:  currentAgeWonderKey,
		CurrentAgeWonderName: currentAgeWonderName,
		NextAge:              nextAge,
		NextAgeName:          nextAgeName,
		NextAgeResReqs:       nextAgeResReqs,
		NextAgeBldReqs:       nextAgeBldReqs,
		Resources:            ge.Resources.Snapshot(),
		Buildings:            ge.Buildings.Snapshot(ge.Resources, ge.buildQueue, ge.Workers.GetAssignedCount),
		BuildQueue:           queue,
		Workers:              workers,
		Research:             ge.Research.Snapshot(ge.age, ageOrder),
		Military:             militarySnap,
		Milestones: ge.Milestones.Snapshot(MilestoneSnapshotParams{
			Tick:            ge.tick,
			Age:             ge.age,
			AgeOrder:        ageOrder,
			Resources:       ge.Resources.GetAll(),
			Buildings:       ge.Buildings.GetAll(),
			Population:      ge.Workers.TotalPop(),
			TechCount:       ge.Research.ResearchedCount(),
			TotalBuilt:      ge.Stats.TotalBuilt,
			SoldierCount:    soldierResource,
			SoldiersTrained: int(ge.Stats.SoldiersTrained),
			WonderCount:     ge.countWonders(),
			KnowledgeCount:  knowledgeCount,
			ResearchedTechs: ge.getResearchedTechMap(),
			Sight:           &sight,
			activeEvents:    ge.Events.GetActive(),
		}),
		ActiveEvents:          ge.Events.GetActive(),
		Prestige:              prestigeSnap,
		Mastery:               ge.Prestige.MasterySnapshot(ge.age),
		Trade:                 ge.Trade.Snapshot(ge.age, ageOrder, ge.Buildings, ge.Diplomacy.DisruptedResources()),
		Diplomacy:             ge.Diplomacy.Snapshot(ge.age, ageOrder),
		Log:                   logCopy,
		Stats:                 ge.Stats.Snapshot(),
		SaveExists:            SaveExists("autosave"),
		TickSpeedBonus:        ge.tickSpeedBonus,
		TickIntervalMs:        int(tickInterval.Milliseconds()),
		CheaterBadge:          ge.cheaterBadge,
		EliteBadge:            ge.eliteBadge,
		DevTouched:            ge.devTouched,
		AccountRecords:        ge.accountForRecordsLocked() != nil,
		Seed:                  ge.seed,
		RNGDraws:              rngDraws,
		QuipDraws:             quipDraws,
		LastAgeAdvanceSummary: ge.lastAgeAdvanceSummary.clone(),
		// Phase 8: epoch fields
		EpochKey:              ge.currentEpoch,
		EpochName:             epochDef.Name,
		EpochIcon:             epochDef.Icon,
		EpochColor:            epochColor,
		EpochSurvived:         ge.survivedEpochs[ge.currentEpoch],
		PendingCatastrophe:    ge.pendingCatastrophe,
		PendingEndure:         pendingEndure,
		CatastropheOutlook:    ge.catastropheOutlook(),
		PendingMemoryTech:     ge.pendingMemoryTech,
		PendingMemoryTechName: ge.Research.defs[ge.pendingMemoryTech].Name,
		EpochEventHistory:     slices.Clone(ge.epochEventHistory), // setCatastropheOutcome edits records in place
		Harbinger:             ge.harbingerView(),
		SessionStart:          ge.sessionStart.clone(),
		HarbingerHistory:      cloneHarbingerHistory(ge.harbingerHistory),
		LegacyBonuses: func() map[string]bool {
			out := make(map[string]bool, len(ge.legacyBonuses))
			for k, v := range ge.legacyBonuses {
				out[k] = v
			}
			return out
		}(),
		CatastropheHistory:    slices.Clone(ge.catastropheHistory),
		CatastrophesEndured:   endured,
		CatastrophesSuccumbed: succumbed,
		SuccumbResearchFactor: ge.succumbResearchFactor(),
		LastPassage:           ge.lastPassageState(prestigeSnap.PendingPoints),
		History:               ge.History.Clone(),
		Morale:                ge.morale,
		MoraleCap:             ge.moraleCap(),
		MoraleMultiplier:      ge.moraleMultiplier(),
		Plan:                  ge.planViews(),
		WonderOverflow:        !ge.wonderOverflowOff,
		PermanentBonuses: func() map[string]float64 {
			out := make(map[string]float64, len(ge.permanentBonuses))
			for k, v := range ge.permanentBonuses {
				out[k] = v
			}
			return out
		}(),
		// Snapshot the resolver's aggregated modifiers for the UI's Active
		// Multipliers panel. buildResolver reads only already-held state and
		// pure config.*, so it's safe under the RLock held here (it never
		// re-acquires a lock). All() returns a fresh copy — no shared mutable
		// state escapes.
		Modifiers: resolver.All(),
		Pools:     ge.bonusPoolsLocked(resolver),
		// Account lifetime stats (Phase 6). We hold ge.mu.RLock here; LifetimeStats
		// takes the account's OWN mutex (a.mu) — consistent lock order ge.mu → a.mu,
		// and the Record* writers never hold a.mu while touching ge.mu, so no deadlock.
		// nil when no account is wired (e.g. headless tests without SetAccount).
		AccountStats: func() *AccountStatsView {
			if ge.account == nil {
				return nil
			}
			s, ach := ge.account.LifetimeStats()
			return &AccountStatsView{
				DisplayName:          ge.account.Name(),
				TotalPrestiges:       s.TotalPrestiges,
				HighestAge:           s.HighestAge,
				CivilizationsStarted: s.CivilizationsStarted,
				SavesCompleted:       s.SavesCompleted,
				Achievements:         ach,
			}
		}(),
	}
}

// buildDoneLog is the log category for a finished copy of def: a wonder is
// notable, any other building is routine (the Buildings panel shows its count).
func buildDoneLog(def config.BuildingDef) string {
	if def.Category == "wonder" {
		return "success"
	}
	return LogRoutine
}

// addLog appends a log entry (must be called with lock held). logType is the
// entry's category; see LogRoutine for which categories reach which log.
func (ge *GameEngine) addLog(logType, message string) {
	entry := LogEntry{
		Tick:    ge.tick,
		Message: message,
		Type:    logType,
	}
	ge.log = append(ge.log, entry)
	if len(ge.log) > MaxLogSize {
		ge.log = ge.log[len(ge.log)-MaxLogSize:]
	}
}

// AddLog adds a log entry (thread-safe, for external use)
func (ge *GameEngine) AddLog(logType, message string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.addLog(logType, message)
}

// GetLogs returns a copy of the full log (thread-safe)
func (ge *GameEngine) GetLogs() []LogEntry {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	logCopy := make([]LogEntry, len(ge.log))
	copy(logCopy, ge.log)
	return logCopy
}

const (
	MaxOfflineTime    = 24 * time.Hour
	OfflineEfficiency = 0.5
	// OfflineStepTicks is the step the offline catch-up advances by: each
	// step credits that many ticks of production (at OfflineEfficiency, up
	// to the caps, overflow to the wonder bank, then to the plan's banks),
	// moves construction and research on, and lets the build plan start what
	// the step paid for. A minute at 1x: 24 hours away is 1,440 steps.
	OfflineStepTicks = 30
)

// applyOfflineProgress applies simulated progress for time spent offline
// (must be called with lock held). Time passes in OfflineStepTicks steps, so
// the build plan starts items as the resources for them come in, the caps
// apply along the way, construction and research finish while the player is
// away, and the worker shares routine staffs and recruits (shares.go). With
// an empty plan, nothing under construction, overflow off and nothing for the
// shares routine to do, it pays exactly the old lump sum: rate x ticks x
// OfflineEfficiency, capped.
func (ge *GameEngine) applyOfflineProgress(elapsed time.Duration) {
	if elapsed < 5*time.Second {
		return // too short to matter
	}
	if elapsed > MaxOfflineTime {
		elapsed = MaxOfflineTime
	}

	bonus := ge.tickSpeedBonus
	mult := ge.speedMultiplier
	if mult < 1.0 {
		mult = 1.0
	}
	tickInterval := time.Duration(float64(BaseTickInterval) / ((1.0 + bonus) * mult))
	if tickInterval < MinTickInterval {
		tickInterval = MinTickInterval
	}

	offlineTicks := int(elapsed / tickInterval)
	if offlineTicks <= 0 {
		return
	}

	ge.addLog("event", fmt.Sprintf("Welcome back. You were away for %s.", textfmt.Duration(elapsed)))

	gains := make(map[string]float64)
	banked := make(map[string]float64)
	bankedInto := ""
	planBanked := make(map[string]float64)
	var starts planStarts
	var staffed staffCounts
	for done := 0; done < offlineTicks; {
		n := min(OfflineStepTicks, offlineTicks-done)
		// Stop at the fate's next event, so a harbinger arrives and a doom
		// strikes at its own tick while the player is away.
		if k := ge.fateNextEventIn(); k > 0 && k < n {
			n = k
		}
		w := ge.overflowWonder()
		losses := ge.overflowScratch[:0]
		ge.Resources.AddProduced(float64(n)*OfflineEfficiency,
			func(res string, g float64) { gains[res] += g },
			func(res string, lost float64) {
				if b := ge.bankOverflow(w, res, lost); b > 0 {
					banked[res] += b
					bankedInto = w
					lost -= b
				}
				if lost > 0 {
					losses = append(losses, overflowLoss{res: res, amount: lost})
				}
			})
		ge.overflowScratch = losses
		// What the wonder didn't take goes toward the plan's queued copies.
		ge.bankPlanOverflow(losses, planBanked)
		ge.tick += n
		done += n
		ge.Trade.DecayPressure(n) // the plan's trades meet a market that recovers as time passes
		changed := ge.advanceBuildQueue(n)
		if ge.advanceResearch(n) {
			changed = true
		}
		if changed {
			ge.recalculateRates()
		}
		ge.runPlan(&starts)
		// The shares routine staffs and recruits as the time passes, so
		// what the plan builds gets workers and the workforce grows.
		if !ge.workersHeld() {
			if c := ge.staffByShares(true); c.any() {
				ge.recalculateRates()
				staffed.add(c)
			}
		}
		// A fated doom keeps its hour offline: the harbinger comes and the doom
		// strikes at its tick, and a struck doom waits, pending, for the player.
		ge.harbingerTickCheck()
	}

	if len(gains) > 0 {
		ge.addLog("info", fmt.Sprintf("Offline progress (at %s efficiency):", textfmt.Percent(OfflineEfficiency)))
		for _, res := range sortedKeys(gains) {
			ge.addLog("info", "  +"+Amount(gains[res], res))
		}
	}
	if len(banked) > 0 {
		ge.addLog("info", fmt.Sprintf("Overflow banked into %s: %s.", ge.Buildings.defs[bankedInto].Name, Amounts(banked)))
	}
	if len(planBanked) > 0 {
		ge.addLog("info", fmt.Sprintf("Overflow banked toward your plan: %s.", Amounts(planBanked)))
	}
	if !starts.empty() {
		ge.addLog("info", "While you were away, your plan "+starts.describe(ge.rules)+".")
	}
	if staffed.any() {
		ge.addLog("info", "While you were away, your worker shares "+staffed.describe(ge.Workers.TotalPop(), ge.popCapLocked())+".")
	}
}

// ExchangeResources performs a resource exchange via the trade system
func (ge *GameEngine) ExchangeResources(from, to string, amount float64) (float64, error) {
	if err := checkAmount(amount); err != nil {
		return 0, err
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.Trade.SetAge(ge.age)
	got, err := ge.Trade.Exchange(from, to, amount, ge.Resources, ge.Buildings, ge.tick)
	if err != nil {
		return 0, err
	}
	ge.addLog(LogRoutine, fmt.Sprintf("Traded %s for %s.", Amount(amount, from), Amount(got, to)))
	return got, nil
}

// StartTradeRoute activates a trade route
func (ge *GameEngine) StartTradeRoute(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	if err := ge.Trade.StartRoute(key, ge.Buildings, ge.age, ageOrder); err != nil {
		return err
	}
	ge.addLog(LogRoutine, fmt.Sprintf("Trade route started: %s.", ge.rules.Name(rules.KindRoute, key)))
	return nil
}

// StopTradeRoute deactivates a trade route
func (ge *GameEngine) StopTradeRoute(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if err := ge.Trade.StopRoute(key); err != nil {
		return err
	}
	ge.addLog(LogRoutine, fmt.Sprintf("Trade route stopped: %s.", ge.rules.Name(rules.KindRoute, key)))
	return nil
}

// SetDiplomaticStatus changes diplomatic status with a faction
func (ge *GameEngine) SetDiplomaticStatus(factionKey, status string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	gold := ge.Resources.Get("gold")
	cost, err := ge.Diplomacy.SetStatus(factionKey, status, gold)
	if err != nil {
		return err
	}
	if cost > 0 {
		ge.Resources.Remove("gold", cost)
	}
	ge.addLog("info", diplomaticStatusLine(ge.rules.Name(rules.KindCiv, factionKey), status, cost))
	return nil
}

// SendGift sends a gift to a faction
func (ge *GameEngine) SendGift(factionKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	gold := ge.Resources.Get("gold")
	before := ge.civOpinion(factionKey)
	cost, err := ge.Diplomacy.SendGift(factionKey, gold)
	if err != nil {
		return err
	}
	ge.Resources.Remove("gold", cost)
	ge.addLog(LogRoutine, fmt.Sprintf("Sent the %s a gift: %s, opinion %s.",
		ge.rules.Name(rules.KindCiv, factionKey), Amount(cost, "gold"), textfmt.Signed(float64(ge.civOpinion(factionKey)-before))))
	return nil
}

// SendTribute pays gold + culture to a civilization at war to sue for peace.
// Cost scales with the civ's strength; the war ends immediately on success.
func (ge *GameEngine) SendTribute(factionKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	gold := ge.Resources.Get("gold")
	culture := ge.Resources.Get("culture")
	goldCost, cultureCost, err := ge.Diplomacy.SendTribute(factionKey, gold, culture)
	if err != nil {
		return err
	}
	ge.Resources.Remove("gold", goldCost)
	ge.Resources.Remove("culture", cultureCost)
	ge.addLog("success", fmt.Sprintf("Paid the %s a tribute of %s. The war is over.",
		ge.rules.Name(rules.KindCiv, factionKey), Amounts(map[string]float64{"gold": goldCost, "culture": cultureCost})))
	return nil
}

// RaidCivRoute raids a discovered civilization's trade route — a provocation
// that tanks opinion and may trigger war if standing is already hostile.
func (ge *GameEngine) RaidCivRoute(factionKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	before := ge.civOpinion(factionKey)
	started, err := ge.Diplomacy.RaidTradeRoute(factionKey, ge.tick)
	if err != nil {
		return err
	}
	name := ge.rules.Name(rules.KindCiv, factionKey)
	if started {
		ge.addLog("warning", fmt.Sprintf("You raided a %s trade route. They declared war.", name))
	} else {
		ge.addLog("warning", fmt.Sprintf("You raided a %s trade route. Opinion %s.",
			name, textfmt.Signed(float64(ge.civOpinion(factionKey)-before))))
	}
	return nil
}

// UpgradeBuilding converts count copies of a legacy building to its pending next-tier
// equivalent, charging the cost delta (new copy cost minus 50% refund on old copy) per unit.
// Pass all=true or count<=0 to upgrade all available copies.
func (ge *GameEngine) UpgradeBuilding(key string, count int, all bool) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	oldDef, hasOld := ge.rules.Building(key)
	if !hasOld {
		return ge.unknownBuildingErr(key)
	}
	newKey, hasPending := ge.Buildings.GetPendingUpgrade(key)
	newDef, hasNew := ge.rules.Building(newKey)
	if !hasPending || !hasNew {
		return fmt.Errorf("%s has no upgrade this age. Type 'upgrade' to list the ones that do.", oldDef.Name)
	}
	if err := ge.techLockErr(newDef); err != nil {
		return err
	}

	oldCount := ge.Buildings.GetCount(key)
	if all || count <= 0 {
		count = oldCount
	}
	if count > oldCount {
		count = oldCount
	}
	if count <= 0 {
		return fmt.Errorf("You have no %s to upgrade.", oldDef.Name)
	}
	if room := ge.Buildings.UpgradeRoom(newKey, count, ge.buildQueue); room < count {
		if room <= 0 {
			return fmt.Errorf("%s is at its max count of %d.", newDef.Name, newDef.MaxCount)
		}
		count = room
	}

	cost, ok := ge.Buildings.UpgradeCost(key, newKey, count)
	if !ok {
		return fmt.Errorf("Could not work out the cost to upgrade %s. Please report this bug.", oldDef.Name)
	}

	if !ge.Resources.CanAfford(cost) {
		return fmt.Errorf("Cannot afford to upgrade %s: need %s.", buildingCountIn(ge.rules, count, key), ge.shortfallText(cost))
	}

	// Deduct resources
	for res, amt := range cost {
		ge.Resources.Add(res, -amt)
	}

	// Perform partial transform
	moved := ge.Buildings.PartialTransform(key, newKey, count, ge.Workers.RenameAssignment)
	ge.rehomeUpgradedWorkers(key, newKey, oldDef, newDef)
	ge.holdStaffing()

	ge.recalculateRates()

	costStr := Amounts(cost)
	if costStr == "nothing" {
		costStr = "free"
	}
	ge.addLog(LogRoutine, fmt.Sprintf("Upgraded %s to %s. Cost: %s.",
		buildingCountIn(ge.rules, moved, key), pluralName(moved, newDef.Name), costStr))
	return nil
}

// rehomeUpgradedWorkers settles workers after an upgrade. A partial upgrade
// left the old key's workers where they were, which could be more than the
// copies left can hold (upgrade 7 of 46 wood camps and 138 workers sat in
// 117 slots). The overflow follows the upgraded copies to the new building
// while it has room; the rest go back to the idle pool. Under the write lock.
func (ge *GameEngine) rehomeUpgradedWorkers(oldKey, newKey string, oldDef, newDef config.BuildingDef) {
	excess := ge.Workers.GetAssignedCount("worker", oldKey) - oldDef.WorkerCapacity*ge.Buildings.GetCount(oldKey)
	if excess <= 0 {
		return
	}
	ge.Workers.Unassign("worker", oldKey, excess)
	room := newDef.WorkerCapacity*ge.Buildings.GetCount(newKey) - ge.Workers.GetAssignedCount("worker", newKey)
	if room > excess {
		room = excess
	}
	if room > 0 {
		ge.Workers.Assign("worker", newKey, room)
	}
}

// GetAvailableUpgrades returns upgrade info for buildings that have a pending player-driven upgrade.
func (ge *GameEngine) GetAvailableUpgrades() []UpgradeInfo {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	var result []UpgradeInfo

	for oldKey, newKey := range ge.Buildings.pendingUpgrades {
		count := ge.Buildings.counts[oldKey]
		if count <= 0 {
			continue
		}
		oldDef, ok1 := ge.rules.Building(oldKey)
		newDef, ok2 := ge.rules.Building(newKey)
		if !ok1 || !ok2 {
			continue
		}
		if count = ge.Buildings.UpgradeRoom(newKey, count, ge.buildQueue); count <= 0 {
			continue
		}
		cost, ok := ge.Buildings.UpgradeCost(oldKey, newKey, count)
		if !ok {
			cost = make(map[string]float64)
		}
		canAfford := ge.Resources.CanAfford(cost)
		result = append(result, UpgradeInfo{
			FromKey:   oldKey,
			ToKey:     newKey,
			FromName:  oldDef.Name,
			ToName:    newDef.Name,
			Count:     count,
			Cost:      cost,
			CanAfford: canAfford,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].FromKey < result[j].FromKey })
	return result
}

// UpgradeInfo describes an available building upgrade for display
type UpgradeInfo struct {
	FromKey   string
	ToKey     string
	FromName  string
	ToName    string
	Count     int
	Cost      map[string]float64
	CanAfford bool
}

// countWonders returns the total number of wonders built (must be called with lock held)
func (ge *GameEngine) countWonders() int {
	count := 0
	for key, c := range ge.Buildings.counts {
		if def, ok := ge.Buildings.defs[key]; ok && def.Category == "wonder" && c > 0 {
			count += c
		}
	}
	return count
}

// getResearchedTechMap returns a map of researched tech keys (must be called with lock held)
func (ge *GameEngine) getResearchedTechMap() map[string]bool {
	m := make(map[string]bool, len(ge.Research.researched))
	for key := range ge.Research.researched {
		m[key] = true
	}
	return m
}

// formatMilestoneRewards formats milestone reward effects for the toast:
// "(+10% all production, +500 food)". Empty when there are no rewards.
func formatMilestoneRewards(effects []config.Effect) string {
	parts := milestoneRewardParts(effects)
	if parts == "" {
		return ""
	}
	return "(" + parts + ")"
}

// milestoneRewardParts lists milestone rewards with display names:
// "+10% all production, +500 food".
func milestoneRewardParts(effects []config.Effect) string {
	var parts []string
	for _, e := range effects {
		switch e.Type {
		case "instant_resource":
			parts = append(parts, "+"+Amount(e.Value, e.Target))
		case "permanent_bonus":
			parts = append(parts, textfmt.SignedPercent(e.Value)+" "+EffectTargetName(e.Target))
		}
	}
	return strings.Join(parts, ", ")
}

// durationWithSpeedBonusLocked is durationLocked for a span during which an
// extra tick_speed bonus applies (a chain boost shortens its own wall-clock
// length). Caller holds ge.mu.
func (ge *GameEngine) durationWithSpeedBonusLocked(ticks int, extra float64) string {
	return DurationText(ticks, ge.tickIntervalWithBonusLocked(ge.tickSpeedBonus+extra))
}

// diplomaticStatusLine announces a status change from the player's side:
// "You are now allied with the Merchant Guild."
func diplomaticStatusLine(civ, status string, cost float64) string {
	switch status {
	case "allied":
		if cost > 0 {
			return fmt.Sprintf("You are now allied with the %s. The alliance cost %s.", civ, Amount(cost, "gold"))
		}
		return fmt.Sprintf("You are now allied with the %s.", civ)
	case "rival":
		return fmt.Sprintf("You declared the %s your rival.", civ)
	case "embargo":
		return fmt.Sprintf("You placed an embargo on the %s.", civ)
	case "neutral":
		return fmt.Sprintf("You are now neutral toward the %s.", civ)
	}
	return fmt.Sprintf("Your status with the %s is now %s.", civ, status)
}

// civOpinion reads a civilization's opinion of the player (0 when unknown).
// Caller holds ge.mu.
func (ge *GameEngine) civOpinion(key string) int {
	if fs, ok := ge.Diplomacy.factions[key]; ok && fs != nil {
		return fs.Opinion
	}
	return 0
}

// prestigeUpgradeLine announces a prestige purchase with its total effect at
// the new tier: "Bought Gather Boost (tier 2): worker output +10%."
func (ge *GameEngine) prestigeUpgradeLine(key string) string {
	def, ok := ge.rules.PrestigeUpgrade(key)
	if !ok {
		return fmt.Sprintf("Bought %s.", ge.rules.Name(rules.KindPrestigeUpgrade, key))
	}
	if def.EffectType == "legacy" {
		// A one-tier kit item: say what it does from now on.
		return fmt.Sprintf("Bought %s. %s.", def.Name, def.Description)
	}
	tier := ge.Prestige.upgrades[key]
	total := float64(def.PerTier * float64(tier))
	var effect string
	switch def.EffectType {
	case "rate_bonus":
		effect = EffectTargetName(def.EffectKey) + " " + textfmt.SignedPercent(total)
	case "flat_bonus":
		switch def.EffectKey {
		case "all":
			effect = "storage for every resource " + textfmt.Signed(total)
		default:
			effect = EffectTargetName(def.EffectKey) + " " + textfmt.Signed(total)
		}
	case "starting_resource":
		effect = "each run starts with +" + Amount(total, def.EffectKey)
	}
	if effect == "" {
		return fmt.Sprintf("Bought %s (tier %d).", def.Name, tier)
	}
	return fmt.Sprintf("Bought %s (tier %d): %s.", def.Name, tier, effect)
}
