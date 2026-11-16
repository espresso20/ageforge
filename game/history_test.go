package game

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

func hv(level, count float64) map[string]float64 {
	return map[string]float64{"level": level, "buildings": count}
}

// The run tier never holds more than historyRunMax buckets: at the 513th the
// neighbours merge in pairs and the span doubles. A constant stays constant, a
// ramp keeps its overall mean, and a count column ends each bucket on its last
// value.
func TestHistoryRunTierMerges(t *testing.T) {
	h := NewRunHistory()
	for i := 1; i <= 512; i++ {
		h.Record(10*i, hv(float64(i), float64(i)))
	}
	if h.Run.N != 512 || h.Run.Span != 10 {
		t.Fatalf("512 samples: %d buckets of %d ticks, want 512 of 10", h.Run.N, h.Run.Span)
	}
	h.Record(5130, hv(513, 513))
	if h.Run.N != 257 || h.Run.Span != 20 {
		t.Fatalf("513 samples: %d buckets of %d ticks, want 257 of 20", h.Run.N, h.Run.Span)
	}
	for i := 514; i <= 20000; i++ {
		h.Record(10*i, hv(float64(i), float64(i)))
		if h.Run.N > historyRunMax {
			t.Fatalf("sample %d: %d buckets", i, h.Run.N)
		}
	}
	if h.Recent.N != historyRecentLen {
		t.Errorf("the recent ring holds %d buckets, want %d", h.Recent.N, historyRecentLen)
	}
	// 20000 samples: spans of 10 doubled until 20000/span <= 512: span 640
	// (31.25 buckets of 640 ticks... 200000/640 = 312.5), so 313 buckets.
	if h.Run.Span != 640 || h.Run.N != 313 {
		t.Errorf("20000 samples: %d buckets of %d ticks, want 313 of 640", h.Run.N, h.Run.Span)
	}
	last := h.Run.Cols["buildings"]
	if last[h.Run.N-1] != 20000 {
		t.Errorf("a count column ends on %v, want its last value 20000", last[h.Run.N-1])
	}
	for j := 0; j < h.Run.N-1; j++ {
		if want := float32((j + 1) * h.Run.Span / historyInterval); last[j] != want {
			t.Fatalf("count bucket %d is %v, want its last sample %v", j, last[j], want)
		}
	}
}

func TestHistoryRunTierMeans(t *testing.T) {
	h := NewRunHistory()
	const n = 2048 // 512 full buckets of 40 ticks
	var sum float64
	for i := 1; i <= n; i++ {
		h.Record(10*i, map[string]float64{"ramp": float64(i), "flat": 3.5})
		sum += float64(i)
	}
	if h.Run.N != 512 || h.Run.Span != 40 || h.Run.Fill != 4 {
		t.Fatalf("%d buckets of %d ticks, fill %d", h.Run.N, h.Run.Span, h.Run.Fill)
	}
	var got float64
	for _, v := range h.Run.Cols["ramp"] {
		got += float64(v)
	}
	got /= float64(h.Run.N)
	if want := sum / n; math.Abs(got-want) > 1e-3*want {
		t.Errorf("the ramp's mean is %v after merging, want %v", got, want)
	}
	for i, v := range h.Run.Cols["flat"] {
		if v != 3.5 {
			t.Fatalf("a constant series reads %v in bucket %d", v, i)
		}
	}
}

// A column that appears late reads zero before its first bucket, in both tiers.
func TestHistoryLateColumn(t *testing.T) {
	h := NewRunHistory()
	for i := 1; i <= 5; i++ {
		h.Record(10*i, map[string]float64{"early": 1})
	}
	for i := 6; i <= 8; i++ {
		h.Record(10*i, map[string]float64{"early": 1, "late": 7})
	}
	for _, tier := range []*HistoryTier{&h.Recent, &h.Run} {
		c := tier.Cols["late"]
		if len(c) != 8 || !slices.Equal(c, []float32{0, 0, 0, 0, 0, 7, 7, 7}) {
			t.Errorf("late column = %v", c)
		}
	}
}

