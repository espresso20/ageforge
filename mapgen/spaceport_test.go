package mapgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapmodel/fixture"
)

func portAt(scale float64, gate bool) *Scene {
	st := fixture.State(fixture.Options{Age: "interstellar_age", Seed: 7, Tick: noon, Scale: scale})
	m := model(st)
	for i := range m.Wonders {
		if m.Wonders[i].Key == gateBuilding {
			m.Wonders[i].Built = gate
		}
	}
	line := "housing"
	return (&Generator{}).DrawAt(m, Moment{Anim: 40, Visit: -1, Close: true, District: line})
}

func portCells(g *Grid, keep func(Thing) bool) int {
	n := 0
	for _, h := range g.Hit {
		if keep(g.Things[h]) {
			n++
		}
	}
	return n
}

// The port is drawn in quadrant blocks, with no plot or lot underneath.
func TestPortIsPixelsWithoutPlots(t *testing.T) {
	g := portAt(1, true).Grid
	quad := 0
	for _, c := range g.Cells {
		if strings.ContainsRune("▘▝▀▖▌▞▛▗▚▐▜▄▙▟█", c.R) {
			quad++
		}
	}
	if quad < 200 {
		t.Errorf("only %d quadrant cells", quad)
	}
	if n := portCells(g, func(th Thing) bool { return th.Kind == KindAmbient || th.Kind == KindRoad }); n != 0 {
		t.Errorf("%d plot or square cells in the port", n)
	}
}

// More buildings give more modules.
func TestPortGrowsWithTheBuildings(t *testing.T) {
	isBuilding := func(th Thing) bool { return th.Kind == KindBuilding }
	few := portCells(portAt(0.2, false).Grid, isBuilding)
	many := portCells(portAt(1, false).Grid, isBuilding)
	if few == 0 || many <= few {
		t.Errorf("module cells: %d for few, %d for many", few, many)
	}
}

// The cursor over a module names one of the player's buildings.
func TestPortModulesNameRealBuildings(t *testing.T) {
	sc := portAt(1, false)
	m := model(fixture.State(fixture.Options{Age: "interstellar_age", Seed: 7, Tick: noon}))
	n := 0
	for _, h := range sc.Grid.Hit {
		th := sc.Grid.Things[h]
		if th.Kind != KindBuilding {
			continue
		}
		n++
		if b := m.Building(th.Key); b == nil || b.Count == 0 {
			t.Fatalf("a module names %q, which the player does not have", th.Key)
		}
	}
	if n == 0 {
		t.Error("no modules")
	}
}

// The warp gate is on the ring only when the player has it.
func TestPortGateNeedsTheNexus(t *testing.T) {
	isGate := func(th Thing) bool { return th.Key == "warp_gate" }
	if portCells(portAt(1, false).Grid, isGate) != 0 {
		t.Error("a gate without the nexus")
	}
	if portCells(portAt(1, true).Grid, isGate) == 0 {
		t.Error("no gate with the nexus")
	}
}

func TestPortSameSeedSameFrame(t *testing.T) {
	a, b := portAt(1, true).Grid, portAt(1, true).Grid
	if !reflect.DeepEqual(a.Cells, b.Cells) || !reflect.DeepEqual(a.Hit, b.Hit) {
		t.Error("the port is not deterministic")
	}
}

// A planet with moons shows them as satellites close to its disc at the inner
// zoom.
func TestInnerPlanetsHaveMoons(t *testing.T) {
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 40)
	g := sc.OverviewInner(160, 48).Grid
	sat, planet := 0, 0
	for _, h := range g.Hit {
		switch th := g.Things[h]; {
		case th.Kind == KindTerrain && th.Key == "satellite":
			sat++
		case th.Kind == KindTerrain && th.Key == "planet":
			planet++
		}
	}
	if sat == 0 || planet == 0 {
		t.Errorf("%d satellite cells, %d planet cells", sat, planet)
	}
}
