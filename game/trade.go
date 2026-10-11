package game

import (
	"fmt"
	"maps"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
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
	rules *rules.Set
	// supplyPressure key is "from:to" (e.g. "wood:gold"); range -1..1.
	// Positive values mean the "from" resource has been oversold, reducing rate.
	supplyPressure map[string]float64
	lastExchange   map[string]int // "from:to" -> tick of last exchange (reserved for future cooldowns)

	activeRoutes map[string]*ActiveRoute // route key -> runtime state

	// feeScale and routeTime are the techs' terms on the market's fee and
	// on a route's time per run (config.MechanicMarketFee as a multiplier
	// on every market rate, config.MechanicRouteTicks). The engine sets
	// them (SetTechTerms); 0 reads as 1, so a manager no engine drives
	// trades at the listed numbers.
	feeScale  float64
	routeTime float64
	// routePay is the techs' term on what a route's run brings in, set the
	// same way; 0 reads as 1.
	routePay float64

	// Cumulative stats for display in the Trade panel. totalExchanged sums
	// both sides of every market trade (kept for old saves); totalSold and
	// totalBought split them.
	totalExchanged map[string]float64
	totalSold      map[string]float64
	totalBought    map[string]float64
	totalImported  map[string]float64
	totalExported  map[string]float64

	// Static config tables, built once at construction so Tick (every engine
	// tick) and Snapshot (every UI refresh) don't rebuild them per call.
	// Read-only: never mutate these or the maps inside the defs; Snapshot hands
	// the UI copies of Export/Import so nothing outside the lock aliases them.
	routeList []config.TradeRouteDef
	routeDefs map[string]config.TradeRouteDef

	// age prices the exchange: two construction resources trade at the
	// current age's parity (rules.Set.MarketRate). The engine sets it before
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

// NewTradeManager creates a new trade manager on the core ruleset.
func NewTradeManager() *TradeManager { return NewTradeManagerWith(rules.Core()) }

// NewTradeManagerWith creates a new trade manager with set's routes and
// market.
func NewTradeManagerWith(set *rules.Set) *TradeManager {
	tm := &TradeManager{
		supplyPressure: make(map[string]float64),
		lastExchange:   make(map[string]int),
		activeRoutes:   make(map[string]*ActiveRoute),
		totalExchanged: make(map[string]float64),
		totalSold:      make(map[string]float64),
		totalBought:    make(map[string]float64),
		totalImported:  make(map[string]float64),
		totalExported:  make(map[string]float64),
	}
	tm.Rebind(set)
	return tm
}

// Rebind moves the manager onto set: it takes set's trade routes and prices
// the exchange from set. Running routes, pressure and totals stay.
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

// SetTechTerms sets the techs' terms: feeScale multiplies every market rate
// (a lower fee pays more), routeTime multiplies a route's time per run and
// routePay what a run brings in.
func (tm *TradeManager) SetTechTerms(feeScale, routeTime, routePay float64) {
	tm.feeScale, tm.routeTime, tm.routePay = feeScale, routeTime, routePay
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

// marketRate is the ruleset's market rate for one from in age with the
// techs' cut of the fee: what the market pays before supply pressure.
func (tm *TradeManager) marketRate(from, to, age string) (float64, bool) {
	base, ok := tm.rules.MarketRate(from, to, age)
	if ok && tm.feeScale > 0 && tm.feeScale != 1 {
		base = float64(base * tm.feeScale)
	}
	return base, ok
}

// RunTicks is how long one run of a route listed at ticks takes with the
// techs' cut of route time: rounded down, one tick at least.
func (tm *TradeManager) RunTicks(ticks int) int {
	return techTimeTicks(ticks, tm.routeTime)
}

// GetExchangeRate returns the current rate for a resource pair, accounting for supply pressure
func (tm *TradeManager) GetExchangeRate(from, to string) float64 {
	base, ok := tm.marketRate(from, to, tm.age)
	if !ok {
		return 0
	}
	pressure := tm.supplyPressure[from+":"+to]
	return base * (1.0 - float64(pressure*0.3))
}

// RateIn is what Exchange would pay now for one from in age: the market rate
// less supply pressure, floored at half the market rate. Read-only.
func (tm *TradeManager) RateIn(from, to, age string) float64 {
	base, ok := tm.marketRate(from, to, age)
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
	_, got, err := tm.exchange(give, get, amount, resources, buildings, tick)
	return got, err
}

// exchange is Exchange for a caller that also needs what was actually given:
// a store with less room than the sale pays for takes a smaller sale, so the
// player is charged only for what it can hold. A store with no room refuses.
func (tm *TradeManager) exchange(give, get string, amount float64, resources *ResourceManager, buildings *BuildingManager, tick int) (gave, got float64, err error) {
	from, to := give, get
	key := from + ":" + to
	base, ok := tm.marketRate(from, to, tm.age)
	if !ok {
		return 0, 0, fmt.Errorf("The market does not trade %s for %s in this age. Type trade list to see the rates.", ResourceName(from), ResourceName(to))
	}

	// Require a trade building: a market or anything its lineage becomes.
	// Checking "market" alone meant upgrading your markets (as the log
	// suggests on reaching the Iron Age) shut the exchange until ports.
	traders := buildings.TradeBuildingCount()
	if traders < 1 {
		return 0, 0, fmt.Errorf("You need a Market to trade.")
	}

	// Check sender has enough
	if have := resources.Get(from); have < amount {
		// Amounts print to three figures: when the two read the same, say
		// how far short instead ("to give 247 (you have 247)" says nothing).
		if textfmt.Number(amount) == textfmt.Number(have) {
			return 0, 0, fmt.Errorf("Not enough %s to give %s: you are %s short.", ResourceName(from), textfmt.Number(amount), textfmt.Number(amount-have))
		}
		return 0, 0, fmt.Errorf("Not enough %s to give %s (you have %s).", ResourceName(from), textfmt.Number(amount), textfmt.Number(have))
	}

	// Calculate received amount with supply pressure
	pressure := tm.supplyPressure[key]
	rate := base * (1.0 - float64(pressure*0.3))
	if rate < base*0.5 {
		rate = base * 0.5 // floor at 50% of base
	}
	got = float64(amount * rate)

	// A full store is a wall: sell only what it can hold, and charge for
	// only that.
	room := resources.GetStorage(to) - resources.Get(to)
	if !(room > 0) {
		return 0, 0, fmt.Errorf("%s storage is full.", textfmt.Capitalize(ResourceName(to)))
	}
	if got > room {
		amount, got = room/rate, room
	}

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
	tm.totalSold[from] += amount
	tm.totalBought[to] += got

	return amount, got, nil
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
	if buildings.GetCount(def.RequiredBld) < def.MinCount {
		return fmt.Errorf("%s needs %s (you have %s).", def.Name, buildingCountIn(tm.rules, def.MinCount, def.RequiredBld), textfmt.Int(buildings.GetCount(def.RequiredBld)))
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

			if canAfford {
				// Consume exports
				for res, amount := range def.Export {
					resources.Remove(res, amount)
					tm.totalExported[res] += amount
				}

				// Add imports (with diplomacy ally bonus + harbour bonus)
				promised, took := map[string]float64{}, map[string]float64{}
				for res, amount := range def.Import {
					bonus := harborBonus
					if diplomacy != nil {
						bonus += diplomacy.GetTradeBonus(res)
					}
					// The techs' term and the ally and harbor bonuses are each
					// a share of the listed import, so each delivers what it
					// says whatever the others are.
					actual := float64(amount * (tm.RoutePay(1) + bonus))
					before := resources.Get(res)
					resources.Add(res, actual)
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
	for _, def := range tm.rules.MarketPairs(age) {
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
		canStart := buildings.GetCount(def.RequiredBld) >= def.MinCount
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
	totalExchanged := make(map[string]float64, len(tm.totalExchanged))
	for k, v := range tm.totalExchanged {
		totalExchanged[k] = v
	}
	totalImported := make(map[string]float64, len(tm.totalImported))
	for k, v := range tm.totalImported {
		totalImported[k] = v
	}

	return TradeState{
		TotalSold:          maps.Clone(tm.totalSold),
		TotalBought:        maps.Clone(tm.totalBought),
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
func (tm *TradeManager) LoadState(s TradeSave) {
	activeRoutes, supplyPressure, totalExchanged, totalImported, totalExported :=
		s.ActiveRoutes, s.SupplyPressure, s.TotalExchanged, s.TotalImported, s.TotalExported
	if s.TotalSold != nil {
		tm.totalSold = s.TotalSold
	}
	if s.TotalBought != nil {
		tm.totalBought = s.TotalBought
	}
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
