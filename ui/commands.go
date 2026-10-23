package ui

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// The command registry: every command a player can type at the prompt, with
// its aliases, subcommands, argument slots, help rows and a Dangerous flag.
// It is the one list the rest of the game reads:
//
//   - completion (suggest.go): the ghost text, Tab and the Enter rules;
//   - the Help panel (overlay_help.go): sections, usage rows, panels, shortcuts;
//   - the smoke docsync scenario: registry against site/docs/commands.md,
//     both ways;
//   - the smoke fuzz scenario: its command corpus.
//
// HandleCommand (input.go) keeps its own switch; TestRegistryMatchesDispatcher
// holds the two together in both directions. To add a command, add it here and
// to HandleCommand, then document it in site/docs/commands.md.

// ArgKind is what an argument slot takes. Completion offers values of the
// kind from the live game state; the Enter rules check a typed value against
// every value the kind can ever take.
type ArgKind int

const (
	ArgText                ArgKind = iota // free text: a name, a path, a code
	ArgNumber                             // a count or an amount (or one of Arg.Words)
	ArgWord                               // one of Arg.Words and nothing else
	ArgBuilding                           // a building buildable now (build)
	ArgPlanBuilding                       // a building the plan takes: this age's, then the next age's
	ArgPlanUpgradeBuilding                // a building the plan can upgrade now
	ArgSellBuilding                       // a built building sell takes: no wonder or storage
	ArgWorkerBuilding                     // a built building that takes workers (assign)
	ArgStaffedBuilding                    // a building with workers in it (unassign, dismiss)
	ArgUpgradeBuilding                    // a building with an upgrade available
	ArgTech                               // a tech available to research now
	ArgPlanTech                           // a tech the plan takes
	ArgResource                           // an unlocked resource
	ArgWonderResource                     // a resource the current wonder's bank still needs
	ArgFaction                            // a discovered civilization
	ArgTheme                              // an unlocked theme
	ArgSave                               // an existing save
	ArgRouteAvailable                     // a trade route that can start
	ArgRouteActive                        // a running trade route
	ArgPrestigeUpgrade                    // a legacy kit item not bought yet
	ArgAccount                            // a local account's name
	ArgPlanItem                           // a plan item's number
	ArgDeal                               // a trade deal's number, of the civilization in the previous argument
	ArgDomain                             // a worker domain (roster)
	ArgPercent                            // a percent: a number, with or without a % sign
)

// Arg is one argument slot.
type Arg struct {
	Kind     ArgKind
	Words    []string // literal words the slot also takes ("max", "all")
	Optional bool
}

// Usage is one help row: the form a player types and what it does.
type Usage struct{ Form, Text string }

// Command is a command or a subcommand.
type Command struct {
	Name    string
	Aliases []string
	Subs    []*Command
	Args    []Arg
	// BareOK: complete with nothing after it, even though it has
	// subcommands or a required argument (bare `plan` opens the panel).
	BareOK bool
	// Dangerous: irreversible. Enter never runs it from a completion guess;
	// it fills the field and waits for a second Enter. A dangerous command
	// with arguments only counts as run once an argument is given (bare
	// `sell` just prints its usage).
	Dangerous bool
	// Section is the Help panel section a top-level command is listed in;
	// "" lists it only under Panels (when Panel is set) or not at all.
	Section string
	// Panel, when set, lists the command under the Help panel's Panels, with
	// this description.
	Panel string
	// Dashboard: run by the dashboard itself, not HandleCommand (quit).
	Dashboard bool
	// Quiet: an old form kept working for habit, not shown: it is parsed
	// and run, but Help does not list it, the completer does not offer it,
	// and it is not among the forms the wiki has to document.
	Quiet bool
	Help  []Usage
}

// Help panel sections, in display order.
const (
	secActions  = "Actions"
	secPlan     = "Build Plan"
	secPeople   = "People"
	secResearch = "Research, Expeditions & Army"
	secTrade    = "Trade & Diplomacy"
	secWonders  = "Wonders & Prestige"
	secGame     = "Game"
	secAccounts = "Accounts"
)

// helpSections is the Help panel's section order, each one's heading and its
// note. The heading is the text of the section's name (a command's Section is
// the name itself, which only the code reads).
var helpSections = []struct {
	name        string
	title, note vibe.Entry
}{
	{secActions, vibe.E("ui.commands.secActions.01", "Actions"), vibe.Entry{}},
	{secPlan, vibe.E("ui.commands.secPlan.01", "Build Plan"), vibe.E("ui.commands.helpSections.01", "Queue builds and techs; each starts, and is paid for, when the resources are there, even while you are away.")},
	{secPeople, vibe.E("ui.commands.secPeople.01", "People"), vibe.E("ui.commands.helpSections.02", "The roster used to be called worker shares: workers share still works.")},
	{secResearch, vibe.E("ui.commands.secResearch.01", "Research, Expeditions & Army"), vibe.Entry{}},
	{secTrade, vibe.E("ui.commands.secTrade.01", "Trade & Diplomacy"), vibe.Entry{}},
	{secWonders, vibe.E("ui.commands.secWonders.01", "Wonders & Prestige"), vibe.Entry{}},
	{secGame, vibe.E("ui.commands.secGame.01", "Game"), vibe.Entry{}},
	{secAccounts, vibe.E("ui.commands.secAccounts.01", "Accounts"), vibe.E("ui.commands.helpSections.03", "Each account is its own slot. New and wipe live in the Accounts panel (main menu).")},
}

