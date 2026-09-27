//go:build race

package smoke

// raceEnabled lets the scenario tests that play thousands of ticks skip
// under -race, where they take minutes; the smoke CI job runs the same
// scenarios without it.
const raceEnabled = true
