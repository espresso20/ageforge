package mapmodel

import (
	"sort"

	"github.com/espresso20/ageforge/game"
)

// Builder builds models. It owns the catalogue and caches the parts that
// depend only on the seed (the world and the town plan, about 10 ms to
// make), plus the last model it built, keyed by the snapshot it came from.
// It is not safe for concurrent use; the UI keeps one per map view.
type Builder struct {
	cat *Catalog

	seed  int64
	world *World
	plan  *Plan

	lastSig uint64
	last    *Model
}

// NewBuilder makes a builder over a catalogue (NewCatalog when nil).
func NewBuilder(cat *Catalog) *Builder {
	if cat == nil {
		cat = NewCatalog()
	}
	return &Builder{cat: cat}
}

// Catalog returns the builder's catalogue.
func (b *Builder) Catalog() *Catalog { return b.cat }

func (b *Builder) seedParts(seed int64) (*World, *Plan) {
	if b.world == nil || b.seed != seed {
		b.seed = seed
		b.world = NewWorld(seed, len(b.cat.Factions))
		b.plan = NewPlan(b.world, b.cat)
	}
	return b.world, b.plan
}

// Model returns the model for a snapshot, reusing the last one when the
// snapshot has not changed in anything the maps draw.
func (b *Builder) Model(st *game.GameState, since *Visit) *Model {
	sig := Signature(st, since)
	if b.last != nil && sig == b.lastSig {
		return b.last
	}
	m := b.Build(st, since)
	b.last, b.lastSig = m, sig
	return m
}

// Signature hashes what the maps read from a snapshot, so a cache can tell
// when a model must be rebuilt: the tick (the clock and weather move with
// it), building counts, staffing and flags, workers, routes, civs,
// expeditions, the harbinger, the pending catastrophe and the baseline.
func Signature(st *game.GameState, since *Visit) uint64 {
	h := Hash(st.Seed, int64(st.Tick), HashStr(st.Age), int64(st.Workers.TotalPop), int64(st.Workers.TotalIdle),
		int64(len(st.BuildQueue)), HashStr(st.PendingCatastrophe), HashStr(st.CurrentAgeWonderKey))
	keys := make([]string, 0, len(st.Buildings))
	for k, bs := range st.Buildings {
		if bs.Count > 0 || bs.RuinCount > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		bs := st.Buildings[k]
		leg := int64(0)
		if bs.IsLegacy {
			leg = 1
		}
		h = Hash(int64(h), HashStr(k), int64(bs.Count), int64(bs.RuinCount), int64(bs.WorkersAssigned), leg)
	}
	fk := make([]string, 0, len(st.Diplomacy.Factions))
	for k := range st.Diplomacy.Factions {
		fk = append(fk, k)
	}
	sort.Strings(fk)
	for _, k := range fk {
		f := st.Diplomacy.Factions[k]
		d, w := int64(0), int64(0)
		if f.Discovered {
			d = 1
		}
		if f.AtWar {
			w = 1
		}
		h = Hash(int64(h), HashStr(k), d, w, HashStr(f.Status), int64(f.Opinion))
	}
	for _, r := range st.Trade.ActiveRoutes {
		dis := int64(0)
		if r.Disrupted {
			dis = 1
		}
		h = Hash(int64(h), HashStr(r.Key), dis)
	}
	if s := st.Military.ActiveScout; s != nil {
		h = Hash(int64(h), HashStr(s.Name), int64(s.TicksLeft))
	}
	if s := st.Military.ActiveMilitary; s != nil {
		h = Hash(int64(h), HashStr(s.Name), int64(s.TicksLeft))
	}
	if hb := st.Harbinger; hb != nil {
		h = Hash(int64(h), HashStr(hb.Name))
	}
	for _, e := range st.ActiveEvents {
		h = Hash(int64(h), HashStr(e.Key))
	}
	if since != nil {
		h = Hash(int64(h), int64(since.Tick), int64(len(since.Buildings)), 1)
	}
	return h
}
