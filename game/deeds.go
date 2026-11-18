package game

import "maps"

// deeds.go counts what a run has done inside the age it is in. The counts are
// plain ints, saved with the run's stats and zeroed on entering an age; the
// callers hold the engine's write lock and nothing here takes a lock.

// AgeDeeds counts what the run has done inside the age it is in.
type AgeDeeds struct {
	Festivals    int            `json:"festivals,omitempty"`     // festivals held
	Upgrades     int            `json:"upgrades,omitempty"`      // upgrades bought
	Pacts        int            `json:"pacts,omitempty"`         // deals, gifts and alliances made
	KeystoneTick int            `json:"keystone_tick,omitempty"` // tick the age's keystone tech was researched, 0 = not yet
	RouteCycles  map[string]int `json:"route_cycles,omitempty"`  // cycles a trade route finished, by route key
	RouteClean   map[string]int `json:"route_clean,omitempty"`   // tick a route's unbroken run began; removed on a miss or a stop
	Wars         map[string]int `json:"wars,omitempty"`          // tick a war began in the age, by faction key
	Tributes     int            `json:"tributes,omitempty"`      // tributes paid
}

// clone copies d so that it shares no map with it.
func (d AgeDeeds) clone() AgeDeeds {
	d.RouteCycles, d.RouteClean, d.Wars = maps.Clone(d.RouteCycles), maps.Clone(d.RouteClean), maps.Clone(d.Wars)
	return d
}

// RouteCycle notes a finished cycle of route at tick. The unbroken run began
// at the first cycle finished since the last miss, stop or age entered.
func (d *AgeDeeds) RouteCycle(route string, tick int) {
	if d == nil {
		return
	}
	if d.RouteCycles == nil {
		d.RouteCycles, d.RouteClean = map[string]int{}, map[string]int{}
	}
	d.RouteCycles[route]++
	if _, ok := d.RouteClean[route]; !ok {
		d.RouteClean[route] = tick
	}
}

// RouteBroken notes a missed cycle or a stop of route: its unbroken run ends.
func (d *AgeDeeds) RouteBroken(route string) {
	if d != nil {
		delete(d.RouteClean, route)
	}
}

// WarBegan notes that a war with faction began at tick.
func (d *AgeDeeds) WarBegan(faction string, tick int) {
	if d == nil {
		return
	}
	if d.Wars == nil {
		d.Wars = map[string]int{}
	}
	d.Wars[faction] = tick
}
