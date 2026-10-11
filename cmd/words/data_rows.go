package main

import (
	"fmt"
	"regexp"
	"strings"
)

// data_rows.go reads the data definitions: the struct literals in config,
// boon, theme and mapmodel that describe buildings, techs, badges and the
// rest. Each string field of those structs is either text a player reads
// (fieldRules: its id, its kind, where it shows) or not (hiddenFields: a key,
// a reference, a number in words). TestConfigFieldsAreAllClassified fails
// on a config field in neither table.

// How often a player meets a line, for the order of first-hours.csv.
const (
	seenAlways  = 1 // on screen all the time: the main window, the menu
	seenOften   = 2 // names of the things a player builds and researches
	seenRegular = 3 // descriptions, cards, panels a player opens
	seenNews    = 4 // log lines and refusals
	seenRare    = 5 // one-off stories: an age's quip, an event
	seenPool    = 6 // one line of many in a pool
)

// fieldRule says how a struct field becomes a row.
type fieldRule struct {
	id    string // template for the row's id
	kind  string
	area  *area
	where string // template for the where sentence
	age   string // template giving a key the age can be read from
	seen  int
	max   int
	rules ruleSet
	fix   func(c *catalog, u *unit, r *row)
}

// Templates may use: {Field} a field of the same struct literal; {^Field} a
// field of the struct literal around it; {k} the map key the text sits
// under; {n} its place in a list; {sn} the struct literal's place in its
// list; {aN} the Nth argument of the call around it; {decl} the variable or
// function it is declared in; {name:X} what the game calls the key X.

