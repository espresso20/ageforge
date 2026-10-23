package game

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// gateState is a hand-written snapshot of a town in age holding fraction f of
// every count the gate into its next age asks for, the wonder built when
// wonder is true.
func gateState(age string, f float64, wonder bool) GameState {
	set := rules.Core()
	st := GameState{Age: age, Buildings: map[string]BuildingState{}}
	next := set.Next(age)
	if a, ok := set.Age(next); ok {
		for key, need := range a.BuildingReqs {
			st.Buildings[key] = BuildingState{Count: int(float64(need) * f)}
		}
	}
	if w := set.Wonder(age); w != "" {
		bs := BuildingState{}
		if wonder {
			bs.Count = 1
		} else {
			// The lock the engine reports on a wonder whose keystone is not in.
			bs.NeedsTech = set.Keystone(age)
		}
		st.Buildings[w] = bs
	}
	return st
}

// Every row of every gate is in exactly one pillar, with the count the gate
// asks for, and the keystone and the wonder have rows of their own.
func TestAdvanceViewHoldsEveryGateRow(t *testing.T) {
	set := rules.Core()
	sawEmpty := map[string]bool{}
	for _, age := range set.AgeKeys() {
		next := set.Next(age)
		v := GameState{Age: age}.AdvanceView()
		if next == "" {
			if !v.Final || v.Pillars != nil || len(v.Blockers) == 0 {
				t.Errorf("%s is the last age: %+v", age, v)
			}
			continue
		}
		if len(v.Pillars) != 4 {
			t.Fatalf("%s: %d pillars", age, len(v.Pillars))
		}
		for i, key := range config.GatePillars() {
			if v.Pillars[i].Key != key || v.Pillars[i].Name != config.PillarName(key) {
				t.Errorf("%s: pillar %d is %s/%s, want %s", age, i, v.Pillars[i].Key, v.Pillars[i].Name, key)
			}
		}
		a, _ := set.Age(next)
		seen := map[string]int{}
		for _, p := range v.Pillars {
			for _, r := range p.Rows {
				if r.Kind != GateRowCount {
					continue
				}
				seen[r.Key]++
				if r.Need != a.BuildingReqs[r.Key] {
					t.Errorf("%s: %s needs %d, the gate asks %d", age, r.Key, r.Need, a.BuildingReqs[r.Key])
				}
				if def, ok := set.Building(r.Key); !ok || config.GatePillar(def) != p.Key {
					t.Errorf("%s: %s sits in %s", age, r.Key, p.Key)
				}
			}
			if len(p.Rows) == 0 {
				sawEmpty[age+"/"+p.Key] = true
				if !p.Met || p.Progress != 1 {
					t.Errorf("%s: the empty %s pillar is not met (%v, %v)", age, p.Key, p.Met, p.Progress)
				}
			}
		}
		for key := range a.BuildingReqs {
			if seen[key] != 1 {
				t.Errorf("%s: gate row %s is in %d pillars", age, key, seen[key])
			}
		}
		if len(seen) != len(a.BuildingReqs) {
			t.Errorf("%s: %d building rows, the gate asks for %d", age, len(seen), len(a.BuildingReqs))
		}
		// The wonder: a bank row and a build row; its keystone leads Knowledge.
		mon := v.Pillars[3]
		if wk := set.Wonder(age); wk != "" {
			if len(mon.Rows) != 2 || mon.Rows[0].Kind != GateRowShare || mon.Rows[1].Kind != GateRowDone || mon.Rows[1].Key != wk {
				t.Errorf("%s: the Monument pillar's rows %+v", age, mon.Rows)
			}
		}
		if tech := set.Keystone(age); tech != "" {
			kn := v.Pillars[0]
			if len(kn.Rows) == 0 || kn.Rows[0].Key != tech || kn.Rows[0].Kind != GateRowDone {
				t.Errorf("%s: Knowledge does not lead with its keystone %s: %+v", age, tech, kn.Rows)
			}
		}
	}
	t.Logf("pillars with nothing needed: %v", sortedBool(sawEmpty))
}

