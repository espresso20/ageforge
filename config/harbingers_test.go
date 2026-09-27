package config

import (
	"regexp"
	"strings"
	"testing"
)

// TestHarbingerRosterCoversEveryAge pins the roster to config's age list: one
// harbinger per age, in age order, with unique keys, and a lookup that agrees
// with the list.
func TestHarbingerRosterCoversEveryAge(t *testing.T) {
	order := AgeOrder()
	roster := Harbingers()
	if len(order) != 22 {
		t.Fatalf("AgeOrder has %d ages; the roster was written for 22", len(order))
	}
	if len(roster) != len(order) {
		t.Fatalf("roster has %d harbingers for %d ages", len(roster), len(order))
	}
	keys := map[string]bool{}
	names := map[string]bool{}
	snake := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	for i, h := range roster {
		if h.Age != order[i] {
			t.Errorf("roster[%d] is for %q, want %q (age order)", i, h.Age, order[i])
		}
		if !snake.MatchString(h.Key) {
			t.Errorf("%s: key %q is not snake_case", h.Age, h.Key)
		}
		if keys[h.Key] {
			t.Errorf("%s: duplicate key %q", h.Age, h.Key)
		}
		keys[h.Key] = true
		if names[h.Name] {
			t.Errorf("%s: duplicate name %q", h.Age, h.Name)
		}
		names[h.Name] = true
		got, ok := HarbingerFor(h.Age)
		if !ok || got != h {
			t.Errorf("HarbingerFor(%q) = %+v, %v; want the roster entry", h.Age, got, ok)
		}
	}
	if _, ok := HarbingerFor("not_an_age"); ok {
		t.Error("HarbingerFor accepted an unknown age")
	}
}

// TestHarbingerLabels holds every player-facing string to the same rules: set,
// trimmed, no markup (the UI wraps them in colour tags), and the three action
// labels distinct within an age so the player is never offered two buttons with
// the same words. Names open lowercase because the flavor package drops them
// mid-sentence ("Half the settlement came out to get a look at the Oracle").
func TestHarbingerLabels(t *testing.T) {
	for _, h := range Harbingers() {
		fields := map[string]string{
			"Name": h.Name, "Description": h.Description,
			"AppeaseLabel": h.AppeaseLabel, "BraceLabel": h.BraceLabel, "InviteLabel": h.InviteLabel,
		}
		for field, v := range fields {
			switch {
			case strings.TrimSpace(v) == "":
				t.Errorf("%s: %s is empty", h.Age, field)
			case v != strings.TrimSpace(v):
				t.Errorf("%s: %s has edge whitespace: %q", h.Age, field, v)
			case strings.ContainsAny(v, "[]%{}"):
				t.Errorf("%s: %s carries markup: %q", h.Age, field, v)
			}
		}
		if first := h.Name[:1]; first != strings.ToLower(first) {
			t.Errorf("%s: Name %q must open lowercase so it reads mid-sentence", h.Age, h.Name)
		}
		labels := []string{h.AppeaseLabel, h.BraceLabel, h.InviteLabel}
		seen := map[string]bool{}
		for _, l := range labels {
			low := strings.ToLower(l)
			if seen[low] {
				t.Errorf("%s: action label %q is used twice", h.Age, l)
			}
			seen[low] = true
		}
	}
}

// TestHarbingerForecastAndFalseProphets pins the two derived curves: precision
// flips from vague to numeric exactly at the Industrial Age, and the
// false-prophet chance starts at 1/8, never rises, and is exactly zero from the
// Industrial Age on, so the ages that publish odds never lie.
func TestHarbingerForecastAndFalseProphets(t *testing.T) {
	industrial := -1
	for i, k := range AgeOrder() {
		if k == "industrial_age" {
			industrial = i
		}
	}
	if industrial < 0 {
		t.Fatal("no industrial_age in AgeOrder")
	}
	roster := Harbingers()
	if roster[0].FalseProphetChance != 0.125 {
		t.Errorf("primitive FalseProphetChance = %v, want 0.125", roster[0].FalseProphetChance)
	}
	prev := 1.0
	for i, h := range roster {
		wantPrecision := ForecastVague
		if i >= industrial {
			wantPrecision = ForecastNumeric
		}
		if h.ForecastPrecision != wantPrecision {
			t.Errorf("%s: ForecastPrecision = %v, want %v", h.Age, h.ForecastPrecision, wantPrecision)
		}
		if h.FalseProphetChance > prev {
			t.Errorf("%s: FalseProphetChance %v rises from %v", h.Age, h.FalseProphetChance, prev)
		}
		prev = h.FalseProphetChance
		if h.FalseProphetChance < 0 || h.FalseProphetChance > 0.125 {
			t.Errorf("%s: FalseProphetChance %v is outside [0, 0.125]", h.Age, h.FalseProphetChance)
		}
		switch {
		case i >= industrial && h.FalseProphetChance != 0:
			t.Errorf("%s: FalseProphetChance = %v; must be exactly 0 from the Industrial Age on", h.Age, h.FalseProphetChance)
		case i < industrial && h.FalseProphetChance == 0:
			t.Errorf("%s: FalseProphetChance is 0 before the Industrial Age", h.Age)
		}
	}
}

// TestHarbingersIsACopy guards the build-once table from callers that edit the
// slice they were handed.
func TestHarbingersIsACopy(t *testing.T) {
	a := Harbingers()
	a[0].Name = "mutated"
	if b := Harbingers(); b[0].Name == "mutated" {
		t.Fatal("Harbingers returned the shared table, not a copy")
	}
}
