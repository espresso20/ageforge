package ui

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// CommandResult is the return value of HandleCommand. The caller (Dashboard)
// logs Message if it is non-empty and Type != "success" (successes are
// ephemeral and only shown as toast/log entries by the engine itself).
// A routine confirmation (Type game.LogRoutine) shows in both logs like any
// other reply; the logs panel marks it with a dot (log_routing.go).
// If OverlayName is set the Dashboard opens that named overlay panel.
type CommandResult struct {
	Message     string
	Type        string // "info", "success", "error", "warning", game.LogRoutine
	OverlayName string // non-empty → dashboard should open this overlay panel
	// OpenCatastrophe asks the dashboard to (re)open the pending catastrophe modal.
	OpenCatastrophe bool
	// MapWorld opens the Map panel (OverlayName "map") on the known world.
	MapWorld bool
	// ResearchZoom ("close" or "far") and ResearchCard (a tech key) say how
	// the Research panel (OverlayName "techs") opens: at that zoom, on that
	// tech's card. Empty leaves the panel as it was.
	ResearchZoom string
	ResearchCard string
	// Icons asks the dashboard to start the guided icons check.
	Icons bool
	// MapPref is a map setting change (map style, map glyphs, minimap) for
	// the dashboard to apply at once. The command has already saved it to
	// the account; with no account loaded the dashboard keeps it for the
	// session.
	MapPref mapPref
	// MapFlows asks the dashboard to turn the Map's flows overlay "on",
	// "off" or over ("switch"); it then writes the reply.
	MapFlows string
	// ToMenu says the command ended the game in progress (an account switch or
	// recovery stopped the tick loop and saved the run to the account it belongs
	// to): the dashboard goes to the main menu, without saving again, and shows
	// Message there, since the log it would go to belongs to the ended run.
	ToMenu bool
}

// HandleCommand parses a raw command string and dispatches to the appropriate
// sub-handler. Single-letter shortcuts (g, b, r, a, u, s, t) are normalised
// to their full equivalents before switching. Commands that purely open an
// overlay return an empty Message with OverlayName set.
func HandleCommand(input string, engine *game.GameEngine) CommandResult {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return CommandResult{}
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help", "h", "?":
		return CommandResult{OverlayName: "help"}
	case "gather", "g":
		return cmdGather(args, engine)
	case "build", "b":
		return cmdBuild(args, engine)
	case "recruit", "r":
		return cmdRecruit(args, engine)
	case "assign", "a":
		return cmdAssign(args, engine)
	case "unassign", "u":
		return cmdUnassign(args, engine)
	case "dismiss":
		return cmdDismiss(args, engine)
	case "sell":
		return cmdSell(args, engine)
	case "status", "s":
		return cmdStatus(engine)
	case "research", "res":
		if len(args) == 0 {
			return CommandResult{OverlayName: "techs"}
		}
		return cmdResearch(args, engine)
	case "expedition", "exp":
		return cmdExpedition(args, engine)
	case "campaign":
		return cmdCampaign(args, engine)
	case "trade", "t":
		return cmdTrade(args, engine)
	case "diplomacy", "dip":
		if len(args) == 0 {
			// Opens the Factions panel under its PRIMARY overlay name, not under
			// "diplomacy": the sidebar highlight matches on the name it is handed,
			// so returning the alias would open the panel without lighting up the
			// entry the player just navigated to.
			return CommandResult{OverlayName: "factions"}
		}
		return cmdDiplomacy(args, engine)
	case "wonder":
		return cmdWonder(args, engine)
	case "plan":
		return cmdPlan(args, engine)
	case "prestige":
		return cmdPrestige(args, engine)
	case "festival":
		return cmdFestival(args, engine)
	case "blackmarket", "bm":
		return cmdBlackMarket(args, engine)
	case "rates":
		return cmdRates(engine)
	case "upgrade":
		return cmdUpgrade(args, engine)
	case "advance":
		return cmdAdvance(engine)
	case "factions":
		return CommandResult{OverlayName: "factions"}
	case "milestones", "ms":
		return CommandResult{OverlayName: "milestones"}
	case "techs":
		return CommandResult{OverlayName: "techs"}
	case "army":
		return CommandResult{OverlayName: "army"}
	case "stats":
		return CommandResult{OverlayName: "stats"}
	case "wonders":
		return CommandResult{OverlayName: "wonders"}
	case "workers":
		return cmdWorkers(args, engine)
	case "logs":
		return CommandResult{OverlayName: "logs"}
	case "epoch":
		return CommandResult{OverlayName: "epoch"}
	case "history":
		return CommandResult{OverlayName: "history"}
	case "buildings":
		return CommandResult{OverlayName: "buildings"}
	case "map", "citymap", "worldmap":
		// citymap and worldmap are the old maps' commands, kept as aliases;
		// worldmap opens on the known world.
		return cmdMap(cmd, args, engine)
	case "style":
		// style is map style's shortcut: players reach for it first.
		return cmdMapStyle(args, engine)
	case "minimap":
		return cmdMinimap(args, engine)
	case "catastrophe", "cat":
		return cmdCatastrophe(args, engine)
	case "harbinger", "harb":
		return cmdHarbinger(args, engine)
	case "dump", "exportlogs":
		return cmdDump(args, engine)
	case "saves":
		return cmdSaveList()
	case "save":
		if len(args) > 0 && args[0] == "list" {
			return cmdSaveList()
		}
		return cmdSave(args, engine)
	case "load":
		return cmdLoad(args, engine)
	case "account", "acct":
		return cmdAccount(args, engine)
	case "theme":
		return cmdTheme(args, engine)
	case "icons":
		if len(args) > 0 {
			return CommandResult{Message: usageFor("icons"), Type: "error"}
		}
		return CommandResult{Icons: true}
	default:
		return CommandResult{Message: unknownCommandText(cmd), Type: "error"}
	}
}

// unknownCommandText is the refusal for a first word that is no command:
// "Unknown command 'biuld'. Did you mean 'build'? Type help for all commands."
func unknownCommandText(cmd string) string {
	msg := fmt.Sprintf("Unknown command '%s'.", cmd)
	if s := closestCommand(cmd); s != "" {
		msg += fmt.Sprintf(" Did you mean '%s'?", s)
	}
	return msg + " Type help for all commands."
}

// usageFor is "Usage: " and the registry's help form for the command path
// (the longest form that starts with it), so a usage line names its slots
// exactly as the Help panel does. The log escapes the [slot] brackets on the
// way to the screen (safeTags).
func usageFor(path string) string {
	return "Usage: " + helpRow(path).Form
}

// helpRow is the registry help row for the command path: the form that is
// the path, or the path followed by its argument slots (the longest such
// form, so `upgrade` finds "upgrade <building> [count|all]"). A path with no
// row gives the path itself (TestUsagePathsExist keeps every caller's path
// real).
func helpRow(path string) Usage {
	best := Usage{Form: path}
	found := false
	fits := func(form string) bool {
		if form == path {
			return true
		}
		rest, ok := strings.CutPrefix(form, path+" ")
		return ok && rest != "" && strings.ContainsAny(rest[:1], "<[")
	}
	var walk func(c *Command)
	walk = func(c *Command) {
		for _, u := range c.Help {
			if fits(u.Form) && (!found || len(u.Form) > len(best.Form)) {
				best, found = u, true
			}
		}
		for _, s := range c.Subs {
			walk(s)
		}
	}
	for _, c := range registry() {
		walk(c)
	}
	return best
}

// subUsage is "Usage: " and every help form under the command, joined with
// " | ": the reply to a subcommand the command does not know.
func subUsage(name string) string {
	c := lookup(registry(), name)
	if c == nil {
		return "Usage: " + name
	}
	var forms []string
	for _, u := range appendHelpRows(nil, c) {
		forms = append(forms, u.Form)
	}
	return "Usage: " + strings.Join(forms, " | ")
}

// maxCommandCount is the largest count any command accepts. No game gets
// anywhere near it, and capping it at the prompt keeps every count-times-cost
// sum finite and every population check free of integer overflow.
const maxCommandCount = 1_000_000

// parseCount reads a count argument: a whole number from 1 to
// maxCommandCount. The error says what was wrong, for the caller to put after
// its usage line.
func parseCount(arg string) (int, error) {
	n, err := strconv.Atoi(arg)
	if err != nil || n < 1 || n > maxCommandCount {
		return 0, fmt.Errorf("the count must be a whole number from 1 to %d (got %q)", maxCommandCount, arg)
	}
	return n, nil
}

// parseAmount reads a resource amount: a positive, finite number. ParseFloat
// happily returns NaN and infinities for "NaN" and "Inf", and NaN slips past
// every comparison, so both are refused here.
func parseAmount(arg string) (float64, error) {
	n, err := strconv.ParseFloat(arg, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n <= 0 {
		return 0, fmt.Errorf("the amount must be a positive number (got %q)", arg)
	}
	return n, nil
}

// usageError is the error result for a bad argument: the usage line, then
// what was wrong with the argument.
func usageError(usage string, err error) CommandResult {
	return CommandResult{Message: usage + " (" + err.Error() + ")", Type: "error"}
}

// errorResult is the refusal for an engine error, as a sentence: capital
// first letter, closing period.
func errorResult(err error) CommandResult {
	return CommandResult{Message: textfmt.Sentence(err.Error()), Type: "error"}
}

// sortedKeysOf returns m's keys in order, so lists built from game maps come
// out the same on every call.
func sortedKeysOf[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func cmdWonder(args []string, engine *game.GameEngine) CommandResult {
	if len(args) >= 1 && strings.ToLower(args[0]) == "overflow" {
		return cmdWonderOverflow(args[1:], engine)
	}
	state := engine.GetState()

	// Find the wonder for the current age
	var curWonder *wonderInfo
	for _, w := range getWonderList() {
		if w.ageKey == state.Age {
			wCopy := w
			curWonder = &wCopy
			break
		}
	}
	if curWonder == nil {
		return CommandResult{Message: "No wonder available this age.", Type: "error"}
	}

	deposit := false
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "collect", "bank":
			deposit = true
		default:
			return CommandResult{Message: fmt.Sprintf("Unknown wonder command %q. %s, or 'wonder overflow [on|off]'.", args[0], wonderBankUsage), Type: "error"}
		}
	}

	bs := state.Buildings[curWonder.key]
	if bs.Count > 0 {
		if deposit {
			return CommandResult{Message: fmt.Sprintf("%s is already built: there is nothing left to bank this age.", curWonder.name), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("[gold]★ %s[-] is already built.", curWonder.name),
			Type:    "info",
		}
	}
	if deposit {
		return cmdWonderCollect(args[1:], curWonder, state, engine)
	}

	// Default: show bank status
	var sb strings.Builder
	fmt.Fprintf(&sb, "[gold::b]%s[-]: wonder bank\n\n", curWonder.name)

	costKeys := make([]string, 0, len(curWonder.def.BaseCost))
	for k := range curWonder.def.BaseCost {
		costKeys = append(costKeys, k)
	}
	sort.Strings(costKeys)

	for _, res := range costKeys {
		need := curWonder.def.BaseCost[res]
		banked := bs.WonderBank[res]
		pct := 0.0
		if need > 0 {
			pct = banked / need * 100
			if pct > 100 {
				pct = 100
			}
		}
		clr := "red"
		if pct >= 100 {
			clr = "green"
		} else if pct > 0 {
			clr = "yellow"
		}
		fmt.Fprintf(&sb, "  [%s]%s: %s / %s (%.0f%%)[-]\n", clr, game.ResourceName(res), FormatNumber(banked), FormatNumber(need), pct)
	}

	if line := wonderKeystoneLine(state, curWonder.key); line != "" {
		fmt.Fprintf(&sb, "\n  %s\n", line)
	}
	switch {
	case bs.WonderBankFull && bs.NeedsTech != "":
		fmt.Fprintf(&sb, "\n[yellow]Bank full. Research %s, then type 'build %s' to start construction.[-]", game.TechName(bs.NeedsTech), curWonder.key)
	case bs.WonderBankFull:
		fmt.Fprintf(&sb, "\n[green]Bank full. Type 'build %s' to start construction.[-]", curWonder.key)
	default:
		fmt.Fprintf(&sb, "\n[gray]Bank resources with 'wonder collect <resource|all> [amount|all|max]'.[-]")
	}
	return CommandResult{Message: sb.String(), Type: "info"}
}

const wonderBankUsage = "Usage: wonder collect|bank <resource|all> [amount|all|max]"

// cmdWonderCollect is `wonder collect|bank <resource|all> [amount|all|max]`:
// bank resources into the current age's unbuilt wonder. `all` or `max` (or
// no amount) banks as much as the wonder still needs, up to what is on hand;
// `all` for the resource does that for every resource the wonder needs.
func cmdWonderCollect(args []string, w *wonderInfo, state game.GameState, engine *game.GameEngine) CommandResult {
	if len(args) == 0 || len(args) > 2 {
		return CommandResult{Message: wonderBankUsage, Type: "error"}
	}
	resource := strings.ToLower(args[0])
	whole := true // bank as much as can go in
	var amount float64
	if len(args) == 2 {
		switch strings.ToLower(args[1]) {
		case "all", "max":
		default:
			if resource == "all" {
				return usageError(wonderBankUsage, fmt.Errorf("'wonder collect all' banks every resource as far as it goes; give an amount for one resource at a time"))
			}
			var err error
			if amount, err = parseAmount(args[1]); err != nil {
				return usageError(wonderBankUsage, err)
			}
			whole = false
		}
	}
	if resource == "all" {
		return wonderCollectAll(w, engine)
	}
	if _, ok := state.Resources[resource]; !ok {
		return CommandResult{Message: fmt.Sprintf("Unknown resource '%s'. Type wonder to see what %s still needs.", args[0], w.name), Type: "error"}
	}

	var deposited float64
	var err error
	if whole {
		deposited, err = engine.BankWonderMax(w.key, resource)
	} else {
		deposited, err = engine.BankWonderResource(w.key, resource, amount)
	}
	if err != nil {
		return CommandResult{Message: "Nothing banked. " + textfmt.Sentence(err.Error()), Type: "error"}
	}
	// The engine logs the deposit (and a full bank); only a short deposit
	// needs a word of its own.
	if !whole && deposited < amount {
		return CommandResult{Message: fmt.Sprintf("Only %s went in: that is all %s still needed.",
			game.Amount(deposited, resource), w.name), Type: "info"}
	}
	return CommandResult{Type: "success"}
}

// wonderCollectAll is `wonder collect all`: bank every resource the wonder
// still needs, each as far as what is on hand goes. The engine logs each
// deposit; the reply names what was not banked and why.
func wonderCollectAll(w *wonderInfo, engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	bs := state.Buildings[w.key]
	var banked int
	var skipped, full []string // full: already banked in full, named only when nothing went in
	for _, res := range sortedMapKeys(w.def.BaseCost) {
		name := game.ResourceName(res)
		switch {
		case bs.WonderBank[res] >= w.def.BaseCost[res]:
			full = append(full, name+" (already has all it needs)")
			continue
		case state.Resources[res].Amount <= 0:
			skipped = append(skipped, name+" (you have none)")
			continue
		}
		if _, err := engine.BankWonderMax(w.key, res); err != nil {
			skipped = append(skipped, name+" ("+strings.TrimRight(lowerFirst(err.Error()), ".")+")")
			continue
		}
		banked++
	}
	if banked == 0 {
		all := append(skipped, full...)
		sort.Strings(all)
		return CommandResult{Message: "Nothing banked: " + strings.Join(all, ", ") + ".", Type: "error"}
	}
	if len(skipped) > 0 {
		return CommandResult{Message: "Not banked: " + strings.Join(skipped, ", ") + ".", Type: "info"}
	}
	return CommandResult{Type: "success"}
}

// lowerFirst lowercases the first letter of s, to set an engine sentence
// inside parentheses.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// cmdWonderOverflow is `wonder overflow [on|off]`: show or set whether
// production a full store would waste goes into the current wonder's bank.
func cmdWonderOverflow(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{Message: wonderOverflowLine(engine.WonderOverflow()), Type: "info"}
	}
	switch strings.ToLower(args[0]) {
	case "on":
		engine.SetWonderOverflow(true)
		return CommandResult{Message: "Wonder overflow on: what a full store would waste now goes into the current wonder's bank.", Type: "info"}
	case "off":
		engine.SetWonderOverflow(false)
		return CommandResult{Message: "Wonder overflow off: production over a storage cap is lost again.", Type: "info"}
	}
	return CommandResult{Message: "Usage: wonder overflow [on|off]", Type: "error"}
}

