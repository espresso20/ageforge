package game

import (
	"github.com/espresso20/ageforge/config"
	"maps"
	"os"
	"strings"
	"testing"
	"time"
)

var recAt = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func TestRecordAgeTimeKeepsTheFastest(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	a.RecordAgeTime("stone_age", 0, recAt, "r0")
	if len(a.Stats.FastestAge) != 0 || a.dirty {
		t.Fatalf("a zero-tick time was recorded: %v", a.Stats.FastestAge)
	}
	a.RecordAgeTime("stone_age", 500, recAt, "slow")
	a.RecordAgeTime("stone_age", 300, recAt.Add(time.Hour), "fast")
	got := a.Stats.FastestAge["stone_age"]
	if got.Ticks != 300 || got.Run != "fast" || !got.At.Equal(recAt.Add(time.Hour)) {
		t.Fatalf("record = %+v, want the faster run replaced whole", got)
	}
	a.dirty = false
	a.RecordAgeTime("stone_age", 400, recAt.Add(2*time.Hour), "later")
	a.RecordAgeTime("stone_age", 300, recAt.Add(2*time.Hour), "tie")
	if a.Stats.FastestAge["stone_age"] != got || a.dirty {
		t.Fatalf("a slower or equal run changed the record: %+v dirty=%v", a.Stats.FastestAge["stone_age"], a.dirty)
	}
}

func TestRecordRunBestsOnlyRise(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	a.RecordRunBests(map[string]int{RunBestBuilt: 10, RunBestRecruited: 4, RunBestResearched: 0}, recAt, "one")
	if _, ok := a.Stats.BestRun[RunBestResearched]; ok {
		t.Fatalf("a zero count was recorded")
	}
	a.RecordRunBests(map[string]int{RunBestBuilt: 7, RunBestRecruited: 9}, recAt.Add(time.Hour), "two")
	b, r := a.Stats.BestRun[RunBestBuilt], a.Stats.BestRun[RunBestRecruited]
	if b.Value != 10 || b.Run != "one" || !b.At.Equal(recAt) {
		t.Errorf("a smaller run lowered or rewrote a best: %+v", b)
	}
	if r.Value != 9 || r.Run != "two" || !r.At.Equal(recAt.Add(time.Hour)) {
		t.Errorf("a larger run did not replace its best whole: %+v", r)
	}
}

func TestEngineRecordsAtAgeAdvanceAndRunEnd(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	ge := NewGameEngine()
	ge.SetAccount(acct)
	noErr(t, ge.StartNewNamedGame("rec"))

	ge.mu.Lock()
	ge.tick = 1200
	ge.Stats.TotalBuilt = 40
	ge.Stats.TotalRecruited = 12
	ge.advanceAge("stone_age")
	ge.mu.Unlock()
	s, _ := acct.LifetimeStats()
	if got := s.FastestAge["stone_age"]; got.Ticks != 1200 || got.Run != "rec" {
		t.Fatalf("age record = %+v", got)
	}
	if _, ok := s.FastestAge["primitive_age"]; ok {
		t.Fatalf("the first age was recorded")
	}
	if s.BestRun[RunBestBuilt].Value != 40 || s.BestRun[RunBestRecruited].Value != 12 {
		t.Fatalf("bests at an age advance = %+v", s.BestRun)
	}

	// A prestige records the run's counts before the reset; a smaller run never lowers them.
	ge.mu.Lock()
	ge.Stats.TotalBuilt = 90
	ge.mu.Unlock()
	prestigeNow(t, ge)
	s, _ = acct.LifetimeStats()
	if s.BestRun[RunBestBuilt].Value != 90 {
		t.Fatalf("best built after a prestige = %d, want 90", s.BestRun[RunBestBuilt].Value)
	}
	ge.mu.Lock()
	ge.Stats.TotalBuilt = 5
	ge.mu.Unlock()
	ge.SetPrestigeFloorForTest("")
	prestigeNow(t, ge)
	s, _ = acct.LifetimeStats()
	if s.BestRun[RunBestBuilt].Value != 90 {
		t.Fatalf("a smaller run lowered the best to %d", s.BestRun[RunBestBuilt].Value)
	}
}

func TestEngineRecordsAtSuccumb(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	ge := NewGameEngine()
	ge.SetAccount(acct)
	noErr(t, ge.StartNewNamedGame("fall"))
	ge.mu.Lock()
	ge.Stats.TotalBuilt = 33
	ge.age = "iron_age"
	ge.currentEpoch = config.EpochForAge("iron_age")
	ge.pendingCatastrophe = "iron_era"
	ge.mu.Unlock()
	noErr(t, ge.Succumb())
	s, _ := acct.LifetimeStats()
	if s.BestRun[RunBestBuilt].Value != 33 {
		t.Fatalf("best built after a Succumb = %d, want 33", s.BestRun[RunBestBuilt].Value)
	}
}

