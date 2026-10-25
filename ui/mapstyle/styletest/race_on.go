//go:build race

package styletest

// RaceEnabled lets the frame-time tests skip under -race, which slows
// drawing about twentyfold; the plain test run still times it.
const RaceEnabled = true
