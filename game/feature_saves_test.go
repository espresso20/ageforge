package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// savedFeatures reads the granted features back from a save file.
func savedFeatures(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatal(err)
	}
	var gs GameSave
	noErr(t, json.Unmarshal(data, &gs))
	return gs.GrantedFeatures
}

// stock gives the game plenty of everything a command might ask for, and an
// army at home to make up any party.
func stock(ge *GameEngine) {
	for _, res := range []string{"food", "wood", "stone", "iron", "gold", "culture", "steel", "coal"} {
		setResource(ge, res, config.Billion)
	}
	setSoldiers(ge, config.Trillion)
}

// clearArmy calls every expedition home, so the next can be sent.
func clearArmy(ge *GameEngine) {
	ge.mu.Lock()
	ge.Military.clearActive()
	ge.mu.Unlock()
}

// grantedList is the run's granted features, sorted and joined.
func grantedList(ge *GameEngine) string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return strings.Join(ge.grantedFeatureKeys(), " ")
}

// TestFeatureLocksRefuseAndOpen: each live lock refuses its command with the
// name of its tech and opens when the tech is researched; missions refuse
// nothing for want of a tech; and a
// new game saves no granted feature until it uses a command without its
// tech.
func TestFeatureLocksRefuseAndOpen(t *testing.T) {
	isolateAccountDir(t)
	at := func(age string) *GameEngine {
		ge := newSeededEngine(5)
		for _, a := range ageKeys() {
			if a == "primitive_age" {
				continue
			}
			noErr(t, ge.EnterAgeForTest(a))
			if a == age {
				break
			}
		}
		stock(ge)
		return ge
	}
	refused := func(what string, err error, want string) {
		t.Helper()
		if err == nil || err.Error() != want {
			t.Errorf("%s: %v, want %q", what, err, want)
		}
	}
	route := func(ge *GameEngine, key string) {
		ge.mu.Lock()
		ge.Buildings.counts[ge.Trade.routeDefs[key].RequiredBld] = ge.Trade.routeDefs[key].MinCount
		ge.mu.Unlock()
	}

	// Missions wait for no tech: a scout, a long mission and a campaign all go
	// out in the Bronze Age, with nothing researched.
	ge := at("bronze_age")
	for _, key := range []string{"scouting_party", "long_hunt", "raid_the_rival_clan"} {
		if err := ge.LaunchExpedition(keyOffer(ge, key)); err != nil {
			t.Errorf("%s without a tech: %v", key, err)
		}
	}
	clearArmy(ge)

	// Trade routes: The Wheel first, then the route's own needs.
	route(ge, "local_barter")
	refused("Local Barter without The Wheel", ge.StartTradeRoute("local_barter"), "Trade routes need The Wheel first. Research it to start one.")
	if f := ge.GetState().Features[config.FeatureTradeRoutes]; !f.Live || f.Open || f.Granted || f.TechName != "The Wheel" {
		t.Errorf("trade routes read %+v, want a live lock, shut", f)
	}
	learn(ge, "the_wheel")
	if err := ge.StartTradeRoute("local_barter"); err != nil {
		t.Errorf("Local Barter with The Wheel: %v", err)
	}
	// Researched, so nothing is granted: the save stays as it was.
	if got := grantedList(ge); got != "" {
		t.Errorf("commands used with their techs researched granted %q", got)
	}
	saveAndLoad(t, ge, "feature_lock_fresh")
	if got := strings.Join(savedFeatures(t, "feature_lock_fresh"), " "); got != "" {
		t.Errorf("the save carries granted features %q, want none", got)
	}

	// Diplomacy and festivals, from the Classical Age.
	ge = at("classical_age")
	meet(ge, "riverlands_tribes", 0)
	const envoys = "Gifts, alliances, rivalries and deals need Envoys first. Research it to deal with other civilizations."
	refused("a gift without Envoys", ge.SendGift("riverlands_tribes"), envoys)
	refused("an alliance without Envoys", ge.SetDiplomaticStatus("riverlands_tribes", "allied"), envoys)
	learn(ge, "envoys")
	if err := ge.SendGift("riverlands_tribes"); err != nil {
		t.Errorf("a gift with Envoys: %v", err)
	}
	refused("a festival without Drama", ge.DoFestival(), "Festivals need Drama first. Research it to hold one.")
	learn(ge, "drama")
	if err := ge.DoFestival(); err != nil {
		t.Errorf("a festival with Drama: %v", err)
	}

	ge = at("renaissance_age")

	// The black market: its age first, then its tech.
	if _, _, err := ge.DoBlackMarket("gold"); err == nil || !strings.HasPrefix(err.Error(), "The black market opens in") {
		t.Errorf("the black market in the Renaissance Age: %v, want its age named", err)
	}
	ge = at("colonial_age")
	if _, _, err := ge.DoBlackMarket("gold"); err == nil || err.Error() != "The black market needs Mercantilism first. Research it to deal there." {
		t.Errorf("the black market without Mercantilism: %v", err)
	}
	learn(ge, "mercantilism")
	if _, _, err := ge.DoBlackMarket("gold"); err != nil {
		t.Errorf("the black market with Mercantilism: %v", err)
	}

	// Rail Freight, from the Industrial Age: trade routes' lock, then the
	// route's own.
	ge = at("industrial_age")
	route(ge, "rail_freight")
	refused("Rail Freight without The Wheel", ge.StartTradeRoute("rail_freight"), "Trade routes need The Wheel first. Research it to start one.")
	learn(ge, "the_wheel")
	refused("Rail Freight without Railroads", ge.StartTradeRoute("rail_freight"), "The Rail Freight route needs Railroads first. Research it to start it.")
	learn(ge, "railroads")
	if err := ge.StartTradeRoute("rail_freight"); err != nil {
		t.Errorf("Rail Freight with Railroads: %v", err)
	}

	// Warp Commerce, from the Interstellar Age: trade routes' lock, then
	// the route's own, the last to find its tech.
	ge = at("interstellar_age")
	learn(ge, "the_wheel")
	route(ge, "warp_commerce")
	refused("Warp Commerce without Interstellar Trade", ge.StartTradeRoute("warp_commerce"), "The Warp Commerce route needs Interstellar Trade first. Research it to start it.")
	if f := ge.GetState().Features[config.FeatureRouteWarpCommerce]; !f.Live || f.Open || f.TechName != "Interstellar Trade" {
		t.Errorf("Warp Commerce reads %+v, want a live lock, shut, that names Interstellar Trade", f)
	}
	learn(ge, "interstellar_trade")
	if err := ge.StartTradeRoute("warp_commerce"); err != nil {
		t.Errorf("Warp Commerce with Interstellar Trade: %v", err)
	}
	if f := ge.GetState().Features[config.FeatureRouteWarpCommerce]; !f.Live || !f.Open {
		t.Errorf("Warp Commerce reads %+v with its tech researched, want a live lock, open", f)
	}
}

