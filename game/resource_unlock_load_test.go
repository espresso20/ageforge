package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestLoadGame_ResourceUnlocksMatchLivePlay loads a save from every age and
// checks the unlocked resources are exactly what live play has there. Coal
// was the one that differed: the load path unlocked it in the Iron Age from
// ResourceDef.Age, while the age table (and live play) unlocks it in the
// Renaissance Age.
func TestLoadGame_ResourceUnlocksMatchLivePlay(t *testing.T) {
	isolateAccountDir(t)
	for _, age := range config.AgeOrder() {
		ge := newSeededEngine(1)
		for _, a := range config.AgeOrder() {
			ge.applyAgeUnlocks(a)
			if a == age {
				break
			}
		}
		ge.age = age
		ge.currentEpoch = config.EpochForAge(age)
		if err := ge.SaveGame("unlocks"); err != nil {
			t.Fatal(err)
		}
		loaded := NewGameEngine()
		if err := loaded.LoadGame("unlocks"); err != nil {
			t.Fatal(err)
		}
		for _, def := range config.BaseResources() {
			live, got := ge.Resources.IsUnlocked(def.Key), loaded.Resources.IsUnlocked(def.Key)
			if live != got {
				t.Errorf("%s: %s unlocked live=%v, after load=%v", age, def.Key, live, got)
			}
		}
		if age == "iron_age" && loaded.Resources.IsUnlocked("coal") {
			t.Error("loading an Iron Age save unlocked coal")
		}
	}
}
