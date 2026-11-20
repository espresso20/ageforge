package game

import (
	"maps"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
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
	// order is the sorted keys of
	// activeRoutes (routes are only added, so a length check keeps it true).
	order []string
	// The throttle's scratch, kept between calls.
	legs          []routeLeg
	demand, share map[string]float64
	// allyMemo is the allies' bonus per good within one Flows call.
	allyMemo map[string]float64
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
	// sized is the route's scaled flows with the inputs they were worked
	// from: they are redone only when one of those changes.
	sized sizedFlow
}

// routeLeg is a running route and its flows, for the throttle.
type routeLeg struct {
	route   *ActiveRoute
	out, in map[string]float64
}

// flowKey is everything a route's flow is sized from.
type flowKey struct {
	age                       string
	speed, pay, harbor, scale float64
	gross                     [4]float64 // exports, then imports
	ally                      [2]float64
}

// sizedFlow is a route's flows at its scale, read-only once stored.
type sizedFlow struct {
	key     flowKey
	ok      bool
	out, in map[string]float64
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
	for _, res := range def.Import {
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
	// Lean leaves the shares out (Flows needs only the sums): no lists are
	// built, and Ally is read as the bonus alone.
	Lean bool
	// AllyBonus is Ally's bonus without the shares, for Lean terms.
	AllyBonus func(res string) float64
}

// timeFactor is how much faster the techs make a route move: its flow is
// multiplied by it (1 with no tech on route time).
func (tm *TradeManager) timeFactor() float64 {
	if tm.routeTime > 0 && tm.routeTime < 1 {
		return 1 / tm.routeTime
	}
	return 1
}

// priceLevel is rules' RoutePriceLevel (kept by the set, so the state view
// can read it under a read lock).
func (tm *TradeManager) priceLevel(res, age string) float64 {
	return tm.rules.RoutePriceLevel(res, age)
}

// routeKeys is the started routes' keys in order, kept between calls.
func (tm *TradeManager) routeKeys() []string {
	if len(tm.order) != len(tm.activeRoutes) {
		tm.order = sortedKeys(tm.activeRoutes)
	}
	return tm.order
}

// idle names the export the town makes none of, or the good no age prices.
// "" when the route can run.
func (tm *TradeManager) idle(def config.TradeRouteDef, age string) string {
	for _, res := range def.Export {
		if tm.gross[res] <= 0 || tm.priceLevel(res, age) <= 0 {
			return res
		}
	}
	for _, res := range def.Import {
		if tm.priceLevel(res, age) <= 0 {
			return res
		}
	}
	return ""
}

// flows is what def moves a tick at its gate. It is sized on the import: the
// first import the town makes comes in at config.RouteShare of that make, and
// the export is what is worth that less config.RouteMargin (by price level, as
// dealRate), so the bonuses below raise the import and not the export. A town
// that makes no import yet sizes it from the export (config.RouteShare of the
// make) so the route can bring a new good. The techs' term on route time
// multiplies both; the techs' pay, the harbor bonus and the allies' shares go
// on the import. The shares are for the state view.
func (tm *TradeManager) flows(def config.TradeRouteDef, terms RouteTerms, age string) (out, in map[string]float64, shares map[string][]PayShare) {
	k := tm.timeFactor()
	out, in = make(map[string]float64, len(def.Export)), make(map[string]float64, len(def.Import))
	if !terms.Lean {
		shares = make(map[string][]PayShare, len(def.Import))
	}
	level := func(res string) float64 { return tm.priceLevel(res, age) }
	// worth is the value, in price units, each import brings before the bonuses.
	worth := 0.0
	for _, res := range def.Import {
		if g := tm.gross[res]; g > 0 && level(res) > 0 {
			worth = float64(config.RouteShare*g*k) / level(res)
			break
		}
	}
	if worth > 0 {
		sent := float64(worth*float64(len(def.Import))) / config.RouteMargin
		for _, res := range def.Export {
			out[res] = float64(sent / float64(len(def.Export)) * level(res))
		}
	} else {
		for _, res := range def.Export {
			out[res] = float64(config.RouteShare * tm.gross[res] * k)
			if l := level(res); l > 0 {
				worth += out[res] / l
			}
		}
		worth *= config.RouteMargin / float64(max(len(def.Import), 1))
	}
	for _, res := range def.Import {
		bonus := terms.Harbor
		if terms.Lean {
			if terms.AllyBonus != nil {
				b, ok := tm.allyMemo[res]
				if !ok {
					b = terms.AllyBonus(res)
					tm.allyMemo[res] = b
				}
				bonus += b
			}
			if p := level(res); p > 0 {
				in[res] = float64(tm.RunPay(worth*p, bonus))
			}
			continue
		}
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
		if p := level(res); p > 0 {
			in[res] = float64(tm.RunPay(worth*p, bonus))
		}
		shares[res] = list
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

// ActiveRouteCount returns the number of started, unpaused trade routes. Used by
// the diplomacy system to drive mercantile-civ opinion drift. Must be called
// under the engine write lock (same as the rest of the manager).
func (tm *TradeManager) ActiveRouteCount() int {
	n := 0
	for _, r := range tm.activeRoutes {
		if !r.Paused {
			n++
		}
	}
	return n
}

// Flows puts the running routes into the resource rates, after production:
// each export is a consumer and each import a producer, with its own line in
// the breakdown. Routes compete for a resource: with D the routes' demand for
// it a tick and S its stock plus what the town makes of it, every route
// exporting it runs at min(1, S/D), both legs, so a shortage slows the routes
// and never skips one. horizon is the ticks the rates will be applied for
// (1 on a live tick; an offline step's ticks), so a long step spends the
// stock once.
func (tm *TradeManager) Flows(rm *ResourceManager, buildings *BuildingManager, diplomacy *DiplomacyManager, terms RouteTerms, age string, horizon float64) {
	horizon = max(horizon, 1)
	if tm.gross == nil {
		tm.gross = map[string]float64{}
	}
	if tm.allyMemo == nil {
		tm.allyMemo, tm.demand, tm.share = map[string]float64{}, map[string]float64{}, map[string]float64{}
	}
	clear(tm.allyMemo)
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
	legs, demand, share := tm.legs[:0], tm.demand, tm.share
	clear(demand)
	clear(share)
	for _, key := range tm.routeKeys() {
		route, def := tm.activeRoutes[key], tm.routeDefs[key]
		if def.Key == "" { // a saved route the game no longer has
			route.Run, route.Scale, route.Research = 0, 0, 0
			continue
		}
		owned := buildings.CountFrom(def.RequiredBld)
		route.Run, route.Research = 0, 0
		route.Owned, route.Scale = owned, 0
		route.NoBuilding = owned < def.MinCount
		if !route.NoBuilding {
			route.Scale = 1 + detmath.Log(float64(owned)/float64(max(def.MinCount, 1)))/math.Ln2
		}
		route.Disrupted = tm.routeDisruptedBy(def, disrupted) != ""
		route.NoIncome = tm.idle(def, age)
		if route.Paused || route.NoBuilding || route.Disrupted || route.NoIncome != "" {
			continue
		}
		out, in := tm.sized(route, def, terms, age)
		for res, a := range out {
			demand[res] += a
		}
		legs = append(legs, routeLeg{route, out, in})
	}
	for res, d := range demand {
		if r := rm.resources[res]; r != nil {
			// Only a surplus goes out: the net rate (what the town makes less
			// what it uses) and, while that is not negative, the stock above
			// what the town keeps (recruitFoodMarginShare of its make). That
			// share stays, so a route never runs the stock dry or stops the
			// people growing.
			keep := float64(recruitFoodMarginShare * tm.gross[res])
			supply := max(0, r.Rate-keep)
			if r.Rate >= 0 {
				supply += max(0, r.Amount-keep) / horizon
			}
			share[res] = min(1, supply/d)
		}
	}
	tm.legs = legs
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

// sized is def's flows at the route's scale. They are kept on the route and
// worked out again only when the age, the techs, the bonuses, the town's
// income of a good or the route's scale changes.
func (tm *TradeManager) sized(route *ActiveRoute, def config.TradeRouteDef, terms RouteTerms, age string) (out, in map[string]float64) {
	k := flowKey{age: age, speed: tm.timeFactor(), pay: tm.RoutePay(1), harbor: terms.Harbor, scale: route.Scale}
	if len(def.Export) > 2 || len(def.Import) > 2 || !terms.Lean {
		out, in, _ = tm.flows(def, terms, age)
		for res, a := range out {
			out[res] = float64(a * route.Scale)
		}
		for res, a := range in {
			in[res] = float64(a * route.Scale)
		}
		return out, in
	}
	for i, res := range def.Export {
		k.gross[i] = tm.gross[res]
	}
	for i, res := range def.Import {
		k.gross[2+i] = tm.gross[res]
		if terms.AllyBonus != nil {
			b, ok := tm.allyMemo[res]
			if !ok {
				b = terms.AllyBonus(res)
				tm.allyMemo[res] = b
			}
			k.ally[i] = b
		}
	}
	if route.sized.ok && route.sized.key == k {
		return route.sized.out, route.sized.in
	}
	out, in, _ = tm.flows(def, terms, age)
	for res, a := range out {
		out[res] = float64(a * route.Scale)
	}
	for res, a := range in {
		in[res] = float64(a * route.Scale)
	}
	route.sized = sizedFlow{key: k, ok: true, out: out, in: in}
	return out, in
}

// researchRate is the research a tick the running routes bring.
func (tm *TradeManager) researchRate() float64 {
	var w float64
	for _, r := range tm.activeRoutes {
		w += r.Research
	}
	return w
}

// Tick is a live tick's trade: the research the routes bring, a word in the
// log when a route stops running or runs again, and once a minute a warmer
// opinion among the civilizations for each route that runs. The goods
// themselves move in Flows.
func (tm *TradeManager) Tick(tick int, diplomacy *DiplomacyManager, research func(work float64) string) []logNote {
	var messages []logNote
	minute := tick%max(minutesToTicks(1), 1) == 0
	for _, key := range tm.routeKeys() {
		route, def := tm.activeRoutes[key], tm.routeDefs[key]
		if def.Key == "" {
			continue
		}
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
	for _, key := range tm.routeKeys() {
		route := tm.activeRoutes[key]
		def := allRoutes[key]
		blockedBy := tm.routeDisruptedBy(def, disrupted)
		out, in, shares := tm.flows(def, terms, age)
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
		out, in, shares := tm.flows(def, terms, age)
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
		Gross:              maps.Clone(tm.gross),
	}
}

// LoadState restores trade state from save.
func (tm *TradeManager) LoadState(s TradeSave) {
	for key, route := range s.ActiveRoutes {
		route := route // copy
		tm.activeRoutes[key] = &route
	}
}
