package config

import (
	"fmt"
	"math"
)

// TechEffectKind says what one effect of a tech changes. A tech's effect is a
// kind, a target where the kind takes one, and a value: "output of gold,
// +10%", "storage, +10%", "build cost, 3% less". The kind is one of the
// constants below, never a free-form key, so a test can list every kind and
// check that the engine delivers each one.
//
// A tech's bonus has its own layer. The game's other bonuses (milestones,
// wonders, events, festivals, boons) add up in pools, and a pool is clamped:
// all production and each resource's output stop at x3. A tech's output bonus
// joins no pool. It multiplies what is left after the clamp, where Era
// Mastery's speed does, so it counts in full whatever else the player holds.
// Inside the layer, bonuses on the same thing add up (two +5% give +10%) and
// nothing caps the sum. A bonus that cuts a time or a price multiplies with
// the others of its kind instead (two 3% cuts leave 0.97 x 0.97) and has a
// floor (TechEffectKey.Limits); TestTechHeadroom checks that every tech
// taken together stays clear of each one.
//
// The kinds are the ones the engine applies today. A kind is added here in
// the same change that teaches the engine to apply it.
type TechEffectKind string

const (
	// EffectOutput adds Value, a fraction (0.1 is +10%), to the output of
	// the resource Target, in the tech layer.
	EffectOutput TechEffectKind = "output"
	// EffectAllOutput adds Value, a fraction, to the output of every
	// resource, in the tech layer. No target.
	EffectAllOutput TechEffectKind = "all_output"
	// EffectFlatOutput adds Value of the resource Target every tick. It is
	// for a tech that is the first source of a resource in its age (Steel
	// Forging's steel), not for a bonus.
	EffectFlatOutput TechEffectKind = "flat_output"
	// EffectStorage adds Value, a fraction, to every store. No target.
	EffectStorage TechEffectKind = "storage"
	// EffectHousing adds Value, a fraction, to housing. No target.
	EffectHousing TechEffectKind = "housing"
	// EffectBuildCost changes what buildings cost by Value, a fraction:
	// -0.03 is 3% less. No target.
	EffectBuildCost TechEffectKind = "build_cost"
	// EffectBuildTime changes how long construction takes by Value, a
	// fraction: -0.08 is 8% less. No target.
	EffectBuildTime TechEffectKind = "build_time"
	// EffectResearchTime changes how long research takes by Value, a
	// fraction: -0.03 is 3% less. No target.
	EffectResearchTime TechEffectKind = "research_time"
	// EffectGameSpeed adds Value, a fraction, to game speed. No target.
	EffectGameSpeed TechEffectKind = "game_speed"
	// EffectMilitaryPower adds Value, a fraction, to military power. No
	// target.
	EffectMilitaryPower TechEffectKind = "military_power"
	// EffectExpeditionReward adds Value, a fraction, to what expeditions
	// bring back. No target.
	EffectExpeditionReward TechEffectKind = "expedition_reward"
	// EffectMechanic changes one number in a mechanic: Target is the
	// mechanic's key (Mechanics), and Value multiplies the number (-0.15
	// leaves 85% of it) or is added to it, as the mechanic says.
	EffectMechanic TechEffectKind = "mechanic"
)

// The floors of the kinds that cut a time or a price: however many techs cut
// it, construction never costs less than BuildCostFloor of its price or
// takes less than BuildTimeFloor of its time, and research never takes less
// than ResearchTimeFloor of its time.
const (
	BuildCostFloor    = 0.10
	BuildTimeFloor    = 0.40
	ResearchTimeFloor = 0.50
)

// TechEffectKinds lists every kind, in the order above.
func TechEffectKinds() []TechEffectKind {
	return []TechEffectKind{
		EffectOutput, EffectAllOutput, EffectFlatOutput,
		EffectStorage, EffectHousing,
		EffectBuildCost, EffectBuildTime, EffectResearchTime,
		EffectGameSpeed, EffectMilitaryPower, EffectExpeditionReward,
		EffectMechanic,
	}
}

