// Command atlas is the Map Lab's "atlas" prototype: the world view as a
// cartographic map drawn in text, rendered from real game states the smoke
// bot produced.
//
//	go run ./lab/atlas -gen                      # play the bot, write states/
//	go run ./lab/atlas -age medieval_age         # interactive
//	go run ./lab/atlas -capture lab/atlas/captures   # every capture
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/espresso20/ageforge/theme"
)

func main() {
	gen := flag.Bool("gen", false, "play the smoke bot and write snapshots")
	seed := flag.Int64("seed", 42, "bot seed for -gen")
	stop := flag.String("stop", "quantum_age", "last age to snapshot with -gen")
	hours := flag.Float64("max-hours", 1500, "simulated-hours cap for -gen")
	dir := flag.String("states", "lab/atlas/states", "snapshot directory")
	age := flag.String("age", "medieval_age", "age snapshot to open")
	themeKey := flag.String("theme", "forge", "theme key")
	capture := flag.String("capture", "", "write every capture into this directory and exit")
	print := flag.Bool("print", false, "print one frame as text and exit")
	w := flag.Int("w", 160, "width for -print")
	h := flag.Int("h", 48, "height for -print")
	plate := flag.String("plate", "", "force a plate (hide, portolan, engraved, survey, satellite, radar, orrery, stars, galaxy)")
	flag.Parse()

	if *gen {
		if err := generate(*dir, *seed, *stop, *hours); err != nil {
			fail(err)
		}
		return
	}
	if err := theme.SetActive(*themeKey); err != nil {
		fail(err)
	}
	lib := newLibrary(*dir)
	if len(lib.ages) == 0 {
		fail(fmt.Errorf("no snapshots in %s: run with -gen first", *dir))
	}
	if *capture != "" {
		var alt *library
		if l := newLibrary(*dir + "-alt"); len(l.ages) > 0 {
			alt = l
		}
		if err := captureAll(lib, alt, *capture); err != nil {
			fail(err)
		}
		return
	}
	a, err := lib.atlas(*age)
	if err != nil {
		fail(err)
	}
	sc := newScene(a)
	if *plate != "" {
		sc.Plate = plates[*plate]
	}
	if *print {
		cv := sc.Draw(*w, *h)
		cells, err := simulate(cv)
		if err != nil {
			fail(err)
		}
		fmt.Print(toText(cells, cv.W, cv.H))
		return
	}
	if err := runApp(lib, sc); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "atlas:", err)
	os.Exit(1)
}

// library loads snapshots and keeps one world per seed.
type library struct {
	dir    string
	ages   []string
	worlds map[int64]*World
	cache  map[string]*Atlas
}

func newLibrary(dir string) *library {
	return &library{dir: dir, ages: availableAges(dir), worlds: map[int64]*World{}, cache: map[string]*Atlas{}}
}

func (l *library) atlas(age string) (*Atlas, error) {
	if a, ok := l.cache[age]; ok {
		return a, nil
	}
	snap, err := loadSnapshot(l.dir, age)
	if err != nil {
		return nil, fmt.Errorf("no snapshot for %s (have: %s)", age, strings.Join(l.ages, ", "))
	}
	w, ok := l.worlds[snap.State.Seed]
	if !ok {
		w = newWorld(snap.State.Seed)
		l.worlds[snap.State.Seed] = w
	}
	// the previous snapshot stands in for the player's last check-in
	var prev *Snapshot
	for i, a := range l.ages {
		if a == age && i > 0 {
			prev, _ = loadSnapshot(l.dir, l.ages[i-1])
		}
	}
	a := buildAtlas(w, snap, prev)
	l.cache[age] = a
	return a, nil
}
