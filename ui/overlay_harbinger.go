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
	"github.com/espresso20/ageforge/theme"
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
func (p *harbingerPanel) provider(state game.GameState, _ int) string {
	return harbingerPanelText(state, p.note, p.noteGood, p.inviteArmed)
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
		err, ok = engine.HarbingerAppease(), "Appeased. The real odds are lower now."
	case harbingerBrace:
		err, ok = engine.HarbingerBrace(), "Braced. An Endure will cost you less."
	case harbingerInvite:
		st := engine.GetState()
		if !p.inviteArmed {
			if st.Harbinger == nil || st.Harbinger.InviteBlocked != "" {
				err = harbingerInviteRefusal(st)
				break
			}
			p.inviteArmed = true
			p.note, p.noteGood = "Press I again to invite the catastrophe. This cannot be undone.", false
			return true
		}
		p.inviteArmed = false
		ok = "Invited. It will come when it was fated to."
		if st.Harbinger != nil && st.Harbinger.LastPassage {
			ok = "Invited. It will come at your next prestige."
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
		return fmt.Errorf("no harbinger is here to invite")
	}
	return fmt.Errorf("cannot invite: %s", st.Harbinger.InviteBlocked)
}

// harbingerPanelText renders the panel. Pure: state in, text out.
func harbingerPanelText(state game.GameState, note string, noteGood, inviteArmed bool) string {
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, "═══ Harbinger ═══") + "\n\n")

	h := state.Harbinger
	if h == nil {
		harbingerAbsentText(&sb, state)
	} else {
		harbingerPresentText(&sb, state, h, inviteArmed)
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
	sb.WriteString(" No harbinger is here.\n\n")
	// No era is named here but the current one: the player has not reached
	// the others (the no-spoiler rule, spoilers.go).
	sb.WriteString(theme.Paint(theme.RoleDim, " A harbinger comes only when doom is on its way, some while before it strikes,\n and the figure changes with each age it lives through. A quiet era is safe, for\n now. In the last era the doom is prestige itself: the Last Passage.") + "\n\n")

	o := state.CatastropheOutlook
	sb.WriteString(theme.Paint(theme.RoleAccent, "── Outlook ──") + "\n")
	switch {
	case o.Passage == game.PassagePrestige && o.Possible:
		fmt.Fprintf(sb, " Prestige here could bring the Last Passage. The risk is %s.\n", harbingerRiskWords(o.Tier))
		if harbingerNumericAge(state) {
			fmt.Fprintf(sb, " Published odds: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(o.Probability)))
		}
		sb.WriteString(theme.Paint(theme.RoleDim, " More faith in storage makes it less likely.") + "\n")
	case o.Passage == game.PassagePrestige && state.LastPassage.Pending:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "The Last Passage has come. Type 'catastrophe' to choose.") + "\n")
	case o.Passage == game.PassagePrestige:
		sb.WriteString(" This is the final epoch. Its passage is prestige, and no harbinger has come.\n")
	default:
		sb.WriteString(" " + eraOutlookText(state) + "\n")
		if o.Possible {
			sb.WriteString(theme.Paint(theme.RoleDim, " Faith in storage makes a doom less likely to strike, and faith and culture\n pay for Appease if a harbinger comes.") + "\n")
		}
		if r := eraDoomRecord(state); r != nil && r.Outcome == game.HarbingerOutcomeDiscredited {
			sb.WriteString(theme.Paint(theme.RoleDim, " "+capFirstUI(r.Name)+"'s warning in this era was invented.") + "\n")
		}
	}
}

