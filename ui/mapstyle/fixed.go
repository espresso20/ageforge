package mapstyle

import (
	"sync"

	"github.com/espresso20/ageforge/theme"
)

// The map's picture (the ground, the city, the sky, the cursor on them) keeps
// one palette in every theme: the full-colour dark one of the default theme,
// Forge. A light theme does not redraw it light and a one-ink theme does not
// draw it in one ink. The words and the frame round it (the header, the news,
// the legend, the inspector, the borders) are the active theme's.

// pictureTheme is the theme the picture is drawn in.
var pictureTheme = sync.OnceValue(func() theme.Theme {
	t, _ := theme.ByKey(theme.DefaultKey)
	return t
})

// pictureRole is a role's colour in the picture's theme.
func pictureRole(r theme.Role) theme.Color { return pictureTheme().Color(r) }
