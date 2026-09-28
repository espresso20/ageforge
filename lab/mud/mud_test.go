package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestEveryStateWalks loads every snapshot, checks that every place is
// reachable from the square, walks every road, and draws both forms at a
// spread of sizes (including silly ones) without panicking.
func TestEveryStateWalks(t *testing.T) {
	files, _ := filepath.Glob("states/*.json.gz")
	if len(files) == 0 {
		t.Skip("no states; run go run ./lab/mud -gen")
	}
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".json.gz")
		st, err := loadState("states", name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		c := BuildCity(st)
		for _, k := range c.Order {
			if k != Square && len(c.Route(Square, k)) == 0 {
				t.Errorf("%s: %s unreachable from the square", name, k)
			}
		}
		v := NewView(c, nil)
		for _, k := range c.Order {
			v.Here = k
			for d := North; d <= West; d++ {
				if to, ok := c.Places[k].Exits[d]; ok {
					if back := c.Places[to].Exits[d.Opposite()]; back != k {
						t.Errorf("%s: road %s %s from %s does not lead back", name, c.DirWord(d), to, k)
					}
				}
			}
			for _, size := range [][2]int{{120, 36}, {100, 30}, {80, 24}, {40, 15}, {20, 8}, {200, 60}} {
				s.SetSize(size[0], size[1])
				v.Draw(s)
				v.DrawMini(s)
			}
		}
		for _, cmd := range []string{"survey", "go gate", "out", "visit farm", "x nothing", "help", "away", "blah"} {
			v.Exec(cmd)
			v.Draw(s)
		}
	}
}

func TestMain(m *testing.M) {
	// Tests run in the package directory; captures and states are relative to it.
	os.Exit(m.Run())
}
