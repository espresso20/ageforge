package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
)

// helpProvider renders the Help overlay: a categorised command reference and a
// list of the panels the player can open. It is intentionally static (it does
// not read game state) so the same reference is available at any point in play.
func helpProvider(_ game.GameState, _ int) string {
	var sb strings.Builder

	sb.WriteString("[gold]═══ Actions ═══[-]\n")
	sb.WriteString("  [cyan]gather[-] <food|wood|stone> [n] - Hand-gather resources (max 10, until Medieval Age)\n")
	sb.WriteString("  [cyan]build[-] <building> [count|max] - Build structure(s) (default: 1)\n")
	sb.WriteString("  [cyan]sell[-] <building> [count]      - Demolish building(s), recover 50% of build cost\n")
	sb.WriteString("  [cyan]advance[-]                       - Advance to the next age (when ready)\n")
	sb.WriteString("  [cyan]upgrade[-]                       - List available building upgrades\n")
	sb.WriteString("  [cyan]upgrade[-] <building> [n|all]    - Upgrade building to next age tier (pays cost delta)\n")

	sb.WriteString("\n[gold]═══ Workers ═══[-]\n")
	sb.WriteString("  [cyan]recruit[-] [count|max]           - Recruit workers from available housing (default: 1)\n")
	sb.WriteString("  [cyan]assign[-] <building> [n|all]     - Assign workers to a building\n")
	sb.WriteString("  [cyan]unassign[-] <building> [n|all]   - Unassign workers from a building\n")
	sb.WriteString("  [cyan]dismiss[-] <building> [n|all]    - Fire workers from a building (removes from pool)\n")

	sb.WriteString("\n[gold]═══ Research, Expeditions & Army ═══[-]\n")
	sb.WriteString("  [cyan]research[-] <tech_key>          - Research a technology\n")
	sb.WriteString("  [cyan]research[-] cancel             - Cancel current research\n")
	sb.WriteString("  [cyan]research[-] list               - List available techs\n")
	sb.WriteString("  [cyan]expedition[-]                  - Open the Expeditions (scouting) panel\n")
	sb.WriteString("  [cyan]expedition[-] <key>            - Send a scouting expedition (costs resources)\n")
	sb.WriteString("  [cyan]expedition[-] list             - List available expeditions\n")
	sb.WriteString("  [cyan]army[-]                        - Open the Army (military) panel\n")
	sb.WriteString("  [cyan]campaign[-] <key>             - Wage a military campaign (costs soldiers)\n")
	sb.WriteString("  [cyan]campaign[-] list              - List available campaigns\n")

	sb.WriteString("\n[gold]═══ Trade & Diplomacy ═══[-]\n")
	sb.WriteString("  [cyan]trade[-] <from> <to> <amount>  - Exchange resources\n")
	sb.WriteString("  [cyan]trade[-] list                  - Show exchange rates\n")
	sb.WriteString("  [cyan]trade[-] route list            - List trade routes\n")
	sb.WriteString("  [cyan]trade[-] route start <key>     - Start a trade route\n")
	sb.WriteString("  [cyan]trade[-] route stop <key>      - Stop a trade route\n")
	sb.WriteString("  [cyan]blackmarket[-] [resource]      - High-risk culture gamble for a resource haul (colonial+)\n")
	sb.WriteString("  [cyan]diplomacy[-]                   - Open the Diplomacy overlay (factions, opinion, status)\n")
	sb.WriteString("  [cyan]diplomacy[-] ally <faction>    - Ally with faction (costs gold)\n")
	sb.WriteString("  [cyan]diplomacy[-] rival <faction>   - Declare rivalry\n")
	sb.WriteString("  [cyan]diplomacy[-] embargo <faction> - Embargo faction\n")
	sb.WriteString("  [cyan]diplomacy[-] gift <faction>    - Send gift (+15 opinion)\n")
	sb.WriteString("  [cyan]diplomacy[-] neutral <faction> - Reset to neutral\n")

	sb.WriteString("\n[gold]═══ Wonders & Prestige ═══[-]\n")
	sb.WriteString("  [cyan]wonder[-]                          - Show current wonder bank status\n")
	sb.WriteString("  [cyan]wonder collect[-] <res> <amt|all> - Bank resources into current wonder\n")
	sb.WriteString("  [cyan]prestige[-]                        - View prestige status\n")
	sb.WriteString("  [cyan]prestige[-] confirm yes            - Reset game with prestige bonus\n")
	sb.WriteString("  [cyan]prestige[-] shop                   - View prestige upgrades\n")
	sb.WriteString("  [cyan]prestige[-] buy <key>              - Buy a prestige upgrade\n")
	sb.WriteString("  [cyan]festival[-]                        - Spend culture for a temporary production boost\n")
	sb.WriteString("  [cyan]festival confirm yes[-]            - Hold the festival now\n")
	sb.WriteString("  [cyan]catastrophe invoke[-]              - Voluntarily trigger the epoch catastrophe\n")

	sb.WriteString("\n[gold]═══ Game ═══[-]\n")
	sb.WriteString("  [cyan]rates[-]                       - Show resource rate breakdown\n")
	sb.WriteString("  [cyan]status[-]                      - Show detailed status\n")
	sb.WriteString("  [cyan]speed[-] [1.0|1.5|2.0|...]     - Set game speed (unlocks per wonder built)\n")
	sb.WriteString("  [cyan]theme[-]                       - Open the theme picker (palettes + accessibility)\n")
	sb.WriteString("  [cyan]theme[-] list                  - List themes with unlock status\n")
	sb.WriteString("  [cyan]theme[-] <key>                 - Switch to a theme by key\n")
	sb.WriteString("  [cyan]save[-] [name]                 - Save game (default: autosave)\n")
	sb.WriteString("  [cyan]load[-] [name]                 - Load game (default: autosave)\n")
	sb.WriteString("  [cyan]saves[-]                       - List all save files\n")
	sb.WriteString("  [cyan]dump[-]                        - Export logs to file for debugging\n")
	sb.WriteString("  [cyan]help[-]                        - Open this Help panel\n")

	sb.WriteString("\n[gold]═══ Accounts ═══[-]\n")
	sb.WriteString("[gray]Each account is its own slot. Switch/new/wipe live in the Accounts panel (main menu).[-]\n")
	sb.WriteString("  [cyan]account[-]                     - Show this account's ID, recovery code & backup help\n")
	sb.WriteString("  [cyan]account[-] list                - List your local accounts\n")
	sb.WriteString("  [cyan]account[-] switch <name>       - Switch to an existing local account\n")
	sb.WriteString("  [cyan]account[-] export [path]       - Back up this account's progress to a file\n")
	sb.WriteString("  [cyan]account[-] backup              - Full snapshot (account.json + saves) to data/backups/\n")
	sb.WriteString("  [cyan]account[-] import <path>       - Restore an account from a backup file\n")
	sb.WriteString("  [cyan]account[-] recover <code>      - Restore your identity from a recovery code\n")
	sb.WriteString("[gray]Wiping or exporting an account also auto-creates a full backup first (last 10 kept).[-]\n")

	sb.WriteString("\n[gold]═══ Panels ═══[-]\n")
	sb.WriteString("[gray]Type the command to open the panel.[-]\n")
	for _, p := range []struct{ cmd, desc string }{
		{"milestones", "Milestone goals & rewards"},
		{"research", "Technology tree & progress"},
		{"expedition", "Scouting expeditions (resource cost)"},
		{"army", "Army overview & military campaigns"},
		{"trade", "Exchange rates & trade routes"},
		{"diplomacy", "Factions, opinion & status"},
		{"stats", "Empire statistics"},
		{"wonders", "Wonder bank & built wonders"},
		{"workers", "Worker domains & assignments"},
		{"logs", "Recent game log entries"},
		{"epoch", "Epoch progress & catastrophe"},
		{"history", "Civilization history timeline"},
		{"buildings", "Built structures by lineage"},
		{"citymap", "Your settlement map (alias: map)"},
		{"worldmap", "Known world — your civ & the diplomacy civs"},
		{"theme", "Theme picker — palettes & accessibility"},
		{"help", "This Help panel"},
	} {
		sb.WriteString("  [cyan]" + padRight(p.cmd, 12) + "[-] — " + p.desc + "\n")
	}
	sb.WriteString("  [cyan]" + padRight("Accounts", 12) + "[-] — Switch/create/back-up accounts [gray](main-menu panel, not a command)[-]\n")

	sb.WriteString("\n[gold]═══ Shortcuts ═══[-]\n")
	sb.WriteString("[gray]g=gather, b=build, r=recruit, a=assign, u=unassign, s=status, res=research, exp=expedition, t=trade, dip=diplomacy[-]\n")

	// Developer Console — only listed when dev mode is active (Ctrl+K passphrase).
	// Hidden entirely otherwise so the reference stays clean for normal play.
	if game.DevModeActive {
		sb.WriteString("\n[gold]═══ Developer Console ═══[-]\n")
		sb.WriteString("[gray]DEV mode active — type these in the [-][cyan]>[-][gray] prompt:[-]\n")
		for _, d := range []struct{ cmd, desc string }{
			{"/god", "Toggle godmode — free costs, instant builds"},
			{"/fill", "Fill all resources to their storage cap"},
			{"/give <resource> <amount>", "Add an amount of a resource"},
			{"/build <building_key>", "Instantly place one building"},
			{"/techs", "Unlock all techs up to the current age"},
			{"/age <age_key>", "Jump to any age"},
			{"/ages", "List all age keys"},
			{"/prestige <level 0-9>", "Set prestige level"},
			{"/speed <multiplier>", "Set the tick-speed multiplier"},
		} {
			sb.WriteString("  [cyan]" + padRight(d.cmd, 28) + "[-] — " + d.desc + "\n")
		}
	}

	return sb.String()
}

// padRight pads s with spaces on the right to width w (no truncation).
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
