package game

import (
	"sort"
	"testing"
)

// TestTradeSnapshot_ActiveRoutesInKeyOrder: the running routes come back in
// key order on every snapshot, so the routes list does not jump between
// frames (the snapshot used to walk the map in its random order).
func TestTradeSnapshot_ActiveRoutesInKeyOrder(t *testing.T) {
	ge := NewGameEngine()
	var keys []string
	for _, def := range ge.Trade.routeList {
		keys = append(keys, def.Key)
	}
	if len(keys) < 6 {
		t.Fatalf("only %d trade routes in the rules, want at least 6 to see an order", len(keys))
	}
	sort.Strings(keys)
	keys = keys[:6]
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for _, key := range keys {
		ge.Trade.activeRoutes[key] = &ActiveRoute{Key: key, TicksLeft: 5}
	}
	for i := 0; i < 50; i++ {
		got := ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil).ActiveRoutes
		if len(got) != len(keys) {
			t.Fatalf("snapshot %d lists %d running routes, want %d", i, len(got), len(keys))
		}
		for j, info := range got {
			if info.Key != keys[j] {
				t.Fatalf("snapshot %d lists route %d as %q, want %q (key order)", i, j, info.Key, keys[j])
			}
		}
	}
}