// TechEffect is one thing a tech does once it is researched.
type TechEffect struct {
	Kind   TechEffectKind
	Target string  // a resource key or a mechanic key for the kinds that name one; "" for the rest
	Value  float64 // a fraction or an amount: see the kind
}

// TechEffectKey is what a tech effect changes: its kind and its target. Two
// effects with the same key stack (TechEffectKey.Fold).
type TechEffectKey struct {
	Kind   TechEffectKind
	Target string
}

// Key is the kind and target e changes.
func (e TechEffect) Key() TechEffectKey { return TechEffectKey{Kind: e.Kind, Target: e.Target} }

// TakesResource reports whether kind k names a resource in its target.
func (k TechEffectKind) TakesResource() bool {
	return k == EffectOutput || k == EffectFlatOutput
}

// Multiplies reports whether effects of this key multiply together: each
// takes its share of what the ones before it left. The kinds that cut a time
// or a price do, and a mechanic says for itself. Every other key adds up.
func (k TechEffectKey) Multiplies() bool {
	switch k.Kind {
	case EffectBuildCost, EffectBuildTime, EffectResearchTime:
		return true
	case EffectMechanic:
		return MechanicByKey()[k.Target].Multiplies
	}
	return false
}

// Start is the key's term with no tech researched: 1 for a key that
// multiplies, 0 for one that adds.
func (k TechEffectKey) Start() float64 {
	if k.Multiplies() {
		return 1
	}
	return 0
}

// Fold takes one more effect of value v into term: term x (1 + v) for a key
// that multiplies, term + v for one that adds.
func (k TechEffectKey) Fold(term, v float64) float64 {
	if k.Multiplies() {
		return float64(term * (1 + v))
	}
	return term + v
}

// Limits is the range the key's term is held to: the floor of a time or a
// price, a mechanic's own limits. -Inf and +Inf where there is none.
func (k TechEffectKey) Limits() (lo, hi float64) {
	switch k.Kind {
	case EffectBuildCost:
		return BuildCostFloor, 1
	case EffectBuildTime:
		return BuildTimeFloor, 1
	case EffectResearchTime:
		return ResearchTimeFloor, 1
	case EffectMechanic:
		if m, ok := MechanicByKey()[k.Target]; ok {
			return m.Min, m.Max
		}
	}
	return math.Inf(-1), math.Inf(1)
}

// Clamp holds term inside the key's limits.
func (k TechEffectKey) Clamp(term float64) float64 {
	lo, hi := k.Limits()
	return math.Min(math.Max(term, lo), hi)
}

// Pool is the bonus pool a fraction of this kind adds to, by the name the
// rest of the game's bonuses use for it ("tick_speed", "military_power"):
// game speed, military power and expedition rewards, which techs share with
// milestones and wonders and which no x3 clamp holds. Every other kind is
// applied in the tech layer and reports false, as an unknown kind does.
func (k TechEffectKey) Pool() (string, bool) {
	switch k.Kind {
	case EffectGameSpeed:
		return "tick_speed", true
	case EffectMilitaryPower:
		return "military_power", true
	case EffectExpeditionReward:
		return "expedition_reward", true
	}
	return "", false
}

// Effect is e as the general Effect every other source of bonuses is written
// in (a building's, a milestone's, an event's), for the code that handles
// all of them alike: cap notes, the bonus guard. Only an effect that joins a
// pool, and a flat output, has one. The layer's kinds read as an Effect of
// type "tech_<kind>", which no pool holds and nothing caps; a kind that is
// not one of the constants does too, and Check refuses it in config.
func (e TechEffect) Effect() Effect {
	if e.Kind == EffectFlatOutput {
		return Effect{Type: "production", Target: e.Target, Value: e.Value}
	}
	if pool, ok := e.Key().Pool(); ok {
		return Effect{Type: "bonus", Target: pool, Value: e.Value}
	}
	return Effect{Type: "tech_" + string(e.Kind), Target: e.Target, Value: e.Value}
}

