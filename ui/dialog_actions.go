package ui

import (
	"strings"

	"github.com/espresso20/ageforge/theme"
)

// dialog_actions.go is what the dialogs that are built from rows share to
// answer with action buttons: reading the keys a line of text named, and a row
// of buttons.

// splitKeyed reads a label that names its key in brackets ("[E] ENDURE") into
// the key and the label.
func splitKeyed(s string) (key, label string) {
	if i := strings.Index(s, "]"); strings.HasPrefix(s, "[") && i > 1 {
		if k := strings.TrimSuffix(s[1:i], "["); len([]rune(k)) == 1 {
			return k, strings.TrimSpace(s[i+1:])
		}
	}
	return "", s
}

// keyHints reads a line of hints written as keys and their words ("Enter:
// confirm · Esc: cancel", in any colour tags) into key and word pairs, and
// what follows the pairs, which is not a key.
func keyHints(s string) (hints [][2]string, rest string) {
	for _, tag := range []string{"[gray]", "[-]"} {
		s = strings.ReplaceAll(s, tag, "")
	}
	parts := strings.Split(s, "·")
	for i, part := range parts {
		k, w, ok := strings.Cut(strings.TrimSpace(part), ": ")
		if !ok || strings.Contains(k, " ") {
			rest = strings.TrimSpace(strings.Join(parts[i:], "·"))
			return
		}
		hints = append(hints, [2]string{k, w})
	}
	return
}

// choice is a button of a dialog's row: what it says, the key it names, its
// colour and sign, and what it does.
type choice struct {
	label, key string
	kind       keyKind
	sign       signKind
	off        bool
	press      func()
}

// choiceRow is the buttons of cs in a row, centred, three rows tall, on the
// ground of the box they are in; the keyboard starts on the button focus.
func choiceRow(cs []choice, gap, focus int) (*nativeFlex, []*nativeButton) {
	row := newNativeFlex().SetDirection(flexColumn).AddItem(surfaceBox(), 0, 1, false)
	var out []*nativeButton
	for i, c := range cs {
		b := newNativeButton(c.label).asAction(c.kind, theme.Of(theme.RoleSurface), c.key, c.sign)
		b.off = c.off
		if c.press != nil && !c.off {
			b.SetSelectedFunc(c.press)
		}
		if i > 0 {
			row.AddItem(surfaceBox(), gap, 0, false)
		}
		row.AddItem(b, b.actionWidth(), 0, i == focus)
		out = append(out, b)
	}
	row.AddItem(surfaceBox(), 0, 1, false)
	return row, out
}
