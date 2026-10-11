package ui

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/game"
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
	ArgText            ArgKind = iota // free text: a name, a path, a code
	ArgNumber                         // a count or an amount (or one of Arg.Words)
	ArgWord                           // one of Arg.Words and nothing else
	ArgBuilding                       // a building buildable now (build)
	ArgPlanBuilding                   // a building the plan takes: this age's, then the next age's
	ArgSellBuilding                   // a built building sell takes: no wonder or storage
	ArgWorkerBuilding                 // a built building that takes workers (assign)
	ArgStaffedBuilding                // a building with workers in it (unassign, dismiss)
	ArgUpgradeBuilding                // a building with an upgrade available
	ArgTech                           // a tech available to research now
	ArgPlanTech                       // a tech the plan takes
	ArgResource                       // an unlocked resource
	ArgWonderResource                 // a resource the current wonder's bank still needs
	ArgTradeFrom                      // a resource the market buys from you
	ArgTradeTo                        // a resource the market sells for the previous argument
	ArgFaction                        // a discovered civilization
	ArgTheme                          // an unlocked theme
	ArgSave                           // an existing save
	ArgExpedition                     // a scouting expedition
	ArgCampaign                       // a military campaign
	ArgRouteAvailable                 // a trade route that can start
	ArgRouteActive                    // a running trade route
	ArgPrestigeUpgrade                // a legacy kit item not bought yet
	ArgAccount                        // a local account's name
	ArgPlanItem                       // a plan item's number
	ArgDeal                           // a trade deal's number, of the civilization in the previous argument
	ArgDomain                         // a worker domain (roster)
	ArgPercent                        // a percent: a number, with or without a % sign
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
	secWorkers  = "Workers"
	secResearch = "Research, Expeditions & Army"
	secTrade    = "Trade & Diplomacy"
	secWonders  = "Wonders & Prestige"
	secGame     = "Game"
	secAccounts = "Accounts"
)

