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
	MaxCount   int                `json:"max_count"`
}

type goldenEffect struct {
	Type   string  `json:"type"`
	Target string  `json:"target"`
	Value  float64 `json:"value"`
}

type goldenTech struct {
	Age   string  `json:"age"`
	Cost  float64 `json:"cost"`
	Ticks int     `json:"ticks"`
}

type goldenGate struct {
	Buildings map[string]int `json:"buildings"`
}

// goldenTables is the record of every number the economy's tables hold once
// the game has loaded them: prices, rates, outputs, storage sizes, build and
// research times, gates, and the price levels and market rates worked out
// from them.
type goldenTables struct {
	Buildings      map[string]goldenBuilding     `json:"buildings"`
	Techs          map[string]goldenTech         `json:"techs"`
	Gates          map[string]goldenGate         `json:"gates"`
	AgeTargetSecs  map[string]float64            `json:"age_target_seconds"`
	PriceLevels    map[string]map[string]float64 `json:"price_levels"`
	MarketRates    map[string]map[string]float64 `json:"market_rates"`
	TypicalIncome  map[string]map[string]float64 `json:"typical_income"`
	TypicalStorage map[string]map[string]float64 `json:"typical_storage"`
	EpochEvents    map[string]int                `json:"epoch_event_ticks"`
	Awakenings     map[string]int                `json:"awakening_ticks"`
	TradeRoutes    map[string]int                `json:"trade_route_ticks"`
	DealLevels     map[string]map[string]float64 `json:"deal_flow_levels"`
}

func liveGoldenTables() goldenTables {
	g := goldenTables{
		Buildings: map[string]goldenBuilding{}, Techs: map[string]goldenTech{}, Gates: map[string]goldenGate{},
		AgeTargetSecs: map[string]float64{}, PriceLevels: map[string]map[string]float64{}, MarketRates: map[string]map[string]float64{},
		EpochEvents: map[string]int{}, Awakenings: map[string]int{}, TradeRoutes: map[string]int{},
	}
	defs := BaseBuildings()
	for _, d := range defs {
		b := goldenBuilding{Category: d.Category, Age: d.RequiredAge, Cost: d.BaseCost, Rate: d.CostScale,
			BuildTicks: d.BuildTicks, Workers: d.WorkerCapacity, MaxCount: d.MaxCount, Effects: []goldenEffect{}}
		for _, e := range d.Effects {
			b.Effects = append(b.Effects, goldenEffect{Type: e.Type, Target: e.Target, Value: e.Value})
		}
		g.Buildings[d.Key] = b
	}
	techs := Technologies()
	for _, t := range techs {
		g.Techs[t.Key] = goldenTech{Age: t.Age, Cost: t.Cost, Ticks: t.ResearchTicks}
	}
	order := AgeOrder()
	for _, a := range Ages() {
		g.Gates[a.Key] = goldenGate{Buildings: a.BuildingReqs}
		g.AgeTargetSecs[a.Key] = AgeTargets[a.Key].Seconds()
		g.PriceLevels[a.Key] = PriceLevels(a.Key)
		rates := map[string]float64{}
		for _, p := range MarketPairs(a.Key) {
			if r, ok := MarketRate(p.From, p.To, a.Key); ok {
				rates[p.From+">"+p.To] = r
			}
		}
		g.MarketRates[a.Key] = rates
	}
	g.TypicalIncome = Incomes(defs, techs, order, AnyResource)
	g.TypicalStorage = TypicalStorages(defs, BaseResources(), order)
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
