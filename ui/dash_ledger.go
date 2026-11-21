package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
)

// dash_ledger.go is what the header's way to the next age reads: the Ledger.
// It works out, once a refresh, the share of the way, the four pillars of
// the gate, what the gate waits on and which resources a blocker is short
// of, and holds the words the header writes about them (dash_header.go draws
// them).
//
// What it says about the gate is the Advance panel's: the pillars, their
// percents, the share of the way and the year are read from the same
// game.AdvanceView (and the engine's held calendar), so the two cannot
// disagree.

// ===== What the Ledger knows =====

// ledgerPillar is one pillar as the Ledger draws it.
type ledgerPillar struct {
	Name  string
	Share float64
	// Pct is the percent the Advance panel writes for the pillar.
	Pct int
	Met bool
	// Count is how many of the pillar's requirements stand, "2/3".
	Count string
	// Status is what ends the pillar's meter in the Advance panel: "✓ stands"
	// or its percent.
	Status string
}

// ledgerData is everything the Ledger draws, worked out once a refresh.
type ledgerData struct {
	v game.AdvanceView
	// Pct is the long bar's percent: the Advance panel's "of the way".
	Pct     int
	Pillars []ledgerPillar
	// Waits are the requirements still to meet, in the Advance panel's order;
	// Then is the wonder that comes after them.
	Waits []string
	Then  string
	// Need are the resources a blocker is short of; Short their names, in
	// the ruleset's order.
	Need  map[string]bool
	Short []string
}

// ledger is the dashboard's Ledger: the last thing it worked out.
type ledger struct {
	data  ledgerData
	ready bool

	// order is the resources in the ruleset's order, cached for the ruleset
	// it was read from.
	set   *rules.Set
	order []string
}

// update works the Ledger out from a snapshot. eng holds the calendar (the
// year never runs back within an age); it may be nil.
func (l *ledger) update(state game.GameState, eng *game.GameEngine, focus string) {
	v := state.AdvanceView()
	if eng != nil {
		v = eng.HoldCalendar(v)
	}
	d := &l.data
	d.v = v
	d.Pct = sharePercent(v.Share, v.Share >= 1)
	d.Pillars = d.Pillars[:0]
	for _, p := range v.Pillars {
		status, _ := pillarStatus(p)
		d.Pillars = append(d.Pillars, ledgerPillar{Name: p.Name, Share: p.Progress, Pct: sharePercent(p.Progress, false), Met: p.Met, Count: pillarCount(p), Status: status})
	}
	l.blockers(state, focus)
	l.ready = true
}

// pillarCount is how many of a pillar's requirements stand, as "2/3".
func pillarCount(p game.GatePillar) string {
	met := 0
	for _, r := range p.Rows {
		if r.Met {
			met++
		}
	}
	return fmt.Sprintf("%d/%d", met, len(p.Rows))
}

// resourceOrder is the ruleset's resources in its own order, then any other
// the snapshot holds by key.
func (l *ledger) resourceOrder(state game.GameState) []string {
	set := state.Ruleset()
	if set != l.set || l.order == nil {
		l.set, l.order = set, l.order[:0]
		seen := map[string]bool{}
		for _, def := range set.Resources() {
			l.order = append(l.order, def.Key)
			seen[def.Key] = true
		}
		rest := keysWhere(state.Resources, func(k string, _ game.ResourceState) bool { return !seen[k] })
		l.order = append(l.order, rest...)
	}
	return l.order
}

// blockers lists what the gate is waiting on and which resources a blocker is
// short of: for a building, the price of its next copy against the stock; for
// the wonder's bank, what is left to bank against the stock.
//
// A resource is named as short only when the building the Next step names
// (focus) or a wonder with something banked needs it and the stock falls
// short: a building going up is paid for.
func (l *ledger) blockers(state game.GameState, focus string) {
	d := &l.data
	d.Waits, d.Then, d.Short = d.Waits[:0], "", d.Short[:0]
	if d.Need == nil {
		d.Need = map[string]bool{}
	}
	clear(d.Need)

	short := func(cost map[string]float64, banked map[string]float64) {
		for res, price := range cost {
			if price-banked[res] > state.Resources[res].Amount {
				d.Need[res] = true
			}
		}
	}

	if _, ok := state.Ruleset().Building(focus); ok {
		short(buildingPrice(&state, focus, state.Buildings[focus]), nil)
	}
	var build *game.GateRow
	bankOpen := false
	for _, p := range d.v.Pillars {
		for i := range p.Rows {
			r := &p.Rows[i]
			if r.Met {
				continue
			}
			switch r.Kind {
			case game.GateRowCount:
				d.Waits = append(d.Waits, fmt.Sprintf("%s %d/%d", r.Label, r.Have, r.Need))
			case game.GateRowAmount:
				d.Waits = append(d.Waits, fmt.Sprintf("%s %s/%s", r.Label, textfmt.Number(r.Value), textfmt.Number(r.Target)))
			case game.GateRowShare:
				bankOpen = true
				d.Waits = append(d.Waits, fmt.Sprintf("%s %d%%", r.Label, r.Have))
				if def, ok := state.Ruleset().Building(r.Key); ok && (r.Key == focus || len(state.Buildings[r.Key].WonderBank) > 0) {
					short(def.BaseCost, state.Buildings[r.Key].WonderBank)
				}
			default:
				if p.Key == config.PillarMonument {
					build = r
					continue
				}
				item := r.Label
				if r.Note != "" {
					item = fmt.Sprintf("%s (%s)", r.Label, r.Note)
				}
				d.Waits = append(d.Waits, item)
			}
		}
	}
	if build != nil {
		if bankOpen || len(d.Waits) > 0 {
			d.Then = build.Label
		} else {
			item := build.Label
			if build.Note != "" {
				item = fmt.Sprintf("%s (%s)", build.Label, build.Note)
			}
			d.Waits = append(d.Waits, item)
		}
	}

	// The names go in the ruleset's order, so the line does not change from
	// frame to frame.
	for _, key := range l.resourceOrder(state) {
		if d.Need[key] {
			d.Short = append(d.Short, state.Resources[key].Name)
		}
	}
}

