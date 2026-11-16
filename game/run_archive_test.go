package game

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// archiveEngine is an engine in age that holds a fresh account, with a run of
// steps ticks behind it.
func archiveEngine(t *testing.T, age string, seed int64, steps int) (*GameEngine, *Account) {
	t.Helper()
	isolateAccountDir(t)
	acct := testAccount(t)
	ge := catEngine(t, age, seed)
	ge.SetAccount(acct)
	ge.StepTicks(steps)
	return ge, acct
}

func onlyRun(t *testing.T, ge *GameEngine) RunRecord {
	t.Helper()
	v := ge.RunArchive()
	if len(v.Runs) != 1 {
		t.Fatalf("%d runs filed, want 1", len(v.Runs))
	}
	return v.Runs[0]
}

func checkArchivedShape(t *testing.T, h *RunHistory) {
	t.Helper()
	if h == nil {
		t.Fatal("the record has no history")
	}
	if h.Run.N > runArchiveBuckets {
		t.Errorf("%d buckets, want at most %d", h.Run.N, runArchiveBuckets)
	}
	if h.Recent.N != 0 {
		t.Errorf("the recent ring is filed: %d buckets", h.Recent.N)
	}
	for k, c := range h.Run.Cols {
		if !slices.Contains(runArchiveCols, k) {
			t.Errorf("column %q is filed", k)
		}
		if len(c) != h.Run.N {
			t.Errorf("column %q has %d values for %d buckets", k, len(c), h.Run.N)
		}
	}
	for _, m := range h.Marks {
		if !slices.Contains(runArchiveMarks, m.Kind) {
			t.Errorf("a %q mark is filed", m.Kind)
		}
	}
}

func TestRunArchivePrestigeFilesTheRun(t *testing.T) {
	ge, _ := archiveEngine(t, "medieval_age", 11, 400)
	setPrestigeLevel(ge, 2)
	ge.advanceAge("modern_age")
	ge.StepTicks(2200) // far enough to merge the run tier at least once
	ge.Stats.TotalBuilt, ge.Stats.TotalRecruited = 41, 17
	ge.Military.completedCount = 3
	ge.history.Mark(5, MarkEvent, "x", 0)
	ge.history.Mark(6, MarkAway, "", 100)
	if len(ge.history.Run.Cols) <= len(runArchiveCols) {
		t.Fatalf("setup: only %d columns", len(ge.history.Run.Cols))
	}
	before, tick := ge.runID, ge.tick
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	r := onlyRun(t, ge)
	if r.ID != before || r.Outcome != RunPrestige || r.Age != "modern_age" || r.Level != 2 || r.Tick != tick {
		t.Errorf("record %+v: want id %s, prestige, modern_age, level 2, tick %d", r, before, tick)
	}
	if r.Built != 41 || r.Recruited != 17 || r.Expeditions != 3 || r.Ended == 0 || r.Started == 0 {
		t.Errorf("counts and times wrong: %+v", r)
	}
	checkArchivedShape(t, r.History)
	if len(r.History.Marks) == 0 {
		t.Error("no marks filed")
	}
	if h := ge.history; h.Run.N != 0 || h.Recent.N != 0 || len(h.Marks) != 0 {
		t.Errorf("the live history holds %d run buckets and %d marks after the prestige", h.Run.N, len(h.Marks))
	}
	if ge.runID == before || len(ge.runID) != 16 {
		t.Errorf("the new run's id is %q (was %q)", ge.runID, before)
	}
	if got := ge.RunArchive().RunID; got != ge.runID {
		t.Errorf("the view names run %q, want %q", got, ge.runID)
	}
}

func TestRunArchiveSuccumbAndShortRuns(t *testing.T) {
	t.Run("succumbed", func(t *testing.T) {
		ge, _ := archiveEngine(t, "bronze_age", 5, 400)
		succumbIn(t, ge, "iron_age")
		if r := onlyRun(t, ge); r.Outcome != RunSuccumbed || r.Age != "iron_age" {
			t.Errorf("record %+v", r)
		}
	})
	t.Run("a short run is not filed", func(t *testing.T) {
		ge, _ := archiveEngine(t, "bronze_age", 5, 100)
		succumbIn(t, ge, "iron_age")
		if n := len(ge.RunArchive().Runs); n != 0 {
			t.Errorf("%d runs filed for a run of %d buckets", n, runArchiveMinBuckets-1)
		}
	})
	t.Run("a short prestige is", func(t *testing.T) {
		ge, _ := archiveEngine(t, "medieval_age", 11, 100)
		ge.advanceAge("modern_age")
		if err := ge.DoPrestige(); err != nil {
			t.Fatal(err)
		}
		if r := onlyRun(t, ge); r.Outcome != RunPrestige {
			t.Errorf("record %+v", r)
		}
	})
	t.Run("a new game files the run it replaces as left", func(t *testing.T) {
		ge, _ := archiveEngine(t, "bronze_age", 5, 400)
		ge.Reset()
		if r := onlyRun(t, ge); r.Outcome != RunLeft {
			t.Errorf("record %+v", r)
		}
	})
	t.Run("a run of another account is not filed", func(t *testing.T) {
		ge, acct := archiveEngine(t, "bronze_age", 5, 400)
		ge.runAccountID = "0123456789abcdef0123456789abcdef"
		succumbIn(t, ge, "iron_age")
		if len(acct.runs) != 0 {
			t.Errorf("%d runs filed on an account that does not own the run", len(acct.runs))
		}
	})
}

