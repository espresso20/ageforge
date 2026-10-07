package config

import "fmt"

// A feature lock is a command, or a part of one, that waits for a tech:
// campaigns wait for Military Tactics, the Rail Freight route for Railroads.
// Until the tech is researched the engine refuses the command and says which
// tech opens it.
//
// A lock names its tech by key. A lock whose tech is not in the tree is
// inert: the command is open, as it was before the locks. So the table lists
// all nine locks the tree is designed around, and five of them wait for
// content: adding a tech with the key a lock names switches that lock on,
// with nothing else to change (LiveFeatureLocks tells the two apart, and the
// static research check lists the ones still waiting).
//
// Old saves: a game that already used a command keeps it for the rest of its
// run, whatever it has researched (the game's GameSave.GrantedFeatures), and
// a save from before the locks has every command open in the age it was in
// (GameSave.TreeGraceAge).

// The keys of the feature locks.
const (
	FeatureTradeRoutes       = "trade_routes"
	FeatureCampaigns         = "campaigns"
	FeatureExpeditions       = "expeditions"
	FeatureDiplomacy         = "diplomacy"
	FeatureFestivals         = "festivals"
	FeatureNavalExpedition   = "naval_expedition"
	FeatureBlackMarket       = "black_market"
	FeatureRouteRailFreight  = "route_rail_freight"
	FeatureRouteWarpCommerce = "route_warp_commerce"
)

// FeatureLockDef is one command that waits for a tech.
type FeatureLockDef struct {
	Key string
	// Name is the command in running text, as the subject of a sentence:
	// "Trade routes", "The black market".
	Name string
	// Tech is the key of the tech that opens it.
	Tech string
	// Needs is the verb that follows Name: "need" for a plural, "needs" for
	// a singular.
	Needs string
	// Then finishes the refusal after "Research it to ": "start one".
	Then string
	// Opens is the lock as an effect of its tech, in player words: "opens
	// trade routes".
	Opens string
}

// FeatureLocks lists every feature lock, in the order the tree opens them.
func FeatureLocks() []FeatureLockDef {
	return []FeatureLockDef{
		{Key: FeatureTradeRoutes, Name: "Trade routes", Tech: "the_wheel",
			Needs: "need", Then: "start one", Opens: "opens trade routes"},
		{Key: FeatureCampaigns, Name: "Campaigns", Tech: "military_tactics",
			Needs: "need", Then: "send one", Opens: "opens campaigns"},
		{Key: FeatureExpeditions, Name: "Expeditions past the Scout Party", Tech: "exploration",
			Needs: "need", Then: "send one", Opens: "opens expeditions past the Scout Party"},
		{Key: FeatureDiplomacy, Name: "Gifts, alliances, rivalries and deals", Tech: "envoys",
			Needs: "need", Then: "deal with other civilizations", Opens: "opens diplomacy"},
		{Key: FeatureFestivals, Name: "Festivals", Tech: "drama",
			Needs: "need", Then: "hold one", Opens: "opens festivals"},
		{Key: FeatureNavalExpedition, Name: "The Naval Expedition", Tech: "navigation",
			Needs: "needs", Then: "send it", Opens: "opens the Naval Expedition"},
		{Key: FeatureBlackMarket, Name: "The black market", Tech: "mercantilism",
			Needs: "needs", Then: "deal there", Opens: "opens the black market"},
		{Key: FeatureRouteRailFreight, Name: "The Rail Freight route", Tech: "railroads",
			Needs: "needs", Then: "start it", Opens: "opens the Rail Freight route"},
		{Key: FeatureRouteWarpCommerce, Name: "The Warp Commerce route", Tech: "interstellar_trade",
			Needs: "needs", Then: "start it", Opens: "opens the Warp Commerce route"},
	}
}

// FeatureLockByKey is FeatureLocks by key.
func FeatureLockByKey() map[string]FeatureLockDef {
	m := make(map[string]FeatureLockDef)
	for _, d := range FeatureLocks() {
		m[d.Key] = d
	}
	return m
}

// Refusal is what the game says when the command is used before its tech,
// given the tech's name: "Campaigns need Military Tactics first. Research it
// to send one."
func (d FeatureLockDef) Refusal(techName string) string {
	return fmt.Sprintf("%s %s %s first. Research it to %s.", d.Name, d.Needs, techName, d.Then)
}

// LiveFeatureLocks splits locks by whether isTech knows each one's tech: a
// lock is live when its tech is in the tree, and waiting (inert, its command
// open) when it is not.
func LiveFeatureLocks(locks []FeatureLockDef, isTech func(string) bool) (live, waiting []FeatureLockDef) {
	for _, d := range locks {
		if isTech(d.Tech) {
			live = append(live, d)
		} else {
			waiting = append(waiting, d)
		}
	}
	return live, waiting
}

// FeaturesOpenedBy lists the locks tech opens, in FeatureLocks order.
func FeaturesOpenedBy(tech string) []FeatureLockDef {
	var out []FeatureLockDef
	for _, d := range FeatureLocks() {
		if d.Tech == tech {
			out = append(out, d)
		}
	}
	return out
}

// RouteFeature is the lock a trade route sits behind on top of the lock on
// trade routes as a whole ("" for none): Rail Freight and Warp Commerce have
// their own.
func RouteFeature(route string) string {
	switch route {
	case "rail_freight":
		return FeatureRouteRailFreight
	case "warp_commerce":
		return FeatureRouteWarpCommerce
	}
	return ""
}
