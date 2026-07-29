package boon

// Rarity weights. Higher = more likely to be drawn. These are the neutral base
// weights; a Profile's WeightMult and RarityScale reshape them at roll time.
const (
	weightCommon   = 10
	weightUncommon = 5
	weightRare     = 2
)

// instantGrowthPerAge is the per-age growth factor applied to INSTANT lump
// grants (Kind InstantResource). A fixed 50–200 crate of goods is a fortune in
// the primitive age and a rounding error in the quantum age, so the lump is
// multiplied by instantGrowthPerAge^ageIndex (see Profile.instantScale).
//
// Calibration: per-BUILDING production in this game roughly doubles per age
// (engineering lineage: 0.10 iron/tick at bronze → 26,214 quantum_flux/tick at
// quantum, ~2.0x per age over 18 ages). 2.2 tracks that curve with a little
// headroom for growing building counts, so a lump stays worth a broadly
// constant number of TICKS of production across the whole run. Storage caps
// grow far faster (~4.5x/age), which is why the lump never threatens to fill
// a late-game stockpile — it is a gift, not a shortcut.
const instantGrowthPerAge = 2.2

// maxDrainFraction is the hard ceiling on a ResourceDrain malus after the
// profile's MagnitudeScale is folded in. A setback empties part of a store,
// never the whole thing by design.
const maxDrainFraction = 0.50

// --- Duration calibration ----------------------------------------------------
//
// Durations are the load-bearing number in this table, because the consumer that
// draws from it (faction encounters) holds timed boons in a fixed number of
// SLOTS. Landed-boon throughput is therefore capped at slots/duration, and if
// that is smaller than the rate encounters arrive, almost every encounter bounces
// off a full inventory and the reward loop stops rewarding.
//
// A measured pass (game/boon_tuning_test.go) caught exactly that: at the original
// 1500–6000 tick durations a continuously-exploring player landed a boon on 2–3%
// of encounters and sat at full capacity 97% of the time. Durations were cut ~5x
// (alongside a ~4x cut to encounter odds and a 3 -> 5 slot bump) to bring offered
// load back in line with capacity; the same harness now measures a ~60% land rate
// at ~33% saturation.
//
// The rule of thumb if these are ever re-tuned: keep
//
//	encounters_per_tick x (timed share ~0.75) x avg_duration  ~=  slots + 1
//
// Setback durations are kept at the same ~1/5 scale for a different reason: a
// timed setback must expire NOTICEABLY sooner than a boon of the same magnitude,
// or a run of bad luck outlasts every gift that could offset it.

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
			DurMin: 600, DurMax: 1200,
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
			DurMin: 600, DurMax: 1000,
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
			DurMin: 600, DurMax: 1200,
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
			DurMin: 400, DurMax: 800,
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
			DurMin: 300, DurMax: 600,
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
			DurMin: 400, DurMax: 800,
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

// MalusCatalog returns the SETBACK table — the negative mirror of Catalog().
// Same engine, same Def shape, same weighted pick; only the sign changes. It is
// drawn instead of Catalog() when a Profile's Polarity is Negative.
//
// Magnitudes here are deliberately modest: a malus should sting for a while, not
// end a run. The timed entries lean on the engine's existing "<res>_rate" /
// "production_all" pools with a NEGATIVE value; the drain and worker-loss
// entries use the two malus-only Kinds and their dedicated Applier methods.
//
// Flavor is dry and in-world, matching the boon table. These strings are
// player-facing production text — keep them that way.
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
				"The expedition succumbs to dysentery on the road home — {n} do not return.",
				"Camp fever moves through the returning column; {n} are buried where they fell.",
				"Bad water at the last ford. {n} of your people are lost to it.",
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
				"Damp got into the stores on the journey back — {frac} of your {res} is fit for nothing.",
				"The {res} was packed badly and travelled worse; {frac} of it is written off.",
				"Vermin found the {res} stores before your quartermaster did — {frac} gone.",
			},
		},
		{
			Kind:     RateBuff,
			Polarity: Negative,
			Name:     "Cursed Relic",
			MagMin:   -0.15, MagMax: -0.08,
			DurMin: 300, DurMax: 600,
			Weight: weightUncommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"They pressed a relic on you as a parting gift. {res} output falls {pct} for {ticks} ticks, and nobody will say why.",
				"The thing your scouts carried home was not meant to leave its shrine — {res} output drops {pct} for {ticks} ticks.",
			},
		},
		{
			Kind:     AllProduction,
			Polarity: Negative,
			Name:     "Bad Omen",
			MagMin:   -0.10, MagMax: -0.05,
			DurMin: 200, DurMax: 500,
			Weight: weightUncommon,
			Target: TargetNone,
			Flavors: []string{
				"Word of the expedition's fate spreads faster than the truth of it — all production falls {pct} for {ticks} ticks.",
				"The augurs read the returning party's account and go quiet. Production drops {pct} across the realm for {ticks} ticks.",
			},
		},
		{
			Kind:      ResourceDrain,
			Polarity:  Negative,
			Name:      "Lost Scouts",
			AmountMin: 0.02, AmountMax: 0.06,
			MagMin: -0.06, MagMax: -0.03,
			DurMin: 120, DurMax: 240,
			Weight: weightCommon,
			Target: TargetRandomAge,
			Flavors: []string{
				"Half the scouting party never came back, and {frac} of the {res} they carried went with them. The rest work {pct} slower for {ticks} ticks.",
				"You are still waiting on names from the last expedition. {frac} of the {res} is unaccounted for and the realm works {pct} slower for {ticks} ticks.",
			},
		},
	}
}
