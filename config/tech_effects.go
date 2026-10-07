package config

import "fmt"

// TechEffectKind says what one effect of a tech changes. A tech's effect is a
// kind, a target where the kind takes one, and a value: "output of gold,
// +30%", "storage of every resource, +25", "build cost, 5% less". The kind is
// one of the constants below, never a free-form key, so a test can list
// every kind and check that the engine delivers each one.
//
// The kinds are the ones the engine applies today. A kind is added here in
// the same change that teaches the engine to apply it.
type TechEffectKind string

const (
	// EffectOutput adds Value, a fraction (0.3 is +30%), to the output of
	// the resource Target.
	EffectOutput TechEffectKind = "output"
	// EffectAllOutput adds Value, a fraction, to all production. No target.
	EffectAllOutput TechEffectKind = "all_output"
	// EffectWorkerOutput adds Value, a fraction, to what workers gather. No
	// target.
	EffectWorkerOutput TechEffectKind = "worker_output"
	// EffectFlatOutput adds Value of the resource Target every tick.
	EffectFlatOutput TechEffectKind = "flat_output"
	// EffectFlatStorage adds Value to the storage of the resource Target, or
	// of every resource when Target is AllResources.
	EffectFlatStorage TechEffectKind = "flat_storage"
	// EffectFlatHousing adds Value to housing. No target.
	EffectFlatHousing TechEffectKind = "flat_housing"
	// EffectBuildCost adds Value, a fraction, to what buildings cost: -0.05
	// is 5% cheaper. No target.
	EffectBuildCost TechEffectKind = "build_cost"
	// EffectResearchTime takes Value, a fraction, off the time research
	// takes: 0.06 is 6% less. No target.
	EffectResearchTime TechEffectKind = "research_time"
	// EffectGameSpeed adds Value, a fraction, to game speed. No target.
	EffectGameSpeed TechEffectKind = "game_speed"
	// EffectMilitaryPower adds Value, a fraction, to military power. No
	// target.
	EffectMilitaryPower TechEffectKind = "military_power"
	// EffectExpeditionReward adds Value, a fraction, to what expeditions
	// bring back. No target.
	EffectExpeditionReward TechEffectKind = "expedition_reward"
)

// AllResources is the target of a storage effect that raises every store.
const AllResources = "all"

// TechEffectKinds lists every kind, in the order above.
func TechEffectKinds() []TechEffectKind {
	return []TechEffectKind{
		EffectOutput, EffectAllOutput, EffectWorkerOutput,
		EffectFlatOutput, EffectFlatStorage, EffectFlatHousing,
		EffectBuildCost, EffectResearchTime, EffectGameSpeed,
		EffectMilitaryPower, EffectExpeditionReward,
	}
}

// TechEffect is one thing a tech does once it is researched.
type TechEffect struct {
	Kind   TechEffectKind
	Target string  // a resource key (or AllResources) for the kinds that name one; "" for the rest
	Value  float64 // a fraction or an amount: see the kind
}

// TechEffectKey is what a tech effect changes: its kind and its target. Two
// effects with the same key add together.
type TechEffectKey struct {
	Kind   TechEffectKind
	Target string
}

// Key is the kind and target e changes.
func (e TechEffect) Key() TechEffectKey { return TechEffectKey{Kind: e.Kind, Target: e.Target} }

// TakesResource reports whether kind k names a resource in its target.
func (k TechEffectKind) TakesResource() bool {
	return k == EffectOutput || k == EffectFlatOutput || k == EffectFlatStorage
}

// Pool is the bonus pool a fraction of this kind adds to, by the name the
// rest of the game's bonuses use for it ("gold_rate", "production_all"):
// techs share each pool with milestones, wonders and events. The flat kinds
// are amounts and join no pool; they report false, as an unknown kind does.
func (k TechEffectKey) Pool() (string, bool) {
	switch k.Kind {
	case EffectOutput:
		return k.Target + "_rate", true
	case EffectAllOutput:
		return "production_all", true
	case EffectWorkerOutput:
		return "gather_rate", true
	case EffectBuildCost:
		return "build_cost", true
	case EffectResearchTime:
		return "research_speed", true
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
// all of them alike: effect text, cap notes, the bonus guard. A kind that is
// not one of the constants reads as an Effect of that type, which nothing
// applies: Check refuses it in config, and the bonus guard calls it dead.
func (e TechEffect) Effect() Effect {
	switch e.Kind {
	case EffectFlatOutput:
		return Effect{Type: "production", Target: e.Target, Value: e.Value}
	case EffectFlatStorage:
		return Effect{Type: "storage", Target: e.Target, Value: e.Value}
	case EffectFlatHousing:
		return Effect{Type: "capacity", Target: "population", Value: e.Value}
	}
	if pool, ok := e.Key().Pool(); ok {
		return Effect{Type: "bonus", Target: pool, Value: e.Value}
	}
	return Effect{Type: string(e.Kind), Target: e.Target, Value: e.Value}
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

// Check reports what is wrong with e ("" when nothing is): a kind that is
// not one of the constants, a target the kind does not take, a resource
// that does not exist, or a value of zero. isResource says which keys are
// resources.
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
	case !e.Kind.TakesResource():
		if e.Target != "" {
			return fmt.Sprintf("%s takes no target, got %q", e.Kind, e.Target)
		}
	case e.Kind == EffectFlatStorage && e.Target == AllResources:
	case !isResource(e.Target):
		return fmt.Sprintf("%s names %q, which is not a resource", e.Kind, e.Target)
	}
	return ""
}
