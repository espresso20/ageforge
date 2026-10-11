package smoke

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// victorianStall is the nightly's seed 1 after a Nuclear Exchange took both
// Industrial Depots on the way into the Victorian Age: 130M caps, a first
// Victorian Vault at 193M steel (build_cost discounts included), and an
// academy requirement priced past either.
func victorianStall(capacity float64) game.GameState {
	res := map[string]game.ResourceState{}
	for _, r := range []string{"coal", "electricity", "gold", "iron", "steel"} {
		res[r] = game.ResourceState{Storage: capacity, Amount: capacity, Unlocked: true}
	}
	return game.GameState{
		Age:            "victorian_age",
		NextAge:        "electric_age",
		NextAgeBldReqs: map[string]int{"academy": 10},
		Resources:      res,
		Buildings: map[string]game.BuildingState{
			"victorian_vault": {Name: "Victorian Vault", Unlocked: true,
				NextCost: map[string]float64{"steel": 193.2e6, "gold": 156.4e6, "iron": 119.6e6}},
			// An older age's storage: the age lock keeps it off the ladder.
			"industrial_depot": {Name: "Industrial Depot", Unlocked: true,
				NextCost: map[string]float64{"steel": 24e6, "iron": 31e6, "coal": 15.6e6}},
			"academy": {Name: "Academy", Unlocked: true,
				NextCost: map[string]float64{"steel": 488e6, "gold": 248e6, "iron": 166e6}},
		},
	}
}

func problemChecks(ps []problem) map[string]string {
	out := map[string]string{}
	for _, p := range ps {
		out[p.check] = p.msg
	}
	return out
}

// The stall the nightly took 36 simulated hours to time out on is named at
// the first sweep: the vault can't fit, so no copy of it adds anything, and
// the academies the gate asks for are over the store the player can reach.
// That, and not the wall itself, is what makes it a stall: a walled store is
// the normal state (see the tests below).
func TestStorageLadder_NamesAStorageStall(t *testing.T) {
	st := victorianStall(130e6)
	got := problemChecks(storageProblems(st, config.BuildingByKey()))
	msg, ok := got["storage_unreachable"]
	if !ok {
		t.Fatalf("no storage_unreachable problem; got %v", got)
	}
	for _, want := range []string{"victorian_vault", "193M steel", "130M cap", "academy #10"} {
		if !strings.Contains(msg, want) {
			t.Errorf("storage_unreachable message %q lacks %q", msg, want)
		}
	}
	if _, ok := got["required_building_over_storage"]; !ok {
		t.Errorf("the academies' last copy should be over the reachable storage; got %v", got)
	}
	if b := Blockers(st); !strings.HasPrefix(b, "storage stuck (next victorian_vault costs 193M steel, over the 130M cap)") {
		t.Errorf("Blockers should lead with the storage stall, got %q", b)
	}
}

// With the two depots standing (478M), the vault fits and the ladder climbs
// copy by copy until the next copy costs more than the store the copies
// before it give: a wall, worked out here by hand from the vault's size and
// its 1.75 climb. The academies are under that store, so nothing is flagged.
func TestStorageLadder_ClimbsWhenTheFirstCopyFits(t *testing.T) {
	st := victorianStall(478e6)
	defs := config.BuildingByKey()
	if ps := storageProblems(st, defs); len(ps) > 0 {
		t.Fatalf("want no problems with the vault in reach, got %v", ps)
	}
	caps, stall := storageLadder(st, defs)
	if stall != nil {
		t.Fatalf("unexpected stall %+v", *stall)
	}
	vault := defs["victorian_vault"]
	store, price, copies := 478e6, st.Buildings["victorian_vault"].NextCost["steel"], 0
	for price <= store {
		store += vault.Effects[0].Value
		price *= vault.CostScale
		copies++
	}
	if copies < 3 || math.IsInf(caps["steel"], 0) || math.Abs(caps["steel"]-store) > 1e-9*store {
		t.Errorf("steel reachable = %.6g, want %.6g (%d copies of the vault, then a copy at %.6g that does not fit)", caps["steel"], store, copies, price)
	}
	if strings.Contains(Blockers(st), "storage stuck") {
		t.Errorf("Blockers names a stall that isn't there: %q", Blockers(st))
	}
}

