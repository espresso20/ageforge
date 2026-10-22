package game

import (
	"fmt"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// TradeManager runs the repeating trade routes. (The market no longer swaps
// one resource for another: a town makes what it needs, and the save keeps
// the old swap's fields only so a signed save still verifies.)
//
// Trade routes: run on a configurable interval (TicksPerRun), automatically
// consuming Export resources and adding Import resources each cycle. Allied
// faction bonuses are applied to imported amounts via DiplomacyManager.
// A route is suspended mid-cycle if its required building is demolished.
type TradeManager struct {
	rules *rules.Set

	activeRoutes map[string]*ActiveRoute // route key -> runtime state

	// routeTime is the techs' term on a route's time per run
	// (config.MechanicRouteTicks). The engine sets it (SetTechTerms); 0 reads
	// as 1, so a manager no engine drives runs routes at the listed numbers.
	routeTime float64
	// routePay is the techs' term on what a route's run brings in, set the
	// same way; 0 reads as 1.
	routePay float64

	// Cumulative route stats for display in the Trade panel.
	totalImported map[string]float64
	totalExported map[string]float64

	// Static config tables, built once at construction so Tick (every engine
	// tick) and Snapshot (every UI refresh) don't rebuild them per call.
	// Read-only: never mutate these or the maps inside the defs; Snapshot hands
	// the UI copies of Export/Import so nothing outside the lock aliases them.
	routeList []config.TradeRouteDef
	routeDefs map[string]config.TradeRouteDef
}

// ActiveRoute represents a running trade route
type ActiveRoute struct {
	Key        string
	TicksLeft  int
	CyclesDone int
	// Disrupted is transient runtime state (recomputed every tick from diplomacy
	// war/embargo) — not persisted. True when the route's imports are blockaded.
	Disrupted bool `json:"-"`
	// Starved is transient: true after a run was skipped because the exports
	// were short, so the warning logs once per shortage, not every run.
	Starved bool `json:"-"`
}

// routeDisruptedBy returns the first imported resource of def that appears in the
// disrupted set (resources blockaded by war/embargo), or "" if the route is
// clear. A non-empty return means the route is currently disrupted.
func (tm *TradeManager) routeDisruptedBy(def config.TradeRouteDef, disrupted map[string]bool) string {
	if len(disrupted) == 0 {
		return ""
	}
	// Deterministic: check imports in sorted key order so the reported resource
	// is stable across runs (Go map iteration is randomised).
	keys := make([]string, 0, len(def.Import))
	for res := range def.Import {
		keys = append(keys, res)
	}
	sort.Strings(keys)
	for _, res := range keys {
		if disrupted[res] {
			return res
		}
	}
	return ""
}

// NewTradeManager creates a new trade manager on the core ruleset.
func NewTradeManager() *TradeManager { return NewTradeManagerWith(rules.Core()) }

// NewTradeManagerWith creates a new trade manager with set's routes.
func NewTradeManagerWith(set *rules.Set) *TradeManager {
	tm := &TradeManager{
		activeRoutes:  make(map[string]*ActiveRoute),
		totalImported: make(map[string]float64),
		totalExported: make(map[string]float64),
	}
	tm.Rebind(set)
	return tm
}

// Rebind moves the manager onto set: it takes set's trade routes. Running
// routes and totals stay.
func (tm *TradeManager) Rebind(set *rules.Set) {
	tm.rules = set
	tm.routeList = set.TradeRoutes()
	tm.routeDefs = make(map[string]config.TradeRouteDef, len(tm.routeList))
	for _, def := range tm.routeList {
		tm.routeDefs[def.Key] = def
	}
}

// copyAmounts returns an independent copy of a resource-amount map so UI
// snapshots never alias the manager's held config defs.
func copyAmounts(m map[string]float64) map[string]float64 {
	if m == nil {
		return nil
	}
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// SetTechTerms sets the techs' terms: routeTime multiplies a route's time per
// run and routePay what a run brings in.
func (tm *TradeManager) SetTechTerms(routeTime, routePay float64) {
	tm.routeTime, tm.routePay = routeTime, routePay
}

// routeImports is a copy of what one run of def brings in, with the techs'
// term on trade route income: what the panels show.
func (tm *TradeManager) routeImports(def config.TradeRouteDef) map[string]float64 {
	out := copyAmounts(def.Import)
	for res, amount := range out {
		out[res] = tm.RoutePay(amount)
	}
	return out
}

// RoutePay is what one run of a route brings in of an import listed at
// amount, with the techs' term on trade route income: the panels and the
// engine read a route's imports through it. The ally and harbor bonuses
// are added to it as further shares of the listed amount.
func (tm *TradeManager) RoutePay(amount float64) float64 {
	if tm.routePay > 0 && tm.routePay != 1 {
		return float64(amount * tm.routePay)
	}
	return amount
}

// RunTicks is how long one run of a route listed at ticks takes with the
// techs' cut of route time: rounded down, one tick at least.
func (tm *TradeManager) RunTicks(ticks int) int {
	return techTimeTicks(ticks, tm.routeTime)
}

// StartRoute activates a trade route
func (tm *TradeManager) StartRoute(key string, buildings *BuildingManager, age string, ageOrder map[string]int) error {
	routes := tm.routeDefs
	def, ok := routes[key]
	if !ok {
		return unknownKeyError("trade route", key, routes, "Type trade route list to see the routes.")
	}

	// Check age requirement
	if ageOrder[def.MinAge] > ageOrder[age] {
		return fmt.Errorf("%s needs %s.", def.Name, laterAgeRef(tm.rules, age, def.MinAge))
	}

	// Check building requirement
	if buildings.CountFrom(def.RequiredBld) < def.MinCount {
		return fmt.Errorf("%s needs %s (you have %s).", def.Name, buildingCountIn(tm.rules, def.MinCount, def.RequiredBld), textfmt.Int(buildings.CountFrom(def.RequiredBld)))
	}

	// Check not already active
	if _, active := tm.activeRoutes[key]; active {
		return fmt.Errorf("%s is already running.", def.Name)
	}

	tm.activeRoutes[key] = &ActiveRoute{
		Key:       key,
		TicksLeft: tm.RunTicks(def.TicksPerRun),
	}
	return nil
}

// StopRoute deactivates a trade route
func (tm *TradeManager) StopRoute(key string) error {
	if _, active := tm.activeRoutes[key]; !active {
		return fmt.Errorf("%s is not running. Type trade route list to see your routes.", tm.rules.Name(rules.KindRoute, key))
	}
	delete(tm.activeRoutes, key)
	return nil
}

// ActiveRouteCount returns the number of currently-active trade routes. Used by
// the diplomacy system to drive mercantile-civ opinion drift. Must be called
// under the engine write lock (same as the rest of the manager).
func (tm *TradeManager) ActiveRouteCount() int {
	return len(tm.activeRoutes)
}

// Tick processes trade routes.
//
// harborBonus is the additive trade-route income multiplier from built harbour
// buildings (e.g. 0.15 = +15%); the engine computes it from the harbor lineage
// and passes it in. Disrupted routes (those importing a resource a hostile civ
// specialises in — see DiplomacyManager.DisruptedResources) are skipped for the
// cycle with a log line.
func (tm *TradeManager) Tick(resources *ResourceManager, buildings *BuildingManager, diplomacy *DiplomacyManager, harborBonus float64, research func(work float64) string) []string {
	var messages []string

	routes := tm.routeDefs

	// Resources currently blockaded by war/embargo. A route whose imports touch
	// any of these is disrupted (income blocked) until the conflict ends.
	var disrupted map[string]bool
	if diplomacy != nil {
		disrupted = diplomacy.DisruptedResources()
	}

	// Process active trade routes in key order: routes can compete for the
	// same export, so which one runs first decides which one is starved.
	for _, key := range sortedKeys(tm.activeRoutes) {
		route := tm.activeRoutes[key]
		def, ok := routes[key]
		if !ok {
			continue
		}

		// Check building still meets requirements
		if buildings.CountFrom(def.RequiredBld) < def.MinCount {
			messages = append(messages, fmt.Sprintf("Trade route %s stopped: it needs %s.", def.Name, buildingCountIn(tm.rules, def.MinCount, def.RequiredBld)))
			delete(tm.activeRoutes, key)
			continue
		}

		// Disruption: if any imported resource is blockaded, the route is dead
		// in the water this cycle — no export consumed, no income, timer still
		// ticks so it resumes automatically once peace returns.
		blockedRes := tm.routeDisruptedBy(def, disrupted)
		route.Disrupted = blockedRes != ""
		if route.Disrupted {
			route.TicksLeft--
			if route.TicksLeft <= 0 {
				messages = append(messages, fmt.Sprintf("Trade route %s disrupted: a war or embargo is blocking %s shipments.", def.Name, ResourceName(blockedRes)))
				route.TicksLeft = tm.RunTicks(def.TicksPerRun)
			}
			continue
		}

		route.TicksLeft--
		if route.TicksLeft <= 0 {
			// Check if we can afford the exports
			canAfford := true
			for res, amount := range def.Export {
				if resources.Get(res) < amount {
					canAfford = false
					break
				}
			}

			// What a run brings in, with the diplomacy ally bonus and the
			// harbor bonus. The techs' term and those bonuses are each a
			// share of the listed import, so each delivers what it says
			// whatever the others are.
			imports := make(map[string]float64, len(def.Import))
			for res, amount := range def.Import {
				bonus := harborBonus
				if diplomacy != nil {
					bonus += diplomacy.GetTradeBonus(res)
				}
				imports[res] = float64(amount * (tm.RoutePay(1) + bonus))
			}

			// A full store is a wall: a run moves only as much as the tightest
			// import store has room for, at the route's rate, and takes only
			// that share of its export. With no room at all the run does not
			// happen, as when the export stock is short.
			scale := float64(1)
			if canAfford {
				for res, actual := range imports {
					if !(actual > 0) || res == config.ResearchTarget {
						continue // research is not stored: it has no store to fill
					}
					room := resources.GetStorage(res) - resources.Get(res)
					if !(room > 0) {
						scale = 0
						break
					}
					if share := room / actual; share < scale {
						scale = share
					}
				}
				if !(scale > 0) {
					canAfford = false
				}
			}

			if canAfford {
				// Consume exports
				for res, amount := range def.Export {
					resources.Remove(res, amount*scale)
					tm.totalExported[res] += float64(amount * scale)
				}

				// Add imports
				promised, took := map[string]float64{}, map[string]float64{}
				for res, actual := range imports {
					if res == config.ResearchTarget {
						if research != nil {
							if said := research(actual * scale); said != "" {
								messages = append(messages, "  → "+def.Name+" gave "+said+".")
							}
						}
						continue
					}
					before := resources.Get(res)
					resources.Add(res, actual*scale)
					promised[res], took[res] = actual, resources.Get(res)-before
					tm.totalImported[res] += took[res]
				}
				// A cycle a full store cut short says what fit.
				if line := clippedLine(promised, took); line != "" {
					messages = append(messages, line)
				}

				route.CyclesDone++
				route.Starved = false
				// A completed cycle warms every civ you have met and are not at war with.
				if diplomacy != nil {
					diplomacy.RecordTrade()
				}
			} else if !route.Starved {
				route.Starved = true
				for _, res := range sortedKeys(def.Export) {
					if resources.Get(res) < def.Export[res] {
						messages = append(messages, fmt.Sprintf("%s skipped a run: it needs %s to export.", def.Name, Amount(def.Export[res], res)))
						break
					}
				}
			}

			// Reset cycle
			route.TicksLeft = tm.RunTicks(def.TicksPerRun)
		}
	}

	return messages
}

// Snapshot returns the trade state for UI consumption. disrupted is the set of
// resources currently blockaded by war/embargo (from DiplomacyManager); routes
// importing one are flagged Disrupted so the overlay can warn the player.
func (tm *TradeManager) Snapshot(age string, ageOrder map[string]int, buildings *BuildingManager, disrupted map[string]bool) TradeState {
	allRoutes := tm.routeDefs

	// Active routes
	var activeRoutes []ActiveRouteInfo
	var disruptedResources []string
	seenDisrupt := make(map[string]bool)
	for key, route := range tm.activeRoutes {
		def := allRoutes[key]
		blockedBy := tm.routeDisruptedBy(def, disrupted)
		activeRoutes = append(activeRoutes, ActiveRouteInfo{
			Name:        def.Name,
			Key:         key,
			TicksLeft:   route.TicksLeft,
			CyclesDone:  route.CyclesDone,
			Export:      copyAmounts(def.Export),
			Import:      tm.routeImports(def),
			Disrupted:   blockedBy != "",
			DisruptedBy: blockedBy,
		})
		if blockedBy != "" && !seenDisrupt[blockedBy] {
			seenDisrupt[blockedBy] = true
			disruptedResources = append(disruptedResources, blockedBy)
		}
	}
	sort.Strings(disruptedResources)

	// Available routes
	var availableRoutes []TradeRouteInfo
	for _, def := range tm.routeList {
		if ageOrder[def.MinAge] > ageOrder[age] {
			continue
		}
		if _, active := tm.activeRoutes[def.Key]; active {
			continue
		}
		canStart := buildings.CountFrom(def.RequiredBld) >= def.MinCount
		availableRoutes = append(availableRoutes, TradeRouteInfo{
			Name:        def.Name,
			Key:         def.Key,
			Export:      copyAmounts(def.Export),
			Import:      tm.routeImports(def),
			CanStart:    canStart,
			RequiredBld: def.RequiredBld,
			MinCount:    def.MinCount,
			Description: def.Description,
		})
	}

	// Deep copy stats
	totalImported := make(map[string]float64, len(tm.totalImported))
	for k, v := range tm.totalImported {
		totalImported[k] = v
	}

	return TradeState{
		ActiveRoutes:       activeRoutes,
		AvailableRoutes:    availableRoutes,
		TotalImported:      totalImported,
		DisruptedResources: disruptedResources,
	}
}

// LoadState restores trade state from save. The old swap's fields in the save
// (supply pressure, exchange totals, sold and bought) are read past: nothing
// uses them any more, and the next save writes them empty.
func (tm *TradeManager) LoadState(s TradeSave) {
	activeRoutes, totalImported, totalExported := s.ActiveRoutes, s.TotalImported, s.TotalExported
	if activeRoutes != nil {
		for k, v := range activeRoutes {
			route := v // copy
			tm.activeRoutes[k] = &route
		}
	}
	if totalImported != nil {
		tm.totalImported = totalImported
	}
	if totalExported != nil {
		tm.totalExported = totalExported
	}
}
