package mapmodel

import "github.com/espresso20/ageforge/theme"

// Class is a glyph's colour class. The maps colour everything through these
// few classes; each one is a theme role leaning a set amount toward an
// identity hue, clamped legible (ui/mapstyle.Palette resolves them). A
// theme switch is free and light themes need no re-keying.
type Class uint8

const (
	CNone    Class = iota
	CGround        // grass stipple, beach
	CFlora         // forest, fields
	CWater         // rivers, sea
	CRock          // mountains, ruins
	CHill          // hills: quieter than mountains
	CRoad          // streets, trails
	CWall          // walls, towers, gates
	CHouse         // housing
	CWork          // production (the epoch hue)
	CCivic         // knowledge, faith, culture, monuments
	CWealth        // trade, storage, wonders
	CMil           // military
	CLife          // workers, caravans, scouts
	CIdle          // idle workers, understaffing: a bottleneck signal
	CDanger        // war, catastrophe, fire
	CMemory        // fogged but remembered terrain
	CStar          // sky dots
	CLabel         // names on the grid
	CLegacy        // superseded buildings (the old town)
	CAlly          // relation: allied
	CFriend        // relation: friendly or trading
	CNeutral       // relation: neutral
	CRival         // relation: rival
	CWar           // relation: at war or embargo
	CFresh         // new since your last visit
	CText          // chrome body text
	CDim           // chrome hints
	CAccent        // chrome titles
	NumClasses
)

// ClassSpec is how a class is derived: start from Role, lean toward Hue by
// Mix, then clamp to at least MinContrast against the background.
// EpochHue makes Hue the epoch's production hue instead.
type ClassSpec struct {
	Role        theme.Role
	Hue         theme.MapHue
	Mix         float64
	MinContrast float64
	EpochHue    bool
}

// ClassSpecs is the class table.
var ClassSpecs = [NumClasses]ClassSpec{
	CNone:    cs(theme.RoleText, theme.HueNone, 0, 1),
	CGround:  cs(theme.RoleDim, theme.HueGrass, 0.5, 1.6),
	CFlora:   cs(theme.RolePositive, theme.HueForest, 0.55, 3.0),
	CWater:   cs(theme.RoleLabel, theme.HueWater, 0.7, 3.0),
	CRock:    cs(theme.RoleDim, theme.HueRock, 0.4, 2.8),
	CHill:    cs(theme.RoleDim, theme.HueHill, 0.5, 1.9),
	CRoad:    ce(theme.RoleDim, theme.HueNone, 0.25, 2.6),
	CWall:    cs(theme.RoleText, theme.HueStone, 0.35, 4.0),
	CHouse:   ce(theme.RoleText, theme.HueNone, 0.30, 4.5),
	CWork:    ce(theme.RoleHighlight, theme.HueNone, 0.65, 4.0),
	CCivic:   cs(theme.RoleAccent, theme.HueCivic, 0.25, 4.0),
	CWealth:  cs(theme.RoleHighlight, theme.HueGold, 0.45, 4.0),
	CMil:     cs(theme.RoleNegative, theme.HueMilitary, 0.3, 4.0),
	CLife:    cs(theme.RoleBright, theme.HueNone, 0, 7),
	CIdle:    cs(theme.RoleWarning, theme.HueIdle, 0.3, 4.5),
	CDanger:  cs(theme.RoleNegative, theme.HueDanger, 0.4, 4.5),
	CMemory:  cs(theme.RoleDim, theme.HueNone, 0, 1.3),
	CStar:    cs(theme.RoleText, theme.HueNone, 0, 1.8),
	CLabel:   cs(theme.RoleLabel, theme.HueNone, 0, 4.5),
	CLegacy:  ce(theme.RoleDim, theme.HueNone, 0.2, 2.4),
	CAlly:    cs(theme.RolePositive, theme.HueAlly, 0.6, 4.0),
	CFriend:  cs(theme.RoleHighlight, theme.HueFriend, 0.6, 4.0),
	CNeutral: cs(theme.RoleLabel, theme.HueNeutral, 0.6, 4.0),
	CRival:   cs(theme.RoleWarning, theme.HueRival, 0.6, 4.0),
	CWar:     cs(theme.RoleNegative, theme.HueWar, 0.6, 4.5),
	CFresh:   cs(theme.RolePositive, theme.HueFresh, 0.3, 4.5),
	CText:    cs(theme.RoleText, theme.HueNone, 0, 4.5),
	CDim:     cs(theme.RoleDim, theme.HueNone, 0, 3.0),
	CAccent:  cs(theme.RoleAccent, theme.HueNone, 0, 4.5),
}

func cs(r theme.Role, h theme.MapHue, mix, min float64) ClassSpec {
	return ClassSpec{Role: r, Hue: h, Mix: mix, MinContrast: min}
}

func ce(r theme.Role, h theme.MapHue, mix, min float64) ClassSpec {
	return ClassSpec{Role: r, Hue: h, Mix: mix, MinContrast: min, EpochHue: true}
}

// LineageClass is the colour class of a lineage's buildings.
func LineageClass(lin string) Class {
	switch lin {
	case LinHousing:
		return CHouse
	case LinFood:
		return CFlora
	case LinKnowledge, LinFaith, LinCulture, LinMonument, LinDiplomacy:
		return CCivic
	case LinTrade, LinStorage, LinWonder:
		return CWealth
	case LinMilitary:
		return CMil
	}
	return CWork
}

// RelationClass is the class of a civ's standing: the fixed signal colours
// both styles use (war red, ally green, friendly gold, neutral steel,
// rival amber).
func RelationClass(r Relation) Class {
	switch r {
	case RelAllied:
		return CAlly
	case RelFriendly:
		return CFriend
	case RelRival:
		return CRival
	case RelWar, RelEmbargo:
		return CWar
	}
	return CNeutral
}
