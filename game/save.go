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
	LastFired map[string]int `json:"last_fired"`
	Active    []ActiveEvent  `json:"active"`
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

	save := GameSave{
		Timestamp:  time.Now(),
		Tick:       ge.tick,
		Age:        ge.age,
		Resources:  ge.Resources.GetAll(),
		Storage:    ge.Resources.GetAllStorage(),
		Buildings:  ge.Buildings.GetAll(),
		Villagers:  ge.Villagers.GetAll(),
		Unlocked:   ge.getUnlockedState(),
		Stats:      ge.Stats,
		BuildQueue: ge.buildQueue,
		Research: ResearchSave{
			Researched:  ge.Research.GetResearched(),
			CurrentTech: ge.Research.currentTech,
			TicksLeft:   ge.Research.ticksLeft,
			TotalTicks:  ge.Research.totalTicks,
		},
		Military: MilitarySave{
			ActiveExpedition: ge.Military.GetActiveForSave(),
			CompletedCount:   ge.Military.completedCount,
			TotalLoot:        ge.Military.totalLoot,
		},
		Events: EventSave{
			LastFired: ge.Events.GetLastFired(),
			Active:    ge.Events.GetActiveForSave(),
		},
		Milestones:       ge.Milestones.GetCompleted(),
		PermanentBonuses: ge.permanentBonuses,
		Prestige: PrestigeSave{
			Level:       ge.Prestige.level,
			TotalEarned: ge.Prestige.totalEarned,
			Available:   ge.Prestige.available,
			Upgrades:    ge.Prestige.upgrades,
		},
	}

	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save: %w", err)
	}

	path := filepath.Join(saveDir, filename+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write save: %w", err)
	}
	return nil
}

// LoadGame restores game state from a file
func (ge *GameEngine) LoadGame(filename string) error {
	path := filepath.Join(saveDir, filename+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read save: %w", err)
	}

	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		return fmt.Errorf("failed to parse save: %w", err)
	}

	ge.tick = save.Tick
	ge.age = save.Age
	ge.Resources.LoadAmounts(save.Resources)
	if save.Storage != nil {
		ge.Resources.LoadStorage(save.Storage)
	}
	ge.Buildings.LoadCounts(save.Buildings)
	ge.Villagers.LoadVillagers(save.Villagers)
	if save.Stats != nil {
		ge.Stats = save.Stats
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
	ge.Events.LoadState(save.Events.LastFired, save.Events.Active)
	ge.Milestones.LoadState(save.Milestones)

	if save.PermanentBonuses != nil {
		ge.permanentBonuses = save.PermanentBonuses
	}

	// Restore prestige
	ge.Prestige.LoadState(save.Prestige.Level, save.Prestige.TotalEarned, save.Prestige.Available, save.Prestige.Upgrades)

	ge.recalculateRates()
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

// SaveExists checks if a save file exists
func SaveExists(filename string) bool {
	path := filepath.Join(saveDir, filename+".json")
	_, err := os.Stat(path)
	return err == nil
}
