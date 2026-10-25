package ui

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// The Harbinger panel: who came, what they said, how worried to be, and the
// three answers (A / B / I). A text overlay with a key handler; the handler
// runs on the tview goroutine and never under the engine lock, so calling the
// engine's action methods (write lock) and GetState (read lock) from it is safe.

// harbingerPanel holds the panel's UI-only state. Owned by the tview
// goroutine: touched only from the key handler, the provider (called by
// OverlayManager.Show/Refresh) and the command path that opens the panel.
type harbingerPanel struct {
	// note is the last action's feedback line ("" for none).
	note     string
	noteGood bool
	// inviteArmed is set by the first I press; the second one invites. Invite
	// cannot be undone, so one stray key must not do it.
	inviteArmed bool
}

// reset clears the feedback and the Invite confirmation, for a fresh open.
func (p *harbingerPanel) reset() {
	p.note, p.noteGood, p.inviteArmed = "", false, false
}

// provider renders the panel for OverlayManager.
func (p *harbingerPanel) provider(state game.GameState, screenW int) string {
	return harbingerPanelTextAt(state, screenW, p.note, p.noteGood, p.inviteArmed)
}

// harbingerAction is one of the panel's three answers.
type harbingerAction int

const (
	harbingerNoAction harbingerAction = iota
	harbingerAppease
	harbingerBrace
	harbingerInvite
)

// harbingerKeyAction maps a key event to an action (A / B / I, any case).
func harbingerKeyAction(event *tcell.EventKey) harbingerAction {
	if event.Key() != tcell.KeyRune {
		return harbingerNoAction
	}
	switch unicode.ToLower(event.Rune()) {
	case 'a':
		return harbingerAppease
	case 'b':
		return harbingerBrace
	case 'i':
		return harbingerInvite
	}
	return harbingerNoAction
}

// handleKey runs the action for event against engine and updates the note.
// Reports whether the key was consumed. Takes engine locks through the public
// action methods only; must not be called under the engine lock.
func (p *harbingerPanel) handleKey(event *tcell.EventKey, engine *game.GameEngine) bool {
	act := harbingerKeyAction(event)
	if act == harbingerNoAction {
		return false
	}
	if act != harbingerInvite {
		p.inviteArmed = false
	}
	var err error
	var ok string
	switch act {
	case harbingerAppease:
		err, ok = engine.HarbingerAppease(), vibe.Line("ui.overlay_harbinger.handleKey.01", "Appeased. The real odds are lower now.")
	case harbingerBrace:
		err, ok = engine.HarbingerBrace(), vibe.Line("ui.overlay_harbinger.handleKey.02", "Braced. An Endure will cost you less.")
	case harbingerInvite:
		st := engine.GetState()
		if !p.inviteArmed {
			if st.Harbinger == nil || st.Harbinger.InviteBlocked != "" {
				err = harbingerInviteRefusal(st)
				break
			}
			p.inviteArmed = true
			p.note, p.noteGood = vibe.Line("ui.overlay_harbinger.handleKey.03", "Press I again to invite the catastrophe. This cannot be undone."), false
			return true
		}
		p.inviteArmed = false
		ok = vibe.Line("ui.overlay_harbinger.handleKey.04", "Invited. It will come when it was fated to.")
		if st.Harbinger != nil && st.Harbinger.LastPassage {
			ok = vibe.Line("ui.overlay_harbinger.handleKey.05", "Invited. It will come at your next prestige.")
		}
		err = engine.HarbingerInvite()
	}
	if err != nil {
		p.note, p.noteGood = err.Error(), false
	} else {
		p.note, p.noteGood = ok, true
	}
	return true
}

// harbingerInviteRefusal is the error for an I press that cannot invite.
func harbingerInviteRefusal(st game.GameState) error {
	if st.Harbinger == nil {
		return vibe.Errorf("ui.overlay_harbinger.harbingerInviteRefusal.01", "no harbinger is here to invite")
	}
	return vibe.Errorf("ui.overlay_harbinger.harbingerInviteRefusal.02", "cannot invite: %s", st.Harbinger.InviteBlocked)
}

