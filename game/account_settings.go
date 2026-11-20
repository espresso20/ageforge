package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/rowid"
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
	// LastPlayed is the save last loaded or saved on this account, and
	// MainGame the save the player marked as the main game, "" for none.
	// Together they say which game the main menu's Continue opens
	// (current_game.go). A name with no save behind it is ignored.
	LastPlayed string `json:"last_played,omitempty"`
	MainGame   string `json:"main_game,omitempty"`
	// Found is the panels the account has found and Opened those it has
	// opened (panels.go in ui): a found panel is on the list for good.
	Found  []string `json:"found,omitempty"`
	Opened []string `json:"opened,omitempty"`
	// Lessons is the keys of the lessons the account has learned (lessons.go
	// in ui): a learned lesson is not shown again.
	Lessons []string `json:"lessons,omitempty"`
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
		return rowid.Errorf("game.account_settings.saveSettingsLocked.01", "cannot save settings for an account with the ID %q", a.AccountID)
	}
	dir := accountDir(a.AccountID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	data, err := json.MarshalIndent(a.settingsLocked(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := writeFileAtomic(filepath.Join(dir, settingsFileName), data); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}
	return nil
}

// readSettings is a copy of the account's settings.
func (a *Account) readSettings() accountSettings {
	a.mu.Lock()
	defer a.mu.Unlock()
	return *a.settingsLocked()
}

// editSettings applies change to the account's settings and writes
// settings.json.
func (a *Account) editSettings(change func(*accountSettings)) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	change(a.settingsLocked())
	return a.saveSettingsLocked()
}

// editSettingsIf is editSettings, writing the file only when change reports
// that it changed something.
func (a *Account) editSettingsIf(change func(*accountSettings) bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !change(a.settingsLocked()) {
		return nil
	}
	return a.saveSettingsLocked()
}

// MotionOn reports whether the game's motion is on (the default): the
// animation of the maps and of badges, and a theme's ambient effect.
func (a *Account) MotionOn() bool {
	return a.readSettings().Motion != "off"
}

// SetMotion persists the motion setting. It writes settings.json only:
// account.json is not touched.
func (a *Account) SetMotion(on bool) error {
	return a.editSettings(func(s *accountSettings) {
		s.Motion = ""
		if !on {
			s.Motion = "off"
		}
	})
}

// SetTitle chooses the title the account wears: one it holds, or "" to go
// back to the title its badge score holds. It writes settings.json only.
// held is the titles the account holds now (BadgeSummary.Titles).
func (a *Account) SetTitle(title string, held []string) error {
	if title != "" {
		found := false
		for _, t := range held {
			if strings.EqualFold(t, title) {
				title, found = t, true
			}
		}
		if !found {
			return rowid.Errorf("game.account_settings.SetTitle.01", "You do not hold the title %q.", title)
		}
	}
	return a.editSettings(func(s *accountSettings) {
		s.Title = title
		a.badgeRev++ // the badge summary carries the title worn
	})
}

// PanelsSeen is the panels the account has found and the ones it has opened.
func (a *Account) PanelsSeen() (found, opened []string) {
	s := a.readSettings()
	return slices.Clone(s.Found), slices.Clone(s.Opened)
}

// NotePanels adds panels to the found and opened lists. It writes the
// settings file only when a list changes.
func (a *Account) NotePanels(found, opened []string) error {
	s := a.readSettings()
	grow := func(have, add []string) []string {
		for _, n := range add {
			if !slices.Contains(have, n) {
				have = append(have, n)
			}
		}
		return have
	}
	f, o := grow(slices.Clone(s.Found), found), grow(slices.Clone(s.Opened), opened)
	if len(f) == len(s.Found) && len(o) == len(s.Opened) {
		return nil
	}
	return a.editSettings(func(s *accountSettings) { s.Found, s.Opened = grow(s.Found, f), grow(s.Opened, o) })
}

// LessonsLearned is the keys of the lessons the account has learned.
func (a *Account) LessonsLearned() []string {
	return slices.Clone(a.readSettings().Lessons)
}

// LearnLessons adds keys to the learned lessons. It writes the settings file
// only when the list grows.
func (a *Account) LearnLessons(keys ...string) error {
	return a.editSettingsIf(func(s *accountSettings) bool {
		n := len(s.Lessons)
		for _, k := range keys {
			if !slices.Contains(s.Lessons, k) {
				s.Lessons = append(s.Lessons, k)
			}
		}
		return len(s.Lessons) != n
	})
}
