package game

import (
	"slices"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// Ruleset returns the ruleset the snapshot was made from, or the core set
// for a snapshot with none (one a test or a fixture wrote by hand).
func (st *GameState) Ruleset() *rules.Set { return orCore(st.Rules) }

// GameState is a read-only snapshot of the entire game state for UI consumption
type GameState struct {
	// Rules is the ruleset the snapshot was made from: the engine's at that
	// moment. Read it through Ruleset, which also covers a snapshot written
	// by hand. A Set never changes, so the UI can read it with no lock and
	// never sees half of a change of rules.
	Rules *rules.Set `json:"-"`

	Tick                 int
	Age                  string
	AgeName              string
	AgeReady             bool   // requirements met — player can type 'advance' to proceed
	CurrentAgeWonderKey  string // wonder key required before advancing; "" if none or already built
	CurrentAgeWonderName string // display name of that wonder
	NextAge              string
	NextAgeName          string
	NextAgeBldReqs       map[string]int
	Resources            map[string]ResourceState
	Buildings            map[string]BuildingState
	BuildQueue           []BuildQueueSnapshot
	Workers              WorkerState
	Research             ResearchState
	// Scholars is the town's scholars and what they do (scholars.go).
	Scholars       ScholarsState
	Military       MilitaryState
	Milestones     MilestoneState
	ActiveEvents   []ActiveEventState
	Prestige       PrestigeState
	PrestigeCue    PrestigeCue
	Mastery        MasteryState
	Trade          TradeState
	Diplomacy      DiplomacyState
	Log            []LogEntry
	Stats          StatsSnapshot
	SaveExists     bool
	TickSpeedBonus float64
	TickIntervalMs int
	CheaterBadge   bool
	EliteBadge     bool
	// DevTouched is true once the developer console has changed this run (saved with
	// the run; see GameEngine.markDevTouchedLocked).
	DevTouched bool
	// AccountRecords is true when this run records to the held account: badges,
	// lifetime stats and theme unlocks. It is false for accountless play and for a run
	// that belongs to another account (the game in memory after a switch). A run the
	// developer console has changed records like any other. The dashboard grants
	// milestone themes only while it is true.
	AccountRecords bool
	// Seed is this run's master RNG seed (see GameEngine.seed) — surfaced for
	// reproducibility/debugging. Persisted via GameSave.Seed, not through this
	// snapshot.
	Seed int64
	// RNGDraws and QuipDraws are the gameplay and quip streams' positions (steps
	// taken since Seed), persisted in the save so a load resumes the stream.
	RNGDraws  uint64
	QuipDraws uint64
	// Phase 7: result of the last age advance transformation pass
	LastAgeAdvanceSummary AgeAdvanceSummary
	// Phase 8: epoch system
	EpochKey           string
	EpochName          string
	EpochIcon          string
	EpochColor         string // tview color tag
	EpochSurvived      bool   // player endured a catastrophe this epoch
	PendingCatastrophe string // epoch key if catastrophe modal should show; "" otherwise
	// PendingEndure is what Endure would cost for the pending catastrophe,
	// Brace and garrison included. Zero when none is pending.
	PendingEndure EndureOutcome
	// CatastropheOutlook reports the catastrophe odds at the NEXT passage: the
	// epoch transition, or prestige in the final epoch (see
	// GameEngine.CatastropheOutlook).
	CatastropheOutlook CatastropheOutlook
	// Ancient Memory (Trello yn98pTQw): tech key of a pending cache offer that the UI
	// should pop an accept/decline modal for; "" when there is no pending offer.
	PendingMemoryTech     string
	PendingMemoryTechName string // resolved display name of PendingMemoryTech ("" if none)
	EpochEventHistory     []EpochEventRecord
	// Harbinger is the live harbinger, nil when none is present (harbinger.go).
	Harbinger *HarbingerView
	// SessionStart is the state the loaded save left, captured before
	// offline catch-up (session_mark.go); nil when the game was not loaded.
	SessionStart *SessionMark `json:",omitempty"`
	// HarbingerHistory lists resolved harbingers this run, oldest first.
	HarbingerHistory []HarbingerRecord
	// Phase 9: civilization history + legacy bonuses
	LegacyBonuses      map[string]bool // epochKey -> true if succumb legacy bonus is active
	CatastropheHistory []string        // narrative log entries
	// Catastrophe outcome counts, derived from CatastropheHistory (so they
	// survive prestige like the history does). A pending catastrophe is neither.
	CatastrophesEndured   int
	CatastrophesSuccumbed int
	// SuccumbResearchFactor is Ancient Knowledge: what research time is
	// multiplied by after the research speed pool (x0.8 per distinct epoch
	// succumbed in), e.g. 0.64 after two epochs and 1 with none.
	SuccumbResearchFactor float64
	// LastPassage is the Cosmic Era's prestige passage: pending choice, the
	// Endure share and the Cosmic Legacy flag (last_passage.go).
	LastPassage LastPassageState
	// History overlay
	History *HistoryCollector
	// Morale system
	Morale    float64 // current morale 0.10–cap
	MoraleCap float64 // current cap (1.0 + 0.05 per wonder, plus the techs' term)
	// MoraleMultiplier is the production multiplier the continuous morale curve
	// currently yields (moraleMultiplier()). Exactly 1.0 at the 0.50 pivot, up to
	// 1.0+moraleMaxBonus at the cap, down to moraleMinMult at the 0.10 floor.
	// Exposed so the UI renders the model without re-deriving the formula.
	MoraleMultiplier float64
	// Plan is the build plan in order, each item with the price of its next
	// start and whether it could start now (plan.go).
	Plan []PlanItemView
	// WonderOverflow reports whether what the caps cut off is banked into the
	// current age's wonder (overflow.go).
	WonderOverflow bool
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
	// Pools is every bonus pool a cap or floor can hold (all production,
	// each resource's own production, worker output, research speed,
	// building costs, game speed), with what it has earned and what the
	// engine applies of it (caps.go). The panels read it to say "capped"
	// beside a bonus a limit is holding back. Populated in GetState(); not
	// stored in save JSON.
	Pools map[string]BonusPool `json:"-"`
	// Features is every feature lock (config.FeatureLocks) by key: the tech
	// that opens the command and whether it is open now (features.go).
	// Populated in GetState(); not stored in save JSON.
	Features map[string]FeatureState `json:"-"`
	// AccountStats carries the account-wide LIFETIME (cross-save) stats and
	// achievements for the Stats overlay (the accounts design §3.3, Phase 6). nil when
	// no account is wired (e.g. tests that build an engine without SetAccount).
	// Distinct from Stats above, which is the per-save ge.Stats snapshot.
	// Populated in GetState() from ge.account.LifetimeStats(); not in save JSON.
	AccountStats *AccountStatsView `json:"-"`
}

// PrestigeCue is the engine's estimate of whether a prestige would reach the
// current age's gate sooner on the next run.
type PrestigeCue struct {
	Ready          bool
	Age            string
	FinishTicks    int
	FreshTicks     int
	Multiplier     float64
	NextMultiplier float64
	Seen           bool
}

// AccountStatsView is the read-only UI projection of the account's lifetime stats
// and badges (the accounts design §3.3 / Phase 6). It is a copy — the account never
// hands the UI its mutable backing slices. Badges is every badge as the player may see
// it: a locked badge the spoiler rules withhold is a silhouette with no text in it
// (BadgeView.Hidden), so no screen has to decide what to hide.
type AccountStatsView struct {
	DisplayName          string
	TotalPrestiges       int
	HighestAge           string
	CivilizationsStarted int
	Badges               []BadgeView
	BadgeSummary         BadgeSummary
}

// AgeAdvanceSummary holds data about what changed during an age advance transition.
type AgeAdvanceSummary struct {
	OldAge               string
	NewAge               string
	BuildingsTransformed []BuildingTransform
	BuildingsLegacy      []string // unused: nothing is marked legacy any more
}

// clone returns a copy that shares no slices with s.
func (s AgeAdvanceSummary) clone() AgeAdvanceSummary {
	s.BuildingsTransformed = slices.Clone(s.BuildingsTransformed)
	s.BuildingsLegacy = slices.Clone(s.BuildingsLegacy)
	return s
}

// BuildingTransform describes one building the age advance made older: the
// copies held, and the building its line has for the new age.
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
	// Outcome is set on catastrophe records only: CatastrophePending until the
	// player chooses, then CatastropheEndured or CatastropheSuccumbed. Empty on
	// other event types, and on catastrophe records from saves that predate it
	// whose outcome could not be reconstructed on load.
	Outcome string `json:"Outcome,omitempty"`
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
	// LegacyRate is what the Cosmic Legacy adds: a tenth of everything the
	// resource makes, after the caps and before the food drain
	// (last_passage.go). 0 without the legacy.
	LegacyRate float64
	// MasteryRate is what Era Mastery adds: the net rate × (k − 1) on known
	// ground (mastery.go), 0 on new ground.
	MasteryRate float64
	// PrestigeRate is what lifetime prestige points add after production pools,
	// at the same final stage as Era Mastery.
	PrestigeRate float64
}