// The recent ring keeps the last historyRecentLen samples, one per interval.
func TestHistoryRecentRing(t *testing.T) {
	h := NewRunHistory()
	for i := 1; i <= 400; i++ {
		h.Record(10*i, map[string]float64{"v": float64(i)})
	}
	r := &h.Recent
	if r.Len() != historyRecentLen || r.TickAt(0) != 1010 || r.TickAt(r.Len()-1) != 4000 {
		t.Fatalf("ring: %d buckets, ticks %d - %d, want %d, 1010 - 4000", r.Len(), r.TickAt(0), r.TickAt(r.Len()-1), historyRecentLen)
	}
	for i, v := range r.Cols["v"] {
		if int(v) != 101+i {
			t.Fatalf("ring slot %d holds %v, want %d", i, v, 101+i)
		}
	}
	// The run tier's ticks line up with its buckets after the merges.
	if tk := h.Run.TickAt(h.Run.N - 1); tk != 4000 {
		t.Errorf("run tier ends at tick %d, want 4000", tk)
	}
}

func TestHistoryMarksCap(t *testing.T) {
	h := NewRunHistory()
	h.Mark(1, MarkAge, "a", 0)
	h.Mark(2, MarkAway, "", 5)
	for i := 0; i < historyMarksMax; i++ {
		h.Mark(10+i, MarkEvent, fmt.Sprint(i), 0)
	}
	if len(h.Marks) != historyMarksMax {
		t.Fatalf("%d marks, want %d", len(h.Marks), historyMarksMax)
	}
	if !slices.ContainsFunc(h.Marks, func(m HistoryMark) bool { return m.Kind == MarkAge }) ||
		!slices.ContainsFunc(h.Marks, func(m HistoryMark) bool { return m.Kind == MarkAway }) {
		t.Fatal("the age or away mark went while events could go")
	}
	if h.Marks[2].Key != "2" {
		t.Errorf("the oldest events should go first, the oldest left is %q", h.Marks[2].Key)
	}
	// With no events left, away marks go next; ages and the rest stay.
	h = NewRunHistory()
	for i := 0; i < 40; i++ {
		h.Mark(i, MarkAge, "x", 0)
	}
	h.Mark(100, MarkAway, "", 1)
	for i := 0; i < historyMarksMax-41; i++ {
		h.Mark(200+i, MarkWonder, "w", 0)
	}
	if len(h.Marks) != historyMarksMax {
		t.Fatalf("%d marks before the cap", len(h.Marks))
	}
	if !slices.ContainsFunc(h.Marks, func(m HistoryMark) bool { return m.Kind == MarkAway }) {
		t.Fatal("the away mark went before the cap was passed")
	}
	h.Mark(3000, MarkWonder, "w", 0)
	if len(h.Marks) != historyMarksMax || slices.ContainsFunc(h.Marks, func(m HistoryMark) bool { return m.Kind == MarkAway }) {
		t.Errorf("%d marks, away mark still there: %v", len(h.Marks), slices.ContainsFunc(h.Marks, func(m HistoryMark) bool { return m.Kind == MarkAway }))
	}
	for _, k := range []string{MarkAge, MarkWonder, MarkCatastrophe, MarkHarbinger} {
		h = NewRunHistory()
		for i := 0; i < historyMarksMax+3; i++ {
			h.Mark(i, k, "k", 0)
		}
		if len(h.Marks) != historyMarksMax+3 {
			t.Errorf("%s marks were dropped: %d left", k, len(h.Marks))
		}
	}
}

func markCount(ge *GameEngine, kind string) int {
	n := 0
	for _, m := range ge.history.Marks {
		if m.Kind == kind {
			n++
		}
	}
	return n
}

