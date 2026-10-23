package game

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// Automatic expedition dispatch (Phase 3 of the faction redesign).
//
// The faction engine is fed by RESOLVED expeditions: every one rolls an encounter,
// which is what discovers civilizations and pays out timed boons. That made the
// whole system hostage to a player sitting at the prompt typing `expedition`. The
// Geographic Society (config/buildings.go, industrial_age) fixes that: once built
// it keeps survey parties on standing orders and sends them out by itself.
//
// Shape of the mechanic, and the rules it must not break:
//
//   - It models the repeating trade route (game/trade.go): a TicksLeft countdown
//     that resets on completion. That is the engine's one existing pattern for a
//     recurring automated action and there was no reason to invent a second.
//   - It sends an offer off the mission board: the one with the best haul in
//     minutes of income for each minute of length, among those it can pay for
//     (pickAutoOffer).
//   - It respects the army. An auto-dispatch only sends an offer whose party
//     the soldiers at home can make up. Nothing is queued.
//   - It pays the full resource cost through the same launch path a player uses
//     (launchExpeditionLocked). An unaffordable cycle is skipped, not discounted.
//   - It is DELIBERATELY SLOWER THAN HANDS-ON PLAY. See the interval constants
//     below; the engagement gradient this preserves is measured in
//     boon_tuning_test.go and is the whole reason the floor exists.
//
// Everything here runs inside doTick under the engine write lock (called from
// processExpeditions), so it may touch ge.Buildings / ge.Workers / ge.Military /
// ge.Resources directly and must never call a lock-acquiring method.

const (
	// autoExpeditionBuildingKey is the building that grants automatic dispatch.
	// The mechanic keys off the building KEY rather than an effect, the same way
	// the trade system keys off "market"/"port" — see the def in
	// config/buildings.go for why it carries no Effects entry.
	autoExpeditionBuildingKey = "geographic_society"

	// --- Cadence -------------------------------------------------------------
	// The interval between automatic dispatches is
	//
	//	interval = round(base × (1 − workerRelief × fill) ÷ count)   [floored]
	//
	// where count is the number of societies standing and fill is their aggregate
	// worker fill (0..1). One unstaffed society is barely better than a player who
	// wanders past every so often; investment buys cadence, and the floor stops
	// investment from buying back-to-back play.
	//
	// CALIBRATION (measured — see TestBoonTuning_Scenarios, scenarios A and F).
	// A resolved mission rolls an encounter (doubled, capped at 0.6), so the
	// encounter rate follows the resolutions per 10k ticks. The era scout is
	// out for 150 to 900 ticks, and a dispatch cycle is max(interval, that
	// duration) because the countdown runs from launch and the slot must be
	// free. Atomic Age, as measured:
	//
	//	hands-on player (both categories, back-to-back):  ~6.9 enc/10k, 90% uptime
	//	casual player (a scout every ~1000 ticks):        ~2.1 enc/10k, 49% uptime
	//	FULL automation at the floor (780-tick cycle):    ~4.0 enc/10k, 76% uptime
	//
	// which is ~58% of the hands-on rate: automation keeps an idle empire fed
	// without making the keyboard pointless. Move autoExpeditionMinInterval and
	// that ratio moves inversely and immediately — re-run the tuning harness.
	//
	// autoExpeditionBaseInterval is the wait for ONE society with no staff.
	// Deliberately just under the ~1000-tick gap the "casual" tuning scenario
	// models: your first society is worth about as much as remembering to send a
	// party yourself now and then.
	autoExpeditionBaseInterval = 900
	// autoExpeditionWorkerRelief is the largest fraction a fully-staffed society
	// cuts off the interval. Staffing is worth about a third; copies are worth more,
	// which keeps the building itself the thing you invest in.
	autoExpeditionWorkerRelief = 0.35
	// autoExpeditionMinInterval is the FLOOR — the fastest an automated empire can
	// ever dispatch, no matter how many societies are staffed. This single number
	// holds the engagement gradient: at 300 ticks a fully-invested idler sits at
	// ~58% of a continuously-exploring player's encounter rate. The era scouts
	// are out for 5 to 30 minutes (150 to 900 ticks), so a floor under their
	// mean length (100 before the era missions) bound nothing and put the idler
	// at 87%. The floor binds at 2 fully-staffed societies (900 × 0.65 ÷ 2 ≈ 293);
	// past that, further copies buy only resilience against a party coming home
	// late.
	autoExpeditionMinInterval = 300
)

