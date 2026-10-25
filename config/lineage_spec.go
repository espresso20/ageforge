package config

// lineageSpec is what every row of one lineage file shares. A row writes only
// what is its own (name, price, effects, age, text, WorkerCapacity); the rest
// comes from here through finishLineage.
type lineageSpec struct {
	Key      string  // LineageKey of every row
	Category string  // Category of every row
	Scale    float64 // CostScale of every row: ProducerRate, KnowledgeRate or the housing rate
	Domain   string  // WorkerDomain of every row; "" = no workers
	Lines    bool    // true when the lineage changes resource: LineKey is Key + "_" + OutputResource
}

// finishLineage fills the fields a lineage's rows would otherwise restate, and
// returns rows. It fills a field only while the row leaves it at its zero
// value, so a row that sets one explicitly (marble_quarry's OutputResource and
// LineKey, a row with a WorkerDomain of its own) keeps it. What it fills:
//   - Category, CostScale, LineageKey and WorkerDomain from the spec;
//   - LineageTier from the row's position (0 is unset, so a row cannot ask for
//     tier 0 in any other place);
//   - OutputResource from the Target of the row's first "production" effect
//     ("" when it has none: housing, research and military rows);
//   - LineKey, only when spec.Lines is set and OutputResource is not empty.
//
// A building that belongs to no lineage (nano_foundry has no LineageKey) is
// appended after the call, so it is never filled.
func finishLineage(rows []BuildingDef, s lineageSpec) []BuildingDef {
	for i := range rows {
		r := &rows[i]
		if r.Category == "" {
			r.Category = s.Category
		}
		if r.CostScale == 0 {
			r.CostScale = s.Scale
		}
		if r.LineageKey == "" {
			r.LineageKey = s.Key
		}
		if r.LineageTier == 0 {
			r.LineageTier = i
		}
		if r.WorkerDomain == "" {
			r.WorkerDomain = s.Domain
		}
		if r.OutputResource == "" {
			r.OutputResource = firstProduction(r.Effects)
		}
		if s.Lines && r.LineKey == "" && r.OutputResource != "" {
			r.LineKey = s.Key + "_" + r.OutputResource
		}
	}
	return rows
}

// firstProduction is the Target of the first "production" effect, or "".
func firstProduction(effects []Effect) string {
	for _, e := range effects {
		if e.Type == "production" {
			return e.Target
		}
	}
	return ""
}
