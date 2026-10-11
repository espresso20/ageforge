package ui

import (
	"fmt"

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

// storageBuildingName is the current age's storage building, "" when the age
// has none.
func storageBuildingName(state game.GameState) string {
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
// when a line needs it.
type storeWall struct {
	state  game.GameState
	build  string
	loaded bool
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
	if !sw.loaded {
		sw.build, sw.loaded = storageBuildingName(sw.state), true
	}
	return storeTooSmallLine(bs.Name, res, bs.NextCost[res], sw.state.Resources[res].Storage, sw.build)
}
