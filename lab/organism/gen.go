package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/smoke"
)

// checkinTicks is the gap between a saved state and its "last visit"
// companion (<age>_prev): an hour of play at 1x.
const checkinTicks = 1800

// generate plays seed with the smoke bot (the same legitimate player the
// smoke suite uses) and saves real games through the engine's own SaveGame,
// under dir (the engine puts them in dir/accounts/saves):
//
//   - <age>: the last decision before each advance, when the age's economy
//     is at its fullest;
//   - <age>_prev: the same run about an hour earlier, for "since your last
//     visit" (rotating checkpoints, so it is 1-2 hours back);
//   - primitive_seedling: the first moment the civ owns two buildings;
//   - <age>_catastrophe: the moment a catastrophe is pending, before the
//     bot endures it.
//
// The loop mirrors smoke's runner: decide every 5 ticks, endure every
// catastrophe, advance as soon as the age is ready, never prestige.
func generate(dir string, seed int64, stopAge string, logf func(string, ...any)) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	restore := game.SetDataDirForTest(abs)
	defer restore()
	saves := filepath.Join(abs, "accounts", "saves")

	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := smoke.NewBot(ge)
	bot.HorizonTicks = (30 * time.Minute).Seconds() / game.BaseTickInterval.Seconds()

	type ckpt struct {
		name, age string
		tick      int
	}
	slots := [2]ckpt{}
	slot := 0
	seedling := false
	catastrophe := map[string]bool{}
	curAge, ageT0 := "", 0

	started := time.Now()
	const maxTicks = 40_000_000
	for t := 0; t < maxTicks; t++ {
		if t%5 != 0 {
			ge.StepTicks(1)
			continue
		}
		st := ge.GetState()
		if st.Age != curAge {
			curAge, ageT0 = st.Age, t
		}
		// The last age never becomes ready: save it after a while there.
		if st.NextAge == "" && t-ageT0 >= 6000 {
			st.AgeReady = true
		}
		if !seedling && st.Stats.TotalBuilt >= 2 {
			seedling = true
			if err := ge.SaveGame("primitive_seedling"); err != nil {
				return err
			}
		}
		if st.PendingCatastrophe != "" {
			if !catastrophe[st.EpochKey] {
				catastrophe[st.EpochKey] = true
				if err := ge.SaveGame(st.Age + "_catastrophe"); err != nil {
					return err
				}
				logf("saved %-18s tick %9d  (catastrophe pending: %s)", st.Age+"_catastrophe", t, st.PendingCatastrophe)
			}
			_ = ge.Endure()
			st = ge.GetState()
		}
		if t > 0 && t%(checkinTicks/2) == 0 {
			name := fmt.Sprintf("_ckpt_%d", slot)
			if err := ge.SaveGame(name); err != nil {
				return err
			}
			slots[slot] = ckpt{name, st.Age, t}
			slot ^= 1
		}
		if st.AgeReady {
			if err := ge.SaveGame(st.Age); err != nil {
				return fmt.Errorf("save %s: %w", st.Age, err)
			}
			// The newest checkpoint at least checkinTicks back, in this age.
			best := -1
			for i, c := range slots {
				if c.name != "" && c.age == st.Age && t-c.tick >= checkinTicks/2 && (best < 0 || c.tick > slots[best].tick) {
					best = i
				}
			}
			if best >= 0 {
				data, err := os.ReadFile(filepath.Join(saves, slots[best].name+".json"))
				if err == nil {
					err = os.WriteFile(filepath.Join(saves, st.Age+"_prev.json"), data, 0o644)
				}
				if err != nil {
					return err
				}
			}
			logf("saved %-18s tick %9d  built %5d  pop %5d  (%s wall)",
				st.Age, t, st.Stats.TotalBuilt, st.Workers.TotalPop, time.Since(started).Round(time.Second))
			if st.Age == stopAge {
				break
			}
			if err := ge.AdvanceAge(); err == nil {
				ge.StepTicks(1)
				continue
			}
		}
		bot.Play(st)
		ge.StepTicks(1)
	}
	for i := range slots {
		os.Remove(filepath.Join(saves, fmt.Sprintf("_ckpt_%d.json", i)))
	}
	return nil
}

// withTrade loads the save age, opens every trade route the player could
// open there (the bot never trades by route), plays a minute, and saves it
// as <age>_trade: the same route command a player types.
func withTrade(dir, age string, logf func(string, ...any)) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	restore := game.SetDataDirForTest(abs)
	defer restore()
	ge := game.NewGameEngine()
	if err := ge.LoadGame(age); err != nil {
		return err
	}
	opened := 0
	for _, rt := range ge.GetState().Trade.AvailableRoutes {
		if rt.CanStart && ge.StartTradeRoute(rt.Key) == nil {
			opened++
		}
	}
	ge.StepTicks(30)
	logf("%s: opened %d trade routes", age, opened)
	return ge.SaveGame(age + "_trade")
}
