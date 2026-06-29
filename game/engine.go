package game

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
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

	// Festival (culture sink) tuning.
	festivalBuffPercent   = 0.20 // +20% production_all while active
	festivalBuffTicks     = 150  // ~5 minutes at 2s/tick
	festivalCooldownTicks = 300  // ~10 minutes between festivals
	festivalMinCost       = 2000.0
	festivalCostFraction  = 0.05 // of culture storage cap

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

	// Permanent bonuses from milestones
	permanentBonuses map[string]float64

	// Dynamic tick speed
	tickSpeedBonus  float64
	speedMultiplier float64

	// Age advancement — set when requirements are met; player must type 'advance' to proceed
	ageReady bool

	// Starvation tracking — counts consecutive ticks with food <= 0 and active drain
	starvationTicks int

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
	//     package default rand. Tests inject a seeded *rand.Rand for determinism.
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
	// boot (accounts.md §2/§8) and held here so the UI/dashboard — already sharing
	// this engine — can reach it via Account(). It is player-level, NOT per-save, so
	// Reset() must NOT clear it (a new game keeps the same player). May be nil if
	// LoadOrCreate failed at boot — account state is non-critical, the game runs anyway.
	account *Account

	// Morale system — a managed two-way dial. Range [0.10, moraleCap()];
	// starts at moraleNeutral (0.50). Drives production via moraleMultiplier().
	morale          float64 // 0.10–moraleCap(); starts at moraleNeutral (0.50)
	lowMoraleWarned bool    // true after morale warning fired; reset when morale rises above 0.40

	// festivalReadyTick is the earliest game tick a new festival may be held
	// (cooldown anti-spam for the `festival` culture sink). 0 = ready now.
	festivalReadyTick int
}

// BuildQueueItem represents a building under construction
type BuildQueueItem struct {
	BuildingKey string
	TicksLeft   int
	TotalTicks  int
}

// NewGameEngine creates a new game engine initialised to the Primitive Age.
// Callers must call Start() to begin the tick loop.
func NewGameEngine() *GameEngine {
	ge := &GameEngine{
		age:              "primitive_age",
		Resources:        NewResourceManager(),
		Buildings:        NewBuildingManager(),
		Workers:          NewWorkerManager(),
		Research:         NewResearchManager(),
		Military:         NewMilitaryManager(),
		Events:           NewEventManager(),
		Milestones:       NewMilestoneManager(),
		Prestige:         NewPrestigeManager(),
		Trade:            NewTradeManager(),
		Diplomacy:        NewDiplomacyManager(),
		Stats:            NewGameStats(),
		Bus:              NewEventBus(),
		progress:         NewProgressManager(),
		permanentBonuses: make(map[string]float64),
		speedMultiplier:  1.0,
		morale:           moraleNeutral,
		stopCh:           make(chan struct{}),
		currentEpoch:     config.EpochForAge("primitive_age"),
		epochEventFired:  make(map[string]bool),
		awakeningsFired:  make(map[string]bool),
		survivedEpochs:   make(map[string]bool),
		legacyBonuses:    make(map[string]bool),
		History:          NewHistoryCollector(),
	}
	ge.applyAgeUnlocks("primitive_age")
	// Give starting resources — enough for first hut + a little food
	ge.Resources.Add("food", 25)
	ge.Resources.Add("wood", 50)
	// Startup flavor — the step-by-step onboarding now lives in the Buildings panel
	// (main-screen polish part 1), so the log stays clean for live events.
	ge.addLog("event", "Welcome to AgeForge! You have nothing but your hands.")
	// Subscribe to age advances to record markers in history.
	// IMPORTANT: Bus handlers run under the engine write lock — do NOT call GetState().
	ge.Bus.Subscribe(EventAgeAdvanced, func(e EventData) {
		ageName, _ := e.Payload["new_age"].(string)
		ge.History.MarkAge(ge.tick, ageName)
	})
	return ge
}

const AutosaveInterval = 60 * time.Second

// moraleCap returns 1.0 + 0.05 per wonder built.
func (ge *GameEngine) moraleCap() float64 {
	cap := 1.0
	for key, count := range ge.Buildings.counts {
		if count > 0 {
			def, ok := ge.Buildings.defs[key]
			if ok && def.Category == "wonder" {
				cap += 0.05 * float64(count)
			}
		}
	}
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
	ge.morale += delta
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
		return 1.0 + frac*moraleMaxBonus
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
		return 1.0 - frac*(1.0-moraleMinMult)
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
	for key, count := range ge.Buildings.counts {
		if count == 0 {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		if !ok {
			continue
		}
		for _, eff := range def.Effects {
			if eff.Type == "morale" {
				moraleFromBuildings += eff.Value * float64(count)
			}
		}
	}
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
		ge.addLog("warning", fmt.Sprintf("⚠ Morale critical: %.0f%% — worker output severely reduced", ge.morale*100))
	} else if ge.morale >= 0.40 && ge.lowMoraleWarned {
		ge.lowMoraleWarned = false
	}
}

// Start begins the game tick loop in the calling goroutine. It blocks until
// Stop is called. Safe to call again after Stop — the stop channel is
// re-initialised so the engine can restart (e.g. ESC → splash → New Game).
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
	ge.mu.Unlock()

	timer := time.NewTimer(ge.getTickInterval())
	defer timer.Stop()

	lastAutosave := time.Now()

	for {
		select {
		case <-timer.C:
			ge.safeTick()

			// Periodic autosave (outside the tick lock) → overwrite the active save
			// slot, not a fixed "autosave" file. ActiveSaveName takes its own RLock;
			// safe here because we are outside the tick write lock.
			if time.Since(lastAutosave) >= AutosaveInterval {
				if err := ge.SaveGame(ge.ActiveSaveName()); err != nil {
					ge.mu.Lock()
					ge.addLog("warning", fmt.Sprintf("Autosave failed: %v", err))
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
						ge.addLog("warning", fmt.Sprintf("Account flush failed: %v", err))
						ge.mu.Unlock()
					}
				}

				lastAutosave = time.Now()
			}

			timer.Reset(ge.getTickInterval())
		case <-ge.stopCh:
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
			ge.addLog("error", fmt.Sprintf("Tick recovered from panic: %v", r))
			ge.mu.Unlock()
		}
	}()
	ge.doTick()
}

// getTickInterval computes the current tick interval from all speed sources.
// Called by the tick goroutine between ticks — no lock needed because
// tickSpeedBonus and speedMultiplier are only written under the write lock
// inside doTick/LoadGame, which runs on the same goroutine before this call.
func (ge *GameEngine) getTickInterval() time.Duration {
	// tickSpeedBonus and speedMultiplier are only written under the write lock
	// in doTick/LoadGame, and this is called from the same goroutine after
	// doTick returns, so a direct read is safe here.
	bonus := ge.tickSpeedBonus
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

// MaxSpeedForAge returns the maximum speed multiplier gated by wonders.
// Each wonder built adds +0.5x on top of the 1.0x base, so players must
// invest in wonders to unlock higher speed settings via the `speed` command.
// NOTE: Caller must hold at least an RLock if called from outside the tick goroutine.
func (ge *GameEngine) MaxSpeedForAge() float64 {
	wonderCount := 0
	for key, count := range ge.Buildings.counts {
		if def, ok := ge.Buildings.defs[key]; ok && def.Category == "wonder" && count > 0 {
			wonderCount++
		}
	}
	return 1.0 + float64(wonderCount)*0.5
}

// SetSpeedMultiplier sets the game speed multiplier (0.5 increments, capped by age)
func (ge *GameEngine) SetSpeedMultiplier(mult float64) error {
	// Validate it's a 0.5 increment and at least 1.0
	if mult < 1.0 || mult != float64(int(mult*2))/2 {
		return fmt.Errorf("invalid speed: %.1f (must be 1.0, 1.5, 2.0, etc.)", mult)
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	maxSpeed := ge.MaxSpeedForAge()
	if mult > maxSpeed {
		return fmt.Errorf("speed %.1fx not unlocked yet (max: %.1fx — build more wonders!)", mult, maxSpeed)
	}
	ge.speedMultiplier = mult
	ge.addLog("info", fmt.Sprintf("Game speed set to %.1fx", mult))
	return nil
}

// GetSpeedMultiplier returns the current speed multiplier
func (ge *GameEngine) GetSpeedMultiplier() float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.speedMultiplier
}

// GetMaxSpeed returns the max speed allowed for the current age (thread-safe)
func (ge *GameEngine) GetMaxSpeed() float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.MaxSpeedForAge()
}

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

// SetAccount installs the per-player account, loaded once at boot. May be nil.
// The account is player-level state and survives Reset (new game / succumb), so it
// is set here rather than in NewGameEngine or Reset (accounts.md §6).
func (ge *GameEngine) SetAccount(a *Account) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.account = a
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
// stamp rides the normal SaveGame path so _sig re-signs over it (accounts.md §6).
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

// SwitchAccount makes account id active and installs it as ge.account (Phase B). It is a
// start-screen operation: game.SwitchAccount repoints the active pointer + loads the slot,
// then on success SetAccount swaps the live account under the write lock. It does NOT reset
// running game state (the UI handles re-theming and any new-game flow). On error ge.account
// is left as-is.
func (ge *GameEngine) SwitchAccount(id string) error {
	acct, err := SwitchAccount(id)
	if err != nil {
		return err
	}
	ge.SetAccount(acct)
	return nil
}

// ImportAccountExport restores a single-account backup blob into the account's OWN slot
// (Phase C) and returns the imported account. It is a thin passthrough to
// game.ImportAccountExport: that function resolves the blob's target slot by its AccountID,
// creates-or-merges the data there, and DELIBERATELY does not change the active account — so
// importing account B's backup never disturbs the live account A. No ge.mu is taken: the work
// is file I/O over the account slots, and it touches no engine state (it does NOT auto-install
// the result as ge.account). The caller decides whether to SwitchAccount to the imported id.
func (ge *GameEngine) ImportAccountExport(blob []byte, merge bool) (*Account, error) {
	return ImportAccountExport(blob, merge)
}

