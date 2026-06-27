package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// saveHMACKey is the HMAC signing key for save integrity. Its presence in the
// binary means a determined player can **** a *** signature, which is
// ***** — the badge system is cosmetic, not security-critical. [[REDACTED]] :)
const saveHMACKey = "ageforge-v1-save-integrity"

// GameSave is the serialised on-disk representation of a game state.
// All subsystem state is flattened into simple slices, maps, and scalars
// so the file remains readable as plain JSON and backward-compatible across
// schema additions (new optional fields default to zero/nil on old saves).
//
// Integrity fields:
//   - Signature: HMAC-SHA256 of the payload (excluding sig/proof). Tampered
//     saves produce a mismatch → cheaterBadge is set on load.
//   - Proof: HMAC of the Signature using forgeMasterKey. Present only when
//     the player has the elite badge; verifying this clears the cheater badge.
type GameSave struct {
	Timestamp time.Time             `json:"timestamp"`
	Tick      int                   `json:"tick"`
	Age       string                `json:"age"`
	Resources map[string]float64    `json:"resources"`
	Storage   map[string]float64    `json:"storage"`
	Buildings map[string]int        `json:"buildings"`
	Workers   map[string]WorkerInfo `json:"workers"`
	Unlocked  UnlockedState         `json:"unlocked"`
	Stats     *GameStats            `json:"stats"`
	// Phase 3 additions
	Research         ResearchSave                  `json:"research"`
	Military         MilitarySave                  `json:"military"`
	Events           EventSave                     `json:"events"`
	Milestones       []string                      `json:"milestones"`
	ChainsCompleted  []string                      `json:"chains_completed,omitempty"`
	CurrentTitle     string                        `json:"current_title,omitempty"`
	PermanentBonuses map[string]float64            `json:"permanent_bonuses"`
	BuildQueue       []BuildQueueItem              `json:"build_queue"`
	Prestige         PrestigeSave                  `json:"prestige"`
	Trade            TradeSave                     `json:"trade"`
	Diplomacy        DiplomacySave                 `json:"diplomacy"`
	SpeedMultiplier  float64                       `json:"speed_multiplier"`
	WonderBanks      map[string]map[string]float64 `json:"wonder_banks,omitempty"`
	// Phase 7: legacy building keys
	LegacyBuildings []string `json:"legacy_buildings,omitempty"`
	// Phase 8: epoch system
	CurrentEpoch       string             `json:"current_epoch,omitempty"`
	EpochEventFired    map[string]bool    `json:"epoch_event_fired,omitempty"`
	SurvivedEpochs     map[string]bool    `json:"survived_epochs,omitempty"`
	PendingCatastrophe string             `json:"pending_catastrophe,omitempty"`
	EpochEventHistory  []EpochEventRecord `json:"epoch_event_history,omitempty"`
	// Phase 9: catastrophe system (persist across Succumb and Prestige)
	Ruins              map[string]int  `json:"ruins,omitempty"`
	LegacyBonuses      map[string]bool `json:"legacy_bonuses,omitempty"`
	CatastropheHistory []string        `json:"catastrophe_history,omitempty"`
	// Morale system
	Morale float64 `json:"morale,omitempty"`
	// History overlay samples
	History *HistoryCollector `json:"history,omitempty"`
	// Integrity fields
	CheaterBadge bool   `json:"cheater_badge,omitempty"`
	EliteBadge   bool   `json:"elite_badge,omitempty"`
	// ParentName records the save this one branched from, for the save-lineage
	// tree (Phase 1: plumbed through but always "" — branching lands in Phase 2).
	// Legacy saves lack the field → "" → a root. omitempty keeps current saves
	// byte-identical when empty.
	ParentName string `json:"parent_name,omitempty"`
	Signature  string `json:"_sig,omitempty"`
	Proof      string `json:"_proof,omitempty"`
}

// signSave returns the HMAC-SHA256 hex of the save payload.
// Signature and Proof are zeroed before marshalling so the signature covers
// the game data only, not a prior signature value.
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
	Workers   []string `json:"workers"`
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

// AutosaveName is the reserved base-name of the automatic save file. The engine
// writes to it on its autosave cadence; the UI uses it to flag the autosave row
// in the Load Game browser so it can't be confused with a player-named save.
const AutosaveName = "autosave"

