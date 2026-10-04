package game

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
	"github.com/espresso20/ageforge/rules"
)

// The Last Passage: the Cosmic Era's passage.
//
// Every other epoch's passage is its transition into the next epoch. The Cosmic
// Era has no next epoch, so its passage is prestige itself. A harbinger thread
// warns of it from the era's first age (harbinger.go; its TargetEpoch is ""),
// alongside the era's own fated doom, the Reality Tear (fate.go), which settles
// first: before the roll at a prestige, and before the choice when both are
// pending. Confirming prestige rolls the Last Passage:
//
//   - The odds are the epoch roll's: (1 - good chance by faith band) × 0.30 ×
//     the Appease multiplier, or certain when invited. One ge.rng draw, always
//     taken, so the stream's shape does not depend on the odds or the invite.
//   - No catastrophe: the verdict (Spared), the ending line, and prestige
//     completes as usual.
//   - Catastrophe: prestige does NOT complete. pendingLastPassage is set, the
//     catastrophe modal offers the choice, and only prestige waits (AdvanceAge
//     and everything else carry on). It is persisted.
//     Endure: prestige completes, keeping LastPassageKeep of the run's points
//     (0.70 / 0.85 at Brace 1 / 2), floored.
//     Succumb: prestige completes with no points from the run, and the Cosmic
//     Legacy flag is set: all production × 1.1 after the ×3 caps, forever
//     (it survives prestige and Succumb; only a full wipe clears it). A player
//     who already carries it can only Endure.
//
// Every prestige, from any age, ends with one RunEnding flavor line in the
// voice of the age the run ended in (in the Cosmic Era, the age's harbinger).
//
// Prestige from before the Cosmic Era never rolls. Everything unexported here
// expects the write lock, except the read-only helpers GetState uses
// (lastPassageState, lastPassageBraceLevel, cosmicLegacyModifiers).

const (
	// PassageEpoch and PassagePrestige are CatastropheOutlook.Passage values:
	// the next passage is an epoch transition, or (in the final epoch) prestige.
	PassageEpoch    = "epoch"
	PassagePrestige = "prestige"

	// LastPassageKeep is the share of the run's prestige points Endure keeps,
	// unbraced. See lastPassageKeepFrac for the braced shares.
	LastPassageKeep = 0.50

	// CosmicLegacyProductionBonus is the Cosmic Legacy's permanent production
	// bonus: everything a resource makes × (1 + this), after the ×3 caps
	// (cosmicLegacyFactor, recalculateRates).
	CosmicLegacyProductionBonus = 0.10

	// cosmicLegacySource is the resolver Source of the Cosmic Legacy bonus.
	cosmicLegacySource = "cosmic_legacy"
)

// lastPassageKeepFrac is Endure's share of the run's points by Brace level
// (index 0 = unbraced).
var lastPassageKeepFrac = [HarbingerMaxBrace + 1]float64{LastPassageKeep, 0.70, 0.85}

// LastPassageKeepFor returns the share of prestige points an Endure keeps at
// Brace level brace (clamped to 0..HarbingerMaxBrace).
func LastPassageKeepFor(brace int) float64 {
	if brace < 0 || brace > HarbingerMaxBrace {
		brace = 0
	}
	return lastPassageKeepFrac[brace]
}

// prestigeEnding is how a run ends, for completePrestige.
type prestigeEnding int

const (
	prestigePlain        prestigeEnding = iota // before the Cosmic Era: no Last Passage
	lastPassageSpared                          // rolled, and it did not come
	lastPassageEndured                         // it came; the player chose Endure
	lastPassageSuccumbed                       // it came; the player chose Succumb
)

// LastPassageState is the UI's picture of the Last Passage (GameState.LastPassage).
type LastPassageState struct {
	// Pending is true while the Last Passage waits for Endure or Succumb.
	Pending bool
	// CosmicLegacy is true once the player carries the Cosmic Legacy; Succumb is
	// then refused.
	CosmicLegacy bool
	// BraceLevel is the Cosmic Era thread's Brace level (0 without a thread).
	BraceLevel int
	// KeepPct is the share of points Endure keeps at BraceLevel, in percent.
	KeepPct int
	// PointsNow is what prestige would earn right now in full;
	// PointsIfEndured is what an Endure would keep of it.
	PointsNow       int
	PointsIfEndured int
	// Invited is true when the Cosmic Era thread invited the Last Passage.
	Invited bool
}

// lastPassageApplies reports whether prestige from here rolls the Last
// Passage: the current epoch is the final one and past the Iron gate. Read-only.
func (ge *GameEngine) lastPassageApplies() bool {
	return ge.rules.IsFinalEra(ge.currentEpoch) && ge.rules.CatastropheAllowed(ge.currentEpoch)
}

// lastPassageBlockErr is the error DoPrestige returns while the Last Passage
// is pending.
func lastPassageBlockErr(set *rules.Set) error {
	name, _ := set.LastPassage()
	return fmt.Errorf("%s is upon you. Type 'catastrophe' to choose Endure or Succumb before you prestige.", name)
}

