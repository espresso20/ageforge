package main

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// Every snapshot, every plate it has earned, at every size from the
// sidebar to a large terminal: no panic, exact output size, and the cursor
// and verbs work.
func TestDrawAllSizes(t *testing.T) {
	lib := newLibrary("states")
	if len(lib.ages) == 0 {
		t.Skip("no snapshots; run go run ./lab/atlas -gen")
	}
	sizes := [][2]int{{40, 15}, {80, 24}, {100, 30}, {160, 48}, {220, 64}}
	for _, th := range []string{"forge", "daylight"} {
		if err := theme.SetActive(th); err != nil {
			t.Fatal(err)
		}
		for _, age := range lib.ages {
			a, err := lib.atlas(age)
			if err != nil {
				t.Fatal(err)
			}
			for n := 0; n < earnedPlates(age); n++ {
				for _, sz := range sizes {
					sc := newScene(a)
					sc.Plate = plates[plateOrder[n]]
					cv := sc.Draw(sz[0], sz[1])
					if cv.W != sz[0] || cv.H != sz[1] || len(cv.C) != sz[0]*sz[1] {
						t.Fatalf("%s %s %v: canvas %dx%d", age, plateOrder[n], sz, cv.W, cv.H)
					}
					for i := 0; i < 6; i++ {
						sc.Next(1)
						sc.Frame++
						sc.Draw(sz[0], sz[1])
						sc.inspect(newPal(sc.Plate))
					}
				}
			}
		}
	}
	_ = theme.SetActive("forge")
}

// The same state always draws the same map.
func TestDeterministic(t *testing.T) {
	a1, err := newLibrary("states").atlas("medieval_age")
	if err != nil {
		t.Skip(err)
	}
	a2, _ := newLibrary("states").atlas("medieval_age")
	c1 := newScene(a1).Draw(160, 48)
	c2 := newScene(a2).Draw(160, 48)
	for i := range c1.C {
		if c1.C[i] != c2.C[i] {
			t.Fatalf("cell %d differs: %q vs %q", i, c1.C[i].R, c2.C[i].R)
		}
	}
}

// The interactive loop, driven through a simulation screen: move, tab,
// zoom, leaf plates, switch age, verb, quit.
func TestInteractive(t *testing.T) {
	lib := newLibrary("states")
	a, err := lib.atlas("industrial_age")
	if err != nil {
		t.Skip(err)
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	sc := newScene(a)
	done := make(chan error, 1)
	go func() { done <- runAppOn(lib, sc, sim) }()
	time.Sleep(150 * time.Millisecond)
	sim.SetSize(140, 44)
	for _, k := range []tcell.Key{tcell.KeyRight, tcell.KeyDown, tcell.KeyTab, tcell.KeyTab, tcell.KeyEnter} {
		sim.InjectKey(k, 0, tcell.ModNone)
	}
	for _, r := range "+-0[]><mtg" {
		sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	}
	time.Sleep(300 * time.Millisecond)
	sim.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("app did not quit")
	}
	_ = theme.SetActive("forge")
}

func BenchmarkDraw200x60(b *testing.B) {
	lib := newLibrary("states")
	for _, age := range []string{"medieval_age", "atomic_age", "cyberpunk_age", "galactic_age"} {
		a, err := lib.atlas(age)
		if err != nil {
			b.Skip(err)
		}
		b.Run(age, func(b *testing.B) {
			sc := newScene(a)
			for i := 0; i < b.N; i++ {
				sc.Frame = i
				sc.Draw(200, 60)
			}
		})
	}
}
