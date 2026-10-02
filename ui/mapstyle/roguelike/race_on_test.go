//go:build race

package roguelike

// raceOn lets the timing tests skip under -race, which slows drawing about
// twentyfold; the plain test run still times it.
const raceOn = true
