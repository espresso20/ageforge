package ui

import "github.com/gdamore/tcell/v2"

// commandInput is the dashboard's command prompt: an input line that shows
// the best completion of what was typed as dim ghost text after the cursor,
// fish-style. Tab takes the completion (again: the next candidate), → at the
// end of the line takes it too; Enter's rules live in completer.enterLine
// and Dashboard.submitInput. There is no dropdown.
type commandInput struct {
	*nativeInput
	comp  *completer
	cycle tabCycle
}

// tabCycle is a run of Tab presses: the candidates for the text before the
// first Tab, which one is in the field, and the text Tab last set (any other
// edit ends the run).
type tabCycle struct {
	cands []string
	idx   int
	set   string
}

func newCommandInput(comp *completer) *commandInput {
	c := &commandInput{nativeInput: newNativeInput(), comp: comp}
	c.in.Ghost = comp.ghost
	return c
}

// Ghost is the completion suffix currently shown after the text.
func (c *commandInput) Ghost() string { return c.comp.ghost(c.GetText()) }

// acceptKey handles the completion keys: Tab and Backtab take and cycle the
// candidates, → takes the ghost when the cursor is at the end. It reports
// whether the key was used.
func (c *commandInput) acceptKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyTab, tcell.KeyBacktab:
		step := 1
		if ev.Key() == tcell.KeyBacktab {
			step = -1
		}
		c.tab(step)
		return true // Tab never leaves the prompt
	case tcell.KeyRight:
		if !c.in.AtEnd() || !c.HasFocus() {
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
		c.setText(line)
		return true
	}
	return false
}

// tab takes the first candidate, or on a repeat press the next one (step
// -1: the previous one). A run with one candidate starts over from what is
// in the field, so Tab after `plan build ` goes on to the building.
func (c *commandInput) tab(step int) {
	text := c.GetText()
	if text == c.cycle.set && len(c.cycle.cands) > 1 {
		n := len(c.cycle.cands)
		c.cycle.idx = ((c.cycle.idx+step)%n + n) % n
		c.setText(c.cycle.cands[c.cycle.idx])
		return
	}
	cands := c.comp.candidates(text)
	if len(cands) == 0 {
		c.cycle = tabCycle{}
		return
	}
	idx := 0
	if step < 0 {
		idx = len(cands) - 1
	}
	c.cycle = tabCycle{cands: cands, idx: idx}
	c.setText(cands[idx])
}

func (c *commandInput) setText(s string) {
	c.SetText(s)
	c.cycle.set = s
}