// CreateAccount creates (or, for an existing same-name slot, opens) a name-derived account
// and installs it as ge.account (Phase B). It is the no-carry-over create: a brand-new
// account starts empty (see game.CreateAccount). Like SwitchAccount it is a start-screen
// operation — SetAccount swaps the live account under the write lock; no running-game reset.
func (ge *GameEngine) CreateAccount(name string) (*Account, error) {
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
func (ge *GameEngine) WipeAccountByID(id string) (string, error) {
	wasActive := ge.AccountID() == id && id != ""
	backupPath, err := WipeAccountByID(id)
	if err != nil {
		return backupPath, err
	}
	if wasActive {
		ge.SetAccount(nil)
	}
	return backupPath, nil
}

// StartNewNamedGame resets to a fresh game, makes `name` the active root save,
// and writes the initial save file. It does NOT start the ticker — the caller
// starts it. Returns the SaveGame error if any.
func (ge *GameEngine) StartNewNamedGame(name string) error {
	ge.Reset()
	// Set names AFTER Reset so it can't clobber them (Reset leaves them alone, but
	// the ordering keeps that guarantee local to this call).
	ge.SetActiveSaveName(name)
	ge.SetActiveParentName("") // root of a new lineage
	return ge.SaveGame(name)
}

// Stop halts the game tick loop. Safe to call multiple times; subsequent
// calls are no-ops. After Stop returns the Start goroutine has exited.
func (ge *GameEngine) Stop() {
	ge.stopOnce.Do(func() {
		ge.mu.Lock()
		ge.running = false
		ge.mu.Unlock()
		close(ge.stopCh)
		// Flush any pending account lifetime stats on a clean exit so a prestige/age-up
		// since the last autosave isn't lost (Phase 6). Outside ge.mu — Save does I/O —
		// and via the locking accessor. No-op when not dirty.
		if acct := ge.Account(); acct != nil {
			_ = acct.FlushIfDirty()
		}
	})
}

// doTick processes one game tick. It holds the write lock for its entire
// duration, which means all Bus handlers that fire here (via Publish) also
// run under the write lock. Bus handlers MUST NOT call GetState() or any
// other method that acquires the lock — doing so will deadlock.
func (ge *GameEngine) doTick() {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.tick++

	// Process build queue
	ge.processBuildQueue()
	if len(ge.buildQueue) > 0 {
		ge.addLog("debug", fmt.Sprintf("Build queue: %d item(s) in progress", len(ge.buildQueue)))
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

	// Apply resource rates (production - consumption)
	ge.Resources.ApplyRates()

	// Credit the lifetime soldiers-trained counter with the post-clamp delta.
	// Soldiers discarded at the storage cap don't count; the helper floors at 0
	// so a net drain never reduces the lifetime total.
	ge.Stats.RecordSoldiersTrained(ge.Resources.Get("soldiers") - soldiersBefore)

	// Log net food rate and capped resources every 10 ticks
	if ge.tick%10 == 0 {
		snap := ge.Resources.Snapshot()
		if f, ok := snap["food"]; ok {
			ge.addLog("debug", fmt.Sprintf("Food: %.1f (rate %+.3f/t), pop=%d", f.Amount, f.Rate, ge.Workers.TotalPop()))
		}
		for key, rs := range snap {
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

	// Starvation: when food is at 0 with active drain, workers die every 5 ticks
	if ge.Resources.Get("food") <= 0 && ge.Workers.FoodDrain() > 0 {
		ge.starvationTicks++
		if ge.starvationTicks == 1 {
			ge.addLog("warning", "⚠ Your people are starving! Food has run out.")
		}
		if ge.starvationTicks%5 == 0 && ge.Workers.TotalPop() > 0 {
			killed := ge.Workers.KillWorker(1)
			if killed > 0 {
				ge.addLog("error", fmt.Sprintf("☠ A worker has died of starvation! (pop: %d)", ge.Workers.TotalPop()))
			}
		}
	} else if ge.starvationTicks > 0 {
		ge.starvationTicks = 0
		ge.addLog("info", "✓ Food supply restored — starvation ended.")
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
			ge.addLog("event", fmt.Sprintf("✦ Ready to advance to the %s! Type 'advance' when you're ready.", nextName))
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
			ProdAll:    ge.permanentBonuses["production_all"],
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
		def := config.TechByKey()[completed]
		ge.addLog("debug", fmt.Sprintf("Research complete: %s", def.Name))
		ge.addLog("success", fmt.Sprintf("Research complete: %s!", def.Name))
		ge.Bus.Publish(EventData{
			Type:    EventResearchDone,
			Payload: map[string]interface{}{"tech": completed},
		})
	} else if ge.Research.currentTech != "" {
		ge.addLog("debug", fmt.Sprintf("Research: %s %d/%d ticks",
			ge.Research.currentTech, ge.Research.totalTicks-ge.Research.ticksLeft, ge.Research.totalTicks))
	}
}

// processEvents handles random events
func (ge *GameEngine) processEvents() {
	ageOrder := ge.progress.GetAgeOrder()
	triggered, expired := ge.Events.Tick(ge.tick, ge.age, ageOrder, ge.currentEpoch)

	for _, def := range triggered {
		ge.addLog("debug", fmt.Sprintf("Event triggered: %s (sentiment: %s)", def.Name, def.Sentiment))
		ge.addLog("event", def.LogMessage)
		// Process instant and on-trigger effects.
		// For timed events (Duration > 0) the losses are also recorded on the active event
		// so that when it expires the "has ended" message includes a yellow summary.
		isTimed := def.Duration > 0
		for _, eff := range def.Effects {
			switch eff.Type {
			case "instant_resource":
				ge.Resources.Add(eff.Target, eff.Value)
				ge.addLog("debug", fmt.Sprintf("Event effect: %s %s %+.1f", eff.Type, eff.Target, eff.Value))
			case "steal_resource":
				current := ge.Resources.Get(eff.Target)
				loss := eff.Value
				if loss > current {
					loss = current
				}
				ge.Resources.Remove(eff.Target, loss)
				ge.addLog("debug", fmt.Sprintf("Event effect: %s %s -%.1f", eff.Type, eff.Target, loss))
				if isTimed && loss > 0 {
					ge.Events.RecordResourceLoss(def.Key, eff.Target, loss)
				}
			case "worker_loss":
				// Value is a percentage (0.0–1.0) of total workers to remove
				lost := int(float64(ge.Workers.TotalPop()) * eff.Value)
				if lost < 1 {
					lost = 1
				}
				ge.Workers.RemovePct(eff.Value)
				if isTimed {
					// Loss will be reported in the "has ended" summary; skip standalone log
					ge.Events.RecordWorkerLoss(def.Key, lost)
				} else {
					ge.addLog("warning", fmt.Sprintf("%d workers fled or were lost.", lost))
				}
				ge.addLog("debug", fmt.Sprintf("Event effect: worker_loss %.0f%%", eff.Value*100))
			}
		}
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
	prestigeBonuses := ge.Prestige.GetBonuses()
	wonderBonuses := ge.getWonderBonuses()
	militaryBonus := ge.Research.GetBonus("military_power") + ge.permanentBonuses["military_power"] + prestigeBonuses["military_power"] + wonderBonuses["military_power"]
	expeditionBonus := ge.Research.GetBonus("expedition_reward") + ge.permanentBonuses["expedition_reward"] + prestigeBonuses["expedition_reward"] + wonderBonuses["expedition_reward"]
	for _, cat := range []string{ExpeditionScouting, ExpeditionMilitary} {
		if active := ge.Military.ActiveByCategory(cat); active != nil {
			ge.addLog("debug", fmt.Sprintf("Expedition: %s %d ticks left", active.Name, active.TicksLeft))
		}
	}
	// Tick all active expeditions (one per category); each may resolve this tick.
	for _, res := range ge.Military.Tick(militaryBonus, expeditionBonus) {
		ge.addLog("debug", fmt.Sprintf("Expedition resolved (rewards: %d types)", len(res.Rewards)))
		ge.addLog("event", res.Message)
		// Add rewards to resources
		for resource, amount := range res.Rewards {
			ge.Resources.Add(resource, amount)
		}
	}
}

// processTrade handles trade route ticks
func (ge *GameEngine) processTrade() {
	messages := ge.Trade.Tick(ge.Resources, ge.Buildings, ge.Diplomacy)
	for _, msg := range messages {
		ge.addLog("warning", msg)
	}
}

// processDiplomacy handles diplomacy ticks
func (ge *GameEngine) processDiplomacy() {
	ageOrder := ge.progress.GetAgeOrder()
	// Mercantile civs warm to trade activity: treat any active trade route as
	// "traded recently" this window. TradeManager.RecordTrade already runs in
	// the same lock, so reading the active count here is safe.
	tradedRecently := ge.Trade.ActiveRouteCount() > 0
	messages := ge.Diplomacy.Tick(ge.age, ageOrder, ge.tick, tradedRecently)
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
	// Apply war raids (resource losses).
	for _, raid := range ge.Diplomacy.TakePendingRaids() {
		ge.Resources.Remove(raid.Resource, raid.Amount)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "war_raid",
			Name:      "Under Raid",
			TicksLeft: lendEventDisplayTicks,
		})
	}

	// Embassies passively generate opinion toward non-hostile factions.
	// Total/tick = Σ over embassy-type buildings of:
	//   perWorkerRate × workerCapacity × count × (0.20 + 0.80 × assigned/totalCap)
	// mirroring the production worker-fill curve so a staffed embassy outperforms
	// an empty one. perWorkerRate comes from the building's "opinion" effect.
	totalOpinion := 0.0
	for _, key := range []string{"embassy", "grand_embassy"} {
		count := ge.Buildings.GetCount(key)
		if count == 0 {
			continue
		}
		def, ok := config.BuildingByKey()[key]
		if !ok {
			continue
		}
		var perWorker float64
		for _, eff := range def.Effects {
			if eff.Type == "opinion" {
				perWorker = eff.Value
				break
			}
		}
		if perWorker <= 0 || def.WorkerCapacity <= 0 {
			continue
		}
		assigned := ge.Workers.GetAssignedCount("worker", key)
		totalCap := float64(count * def.WorkerCapacity)
		fill := float64(assigned) / totalCap
		if fill > 1.0 {
			fill = 1.0
		}
		totalOpinion += perWorker * float64(def.WorkerCapacity) * float64(count) * (0.20 + 0.80*fill)
	}
	if totalOpinion > 0 {
		ge.Diplomacy.AddPassiveOpinion(totalOpinion)
	}
}

// checkMilestones checks for newly completed milestones and chains
func (ge *GameEngine) checkMilestones() {
	ageOrder := ge.progress.GetAgeOrder()
	researchedTechs := make(map[string]bool)
	for _, key := range ge.Research.GetResearched() {
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
		ge.addLog("success", fmt.Sprintf("Milestone achieved: %s!", ms.Name))
		// Apply rewards
		for _, eff := range ms.Rewards {
			switch eff.Type {
			case "instant_resource":
				ge.Resources.Add(eff.Target, eff.Value)
			case "permanent_bonus":
				ge.permanentBonuses[eff.Target] += eff.Value
			}
		}
		// Publish milestone event
		ge.Bus.Publish(EventData{
			Type: EventMilestoneCompleted,
			Payload: map[string]interface{}{
				"name":        ms.Name,
				"key":         ms.Key,
				"reward_text": rewardText,
			},
		})
	}

	// Check chains
	newChains := ge.Milestones.CheckChains()
	for _, chain := range newChains {
		ge.addLog("success", fmt.Sprintf("Chain complete: %s! Title: %s", chain.Name, chain.Title))
		// Inject speed boost event
		ge.Events.InjectEvent(ActiveEvent{
			Key:       chain.Key + "_boost",
			Name:      chain.Name + " Speed Boost",
			TicksLeft: chain.BoostDuration,
			Effects: []config.Effect{
				{Type: "tick_speed", Target: "tick_speed", Value: chain.BoostValue},
			},
		})
		// Publish chain event
		ge.Bus.Publish(EventData{
			Type: EventChainCompleted,
			Payload: map[string]interface{}{
				"name":  chain.Name,
				"key":   chain.Key,
				"title": chain.Title,
			},
		})
	}

	// Recalculate title
	ge.Milestones.recalculateTitle()
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
	for res, rate := range ge.Buildings.WorkerScaledProduction(ge.Workers.GetAssignedCount) {
		moraleRate := rate * mMult
		r := ge.Resources.resources[res]
		if r != nil {
			r.Rate += moraleRate
			r.Breakdown.BuildingRate += moraleRate
		}
	}

	// Worker production (returns empty in Phase 6+ — contribution folded into BuildingRate)
	for res, rate := range ge.Workers.GetProductionRates() {
		r := ge.Resources.resources[res]
		if r != nil {
			r.Rate += rate
			r.Breakdown.WorkerRate += rate
		}
	}

	// Research bonuses to production rates
	researchBonuses := ge.Research.GetBonuses()
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
	costMult := clamp(1.0+r.AddTotal("build_cost"), buildCostFloor, buildCostCap)
	ge.Buildings.SetCostMultiplier(costMult)

	// Apply production_all bonus (multiplier on all positive rates).
	// Pool: research + permanent + prestige + wonders + active-event production_all.
	// Fix B: UNgated with a floor. Previously gated `if prodAllBonus > 0`, which
	// silently swallowed negative additive bonuses (e.g. the Reconstruction Effort
	// catastrophe's -0.10 production_all) whenever the player lacked ≥10% positive
	// bonuses. Now always applied as ×max(productionFloor, 1+Σ), so the debuff
	// lands but production can't drop below 10% of its pre-bonus value.
	prodAllBonus := r.AddTotal("production_all")
	prodAllFactor := math.Max(productionFloor, 1.0+prodAllBonus)
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
	// reapplyLegacyBonuses). Fix B: same ungated+floored treatment as above.
	for _, def := range ge.Resources.defs {
		bonusKey := def.Key + "_rate"
		bonus := r.AddTotal(bonusKey)
		factor := math.Max(productionFloor, 1.0+bonus)
		if factor != 1.0 {
			r := ge.Resources.resources[def.Key]
			if r != nil && r.Rate > 0 {
				r.Rate *= factor
			}
		}
	}

	// Apply gather_rate bonus to worker-generated rates. This is ADDITIVE on the
	// base worker rates — re-add the bonus portion (the production_all multiply
	// above has already touched these rates, so we add the gather delta on top of
	// the base). Fix B: ungated with the same floor. A negative gather_rate now
	// reduces worker output, but the floored factor max(productionFloor, 1+Σ)
	// means the effective worker contribution can't drop below 10% of base.
	gatherBonus := r.AddTotal("gather_rate")
	gatherDelta := math.Max(productionFloor, 1.0+gatherBonus) - 1.0
	if gatherDelta != 0 {
		for res, rate := range ge.Workers.GetProductionRates() {
			r := ge.Resources.resources[res]
			if r != nil {
				r.Rate += rate * gatherDelta
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
				tradeBonus := r.Rate * bonus
				r.Rate += tradeBonus
				r.Breakdown.TradeRate += tradeBonus
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
				r.Breakdown.ResearchRate + r.Breakdown.EventRate + r.Breakdown.TradeRate + r.Breakdown.FoodDrain
			r.Breakdown.BonusRate = r.Rate - knownComponents
		}
	}

	// Recalculate storage from buildings + research + milestones
	storageBonuses := ge.Buildings.GetStorageBonuses()
	allBonus := storageBonuses["all"]
	// Add storage bonuses from research
	allBonus += researchBonuses["all"] // storage type effects
	allBonus += permanentBonuses["all"]

	for _, def := range ge.Resources.defs {
		specific := storageBonuses[def.Key]
		specific += researchBonuses[def.Key]
		specific += permanentBonuses[def.Key]
		ge.Resources.resources[def.Key].Storage = def.BaseStorage + allBonus + specific
	}
}

// getAllResearchProductionEffects returns production effects from researched techs
func (ge *GameEngine) getAllResearchProductionEffects() []config.Effect {
	var effects []config.Effect
	allTechs := config.TechByKey()
	for _, key := range ge.Research.GetResearched() {
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
	// carryoverStarterBuildings caps a carried-over resource to roughly this many
	// of the cheapest new-age building that uses it — a small head start.
	carryoverStarterBuildings = 8
	// carryoverResidualPct is the fallback fraction kept for resources that no
	// new-age (non-wonder) building uses as a build cost. Kept at the legacy 10%
	// because this branch mostly catches food (the worker-sustain resource) —
	// cutting it harder risks a starvation spiral right at the age transition,
	// and the mass-buy problem this rebalance fixes lives entirely in the
	// build-cost cap above.
	carryoverResidualPct = 0.10
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
	ge.age = newAge
	ge.ageReady = false
	ge.Workers.SetAge(newAge)

	// Phase 7: Building lineage transformation pass.
	// Collect transforms first (safe iteration), then apply.
	type pendingTransform struct {
		oldKey, oldName, newKey, newName string
		count                            int
	}
	var transforms []pendingTransform
	for key, count := range ge.Buildings.counts {
		if count == 0 {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		if !ok || def.LineageKey == "" || def.LineageKey == "wonder" {
			continue
		}
		next := config.BuildingNextTierForAge(def.LineageKey, def.LineageTier, newAge)
		if next == nil {
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
		ge.addLog("info", fmt.Sprintf("↑ %s → %s available (×%d) — type: upgrade %s",
			t.oldName, t.newName, t.count, t.oldKey))
	}
	// Mark buildings as legacy if their lineage now has a higher-tier unlocked equivalent.
	for key, count := range ge.Buildings.counts {
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
	// Order comes from the pure config age table (no locks) — the account stays
	// config-free and ranks ages by this int rather than re-deriving order itself.
	if ge.account != nil {
		ge.account.RecordAgeReached(newAge, config.AgeByKey()[newAge].Order)
	}

	// note: Age-transition carryover model (EPIC: age-pacing economy rebalance).
	// The old flat-10% reduction still left a huge stockpile (10% of a hoard is
	// plenty to mass-buy a new age's buildings). Instead we cap each resource to
	// ~carryoverStarterBuildings of the CHEAPEST new-age building that uses it —
	// a small head start, not a fresh stockpile. Resources no new-age building
	// uses fall back to a small residual percentage. Players who didn't over-
	// accumulate keep what they had (amount below the cap is untouched).
	// Faith is excluded — it's cumulative.
	entryCosts := config.AgeEntryCosts(newAge)
	for key, r := range ge.Resources.resources {
		if key == "faith" {
			continue
		}
		if entry, ok := entryCosts[key]; ok && entry > 0 {
			capAmt := carryoverStarterBuildings * entry
			if r.Amount > capAmt {
				r.Amount = capAmt
			}
			// else: kept as-is — they didn't over-accumulate this resource.
		} else {
			r.Amount *= carryoverResidualPct
		}
	}
	ge.addLog("info", "Age transition: resources reduced to a starter head start")

	oldName := ge.progress.GetAgeName(oldAge)
	newName := ge.progress.GetAgeName(newAge)
	unlocks := ge.progress.GetUnlocks(newAge)
	ge.addLog("debug", fmt.Sprintf("Age advance: %s → %s (unlocks: %d buildings, %d resources, %d workers)",
		oldAge, newAge, len(unlocks.UnlockBuildings), len(unlocks.UnlockResources), len(unlocks.UnlockVillagers)))
	ge.addLog("success", fmt.Sprintf("Advanced from %s to %s!", oldName, newName))

	// Notify player about the wonder available in this age
	for _, bKey := range unlocks.UnlockBuildings {
		if def, ok := ge.Buildings.defs[bKey]; ok && def.Category == "wonder" {
			ge.addLog("event", fmt.Sprintf("★ Wonder available: %s — build it to unlock a permanent +0.5x speed bonus!", def.Name))
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
// and fires the epoch event roll when an epoch boundary is crossed. Must be
// called at the end of advanceAge while the engine write lock is held.
// Each epoch fires its roll at most once per civilisation cycle
// (epochEventFired prevents double-fire on load or re-entry).
func (ge *GameEngine) detectEpochTransition(newAge string) {
	newEpoch := config.EpochForAge(newAge)
	if newEpoch == ge.currentEpoch {
		return // same epoch, no transition
	}
	ge.currentEpoch = newEpoch
	ep := config.EpochByKey()[newEpoch]
	ge.addLog("event", fmt.Sprintf("[%s]✦ The %s Dawns — %s[-]", ep.Color, ep.Name, ep.Description))
	ge.Bus.Publish(EventData{
		Type: EventEpochAdvanced,
		Payload: map[string]interface{}{
			"epoch_key":  newEpoch,
			"epoch_name": ep.Name,
			"epoch_icon": ep.Icon,
		},
	})
	ge.rollEpochEvent(newEpoch)
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
	def, ok := config.AwakeningForAge(newAge)
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

	ep := config.EpochByKey()[def.EpochKey]
	// One log line per awakening — the pivotal "new era" beat. Coloured by the epoch
	// so the awakening visually belongs to the era it ushers in.
	ge.addLog("event", fmt.Sprintf("[%s]✦ Awakening: %s — %s[-]", ep.Color, def.Name, def.FlavorText))

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
//   - Faith fill % gates good-event probability: <25% → 40%, >75% → 60%, else 50%.
//   - On a bad roll, a further 30% chance escalates to a catastrophe (modal prompt).
//   - Otherwise a challenging (non-catastrophe) bad event is applied immediately.
//
// Must be called under engine write lock.
func (ge *GameEngine) rollEpochEvent(epochKey string) {
	// Prevent double-fire per epoch
	if ge.epochEventFired[epochKey] {
		return
	}
	ge.epochEventFired[epochKey] = true

	faithStorage := ge.Resources.GetStorage("faith")
	goodChance := 0.50
	if faithStorage > 0 {
		faithPct := ge.Resources.Get("faith") / faithStorage
		if faithPct < 0.25 {
			goodChance = 0.40
		} else if faithPct > 0.75 {
			goodChance = 0.60
		}
	}

	if rand.Float64() < goodChance {
		ge.rollGoodEpochEvent()
	} else {
		if rand.Float64() < 0.30 {
			// Catastrophe
			ge.pendingCatastrophe = epochKey
			ep := config.EpochByKey()[epochKey]
			ge.addLog("warning", fmt.Sprintf("☄ A great catastrophe threatens the %s — prepare yourself.", ep.Name))
			record := EpochEventRecord{
				EpochKey: epochKey, EpochName: ep.Name,
				EventKey: ep.CatastropheKey, EventName: "Catastrophe", EventType: "catastrophe",
				Tick: ge.tick,
			}
			ge.epochEventHistory = append(ge.epochEventHistory, record)
		} else {
			ge.rollChallengingEpochEvent(epochKey)
		}
	}
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
		if culturePct > 0.75 && rand.Float64() < 0.15 {
			tier = "legendary"
		} else if culturePct > 0.40 {
			tier = "major"
		}
	}

	pool := config.GoodEpochEvents()
	var eligible []config.EpochEventDef
	for _, ev := range pool {
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
	ev := eligible[rand.Intn(len(eligible))]
	ge.applyGoodEpochEvent(ev)

	ep := config.EpochByKey()[ge.currentEpoch]
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
	pool := config.ChallengingEpochEvents()
	if len(pool) == 0 {
		return
	}
	ev := pool[rand.Intn(len(pool))]
	ge.applyChallengingEpochEvent(ev, epochKey)

	ep := config.EpochByKey()[epochKey]
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

// applyGoodEpochEvent applies the effects of a good epoch transition event.
func (ge *GameEngine) applyGoodEpochEvent(ev config.EpochEventDef) {
	ge.addLog("success", fmt.Sprintf("✦ %s — %s", ev.Name, ev.FlavorText))
	ageOrder := ge.progress.GetAgeOrder()
	switch ev.Key {
	case "age_of_plenty":
		// ×2 all production for Duration ticks via production_all effect
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_age_of_plenty",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production_all", Value: 1.0}},
		})
	case "population_surge":
		// +15% workers across all domains, instant
		ge.Workers.AddPctAll(0.15)
	case "ancient_cache":
		// Fill 40% of each resource's storage cap
		for _, def := range ge.Resources.defs {
			cap := ge.Resources.GetStorage(def.Key)
			if cap > 0 && ge.Resources.IsUnlocked(def.Key) {
				ge.Resources.Add(def.Key, cap*0.40)
			}
		}
	case "trade_winds":
		// Gold ×2 production for Duration ticks
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_trade_winds",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: "gold", Value: 5.0}},
		})
	case "cultural_festival":
		// Instant culture +30%, faith +20%; timed production boost
		ge.Resources.Add("culture", ge.Resources.Get("culture")*0.30)
		ge.Resources.Add("faith", ge.Resources.Get("faith")*0.20)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_cultural_festival",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects: []config.Effect{
				{Type: "production", Target: "culture", Value: 1.0},
				{Type: "production", Target: "faith", Value: 1.0},
			},
		})
	case "grand_discovery":
		// Complete 3 free techs from current age
		completed := ge.Research.ForceCompleteN(3, ge.age, ageOrder)
		for _, key := range completed {
			def := config.TechByKey()[key]
			ge.addLog("success", fmt.Sprintf("  → Free tech: %s", def.Name))
		}
	case "worker_innovation":
		// Permanent +10% production_all
		ge.permanentBonuses["production_all"] += 0.10
		ge.addLog("success", "  → All production permanently +10%")
	case "architects_gift":
		// 10 free buildings of the most common built non-wonder type
		bestKey := ""
		bestCount := 0
		for key, count := range ge.Buildings.counts {
			if def, ok := ge.Buildings.defs[key]; ok && def.Category != "wonder" && count > bestCount {
				bestKey = key
				bestCount = count
			}
		}
		if bestKey != "" {
			ge.Buildings.counts[bestKey] += 10
			def := ge.Buildings.defs[bestKey]
			ge.addLog("success", fmt.Sprintf("  → 10 free %s", def.Name))
		}
	case "peaceful_century":
		// +20% all production for Duration ticks
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_peaceful_century",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production_all", Value: 0.20}},
		})
	case "epoch_blessing":
		// Permanent +15% production_all; recorded as a golden age
		ge.permanentBonuses["production_all"] += 0.15
		ge.addLog("success", "  → Epoch Blessing: all production permanently +15%")
	}
}

