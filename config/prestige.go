package config

// PrestigeBonusCoefficient is the square-root form used by Cookie Clicker:
// prestige level grows sublinearly with what was earned. 201 total points
// calibrates to 5x; 120 to about 4.09x.
const PrestigeBonusCoefficient = 0.282

// PrestigeUpgradeDef is retained only to read old signed saves and to keep the
// now-empty shop API wire-compatible during the UI transition.
// Upgrades survive across full resets (DoPrestige/Succumb) — they are the primary
// permanent progression mechanism. Keys are saved: never rename or remove one.
//
// EffectType semantics:
//
//	"rate_bonus"        — multiplier added to EffectKey rate (PerTier = fractional, e.g. 0.05 = +5%)
//	"starting_resource" — added to the player's starting amount of EffectKey resource on reset
//	"legacy"            — a legacy kit item: carries EffectKey's automation across a prestige
type PrestigeUpgradeDef struct {
	Key         string
	Name        string
	Description string
	EffectKey   string  // engine bonus key (e.g. "gather_rate", "population") or resource key for starting_resource
	EffectType  string  // "rate_bonus", "starting_resource" or "legacy"
	PerTier     float64 // bonus increment per tier (meaning depends on EffectType)
	MaxTier     int     // maximum purchasable tier; upgrade is "maxed" when Tier == MaxTier
	Costs       []int   // prestige point cost at each tier; len must equal MaxTier
	// Retired is retained for callers during the UI transition.
	Retired bool
}

// The former legacy kit's keys remain stable for old saves.
const (
	LegacyPlan     = "legacy_plan"
	LegacyWorkers  = "legacy_workers"
	LegacyFactions = "legacy_factions"
)

// LegacyKit lists the former kit keys for compatibility with old saves.
func LegacyKit() []string {
	return []string{LegacyPlan, LegacyWorkers, LegacyFactions}
}

// PrestigeUpgrades no longer exposes a shop. Old save entries remain in the
// serialized upgrades map but have no definitions or effects.
func PrestigeUpgrades() []PrestigeUpgradeDef { return nil }

// PrestigeUpgradeByKey returns no entries now that the shop is gone.
func PrestigeUpgradeByKey() map[string]PrestigeUpgradeDef {
	m := make(map[string]PrestigeUpgradeDef)
	for _, u := range PrestigeUpgrades() {
		m[u.Key] = u
	}
	return m
}

// ActivePrestigeUpgrades is retained for callers during the UI transition.
func ActivePrestigeUpgrades() []PrestigeUpgradeDef { return nil }

// ===== Depth points (Pacing v2) =====

// DepthWeight is what completing age pays at prestige: 3^epoch, counting
// the Stone Era as epoch 0. That is 1 for each Stone Era age, 3 Iron, 9
// Steel, 27 Electric, 81 Digital, 243 Neon and 729 Cosmic; 0 for an
// unknown age. The 22 ages weigh 4,008 in all.
func DepthWeight(age string) int {
	for _, ep := range Epochs() {
		for _, a := range ep.Ages {
			if a != age {
				continue
			}
			w := 1
			for i := 0; i < ep.Order; i++ {
				w *= 3
			}
			return w
		}
	}
	return 0
}

// DepthPoints is what a prestige from age pays: the sum of DepthWeight over
// every age before it (the ages the run completed), with no divisor. The
// Medieval Age pays 9, the Modern Age 120, the Information Age 201, the
// Cyberpunk Age (a run through the Digital Age) 363 and the Interstellar Age
// (through the Space Age) 1,092. 0 for an unknown age.
func DepthPoints(age string) int {
	total := 0
	for _, a := range AgeOrder() {
		if a == age {
			return total
		}
		total += DepthWeight(a)
	}
	return 0
}