var fieldRules = map[string]fieldRule{
	// ----- ages and eras -----
	"config.AgeDef.Name": {id: "age.{Key}.name", kind: kindName, area: areaAges, age: "{Key}", seen: seenAlways,
		where: "The name of an age: in the header of the main window, on the splash when you reach it, and wherever the game names it. Several screens cut a closing “ Age” off the name, so keep that ending."},
	"config.AgeDef.Description": {id: "age.{Key}.description", kind: kindVoice, area: areaAges, age: "{Key}", seen: seenRare,
		where: "The one-line description of the {Name}, in quotes on the splash when you reach it."},
	"config.AgeDef.Quip": {id: "age.{Key}.quip", kind: kindVoice, area: areaAges, age: "{Key}", seen: seenRare,
		where: "The quip under the description on the splash when you reach the {Name}."},
	"config.EpochDef.Name": {id: "era.{Key}.name", kind: kindName, area: areaAges, age: "{Key}", seen: seenOften,
		where: "The name of an era (a group of ages): in the Epoch panel, in the log when the era begins, and in badges."},
	"config.EpochDef.Description": {id: "era.{Key}.description", kind: kindVoice, area: areaAges, age: "{Key}", seen: seenRare,
		where: "What the log says after “The {Name} begins.”, and the era's line in the Epoch panel."},
	"config.ResourceDef.Name": {id: "resource.{Key}.name", kind: kindName, area: areaAges, age: "res:{Key}", seen: seenAlways,
		where: "The name of a resource: in the Resources box of the main window, in every cost and in the log."},
	"config.ResourceDef.Description": {id: "resource.{Key}.description", kind: kindVoice, area: areaAges, age: "res:{Key}", seen: seenRegular,
		where: "The one-line description of the resource {Name}."},
	"config.WorkerClassDef.ClassName": {id: "worker_class.{Domain}.{AgeKey}", kind: kindName, area: areaAges, age: "{AgeKey}", seen: seenRegular,
		where: "What the Workers panel calls the workers of the {Domain} domain from the {name:AgeKey} on."},

	// ----- buildings and wonders -----
	"config.BuildingDef.Name": {id: "building.{Key}.name", kind: kindName, area: areaBuildings, age: "{Key}", seen: seenOften, fix: wonderArea,
		where: "The name of a building: in the Buildings list, on the map, in costs and in the log."},
	"config.BuildingDef.Description": {id: "building.{Key}.description", kind: kindVoice, area: areaBuildings, age: "{Key}", seen: seenRegular, fix: wonderArea,
		where: "The description of {Name} in the Buildings list and the Buildings panel. The game writes what the building does after it."},
	"config.BuildingDef.Flavor": {id: "building.{Key}.flavor", kind: kindVoice, area: areaBuildings, age: "{Key}", seen: seenRegular, fix: wonderArea,
		where: "The line in italics under the description of {Name} in the Buildings list and the Buildings panel."},

	// ----- techs -----
	"config.TechDef.Name": {id: "tech.{Key}.name", kind: kindName, area: areaTechs, age: "tech:{Key}", seen: seenOften,
		where: "The name of a tech: on its card in the Research panel, in the research list and in the log."},
	"config.TechDef.Code": {id: "tech.{Key}.code", kind: kindLabel, area: areaTechs, age: "tech:{Key}", seen: seenRegular, max: 5,
		where: "The short code in capitals on the card of {Name} in the Research panel. At most five capital letters, and no two techs may share one."},
	"config.TechDef.Description": {id: "tech.{Key}.description", kind: kindVoice, area: areaTechs, age: "tech:{Key}", seen: seenRegular,
		where: "The description on the card of {Name} in the Research panel."},
	"config.TechLaneDef.Name": {id: "tech_lane.{Key}.name", kind: kindLabel, area: areaTechs, age: "primitive_age", seen: seenRegular,
		where: "The name of a lane of the tech tree, written down the side of the Research panel."},
	"config.MechanicDef.Name": {id: "mechanic.{Key}.name", kind: kindLabel, area: areaTechs, seen: seenRegular,
		where: "What the game calls a number a tech can change, in the lists of what a tech does."},
	"config.MechanicDef.Text": {id: "mechanic.{Key}.text", kind: kindMessage, area: areaTechs, seen: seenRegular,
		where: "How a tech's card says it changes “{Name}”. The %s is the size of the change (“15%”)."},
	"config.FeatureLockDef.Name": {id: "feature.{Key}.name", kind: kindMessage, area: areaTechs, age: "feature:{Key}", seen: seenNews,
		where: "The start of the refusal when a command waits for a tech: “{Name} {Needs} <tech> first. Research it to {Then}.” It also starts the log line when the tech is done: “{Name} {Opens}.”"},
	"config.FeatureLockDef.Needs": {id: "feature.{Key}.needs", kind: kindMessage, area: areaTechs, age: "feature:{Key}", seen: seenNews,
		where: "The verb of the refusal “{Name} {Needs} <tech> first. Research it to {Then}.” It must agree with “{Name}”: need after a plural, needs after a singular."},
	"config.FeatureLockDef.Then": {id: "feature.{Key}.then", kind: kindMessage, area: areaTechs, age: "feature:{Key}", seen: seenNews,
		where: "The end of the refusal “{Name} {Needs} <tech> first. Research it to {Then}.”"},
	"config.FeatureLockDef.Opens": {id: "feature.{Key}.opens", kind: kindMessage, area: areaTechs, age: "feature:{Key}", seen: seenRegular,
		where: "What a tech's card says it unlocks, after the word “It” (“It {Opens}.”), and the log line when the tech is done."},

	// ----- milestones -----
	"config.MilestoneDef.Name": {id: "milestone.{Key}.name", kind: kindName, area: areaMilestones, age: "milestone:{Key}", seen: seenRegular,
		where: "The name of a milestone: in the Milestones panel and in the log line and toast when it is completed."},
	"config.MilestoneDef.Description": {id: "milestone.{Key}.description", kind: kindMessage, area: areaMilestones, age: "milestone:{Key}", seen: seenRegular,
		where: "What the Milestones panel says you must do for {Name}."},
	"config.MilestoneDef.Flavor": {id: "milestone.{Key}.flavor", kind: kindVoice, area: areaMilestones, age: "milestone:{Key}", seen: seenRare, rules: rulesNoMark,
		where: "The quip the log adds in gray when the milestone {Name} is completed."},
	"config.MilestoneChainDef.Name": {id: "milestone_chain.{Key}.name", kind: kindName, area: areaMilestones, seen: seenRegular,
		where: "The name of a chain of milestones, in the Milestones panel and in the log when the chain is completed."},
	"config.MilestoneChainDef.Title": {id: "milestone_chain.{Key}.title", kind: kindName, area: areaMilestones, seen: seenRare,
		where: "The title the chain {Name} gives when it is completed."},
	"config.MilestoneChainDef.Flavor": {id: "milestone_chain.{Key}.flavor", kind: kindVoice, area: areaMilestones, seen: seenRare, rules: rulesNoMark,
		where: "The quip the log adds in gray when the chain {Name} is completed."},
	"config.TitleDef.Title": {id: "milestone_title.{sn}", kind: kindName, area: areaMilestones, seen: seenRare,
		where: "A title a run earns by the number of milestones completed."},

	// ----- badges -----
	"config.BadgeDef.Name": {id: "badge.{Key}.name", kind: kindName, area: areaBadges, age: "{InAge}", seen: seenRegular,
		where: "The name of a badge: in the badge case, and in the toast and log line when it is earned."},
	"config.BadgeDef.Desc": {id: "badge.{Key}.description", kind: kindVoice, area: areaBadges, age: "{InAge}", seen: seenRegular,
		where: "What the badge case says the badge {Name} is for."},
	"config.BadgeDef.Hint": {id: "badge.{Key}.hint", kind: kindMessage, area: areaBadges, age: "{InAge}", seen: seenRegular,
		where: "The hint the badge case shows in place of the description while the badge {Name} is still secret."},
	"config.BadgeDef.Ladder": {id: "badge.{Key}.ladder", kind: kindLabel, area: areaBadges, seen: seenRegular,
		where: "The heading the badge case puts over the ladder the badge {Name} belongs to."},
	"config.BadgeFamilyDef.Name": {id: "badge_family.{Key}.name", kind: kindName, area: areaBadges, seen: seenRegular,
		where: "The name of each badge of the “{Family}” family. The game fills in the slot for each badge."},
	"config.BadgeFamilyDef.Desc": {id: "badge_family.{Key}.description", kind: kindVoice, area: areaBadges, seen: seenRegular,
		where: "What the badge case says each badge of the “{Family}” family is for. The game fills in the slots for each badge."},
	"config.BadgeFamilyDef.Names": {id: "badge.{k}.name", kind: kindName, area: areaBadges, age: "{k}", seen: seenRegular,
		where: "The name of the badge {k}, used in place of its family's usual name: in the badge case, and in the toast and log line when it is earned."},
	"config.BadgeFamilyDef.Descs": {id: "badge.{k}.description", kind: kindVoice, area: areaBadges, age: "{k}", seen: seenRegular,
		where: "What the badge case says the badge {k} is for, in place of its family's usual line."},
	"config.BadgeFamilyDef.Ladder": {id: "badge_family.{Key}.ladder", kind: kindLabel, area: areaBadges, seen: seenRegular,
		where: "The heading the badge case puts over each ladder of the “{Family}” family. The game fills in the slot."},
	"config.BadgeFamilyDef.Hint": {id: "badge_family.{Key}.hint", kind: kindMessage, area: areaBadges, seen: seenRegular,
		where: "The hint the badge case shows while a badge of the “{Family}” family is still secret."},
	"config.BadgeRung.Name": {id: "badge_rung.{decl}.{sn}", kind: kindName, area: areaBadges, seen: seenRegular,
		where: "The name of rung {sn} of a badge ladder ({decl}). A ladder's badges are named by their rung."},
	"config.BadgeReward.Title": {id: "badge_reward.{k}{^Key}.title", kind: kindName, area: areaBadges, seen: seenRare,
		where: "The title the badge {k}{^Key} gives. Keep it the same as the title's own row."},
	"config.BadgeTitleDef.Title": {id: "badge_title.{Badge}{Set}", kind: kindName, area: areaBadges, seen: seenRare,
		where: "A title an account can wear once it holds the badge or set {Badge}{Set}."},
	"config.BadgeScoreTitle.Title": {id: "badge_score_title.{sn}", kind: kindName, area: areaBadges, seen: seenRare,
		where: "A title an account earns by its badge points."},
	"config.BadgeExpeditionDef.Name": {id: "badge_expedition.{Key}.name", kind: kindName, area: areaBadges, age: "{MinAge}", seen: seenRegular,
		where: "The name of an expedition as the badge case writes it. Keep it the same as the expedition's own name."},
	"config.BadgeThemeDef.Name": {id: "badge_theme.{Key}.name", kind: kindName, area: areaBadges, seen: seenRegular,
		where: "The name of a theme as the badge case writes it. Keep it the same as the theme's own name."},

	// ----- events, catastrophes and harbingers -----
	"config.EventDef.Name": {id: "event.{Key}.name", kind: kindName, area: areaEvents, age: "{MinAge}", seen: seenNews,
		where: "The name of a random event: in the Events box of the main window while it lasts, and in the toast when it starts."},
	"config.EventDef.Description": {id: "event.{Key}.description", kind: kindVoice, area: areaEvents, age: "{MinAge}", seen: seenRare,
		where: "The description of the event {Name}, shown with it while it lasts."},
	"config.EventDef.LogMessage": {id: "event.{Key}.log", kind: kindVoice, area: areaEvents, age: "{MinAge}", seen: seenNews, rules: rulesNoMark,
		where: "What the log says when the event {Name} happens. It carries no numbers: the game writes what the event did after it (what was lost and gained, each rate and how long it lasts), so leave numbers out."},
	"config.EpochEventDef.Name": {id: "era_event.{Key}.name", kind: kindName, area: areaEvents, age: "iron_age", seen: seenRare,
		where: "The name of an era event: in the log when it happens, in the Epoch panel and on the age splash."},
	"config.EpochEventDef.FlavorText": {id: "era_event.{Key}.text", kind: kindVoice, area: areaEvents, age: "iron_age", seen: seenRare,
		where: "What the log says after the name when the era event {Name} happens. The numbers in it are checked against what the event does; {dur} is how long it lasts."},
	"config.AwakeningDef.Name": {id: "awakening.{Key}.name", kind: kindName, area: areaEvents, age: "{TriggerAge}", seen: seenRare,
		where: "The name of an awakening: a one-time discovery the log announces on reaching the {name:TriggerAge}."},
	"config.AwakeningDef.FlavorText": {id: "awakening.{Key}.text", kind: kindVoice, area: areaEvents, age: "{TriggerAge}", seen: seenRare,
		where: "What the log says after “Awakening: {Name}.” The numbers in it are checked against what the awakening gives."},
	"config.HarbingerDef.Name": {id: "harbinger.{Key}.name", kind: kindName, area: areaEvents, age: "{Age}", seen: seenRare,
		where: "The name of the harbinger of the {name:Age}: in the Harbinger panel, the log and the flavor lines about it. It starts in lower case because it is used in the middle of sentences."},
	"config.HarbingerDef.Description": {id: "harbinger.{Key}.description", kind: kindVoice, area: areaEvents, age: "{Age}", seen: seenRare,
		where: "The description of {Name} at the top of the Harbinger panel."},
	"config.HarbingerDef.AppeaseLabel": {id: "harbinger.{Key}.appease", kind: kindLabel, area: areaEvents, age: "{Age}", seen: seenRare,
		where: "What appeasing {Name} is called: after “Appease: ” in the Harbinger panel, and in the log when you do it."},
	"config.HarbingerDef.BraceLabel": {id: "harbinger.{Key}.brace", kind: kindLabel, area: areaEvents, age: "{Age}", seen: seenRare,
		where: "What bracing in the {name:Age} is called: after “Brace: ” in the Harbinger panel, and in the log when you do it."},
	"config.HarbingerDef.InviteLabel": {id: "harbinger.{Key}.invite", kind: kindLabel, area: areaEvents, age: "{Age}", seen: seenRare,
		where: "What inviting the doom is called with {Name}: after “Invite: ” in the Harbinger panel, and in the log when you do it."},

	// ----- trade, civilizations, prestige, boons -----
	"config.TradeRouteDef.Name": {id: "route.{Key}.name", kind: kindName, area: areaWorld, age: "{MinAge}", seen: seenRegular,
		where: "The name of a trade route, in the Trade panel and the log. The map picks how the route travels (land, sea or air) from words in its name and description, such as rail, ship or warp."},
	"config.TradeRouteDef.Description": {id: "route.{Key}.description", kind: kindVoice, area: areaWorld, age: "{MinAge}", seen: seenRegular,
		where: "The description of the route {Name} in the Trade panel. It must read “<what you send> for <what you get>”, naming each resource. The map picks how the route travels from words in it, such as rail, ship or warp."},
	"config.FactionDef.Name": {id: "civ.{Key}.name", kind: kindName, area: areaWorld, age: "{MinAge}", seen: seenRegular,
		where: "The name of another civilization: in the Factions panel, on the map and in the log."},
	"config.FactionDef.Backstory": {id: "civ.{Key}.backstory", kind: kindVoice, area: areaWorld, age: "{MinAge}", seen: seenRare,
		where: "The introduction of the {Name}: in the log when you first meet them, and under their name in the Factions panel."},
	"config.FactionDef.Description": {id: "civ.{Key}.description", kind: kindVoice, area: areaWorld, age: "{MinAge}", seen: seenRare,
		where: "The one-line description of the {Name}."},
	"config.PrestigeUpgradeDef.Name": {id: "prestige.{Key}.name", kind: kindName, area: areaWorld, age: "medieval_age", seen: seenRare,
		where: "The name of an item in the prestige shop."},
	"config.PrestigeUpgradeDef.Description": {id: "prestige.{Key}.description", kind: kindMessage, area: areaWorld, age: "medieval_age", seen: seenRare,
		where: "What the prestige shop says {Name} does, and the log when you buy it."},
	"boon.Def.Name": {id: "boon.{decl}.{sn}.name", kind: kindName, area: areaWorld, age: "bronze_age", seen: seenRare,
		where: "The name of a boon or setback another civilization can leave you with, in the log and the Factions panel."},
	"boon.Def.Flavors": {id: "boon.{decl}.{sn}.line.{n}", kind: kindVoice, area: areaWorld, age: "bronze_age", seen: seenPool,
		where: "One of the lines the log picks from when an expedition comes back with “{Name}”. The slots are filled with what it gives or takes."},

	// ----- themes and the map -----
	"theme.Theme.Name": {id: "theme.{Key}.name", kind: kindName, area: areaMap, age: "primitive_age", seen: seenRegular,
		where: "The name of a color theme, in the theme picker."},
	"theme.Theme.Blurb": {id: "theme.{Key}.blurb", kind: kindVoice, area: areaMap, age: "primitive_age", seen: seenRegular,
		where: "The description of the theme {Name} in the theme picker."},
	"theme.Theme.UnlockHint": {id: "theme.{Key}.unlock_hint", kind: kindMessage, area: areaMap, age: "primitive_age", seen: seenRegular,
		where: "What the theme picker says you must do to unlock the theme {Name}."},
	"mapmodel.cityFeatureDef.name": {id: "map.city.{key}.name", kind: kindName, area: areaMap, age: "{from}", seen: seenRare,
		where: "What the map calls a part of the late-game city in a sentence."},
	"mapmodel.cityFeatureDef.title": {id: "map.city.{key}.title", kind: kindLabel, area: areaMap, age: "{from}", seen: seenRare,
		where: "The heading when you inspect this part of the city on the map."},
	"mapmodel.cityFeatureDef.lines": {id: "map.city.{key}.line.{n}", kind: kindVoice, area: areaMap, age: "{from}", seen: seenRare,
		where: "A line the map shows when you inspect “{title}” in the city."},
	"mapmodel.moverDef.name": {id: "map.mover.{key}.name", kind: kindName, area: areaMap, age: "{from}", seen: seenRare,
		where: "What the map calls something that moves across it, in a sentence."},
	"mapmodel.moverDef.title": {id: "map.mover.{key}.title", kind: kindLabel, area: areaMap, age: "{from}", seen: seenRare,
		where: "The heading when you inspect this moving thing on the map."},
	"mapmodel.moverDef.lines": {id: "map.mover.{key}.line.{n}", kind: kindVoice, area: areaMap, age: "{from}", seen: seenRare,
		where: "A line the map shows when you inspect “{title}” as it moves."},
	"mapmodel.cityLookDef.name": {id: "map.city_look.{key}.name", kind: kindName, area: areaMap, age: "{key}", seen: seenRare,
		where: "What the map calls the city of the {name:key}."},
	"mapmodel.cityLookDef.structure": {id: "map.city_look.{key}.structure", kind: kindVoice, area: areaMap, age: "{key}", seen: seenRare,
		where: "What the map says the city of the {name:key} is built of."},
	"mapmodel.cityLookDef.ambient": {id: "map.city_look.{key}.ambient", kind: kindVoice, area: areaMap, age: "{key}", seen: seenRare,
		where: "What the map says moves in the city of the {name:key}."},
	"mapmodel.skyPartDef.name": {id: "map.sky.{k2}.{k}", kind: kindName, area: areaMap, age: "space_age", seen: seenRare,
		where: "What the map calls a part of your station or fleet in the space ages."},
	"mapmodel.AlienKind.Color": {id: "map.alien.{sn}.color", kind: kindLabel, area: areaMap, age: "galactic_age", seen: seenRare,
		where: "The color the map names a visiting craft by when you inspect it."},
	"mapmodel.AlienKind.Lines": {id: "map.alien.{sn}.line.{n}", kind: kindVoice, area: areaMap, age: "galactic_age", seen: seenRare,
		where: "A line the map shows when you inspect a visiting craft."},
}