// autoExpeditionInvestment returns how many Geographic Societies stand and their
// aggregate worker fill (0..1). fill mirrors the production/embassy worker curve's
// notion of fill: assigned ÷ (count × WorkerCapacity), clamped to 1.
func (ge *GameEngine) autoExpeditionInvestment() (count int, fill float64) {
	count = ge.Buildings.GetCount(autoExpeditionBuildingKey)
	if count <= 0 {
		return 0, 0
	}
	// Read the def off the BuildingManager's cached map, NOT config.BuildingByKey():
	// that helper reconstructs and re-normalizes all 284 building defs on every call,
	// and this runs every tick for the life of the run.
	def, ok := ge.Buildings.defs[autoExpeditionBuildingKey]
	if !ok || def.WorkerCapacity <= 0 {
		return count, 0
	}
	totalCap := float64(count * def.WorkerCapacity)
	assigned := float64(ge.Workers.GetAssignedCount("worker", autoExpeditionBuildingKey))
	fill = assigned / totalCap
	if fill > 1 {
		fill = 1
	}
	if fill < 0 {
		fill = 0
	}
	return count, fill
}

// autoExpeditionSnapshot builds the UI-facing view of automatic dispatch.
//
// It lives HERE, beside the fields and the building key it reads, so that
// GetState — and through it every panel — never has to know
// autoExpeditionBuildingKey or reach into state.Buildings for it. Before this
// existed none of automatic dispatch was on GameState at all, which is why a
// built society was invisible to the player.
//
// Caller must hold the engine lock (GetState's read lock suffices): this reads
// ge.Buildings / ge.Workers / the countdown fields directly and acquires
// nothing itself.
func (ge *GameEngine) autoExpeditionSnapshot() AutoExpeditionState {
	count, fill := ge.autoExpeditionInvestment()
	if count <= 0 {
		// Nothing built: automation is off, and a zero value says so without
		// the UI having to special-case a stale countdown.
		return AutoExpeditionState{}
	}
	capacity := 0
	// Same reason as autoExpeditionInvestment for reading defs directly rather
	// than config.BuildingByKey(): that helper rebuilds all 284 defs per call.
	if def, ok := ge.Buildings.defs[autoExpeditionBuildingKey]; ok && def.WorkerCapacity > 0 {
		capacity = count * def.WorkerCapacity
	}
	return AutoExpeditionState{
		Active:    true,
		TicksLeft: ge.autoExpeditionTicksLeft,
		Interval:  autoExpeditionInterval(ge.rules, ge.age, count, fill),
		Starved:   ge.autoExpeditionStarved,
		Count:     count,
		Assigned:  ge.Workers.GetAssignedCount("worker", autoExpeditionBuildingKey),
		Capacity:  capacity,
	}
}

// autoExpeditionIntervalFor is the pure cadence formula: ticks between automatic
// dispatches for a given investment. Returns 0 when nothing is built (automation
// off). Never returns a value below autoExpeditionMinInterval otherwise, and is
// panic-safe for any input (a NaN/absurd fill is clamped).
func autoExpeditionIntervalFor(count int, fill float64) int {
	if count <= 0 {
		return 0
	}
	if math.IsNaN(fill) || fill < 0 {
		fill = 0
	}
	if fill > 1 {
		fill = 1
	}
	relief := 1 - float64(autoExpeditionWorkerRelief*fill)
	interval := int(math.Round(autoExpeditionBaseInterval * relief / float64(count)))
	if interval < autoExpeditionMinInterval {
		interval = autoExpeditionMinInterval
	}
	return interval
}

// autoExpeditionInterval is the dispatch interval in age: the cadence
// formula, typed for the base curve, stretched like the age
// (set.StretchTicks). Expeditions stretch the same way, so the society
// still sends as many parties per age, and the engagement gradient measured
// in boon_tuning_test.go holds.
func autoExpeditionInterval(set *rules.Set, age string, count int, fill float64) int {
	return set.StretchTicks(age, autoExpeditionIntervalFor(count, fill))
}

