package mapstyle

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

// world.go says what the cursor says on the world round the town, in the
// game's own words: a civilization met, the route to it, the harbinger, a
// party that is out and the idle workers at the square.

// describeCiv is a civilization met: how it stands toward you, its strength,
// what it trades and how much you have traded.
func describeCiv(m *mapmodel.Model, civ *mapmodel.Faction) Inspection {
	in := Inspection{Title: civ.Name, Command: m.FactionCommand(civ),
		Lines: []string{civ.Relation.String() + vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.01", ", opinion ") + strconv.Itoa(civ.Opinion),
			vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.02", "strength ") + strconv.Itoa(civ.Strength)}}
	if civ.Specialty != "" {
		in.Lines[1] += vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.03", ", trades in ") + civ.Specialty
	}
	if civ.TradeCount > 0 {
		in.Lines = append(in.Lines, plural(civ.TradeCount, vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.04", "trade"),
			vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.05", "trades"))+vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.06", " so far"))
	}
	if civ.LentWorkers > 0 {
		in.Lines = append(in.Lines, plural(civ.LentWorkers, vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.07", "worker"),
			vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.08", "workers"))+vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.09", " lent to you"))
	}
	return in
}

// routeWords are the words for a route on the land (a trail) and in space (a
// lane).
var routeWords = [2]struct{ title, closed, brings, plain, none vibe.Entry }{
	{vibe.E("ui.mapstyle.roguelike.inspect.describeTrail.01", "Trail to the "), vibe.E("ui.mapstyle.roguelike.inspect.describeTrail.02", "at war: the trail is closed"),
		vibe.E("ui.mapstyle.roguelike.inspect.describeTrail.04", " caravans bring "), vibe.E("ui.mapstyle.roguelike.inspect.describeTrail.05", " caravans"),
		vibe.E("ui.mapstyle.roguelike.inspect.describeTrail.06", "no caravans on it")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.describeLane.01", "Lane to the "), vibe.E("ui.mapstyle.roguelike.space_inspect.describeLane.02", "at war: the lane is closed"),
		vibe.E("ui.mapstyle.roguelike.space_inspect.describeLane.04", " brings "), vibe.Same(""),
		vibe.E("ui.mapstyle.roguelike.space_inspect.describeLane.05", "no freight on it")},
}

// describeRoute is the route to a civilization: a trail on the land, a lane in
// space (lane 1), and what it carries.
func describeRoute(m *mapmodel.Model, civ *mapmodel.Faction, lane int) Inspection {
	words := routeWords[lane]
	in := Inspection{Title: words.title.String() + civ.Name, Command: m.FactionCommand(civ)}
	var route *mapmodel.Route
	for i := range m.Routes {
		if m.Routes[i].Civ == civ.Key && route == nil {
			route = &m.Routes[i]
			in.Command = m.RouteCommand(route)
		}
	}
	switch {
	case civ.Relation == mapmodel.RelWar:
		in.Lines = []string{words.closed.String()}
	case route != nil && route.Disrupted:
		in.Lines = []string{route.Name + vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.03", " is disrupted")}
	case route != nil && len(route.Imports) > 0:
		in.Lines = []string{route.Name + words.brings.String() + strings.Join(route.Imports, ", ")}
	case route != nil:
		in.Lines = []string{route.Name + words.plain.String()}
	default:
		in.Lines = []string{words.none.String(), civ.Relation.String()}
	}
	return in
}

// describeHarbinger is the harbinger at the square: what it warns of, the
// odds when the game gives them, and how bad it could be.
func describeHarbinger(h *mapmodel.Harbinger) Inspection {
	in := Inspection{Title: h.Name, Command: mapmodel.CmdHarbinger, Lines: []string{h.WarningLine()}}
	if h.Numeric {
		in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describe.02", "odds ")+strconv.Itoa(int(float64(h.Probability*100)+0.5))+"%")
	}
	if h.Tier != "" {
		in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describe.03", "threat ")+h.Tier)
	}
	return in
}

// missionWords are how a party that is out is named: the scouts and the army.
var missionWords = map[string]struct {
	title   vibe.Entry
	command string
}{
	"scout":    {vibe.E("ui.mapstyle.roguelike.inspect.describe.11", "Scouts: "), mapmodel.CmdExpedition},
	"military": {vibe.E("ui.mapstyle.mission.military.01", "Army: "), mapmodel.CmdArmy},
}

// describeMission is a party that is out, and how long until it reports back.
func describeMission(trip *mapmodel.Trip, key string) Inspection {
	words := missionWords[key]
	return Inspection{Title: words.title.String() + trip.Name, Command: words.command,
		Lines: []string{strconv.Itoa(trip.TicksLeft) + vibe.Line("ui.mapstyle.roguelike.inspect.describe.12", " ticks until they report back")}}
}

// describeIdle is the idle workers at the square.
func describeIdle(m *mapmodel.Model) Inspection {
	return Inspection{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describe.04", "Idle workers"), Command: m.IdleCommand(),
		Lines: []string{plural(m.Workers.Idle, vibe.Line("ui.mapstyle.roguelike.inspect.describe.05", "person"),
			vibe.Line("ui.mapstyle.roguelike.inspect.describe.06", "people")) + vibe.Line("ui.mapstyle.roguelike.inspect.describe.07", " with nothing to do")}}
}