// hiddenFields are struct fields no player reads as text, by "Type.Field" or
// by the field's bare name when it means the same in every struct.
var hiddenFields = map[string]string{
	"Key": "a key", "key": "a key", "Category": "a key", "RequiredAge": "an age key", "RequiredTech": "a tech key",
	"LineageKey": "a key", "WorkerDomain": "a key", "EpochKey": "an era key", "OutputResource": "a resource key",
	"MinAge": "an age key", "Age": "an age key", "AgeKey": "an age key", "TriggerAge": "an age key", "InAge": "an age key",
	"Type": "a key", "Target": "a key", "Family": "a key", "Subject": "a key", "Counter": "a key", "Event": "a key",
	"Pred": "a key", "Emblem": "an emblem's name or glyph", "Set": "a key", "Aliases": "keys older saves used",
	"Only": "keys", "Except": "keys", "Emblems": "emblem names", "Theme": "a theme key", "Badge": "a badge key",
	"Rule": "a key", "Why": "a note for developers", "Fact": "a key", "Lane": "a key", "Prerequisites": "tech keys",
	"AnyOf": "tech keys", "Icon": "a glyph", "Color": "a color name", "PrimaryResource": "a resource key", "BraceMaterials": "resource keys",
	"EnergyResource": "a resource key", "CatastropheKey": "a key", "Sentiment": "a key", "Tech": "a tech key",
	"EffectKey": "a key", "EffectType": "a key", "From": "a key", "To": "a key", "RequiredBld": "a building key",
	"Specialty": "a resource key", "Personality": "a key the code and the flavor catalogs match on",
	"Domain": "a key", "Hue": "a color", "Atom": "a key", "Measure": "a key", "Keys": "keys", "MilestoneKeys": "keys",
	"RequiredTechs": "tech keys", "UnlockBuildings": "keys", "UnlockResources": "keys", "UnlockVillagers": "keys",
	"Ages": "age keys", "Resource": "a resource key", "UnlockBadge": "a badge key", "Effect": "a key",
	"from": "an age key", "until": "an age key", "GainGlyph": "a glyph", "LossGlyph": "a glyph",

	"config.BadgeFamilyDef.Key":           "a template for badge keys",
	"config.BadgeDef.Aliases":             "keys older saves used",
	"config.BadgeReveal.Key":              "a key",
	"config.BadgeReward.Theme":            "a theme key",
	"config.AgeDef.BuildingReqs":          "building keys",
	"config.BuildingDef.BaseCost":         "resource keys",
	"config.MilestoneDef.MinResources":    "resource keys",
	"config.MilestoneDef.MinBuildings":    "building keys",
	"config.TradeRouteDef.Export":         "resource keys",
	"config.TradeRouteDef.Import":         "resource keys",
	"config.BadgeFamilyDef.Tiers":         "badge keys",
	"config.BadgeFamilyDef.Ladders":       "keys",
	"config.BadgeFamilyDef.Aliases":       "keys older saves used",
	"config.BadgeFamilyDef.Rewards":       "badge keys (the titles are read from BadgeReward)",
	"config.BadgeFamilyDef.Thresholds":    "keys",
	"config.unlockOrder.ages":             "not a definition",
	"config.unlockOrder.resources":        "not a definition",
	"config.TechEffect.Target":            "a key",
	"config.TechEffectKey.Target":         "a key",
	"config.BadgeHeldOutDef.Why":          "a note for developers",
	"config.BadgeProof.Why":               "a note for developers",
	"config.FeatureLockDef.Tech":          "a tech key",
	"config.BadgeExpeditionDef.MinAge":    "an age key",
	"config.BadgeThemeDef.Badge":          "a badge key",
	"config.BadgeTitleDef.Badge":          "a badge key",
	"config.BadgeTitleDef.Set":            "a key",
	"config.TechDef.Emblem":               "a glyph",
	"config.TechLaneDef.Emblem":           "a glyph",
	"config.BadgeDef.Emblem":              "an emblem's name",
	"config.BadgeFamilyDef.Emblem":        "an emblem's name",
	"config.BadgeDef.Family":              "a key",
	"config.BadgeFamilyDef.Family":        "a key",
	"config.MilestoneChainDef.Category":   "a key",
	"config.MilestoneDef.Category":        "a key",
	"config.BuildingDef.Category":         "a key",
	"config.EpochEventDef.Type":           "a key",
	"config.Effect.Type":                  "a key",
	"config.Effect.Target":                "a key",
	"config.BuildingSum.Keys":             "building keys",
	"config.BadgeFamilyDef.Only":          "keys",
	"config.BadgeFamilyDef.Except":        "keys",
	"config.BadgeFamilyDef.Emblems":       "emblem names",
	"config.BadgeFamilyDef.Event":         "a key",
	"config.BadgeFamilyDef.Counter":       "a key",
	"config.BadgeFamilyDef.InAge":         "an age key or a slot",
	"config.BadgeFamilyDef.Set":           "a key",
	"config.buildingMetaEntry.EpochKey":   "an era key",
	"config.buildingMetaEntry.LineageKey": "a key",
}

