package smoke

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

func bad(v float64) bool { return math.IsNaN(v) || math.IsInf(v, 0) }

// checkInvariants runs the per-sweep checks against one snapshot.
func (r *runner) checkInvariants(st game.GameState) {
	for _, p := range invariantProblems(st, r.bot.defs) {
		r.anomaly(KindInvariant, p.check, p.msg, st, true)
	}
}

// checkStorageFeasible flags a next-age requirement that no amount of
// building in this age can satisfy.
func (r *runner) checkStorageFeasible(st game.GameState) {
	for _, p := range storageProblems(st, r.bot.defs) {
		r.anomaly(KindInvariant, p.check, p.msg, st, true)
	}
}

// invariantProblems is every invariant st breaks: non-finite or negative
// resources, amounts over their cap, bad multipliers, worker bookkeeping,
// and requirements no storage can hold. Scenarios other than progression
// (offline, fuzz, saveload) share it.
func invariantProblems(st game.GameState, defs map[string]config.BuildingDef) []problem {
	var out []problem
	add := func(check, msg string) { out = append(out, problem{check, msg}) }
	for _, key := range sortedKeys(st.Resources) {
		rs := st.Resources[key]
		switch {
		case bad(rs.Amount) || bad(rs.Rate) || bad(rs.Storage):
			add("nan_resource", fmt.Sprintf("%s has a non-finite value: amount=%v rate=%v storage=%v", key, rs.Amount, rs.Rate, rs.Storage))
		case rs.Amount < 0:
			add("negative_resource", fmt.Sprintf("%s is negative: %v", key, rs.Amount))
		case rs.Amount > rs.Storage*(1+1e-9)+1e-6:
			add("over_storage", fmt.Sprintf("%s amount %s is above its storage cap %s", key, num(rs.Amount), num(rs.Storage)))
		}
	}
	for _, v := range []struct {
		name string
		val  float64
	}{
		{"morale", st.Morale}, {"morale_cap", st.MoraleCap}, {"morale_multiplier", st.MoraleMultiplier},
		{"tick_speed_bonus", st.TickSpeedBonus}, {"succumb_research_bonus", st.SuccumbResearchBonus},
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
	return append(out, storageProblems(st, defs)...)
}

// storageProblems flags a next-age resource requirement that no amount of
// storage building in this age can hold, and a required building that can't
// be built or whose last copy can't fit under any reachable cap.
func storageProblems(st game.GameState, defs map[string]config.BuildingDef) []problem {
	var out []problem
	for _, res := range sortedKeys(st.NextAgeResReqs) {
		need := st.NextAgeResReqs[res]
		if got := achievableStorage(st, res, defs); need > got {
			out = append(out, problem{"requirement_over_storage",
				fmt.Sprintf("advancing to %s needs %s %s but the most storage buildable in %s is %s",
					st.NextAge, num(need), res, st.Age, num(got))})
		}
	}
	for _, bld := range sortedKeys(st.NextAgeBldReqs) {
		need, bs, def := st.NextAgeBldReqs[bld], st.Buildings[bld], defs[bld]
		if bs.Count < need && (bs.IsLegacy || (def.RequiredAge != "" && def.RequiredAge != st.Age)) {
			out = append(out, problem{"required_building_unbuildable",
				fmt.Sprintf("advancing to %s needs %d %s but only %d exist and %s can't be built in %s",
					st.NextAge, need, bld, bs.Count, bld, st.Age)})
			continue
		}
		res, cost, capacity, ok := lastCopyOverStorage(st, bld, defs)
		if !ok {
			continue
		}
		out = append(out, problem{"required_building_over_storage",
			fmt.Sprintf("advancing to %s needs %d %s; copy #%d costs %s %s but the most %s storage buildable in %s is %s",
				st.NextAge, need, bld, need, num(cost), res, res, st.Age, num(capacity))})
	}
	return out
}

// achievableStorage is the highest cap reachable for res in the current age:
// today's cap plus every storage copy still buildable here. +Inf when an
// uncapped storage building covers it.
func achievableStorage(st game.GameState, res string, defs map[string]config.BuildingDef) float64 {
	capacity := st.Resources[res].Storage
	for _, key := range sortedKeys(st.Buildings) {
		bs := st.Buildings[key]
		def := defs[key]
		if !bs.Unlocked || bs.IsLegacy || def.Category == "wonder" {
			continue
		}
		if def.RequiredAge != "" && def.RequiredAge != st.Age {
			continue
		}
		for _, e := range def.Effects {
			if e.Type != "storage" || e.Value <= 0 || (e.Target != res && e.Target != "all") {
				continue
			}
			if def.MaxCount == 0 {
				return math.Inf(1)
			}
			if left := def.MaxCount - bs.Count; left > 0 {
				capacity += e.Value * float64(left)
			}
		}
	}
	return capacity
}

// lastCopyOverStorage checks whether the last copy of bld the next age asks
// for costs more of some resource than the most storage reachable in this
// age. Costs use the current build_cost multiplier (already in NextCost).
func lastCopyOverStorage(st game.GameState, bld string, defs map[string]config.BuildingDef) (res string, cost, capacity float64, over bool) {
	need := st.NextAgeBldReqs[bld]
	bs := st.Buildings[bld]
	if bs.Count >= need || len(bs.NextCost) == 0 {
		return "", 0, 0, false
	}
	queued := 0
	for _, q := range st.BuildQueue {
		if q.Name == bs.Name {
			queued++
		}
	}
	steps := float64(need - 1 - bs.Count - queued)
	if steps < 0 {
		steps = 0
	}
	mult := math.Pow(defs[bld].CostScale, steps)
	for _, r := range sortedKeys(bs.NextCost) {
		c := bs.NextCost[r] * mult
		if got := achievableStorage(st, r, defs); c > got {
			return r, c, got, true
		}
	}
	return "", 0, 0, false
}

// Blockers lists what stands between st and the next age, one clause per
// unmet requirement.
func Blockers(st game.GameState) string {
	var out []string
	for _, res := range sortedKeys(st.NextAgeResReqs) {
		need := st.NextAgeResReqs[res]
		rs := st.Resources[res]
		if rs.Amount >= need {
			continue
		}
		note := ""
		switch {
		case rs.Storage < need:
			note = ", cap below requirement"
		case rs.Rate <= 0:
			note = ", not being produced"
		}
		out = append(out, fmt.Sprintf("%s %s/%s (cap %s, %+.3g/t%s)", res, num(rs.Amount), num(need), num(rs.Storage), rs.Rate, note))
	}
	for _, bld := range sortedKeys(st.NextAgeBldReqs) {
		need := st.NextAgeBldReqs[bld]
		bs := st.Buildings[bld]
		if bs.Count < need {
			note := ""
			if res, c, capacity, over := lastCopyOverStorage(st, bld, config.BuildingByKey()); over {
				note = fmt.Sprintf("; copy #%d costs %s %s, over the %s storage reachable this age", need, num(c), res, num(capacity))
			}
			out = append(out, fmt.Sprintf("%s %d/%d (next costs %s%s)", bld, bs.Count, need, costStr(bs.NextCost), note))
		}
	}
	if w := st.CurrentAgeWonderKey; w != "" {
		out = append(out, fmt.Sprintf("wonder %s (bank %s)", w, bankStr(st, w)))
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
		fmt.Fprintf(&sb, "  %-22s %12s / %-12s %+.4g/t\n", k, num(rs.Amount), num(rs.Storage), rs.Rate)
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
func botNoise(msg string) bool {
	for _, p := range []string{"Gathered ", "Assigned ", "Unassigned ", "Recruited ", "Banked ", "Started building ", "Queued "} {
		if strings.HasPrefix(msg, p) {
			return true
		}
	}
	return false
}

func bankStr(st game.GameState, w string) string {
	def := config.BuildingByKey()[w]
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

// num formats a number compactly (1.2K, 3.4M, ...).
func num(v float64) string {
	a := math.Abs(v)
	switch {
	case bad(v):
		return fmt.Sprint(v)
	case a >= 1e12:
		return fmt.Sprintf("%.3gT", v/1e12)
	case a >= 1e9:
		return fmt.Sprintf("%.3gB", v/1e9)
	case a >= 1e6:
		return fmt.Sprintf("%.3gM", v/1e6)
	case a >= 1e4:
		return fmt.Sprintf("%.3gK", v/1e3)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}
