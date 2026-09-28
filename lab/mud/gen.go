package main

// gen.go plays a real game headlessly with the smoke bot and writes a JSON
// snapshot of game.GameState at points of interest, so the renderer never sees
// hand-made data. Snapshots, per age the run passes through:
//
//	states/<age>_early.json  ~an hour of play after the age was entered
//	states/<age>.json        the last decision before the advance out of it
//
// The pair is what the idle-return demo diffs ("while you were away").

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/smoke"
)

// earlyAfter is how long into an age the _early snapshot is taken.
const earlyAfter = 45 * time.Minute

func generate(dir string, seed int64, stopAge string, maxSim time.Duration) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Keep any autosave the engine might attempt out of the real data dir.
	restore := game.SetDataDirForTest(filepath.Join(os.TempDir(), "ageforge-mudlab"))
	defer restore()

	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := smoke.NewBot(ge)
	bot.Harbinger = smoke.HarbingerIgnore

	idx := map[string]int{}
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}

	var sim, ageStart time.Duration
	age := ""
	early := map[string]bool{}
	started := time.Now()
	for ticks := 0; sim < maxSim; ticks++ {
		if ticks%5 == 0 {
			st := ge.GetState()
			if st.Age != age {
				age, ageStart = st.Age, sim
				fmt.Printf("%8s sim  %6s wall  entered %s\n", sim.Round(time.Minute), time.Since(started).Round(time.Second), age)
			}
			// The very first impression: a few minutes in, two or three huts.
			if age == "primitive_age" && !early["first"] && sim >= 4*time.Minute {
				early["first"] = true
				if err := writeState(dir, "primitive_age_first", st); err != nil {
					return err
				}
			}
			if !early[age] && sim-ageStart >= earlyAfter {
				early[age] = true
				if err := writeState(dir, age+"_early", st); err != nil {
					return err
				}
			}
			if st.PendingCatastrophe != "" {
				_ = ge.Endure()
				st = ge.GetState()
			}
			if st.AgeReady {
				if err := writeState(dir, st.Age, st); err != nil {
					return err
				}
				if st.Age == stopAge {
					return nil
				}
				if err := ge.AdvanceAge(); err == nil {
					st = ge.GetState()
				}
			}
			if stopAge != "" && idx[st.Age] > idx[stopAge] {
				return nil
			}
			bot.Play(st)
			if ticks%300 == 0 {
				curious(ge, st)
			}
		}
		sim += ge.StepTicks(1)
	}
	// Out of budget: keep whatever age we are in.
	return writeState(dir, ge.GetState().Age, ge.GetState())
}

// curious is the part of a player the greedy bot leaves out: it sends a
// scouting party out when none is away and opens any trade route it can. Both
// go through the same engine calls the expedition and trade commands use, so
// the states stay real; they just have a gate and a harbour worth walking to.
func curious(ge *game.GameEngine, st game.GameState) {
	if st.Military.ActiveScout == nil {
		for _, e := range st.Military.Expeditions {
			if e.Category == "scouting" && e.CanLaunch {
				if ge.LaunchExpedition(e.Key) == nil {
					break
				}
			}
		}
	}
	for _, r := range st.Trade.AvailableRoutes {
		if r.CanStart {
			_ = ge.StartTradeRoute(r.Key)
		}
	}
}

func writeState(dir, name string, st game.GameState) error {
	st.History = nil // large and not used by the renderer
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, name+".json.gz"))
	if err != nil {
		return err
	}
	zw, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	if _, err := zw.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func loadState(dir, name string) (game.GameState, error) {
	var st game.GameState
	f, err := os.Open(filepath.Join(dir, name+".json.gz"))
	if err != nil {
		return st, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return st, err
	}
	b, err := io.ReadAll(zr)
	if err != nil {
		return st, err
	}
	err = json.Unmarshal(b, &st)
	return st, err
}
