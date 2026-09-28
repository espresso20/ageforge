package main

// commands.go is the walk's verb set. It mirrors the game's prompt: building
// names and keys resolve the same way `build <key>` does, and nothing here
// shadows a dashboard command (see DESIGN.md, "Coexisting with the prompt").

import (
	"math/rand"
	"strings"

	"github.com/espresso20/ageforge/theme"
)

func newRand(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

// dirByWord accepts both registers everywhere, so muscle memory survives the
// move to orbit.
var dirByWord = map[string]Dir{
	"n": North, "north": North, "hubward": North, "up": North,
	"e": East, "east": East, "spinward": East, "right": East,
	"s": South, "south": South, "rimward": South, "down": South,
	"w": West, "west": West, "antispinward": West, "left": West,
}

// Exec runs one typed line and records it in the transcript. It reports
// false when the player leaves the walk.
func (v *View) Exec(line string) bool {
	ok := v.exec(line)
	if v.Echo != "" {
		v.History = append(v.History, Para{{Text: "> " + v.Echo, Role: theme.RoleDim}})
		if len(v.Reply) > 0 {
			v.History = append(v.History, v.Reply)
		}
		if len(v.History) > 40 {
			v.History = v.History[len(v.History)-40:]
		}
	}
	return ok
}

func (v *View) exec(line string) bool {
	line = strings.TrimSpace(line)
	v.Echo = line
	v.Reply = nil
	if line == "" {
		v.Echo = ""
		return true
	}
	verb, arg, _ := strings.Cut(strings.ToLower(line), " ")
	arg = strings.TrimSpace(arg)
	c := v.City
	switch verb {
	case "q", "quit", "back", "exit", "leave":
		return false
	case "l", "look":
		if arg != "" {
			return v.exec("examine " + arg)
		}
		v.Mode, v.Scroll = ModeRoom, 0
	case "?", "help", "h":
		v.Mode, v.Scroll = ModeHelp, 0
	case "survey", "map":
		v.Mode, v.Scroll = ModeSurvey, 0
		v.Reply = Para{{Text: "You take it all in.", Role: theme.RoleText}}
	case "away", "news":
		if v.Prev == nil {
			v.Reply = Para{{Text: "You have not been away.", Role: theme.RoleDim}}
			return true
		}
		v.Mode, v.Scroll = ModeAway, 0
	case "out", "world", "beyond":
		if v.Here != Gate {
			v.Reply = Para{{Text: "The way out is at " + c.Places[Gate].Name + ". ", Role: theme.RoleText}, {Text: "go gate", Role: theme.RoleHighlight}}
			return true
		}
		v.Mode, v.Scroll = ModeWorld, 0
		v.Reply = Para{{Text: "You stand at the edge of the settlement and look out.", Role: theme.RoleText}}
	case "go", "walk", "g":
		if d, ok := dirByWord[arg]; ok {
			v.step(d)
			return true
		}
		v.walkTo(arg, false)
	case "visit", "find":
		v.walkTo(arg, true)
	case "x", "examine", "inspect", "info":
		p := c.Places[v.Here]
		k, b, ok := c.Resolve(arg)
		switch {
		case !ok || b == "":
			v.Reply = Para{{Text: "You see no " + arg + " here.", Role: theme.RoleDim}}
		case k != v.Here:
			v.Reply = Para{{Text: "That is in " + c.Places[k].Name + ". ", Role: theme.RoleText}, {Text: "visit " + arg, Role: theme.RoleHighlight}}
		default:
			_ = p
			v.Mode, v.ExamineKey, v.Scroll = ModeExamine, b, 0
		}
	default:
		if d, ok := dirByWord[verb]; ok && arg == "" {
			v.step(d)
			return true
		}
		// A bare place or building name walks there.
		if _, _, ok := c.Resolve(line); ok {
			v.walkTo(line, true)
			return true
		}
		v.Reply = Para{{Text: "Nobody here knows what \"" + line + "\" means. ", Role: theme.RoleDim}, {Text: "help", Role: theme.RoleHighlight}}
	}
	return true
}

func (v *View) arrive(k PlaceKey) {
	v.Here, v.Mode, v.Scroll = k, ModeRoom, 0
	v.Visited[k] = true
}

func (v *View) step(d Dir) {
	c := v.City
	to, ok := c.Places[v.Here].Exits[d]
	if !ok {
		v.Reply = Para{{Text: "There is no road " + c.DirWord(d) + " from here.", Role: theme.RoleDim}}
		return
	}
	v.arrive(to)
	v.Reply = Para{{Text: "You walk " + c.DirWord(d) + " to ", Role: theme.RoleText}, {Text: c.Places[to].Name, Role: theme.RoleBright}, {Text: ".", Role: theme.RoleText}}
}

// walkTo follows the roads to a place, or to the place a building stands in.
func (v *View) walkTo(q string, examine bool) {
	c := v.City
	k, b, ok := c.Resolve(q)
	if !ok {
		v.Reply = Para{{Text: "Nobody can tell you the way to \"" + q + "\".", Role: theme.RoleDim}}
		return
	}
	route := c.Route(v.Here, k)
	if k == v.Here {
		route = nil
	}
	v.arrive(k)
	if b != "" && examine {
		v.Mode, v.ExamineKey = ModeExamine, b
	}
	var r Para
	switch {
	case len(route) == 0:
		r = Para{{Text: "You are already in " + c.Places[k].Name + ".", Role: theme.RoleText}}
	case len(route) == 1:
		r = Para{{Text: "You walk to ", Role: theme.RoleText}, {Text: c.Places[k].Name, Role: theme.RoleBright}, {Text: ".", Role: theme.RoleText}}
	default:
		var via []string
		for _, s := range route[:len(route)-1] {
			via = append(via, c.Places[s].Name)
		}
		r = Para{{Text: "You walk through " + joinAnd(via) + " to ", Role: theme.RoleText}, {Text: c.Places[k].Name, Role: theme.RoleBright}, {Text: ".", Role: theme.RoleText}}
	}
	v.Reply = r
}

func joinAnd(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}