var (
	optCount    = Arg{Kind: ArgNumber, Optional: true}
	optCountMax = Arg{Kind: ArgNumber, Words: []string{"max"}, Optional: true}
	optCountAll = Arg{Kind: ArgNumber, Words: []string{"all"}, Optional: true}
	civArg      = []Arg{{Kind: ArgFaction}}
	// mapStyleArg is the one slot of map style and its shortcut style.
	mapStyleArg = Arg{Kind: ArgWord, Words: []string{"roguelike", "skyline"}, Optional: true}
)

// sub builds a subcommand with one help row.
func sub(name, form, text string, args ...Arg) *Command {
	return &Command{Name: name, Args: args, Help: []Usage{{form, text}}}
}

// panel builds a command that only opens a panel.
func panel(name, desc string, aliases ...string) *Command {
	return &Command{Name: name, Aliases: aliases, Panel: desc}
}

// registry is the player command list, in Help panel order within each
// section. Built fresh per call: callers may not modify it.
func registry() []*Command {
	confirmYes := func(form, text string) *Command {
		return &Command{Name: "confirm", BareOK: true, Subs: []*Command{{Name: "yes", Dangerous: true}}, Help: []Usage{{form, text}}}
	}
	prestigeConfirm := confirmYes("prestige confirm yes", vibe.Line("ui.help.commands.prestige_confirm_yes", "Prestige: a new run, keeping your points, carry-overs and Era Mastery"))
	festivalConfirm := confirmYes("festival confirm yes", vibe.Line("ui.help.commands.festival_confirm_yes", "Hold the festival now"))

	return []*Command{
		// Actions
		{Name: "gather", Aliases: []string{"g"}, Section: secActions,
			Args: []Arg{{Kind: ArgWord, Words: []string{"food", "wood", "stone"}}, optCount},
			Help: []Usage{{"gather <food|wood|stone> [amount]", vibe.Line("ui.help.commands.gather", "Gather food, wood or stone by hand (default ") + strconv.Itoa(int(gatherDefaultYield)) +
				vibe.Line("ui.help.commands.gather.2", ", max ") + strconv.Itoa(int(gatherMaxYield)) + vibe.Line("ui.help.commands.gather.3", " per use; the early ages only)")}}},
		{Name: "build", Aliases: []string{"b"}, Section: secActions,
			Args: []Arg{{Kind: ArgBuilding, Optional: true}, optCountMax},
			Help: []Usage{{"build <building> [count|max]", vibe.Line("ui.help.commands.build", "Build copies of a building (default 1)")}}},
		{Name: "sell", Section: secActions, Dangerous: true,
			Args: []Arg{{Kind: ArgSellBuilding}, optCount},
			Help: []Usage{{"sell <building> [count]", vibe.Line("ui.help.commands.sell", "Demolish copies of a building and get back 50% of the build cost (not wonders, storage or knowledge buildings)")}}},
		{Name: "advance", Section: secActions,
			Help: []Usage{{"advance", vibe.Line("ui.help.commands.advance", "Advance to the next age (when ready)")}}},
		{Name: "charter", Section: secActions, Panel: vibe.Line("ui.help.panels.charter", "The Advance panel: the four pillars the next age asks for, the march of ages and what the next age brings"),
			Help: []Usage{{"charter", vibe.Line("ui.help.commands.charter", "Open the Advance panel (Enter there advances when all four pillars stand)")}}},
		{Name: "upgrade", Section: secActions,
			Args: []Arg{{Kind: ArgUpgradeBuilding, Optional: true}, optCountAll},
			Help: []Usage{
				{"upgrade", vibe.Line("ui.help.commands.upgrade", "List the older buildings you hold and what upgrading them costs")},
				{"upgrade <building> [count|all]", vibe.Line("ui.help.commands.upgrade.2", "Upgrade copies of an older building to this age's building at half the price of a new copy (no count: as many as you can pay for)")},
			}},

		// Build Plan
		{Name: "plan", Section: secPlan, BareOK: true, Panel: vibe.Line("ui.help.panels.plan", "Build plan: queued builds & techs, started as resources come in"),
			Help: []Usage{{"plan", vibe.Line("ui.help.commands.plan", "Open the Plan panel (reorder and remove with keys)")}},
			Subs: []*Command{
				sub("build", "plan build <building> [count]", vibe.Line("ui.help.commands.plan.build", "Add copies of a building (this age's, or the next age's to build after you advance)"), Arg{Kind: ArgPlanBuilding}, optCount),
				sub("upgrade", "plan upgrade <building> [count|all]", vibe.Line("ui.help.commands.plan.upgrade", "Plan upgrades for held copies, one at a time as each can be paid"), Arg{Kind: ArgPlanUpgradeBuilding}, optCountAll),
				{Name: "research", Aliases: []string{"res"}, Args: []Arg{{Kind: ArgPlanTech}},
					Help: []Usage{{"plan research <tech>", vibe.Line("ui.help.commands.plan.research", "Add a tech and, before it, the techs it still needs (techs start one at a time, in order)")}}},
				sub("advance", "plan advance", vibe.Line("ui.help.commands.plan.advance", "Advance as soon as the next age is ready")),
				sub("deal", "plan deal <civ> <n>", vibe.Line("ui.help.commands.plan.deal", "Take a civilization's trade deal n once its price is there"),
					Arg{Kind: ArgFaction}, Arg{Kind: ArgDeal}),
				sub("list", "plan list", vibe.Line("ui.help.commands.plan.list", "Print the plan with each item's status")),
				{Name: "remove", Aliases: []string{"rm"}, Args: []Arg{{Kind: ArgPlanItem}},
					Help: []Usage{{"plan remove <n>", vibe.Line("ui.help.commands.plan.remove", "Remove item n")}}},
				{Name: "optimize", Aliases: []string{"reorder"},
					Help: []Usage{{"plan optimize", vibe.Line("ui.help.commands.plan.optimize", "Reorder the planned buildings for the fastest growth (homes and producers mixed), and say how much sooner the plan finishes")}}},
				sub("up", "plan up <n>", vibe.Line("ui.help.commands.plan.up", "Move item n one place up"), Arg{Kind: ArgPlanItem}),
				sub("down", "plan down <n>", vibe.Line("ui.help.commands.plan.down", "Move item n one place down"), Arg{Kind: ArgPlanItem}),
				{Name: "clear", Dangerous: true, Help: []Usage{{"plan clear", vibe.Line("ui.help.commands.plan.clear", "Empty the plan")}}},
			}},

		// Workers
		{Name: "recruit", Aliases: []string{"r"}, Section: secPeople,
			Args: []Arg{optCountMax},
			Help: []Usage{{"recruit [count|max]", vibe.Line("ui.help.commands.recruit", "Recruit workers into free housing (default 1). They start idle.")}}},
		{Name: "assign", Aliases: []string{"a"}, Section: secPeople,
			Args: []Arg{{Kind: ArgWorkerBuilding}, optCountAll},
			Help: []Usage{{"assign <building> [count|all]", vibe.Line("ui.help.commands.assign", "Put idle workers to work in a building (default 1)")}}},
		{Name: "unassign", Aliases: []string{"u"}, Section: secPeople,
			Args: []Arg{{Kind: ArgStaffedBuilding}, optCountAll},
			Help: []Usage{{"unassign <building> [count|all]", vibe.Line("ui.help.commands.unassign", "Take workers out of a building; they go idle (default 1)")}}},
		{Name: "dismiss", Section: secPeople, Dangerous: true,
			Args: []Arg{{Kind: ArgStaffedBuilding}, optCountAll},
			Help: []Usage{{"dismiss <building> [count|all]", vibe.Line("ui.help.commands.dismiss", "Dismiss workers from a building; they leave your population (default 1)")}}},
		{Name: "roster", Section: secPeople, BareOK: true,
			Args: []Arg{{Kind: ArgDomain, Words: []string{"auto"}}, {Kind: ArgPercent, Words: []string{"auto"}, Optional: true}},
			Help: []Usage{
				{"roster", vibe.Line("ui.help.commands.roster", "Show your roster: each domain's part of the workforce")},
				{"roster <domain> [percent|auto]", vibe.Line("ui.help.commands.roster.2", "Set a domain's share of your workers (0 keeps it empty; auto, the default, follows its buildings' slots); no percent shows it")},
				{"roster auto", vibe.Line("ui.help.commands.roster.3", "Put every domain back on auto")},
			}},
		{Name: "people", Section: secPeople, BareOK: true, Panel: vibe.Line("ui.help.panels.people", "People: workers, the reserve, soldiers and scholars"),
			Help: []Usage{{"people", vibe.Line("ui.help.commands.people", "Open the People panel: your workers and their work areas, the reserve, your soldiers and your scholars")}}},
		{Name: "workers", Section: secPeople, BareOK: true,
			Help: []Usage{{"workers", vibe.Line("ui.help.commands.workers", "Open the People panel (the same as people)")}},
			Subs: []*Command{
				// The roster's old name: `workers share ...` is `roster ...`.
				{Name: "share", BareOK: true, Quiet: true,
					Args: []Arg{{Kind: ArgDomain, Words: []string{"auto"}}, {Kind: ArgPercent, Words: []string{"auto"}, Optional: true}}},
				{Name: "auto-recruit", Aliases: []string{"autorecruit"},
					Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
					Help: []Usage{{"workers auto-recruit [on|off]", vibe.Line("ui.help.commands.workers.auto-recruit", "Show or set whether the game recruits into empty worker slots as housing and food allow (on by default)")}}},
			}},

		// Research, Expeditions & Army
		{Name: "research", Aliases: []string{"res"}, Section: secResearch, Panel: vibe.Line("ui.help.panels.research", "The tech tree: a map of every tech in sight"),
			Args: []Arg{{Kind: ArgTech, Optional: true}},
			Help: []Usage{{"research <tech>", vibe.Line("ui.help.commands.research", "Research a tech")}},
			Subs: []*Command{
				{Name: "cancel", Dangerous: true, Help: []Usage{{"research cancel", vibe.Line("ui.help.commands.research.cancel", "Cancel current research (the work done on it is lost)")}}},
				sub("list", "research list", vibe.Line("ui.help.commands.research.list", "List available techs")),
				sub("tree", "research tree [close|far]", vibe.Line("ui.help.commands.research.tree", "Open the tech tree, zoomed in on big badges (close) or out on the whole tree (far)"), Arg{Kind: ArgWord, Words: []string{"close", "far"}, Optional: true}),
				sub("card", "research card <tech>", vibe.Line("ui.help.commands.research.card", "Open the tech tree on a tech's card: what it does, how long it takes and what it needs"), Arg{Kind: ArgTech}),
			}},
		{Name: "scholars", Section: secResearch, BareOK: true,
			Args: []Arg{{Kind: ArgPercent, Optional: true}},
			Help: []Usage{
				{"scholars", vibe.Line("ui.help.commands.scholars", "Show your scholars: their kinds, how many research and teach, and what teaching adds")},
				{"scholars <percent>", vibe.Line("ui.help.commands.scholars.2", "Set the share of your scholars that research (100 by default); the rest teach")},
			}},
		panel("techs", ""),
		{Name: "expedition", Aliases: []string{"exp"}, Section: secResearch, Panel: vibe.Line("ui.help.panels.expedition", "The Expeditions panel (Ready Room and Expedition Report)"),
			Args: []Arg{{Kind: ArgText, Optional: true}},
			Help: []Usage{
				{"expedition", vibe.Line("ui.help.commands.expedition", "Open the Expeditions panel: the map of the age's lands with its offers on it, and the report of the parties out")},
				{"expedition <name>", vibe.Line("ui.help.commands.expedition.2", "Send the offer with that name (a part of it will do); it takes a party of soldiers, and your army sets how many are out")},
			},
			Subs: []*Command{sub("list", "expedition list", vibe.Line("ui.help.commands.expedition.list", "Print the mission board"))}},
		{Name: "army", Section: secResearch, Panel: vibe.Line("ui.help.panels.army", "Army overview & military campaigns"),
			Help: []Usage{{"army", vibe.Line("ui.help.commands.army", "Open the Army (military) panel")}}},
		{Name: "campaign", Section: secResearch,
			Args: []Arg{{Kind: ArgText, Optional: true}},
			Help: []Usage{{"campaign <name>", vibe.Line("ui.help.commands.campaign", "Same as expedition: send the offer with that name")}},
			Subs: []*Command{sub("list", "campaign list", vibe.Line("ui.help.commands.campaign.list", "Print the mission board"))}},

		// Trade & Diplomacy
		{Name: "trade", Aliases: []string{"t"}, Section: secTrade, BareOK: true, Panel: vibe.Line("ui.help.panels.trade", "Trade routes"),
			Subs: []*Command{
				{Name: "route", BareOK: true, Subs: []*Command{
					sub("list", "trade route list", vibe.Line("ui.help.commands.trade.route.list", "List trade routes")),
					sub("start", "trade route start <route>", vibe.Line("ui.help.commands.trade.route.start", "Start a trade route"), Arg{Kind: ArgRouteAvailable}),
					sub("stop", "trade route stop <route>", vibe.Line("ui.help.commands.trade.route.stop", "Stop a trade route"), Arg{Kind: ArgRouteActive}),
				}},
				sub("black", "trade black [resource]", vibe.Line("ui.help.commands.trade.black", "Same as blackmarket"), Arg{Kind: ArgResource, Optional: true}),
			}},
		{Name: "blackmarket", Aliases: []string{"bm"}, Section: secTrade,
			Args: []Arg{{Kind: ArgResource, Optional: true}},
			Help: []Usage{{"blackmarket [resource]", vibe.Line("ui.help.commands.blackmarket", "Smuggling run: gamble culture on a haul of one resource (bare blackmarket shows the odds, or when it opens)")}}},
		{Name: "factions", Section: secTrade, Panel: vibe.Line("ui.help.panels.factions", "Boons, the Geographic Society & opinion of each civilization (alias: diplomacy)"),
			Help: []Usage{{"factions", vibe.Line("ui.help.commands.factions", "Open the Factions panel (boons, the Geographic Society, opinion)")}}},
		{Name: "diplomacy", Aliases: []string{"dip"}, Section: secTrade, BareOK: true,
			Help: []Usage{{"diplomacy", vibe.Line("ui.help.commands.diplomacy", "Alias for factions (opens the same panel)")}},
			Subs: []*Command{
				sub("ally", "diplomacy ally <civ>", vibe.Line("ui.help.commands.diplomacy.ally", "Ally with a civilization (")+game.Amount(game.AllyCost, "gold")+vibe.Line("ui.help.commands.diplomacy.ally.2", " or less, needs opinion ")+strconv.Itoa(game.AllyOpinion)+")", civArg...),
				sub("rival", "diplomacy rival <civ>", vibe.Line("ui.help.commands.diplomacy.rival", "Declare a civilization your rival"), civArg...),
				sub("embargo", "diplomacy embargo <civ>", vibe.Line("ui.help.commands.diplomacy.embargo", "Embargo a civilization (a provocation: it can start a war)"), civArg...),
				sub("gift", "diplomacy gift <civ>", vibe.Line("ui.help.commands.diplomacy.gift", "Send a gift of gold for +")+strconv.Itoa(game.GiftOpinion)+vibe.Line("ui.help.commands.diplomacy.gift.2", " opinion or more (the Factions panel shows today's price and what it earns)"), civArg...),
				sub("neutral", "diplomacy neutral <civ>", vibe.Line("ui.help.commands.diplomacy.neutral", "Return to neutral with a civilization"), civArg...),
				sub("tribute", "diplomacy tribute <civ>", vibe.Line("ui.help.commands.diplomacy.tribute", "Sue for peace with a civilization at war"), civArg...),
				sub("deals", "diplomacy deals [civ]", vibe.Line("ui.help.commands.diplomacy.deals", "List trade deals (one civilization, or every one you have met)"),
					Arg{Kind: ArgFaction, Optional: true}),
				sub("accept", "diplomacy accept <civ> <n>", vibe.Line("ui.help.commands.diplomacy.accept", "Take a civilization's trade deal n"),
					Arg{Kind: ArgFaction}, Arg{Kind: ArgDeal}),
				{Name: "raid", Dangerous: true, Args: civArg,
					Help: []Usage{{"diplomacy raid <civ>", vibe.Line("ui.help.commands.diplomacy.raid", "Raid a civilization's trade route (a war provocation)")}}},
			}},

		// Wonders & Prestige
		{Name: "wonder", Section: secWonders, BareOK: true,
			Help: []Usage{{"wonder", vibe.Line("ui.help.commands.wonder", "Show current wonder bank status")}},
			Subs: []*Command{
				{Name: "collect", Aliases: []string{"bank"},
					Args: []Arg{{Kind: ArgWonderResource, Words: []string{"all"}}, {Kind: ArgNumber, Words: []string{"all", "max"}, Optional: true}},
					Help: []Usage{{"wonder collect <resource|all> [amount|all|max]", vibe.Line("ui.help.commands.wonder.collect", "Bank resources into the current wonder (alias: bank; no amount: as much as it needs)")}}},
			}},
		panel("wonders", vibe.Line("ui.help.panels.wonders", "Wonder bank, built wonders and a gallery of them (G)")),
		{Name: "prestige", Section: secWonders, BareOK: true,
			Help: []Usage{{"prestige", vibe.Line("ui.help.commands.prestige", "View prestige status")}},
			Subs: []*Command{
				prestigeConfirm,
			}},
		{Name: "festival", Section: secWonders, BareOK: true,
			Help: []Usage{{"festival", vibe.Line("ui.help.commands.festival", "Spend culture for a temporary production boost")}},
			Subs: []*Command{festivalConfirm}},
		{Name: "catastrophe", Aliases: []string{"cat"}, Section: secWonders,
			Help: []Usage{{"catastrophe", vibe.Line("ui.help.commands.catastrophe", "Reopen a pending catastrophe or Last Passage (or show the outlook)")}}},
		{Name: "harbinger", Aliases: []string{"harb"}, Section: secWonders, BareOK: true,
			Panel: vibe.Line("ui.help.panels.harbinger", "The harbinger's warning & your answers (alias: harb)"),
			Help: []Usage{
				{"harbinger", vibe.Line("ui.help.commands.harbinger", "Open the Harbinger panel (alias: harb)")},
				{"harbinger appease|brace|invite", vibe.Line("ui.help.commands.harbinger.2", "Answer the harbinger without the panel")},
			},
			Subs: []*Command{{Name: "appease"}, {Name: "brace"}, {Name: "invite", Dangerous: true}}},

		// Game
		{Name: "rates", Section: secGame, Help: []Usage{{"rates", vibe.Line("ui.help.commands.rates", "Show resource rate breakdown")}}},
		{Name: "status", Aliases: []string{"s"}, Section: secGame, Help: []Usage{{"status", vibe.Line("ui.help.commands.status", "Show detailed status")}}},
		{Name: "theme", Section: secGame, Panel: vibe.Line("ui.help.panels.theme", "Theme picker: palettes & accessibility"),
			Args: []Arg{{Kind: ArgTheme, Optional: true}},
			Help: []Usage{
				{"theme", vibe.Line("ui.help.commands.theme", "Open the theme picker (palettes + accessibility)")},
				{"theme <key>", vibe.Line("ui.help.commands.theme.2", "Switch to a theme by key")},
			},
			Subs: []*Command{sub("list", "theme list", vibe.Line("ui.help.commands.theme.list", "List themes with unlock status"))}},
		{Name: "map", Aliases: []string{"citymap", "worldmap"}, Section: secGame, BareOK: true,
			Panel: vibe.Line("ui.help.panels.map", "The map: your settlement and the known world (aliases: citymap, worldmap)"),
			Help: []Usage{
				{"map", vibe.Line("ui.help.commands.map", "Open the Map panel (worldmap opens it on the known world); the prompt keeps working while it is open")},
			},
			Subs: []*Command{
				sub("style", "map style [roguelike|skyline]", vibe.Line("ui.help.commands.map.style", "Show or set the map style (default roguelike)"), mapStyleArg),
				sub("glyphs", "map glyphs [ascii|unicode|nerd]", vibe.Line("ui.help.commands.map.glyphs", "Show or set the map's glyphs (default unicode; nerd needs a Nerd Font)"),
					Arg{Kind: ArgWord, Words: []string{"ascii", "unicode", "nerd"}, Optional: true}),
				sub("flows", "map flows [on|off]", vibe.Line("ui.help.commands.map.flows", "Turn the Map's flows overlay on or off (bare: switch it)"),
					Arg{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}),
			}},
		{Name: "style", Section: secGame, Args: []Arg{mapStyleArg},
			Help: []Usage{{"style [roguelike|skyline]", vibe.Line("ui.help.commands.style", "Same as map style")}}},
		{Name: "minimap", Section: secGame, Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
			Help: []Usage{{"minimap [on|off]", vibe.Line("ui.help.commands.minimap", "Show or set the mini map above the Buildings list (default off)")}}},
		{Name: "motion", Section: secGame, Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
			Help: []Usage{{"motion [on|off]", vibe.Line("ui.help.commands.motion", "Show or set motion: whether the maps, the badge case and theme effects move (default on)")}}},
		{Name: "vibemode", Section: secGame, Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
			Help: []Usage{{"vibemode [on|off]", vibe.Line("ui.help.commands.vibemode", "Show or set VibeMode: the game's words in a second, GenZ script (default off)")}}},
		{Name: "icons", Section: secGame,
			Help: []Usage{{"icons", vibe.Line("ui.help.commands.icons", "Check whether your font shows Nerd Font icons, and install one if it doesn't")}}},
		{Name: "save", Section: secGame, Args: []Arg{{Kind: ArgText, Optional: true}},
			Help: []Usage{{"save [name]", vibe.Line("ui.help.commands.save", "Save (no name: overwrite or branch; a name: branch a new save)")}},
			Subs: []*Command{sub("list", "save list", vibe.Line("ui.help.commands.save.list", "Same as saves"))}},
		{Name: "load", Section: secGame, Dangerous: true, Args: []Arg{{Kind: ArgSave, Optional: true}},
			Help: []Usage{{"load [name]", vibe.Line("ui.help.commands.load", "Load a save (no name: open the save browser)")}}},
		{Name: "saves", Section: secGame, Help: []Usage{{"saves", vibe.Line("ui.help.commands.saves", "List all save files")}}},
		{Name: "dump", Aliases: []string{"exportlogs"}, Section: secGame,
			Help: []Usage{{"dump", vibe.Line("ui.help.commands.dump", "Export logs to file for debugging")}}},
		{Name: "help", Aliases: []string{"h", "?"}, Section: secGame, Panel: vibe.Line("ui.help.panels.help", "This Help panel"),
			Help: []Usage{{"help", vibe.Line("ui.help.commands.help", "Open this Help panel")}}},
		{Name: "quit", Section: secGame, Dashboard: true, Dangerous: true,
			Help: []Usage{{"quit", vibe.Line("ui.help.commands.quit", "Save and quit the game")}}},

		// Accounts
		{Name: "account", Aliases: []string{"acct"}, Section: secAccounts, BareOK: true,
			Help: []Usage{{"account", vibe.Line("ui.help.commands.account", "Show this account's ID & backup help")}},
			Subs: []*Command{
				sub("list", "account list", vibe.Line("ui.help.commands.account.list", "List your local accounts")),
				sub("badges", "account badges", vibe.Line("ui.help.commands.account.badges", "List this account's badges, earned and locked")),
				{Name: "show", Subs: []*Command{sub("code", "account show code", vibe.Line("ui.help.commands.account.show.code", "Show this account's recovery code"))}},
				{Name: "switch", Dangerous: true, Args: []Arg{{Kind: ArgAccount}},
					Help: []Usage{{"account switch <name>", vibe.Line("ui.help.commands.account.switch", "Switch to an existing local account")}}},
				sub("export", "account export [path]", vibe.Line("ui.help.commands.account.export", "Back up this account's progress to a file"), Arg{Kind: ArgText, Optional: true}),
				sub("backup", "account backup", vibe.Line("ui.help.commands.account.backup", "Full snapshot (account files + saves) to data/backups/")),
				{Name: "import", Dangerous: true, Args: []Arg{{Kind: ArgText}, {Kind: ArgWord, Words: []string{"replace"}, Optional: true}},
					Help: []Usage{{"account import <path> [replace]", vibe.Line("ui.help.commands.account.import", "Restore an account from a backup file")}}},
				{Name: "recover", Dangerous: true, Args: []Arg{{Kind: ArgText}, {Kind: ArgWord, Words: []string{"confirm"}, Optional: true}},
					Help: []Usage{{"account recover <code> [confirm]", vibe.Line("ui.help.commands.account.recover", "Restore your account ID from a recovery code")}}},
				{Name: "wipe", Dangerous: true,
					Help: []Usage{{"account wipe", vibe.Line("ui.help.commands.account.wipe", "Where to wipe an account (the Accounts panel)")}}},
			}},

		{Name: "badges", Aliases: []string{"achievements"}, Section: secAccounts, BareOK: true,
			Panel: vibe.Line("ui.help.panels.badges", "The badge case: every badge your account holds or can earn (alias: achievements)"),
			Args:  []Arg{{Kind: ArgText, Words: []string{"next", "all"}, Optional: true}},
			Help: []Usage{
				{"badges", vibe.Line("ui.help.commands.badges", "Open the badge case (alias: achievements); the prompt keeps working while it is open")},
				{"badges <family>", vibe.Line("ui.help.commands.badges.2", "Open it on a family's tab: ages, lineages, ladders, specials and the rest (all: every family)")},
				{"badges next", vibe.Line("ui.help.commands.badges.3", "Open it on the badges you are closest to")},
				{"badges <name>", vibe.Line("ui.help.commands.badges.4", "Open a badge's detail by its name, or a part of it")},
			}},
		{Name: "title", Section: secAccounts, BareOK: true,
			Args: []Arg{{Kind: ArgText, Words: []string{"default"}, Optional: true}},
			Help: []Usage{
				{"title", vibe.Line("ui.help.commands.title", "List the titles your account holds, and the one it wears")},
				{"title <name>", vibe.Line("ui.help.commands.title.2", "Wear a title you hold (default: the one your badge score gives)")},
			}},

		// Panels only
		panel("milestones", vibe.Line("ui.help.panels.milestones", "Milestone goals & rewards"), "ms"),
		panel("stats", vibe.Line("ui.help.panels.stats", "Empire statistics")),
		panel("logs", vibe.Line("ui.help.panels.logs", "Recent game log entries")),
		panel("epoch", vibe.Line("ui.help.panels.epoch", "Epoch progress & catastrophe")),
		panel("history", vibe.Line("ui.help.panels.history", "Civilization history timeline")),
		panel("buildings", vibe.Line("ui.help.panels.buildings", "Built structures by lineage")),
	}
}

