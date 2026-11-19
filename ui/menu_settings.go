package ui

import (
	"runtime"
	"strconv"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
)

// menu_settings.go is the main menu's Settings page: the entries are swapped
// for one row per display setting, each with its value at the right. A change
// is made through the calls the commands make (the account's setters), so a
// command and the menu always agree. Esc, or Back, returns
// to the main menu.

// settingsItems are the Settings page's rows.
func (m *mainMenu) settingsItems() []menuItem {
	items := []menuItem{
		{id: miSetTheme, label: rowid.Line("ui.menu.buildViewWith.05", "Themes"), key: 't', extras: []string{theme.Active().Name.String()}},
		{id: miSetGlyphs, label: rowid.Line("ui.menu.settings.02", "Map glyphs"), key: 'g', extras: []string{m.set.Tier.String()}},
		{id: miSetMotion, label: rowid.Line("ui.menu.settings.03", "Motion"), key: 'm', extras: []string{settingOnOff(m.set.Motion)}, sw: true, on: m.set.Motion},
		{id: miSetMinimap, label: rowid.Line("ui.menu.settings.04", "Minimap"), key: 'i', extras: []string{settingOnOff(m.set.Minimap)}, sw: true, on: m.set.Minimap},
	}
	// The font, the text size and fullscreen are rows only when the game is
	// in its own window.
	if m.window != nil {
		items = append(items,
			menuItem{id: miSetFont, label: txtFont.String(), key: 'f', extras: []string{fontName(m.window.FontChoice())}},
			menuItem{id: miSetSize, label: rowid.Line("ui.menu.settings.09", "Text size"), key: 'z', extras: []string{strconv.Itoa(m.window.TextSize())}, size: true, hint: zoomKeysHint()},
			menuItem{id: miSetFull, label: rowid.Line("ui.menu.settings.10", "Fullscreen"), key: 'u', extras: []string{settingOnOff(m.window.Fullscreen())}, sw: true, on: m.window.Fullscreen(), hint: keyF11})
	}
	// Discord status is a row only in a build that has a Discord application id.
	if m.appID != "" {
		items = append(items, menuItem{id: miSetDiscord, label: rowid.Line("ui.menu.settings.08", "Discord status"), key: 'd', extras: []string{settingOnOff(m.discordOn())}, sw: true, on: m.discordOn(), off: m.acct == nil})
	}
	return append(items, menuItem{id: miSetBack, label: rowid.Line("ui.menu.settings.05", "Back"), key: 'b'})
}

var (
	txtFont      = rowid.E("ui.window.font.title", "Font")
	txtFontPixel = rowid.E("ui.window.font.pixel", "Pixel")
	txtFontMono  = rowid.E("ui.window.font.mono", "JetBrains Mono")
)

func fontName(choice string) string {
	if choice == "pixel" {
		return txtFontPixel.String()
	}
	return txtFontMono.String()
}

// discordOn reads the Discord status setting from the account; it is on with none.
func (m *mainMenu) discordOn() bool {
	return m.acct == nil || m.acct.DiscordOn()
}

func settingOnOff(on bool) string {
	if on {
		return rowid.Line("ui.menu.settings.06", "On")
	}
	return rowid.Line("ui.menu.settings.07", "Off")
}

// changeSetting changes the setting of row id: a switch flips; the glyph
// tier steps through the tiers (dir 1 forward, -1 back). It saves to the
// account at once, or for the menu's life when no account is loaded.
func (m *mainMenu) changeSetting(id string, dir int) {
	set := m.set
	m.settingsSel = m.view.sel // refresh rebuilds the rows on this one
	switch id {
	case miSetGlyphs:
		n := len(mapmodel.TierNames)
		cur := 0
		for i, name := range mapmodel.TierNames {
			if name == set.Tier.String() {
				cur = i
			}
		}
		next := mapmodel.TierNames[(cur+dir+n)%n]
		set.Tier, _ = mapmodel.ParseTier(next)
		if m.acct != nil {
			if err := m.acct.SetMapGlyphs(next); err != nil {
				return
			}
		}
	case miSetMotion:
		set.Motion = !set.Motion
		if m.acct != nil {
			if err := m.acct.SetMotion(set.Motion); err != nil {
				return
			}
		}
	case miSetMinimap:
		set.Minimap = !set.Minimap
		if m.acct != nil {
			if err := m.acct.SetMinimap(set.Minimap); err != nil {
				return
			}
		}
	case miSetFont:
		if m.window == nil {
			return
		}
		m.window.NextFont()
		m.refresh(false)
		return
	case miSetSize:
		if m.window == nil {
			return
		}
		m.zoom(dir)
		return
	case miSetFull:
		if m.window == nil {
			return
		}
		m.window.ToggleFullscreen()
		m.refresh(false)
		return
	case miSetDiscord:
		// Held by the account only: with none loaded there is nothing to keep it in.
		if m.acct == nil {
			return
		}
		if err := m.acct.SetDiscord(!m.acct.DiscordOn()); err != nil {
			return
		}
		if m.onDiscord != nil {
			m.onDiscord()
		}
	default:
		return
	}
	if m.acct == nil {
		m.local = &set
	}
	m.refresh(false)
}

// zoom changes the text size by steps, through the call the window's own
// keys make.
func (m *mainMenu) zoom(steps int) {
	if m.window == nil {
		return
	}
	m.settingsSel = m.view.sel
	m.window.ZoomBy(steps)
	m.refresh(false)
}

// keyF11 is the key the window's own fullscreen is on.
const keyF11 = "F11"

// zoomKeysHint is the window's own keys for the text size.
func zoomKeysHint() string {
	if runtime.GOOS == "darwin" {
		return "Cmd+/-"
	}
	return "Ctrl+/-"
}

// closeSettings returns to the main menu, on the Settings entry.
func (m *mainMenu) closeSettings() {
	m.settings = false
	m.buildView()
	for i, it := range m.view.items {
		if it.id == miSettings {
			m.view.sel = i
		}
	}
}
