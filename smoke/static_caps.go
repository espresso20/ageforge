package smoke

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The caps report: what each bonus pool has earned and what it applies, age
// by age, for a player who holds everything a run can hold by then.
//
// A pool is the game's "+X%" bonuses on one thing added up: all production,
// or one resource's output. A production pool follows the soft cap
// (config.ProductionSoftCap): it applies in full up to its knee, +200%, and
// a quarter of every point past it. Until that rule the pools stopped dead
// at +200% (x3), and this report showed a full set of milestones and wonders
// earning +661% by the last age for +200% applied. Techs are in a layer of
// their own, applied after the pools, so what a pool holds is milestone
// rewards, wonders and monuments (which stand for the run) and festivals,
// awakenings and era events (which come and go). This report is the data on
// how much of that counts.
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
// the age a pool passes its knee here is the earliest it can.
const (
	// PoolKnee is the most of a production pool the engine applies in
	// full: +200% (config.ProductionKnee). Past it a point counts
	// config.ProductionPastKnee, a quarter.
	PoolKnee = config.ProductionKnee
	// FestivalBonus is what a festival adds to all production while it
	// lasts, and PlentyBonus the largest timed boost an era event gives: an
	// Age of Plenty doubles production for its length. (The event named
	// Power Surge is no part of this: it adds electricity, not a bonus.)
	FestivalBonus = 0.20
	PlentyBonus   = 1.00
)

// PoolApplied is what a production pool that has earned bonuses in all
// applies under the soft cap.
func PoolApplied(earned float64) float64 { return config.ProductionSoftCap().Applied(earned) }

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

// All is the all-production pool's standing total, as earned.
func (r CapRow) All() float64 { return r.Milestones + r.Wonders + r.Monuments }

// Applied is what the engine applies of All under the soft cap.
func (r CapRow) Applied() float64 { return PoolApplied(r.All()) }

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

// KneePassedIn is the first age the standing pool read by of is past the
// knee with extra added on top ("" when it never is).
func KneePassedIn(rows []CapRow, extra float64, of func(CapRow) float64) string {
	for _, r := range rows {
		if of(r)+extra > PoolKnee+1e-9 {
			return r.Age
		}
	}
	return ""
}

// writeCaps renders the caps report.
func writeCaps(sb *strings.Builder, rows []CapRow) {
	fmt.Fprintf(sb, "What each bonus pool has earned and what it applies, age by age, for a player who holds everything a run can hold by then: every milestone from the first age it can be completed in, every earlier age's wonder, every monument. A production pool applies in full up to +%.0f%% and %.0f%% of every point past it (config.ProductionSoftCap), so nothing earned is lost; before that rule it applied +%.0f%% at most. Techs are not in these pools: a tech's bonus is applied after them and always counts in full. Timed boosts come on top while they last: a festival +%.0f%%, an awakening up to +25%%, an Age of Plenty +%.0f%%; a good era can add +10%% or +15%% for the run. A cell with an arrow reads earned → applied.\n\n",
		PoolKnee*100, config.ProductionPastKnee*100, PoolKnee*100, FestivalBonus*100, PlentyBonus*100)
	sb.WriteString("| age | all production: milestones | wonders | monuments | earned | applied | share that counts | with a festival | with a festival and an Age of Plenty | knowledge | gold | tightest other pool | the pacing model holds |\n|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	pct := func(v float64) string { return fmt.Sprintf("+%.0f%%", v*100) }
	// both is a pool as earned, and what it applies when that is less.
	both := func(earned float64) string {
		if earned > PoolKnee+1e-9 {
			return pct(earned) + " → " + pct(PoolApplied(earned))
		}
		return pct(earned)
	}
	for _, r := range rows {
		other, top := "-", 0.0
		for _, res := range sortedKeys(r.Resources) {
			if res != "knowledge" && res != "gold" && r.Resources[res] > top {
				other, top = res+" "+both(r.Resources[res]), r.Resources[res]
			}
		}
		share := 100.0
		if r.All() > 0 {
			share = 100 * r.Applied() / r.All()
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %.0f%% | %s | %s | %s | %s | %s | %s |\n", ageName(r.Age),
			pct(r.Milestones), pct(r.Wonders), pct(r.Monuments), pct(r.All()), pct(r.Applied()), share,
			both(r.All()+FestivalBonus), both(r.All()+FestivalBonus+PlentyBonus),
			both(r.Resources["knowledge"]), both(r.Resources["gold"]), other, both(config.ProductionAllHeld[r.Age]))
	}
	all := func(r CapRow) float64 { return r.All() }
	when := func(age string) string {
		if age == "" {
			return "never"
		}
		return "from the " + ageName(age) + " Age"
	}
	fmt.Fprintf(sb, "\nAll production is past its knee %s on what stands alone, %s with a festival running, and %s with a festival and an Age of Plenty. ",
		when(KneePassedIn(rows, 0, all)), when(KneePassedIn(rows, FestivalBonus, all)), when(KneePassedIn(rows, FestivalBonus+PlentyBonus, all)))
	var reached []string
	last := rows[len(rows)-1]
	for _, res := range sortedKeys(last.Resources) {
		if age := KneePassedIn(rows, 0, func(r CapRow) float64 { return r.Resources[res] }); age != "" {
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
		fmt.Fprintf(sb, "No resource's own pool passes its knee in any age, so each applies all it has earned; the fullest at the end are %s.\n", strings.Join(tops, ", "))
	} else {
		fmt.Fprintf(sb, "A resource's own pool passes its knee for: %s.\n", strings.Join(reached, ", "))
	}
	fmt.Fprintf(sb, "\nIn the last age the full set has earned %s and applies %s; under the old clamp it applied %s, and the other %s did nothing.\n", pct(last.All()), pct(last.Applied()), pct(PoolKnee), pct(last.All()-PoolKnee))
	fmt.Fprintf(sb, "\nThe pacing model's all-production pool (config.ProductionAllHeld, the last column) is held as earned and goes through the same soft cap. It must stay at or under the earned column: it is what a well-played game holds, never more than a game can.\n")
}
