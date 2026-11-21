package ui

import (
	"slices"
	"strings"

	"github.com/espresso20/ageforge/rules"
)

// nameKinds is, for each argument slot that takes a game thing by its key,
// what the key names.
var nameKinds = map[ArgKind]rules.Kind{
	ArgBuilding:            rules.KindBuilding,
	ArgPlanBuilding:        rules.KindBuilding,
	ArgPlanUpgradeBuilding: rules.KindBuilding,
	ArgSellBuilding:        rules.KindBuilding,
	ArgWorkerBuilding:      rules.KindBuilding,
	ArgStaffedBuilding:     rules.KindBuilding,
	ArgUpgradeBuilding:     rules.KindBuilding,
	ArgTech:                rules.KindTech,
	ArgPlanTech:            rules.KindTech,
	ArgResource:            rules.KindResource,
	ArgWonderResource:      rules.KindResource,
	ArgGatherResource:      rules.KindResource,
	ArgFaction:             rules.KindCiv,
	ArgRouteAvailable:      rules.KindRoute,
	ArgRouteActive:         rules.KindRoute,
}

// plainNames returns words with each game thing spelled as its key, so a
// command takes a thing the way the screen names it: `build Lumber Mill 2`
// becomes `build lumber_mill 2`. In a slot that takes a key, the longest run
// of words that spells one (rules.Set.KeyNamed) is replaced by it. Every
// other word, and a line that is no command in cmds, is returned as it was.
func plainNames(cmds []*Command, words []string, rs *rules.Set) []string {
	p, ok := parse(cmds, words)
	if !ok || len(p.args) == 0 {
		return words
	}
	slots, rest := p.last().Args, p.args
	out := slices.Clone(words[:len(words)-len(rest)])
	for i := 0; len(rest) > 0; i++ {
		if i >= len(slots) {
			return append(out, rest...)
		}
		word, n := rest[0], 1
		if kind, named := nameKinds[slots[i].Kind]; named && !slices.ContainsFunc(slots[i].Words, func(w string) bool { return strings.EqualFold(w, word) }) {
			for run := len(rest); run > 0; run-- {
				if key, ok := rs.KeyNamed(kind, strings.Join(rest[:run], " ")); ok {
					word, n = key, run
					break
				}
			}
		}
		out = append(out, word)
		rest = rest[n:]
	}
	return out
}