// applyChallengingEpochEvent applies a challenging (non-catastrophe) bad epoch event.
func (ge *GameEngine) applyChallengingEpochEvent(ev config.EpochEventDef, epochKey string) {
	ge.addLog("warning", fmt.Sprintf("⚠ %s — %s", ev.Name, ev.FlavorText))
	switch ev.Key {
	case "the_famine":
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_famine",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: "food", Value: -3.0}},
		})
	case "merchant_betrayal":
		ge.Resources.Remove("gold", ge.Resources.Get("gold")*0.50)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_merchant_betrayal",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: "gold", Value: -2.0}},
		})
	case "the_great_fire":
		destroyed := ge.Buildings.DestroyRandom(8)
		for _, desc := range destroyed {
			ge.addLog("warning", fmt.Sprintf("  → Destroyed: %s", desc))
		}
	case "epidemic":
		ge.Workers.RemovePct(0.20)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_epidemic",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: "food", Value: -1.5}},
		})
	case "resource_drought":
		// Debuff epoch's primary resource
		primaryRes := "wood" // fallback
		if ep, ok := config.EpochByKey()[epochKey]; ok {
			primaryRes = ep.PrimaryResource
		}
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_resource_drought",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: primaryRes, Value: -3.0}},
		})
	case "political_instability":
		ge.Resources.Remove("faith", ge.Resources.Get("faith")*0.60)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_political_instability",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects: []config.Effect{
				{Type: "production", Target: "knowledge", Value: -2.0},
			},
		})
	case "economic_crash":
		ge.Resources.Remove("gold", ge.Resources.Get("gold")*0.50)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_economic_crash",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: "gold", Value: -3.0}},
		})
	case "the_dark_age":
		ge.Research.CancelResearch()
		ge.Resources.Remove("knowledge", ge.Resources.Get("knowledge")*0.80)
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "epoch_dark_age",
			Name:      ev.Name,
			TicksLeft: ev.Duration,
			Effects:   []config.Effect{{Type: "production", Target: "knowledge", Value: -3.0}},
		})
	}
}

