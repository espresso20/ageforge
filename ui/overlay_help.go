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
func helpProvider(_ game.GameState, _ int) string {
	var sb strings.Builder
	reg := registry()

	for i, sec := range helpSections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("[gold]═══ " + sec.name + " ═══[-]\n")
		if sec.note != "" {
			sb.WriteString("[gray]" + sec.note + "[-]\n")
		}
		var rows []Usage
		for _, c := range reg {
			if c.Section == sec.name {
				rows = appendHelpRows(rows, c)
			}
		}
		width := 0
		for _, r := range rows {
			if n := len([]rune(r.Form)); n > width {
				width = n
			}
		}
		for _, r := range rows {
			sb.WriteString("  " + helpForm(r.Form) + strings.Repeat(" ", width-len([]rune(r.Form))) + " - " + r.Text + "\n")
		}
	}

	sb.WriteString("\n[gold]═══ Panels ═══[-]\n")
	sb.WriteString("[gray]Type the command to open the panel.[-]\n")
	for _, name := range panelOrder {
		if c := lookup(reg, name); c != nil && c.Panel != "" {
			sb.WriteString("  [cyan]" + padRight(c.Name, 12) + "[-] — " + c.Panel + "\n")
		}
	}
	sb.WriteString("  [cyan]" + padRight("Accounts", 12) + "[-] — Switch/create/back-up accounts [gray](main-menu panel, not a command)[-]\n")

	sb.WriteString("\n[gold]═══ The Prompt ═══[-]\n")
	sb.WriteString("[gray]As you type, the best completion shows in dim text after the cursor.[-]\n")
	sb.WriteString("  [cyan]Tab[-]    - Take the completion; press again for the next one\n")
	sb.WriteString("  [cyan]→[-]      - Take the completion (cursor at the end of the line)\n")
	sb.WriteString("  [cyan]Enter[-]  - Run the line; an unfinished line runs its completion\n")
	sb.WriteString("           (irreversible ones, like sell, are only filled in: Enter again runs them)\n")
	sb.WriteString("  [cyan]↑/↓[-]    - Command history\n")

	sb.WriteString("\n[gold]═══ Shortcuts ═══[-]\n")
	var short []string
	for _, c := range reg {
		for _, a := range c.Aliases {
			short = append(short, a+"="+c.Name)
		}
	}
	sb.WriteString("[gray]" + strings.Join(short, ", ") + "[-]\n")

	// Developer Console — only listed when dev mode is active (Ctrl+K passphrase).
	// Hidden entirely otherwise so the reference stays clean for normal play.
	if game.DevModeActive {
		sb.WriteString("\n[gold]═══ Developer Console ═══[-]\n")
		sb.WriteString("[gray]DEV mode active — type these in the [-][cyan]>[-][gray] prompt:[-]\n")
		for _, d := range devCommands {
			sb.WriteString("  [cyan]" + padRight(d.form, 28) + "[-] — " + d.text + "\n")
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

// helpForm colours a usage form: the command words cyan, placeholders plain.
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
