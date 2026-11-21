package ui

import (
	"slices"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// panels_found.go decides which panels the Panels list shows. A panel is
// found by the account once, and stays found; a found panel outside the
// start set is new until it is opened.

// startPanels are found from the start and never marked new.
var startPanels = []string{"people", "research", "logs", "help"}

// plainPanels are found once the town gives them something to show (see
// panelsReached) and are never marked new.
var plainPanels = []string{"charter", "milestones", "badges", "stats", "map"}

// panelView is what the list shows: the panels found and the ones opened.
// A view with no found map shows every panel, none new.
type panelView struct{ found, opened map[string]bool }

func (v panelView) has(name string) bool { return v.found == nil || v.found[name] }

// fresh reports whether name is a found panel that is new: outside the start
// set and not opened yet.
func (v panelView) fresh(name string) bool {
	return v.found != nil && v.found[name] && !v.opened[name] && !slices.Contains(startPanels, name) && !slices.Contains(plainPanels, name)
}

func (v panelView) newCount() int {
	n := 0
	for name := range v.found {
		if v.fresh(name) {
			n++
		}
	}
	return n
}

// panelsReached is the panels the state has made findable. Charter and the
// map and stats: a building stands. Milestones: one has progress. Badges: one
// is earned. Expedition: the
// board has a slot, or a party is out or has come home. Army: a post stands.
// Harbinger: one is present or has been. Epoch: a catastrophe is possible,
// pending or on record.
func panelsReached(st game.GameState) []string {
	out := slices.Clone(startPanels)
	if anyBuilt(st) {
		out = append(out, "charter", "stats", "map")
	}
	if anyMilestoneProgress(st) {
		out = append(out, "milestones")
	}
	if st.AccountStats != nil && st.AccountStats.BadgeSummary.Earned > 0 {
		out = append(out, "badges")
	}
	if i := slices.Index(st.Ruleset().AgeKeys(), st.Age); i >= 1 {
		out = append(out, "plan", "history")
	}
	if anyBuilt(st) && currentWonder(st) != nil {
		out = append(out, "wonders")
	}
	if _, locked := featureRefusal(st, config.FeatureTradeRoutes); !locked {
		out = append(out, "trade")
	}
	if _, locked := featureRefusal(st, config.FeatureDiplomacy); !locked {
		out = append(out, "factions")
	}
	m := st.Military
	if len(m.Board) > 0 || len(m.Active) > 0 || m.CompletedCount > 0 {
		out = append(out, "expedition")
	}
	if m.SoldierCap > 0 || m.SoldiersAway > 0 {
		out = append(out, "army")
	}
	if st.Harbinger != nil || len(st.HarbingerHistory) > 0 {
		out = append(out, "harbinger")
	}
	if st.CatastropheOutlook.Possible || st.PendingCatastrophe != "" || len(st.EpochEventHistory) > 0 {
		out = append(out, "epoch")
	}
	return out
}

// anyBuilt reports whether any building stands.
func anyBuilt(st game.GameState) bool {
	for _, b := range st.Buildings {
		if b.Count > 0 {
			return true
		}
	}
	return false
}

// anyMilestoneProgress reports whether a milestone is done or has a
// condition under way.
func anyMilestoneProgress(st game.GameState) bool {
	for _, m := range st.Milestones.Milestones {
		if m.Completed {
			return true
		}
		for _, p := range m.Progress {
			if p.Current > 0 {
				return true
			}
		}
	}
	return false
}

// panelNamed is the panel a typed first word opens, or "" for none.
func panelNamed(word string) string {
	w := strings.ToLower(word)
	for _, c := range registry() {
		if !c.Dashboard && (c.Name == w || slices.Contains(c.Aliases, w)) {
			w = c.Name
			break
		}
	}
	if isPanel(w) {
		return w
	}
	return ""
}

// notePanels adds names to the found list, and to the opened list when
// opened, and writes the account's settings when either list grew.
func (d *Dashboard) notePanels(names []string, opened bool) {
	if d.panels.found == nil {
		d.panels = panelView{map[string]bool{}, map[string]bool{}}
		if acct := d.engine.Account(); acct != nil {
			f, o := acct.PanelsSeen()
			for _, n := range f {
				d.panels.found[n] = true
			}
			for _, n := range o {
				d.panels.opened[n] = true
			}
		}
	}
	var f, o []string
	for _, n := range names {
		if !d.panels.found[n] {
			d.panels.found[n], f = true, append(f, n)
		}
		if opened && !d.panels.opened[n] {
			d.panels.opened[n], o = true, append(o, n)
		}
	}
	d.body.view = d.panels
	if len(f)+len(o) == 0 {
		return
	}
	if acct := d.engine.Account(); acct != nil {
		_ = acct.NotePanels(f, o) // a display record: a failed write is not worth a line in the log
	}
	d.nextBox.changed()
}

// newPlate is a panel's name on the cyan plate, one cell wider at each side.
func newPlate(s string) string {
	return theme.TagFgBgAttr(theme.RoleBackground, theme.RoleNew, "b") + " " + s + " " + theme.Reset
}
