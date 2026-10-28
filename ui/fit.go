package ui

import (
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// fit.go holds the helpers that lay text out for the width it will be drawn
// at, so a panel breaks its lines on purpose instead of leaving it to the
// terminal: tview wraps an overlong line back to column 0, which splits a
// rate after its slash and drops a wrapped description under the command
// column.

// visibleLen is how many cells s takes once drawn: color tags take none,
// and a bracketed word tview would not eat counts in full.
func visibleLen(s string) int {
	return tview.TaggedStringWidth(safeTags(s))
}

// overlayTextWidth is the width of the text area of a panel opened over a
// terminal screenW columns wide: the whole width, less the frame's edges and
// the air inside them (panel_frame.go).
func overlayTextWidth(screenW int) int {
	return max(20, screenW-2*panelPadX)
}

// glue stands for a space that must not break: wrapWords keeps the words on
// either side of it together and prints it as a space.
const glue = "\x1f"

// rateInText finds a rate written in prose: a signed number, a space, then
// what it is a rate of ("+0.008 faith/tick", "+13.1K gold/tick").
var rateInText = regexp.MustCompile(`([+-][0-9][0-9.,]*\pL{0,2}) (\S+/tick)`)

// crewInText finds the crew a building's rate is quoted for: "(3 workers)".
var crewInText = regexp.MustCompile(`\((\d+) (workers?\))`)

// glueRates keeps every rate in s whole when it is wrapped: the number
// stays with its unit, and a crew's count with its word.
func glueRates(s string) string {
	s = rateInText.ReplaceAllString(s, "${1}"+glue+"${2}")
	return crewInText.ReplaceAllString(s, "(${1}"+glue+"${2}")
}

// wrapLines packs words into lines at most width cells wide, one space
// between words, as size measures them. A word wider than the width gets a
// line to itself.
func wrapLines(words []string, width int, size func(string) int) []string {
	var lines []string
	var line strings.Builder
	lineSize := 0
	for _, word := range words {
		n := size(word)
		if lineSize > 0 && lineSize+1+n > width {
			lines = append(lines, line.String())
			line.Reset()
			lineSize = 0
		}
		if lineSize > 0 {
			line.WriteByte(' ')
			lineSize++
		}
		line.WriteString(word)
		lineSize += n
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines
}

// wrapWords breaks s at its spaces into lines at most width cells wide. A
// word longer than the width gets a line to itself. Color tags are kept
// with the word they are written against and take no room, and words
// joined with glue stay together. Text with no words is one empty line.
func wrapWords(s string, width int) []string {
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.ReplaceAll(word, glue, " ")
	}
	lines := wrapLines(words, max(width, 1), visibleLen)
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

// wrapText breaks plain text (no color tags) into lines of at most w runes.
func wrapText(s string, w int) []string {
	return wrapLines(strings.Fields(s), w, runeLen)
}

// wrapCells is wrapText for a column that cannot spill: a word longer than
// the line is cut into lines of its own.
func wrapCells(s string, w int) []string {
	if w < 1 {
		return nil
	}
	var words []string
	for _, word := range strings.Fields(s) {
		for runeLen(word) > w {
			runes := []rune(word)
			words = append(words, string(runes[:w]))
			word = string(runes[w:])
		}
		words = append(words, word)
	}
	return wrapLines(words, w, runeLen)
}

// wrapItems breaks a list of items into lines of at most w cells, comma
// separated, never inside an item.
func wrapItems(items []string, w int) []string {
	words := make([]string, len(items))
	for i, item := range items {
		words[i] = item
		if i < len(items)-1 {
			words[i] += ","
		}
	}
	return wrapLines(words, w, runeLen)
}

// truncate shortens s to at most max runes, appending an ellipsis when cut.
// Operates on runes so multibyte backstory text is never split mid-character.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

// fitText is truncate for the grid panels, which show the ellipsis alone
// where a cut string has one cell.
func fitText(s string, room int) string {
	if room == 1 && runeLen(s) > 1 {
		return "…"
	}
	return truncate(s, room)
}

// hangingRow is a two-column row: lead (already padded to its column, tags
// and all), then text wrapped to what is left of width, each further line
// indented to start under the first. When the lead leaves the text under
// minText cells, the text goes on the lines below it instead, indented a
// little, so a long lead on a narrow panel still reads.
func hangingRow(lead, text string, width int) string {
	const minText, fallbackIndent = 24, 6
	indent := visibleLen(lead)
	if width-indent < minText {
		lines := wrapWords(text, width-fallbackIndent)
		return lead + "\n" + strings.Repeat(" ", fallbackIndent) + strings.Join(lines, "\n"+strings.Repeat(" ", fallbackIndent)) + "\n"
	}
	lines := wrapWords(text, width-indent)
	return lead + strings.Join(lines, "\n"+strings.Repeat(" ", indent)) + "\n"
}

// fitView is a text box whose text is written for the size it is drawn at.
// Its render function is handed the box's inner width and height and
// returns lines that fit them; the box never wraps, so a line that did not
// fit would be cut, not broken. The text is written again when the size
// changes or changed is called (new game state).
type fitView struct {
	*tview.TextView
	render func(w, h int) string
	w, h   int
	stale  bool
	// drawn: the box has been drawn, so it has a real size. Until then its
	// size is tview's default and means nothing.
	drawn bool
}

func newFitView(render func(w, h int) string) *fitView {
	return &fitView{TextView: tview.NewTextView().SetDynamicColors(true).SetWrap(false), render: render, stale: true}
}

// changed says what the box shows has changed, and writes it again: for the
// box's size once it has been drawn, and for an unbounded one until then
// (the size of a box never drawn is tview's default and means nothing), so
// code that reads the text back without a screen gets it whole.
func (v *fitView) changed() {
	v.stale = true
	if v.drawn {
		v.layout()
		return
	}
	const unbounded = 1 << 16
	v.SetText(safeTags(v.render(unbounded, unbounded)))
}

// layout writes the text for the box's current size if it is out of date.
func (v *fitView) layout() {
	_, _, w, h := v.GetInnerRect()
	if v.stale || w != v.w || h != v.h {
		v.SetText(safeTags(v.render(w, h)))
		v.w, v.h, v.stale = w, h, false
	}
}

func (v *fitView) Draw(screen tcell.Screen) {
	if !v.drawn {
		v.drawn, v.stale = true, true
	}
	v.layout()
	v.TextView.Draw(screen)
}