// A new game started over an unfinished run is a run end too: the abandoned run's
// counts reach the records under its own save name.
func TestEngineRecordsAtNewGame(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	ge := NewGameEngine()
	ge.SetAccount(acct)
	noErr(t, ge.StartNewNamedGame("first"))
	ge.mu.Lock()
	ge.Stats.TotalRecruited = 21
	ge.mu.Unlock()
	noErr(t, ge.StartNewNamedGame("second"))
	s, _ := acct.LifetimeStats()
	if got := s.BestRun[RunBestRecruited]; got.Value != 21 || got.Run != "first" {
		t.Fatalf("best recruited after a new game = %+v, want 21 from the run \"first\"", got)
	}
}

func TestMergeRecords(t *testing.T) {
	isolateAccountDir(t)
	mk := func(ticks int, run string, built int) AccountStats {
		return AccountStats{
			FastestAge: map[string]AgeRecord{"stone_age": {Ticks: ticks, At: recAt, Run: run}},
			BestRun:    map[string]RunBest{RunBestBuilt: {Value: built, At: recAt, Run: run}},
		}
	}
	for _, tc := range []struct {
		name         string
		local, other AccountStats
		wantRun      string
		wantTicks    int
		wantBuilt    int
	}{
		{"import is faster and larger", mk(500, "local", 10), mk(300, "other", 20), "other", 300, 20},
		{"local is faster and larger", mk(300, "local", 20), mk(500, "other", 10), "local", 300, 20},
	} {
		a := testAccount(t)
		a.Stats = tc.local
		a.applyExportLocked(&progressExport{Stats: tc.other, BadgeStore: &badgeFile{}}, true)
		got := a.Stats.FastestAge["stone_age"]
		if got.Ticks != tc.wantTicks || got.Run != tc.wantRun || a.Stats.BestRun[RunBestBuilt].Value != tc.wantBuilt || a.Stats.BestRun[RunBestBuilt].Run != tc.wantRun {
			t.Errorf("%s: %+v %+v", tc.name, got, a.Stats.BestRun[RunBestBuilt])
		}
	}
	// One-sided records are kept both ways.
	a := testAccount(t)
	a.Stats = AccountStats{FastestAge: map[string]AgeRecord{"iron_age": {Ticks: 9}}}
	a.applyExportLocked(&progressExport{BadgeStore: &badgeFile{}, Stats: AccountStats{
		FastestAge: map[string]AgeRecord{"stone_age": {Ticks: 4}},
		BestRun:    map[string]RunBest{RunBestRecruited: {Value: 3}},
	}}, true)
	if a.Stats.FastestAge["iron_age"].Ticks != 9 || a.Stats.FastestAge["stone_age"].Ticks != 4 || a.Stats.BestRun[RunBestRecruited].Value != 3 {
		t.Errorf("one-sided records lost: %+v", a.Stats)
	}
}

func TestRecordsSaveLoadVerifyAndOldFilesLoad(t *testing.T) {
	isolateAccountDir(t)
	a := testAccount(t)
	// An account saved with no records writes none of the new keys.
	noErr(t, a.Save())
	raw, err := os.ReadFile(accountPath())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "fastest_age") || strings.Contains(string(raw), "best_run") {
		t.Fatalf("an empty account wrote record keys: %s", raw)
	}
	if back := reloadAccount(t); back.Tampered {
		t.Fatalf("a file without the new fields does not verify")
	}

	a.RecordAgeTime("stone_age", 321, recAt, "save1")
	a.RecordRunBests(map[string]int{RunBestExpeditions: 6}, recAt, "save1")
	noErr(t, a.Save())
	back := reloadAccount(t)
	if back.Tampered {
		t.Fatalf("an account with records fails to verify")
	}
	if got := back.Stats.FastestAge["stone_age"]; got.Ticks != 321 || got.Run != "save1" || !got.At.Equal(recAt) {
		t.Errorf("age record after reload = %+v", got)
	}
	if got := back.Stats.BestRun[RunBestExpeditions]; got.Value != 6 || got.Run != "save1" {
		t.Errorf("run best after reload = %+v", got)
	}
}

func TestViewRecordsAreCopies(t *testing.T) {
	isolateAccountDir(t)
	acct := testAccount(t)
	ge := NewGameEngine()
	ge.SetAccount(acct)
	acct.RecordAgeTime("stone_age", 100, recAt, "x")
	acct.RecordRunBests(map[string]int{RunBestBuilt: 8}, recAt, "x")
	v := ge.GetState().AccountStats
	if v.FastestAge["stone_age"].Ticks != 100 || v.BestRun[RunBestBuilt].Value != 8 {
		t.Fatalf("view = %+v", v)
	}
	maps.DeleteFunc(v.FastestAge, func(string, AgeRecord) bool { return true })
	maps.DeleteFunc(v.BestRun, func(string, RunBest) bool { return true })
	if len(acct.Stats.FastestAge) != 1 || len(acct.Stats.BestRun) != 1 {
		t.Fatalf("changing the view changed the account")
	}
}
