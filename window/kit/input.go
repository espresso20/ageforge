package kit

import (
	"image"
	"unicode"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/theme"
)

// Input is one line the player types into: a prompt, the text, a cursor. It
// edits (the arrows, Home, End, Backspace, Delete, and Ctrl with A, E, U, K
// and W as a shell has them) and calls OnEnter on Enter. Every other key is
// left for whatever is around it, which is how a command line gets its
// history on Up and its completion on Tab.
type Input struct {
	Base
	// Prompt is tagged text before the line, painted in PromptStyle (unset:
	// the theme's highlight color).
	Prompt      string
	PromptStyle Style
	// Style is what the typed text is painted in; its ground fills the
	// field, the cells after the prompt.
	Style Style
	// Width is the field's width in cells; 0 takes the rest of the row.
	Width int
	// Mask, when set, is painted in place of every typed character (a
	// password box).
	Mask rune
	// Ghost, when set, is asked for the text to show dim after the line
	// while the cursor is at its end: a completion on offer.
	Ghost func(text string) string
	// OnChange is called when the player changes the text, OnEnter with the
	// text on Enter.
	OnChange func(text string)
	OnEnter  func(text string)

	text []rune
	cur  int // the cursor, as an index into text
	off  int // the first rune in view: a long line scrolls
}

// NewInput is an empty line after a prompt.
func NewInput(prompt string) *Input { return &Input{Prompt: prompt} }

// Text is what has been typed.
func (in *Input) Text() string { return string(in.text) }

// SetText replaces the text and puts the cursor at its end. It is not the
// player's change, so OnChange is not called.
func (in *Input) SetText(s string) {
	in.text = []rune(s)
	in.cur, in.off = len(in.text), 0
}

// AtEnd reports whether the cursor is after the last character, where a
// ghost shows.
func (in *Input) AtEnd() bool { return in.cur == len(in.text) }

func (in *Input) CanFocus() bool { return true }

func (in *Input) Layout(r image.Rectangle) { in.Place(r) }

// Size is the whole width, one row high.
func (in *Input) Size(most image.Point) image.Point { return image.Pt(most.X, min(1, most.Y)) }

func runeWidth(r rune) int { return uniseg.StringWidth(string(r)) }

func runesWidth(rs []rune) (w int) {
	for _, r := range rs {
		w += runeWidth(r)
	}
	return w
}

// field is where the text starts and how many cells it has.
func (in *Input) field() (x, w int) {
	pw := Width(in.Prompt)
	x, w = in.rect.Min.X+pw, in.rect.Dx()-pw
	if in.Width > 0 {
		w = min(w, in.Width)
	}
	return x, w
}

func (in *Input) Paint(p *Painter) {
	r := in.rect
	if r.Empty() {
		return
	}
	p = p.Clip(r)
	ps := in.PromptStyle
	if !ps.Fg.IsSet() {
		ps.Fg = Role(theme.RoleHighlight)
	}
	p.Print(r.Min.X, r.Min.Y, r.Dx(), in.Prompt, ps)
	x, w := in.field()
	if w <= 0 {
		return
	}
	p.Fill(image.Rect(x, r.Min.Y, x+w, r.Min.Y+1), in.Style.Bg)
	// Scroll so the cursor has a cell in view.
	in.off = min(in.off, in.cur)
	for in.off < in.cur && runesWidth(in.text[in.off:in.cur]) >= w {
		in.off++
	}
	at := x
	for i := in.off; i < len(in.text) && at+runeWidth(in.text[i]) <= x+w; i++ {
		ch := in.text[i]
		if in.Mask != 0 {
			ch = in.Mask
		}
		p.Put(at, r.Min.Y, ch, in.Style)
		if rw := runeWidth(ch); rw == 2 {
			p.put(at, r.Min.Y, &glyph{s: string(ch), w: 2}, in.Style)
		}
		at += runeWidth(ch)
	}
	curX := x + runesWidth(in.text[in.off:in.cur])
	if in.AtEnd() && in.Ghost != nil {
		p.Print(curX, r.Min.Y, x+w-curX, Escape(in.Ghost(string(in.text))), Style{Fg: Role(theme.RoleDim), Bg: in.Style.Bg})
	}
	if in.focused {
		p.Cursor(curX, r.Min.Y)
	}
}

// edit replaces text[from:to] with ins, leaves the cursor after it and tells
// OnChange.
func (in *Input) edit(from, to int, ins ...rune) {
	in.text = append(in.text[:from:from], append(ins, in.text[to:]...)...)
	in.cur = from + len(ins)
	if in.OnChange != nil {
		in.OnChange(string(in.text))
	}
}

// wordStart is where the word before the cursor starts.
func (in *Input) wordStart() int {
	i := in.cur
	for i > 0 && unicode.IsSpace(in.text[i-1]) {
		i--
	}
	for i > 0 && !unicode.IsSpace(in.text[i-1]) {
		i--
	}
	return i
}

func (in *Input) Key(k Key) bool {
	n := len(in.text)
	if k.Code == KeyRune && k.Mod&(ModCtrl|ModAlt) == 0 {
		in.edit(in.cur, in.cur, k.Rune)
		return true
	}
	if k.Code == KeyRune && k.Mod&ModCtrl != 0 {
		switch k.Rune {
		case 'a':
			in.cur = 0
		case 'e':
			in.cur = n
		case 'u':
			in.edit(0, in.cur)
		case 'k':
			in.edit(in.cur, n)
		case 'w':
			in.edit(in.wordStart(), in.cur)
		default:
			return false
		}
		return true
	}
	switch k.Code {
	case KeyLeft:
		in.cur = max(in.cur-1, 0)
	case KeyRight:
		if in.cur == n {
			return false // at the end, Right is the owner's: it may take the ghost
		}
		in.cur++
	case KeyHome:
		in.cur = 0
	case KeyEnd:
		in.cur = n
	case KeyBackspace:
		if in.cur > 0 {
			in.edit(in.cur-1, in.cur)
		}
	case KeyDelete:
		if in.cur < n {
			in.edit(in.cur, in.cur+1)
		}
	case KeyEnter:
		if in.OnEnter != nil {
			in.OnEnter(string(in.text))
		}
	default:
		return false
	}
	return true
}

// Mouse: a click puts the cursor at the character clicked.
func (in *Input) Mouse(m Mouse) bool {
	if m.Kind != MousePress || m.Button != ButtonLeft {
		return false
	}
	x, _ := in.field()
	in.cur = in.off
	for in.cur < len(in.text) && x+runeWidth(in.text[in.cur]) <= m.X {
		x += runeWidth(in.text[in.cur])
		in.cur++
	}
	return true
}
