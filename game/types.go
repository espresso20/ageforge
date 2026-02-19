package game

import "time"

// GameState is a read-only snapshot of the entire game state for UI consumption
type GameState struct {
	Tick           int
	Age            string
	AgeName        string
	NextAge        string
	NextAgeName    string
	NextAgeResReqs map[string]float64
	NextAgeBldReqs map[string]int
	Resources      map[string]ResourceState
	Buildings      map[string]BuildingState
	BuildQueue     []BuildQueueSnapshot
	Villagers      VillagerState
	Research       ResearchState
	Military       MilitaryState
	Milestones     MilestoneState
	ActiveEvents   []ActiveEventState
	Prestige       PrestigeState
	Log            []LogEntry
	Stats          StatsSnapshot
	SaveExists     bool
}

// BuildQueueSnapshot represents a building under construction for UI
type BuildQueueSnapshot struct {
	Name       string
	TicksLeft  int
	TotalTicks int
}

// ResourceState represents a single resource's current state
type ResourceState struct {
	Amount   float64
	Rate     float64
	Storage  float64
	Name     string
	Unlocked bool
}

// BuildingState represents a building type's current state
type BuildingState struct {
	Count       int
	Name        string
	Category    string
	Description string
	Unlocked    bool
	// Cost for next building
	NextCost map[string]float64
	CanBuild bool
}

// VillagerState represents all villager info
type VillagerState struct {
	Types     map[string]VillagerTypeState
	TotalPop  int
	MaxPop    int
	TotalIdle int
	FoodDrain float64
}

// VillagerTypeState represents one villager type's state
type VillagerTypeState struct {
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

// VillagerInfo is used for save/load serialization
type VillagerInfo struct {
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
	SoldierCount     int
	DefenseRating    float64
	MilitaryBonus    float64
	ExpeditionBonus  float64
	ActiveExpedition *ExpeditionSnapshot
	Expeditions      []ExpeditionInfo
	CompletedCount   int
	TotalLoot        map[string]float64
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
	SoldiersNeeded int
	Duration       int
	Difficulty     float64
	Description    string
	CanLaunch      bool
}

// === Milestone Types ===

// MilestoneState represents milestone system state for UI
type MilestoneState struct {
	Milestones     map[string]MilestoneInfo
	CompletedCount int
	TotalCount     int
}

// MilestoneInfo represents one milestone for UI
type MilestoneInfo struct {
	Name        string
	Description string
	Completed   bool
}

// === Prestige Types ===

// PrestigeState represents the prestige system state for UI
type PrestigeState struct {
	Level        int
	TotalEarned  int
	Available    int
	Upgrades     map[string]PrestigeUpgradeState
	PendingPoints int  // points you'd get if you prestige now
	CanPrestige  bool
	PassiveBonus float64 // current production_all bonus
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

// ActiveEventState represents an active timed event for UI
type ActiveEventState struct {
	Name      string
	Key       string
	TicksLeft int
}