// SaveGame serialises current engine state to filename.json in the save directory.
// The save is written atomically (temp file + rename) to prevent corruption if
// the process is killed mid-write. The payload is HMAC-signed before writing.
// NOTE: SaveGame acquires only an RLock for the snapshot, so it can run
// concurrently with reads but not concurrent writes (doTick).
func (ge *GameEngine) SaveGame(filename string) error {
	dir := saveDirectory()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Take the read lock for the snapshot + marshal so doTick cannot mutate
	// maps/slices while json.Marshal is iterating over them.
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
		Workers:   ge.Workers.GetAll(),
		Unlocked:  ge.getUnlockedState(),
		Stats: &GameStats{
			TotalBuilt:      ge.Stats.TotalBuilt,
			TotalRecruited:  ge.Stats.TotalRecruited,
			TotalGathered:   statsGathered,
			GameStarted:     ge.Stats.GameStarted,
			AgesReached:     agesReached,
			SoldiersTrained: ge.Stats.SoldiersTrained,
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
		SpeedMultiplier:    ge.speedMultiplier,
		WonderBanks:        ge.Buildings.GetWonderBanks(),
		LegacyBuildings:    ge.Buildings.GetLegacyBuildings(),
		CheaterBadge:       ge.cheaterBadge,
		EliteBadge:         ge.eliteBadge,
		ParentName:         ge.activeParentName,
		CurrentEpoch:       ge.currentEpoch,
		EpochEventFired:    copyBoolMap(ge.epochEventFired),
		SurvivedEpochs:     copyBoolMap(ge.survivedEpochs),
		PendingCatastrophe: ge.pendingCatastrophe,
		EpochEventHistory:  append([]EpochEventRecord(nil), ge.epochEventHistory...),
		Ruins:              ge.Buildings.GetAllRuins(),
		LegacyBonuses:      copyBoolMap(ge.legacyBonuses),
		CatastropheHistory: append([]string(nil), ge.catastropheHistory...),
		Morale:             ge.morale,
		History:            ge.History,
	}
}

