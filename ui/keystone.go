package ui

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// The tech tree's lock, as the panels show it. An age's wonder needs one
// tech, its keystone, before it can be built (its bank fills without it).
// Every place that shows what a wonder needs shows that tech and whether it
// is researched, and every place that lists a tech marks the keystones.

// wonderKeystone reports the keystone tech of wonder key and whether it is
// researched. ok is false for a wonder that needs no tech.
func wonderKeystone(state game.GameState, key string) (name string, researched, ok bool) {
	def, found := state.Ruleset().Building(key)
	if !found || def.RequiredTech == "" {
		return "", false, false
	}
	ts := state.Research.Techs[def.RequiredTech]
	name = ts.Name
	if name == "" {
		name = game.TechName(def.RequiredTech)
	}
	return name, ts.Researched, true
}

// wonderKeystoneLine is the line a wonder's requirements carry for its
// keystone, in the marks the bank lines use: green ✓ when researched, red ✗
// with the command when not. "" for a wonder that needs no tech.
func wonderKeystoneLine(state game.GameState, key string) string {
	name, researched, ok := wonderKeystone(state, key)
	if !ok {
		return ""
	}
	def, _ := state.Ruleset().Building(key)
	if researched {
		return vibe.Sprintf("ui.keystone.wonderKeystoneLine.01", "[green]✓ Keystone: %s, researched[-]", name)
	}
	return vibe.Sprintf("ui.keystone.wonderKeystoneLine.03", "[red]✗ Keystone: %s, not researched (research %s)[-]", name, state.Ruleset().Typed(rules.KindTech, def.RequiredTech))
}

// wonderBuildLine is what a full wonder bank says to do next: build it, or
// research its keystone first.
func wonderBuildLine(state game.GameState, key string) string {
	if name, researched, ok := wonderKeystone(state, key); ok && !researched {
		return vibe.Sprintf("ui.keystone.wonderBuildLine.01", "[yellow]✓ The bank is full. Research %s, then build it with: build %s[-]", name, state.Ruleset().Typed(rules.KindBuilding, key))
	}
	return vibe.Sprintf("ui.keystone.wonderBuildLine.02", "[green]✓ The bank is full. Build it with: build %s[-]", state.Ruleset().Typed(rules.KindBuilding, key))
}

// keystoneMark is the tag a tech carries in the research lists when a
// wonder cannot be built without it: " ★ keystone: Colosseum". "" for any
// other tech.
func keystoneMark(ts game.TechState) string {
	if ts.Kind != config.TechKeystone || ts.KeystoneOf == "" {
		return ""
	}
	return vibe.Sprintf("ui.keystone.keystoneMark.01", "  [gold]★ keystone: %s[-]", game.BuildingName(ts.KeystoneOf))
}