// InvokeCatastrophe voluntarily triggers the catastrophe for the current epoch.
// Returns an error if a catastrophe has already been invoked this epoch (random or voluntary).
func (ge *GameEngine) InvokeCatastrophe() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.epochEventFired[ge.currentEpoch] {
		return fmt.Errorf("a catastrophe event has already occurred this epoch (%s)", ge.currentEpoch)
	}
	if ge.pendingCatastrophe != "" {
		return fmt.Errorf("a catastrophe is already pending")
	}
	ge.epochEventFired[ge.currentEpoch] = true
	ge.pendingCatastrophe = ge.currentEpoch
	ep := config.EpochByKey()[ge.currentEpoch]
	ge.addLog("warning", fmt.Sprintf("☄ Voluntary catastrophe invoked for the %s.", ep.Name))
	record := EpochEventRecord{
		EpochKey: ge.currentEpoch, EpochName: ep.Name,
		EventKey: ep.CatastropheKey, EventName: "Voluntary Catastrophe", EventType: "catastrophe",
		Tick: ge.tick,
	}
	ge.epochEventHistory = append(ge.epochEventHistory, record)
	return nil
}

// DeferCatastrophe signals the player's intent to decide later.
// The pendingCatastrophe remains set; the UI handles hiding the modal.
// This method exists for command-based invocation (no engine state change needed).
func (ge *GameEngine) DeferCatastrophe() {
	// No-op on the engine side — pendingCatastrophe stays set.
	// The UI is responsible for not re-showing the modal until next session refresh.
}

// reapplyLegacyBonuses restores all Succumb legacy rate bonuses into
// permanentBonuses after a reset. Must be called whenever permanentBonuses is
// cleared (prestige or succumb resets) so cross-run bonuses are not lost.
func (ge *GameEngine) reapplyLegacyBonuses() {
	for epochKey := range ge.legacyBonuses {
		for res, mult := range config.LegacyBonusForEpoch(epochKey) {
			ge.permanentBonuses[res+"_rate"] += mult
		}
	}
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
		Cost:          ge.festivalCost(),
		Culture:       ge.Resources.Get("culture"),
		BuffPercent:   festivalBuffPercent,
		BuffTicks:     festivalBuffTicks,
		CooldownTicks: festivalCooldownTicks,
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
		return fmt.Errorf("festival is on cooldown (%d ticks remaining)", ge.festivalReadyTick-ge.tick)
	}
	cost := ge.festivalCost()
	have := ge.Resources.Get("culture")
	if have < cost {
		return fmt.Errorf("not enough culture for a festival: need %.0f, have %.0f", cost, have)
	}
	ge.Resources.Remove("culture", cost)
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "cultural_festival",
		Name:      "Cultural Festival",
		TicksLeft: festivalBuffTicks,
		Effects:   []config.Effect{{Type: "production_all", Value: festivalBuffPercent}},
	})
	ge.festivalReadyTick = ge.tick + festivalCooldownTicks
	ge.addLog("success", fmt.Sprintf("Held a cultural festival — spent %.0f culture for +%.0f%% production (%d ticks).",
		cost, festivalBuffPercent*100, festivalBuffTicks))
	return nil
}

// getWonderBonuses returns a map of bonus-type effects from all currently built
// wonders and cultural monuments (those with count > 0). Effects with Type "bonus"
// represent percentage multipliers (e.g. production_all, knowledge_rate,
// expedition_reward) rather than flat resource production rates. The returned map
// is keyed by Target and holds the summed Value across all built wonders and monuments.
// Must be called with the engine lock held.
func (ge *GameEngine) getWonderBonuses() map[string]float64 {
	out := make(map[string]float64)
	for key, count := range ge.Buildings.counts {
		if count == 0 {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		if !ok || (def.Category != "wonder" && def.Category != "monument") {
			continue
		}
		for _, eff := range def.Effects {
			if eff.Type == "bonus" {
				out[eff.Target] += eff.Value * float64(count)
			}
		}
	}
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
	r.AddAll(ge.eventModifiers())
	r.AddAll(ge.moraleModifiers())
	r.AddAll(ge.diplomacyModifiers())
	return r
}

// Endure executes the Endure consequences for the pending catastrophe:
//   - 20% of buildings randomly destroyed
//   - All resources drop to 15% of current stored amount
//   - 25% of workers removed
//   - Lasting timed debuffs injected (building costs conceptually +20%, food drain +10%)
//   - Permanent rewards: Survived marker, unique titled logged in civilization history
func (ge *GameEngine) Endure() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingCatastrophe == "" {
		return fmt.Errorf("no pending catastrophe to endure")
	}
	epochKey := ge.pendingCatastrophe
	ge.pendingCatastrophe = ""
	ge.survivedEpochs[epochKey] = true

	catName, catFlavor := config.CatastropheInfo(epochKey)

	// 20% of buildings destroyed
	totalBuilt := 0
	for _, c := range ge.Buildings.counts {
		totalBuilt += c
	}
	destroyCount := totalBuilt / 5
	if destroyCount < 1 && totalBuilt > 0 {
		destroyCount = 1
	}
	destroyed := ge.Buildings.DestroyRandom(destroyCount)

	// Resources → 15% of current
	for key, r := range ge.Resources.resources {
		if r != nil && ge.Resources.IsUnlocked(key) {
			r.Amount *= 0.15
		}
	}

	// -25% workers
	ge.Workers.RemovePct(0.25)

	// Timed debuffs: worker food drain +10% for 216 ticks, production -10% for 216 ticks
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "endure_reconstruction",
		Name:      "Reconstruction Effort",
		TicksLeft: 216,
		Effects: []config.Effect{
			{Type: "production_all", Value: -0.10},
		},
	})

	// Log consequences
	ge.addLog("warning", fmt.Sprintf("☄ ENDURE: %s — %s", catName, catFlavor))
	ge.addLog("warning", fmt.Sprintf("  Buildings destroyed: %d", destroyCount))
	for _, desc := range destroyed {
		ge.addLog("warning", fmt.Sprintf("  → %s lost", desc))
	}
	ge.addLog("warning", "  All resources reduced to 15% of stored amounts.")
	ge.addLog("warning", "  25% of workers lost.")
	ge.addLog("info", "  Timed: production -10% for 216 ticks (reconstruction period).")
	ge.addLog("success", fmt.Sprintf("  ✦ Survived marker earned for %s badge.", config.EpochByKey()[epochKey].Name))

	// Civilization history entry
	histEntry := fmt.Sprintf("Tick %d — Endured %s (%s). %d buildings lost.", ge.tick, catName, config.EpochByKey()[epochKey].Name, destroyCount)
	ge.catastropheHistory = append(ge.catastropheHistory, histEntry)

	// Catastrophe survival hurts morale
	ge.applyMorale(-0.10)

	return nil
}