// harbingerPresentText is the panel with a harbinger present.
func harbingerPresentText(sb *strings.Builder, state game.GameState, h *game.HarbingerView, inviteArmed bool) {
	fmt.Fprintf(sb, " %s   %s\n", theme.Paint(theme.RoleBright, capFirstUI(h.Name)),
		theme.Paint(theme.RoleDim, h.AgeName+" harbinger"))
	sb.WriteString(" " + theme.Paint(theme.RoleLabel, h.Description) + "\n")
	if len(h.Earlier) > 0 {
		names := make([]string, len(h.Earlier))
		for i, n := range h.Earlier {
			names[i] = capFirstUI(n)
		}
		sb.WriteString(" " + theme.Paint(theme.RoleDim, "Took up the warning from "+strings.Join(names, ", then ")+". Your answers stand.") + "\n")
	}
	sb.WriteString("\n")
	switch {
	case h.LastPassage:
		fmt.Fprintf(sb, " Warning of %s: the end of this civilization, when you next prestige.\n\n",
			theme.Paint(theme.RoleHighlight, h.TargetEpochName))
	case h.WhenText != "":
		// TargetEpochName is the warning ("impending doom"), never an era.
		fmt.Fprintf(sb, " Warning of %s %s.\n\n", theme.Paint(theme.RoleHighlight, h.TargetEpochName), h.WhenText)
	default:
		fmt.Fprintf(sb, " Warning of %s. %s gives no word of when.\n\n", theme.Paint(theme.RoleHighlight, h.TargetEpochName), capFirstUI(h.Name))
	}

	for _, l := range h.Lines {
		sb.WriteString(" " + theme.Paint(theme.RoleDim, "“"+l+"”") + "\n")
	}
	if h.LastPassageWaiting {
		sb.WriteString(" " + theme.Paint(theme.RoleDim, "The Last Passage still waits at your next prestige. Your answers to it stand.") + "\n")
	}
	sb.WriteString("\n")

	switch {
	case h.PassageCame:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "The Last Passage has come. Prestige waits: type 'catastrophe' to choose.") + "\n")
	case h.Numeric:
		fmt.Fprintf(sb, " Severity: %s\n", harbingerSeverityText(h.Tier))
		fmt.Fprintf(sb, " Published odds: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(h.Probability)))
	default:
		fmt.Fprintf(sb, " Severity: %s\n", harbingerSeverityText(h.Tier))
		sb.WriteString(theme.Paint(theme.RoleDim, " The omens give no figure. Their words are all you have to go on.") + "\n")
	}
	switch {
	case h.Invited && h.LastPassage && !h.PassageCame:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "You have invited it. The Last Passage will come when you prestige.") + "\n")
	case h.Invited && !h.LastPassage:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "You have invited it. The catastrophe will come when it was fated to.") + "\n")
	}

	sb.WriteString("\n" + theme.Paint(theme.RoleAccent, "── Answers ──") + "\n\n")

	// Appease.
	fmt.Fprintf(sb, " %s %s   %s\n", theme.Keycap("A"), theme.Paint(theme.RoleBright, "Appease: "+h.AppeaseLabel),
		harbingerLevelText(h.AppeaseLevel, game.HarbingerMaxAppease))
	sb.WriteString(theme.Paint(theme.RoleDim, "     Each level multiplies the real chance it strikes by 0.6.") + "\n")
	harbingerCostLine(sb, state, h.AppeaseBlocked, h.AppeaseCost)
	sb.WriteString("\n")

	// Brace.
	fmt.Fprintf(sb, " %s %s   %s\n", theme.Keycap("B"), theme.Paint(theme.RoleBright, "Brace: "+h.BraceLabel),
		harbingerLevelText(h.BraceLevel, game.HarbingerMaxBrace))
	if h.LastPassage {
		fmt.Fprintf(sb, theme.Paint(theme.RoleDim, "     If it comes and you Endure: you keep %d%% of the run's prestige points.")+"\n",
			h.EndurePointsPct)
		if h.BraceBlocked == "" {
			fmt.Fprintf(sb, theme.Paint(theme.RoleDim, "     Next level: %d%% kept.")+"\n", h.NextEndurePointsPct)
		}
	} else {
		fmt.Fprintf(sb, theme.Paint(theme.RoleDim, "     If it comes and you Endure: %d%% of buildings fall, %d%% of stock is kept.")+"\n",
			h.EndureDestroyPct, h.EndureKeepPct)
		if h.BraceBlocked == "" {
			fmt.Fprintf(sb, theme.Paint(theme.RoleDim, "     Next level: %d%% fall, %d%% kept.")+"\n", h.NextEndureDestroyPct, h.NextEndureKeepPct)
		}
		sb.WriteString(harbingerGarrisonLine(h) + "\n")
	}
	harbingerCostLine(sb, state, h.BraceBlocked, h.BraceCost)
	sb.WriteString("\n")

	// Invite.
	fmt.Fprintf(sb, " %s %s\n", theme.Keycap("I"), theme.Paint(theme.RoleBright, "Invite: "+h.InviteLabel))
	if h.LastPassage {
		sb.WriteString(theme.Paint(theme.RoleDim, "     Guarantees the Last Passage at your next prestige; Succumb then earns the\n     Cosmic Legacy. Free. Cannot be undone.") + "\n")
	} else {
		sb.WriteString(theme.Paint(theme.RoleDim, "     Guarantees the catastrophe; it still comes when it was fated to. Free.\n     Cannot be undone.") + "\n")
	}
	switch {
	case h.InviteBlocked != "" && h.Invited:
		sb.WriteString("     " + theme.Paint(theme.RoleDim, "Done: "+h.InviteBlocked+".") + "\n")
	case h.InviteBlocked != "":
		sb.WriteString("     " + theme.Paint(theme.RoleDim, "Unavailable: "+h.InviteBlocked+".") + "\n")
	case inviteArmed:
		sb.WriteString("     " + theme.Paint(theme.RoleNegative, "Press I again to confirm.") + "\n")
	}

	sb.WriteString("\n " + theme.KeycapButton("A", "Appease") + "  " + theme.KeycapButton("B", "Brace") + "  " +
		theme.KeycapButton("I", "Invite (twice)") + "  " + theme.KeycapButton("Esc", "Close") + "\n")
	sb.WriteString(theme.Paint(theme.RoleDim, " Nothing here expires. The price is the same in every age of the epoch.") + "\n")
}

