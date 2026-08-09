package game

import (
	"fmt"
	"math"
	"sort"
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
//   - SCOUTING ONLY. Scouting is what feeds encounters; automating conquest is a
//     different feature with a different risk profile and is out of scope.
//   - It respects the existing single-slot-per-category scheduler. An auto-dispatch
//     happens only when the scouting slot is genuinely FREE. Nothing is queued and
//     the one-active-per-category rule is never bypassed.
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
	// A resolved scouting run rolls an encounter at encounterChanceScoutSuccess /
	// Fail, so the encounter rate is (resolutions per 10k ticks) × ~0.156. The
	// shortest scouting expedition available from the bronze age on runs 60-100
	// ticks, and a dispatch cycle is max(interval, that duration) because the
	// countdown runs from launch and the slot must be free. Therefore:
	//
	//	hands-on player (both categories, back-to-back):  ~26 enc/10k, 97.5% uptime
	//	casual player (a scout every ~1000 ticks):        ~1.4 enc/10k, 15.8% uptime
	//	FULL automation at the floor (100-tick cycle):    ~15.6 enc/10k
	//
	// which is ~60% of the hands-on rate: automation keeps an idle empire fed
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
	// holds the engagement gradient: at 100 ticks a fully-invested idler sits at
	// ~60% of a continuously-exploring player's encounter rate. The floor binds at
	// 6 fully-staffed societies (900 × 0.65 ÷ 6 ≈ 98); past that, further copies buy
	// only resilience against a party coming home late.
	autoExpeditionMinInterval = 100
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
		Interval:  autoExpeditionIntervalFor(count, fill),
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
	relief := 1 - autoExpeditionWorkerRelief*fill
	interval := int(math.Round(autoExpeditionBaseInterval * relief / float64(count)))
	if interval < autoExpeditionMinInterval {
		interval = autoExpeditionMinInterval
	}
	return interval
}

// pickAutoScoutExpedition chooses which scouting expedition the society sends.
//
// RULE: the CHEAPEST expedition it can currently afford, where "cheapest" is the
// plain sum of its Cost amounts. Ties break on the shorter DurationMin, then on Key
// so the choice is fully deterministic and needs no RNG. The rule is frugal on
// purpose — an automated office should be sending the routine local sweep, not
// spending the treasury on a grand voyage the player never authorised — and the
// crude cross-resource sum is safe here because the scouting pool's costs sit in
// the same order of magnitude as each other at every age.
//
// Returns nil when the age offers no scouting expedition, or when none is
// affordable right now.
func (ge *GameEngine) pickAutoScoutExpedition(ageOrder map[string]int) *ExpeditionDef {
	defs := ge.Military.GetAvailableExpeditionsByCategory(ExpeditionScouting, ge.age, ageOrder)
	if len(defs) == 0 {
		return nil
	}

	affordable := make([]ExpeditionDef, 0, len(defs))
	for _, def := range defs {
		if def.SoldiersNeeded > 0 && int(ge.Resources.Get("soldiers")) < def.SoldiersNeeded {
			continue
		}
		ok := true
		for res, amount := range def.Cost {
			if ge.Resources.Get(res) < amount {
				ok = false
				break
			}
		}
		if ok {
			affordable = append(affordable, def)
		}
	}
	if len(affordable) == 0 {
		return nil
	}

	sort.Slice(affordable, func(i, j int) bool {
		ci, cj := totalExpeditionCost(affordable[i]), totalExpeditionCost(affordable[j])
		if ci != cj {
			return ci < cj
		}
		if affordable[i].DurationMin != affordable[j].DurationMin {
			return affordable[i].DurationMin < affordable[j].DurationMin
		}
		return affordable[i].Key < affordable[j].Key
	})
	best := affordable[0]
	return &best
}

// totalExpeditionCost sums an expedition's resource Cost. See
// pickAutoScoutExpedition for why a naive sum is acceptable as a "cheapness" score.
func totalExpeditionCost(def ExpeditionDef) float64 {
	total := 0.0
	for _, amount := range def.Cost {
		total += amount
	}
	return total
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

	interval := autoExpeditionIntervalFor(count, fill)

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

	// The countdown is spent: a dispatch is DUE. It may still not happen — if the
	// slot is busy or nothing is affordable we hold at zero and retry next tick
	// rather than forfeiting the cycle.
	if ge.Military.ActiveByCategory(ExpeditionScouting) != nil {
		return
	}

	def := ge.pickAutoScoutExpedition(ge.progress.GetAgeOrder())
	if def == nil {
		// Either the age has no scouting expedition at all, or the stores are too
		// thin to outfit one. Warn ONCE per dry spell (state change only) — this is
		// checked every tick and would otherwise bury the log.
		if !ge.autoExpeditionStarved {
			ge.autoExpeditionStarved = true
			ge.addLog("warning", fmt.Sprintf("%s: a survey party stands ready, but there are no supplies to outfit it.",
				ge.autoExpeditionBuildingName()))
		}
		return
	}

	if err := ge.launchExpeditionLocked(def.Key); err != nil {
		// Defensive: the affordability and slot checks above already cover every
		// rejection the launch path can produce. Hold at zero and retry.
		ge.addLog("debug", fmt.Sprintf("auto-dispatch declined: %v", err))
		return
	}

	ge.autoExpeditionStarved = false
	ge.autoExpeditionTicksLeft = interval
	ge.addLog("info", fmt.Sprintf("%s dispatches a survey party: %s.", ge.autoExpeditionBuildingName(), def.Name))
}

// autoExpeditionBuildingName is the society's player-facing name, read from the
// loaded defs so the log lines cannot drift from the building definition.
func (ge *GameEngine) autoExpeditionBuildingName() string {
	if def, ok := ge.Buildings.defs[autoExpeditionBuildingKey]; ok && def.Name != "" {
		return def.Name
	}
	return "The survey office"
}
