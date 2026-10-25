package smoke

import "github.com/espresso20/ageforge/game"

// PriceRow is the static price of one thread's harbinger answers, in the age
// its harbinger arrives in.
type PriceRow struct {
	Epoch       string `json:"epoch"`
	TargetEpoch string `json:"target_epoch"`
	// Age is the age the harbinger arrives in. Appease is priced on it (what
	// that age makes in the thread's shortest warning); the price then holds
	// for the whole thread.
	Age string `json:"age"`
	// WarningTicks is that shortest warning in ticks at 1x: a fifth of the
	// age's pacing target for a doom, two thirds of it for the Last Passage.
	WarningTicks float64 `json:"appease_warning_ticks"`
	// AppeaseL1 and AppeaseL2 are each level's own price: level 2 costs
	// the same again as level 1.
	AppeaseL1 map[string]float64 `json:"appease_l1_cost"`
	AppeaseL2 map[string]float64 `json:"appease_l2_cost"`
	// BraceL1 and BraceL2 likewise. Brace is priced on the same warning as
	// Appease, in the materials the arrival age makes; FinalEra marks the
	// final era's two threads (its fated doom's and the Last Passage's),
	// whose Brace costs more of the warning.
	BraceL1  map[string]float64 `json:"brace_l1_cost"`
	BraceL2  map[string]float64 `json:"brace_l2_cost"`
	FinalEra bool               `json:"final_era,omitempty"`
	// BraceCeiling is the most a thread's Brace level 1 may ask of each
	// material (three fifths of the arrival age's stock scale, rounded down
	// to two figures), BraceCeilingHours the same ceiling in hours of a
	// moderate economy's income of the material, and StockScale the age's
	// scale (config.StockScale).
	BraceCeiling      map[string]float64 `json:"brace_ceiling,omitempty"`
	BraceCeilingHours map[string]float64 `json:"brace_ceiling_hours,omitempty"`
	StockScale        float64            `json:"stock_scale,omitempty"`
}

// HarbingerPrices is the prices of every thread whose doom can be
// answered (game.HarbingerPriceTable), for each age its harbinger can arrive
// in: a fated doom's, and the Last Passage's (a TargetEpoch of ""), next to
// the most storage reachable in that age. The Stone Era is skipped: only
// false prophets come there, and nothing can strike.
func HarbingerPrices() []PriceRow {
	var rows []PriceRow
	for _, p := range game.HarbingerPriceTable() {
		row := PriceRow{
			Epoch: p.Epoch, TargetEpoch: p.Epoch, Age: p.Age, WarningTicks: p.WarningTicks,
			AppeaseL1: p.AppeaseL1, AppeaseL2: p.AppeaseL2, BraceL1: p.BraceL1, BraceL2: p.BraceL2,
			FinalEra: p.FinalEra, BraceCeiling: p.BraceCeiling, BraceCeilingHours: p.BraceCeilingHours, StockScale: p.StockScale,
		}
		if p.LastPassage {
			row.TargetEpoch = ""
		}
		rows = append(rows, row)
	}
	return rows
}
