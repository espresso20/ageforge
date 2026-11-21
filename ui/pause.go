package ui

import (
	"image"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// pause.go is the pause card: a card over the dimmed town, opened by Esc with
// no panel open. While it is open the engine is held (game/hold.go), the
// pointer is the card's over the whole page and the keys are its own, so
// nothing behind it moves, is clicked or is typed into. The Command box stays
// as it is: the card and the dimming cover the body only. Settings are three
// tabs of the same card, each control on a setting that already exists.

const (
	pausePage  = "__pause__"
	pauseCardW = 72
	pauseCardH = 27
)

type pauseView int

const (
	pauseMain pauseView = iota
	pauseDisplay
	pauseMotion
	pauseKeysTab
)

// pauseRow is a setting: its label, its value, and the buttons that change it.
type pauseRow struct {
	label func() string
	btns  []*nativeButton
}

type pauseLayer struct {
	*nativeHost
	kit.Base
	d     *Dashboard
	view  pauseView
	btns  []*nativeButton // the ring: what Tab and the arrows walk
	at    int
	rows  []pauseRow
	acts  []*nativeButton // the four buttons of the card
	note  string
	rect  image.Rectangle
	box   image.Rectangle
	press map[rune]func() // the letters of the card
}

// openPause shows the card and holds the engine.
func (d *Dashboard) openPause() *pauseLayer {
	p := &pauseLayer{d: d}
	p.nativeHost = newNativeHost()
	d.engine.Hold()
	p.build()
	d.pages.AddPage(pausePage, p, true, true)
	d.app.SetFocus(p.btns[0])
	return p
}

// closePause takes the card away and lets the engine go on.
func (d *Dashboard) closePause() {
	d.pages.RemovePage(pausePage)
	d.engine.Release()
	d.app.SetFocus(d.inputField)
}

// toMenu saves and leaves for the main menu, as Esc did before the card.
func (d *Dashboard) toMenu() {
	d.closePause()
	d.engine.SaveGame(d.engine.ActiveSaveName())
	d.engine.Stop()
	d.pages.SwitchToPage("splash")
}

// build makes the buttons of the view the card is on.
func (p *pauseLayer) build() {
	p.btns, p.rows, p.acts, p.press = nil, nil, nil, map[rune]func(){}
	ground := kit.Role(theme.RoleBackground)
	if p.view == pauseMain {
		for _, a := range []struct {
			label, key string
			kind       keyKind
			do         func()
		}{
			{rowid.Line("ui.pause.build.01", "Resume"), "Esc", keyAccent, p.d.closePause},
			{rowid.Line("ui.pause.build.02", "Settings"), "S", keySteel, func() { p.go2(pauseDisplay) }},
			{rowid.Line("ui.pause.build.03", "Save"), "V", keySteel, p.save},
			{rowid.Line("ui.pause.build.04", "Main menu"), "M", keySteel, func() { p.d.pressLine("menu") }},
		} {
			b := newNativeButton(a.label).SetSelectedFunc(a.do).asAction(a.kind, ground, a.key)
			p.acts, p.btns = append(p.acts, b), append(p.btns, b)
			if a.key != "Esc" {
				p.press[rune(strings.ToLower(a.key)[0])] = a.do
			}
		}
		return
	}
	for i, t := range []string{rowid.Line("ui.pause.build.05", "Display"), rowid.Line("ui.pause.build.06", "Motion"), rowid.Line("ui.pause.build.07", "Keys")} {
		b := newNativeButton(t).SetSelectedFunc(func() { p.go2(pauseView(i + 1)) })
		if pauseView(i+1) == p.view {
			b.SetLabelColor(kit.Role(theme.RoleAccent))
		}
		p.btns = append(p.btns, b)
	}
	p.at = int(p.view) - 1
	switch p.view {
	case pauseDisplay:
		p.display()
	case pauseMotion:
		p.motion()
	}
}

// go2 switches to a view, with the keyboard on its first button.
func (p *pauseLayer) go2(v pauseView) {
	p.view, p.note = v, ""
	p.build()
	p.refocus()
}

// refocus relays the card out and puts the keyboard on the button at p.at.
func (p *pauseLayer) refocus() {
	p.Layout(p.rect)
	p.at = min(p.at, len(p.btns)-1)
	p.d.app.SetFocus(p.btns[p.at])
}

// row adds a setting; each button runs its function, and the card is built
// again so every value reads as it is now.
func (p *pauseLayer) row(label func() string, btns ...[2]any) {
	r := pauseRow{label: label}
	for _, bt := range btns {
		do := bt[1].(func())
		b := newNativeButton(bt[0].(string)).SetSelectedFunc(func() {
			at := slices.Index(p.btns, p.focused())
			do()
			p.build()
			p.at = at
			p.refocus()
		})
		r.btns, p.btns = append(r.btns, b), append(p.btns, b)
	}
	p.rows = append(p.rows, r)
}

func (p *pauseLayer) focused() *nativeButton {
	for _, b := range p.btns {
		if b.HasFocus() {
			return b
		}
	}
	return nil
}

// display wires the text size, the window mode, the font and the theme.
func (p *pauseLayer) display() {
	d := p.d
	if w := d.window; w != nil {
		p.row(func() string { return rowid.Sprintf("ui.pause.display.01", "Text size  %d", w.TextSize()) },
			[2]any{"-", func() { w.ZoomBy(-1) }}, [2]any{"+", func() { w.ZoomBy(1) }})
		mode := rowid.Line("ui.pause.display.02", "Windowed")
		if w.Fullscreen() {
			mode = rowid.Line("ui.pause.display.03", "Fullscreen")
		}
		p.row(func() string { return rowid.Line("ui.pause.display.04", "Window") }, [2]any{mode, w.ToggleFullscreen})
		p.row(func() string { return rowid.Line("ui.pause.display.05", "Font") }, [2]any{w.FontChoice(), w.NextFont})
	}
	p.row(func() string { return rowid.Line("ui.pause.display.06", "Theme") }, [2]any{theme.Active().Name.String(), d.nextTheme})
}

// nextTheme applies the next theme the account may use, as the theme command does.
func (d *Dashboard) nextTheme() {
	acct := themeAccount(d.engine)
	var open []theme.Theme
	for _, t := range theme.All() {
		if themeAvailable(acct, t) {
			open = append(open, t)
		}
	}
	for i, t := range open {
		if t.Key == theme.Active().Key {
			cmdTheme([]string{open[(i+1)%len(open)].Key}, d.engine)
		}
	}
}

// motion wires the motion, mini map and map glyph settings.
func (p *pauseLayer) motion() {
	d := p.d
	s := d.mapSettings
	set := func(f func(*mapSettings)) func() {
		return func() { v := s(); f(&v); d.saveMapSettings(v) }
	}
	p.row(func() string { return rowid.Line("ui.pause.motion.01", "Motion") },
		[2]any{onOff(s().Motion), set(func(v *mapSettings) { v.Motion = !v.Motion })})
	p.row(func() string { return rowid.Line("ui.pause.motion.02", "Minimap") },
		[2]any{onOff(s().Minimap), set(func(v *mapSettings) { v.Minimap = !v.Minimap })})
	p.row(func() string { return rowid.Line("ui.pause.motion.03", "Map glyphs") },
		[2]any{s().Tier.String(), set(func(v *mapSettings) {
			v.Tier = mapmodel.TierUnicode
			if s().Tier == mapmodel.TierUnicode {
				v.Tier = mapmodel.TierASCII
			}
		})})
}

// save saves now and says so on the card.
func (p *pauseLayer) save() {
	if p.d.engine.SaveGame(p.d.engine.ActiveSaveName()) == nil {
		p.note = rowid.Line("ui.pause.save.01", "Saved")
	}
}

func (p *pauseLayer) Children() []kit.Widget {
	out := make([]kit.Widget, len(p.btns))
	for i, b := range p.btns {
		out[i] = b
	}
	return out
}

// Layout centres the card in the body, above the Command box, and places the
// buttons of the view.
func (p *pauseLayer) Layout(r image.Rectangle) {
	p.Place(r)
	if p.nativeHost == nil {
		return
	}
	p.rect = r
	off := screenOffset(p, r)
	w, h := min(pauseCardW, r.Dx()-2), min(pauseCardH, r.Dy()-p.d.bar()-2)
	x, y := r.Min.X+(r.Dx()-w)/2, r.Min.Y+(r.Dy()-p.d.bar()-h)/2
	p.box = image.Rect(x, y, x+w, y+h)
	x0 := x + 3
	if p.view == pauseMain {
		bw := 0
		for _, b := range p.acts {
			bw = max(bw, b.actionWidth())
		}
		for i, b := range p.acts {
			placeAt(b, x0, y+8+4*i, bw, chipRows, off)
		}
		return
	}
	tx := x0
	for i := 0; i < 3; i++ {
		bw := kit.Width(p.btns[i].GetLabel()) + 4
		placeAt(p.btns[i], tx, y+7, bw, 1, off)
		tx += bw + 1
	}
	for i, row := range p.rows {
		bx := x0 + 26
		for _, b := range row.btns {
			bw := max(kit.Width(b.GetLabel())+4, 5)
			placeAt(b, bx, y+9+2*i, bw, 1, off)
			bx += bw + 1
		}
	}
}

// pauseKeyList is the Keys tab: a key and what it does, in two columns.
func pauseKeyList() [][2]string {
	return [][2]string{
		{"Esc", rowid.Line("ui.pause.keys.01", "Pause or close panel")},
		{"Enter", rowid.Line("ui.pause.keys.02", "Run the command")},
		{"Tab", rowid.Line("ui.pause.keys.03", "Next step")},
		{"PgUp/PgDn", rowid.Line("ui.pause.keys.04", "Scroll Buildings")},
		{resourcePageKeyName, rowid.Line("ui.pause.keys.05", "Next Resources page")},
		{keyF11, rowid.Line("ui.pause.keys.06", "Fullscreen")},
		{zoomKeysHint(), rowid.Line("ui.pause.keys.07", "Text size")},
		{"S  V  M", rowid.Line("ui.pause.keys.08", "Settings, Save, Menu")},
	}
}

func (p *pauseLayer) Paint(pt *kit.Painter) {
	if p.box.Empty() {
		return
	}
	bg := kit.Role(theme.RoleBackground)
	pt.Shade(image.Rect(p.rect.Min.X, p.rect.Min.Y, p.rect.Max.X, p.rect.Max.Y-p.d.bar()), bg, .65)
	pt.Fill(p.box, bg)
	pt.Box(p.box, kit.BorderLine, kit.Style{Fg: kit.Role(theme.RoleAccent)})
	x0, y := p.box.Min.X+3, p.box.Min.Y
	paintBig(pt, x0, y+2, "PAUSED", 6, 4, pixelInk(kit.Role(theme.RoleAccent)))
	sty := func(r theme.Role) kit.Style { return kit.Style{Fg: kit.Role(r), Bg: bg} }
	dim, text := sty(theme.RoleDim), sty(theme.RoleText)
	foot := rowid.Line("ui.pause.paint.01", "Leaving saves your game first.")
	if p.view == pauseMain {
		x1 := x0 + 30
		pt.Print(x1, y+8, p.box.Max.X-x1-2, rowid.Line("ui.pause.paint.02", "This sitting"), sty(theme.RoleAccent))
		st := p.d.engine.GetState()
		saved := rowid.Line("ui.pause.paint.07", "not yet")
		if t := p.d.engine.SavedAt(); !t.IsZero() {
			saved = relativeTime(t)
		}
		for i, kv := range [][2]string{
			{rowid.Line("ui.pause.paint.03", "Age"), ageDisplay(st.Age)},
			{rowid.Line("ui.pause.paint.04", "Played"), textfmt.Duration(st.Stats.PlayTime)},
			{rowid.Line("ui.pause.paint.05", "Last saved"), saved},
			{rowid.Line("ui.pause.paint.06", "Next autosave"), textfmt.Duration(p.d.engine.NextAutosave())},
		} {
			pt.Print(x1, y+10+2*i, 14, kv[0], dim)
			pt.Print(x1+15, y+10+2*i, p.box.Max.X-x1-17, kv[1], text)
		}
		pt.Print(x1, y+19, 12, p.note, sty(theme.RolePositive))
	} else {
		foot = rowid.Line("ui.pause.paint.08", "Esc goes back to the pause card.")
		for i, row := range p.rows {
			pt.Print(x0, y+9+2*i, 24, row.label(), text)
		}
		if p.view == pauseKeysTab {
			for i, k := range pauseKeyList() {
				cx, cy := x0+(i/4)*33, y+9+2*(i%4)
				pt.Print(cx, cy, 9, k[0], sty(theme.RoleHighlight))
				pt.Print(cx+10, cy, 22, k[1], text)
			}
		}
	}
	pt.PrintAligned(p.box.Min.X+1, p.box.Max.Y-3, p.box.Dx()-2, foot, dim, kit.AlignCenter)
	for _, b := range p.btns {
		if !b.Bounds().Empty() {
			b.ground = bg
			b.Paint(pt)
		}
	}
}

// Covers: the card is the pointer's over the whole page, so a click on the
// dimmed town (or the Command box) does nothing.
func (p *pauseLayer) Covers(image.Point) bool { return true }

func (p *pauseLayer) covered() []image.Rectangle { return []image.Rectangle{p.box} }

// Key: Esc steps back (from a tab to the card, from the card to the game),
// the letters press the card's buttons, and Tab and the arrows walk the ring.
func (p *pauseLayer) Key(k kit.Key) bool {
	switch k.Code {
	case kit.KeyEsc:
		if p.view != pauseMain {
			p.at = 1
			p.go2(pauseMain)
		} else {
			p.d.closePause()
		}
	case kit.KeyTab, kit.KeyDown, kit.KeyRight, kit.KeyBacktab, kit.KeyUp, kit.KeyLeft:
		by := len(p.btns) + 1 // forward; back is one short of a lap
		if k.Code == kit.KeyBacktab || k.Code == kit.KeyUp || k.Code == kit.KeyLeft {
			by = len(p.btns) - 1
		}
		p.at = (p.at + by) % len(p.btns)
		p.d.app.SetFocus(p.btns[p.at])
	case kit.KeyRune:
		if do, ok := p.press[k.Rune|0x20]; ok {
			do()
		}
	}
	return true
}

// Focus gives the keyboard to the button it was last on.
func (p *pauseLayer) Focus(delegate func(p Primitive)) {
	if len(p.btns) == 0 {
		p.primBase.Focus(delegate)
		return
	}
	delegate(p.btns[min(p.at, len(p.btns)-1)])
}

// HasFocus reports whether the card or one of its buttons has the keyboard.
func (p *pauseLayer) HasFocus() bool { return p.focused() != nil || p.primBase.HasFocus() }
