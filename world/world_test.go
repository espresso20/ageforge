package world

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

var testCatalog = mapmodel.NewCatalog()

func model(age string, seed int64, o fixture.Options) *mapmodel.Model {
	o.Age, o.Seed = age, seed
	if o.Tick == 0 {
		o.Tick = 43200
	}
	st := fixture.State(o)
	return mapmodel.NewBuilder(testCatalog).Build(&st, nil)
}

func TestPackageIsPure(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	banned := []string{"github.com/gdamore/tcell", "github.com/rivo/tview", "github.com/hajimehoshi/ebiten",
		"github.com/espresso20/ageforge/ui", "github.com/espresso20/ageforge/theme", "github.com/espresso20/ageforge/mapgen"}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if slices.ContainsFunc(banned, func(b string) bool { return strings.HasPrefix(p, b) }) {
				t.Errorf("%s imports %s", path, p)
			}
		}
	}
	if _, err := os.Stat("doc.go"); err != nil {
		t.Error(err)
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	for _, age := range SpaceAges {
		a, b := Build(model(age, 7, fixture.Options{Harbinger: true}), 7), Build(model(age, 7, fixture.Options{Harbinger: true}), 7)
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s: two builds differ", age)
		}
	}
	if reflect.DeepEqual(Build(model(AgeGalactic, 7, fixture.Options{}), 7), Build(model(AgeGalactic, 8, fixture.Options{}), 8)) {
		t.Error("two seeds give the same sky")
	}
}

func TestGroundAgesHaveNoSky(t *testing.T) {
	if w := Build(model("stone_age", 7, fixture.Options{}), 7); w.Root != nil {
		t.Error("a ground age has a sky")
	}
}

// TestHomeSystem: every age that shows the system has the eight planets in
// order, the home world third, with its moon.
func TestHomeSystem(t *testing.T) {
	for _, age := range SpaceAges {
		w := Build(model(age, 7, fixture.Options{}), 7)
		sys := w.Root.Find(FrameSystem)
		if sys == nil {
			t.Fatalf("%s: no system", age)
		}
		var names []string
		for _, o := range sys.Objects {
			if o.Kind == KindPlanet && o.Body.Kind != "dwarf" {
				names = append(names, o.Key)
			}
		}
		if !slices.Equal(names, SolNames[:]) {
			t.Errorf("%s: planets %v", age, names)
		}
		home := w.Root.Find(FramePlanet)
		if home == nil || home.Count(KindMoon) != 1 || home.Count(KindShip) != 1 {
			t.Errorf("%s: home frame wrong", age)
		}
		if StarAge(age) != (w.Root.Kind <= FrameLocalGroup) {
			t.Errorf("%s: wrong root %v", age, w.Root.Kind)
		}
	}
}

func TestStarAgeLadder(t *testing.T) {
	w := Build(model(AgeGalactic, 7, fixture.Options{}), 7)
	f := w.Root
	for _, k := range []FrameKind{FrameLocalGroup, FrameGalaxy, FrameRegion, FrameNeighbourhood, FrameSystem, FramePlanet} {
		if f == nil || f.Kind != k {
			t.Fatalf("ladder broken at %v", k)
		}
		if len(f.Children) > 0 {
			f = f.Children[0]
		}
	}
	if n := w.Root.Count(KindStar); n < len(NearStars) {
		t.Errorf("%d stars", n)
	}
	if n := w.Root.Find(FrameRegion).Count(KindFeature); n < ArmObjects+11 {
		t.Errorf("%d arm features", n)
	}
	if w.Root.Count(KindGate) != 1 {
		t.Error("a star age's system has the warp gate")
	}
}

// TestGameStateGatesWhatIsLit: colonies, lanes, sector stars, rivals and the
// rift come from the state; stations of the ring only in the space age.
func TestGameStateGatesWhatIsLit(t *testing.T) {
	bare := model(AgeGalactic, 7, fixture.Options{Scale: 0.01})
	for _, l := range bare.Lineages {
		l.Count = 0
	}
	bare.Harbinger = nil
	for i := range bare.Factions {
		bare.Factions[i].Discovered = false
	}
	w := Build(bare, 7)
	for _, k := range []ObjKind{KindColony, KindLane, KindSector, KindRival, KindRift} {
		if n := w.Root.Count(k); n != 0 {
			t.Errorf("kind %d: %d objects with nothing built", k, n)
		}
	}
	full := model(AgeGalactic, 7, fixture.Options{Harbinger: true})
	wf := Build(full, 7)
	for _, k := range []ObjKind{KindColony, KindLane, KindSector, KindRival, KindRift} {
		if wf.Root.Count(k) == 0 {
			t.Errorf("kind %d: none with a grown town", k)
		}
	}
	mods := func(age string) int {
		n := 0
		for _, o := range Build(model(age, 7, fixture.Options{}), 7).Root.Find(FramePlanet).Objects {
			if o.Kind == KindStation && strings.HasPrefix(o.Key, "module_") {
				n++
			}
		}
		return n
	}
	if mods(AgeSpace) == 0 || mods(AgeInterstellar) != 0 {
		t.Errorf("ring modules: space %d, interstellar %d", mods(AgeSpace), mods(AgeInterstellar))
	}
	gate := func(m *mapmodel.Model) int { return Build(m, 7).Root.Count(KindGate) }
	m := model(AgeInterstellar, 7, fixture.Options{})
	m.Wonders = nil
	if gate(m) != 0 {
		t.Error("a gate without the wonder")
	}
	m.Wonders = []mapmodel.Wonder{{Key: GateWonder, Built: true}}
	if gate(m) != 1 {
		t.Error("no gate with the wonder")
	}
}

func TestChangeKey(t *testing.T) {
	m := model(AgeGalactic, 7, fixture.Options{})
	k := ChangeKey(m, 7)
	m.Tick += 5000
	if ChangeKey(m, 7) != k {
		t.Error("the tick moves the key")
	}
	if ChangeKey(m, 8) == k {
		t.Error("the seed does not move the key")
	}
	m.Harbinger = nil
	h := model(AgeGalactic, 7, fixture.Options{Harbinger: true})
	if ChangeKey(h, 7) == ChangeKey(m, 7) {
		t.Error("the harbinger does not move the key")
	}
	n := len(m.Lineages)
	if n == 0 {
		t.Fatal("no lineages")
	}
	before := ChangeKey(m, 7)
	m.Lineages[0].Count++
	if ChangeKey(m, 7) == before {
		t.Error("a building does not move the key")
	}
	if Build(m, 7).Key != ChangeKey(m, 7) {
		t.Error("World.Key is not the change key")
	}
}