// thinCheck thins a copy of the run tier to limit and checks what must hold.
func thinCheck(t *testing.T, h *RunHistory, limit int) {
	t.Helper()
	orig := h.Run.clone()
	tier := h.Run.clone()
	tier.thin(limit)
	label := fmt.Sprintf("%d buckets (fill %d) thinned to %d", orig.N, orig.Fill, limit)
	if tier.N > limit && limit >= 1 {
		t.Fatalf("%s: %d buckets", label, tier.N)
	}
	if tier.End != orig.End {
		t.Errorf("%s: End moved, %d to %d", label, orig.End, tier.End)
	}
	if got, want := tier.TickAt(0)-tier.Span, orig.TickAt(0)-orig.Span; got != want {
		t.Errorf("%s: the first bucket starts at %d, was %d", label, got, want)
	}
	covered := func(x HistoryTier) int { return (x.N-1)*x.Span + x.Fill*historyInterval }
	if covered(tier) != covered(orig) {
		t.Errorf("%s: covers %d ticks, was %d", label, covered(tier), covered(orig))
	}
	if tier.Fill < 1 || tier.Fill > tier.Span/historyInterval {
		t.Errorf("%s: fill %d of %d", label, tier.Fill, tier.Span/historyInterval)
	}
	if orig.N%2 == 1 && tier.Span == 2*orig.Span {
		// The odd bucket stands alone: it starts where it did.
		if got, want := tier.TickAt(tier.N-1)-tier.Span, orig.TickAt(orig.N-1)-orig.Span; got != want {
			t.Errorf("%s: the last bucket starts at %d, was %d", label, got, want)
		}
	}
	if c, o := tier.Cols["buildings"], orig.Cols["buildings"]; c[len(c)-1] != o[len(o)-1] {
		t.Errorf("%s: the count column ends on %v, was %v", label, c[len(c)-1], o[len(o)-1])
	}
	// The mean of a column is the mean of its samples, however the buckets merged.
	mean := func(x HistoryTier) float64 {
		var sum, w float64
		for i, v := range x.Cols["ramp"] {
			n := float64(x.Span / historyInterval)
			if i == x.N-1 {
				n = float64(x.Fill)
			}
			sum += float64(v) * n
			w += n
		}
		return sum / w
	}
	if a, b := mean(tier), mean(orig); math.Abs(a-b) > 1e-4*b {
		t.Errorf("%s: the mean is %v, was %v", label, a, b)
	}
}

func TestRunHistoryThinsAnyTier(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5, 7, 100, 255, 257, 511, 512, 513, 777, 1000, 1025, 1300, 2049} {
		h := NewRunHistory()
		for i := 1; i <= n; i++ {
			h.Record(10*i, map[string]float64{"ramp": float64(i), "buildings": float64(i)})
		}
		for _, limit := range []int{1, 3, 4, 7, 256} {
			thinCheck(t, h, limit)
		}
	}
}

// testRun is a record with a small history.
func testRun(id string, ended int64, age string, tick int, history bool) RunRecord {
	r := RunRecord{ID: id, Ended: ended, Tick: tick, Age: age, Outcome: RunPrestige, Started: ended - 100}
	if history {
		h := NewRunHistory()
		for i := 1; i <= 40; i++ {
			h.Record(10*i, map[string]float64{"population": float64(i)})
		}
		r.History = h.archived()
	}
	return r
}