// panelOrder is the Help panel's Panels list order (the sidebar's, then the rest).
var panelOrder = []string{
	"milestones", "badges", "research", "plan", "charter", "expedition", "army", "trade", "factions", "stats", "wonders",
	"people", "logs", "epoch", "harbinger", "history", "buildings", "map", "theme", "help",
}

// devCommand is one dev-console command. They are not player commands: the
// dashboard sends a /line straight to game.DevConsoleCommand, and they are
// listed and completed only while dev mode is active.
type devCommand struct{ name, form, text string }

var devCommands = []devCommand{
	{"/god", "/god", "Toggle godmode: free costs, instant builds"},
	{"/fill", "/fill", "Add a large amount of every resource"},
	{"/give", "/give <resource> <amount>", "Add an amount of a resource"},
	{"/build", "/build <building_key>", "Instantly place one building"},
	{"/techs", "/techs", "Unlock all techs up to the current age"},
	{"/age", "/age <age_key>", "Jump to any age"},
	{"/ages", "/ages", "List all age keys"},
	{"/prestige", "/prestige <level 0-9>", "Set prestige level"},
	{"/speed", "/speed <multiplier>", "Set the tick-speed multiplier"},
	{"/catastrophe", "/catastrophe", "Force the current epoch's catastrophe (Iron Era on)"},
	{"/harbinger", "/harbinger", "Fate a doom now and bring its harbinger (a false prophet in the Stone Era)"},
	{"/lastpassage", "/lastpassage", "Make the Last Passage pending (final epoch)"},
	{"/mastery", "/mastery <age|all> <0-10>", "Set Era Mastery for one age or every age"},
	{"/record", "/record <age_key>", "Set the record (the deepest age ever entered)"},
}

