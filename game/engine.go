package game

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/user/ageforge/config"
)

const (
	BaseTickInterval = 2 * time.Second
	MinTickInterval  = 200 * time.Millisecond
	MaxLogSize       = 500
)

// GameEngine is the central game coordinator
type GameEngine struct {
	mu sync.RWMutex

	tick int
	age  string

	Resources  *ResourceManager
	Buildings  *BuildingManager
	Villagers  *VillagerManager
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
}

// BuildQueueItem represents a building under construction
type BuildQueueItem struct {
	BuildingKey string
	TicksLeft   int
	TotalTicks  int
}

// NewGameEngine creates a new game engine
func NewGameEngine() *GameEngine {
	ge := &GameEngine{
		age:              "primitive_age",
		Resources:        NewResourceManager(),
		Buildings:        NewBuildingManager(),
		Villagers:        NewVillagerManager(),
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
		stopCh:           make(chan struct{}),
	}
	ge.applyAgeUnlocks("primitive_age")
	// Give starting resources — enough for first hut + a little food
	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)
	// Startup tutorial
	ge.addLog("event", "Welcome to AgeForge! You have nothing but your hands.")
	ge.addLog("info", "[gold]Getting Started:[-]")
	ge.addLog("info", "  1. [cyan]gather wood[-] — collect wood by hand")
	ge.addLog("info", "  2. [cyan]gather food[-] — forage for food")
	ge.addLog("info", "  3. [cyan]build hut[-] — build shelter (costs 10 wood)")
	ge.addLog("info", "  4. [cyan]recruit worker[-] — recruit your first worker")
	ge.addLog("event", "★ Wonder available: Sacred Grove — build it to unlock +0.5x speed!")
	ge.addLog("info", "  5. [cyan]assign worker food[-] — put them to work!")
	ge.addLog("info", "  Type [cyan]help[-] for all commands.")
	return ge
}

const AutosaveInterval = 60 * time.Second

// Start begins the game tick loop
func (ge *GameEngine) Start() {
	ge.mu.Lock()
	ge.running = true
	ge.mu.Unlock()

	timer := time.NewTimer(ge.getTickInterval())
	defer timer.Stop()

	lastAutosave := time.Now()

	for {
		select {
		case <-timer.C:
			ge.safeTick()

			// Periodic autosave (outside the tick lock)
			if time.Since(lastAutosave) >= AutosaveInterval {
				if err := ge.SaveGame("autosave"); err != nil {
					ge.mu.Lock()
					ge.addLog("warning", fmt.Sprintf("Autosave failed: %v", err))
					ge.mu.Unlock()
				} else {
					ge.mu.Lock()
					ge.addLog("debug", "Autosave complete")
					ge.mu.Unlock()
				}
				lastAutosave = time.Now()
			}

			timer.Reset(ge.getTickInterval())
		case <-ge.stopCh:
			return
		}
	}
}

// safeTick wraps doTick with panic recovery to prevent the tick goroutine from dying
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

// getTickInterval returns the current tick interval based on tick speed bonus and speed multiplier
func (ge *GameEngine) getTickInterval() time.Duration {
	// tickSpeedBonus and speedMultiplier are only written under the write lock
	// in doTick/LoadGame, and this is called from the same goroutine after
	// doTick returns, so a direct read is safe here.
	bonus := ge.tickSpeedBonus
	mult := ge.speedMultiplier
	if mult < 1.0 {
		mult = 1.0
	}

	interval := time.Duration(float64(BaseTickInterval) / ((1.0 + bonus) * mult))
	if interval < MinTickInterval {
		interval = MinTickInterval
	}
	return interval
}

// recalculateTickSpeed sums all tick speed bonuses (must be called with lock held)
func (ge *GameEngine) recalculateTickSpeed() {
	oldBonus := ge.tickSpeedBonus
	bonus := ge.Research.GetBonus("tick_speed") +
		ge.permanentBonuses["tick_speed"] +
		ge.Prestige.GetBonuses()["tick_speed"]
	// Add active event tick_speed effects (e.g., milestone chain boosts)
	for _, eff := range ge.Events.GetActiveEffects() {
		if eff.Type == "tick_speed" {
			bonus += eff.Value
		}
	}
	ge.tickSpeedBonus = bonus

	if bonus != oldBonus {
		interval := time.Duration(float64(BaseTickInterval) / (1.0 + bonus))
		if interval < MinTickInterval {
			interval = MinTickInterval
		}
		ge.addLog("debug", fmt.Sprintf("Tick speed: +%.0f%% (interval: %dms)", bonus*100, interval.Milliseconds()))
	}
}

