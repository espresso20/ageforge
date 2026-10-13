package game

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
)

// The knowledge store (config/knowledge_rule.go). Knowledge is not a
// physical thing, so it has a store of its own: the base the knowledge
// resource carries (config.KnowledgeBase) plus, for every knowledge building
// standing, its copies times what one copy holds (BuildingDef.KnowledgeStore).
// Nothing else raises it: not a storage building, not a tech's or a wonder's
// bonus to every store, and not Era Mastery, which scales production and the
// general stores but not this one (tech prices do not scale with mastery, so
// the store they are set against does not either). A bonus that names
// knowledge storage by itself (a "storage" effect aimed at "knowledge") still
// applies, unscaled; none exists today (TestKnowledgeStore_NoBonusNamesIt).

// KnowledgeStorePart is one knowledge building's share of the knowledge
// store: what the screens list under "what raises it".
type KnowledgeStorePart struct {
	Key string
	// Count is the copies standing; Each is what one copy holds.
	Count int
	Each  float64
	// Total is Count × Each, the share this building adds.
	Total float64
}

// knowledgeStoreParts is the knowledge buildings standing, in building
// order, with the share of the knowledge store each adds.
func (bm *BuildingManager) knowledgeStoreParts() []KnowledgeStorePart {
	var parts []KnowledgeStorePart
	bm.eachBuilt(func(key string, count int, def config.BuildingDef) {
		// Only a knowledge building carries a figure (a test in config holds
		// it to that), so the figure is the whole test.
		if def.KnowledgeStore <= 0 {
			return
		}
		parts = append(parts, KnowledgeStorePart{
			Key: key, Count: count, Each: def.KnowledgeStore,
			Total: float64(float64(count) * def.KnowledgeStore),
		})
	})
	return parts
}

// knowledgeStoreSize is the size of the knowledge store: base plus the
// buildings' shares plus a bonus that names knowledge storage (specific).
func (bm *BuildingManager) knowledgeStoreSize(base, specific float64) float64 {
	size := base
	for _, p := range bm.knowledgeStoreParts() {
		size += p.Total
	}
	return size + specific
}

// KnowledgeStoreState is what a snapshot says of the knowledge store.
type KnowledgeStoreState struct {
	// Store is the size of the store (the same figure as
	// Resources["knowledge"].Storage). Base is what a town holds with no
	// knowledge building.
	Store, Base float64
	// Parts is each standing knowledge building's share.
	Parts []KnowledgeStorePart
	// NextBuilding is the knowledge building of the current age the player
	// can build ("" when the age has none): what to build to raise the store.
	NextBuilding string
	// NextEach is what one copy of NextBuilding adds to the store.
	NextEach float64
}

// knowledgeBuildingOfAge is the knowledge building of the current age (the
// newest one the player can build: an older age's cannot be rebuilt, and
// its figure is already in the store). ok is false when the age has none.
func (ge *GameEngine) knowledgeBuildingOfAge() (config.BuildingDef, bool) {
	var found config.BuildingDef
	ok := false
	for _, d := range ge.rules.Buildings() {
		if d.Category != "wonder" && d.KnowledgeStore > 0 && d.RequiredAge == ge.age {
			found, ok = d, true
		}
	}
	return found, ok
}

// knowledgeWall refuses a tech priced over the knowledge store that the
// stock held cannot pay for either (stock kept above the store from an older
// save can still be spent). The refusal is the one every store gives
// (storeTooSmallText). Under the write lock; dev god mode waives prices.
func (ge *GameEngine) knowledgeWall(def config.TechDef) error {
	if DevGodMode {
		return nil
	}
	if ge.overStore(map[string]float64{"knowledge": def.Cost}) == "" {
		return nil
	}
	return fmt.Errorf("%s", ge.storeTooSmallText(def.Name, "knowledge", def.Cost))
}