// wonderArea moves a wonder's rows to the wonders sheet.
func wonderArea(c *catalog, u *unit, r *row) {
	if c.w.wonder[u.st().sib["Key"]] {
		r.area = areaAges
		r.where = strings.Replace(r.where, "a building:", "a wonder:", 1)
		r.where = strings.Replace(r.where, "in the Buildings list, on the map, in costs and in the log", "in the Wonders panel, on the map and in the log", 1)
		r.where = strings.Replace(r.where, "in the Buildings list and the Buildings panel", "in the Wonders panel", 2)
	}
}

// readField makes the row for a unit that fills a struct field, or says why
// not. It returns nil and "" for a struct the tables do not know, which
// leaves the unit to the code reader.
func readField(c *catalog, u *unit) (*row, string) {
	s := u.st()
	if s.typ == "" || s.field == "" {
		return nil, ""
	}
	if len(u.calls) > 0 && u.calls[0].out == 0 {
		return nil, "" // an argument of a call inside the literal, not the field's own text
	}
	full := s.typ + "." + s.field
	rule, ok := fieldRules[full]
	if !ok {
		if why, ok := hiddenFields[full]; ok {
			return nil, why
		}
		if why, ok := hiddenFields[s.field]; ok {
			return nil, why
		}
		if strings.HasPrefix(s.typ, "config.") && u.f.dir == "config" {
			// A config field nobody has classified: put it in front of the
			// owner, marked, rather than lose it.
			if !hasWord(u.text) {
				return nil, whyNoWords
			}
			return &row{
				id:     fmt.Sprintf("unclassified.%s.%d_%d", idSafe(full), u.line, u.start),
				kind:   kindVoice,
				where:  "A field of a definition (" + full + ") the tool has not been told about.",
				unsure: true, age: -1, seen: seenRare, area: areaUnsure,
			}, ""
		}
		return nil, ""
	}
	if !hasWord(u.text) {
		return nil, whyNoWords
	}
	return c.fromRule(rule, u), ""
}