// Succumb executes the Succumb consequences for the pending catastrophe:
//   - Generates 8 ruins from current buildings (persist into next run)
//   - Awards epoch legacy bonus (permanent production multiplier)
//   - Awards Ancient Knowledge (+25% research speed, permanent)
//   - Records civilization history entry
//   - Full reset (all buildings, resources, workers, etc.)
//   - Restores ruins and legacy bonuses into the fresh civilization
func (ge *GameEngine) Succumb() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingCatastrophe == "" {
		return fmt.Errorf("no pending catastrophe to succumb to")
	}
	epochKey := ge.pendingCatastrophe
	catName, _ := config.CatastropheInfo(epochKey)
	ep := config.EpochByKey()[epochKey]

	// Civilization history entry (before reset clears tick/age)
	histEntry := fmt.Sprintf("Tick %d — Succumbed to %s (%s). Civilization reset. Legacy bonus earned.", ge.tick, catName, ep.Name)
	ge.catastropheHistory = append(ge.catastropheHistory, histEntry)

	// Generate 8 ruins from current buildings
	ge.Buildings.GenerateRuins(8)
	savedRuins := ge.Buildings.GetAllRuins()

	// Award legacy bonus for this epoch
	ge.legacyBonuses[epochKey] = true
	savedLegacy := copyBoolMap(ge.legacyBonuses)
	savedCatHistory := append([]string(nil), ge.catastropheHistory...)
	savedEpochHistory := append([]EpochEventRecord(nil), ge.epochEventHistory...)

	// Preserve cross-run state
	savedPrestige := ge.Prestige

	// Full reset — Bus intentionally kept so dashboard subscriptions survive.
	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManager()
	ge.Buildings = NewBuildingManager()
	ge.Workers = NewWorkerManager()
	ge.Research = NewResearchManager()
	ge.Military = NewMilitaryManager()
	ge.Events = NewEventManager()
	ge.Milestones = NewMilestoneManager()
	ge.Trade = NewTradeManager()
	ge.Diplomacy = NewDiplomacyManager()
	ge.Stats = NewGameStats()
	ge.permanentBonuses = make(map[string]float64)
	ge.buildQueue = nil
	ge.log = nil
	ge.speedMultiplier = 1.0
	ge.tickSpeedBonus = 0
	ge.ageReady = false
	ge.starvationTicks = 0
	ge.currentEpoch = config.EpochForAge("primitive_age")
	ge.epochEventFired = make(map[string]bool)
	ge.awakeningsFired = make(map[string]bool)
	ge.survivedEpochs = make(map[string]bool)
	ge.pendingCatastrophe = ""
	ge.morale = 0.50
	ge.lowMoraleWarned = false
	// Fresh run after the fall: eligible to roll a new Ancient Memory.
	ge.ancientMemoryUsed = false
	ge.pendingMemoryTech = ""

	// Restore persistent cross-run state
	ge.Prestige = savedPrestige
	ge.Buildings.LoadRuins(savedRuins)
	ge.legacyBonuses = savedLegacy
	ge.catastropheHistory = savedCatHistory
	ge.epochEventHistory = savedEpochHistory

	// Ancient Knowledge: +25% research speed (permanent bonus)
	ge.permanentBonuses["research_speed"] += 0.25

	// Apply all active legacy bonuses
	ge.reapplyLegacyBonuses()

	// Apply age unlocks and starting resources
	ge.applyAgeUnlocks("primitive_age")
	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)
	for res, amount := range ge.Prestige.GetStartingResources() {
		ge.Resources.Add(res, amount)
	}

	ge.recalculateTickSpeed()

	// Welcome-back log
	ge.addLog("event", fmt.Sprintf("☄ %s — civilization has fallen. A new dawn.", catName))
	ge.addLog("success", fmt.Sprintf("Legacy Bonus: %s production permanently boosted.", ep.Name))
	ge.addLog("success", "Ancient Knowledge: research speed +25% (permanent).")
	if len(savedRuins) > 0 {
		ge.addLog("info", fmt.Sprintf("%d ruin type(s) from the fallen civilization carry forward.", len(savedRuins)))
	}
	ge.addLog("info", "Type [cyan]help[-] to rebuild.")

	// Roll for an Ancient Memory cache (only when this account has prestiged before;
	// a first-ever Succumb with no prestige history offers nothing — see the gate).
	ge.maybeOfferAncientMemory()

	return nil
}

// processBuildQueue advances construction on queued buildings
func (ge *GameEngine) processBuildQueue() {
	var remaining []BuildQueueItem
	for _, item := range ge.buildQueue {
		item.TicksLeft--
		if item.TicksLeft <= 0 {
			ge.Buildings.counts[item.BuildingKey]++
			def := ge.Buildings.defs[item.BuildingKey]
			ge.addLog("debug", fmt.Sprintf("Build complete: %s (count now %d)", def.Name, ge.Buildings.GetCount(item.BuildingKey)))
			ge.addLog("success", fmt.Sprintf("%s completed! (#%d)", def.Name, ge.Buildings.GetCount(item.BuildingKey)))
			ge.Stats.RecordBuild()
			ge.Bus.Publish(EventData{
				Type:    EventBuildingBuilt,
				Payload: map[string]interface{}{"building": item.BuildingKey},
			})
		} else {
			def := ge.Buildings.defs[item.BuildingKey]
			ge.addLog("debug", fmt.Sprintf("Build queue: %s %d/%d ticks", def.Name, item.TotalTicks-item.TicksLeft, item.TotalTicks))
			remaining = append(remaining, item)
		}
	}
	ge.buildQueue = remaining
}

// --- Public API for commands ---

// AdvanceAge manually advances to the next age if requirements are met.
func (ge *GameEngine) AdvanceAge() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if !ge.ageReady {
		nextAge := ge.progress.CheckAdvancement(ge.age, ge.Resources, ge.Buildings)
		if nextAge == "" {
			// Check if wonder is the only blocker
			wonderKey := ge.progress.WonderForAge(ge.age)
			if wonderKey != "" && ge.Buildings.GetCount(wonderKey) < 1 {
				if def, ok := ge.Buildings.defs[wonderKey]; ok {
					return fmt.Errorf("you must complete the %s wonder before advancing — use 'wonder collect' then 'build %s'", def.Name, wonderKey)
				}
			}
			return fmt.Errorf("age requirements not met yet — check the Stats tab for what's needed")
		}
	}
	nextAge := ge.progress.GetNextAge(ge.age)
	if nextAge == "" {
		return fmt.Errorf("you are already at the final age")
	}
	ge.advanceAge(nextAge)
	return nil
}

// pastMedievalForGather reports whether the given age is strictly later than the
// Medieval Age in the canonical age order. Used to gate hand-gathering. It is
// pure (relies only on config.AgeOrder) and acquires no locks, so it is safe to
// call while the engine write lock is held. Fails safe: if either age key is
// absent from the order, it returns false (gathering allowed) rather than panic.
func pastMedievalForGather(age string) bool {
	order := config.AgeOrder()
	curIdx, medievalIdx := -1, -1
	for i, key := range order {
		switch key {
		case age:
			curIdx = i
		case "medieval_age":
			medievalIdx = i
		}
	}
	if curIdx == -1 || medievalIdx == -1 {
		return false
	}
	return curIdx > medievalIdx
}

// GatherResource manually gathers a resource
func (ge *GameEngine) GatherResource(resource string, amount float64) (float64, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	// Hand-gathering is only practical through the Medieval Age. Past it, the
	// economy is expected to run on buildings and workers. Lock is held here, so
	// we use the ge.age field and pure config.AgeOrder() — no GetState().
	if pastMedievalForGather(ge.age) {
		return 0, fmt.Errorf("gathering by hand is no longer practical past the Medieval Age — your economy runs on buildings and workers now")
	}

	if !ge.Resources.IsUnlocked(resource) {
		return 0, fmt.Errorf("resource '%s' is not yet unlocked", resource)
	}
	actual := ge.Resources.Add(resource, amount)
	ge.Stats.RecordGather(resource, amount)
	ge.addLog("debug", fmt.Sprintf("Gather: %s +%.1f (total: %.1f)", resource, amount, actual))
	ge.addLog("success", fmt.Sprintf("Gathered %.0f %s", amount, resource))
	return actual, nil
}

// BuildBuilding constructs a building (instant or queued)
// BankWonderResource deposits resources from player storage into a wonder's bank.
func (ge *GameEngine) BankWonderResource(wonderKey, resource string, amount float64) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	deposited, err := ge.Buildings.BankResource(wonderKey, resource, amount, ge.Resources)
	if err != nil {
		return err
	}

	def := ge.Buildings.defs[wonderKey]
	banked := ge.Buildings.wonderBanks[wonderKey][resource]
	need := def.BaseCost[resource]
	ge.addLog("info", fmt.Sprintf("Banked %.0f %s toward %s (%.0f / %.0f)", deposited, resource, def.Name, banked, need))

	if ge.Buildings.IsWonderBankFull(wonderKey) {
		ge.addLog("success", fmt.Sprintf("%s bank is full! Type 'build %s' to begin construction.", def.Name, wonderKey))
	}
	return nil
}

