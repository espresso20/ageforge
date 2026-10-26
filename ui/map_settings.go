package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

// map_settings.go holds the display settings a command sets: "map glyphs",
// "minimap" and "motion". They travel with the account (AccountPrefs, like
// the active theme; motion in the account's settings file), so a save/load
// keeps them and an account switch swaps them. The game still runs without
// an account: the dashboard then keeps them for the session only.

// mapSettings is the resolved settings, defaults filled in.
type mapSettings struct {
	Tier mapmodel.GlyphTier
	// Minimap: the dashboard shows the mini map above the Buildings list.
	// On unless the player turned it off (minimap off).
	Minimap bool
	// HintShown: the one-time "Type icons" hint was shown on this account.
	HintShown bool
	// Motion: things move (the default). Off holds the maps, the badge
	// case and a theme's ambient effect still.
	Motion bool
}

// defaultMapSettings is the settings with nothing chosen.
func defaultMapSettings() mapSettings {
	return mapSettings{Tier: mapmodel.TierUnicode, Minimap: true, Motion: true}
}

// resolveMapSettings reads the account's map settings. An empty or unknown
// tier is unicode; the mini map is on unless turned off. acct may be nil
// (defaults).
func resolveMapSettings(acct *game.Account) mapSettings {
	out := defaultMapSettings()
	if acct == nil {
		return out
	}
	glyphs, hint := acct.MapPrefs()
	if t, ok := mapmodel.ParseTier(glyphs); ok {
		out.Tier = t
	}
	out.Minimap = acct.MinimapOn()
	out.HintShown = hint
	out.Motion = acct.MotionOn()
	return out
}

// mapSettings is the dashboard's current map settings: the account's, or
// the session's when no account is loaded.
func (d *Dashboard) mapSettings() mapSettings {
	if d.engine != nil {
		if acct := d.engine.Account(); acct != nil {
			return resolveMapSettings(acct)
		}
	}
	if d.mapLocal == nil {
		s := resolveMapSettings(nil)
		d.mapLocal = &s
	}
	return *d.mapLocal
}

// saveMapSettings persists a settings change: to the account, or for the
// session when no account is loaded. A failed write says so in the log
// (the account file is the only thing that failed).
func (d *Dashboard) saveMapSettings(s mapSettings) {
	var acct *game.Account
	if d.engine != nil {
		acct = d.engine.Account()
	}
	if acct == nil {
		d.mapLocal = &s
		return
	}
	cur := resolveMapSettings(acct)
	var err error
	if cur.Tier != s.Tier {
		err = acct.SetMapGlyphs(s.Tier.String())
	}
	if cur.Minimap != s.Minimap && err == nil {
		err = acct.SetMinimap(s.Minimap)
	}
	if cur.Motion != s.Motion && err == nil {
		err = acct.SetMotion(s.Motion)
	}
	if err != nil {
		d.engine.AddLog("warning", vibe.Line("ui.map_settings.saveMapSettings.01", "The setting could not be saved to your account: ")+err.Error())
	}
}

// mapPref is one setting a command changed: "glyphs", "minimap" or
// "motion", and its new value. The zero value changes nothing.
type mapPref struct{ Key, Value string }

// apply writes the change into s.
func (p mapPref) apply(s *mapSettings) {
	switch p.Key {
	case "glyphs":
		if t, ok := mapmodel.ParseTier(p.Value); ok {
			s.Tier = t
		}
	case "minimap":
		s.Minimap = p.Value == "on"
	case "motion":
		s.Motion = p.Value == "on"
	}
}

// applyMapPref applies a map command's setting change at once: for the
// session when no account is loaded (with one, the command has already
// saved it and this changes nothing).
func (d *Dashboard) applyMapPref(p mapPref) {
	s := d.mapSettings()
	p.apply(&s)
	d.saveMapSettings(s)
}

// markMapHintShown records that the icons hint was shown.
func (d *Dashboard) markMapHintShown() {
	var acct *game.Account
	if d.engine != nil {
		acct = d.engine.Account()
	}
	if acct == nil {
		s := d.mapSettings()
		s.HintShown = true
		d.mapLocal = &s
		return
	}
	if err := acct.SetMapIconsHintShown(); err != nil {
		d.engine.AddLog("warning", vibe.Line("ui.map_settings.markMapHintShown.01", "The map setting could not be saved to your account: ")+err.Error())
	}
}