func TestRunArchiveLimits(t *testing.T) {
	a := &Account{AccountID: "0123456789abcdef0123456789abcdef"}
	// 25 runs, the oldest of them the best.
	a.fileRun(testRun("old-best", 1, "modern_age", 9000, true))
	for i := 1; i < 25; i++ {
		a.fileRun(testRun(fmt.Sprintf("r%02d", i), int64(1+i), "stone_age", 100*i, true))
	}
	runs := a.runRecords()
	if len(runs) != 25 || runs[0].ID != "r24" {
		t.Fatalf("%d runs, newest %q", len(runs), runs[0].ID)
	}
	for i, r := range runs {
		want := i < runArchiveFull || r.ID == "old-best"
		if (r.History != nil) != want {
			t.Errorf("run %d (%s): history %v, want %v", i, r.ID, r.History != nil, want)
		}
	}
	// The same id twice is one record, the later filing.
	a.fileRun(testRun("r24", 99, "stone_age", 5, true))
	runs = a.runRecords()
	if len(runs) != 25 || runs[0].ID != "r24" || runs[0].Tick != 5 {
		t.Errorf("a run filed twice: %d records, newest %+v", len(runs), runs[0])
	}
	// A second best by age then ticks.
	a.fileRun(testRun("faster", 2, "modern_age", 8000, true))
	if r := a.runRecords(); r[len(r)-1].ID != "old-best" || r[len(r)-1].History != nil {
		t.Errorf("the run beaten by a faster one of its age keeps its history")
	}
	// Past 200 the oldest goes.
	b := &Account{AccountID: a.AccountID}
	for i := 1; i <= runArchiveMax+1; i++ {
		b.fileRun(testRun(fmt.Sprintf("s%03d", i), int64(i), "stone_age", i, false))
	}
	runs = b.runRecords()
	if len(runs) != runArchiveMax || runs[0].ID != "s201" || runs[len(runs)-1].ID != "s002" {
		t.Errorf("%d summaries, from %s to %s", len(runs), runs[0].ID, runs[len(runs)-1].ID)
	}
	// What the view hands out is a copy.
	runs[0].Outcome = "changed"
	if b.runRecords()[0].Outcome == "changed" {
		t.Error("changing a record of the view changed the account's")
	}
	c := a.runRecords()
	c[0].History.Run.Cols["population"][0] = -1
	if a.runRecords()[0].History.Run.Cols["population"][0] == -1 {
		t.Error("changing a history of the view changed the account's")
	}
}

