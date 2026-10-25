package config

import (
	"strings"
	"testing"
)

// Stock has no ceiling, so nothing raises one: no building, tech, milestone
// reward or event carries a storage effect, and no effect aims at a store
// ("all" or "<resource>_storage"). A tech that once raised every store
// raises all production now, by a little under a tenth of what it raised storage by.

// isStorageEffect reports whether e raises a store: a "storage" effect or one
// aimed at "<resource>_storage".
func isStorageEffect(e Effect) bool {
	return e.Type == "storage" || strings.HasSuffix(e.Target, "_storage")
}

func TestNothingHasAStorageEffect(t *testing.T) {
	isResKey := func(k string) bool { _, ok := ResourceByKey()[k]; return ok }
	for _, d := range BaseBuildings() {
		for _, e := range d.Effects {
			if isStorageEffect(e) {
				t.Errorf("building %s has a storage effect: %+v", d.Key, e)
			}
		}
	}
	for _, m := range Milestones() {
		for _, e := range m.Rewards {
			if isStorageEffect(e) {
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
			if isStorageEffect(e) {
				t.Errorf("event %s has a storage effect: %+v", ev.Key, e)
			}
		}
	}
	for _, list := range [][]EpochEventDef{GoodEpochEvents(), ChallengingEpochEvents()} {
		for _, ev := range list {
			for _, e := range ev.Rates {
				if isStorageEffect(e) {
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
