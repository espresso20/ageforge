package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestGateCovenant is the regression guard for the whole "requirement you
// can never store" class of soft-lock (the stash deadlock, the 50-longhouse
// Stone Age wall, the Atomic 30-bunker wall): every age advance must pass the
// Gate Covenant from config alone. It runs in plain `go test ./...`, so CI
// catches a balance change that breaks a gate before any bot has to play
// into it. See design-and-architecture/economy.md, "Gate Covenant".
func TestGateCovenant(t *testing.T) {
	problems, slack := StaticGates()
	for _, g := range problems {
		switch g.Kind {
		case "unbuildable":
			t.Errorf("%s -> %s: requires %d %s, which can only be built in %s; name the building its lineage has in %s",
				g.From, g.To, g.Count, g.Key, g.BuiltIn, g.From)
		case "unsourced":
			t.Errorf("%s -> %s: needs %s (for %s) but nothing supplies it by %s without first spending some",
				g.From, g.To, g.Resource, g.Key, g.From)
		case "dead_building":
			t.Errorf("%s (%s) costs %s, which has no source in that age", g.Key, g.From, g.Resource)
		case "building":
			t.Errorf("%s -> %s: copy #%d of %s costs %s %s, over 1/%g of the %s storage buildable in %s",
				g.From, g.To, g.Count, g.Key, num(g.Need), g.Resource, g.Margin, num(g.MaxStorage), g.From)
		case "ladder":
			t.Errorf("%s -> %s: the first %s costs %s %s, over 1/%g of the %s storage the gate forces (%s); a player who met the gate with nothing to spare could never raise a cap in %s",
				g.From, g.To, g.Key, num(g.Need), g.Resource, g.Margin, num(g.MaxStorage), g.ForcedBy, g.To)
		default:
			t.Errorf("%s -> %s: needs %s %s, more than 1/%g of the %s storage buildable in %s",
				g.From, g.To, num(g.Need), g.Resource, g.Margin, num(g.MaxStorage), g.From)
		}
	}
	if want := len(config.Ages()) - 1; len(slack) != want {
		t.Errorf("slack rows = %d, want one per advance (%d)", len(slack), want)
	}
}

// TestStorageCovenant: the most storage buildable in every age holds
// config.StorageHoldHours of the age's typical production of each of its
// construction resources (economy.md, Law 1). A store that fills in minutes
// throws away most of what a player makes between visits.
func TestStorageCovenant(t *testing.T) {
	rows := StaticStorage()
	if len(rows) == 0 {
		t.Fatal("no ages checked")
	}
	for _, r := range rows {
		if !r.OK() {
			t.Errorf("%s: the most %s storage buildable (%s) holds %.2f h of typical income (%s/tick), under %g h; raise the age's storage per copy",
				r.Age, r.Resource, num(r.MaxStorage), r.Hours, num(r.Income), config.StorageHoldHours)
		}
	}
}

// TestStorageCovenantCatchesBrokenStorage keeps the guard honest: the
// storage the Renaissance to Victorian Ages had before the covenant (500K,
// 10M, 50M and 350M per copy, where a full store held under half an hour of
// gold or steel) must each be flagged, and so must halving every storage
// building, which puts nearly every age under.
func TestStorageCovenantCatchesBrokenStorage(t *testing.T) {
	old := map[string]float64{
		"renaissance_vault":  500e3,
		"colonial_warehouse": 10e6,
		"industrial_depot":   50e6,
		"victorian_vault":    350e6,
	}
	withStorage := func(per func(key string, v float64) float64) map[string]config.BuildingDef {
		defs := config.BuildingByKey()
		for k, d := range defs {
			if d.Category != "storage" {
				continue
			}
			effs := append([]config.Effect(nil), d.Effects...)
			for i, e := range effs {
				if e.Type == "storage" {
					effs[i].Value = per(k, e.Value)
				}
			}
			d.Effects = effs
			defs[k] = d
		}
		return defs
	}
	broken := withStorage(func(k string, v float64) float64 {
		if o, ok := old[k]; ok {
			return o
		}
		return v
	})
	flagged := map[string]bool{}
	for _, r := range staticStorage(broken, config.TypicalIncome) {
		if !r.OK() {
			flagged[r.Age] = true
		}
	}
	for _, age := range []string{"renaissance_age", "colonial_age", "industrial_age", "victorian_age"} {
		if !flagged[age] {
			t.Errorf("%s: the pre-covenant storage was not flagged", age)
		}
	}
	halved := 0
	for _, r := range staticStorage(withStorage(func(_ string, v float64) float64 { return v / 2 }), config.TypicalIncome) {
		if !r.OK() {
			halved++
		}
	}
	if halved < 10 {
		t.Errorf("halving every storage building flagged only %d ages", halved)
	}
}

