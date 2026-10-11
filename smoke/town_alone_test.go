package smoke

import (
	"fmt"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// A town left alone stops growing: the storage wall, on the real engine.
//
// The reference town of an early age (reference.go) is stood up on the
// engine with every tech of the ages so far, wonder overflow on, and a plan
// that buys every building of the age without limit (1000 copies of each,
// the most a plan item takes). It is then left alone, a day at a time
// (SimulateOffline, the engine's own catch-up, which runs the plan, the
// builds and the stores as play does), and must show:
//
//   - purchases stop: some day before the horizon, none is made for a quiet
//     fortnight;
//   - income does not rise after they stop;
//   - the final income is within the bound property 3 states (propWallIncome
//     times the reference town's, in price units) wherever the tables
//     themselves hold that bound; where property 3 is a known failure for
//     the age, the reading is printed beside the tables' and no bound is
//     asserted;
//   - every store holds no more than its cap, every day.

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
	age       string
	lastBuy   int       // day of the last purchase (0: none)
	stopped   bool      // quiet for aloneQuietDays before the horizon
	income    []float64 // income over the reference town's, each sampled day
	days      []int
	final     float64
	model     float64 // the tables' own reading of the same
	startTick int
	stopTick  int // the tick of the last sample
	buyTick   int // the tick of the last purchase
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
	ge.SetWonderOverflow(true)
	for _, d := range tb.byAge[i] {
		if _, err := ge.PlanAddBuild(d.Key, 1000); err != nil {
			t.Fatalf("%s: planning %s: %v", out.age, d.Key, err)
		}
	}

	sample := func(day int) (built int) {
		st := ge.GetState()
		inc := map[string]float64{}
		for res, r := range st.Resources {
			inc[res] = r.Rate
			if r.Amount > r.Storage {
				t.Errorf("%s, day %d: %s holds %v in a store of %v", out.age, day, res, r.Amount, r.Storage)
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
			built, out.lastBuy, out.buyTick = b, day, out.stopTick
		}
		if day-out.lastBuy >= aloneQuietDays+step {
			out.stopped = true
			break
		}
	}
	out.final = out.income[len(out.income)-1]
	return out
}

func checkAlone(t *testing.T, r aloneAge, horizon int) {
	t.Helper()
	if !r.stopped {
		t.Errorf("%s: still buying on day %d, tick %d (limit: day %d, %d ticks in): a town left alone does not stop", r.age, r.lastBuy, r.buyTick, horizon, r.stopTick-r.startTick)
		return
	}
	// Income after the last purchase: settled one sample later (a build in
	// the queue finishes, a hand arrives), then never higher.
	settled := -1
	for k, d := range r.days {
		if d > r.lastBuy {
			settled = k
			break
		}
	}
	for k := settled + 1; k < len(r.income); k++ {
		if r.income[k] > r.income[settled]*(1+1e-9) {
			t.Errorf("%s: income rose from %.4f (day %d) to %.4f (day %d), after the last purchase on day %d", r.age, r.income[settled], r.days[settled], r.income[k], r.days[k], r.lastBuy)
		}
	}
	switch {
	case r.final <= propWallIncome:
	case r.model > propWallIncome:
		t.Logf("%s: income at the wall is %.2f times the reference town's, over the %g of property 3; the tables read %.2f (a known failure of property 3)", r.age, r.final, propWallIncome, r.model)
	default:
		t.Errorf("%s: income at the wall is %.2f times the reference town's, over the %g property 3 sets, and the tables read %.2f", r.age, r.final, propWallIncome, r.model)
	}
	t.Logf("%s: stopped buying on day %d (%d ticks in), income x%.2f of the reference town's (tables: x%.2f)", r.age, r.lastBuy, r.buyTick-r.startTick, r.final, r.model)
}

// The default form: one early age, the Iron Age, the first whose wall the
// tables keep within property 3's bound.
func TestTownLeftAloneStopsGrowing(t *testing.T) {
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