// ResourceState represents a single resource's current state
type ResourceState struct {
	Amount    float64
	Rate      float64
	Storage   float64
	Name      string
	Unlocked  bool
	Breakdown RateBreakdown
	// OverCapGrace marks stock kept above a cap that shrank when Era
	// Mastery's speed dropped (the grace rule, mastery.go): it stays until
	// spent, and production adds nothing to it meanwhile.
	OverCapGrace bool
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
	// NeedsTech is the tech key a building of an unlocked age still waits
	// for (Unlocked is false until it is researched); "" otherwise. A wonder
	// waiting for its keystone stays Unlocked, so its bank fills meanwhile:
	// only CanBuild waits.
	NeedsTech string
	// LockTech is the display name of the tech a building of an age that
	// has unlocked it still waits for; "" when it needs none or the tech
	// is researched. LockTechPath is the display names of that tech's
	// prerequisites not researched yet, nearest first and recursively;
	// empty when LockTech can be researched now.
	LockTech     string
	LockTechPath []string
	// Cost for next building
	NextCost   map[string]float64
	CanBuild   bool
	AtMaxCount bool
	// StoreLimit marks a storage building whose next copy costs more than
	// the store would hold with every copy already queued built: the price
	// curve has run past the store, so no more copies can be bought this
	// age. It is not AtMaxCount: storage has no copy limit, only a price.
	StoreLimit bool
	// Wonder-specific: resources banked toward construction
	WonderBank     map[string]float64
	WonderBankFull bool
	// Phase 6: worker assignment fields
	WorkerDomain    string
	WorkerCapacity  int // per-building-instance capacity
	WorkersAssigned int // total workers assigned across all instances
	// IsLegacy is set for an older building (older.go): functional but its
	// line has a newer rung for this age, so no more are built. Read live.
	IsLegacy bool
	// PendingUpgrade is the building an older building held in copies can be
	// upgraded to; "" if none.
	PendingUpgrade string
	// HeldWithLater is the copies held of this building and of every later
	// rung of its line: what a requirement that names this building counts.
	HeldWithLater int
	// Scholars is the scholars one copy of this building comes with (0 for
	// every building but a knowledge building) and ScholarCap the most copies
	// that can be built (0 where there is no cap).
	Scholars   int
	ScholarCap int
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
	// Shares is the split of the workforce the player set (shares.go):
	// domain → percent, only the domains set; nil when every domain is on
	// auto. ShareRows works out what each domain gets.
	Shares map[string]float64
	// AutoRecruit reports whether the shares routine recruits.
	AutoRecruit bool
	// HoldTicks is how long the routine still waits after a worker command
	// (0: it is running).
	HoldTicks int
	// HeldBy is what keeps the population from growing toward the housing:
	// HeldHousing (no room), HeldFood (the food margin test fails) or "" (it
	// can grow). Reserve is the idle villagers, the same number as TotalIdle.
	// FoodNet is the net food a tick before Era Mastery, in the units the
	// villagers eat in, and FoodFeeds the villagers it would feed (FoodNet
	// over one villager's food, 0 when it is not above zero).
	HeldBy    string
	FoodNet   float64
	FoodFeeds int
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
	// Type is the category: "info", "success", "warning", "error", "event",
	// LogRoutine or "debug". It decides the color, the logs panel's tag, and
	// whether a player sees the line at all ("debug" is for dumps only).
	Type string
}