// An age advance, a completed wonder, a catastrophe, a harbinger and an epoch
// event each leave one mark.
func TestHistoryMarksFromTheEngine(t *testing.T) {
	ge := catEngine(t, "bronze_age", 5)
	ge.StepTicks(50)
	ge.advanceAge("iron_age")
	if markCount(ge, MarkAge) != 1 || !slices.Contains(ge.history.Marks, HistoryMark{Tick: ge.tick, Kind: MarkAge, Key: "iron_age"}) {
		t.Errorf("age marks: %+v", ge.history.Marks)
	}

	wonder := ge.progress.WonderForAge(ge.age)
	ge.noteBuilt(wonder, 1)
	if markCount(ge, MarkWonder) != 1 || !slices.Contains(ge.history.Marks, HistoryMark{Tick: ge.tick, Kind: MarkWonder, Key: wonder}) {
		t.Errorf("wonder marks: %+v", ge.history.Marks)
	}
	ge.noteBuilt("hut", 1)
	if markCount(ge, MarkWonder) != 1 {
		t.Errorf("an ordinary building left a wonder mark")
	}

	events := markCount(ge, MarkEvent)
	ge.recordEpochEvent(ge.currentEpoch, config.EpochEventDef{Key: "test_event", Name: "Test", Type: "good_minor"})
	if markCount(ge, MarkEvent) != events+1 || markCount(ge, MarkCatastrophe) != 0 {
		t.Errorf("event marks: %+v", ge.history.Marks)
	}

	ge.pendingCatastrophe = ""
	if err := ge.ForceCatastropheForTest(); err != nil {
		t.Fatal(err)
	}
	if markCount(ge, MarkCatastrophe) != 1 || markCount(ge, MarkEvent) != events+1 {
		t.Errorf("catastrophe marks: %+v", ge.history.Marks)
	}
}

func TestHistoryHarbingerMark(t *testing.T) {
	ge := fateEngine(t, "medieval_age", 9)
	forceFate(t, ge, int(baseEraTicks("iron_era"))-1)
	ge.fateArrive()
	ge.rng = riggedRNG(0.01)
	readyToAdvance(ge)
	_ = ge.AdvanceAge()
	if len(ge.harbingerHistory) != 1 || markCount(ge, MarkHarbinger) != 1 {
		t.Errorf("%d resolved threads, harbinger marks: %+v", len(ge.harbingerHistory), ge.history.Marks)
	}
}

// Every run starts its history over, and the new run's own age mark is kept.
func TestHistoryStartsOverWithEachRun(t *testing.T) {
	check := func(t *testing.T, ge *GameEngine) {
		t.Helper()
		h := ge.history
		if h.Recent.N != 0 || h.Run.N != 0 || len(h.Marks) != 0 || len(h.Run.Cols) != 0 {
			t.Fatalf("the new run starts with %d recent and %d run buckets and %d marks, want none", h.Recent.N, h.Run.N, len(h.Marks))
		}
		ge.StepTicks(100)
		ge.advanceAge("stone_age")
		ge.StepTicks(50)
		if first := h.Recent.TickAt(0); first != historyInterval || h.Recent.TickAt(h.Recent.N-1) != 150 {
			t.Errorf("the new run's samples run %d - %d, want %d - 150", first, h.Recent.TickAt(h.Recent.N-1), historyInterval)
		}
		if !slices.Contains(h.Marks, HistoryMark{Tick: 100, Kind: MarkAge, Key: "stone_age"}) {
			t.Errorf("the new run's age mark is gone: %+v", h.Marks)
		}
	}
	t.Run("prestige", func(t *testing.T) {
		ge := catEngine(t, "medieval_age", 11)
		ge.StepTicks(400)
		ge.advanceAge("modern_age")
		ge.StepTicks(200)
		if err := ge.DoPrestige(); err != nil {
			t.Fatalf("DoPrestige: %v", err)
		}
		check(t, ge)
	})
	t.Run("succumb", func(t *testing.T) {
		ge := catEngine(t, "bronze_age", 5)
		ge.StepTicks(400)
		ge.advanceAge("iron_age")
		ge.StepTicks(200)
		succumbIn(t, ge, "iron_age")
		check(t, ge)
	})
	t.Run("new game", func(t *testing.T) {
		ge := newSeededEngine(3)
		ge.StepTicks(400)
		ge.advanceAge("stone_age")
		ge.StepTicks(200)
		ge.Reset()
		check(t, ge)
	})
}

