package smoke

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The caps report: how close each bonus pool gets to its clamp, age by age,
// for a player who holds everything a run can hold by then.
//
// A pool is the game's "+X%" bonuses on one thing added up: all production,
// or one resource's output. The engine applies at most +200% of a pool (x3).
// Techs used to fill these pools on their own. They are in a layer of their
// own now, applied after the clamp, so what is left in a pool is milestone
// rewards, wonders and monuments (which stand for the run) and festivals,
// awakenings and era events (which come and go). This report is the data on
// whether the clamp still bites.
//
// "Everything a run can hold" is generous on purpose:
//
//   - every milestone, from the first age the Milestone Covenant proves it
//     can be completed in (StaticMilestones);
//   - every wonder of an earlier age (an advance needs its age's wonder, so
//     they stand; the age's own wonder is built at its end);
//   - every monument of the age or an earlier one.
//
// A player who plays an age through holds less than that for most of it, so
// the age a pool reaches its clamp here is the earliest it can.
const (
	// PoolCapBonus is the most of a production pool the engine applies:
	// +200%, a factor of 3 (config.ProductionAllCap).
	PoolCapBonus = config.ProductionAllCap - 1
	// FestivalBonus is what a festival adds to all production while it
	// lasts, and SurgeBonus the largest timed boost an era event gives (the
	// power surge doubles production for its length).
	FestivalBonus = 0.20
	SurgeBonus    = 1.00
)

// CapRow is one age of the caps report.
type CapRow struct {
	Age string `json:"age"`
	// The all-production pool's standing parts, as bonuses (0.8 is +80%).
	Milestones float64 `json:"milestones"`
	Wonders    float64 `json:"wonders"`
	Monuments  float64 `json:"monuments"`
	// Resources is every resource's own pool that holds anything: its
	// milestone rewards and wonders together.
	Resources map[string]float64 `json:"resources,omitempty"`
}

// All is the all-production pool's standing total.
func (r CapRow) All() float64 { return r.Milestones + r.Wonders + r.Monuments }

// StaticCaps computes the caps report for the core ruleset, one row per age.
func StaticCaps() []CapRow {
	ages := config.AgeOrder()
	idx := config.AgePositions(ages)
	rows := make([]CapRow, len(ages))
	for i, a := range ages {
		rows[i] = CapRow{Age: a, Resources: map[string]float64{}}
	}
	from := func(at int, add func(r *CapRow)) {
		for i := at; i >= 0 && i < len(rows); i++ {
			add(&rows[i])
		}
	}
	pool := func(target string) (res string, all, ok bool) {
		if target == "production_all" {
			return "", true, true
		}
		res, ok = strings.CutSuffix(target, "_rate")
		if _, isRes := config.ResourceByKey()[res]; !ok || !isRes {
			return "", false, false
		}
		return res, false, true
	}

	_, reach := StaticMilestones()
	defs := config.MilestoneByKey()
	for _, m := range reach {
		at, ok := idx[m.Earliest]
		if !ok {
			continue
		}
		for _, e := range defs[m.Key].Rewards {
			if e.Type != "permanent_bonus" {
				continue
			}
			res, all, ok := pool(e.Target)
			if !ok {
				continue
			}
			from(at, func(r *CapRow) {
				if all {
					r.Milestones += e.Value
				} else {
					r.Resources[res] += e.Value
				}
			})
		}
	}
	for _, b := range config.BaseBuildings() {
		at, ok := idx[b.RequiredAge]
		if !ok || (b.Category != "wonder" && b.Category != "monument") {
			continue
		}
		if b.Category == "wonder" {
			at++ // it stands from the next age on
		}
		for _, e := range b.Effects {
			if e.Type != "bonus" {
				continue
			}
			res, all, ok := pool(e.Target)
			if !ok {
				continue
			}
			from(at, func(r *CapRow) {
				switch {
				case !all:
					r.Resources[res] += e.Value
				case b.Category == "wonder":
					r.Wonders += e.Value
				default:
					r.Monuments += e.Value
				}
			})
		}
	}
	return rows
}

