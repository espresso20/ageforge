package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// completer answers the command prompt's questions from the registry
// (commands.go) and the game: what completes the text so far (Tab, the ghost
// text).
//
// It runs on every keystroke and every draw, on the UI goroutine. Values
// read from the game are cached per snapshot: the dashboard hands out the
// state of its last refresh, and the cache is rebuilt when that changes.
type completer struct {
	engine *game.GameEngine
	cmds   []*Command
	// snapshot returns the state to suggest from. nil: engine.GetState()
	// on every call, uncached (the smoke suite's use).
	snapshot func() *game.GameState

	// Every value each game-thing kind can ever take, for checking typed
	// words; built on first use.
	universe map[ArgKind]map[string]bool

	cachedFor *game.GameState
	cache     map[string][]string
}

func newCompleter(engine *game.GameEngine, snapshot func() *game.GameState) *completer {
	return &completer{engine: engine, cmds: registry(), snapshot: snapshot}
}

// NewAutoCompleter returns the prompt's completions for a line, for callers
// outside the dashboard (the smoke fuzz scenario walks commands with it):
// every full line that completes the text, best first, without the trailing
// space Tab adds. It reads a fresh snapshot on every call.
func NewAutoCompleter(engine *game.GameEngine) func(string) []string {
	c := newCompleter(engine, nil)
	return func(text string) []string {
		cands := c.candidates(text)
		for i, s := range cands {
			cands[i] = strings.TrimRight(s, " ")
		}
		return cands
	}
}

// state returns the snapshot to suggest from, dropping the cache when it
// has moved on.
func (c *completer) state() *game.GameState {
	var st *game.GameState
	if c.snapshot != nil {
		st = c.snapshot()
	}
	if st == nil {
		s := c.engine.GetState()
		st = &s
	}
	if st != c.cachedFor || c.cache == nil {
		c.cachedFor, c.cache = st, map[string][]string{}
	}
	return st
}

// valid reports whether w is a value the slot's kind can ever take. Typed
// words are checked against everything the game has, not only what is
// suggested now: a real but locked building is a complete command, and the
// game says why it can't be built.
func (c *completer) valid(a Arg, w string) bool {
	w = strings.ToLower(w)
	switch a.Kind {
	case ArgSave:
		for _, s := range c.values(a.Kind, nil) {
			if strings.EqualFold(s, w) {
				return true
			}
		}
		return false
	case ArgAccount:
		return true
	}
	if c.universe == nil {
		c.universe = buildUniverse()
	}
	return c.universe[a.Kind][w]
}

// buildUniverse lists every key each game-thing kind can take.
func buildUniverse() map[ArgKind]map[string]bool {
	set := func(keys []string) map[string]bool {
		m := make(map[string]bool, len(keys))
		for _, k := range keys {
			m[k] = true
		}
		return m
	}
	var buildings, techs, resources, factions, themes, routes, upgrades, expeditions []string
	for k := range config.BuildingByKey() {
		buildings = append(buildings, k)
	}
	for _, t := range config.Technologies() {
		techs = append(techs, t.Key)
	}
	for _, r := range config.BaseResources() {
		resources = append(resources, r.Key)
	}
	for _, f := range config.BaseFactions() {
		factions = append(factions, f.Key)
	}
	for _, t := range theme.All() {
		themes = append(themes, t.Key)
	}
	for _, r := range config.BaseTradeRoutes() {
		routes = append(routes, r.Key)
	}
	for _, u := range config.PrestigeUpgrades() {
		upgrades = append(upgrades, u.Key)
	}
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	mm := game.NewMilitaryManager()
	for _, a := range config.AgeOrder() {
		for _, x := range mm.GetAvailableExpeditions(a, order) {
			expeditions = append(expeditions, x.Key)
		}
	}
	b, r := set(buildings), set(resources)
	return map[ArgKind]map[string]bool{
		ArgBuilding: b, ArgPlanBuilding: b, ArgBuiltBuilding: b, ArgWorkerBuilding: b,
		ArgStaffedBuilding: b, ArgUpgradeBuilding: b,
		ArgTech: set(techs), ArgPlanTech: set(techs),
		ArgResource: r, ArgTradeFrom: r, ArgTradeTo: r,
		ArgFaction: set(factions), ArgTheme: set(themes),
		ArgExpedition: set(expeditions), ArgCampaign: set(expeditions),
		ArgRouteAvailable: set(routes), ArgRouteActive: set(routes),
		ArgPrestigeUpgrade: set(upgrades),
	}
}

