package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/cabinet"
	"github.com/espresso20/ageforge/pkg/coins"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// cabinet_panel.go is the coin cabinet: the account's collection as a wall of
// drawers, one per age and then Ledgers, Curios and Themes. The case is a
// picture (pkg/cabinet), drawn once for a layout and what the drawers hold;
// the words on the plates are cell text over it. Like the Map and Research
// panels it leaves the keyboard with the command bar and takes only the keys
// that print nothing: the arrows move across the drawers, Home and End go to
// the first and the last. Esc closes it.

// cabinetKeys are the cabinet's keys as the Help panel lists them.
var cabinetKeys = []keyHelp{
	{"Arrows", rowid.E("ui.cabinet_panel.cabinetKeys.01", "Move from drawer to drawer")},
	{"Home/End", rowid.E("ui.cabinet_panel.cabinetKeys.02", "The first drawer and the last")},
	{"Esc", rowid.E("ui.cabinet_panel.cabinetKeys.03", "Close the cabinet")},
	{"Enter", rowid.E("ui.cabinet_panel.cabinetKeys.04", "Open the drawer; arrows then move between its wells")},
	{"PgUp/PgDn", rowid.E("ui.cabinet_panel.cabinetKeys.06", "Turn the page of Ledgers, Curios or Themes")},
	{"Esc, right-click", rowid.E("ui.cabinet_panel.cabinetKeys.05", "Close the open drawer")},
}

// cabinetAgeless are the drawers after the ages: the key a command names
// them by, and the word on the plate.
var cabinetAgeless = []struct{ key, name string }{
	{"ledgers", rowid.Line("ui.cabinet_panel.cabinetAgeless.01", "Ledgers")},
	{"curios", rowid.Line("ui.cabinet_panel.cabinetAgeless.02", "Curios")},
	{"themes", rowid.Line("ui.cabinet_panel.cabinetAgeless.03", "Themes")},
}

// cabinetPanel is the cabinet's overlay primitive.
type cabinetPanel struct {
	barPanel
	kit.Base
	// account reads the account whose collection is shown; nil means none.
	account func() *game.Account
	mint    *coins.Mint
	set     *rules.Set // the tallies and badges the ageless drawers list

	keys    [cabinet.Slots]string
	names   [cabinet.Slots]string
	dim     [cabinet.Slots]bool
	drawers [cabinet.Slots]cabinet.Drawer
	sel     int

	// The picture is kept while the layout and what the drawers hold stay
	// the same; the selected drawer's patch is kept for one selection.
	drawn    cabinetKey
	front    *image.RGBA
	renders  int
	patchSel int
	patchAt  image.Rectangle
	patchImg *image.RGBA
	tray     cabinetTray
}

// cabinetKey is what the front's picture depends on.
type cabinetKey struct {
	cells, cell image.Point
	ds          [cabinet.Slots]cabinet.Drawer
}

func newCabinetPanel(account func() *game.Account, set *rules.Set) *cabinetPanel {
	p := &cabinetPanel{account: account, set: set, mint: coins.NewMint(256), patchSel: -1, tray: cabinetTray{slot: -1, liftWell: -1}}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *cabinetPanel) open(state game.GameState) Primitive {
	p.update(state)
	return overBar(p)
}

// request selects the drawer a command named by its key and opens it when it
// is an age; a well's key after a space selects that well.
func (p *cabinetPanel) request(key string) {
	key, typ, _ := strings.Cut(key, " ")
	for i, k := range p.keys {
		if k == key && key != "" {
			p.sel = i
			if p.openDrawer(i) && typ != "" {
				p.tray.well = max(cabinetWell(typ), 0)
			}
		}
	}
}

// update reads the account into the drawers.
func (p *cabinetPanel) update(state game.GameState) {
	sight := game.SightOf(&state)
	var acct *game.Account
	if p.account != nil {
		acct = p.account()
	}
	for i, a := range config.Ages() {
		d := cabinet.Drawer{Money: a.Key, Reached: sight.Age(a.Key), Total: len(config.CoinTypes) + 1}
		if acct != nil {
			for _, typ := range config.CoinTypes {
				if b := acct.CoinBest(typ, a.Key); b > 0 {
					d.Best, d.Held = max(d.Best, b), d.Held+1
				}
			}
			if d.Legend = acct.CoinBest(config.CoinFeat, a.Key) > 0; d.Legend {
				d.Held++
			}
		}
		p.keys[i], p.names[i], p.dim[i], p.drawers[i] = a.Key, a.Name, !d.Reached, d
		if !d.Reached {
			p.names[i] = textfmt.Capitalize(sight.AgeRef(a.Key))
		}
	}
	for j, l := range cabinetAgeless {
		i := len(config.Ages()) + j
		d := cabinet.Drawer{Reached: true}
		p.keys[i], p.names[i], p.dim[i] = l.key, l.name, false
		for _, w := range p.agelessWells(i) {
			d.Total++
			if w.held {
				d.Held++
			}
		}
		p.drawers[i] = d
	}
	p.sel = min(p.sel, cabinet.Slots-1)
	p.loadTray()
}

