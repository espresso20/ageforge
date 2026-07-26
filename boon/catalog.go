package boon

// Rarity weights. Higher = more likely to be drawn. These are the neutral base
// weights; a Profile's WeightMult and RarityScale reshape them at roll time.
const (
	weightCommon   = 10
	weightUncommon = 5
	weightRare     = 2
)

// Catalog returns the starting boon table. Everything here is plain, tunable
// DATA — magnitudes, durations, lump ranges, weights, and flavor pools live in
// these literals, so balancing is an edit here and nowhere else.
//
// StorageCap and TradeIncome are deliberately absent: they need engine plumbing
// that does not exist yet (a storage-cap pool, a trade-income modifier). The
// Kinds are reserved in boon.go; add their Defs — and the matching Applier
// methods — once that machinery lands.
func Catalog() []Def {
	return []Def{
		{
			Kind:   RateBuff,
			Name:   "Specialty Windfall",
			MagMin: 0.08, MagMax: 0.20,
			DurMin: 3000, DurMax: 6000,
			Weight: weightCommon,
			Target: TargetSpecialty,
			Flavors: []string{
				"Their master artisans share a trade secret — +{pct} {res} for {ticks} ticks.",
				"A gift of hard-won expertise flows your way — +{pct} {res} production.",
			},
		},
		{
			Kind:   RateBuff,
			Name:   "Resource Surge",
			MagMin: 0.08, MagMax: 0.18,
			DurMin: 3000, DurMax: 5000,
			Weight: weightCommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"A sudden bounty of {res} — output climbs +{pct} for {ticks} ticks.",
				"Fortune favours your {res} stores — +{pct} while it lasts.",
			},
		},
		{
			Kind:   RateBuff,
			Name:   "Enlightenment",
			MagMin: 0.12, MagMax: 0.25,
			DurMin: 3000, DurMax: 6000,
			Weight:   weightUncommon,
			Target:   TargetSpecificResource,
			Resource: "knowledge",
			Flavors: []string{
				"Foreign scholars open their libraries — +{pct} knowledge for {ticks} ticks.",
				"A spark of shared insight takes hold — +{pct} knowledge.",
			},
		},
		{
			Kind:   AllProduction,
			Name:   "Industrious Spell",
			MagMin: 0.05, MagMax: 0.12,
			DurMin: 2000, DurMax: 4000,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"A wave of industry sweeps the realm — +{pct} to all production for {ticks} ticks.",
				"Every workshop hums a little louder — +{pct} production across the board.",
			},
		},
		{
			Kind:   TickSpeed,
			Name:   "Time Dilation",
			MagMin: 0.08, MagMax: 0.15,
			DurMin: 1500, DurMax: 3000,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"The days seem to quicken — +{pct} tick speed for {ticks} ticks.",
				"Time itself leans in your favour — +{pct} tempo for a while.",
			},
		},
		{
			Kind:      InstantResource,
			Name:      "Supply Drop",
			AmountMin: 50, AmountMax: 200,
			Weight: weightCommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"A caravan arrives bearing {amt} {res}.",
				"Crates of {res} land at your gates — {amt} in all.",
			},
		},
		{
			Kind:      InstantResource,
			Name:      "Grand Cache",
			AmountMin: 500, AmountMax: 1500,
			Weight: weightRare,
			Target: TargetRandomRare,
			Flavors: []string{
				"A lost vault swings open — {amt} {res} spills into your coffers.",
				"An ancient hoard of {res} is yours: {amt} units.",
			},
		},
		{
			Kind:   TempWorkers,
			Name:   "Extra Hands",
			DurMin: 2000, DurMax: 4000,
			AmountMin: 3, AmountMax: 8,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"{n} skilled hands arrive to lend their labour for {ticks} ticks.",
				"A work-gang of {n} joins your cause for a season.",
			},
		},
	}
}