// CommandInfo is a registry entry as the smoke suite sees it.
type CommandInfo struct {
	Names []string // the name, then its aliases
	// Subs are the subcommand paths under the command, one or two words deep
	// ("route", "route start"), including literal argument words ("max",
	// "confirm yes"), primary names only.
	Subs []string
	// Accepts is every word the command takes right after its name: its
	// subcommands and their aliases, and its first slot's literal words.
	Accepts   map[string]bool
	Dangerous bool
}

// Commands lists the player commands from the registry, for the smoke
// suite's docsync and fuzz scenarios.
func Commands() []CommandInfo {
	var out []CommandInfo
	for _, c := range registry() {
		info := CommandInfo{Names: append([]string{c.Name}, c.Aliases...), Accepts: map[string]bool{}, Dangerous: c.Dangerous}
		for _, w := range nextWords(c, true) {
			info.Accepts[w] = true
		}
		for _, w := range nextWords(c, false) {
			info.Subs = append(info.Subs, w)
			if s := lookup(c.Subs, w); s != nil {
				for _, w2 := range nextWords(s, false) {
					info.Subs = append(info.Subs, w+" "+w2)
				}
			}
		}
		out = append(out, info)
	}
	return out
}

// nextWords is the literal words that can follow c: its subcommands (with
// aliases when aliases is set) and its first slot's words.
func nextWords(c *Command, aliases bool) []string {
	var out []string
	for _, s := range c.Subs {
		if s.Quiet && !aliases {
			continue // an old form: accepted like an alias, not a form of its own
		}
		out = append(out, s.Name)
		if aliases {
			out = append(out, s.Aliases...)
		}
	}
	if len(c.Args) > 0 {
		out = append(out, c.Args[0].Words...)
	}
	return out
}

