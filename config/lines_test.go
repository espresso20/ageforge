package config

import (
	"sort"
	"strings"
	"testing"
)

// lineMakes is the resources a building makes: its production effects, main
// or side output.
func lineMakes(d BuildingDef) map[string]bool {
	out := map[string]bool{}
	for _, e := range d.Effects {
		if e.Type == "production" && e.Value > 0 {
			out[e.Target] = true
		}
	}
	return out
}

func lineList(m map[string]bool) string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	sort.Strings(s)
	return strings.Join(s, "+")
}

// productionLines groups every lineage building (storage, wonders and
// monuments are not lines) by its line.
func productionLines() map[string][]BuildingDef {
	lines := map[string][]BuildingDef{}
	for _, d := range BaseBuildings() {
		switch d.LineageKey {
		case "", "storage", "wonder", "monument":
			continue
		}
		lines[d.Line()] = append(lines[d.Line()], d)
	}
	return lines
}

// lineResource is the resource a line makes: for a line split off a lineage
// it is the suffix of the line key (<lineage>_<resource>); for a lineage that
// never changes resource it is the resource all its producing buildings share
// ("" when none of them produces, as housing and knowledge).
func lineResource(line string, defs []BuildingDef) (res string, bad bool) {
	if defs[0].LineKey != "" {
		return strings.TrimPrefix(line, defs[0].LineageKey+"_"), false
	}
	var common map[string]bool
	for _, d := range defs {
		got := lineMakes(d)
		if common == nil {
			common = got
			continue
		}
		for r := range common {
			if !got[r] {
				delete(common, r)
			}
		}
	}
	if len(common) == 0 {
		for _, d := range defs {
			if len(lineMakes(d)) > 0 {
				return "", true
			}
		}
		return "", false
	}
	return lineList(common), false
}

// TestLinesNeverSwitchResource: a building line makes one resource from its
// first rung to its last (pausing and resuming is still one line).
func TestLinesNeverSwitchResource(t *testing.T) {
	lines := productionLines()
	for line, defs := range lines {
		for _, d := range defs {
			if (d.LineKey != "") != (defs[0].LineKey != "") || d.LineageKey != defs[0].LineageKey {
				t.Errorf("line %s mixes lineages or keyed and unkeyed buildings: %s", line, d.Key)
			}
		}
		res, bad := lineResource(line, defs)
		if bad {
			for _, d := range defs {
				t.Errorf("line %s has no resource in common: %s (%s) makes %s", line, d.Key, d.RequiredAge, lineList(lineMakes(d)))
			}
			continue
		}
		if res == "" {
			continue // a line that makes nothing (housing, knowledge)
		}
		if defs[0].LineKey != "" {
			for _, d := range defs {
				if !lineMakes(d)[res] {
					t.Errorf("line %s makes %s, but %s (%s) makes %s", line, res, d.Key, d.RequiredAge, lineList(lineMakes(d)))
				}
			}
		}
	}
	if len(lines) != 39 {
		t.Errorf("%d building lines, want 39", len(lines))
	}
}

// TestUpgradesKeepTheResource: whatever a building upgrades to as an age
// turns, it is the next rung of the same line and makes the same resource.
func TestUpgradesKeepTheResource(t *testing.T) {
	defs := BaseBuildings()
	offers := 0
	for line, members := range productionLines() {
		res, _ := lineResource(line, members)
		for _, d := range members {
			next, ok := NextRung(defs, d.Line(), d.LineageTier)
			if !ok {
				continue
			}
			offers++
			if next.Line() != line || next.LineageTier <= d.LineageTier {
				t.Errorf("%s upgrades to %s, which is not a later rung of line %s", d.Key, next.Key, line)
			}
			if d.LineKey != "" && (!lineMakes(d)[res] || !lineMakes(next)[res]) {
				t.Errorf("%s (%s) upgrades to %s, but not both make %s", d.Key, line, next.Key, res)
			}
		}
	}
	if offers == 0 {
		t.Fatal("no upgrade offers found")
	}
}

// TestLastRungOfALineHasNoUpgrade: the building whose line has no later rung
// is the last of its line: nothing is above it, in any age.
func TestLastRungOfALineHasNoUpgrade(t *testing.T) {
	defs := BaseBuildings()
	for _, key := range []string{"sawmill", "steam_coal_plant", "smart_refinery", "molecular_synthesizer", "black_market", "coal_plant", "ironmonger"} {
		var d BuildingDef
		for _, b := range defs {
			if b.Key == key {
				d = b
			}
		}
		if next, ok := NextRung(defs, d.Line(), d.LineageTier); ok {
			t.Errorf("%s is the last of line %s but %s is above it", key, d.Line(), next.Key)
		}
	}
	// A line that pauses still has its next rung: titanium after Modern.
	if next, ok := NextRung(defs, "metallurgy_titanium", 11); !ok || next.Key != "orbital_refinery" || next.RequiredAge != "space_age" {
		t.Errorf("nano_alloy_plant should be below orbital_refinery in the Space Age, got %v", next)
	}
}

// NextRung is the building one rung up the line from tier: the one with the
// lowest tier above it in the same line. A line that pauses (titanium after
// Modern, electricity while oil runs) has its next rung in a later age, which
// is still the next rung. ok is false when tier is the last of its line.
func NextRung(defs []BuildingDef, line string, tier int) (next BuildingDef, ok bool) {
	if line == "" {
		return BuildingDef{}, false
	}
	for _, b := range defs {
		if b.Line() != line || b.LineageTier <= tier {
			continue
		}
		if !ok || b.LineageTier < next.LineageTier {
			next, ok = b, true
		}
	}
	return next, ok
}