// rollLastPassage rolls the Last Passage at a confirmed prestige and reports
// whether it came. On a hit it makes the choice pending and tells the UI.
func (ge *GameEngine) rollLastPassage() bool {
	out := ge.catastropheOutlook()
	// Always drawn, so the stream's shape does not depend on the odds or on an
	// invite.
	roll := ge.gameRNG().Float64()
	invited := ge.catastropheInvited
	if !invited && roll >= out.Probability {
		return false
	}
	ge.catastropheInvited = false
	source := catastropheRolled
	if invited {
		source = catastropheInvited
	}
	ge.triggerLastPassage(source)
	return true
}

// triggerLastPassage makes the Last Passage pending: log lines and a bus event
// so the dashboard toast fires. Bus handlers must not take the engine lock.
func (ge *GameEngine) triggerLastPassage(source string) {
	ge.pendingLastPassage = true
	name, _ := ge.rules.LastPassage()
	ge.addLog("warning", fmt.Sprintf("☄ %s has come. The civilization cannot pass out of this age unmarked.", name))
	ge.addLog("warning", "  Type 'catastrophe' to choose Endure or Succumb. Only prestige waits; the run goes on until you choose.")
	ge.Bus.Publish(EventData{
		Type: EventEpochEventFired,
		Payload: map[string]interface{}{
			"event_key":  config.LastPassageKey,
			"event_name": name,
			"event_type": "catastrophe",
			"epoch_key":  ge.currentEpoch,
			"source":     source,
		},
	})
}

// forceLastPassage makes the Last Passage pending without a roll: the dev
// console's testing tool. Same gates as a real roll.
func (ge *GameEngine) forceLastPassage() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	// /age jumps the age without the epoch; line them up first.
	if ep := ge.rules.EraOf(ge.age); ep != ge.currentEpoch {
		ge.currentEpoch = ep
	}
	switch {
	case ge.pendingLastPassage:
		return fmt.Errorf("The Last Passage is already pending.")
	case ge.pendingCatastrophe != "":
		return fmt.Errorf("A catastrophe is already pending.")
	case !ge.lastPassageApplies():
		return fmt.Errorf("The Last Passage only comes in the final epoch.")
	}
	ge.triggerLastPassage(catastropheForced)
	return nil
}

// ForceLastPassageForTest is a test hook for other packages (the UI's modal
// and theme-sweep tests): it places the engine in age (a Cosmic Era age), with
// that age's epoch, and makes the Last Passage pending. Not reachable from play.
func (ge *GameEngine) ForceLastPassageForTest(age string) error {
	ge.mu.Lock()
	if _, ok := ge.rules.Age(age); !ok {
		ge.mu.Unlock()
		return fmt.Errorf("Unknown age '%s'.", age)
	}
	ge.age = age
	ge.currentEpoch = ge.rules.EraOf(age)
	ge.mu.Unlock()
	return ge.forceLastPassage()
}

// EndureLastPassage answers a pending Last Passage with Endure: prestige
// completes, keeping part of the run's points. Takes the write lock.
func (ge *GameEngine) EndureLastPassage() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.resolveLastPassage(lastPassageEndured)
}

// SuccumbLastPassage answers a pending Last Passage with Succumb: prestige
// completes with no points from the run and grants the Cosmic Legacy. Refused
// when the player already carries it. Takes the write lock.
func (ge *GameEngine) SuccumbLastPassage() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	return ge.resolveLastPassage(lastPassageSuccumbed)
}

// resolveLastPassage completes the pending prestige the way how says. A
// catastrophe pending beside it (the Cosmic Era's fated doom struck while
// prestige waited) is answered first.
func (ge *GameEngine) resolveLastPassage(how prestigeEnding) error {
	if !ge.pendingLastPassage {
		return fmt.Errorf("The Last Passage has not come.")
	}
	if ge.pendingCatastrophe != "" {
		name, _ := ge.rules.Catastrophe(ge.pendingCatastrophe)
		return fmt.Errorf("%s came first. Answer it before the Last Passage.", name)
	}
	if how == lastPassageSuccumbed && ge.cosmicLegacy {
		return fmt.Errorf("You already carry the Cosmic Legacy, so the Last Passage can only be endured.")
	}
	ge.pendingLastPassage = false
	ge.completePrestige(how)
	return nil
}

// lastPassageBraceLevel is the Brace level bought in the Cosmic Era's Last
// Passage thread (live, or parked behind the era's fated doom), 0 without
// one. Read-only.
func (ge *GameEngine) lastPassageBraceLevel() int {
	h := ge.lastPassageThread()
	if h == nil || h.BraceLevel < 0 || h.BraceLevel > HarbingerMaxBrace {
		return 0
	}
	return h.BraceLevel
}

// lastPassagePoints applies how to the run's full points: all of them, the
// Endure share (floored, never below 0), or none.
func (ge *GameEngine) lastPassagePoints(how prestigeEnding, full int) int {
	switch how {
	case lastPassageEndured:
		p := int(math.Floor(float64(full) * LastPassageKeepFor(ge.lastPassageBraceLevel())))
		if p < 0 {
			p = 0
		}
		return p
	case lastPassageSuccumbed:
		return 0
	}
	return full
}

