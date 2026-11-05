package mapmodel

// Dinosaur and Late Cretaceous asteroid ambient life has been removed from the
// Primitive Age. The Primitive Age is grounded in human prehistory with clean
// wilderness, foragers, flint tools, and campfires.

// The ambient kinds (deprecated).
const (
	SightSauropod    SightingKind = iota + SightWalker + 1 // a long-necked giant, slow (deprecated)
	SightRaptor                                            // a small fast walker (deprecated)
	SightPterodactyl                                       // a flyby across the upper rows (deprecated)
	SightAsteroid                                          // the fireball (deprecated)
)

// DinoFactor returns 0 as dinosaurs have been removed.
func DinoFactor(share float64) float64 {
	return 0
}

// DinoSlots returns 0.
func DinoSlots(k int, share float64) int {
	return 0
}

// DinosAt returns nil: no dinosaurs roam the map.
func DinosAt(seed int64, share float64, anim int) []Sighting {
	return nil
}

// Dinos lists the dinosaurs showing at animation frame anim: none.
func (m *Model) Dinos(anim int) []Sighting {
	return nil
}

// AsteroidShare returns 1 (no asteroid window active in Primitive Age).
func AsteroidShare() float64 {
	return 1
}

// AsteroidAt returns false (no asteroid streak in Primitive Age).
func (m *Model) AsteroidAt(anim int) (Sighting, bool) {
	return Sighting{}, false
}
