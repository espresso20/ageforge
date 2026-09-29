// Package fixture builds synthetic, deterministic game states for map tests:
// any age, any seed, as built-up as asked, without running the engine. The
// captures use real smoke-bot states instead (see ui/mapstyle/capture); these
// are for the fast checks (every age at every size, determinism, placement
// stability, traffic density).
package fixture

import (
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// Options shape a synthetic state.
type Options struct {
	Age   string
	Seed  int64
	Tick  int     // 0: a tick that suits the age
	Scale float64 // building density; 0 means 1
	// Routes is how many trade routes run (-1: none; 0: two, when any exist).
	Routes      int
	Wars        int // civs at war with you
	Harbinger   bool
	Catastrophe bool
	Ruins       bool
	Idle        int // idle workers (0: a handful)
}

// State returns a synthetic snapshot.
func State(o Options) game.GameState {
	if o.Scale == 0 {
		o.Scale = 1
	}
	order := config.AgeOrder()
	ageIdx := map[string]int{}
	for i, a := range order {
		ageIdx[a] = i
	}
	ai, ok := ageIdx[o.Age]
	if !ok {
		o.Age, ai = order[0], 0
	}
	ages := config.AgeByKey()
	st := game.GameState{
		Seed: o.Seed, Tick: o.Tick, Age: o.Age, AgeName: ages[o.Age].Name,
		Buildings: map[string]game.BuildingState{}, Resources: map[string]game.ResourceState{},
		Diplomacy: game.DiplomacyState{Factions: map[string]game.FactionInfo{}},
	}
	if st.Tick == 0 {
		st.Tick = 1200 + ai*4700
	}
	if ai+1 < len(order) {
		st.NextAge, st.NextAgeName = order[ai+1], ages[order[ai+1]].Name
	}
	st.EpochKey = config.EpochForAge(o.Age)
	st.EpochName = config.EpochByKey()[st.EpochKey].Name
	h := func(k string, salt int64) uint64 { return mapmodel.Hash(o.Seed, mapmodel.HashStr(k), salt) }

	defs := config.BuildingByKey()
	keys := make([]string, 0, len(defs))
	for k := range defs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// the highest tier each lineage has reached
	top := map[string]int{}
	for _, k := range keys {
		d := defs[k]
		if ageIdx[d.RequiredAge] <= ai && d.LineageKey != "" && d.Category != "wonder" {
			if t, seen := top[d.LineageKey]; !seen || d.LineageTier > t {
				top[d.LineageKey] = d.LineageTier
			}
		}
	}
	pop, staffed := 0, 0
	for _, k := range keys {
		d := defs[k]
		da := ageIdx[d.RequiredAge]
		if da > ai {
			continue
		}
		bs := game.BuildingState{Name: d.Name, Category: d.Category, AgeKey: d.RequiredAge,
			WorkerCapacity: d.WorkerCapacity, Unlocked: true}
		if d.Category == "wonder" {
			if da < ai {
				bs.Count = 1
				st.Buildings[k] = bs
			} else if st.CurrentAgeWonderKey == "" {
				st.CurrentAgeWonderKey, st.CurrentAgeWonderName = k, d.Name
				bs.NextCost = map[string]float64{"gold": 1000, "stone": 500}
				bs.WonderBank = map[string]float64{"gold": 600, "stone": 100}
				st.Buildings[k] = bs
			}
			continue
		}
		if h(k, 1)%6 == 0 && da < ai {
			continue // not every player builds everything
		}
		legacy := d.LineageKey != "" && d.LineageTier < top[d.LineageKey] && d.Category != "storage"
		n := 2 + int(h(k, 2)%9)
		if da == ai {
			n = 1 + int(h(k, 3)%6)
		}
		if d.Category == "storage" {
			n = 1 + int(h(k, 4)%3)
		}
		if d.Category == "housing" {
			n += 4
		}
		n = int(float64(n)*o.Scale + 0.5)
		if n < 1 {
			n = 1
		}
		if d.MaxCount > 0 && n > d.MaxCount {
			n = d.MaxCount
		}
		bs.Count, bs.IsLegacy = n, legacy
		if d.WorkerCapacity > 0 {
			fill := [...]float64{1, 1, 0.9, 0.75, 0.5, 0.3, 0}[h(k, 5)%7]
			bs.WorkersAssigned = int(float64(n*d.WorkerCapacity)*fill + 0.5)
			staffed += bs.WorkersAssigned
		}
		if o.Ruins && h(k, 6)%5 == 0 {
			bs.RuinCount = 1 + int(h(k, 7)%3)
		}
		st.Buildings[k] = bs
	}
	idle := o.Idle
	if idle == 0 {
		idle = int(mapmodel.Hash(o.Seed, 11) % 7)
	}
	pop = staffed + idle
	st.Workers = game.WorkerState{TotalPop: pop, MaxPop: pop + 20 + ai*30, TotalIdle: idle}

	for i, r := range config.BaseResources() {
		if ageIdx[r.Age] > ai {
			continue
		}
		storage := r.BaseStorage * float64(10+ai*20)
		fill := float64(mapmodel.Hash(o.Seed, int64(i), 21)%100) / 100
		rate := 1.0 + float64(i%5)
		if i%7 == 3 {
			fill = 1 // a full store
		}
		if i%11 == 5 {
			rate, fill = -storage/10, 0.05 // one draining
		}
		st.Resources[r.Key] = game.ResourceState{Name: r.Name, Amount: storage * fill, Storage: storage, Rate: rate,
			Unlocked: true, Breakdown: game.RateBreakdown{BuildingRate: rate}}
	}

	wars := o.Wars
	statuses := []string{"friendly", "neutral", "allied", "rival", "neutral", "embargo"}
	for i, f := range config.BaseFactions() {
		fi := game.FactionInfo{Name: f.Name, Specialty: f.Specialty, Personality: f.Personality, Strength: f.Strength,
			Discovered: ageIdx[f.MinAge] <= ai, Status: statuses[i%len(statuses)], Opinion: 40 + i*5,
			TradeCount: i % 4}
		if fi.Discovered && wars > 0 && fi.Status != "allied" {
			fi.AtWar = true
			wars--
		}
		st.Diplomacy.Factions[f.Key] = fi
	}

	n := o.Routes
	if n == 0 {
		n = 2
	}
	for _, r := range config.BaseTradeRoutes() {
		if n <= 0 {
			break
		}
		if ageIdx[r.MinAge] > ai {
			continue
		}
		st.Trade.ActiveRoutes = append(st.Trade.ActiveRoutes, game.ActiveRouteInfo{Name: r.Name, Key: r.Key,
			Import: r.Import, Export: r.Export, CyclesDone: 3})
		n--
	}

	st.Military = game.MilitaryState{SoldierCount: int(float64(8*(ai+1)) * o.Scale), SoldierCap: 20 * (ai + 1),
		CompletedCount: ai}
	if ai >= 2 && o.Seed%2 == 0 {
		st.Military.ActiveScout = &game.ExpeditionSnapshot{Name: "Scout the hills", TicksLeft: 40}
	}
	if o.Harbinger {
		if hd, ok := config.HarbingerFor(o.Age); ok {
			st.Harbinger = &game.HarbingerView{Key: hd.Key, Name: hd.Name, Age: o.Age, AgeName: st.AgeName,
				Tier: game.CatastropheTierMedium, Probability: 0.2, TargetEpochName: "the next era"}
		}
	}
	if o.Catastrophe {
		st.PendingCatastrophe = st.EpochKey
	}
	st.BuildQueue = []game.BuildQueueSnapshot{{Name: "Scaffold", TicksLeft: 10, TotalTicks: 30}}
	return st
}

// Grow returns a copy of st with extra copies of existing and new buildings,
// as a player's next hour would add them: more of this age's types, and a
// type or two not built before. Legacy types never grow (the game forbids
// it). Used by the placement stability tests.
func Grow(st game.GameState, step int) game.GameState {
	out := st
	out.Buildings = make(map[string]game.BuildingState, len(st.Buildings))
	for k, v := range st.Buildings {
		out.Buildings[k] = v
	}
	defs := config.BuildingByKey()
	ageIdx := map[string]int{}
	for i, a := range config.AgeOrder() {
		ageIdx[a] = i
	}
	ai := ageIdx[st.Age]
	keys := make([]string, 0, len(defs))
	for k := range defs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		d := defs[k]
		if ageIdx[d.RequiredAge] != ai || d.Category == "wonder" {
			continue
		}
		bs, had := out.Buildings[k]
		if bs.IsLegacy {
			continue
		}
		add := int(mapmodel.Hash(st.Seed, mapmodel.HashStr(k), int64(step)) % 4)
		if !had {
			if mapmodel.Hash(st.Seed, mapmodel.HashStr(k), int64(step), 9)%3 != 0 {
				continue
			}
			bs = game.BuildingState{Name: d.Name, Category: d.Category, AgeKey: d.RequiredAge,
				WorkerCapacity: d.WorkerCapacity, Unlocked: true}
			add++
		}
		if d.MaxCount > 0 && bs.Count+add > d.MaxCount {
			add = d.MaxCount - bs.Count
		}
		bs.Count += add
		out.Buildings[k] = bs
	}
	out.Tick += 600
	return out
}