// ===== What the Ledger says =====

// leadText is the long bar's label.
func (m *ledgerData) leadText() string {
	return rowid.Sprintf("ui.dash_ledger.leadText.01", "To the %s", m.v.NextName)
}

// finalRow is what the bar's row says in the last age.
func (m *ledgerData) finalRow() string {
	return rowid.Line("ui.dash_ledger.finalRow.01", "You have reached the final age.")
}

// readyText is what the waiting line says once every pillar stands, and
// whether it is good news (advancing is open) or a warning.
func (m *ledgerData) readyText() (good bool, text string) {
	if m.v.Ready {
		return true, rowid.Line("ui.dash_ledger.readyText.01", "Every pillar stands. Type advance to go on.")
	}
	return false, rowid.Line("ui.dash_ledger.readyText.02", "Every pillar stands. Answer the catastrophe first (type catastrophe).")
}

// waitText is the line of what the gate waits on, in at most room cells: as
// many of the requirements as fit, then how many more there are, or all of
// them and the wonder that comes after. A row too short for its label and
// one requirement drops the label; one too short even for that cuts the
// first requirement, so the line never says only how many there are.
func (m *ledgerData) waitText(room int) []xseg {
	if len(m.Waits) == 0 {
		return nil
	}
	lead := rowid.Line("ui.dash_ledger.waitText.02", "Waiting on: ")
	parts, shown := m.waitParts(lead, room)
	if shown == 0 {
		parts, shown = m.waitParts("", room)
	}
	if shown == 0 {
		more := m.waitMore(1)
		if cut := room - runeLen(more) - 3; len(m.Waits) > 1 && cut >= 8 {
			return []xseg{{fitText(m.Waits[0], cut), xText}, {" · ", xFaint}, {more, xDim}}
		}
		return []xseg{{fitText(m.Waits[0], max(room, 1)), xText}}
	}
	return parts
}

// waitMore says how many requirements a line leaves out after the first i.
func (m *ledgerData) waitMore(i int) string {
	return rowid.Sprintf("ui.dash_ledger.waitText.04", "+%d", len(m.Waits)-i)
}

// waitNames is the most requirements the waiting line names; the rest are
// counted.
const waitNames = 4

// waitParts is the waiting line after lead in room cells, and how many
// requirements it lists whole.
func (m *ledgerData) waitParts(lead string, room int) (parts []xseg, shown int) {
	const sep = " · "
	sepW := runeLen(sep)
	if lead != "" {
		parts = append(parts, xseg{lead, xDim})
	}
	x := runeLen(lead)
	for i, it := range m.Waits {
		more := m.waitMore(i)
		if i == waitNames {
			return append(parts, xseg{sep, xFaint}, xseg{more, xDim}), i
		}
		need := runeLen(it)
		if i > 0 {
			need += sepW
		}
		if i < len(m.Waits)-1 {
			need += sepW + runeLen(more)
		}
		if x+need > room {
			if i > 0 {
				parts = append(parts, xseg{sep, xFaint})
			}
			return append(parts, xseg{more, xDim}), i
		}
		if i > 0 {
			parts = append(parts, xseg{sep, xFaint})
			x += sepW
		}
		parts = append(parts, xseg{it, xText})
		x += runeLen(it)
	}
	if m.Then != "" {
		if then := rowid.Sprintf("ui.dash_ledger.waitText.01", "then %s", m.Then); x+sepW+runeLen(then) <= room {
			parts = append(parts, xseg{sep, xFaint}, xseg{then, xDim})
		}
	}
	return parts, len(m.Waits)
}

// shortText is the label of the resources a blocker is short of.
func (m *ledgerData) shortText() string {
	return strings.TrimSpace(rowid.Line("ui.dash_ledger.shortText.01", "Short of: "))
}
