package main

import (
	"regexp"
	"strings"
	"unicode"
)

// vibe_screens.go: what vibe-wire decides about the screens-and-help sheet
// that the syntax alone does not say.
//
// Most of the sheet is decided like the log: a text in a function is wired,
// and a text in a table made once becomes an entry that carries its id
// (vibe.E). What is judged here is the text that is not words at all, or that
// the player types, or that is another program's.

const (
	leftKeycap   = "keycap: the name of a key on the keyboard, which reads the same in any script"
	leftNoWords  = "no words: nothing but marks, numbers and format verbs"
	leftCommand  = "command: syntax the player types, which stays exactly as it is"
	leftOtherApp = "other program: a step that names another program's menus, settings and config lines, which must read as that program shows them"
	leftBoot     = "boot: printed before any account's settings are read, so VibeMode is always off"
	leftArticle  = "article: put in front of a name unless the name begins with it, and tested against the names of the ruleset"
)

func init() {
	leftAloneFor(leftDev,
		// The file `dump` writes for a bug report (the screen's own lines, 01, 17 and 18, are wired).
		"ui.input.cmdDump.02", "ui.input.cmdDump.03", "ui.input.cmdDump.04", "ui.input.cmdDump.05",
		"ui.input.cmdDump.06", "ui.input.cmdDump.07", "ui.input.cmdDump.08", "ui.input.cmdDump.09",
		"ui.input.cmdDump.10", "ui.input.cmdDump.11", "ui.input.cmdDump.12", "ui.input.cmdDump.13",
		"ui.input.cmdDump.14", "ui.input.cmdDump.15", "ui.input.cmdDump.16",
		"ui.dashboard.submitInput.01", // the echo of a developer's command
	)
	leftAloneFor(leftCommand,
		"ui.input.cmdTrade.01",                                                 // "trade " and the word the player typed after it
		"ui.input.wonderBankUsage.01",                                          // Usage: wonder collect|bank <resource|all> [amount|all|max]
		"ui.input.cmdTradeRoute.01",                                            // Usage: trade route start|stop <route>
		"ui.overlay_plan.planPanelText.04", "ui.overlay_plan.planPanelText.05", // plan build <building> [count], and an example of it
		"ui.people_tabs.work.05", "ui.people_tabs.work.15", "ui.people_tabs.work.16", "ui.people_tabs.work.20", // roster <area> <%|auto>
		"ui.overlay_help.helpProvider.09", // the key name Tab in the help page (the word help beside it is not a row)
	)
	leftAloneFor(leftKey,
		"ui.research_tree.laneLabel.01", // AEIOU: the vowels a letter is tested against
		// The names of the Help panel's sections, which a command's Section is
		// compared with; the headings the player reads are the entries beside them.
		"ui.commands.secActions.02", "ui.commands.secPlan.02", "ui.commands.secPeople.02",
		"ui.commands.secResearch.02", "ui.commands.secTrade.02", "ui.commands.secWonders.02",
		"ui.commands.secGame.02", "ui.commands.secAccounts.02",
	)
	leftAloneFor(leftBoot, "main.main.main.01", "main.main.main.02", "main.main.leave.01")
	leftAloneFor(leftArticle, "ui.overlay_stats.statsProvider.42")
}

// wireDespiteTwins are texts that link marks as compared because another
// literal of the same words sits in a text operation, but that are not a
// comparison.
var wireDespiteTwins = map[string]bool{
	"ui.research_tree.canvas.01": true, // " or " between two lines of the tree; the other " or " is a Join separator
}

// keycapWords are the words a key hint may be made of and still be only the
// name of a key.
var keycapWords = map[string]bool{
	"enter": true, "esc": true, "tab": true, "shift": true, "arrows": true, "up": true, "down": true,
	"left": true, "right": true, "pgup": true, "pgdn": true, "home": true, "end": true, "backspace": true,
	"space": true, "delete": true, "option": true, "alt": true,
	"◄►": true, "↑↓": true, // the marks the map's key hints use for the arrow keys
}

var keycapSplit = regexp.MustCompile(`[\s/\-+]+`)
var styleTagRe = regexp.MustCompile(`\[[A-Za-z0-9#:\-]*\]`)

// screenReason says why a row of the screens sheet is not wired, when its text
// says so by itself: the name of a key, or nothing a script could write.
func screenReason(r *row) string {
	if r.area != areaScreens && r.area != areaMap {
		return "" // the map's key hints are named like the screens' are
	}
	// How to pick a font in each terminal: the names of their menus and the lines of their config files.
	if strings.HasPrefix(r.id, "ui.icons.terminalFontSteps.") {
		return leftOtherApp
	}
	text := styleTagRe.ReplaceAllString(r.current, "")
	words := keycapSplit.Split(strings.TrimSpace(text), -1)
	if len(words) > 0 && words[0] != "" {
		key, ctrl := true, false
		for _, w := range words {
			lw := strings.ToLower(w)
			switch {
			case lw == "ctrl":
				ctrl = true
			case keycapWords[lw]:
				ctrl = false
			case ctrl && len(lw) == 1: // Ctrl+R
				ctrl = false
			default:
				key = false
			}
		}
		if key {
			return leftKeycap
		}
	}
	rest := verbRe.ReplaceAllString(text, "")
	for _, ch := range rest {
		if unicode.IsLetter(ch) {
			return ""
		}
	}
	return leftNoWords
}

// tableEntry says whether a text sits in a table made once: a list or a
// struct of a package-level var, or the var itself.
func tableEntry(near string) bool {
	switch near {
	case "elem", "field", "var":
		return true
	}
	return false
}