// helpSections is the Help panel's section order, and each one's note.
var helpSections = []struct{ name, note string }{
	{secActions, ""},
	{secPlan, "Queue builds and techs; each starts, and is paid for, when the resources are there, even while you are away."},
	{secWorkers, "The roster used to be called worker shares: workers share still works."},
	{secResearch, ""},
	{secTrade, ""},
	{secWonders, ""},
	{secGame, ""},
	{secAccounts, "Each account is its own slot. New and wipe live in the Accounts panel (main menu)."},
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
	prestigeConfirm := confirmYes("prestige confirm yes", "Prestige: a new run, keeping your points, the legacy kit and Era Mastery")
	festivalConfirm := confirmYes("festival confirm yes", "Hold the festival now")

	return []*Command{
		// Actions
		{Name: "gather", Aliases: []string{"g"}, Section: secActions,
			Args: []Arg{{Kind: ArgWord, Words: []string{"food", "wood", "stone"}}, optCount},
			Help: []Usage{{"gather <food|wood|stone> [amount]", "Gather food, wood or stone by hand (default " + strconv.Itoa(int(gatherDefaultYield)) +
				", max " + strconv.Itoa(int(gatherMaxYield)) + " per use; the early ages only)"}}},
		{Name: "build", Aliases: []string{"b"}, Section: secActions,
			Args: []Arg{{Kind: ArgBuilding, Optional: true}, optCountMax},
			Help: []Usage{{"build <building> [count|max]", "Build copies of a building (default 1)"}}},
		{Name: "sell", Section: secActions, Dangerous: true,
			Args: []Arg{{Kind: ArgSellBuilding}, optCount},
			Help: []Usage{{"sell <building> [count]", "Demolish copies of a building and get back 50% of the build cost (not wonders or storage)"}}},
		{Name: "advance", Section: secActions,
			Help: []Usage{{"advance", "Advance to the next age (when ready)"}}},
		{Name: "upgrade", Section: secActions,
			Args: []Arg{{Kind: ArgUpgradeBuilding, Optional: true}, optCountAll},
			Help: []Usage{
				{"upgrade", "List the building upgrades you can make"},
				{"upgrade <building> [count|all]", "Upgrade copies to the next tier (no count: all of them; pays the difference in cost)"},
			}},

		// Build Plan
		{Name: "plan", Section: secPlan, BareOK: true, Panel: "Build plan: queued builds & techs, started as resources come in",
			Help: []Usage{{"plan", "Open the Plan panel (reorder and remove with keys)"}},
			Subs: []*Command{
				sub("build", "plan build <building> [count]", "Add copies of a building (this age's, or the next age's to build after you advance)", Arg{Kind: ArgPlanBuilding}, optCount),
				{Name: "research", Aliases: []string{"res"}, Args: []Arg{{Kind: ArgPlanTech}},
					Help: []Usage{{"plan research <tech>", "Add a tech and, before it, the techs it still needs (techs start one at a time, in order)"}}},
				sub("trade", "plan trade <give> <get> [amount]", "Sell <give> for <get> as it comes in, until <amount> <get> is bought (no amount: until you remove it)",
					Arg{Kind: ArgTradeFrom}, Arg{Kind: ArgTradeTo}, Arg{Kind: ArgNumber, Optional: true}),
				sub("advance", "plan advance", "Advance as soon as the next age is ready"),
				sub("deal", "plan deal <civ> <n>", "Take a civilization's trade deal n once its price is there",
					Arg{Kind: ArgFaction}, Arg{Kind: ArgDeal}),
				sub("list", "plan list", "Print the plan with each item's status"),
				{Name: "remove", Aliases: []string{"rm"}, Args: []Arg{{Kind: ArgPlanItem}},
					Help: []Usage{{"plan remove <n>", "Remove item n"}}},
				sub("up", "plan up <n>", "Move item n one place up", Arg{Kind: ArgPlanItem}),
				sub("down", "plan down <n>", "Move item n one place down", Arg{Kind: ArgPlanItem}),
				{Name: "clear", Dangerous: true, Help: []Usage{{"plan clear", "Empty the plan"}}},
			}},

		// Workers
		{Name: "recruit", Aliases: []string{"r"}, Section: secWorkers,
			Args: []Arg{optCountMax},
			Help: []Usage{{"recruit [count|max]", "Recruit workers into free housing (default 1). They start idle."}}},
		{Name: "assign", Aliases: []string{"a"}, Section: secWorkers,
			Args: []Arg{{Kind: ArgWorkerBuilding}, optCountAll},
			Help: []Usage{{"assign <building> [count|all]", "Put idle workers to work in a building (default 1)"}}},
		{Name: "unassign", Aliases: []string{"u"}, Section: secWorkers,
			Args: []Arg{{Kind: ArgStaffedBuilding}, optCountAll},
			Help: []Usage{{"unassign <building> [count|all]", "Take workers out of a building; they go idle (default 1)"}}},
		{Name: "dismiss", Section: secWorkers, Dangerous: true,
			Args: []Arg{{Kind: ArgStaffedBuilding}, optCountAll},
			Help: []Usage{{"dismiss <building> [count|all]", "Dismiss workers from a building; they leave your population (default 1)"}}},
		{Name: "roster", Section: secWorkers, BareOK: true,
			Args: []Arg{{Kind: ArgDomain, Words: []string{"auto"}}, {Kind: ArgPercent, Words: []string{"auto"}, Optional: true}},
			Help: []Usage{
				{"roster", "Show your roster: each domain's part of the workforce"},
				{"roster <domain> [percent|auto]", "Set a domain's share of your workers (0 keeps it empty; auto, the default, follows its buildings' slots); no percent shows it"},
				{"roster auto", "Put every domain back on auto"},
			}},
		{Name: "workers", Section: secWorkers, BareOK: true, Panel: "Workers: domains, roster & assignments",
			Help: []Usage{{"workers", "Open the Workers panel (domains, roster, assignments)"}},
			Subs: []*Command{
				// The roster's old name: `workers share ...` is `roster ...`.
				{Name: "share", BareOK: true, Quiet: true,
					Args: []Arg{{Kind: ArgDomain, Words: []string{"auto"}}, {Kind: ArgPercent, Words: []string{"auto"}, Optional: true}}},
				{Name: "auto-recruit", Aliases: []string{"autorecruit"},
					Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
					Help: []Usage{{"workers auto-recruit [on|off]", "Show or set whether the game recruits into empty worker slots as housing and food allow (on by default)"}}},
			}},

		// Research, Expeditions & Army
		{Name: "research", Aliases: []string{"res"}, Section: secResearch, Panel: "The tech tree: a map of every tech in sight",
			Args: []Arg{{Kind: ArgTech, Optional: true}},
			Help: []Usage{{"research <tech>", "Research a tech"}},
			Subs: []*Command{
				{Name: "cancel", Dangerous: true, Help: []Usage{{"research cancel", "Cancel current research (the knowledge spent is not refunded)"}}},
				sub("list", "research list", "List available techs"),
				sub("tree", "research tree [close|far]", "Open the tech tree, zoomed in on big badges (close) or out on the whole tree (far)", Arg{Kind: ArgWord, Words: []string{"close", "far"}, Optional: true}),
				sub("card", "research card <tech>", "Open the tech tree on a tech's card: what it does, costs and needs", Arg{Kind: ArgTech}),
			}},
		panel("techs", ""),
		{Name: "expedition", Aliases: []string{"exp"}, Section: secResearch, Panel: "Scouting expeditions (resource cost)",
			Args: []Arg{{Kind: ArgExpedition, Optional: true}},
			Help: []Usage{
				{"expedition", "Open the Expeditions (scouting) panel"},
				{"expedition <expedition>", "Send a scouting expedition (costs resources)"},
			},
			Subs: []*Command{sub("list", "expedition list", "List available expeditions")}},
		{Name: "army", Section: secResearch, Panel: "Army overview & military campaigns",
			Help: []Usage{{"army", "Open the Army (military) panel"}}},
		{Name: "campaign", Section: secResearch,
			Args: []Arg{{Kind: ArgCampaign, Optional: true}},
			Help: []Usage{{"campaign <campaign>", "Wage a military campaign (costs soldiers)"}},
			Subs: []*Command{sub("list", "campaign list", "List available campaigns")}},

		// Trade & Diplomacy
		{Name: "trade", Aliases: []string{"t"}, Section: secTrade, BareOK: true, Panel: "Market rates & trade routes",
			Args: []Arg{{Kind: ArgTradeFrom}, {Kind: ArgTradeTo}, {Kind: ArgNumber}},
			Help: []Usage{{"trade <give> <get> <amount>", "Sell <amount> of <give> for <get> at the market rate"}},
			Subs: []*Command{
				sub("list", "trade list", "Show market rates"),
				{Name: "route", BareOK: true, Subs: []*Command{
					sub("list", "trade route list", "List trade routes"),
					sub("start", "trade route start <route>", "Start a trade route", Arg{Kind: ArgRouteAvailable}),
					sub("stop", "trade route stop <route>", "Stop a trade route", Arg{Kind: ArgRouteActive}),
				}},
				sub("black", "trade black [resource]", "Same as blackmarket", Arg{Kind: ArgResource, Optional: true}),
			}},
		{Name: "blackmarket", Aliases: []string{"bm"}, Section: secTrade,
			Args: []Arg{{Kind: ArgResource, Optional: true}},
			Help: []Usage{{"blackmarket [resource]", "Smuggling run: gamble culture on a haul of one resource (bare blackmarket shows the odds, or when it opens)"}}},
		{Name: "factions", Section: secTrade, Panel: "Boons, the Geographic Society & opinion of each civilization (alias: diplomacy)",
			Help: []Usage{{"factions", "Open the Factions panel (boons, the Geographic Society, opinion)"}}},
		{Name: "diplomacy", Aliases: []string{"dip"}, Section: secTrade, BareOK: true,
			Help: []Usage{{"diplomacy", "Alias for factions (opens the same panel)"}},
			Subs: []*Command{
				sub("ally", "diplomacy ally <civ>", "Ally with a civilization ("+game.Amount(game.AllyCost, "gold")+" or less, needs opinion "+strconv.Itoa(game.AllyOpinion)+")", civArg...),
				sub("rival", "diplomacy rival <civ>", "Declare a civilization your rival", civArg...),
				sub("embargo", "diplomacy embargo <civ>", "Embargo a civilization (a provocation: it can start a war)", civArg...),
				sub("gift", "diplomacy gift <civ>", "Send a gift of gold for +"+strconv.Itoa(game.GiftOpinion)+" opinion or more (the Factions panel shows today's price and what it earns)", civArg...),
				sub("neutral", "diplomacy neutral <civ>", "Return to neutral with a civilization", civArg...),
				sub("tribute", "diplomacy tribute <civ>", "Sue for peace with a civilization at war", civArg...),
				sub("deals", "diplomacy deals [civ]", "List trade deals (one civilization, or every one you have met)",
					Arg{Kind: ArgFaction, Optional: true}),
				sub("accept", "diplomacy accept <civ> <n>", "Take a civilization's trade deal n",
					Arg{Kind: ArgFaction}, Arg{Kind: ArgDeal}),
				{Name: "raid", Dangerous: true, Args: civArg,
					Help: []Usage{{"diplomacy raid <civ>", "Raid a civilization's trade route (a war provocation)"}}},
			}},

		// Wonders & Prestige
		{Name: "wonder", Section: secWonders, BareOK: true,
			Help: []Usage{{"wonder", "Show current wonder bank status"}},
			Subs: []*Command{
				{Name: "collect", Aliases: []string{"bank"},
					Args: []Arg{{Kind: ArgWonderResource, Words: []string{"all"}}, {Kind: ArgNumber, Words: []string{"all", "max"}, Optional: true}},
					Help: []Usage{{"wonder collect <resource|all> [amount|all|max]", "Bank resources into the current wonder (alias: bank; no amount: as much as it needs)"}}},
				sub("overflow", "wonder overflow [on|off]", "Bank what full stores would waste (on by default)",
					Arg{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}),
			}},
		panel("wonders", "Wonder bank & built wonders"),
		{Name: "prestige", Section: secWonders, BareOK: true,
			Help: []Usage{{"prestige", "View prestige status"}},
			Subs: []*Command{
				prestigeConfirm,
				sub("shop", "prestige shop", "View the legacy kit"),
				sub("buy", "prestige buy <item>", "Buy a legacy kit item", Arg{Kind: ArgPrestigeUpgrade}),
			}},
		{Name: "festival", Section: secWonders, BareOK: true,
			Help: []Usage{{"festival", "Spend culture for a temporary production boost"}},
			Subs: []*Command{festivalConfirm}},
		{Name: "catastrophe", Aliases: []string{"cat"}, Section: secWonders,
			Help: []Usage{{"catastrophe", "Reopen a pending catastrophe or Last Passage (or show the outlook)"}}},
		{Name: "harbinger", Aliases: []string{"harb"}, Section: secWonders, BareOK: true,
			Panel: "The harbinger's warning & your answers (alias: harb)",
			Help: []Usage{
				{"harbinger", "Open the Harbinger panel (alias: harb)"},
				{"harbinger appease|brace|invite", "Answer the harbinger without the panel"},
			},
			Subs: []*Command{{Name: "appease"}, {Name: "brace"}, {Name: "invite", Dangerous: true}}},

		// Game
		{Name: "rates", Section: secGame, Help: []Usage{{"rates", "Show resource rate breakdown"}}},
		{Name: "status", Aliases: []string{"s"}, Section: secGame, Help: []Usage{{"status", "Show detailed status"}}},
		{Name: "theme", Section: secGame, Panel: "Theme picker: palettes & accessibility",
			Args: []Arg{{Kind: ArgTheme, Optional: true}},
			Help: []Usage{
				{"theme", "Open the theme picker (palettes + accessibility)"},
				{"theme <key>", "Switch to a theme by key"},
			},
			Subs: []*Command{sub("list", "theme list", "List themes with unlock status")}},
		{Name: "map", Aliases: []string{"citymap", "worldmap"}, Section: secGame, BareOK: true,
			Panel: "The map: your settlement and the known world (aliases: citymap, worldmap)",
			Help: []Usage{
				{"map", "Open the Map panel (worldmap opens it on the known world); the prompt keeps working while it is open"},
			},
			Subs: []*Command{
				sub("style", "map style [roguelike|skyline]", "Show or set the map style (default roguelike)", mapStyleArg),
				sub("glyphs", "map glyphs [ascii|unicode|nerd]", "Show or set the map's glyphs (default unicode; nerd needs a Nerd Font)",
					Arg{Kind: ArgWord, Words: []string{"ascii", "unicode", "nerd"}, Optional: true}),
				sub("flows", "map flows [on|off]", "Turn the Map's flows overlay on or off (bare: switch it)",
					Arg{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}),
			}},
		{Name: "style", Section: secGame, Args: []Arg{mapStyleArg},
			Help: []Usage{{"style [roguelike|skyline]", "Same as map style"}}},
		{Name: "minimap", Section: secGame, Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
			Help: []Usage{{"minimap [on|off]", "Show or set the mini map above the Buildings list (default off)"}}},
		{Name: "motion", Section: secGame, Args: []Arg{{Kind: ArgWord, Words: []string{"on", "off"}, Optional: true}},
			Help: []Usage{{"motion [on|off]", "Show or set motion: whether the maps, the badge case and theme effects move (default on)"}}},
		{Name: "icons", Section: secGame,
			Help: []Usage{{"icons", "Check whether your font shows Nerd Font icons, and install one if it doesn't"}}},
		{Name: "save", Section: secGame, Args: []Arg{{Kind: ArgText, Optional: true}},
			Help: []Usage{{"save [name]", "Save (no name: overwrite or branch; a name: branch a new save)"}},
			Subs: []*Command{sub("list", "save list", "Same as saves")}},
		{Name: "load", Section: secGame, Dangerous: true, Args: []Arg{{Kind: ArgSave, Optional: true}},
			Help: []Usage{{"load [name]", "Load a save (no name: open the save browser)"}}},
		{Name: "saves", Section: secGame, Help: []Usage{{"saves", "List all save files"}}},
		{Name: "dump", Aliases: []string{"exportlogs"}, Section: secGame,
			Help: []Usage{{"dump", "Export logs to file for debugging"}}},
		{Name: "help", Aliases: []string{"h", "?"}, Section: secGame, Panel: "This Help panel",
			Help: []Usage{{"help", "Open this Help panel"}}},
		{Name: "quit", Section: secGame, Dashboard: true, Dangerous: true,
			Help: []Usage{{"quit", "Save and quit the game"}}},

		// Accounts
		{Name: "account", Aliases: []string{"acct"}, Section: secAccounts, BareOK: true,
			Help: []Usage{{"account", "Show this account's ID, recovery code & backup help"}},
			Subs: []*Command{
				sub("list", "account list", "List your local accounts"),
				sub("badges", "account badges", "List this account's badges, earned and locked"),
				{Name: "switch", Dangerous: true, Args: []Arg{{Kind: ArgAccount}},
					Help: []Usage{{"account switch <name>", "Switch to an existing local account"}}},
				sub("export", "account export [path]", "Back up this account's progress to a file", Arg{Kind: ArgText, Optional: true}),
				sub("backup", "account backup", "Full snapshot (account files + saves) to data/backups/"),
				{Name: "import", Dangerous: true, Args: []Arg{{Kind: ArgText}, {Kind: ArgWord, Words: []string{"replace"}, Optional: true}},
					Help: []Usage{{"account import <path> [replace]", "Restore an account from a backup file"}}},
				{Name: "recover", Dangerous: true, Args: []Arg{{Kind: ArgText}, {Kind: ArgWord, Words: []string{"confirm"}, Optional: true}},
					Help: []Usage{{"account recover <code> [confirm]", "Restore your account ID from a recovery code"}}},
				{Name: "wipe", Dangerous: true,
					Help: []Usage{{"account wipe", "Where to wipe an account (the Accounts panel)"}}},
			}},

		{Name: "badges", Aliases: []string{"achievements"}, Section: secAccounts, BareOK: true,
			Panel: "The badge case: every badge your account holds or can earn (alias: achievements)",
			Args:  []Arg{{Kind: ArgText, Words: []string{"next", "all"}, Optional: true}},
			Help: []Usage{
				{"badges", "Open the badge case (alias: achievements); the prompt keeps working while it is open"},
				{"badges <family>", "Open it on a family's tab: ages, lineages, ladders, specials and the rest (all: every family)"},
				{"badges next", "Open it on the badges you are closest to"},
				{"badges <name>", "Open a badge's detail by its name, or a part of it"},
			}},
		{Name: "title", Section: secAccounts, BareOK: true,
			Args: []Arg{{Kind: ArgText, Words: []string{"default"}, Optional: true}},
			Help: []Usage{
				{"title", "List the titles your account holds, and the one it wears"},
				{"title <name>", "Wear a title you hold (default: the one your badge score gives)"},
			}},

		// Panels only
		panel("milestones", "Milestone goals & rewards", "ms"),
		panel("stats", "Empire statistics"),
		panel("logs", "Recent game log entries"),
		panel("epoch", "Epoch progress & catastrophe"),
		panel("history", "Civilization history timeline"),
		panel("buildings", "Built structures by lineage"),
	}
}

// panelOrder is the Help panel's Panels list order (the sidebar's, then the rest).
var panelOrder = []string{
	"milestones", "badges", "research", "plan", "expedition", "army", "trade", "factions", "stats", "wonders",
	"workers", "logs", "epoch", "harbinger", "history", "buildings", "map", "theme", "help",
}

// devCommand is one dev-console command. They are not player commands: the
// dashboard sends a /line straight to game.DevConsoleCommand, and they are
// listed and completed only while dev mode is active.
type devCommand struct{ name, form, text string }

var devCommands = []devCommand{
	{"/god", "/god", "Toggle godmode: free costs, instant builds"},
	{"/fill", "/fill", "Fill all resources to their storage cap"},
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
