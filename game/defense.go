package game

import (
	"fmt"
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// Army defense (Trello qPhJ5YxH, Phase 3): the Defense Rating finally does
// something. It blunts raid events, war raids and the losses of an Endure; the
// curve and its constants live in config/defense.go.
//
// Everything here is read-only over engine state except the record* helpers,
// which write GameStats.Defense and are only called from doTick paths and
// Endure, under the write lock. The read-only helpers are safe from GetState.

// DefenseTally is what the army has saved this run (it resets with the run,
// like the rest of GameStats). Saved as GameStats.Defense, omitted while the
// army has saved nothing, so a save from a player who never trained a soldier
// is byte-identical to one from before the army mattered.
type DefenseTally struct {
	// Resources kept that a raid event, a war raid or an Endure would have
	// taken.
	Resources map[string]float64 `json:"resources,omitempty"`
	// Workers who stayed when a raid event would have driven them off.
	Workers int `json:"workers,omitempty"`
	// Buildings still standing after an Endure that would have destroyed them.
	Buildings int `json:"buildings,omitempty"`
	// Raids is the number of raids (events and war raids) the garrison
	// blunted.
	Raids int `json:"raids,omitempty"`
}

// clone returns a deep copy, or nil.
func (t *DefenseTally) clone() *DefenseTally {
	if t == nil {
		return nil
	}
	c := *t
	if t.Resources != nil {
		c.Resources = make(map[string]float64, len(t.Resources))
		for k, v := range t.Resources {
			c.Resources[k] = v
		}
	}
	return &c
}

// defenseTally returns the run's tally, creating it on first use.
func (ge *GameEngine) defenseTally() *DefenseTally {
	if ge.Stats.Defense == nil {
		ge.Stats.Defense = &DefenseTally{}
	}
	return ge.Stats.Defense
}

// recordSavedResource adds amount of res to the tally. Non-positive amounts
// are ignored.
func (ge *GameEngine) recordSavedResource(res string, amount float64) {
	if amount <= 0 {
		return
	}
	t := ge.defenseTally()
	if t.Resources == nil {
		t.Resources = make(map[string]float64)
	}
	t.Resources[res] += amount
}

// militaryPower is the summed military_power bonus: research, milestones,
// prestige upgrades and wonders. Read-only.
func (ge *GameEngine) militaryPower() float64 {
	prestige := ge.Prestige.GetBonuses()
	wonders := ge.getWonderBonuses()
	return ge.Research.Bonus(config.EffectMilitaryPower, "") + ge.permanentBonuses["military_power"] + prestige["military_power"] + wonders["military_power"]
}

// expeditionReward is the summed expedition_reward bonus, from the same
// sources as militaryPower. Read-only.
func (ge *GameEngine) expeditionReward() float64 {
	prestige := ge.Prestige.GetBonuses()
	wonders := ge.getWonderBonuses()
	return ge.Research.Bonus(config.EffectExpeditionReward, "") + ge.permanentBonuses["expedition_reward"] + prestige["expedition_reward"] + wonders["expedition_reward"]
}

// defenseRating is the live Defense Rating (MilitaryState.DefenseRating).
func (ge *GameEngine) defenseRating() float64 {
	return ge.Military.CalculateDefense(int(ge.Resources.Get("soldiers")), ge.militaryPower())
}

// ageThreat is the raid threat of age (config.AgeThreat by its order).
func (ge *GameEngine) ageThreat(age string) float64 {
	return config.AgeThreat(ge.progress.GetAgeOrder()[age])
}

// raidMitigation is the share of a raid's losses the garrison blunts right
// now, against the current age's threat.
func (ge *GameEngine) raidMitigation() float64 {
	return config.DefenseMitigation(ge.defenseRating(), ge.ageThreat(ge.age))
}

// EndureOutcome is what an Endure costs once the Harbinger's Brace and the
// garrison have both had their say. Endure applies it and the catastrophe
// modal and harbinger panel preview it, from the same function, so a preview
// can never promise what Endure will not deliver.
type EndureOutcome struct {
	BraceLevel int
	// Garrison is the share the garrison would blunt of a raid of the age the
	// catastrophe strikes in (before the combined cap).
	Garrison float64
	// DestroyPct is the share of destroyable buildings (neither wonders nor
	// storage) destroyed after Brace and the garrison, in percent (e.g. 16.5).
	DestroyPct float64
	// KeepFrac is the share of every stock kept (0.15 unbraced, no garrison).
	KeepFrac float64
	// BracedDestroyPct and BracedKeepFrac are the same numbers with Brace
	// only, no garrison: what the player would face without an army.
	BracedDestroyPct float64
	BracedKeepFrac   float64
	// Capped reports that the combined cap (config.EndureReductionCap) cut the
	// garrison's share of the building or stock loss.
	Capped bool
	// Destroyable, DestroyCount and BuildingsSaved are counts for the current
	// empire: buildings that could fall, buildings that will, and buildings
	// the garrison keeps standing.
	Destroyable    int
	DestroyCount   int
	BuildingsSaved int
}

// computeEndure computes an Endure's losses for destroyable buildings at the
// given Brace level with a garrison that blunts garrison (0..cap) of a raid.
//
// Brace first: braceDestroyPct / braceKeepFrac. Then the garrison blunts that
// share of what is left, but never so much that Brace and garrison together
// cut the unbraced loss by more than config.EndureReductionCap. With no
// garrison the numbers are exactly the Brace-only numbers, bit for bit.
func computeEndure(destroyable, brace int, garrison float64) EndureOutcome {
	if brace < 0 || brace > HarbingerMaxBrace {
		brace = 0
	}
	if garrison < 0 {
		garrison = 0
	}
	destroyPct, keep := braceDestroyPct[brace], braceKeepFrac[brace]
	out := EndureOutcome{
		BraceLevel:       brace,
		Garrison:         garrison,
		BracedDestroyPct: float64(destroyPct),
		BracedKeepFrac:   keep,
		DestroyPct:       float64(destroyPct),
		KeepFrac:         keep,
		Destroyable:      destroyable,
	}

	// The building loss, as Endure has always counted it: a floor of the
	// braced percentage, at least one if anything can fall.
	count := destroyable * destroyPct / 100
	if count < 1 && destroyable > 0 {
		count = 1
	}
	out.DestroyCount = count
	if garrison <= 0 {
		return out
	}

	// Buildings. floor keeps the garrison's share at or under its percentage.
	// Worked in percent, where the Brace shares are whole numbers, so the
	// cap lands on round values (level 2: at most 10% -> 8%, a fifth).
	bSave := garrison
	floorPct := float64(float64(braceDestroyPct[0]) * (1 - config.EndureReductionCap))
	if braced := float64(destroyPct); braced > 0 {
		if most := (braced - floorPct) / braced; bSave > most {
			bSave = math.Max(most, 0)
			out.Capped = true
		}
	}
	out.DestroyPct = float64(float64(destroyPct) * (1 - bSave))
	// The epsilon keeps a product that is a whole number on paper (10 x 0.2)
	// from flooring one short.
	saved := int(math.Floor(float64(float64(count)*bSave) + 1e-9))
	if count-saved < 1 && destroyable > 0 {
		saved = count - 1
	}
	if saved < 0 {
		saved = 0
	}
	out.BuildingsSaved = saved
	out.DestroyCount = count - saved

	// Stock.
	sSave := garrison
	unbracedLoss := 1 - braceKeepFrac[0]
	floorLoss := float64(unbracedLoss * (1 - config.EndureReductionCap))
	if loss := 1 - keep; loss > 0 {
		if most := 1 - floorLoss/loss; sSave > most {
			sSave = math.Max(most, 0)
			out.Capped = true
		}
		out.KeepFrac = keep + float64(loss*sSave)
	}
	return out
}

// DefaultEndureOutcome is an Endure with no Brace and no garrison: the losses
// the catastrophe modal shows when no pending catastrophe carries a preview.
func DefaultEndureOutcome() EndureOutcome { return computeEndure(0, 0, 0) }

// endurePreview is the outcome an Endure would have right now at the given
// Brace level, with the garrison measured against age's threat. Read-only.
func (ge *GameEngine) endurePreview(brace int, age string) EndureOutcome {
	g := config.DefenseMitigation(ge.defenseRating(), ge.ageThreat(age))
	return computeEndure(ge.Buildings.DestroyableCount(), brace, g)
}

// garrisonSavedLine is the log line that tells the player what the garrison
// kept from a raid event: m is the share it blunted, resources and workers
// what it saved. Empty when it saved nothing.
func garrisonSavedLine(m float64, resources map[string]float64, workers int) string {
	parts := lossParts(resources, workers)
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("  [green]Your garrison blunted about %.0f%% of the raid: you kept %s.[-]", m*100, joinAnd(parts))
}

// endureGarrisonLine is the Endure log line for what the garrison saved, or ""
// when there was no garrison to speak of.
func endureGarrisonLine(o EndureOutcome) string {
	if o.Garrison <= 0 || (o.BuildingsSaved == 0 && o.KeepFrac == o.BracedKeepFrac) {
		return ""
	}
	stood := ""
	switch o.BuildingsSaved {
	case 0:
	case 1:
		stood = "1 building still stands that would have fallen, and "
	default:
		stood = fmt.Sprintf("%d buildings still stand that would have fallen, and ", o.BuildingsSaved)
	}
	line := fmt.Sprintf("  [green]Your garrison held the line: %syou keep %.0f%% of your stock instead of %.0f%%.[-]",
		stood, o.KeepFrac*100, o.BracedKeepFrac*100)
	if o.Capped {
		line += fmt.Sprintf(" [gray](Brace and garrison together soften an Endure by at most %.0f%%.)[-]", config.EndureReductionCap*100)
	}
	return line
}

// applyWarRaids drains the war raids DiplomacyManager queued this tick, applies
// the resource losses (blunted by the garrison) and announces them. Under the
// write lock.
func (ge *GameEngine) applyWarRaids() {
	// Apply war raids (resource losses) and announce them. The announcement lives
	// here rather than in DiplomacyManager.Tick because the flavour half is drawn
	// off ge.rng, which the manager has no access to.
	//
	// Sorted by faction key first: pendingRaids is built by walking a MAP, so its
	// order is randomised, and drawing prose in that order would make which raid
	// got which sentence unreproducible from the seed.
	raids := ge.Diplomacy.TakePendingRaids()
	sort.Slice(raids, func(i, j int) bool { return raids[i].FactionKey < raids[j].FactionKey })
	guard := 0.0
	if len(raids) > 0 {
		guard = ge.raidMitigation()
	}
	for _, raid := range raids {
		// The garrison blunts part of every war raid (config/defense.go). A
		// raid bigger than the stock takes nothing (Remove refuses it), so
		// only a raid that would have landed in full is blunted: an army
		// never turns a raid that missed into one that lands.
		def, known := ge.rules.Faction(raid.FactionKey)
		if !(ge.Resources.Get(raid.Resource) >= raid.Amount) {
			// Remove would refuse: the raid takes nothing. Say so rather
			// than report a loss that never happened.
			if known {
				raid.Message = raidMissedMessage(def, raid.Resource)
			}
		} else {
			// The techs' cut of raid losses comes first: the raid that
			// lands is that much smaller, and the garrison meets what is
			// left. (A raid that missed stays missed: the cut never turns
			// one too big to land into one that does.)
			if f := ge.raidLossFactor(); f != 1 {
				raid.Amount = float64(raid.Amount * f)
				if known {
					raid.Message = raidMessage(def, raid.Amount, raid.Resource)
				}
			}
			if guard > 0 {
				kept := float64(raid.Amount * guard)
				raid.Amount -= kept
				if known {
					raid.Message = raidMessageDefended(def, raid.Amount, kept, raid.Resource, guard)
				}
				ge.recordSavedResource(raid.Resource, kept)
				ge.defenseTally().Raids++
			}
			ge.Resources.Remove(raid.Resource, raid.Amount)
		}
		ge.addLog("event", ge.raidLogLine(raid))
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "war_raid",
			Name:      "Under Raid",
			TicksLeft: lendEventDisplayTicks,
		})
	}
}
