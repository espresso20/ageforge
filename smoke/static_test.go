package smoke

import (
	"maps"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// TestGateCovenant is the regression guard for the whole "requirement you
// can never meet" class of soft-lock (a building only an older age can build,
// a price in a resource nothing makes yet): every age advance must pass the
// Gate Covenant from config alone. It runs in plain `go test ./...`, so CI
// catches a balance change that breaks a gate before any bot has to play
// into it. The covenant's rules are listed at the top of static.go.
func TestGateCovenant(t *testing.T) {
	for _, g := range StaticGates() {
		switch g.Kind {
		case "unbuildable":
			t.Errorf("%s -> %s: requires %d %s, which can only be built in %s; name the building its lineage has in %s",
				g.From, g.To, g.Count, g.Key, g.BuiltIn, g.From)
		case "unsourced":
			t.Errorf("%s -> %s: needs %s (for %s) but nothing supplies it by %s without first spending some",
				g.From, g.To, g.Resource, g.Key, g.From)
		case "dead_building":
			t.Errorf("%s (%s) costs %s, which has no source in that age", g.Key, g.From, g.Resource)
		default:
			t.Errorf("%s -> %s: %s needs %s %s (%s)", g.From, g.To, g.Key, num(g.Need), g.Resource, g.Kind)
		}
	}
}

// TestStorageCheck: the storage tables hold what is left of the storage
// rule, checked from the tables alone (see static_storage.go): the price
// curve with no copy limit, a building that only boosts production, gates that
// ask for the five copies, and the five copies payable from the age's income.
func TestStorageCheck(t *testing.T) {
	problems, rows := StaticStorage()
	if len(rows) < 20 {
		t.Fatalf("only %d ages have a storage building", len(rows))
	}
	for _, p := range problems {
		t.Errorf("%s (%s)", p, p.Rule)
	}
	for _, r := range rows {
		if r.Share <= 0 || r.Share > 1 {
			t.Errorf("%s: the fifth copy costs %.3fx of the age's income over one target, want a share in (0, 1]", r.Age, r.Share)
		}
	}
	var sb strings.Builder
	writeStorage(&sb, problems, rows)
	if out := sb.String(); !strings.Contains(out, "No problems.") || strings.Count(out, "\n| ") != len(rows)+1 {
		t.Errorf("the report section:\n%s", out)
	}
}

// editedGate changes the building requirements of the gate into `to` in
// ages by change, which is handed a copy (the real map is shared), and
// returns ages.
func editedGate(ages []config.AgeDef, to string, change func(b map[string]int)) []config.AgeDef {
	for i := range ages {
		if ages[i].Key != to {
			continue
		}
		b := maps.Clone(ages[i].BuildingReqs)
		change(b)
		ages[i].BuildingReqs = b
	}
	return ages
}

// withDef changes the building key in defs by change, which is handed a copy
// whose BaseCost and Effects it may edit freely (the real ones are shared),
// and returns defs.
func withDef(defs map[string]config.BuildingDef, key string, change func(d *config.BuildingDef)) map[string]config.BuildingDef {
	d := defs[key]
	d.BaseCost = maps.Clone(d.BaseCost)
	d.Effects = slices.Clone(d.Effects)
	change(&d)
	defs[key] = d
	return defs
}

// expectProblems holds problems to exactly the wanted ones: key names each by
// kind and subject, a problem nobody wanted is unexpected, and a wanted one
// that never came is a guard that missed.
func expectProblems(t *testing.T, problems []GateProblem, key func(GateProblem) string, want ...string) {
	t.Helper()
	seen := map[string]bool{}
	for _, g := range problems {
		k := key(g)
		if !slices.Contains(want, k) {
			t.Errorf("unexpected problem %+v", g)
			continue
		}
		seen[k] = true
	}
	for _, k := range want {
		if !seen[k] {
			t.Errorf("guard missed %s", k)
		}
	}
}

// TestStorageCheckCatchesBrokenStorage keeps the guard honest: each way the
// tables can part from the storage check must be flagged under the rule that
// names it, in the age it breaks, and the real tables must flag nothing.
func TestStorageCheckCatchesBrokenStorage(t *testing.T) {
	income := config.Incomes(config.BaseBuildings(), config.Technologies(), config.AgeOrder(), config.AnyResource)
	flagged := func(defs map[string]config.BuildingDef, ages []config.AgeDef) map[string]bool {
		problems, _ := staticStorage(defs, ages, income)
		out := map[string]bool{}
		for _, p := range problems {
			out[p.Rule+"/"+p.Age] = true
		}
		return out
	}
	// edit returns the real tables with one storage building changed.
	edit := func(key string, change func(d *config.BuildingDef)) map[string]config.BuildingDef {
		return withDef(config.BuildingByKey(), key, change)
	}
	priceTimes := func(k float64) func(d *config.BuildingDef) {
		return func(d *config.BuildingDef) {
			cost := map[string]float64{}
			for r, v := range d.BaseCost {
				cost[r] = v * k
			}
			d.BaseCost = cost
		}
	}
	defs, ages := config.BuildingByKey(), config.Ages()
	if got := flagged(defs, ages); len(got) > 0 {
		t.Fatalf("the real tables are flagged: %v", got)
	}
	cases := []struct {
		name  string
		defs  map[string]config.BuildingDef
		ages  []config.AgeDef
		wants []string
	}{
		// A first storage copy nobody can pay: its fifth copy costs more than
		// the age's income makes in a target.
		{"a Victorian Vault priced at a hundred times", edit("victorian_vault", priceTimes(100)), ages, []string{"afford/victorian_age"}},
		{"a dear Stash", edit("stash", priceTimes(1000)), ages, []string{"afford/primitive_age"}},
		// Storage with a copy limit, or climbing at the wrong rate.
		{"a copy limit on the Warehouse", edit("warehouse", func(d *config.BuildingDef) { d.MaxCount = 25 }), ages, []string{"curve/bronze_age"}},
		{"a slow Granary", edit("granary", func(d *config.BuildingDef) { d.CostScale = 1.15 }), ages, []string{"curve/iron_age"}},
		{"a slow Fusion Vault", edit("fusion_vault", func(d *config.BuildingDef) { d.CostScale = 1.2 }), ages, []string{"curve/fusion_age"}},
		// A storage building that still stores, or boosts the wrong amount.
		{"a Keep that stores", edit("keep", func(d *config.BuildingDef) {
			d.Effects = append(d.Effects, config.Effect{Type: "storage", Target: "all", Value: config.Million})
		}), ages, []string{"boost/medieval_age"}},
		{"a Stash that boosts five percent", edit("stash", func(d *config.BuildingDef) {
			d.Effects = []config.Effect{{Type: "bonus", Target: "production_all", Value: 0.05}}
		}), ages, []string{"boost/primitive_age"}},
		{"a Storage Pit that boosts nothing", edit("storage_pit", func(d *config.BuildingDef) { d.Effects = nil }), ages, []string{"boost/stone_age"}},
		// Gates.
		{"a gate without the storage copies", config.BuildingByKey(), editedGate(config.Ages(), "iron_age", func(b map[string]int) { delete(b, "warehouse") }), []string{"gate_copies/bronze_age"}},
		{"a gate asking for three copies", config.BuildingByKey(), editedGate(config.Ages(), "classical_age", func(b map[string]int) { b["granary"] = 3 }), []string{"gate_copies/iron_age"}},
	}
	for _, c := range cases {
		got := flagged(c.defs, c.ages)
		for _, w := range c.wants {
			if !got[w] {
				t.Errorf("%s: want %s flagged, got %v", c.name, w, got)
			}
		}
	}
}

// TestGateCovenantCatchesBrokenGates keeps the guard honest: numbers that
// break the covenant must each be flagged, added to the real gates. 30
// barracks for Medieval named a Bronze Age building the age lock forbids
// building later; and a Bronze Age smithy priced in coal (which unlocks in
// the Renaissance) could never be built. The smithy is also the Bronze Age's
// only maker of iron, and a town brings no iron into the age (the gates ask
// for no amount of anything), so the chain breaks with it: every Bronze Age
// building priced in iron is dead, and so is everything priced in the gold
// the market makes. The gate into the Iron Age asks for the smithy, so it is
// an unsourced row there too. (The knowledge the scriptorium makes is not
// part of the chain: the Stone Age's elders hall, which an earlier gate
// asked for, still makes it.)
//
// The flow and sourcing checks, with numbers a source check passes: the
// Sistine Chapel's 6M faith, more than the age's income makes in its
// target; the Stellar Cradle's 940T uranium, which only the Atomic Age's
// mines make and a Fusion Age town carries as 8 works, 0.9M a tick against
// the 94,000 ticks of the age's target. (No gate asks for an amount of a
// resource, so none of the cases is a gate's resource amount.)
//
// A price over what any store held is no problem now that stock has no cap
// (52 longhouses, once the Stone Age wall, are simply built), and the covenant
// reports no "building", "wonder" or "ladder" row at all: the case below
// would show one as unexpected.
func TestGateCovenantCatchesBrokenGates(t *testing.T) {
	// 52 longhouses: nothing to flag without a store to outgrow.
	ages := editedGate(config.Ages(), "bronze_age", func(b map[string]int) { b["longhouse"] = 52 })
	ages = editedGate(ages, "medieval_age", func(b map[string]int) { b["barracks"] = 30 })
	defs := withDef(config.BuildingByKey(), "smithy", func(d *config.BuildingDef) {
		d.BaseCost = map[string]float64{"wood": 900, "coal": 100}
	})
	withDef(defs, "sistine_chapel", func(d *config.BuildingDef) {
		d.BaseCost = map[string]float64{"stone": 200 * config.Million, "gold": 10 * config.Million, "faith": 6 * config.Million, "culture": 8 * config.Million}
	})
	withDef(defs, "stellar_cradle", func(d *config.BuildingDef) { d.BaseCost["uranium"] = 940 * config.Trillion })
	expectProblems(t, staticGates(ages, defs), func(g GateProblem) string { return g.Kind + "/" + g.Key },
		"unbuildable/barracks", "dead_building/smithy", "flow/sistine_chapel", "unsourced/stellar_cradle",
		// the chain from the smithy's coal: no iron, so no market (gold)
		"unsourced/smithy", "unsourced/lumber_mill", "unsourced/altar",
		"dead_building/altar", "dead_building/barracks",
		"dead_building/lumber_mill", "dead_building/market", "dead_building/quarry",
		"dead_building/scriptorium", "dead_building/warehouse")
}

// TestGateCovenantCatchesColdStartTraps feeds the covenant the Iron Age
// trading post as it was, priced in gold. The post is the Iron Age's only
// gold producer, so a town that arrived without a gold producer could never
// get gold in the Iron Age: not for the trading post, nor for the agoras,
// legion forts and temples.
//
// A town arrives with empty stores and the buildings every gate so far asked
// for, so whether that trap is open depends on the gate. The gate out of the
// Bronze Age asks for the Bronze Age market (the gate rule: it makes gold), so
// the town brings gold with it and the same price is fine: that is the real
// tables, the last case. The trap is a gate that lets a town through without
// the market building, the first.
func TestGateCovenantCatchesColdStartTraps(t *testing.T) {
	oldPost := func() map[string]config.BuildingDef {
		return withDef(config.BuildingByKey(), "trading_post", func(d *config.BuildingDef) { d.BaseCost["gold"] = 8800 })
	}
	if problems := staticGates(config.Ages(), config.BuildingByKey()); len(problems) > 0 {
		t.Fatalf("setup: the real tables, and no trap, should break nothing: %+v", problems)
	}
	withoutMarketBuilding := editedGate(config.Ages(), "iron_age", func(b map[string]int) { delete(b, "market") })
	expectProblems(t, staticGates(withoutMarketBuilding, oldPost()), func(g GateProblem) string {
		if g.Kind == "unsourced" {
			return g.Kind + "/" + g.Key + "/" + g.Resource
		}
		return g.Kind + "/" + g.Key
	}, "dead_building/trading_post", "dead_building/agora", "dead_building/legion_fort", "dead_building/temple", "unsourced/agora/gold")

	// The real gate asks for the market, so the town arrives making gold.
	if problems := staticGates(config.Ages(), oldPost()); len(problems) > 0 {
		t.Errorf("with the Bronze Age market asked for by the gate, want no problems, got %+v", problems)
	}
}

// TestGateCovenantCatchesAnUnpayableFirstCopy: the first copy of a building
// of the age a town arrives in is held to the bound the covenant puts on a
// flow resource, with the same constant. A Classical Age building no gate asks
// for (the Cultural Obelisk), priced in more culture than the reference income
// makes in the age's pacing target, is flagged; one priced at a hundredth of
// what it makes is not.
// The first tech is held to the same: a town that arrives with nothing that
// makes knowledge cannot pay the age's cheapest tech, and is flagged as
// unsourced.
func TestGateCovenantCatchesAnUnpayableFirstCopy(t *testing.T) {
	priced := func(culture float64) map[string]config.BuildingDef {
		return withDef(config.BuildingByKey(), "cultural_obelisk", func(d *config.BuildingDef) {
			d.BaseCost = map[string]float64{"stone": 17000, "culture": culture}
		})
	}
	made := float64(rules.Core().FlowIncome("culture", "classical_age") * config.AgeTargetTicks("classical_age"))
	if made <= 0 {
		t.Fatalf("setup: the Classical Age makes no culture (%g)", made)
	}
	problems := staticGates(config.Ages(), priced(made/100))
	if len(problems) > 0 {
		t.Errorf("a first copy at a hundredth of what the age makes: want no problems, got %+v", problems)
	}
	problems = staticGates(config.Ages(), priced(made*100))
	found := false
	for _, g := range problems {
		if g.Kind == "flow" && g.Key == "cultural_obelisk" && g.Resource == "culture" && g.To == "classical_age" {
			found = true
		}
	}
	if !found {
		t.Errorf("a first copy at a hundred times what the age makes: want a flow row for the obelisk, got %+v", problems)
	}
}

// TestStaticFeatureLocks pins which feature locks are live and which wait
// for their tech. The tree holds every tech a lock names now, so all
// are live and none waits; a lock added for a tech still to come would show
// in the second list, here and in the report.
func TestStaticFeatureLocks(t *testing.T) {
	rows := StaticFeatureLocks()
	var live []string
	for _, r := range rows {
		if r.Live {
			live = append(live, r.Key+" <- "+r.Tech+" ("+r.TechAge+")")
		}
	}
	if got, want := strings.Join(live, "; "), "trade_routes <- the_wheel (bronze_age); diplomacy <- envoys (classical_age); festivals <- drama (classical_age); black_market <- mercantilism (colonial_age); route_rail_freight <- railroads (industrial_age); route_warp_commerce <- interstellar_trade (interstellar_age)"; got != want {
		t.Errorf("live locks:\n got %s\nwant %s", got, want)
	}
	if got := strings.Join(FeatureLocksWaiting(rows), " "); got != "" {
		t.Errorf("locks waiting for their tech: %s; want none", got)
	}
	var sb strings.Builder
	writeFeatureLocks(&sb, rows)
	if !strings.Contains(sb.String(), "| Trade routes | The Wheel | Bronze | live |") || !strings.Contains(sb.String(), "| Festivals | Drama | Classical | live |") ||
		!strings.Contains(sb.String(), "| The Warp Commerce route | Interstellar Trade | Interstellar | live |") {
		t.Errorf("the report's table is off:\n%s", sb.String())
	}
}

// TestStaticCaps pins the caps report's reading: with techs in a layer of
// their own, a player who holds every milestone, wonder and monument, and the
// five copies of every storage building the gates ask for, holds +766% all
// production by the last age, and the pacing model's all-production pool
// never holds more than such a player can. A pool applies in full, so the
// report has nothing to say about what is applied.
func TestStaticCaps(t *testing.T) {
	rows := StaticCaps()
	if len(rows) != len(config.AgeOrder()) {
		t.Fatalf("%d rows for %d ages", len(rows), len(config.AgeOrder()))
	}
	last := rows[len(rows)-1]
	near := func(got, want float64) bool { return math.Abs(got-want) < 1e-9 }
	if !near(last.Milestones, 4.35) || !near(last.Wonders, 2.15) || !near(last.Monuments, 0.11) || !near(last.Storage, 1.05) || !near(last.Resources["knowledge"], 1.45) || !near(last.Resources["gold"], 0.70) {
		t.Errorf("the last age holds %+v", last)
	}
	if !near(last.All(), 7.66) {
		t.Errorf("the last age has earned %+.4f all production, want +7.66", last.All())
	}
	for i, r := range rows {
		if i > 0 && r.All() < rows[i-1].All() {
			t.Errorf("%s holds less all production than the age before it", r.Age)
		}
		if held := config.ProductionAllHeld[r.Age]; held > r.All()+1e-9 {
			t.Errorf("%s: the pacing model holds +%.0f%% all production, more than the +%.0f%% a game can hold by then", r.Age, held*100, r.All()*100)
		}
	}
	for age := range config.ProductionAllHeld {
		if _, ok := config.AgeByKey()[age]; !ok {
			t.Errorf("config.ProductionAllHeld lists %s, which is not an age", age)
		}
	}
	var sb strings.Builder
	writeCaps(&sb, rows)
	if strings.Contains(sb.String(), "knee") || strings.Contains(sb.String(), "soft cap") {
		t.Errorf("the report still talks of a knee:\n%s", sb.String())
	}
}

// TestGateCovenantCatchesABuildingBehindAnOptionalTech is the content rule
// for moving a building onto a tech: a tech may hold a building only if the
// age keeps another way to make what it makes, or the tech is one no run
// leaves the age without. The Market is the Bronze Age's only gold producer
// and the Iron Age asks for gold (faith left the gates, which is why this is
// not the Cathedral any more). Behind Language, the first age's spine, the
// gate still stands. Behind Alchemy, which a run may skip, it has no source
// left, and the guard says so.
func TestGateCovenantCatchesABuildingBehindAnOptionalTech(t *testing.T) {
	defs := config.BuildingByKey()
	if got := defs["market"].RequiredTech; got != "" {
		t.Fatalf("the Market waits for %q, want no tech: the test's premise", got)
	}
	if problems := staticGates(config.Ages(), defs); len(problems) != 0 {
		t.Fatalf("today's tables break the covenant: %+v", problems)
	}
	// Behind a tech every run researches (the Bronze Age's keystone chain
	// starts with Language), the Market still counts as the age's gold.
	withDef(defs, "market", func(d *config.BuildingDef) { d.RequiredTech = "language" })
	if problems := staticGates(config.Ages(), defs); len(problems) != 0 {
		t.Errorf("with the Market behind a spine tech nothing should break: %+v", problems)
	}
	// Behind one a run may never research, it does not: the Market is the
	// Bronze Age's only gold, and the Iron Age asks for it.
	withDef(defs, "market", func(d *config.BuildingDef) { d.RequiredTech = "alchemy" })
	problems := staticGates(config.Ages(), defs)
	found := false
	for _, g := range problems {
		if g.Kind == "unsourced" && g.Resource == "gold" && g.From == "bronze_age" {
			found = true
		}
	}
	if !found {
		t.Errorf("with the Market behind an optional tech the Iron Age's gold should have no source; problems: %+v", problems)
	}
}

// TestEveryAgeStandsWithoutTheMarket: the market's exchange is not a source of
// anything. Every resource an age's gate and buildings are priced in has a way
// into a town that does not pass through it, age by age: a maker the town can
// pay for in the age, a maker an earlier gate required, or hand gathering. The
// covenant reports no problem for any age, and its report says nothing of
// buying.
func TestEveryAgeStandsWithoutTheMarket(t *testing.T) {
	problems := StaticGates()
	byAge := map[string][]GateProblem{}
	for _, g := range problems {
		byAge[g.From] = append(byAge[g.From], g)
	}
	for _, a := range config.Ages() {
		t.Run(a.Key, func(t *testing.T) {
			for _, g := range byAge[a.Key] {
				t.Errorf("%s has no way to make %s without the market: %+v", a.Key, g.Resource, g)
			}
		})
	}
	var sb strings.Builder
	writeGates(&sb, []GateProblem{{From: "a", To: "b", Kind: "flow", Key: "k", Resource: "culture", Need: 2, Made: 1}})
	for _, word := range []string{"market", "bought", "price units"} {
		if strings.Contains(strings.ToLower(sb.String()), word) {
			t.Errorf("the gate report still talks of %q:\n%s", word, sb.String())
		}
	}
}

// TestColdStartIgnoresThinCarriedSources: a source the gates carried into an
// age counts toward the first copy of a building of that age only if the
// carried required copies, at their listed output, make the copy's price in
// one target of the age. The Atomic Age's Uranium Processing Works was priced
// in 1.2 billion uranium, the only uranium its age makes; the Victorian and
// Electric mines every town carries make about 983 a tick, 55 million in the
// age's target. That is a trickle against the price, so with the old price the
// age had no way to its first uranium and everything priced in it was dead.
// (The works lost the uranium part, the rest scaled up so the price in units
// did not move: the real tables.)
func TestColdStartIgnoresThinCarriedSources(t *testing.T) {
	ages := config.Ages()
	atomic := -1
	for i, a := range ages {
		if a.Key == "atomic_age" {
			atomic = i
		}
	}
	oldWorks := func() map[string]config.BuildingDef {
		return withDef(config.BuildingByKey(), "uranium_processing_works", func(d *config.BuildingDef) {
			d.BaseCost = map[string]float64{"steel": 15 * config.Billion, "electricity": 6 * config.Billion, "uranium": 1.2 * config.Billion}
		})
	}
	cs := coldStarts(ages, oldWorks())[atomic]
	made := cs.made["uranium"] * config.AgeTargetTicks("atomic_age")
	if made <= 0 || made >= 1.2*config.Billion {
		t.Fatalf("setup: the carried mines make %.4g uranium in the Atomic Age's target, want something under the works' 1.2B", made)
	}
	if !cs.affords("uranium", made/2) {
		t.Errorf("a price of half what the carried copies make in a target: want it payable")
	}
	if cs.affords("uranium", made*2) {
		t.Errorf("a price of twice what the carried copies make in a target: want it a trickle against the price")
	}
	dead := func(defs map[string]config.BuildingDef) int {
		problems := staticGates(ages, defs)
		n := 0
		for _, g := range problems {
			if g.From == "atomic_age" && g.Resource == "uranium" {
				n++
			}
		}
		return n
	}
	if n := dead(oldWorks()); n == 0 {
		t.Errorf("the works priced in uranium: want the Atomic Age's uranium buildings flagged, got none")
	}
	if n := dead(config.BuildingByKey()); n != 0 {
		t.Errorf("the real tables: want no uranium problem in the Atomic Age, got %d", n)
	}
}

// TestColdStartOpenIsTheOneException: the real tables have exactly one
// cold-start problem, the Transcendent Age's Singularity Core, and ColdStartOpen
// names it. If a maker ever pays for it the row goes and this test says to
// delete the exception; if anything else breaks, it is reported with the
// covenant's other problems and not hidden here.
func TestColdStartOpenIsTheOneException(t *testing.T) {
	all := staticGatesAll(config.Ages(), config.BuildingByKey())
	if len(all) != len(ColdStartOpen) {
		t.Fatalf("the real tables have %d cold-start problems, ColdStartOpen names %d: %+v", len(all), len(ColdStartOpen), all)
	}
	for i, o := range ColdStartOpen {
		g := all[i]
		if g.From != o.From || g.Kind != o.Kind || g.Key != o.Key || g.Resource != o.Resource {
			t.Errorf("ColdStartOpen[%d] = %+v, the covenant finds %+v", i, o, g)
		}
	}
	open := false
	for _, h := range config.MakerOpenHoles() {
		open = open || (h.Age == "transcendent_age" && h.Material == "quantum_flux")
	}
	if !open {
		t.Errorf("the Transcendent Age's quantum flux is a cold-start exception and not an open hole in config.MakerOpenHoles")
	}
}
