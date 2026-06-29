package game

import "time"

// GameState is a read-only snapshot of the entire game state for UI consumption
type GameState struct {
	Tick                 int
	Age                  string
	AgeName              string
	AgeReady             bool   // requirements met — player can type 'advance' to proceed
	CurrentAgeWonderKey  string // wonder key required before advancing; "" if none or already built
	CurrentAgeWonderName string // display name of that wonder
	NextAge              string
	NextAgeName          string
	NextAgeResReqs       map[string]float64
	NextAgeBldReqs       map[string]int
	Resources            map[string]ResourceState
	Buildings            map[string]BuildingState
	BuildQueue           []BuildQueueSnapshot
	Workers              WorkerState
	Research             ResearchState
	Military             MilitaryState
	Milestones           MilestoneState
	ActiveEvents         []ActiveEventState
	Prestige             PrestigeState
	Trade                TradeState
	Diplomacy            DiplomacyState
	Log                  []LogEntry
	Stats                StatsSnapshot
	SaveExists           bool
	TickSpeedBonus       float64
	TickIntervalMs       int
	SpeedMultiplier      float64
	CheaterBadge         bool
	EliteBadge           bool
	// Phase 7: result of the last age advance transformation pass
	LastAgeAdvanceSummary AgeAdvanceSummary
	// Phase 8: epoch system
	EpochKey           string
	EpochName          string
	EpochIcon          string
	EpochColor         string // tview color tag
	EpochSurvived      bool   // player endured a catastrophe this epoch
	PendingCatastrophe string // epoch key if catastrophe modal should show; "" otherwise
	// Ancient Memory (Trello yn98pTQw): tech key of a pending cache offer that the UI
	// should pop an accept/decline modal for; "" when there is no pending offer.
	PendingMemoryTech     string
	PendingMemoryTechName string // resolved display name of PendingMemoryTech ("" if none)
	EpochEventHistory     []EpochEventRecord
	// Phase 9: civilization history + legacy bonuses
	LegacyBonuses      map[string]bool // epochKey -> true if succumb legacy bonus is active
	CatastropheHistory []string        // narrative log entries
	// History overlay
	History *HistoryCollector
	// Morale system
	Morale    float64 // current morale 0.10–cap
	MoraleCap float64 // current cap (1.0 + 0.05 per wonder)
	// MoraleMultiplier is the production multiplier the continuous morale curve
	// currently yields (moraleMultiplier()). Exactly 1.0 at the 0.50 pivot, up to
	// 1.0+moraleMaxBonus at the cap, down to moraleMinMult at the 0.10 floor.
	// Exposed so the UI renders the model without re-deriving the formula.
	MoraleMultiplier float64
	// PermanentBonuses is the authoritative runtime map of all cumulative
	// permanent bonuses (epoch events, legacy, milestones, etc.).
	// Populated in GetState(); not stored in save JSON.
	PermanentBonuses map[string]float64 `json:"-"`
	// Modifiers is a flat snapshot of every modifier the engine's resolver
	// aggregated this tick (research, prestige, wonders, permanent, morale,
	// active events). The UI rebuilds a Resolver from it (NewResolver +
	// AddAll) to render the Active Multipliers panel — same source of truth
	// the engine uses for its rates, so the panel can never drift from the
	// math. Populated in GetState(); not stored in save JSON.
	Modifiers []Modifier `json:"-"`
	// AccountStats carries the account-wide LIFETIME (cross-save) stats and
	// achievements for the Stats overlay (accounts.md §3.3, Phase 6). nil when
	// no account is wired (e.g. tests that build an engine without SetAccount).
	// Distinct from Stats above, which is the per-save ge.Stats snapshot.
	// Populated in GetState() from ge.account.LifetimeStats(); not in save JSON.
	AccountStats *AccountStatsView `json:"-"`
}

// AccountStatsView is the read-only UI projection of the account's lifetime stats
// and achievements (accounts.md §3.3 / Phase 6). It is a copy — the account never
// hands the UI its mutable backing slices. Achievements holds unlocked keys; the UI
// resolves human names via game.AchievementName.
type AccountStatsView struct {
	DisplayName          string
	TotalPrestiges       int
	HighestAge           string
	CivilizationsStarted int
	SavesCompleted       int
	Achievements         []string
}

// AgeAdvanceSummary holds data about what changed during an age advance transition.
type AgeAdvanceSummary struct {
	OldAge               string
	NewAge               string
	BuildingsTransformed []BuildingTransform
	BuildingsLegacy      []string // keys of buildings newly marked legacy this transition
}

// BuildingTransform describes one building that upgraded during an age advance.
type BuildingTransform struct {
	OldKey  string
	OldName string
	NewKey  string
	NewName string
	Count   int
}

// RuinState represents one ruin entry (building type + count) persisting across Succumb resets.
type RuinState struct {
	Key   string
	Name  string
	Count int
}