// cmdWorkers is `workers`: bare opens the Workers panel, `workers share`
// shows or sets the worker shares and `workers auto-recruit [on|off]` shows
// or sets recruiting.
func cmdWorkers(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "workers"}
	}
	switch strings.ToLower(args[0]) {
	case "share":
		return cmdWorkersShare(args[1:], engine)
	case "auto-recruit", "autorecruit":
		return cmdAutoRecruit(args[1:], engine)
	}
	return CommandResult{Message: fmt.Sprintf("Unknown workers command '%s'. %s", args[0], subUsage("workers")), Type: "error"}
}

// cmdWorkersShare is `workers share [domain] [percent|auto]` and `workers
// share auto`: show the shares, show one domain's, set one, or put one or
// all back on auto. A change replies with the shares now and what moved.
func cmdWorkersShare(args []string, engine *game.GameEngine) CommandResult {
	usage := usageFor("workers share")
	if len(args) == 0 {
		return CommandResult{Message: sharesText(engine.GetState(), ""), Type: "info"}
	}
	if len(args) > 2 {
		return CommandResult{Message: usage, Type: "error"}
	}
	domain := strings.ToLower(args[0])
	var reply game.ShareReply
	var err error
	switch {
	case domain == "auto" && len(args) == 1:
		reply, err = engine.ClearWorkerShare("")
	case !game.IsWorkerDomain(engine.Rules(), domain):
		return CommandResult{Message: game.UnknownDomainError(engine.Rules(), args[0]).Error(), Type: "error"}
	case len(args) == 1:
		return CommandResult{Message: sharesText(engine.GetState(), domain), Type: "info"}
	case strings.ToLower(args[1]) == "auto":
		reply, err = engine.ClearWorkerShare(domain)
	default:
		pct, perr := parsePercent(args[1])
		if perr != nil {
			return usageError(usage, perr)
		}
		reply, err = engine.SetWorkerShare(domain, pct)
	}
	if err != nil {
		return errorResult(err)
	}
	if reply.Warning {
		return CommandResult{Message: reply.Line, Type: "warning"}
	}
	return CommandResult{Message: reply.Line, Type: game.LogRoutine}
}

// parsePercent reads a share: a number from 0 to 100, with or without a %
// sign.
func parsePercent(arg string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSuffix(arg, "%"), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 100 {
		return 0, fmt.Errorf("the share must be a percent from 0 to 100 (got %q)", arg)
	}
	return v, nil
}

// sharesText is `workers share`: each domain's share of the workforce with
// its workers and slots, and whether auto-recruit is on. With domain set,
// that domain's line only.
func sharesText(st game.GameState, domain string) string {
	rows := game.ShareRows(st)
	if domain != "" {
		for _, r := range rows {
			if r.Domain == domain {
				return shareRowText(r)
			}
		}
		return fmt.Sprintf("%s: on auto. You have no %s buildings yet.", game.DomainName(domain), strings.ToLower(game.DomainName(domain)))
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "[gold]Worker shares[-] (auto-recruit %s)\n", onOff(st.Workers.AutoRecruit))
	if len(rows) == 0 {
		sb.WriteString("  No worker buildings yet: every domain is on auto.\n")
	}
	for _, r := range rows {
		sb.WriteString("  " + shareRowText(r) + "\n")
	}
	sb.WriteString("Set one with 'workers share <domain> <percent|auto>'.")
	return sb.String()
}

// shareRowText is one domain's line: "Knowledge: 40% (set), 8 workers in 10
// slots".
func shareRowText(r game.ShareRow) string {
	share := textfmt.Percent(r.Percent/100) + " (auto)"
	if r.Set {
		share = game.SharePercent(math.Round(r.Percent*10)/10) + " (set)"
	}
	return fmt.Sprintf("%s: %s, %s in %s", r.Name, share, textfmt.Count(r.Workers, "worker", "workers"), textfmt.Count(r.Slots, "slot", "slots"))
}

// onOff is "on" or "off".
func onOff(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

// cmdAutoRecruit is `workers auto-recruit [on|off]`: show or set whether the
// game recruits into empty worker slots.
func cmdAutoRecruit(args []string, engine *game.GameEngine) CommandResult {
	switch {
	case len(args) == 0:
		if engine.AutoRecruit() {
			return CommandResult{Message: "Auto-recruit is on: the game recruits into empty worker slots while housing and food allow. Turn it off with 'workers auto-recruit off'.", Type: "info"}
		}
		return CommandResult{Message: "Auto-recruit is off: recruit by hand with 'recruit'. Turn it on with 'workers auto-recruit on'.", Type: "info"}
	case len(args) > 1:
		return CommandResult{Message: usageFor("workers auto-recruit"), Type: "error"}
	}
	switch strings.ToLower(args[0]) {
	case "on":
		engine.SetAutoRecruit(true)
		return CommandResult{Message: "Auto-recruit on: the game recruits into empty worker slots while housing and food allow.", Type: game.LogRoutine}
	case "off":
		engine.SetAutoRecruit(false)
		return CommandResult{Message: "Auto-recruit off: recruit by hand with 'recruit'. Idle workers still go to work by your shares.", Type: game.LogRoutine}
	}
	return CommandResult{Message: usageFor("workers auto-recruit"), Type: "error"}
}

func cmdAdvance(engine *game.GameEngine) CommandResult {
	if err := engine.AdvanceAge(); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs the new age
}

func cmdUpgrade(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		// List available upgrades
		upgrades := engine.GetAvailableUpgrades()
		if len(upgrades) == 0 {
			return CommandResult{Message: "No building upgrades available right now.", Type: "info"}
		}
		var lines []string
		lines = append(lines, "[gold]Upgrades you can make[-] (pays the difference in cost: the new copy's cost less a 50% refund on the old one)")
		sort.Slice(upgrades, func(i, j int) bool { return upgrades[i].FromKey < upgrades[j].FromKey })
		for _, u := range upgrades {
			affordable := "[red]✗[-]"
			if u.CanAfford {
				affordable = "[green]✓[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] → [cyan]%s[-]: %s, cost for all: %s",
				affordable, u.FromKey, u.ToKey, textfmt.Count(u.Count, "copy", "copies"), FormatCost(u.Cost)))
		}
		lines = append(lines, "\n  Type [cyan]"+lit(strings.TrimPrefix(usageFor("upgrade"), "Usage: "))+"[-]")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}

	building := strings.ToLower(args[0])
	all := false
	count := 0
	if len(args) >= 2 {
		if strings.ToLower(args[1]) == "all" {
			all = true
		} else {
			n, err := parseCount(args[1])
			if err != nil {
				return usageError(usageFor("upgrade"), err)
			}
			count = n
		}
	} else {
		all = true // default: upgrade all
	}

	if err := engine.UpgradeBuilding(building, count, all); err != nil {
		return errorResult(err)
	}
	return CommandResult{Message: "", Type: "success"}
}

func cmdDump(args []string, engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	logs := engine.GetLogs()

	// Dumps go in a logs folder in the active account's data directory, where
	// its saves live. (A relative "data/logs" landed wherever the game was
	// launched from.)
	dir := filepath.Join(game.DataDir(), "logs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return CommandResult{Message: "Could not create the logs folder: " + shortIOError(err) + ".", Type: "error"}
	}

	// Generate timestamped filename
	ts := time.Now().Format("2006-01-02_150405")
	filename := filepath.Join(dir, fmt.Sprintf("dump_%s.log", ts))

	var sb strings.Builder

	// Header with engine state
	sb.WriteString("=== AgeForge Log Dump ===\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Tick: %d\n", state.Tick))
	sb.WriteString(fmt.Sprintf("Age: %s (%s)\n", state.AgeName, state.Age))
	sb.WriteString(fmt.Sprintf("Population: %d/%d (idle: %d, food drain: %.2f/tick)\n",
		state.Workers.TotalPop, state.Workers.MaxPop, state.Workers.TotalIdle, state.Workers.FoodDrain))
	sb.WriteString("\n--- Resources ---\n")
	for _, rs := range state.Resources {
		if !rs.Unlocked {
			continue
		}
		sb.WriteString(fmt.Sprintf("  %-12s %8.1f / %8.0f  rate: %+.3f/tick\n",
			rs.Name, rs.Amount, rs.Storage, rs.Rate))
	}
	sb.WriteString("\n--- Build Queue ---\n")
	if len(state.BuildQueue) == 0 {
		sb.WriteString("  (empty)\n")
	}
	for _, bq := range state.BuildQueue {
		sb.WriteString(fmt.Sprintf("  %s: %d/%d ticks\n", bq.Name, bq.TotalTicks-bq.TicksLeft, bq.TotalTicks))
	}
	sb.WriteString("\n--- Active Events ---\n")
	if len(state.ActiveEvents) == 0 {
		sb.WriteString("  (none)\n")
	}
	for _, evt := range state.ActiveEvents {
		// This file is a DEBUG dump read by us, not by players, so it keeps the
		// raw tick count — with the wall-clock reading alongside it.
		sb.WriteString(fmt.Sprintf("  %s: %d ticks left (%s)\n",
			evt.Name, evt.TicksLeft, formatTicks(evt.TicksLeft, state)))
	}
	if state.Research.CurrentTech != "" {
		sb.WriteString(fmt.Sprintf("\n--- Research ---\n  %s: %d/%d ticks\n",
			state.Research.CurrentTechName,
			state.Research.TotalTicks-state.Research.TicksLeft,
			state.Research.TotalTicks))
	}

	// All log entries
	sb.WriteString(fmt.Sprintf("\n=== Log Entries (%d) ===\n", len(logs)))
	for _, entry := range logs {
		sb.WriteString(fmt.Sprintf("T%-5d [%-7s] %s\n", entry.Tick, entry.Type, entry.Message))
	}

	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return CommandResult{Message: "Could not write the dump: " + shortIOError(err) + ".", Type: "error"}
	}

	return CommandResult{
		Message: fmt.Sprintf("Logs exported to %s (%d entries)", filename, len(logs)),
		Type:    "info",
	}
}

