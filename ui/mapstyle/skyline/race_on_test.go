//go:build race

package skyline

// raceEnabled lets the frame-time test skip under -race, which slows
// drawing about twentyfold; the plain test run still times it.
const raceEnabled = true