// totals is the line on the cornice: the pieces held and the drawers complete.
func (p *cabinetPanel) totals() string {
	pieces, all, sets, ages := 0, 0, 0, 0
	for _, d := range p.drawers {
		if d.Money == "" {
			continue
		}
		pieces, all, ages = pieces+d.Held, all+d.Total, ages+1
		if d.Reached && d.Held == d.Total {
			sets++
		}
	}
	return fmt.Sprintf(rowid.Line("ui.cabinet_panel.totals.01", "Pieces %d of %d   Sets %d of %d"), pieces, all, sets, ages)
}

// Layout takes the panel's rectangle.
func (p *cabinetPanel) Layout(r image.Rectangle) { p.Place(r) }

// plan is the layout for the panel's size, and whether the cabinet fits.
func (p *cabinetPanel) plan() (cabinet.Layout, bool) {
	cw, ch := cellPixels()
	r := p.Bounds()
	return cabinet.Plan(r.Dx(), r.Dy(), cw, ch)
}

// Paint puts the cabinet's picture, the selected drawer over it, and the words.
func (p *cabinetPanel) Paint(p2 *kit.Painter) {
	r := p.Bounds()
	if r.Empty() {
		return
	}
	lay, ok := p.plan()
	if ok && p.tray.slot >= 0 && p.paintTray(p2, r) {
		return
	}
	if !ok {
		ink, face := cabinet.PlateInk(false), cabinet.PlateFace(false)
		st := kit.Style{}.Foreground(theme.RGB(ink.R, ink.G, ink.B)).Background(theme.RGB(face.R, face.G, face.B))
		p2.Fill(r, st.Bg)
		p2.Print(r.Min.X, r.Min.Y, r.Dx(), kit.Escape(rowid.Line("ui.cabinet_panel.Paint.01", "The window is too small for the cabinet.")), st)
		return
	}
	if k := (cabinetKey{lay.Cells, lay.Cell, p.drawers}); p.front == nil || k != p.drawn {
		p.drawn, p.front, p.patchSel = k, cabinet.Front(lay, p.drawers[:], p.mint), -1
		p.renders++
	}
	if p.patchSel != p.sel {
		p.patchSel = p.sel
		p.patchAt, p.patchImg = cabinet.Patch(lay, p.drawers[:], p.sel, p.mint)
	}
	p2.Picture(r, p.front)
	p2.Picture(p.patchAt.Add(r.Min), p.patchImg)
	for i, c := range lay.Plates {
		p.printPlate(p2, r, c, p.names[i], p.dim[i])
	}
	p.printPlate(p2, r, lay.Title, p.totals(), false)
	p.printPlate(p2, r, lay.Hint, rowid.Line("ui.cabinet_panel.Paint.02", "Arrows move   Enter opens   Esc closes"), false)
}

// move changes the selection by a key; the grid does not wrap.
func (p *cabinetPanel) move(key kit.KeyCode) {
	col, row := p.sel%cabinet.Cols, p.sel/cabinet.Cols
	switch key {
	case kit.KeyLeft:
		col = max(col-1, 0)
	case kit.KeyRight:
		col = min(col+1, cabinet.Cols-1)
	case kit.KeyUp:
		row = max(row-1, 0)
	case kit.KeyDown:
		row = min(row+1, cabinet.Rows-1)
	case kit.KeyHome:
		col, row = 0, 0
	case kit.KeyEnd:
		col, row = cabinet.Cols-1, cabinet.Rows-1
	}
	p.sel = row*cabinet.Cols + col
}