// EpochEventRecord records one epoch transition event for the civilization history log.
type EpochEventRecord struct {
	EpochKey  string
	EpochName string
	EventKey  string
	EventName string
	EventType string // good_minor/good_major/good_legendary/bad_challenging/catastrophe
	Tick      int
}

// BuildQueueSnapshot represents a building under construction for UI
type BuildQueueSnapshot struct {
	Name       string
	TicksLeft  int
	TotalTicks int
}

// RateBreakdown shows the components that make up a resource's net rate
type RateBreakdown struct {
	BuildingRate float64
	WorkerRate   float64
	ResearchRate float64
	EventRate    float64
	TradeRate    float64
	FoodDrain    float64
	BonusRate    float64
}

// ResourceState represents a single resource's current state
type ResourceState struct {
	Amount    float64
	Rate      float64
	Storage   float64
	Name      string
	Unlocked  bool
	Breakdown RateBreakdown
}

// BuildingState represents a building type's current state
type BuildingState struct {
	Count       int
	Name        string
	Category    string
	Description string
	Flavor      string // cosmetic personality line; mirrors BuildingDef.Flavor, may be empty
	Unlocked    bool
	AgeKey      string // age this building first becomes available
	// Cost for next building
	NextCost   map[string]float64
	CanBuild   bool
	AtMaxCount bool
	// Wonder-specific: resources banked toward construction
	WonderBank     map[string]float64
	WonderBankFull bool
	// Phase 6: worker assignment fields
	WorkerDomain    string
	WorkerCapacity  int // per-building-instance capacity
	WorkersAssigned int // total workers assigned across all instances
	// Phase 7: lineage legacy flag
	IsLegacy bool // functional but superseded — can't build more; grayed in UI
	// Player-driven upgrade: key of next-tier building this can be upgraded to; "" if none
	PendingUpgrade string
	// Phase 9: ruins
	RuinCount int // ruins of this building type (produce at 50%, no workers, can't rebuild)
}

// WorkerState represents all worker info
type WorkerState struct {
	Types     map[string]WorkerDomainState
	TotalPop  int
	MaxPop    int
	TotalIdle int
	FoodDrain float64
}

// WorkerDomainState represents one worker domain's state
type WorkerDomainState struct {
	Name        string
	Count       int
	IdleCount   int
	Assignments map[string]int
	Unlocked    bool
}

// LogEntry is a timestamped game log message
type LogEntry struct {
	Tick    int
	Message string
	Type    string // "info", "success", "warning", "error", "event"
}

// StatsSnapshot is the stats for UI display
type StatsSnapshot struct {
	TotalTicks     int
	TotalBuilt     int
	TotalRecruited int
	TotalGathered  map[string]float64
	GameStarted    time.Time
	PlayTime       time.Duration
	AgesReached    []string
}

// WorkerInfo is used for save/load serialization
type WorkerInfo struct {
	Count      int            `json:"count"`
	FoodCost   float64        `json:"food_cost"`
	Assignment map[string]int `json:"assignment"`
}

// === Research Types ===

// ResearchState represents the research system state for UI
type ResearchState struct {
	Techs           map[string]TechState
	CurrentTech     string
	CurrentTechName string
	TicksLeft       int
	TotalTicks      int
	TotalResearched int
	Bonuses         map[string]float64
}

// TechState represents one technology's state for UI
type TechState struct {
	Name          string
	Age           string
	Cost          float64
	Prerequisites []string
	Description   string
	Researched    bool
	Available     bool // meets age + prereqs and not yet researched
	PrereqsMet    bool
}

// === Military Types ===

// MilitaryState represents military system state for UI
type MilitaryState struct {
	// SoldierCount is the current soldiers resource amount.
	SoldierCount int
	// SoldierCap is the soldiers resource storage cap (sum of built military
	// buildings' storage effects). SoldierRate is the per-tick soldiers
	// production rate (net). Both are populated from the soldiers resource.
	SoldierCap       int
	SoldierRate      float64
	DefenseRating   float64
	MilitaryBonus   float64
	ExpeditionBonus float64
	// ActiveScout / ActiveMilitary are the per-category active expeditions. A
	// scouting and a military expedition can run concurrently, so either, both,
	// or neither may be non-nil.
	ActiveScout    *ExpeditionSnapshot
	ActiveMilitary *ExpeditionSnapshot
	Expeditions    []ExpeditionInfo
	CompletedCount int
	TotalLoot      map[string]float64
}

// ExpeditionSnapshot represents an active expedition for UI
type ExpeditionSnapshot struct {
	Name      string
	Soldiers  int
	TicksLeft int
}

// ExpeditionInfo represents an available expedition for UI
type ExpeditionInfo struct {
	Name           string
	Key            string
	Category       string // "scouting" or "military"
	SoldiersNeeded int
	Duration       int
	Difficulty     float64
	Cost           map[string]float64
	Description    string
	CanLaunch      bool
	// LaunchBlockReason is a short, player-facing explanation of why the
	// expedition can't be launched right now (e.g. "need 3 soldiers",
	// "need 30 food"). Empty when CanLaunch is true.
	LaunchBlockReason string
}

