package game

import (
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Stock has no ceiling, so the engine keeps no store size and shows none: a
// resource is an amount and a rate, a save writes no store, and what a
// storage tech raised is production.

func TestStateHasNoStorageField(t *testing.T) {
	for _, v := range []any{ResourceState{}, ResearchState{}, Resource{}} {
		typ := reflect.TypeOf(v)
		for i := 0; i < typ.NumField(); i++ {
			if f := typ.Field(i); strings.Contains(strings.ToLower(f.Name), "storage") {
				t.Errorf("%s has a field %s", typ.Name(), f.Name)
			}
		}
	}
}

func TestASaveWritesNoStorage(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(3)
	ge.StepTicks(20)
	if err := ge.SaveGame("nostorage"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath("nostorage"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"storage"`) {
		t.Error("a save still writes the stores' sizes")
	}
	if err := newSeededEngine(4).LoadGame("nostorage"); err != nil {
		t.Errorf("a save with no stores did not load: %v", err)
	}
}

// The devmode /fill adds a large amount of every resource a town can save
// up; nothing else reads it.
func TestDevFillAddsALotOfEverythingSavable(t *testing.T) {
	ge := newSeededEngine(1)
	fill := ge.Resources.GetAllStorage()
	if fill["food"] != devFillAmount || fill["wood"] != devFillAmount {
		t.Errorf("fill = %v, want %v of food and wood", fill, devFillAmount)
	}
	if _, ok := fill["soldiers"]; ok {
		t.Error("fill lists soldiers, which are people and no resource")
	}
}

// The seven techs that raised every store raise all production, each by about
// a tenth of the share they raised storage by: the rate of a resource they do
// not name rises by exactly that share.
func TestConvertedTechsAddTheirProduction(t *testing.T) {
	for _, c := range []struct {
		tech string
		by   float64
	}{
		{"pottery", 0.008}, {"masonry", 0.008}, {"prefabrication", 0.004}, {"grid_storage", 0.004},
		{"cloud_computing", 0.006}, {"superconductors", 0.006}, {"space_elevator", 0.006},
	} {
		ge := newTruthEngine("bronze_age", truthClean)
		before := rateOf(ge, "iron")
		if before <= 0 {
			t.Fatalf("setup: the lab town makes no iron")
		}
		learn(ge, c.tech)
		after := rateOf(ge, "iron")
		if got := after/before - 1; math.Abs(got-c.by) > 1e-9 {
			t.Errorf("%s moves iron by %.4f, want %.4f", c.tech, got, c.by)
		}
	}
}
