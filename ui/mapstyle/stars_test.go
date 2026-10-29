package mapstyle

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestSectorsAndLanesInspectByName: on the star map a sector names its line,
// its systems and its biggest buildings, a lane its ends and its traffic, and
// the border says whose it is.
func TestSectorsAndLanesInspectByName(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range []string{"galactic_age", "transcendent_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 6, Wars: 1, Routes: 3})
		v := New().(*view)
		f := Frame{Model: m, Anim: 3, Clock: 3}
		sectors, lanes, borders := 0, 0, 0
		for th, p := range thingsOf(v.scene(f).Grid) {
			v.cur = p
			in, _ := v.Inspect(f)
			switch th.Kind {
			case mapgen.KindSector:
				if th.Key == "border" {
					borders++
					continue
				}
				sectors++
				if !strings.HasPrefix(in.Title, "Sector: "+placeName(m, th.Key)+" · ") || !strings.Contains(in.Title, "system") || len(in.Lines) == 0 || !strings.Contains(in.Lines[0], "×") {
					t.Errorf("%s: a sector of %s inspects as %+v", age, th.Key, in)
				}
			case mapgen.KindLane:
				lanes++
				if !strings.HasPrefix(in.Title, "Lane: ") || !strings.Contains(in.Title, " to ") || len(in.Lines) != 1 || !strings.HasPrefix(in.Lines[0], "traffic ") {
					t.Errorf("%s: a lane inspects as %+v", age, in)
				}
			}
		}
		if sectors < 5 || lanes < 3 || borders == 0 {
			t.Errorf("%s: %d sectors, %d lanes, %d border cells", age, sectors, lanes, borders)
		}
	}
}

// TestTheCursorVisitsSectorsAndRivals: Tab on the star map lands on the hub of
// every sector with something built and on every civ met.
func TestTheCursorVisitsSectorsAndRivals(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "transcendent_age", Seed: 6, Wars: 1, Routes: 3})
	sc := New().(*view).scene(Frame{Model: m})
	hubs, civs := 0, 0
	for _, p := range sc.Targets {
		switch th := sc.Grid.ThingAt(p.X, p.Y); {
		case th.Kind == mapgen.KindSector && th.Ref == 1:
			hubs++
		case th.Kind == mapgen.KindCiv:
			civs++
		}
	}
	if hubs != len(m.Lineages) || civs != m.DiscoveredCount() || len(sc.Districts) != len(m.Lineages) {
		t.Errorf("%d hubs, %d civs and %d districts for %d lines and %d civs", hubs, civs, len(sc.Districts), len(m.Lineages), m.DiscoveredCount())
	}
}