// gatherDefaultYield is what one bare `gather` brings in; gatherMaxYield is
// the per-use cap.
const (
	gatherDefaultYield = 3.0
	gatherMaxYield     = 25.0
)

func cmdGather(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: usageFor("gather") + " (max " + textfmt.Number(gatherMaxYield) + " per use)", Type: "error"}
	}
	resource := strings.ToLower(args[0])
	if resource != "food" && resource != "wood" && resource != "stone" {
		return CommandResult{Message: "You can only gather food, wood or stone by hand.", Type: "error"}
	}
	amount := gatherDefaultYield
	if len(args) >= 2 {
		n, err := parseAmount(args[1])
		if err != nil {
			return usageError(usageFor("gather")+" (max "+textfmt.Number(gatherMaxYield)+" per use)", err)
		}
		amount = n
	}
	if amount > gatherMaxYield {
		amount = gatherMaxYield
	}
	if _, err := engine.GatherResource(resource, amount); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs what was gathered
}

func cmdBuild(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		// Show available buildings
		state := engine.GetState()
		var lines []string
		lines = append(lines, "[gold]Available buildings:[-]")
		for _, key := range sortedKeysOf(state.Buildings) {
			b := state.Buildings[key]
			if !b.Unlocked {
				continue
			}
			affordable := ""
			if b.CanBuild {
				affordable = "[green]✓[-]"
			} else {
				affordable = "[red]✗[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] (%d built) - Cost: %s %s",
				affordable, key, b.Count, FormatCost(b.NextCost), b.Description))
		}
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}
	key := strings.ToLower(args[0])

	// "build <key> [count|max]" — passing 10000 as the count is a sentinel
	// that tells BuildMultiple "keep building until you can't afford it or
	// hit the MaxCount limit". BuildMultiple returns the actual count built.
	if len(args) >= 2 {
		countArg := strings.ToLower(args[1])
		count := 10000 // "max": BuildMultiple will stop when resources run out or max is hit
		if countArg != "max" {
			n, err := parseCount(countArg)
			if err != nil {
				return usageError(usageFor("build"), err)
			}
			count = n
		}
		if _, err := engine.BuildMultiple(key, count); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs each build
	}

	if err := engine.BuildBuilding(key); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs the build
}

func cmdRecruit(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		if err := engine.RecruitWorker("worker", 1); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the recruits
	}

	arg := strings.ToLower(args[0])
	if arg == "max" {
		if _, err := engine.RecruitMax("worker"); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"}
	}

	n, err := parseCount(arg)
	if err != nil {
		return usageError(usageFor("recruit")+". New workers start idle; put them to work with 'assign <building>'.", err)
	}
	if err := engine.RecruitWorker("worker", n); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"}
}

func cmdAssign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: usageFor("assign"), Type: "error"}
	}
	building := strings.ToLower(args[0])
	if len(args) >= 2 && strings.ToLower(args[1]) == "all" {
		if _, err := engine.AssignAll(building); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the assignment
	}
	count := 1
	if len(args) >= 2 {
		n, err := parseCount(args[1])
		if err != nil {
			return usageError(usageFor("assign"), err)
		}
		count = n
	}
	if err := engine.AssignWorker(building, count); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"}
}

func cmdUnassign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: usageFor("unassign"), Type: "error"}
	}
	building := strings.ToLower(args[0])
	if len(args) >= 2 && strings.ToLower(args[1]) == "all" {
		if _, err := engine.UnassignAll(building); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the change
	}
	count := 1
	if len(args) >= 2 {
		n, err := parseCount(args[1])
		if err != nil {
			return usageError(usageFor("unassign"), err)
		}
		count = n
	}
	if err := engine.UnassignWorker(building, count); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"}
}

