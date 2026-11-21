package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// helpProvider renders the Help overlay from the command registry
// (commands.go): each section's usage rows, the panels the player can open,
// the prompt's keys and the shortcuts. It is intentionally static (it does
// not read game state) so the same reference is available at any point in
// play, apart from the lessons the account has learned.
func helpProvider(learned []string, _ game.GameState, screenW int) string {
	var sb strings.Builder
	reg := registry()
	// Every row is a key or a command, then what it does: the description
	// wraps under itself, so the two columns hold at any width.
	width := overlayTextWidth(screenW)
	if screenW <= 0 {
		width = 1 << 20 // no screen to fit: nothing wraps
	}
	row := func(lead, text string) { sb.WriteString(hangingRow(lead, text, width)) }
	// head is a section's name, with a rule when the text has a width.
	head := func(name string) string { return headLine(name, width) + "\n" }
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

	sb.WriteString(lessonsHelp(learned, width))
	sb.WriteString(head(rowid.Line("ui.overlay_help.helpProvider.01", "Getting around")))
	note(rowid.Line("ui.overlay_help.helpProvider.02", "Everything in AgeForge is a command: type it at the prompt at the bottom of the screen and press Enter. Panels open over the whole screen, and the prompt keeps working in most of them. The Wonders, Harbinger and Build Plan panels use letter keys, so press Esc before you type a command there."))
	keys([2]string{"PgUp/PgDn", rowid.Line("ui.overlay_help.helpProvider.04", "Scroll a panel that is longer than the screen, like this one")},
		[2]string{"Option+arrows", rowid.Line("ui.overlay_help.helpProvider.67", "No Page Up, Page Down, Home or End on your keyboard? Option+Up and Option+Down are PgUp and PgDn, and Option+Left and Option+Right are Home and End, in every panel. Alt does the same on other keyboards")},
		[2]string{"Esc", rowid.Line("ui.overlay_help.helpProvider.06", "Close the panel (with no panel open: pause the game)")},
		[2]string{"Tab", rowid.Line("ui.overlay_help.helpProvider.08", "Finish the word you are typing (press again for the next match)")},
		[2]string{"help", rowid.Line("ui.overlay_help.helpProvider.10", "Open this page again at any time")})
	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.12", "The dashboard")))
	note(rowid.Line("ui.overlay_help.helpProvider.13", "With no panel open."))
	keys([2]string{"PgUp/PgDn", rowid.Line("ui.overlay_help.helpProvider.15", "Scroll the Buildings list")},
		[2]string{resourcePageKeyName, rowid.Line("ui.overlay_help.helpProvider.16", "The next page of the Resources box, when it has more than it can show")},
		[2]string{"Esc", rowid.Line("ui.overlay_help.helpProvider.18", "Pause the game, with Resume, Settings, Save and Main menu (with a panel open: close it)")})

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.19", "The prompt")))
	note(rowid.Line("ui.overlay_help.helpProvider.20", "As you type, the best completion shows in dim text after the cursor."))
	keys([2]string{"Tab", rowid.Line("ui.overlay_help.helpProvider.22", "Take the completion; press again for the next one")},
		[2]string{"→", rowid.Line("ui.overlay_help.helpProvider.23", "Take the completion (cursor at the end of the line)")},
		[2]string{"Enter", rowid.Line("ui.overlay_help.helpProvider.25", "Run the line; an unfinished line runs its completion (irreversible ones, like sell, are only filled in: Enter again runs them)")},
		[2]string{"↑/↓", rowid.Line("ui.overlay_help.helpProvider.26", "Command history")})

	sb.WriteString("\n")
	for i, sec := range helpSections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(head(sec.title.String()))
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

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.27", "Panels")))
	note(rowid.Line("ui.overlay_help.helpProvider.28", "Type the command to open the panel."))
	for _, name := range panelOrder {
		if c := lookup(reg, name); c != nil && c.Panel != "" {
			row("  [cyan]"+padRight(c.Name, 12)+"[-] - ", c.Panel)
		}
	}
	row("  [cyan]"+padRight(rowid.Line("ui.overlay_help.helpProvider.29", "Accounts"), 12)+"[-] - ", rowid.Line("ui.overlay_help.helpProvider.30", "Switch, create or back up accounts [gray](main-menu panel, not a command)[-]"))

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.31", "The Map panel")))
	note(rowid.Line("ui.overlay_help.helpProvider.32", "The prompt keeps working while the Map is open: type commands as usual. The map takes the keys that print nothing, and + - = while the prompt is empty."))
	keys([2]string{"Arrows", rowid.Line("ui.overlay_help.helpProvider.34", "Move the cursor; Shift moves further")},
		[2]string{"Tab", rowid.Line("ui.overlay_help.helpProvider.36", "Next building, wonder or civilization (Shift-Tab: the one before)")},
		[2]string{"PgUp/PgDn", rowid.Line("ui.overlay_help.helpProvider.38", "Zoom out and in (Option+Up and Option+Down do the same)")},
		[2]string{"+/-", rowid.Line("ui.overlay_help.helpProvider.39", "Zoom in and out with no Page keys (= works as +)")},
		[2]string{"Home/End", rowid.Line("ui.overlay_help.helpProvider.42", "The town square")},
		[2]string{"Enter", rowid.Line("ui.overlay_help.helpProvider.44", "Put the command for what the cursor is on in the prompt (with something typed, Tab and Enter act on the prompt instead)")},
		[2]string{"Esc", rowid.Line("ui.overlay_help.helpProvider.46", "Close the panel")})
	note(rowid.Line("ui.overlay_help.helpProvider.47", "Settings are commands: map glyphs, map flows, minimap."))

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.48", "The Research panel")))
	note(rowid.Line("ui.overlay_help.helpProvider.49", "The board. The prompt keeps working while it is open; the board takes the keys that print nothing."))
	keys(helpPairs(boardKeys)...)
	note(rowid.Line("ui.overlay_help.helpProvider.50", "Zoom is a command too: research tree close, research tree far."))
	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.68", "The tech tree")))
	note(rowid.Line("ui.overlay_help.helpProvider.69", "The whole tree as a map, opened with T on the board. Esc goes back to the board."))
	keys(helpPairs(treeKeys)...)

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.51", "The Ranks screen")))
	note(rowid.Line("ui.overlay_help.helpProvider.52", "Your account's badges, as a rank for each worker domain. The prompt keeps working while it is open; the screen takes the keys that print nothing."))
	keys(helpPairs(badgeKeys)...)
	note(rowid.Line("ui.overlay_help.helpProvider.53", "Finding one is a command: badges all, badges recent, badges <domain>, badges <name>."))

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.54", "The Expeditions panel")))
	note(rowid.Line("ui.overlay_help.helpProvider.55", "The Ready Room shows the age's lands with every offer as a pin; the Expedition Report follows the parties out. The prompt keeps working while it is open; the panel takes the keys that print nothing."))
	keys(helpPairs(expeditionKeys)...)
	note(rowid.Line("ui.overlay_help.helpProvider.56", "Sending by name is a command: expedition <name>, with as much of the name as it takes to tell the offers apart."))

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.57", "The People panel")))
	note(rowid.Line("ui.overlay_help.helpProvider.58", "Every person in the town: workers, the reserve, soldiers and scholars, each by the name the age gives them. The prompt keeps working while it is open; the panel takes the keys that print nothing."))
	keys(helpPairs(peopleKeys)...)
	note(rowid.Line("ui.overlay_help.helpProvider.59", "Shares are a command: roster <area> <percent|auto>. The panel opens with people or workers."))

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.60", "The Advance panel")))
	note(rowid.Line("ui.overlay_help.helpProvider.61", "The gate into the next age as four pillars (Knowledge, People, Industry and Monument), with the march of ages and a calendar year that moves as you close in. The prompt keeps working while it is open; the panel takes the keys that print nothing."))
	keys(helpPairs(advanceKeys)...)
	note(rowid.Line("ui.overlay_help.helpProvider.62", "The panel opens with charter. Typing advance does the same as Enter on it."))

	sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.63", "Shortcuts")))
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
		sb.WriteString("\n" + head(rowid.Line("ui.overlay_help.helpProvider.64", "Developer console")))
		sb.WriteString(rowid.Line("ui.overlay_help.helpProvider.65", "[gray]Dev mode is on. Type these in the [-][cyan]>[-][gray] prompt:[-]\n"))
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
	out := "[cyan]" + kit.Escape(head) + "[-]"
	if rest != "" {
		out += " " + kit.Escape(rest)
	}
	return out
}

// padRight pads s with spaces on the right to width w (no truncation).
func padRight(s string, w int) string {
	n := runeLen(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}