// harbingerPanelText renders the panel for the narrowest screen the game
// draws. Pure: state in, text out.
func harbingerPanelText(state game.GameState, note string, noteGood, inviteArmed bool) string {
	return harbingerPanelTextAt(state, narrowestScreenW, note, noteGood, inviteArmed)
}

// narrowestScreenW is the narrowest terminal the game is laid out for.
const narrowestScreenW = 80

// harbingerPanelTextAt renders the panel for a screen screenW cells wide, so
// the lines that can run long (the prices) are broken to fit the box.
func harbingerPanelTextAt(state game.GameState, screenW int, note string, noteGood, inviteArmed bool) string {
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, vibe.Line("ui.overlay_harbinger.harbingerPanelTextAt.01", "═══ Harbinger ═══")) + "\n\n")

	h := state.Harbinger
	if h == nil {
		harbingerAbsentText(&sb, state)
	} else {
		harbingerPresentText(&sb, state, h, overlayTextWidth(screenW), inviteArmed)
	}

	if note != "" {
		role := theme.RoleNegative
		if noteGood {
			role = theme.RolePositive
		}
		sb.WriteString("\n " + theme.Paint(role, tview.Escape(note)) + "\n")
	}
	return sb.String()
}

// harbingerAbsentText is the panel with nobody there: how harbingers come,
// and the outlook in plain words. It reads only what the player has seen, so
// an era with a doom fated and a quiet one look the same until one comes.
func harbingerAbsentText(sb *strings.Builder, state game.GameState) {
	sb.WriteString(vibe.Line("ui.overlay_harbinger.harbingerAbsentText.01", " No harbinger is here.\n\n"))
	// No era is named here but the current one: the player has not reached
	// the others (the no-spoiler rule, spoilers.go).
	sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerAbsentText.02", " A harbinger comes only when doom is on its way, some while before it strikes,\n and the figure changes with each age it lives through. A quiet era is safe, for\n now. In the last era the doom is prestige itself: the Last Passage.")) + "\n\n")

	o := state.CatastropheOutlook
	sb.WriteString(theme.Paint(theme.RoleAccent, vibe.Line("ui.overlay_harbinger.harbingerAbsentText.03", "── Outlook ──")) + "\n")
	switch {
	case o.Passage == game.PassagePrestige && o.Possible:
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerAbsentText.04", " Prestige here could bring the Last Passage. The risk is %s.\n", harbingerRiskWords(o.Tier)))
		if harbingerNumericAge(state) {
			fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerAbsentText.05", " Published odds: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(o.Probability))))
		}
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerAbsentText.06", " Faith strength: ")+faithStrengthFigure(state)+vibe.Line("ui.overlay_harbinger.harbingerAbsentText.07", ". Stronger faith makes it less likely.")) + "\n")
	case o.Passage == game.PassagePrestige && state.LastPassage.Pending:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, vibe.Line("ui.overlay_harbinger.harbingerAbsentText.08", "The Last Passage has come. Type 'catastrophe' to choose.")) + "\n")
	case o.Passage == game.PassagePrestige:
		sb.WriteString(vibe.Line("ui.overlay_harbinger.harbingerAbsentText.09", " This is the final epoch. Its passage is prestige, and no harbinger has come.\n"))
	default:
		sb.WriteString(" " + eraOutlookText(state) + "\n")
		if o.Possible {
			sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerAbsentText.10", " Faith strength: ")+faithStrengthFigure(state)+vibe.Line("ui.overlay_harbinger.harbingerAbsentText.11", ".\n Stronger faith makes a doom less likely to strike, and faith and culture pay\n for Appease if a harbinger comes.")) + "\n")
		}
		if r := eraDoomRecord(state); r != nil && r.Outcome == game.HarbingerOutcomeDiscredited {
			sb.WriteString(theme.Paint(theme.RoleDim, " "+textfmt.Capitalize(r.Name)+vibe.Line("ui.overlay_harbinger.harbingerAbsentText.12", "'s warning in this era was invented.")) + "\n")
		}
	}
}