// harbingerCostLine prints the next level's cost, each resource colored by
// whether you have it, or why the action is unavailable.
func harbingerCostLine(sb *strings.Builder, state game.GameState, blocked string, cost map[string]float64) {
	if blocked != "" {
		sb.WriteString("     " + theme.Paint(theme.RoleDim, "Unavailable: "+blocked+".") + "\n")
		return
	}
	if len(cost) == 0 {
		sb.WriteString("     " + theme.Paint(theme.RoleDim, "Nothing to pay with yet.") + "\n")
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
			parts = append(parts, theme.Paint(theme.RoleNegative, item+" (have "+FormatNumber(math.Floor(have))+")"))
		}
	}
	sb.WriteString("     Next level costs: " + strings.Join(parts, ", ") + "\n")
}

// harbingerLevelText renders "Level 1 / 2".
func harbingerLevelText(level, max int) string {
	return theme.Paint(theme.RoleLabel, fmt.Sprintf("Level %d / %d", level, max))
}

// harbingerSeverityText renders the vague severity with its color, in the
// same words as the outlook line (harbingerRiskWords).
func harbingerSeverityText(t game.CatastropheTier) string {
	switch t {
	case game.CatastropheTierHigh:
		return theme.Paint(theme.RoleNegative, "high") + ": the harbinger is in open terror"
	case game.CatastropheTierMedium:
		return theme.Paint(theme.RoleWarning, "medium") + ": a real danger, worth paying to lessen"
	case game.CatastropheTierLow:
		return theme.Paint(theme.RolePositive, "low") + ": uneasy rather than afraid"
	}
	return theme.Paint(theme.RoleDim, "none") + ": nothing to fear"
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

// capFirstUI upper-cases the first letter: "the Oracle" → "The Oracle".
func capFirstUI(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// harbingerGarrisonLine says whether the army is counted in the Brace
// preview above it, and by how much.
func harbingerGarrisonLine(h *game.HarbingerView) string {
	if h.GarrisonPct <= 0 {
		return theme.Paint(theme.RoleDim, "     No garrison counted: soldiers would soften an Endure further (Army panel).")
	}
	line := theme.Paint(theme.RolePositive, fmt.Sprintf("     Your garrison is counted: it blunts about %d%% of what Brace leaves.", h.GarrisonPct))
	if h.GarrisonCapped {
		line += "\n" + theme.Paint(theme.RoleDim, fmt.Sprintf("     Brace and garrison together soften an Endure by at most %.0f%%.", config.EndureReductionCap*100))
	}
	return line
}
