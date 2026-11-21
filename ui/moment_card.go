package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// momentPage is the page of a moment card: a small window over the game for
// something that should not pass as a line in the log (a return after time
// away, a prestige in reach, the answer to a command that asks for a choice).
const momentPage = "moment"

const momentWidth = 64

// momentBtn is one button of a moment card. Its line is the command it runs,
// so every choice can be typed as well; an empty line only closes the card.
type momentBtn struct {
	label, line string
	kind        keyKind
}

// momentCard is a moment card's words and buttons. The keyboard starts on the
// first button.
type momentCard struct {
	title, body string
	btns        []momentBtn
}

// showMoment opens c over the game, in place of a card already open. Enter or
// a click presses a button and Esc closes the card. A typed character closes
// it too and goes to the command line, so the card never holds up typing.
func (d *Dashboard) showMoment(c momentCard) {
	d.closeMoment()
	var lines []string
	for _, line := range strings.Split(c.body, "\n") {
		for _, l := range wrapWords(strings.TrimSpace(line), momentWidth-4) {
			lines = append(lines, " "+l)
		}
	}
	body := newNativeText().SetDynamicColors(true).SetText(strings.Join(lines, "\n"))

	choices := make([]choice, len(c.btns))
	for i, b := range c.btns {
		choices[i] = choice{label: b.label, kind: b.kind, press: func() {
			d.closeMoment()
			if b.line != "" {
				d.pressLine(b.line)
			}
		}}
	}
	btnRow, btns := choiceRow(choices, 2, 0)
	inner := newNativeFlex().SetDirection(flexRow).
		AddItem(body, len(lines), 0, false).
		AddItem(newNativeBox(), 1, 0, false).
		AddItem(btnRow, chipRows, 0, true)
	framed(inner, " "+c.title+" ", theme.RoleAccent)
	modal := floatingModal(inner, momentWidth, len(lines)+3+chipRows)

	ring := newButtonRing(d.app, func(p Primitive) { d.momentFocus = p })
	for _, b := range btns {
		ring.add(b)
	}
	modal.SetInputCapture(func(ev *kit.Key) *kit.Key {
		switch {
		case ev.Code == kit.KeyEsc:
			d.closeMoment()
		case ev.Code == kit.KeyRune && ev.Rune != ' ':
			d.closeMoment()
			d.inputField.SetText(d.inputField.GetText() + string(ev.Rune))
		case !ring.key(ev):
			return ev
		}
		return nil
	})
	d.pages.AddPage(momentPage, modal, true, true)
	ring.focus(0)
}

// closeMoment removes the moment card, if one is open, and hands the keyboard
// back.
func (d *Dashboard) closeMoment() {
	if !d.pages.HasPage(momentPage) {
		return
	}
	d.pages.RemovePage(momentPage)
	d.momentFocus = nil
	d.returnFocus()
}

// showCard opens the card a command's reply asks for, with the reply as its
// words.
func (d *Dashboard) showCard(res CommandResult) {
	c := *res.Card
	c.body = res.Message
	d.showMoment(c)
}

// momentClear reports whether a card may open unasked: nothing else is over
// the game and no command is half typed.
func (d *Dashboard) momentClear() bool {
	return !d.overlayMgr.HasActive() && !d.overlayMgr.FullScreenUp() &&
		!d.pages.HasPage(catastrophePage) && !d.pages.HasPage(ancientMemoryPage) &&
		!d.pages.HasPage(momentPage) && !d.pages.HasPage(arrivalPageName) &&
		d.inputField.GetText() == ""
}

// maybeShowReturn opens the report of a long time away, once.
func (d *Dashboard) maybeShowReturn(state game.GameState) {
	if len(state.Return.Lines) == 0 || state.Return.Seen || !d.momentClear() {
		return
	}
	d.showCard(returnReply(state))
	d.engine.MarkReturnSeen()
}