func (ge *GameEngine) BuildBuilding(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, exists := ge.Buildings.defs[key]
	if !exists {
		// Unknown building key — suggest closest match
		if suggestion := ge.Buildings.SuggestKey(key); suggestion != "" {
			return fmt.Errorf("unknown building '%s' — did you mean '%s'?", key, suggestion)
		}
		return fmt.Errorf("unknown building '%s'. Type 'build' to see available buildings.", key)
	}
	if !ge.Buildings.IsUnlocked(key) {
		return fmt.Errorf("building '%s' is not yet unlocked", def.Name)
	}
	// Age lock: only allow building structures that belong to the current age.
	// Storage and wonder categories with no RequiredAge (empty string) are exempt.
	// Wonders always match because their RequiredAge equals the age they unlock in,
	// and the player can only be in that age when they attempt to build the wonder.
	if def.RequiredAge != "" && def.RequiredAge != ge.age {
		return fmt.Errorf("%s belongs to a previous age — use 'upgrade' to advance your buildings", def.Name)
	}
	if def.MaxCount > 0 {
		inQueue := ge.Buildings.GetQueueCount(key, ge.buildQueue)
		if ge.Buildings.GetCount(key)+inQueue >= def.MaxCount {
			return fmt.Errorf("%s is at max count (%d)", def.Name, def.MaxCount)
		}
	}

	// Check if already building this (for unique buildings)
	for _, item := range ge.buildQueue {
		if item.BuildingKey == key && def.MaxCount > 0 {
			return fmt.Errorf("%s is already under construction (%d ticks left)", def.Name, item.TicksLeft)
		}
	}

	if DevGodMode {
		// godmode: skip all cost/bank checks, build instantly below
	} else if def.Category == "wonder" {
		if !ge.Buildings.IsWonderBankFull(key) {
			return fmt.Errorf("%s bank is not full — use 'wonder collect <resource> <amount>' to bank resources first", def.Name)
		}
		// Resources were already deducted when banked; nothing to pay here.
	} else {
		// Use queue-aware cost so that items already in the build queue are
		// factored into the cost curve (fixes queue-blindness exploit).
		cost, _ := ge.Buildings.BuildBatchCost(key, 1, ge.buildQueue)
		if !ge.Resources.Pay(cost) {
			return fmt.Errorf("cannot afford %s (need: %s)", def.Name, formatCost(cost))
		}
	}

	ge.addLog("debug", fmt.Sprintf("Build start: %s", def.Name))
	if !DevGodMode && def.BuildTicks > 0 {
		// Queue for construction
		ge.buildQueue = append(ge.buildQueue, BuildQueueItem{
			BuildingKey: key,
			TicksLeft:   def.BuildTicks,
			TotalTicks:  def.BuildTicks,
		})
		ge.addLog("info", fmt.Sprintf("Started building %s (%d ticks)", def.Name, def.BuildTicks))
	} else {
		// Instant build
		ge.Buildings.counts[key]++
		ge.Stats.RecordBuild()
		ge.recalculateRates()
		ge.addLog("success", fmt.Sprintf("Built %s (#%d)", def.Name, ge.Buildings.GetCount(key)))
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
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def, exists := ge.Buildings.defs[key]
	if !exists {
		if suggestion := ge.Buildings.SuggestKey(key); suggestion != "" {
			return 0, fmt.Errorf("unknown building '%s' — did you mean '%s'?", key, suggestion)
		}
		return 0, fmt.Errorf("unknown building '%s'. Type 'build' to see available buildings.", key)
	}
	if !ge.Buildings.IsUnlocked(key) {
		return 0, fmt.Errorf("building '%s' is not yet unlocked", def.Name)
	}
	// Age lock: only allow building structures that belong to the current age.
	if def.RequiredAge != "" && def.RequiredAge != ge.age {
		return 0, fmt.Errorf("%s belongs to a previous age — use 'upgrade' to advance your buildings", def.Name)
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
			ge.buildQueue = append(ge.buildQueue, BuildQueueItem{
				BuildingKey: key,
				TicksLeft:   def.BuildTicks,
				TotalTicks:  def.BuildTicks,
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
				return 0, fmt.Errorf("%s is at max count (%d)", def.Name, def.MaxCount)
			}
		}
		unitCost, _ := ge.Buildings.BuildBatchCost(key, 1, ge.buildQueue)
		return 0, fmt.Errorf("cannot afford %s (need: %s)", def.Name, formatCost(unitCost))
	}

	if def.BuildTicks > 0 {
		ge.addLog("info", fmt.Sprintf("Queued %d %s for construction", built, def.Name))
	} else {
		ge.recalculateRates()
		ge.addLog("success", fmt.Sprintf("Built %d %s (total: %d)", built, def.Name, ge.Buildings.GetCount(key)))
	}
	return built, nil
}

// RecruitMax recruits as many workers as possible up to the pop cap
func (ge *GameEngine) RecruitMax(vType string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if !ge.Workers.IsUnlocked(vType) {
		return 0, fmt.Errorf("worker type '%s' is not yet unlocked", vType)
	}

	popCap := ge.Buildings.GetPopCapacity()
	popCap += int(ge.Research.GetBonus("population") + ge.permanentBonuses["population"] + ge.Prestige.GetBonuses()["population"])

	available := popCap - ge.Workers.TotalPop()
	if available <= 0 {
		return 0, fmt.Errorf("population cap reached (%d/%d)", ge.Workers.TotalPop(), popCap)
	}

	if !ge.Workers.Recruit(vType, available, popCap) {
		return 0, fmt.Errorf("cannot recruit %s(s)", vType)
	}
	ge.Stats.RecordRecruit(available)
	ge.addLog("info", fmt.Sprintf("Recruited %d worker(s) (pop: %d/%d)", available, ge.Workers.TotalPop(), popCap))
	return available, nil
}

// RecruitWorker recruits workers
func (ge *GameEngine) RecruitWorker(vType string, count int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	popCap := ge.Buildings.GetPopCapacity()
	// Add population capacity from research/milestones/prestige
	popCap += int(ge.Research.GetBonus("population") + ge.permanentBonuses["population"] + ge.Prestige.GetBonuses()["population"])

	if !ge.Workers.Recruit(vType, count, popCap) {
		totalPop := ge.Workers.TotalPop()
		if !ge.Workers.IsUnlocked(vType) {
			return fmt.Errorf("worker type '%s' is not yet unlocked", vType)
		}
		return fmt.Errorf("cannot recruit %d %s(s) (pop: %d/%d)", count, vType, totalPop, popCap)
	}
	ge.Stats.RecordRecruit(count)
	ge.addLog("debug", fmt.Sprintf("Recruit: %d worker(s) (pop: %d/%d)", count, ge.Workers.TotalPop(), popCap))
	ge.addLog("info", fmt.Sprintf("Recruited %d worker(s)", count))
	return nil
}

// AssignWorker assigns workers to a building.
// Any worker can be assigned to any building with WorkerCapacity > 0.
func (ge *GameEngine) AssignWorker(buildingKey string, count int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.Buildings.GetCount(buildingKey) == 0 {
		return fmt.Errorf("no %s built yet — build one first", buildingKey)
	}
	byKey := config.BuildingByKey()
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerCapacity == 0 {
		return fmt.Errorf("building %s does not accept workers", buildingKey)
	}
	// Enforce capacity cap
	totalCap := def.WorkerCapacity * ge.Buildings.GetCount(buildingKey)
	alreadyAssigned := ge.Workers.GetAssignedCount("worker", buildingKey)
	available := totalCap - alreadyAssigned
	if available <= 0 {
		return fmt.Errorf("all %d worker slot(s) for %s are full", totalCap, buildingKey)
	}
	if count > available {
		return fmt.Errorf("only %d worker slot(s) available for %s (%d/%d filled)", available, buildingKey, alreadyAssigned, totalCap)
	}
	if !ge.Workers.Assign("worker", buildingKey, count) {
		idle := ge.Workers.IdleCount("worker")
		return fmt.Errorf("cannot assign %d workers to %s (idle: %d)", count, buildingKey, idle)
	}
	ge.recalculateRates()
	ge.addLog("debug", fmt.Sprintf("Assign: %d → %s", count, buildingKey))
	ge.addLog("info", fmt.Sprintf("Assigned %d worker(s) to %s", count, buildingKey))
	return nil
}

// AssignAll assigns all idle workers to a building.
// Any worker can be assigned to any building with WorkerCapacity > 0.
func (ge *GameEngine) AssignAll(buildingKey string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.Buildings.GetCount(buildingKey) == 0 {
		return 0, fmt.Errorf("no %s built yet — build one first", buildingKey)
	}
	byKey := config.BuildingByKey()
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerCapacity == 0 {
		return 0, fmt.Errorf("building %s does not accept workers", buildingKey)
	}
	// Cap at available capacity
	toAssign := ge.Workers.IdleCount("worker")
	if toAssign <= 0 {
		return 0, fmt.Errorf("no idle workers to assign")
	}
	totalCap := def.WorkerCapacity * ge.Buildings.GetCount(buildingKey)
	alreadyAssigned := ge.Workers.GetAssignedCount("worker", buildingKey)
	available := totalCap - alreadyAssigned
	if available <= 0 {
		return 0, fmt.Errorf("all %d worker slot(s) for %s are full", totalCap, buildingKey)
	}
	if toAssign > available {
		toAssign = available
	}
	if !ge.Workers.Assign("worker", buildingKey, toAssign) {
		return 0, fmt.Errorf("cannot assign workers to %s", buildingKey)
	}
	ge.recalculateRates()
	ge.addLog("info", fmt.Sprintf("Assigned all %d worker(s) to %s", toAssign, buildingKey))
	return toAssign, nil
}

// UnassignAll removes all workers from a building.
func (ge *GameEngine) UnassignAll(buildingKey string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	byKey := config.BuildingByKey()
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerCapacity == 0 {
		return 0, fmt.Errorf("building %s does not accept workers", buildingKey)
	}
	assigned := ge.Workers.GetAssignedCount("worker", buildingKey)
	if assigned <= 0 {
		return 0, fmt.Errorf("no workers assigned to %s", buildingKey)
	}
	if !ge.Workers.Unassign("worker", buildingKey, assigned) {
		return 0, fmt.Errorf("cannot unassign workers from %s", buildingKey)
	}
	ge.recalculateRates()
	ge.addLog("info", fmt.Sprintf("Unassigned all %d worker(s) from %s", assigned, buildingKey))
	return assigned, nil
}

// UnassignWorker removes a specific number of workers from a building.
func (ge *GameEngine) UnassignWorker(buildingKey string, count int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	byKey := config.BuildingByKey()
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerCapacity == 0 {
		return fmt.Errorf("building %s does not accept workers", buildingKey)
	}
	if !ge.Workers.Unassign("worker", buildingKey, count) {
		return fmt.Errorf("cannot unassign %d workers from %s", count, buildingKey)
	}
	ge.recalculateRates()
	ge.addLog("debug", fmt.Sprintf("Unassign: %d ← %s", count, buildingKey))
	ge.addLog("info", fmt.Sprintf("Unassigned %d worker(s) from %s", count, buildingKey))
	return nil
}

// DismissWorkers removes workers from a building and from the population pool entirely.
func (ge *GameEngine) DismissWorkers(buildingKey string, count int, all bool) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if all {
		count = ge.Workers.GetAssignedCount("worker", buildingKey)
	}
	if count <= 0 {
		return fmt.Errorf("no workers assigned to %s", buildingKey)
	}
	dismissed := ge.Workers.Dismiss(buildingKey, count)
	if dismissed == 0 {
		return fmt.Errorf("no workers assigned to %s", buildingKey)
	}
	byKey := config.BuildingByKey()
	def, _ := byKey[buildingKey]
	ge.recalculateRates()
	ge.addLog("info", fmt.Sprintf("Dismissed %d workers from %s (pop: %d)", dismissed, def.Name, ge.Workers.TotalPop()))
	return nil
}

// formatResourceMap formats a map[string]float64 as "key1 45, key2 20" sorted by key.
func formatResourceMap(m map[string]float64) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %.0f", k, m[k]))
	}
	return strings.Join(parts, ", ")
}

