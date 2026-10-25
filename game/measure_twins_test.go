package game

import "testing"

// TestAccrueMeasure_Pins: faith and culture add a tick's worth of their rates
// to the run's measure the same way, with each product rounded on its own
// before it is added. The culture measure waits for the resource to be
// unlocked; the faith measure does not. The sum of own and other is the whole
// income only when the rate is above what the town's own buildings make.
func TestAccrueMeasure_Pins(t *testing.T) {
	// Variables, not constants: constants would be worked out exactly, not in float64.
	var own, moderate, rate, ticks float64 = 0.3, 0.45, 0.7, 0.4
	type twin struct {
		name    string
		res     string
		setRate func(ge *GameEngine)
		accrue  func(ge *GameEngine, ticks float64)
		measure func(ge *GameEngine) FaithSave
	}
	twins := []twin{
		{"faith", "faith",
			func(ge *GameEngine) { ge.faithRate = faithRates{own: own, moderate: moderate} },
			func(ge *GameEngine, n float64) { ge.accrueFaith(n) },
			func(ge *GameEngine) FaithSave { return ge.faithMeasure }},
		{"culture", "culture",
			func(ge *GameEngine) { ge.cultureRate = faithRates{own: own, moderate: moderate} },
			func(ge *GameEngine, n float64) { ge.accrueCulture(n) },
			func(ge *GameEngine) FaithSave { return ge.cultureMeasure }},
	}
	for _, tw := range twins {
		t.Run(tw.name, func(t *testing.T) {
			ge := NewGameEngine()
			ge.Resources.UnlockResource(tw.res)
			tw.setRate(ge)
			ge.Resources.resources[tw.res].Rate = rate

			var want FaithSave
			for i := 0; i < 3; i++ {
				tw.accrue(ge, ticks)
				want.Moderate += float64(moderate * ticks)
				want.Own += float64(own * ticks)
				want.Other += float64(float64(rate-own) * ticks)
				if got := tw.measure(ge); got != want {
					t.Fatalf("after %d accruals the measure is %+v, want %+v", i+1, got, want)
				}
			}

			// Income at or below what the town's own buildings make adds no "other".
			ge.Resources.resources[tw.res].Rate = own
			before := tw.measure(ge)
			tw.accrue(ge, ticks)
			after := tw.measure(ge)
			if after.Other != before.Other {
				t.Errorf("other grew from %v to %v with no income beyond the town's own", before.Other, after.Other)
			}
			if after.Own <= before.Own || after.Moderate <= before.Moderate {
				t.Errorf("own and moderate did not grow: %+v -> %+v", before, after)
			}

			// No ticks, no change.
			tw.accrue(ge, 0)
			if got := tw.measure(ge); got != after {
				t.Errorf("zero ticks changed the measure: %+v -> %+v", after, got)
			}
		})
	}

	// A culture resource not unlocked yet is not measured.
	ge := NewGameEngine()
	ge.cultureRate = faithRates{own: own, moderate: moderate}
	ge.Resources.resources["culture"].Rate = rate
	if ge.Resources.unlocked["culture"] {
		t.Fatal("setup: culture starts unlocked, so there is no locked case to test")
	}
	ge.accrueCulture(ticks)
	if ge.cultureMeasure != (FaithSave{}) {
		t.Errorf("a locked culture accrued %+v", ge.cultureMeasure)
	}
}