// LogRoutine is the log category for routine confirmations: a line whose
// effect a panel already shows. A build started, queued or finished (a
// wonder's completion excepted), a sale, an upgrade, workers recruited,
// assigned, unassigned or dismissed, a gather, a wonder deposit, a trade, a
// route started or stopped, a research started, an expedition sent, a gift,
// a plan item added or started.
//
// The rule, for any new line: a plain confirmation of something a panel
// already shows is LogRoutine; anything worth noticing takes another
// category. Every category but "debug" (dumps only) shows in both logs, so
// the main window always says what a command did. There, routine lines take
// the plain text color and the notable ones stand out in theirs; the logs
// panel marks routine lines with a dot. Every command that succeeds writes a
// line the main log shows (TestSuccessRepliesAreLogged).
const LogRoutine = "routine"

// StatsSnapshot is the stats for UI display
type StatsSnapshot struct {
	TotalTicks     int
	TotalBuilt     int
	TotalRecruited int
	TotalGathered  map[string]float64
	GameStarted    time.Time
	PlayTime       time.Duration
	AgesReached    []string
	AgeLog         []AgeEntry
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
	// Progress is the work done on the tech in progress and Price the work
	// it needs (its price); both 0 with none in progress.
	Progress, Price float64
	// PerTick is the work the researching scholars do each tick, with every
	// bonus. TicksLeft and TotalTicks are the time the tech in progress has
	// left and takes at that work (0 when no scholar is researching: it has
	// no time to show, see NeedsScholars).
	PerTick       float64
	TicksLeft     int
	TotalTicks    int
	NeedsScholars bool
	// Queue is the techs planned to be researched next, in order.
	Queue           []string
	TotalResearched int
	// Bonuses is what researched techs add to each bonus pool they share
	// with the rest of the game, as fractions ("military_power": 0.5 is
	// +50%): game speed, military power, expedition rewards. Flat is what
	// they add per tick to each resource (a first source of it).
	Bonuses map[string]float64
	Flat    map[string]float64
	// The tech layer, which no cap holds. Output is the techs' bonus on
	// each resource's output and AllOutput their bonus on every resource's
	// (fractions; a resource's layer factor is 1 + its Output + AllOutput).
	// Storage and Housing are their bonus on every store and on housing
	// (fractions).
	Output    map[string]float64
	AllOutput float64
	Storage   float64
	Housing   float64
	// BuildCost, BuildTime and ResearchTime are the techs' factors on what
	// buildings cost, how long construction takes and how long research
	// takes: 1 with no tech, 0.97 after one 3% cut. Mechanics is each
	// mechanic number's tech term (config.Mechanics), for the numbers a
	// researched tech moves.
	BuildCost    float64
	BuildTime    float64
	ResearchTime float64
	Mechanics    map[string]float64
}