// fromRule fills a rule's templates for a unit.
func (c *catalog) fromRule(rule fieldRule, u *unit) *row {
	r := &row{
		id:    c.expand(rule.id, u, true),
		kind:  rule.kind,
		area:  rule.area,
		where: c.expand(rule.where, u, false),
		age:   -1,
		seen:  rule.seen,
		max:   rule.max,
		rules: rule.rules,
	}
	if rule.age != "" {
		r.age = c.w.ageOf(c.expand(rule.age, u, false))
	}
	switch u.f.dir {
	case "config":
		r.rules |= rulesConfig | rulesDocs
	case "boon":
		r.rules |= rulesUI
	}
	if rule.fix != nil {
		rule.fix(c, u, r)
	}
	if u.decl == "BadgesHeldOut" {
		r.unsure = true
		r.where += " This badge is held out of the game for now (BadgesHeldOut), so nobody can see it yet."
	}
	return r
}

var tmplRe = regexp.MustCompile(`\{(name:)?(\^)?([A-Za-z]+[0-9]*)\}`)

// expand fills a template's slots from the unit's surroundings. For an id,
// each value is made safe to be part of one.
func (c *catalog) expand(t string, u *unit, forID bool) string {
	return tmplRe.ReplaceAllStringFunc(t, func(m string) string {
		parts := tmplRe.FindStringSubmatch(m)
		v, known := u.lookup(parts[3], parts[2] == "^")
		if !known {
			return m
		}
		if parts[1] != "" {
			return c.w.called(v)
		}
		if forID {
			return idSafe(v)
		}
		return v
	})
}