// SellBuilding removes n copies of a built building, refunds 50% of cost,
// and unassigns any workers that were in the sold slots.
func (ge *GameEngine) SellBuilding(key string, n int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.age == "primitive_age" {
		return fmt.Errorf("sell is not available in the primitive age")
	}

	byKey := config.BuildingByKey()
	def, ok := byKey[key]
	if !ok {
		if suggestion := ge.Buildings.SuggestKey(key); suggestion != "" {
			return fmt.Errorf("unknown building '%s' — did you mean '%s'?", key, suggestion)
		}
		return fmt.Errorf("unknown building '%s'", key)
	}

	if def.Category == "wonder" {
		return fmt.Errorf("wonders cannot be sold")
	}

	current := ge.Buildings.GetCount(key)
	if current == 0 {
		return fmt.Errorf("no %s built", def.Name)
	}

	if n > current {
		n = current
	}

	// Check build queue — reject if any copy of this building is queued
	for _, item := range ge.buildQueue {
		if item.BuildingKey == key {
			return fmt.Errorf("cannot sell a building that is under construction")
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

	// Add refund to resources
	for res, amount := range refund {
		ge.Resources.Add(res, amount)
	}

	ge.recalculateRates()
	ge.addLog("info", fmt.Sprintf("Sold %d %s — returned: %s", n, def.Name, formatResourceMap(refund)))
	return nil
}

// StartResearch begins researching a technology
func (ge *GameEngine) StartResearch(techKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	knowledge := ge.Resources.Get("knowledge")

	// Combine research_speed from all sources: techs + permanent bonuses + prestige.
	// This must be done before StartResearch so the combined value reduces tick count.
	combinedResearchSpeed := ge.Research.GetBonus("research_speed") +
		ge.permanentBonuses["research_speed"] +
		ge.Prestige.GetBonuses()["research_speed"]
	if err := ge.Research.StartResearchWithSpeed(techKey, ge.age, ageOrder, knowledge, combinedResearchSpeed); err != nil {
		return err
	}

	// Pay knowledge cost (waived in godmode)
	def := config.TechByKey()[techKey]
	if !DevGodMode {
		ge.Resources.Remove("knowledge", def.Cost)
	}
	if DevGodMode {
		// complete immediately
		ge.Research.ticksLeft = 0
	}
	ge.addLog("debug", fmt.Sprintf("Research start: %s (cost: %.0f knowledge, %d ticks)", def.Name, def.Cost, ge.Research.totalTicks))
	ge.addLog("info", fmt.Sprintf("Started researching %s (%d ticks)", def.Name, ge.Research.totalTicks))
	return nil
}

// CancelResearch cancels current research (no refund)
func (ge *GameEngine) CancelResearch() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	tech, ok := ge.Research.CancelResearch()
	if !ok {
		return fmt.Errorf("no research in progress")
	}
	def := config.TechByKey()[tech]
	ge.addLog("warning", fmt.Sprintf("Cancelled research: %s (no refund)", def.Name))
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

// memRandFloat returns a [0,1) float from the injected seam, or the package
// default rand if no seam is set. Lets tests force/suppress the roll.
func (ge *GameEngine) memRandFloat() float64 {
	if ge.memoryRand != nil {
		return ge.memoryRand.Float64()
	}
	return rand.Float64()
}

// memRandIntn returns a non-negative int in [0,n) from the injected seam (or the
// package default). n must be > 0.
func (ge *GameEngine) memRandIntn(n int) int {
	if ge.memoryRand != nil {
		return ge.memoryRand.Intn(n)
	}
	return rand.Intn(n)
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
	def := config.TechByKey()[techKey]
	ge.addLog("event", fmt.Sprintf("✦ %s A memory of [cyan]%s[-] stirs — research it free of prerequisites, but at half speed.", ancientMemoryFlavor, def.Name))
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
	for _, t := range config.Technologies() {
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
		return fmt.Errorf("no ancient memory to accept")
	}
	techKey := ge.pendingMemoryTech
	ge.pendingMemoryTech = ""

	// Same combined research_speed sources a normal research gets; the memory
	// penalty (2x ticks) is applied on top inside StartMemoryResearch.
	combinedResearchSpeed := ge.Research.GetBonus("research_speed") +
		ge.permanentBonuses["research_speed"] +
		ge.Prestige.GetBonuses()["research_speed"]
	if err := ge.Research.StartMemoryResearch(techKey, combinedResearchSpeed); err != nil {
		return err
	}
	def := config.TechByKey()[techKey]
	ge.addLog("success", fmt.Sprintf("Recovered the memory of %s — researching at half speed (%d ticks).", def.Name, ge.Research.totalTicks))
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

	ageOrder := ge.progress.GetAgeOrder()

	def := ge.Military.ExpeditionDefByKey(key)
	if def == nil {
		return fmt.Errorf("unknown expedition: %s", key)
	}

	// --- Validate everything BEFORE any deduction so a failed launch never
	// partially charges the player. ---

	// Soldiers resource check (soldiers are now a real resource, not workers).
	haveSoldiers := int(ge.Resources.Get("soldiers"))
	if haveSoldiers < def.SoldiersNeeded {
		return fmt.Errorf("%s needs %d soldiers (have %d)", def.Name, def.SoldiersNeeded, haveSoldiers)
	}

	// Additional resource cost check.
	for res, amount := range def.Cost {
		if ge.Resources.Get(res) < amount {
			return fmt.Errorf("not enough %s: need %.0f, have %.0f", res, amount, ge.Resources.Get(res))
		}
	}

	// Age range + active-expedition validation (does NOT touch resources).
	if err := ge.Military.LaunchExpedition(key, ge.age, ageOrder); err != nil {
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
	ge.addLog("info", fmt.Sprintf("Expedition launched: %s", def.Name))
	return nil
}

// DoPrestige resets the game with prestige bonuses
func (ge *GameEngine) DoPrestige() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	if !ge.Prestige.CanPrestige(ge.age, ageOrder) {
		return fmt.Errorf("must reach Modern Age or later to prestige")
	}

	points := ge.Prestige.CalculatePoints(
		ge.age, ageOrder,
		ge.Milestones.CompletedCount(),
		ge.Research.ResearchedCount(),
		ge.Stats.TotalBuilt,
	)

	ge.Prestige.Prestige(points)

	// Preserve cross-run state before resetting managers
	savedRuins := ge.Buildings.GetAllRuins()

	// Reset all game systems
	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManager()
	ge.Buildings = NewBuildingManager()
	ge.Workers = NewWorkerManager()
	ge.Research = NewResearchManager()
	ge.Military = NewMilitaryManager()
	ge.Events = NewEventManager()
	ge.Milestones = NewMilestoneManager()
	ge.Trade = NewTradeManager()
	ge.Diplomacy = NewDiplomacyManager()
	ge.Stats = NewGameStats()
	// Bus intentionally kept — dashboard subscriptions must survive across resets.
	ge.permanentBonuses = make(map[string]float64)
	ge.buildQueue = nil
	ge.log = nil
	ge.currentEpoch = config.EpochForAge("primitive_age")
	ge.epochEventFired = make(map[string]bool)
	ge.awakeningsFired = make(map[string]bool)
	ge.survivedEpochs = make(map[string]bool)
	ge.pendingCatastrophe = ""
	ge.epochEventHistory = nil
	ge.morale = 0.70
	ge.lowMoraleWarned = false
	// Fresh run: this prestige cycle may roll a new Ancient Memory.
	ge.ancientMemoryUsed = false
	ge.pendingMemoryTech = ""

	// Restore cross-run state
	ge.Buildings.LoadRuins(savedRuins)
	ge.reapplyLegacyBonuses()

	// Apply age unlocks for primitive age
	ge.applyAgeUnlocks("primitive_age")

	// Apply starting resources (base + prestige bonus)
	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)
	for res, amount := range ge.Prestige.GetStartingResources() {
		ge.Resources.Add(res, amount)
	}

	ge.recalculateTickSpeed()

	ge.addLog("success", fmt.Sprintf("Prestige complete! Level %d (+%d points)", ge.Prestige.GetLevel(), points))
	ge.addLog("info", fmt.Sprintf("Passive bonus: +%.0f%% production, +%.0f%% tick speed",
		float64(ge.Prestige.GetLevel())*2, ge.tickSpeedBonus*100))
	if len(ge.legacyBonuses) > 0 {
		ge.addLog("info", fmt.Sprintf("Legacy bonuses active: %d epoch(s)", len(ge.legacyBonuses)))
	}
	if len(savedRuins) > 0 {
		ge.addLog("info", fmt.Sprintf("%d ruin type(s) carry forward from past civilizations.", len(savedRuins)))
	}
	ge.addLog("info", "Type [cyan]help[-] to get started again.")

	// Account lifetime stat (Phase 6): record the prestige IN-MEMORY only — we hold
	// ge.mu here, so RecordPrestige must not do I/O or re-enter the engine. The write
	// is deferred to FlushIfDirty in the autosave block (outside ge.mu).
	if ge.account != nil {
		ge.account.RecordPrestige()
	}

	// Roll for an Ancient Memory cache — prestige level is now >= 1, the age is
	// primitive, and the flag was just cleared above, so this fresh run is eligible.
	ge.maybeOfferAncientMemory()

	return nil
}

// BuyPrestigeUpgrade purchases a prestige upgrade tier
func (ge *GameEngine) BuyPrestigeUpgrade(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if err := ge.Prestige.BuyUpgrade(key); err != nil {
		return err
	}
	ge.addLog("success", fmt.Sprintf("Purchased prestige upgrade: %s", key))
	return nil
}

// Reset completely reinitializes the engine to a fresh state (including prestige)
func (ge *GameEngine) Reset() {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManager()
	ge.Buildings = NewBuildingManager()
	ge.Workers = NewWorkerManager()
	ge.Research = NewResearchManager()
	ge.Military = NewMilitaryManager()
	ge.Events = NewEventManager()
	ge.Milestones = NewMilestoneManager()
	ge.Prestige = NewPrestigeManager()
	ge.Trade = NewTradeManager()
	ge.Diplomacy = NewDiplomacyManager()
	ge.Stats = NewGameStats()
	// Bus intentionally kept — dashboard subscriptions must survive across resets.
	ge.permanentBonuses = make(map[string]float64)
	ge.tickSpeedBonus = 0
	ge.speedMultiplier = 1.0
	ge.buildQueue = nil
	ge.log = nil

	ge.applyAgeUnlocks("primitive_age")
	ge.Resources.Add("food", 25)
	ge.Resources.Add("wood", 50)

	ge.cheaterBadge = false
	ge.eliteBadge = false
	ge.currentEpoch = config.EpochForAge("primitive_age")
	ge.epochEventFired = make(map[string]bool)
	ge.awakeningsFired = make(map[string]bool)
	ge.survivedEpochs = make(map[string]bool)
	ge.pendingCatastrophe = ""
	ge.epochEventHistory = nil
	ge.legacyBonuses = make(map[string]bool)
	ge.catastropheHistory = nil
	ge.morale = moraleNeutral
	ge.lowMoraleWarned = false
	// Full wipe: no previous civilization, so no cache. Clear the run flag.
	ge.ancientMemoryUsed = false
	ge.pendingMemoryTech = ""

	ge.addLog("event", "Game wiped! Starting fresh.")
	ge.addLog("info", "Type [cyan]help[-] for commands.")
}

// GetState returns a snapshot of the game state for UI
func (ge *GameEngine) GetState() GameState {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	popCap := ge.Buildings.GetPopCapacity()
	popCap += int(ge.Research.GetBonus("population") + ge.permanentBonuses["population"] + ge.Prestige.GetBonuses()["population"])
	nextAge := ge.progress.GetNextAge(ge.age)

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
	prestigeBonuses := ge.Prestige.GetBonuses()
	wonderBonuses := ge.getWonderBonuses()
	militaryBonus := ge.Research.GetBonus("military_power") + ge.permanentBonuses["military_power"] + prestigeBonuses["military_power"] + wonderBonuses["military_power"]
	expeditionBonus := ge.Research.GetBonus("expedition_reward") + ge.permanentBonuses["expedition_reward"] + prestigeBonuses["expedition_reward"] + wonderBonuses["expedition_reward"]

	// Prestige snapshot with pending points
	prestigeSnap := ge.Prestige.Snapshot()
	prestigeSnap.CanPrestige = ge.Prestige.CanPrestige(ge.age, ageOrder)
	prestigeSnap.PendingPoints = ge.Prestige.CalculatePoints(
		ge.age, ageOrder,
		ge.Milestones.CompletedCount(),
		ge.Research.ResearchedCount(),
		ge.Stats.TotalBuilt,
	)

	speedMult := ge.speedMultiplier
	if speedMult < 1.0 {
		speedMult = 1.0
	}
	tickInterval := time.Duration(float64(BaseTickInterval) / ((1.0 + ge.tickSpeedBonus) * speedMult))
	if tickInterval < MinTickInterval {
		tickInterval = MinTickInterval
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

	return GameState{
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
		Workers:              ge.Workers.Snapshot(popCap),
		Research:             ge.Research.Snapshot(ge.age, ageOrder),
		Military:             ge.Military.Snapshot(ge.age, ageOrder, soldierResource, int(ge.Resources.GetStorage("soldiers")), ge.Resources.GetRate("soldiers"), ge.Resources.GetAll(), militaryBonus, expeditionBonus),
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
			activeEvents:    ge.Events.GetActive(),
		}),
		ActiveEvents:          ge.Events.GetActive(),
		Prestige:              prestigeSnap,
		Trade:                 ge.Trade.Snapshot(ge.age, ageOrder, ge.Buildings),
		Diplomacy:             ge.Diplomacy.Snapshot(ge.age, ageOrder),
		Log:                   logCopy,
		Stats:                 ge.Stats.Snapshot(),
		SaveExists:            SaveExists("autosave"),
		TickSpeedBonus:        ge.tickSpeedBonus,
		TickIntervalMs:        int(tickInterval.Milliseconds()),
		SpeedMultiplier:       speedMult,
		CheaterBadge:          ge.cheaterBadge,
		EliteBadge:            ge.eliteBadge,
		LastAgeAdvanceSummary: ge.lastAgeAdvanceSummary,
		// Phase 8: epoch fields
		EpochKey: ge.currentEpoch,
		EpochName: func() string {
			if ep, ok := config.EpochByKey()[ge.currentEpoch]; ok {
				return ep.Name
			}
			return ""
		}(),
		EpochIcon: func() string {
			if ep, ok := config.EpochByKey()[ge.currentEpoch]; ok {
				return ep.Icon
			}
			return ""
		}(),
		EpochColor: func() string {
			if ep, ok := config.EpochByKey()[ge.currentEpoch]; ok {
				return ep.Color
			}
			return "white"
		}(),
		EpochSurvived:         ge.survivedEpochs[ge.currentEpoch],
		PendingCatastrophe:    ge.pendingCatastrophe,
		PendingMemoryTech:     ge.pendingMemoryTech,
		PendingMemoryTechName: config.TechByKey()[ge.pendingMemoryTech].Name,
		EpochEventHistory:     ge.epochEventHistory,
		LegacyBonuses: func() map[string]bool {
			out := make(map[string]bool, len(ge.legacyBonuses))
			for k, v := range ge.legacyBonuses {
				out[k] = v
			}
			return out
		}(),
		CatastropheHistory: ge.catastropheHistory,
		History:            ge.History,
		Morale:             ge.morale,
		MoraleCap:          ge.moraleCap(),
		MoraleMultiplier:   ge.moraleMultiplier(),
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
		Modifiers: ge.buildResolver().All(),
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

// addLog appends a log entry (must be called with lock held)
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
)

// applyOfflineProgress applies simulated progress for time spent offline (must be called with lock held)
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

	gains := make(map[string]float64)
	for key, r := range ge.Resources.resources {
		if !ge.Resources.unlocked[key] || r.Rate <= 0 {
			continue
		}
		amount := r.Rate * float64(offlineTicks) * OfflineEfficiency
		if r.Amount+amount > r.Storage {
			amount = r.Storage - r.Amount
		}
		if amount > 0 {
			ge.Resources.Add(key, amount)
			gains[key] = amount
		}
	}

	ge.tick += offlineTicks

	// Log welcome back message
	minutes := int(elapsed.Minutes())
	hours := minutes / 60
	mins := minutes % 60
	var timeStr string
	if hours > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hours, mins)
	} else {
		timeStr = fmt.Sprintf("%dm", mins)
	}

	ge.addLog("event", fmt.Sprintf("Welcome back! You were away for %s.", timeStr))
	if len(gains) > 0 {
		ge.addLog("info", fmt.Sprintf("Offline progress (%d ticks at 50%% efficiency):", offlineTicks))
		for res, amount := range gains {
			ge.addLog("info", fmt.Sprintf("  +%.1f %s", amount, res))
		}
	}
}

