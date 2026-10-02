package smoke

import (
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
		NextAgeResReqs: map[string]float64{"electricity": 1.1e6},
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
// the first sweep: the vault can't fit, so the 25 copies the old check
// counted as buildable add nothing, and the academies are out of reach too.
func TestStorageLadder_NamesAStorageStall(t *testing.T) {
	st := victorianStall(130e6)
	got := problemChecks(storageProblems(st, config.BuildingByKey()))
	msg, ok := got["storage_unreachable"]
	if !ok {
		t.Fatalf("no storage_unreachable problem; got %v", got)
	}
	for _, want := range []string{"victorian_vault", "193M steel", "130M cap"} {
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

// With the two depots standing (478M), the vault fits, every copy after it
// fits the caps the ones before raised, and nothing is flagged.
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
	if want := 478e6 + float64(vault.MaxCount)*vault.Effects[0].Value; caps["steel"] < want*(1-1e-9) {
		t.Errorf("steel reachable = %.4g, want all %d vaults: %.4g", caps["steel"], vault.MaxCount, want)
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
	if _, stall := storageLadder(st, config.BuildingByKey()); stall != nil {
		t.Errorf("a queued vault should lift the caps, got stall %+v", *stall)
	}
}

// All copies built is a full store, not a stall.
func TestStorageLadder_MaxedStorageIsNoStall(t *testing.T) {
	st := victorianStall(130e6)
	vault := st.Buildings["victorian_vault"]
	vault.Count = config.BuildingByKey()["victorian_vault"].MaxCount
	vault.AtMaxCount = true
	st.Buildings["victorian_vault"] = vault
	if _, stall := storageLadder(st, config.BuildingByKey()); stall != nil {
		t.Errorf("maxed storage reported as a stall: %+v", *stall)
	}
}
