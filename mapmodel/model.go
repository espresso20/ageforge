// Package mapmodel is the one map model every map style draws from.
//
// It is a pure function of (game snapshot, seed, tick): Build reads a
// game.GameState and returns a Model holding everything a renderer needs
// (buildings grouped by lineage and by age, stable placement for the town
// and the skyline, workers, wonders, trade routes, civilizations,
// expeditions, the harbinger and catastrophe pressure, the clock and the
// weather, the flows summary and the since-last-visit recap). A new game
// feature is modelled here once; styles stay thin.
//
// The model is deterministic across machines (it follows the float rules
// for simulation code and the smoke harness fingerprints it), cheap (a few
// milliseconds for a late-game state, the seed-derived world and town plan
// cached per seed by Builder) and never touches the engine.
package mapmodel

import (
	"sort"

	"github.com/espresso20/ageforge/game"
)

// Model is one snapshot laid out for the maps.
type Model struct {
	Seed int64
	Tick int

	Age, AgeName         string
	AgeIdx               int
	NextAge, NextAgeName string
	AgeReady             bool
	Epoch                int // epoch index (0 Stone … 6 Cosmic)
	EpochKey, EpochName  string

	Clock   Clock
	Weather Weather

	// Buildings lists every type the player owns (or holds ruins of), in
	// lineage order then rank (oldest first).
	Buildings []*Building
	// Lineages groups them; only lineages with something built appear.
	Lineages []*Lineage

	Town    Town    // the roguelike town: world, plan and placed tiles
	Skyline Skyline // the panorama: one district per age, placed lots

	Wonders     []Wonder
	Workers     Workers
	Army        Army
	Routes      []Route
	Factions    []Faction // every civ, discovered or not, in site order
	Expeditions Expeditions
	Harbinger   *Harbinger
	Catastrophe Catastrophe
	Queue       []QueueItem
	Activity    Activity
	Flows       Flows
	Recap       Recap

	// LayoutKey hashes everything a style's static layer depends on (what
	// is placed, staffing, civs, routes, expeditions, the harbinger and
	// catastrophe) but not the tick, so a style can keep its scene while
	// only the clock moves.
	LayoutKey uint64

	Catalog *Catalog
	byKey   map[string]*Building
}

// Building is one owned building type.
type Building struct {
	*Def
	Count, Ruins int
	Workers      int // assigned across all instances
	Capacity     int // worker slots across all instances
	// Staffing is Workers/Capacity in [0,1], or -1 for a type that takes
	// no workers.
	Staffing float64
	Legacy   bool
	Upgrade  string // key of the next tier it can be upgraded to, "" if none
	// Rate is the approximate output of all instances per second, of
	// RateName; 0 when the type produces nothing measurable.
	Rate     float64
	RateName string
	Delta    int // count change since the last visit (0 without a baseline)
	// Class and Sym are the colour class and symbol the maps draw it with.
	Class Class
	Sym   Sym
}

// Understaffed reports a type that takes workers but has under half.
func (b *Building) Understaffed() bool { return b.Staffing >= 0 && b.Staffing < 0.5 && b.Count > 0 }

// Lineage is one lineage's buildings and totals.
type Lineage struct {
	Key, Name string
	Buildings []*Building // rank order
	Count     int         // all copies
	Legacy    int         // copies of legacy types
	Ruins     int
	Workers   int
	Capacity  int
	Share     float64 // share of the settlement's working output, 0..1
	Class     Class
	Sym       Sym
	// Tiles is how many town tiles the lineage draws (TownTiles).
	Tiles int
}

// Wonder is a built wonder or the current age's one in progress.
type Wonder struct {
	Key, Name string
	Age       int
	Built     bool
	Current   bool    // the wonder this age needs
	Progress  float64 // banked share of its cost, 0..1
	Delta     bool    // built since the last visit
}

// Workers is the population picture.
type Workers struct {
	Pop, MaxPop, Idle int
	Staffed           int // assigned to buildings
	Capacity          int // worker slots in all buildings
}

// Army is the military picture.
type Army struct {
	Soldiers, Cap int
	Defense       float64
}

// Route is an active trade route.
type Route struct {
	Key, Name string
	Disrupted bool
	Imports   []string
	Cycles    int
	// Mode is how its goods travel (land, sea or air), from the route data.
	Mode RouteMode
	// Civ is the civ the route's caravans travel to: a cosmetic,
	// deterministic pick among the civs you trade with ("" when you have
	// met none).
	Civ string
}

// Relation is a civ's standing toward the player.
type Relation uint8

const (
	RelNeutral Relation = iota
	RelFriendly
	RelAllied
	RelRival
	RelEmbargo
	RelWar
)

func (r Relation) String() string {
	return [...]string{"neutral", "friendly", "allied", "rival", "embargo", "at war"}[r]
}