// CapReachedIn is the first age the standing pool read by of reaches the
// clamp with extra added on top ("" when it never does).
func CapReachedIn(rows []CapRow, extra float64, of func(CapRow) float64) string {
	for _, r := range rows {
		if of(r)+extra >= PoolCapBonus-1e-9 {
			return r.Age
		}
	}
	return ""
}

// writeCaps renders the caps report.
func writeCaps(sb *strings.Builder, rows []CapRow) {
	fmt.Fprintf(sb, "How close each bonus pool gets to its clamp (the engine applies at most +%.0f%% of a pool, x%g), age by age, for a player who holds everything a run can hold by then: every milestone from the first age it can be completed in, every earlier age's wonder, every monument. Techs are not in these pools: a tech's bonus is applied after the clamp and always counts. Timed boosts come on top while they last: a festival +%.0f%%, an awakening up to +25%%, a power surge +%.0f%%; a good era can add +10%% or +15%% for the run.\n\n",
		PoolCapBonus*100, config.ProductionAllCap, FestivalBonus*100, SurgeBonus*100)
	sb.WriteString("| age | all production: milestones | wonders | monuments | standing | of the clamp | with a festival | with a festival and a surge | knowledge | gold | tightest other pool |\n|---|---|---|---|---|---|---|---|---|---|---|\n")
	pct := func(v float64) string { return fmt.Sprintf("+%.0f%%", v*100) }
	mark := func(v float64) string {
		if v >= PoolCapBonus-1e-9 {
			return pct(v) + " (clamped)"
		}
		return pct(v)
	}
	for _, r := range rows {
		other, top := "-", 0.0
		for _, res := range sortedKeys(r.Resources) {
			if res != "knowledge" && res != "gold" && r.Resources[res] > top {
				other, top = res+" "+pct(r.Resources[res]), r.Resources[res]
			}
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %.0f%% | %s | %s | %s | %s | %s |\n", ageName(r.Age),
			pct(r.Milestones), pct(r.Wonders), pct(r.Monuments), mark(r.All()), 100*r.All()/PoolCapBonus,
			mark(r.All()+FestivalBonus), mark(r.All()+FestivalBonus+SurgeBonus),
			mark(r.Resources["knowledge"]), mark(r.Resources["gold"]), other)
	}
	all := func(r CapRow) float64 { return r.All() }
	when := func(age string) string {
		if age == "" {
			return "never"
		}
		return "from the " + ageName(age) + " Age"
	}
	fmt.Fprintf(sb, "\nAll production reaches its clamp %s on what stands alone, %s with a festival running, and %s with a festival and a power surge. ",
		when(CapReachedIn(rows, 0, all)), when(CapReachedIn(rows, FestivalBonus, all)), when(CapReachedIn(rows, FestivalBonus+SurgeBonus, all)))
	var reached []string
	last := rows[len(rows)-1]
	for _, res := range sortedKeys(last.Resources) {
		if age := CapReachedIn(rows, 0, func(r CapRow) float64 { return r.Resources[res] }); age != "" {
			reached = append(reached, fmt.Sprintf("%s (%s)", res, when(age)))
		}
	}
	if len(reached) == 0 {
		tops := make([]string, 0, len(last.Resources))
		for _, res := range sortedKeys(last.Resources) {
			tops = append(tops, res)
		}
		sort.SliceStable(tops, func(i, j int) bool { return last.Resources[tops[i]] > last.Resources[tops[j]] })
		if len(tops) > 3 {
			tops = tops[:3]
		}
		for i, res := range tops {
			tops[i] = fmt.Sprintf("%s %s", res, pct(last.Resources[res]))
		}
		fmt.Fprintf(sb, "No resource's own pool reaches its clamp in any age; the fullest at the end are %s.\n", strings.Join(tops, ", "))
	} else {
		fmt.Fprintf(sb, "A resource's own pool reaches its clamp for: %s.\n", strings.Join(reached, ", "))
	}
	fmt.Fprintf(sb, "\nThe pacing model's all-production pool (config.ProductionAllHeld) must stay at or under the standing column: it is what a well-played game holds, never more than a game can.\n")
}