// LoadGame reads a save file, verifies its integrity, and restores all engine
// state under the write lock. File I/O is done outside the lock to avoid
// blocking doTick for the duration of a slow disk read.
// Integrity check: if the signature is present and invalid, cheaterBadge is
// set. If a valid elite proof is found, eliteBadge is set and cheaterBadge cleared.
func (ge *GameEngine) LoadGame(filename string) error {
	// File I/O outside the lock — avoids holding the write lock during disk access
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
	ge.Workers.SetAge(save.Age)
	ge.Resources.LoadAmounts(save.Resources)
	if save.Storage != nil {
		ge.Resources.LoadStorage(save.Storage)
	}
	ge.Buildings.LoadCounts(save.Buildings)
	ge.Workers.LoadWorkers(save.Workers)
	if save.Stats != nil {
		// Deep copy stats to avoid aliasing with the deserialized save
		gathered := make(map[string]float64, len(save.Stats.TotalGathered))
		for k, v := range save.Stats.TotalGathered {
			gathered[k] = v
		}
		ages := make([]string, len(save.Stats.AgesReached))
		copy(ages, save.Stats.AgesReached)
		ge.Stats = &GameStats{
			TotalBuilt:      save.Stats.TotalBuilt,
			TotalRecruited:  save.Stats.TotalRecruited,
			TotalGathered:   gathered,
			GameStarted:     save.Stats.GameStarted,
			AgesReached:     ages,
			SoldiersTrained: save.Stats.SoldiersTrained,
		}
	}
	ge.buildQueue = save.BuildQueue

	// Restore unlocks
	for _, key := range save.Unlocked.Resources {
		ge.Resources.UnlockResource(key)
	}
	// Reconcile resource unlocks against the loaded age. Old saves predating a
	// resource's introduction (e.g. the `soldiers` resource added in the military
	// rework) won't have it in their serialized unlock set even when the player is
	// already past its unlock age — unlock anything whose unlock-age has been reached.
	currentOrder := ge.progress.ageIndex[save.Age]
	for _, def := range config.BaseResources() {
		if order, ok := ge.progress.ageIndex[def.Age]; ok && order <= currentOrder {
			ge.Resources.UnlockResource(def.Key)
		}
	}
	for _, key := range save.Unlocked.Buildings {
		ge.Buildings.UnlockBuilding(key)
	}
	if len(save.Unlocked.Workers) > 0 {
		ge.Workers.UnlockType("worker")
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

	// Reconstruct pending upgrades from legacy buildings.
	// pendingUpgrades is not persisted to disk; we derive it from the
	// legacy set + current age so the 'upgrade' command works after a reload.
	for _, key := range save.LegacyBuildings {
		if ge.Buildings.GetCount(key) <= 0 {
			continue
		}
		def, ok := ge.Buildings.defs[key]
		if !ok || def.LineageKey == "" || def.LineageKey == "wonder" {
			continue
		}
		next := config.BuildingNextTierForAge(def.LineageKey, def.LineageTier, save.Age)
		if next == nil {
			continue
		}
		ge.Buildings.SetPendingUpgrade(key, next.Key)
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

	// Restore history collector
	if save.History != nil {
		ge.History = save.History
	} else {
		ge.History = NewHistoryCollector()
	}

	// Restore morale (default to 1.0 for old saves that have zero value)
	if save.Morale > 0 {
		ge.morale = save.Morale
	} else {
		ge.morale = 1.0
	}
	ge.lowMoraleWarned = false

	ge.recalculateRates()
	ge.recalculateTickSpeed()

	// Apply offline progress for time since save
	ge.applyOfflineProgress(time.Since(save.Timestamp))

	// Load succeeded: this is now the active slot a bare `save` writes to. We're
	// still under the write lock (ge.mu.Lock at the top of this function), so set
	// the field DIRECTLY — calling SetActiveSaveName would re-acquire the lock and
	// deadlock.
	ge.activeSaveName = filename
	// Adopt the loaded save's lineage parent (empty for legacy/root saves).
	ge.activeParentName = save.ParentName

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

// getUnlockedState collects all resource, building, and worker unlocks for
// the current age and all prior ages. On load, these lists are replayed to
// restore the Buildings.unlocked / Resources.unlocked maps without having to
// re-run the full age advance sequence.
// NOTE: Must be called with the lock held (called from buildSaveSnapshot).
func (ge *GameEngine) getUnlockedState() UnlockedState {
	state := UnlockedState{}
	for _, def := range ge.progress.ages {
		order := ge.progress.ageIndex[def.Key]
		currentOrder := ge.progress.ageIndex[ge.age]
		if order <= currentOrder {
			state.Resources = append(state.Resources, def.UnlockResources...)
			state.Buildings = append(state.Buildings, def.UnlockBuildings...)
			state.Workers = append(state.Workers, def.UnlockVillagers...)
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

// SaveInfo holds metadata about a save file, parsed from the save's JSON
// header without loading the full engine state. Fields beyond Name/Timestamp/Age
// drive the Load Game browser's detail pane.
type SaveInfo struct {
	Name          string
	Timestamp     time.Time
	Age           string
	Tick          int
	PrestigeLevel int
	Morale        float64
	Modified      bool // save's cheater_badge flag (HMAC mismatch on a prior load)
	Elite         bool // save's elite_badge flag
	// Rich detail-pane metadata (all derived from the save header; zero on a
	// corrupt save).
	Title              string // current_title — the civ's earned title (may be "")
	Epoch              string // current_epoch, mapped to its display name
	Population         int    // sum of worker Count across all domains (assigned + idle)
	Buildings          int    // total building instances built
	Wonders            int    // built buildings whose Category == "wonder"
	Techs              int    // number of researched techs
	Soldiers           int    // resources["soldiers"], truncated to int
	PrestigeTotal      int    // prestige.total_earned (lifetime prestige points)
	PendingCatastrophe string // pending_catastrophe → catastrophe display name ("" if none)
	MilestonesDone     int    // count of completed milestones
	MilestonesTotal    int    // total milestones defined in config (0 if unavailable)
	// ParentName is the lineage parent of this save ("" for a root). Surfaced
	// here so Phase 3 can assemble the save-tree without re-reading files.
	ParentName string
	// Corrupt is set when the file could not be read or JSON-parsed. The entry
	// is still returned (Name + mtime Timestamp) so the UI can show it as
	// greyed/unloadable rather than silently dropping it.
	Corrupt bool
}

// ListSaveDetails returns metadata for each save file. A file that fails to read
// or parse is not skipped: it is returned with Corrupt=true, its Name, and a
// Timestamp taken from the file's mtime, so a single bad file never breaks the
// listing for the rest.
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

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			saves = append(saves, corruptInfo(name, e))
			continue
		}
		// Parse only the subset of fields the detail pane needs. The json tags
		// here must match GameSave / PrestigeSave / ResearchSave / WorkerInfo above.
		var header struct {
			Timestamp time.Time `json:"timestamp"`
			Tick      int       `json:"tick"`
			Age       string    `json:"age"`
			Morale    float64   `json:"morale"`
			Prestige  struct {
				Level       int `json:"level"`
				TotalEarned int `json:"total_earned"`
			} `json:"prestige"`
			CurrentTitle       string                `json:"current_title"`
			CurrentEpoch       string                `json:"current_epoch"`
			PendingCatastrophe string                `json:"pending_catastrophe"`
			Milestones         []string              `json:"milestones"`
			Resources          map[string]float64    `json:"resources"`
			Buildings          map[string]int        `json:"buildings"`
			Workers            map[string]WorkerInfo `json:"workers"`
			Research           struct {
				Researched []string `json:"researched"`
			} `json:"research"`
			CheaterBadge bool   `json:"cheater_badge"`
			EliteBadge   bool   `json:"elite_badge"`
			ParentName   string `json:"parent_name"`
		}
		if err := json.Unmarshal(data, &header); err != nil {
			saves = append(saves, corruptInfo(name, e))
			continue
		}

		// Derive aggregates. Ranges over nil maps are no-ops, so a save missing
		// any of these sections simply yields zero — no panic.
		population := 0
		for _, w := range header.Workers {
			population += w.Count
		}
		buildings, wonders := 0, 0
		buildingDefs := config.BuildingByKey()
		for key, count := range header.Buildings {
			if count <= 0 {
				continue
			}
			buildings += count
			if def, ok := buildingDefs[key]; ok && def.Category == "wonder" {
				wonders += count
			}
		}

		saves = append(saves, SaveInfo{
			Name:               name,
			Timestamp:          header.Timestamp,
			Age:                header.Age,
			Tick:               header.Tick,
			PrestigeLevel:      header.Prestige.Level,
			Morale:             header.Morale,
			Modified:           header.CheaterBadge,
			Elite:              header.EliteBadge,
			Title:              header.CurrentTitle,
			Epoch:              epochDisplayName(header.CurrentEpoch),
			Population:         population,
			Buildings:          buildings,
			Wonders:            wonders,
			Techs:              len(header.Research.Researched),
			Soldiers:           int(header.Resources["soldiers"]),
			PrestigeTotal:      header.Prestige.TotalEarned,
			PendingCatastrophe: catastropheDisplayName(header.PendingCatastrophe),
			MilestonesDone:     len(header.Milestones),
			MilestonesTotal:    len(config.Milestones()),
			ParentName:         header.ParentName,
		})
	}
	return saves, nil
}

// epochDisplayName maps an epoch key to its display name via config, falling
// back to the raw key. An empty key stays empty.
func epochDisplayName(key string) string {
	if key == "" {
		return ""
	}
	if def, ok := config.EpochByKey()[key]; ok && def.Name != "" {
		return def.Name
	}
	return key
}

// catastropheDisplayName maps a pending-catastrophe key (an epoch key — see
// GameState.PendingCatastrophe) to the catastrophe's display name. An empty key
// stays empty so the UI can omit the warning line.
func catastropheDisplayName(epochKey string) string {
	if epochKey == "" {
		return ""
	}
	name, _ := config.CatastropheInfo(epochKey)
	return name
}

// corruptInfo builds a SaveInfo for an unreadable/unparseable save, falling back
// to the file's mtime for the Timestamp so the UI still has something to sort on.
func corruptInfo(name string, e os.DirEntry) SaveInfo {
	info := SaveInfo{Name: name, Corrupt: true}
	if fi, err := e.Info(); err == nil {
		info.Timestamp = fi.ModTime()
	}
	return info
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

// maxSaveNameLen bounds save base-names. Generous for player-chosen names while
// keeping us clear of filesystem path limits.
const maxSaveNameLen = 64

// sanitizeSaveName validates a save base-name that may originate from user input
// (e.g. the rename dialog) and returns the cleaned name. A trailing ".json" the
// user may have typed is stripped first. It REJECTS names that are empty/
// whitespace-only, contain a path separator ('/' or '\'), contain "..", begin
// with a dot, contain a null byte, or exceed maxSaveNameLen — anything that could
// escape the saves directory or produce a hidden/awkward file. The returned name
// is safe to join onto the saves directory.
func sanitizeSaveName(name string) (string, error) {
	// Strip a trailing .json the user may have included.
	if ext := filepath.Ext(name); ext == ".json" {
		name = name[:len(name)-len(ext)]
	}
	name = strings.TrimSpace(name)

	if name == "" {
		return "", fmt.Errorf("save name cannot be empty")
	}
	if len(name) > maxSaveNameLen {
		return "", fmt.Errorf("save name too long (max %d characters)", maxSaveNameLen)
	}
	if strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("save name contains an invalid character")
	}
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("save name cannot contain path separators")
	}
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("save name cannot contain '..'")
	}
	if strings.HasPrefix(name, ".") {
		return "", fmt.Errorf("save name cannot start with a dot")
	}
	return name, nil
}

// ValidateSaveName is the exported entry point to the same validation the rename
// and save paths use: it strips a trailing ".json", trims whitespace, and rejects
// names that are empty, too long, contain a null byte / path separator / "..", or
// begin with a dot. It returns the cleaned, filesystem-safe name. UI code (e.g. the
// New Game name prompt) calls this so the rule lives in exactly one place.
func ValidateSaveName(name string) (string, error) {
	return sanitizeSaveName(name)
}

// DeleteSave removes a save file. The name is sanitized first so a crafted value
// cannot escape the saves directory. Returns a clear error if the file is absent.
func DeleteSave(filename string) error {
	name, err := sanitizeSaveName(filename)
	if err != nil {
		return err
	}
	if !SaveExists(name) {
		return fmt.Errorf("save %q does not exist", name)
	}
	if err := os.Remove(savePath(name)); err != nil {
		return fmt.Errorf("failed to delete save: %w", err)
	}
	return nil
}

// RenameSave renames a save file from oldName to newName. Both names are
// sanitized. Errors if the source does not exist, if newName collides with an
// existing save, or if either name is invalid. Because only the filename changes
// (not the bytes), the save's HMAC signature stays valid.
func RenameSave(oldName, newName string) error {
	src, err := sanitizeSaveName(oldName)
	if err != nil {
		return fmt.Errorf("invalid source name: %w", err)
	}
	dst, err := sanitizeSaveName(newName)
	if err != nil {
		return fmt.Errorf("invalid new name: %w", err)
	}
	if !SaveExists(src) {
		return fmt.Errorf("save %q does not exist", src)
	}
	if src == dst {
		return fmt.Errorf("new name is the same as the current name")
	}
	if SaveExists(dst) {
		return fmt.Errorf("a save named %q already exists", dst)
	}
	srcPath := savePath(src)
	// Write the renamed file into the same directory the source lives in so a
	// save in the legacy CWD location isn't orphaned by a canonical-dir write.
	dstPath := filepath.Join(filepath.Dir(srcPath), dst+".json")
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to rename save: %w", err)
	}
	return nil
}

