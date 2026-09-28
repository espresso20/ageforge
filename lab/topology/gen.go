package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/smoke"
)

// Fixture is one real game state produced by the smoke bot, plus the short
// history the topology view needs for sparklines. A production widget would
// keep the same ring buffer itself by sampling GetState on each refresh.
type Fixture struct {
	Seed    int64
	Age     string
	Ticks   int
	SimSecs float64
	State   game.GameState
	// Prev is the state about checkInTicks earlier (the player's last look),
	// for the "since you last looked" diff. Nil in very short runs.
	Prev *game.GameState
	// Scenario notes player actions layered on the bot's run, "" for none.
	Scenario string
	// Hist holds up to histLen samples (oldest first), one every histEvery
	// ticks: "rate:<res>" net rate, "fill:<res>" amount/storage,
	// "lin:<lineage>" the lineage's attributed output.
	Hist map[string][]float64
}

const (
	histLen      = 48
	histEvery    = 10
	checkInTicks = 300 // ten minutes at 1x
)

// generate plays seed with the smoke bot until it has spent dwell ticks in
// target (never advancing past it) and returns the final state. With
// holdCatastrophe the run stops as soon as a catastrophe is pending in the
// target age, before the bot answers it, so the alarm state can be rendered.
func generate(target string, seed int64, dwell int, holdCatastrophe, incident bool, logf func(string, ...any)) (*Fixture, error) {
	var prevs []*game.GameState
	order := map[string]int{}
	for i, k := range config.AgeOrder() {
		order[k] = i
	}
	ti, ok := order[target]
	if !ok {
		return nil, fmt.Errorf("unknown age %q", target)
	}
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := smoke.NewBot(ge)
	fx := &Fixture{Seed: seed, Age: target, Hist: map[string][]float64{}}
	var sim time.Duration
	ticks, inTarget := 0, 0
	lastAge := ""
	const decide = 5
	for {
		st := ge.GetState()
		if st.Age != lastAge {
			logf("  tick %8d  %s", ticks, st.AgeName)
			lastAge = st.Age
		}
		if st.LastPassage.Pending && st.PendingCatastrophe == "" {
			_ = ge.EndureLastPassage()
			st = ge.GetState()
		}
		if st.PendingCatastrophe != "" {
			if holdCatastrophe && order[st.Age] >= ti {
				fx.State = st
				break
			}
			if err := ge.Endure(); err != nil {
				return nil, fmt.Errorf("endure: %w", err)
			}
			st = ge.GetState()
		}
		if st.AgeReady && order[st.Age] < ti {
			if err := ge.AdvanceAge(); err == nil {
				st = ge.GetState()
			}
		}
		if order[st.Age] > ti {
			return nil, fmt.Errorf("overshot %s into %s", target, st.Age)
		}
		if order[st.Age] == ti {
			if inTarget >= dwell && !holdCatastrophe {
				fx.State = st
				break
			}
			// Keep the state from about checkInTicks ago: "since you last
			// looked" is the diff between it and the final state.
			if inTarget%100 == 0 {
				s := st
				prevs = append(prevs, &s)
				if len(prevs) > checkInTicks/100+1 {
					prevs = prevs[1:]
				}
			}
			inTarget += decide
		}
		if ticks > 6_000_000 {
			return nil, fmt.Errorf("gave up at %s after %d ticks", st.Age, ticks)
		}
		bot.Play(st)
		for i := 0; i < decide; i++ {
			sim += ge.StepTicks(1)
			ticks++
			if ticks%histEvery == 0 {
				sample(fx, ge.GetState())
			}
		}
	}
	if len(prevs) > 0 {
		fx.Prev = prevs[0]
	}
	if incident {
		// Real player actions on top of the bot's run: raid a discovered
		// civ's route until it declares war (the `diplomacy raid` command),
		// then the dev console's /catastrophe. A few ticks let the blockade
		// reach the trade routes.
		for _, key := range sortedMapKeys(fx.State.Diplomacy.Factions) {
			f := fx.State.Diplomacy.Factions[key]
			if !f.Discovered {
				continue
			}
			for i := 0; i < 8 && !ge.GetState().Diplomacy.Factions[key].AtWar; i++ {
				if err := ge.RaidCivRoute(key); err != nil {
					logf("  raid %s: %v", key, err)
					break
				}
			}
			if ge.GetState().Diplomacy.Factions[key].AtWar {
				logf("  %s declared war", f.Name)
				break
			}
		}
		for i := 0; i < 6; i++ {
			bot.Play(ge.GetState())
			sim += ge.StepTicks(decide)
			ticks += decide
		}
		if err := ge.ForceCatastropheForTest(); err != nil {
			logf("  catastrophe: %v", err)
		}
		fx.State = ge.GetState()
		fx.Scenario = "incident: war declared by raiding, catastrophe forced"
	}
	fx.Ticks, fx.SimSecs = ticks, sim.Seconds()
	fx.State.Log = trimLog(fx.State.Log)
	if fx.Prev != nil {
		fx.Prev.Log = nil
		fx.Prev.History = nil
	}
	return fx, nil
}

func trimLog(l []game.LogEntry) []game.LogEntry {
	if len(l) > 40 {
		return l[len(l)-40:]
	}
	return l
}

// sample appends one history point, keeping the last histLen.
func sample(fx *Fixture, st game.GameState) {
	push := func(k string, v float64) {
		h := append(fx.Hist[k], v)
		if len(h) > histLen {
			h = h[len(h)-histLen:]
		}
		fx.Hist[k] = h
	}
	for k, r := range st.Resources {
		if !r.Unlocked {
			continue
		}
		push("rate:"+k, r.Rate)
		if r.Storage > 0 {
			push("fill:"+k, r.Amount/r.Storage)
		}
	}
	for lin, v := range lineageOutput(st) {
		push("lin:"+lin, v)
	}
}

func saveFixture(fx *Fixture, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	if err := json.NewEncoder(zw).Encode(fx); err != nil {
		f.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func loadFixture(path string) (*Fixture, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	var fx Fixture
	if err := json.NewDecoder(zr).Decode(&fx); err != nil {
		return nil, err
	}
	return &fx, nil
}