// TestGateCovenantCatchesBrokenGates keeps the guard honest: the pre-fix
// numbers must each be flagged. 50 longhouses was the Stone Age wall, 30
// barracks for Medieval named a Bronze Age building the age lock forbids
// building later (and, with the gate left asking only 220K stone, nothing
// made a player hold enough storage for a 340K stone Strongroom, the only
// storage the Medieval Age builds: the storage ladder), 80K food cannot fit
// 1.25x under Stone Age storage, and
// iron does not exist before the Bronze Age; and a Bronze Age smithy priced
// in coal (which unlocks in the Renaissance) could never be built. The 80K
// food is also far more than a Stone Age economy makes in 45 minutes, with no
// market yet to buy the rest.
//
// The wonder and flow checks, with the numbers that slipped past the older
// covenant: the Renaissance's 44K faith (about 1.5 faith/tick in the Medieval
// Age) and the Sistine Chapel's 6M faith, which no market sells; the Stellar
// Cradle's 940T uranium, which only Atomic Age mines (unbuildable by the
// Fusion Age) produced; and a Sistine Chapel whose stone outgrows every
// Renaissance warehouse.
func TestGateCovenantCatchesBrokenGates(t *testing.T) {
	ages := config.Ages()
	for i := range ages {
		switch ages[i].Key {
		case "bronze_age":
			ages[i].BuildingReqs = map[string]int{"longhouse": 50}
			ages[i].ResourceReqs = map[string]float64{"food": 80000, "iron": 10}
		case "medieval_age":
			ages[i].BuildingReqs = map[string]int{"barracks": 30}
		case "renaissance_age":
			reqs := map[string]float64{}
			for k, v := range ages[i].ResourceReqs {
				reqs[k] = v
			}
			reqs["faith"] = 44000
			ages[i].ResourceReqs = reqs
		}
	}
	defs := config.BuildingByKey()
	smithy := defs["smithy"]
	smithy.BaseCost = map[string]float64{"wood": 900, "coal": 100}
	defs["smithy"] = smithy
	sistine := defs["sistine_chapel"]
	sistine.BaseCost = map[string]float64{"stone": 100e6, "gold": 10e6, "faith": 6e6, "culture": 8e6}
	defs["sistine_chapel"] = sistine
	cradle := defs["stellar_cradle"]
	cradle.BaseCost = map[string]float64{"uranium": 940e12}
	for k, v := range defs["stellar_cradle"].BaseCost {
		cradle.BaseCost[k] = v
	}
	defs["stellar_cradle"] = cradle
	problems, _ := staticGates(ages, defs)
	want := map[string]bool{"building/longhouse": false, "resource/food": false, "unbuildable/barracks": false, "ladder/keep": false,
		"unsourced/bronze_age requirement": false, "dead_building/smithy": false,
		"flow/bronze_age requirement": false, "flow/renaissance_age requirement": false, "flow/sistine_chapel": false,
		"wonder/sistine_chapel": false, "unsourced/stellar_cradle": false}
	for _, g := range problems {
		k := g.Kind + "/" + g.Key
		if _, ok := want[k]; ok {
			want[k] = true
			continue
		}
		t.Errorf("unexpected problem %+v", g)
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("guard missed %s", k)
		}
	}
}

