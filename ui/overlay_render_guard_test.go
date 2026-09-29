package ui

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// overlay_render_guard_test.go renders every text panel the dashboard
// registers against a busy mid-game state and checks what the player would
// actually read (wording audit, systemic fixes 1, 2 and 11):
//
//  1. no [word] the panel meant to print is eaten by tview as a color tag;
//  2. no raw config key (lumber_mill, iron_ore, bronze_age) reaches the
//     screen, apart from the listed spots that echo a key on purpose;
//  3. the same state renders the same text every time (no map-order drift).

// guardPanelWidth is the terminal width the panels are rendered at.
const guardPanelWidth = 120

// guardPanels builds a dashboard and returns its registered text panels by
// name, so a panel added to the dashboard is covered without editing this
// test.
func guardPanels(t *testing.T, engine *game.GameEngine) map[string]OverlayProvider {
	t.Helper()
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), engine, pages)
	out := make(map[string]OverlayProvider, len(d.overlayMgr.entries))
	for name, e := range d.overlayMgr.entries {
		out[name] = e.provide
	}
	if len(out) < 15 {
		t.Fatalf("only %d text panels registered; the dashboard registration list moved?", len(out))
	}
	return out
}

// guardStates returns a fresh game and a busy mid-game one: met civs with
// deals and live boons, market pressure, blockaded routes, a plan, research
// in progress, workers, history, legacy bonuses, loot, a harbinger and a log
// with every severity.
func guardStates(t *testing.T) (map[string]game.GameState, *game.GameEngine) {
	t.Helper()
	engine := game.NewGameEngine()
	fresh := engine.GetState()

	for key, opinion := range map[string]int{"riverlands_tribes": 60, "ironhold_clans": -40, "merchant_guild": 30} {
		if err := engine.MeetFactionForTest(key, opinion); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.SummonHarbingerForTest("industrial_age"); err != nil {
		t.Fatal(err)
	}
	s := engine.GetState()
	s.Age, s.AgeName = "industrial_age", "Industrial Age"
	s.TickIntervalMs = 2000

	// Buildings: everything up to the Industrial Age unlocked, some built
	// and staffed.
	order := config.AgeOrder()
	ageIdx := map[string]int{}
	for i, a := range order {
		ageIdx[a] = i
	}
	bkeys := make([]string, 0, len(s.Buildings))
	for k := range s.Buildings {
		bkeys = append(bkeys, k)
	}
	sort.Strings(bkeys)
	for i, k := range bkeys {
		bs := s.Buildings[k]
		if ageIdx[bs.AgeKey] <= ageIdx["industrial_age"] {
			bs.Unlocked = true
			if i%2 == 0 {
				bs.Count = 2
				if bs.WorkerDomain != "" {
					bs.WorkersAssigned = bs.WorkerCapacity
				}
			}
		}
		s.Buildings[k] = bs
	}

	// Research: a third researched, a third available, one in progress.
	tkeys := make([]string, 0, len(s.Research.Techs))
	for k := range s.Research.Techs {
		tkeys = append(tkeys, k)
	}
	sort.Strings(tkeys)
	for i, k := range tkeys {
		ts := s.Research.Techs[k]
		switch i % 3 {
		case 0:
			ts.Researched = true
		case 1:
			ts.Available, ts.PrereqsMet = true, true
			if s.Research.CurrentTech == "" {
				s.Research.CurrentTech, s.Research.CurrentTechName = k, ts.Name
				s.Research.TicksLeft, s.Research.TotalTicks = 40, 100
			}
		}
		s.Research.Techs[k] = ts
	}
	s.Research.Bonuses = map[string]float64{"production_all": 0.1, "gather_rate": 0.05, "iron_rate": 0.2, "research_speed": 0.15}

	// Workers.
	wt := s.Workers.Types["worker"]
	wt.Unlocked, wt.Count, wt.IdleCount = true, 40, 3
	wt.Assignments = map[string]int{"farm": 10, "lumber_mill": 8, "iron_mine": 6}
	s.Workers.Types["worker"] = wt
	s.Workers.TotalPop, s.Workers.MaxPop, s.Workers.TotalIdle, s.Workers.FoodDrain = 40, 60, 3, 4.8
	s.Morale, s.MoraleCap = 0.7, 1.2

	// Resources: a few unlocked with rates, food falling.
	for k, rate := range map[string]float64{"food": -1.25, "wood": 3.5, "iron_ore": 0.04, "knowledge": 12} {
		rs := s.Resources[k]
		rs.Unlocked, rs.Rate, rs.Amount = true, rate, 500
		s.Resources[k] = rs
	}

	// Live civilization effects and a plain event.
	s.ActiveEvents = []game.ActiveEventState{
		{Name: "Specialty Windfall", Key: "faction_boon_riverlands_tribes", TicksLeft: 142,
			Effects: []game.EventEffectInfo{{Type: "food_rate", Target: "food", Value: 0.13}}},
		{Name: "Cursed Relic", Key: "faction_malus_ironhold_clans", TicksLeft: 20,
			Effects: []game.EventEffectInfo{{Type: "iron_ore_rate", Value: -0.11}, {Type: "production", Target: "iron_ore", Value: -0.5}}},
		{Name: "Golden Age", Key: "golden_age", TicksLeft: 60,
			Effects: []game.EventEffectInfo{{Type: "production_all", Value: 0.08}, {Type: "tick_speed", Value: 0.1}, {Type: "gather_rate", Value: 0.2}}},
	}
	iron := s.Diplomacy.Factions["ironhold_clans"]
	iron.AtWar, iron.Status, iron.DealsBlocked = true, "rival", "at war with you"
	s.Diplomacy.Factions["ironhold_clans"] = iron
	guild := s.Diplomacy.Factions["merchant_guild"]
	guild.DealRefreshIn = 90
	guild.Deals = []game.DealInfo{
		{Num: 1, Kind: game.DealWant, Give: "iron_ore", GiveAmt: 2100, Get: "food", GetAmt: 3000, Edge: 0.13},
		{Num: 2, Kind: game.DealFavor, Give: "iron_ore", GiveAmt: 727, Standing: 5},
		{Num: 3, Kind: game.DealRare, Give: "gold", GiveAmt: 10, Get: "dark_matter", GetAmt: 1},
		{Num: 4, Kind: "", Give: "wood", GiveAmt: 10, Get: "iron_ore", GetAmt: 12, Taken: true},
	}
	s.Diplomacy.Factions["merchant_guild"] = guild
	river := s.Diplomacy.Factions["riverlands_tribes"]
	river.Status, river.TradeBonus, river.TradeCount, river.LentWorkers = "allied", 0.15, 4, 3
	s.Diplomacy.Factions["riverlands_tribes"] = river
	s.Military.AutoExpedition = game.AutoExpeditionState{Active: true, TicksLeft: 50, Interval: 200, Count: 1, Assigned: 2, Capacity: 5}
	s.Military.TotalLoot = map[string]float64{"iron_ore": 1234, "gold": 56}
	s.Military.ExpeditionBonus = 0.1

	// Trade: pressure both ways, a blockade, routes in every state.
	s.Trade.TradeBuildings = 0
	s.Trade.ExchangeRates = map[string]game.ExchangeRateInfo{
		"food:wood":     {From: "food", To: "wood", Rate: 0.3, BaseRate: 1, Pressure: 2.5},
		"wood:iron_ore": {From: "wood", To: "iron_ore", Rate: 0.55, BaseRate: 0.5, Pressure: -0.3},
		"stone:gold":    {From: "stone", To: "gold", Rate: 0.1, BaseRate: 0.1},
	}
	s.Trade.TotalExchanged = map[string]float64{"food": 1500, "iron_ore": 20}
	s.Trade.TotalSold = map[string]float64{"food": 1500}
	s.Trade.TotalBought = map[string]float64{"iron_ore": 20, "dark_matter": 3}
	s.Trade.DisruptedResources = []string{"iron_ore"}
	s.Trade.ActiveRoutes = []game.ActiveRouteInfo{
		{Name: "Tin Road", Key: "tin_road", TicksLeft: 30, CyclesDone: 1,
			Export: map[string]float64{"wood": 50}, Import: map[string]float64{"iron_ore": 20}, Disrupted: true, DisruptedBy: "iron_ore"},
		{Name: "Amber Route", Key: "amber_route", TicksLeft: 12, CyclesDone: 3,
			Export: map[string]float64{"food": 40, "stone": 10}, Import: map[string]float64{"gold": 5}},
	}
	s.Trade.AvailableRoutes = []game.TradeRouteInfo{
		{Name: "Salt Path", Key: "salt_path", Export: map[string]float64{"food": 10}, Import: map[string]float64{"gold": 2},
			CanStart: false, RequiredBld: "black_market", MinCount: 2, Description: "Salt for gold."},
		{Name: "Spice Lane", Key: "spice_lane", Export: map[string]float64{"wood": 10}, Import: map[string]float64{"gold": 2},
			CanStart: true, RequiredBld: "market", MinCount: 1, Description: "Wood for gold."},
	}

	// Plan: every kind and status.
	s.Plan = []game.PlanItemView{
		{Kind: game.PlanBuild, Key: "lumber_mill", Name: "Lumber Mill", Count: 3, Started: 1, Status: game.PlanStatusWaiting,
			Progress: 0.4, Short: "iron_ore", Cost: map[string]float64{"iron_ore": 300, "wood": 100}},
		{Kind: game.PlanTrade, Key: "food", To: "iron_ore", Amount: 500, Got: 300, Status: game.PlanStatusReady, Count: 1},
		{Kind: game.PlanTrade, Key: "wood", To: "gold", Status: game.PlanStatusBlocked, Note: "no market", Count: 1},
		{Kind: game.PlanResearch, Key: tkeys[1], Name: s.Research.Techs[tkeys[1]].Name, Status: game.PlanStatusReady, Count: 1,
			Cost: map[string]float64{"knowledge": 1200}},
	}

	// History with a negative food rate, epochs, legacy and the log.
	s.History = &game.HistoryCollector{
		Samples: []game.HistorySample{
			{Tick: 10, Population: 5, FoodRate: 1, KnowRate: 2, Faith: 3, ProdAll: 0.1, TickSpeed: 1, Morale: 0.5},
			{Tick: 20, Population: 9, FoodRate: -3.2, KnowRate: 4, Faith: 5, ProdAll: 0.2, TickSpeed: 1.5, Morale: 0.6},
			{Tick: 30, Population: 12, FoodRate: -1.5, KnowRate: 5, Faith: 8, ProdAll: 0.2, TickSpeed: 1.5, Morale: 0.7},
		},
		AgeMarkers: []game.AgeMarker{{Tick: 20, AgeName: "Bronze Age"}},
	}
	s.LegacyBonuses = map[string]bool{"stone_era": true, "iron_era": true}
	s.CatastrophesEndured, s.CatastrophesSuccumbed = 1, 1
	s.Stats.AgesReached = []string{"primitive_age", "stone_age", "bronze_age", "industrial_age"}
	s.Stats.TotalGathered = map[string]float64{"food": 12000, "iron_ore": 3400, "dark_matter": 1}
	s.BuildQueue = []game.BuildQueueSnapshot{{Name: "Farm", TicksLeft: 5, TotalTicks: 10}}
	s.Log = []game.LogEntry{
		{Tick: 1, Type: "success", Message: "Farm built (you now have 3)."},
		{Tick: 2, Type: "warning", Message: "Food is low."},
		{Tick: 3, Type: "error", Message: "Usage: sell <building> [count]"},
		{Tick: 4, Type: "event", Message: "A festival starts."},
		{Tick: 5, Type: "info", Message: "Saved."},
	}
	s.Prestige.Level = 1

	return map[string]game.GameState{"fresh": fresh, "midgame": s}, engine
}

// guardTagWordRe finds [word] spans tview could read as a style tag.
var guardTagWordRe = regexp.MustCompile(`\[([A-Za-z][A-Za-z0-9#:\-]*)\]`)

// guardEscapedWordRe finds tview-escaped "[word[]" spans, which print as
// "[word]".
var guardEscapedWordRe = regexp.MustCompile(`\[([A-Za-z][A-Za-z0-9#:\-]*)\[\]`)

// guardRendered is what a panel's TextView shows: the provider output after
// the OverlayManager's safeTags pass, with every tag stripped.
func guardRendered(raw string) string {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetText(safeTags(raw))
	return tv.GetText(true)
}

// TestPanelsKeepBracketedWords: every [word] a panel writes that is not a
// real color tag, and every escaped "[word[]", must still be on screen.
func TestPanelsKeepBracketedWords(t *testing.T) {
	states, engine := guardStates(t)
	for name, provide := range guardPanels(t, engine) {
		for sname, st := range states {
			raw := provide(st, guardPanelWidth)
			shown := guardRendered(raw)
			var want []string
			for _, m := range guardTagWordRe.FindAllStringSubmatch(raw, -1) {
				if !isStyleTag(m[1]) {
					want = append(want, "["+m[1]+"]")
				}
			}
			for _, m := range guardEscapedWordRe.FindAllStringSubmatch(raw, -1) {
				want = append(want, "["+m[1]+"]")
			}
			for _, w := range want {
				if !strings.Contains(shown, w) {
					t.Errorf("%s panel (%s): %q is eaten as a color tag", name, sname, w)
				}
			}
		}
	}
}

// guardKeyRe matches a snake_case config key token.
var guardKeyRe = regexp.MustCompile(`\b[a-z]+_[a-z_]+\b`)

// guardKeyEchoes are the spots that print a key on purpose, matched against
// the text just before the key on its line. Each is a command hint or a
// "Name (key)" listing: the player needs the key to type the command.
var guardKeyEchoes = []*regexp.Regexp{
	// "Lumber Mill (lumber_mill)": Buildings and Research list the key the
	// build/research command takes beside each name.
	regexp.MustCompile(`\($`),
	// Command hints that end in the key they act on.
	regexp.MustCompile(`\b(build|research|expedition|campaign|trade route start|trade route stop|plan deal|diplomacy [a-z/]+|ally/rival/embargo/neutral)\s+$`),
}

// guardKeyExemptPanels are panels whose text another change owns.
var guardKeyExemptPanels = map[string]string{}

// guardKeyWords are snake_case words that are English, not keys.
var guardKeyWords = map[string]bool{}

// guardKnownLeaks are raw keys that come from text outside ui/ and are still
// being fixed at the source. Each entry names the panel and a pattern on the
// line; delete it once the source is fixed so the guard covers that text too.
var guardKnownLeaks = []struct {
	panel string
	line  *regexp.Regexp
	why   string
}{}

// guardKnownLeak reports whether line on panel is a listed known leak.
func guardKnownLeak(panel, line string) bool {
	for _, k := range guardKnownLeaks {
		if k.panel == panel && k.line.MatchString(line) {
			return true
		}
	}
	return false
}

// TestPanelsShowNoRawKeys: the tag-stripped panel text carries no config key
// outside the key-echo spots above.
func TestPanelsShowNoRawKeys(t *testing.T) {
	states, engine := guardStates(t)
	for name, provide := range guardPanels(t, engine) {
		if _, skip := guardKeyExemptPanels[name]; skip {
			continue
		}
		for sname, st := range states {
			shown := guardRendered(provide(st, guardPanelWidth))
			for _, line := range strings.Split(shown, "\n") {
				if guardKnownLeak(name, line) {
					continue
				}
				for _, m := range guardKeyRe.FindAllStringIndex(line, -1) {
					key := line[m[0]:m[1]]
					if guardKeyWords[key] {
						continue
					}
					before := line[:m[0]]
					echo := false
					for _, re := range guardKeyEchoes {
						if re.MatchString(before) {
							echo = true
							break
						}
					}
					if !echo {
						t.Errorf("%s panel (%s): raw key %q in %q", name, sname, key, strings.TrimSpace(line))
					}
				}
			}
		}
	}
}

// TestPanelsRenderStably: the same state renders the same text every time.
// Several renders give a map-order bug a fair chance to show.
func TestPanelsRenderStably(t *testing.T) {
	states, engine := guardStates(t)
	for name, provide := range guardPanels(t, engine) {
		for sname, st := range states {
			first := provide(st, guardPanelWidth)
			for i := 0; i < 8; i++ {
				if again := provide(st, guardPanelWidth); again != first {
					t.Errorf("%s panel (%s): two renders of one state differ:\n--- first\n%s\n--- again\n%s", name, sname, first, again)
					break
				}
			}
		}
	}
}
