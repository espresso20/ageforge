package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/smoke"
)

// states.go builds the real game states the prototype renders. The smoke
// bot (the same greedy autoplayer the pacing suite uses) plays a seeded
// engine headlessly; at the last decision before each age advance, when the
// age is as built-up as it will get, the GameState snapshot is written to
// states/<age>.json.gz. Harbinger sightings and pending catastrophes are
// saved as extra states so the event overlays are drawn from real data too.

const stateDir = "lab/skyline/states"

// trimState drops the parts of a snapshot the renderer never reads and that
// would bloat the file (the log, the history collector, account stats).
func trimState(st game.GameState) game.GameState {
	st.History = nil
	st.AccountStats = nil
	st.Modifiers = nil
	if len(st.Log) > 12 {
		st.Log = st.Log[len(st.Log)-12:]
	}
	return st
}

func writeState(dir, name string, st game.GameState) error {
	st = trimState(st)
	f, err := os.Create(filepath.Join(dir, name+".json.gz"))
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	enc := json.NewEncoder(zw)
	if err := enc.Encode(st); err != nil {
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
	err = json.NewDecoder(zr).Decode(&st)
	return st, err
}

// availableStates lists the saved state names, in age order first.
func availableStates(dir string) []string {
	m, _ := filepath.Glob(filepath.Join(dir, "*.json.gz"))
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	var out []string
	for _, p := range m {
		b := filepath.Base(p)
		out = append(out, b[:len(b)-len(".json.gz")])
	}
	sort.Slice(out, func(i, j int) bool {
		oi, iok := order[out[i]]
		oj, jok := order[out[j]]
		if iok && jok {
			return oi < oj
		}
		if iok != jok {
			return iok
		}
		return out[i] < out[j]
	})
	return out
}

// generate plays seed up to stopAge, saving one state per age plus event
// states. It mirrors the smoke runner's control loop (endure catastrophes,
// advance when ready) without its invariant bookkeeping.
func generate(dir string, seed int64, stopAge string, maxHours float64) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := smoke.NewBot(ge)
	bot.Harbinger = smoke.HarbingerIgnore

	idx := map[string]int{}
	for i, a := range config.AgeOrder() {
		idx[a] = i
	}
	var sim time.Duration
	budget := time.Duration(maxHours * float64(time.Hour))
	harbSaved := map[string]bool{}
	earlySaved := false
	enteredStop := 0
	forced := false
	lastAge, ageT0 := "", 0
	// visit is a rolling snapshot about two hours of play old: saved next to
	// each age as prev_<age>, it stands in for "your last check-in" so the
	// renderer can mark what changed since.
	var visit, pending game.GameState
	const visitEvery = 3600
	catSaved := map[string]bool{}
	var last game.GameState
	started := time.Now()
	for tick := 0; sim < budget; tick++ {
		if tick%5 == 0 {
			st := ge.GetState()
			last = st
			if tick%visitEvery == 0 {
				visit, pending = pending, st
			}
			if !earlySaved && totalBuildings(st) >= 3 {
				earlySaved = true
				if err := writeState(dir, "primitive_early", st); err != nil {
					return err
				}
			}
			if h := st.Harbinger; h != nil && !harbSaved[st.EpochKey] {
				harbSaved[st.EpochKey] = true
				if err := writeState(dir, "harbinger_"+st.Age, st); err != nil {
					return err
				}
				fmt.Printf("  harbinger %q at %s (tick %d)\n", h.Name, st.Age, st.Tick)
			}
			// Catastrophes are rare; to have the Digital Era's collapse on
			// record the generator calls the dev console's force hook once,
			// halfway through the Information Age (the passage then plays out
			// exactly as a rolled one would).
			if st.Age == "information_age" && !forced && st.Tick-ageT0 > 6000 {
				forced = ge.ForceCatastropheForTest() == nil
				st = ge.GetState()
			}
			if st.Age != lastAge {
				lastAge, ageT0 = st.Age, st.Tick
			}
			if st.PendingCatastrophe != "" {
				if !catSaved[st.PendingCatastrophe] && len(catSaved) < 3 {
					catSaved[st.PendingCatastrophe] = true
					if err := writeState(dir, "catastrophe_"+st.Age, st); err != nil {
						return err
					}
					fmt.Printf("  catastrophe pending at %s (tick %d)\n", st.Age, st.Tick)
				}
				_ = ge.Endure()
				st = ge.GetState()
			}
			// the last age never becomes "ready": settle in for a while
			// (about two hours of play) and keep it as it stands
			if st.Age == stopAge && !st.AgeReady {
				if enteredStop == 0 {
					enteredStop = tick
				} else if tick-enteredStop > 3600 {
					fmt.Printf("%-18s tick %7d  sim %6.1fh  (settled)\n", st.Age, st.Tick, sim.Hours())
					if visit.Age != "" {
						_ = writeState(dir, "prev_"+st.Age, visit)
					}
					return writeState(dir, st.Age, st)
				}
			}
			if st.AgeReady {
				if err := writeState(dir, st.Age, st); err != nil {
					return err
				}
				if visit.Age != "" {
					if err := writeState(dir, "prev_"+st.Age, visit); err != nil {
						return err
					}
				}
				fmt.Printf("%-18s tick %7d  sim %6.1fh  wall %s  bld %d\n", st.Age, st.Tick, sim.Hours(),
					time.Since(started).Round(time.Second), totalBuildings(st))
				if st.Age == stopAge {
					return nil
				}
				if err := ge.AdvanceAge(); err == nil {
					st = ge.GetState()
				}
			}
			bot.Play(st)
			// The greedy bot never opens trade routes; a real player does,
			// and the skyline draws them, so open whatever can start (up
			// to three at a time), the same StartTradeRoute the command
			// line reaches.
			if tick%50 == 0 && len(st.Trade.ActiveRoutes) < 3 {
				for _, r := range st.Trade.AvailableRoutes {
					if r.CanStart && ge.StartTradeRoute(r.Key) == nil {
						break
					}
				}
			}
		}
		sim += ge.StepTicks(1)
	}
	// Out of budget: keep whatever age we are in, as it stands.
	fmt.Printf("budget spent in %s; saving it as is\n", last.Age)
	return writeState(dir, last.Age, last)
}

func totalBuildings(st game.GameState) int {
	n := 0
	for _, b := range st.Buildings {
		n += b.Count
	}
	return n
}