// DuplicateSave copies a save to a new name and returns the new base name. The
// new name is "<filename>-copy", or "<filename>-copy-2", "-copy-3", ... if a
// prior copy already exists. Bytes are copied verbatim, so the HMAC signature
// remains valid on the duplicate.
func DuplicateSave(filename string) (string, error) {
	src, err := sanitizeSaveName(filename)
	if err != nil {
		return "", err
	}
	if !SaveExists(src) {
		return "", fmt.Errorf("save %q does not exist", src)
	}

	// Find the first free "-copy" variant. The base + suffix must still satisfy
	// the sanitizer (length bound in particular), so validate before using it.
	dst := src + "-copy"
	for n := 2; SaveExists(dst); n++ {
		dst = fmt.Sprintf("%s-copy-%d", src, n)
	}
	if _, err := sanitizeSaveName(dst); err != nil {
		return "", fmt.Errorf("cannot build a valid copy name: %w", err)
	}

	srcPath := savePath(src)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read save: %w", err)
	}
	// Place the copy alongside the source (canonical or legacy dir) for the same
	// reason as RenameSave. Write atomically (temp + rename).
	dir := filepath.Dir(srcPath)
	dstPath := filepath.Join(dir, dst+".json")
	tmpPath := dstPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write copy: %w", err)
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to finalize copy: %w", err)
	}
	return dst, nil
}
