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
		if !p.inviteArmed {
			if st := engine.GetState(); st.Harbinger == nil || st.Harbinger.InviteBlocked != "" {
				err = harbingerInviteRefusal(st)
				break
			}
			p.inviteArmed = true
			p.note, p.noteGood = "Press I again to invite the catastrophe. This cannot be undone.", false
			return true
		}
		p.inviteArmed = false
		err, ok = engine.HarbingerInvite(), "Invited. It will come at the passage."
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

// harbingerAbsentText is the panel with nobody there: when one comes, and the
// next transition's outlook in plain words.
func harbingerAbsentText(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" No harbinger is here.\n\n")
	sb.WriteString(theme.Paint(theme.RoleDim, " Harbingers walk through every epoch whose passage into the next could bring a\n catastrophe, one figure per age, from its first age until that passage. In the\n Cosmic Era the passage is prestige itself: the Last Passage.") + "\n\n")

	o := state.CatastropheOutlook
	sb.WriteString(theme.Paint(theme.RoleAccent, "── Outlook ──") + "\n")
	switch {
	case o.Passage == game.PassagePrestige && o.Possible:
		fmt.Fprintf(sb, " Prestige here could bring the Last Passage. The risk is %s.\n", harbingerRiskWords(o.Tier))
		if harbingerNumericAge(state.Age) {
			fmt.Fprintf(sb, " Published odds: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(o.Probability)))
		}
		sb.WriteString(theme.Paint(theme.RoleDim, " More faith in storage makes it less likely.") + "\n")
	case o.Passage == game.PassagePrestige && state.LastPassage.Pending:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "The Last Passage has come. Type 'catastrophe' to choose.") + "\n")
	case o.NextEpochKey == "":
		sb.WriteString(" This is the final epoch. Its passage is prestige, and no harbinger has come.\n")
	case !o.Possible:
		fmt.Fprintf(sb, " The passage into the %s cannot bring a catastrophe.\n", config.EpochByKey()[o.NextEpochKey].Name)
	default:
		fmt.Fprintf(sb, " The passage into the %s could bring a catastrophe. The risk is %s.\n",
			config.EpochByKey()[o.NextEpochKey].Name, harbingerRiskWords(o.Tier))
		if harbingerNumericAge(state.Age) {
			fmt.Fprintf(sb, " Published odds: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(o.Probability)))
		}
		sb.WriteString(theme.Paint(theme.RoleDim, " More faith in storage makes it less likely.") + "\n")
	}
}

