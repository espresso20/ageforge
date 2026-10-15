package game

import (
	"testing"
	"time"
)

func TestRecordAgeLogsEachAgeOnce(t *testing.T) {
	gs := NewGameStats()
	at := time.Now()
	gs.RecordAge("stone_age", 120, at)
	gs.RecordAge("stone_age", 300, at.Add(time.Minute))

	if len(gs.AgeLog) != 2 {
		t.Fatalf("AgeLog has %d entries, want 2 (primitive, stone): %+v", len(gs.AgeLog), gs.AgeLog)
	}
	got := gs.AgeLog[1]
	if got.Age != "stone_age" || got.Tick != 120 || !got.At.Equal(at) {
		t.Errorf("stone entry = %+v, want stone_age at tick 120, %v", got, at)
	}
	if len(gs.AgesReached) != 2 {
		t.Errorf("AgesReached = %v, want primitive and stone", gs.AgesReached)
	}
}

func TestNewGameStatsLogsThePrimitiveAgeAtTickZero(t *testing.T) {
	gs := NewGameStats()
	if len(gs.AgeLog) != 1 {
		t.Fatalf("AgeLog has %d entries, want 1", len(gs.AgeLog))
	}
	if e := gs.AgeLog[0]; e.Age != "primitive_age" || e.Tick != 0 || e.At.IsZero() {
		t.Errorf("first entry = %+v, want primitive_age at tick 0 with a time", e)
	}
}

func TestAdvanceAgeRecordsTheTickOfEntry(t *testing.T) {
	ge := NewGameEngine()
	ge.tick = 500
	ge.advanceAge("stone_age")

	log := ge.Stats.AgeLog
	if len(log) != 2 {
		t.Fatalf("AgeLog has %d entries, want 2: %+v", len(log), log)
	}
	if log[1].Age != "stone_age" || log[1].Tick != 500 {
		t.Errorf("entry = %+v, want stone_age at tick 500", log[1])
	}
}

func TestPrestigeStartsANewAgeLog(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "modern_age", 11)
	ge.tick = 4000
	ge.Stats.RecordAge("modern_age", 3900, time.Now())

	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	log := ge.Stats.AgeLog
	if len(log) != 1 {
		t.Fatalf("after prestige AgeLog has %d entries, want 1: %+v", len(log), log)
	}
	if log[0].Age != "primitive_age" || log[0].Tick != ge.tick {
		t.Errorf("after prestige entry = %+v, want primitive_age at the current tick %d", log[0], ge.tick)
	}
}

func TestAgeLogSurvivesSaveAndLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := NewGameEngine()
	ge.tick = 500
	ge.advanceAge("stone_age")
	ge.tick = 900
	ge.advanceAge("bronze_age")
	if err := ge.SaveGame("age_log"); err != nil {
		t.Fatal(err)
	}

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("age_log"); err != nil {
		t.Fatal(err)
	}
	want, got := ge.Stats.AgeLog, ge2.Stats.AgeLog
	if len(got) != len(want) {
		t.Fatalf("loaded AgeLog has %d entries, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Age != want[i].Age || got[i].Tick != want[i].Tick || !got[i].At.Equal(want[i].At) {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}
