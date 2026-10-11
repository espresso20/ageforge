package smoke

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestGateCovenant is the regression guard for the whole "requirement you
// can never store" class of soft-lock (the stash deadlock, the 50-longhouse
// Stone Age wall, the Atomic 30-bunker wall): every age advance must pass the
// Gate Covenant from config alone. It runs in plain `go test ./...`, so CI
// catches a balance change that breaks a gate before any bot has to play
// into it. The covenant's rules are listed at the top of static.go.
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

// TestStorageCovenant: the storage tables follow the storage rule, checked
// from the tables alone (see static_storage.go): sizes that hold the dearest
// price of the age and a quarter more, copies that each fit under the ones
// before, a town that arrives able to pay the next age's first storage and
// housing, gates that ask for the five copies and a quarter of each store, and
// the wall in the first four ages at exactly five copies.
func TestStorageCovenant(t *testing.T) {
	problems, rows := StaticStorage()
	if len(rows) < 20 {
		t.Fatalf("only %d ages have a storage building", len(rows))
	}
	for _, p := range problems {
		t.Errorf("%s (%s)", p, p.Rule)
	}
}

// TestStorageCovenantCatchesBrokenStorage keeps the guard honest: each way
// the tables can part from the storage rule must be flagged under the rule
// that names it, in the age it breaks, and the real tables must flag nothing.
func TestStorageCovenantCatchesBrokenStorage(t *testing.T) {
	flagged := func(defs map[string]config.BuildingDef, ages []config.AgeDef) map[string]bool {
		problems, _ := staticStorage(defs, ages)
		out := map[string]bool{}
		for _, p := range problems {
			out[p.Rule+"/"+p.Age] = true
		}
		return out
	}
	// edit returns the real tables with one storage building changed.
	edit := func(key string, change func(d *config.BuildingDef)) map[string]config.BuildingDef {
		defs := config.BuildingByKey()
		d := defs[key]
		d.Effects = append([]config.Effect(nil), d.Effects...)
		change(&d)
		defs[key] = d
		return defs
	}
	scaleSize := func(by float64) func(d *config.BuildingDef) {
		return func(d *config.BuildingDef) {
			for i, e := range d.Effects {
				if e.Type == "storage" {
					d.Effects[i].Value = e.Value * by
				}
			}
		}
	}
	// editGate returns the real ages with one gate's requirements changed.
	editGate := func(to string, change func(b map[string]int, r map[string]float64)) []config.AgeDef {
		ages := config.Ages()
		for i := range ages {
			if ages[i].Key != to {
				continue
			}
			b, r := map[string]int{}, map[string]float64{}
			for k, v := range ages[i].BuildingReqs {
				b[k] = v
			}
			for k, v := range ages[i].ResourceReqs {
				r[k] = v
			}
			change(b, r)
			ages[i].BuildingReqs, ages[i].ResourceReqs = b, r
		}
		return ages
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
		// A store a tenth too small for what the age asks: five copies must
		// hold the dearest price and a quarter more.
		{"a small Classical Vault", edit("classical_vault", scaleSize(0.1)), ages, []string{"hold/classical_age"}},
		{"a small Stash", edit("stash", scaleSize(0.1)), ages, []string{"hold/primitive_age"}},
		// A store the next age's first storage building does not fit in.
		{"a Victorian Vault priced past the Industrial store", edit("victorian_vault", func(d *config.BuildingDef) {
			cost := map[string]float64{}
			for k, v := range d.BaseCost {
				cost[k] = v * 100
			}
			d.BaseCost = cost
		}), ages, []string{"arrive/industrial_age", "climb/victorian_age"}},
		// Storage with a copy limit, or climbing at the wrong rate.
		{"a copy limit on the Warehouse", edit("warehouse", func(d *config.BuildingDef) { d.MaxCount = 25 }), ages, []string{"curve/bronze_age"}},
		{"a slow Granary", edit("granary", func(d *config.BuildingDef) { d.CostScale = 1.15 }), ages, []string{"curve/iron_age", "wall/iron_age"}},
		// A late store that climbs too slowly walls far past the reference.
		{"a slow Fusion Vault", edit("fusion_vault", func(d *config.BuildingDef) { d.CostScale = 1.2 }), ages, []string{"curve/fusion_age", "wall/fusion_age"}},
		// Gates.
		{"a gate without the storage copies", config.BuildingByKey(), editGate("iron_age", func(b map[string]int, r map[string]float64) { delete(b, "warehouse") }), []string{"gate_copies/bronze_age"}},
		{"a gate asking for three copies", config.BuildingByKey(), editGate("classical_age", func(b map[string]int, r map[string]float64) { b["granary"] = 3 }), []string{"gate_copies/iron_age"}},
		{"a gate asking twice the stone", config.BuildingByKey(), editGate("medieval_age", func(b map[string]int, r map[string]float64) { r["stone"] *= 2 }), []string{"gate_amount/classical_age"}},
		{"a gate asking for faith", config.BuildingByKey(), editGate("renaissance_age", func(b map[string]int, r map[string]float64) { r["faith"] = 44000 }), []string{"gate_amount/medieval_age"}},
		{"a gate that forgets a material", config.BuildingByKey(), editGate("medieval_age", func(b map[string]int, r map[string]float64) { delete(r, "iron") }), []string{"gate_amount/classical_age"}},
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

// TestGateCovenantCatchesBrokenGates keeps the guard honest: the pre-fix
// numbers must each be flagged. 50 longhouses was the Stone Age wall (52
// here: the Stone Age's storage has grown since, and the 50th now fits), 30
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
			ages[i].BuildingReqs = map[string]int{"longhouse": 52}
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
	sistine.BaseCost = map[string]float64{"stone": 200e6, "gold": 10e6, "faith": 6e6, "culture": 8e6}
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
// temples, nor for the Classical Age's gold (and the knowledge Mathematics,
// the Colosseum's keystone, costs, which only the gold-priced agora makes).
// The old rule counted market parity without asking whether a trade building
// could stand.
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
		"unsourced/classical_age requirement/gold": false, "unsourced/mathematics/knowledge": false}
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

// TestStaticFeatureLocks pins which feature locks are live and which wait
// for their tech. The tree holds every tech a lock names now, so all nine
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
	if got, want := strings.Join(live, "; "), "trade_routes <- the_wheel (bronze_age); campaigns <- military_tactics (bronze_age); expeditions <- exploration (iron_age); diplomacy <- envoys (classical_age); festivals <- drama (classical_age); naval_expedition <- navigation (renaissance_age); black_market <- mercantilism (colonial_age); route_rail_freight <- railroads (industrial_age); route_warp_commerce <- interstellar_trade (interstellar_age)"; got != want {
		t.Errorf("live locks:\n got %s\nwant %s", got, want)
	}
	if got := strings.Join(FeatureLocksWaiting(rows), " "); got != "" {
		t.Errorf("locks waiting for their tech: %s; want none", got)
	}
	var sb strings.Builder
	writeFeatureLocks(&sb, rows)
	if !strings.Contains(sb.String(), "| Campaigns | Military Tactics | Bronze | live |") || !strings.Contains(sb.String(), "| Festivals | Drama | Classical | live |") ||
		!strings.Contains(sb.String(), "| The Warp Commerce route | Interstellar Trade | Interstellar | live |") {
		t.Errorf("the report's table is off:\n%s", sb.String())
	}
}

