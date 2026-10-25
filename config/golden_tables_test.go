package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite config/testdata/golden_tables.json from the live tables")

// goldenBuilding is every number a building carries into play.
type goldenBuilding struct {
	Category   string             `json:"category"`
	Age        string             `json:"age"`
	Cost       map[string]float64 `json:"cost"`
	Rate       float64            `json:"rate"`
	Effects    []goldenEffect     `json:"effects"`
	BuildTicks int                `json:"build_ticks"`
	Workers    int                `json:"workers"`
	// The lineage fields sit before MaxCount so a record written before they
	// existed is only added to, never changed. LineKey, WorkerDomain and
	// OutputResource are left out when empty.
	LineageKey     string  `json:"lineage_key"`
	LineageTier    int     `json:"lineage_tier"`
	LineKey        string  `json:"line_key,omitempty"`
	WorkerDomain   string  `json:"worker_domain,omitempty"`
	OutputResource string  `json:"output_resource,omitempty"`
	CostScale      float64 `json:"cost_scale"`
	MaxCount       int     `json:"max_count"`
	Soldiers       int     `json:"soldiers,omitempty"`
	// FullRateCopies is a knowledge building's soft cap: the copies that
	// research at full rate, half past it.
	FullRateCopies int `json:"full_rate_copies,omitempty"`
}

type goldenEffect struct {
	Type   string  `json:"type"`
	Target string  `json:"target"`
	Value  float64 `json:"value"`
}

// goldenTech is a tech's place in the tree and its price. Cost stays last so
// a record written before the tree fields existed is only added to. Needs is
// Prerequisites in declared order; AnyOf is the either-or group in declared
// order. Both are left out when empty.
type goldenTech struct {
	Age   string   `json:"age"`
	Lane  string   `json:"lane"`
	Kind  TechKind `json:"kind"`
	Needs []string `json:"needs,omitempty"`
	AnyOf []string `json:"any_of,omitempty"`
	Cost  float64  `json:"cost"`
}

type goldenGate struct {
	Buildings map[string]int `json:"buildings"`
}

// goldenTables is the record of every number the economy's tables hold once
// the game has loaded them: prices, rates, outputs, build and
// research times, gates, and the price levels worked out from them.
type goldenTables struct {
	Buildings     map[string]goldenBuilding     `json:"buildings"`
	Techs         map[string]goldenTech         `json:"techs"`
	Gates         map[string]goldenGate         `json:"gates"`
	AgeTargetSecs map[string]float64            `json:"age_target_seconds"`
	PriceLevels   map[string]map[string]float64 `json:"price_levels"`
	TypicalIncome map[string]map[string]float64 `json:"typical_income"`
	EpochEvents   map[string]int                `json:"epoch_event_ticks"`
	Awakenings    map[string]int                `json:"awakening_ticks"`
	TradeRoutes   map[string]int                `json:"trade_route_ticks"`
	DealLevels    map[string]map[string]float64 `json:"deal_flow_levels"`
}

func liveGoldenTables() goldenTables {
	g := goldenTables{
		Buildings: map[string]goldenBuilding{}, Techs: map[string]goldenTech{}, Gates: map[string]goldenGate{},
		AgeTargetSecs: map[string]float64{}, PriceLevels: map[string]map[string]float64{},
		EpochEvents: map[string]int{}, Awakenings: map[string]int{}, TradeRoutes: map[string]int{},
	}
	defs := BaseBuildings()
	for _, d := range defs {
		b := goldenBuilding{Category: d.Category, Age: d.RequiredAge, Cost: d.BaseCost, Rate: d.CostScale,
			BuildTicks: d.BuildTicks, Workers: d.WorkerCapacity, MaxCount: d.MaxCount, Soldiers: d.Soldiers, FullRateCopies: d.FullRateCopies, Effects: []goldenEffect{},
			LineageKey: d.LineageKey, LineageTier: d.LineageTier, LineKey: d.LineKey, WorkerDomain: d.WorkerDomain, OutputResource: d.OutputResource, CostScale: d.CostScale}
		for _, e := range d.Effects {
			b.Effects = append(b.Effects, goldenEffect{Type: e.Type, Target: e.Target, Value: e.Value})
		}
		g.Buildings[d.Key] = b
	}
	techs := Technologies()
	kinds := TechKinds(techs, defs)
	for _, t := range techs {
		g.Techs[t.Key] = goldenTech{Age: t.Age, Lane: t.Lane, Kind: kinds[t.Key], Needs: t.Prerequisites, AnyOf: t.AnyOf, Cost: t.Cost}
	}
	order := AgeOrder()
	for _, a := range Ages() {
		g.Gates[a.Key] = goldenGate{Buildings: a.BuildingReqs}
		g.AgeTargetSecs[a.Key] = AgeTargets[a.Key].Seconds()
		g.PriceLevels[a.Key] = PriceLevels(a.Key)
	}
	g.TypicalIncome = Incomes(defs, techs, order, AnyResource)
	g.DealLevels = FlowDealLevels(defs, AgePositions(order))
	for _, e := range append(GoodEpochEvents(), ChallengingEpochEvents()...) {
		g.EpochEvents[e.Key] = e.Duration
	}
	for _, a := range Awakenings() {
		g.Awakenings[a.Key] = a.Duration
	}
	for _, r := range BaseTradeRoutes() {
		g.TradeRoutes[r.Key] = r.TicksPerRun
	}
	return g
}

// TestGoldenTables holds every number in the economy's tables to the record
// in testdata/golden_tables.json, to the unit. A change to any price, rate,
// output, size, time or gate shows up here by name; one that is meant is
// recorded with: go test ./config -run TestGoldenTables -update-golden
func TestGoldenTables(t *testing.T) {
	path := filepath.Join("testdata", "golden_tables.json")
	live, err := json.MarshalIndent(liveGoldenTables(), "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if *updateGolden {
		if err := os.WriteFile(path, append(live, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no golden record: %v (write one with -update-golden)", err)
	}
	var a, b map[string]map[string]json.RawMessage
	if err := json.Unmarshal(want, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(live, &b); err != nil {
		t.Fatal(err)
	}
	var diffs []string
	for _, table := range sortedKeys(a, b) {
		for _, key := range sortedKeys(a[table], b[table]) {
			if string(compact(a[table][key])) != string(compact(b[table][key])) {
				diffs = append(diffs, table+"/"+key+": recorded "+string(compact(a[table][key]))+", live "+string(compact(b[table][key])))
			}
		}
	}
	if len(diffs) > 0 {
		shown := diffs
		if len(shown) > 25 {
			shown = shown[:25]
		}
		for _, d := range shown {
			t.Error(d)
		}
		t.Fatalf("%d entries differ from the golden record", len(diffs))
	}
}

func compact(raw json.RawMessage) []byte {
	if raw == nil {
		return []byte("absent")
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return raw
	}
	out, _ := json.Marshal(v)
	return out
}

func sortedKeys[V any](ms ...map[string]V) []string {
	seen := map[string]bool{}
	for _, m := range ms {
		for k := range m {
			seen[k] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
