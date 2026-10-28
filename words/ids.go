// Package words holds what the copy tools and VibeMode share: the ids of the
// text a player reads, and the VibeMode script itself.
//
// Every line of text in the game has a row id (building.lumber_mill.name,
// tech.pottery.description). cmd/words reads the Go source to export the rows to
// sheets, and the ruleset overlay reads the same ids off the live definitions to
// swap a row's text for its VibeMode line. The table both use is here: FieldIDs
// says how a definition's text field becomes an id, so the two cannot name a row
// differently.
package words

import (
	"regexp"
	"strings"
)

// FieldIDs maps a struct field that holds player text ("package.Type.Field")
// to the template of its row id. A template may use:
//
//	{Field}   a field of the same struct literal
//	{^Field}  a field of the struct literal around it
//	{k}       the map key the text sits under ({k2}: the key above that)
//	{n}       its place in a list, from 01
//	{sn}      the struct literal's place in its list, from 01
//	{decl}    the variable or function it is declared in (the source only)
//	{aN}      the Nth argument of the call around it (the source only)
//
// A template with {decl} or {aN} can be filled from the source but not from a
// running ruleset, so the overlay skips those rows. The export still has them.
var FieldIDs = map[string]string{
	// ----- ages and eras -----
	"config.AgeDef.Name":              "age.{Key}.name",
	"config.AgeDef.Description":       "age.{Key}.description",
	"config.AgeDef.Quip":              "age.{Key}.quip",
	"config.EpochDef.Name":            "era.{Key}.name",
	"config.EpochDef.Description":     "era.{Key}.description",
	"config.ResourceDef.Name":         "resource.{Key}.name",
	"config.WorkerClassDef.ClassName": "worker_class.{Domain}.{AgeKey}",

	// ----- buildings and wonders -----
	"config.BuildingDef.Name":        "building.{Key}.name",
	"config.BuildingDef.Description": "building.{Key}.description",
	"config.BuildingDef.Flavor":      "building.{Key}.flavor",

	// ----- techs -----
	"config.TechDef.Name":         "tech.{Key}.name",
	"config.TechDef.Code":         "tech.{Key}.code",
	"config.TechDef.Description":  "tech.{Key}.description",
	"config.TechLaneDef.Name":     "tech_lane.{Key}.name",
	"config.MechanicDef.Name":     "mechanic.{Key}.name",
	"config.MechanicDef.Text":     "mechanic.{Key}.text",
	"config.FeatureLockDef.Name":  "feature.{Key}.name",
	"config.FeatureLockDef.Needs": "feature.{Key}.needs",
	"config.FeatureLockDef.Then":  "feature.{Key}.then",
	"config.FeatureLockDef.Opens": "feature.{Key}.opens",

	// ----- milestones -----
	"config.MilestoneDef.Name":        "milestone.{Key}.name",
	"config.MilestoneDef.Description": "milestone.{Key}.description",
	"config.MilestoneDef.Flavor":      "milestone.{Key}.flavor",
	"config.MilestoneChainDef.Name":   "milestone_chain.{Key}.name",
	"config.MilestoneChainDef.Title":  "milestone_chain.{Key}.title",
	"config.MilestoneChainDef.Flavor": "milestone_chain.{Key}.flavor",
	"config.TitleDef.Title":           "milestone_title.{sn}",

	// ----- badges -----
	"config.BadgeDef.Name":                   "badge.{Key}.name",
	"config.BadgeDef.Desc":                   "badge.{Key}.description",
	"config.BadgeDef.Hint":                   "badge.{Key}.hint",
	"config.BadgeDef.Ladder":                 "badge.{Key}.ladder",
	"config.BadgeFamilyDef.Name":             "badge_family.{Key}.name",
	"config.BadgeFamilyDef.Desc":             "badge_family.{Key}.description",
	"config.BadgeFamilyDef.Names":            "badge.{k}.name",
	"config.BadgeFamilyDef.Descs":            "badge.{k}.description",
	"config.BadgeFamilyDef.Ladder":           "badge_family.{Key}.ladder",
	"config.BadgeFamilyDef.Hint":             "badge_family.{Key}.hint",
	"config.BadgeRung.Name":                  "badge_rung.{decl}.{sn}",
	"config.BadgeReward.Title":               "badge_reward.{k}{^Key}.title",
	"config.BadgeTitleDef.Title":             "badge_title.{Badge}{Set}",
	"config.BadgeScoreTitle.Title":           "badge_score_title.{sn}",
	"config.BadgeExpeditionDef.Name":         "badge_expedition.{Key}.name",
	"config.MissionWords.Tiers":              "mission.{Age}.tier.{n}",
	"config.MissionWords.Places":             "mission.{Age}.place.{n}",
	"config.MissionWords.ShortObjectives":    "mission.{Age}.short.{n}",
	"config.MissionWords.LongObjectives":     "mission.{Age}.long.{n}",
	"config.MissionWords.MilitaryObjectives": "mission.{Age}.military.{n}",
	"config.MissionWords.CivilObjectives":    "mission.{Age}.civil.{n}",
	"config.MissionWords.Twists":             "mission.{Age}.twist.{n}",
	"config.MissionWords.Weather":            "mission.{Age}.weather.{n}",
	"config.MissionWords.Foes":               "mission.{Age}.foes.{n}",
	"config.MissionWords.Helpers":            "mission.{Age}.helpers.{n}",
	"config.MissionWords.Finds":              "mission.{Age}.finds.{n}",
	"config.BadgeThemeDef.Name":              "badge_theme.{Key}.name",

	// ----- events, catastrophes and harbingers -----
	"config.EventDef.Name":             "event.{Key}.name",
	"config.EventDef.LogMessage":       "event.{Key}.log",
	"config.EpochEventDef.Name":        "era_event.{Key}.name",
	"config.EpochEventDef.FlavorText":  "era_event.{Key}.text",
	"config.AwakeningDef.Name":         "awakening.{Key}.name",
	"config.AwakeningDef.FlavorText":   "awakening.{Key}.text",
	"config.HarbingerDef.Name":         "harbinger.{Key}.name",
	"config.HarbingerDef.Description":  "harbinger.{Key}.description",
	"config.HarbingerDef.AppeaseLabel": "harbinger.{Key}.appease",
	"config.HarbingerDef.BraceLabel":   "harbinger.{Key}.brace",
	"config.HarbingerDef.InviteLabel":  "harbinger.{Key}.invite",

	// ----- trade, civilizations, prestige, boons -----
	"config.TradeRouteDef.Name":             "route.{Key}.name",
	"config.TradeRouteDef.Description":      "route.{Key}.description",
	"config.FactionDef.Name":                "civ.{Key}.name",
	"config.FactionDef.Backstory":           "civ.{Key}.backstory",
	"config.PrestigeUpgradeDef.Name":        "prestige.{Key}.name",
	"config.PrestigeUpgradeDef.Description": "prestige.{Key}.description",
	"boon.Def.Name":                         "boon.{decl}.{sn}.name",
	"boon.Def.Flavors":                      "boon.{decl}.{sn}.line.{n}",

	// ----- themes and the map -----
	"theme.Theme.Name":       "theme.{Key}.name",
	"theme.Theme.Blurb":      "theme.{Key}.blurb",
	"theme.Theme.UnlockHint": "theme.{Key}.unlock_hint",
}

