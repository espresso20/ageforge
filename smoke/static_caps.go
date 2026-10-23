package smoke

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The caps report: what each bonus pool has earned, age by age, for a player
// who holds everything a run can hold by then.
//
// A pool is the game's "+X%" bonuses on one thing added up: all production,
// or one resource's output. A production pool has no ceiling: it applies in
// full, however large it grows. Techs are in a layer of their own, applied
// after the pools, so what a pool holds is milestone rewards, wonders and
// monuments (which stand for the run) and festivals, awakenings and era
// events (which come and go). This report is the data on how much of that
// a run can hold, and the bound the pacing model's pool is held under.
//
// "Everything a run can hold" is generous on purpose:
//
//   - every milestone, from the first age the Milestone Covenant proves it
//     can be completed in (StaticMilestones);
//   - every wonder of an earlier age (an advance needs its age's wonder, so
//     they stand; the age's own wonder is built at its end);
//   - every monument of the age or an earlier one;
//   - the storage buildings the gates force: config.StorageRuleCopies copies
//     of each age's, from the age they are built in (they boost all
//     production by config.StorageBoostPercent a copy; a town may build more,
//     the price being the only limit, so this is the least a town holds).
//
// A player who plays an age through holds less than that for most of it, so
// the pool a row shows is the most a run can hold by then.
const (
	// FestivalBonus is what a festival adds to all production while it
	// lasts, and PlentyBonus the largest timed boost an era event gives: an
	// Age of Plenty doubles production for its length. (The event named
	// Power Surge is no part of this: it adds electricity, not a bonus.)
	FestivalBonus = 0.20
	PlentyBonus   = 1.00
)

// CapRow is one age of the caps report.
type CapRow struct {
	Age string `json:"age"`
	// The all-production pool's standing parts, as bonuses (0.8 is +80%).
	Milestones float64 `json:"milestones"`
	Wonders    float64 `json:"wonders"`
	Monuments  float64 `json:"monuments"`
	Storage    float64 `json:"storage"`
	// Resources is every resource's own pool that holds anything: its
	// milestone rewards and wonders together.
	Resources map[string]float64 `json:"resources,omitempty"`
}

// All is the all-production pool's standing total, as earned.
func (r CapRow) All() float64 { return r.Milestones + r.Wonders + r.Monuments + r.Storage }

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
		// "knowledge_rate" is the research pool: the scholars' work is not a
		// resource, but its pool follows the same rule as the others.
		if _, isRes := config.ResourceByKey()[res]; !ok || (!isRes && res != "knowledge") {
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
		if ok && b.Category == "storage" {
			for _, e := range b.Effects {
				if e.Type == "bonus" && e.Target == "production_all" {
					v := float64(e.Value * config.StorageRuleCopies)
					from(at, func(r *CapRow) { r.Storage += v })
				}
			}
			continue
		}
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

// writeCaps renders the caps report.
func writeCaps(sb *strings.Builder, rows []CapRow) {
	fmt.Fprintf(sb, "What each bonus pool has earned, age by age, for a player who holds everything a run can hold by then: every milestone from the first age it can be completed in, every earlier age's wonder, every monument, five copies of every storage building. A production pool applies in full, however large it grows, so earned is applied. Techs are not in these pools: a tech's bonus is applied after them and always counts in full. Timed boosts come on top while they last: a festival +%.0f%%, an awakening up to +25%%, an Age of Plenty +%.0f%%; a good era can add +10%% or +15%% for the run.\n\n",
		FestivalBonus*100, PlentyBonus*100)
	sb.WriteString("| age | all production: milestones | wonders | monuments | storage | earned | with a festival | with a festival and an Age of Plenty | knowledge | gold | largest other pool | the pacing model holds |\n|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	pct := func(v float64) string { return fmt.Sprintf("+%.0f%%", v*100) }
	for _, r := range rows {
		other, top := "-", 0.0
		for _, res := range sortedKeys(r.Resources) {
			if res != "knowledge" && res != "gold" && r.Resources[res] > top {
				other, top = res+" "+pct(r.Resources[res]), r.Resources[res]
			}
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", ageName(r.Age),
			pct(r.Milestones), pct(r.Wonders), pct(r.Monuments), pct(r.Storage), pct(r.All()),
			pct(r.All()+FestivalBonus), pct(r.All()+FestivalBonus+PlentyBonus),
			pct(r.Resources["knowledge"]), pct(r.Resources["gold"]), other, pct(config.ProductionAllHeld[r.Age]))
	}
	last := rows[len(rows)-1]
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
	fmt.Fprintf(sb, "\nIn the last age the full set has earned %s of all production, all of it applied; the fullest resource pools are %s.\n", pct(last.All()), strings.Join(tops, ", "))
	fmt.Fprintf(sb, "\nThe pacing model's all-production pool (config.ProductionAllHeld, the last column) is held as earned and applied in full. It must stay at or under the earned column: it is what a well-played game holds, never more than a game can.\n")
}
