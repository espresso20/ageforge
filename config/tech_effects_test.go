package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// techEffectsBeforeTree is every tech's effects as the game wrote them
// before effects were typed: the Effects of each tech in the config export
// (cmd/export_config, techs.json) at the last commit before the tech tree
// work, a461a0e.
const techEffectsBeforeTree = "testdata/tech_effects_before_tree.json"

// TestTechEffectsUnchangedByTyping: typing the effects changed how they are
// written, not what any tech does. Every tech's typed effects, read back as
// general Effects, are the type, target and value it had before, in the
// same order, and no tech was added or lost.
//
// The tree's third PR gives the 77 techs new effects. This test and its
// file go then: they pin the old ones.
func TestTechEffectsUnchangedByTyping(t *testing.T) {
	data, err := os.ReadFile(techEffectsBeforeTree)
	if err != nil {
		t.Fatal(err)
	}
	var before map[string][]Effect
	if err := json.Unmarshal(data, &before); err != nil {
		t.Fatal(err)
	}
	techs := Technologies()
	if len(techs) != len(before) {
		t.Errorf("%d techs now, %d before", len(techs), len(before))
	}
	for _, tech := range techs {
		want, ok := before[tech.Key]
		if !ok {
			t.Errorf("%s was not a tech before", tech.Key)
			continue
		}
		if got := tech.GeneralEffects(); !reflect.DeepEqual(got, want) {
			t.Errorf("%s does %+v now, did %+v before", tech.Key, got, want)
		}
	}
}

// TestTechEffectKindsMapToPools pins which pool each kind adds to and how
// each kind reads as a general Effect. The engine applies a tech's bonus
// through these names, so a kind that moved pools would change what its
// techs do.
func TestTechEffectKindsMapToPools(t *testing.T) {
	cases := []struct {
		eff  TechEffect
		pool string // "" for the flat kinds, which join no pool
		want Effect
	}{
		{TechEffect{EffectOutput, "gold", 0.3}, "gold_rate", Effect{"bonus", "gold_rate", 0.3}},
		{TechEffect{EffectAllOutput, "", 0.5}, "production_all", Effect{"bonus", "production_all", 0.5}},
		{TechEffect{EffectWorkerOutput, "", 0.15}, "gather_rate", Effect{"bonus", "gather_rate", 0.15}},
		{TechEffect{EffectBuildCost, "", -0.05}, "build_cost", Effect{"bonus", "build_cost", -0.05}},
		{TechEffect{EffectResearchTime, "", 0.06}, "research_speed", Effect{"bonus", "research_speed", 0.06}},
		{TechEffect{EffectGameSpeed, "", 0.05}, "tick_speed", Effect{"bonus", "tick_speed", 0.05}},
		{TechEffect{EffectMilitaryPower, "", 0.2}, "military_power", Effect{"bonus", "military_power", 0.2}},
		{TechEffect{EffectExpeditionReward, "", 0.3}, "expedition_reward", Effect{"bonus", "expedition_reward", 0.3}},
		{TechEffect{EffectFlatOutput, "food", 0.1}, "", Effect{"production", "food", 0.1}},
		{TechEffect{EffectFlatStorage, "gold", 100}, "", Effect{"storage", "gold", 100}},
		{TechEffect{EffectFlatStorage, AllResources, 25}, "", Effect{"storage", "all", 25}},
		{TechEffect{EffectFlatHousing, "", 5}, "", Effect{"capacity", "population", 5}},
	}
	covered := map[TechEffectKind]bool{}
	for _, c := range cases {
		covered[c.eff.Kind] = true
		pool, ok := c.eff.Key().Pool()
		if pool != c.pool || ok != (c.pool != "") {
			t.Errorf("%+v adds to pool %q (a pool: %v), want %q", c.eff, pool, ok, c.pool)
		}
		if got := c.eff.Effect(); got != c.want {
			t.Errorf("%+v reads as %+v, want %+v", c.eff, got, c.want)
		}
	}
	for _, k := range TechEffectKinds() {
		if !covered[k] {
			t.Errorf("kind %q has no case here: say which pool it adds to", k)
		}
	}
	// A kind nobody defined keeps its name, so what reads it can say what
	// it is looking at; it joins no pool.
	stray := TechEffect{Kind: "made_up", Target: "food", Value: 1}
	if got := stray.Effect(); got != (Effect{"made_up", "food", 1}) {
		t.Errorf("an unknown kind reads as %+v, want its own name as the type", got)
	}
	if pool, ok := stray.Key().Pool(); ok {
		t.Errorf("an unknown kind adds to pool %q, want none", pool)
	}
}

// TestTechEffectPoolsNeverCollide: a pool belongs to one kind and target.
// Were a resource ever named so that its output pool spelled another kind's
// pool ("gather" would give "gather_rate"), two different effects would add
// into one number.
func TestTechEffectPoolsNeverCollide(t *testing.T) {
	seen := map[string]TechEffectKey{}
	claim := func(k TechEffectKey) {
		pool, ok := k.Pool()
		if !ok {
			return
		}
		if other, taken := seen[pool]; taken && other != k {
			t.Errorf("pool %q is fed by both %+v and %+v", pool, other, k)
		}
		seen[pool] = k
	}
	for _, kind := range TechEffectKinds() {
		if !kind.TakesResource() {
			claim(TechEffectKey{Kind: kind})
			continue
		}
		for _, r := range BaseResources() {
			claim(TechEffectKey{Kind: kind, Target: r.Key})
		}
	}
}

// TestTechEffectCheck: the check that guards config catches each way a
// typed effect can be wrong.
func TestTechEffectCheck(t *testing.T) {
	isResource := func(key string) bool { return key == "food" || key == "gold" }
	good := []TechEffect{
		{EffectOutput, "food", 0.1},
		{EffectFlatOutput, "gold", 2},
		{EffectFlatStorage, "gold", 100},
		{EffectFlatStorage, AllResources, 25},
		{EffectFlatHousing, "", 5},
		{EffectBuildCost, "", -0.05},
	}
	for _, e := range good {
		if problem := e.Check(isResource); problem != "" {
			t.Errorf("%+v: %s", e, problem)
		}
	}
	bad := map[string]TechEffect{
		"an unknown kind":                 {"gather_rate", "", 0.1},
		"a resource that does not exist":  {EffectOutput, "foods", 0.1},
		"every resource, for output":      {EffectOutput, AllResources, 0.1},
		"a missing resource":              {EffectFlatOutput, "", 1},
		"a target on a kind without one":  {EffectGameSpeed, "food", 0.05},
		"a housing target":                {EffectFlatHousing, "population", 5},
		"no value":                        {EffectMilitaryPower, "", 0},
		"an old pool name as the target":  {EffectOutput, "gold_rate", 0.3},
		"an old bonus key as the kind":    {"bonus", "gold_rate", 0.3},
		"storage for a made-up resource":  {EffectFlatStorage, "everything", 10},
		"all production written as a res": {EffectAllOutput, AllResources, 0.3},
	}
	for name, e := range bad {
		if e.Check(isResource) == "" {
			t.Errorf("%s (%+v) passed the check", name, e)
		}
	}
}