// GeneralEffects is every effect of the tech as a general Effect, in order
// (see TechEffect.Effect).
func (t TechDef) GeneralEffects() []Effect {
	if len(t.Effects) == 0 {
		return nil
	}
	out := make([]Effect, len(t.Effects))
	for i, e := range t.Effects {
		out[i] = e.Effect()
	}
	return out
}

// Text is e in player words, without a capital or a full stop, the way the
// Research panel lists effects: "+10% food production", "+10% storage",
// "buildings cost 3% less", "market fee 3 points lower".
func (e TechEffect) Text() string {
	pct := FormatPercent(e.Value)
	switch e.Kind {
	case EffectOutput:
		return signed(e.Value, pct) + " " + ResourceLabel(e.Target) + " production"
	case EffectAllOutput:
		return signed(e.Value, pct) + " all production"
	case EffectFlatOutput:
		return signed(e.Value, FormatAmount(e.Value)) + " " + ResourceLabel(e.Target) + "/tick"
	case EffectStorage:
		return signed(e.Value, pct) + " storage"
	case EffectHousing:
		return signed(e.Value, pct) + " housing"
	case EffectBuildCost:
		return "buildings cost " + pct + " " + lessOrMore(e.Value)
	case EffectBuildTime:
		return "construction takes " + pct + " " + lessOrMore(e.Value) + " time"
	case EffectResearchTime:
		return "research takes " + pct + " " + lessOrMore(e.Value) + " time"
	case EffectGameSpeed:
		return signed(e.Value, pct) + " game speed"
	case EffectMilitaryPower:
		return signed(e.Value, pct) + " military power"
	case EffectExpeditionReward:
		return signed(e.Value, pct) + " expedition rewards"
	case EffectMechanic:
		if m, ok := MechanicByKey()[e.Target]; ok {
			return m.EffectText(e.Value)
		}
	}
	return string(e.Kind) + " " + signed(e.Value, FormatAmount(e.Value))
}

// lessOrMore is "less" for a cut and "more" for a rise.
func lessOrMore(v float64) string {
	if v < 0 {
		return "less"
	}
	return "more"
}

// Check reports what is wrong with e ("" when nothing is): a kind that is
// not one of the constants, a target the kind does not take, a resource or a
// mechanic that does not exist, a value of zero, or a cut so deep it leaves
// nothing (a time or a price multiplied by zero or less). isResource says
// which keys are resources.
func (e TechEffect) Check(isResource func(string) bool) string {
	known := false
	for _, k := range TechEffectKinds() {
		known = known || k == e.Kind
	}
	switch {
	case !known:
		return fmt.Sprintf("unknown effect kind %q", e.Kind)
	case e.Value == 0:
		return fmt.Sprintf("%s effect with no value", e.Kind)
	case e.Kind == EffectMechanic:
		if _, ok := MechanicByKey()[e.Target]; !ok {
			return fmt.Sprintf("%s names %q, which is not a mechanic", e.Kind, e.Target)
		}
	case !e.Kind.TakesResource():
		if e.Target != "" {
			return fmt.Sprintf("%s takes no target, got %q", e.Kind, e.Target)
		}
	case !isResource(e.Target):
		return fmt.Sprintf("%s names %q, which is not a resource", e.Kind, e.Target)
	}
	if e.Key().Multiplies() && e.Value <= -1 {
		return fmt.Sprintf("%s of %v leaves nothing", e.Kind, e.Value)
	}
	return ""
}

// TechTerms folds the effects of techs into one term per key, in the order
// given: the sum for a key that adds, the product for one that multiplies,
// before any limit (TechEffectKey.Clamp applies it). A key no tech touches
// is absent: read it with TechEffectKey.Start.
func TechTerms(techs []TechDef) map[TechEffectKey]float64 {
	out := map[TechEffectKey]float64{}
	for _, t := range techs {
		for _, e := range t.Effects {
			k := e.Key()
			cur, ok := out[k]
			if !ok {
				cur = k.Start()
			}
			out[k] = k.Fold(cur, e.Value)
		}
	}
	return out
}