// runEndingLines writes the run's last lines (the verdict, the Last Passage
// outcome, the ending) aside and returns them, so the prestige reset, which
// clears the log, can carry them into the new run. Draws from ge.rng in a
// fixed order: the verdict line, then the ending line.
func (ge *GameEngine) runEndingLines(how prestigeEnding, points, full int) []LogEntry {
	saved := ge.log
	ge.log = nil

	name, flavorText := ge.rules.LastPassage()
	switch how {
	case lastPassageSpared:
		ge.resolveHarbinger("", false)
	case lastPassageEndured:
		keep := LastPassageKeepFor(ge.lastPassageBraceLevel())
		ge.resolveHarbinger("", true)
		ge.addLog("warning", fmt.Sprintf("☄ Endure: %s. %s", name, flavorText))
		ge.addLog("warning", fmt.Sprintf("  You keep %.0f%% of this run's prestige points: %d of %d.", keep*100, points, full))
	case lastPassageSuccumbed:
		ge.resolveHarbinger("", true)
		ge.addLog("event", fmt.Sprintf("☄ Succumb: %s took everything this run had. No prestige points from it.", name))
	}
	ge.logRunEnding()

	lines := ge.log
	ge.log = saved
	for i := range lines {
		lines[i].Tick = 0
	}
	return lines
}

// logRunEnding writes the run's closing flavor line in the voice of the age it
// ended in; in the Cosmic Era the age's harbinger is the Subject.
func (ge *GameEngine) logRunEnding() {
	req := flavor.Request{Moment: flavor.RunEnding, Age: ge.age}
	if ge.rules.IsFinalEra(ge.rules.EraOf(ge.age)) {
		if def, ok := ge.rules.Harbinger(ge.age); ok {
			req.Subject = def.Name
		}
	}
	if l := ge.flavorStream().Line(req, ge.gameRNG()); l != "" {
		ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", l))
	}
}

// recordLastPassageOutcome writes the civilization-log entry for an Endure or
// Succumb (it survives prestige, like every catastrophe entry). The entries
// carry the markers countCatastropheOutcomes matches, so the Last Passage
// counts as a catastrophe endured or succumbed to.
func (ge *GameEngine) recordLastPassageOutcome(how prestigeEnding, points, full int) {
	name, _ := ge.rules.LastPassage()
	epName := eraName(ge.rules, ge.currentEpoch)
	switch how {
	case lastPassageEndured:
		ge.catastropheHistory = append(ge.catastropheHistory,
			fmt.Sprintf("Tick %d"+catHistEnduredMarker+"%s (%s). Prestige kept %d of %d points.", ge.tick, name, epName, points, full))
	case lastPassageSuccumbed:
		ge.catastropheHistory = append(ge.catastropheHistory,
			fmt.Sprintf("Tick %d"+catHistSuccumbedMarker+"%s (%s). No prestige points. Cosmic Legacy earned.", ge.tick, name, epName))
	}
}

// cosmicLegacyFactor is what the Cosmic Legacy multiplies production by: 1.1
// with the legacy, 1 without. recalculateRates applies it after the ×3 caps,
// so it is never in the all-production pool and no cap can swallow it.
// Derived from the flag, never stored as a bonus value, so no reset can lose
// or double it. Read-only.
func (ge *GameEngine) cosmicLegacyFactor() float64 {
	if !ge.cosmicLegacy {
		return 1
	}
	return 1 + CosmicLegacyProductionBonus
}

// cosmicLegacyModifiers is the Cosmic Legacy as the Active Multipliers panel
// shows it (Source "cosmic_legacy"): a multiplier of its own beside the
// all-production pool, like morale's. The pool's sum (AddTotal) ignores it;
// recalculateRates applies cosmicLegacyFactor itself. Read-only.
func (ge *GameEngine) cosmicLegacyModifiers() []Modifier {
	if !ge.cosmicLegacy {
		return nil
	}
	return []Modifier{{Source: cosmicLegacySource, Target: "production_all", Op: OpMul, Value: ge.cosmicLegacyFactor()}}
}

// lastPassageState builds GameState.LastPassage. pointsNow is the run's full
// prestige points. Read-only.
func (ge *GameEngine) lastPassageState(pointsNow int) LastPassageState {
	brace := ge.lastPassageBraceLevel()
	keep := LastPassageKeepFor(brace)
	s := LastPassageState{
		Pending:         ge.pendingLastPassage,
		CosmicLegacy:    ge.cosmicLegacy,
		BraceLevel:      brace,
		KeepPct:         int(math.Round(keep * 100)),
		PointsNow:       pointsNow,
		PointsIfEndured: int(math.Floor(float64(pointsNow) * keep)),
	}
	if h := ge.lastPassageThread(); h != nil {
		s.Invited = h.Invited
	}
	return s
}

// SetCosmicLegacyForTest sets the Cosmic Legacy flag: a test hook for other
// packages (the UI's modal tests, where Succumb must show as closed). Not
// reachable from play. Takes the write lock.
func (ge *GameEngine) SetCosmicLegacyForTest(held bool) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.cosmicLegacy = held
}