// routeKey takes a key if it is one of the cabinet's and reports whether it did.
func (p *cabinetPanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyEnter:
		if prompt != "" {
			return false
		}
		if p.tray.slot >= 0 {
			p.openFan()
			return true
		}
		return p.openDrawer(p.sel)
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight, kit.KeyHome, kit.KeyEnd, kit.KeyPgUp, kit.KeyPgDn:
		if p.tray.slot >= 0 && agelessSlot(p.tray.slot) {
			p.moveGrid(ev.Code, len(p.agelessWells(p.tray.slot)))
		} else if ev.Code == kit.KeyPgUp || ev.Code == kit.KeyPgDn {
			return false
		} else if p.tray.fan {
			p.stepFan(ev.Code)
		} else if p.tray.slot >= 0 {
			p.moveWell(ev.Code)
		} else {
			p.move(ev.Code)
		}
		return true
	}
	return false
}

// drawerAt is the drawer under a screen cell, or -1.
func (p *cabinetPanel) drawerAt(at image.Point) int {
	lay, ok := p.plan()
	if !ok {
		return -1
	}
	at = at.Sub(p.Bounds().Min)
	return lay.At(at.X, at.Y)
}

// Mouse: a click selects the drawer under the pointer; the wheel moves to
// the next or the previous drawer.
func (p *cabinetPanel) Mouse(m kit.Mouse) bool {
	if p.tray.slot >= 0 {
		w := p.wellAt(m.Pt())
		if m.Kind == kit.MousePress && m.Button == kit.ButtonLeft && w >= 0 {
			if w == p.tray.well && !p.openFan() || w != p.tray.well {
				p.tray.well, p.tray.fan = w, false
			}
		}
		return m.Kind == kit.MousePress && m.Button == kit.ButtonRight && p.closeDrawer() || w >= 0
	}
	switch {
	case m.Kind == kit.MouseWheel && m.Wheel != 0:
		step := 1
		if m.Wheel < 0 {
			step = -1
		}
		p.sel = max(0, min(p.sel+step, cabinet.Slots-1))
		return true
	case m.Kind == kit.MousePress && m.Button == kit.ButtonLeft:
		if i := p.drawerAt(m.Pt()); i >= 0 {
			if i == p.sel {
				p.openDrawer(i)
			}
			p.sel = i
			return true
		}
	}
	return false
}

// PointerAt: the hand over a drawer.
func (p *cabinetPanel) PointerAt(at image.Point) kit.Pointer {
	if p.tray.slot >= 0 {
		return handIf(p.wellAt(at) >= 0)
	}
	return handIf(p.drawerAt(at) >= 0)
}

// findCabinetDrawer reads a word as a drawer: an age by its key or its name
// (with or without "age"), or Ledgers, Curios or Themes. An age the player
// cannot see named has no name to find.
func findCabinetDrawer(state game.GameState, word string) (string, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	sight := game.SightOf(&state)
	for _, a := range config.Ages() {
		n := strings.ToLower(a.Name)
		if sight.Age(a.Key) && (word == a.Key || word == strings.ReplaceAll(a.Key, "_", " ") || word == n || word+" age" == n) {
			return a.Key, true
		}
	}
	for _, l := range cabinetAgeless {
		if word == l.key {
			return l.key, true
		}
	}
	return "", false
}

// menuCabinetPage is the page the cabinet is shown on from the main menu.
const menuCabinetPage = "menu_cabinet"

// menuCabinet is the cabinet as a page of its own, over the menu: Esc or q
// closes it.
type menuCabinet struct {
	*cabinetPanel
	close func()
}

// Key: the cabinet's own keys, and Esc (or q) to close.
func (c *menuCabinet) Key(k kit.Key) bool {
	if k.Code == kit.KeyEsc && c.closeDrawer() {
		return true
	}
	if k.Code == kit.KeyEsc || k.Code == kit.KeyRune && (k.Rune == 'q' || k.Rune == 'Q') {
		c.close()
	} else {
		c.routeKey(&k, "")
	}
	return true
}

// openMenuCabinet shows the cabinet over the menu and returns to returnPage
// when it is closed. It draws from the account, which needs no game loaded.
func openMenuCabinet(app *appLoop, pages *nativePages, engine *game.GameEngine, returnPage string) {
	p := newCabinetPanel(engine.Account, engine.Rules())
	st := game.GameState{}
	if a := engine.Account(); a != nil {
		s, _ := a.LifetimeStats()
		st.AccountStats = &game.AccountStatsView{HighestAge: s.HighestAge}
	}
	p.update(st)
	c := &menuCabinet{cabinetPanel: p}
	c.close = func() {
		pages.RemovePage(menuCabinetPage)
		pages.SwitchToPage(returnPage)
	}
	pages.AddPage(menuCabinetPage, c, true, true)
	app.SetFocus(c)
}