// pickAutoOffer chooses which offer the society sends.
//
// RULE: among the offers on the board it can send right now (the town has the
// party at home and can pay the cost), the one with the
// best haul minutes per minute of length, among the short offers only: the
// Society sends the era's scout, and long and grand missions stay the
// player's call. Ties go to the bigger haul, then to
// the lower number, so the choice is fully deterministic and needs no RNG.
//
// Returns 0 when nothing can be sent right now.
func (ge *GameEngine) pickAutoOffer() int {
	best, bestRate, bestHaul := 0, 0.0, 0.0
	for i, slot := range ge.Military.board.Slots {
		offer := slot.Offer
		if offer == nil || offer.LengthMinutes <= 0 {
			continue
		}
		// The Society sends the era's scout: long and grand offers are the
		// player's call.
		if offer.Tier != TierShort {
			continue
		}
		if _, err := ge.offerRefusalLocked(*offer, ge.missionCostLocked(*offer)); err != nil {
			continue
		}
		rate := offer.HaulMinutes / offer.LengthMinutes
		if best == 0 || rate > bestRate || (rate == bestRate && offer.HaulMinutes > bestHaul) {
			best, bestRate, bestHaul = i+1, rate, offer.HaulMinutes
		}
	}
	return best
}

// processAutoExpeditions runs one tick of the Geographic Society's standing
// orders. Called from processExpeditions AFTER the active expeditions have been
// ticked, so a party that came home this tick frees the slot for the next one
// immediately — which makes the real dispatch cycle exactly
// max(interval, expedition duration).
//
// Must be called under the engine write lock.
func (ge *GameEngine) processAutoExpeditions() {
	count, fill := ge.autoExpeditionInvestment()
	if count == 0 {
		// No society (or all of them demolished): automation is off and any
		// pending countdown / warning state is dropped.
		ge.autoExpeditionTicksLeft = 0
		ge.autoExpeditionStarved = false
		return
	}

	interval := autoExpeditionInterval(ge.rules, ge.age, count, fill)

	// Investment can rise mid-countdown (a new society finishes, workers are
	// assigned). Clamp the outstanding wait DOWN to the new interval so the
	// improvement is felt at once; never clamp it up, so losing staff cannot
	// retroactively lengthen a wait already served.
	if ge.autoExpeditionTicksLeft > interval {
		ge.autoExpeditionTicksLeft = interval
	}

	if ge.autoExpeditionTicksLeft > 0 {
		ge.autoExpeditionTicksLeft--
		if ge.autoExpeditionTicksLeft > 0 {
			return
		}
	}

	// The countdown is spent: a dispatch is DUE. It may still not happen — if
	// every slot is busy or nothing is affordable we hold at zero and retry next
	// tick rather than forfeiting the cycle.
	ge.ensureMissionBoardLocked()
	n := ge.pickAutoOffer()
	if n == 0 {
		// Either the board has nothing the town can send, or the stores are too
		// thin to outfit anything. Warn ONCE per dry spell (state change only) —
		// this is checked every tick and would otherwise bury the log.
		if !ge.autoExpeditionStarved {
			ge.autoExpeditionStarved = true
			ge.addLog("warning", vibe.Sprintf("game.auto_expedition.processAutoExpeditions.01", "%s: a survey party stands ready, but there are no supplies to outfit it.",
				ge.autoExpeditionBuildingName()))
		}
		return
	}

	offer, _ := ge.Military.BoardOffer(n)
	if err := ge.launchExpeditionLocked(n); err != nil {
		// Defensive: the affordability and slot checks above already cover every
		// rejection the launch path can produce. Hold at zero and retry.
		ge.addLog("debug", fmt.Sprintf("auto-dispatch declined: %v", err))
		return
	}

	ge.autoExpeditionStarved = false
	ge.autoExpeditionTicksLeft = interval
	ge.addLog(LogRoutine, vibe.Sprintf("game.auto_expedition.processAutoExpeditions.03", "%s dispatches a survey party: %s.", ge.autoExpeditionBuildingName(), offer.Title))
}

// autoExpeditionBuildingName is the society's player-facing name, read from the
// loaded defs so the log lines cannot drift from the building definition.
func (ge *GameEngine) autoExpeditionBuildingName() string {
	if def, ok := ge.Buildings.defs[autoExpeditionBuildingKey]; ok && def.Name != "" {
		return def.Name
	}
	return vibe.Line("game.auto_expedition.autoExpeditionBuildingName.01", "The survey office")
}
