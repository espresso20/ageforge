package ui

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
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

	for i, sec := range helpSections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("[gold]═══ " + sec.name + " ═══[-]\n")
		if sec.note != "" {
			note(sec.note)
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

	sb.WriteString("\n[gold]═══ Panels ═══[-]\n")
	note("Type the command to open the panel.")
	for _, name := range panelOrder {
		if c := lookup(reg, name); c != nil && c.Panel != "" {
			row("  [cyan]"+padRight(c.Name, 12)+"[-] - ", c.Panel)
		}
	}
	row("  [cyan]"+padRight("Accounts", 12)+"[-] - ", "Switch, create or back up accounts [gray](main-menu panel, not a command)[-]")

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

	sb.WriteString("\n[gold]═══ The dashboard ═══[-]\n")
	note("With no panel open.")
	keys([2]string{"PgUp/PgDn", "Scroll the Buildings list"},
		[2]string{resourcePageKeyName, "The next page of the Resources box, when it has more than it can show"},
		[2]string{"Esc", "Save and go to the main menu (with a panel open: close it)"})

	sb.WriteString("\n[gold]═══ The prompt ═══[-]\n")
	note("As you type, the best completion shows in dim text after the cursor.")
	keys([2]string{"Tab", "Take the completion; press again for the next one"},
		[2]string{"→", "Take the completion (cursor at the end of the line)"},
		[2]string{"Enter", "Run the line; an unfinished line runs its completion (irreversible ones, like sell, are only filled in: Enter again runs them)"},
		[2]string{"↑/↓", "Command history"})

	sb.WriteString("\n[gold]═══ The Map panel ═══[-]\n")
	note("The prompt keeps working while the Map is open: type commands as usual. The map takes the keys that print nothing.")
	keys([2]string{"Arrows", "Move the cursor (roguelike) or scroll (skyline); Shift moves further"},
		[2]string{"Tab", "Next building, wonder or civilization (Shift-Tab: the one before)"},
		[2]string{"PgUp/PgDn", "Zoom out and in (roguelike) or scroll half a screen (skyline)"},
		[2]string{"Home/End", "The town square (roguelike); the oldest district and the present (skyline)"},
		[2]string{"Enter", "Put the command for what the cursor is on in the prompt (with something typed, Tab and Enter act on the prompt instead)"},
		[2]string{"Esc", "Close the panel"})
	note("Settings are commands: map style, map glyphs, map flows, minimap.")

	sb.WriteString("\n[gold]═══ Shortcuts ═══[-]\n")
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
		sb.WriteString("\n[gold]═══ Developer console ═══[-]\n")
		sb.WriteString("[gray]Dev mode is on. Type these in the [-][cyan]>[-][gray] prompt:[-]\n")
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