// TestFeatureLockSwitchesOnWithItsTech: a lock whose tech the tree does not
// hold is inert, its command open, and it switches on when a ruleset adds a
// tech with that key, with no other change. Every lock of the game has its
// tech now, so the test makes the case: a ruleset where Warp Commerce's lock
// names a tech that does not exist, and one that adds it. A game on the
// first starts the route, and the save remembers it. A new game on the
// second meets the lock. The game that was already using the command,
// saved before the tech existed, keeps it for the rest of its run.
func TestFeatureLockSwitchesOnWithItsTech(t *testing.T) {
	isolateAccountDir(t)
	src := rules.FromConfig()
	locks := append([]config.FeatureLockDef(nil), src.FeatureLocks...)
	for i := range locks {
		if locks[i].Key == config.FeatureRouteWarpCommerce {
			locks[i].Tech = "wormhole_charters"
		}
	}
	src.FeatureLocks = locks
	inert := rules.Compile(src)
	src.Techs = append(append([]config.TechDef(nil), src.Techs...), config.TechDef{
		Key: "wormhole_charters", Name: "Wormhole Charters", Age: "interstellar_age", Lane: config.LaneTrade, Cost: 10,
	})
	withTech := rules.Compile(src)
	const warp = "warp_commerce"
	port := func(ge *GameEngine) {
		ge.mu.Lock()
		def := ge.Trade.routeDefs[warp]
		ge.Buildings.counts[def.RequiredBld] = def.MinCount
		ge.age = def.MinAge
		ge.mu.Unlock()
		learn(ge, "the_wheel") // trade routes themselves are open
		stock(ge)
	}

	// The tree without the tech: the lock is inert, the route starts, and
	// the save remembers it.
	old := NewGameEngineWith(inert)
	old.SeedRNG(6)
	port(old)
	if f := old.GetState().Features[config.FeatureRouteWarpCommerce]; f.Live || !f.Open {
		t.Fatalf("on a tree without its tech Warp Commerce reads %+v, want an inert lock, open", f)
	}
	if err := old.StartTradeRoute(warp); err != nil {
		t.Fatalf("Warp Commerce on a tree without its tech: %v", err)
	}
	if got := grantedList(old); got != "route_warp_commerce" {
		t.Errorf("a route started under an inert lock granted %q, want route_warp_commerce", got)
	}
	if f := old.GetState().Features[config.FeatureRouteWarpCommerce]; f.Live || !f.Open || !f.Granted {
		t.Errorf("Warp Commerce reads %+v, want an inert lock, open, granted", f)
	}
	noErr(t, old.SaveGame("before_wormhole_charters"))
	if got := strings.Join(savedFeatures(t, "before_wormhole_charters"), " "); got != "route_warp_commerce" {
		t.Errorf("the save carries granted features %q, want route_warp_commerce", got)
	}

	// The tree with the tech: a new game is refused until it researches it.
	fresh := NewGameEngineWith(withTech)
	port(fresh)
	if err := fresh.StartTradeRoute(warp); err == nil || err.Error() != "The Warp Commerce route needs Wormhole Charters first. Research it to start it." {
		t.Errorf("Warp Commerce on a tree with its tech, not researched: %v", err)
	}
	learn(fresh, "wormhole_charters")
	if err := fresh.StartTradeRoute(warp); err != nil {
		t.Errorf("Warp Commerce with its tech researched: %v", err)
	}

	// The game saved before the tech existed, loaded on the new tree: its
	// route is running, and it can stop it and start it again.
	loaded := NewGameEngineWith(withTech)
	noErr(t, loaded.LoadGame("before_wormhole_charters"))
	if loaded.Research.IsResearched("wormhole_charters") {
		t.Fatal("the lock's tech is researched: the save does not test the grant")
	}
	if err := loaded.StopTradeRoute(warp); err != nil {
		t.Fatalf("stopping Warp Commerce after the load: %v", err)
	}
	if err := loaded.StartTradeRoute(warp); err != nil {
		t.Errorf("Warp Commerce in a game that ran it before its tech existed: %v", err)
	}
}

// TestResearchSaysWhatItOpens: finishing a tech that opens a command logs
// it, in the game's words.
func TestResearchSaysWhatItOpens(t *testing.T) {
	ge := newSeededEngine(8)
	noErr(t, ge.EnterAgeForTest("stone_age"))
	noErr(t, ge.EnterAgeForTest("bronze_age"))
	ge.mu.Lock()
	ge.Research.currentTech = "drama"
	ge.Research.progress = ge.Research.defs["drama"].Cost - 1
	ge.researchPerTick = 1
	ge.processResearch()
	ge.mu.Unlock()
	if !ge.Research.IsResearched("drama") || !logHas(ge, "Drama opens festivals.") {
		t.Errorf("finishing Drama did not say it opens festivals; the log ends %q", lastLog(ge))
	}
}
