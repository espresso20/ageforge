package ui

import (
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// menu_settings.go is the main menu's Settings page: the entries are swapped
// for one row per display setting, each with its value at the right. A change
// is made through the calls the commands make (the account's setters,
// SetVibeMode), so a command and the menu always agree. Esc, or Back, returns
// to the main menu.

// settingsItems are the Settings page's rows.
func (m *mainMenu) settingsItems() []menuItem {
	items := []menuItem{
		{id: miSetTheme, label: vibe.Line("ui.menu.buildViewWith.05", "Themes"), key: 't', extras: []string{theme.Active().Name.String()}},
		{id: miSetVibe, label: "VibeMode", key: 'v', extras: []string{settingOnOff(vibe.On())}},
		{id: miSetGlyphs, label: vibe.Line("ui.menu.settings.02", "Map glyphs"), key: 'g', extras: []string{m.set.Tier.String()}},
		{id: miSetMotion, label: vibe.Line("ui.menu.settings.03", "Motion"), key: 'm', extras: []string{settingOnOff(m.set.Motion)}},
		{id: miSetMinimap, label: vibe.Line("ui.menu.settings.04", "Minimap"), key: 'i', extras: []string{settingOnOff(m.set.Minimap)}},
	}
	// Discord status is a row only in a build that has a Discord application id.
	if m.appID != "" {
		items = append(items, menuItem{id: miSetDiscord, label: vibe.Line("ui.menu.settings.08", "Discord status"), key: 'd', extras: []string{settingOnOff(m.discordOn())}})
	}
	return append(items, menuItem{id: miSetBack, label: vibe.Line("ui.menu.settings.05", "Back"), key: 'b'})
}

// discordOn reads the Discord status setting from the account; it is on with none.
func (m *mainMenu) discordOn() bool {
	return m.acct == nil || m.acct.DiscordOn()
}

func settingOnOff(on bool) string {
	if on {
		return vibe.Line("ui.menu.settings.06", "On")
	}
	return vibe.Line("ui.menu.settings.07", "Off")
}

// changeSetting changes the setting of row id: a switch flips; the glyph
// tier steps through the tiers (dir 1 forward, -1 back). It saves to the
// account at once, or for the menu's life when no account is loaded.
func (m *mainMenu) changeSetting(id string, dir int) {
	set := m.set
	m.settingsSel = m.view.sel // refresh rebuilds the rows on this one
	switch id {
	case miSetVibe:
		// The same call the vibemode command makes.
		if err := m.engine.SetVibeMode(!vibe.On()); err != nil {
			return
		}
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
	if m.acct == nil && id != miSetVibe {
		m.local = &set
	}
	m.refresh(false)
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