// TechState represents one technology's state for UI
type TechState struct {
	Name string
	Age  string
	Cost float64
	// Prerequisites must all be researched. AnyOf is the tech's either-or
	// group: one of its keys must be researched too (nil for a tech with no
	// such group).
	Prerequisites []string
	AnyOf         []string
	Description   string
	Researched    bool
	Available     bool // meets age + prereqs and not yet researched
	// TicksNeeded is how long the tech takes the scholars researching, from
	// the start. For a tech not in progress it counts the scholars the share
	// setting puts on research, even while all teach for want of a tech (0
	// when there are none).
	TicksNeeded int
	PrereqsMet  bool // every prerequisite, and one of AnyOf
	// Kind is the tech's place in the tree: a keystone, on the spine, a
	// capstone or optional (rules.Set.TechKind).
	Kind config.TechKind
	// KeystoneOf is the key of the wonder that cannot be built without this
	// tech ("" for every tech but a keystone).
	KeystoneOf string
}

// === Military Types ===

// MilitaryState represents military system state for UI
type MilitaryState struct {
	// SoldierCount is the current soldiers resource amount.
	SoldierCount int
	// SoldierCap is the soldiers resource storage cap (sum of built military
	// buildings' storage effects). SoldierRate is the per-tick soldiers
	// production rate (net). Both are populated from the soldiers resource.
	SoldierCap    int
	SoldierRate   float64
	DefenseRating float64
	// Threat is the current age's raid threat (rules.Set.AgeThreat) and
	// Mitigation the share of a raid the garrison would blunt against it
	// (config.DefenseMitigation, 0..config.DefenseMitigationCap).
	Threat     float64
	Mitigation float64
	// Saved is what the garrison has saved this run; nil until it saves
	// anything.
	Saved           *DefenseTally
	MilitaryBonus   float64
	ExpeditionBonus float64
	// ActiveShort / ActiveLong / ActiveGrand are the missions under way, one
	// slot per tier; any, all or none may be non-nil.
	ActiveShort *ExpeditionSnapshot
	ActiveLong  *ExpeditionSnapshot
	ActiveGrand *ExpeditionSnapshot
	// ActiveScout (the short mission, else the long one) and ActiveMilitary
	// (the grand mission) are the two slots the older panels and the map
	// read.
	ActiveScout    *ExpeditionSnapshot
	ActiveMilitary *ExpeditionSnapshot
	// Expeditions is the offers on the board, in number order; a slot
	// waiting for its next offer has none. Board has all nine slots.
	Expeditions    []ExpeditionInfo
	Board          []BoardSlotInfo
	CompletedCount int
	TotalLoot      map[string]float64
	// Returns is the last missions to come home, newest first (at most
	// config.MissionReturnsKept). ReturnsMade counts every mission that has
	// come home this run; ReturnsSeen how many of them the player has seen
	// on the Expedition Report. The difference is the new ones.
	Returns     []MissionReturn
	ReturnsMade int
	ReturnsSeen int
	// AutoExpedition is the Geographic Society's standing-orders state. It is
	// NOT filled by MilitaryManager.Snapshot (which has no engine receiver and
	// cannot see the buildings or the countdown) — GetState populates it.
	AutoExpedition AutoExpeditionState
}