// The columns the engine feeds: the three counts, a rate for each resource the
// run has met, and none for one it has not.
func TestHistoryEngineColumns(t *testing.T) {
	ge := newSeededEngine(1)
	ge.StepTicks(200)
	r := &ge.history.Recent
	for _, k := range []string{"population", "food_rate", "know_rate", "faith", "morale", "prod_all", "tick_speed", "buildings", "techs", "expeditions", "rate.food"} {
		if len(r.Cols[k]) != r.N {
			t.Errorf("column %s has %d values for %d buckets", k, len(r.Cols[k]), r.N)
		}
	}
	for k := range ge.Resources.resources {
		if _, has := r.Cols["rate."+k]; has != ge.Resources.IsUnlocked(k) {
			t.Errorf("resource %s: rate column %v, unlocked %v", k, has, ge.Resources.IsUnlocked(k))
		}
	}
	if ge.Resources.IsUnlocked("food") && r.Cols["rate.food"][r.N-1] != r.Cols["food_rate"][r.N-1] {
		t.Errorf("rate.food %v differs from food_rate %v", r.Cols["rate.food"][r.N-1], r.Cols["food_rate"][r.N-1])
	}
}

// A 24-hour catch-up leaves no gap in the run tier and one away mark.
func TestHistoryTimeAway(t *testing.T) {
	ge := newSeededEngine(1)
	ge.StepTicks(1500)
	ge.SimulateOffline(MaxOfflineTime)
	want := int(MaxOfflineTime / ge.getTickInterval())
	h := ge.history
	var away []HistoryMark
	for _, m := range h.Marks {
		if m.Kind == MarkAway {
			away = append(away, m)
		}
	}
	if len(away) != 1 || away[0].Tick != 1500 || away[0].Ticks != want {
		t.Fatalf("away marks %+v, want one at tick 1500 of %d ticks", away, want)
	}
	end := ge.tick / historyInterval * historyInterval
	for name, tier := range map[string]*HistoryTier{"recent": &h.Recent, "run": &h.Run} {
		if tier.End != end {
			t.Errorf("%s ends at tick %d, want %d", name, tier.End, end)
		}
	}
	run := &h.Run
	if got := (run.N-1)*run.Span + run.Fill*historyInterval; got != end {
		t.Errorf("run tier covers %d ticks, want %d: a gap in the time away", got, end)
	}
	if run.N > historyRunMax {
		t.Errorf("%d run buckets", run.N)
	}
	// The held values leave no spike: the population column stays within what
	// it was before the catch-up and what it is after.
	pop := run.Cols["population"]
	hi := slices.Max(pop)
	if lim := float32(ge.popLocked()) * 1.01; hi > lim && hi > pop[0]*1000 {
		t.Errorf("population peaks at %v in the history, now %v", hi, ge.popLocked())
	}
}

func TestHistoryAccessorCopies(t *testing.T) {
	ge := newSeededEngine(1)
	ge.StepTicks(100)
	ge.advanceAge("stone_age")
	c := ge.History()
	c.Recent.Cols["population"][0] = -1
	c.Run.Cols["population"][0] = -1
	c.Marks[0].Key = "changed"
	c.Marks = append(c.Marks, HistoryMark{})
	if ge.history.Recent.Cols["population"][0] == -1 || ge.history.Run.Cols["population"][0] == -1 ||
		ge.history.Marks[0].Key == "changed" || len(ge.history.Marks) != 1 {
		t.Error("changing the copy changed the engine's history")
	}
}

// fullHistory is a recorder with both tiers full of incompressible columns.
func fullHistory(cols, marks int) *RunHistory {
	rng := rand.New(rand.NewSource(1))
	h := NewRunHistory()
	for i := 1; i <= 4*historyRunMax*historyRecentLen; i++ {
		vals := make(map[string]float64, cols)
		for c := 0; c < cols; c++ {
			vals[fmt.Sprintf("rate.resource_%02d", c)] = rng.Float64() * math.Pow(10, float64(rng.Intn(20)))
		}
		h.Record(10*i, vals)
		if h.Run.N == historyRunMax && h.Run.Fill == h.Run.Span/historyInterval && h.Recent.N == historyRecentLen {
			break
		}
	}
	for i := 0; i < marks; i++ {
		h.Mark(i*1000, MarkAge, "classical_age", 0)
	}
	return h
}