// harbingerPresentText is the panel with a harbinger present.
func harbingerPresentText(sb *strings.Builder, state game.GameState, h *game.HarbingerView, width int, inviteArmed bool) {
	fmt.Fprintf(sb, " %s   %s\n", theme.Paint(theme.RoleBright, textfmt.Capitalize(h.Name)),
		theme.Paint(theme.RoleDim, h.AgeName+vibe.Line("ui.overlay_harbinger.harbingerPresentText.01", " harbinger")))
	sb.WriteString(" " + theme.Paint(theme.RoleLabel, h.Description) + "\n")
	if len(h.Earlier) > 0 {
		names := make([]string, len(h.Earlier))
		for i, n := range h.Earlier {
			names[i] = textfmt.Capitalize(n)
		}
		sb.WriteString(" " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.02", "Took up the warning from ")+strings.Join(names, ", then ")+vibe.Line("ui.overlay_harbinger.harbingerPresentText.03", ". Your answers stand.")) + "\n")
	}
	sb.WriteString("\n")
	switch {
	case h.LastPassage:
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerPresentText.04", " Warning of %s: the end of this civilization, when you next prestige.\n\n",
			theme.Paint(theme.RoleHighlight, h.TargetEpochName)))
	case h.WhenText != "":
		// TargetEpochName is the warning ("impending doom"), never an era.
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerPresentText.05", " Warning of %s %s.\n\n", theme.Paint(theme.RoleHighlight, h.TargetEpochName), h.WhenText))
	default:
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerPresentText.06", " Warning of %s. %s gives no word of when.\n\n", theme.Paint(theme.RoleHighlight, h.TargetEpochName), textfmt.Capitalize(h.Name)))
	}

	for _, l := range h.Lines {
		sb.WriteString(" " + theme.Paint(theme.RoleComment, "“"+l+"”") + "\n")
	}
	if h.LastPassageWaiting {
		sb.WriteString(" " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.07", "The Last Passage still waits at your next prestige. Your answers to it stand.")) + "\n")
	}
	sb.WriteString("\n")

	switch {
	case h.PassageCame:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, vibe.Line("ui.overlay_harbinger.harbingerPresentText.08", "The Last Passage has come. Prestige waits: type 'catastrophe' to choose.")) + "\n")
	case h.Numeric:
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerPresentText.09", " Severity: %s\n", harbingerSeverityText(h.Tier)))
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerPresentText.10", " Published odds: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(h.Probability))))
	default:
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_harbinger.harbingerPresentText.11", " Severity: %s\n", harbingerSeverityText(h.Tier)))
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.12", " The omens give no figure. Their words are all you have to go on.")) + "\n")
	}
	if !h.PassageCame {
		// The strength the odds are read from, as the faith row shows it.
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.13", " Faith strength: ")+faithStrengthFigure(state)+vibe.Line("ui.overlay_harbinger.harbingerPresentText.14", ". Stronger faith lowers the odds.")) + "\n")
	}
	switch {
	case h.Invited && h.LastPassage && !h.PassageCame:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, vibe.Line("ui.overlay_harbinger.harbingerPresentText.15", "You have invited it. The Last Passage will come when you prestige.")) + "\n")
	case h.Invited && !h.LastPassage:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, vibe.Line("ui.overlay_harbinger.harbingerPresentText.16", "You have invited it. The catastrophe will come when it was fated to.")) + "\n")
	}

	sb.WriteString("\n" + theme.Paint(theme.RoleAccent, vibe.Line("ui.overlay_harbinger.harbingerPresentText.17", "── Answers ──")) + "\n\n")

	// Appease.
	fmt.Fprintf(sb, " %s %s   %s\n", theme.Keycap("A"), theme.Paint(theme.RoleBright, vibe.Line("ui.overlay_harbinger.harbingerPresentText.18", "Appease: ")+h.AppeaseLabel),
		harbingerLevelText(h.AppeaseLevel, game.HarbingerMaxAppease))
	sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.19", "     Each level multiplies the real chance it strikes by 0.6.")) + "\n")
	harbingerCostLine(sb, state, width, h.AppeaseBlocked, h.AppeaseCost)
	sb.WriteString("\n")

	// Brace.
	fmt.Fprintf(sb, " %s %s   %s\n", theme.Keycap("B"), theme.Paint(theme.RoleBright, vibe.Line("ui.overlay_harbinger.harbingerPresentText.20", "Brace: ")+h.BraceLabel),
		harbingerLevelText(h.BraceLevel, game.HarbingerMaxBrace))
	if h.LastPassage {
		fmt.Fprintf(sb, theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.21", "     If it comes and you Endure: you keep %d%% of the run's prestige points."))+"\n",
			h.EndurePointsPct)
		if h.BraceBlocked == "" {
			fmt.Fprintf(sb, theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.22", "     Next level: %d%% kept."))+"\n", h.NextEndurePointsPct)
		}
	} else {
		fmt.Fprintf(sb, theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.23", "     If it comes and you Endure: %d%% of buildings fall, %d%% of stock is kept."))+"\n",
			h.EndureDestroyPct, h.EndureKeepPct)
		if h.BraceBlocked == "" {
			fmt.Fprintf(sb, theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.24", "     Next level: %d%% fall, %d%% kept."))+"\n", h.NextEndureDestroyPct, h.NextEndureKeepPct)
		}
		sb.WriteString(harbingerGarrisonLine(h) + "\n")
	}
	harbingerCostLine(sb, state, width, h.BraceBlocked, h.BraceCost)
	sb.WriteString("\n")

	// Invite.
	fmt.Fprintf(sb, " %s %s\n", theme.Keycap("I"), theme.Paint(theme.RoleBright, vibe.Line("ui.overlay_harbinger.harbingerPresentText.25", "Invite: ")+h.InviteLabel))
	if h.LastPassage {
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.26", "     Guarantees the Last Passage at your next prestige; Succumb then earns the\n     Cosmic Legacy. Free. Cannot be undone.")) + "\n")
	} else {
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.27", "     Guarantees the catastrophe; it still comes when it was fated to. Free.\n     Cannot be undone.")) + "\n")
	}
	switch {
	case h.InviteBlocked != "" && h.Invited:
		sb.WriteString("     " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.28", "Done: ")+h.InviteBlocked+".") + "\n")
	case h.InviteBlocked != "":
		sb.WriteString("     " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.29", "Unavailable: ")+h.InviteBlocked+".") + "\n")
	case inviteArmed:
		sb.WriteString("     " + theme.Paint(theme.RoleNegative, vibe.Line("ui.overlay_harbinger.harbingerPresentText.30", "Press I again to confirm.")) + "\n")
	}

	sb.WriteString("\n " + theme.KeycapButton("A", vibe.Line("ui.overlay_harbinger.harbingerPresentText.31", "Appease")) + "  " + theme.KeycapButton("B", vibe.Line("ui.overlay_harbinger.harbingerPresentText.32", "Brace")) + "  " +
		theme.KeycapButton("I", vibe.Line("ui.overlay_harbinger.harbingerPresentText.33", "Invite (twice)")) + "  " + theme.KeycapButton("Esc", vibe.Line("ui.overlay_harbinger.harbingerPresentText.35", "Close")) + "\n")
	sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerPresentText.36", " Nothing here expires. The price is the same in every age of the epoch.")) + "\n")
}

// harbingerCostLine prints the next level's cost, each resource colored by
// whether you have it, or why the action is unavailable. A price in any
// number of materials fits a box width cells wide: when the line would run
// past it the materials continue on the next line, under the first.
func harbingerCostLine(sb *strings.Builder, state game.GameState, width int, blocked string, cost map[string]float64) {
	if blocked != "" {
		sb.WriteString("     " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerCostLine.01", "Unavailable: ")+blocked+".") + "\n")
		return
	}
	if len(cost) == 0 {
		sb.WriteString("     " + theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerCostLine.02", "Nothing to pay with yet.")) + "\n")
		return
	}
	var parts []string
	for _, def := range state.Ruleset().Resources() {
		need, ok := cost[def.Key]
		if !ok {
			continue
		}
		have := state.Resources[def.Key].Amount
		item := game.Amount(need, def.Key)
		if have >= need {
			parts = append(parts, theme.Paint(theme.RolePositive, item))
		} else {
			parts = append(parts, theme.Paint(theme.RoleNegative, item+vibe.Line("ui.overlay_harbinger.harbingerCostLine.03", " (have ")+textfmt.Number(math.Floor(have))+")"))
		}
	}
	lead := vibe.Line("ui.overlay_harbinger.harbingerCostLine.04", "     Next level costs: ")
	indent := strings.Repeat(" ", runeLen(lead))
	line, lineW := lead, runeLen(lead)
	for i, part := range parts {
		w := visibleLen(part)
		if i > 0 {
			if lineW+1+w > width && lineW > len(indent) {
				sb.WriteString(strings.TrimRight(line, " ") + "\n")
				line, lineW = indent, len(indent)
			} else {
				line, lineW = line+" ", lineW+1
			}
		}
		line, lineW = line+part, lineW+w
		if i < len(parts)-1 {
			line, lineW = line+",", lineW+1
		}
	}
	sb.WriteString(line + "\n")
}

// harbingerLevelText renders "Level 1 / 2".
func harbingerLevelText(level, max int) string {
	return theme.Paint(theme.RoleLabel, vibe.Sprintf("ui.overlay_harbinger.harbingerLevelText.01", "Level %d / %d", level, max))
}

// harbingerSeverityText renders the vague severity with its color, in the
// same words as the outlook line (harbingerRiskWords).
func harbingerSeverityText(t game.CatastropheTier) string {
	switch t {
	case game.CatastropheTierHigh:
		return theme.Paint(theme.RoleNegative, "high") + vibe.Line("ui.overlay_harbinger.harbingerSeverityText.02", ": the harbinger is in open terror")
	case game.CatastropheTierMedium:
		return theme.Paint(theme.RoleWarning, "medium") + vibe.Line("ui.overlay_harbinger.harbingerSeverityText.04", ": a real danger, worth paying to lessen")
	case game.CatastropheTierLow:
		return theme.Paint(theme.RolePositive, "low") + vibe.Line("ui.overlay_harbinger.harbingerSeverityText.06", ": uneasy rather than afraid")
	}
	return theme.Paint(theme.RoleDim, "none") + vibe.Line("ui.overlay_harbinger.harbingerSeverityText.08", ": nothing to fear")
}

// harbingerRiskWords is the tier in plain words for the outlook line.
func harbingerRiskWords(t game.CatastropheTier) string {
	switch t {
	case game.CatastropheTierHigh:
		return theme.Paint(theme.RoleNegative, "high")
	case game.CatastropheTierMedium:
		return theme.Paint(theme.RoleWarning, "medium")
	case game.CatastropheTierLow:
		return theme.Paint(theme.RolePositive, "low")
	}
	return "none"
}

// harbingerNumericAge reports whether the forecasts of the snapshot's age
// print the odds.
func harbingerNumericAge(state game.GameState) bool {
	def, ok := state.Ruleset().Harbinger(state.Age)
	return ok && def.ForecastPrecision == config.ForecastNumeric
}

// harbingerPercent renders a probability as a whole percentage ("14%").
func harbingerPercent(p float64) string {
	return fmt.Sprintf("%.0f%%", p*100)
}

// harbingerGarrisonLine says whether the army is counted in the Brace
// preview above it, and by how much.
func harbingerGarrisonLine(h *game.HarbingerView) string {
	if h.GarrisonPct <= 0 {
		return theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_harbinger.harbingerGarrisonLine.01", "     No garrison counted: soldiers would soften an Endure further (Army panel)."))
	}
	line := theme.Paint(theme.RolePositive, vibe.Sprintf("ui.overlay_harbinger.harbingerGarrisonLine.02", "     Your garrison is counted: it blunts about %d%% of what Brace leaves.", h.GarrisonPct))
	if h.GarrisonCapped {
		line += "\n" + theme.Paint(theme.RoleDim, vibe.Sprintf("ui.overlay_harbinger.harbingerGarrisonLine.03", "     Brace and garrison together soften an Endure by at most %.0f%%.", config.EndureReductionCap*100))
	}
	return line
}