// AutoExpeditionState is the player-facing view of automatic expedition
// dispatch (game/auto_expedition.go). Before this existed the whole mechanic
// ran on engine-internal fields, so a built Geographic Society was invisible:
// nothing in the UI could say it existed, when it would next dispatch, or that
// it was sitting starved for supplies.
//
// Count/Assigned/Capacity are surfaced HERE rather than left to the UI to dig
// out of state.Buildings on purpose — the building key belongs to
// auto_expedition.go and must not be hardcoded in a panel.
type AutoExpeditionState struct {
	// Active is true when at least one society stands, i.e. automation is on.
	Active bool
	// TicksLeft counts down to the next dispatch. 0 means a dispatch is DUE
	// and is being retried each tick until the scouting slot frees and the
	// cost is covered.
	TicksLeft int
	// Interval is the effective ticks-between-dispatches at the current
	// investment; 0 when nothing is built.
	Interval int
	// Starved is true while a due dispatch is blocked for want of supplies.
	Starved bool
	// Count is how many societies are built. Assigned / Capacity are the
	// workers on them and the total worker capacity across them — together
	// they are the "fill" that buys cadence.
	Count    int
	Assigned int
	Capacity int
}

// ExpeditionSnapshot represents an active expedition for UI
type ExpeditionSnapshot struct {
	Name string
	// Place is the place the mission goes to; ExpiresTick is the tick its
	// offer would have expired (with Name it is the offer's identity, which
	// the panel's map places the destination from); TotalTicks is its whole
	// length in ticks, of which TicksLeft remain.
	Place       string
	ExpiresTick int
	TotalTicks  int
	// Label is the mission's tier name in the age it was sent in.
	Label         string
	Tier          string
	Grand         string
	Rare          bool
	Soldiers      int
	TicksLeft     int
	LengthMinutes float64
	SuccessChance float64
	HaulMinutes   float64
	SuccessHaul   map[string]float64
	FailureHaul   map[string]float64
}

