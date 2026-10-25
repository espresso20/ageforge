package config

import "testing"

// Nanobot System (Trello 0kLti5GR). Nanobots unlock as a resource in the Modern
// age, so a nanobot producer must be buildable by then.

// prodValue returns a building/tech's "production" effect value for a target.
func prodEffectValue(effects []Effect, target string) (float64, bool) {
	for _, e := range effects {
		if e.Type == "production" && e.Target == target {
			return e.Value, true
		}
	}
	return 0, false
}

// TestNanobots_ProducerUnlocksWithResource guards the core fix: nanobots unlock
// as a resource in the Modern Age, and there must now be a nanobot producer
// buildable in that same age (previously the first producer was 2 ages later).
func TestNanobots_ProducerUnlocksWithResource(t *testing.T) {
	ageOrder := map[string]int{}
	for _, a := range Ages() {
		ageOrder[a.Key] = a.Order
	}
	resAge, ok := ResourceByKey()["nanobots"]
	if !ok {
		t.Fatal("nanobots resource not defined")
	}
	unlockOrder := ageOrder[resAge.Age]

	earliest := 1 << 30
	for _, b := range BaseBuildings() {
		if _, prod := prodEffectValue(b.Effects, "nanobots"); !prod {
			continue
		}
		if o := ageOrder[b.RequiredAge]; o < earliest {
			earliest = o
		}
	}
	if earliest > unlockOrder {
		t.Errorf("no nanobot producer at or before the unlock age %q (order %d); earliest producer age order = %d",
			resAge.Age, unlockOrder, earliest)
	}
}
