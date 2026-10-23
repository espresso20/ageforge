package smoke

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// A town left alone slows down: prices climb with every copy, so its income
// grows more slowly the longer it is left.
//
// The reference town of an early age (reference.go) is stood up on the
// engine with every tech of the ages so far, and a plan
// that buys every building of the age without limit (1000 copies of each,
// the most a plan item takes). It is then left alone, a day at a time
// (SimulateOffline, the engine's own catch-up, which runs the plan, the
// builds and the stores as play does), and must slow down: the income it gains in the last third
// of the samples is at most half of what it gained in the first third. Every
// store holds a finite amount, never negative, every day.

// A day is 43200 ticks here (a tick is two seconds). The Iron Age stops on
// day 8, so the default form allows a month (1,296,000 ticks); the long form
// allows 1000 days, for a Colonial Age that takes about 750.
const (
	aloneQuietDays    = 14   // days without a purchase that count as stopped
	aloneShortHorizon = 30   // days the default form may take to stop
	aloneLongHorizon  = 1000 // days the long form may take to stop
)

// aloneAge is one age left alone.
type aloneAge struct {
	age          string
	income       []float64 // income over the reference town's, each sampled day
	days         []int
	final        float64
	model        float64 // the tables' own reading of the same
	startTick    int
	stopTick     int // the tick of the last sample
	buyTick      int // the tick of the last purchase
	purchaseDays []int
}

// leaveTownAlone stands age i's reference town on the engine and leaves it
// for up to horizon days, sampling every step days.
func leaveTownAlone(t *testing.T, tb *refTables, i, horizon, step int) aloneAge {
	t.Helper()
	restore := game.SetDataDirForTest(t.TempDir())
	defer restore()
	town := tb.refTown(i)
	ref := tb.state(town, i, 1)
	refIncome := units(ref.income, tb.levels[i])
	wall := tb.stay(i, town, refPlay{rate: 1, speed: 1, refIncome: refIncome, horizon: 2e7})
	out := aloneAge{age: tb.ages[i].Name, model: units(wall.end.income, tb.levels[i]) / refIncome}

	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for j := 1; j <= i; j++ {
		if err := ge.EnterAgeForTest(tb.ages[j].Key); err != nil {
			t.Fatalf("%s: entering the %s: %v", out.age, tb.ages[j].Name, err)
		}
	}
	_ = ge.ForceQuietFateForTest(tb.ages[i].EpochKey)
	ge.GrantTechsForTest()
	ge.SetTownForTest(town, int(ref.workers))
	for _, d := range tb.byAge[i] {
		n := 1000
		if d.MaxCount > 0 {
			// A knowledge building stops at the count its gate asks for, and
			// the reference town holds exactly that.
			if town[d.Key] >= d.MaxCount {
				continue
			}
			n = d.MaxCount - town[d.Key]
		}
		if _, err := ge.PlanAddBuild(d.Key, n); err != nil {
			t.Fatalf("%s: planning %s: %v", out.age, d.Key, err)
		}
	}

	sample := func(day int) (built int) {
		st := ge.GetState()
		inc := map[string]float64{}
		for res, r := range st.Resources {
			inc[res] = r.Rate
			if math.IsNaN(r.Amount) || math.IsInf(r.Amount, 0) || r.Amount < 0 {
				t.Errorf("%s, day %d: %s holds %v", out.age, day, res, r.Amount)
			}
		}
		for _, d := range tb.byAge[i] {
			built += st.Buildings[d.Key].Count
		}
		built += len(st.BuildQueue)
		out.income = append(out.income, units(inc, tb.levels[i])/refIncome)
		out.days = append(out.days, day)
		out.stopTick = st.Tick
		return built
	}
	out.startTick = ge.GetState().Tick
	out.buyTick = out.startTick
	built := sample(0)
	for day := step; day <= horizon; day += step {
		for d := 0; d < step; d++ {
			ge.SimulateOffline(24 * time.Hour) // the most one catch-up credits (MaxOfflineTime)
		}
		if b := sample(day); b != built {
			built = b
		}
	}
	out.final = out.income[len(out.income)-1]
	return out
}

func checkAlone(t *testing.T, r aloneAge, horizon int) {
	t.Helper()
	if len(r.income) < 8 {
		t.Errorf("%s: only %d income samples; need at least eight", r.age, len(r.income))
		return
	}
	n := len(r.income)
	for _, v := range r.income {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("%s: income is not finite (%v)", r.age, r.income)
			return
		}
	}
	// A town left alone slows down: what its income grows in the last third
	// of the samples is at most half of what it grew in the first third.
	first := r.income[n/3] - r.income[0]
	lastThird := r.income[n-1] - r.income[n-1-n/3]
	t.Logf("%s: income %.4f at the start, %.4f a third in, %.4f two thirds in, %.4f at the end (grew %.4f in the first third, %.4f in the last)",
		r.age, r.income[0], r.income[n/3], r.income[n-1-n/3], r.income[n-1], first, lastThird)
	if lastThird > first*0.5 {
		t.Errorf("%s: income grew %.4f in the last third of the samples and %.4f in the first (want the last at most half the first)", r.age, lastThird, first)
	}
}

// The default form: one early age, the Iron Age, the first whose wall the
// tables keep within property 3's bound.
func TestTownLeftAloneSlowsDown(t *testing.T) {
	tb := newRefTables()
	if testing.Short() {
		t.Skip("plays a month of an age on the engine")
	}
	i := tb.idx["iron_age"]
	r := leaveTownAlone(t, tb, i, aloneShortHorizon, 1)
	checkAlone(t, r, aloneShortHorizon)
}

// The long form, over the first eight ages. AGEFORGE_LONG_TESTS=1.
func TestTownLeftAloneStopsGrowing_Ages(t *testing.T) {
	if !longTests() {
		t.Skip("set AGEFORGE_LONG_TESTS=1 for the town left alone in each of the first eight ages")
	}
	tb := newRefTables()
	for i := 0; i < 8; i++ {
		t.Run(fmt.Sprintf("%02d_%s", i, tb.ages[i].Key), func(t *testing.T) {
			r := leaveTownAlone(t, tb, i, aloneLongHorizon, 10)
			checkAlone(t, r, aloneLongHorizon)
		})
	}
}
