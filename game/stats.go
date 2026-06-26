package game

import "time"

// GameStats tracks game statistics
type GameStats struct {
	TotalBuilt     int                `json:"total_built"`
	TotalRecruited int                `json:"total_recruited"`
	TotalGathered  map[string]float64 `json:"total_gathered"`
	GameStarted    time.Time          `json:"game_started"`
	AgesReached    []string           `json:"ages_reached"`
	// SoldiersTrained is the cumulative lifetime count of soldiers actually
	// added to the soldiers resource (post-storage-clamp). Soldiers spent on
	// expeditions never decrement this — it only ever grows. Drives the soldier
	// milestone chain (first_soldiers, standing_army, …). Old saves without this
	// field default to 0 and re-accrue progress as new soldiers are trained.
	SoldiersTrained float64 `json:"soldiers_trained"`
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

// RecordRecruit records worker recruitment
func (gs *GameStats) RecordRecruit(count int) {
	gs.TotalRecruited += count
}

// RecordGather records resource gathering
func (gs *GameStats) RecordGather(resource string, amount float64) {
	gs.TotalGathered[resource] += amount
}

// RecordSoldiersTrained adds to the cumulative lifetime soldiers-trained count.
// amount is the post-clamp delta of soldiers added in a tick; negative or zero
// deltas are ignored so spending soldiers never reduces the lifetime total.
func (gs *GameStats) RecordSoldiersTrained(amount float64) {
	if amount > 0 {
		gs.SoldiersTrained += amount
	}
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
