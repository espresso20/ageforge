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
// terminal screenW columns wide: the middle 17 of 19 parts (the way the
// overlay's Flex divides them), less the border.
func overlayTextWidth(screenW int) int {
	side := screenW / 19
	w := (screenW-side)*17/18 - 2
	if w < 20 {
		w = 20
	}
	return w
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

// wrapWords breaks s at its spaces into lines at most width cells wide. A
// word longer than the width gets a line to itself. Color tags are kept
// with the word they are written against and take no room, and words
// joined with glue stay together.
func wrapWords(s string, width int) []string {
	if width < 1 {
		width = 1
	}
	var lines []string
	var cur strings.Builder
	curLen := 0
	for _, word := range strings.Fields(s) {
		word = strings.ReplaceAll(word, glue, " ")
		n := visibleLen(word)
		if curLen > 0 && curLen+1+n > width {
			lines = append(lines, cur.String())
			cur.Reset()
			curLen = 0
		}
		if curLen > 0 {
			cur.WriteByte(' ')
			curLen++
		}
		cur.WriteString(word)
		curLen += n
	}
	if cur.Len() > 0 || len(lines) == 0 {
		lines = append(lines, cur.String())
	}
	return lines
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
