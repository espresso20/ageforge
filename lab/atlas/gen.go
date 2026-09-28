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

// Snapshot is one real game state captured from a headless smoke-bot run,
// plus the expedition journal the driver kept (the engine keeps only a count
// of finished expeditions, so the atlas's trails need this; see DESIGN.md,
// "engine additions").
type Snapshot struct {
	Seed    int64
	Age     string
	Ticks   int
	State   game.GameState
	Journal []JournalEntry
}

// JournalEntry is one expedition the driver launched.
type JournalEntry struct {
	Key      string
	Name     string
	Category string
	Tick     int
	Age      string
}

// snapshotAges are the ages the generator keeps: at least one per epoch,
// two where the atlas changes its style mid-epoch.
var snapshotAges = map[string]bool{
	"primitive_age": true, "stone_age": true, "bronze_age": true,
	"iron_age": true, "medieval_age": true,
	"renaissance_age": true, "industrial_age": true,
	"victorian_age": true, "atomic_age": true,
	"modern_age": true, "digital_age": true,
	"cyberpunk_age": true, "space_age": true,
	"interstellar_age": true, "galactic_age": true, "quantum_age": true,
}

// generate plays seed with the smoke bot (greedy style, deals on, scouting
// expeditions launched whenever the slot is free) and writes a snapshot of
// each kept age at the moment that age is complete (just before advancing),
// which is the fullest picture of it. Stops after stopAge.
func generate(dir string, seed int64, stopAge string, maxHours float64) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := smoke.NewBot(ge)
	bot.Deals = true
	bot.Harbinger = smoke.HarbingerIgnore

	var journal []JournalEntry
	var sim time.Duration
	start := time.Now()
	maxSim := time.Duration(maxHours * float64(time.Hour))
	ticks := 0
	save := func(st game.GameState) error {
		if !snapshotAges[st.Age] {
			return nil
		}
		trim(&st)
		snap := Snapshot{Seed: seed, Age: st.Age, Ticks: ticks, State: st, Journal: append([]JournalEntry(nil), journal...)}
		path := filepath.Join(dir, st.Age+".json.gz")
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		zw := gzip.NewWriter(f)
		enc := json.NewEncoder(zw)
		if err := enc.Encode(snap); err != nil {
			f.Close()
			return err
		}
		zw.Close()
		fmt.Printf("  saved %-18s tick %7d  sim %6.1fh  wall %s  civs %d  routes %d  exp %d\n", st.Age, ticks,
			sim.Hours(), time.Since(start).Round(time.Second), discovered(st), len(st.Trade.ActiveRoutes), len(journal))
		return f.Close()
	}
	for sim < maxSim {
		if ticks%5 == 0 {
			st := ge.GetState()
			if st.LastPassage.Pending {
				_ = ge.EndureLastPassage()
				st = ge.GetState()
			}
			if st.PendingCatastrophe != "" {
				_ = ge.Endure()
				st = ge.GetState()
			}
			if st.AgeReady {
				if err := save(st); err != nil {
					return err
				}
				if st.Age == stopAge {
					return nil
				}
				if err := ge.AdvanceAge(); err == nil {
					st = ge.GetState()
				}
			}
			if st.Military.ActiveScout == nil || st.Military.ActiveMilitary == nil {
				for _, ex := range st.Military.Expeditions {
					if !ex.CanLaunch {
						continue
					}
					busy := (ex.Category == "scouting" && st.Military.ActiveScout != nil) ||
						(ex.Category != "scouting" && st.Military.ActiveMilitary != nil)
					// Military expeditions only when soldiers are plentiful, so
					// the driver does not starve the bot's own plan.
					if busy || (ex.Category != "scouting" && st.Military.SoldierCount < 3*ex.SoldiersNeeded) {
						continue
					}
					if ge.LaunchExpedition(ex.Key) == nil {
						journal = append(journal, JournalEntry{Key: ex.Key, Name: ex.Name, Category: ex.Category, Tick: ticks, Age: st.Age})
						st = ge.GetState()
					}
				}
			}
			if ticks%400 == 0 {
				statesman(ge, st, ticks)
			}
			bot.Play(st)
		}
		sim += ge.StepTicks(1)
		ticks++
	}
	return fmt.Errorf("seed %d did not reach %s within %.0fh simulated", seed, stopAge, maxHours)
}

// statesman is the part of a player the greedy bot does not play: it opens
// every trade route it can, courts the peaceful and mercantile civs, keeps a
// standing feud with the Ironhold Clans (a raid every 600 ticks, which is how
// wars start and stay started), and embargoes the isolationists. It exists so the fixtures carry routes, alliances,
// wars and embargoes for the atlas to draw.
func statesman(ge *game.GameEngine, st game.GameState, ticks int) {
	active := map[string]bool{}
	for _, r := range st.Trade.ActiveRoutes {
		active[r.Key] = true
	}
	for _, r := range st.Trade.AvailableRoutes {
		if r.CanStart && !active[r.Key] {
			_ = ge.StartTradeRoute(r.Key)
		}
	}
	for _, key := range sortedFactionKeys(st) {
		f := st.Diplomacy.Factions[key]
		if !f.Discovered {
			continue
		}
		switch f.Personality {
		case "peaceful", "mercantile":
			if f.Opinion >= 50 && f.Status != "allied" {
				_ = ge.SetDiplomaticStatus(key, "allied")
			} else if f.Opinion < 60 {
				_ = ge.SendGift(key)
			}
		case "aggressive":
			if key == "ironhold_clans" && ticks%600 == 0 { // a standing feud: keep provoking them
				_ = ge.RaidCivRoute(key)
			}
		case "isolationist":
			if f.Status == "neutral" && f.Opinion < 10 {
				_ = ge.SetDiplomaticStatus(key, "embargo")
			}
		}
	}
}

func sortedFactionKeys(st game.GameState) []string {
	var keys []string
	for k := range st.Diplomacy.Factions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func discovered(st game.GameState) int {
	n := 0
	for _, f := range st.Diplomacy.Factions {
		if f.Discovered {
			n++
		}
	}
	return n
}

// trim drops the parts of the snapshot the atlas never reads, to keep the
// committed fixtures small.
func trim(st *game.GameState) {
	if len(st.Log) > 60 {
		st.Log = st.Log[len(st.Log)-60:]
	}
	st.Research.Techs = nil
	st.Milestones.Milestones = nil
	st.Milestones.Chains = nil
	st.Trade.ExchangeRates = nil
	st.History = nil
	st.Plan = nil
}

// loadSnapshot reads one generated snapshot.
func loadSnapshot(dir, age string) (*Snapshot, error) {
	f, err := os.Open(filepath.Join(dir, age+".json.gz"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	var s Snapshot
	if err := json.NewDecoder(zr).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// availableAges lists the snapshot ages present in dir, in age order.
func availableAges(dir string) []string {
	var out []string
	for _, a := range config.AgeOrder() {
		if _, err := os.Stat(filepath.Join(dir, a+".json.gz")); err == nil {
			out = append(out, a)
		}
	}
	return out
}
