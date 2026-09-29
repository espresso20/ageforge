package game

import (
	"fmt"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// TradeManager handles both instant resource exchange and repeating trade routes.
//
// Exchange: converts resource A → B at the configured base rate adjusted by
// supply pressure. Repeated selling of the same pair increases pressure
// (up to +30% rate reduction); pressure decays 2% per tick toward zero.
// More markets/ports reduce pressure accumulation per trade.
//
// Trade routes: run on a configurable interval (TicksPerRun), automatically
// consuming Export resources and adding Import resources each cycle. Allied
// faction bonuses are applied to imported amounts via DiplomacyManager.
// A route is suspended mid-cycle if its required building is demolished.
type TradeManager struct {
	// supplyPressure key is "from:to" (e.g. "wood:gold"); range -1..1.
	// Positive values mean the "from" resource has been oversold, reducing rate.
	supplyPressure map[string]float64
	lastExchange   map[string]int // "from:to" -> tick of last exchange (reserved for future cooldowns)

	activeRoutes map[string]*ActiveRoute // route key -> runtime state

	// Cumulative stats for display in the Trade tab.
	totalExchanged map[string]float64
	totalImported  map[string]float64
	totalExported  map[string]float64

	// Static config tables, built once at construction so Tick (every engine
	// tick) and Snapshot (every UI refresh) don't rebuild them per call.
	// Read-only: never mutate these or the maps inside the defs; Snapshot hands
	// the UI copies of Export/Import so nothing outside the lock aliases them.
	routeList []config.TradeRouteDef
	routeDefs map[string]config.TradeRouteDef

	// age prices the exchange: two construction resources trade at the
	// current age's parity (config.MarketRate). The engine sets it before
	// every exchange; "" falls back to the listed pairs at their base rates.
	age string
}

// SetAge sets the age whose price parity the exchange trades at.
func (tm *TradeManager) SetAge(age string) { tm.age = age }

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

// NewTradeManager creates a new trade manager
func NewTradeManager() *TradeManager {
	routes := config.BaseTradeRoutes()
	routeDefs := make(map[string]config.TradeRouteDef, len(routes))
	for _, def := range routes {
		routeDefs[def.Key] = def
	}
	return &TradeManager{
		supplyPressure: make(map[string]float64),
		lastExchange:   make(map[string]int),
		activeRoutes:   make(map[string]*ActiveRoute),
		totalExchanged: make(map[string]float64),
		totalImported:  make(map[string]float64),
		totalExported:  make(map[string]float64),
		routeList:      routes,
		routeDefs:      routeDefs,
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

// GetExchangeRate returns the current rate for a resource pair, accounting for supply pressure
func (tm *TradeManager) GetExchangeRate(from, to string) float64 {
	base, ok := config.MarketRate(from, to, tm.age)
	if !ok {
		return 0
	}
	pressure := tm.supplyPressure[from+":"+to]
	return base * (1.0 - float64(pressure*0.3))
}

// RateIn is what Exchange would pay now for one from in age: the market rate
// less supply pressure, floored at half the market rate. Read-only.
func (tm *TradeManager) RateIn(from, to, age string) float64 {
	base, ok := config.MarketRate(from, to, age)
	if !ok {
		return 0
	}
	return math.Max(base*(1.0-float64(tm.supplyPressure[from+":"+to]*0.3)), base*0.5)
}

// Pressure is the supply pressure on from → to (0 when the market has fully
// recovered). Read-only.
func (tm *TradeManager) Pressure(from, to string) float64 {
	return tm.supplyPressure[from+":"+to]
}

// Exchange sells amount of give for get at the market rate and returns how
// much get the player received. amount always counts the give side, matching
// the command `trade <give> <get> <amount>`.
func (tm *TradeManager) Exchange(give, get string, amount float64, resources *ResourceManager, buildings *BuildingManager, tick int) (float64, error) {
	from, to := give, get
	key := from + ":" + to
	base, ok := config.MarketRate(from, to, tm.age)
	if !ok {
		return 0, fmt.Errorf("The market does not trade %s for %s in this age. Type trade list to see the rates.", ResourceName(from), ResourceName(to))
	}

	// Require a trade building: a market or anything its lineage becomes.
	// Checking "market" alone meant upgrading your markets (as the log
	// suggests on reaching the Iron Age) shut the exchange until ports.
	traders := buildings.TradeBuildingCount()
	if traders < 1 {
		return 0, fmt.Errorf("You need a Market to trade.")
	}

	// Check sender has enough
	if resources.Get(from) < amount {
		return 0, fmt.Errorf("Not enough %s to give %s (you have %s).", ResourceName(from), textfmt.Number(amount), textfmt.Number(resources.Get(from)))
	}

	// Calculate received amount with supply pressure
	pressure := tm.supplyPressure[key]
	rate := base * (1.0 - float64(pressure*0.3))
	if rate < base*0.5 {
		rate = base * 0.5 // floor at 50% of base
	}
	got := float64(amount * rate)

	// Execute trade
	resources.Remove(from, amount)
	resources.Add(to, got)

	// Update supply pressure (selling more pushes rate down)
	// More markets reduce pressure impact
	marketCount := float64(traders)
	pressureIncrease := 0.1 / (1.0 + float64(marketCount*0.2))
	tm.supplyPressure[key] += pressureIncrease
	if tm.supplyPressure[key] > 1.0 {
		tm.supplyPressure[key] = 1.0
	}

	tm.lastExchange[key] = tick
	tm.totalExchanged[from] += amount
	tm.totalExchanged[to] += got

	return got, nil
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
		return fmt.Errorf("%s needs the %s.", def.Name, AgeName(def.MinAge))
	}

	// Check building requirement
	if buildings.GetCount(def.RequiredBld) < def.MinCount {
		return fmt.Errorf("%s needs %s (you have %s).", def.Name, BuildingCount(def.MinCount, def.RequiredBld), textfmt.Int(buildings.GetCount(def.RequiredBld)))
	}

	// Check not already active
	if _, active := tm.activeRoutes[key]; active {
		return fmt.Errorf("%s is already running.", def.Name)
	}

	tm.activeRoutes[key] = &ActiveRoute{
		Key:       key,
		TicksLeft: def.TicksPerRun,
	}
	return nil
}

// StopRoute deactivates a trade route
func (tm *TradeManager) StopRoute(key string) error {
	if _, active := tm.activeRoutes[key]; !active {
		return fmt.Errorf("%s is not running. Type trade route list to see your routes.", RouteName(key))
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

// Tick processes trade routes and decays supply pressure.
//
// harborBonus is the additive trade-route income multiplier from built harbour
// buildings (e.g. 0.15 = +15%); the engine computes it from the harbor lineage
// and passes it in. Disrupted routes (those importing a resource a hostile civ
// specialises in — see DiplomacyManager.DisruptedResources) are skipped for the
// cycle with a log line.
func (tm *TradeManager) Tick(resources *ResourceManager, buildings *BuildingManager, diplomacy *DiplomacyManager, harborBonus float64) []string {
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
		if buildings.GetCount(def.RequiredBld) < def.MinCount {
			messages = append(messages, fmt.Sprintf("Trade route %s stopped: it needs %s.", def.Name, BuildingCount(def.MinCount, def.RequiredBld)))
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
				route.TicksLeft = def.TicksPerRun
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

			if canAfford {
				// Consume exports
				for res, amount := range def.Export {
					resources.Remove(res, amount)
					tm.totalExported[res] += amount
				}

				// Add imports (with diplomacy ally bonus + harbour bonus)
				for res, amount := range def.Import {
					bonus := harborBonus
					if diplomacy != nil {
						bonus += diplomacy.GetTradeBonus(res)
					}
					actual := float64(amount * (1.0 + bonus))
					resources.Add(res, actual)
					tm.totalImported[res] += actual
				}

				route.CyclesDone++
				route.Starved = false
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
			route.TicksLeft = def.TicksPerRun
		}
	}

	// Decay supply pressure (2% per tick toward 0)
	tm.DecayPressure(1)

	return messages
}

// DecayPressure lets n ticks of supply-pressure decay pass (2% a tick toward
// 0), as Tick does one tick at a time; offline catch-up passes many at once.
// Keys are walked in sorted order so the float results never depend on map
// order.
func (tm *TradeManager) DecayPressure(n int) {
	f := detmath.Pow(0.98, float64(n))
	for _, key := range sortedKeys(tm.supplyPressure) {
		p := tm.supplyPressure[key] * f
		if math.Abs(p) < 0.001 {
			delete(tm.supplyPressure, key)
			continue
		}
		tm.supplyPressure[key] = p
	}
}

// Snapshot returns the trade state for UI consumption. disrupted is the set of
// resources currently blockaded by war/embargo (from DiplomacyManager); routes
// importing one are flagged Disrupted so the overlay can warn the player.
func (tm *TradeManager) Snapshot(age string, ageOrder map[string]int, buildings *BuildingManager, disrupted map[string]bool) TradeState {
	allRoutes := tm.routeDefs

	// Exchange rates: what the market offers in this age (listed pairs plus
	// every pair of the age's construction resources, at parity).
	exchangeRates := make(map[string]ExchangeRateInfo)
	for _, def := range config.MarketPairs(age) {
		key := def.From + ":" + def.To
		pressure := tm.supplyPressure[key]
		base := def.BaseRate
		// The same 50% floor Exchange applies, so the listed rate is the
		// rate the player actually gets.
		currentRate := math.Max(base*(1.0-float64(pressure*0.3)), base*0.5)
		exchangeRates[key] = ExchangeRateInfo{
			From:     def.From,
			To:       def.To,
			Rate:     currentRate,
			BaseRate: base,
			Pressure: pressure,
		}
	}

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
			Import:      copyAmounts(def.Import),
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
		canStart := buildings.GetCount(def.RequiredBld) >= def.MinCount
		availableRoutes = append(availableRoutes, TradeRouteInfo{
			Name:        def.Name,
			Key:         def.Key,
			Export:      copyAmounts(def.Export),
			Import:      copyAmounts(def.Import),
			CanStart:    canStart,
			RequiredBld: def.RequiredBld,
			MinCount:    def.MinCount,
			Description: def.Description,
		})
	}

	// Deep copy stats
	totalExchanged := make(map[string]float64, len(tm.totalExchanged))
	for k, v := range tm.totalExchanged {
		totalExchanged[k] = v
	}
	totalImported := make(map[string]float64, len(tm.totalImported))
	for k, v := range tm.totalImported {
		totalImported[k] = v
	}

	return TradeState{
		ExchangeRates:      exchangeRates,
		ActiveRoutes:       activeRoutes,
		AvailableRoutes:    availableRoutes,
		TotalExchanged:     totalExchanged,
		TotalImported:      totalImported,
		DisruptedResources: disruptedResources,
		TradeBuildings:     buildings.TradeBuildingCount(),
	}
}

// LoadState restores trade state from save
func (tm *TradeManager) LoadState(activeRoutes map[string]ActiveRoute, supplyPressure, totalExchanged, totalImported, totalExported map[string]float64) {
	if activeRoutes != nil {
		for k, v := range activeRoutes {
			route := v // copy
			tm.activeRoutes[k] = &route
		}
	}
	if supplyPressure != nil {
		tm.supplyPressure = supplyPressure
	}
	if totalExchanged != nil {
		tm.totalExchanged = totalExchanged
	}
	if totalImported != nil {
		tm.totalImported = totalImported
	}
	if totalExported != nil {
		tm.totalExported = totalExported
	}
}