// ExpeditionInfo represents an available expedition for UI
type ExpeditionInfo struct {
	Name string
	// Key is the offer's number on the board as text ("1" to "9"): what
	// `expedition <number>` takes.
	Key    string
	Number int
	// Place is the place the offer goes to. ExpiresTick is the tick it
	// expires at: it never changes, so with Name it identifies the offer.
	Place       string
	ExpiresTick int
	// Label is the offer's tier name in its age ("Scout", "Trade Trek").
	Label       string
	Tier        string
	Grand       string
	Rare        bool
	Description string
	// LengthMinutes, SuccessChance and HaulMinutes are the offer's rolls
	// (SuccessChance counts the army for a military grand offer, now).
	// ExpiresTicks is how long the offer stays up.
	LengthMinutes   float64
	SuccessChance   float64
	HaulMinutes     float64
	ExpiresTicks    int
	Kind            string
	MinMinutes      float64
	MaxMinutes      float64
	BaseSuccess     float64
	ExpectedHaulMin map[string]float64
	ExpectedHaulMax map[string]float64
	Category        string // "scouting" or "military"
	SoldiersNeeded  int
	// DurationMin/DurationMax mirror the ExpeditionDef's randomized active-duration
	// bounds so the available-expeditions preview can show the rolled range.
	DurationMin int
	DurationMax int
	Difficulty  float64
	// Chance is the mission's chance of failure now, with the player's
	// military power counted on the current age's yardstick
	// (config.MissionDifficulty). Difficulty is the listed one, before it.
	Chance    float64
	Cost      map[string]float64
	CanLaunch bool
	// LaunchBlockReason is a short, player-facing explanation of why the
	// expedition can't be launched right now (e.g. "need 3 soldiers",
	// "need 30 food"). Empty when CanLaunch is true.
	LaunchBlockReason string
}

// BoardSlotInfo is one place on the mission board for the UI: its number, its
// tier, the offer in it (nil while a taken offer's replacement is awaited) and
// how long until that replacement.
type BoardSlotInfo struct {
	Number      int
	Tier        string
	Offer       *ExpeditionInfo
	RefillTicks int
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
	// Rewards is the milestone's rewards as effects, for the panel's
	// "capped" notes beside a bonus a limit holds back (caps.go).
	Rewards  []config.Effect
	Progress []MilestoneProgress
	ChainKey string
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
	KnowledgeCount  int // scholars
	ResearchedTechs map[string]bool
	// Sight, when set, hides every unfinished milestone that needs an age the
	// player cannot see named yet (spoilers.go). nil shows them all.
	Sight        *AgeSight
	activeEvents []ActiveEventState // unexported — only set by engine
	// held, when set (the engine sets it), counts a building named by key as
	// the building and every later rung of its line (BuildingManager.CountFrom);
	// without it the counts in Buildings are read as they are.
	held *BuildingManager
}

// heldFrom is the copies counted toward a requirement that names bld.
func (p MilestoneSnapshotParams) heldFrom(bld string) int {
	if p.held != nil {
		return p.held.CountFrom(bld)
	}
	return p.Buildings[bld]
}

