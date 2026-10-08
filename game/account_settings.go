package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// account_settings.go keeps the account's settings that came after
// account.json's format was fixed.
//
// account.json is signed over every field it holds, preferences included,
// and a build that does not know a field drops it and signs what is left:
// a new preference written there would make an older build read a healthy
// account as edited. So a setting newer than the file's format lives in a
// file of its own beside it, settings.json. It holds display preferences
// and nothing an account earns, so it is plain and unsigned; a missing or
// unreadable one reads as the defaults. `account backup` copies it; an
// export does not carry it.

// settingsFileName is the settings file's name in an account's slot:
// <root>/accounts/<account_id>/settings.json.
const settingsFileName = "settings.json"

// settingsFileVersion is the version settings.json is written at.
const settingsFileVersion = 1

// accountSettings is settings.json.
type accountSettings struct {
	Version int `json:"version"`
	// Motion is the motion setting: "off", or empty for the default (on).
	Motion string `json:"motion,omitempty"`
	// Title is the title the account chose to wear, "" for the title its
	// badge score holds. A title it no longer holds is not worn.
	Title string `json:"title,omitempty"`
}

// settingsLocked returns the account's settings, read from its slot the
// first time they are asked for. Callers hold a.mu.
func (a *Account) settingsLocked() *accountSettings {
	if a.settings != nil {
		return a.settings
	}
	a.settings = &accountSettings{Version: settingsFileVersion}
	if !validAccountID(a.AccountID) {
		return a.settings
	}
	data, err := os.ReadFile(filepath.Join(accountDir(a.AccountID), settingsFileName))
	if err != nil {
		return a.settings
	}
	var s accountSettings
	if json.Unmarshal(data, &s) == nil {
		s.Version = settingsFileVersion
		a.settings = &s
	}
	return a.settings
}

// saveSettingsLocked writes settings.json into the account's slot,
// atomically. Callers hold a.mu.
func (a *Account) saveSettingsLocked() error {
	if !validAccountID(a.AccountID) {
		return fmt.Errorf("cannot save settings for an account with the ID %q", a.AccountID)
	}
	dir := accountDir(a.AccountID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	data, err := json.MarshalIndent(a.settingsLocked(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	path := filepath.Join(dir, settingsFileName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("failed to finalize settings: %w", err)
	}
	return nil
}

// MotionOn reports whether the game's motion is on (the default): the
// animation of the maps and of badges, and a theme's ambient effect.
func (a *Account) MotionOn() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settingsLocked().Motion != "off"
}

// SetMotion persists the motion setting. It writes settings.json only:
// account.json is not touched.
func (a *Account) SetMotion(on bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.settingsLocked()
	s.Motion = ""
	if !on {
		s.Motion = "off"
	}
	return a.saveSettingsLocked()
}

// SetTitle chooses the title the account wears: one it holds, or "" to go
// back to the title its badge score holds. It writes settings.json only.
// held is the titles the account holds now (BadgeSummary.Titles).
func (a *Account) SetTitle(title string, held []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if title != "" {
		found := false
		for _, t := range held {
			if strings.EqualFold(t, title) {
				title, found = t, true
			}
		}
		if !found {
			return fmt.Errorf("You do not hold the title %q.", title)
		}
	}
	a.settingsLocked().Title = title
	a.badgeRev++ // the badge summary carries the title worn
	return a.saveSettingsLocked()
}