// === Milestone Types ===

// MilestoneState represents milestone system state for UI
type MilestoneState struct {
	Milestones     map[string]MilestoneInfo
	CompletedCount int
	TotalCount     int
	VisibleCount   int
	Chains         []ChainInfo
	CurrentTitle   string
}

// MilestoneInfo represents one milestone for UI
type MilestoneInfo struct {
	Name        string
	Description string
	Category    string
	Hidden      bool
	Visible     bool // computed: completed || !hidden || progress > 0.5
	Completed   bool
	RewardText  string
	Progress    []MilestoneProgress
	ChainKey    string
}

// MilestoneProgress represents progress toward one condition of a milestone
type MilestoneProgress struct {
	Label   string
	Current float64
	Target  float64
	Met     bool
}

// ChainInfo represents a milestone chain for UI
type ChainInfo struct {
	Name           string
	Key            string
	Category       string
	CompletedCount int
	TotalCount     int
	Complete       bool
	Title          string
	BoostActive    bool
}

// MilestoneSnapshotParams holds data needed to compute milestone progress/visibility
type MilestoneSnapshotParams struct {
	Tick            int
	Age             string
	AgeOrder        map[string]int
	Resources       map[string]float64
	Buildings       map[string]int
	Population      int
	TechCount       int
	TotalBuilt      int
	SoldierCount    int // soldiers resource amount (live); used by non-milestone consumers
	SoldiersTrained int // cumulative lifetime soldiers trained; drives soldier milestones
	WonderCount     int
	KnowledgeCount  int
	ResearchedTechs map[string]bool
	activeEvents    []ActiveEventState // unexported — only set by engine
}

// === Prestige Types ===

// PrestigeState represents the prestige system state for UI
type PrestigeState struct {
	Level         int
	TotalEarned   int
	Available     int
	Upgrades      map[string]PrestigeUpgradeState
	PendingPoints int // points you'd get if you prestige now
	CanPrestige   bool
	PassiveBonus  float64 // current production_all bonus
}

// PrestigeUpgradeState represents one prestige upgrade for UI
type PrestigeUpgradeState struct {
	Name        string
	Description string
	Tier        int
	MaxTier     int
	NextCost    int // 0 if maxed
	Effect      string
}

// === Event Types ===

// EventEffectInfo is one ongoing effect of an active event, for UI display.
type EventEffectInfo struct {
	Type   string  // config.Effect.Type: "production" | "production_all" | "tick_speed"
	Target string  // resource key for "production"; empty/"" for the global ones
	Value  float64 // per-tick delta for "production"; fraction for "production_all"/"tick_speed"
}

// ActiveEventState represents an active timed event for UI
type ActiveEventState struct {
	Name      string
	Key       string
	TicksLeft int
	Effects   []EventEffectInfo
}

// === Trade Types ===

// TradeState represents the trade system state for UI
type TradeState struct {
	ExchangeRates   map[string]ExchangeRateInfo
	ActiveRoutes    []ActiveRouteInfo
	AvailableRoutes []TradeRouteInfo
	TotalExchanged  map[string]float64
	TotalImported   map[string]float64
	// DisruptedResources lists resources currently blockaded by war/embargo; any
	// active route importing one is suspended until the conflict ends.
	DisruptedResources []string
}

// ExchangeRateInfo represents a single exchange rate for UI
type ExchangeRateInfo struct {
	From     string
	To       string
	Rate     float64
	BaseRate float64
	Pressure float64
}

// ActiveRouteInfo represents an active trade route for UI
type ActiveRouteInfo struct {
	Name        string
	Key         string
	TicksLeft   int
	CyclesDone  int
	Export      map[string]float64
	Import      map[string]float64
	Disrupted   bool   // true when blockaded by war/embargo (income suspended)
	DisruptedBy string // the imported resource that is blockaded (if Disrupted)
}

// TradeRouteInfo represents an available trade route for UI
type TradeRouteInfo struct {
	Name        string
	Key         string
	Export      map[string]float64
	Import      map[string]float64
	CanStart    bool
	RequiredBld string
	MinCount    int
	Description string
}

// === Diplomacy Types ===

// DiplomacyState represents the diplomacy system state for UI
type DiplomacyState struct {
	Factions map[string]FactionInfo
}

// FactionInfo represents an NPC civilization for UI
type FactionInfo struct {
	Name        string
	Discovered  bool
	Opinion     int
	Status      string
	Specialty   string
	TradeBonus  float64
	TradeCount  int
	Personality string // "aggressive" | "peaceful" | "mercantile" | "isolationist"
	Backstory   string // flavour shown in the overlay once discovered
	AtWar       bool   // true while this civ is waging war on the player
	LentWorkers int    // workers currently on loan from this civ (0 if none)
	LentReturn  int    // ticks until lent workers return (0 = none / permanent)
	LentPerm    bool   // lent workers are permanent (opinion was > 80 at lend time)
}
