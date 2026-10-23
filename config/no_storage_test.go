package config

import (
	"math"
	"strings"
	"testing"
)

// Stock has no ceiling, so nothing raises one: no building, tech, milestone
// reward or event carries a storage effect, and no effect aims at a store
// ("all" or "<resource>_storage"). A tech that once raised every store
// raises all production now, by a little under a tenth of what it raised storage by.

func TestNothingHasAStorageEffect(t *testing.T) {
	isResKey := func(k string) bool { _, ok := ResourceByKey()[k]; return ok }
	for _, d := range BaseBuildings() {
		for _, e := range d.Effects {
			if e.Type == "storage" || strings.HasSuffix(e.Target, "_storage") {
				t.Errorf("building %s has a storage effect: %+v", d.Key, e)
			}
		}
	}
	for _, m := range Milestones() {
		for _, e := range m.Rewards {
			if e.Type == "storage" || strings.HasSuffix(e.Target, "_storage") {
				t.Errorf("milestone %s rewards storage: %+v", m.Key, e)
			}
			// A permanent bonus aimed at "all" or at a bare resource was a
			// store's flat size; nothing reads one now.
			if e.Type == "permanent_bonus" && (e.Target == "all" || isResKey(e.Target)) {
				t.Errorf("milestone %s gives a flat bonus to %q, which only a store read: %+v", m.Key, e.Target, e)
			}
		}
	}
	for _, ev := range RandomEvents() {
		for _, e := range ev.Effects {
			if e.Type == "storage" || strings.HasSuffix(e.Target, "_storage") {
				t.Errorf("event %s has a storage effect: %+v", ev.Key, e)
			}
		}
	}
	for _, list := range [][]EpochEventDef{GoodEpochEvents(), ChallengingEpochEvents()} {
		for _, ev := range list {
			for _, e := range ev.Rates {
				if e.Type == "storage" || strings.HasSuffix(e.Target, "_storage") {
					t.Errorf("epoch event %s has a storage effect: %+v", ev.Key, e)
				}
			}
		}
	}
	for _, tc := range Technologies() {
		for _, e := range tc.Effects {
			if string(e.Kind) == "storage" {
				t.Errorf("tech %s has a storage effect", tc.Key)
			}
		}
	}
	for _, k := range TechEffectKinds() {
		if string(k) == "storage" {
			t.Error("the tech effect kinds still list storage")
		}
	}
	isRes := func(k string) bool { _, ok := ResourceByKey()[k]; return ok }
	if msg := (TechEffect{Kind: "storage", Value: 0.1}).Check(isRes); !strings.Contains(msg, "unknown effect kind") {
		t.Errorf("a storage tech effect checks as %q, want it refused as an unknown kind", msg)
	}
}

// The seven techs that raised every store raise all production instead, by
// about a tenth of the share they raised storage by (0.8, 0.8, 0.4, 0.4, 0.6,
// 0.6 and 0.6 percent).
func TestStorageTechsRaiseProductionInstead(t *testing.T) {
	want := map[string]float64{ // the share of all production each gives now
		"pottery": 0.008, "masonry": 0.008, "prefabrication": 0.004, "grid_storage": 0.004,
		"cloud_computing": 0.006, "superconductors": 0.006, "space_elevator": 0.006,
	}
	byKey := map[string]TechDef{}
	for _, tc := range Technologies() {
		byKey[tc.Key] = tc
	}
	for key, w := range want {
		tc, ok := byKey[key]
		if !ok {
			t.Errorf("no tech %s", key)
			continue
		}
		got := 0.0
		for _, e := range tc.Effects {
			if e.Kind == EffectAllOutput {
				got += e.Value
			}
		}
		if math.Abs(got-w) > 1e-12 {
			t.Errorf("%s raises all production by %v, want %v", key, got, w)
		}
		if len(TechChecks(tc, BaseBuildings(), FeatureLocks())) == 0 {
			t.Errorf("%s promises nothing", key)
		}
	}
}
