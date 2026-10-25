//go:build !race

package styletest

// RaceEnabled is false unless the tests run under -race.
const RaceEnabled = false
