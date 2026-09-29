package mapmodel

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestRouteModes: route modes come from the route data, and a route the
// catalogue does not know still gets the same mode every time.
func TestRouteModes(t *testing.T) {
	c := NewCatalog()
	want := map[string]RouteMode{
		"local_barter": ModeLand, "silk_road": ModeLand, "rail_freight": ModeLand, "oil_pipeline": ModeLand,
		"spice_trade": ModeSea, "tea_clippers": ModeSea, "coal_barges": ModeSea, "steamship_line": ModeSea,
		"warp_commerce": ModeAir, "stellar_exchange": ModeAir,
	}
	for k, m := range want {
		if got := c.RouteMode(k); got != m {
			t.Errorf("%s: %s, want %s", k, got, m)
		}
	}
	for _, r := range config.BaseTradeRoutes() {
		t.Logf("%-20s %s", r.Key, c.RouteMode(r.Key))
	}
	if a, b := c.RouteMode("no_such_route"), c.RouteMode("no_such_route"); a != b {
		t.Error("an unknown route's mode is not stable")
	}
}
