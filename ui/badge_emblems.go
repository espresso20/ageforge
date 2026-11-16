package ui

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// badge_emblems.go is the badge case's emblem table: the names a badge's
// emblem may take (config.BadgeDef.Emblem) and the sign, if any, the badge's
// coin wears for each. TestBadgeEmblemsAreKnown holds the catalog to it.

// badgeEmblem is an emblem: the sign the badge's coin wears, if it has one.
type badgeEmblem struct {
	special string
}

// emblemStar is the emblem of a badge that names none, and of the ladders
// that belong to no table: no sign.
var emblemStar = badgeEmblem{}

// badgeEmblems are the named emblems.
var badgeEmblems = map[string]badgeEmblem{
	"star":  emblemStar,
	"hut":   {},
	"sun":   {},
	"trade": {},
	// What the families wear.
	"wonder":    {},
	"hall":      {},
	"housing":   {},
	"knowledge": {},
	"faith":     {},
	"culture":   {},
	"monument":  {},
	"engineer":  {},
	"military":  {},
	"war":       {},
	"diplomacy": {},
	"civ":       {},
	"harbinger": {},
	"hazard":    {},
	"ruin":      {},
	"alien":     {},
	// The resources: a store, or the symbol of what makes one.
	"store":  {},
	"food":   {},
	"wood":   {},
	"mine":   {},
	"metal":  {},
	"gold":   {},
	"energy": {},
	"data":   {},
	// The integrity badges: a store for the jar, the scholars' mark for
	// the ledger, the hackers' for the source. Each has a sprite.
	"cookie_jar": {special: specialCookieJar},
	"ledger":     {special: specialLedger},
	"source":     {special: specialSource},
	// The legendary badges' signs (medalIcons, medal_pix.go).
	"sprite.chains":    {special: specialChains},
	"sprite.boxes":     {special: specialBoxes},
	"sprite.museum":    {special: specialMuseum},
	"sprite.wonders":   {special: specialWonders},
	"sprite.rogues":    {special: specialRogues},
	"sprite.endings":   {special: specialEndings},
	"sprite.undying":   {special: specialUndying},
	"sprite.transcend": {special: specialTranscend},
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
				return badgeEmblem{}, true
			}
		}
		return emblemStar, false
	case strings.HasPrefix(name, emblemCentre):
		n, err := strconv.Atoi(strings.TrimPrefix(name, emblemCentre))
		if err != nil || n < 0 {
			return emblemStar, false
		}
		return badgeEmblem{}, true
	}
	if e, ok := badgeEmblems[name]; ok {
		return e, true
	}
	return emblemStar, false
}

// medalState is whether a badge is earned, in sight but not earned, or
// withheld by the spoiler rules.
type medalState uint8

const (
	medalEarned medalState = iota
	medalLocked
	medalHidden
)

// medal is what the renderer needs to know of a badge.
type medal struct {
	// tier is the badge's tier; BadgeNoTier is an integrity badge.
	tier  config.BadgeTier
	state medalState
	// crossed: earned in a modified game.
	crossed bool
	// special names a sign of medalIcons (medal_pix.go), "" for none.
	special string
}

// medalOf is a badge as the renderer draws it.
func medalOf(v game.BadgeView) medal {
	e, _ := emblemOf(v.Emblem)
	m := medal{tier: v.Level, crossed: v.Crossed, special: e.special}
	switch {
	case v.Hidden:
		m.state = medalHidden
	case !v.Earned:
		m.state = medalLocked
	}
	return m
}

// animated reports whether a badge's coin moves: platinum glints, the
// legendary ring turns and the signs of the specials glitch. Bronze,
// silver and gold hold still.
func (m medal) animated() bool {
	if m.state != medalEarned {
		return false
	}
	return m.special != "" || m.tier == config.BadgePlatinum || m.tier == config.BadgeLegendary
}
