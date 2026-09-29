// Command skyline is the Map Lab "skyline" concept: the civilisation drawn
// side-on as a BBS/ANSI-art city panorama that grows with the empire.
//
//	go run ./lab/skyline -gen                  # play the smoke bot, save states
//	go run ./lab/skyline -age medieval_age     # interactive viewer
//	go run ./lab/skyline -captures             # write lab/skyline/captures/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	gen := flag.Bool("gen", false, "play the smoke bot and save one state per age")
	genOut := flag.String("gen-out", stateDir, "directory -gen writes to")
	seed := flag.Int64("seed", 7, "seed for -gen")
	stop := flag.String("stop", "quantum_age", "last age -gen saves")
	hours := flag.Float64("hours", 1400, "simulated-hour budget for -gen")
	age := flag.String("age", "medieval_age", "state to view (an age key, or harbinger_<age>/catastrophe_<age>)")
	themeKey := flag.String("theme", "forge", "theme key")
	caps := flag.Bool("captures", false, "write the capture set to lab/skyline/captures")
	sheet := flag.String("sheet", "", "write a sprite sheet of every form and wonder to this HTML file")
	review := flag.String("review", "", "write an 80-column review crop: state:tod:screensBack[:theme] (comma-separated list)")
	flag.Parse()

	var err error
	switch {
	case *gen:
		err = generate(*genOut, *seed, *stop, *hours)
	case *sheet != "":
		for i, r := range [][2]int{{0, 4}, {5, 8}, {9, 13}, {14, 17}, {18, 21}} {
			if err = writeSheet(fmt.Sprintf("%s_%d.html", *sheet, i+1), r[0] >= 9, r[0], r[1]); err != nil {
				break
			}
		}
	case *review != "":
		for _, r := range strings.Split(*review, ",") {
			if err = writeReview(stateDir, "lab/skyline/captures", r); err != nil {
				break
			}
		}
	case *caps:
		err = writeCaptures(stateDir, "lab/skyline/captures")
	default:
		err = runInteractive(stateDir, *age, *themeKey)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "skyline:", err)
		os.Exit(1)
	}
}