// complete reports whether line is a whole command as typed: Enter runs it
// without looking at suggestions. Dev-console lines count while dev mode is
// on; the console judges them.
func (c *completer) complete(line string) bool {
	words := strings.Fields(line)
	if game.DevModeActive && len(words) > 0 && strings.HasPrefix(words[0], "/") {
		return true
	}
	p, ok := parse(c.cmds, words)
	return ok && p.complete(c.valid)
}

// ghost is the dim text shown after what was typed: the rest of the best
// completion. Nothing for an empty prompt, or for a line that is already a
// whole command (Enter would run it as typed).
func (c *completer) ghost(text string) string {
	if strings.TrimSpace(text) == "" || c.complete(text) {
		return ""
	}
	cands := c.candidates(text)
	if len(cands) == 0 {
		return ""
	}
	typed := strings.TrimLeft(text, " ")
	best := strings.TrimRight(cands[0], " ")
	if len(best) <= len(typed) || !strings.EqualFold(best[:len(typed)], typed) {
		return ""
	}
	return best[len(typed):]
}

// candidates returns every full line that completes text, best first. A
// candidate that leaves more to type ends in a space.
func (c *completer) candidates(text string) []string {
	text = strings.TrimLeft(text, " ")
	if text == "" {
		return nil
	}
	words := strings.Fields(text)
	trailing := strings.HasSuffix(text, " ")
	if len(words) == 1 && !trailing {
		return c.nameCandidates(words[0])
	}
	partial := ""
	done := words[1:]
	if !trailing {
		partial = words[len(words)-1]
		done = words[1 : len(words)-1]
	}
	prefix := text[:len(text)-len(partial)]
	p, ok := parse(c.cmds, append([]string{words[0]}, done...))
	if !ok {
		return nil
	}
	node := p.last()
	slot := len(p.args)
	var out []string
	add := func(word string, more bool) {
		line := prefix + word
		if more {
			line += " "
		}
		for _, o := range out {
			if o == line {
				return
			}
		}
		out = append(out, line)
	}
	lp := strings.ToLower(partial)
	if slot < len(node.Args) {
		a := node.Args[slot]
		more := slot+1 < len(node.Args)
		for _, v := range c.values(a.Kind, p.args) {
			if strings.HasPrefix(strings.ToLower(v), lp) {
				add(v, more)
			}
		}
		if slot == 0 {
			for _, s := range sortedSubs(node) {
				if strings.HasPrefix(s.Name, lp) {
					add(s.Name, len(s.Subs) > 0 || len(s.Args) > 0)
				}
			}
		}
		for _, w := range sortedWords(a.Words) {
			if strings.HasPrefix(w, lp) {
				add(w, more)
			}
		}
	} else if slot == 0 {
		for _, s := range sortedSubs(node) {
			if strings.HasPrefix(s.Name, lp) {
				add(s.Name, len(s.Subs) > 0 || len(s.Args) > 0)
			}
		}
	}
	return out
}

// nameCandidates completes a command name: an exact name first, then the
// other names, then aliases. Dev-console commands only while dev mode is on.
func (c *completer) nameCandidates(partial string) []string {
	lp := strings.ToLower(partial)
	if strings.HasPrefix(lp, "/") {
		if !game.DevModeActive {
			return nil
		}
		var out []string
		for _, d := range devCommands {
			if strings.HasPrefix(d.name, lp) {
				out = append(out, d.name)
			}
		}
		return out
	}
	var exact, names, aliases []string
	for _, cmd := range c.cmds {
		more := len(cmd.Subs) > 0 || len(cmd.Args) > 0
		for i, n := range append([]string{cmd.Name}, cmd.Aliases...) {
			if !strings.HasPrefix(n, lp) {
				continue
			}
			line := n
			if more {
				line += " "
			}
			switch {
			case n == lp:
				exact = append(exact, line)
			case i == 0:
				names = append(names, line)
			default:
				aliases = append(aliases, line)
			}
		}
	}
	sort.Strings(names)
	sort.Strings(aliases)
	return append(append(exact, names...), aliases...)
}

