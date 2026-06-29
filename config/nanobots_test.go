package config

import "testing"

// Nanobot System (Trello 0kLti5GR). These tests pin the three pieces that give
// nanobots a real economic role: a Modern-age producer, the nanobot tech line,
// and digital/cyberpunk buildings that consume nanobots as a build material.

// prodValue returns a building/tech's "production" effect value for a target.
func prodEffectValue(effects []Effect, target string) (float64, bool) {
	for _, e := range effects {
		if e.Type == "production" && e.Target == target {
			return e.Value, true
		}
	}
	return 0, false
}

func bonusEffectValue(effects []Effect, target string) (float64, bool) {
	for _, e := range effects {
		if e.Type == "bonus" && e.Target == target {
			return e.Value, true
		}
	}
	return 0, false
}

// TestNanobots_FoundryProducerExists asserts the Modern-age nano_foundry exists,
// is gated to modern_age, produces a positive amount of nanobots, has a worker
// domain, and is wired into the Modern Age's unlock list.
func TestNanobots_FoundryProducerExists(t *testing.T) {
	def, ok := BuildingByKey()["nano_foundry"]
	if !ok {
		t.Fatal("nano_foundry building is not defined")
	}
	if def.RequiredAge != "modern_age" {
		t.Errorf("nano_foundry RequiredAge = %q, want modern_age", def.RequiredAge)
	}
	if def.Category != "production" {
		t.Errorf("nano_foundry Category = %q, want production", def.Category)
	}
	rate, hasProd := prodEffectValue(def.Effects, "nanobots")
	if !hasProd || rate <= 0 {
		t.Errorf("nano_foundry must produce nanobots; got rate %v (present=%v)", rate, hasProd)
	}
	if def.WorkerDomain == "" || def.WorkerCapacity <= 0 {
		t.Errorf("nano_foundry needs a worker domain + capacity; got domain=%q cap=%d",
			def.WorkerDomain, def.WorkerCapacity)
	}

	// Must be unlocked by the Modern Age (reachability + it shows up in views).
	unlocked := false
	for _, age := range Ages() {
		if age.Key != "modern_age" {
			continue
		}
		for _, k := range age.UnlockBuildings {
			if k == "nano_foundry" {
				unlocked = true
			}
		}
	}
	if !unlocked {
		t.Error("nano_foundry is not listed in modern_age UnlockBuildings")
	}
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

// TestNanobots_TechsExistAndGated asserts the three nanobot techs exist, are
// gated to modern+ ages, cost knowledge, and carry their intended effects.
func TestNanobots_TechsExistAndGated(t *testing.T) {
	byKey := TechByKey()
	ageOrder := map[string]int{}
	for _, a := range Ages() {
		ageOrder[a.Key] = a.Order
	}
	modernOrder := ageOrder["modern_age"]

	for _, key := range []string{"nanofabrication", "medical_nanobots", "self_replication"} {
		tech, ok := byKey[key]
		if !ok {
			t.Errorf("tech %q is not defined", key)
			continue
		}
		if ageOrder[tech.Age] < modernOrder {
			t.Errorf("tech %q age %q is before modern_age", key, tech.Age)
		}
		if tech.Cost <= 0 {
			t.Errorf("tech %q has non-positive knowledge cost %v", key, tech.Cost)
		}
	}

	// Nanofabrication reduces build cost (negative build_cost bonus).
	if v, ok := bonusEffectValue(byKey["nanofabrication"].Effects, "build_cost"); !ok || v >= 0 {
		t.Errorf("nanofabrication must reduce build_cost; got %v (present=%v)", v, ok)
	}
	// Self-Replication boosts nanobot production.
	if v, ok := prodEffectValue(byKey["self_replication"].Effects, "nanobots"); !ok || v <= 0 {
		t.Errorf("self_replication must add nanobot production; got %v (present=%v)", v, ok)
	}
	// Medical Nanobots is a clearly-beneficial tech (pop cap and/or food).
	mn := byKey["medical_nanobots"].Effects
	_, hasPop := func() (float64, bool) {
		for _, e := range mn {
			if e.Type == "capacity" && e.Target == "population" {
				return e.Value, true
			}
		}
		return 0, false
	}()
	_, hasFood := prodEffectValue(mn, "food")
	if !hasPop && !hasFood {
		t.Error("medical_nanobots must grant a beneficial effect (population cap and/or food)")
	}
}

// TestNanobots_ConsumedAsBuildMaterial asserts that several digital/cyberpunk
// buildings now require nanobots to construct, giving produced nanobots a sink.
func TestNanobots_ConsumedAsBuildMaterial(t *testing.T) {
	want := []string{
		"data_center", "neural_grid", "quantum_battery_array", "nano_alloy_plant",
		"cyber_hub", "augmentation_foundry", "dark_energy_tap",
	}
	byKey := BuildingByKey()
	count := 0
	for _, key := range want {
		def, ok := byKey[key]
		if !ok {
			t.Errorf("expected building %q not found", key)
			continue
		}
		if amt, has := def.BaseCost["nanobots"]; !has || amt <= 0 {
			t.Errorf("building %q should cost nanobots to build; got %v (present=%v)", key, amt, has)
		} else {
			count++
		}
	}
	if count < 4 {
		t.Errorf("expected at least 4 buildings to consume nanobots, got %d", count)
	}
}
