package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// map_settings.go holds the two map settings, "map style" and "map glyphs".
// They travel with the account (AccountPrefs, like the active theme), so a
// save/load keeps them and an account switch swaps them. The game still runs
// without an account: the dashboard then keeps them for the session only.

// mapSettings is the resolved setting pair, defaults filled in.
type mapSettings struct {
	Style string
	Tier  mapmodel.GlyphTier
	// HintShown: the one-time "Type icons" hint was shown on this account.
	HintShown bool
}

// resolveMapSettings reads the account's map settings against the style
// registry. An empty or unknown style is the registry default (roguelike);
// an empty or unknown tier is unicode. acct may be nil (defaults).
func resolveMapSettings(acct *game.Account, reg *mapstyle.Registry) mapSettings {
	out := mapSettings{Style: reg.Default(), Tier: mapmodel.TierUnicode}
	if acct == nil {
		return out
	}
	style, glyphs, hint := acct.MapPrefs()
	if _, ok := reg.New(style); ok {
		out.Style = style
	}
	if t, ok := mapmodel.ParseTier(glyphs); ok {
		out.Tier = t
	}
	out.HintShown = hint
	return out
}

// styleTitle is a style's display name ("Roguelike"), or the key.
func styleTitle(reg *mapstyle.Registry, name string) string {
	for _, e := range reg.Entries() {
		if e.Name == name {
			return e.Title
		}
	}
	return name
}

// nextStyle is the style after name in registry order, wrapping.
func nextStyle(reg *mapstyle.Registry, name string) string {
	names := reg.Names()
	for i, n := range names {
		if n == name {
			return names[(i+1)%len(names)]
		}
	}
	return reg.Default()
}

// nextTier is the glyph tier after t in the setting's order
// (ascii, unicode, nerd), wrapping.
func nextTier(t mapmodel.GlyphTier) mapmodel.GlyphTier {
	for i, n := range mapmodel.TierNames {
		if n == t.String() {
			next, _ := mapmodel.ParseTier(mapmodel.TierNames[(i+1)%len(mapmodel.TierNames)])
			return next
		}
	}
	return mapmodel.TierUnicode
}

// mapSettings is the dashboard's current map settings: the account's, or
// the session's when no account is loaded.
func (d *Dashboard) mapSettings() mapSettings {
	if d.engine != nil {
		if acct := d.engine.Account(); acct != nil {
			return resolveMapSettings(acct, d.mapViews.reg)
		}
	}
	if d.mapLocal == nil {
		s := resolveMapSettings(nil, d.mapViews.reg)
		d.mapLocal = &s
	}
	return *d.mapLocal
}

// saveMapSettings persists a style or glyph change from the Map panel's
// keys. A failed write keeps the change for the session and says so in
// the log (the account file is the only thing that failed).
func (d *Dashboard) saveMapSettings(s mapSettings) {
	var acct *game.Account
	if d.engine != nil {
		acct = d.engine.Account()
	}
	if acct == nil {
		d.mapLocal = &s
		return
	}
	style, glyphs, _ := acct.MapPrefs()
	var err error
	if style != s.Style {
		err = acct.SetMapStyle(s.Style)
	}
	if glyphs != s.Tier.String() && err == nil {
		err = acct.SetMapGlyphs(s.Tier.String())
	}
	if err != nil {
		d.engine.AddLog("warning", "The map setting could not be saved to your account: "+err.Error())
	}
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
		d.engine.AddLog("warning", "The map setting could not be saved to your account: "+err.Error())
	}
}
