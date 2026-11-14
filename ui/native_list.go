package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeList stands where a tview.List with one line per item stood: the
// same calls and keys, drawn by the kit. A click selects a row and chooses it,
// as Enter does. Not carried over: the secondary line, the "(x)" shortcut
// labels and the sideways scroll of Left and Right.
type nativeList struct {
	*nativeHost
	list *kit.List
	box  *kit.Panel

	items    []nativeListItem
	border   bool
	title    string
	wrap     bool
	selected func(i int, main, secondary string, shortcut rune)
	changed  func(i int, main, secondary string, shortcut rune)
}

type nativeListItem struct {
	main, secondary string
	shortcut        rune
	selected        func()
}

// newNativeList is a list as tview makes one: text on the page's ground, the
// selection inverted, and a move past either end coming round.
func newNativeList() *nativeList {
	l := &nativeList{list: kit.NewList(), wrap: true}
	l.list.Bar, l.list.ClickChooses, l.list.TextOnly = false, true, true
	l.list.Style = styleOf(tview.Styles.PrimaryTextColor, tview.Styles.PrimitiveBackgroundColor)
	l.list.SelectedStyle = styleOf(tview.Styles.PrimitiveBackgroundColor, tview.Styles.PrimaryTextColor)
	l.list.OnMove = func(i int) { l.fire(l.changed, i) }
	l.list.OnChoose = l.choose
	l.box = &kit.Panel{Child: l.list, Ground: kit.Role(theme.RoleBackground), TitleAlign: kit.AlignCenter}
	l.nativeHost = newNativeHost(l.box)
	return l
}

func (l *nativeList) fire(fn func(int, string, string, rune), i int) {
	if fn != nil && i >= 0 && i < len(l.items) {
		it := l.items[i]
		fn(i, it.main, it.secondary, it.shortcut)
	}
}

// choose is what Enter, a shortcut or a click does to row i.
func (l *nativeList) choose(i int) {
	if i < 0 || i >= len(l.items) {
		return
	}
	if fn := l.items[i].selected; fn != nil {
		fn()
	}
	l.fire(l.selected, i)
}

func (l *nativeList) sync() {
	rows := make([]string, len(l.items))
	for i, it := range l.items {
		rows[i] = it.main
	}
	l.list.SetItems(rows)
}

// AddItem adds a row; the secondary text is kept for the callbacks only.
func (l *nativeList) AddItem(main, secondary string, shortcut rune, selected func()) *nativeList {
	l.items = append(l.items, nativeListItem{main, secondary, shortcut, selected})
	l.sync()
	if len(l.items) == 1 {
		l.fire(l.changed, 0)
	}
	return l
}

func (l *nativeList) Clear() *nativeList {
	l.items = nil
	l.sync()
	return l
}

func (l *nativeList) GetItemCount() int { return len(l.items) }

// GetItemText is a row's main and secondary text.
func (l *nativeList) GetItemText(i int) (main, secondary string) {
	if i < 0 || i >= len(l.items) {
		return "", ""
	}
	return l.items[i].main, l.items[i].secondary
}

func (l *nativeList) GetCurrentItem() int { return max(l.list.Selected(), 0) }

// SetCurrentItem selects row i; a negative i counts from the end.
func (l *nativeList) SetCurrentItem(i int) *nativeList {
	if i < 0 {
		i += len(l.items)
	}
	l.list.Select(i)
	return l
}

func (l *nativeList) SetSelectedFunc(fn func(int, string, string, rune)) *nativeList {
	l.selected = fn
	return l
}

func (l *nativeList) SetChangedFunc(fn func(int, string, string, rune)) *nativeList {
	l.changed = fn
	return l
}

func (l *nativeList) SetMainTextColor(c tcell.Color) *nativeList {
	l.list.Style.Fg = kitColor(c)
	return l
}

func (l *nativeList) SetSelectedTextColor(c tcell.Color) *nativeList {
	l.list.SelectedStyle.Fg = kitColor(c)
	return l
}

func (l *nativeList) SetSelectedBackgroundColor(c tcell.Color) *nativeList {
	l.list.SelectedStyle.Bg = kitColor(c)
	return l
}

// ShowSecondaryText is accepted; the list always shows one line a row.
func (l *nativeList) ShowSecondaryText(bool) *nativeList { return l }

// SetHighlightFullLine colors the selected row across, not just its text.
func (l *nativeList) SetHighlightFullLine(on bool) *nativeList {
	l.list.TextOnly = !on
	return l
}

func (l *nativeList) SetBorder(on bool) *nativeList {
	l.border = on
	return l
}

func (l *nativeList) SetTitle(s string) *nativeList {
	l.title = s
	return l
}

func (l *nativeList) SetBorderColor(c tcell.Color) *nativeList {
	l.box.BorderStyle.Fg = kitColor(c)
	return l
}

func (l *nativeList) SetBackgroundColor(c tcell.Color) *nativeList {
	l.box.Ground = kitColor(c)
	return l
}

// GetInnerRect is the rectangle the rows are in: the box, less its border.
func (l *nativeList) GetInnerRect() (x, y, w, h int) {
	x, y, w, h = l.GetRect()
	if l.border {
		x, y, w, h = x+1, y+1, w-2, h-2
	}
	return x, y, max(w, 0), max(h, 0)
}

// Draw draws the list, with the doubled border tview gives the primitive that
// has the focus.
func (l *nativeList) Draw(scr tcell.Screen) {
	l.box.Border, l.box.Title = kit.BorderNone, ""
	if l.border {
		l.box.Title = l.title
		l.box.Border = kit.BorderLine
		if l.HasFocus() {
			l.box.Border = kit.BorderDouble
		}
	}
	l.nativeHost.Draw(scr)
}

// InputHandler takes a list's keys as tview's does: Up and Down (and Shift+Tab
// and Tab) move, coming round at the ends, Home, End, PgUp and PgDn jump, Enter
// chooses, and a row's shortcut selects and chooses it. Space chooses the
// selected row.
func (l *nativeList) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return keysIn(l.Box, func(ev *kit.Key) {
		n := len(l.items)
		if n == 0 {
			return
		}
		cur := l.GetCurrentItem()
		to := cur
		switch ev.Code {
		case kit.KeyTab, kit.KeyDown:
			to++
		case kit.KeyBacktab, kit.KeyUp:
			to--
		case kit.KeyHome:
			to = 0
		case kit.KeyEnd:
			to = n - 1
		case kit.KeyPgDn:
			_, _, _, h := l.GetInnerRect()
			to = min(cur+h, n-1)
		case kit.KeyPgUp:
			_, _, _, h := l.GetInnerRect()
			to = max(cur-h, 0)
		case kit.KeyEnter:
			l.choose(cur)
			return
		case kit.KeyRune:
			if r := ev.Rune; r != ' ' {
				found := false
				for i, it := range l.items {
					if it.shortcut == r {
						to, found = i, true
						break
					}
				}
				if !found {
					return
				}
			}
			l.list.Select(to)
			l.choose(l.GetCurrentItem())
			return
		default:
			return
		}
		if to < 0 || to >= n {
			if l.wrap {
				to = (to + n) % n
			} else {
				to = max(0, min(to, n-1))
			}
		}
		l.list.Select(to)
	})
}
