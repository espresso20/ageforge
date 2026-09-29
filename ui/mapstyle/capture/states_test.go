//go:build mapcapture

package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/smoke"
)

// TestGenerateStates plays the smoke bot (the greedy autoplayer the pacing
// suite uses) through a seeded engine and saves real snapshots for the map
// captures: one per age at the moment it is ready to advance (as built-up
// as that age gets), prev_<age> from about two hours of play earlier (the
// "last visit" baseline), primitive_early (three buildings in), and the
// first harbinger and catastrophe states. Like a real player, and unlike the
// bot, it sends scouts out and opens trade routes. It takes about ten
// minutes to the Transcendent Age.
//
//	go test -tags mapcapture -run TestGenerateStates -timeout 60m ./ui/mapstyle/capture
//
// MAP_STATES_DIR picks the folder (default map_captures/states at the repo
// root), MAP_STOP_AGE the last age.
func TestGenerateStates(t *testing.T) {
	dir := os.Getenv("MAP_STATES_DIR")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "map_captures", "states")
	}
	stop := os.Getenv("MAP_STOP_AGE")
	if stop == "" {
		stop = "transcendent_age"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	defer game.SetDataDirForTest(t.TempDir())()
	save := func(name string, st game.GameState) {
		if err := SaveState(filepath.Join(dir, name+".json.gz"), st); err != nil {
			t.Fatal(err)
		}
	}
	ge := game.NewGameEngine()
	ge.SeedRNG(7)
	bot := smoke.NewBot(ge)
	bot.Harbinger = smoke.HarbingerIgnore
	var sim time.Duration
	var visit, pending game.GameState
	const visitEvery = 3600
	early, forced := false, false
	harb, cat := map[string]bool{}, map[string]bool{}
	lastAge, ageT0, settled := "", 0, 0
	started := time.Now()
	for tick := 0; sim < 4000*time.Hour; tick++ {
		if tick%5 == 0 {
			st := ge.GetState()
			if tick%visitEvery == 0 {
				visit, pending = pending, st
			}
			if st.Age != lastAge {
				lastAge, ageT0 = st.Age, st.Tick
			}
			n := 0
			for _, b := range st.Buildings {
				n += b.Count
			}
			if !early && n >= 3 {
				early = true
				save("primitive_early", st)
			}
			if st.Harbinger != nil && !harb[st.EpochKey] {
				harb[st.EpochKey] = true
				save("harbinger_"+st.Age, st)
			}
			// Catastrophes are rare: force one, halfway through the
			// Information Age, through the dev hook (it then plays out as a
			// rolled one would).
			if st.Age == "information_age" && !forced && st.Tick-ageT0 > 6000 {
				forced = ge.ForceCatastropheForTest() == nil
				st = ge.GetState()
			}
			if st.PendingCatastrophe != "" {
				if !cat[st.PendingCatastrophe] {
					cat[st.PendingCatastrophe] = true
					save("catastrophe_"+st.Age, st)
				}
				_ = ge.Endure()
				st = ge.GetState()
			}
			if st.Age == stop && !st.AgeReady {
				if settled == 0 {
					settled = tick
				} else if tick-settled > 3600 {
					save(st.Age, st)
					if visit.Age != "" {
						save("prev_"+st.Age, visit)
					}
					t.Logf("%s settled at tick %d (%s)", st.Age, st.Tick, time.Since(started).Round(time.Second))
					return
				}
			}
			if st.AgeReady {
				save(st.Age, st)
				if visit.Age != "" {
					save("prev_"+st.Age, visit)
				}
				t.Logf("%-18s tick %7d sim %6.1fh wall %s buildings %d", st.Age, st.Tick, sim.Hours(),
					time.Since(started).Round(time.Second), n)
				if st.Age == stop {
					return
				}
				if err := ge.AdvanceAge(); err == nil {
					st = ge.GetState()
				}
			}
			bot.Play(st)
			if tick%50 == 0 {
				explore(ge, st)
			}
		}
		sim += ge.StepTicks(1)
	}
	t.Fatal(fmt.Sprintf("did not reach %s within the budget", stop))
}

// explore does what the bot skips and a player does: sends scouts out and
// keeps up to three trade routes running.
func explore(ge *game.GameEngine, st game.GameState) {
	if st.Military.ActiveScout == nil {
		for _, e := range st.Military.Expeditions {
			if e.Category == "scouting" && e.CanLaunch {
				_ = ge.LaunchExpedition(e.Key)
				break
			}
		}
	}
	if len(st.Trade.ActiveRoutes) < 3 {
		for _, r := range st.Trade.AvailableRoutes {
			if r.CanStart && ge.StartTradeRoute(r.Key) == nil {
				break
			}
		}
	}
}
