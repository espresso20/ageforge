package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The tech tree's lock, as the panels show it. An age's wonder needs one
// tech, its keystone, before it can be built (its bank fills without it).
// Every place that shows what a wonder needs shows that tech and whether it
// is researched, and every place that lists a tech marks the keystones.

// wonderKeystone reports the keystone tech of wonder key and where it
// stands: researched, still locking the wonder, or waived (a save from
// before the lock keeps the age it was in exempt; the engine then reports
// no tech for the wonder to wait for). ok is false for a wonder that needs
// no tech.
func wonderKeystone(state game.GameState, key string) (name string, researched, waived, ok bool) {
	def, found := state.Ruleset().Building(key)
	if !found || def.RequiredTech == "" {
		return "", false, false, false
	}
	ts := state.Research.Techs[def.RequiredTech]
	name = ts.Name
	if name == "" {
		name = game.TechName(def.RequiredTech)
	}
	return name, ts.Researched, !ts.Researched && state.Buildings[key].NeedsTech == "", true
}

// wonderKeystoneLine is the line a wonder's requirements carry for its
// keystone, in the marks the bank lines use: green ✓ when researched, red ✗
// with the command when not, gray when this age is exempt. "" for a wonder
// that needs no tech.
func wonderKeystoneLine(state game.GameState, key string) string {
	name, researched, waived, ok := wonderKeystone(state, key)
	if !ok {
		return ""
	}
	def, _ := state.Ruleset().Building(key)
	switch {
	case researched:
		return fmt.Sprintf("[green]✓ Keystone: %s, researched[-]", name)
	case waived:
		return fmt.Sprintf("[gray]Keystone: %s, not needed in this age (the lock starts when you next advance)[-]", name)
	}
	return fmt.Sprintf("[red]✗ Keystone: %s, not researched (research %s)[-]", name, def.RequiredTech)
}

// wonderBuildLine is what a full wonder bank says to do next: build it, or
// research its keystone first.
func wonderBuildLine(state game.GameState, key string) string {
	if name, researched, waived, ok := wonderKeystone(state, key); ok && !researched && !waived {
		return fmt.Sprintf("[yellow]✓ The bank is full. Research %s, then build it with: build %s[-]", name, key)
	}
	return fmt.Sprintf("[green]✓ The bank is full. Build it with: build %s[-]", key)
}

// keystoneMark is the tag a tech carries in the research lists when a
// wonder cannot be built without it: " ★ keystone: Colosseum". "" for any
// other tech.
func keystoneMark(ts game.TechState) string {
	if ts.Kind != config.TechKeystone || ts.KeystoneOf == "" {
		return ""
	}
	return fmt.Sprintf("  [gold]★ keystone: %s[-]", game.BuildingName(ts.KeystoneOf))
}

// keystoneLegend explains the mark, once per list.
const keystoneLegend = "[gray]★ keystone: the one tech its age's wonder needs before it can be built.[-]"
