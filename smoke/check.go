package smoke

import (
	"fmt"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// checkInvariants runs the per-sweep checks against one snapshot.
func (r *runner) checkInvariants(st game.GameState) {
	for _, p := range invariantProblems(st, r.bot.defs) {
		r.anomaly(KindInvariant, p.check, p.msg, st, true)
	}
}

// checkRequiredBuildings flags a next-age requirement that no amount of
// building in this age can satisfy.
func (r *runner) checkRequiredBuildings(st game.GameState) {
	for _, p := range requiredBuildingProblems(st, r.bot.defs) {
		r.anomaly(KindInvariant, p.check, p.msg, st, true)
	}
}

// invariantProblems is every invariant st breaks: non-finite or negative
// resources, bad multipliers, worker bookkeeping,
// and requirements nothing can meet. Scenarios other than progression
// (offline, fuzz, saveload) share it.
func invariantProblems(st game.GameState, defs map[string]config.BuildingDef) []problem {
	var out []problem
	add := func(check, msg string) { out = append(out, problem{check, msg}) }
	for _, key := range sortedKeys(st.Resources) {
		rs := st.Resources[key]
		switch {
		case bad(rs.Amount) || bad(rs.Rate):
			add("nan_resource", fmt.Sprintf("%s has a non-finite value: amount=%v rate=%v", key, rs.Amount, rs.Rate))
		case rs.Amount < 0:
			add("negative_resource", fmt.Sprintf("%s is negative: %v", key, rs.Amount))
		}
	}
	for _, v := range []struct {
		name string
		val  float64
	}{
		{"morale", st.Morale}, {"morale_cap", st.MoraleCap}, {"morale_multiplier", st.MoraleMultiplier},
		{"tick_speed_bonus", st.TickSpeedBonus}, {"succumb_research_factor", st.SuccumbResearchFactor},
		{"prestige_passive_bonus", st.Prestige.PassiveBonus},
	} {
		if bad(v.val) {
			add("nan_value", fmt.Sprintf("%s is non-finite: %v", v.name, v.val))
		}
	}
	for _, k := range sortedKeys(st.PermanentBonuses) {
		if bad(st.PermanentBonuses[k]) {
			add("nan_value", fmt.Sprintf("permanent bonus %s is non-finite", k))
		}
	}
	for _, m := range st.Modifiers {
		if bad(m.Value) {
			add("nan_value", fmt.Sprintf("modifier %+v is non-finite", m))
		}
	}
	if st.TickIntervalMs <= 0 {
		add("tick_interval", fmt.Sprintf("tick interval is %dms", st.TickIntervalMs))
	}

	ws := st.Workers
	assigned := 0
	for _, key := range sortedKeys(st.Buildings) {
		bs := st.Buildings[key]
		assigned += bs.WorkersAssigned
		if bs.WorkersAssigned < 0 {
			add("negative_assignment", fmt.Sprintf("%s has %d workers assigned", key, bs.WorkersAssigned))
		}
		if capacity := bs.Count * bs.WorkerCapacity; bs.WorkersAssigned > capacity {
			add("over_capacity",
				fmt.Sprintf("%s has %d workers assigned but %d copies x %d slots = %d", key, bs.WorkersAssigned, bs.Count, bs.WorkerCapacity, capacity))
		}
		if bs.Count < 0 {
			add("negative_building", fmt.Sprintf("%s count is %d", key, bs.Count))
		}
	}
	if assigned > ws.TotalPop {
		add("over_population", fmt.Sprintf("%d workers assigned but population is %d", assigned, ws.TotalPop))
	}
	if ws.TotalIdle < 0 {
		add("negative_idle", fmt.Sprintf("idle workers is %d", ws.TotalIdle))
	}
	return append(out, requiredBuildingProblems(st, defs)...)
}

// requiredBuildingProblems flags a building the next age asks for that can no
// longer be built: its age has passed (or it was upgraded away) and fewer
// copies stand than the gate wants. Stock has no cap, so a price is never
// "over the storage" and nothing else can strand a required copy.
func requiredBuildingProblems(st game.GameState, defs map[string]config.BuildingDef) []problem {
	var out []problem
	for _, bld := range sortedKeys(st.NextAgeBldReqs) {
		need, bs, def := st.NextAgeBldReqs[bld], st.Buildings[bld], defs[bld]
		if bs.Count < need && (bs.IsLegacy || (def.RequiredAge != "" && def.RequiredAge != st.Age)) {
			out = append(out, problem{"required_building_unbuildable",
				fmt.Sprintf("advancing to %s needs %d %s but only %d exist and %s can't be built in %s",
					st.NextAge, need, bld, bs.Count, bld, st.Age)})
		}
	}
	return out
}

// Blockers lists what stands between st and the next age, one clause per
// unmet requirement.
func Blockers(st game.GameState) string {
	var out []string
	for _, bld := range sortedKeys(st.NextAgeBldReqs) {
		need := st.NextAgeBldReqs[bld]
		bs := st.Buildings[bld]
		if bs.Count < need {
			out = append(out, fmt.Sprintf("%s %d/%d (next costs %s)", bld, bs.Count, need, costStr(bs.NextCost)))
		}
	}
	if w := st.CurrentAgeWonderKey; w != "" {
		note := ""
		for _, key := range keystoneChain(st) {
			note += fmt.Sprintf("; needs %s researched (%s of work)", key, num(st.Research.Techs[key].Cost))
		}
		out = append(out, fmt.Sprintf("wonder %s (bank %s%s)", w, bankStr(st, w), note))
	}
	if st.PendingCatastrophe != "" {
		out = append(out, "pending catastrophe "+st.PendingCatastrophe)
	}
	if len(out) == 0 {
		if st.NextAge == "" {
			return "nothing (final age)"
		}
		return "nothing (requirements met)"
	}
	return strings.Join(out, "; ")
}

// Dump renders a snapshot for a report.
func Dump(st game.GameState, cycle int, sim time.Duration) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "cycle %d, age %s, game tick %d, %s simulated at 1x, prestige level %d\n",
		cycle, st.Age, st.Tick, sim.Round(time.Second), st.Prestige.Level)
	fmt.Fprintf(&sb, "next age: %s\n", st.NextAge)
	fmt.Fprintf(&sb, "blocked on: %s\n", Blockers(st))
	ws := st.Workers
	fmt.Fprintf(&sb, "population %d/%d, idle %d, food drain %.3g/t, morale %.2f\n", ws.TotalPop, ws.MaxPop, ws.TotalIdle, ws.FoodDrain, st.Morale)
	if st.PendingCatastrophe != "" {
		fmt.Fprintf(&sb, "pending catastrophe: %s\n", st.PendingCatastrophe)
	}
	if st.Research.CurrentTech != "" {
		fmt.Fprintf(&sb, "researching %s (%d ticks left), %d techs done\n", st.Research.CurrentTech, st.Research.TicksLeft, st.Research.TotalResearched)
	} else {
		fmt.Fprintf(&sb, "no research running, %d techs done\n", st.Research.TotalResearched)
	}
	sb.WriteString("resources:\n")
	for _, k := range sortedKeys(st.Resources) {
		rs := st.Resources[k]
		if !rs.Unlocked {
			continue
		}
		fmt.Fprintf(&sb, "  %-22s %12s / %-12s %+.4g/t\n", k, num(rs.Amount), num(goodStockOf(rules.Core(), st, k)), rs.Rate)
	}
	sb.WriteString("buildings:\n")
	for _, k := range sortedKeys(st.Buildings) {
		bs := st.Buildings[k]
		if bs.Count == 0 && bs.RuinCount == 0 {
			continue
		}
		flags := ""
		if bs.IsLegacy {
			flags += " legacy"
		}
		if bs.RuinCount > 0 {
			flags += fmt.Sprintf(" ruins=%d", bs.RuinCount)
		}
		fmt.Fprintf(&sb, "  %-28s x%-4d workers %d/%d%s\n", k, bs.Count, bs.WorkersAssigned, bs.Count*bs.WorkerCapacity, flags)
	}
	if len(st.BuildQueue) > 0 {
		fmt.Fprintf(&sb, "build queue: %d item(s)\n", len(st.BuildQueue))
	}
	if len(st.ActiveEvents) > 0 {
		sb.WriteString("active events:")
		for _, e := range st.ActiveEvents {
			fmt.Fprintf(&sb, " %s(%d)", e.Key, e.TicksLeft)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("last log lines:\n")
	n := 0
	for i := len(st.Log) - 1; i >= 0 && n < 12; i-- {
		if st.Log[i].Type == "debug" || botNoise(st.Log[i].Message) {
			continue
		}
		fmt.Fprintf(&sb, "  [%d %s] %s\n", st.Log[i].Tick, st.Log[i].Type, st.Log[i].Message)
		n++
	}
	return sb.String()
}

// botNoise reports log lines that only echo the bot's own routine actions.
// keystoneChain is the techs this age's unbuilt wonder still waits for: its
// keystone and what that stands on, what it stands on first. nil when the
// wonder is built, needs no tech, or has it.
func keystoneChain(st game.GameState) []string {
	w := st.CurrentAgeWonderKey
	if w == "" || st.Buildings[w].NeedsTech == "" {
		return nil
	}
	return techChain(st.Ruleset(), st.Research.Techs, st.Buildings[w].NeedsTech, map[string]bool{}, nil)
}

func botNoise(msg string) bool {
	for _, p := range []string{"Gathered ", "Assigned ", "Unassigned ", "Recruited ", "Banked ", "Started building ", "Queued "} {
		if strings.HasPrefix(msg, p) {
			return true
		}
	}
	return false
}

func bankStr(st game.GameState, w string) string {
	def, _ := st.Ruleset().Building(w)
	bank := st.Buildings[w].WonderBank
	var parts []string
	for _, res := range sortedKeys(def.BaseCost) {
		parts = append(parts, fmt.Sprintf("%s %s/%s", res, num(bank[res]), num(def.BaseCost[res])))
	}
	return strings.Join(parts, ", ")
}

func costStr(c map[string]float64) string {
	var parts []string
	for _, k := range sortedKeys(c) {
		parts = append(parts, fmt.Sprintf("%s %s", k, num(c[k])))
	}
	return strings.Join(parts, ", ")
}
