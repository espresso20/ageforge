package main

import (
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

// MapView is the narrow contract between the engine and the glyph map. It is
// built from a game.GameState and holds only what the map draws, so the
// renderer never touches the engine and a snapshot can be cached as JSON.
type MapView struct {
	Seed      int64
	Tick      int
	Age       string
	AgeName   string
	AgeIndex  int
	Epoch     string
	EpochName string

	Buildings []BuildingView
	Pop       int
	MaxPop    int
	Idle      int

	Factions []FactionView
	Routes   []RouteView

	ScoutActive    string // name of the active scouting expedition, "" if none
	MilitaryActive string
	Expeditions    int // completed expeditions: lifts the fog

	PendingCatastrophe string // epoch key, "" if none
	CatastropheName    string
	Harbinger          string // harbinger name, "" if none
	Morale             float64
	Survived           int // catastrophes endured
	Succumbed          int

	// Prev is the building counts at the player's previous check-in, so the
	// map can mark what changed while they were away.
	Prev map[string]int
}

// BuildingView is one building type the player owns.
type BuildingView struct {
	Key      string
	Name     string
	Lineage  string
	Category string
	Tier     int
	Count    int
	Ruins    int
	Workers  int
	WorkCap  int
	Legacy   bool
	Output   string  // primary output resource
	Rate     float64 // approx. per-second output of all instances
	RateName string
}

// FactionView is an NPC civilisation.
type FactionView struct {
	Key        string
	Name       string
	Discovered bool
	Status     string
	Opinion    int
	AtWar      bool
	Strength   int
	TradeCount int
	Specialty  string
}

// RouteView is an active trade route.
type RouteView struct {
	Key       string
	Name      string
	Disrupted bool
	Imports   []string
}

// BuildView converts an engine snapshot into the map's view.
func BuildView(st game.GameState) MapView {
	idx := map[string]int{}
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}
	v := MapView{
		Seed: st.Seed, Tick: st.Tick, Age: st.Age, AgeName: st.AgeName, AgeIndex: idx[st.Age],
		Epoch: st.EpochKey, EpochName: st.EpochName,
		Pop: st.Workers.TotalPop, MaxPop: st.Workers.MaxPop, Idle: st.Workers.TotalIdle,
		Expeditions:        st.Military.CompletedCount,
		PendingCatastrophe: st.PendingCatastrophe,
		Morale:             st.Morale,
		Survived:           st.CatastrophesEndured,
		Succumbed:          st.CatastrophesSuccumbed,
	}
	if v.Epoch == "" {
		v.Epoch = config.EpochForAge(st.Age)
	}
	if st.PendingCatastrophe != "" {
		v.CatastropheName, _ = config.CatastropheInfo(st.PendingCatastrophe)
	}
	if st.Harbinger != nil {
		v.Harbinger = st.Harbinger.Name
	}
	if s := st.Military.ActiveScout; s != nil {
		v.ScoutActive = s.Name
	}
	if s := st.Military.ActiveMilitary; s != nil {
		v.MilitaryActive = s.Name
	}
	defs := config.BuildingByKey()

	// Share each resource's building rate across its producers by their
	// base production, so a tile can say "+3.2 iron/s".
	weight := map[string]float64{}
	base := func(d config.BuildingDef, res string) float64 {
		t := 0.0
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == res {
				t += e.Value
			}
		}
		return t
	}
	for k, b := range st.Buildings {
		if b.Count == 0 {
			continue
		}
		d := defs[k]
		if d.OutputResource != "" {
			weight[d.OutputResource] += base(d, d.OutputResource) * float64(b.Count)
		}
	}
	keys := make([]string, 0, len(st.Buildings))
	for k := range st.Buildings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b := st.Buildings[k]
		if b.Count == 0 && b.RuinCount == 0 {
			continue
		}
		d := defs[k]
		bv := BuildingView{
			Key: k, Name: b.Name, Lineage: d.LineageKey, Category: b.Category, Tier: d.LineageTier,
			Count: b.Count, Ruins: b.RuinCount, Workers: b.WorkersAssigned,
			WorkCap: b.WorkerCapacity * b.Count, Legacy: b.IsLegacy, Output: d.OutputResource,
		}
		if bv.Lineage == "" {
			bv.Lineage = b.Category
		}
		if r, ok := st.Resources[d.OutputResource]; ok && weight[d.OutputResource] > 0 {
			share := base(d, d.OutputResource) * float64(b.Count) / weight[d.OutputResource]
			bv.Rate = r.Breakdown.BuildingRate * share / game.BaseTickInterval.Seconds()
			bv.RateName = r.Name
		}
		v.Buildings = append(v.Buildings, bv)
	}
	fkeys := make([]string, 0, len(st.Diplomacy.Factions))
	for k := range st.Diplomacy.Factions {
		fkeys = append(fkeys, k)
	}
	sort.Strings(fkeys)
	for _, k := range fkeys {
		f := st.Diplomacy.Factions[k]
		v.Factions = append(v.Factions, FactionView{
			Key: k, Name: f.Name, Discovered: f.Discovered, Status: f.Status, Opinion: f.Opinion,
			AtWar: f.AtWar, Strength: f.Strength, Specialty: f.Specialty, TradeCount: f.TradeCount,
		})
	}
	for _, r := range st.Trade.ActiveRoutes {
		rv := RouteView{Key: r.Key, Name: r.Name, Disrupted: r.Disrupted}
		for res := range r.Import {
			rv.Imports = append(rv.Imports, res)
		}
		sort.Strings(rv.Imports)
		v.Routes = append(v.Routes, rv)
	}
	return v
}

