package game

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
)

// TradeManager runs the standing trade routes. A running route is a flow: it
// takes its Export out of the resource rates and puts its Import into them,
// every tick, until it is paused. At the route's gate (MinCount of its
// building line) it sends config.RouteShare of the town's gross production of
// its export; the flow grows with the count of the line. A route that cannot pay its
// export runs at a share of it (Flows). Allied faction bonuses are applied to
// the import via DiplomacyManager. A route whose building is demolished is
// held at 0 until the building is back.
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

	// Static config tables, built once at construction so Tick (every engine
	// tick) and Snapshot (every UI refresh) don't rebuild them per call.
	// Read-only: never mutate these or the maps inside the defs; Snapshot hands
	// the UI copies of Export/Import so nothing outside the lock aliases them.
	routeList []config.TradeRouteDef
	routeDefs map[string]config.TradeRouteDef

	// gross is what the town makes of each resource a tick before any trade
	// flow (set by Flows), which sizes every route.
	gross map[string]float64
}

// ActiveRoute is a route the player has started.
type ActiveRoute struct {
	Key    string
	Paused bool
	// The rest is recomputed by Flows every time the rates are, not saved.
	// Run is the share of its flow the route moves now (0 to 1): 0 when
	// paused, blockaded or short of its building, less than 1 when the
	// export is short. Scale is the flow's multiple of the gate's: 1 at the
	// gate, one more for each doubling of the line (Owned); Research is the research a tick the route brings.
	Run, Scale, Research float64 `json:"-"`
	Owned                int     `json:"-"`
	NoIncome             string  `json:"-"` // the export the town makes none of
	Disrupted            bool    `json:"-"`
	NoBuilding           bool    `json:"-"`
	said                 string
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
	for _, res := range sortedKeys(def.Import) {
		if disrupted[res] {
			return res
		}
	}
	return ""
}

