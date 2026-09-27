package game

import "math/rand"

// testSeed is the default seed for tests that drive a manager directly and do
// not care about the particular roll, only that it is the same every run.
const testSeed int64 = 0x5EED

// testRNG returns a fresh *rand.Rand seeded with testSeed. Managers take their
// randomness as an argument (EventManager.Tick, MilitaryManager.Tick and
// LaunchExpedition, DiplomacyManager.Tick), and this is what tests pass when
// they call those methods without an engine.
func testRNG() *rand.Rand { return rand.New(rand.NewSource(testSeed)) }

// newSeededEngine returns a NewGameEngine whose run RNG is seeded with seed, so
// every gameplay roll the engine makes is reproducible.
func newSeededEngine(seed int64) *GameEngine {
	ge := NewGameEngine()
	ge.SeedRNG(seed)
	return ge
}