// lookup resolves one template slot.
func (u *unit) lookup(name string, outer bool) (string, bool) {
	switch {
	case outer:
		if len(u.structs) > 1 {
			return u.structs[1].sib[name], true
		}
		return "", true
	case name == "k" || name == "k2":
		var keys []string
		for _, st := range u.coll {
			if st.isMap {
				keys = append(keys, st.key)
			}
		}
		for _, s := range u.structs {
			for _, st := range s.coll {
				if st.isMap {
					keys = append(keys, st.key)
				}
			}
		}
		if i := len(name) - 1; i < len(keys) {
			return keys[i], true
		}
		return "", true
	case name == "n":
		for _, st := range u.coll {
			if !st.isMap {
				return fmt.Sprintf("%02d", st.index+1), true
			}
		}
		return "01", true
	case name == "sn":
		if len(u.structs) > 0 {
			for _, st := range u.structs[0].coll {
				if !st.isMap {
					return fmt.Sprintf("%02d", st.index+1), true
				}
			}
		}
		return "01", true
	case name == "decl":
		return u.decl, true
	case len(name) == 2 && name[0] == 'a' && name[1] >= '0' && name[1] <= '9':
		if len(u.calls) > 0 {
			if i := int(name[1] - '0'); i < len(u.calls[0].args) {
				return u.calls[0].args[i], true
			}
		}
		return "", true
	}
	if len(u.structs) > 0 {
		if v, ok := u.structs[0].sib[name]; ok {
			return v, true
		}
		// A field name that the literal leaves out reads as empty; any other
		// word in braces is not a slot.
		if name[0] >= 'A' && name[0] <= 'Z' || name == "key" || name == "from" || name == "title" || name == "name" {
			return "", true
		}
	}
	return "", false
}
