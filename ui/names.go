package ui

import (
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// names.go: the command line and the names VibeMode shows.
//
// A command takes a config key ("hut"). While VibeMode is on the screen shows
// another name for most of them ("Crash Pad"), so the prompt offers that name
// in the shape of a key ("crash_pad") and every command takes it, the key and
// the plain name as well. Nothing here runs with VibeMode off.

// nameKind is the kind of name a slot of kind k takes in the ruleset.
func nameKind(k ArgKind) (rules.Kind, bool) {
	switch k {
	case ArgBuilding, ArgPlanBuilding, ArgPlanUpgradeBuilding, ArgSellBuilding,
		ArgWorkerBuilding, ArgStaffedBuilding, ArgUpgradeBuilding:
		return rules.KindBuilding, true
	case ArgTech, ArgPlanTech:
		return rules.KindTech, true
	case ArgResource, ArgWonderResource, ArgGatherResource:
		return rules.KindResource, true
	case ArgFaction:
		return rules.KindCiv, true
	case ArgRouteAvailable, ArgRouteActive:
		return rules.KindRoute, true
	}
	return 0, false
}

// typedName is key as a player types it while VibeMode is on: the slug of the
// name the screen shows.
func typedName(set *rules.Set, k ArgKind, key string) string {
	if kind, ok := nameKind(k); ok {
		return set.Typed(kind, key)
	}
	if k == ArgTheme {
		if t, ok := theme.ByKey(key); ok && vibe.On() {
			if s := rules.Slug(t.Name.String()); s != "" {
				return s
			}
		}
	}
	return key
}

// typedNames is keys as typedName spells each one.
func typedNames(set *rules.Set, k ArgKind, keys []string) []string {
	if !vibe.On() {
		return keys
	}
	out := make([]string, len(keys))
	for i, key := range keys {
		out[i] = typedName(set, k, key)
	}
	return out
}

// keyOfTyped reads w, typed for a slot of kind k, back to the key it names:
// the key, the displayed name or the plain name, and with prefix a first word
// of a name that no other starts. ok is false when it is none.
func keyOfTyped(set *rules.Set, k ArgKind, w string, prefix bool) (string, bool) {
	w = strings.ToLower(w)
	if kind, ok := nameKind(k); ok {
		return set.KeyFor(kind, w, prefix)
	}
	if k != ArgTheme {
		return "", false
	}
	t := rules.Slug(w)
	var found string
	for _, th := range theme.All() {
		switch {
		case th.Key == t || rules.Slug(th.Name.String()) == t || rules.Slug(th.Name.Plain()) == t:
			return th.Key, true
		case prefix && strings.HasPrefix(rules.Slug(th.Name.String()), t+"_"):
			if found != "" {
				return "", false
			}
			found = th.Key
		}
	}
	return found, found != ""
}

// vibeArgs puts the keys in for the names typed after command cmd: each word
// (or run of words, for a name typed with spaces) in a slot that takes a game
// thing becomes its key. Words that name nothing are left as typed, and the
// command says what is wrong with them. With VibeMode off the args are as
// they came.
func vibeArgs(cmd string, args []string, engine *game.GameEngine) []string {
	if !vibe.On() || engine == nil || len(args) == 0 {
		return args
	}
	p, ok := parse(registry(), append([]string{cmd}, args...))
	if !ok {
		return args
	}
	set := engine.Rules()
	node := p.last()
	skip := len(args) - len(p.args)
	out := slices.Clone(args[:skip])
	rest := p.args
	j := 0
	for slot := 0; j < len(rest); slot++ {
		if slot >= len(node.Args) {
			out = append(out, rest[j:]...)
			break
		}
		a := node.Args[slot]
		word, used := rest[j], 1
		_, named := nameKind(a.Kind)
		if (named || a.Kind == ArgTheme) && !slices.ContainsFunc(a.Words, func(x string) bool { return strings.EqualFold(x, word) }) {
			left := len(rest) - j
			longest := min(4, left)
			if a.Kind == ArgTech || a.Kind == ArgPlanTech {
				longest = left // a tech's words are all the rest of the line
			}
			resolved := false
			for n := longest; n >= 2 && !resolved; n-- {
				if key, ok := keyOfTyped(set, a.Kind, strings.Join(rest[j:j+n], "_"), false); ok {
					word, used, resolved = key, n, true
				}
			}
			if !resolved && !(longest == left && left > 1 && (a.Kind == ArgTech || a.Kind == ArgPlanTech)) {
				if key, ok := keyOfTyped(set, a.Kind, word, true); ok {
					word = key
				}
			}
		}
		out = append(out, word)
		j += used
	}
	return out
}
