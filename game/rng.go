package game

import "math/rand"

// countingSource is math/rand's seeded source with a step counter, so a save
// can record how far into the stream a run is and a load can replay to that
// point. math/rand's source can't be serialized, but it can be re-seeded and
// stepped, and every draw a *rand.Rand makes (Intn, Float64, Perm, ...) is
// some number of Int63/Uint64 calls on its source, each exactly one step.
// Wrapping the stock source keeps every seed's stream bit-identical to what
// it was before the counter existed.
type countingSource struct {
	src   rand.Source64
	draws uint64
}

func newCountingSource(seed int64) *countingSource {
	return &countingSource{src: rand.NewSource(seed).(rand.Source64)}
}

func (s *countingSource) Int63() int64 {
	s.draws++
	return s.src.Int63()
}

func (s *countingSource) Uint64() uint64 {
	s.draws++
	return s.src.Uint64()
}

func (s *countingSource) Seed(seed int64) {
	s.src.Seed(seed)
	s.draws = 0
}

// skip advances the stream n steps without handing out the values.
func (s *countingSource) skip(n uint64) {
	for i := uint64(0); i < n; i++ {
		s.src.Uint64()
	}
	s.draws += n
}

// maxRNGReplay caps how many steps a load will replay to restore a stream
// position: a few seconds of stepping at worst, and orders of magnitude past
// any real run (the smoke suite's longest runs draw well under one step per
// tick). A save claiming more has been edited; it gets a stream restarted
// from the seed instead of a hung load.
const maxRNGReplay = 1 << 31

// rngDraws reports the stream positions to save: how many steps the gameplay
// and quip streams have taken since the seed. A stream a test replaced with
// its own *rand.Rand reports 0. Call under the lock.
func (ge *GameEngine) rngDraws() (gameplay, quip uint64) {
	if ge.rngSrc != nil && ge.rng == ge.rngOwner {
		gameplay = ge.rngSrc.draws
	}
	if ge.quipSrc != nil && ge.quip == ge.quipOwner {
		quip = ge.quipSrc.draws
	}
	return gameplay, quip
}

// restoreRNG re-seeds both streams from seed and replays them to the saved
// positions, so a loaded game draws exactly what the saved one would have
// drawn next. Positions past maxRNGReplay are ignored (the stream restarts
// from the seed, as every load did before positions were saved). Returns
// whether the positions were applied. Call under the write lock.
func (ge *GameEngine) restoreRNG(seed int64, gameplay, quip uint64) bool {
	ge.SeedRNG(seed)
	if gameplay > maxRNGReplay || quip > maxRNGReplay {
		return false
	}
	ge.rngSrc.skip(gameplay)
	ge.quipSrc.skip(quip)
	return true
}