// heldFromAny is the copies counted toward a requirement that names keys
// together, each building once.
func (p MilestoneSnapshotParams) heldFromAny(keys []string) int {
	if p.held != nil {
		return p.held.CountFromAny(keys)
	}
	n := 0
	for _, k := range keys {
		n += p.Buildings[k]
	}
	return n
}

// === Prestige Types ===

// PrestigeState represents the prestige system state for UI
type PrestigeState struct {
	Level         int
	Floor         string
	TotalEarned   int
	Available     int
	Upgrades      map[string]PrestigeUpgradeState
	PendingPoints int // points you'd get if you prestige now
	CanPrestige   bool
	FloorBlocks   bool    // true only when the prestige age floor is the current blocker
	PassiveBonus  float64 // retired: always 0 (the passive became Era Mastery)
	// ProductionMultiplier is the lasting multiplier from all points ever
	// earned; NextProductionMultiplier includes PendingPoints.
	ProductionMultiplier     float64
	NextProductionMultiplier float64
	// NextAge is the age after the current one ("" in the final age) and
	// NextAgePoints what a prestige from it would pay, so the prestige
	// screen can say what one more age adds.
	NextAge       string
	NextAgePoints int
	// ShopVersion is the shop the game is on (config.PrestigeShopVersion
	// once any old perks are refunded).
	ShopVersion int
	// Kit is what the legacy kit remembers (legacy.go).
	Kit LegacyKitState
}

// PrestigeUpgradeState represents one prestige upgrade for UI
type PrestigeUpgradeState struct {
	Name        string
	Description string
	Tier        int
	MaxTier     int
	NextCost    int // 0 if maxed or retired
	Effect      string
	// Retired marks a perk of the first shop: hidden, no effect, refunded.
	Retired bool
	// Kit marks a legacy kit item (one tier).
	Kit bool
}

// === Event Types ===

// EventEffectInfo is one ongoing effect of an active event, for UI display.
type EventEffectInfo struct {
	// Type is the config.Effect.Type: "production" | "production_all" |
	// "tick_speed" | "<res>_rate". The "<res>_rate" form is what a faction
	// specialty boon/setback arrives as (see boon/apply.go) — renderers must
	// handle the suffix, not just the three fixed names.
	Type string
	// Target is the resource key for "production" and "<res>_rate"; empty for
	// the global types.
	Target string
	// Value is a per-tick delta for "production" and a fraction for
	// "production_all" / "tick_speed" / "<res>_rate".
	Value float64
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
	ActiveRoutes    []ActiveRouteInfo
	AvailableRoutes []TradeRouteInfo
	TotalImported   map[string]float64
	// DisruptedResources lists resources currently blockaded by war/embargo; any
	// active route importing one is suspended until the conflict ends.
	DisruptedResources []string
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
	// GiftCost is what a gift costs now: GiftCost with the techs' cut.
	GiftCost float64
	// GiftGain is the opinion a gift earns now: GiftOpinion with the techs'
	// share on top.
	GiftGain int
	// AllyCost is what an alliance costs now: AllyCost with the techs' cut.
	AllyCost float64
	// BoonCrews are the workers faction boons have lent (Extra Hands), one
	// entry per crew, oldest first. They go home when TicksLeft runs out.
	BoonCrews []BoonWorkerLoan
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
	Strength    int    // civ power rating 1-5 (from FactionDef) — drives worldmap dot size
	AtWar       bool   // true while this civ is waging war on the player
	LentWorkers int    // workers currently on loan from this civ (0 if none)
	LentReturn  int    // ticks until lent workers return (0 = none / permanent)
	LentPerm    bool   // lent workers are permanent (opinion was > 80 at lend time)
	// Trade deals (deals.go): the current offers, numbered for `diplomacy
	// accept`; why the civ won't trade ("" if it will); and the ticks of
	// play until the offers rotate.
	Deals         []DealInfo
	DealsBlocked  string
	DealRefreshIn int
}
