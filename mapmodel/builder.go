package mapmodel

import (
	"maps"
	"slices"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// Builder builds models. It owns the catalogue and caches the parts that
// depend only on the seed (the world and the town plan, about 10 ms to
// make), plus the last model it built, keyed by the snapshot it came from.
// It is not safe for concurrent use; the UI keeps one per map view.
type Builder struct {
	cat *Catalog

	// ownCat says the builder made its catalogue itself, so it makes it again
	// when VibeMode is switched; vibeOn is the switch as it last looked.
	ownCat bool
	vibeOn bool

	seed  int64
	world *World
	plan  *Plan

	lastSig uint64
	last    *Model
}

// NewBuilder makes a builder over a catalogue (NewCatalog when nil).
func NewBuilder(cat *Catalog) *Builder {
	own := cat == nil
	if own {
		cat = NewCatalog()
	}
	return &Builder{cat: cat, ownCat: own, vibeOn: vibe.On()}
}

// follow lets what the builder holds follow the VibeMode switch: the words of
// a catalogue it made, the name the world was given and the last model are
// built again when the switch has moved since it last looked. A switch is
// rare, and a world and its plan take about 10 ms.
func (b *Builder) follow() {
	if on := vibe.On(); on != b.vibeOn {
		b.vibeOn = on
		if b.ownCat {
			b.cat = NewCatalog()
		}
		b.world, b.plan, b.last = nil, nil, nil
	}
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
	b.follow()
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
	for _, k := range slices.Sorted(maps.Keys(st.Buildings)) {
		bs := st.Buildings[k]
		if bs.Count > 0 || bs.RuinCount > 0 {
			h = Hash(int64(h), HashStr(k), int64(bs.Count), int64(bs.RuinCount), int64(bs.WorkersAssigned), b2i(bs.IsLegacy))
		}
	}
	for _, k := range slices.Sorted(maps.Keys(st.Diplomacy.Factions)) {
		f := st.Diplomacy.Factions[k]
		h = Hash(int64(h), HashStr(k), b2i(f.Discovered), b2i(f.AtWar), HashStr(f.Status), int64(f.Opinion))
	}
	for _, r := range st.Trade.ActiveRoutes {
		h = Hash(int64(h), HashStr(r.Key), b2i(r.Disrupted))
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
