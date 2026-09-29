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
// text) and what Enter should run.
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
		ArgResource: r, ArgWonderResource: r, ArgTradeFrom: r, ArgTradeTo: r,
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

// dangerous reports whether line would run a Dangerous command.
func (c *completer) dangerous(line string) bool {
	p, ok := parse(c.cmds, strings.Fields(line))
	return ok && p.dangerous()
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

// enterLine decides what Enter runs. A whole command runs as typed.
// Otherwise a ghost completion that makes a whole command runs, unless it is
// Dangerous: then the completion goes into the field (run false) and a
// second Enter runs it. Anything else runs as typed, and the game says what
// is wrong with it.
func (c *completer) enterLine(text string) (line string, run bool) {
	if strings.TrimSpace(text) == "" || c.complete(text) {
		return text, true
	}
	g := c.ghost(text)
	if g == "" {
		return text, true
	}
	full := strings.TrimSpace(text + g)
	if !c.complete(full) {
		return text, true
	}
	if c.dangerous(full) {
		return full, false
	}
	return full, true
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

// closestCommand is the registry command name nearest to word by edit
// distance (at most 2, or 1 for a word of three letters or fewer), or "" when
// none is that close. Names are compared in sorted order so a tie always
// gives the same suggestion.
func closestCommand(word string) string {
	word = strings.ToLower(word)
	limit := 2
	if len([]rune(word)) <= 3 {
		limit = 1
	}
	var names []string
	for _, c := range registry() {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	best, bestDist := "", limit+1
	for _, n := range names {
		if d := editDistance(word, n); d < bestDist {
			best, bestDist = n, d
		}
	}
	return best
}

// editDistance is the Levenshtein distance between a and b, by rune.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr := make([]int, len(rb)+1)
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = curr
	}
	return prev[len(rb)]
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
	if (k == ArgTradeTo || k == ArgDeal) && len(prev) > 0 {
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
	case ArgWonderResource:
		return wonderNeedKeys(st)
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
		return unlockedThemeKeys(c.engine)
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
	case ArgDeal:
		if len(prev) == 0 {
			return nil
		}
		return dealNumbers(st, strings.ToLower(prev[0]))
	case ArgPlanItem:
		nums := make([]string, len(st.Plan))
		for i := range st.Plan {
			nums[i] = strconv.Itoa(i + 1)
		}
		return nums
	}
	return nil
}

// byRank sorts keys by rank (lower first), then alphabetically.
func byRank(keys []string, rank func(string) int) []string {
	sort.Slice(keys, func(i, j int) bool {
		ri, rj := rank(keys[i]), rank(keys[j])
		if ri != rj {
			return ri < rj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// buildableBuildingKeys is what `build` suggests: buildings that can be
// built in this age (unlocked, not superseded, short of their limit), the
// affordable ones first.
func buildableBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Unlocked && !bs.IsLegacy && !bs.AtMaxCount {
			keys = append(keys, key)
		}
	}
	return byRank(keys, func(k string) int {
		if state.Buildings[k].CanBuild {
			return 0
		}
		return 1
	})
}

// plannableBuildingKeys is what `plan build` takes and suggests: this age's
// unlocked buildings (its wonder included) short of their MaxCount, the
// affordable ones first, then the next age's, which wait for the advance.
func plannableBuildingKeys(state game.GameState) []string {
	defs := config.BuildingByKey()
	rank := map[string]int{}
	var keys []string
	for key, bs := range state.Buildings {
		d := defs[key]
		switch {
		case bs.Unlocked && !bs.IsLegacy && !bs.AtMaxCount && d.RequiredAge == state.Age:
			rank[key] = 1
			if bs.CanBuild {
				rank[key] = 0
			}
		case state.NextAge != "" && d.RequiredAge == state.NextAge:
			rank[key] = 2
		default:
			continue
		}
		keys = append(keys, key)
	}
	return byRank(keys, func(k string) int { return rank[k] })
}

// techAffordable reports whether the knowledge to start t is in store.
func techAffordable(state game.GameState, t game.TechState) bool {
	return state.Resources["knowledge"].Amount >= t.Cost
}

// availableTechKeys is what `research` suggests: techs that can start now,
// the affordable ones first.
func availableTechKeys(state game.GameState) []string {
	var keys []string
	for key, ts := range state.Research.Techs {
		if ts.Available {
			keys = append(keys, key)
		}
	}
	return byRank(keys, func(k string) int {
		if techAffordable(state, state.Research.Techs[k]) {
			return 0
		}
		return 1
	})
}

// plannableTechKeys is what `plan research` takes: unresearched techs of
// this age, an earlier one or the next that are neither in progress nor
// planned already. Prerequisites may still be missing; they can be planned
// first. Suggested in the order research would take them: available and
// affordable, available, then the rest.
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
	return byRank(keys, func(k string) int {
		ts := state.Research.Techs[k]
		switch {
		case ts.Available && techAffordable(state, ts):
			return 0
		case ts.Available:
			return 1
		}
		return 2
	})
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

// wonderNeedKeys is what `wonder collect` suggests: the resources the current
// age's unbuilt wonder still needs, the ones on hand first.
func wonderNeedKeys(state game.GameState) []string {
	w := state.CurrentAgeWonderKey
	if w == "" {
		return nil
	}
	bank := state.Buildings[w].WonderBank
	var keys []string
	for res, need := range config.BuildingByKey()[w].BaseCost {
		if need-bank[res] > 0.001 {
			keys = append(keys, res)
		}
	}
	return byRank(keys, func(k string) int {
		if state.Resources[k].Amount > 0 {
			return 0
		}
		return 1
	})
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

// workerBuildingKeys is what `assign` suggests: built buildings that take
// workers, the ones with a free slot first.
func workerBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Count > 0 && bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
			keys = append(keys, key)
		}
	}
	return byRank(keys, func(k string) int {
		bs := state.Buildings[k]
		if bs.WorkersAssigned < bs.Count*bs.WorkerCapacity {
			return 0
		}
		return 1
	})
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

// prestigeUpgradeKeys returns the prestige upgrades with a tier left, the
// ones the points on hand can buy first.
func prestigeUpgradeKeys(state game.GameState) []string {
	var keys []string
	for key, u := range state.Prestige.Upgrades {
		if u.NextCost > 0 {
			keys = append(keys, key)
		}
	}
	return byRank(keys, func(k string) int {
		if state.Prestige.Upgrades[k].NextCost <= state.Prestige.Available {
			return 0
		}
		return 1
	})
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

// dealNumbers is what `diplomacy accept <civ>` and `plan deal <civ>` suggest:
// the numbers of the civ's offers still open.
func dealNumbers(state game.GameState, civ string) []string {
	var nums []string
	for _, d := range state.Diplomacy.Factions[civ].Deals {
		if !d.Taken {
			nums = append(nums, strconv.Itoa(d.Num))
		}
	}
	return nums
}

// unlockedThemeKeys returns the themes the active account may switch to,
// in the registry's order.
func unlockedThemeKeys(engine *game.GameEngine) []string {
	acct := themeAccount(engine)
	var keys []string
	for _, t := range theme.All() {
		if themeAvailable(acct, t) {
			keys = append(keys, t.Key)
		}
	}
	return keys
}

// availableSpeedOptions returns the speed multipliers from 1.0 up to the
// current max, in steps of config.WonderSpeedCapStep (the max rises that much
// per wonder built).
func availableSpeedOptions(engine *game.GameEngine) []string {
	maxSpeed := engine.GetMaxSpeed()
	var options []string
	for s := 1.0; s <= maxSpeed; s += config.WonderSpeedCapStep {
		options = append(options, fmt.Sprintf("%.1f", s))
	}
	return options
}

// upgradeableBuildingKeys returns the buildings with an upgrade available,
// the affordable ones first.
func upgradeableBuildingKeys(engine *game.GameEngine) []string {
	afford := map[string]bool{}
	var keys []string
	for _, u := range engine.GetAvailableUpgrades() {
		if _, seen := afford[u.FromKey]; !seen {
			keys = append(keys, u.FromKey)
		}
		afford[u.FromKey] = afford[u.FromKey] || u.CanAfford
	}
	return byRank(keys, func(k string) int {
		if afford[k] {
			return 0
		}
		return 1
	})
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
