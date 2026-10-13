package game

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

// Saves from before the raw materials were priced, loaded by the game after
// it. The fixtures were written and signed by the old game
// (raw_materials_fixture_gen_test.go says how):
//
//   - pre_raw_materials_modern.json: a Modern Age town with 4 Titanium Mines.
//   - pre_raw_materials_digital.json: a Digital Age town with 3 each of the
//     Titanium Mine, the Precision Mine and the Nano Drill Complex.
//   - pre_raw_materials_iron.json: an Iron Age town whose Colosseum bank was
//     full at the old price, with an Agora in the build queue and two plan
//     items (Timber Yard x2, and a Legion Fort with part of its price banked).
const (
	rawModernFixture  = "testdata/pre_raw_materials_modern.json"
	rawDigitalFixture = "testdata/pre_raw_materials_digital.json"
	rawIronFixture    = "testdata/pre_raw_materials_iron.json"
)

// loadSaveFile reads a fixture as the save it holds, without loading it.
func loadSaveFile(t *testing.T, path string) GameSave {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatal(err)
	}
	return save
}

// loadFixtureAway loads a fixture as if it had been saved away ago. The
// save's clock is the one thing rewritten: the load credits the time since the
// save (up to a day) as offline play, so a fixture loaded as it stands plays
// a longer absence every day it sits in the repository. The save is signed
// again over its new clock; every other byte is the old game's.
func loadFixtureAway(t *testing.T, path, name string, away time.Duration) *GameEngine {
	t.Helper()
	save := loadSaveFile(t, path)
	if sigValid, _ := verifySave(&save); !sigValid || save.Signature == "" {
		t.Fatalf("setup: the fixture %s is not signed and valid as the old game wrote it", path)
	}
	save.Timestamp = time.Now().Add(-away)
	save.Signature = signSave(save, saveHMACKey)
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath(name), data, 0o644); err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	if err := ge.LoadGame(name); err != nil {
		t.Fatal(err)
	}
	if ge.cheaterBadge {
		t.Fatalf("%s: the save fails its signature check after the load", path)
	}
	return ge
}

// TestRawMaterialsSave_TitaniumOreOpensAndTheMinesMakeIt: a save from before
// titanium ore opened in the Modern Age (it was shut until the Space Age, so
// the save's unlocked list lacks it) loads with the ore open and shown, and
// its titanium mines make it, at the rate the mines now state.
func TestRawMaterialsSave_TitaniumOreOpensAndTheMinesMakeIt(t *testing.T) {
	for _, c := range []struct {
		name, path, age string
		mines           map[string]int
	}{
		{"modern", rawModernFixture, "modern_age", map[string]int{"titanium_mine": 4}},
		{"digital", rawDigitalFixture, "digital_age", map[string]int{"titanium_mine": 3, "precision_mine": 3, "nano_drill_complex": 3}},
	} {
		t.Run(c.name, func(t *testing.T) {
			isolateAccountDir(t)
			save := loadSaveFile(t, c.path)
			for _, k := range save.Unlocked.Resources {
				if k == "titanium_ore" {
					t.Fatal("setup: the old save already lists titanium ore as open")
				}
			}
			if save.Age != c.age {
				t.Fatalf("setup: the fixture is in %s, want %s", save.Age, c.age)
			}

			ge := loadFixtureAway(t, c.path, "raws_"+c.name, 0)
			rs := ge.GetState().Resources["titanium_ore"]
			if !rs.Unlocked {
				t.Fatal("titanium ore is not open after the load")
			}
			if rs.Name != "Titanium Ore" {
				t.Errorf("titanium ore is shown as %q", rs.Name)
			}
			// Open is what the resource box draws (ui/resources_box.go lists
			// every resource whose Unlocked is set), so this is shown.
			want := 0.0
			for key, n := range c.mines {
				if got := ge.Buildings.GetCount(key); got != n {
					t.Fatalf("%s: %d after the load, the save has %d", key, got, n)
				}
				def, _ := ge.Buildings.defs[key]
				for _, e := range def.Effects {
					if e.Type == "production" && e.Target == "titanium_ore" {
						want += e.Value * float64(n)
					}
				}
			}
			if want <= 0 {
				t.Fatal("setup: the mines in the fixture state no titanium ore output")
			}
			ge.StepTicks(5)
			after := ge.GetState().Resources["titanium_ore"]
			t.Logf("%s: rate %.4g/s (mines state %.4g/s base), amount %.4g of %.4g after 5 ticks", c.name, after.Rate, want, after.Amount, after.Storage)
			if after.Rate <= 0 {
				t.Errorf("titanium ore has no rate after the load: %v", after.Rate)
			}
			if !(after.Amount > 0) {
				t.Errorf("the mines made no titanium ore in 5 ticks (amount %v of %v)", after.Amount, after.Storage)
			}
			if after.Amount > after.Storage+1e-6 {
				t.Errorf("titanium ore %v is over its store %v", after.Amount, after.Storage)
			}
		})
	}
}