// NewTradeManagerWith creates a new trade manager with set's routes.
func NewTradeManagerWith(set *rules.Set) *TradeManager {
	tm := &TradeManager{
		activeRoutes: make(map[string]*ActiveRoute),
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

// SetTechTerms sets the techs' terms: routeTime multiplies a route's time per
// run and routePay what a run brings in.
func (tm *TradeManager) SetTechTerms(routeTime, routePay float64) {
	tm.routeTime, tm.routePay = routeTime, routePay
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

// RunPay is what one run brings in of an import listed at amount: the techs'
// term plus bonus (the harbor and ally shares), each a share of the listed
// amount. Tick pays from it and the state view reports it.
func (tm *TradeManager) RunPay(amount, bonus float64) float64 {
	return float64(amount * (tm.RoutePay(1) + bonus))
}

// PayShare is one share of a run's pay: a label, a percent of the listed
// import, and where it comes from (ShareTech, ShareBuilding or ShareAlly).
type PayShare struct {
	Label   string
	Percent float64
	Kind    string
}

// The kinds of pay share.
const (
	ShareTech     = "tech"
	ShareBuilding = "building"
	ShareAlly     = "ally"
)

// RouteTerms is what the engine hands Snapshot beyond the routes: the harbor
// bonus and its shares by building (sorted: percent, then label), and the
// allies' bonus and shares for an imported resource.
type RouteTerms struct {
	Harbor       float64
	HarborShares []PayShare
	Ally         func(res string) (float64, []PayShare)
}

// timeFactor is how much faster the techs make a route move: its flow is
// multiplied by it (1 with no tech on route time).
func (tm *TradeManager) timeFactor() float64 {
	if tm.routeTime > 0 && tm.routeTime < 1 {
		return 1 / tm.routeTime
	}
	return 1
}

// volume is the runs a tick def makes at its gate: config.RouteShare of the
// gross production of its scarcest export, in listed amounts. It is 0, with the
// export named, when the town makes none of one.
func (tm *TradeManager) volume(def config.TradeRouteDef) (u float64, none string) {
	u = math.Inf(1)
	for _, res := range sortedKeys(def.Export) {
		g := float64(config.RouteShare * tm.gross[res])
		if g <= 0 {
			return 0, res
		}
		u = min(u, g/def.Export[res])
	}
	if math.IsInf(u, 1) {
		return 0, ""
	}
	return u, ""
}

// flows is what def moves a tick at its gate: what it takes of each export,
// and what it brings of each import with the techs' term, the harbor bonus and
// the allies' shares. payFor adds the shares for the state view.
func (tm *TradeManager) flows(def config.TradeRouteDef, terms RouteTerms) (out, in map[string]float64, shares map[string][]PayShare) {
	u, _ := tm.volume(def)
	k := float64(tm.timeFactor() * u)
	out, in, shares = make(map[string]float64, len(def.Export)), make(map[string]float64, len(def.Import)), make(map[string][]PayShare, len(def.Import))
	for res, amount := range def.Export {
		out[res] = float64(amount * k)
	}
	for res, amount := range def.Import {
		bonus := terms.Harbor
		list := []PayShare{}
		if t := tm.RoutePay(1); t != 1 {
			list = append(list, PayShare{Label: rowid.Line("game.trade.payFor.01", "techs"), Percent: float64((t - 1) * 100), Kind: ShareTech})
		}
		list = append(list, terms.HarborShares...)
		if terms.Ally != nil {
			b, ally := terms.Ally(res)
			bonus += b
			list = append(list, ally...)
		}
		in[res], shares[res] = float64(tm.RunPay(amount, bonus)*k), list
	}
	return out, in, shares
}

// StartRoute activates a trade route
func (tm *TradeManager) StartRoute(key string, buildings *BuildingManager, age string, ageOrder map[string]int) error {
	routes := tm.routeDefs
	def, ok := routes[key]
	if !ok {
		return unknownKeyError(rowid.Line("game.trade.StartRoute.01", "trade route"), key, routes, rowid.Line("game.trade.StartRoute.02", "Type trade route list to see the routes."))
	}

	// Check age requirement
	if ageOrder[def.MinAge] > ageOrder[age] {
		return rowid.Errorf("game.trade.StartRoute.03", "%s needs %s.", def.Name, laterAgeRef(tm.rules, age, def.MinAge))
	}

	// Check building requirement
	if buildings.CountFrom(def.RequiredBld) < def.MinCount {
		return rowid.Errorf("game.trade.StartRoute.04", "%s needs %s (you have %s).", def.Name, buildingCountIn(tm.rules, def.MinCount, def.RequiredBld), textfmt.Int(buildings.CountFrom(def.RequiredBld)))
	}

	// A paused route starts again; a running one is already running.
	if r, active := tm.activeRoutes[key]; active {
		if !r.Paused {
			return rowid.Errorf("game.trade.StartRoute.05", "%s is already running.", def.Name)
		}
		r.Paused = false
		return nil
	}

	tm.activeRoutes[key] = &ActiveRoute{Key: key}
	return nil
}

// StopRoute pauses a trade route: it stays on the list at 0 until started.
func (tm *TradeManager) StopRoute(key string) error {
	if r, active := tm.activeRoutes[key]; !active || r.Paused {
		return rowid.Errorf("game.trade.StopRoute.01", "%s is not running. Type trade route list to see your routes.", tm.rules.Name(rules.KindRoute, key))
	}
	tm.activeRoutes[key].Paused = true
	return nil
}

// ActiveRouteCount returns the number of currently-active trade routes. Used by
// the diplomacy system to drive mercantile-civ opinion drift. Must be called
// under the engine write lock (same as the rest of the manager).
func (tm *TradeManager) ActiveRouteCount() int {
	return len(tm.activeRoutes)
}

// Flows puts the running routes into the resource rates, after production:
// each export is a consumer and each import a producer, with its own line in
// the breakdown. Routes compete for a resource: with D the routes' demand for
// it a tick and S its stock plus what the town makes of it, every route
// exporting it runs at min(1, S/D), both legs, so a shortage slows the routes
// and never skips one. horizon is the ticks the rates will be applied for
// (1 on a live tick; an offline step's ticks), so a long step spends the
// stock once.
func (tm *TradeManager) Flows(rm *ResourceManager, buildings *BuildingManager, diplomacy *DiplomacyManager, terms RouteTerms, horizon float64) {
	horizon = max(horizon, 1)
	if tm.gross == nil {
		tm.gross = map[string]float64{}
	}
	for res, r := range rm.resources {
		tm.gross[res] = max(0, float64(r.Rate-r.Breakdown.FoodDrain))
	}
	if len(tm.activeRoutes) == 0 {
		return
	}
	var disrupted map[string]bool
	if diplomacy != nil {
		disrupted = diplomacy.DisruptedResources()
	}
	type leg struct {
		route   *ActiveRoute
		out, in map[string]float64
	}
	var legs []leg
	demand := map[string]float64{}
	for _, key := range sortedKeys(tm.activeRoutes) {
		route, def := tm.activeRoutes[key], tm.routeDefs[key]
		owned := buildings.CountFrom(def.RequiredBld)
		route.Run, route.Research = 0, 0
		route.Owned, route.Scale = owned, 0
		route.NoBuilding = owned < def.MinCount
		if !route.NoBuilding {
			route.Scale = 1 + math.Log2(float64(owned)/float64(max(def.MinCount, 1)))
		}
		route.Disrupted = tm.routeDisruptedBy(def, disrupted) != ""
		_, route.NoIncome = tm.volume(def)
		if route.Paused || route.NoBuilding || route.Disrupted || route.NoIncome != "" {
			continue
		}
		out, in, _ := tm.flows(def, terms)
		for res, a := range out {
			out[res] = float64(a * route.Scale)
			demand[res] += out[res]
		}
		for res, a := range in {
			in[res] = float64(a * route.Scale)
		}
		legs = append(legs, leg{route, out, in})
	}
	share := make(map[string]float64, len(demand))
	for res, d := range demand {
		if r := rm.resources[res]; r != nil {
			// Only a surplus goes out: the stock and the net rate (what the
			// town makes less what it uses), never below nothing.
			share[res] = min(1, max(0, r.Amount+float64(r.Rate*horizon))/float64(d*horizon))
		}
	}
	for _, l := range legs {
		l.route.Run = 1
		for res := range l.out {
			l.route.Run = min(l.route.Run, share[res])
		}
		for res, a := range l.out {
			if r := rm.resources[res]; r != nil {
				r.Rate -= float64(a * l.route.Run)
				r.Breakdown.TradeFlow -= float64(a * l.route.Run)
			}
		}
		for res, a := range l.in {
			if res == config.ResearchTarget {
				l.route.Research = float64(a * l.route.Run)
			} else if r := rm.resources[res]; r != nil {
				r.Rate += float64(a * l.route.Run)
				r.Breakdown.TradeFlow += float64(a * l.route.Run)
			}
		}
	}
}

// Tick is a live tick's trade: the research the routes bring, a word in the
// log when a route stops running or runs again, and once a minute a warmer
// opinion among the civilizations for each route that runs. The goods
// themselves move in Flows.
func (tm *TradeManager) Tick(tick int, diplomacy *DiplomacyManager, research func(work float64) string) []logNote {
	var messages []logNote
	minute := tick%max(minutesToTicks(1), 1) == 0
	for _, key := range sortedKeys(tm.activeRoutes) {
		route, def := tm.activeRoutes[key], tm.routeDefs[key]
		state := ""
		switch {
		case route.Paused:
			state = "paused"
		case route.NoBuilding:
			state = "building"
			if route.said != state {
				messages = append(messages, note(rowid.Sprintf("game.trade.Tick.01", "Trade route %s is held: it needs %s.", def.Name, buildingCountIn(tm.rules, def.MinCount, def.RequiredBld)), logSubj(key), logTrouble))
			}
		case route.Disrupted:
			state = "war"
			if route.said != state {
				messages = append(messages, note(rowid.Sprintf("game.trade.Tick.02", "Trade route %s is held: a war or embargo is blocking %s shipments.", def.Name, ResourceName(tm.routeDisruptedBy(def, diplomacy.DisruptedResources()))), logSubj(key), logTrouble))
			}
		case route.Run < 1:
			state = "short"
			if route.said != state {
				messages = append(messages, note(rowid.Sprintf("game.trade.Tick.04", "Trade route %s is slowed: the town is short of what it exports.", def.Name), logSubj(key), logTrouble))
			}
		}
		if state == "" && route.said != "" && route.said != "paused" {
			messages = append(messages, note(rowid.Sprintf("game.trade.Tick.05", "Trade route %s is running at full again.", def.Name), logSubj(key), logAs("info")))
		}
		route.said = state
		if route.Run <= 0 {
			continue
		}
		if route.Research > 0 && research != nil {
			research(route.Research)
		}
		if minute && diplomacy != nil {
			diplomacy.RecordTrade()
		}
	}
	return messages
}

// Snapshot returns the trade state for UI consumption. disrupted is the set of
// resources currently blockaded by war/embargo (from DiplomacyManager); routes
// importing one are flagged Disrupted so the overlay can warn the player.
func (tm *TradeManager) Snapshot(age string, ageOrder map[string]int, buildings *BuildingManager, disrupted map[string]bool, terms RouteTerms) TradeState {
	allRoutes := tm.routeDefs

	// Active routes
	var activeRoutes []ActiveRouteInfo
	if len(tm.activeRoutes) > 0 {
		activeRoutes = make([]ActiveRouteInfo, 0, len(tm.activeRoutes))
	}
	var disruptedResources []string
	seenDisrupt := make(map[string]bool)
	for _, key := range sortedKeys(tm.activeRoutes) {
		route := tm.activeRoutes[key]
		def := allRoutes[key]
		blockedBy := tm.routeDisruptedBy(def, disrupted)
		out, in, shares := tm.flows(def, terms)
		activeRoutes = append(activeRoutes, ActiveRouteInfo{
			Name:        def.Name,
			Key:         key,
			Export:      out,
			Import:      in,
			Disrupted:   blockedBy != "",
			DisruptedBy: blockedBy,
			Paused:      route.Paused,
			NoBuilding:  route.NoBuilding,
			NoIncome:    route.NoIncome,
			Run:         route.Run,
			Scale:       route.Scale,
			Owned:       route.Owned,
			Shares:      shares,
		})
		if blockedBy != "" && !seenDisrupt[blockedBy] {
			seenDisrupt[blockedBy] = true
			disruptedResources = append(disruptedResources, blockedBy)
		}
	}
	sort.Strings(disruptedResources)

	// Available routes
	var availableRoutes []TradeRouteInfo
	open := 0
	for _, def := range tm.routeList {
		if ageOrder[def.MinAge] <= ageOrder[age] {
			if _, active := tm.activeRoutes[def.Key]; !active {
				open++
			}
		}
	}
	if open > 0 {
		availableRoutes = make([]TradeRouteInfo, 0, open)
	}
	for _, def := range tm.routeList {
		if ageOrder[def.MinAge] > ageOrder[age] {
			continue
		}
		if _, active := tm.activeRoutes[def.Key]; active {
			continue
		}
		canStart := buildings.CountFrom(def.RequiredBld) >= def.MinCount
		out, in, shares := tm.flows(def, terms)
		availableRoutes = append(availableRoutes, TradeRouteInfo{
			Name:        def.Name,
			Key:         def.Key,
			Export:      out,
			Import:      in,
			CanStart:    canStart,
			RequiredBld: def.RequiredBld,
			MinCount:    def.MinCount,
			Description: def.Description,
			Shares:      shares,
		})
	}

	// Deep copy stats
	return TradeState{
		ActiveRoutes:       activeRoutes,
		AvailableRoutes:    availableRoutes,
		DisruptedResources: disruptedResources,
		Speed:              tm.timeFactor(),
	}
}

// LoadState restores trade state from save.
func (tm *TradeManager) LoadState(s TradeSave) {
	for key, route := range s.ActiveRoutes {
		route := route // copy
		tm.activeRoutes[key] = &route
	}
}