// harbingerPresentText is the panel with a harbinger present.
func harbingerPresentText(sb *strings.Builder, state game.GameState, h *game.HarbingerView, inviteArmed bool) {
	fmt.Fprintf(sb, " %s   %s\n", theme.Paint(theme.RoleBright, strings.ToUpper(capFirstUI(h.Name))),
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
	if h.LastPassage {
		fmt.Fprintf(sb, " Warning of %s: the end of this civilization, when you next prestige.\n\n",
			theme.Paint(theme.RoleHighlight, h.TargetEpochName))
	} else {
		fmt.Fprintf(sb, " Warning of the passage into the %s.\n\n", theme.Paint(theme.RoleHighlight, h.TargetEpochName))
	}

	for _, l := range h.Lines {
		sb.WriteString(" " + theme.Paint(theme.RoleDim, "“"+l+"”") + "\n")
	}
	sb.WriteString("\n")

	switch {
	case h.PassageCame:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "The Last Passage has come. Prestige waits: type 'catastrophe' to choose.") + "\n")
	case h.Numeric:
		fmt.Fprintf(sb, " Severity: %s\n", harbingerSeverityText(h.Tier))
		fmt.Fprintf(sb, " Odds published: %s\n", theme.Paint(theme.RoleHighlight, harbingerPercent(h.Probability)))
	default:
		fmt.Fprintf(sb, " Severity: %s\n", harbingerSeverityText(h.Tier))
		sb.WriteString(theme.Paint(theme.RoleDim, " The omens give no figure. Their words are all you have to go on.") + "\n")
	}
	switch {
	case h.Invited && h.LastPassage && !h.PassageCame:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "You have invited it. The Last Passage will come when you prestige.") + "\n")
	case h.Invited && !h.LastPassage:
		sb.WriteString(" " + theme.Paint(theme.RoleNegative, "You have invited it. The catastrophe will come at the passage.") + "\n")
	}

	sb.WriteString("\n" + theme.Paint(theme.RoleAccent, "── Answers ──") + "\n\n")

	// Appease.
	fmt.Fprintf(sb, " %s %s   %s\n", theme.Keycap("A"), theme.Paint(theme.RoleBright, "Appease — "+h.AppeaseLabel),
		harbingerLevelText(h.AppeaseLevel, game.HarbingerMaxAppease))
	sb.WriteString(theme.Paint(theme.RoleDim, "     Each level multiplies the real catastrophe chance by 0.6.") + "\n")
	harbingerCostLine(sb, state, h.AppeaseBlocked, h.AppeaseCost)
	sb.WriteString("\n")

	// Brace.
	fmt.Fprintf(sb, " %s %s   %s\n", theme.Keycap("B"), theme.Paint(theme.RoleBright, "Brace — "+h.BraceLabel),
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
	}
	harbingerCostLine(sb, state, h.BraceBlocked, h.BraceCost)
	sb.WriteString("\n")

	// Invite.
	fmt.Fprintf(sb, " %s %s\n", theme.Keycap("I"), theme.Paint(theme.RoleBright, "Invite — "+h.InviteLabel))
	if h.LastPassage {
		sb.WriteString(theme.Paint(theme.RoleDim, "     Guarantees the Last Passage at your next prestige; Succumb then earns the\n     Cosmic Legacy. Free. Cannot be undone.") + "\n")
	} else {
		sb.WriteString(theme.Paint(theme.RoleDim, "     Guarantees the catastrophe at this passage. Free. Cannot be undone.") + "\n")
	}
	switch {
	case h.InviteBlocked != "":
		sb.WriteString("     " + theme.Paint(theme.RoleDim, "Done: "+h.InviteBlocked+".") + "\n")
	case inviteArmed:
		sb.WriteString("     " + theme.Paint(theme.RoleNegative, "Press I again to confirm.") + "\n")
	}

	sb.WriteString("\n " + theme.KeycapButton("A", "Appease") + "  " + theme.KeycapButton("B", "Brace") + "  " +
		theme.KeycapButton("I", "Invite (twice)") + "  " + theme.KeycapButton("Esc", "Close") + "\n")
	sb.WriteString(theme.Paint(theme.RoleDim, " Nothing here expires. The price is the same in every age of the epoch.") + "\n")
}

// harbingerCostLine prints the next level's cost, each resource coloured by
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
	for _, def := range config.BaseResources() {
		need, ok := cost[def.Key]
		if !ok {
			continue
		}
		have := state.Resources[def.Key].Amount
		item := fmt.Sprintf("%s %s", FormatNumber(need), def.Key)
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

// harbingerSeverityText renders the vague severity with its colour.
func harbingerSeverityText(t game.CatastropheTier) string {
	switch t {
	case game.CatastropheTierHigh:
		return theme.Paint(theme.RoleNegative, "HIGH") + " — the harbinger is in open terror"
	case game.CatastropheTierMedium:
		return theme.Paint(theme.RoleWarning, "MEDIUM") + " — a real danger, worth paying to lessen"
	case game.CatastropheTierLow:
		return theme.Paint(theme.RolePositive, "LOW") + " — uneasy rather than afraid"
	}
	return theme.Paint(theme.RoleDim, "NONE") + " — nothing to fear"
}

// harbingerRiskWords is the tier in plain words for the outlook line.
func harbingerRiskWords(t game.CatastropheTier) string {
	switch t {
	case game.CatastropheTierHigh:
		return theme.Paint(theme.RoleNegative, "high")
	case game.CatastropheTierMedium:
		return theme.Paint(theme.RoleWarning, "moderate")
	case game.CatastropheTierLow:
		return theme.Paint(theme.RolePositive, "low")
	}
	return "nil"
}

// harbingerNumericAge reports whether age's forecasts print the odds.
func harbingerNumericAge(age string) bool {
	def, ok := config.HarbingerFor(age)
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