// FieldID returns the id template for a field, or "" if it has none. typ is
// "package.Type" and field is the field's name.
func FieldID(typ, field string) string { return FieldIDs[typ+"."+field] }

// TemplateRe finds the slots of an id or "where" template.
var TemplateRe = regexp.MustCompile(`\{(name:)?(\^)?([A-Za-z]+[0-9]*)\}`)

// Slot answers one slot of a template. It gets the name between the braces and
// whether the slot was written {^name} (a field of the struct around). For a
// slot it does not know it returns known false, and Expand leaves the slot as
// written.
type Slot func(name string, outer bool) (value string, known bool)

// Expand fills the slots of t. For an id (forID) each value is made safe to be
// part of one (IDSafe). A {name:X} slot goes through called, which says what the
// game calls the key X. With no called it is left as written.
func Expand(t string, slot Slot, forID bool, called func(string) string) string {
	return TemplateRe.ReplaceAllStringFunc(t, func(m string) string {
		parts := TemplateRe.FindStringSubmatch(m)
		v, known := slot(parts[3], parts[2] == "^")
		if !known {
			return m
		}
		if parts[1] != "" {
			if called == nil {
				return m
			}
			return called(v)
		}
		if forID {
			return IDSafe(v)
		}
		return v
	})
}

// IDSafe makes a piece of an id out of free text. It keeps ASCII letters (in
// their case), digits and dashes. Each run of any other characters (spaces,
// underscores, punctuation, other letters) becomes one underscore between kept
// characters, and a run at the start or end is dropped.
func IDSafe(s string) string {
	var b strings.Builder
	gap := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r >= 'A' && r <= 'Z', r == '-':
			if gap && b.Len() > 0 {
				b.WriteByte('_')
			}
			gap = false
			b.WriteRune(r)
		default:
			gap = true
		}
	}
	return b.String()
}
