package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// epoch_card.go is the Discord card: the epoch screen's facts as plain text,
// about 60 columns wide, no color codes, fenced so that it pastes as a code
// block. `epoch card` writes it next to the saves.

// epochCardWidth is the card's width in columns.
const epochCardWidth = 60

// epochCardFile is the card's file name, in the account's data directory.
const epochCardFile = "epoch-card.txt"

var cardTagRe = regexp.MustCompile(`\[[a-zA-Z0-9#:\-]*\]`)

// cardPlain strips tview color tags and leaves ASCII-safe punctuation.
func cardPlain(s string) string {
	s = cardTagRe.ReplaceAllString(s, "")
	return strings.NewReplacer("[]", "[", "·", "-").Replace(s)
}

// epochCardText is the card for a state, fence included. It names no age the
// player has not seen (the ages ahead are counted).
func epochCardText(state game.GameState) string {
	v := newHallView(state)
	w := epochCardWidth
	rule := strings.Repeat("=", w)
	var lines []string
	add := func(ss ...string) { lines = append(lines, ss...) }
	wrapped := func(prefix, s string) {
		for i, l := range wrapText(s, w-len(prefix)) {
			if i > 0 {
				l = strings.Repeat(" ", len(prefix)) + l
			} else {
				l = prefix + l
			}
			add(l)
		}
	}
	heading := func(h string) { add("", h, strings.Repeat("-", len(h))) }

	add(rule)
	for _, l := range wrapText(strings.ToUpper(cardPlain(v.name)), w) {
		add(l)
	}
	where := []string{v.town, v.age}
	if v.year != "" {
		where = append(where, v.year)
	}
	for _, l := range wrapText(cardPlain(strings.Join(where, " - ")), w) {
		add(l)
	}
	add(rule)

	heading(vibe.Line("ui.epoch_card.epochCardText.01", "Highlights"))
	for _, s := range v.stats {
		add(fmt.Sprintf("%-14s %s", s.label, s.value))
	}

	heading(vibe.Line("ui.epoch_card.epochCardText.02", "Ages"))
	if len(v.done) > 0 {
		wrapped(vibe.Line("ui.epoch_card.epochCardText.03", "Done: "), strings.Join(v.done, ", "))
	}
	wrapped(vibe.Line("ui.epoch_card.epochCardText.04", "Now: "), v.cur)
	if v.unseen > 0 || len(v.ahead) > 0 {
		next := append([]string{}, v.ahead...)
		if v.unseen > 0 {
			next = append(next, vibe.Sprintf("ui.epoch_card.epochCardText.05", "%d more to come", v.unseen))
		}
		wrapped(vibe.Line("ui.epoch_card.epochCardText.06", "Ahead: "), strings.Join(next, ", "))
	}

	heading(vibe.Line("ui.epoch_card.epochCardText.07", "Top badges"))
	if len(v.badges) == 0 {
		add(vibe.Line("ui.epoch_card.epochCardText.08", "No badges earned yet."))
	}
	for _, b := range v.badges {
		wrapped("* ", b.Name)
	}

	heading(vibe.Line("ui.epoch_card.epochCardText.09", "Feats"))
	if len(v.feats) == 0 {
		add(vibe.Line("ui.epoch_card.epochCardText.10", "No feats yet."))
	}
	for _, f := range v.feats {
		wrapped("* ", cardPlain(f))
	}
	add("", rule)
	return "```\n" + strings.Join(lines, "\n") + "\n```\n"
}

// cmdEpoch opens the epoch screen, or with `card` writes the Discord card.
func cmdEpoch(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return panelReply("epoch")
	}
	if strings.ToLower(args[0]) != "card" || len(args) > 1 {
		return infoReply(vibe.Line("ui.epoch_card.cmdEpoch.01", "Usage: epoch (opens the Epoch screen), or epoch card (saves a card to paste into Discord)"))
	}
	dir := game.DataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errorReply(vibe.Line("ui.epoch_card.cmdEpoch.02", "Could not create the data folder: ") + shortIOError(err) + ".")
	}
	path := filepath.Join(dir, epochCardFile)
	if err := os.WriteFile(path, []byte(epochCardText(engine.GetState())), 0644); err != nil {
		return errorReply(vibe.Line("ui.epoch_card.cmdEpoch.03", "Could not save the card: ") + shortIOError(err) + ".")
	}
	return CommandResult{Message: vibe.Sprintf("ui.epoch_card.cmdEpoch.04", "Saved your epoch card to %s. Paste its contents into Discord.", path), Type: "success"}
}
