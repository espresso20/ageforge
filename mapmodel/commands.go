package mapmodel

// commands.go is the command vocabulary the map teaches. Whatever the
// cursor is on, the model says which command a player would type for it,
// so the map speaks the registry's words and a test can check them against
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
	case b.Legacy:
		return ""
	}
	return "build " + b.Key
}

// WonderCommand is the command for a wonder: bank into the one rising,
// otherwise open the wonders panel.
func (m *Model) WonderCommand(w *Wonder) string {
	if w != nil && w.Current && !w.Built {
		return "wonder collect all"
	}
	return "wonders"
}

// Fixed commands for the rest of the map.
const (
	CmdStatus  = "status"
	CmdAdvance = "advance"
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
	add(m.SquareCommand())
	add(CmdStatus)
	return out
}
