package ui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// Adding an item from the panel: A opens a name box at the foot. The name is
// matched the way the prompt matches (matchOffers) against what the plan
// takes: this age's buildings, upgrades, techs and the board's missions.
// Enter adds when one thing fits, asks which kind when several do, and puts
// the item below the selection. Esc cancels.

// planPick is one thing the typed name can mean.
type planPick struct {
	kind  string // game.PlanBuild, PlanUpgrade, PlanResearch or PlanExpedition
	key   string // the building or tech key, or the offer's title
	label string
}

// planKindWord is the kind as a player reads it in a choice. One lower-case
// word is a key to the export, so these are plain.
func planKindWord(kind string) string {
	switch kind {
	case game.PlanUpgrade:
		return "upgrade"
	case game.PlanResearch:
		return "research"
	case game.PlanExpedition:
		return "expedition"
	}
	return "build"
}

// planPicks is every item the typed name can mean, builds first.
func planPicks(state game.GameState, engine *game.GameEngine, input string) []planPick {
	var out []planPick
	collect := func(kind string, keys []string, name func(string) string) {
		names := make([]string, len(keys))
		for i, k := range keys {
			names[i] = name(k)
		}
		// Match the name, and the key spelled out as words ("wood_camp").
		seen := map[int]bool{}
		for _, i := range matchOffers(names, input) {
			seen[i] = true
		}
		for _, i := range matchOffers(keys, input) {
			seen[i] = true
		}
		for i := range keys {
			if seen[i] {
				out = append(out, planPick{kind: kind, key: keys[i], label: names[i]})
			}
		}
	}
	collect(game.PlanBuild, plannableBuildingKeys(state), game.BuildingName)
	if engine != nil {
		collect(game.PlanUpgrade, upgradeableBuildingKeys(engine), game.BuildingName)
	}
	collect(game.PlanResearch, plannableTechKeys(state), func(k string) string {
		def, _ := state.Ruleset().Tech(k)
		return def.Name
	})
	var titles []string
	for _, slot := range state.Military.Board {
		if slot.Offer != nil {
			titles = append(titles, slot.Offer.Name)
		}
	}
	collect(game.PlanExpedition, titles, func(t string) string { return t })
	return out
}

// handleAddKey is the key handler while the name box (or the choice) is open.
func (p *planPanel) handleAddKey(event *tcell.EventKey, engine *game.GameEngine) bool {
	switch event.Key() {
	case tcell.KeyEscape:
		p.mode, p.input, p.picks = planModeQueue, "", nil
		p.note = ""
	case tcell.KeyUp:
		if p.mode == planModePick {
			p.pickSel = clampSel(p.pickSel-1, len(p.picks))
		}
	case tcell.KeyDown:
		if p.mode == planModePick {
			p.pickSel = clampSel(p.pickSel+1, len(p.picks))
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if r := []rune(p.input); len(r) > 0 {
			p.input = string(r[:len(r)-1])
		}
		p.mode, p.picks = planModeAdd, nil
	case tcell.KeyEnter:
		p.enterAdd(engine)
	case tcell.KeyRune:
		if unicode.IsPrint(event.Rune()) {
			p.input += string(event.Rune())
			p.mode, p.picks = planModeAdd, nil
		}
	}
	return true
}

// enterAdd acts on Enter in the name box.
func (p *planPanel) enterAdd(engine *game.GameEngine) {
	if p.mode == planModePick {
		p.add(engine, p.picks[clampSel(p.pickSel, len(p.picks))])
		return
	}
	picks := planPicks(engine.GetState(), engine, p.input)
	switch len(picks) {
	case 0:
		p.note, p.noteGood = vibe.Sprintf("ui.plan_add.enterAdd.01", "Nothing called %q fits the plan.", p.input), false
	case 1:
		p.add(engine, picks[0])
	default:
		p.mode, p.picks, p.pickSel = planModePick, picks, 0
		p.note = ""
	}
}

// add puts pick into the plan, below the selection, and closes the box.
func (p *planPanel) add(engine *game.GameEngine, pick planPick) {
	before := len(engine.GetState().Plan)
	var err error
	switch pick.kind {
	case game.PlanBuild:
		_, err = engine.PlanAddBuild(pick.key, 1)
	case game.PlanUpgrade:
		_, err = engine.PlanAddUpgrade(pick.key, 1)
	case game.PlanResearch:
		_, err = engine.PlanAddResearchChain(pick.key)
	case game.PlanExpedition:
		err = engine.PlanAddExpedition(pick.key)
	}
	if err != nil {
		p.mode, p.picks = planModeQueue, nil
		p.note, p.noteGood = err.Error(), false
		return
	}
	// New items land at the end; take them up to just below the selection.
	after := len(engine.GetState().Plan)
	at := p.sel + 2
	if before == 0 {
		at = 1
	}
	for i := 0; i < after-before; i++ {
		from := before + 1 + i
		if _, err := engine.PlanMove(from, at+i-from); err != nil {
			break
		}
	}
	if after > before {
		p.sel = at - 1
	} else {
		p.sel = max(0, after-1) // merged into the last item
	}
	p.mode, p.input, p.picks = planModeQueue, "", nil
	p.note, p.noteGood = vibe.Sprintf("ui.plan_add.add.01", "Added %s.", pick.label), true
}

// planAddLines is the name box and what the name fits so far.
func planAddLines(sc planScreen) []string {
	box := " " + theme.Paint(theme.RoleAccent, vibe.Line("ui.plan_add.planAddLines.01", "Add: ")) + tview.Escape(sc.input) + theme.Paint(theme.RoleHighlight, "▏")
	hint := vibe.Line("ui.plan_add.planAddLines.02", "type a building, tech or mission   Enter add below   Esc cancel")
	if sc.input != "" {
		fits := planPicks(sc.state, nil, sc.input)
		var names []string
		for i, f := range fits {
			if i == 4 {
				names = append(names, "…")
				break
			}
			names = append(names, f.label)
		}
		if len(names) > 0 {
			hint = strings.Join(names, ", ")
		} else {
			hint = vibe.Line("ui.plan_add.planAddLines.03", "no match")
		}
	}
	return []string{box, " " + theme.Paint(theme.RoleDim, tview.Escape(hint))}
}

// planPickLines is the choice when the name fits several items.
func planPickLines(sc planScreen) []string {
	var parts []string
	for i, f := range sc.picks {
		item := tview.Escape(f.label) + " (" + planKindWord(f.kind) + ")"
		if i == sc.pickSel {
			item = theme.Selected(" " + item + " ")
		}
		parts = append(parts, item)
		if i == 5 {
			break
		}
	}
	return []string{
		" " + theme.Paint(theme.RoleAccent, vibe.Line("ui.plan_add.planPickLines.01", "Which one? ")) + strings.Join(parts, "  "),
		" " + theme.Paint(theme.RoleDim, vibe.Line("ui.plan_add.planPickLines.02", "↑↓ choose   Enter add below   Esc cancel")),
	}
}