// TestRawMaterialsSave_ColosseumBankKeepsWhatItHeldAndAsksTheRest: an Iron Age
// save whose Colosseum bank was full at the old price (stone, iron, gold)
// keeps the bank as saved. The wonder now also asks marble and iron ore, so
// the bank is no longer full, and what it lacks is exactly those two. Stone
// and iron are held at twice their new price: the surplus stays in the bank,
// is neither refunded nor counted toward marble or iron ore, and stays after
// the wonder is built.
func TestRawMaterialsSave_ColosseumBankKeepsWhatItHeldAndAsksTheRest(t *testing.T) {
	isolateAccountDir(t)
	save := loadSaveFile(t, rawIronFixture)
	oldBank := save.WonderBanks["colosseum"]
	if oldBank["stone"] != 512000 || oldBank["iron"] != 113600 || oldBank["gold"] != 100800 || len(oldBank) != 3 {
		t.Fatalf("setup: the fixture's Colosseum bank is %v", oldBank)
	}
	ge := loadFixtureAway(t, rawIronFixture, "raws_iron", 0)
	cost := ge.Buildings.defs["colosseum"].BaseCost
	for _, res := range []string{"stone", "marble", "iron", "iron_ore", "gold"} {
		if cost[res] <= 0 {
			t.Fatalf("setup: the Colosseum no longer asks %s (%v)", res, cost)
		}
	}

	// The bank is as saved, to the unit.
	bank := ge.Buildings.GetWonderBank("colosseum")
	if len(bank) != len(oldBank) {
		t.Errorf("the bank after the load is %v, the save held %v", bank, oldBank)
	}
	for res, amt := range oldBank {
		if bank[res] != amt {
			t.Errorf("bank %s is %v after the load, the save held %v", res, bank[res], amt)
		}
	}

	// It is no longer full, and lacks exactly the two new parts.
	if ge.Buildings.IsWonderBankFull("colosseum") {
		t.Fatal("the old full bank still reads full")
	}
	st := ge.GetState().Buildings["colosseum"]
	if st.WonderBankFull || st.CanBuild {
		t.Errorf("the Colosseum shows bank full %v, can build %v", st.WonderBankFull, st.CanBuild)
	}
	if err := ge.BuildBuilding("colosseum"); err == nil || !strings.Contains(err.Error(), "not fully banked") {
		t.Errorf("building the Colosseum with the old bank: %v", err)
	}
	rest := ge.wonderBankRest("colosseum")
	if len(rest) != 2 || rest["marble"] != cost["marble"] || rest["iron_ore"] != cost["iron_ore"] {
		t.Errorf("the bank lacks %v, want only marble %v and iron ore %v", rest, cost["marble"], cost["iron_ore"])
	}

	// Surplus: stone and iron are held beyond their new price. Banking more
	// of either is refused, and neither pays for marble or iron ore.
	surplus := map[string]float64{"stone": oldBank["stone"] - cost["stone"], "iron": oldBank["iron"] - cost["iron"]}
	if surplus["stone"] <= 0 || surplus["iron"] <= 0 {
		t.Fatalf("setup: no surplus (%v)", surplus)
	}
	setResource(ge, "stone", 1000)
	if _, err := ge.BankWonderResource("colosseum", "stone", 1000); err == nil || !strings.Contains(err.Error(), "already has all") {
		t.Errorf("banking stone into a bank that holds %v of %v: %v", bank["stone"], cost["stone"], err)
	}
	if got := ge.Resources.Get("stone"); got != 1000 {
		t.Errorf("the refused deposit still took stone: %v left of 1000", got)
	}

	// Save and load again: the bank, surplus included, is written back as it is.
	again := saveAndLoad(t, ge, "raws_iron_again")
	for res, amt := range oldBank {
		if got := again.Buildings.GetWonderBank("colosseum")[res]; got != amt {
			t.Errorf("after a second save, bank %s is %v, want %v", res, got, amt)
		}
	}

	// Bank the two new parts: now it is full, with the surplus still in it.
	for _, res := range []string{"marble", "iron_ore"} {
		setResource(ge, res, cost[res])
		got, err := ge.BankWonderMax("colosseum", res)
		if err != nil || got != cost[res] {
			t.Fatalf("banking %s: put in %v, err %v (want %v)", res, got, err, cost[res])
		}
	}
	if !ge.Buildings.IsWonderBankFull("colosseum") {
		t.Fatal("the bank is not full with every part in")
	}
	b2 := ge.Buildings.GetWonderBank("colosseum")
	for res, want := range map[string]float64{"stone": oldBank["stone"], "iron": oldBank["iron"], "marble": cost["marble"], "iron_ore": cost["iron_ore"], "gold": cost["gold"]} {
		if b2[res] != want {
			t.Errorf("full bank: %s is %v, want %v", res, b2[res], want)
		}
	}
	// Building it takes nothing from the stores and gives nothing back, and
	// the surplus is still recorded in the bank after the wonder stands.
	setResource(ge, "stone", 5000)
	if err := ge.BuildBuilding("colosseum"); err != nil {
		t.Fatalf("building the Colosseum from a full bank: %v", err)
	}
	for i := 0; i < 5000 && ge.Buildings.GetCount("colosseum") == 0; i++ {
		ge.StepTicks(1)
	}
	if ge.Buildings.GetCount("colosseum") != 1 {
		t.Fatal("the Colosseum did not finish")
	}
	if got := ge.Buildings.GetWonderBank("colosseum")["stone"]; got != oldBank["stone"] {
		t.Errorf("after the build the bank holds %v stone, want the surplus kept (%v)", got, oldBank["stone"])
	}
	t.Logf("surplus left in the bank after the build: stone %.6g, iron %.6g", surplus["stone"], surplus["iron"])
}

