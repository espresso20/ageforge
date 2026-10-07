package smoke

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
		case rs.Amount > float64(rs.Storage*(1+1e-9))+1e-6 && !rs.OverCapGrace:
			// Graced stock (Era Mastery's grace rule) may sit over the cap.
			add("over_storage", fmt.Sprintf("%s amount %s is above its storage cap %s", key, num(rs.Amount), num(rs.Storage)))
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
	return append(out, storageProblems(st, defs)...)
}

// storageProblems flags storage that can never grow again in this age, a
// next-age resource requirement no reachable storage can hold, and a required
// building that can't be built or whose last copy can't fit under any
// reachable cap.
func storageProblems(st game.GameState, defs map[string]config.BuildingDef) []problem {
	var out []problem
	caps, stall := storageLadder(st, defs)
	if stall != nil {
		out = append(out, problem{"storage_unreachable", stall.message(st)})
	}
	for _, res := range sortedKeys(st.NextAgeResReqs) {
		need := st.NextAgeResReqs[res]
		if got := caps[res]; need > got {
			out = append(out, problem{"requirement_over_storage",
				fmt.Sprintf("advancing to %s needs %s %s but the most storage reachable in %s is %s",
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
		res, cost, capacity, ok := lastCopyOverStorage(st, bld, defs, caps)
		if !ok {
			continue
		}
		out = append(out, problem{"required_building_over_storage",
			fmt.Sprintf("advancing to %s needs %d %s; copy #%d costs %s %s but the most %s storage reachable in %s is %s",
				st.NextAge, need, bld, need, num(cost), res, res, st.Age, num(capacity))})
	}
	return out
}

// storageStall is storage that can never grow again: this age's storage
// building has copies left, but the next one costs more than the cap it must
// fit under, and the age lock forbids every older storage building. It is
// what a storage loss used to cause (an Endure took both Industrial Depots
// on the way into the Victorian Age, leaving 130M against a 193M vault), and
// what a gate that forced too little storage would cause.
type storageStall struct {
	key       string  // this age's storage building
	res       string  // the resource whose cap its next copy overflows most
	cost, cap float64 // that copy's price in res, and the res cap
}

func (s *storageStall) message(st game.GameState) string {
	return fmt.Sprintf("storage can never grow in %s: the next %s costs %s %s, over the %s cap, and no older storage can be built",
		st.Age, s.key, num(s.cost), s.res, num(s.cap))
}

// storageLadder is the most storage reachable in this age, built a copy at a
// time the way a player has to: a copy counts only once its price fits under
// the caps the copies before it reached, so storage whose next copy costs
// more than today's cap adds nothing. Copies under construction are paid for
// and count from the start. Only this age's buildings can be bought (the age
// lock). A cap an uncapped building raises (MaxCount 0) is +Inf once its next
// copy fits. stall is set when this age's storage still has copies left but
// not one can be bought.
func storageLadder(st game.GameState, defs map[string]config.BuildingDef) (map[string]float64, *storageStall) {
	caps := make(map[string]float64, len(st.Resources))
	for _, k := range sortedKeys(st.Resources) {
		caps[k] = st.Resources[k].Storage
	}
	// Storage grows with Era Mastery's k, a building's share included.
	k := st.Mastery.K
	if k <= 0 {
		k = 1
	}
	raise := func(def config.BuildingDef, n float64) {
		for _, e := range def.Effects {
			if e.Type != "storage" || e.Value <= 0 {
				continue
			}
			add := float64(float64(e.Value*n) * k)
			if e.Target != "all" {
				caps[e.Target] += add
				continue
			}
			for _, r := range sortedKeys(caps) {
				caps[r] += add
			}
		}
	}
	keyByName := make(map[string]string, len(st.Buildings))
	for _, k := range sortedKeys(st.Buildings) {
		keyByName[st.Buildings[k].Name] = k
	}
	queued := map[string]int{}
	for _, q := range st.BuildQueue {
		if k, ok := keyByName[q.Name]; ok {
			queued[k]++
			raise(defs[k], 1)
		}
	}

	type rung struct {
		key  string
		def  config.BuildingDef
		next map[string]float64 // price of the next copy
		left int                // copies left under MaxCount; -1 when uncapped
	}
	var rungs []*rung
	for _, key := range sortedKeys(st.Buildings) {
		bs, def := st.Buildings[key], defs[key]
		if !bs.Unlocked || bs.IsLegacy || def.Category == "wonder" {
			continue
		}
		if def.RequiredAge != "" && def.RequiredAge != st.Age {
			continue
		}
		stores := false
		for _, e := range def.Effects {
			stores = stores || (e.Type == "storage" && e.Value > 0)
		}
		if !stores {
			continue
		}
		left := -1
		if def.MaxCount > 0 {
			if left = def.MaxCount - bs.Count - queued[key]; left <= 0 {
				continue
			}
		}
		next := make(map[string]float64, len(bs.NextCost))
		for r, c := range bs.NextCost {
			next[r] = c
		}
		rungs = append(rungs, &rung{key: key, def: def, next: next, left: left})
	}
	fits := func(cost map[string]float64) bool {
		for r, c := range cost {
			if c > caps[r] {
				return false
			}
		}
		return true
	}

	grew := false
	for progress := true; progress; {
		progress = false
		for _, g := range rungs {
			if g.left == 0 || !fits(g.next) {
				continue
			}
			if g.left < 0 {
				// Uncapped: every later copy fits too, so its caps have no top.
				for _, e := range g.def.Effects {
					if e.Type != "storage" || e.Value <= 0 {
						continue
					}
					for _, r := range sortedKeys(caps) {
						if e.Target == "all" || e.Target == r {
							caps[r] = math.Inf(1)
						}
					}
				}
				g.left = 0
			} else {
				raise(g.def, 1)
				g.left--
				for r, c := range g.next {
					g.next[r] = float64(c * g.def.CostScale)
				}
			}
			grew = grew || g.def.Category == "storage"
			progress = true
		}
	}
	if grew {
		return caps, nil
	}
	for _, g := range rungs {
		if g.def.Category != "storage" || g.left == 0 {
			continue
		}
		s := &storageStall{key: g.key}
		for _, r := range sortedKeys(g.next) {
			if c := g.next[r]; c > caps[r] && (s.res == "" || c/caps[r] > s.cost/s.cap) {
				s.res, s.cost, s.cap = r, c, caps[r]
			}
		}
		return caps, s
	}
	return caps, nil
}

// lastCopyOverStorage checks whether the last copy of bld the next age asks
// for costs more of some resource than caps, the most storage reachable in
// this age (storageLadder). Costs use the current build_cost multiplier
// (already in NextCost).
func lastCopyOverStorage(st game.GameState, bld string, defs map[string]config.BuildingDef, caps map[string]float64) (res string, cost, capacity float64, over bool) {
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
	mult := detmath.Pow(defs[bld].CostScale, steps)
	for _, r := range sortedKeys(bs.NextCost) {
		c := bs.NextCost[r] * mult
		if got := caps[r]; c > got {
			return r, c, got, true
		}
	}
	return "", 0, 0, false
}

// Blockers lists what stands between st and the next age, one clause per
// unmet requirement.
func Blockers(st game.GameState) string {
	var out []string
	defs := st.Ruleset().BuildingMap()
	caps, stall := storageLadder(st, defs)
	if stall != nil {
		// The root cause when it is there: nothing below can be fixed.
		out = append(out, fmt.Sprintf("storage stuck (next %s costs %s %s, over the %s cap)", stall.key, num(stall.cost), stall.res, num(stall.cap)))
	}
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
			if res, c, capacity, over := lastCopyOverStorage(st, bld, defs, caps); over {
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

// num formats a number compactly: whole below ten thousand, and from there
// the game's own format (textfmt.Number: 12.5K, 3.4M, 1.56T, 8.9Q), so the
// report reads as the game does. It used to stop at T and printed the Cosmic
// Era's quadrillions as 8.9e+03T.
func num(v float64) string {
	switch {
	case bad(v):
		return fmt.Sprint(v)
	case math.Abs(v) >= 1e4:
		return textfmt.Number(v)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}