// Faction is a civilization.
type Faction struct {
	Key, Name   string
	Site        int // index into Town.World.Sites
	Discovered  bool
	Relation    Relation
	Opinion     int
	Strength    int
	TradeCount  int
	Personality string
	Specialty   string
	LentWorkers int
}

// Trip is an expedition under way.
type Trip struct {
	Name      string
	TicksLeft int
}

// Expeditions is the scouting and military picture.
type Expeditions struct {
	Scout, Military *Trip
	Completed       int
	Auto            bool // a Geographic Society dispatches scouts
}

// Harbinger is the live harbinger.
type Harbinger struct {
	Key, Name   string
	Tier        string
	Probability float64
	Numeric     bool
	Invited     bool
	Target      string // the epoch (or "the Last Passage") it warns of
}

// Catastrophe is the pending catastrophe and the pressure toward one.
type Catastrophe struct {
	Pending     string // epoch key of a pending catastrophe, "" if none
	PendingName string
	// Pressure is 0..1: 1 while one is pending, rising with the
	// harbinger's odds and the outlook otherwise.
	Pressure float64
	Tier     string
}

// QueueItem is a building under construction.
type QueueItem struct {
	Name     string
	Progress float64 // 0..1
}

// Build lays a snapshot out. since is the baseline for the recap (nil for
// none). It is a pure function of its inputs; Builder caches the parts that
// depend only on the seed.
func (b *Builder) Build(st *game.GameState, since *Visit) *Model {
	cat := b.cat
	m := &Model{Seed: st.Seed, Tick: st.Tick, Age: st.Age, AgeName: st.AgeName, AgeIdx: cat.AgeIdx[st.Age],
		NextAge: st.NextAge, NextAgeName: st.NextAgeName, AgeReady: st.AgeReady, Catalog: cat,
		byKey: map[string]*Building{}}
	m.Epoch = cat.EpochOfAge(m.AgeIdx)
	m.EpochKey = cat.Epochs[m.Epoch]
	m.EpochName = cat.EpochName[m.Epoch]
	if m.AgeName == "" {
		m.AgeName = cat.AgeNames[m.AgeIdx]
	}
	m.Clock = ClockAt(st.Tick)
	m.Weather = weatherFor(st, m.Clock, m.Epoch)
	m.buildings(st, since)
	m.lineages()
	m.people(st)
	m.world(st)
	m.Town = b.town(m)
	m.Skyline = skylineFor(m)
	m.Recap = recap(m, st, since)
	m.Flows = flowsFor(m, st)
	m.Activity = activityFor(m)
	m.LayoutKey = layoutKey(m)
	return m
}

func layoutKey(m *Model) uint64 {
	b2i := func(b bool) int64 {
		if b {
			return 1
		}
		return 0
	}
	h := Hash(m.Seed, int64(m.AgeIdx), int64(m.Workers.Staffed), int64(m.Workers.Idle), int64(m.Workers.Pop),
		HashStr(m.Catastrophe.Pending), int64(m.Expeditions.Completed))
	for _, b := range m.Buildings {
		h = Hash(int64(h), HashStr(b.Key), int64(b.Count), int64(b.Ruins), int64(b.Workers), b2i(b.Legacy), int64(b.Delta))
	}
	for _, t := range m.Town.Tiles {
		h = Hash(int64(h), int64(t.X), int64(t.Y), HashStr(t.Key), b2i(t.Fresh), b2i(t.Legacy), b2i(t.Ruin))
	}
	for _, w := range m.Wonders {
		h = Hash(int64(h), HashStr(w.Key), b2i(w.Built), b2i(w.Delta))
	}
	for _, f := range m.Factions {
		h = Hash(int64(h), HashStr(f.Key), b2i(f.Discovered), int64(f.Relation))
	}
	for _, r := range m.Routes {
		h = Hash(int64(h), HashStr(r.Key), HashStr(r.Civ), b2i(r.Disrupted))
	}
	if e := m.Expeditions.Scout; e != nil {
		h = Hash(int64(h), HashStr(e.Name))
	}
	if e := m.Expeditions.Military; e != nil {
		h = Hash(int64(h), HashStr(e.Name), 2)
	}
	if m.Harbinger != nil {
		h = Hash(int64(h), HashStr(m.Harbinger.Name))
	}
	return h
}

// Building returns the owned type with key k, or nil.
func (m *Model) Building(k string) *Building { return m.byKey[k] }

// Faction returns the civ with key k, or nil.
func (m *Model) Faction(k string) *Faction {
	for i := range m.Factions {
		if m.Factions[i].Key == k {
			return &m.Factions[i]
		}
	}
	return nil
}

// Lineage returns the lineage group with key k, or nil.
func (m *Model) Lineage(k string) *Lineage {
	for _, l := range m.Lineages {
		if l.Key == k {
			return l
		}
	}
	return nil
}

