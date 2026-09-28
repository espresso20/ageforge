// Command mud is the Map Lab "mud" prototype: the settlement as a small text
// world you walk through. See DESIGN.md.
//
//	go run ./lab/mud -gen                       # play a seed, write states/*.json
//	go run ./lab/mud -age medieval_age          # walk the city interactively
//	go run ./lab/mud -capture                   # write captures/*.txt|html
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	gen := flag.Bool("gen", false, "play a seed with the smoke bot and write state snapshots")
	seed := flag.Int64("seed", 7, "seed for -gen")
	stop := flag.String("stop", "galactic_age", "last age -gen snapshots")
	maxSim := flag.Duration("max-sim", 2000*time.Hour, "simulated-time budget for -gen")
	states := flag.String("states", "lab/mud/states", "state snapshot directory")
	age := flag.String("age", "medieval_age", "state to walk (a snapshot name in -states)")
	since := flag.String("since", "", "an earlier snapshot to diff against (idle return)")
	themeKey := flag.String("theme", "forge", "theme key")
	capture := flag.Bool("capture", false, "write the capture set and exit")
	out := flag.String("out", "lab/mud/captures", "capture directory")
	flag.Parse()

	var err error
	switch {
	case *gen:
		err = generate(*states, *seed, *stop, *maxSim)
	case *capture:
		err = captureAll(*states, *out)
	default:
		err = runInteractive(*states, *age, *since, *themeKey)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "mud:", err)
		os.Exit(1)
	}
}
