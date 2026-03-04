package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/espresso20/ageforge/config"
)

// saveHMACKey is used to sign the save payload. Knowing this key allows clearing the shame badge.
const saveHMACKey = "ageforge-v1-save-integrity"

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
	ChainsCompleted  []string       `json:"chains_completed,omitempty"`
	CurrentTitle     string         `json:"current_title,omitempty"`
	PermanentBonuses map[string]float64 `json:"permanent_bonuses"`
	BuildQueue       []BuildQueueItem   `json:"build_queue"`
	Prestige         PrestigeSave        `json:"prestige"`
	Trade            TradeSave           `json:"trade"`
	Diplomacy        DiplomacySave       `json:"diplomacy"`
	SpeedMultiplier  float64             `json:"speed_multiplier"`
	WonderBanks      map[string]map[string]float64 `json:"wonder_banks,omitempty"`
	// Phase 7: legacy building keys
	LegacyBuildings []string `json:"legacy_buildings,omitempty"`
	// Phase 8: epoch system
	CurrentEpoch       string            `json:"current_epoch,omitempty"`
	EpochEventFired    map[string]bool   `json:"epoch_event_fired,omitempty"`
	SurvivedEpochs     map[string]bool   `json:"survived_epochs,omitempty"`
	PendingCatastrophe string            `json:"pending_catastrophe,omitempty"`
	EpochEventHistory  []EpochEventRecord `json:"epoch_event_history,omitempty"`
	// Phase 9: catastrophe system (persist across Succumb and Prestige)
	Ruins              map[string]int    `json:"ruins,omitempty"`
	LegacyBonuses      map[string]bool   `json:"legacy_bonuses,omitempty"`
	CatastropheHistory []string          `json:"catastrophe_history,omitempty"`
	// Integrity fields
	CheaterBadge bool   `json:"cheater_badge,omitempty"`
	EliteBadge   bool   `json:"elite_badge,omitempty"`
	Signature    string `json:"_sig,omitempty"`
	Proof        string `json:"_proof,omitempty"`
}

// signSave returns the HMAC-SHA256 hex of the save payload (sig and proof cleared).
func signSave(gs GameSave, key string) string {
	gs.Signature = ""
	gs.Proof = ""
	data, _ := json.Marshal(gs)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// verifySave checks the save's HMAC signature and optional elite proof.
// sigValid is true when the save is unsigned (legacy) or the sig matches.
// eliteValid is true when a valid proof is present alongside a valid sig.
func verifySave(gs *GameSave) (sigValid, eliteValid bool) {
	savedSig := gs.Signature
	savedProof := gs.Proof

	// Unsigned save (pre-integrity era) — grant benefit of the doubt
	if savedSig == "" {
		return true, false
	}

	expectedSig := signSave(*gs, saveHMACKey)
	sigValid = hmac.Equal([]byte(savedSig), []byte(expectedSig))

	if sigValid && savedProof != "" {
		mac := hmac.New(sha256.New, []byte(forgeMasterKey))
		mac.Write([]byte(savedSig))
		expectedProof := hex.EncodeToString(mac.Sum(nil))
		eliteValid = hmac.Equal([]byte(savedProof), []byte(expectedProof))
	}
	return
}

// PeekSaveBadges reads a save file and returns its badge state without loading the engine.
// Used by the splash screen to show the elite badge before the game starts.
func PeekSaveBadges(filename string) (cheater, elite bool) {
	path := savePath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}
	var gs GameSave
	if err := json.Unmarshal(data, &gs); err != nil {
		return false, false
	}
	sigValid, eliteValid := verifySave(&gs)
	if !sigValid {
		return true, false
	}
	if eliteValid {
		return false, true
	}
	return gs.CheaterBadge, false
}

// TradeSave holds trade state for save
type TradeSave struct {
	ActiveRoutes   map[string]ActiveRoute `json:"active_routes"`
	SupplyPressure map[string]float64     `json:"supply_pressure"`
	TotalExchanged map[string]float64     `json:"total_exchanged"`
	TotalImported  map[string]float64     `json:"total_imported"`
	TotalExported  map[string]float64     `json:"total_exported"`
}

