package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// storage_wall.go is the storage wall as the screens show it: a price larger
// than a store cannot be paid however long the player waits, so a building
// whose next copy costs more than a store holds says so before the player
// tries, in the words of the refusal the engine gives when they do.

// overStoreRes is the first resource (in key order) of which cost asks more
// than its store holds and more than the town holds, or "". It is the
// engine's own test (stock kept above a store from an older save can still
// be spent, so a price it covers is not over), read from the state.
func overStoreRes(state game.GameState, cost map[string]float64) string {
	for _, res := range sortedMapKeys(cost) {
		rs := state.Resources[res]
		if cost[res] > rs.Storage && cost[res] > rs.Amount {
			return res
		}
	}
	return ""
}

// storageBuildingName is the building that raises the store of res in the
// current age: the age's storage building, or for knowledge the age's
// knowledge building (the engine's own choice). "" when the age has none.
func storageBuildingName(state game.GameState, res string) string {
	if res == "knowledge" {
		if def, ok := state.Ruleset().Building(state.KnowledgeStore.NextBuilding); ok {
			return def.Name
		}
		return ""
	}
	for _, d := range state.Ruleset().Buildings() {
		if d.Category == "storage" && d.RequiredAge == state.Age {
			return d.Name
		}
	}
	return ""
}

// storeTooSmallLine is the engine's refusal for a price larger than a store
// (game.storeTooSmallText), word for word: "Granary costs 150 food, more
// than your Food storage holds (100). Build more storage first: Stash."
func storeTooSmallLine(what, res string, amount, store float64, build string) string {
	line := fmt.Sprintf("%s costs %s, more than your %s storage holds (%s).", what, game.Amount(amount, res), game.ResourceName(res), FormatNumber(store))
	if build != "" {
		line += " Build more storage first: " + build + "."
	}
	return line
}

// storeWall tells, for each building of a state, whether its next copy is
// over a store. The storage building it names is looked up once, and only
// when a line needs it (per resource: knowledge has its own).
type storeWall struct {
	state game.GameState
	build map[string]string
}

// line is the over-the-store line for bs, or "" for a building whose next
// copy can be paid. A wonder is paid through its bank, a superseded building
// cannot be built at all, and one at its limit has no next copy.
func (sw *storeWall) line(bs game.BuildingState) string {
	if bs.Category == "wonder" || bs.IsLegacy || bs.AtMaxCount || !bs.Unlocked {
		return ""
	}
	res := overStoreRes(sw.state, bs.NextCost)
	if res == "" {
		return ""
	}
	return storeTooSmallLine(bs.Name, res, bs.NextCost[res], sw.state.Resources[res].Storage, sw.buildingFor(res))
}

// buildingFor is the building the over-the-store line names for res, looked
// up once per resource.
func (sw *storeWall) buildingFor(res string) string {
	if b, ok := sw.build[res]; ok {
		return b
	}
	if sw.build == nil {
		sw.build = map[string]string{}
	}
	sw.build[res] = storageBuildingName(sw.state, res)
	return sw.build[res]
}

// buildingText is a building's description as the lists show it. A knowledge
// building adds what one copy holds, in the form a storage building's
// description gives its figure: "+140 knowledge storage."
func buildingText(bs game.BuildingState) string {
	if bs.KnowledgeStore <= 0 {
		return bs.Description
	}
	return strings.TrimSpace(bs.Description + " +" + config.FormatAmount(bs.KnowledgeStore) + " knowledge storage.")
}