func cmdStatus(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string

	lines = append(lines, fmt.Sprintf("[gold]Age:[-] %s  [gold]Game time:[-] %s", state.AgeName, formatTicks(state.Tick, state)))
	lines = append(lines, "")

	// Resources
	lines = append(lines, "[gold]Resources:[-]")
	for _, key := range sortedKeysOf(state.Resources) {
		rs := state.Resources[key]
		if !rs.Unlocked {
			continue
		}
		bar := ProgressBar(rs.Amount, rs.Storage, 15)
		lines = append(lines, fmt.Sprintf("  %-10s %s/%s %s %s",
			rs.Name, FormatNumber(rs.Amount), FormatNumber(rs.Storage), FormatRateTick(rs.Rate), bar))
	}
	lines = append(lines, "")

	// Population
	v := state.Workers
	lines = append(lines, fmt.Sprintf("[gold]Population:[-] %d/%d (idle: %d, food drain: %s/tick)",
		v.TotalPop, v.MaxPop, v.TotalIdle, textfmt.Number(v.FoodDrain)))
	for _, vt := range v.Types {
		if !vt.Unlocked {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %-10s %d (idle: %d)", vt.Name, vt.Count, vt.IdleCount))
		for _, building := range sortedKeysOf(vt.Assignments) {
			if count := vt.Assignments[building]; count > 0 {
				lines = append(lines, fmt.Sprintf("    → %s: %d", game.BuildingName(building), count))
			}
		}
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// shortAccountID returns a human-friendly short form of a 32-char hex account ID:
// the first 8 hex chars (enough to recognize an account at a glance). Falls back to
// the whole string when it's shorter than 8 chars.
func shortAccountID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// cmdAccount handles the `account` command family (the accounts design §9 Phase 4):
//
//	account                 → show the short ID + recovery code + honest "identity,
//	                          not progress" copy.
//	account switch <name>   → make another local account the live one. In a game this
//	                          saves the game to its own account and returns to the menu.
//	account import <path>   → restore a backup into its own account's slot. Never
//	                          switches accounts.
//	account recover <code>  → restore the identity in a recovery code and switch to it.
//	                          Guarded: if the current account holds any progress,
//	                          require `account recover <code> confirm` first.
//
// It talks to the engine's account methods directly — no GameState snapshot needed.
func cmdAccount(args []string, engine *game.GameEngine) CommandResult {
	acct := engine.Account()

	if len(args) == 0 {
		if acct == nil {
			return CommandResult{
				Message: "Accounts are unavailable (no account is loaded).",
				Type:    "warning",
			}
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("[gold]Account:[-] %s  (%s)", acct.Name(), shortAccountID(acct.AccountID)))
		lines = append(lines, fmt.Sprintf("[gold]Recovery code:[-] %s", acct.RecoveryCode()))
		lines = append(lines, "")
		lines = append(lines, recoveryCodeNote...)
		lines = append(lines, "It is not a password: it only says which account you are. Write it down.")
		lines = append(lines, "")
		lines = append(lines, "Restore your account ID on another machine with:  account recover <code>")
		lines = append(lines, "")
		lines = append(lines, "[gold]Progress backups:[-] each account is its own slot, and an import adds or")
		lines = append(lines, "restores that account next to your others. Switch accounts in the [gold]Accounts[-] panel on the main menu.")
		for _, path := range []string{"account list", "account switch", "account export", "account backup", "account import"} {
			u := helpRow(path)
			lines = append(lines, fmt.Sprintf("  %-33s %s", u.Form, u.Text))
		}
		lines = append(lines, "")
		lines = append(lines, "[red]Wipe account[-] (delete an account's ID, unlocks, stats and saves for good)")
		lines = append(lines, "lives in the [gold]Accounts[-] panel on the main menu, behind a type-your-name confirmation.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}

	sub := strings.ToLower(args[0])
	switch sub {
	case "list":
		// List every local account slot: name, short id, an active marker, and highest age.
		summaries := engine.ListAccounts()
		if len(summaries) == 0 {
			return CommandResult{Message: "No local accounts found.", Type: "info"}
		}
		var lines []string
		lines = append(lines, "[gold]Local accounts:[-]")
		for _, s := range summaries {
			marker := "  "
			if s.Active {
				marker = "[label]●[-] "
			}
			name := s.DisplayName
			if strings.TrimSpace(name) == "" {
				name = "(unnamed)"
			}
			age := s.HighestAge
			if age == "" {
				age = "none yet"
			}
			tampered := ""
			if s.Tampered {
				tampered = "  [red]⚠ modified[-]"
			}
			lines = append(lines, fmt.Sprintf("%s%s  [gray](%s)[-]  [gray]age:[-] %s%s",
				marker, lit(name), shortAccountID(s.AccountID), age, tampered))
		}
		lines = append(lines, "")
		lines = append(lines, "Switch with:  account switch <name>   (or use the Accounts panel on the main menu)")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}

	case "switch":
		if len(args) < 2 {
			return CommandResult{Message: usageFor("account switch"), Type: "error"}
		}
		// Resolve name→id via the shared derivation, then switch only if that slot exists.
		// A non-existent slot errors with guidance rather than minting an empty account
		// (use the Accounts panel, or `account import`, to create/restore one).
		name := strings.TrimSpace(strings.Join(args[1:], " "))
		id := game.AccountIDForName(name)
		if !accountListed(engine, id) {
			return CommandResult{
				Message: fmt.Sprintf("No account named %q. Create it from the Accounts panel on the main menu, or type 'account list' to see yours.", lit(name)),
				Type:    "error",
			}
		}
		if acct != nil && acct.AccountID == id {
			return CommandResult{Message: fmt.Sprintf("You are already playing as %s.", lit(acct.Name())), Type: "info"}
		}
		// The game in progress belongs to the account in use: SwitchAccount saves it to
		// that account and stops it before switching (endedRun), and the dashboard then
		// goes to the main menu.
		saveName := engine.ActiveSaveName()
		endedRun, err := engine.SwitchAccount(id)
		if err != nil {
			return CommandResult{Message: "Could not switch accounts: " + textfmt.Sentence(err.Error()), Type: "error", ToMenu: endedRun}
		}
		// Re-resolve the active theme against the now-active account so the UI doesn't keep
		// the prior account's theme after the swap (the theming design §6).
		applyAccountTheme(engine)
		// Info, not success: the engine logs nothing here, and the dashboard
		// drops success replies.
		return CommandResult{
			Message: switchedReply(acct, engine.Account(), id, saveName, endedRun),
			Type:    "info",
			ToMenu:  endedRun,
		}

	case "export":
		if acct == nil {
			return CommandResult{
				Message: "Accounts are unavailable (no account is loaded).",
				Type:    "warning",
			}
		}
		blob, err := acct.ExportProgress()
		if err != nil {
			return CommandResult{Message: "Export failed: " + shortIOError(err), Type: "error"}
		}
		// Resolve the destination: explicit path arg, or a default that names the account so
		// multiple accounts' exports don't collide in a shared directory. The blob carries the
		// id regardless; the filename is just a convenience. Defaults beside the live account.
		var path string
		if len(args) >= 2 && args[1] != "" {
			path = args[1]
		} else {
			path = filepath.Join(game.DataDir(), fmt.Sprintf("account-%s-export.json", shortAccountID(acct.AccountID)))
		}
		if dir := filepath.Dir(path); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return CommandResult{Message: "Export failed: " + shortIOError(err), Type: "error"}
			}
		}
		if err := os.WriteFile(path, blob, 0644); err != nil {
			return CommandResult{Message: "Export failed: " + shortIOError(err), Type: "error"}
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("Progress exported to %s. Restore it with: account import %s", path, path))
		lines = append(lines, "This file backs up your progress (unlocks, stats, achievements). Your recovery")
		lines = append(lines, "code is separate and restores only your account ID.")
		// Also take a full slot snapshot (account.json + saves/). A backup failure must not fail
		// the export, so it only adds a line when it works.
		if backupPath, bErr := engine.BackupAccount(acct.AccountID); bErr == nil {
			lines = append(lines, fmt.Sprintf("Full backup (account.json and saves) written to %s.", backupPath))
		}
		// Info, not success: the path must reach the log, and the engine logs nothing here.
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}

	case "backup":
		// A full snapshot of the active account's slot (account.json + saves/) into
		// <root>/backups/. Distinct from export, which serializes only meta-progression.
		if acct == nil {
			return CommandResult{Message: "No account to back up.", Type: "warning"}
		}
		backupPath, err := engine.BackupAccount(acct.AccountID)
		if err != nil {
			return CommandResult{Message: "Backup failed: " + shortIOError(err), Type: "error"}
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("Full backup saved to %s.", backupPath))
		lines = append(lines, "It holds this account's account.json and every save in its slot. To restore,")
		lines = append(lines, "copy the folder's contents back into data/accounts/<id>/. The 10 most recent backups per account are kept.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}

	case "import":
		if acct == nil {
			return CommandResult{
				Message: "Accounts are unavailable (no account is loaded).",
				Type:    "warning",
			}
		}
		if len(args) < 2 {
			return CommandResult{Message: usageFor("account import"), Type: "error"}
		}
		path := args[1]
		// merge by default; the `replace` token switches to wholesale replacement.
		merge := !(len(args) >= 3 && strings.EqualFold(args[2], "replace"))
		blob, err := os.ReadFile(path)
		if err != nil {
			return CommandResult{Message: fmt.Sprintf("Import failed: cannot read %s (%s).", path, shortIOError(err)), Type: "error"}
		}
		// An export is a single-account backup: it lands in its own account's slot (keyed by
		// the blob's account id). A backup of the account in use is folded into it. Import
		// never switches accounts, so the game in progress carries on under the account it
		// belongs to; switching is its own step.
		imported, err := engine.ImportAccountExport(blob, merge)
		if err != nil {
			return CommandResult{Message: "Import failed: " + textfmt.Sentence(err.Error()), Type: "error"}
		}
		mode := "merged"
		if !merge {
			mode = "replaced"
		}
		name := imported.Name()
		if name == "" {
			name = "(unnamed)"
		}
		themes := textfmt.Count(len(imported.UnlockedThemes()), "theme", "themes")
		if imported.AccountID == acct.AccountID {
			// The backup may carry a theme choice (replace, or a first theme): re-apply it.
			applyAccountTheme(engine)
			return CommandResult{
				Message: fmt.Sprintf("Imported the backup into %q (%s), the account you are using: %s, progress %s.",
					lit(name), shortAccountID(imported.AccountID), themes, mode),
				Type: "info",
			}
		}
		return CommandResult{
			Message: fmt.Sprintf("Imported account %q (%s): %s, progress %s. %s",
				lit(name), shortAccountID(imported.AccountID), themes, mode, switchHint(imported)),
			Type: "info",
		}

	case "recover":
		if len(args) < 2 {
			return CommandResult{Message: usageFor("account recover"), Type: "error"}
		}
		code := args[1]
		confirmed := len(args) >= 3 && strings.EqualFold(args[2], "confirm")

		// Check the code before anything else, so a typo is refused before any question
		// is asked or anything changes.
		id, err := game.RecoveryCodeID(code)
		if err != nil {
			return errorResult(err)
		}
		if acct != nil && acct.AccountID == id {
			return CommandResult{
				Message: fmt.Sprintf("That is the recovery code of the account you are using (%s), so there is nothing to restore.", accountLabel(acct)),
				Type:    "info",
			}
		}
		// Confirm guard: recovering switches this machine to another account. The account
		// in use keeps everything in its own slot, but ask first whenever it holds any
		// progress at all (theme unlocks, achievements or lifetime stats), and say what.
		if acct != nil && !confirmed {
			if held := accountHoldings(acct); held != "" {
				var lines []string
				lines = append(lines, fmt.Sprintf("[gold]Recovering switches accounts.[-] The code is for a different account from %s, the one you are using.", accountLabel(acct)))
				lines = append(lines, fmt.Sprintf("%s keeps what it holds (%s) in its own slot. %s", accountLabel(acct), held, switchHint(acct)))
				lines = append(lines, recoveryCodeNote...)
				lines = append(lines, "A game in progress is saved to its own account first, and you go back to the main menu.")
				lines = append(lines, "")
				lines = append(lines, fmt.Sprintf("To go ahead:  account recover %s confirm", lit(code)))
				return CommandResult{Message: strings.Join(lines, "\n"), Type: "warning"}
			}
		}

		// The code's account lands in its own slot (an account already there is opened, never
		// overwritten), then the switch works as account switch's does.
		existed := accountListed(engine, id)
		saveName := engine.ActiveSaveName()
		restored, endedRun, err := engine.RecoverAccount(code)
		if err != nil {
			return CommandResult{Message: "Could not recover the account: " + textfmt.Sentence(err.Error()), Type: "error", ToMenu: endedRun}
		}
		// Re-resolve the active theme against the now-installed account so the UI
		// doesn't keep the prior account's theme after an identity swap (the theming design
		// §6). A recovery code carries identity only, so a fresh restore resolves to
		// Forge; a restore of an account with a stored theme honors it.
		applyAccountTheme(engine)
		msg := fmt.Sprintf("Account ID restored: %s. Import a progress backup with 'account import <path>' to bring back unlocks and stats.", shortAccountID(restored.AccountID))
		if existed {
			msg = fmt.Sprintf("That account was already on this machine, with its progress. Now playing as %s.", accountWithID(restored))
		}
		if endedRun {
			msg = savedRunLine(acct, saveName) + " " + msg + " " + menuNextStep
		}
		return CommandResult{Message: msg, Type: "info", ToMenu: endedRun}

	case "wipe":
		// The destructive wipe lives behind the Accounts panel's type-your-name gate; we
		// deliberately do not wipe from a bare command. Direct the player there.
		return CommandResult{
			Message: "Wiping an account is permanent, so it lives in the Accounts panel on the main menu (press Esc, choose Accounts, then press w on the account). It deletes that account's ID, theme unlocks, lifetime stats, achievements and every save in its slot. A backup goes to data/backups/ first; restoring it is manual.",
			Type:    "warning",
		}

	default:
		return CommandResult{Message: subUsage("account"), Type: "error"}
	}
}

// recoveryCodeNote says what a recovery code does and does not restore.
var recoveryCodeNote = []string{
	"This code restores your account ID on another machine. It does not restore",
	"progress (unlocks, stats, achievements). Back those up with account export.",
}

// menuNextStep ends the reply to an account change that ended a game in progress.
const menuNextStep = "Load one of this account's games or start a new one."

// accountListed reports whether a local account slot holds the account with this ID.
func accountListed(engine *game.GameEngine, id string) bool {
	for _, s := range engine.ListAccounts() {
		if s.AccountID == id {
			return true
		}
	}
	return false
}

// accountLabel names an account in a reply: its display name, or its short ID when it
// has none (an identity restored from a recovery code). Escaped for tview.
func accountLabel(acct *game.Account) string {
	if name := strings.TrimSpace(acct.Name()); name != "" {
		return lit(name)
	}
	return shortAccountID(acct.AccountID)
}

// accountWithID is accountLabel plus the short ID in parentheses, when the label is a name.
func accountWithID(acct *game.Account) string {
	if strings.TrimSpace(acct.Name()) == "" {
		return shortAccountID(acct.AccountID)
	}
	return fmt.Sprintf("%s (%s)", accountLabel(acct), shortAccountID(acct.AccountID))
}

// switchHint says how to switch to an account: by name when its ID comes from that name,
// otherwise from the Accounts panel (an identity restored from a recovery code has no
// name to type).
func switchHint(acct *game.Account) string {
	name := strings.TrimSpace(acct.Name())
	if name != "" && game.AccountIDForName(name) == acct.AccountID {
		return fmt.Sprintf("Switch to it any time with: account switch %s", lit(name))
	}
	return "Switch to it any time from the Accounts panel on the main menu."
}

// accountHoldings lists the progress an account holds, for the recover guard: theme
// unlocks, achievements and every lifetime stat. "" when it holds none.
func accountHoldings(acct *game.Account) string {
	stats, achievements := acct.LifetimeStats()
	var parts []string
	if n := len(acct.UnlockedThemes()); n > 0 {
		parts = append(parts, textfmt.Count(n, "theme unlock", "theme unlocks"))
	}
	if n := len(achievements); n > 0 {
		parts = append(parts, textfmt.Count(n, "achievement", "achievements"))
	}
	if n := stats.TotalPrestiges; n > 0 {
		parts = append(parts, textfmt.Count(n, "prestige", "prestiges"))
	}
	if stats.HighestAge != "" {
		parts = append(parts, "highest age "+game.AgeName(stats.HighestAge))
	}
	if n := stats.CivilizationsStarted; n > 0 {
		parts = append(parts, textfmt.Count(n, "civilization started", "civilizations started"))
	}
	if n := stats.SavesCompleted; n > 0 {
		parts = append(parts, textfmt.Count(n, "save completed", "saves completed"))
	}
	return strings.Join(parts, ", ")
}

// savedRunLine says where the game an account change ended was saved: to the account it
// belongs to (prev, the account in use before the change).
func savedRunLine(prev *game.Account, saveName string) string {
	if prev == nil {
		return fmt.Sprintf("Saved your game %s.", lit(saveName))
	}
	return fmt.Sprintf("Saved your game %s to the account %s.", lit(saveName), accountLabel(prev))
}

// switchedReply is the reply to a successful account switch from prev to now. When the
// switch ended a game in progress it says where that game was saved and what to do next.
func switchedReply(prev, now *game.Account, id, saveName string, endedRun bool) string {
	who := shortAccountID(id)
	if now != nil {
		who = accountWithID(now)
	}
	msg := fmt.Sprintf("Now playing as %s.", who)
	if endedRun {
		msg = savedRunLine(prev, saveName) + " " + msg + " " + menuNextStep
	}
	return msg
}

// shortIOError is an I/O error without the file path the OS puts in it:
// "permission denied", "no such file or directory".
func shortIOError(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err.Error()
	}
	return err.Error()
}

// oneWordName refuses a save name typed as several words: `save my run`
// used to save as "my".
func oneWordName(verb string, args []string) (string, *CommandResult) {
	if len(args) > 1 {
		return "", &CommandResult{
			Message: fmt.Sprintf("Save names are one word (try '%s %s').", verb, strings.Join(args, "_")),
			Type:    "error",
		}
	}
	return args[0], nil
}

func cmdSave(args []string, engine *game.GameEngine) CommandResult {
	// With a name: branch a new save off the current run (BranchSave switches the
	// active slot + sets the parent, so autosave follows the branch). Without a
	// name: overwrite the active slot. The dashboard UI intercepts a bare `save`
	// to pop the Overwrite/Branch modal, but this path stays sane for tests and
	// other callers.
	if len(args) > 0 {
		name, refused := oneWordName("save", args)
		if refused != nil {
			return *refused
		}
		if err := engine.BranchSave(name); err != nil {
			return errorResult(err)
		}
		return CommandResult{Message: fmt.Sprintf("Branched a new save '%s'. Autosave now follows it.", name), Type: "info"}
	}
	active := engine.ActiveSaveName()
	if err := engine.SaveGame(active); err != nil {
		return CommandResult{Message: "Could not save: " + shortIOError(err) + ".", Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Saved to '%s'.", active), Type: "info"}
}

func cmdLoad(args []string, engine *game.GameEngine) CommandResult {
	// No name: do NOT silently load autosave. Guide the player to the browser.
	// The dashboard intercepts a bare `load` to open the Load Game tree; this
	// fallback keeps other callers (and tests) from loading a slot by surprise.
	if len(args) == 0 {
		return CommandResult{Message: "Type 'load <name>' to load a save, or open Load Game from the menu (press Esc) to browse your save tree.", Type: "info"}
	}
	name, refused := oneWordName("load", args)
	if refused != nil {
		return *refused
	}
	if err := engine.LoadGame(name); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return CommandResult{Message: fmt.Sprintf("No save named '%s'. Type saves to list them.", name), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Could not load '%s': %s.", name, strings.TrimRight(shortIOError(err), ".")), Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Game loaded from '%s'.", name), Type: "info"}
}

func cmdRates(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Resource Rate Breakdown:[-]")

	for _, key := range sortedKeysOf(state.Resources) {
		rs := state.Resources[key]
		if !rs.Unlocked || (rs.Rate == 0 && rs.Breakdown == (game.RateBreakdown{})) {
			continue
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-]:  %s", rs.Name, FormatRateTick(rs.Rate)))
		b := rs.Breakdown
		var parts []string
		if b.BuildingRate != 0 {
			parts = append(parts, fmt.Sprintf("Buildings: %s", textfmt.RateValue(b.BuildingRate)))
		}
		if b.WorkerRate != 0 {
			// What worker output bonuses add (game/engine.go, gather_rate).
			parts = append(parts, fmt.Sprintf("Worker bonus: %s", textfmt.RateValue(b.WorkerRate)))
		}
		if b.ResearchRate != 0 {
			// What the techs add: their bonus on what the resource's
			// buildings make, counted after the caps, and the flat output
			// of a tech that is its first source.
			parts = append(parts, fmt.Sprintf("Research: %s", textfmt.RateValue(b.ResearchRate)))
		}
		if b.EventRate != 0 {
			parts = append(parts, fmt.Sprintf("Events: %s", textfmt.RateValue(b.EventRate)))
		}
		if b.TradeRate != 0 {
			parts = append(parts, fmt.Sprintf("Trade: %s", textfmt.RateValue(b.TradeRate)))
		}
		if b.BonusRate != 0 {
			parts = append(parts, fmt.Sprintf("Bonuses: %s", textfmt.RateValue(b.BonusRate)))
		}
		if b.LegacyRate != 0 {
			parts = append(parts, fmt.Sprintf("Cosmic Legacy: %s", textfmt.RateValue(b.LegacyRate)))
		}
		if b.MasteryRate != 0 {
			parts = append(parts, fmt.Sprintf("Era Mastery: %s", textfmt.RateValue(b.MasteryRate)))
		}
		if b.FoodDrain != 0 {
			parts = append(parts, fmt.Sprintf("Drain: %s", textfmt.RateValue(b.FoodDrain)))
		}
		if len(parts) > 0 {
			lines = append(lines, fmt.Sprintf("    %s", strings.Join(parts, "  ")))
		}
	}

	if len(lines) == 1 {
		lines = append(lines, "  [gray]No active resource rates[-]")
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdSaveList() CommandResult {
	saves, err := game.ListSaveDetails()
	if err != nil {
		return CommandResult{Message: "Could not list saves: " + shortIOError(err) + ".", Type: "error"}
	}
	if len(saves) == 0 {
		return CommandResult{Message: "No save files found.", Type: "info"}
	}
	var lines []string
	lines = append(lines, "[gold]Save Files:[-]")
	for _, s := range saves {
		age := s.Age
		if age == "" {
			age = "unknown"
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%-15s[-] %s  [gray](%s)[-]",
			s.Name, s.Timestamp.Format("2006-01-02 15:04:05"), age))
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// cmdTheme handles the `theme` command's text paths:
//   - `theme`        → directs the player to the picker (the interactive picker
//     page is opened by the dashboard intercept; this fallback covers the
//     command-table path and tests, which can't open an interactive page).
//   - `theme list`   → lists every theme (name, key, active marker, accessible note).
//   - `theme <key>`  → switches directly to a theme by key if it exists, else an
//     error listing the valid keys.
//
// theme.SetActive applies the name-remap + restyle; the dashboard redraws on its
// next tick, so the switch shows up live without an explicit Draw here.
//
// engine bridges to the account layer (the theming design §6): on a successful switch we
// persist the new active theme account-wide so a CLI switch survives saves, and the
// theme is gated through themeAvailable so locked flavor themes (Phase 3) refuse.
// engine/Account() are nil-guarded — the game must run accountless.
func cmdTheme(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{
			Message: "Usage: theme list | theme <key>. Type 'theme' at the prompt for the live picker.",
			Type:    "info",
		}
	}
	if strings.ToLower(args[0]) == "list" {
		var st *game.GameState
		if engine != nil {
			s := engine.GetState()
			st = &s
		}
		return cmdThemeList(themeAccount(engine), st)
	}

	key := strings.ToLower(args[0])
	t, ok := theme.ByKey(key)
	if !ok {
		return CommandResult{
			Message: fmt.Sprintf("Unknown theme '%s'. Themes: %s.", key, strings.Join(themeKeys(), ", ")),
			Type:    "error",
		}
	}
	// Unlock gate (the theming design §4/§5): refuse a known-but-locked theme. No theme is
	// locked today (all shipped themes are Accessible/Forge); this is the Phase-3 seam.
	if !themeAvailable(themeAccount(engine), t) {
		return CommandResult{Message: themeUnavailableMsg, Type: "error"}
	}
	if err := theme.SetActive(t.Key); err != nil {
		return errorResult(err)
	}
	theme.Restyle()
	// Persist account-wide so a CLI switch survives saves (the theming design §6). Nil-guarded;
	// the Save error is non-fatal — the theme is already applied, so we report success
	// regardless rather than rolling back a working visual change over a write hiccup.
	if acct := themeAccount(engine); acct != nil {
		_ = acct.SetActiveTheme(t.Key)
	}
	// Info, not success: the engine logs nothing for a theme switch.
	return CommandResult{Message: fmt.Sprintf("Theme set to %s.", t.Name), Type: "info"}
}

// themeAccount returns engine's account, or nil when accountless. Mirrors the
// picker's account() guard so the command + picker share one nil-safe accessor.
func themeAccount(engine *game.GameEngine) *game.Account {
	if engine == nil {
		return nil
	}
	return engine.Account()
}

// cmdThemeList renders the `theme list` output: one line per theme with its name,
// key, an active marker, and a status note. Unlocked themes show "(accessible)" when
// applicable; a LOCKED flavor theme shows "🔒 <unlock hint>" instead, so the player
// sees exactly how to earn it (the theming design §5/§7), e.g. "monochrome  🔒 Reach the
// Information Age".
//
// acct may be nil (accountless play / tests): with no account only the always-
// available set (Accessible + Forge) is unlocked, so the flavor themes correctly
// render as locked with their hints. st, when set, keeps a hint from naming an
// age the player cannot see yet (themeUnlockHint); nil shows the hints as written.
func cmdThemeList(acct *game.Account, st *game.GameState) CommandResult {
	activeKey := theme.Active().Key
	var lines []string
	lines = append(lines, "[gold]Themes:[-]")
	for _, t := range theme.All() {
		marker := "  "
		if t.Key == activeKey {
			marker = "[gold]●[-] "
		}
		note := ""
		switch {
		case !themeAvailable(acct, t):
			// Locked flavor theme: show the unlock condition rather than nothing.
			hint := t.UnlockHint
			if st != nil {
				hint = themeUnlockHint(t, *st)
			}
			if hint == "" {
				hint = "unlock via a milestone"
			}
			note = fmt.Sprintf("  [red]🔒[-] [gray]%s[-]", hint)
		case t.Accessible:
			note = "  [cyan](accessible)[-]"
		}
		lines = append(lines, fmt.Sprintf("%s[white]%-20s[-] [gray]%-20s[-] [dim]%-5s[-]%s",
			marker, t.Name, t.Key, strings.ToLower(t.Variant()), note))
	}
	lines = append(lines, "[gray]Switch with 'theme <key>'.[-]")
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// themeKeys returns every registered theme key in display order, for usage/error
// messages and autocomplete.
func themeKeys() []string {
	all := theme.All()
	keys := make([]string, 0, len(all))
	for _, t := range all {
		keys = append(keys, t.Key)
	}
	return keys
}

func cmdResearch(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return cmdResearchList(engine)
	}
	subcmd := strings.ToLower(args[0])

	if subcmd == "list" {
		return cmdResearchList(engine)
	}
	if subcmd == "tree" {
		switch zoom := strings.ToLower(strings.Join(args[1:], " ")); zoom {
		case "", "close", "far":
			return CommandResult{OverlayName: "techs", ResearchZoom: zoom}
		}
		return CommandResult{Message: usageFor("research tree"), Type: "error"}
	}
	if subcmd == "card" {
		// The card names a tech, so it opens only on one the player may see.
		state := engine.GetState()
		key := strings.ToLower(strings.Join(args[1:], "_"))
		if def, ok := state.Ruleset().Tech(key); !ok || !game.SightOf(&state).Age(def.Age) {
			return CommandResult{Message: "No tech '" + strings.Join(args[1:], " ") + "' in sight. Type research to see the tree.", Type: "error"}
		}
		return CommandResult{OverlayName: "techs", ResearchCard: key}
	}
	if subcmd == "cancel" {
		if err := engine.CancelResearch(); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the cancellation
	}

	// Support multi-word keys entered with spaces by joining all remaining args
	// with underscores (e.g. "research bronze working" → "bronze_working").
	techKey := strings.ToLower(strings.Join(args, "_"))
	if err := engine.StartResearch(techKey); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs the start
}

func cmdResearchList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available techs:[-]")

	for _, key := range sortedKeysOf(state.Research.Techs) {
		ts := state.Research.Techs[key]
		if !ts.Available {
			continue
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-] - %s (%s)%s", key, ts.Name, game.Amount(ts.Cost, "knowledge"), keystoneMark(ts)))
	}

	if state.Research.CurrentTech != "" {
		lines = append(lines, fmt.Sprintf("\n[yellow]Currently researching: %s (%s left)[-]",
			state.Research.CurrentTechName, formatTicks(state.Research.TicksLeft, state)))
	}

	if len(lines) == 1 {
		lines = append(lines, "  [gray]No techs to research right now.[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// cmdExpedition handles the `expedition`/`exp` command — the civilian SCOUTING
// surface. No args opens the Expeditions panel; "list" prints the scouting list;
// a key launches a scouting expedition. Military keys are redirected to the
// `campaign` command rather than launched here.
func cmdExpedition(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{OverlayName: "expedition"}
	}
	subcmd := strings.ToLower(args[0])
	if subcmd == "list" {
		return cmdScoutingList(engine)
	}

	expKey := strings.ToLower(strings.Join(args, "_"))

	// Reject military keys here — they belong to `campaign`.
	if def := engine.Military.ExpeditionDefByKey(expKey); def != nil && def.Category != game.ExpeditionScouting {
		return CommandResult{
			Message: fmt.Sprintf("%s is a military campaign. Wage it with 'campaign %s'.", def.Name, expKey),
			Type:    "info",
		}
	}

	if err := engine.LaunchExpedition(expKey); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs the launch
}

// cmdCampaign handles the `campaign` command — the MILITARY surface. No args or
// "list" prints the campaign list; a key wages a military campaign (costs
// soldiers). Scouting keys are redirected to the `expedition` command.
func cmdCampaign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return cmdCampaignList(engine)
	}
	subcmd := strings.ToLower(args[0])
	if subcmd == "list" {
		return cmdCampaignList(engine)
	}

	expKey := strings.ToLower(strings.Join(args, "_"))

	// Reject scouting keys here — they belong to `expedition`.
	if def := engine.Military.ExpeditionDefByKey(expKey); def != nil && def.Category != game.ExpeditionMilitary {
		return CommandResult{
			Message: fmt.Sprintf("%s is a scouting expedition. Send it with 'expedition %s'.", def.Name, expKey),
			Type:    "info",
		}
	}

	if err := engine.LaunchExpedition(expKey); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs the launch
}

// cmdFestival handles the `festival` culture-sink command. Bare `festival`
// shows status (cost, cooldown, what it does). `festival confirm yes` spends a
// lump of culture to inject a temporary empire-wide production buff. Mirrors the
// prestige confirm UX so the muscle memory carries over.
func cmdFestival(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdFestivalStatus(engine)
	}
	subcmd := strings.ToLower(args[0])

	switch subcmd {
	case "confirm":
		if len(args) >= 2 && strings.ToLower(args[1]) == "yes" {
			if err := engine.DoFestival(); err != nil {
				return errorResult(err)
			}
			return CommandResult{Type: "success"} // the engine logs the festival
		}
		// Show the confirm prompt with the live cost.
		st := engine.FestivalStatus()
		state := engine.GetState()
		if !st.Ready {
			return CommandResult{
				Message: fmt.Sprintf("[yellow]Festival on cooldown.[-] The next one can be held in %s.", formatTicks(st.CooldownLeft, state)),
				Type:    "warning",
			}
		}
		var lines []string
		lines = append(lines, "[gold]Hold a Cultural Festival?[-]")
		lines = append(lines, fmt.Sprintf("  Cost: [cyan]%s[-] (you have %s)", game.Amount(st.Cost, "culture"), textfmt.Number(st.Culture)))
		lines = append(lines, fmt.Sprintf("  Effect: [green]+%.0f%%[-] to all production for [cyan]%s[-].", st.BuffPercent*100, formatTicks(st.BuffTicks, state)))
		lines = append(lines, festivalCapLines(st)...)
		lines = append(lines, fmt.Sprintf("  Cooldown afterward: [cyan]%s[-].", formatTicks(st.CooldownTicks, state)))
		if st.Culture < st.Cost {
			lines = append(lines, "")
			lines = append(lines, "  [red]Not enough culture.[-]")
		}
		lines = append(lines, "")
		lines = append(lines, "  Type [cyan]festival confirm yes[-] to celebrate.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "warning"}
	default:
		return CommandResult{Message: subUsage("festival"), Type: "error"}
	}
}

// festivalCapLines warns, before the culture is spent, that the
// all-production cap would hold back a festival held now: nothing when the
// whole buff would count.
func festivalCapLines(st game.FestivalStatus) []string {
	if st.CapNote == "" {
		return nil
	}
	return []string{"  [yellow]Right now it is " + st.CapNote + ". All production counts up to +200% (see stats).[-]"}
}

// cmdFestivalStatus renders the bare `festival` status panel.
func cmdFestivalStatus(engine *game.GameEngine) CommandResult {
	st := engine.FestivalStatus()
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Cultural Festival[-]")
	lines = append(lines, "  Spend a lump of culture to raise all production for a while.")
	lines = append(lines, fmt.Sprintf("  Cost: [cyan]%s[-]  (you have %s)", game.Amount(st.Cost, "culture"), textfmt.Number(st.Culture)))
	lines = append(lines, fmt.Sprintf("  Effect: [green]+%.0f%%[-] to all production for [cyan]%s[-].", st.BuffPercent*100, formatTicks(st.BuffTicks, state)))
	lines = append(lines, festivalCapLines(st)...)
	if st.Ready {
		if st.Culture >= st.Cost {
			lines = append(lines, "  Status: [green]ready[-]. Type [cyan]festival confirm yes[-].")
		} else {
			lines = append(lines, "  Status: [yellow]not enough culture yet.[-]")
		}
	} else {
		lines = append(lines, fmt.Sprintf("  Status: [yellow]on cooldown[-], %s left.", formatTicks(st.CooldownLeft, state)))
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// cmdBlackMarket handles the `blackmarket` / `trade black` command. Bare form
// shows the status panel (cost, odds, cooldown). `blackmarket <resource>` spends
// a lump of culture on a high-risk/high-reward deal that may pay out a large
// amount of the chosen resource — or vanish with the culture.
func cmdBlackMarket(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdBlackMarketStatus(engine)
	}
	resource := strings.ToLower(args[0])

	if _, _, err := engine.DoBlackMarket(resource); err != nil {
		return errorResult(err)
	}
	// The engine logs the run, paid off or failed; no second line here.
	return CommandResult{Type: "success"}
}

// cmdBlackMarketStatus renders the bare `blackmarket` status panel.
func cmdBlackMarketStatus(engine *game.GameEngine) CommandResult {
	st := engine.BlackMarketStatus()
	var lines []string
	lines = append(lines, "[gold]Black Market[-]")
	if !st.Available {
		lines = append(lines, "  "+theme.Paint(theme.RoleDim, "Smuggling networks open in "+ageRef(engine.GetState(), "colonial_age")+"."))
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}
	lines = append(lines, "  Spend culture on a smuggling run: a gamble on a big haul of one resource.")
	lines = append(lines, fmt.Sprintf("  Cost: [cyan]%s[-] per run  (you have %s)", game.Amount(st.Cost, "culture"), textfmt.Number(st.Culture)))
	lines = append(lines, fmt.Sprintf("  Odds: [green]%.0f%%[-] chance of a haul worth [green]%.1fx[-] the culture; otherwise the culture is lost.", st.WinChance*100, st.WinMult))
	if st.Ready {
		if st.Culture >= st.Cost {
			lines = append(lines, "  Status: [green]ready[-]. Type [cyan]blackmarket <resource>[-] (for example, blackmarket gold).")
		} else {
			lines = append(lines, "  Status: [yellow]not enough culture yet.[-]")
		}
	} else {
		lines = append(lines, fmt.Sprintf("  Status: [yellow]lying low[-]. The next run is possible in %s.", formatTicks(st.CooldownLeft, engine.GetState())))
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdPrestige(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdPrestigeStatus(engine)
	}
	subcmd := strings.ToLower(args[0])

	switch subcmd {
	case "confirm":
		// Require "prestige confirm yes" to actually execute
		if len(args) >= 2 && strings.ToLower(args[1]) == "yes" {
			if err := engine.DoPrestige(); err != nil {
				return errorResult(err)
			}
			if engine.GetState().LastPassage.Pending {
				return CommandResult{
					Message:         "☄ The Last Passage has come. Prestige waits until you choose Endure or Succumb.",
					Type:            "warning",
					OpenCatastrophe: true,
				}
			}
			return CommandResult{Type: "success"} // the engine logs the prestige
		}
		// Show warning
		state := engine.GetState()
		p := state.Prestige
		var lines []string
		lines = append(lines, "[yellow]⚠ Prestige warning[-]")
		lines = append(lines, fmt.Sprintf("  You will earn [cyan]%d[-] prestige points.", p.PendingPoints))
		// An early prestige pays little: say so before the player confirms.
		if early := game.EarlyPrestigeLine(game.SightOf(&state), state.Age, p.PendingPoints, false); early != "" {
			lines = append(lines, "  "+early)
		}
		lines = append(lines, "  [red]All progress is reset:[-] resources, buildings, workers, research and military.")
		lines = append(lines, "  Kept: prestige points, the legacy kit and what it remembers, and Era Mastery.")
		lines = append(lines, lastPassageWarningLines(state)...)
		lines = append(lines, "")
		lines = append(lines, "  Type [cyan]prestige confirm yes[-] to proceed.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "warning"}
	case "shop":
		return cmdPrestigeShop(engine)
	case "buy":
		if len(args) < 2 {
			return CommandResult{Message: usageFor("prestige buy"), Type: "error"}
		}
		key := strings.ToLower(strings.Join(args[1:], "_"))
		if err := engine.BuyPrestigeUpgrade(key); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the purchase
	default:
		return CommandResult{Message: subUsage("prestige"), Type: "error"}
	}
}

func cmdPrestigeStatus(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	p := state.Prestige
	var lines []string

	lines = append(lines, "[gold]Prestige Status[-]")
	lines = append(lines, fmt.Sprintf("  Level: [cyan]%d[-]", p.Level))
	lines = append(lines, fmt.Sprintf("  Points: [cyan]%d[-] available / [cyan]%d[-] total earned", p.Available, p.TotalEarned))

	lines = append(lines, masteryStatusLines(state)...)

	if state.LastPassage.CosmicLegacy {
		lines = append(lines, fmt.Sprintf("  Cosmic Legacy: [gold]+%.0f%%[-] production (permanent)", game.CosmicLegacyProductionBonus*100))
	}

	lines = append(lines, kitStatusLine(state))
	lines = append(lines, prestigePointsLines(state)...)
	switch {
	case state.LastPassage.Pending:
		lines = append(lines, "\n  [red]☄ The Last Passage has come. Prestige waits for your answer.[-]")
		lines = append(lines, "  Type [cyan]catastrophe[-] to choose Endure or Succumb.")
	case p.CanPrestige:
		lines = append(lines, lastPassageStatusLines(state)...)
		lines = append(lines, "  Type [cyan]prestige confirm[-] to reset with bonuses.")
	}

	lines = append(lines, "\n  Type [cyan]prestige shop[-] to see the legacy kit.")
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdPrestigeShop(engine *game.GameEngine) CommandResult {
	return CommandResult{Message: strings.Join(prestigeShopLines(engine.GetState()), "\n"), Type: "info"}
}

// cmdScoutingList prints only the available SCOUTING expeditions, with the
// active scout (if any) as a footer. This backs `expedition list`.
func cmdScoutingList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available Expeditions:[-]")

	appendExpeditionGroup(&lines, "Scouting", state.Military.Expeditions, game.ExpeditionScouting, state)

	if state.Military.ActiveScout != nil {
		lines = append(lines, fmt.Sprintf("\n[yellow]Active expedition: %s (%s left)[-]",
			state.Military.ActiveScout.Name, formatTicks(state.Military.ActiveScout.TicksLeft, state)))
	}

	if !hasCategory(state.Military.Expeditions, game.ExpeditionScouting) {
		lines = append(lines, "  [gray]No expeditions available yet[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// cmdCampaignList prints only the available MILITARY campaigns, with the active
// campaign (if any) as a footer. This backs `campaign` and `campaign list`.
func cmdCampaignList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available Campaigns:[-]")
	if note := strings.TrimRight(lockNotes(state, config.FeatureCampaigns), "\n"); note != "" {
		lines = append(lines, note)
	}

	appendExpeditionGroup(&lines, "Campaigns", state.Military.Expeditions, game.ExpeditionMilitary, state)

	if state.Military.ActiveMilitary != nil {
		lines = append(lines, fmt.Sprintf("\n[yellow]Active campaign: %s (%s left)[-]",
			state.Military.ActiveMilitary.Name, formatTicks(state.Military.ActiveMilitary.TicksLeft, state)))
	}

	if !hasCategory(state.Military.Expeditions, game.ExpeditionMilitary) {
		lines = append(lines, "  [gray]No campaigns available yet[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// hasCategory reports whether any expedition in exps matches the given category.
func hasCategory(exps []game.ExpeditionInfo, category string) bool {
	for _, exp := range exps {
		if exp.Category == category {
			return true
		}
	}
	return false
}

// appendExpeditionGroup appends the subset of exps matching category to lines,
// under a labeled header (e.g. "Scouting"). The header is omitted when no
// expedition matches, so empty subsections produce no output.
func appendExpeditionGroup(lines *[]string, label string, exps []game.ExpeditionInfo, category string, state game.GameState) {
	first := true
	for _, exp := range exps {
		if exp.Category != category {
			continue
		}
		if first {
			*lines = append(*lines, fmt.Sprintf("[yellow]%s:[-]", label))
			first = false
		}
		canStr := "[red]✗[-]"
		if exp.CanLaunch {
			canStr = "[green]✓[-]"
		}
		// Soldier-free scouting expeditions omit the soldier prefix; military
		// campaigns lead with their soldier requirement.
		var reqParts []string
		if exp.SoldiersNeeded > 0 {
			reqParts = append(reqParts, fmt.Sprintf("%d soldiers", exp.SoldiersNeeded))
		}
		if cost := formatExpeditionCost(exp.Cost, state); cost != "" {
			reqParts = append(reqParts, cost)
		}
		// Duration is rolled per launch, so list the def's range, not one value.
		reqParts = append(reqParts, formatTickRange(exp.DurationMin, exp.DurationMax, state))
		reqs := strings.Join(reqParts, ", ")
		line := fmt.Sprintf("  %s [cyan]%s[-] - %s (%s)", canStr, exp.Key, exp.Name, reqs)
		if !exp.CanLaunch && exp.LaunchBlockReason != "" {
			line += fmt.Sprintf(" [red](%s)[-]", strings.TrimRight(exp.LaunchBlockReason, "."))
		}
		*lines = append(*lines, line)
	}
}

func cmdTrade(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "trade"}
	}
	subcmd := strings.ToLower(args[0])

	if subcmd == "list" {
		return cmdTradeList(engine)
	}
	if subcmd == "route" {
		return cmdTradeRoute(args[1:], engine)
	}
	if subcmd == "black" {
		// `trade black [resource]` is an alias for the black-market command.
		return cmdBlackMarket(args[1:], engine)
	}

	// Market trade: trade <give> <get> <amount>, where amount is how much of
	// give to sell.
	if len(args) < 3 {
		return CommandResult{Message: usageFor("trade") + ". Or 'trade list' for market rates, 'trade route list' for routes.", Type: "error"}
	}
	give := strings.ToLower(args[0])
	get := strings.ToLower(args[1])
	amount, err := parseAmount(args[2])
	if err != nil {
		return usageError(usageFor("trade"), err)
	}

	if _, err := engine.ExchangeResources(give, get, amount); err != nil {
		return errorResult(err)
	}
	return CommandResult{Type: "success"} // the engine logs both sides of the trade
}

// cmdTradeList is `trade list`: the market rate of every pair that trades
// this age, as "food → gold: 0.25 gold per food". The rates are listed with
// or without a trade building; trading itself needs one.
func cmdTradeList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	trade := state.Trade
	var lines []string
	lines = append(lines, "[gold]Market rates:[-]")
	if trade.TradeBuildings == 0 {
		lines = append(lines, "  [yellow]You need a Market to trade.[-]")
	}
	if len(trade.ExchangeRates) == 0 {
		lines = append(lines, "  [gray]Nothing trades at the market this age.[-]")
	}
	for _, key := range sortedKeysOf(trade.ExchangeRates) {
		info := trade.ExchangeRates[key]
		line := fmt.Sprintf("  [cyan]%s → %s[-]: %s %s per %s", info.From, info.To,
			strings.TrimPrefix(textfmt.RateValue(info.Rate), "+"), game.ResourceName(info.To), game.ResourceName(info.From))
		if info.Pressure > 0.05 {
			// The market takes 30% off the rate at full supply pressure, and
			// never more than half.
			line += fmt.Sprintf(" (%s lower after recent sales)", textfmt.Percent(math.Min(info.Pressure*0.3, 0.5)))
		}
		lines = append(lines, line)
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdTradeRoute(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 || strings.ToLower(args[0]) == "list" {
		return cmdTradeRouteList(engine)
	}
	subcmd := strings.ToLower(args[0])

	const usage = "Usage: trade route start|stop <route>"
	if len(args) < 2 {
		return CommandResult{Message: usage, Type: "error"}
	}
	routeKey := strings.ToLower(strings.Join(args[1:], "_"))

	switch subcmd {
	case "start":
		if err := engine.StartTradeRoute(routeKey); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the route
	case "stop":
		if err := engine.StopTradeRoute(routeKey); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"}
	default:
		return CommandResult{Message: usage, Type: "error"}
	}
}

func cmdTradeRouteList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	trade := state.Trade
	var lines []string
	lines = append(lines, "[gold]Trade routes:[-]")

	if len(trade.ActiveRoutes) > 0 {
		lines = append(lines, "\n[green]Active:[-]")
		for _, route := range trade.ActiveRoutes {
			lines = append(lines, fmt.Sprintf("  [cyan]%s[-] (%s): %s left, %s done",
				route.Name, route.Key, formatTicks(route.TicksLeft, state), textfmt.Count(route.CyclesDone, "cycle", "cycles")))
		}
	}

	if len(trade.AvailableRoutes) > 0 {
		lines = append(lines, "\n[yellow]Available:[-]")
		for _, route := range trade.AvailableRoutes {
			status := "[red]✗[-]"
			if route.CanStart {
				status = "[green]✓[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] - %s", status, route.Key, route.Name))
			lines = append(lines, fmt.Sprintf("    %s", route.Description))
		}
	}

	if len(trade.ActiveRoutes) == 0 && len(trade.AvailableRoutes) == 0 {
		lines = append(lines, "  [gray]No trade routes available yet.[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// civKey reads a civilization typed as one or more words, joined with "_"
// the way the keys are: `diplomacy ally merchant guild` is merchant_guild.
func civKey(words []string) string {
	return strings.ToLower(strings.Join(words, "_"))
}

func cmdDiplomacy(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "factions"}
	}
	subcmd := strings.ToLower(args[0])

	// The status changes, gift, tribute and raid each log their own line in
	// the engine (with the civilization's name), so the replies carry no text.
	var err error
	switch subcmd {
	case "ally", "rival", "embargo", "neutral", "gift", "tribute", "raid":
		if len(args) < 2 {
			return CommandResult{Message: usageFor("diplomacy " + subcmd), Type: "error"}
		}
		civ := civKey(args[1:])
		switch subcmd {
		case "ally":
			err = engine.SetDiplomaticStatus(civ, "allied")
		case "rival":
			err = engine.SetDiplomaticStatus(civ, "rival")
		case "embargo":
			err = engine.SetDiplomaticStatus(civ, "embargo")
		case "neutral":
			err = engine.SetDiplomaticStatus(civ, "neutral")
		case "gift":
			err = engine.SendGift(civ)
		case "tribute":
			err = engine.SendTribute(civ)
		case "raid":
			err = engine.RaidCivRoute(civ)
		}
		if err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"}

	case "deals":
		return cmdDiplomacyDeals(civKey(args[1:]), engine)

	case "accept":
		usage := usageFor("diplomacy accept")
		if len(args) < 3 {
			return CommandResult{Message: usage, Type: "error"}
		}
		n, err := parseCount(args[len(args)-1])
		if err != nil {
			return usageError(usage, err)
		}
		if _, err := engine.AcceptFactionDeal(civKey(args[1:len(args)-1]), n); err != nil {
			return errorResult(err)
		}
		return CommandResult{Type: "success"} // the engine logs the deal and its terms

	default:
		return CommandResult{Message: subUsage("diplomacy"), Type: "error"}
	}
}

// notMetReply refuses civ, a civilization the player has not met. An unmet
// civilization and a made-up one read alike, and neither is named, so the
// refusal gives nothing away (spoilers.go).
func notMetReply(civ string) CommandResult {
	return CommandResult{Message: fmt.Sprintf("You have not met a civilization called '%s'. Type 'diplomacy deals' to see the ones you have met.", civ), Type: "error"}
}

// cmdDiplomacyDeals lists the trade deals of one civilization, or of every
// one met when civ is "".
func cmdDiplomacyDeals(civ string, engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	for _, def := range state.Ruleset().Factions() {
		f, ok := state.Diplomacy.Factions[def.Key]
		if civ != "" && def.Key != civ {
			continue
		}
		if !ok || !f.Discovered {
			if civ != "" {
				return notMetReply(civ)
			}
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s", theme.Paint(theme.RoleAccent, def.Name), theme.Paint(theme.RoleDim, "("+def.Key+")")))
		var sb strings.Builder
		writeFactionDeals(&sb, f, state, 100)
		lines = append(lines, strings.TrimRight(sb.String(), "\n"))
	}
	if len(lines) == 0 {
		if civ != "" {
			return notMetReply(civ)
		}
		return CommandResult{Message: "You have not met anyone to trade with yet. Scouting expeditions make first contact.", Type: "info"}
	}
	lines = append(lines, theme.Paint(theme.RoleDim, "Take one with: diplomacy accept <civ> <n> (or plan deal <civ> <n>)"))
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdSell(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: usageFor("sell"), Type: "error"}
	}
	building := strings.ToLower(args[0])
	count := 1
	if len(args) >= 2 {
		n, err := parseCount(args[1])
		if err != nil {
			return usageError(usageFor("sell"), err)
		}
		count = n
	}
	if err := engine.SellBuilding(building, count); err != nil {
		return errorResult(err)
	}
	return CommandResult{Message: "", Type: "success"}
}

func cmdDismiss(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: usageFor("dismiss"), Type: "error"}
	}
	building := strings.ToLower(args[0])
	all := false
	count := 1
	if len(args) >= 2 {
		if strings.ToLower(args[1]) == "all" {
			all = true
		} else {
			n, err := parseCount(args[1])
			if err != nil {
				return usageError(usageFor("dismiss"), err)
			}
			count = n
		}
	}
	if err := engine.DismissWorkers(building, count, all); err != nil {
		return errorResult(err)
	}
	return CommandResult{Message: "", Type: "success"}
}

func cmdCatastrophe(args []string, engine *game.GameEngine) CommandResult {
	if len(args) > 0 {
		return CommandResult{
			Message: "Usage: catastrophe (reopens a pending catastrophe or Last Passage choice, or shows the outlook)",
			Type:    "info",
		}
	}
	// Bare `catastrophe`: reopen the pending choice, or report the outlook.
	state := engine.GetState()
	if pendingChoiceKey(state) != "" {
		return CommandResult{Type: "success", OpenCatastrophe: true}
	}
	return CommandResult{Message: catastropheOutlookText(state), Type: "info"}
}

// lastPassageStatusLines is the Last Passage risk line for `prestige`, empty
// before the Cosmic Era. The odds follow the harbinger's precision (numeric by
// then), and while a harbinger is present its figure is the one shown.
func lastPassageStatusLines(state game.GameState) []string {
	o := state.CatastropheOutlook
	if o.Passage != game.PassagePrestige || !o.Possible {
		return nil
	}
	lines := []string{fmt.Sprintf("  [red]☄ The Last Passage:[-] %s when you prestige.", lastPassageRiskText(state))}
	if h := state.Harbinger; h != nil && h.LastPassage {
		lines = append(lines, fmt.Sprintf("  %s is warning of it. Type [cyan]harbinger[-] to answer.", capFirstUI(h.Name)))
	}
	return lines
}

// lastPassageRiskText is the Last Passage chance the way the current age can
// know it (a figure in the Cosmic Era); while a harbinger is present its figure
// is the one shown, as everywhere else.
func lastPassageRiskText(state game.GameState) string {
	o := state.CatastropheOutlook
	tier, numeric, prob := o.Tier, harbingerNumericAge(state), o.Probability
	// Only the Last Passage's own thread: the Cosmic Era's fated doom, when
	// it speaks instead, warns of something else.
	if h := state.Harbinger; h != nil && h.LastPassage {
		tier, numeric, prob = h.Tier, h.Numeric, h.Probability
	}
	if numeric {
		return fmt.Sprintf("%.0f%% chance (%s)", prob*100, tier)
	}
	return fmt.Sprintf("%s risk", tier)
}

// lastPassageWarningLines explains, on `prestige confirm`, what the Last
// Passage does if the roll goes against you. Empty before the Cosmic Era.
func lastPassageWarningLines(state game.GameState) []string {
	o := state.CatastropheOutlook
	if o.Passage != game.PassagePrestige || !o.Possible {
		return nil
	}
	lp := state.LastPassage
	lines := []string{
		"",
		fmt.Sprintf("  [red]☄ In the Cosmic Era prestige can bring the Last Passage: %s.[-]", lastPassageRiskText(state)),
		"  If it comes, prestige waits for your choice:",
		fmt.Sprintf("    Endure: keep %d%% of this run's points (%d of %d).", lp.KeepPct, lp.PointsIfEndured, lp.PointsNow),
	}
	if lp.CosmicLegacy {
		lines = append(lines, "    Succumb is closed: you already carry the Cosmic Legacy.")
	} else {
		lines = append(lines, fmt.Sprintf("    Succumb: no points from this run, and the Cosmic Legacy (+%.0f%% production, permanent).", game.CosmicLegacyProductionBonus*100))
	}
	if lp.Invited {
		lines = append(lines, "  [red]You invited it. It will come.[-]")
	}
	return lines
}

// cmdHarbinger opens the Harbinger panel, or answers the harbinger directly
// with `harbinger appease|brace|invite`.
func cmdHarbinger(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "harbinger"}
	}
	var err error
	switch strings.ToLower(args[0]) {
	case "appease":
		err = engine.HarbingerAppease()
	case "brace":
		err = engine.HarbingerBrace()
	case "invite":
		err = engine.HarbingerInvite()
	default:
		return CommandResult{
			Message: "Usage: harbinger (opens the Harbinger panel), or harbinger appease|brace|invite (answers it directly)",
			Type:    "info",
		}
	}
	if err != nil {
		return errorResult(err)
	}
	// The engine logs the action itself.
	return CommandResult{Type: "success"}
}

// faithStrengthText says the faith strength the odds are read from, with the
// two things it comes from: "faith strength 22% (devotion 1.0x, 100% of your
// faith kept)". The same figure the Economy panel's faith row shows.
func faithStrengthText(state game.GameState) string {
	return "faith strength " + faithStrengthFigure(state)
}

// faithStrengthFigure is the figure in faithStrengthText. Devotion is what
// the town's faith buildings have made this run against a moderate set (five
// staffed copies of each); kept is the share of the run's faith still held.
func faithStrengthFigure(state game.GameState) string {
	o := state.CatastropheOutlook
	if o.FaithDevotion <= 0 {
		return fmt.Sprintf("%.0f%% (your faith buildings have made no faith yet)", o.FaithStrength*100)
	}
	devotion := fmt.Sprintf("%.1f", o.FaithDevotion)
	if o.FaithDevotion >= 10 {
		devotion = fmt.Sprintf("%.0f", o.FaithDevotion)
	} else if o.FaithDevotion < 0.1 {
		devotion = fmt.Sprintf("%.2f", o.FaithDevotion)
	}
	return fmt.Sprintf("%.0f%% (devotion %sx, %.0f%% of your faith kept)", o.FaithStrength*100, devotion, o.FaithKept*100)
}

// catastropheOutlookText renders the no-pending status line for the bare
// `catastrophe` command: what the player can know. In an era that can be
// fated, a doom is only ever known through its harbinger, so with none here
// the era reads quiet whether or not one is fated. In the final epoch it is
// the Last Passage's odds at prestige.
func catastropheOutlookText(state game.GameState) string {
	o := state.CatastropheOutlook
	var sb strings.Builder
	sb.WriteString("No catastrophe pending.\n")
	if o.Passage == game.PassagePrestige && o.Warned {
		// The Cosmic Era's fated doom, foretold: it comes before the Last
		// Passage can.
		fmt.Fprintf(&sb, "  %s, %s.\n", doomWarningText(state), faithStrengthText(state))
	}
	switch {
	case o.Passage == game.PassagePrestige && o.Possible:
		fmt.Fprintf(&sb, "  Next passage (prestige, the Last Passage): %s, %s.",
			outlookRiskText(state), faithStrengthText(state))
	case o.Passage == game.PassagePrestige:
		sb.WriteString("  This is the final epoch: its passage is prestige, and the Last Passage cannot strike now.")
	case o.Warned:
		fmt.Fprintf(&sb, "  %s, %s.", doomWarningText(state), faithStrengthText(state))
	default:
		sb.WriteString("  " + eraOutlookText(state))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// eraOutlookText is the outlook of an era with no harbinger present, in one
// sentence. It reads only what the player has seen (the era's rule and its
// harbinger history), so a fated era and a quiet one read the same until the
// harbinger comes.
func eraOutlookText(state game.GameState) string {
	era := currentEraName(state)
	switch {
	case !state.Ruleset().CatastropheAllowed(state.EpochKey):
		return fmt.Sprintf("No catastrophe can strike in the %s.", era)
	case state.CatastropheOutlook.Possible:
		return fmt.Sprintf("No harbinger has come: the %s is quiet, for now. A doom is always foretold before it strikes.", era)
	}
	if r := eraDoomRecord(state); r != nil && r.Outcome == game.HarbingerOutcomeSpared {
		return fmt.Sprintf("The doom %s foretold passed you by. Nothing more will strike before the %s ends.", r.Name, era)
	}
	return fmt.Sprintf("The %s's doom has come. Nothing more will strike before it ends.", era)
}

// eraDoomRecord is the newest resolved harbinger of the current era's doom,
// or nil.
func eraDoomRecord(state game.GameState) *game.HarbingerRecord {
	for i := len(state.HarbingerHistory) - 1; i >= 0; i-- {
		if r := &state.HarbingerHistory[i]; r.EpochKey == state.EpochKey && r.TargetEpochKey == state.EpochKey {
			return r
		}
	}
	return nil
}

// doomWarningText says what the harbinger present foretells of the era's
// doom: who, when (as far as the figure can tell) and how likely.
func doomWarningText(state game.GameState) string {
	h := state.Harbinger
	when := ", with no word of when"
	if h.WhenText != "" {
		when = " " + h.WhenText
	}
	return fmt.Sprintf("%s warns of doom%s: %s", capFirstUI(h.Name), when, riskText(state))
}

// outlookRiskText describes the Last Passage's risk the way the current age
// can know it, with its thread as the speaker when that is the thread present
// (the Cosmic Era's fated doom may be speaking instead).
func outlookRiskText(state game.GameState) string {
	if h := state.Harbinger; h != nil && h.LastPassage {
		return fmt.Sprintf("%s warns of %s", capFirstUI(h.Name), riskText(state))
	}
	o := state.CatastropheOutlook
	return riskFrom(state, o.Tier, harbingerNumericAge(state), o.Probability)
}

// riskText is the risk the way the current age can know it. From the
// Industrial Age on the odds are published: "75% catastrophe chance
// (medium)". Before it there is no figure, only a severity. While a
// harbinger is present its severity is the one shown everywhere, so no other
// screen can contradict (and so expose) a false prophet.
func riskText(state game.GameState) string {
	o := state.CatastropheOutlook
	tier, numeric, prob := o.Tier, harbingerNumericAge(state), o.Probability
	if h := state.Harbinger; h != nil {
		tier, numeric, prob = h.Tier, h.Numeric, h.Probability
	}
	return riskFrom(state, tier, numeric, prob)
}

// riskFrom words a risk: the odds from the Industrial Age on, a severity
// before it.
func riskFrom(state game.GameState, tier game.CatastropheTier, numeric bool, prob float64) string {
	switch {
	case numeric:
		return fmt.Sprintf("%.0f%% catastrophe chance (%s)", prob*100, tier)
	case game.SightOf(&state).Age("industrial_age"):
		return fmt.Sprintf("%s risk of catastrophe (no figures before %s)", tier, ageRef(state, "industrial_age"))
	}
	// The age that prints the odds is named only once the player can see it.
	return fmt.Sprintf("%s risk of catastrophe (no figures this early)", tier)
}

// cmdPlan is the `plan` command. Bare `plan` opens the Plan panel.
func cmdPlan(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "plan"}
	}
	sub := strings.ToLower(args[0])
	rest := args[1:]
	switch sub {
	case "build":
		if len(rest) == 0 || len(rest) > 2 {
			return CommandResult{Message: usageFor("plan build"), Type: "error"}
		}
		key := strings.ToLower(rest[0])
		count := 1
		if len(rest) == 2 {
			n, err := parseCount(rest[1])
			if err != nil {
				return usageError(usageFor("plan build"), err)
			}
			count = n
		}
		added, err := engine.PlanAddBuild(key, count)
		if err != nil {
			return errorResult(err)
		}
		msg := fmt.Sprintf("Planned %s. It starts as soon as the resources are there.", game.BuildingCount(added, key))
		if added < count {
			msg = fmt.Sprintf("Planned %s (the most its limit allows). It starts as soon as the resources are there.", game.BuildingCount(added, key))
		}
		return CommandResult{Message: msg, Type: game.LogRoutine}
	case "research", "res":
		if len(rest) == 0 {
			return CommandResult{Message: usageFor("plan research"), Type: "error"}
		}
		key := strings.ToLower(strings.Join(rest, "_"))
		if err := engine.PlanAddResearch(key); err != nil {
			return errorResult(err)
		}
		planned, _ := engine.Rules().Tech(key)
		return CommandResult{Message: fmt.Sprintf("Planned research: %s. Techs start one at a time, in plan order.", planned.Name), Type: game.LogRoutine}
	case "list":
		return CommandResult{Message: planListText(engine.GetState()), Type: "info"}
	case "remove", "rm":
		n, err := planIndexArg(rest, "remove")
		if err != nil {
			return usageError(usageFor("plan remove"), err)
		}
		what, err := engine.PlanRemove(n)
		if err != nil {
			return errorResult(err)
		}
		return CommandResult{Message: "Removed " + what + " from the plan.", Type: game.LogRoutine}
	case "up", "down":
		n, err := planIndexArg(rest, sub)
		if err != nil {
			return usageError(usageFor("plan "+sub), err)
		}
		delta := -1
		if sub == "down" {
			delta = 1
		}
		to, err := engine.PlanMove(n, delta)
		if err != nil {
			return errorResult(err)
		}
		return CommandResult{Message: fmt.Sprintf("Plan item %d is now number %d.", n, to), Type: game.LogRoutine}
	case "clear":
		n := engine.PlanClear()
		return CommandResult{Message: fmt.Sprintf("Cleared the plan (%s).", textfmt.Count(n, "item", "items")), Type: game.LogRoutine}
	case "trade":
		// plan trade <give> <get> [amount]: amount is how much of get to buy.
		if len(rest) < 2 || len(rest) > 3 {
			return CommandResult{Message: usageFor("plan trade"), Type: "error"}
		}
		give, get := strings.ToLower(rest[0]), strings.ToLower(rest[1])
		amount := 0.0
		if len(rest) == 3 {
			a, err := parseAmount(rest[2])
			if err != nil {
				return usageError(usageFor("plan trade"), err)
			}
			amount = a
		}
		if err := engine.PlanAddTrade(give, get, amount); err != nil {
			return errorResult(err)
		}
		if amount > 0 {
			return CommandResult{Message: fmt.Sprintf("Plan: buy %s with %s as it comes in.",
				game.Amount(amount, get), game.ResourceName(give)), Type: game.LogRoutine}
		}
		return CommandResult{Message: fmt.Sprintf("Plan: buy %s with %s as it comes in, until you remove the item.",
			game.ResourceName(get), game.ResourceName(give)), Type: game.LogRoutine}
	case "deal":
		usage := usageFor("plan deal")
		if len(rest) < 2 {
			return CommandResult{Message: usage, Type: "error"}
		}
		n, err := parseCount(rest[len(rest)-1])
		if err != nil {
			return usageError(usage, err)
		}
		civ := civKey(rest[:len(rest)-1])
		if err := engine.PlanAddDeal(civ, n); err != nil {
			return errorResult(err)
		}
		return CommandResult{Message: fmt.Sprintf("Planned: take deal %d with the %s as soon as its price is there.", n, game.CivName(civ)), Type: game.LogRoutine}
	case "advance":
		if len(rest) != 0 {
			return CommandResult{Message: usageFor("plan advance"), Type: "error"}
		}
		if err := engine.PlanAddAdvance(); err != nil {
			return errorResult(err)
		}
		return CommandResult{Message: "Planned: advance as soon as the next age's requirements are met.", Type: game.LogRoutine}
	}
	return CommandResult{Message: subUsage("plan"), Type: "error"}
}

// planIndexArg reads the one item number a plan subcommand takes.
func planIndexArg(rest []string, sub string) (int, error) {
	if len(rest) != 1 {
		return 0, fmt.Errorf("plan %s takes one item number", sub)
	}
	return parseCount(rest[0])
}

// cmdMap handles map and its aliases citymap and worldmap. Bare, it opens
// the Map panel (worldmap on the known world); "style" and "glyphs" read or
// change the settings.
func cmdMap(cmd string, args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "map", MapWorld: cmd == "worldmap"}
	}
	switch strings.ToLower(args[0]) {
	case "style":
		return cmdMapStyle(args[1:], engine)
	case "glyphs":
		return cmdMapGlyphs(args[1:], engine)
	case "flows":
		return cmdMapFlows(args[1:])
	}
	return CommandResult{Message: subUsage("map"), Type: "error"}
}

// cmdMapFlows turns the Map's flows overlay (full stores, understaffed
// buildings, idle workers) on, off, or over when bare. It is a view option
// for the session, applied by the dashboard.
func cmdMapFlows(args []string) CommandResult {
	if len(args) > 1 {
		return usageError(usageFor("map flows"), fmt.Errorf("on or off, please"))
	}
	if len(args) == 0 {
		return CommandResult{Type: game.LogRoutine, Message: "Flows overlay switched.", MapFlows: "switch"}
	}
	mode := strings.ToLower(args[0])
	if mode != "on" && mode != "off" {
		return usageError(usageFor("map flows"), fmt.Errorf("map flows takes on or off, not %q", args[0]))
	}
	return CommandResult{Type: game.LogRoutine, Message: flowsReply(mode == "on"), MapFlows: mode}
}

// flowsReply is the reply once the flows overlay is on or off.
func flowsReply(on bool) string {
	if on {
		return "Flows overlay on: full stores, understaffed buildings and idle workers show on the Map."
	}
	return "Flows overlay off."
}

// mapSessionOnly ends a map setting's reply when no account is loaded to
// keep it on.
const mapSessionOnly = " No account is loaded, so it lasts for this session."

func cmdMapStyle(args []string, engine *game.GameEngine) CommandResult {
	reg := all.Registry()
	acct := engine.Account()
	cur := resolveMapSettings(acct, reg)
	if len(args) == 0 {
		return CommandResult{Type: "info", Message: fmt.Sprintf("Map style: %s. Styles: %s. Type map style <name> to switch.",
			styleTitle(reg, cur.Style), strings.Join(reg.Names(), ", "))}
	}
	if len(args) > 1 {
		return usageError(usageFor("map style"), fmt.Errorf("one style name, please"))
	}
	name := strings.ToLower(args[0])
	switch name {
	case "roguelike", "skyline":
	default:
		return usageError(usageFor("map style"), fmt.Errorf("there is no map style %q", args[0]))
	}
	// A routine confirmation (the map shows the change); without an account
	// the reply carries a caveat, so it is an info line instead.
	res := CommandResult{Type: game.LogRoutine, Message: fmt.Sprintf("Map style set to %s.", styleTitle(reg, name)),
		MapPref: mapPref{Key: "style", Value: name}}
	if acct == nil {
		res.Type, res.Message = "info", res.Message+mapSessionOnly
		return res
	}
	if err := acct.SetMapStyle(name); err != nil {
		return errorResult(fmt.Errorf("the map style could not be saved: %w", err))
	}
	return res
}

func cmdMapGlyphs(args []string, engine *game.GameEngine) CommandResult {
	acct := engine.Account()
	cur := resolveMapSettings(acct, all.Registry())
	if len(args) == 0 {
		return CommandResult{Type: "info", Message: fmt.Sprintf("Map glyphs: %s. Tiers: %s. Type map glyphs <tier> to switch (icons checks your font for nerd).",
			cur.Tier, strings.Join(mapmodel.TierNames, ", "))}
	}
	if len(args) > 1 {
		return usageError(usageFor("map glyphs"), fmt.Errorf("one glyph tier, please"))
	}
	name := strings.ToLower(args[0])
	switch name {
	case "ascii", "unicode", "nerd":
	default:
		return usageError(usageFor("map glyphs"), fmt.Errorf("there is no glyph tier %q", args[0]))
	}
	res := CommandResult{Type: game.LogRoutine, Message: fmt.Sprintf("Map glyphs set to %s.", name),
		MapPref: mapPref{Key: "glyphs", Value: name}}
	if name == "nerd" {
		res.Type = "info" // advice, not a plain confirmation
		res.Message += " If the map shows boxes or question marks, type icons."
	}
	if acct == nil {
		res.Type, res.Message = "info", res.Message+mapSessionOnly
		return res
	}
	if err := acct.SetMapGlyphs(name); err != nil {
		return errorResult(fmt.Errorf("the map glyphs could not be saved: %w", err))
	}
	return res
}

// cmdMinimap shows or sets the minimap setting: the dashboard's mini map
// above the Buildings list, on by default.
func cmdMinimap(args []string, engine *game.GameEngine) CommandResult {
	acct := engine.Account()
	if len(args) == 0 {
		state := "off"
		if resolveMapSettings(acct, all.Registry()).Minimap {
			state = "on"
		}
		return CommandResult{Type: "info", Message: "Mini map: " + state + ". Type minimap on or minimap off to change it."}
	}
	if len(args) > 1 {
		return usageError(usageFor("minimap"), fmt.Errorf("on or off, please"))
	}
	val := strings.ToLower(args[0])
	if val != "on" && val != "off" {
		return usageError(usageFor("minimap"), fmt.Errorf("minimap takes on or off, not %q", args[0]))
	}
	on := val == "on"
	res := CommandResult{Type: game.LogRoutine, Message: "Mini map off. Type minimap on to bring it back.",
		MapPref: mapPref{Key: "minimap", Value: val}}
	if on {
		res.Message = "Mini map on. It shows above the Buildings list when the terminal has room (about 120x40 and up)."
	}
	if acct == nil {
		res.Type, res.Message = "info", res.Message+mapSessionOnly
		return res
	}
	if err := acct.SetMinimap(on); err != nil {
		return errorResult(fmt.Errorf("the minimap setting could not be saved: %w", err))
	}
	return res
}
