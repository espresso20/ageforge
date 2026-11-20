package game

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
)

// advance_view_v2.go is the Advance panel's pillars under the measured gate
// (gate_v2.go): People is one row of population, Industry one row of income,
// Knowledge the keystone and the key tech. Monument is as it always is.

// measureFor is the measure of the gate out of the snapshot's age, when the
// snapshot's ruleset measures its gates and the age has a measure.
func (st GameState) measureFor(set *rules.Set) (config.GateMeasure, bool) {
	if !set.GatesV2() {
		return config.GateMeasure{}, false
	}
	return set.GateMeasure(st.Age)
}

// measuredRows adds the measured gate's rows to rows: the key tech behind the
// keystone's row, the population, and the income of the main materials.
func (st GameState) measuredRows(set *rules.Set, m config.GateMeasure, rows map[string][]GateRow) {
	town := measureTown(m, st.Workers.TotalPop,
		func(res string) float64 { return st.Resources[res].Rate },
		func(tech string) bool { return st.Research.Techs[tech].Researched })

	if m.KeyTech != "" {
		rows[config.PillarKnowledge] = append(rows[config.PillarKnowledge], st.keyTechRow(set, m.KeyTech))
	}

	pop := amountRow("population", rowid.Line("game.advance_view_v2.measuredRows.01", "Population"), float64(town.Population), float64(m.People), false)
	pop.Met = town.PeopleMet()
	rows[config.PillarPeople] = append(rows[config.PillarPeople], pop)

	names := make([]string, 0, len(m.Materials))
	for _, res := range m.Materials {
		names = append(names, set.Name(rules.KindResource, res))
	}
	income := amountRow("income", strings.Join(names, " + ")+rowid.Line("game.advance_view_v2.measuredRows.02", " income"), town.Income, m.Income, true)
	income.Met = town.IndustryMet()
	rows[config.PillarIndustry] = append(rows[config.PillarIndustry], income)
}

// amountRow is a row of an amount the town has against the amount the gate
// measures.
func amountRow(key, label string, have, need float64, perTick bool) GateRow {
	progress := 1.0
	if have < need {
		progress = clampShare(have / need)
	}
	return GateRow{
		Kind: GateRowAmount, Key: key, Label: label,
		Have: intOf(have), Need: intOf(need), Value: have, Target: need, PerTick: perTick,
		Met: have >= need, Progress: progress,
	}
}

// keyTechRow is the age's key tech as a row: met when it is researched.
func (st GameState) keyTechRow(set *rules.Set, tech string) GateRow {
	ts := st.Research.Techs[tech]
	r := GateRow{Kind: GateRowDone, Key: tech, Label: rowid.Line("game.advance_view_v2.keyTechRow.01", "Key tech: ") + set.Name(rules.KindTech, tech), Need: 1}
	switch {
	case ts.Researched:
		r.Have, r.Met, r.Progress, r.Note = 1, true, 1, "researched"
	case st.Research.CurrentTech == tech:
		if st.Research.Price > 0 {
			r.Progress = clampShare(st.Research.Progress / st.Research.Price)
		}
		r.Note = rowid.Sprintf("game.advance_view_v2.keyTechRow.02", "researching, %d%%", int(r.Progress*100))
	default:
		r.Note = rowid.Line("game.advance_view_v2.keyTechRow.03", "not started")
		for _, q := range st.Research.Queue {
			if q == tech {
				r.Note = "queued"
			}
		}
	}
	return r
}
