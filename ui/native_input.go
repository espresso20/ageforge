package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeInput stands where a tview.InputField stood: the same calls, drawn
// by the kit (a kit.Input in a kit.Panel on a host). A click puts the cursor
// where it lands. Like nativeText it goes when the screen around it is a kit
// tree.
type nativeInput struct {
	*nativeHost
	in  *kit.Input
	box *kit.Panel

	border  bool
	done    func(kit.KeyCode)
	changed func(text string)
}

// newNativeInput is an input field as tview makes one: the field on the
// selection's ground, the label dim before it.
func newNativeInput() *nativeInput {
	f := &nativeInput{in: kit.NewInput("")}
	f.in.Style = kit.Style{Fg: kit.Role(theme.RoleText), Bg: kit.Role(theme.RoleSelection)}
	f.in.PromptStyle = kit.Fg(theme.RoleDim)
	f.in.OnChange = func(text string) {
		if f.changed != nil {
			f.changed(text)
		}
	}
	f.box = &kit.Panel{Child: f.in, Ground: kit.Role(theme.RoleBackground), TitleAlign: kit.AlignCenter}
	f.nativeHost = newNativeHost(f.box)
	return f
}

func (f *nativeInput) SetLabel(s string) *nativeInput {
	f.in.Prompt = s
	return f
}

func (f *nativeInput) GetLabel() string { return f.in.Prompt }

// SetFieldWidth sets the field's width in cells; 0 takes the rest of the row.
func (f *nativeInput) SetFieldWidth(n int) *nativeInput {
	f.in.Width = n
	return f
}

func (f *nativeInput) SetText(s string) *nativeInput {
	f.in.SetText(s)
	if f.changed != nil {
		f.changed(s)
	}
	return f
}

func (f *nativeInput) GetText() string { return f.in.Text() }

func (f *nativeInput) SetLabelColor(c tcell.Color) *nativeInput {
	f.in.PromptStyle.Fg = kitColor(c)
	return f
}

func (f *nativeInput) SetFieldTextColor(c tcell.Color) *nativeInput {
	f.in.Style.Fg = kitColor(c)
	return f
}

func (f *nativeInput) SetFieldBackgroundColor(c tcell.Color) *nativeInput {
	f.in.Style.Bg = kitColor(c)
	return f
}

// SetMaskCharacter shows r in place of every typed character.
func (f *nativeInput) SetMaskCharacter(r rune) *nativeInput {
	f.in.Mask = r
	return f
}

func (f *nativeInput) SetBackgroundColor(c tcell.Color) *nativeInput {
	f.box.Ground = kitColor(c)
	return f
}

// SetChangedFunc sets what is called when the text changes.
func (f *nativeInput) SetChangedFunc(fn func(text string)) *nativeInput {
	f.changed = fn
	return f
}

// SetDoneFunc sets what Enter, Esc, Tab and Shift+Tab call.
func (f *nativeInput) SetDoneFunc(fn func(key kit.KeyCode)) *nativeInput {
	f.done = fn
	return f
}

func (f *nativeInput) SetBorder(on bool) *nativeInput {
	f.border = on
	return f
}

func (f *nativeInput) SetTitle(s string) *nativeInput {
	f.box.Title = s
	return f
}

func (f *nativeInput) GetTitle() string { return f.box.Title }

func (f *nativeInput) SetTitleAlign(align int) *nativeInput {
	f.box.TitleAlign = kitAligns[align]
	return f
}

// SetTitleColor is accepted and does nothing: a kit title is the theme's
// accent, which is what every caller sets.
func (f *nativeInput) SetTitleColor(tcell.Color) *nativeInput { return f }

func (f *nativeInput) SetBorderColor(c tcell.Color) *nativeInput {
	f.box.BorderStyle.Fg = kitColor(c)
	return f
}

// GetInnerRect is the rectangle the field's row is in: the box, less its
// border.
func (f *nativeInput) GetInnerRect() (x, y, w, h int) {
	x, y, w, h = f.GetRect()
	if f.border {
		x, y, w, h = x+1, y+1, w-2, h-2
	}
	return x, y, max(w, 0), max(h, 0)
}

// Draw draws the field, with the doubled border tview gives the primitive
// that has the focus.
func (f *nativeInput) Draw(scr tcell.Screen) {
	f.box.Border = kit.BorderNone
	if f.border {
		f.box.Border = kit.BorderLine
		if f.HasFocus() {
			f.box.Border = kit.BorderDouble
		}
	}
	f.nativeHost.Draw(scr)
}

// InputHandler takes an input field's keys: Enter, Esc, Tab and Shift+Tab
// are done, the rest edit the line.
func (f *nativeInput) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return keysIn(f.Box, func(ev *kit.Key) {
		switch ev.Code {
		case kit.KeyEnter, kit.KeyEsc, kit.KeyTab, kit.KeyBacktab:
			if f.done != nil {
				f.done(ev.Code)
			}
			return
		}
		f.root.Key(*ev)
	})
}