// MaxSpeedForAge returns the maximum speed multiplier based on wonders built.
// Base speed is 1.0x. Each wonder built adds +0.5x.
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

// Stop halts the game engine (safe to call multiple times)
func (ge *GameEngine) Stop() {
	ge.stopOnce.Do(func() {
		ge.mu.Lock()
		ge.running = false
		ge.mu.Unlock()
		close(ge.stopCh)
	})
}

// doTick processes one game tick
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

	// Apply resource rates (production - consumption)
	ge.Resources.ApplyRates()

	// Log net food rate and capped resources every 10 ticks
	if ge.tick%10 == 0 {
		snap := ge.Resources.Snapshot()
		if f, ok := snap["food"]; ok {
			ge.addLog("debug", fmt.Sprintf("Food: %.1f (rate %+.3f/t), pop=%d", f.Amount, f.Rate, ge.Villagers.TotalPop()))
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

	// Check food - starve if negative
	if ge.Resources.Get("food") <= 0 && ge.Villagers.FoodDrain() > 0 {
		ge.addLog("warning", "Your people are starving! Food has run out.")
	}

	// Periodic debug snapshot every 50 ticks
	if ge.tick%50 == 0 {
		snap := ge.Resources.Snapshot()
		foodAmt := snap["food"].Amount
		foodRate := snap["food"].Rate
		ge.addLog("debug", fmt.Sprintf("Tick %d snapshot: food=%.1f (%+.3f/t), pop=%d, queue=%d",
			ge.tick, foodAmt, foodRate, ge.Villagers.TotalPop(), len(ge.buildQueue)))
	}

	// Check milestones
	ge.checkMilestones()

	// Check age advancement
	if nextAge := ge.progress.CheckAdvancement(ge.age, ge.Resources, ge.Buildings); nextAge != "" {
		ge.advanceAge(nextAge)
	}

	// Recalculate tick speed from all sources
	ge.recalculateTickSpeed()
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
	triggered, expired := ge.Events.Tick(ge.tick, ge.age, ageOrder)

	for _, def := range triggered {
		ge.addLog("debug", fmt.Sprintf("Event triggered: %s (sentiment: %s)", def.Name, def.Sentiment))
		ge.addLog("event", def.LogMessage)
		// Process instant effects
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
			}
		}
	}

	for _, key := range expired {
		def := config.EventByKey()[key]
		ge.addLog("debug", fmt.Sprintf("Event expired: %s", key))
		ge.addLog("info", fmt.Sprintf("%s has ended.", def.Name))
	}
}

