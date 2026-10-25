package boon

// Rarity weights. Higher means more likely to be drawn. These are the base
// weights; a Profile's WeightMult and RarityScale reshape them at roll time.
const (
	weightCommon   = 10
	weightUncommon = 5
	weightRare     = 2
)

// instantGrowthPerAge is the per-age growth factor for instant lump grants (Kind
// InstantResource). A fixed crate of 50 to 200 is a fortune in the primitive age
// and a rounding error in the quantum age, so the lump is multiplied by
// instantGrowthPerAge^ageIndex (see Profile.instantScale). 2.2 is meant to follow
// the growth of one building's output per age, so a lump stays worth about the
// same number of ticks of production through the whole run.
const instantGrowthPerAge = 2.2

// maxDrainFraction is the ceiling on a ResourceDrain after the profile's
// MagnitudeScale is applied. A setback takes part of a store, never all of it.
const maxDrainFraction = 0.50

// Duration calibration.
//
// Durations matter more than any other number in this table, because the consumer
// (faction encounters) holds timed boons in a fixed number of slots
// (game.MaxConcurrentFactionBoons). Throughput is slots/duration. If that is
// lower than the rate at which encounters arrive, most encounters bounce off a
// full inventory and the reward loop stops rewarding.
//
// Rule of thumb for re-tuning, as offered load in slot-equivalents:
//
//	encounters_per_tick x (share of rolls that hold a slot, ~0.70) x avg_duration  ~=  0.75 x slots
//
// Below ~0.2 x slots the buffs stop overlapping and exploring feels unrewarded.
// Above ~1.3 x slots more than 40% of timed rolls are refused and the loop
// starts to read as rejection. The target is a median combined uplift of x1.2 to
// x1.6 (measure it with go test ./game -run TestBoonTuning -v).
//
// Timed setbacks last about 40% as long as boons. A setback must expire
// noticeably sooner than a boon of the same magnitude, or a run of bad luck
// outlasts every gift that could offset it.

// Catalog returns the boon table. Magnitudes, durations, lump ranges, weights and
// flavor pools are plain data in these literals, so balancing is an edit here.
//
// StorageCap and TradeIncome have no entries (see Kind in boon.go).
func Catalog() []Def {
	return []Def{
		{
			Kind:   RateBuff,
			Name:   "Specialty Windfall",
			MagMin: 0.08, MagMax: 0.20,
			DurMin: 1500, DurMax: 3000,
			Weight: weightCommon,
			Target: TargetSpecialty,
			Flavors: []string{
				"Their master artisans share a trade secret: +{pct} {res} for {ticks}.",
				"A gift of hard-won expertise comes your way: +{pct} {res} production for {ticks}.",
			},
		},
		{
			Kind:   RateBuff,
			Name:   "Resource Surge",
			MagMin: 0.08, MagMax: 0.18,
			DurMin: 1500, DurMax: 2500,
			Weight: weightCommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"A sudden bounty of {res}. Output climbs +{pct} for {ticks}.",
				"Luck runs your way: +{pct} {res} output for {ticks}.",
			},
		},
		{
			Kind:   RateBuff,
			Name:   "Enlightenment",
			MagMin: 0.12, MagMax: 0.25,
			DurMin: 1500, DurMax: 3000,
			Weight:   weightUncommon,
			Target:   TargetSpecificResource,
			Resource: "knowledge",
			Flavors: []string{
				"Foreign scholars open their libraries: +{pct} research speed for {ticks}.",
				"A spark of shared insight takes hold: +{pct} research speed for {ticks}.",
			},
		},
		{
			Kind:   AllProduction,
			Name:   "Industrious Spell",
			MagMin: 0.05, MagMax: 0.12,
			DurMin: 1000, DurMax: 2000,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"A wave of industry sweeps across your lands: +{pct} all production for {ticks}.",
				"Every workshop hums a little louder: +{pct} all production for {ticks}.",
			},
		},
		{
			Kind:   TickSpeed,
			Name:   "Time Dilation",
			MagMin: 0.08, MagMax: 0.15,
			DurMin: 750, DurMax: 1500,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"The days seem to quicken: +{pct} game speed for {ticks}.",
				"Time itself leans your way: +{pct} game speed for {ticks}.",
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
				"Crates of {res} land at your gates, {amt} in all.",
			},
		},
		{
			Kind:      InstantResource,
			Name:      "Grand Cache",
			AmountMin: 500, AmountMax: 1500,
			Weight: weightRare,
			Target: TargetRandomRare,
			Flavors: []string{
				"A lost vault swings open: {amt} {res} spills into your coffers.",
				"An ancient hoard of {res} is yours: {amt} units.",
			},
		},
		{
			Kind:   TempWorkers,
			Name:   "Extra Hands",
			DurMin: 1000, DurMax: 2000,
			AmountMin: 3, AmountMax: 8,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"{n} skilled workers arrive to lend a hand for {ticks}.",
				"A crew of {n} workers joins your cause for {ticks}.",
			},
		},
	}
}

