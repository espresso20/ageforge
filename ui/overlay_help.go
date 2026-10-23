package ui

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// helpProvider renders the Help overlay from the command registry
// (commands.go): each section's usage rows, the panels the player can open,
// the prompt's keys and the shortcuts. It is intentionally static (it does
// not read game state) so the same reference is available at any point in
// play.
func helpProvider(_ game.GameState, screenW int) string {
	var sb strings.Builder
	reg := registry()
	// Every row is a key or a command, then what it does: the description
	// wraps under itself, so the two columns hold at any width.
	width := overlayTextWidth(screenW)
	if screenW <= 0 {
		width = 1 << 20 // no screen to fit: nothing wraps
	}
	row := func(lead, text string) { sb.WriteString(hangingRow(lead, text, width)) }
	// note writes a line of prose in gray, broken at its spaces.
	note := func(text string) {
		for _, line := range wrapWords(text, width) {
			sb.WriteString("[gray]" + line + "[-]\n")
		}
	}

	// keys writes a section of keys: each key padded to the longest, then
	// what it does.
	keys := func(rows ...[2]string) {
		keyW := 0
		for _, r := range rows {
			keyW = max(keyW, runeLen(r[0]))
		}
		for _, r := range rows {
			row("  [cyan]"+r[0]+"[-]"+strings.Repeat(" ", keyW-runeLen(r[0]))+" - ", r[1])
		}
	}

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.01", "[gold]═══ Getting around ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.02", "Everything in AgeForge is a command: type it at the prompt at the bottom of the screen and press Enter. Panels open on top of the town; the prompt keeps working while one is open."))
	keys([2]string{"PgUp/PgDn", vibe.Line("ui.overlay_help.helpProvider.04", "Scroll a panel that is longer than the screen, like this one")},
		[2]string{"Esc", vibe.Line("ui.overlay_help.helpProvider.06", "Close the panel (with no panel open: save and go to the main menu)")},
		[2]string{"Tab", vibe.Line("ui.overlay_help.helpProvider.08", "Finish the word you are typing (press again for the next match)")},
		[2]string{"help", vibe.Line("ui.overlay_help.helpProvider.10", "Open this page again at any time")})
	note(vibe.Line("ui.overlay_help.helpProvider.11", "New here? Try: build hut, then status, then research. Each panel below lists its own keys."))
	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.12", "\n[gold]═══ The dashboard ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.13", "With no panel open."))
	keys([2]string{"PgUp/PgDn", vibe.Line("ui.overlay_help.helpProvider.15", "Scroll the Buildings list")},
		[2]string{resourcePageKeyName, vibe.Line("ui.overlay_help.helpProvider.16", "The next page of the Resources box, when it has more than it can show")},
		[2]string{"Esc", vibe.Line("ui.overlay_help.helpProvider.18", "Save and go to the main menu (with a panel open: close it)")})

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.19", "\n[gold]═══ The prompt ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.20", "As you type, the best completion shows in dim text after the cursor."))
	keys([2]string{"Tab", vibe.Line("ui.overlay_help.helpProvider.22", "Take the completion; press again for the next one")},
		[2]string{"→", vibe.Line("ui.overlay_help.helpProvider.23", "Take the completion (cursor at the end of the line)")},
		[2]string{"Enter", vibe.Line("ui.overlay_help.helpProvider.25", "Run the line; an unfinished line runs its completion (irreversible ones, like sell, are only filled in: Enter again runs them)")},
		[2]string{"↑/↓", vibe.Line("ui.overlay_help.helpProvider.26", "Command history")})

	sb.WriteString("\n")
	for i, sec := range helpSections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("[gold]═══ " + sec.title.String() + " ═══[-]\n")
		if sec.note.Plain() != "" {
			note(sec.note.String())
		}
		var rows []Usage
		for _, c := range reg {
			if c.Section == sec.name {
				rows = appendHelpRows(rows, c)
			}
		}
		formW := 0
		for _, r := range rows {
			if n := len([]rune(r.Form)); n > formW {
				formW = n
			}
		}
		for _, r := range rows {
			row("  "+helpForm(r.Form)+strings.Repeat(" ", formW-len([]rune(r.Form)))+" - ", r.Text)
		}
	}

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.27", "\n[gold]═══ Panels ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.28", "Type the command to open the panel."))
	for _, name := range panelOrder {
		if c := lookup(reg, name); c != nil && c.Panel != "" {
			row("  [cyan]"+padRight(c.Name, 12)+"[-] - ", c.Panel)
		}
	}
	row("  [cyan]"+padRight(vibe.Line("ui.overlay_help.helpProvider.29", "Accounts"), 12)+"[-] - ", vibe.Line("ui.overlay_help.helpProvider.30", "Switch, create or back up accounts [gray](main-menu panel, not a command)[-]"))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.31", "\n[gold]═══ The Map panel ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.32", "The prompt keeps working while the Map is open: type commands as usual. The map takes the keys that print nothing, and + - = [ ] while the prompt is empty."))
	keys([2]string{"Arrows", vibe.Line("ui.overlay_help.helpProvider.34", "Move the cursor (roguelike) or scroll (skyline); Shift moves further, half a screen in the skyline")},
		[2]string{"Tab", vibe.Line("ui.overlay_help.helpProvider.36", "Next building, wonder or civilization (Shift-Tab: the one before)")},
		[2]string{"PgUp/PgDn", vibe.Line("ui.overlay_help.helpProvider.38", "Zoom out and in")},
		[2]string{"+/-", vibe.Line("ui.overlay_help.helpProvider.39", "Zoom in and out in the skyline (= works as +): far shows the whole town, close every building twice the size; the level is kept for next time")},
		[2]string{"[ ]", vibe.Line("ui.overlay_help.helpProvider.40", "Skyline: jump to the previous and next district (the bar at the bottom shows your skyline and a box where you are looking)")},
		[2]string{"Home/End", vibe.Line("ui.overlay_help.helpProvider.42", "The town square (roguelike); the oldest district and the present (skyline)")},
		[2]string{"Enter", vibe.Line("ui.overlay_help.helpProvider.44", "Put the command for what the cursor is on in the prompt (with something typed, Tab and Enter act on the prompt instead)")},
		[2]string{"Esc", vibe.Line("ui.overlay_help.helpProvider.46", "Close the panel")})
	note(vibe.Line("ui.overlay_help.helpProvider.47", "Settings are commands: map style, map glyphs, map flows, minimap."))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.48", "\n[gold]═══ The Research panel ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.49", "The tech tree. The prompt keeps working while it is open; the tree takes the keys that print nothing."))
	keys(researchKeys...)
	note(vibe.Line("ui.overlay_help.helpProvider.50", "Zoom is a command too: research tree close, research tree far."))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.51", "\n[gold]═══ The badge case ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.52", "Your account's badges. The prompt keeps working while it is open; the case takes the keys that print nothing."))
	keys(badgeKeys...)
	note(vibe.Line("ui.overlay_help.helpProvider.53", "Finding one is a command: badges next, badges <family>, badges <name>. Motion is a setting: motion off holds everything still."))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.54", "\n[gold]═══ The Expeditions panel ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.55", "The Ready Room shows the age's lands with every offer as a pin; the Expedition Report follows the parties out. The prompt keeps working while it is open; the panel takes the keys that print nothing."))
	keys(expeditionKeys...)
	note(vibe.Line("ui.overlay_help.helpProvider.56", "Sending by name is a command: expedition <name>, with as much of the name as it takes to tell the offers apart."))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.57", "\n[gold]═══ The People panel ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.58", "Every person in the town: workers, the reserve, soldiers and scholars, each by the name the age gives them. The prompt keeps working while it is open; the panel takes the keys that print nothing."))
	keys(peopleKeys...)
	note(vibe.Line("ui.overlay_help.helpProvider.59", "Shares are a command: roster <area> <percent|auto>. The panel opens with people or workers."))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.60", "\n[gold]═══ The Advance panel ═══[-]\n"))
	note(vibe.Line("ui.overlay_help.helpProvider.61", "The gate into the next age as four pillars (Knowledge, People, Industry and Monument), with the march of ages and a calendar year that moves as you close in. The prompt keeps working while it is open; the panel takes the keys that print nothing."))
	keys(advanceKeys...)
	note(vibe.Line("ui.overlay_help.helpProvider.62", "The panel opens with charter. Typing advance does the same as Enter on it."))

	sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.63", "\n[gold]═══ Shortcuts ═══[-]\n"))
	var short []string
	for _, c := range reg {
		for _, a := range c.Aliases {
			short = append(short, a+"="+c.Name)
		}
	}
	note(strings.Join(short, ", "))

	// Developer Console — only listed when dev mode is active (Ctrl+K passphrase).
	// Hidden entirely otherwise so the reference stays clean for normal play.
	if game.DevModeActive {
		sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.64", "\n[gold]═══ Developer console ═══[-]\n"))
		sb.WriteString(vibe.Line("ui.overlay_help.helpProvider.65", "[gray]Dev mode is on. Type these in the [-][cyan]>[-][gray] prompt:[-]\n"))
		for _, d := range devCommands {
			row("  [cyan]"+padRight(d.form, 28)+"[-] - ", d.text)
		}
	}

	return sb.String()
}

// appendHelpRows adds c's help rows, then its subcommands', depth first.
func appendHelpRows(rows []Usage, c *Command) []Usage {
	rows = append(rows, c.Help...)
	for _, s := range c.Subs {
		rows = appendHelpRows(rows, s)
	}
	return rows
}

// helpForm colors a usage form: the command words cyan, placeholders plain.
func helpForm(form string) string {
	words := strings.Fields(form)
	n := 0
	for n < len(words) && !strings.ContainsAny(words[n][:1], "<[") {
		n++
	}
	head := strings.Join(words[:n], " ")
	rest := strings.Join(words[n:], " ")
	out := "[cyan]" + tview.Escape(head) + "[-]"
	if rest != "" {
		out += " " + tview.Escape(rest)
	}
	return out
}

// padRight pads s with spaces on the right to width w (no truncation).
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
