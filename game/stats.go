package game

import "time"

// GameStats tracks game statistics
type GameStats struct {
	TotalBuilt     int                `json:"total_built"`
	TotalRecruited int                `json:"total_recruited"`
	TotalGathered  map[string]float64 `json:"total_gathered"`
	GameStarted    time.Time          `json:"game_started"`
	AgesReached    []string           `json:"ages_reached"`
}

// NewGameStats creates a new stats tracker
func NewGameStats() *GameStats {
	return &GameStats{
		TotalGathered: make(map[string]float64),
		GameStarted:   time.Now(),
		AgesReached:   []string{"primitive_age"},
	}
}

// RecordBuild records a building construction
func (gs *GameStats) RecordBuild() {
	gs.TotalBuilt++
}

// RecordRecruit records villager recruitment
func (gs *GameStats) RecordRecruit(count int) {
	gs.TotalRecruited += count
}

// RecordGather records resource gathering
func (gs *GameStats) RecordGather(resource string, amount float64) {
	gs.TotalGathered[resource] += amount
}

// RecordAge records reaching a new age
func (gs *GameStats) RecordAge(age string) {
	for _, a := range gs.AgesReached {
		if a == age {
			return
		}
	}
	gs.AgesReached = append(gs.AgesReached, age)
}

// Snapshot returns a stats snapshot for UI
func (gs *GameStats) Snapshot() StatsSnapshot {
	gathered := make(map[string]float64)
	for k, v := range gs.TotalGathered {
		gathered[k] = v
	}
	ages := make([]string, len(gs.AgesReached))
	copy(ages, gs.AgesReached)
	return StatsSnapshot{
		TotalBuilt:     gs.TotalBuilt,
		TotalRecruited: gs.TotalRecruited,
		TotalGathered:  gathered,
		GameStarted:    gs.GameStarted,
		PlayTime:       time.Since(gs.GameStarted),
		AgesReached:    ages,
	}
}