// MalusCatalog returns the setback table, the negative mirror of Catalog. It has
// the same Def shape and the same weighted pick. RollBoon draws from it when a
// Profile's Polarity is Negative.
//
// Magnitudes are modest on purpose: a malus should sting for a while, not end a
// run. The timed entries use the engine's "<res>_rate" and "production_all"
// pools with a negative value. The drain and worker-loss entries use the two
// malus-only Kinds and their own Applier methods.
//
// The flavor strings are player-facing text.
func MalusCatalog() []Def {
	return []Def{
		{
			Kind:      WorkerLoss,
			Polarity:  Negative,
			Name:      "Dysentery",
			AmountMin: 1, AmountMax: 3,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"Dysentery takes hold on the march back. Workers lost on the road home: {n}.",
				"Camp fever moves through the returning column. Workers lost: {n}.",
				"Bad water at the last ford. Workers lost to it: {n}.",
			},
		},
		{
			Kind:      ResourceDrain,
			Polarity:  Negative,
			Name:      "Spoiled Supplies",
			AmountMin: 0.05, AmountMax: 0.15,
			Weight: weightCommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"Damp got into the stores on the journey back: {frac} of your {res} is fit for nothing.",
				"The {res} was packed badly and traveled worse; {frac} of it is written off.",
				"Vermin found the {res} stores before your quartermaster did and took {frac} of it.",
			},
		},
		{
			Kind:     RateBuff,
			Polarity: Negative,
			Name:     "Cursed Relic",
			MagMin:   -0.15, MagMax: -0.08,
			DurMin: 750, DurMax: 1500,
			Weight: weightUncommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"They pressed a relic on you as a parting gift. Your {res} output falls {pct} for {ticks}, and nobody will say why.",
				"The thing your scouts carried home was not meant to leave its shrine: {res} output drops {pct} for {ticks}.",
			},
		},
		{
			Kind:     AllProduction,
			Polarity: Negative,
			Name:     "Bad Omen",
			MagMin:   -0.10, MagMax: -0.05,
			DurMin: 500, DurMax: 1250,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"Word of the expedition's fate spreads faster than the truth of it. All production falls {pct} for {ticks}.",
				"The augurs read the returning party's account and go quiet. All production drops {pct} for {ticks}.",
			},
		},
		{
			Kind:      ResourceDrain,
			Polarity:  Negative,
			Name:      "Lost Scouts",
			AmountMin: 0.02, AmountMax: 0.06,
			MagMin: -0.06, MagMax: -0.03,
			DurMin: 300, DurMax: 600,
			Weight: weightCommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"Part of the party never came back, and {frac} of the {res} they carried went with them. All production falls {pct} for {ticks}.",
				"You are still waiting on names from the last expedition. {frac} of the {res} is unaccounted for, and all production runs {pct} slower for {ticks}.",
			},
		},
	}
}