func sortedBool(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// All four pillars standing is CheckAdvancement saying ready, in every age:
// with the gate met exactly, with each count one short, with the wonder
// missing, and with nothing at all.
func TestAdvancePillarsAgreeWithCheckAdvancement(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	ge := NewGameEngine()
	ge.SeedRNG(1)
	ages := config.Ages()
	for g := 0; g+1 < len(ages); g++ {
		from, to := ages[g], ages[g+1]
		_ = ge.ForceQuietFateForTest(from.EpochKey)
		wonder := ge.progress.WonderForAge(from.Key)

		// notReady counts the checks that left the gate shut, so a gate that
		// nothing closes does not pass for want of looking.
		notReady := 0
		check := func(label string, counts map[string]int, want *bool) {
			t.Helper()
			ge.SetTownForTest(counts, 0)
			engineReady := ge.progress.CheckAdvancement(ge.age, ge.Buildings) != ""
			v := ge.GetState().AdvanceView()
			if want != nil && engineReady != *want {
				t.Fatalf("%s, %s: CheckAdvancement ready = %v, want %v", from.Key, label, engineReady, *want)
			}
			if !engineReady {
				notReady++
			}
			if v.AllStand != engineReady {
				var rows []string
				for _, p := range v.Pillars {
					for _, r := range p.Rows {
						rows = append(rows, fmt.Sprintf("%s/%s %d/%d met=%v", p.Key, r.Key, r.Have, r.Need, r.Met))
					}
				}
				t.Errorf("%s, %s: pillars all stand = %v, CheckAdvancement ready = %v\n%s", from.Key, label, v.AllStand, engineReady, strings.Join(rows, "\n"))
			}
			if v.Ready != (engineReady && v.Blockers == nil) {
				t.Errorf("%s, %s: Ready = %v with all standing %v and blockers %v", from.Key, label, v.Ready, v.AllStand, v.Blockers)
			}
		}
		yes, no := true, false

		full := map[string]int{}
		for k, n := range to.BuildingReqs {
			full[k] = n
		}
		if wonder != "" {
			full[wonder] = 1
		}
		zero := map[string]int{}
		for k := range full {
			zero[k] = 0
		}
		check("nothing built", zero, &no)
		check("the gate met exactly", full, &yes)

		keys := make([]string, 0, len(full))
		for k := range full {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			short := map[string]int{}
			for kk, n := range full {
				short[kk] = n
			}
			short[k]--
			check("one short of "+k, short, nil) // a later rung of the same line can stand in
		}
		// The wonder missing is covered above (it is in keys when there is
		// one). Over the gate is still ready.
		over := map[string]int{}
		for k, n := range full {
			over[k] = n + 3
		}
		if wonder != "" {
			over[wonder] = 1
		}
		check("over the gate", over, &yes)
		if notReady < 2 {
			t.Errorf("%s: only %d of the checks left the gate shut", from.Key, notReady)
		}

		if err := ge.EnterAgeForTest(to.Key); err != nil {
			t.Fatal(err)
		}
	}
}

// A copy upgraded into a later rung of the line counts toward the gate, in the
// pillars as in CheckAdvancement.
func TestAdvanceViewCountsLaterRungs(t *testing.T) {
	st := gateState("primitive_age", 1, true)
	key := ""
	for k := range rules.Core().Ages()[1].BuildingReqs {
		key = k
		break
	}
	bs := st.Buildings[key]
	bs.HeldWithLater, bs.Count = bs.Count, 0
	st.Buildings[key] = bs
	for _, p := range st.AdvanceView().Pillars {
		for _, r := range p.Rows {
			if r.Key == key && !r.Met {
				t.Errorf("%s held as a later rung is not counted: %+v", key, r)
			}
		}
	}
}

// The year is the age's start at nothing and the next age's start at all of
// it, and never falls as more of the gate is met.
func TestAdvanceYearFollowsTheGate(t *testing.T) {
	set := rules.Core()
	keys := set.AgeKeys()
	years := config.AgeStartYears(keys)
	for i, age := range keys {
		if set.Next(age) == "" {
			continue
		}
		v0 := gateState(age, 0, false).AdvanceView()
		if v0.Year != years[i] {
			t.Errorf("%s at nothing: year %d, want %d", age, v0.Year, years[i])
		}
		v1 := gateState(age, 1, true).AdvanceView()
		if v1.Progress < 1 && v1.Progress > 0.999 {
			t.Errorf("%s: %v", age, v1.Progress)
		}
		// Bank and keystone are the gate's too: with the wonder standing the
		// pillars are all met, so progress is 1 and the year the next start.
		if !v1.AllStand || v1.Progress != 1 || v1.Year != years[i+1] {
			t.Errorf("%s at the whole gate: stand=%v progress=%v year %d, want %d", age, v1.AllStand, v1.Progress, v1.Year, years[i+1])
		}
		prev := v0.Year
		prevP := -1.0
		for f := 0.0; f <= 1.0001; f += 0.05 {
			v := gateState(age, min(f, 1), f >= 1).AdvanceView()
			if v.Progress < prevP {
				t.Fatalf("%s: progress fell from %v to %v at %.2f", age, prevP, v.Progress, f)
			}
			if v.Year < prev {
				t.Fatalf("%s: the year fell from %d to %d at %.2f", age, prev, v.Year, f)
			}
			prev, prevP = v.Year, v.Progress
		}
	}
	// The last age shows its word, not a number.
	last := keys[len(keys)-1]
	v := GameState{Age: last}.AdvanceView()
	if v.YearText != "Beyond time" || v.March[len(v.March)-1].Year != "Beyond time" {
		t.Errorf("the last age reads %q / %q", v.YearText, v.March[len(v.March)-1].Year)
	}
}

// The calendar is held at the furthest point shown, within an age, and starts
// over in a new age, a new run and a loaded game.
func TestHoldCalendarNeverRunsBackward(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	ge := NewGameEngine()
	ge.SeedRNG(1)

	high := gateState("primitive_age", 0.8, false)
	high.Tick = 100
	low := gateState("primitive_age", 0.2, false)
	low.Tick = 101

	a := ge.HoldCalendar(high.AdvanceView())
	b := ge.HoldCalendar(low.AdvanceView())
	if b.Year != a.Year || b.Share != a.Share {
		t.Errorf("the calendar went back: %d (%.3f) then %d (%.3f)", a.Year, a.Share, b.Year, b.Share)
	}
	if b.Progress >= a.Progress {
		t.Errorf("the pillars should show the truth: progress %.3f then %.3f", a.Progress, b.Progress)
	}
	// Further on moves it on.
	best := gateState("primitive_age", 1, true)
	best.Tick = 102
	if c := ge.HoldCalendar(best.AdvanceView()); c.Year <= a.Year {
		t.Errorf("the calendar did not move on: %d then %d", a.Year, c.Year)
	}
	// An earlier tick (a new run started over) forgets the mark.
	again := gateState("primitive_age", 0.1, false)
	again.Tick = 3
	if d := ge.HoldCalendar(again.AdvanceView()); d.Share > 0.2 {
		t.Errorf("a new run kept the old mark: %.3f", d.Share)
	}
	// So does a new age, by name, and the real advance.
	ge.HoldCalendar(high.AdvanceView())
	stone := gateState("stone_age", 0.1, false)
	stone.Tick = 200
	if e := ge.HoldCalendar(stone.AdvanceView()); e.Share > 0.2 {
		t.Errorf("a new age kept the old mark: %.3f", e.Share)
	}
	ge.HoldCalendar(high.AdvanceView())
	if err := ge.EnterAgeForTest("stone_age"); err != nil {
		t.Fatal(err)
	}
	if ge.calendar != (calendarMark{}) {
		t.Errorf("advancing kept the mark: %+v", ge.calendar)
	}
	// The final age has nothing to hold.
	var nilEngine *GameEngine
	v := GameState{Age: "transcendent_age"}.AdvanceView()
	if got := ge.HoldCalendar(v); got.Year != v.Year {
		t.Errorf("the final age moved: %d to %d", v.Year, got.Year)
	}
	if got := nilEngine.HoldCalendar(high.AdvanceView()); got.Year != high.AdvanceView().Year {
		t.Error("a nil engine changed the view")
	}
}

// The march names the ages the player may see and leaves the rest unnamed.
func TestAdvanceMarchKeepsLaterAgesDark(t *testing.T) {
	set := rules.Core()
	v := GameState{Age: "primitive_age"}.AdvanceView()
	for i, s := range v.March {
		named := i <= 1
		if (s.Name != "") != named || (s.Year != "") != named {
			t.Errorf("stop %d (%s) reads %q %q: named should be %v", i, s.Key, s.Name, s.Year, named)
		}
	}
	if v.Here != 0 || v.March[0].Key != "primitive_age" || len(v.March) != set.NumAges() {
		t.Errorf("here %d, %d stops", v.Here, len(v.March))
	}
	// An age reached before, on the account, may be named.
	st := GameState{Age: "primitive_age", AccountStats: &AccountStatsView{HighestAge: "bronze_age"}}
	w := st.AdvanceView()
	if w.March[2].Name == "" || w.March[3].Name != "" {
		t.Errorf("with the Bronze Age reached on the account: %q %q", w.March[2].Name, w.March[3].Name)
	}
}

// Anything but the pillars that stops the advance is listed: a catastrophe
// pending is a blocker, an open harbinger's word is a notice.
func TestAdvanceViewBlockersAndNotices(t *testing.T) {
	st := gateState("bronze_age", 1, true)
	st.Research.Techs = map[string]TechState{}
	if v := st.AdvanceView(); !v.Ready || len(v.Blockers) != 0 {
		t.Fatalf("a met gate with nothing in the way: ready %v blockers %v", v.Ready, v.Blockers)
	}
	st.PendingCatastrophe = "iron_era"
	v := st.AdvanceView()
	if v.Ready || !v.AllStand || len(v.Blockers) != 1 || !strings.Contains(v.Blockers[0], "'catastrophe'") {
		t.Errorf("a pending catastrophe: ready %v stand %v blockers %v", v.Ready, v.AllStand, v.Blockers)
	}
	st.PendingCatastrophe = ""
	st.Harbinger = &HarbingerView{Name: "The Seer", When: WhenThisAge}
	v = st.AdvanceView()
	if !v.Ready || len(v.Notices) != 1 || !strings.Contains(v.Notices[0], "The Seer") || !strings.Contains(v.Notices[0], "'harbinger'") {
		t.Errorf("a harbinger's word: ready %v notices %v", v.Ready, v.Notices)
	}
	// A doom promised for the era stands in the way of an advance only when
	// the advance leaves the era.
	set := rules.Core()
	for _, age := range set.AgeKeys() {
		next := set.Next(age)
		if next == "" {
			continue
		}
		g := gateState(age, 1, true)
		g.Harbinger = &HarbingerView{Name: "The Seer", When: WhenThisEra}
		leaves := set.EraOf(next) != set.EraOf(age)
		if got := len(g.AdvanceView().Notices) == 1; got != leaves {
			t.Errorf("%s to %s, an era's doom: notice %v, advance leaves the era %v", age, next, got, leaves)
		}
	}
}

// What the next age brings reads from the rules, in the age's own words for
// people, and counts the wonder apart.
func TestAdvanceBrings(t *testing.T) {
	v := gateState("bronze_age", 0, false).AdvanceView()
	b := v.Brings
	if b.Buildings == 0 || b.Wonder == "" || len(b.People) != 4 {
		t.Fatalf("brings %+v", b)
	}
	if want := capFirst(config.Term("iron_age", config.TermScholars).Many); b.People[3] != want {
		t.Errorf("the people of the Iron Age: %v, want ...%s", b.People, want)
	}
	for _, age := range rules.Core().Ages() {
		unlocked := 0
		for _, k := range age.UnlockBuildings {
			if def, ok := rules.Core().Building(k); ok && def.Category != "wonder" {
				unlocked++
			}
		}
		prev := rules.Core().Ages()[max(0, age.Order-1)]
		if age.Order == 0 {
			continue
		}
		got := gateState(prev.Key, 0, false).AdvanceView().Brings.Buildings
		if got != unlocked {
			t.Errorf("%s brings %d kinds of building, the table has %d", age.Key, got, unlocked)
		}
	}
}

// The keystone row: not started, in progress, queued, researched, and not
// needed once the wonder stands.
func TestAdvanceKeystoneRow(t *testing.T) {
	set := rules.Core()
	age := "stone_age"
	tech := set.Keystone(age)
	if tech == "" {
		t.Fatal("the Stone Age has no keystone")
	}
	wk := set.Wonder(age)
	row := func(st GameState) GateRow {
		t.Helper()
		return st.AdvanceView().Pillars[0].Rows[0]
	}
	st := gateState(age, 0, false)
	st.Buildings[wk] = BuildingState{NeedsTech: tech}
	st.Research.Techs = map[string]TechState{tech: {}}
	if r := row(st); r.Met || r.Note != "not started" || r.Progress != 0 {
		t.Errorf("not started: %+v", r)
	}
	st.Research.Queue = []string{tech}
	if r := row(st); r.Note != "queued" {
		t.Errorf("queued: %+v", r)
	}
	st.Research.CurrentTech, st.Research.Progress, st.Research.Price = tech, 25, 100
	if r := row(st); r.Met || r.Progress != 0.25 || r.Note != "researching, 25%" {
		t.Errorf("in progress: %+v", r)
	}
	st.Research.Techs = map[string]TechState{tech: {Researched: true}}
	if r := row(st); !r.Met || r.Progress != 1 || r.Note != "researched" {
		t.Errorf("researched: %+v", r)
	}
	st.Research.Techs = map[string]TechState{tech: {}}
	st.Buildings[wk] = BuildingState{Count: 1}
	if r := row(st); !r.Met || r.Note != "not needed" {
		t.Errorf("the wonder stands: %+v", r)
	}
	// And the Monument pillar says what the wonder waits for.
	st = gateState(age, 0, false)
	st.Buildings[wk] = BuildingState{NeedsTech: tech}
	if m := st.AdvanceView().Pillars[3].Rows; m[1].Note != "needs "+set.Name(rules.KindTech, tech)+" first" {
		t.Errorf("the wonder row reads %q", m[1].Note)
	}
	def, _ := set.Building(wk)
	full := map[string]float64{}
	for r, c := range def.BaseCost {
		full[r] = c
	}
	st.Buildings[wk] = BuildingState{WonderBank: full, WonderBankFull: true}
	m := st.AdvanceView().Pillars[3].Rows
	if !m[0].Met || m[0].Have != 100 || m[1].Met || m[1].Progress != 0 {
		t.Errorf("a full bank, the wonder not begun: %+v", m)
	}
	st.BuildQueue = []BuildQueueSnapshot{{Name: def.Name, TicksLeft: 30, TotalTicks: 100}}
	if m := st.AdvanceView().Pillars[3].Rows; m[1].Progress != 0.7 || m[1].Note != "building, 70%" {
		t.Errorf("the wonder under construction: %+v", m[1])
	}
}