// TestRawMaterialsSave_QueueAndPlanSeeTheNewPrice: a building in the queue
// finishes without a further charge (it was paid when it was queued), and the
// plan's items for buildings that now cost a raw show the new price. A plan
// item's bank of stone and iron pays only those parts of the new price, up to
// each part, and is credited to no other.
func TestRawMaterialsSave_QueueAndPlanSeeTheNewPrice(t *testing.T) {
	isolateAccountDir(t)
	ge := loadFixtureAway(t, rawIronFixture, "raws_iron_plan", 0)
	if len(ge.buildQueue) != 1 || ge.buildQueue[0].BuildingKey != "agora" {
		t.Fatalf("the build queue after the load is %+v", ge.buildQueue)
	}

	state := ge.GetState()
	if len(state.Plan) != 2 {
		t.Fatalf("%d plan items after the load, want 2", len(state.Plan))
	}
	for _, v := range state.Plan {
		cost, _ := ge.Buildings.BuildBatchCost(v.Key, 1, ge.buildQueue)
		if cost["marble"] <= 0 || cost["iron_ore"] <= 0 {
			t.Fatalf("setup: %s does not cost marble and iron ore (%v)", v.Key, cost)
		}
		if len(v.Cost) != len(cost) {
			t.Errorf("plan %s shows %v, the price is %v", v.Key, v.Cost, cost)
		}
		for res, c := range cost {
			if math.Abs(v.Cost[res]-c) > 1e-6 {
				t.Errorf("plan %s shows %s %v, the price is %v", v.Key, res, v.Cost[res], c)
			}
		}
	}

	// The Legion Fort's bank (14,000 stone, 7,000 iron) pays up to each part
	// of the new price and nothing else.
	fort := state.Plan[1]
	if fort.Key != "legion_fort" {
		t.Fatalf("the second plan item is %s", fort.Key)
	}
	for res, banked := range map[string]float64{"stone": 14000, "iron": 7000} {
		if want := math.Min(banked, fort.Cost[res]); math.Abs(fort.Banked[res]-want) > 1e-6 {
			t.Errorf("legion fort: %s from the bank is %v, want %v", res, fort.Banked[res], want)
		}
	}
	if fort.Banked["marble"] != 0 || fort.Banked["iron_ore"] != 0 || len(fort.Banked) != 2 {
		t.Errorf("legion fort: the bank is credited to another part: %v", fort.Banked)
	}

	// The queued Agora finishes, and the stores are not charged for it again.
	before := map[string]float64{}
	for _, res := range []string{"stone", "iron", "gold", "marble", "iron_ore"} {
		before[res] = ge.Resources.Get(res)
	}
	ge.StepTicks(ge.buildQueue[0].TicksLeft + 2)
	if got := ge.Buildings.GetCount("agora"); got != 1 {
		t.Fatalf("the queued Agora count is %d after its time", got)
	}
	for _, res := range []string{"stone", "iron", "gold"} {
		if got := ge.Resources.Get(res); got < before[res]-1e-6 {
			t.Errorf("finishing the queued Agora took %s: %v before, %v after", res, before[res], got)
		}
	}

	// The plan then waits on a part of the new price that the town lacks, and
	// says which; nothing was started on the old price.
	for _, v := range ge.GetState().Plan {
		if v.Started != 0 {
			t.Errorf("plan %s started %d copies on the old price", v.Key, v.Started)
		}
		if v.Status == PlanStatusReady {
			continue
		}
		if v.Short == "" && v.Status == PlanStatusWaiting {
			t.Errorf("plan %s waits and names nothing short", v.Key)
		}
		t.Logf("plan %s: %s, short of %q, note %q", v.Key, v.Status, v.Short, v.Note)
	}
}
