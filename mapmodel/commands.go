package mapmodel

import (
	"sort"
	"strconv"
)

// commands.go is the command vocabulary the maps teach. Whatever a style's
// cursor is on, the model says which command a player would type for it,
// so every style speaks the same words and a test can check them against
// the command registry.

// BuildingCommand is the command for a building type: finish an upgrade,
// staff it when it is short of hands and people are idle, otherwise build
// another. Legacy types with nothing to do say so with "".
func (m *Model) BuildingCommand(b *Building) string {
	switch {
	case b == nil:
		return ""
	case b.Wonder:
		return "wonders"
	case b.Upgrade != "" && b.Count > 0:
		return "upgrade " + b.Key
	case b.Understaffed() && m.Workers.Idle > 0:
		return "assign " + b.Key
	case b.Staffing == 0 && b.Count > 0 && m.Workers.Idle > 0:
		return "assign " + b.Key
	case b.Legacy:
		return ""
	}
	return "build " + b.Key
}

// WarningLine is the harbinger's inspect line: what it warns of, in the
// game's words, never the era to come.
func (h *Harbinger) WarningLine() string {
	if h == nil || h.Warning == "" {
		return "warns of impending doom"
	}
	return "warns of " + h.Warning
}

// WonderCommand is the command for a wonder: bank into the one rising,
// otherwise open the wonders panel.
func (m *Model) WonderCommand(w *Wonder) string {
	if w != nil && w.Current && !w.Built {
		return "wonder collect all"
	}
	return "wonders"
}

// FactionCommand is the command for a civ: go and find it, sue for peace,
// mend a feud, or send a gift.
func (m *Model) FactionCommand(f *Faction) string {
	switch {
	case f == nil:
		return ""
	case !f.Discovered:
		return "expedition"
	case f.Relation == RelWar:
		return "diplomacy tribute " + f.Key
	case f.Relation == RelEmbargo || f.Relation == RelRival:
		return "diplomacy neutral " + f.Key
	}
	return "diplomacy gift " + f.Key
}

// RouteCommand is the command for a trade route.
func (m *Model) RouteCommand(*Route) string { return "trade route list" }

// IdleCommand is the command for idle hands: put them in the building
// most short of workers, or recruit when nobody is idle.
func (m *Model) IdleCommand() string {
	var cands []*Building
	for _, b := range m.Buildings {
		if b.Count > 0 && b.Staffing >= 0 && b.Staffing < 1 && !b.Legacy {
			cands = append(cands, b)
		}
	}
	if m.Workers.Idle == 0 || len(cands) == 0 {
		return "recruit"
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].Staffing < cands[j].Staffing })
	return "assign " + cands[0].Key + " " + strconv.Itoa(m.Workers.Idle)
}

// Fixed commands for the rest of the map.
const (
	CmdHarbinger   = "harbinger"
	CmdCatastrophe = "catastrophe"
	CmdExpedition  = "expedition"
	CmdStatus      = "status"
	CmdArmy        = "army"
	CmdAdvance     = "advance"
)

// SquareCommand is the command for the settlement's centre: advance when
// the age is ready, otherwise the status report.
func (m *Model) SquareCommand() string {
	if m.AgeReady {
		return CmdAdvance
	}
	return CmdStatus
}

// AllCommands lists every command the model can produce for this state,
// for tests that check them against the registry.
func (m *Model) AllCommands() []string {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, b := range m.Buildings {
		add(m.BuildingCommand(b))
	}
	for i := range m.Wonders {
		add(m.WonderCommand(&m.Wonders[i]))
	}
	for i := range m.Factions {
		add(m.FactionCommand(&m.Factions[i]))
	}
	for i := range m.Routes {
		add(m.RouteCommand(&m.Routes[i]))
	}
	add(m.IdleCommand())
	add(m.SquareCommand())
	for _, c := range []string{CmdHarbinger, CmdCatastrophe, CmdExpedition, CmdStatus, CmdArmy} {
		add(c)
	}
	return out
}
