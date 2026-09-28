package main

import (
	"hash/fnv"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The organism reads the game through this small model. Everything the
// renderer draws comes from here, and everything here comes from one
// game.GameState, so a loaded save draws the same tree it was saved with.

// limbSpec is a production lineage that grows as a limb of the tree. The
// list is the crown's left-to-right order, and the spectrum bar's: the land
// (food, timber, stone, metal, craft, energy) on the left, the mind
// (knowledge, culture, faith) in the middle, the outward-facing lineages
// (trade, harbor, military, net, orbit) on the right. Limbs never swap
// places; a limb's wedge of the crown widens and narrows with its share.
type limbSpec struct {
	Key   string
	Label string
	Leaf  rune   // the lineage's leaf glyph: shape carries the meaning
	Hue   uint32 // identity hue for a dark background; Legible() adapts it
}

var limbSpecs = []limbSpec{
	{"food", "food", '&', 0x9ad44a},
	{"organic_extraction", "timber", '%', 0x3fae5a},
	{"geological_extraction", "stone", '#', 0xb8ab8c},
	{"metallurgy", "metal", '◊', 0xd0784e},
	{"engineering", "craft", '¤', 0xe3b341},
	{"energy", "energy", '!', 0xffe14d},
	{"knowledge", "knowledge", '*', 0x5cb6ff},
	{"culture_arts", "culture", '♪', 0xff7eb8},
	{"faith", "faith", '+', 0xc3a2ff},
	{"trade", "trade", '$', 0xffa23f},
	{"harbor", "harbor", 'ω', 0x3fd6c6},
	{"military", "military", '^', 0xff5a5a},
	{"hacker", "net", '@', 0x2fe8ff},
	{"astronaut", "orbit", 'o', 0xe6ecff},
}

// Limb is one lineage's share of the civilisation.
type Limb struct {
	Spec      limbSpec
	Index     int // position in limbSpecs, fixes the limb's angle
	Count     int // buildings owned (legacy included)
	Tier      int // highest lineage tier owned
	Workers   int
	WorkerCap int
	Rate      float64 // net rate of the lineage's main output
	Buildings []named
}

type named struct {
	Key, Name string
	Count     int
}

// Vigor is how well the limb is staffed, 0..1; unstaffable lineages count
// as fully vigorous.
func (l Limb) Vigor() float64 {
	if l.WorkerCap <= 0 {
		return 1
	}
	return math.Min(1, float64(l.Workers)/float64(l.WorkerCap))
}

type Wonder struct {
	Key, Name string
	Built     bool
	Progress  float64 // 0..1 of the bank, for an unbuilt wonder of this age
}

// Stratum is one epoch's layer of soil under the tree.
type Stratum struct {
	EpochKey, Name, Icon string
	Ticks                int      // weight: ages reached in the epoch
	Ages                 []string // age names inside it, oldest first
	Current              bool
}

// Fossil is an event buried in the layer of the epoch it happened in.
type Fossil struct {
	EpochKey string
	Kind     string // good_minor ... catastrophe, harbinger, ruin
	Name     string
	Outcome  string
	Tick     int
}

type Neighbor struct {
	Key, Name string
	Strength  int
	Opinion   int
	AtWar     bool
	Status    string
}

type Route struct {
	Key, Name string
	Disrupted bool
	Export    string
	Import    string
}

type Organism struct {
	Seed               int64
	Tick               int
	Age, AgeName       string
	AgeIdx, AgeCount   int
	EpochKey, Epoch    string
	EpochIcon          string
	EpochIdx           int
	Limbs              []Limb // lineages with at least one building, in crown order
	Pop, MaxPop, Idle  int
	Housing            int
	Storage            int
	Wonders            []Wonder
	Morale             float64
	FoodRate           float64
	Techs              int
	Built              int
	Strata             []Stratum
	Fossils            []Fossil
	Neighbors          []Neighbor
	Routes             []Route
	Harbinger          *game.HarbingerView
	PendingCatastrophe string
	Events             []string
	Ruins              int
	Headlines          []game.LogEntry
	Soldiers           int
	AgeReady           bool
	NextAgeName        string

	genomes map[uint64]genome // grown once per state; see genome()
}

// genome returns the genome for seed, growing it on first use. Genomes
// depend only on the seed, so a state keeps its own and a frame costs no
// regrowth.
func (o *Organism) genome(seed uint64, depth int) genome {
	if g, ok := o.genomes[seed]; ok {
		return g
	}
	if o.genomes == nil {
		o.genomes = map[uint64]genome{}
	}
	g := grow(seed, depth)
	o.genomes[seed] = g
	return g
}

// Total buildings on the limbs.
func (o *Organism) LimbTotal() int {
	n := 0
	for _, l := range o.Limbs {
		n += l.Count
	}
	return n
}

func hash64(parts ...string) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	// FNV alone avalanches poorly on near-identical keys ("star 1",
	// "star 2"), which shows up as diagonal stripes; finish with splitmix.
	z := h.Sum64() + 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// Build distils a GameState into the organism model.
func Build(st game.GameState) *Organism {
	defs := config.BuildingByKey()
	ages := config.AgeOrder()
	ageIdx := map[string]int{}
	for i, k := range ages {
		ageIdx[k] = i
	}
	epochs := config.Epochs()
	o := &Organism{
		Seed: st.Seed, Tick: st.Tick, Age: st.Age, AgeName: st.AgeName,
		AgeIdx: ageIdx[st.Age], AgeCount: len(ages),
		EpochKey: st.EpochKey, Epoch: st.EpochName, EpochIcon: st.EpochIcon,
		Pop: st.Workers.TotalPop, MaxPop: st.Workers.MaxPop, Idle: st.Workers.TotalIdle,
		Morale: st.Morale, Techs: st.Research.TotalResearched, Built: st.Stats.TotalBuilt,
		Harbinger: st.Harbinger, PendingCatastrophe: st.PendingCatastrophe,
		AgeReady: st.AgeReady, NextAgeName: st.NextAgeName,
	}
	for _, e := range epochs {
		if e.Key == st.EpochKey {
			o.EpochIdx = e.Order
		}
	}
	if r, ok := st.Resources["food"]; ok {
		o.FoodRate = r.Rate
	}
	if r, ok := st.Resources["soldiers"]; ok {
		o.Soldiers = int(r.Amount)
	}

	limbAt := map[string]int{}
	for i, s := range limbSpecs {
		limbAt[s.Key] = i
	}
	limbs := make([]*Limb, len(limbSpecs))
	for _, key := range sortedKeys(st.Buildings) {
		b := st.Buildings[key]
		o.Ruins += b.RuinCount
		if b.Count <= 0 {
			continue
		}
		def := defs[key]
		switch {
		case b.Category == "wonder" || def.LineageKey == "wonder" || def.LineageKey == "monument":
			continue // fruit, below
		case def.LineageKey == "housing" || b.Category == "housing":
			o.Housing += b.Count
			continue
		case def.LineageKey == "storage" || b.Category == "storage":
			o.Storage += b.Count
			continue
		}
		i, ok := limbAt[def.LineageKey]
		if !ok {
			continue
		}
		if limbs[i] == nil {
			limbs[i] = &Limb{Spec: limbSpecs[i], Index: i}
		}
		l := limbs[i]
		l.Count += b.Count
		if def.LineageTier > l.Tier {
			l.Tier = def.LineageTier
		}
		l.Workers += b.WorkersAssigned
		l.WorkerCap += b.WorkerCapacity * b.Count
		l.Buildings = append(l.Buildings, named{key, b.Name, b.Count})
		if out := def.OutputResource; out != "" {
			if r, ok := st.Resources[out]; ok && r.Rate > l.Rate {
				l.Rate = r.Rate
			}
		}
	}
	for _, l := range limbs {
		if l != nil {
			sort.Slice(l.Buildings, func(a, b int) bool { return l.Buildings[a].Count > l.Buildings[b].Count })
			o.Limbs = append(o.Limbs, *l)
		}
	}

	// Wonders: every wonder of an age reached, built or not; the current
	// age's unbuilt one carries its bank progress.
	for _, key := range sortedKeys(st.Buildings) {
		b := st.Buildings[key]
		def := defs[key]
		if !(b.Category == "wonder" || def.LineageKey == "wonder" || def.LineageKey == "monument") {
			continue
		}
		if ai, ok := ageIdx[def.RequiredAge]; !ok || ai > o.AgeIdx {
			continue
		}
		w := Wonder{Key: key, Name: b.Name, Built: b.Count > 0}
		if !w.Built && len(b.NextCost) > 0 {
			var paid, need float64
			for res, c := range b.NextCost {
				need += c
				paid += math.Min(c, b.WonderBank[res])
			}
			if need > 0 {
				w.Progress = paid / need
			}
		}
		if w.Built || def.RequiredAge == st.Age {
			o.Wonders = append(o.Wonders, w)
		}
	}

	// Strata: one per epoch reached, holding the ages reached in it. The
	// history collector keeps only a short window, so thickness comes from
	// the number of ages, not the time spent.
	ageNames := map[string]string{}
	for _, a := range config.Ages() {
		ageNames[a.Key] = a.Name
	}
	for _, e := range epochs {
		if e.Order > o.EpochIdx {
			break
		}
		s := Stratum{EpochKey: e.Key, Name: e.Name, Icon: e.Icon, Current: e.Key == st.EpochKey}
		for _, a := range e.Ages {
			if ageIdx[a] <= o.AgeIdx {
				s.Ages = append(s.Ages, ageNames[a])
			}
		}
		s.Ticks = len(s.Ages)
		o.Strata = append(o.Strata, s)
	}

	for _, ev := range st.EpochEventHistory {
		o.Fossils = append(o.Fossils, Fossil{EpochKey: ev.EpochKey, Kind: ev.EventType, Name: ev.EventName, Outcome: ev.Outcome, Tick: ev.Tick})
	}
	for _, h := range st.HarbingerHistory {
		o.Fossils = append(o.Fossils, Fossil{EpochKey: h.EpochKey, Kind: "harbinger", Name: h.Name, Outcome: h.Outcome, Tick: h.Tick})
	}

	for _, key := range sortedKeys(st.Diplomacy.Factions) {
		f := st.Diplomacy.Factions[key]
		if !f.Discovered {
			continue
		}
		o.Neighbors = append(o.Neighbors, Neighbor{Key: key, Name: f.Name, Strength: f.Strength, Opinion: f.Opinion, AtWar: f.AtWar, Status: f.Status})
	}
	for _, r := range st.Trade.ActiveRoutes {
		o.Routes = append(o.Routes, Route{Key: r.Key, Name: r.Name, Disrupted: r.Disrupted, Export: firstKey(r.Export), Import: firstKey(r.Import)})
	}
	for _, e := range st.ActiveEvents {
		o.Events = append(o.Events, e.Name)
	}
	// Headlines: the newest notable log lines, newest first.
	for i := len(st.Log) - 1; i >= 0 && len(o.Headlines) < 4; i-- {
		e := st.Log[i]
		if e.Type == "info" || strings.TrimSpace(e.Message) == "" || strings.HasPrefix(e.Message, "Welcome back") {
			continue
		}
		o.Headlines = append(o.Headlines, e)
	}
	return o
}

func firstKey[V any](m map[string]V) string {
	ks := sortedKeys(m)
	if len(ks) == 0 {
		return ""
	}
	return ks[0]
}

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
