package game

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"testing"
	"time"
)

// olderBuildingsGame builds a Bronze Age game the way live play gets there:
// every Primitive Age building built, then two age advances. Gathering camps,
// huts and the rest are older buildings in it.
func olderBuildingsGame(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	for key, def := range ge.Buildings.defs {
		if def.RequiredAge == "primitive_age" && def.Category != "wonder" && ge.Buildings.IsUnlocked(key) {
			ge.Buildings.counts[key] = 2
		}
	}
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	ge.advanceAge("bronze_age")
	ge.pendingCatastrophe = ""
	if !ge.Buildings.IsOlder("gathering_camp") {
		t.Fatal("setup: gathering_camp is not older in the Bronze Age")
	}
	return ge
}

// TestSaveGame_ByteStable: two saves of the same state must be the same
// bytes. A list written in map order would not be (the smoke suite's
// save_map_order warning). Map iteration order is randomized per range, so a
// map-ordered slice shows up within a few tries.
func TestSaveGame_ByteStable(t *testing.T) {
	isolateAccountDir(t)
	ge := olderBuildingsGame(t)
	ge.StepTicks(50)
	if n := len(ge.Buildings.OlderKeys()); n < 3 {
		t.Fatalf("setup: only %d older buildings; the test needs several to catch map order", n)
	}

	marshal := func() []byte {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		s := ge.buildSaveSnapshot()
		s.Timestamp = time.Time{}
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	first := marshal()
	for i := 0; i < 20; i++ {
		if got := marshal(); !bytes.Equal(got, first) {
			t.Fatalf("save %d differs from the first save of the same state", i+2)
		}
	}

	// And through SaveGame itself, where only the write time (and the
	// signature over it) may differ.
	stamp := regexp.MustCompile(`(?m)^  "(timestamp|_sig|_proof)": .*$`)
	read := func(name string) []byte {
		if err := ge.SaveGame(name); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(savePath(name))
		if err != nil {
			t.Fatal(err)
		}
		return stamp.ReplaceAll(data, nil)
	}
	a := read("bytes-a")
	for i := 0; i < 5; i++ {
		if b := read("bytes-b"); !bytes.Equal(a, b) {
			t.Fatalf("SaveGame wrote different bytes for the same state (try %d)", i+1)
		}
	}
}
