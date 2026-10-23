package config

import (
	"math"
	"testing"
)

// TestStockScaleCoversEveryAge: every age has a scale, it never falls, and it
// grows from age to age until the Quantum Age, whose scale the Transcendent
// Age keeps (it has no storage building of its own).
func TestStockScaleCoversEveryAge(t *testing.T) {
	order := AgeOrder()
	prev := 0.0
	for i, age := range order {
		got := StockScale(age)
		if got <= 0 {
			t.Fatalf("%s: no stock scale", age)
		}
		if i < len(order)-1 && got <= prev {
			t.Errorf("%s: scale %v does not grow on %v", age, got, prev)
		}
		prev = got
	}
	if StockScale("transcendent_age") != StockScale("quantum_age") {
		t.Errorf("the Transcendent Age's scale %v is not the Quantum Age's %v", StockScale("transcendent_age"), StockScale("quantum_age"))
	}
	if StockScale("no_such_age") != 0 {
		t.Error("an unknown age has a scale")
	}
}

// TestStockOfFollowsTheIncome: a town making the age's typical income of a
// resource gets the share of the scale itself; one making less or more is
// sized on its income, within the band around the typical (no less than a
// quarter of it, no more than four times it); with no typical income to
// measure against, the share of the scale stands.
func TestStockOfFollowsTheIncome(t *testing.T) {
	const age, typical = "industrial_age", 1000.0
	scale := StockScale(age)
	near := func(got, want float64, what string) {
		t.Helper()
		if math.Abs(got-want) > 1e-9*want {
			t.Errorf("%s: %v, want %v", what, got, want)
		}
	}
	near(StockOf(age, 0.5, typical, typical), 0.5*scale, "the typical income")
	near(StockOf(age, 1, typical, 2*typical), 2*scale, "twice the typical income")
	near(StockOf(age, 1, typical, 100*typical), EventIncomeCeil*scale, "far over the typical income")
	near(StockOf(age, 1, typical, 0), EventIncomeFloor*scale, "no income")
	near(StockOf(age, 1, typical, -5), EventIncomeFloor*scale, "a negative income")
	near(StockOf(age, 0.8, 0, 123), 0.8*scale, "no typical income")
	// The minutes are what the share comes to at the typical income.
	minutes := StockMinutes(age, 1, typical)
	near(minutes*TicksPerMinute*typical, scale, "the minutes of the typical income")
	if StockMinutes(age, 1, 0) != 0 {
		t.Error("minutes of an income that is not there")
	}
}

// TestStakeMinutesHoldsTheShareToACeiling: a stake asks the minutes its share
// of the scale comes to at the age's typical income, and no more than the
// ceiling; with no typical income it asks nothing, ceiling or not.
func TestStakeMinutesHoldsTheShareToACeiling(t *testing.T) {
	const typical = 1000.0
	for _, age := range AgeOrder() {
		share := StockMinutes(age, 0.05, typical)
		if got := StakeMinutes(age, 0.05, typical, 1e300); got != share {
			t.Errorf("%s: below its ceiling the stake asks %v minutes, want %v", age, got, share)
		}
		if got := StakeMinutes(age, 0.05, typical, share/2); got != share/2 {
			t.Errorf("%s: over its ceiling the stake asks %v minutes, want %v", age, got, share/2)
		}
		if got := StakeMinutes(age, 0.05, 0, 6000); got != 0 {
			t.Errorf("%s: with no typical income the stake asks %v minutes, want 0", age, got)
		}
	}
}
