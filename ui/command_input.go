package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/theme"
)

// commandInput is the dashboard's command prompt: a tview InputField that
// shows the best completion of what was typed as dim ghost text after the
// cursor, fish-style. Tab takes the completion, → at the end of the line
// takes it too; Enter lives in
// Dashboard.submitInput. There is no dropdown.
type commandInput struct {
	*tview.InputField
	comp *completer

	// atEnd: at the last draw the cursor sat right after the text, so the
	// ghost was showing and → takes it.
	atEnd bool
}

func newCommandInput(comp *completer) *commandInput {
	return &commandInput{InputField: tview.NewInputField(), comp: comp}
}

// Ghost is the completion suffix currently shown after the text.
func (c *commandInput) Ghost() string { return c.comp.ghost(c.GetText()) }

// Draw draws the field, then the ghost text from the cursor on, when the
// cursor sits at the end of the text and the field has room.
func (c *commandInput) Draw(screen tcell.Screen) {
	spy := &cursorSpy{Screen: screen}
	c.InputField.Draw(spy)
	c.atEnd = false
	if !spy.shown || !c.HasFocus() {
		return
	}
	text := c.GetText()
	x, y, width, _ := c.GetInnerRect()
	start := x + tview.TaggedStringWidth(c.GetLabel()) + uniseg.StringWidth(text)
	// The text area scrolls a long line; then the cursor is not where an
	// unscrolled line ends, and there is no ghost.
	if spy.x != start || spy.y != y {
		return
	}
	c.atEnd = true
	ghost := c.Ghost()
	if ghost == "" {
		return
	}
	style := tcell.StyleDefault.
		Foreground(theme.Color(theme.RoleDim)).
		Background(theme.Color(theme.RoleBackground))
	col := spy.x
	for _, r := range ghost {
		if col >= x+width {
			break
		}
		screen.SetContent(col, y, r, nil, style)
		col++
	}
}

// acceptKey handles the completion keys: Tab takes the first candidate, →
// takes the ghost when the cursor is at the end. It reports whether the key
// was used.
func (c *commandInput) acceptKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyTab:
		if cands := c.comp.candidates(c.GetText()); len(cands) > 0 {
			c.SetText(cands[0])
		}
		return true // Tab never leaves the prompt
	case tcell.KeyRight:
		if !c.atEnd {
			return false
		}
		g := c.Ghost()
		if g == "" {
			return false
		}
		cands := c.comp.candidates(c.GetText())
		line := c.GetText() + g
		if len(cands) > 0 {
			line = cands[0] // keeps the trailing space when more follows
		}
		c.SetText(line)
		return true
	}
	return false
}

// cursorSpy is a screen that remembers where the cursor was put.
type cursorSpy struct {
	tcell.Screen
	x, y  int
	shown bool
}

func (s *cursorSpy) ShowCursor(x, y int) {
	s.x, s.y, s.shown = x, y, true
	s.Screen.ShowCursor(x, y)
}