// TotalBuildings is the number of building copies owned.
func (m *Model) TotalBuildings() int {
	n := 0
	for _, b := range m.Buildings {
		n += b.Count
	}
	return n
}

// DiscoveredCount is the number of civs met.
func (m *Model) DiscoveredCount() int {
	n := 0
	for _, f := range m.Factions {
		if f.Discovered {
			n++
		}
	}
	return n
}

func (m *Model) buildings(st *game.GameState, since *Visit) {
	cat := m.Catalog
	// Share each resource's building rate across its producers by their
	// base production, so a building can say "+3.2 iron/s".
	weight := map[string]float64{}
	keys := make([]string, 0, len(st.Buildings))
	for k := range st.Buildings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		bs := st.Buildings[k]
		d := cat.Defs[k]
		if d == nil || bs.Count == 0 || d.Output == "" {
			continue
		}
		weight[d.Output] += float64(d.BaseOutput * float64(bs.Count))
	}
	for _, k := range keys {
		bs := st.Buildings[k]
		d := cat.Defs[k]
		if d == nil || bs.Count == 0 && bs.RuinCount == 0 {
			continue
		}
		b := &Building{Def: d, Count: bs.Count, Ruins: bs.RuinCount, Workers: bs.WorkersAssigned,
			Capacity: bs.WorkerCapacity * bs.Count, Legacy: bs.IsLegacy, Upgrade: bs.PendingUpgrade, Staffing: -1}
		if b.Capacity > 0 {
			b.Staffing = clamp01(float64(b.Workers) / float64(b.Capacity))
		}
		if r, ok := st.Resources[d.Output]; ok && weight[d.Output] > 0 && bs.Count > 0 {
			share := float64(d.BaseOutput*float64(bs.Count)) / weight[d.Output]
			b.Rate = float64(r.Breakdown.BuildingRate*share) / game.BaseTickInterval.Seconds()
			b.RateName = r.Name
		}
		if since != nil {
			b.Delta = b.Count - since.Buildings[k]
		}
		b.Class = LineageClass(d.Lineage)
		if b.Legacy {
			b.Class = CLegacy
		}
		b.Sym = LineageSym(d.Lineage, m.Epoch)
		m.Buildings = append(m.Buildings, b)
		m.byKey[k] = b
	}
	order := map[string]int{}
	for i, l := range LineageOrder {
		order[l] = i
	}
	order[LinWonder] = len(LineageOrder)
	sort.SliceStable(m.Buildings, func(i, j int) bool {
		a, c := m.Buildings[i], m.Buildings[j]
		if order[a.Lineage] != order[c.Lineage] {
			return order[a.Lineage] < order[c.Lineage]
		}
		return a.Rank < c.Rank
	})
	// Wonders: every one built, plus the one this age needs.
	for _, d := range cat.Wonders {
		bs, ok := st.Buildings[d.Key]
		built := ok && bs.Count > 0
		current := d.Key == st.CurrentAgeWonderKey
		if !built && !current {
			continue
		}
		w := Wonder{Key: d.Key, Name: d.Name, Age: d.Age, Built: built, Current: current}
		if built {
			w.Progress = 1
			w.Delta = since != nil && since.Buildings[d.Key] == 0
		} else {
			w.Progress = bankProgress(bs)
		}
		m.Wonders = append(m.Wonders, w)
	}
}

// bankProgress is how much of a wonder's cost is banked, 0..1.
func bankProgress(bs game.BuildingState) float64 {
	if bs.WonderBankFull {
		return 1
	}
	var have, need float64
	res := make([]string, 0, len(bs.NextCost))
	for r := range bs.NextCost {
		res = append(res, r)
	}
	sort.Strings(res)
	for _, r := range res {
		c := bs.NextCost[r]
		if c <= 0 {
			continue
		}
		b := bs.WonderBank[r]
		if b > c {
			b = c
		}
		have += b
		need += c
	}
	if need <= 0 {
		return 0
	}
	return clamp01(have / need)
}

func (m *Model) lineages() {
	idx := map[string]*Lineage{}
	for _, b := range m.Buildings {
		if b.Wonder {
			continue
		}
		l := idx[b.Lineage]
		if l == nil {
			l = &Lineage{Key: b.Lineage, Name: LineageNames[b.Lineage], Class: LineageClass(b.Lineage),
				Sym: LineageSym(b.Lineage, m.Epoch)}
			idx[b.Lineage] = l
			m.Lineages = append(m.Lineages, l)
		}
		l.Buildings = append(l.Buildings, b)
		l.Count += b.Count
		l.Ruins += b.Ruins
		if b.Legacy {
			l.Legacy += b.Count
		}
		l.Workers += b.Workers
		l.Capacity += b.Capacity
	}
	// Share of working output: every copy counts, staffed copies fully and
	// idle ones at a fifth (the engine's 0.20 + 0.80 × fill shape).
	total := 0.0
	act := map[string]float64{}
	for _, l := range m.Lineages {
		a := 0.0
		for _, b := range l.Buildings {
			f := 1.0
			if b.Staffing >= 0 {
				f = 0.2 + float64(0.8*b.Staffing)
			}
			a += float64(float64(b.Count) * f)
		}
		act[l.Key] = a
		total += a
	}
	for _, l := range m.Lineages {
		if total > 0 {
			l.Share = act[l.Key] / total
		}
		l.Tiles = TownTiles(l.Key, l.Count+l.Ruins)
	}
}