// A vault already under construction is paid for: its storage counts, and
// the next copy (priced past the queued one) fits under it.
func TestStorageLadder_QueuedStorageCounts(t *testing.T) {
	st := victorianStall(130e6)
	st.BuildQueue = []game.BuildQueueSnapshot{{Name: "Victorian Vault", TicksLeft: 100, TotalTicks: 500}}
	vault := st.Buildings["victorian_vault"]
	next := map[string]float64{}
	for r, c := range vault.NextCost {
		next[r] = c * 1.13
	}
	vault.NextCost = next
	st.Buildings["victorian_vault"] = vault
	defs := config.BuildingByKey()
	caps, stall := storageLadder(st, defs)
	if stall != nil {
		t.Errorf("a queued vault should lift the caps, got stall %+v", *stall)
	}
	if want := 130e6 + defs["victorian_vault"].Effects[0].Value; caps["steel"] <= want {
		t.Errorf("steel reachable = %.6g, want more than the queued vault's %.6g", caps["steel"], want)
	}
}

// A walled store is not a stall. Storage has no copy limit since the storage
// rule; the last copy a player can buy is the one that fits, and the next one
// costs more than the cap. With the vault over the cap and nothing the gate
// asks for over it, there is nothing the player cannot buy.
func TestStorageLadder_MaxedStorageIsNoStall(t *testing.T) {
	st := victorianStall(130e6)
	st.NextAgeBldReqs = map[string]int{}
	defs := config.BuildingByKey()
	if _, stall := storageLadder(st, defs); stall != nil {
		t.Errorf("a walled store with nothing over it reported as a stall: %+v", *stall)
	}
	if ps := storageProblems(st, defs); len(ps) > 0 {
		t.Errorf("a walled store with nothing over it: want no problems, got %v", ps)
	}
	if strings.Contains(Blockers(st), "storage stuck") {
		t.Errorf("Blockers names a stall that isn't there: %q", Blockers(st))
	}
	// The same wall with a requirement over it is the stall.
	st.NextAgeBldReqs = map[string]int{"academy": 10}
	if _, stall := storageLadder(st, defs); stall == nil {
		t.Errorf("10 academies asked for, the last over a 130M cap with the vault out of reach: want a stall")
	}
}

// A required copy in the build queue is bought: the next cost the game shows
// is already the price of the copy after it, so a gate that asks for five
// Warehouses is not blocked by the price of a sixth while the fifth is
// building (the perf scenario's long run was flagged for this).
func TestRequiredCopyInTheQueueIsBought(t *testing.T) {
	defs := config.BuildingByKey()
	st := victorianStall(130e6)
	st.NextAgeBldReqs = map[string]int{"victorian_vault": 1}
	vault := st.Buildings["victorian_vault"]
	next := map[string]float64{}
	for r, c := range vault.NextCost {
		next[r] = c * 100 // far over any store: the price of the copy after the queued one
	}
	vault.NextCost = next
	st.Buildings["victorian_vault"] = vault

	if _, ok := problemChecks(storageProblems(st, defs))["required_building_over_storage"]; !ok {
		t.Fatalf("with the required vault unbought and priced over every store, want it flagged")
	}
	st.BuildQueue = []game.BuildQueueSnapshot{{Name: "Victorian Vault", TicksLeft: 100, TotalTicks: 500}}
	if msg, ok := problemChecks(storageProblems(st, defs))["required_building_over_storage"]; ok {
		t.Errorf("the one vault the gate asks for is in the queue; flagged: %s", msg)
	}
}