func TestHistorySaveRoundTripAndSize(t *testing.T) {
	h := fullHistory(34, 0)
	if h.Run.N != historyRunMax || h.Recent.N != historyRecentLen || len(h.Run.Cols) != 34 {
		t.Fatalf("setup: %d run buckets, %d recent, %d columns", h.Run.N, h.Recent.N, len(h.Run.Cols))
	}
	h.Mark(12, MarkAge, "stone_age", 0)
	h.Mark(30, MarkAway, "", 4000)
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	var back RunHistory
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if !historyEqual(h, &back) {
		t.Error("the history changed in a save and load")
	}
	t.Logf("34 columns, both tiers full, 2 marks: %d bytes", len(b))
	if len(b) >= 160*1000 {
		t.Errorf("a full history saves in %d bytes, budget 160 KB", len(b))
	}
	full, _ := json.Marshal(fullHistory(34, historyMarksMax))
	t.Logf("34 columns, both tiers full, %d marks: %d bytes", historyMarksMax, len(full))
	mid, _ := json.Marshal(fullHistory(14, 60))
	t.Logf("14 columns, both tiers full, 60 marks: %d bytes", len(mid))
}

func historyEqual(a, b *RunHistory) bool {
	tiers := func(x, y *HistoryTier) bool {
		if x.Span != y.Span || x.N != y.N || x.Fill != y.Fill || x.End != y.End || len(x.Cols) != len(y.Cols) {
			return false
		}
		for k, c := range x.Cols {
			if !slices.Equal(c, y.Cols[k]) {
				return false
			}
		}
		return true
	}
	return tiers(&a.Recent, &b.Recent) && tiers(&a.Run, &b.Run) && slices.Equal(a.Marks, b.Marks)
}

// A save written with the old history key and none of the new loads to an
// empty recorder, and a damaged column leaves the rest of the save loadable.
func TestHistoryLoadsOldAndDamagedSaves(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.StepTicks(120)
	ge.advanceAge("stone_age")
	b := saveAndLoad(t, ge, "hist_roundtrip")
	if !historyEqual(ge.history, b.history) {
		t.Error("the recorder changed through SaveGame and LoadGame")
	}

	ge.mu.Lock()
	ge.history = nil
	ge.mu.Unlock()
	old := saveWithout(t, ge, "hist_old", func(m map[string]any) {
		m["history"] = map[string]any{"samples": []any{map[string]any{"tick": 10, "population": 5}}, "age_markers": []any{}}
	})
	if old.history == nil || old.history.Run.N != 0 || old.history.Recent.N != 0 || len(old.history.Marks) != 0 {
		t.Errorf("an old save loads to %+v, want an empty recorder", old.history)
	}
	bad := saveWithout(t, b, "hist_bad", func(m map[string]any) {
		m["run_history"].(map[string]any)["run"].(map[string]any)["cols"].(map[string]any)["population"] = "!!not base64"
	})
	if bad.history.Run.N != 0 {
		t.Errorf("a damaged tier loaded %d buckets", bad.history.Run.N)
	}
}

// saveWithout saves ge, edits the save's JSON with edit (dropping the
// integrity signature, which the loader then flags but does not refuse) and
// loads it into a new engine.
func saveWithout(t *testing.T, ge *GameEngine, name string, edit func(map[string]any)) *GameEngine {
	t.Helper()
	if ge.history == nil {
		ge.history = NewRunHistory()
	}
	noErr(t, ge.SaveGame(name))
	path := savePath(name)
	data, err := os.ReadFile(path)
	noErr(t, err)
	var m map[string]any
	noErr(t, json.Unmarshal(data, &m))
	delete(m, "history")
	delete(m, "run_history")
	if hist, ok := loadedHistory(data); ok {
		m["run_history"] = hist
	}
	edit(m)
	out, err := json.Marshal(m)
	noErr(t, err)
	noErr(t, os.WriteFile(path, out, 0o600))
	again := NewGameEngine()
	noErr(t, again.LoadGame(name))
	return again
}

func loadedHistory(data []byte) (any, bool) {
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return nil, false
	}
	h, ok := m["run_history"]
	return h, ok
}