// processExpeditions handles military expedition progress
func (ge *GameEngine) processExpeditions() {
	prestigeBonuses := ge.Prestige.GetBonuses()
	militaryBonus := ge.Research.GetBonus("military_power") + ge.permanentBonuses["military_power"] + prestigeBonuses["military_power"]
	expeditionBonus := ge.Research.GetBonus("expedition_reward") + ge.permanentBonuses["expedition_reward"] + prestigeBonuses["expedition_reward"]

	if ge.Military.active != nil {
		ge.addLog("debug", fmt.Sprintf("Expedition: %s %d ticks left", ge.Military.active.Name, ge.Military.active.TicksLeft))
	}
	rewards, message, soldiersLost := ge.Military.Tick(militaryBonus, expeditionBonus)
	if message != "" {
		ge.addLog("debug", fmt.Sprintf("Expedition resolved (soldiers lost: %d, rewards: %d types)", soldiersLost, len(rewards)))
		ge.addLog("event", message)
		// Add rewards to resources
		for res, amount := range rewards {
			ge.Resources.Add(res, amount)
		}
		// Remove lost soldiers
		if soldiersLost > 0 {
			ge.Villagers.RemoveSoldiers(soldiersLost)
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
	messages := ge.Diplomacy.Tick(ge.age, ageOrder, ge.tick)
	for _, msg := range messages {
		ge.addLog("event", msg)
	}
}

// checkMilestones checks for newly completed milestones and chains
func (ge *GameEngine) checkMilestones() {
	ageOrder := ge.progress.GetAgeOrder()
	researchedTechs := make(map[string]bool)
	for _, key := range ge.Research.GetResearched() {
		researchedTechs[key] = true
	}

	// Count soldiers
	soldierCount := 0
	if st, ok := ge.Villagers.types["soldier"]; ok {
		soldierCount = st.count
	}

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
		ge.Villagers.TotalPop(),
		ge.Research.ResearchedCount(),
		ge.Stats.TotalBuilt,
		researchedTechs,
		soldierCount,
		wonderCount,
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

	// Building production
	for _, eff := range ge.Buildings.GetEffects() {
		if eff.Type == "production" {
			r := ge.Resources.resources[eff.Target]
			if r != nil {
				r.Rate += eff.Value
				r.Breakdown.BuildingRate += eff.Value
			}
		}
	}

	// Villager production
	for res, rate := range ge.Villagers.GetProductionRates() {
		r := ge.Resources.resources[res]
		if r != nil {
			r.Rate += rate
			r.Breakdown.VillagerRate += rate
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

	// Apply production_all bonus (multiplier on all positive rates)
	prodAllBonus := researchBonuses["production_all"] + permanentBonuses["production_all"]
	if prodAllBonus > 0 {
		for _, def := range ge.Resources.defs {
			r := ge.Resources.resources[def.Key]
			if r != nil && r.Rate > 0 {
				r.Rate *= (1.0 + prodAllBonus)
			}
		}
	}

	// Apply per-resource rate bonuses (e.g., "gold_rate", "iron_rate")
	for _, def := range ge.Resources.defs {
		bonusKey := def.Key + "_rate"
		bonus := researchBonuses[bonusKey] + permanentBonuses[bonusKey]
		if bonus > 0 {
			r := ge.Resources.resources[def.Key]
			if r != nil && r.Rate > 0 {
				r.Rate *= (1.0 + bonus)
			}
		}
	}

	// Apply gather_rate bonus to villager-generated rates
	gatherBonus := researchBonuses["gather_rate"] + permanentBonuses["gather_rate"]
	if gatherBonus > 0 {
		// Already applied via multiplier above
		// This is additive on base villager rates — re-add the bonus portion
		for res, rate := range ge.Villagers.GetProductionRates() {
			r := ge.Resources.resources[res]
			if r != nil {
				r.Rate += rate * gatherBonus
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
	drain := ge.Villagers.FoodDrain()
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
			knownComponents := r.Breakdown.BuildingRate + r.Breakdown.VillagerRate +
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

// advanceAge advances to the next age
func (ge *GameEngine) advanceAge(newAge string) {
	oldAge := ge.age
	ge.age = newAge
	ge.applyAgeUnlocks(newAge)
	ge.Stats.RecordAge(newAge)

	// Reduce all resources to 25% on age transition
	for _, r := range ge.Resources.resources {
		r.Amount *= 0.25
	}
	ge.addLog("info", "Age transition: resources reduced to 25%")

	oldName := ge.progress.GetAgeName(oldAge)
	newName := ge.progress.GetAgeName(newAge)
	unlocks := ge.progress.GetUnlocks(newAge)
	ge.addLog("debug", fmt.Sprintf("Age advance: %s → %s (unlocks: %d buildings, %d resources, %d villagers)",
		oldAge, newAge, len(unlocks.UnlockBuildings), len(unlocks.UnlockResources), len(unlocks.UnlockVillagers)))
	ge.addLog("success", fmt.Sprintf("Advanced from %s to %s!", oldName, newName))

	// Notify player about the wonder available in this age
	for _, bKey := range unlocks.UnlockBuildings {
		if def, ok := ge.Buildings.defs[bKey]; ok && def.Category == "wonder" {
			ge.addLog("event", fmt.Sprintf("★ Wonder available: %s — build it to unlock +0.5x speed!", def.Name))
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
		ge.Villagers.UnlockType(v)
	}
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

// GatherResource manually gathers a resource
func (ge *GameEngine) GatherResource(resource string, amount float64) (float64, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

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
	if def.MaxCount > 0 && ge.Buildings.GetCount(key) >= def.MaxCount {
		return fmt.Errorf("%s is at max count (%d)", def.Name, def.MaxCount)
	}

	// Check if already building this (for unique buildings)
	for _, item := range ge.buildQueue {
		if item.BuildingKey == key && def.MaxCount > 0 {
			return fmt.Errorf("%s is already under construction (%d ticks left)", def.Name, item.TicksLeft)
		}
	}

	cost := ge.Buildings.GetCost(key)
	if !ge.Resources.Pay(cost) {
		return fmt.Errorf("cannot afford %s (need: %s)", def.Name, formatCost(cost))
	}

	ge.addLog("debug", fmt.Sprintf("Build start: %s (cost: %s)", def.Name, formatCost(cost)))
	if def.BuildTicks > 0 {
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

	built := 0
	for i := 0; i < count; i++ {
		if def.MaxCount > 0 && ge.Buildings.GetCount(key) >= def.MaxCount {
			break
		}
		// For unique buildings, check build queue
		if def.MaxCount > 0 {
			inQueue := false
			for _, item := range ge.buildQueue {
				if item.BuildingKey == key {
					inQueue = true
					break
				}
			}
			if inQueue {
				break
			}
		}

		cost := ge.Buildings.GetCost(key)
		if !ge.Resources.Pay(cost) {
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
		if def.MaxCount > 0 && ge.Buildings.GetCount(key) >= def.MaxCount {
			return 0, fmt.Errorf("%s is at max count (%d)", def.Name, def.MaxCount)
		}
		cost := ge.Buildings.GetCost(key)
		return 0, fmt.Errorf("cannot afford %s (need: %s)", def.Name, formatCost(cost))
	}

	if def.BuildTicks > 0 {
		ge.addLog("info", fmt.Sprintf("Queued %d %s for construction", built, def.Name))
	} else {
		ge.recalculateRates()
		ge.addLog("success", fmt.Sprintf("Built %d %s (total: %d)", built, def.Name, ge.Buildings.GetCount(key)))
	}
	return built, nil
}

// RecruitMax recruits as many villagers as possible up to the pop cap
func (ge *GameEngine) RecruitMax(vType string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if !ge.Villagers.IsUnlocked(vType) {
		return 0, fmt.Errorf("villager type '%s' is not yet unlocked", vType)
	}

	popCap := ge.Buildings.GetPopCapacity()
	popCap += int(ge.Research.GetBonus("population") + ge.permanentBonuses["population"] + ge.Prestige.GetBonuses()["population"])

	available := popCap - ge.Villagers.TotalPop()
	if available <= 0 {
		return 0, fmt.Errorf("population cap reached (%d/%d)", ge.Villagers.TotalPop(), popCap)
	}

	if !ge.Villagers.Recruit(vType, available, popCap) {
		return 0, fmt.Errorf("cannot recruit %s(s)", vType)
	}
	ge.Stats.RecordRecruit(available)
	ge.addLog("info", fmt.Sprintf("Recruited %d %s(s) (pop: %d/%d)", available, vType, ge.Villagers.TotalPop(), popCap))
	return available, nil
}

// RecruitVillager recruits villagers
func (ge *GameEngine) RecruitVillager(vType string, count int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	popCap := ge.Buildings.GetPopCapacity()
	// Add population capacity from research/milestones/prestige
	popCap += int(ge.Research.GetBonus("population") + ge.permanentBonuses["population"] + ge.Prestige.GetBonuses()["population"])

	if !ge.Villagers.Recruit(vType, count, popCap) {
		totalPop := ge.Villagers.TotalPop()
		if !ge.Villagers.IsUnlocked(vType) {
			return fmt.Errorf("villager type '%s' is not yet unlocked", vType)
		}
		return fmt.Errorf("cannot recruit %d %s(s) (pop: %d/%d)", count, vType, totalPop, popCap)
	}
	ge.Stats.RecordRecruit(count)
	ge.addLog("debug", fmt.Sprintf("Recruit: %d %s (pop: %d/%d)", count, vType, ge.Villagers.TotalPop(), popCap))
	ge.addLog("info", fmt.Sprintf("Recruited %d %s(s)", count, vType))
	return nil
}

// AssignVillager assigns villagers to gather a resource
func (ge *GameEngine) AssignVillager(vType, resource string, count int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if !ge.Villagers.Assign(vType, resource, count) {
		idle := ge.Villagers.IdleCount(vType)
		return fmt.Errorf("cannot assign %d %s(s) to %s (idle: %d)", count, vType, resource, idle)
	}
	ge.recalculateRates()
	ge.addLog("debug", fmt.Sprintf("Assign: %d %s → %s", count, vType, resource))
	ge.addLog("info", fmt.Sprintf("Assigned %d %s(s) to %s", count, vType, resource))
	return nil
}

// AssignAll assigns all idle villagers of a type to a resource
func (ge *GameEngine) AssignAll(vType, resource string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	idle := ge.Villagers.IdleCount(vType)
	if idle <= 0 {
		return 0, fmt.Errorf("no idle %s(s) to assign", vType)
	}
	if !ge.Villagers.Assign(vType, resource, idle) {
		return 0, fmt.Errorf("cannot assign %s(s) to %s", vType, resource)
	}
	ge.recalculateRates()
	ge.addLog("info", fmt.Sprintf("Assigned all %d %s(s) to %s", idle, vType, resource))
	return idle, nil
}

// UnassignAll removes all villagers of a type from a resource
func (ge *GameEngine) UnassignAll(vType, resource string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	st, ok := ge.Villagers.types[vType]
	if !ok {
		return 0, fmt.Errorf("unknown villager type: %s", vType)
	}
	assigned := st.assignment[resource]
	if assigned <= 0 {
		return 0, fmt.Errorf("no %s(s) assigned to %s", vType, resource)
	}
	if !ge.Villagers.Unassign(vType, resource, assigned) {
		return 0, fmt.Errorf("cannot unassign %s(s) from %s", vType, resource)
	}
	ge.recalculateRates()
	ge.addLog("info", fmt.Sprintf("Unassigned all %d %s(s) from %s", assigned, vType, resource))
	return assigned, nil
}

// UnassignVillager removes villagers from a resource assignment
func (ge *GameEngine) UnassignVillager(vType, resource string, count int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if !ge.Villagers.Unassign(vType, resource, count) {
		return fmt.Errorf("cannot unassign %d %s(s) from %s", count, vType, resource)
	}
	ge.recalculateRates()
	ge.addLog("debug", fmt.Sprintf("Unassign: %d %s ← %s", count, vType, resource))
	ge.addLog("info", fmt.Sprintf("Unassigned %d %s(s) from %s", count, vType, resource))
	return nil
}

// StartResearch begins researching a technology
func (ge *GameEngine) StartResearch(techKey string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	knowledge := ge.Resources.Get("knowledge")

	if err := ge.Research.StartResearch(techKey, ge.age, ageOrder, knowledge); err != nil {
		return err
	}

	// Pay knowledge cost
	def := config.TechByKey()[techKey]
	ge.Resources.Remove("knowledge", def.Cost)
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

// LaunchExpedition starts a military expedition
func (ge *GameEngine) LaunchExpedition(key string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	soldierCount := 0
	if st, ok := ge.Villagers.types["soldier"]; ok {
		soldierCount = st.count
	}
	militaryBonus := ge.Research.GetBonus("military_power") + ge.permanentBonuses["military_power"] + ge.Prestige.GetBonuses()["military_power"]

	if err := ge.Military.LaunchExpedition(key, soldierCount, ge.age, ageOrder, militaryBonus); err != nil {
		return err
	}

	ge.addLog("debug", fmt.Sprintf("Expedition start: %s (soldiers: %d, bonus: %.1f%%)", ge.Military.active.Name, soldierCount, militaryBonus*100))
	ge.addLog("info", fmt.Sprintf("Expedition launched: %s", ge.Military.active.Name))
	return nil
}

// DoPrestige resets the game with prestige bonuses
func (ge *GameEngine) DoPrestige() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	if !ge.Prestige.CanPrestige(ge.age, ageOrder) {
		return fmt.Errorf("must reach Medieval Age or later to prestige")
	}

	points := ge.Prestige.CalculatePoints(
		ge.age, ageOrder,
		ge.Milestones.CompletedCount(),
		ge.Research.ResearchedCount(),
		ge.Stats.TotalBuilt,
	)

	ge.Prestige.Prestige(points)

	// Reset all game systems
	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManager()
	ge.Buildings = NewBuildingManager()
	ge.Villagers = NewVillagerManager()
	ge.Research = NewResearchManager()
	ge.Military = NewMilitaryManager()
	ge.Events = NewEventManager()
	ge.Milestones = NewMilestoneManager()
	ge.Trade = NewTradeManager()
	ge.Diplomacy = NewDiplomacyManager()
	ge.Stats = NewGameStats()
	ge.Bus = NewEventBus()
	ge.permanentBonuses = make(map[string]float64)
	ge.buildQueue = nil
	ge.log = nil

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
	ge.addLog("info", "Type [cyan]help[-] to get started again.")

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
	ge.Villagers = NewVillagerManager()
	ge.Research = NewResearchManager()
	ge.Military = NewMilitaryManager()
	ge.Events = NewEventManager()
	ge.Milestones = NewMilestoneManager()
	ge.Prestige = NewPrestigeManager()
	ge.Trade = NewTradeManager()
	ge.Diplomacy = NewDiplomacyManager()
	ge.Stats = NewGameStats()
	ge.Bus = NewEventBus()
	ge.permanentBonuses = make(map[string]float64)
	ge.tickSpeedBonus = 0
	ge.speedMultiplier = 1.0
	ge.buildQueue = nil
	ge.log = nil

	ge.applyAgeUnlocks("primitive_age")
	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)

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
	soldierCount := 0
	if st, ok := ge.Villagers.types["soldier"]; ok {
		soldierCount = st.count
	}
	prestigeBonuses := ge.Prestige.GetBonuses()
	militaryBonus := ge.Research.GetBonus("military_power") + ge.permanentBonuses["military_power"] + prestigeBonuses["military_power"]
	expeditionBonus := ge.Research.GetBonus("expedition_reward") + ge.permanentBonuses["expedition_reward"] + prestigeBonuses["expedition_reward"]

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

	return GameState{
		Tick:             ge.tick,
		Age:              ge.age,
		AgeName:          ge.progress.GetAgeName(ge.age),
		NextAge:          nextAge,
		NextAgeName:      nextAgeName,
		NextAgeResReqs:   nextAgeResReqs,
		NextAgeBldReqs:   nextAgeBldReqs,
		Resources:        ge.Resources.Snapshot(),
		Buildings:        ge.Buildings.Snapshot(ge.Resources),
		BuildQueue:       queue,
		Villagers:        ge.Villagers.Snapshot(popCap),
		Research:         ge.Research.Snapshot(ge.age, ageOrder),
		Military:         ge.Military.Snapshot(ge.age, ageOrder, soldierCount, militaryBonus, expeditionBonus),
		Milestones: ge.Milestones.Snapshot(MilestoneSnapshotParams{
			Tick:            ge.tick,
			Age:             ge.age,
			AgeOrder:        ageOrder,
			Resources:       ge.Resources.GetAll(),
			Buildings:       ge.Buildings.GetAll(),
			Population:      ge.Villagers.TotalPop(),
			TechCount:       ge.Research.ResearchedCount(),
			TotalBuilt:      ge.Stats.TotalBuilt,
			SoldierCount:    soldierCount,
			WonderCount:     ge.countWonders(),
			ResearchedTechs: ge.getResearchedTechMap(),
			activeEvents:    ge.Events.GetActive(),
		}),
		ActiveEvents:     ge.Events.GetActive(),
		Prestige:         prestigeSnap,
		Trade:            ge.Trade.Snapshot(ge.age, ageOrder, ge.Buildings),
		Diplomacy:        ge.Diplomacy.Snapshot(ge.age, ageOrder),
		Log:              logCopy,
		Stats:            ge.Stats.Snapshot(),
		SaveExists:       SaveExists("autosave"),
		TickSpeedBonus:   ge.tickSpeedBonus,
		TickIntervalMs:   int(tickInterval.Milliseconds()),
		SpeedMultiplier:  speedMult,
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

// UpgradeBuilding upgrades all buildings of fromKey to the next tier.
// Returns the number upgraded and any error.
func (ge *GameEngine) UpgradeBuilding(fromKey string) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	upgrades := config.UpgradesFromKey()
	upg, ok := upgrades[fromKey]
	if !ok {
		return 0, fmt.Errorf("no upgrade path for '%s'", fromKey)
	}

	// Check age requirement
	ageOrder := ge.progress.GetAgeOrder()
	currentOrder, ok1 := ageOrder[ge.age]
	requiredOrder, ok2 := ageOrder[upg.MinAge]
	if !ok1 || !ok2 || currentOrder < requiredOrder {
		return 0, fmt.Errorf("upgrade requires %s", ge.progress.GetAgeName(upg.MinAge))
	}

	// Ensure target building is unlocked
	if !ge.Buildings.IsUnlocked(upg.To) {
		ge.Buildings.UnlockBuilding(upg.To)
	}

	// Get base cost of target building, scaled by upgrade discount
	toDef, toExists := ge.Buildings.defs[upg.To]
	if !toExists {
		return 0, fmt.Errorf("target building '%s' not found", upg.To)
	}

	count := ge.Buildings.counts[fromKey]
	if count <= 0 {
		return 0, fmt.Errorf("no %s to upgrade", fromKey)
	}

	upgraded := 0
	for i := 0; i < count; i++ {
		// Check max count on target
		if toDef.MaxCount > 0 && ge.Buildings.counts[upg.To] >= toDef.MaxCount {
			break
		}
		// Calculate discounted cost
		cost := make(map[string]float64)
		for res, base := range toDef.BaseCost {
			cost[res] = base * upg.CostScale
		}
		if !ge.Resources.Pay(cost) {
			break
		}
		ge.Buildings.counts[fromKey]--
		ge.Buildings.counts[upg.To]++
		upgraded++
	}

	if upgraded == 0 {
		cost := make(map[string]float64)
		for res, base := range toDef.BaseCost {
			cost[res] = base * upg.CostScale
		}
		return 0, fmt.Errorf("cannot afford upgrade to %s (need: %s)", toDef.Name, formatCost(cost))
	}

	fromDef := ge.Buildings.defs[fromKey]
	ge.recalculateRates()
	ge.addLog("success", fmt.Sprintf("Upgraded %d %s → %s", upgraded, fromDef.Name, toDef.Name))
	return upgraded, nil
}

// UpgradeAll upgrades all affordable buildings across all chains.
// Returns a map of fromKey -> count upgraded.
func (ge *GameEngine) UpgradeAll() (map[string]int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ageOrder := ge.progress.GetAgeOrder()
	currentOrder := ageOrder[ge.age]
	upgrades := config.BuildingUpgrades()
	result := make(map[string]int)

	for _, upg := range upgrades {
		requiredOrder, ok := ageOrder[upg.MinAge]
		if !ok || currentOrder < requiredOrder {
			continue
		}

		count := ge.Buildings.counts[upg.From]
		if count <= 0 {
			continue
		}

		toDef, toExists := ge.Buildings.defs[upg.To]
		if !toExists {
			continue
		}

		if !ge.Buildings.IsUnlocked(upg.To) {
			ge.Buildings.UnlockBuilding(upg.To)
		}

		for i := 0; i < count; i++ {
			if toDef.MaxCount > 0 && ge.Buildings.counts[upg.To] >= toDef.MaxCount {
				break
			}
			cost := make(map[string]float64)
			for res, base := range toDef.BaseCost {
				cost[res] = base * upg.CostScale
			}
			if !ge.Resources.Pay(cost) {
				break
			}
			ge.Buildings.counts[upg.From]--
			ge.Buildings.counts[upg.To]++
			result[upg.From]++
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("nothing to upgrade (no affordable upgrades available)")
	}

	ge.recalculateRates()
	for from, n := range result {
		fromDef := ge.Buildings.defs[from]
		upg := config.UpgradesFromKey()[from]
		toDef := ge.Buildings.defs[upg.To]
		ge.addLog("success", fmt.Sprintf("Upgraded %d %s → %s", n, fromDef.Name, toDef.Name))
	}
	return result, nil
}

// GetAvailableUpgrades returns upgrade info for display
func (ge *GameEngine) GetAvailableUpgrades() []UpgradeInfo {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	ageOrder := ge.progress.GetAgeOrder()
	currentOrder := ageOrder[ge.age]
	upgrades := config.BuildingUpgrades()
	var result []UpgradeInfo

	for _, upg := range upgrades {
		requiredOrder, ok := ageOrder[upg.MinAge]
		if !ok || currentOrder < requiredOrder {
			continue
		}

		count := ge.Buildings.counts[upg.From]
		if count <= 0 {
			continue
		}

		fromDef := ge.Buildings.defs[upg.From]
		toDef := ge.Buildings.defs[upg.To]
		cost := make(map[string]float64)
		for res, base := range toDef.BaseCost {
			cost[res] = base * upg.CostScale
		}

		canAfford := true
		for res, need := range cost {
			if ge.Resources.Get(res) < need {
				canAfford = false
				break
			}
		}

		result = append(result, UpgradeInfo{
			FromKey:   upg.From,
			ToKey:     upg.To,
			FromName:  fromDef.Name,
			ToName:    toDef.Name,
			Count:     count,
			Cost:      cost,
			CanAfford: canAfford,
		})
	}
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
