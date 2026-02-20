package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GameSave represents a saved game state
type GameSave struct {
	Timestamp  time.Time               `json:"timestamp"`
	Tick       int                     `json:"tick"`
	Age        string                  `json:"age"`
	Resources  map[string]float64      `json:"resources"`
	Storage    map[string]float64      `json:"storage"`
	Buildings  map[string]int          `json:"buildings"`
	Villagers  map[string]VillagerInfo `json:"villagers"`
	Unlocked   UnlockedState           `json:"unlocked"`
	Stats      *GameStats              `json:"stats"`
	// Phase 3 additions
	Research         ResearchSave   `json:"research"`
	Military         MilitarySave   `json:"military"`
	Events           EventSave      `json:"events"`
	Milestones       []string       `json:"milestones"`
	PermanentBonuses map[string]float64 `json:"permanent_bonuses"`
	BuildQueue       []BuildQueueItem   `json:"build_queue"`
	Prestige         PrestigeSave        `json:"prestige"`
	SpeedMultiplier  float64             `json:"speed_multiplier"`
}

// PrestigeSave holds prestige state for save
type PrestigeSave struct {
	Level       int            `json:"level"`
	TotalEarned int            `json:"total_earned"`
	Available   int            `json:"available"`
	Upgrades    map[string]int `json:"upgrades"`
}

// ResearchSave holds research state for save
type ResearchSave struct {
	Researched  []string `json:"researched"`
	CurrentTech string   `json:"current_tech"`
	TicksLeft   int      `json:"ticks_left"`
	TotalTicks  int      `json:"total_ticks"`
}

// MilitarySave holds military state for save
type MilitarySave struct {
	ActiveExpedition *ActiveExpedition  `json:"active_expedition"`
	CompletedCount   int                `json:"completed_count"`
	TotalLoot        map[string]float64 `json:"total_loot"`
}

// EventSave holds event state for save
type EventSave struct {
	LastFired     map[string]int `json:"last_fired"`
	Active        []ActiveEvent  `json:"active"`
	NextEventTick int            `json:"next_event_tick"`
	GoodStreak    int            `json:"good_streak"`
	BadStreak     int            `json:"bad_streak"`
}

// UnlockedState tracks what's been unlocked
type UnlockedState struct {
	Resources []string `json:"resources"`
	Buildings []string `json:"buildings"`
	Villagers []string `json:"villagers"`
}

const saveDir = "data/saves"

// SaveGame saves the current game state
func (ge *GameEngine) SaveGame(filename string) error {
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Snapshot + marshal under lock. The data is small so marshal is fast,
	// and this avoids aliasing bugs where doTick mutates shared maps/slices
	// while json.Marshal reads them concurrently.
	ge.mu.RLock()
	save := ge.buildSaveSnapshot()
	data, err := json.MarshalIndent(save, "", "  ")
	ge.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal save: %w", err)
	}

	// Atomic write: temp file + rename to prevent corruption on crash
	path := filepath.Join(saveDir, filename+".json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write save: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to finalize save: %w", err)
	}
	return nil
}

// buildSaveSnapshot creates a GameSave from current state (must be called with lock held)
func (ge *GameEngine) buildSaveSnapshot() GameSave {
	// Deep copy build queue
	queue := make([]BuildQueueItem, len(ge.buildQueue))
	copy(queue, ge.buildQueue)

	// Deep copy permanent bonuses
	permBonuses := make(map[string]float64, len(ge.permanentBonuses))
	for k, v := range ge.permanentBonuses {
		permBonuses[k] = v
	}

	// Deep copy military loot
	totalLoot := make(map[string]float64, len(ge.Military.totalLoot))
	for k, v := range ge.Military.totalLoot {
		totalLoot[k] = v
	}

	// Deep copy prestige upgrades
	upgrades := make(map[string]int, len(ge.Prestige.upgrades))
	for k, v := range ge.Prestige.upgrades {
		upgrades[k] = v
	}

	// Deep copy stats
	statsGathered := make(map[string]float64, len(ge.Stats.TotalGathered))
	for k, v := range ge.Stats.TotalGathered {
		statsGathered[k] = v
	}
	agesReached := make([]string, len(ge.Stats.AgesReached))
	copy(agesReached, ge.Stats.AgesReached)

	return GameSave{
		Timestamp: time.Now(),
		Tick:      ge.tick,
		Age:       ge.age,
		Resources: ge.Resources.GetAll(),
		Storage:   ge.Resources.GetAllStorage(),
		Buildings: ge.Buildings.GetAll(),
		Villagers: ge.Villagers.GetAll(),
		Unlocked:  ge.getUnlockedState(),
		Stats: &GameStats{
			TotalBuilt:     ge.Stats.TotalBuilt,
			TotalRecruited: ge.Stats.TotalRecruited,
			TotalGathered:  statsGathered,
			GameStarted:    ge.Stats.GameStarted,
			AgesReached:    agesReached,
		},
		BuildQueue: queue,
		Research: ResearchSave{
			Researched:  ge.Research.GetResearched(),
			CurrentTech: ge.Research.currentTech,
			TicksLeft:   ge.Research.ticksLeft,
			TotalTicks:  ge.Research.totalTicks,
		},
		Military: MilitarySave{
			ActiveExpedition: ge.Military.GetActiveForSave(),
			CompletedCount:   ge.Military.completedCount,
			TotalLoot:        totalLoot,
		},
		Events: EventSave{
			LastFired:     ge.Events.GetLastFired(),
			Active:        ge.Events.GetActiveForSave(),
			NextEventTick: ge.Events.GetNextEventTick(),
			GoodStreak:    ge.Events.goodStreak,
			BadStreak:     ge.Events.badStreak,
		},
		Milestones:       ge.Milestones.GetCompleted(),
		PermanentBonuses: permBonuses,
		Prestige: PrestigeSave{
			Level:       ge.Prestige.level,
			TotalEarned: ge.Prestige.totalEarned,
			Available:   ge.Prestige.available,
			Upgrades:    upgrades,
		},
		SpeedMultiplier: ge.speedMultiplier,
	}
}

