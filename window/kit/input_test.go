package kit

import (
	"testing"

	"github.com/espresso20/ageforge/theme"
)

func typeText(r *Root, s string) {
	for _, c := range s {
		r.Key(Key{Code: KeyRune, Rune: c})
	}
}

func ctrl(c rune) Key { return Key{Code: KeyChord, Rune: c, Mod: ModCtrl} }

func TestInputEditsALine(t *testing.T) {
	var changes int
	var entered string
	in := NewInput("[gold]>[-] ")
	in.OnChange = func(string) { changes++ }
	in.OnEnter = func(s string) { entered = s }
	r, b := screen(in, 20, 1)
	typeText(r, "build mill")
	if in.Text() != "build mill" || changes != 10 {
		t.Fatalf("typed %q with %d changes", in.Text(), changes)
	}
	r.Paint(b)
	if got := b.Line(0); got != "> build mill" {
		t.Errorf("painted %q", got)
	}
	if !b.CursorOn || b.Cursor.X != 12 {
		t.Errorf("the cursor should sit after the text, at %v (on %v)", b.Cursor, b.CursorOn)
	}
	if b.At(0, 0).Fg != theme.RGBA(theme.RoleAccent) {
		t.Error("the prompt keeps its tags")
	}
	for _, k := range []Key{key(KeyHome), key(KeyDelete), key(KeyEnd), key(KeyBackspace), key(KeyLeft), key(KeyLeft)} {
		r.Key(k)
	}
	r.Key(Key{Code: KeyRune, Rune: 'X'})
	if got := in.Text(); got != "uild mXil" {
		t.Errorf("after the edits: %q", got)
	}
	r.Key(ctrl('w'))
	if got := in.Text(); got != "uild il" {
		t.Errorf("Ctrl+W takes the word before the cursor: %q", got)
	}
	r.Key(ctrl('k'))
	r.Key(ctrl('a'))
	typeText(r, "b")
	if got := in.Text(); got != "build " {
		t.Errorf("Ctrl+K then Ctrl+A and a letter: %q", got)
	}
	r.Key(key(KeyEnter))
	if entered != "build " {
		t.Errorf("Enter gave %q", entered)
	}
	r.Key(ctrl('e'))
	r.Key(ctrl('u'))
	if in.Text() != "" {
		t.Errorf("Ctrl+U at the end clears the line: %q", in.Text())
	}
	before := changes
	in.SetText("set by the game")
	if changes != before || !in.AtEnd() {
		t.Error("SetText is not the player's change, and leaves the cursor at the end")
	}
}

func TestInputLeavesOtherKeysToItsOwner(t *testing.T) {
	in := NewInput("> ")
	r, _ := screen(in, 20, 1)
	var seen []KeyCode
	r.Keys = func(k Key) bool { seen = append(seen, k.Code); return true }
	typeText(r, "ab")
	for _, c := range []KeyCode{KeyUp, KeyDown, KeyEsc, KeyRight, KeyPgUp} {
		r.Key(key(c))
	}
	if len(seen) != 5 {
		t.Errorf("history, Esc and Right at the end are the owner's keys, it saw %v", seen)
	}
	seen = nil
	r.Key(key(KeyLeft))
	r.Key(key(KeyRight))
	if len(seen) != 0 {
		t.Errorf("inside the line the arrows are the input's: %v", seen)
	}
}

func TestInputShowsAGhostAtTheEnd(t *testing.T) {
	in := NewInput("> ")
	in.Ghost = func(s string) string {
		if s == "bu" {
			return "ild [x]"
		}
		return ""
	}
	r, b := screen(in, 20, 1)
	typeText(r, "bu")
	r.Paint(b)
	if got := b.Line(0); got != "> build [x]" {
		t.Errorf("painted %q", got)
	}
	if b.At(4, 0).Fg != theme.RGBA(theme.RoleDim) || b.At(3, 0).Fg == theme.RGBA(theme.RoleDim) {
		t.Error("the ghost is dim and the typed text is not")
	}
	r.Key(key(KeyLeft))
	r.Paint(b)
	if got := b.Line(0); got != "> bu" {
		t.Errorf("with the cursor inside the line there is no ghost: %q", got)
	}
}

func TestInputScrollsALongLineAndTakesAClick(t *testing.T) {
	in := NewInput("> ")
	r, b := screen(in, 8, 1) // six cells of text
	typeText(r, "abcdefghij")
	r.Paint(b)
	if got := b.Line(0); got != "> fghij" {
		t.Errorf("the end of a long line stays in view: %q", got)
	}
	if b.Cursor.X != 7 {
		t.Errorf("the cursor is at %d", b.Cursor.X)
	}
	r.Key(key(KeyHome))
	r.Paint(b)
	if got := b.Line(0); got != "> abcdef" {
		t.Errorf("Home scrolls back: %q", got)
	}
	r.Mouse(Mouse{Kind: MousePress, X: 5, Y: 0})
	typeText(r, "X")
	if got := in.Text(); got != "abcXdefghij" {
		t.Errorf("a click puts the cursor at the character: %q", got)
	}
}