func sortedSubs(c *Command) []*Command {
	out := append([]*Command(nil), c.Subs...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func sortedWords(ws []string) []string {
	out := append([]string(nil), ws...)
	sort.Strings(out)
	return out
}

// values returns what a slot of kind k can take right now, best first. prev
// is the arguments already typed before the slot.
func (c *completer) values(k ArgKind, prev []string) []string {
	st := c.state()
	key := strconv.Itoa(int(k))
	if k == ArgTradeTo && len(prev) > 0 {
		key += ":" + strings.ToLower(prev[0])
	}
	if v, ok := c.cache[key]; ok {
		return v
	}
	v := c.compute(k, prev, *st)
	c.cache[key] = v
	return v
}

func (c *completer) compute(k ArgKind, prev []string, st game.GameState) []string {
	switch k {
	case ArgBuilding:
		return buildableBuildingKeys(st)
	case ArgPlanBuilding:
		return plannableBuildingKeys(st)
	case ArgBuiltBuilding:
		return builtBuildingKeys(st)
	case ArgWorkerBuilding:
		return workerBuildingKeys(st)
	case ArgStaffedBuilding:
		return assignedBuildingKeysAll(st)
	case ArgUpgradeBuilding:
		return upgradeableBuildingKeys(c.engine)
	case ArgTech:
		return availableTechKeys(st)
	case ArgPlanTech:
		return plannableTechKeys(st)
	case ArgResource:
		return unlockedResourceKeys(st)
	case ArgTradeFrom:
		return tradeKeys(st, "")
	case ArgTradeTo:
		from := ""
		if len(prev) > 0 {
			from = prev[0]
		}
		return tradeKeys(st, from)
	case ArgFaction:
		return discoveredFactionKeys(st)
	case ArgTheme:
		return themeKeys()
	case ArgSave:
		return saveNames()
	case ArgExpedition:
		return expeditionKeysByCategory(st, game.ExpeditionScouting)
	case ArgCampaign:
		return expeditionKeysByCategory(st, game.ExpeditionMilitary)
	case ArgRouteAvailable:
		return availableTradeRouteKeys(st)
	case ArgRouteActive:
		return activeTradeRouteKeys(st)
	case ArgPrestigeUpgrade:
		return prestigeUpgradeKeys(st)
	case ArgSpeed:
		return availableSpeedOptions(c.engine)
	case ArgAccount:
		return localAccountNames(c.engine)
	case ArgPlanItem:
		nums := make([]string, len(st.Plan))
		for i := range st.Plan {
			nums[i] = strconv.Itoa(i + 1)
		}
		return nums
	}
	return nil
}

// buildableBuildingKeys returns the unlocked buildings, sorted.
func buildableBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Unlocked {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// plannableBuildingKeys is what `plan build` can take: this age's unlocked
// buildings (its wonder included) short of their MaxCount, and the next
// age's, which wait for the advance.
func plannableBuildingKeys(state game.GameState) []string {
	defs := config.BuildingByKey()
	var keys []string
	for key, bs := range state.Buildings {
		d := defs[key]
		now := bs.Unlocked && !bs.IsLegacy && !bs.AtMaxCount && d.RequiredAge == state.Age
		if now || (state.NextAge != "" && d.RequiredAge == state.NextAge) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// availableTechKeys returns the techs that can start now, sorted.
func availableTechKeys(state game.GameState) []string {
	var keys []string
	for key, ts := range state.Research.Techs {
		if ts.Available {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// plannableTechKeys is what `plan research` can take: unresearched techs of
// this age, an earlier one or the next that are neither in progress nor
// planned already. Prerequisites may still be missing; they can be planned
// first.
func plannableTechKeys(state game.GameState) []string {
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	planned := map[string]bool{state.Research.CurrentTech: true}
	for _, v := range state.Plan {
		if v.Kind == game.PlanResearch {
			planned[v.Key] = true
		}
	}
	var keys []string
	for key, ts := range state.Research.Techs {
		if !ts.Researched && !planned[key] && (order[ts.Age] <= order[state.Age] || ts.Age == state.NextAge) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// unlockedResourceKeys returns the unlocked resources, sorted.
func unlockedResourceKeys(state game.GameState) []string {
	var keys []string
	for key, rs := range state.Resources {
		if rs.Unlocked {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// tradeKeys is the market's side of an exchange: with from empty, what it
// buys from you; otherwise what it sells for from. Before there is a market
// it falls back to every unlocked resource (the game then says a market is
// needed).
func tradeKeys(state game.GameState, from string) []string {
	if len(state.Trade.ExchangeRates) == 0 {
		return unlockedResourceKeys(state)
	}
	seen := map[string]bool{}
	var keys []string
	for _, x := range state.Trade.ExchangeRates {
		k := x.From
		if from != "" {
			if !strings.EqualFold(x.From, from) {
				continue
			}
			k = x.To
		}
		if !seen[k] && state.Resources[k].Unlocked {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// builtBuildingKeys returns the buildings with at least one copy, sorted.
func builtBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Count > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// workerBuildingKeys returns the unlocked buildings that take workers,
// sorted.
func workerBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Unlocked && bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// assignedBuildingKeysAll returns the buildings with at least one worker
// assigned, in any domain, sorted.
func assignedBuildingKeysAll(state game.GameState) []string {
	seen := map[string]bool{}
	for _, vt := range state.Workers.Types {
		for buildingKey, count := range vt.Assignments {
			if count > 0 {
				seen[buildingKey] = true
			}
		}
	}
	var keys []string
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// expeditionKeysByCategory returns the visible expedition keys of one
// category (game.ExpeditionScouting or game.ExpeditionMilitary), sorted.
// Locked entries are included; the engine enforces eligibility on launch.
func expeditionKeysByCategory(state game.GameState, category string) []string {
	var keys []string
	for _, exp := range state.Military.Expeditions {
		if exp.Category == category {
			keys = append(keys, exp.Key)
		}
	}
	sort.Strings(keys)
	return keys
}

// prestigeUpgradeKeys returns the prestige upgrades with a tier left, sorted.
func prestigeUpgradeKeys(state game.GameState) []string {
	var keys []string
	for key, u := range state.Prestige.Upgrades {
		if u.NextCost > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// availableTradeRouteKeys returns the routes that can start, sorted.
func availableTradeRouteKeys(state game.GameState) []string {
	var keys []string
	for _, route := range state.Trade.AvailableRoutes {
		keys = append(keys, route.Key)
	}
	sort.Strings(keys)
	return keys
}

// activeTradeRouteKeys returns the running routes, sorted.
func activeTradeRouteKeys(state game.GameState) []string {
	var keys []string
	for _, route := range state.Trade.ActiveRoutes {
		keys = append(keys, route.Key)
	}
	sort.Strings(keys)
	return keys
}

// discoveredFactionKeys returns the civilizations met so far, sorted.
func discoveredFactionKeys(state game.GameState) []string {
	var keys []string
	for key, f := range state.Diplomacy.Factions {
		if f.Discovered {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// availableSpeedOptions returns the speed multipliers from 1.0 up to the
// current max, in steps of 0.5 (the max rises 0.5x per wonder built).
func availableSpeedOptions(engine *game.GameEngine) []string {
	maxSpeed := engine.GetMaxSpeed()
	var options []string
	for s := 1.0; s <= maxSpeed; s += 0.5 {
		options = append(options, fmt.Sprintf("%.1f", s))
	}
	return options
}

// upgradeableBuildingKeys returns the buildings with an upgrade available,
// sorted.
func upgradeableBuildingKeys(engine *game.GameEngine) []string {
	var keys []string
	for _, u := range engine.GetAvailableUpgrades() {
		keys = append(keys, u.FromKey)
	}
	sort.Strings(keys)
	return keys
}

func saveNames() []string {
	saves, err := game.ListSaves()
	if err != nil {
		return nil
	}
	return saves
}

// localAccountNames returns the display names of every named local account,
// for `account switch <name>`. Read-only: ListAccounts never changes the
// active account.
func localAccountNames(engine *game.GameEngine) []string {
	var names []string
	for _, s := range engine.ListAccounts() {
		if strings.TrimSpace(s.DisplayName) != "" {
			names = append(names, s.DisplayName)
		}
	}
	sort.Strings(names)
	return names
}
