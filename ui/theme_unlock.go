package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/theme"
)

// A gated theme is the reward of an account badge. The engine unlocks it as
// it grants the badge (game.Account), so the UI's part is only to say so:
// when a badge that gives a theme is announced, the log names the theme.

// themeUnlockToast renders the unlock notification text. The accent
// color comes from the active theme's role tag so the toast reads in whatever theme
// is live. Kept separate (and pure) so the message format is testable.
func themeUnlockToast(themeName string) string {
	return fmt.Sprintf("%s🎨 New theme unlocked: %s. Type theme to use it.[-]",
		theme.Tag(theme.RoleAccent), themeName)
}