// generate plays a fresh engine with the smoke bot until target age is
// entered, then keeps playing `linger` of simulated time inside it so the
// age has buildings of its own. Catastrophes are endured.
func generate(target string, seed int64, linger time.Duration) (MapView, error) {
	order := config.AgeOrder()
	ti := -1
	for i, k := range order {
		if k == target {
			ti = i
		}
	}
	if ti < 0 {
		return MapView{}, fmt.Errorf("unknown age %q", target)
	}
	ge := game.NewGameEngine()
	ge.SeedRNG(seed)
	bot := smoke.NewBot(ge)
	bot.HorizonTicks = (30 * time.Minute).Seconds() / game.BaseTickInterval.Seconds()
	var sim time.Duration
	var entered time.Duration = -1
	var prev map[string]int
	for ticks := 0; sim < 3000*time.Hour; ticks++ {
		if ticks%5 == 0 {
			st := ge.GetState()
			if st.Age == target && entered < 0 {
				entered = sim
			}
			if entered >= 0 && prev == nil && sim-entered >= linger/2 {
				prev = map[string]int{}
				for k, b := range st.Buildings {
					if b.Count > 0 {
						prev[k] = b.Count
					}
				}
			}
			if entered >= 0 && sim-entered >= linger {
				v := BuildView(st)
				v.Prev = prev
				return v, nil
			}
			if st.PendingCatastrophe != "" {
				_ = ge.Endure()
				st = ge.GetState()
			}
			if st.AgeReady && st.Age != target {
				_ = ge.AdvanceAge()
				st = ge.GetState()
			}
			bot.Play(st)
			if ticks%50 == 0 {
				explore(ge, st)
			}
		}
		sim += ge.StepTicks(1)
	}
	return MapView{}, fmt.Errorf("did not reach %s within the sim budget", target)
}

func stateFile(dir, age string, seed int64) string {
	if seed == 7 {
		return filepath.Join(dir, age+".json")
	}
	return filepath.Join(dir, fmt.Sprintf("%s_seed%d.json", age, seed))
}

func loadView(dir, age string, seed int64) (MapView, error) {
	var v MapView
	b, err := os.ReadFile(stateFile(dir, age, seed))
	if err != nil {
		return v, err
	}
	return v, json.Unmarshal(b, &v)
}

func saveView(dir string, v MapView, seed int64) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile(dir, v.Age, seed), b, 0o644)
}
