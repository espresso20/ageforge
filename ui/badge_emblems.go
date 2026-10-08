package ui

import (
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// badge_emblems.go is the badge case's emblem table: the one glyph in the
// middle of a badge. A badge names its emblem (config.BadgeDef.Emblem) and
// the case draws it in the account's glyph tier.
//
// Every emblem is a symbol the maps already draw, in all three tiers: a
// lineage wears its map symbol, an age the town centre of its era, and the
// named ones below borrow a map symbol each. So a badge teaches no glyph
// the player has not met on the map, and whatever font shows the map shows
// the badges. TestBadgeEmblemsAreMapGlyphs holds the table to that.

// badgeEmblem is an emblem: its glyph in each tier, and the hand-drawn
// sprite the badge wears at full size, if it has one.
type badgeEmblem struct {
	glyph   mapmodel.Glyph
	special string
}

// emblemStar is the emblem of a badge that names none, and of the ladders
// that belong to no table: the wonders' star.
var emblemStar = badgeEmblem{glyph: mapmodel.Glyph{ASCII: '*', Unicode: mapmodel.G(mapmodel.SymWonder).Unicode, Nerd: mapmodel.G(mapmodel.SymStarBig).Nerd}}

// badgeEmblems are the named emblems.
var badgeEmblems = map[string]badgeEmblem{
	"star":  emblemStar,
	"hut":   {glyph: mapmodel.G(mapmodel.SymHut)},
	"sun":   {glyph: mapmodel.G(mapmodel.SymSun)},
	"trade": {glyph: mapmodel.G(mapmodel.SymTrade)},
	// What the families wear.
	"wonder":    {glyph: mapmodel.G(mapmodel.SymWonder)},
	"hall":      {glyph: mapmodel.G(mapmodel.SymHall)},
	"housing":   {glyph: mapmodel.G(mapmodel.SymHouse)},
	"knowledge": {glyph: mapmodel.G(mapmodel.SymKnowledge)},
	"faith":     {glyph: mapmodel.G(mapmodel.SymFaith)},
	"culture":   {glyph: mapmodel.G(mapmodel.SymCulture)},
	"monument":  {glyph: mapmodel.G(mapmodel.SymMonument)},
	"engineer":  {glyph: mapmodel.G(mapmodel.SymEngineer)},
	"military":  {glyph: mapmodel.G(mapmodel.SymMilitary)},
	"war":       {glyph: mapmodel.G(mapmodel.SymWar)},
	"diplomacy": {glyph: mapmodel.G(mapmodel.SymDiplomacy)},
	"civ":       {glyph: mapmodel.G(mapmodel.SymCiv)},
	"harbinger": {glyph: mapmodel.G(mapmodel.SymHarbinger)},
	"hazard":    {glyph: mapmodel.G(mapmodel.SymHazard)},
	"ruin":      {glyph: mapmodel.G(mapmodel.SymRuin)},
	"alien":     {glyph: mapmodel.G(mapmodel.SymUFO)},
	// The resources: a store, or the symbol of what makes one.
	"store":  {glyph: mapmodel.G(mapmodel.SymStore)},
	"food":   {glyph: mapmodel.G(mapmodel.SymFood)},
	"wood":   {glyph: mapmodel.G(mapmodel.SymWood)},
	"mine":   {glyph: mapmodel.G(mapmodel.SymMine)},
	"metal":  {glyph: mapmodel.G(mapmodel.SymMetal)},
	"gold":   {glyph: mapmodel.G(mapmodel.SymGold)},
	"energy": {glyph: mapmodel.G(mapmodel.SymEnergy)},
	"data":   {glyph: mapmodel.G(mapmodel.SymDataStore)},
	// The integrity badges: a store for the jar, the scholars' mark for
	// the ledger, the hackers' for the source. Each has a sprite.
	"cookie_jar": {glyph: mapmodel.G(mapmodel.SymStore), special: specialCookieJar},
	"ledger":     {glyph: mapmodel.G(mapmodel.SymKnowledge), special: specialLedger},
	"source":     {glyph: mapmodel.G(mapmodel.SymHacker), special: specialSource},
	// The hand-drawn legendary badges (badge_art_legend.go). The glyph is
	// what the grid and the toast show.
	"sprite.chains":    {glyph: emblemStar.glyph, special: specialChains},
	"sprite.boxes":     {glyph: emblemStar.glyph, special: specialBoxes},
	"sprite.museum":    {glyph: mapmodel.G(mapmodel.SymMonument), special: specialMuseum},
	"sprite.wonders":   {glyph: mapmodel.G(mapmodel.SymWonder), special: specialWonders},
	"sprite.rogues":    {glyph: mapmodel.G(mapmodel.SymHarbinger), special: specialRogues},
	"sprite.endings":   {glyph: mapmodel.G(mapmodel.SymRuin), special: specialEndings},
	"sprite.undying":   {glyph: mapmodel.G(mapmodel.SymHazard), special: specialUndying},
	"sprite.transcend": {glyph: mapmodel.G(mapmodel.SymPlazaCosmic), special: specialTranscend},
}

// The emblems a family writes from its subject.
const (
	emblemLineage = "lineage." // + the lineage's key
	emblemCentre  = "centre."  // + the order of the era
)

// emblemOf resolves a badge's emblem name. ok is false for a name the
// table does not have: the star stands in.
func emblemOf(name string) (e badgeEmblem, ok bool) {
	switch {
	case name == "":
		return emblemStar, true
	case strings.HasPrefix(name, emblemLineage):
		lin := strings.TrimPrefix(name, emblemLineage)
		for _, known := range mapmodel.LineageOrder {
			if known == lin {
				// The lineage as the Iron Era draws it: past the first
				// huts and fields, before the towers.
				return badgeEmblem{glyph: mapmodel.G(mapmodel.LineageSym(lin, 1))}, true
			}
		}
		return emblemStar, false
	case strings.HasPrefix(name, emblemCentre):
		n, err := strconv.Atoi(strings.TrimPrefix(name, emblemCentre))
		if err != nil || n < 0 {
			return emblemStar, false
		}
		return badgeEmblem{glyph: mapmodel.G(mapmodel.CentreSym(n))}, true
	}
	if e, ok := badgeEmblems[name]; ok {
		return e, true
	}
	return emblemStar, false
}

// badgePhase is a badge's own offset into the animations, from its key.
func badgePhase(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % 24)
}

// medalOf is a badge as the renderer draws it, in a glyph tier.
func medalOf(v game.BadgeView, tier mapmodel.GlyphTier) medal {
	e, _ := emblemOf(v.Emblem)
	m := medal{tier: v.Level, emblem: e.glyph.In(tier), crossed: v.Crossed, special: e.special, phase: badgePhase(v.Key)}
	switch {
	case v.Hidden:
		m.state = medalHidden
	case !v.Earned:
		m.state = medalLocked
	}
	return m
}

// animated reports whether a badge's full art moves: platinum glints, the
// legendary ring turns and the hand-drawn sprites glitch. Bronze, silver
// and gold hold still.
func (m medal) animated() bool {
	if m.state != medalEarned {
		return false
	}
	return m.special != "" || m.tier == config.BadgePlatinum || m.tier == config.BadgeLegendary
}
