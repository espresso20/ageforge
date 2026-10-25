package game

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// savedFixture reads a save fixture's own fields, as the file holds them.
func savedFixture(t *testing.T, path string) GameSave {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var gs GameSave
	noErr(t, json.Unmarshal(data, &gs))
	return gs
}

// TestSaveFromBeforeTheBiggerTownLoads: a save written when an age was
// priced for a town of five loads whole under the town of eight. What it
// built stands and its wonder bank holds what it held; it simply reads the
// new prices (the wonder costs the tables' price, so a bank that was half
// full is now under half) and the new requirements (the gate asks the
// tables' counts). Nothing is topped up or taken back.
func TestSaveFromBeforeTheBiggerTownLoads(t *testing.T) {
	saved := savedFixture(t, keystoneFixtureBanked)
	if len(saved.Buildings) == 0 || len(saved.WonderBanks["colosseum"]) == 0 {
		t.Fatal("the fixture holds no buildings or no Colosseum bank")
	}
	ge := loadKeystoneFixture(t, keystoneFixtureBanked, "pre_bigger_town_iron")

	// What it built stands.
	for key, n := range saved.Buildings {
		if got := ge.Buildings.GetCount(key); got != n {
			t.Errorf("the save held %d %s, the loaded game has %d", n, key, got)
		}
	}
	// The bank holds what the save's did, part by part.
	def := ge.Buildings.defs["colosseum"]
	for res, had := range saved.WonderBanks["colosseum"] {
		if got := ge.Buildings.wonderBanks["colosseum"][res]; got != had {
			t.Errorf("the Colosseum's bank holds %v %s, the save held %v", got, res, had)
		}
	}
	// The price is the tables' price now, and the bank that was half of the
	// old one is short of it.
	if want := config.BuildingByKey()["colosseum"].BaseCost; len(def.BaseCost) != len(want) {
		t.Fatalf("the game prices the Colosseum in %d parts, the tables in %d", len(def.BaseCost), len(want))
	}
	for res, need := range config.BuildingByKey()["colosseum"].BaseCost {
		if got := def.BaseCost[res]; got != need {
			t.Errorf("the Colosseum costs %v %s, the tables say %v", got, res, need)
		}
		// A worked form shares its price with its raw (iron with iron ore):
		// the save is from before the split, so it is measured against both.
		whole := need
		for form, raw := range map[string]string{"iron": "iron_ore", "stone": "marble"} {
			if res == form {
				whole += config.BuildingByKey()["colosseum"].BaseCost[raw]
			}
		}
		if had := saved.WonderBanks["colosseum"][res]; had >= whole/2 {
			t.Errorf("the save's bank of %v %s is %v of the new price: the fixture does not predate the change", had, res, had/whole)
		}
	}
	if ge.Buildings.IsWonderBankFull("colosseum") {
		t.Error("the Colosseum's bank is full: the load topped it up")
	}
	// The gate asks for the tables' counts.
	next := config.AgeByKey()[ge.progress.GetNextAge(ge.age)]
	got := ge.progress.GetRequirementsForNext(ge.age)
	if len(got) == 0 || len(got) != len(next.BuildingReqs) {
		t.Fatalf("the gate out of %s asks for %d buildings, the tables ask for %d", ge.age, len(got), len(next.BuildingReqs))
	}
	towns := 0
	for key, n := range next.BuildingReqs {
		if got[key] != n {
			t.Errorf("the gate asks for %d %s, the tables say %d", got[key], key, n)
		}
		if n >= config.TownCopies {
			towns++
		}
	}
	if towns == 0 {
		t.Errorf("no building in the gate out of %s is asked for in %d copies", ge.age, config.TownCopies)
	}
}
