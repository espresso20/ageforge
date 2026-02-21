package game

import (
	"fmt"

	"github.com/user/ageforge/config"
)

// TradeManager handles resource exchange and trade routes
type TradeManager struct {
	// Exchange system
	supplyPressure map[string]float64 // "from:to" -> pressure (-1 to 1)
	lastExchange   map[string]int     // "from:to" -> last tick exchanged

	// Trade routes
	activeRoutes map[string]*ActiveRoute // route key -> state

	// Stats
	totalExchanged map[string]float64
	totalImported  map[string]float64
	totalExported  map[string]float64
}

// ActiveRoute represents a running trade route
type ActiveRoute struct {
	Key        string
	TicksLeft  int
	CyclesDone int
}

// NewTradeManager creates a new trade manager
func NewTradeManager() *TradeManager {
	return &TradeManager{
		supplyPressure: make(map[string]float64),
		lastExchange:   make(map[string]int),
		activeRoutes:   make(map[string]*ActiveRoute),
		totalExchanged: make(map[string]float64),
		totalImported:  make(map[string]float64),
		totalExported:  make(map[string]float64),
	}
}

// GetExchangeRate returns the current rate for a resource pair, accounting for supply pressure
func (tm *TradeManager) GetExchangeRate(from, to string) float64 {
	rates := config.ExchangeRateByKey()
	key := from + ":" + to
	def, ok := rates[key]
	if !ok {
		return 0
	}
	pressure := tm.supplyPressure[key]
	return def.BaseRate * (1.0 - pressure*0.3)
}

// Exchange performs an instant resource exchange
func (tm *TradeManager) Exchange(from, to string, amount float64, resources *ResourceManager, buildings *BuildingManager, tick int) (float64, error) {
	rates := config.ExchangeRateByKey()
	key := from + ":" + to
	def, ok := rates[key]
	if !ok {
		return 0, fmt.Errorf("no exchange rate for %s → %s", from, to)
	}

	// Require at least 1 market or port
	if buildings.GetCount("market") < 1 && buildings.GetCount("port") < 1 {
		return 0, fmt.Errorf("need a market or port to trade")
	}

	// Check sender has enough
	if resources.Get(from) < amount {
		return 0, fmt.Errorf("not enough %s (have: %.0f, need: %.0f)", from, resources.Get(from), amount)
	}

	// Calculate received amount with supply pressure
	pressure := tm.supplyPressure[key]
	rate := def.BaseRate * (1.0 - pressure*0.3)
	if rate < def.BaseRate*0.5 {
		rate = def.BaseRate * 0.5 // floor at 50% of base
	}
	got := amount * rate

	// Execute trade
	resources.Remove(from, amount)
	resources.Add(to, got)

	// Update supply pressure (selling more pushes rate down)
	// More markets reduce pressure impact
	marketCount := float64(buildings.GetCount("market") + buildings.GetCount("port"))
	pressureIncrease := 0.1 / (1.0 + marketCount*0.2)
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
	routes := config.TradeRouteByKey()
	def, ok := routes[key]
	if !ok {
		return fmt.Errorf("unknown trade route: %s", key)
	}

	// Check age requirement
	if ageOrder[def.MinAge] > ageOrder[age] {
		return fmt.Errorf("%s requires %s", def.Name, def.MinAge)
	}

	// Check building requirement
	if buildings.GetCount(def.RequiredBld) < def.MinCount {
		return fmt.Errorf("%s requires %d %s(s) (have: %d)", def.Name, def.MinCount, def.RequiredBld, buildings.GetCount(def.RequiredBld))
	}

	// Check not already active
	if _, active := tm.activeRoutes[key]; active {
		return fmt.Errorf("%s is already active", def.Name)
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
		return fmt.Errorf("trade route %s is not active", key)
	}
	delete(tm.activeRoutes, key)
	return nil
}

// Tick processes trade routes and decays supply pressure
func (tm *TradeManager) Tick(resources *ResourceManager, buildings *BuildingManager, diplomacy *DiplomacyManager) []string {
	var messages []string

	routes := config.TradeRouteByKey()

	// Process active trade routes
	for key, route := range tm.activeRoutes {
		def, ok := routes[key]
		if !ok {
			continue
		}

		// Check building still meets requirements
		if buildings.GetCount(def.RequiredBld) < def.MinCount {
			messages = append(messages, fmt.Sprintf("Trade route %s stopped: not enough %s", def.Name, def.RequiredBld))
			delete(tm.activeRoutes, key)
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

				// Add imports (with diplomacy bonus)
				for res, amount := range def.Import {
					bonus := 0.0
					if diplomacy != nil {
						bonus = diplomacy.GetTradeBonus(res)
					}
					actual := amount * (1.0 + bonus)
					resources.Add(res, actual)
					tm.totalImported[res] += actual
				}

				route.CyclesDone++
			}

			// Reset cycle
			route.TicksLeft = def.TicksPerRun
		}
	}

	// Decay supply pressure (2% per tick toward 0)
	for key, pressure := range tm.supplyPressure {
		if pressure > 0 {
			tm.supplyPressure[key] = pressure * 0.98
			if tm.supplyPressure[key] < 0.001 {
				delete(tm.supplyPressure, key)
			}
		} else if pressure < 0 {
			tm.supplyPressure[key] = pressure * 0.98
			if tm.supplyPressure[key] > -0.001 {
				delete(tm.supplyPressure, key)
			}
		}
	}

	return messages
}

// Snapshot returns the trade state for UI consumption
func (tm *TradeManager) Snapshot(age string, ageOrder map[string]int, buildings *BuildingManager) TradeState {
	rates := config.ExchangeRateByKey()
	allRoutes := config.TradeRouteByKey()

	// Exchange rates
	exchangeRates := make(map[string]ExchangeRateInfo)
	for key, def := range rates {
		if ageOrder[def.MinAge] > ageOrder[age] {
			continue
		}
		pressure := tm.supplyPressure[key]
		currentRate := def.BaseRate * (1.0 - pressure*0.3)
		exchangeRates[key] = ExchangeRateInfo{
			From:     def.From,
			To:       def.To,
			Rate:     currentRate,
			BaseRate: def.BaseRate,
			Pressure: pressure,
		}
	}

	// Active routes
	var activeRoutes []ActiveRouteInfo
	for key, route := range tm.activeRoutes {
		def := allRoutes[key]
		activeRoutes = append(activeRoutes, ActiveRouteInfo{
			Name:       def.Name,
			Key:        key,
			TicksLeft:  route.TicksLeft,
			CyclesDone: route.CyclesDone,
			Export:     def.Export,
			Import:     def.Import,
		})
	}

	// Available routes
	var availableRoutes []TradeRouteInfo
	for _, def := range config.BaseTradeRoutes() {
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
			Export:      def.Export,
			Import:      def.Import,
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
		ExchangeRates:   exchangeRates,
		ActiveRoutes:    activeRoutes,
		AvailableRoutes: availableRoutes,
		TotalExchanged:  totalExchanged,
		TotalImported:   totalImported,
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