// ExchangeResources performs a resource exchange via the trade system
func (ge *GameEngine) ExchangeResources(from, to string, amount float64) (float64, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	got, err := ge.Trade.Exchange(from, to, amount, ge.Resources, ge.Buildings, ge.tick)
	if err != nil {
		return 0, err
	}
	ge.addLog("info", fmt.Sprintf("Traded %.0f %s → %.1f %s", amount, from, got, to))
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
	ge.addLog("info", fmt.Sprintf("Trade route started: %s", key))
	return nil
}

// StopTradeRoute deactivates a trade route
func (ge *GameEngine) StopTradeRoute(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if err := ge.Trade.StopRoute(key); err != nil {
		return err
	}
	ge.addLog("info", fmt.Sprintf("Trade route stopped: %s", key))
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
	ge.addLog("info", fmt.Sprintf("Diplomatic status with %s set to %s", factionKey, status))
	return nil
}

// SendGift sends a gift to a faction
func (ge *GameEngine) SendGift(factionKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	gold := ge.Resources.Get("gold")
	cost, err := ge.Diplomacy.SendGift(factionKey, gold)
	if err != nil {
		return err
	}
	ge.Resources.Remove("gold", cost)
	ge.addLog("info", fmt.Sprintf("Sent gift to %s (+15 opinion)", factionKey))
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
	name := factionKey
	if def, ok := config.FactionByKey()[factionKey]; ok {
		name = def.Name
	}
	ge.addLog("success", fmt.Sprintf("Tribute paid to %s (%.0f gold, %.0f culture) — the war is over.", name, goldCost, cultureCost))
	return nil
}

// RaidCivRoute raids a discovered civilization's trade route — a provocation
// that tanks opinion and may trigger war if standing is already hostile.
func (ge *GameEngine) RaidCivRoute(factionKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	started, err := ge.Diplomacy.RaidTradeRoute(factionKey, ge.tick)
	if err != nil {
		return err
	}
	name := factionKey
	if def, ok := config.FactionByKey()[factionKey]; ok {
		name = def.Name
	}
	if started {
		ge.addLog("warning", fmt.Sprintf("You raided the %s's trade route — they have declared WAR!", name))
	} else {
		ge.addLog("warning", fmt.Sprintf("You raided the %s's trade route (-20 opinion).", name))
	}
	return nil
}

// UpgradeBuilding converts count copies of a legacy building to its pending next-tier
// equivalent, charging the cost delta (new copy cost minus 50% refund on old copy) per unit.
// Pass all=true or count<=0 to upgrade all available copies.
func (ge *GameEngine) UpgradeBuilding(key string, count int, all bool) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	newKey, hasPending := ge.Buildings.GetPendingUpgrade(key)
	if !hasPending {
		return fmt.Errorf("no upgrade available for %s", key)
	}

	byKey := config.BuildingByKey()
	oldDef, hasOld := byKey[key]
	newDef, hasNew := byKey[newKey]
	if !hasOld || !hasNew {
		return fmt.Errorf("building definition not found")
	}

	oldCount := ge.Buildings.GetCount(key)
	if all || count <= 0 {
		count = oldCount
	}
	if count > oldCount {
		count = oldCount
	}
	if count <= 0 {
		return fmt.Errorf("no %s to upgrade", oldDef.Name)
	}

	cost, ok := ge.Buildings.UpgradeCost(key, newKey, count)
	if !ok {
		return fmt.Errorf("could not calculate upgrade cost")
	}

	if !ge.Resources.CanAfford(cost) {
		var needed []string
		for res, amt := range cost {
			have := ge.Resources.Get(res)
			if have < amt {
				needed = append(needed, fmt.Sprintf("%s %.0f/%.0f", res, have, amt))
			}
		}
		sort.Strings(needed)
		return fmt.Errorf("insufficient resources: %s", strings.Join(needed, ", "))
	}

	// Deduct resources
	for res, amt := range cost {
		ge.Resources.Add(res, -amt)
	}

	// Perform partial transform
	moved := ge.Buildings.PartialTransform(key, newKey, count, ge.Workers.RenameAssignment)

	ge.recalculateRates()

	costStr := formatResourceMap(cost)
	if costStr == "" {
		costStr = "free"
	}
	ge.addLog("success", fmt.Sprintf("Upgraded %d %s → %s (cost: %s)",
		moved, oldDef.Name, newDef.Name, costStr))
	return nil
}

// GetAvailableUpgrades returns upgrade info for buildings that have a pending player-driven upgrade.
func (ge *GameEngine) GetAvailableUpgrades() []UpgradeInfo {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	byKey := config.BuildingByKey()
	var result []UpgradeInfo

	for oldKey, newKey := range ge.Buildings.pendingUpgrades {
		count := ge.Buildings.counts[oldKey]
		if count <= 0 {
			continue
		}
		oldDef, ok1 := byKey[oldKey]
		newDef, ok2 := byKey[newKey]
		if !ok1 || !ok2 {
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
	m := make(map[string]bool)
	for _, key := range ge.Research.GetResearched() {
		m[key] = true
	}
	return m
}

// formatMilestoneRewards formats milestone reward effects for display
func formatMilestoneRewards(effects []config.Effect) string {
	var parts []string
	for _, e := range effects {
		switch e.Type {
		case "instant_resource":
			parts = append(parts, fmt.Sprintf("+%.0f %s", e.Value, e.Target))
		case "permanent_bonus":
			if e.Value < 0 {
				parts = append(parts, fmt.Sprintf("%.0f%% %s", e.Value*100, e.Target))
			} else {
				parts = append(parts, fmt.Sprintf("+%.0f%% %s", e.Value*100, e.Target))
			}
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// formatCost formats a cost map for display
func formatCost(cost map[string]float64) string {
	s := ""
	for k, v := range cost {
		if s != "" {
			s += ", "
		}
		s += fmt.Sprintf("%s: %.0f", k, v)
	}
	return s
}
