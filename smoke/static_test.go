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

// TestStorageCovenant: the most storage buildable in every age holds
// config.StorageHold(age) hours of the age's typical production of each of
// its construction resources (the economy design's Law 1): 4.5 from the Bronze Age
// on, 1.5 in the Primitive and Stone Ages. A store that fills in minutes
// throws away most of what a player makes between visits.
func TestStorageCovenant(t *testing.T) {
	rows := StaticStorage()
	if len(rows) == 0 {
		t.Fatal("no ages checked")
	}
	for _, r := range rows {
		if !r.OK() {
			t.Errorf("%s: the most %s storage buildable (%s) holds %.2f h of typical income (%s/tick), under %g h; raise the age's storage per copy",
				r.Age, r.Resource, num(r.MaxStorage), r.Hours, num(r.Income), r.Want())
		}
	}
}

// TestStorageCovenantCatchesBrokenStorage keeps the guard honest: the
// storage the Renaissance to Victorian Ages had before the covenant (500K,
// 10M, 50M and 350M per copy, where a full store held under half an hour of
// gold or steel) must each be flagged, and so must halving every storage
// building, which puts nearly every age under. The covenant counts hours of
// typical income, and on the one-week curve a tick from the Bronze Age on
// makes 1/config.PacingStretch as much, so every store holds that much
// longer: the broken numbers are divided by the stretch of the building's age
// to stay as short as they were.
func TestStorageCovenantCatchesBrokenStorage(t *testing.T) {
	defsByKey := config.BuildingByKey()
	stretch := func(k string) float64 { return config.AgeStretch(defsByKey[k].RequiredAge) }
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
			return o / stretch(k)
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
	for _, r := range staticStorage(withStorage(func(k string, v float64) float64 { return v / 2 / stretch(k) }), config.TypicalIncome) {
		if !r.OK() {
			halved++
		}
	}
	if halved < 10 {
		t.Errorf("halving every storage building flagged only %d ages", halved)
	}

	// The 4.5-hour threshold (Pacing v2's away-proofing). Each storage
	// building it raised, put back alone on its old storage per copy, must
	// leave its own age short: the raises were the least that pass, at two
	// significant figures. Under the old 1.5-hour covenant none of them was.
	// The Victorian Vault (1.1B) and the Electric Warehouse (3.5B) left the
	// list when their ages' producers were set to repay more slowly
	// (config.PaybackAdjust): typical income fell there, and the old sizes
	// would hold 4.5 hours again. Their storage was left where it is.
	before := map[string]float64{
		"warehouse": 11e3, "classical_vault": 110e3, "keep": 410e3,
		"colonial_warehouse": 33e6, "industrial_depot": 170e6,
		"info_vault": 790e9, "cyber_vault": 8e12,
	}
	for _, k := range sortedKeys(before) {
		reverted := withStorage(func(key string, v float64) float64 {
			if key == k {
				return before[k]
			}
			return v
		})
		age := defsByKey[k].RequiredAge
		for _, r := range staticStorage(reverted, config.TypicalIncome) {
			if r.Age == age && r.OK() {
				t.Errorf("%s back at %s per copy: %s still holds %.2f h of typical income; the 4.5-hour covenant should flag it", k, num(before[k]), age, r.Hours)
			}
		}
	}
}

// TestStorageCovenantHours: the first hour of the game keeps its pace. The
// Primitive and Stone Ages are graded at config.EarlyStorageHoldHours (the
// Stone Age holds about 1.54 hours and keeps its storage), every later age
// at config.StorageHoldHours.
func TestStorageCovenantHours(t *testing.T) {
	if config.StorageHoldHours != 4.5 || config.EarlyStorageHoldHours != 1.5 {
		t.Fatalf("covenant hours %g and %g; the docs and the storage table say 4.5 and 1.5", config.StorageHoldHours, config.EarlyStorageHoldHours)
	}
	for _, r := range StaticStorage() {
		want := config.StorageHoldHours
		if r.Age == "primitive_age" || r.Age == "stone_age" {
			want = config.EarlyStorageHoldHours
		}
		if r.Want() != want {
			t.Errorf("%s is graded at %g h, want %g h", r.Age, r.Want(), want)
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