// TestStaticCaps pins the caps report's reading: with techs in a layer of
// their own, a player who holds every milestone, wonder and monument still
// takes the all-production pool past its knee, from the Electric Age at the
// earliest, and no resource's own pool passes its knee in any age. Past the
// knee the pool applies a quarter of each point: the last age's +661%
// applies +315%, where the old clamp applied +200%. And the pacing model's
// all-production pool never holds more than such a player can.
func TestStaticCaps(t *testing.T) {
	rows := StaticCaps()
	if len(rows) != len(config.AgeOrder()) {
		t.Fatalf("%d rows for %d ages", len(rows), len(config.AgeOrder()))
	}
	all := func(r CapRow) float64 { return r.All() }
	if got := KneePassedIn(rows, 0, all); got != "electric_age" {
		t.Errorf("all production passes its knee on what stands alone in %q, want the Electric Age", got)
	}
	if got := KneePassedIn(rows, FestivalBonus+PlentyBonus, all); got != "industrial_age" {
		t.Errorf("all production passes its knee with a festival and an Age of Plenty in %q, want the Industrial Age", got)
	}
	last := rows[len(rows)-1]
	for res := range last.Resources {
		if age := KneePassedIn(rows, 0, func(r CapRow) float64 { return r.Resources[res] }); age != "" {
			t.Errorf("%s's own pool passes its knee in the %s", res, age)
		}
	}
	near := func(got, want float64) bool { return math.Abs(got-want) < 1e-9 }
	if !near(last.Milestones, 4.35) || !near(last.Wonders, 2.15) || !near(last.Monuments, 0.11) || !near(last.Resources["knowledge"], 1.45) || !near(last.Resources["gold"], 0.70) {
		t.Errorf("the last age holds %+v", last)
	}
	// Earned against applied, by the rule written out: all of it up to
	// +200%, a quarter of the rest.
	if !near(last.All(), 6.61) || !near(last.Applied(), 2+4.61*0.25) {
		t.Errorf("the last age has earned %+.4f and applies %+.4f, want +6.61 and +3.1525", last.All(), last.Applied())
	}
	for i, r := range rows {
		if i > 0 && r.All() < rows[i-1].All() {
			t.Errorf("%s holds less all production than the age before it", r.Age)
		}
		want := r.All()
		if want > 2 {
			want = 2 + (want-2)*0.25
		}
		if !near(r.Applied(), want) {
			t.Errorf("%s: +%.0f%% earned applies +%.2f%%, want +%.2f%%", r.Age, r.All()*100, r.Applied()*100, want*100)
		}
		if i > 0 && r.Applied() < rows[i-1].Applied() {
			t.Errorf("%s applies less all production than the age before it", r.Age)
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
}

// TestGateCovenantCatchesABuildingBehindAnOptionalTech is the content rule
// for moving a building onto a tech: a tech may hold a building only if the
// age keeps another way to make what it makes, or the tech is one no run
// leaves the age without. The Cathedral is the Medieval Age's only faith
// producer and the Renaissance asks for faith. Behind Theology, the age's
// keystone, the gate still stands. Behind Alchemy, which a run may skip, it
// has no source left, and the guard says so.
func TestGateCovenantCatchesABuildingBehindAnOptionalTech(t *testing.T) {
	defs := config.BuildingByKey()
	if got := defs["cathedral"].RequiredTech; got != "theology" {
		t.Fatalf("the Cathedral waits for %q, want theology: the test's premise", got)
	}
	if problems, _ := staticGates(config.Ages(), defs); len(problems) != 0 {
		t.Fatalf("today's tables break the covenant: %+v", problems)
	}
	cathedral := defs["cathedral"]
	cathedral.RequiredTech = "alchemy"
	defs["cathedral"] = cathedral
	problems, _ := staticGates(config.Ages(), defs)
	found := false
	for _, g := range problems {
		if g.Kind == "unsourced" && g.Resource == "faith" && g.From == "medieval_age" {
			found = true
		}
	}
	if !found {
		t.Errorf("with the Cathedral behind an optional tech the Renaissance's faith should have no source; problems: %+v", problems)
	}
}