func (m *Model) people(st *game.GameState) {
	w := &m.Workers
	w.Pop, w.MaxPop, w.Idle = st.Workers.TotalPop, st.Workers.MaxPop, st.Workers.TotalIdle
	for _, b := range m.Buildings {
		w.Staffed += b.Workers
		w.Capacity += b.Capacity
	}
	m.Army = Army{Soldiers: st.Military.SoldierCount, Cap: st.Military.SoldierCap, Defense: st.Military.DefenseRating}
	e := &m.Expeditions
	if s := st.Military.ActiveScout; s != nil {
		e.Scout = &Trip{Name: s.Name, TicksLeft: s.TicksLeft}
	}
	if s := st.Military.ActiveMilitary; s != nil {
		e.Military = &Trip{Name: s.Name, TicksLeft: s.TicksLeft}
	}
	e.Completed = st.Military.CompletedCount
	e.Auto = st.Military.AutoExpedition.Active
	for _, q := range st.BuildQueue {
		p := 0.0
		if q.TotalTicks > 0 {
			p = clamp01(1 - float64(q.TicksLeft)/float64(q.TotalTicks))
		}
		m.Queue = append(m.Queue, QueueItem{Name: q.Name, Progress: p})
	}
}

func (m *Model) world(st *game.GameState) {
	cat := m.Catalog
	for i, k := range cat.Factions {
		fi, ok := st.Diplomacy.Factions[k]
		if !ok {
			continue
		}
		f := Faction{Key: k, Name: fi.Name, Site: i, Discovered: fi.Discovered, Opinion: fi.Opinion,
			Strength: fi.Strength, TradeCount: fi.TradeCount, Personality: fi.Personality,
			Specialty: fi.Specialty, LentWorkers: fi.LentWorkers}
		switch {
		case fi.AtWar:
			f.Relation = RelWar
		case fi.Status == "embargo":
			f.Relation = RelEmbargo
		case fi.Status == "rival":
			f.Relation = RelRival
		case fi.Status == "allied":
			f.Relation = RelAllied
		case fi.Status == "friendly":
			f.Relation = RelFriendly
		}
		m.Factions = append(m.Factions, f)
	}
	// Caravans go to civs you can trade with: a stable pick per route.
	var partners []string
	for _, f := range m.Factions {
		if f.Discovered && f.Relation != RelWar && f.Relation != RelEmbargo {
			partners = append(partners, f.Key)
		}
	}
	for _, r := range st.Trade.ActiveRoutes {
		rt := Route{Key: r.Key, Name: r.Name, Disrupted: r.Disrupted, Cycles: r.CyclesDone, Mode: cat.RouteMode(r.Key)}
		for res := range r.Import {
			rt.Imports = append(rt.Imports, res)
		}
		sort.Strings(rt.Imports)
		if len(partners) > 0 {
			rt.Civ = partners[Hash(m.Seed, HashStr(r.Key))%uint64(len(partners))]
		}
		m.Routes = append(m.Routes, rt)
	}
	sort.Slice(m.Routes, func(i, j int) bool { return m.Routes[i].Key < m.Routes[j].Key })
	if h := st.Harbinger; h != nil {
		m.Harbinger = &Harbinger{Key: h.Key, Name: h.Name, Tier: string(h.Tier), Probability: h.Probability,
			Numeric: h.Numeric, Invited: h.Invited, Target: h.TargetEpochName}
	}
	c := &m.Catastrophe
	c.Tier = string(st.CatastropheOutlook.Tier)
	if st.PendingCatastrophe != "" {
		c.Pending = st.PendingCatastrophe
		c.PendingName = cat.Catastrophes[st.PendingCatastrophe]
		if c.PendingName == "" {
			c.PendingName = "Catastrophe"
		}
		c.Pressure = 1
	} else {
		p := 0.0
		if st.CatastropheOutlook.Possible {
			p = float64(st.CatastropheOutlook.Probability * 0.8)
		}
		if h := st.Harbinger; h != nil {
			hp := 0.35 + float64(h.Probability*2)
			if hp > p {
				p = hp
			}
		}
		c.Pressure = clamp01(p)
	}
}
