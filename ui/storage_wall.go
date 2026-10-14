package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
// current age: the age's storage building. "" when the age has none.
func storageBuildingName(state game.GameState, res string) string {
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

// storageLimitLine is the same refusal for a store that cannot grow: the
// age's storage building has run out of copies, so it names no building to
// build and says where bigger storage comes from.
func storageLimitLine(what, res string, amount, store float64, nextAge bool) string {
	line := fmt.Sprintf("%s costs %s, more than your %s storage holds (%s). Your storage is at its limit for this age.", what, game.Amount(amount, res), game.ResourceName(res), FormatNumber(store))
	if nextAge {
		line += " The next age has a bigger storage building."
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
	if bs.StoreLimit || sw.storageAtLimit() {
		return storageLimitLine(bs.Name, res, bs.NextCost[res], sw.state.Resources[res].Storage, sw.state.NextAge != "" && sw.state.Ruleset().AgeHasStorageBuilding(sw.state.NextAge))
	}
	return storeTooSmallLine(bs.Name, res, bs.NextCost[res], sw.state.Resources[res].Storage, sw.buildingFor(res))
}

// storageAtLimit reports whether the age's storage building has run out of
// copies the store could pay for: naming it would send the player in a circle.
func (sw *storeWall) storageAtLimit() bool {
	for key, d := range sw.state.Ruleset().BuildingMap() {
		if d.Category == "storage" && d.RequiredAge == sw.state.Age {
			return sw.state.Buildings[key].StoreLimit
		}
	}
	return false
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
// building also gives its copies against the most that can be built ("3 of 5
// built.") and, when the description does not say it (it does: "comes with 2
// scholars"), its scholars a copy.
func buildingText(bs game.BuildingState) string {
	if bs.Scholars <= 0 {
		return bs.Description
	}
	count := textfmt.Count(bs.Scholars, "scholar", "scholars")
	line := ""
	if !strings.Contains(bs.Description, count) {
		line = count + " a copy"
	}
	if bs.ScholarCap > 0 {
		if line != "" {
			line += ", "
		}
		line += fmt.Sprintf("%d of %d built", bs.Count, bs.ScholarCap)
	}
	if line == "" {
		return bs.Description
	}
	return strings.TrimSpace(bs.Description + " " + textfmt.Capitalize(line) + ".")
}

// atLimit is the words the Buildings panel gives a building that can be
// built no further, for a knowledge building at its cap in the lists that do
// not show them already.
func atLimit(bs game.BuildingState) string {
	if bs.Scholars > 0 && bs.AtMaxCount {
		return "Building limit reached."
	}
	return ""
}