// lookup finds the command named w (or aliased w) in cs.
func lookup(cs []*Command, w string) *Command {
	w = strings.ToLower(w)
	for _, c := range cs {
		if c.Name == w {
			return c
		}
		for _, a := range c.Aliases {
			if a == w {
				return c
			}
		}
	}
	return nil
}

// parsed is a command line read against the registry.
type parsed struct {
	path []*Command // the command, then each subcommand
	args []string   // the words after the last of them
}

func (p parsed) last() *Command { return p.path[len(p.path)-1] }

// parse reads words against cmds: the command, then subcommands for as
// long as the words name them; the rest are arguments. ok is false when the
// first word is no command.
func parse(cmds []*Command, words []string) (parsed, bool) {
	if len(words) == 0 {
		return parsed{}, false
	}
	c := lookup(cmds, words[0])
	if c == nil {
		return parsed{}, false
	}
	p := parsed{path: []*Command{c}}
	i := 1
	for ; i < len(words); i++ {
		s := lookup(p.last().Subs, words[i])
		if s == nil {
			break
		}
		p.path = append(p.path, s)
	}
	p.args = words[i:]
	return p, true
}

// complete reports whether p is a whole command: every word fits a slot
// (valid decides the ones that name game things) and every required slot
// is filled.
func (p parsed) complete(valid func(Arg, string) bool) bool {
	c := p.last()
	if len(p.args) == 0 {
		return c.BareOK || len(c.Args) == 0 && len(c.Subs) == 0 || len(c.Args) > 0 && c.Args[0].Optional
	}
	if len(p.args) > len(c.Args) {
		return false
	}
	for i, w := range p.args {
		if !slotTakes(c.Args[i], w, valid) {
			return false
		}
	}
	for _, a := range c.Args[len(p.args):] {
		if !a.Optional {
			return false
		}
	}
	return true
}

// dangerous reports whether running p would do something irreversible.
func (p parsed) dangerous() bool {
	for i, c := range p.path {
		if !c.Dangerous {
			continue
		}
		if i < len(p.path)-1 || len(c.Args) == 0 || len(p.args) > 0 {
			return true
		}
	}
	return false
}

// slotTakes reports whether the slot a takes the word w.
func slotTakes(a Arg, w string, valid func(Arg, string) bool) bool {
	for _, x := range a.Words {
		if strings.EqualFold(x, w) {
			return true
		}
	}
	switch a.Kind {
	case ArgText:
		return true
	case ArgWord:
		return false
	case ArgNumber, ArgPlanItem, ArgDeal:
		_, err := strconv.ParseFloat(w, 64)
		return err == nil
	case ArgPercent:
		_, err := strconv.ParseFloat(strings.TrimSuffix(w, "%"), 64)
		return err == nil
	}
	return valid(a, w)
}
