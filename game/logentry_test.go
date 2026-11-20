package game

import (
	"fmt"
	"testing"
	"time"
)

func logEngine() *GameEngine {
	ge := NewGameEngine()
	ge.log = nil
	return ge
}

func foldLine(ge *GameEngine, tick int, n float64) {
	ge.tick = tick
	ge.addLog(OfficeBuilding, "success", "Hut built.", logID("t.hut"), logSubj("hut"),
		logPaid(LogAmt{"wood", n}), logGot(LogAmt{"huts", 1}))
}

func TestFoldJoinsAt70NotAt71(t *testing.T) {
	ge := logEngine()
	foldLine(ge, 100, 5)
	foldLine(ge, 170, 5) // 70 ticks after the newest member: joins
	if len(ge.log) != 1 || ge.log[0].Count != 2 {
		t.Fatalf("after a line 70 ticks on: %d entries, count %d", len(ge.log), ge.log[0].Count)
	}
	foldLine(ge, 241, 5) // 71 ticks after the fold's newest member: a new line
	if len(ge.log) != 2 {
		t.Fatalf("a line 71 ticks on folded: %d entries", len(ge.log))
	}
}

func TestFoldSumsAndKeepsTheNewestTwelve(t *testing.T) {
	ge := logEngine()
	for i := 0; i < 20; i++ {
		foldLine(ge, 100+i, 2)
	}
	e := ge.log[0]
	if len(ge.log) != 1 || e.Count != 20 || len(e.Members) != foldMembers || e.Earlier() != 8 {
		t.Fatalf("count %d, members %d, earlier %d", e.Count, len(e.Members), e.Earlier())
	}
	if e.Paid[0].N != 40 || e.Gained[0].N != 20 {
		t.Errorf("paid %v, gained %v, want 40 wood and 20 huts", e.Paid, e.Gained)
	}
	if e.FirstAt == 0 || e.Tick != 119 {
		t.Errorf("first %d, newest tick %d", e.FirstAt, e.Tick)
	}
}

func TestFoldNeedsSameOfficeSubjectAndID(t *testing.T) {
	ge := logEngine()
	ge.addLog(OfficeBuilding, "info", "a", logID("x"), logSubj("one"))
	ge.addLog(OfficeBuilding, "info", "b", logID("x"), logSubj("two"))
	ge.addLog(OfficeResearch, "info", "c", logID("x"), logSubj("one"))
	ge.addLog(OfficeBuilding, "info", "d", logSubj("one"))
	ge.addLog(OfficeBuilding, "info", "e", logSubj("one"))
	if len(ge.log) != 5 {
		t.Errorf("%d entries, want 5: nothing here should fold", len(ge.log))
	}
}

func TestLogCapAndLevels(t *testing.T) {
	ge := logEngine()
	for i := 0; i < MaxLogSize+50; i++ {
		ge.tick = i
		ge.addLog(OfficeGame, "info", fmt.Sprint(i))
	}
	if len(ge.log) != MaxLogSize || ge.log[0].Message != "50" {
		t.Fatalf("%d entries, first %q", len(ge.log), ge.log[0].Message)
	}
	ge.addLog(OfficeGame, LogRoutine, "r")
	ge.addLog(OfficeGame, "event", "h", logHead)
	if ge.log[len(ge.log)-2].Level != LevelEverything || ge.log[len(ge.log)-1].Level != LevelHeadline || ge.log[0].Level != LevelTown {
		t.Error("levels: routine is Everything, a headline option Headline, the rest Town")
	}
}

func TestLogSavesAndLoads(t *testing.T) {
	ge := logEngine()
	for i := 0; i < MaxLogSize; i++ {
		ge.tick = i
		ge.addLog(Office(i%int(officeCount)), "info", fmt.Sprintf("Line %d: some words that read like a real log line.", i),
			logSubj(fmt.Sprint(i%37)), logPaid(LogAmt{"wood", float64(i)}))
	}
	blob := ge.logForSave()
	if len(blob) > 400_000 {
		t.Errorf("2,000 lines save as %d bytes, want under 400 KB", len(blob))
	}
	back := logFromSave(blob)
	if len(back) != MaxLogSize || back[1999].Message != ge.log[1999].Message || back[5].Paid[0].N != 5 {
		t.Fatalf("round trip lost lines: %d", len(back))
	}
	if logFromSave("not a log") != nil || logFromSave("") != nil {
		t.Error("a blob that does not read should be an empty log")
	}
}

func TestLoadWritesOneSeam(t *testing.T) {
	ge := logEngine()
	ge.SetClock(func() time.Time { return time.Unix(5000, 0) })
	ge.addLog(OfficeGame, "info", "a", logID("x"))
	ge.addLog(OfficeGame, "info", "b")
	ge.logSeam()
	ge.logSeam() // nothing new since: no second seam
	if len(ge.log) != 3 || !ge.log[2].Seam || ge.log[2].Count != 2 || ge.log[2].FirstAt != 5000 {
		t.Fatalf("log %+v", ge.log)
	}
	ge.tick = 1
	ge.addLog(OfficeGame, "info", "c", logID("x"))
	if len(ge.log) != 4 {
		t.Errorf("%d entries: the fold should have stopped at the seam", len(ge.log))
	}
}

func TestEveryLineHasAnOfficeAndLevel(t *testing.T) {
	ge := NewGameEngine()
	ge.SeedRNG(3)
	for i := 0; i < 30; i++ {
		_, _ = ge.GatherResource("wood", 3)
		_, _ = ge.GatherResource("food", 3)
	}
	_ = ge.BuildBuilding("hut")
	ge.StepTicks(400)
	for _, e := range ge.log {
		if int(e.Office) < 1 || int(e.Office) > officeCount || e.Level < LevelHeadline || e.Level > LevelEverything || e.At == 0 {
			t.Errorf("%q: office %d, level %d, at %d", e.Message, e.Office, e.Level, e.At)
		}
	}
}

func TestFoodWarningFiresOnceAndCloses(t *testing.T) {
	ge := logEngine()
	ge.Resources.resources["food"].Amount = 100
	ge.Resources.resources["food"].Rate = -1 // gone in 100 ticks
	ge.tick = 1
	ge.foodWarning()
	ge.foodWarning()
	if len(ge.log) != 1 || !ge.log[0].Trouble || ge.log[0].Level != LevelTown {
		t.Fatalf("log %+v", ge.log)
	}
	ge.tick = 50
	ge.Resources.resources["food"].Rate = 1
	ge.foodWarning()
	if len(ge.log) != 2 || !ge.log[1].Trouble || ge.foodWarned {
		t.Fatalf("the closing line is missing: %+v", ge.log)
	}
}