// TestGateCovenantCatchesColdStartTraps feeds the covenant the Iron Age
// trading post as it was, priced in gold. It was the Iron Age's only gold
// producer and its only trade building (the market needs one), so a player
// who skipped the optional Bronze Age market could never get gold in the
// Iron Age: not for the trading post, nor for the agoras, legion forts and
// temples, nor for the Classical Age's gold (and its knowledge, which only
// the gold-priced agora makes). The old rule counted market parity without
// asking whether a trade building could stand.
//
// The second case is the carry-over assumption at work: had the Iron Age
// gate required a market, that market would stand in the Iron Age, the
// exchange would be open, and the same price would be fine.
func TestGateCovenantCatchesColdStartTraps(t *testing.T) {
	oldPost := func() map[string]config.BuildingDef {
		defs := config.BuildingByKey()
		post := defs["trading_post"]
		post.BaseCost = map[string]float64{"stone": 32000, "iron": 15000, "gold": 8800}
		defs["trading_post"] = post
		return defs
	}
	problems, _ := staticGates(config.Ages(), oldPost())
	want := map[string]bool{"dead_building/trading_post": false, "dead_building/agora": false,
		"dead_building/legion_fort": false, "dead_building/temple": false,
		"unsourced/classical_age requirement/gold": false, "unsourced/classical_age requirement/knowledge": false}
	for _, g := range problems {
		k := g.Kind + "/" + g.Key
		if g.Kind == "unsourced" {
			k += "/" + g.Resource
		}
		if _, ok := want[k]; ok {
			want[k] = true
			continue
		}
		t.Errorf("unexpected problem %+v", g)
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("guard missed %s", k)
		}
	}

	ages := config.Ages()
	for i := range ages {
		if ages[i].Key == "iron_age" {
			reqs := map[string]int{"market": 1}
			for k, v := range ages[i].BuildingReqs {
				reqs[k] = v
			}
			ages[i].BuildingReqs = reqs
		}
	}
	if problems, _ := staticGates(ages, oldPost()); len(problems) > 0 {
		t.Errorf("with a required Bronze Age market carried into the Iron Age, want no problems, got %+v", problems)
	}
}

// TestGateCovenantCatchesBrokenLadder: the storage ladder. Entering the
// Victorian Age, the only storage a player can build is the Victorian Vault
// (about 210M steel for the first copy), so the gate must make them hold
// more than that first. Today the 30th tenement (381M stone) does. A retune
// to 10 tenements would leave the fifth steel mill (210M steel) as the
// biggest price the gate forces, and a vault priced at twice today's would
// outgrow even the 381M; either way a player who met the gate with nothing
// to spare could never raise a cap in the Victorian Age.
func TestGateCovenantCatchesBrokenLadder(t *testing.T) {
	ladderRows := func(problems []GateProblem) []GateProblem {
		var out []GateProblem
		for _, g := range problems {
			if g.Kind != "ladder" || g.Key != "victorian_vault" {
				t.Errorf("unexpected problem %+v", g)
				continue
			}
			out = append(out, g)
		}
		return out
	}

	ages := config.Ages()
	for i := range ages {
		if ages[i].Key == "victorian_age" {
			reqs := map[string]int{}
			for k, v := range ages[i].BuildingReqs {
				reqs[k] = v
			}
			reqs["tenement"] = 10
			ages[i].BuildingReqs = reqs
		}
	}
	problems, _ := staticGates(ages, config.BuildingByKey())
	if rows := ladderRows(problems); len(rows) != 1 || rows[0].From != "industrial_age" || rows[0].Resource != "steel" {
		t.Errorf("10 tenements: want one ladder row for the vault's steel, got %+v", rows)
	}

	defs := config.BuildingByKey()
	vault := defs["victorian_vault"]
	vault.BaseCost = map[string]float64{}
	for k, v := range defs["victorian_vault"].BaseCost {
		vault.BaseCost[k] = 2 * v
	}
	defs["victorian_vault"] = vault
	problems, _ = staticGates(config.Ages(), defs)
	if rows := ladderRows(problems); len(rows) != 1 || !strings.Contains(rows[0].ForcedBy, "tenement #30") {
		t.Errorf("a doubled vault: want one ladder row against the 30th tenement, got %+v", rows)
	}
}