func recordsJSON(t *testing.T, runs []RunRecord) string {
	t.Helper()
	b, err := json.Marshal(runs)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRunArchiveFileRoundTripAndDamage(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	acct.fileRun(testRun("a", 10, "stone_age", 10, true))
	acct.fileRun(testRun("b", 20, "bronze_age", 20, true))
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(accountDir(acct.AccountID), runArchiveFileName)
	good, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloadAccount(t); recordsJSON(t, got.runs) != recordsJSON(t, acct.runs) || len(got.runs) != 2 {
		t.Fatalf("the records changed in a save and load: %d", len(got.runs))
	}
	var f runArchiveFile
	if err := json.Unmarshal(good, &f); err != nil {
		t.Fatal(err)
	}
	f.Runs[0].Built = 999 // edited, signature kept
	edited, _ := json.Marshal(&f)
	foreign := runArchiveFile{Version: runArchiveVersion, AccountID: "ffffffffffffffffffffffffffffffff", Runs: f.Runs}
	foreign.Signature = signRunArchive(&foreign)
	foreignData, _ := json.Marshal(&foreign)
	for name, data := range map[string][]byte{"garbage": []byte("{not json"), "bad signature": edited, "another account's": foreignData, "empty": {}} {
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		got := reloadAccount(t)
		if len(got.runs) != 0 || got.Tampered || got.BadgesTampered {
			t.Errorf("%s: %d runs, tampered %v/%v", name, len(got.runs), got.Tampered, got.BadgesTampered)
		}
		// Overwritten at the next filing.
		got.fileRun(testRun("c", 30, "stone_age", 30, false))
		if err := got.FlushIfDirty(); err != nil {
			t.Fatal(err)
		}
		if again := reloadAccount(t); len(again.runs) != 1 || again.runs[0].ID != "c" {
			t.Errorf("%s: the file was not rewritten: %d runs", name, len(again.runs))
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got := reloadAccount(t); len(got.runs) != 0 {
		t.Errorf("a missing file loaded %d runs", len(got.runs))
	}
}

func TestRunArchiveExportAndImport(t *testing.T) {
	isolateAccountDir(t)
	src := testAccount(t)
	src.fileRun(testRun("r1", 10, "stone_age", 10, true))
	src.fileRun(testRun("r2", 20, "bronze_age", 20, true))
	blob, err := src.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}
	want := recordsJSON(t, src.runs)

	// Into a fresh data root: the slot is made from the export.
	isolateAccountDir(t)
	got, err := ImportAccountExport(blob, false)
	if err != nil {
		t.Fatal(err)
	}
	if recordsJSON(t, got.runs) != want {
		t.Errorf("a fresh account holds %d runs after the import, want the export's", len(got.runs))
	}
	if disk := reloadSlot(t, got.AccountID); recordsJSON(t, disk.runs) != want {
		t.Error("the imported runs were not saved")
	}

	// Into an account with its own runs: a merge unions by id.
	isolateAccountDir(t)
	local := testAccount(t)
	local.fileRun(testRun("r2", 25, "bronze_age", 99, true)) // same id, later end tick
	local.fileRun(testRun("r3", 30, "stone_age", 30, true))
	for i := 0; i < 198; i++ {
		local.fileRun(testRun(fmt.Sprintf("x%03d", i), int64(100+i), "stone_age", i, false))
	}
	if err := local.Save(); err != nil {
		t.Fatal(err)
	}
	merged, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.runs) != runArchiveMax {
		t.Errorf("%d runs after a merge, want the limit %d", len(merged.runs), runArchiveMax)
	}
	ids := map[string]RunRecord{}
	for _, r := range merged.runs {
		ids[r.ID] = r
	}
	if _, ok := ids["r1"]; ok {
		t.Error("the oldest run r1 survived the limit")
	}
	if r := ids["r2"]; r.Tick != 99 {
		t.Errorf("r2 is %+v, want the one with the larger end tick (99)", r)
	}
	if _, ok := ids["r3"]; !ok {
		t.Error("the account's own run r3 is missing")
	}
	// A replace import takes the export's records alone.
	replaced, err := ImportAccountExport(blob, false)
	if err != nil {
		t.Fatal(err)
	}
	if recordsJSON(t, replaced.runs) != want {
		t.Errorf("a replace import left %d runs", len(replaced.runs))
	}
	// An export whose run store was edited brings none, and imports otherwise.
	var exp progressExport
	if err := json.Unmarshal(blob, &exp); err != nil {
		t.Fatal(err)
	}
	exp.RunStore.Runs[0].Built = 5
	bad, _ := json.Marshal(&exp)
	isolateAccountDir(t)
	none, err := ImportAccountExport(bad, true)
	if err != nil || len(none.runs) != 0 {
		t.Errorf("an edited run store: err %v, %d runs", err, len(none.runs))
	}
}

func reloadSlot(t *testing.T, id string) *Account {
	t.Helper()
	acct, found, err := loadAccountFromSlot(id)
	if err != nil || !found {
		t.Fatalf("loadAccountFromSlot: %v %v", found, err)
	}
	return acct
}

// The worst the file gets: 21 records with a full history of ten incompressible
// columns and 80 marks, and 200 summaries.
func TestRunArchiveSize(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	full := func(i int, marks int) RunRecord {
		h := NewRunHistory()
		h.Run = HistoryTier{Span: 40, N: runArchiveBuckets, Fill: 4, End: 40 * runArchiveBuckets, Cols: map[string][]float32{}}
		for _, k := range runArchiveCols {
			c := make([]float32, runArchiveBuckets)
			for j := range c {
				c[j] = float32(rng.Float64() * math.Pow(10, float64(rng.Intn(20))))
			}
			h.Run.Cols[k] = c
		}
		for m := 0; m < marks; m++ {
			h.Marks = append(h.Marks, HistoryMark{Tick: m * 1000, Kind: MarkWonder, Key: "hanging_gardens_of_babylon"})
		}
		return RunRecord{ID: fmt.Sprintf("%016x", i), Run: "a long save name", Started: 1700000000, Ended: 1700000000 + int64(i), Tick: 1000000 + i,
			Outcome: RunSuccumbed, Age: "transcendent_age", Built: 99999, Recruited: 99999, Researched: 999, Expeditions: 999, Level: 30, History: h}
	}
	f := &runArchiveFile{Version: runArchiveVersion, AccountID: "0123456789abcdef0123456789abcdef"}
	for i := 0; i < runArchiveFull+1; i++ {
		f.Runs = append(f.Runs, full(i, 80))
	}
	for i := 0; i < runArchiveMax; i++ {
		s := full(1000+i, 0)
		s.History = nil
		f.Runs = append(f.Runs, s)
	}
	f.Signature = signRunArchive(f)
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	one, _ := json.Marshal(full(0, 40))
	t.Logf("worst case file: %d bytes; one full record (256 buckets, 40 marks): %d bytes", len(data), len(one))
	if len(data) > 512000 {
		t.Errorf("the run file is %d bytes, budget 512,000", len(data))
	}
}