// DiplomacySave holds diplomacy state for save
type DiplomacySave struct {
	Factions map[string]FactionStateSave `json:"factions"`
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

// saveDirectory returns the canonical save directory: data/saves/ next to the binary.
// Falls back to a CWD-relative path if the binary path cannot be determined.
func saveDirectory() string {
	exe, err := os.Executable()
	if err != nil {
		return "data/saves"
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "data/saves"
	}
	return filepath.Join(filepath.Dir(exe), "data", "saves")
}

// savePath returns the full path for a named save file.
// It checks the canonical (binary-relative) location first, then falls back to
// the legacy CWD-relative location so saves from older versions are still found.
func savePath(filename string) string {
	primary := filepath.Join(saveDirectory(), filename+".json")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	legacy := filepath.Join("data", "saves", filename+".json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return primary // default to canonical for new saves
}

// SaveGame saves the current game state
func (ge *GameEngine) SaveGame(filename string) error {
	dir := saveDirectory()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Snapshot + sign + marshal under lock. The data is small so marshal is fast,
	// and this avoids aliasing bugs where doTick mutates shared maps/slices
	// while json.Marshal reads them concurrently.
	ge.mu.RLock()
	save := ge.buildSaveSnapshot()
	// Sign the payload (sig/proof are empty in snapshot)
	sig := signSave(save, saveHMACKey)
	save.Signature = sig
	if ge.eliteBadge {
		mac := hmac.New(sha256.New, []byte(forgeMasterKey))
		mac.Write([]byte(sig))
		save.Proof = hex.EncodeToString(mac.Sum(nil))
	}
	data, err := json.MarshalIndent(save, "", "  ")
	ge.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal save: %w", err)
	}

	// Atomic write: temp file + rename to prevent corruption on crash
	path := filepath.Join(dir, filename+".json")
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

	// Deep copy trade state
	tradeActiveRoutes := make(map[string]ActiveRoute, len(ge.Trade.activeRoutes))
	for k, v := range ge.Trade.activeRoutes {
		tradeActiveRoutes[k] = *v
	}
	tradePressure := make(map[string]float64, len(ge.Trade.supplyPressure))
	for k, v := range ge.Trade.supplyPressure {
		tradePressure[k] = v
	}
	tradeExchanged := make(map[string]float64, len(ge.Trade.totalExchanged))
	for k, v := range ge.Trade.totalExchanged {
		tradeExchanged[k] = v
	}
	tradeImported := make(map[string]float64, len(ge.Trade.totalImported))
	for k, v := range ge.Trade.totalImported {
		tradeImported[k] = v
	}
	tradeExported := make(map[string]float64, len(ge.Trade.totalExported))
	for k, v := range ge.Trade.totalExported {
		tradeExported[k] = v
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
		ChainsCompleted:  ge.Milestones.GetChainsCompleted(),
		CurrentTitle:     ge.Milestones.GetCurrentTitle(),
		PermanentBonuses: permBonuses,
		Prestige: PrestigeSave{
			Level:       ge.Prestige.level,
			TotalEarned: ge.Prestige.totalEarned,
			Available:   ge.Prestige.available,
			Upgrades:    upgrades,
		},
		Trade: TradeSave{
			ActiveRoutes:   tradeActiveRoutes,
			SupplyPressure: tradePressure,
			TotalExchanged: tradeExchanged,
			TotalImported:  tradeImported,
			TotalExported:  tradeExported,
		},
		Diplomacy: DiplomacySave{
			Factions: ge.Diplomacy.GetFactionsForSave(),
		},
		SpeedMultiplier: ge.speedMultiplier,
		WonderBanks:     ge.Buildings.GetWonderBanks(),
		LegacyBuildings:    ge.Buildings.GetLegacyBuildings(),
		CheaterBadge:       ge.cheaterBadge,
		EliteBadge:         ge.eliteBadge,
		CurrentEpoch:       ge.currentEpoch,
		EpochEventFired:    copyBoolMap(ge.epochEventFired),
		SurvivedEpochs:     copyBoolMap(ge.survivedEpochs),
		PendingCatastrophe: ge.pendingCatastrophe,
		EpochEventHistory:  append([]EpochEventRecord(nil), ge.epochEventHistory...),
		Ruins:              ge.Buildings.GetAllRuins(),
		LegacyBonuses:      copyBoolMap(ge.legacyBonuses),
		CatastropheHistory: append([]string(nil), ge.catastropheHistory...),
	}
}

// LoadGame restores game state from a file
func (ge *GameEngine) LoadGame(filename string) error {
	// File I/O outside the lock
	path := savePath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read save: %w", err)
	}

	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		return fmt.Errorf("failed to parse save: %w", err)
	}

	// Verify integrity before loading — set badge flags on the save struct
	sigValid, eliteValid := verifySave(&save)
	if !sigValid {
		// Sig present but invalid — save was tampered with
		save.CheaterBadge = true
	}
	if eliteValid {
		save.EliteBadge = true
		save.CheaterBadge = false
	}

	// All state mutations under write lock to avoid racing with doTick
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.tick = save.Tick
	ge.age = save.Age
	ge.Villagers.SetAge(save.Age)
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
	ge.Milestones.LoadState(save.Milestones, save.ChainsCompleted, save.CurrentTitle)
	// Reconstruct chains and title for old saves that don't have them
	if len(save.ChainsCompleted) == 0 {
		ge.Milestones.CheckChains()
		ge.Milestones.recalculateTitle()
	}

	if save.PermanentBonuses != nil {
		ge.permanentBonuses = save.PermanentBonuses
	}

	// Restore prestige
	ge.Prestige.LoadState(save.Prestige.Level, save.Prestige.TotalEarned, save.Prestige.Available, save.Prestige.Upgrades)

	// Restore trade
	ge.Trade.LoadState(save.Trade.ActiveRoutes, save.Trade.SupplyPressure, save.Trade.TotalExchanged, save.Trade.TotalImported, save.Trade.TotalExported)

	// Restore diplomacy
	ge.Diplomacy.LoadState(save.Diplomacy.Factions)

	// Restore wonder banks
	if save.WonderBanks != nil {
		ge.Buildings.LoadWonderBanks(save.WonderBanks)
	}

	// Restore Phase 7: legacy buildings
	if len(save.LegacyBuildings) > 0 {
		ge.Buildings.LoadLegacyBuildings(save.LegacyBuildings)
	}

	// Restore speed multiplier
	ge.speedMultiplier = save.SpeedMultiplier
	if ge.speedMultiplier < 1.0 {
		ge.speedMultiplier = 1.0
	}

	// Restore badge state (verified above)
	ge.cheaterBadge = save.CheaterBadge
	ge.eliteBadge = save.EliteBadge

	// Restore Phase 8: epoch system
	if save.CurrentEpoch != "" {
		ge.currentEpoch = save.CurrentEpoch
	} else {
		ge.currentEpoch = config.EpochForAge(save.Age)
	}
	if save.EpochEventFired != nil {
		ge.epochEventFired = save.EpochEventFired
	} else {
		ge.epochEventFired = make(map[string]bool)
	}
	if save.SurvivedEpochs != nil {
		ge.survivedEpochs = save.SurvivedEpochs
	} else {
		ge.survivedEpochs = make(map[string]bool)
	}
	ge.pendingCatastrophe = save.PendingCatastrophe
	ge.epochEventHistory = save.EpochEventHistory

	// Restore Phase 9: catastrophe system
	if save.Ruins != nil {
		ge.Buildings.LoadRuins(save.Ruins)
	}
	if save.LegacyBonuses != nil {
		ge.legacyBonuses = save.LegacyBonuses
	} else {
		ge.legacyBonuses = make(map[string]bool)
	}
	ge.catastropheHistory = save.CatastropheHistory

	ge.recalculateRates()
	ge.recalculateTickSpeed()

	// Apply offline progress for time since save
	ge.applyOfflineProgress(time.Since(save.Timestamp))

	return nil
}

// copyBoolMap returns a deep copy of a map[string]bool.
func copyBoolMap(m map[string]bool) map[string]bool {
	if m == nil {
		return nil
	}
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
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
	entries, err := os.ReadDir(saveDirectory())
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
	dir := saveDirectory()
	entries, err := os.ReadDir(dir)
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
		path := filepath.Join(dir, e.Name())
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

// WipeAllSaves deletes all save files
func WipeAllSaves() error {
	dir := saveDirectory()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return nil
}

// SaveExists checks if a save file exists at either the canonical or legacy path.
func SaveExists(filename string) bool {
	_, err := os.Stat(savePath(filename))
	return err == nil
}
