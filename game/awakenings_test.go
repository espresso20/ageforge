package game

import (
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// activeEventByKey returns the active ActiveEvent with the given key, or nil.
func activeEventByKey(ge *GameEngine, key string) *ActiveEvent {
	for i := range ge.Events.active {
		if ge.Events.active[i].Key == key {
			return &ge.Events.active[i]
		}
	}
	return nil
}

// TestAwakening_FiresOnceOnTriggerAge verifies that advancing into an epoch's
// trigger age fires that epoch's awakening exactly once: it injects the timed
// boost event and records the awakening as fired.
func TestAwakening_FiresOnceOnTriggerAge(t *testing.T) {
	ge := NewGameEngine()

	// stone_age is the Stone Era's awakening trigger (Pottery Mastery).
	def, ok := config.AwakeningForAge("stone_age")
	if !ok {
		t.Fatal("expected an awakening to trigger on stone_age")
	}

	if ge.awakeningsFired[def.Key] {
		t.Fatalf("awakening %q should not be fired before advancing", def.Key)
	}

	ge.advanceAge("stone_age")

	if !ge.awakeningsFired[def.Key] {
		t.Errorf("awakening %q should be recorded as fired after advancing to stone_age", def.Key)
	}

	ae := activeEventByKey(ge, def.Key)
	if ae == nil {
		t.Fatalf("awakening %q should have injected an active event", def.Key)
	}
	if ae.Name != def.Name {
		t.Errorf("injected event name = %q, want %q", ae.Name, def.Name)
	}
	if ae.TicksLeft != def.Duration {
		t.Errorf("injected event TicksLeft = %d, want %d", ae.TicksLeft, def.Duration)
	}
	if len(ae.Effects) != len(def.Effects) {
		t.Errorf("injected event has %d effects, want %d", len(ae.Effects), len(def.Effects))
	}
}

// TestAwakening_DoesNotRefireOnReentry verifies that advancing into the same
// trigger age twice fires the awakening only once (no duplicate injected event,
// fired flag stays set).
func TestAwakening_DoesNotRefireOnReentry(t *testing.T) {
	ge := NewGameEngine()
	def, _ := config.AwakeningForAge("stone_age")

	ge.advanceAge("stone_age")
	// Drain the first injected event so a second fire would be unmistakable.
	ge.Events.active = nil

	// Re-enter stone_age (e.g. an age-set edge case). Must NOT re-fire.
	ge.advanceAge("stone_age")

	if ae := activeEventByKey(ge, def.Key); ae != nil {
		t.Errorf("awakening %q re-fired on re-entry; should fire at most once per run", def.Key)
	}
	if !ge.awakeningsFired[def.Key] {
		t.Errorf("awakening %q fired flag should remain set", def.Key)
	}
}

// TestAwakening_NoFireOnNonTriggerAge verifies that advancing into an age that is
// not any awakening's trigger age fires nothing. primitive_age (the starting age)
// is deliberately not an awakening trigger.
func TestAwakening_NoFireOnNonTriggerAge(t *testing.T) {
	ge := NewGameEngine()

	if _, ok := config.AwakeningForAge("primitive_age"); ok {
		t.Fatal("primitive_age should NOT be an awakening trigger age")
	}

	ge.advanceAge("primitive_age")

	if len(ge.awakeningsFired) != 0 {
		t.Errorf("no awakening should fire on primitive_age, got fired set: %v", ge.awakeningsFired)
	}
}

// TestAwakening_ReloadDoesNotRefire verifies the one-time guard survives a real
// save/load round-trip: after firing an awakening and saving, loading the save and
// re-advancing into the trigger age does NOT re-fire it.
func TestAwakening_ReloadDoesNotRefire(t *testing.T) {
	ge := NewGameEngine()
	def, _ := config.AwakeningForAge("stone_age")

	ge.advanceAge("stone_age")
	if !ge.awakeningsFired[def.Key] {
		t.Fatalf("precondition: awakening %q should be fired", def.Key)
	}

	if err := ge.SaveGame("test_awakening_reload"); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove("data/saves/test_awakening_reload.json")

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("test_awakening_reload"); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	// Fired set must have survived the round-trip.
	if !ge2.awakeningsFired[def.Key] {
		t.Errorf("awakening %q fired flag did not persist across save/load", def.Key)
	}

	// Re-advancing into the trigger age on the loaded engine must not re-fire.
	ge2.Events.active = nil
	ge2.advanceAge("stone_age")
	if ae := activeEventByKey(ge2, def.Key); ae != nil {
		t.Errorf("awakening %q re-fired after reload; one-time guard not honoured", def.Key)
	}
}

// TestAwakening_PrestigeClearsFiredSet verifies that the fired set resets on
// prestige/reset so awakenings can fire again next run. We poke the reset path
// that DoPrestige/Reset use (re-initialising the map) and confirm a fresh fire.
func TestAwakening_PrestigeClearsFiredSet(t *testing.T) {
	ge := NewGameEngine()
	def, _ := config.AwakeningForAge("stone_age")

	ge.advanceAge("stone_age")
	if !ge.awakeningsFired[def.Key] {
		t.Fatalf("precondition: awakening %q should be fired", def.Key)
	}

	// Reset (mirrors the prestige/succumb/wipe paths, which all re-init this map
	// alongside epochEventFired). After reset the run should be able to fire again.
	ge.Reset()

	if len(ge.awakeningsFired) != 0 {
		t.Errorf("awakeningsFired should be cleared after Reset, got: %v", ge.awakeningsFired)
	}

	// Bring the engine back to the age before the trigger so advanceAge crosses in.
	ge.age = "primitive_age"
	ge.advanceAge("stone_age")
	if !ge.awakeningsFired[def.Key] {
		t.Errorf("awakening %q should fire again in the new run after reset", def.Key)
	}
	if activeEventByKey(ge, def.Key) == nil {
		t.Errorf("awakening %q should re-inject its event in the new run", def.Key)
	}
}

// TestAwakening_AllDefinitionsValid verifies there are exactly 7 awakenings (one
// per epoch) and that every awakening references real config: a valid epoch key, a
// trigger age that exists and belongs to that epoch, a non-empty name/flavor, a
// positive duration, and effects whose targets are real resources (or no target for
// production_all). Guards against typos in keys/targets at the config layer.
func TestAwakening_AllDefinitionsValid(t *testing.T) {
	awakenings := config.Awakenings()

	epochs := config.EpochByKey()
	if len(awakenings) != len(epochs) {
		t.Errorf("got %d awakenings, want one per epoch (%d)", len(awakenings), len(epochs))
	}

	// Build the set of valid resource keys for effect-target validation.
	validResource := make(map[string]bool)
	for _, r := range config.BaseResources() {
		validResource[r.Key] = true
	}

	seenEpoch := make(map[string]bool)
	seenKey := make(map[string]bool)

	for _, a := range awakenings {
		if a.Key == "" {
			t.Errorf("awakening for epoch %q has empty Key", a.EpochKey)
		}
		if seenKey[a.Key] {
			t.Errorf("duplicate awakening key %q", a.Key)
		}
		seenKey[a.Key] = true

		ep, ok := epochs[a.EpochKey]
		if !ok {
			t.Errorf("awakening %q references unknown epoch %q", a.Key, a.EpochKey)
			continue
		}
		if seenEpoch[a.EpochKey] {
			t.Errorf("epoch %q has more than one awakening", a.EpochKey)
		}
		seenEpoch[a.EpochKey] = true

		// Trigger age must exist and belong to this awakening's epoch.
		if config.EpochForAge(a.TriggerAge) != a.EpochKey {
			t.Errorf("awakening %q trigger age %q does not belong to epoch %q",
				a.Key, a.TriggerAge, a.EpochKey)
		}
		ageInEpoch := false
		for _, age := range ep.Ages {
			if age == a.TriggerAge {
				ageInEpoch = true
				break
			}
		}
		if !ageInEpoch {
			t.Errorf("awakening %q trigger age %q not listed in epoch %q ages %v",
				a.Key, a.TriggerAge, a.EpochKey, ep.Ages)
		}

		if a.Name == "" {
			t.Errorf("awakening %q has empty Name", a.Key)
		}
		if a.FlavorText == "" {
			t.Errorf("awakening %q has empty FlavorText", a.Key)
		}
		if a.Duration <= 0 {
			t.Errorf("awakening %q has non-positive Duration %d", a.Key, a.Duration)
		}
		if len(a.Effects) == 0 {
			t.Errorf("awakening %q has no effects", a.Key)
		}
		for _, eff := range a.Effects {
			switch eff.Type {
			case "production_all":
				// No target required; value is a multiplier.
				if eff.Value <= 0 {
					t.Errorf("awakening %q production_all value %.2f should be positive", a.Key, eff.Value)
				}
			case "production":
				if !validResource[eff.Target] {
					t.Errorf("awakening %q production effect targets unknown resource %q", a.Key, eff.Target)
				}
				if eff.Value <= 0 {
					t.Errorf("awakening %q production effect on %q has non-positive value %.2f",
						a.Key, eff.Target, eff.Value)
				}
			default:
				t.Errorf("awakening %q uses unexpected effect type %q (use production / production_all)",
					a.Key, eff.Type)
			}
		}
	}

	// Sanity: the card's 7 named trigger ages must each resolve to an awakening.
	for _, age := range []string{
		"stone_age", "iron_age", "industrial_age", "victorian_age",
		"modern_age", "cyberpunk_age", "interstellar_age",
	} {
		if _, ok := config.AwakeningForAge(age); !ok {
			t.Errorf("expected an awakening to trigger on %q", age)
		}
	}
}
