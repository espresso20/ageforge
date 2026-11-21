package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
)

// shell.go is the frame every full panel is drawn in: a box with the panel's
// name in its top edge and the close button at its right, the panel's tabs on
// the row under it, and the keys along the foot. A panel draws its own
// content in the cells the shell hands back and nothing of the frame itself,
// so what a panel's frame looks like is decided here and nowhere else.

// shell draws the frame on the whole grid: the box, title in its top edge,
// the close button, and the tabs (none for a panel without) with the one
// numbered open showing. It returns the panel's own cells: inside the edges
// with a column of air each side, under the tabs and above the key row.
func (g *xgrid) shell(title string, open int, tabs ...string) image.Rectangle {
	g.box(0, 0, g.w, g.h, xBorder, title, xTitle, false)
	g.closeButton(g.w)
	top, x := 1, panelPadX
	for i, name := range tabs {
		k := xDim
		if i == open {
			k = xSel
		}
		end := g.put(x, 1, " "+name+" ", k)
		g.tab(image.Rect(x, 1, end, 2), name, i, i == open)
		x, top = end+1, 2
	}
	return image.Rect(panelPadX, top, max(panelPadX+1, g.w-panelPadX), max(top+1, g.h-2))
}

// foot writes the panel's keys on the shell's key row, on a grid tall enough
// to have one.
func (g *xgrid) foot(keys [][2]string, hint string) {
	if g.h >= panelMinRows {
		g.keyRow(g.h-2, keys, hint)
	}
}

// keyRow writes keys along row y as buttons, each a key and what it does. The
// last one (how to close the panel) is there whenever it fits; the others are
// those that fit before it, in order. hint is a command the row says may be
// typed, dropped first: a click on it writes it in the command bar, up to the
// first part the player has to fill in.
func (g *xgrid) keyRow(y int, keys [][2]string, hint string) {
	x := g.keyButtons(y, keys)
	if hint == "" {
		return
	}
	lead := typeLead()
	if x+runeLen(lead)+chipW(hint, "") <= g.w-2 {
		g.chip(g.put(x, y, lead, xFaint), y, chipSpec{kind: keySteel, label: hint, cmd: typedPart(hint), fill: true})
	}
}

// typedPart is the start of a hint that is typed as it stands: the words
// before the first one that names a choice or a value ("a/b", "<key>",
// "[name]").
func typedPart(hint string) string {
	// A leading "/" marks a developer command, not a choice.
	i := strings.IndexAny(hint[min(1, len(hint)):], "<[|/")
	if i < 0 {
		return hint
	}
	i++
	return hint[:strings.LastIndex(hint[:i], " ")+1]
}

// footWords is foot for a panel that says only that a command may be typed,
// with no command to offer: the words are plain text, not a button.
func (g *xgrid) footWords(keys [][2]string, words string) {
	if g.h < panelMinRows {
		return
	}
	y := g.h - 2
	x := g.keyButtons(y, keys)
	lead := typeLead()
	if words != "" && x+runeLen(lead)+runeLen(words) <= g.w-2 {
		g.put(g.put(x, y, lead, xFaint), y, words, xText)
	}
}

// typeLead is the words before what a key row says may be typed.
func typeLead() string { return rowid.Line("ui.expedition_view.expKeyBar.01", " · or type ") }

// keyButtons writes the keys along row y and returns the next free x.
func (g *xgrid) keyButtons(y int, keys [][2]string) int {
	x := 2
	room := g.w - 2
	last, closeW := len(keys)-1, 0
	if last >= 0 {
		closeW = keyButtonW(keys[last][0], keys[last][1])
	}
	for i, kv := range keys {
		at := x
		if x > 2 {
			at++
		}
		end := at + keyButtonW(kv[0], kv[1])
		if i < last {
			end += 1 + closeW
		}
		if end <= room {
			x = g.keyButton(at, y, kv[0], kv[1])
		}
	}
	return x
}

// head is a section's name in the title style with a faint rule across the
// rest of the w cells of its row.
func (g *xgrid) head(x, y, w int, s string) {
	for end := g.put(x, y, s, xTitle); end < x+w; end++ {
		g.set(end, y, '─', xFaint)
	}
}

// headLine is head for a text panel: the section's name as a markup line (no
// newline), in the title colour and bold, then a faint rule out to width
// cells. A width of 0, or one that stands for no limit, draws no rule.
func headLine(s string, width int) string {
	out := "[" + theme.TagName(theme.RoleHighlight) + "::b]" + s + theme.Reset
	if n := width - runeLen(s) - 1; n > 0 && width < 1<<10 {
		out += " " + theme.Paint(theme.RoleDim, strings.Repeat("─", n))
	}
	return out
}