// LoadGame restores game state from a file
func (ge *GameEngine) LoadGame(filename string) error {
	// File I/O outside the lock
	path := filepath.Join(saveDir, filename+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read save: %w", err)
	}

	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		return fmt.Errorf("failed to parse save: %w", err)
	}

	// All state mutations under write lock to avoid racing with doTick
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.tick = save.Tick
	ge.age = save.Age
	ge.Resources.LoadAmounts(save.Resources)
	if save.Storage != nil {
		ge.Resources.LoadStorage(save.Storage)
	}
	ge.Buildings.LoadCounts(save.Buildings)
	ge.Villagers.LoadVillagers(save.Villagers)
	if save.Stats != nil {
		// Deep copy stats to avoid aliasing with the deserialized save
		gathered := make(map[string]float64, len(save.Stats.TotalGathered))
		for k, v := range save.Stats.TotalGathered {
			gathered[k] = v
		}
		ages := make([]string, len(save.Stats.AgesReached))
		copy(ages, save.Stats.AgesReached)
		ge.Stats = &GameStats{
			TotalBuilt:     save.Stats.TotalBuilt,
			TotalRecruited: save.Stats.TotalRecruited,
			TotalGathered:  gathered,
			GameStarted:    save.Stats.GameStarted,
			AgesReached:    ages,
		}
	}
	ge.buildQueue = save.BuildQueue

	// Restore unlocks
	for _, key := range save.Unlocked.Resources {
		ge.Resources.UnlockResource(key)
	}
	for _, key := range save.Unlocked.Buildings {
		ge.Buildings.UnlockBuilding(key)
	}
	for _, key := range save.Unlocked.Villagers {
		ge.Villagers.UnlockType(key)
	}

	// Restore Phase 3 systems
	ge.Research.LoadState(save.Research.Researched, save.Research.CurrentTech, save.Research.TicksLeft, save.Research.TotalTicks)
	ge.Military.LoadState(save.Military.ActiveExpedition, save.Military.CompletedCount, save.Military.TotalLoot)
	ge.Events.LoadState(save.Events.LastFired, save.Events.Active, save.Events.NextEventTick, save.Events.GoodStreak, save.Events.BadStreak)
	ge.Milestones.LoadState(save.Milestones)

	if save.PermanentBonuses != nil {
		ge.permanentBonuses = save.PermanentBonuses
	}

	// Restore prestige
	ge.Prestige.LoadState(save.Prestige.Level, save.Prestige.TotalEarned, save.Prestige.Available, save.Prestige.Upgrades)

	// Restore speed multiplier
	ge.speedMultiplier = save.SpeedMultiplier
	if ge.speedMultiplier < 1.0 {
		ge.speedMultiplier = 1.0
	}

	ge.recalculateRates()
	ge.recalculateTickSpeed()

	// Apply offline progress for time since save
	ge.applyOfflineProgress(time.Since(save.Timestamp))

	return nil
}

// getUnlockedState collects all unlock states for saving
func (ge *GameEngine) getUnlockedState() UnlockedState {
	state := UnlockedState{}
	for _, def := range ge.progress.ages {
		order := ge.progress.ageIndex[def.Key]
		currentOrder := ge.progress.ageIndex[ge.age]
		if order <= currentOrder {
			state.Resources = append(state.Resources, def.UnlockResources...)
			state.Buildings = append(state.Buildings, def.UnlockBuildings...)
			state.Villagers = append(state.Villagers, def.UnlockVillagers...)
		}
	}
	return state
}

// ListSaves returns available save files
func ListSaves() ([]string, error) {
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var saves []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			name := e.Name()
			saves = append(saves, name[:len(name)-5]) // strip .json
		}
	}
	return saves, nil
}

// SaveInfo holds metadata about a save file
type SaveInfo struct {
	Name      string
	Timestamp time.Time
	Age       string
}

// ListSaveDetails returns metadata for each save file
func ListSaveDetails() ([]SaveInfo, error) {
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var saves []SaveInfo
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5]
		path := filepath.Join(saveDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var header struct {
			Timestamp time.Time `json:"timestamp"`
			Age       string    `json:"age"`
		}
		if err := json.Unmarshal(data, &header); err != nil {
			continue
		}
		saves = append(saves, SaveInfo{
			Name:      name,
			Timestamp: header.Timestamp,
			Age:       header.Age,
		})
	}
	return saves, nil
}

// SaveExists checks if a save file exists
func SaveExists(filename string) bool {
	path := filepath.Join(saveDir, filename+".json")
	_, err := os.Stat(path)
	return err == nil
}
